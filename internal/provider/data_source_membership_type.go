package provider

import (
	"context"
	"fmt"
	"strconv"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var _ datasource.DataSource = &MembershipTypeDataSource{}
var _ datasource.DataSourceWithConfigure = &MembershipTypeDataSource{}

type MembershipTypeDataSource struct {
	client *Client
}

type MembershipTypeDataSourceModel struct {
	ID                     types.Int64   `tfsdk:"id"`
	Name                   types.String  `tfsdk:"name"`
	Description            types.String  `tfsdk:"description"`
	MemberOfContactID      types.Int64   `tfsdk:"member_of_contact_id"`
	FinancialTypeID        types.Int64   `tfsdk:"financial_type_id"`
	MinimumFee             types.Float64 `tfsdk:"minimum_fee"`
	DurationUnit           types.String  `tfsdk:"duration_unit"`
	DurationInterval       types.Int64   `tfsdk:"duration_interval"`
	PeriodType             types.String  `tfsdk:"period_type"`
	FixedPeriodStartDay    types.Int64   `tfsdk:"fixed_period_start_day"`
	FixedPeriodRolloverDay types.Int64   `tfsdk:"fixed_period_rollover_day"`
	RelationshipTypeID     types.Int64   `tfsdk:"relationship_type_id"`
	RelationshipDirection  types.String  `tfsdk:"relationship_direction"`
	MaxRelated             types.Int64   `tfsdk:"max_related"`
	Visibility             types.String  `tfsdk:"visibility"`
	Weight                 types.Int64   `tfsdk:"weight"`
	ReceiptTextSignup      types.String  `tfsdk:"receipt_text_signup"`
	ReceiptTextRenewal     types.String  `tfsdk:"receipt_text_renewal"`
	AutoRenew              types.Bool    `tfsdk:"auto_renew"`
	IsActive               types.Bool    `tfsdk:"is_active"`
}

func NewMembershipTypeDataSource() datasource.DataSource {
	return &MembershipTypeDataSource{}
}

func (d *MembershipTypeDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_membership_type"
}

func (d *MembershipTypeDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Fetches a CiviCRM Membership Type by ID or name.",
		Attributes: map[string]schema.Attribute{
			"id": schema.Int64Attribute{
				Description: "The unique identifier. Specify either id or name.",
				Optional:    true,
				Computed:    true,
			},
			"name": schema.StringAttribute{
				Description: "The name of the membership type. Specify either id or name.",
				Optional:    true,
				Computed:    true,
			},
			"description":              schema.StringAttribute{Description: "Description.", Computed: true},
			"member_of_contact_id":     schema.Int64Attribute{Description: "The contact ID of the owning organization.", Computed: true},
			"financial_type_id":        schema.Int64Attribute{Description: "The financial type ID.", Computed: true},
			"minimum_fee":              schema.Float64Attribute{Description: "The minimum fee.", Computed: true},
			"duration_unit":            schema.StringAttribute{Description: "Duration unit (day, month, year, lifetime).", Computed: true},
			"duration_interval":        schema.Int64Attribute{Description: "Duration interval count.", Computed: true},
			"period_type":              schema.StringAttribute{Description: "Period type (rolling or fixed).", Computed: true},
			"fixed_period_start_day":   schema.Int64Attribute{Description: "Fixed period start day (MMDD).", Computed: true},
			"fixed_period_rollover_day": schema.Int64Attribute{Description: "Fixed period rollover day (MMDD).", Computed: true},
			"relationship_type_id":     schema.Int64Attribute{Description: "Relationship type ID for inherited memberships.", Computed: true},
			"relationship_direction":   schema.StringAttribute{Description: "Relationship direction for inherited memberships.", Computed: true},
			"max_related":              schema.Int64Attribute{Description: "Maximum number of related memberships.", Computed: true},
			"visibility":               schema.StringAttribute{Description: "Visibility (Public or Admin).", Computed: true},
			"weight":                   schema.Int64Attribute{Description: "Sort weight.", Computed: true},
			"receipt_text_signup":      schema.StringAttribute{Description: "Signup receipt text.", Computed: true},
			"receipt_text_renewal":     schema.StringAttribute{Description: "Renewal receipt text.", Computed: true},
			"auto_renew":               schema.BoolAttribute{Description: "Whether auto-renewal is enabled.", Computed: true},
			"is_active":                schema.BoolAttribute{Description: "Whether the membership type is active.", Computed: true},
		},
	}
}

func (d *MembershipTypeDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	client, ok := req.ProviderData.(*Client)
	if !ok {
		resp.Diagnostics.AddError("Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *Client, got: %T.", req.ProviderData))
		return
	}
	d.client = client
}

