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
	_ resource.Resource                = &PriceSetResource{}
	_ resource.ResourceWithConfigure   = &PriceSetResource{}
	_ resource.ResourceWithImportState = &PriceSetResource{}
)

type PriceSetResource struct {
	client *Client
}

type PriceSetResourceModel struct {
	ID              types.Int64   `tfsdk:"id"`
	DomainID        types.Int64   `tfsdk:"domain_id"`
	Name            types.String  `tfsdk:"name"`
	Title           types.String  `tfsdk:"title"`
	IsActive        types.Bool    `tfsdk:"is_active"`
	HelpPre         types.String  `tfsdk:"help_pre"`
	HelpPost        types.String  `tfsdk:"help_post"`
	Javascript      types.String  `tfsdk:"javascript"`
	Extends         types.String  `tfsdk:"extends"`
	FinancialTypeID types.Int64   `tfsdk:"financial_type_id"`
	IsQuickConfig   types.Bool    `tfsdk:"is_quick_config"`
	IsReserved      types.Bool    `tfsdk:"is_reserved"`
	MinAmount       types.Float64 `tfsdk:"min_amount"`
}

func NewPriceSetResource() resource.Resource {
	return &PriceSetResource{}
}

func (r *PriceSetResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_price_set"
}

func (r *PriceSetResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a CiviCRM Price Set.",
		Attributes: map[string]schema.Attribute{
			"id": schema.Int64Attribute{
				Description: "The unique identifier of the price set.",
				Computed:    true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"domain_id": schema.Int64Attribute{
				Description: "Which Domain is this price set for.",
				Optional:    true,
				Computed:    true,
			},
			"name": schema.StringAttribute{
				Description: "Variable name/programmatic handle for this set of price fields.",
				Required:    true,
			},
			"title": schema.StringAttribute{
				Description: "Displayed title for the Price Set.",
				Required:    true,
			},
			"is_active": schema.BoolAttribute{
				Description: "Is this price set active. Default: true.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(true),
			},
			"help_pre": schema.StringAttribute{
				Description: "Description and/or help text to display before fields in form.",
				Optional:    true,
				Computed:    true,
			},
			"help_post": schema.StringAttribute{
				Description: "Description and/or help text to display after fields in form.",
				Optional:    true,
				Computed:    true,
			},
			"javascript": schema.StringAttribute{
				Description: "Optional Javascript script function(s) included on the form with this price set.",
				Optional:    true,
				Computed:    true,
			},
			"extends": schema.StringAttribute{
				Description: "What components are using this price set (e.g. CiviMember, CiviEvent).",
				Required:    true,
			},
			"financial_type_id": schema.Int64Attribute{
				Description: "FK to Financial Type (for membership price sets only).",
				Optional:    true,
				Computed:    true,
			},
			"is_quick_config": schema.BoolAttribute{
				Description: "Is set if edited on Contribution or Event Page rather than through Manage Price Sets. Default: false.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
			},
			"is_reserved": schema.BoolAttribute{
				Description: "Is this a predefined system price set (i.e. it cannot be deleted or edited). Default: false.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
			},
			"min_amount": schema.Float64Attribute{
				Description: "Minimum amount required for this set.",
				Optional:    true,
				Computed:    true,
			},
		},
	}
}

func (r *PriceSetResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *PriceSetResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan PriceSetResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Creating PriceSet", map[string]any{"name": plan.Name.ValueString()})

	values := map[string]any{
		"name":            plan.Name.ValueString(),
		"title":           plan.Title.ValueString(),
		"extends":         plan.Extends.ValueString(),
		"is_active":       plan.IsActive.ValueBool(),
		"is_quick_config": plan.IsQuickConfig.ValueBool(),
		"is_reserved":     plan.IsReserved.ValueBool(),
	}

	if !plan.DomainID.IsNull() {
		values["domain_id"] = plan.DomainID.ValueInt64()
	}
	if !plan.HelpPre.IsNull() {
		values["help_pre"] = plan.HelpPre.ValueString()
	}
	if !plan.HelpPost.IsNull() {
		values["help_post"] = plan.HelpPost.ValueString()
	}
	if !plan.Javascript.IsNull() {
		values["javascript"] = plan.Javascript.ValueString()
	}
	if !plan.FinancialTypeID.IsNull() {
		values["financial_type_id"] = plan.FinancialTypeID.ValueInt64()
	}
	if !plan.MinAmount.IsNull() {
		values["min_amount"] = plan.MinAmount.ValueFloat64()
	}

	result, err := r.client.Create("PriceSet", values)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating PriceSet",
			"Could not create PriceSet: "+err.Error(),
		)
		return
	}

	r.mapResultToState(result, &plan)

	tflog.Debug(ctx, "Created PriceSet", map[string]any{"id": plan.ID.ValueInt64()})

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

func (r *PriceSetResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state PriceSetResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Reading PriceSet", map[string]any{"id": state.ID.ValueInt64()})

	result, err := r.client.GetByID("PriceSet", state.ID.ValueInt64(), nil)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading PriceSet",
			"Could not read PriceSet ID "+strconv.FormatInt(state.ID.ValueInt64(), 10)+": "+err.Error(),
		)
		return
	}

	r.mapResultToState(result, &state)

	diags = resp.State.Set(ctx, state)
	resp.Diagnostics.Append(diags...)
}

