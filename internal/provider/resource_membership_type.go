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
	_ resource.Resource                = &MembershipTypeResource{}
	_ resource.ResourceWithConfigure   = &MembershipTypeResource{}
	_ resource.ResourceWithImportState = &MembershipTypeResource{}
)

type MembershipTypeResource struct {
	client *Client
}

type MembershipTypeResourceModel struct {
	ID                    types.Int64   `tfsdk:"id"`
	Name                  types.String  `tfsdk:"name"`
	Description           types.String  `tfsdk:"description"`
	MemberOfContactID     types.Int64   `tfsdk:"member_of_contact_id"`
	FinancialTypeID       types.Int64   `tfsdk:"financial_type_id"`
	MinimumFee            types.Float64 `tfsdk:"minimum_fee"`
	DurationUnit          types.String  `tfsdk:"duration_unit"`
	DurationInterval      types.Int64   `tfsdk:"duration_interval"`
	PeriodType            types.String  `tfsdk:"period_type"`
	FixedPeriodStartDay   types.Int64   `tfsdk:"fixed_period_start_day"`
	FixedPeriodRolloverDay types.Int64  `tfsdk:"fixed_period_rollover_day"`
	RelationshipTypeID    types.Int64   `tfsdk:"relationship_type_id"`
	RelationshipDirection types.String  `tfsdk:"relationship_direction"`
	MaxRelated            types.Int64   `tfsdk:"max_related"`
	Visibility            types.String  `tfsdk:"visibility"`
	Weight                types.Int64   `tfsdk:"weight"`
	ReceiptTextSignup     types.String  `tfsdk:"receipt_text_signup"`
	ReceiptTextRenewal    types.String  `tfsdk:"receipt_text_renewal"`
	AutoRenew             types.Bool    `tfsdk:"auto_renew"`
	IsActive              types.Bool    `tfsdk:"is_active"`
}

func NewMembershipTypeResource() resource.Resource {
	return &MembershipTypeResource{}
}

func (r *MembershipTypeResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_membership_type"
}

func (r *MembershipTypeResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a CiviCRM Membership Type.",
		Attributes: map[string]schema.Attribute{
			"id": schema.Int64Attribute{
				Description: "The unique identifier of the membership type.",
				Computed:    true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Description: "The name of the membership type.",
				Required:    true,
			},
			"description": schema.StringAttribute{
				Description: "A description of the membership type.",
				Optional:    true,
			},
			"member_of_contact_id": schema.Int64Attribute{
				Description: "The contact ID of the organization this membership type belongs to.",
				Required:    true,
			},
			"financial_type_id": schema.Int64Attribute{
				Description: "The financial type ID used for membership fees.",
				Required:    true,
			},
			"minimum_fee": schema.Float64Attribute{
				Description: "The minimum fee for this membership type.",
				Optional:    true,
				Computed:    true,
			},
			"duration_unit": schema.StringAttribute{
				Description: "The unit of duration (day, month, year, lifetime).",
				Required:    true,
			},
			"duration_interval": schema.Int64Attribute{
				Description: "The number of duration units.",
				Optional:    true,
				Computed:    true,
			},
			"period_type": schema.StringAttribute{
				Description: "The period type: rolling or fixed.",
				Required:    true,
			},
			"fixed_period_start_day": schema.Int64Attribute{
				Description: "For fixed period memberships, the start day (MMDD format, e.g. 101 for January 1).",
				Optional:    true,
				Computed:    true,
			},
			"fixed_period_rollover_day": schema.Int64Attribute{
				Description: "For fixed period memberships, the rollover day (MMDD format).",
				Optional:    true,
				Computed:    true,
			},
			"relationship_type_id": schema.Int64Attribute{
				Description: "The relationship type ID for inherited memberships.",
				Optional:    true,
				Computed:    true,
			},
			"relationship_direction": schema.StringAttribute{
				Description: "The relationship direction for inherited memberships.",
				Optional:    true,
				Computed:    true,
			},
			"max_related": schema.Int64Attribute{
				Description: "The maximum number of related memberships.",
				Optional:    true,
				Computed:    true,
			},
			"visibility": schema.StringAttribute{
				Description: "Visibility of the membership type (Public or Admin).",
				Optional:    true,
				Computed:    true,
			},
			"weight": schema.Int64Attribute{
				Description: "The sort weight of the membership type.",
				Optional:    true,
				Computed:    true,
			},
			"receipt_text_signup": schema.StringAttribute{
				Description: "Text to include in signup receipts.",
				Optional:    true,
				Computed:    true,
			},
			"receipt_text_renewal": schema.StringAttribute{
				Description: "Text to include in renewal receipts.",
				Optional:    true,
				Computed:    true,
			},
			"auto_renew": schema.BoolAttribute{
				Description: "Whether this membership type supports auto-renewal. Default: false.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
			},
			"is_active": schema.BoolAttribute{
				Description: "Whether the membership type is active. Default: true.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(true),
			},
		},
	}
}

