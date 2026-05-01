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
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var (
	_ resource.Resource                = &ACLRoleResource{}
	_ resource.ResourceWithConfigure   = &ACLRoleResource{}
	_ resource.ResourceWithImportState = &ACLRoleResource{}
)

// ACLRoleResource manages ACL roles in CiviCRM.
// ACL Roles are stored as OptionValues in the "acl_role" option group.
type ACLRoleResource struct {
	client *Client
}

type ACLRoleResourceModel struct {
	ID          types.Int64  `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	Label       types.String `tfsdk:"label"`
	Description types.String `tfsdk:"description"`
	IsActive    types.Bool   `tfsdk:"is_active"`
	Weight      types.Int64  `tfsdk:"weight"`
	Value       types.String `tfsdk:"value"`
}

func NewACLRoleResource() resource.Resource {
	return &ACLRoleResource{}
}

func (r *ACLRoleResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_acl_role"
}

func (r *ACLRoleResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a CiviCRM ACL Role. ACL Roles define permission sets that can be assigned to groups via ACL Entity Roles.",
		Attributes: map[string]schema.Attribute{
			"id": schema.Int64Attribute{
				Description: "The unique identifier of the ACL role (OptionValue ID).",
				Computed:    true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Description: "The machine name of the ACL role.",
				Required:    true,
			},
			"label": schema.StringAttribute{
				Description: "The display label of the ACL role.",
				Required:    true,
			},
			"description": schema.StringAttribute{
				Description: "A description of the ACL role.",
				Optional:    true,
			},
			"is_active": schema.BoolAttribute{
				Description: "Whether the ACL role is active. Default: true.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(true),
			},
			"weight": schema.Int64Attribute{
				Description: "The sort weight of the ACL role.",
				Optional:    true,
				Computed:    true,
			},
			"value": schema.StringAttribute{
				Description: "The value of the ACL role (used internally by CiviCRM to link ACL rules and entity roles). If not specified, CiviCRM auto-generates it. Use this value when referencing the role in civicrm_acl and civicrm_acl_entity_role resources.",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

func (r *ACLRoleResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *ACLRoleResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan ACLRoleResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Creating ACL role", map[string]any{
		"name":  plan.Name.ValueString(),
		"label": plan.Label.ValueString(),
	})

	// Look up the acl_role option group ID
	optionGroupID, err := r.client.GetOptionGroupID("acl_role")
	if err != nil {
		resp.Diagnostics.AddError(
			"Error looking up option group",
			"Could not find acl_role option group: "+err.Error(),
		)
		return
	}

	// Build values for API call
	// ACL Roles are stored as OptionValues in the acl_role option group
	values := map[string]any{
		"option_group_id": optionGroupID,
		"name":            plan.Name.ValueString(),
		"label":           plan.Label.ValueString(),
		"is_active":       plan.IsActive.ValueBool(),
	}

	if !plan.Description.IsNull() {
		values["description"] = plan.Description.ValueString()
	}

	if !plan.Weight.IsNull() {
		values["weight"] = plan.Weight.ValueInt64()
	}

	if !plan.Value.IsNull() {
		values["value"] = plan.Value.ValueString()
	}

	// Call API
	result, err := r.client.Create("OptionValue", values)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating ACL role",
			"Could not create ACL role, unexpected error: "+err.Error(),
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

	if label, ok := GetString(result, "label"); ok {
		plan.Label = types.StringValue(label)
	}

	if desc, ok := GetString(result, "description"); ok {
		plan.Description = types.StringValue(desc)
	} else {
		plan.Description = types.StringNull()
	}

	if active, ok := GetBool(result, "is_active"); ok {
		plan.IsActive = types.BoolValue(active)
	}

	if weight, ok := GetInt64(result, "weight"); ok {
		plan.Weight = types.Int64Value(weight)
	}

	if value, ok := GetString(result, "value"); ok {
		plan.Value = types.StringValue(value)
	}

	tflog.Debug(ctx, "Created ACL role", map[string]any{
		"id": plan.ID.ValueInt64(),
	})

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

func (r *ACLRoleResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state ACLRoleResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Reading ACL role", map[string]any{
		"id": state.ID.ValueInt64(),
	})

	result, err := r.client.GetByID("OptionValue", state.ID.ValueInt64(), nil)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading ACL role",
			"Could not read ACL role ID "+strconv.FormatInt(state.ID.ValueInt64(), 10)+": "+err.Error(),
		)
		return
	}

	// Update state
	if name, ok := GetString(result, "name"); ok {
		state.Name = types.StringValue(name)
	}

	if label, ok := GetString(result, "label"); ok {
		state.Label = types.StringValue(label)
	}

	if desc, ok := GetString(result, "description"); ok && desc != "" {
		state.Description = types.StringValue(desc)
	} else {
		state.Description = types.StringNull()
	}

	if active, ok := GetBool(result, "is_active"); ok {
		state.IsActive = types.BoolValue(active)
	}

	if weight, ok := GetInt64(result, "weight"); ok {
		state.Weight = types.Int64Value(weight)
	}

	if value, ok := GetString(result, "value"); ok {
		state.Value = types.StringValue(value)
	}

	diags = resp.State.Set(ctx, state)
	resp.Diagnostics.Append(diags...)
}

func (r *ACLRoleResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan ACLRoleResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var state ACLRoleResourceModel
	diags = req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Updating ACL role", map[string]any{
		"id": state.ID.ValueInt64(),
	})

	// Build values for API call
	values := map[string]any{
		"name":      plan.Name.ValueString(),
		"label":     plan.Label.ValueString(),
		"is_active": plan.IsActive.ValueBool(),
	}

	if !plan.Description.IsNull() {
		values["description"] = plan.Description.ValueString()
	} else {
		values["description"] = nil
	}

	if !plan.Weight.IsNull() {
		values["weight"] = plan.Weight.ValueInt64()
	}

	if !plan.Value.IsNull() {
		values["value"] = plan.Value.ValueString()
	}

	// Call API
	result, err := r.client.Update("OptionValue", state.ID.ValueInt64(), values)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error updating ACL role",
			"Could not update ACL role ID "+strconv.FormatInt(state.ID.ValueInt64(), 10)+": "+err.Error(),
		)
		return
	}

	// Update state
	plan.ID = state.ID

	if name, ok := GetString(result, "name"); ok {
		plan.Name = types.StringValue(name)
	}

	if label, ok := GetString(result, "label"); ok {
		plan.Label = types.StringValue(label)
	}

	if desc, ok := GetString(result, "description"); ok && desc != "" {
		plan.Description = types.StringValue(desc)
	} else {
		plan.Description = types.StringNull()
	}

	if active, ok := GetBool(result, "is_active"); ok {
		plan.IsActive = types.BoolValue(active)
	}

	if weight, ok := GetInt64(result, "weight"); ok {
		plan.Weight = types.Int64Value(weight)
	}

	if value, ok := GetString(result, "value"); ok {
		plan.Value = types.StringValue(value)
	}

	tflog.Debug(ctx, "Updated ACL role", map[string]any{
		"id": plan.ID.ValueInt64(),
	})

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

func (r *ACLRoleResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state ACLRoleResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Deleting ACL role", map[string]any{
		"id": state.ID.ValueInt64(),
	})

	err := r.client.Delete("OptionValue", state.ID.ValueInt64())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error deleting ACL role",
			"Could not delete ACL role ID "+strconv.FormatInt(state.ID.ValueInt64(), 10)+": "+err.Error(),
		)
		return
	}

	tflog.Debug(ctx, "Deleted ACL role", map[string]any{
		"id": state.ID.ValueInt64(),
	})
}

func (r *ACLRoleResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
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
