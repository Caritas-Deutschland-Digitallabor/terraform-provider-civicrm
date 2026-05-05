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
	_ resource.Resource                = &CaseStatusResource{}
	_ resource.ResourceWithConfigure   = &CaseStatusResource{}
	_ resource.ResourceWithImportState = &CaseStatusResource{}
)

// CaseStatusResource manages CiviCRM case statuses.
// Case statuses are OptionValues in the "case_status" OptionGroup.
// The grouping field controls which UI category the status appears in:
//   "Opened"   — open/active cases
//   "Closed"   — resolved/closed cases
type CaseStatusResource struct {
	client *Client
}

type CaseStatusResourceModel struct {
	ID       types.Int64  `tfsdk:"id"`
	Name     types.String `tfsdk:"name"`
	Label    types.String `tfsdk:"label"`
	Grouping types.String `tfsdk:"grouping"`
	IsActive types.Bool   `tfsdk:"is_active"`
	IsReserved types.Bool `tfsdk:"is_reserved"`
	Weight   types.Int64  `tfsdk:"weight"`
	Value    types.String `tfsdk:"value"`
}

func NewCaseStatusResource() resource.Resource {
	return &CaseStatusResource{}
}

func (r *CaseStatusResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_case_status"
}

func (r *CaseStatusResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a CiviCRM Case Status. Case statuses are OptionValues in the 'case_status' option group. They control which lifecycle stage a case is in and how it appears in reports and searches.",
		Attributes: map[string]schema.Attribute{
			"id": schema.Int64Attribute{
				Description: "Unique ID of the OptionValue.",
				Computed:    true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Description: "Machine name of the status (e.g. 'in_progress', 'housing_secured'). Must be unique within the case_status group.",
				Required:    true,
			},
			"label": schema.StringAttribute{
				Description: "Display label shown in the UI (e.g. 'In Progress', 'Housing Secured').",
				Required:    true,
			},
			"grouping": schema.StringAttribute{
				Description: "Category bucket for the status. Use 'Opened' for active/open cases and 'Closed' for resolved/ended cases. CiviCRM uses this to filter cases in dashboards and reports.",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"is_active": schema.BoolAttribute{
				Description: "Whether the status is active. Default: true.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(true),
			},
			"is_reserved": schema.BoolAttribute{
				Description: "Whether the status is reserved (protected from deletion by CiviCRM). Default: false.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
			},
			"weight": schema.Int64Attribute{
				Description: "Sort weight. Controls display order in dropdowns.",
				Optional:    true,
				Computed:    true,
			},
			"value": schema.StringAttribute{
				Description: "Internal value used by CiviCRM to identify this status in case activity records. Auto-generated if not set.",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

func (r *CaseStatusResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	client, ok := req.ProviderData.(*Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *Client, got: %T.", req.ProviderData),
		)
		return
	}
	r.client = client
}

func (r *CaseStatusResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan CaseStatusResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Creating CaseStatus", map[string]any{"name": plan.Name.ValueString()})

	optionGroupID, err := r.client.GetOptionGroupID("case_status")
	if err != nil {
		resp.Diagnostics.AddError("Error looking up option group", "Could not find case_status option group: "+err.Error())
		return
	}

	values := map[string]any{
		"option_group_id": optionGroupID,
		"name":            plan.Name.ValueString(),
		"label":           plan.Label.ValueString(),
		"is_active":       plan.IsActive.ValueBool(),
		"is_reserved":     plan.IsReserved.ValueBool(),
	}
	if !plan.Grouping.IsNull() && plan.Grouping.ValueString() != "" {
		values["grouping"] = plan.Grouping.ValueString()
	}
	if !plan.Weight.IsNull() {
		values["weight"] = plan.Weight.ValueInt64()
	}
	if !plan.Value.IsNull() && plan.Value.ValueString() != "" {
		values["value"] = plan.Value.ValueString()
	}

	result, err := r.client.Create("OptionValue", values)
	if err != nil {
		resp.Diagnostics.AddError("Error creating CaseStatus", err.Error())
		return
	}

	r.mapResultToState(result, &plan)
	tflog.Debug(ctx, "Created CaseStatus", map[string]any{"id": plan.ID.ValueInt64()})
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *CaseStatusResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state CaseStatusResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	result, err := r.client.GetByID("OptionValue", state.ID.ValueInt64(), nil)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading CaseStatus",
			"Could not read CaseStatus ID "+strconv.FormatInt(state.ID.ValueInt64(), 10)+": "+err.Error(),
		)
		return
	}

	r.mapResultToState(result, &state)
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (r *CaseStatusResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan CaseStatusResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	var state CaseStatusResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	values := map[string]any{
		"name":        plan.Name.ValueString(),
		"label":       plan.Label.ValueString(),
		"is_active":   plan.IsActive.ValueBool(),
		"is_reserved": plan.IsReserved.ValueBool(),
	}
	if !plan.Grouping.IsNull() && plan.Grouping.ValueString() != "" {
		values["grouping"] = plan.Grouping.ValueString()
	} else {
		values["grouping"] = nil
	}
	if !plan.Weight.IsNull() {
		values["weight"] = plan.Weight.ValueInt64()
	}
	if !plan.Value.IsNull() && plan.Value.ValueString() != "" {
		values["value"] = plan.Value.ValueString()
	}

	result, err := r.client.Update("OptionValue", state.ID.ValueInt64(), values)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error updating CaseStatus",
			"Could not update CaseStatus ID "+strconv.FormatInt(state.ID.ValueInt64(), 10)+": "+err.Error(),
		)
		return
	}

	plan.ID = state.ID
	r.mapResultToState(result, &plan)
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *CaseStatusResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state CaseStatusResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.client.Delete("OptionValue", state.ID.ValueInt64()); err != nil {
		resp.Diagnostics.AddError(
			"Error deleting CaseStatus",
			"Could not delete CaseStatus ID "+strconv.FormatInt(state.ID.ValueInt64(), 10)+": "+err.Error(),
		)
	}
}

func (r *CaseStatusResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	id, err := strconv.ParseInt(req.ID, 10, 64)
	if err != nil {
		resp.Diagnostics.AddError("Invalid import ID", "Could not parse import ID as integer: "+err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), id)...)
}

func (r *CaseStatusResource) mapResultToState(result map[string]any, model *CaseStatusResourceModel) {
	if id, ok := GetInt64(result, "id"); ok {
		model.ID = types.Int64Value(id)
	}
	if name, ok := GetString(result, "name"); ok {
		model.Name = types.StringValue(name)
	}
	if label, ok := GetString(result, "label"); ok {
		model.Label = types.StringValue(label)
	}
	if grouping, ok := GetString(result, "grouping"); ok && grouping != "" {
		model.Grouping = types.StringValue(grouping)
	} else {
		model.Grouping = types.StringNull()
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
	if value, ok := GetString(result, "value"); ok && value != "" {
		model.Value = types.StringValue(value)
	} else {
		model.Value = types.StringNull()
	}
}