func (d *MembershipTypeDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config MembershipTypeDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if config.ID.IsNull() && config.Name.IsNull() {
		resp.Diagnostics.AddError("Missing Filter", "At least one of 'id' or 'name' must be specified.")
		return
	}

	where := [][]any{}
	if !config.ID.IsNull() {
		where = append(where, []any{"id", "=", config.ID.ValueInt64()})
	}
	if !config.Name.IsNull() {
		where = append(where, []any{"name", "=", config.Name.ValueString()})
	}

	tflog.Debug(ctx, "Reading MembershipType data source", map[string]any{"where": where})

	results, err := d.client.Get("MembershipType", where, nil)
	if err != nil {
		resp.Diagnostics.AddError("Error reading MembershipType", err.Error())
		return
	}
	if len(results) == 0 {
		resp.Diagnostics.AddError("MembershipType not found", "No membership type found matching the specified criteria.")
		return
	}

	result := results[0]

	if id, ok := GetInt64(result, "id"); ok {
		config.ID = types.Int64Value(id)
	}
	if v, ok := GetString(result, "name"); ok {
		config.Name = types.StringValue(v)
	}
	if v, ok := GetString(result, "description"); ok && v != "" {
		config.Description = types.StringValue(v)
	} else {
		config.Description = types.StringNull()
	}
	if v, ok := GetInt64(result, "member_of_contact_id"); ok {
		config.MemberOfContactID = types.Int64Value(v)
	}
	if v, ok := GetInt64(result, "financial_type_id"); ok {
		config.FinancialTypeID = types.Int64Value(v)
	}
	if v, ok := result["minimum_fee"]; ok && v != nil {
		switch val := v.(type) {
		case float64:
			config.MinimumFee = types.Float64Value(val)
		case string:
			if f, err := strconv.ParseFloat(val, 64); err == nil {
				config.MinimumFee = types.Float64Value(f)
			} else {
				config.MinimumFee = types.Float64Null()
			}
		default:
			config.MinimumFee = types.Float64Null()
		}
	} else {
		config.MinimumFee = types.Float64Null()
	}
	if v, ok := GetString(result, "duration_unit"); ok {
		config.DurationUnit = types.StringValue(v)
	}
	if v, ok := GetInt64(result, "duration_interval"); ok {
		config.DurationInterval = types.Int64Value(v)
	} else {
		config.DurationInterval = types.Int64Null()
	}
	if v, ok := GetString(result, "period_type"); ok {
		config.PeriodType = types.StringValue(v)
	}
	if v, ok := GetInt64(result, "fixed_period_start_day"); ok {
		config.FixedPeriodStartDay = types.Int64Value(v)
	} else {
		config.FixedPeriodStartDay = types.Int64Null()
	}
	if v, ok := GetInt64(result, "fixed_period_rollover_day"); ok {
		config.FixedPeriodRolloverDay = types.Int64Value(v)
	} else {
		config.FixedPeriodRolloverDay = types.Int64Null()
	}
	if v, ok := GetInt64(result, "relationship_type_id"); ok {
		config.RelationshipTypeID = types.Int64Value(v)
	} else {
		config.RelationshipTypeID = types.Int64Null()
	}
	if v, ok := GetString(result, "relationship_direction"); ok && v != "" {
		config.RelationshipDirection = types.StringValue(v)
	} else {
		config.RelationshipDirection = types.StringNull()
	}
	if v, ok := GetInt64(result, "max_related"); ok {
		config.MaxRelated = types.Int64Value(v)
	} else {
		config.MaxRelated = types.Int64Null()
	}
	if v, ok := GetString(result, "visibility"); ok && v != "" {
		config.Visibility = types.StringValue(v)
	} else {
		config.Visibility = types.StringNull()
	}
	if v, ok := GetInt64(result, "weight"); ok {
		config.Weight = types.Int64Value(v)
	} else {
		config.Weight = types.Int64Null()
	}
	if v, ok := GetString(result, "receipt_text_signup"); ok && v != "" {
		config.ReceiptTextSignup = types.StringValue(v)
	} else {
		config.ReceiptTextSignup = types.StringNull()
	}
	if v, ok := GetString(result, "receipt_text_renewal"); ok && v != "" {
		config.ReceiptTextRenewal = types.StringValue(v)
	} else {
		config.ReceiptTextRenewal = types.StringNull()
	}
	if v, ok := GetBool(result, "auto_renew"); ok {
		config.AutoRenew = types.BoolValue(v)
	}
	if v, ok := GetBool(result, "is_active"); ok {
		config.IsActive = types.BoolValue(v)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, config)...)
}
