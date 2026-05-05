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
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var (
	_ resource.Resource                = &CiviRulesRuleResource{}
	_ resource.ResourceWithConfigure   = &CiviRulesRuleResource{}
	_ resource.ResourceWithImportState = &CiviRulesRuleResource{}
)

// CiviRulesRuleResource manages a CiviRules rule (entity: CiviRulesRule).
// A rule pairs one trigger with optional conditions and actions.
// Conditions are managed via civicrm_civirules_condition,
// actions via civicrm_civirules_action.
type CiviRulesRuleResource struct {
	client *Client
}

type CiviRulesRuleResourceModel struct {
	ID             types.Int64  `tfsdk:"id"`
	Name           types.String `tfsdk:"name"`
	Label          types.String `tfsdk:"label"`
	TriggerID      types.Int64  `tfsdk:"trigger_id"`
	TriggerParams  types.String `tfsdk:"trigger_params"`
	Description    types.String `tfsdk:"description"`
	HelpText       types.String `tfsdk:"help_text"`
	IsActive       types.Bool   `tfsdk:"is_active"`
	IsDebug        types.Bool   `tfsdk:"is_debug"`
}

func NewCiviRulesRuleResource() resource.Resource {
	return &CiviRulesRuleResource{}
}

func (r *CiviRulesRuleResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_civirules_rule"
}

func (r *CiviRulesRuleResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a CiviRules Rule. A rule consists of a trigger, optional conditions (civicrm_civirules_condition), and one or more actions (civicrm_civirules_action). Requires the CiviRules extension to be installed.",
		Attributes: map[string]schema.Attribute{
			"id": schema.Int64Attribute{
				Description: "Unique ID of the CiviRules rule.",
				Computed:    true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Description: "Machine name of the rule. Must be unique. Used for managed entity matching.",
				Required:    true,
			},
			"label": schema.StringAttribute{
				Description: "Human-readable label shown in the CiviRules UI.",
				Required:    true,
			},
			"trigger_id": schema.Int64Attribute{
				Description: "ID of the CiviRules trigger that fires this rule (e.g. 'Case is changed', 'Activity is added'). Look up available triggers via the CiviRules UI or CiviRulesTrigger.get API.",
				Required:    true,
			},
			"trigger_params": schema.StringAttribute{
				Description: "JSON-encoded parameters for the trigger. Content depends on the trigger type. For case triggers this typically contains the case_type_id. Leave empty if the trigger requires no parameters.",
				Optional:    true,
			},
			"description": schema.StringAttribute{
				Description: "Optional description of what this rule does.",
				Optional:    true,
			},
			"help_text": schema.StringAttribute{
				Description: "Optional help text shown to admins in the CiviRules UI.",
				Optional:    true,
			},
			"is_active": schema.BoolAttribute{
				Description: "Whether the rule is active. Default: true.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(true),
			},
			"is_debug": schema.BoolAttribute{
				Description: "Enable debug logging for this rule. Default: false.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
			},
		},
	}
}

func (r *CiviRulesRuleResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	client, ok := req.ProviderData.(*Client)
	if !ok {
		resp.Diagnostics.AddError("Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *Client, got: %T.", req.ProviderData))
		return
	}
	r.client = client
}