func (r *MembershipTypeResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *MembershipTypeResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan MembershipTypeResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Creating MembershipType", map[string]any{"name": plan.Name.ValueString()})

	values := map[string]any{
		"name":                plan.Name.ValueString(),
		"member_of_contact_id": plan.MemberOfContactID.ValueInt64(),
		"financial_type_id":   plan.FinancialTypeID.ValueInt64(),
		"duration_unit":       plan.DurationUnit.ValueString(),
		"period_type":         plan.PeriodType.ValueString(),
		"auto_renew":          plan.AutoRenew.ValueBool(),
		"is_active":           plan.IsActive.ValueBool(),
	}

	if !plan.Description.IsNull() {
		values["description"] = plan.Description.ValueString()
	}
	if !plan.MinimumFee.IsNull() {
		values["minimum_fee"] = plan.MinimumFee.ValueFloat64()
	}
	if !plan.DurationInterval.IsNull() {
		values["duration_interval"] = plan.DurationInterval.ValueInt64()
	}
	if !plan.FixedPeriodStartDay.IsNull() {
		values["fixed_period_start_day"] = plan.FixedPeriodStartDay.ValueInt64()
	}
	if !plan.FixedPeriodRolloverDay.IsNull() {
		values["fixed_period_rollover_day"] = plan.FixedPeriodRolloverDay.ValueInt64()
	}
	if !plan.RelationshipTypeID.IsNull() {
		values["relationship_type_id"] = plan.RelationshipTypeID.ValueInt64()
	}
	if !plan.RelationshipDirection.IsNull() {
		values["relationship_direction"] = plan.RelationshipDirection.ValueString()
	}
	if !plan.MaxRelated.IsNull() {
		values["max_related"] = plan.MaxRelated.ValueInt64()
	}
	if !plan.Visibility.IsNull() {
		values["visibility"] = plan.Visibility.ValueString()
	}
	if !plan.Weight.IsNull() {
		values["weight"] = plan.Weight.ValueInt64()
	}
	if !plan.ReceiptTextSignup.IsNull() {
		values["receipt_text_signup"] = plan.ReceiptTextSignup.ValueString()
	}
	if !plan.ReceiptTextRenewal.IsNull() {
		values["receipt_text_renewal"] = plan.ReceiptTextRenewal.ValueString()
	}

	result, err := r.client.Create("MembershipType", values)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating MembershipType",
			"Could not create MembershipType: "+err.Error(),
		)
		return
	}

	r.mapResultToState(result, &plan)

	tflog.Debug(ctx, "Created MembershipType", map[string]any{"id": plan.ID.ValueInt64()})

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

func (r *MembershipTypeResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state MembershipTypeResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Reading MembershipType", map[string]any{"id": state.ID.ValueInt64()})

	result, err := r.client.GetByID("MembershipType", state.ID.ValueInt64(), nil)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading MembershipType",
			"Could not read MembershipType ID "+strconv.FormatInt(state.ID.ValueInt64(), 10)+": "+err.Error(),
		)
		return
	}

	r.mapResultToState(result, &state)

	diags = resp.State.Set(ctx, state)
	resp.Diagnostics.Append(diags...)
}

