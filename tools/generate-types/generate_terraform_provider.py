#!/usr/bin/env python3
"""
Generate Go Terraform provider resource + data source files from CiviCRM API v4 getFields.

Output per entity (written to --output-dir):
  resource_<snake>.go        — full CRUD resource
  data_source_<snake>.go     — read-only data source (lookup by id or name)

Usage:
    python generate_terraform_provider.py \\
        --url https://example.org --api-key YOUR_KEY --entities PriceSet,PriceField

    python generate_terraform_provider.py \\
        --url https://example.org --api-key YOUR_KEY \\
        --entities MembershipType --output-dir internal/provider

    python generate_terraform_provider.py \\
        --url https://example.org --api-key YOUR_KEY --insecure \\
        --entities Tag

After generation, register the new constructors in provider.go:
  Resources():    New<Entity>Resource
  DataSources():  New<Entity>DataSource
"""

import argparse
import json
import re
import sys
import urllib.error
import urllib.parse
import urllib.request
import ssl
from pathlib import Path

# ─────────────────────────────────────────────────────────────
# Type mapping: CiviCRM data_type → Go terraform-plugin-framework
# ─────────────────────────────────────────────────────────────

# (go_type, schema_type, getter)
TYPE_MAP = {
    "Integer":   ("types.Int64",   "schema.Int64Attribute",   "Int64"),
    "Float":     ("types.Float64", "schema.Float64Attribute", "Float64"),
    "Money":     ("types.Float64", "schema.Float64Attribute", "Money"),   # special reader
    "Boolean":   ("types.Bool",    "schema.BoolAttribute",    "Bool"),
    "String":    ("types.String",  "schema.StringAttribute",  "String"),
    "Text":      ("types.String",  "schema.StringAttribute",  "String"),
    "Blob":      ("types.String",  "schema.StringAttribute",  "String"),
    "Date":      ("types.String",  "schema.StringAttribute",  "String"),
    "Timestamp": ("types.String",  "schema.StringAttribute",  "String"),
    "Array":     ("types.String",  "schema.StringAttribute",  "String"),  # serialized
}

DEFAULT_TYPE = ("types.String", "schema.StringAttribute", "String")

# ─────────────────────────────────────────────────────────────
# HTTP helpers (same pattern as generate_civicrm_types.py)
# ─────────────────────────────────────────────────────────────

def _ssl_ctx(insecure: bool) -> ssl.SSLContext:
    ctx = ssl.create_default_context()
    if insecure:
        ctx.check_hostname = False
        ctx.verify_mode = ssl.CERT_NONE
    return ctx


def _api_call(base_url: str, api_key: str, entity: str, action: str,
              params: dict, insecure: bool) -> list:
    url = f"{base_url.rstrip('/')}/civicrm/ajax/api4/{entity}/{action}"
    params_json = json.dumps(params)
    ctx = _ssl_ctx(insecure)

    for method in ("POST", "GET"):
        try:
            if method == "GET":
                full = url + "?" + urllib.parse.urlencode({"params": params_json})
                req = urllib.request.Request(full, method="GET")
            else:
                data = urllib.parse.urlencode({"params": params_json}).encode()
                req = urllib.request.Request(url, data=data, method="POST")
                req.add_header("Content-Type", "application/x-www-form-urlencoded")

            req.add_header("Authorization", f"Bearer {api_key}")
            req.add_header("X-Requested-With", "XMLHttpRequest")
            req.add_header("Accept", "application/json")

            with urllib.request.urlopen(req, context=ctx, timeout=15) as resp:
                body = resp.read()
            break
        except urllib.error.HTTPError as e:
            if e.code == 405 and method == "POST":
                continue
            print(f"  HTTP {e.code} for {entity}.{action}: {e.reason}", file=sys.stderr)
            return []
        except Exception as e:
            print(f"  Error for {entity}.{action}: {e}", file=sys.stderr)
            return []
    else:
        return []

    try:
        parsed = json.loads(body.decode())
    except json.JSONDecodeError:
        print(f"  Could not parse JSON for {entity}.{action}", file=sys.stderr)
        return []

    if parsed.get("error_message"):
        print(f"  API error for {entity}.{action}: {parsed['error_message']}", file=sys.stderr)
        return []

    return parsed.get("values", [])


def fetch_fields(base_url: str, api_key: str, entity: str, insecure: bool) -> list[dict]:
    return _api_call(base_url, api_key, entity, "getFields",
                     {"checkPermissions": False, "action": "create"}, insecure)

# ─────────────────────────────────────────────────────────────
# Naming helpers
# ─────────────────────────────────────────────────────────────

def snake(name: str) -> str:
    """PriceFieldValue → price_field_value"""
    s = re.sub(r"([A-Z]+)([A-Z][a-z])", r"\1_\2", name)
    s = re.sub(r"([a-z0-9])([A-Z])", r"\1_\2", s)
    return s.lower()


def go_field_name(api_name: str) -> str:
    """member_of_contact_id → MemberOfContactID"""
    parts = api_name.split("_")
    result = ""
    for p in parts:
        if not p:
            continue
        # Keep common acronyms uppercase
        if p.upper() in ("ID", "URL", "API", "HTML", "FK"):
            result += p.upper()
        else:
            result += p.capitalize()
    # Make trailing "Id" → "ID"
    result = re.sub(r"Id$", "ID", result)
    result = re.sub(r"Ids$", "IDs", result)
    return result

