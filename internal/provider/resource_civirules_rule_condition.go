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
	_ resource.Resource                = &CiviRulesRuleConditionResource{}
	_ resource.ResourceWithConfigure   = &CiviRulesRuleConditionResource{}
	_ resource.ResourceWithImportState = &CiviRulesRuleConditionResource{}
)

// CiviRulesRuleConditionResource links a condition type to a rule (entity: CiviRulesRuleCondition).
// One rule can have multiple conditions. All conditions must pass (AND logic) unless
// the CiviRules extension is configured differently.
type CiviRulesRuleConditionResource struct {
	client *Client
}

type CiviRulesRuleConditionResourceModel struct {
	ID              types.Int64  `tfsdk:"id"`
	RuleID          types.Int64  `tfsdk:"rule_id"`
	ConditionID     types.Int64  `tfsdk:"condition_id"`
	ConditionParams types.String `tfsdk:"condition_params"`
	IsActive        types.Bool   `tfsdk:"is_active"`
	Negate          types.Bool   `tfsdk:"negate"`
}

func NewCiviRulesRuleConditionResource() resource.Resource {
	return &CiviRulesRuleConditionResource{}
}

func (r *CiviRulesRuleConditionResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_civirules_rule_condition"
}

func (r *CiviRulesRuleConditionResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Attaches a condition to a CiviRules rule (entity: CiviRulesRuleCondition). The condition_id references a CiviRulesCondition type. Parameters are passed as a JSON string whose structure depends on the condition class.",
		Attributes: map[string]schema.Attribute{
			"id": schema.Int64Attribute{
				Description: "Unique ID of the rule-condition link.",
				Computed:    true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"rule_id": schema.Int64Attribute{
				Description: "ID of the civicrm_civirules_rule this condition belongs to.",
				Required:    true,
			},
			"condition_id": schema.Int64Attribute{
				Description: "ID of the CiviRulesCondition type to apply. Look up available condition IDs via the CiviRules UI or CiviRulesCondition.get API.",
				Required:    true,
			},
			"condition_params": schema.StringAttribute{
				Description: "JSON-encoded parameters passed to the condition class. Structure depends on the condition type (e.g. '{\"case_type_id\": \"3\"}' for a case-type condition).",
				Optional:    true,
			},
			"is_active": schema.BoolAttribute{
				Description: "Whether this condition is active. Default: true.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(true),
			},
			"negate": schema.BoolAttribute{
				Description: "When true, the condition logic is inverted (NOT). Default: false.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
			},
		},
	}
}

func (r *CiviRulesRuleConditionResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *CiviRulesRuleConditionResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan CiviRulesRuleConditionResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Creating CiviRulesRuleCondition", map[string]any{
		"rule_id": plan.RuleID.ValueInt64(), "condition_id": plan.ConditionID.ValueInt64(),
	})

	values := map[string]any{
		"rule_id":      plan.RuleID.ValueInt64(),
		"condition_id": plan.ConditionID.ValueInt64(),
		"is_active":    plan.IsActive.ValueBool(),
		"negate":       plan.Negate.ValueBool(),
	}
	if !plan.ConditionParams.IsNull() && plan.ConditionParams.ValueString() != "" {
		values["condition_params"] = plan.ConditionParams.ValueString()
	}

	result, err := r.client.Create("CiviRulesRuleCondition", values)
	if err != nil {
		resp.Diagnostics.AddError("Error creating CiviRulesRuleCondition", err.Error())
		return
	}

	r.mapResultToState(result, &plan)
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *CiviRulesRuleConditionResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state CiviRulesRuleConditionResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	result, err := r.client.GetByID("CiviRulesRuleCondition", state.ID.ValueInt64(), nil)
	if err != nil {
		resp.Diagnostics.AddError("Error reading CiviRulesRuleCondition",
			"Could not read CiviRulesRuleCondition ID "+strconv.FormatInt(state.ID.ValueInt64(), 10)+": "+err.Error())
		return
	}

	r.mapResultToState(result, &state)
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (r *CiviRulesRuleConditionResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan CiviRulesRuleConditionResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	var state CiviRulesRuleConditionResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	values := map[string]any{
		"rule_id":      plan.RuleID.ValueInt64(),
		"condition_id": plan.ConditionID.ValueInt64(),
		"is_active":    plan.IsActive.ValueBool(),
		"negate":       plan.Negate.ValueBool(),
	}
	if !plan.ConditionParams.IsNull() && plan.ConditionParams.ValueString() != "" {
		values["condition_params"] = plan.ConditionParams.ValueString()
	} else {
		values["condition_params"] = nil
	}

	result, err := r.client.Update("CiviRulesRuleCondition", state.ID.ValueInt64(), values)
	if err != nil {
		resp.Diagnostics.AddError("Error updating CiviRulesRuleCondition",
			"Could not update CiviRulesRuleCondition ID "+strconv.FormatInt(state.ID.ValueInt64(), 10)+": "+err.Error())
		return
	}

	plan.ID = state.ID
	r.mapResultToState(result, &plan)
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *CiviRulesRuleConditionResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state CiviRulesRuleConditionResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.client.Delete("CiviRulesRuleCondition", state.ID.ValueInt64()); err != nil {
		resp.Diagnostics.AddError("Error deleting CiviRulesRuleCondition",
			"Could not delete CiviRulesRuleCondition ID "+strconv.FormatInt(state.ID.ValueInt64(), 10)+": "+err.Error())
	}
}

func (r *CiviRulesRuleConditionResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	id, err := strconv.ParseInt(req.ID, 10, 64)
	if err != nil {
		resp.Diagnostics.AddError("Invalid import ID", "Could not parse import ID as integer: "+err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), id)...)
}

func (r *CiviRulesRuleConditionResource) mapResultToState(result map[string]any, model *CiviRulesRuleConditionResourceModel) {
	if id, ok := GetInt64(result, "id"); ok {
		model.ID = types.Int64Value(id)
	}
	if v, ok := GetInt64(result, "rule_id"); ok {
		model.RuleID = types.Int64Value(v)
	}
	if v, ok := GetInt64(result, "condition_id"); ok {
		model.ConditionID = types.Int64Value(v)
	}
	if v, ok := GetString(result, "condition_params"); ok && v != "" {
		model.ConditionParams = types.StringValue(v)
	} else {
		model.ConditionParams = types.StringNull()
	}
	if v, ok := GetBool(result, "is_active"); ok {
		model.IsActive = types.BoolValue(v)
	}
	if v, ok := GetBool(result, "negate"); ok {
		model.Negate = types.BoolValue(v)
	}
}
