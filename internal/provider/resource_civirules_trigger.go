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
	_ resource.Resource                = &CiviRulesTriggerResource{}
	_ resource.ResourceWithConfigure   = &CiviRulesTriggerResource{}
	_ resource.ResourceWithImportState = &CiviRulesTriggerResource{}
)

// CiviRulesTriggerResource manages a CiviRules trigger definition (entity: CiviRulesTrigger).
// Triggers are the "when" of a CiviRules rule. Most triggers are provided by the
// CiviRules extension itself (e.g. "Case is changed") and do not need to be created
// via Terraform. Use this resource only for custom/cron triggers.
// Reference an existing trigger by its ID in civicrm_civirules_rule.trigger_id.
type CiviRulesTriggerResource struct {
	client *Client
}

type CiviRulesTriggerResourceModel struct {
	ID         types.Int64  `tfsdk:"id"`
	Name       types.String `tfsdk:"name"`
	Label      types.String `tfsdk:"label"`
	ObjectName types.String `tfsdk:"object_name"`
	Op         types.String `tfsdk:"op"`
	Cron       types.Bool   `tfsdk:"cron"`
	ClassName  types.String `tfsdk:"class_name"`
	IsActive   types.Bool   `tfsdk:"is_active"`
}

func NewCiviRulesTriggerResource() resource.Resource {
	return &CiviRulesTriggerResource{}
}

func (r *CiviRulesTriggerResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_civirules_trigger"
}

func (r *CiviRulesTriggerResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a CiviRules Trigger definition. Most standard triggers (case changed, activity added, etc.) are shipped by the CiviRules extension and should be referenced by ID in civicrm_civirules_rule rather than created here. Use this resource only for custom cron triggers or triggers provided by a custom PHP class.",
		Attributes: map[string]schema.Attribute{
			"id": schema.Int64Attribute{
				Description: "Unique ID of the CiviRules trigger.",
				Computed:    true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Description: "Machine name of the trigger. Must be unique.",
				Required:    true,
			},
			"label": schema.StringAttribute{
				Description: "Human-readable label shown in the CiviRules rule form.",
				Required:    true,
			},
			"object_name": schema.StringAttribute{
				Description: "CiviCRM entity this trigger fires on (e.g. 'Case', 'Activity', 'Contact'). Leave empty for cron triggers.",
				Optional:    true,
			},
			"op": schema.StringAttribute{
				Description: "Operation that fires the trigger: 'create', 'edit', 'delete', or empty for cron.",
				Optional:    true,
			},
			"cron": schema.BoolAttribute{
				Description: "Whether this is a cron-based (scheduled) trigger. Default: false.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
			},
			"class_name": schema.StringAttribute{
				Description: "Fully-qualified PHP class name that implements the trigger logic (e.g. 'CRM_CivirulesMyExt_Trigger_MyTrigger').",
				Optional:    true,
			},
			"is_active": schema.BoolAttribute{
				Description: "Whether this trigger is active. Default: true.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(true),
			},
		},
	}
}