# ─────────────────────────────────────────────────────────────
# Field analysis
# ─────────────────────────────────────────────────────────────

def classify_field(f: dict) -> dict:
    """Return enriched field dict with go_type, schema_type, getter, etc."""
    data_type = f.get("data_type", "String")
    go_type, schema_type, getter = TYPE_MAP.get(data_type, DEFAULT_TYPE)

    required = bool(f.get("required", False))
    nullable = bool(f.get("nullable", True))
    name = f.get("name", "")
    is_id = name == "id"
    is_money = data_type == "Money"

    return {
        **f,
        "go_type": go_type,
        "schema_type": schema_type,
        "getter": getter,
        "is_required": required,
        "is_nullable": nullable,
        "is_id": is_id,
        "is_money": is_money,
        "go_field": go_field_name(name),
        "description": (f.get("description") or f.get("title") or name).strip(),
    }

# ─────────────────────────────────────────────────────────────
# Code generation helpers
# ─────────────────────────────────────────────────────────────

def _map_result_block(f: dict, receiver: str) -> list[str]:
    """Generate the mapResultToState lines for one field."""
    name = f["name"]
    go_field = f["go_field"]
    go_type = f["go_type"]
    getter = f["getter"]
    is_nullable = f["is_nullable"]
    is_money = f["is_money"]

    lines = []

    if is_money:
        lines += [
            f'\tif v, ok := result["{name}"]; ok && v != nil {{',
            "\t\tswitch val := v.(type) {",
            "\t\tcase float64:",
            f"\t\t\t{receiver}.{go_field} = types.Float64Value(val)",
            "\t\tcase string:",
            "\t\t\tif f, err := strconv.ParseFloat(val, 64); err == nil {",
            f"\t\t\t\t{receiver}.{go_field} = types.Float64Value(f)",
            "\t\t\t} else {",
            f"\t\t\t\t{receiver}.{go_field} = types.Float64Null()",
            "\t\t\t}",
            "\t\tdefault:",
            f"\t\t\t{receiver}.{go_field} = types.Float64Null()",
            "\t\t}",
            "\t} else {",
            f"\t\t{receiver}.{go_field} = types.Float64Null()",
            "\t}",
        ]
    elif go_type == "types.String":
        if is_nullable:
            lines += [
                f'\tif v, ok := GetString(result, "{name}"); ok && v != "" {{',
                f"\t\t{receiver}.{go_field} = types.StringValue(v)",
                "\t} else {",
                f"\t\t{receiver}.{go_field} = types.StringNull()",
                "\t}",
            ]
        else:
            lines += [
                f'\tif v, ok := GetString(result, "{name}"); ok {{',
                f"\t\t{receiver}.{go_field} = types.StringValue(v)",
                "\t}",
            ]
    elif go_type == "types.Int64":
        if is_nullable:
            lines += [
                f'\tif v, ok := GetInt64(result, "{name}"); ok {{',
                f"\t\t{receiver}.{go_field} = types.Int64Value(v)",
                "\t} else {",
                f"\t\t{receiver}.{go_field} = types.Int64Null()",
                "\t}",
            ]
        else:
            lines += [
                f'\tif v, ok := GetInt64(result, "{name}"); ok {{',
                f"\t\t{receiver}.{go_field} = types.Int64Value(v)",
                "\t}",
            ]
    elif go_type == "types.Float64":
        if is_nullable:
            lines += [
                f'\tif v, ok := GetFloat64(result, "{name}"); ok {{',
                f"\t\t{receiver}.{go_field} = types.Float64Value(v)",
                "\t} else {",
                f"\t\t{receiver}.{go_field} = types.Float64Null()",
                "\t}",
            ]
        else:
            lines += [
                f'\tif v, ok := GetFloat64(result, "{name}"); ok {{',
                f"\t\t{receiver}.{go_field} = types.Float64Value(v)",
                "\t}",
            ]
    elif go_type == "types.Bool":
        lines += [
            f'\tif v, ok := GetBool(result, "{name}"); ok {{',
            f"\t\t{receiver}.{go_field} = types.BoolValue(v)",
            "\t}",
        ]

    return lines


def _values_block(fields: list[dict], required_names: set[str]) -> list[str]:
    """Generate the values map[string]any{} block for Create."""
    lines = ["\tvalues := map[string]any{"]
    optional = []
    for f in fields:
        name = f["name"]
        go_field = f["go_field"]
        go_type = f["go_type"]
        if name in required_names:
            if go_type == "types.String":
                lines.append(f'\t\t"{name}": plan.{go_field}.ValueString(),')
            elif go_type == "types.Int64":
                lines.append(f'\t\t"{name}": plan.{go_field}.ValueInt64(),')
            elif go_type in ("types.Float64",):
                lines.append(f'\t\t"{name}": plan.{go_field}.ValueFloat64(),')
            elif go_type == "types.Bool":
                lines.append(f'\t\t"{name}": plan.{go_field}.ValueBool(),')
        else:
            optional.append(f)
    lines.append("\t}")
    lines.append("")

    for f in optional:
        name = f["name"]
        go_field = f["go_field"]
        go_type = f["go_type"]
        lines.append(f'\tif !plan.{go_field}.IsNull() {{')
        if go_type == "types.String":
            lines.append(f'\t\tvalues["{name}"] = plan.{go_field}.ValueString()')
        elif go_type == "types.Int64":
            lines.append(f'\t\tvalues["{name}"] = plan.{go_field}.ValueInt64()')
        elif go_type in ("types.Float64",):
            lines.append(f'\t\tvalues["{name}"] = plan.{go_field}.ValueFloat64()')
        elif go_type == "types.Bool":
            lines.append(f'\t\tvalues["{name}"] = plan.{go_field}.ValueBool()')
        lines.append("\t}")

    return lines


