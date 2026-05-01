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
	_ resource.Resource                = &ACLEntityRoleResource{}
	_ resource.ResourceWithConfigure   = &ACLEntityRoleResource{}
	_ resource.ResourceWithImportState = &ACLEntityRoleResource{}
)

// ACLEntityRoleResource manages ACL entity role assignments in CiviCRM.
// This assigns ACL roles to groups, determining which users get which ACL permissions.
type ACLEntityRoleResource struct {
	client *Client
}

type ACLEntityRoleResourceModel struct {
	ID          types.Int64  `tfsdk:"id"`
	ACLRoleID   types.Int64  `tfsdk:"acl_role_id"`
	EntityTable types.String `tfsdk:"entity_table"`
	EntityID    types.Int64  `tfsdk:"entity_id"`
	IsActive    types.Bool   `tfsdk:"is_active"`
}

func NewACLEntityRoleResource() resource.Resource {
	return &ACLEntityRoleResource{}
}

func (r *ACLEntityRoleResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_acl_entity_role"
}

func (r *ACLEntityRoleResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a CiviCRM ACL Entity Role assignment. This assigns ACL roles to groups, " +
			"giving members of the group the permissions defined by the ACL role.",
		Attributes: map[string]schema.Attribute{
			"id": schema.Int64Attribute{
				Description: "The unique identifier of the ACL entity role assignment.",
				Computed:    true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"acl_role_id": schema.Int64Attribute{
				Description: "The value field of the ACL role to assign. Use tonumber(civicrm_acl_role.example.value) to reference an ACL role.",
				Required:    true,
			},
			"entity_table": schema.StringAttribute{
				Description: "The table containing the entity to assign the role to. Default: 'civicrm_group'.",
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString("civicrm_group"),
			},
			"entity_id": schema.Int64Attribute{
				Description: "The ID of the group (or other entity) to assign the ACL role to.",
				Required:    true,
			},
			"is_active": schema.BoolAttribute{
				Description: "Whether this role assignment is active. Default: true.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(true),
			},
		},
	}
}

func (r *ACLEntityRoleResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *ACLEntityRoleResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan ACLEntityRoleResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Creating ACL entity role", map[string]any{
		"acl_role_id":  plan.ACLRoleID.ValueInt64(),
		"entity_table": plan.EntityTable.ValueString(),
		"entity_id":    plan.EntityID.ValueInt64(),
	})

	// Build values for API call
	values := map[string]any{
		"acl_role_id":  plan.ACLRoleID.ValueInt64(),
		"entity_table": plan.EntityTable.ValueString(),
		"entity_id":    plan.EntityID.ValueInt64(),
		"is_active":    plan.IsActive.ValueBool(),
	}

	// Call API
	result, err := r.client.Create("ACLEntityRole", values)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating ACL entity role",
			"Could not create ACL entity role, unexpected error: "+err.Error(),
		)
		return
	}

	// Update state with response
	if id, ok := GetInt64(result, "id"); ok {
		plan.ID = types.Int64Value(id)
	}

	if aclRoleID, ok := GetInt64(result, "acl_role_id"); ok {
		plan.ACLRoleID = types.Int64Value(aclRoleID)
	}

	if entityTable, ok := GetString(result, "entity_table"); ok {
		plan.EntityTable = types.StringValue(entityTable)
	}

	if entityID, ok := GetInt64(result, "entity_id"); ok {
		plan.EntityID = types.Int64Value(entityID)
	}

	if active, ok := GetBool(result, "is_active"); ok {
		plan.IsActive = types.BoolValue(active)
	}

	tflog.Debug(ctx, "Created ACL entity role", map[string]any{
		"id": plan.ID.ValueInt64(),
	})

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

func (r *ACLEntityRoleResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state ACLEntityRoleResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Reading ACL entity role", map[string]any{
		"id": state.ID.ValueInt64(),
	})

	result, err := r.client.GetByID("ACLEntityRole", state.ID.ValueInt64(), nil)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading ACL entity role",
			"Could not read ACL entity role ID "+strconv.FormatInt(state.ID.ValueInt64(), 10)+": "+err.Error(),
		)
		return
	}

	// Update state
	if aclRoleID, ok := GetInt64(result, "acl_role_id"); ok {
		state.ACLRoleID = types.Int64Value(aclRoleID)
	}

	if entityTable, ok := GetString(result, "entity_table"); ok {
		state.EntityTable = types.StringValue(entityTable)
	}

	if entityID, ok := GetInt64(result, "entity_id"); ok {
		state.EntityID = types.Int64Value(entityID)
	}

	if active, ok := GetBool(result, "is_active"); ok {
		state.IsActive = types.BoolValue(active)
	}

	diags = resp.State.Set(ctx, state)
	resp.Diagnostics.Append(diags...)
}

func (r *ACLEntityRoleResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan ACLEntityRoleResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var state ACLEntityRoleResourceModel
	diags = req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Updating ACL entity role", map[string]any{
		"id": state.ID.ValueInt64(),
	})

	// Build values for API call
	values := map[string]any{
		"acl_role_id":  plan.ACLRoleID.ValueInt64(),
		"entity_table": plan.EntityTable.ValueString(),
		"entity_id":    plan.EntityID.ValueInt64(),
		"is_active":    plan.IsActive.ValueBool(),
	}

	// Call API
	result, err := r.client.Update("ACLEntityRole", state.ID.ValueInt64(), values)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error updating ACL entity role",
			"Could not update ACL entity role ID "+strconv.FormatInt(state.ID.ValueInt64(), 10)+": "+err.Error(),
		)
		return
	}

	// Update state
	plan.ID = state.ID

	if aclRoleID, ok := GetInt64(result, "acl_role_id"); ok {
		plan.ACLRoleID = types.Int64Value(aclRoleID)
	}

	if entityTable, ok := GetString(result, "entity_table"); ok {
		plan.EntityTable = types.StringValue(entityTable)
	}

	if entityID, ok := GetInt64(result, "entity_id"); ok {
		plan.EntityID = types.Int64Value(entityID)
	}

	if active, ok := GetBool(result, "is_active"); ok {
		plan.IsActive = types.BoolValue(active)
	}

	tflog.Debug(ctx, "Updated ACL entity role", map[string]any{
		"id": plan.ID.ValueInt64(),
	})

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

func (r *ACLEntityRoleResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state ACLEntityRoleResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Deleting ACL entity role", map[string]any{
		"id": state.ID.ValueInt64(),
	})

	err := r.client.Delete("ACLEntityRole", state.ID.ValueInt64())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error deleting ACL entity role",
			"Could not delete ACL entity role ID "+strconv.FormatInt(state.ID.ValueInt64(), 10)+": "+err.Error(),
		)
		return
	}

	tflog.Debug(ctx, "Deleted ACL entity role", map[string]any{
		"id": state.ID.ValueInt64(),
	})
}

func (r *ACLEntityRoleResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
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
