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
	_ resource.Resource                = &AddressResource{}
	_ resource.ResourceWithConfigure   = &AddressResource{}
	_ resource.ResourceWithImportState = &AddressResource{}
)

type AddressResource struct {
	client *Client
}

type AddressResourceModel struct {
	ID                   types.Int64  `tfsdk:"id"`
	ContactID            types.Int64  `tfsdk:"contact_id"`
	LocationTypeID       types.Int64  `tfsdk:"location_type_id"`
	IsPrimary            types.Bool   `tfsdk:"is_primary"`
	IsBilling            types.Bool   `tfsdk:"is_billing"`
	StreetAddress        types.String `tfsdk:"street_address"`
	SupplementalAddress1 types.String `tfsdk:"supplemental_address_1"`
	SupplementalAddress2 types.String `tfsdk:"supplemental_address_2"`
	City                 types.String `tfsdk:"city"`
	PostalCode           types.String `tfsdk:"postal_code"`
	StateProvinceID      types.Int64  `tfsdk:"state_province_id"`
	CountryID            types.Int64  `tfsdk:"country_id"`
}

func NewAddressResource() resource.Resource {
	return &AddressResource{}
}

func (r *AddressResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_address"
}

func (r *AddressResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a CiviCRM Address associated with a contact.",
		Attributes: map[string]schema.Attribute{
			"id": schema.Int64Attribute{
				Description: "Unique address ID assigned by CiviCRM.",
				Computed:    true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"contact_id": schema.Int64Attribute{
				Description: "ID of the contact this address belongs to.",
				Required:    true,
			},
			"location_type_id": schema.Int64Attribute{
				Description: "Location type ID (e.g. 1=Home, 2=Work). Use civicrm_option_value to look up IDs.",
				Optional:    true,
			},
			"is_primary": schema.BoolAttribute{
				Description: "Whether this is the primary address for the contact. Default: false.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
			},
			"is_billing": schema.BoolAttribute{
				Description: "Whether this is the billing address for the contact. Default: false.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
			},
			"street_address": schema.StringAttribute{
				Description: "Street address line.",
				Optional:    true,
			},
			"supplemental_address_1": schema.StringAttribute{
				Description: "Additional address line 1 (e.g. apartment, suite).",
				Optional:    true,
			},
			"supplemental_address_2": schema.StringAttribute{
				Description: "Additional address line 2.",
				Optional:    true,
			},
			"city": schema.StringAttribute{
				Description: "City.",
				Optional:    true,
			},
			"postal_code": schema.StringAttribute{
				Description: "Postal/ZIP code.",
				Optional:    true,
			},
			"state_province_id": schema.Int64Attribute{
				Description: "State or province ID.",
				Optional:    true,
			},
			"country_id": schema.Int64Attribute{
				Description: "Country ID (CiviCRM numeric country ID, e.g. 1082 for Germany).",
				Optional:    true,
			},
		},
	}
}