func (r *MembershipTypeResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan MembershipTypeResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var state MembershipTypeResourceModel
	diags = req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Updating MembershipType", map[string]any{"id": state.ID.ValueInt64()})

	values := map[string]any{
		"name":                plan.Name.ValueString(),
		"member_of_contact_id": plan.MemberOfContactID.ValueInt64(),
		"financial_type_id":   plan.FinancialTypeID.ValueInt64(),
		"duration_unit":       plan.DurationUnit.ValueString(),
		"period_type":         plan.PeriodType.ValueString(),
		"auto_renew":          plan.AutoRenew.ValueBool(),
		"is_active":           plan.IsActive.ValueBool(),
	}

	if !plan.Description.IsNull() {
		values["description"] = plan.Description.ValueString()
	} else {
		values["description"] = nil
	}
	if !plan.MinimumFee.IsNull() {
		values["minimum_fee"] = plan.MinimumFee.ValueFloat64()
	} else {
		values["minimum_fee"] = nil
	}
	if !plan.DurationInterval.IsNull() {
		values["duration_interval"] = plan.DurationInterval.ValueInt64()
	} else {
		values["duration_interval"] = nil
	}
	if !plan.FixedPeriodStartDay.IsNull() {
		values["fixed_period_start_day"] = plan.FixedPeriodStartDay.ValueInt64()
	} else {
		values["fixed_period_start_day"] = nil
	}
	if !plan.FixedPeriodRolloverDay.IsNull() {
		values["fixed_period_rollover_day"] = plan.FixedPeriodRolloverDay.ValueInt64()
	} else {
		values["fixed_period_rollover_day"] = nil
	}
	if !plan.RelationshipTypeID.IsNull() {
		values["relationship_type_id"] = plan.RelationshipTypeID.ValueInt64()
	} else {
		values["relationship_type_id"] = nil
	}
	if !plan.RelationshipDirection.IsNull() {
		values["relationship_direction"] = plan.RelationshipDirection.ValueString()
	} else {
		values["relationship_direction"] = nil
	}
	if !plan.MaxRelated.IsNull() {
		values["max_related"] = plan.MaxRelated.ValueInt64()
	} else {
		values["max_related"] = nil
	}
	if !plan.Visibility.IsNull() {
		values["visibility"] = plan.Visibility.ValueString()
	} else {
		values["visibility"] = nil
	}
	if !plan.Weight.IsNull() {
		values["weight"] = plan.Weight.ValueInt64()
	} else {
		values["weight"] = nil
	}
	if !plan.ReceiptTextSignup.IsNull() {
		values["receipt_text_signup"] = plan.ReceiptTextSignup.ValueString()
	} else {
		values["receipt_text_signup"] = nil
	}
	if !plan.ReceiptTextRenewal.IsNull() {
		values["receipt_text_renewal"] = plan.ReceiptTextRenewal.ValueString()
	} else {
		values["receipt_text_renewal"] = nil
	}

	result, err := r.client.Update("MembershipType", state.ID.ValueInt64(), values)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error updating MembershipType",
			"Could not update MembershipType ID "+strconv.FormatInt(state.ID.ValueInt64(), 10)+": "+err.Error(),
		)
		return
	}

	plan.ID = state.ID
	r.mapResultToState(result, &plan)

	tflog.Debug(ctx, "Updated MembershipType", map[string]any{"id": plan.ID.ValueInt64()})

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

func (r *MembershipTypeResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state MembershipTypeResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Deleting MembershipType", map[string]any{"id": state.ID.ValueInt64()})

	err := r.client.Delete("MembershipType", state.ID.ValueInt64())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error deleting MembershipType",
			"Could not delete MembershipType ID "+strconv.FormatInt(state.ID.ValueInt64(), 10)+": "+err.Error(),
		)
		return
	}

	tflog.Debug(ctx, "Deleted MembershipType", map[string]any{"id": state.ID.ValueInt64()})
}