def _values_block_update(fields: list[dict], required_names: set[str]) -> list[str]:
    """Generate the values map for Update (optional fields explicitly set to nil when absent)."""
    lines = ["\tvalues := map[string]any{"]
    optional = []
    for f in fields:
        name = f["name"]
        go_field = f["go_field"]
        go_type = f["go_type"]
        if name in required_names:
            if go_type == "types.String":
                lines.append(f'\t\t"{name}": plan.{go_field}.ValueString(),')
            elif go_type == "types.Int64":
                lines.append(f'\t\t"{name}": plan.{go_field}.ValueInt64(),')
            elif go_type in ("types.Float64",):
                lines.append(f'\t\t"{name}": plan.{go_field}.ValueFloat64(),')
            elif go_type == "types.Bool":
                lines.append(f'\t\t"{name}": plan.{go_field}.ValueBool(),')
        else:
            optional.append(f)
    lines.append("\t}")
    lines.append("")

    for f in optional:
        name = f["name"]
        go_field = f["go_field"]
        go_type = f["go_type"]
        lines.append(f'\tif !plan.{go_field}.IsNull() {{')
        if go_type == "types.String":
            lines.append(f'\t\tvalues["{name}"] = plan.{go_field}.ValueString()')
        elif go_type == "types.Int64":
            lines.append(f'\t\tvalues["{name}"] = plan.{go_field}.ValueInt64()')
        elif go_type in ("types.Float64",):
            lines.append(f'\t\tvalues["{name}"] = plan.{go_field}.ValueFloat64()')
        elif go_type == "types.Bool":
            lines.append(f'\t\tvalues["{name}"] = plan.{go_field}.ValueBool()')
        lines.append("\t} else {")
        lines.append(f'\t\tvalues["{name}"] = nil')
        lines.append("\t}")

    return lines

# ─────────────────────────────────────────────────────────────
# Resource generator
# ─────────────────────────────────────────────────────────────