func (r *AddressResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *AddressResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan AddressResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	values := map[string]any{
		"contact_id": plan.ContactID.ValueInt64(),
		"is_primary": plan.IsPrimary.ValueBool(),
		"is_billing": plan.IsBilling.ValueBool(),
	}

	if !plan.LocationTypeID.IsNull() {
		values["location_type_id"] = plan.LocationTypeID.ValueInt64()
	}
	if !plan.StreetAddress.IsNull() {
		values["street_address"] = plan.StreetAddress.ValueString()
	}
	if !plan.SupplementalAddress1.IsNull() {
		values["supplemental_address_1"] = plan.SupplementalAddress1.ValueString()
	}
	if !plan.SupplementalAddress2.IsNull() {
		values["supplemental_address_2"] = plan.SupplementalAddress2.ValueString()
	}
	if !plan.City.IsNull() {
		values["city"] = plan.City.ValueString()
	}
	if !plan.PostalCode.IsNull() {
		values["postal_code"] = plan.PostalCode.ValueString()
	}
	if !plan.StateProvinceID.IsNull() {
		values["state_province_id"] = plan.StateProvinceID.ValueInt64()
	}
	if !plan.CountryID.IsNull() {
		values["country_id"] = plan.CountryID.ValueInt64()
	}

	tflog.Debug(ctx, "Creating address", map[string]any{"contact_id": plan.ContactID.ValueInt64()})

	result, err := r.client.Create("Address", values)
	if err != nil {
		resp.Diagnostics.AddError("Error creating address", err.Error())
		return
	}

	r.mapResponseToModel(result, &plan)
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *AddressResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state AddressResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	result, err := r.client.GetByID("Address", state.ID.ValueInt64(), nil)
	if err != nil {
		resp.Diagnostics.AddError("Error reading address",
			"Could not read address ID "+strconv.FormatInt(state.ID.ValueInt64(), 10)+": "+err.Error())
		return
	}

	r.mapResponseToModel(result, &state)
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (r *AddressResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan AddressResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	var state AddressResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	values := map[string]any{
		"contact_id": plan.ContactID.ValueInt64(),
		"is_primary": plan.IsPrimary.ValueBool(),
		"is_billing": plan.IsBilling.ValueBool(),
	}

	if !plan.LocationTypeID.IsNull() {
		values["location_type_id"] = plan.LocationTypeID.ValueInt64()
	} else {
		values["location_type_id"] = nil
	}
	if !plan.StreetAddress.IsNull() {
		values["street_address"] = plan.StreetAddress.ValueString()
	} else {
		values["street_address"] = nil
	}
	if !plan.SupplementalAddress1.IsNull() {
		values["supplemental_address_1"] = plan.SupplementalAddress1.ValueString()
	} else {
		values["supplemental_address_1"] = nil
	}
	if !plan.SupplementalAddress2.IsNull() {
		values["supplemental_address_2"] = plan.SupplementalAddress2.ValueString()
	} else {
		values["supplemental_address_2"] = nil
	}
	if !plan.City.IsNull() {
		values["city"] = plan.City.ValueString()
	} else {
		values["city"] = nil
	}
	if !plan.PostalCode.IsNull() {
		values["postal_code"] = plan.PostalCode.ValueString()
	} else {
		values["postal_code"] = nil
	}
	if !plan.StateProvinceID.IsNull() {
		values["state_province_id"] = plan.StateProvinceID.ValueInt64()
	} else {
		values["state_province_id"] = nil
	}
	if !plan.CountryID.IsNull() {
		values["country_id"] = plan.CountryID.ValueInt64()
	} else {
		values["country_id"] = nil
	}

	result, err := r.client.Update("Address", state.ID.ValueInt64(), values)
	if err != nil {
		resp.Diagnostics.AddError("Error updating address",
			"Could not update address ID "+strconv.FormatInt(state.ID.ValueInt64(), 10)+": "+err.Error())
		return
	}

	plan.ID = state.ID
	r.mapResponseToModel(result, &plan)
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *AddressResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state AddressResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.Delete("Address", state.ID.ValueInt64())
	if err != nil {
		resp.Diagnostics.AddError("Error deleting address",
			"Could not delete address ID "+strconv.FormatInt(state.ID.ValueInt64(), 10)+": "+err.Error())
		return
	}
}

func (r *AddressResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	id, err := strconv.ParseInt(req.ID, 10, 64)
	if err != nil {
		resp.Diagnostics.AddError("Invalid import ID", "Could not parse import ID as integer: "+err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), id)...)
}

func (r *AddressResource) mapResponseToModel(result map[string]any, model *AddressResourceModel) {
	if id, ok := GetInt64(result, "id"); ok {
		model.ID = types.Int64Value(id)
	}
	if v, ok := GetInt64(result, "contact_id"); ok {
		model.ContactID = types.Int64Value(v)
	}
	if v, ok := GetInt64(result, "location_type_id"); ok && v != 0 {
		model.LocationTypeID = types.Int64Value(v)
	} else {
		model.LocationTypeID = types.Int64Null()
	}
	if v, ok := GetBool(result, "is_primary"); ok {
		model.IsPrimary = types.BoolValue(v)
	}
	if v, ok := GetBool(result, "is_billing"); ok {
		model.IsBilling = types.BoolValue(v)
	}
	if v, ok := GetString(result, "street_address"); ok && v != "" {
		model.StreetAddress = types.StringValue(v)
	} else {
		model.StreetAddress = types.StringNull()
	}
	if v, ok := GetString(result, "supplemental_address_1"); ok && v != "" {
		model.SupplementalAddress1 = types.StringValue(v)
	} else {
		model.SupplementalAddress1 = types.StringNull()
	}
	if v, ok := GetString(result, "supplemental_address_2"); ok && v != "" {
		model.SupplementalAddress2 = types.StringValue(v)
	} else {
		model.SupplementalAddress2 = types.StringNull()
	}
	if v, ok := GetString(result, "city"); ok && v != "" {
		model.City = types.StringValue(v)
	} else {
		model.City = types.StringNull()
	}
	if v, ok := GetString(result, "postal_code"); ok && v != "" {
		model.PostalCode = types.StringValue(v)
	} else {
		model.PostalCode = types.StringNull()
	}
	if v, ok := GetInt64(result, "state_province_id"); ok && v != 0 {
		model.StateProvinceID = types.Int64Value(v)
	} else {
		model.StateProvinceID = types.Int64Null()
	}
	if v, ok := GetInt64(result, "country_id"); ok && v != 0 {
		model.CountryID = types.Int64Value(v)
	} else {
		model.CountryID = types.Int64Null()
	}
}
