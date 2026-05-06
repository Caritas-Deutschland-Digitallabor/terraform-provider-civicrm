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
	_ resource.Resource                = &ContactResource{}
	_ resource.ResourceWithConfigure   = &ContactResource{}
	_ resource.ResourceWithImportState = &ContactResource{}
)

type ContactResource struct {
	client *Client
}

type ContactResourceModel struct {
	ID                 types.Int64  `tfsdk:"id"`
	ContactType        types.String `tfsdk:"contact_type"`
	FirstName          types.String `tfsdk:"first_name"`
	LastName           types.String `tfsdk:"last_name"`
	OrganizationName   types.String `tfsdk:"organization_name"`
	HouseholdName      types.String `tfsdk:"household_name"`
	DisplayName        types.String `tfsdk:"display_name"`
	JobTitle           types.String `tfsdk:"job_title"`
	EmployerID         types.Int64  `tfsdk:"employer_id"`
	ExternalIdentifier types.String `tfsdk:"external_identifier"`
	ContactSubType     types.String `tfsdk:"contact_sub_type"`
	IsActive           types.Bool   `tfsdk:"is_active"`
}

func NewContactResource() resource.Resource {
	return &ContactResource{}
}

func (r *ContactResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_contact"
}

func (r *ContactResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a CiviCRM Contact (Individual, Organization, or Household). Use this to register persons, organizations, and departments that can be assigned case roles or linked to other entities.",
		Attributes: map[string]schema.Attribute{
			"id": schema.Int64Attribute{
				Description: "Unique contact ID assigned by CiviCRM.",
				Computed:    true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"contact_type": schema.StringAttribute{
				Description: "Contact type: 'Individual', 'Organization', or 'Household'. Cannot be changed after creation.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"first_name": schema.StringAttribute{
				Description: "First name (Individual only).",
				Optional:    true,
			},
			"last_name": schema.StringAttribute{
				Description: "Last name (Individual only).",
				Optional:    true,
			},
			"organization_name": schema.StringAttribute{
				Description: "Organization or department name (Organization only).",
				Optional:    true,
			},
			"household_name": schema.StringAttribute{
				Description: "Household name (Household only).",
				Optional:    true,
			},
			"display_name": schema.StringAttribute{
				Description: "Display name as computed/stored by CiviCRM.",
				Computed:    true,
			},
			"job_title": schema.StringAttribute{
				Description: "Job title (Individual only).",
				Optional:    true,
			},
			"employer_id": schema.Int64Attribute{
				Description: "ID of the employer Organization contact (Individual only).",
				Optional:    true,
			},
			"external_identifier": schema.StringAttribute{
				Description: "External identifier for cross-system reference.",
				Optional:    true,
			},
			"contact_sub_type": schema.StringAttribute{
				Description: "Contact subtype name.",
				Optional:    true,
			},
			"is_active": schema.BoolAttribute{
				Description: "Whether the contact is active. Default: true.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(true),
			},
		},
	}
}