def generate_resource(entity: str, raw_fields: list[dict]) -> str:
    fields = [classify_field(f) for f in raw_fields if f.get("name") != "id"]
    id_field = classify_field(next((f for f in raw_fields if f.get("name") == "id"),
                                   {"name": "id", "data_type": "Integer", "description": "ID",
                                    "required": False, "nullable": False}))

    entity_snake = snake(entity)

    # Determine which fields are truly required (required=True, not the id)
    required_names = {f["name"] for f in fields if f["is_required"]}
    # Bool fields that have defaults should always be in the required map
    bool_fields_with_defaults = {f["name"] for f in fields if f["go_type"] == "types.Bool"}
    required_names |= bool_fields_with_defaults

    # ── imports ──────────────────────────────────────────────
    needs_strconv = True
    imports = [
        '"context"',
        '"fmt"',
        '"strconv"',
        "",
        '"github.com/hashicorp/terraform-plugin-framework/path"',
        '"github.com/hashicorp/terraform-plugin-framework/resource"',
        '"github.com/hashicorp/terraform-plugin-framework/resource/schema"',
        '"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"',
        '"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"',
        '"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"',
        '"github.com/hashicorp/terraform-plugin-framework/types"',
        '"github.com/hashicorp/terraform-plugin-log/tflog"',
    ]

    # ── struct fields ─────────────────────────────────────────
    struct_lines = [
        f"type {entity}ResourceModel struct {{",
        f"\tID {id_field['go_type']} `tfsdk:\"id\"`",
    ]
    for f in fields:
        struct_lines.append(f'\t{f["go_field"]} {f["go_type"]} `tfsdk:"{f["name"]}"`')
    struct_lines.append("}")

    # ── schema attributes ─────────────────────────────────────
    schema_lines = [
        '\t\t\t"id": schema.Int64Attribute{',
        f'\t\t\t\tDescription: "The unique identifier of the {entity_snake}.",',
        "\t\t\t\tComputed:    true,",
        "\t\t\t\tPlanModifiers: []planmodifier.Int64{",
        "\t\t\t\t\tint64planmodifier.UseStateForUnknown(),",
        "\t\t\t\t},",
        "\t\t\t},",
    ]
    for f in fields:
        desc = f["description"]
        stype = f["schema_type"]
        name = f["name"]
        is_req = f["is_required"]
        is_nullable = f["is_nullable"]
        go_type = f["go_type"]

        schema_lines.append(f'\t\t\t"{name}": {stype}{{')
        schema_lines.append(f'\t\t\t\tDescription: "{desc}.",')

        if is_req:
            schema_lines.append("\t\t\t\tRequired:    true,")
        elif go_type == "types.Bool":
            # Bool fields get Optional+Computed+Default
            schema_lines.append("\t\t\t\tOptional:    true,")
            schema_lines.append("\t\t\t\tComputed:    true,")
            schema_lines.append("\t\t\t\tDefault:     booldefault.StaticBool(false),")
        else:
            schema_lines.append("\t\t\t\tOptional:    true,")
            schema_lines.append("\t\t\t\tComputed:    true,")

        schema_lines.append("\t\t\t},")

    # ── mapResultToState ──────────────────────────────────────
    map_lines = [
        f"func (r *{entity}Resource) mapResultToState(result map[string]any, model *{entity}ResourceModel) {{",
        f'\tif id, ok := GetInt64(result, "id"); ok {{',
        "\t\tmodel.ID = types.Int64Value(id)",
        "\t}",
        "",
    ]
    for f in fields:
        map_lines += _map_result_block(f, "model")
        map_lines.append("")
    map_lines.append("}")

    # ── values blocks ─────────────────────────────────────────
    create_values = _values_block(fields, required_names)
    update_values = _values_block_update(fields, required_names)

    # ── assemble ──────────────────────────────────────────────
    log_field = fields[0]["name"] if fields else "id"
    log_go_field = fields[0]["go_field"] if fields else "ID"
    log_value_expr = (
        f"plan.{log_go_field}.ValueString()"
        if fields and fields[0]["go_type"] == "types.String"
        else f"plan.{log_go_field}.ValueInt64()"
        if fields and fields[0]["go_type"] == "types.Int64"
        else f"plan.{log_go_field}.ValueBool()"
    )

    out = [
        "package provider",
        "",
        "import (",
    ]
    for imp in imports:
        if imp:
            out.append(f"\t{imp}")
        else:
            out.append("")
    out += [
        ")",
        "",
        "var (",
        f"\t_ resource.Resource                = &{entity}Resource{{}}",
        f"\t_ resource.ResourceWithConfigure   = &{entity}Resource{{}}",
        f"\t_ resource.ResourceWithImportState = &{entity}Resource{{}}",
        ")",
        "",
        f"type {entity}Resource struct {{",
        "\tclient *Client",
        "}",
        "",
    ]
    out += struct_lines
    out += [
        "",
        f"func New{entity}Resource() resource.Resource {{",
        f"\treturn &{entity}Resource{{}}",
        "}",
        "",
        f"func (r *{entity}Resource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {{",
        f'\tresp.TypeName = req.ProviderTypeName + "_{entity_snake}"',
        "}",
        "",
        f"func (r *{entity}Resource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {{",
        "\tresp.Schema = schema.Schema{",
        f'\t\tDescription: "Manages a CiviCRM {entity}.",',
        "\t\tAttributes: map[string]schema.Attribute{",
    ]
    out += schema_lines
    out += [
        "\t\t},",
        "\t}",
        "}",
        "",
        f"func (r *{entity}Resource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {{",
        "\tif req.ProviderData == nil {",
        "\t\treturn",
        "\t}",
        "\tclient, ok := req.ProviderData.(*Client)",
        "\tif !ok {",
        "\t\tresp.Diagnostics.AddError(",
        '\t\t\t"Unexpected Resource Configure Type",',
        f'\t\t\tfmt.Sprintf("Expected *Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),',
        "\t\t)",
        "\t\treturn",
        "\t}",
        "\tr.client = client",
        "}",
        "",
        f"func (r *{entity}Resource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {{",
        f"\tvar plan {entity}ResourceModel",
        "\tdiags := req.Plan.Get(ctx, &plan)",
        "\tresp.Diagnostics.Append(diags...)",
        "\tif resp.Diagnostics.HasError() {",
        "\t\treturn",
        "\t}",
        "",
        f'\ttflog.Debug(ctx, "Creating {entity}", map[string]any{{"{log_field}": {log_value_expr}}})',
        "",
    ]
    out += create_values
    out += [
        "",
        f'\tresult, err := r.client.Create("{entity}", values)',
        "\tif err != nil {",
        "\t\tresp.Diagnostics.AddError(",
        f'\t\t\t"Error creating {entity}",',
        f'\t\t\t"Could not create {entity}: "+err.Error(),',
        "\t\t)",
        "\t\treturn",
        "\t}",
        "",
        "\tr.mapResultToState(result, &plan)",
        "",
        f'\ttflog.Debug(ctx, "Created {entity}", map[string]any{{"id": plan.ID.ValueInt64()}})',
        "",
        "\tdiags = resp.State.Set(ctx, plan)",
        "\tresp.Diagnostics.Append(diags...)",
        "}",
        "",
        f"func (r *{entity}Resource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {{",
        f"\tvar state {entity}ResourceModel",
        "\tdiags := req.State.Get(ctx, &state)",
        "\tresp.Diagnostics.Append(diags...)",
        "\tif resp.Diagnostics.HasError() {",
        "\t\treturn",
        "\t}",
        "",
        f'\ttflog.Debug(ctx, "Reading {entity}", map[string]any{{"id": state.ID.ValueInt64()}})',
        "",
        f'\tresult, err := r.client.GetByID("{entity}", state.ID.ValueInt64(), nil)',
        "\tif err != nil {",
        "\t\tresp.Diagnostics.AddError(",
        f'\t\t\t"Error reading {entity}",',
        f'\t\t\t"Could not read {entity} ID "+strconv.FormatInt(state.ID.ValueInt64(), 10)+": "+err.Error(),',
        "\t\t)",
        "\t\treturn",
        "\t}",
        "",
        "\tr.mapResultToState(result, &state)",
        "",
        "\tdiags = resp.State.Set(ctx, state)",
        "\tresp.Diagnostics.Append(diags...)",
        "}",
        "",
        f"func (r *{entity}Resource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {{",
        f"\tvar plan {entity}ResourceModel",
        "\tdiags := req.Plan.Get(ctx, &plan)",
        "\tresp.Diagnostics.Append(diags...)",
        "\tif resp.Diagnostics.HasError() {",
        "\t\treturn",
        "\t}",
        f"\tvar state {entity}ResourceModel",
        "\tdiags = req.State.Get(ctx, &state)",
        "\tresp.Diagnostics.Append(diags...)",
        "\tif resp.Diagnostics.HasError() {",
        "\t\treturn",
        "\t}",
        "",
        f'\ttflog.Debug(ctx, "Updating {entity}", map[string]any{{"id": state.ID.ValueInt64()}})',
        "",
    ]
    out += update_values
    out += [
        "",
        f'\tresult, err := r.client.Update("{entity}", state.ID.ValueInt64(), values)',
        "\tif err != nil {",
        "\t\tresp.Diagnostics.AddError(",
        f'\t\t\t"Error updating {entity}",',
        f'\t\t\t"Could not update {entity} ID "+strconv.FormatInt(state.ID.ValueInt64(), 10)+": "+err.Error(),',
        "\t\t)",
        "\t\treturn",
        "\t}",
        "",
        "\tplan.ID = state.ID",
        "\tr.mapResultToState(result, &plan)",
        "",
        f'\ttflog.Debug(ctx, "Updated {entity}", map[string]any{{"id": plan.ID.ValueInt64()}})',
        "",
        "\tdiags = resp.State.Set(ctx, plan)",
        "\tresp.Diagnostics.Append(diags...)",
        "}",
        "",
        f"func (r *{entity}Resource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {{",
        f"\tvar state {entity}ResourceModel",
        "\tdiags := req.State.Get(ctx, &state)",
        "\tresp.Diagnostics.Append(diags...)",
        "\tif resp.Diagnostics.HasError() {",
        "\t\treturn",
        "\t}",
        "",
        f'\ttflog.Debug(ctx, "Deleting {entity}", map[string]any{{"id": state.ID.ValueInt64()}})',
        "",
        f'\terr := r.client.Delete("{entity}", state.ID.ValueInt64())',
        "\tif err != nil {",
        "\t\tresp.Diagnostics.AddError(",
        f'\t\t\t"Error deleting {entity}",',
        f'\t\t\t"Could not delete {entity} ID "+strconv.FormatInt(state.ID.ValueInt64(), 10)+": "+err.Error(),',
        "\t\t)",
        "\t\treturn",
        "\t}",
        "",
        f'\ttflog.Debug(ctx, "Deleted {entity}", map[string]any{{"id": state.ID.ValueInt64()}})',
        "}",
        "",
        f"func (r *{entity}Resource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {{",
        "\tid, err := strconv.ParseInt(req.ID, 10, 64)",
        "\tif err != nil {",
        "\t\tresp.Diagnostics.AddError(",
        '\t\t\t"Invalid import ID",',
        '\t\t\t"Could not parse import ID as integer: "+err.Error(),',
        "\t\t)",
        "\t\treturn",
        "\t}",
        "\tresp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root(\"id\"), id)...)",
        "}",
        "",
    ]
    out += map_lines

    return "\n".join(out) + "\n"

