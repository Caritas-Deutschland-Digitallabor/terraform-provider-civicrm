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
	_ resource.Resource                = &PriceFieldResource{}
	_ resource.ResourceWithConfigure   = &PriceFieldResource{}
	_ resource.ResourceWithImportState = &PriceFieldResource{}
)

type PriceFieldResource struct {
	client *Client
}

type PriceFieldResourceModel struct {
	ID               types.Int64  `tfsdk:"id"`
	PriceSetID       types.Int64  `tfsdk:"price_set_id"`
	Name             types.String `tfsdk:"name"`
	Label            types.String `tfsdk:"label"`
	HtmlType         types.String `tfsdk:"html_type"`
	IsEnterQty       types.Bool   `tfsdk:"is_enter_qty"`
	HelpPre          types.String `tfsdk:"help_pre"`
	HelpPost         types.String `tfsdk:"help_post"`
	Weight           types.Int64  `tfsdk:"weight"`
	IsDisplayAmounts types.Bool   `tfsdk:"is_display_amounts"`
	OptionsPerLine   types.Int64  `tfsdk:"options_per_line"`
	IsActive         types.Bool   `tfsdk:"is_active"`
	IsRequired       types.Bool   `tfsdk:"is_required"`
	ActiveOn         types.String `tfsdk:"active_on"`
	ExpireOn         types.String `tfsdk:"expire_on"`
	Javascript       types.String `tfsdk:"javascript"`
	VisibilityID     types.Int64  `tfsdk:"visibility_id"`
}

func NewPriceFieldResource() resource.Resource {
	return &PriceFieldResource{}
}

func (r *PriceFieldResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_price_field"
}

func (r *PriceFieldResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a CiviCRM Price Field.",
		Attributes: map[string]schema.Attribute{
			"id": schema.Int64Attribute{
				Description: "The unique identifier of the price field.",
				Computed:    true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"price_set_id": schema.Int64Attribute{
				Description: "The price set this field belongs to.",
				Required:    true,
			},
			"name": schema.StringAttribute{
				Description: "Variable name/programmatic handle for this price field.",
				Required:    true,
			},
			"label": schema.StringAttribute{
				Description: "Displayed label for the price field.",
				Required:    true,
			},
			"html_type": schema.StringAttribute{
				Description: "HTML input type (Text, Select, Radio, CheckBox, Text, Button).",
				Required:    true,
			},
			"is_enter_qty": schema.BoolAttribute{
				Description: "Whether participants must enter a quantity for this field. Default: false.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
			},
			"help_pre": schema.StringAttribute{
				Description: "Help text displayed before the field.",
				Optional:    true,
				Computed:    true,
			},
			"help_post": schema.StringAttribute{
				Description: "Help text displayed after the field.",
				Optional:    true,
				Computed:    true,
			},
			"weight": schema.Int64Attribute{
				Description: "Order in which the fields are displayed.",
				Optional:    true,
				Computed:    true,
			},
			"is_display_amounts": schema.BoolAttribute{
				Description: "Whether to display amounts for this field. Default: true.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(true),
			},
			"options_per_line": schema.Int64Attribute{
				Description: "Number of options per line for Radio/CheckBox fields.",
				Optional:    true,
				Computed:    true,
			},
			"is_active": schema.BoolAttribute{
				Description: "Is this price field active. Default: true.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(true),
			},
			"is_required": schema.BoolAttribute{
				Description: "Is this price field required. Default: true.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(true),
			},
			"active_on": schema.StringAttribute{
				Description: "Date and time this field becomes active (ISO 8601).",
				Optional:    true,
				Computed:    true,
			},
			"expire_on": schema.StringAttribute{
				Description: "Date and time this field expires (ISO 8601).",
				Optional:    true,
				Computed:    true,
			},
			"javascript": schema.StringAttribute{
				Description: "Optional JavaScript for this price field.",
				Optional:    true,
				Computed:    true,
			},
			"visibility_id": schema.Int64Attribute{
				Description: "Visibility of this price field (1=public, 2=admin).",
				Optional:    true,
				Computed:    true,
			},
		},
	}
}

func (r *PriceFieldResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *PriceFieldResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan PriceFieldResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Creating PriceField", map[string]any{"name": plan.Name.ValueString()})

	values := map[string]any{
		"price_set_id":       plan.PriceSetID.ValueInt64(),
		"name":               plan.Name.ValueString(),
		"label":              plan.Label.ValueString(),
		"html_type":          plan.HtmlType.ValueString(),
		"is_enter_qty":       plan.IsEnterQty.ValueBool(),
		"is_display_amounts": plan.IsDisplayAmounts.ValueBool(),
		"is_active":          plan.IsActive.ValueBool(),
		"is_required":        plan.IsRequired.ValueBool(),
	}

	if !plan.HelpPre.IsNull() {
		values["help_pre"] = plan.HelpPre.ValueString()
	}
	if !plan.HelpPost.IsNull() {
		values["help_post"] = plan.HelpPost.ValueString()
	}
	if !plan.Weight.IsNull() {
		values["weight"] = plan.Weight.ValueInt64()
	}
	if !plan.OptionsPerLine.IsNull() {
		values["options_per_line"] = plan.OptionsPerLine.ValueInt64()
	}
	if !plan.ActiveOn.IsNull() {
		values["active_on"] = plan.ActiveOn.ValueString()
	}
	if !plan.ExpireOn.IsNull() {
		values["expire_on"] = plan.ExpireOn.ValueString()
	}
	if !plan.Javascript.IsNull() {
		values["javascript"] = plan.Javascript.ValueString()
	}
	if !plan.VisibilityID.IsNull() {
		values["visibility_id"] = plan.VisibilityID.ValueInt64()
	}

	result, err := r.client.Create("PriceField", values)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating PriceField",
			"Could not create PriceField: "+err.Error(),
		)
		return
	}

	r.mapResultToState(result, &plan)

	tflog.Debug(ctx, "Created PriceField", map[string]any{"id": plan.ID.ValueInt64()})

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