func (r *CiviRulesTriggerResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *CiviRulesTriggerResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan CiviRulesTriggerResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Creating CiviRulesTrigger", map[string]any{"name": plan.Name.ValueString()})

	values := map[string]any{
		"name":      plan.Name.ValueString(),
		"label":     plan.Label.ValueString(),
		"cron":      plan.Cron.ValueBool(),
		"is_active": plan.IsActive.ValueBool(),
	}
	if !plan.ObjectName.IsNull() && plan.ObjectName.ValueString() != "" {
		values["object_name"] = plan.ObjectName.ValueString()
	}
	if !plan.Op.IsNull() && plan.Op.ValueString() != "" {
		values["op"] = plan.Op.ValueString()
	}
	if !plan.ClassName.IsNull() && plan.ClassName.ValueString() != "" {
		values["class_name"] = plan.ClassName.ValueString()
	}

	result, err := r.client.Create("CiviRulesTrigger", values)
	if err != nil {
		resp.Diagnostics.AddError("Error creating CiviRulesTrigger", err.Error())
		return
	}

	r.mapResultToState(result, &plan)
	tflog.Debug(ctx, "Created CiviRulesTrigger", map[string]any{"id": plan.ID.ValueInt64()})
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *CiviRulesTriggerResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state CiviRulesTriggerResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	result, err := r.client.GetByID("CiviRulesTrigger", state.ID.ValueInt64(), nil)
	if err != nil {
		resp.Diagnostics.AddError("Error reading CiviRulesTrigger",
			"Could not read CiviRulesTrigger ID "+strconv.FormatInt(state.ID.ValueInt64(), 10)+": "+err.Error())
		return
	}

	r.mapResultToState(result, &state)
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (r *CiviRulesTriggerResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan CiviRulesTriggerResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	var state CiviRulesTriggerResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	values := map[string]any{
		"name":      plan.Name.ValueString(),
		"label":     plan.Label.ValueString(),
		"cron":      plan.Cron.ValueBool(),
		"is_active": plan.IsActive.ValueBool(),
	}
	if !plan.ObjectName.IsNull() && plan.ObjectName.ValueString() != "" {
		values["object_name"] = plan.ObjectName.ValueString()
	} else {
		values["object_name"] = nil
	}
	if !plan.Op.IsNull() && plan.Op.ValueString() != "" {
		values["op"] = plan.Op.ValueString()
	} else {
		values["op"] = nil
	}
	if !plan.ClassName.IsNull() && plan.ClassName.ValueString() != "" {
		values["class_name"] = plan.ClassName.ValueString()
	} else {
		values["class_name"] = nil
	}

	result, err := r.client.Update("CiviRulesTrigger", state.ID.ValueInt64(), values)
	if err != nil {
		resp.Diagnostics.AddError("Error updating CiviRulesTrigger",
			"Could not update CiviRulesTrigger ID "+strconv.FormatInt(state.ID.ValueInt64(), 10)+": "+err.Error())
		return
	}

	plan.ID = state.ID
	r.mapResultToState(result, &plan)
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *CiviRulesTriggerResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state CiviRulesTriggerResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.client.Delete("CiviRulesTrigger", state.ID.ValueInt64()); err != nil {
		resp.Diagnostics.AddError("Error deleting CiviRulesTrigger",
			"Could not delete CiviRulesTrigger ID "+strconv.FormatInt(state.ID.ValueInt64(), 10)+": "+err.Error())
	}
}

func (r *CiviRulesTriggerResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	id, err := strconv.ParseInt(req.ID, 10, 64)
	if err != nil {
		resp.Diagnostics.AddError("Invalid import ID", "Could not parse import ID as integer: "+err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), id)...)
}

func (r *CiviRulesTriggerResource) mapResultToState(result map[string]any, model *CiviRulesTriggerResourceModel) {
	if id, ok := GetInt64(result, "id"); ok {
		model.ID = types.Int64Value(id)
	}
	if v, ok := GetString(result, "name"); ok {
		model.Name = types.StringValue(v)
	}
	if v, ok := GetString(result, "label"); ok {
		model.Label = types.StringValue(v)
	}
	if v, ok := GetString(result, "object_name"); ok && v != "" {
		model.ObjectName = types.StringValue(v)
	} else {
		model.ObjectName = types.StringNull()
	}
	if v, ok := GetString(result, "op"); ok && v != "" {
		model.Op = types.StringValue(v)
	} else {
		model.Op = types.StringNull()
	}
	if v, ok := GetBool(result, "cron"); ok {
		model.Cron = types.BoolValue(v)
	}
	if v, ok := GetString(result, "class_name"); ok && v != "" {
		model.ClassName = types.StringValue(v)
	} else {
		model.ClassName = types.StringNull()
	}
	if v, ok := GetBool(result, "is_active"); ok {
		model.IsActive = types.BoolValue(v)
	}
}