# ─────────────────────────────────────────────────────────────
# Data source generator
# ─────────────────────────────────────────────────────────────

def generate_data_source(entity: str, raw_fields: list[dict]) -> str:
    fields = [classify_field(f) for f in raw_fields if f.get("name") != "id"]
    entity_snake = snake(entity)

    # Check whether entity has a 'name' field (used as second lookup key)
    has_name = any(f["name"] == "name" for f in fields)
    name_field = next((f for f in fields if f["name"] == "name"), None)

    # Schema for all non-id, non-name fields: Computed only
    other_fields = [f for f in fields if f["name"] != "name"]

    imports = [
        '"context"',
        '"fmt"',
        '"strconv"',
        "",
        '"github.com/hashicorp/terraform-plugin-framework/datasource"',
        '"github.com/hashicorp/terraform-plugin-framework/datasource/schema"',
        '"github.com/hashicorp/terraform-plugin-framework/types"',
        '"github.com/hashicorp/terraform-plugin-log/tflog"',
    ]

    # ── struct ────────────────────────────────────────────────
    struct_lines = [
        f"type {entity}DataSourceModel struct {{",
        "\tID types.Int64 `tfsdk:\"id\"`",
    ]
    if has_name:
        struct_lines.append(f'\tName types.String `tfsdk:"name"`')
    for f in other_fields:
        struct_lines.append(f'\t{f["go_field"]} {f["go_type"]} `tfsdk:"{f["name"]}"`')
    struct_lines.append("}")

    # ── schema ────────────────────────────────────────────────
    schema_lines = [
        '\t\t\t"id": schema.Int64Attribute{',
        f'\t\t\t\tDescription: "The unique identifier. Specify either id{" or name" if has_name else ""}.",',
        "\t\t\t\tOptional:    true,",
        "\t\t\t\tComputed:    true,",
        "\t\t\t},",
    ]
    if has_name:
        schema_lines += [
            '\t\t\t"name": schema.StringAttribute{',
            f'\t\t\t\tDescription: "The name of the {entity_snake}. Specify either id or name.",',
            "\t\t\t\tOptional:    true,",
            "\t\t\t\tComputed:    true,",
            "\t\t\t},",
        ]
    for f in other_fields:
        desc = f["description"]
        stype = f["schema_type"]
        schema_lines += [
            f'\t\t\t"{f["name"]}": {stype}{{Description: "{desc}.", Computed: true}},',
        ]

    # ── Read body: where clause ───────────────────────────────
    read_where_lines = []
    if has_name:
        read_where_lines += [
            f'\tif config.ID.IsNull() && config.Name.IsNull() {{',
            '\t\tresp.Diagnostics.AddError("Missing Filter", "At least one of \'id\' or \'name\' must be specified.")',
            "\t\treturn",
            "\t}",
            "",
            "\twhere := [][]any{}",
            "\tif !config.ID.IsNull() {",
            "\t\twhere = append(where, []any{\"id\", \"=\", config.ID.ValueInt64()})",
            "\t}",
            "\tif !config.Name.IsNull() {",
            "\t\twhere = append(where, []any{\"name\", \"=\", config.Name.ValueString()})",
            "\t}",
        ]
    else:
        read_where_lines += [
            "\tif config.ID.IsNull() {",
            '\t\tresp.Diagnostics.AddError("Missing Filter", "\'id\' must be specified.")',
            "\t\treturn",
            "\t}",
            "",
            "\twhere := [][]any{{\"id\", \"=\", config.ID.ValueInt64()}}",
        ]

    # ── Read body: field mapping ──────────────────────────────
    read_map_lines = [
        '\tif id, ok := GetInt64(result, "id"); ok {',
        "\t\tconfig.ID = types.Int64Value(id)",
        "\t}",
    ]
    if has_name:
        read_map_lines += [
            '\tif v, ok := GetString(result, "name"); ok {',
            "\t\tconfig.Name = types.StringValue(v)",
            "\t}",
        ]
    for f in other_fields:
        read_map_lines += _map_result_block(f, "config")

    # ── assemble ──────────────────────────────────────────────
    out = [
        "package provider",
        "",
        "import (",
    ]
    for imp in imports:
        if imp:
            out.append(f"\t{imp}")
        else:
            out.append("")
    out += [
        ")",
        "",
        f"var _ datasource.DataSource = &{entity}DataSource{{}}",
        f"var _ datasource.DataSourceWithConfigure = &{entity}DataSource{{}}",
        "",
        f"type {entity}DataSource struct {{",
        "\tclient *Client",
        "}",
        "",
    ]
    out += struct_lines
    out += [
        "",
        f"func New{entity}DataSource() datasource.DataSource {{",
        f"\treturn &{entity}DataSource{{}}",
        "}",
        "",
        f"func (d *{entity}DataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {{",
        f'\tresp.TypeName = req.ProviderTypeName + "_{entity_snake}"',
        "}",
        "",
        f"func (d *{entity}DataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {{",
        "\tresp.Schema = schema.Schema{",
        f'\t\tDescription: "Fetches a CiviCRM {entity} by ID{" or name" if has_name else ""}.",',
        "\t\tAttributes: map[string]schema.Attribute{",
    ]
    out += schema_lines
    out += [
        "\t\t},",
        "\t}",
        "}",
        "",
        f"func (d *{entity}DataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {{",
        "\tif req.ProviderData == nil {",
        "\t\treturn",
        "\t}",
        "\tclient, ok := req.ProviderData.(*Client)",
        "\tif !ok {",
        '\t\tresp.Diagnostics.AddError("Unexpected Data Source Configure Type",',
        '\t\t\tfmt.Sprintf("Expected *Client, got: %T.", req.ProviderData))',
        "\t\treturn",
        "\t}",
        "\td.client = client",
        "}",
        "",
        f"func (d *{entity}DataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {{",
        f"\tvar config {entity}DataSourceModel",
        "\tresp.Diagnostics.Append(req.Config.Get(ctx, &config)...)",
        "\tif resp.Diagnostics.HasError() {",
        "\t\treturn",
        "\t}",
        "",
    ]
    out += read_where_lines
    out += [
        "",
        f'\ttflog.Debug(ctx, "Reading {entity} data source", map[string]any{{"where": where}})',
        "",
        f'\tresults, err := d.client.Get("{entity}", where, nil)',
        "\tif err != nil {",
        f'\t\tresp.Diagnostics.AddError("Error reading {entity}", err.Error())',
        "\t\treturn",
        "\t}",
        "\tif len(results) == 0 {",
        f'\t\tresp.Diagnostics.AddError("{entity} not found", "No {entity_snake} found matching the specified criteria.")',
        "\t\treturn",
        "\t}",
        "",
        "\tresult := results[0]",
        "",
    ]
    out += read_map_lines
    out += [
        "",
        "\tresp.Diagnostics.Append(resp.State.Set(ctx, config)...)",
        "}",
    ]

    # suppress unused strconv if not needed
    src = "\n".join(out) + "\n"
    if "strconv." not in src:
        src = src.replace('\t"strconv"\n', "")
    return src