func (r *ContactResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *ContactResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan ContactResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	values := map[string]any{
		"contact_type": plan.ContactType.ValueString(),
		"is_active":    plan.IsActive.ValueBool(),
	}

	if !plan.FirstName.IsNull() {
		values["first_name"] = plan.FirstName.ValueString()
	}
	if !plan.LastName.IsNull() {
		values["last_name"] = plan.LastName.ValueString()
	}
	if !plan.OrganizationName.IsNull() {
		values["organization_name"] = plan.OrganizationName.ValueString()
	}
	if !plan.HouseholdName.IsNull() {
		values["household_name"] = plan.HouseholdName.ValueString()
	}
	if !plan.JobTitle.IsNull() {
		values["job_title"] = plan.JobTitle.ValueString()
	}
	if !plan.EmployerID.IsNull() {
		values["employer_id"] = plan.EmployerID.ValueInt64()
	}
	if !plan.ExternalIdentifier.IsNull() {
		values["external_identifier"] = plan.ExternalIdentifier.ValueString()
	}
	if !plan.ContactSubType.IsNull() {
		values["contact_sub_type"] = plan.ContactSubType.ValueString()
	}

	tflog.Debug(ctx, "Creating contact", map[string]any{"contact_type": plan.ContactType.ValueString()})

	result, err := r.client.Create("Contact", values)
	if err != nil {
		resp.Diagnostics.AddError("Error creating contact", err.Error())
		return
	}

	r.mapResponseToModel(result, &plan)
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *ContactResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state ContactResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	result, err := r.client.GetByID("Contact", state.ID.ValueInt64(), nil)
	if err != nil {
		resp.Diagnostics.AddError("Error reading contact",
			"Could not read contact ID "+strconv.FormatInt(state.ID.ValueInt64(), 10)+": "+err.Error())
		return
	}

	r.mapResponseToModel(result, &state)
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (r *ContactResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan ContactResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	var state ContactResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	values := map[string]any{
		"is_active": plan.IsActive.ValueBool(),
	}

	if !plan.FirstName.IsNull() {
		values["first_name"] = plan.FirstName.ValueString()
	} else {
		values["first_name"] = nil
	}
	if !plan.LastName.IsNull() {
		values["last_name"] = plan.LastName.ValueString()
	} else {
		values["last_name"] = nil
	}
	if !plan.OrganizationName.IsNull() {
		values["organization_name"] = plan.OrganizationName.ValueString()
	} else {
		values["organization_name"] = nil
	}
	if !plan.HouseholdName.IsNull() {
		values["household_name"] = plan.HouseholdName.ValueString()
	} else {
		values["household_name"] = nil
	}
	if !plan.JobTitle.IsNull() {
		values["job_title"] = plan.JobTitle.ValueString()
	} else {
		values["job_title"] = nil
	}
	if !plan.EmployerID.IsNull() {
		values["employer_id"] = plan.EmployerID.ValueInt64()
	} else {
		values["employer_id"] = nil
	}
	if !plan.ExternalIdentifier.IsNull() {
		values["external_identifier"] = plan.ExternalIdentifier.ValueString()
	} else {
		values["external_identifier"] = nil
	}
	if !plan.ContactSubType.IsNull() {
		values["contact_sub_type"] = plan.ContactSubType.ValueString()
	} else {
		values["contact_sub_type"] = nil
	}

	result, err := r.client.Update("Contact", state.ID.ValueInt64(), values)
	if err != nil {
		resp.Diagnostics.AddError("Error updating contact",
			"Could not update contact ID "+strconv.FormatInt(state.ID.ValueInt64(), 10)+": "+err.Error())
		return
	}

	plan.ID = state.ID
	r.mapResponseToModel(result, &plan)
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *ContactResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state ContactResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.Delete("Contact", state.ID.ValueInt64())
	if err != nil {
		resp.Diagnostics.AddError("Error deleting contact",
			"Could not delete contact ID "+strconv.FormatInt(state.ID.ValueInt64(), 10)+": "+err.Error())
		return
	}
}

func (r *ContactResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	id, err := strconv.ParseInt(req.ID, 10, 64)
	if err != nil {
		resp.Diagnostics.AddError("Invalid import ID", "Could not parse import ID as integer: "+err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), id)...)
}

func (r *ContactResource) mapResponseToModel(result map[string]any, model *ContactResourceModel) {
	if id, ok := GetInt64(result, "id"); ok {
		model.ID = types.Int64Value(id)
	}
	if v, ok := GetString(result, "contact_type"); ok {
		model.ContactType = types.StringValue(v)
	}
	if v, ok := GetString(result, "first_name"); ok && v != "" {
		model.FirstName = types.StringValue(v)
	} else {
		model.FirstName = types.StringNull()
	}
	if v, ok := GetString(result, "last_name"); ok && v != "" {
		model.LastName = types.StringValue(v)
	} else {
		model.LastName = types.StringNull()
	}
	if v, ok := GetString(result, "organization_name"); ok && v != "" {
		model.OrganizationName = types.StringValue(v)
	} else {
		model.OrganizationName = types.StringNull()
	}
	if v, ok := GetString(result, "household_name"); ok && v != "" {
		model.HouseholdName = types.StringValue(v)
	} else {
		model.HouseholdName = types.StringNull()
	}
	if v, ok := GetString(result, "display_name"); ok {
		model.DisplayName = types.StringValue(v)
	}
	if v, ok := GetString(result, "job_title"); ok && v != "" {
		model.JobTitle = types.StringValue(v)
	} else {
		model.JobTitle = types.StringNull()
	}
	if v, ok := GetInt64(result, "employer_id"); ok && v != 0 {
		model.EmployerID = types.Int64Value(v)
	} else {
		model.EmployerID = types.Int64Null()
	}
	if v, ok := GetString(result, "external_identifier"); ok && v != "" {
		model.ExternalIdentifier = types.StringValue(v)
	} else {
		model.ExternalIdentifier = types.StringNull()
	}
	if v, ok := GetString(result, "contact_sub_type"); ok && v != "" {
		model.ContactSubType = types.StringValue(v)
	} else {
		model.ContactSubType = types.StringNull()
	}
	if v, ok := GetBool(result, "is_active"); ok {
		model.IsActive = types.BoolValue(v)
	}
}
