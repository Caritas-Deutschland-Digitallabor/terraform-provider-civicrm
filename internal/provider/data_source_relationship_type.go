package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var _ datasource.DataSource = &RelationshipTypeDataSource{}
var _ datasource.DataSourceWithConfigure = &RelationshipTypeDataSource{}

type RelationshipTypeDataSource struct {
	client *Client
}

type RelationshipTypeDataSourceModel struct {
	ID             types.Int64  `tfsdk:"id"`
	NameAB         types.String `tfsdk:"name_a_b"`
	LabelAB        types.String `tfsdk:"label_a_b"`
	NameBA         types.String `tfsdk:"name_b_a"`
	LabelBA        types.String `tfsdk:"label_b_a"`
	Description    types.String `tfsdk:"description"`
	ContactTypeA   types.String `tfsdk:"contact_type_a"`
	ContactTypeB   types.String `tfsdk:"contact_type_b"`
	ContactSubTypeA types.String `tfsdk:"contact_sub_type_a"`
	ContactSubTypeB types.String `tfsdk:"contact_sub_type_b"`
	IsReserved     types.Bool   `tfsdk:"is_reserved"`
	IsActive       types.Bool   `tfsdk:"is_active"`
}

func NewRelationshipTypeDataSource() datasource.DataSource {
	return &RelationshipTypeDataSource{}
}

func (d *RelationshipTypeDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_relationship_type"
}

func (d *RelationshipTypeDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Fetches a CiviCRM RelationshipType by ID or name_a_b. Use this to reference existing relationship types (e.g. case roles) without managing them via Terraform.",
		Attributes: map[string]schema.Attribute{
			"id": schema.Int64Attribute{
				Description: "Unique ID of the relationship type. Specify either id or name_a_b.",
				Optional:    true,
				Computed:    true,
			},
			"name_a_b": schema.StringAttribute{
				Description: "Machine name from A→B (e.g. 'Housing Coordinator'). Specify either id or name_a_b.",
				Optional:    true,
				Computed:    true,
			},
			"label_a_b":        schema.StringAttribute{Description: "Display label from A→B.", Computed: true},
			"name_b_a":         schema.StringAttribute{Description: "Machine name from B→A.", Computed: true},
			"label_b_a":        schema.StringAttribute{Description: "Display label from B→A.", Computed: true},
			"description":      schema.StringAttribute{Description: "Description.", Computed: true},
			"contact_type_a":   schema.StringAttribute{Description: "Required contact type for side A, or empty for any.", Computed: true},
			"contact_type_b":   schema.StringAttribute{Description: "Required contact type for side B, or empty for any.", Computed: true},
			"contact_sub_type_a": schema.StringAttribute{Description: "Required contact subtype for side A.", Computed: true},
			"contact_sub_type_b": schema.StringAttribute{Description: "Required contact subtype for side B.", Computed: true},
			"is_reserved":      schema.BoolAttribute{Description: "Whether this is a system-reserved type.", Computed: true},
			"is_active":        schema.BoolAttribute{Description: "Whether active.", Computed: true},
		},
	}
}

func (d *RelationshipTypeDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *RelationshipTypeDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config RelationshipTypeDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if config.ID.IsNull() && config.NameAB.IsNull() {
		resp.Diagnostics.AddError("Missing Filter", "At least one of 'id' or 'name_a_b' must be specified.")
		return
	}

	where := [][]any{}
	if !config.ID.IsNull() {
		where = append(where, []any{"id", "=", config.ID.ValueInt64()})
	}
	if !config.NameAB.IsNull() {
		where = append(where, []any{"name_a_b", "=", config.NameAB.ValueString()})
	}

	tflog.Debug(ctx, "Reading RelationshipType data source", map[string]any{"where": where})

	results, err := d.client.Get("RelationshipType", where, nil)
	if err != nil {
		resp.Diagnostics.AddError("Error reading RelationshipType", err.Error())
		return
	}
	if len(results) == 0 {
		resp.Diagnostics.AddError("RelationshipType not found", "No relationship type found matching the specified criteria.")
		return
	}

	result := results[0]

	if id, ok := GetInt64(result, "id"); ok {
		config.ID = types.Int64Value(id)
	}
	if v, ok := GetString(result, "name_a_b"); ok {
		config.NameAB = types.StringValue(v)
	}
	if v, ok := GetString(result, "label_a_b"); ok {
		config.LabelAB = types.StringValue(v)
	}
	if v, ok := GetString(result, "name_b_a"); ok {
		config.NameBA = types.StringValue(v)
	}
	if v, ok := GetString(result, "label_b_a"); ok {
		config.LabelBA = types.StringValue(v)
	}
	if v, ok := GetString(result, "description"); ok && v != "" {
		config.Description = types.StringValue(v)
	} else {
		config.Description = types.StringNull()
	}
	if v, ok := GetString(result, "contact_type_a"); ok && v != "" {
		config.ContactTypeA = types.StringValue(v)
	} else {
		config.ContactTypeA = types.StringNull()
	}
	if v, ok := GetString(result, "contact_type_b"); ok && v != "" {
		config.ContactTypeB = types.StringValue(v)
	} else {
		config.ContactTypeB = types.StringNull()
	}
	if v, ok := GetString(result, "contact_sub_type_a"); ok && v != "" {
		config.ContactSubTypeA = types.StringValue(v)
	} else {
		config.ContactSubTypeA = types.StringNull()
	}
	if v, ok := GetString(result, "contact_sub_type_b"); ok && v != "" {
		config.ContactSubTypeB = types.StringValue(v)
	} else {
		config.ContactSubTypeB = types.StringNull()
	}
	if v, ok := GetBool(result, "is_reserved"); ok {
		config.IsReserved = types.BoolValue(v)
	}
	if v, ok := GetBool(result, "is_active"); ok {
		config.IsActive = types.BoolValue(v)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, config)...)
}