# ─────────────────────────────────────────────────────────────
# Docs generation
# ─────────────────────────────────────────────────────────────

def _go_type_label(go_type: str) -> str:
    return {
        "types.Int64":   "Number",
        "types.Float64": "Number",
        "types.Bool":    "Boolean",
        "types.String":  "String",
    }.get(go_type, "String")


def generate_resource_doc(entity: str, raw_fields: list[dict]) -> str:
    fields = [classify_field(f) for f in raw_fields if f.get("name") != "id"]
    entity_snake = snake(entity)
    tf_name = f"civicrm_{entity_snake}"

    required = [f for f in fields if f["is_required"]]
    optional = [f for f in fields if not f["is_required"]]

    # Build a minimal example using only required fields + is_active if present
    example_fields = list(required)
    active = next((f for f in optional if f["name"] == "is_active"), None)
    if active:
        example_fields.append(active)

    example_lines = [f'resource "{tf_name}" "example" {{']
    for f in example_fields:
        go_type = f["go_type"]
        name = f["name"]
        if go_type == "types.String":
            example_lines.append(f'  {name:<30} = "value"')
        elif go_type == "types.Int64":
            example_lines.append(f'  {name:<30} = 1')
        elif go_type == "types.Float64":
            example_lines.append(f'  {name:<30} = 0.00')
        elif go_type == "types.Bool":
            example_lines.append(f'  {name:<30} = true')
    example_lines.append("}")
    example_block = "\n".join(example_lines)

    # Required args
    req_lines = []
    for f in required:
        label = _go_type_label(f["go_type"])
        req_lines.append(f'- `{f["name"]}` ({label}) {f["description"]}.')

    # Optional args
    opt_lines = []
    for f in optional:
        label = _go_type_label(f["go_type"])
        opt_lines.append(f'- `{f["name"]}` ({label}, Optional) {f["description"]}.')

    req_section = "\n".join(req_lines) if req_lines else "_None._"
    opt_section = "\n".join(opt_lines) if opt_lines else "_None._"

    return f"""\
---
page_title: "{tf_name} Resource - CiviCRM"
subcategory: ""
description: |-
  Manages a CiviCRM {entity}.
---

# {tf_name} (Resource)

Manages a CiviCRM {entity}.

## Example Usage

```terraform
{example_block}
```

## Argument Reference

### Required

{req_section}

### Optional

{opt_section}

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

- `id` (Number) The unique identifier of the {entity_snake}.

## Import

{entity} resources can be imported using the numeric ID:

```shell
terraform import {tf_name}.example 42
```
"""