func (r *CiviRulesRuleResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan CiviRulesRuleResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Creating CiviRulesRule", map[string]any{"name": plan.Name.ValueString()})

	values := map[string]any{
		"name":      plan.Name.ValueString(),
		"label":     plan.Label.ValueString(),
		"trigger_id": plan.TriggerID.ValueInt64(),
		"is_active": plan.IsActive.ValueBool(),
		"is_debug":  plan.IsDebug.ValueBool(),
	}
	if !plan.TriggerParams.IsNull() && plan.TriggerParams.ValueString() != "" {
		values["trigger_params"] = plan.TriggerParams.ValueString()
	}
	if !plan.Description.IsNull() {
		values["description"] = plan.Description.ValueString()
	}
	if !plan.HelpText.IsNull() {
		values["help_text"] = plan.HelpText.ValueString()
	}

	result, err := r.client.Create("CiviRulesRule", values)
	if err != nil {
		resp.Diagnostics.AddError("Error creating CiviRulesRule", err.Error())
		return
	}

	r.mapResultToState(result, &plan)
	tflog.Debug(ctx, "Created CiviRulesRule", map[string]any{"id": plan.ID.ValueInt64()})
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *CiviRulesRuleResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state CiviRulesRuleResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	result, err := r.client.GetByID("CiviRulesRule", state.ID.ValueInt64(), nil)
	if err != nil {
		resp.Diagnostics.AddError("Error reading CiviRulesRule",
			"Could not read CiviRulesRule ID "+strconv.FormatInt(state.ID.ValueInt64(), 10)+": "+err.Error())
		return
	}

	r.mapResultToState(result, &state)
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (r *CiviRulesRuleResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan CiviRulesRuleResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	var state CiviRulesRuleResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	values := map[string]any{
		"name":       plan.Name.ValueString(),
		"label":      plan.Label.ValueString(),
		"trigger_id": plan.TriggerID.ValueInt64(),
		"is_active":  plan.IsActive.ValueBool(),
		"is_debug":   plan.IsDebug.ValueBool(),
	}
	if !plan.TriggerParams.IsNull() && plan.TriggerParams.ValueString() != "" {
		values["trigger_params"] = plan.TriggerParams.ValueString()
	} else {
		values["trigger_params"] = nil
	}
	if !plan.Description.IsNull() {
		values["description"] = plan.Description.ValueString()
	} else {
		values["description"] = nil
	}
	if !plan.HelpText.IsNull() {
		values["help_text"] = plan.HelpText.ValueString()
	} else {
		values["help_text"] = nil
	}

	result, err := r.client.Update("CiviRulesRule", state.ID.ValueInt64(), values)
	if err != nil {
		resp.Diagnostics.AddError("Error updating CiviRulesRule",
			"Could not update CiviRulesRule ID "+strconv.FormatInt(state.ID.ValueInt64(), 10)+": "+err.Error())
		return
	}

	plan.ID = state.ID
	r.mapResultToState(result, &plan)
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *CiviRulesRuleResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state CiviRulesRuleResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.client.Delete("CiviRulesRule", state.ID.ValueInt64()); err != nil {
		resp.Diagnostics.AddError("Error deleting CiviRulesRule",
			"Could not delete CiviRulesRule ID "+strconv.FormatInt(state.ID.ValueInt64(), 10)+": "+err.Error())
	}
}

func (r *CiviRulesRuleResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	id, err := strconv.ParseInt(req.ID, 10, 64)
	if err != nil {
		resp.Diagnostics.AddError("Invalid import ID", "Could not parse import ID as integer: "+err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), id)...)
}

func (r *CiviRulesRuleResource) mapResultToState(result map[string]any, model *CiviRulesRuleResourceModel) {
	if id, ok := GetInt64(result, "id"); ok {
		model.ID = types.Int64Value(id)
	}
	if v, ok := GetString(result, "name"); ok {
		model.Name = types.StringValue(v)
	}
	if v, ok := GetString(result, "label"); ok {
		model.Label = types.StringValue(v)
	}
	if v, ok := GetInt64(result, "trigger_id"); ok {
		model.TriggerID = types.Int64Value(v)
	}
	if v, ok := GetString(result, "trigger_params"); ok && v != "" {
		model.TriggerParams = types.StringValue(v)
	} else {
		model.TriggerParams = types.StringNull()
	}
	if v, ok := GetString(result, "description"); ok && v != "" {
		model.Description = types.StringValue(v)
	} else {
		model.Description = types.StringNull()
	}
	if v, ok := GetString(result, "help_text"); ok && v != "" {
		model.HelpText = types.StringValue(v)
	} else {
		model.HelpText = types.StringNull()
	}
	if v, ok := GetBool(result, "is_active"); ok {
		model.IsActive = types.BoolValue(v)
	}
	if v, ok := GetBool(result, "is_debug"); ok {
		model.IsDebug = types.BoolValue(v)
	}
}