func (r *MembershipTypeResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
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

func (r *MembershipTypeResource) mapResultToState(result map[string]any, model *MembershipTypeResourceModel) {
	if id, ok := GetInt64(result, "id"); ok {
		model.ID = types.Int64Value(id)
	}

	if v, ok := GetString(result, "name"); ok {
		model.Name = types.StringValue(v)
	}

	if v, ok := GetString(result, "description"); ok && v != "" {
		model.Description = types.StringValue(v)
	} else {
		model.Description = types.StringNull()
	}

	if v, ok := GetInt64(result, "member_of_contact_id"); ok {
		model.MemberOfContactID = types.Int64Value(v)
	}

	if v, ok := GetInt64(result, "financial_type_id"); ok {
		model.FinancialTypeID = types.Int64Value(v)
	}

	if v, ok := result["minimum_fee"]; ok && v != nil {
		switch val := v.(type) {
		case float64:
			model.MinimumFee = types.Float64Value(val)
		case string:
			if f, err := strconv.ParseFloat(val, 64); err == nil {
				model.MinimumFee = types.Float64Value(f)
			} else {
				model.MinimumFee = types.Float64Null()
			}
		default:
			model.MinimumFee = types.Float64Null()
		}
	} else {
		model.MinimumFee = types.Float64Null()
	}

	if v, ok := GetString(result, "duration_unit"); ok {
		model.DurationUnit = types.StringValue(v)
	}

	if v, ok := GetInt64(result, "duration_interval"); ok {
		model.DurationInterval = types.Int64Value(v)
	} else {
		model.DurationInterval = types.Int64Null()
	}

	if v, ok := GetString(result, "period_type"); ok {
		model.PeriodType = types.StringValue(v)
	}

	if v, ok := GetInt64(result, "fixed_period_start_day"); ok {
		model.FixedPeriodStartDay = types.Int64Value(v)
	} else {
		model.FixedPeriodStartDay = types.Int64Null()
	}

	if v, ok := GetInt64(result, "fixed_period_rollover_day"); ok {
		model.FixedPeriodRolloverDay = types.Int64Value(v)
	} else {
		model.FixedPeriodRolloverDay = types.Int64Null()
	}

	if v, ok := GetInt64(result, "relationship_type_id"); ok {
		model.RelationshipTypeID = types.Int64Value(v)
	} else {
		model.RelationshipTypeID = types.Int64Null()
	}

	if v, ok := GetString(result, "relationship_direction"); ok && v != "" {
		model.RelationshipDirection = types.StringValue(v)
	} else {
		model.RelationshipDirection = types.StringNull()
	}

	if v, ok := GetInt64(result, "max_related"); ok {
		model.MaxRelated = types.Int64Value(v)
	} else {
		model.MaxRelated = types.Int64Null()
	}

	if v, ok := GetString(result, "visibility"); ok && v != "" {
		model.Visibility = types.StringValue(v)
	} else {
		model.Visibility = types.StringNull()
	}

	if v, ok := GetInt64(result, "weight"); ok {
		model.Weight = types.Int64Value(v)
	} else {
		model.Weight = types.Int64Null()
	}

	if v, ok := GetString(result, "receipt_text_signup"); ok && v != "" {
		model.ReceiptTextSignup = types.StringValue(v)
	} else {
		model.ReceiptTextSignup = types.StringNull()
	}

	if v, ok := GetString(result, "receipt_text_renewal"); ok && v != "" {
		model.ReceiptTextRenewal = types.StringValue(v)
	} else {
		model.ReceiptTextRenewal = types.StringNull()
	}

	if v, ok := GetBool(result, "auto_renew"); ok {
		model.AutoRenew = types.BoolValue(v)
	}

	if v, ok := GetBool(result, "is_active"); ok {
		model.IsActive = types.BoolValue(v)
	}
}