def generate_datasource_doc(entity: str, raw_fields: list[dict]) -> str:
    fields = [classify_field(f) for f in raw_fields if f.get("name") != "id"]
    entity_snake = snake(entity)
    tf_name = f"civicrm_{entity_snake}"

    has_name = any(f["name"] == "name" for f in fields)
    other_fields = [f for f in fields if f["name"] != "name"]

    if has_name:
        example_block = f"""\
# Look up by name
data "{tf_name}" "example" {{
  name = "example"
}}

# Look up by ID
data "{tf_name}" "by_id" {{
  id = 1
}}"""
        filter_note = "At least one of `id` or `name` must be specified."
        filter_args = (
            "- `id` (Number, Optional) The unique identifier.\n"
            "- `name` (String, Optional) The name of the {entity_snake}."
        ).format(entity_snake=entity_snake)
    else:
        example_block = f"""\
data "{tf_name}" "example" {{
  id = 1
}}"""
        filter_note = "`id` must be specified."
        filter_args = f"- `id` (Number, Required) The unique identifier of the {entity_snake}."

    attr_lines = []
    for f in other_fields:
        label = _go_type_label(f["go_type"])
        attr_lines.append(f'- `{f["name"]}` ({label}) {f["description"]}.')
    attrs = "\n".join(attr_lines) if attr_lines else "_None._"

    return f"""\
---
page_title: "{tf_name} Data Source - CiviCRM"
subcategory: ""
description: |-
  Fetches a CiviCRM {entity} by ID{" or name" if has_name else ""}.
---

# {tf_name} (Data Source)

Fetches a CiviCRM {entity} by ID{" or name" if has_name else ""}. {filter_note}

## Example Usage

```terraform
{example_block}
```

## Argument Reference

{filter_args}

## Attributes Reference

In addition to the arguments above, the following attributes are exported:

{attrs}
"""


# ─────────────────────────────────────────────────────────────
# provider.go registration
# ─────────────────────────────────────────────────────────────

SENTINEL_RESOURCES   = "\t\t// END GENERATED RESOURCES"
SENTINEL_DATASOURCES = "\t\t// END GENERATED DATASOURCES"


def _ensure_sentinels(src: str) -> str:
    """Insert sentinel comments if provider.go was written without them."""
    if SENTINEL_RESOURCES not in src:
        # Find the closing brace of the Resources() return slice
        src = src.replace(
            "\t}\n}\n\nfunc (p *CiviCRMProvider) DataSources",
            f"\t\t{SENTINEL_RESOURCES.strip()}\n\t}}\n}}\n\nfunc (p *CiviCRMProvider) DataSources",
            1,
        )
    if SENTINEL_DATASOURCES not in src:
        # Find the closing brace of the DataSources() return slice — last }}
        idx = src.rfind("\t}\n}")
        if idx != -1:
            src = src[:idx] + f"\t\t{SENTINEL_DATASOURCES.strip()}\n\t}}\n}}" + src[idx + len("\t}\n}"):]
    return src


def register_in_provider(entities: list[str], provider_path: Path) -> None:
    if not provider_path.exists():
        print(f"  provider.go not found at {provider_path} — skipping registration", file=sys.stderr)
        return

    src = provider_path.read_text(encoding="utf-8")
    src = _ensure_sentinels(src)

    changed = False

    for entity in entities:
        res_entry  = f"\t\tNew{entity}Resource,"
        ds_entry   = f"\t\tNew{entity}DataSource,"

        if res_entry not in src:
            src = src.replace(
                SENTINEL_RESOURCES,
                f"{res_entry}\n{SENTINEL_RESOURCES}",
            )
            print(f"  Registered New{entity}Resource in Resources()", file=sys.stderr)
            changed = True
        else:
            print(f"  New{entity}Resource already registered — skipped", file=sys.stderr)

        if ds_entry not in src:
            src = src.replace(
                SENTINEL_DATASOURCES,
                f"{ds_entry}\n{SENTINEL_DATASOURCES}",
            )
            print(f"  Registered New{entity}DataSource in DataSources()", file=sys.stderr)
            changed = True
        else:
            print(f"  New{entity}DataSource already registered — skipped", file=sys.stderr)

    if changed:
        provider_path.write_text(src, encoding="utf-8")
        print(f"  Updated: {provider_path}", file=sys.stderr)

