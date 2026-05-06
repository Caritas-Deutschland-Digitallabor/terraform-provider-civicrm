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
	_ resource.Resource                = &PriceFieldValueResource{}
	_ resource.ResourceWithConfigure   = &PriceFieldValueResource{}
	_ resource.ResourceWithImportState = &PriceFieldValueResource{}
)

type PriceFieldValueResource struct {
	client *Client
}

type PriceFieldValueResourceModel struct {
	ID                  types.Int64   `tfsdk:"id"`
	PriceFieldID        types.Int64   `tfsdk:"price_field_id"`
	Name                types.String  `tfsdk:"name"`
	Label               types.String  `tfsdk:"label"`
	Description         types.String  `tfsdk:"description"`
	HelpPre             types.String  `tfsdk:"help_pre"`
	HelpPost            types.String  `tfsdk:"help_post"`
	Amount              types.Float64 `tfsdk:"amount"`
	Count               types.Int64   `tfsdk:"count"`
	MaxValue            types.Int64   `tfsdk:"max_value"`
	Weight              types.Int64   `tfsdk:"weight"`
	MembershipTypeID    types.Int64   `tfsdk:"membership_type_id"`
	MembershipNumTerms  types.Int64   `tfsdk:"membership_num_terms"`
	IsDefault           types.Bool    `tfsdk:"is_default"`
	IsActive            types.Bool    `tfsdk:"is_active"`
	FinancialTypeID     types.Int64   `tfsdk:"financial_type_id"`
	NonDeductibleAmount types.Float64 `tfsdk:"non_deductible_amount"`
	VisibilityID        types.Int64   `tfsdk:"visibility_id"`
}

func NewPriceFieldValueResource() resource.Resource {
	return &PriceFieldValueResource{}
}

func (r *PriceFieldValueResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_price_field_value"
}

func (r *PriceFieldValueResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a CiviCRM Price Field Value.",
		Attributes: map[string]schema.Attribute{
			"id": schema.Int64Attribute{
				Description: "The unique identifier of the price field value.",
				Computed:    true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"price_field_id": schema.Int64Attribute{
				Description: "The price field this value belongs to.",
				Required:    true,
			},
			"name": schema.StringAttribute{
				Description: "Variable name/programmatic handle for this price field value.",
				Required:    true,
			},
			"label": schema.StringAttribute{
				Description: "Displayed label for the price field value.",
				Required:    true,
			},
			"description": schema.StringAttribute{
				Description: "Description of the price field value.",
				Optional:    true,
				Computed:    true,
			},
			"help_pre": schema.StringAttribute{
				Description: "Help text displayed before the value.",
				Optional:    true,
				Computed:    true,
			},
			"help_post": schema.StringAttribute{
				Description: "Help text displayed after the value.",
				Optional:    true,
				Computed:    true,
			},
			"amount": schema.Float64Attribute{
				Description: "Price amount for this value.",
				Required:    true,
			},
			"count": schema.Int64Attribute{
				Description: "Number of participants this value counts for.",
				Optional:    true,
				Computed:    true,
			},
			"max_value": schema.Int64Attribute{
				Description: "Maximum number of this value that can be selected.",
				Optional:    true,
				Computed:    true,
			},
			"weight": schema.Int64Attribute{
				Description: "Order in which the values are displayed.",
				Optional:    true,
				Computed:    true,
			},
			"membership_type_id": schema.Int64Attribute{
				Description: "FK to Membership Type (for membership price fields only).",
				Optional:    true,
				Computed:    true,
			},
			"membership_num_terms": schema.Int64Attribute{
				Description: "Number of membership terms this value represents.",
				Optional:    true,
				Computed:    true,
			},
			"is_default": schema.BoolAttribute{
				Description: "Whether this is the default selected value. Default: false.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
			},
			"is_active": schema.BoolAttribute{
				Description: "Is this price field value active. Default: true.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(true),
			},
			"financial_type_id": schema.Int64Attribute{
				Description: "FK to Financial Type for this value.",
				Optional:    true,
				Computed:    true,
			},
			"non_deductible_amount": schema.Float64Attribute{
				Description: "Non-deductible portion of the amount.",
				Optional:    true,
				Computed:    true,
			},
			"visibility_id": schema.Int64Attribute{
				Description: "Visibility of this price field value (1=public, 2=admin).",
				Optional:    true,
				Computed:    true,
			},
		},
	}
}

