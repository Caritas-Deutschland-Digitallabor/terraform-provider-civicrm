package provider

import (
	"context"
	"fmt"
	"strconv"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var (
	_ resource.Resource                = &ACLResource{}
	_ resource.ResourceWithConfigure   = &ACLResource{}
	_ resource.ResourceWithImportState = &ACLResource{}
)

// ACLResource manages ACL rules in CiviCRM.
// ACL rules define what operations a role can perform on specific data.
type ACLResource struct {
	client *Client
}

type ACLResourceModel struct {
	ID          types.Int64  `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	Deny        types.Bool   `tfsdk:"deny"`
	EntityTable types.String `tfsdk:"entity_table"`
	EntityID    types.Int64  `tfsdk:"entity_id"`
	Operation   types.String `tfsdk:"operation"`
	ObjectTable types.String `tfsdk:"object_table"`
	ObjectID    types.Int64  `tfsdk:"object_id"`
	AclTable    types.String `tfsdk:"acl_table"`
	AclID       types.Int64  `tfsdk:"acl_id"`
	IsActive    types.Bool   `tfsdk:"is_active"`
	Priority    types.Int64  `tfsdk:"priority"`
}

func NewACLResource() resource.Resource {
	return &ACLResource{}
}

func (r *ACLResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_acl"
}

func (r *ACLResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a CiviCRM ACL rule. ACL rules define what operations a role can perform on specific data.",
		Attributes: map[string]schema.Attribute{
			"id": schema.Int64Attribute{
				Description: "The unique identifier of the ACL.",
				Computed:    true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Description: "The name of the ACL rule.",
				Required:    true,
			},
			"entity_table": schema.StringAttribute{
				Description: "The entity table that owns this ACL (typically 'civicrm_acl_role'). Default: 'civicrm_acl_role'.",
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString("civicrm_acl_role"),
			},
			"entity_id": schema.Int64Attribute{
				Description: "The value field of the ACL role this rule belongs to. Use tonumber(civicrm_acl_role.example.value) to reference an ACL role.",
				Required:    true,
			},
			"operation": schema.StringAttribute{
				Description: "The operation this ACL grants. Options: 'Edit', 'View', 'Create', 'Delete', 'Search', 'All'.",
				Required:    true,
			},
			"object_table": schema.StringAttribute{
				Description: "The type of object being permissioned (e.g., 'civicrm_group', 'civicrm_saved_search', 'civicrm_uf_group').",
				Required:    true,
			},
			"object_id": schema.Int64Attribute{
				Description: "The ID of the specific object being permissioned. Leave empty (null) for all objects of the given type.",
				Optional:    true,
			},
			"is_active": schema.BoolAttribute{
				Description: "Whether the ACL rule is active. Default: true.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(true),
			},
			"deny": schema.BoolAttribute{
				Description: "Whether this ACL denies rather than allows access. Default: false.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
			},
			"acl_table": schema.StringAttribute{
				Description: "The ACL table for nested/linked ACLs.",
				Optional:    true,
			},
			"acl_id": schema.Int64Attribute{
				Description: "The ID of a linked ACL rule.",
				Optional:    true,
			},
			"priority": schema.Int64Attribute{
				Description: "The priority of the ACL rule (higher priority rules are evaluated first).",
				Optional:    true,
				Computed:    true,
			},
		},
	}
}

func (r *ACLResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *ACLResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan ACLResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Creating ACL", map[string]any{
		"name":      plan.Name.ValueString(),
		"operation": plan.Operation.ValueString(),
	})

	// Build values for API call
	values := map[string]any{
		"name":         plan.Name.ValueString(),
		"entity_table": plan.EntityTable.ValueString(),
		"entity_id":    plan.EntityID.ValueInt64(),
		"operation":    plan.Operation.ValueString(),
		"object_table": plan.ObjectTable.ValueString(),
		"is_active":    plan.IsActive.ValueBool(),
		"deny":         plan.Deny.ValueBool(),
	}

	if !plan.ObjectID.IsNull() {
		values["object_id"] = plan.ObjectID.ValueInt64()
	}

	if !plan.AclTable.IsNull() {
		values["acl_table"] = plan.AclTable.ValueString()
	}

	if !plan.AclID.IsNull() {
		values["acl_id"] = plan.AclID.ValueInt64()
	}

	if !plan.Priority.IsNull() {
		values["priority"] = plan.Priority.ValueInt64()
	}

	// Call API
	result, err := r.client.Create("ACL", values)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating ACL",
			"Could not create ACL, unexpected error: "+err.Error(),
		)
		return
	}

	// Update state with response
	if id, ok := GetInt64(result, "id"); ok {
		plan.ID = types.Int64Value(id)
	}

	if name, ok := GetString(result, "name"); ok {
		plan.Name = types.StringValue(name)
	}

	if entityTable, ok := GetString(result, "entity_table"); ok {
		plan.EntityTable = types.StringValue(entityTable)
	}

	if entityID, ok := GetInt64(result, "entity_id"); ok {
		plan.EntityID = types.Int64Value(entityID)
	}

	if operation, ok := GetString(result, "operation"); ok {
		plan.Operation = types.StringValue(operation)
	}

	if objectTable, ok := GetString(result, "object_table"); ok {
		plan.ObjectTable = types.StringValue(objectTable)
	}

	if objectID, ok := GetInt64(result, "object_id"); ok {
		plan.ObjectID = types.Int64Value(objectID)
	} else {
		plan.ObjectID = types.Int64Null()
	}

	if aclTable, ok := GetString(result, "acl_table"); ok && aclTable != "" {
		plan.AclTable = types.StringValue(aclTable)
	} else {
		plan.AclTable = types.StringNull()
	}

	if aclID, ok := GetInt64(result, "acl_id"); ok {
		plan.AclID = types.Int64Value(aclID)
	} else {
		plan.AclID = types.Int64Null()
	}

	if active, ok := GetBool(result, "is_active"); ok {
		plan.IsActive = types.BoolValue(active)
	}

	if deny, ok := GetBool(result, "deny"); ok {
		plan.Deny = types.BoolValue(deny)
	}

	if priority, ok := GetInt64(result, "priority"); ok {
		plan.Priority = types.Int64Value(priority)
	}

	tflog.Debug(ctx, "Created ACL", map[string]any{
		"id": plan.ID.ValueInt64(),
	})

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

func (r *ACLResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state ACLResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Reading ACL", map[string]any{
		"id": state.ID.ValueInt64(),
	})

	result, err := r.client.GetByID("ACL", state.ID.ValueInt64(), nil)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading ACL",
			"Could not read ACL ID "+strconv.FormatInt(state.ID.ValueInt64(), 10)+": "+err.Error(),
		)
		return
	}

	// Update state
	if name, ok := GetString(result, "name"); ok {
		state.Name = types.StringValue(name)
	}

	if entityTable, ok := GetString(result, "entity_table"); ok {
		state.EntityTable = types.StringValue(entityTable)
	}

	if entityID, ok := GetInt64(result, "entity_id"); ok {
		state.EntityID = types.Int64Value(entityID)
	}

	if operation, ok := GetString(result, "operation"); ok {
		state.Operation = types.StringValue(operation)
	}

	if objectTable, ok := GetString(result, "object_table"); ok {
		state.ObjectTable = types.StringValue(objectTable)
	}

	if objectID, ok := GetInt64(result, "object_id"); ok {
		state.ObjectID = types.Int64Value(objectID)
	} else {
		state.ObjectID = types.Int64Null()
	}

	if aclTable, ok := GetString(result, "acl_table"); ok && aclTable != "" {
		state.AclTable = types.StringValue(aclTable)
	} else {
		state.AclTable = types.StringNull()
	}

	if aclID, ok := GetInt64(result, "acl_id"); ok {
		state.AclID = types.Int64Value(aclID)
	} else {
		state.AclID = types.Int64Null()
	}

	if active, ok := GetBool(result, "is_active"); ok {
		state.IsActive = types.BoolValue(active)
	}

	if deny, ok := GetBool(result, "deny"); ok {
		state.Deny = types.BoolValue(deny)
	}

	if priority, ok := GetInt64(result, "priority"); ok {
		state.Priority = types.Int64Value(priority)
	}

	diags = resp.State.Set(ctx, state)
	resp.Diagnostics.Append(diags...)
}

func (r *ACLResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan ACLResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var state ACLResourceModel
	diags = req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Updating ACL", map[string]any{
		"id": state.ID.ValueInt64(),
	})

	// Build values for API call
	values := map[string]any{
		"name":         plan.Name.ValueString(),
		"entity_table": plan.EntityTable.ValueString(),
		"entity_id":    plan.EntityID.ValueInt64(),
		"operation":    plan.Operation.ValueString(),
		"object_table": plan.ObjectTable.ValueString(),
		"is_active":    plan.IsActive.ValueBool(),
		"deny":         plan.Deny.ValueBool(),
	}

	if !plan.ObjectID.IsNull() {
		values["object_id"] = plan.ObjectID.ValueInt64()
	} else {
		values["object_id"] = nil
	}

	if !plan.AclTable.IsNull() {
		values["acl_table"] = plan.AclTable.ValueString()
	} else {
		values["acl_table"] = nil
	}

	if !plan.AclID.IsNull() {
		values["acl_id"] = plan.AclID.ValueInt64()
	} else {
		values["acl_id"] = nil
	}

	if !plan.Priority.IsNull() {
		values["priority"] = plan.Priority.ValueInt64()
	} else {
		values["priority"] = nil
	}

	// Call API
	result, err := r.client.Update("ACL", state.ID.ValueInt64(), values)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error updating ACL",
			"Could not update ACL ID "+strconv.FormatInt(state.ID.ValueInt64(), 10)+": "+err.Error(),
		)
		return
	}

	// Update state
	plan.ID = state.ID

	if name, ok := GetString(result, "name"); ok {
		plan.Name = types.StringValue(name)
	}

	if entityTable, ok := GetString(result, "entity_table"); ok {
		plan.EntityTable = types.StringValue(entityTable)
	}

	if entityID, ok := GetInt64(result, "entity_id"); ok {
		plan.EntityID = types.Int64Value(entityID)
	}

	if operation, ok := GetString(result, "operation"); ok {
		plan.Operation = types.StringValue(operation)
	}

	if objectTable, ok := GetString(result, "object_table"); ok {
		plan.ObjectTable = types.StringValue(objectTable)
	}

	if objectID, ok := GetInt64(result, "object_id"); ok {
		plan.ObjectID = types.Int64Value(objectID)
	} else {
		plan.ObjectID = types.Int64Null()
	}

	if aclTable, ok := GetString(result, "acl_table"); ok && aclTable != "" {
		plan.AclTable = types.StringValue(aclTable)
	} else {
		plan.AclTable = types.StringNull()
	}

	if aclID, ok := GetInt64(result, "acl_id"); ok {
		plan.AclID = types.Int64Value(aclID)
	} else {
		plan.AclID = types.Int64Null()
	}

	if active, ok := GetBool(result, "is_active"); ok {
		plan.IsActive = types.BoolValue(active)
	}

	if deny, ok := GetBool(result, "deny"); ok {
		plan.Deny = types.BoolValue(deny)
	}

	if priority, ok := GetInt64(result, "priority"); ok {
		plan.Priority = types.Int64Value(priority)
	}

	tflog.Debug(ctx, "Updated ACL", map[string]any{
		"id": plan.ID.ValueInt64(),
	})

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

func (r *ACLResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state ACLResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Deleting ACL", map[string]any{
		"id": state.ID.ValueInt64(),
	})

	err := r.client.Delete("ACL", state.ID.ValueInt64())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error deleting ACL",
			"Could not delete ACL ID "+strconv.FormatInt(state.ID.ValueInt64(), 10)+": "+err.Error(),
		)
		return
	}

	tflog.Debug(ctx, "Deleted ACL", map[string]any{
		"id": state.ID.ValueInt64(),
	})
}

func (r *ACLResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
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