func (r *PriceSetResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan PriceSetResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var state PriceSetResourceModel
	diags = req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Updating PriceSet", map[string]any{"id": state.ID.ValueInt64()})

	values := map[string]any{
		"name":            plan.Name.ValueString(),
		"title":           plan.Title.ValueString(),
		"extends":         plan.Extends.ValueString(),
		"is_active":       plan.IsActive.ValueBool(),
		"is_quick_config": plan.IsQuickConfig.ValueBool(),
		"is_reserved":     plan.IsReserved.ValueBool(),
	}

	if !plan.DomainID.IsNull() {
		values["domain_id"] = plan.DomainID.ValueInt64()
	} else {
		values["domain_id"] = nil
	}
	if !plan.HelpPre.IsNull() {
		values["help_pre"] = plan.HelpPre.ValueString()
	} else {
		values["help_pre"] = nil
	}
	if !plan.HelpPost.IsNull() {
		values["help_post"] = plan.HelpPost.ValueString()
	} else {
		values["help_post"] = nil
	}
	if !plan.Javascript.IsNull() {
		values["javascript"] = plan.Javascript.ValueString()
	} else {
		values["javascript"] = nil
	}
	if !plan.FinancialTypeID.IsNull() {
		values["financial_type_id"] = plan.FinancialTypeID.ValueInt64()
	} else {
		values["financial_type_id"] = nil
	}
	if !plan.MinAmount.IsNull() {
		values["min_amount"] = plan.MinAmount.ValueFloat64()
	} else {
		values["min_amount"] = nil
	}

	result, err := r.client.Update("PriceSet", state.ID.ValueInt64(), values)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error updating PriceSet",
			"Could not update PriceSet ID "+strconv.FormatInt(state.ID.ValueInt64(), 10)+": "+err.Error(),
		)
		return
	}

	plan.ID = state.ID
	r.mapResultToState(result, &plan)

	tflog.Debug(ctx, "Updated PriceSet", map[string]any{"id": plan.ID.ValueInt64()})

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

func (r *PriceSetResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state PriceSetResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Deleting PriceSet", map[string]any{"id": state.ID.ValueInt64()})

	err := r.client.Delete("PriceSet", state.ID.ValueInt64())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error deleting PriceSet",
			"Could not delete PriceSet ID "+strconv.FormatInt(state.ID.ValueInt64(), 10)+": "+err.Error(),
		)
		return
	}

	tflog.Debug(ctx, "Deleted PriceSet", map[string]any{"id": state.ID.ValueInt64()})
}

func (r *PriceSetResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
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

func (r *PriceSetResource) mapResultToState(result map[string]any, model *PriceSetResourceModel) {
	if id, ok := GetInt64(result, "id"); ok {
		model.ID = types.Int64Value(id)
	}

	if v, ok := GetInt64(result, "domain_id"); ok {
		model.DomainID = types.Int64Value(v)
	} else {
		model.DomainID = types.Int64Null()
	}

	if v, ok := GetString(result, "name"); ok {
		model.Name = types.StringValue(v)
	}

	if v, ok := GetString(result, "title"); ok {
		model.Title = types.StringValue(v)
	}

	if v, ok := GetBool(result, "is_active"); ok {
		model.IsActive = types.BoolValue(v)
	}

	if v, ok := GetString(result, "help_pre"); ok && v != "" {
		model.HelpPre = types.StringValue(v)
	} else {
		model.HelpPre = types.StringNull()
	}

	if v, ok := GetString(result, "help_post"); ok && v != "" {
		model.HelpPost = types.StringValue(v)
	} else {
		model.HelpPost = types.StringNull()
	}

	if v, ok := GetString(result, "javascript"); ok && v != "" {
		model.Javascript = types.StringValue(v)
	} else {
		model.Javascript = types.StringNull()
	}

	if v, ok := GetString(result, "extends"); ok {
		model.Extends = types.StringValue(v)
	}

	if v, ok := GetInt64(result, "financial_type_id"); ok {
		model.FinancialTypeID = types.Int64Value(v)
	} else {
		model.FinancialTypeID = types.Int64Null()
	}

	if v, ok := GetBool(result, "is_quick_config"); ok {
		model.IsQuickConfig = types.BoolValue(v)
	}

	if v, ok := GetBool(result, "is_reserved"); ok {
		model.IsReserved = types.BoolValue(v)
	}

	if v, ok := result["min_amount"]; ok && v != nil {
		switch val := v.(type) {
		case float64:
			model.MinAmount = types.Float64Value(val)
		case string:
			if f, err := strconv.ParseFloat(val, 64); err == nil {
				model.MinAmount = types.Float64Value(f)
			} else {
				model.MinAmount = types.Float64Null()
			}
		default:
			model.MinAmount = types.Float64Null()
		}
	} else {
		model.MinAmount = types.Float64Null()
	}
}