func (r *PriceFieldValueResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *PriceFieldValueResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan PriceFieldValueResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Creating PriceFieldValue", map[string]any{"name": plan.Name.ValueString()})

	values := map[string]any{
		"price_field_id": plan.PriceFieldID.ValueInt64(),
		"name":           plan.Name.ValueString(),
		"label":          plan.Label.ValueString(),
		"amount":         plan.Amount.ValueFloat64(),
		"is_default":     plan.IsDefault.ValueBool(),
		"is_active":      plan.IsActive.ValueBool(),
	}

	if !plan.Description.IsNull() {
		values["description"] = plan.Description.ValueString()
	}
	if !plan.HelpPre.IsNull() {
		values["help_pre"] = plan.HelpPre.ValueString()
	}
	if !plan.HelpPost.IsNull() {
		values["help_post"] = plan.HelpPost.ValueString()
	}
	if !plan.Count.IsNull() {
		values["count"] = plan.Count.ValueInt64()
	}
	if !plan.MaxValue.IsNull() {
		values["max_value"] = plan.MaxValue.ValueInt64()
	}
	if !plan.Weight.IsNull() {
		values["weight"] = plan.Weight.ValueInt64()
	}
	if !plan.MembershipTypeID.IsNull() {
		values["membership_type_id"] = plan.MembershipTypeID.ValueInt64()
	}
	if !plan.MembershipNumTerms.IsNull() {
		values["membership_num_terms"] = plan.MembershipNumTerms.ValueInt64()
	}
	if !plan.FinancialTypeID.IsNull() {
		values["financial_type_id"] = plan.FinancialTypeID.ValueInt64()
	}
	if !plan.NonDeductibleAmount.IsNull() {
		values["non_deductible_amount"] = plan.NonDeductibleAmount.ValueFloat64()
	}
	if !plan.VisibilityID.IsNull() {
		values["visibility_id"] = plan.VisibilityID.ValueInt64()
	}

	result, err := r.client.Create("PriceFieldValue", values)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating PriceFieldValue",
			"Could not create PriceFieldValue: "+err.Error(),
		)
		return
	}

	r.mapResultToState(result, &plan)

	tflog.Debug(ctx, "Created PriceFieldValue", map[string]any{"id": plan.ID.ValueInt64()})

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

func (r *PriceFieldValueResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state PriceFieldValueResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Reading PriceFieldValue", map[string]any{"id": state.ID.ValueInt64()})

	result, err := r.client.GetByID("PriceFieldValue", state.ID.ValueInt64(), nil)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading PriceFieldValue",
			"Could not read PriceFieldValue ID "+strconv.FormatInt(state.ID.ValueInt64(), 10)+": "+err.Error(),
		)
		return
	}

	r.mapResultToState(result, &state)

	diags = resp.State.Set(ctx, state)
	resp.Diagnostics.Append(diags...)
}

func (r *PriceFieldValueResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan PriceFieldValueResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var state PriceFieldValueResourceModel
	diags = req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Updating PriceFieldValue", map[string]any{"id": state.ID.ValueInt64()})

	values := map[string]any{
		"price_field_id": plan.PriceFieldID.ValueInt64(),
		"name":           plan.Name.ValueString(),
		"label":          plan.Label.ValueString(),
		"amount":         plan.Amount.ValueFloat64(),
		"is_default":     plan.IsDefault.ValueBool(),
		"is_active":      plan.IsActive.ValueBool(),
	}

	if !plan.Description.IsNull() {
		values["description"] = plan.Description.ValueString()
	} else {
		values["description"] = nil
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
	if !plan.Count.IsNull() {
		values["count"] = plan.Count.ValueInt64()
	} else {
		values["count"] = nil
	}
	if !plan.MaxValue.IsNull() {
		values["max_value"] = plan.MaxValue.ValueInt64()
	} else {
		values["max_value"] = nil
	}
	if !plan.Weight.IsNull() {
		values["weight"] = plan.Weight.ValueInt64()
	} else {
		values["weight"] = nil
	}
	if !plan.MembershipTypeID.IsNull() {
		values["membership_type_id"] = plan.MembershipTypeID.ValueInt64()
	} else {
		values["membership_type_id"] = nil
	}
	if !plan.MembershipNumTerms.IsNull() {
		values["membership_num_terms"] = plan.MembershipNumTerms.ValueInt64()
	} else {
		values["membership_num_terms"] = nil
	}
	if !plan.FinancialTypeID.IsNull() {
		values["financial_type_id"] = plan.FinancialTypeID.ValueInt64()
	} else {
		values["financial_type_id"] = nil
	}
	if !plan.NonDeductibleAmount.IsNull() {
		values["non_deductible_amount"] = plan.NonDeductibleAmount.ValueFloat64()
	} else {
		values["non_deductible_amount"] = nil
	}
	if !plan.VisibilityID.IsNull() {
		values["visibility_id"] = plan.VisibilityID.ValueInt64()
	} else {
		values["visibility_id"] = nil
	}

	result, err := r.client.Update("PriceFieldValue", state.ID.ValueInt64(), values)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error updating PriceFieldValue",
			"Could not update PriceFieldValue ID "+strconv.FormatInt(state.ID.ValueInt64(), 10)+": "+err.Error(),
		)
		return
	}

	plan.ID = state.ID
	r.mapResultToState(result, &plan)

	tflog.Debug(ctx, "Updated PriceFieldValue", map[string]any{"id": plan.ID.ValueInt64()})

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