# ─────────────────────────────────────────────────────────────
# Main
# ─────────────────────────────────────────────────────────────

def main() -> None:
    parser = argparse.ArgumentParser(
        description="Generate Go Terraform provider resource + data source from CiviCRM getFields",
        formatter_class=argparse.RawDescriptionHelpFormatter,
        epilog=__doc__,
    )
    parser.add_argument("--url", required=True, help="CiviCRM base URL (e.g. https://example.org)")
    parser.add_argument("--api-key", required=True, dest="api_key", help="CiviCRM API key")
    parser.add_argument(
        "--entities", required=True,
        help="Comma-separated CiviCRM entity names (e.g. PriceSet,PriceField,Tag)",
    )
    parser.add_argument(
        "--output-dir", default="internal/provider", dest="output_dir",
        help="Directory to write generated .go files (default: internal/provider)",
    )
    parser.add_argument("--insecure", action="store_true", help="Skip TLS certificate verification")
    parser.add_argument("--dry-run", action="store_true", dest="dry_run",
                        help="Print generated code to stdout instead of writing files")
    parser.add_argument(
        "--provider-go", default="internal/provider/provider.go", dest="provider_go",
        help="Path to provider.go for automatic registration (default: internal/provider/provider.go)",
    )
    parser.add_argument("--no-register", action="store_true", dest="no_register",
                        help="Skip automatic registration in provider.go")
    parser.add_argument(
        "--docs-dir", default=None, dest="docs_dir",
        help="Generate Markdown docs under <docs-dir>/resources/ and <docs-dir>/data-sources/ "
             "(e.g. --docs-dir docs). Skipped if not set.",
    )
    args = parser.parse_args()

    entity_names = [e.strip() for e in args.entities.split(",") if e.strip()]
    if not entity_names:
        parser.error("--entities must contain at least one entity name")

    output_dir = Path(args.output_dir)
    if not args.dry_run:
        output_dir.mkdir(parents=True, exist_ok=True)

    generated: list[str] = []

    for entity in entity_names:
        print(f"Fetching fields for {entity}...", file=sys.stderr)
        raw_fields = fetch_fields(args.url, args.api_key, entity, args.insecure)
        if not raw_fields:
            print(f"  No fields returned — skipping {entity}", file=sys.stderr)
            continue

        print(f"  {len(raw_fields)} fields found", file=sys.stderr)

        resource_code = generate_resource(entity, raw_fields)
        datasource_code = generate_data_source(entity, raw_fields)

        entity_snake = snake(entity)
        res_filename = f"resource_{entity_snake}.go"
        ds_filename = f"data_source_{entity_snake}.go"

        res_doc = generate_resource_doc(entity, raw_fields)
        ds_doc  = generate_datasource_doc(entity, raw_fields)

        if args.dry_run:
            print(f"\n{'='*60}", flush=True)
            print(f"=== {res_filename}", flush=True)
            print(f"{'='*60}", flush=True)
            print(resource_code, flush=True)
            print(f"\n{'='*60}", flush=True)
            print(f"=== {ds_filename}", flush=True)
            print(f"{'='*60}", flush=True)
            print(datasource_code, flush=True)
            if args.docs_dir:
                print(f"\n{'='*60}", flush=True)
                print(f"=== docs/resources/{entity_snake}.md", flush=True)
                print(f"{'='*60}", flush=True)
                print(res_doc, flush=True)
                print(f"\n{'='*60}", flush=True)
                print(f"=== docs/data-sources/{entity_snake}.md", flush=True)
                print(f"{'='*60}", flush=True)
                print(ds_doc, flush=True)
        else:
            res_path = output_dir / res_filename
            ds_path = output_dir / ds_filename
            res_path.write_text(resource_code, encoding="utf-8")
            ds_path.write_text(datasource_code, encoding="utf-8")
            print(f"  Written: {res_path}", file=sys.stderr)
            print(f"  Written: {ds_path}", file=sys.stderr)

            if args.docs_dir:
                docs_root = Path(args.docs_dir)
                res_doc_path = docs_root / "resources" / f"{entity_snake}.md"
                ds_doc_path  = docs_root / "data-sources" / f"{entity_snake}.md"
                res_doc_path.parent.mkdir(parents=True, exist_ok=True)
                ds_doc_path.parent.mkdir(parents=True, exist_ok=True)
                res_doc_path.write_text(res_doc, encoding="utf-8")
                ds_doc_path.write_text(ds_doc, encoding="utf-8")
                print(f"  Written: {res_doc_path}", file=sys.stderr)
                print(f"  Written: {ds_doc_path}", file=sys.stderr)

        generated.append(entity)

    if generated and not args.dry_run:
        if not args.no_register:
            print("\nRegistering in provider.go...", file=sys.stderr)
            register_in_provider(generated, Path(args.provider_go))
        print(
            "\nDon't forget to run 'go build ./...' to check for compile errors.",
            file=sys.stderr,
        )


if __name__ == "__main__":
    main()
