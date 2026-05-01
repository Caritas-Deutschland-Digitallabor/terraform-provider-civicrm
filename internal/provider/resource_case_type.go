package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var (
	_ resource.Resource                = &CaseTypeResource{}
	_ resource.ResourceWithConfigure   = &CaseTypeResource{}
	_ resource.ResourceWithImportState = &CaseTypeResource{}
)

type CaseTypeResource struct {
	client *Client
}

type CaseTypeResourceModel struct {
	ID          types.Int64  `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	Title       types.String `tfsdk:"title"`
	Description types.String `tfsdk:"description"`
	IsActive    types.Bool   `tfsdk:"is_active"`
	IsReserved  types.Bool   `tfsdk:"is_reserved"`
	Weight      types.Int64  `tfsdk:"weight"`
	Definition  types.String `tfsdk:"definition"`
}

func NewCaseTypeResource() resource.Resource {
	return &CaseTypeResource{}
}

func (r *CaseTypeResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_case_type"
}

func (r *CaseTypeResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a CiviCRM Case Type. Case Types define the structure of cases including activity types, timelines, and case roles.",
		Attributes: map[string]schema.Attribute{
			"id": schema.Int64Attribute{
				Description: "The unique identifier of the case type.",
				Computed:    true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Description: "The machine name of the case type (must be unique).",
				Required:    true,
			},
			"title": schema.StringAttribute{
				Description: "The display title of the case type.",
				Required:    true,
			},
			"description": schema.StringAttribute{
				Description: "A description of the case type.",
				Optional:    true,
			},
			"is_active": schema.BoolAttribute{
				Description: "Whether the case type is active. Default: true.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(true),
			},
			"is_reserved": schema.BoolAttribute{
				Description: "Whether the case type is reserved (system-defined). Default: false.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
			},
			"weight": schema.Int64Attribute{
				Description: "The sort weight of the case type.",
				Optional:    true,
				Computed:    true,
			},
			"definition": schema.StringAttribute{
				Description: "The case type definition as a JSON string. Defines activity types, activity sets, timeline activity types, and case roles. On read, the API returns a structured object which is serialized to JSON.",
				Optional:    true,
				Computed:    true,
			},
		},
	}
}

func (r *CaseTypeResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}

	r.client = client
}

func (r *CaseTypeResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan CaseTypeResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Creating CaseType", map[string]any{
		"name":  plan.Name.ValueString(),
		"title": plan.Title.ValueString(),
	})

	values := map[string]any{
		"name":        plan.Name.ValueString(),
		"title":       plan.Title.ValueString(),
		"is_active":   plan.IsActive.ValueBool(),
		"is_reserved": plan.IsReserved.ValueBool(),
	}

	if !plan.Description.IsNull() {
		values["description"] = plan.Description.ValueString()
	}

	if !plan.Weight.IsNull() {
		values["weight"] = plan.Weight.ValueInt64()
	}

	if !plan.Definition.IsNull() && plan.Definition.ValueString() != "" {
		var def any
		if err := json.Unmarshal([]byte(plan.Definition.ValueString()), &def); err != nil {
			resp.Diagnostics.AddError(
				"Invalid definition JSON",
				"The definition field must be valid JSON: "+err.Error(),
			)
			return
		}
		values["definition"] = def
	}

	result, err := r.client.Create("CaseType", values)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating CaseType",
			"Could not create CaseType: "+err.Error(),
		)
		return
	}

	r.mapResultToState(result, &plan)

	tflog.Debug(ctx, "Created CaseType", map[string]any{"id": plan.ID.ValueInt64()})

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

func (r *CaseTypeResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state CaseTypeResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Reading CaseType", map[string]any{"id": state.ID.ValueInt64()})

	result, err := r.client.GetByID("CaseType", state.ID.ValueInt64(), nil)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading CaseType",
			"Could not read CaseType ID "+strconv.FormatInt(state.ID.ValueInt64(), 10)+": "+err.Error(),
		)
		return
	}

	r.mapResultToState(result, &state)

	diags = resp.State.Set(ctx, state)
	resp.Diagnostics.Append(diags...)
}

func (r *CaseTypeResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan CaseTypeResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var state CaseTypeResourceModel
	diags = req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Updating CaseType", map[string]any{"id": state.ID.ValueInt64()})

	values := map[string]any{
		"name":        plan.Name.ValueString(),
		"title":       plan.Title.ValueString(),
		"is_active":   plan.IsActive.ValueBool(),
		"is_reserved": plan.IsReserved.ValueBool(),
	}

	if !plan.Description.IsNull() {
		values["description"] = plan.Description.ValueString()
	} else {
		values["description"] = nil
	}

	if !plan.Weight.IsNull() {
		values["weight"] = plan.Weight.ValueInt64()
	}

	if !plan.Definition.IsNull() && plan.Definition.ValueString() != "" {
		var def any
		if err := json.Unmarshal([]byte(plan.Definition.ValueString()), &def); err != nil {
			resp.Diagnostics.AddError(
				"Invalid definition JSON",
				"The definition field must be valid JSON: "+err.Error(),
			)
			return
		}
		values["definition"] = def
	}

	result, err := r.client.Update("CaseType", state.ID.ValueInt64(), values)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error updating CaseType",
			"Could not update CaseType ID "+strconv.FormatInt(state.ID.ValueInt64(), 10)+": "+err.Error(),
		)
		return
	}

	plan.ID = state.ID
	r.mapResultToState(result, &plan)

	tflog.Debug(ctx, "Updated CaseType", map[string]any{"id": plan.ID.ValueInt64()})

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

func (r *CaseTypeResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state CaseTypeResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Deleting CaseType", map[string]any{"id": state.ID.ValueInt64()})

	err := r.client.Delete("CaseType", state.ID.ValueInt64())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error deleting CaseType",
			"Could not delete CaseType ID "+strconv.FormatInt(state.ID.ValueInt64(), 10)+": "+err.Error(),
		)
		return
	}

	tflog.Debug(ctx, "Deleted CaseType", map[string]any{"id": state.ID.ValueInt64()})
}

func (r *CaseTypeResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	id, err := strconv.ParseInt(req.ID, 10, 64)
	if err != nil {
		resp.Diagnostics.AddError(
			"Invalid import ID",
			"Could not parse import ID as integer: "+err.Error(),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), id)...)
}

func (r *CaseTypeResource) mapResultToState(result map[string]any, model *CaseTypeResourceModel) {
	if id, ok := GetInt64(result, "id"); ok {
		model.ID = types.Int64Value(id)
	}

	if name, ok := GetString(result, "name"); ok {
		model.Name = types.StringValue(name)
	}

	if title, ok := GetString(result, "title"); ok {
		model.Title = types.StringValue(title)
	}

	if desc, ok := GetString(result, "description"); ok && desc != "" {
		model.Description = types.StringValue(desc)
	} else {
		model.Description = types.StringNull()
	}

	if active, ok := GetBool(result, "is_active"); ok {
		model.IsActive = types.BoolValue(active)
	}

	if reserved, ok := GetBool(result, "is_reserved"); ok {
		model.IsReserved = types.BoolValue(reserved)
	}

	if weight, ok := GetInt64(result, "weight"); ok {
		model.Weight = types.Int64Value(weight)
	}

	if def, ok := result["definition"]; ok && def != nil {
		switch v := def.(type) {
		case string:
			if v != "" {
				model.Definition = types.StringValue(v)
			} else {
				model.Definition = types.StringNull()
			}
		default:
			// API returned a structured object — serialize to JSON for state
			jsonBytes, err := json.Marshal(v)
			if err == nil {
				model.Definition = types.StringValue(string(jsonBytes))
			} else {
				model.Definition = types.StringNull()
			}
		}
	} else {
		model.Definition = types.StringNull()
	}
}