func (r *PriceFieldValueResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state PriceFieldValueResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Deleting PriceFieldValue", map[string]any{"id": state.ID.ValueInt64()})

	err := r.client.Delete("PriceFieldValue", state.ID.ValueInt64())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error deleting PriceFieldValue",
			"Could not delete PriceFieldValue ID "+strconv.FormatInt(state.ID.ValueInt64(), 10)+": "+err.Error(),
		)
		return
	}

	tflog.Debug(ctx, "Deleted PriceFieldValue", map[string]any{"id": state.ID.ValueInt64()})
}

func (r *PriceFieldValueResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
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

func (r *PriceFieldValueResource) mapResultToState(result map[string]any, model *PriceFieldValueResourceModel) {
	if id, ok := GetInt64(result, "id"); ok {
		model.ID = types.Int64Value(id)
	}

	if v, ok := GetInt64(result, "price_field_id"); ok {
		model.PriceFieldID = types.Int64Value(v)
	}

	if v, ok := GetString(result, "name"); ok {
		model.Name = types.StringValue(v)
	}

	if v, ok := GetString(result, "label"); ok {
		model.Label = types.StringValue(v)
	}

	if v, ok := GetString(result, "description"); ok && v != "" {
		model.Description = types.StringValue(v)
	} else {
		model.Description = types.StringNull()
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

	if v, ok := result["amount"]; ok && v != nil {
		switch val := v.(type) {
		case float64:
			model.Amount = types.Float64Value(val)
		case string:
			if f, err := strconv.ParseFloat(val, 64); err == nil {
				model.Amount = types.Float64Value(f)
			}
		}
	}

	if v, ok := GetInt64(result, "count"); ok {
		model.Count = types.Int64Value(v)
	} else {
		model.Count = types.Int64Null()
	}

	if v, ok := GetInt64(result, "max_value"); ok {
		model.MaxValue = types.Int64Value(v)
	} else {
		model.MaxValue = types.Int64Null()
	}

	if v, ok := GetInt64(result, "weight"); ok {
		model.Weight = types.Int64Value(v)
	} else {
		model.Weight = types.Int64Null()
	}

	if v, ok := GetInt64(result, "membership_type_id"); ok {
		model.MembershipTypeID = types.Int64Value(v)
	} else {
		model.MembershipTypeID = types.Int64Null()
	}

	if v, ok := GetInt64(result, "membership_num_terms"); ok {
		model.MembershipNumTerms = types.Int64Value(v)
	} else {
		model.MembershipNumTerms = types.Int64Null()
	}

	if v, ok := GetBool(result, "is_default"); ok {
		model.IsDefault = types.BoolValue(v)
	}

	if v, ok := GetBool(result, "is_active"); ok {
		model.IsActive = types.BoolValue(v)
	}

	if v, ok := GetInt64(result, "financial_type_id"); ok {
		model.FinancialTypeID = types.Int64Value(v)
	} else {
		model.FinancialTypeID = types.Int64Null()
	}

	if v, ok := result["non_deductible_amount"]; ok && v != nil {
		switch val := v.(type) {
		case float64:
			model.NonDeductibleAmount = types.Float64Value(val)
		case string:
			if f, err := strconv.ParseFloat(val, 64); err == nil {
				model.NonDeductibleAmount = types.Float64Value(f)
			} else {
				model.NonDeductibleAmount = types.Float64Null()
			}
		default:
			model.NonDeductibleAmount = types.Float64Null()
		}
	} else {
		model.NonDeductibleAmount = types.Float64Null()
	}

	if v, ok := GetInt64(result, "visibility_id"); ok {
		model.VisibilityID = types.Int64Value(v)
	} else {
		model.VisibilityID = types.Int64Null()
	}
}