func (r *PriceFieldResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state PriceFieldResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Reading PriceField", map[string]any{"id": state.ID.ValueInt64()})

	result, err := r.client.GetByID("PriceField", state.ID.ValueInt64(), nil)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading PriceField",
			"Could not read PriceField ID "+strconv.FormatInt(state.ID.ValueInt64(), 10)+": "+err.Error(),
		)
		return
	}

	r.mapResultToState(result, &state)

	diags = resp.State.Set(ctx, state)
	resp.Diagnostics.Append(diags...)
}

func (r *PriceFieldResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan PriceFieldResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var state PriceFieldResourceModel
	diags = req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Updating PriceField", map[string]any{"id": state.ID.ValueInt64()})

	values := map[string]any{
		"price_set_id":       plan.PriceSetID.ValueInt64(),
		"name":               plan.Name.ValueString(),
		"label":              plan.Label.ValueString(),
		"html_type":          plan.HtmlType.ValueString(),
		"is_enter_qty":       plan.IsEnterQty.ValueBool(),
		"is_display_amounts": plan.IsDisplayAmounts.ValueBool(),
		"is_active":          plan.IsActive.ValueBool(),
		"is_required":        plan.IsRequired.ValueBool(),
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
	if !plan.Weight.IsNull() {
		values["weight"] = plan.Weight.ValueInt64()
	} else {
		values["weight"] = nil
	}
	if !plan.OptionsPerLine.IsNull() {
		values["options_per_line"] = plan.OptionsPerLine.ValueInt64()
	} else {
		values["options_per_line"] = nil
	}
	if !plan.ActiveOn.IsNull() {
		values["active_on"] = plan.ActiveOn.ValueString()
	} else {
		values["active_on"] = nil
	}
	if !plan.ExpireOn.IsNull() {
		values["expire_on"] = plan.ExpireOn.ValueString()
	} else {
		values["expire_on"] = nil
	}
	if !plan.Javascript.IsNull() {
		values["javascript"] = plan.Javascript.ValueString()
	} else {
		values["javascript"] = nil
	}
	if !plan.VisibilityID.IsNull() {
		values["visibility_id"] = plan.VisibilityID.ValueInt64()
	} else {
		values["visibility_id"] = nil
	}

	result, err := r.client.Update("PriceField", state.ID.ValueInt64(), values)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error updating PriceField",
			"Could not update PriceField ID "+strconv.FormatInt(state.ID.ValueInt64(), 10)+": "+err.Error(),
		)
		return
	}

	plan.ID = state.ID
	r.mapResultToState(result, &plan)

	tflog.Debug(ctx, "Updated PriceField", map[string]any{"id": plan.ID.ValueInt64()})

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

func (r *PriceFieldResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state PriceFieldResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Deleting PriceField", map[string]any{"id": state.ID.ValueInt64()})

	err := r.client.Delete("PriceField", state.ID.ValueInt64())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error deleting PriceField",
			"Could not delete PriceField ID "+strconv.FormatInt(state.ID.ValueInt64(), 10)+": "+err.Error(),
		)
		return
	}

	tflog.Debug(ctx, "Deleted PriceField", map[string]any{"id": state.ID.ValueInt64()})
}

func (r *PriceFieldResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
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

func (r *PriceFieldResource) mapResultToState(result map[string]any, model *PriceFieldResourceModel) {
	if id, ok := GetInt64(result, "id"); ok {
		model.ID = types.Int64Value(id)
	}

	if v, ok := GetInt64(result, "price_set_id"); ok {
		model.PriceSetID = types.Int64Value(v)
	}

	if v, ok := GetString(result, "name"); ok {
		model.Name = types.StringValue(v)
	}

	if v, ok := GetString(result, "label"); ok {
		model.Label = types.StringValue(v)
	}

	if v, ok := GetString(result, "html_type"); ok {
		model.HtmlType = types.StringValue(v)
	}

	if v, ok := GetBool(result, "is_enter_qty"); ok {
		model.IsEnterQty = types.BoolValue(v)
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

	if v, ok := GetInt64(result, "weight"); ok {
		model.Weight = types.Int64Value(v)
	} else {
		model.Weight = types.Int64Null()
	}

	if v, ok := GetBool(result, "is_display_amounts"); ok {
		model.IsDisplayAmounts = types.BoolValue(v)
	}

	if v, ok := GetInt64(result, "options_per_line"); ok {
		model.OptionsPerLine = types.Int64Value(v)
	} else {
		model.OptionsPerLine = types.Int64Null()
	}

	if v, ok := GetBool(result, "is_active"); ok {
		model.IsActive = types.BoolValue(v)
	}

	if v, ok := GetBool(result, "is_required"); ok {
		model.IsRequired = types.BoolValue(v)
	}

	if v, ok := GetString(result, "active_on"); ok && v != "" {
		model.ActiveOn = types.StringValue(v)
	} else {
		model.ActiveOn = types.StringNull()
	}

	if v, ok := GetString(result, "expire_on"); ok && v != "" {
		model.ExpireOn = types.StringValue(v)
	} else {
		model.ExpireOn = types.StringNull()
	}

	if v, ok := GetString(result, "javascript"); ok && v != "" {
		model.Javascript = types.StringValue(v)
	} else {
		model.Javascript = types.StringNull()
	}

	if v, ok := GetInt64(result, "visibility_id"); ok {
		model.VisibilityID = types.Int64Value(v)
	} else {
		model.VisibilityID = types.Int64Null()
	}
}
