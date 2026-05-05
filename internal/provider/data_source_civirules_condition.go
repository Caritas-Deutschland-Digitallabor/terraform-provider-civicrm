package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var _ datasource.DataSource = &CiviRulesConditionDataSource{}
var _ datasource.DataSourceWithConfigure = &CiviRulesConditionDataSource{}

// CiviRulesConditionDataSource looks up a CiviRulesCondition type definition.
// Use this to resolve condition_id for civicrm_civirules_rule_condition
// without hardcoding numeric IDs.
type CiviRulesConditionDataSource struct {
	client *Client
}

type CiviRulesConditionDataSourceModel struct {
	ID        types.Int64  `tfsdk:"id"`
	Name      types.String `tfsdk:"name"`
	Label     types.String `tfsdk:"label"`
	ClassName types.String `tfsdk:"class_name"`
	IsActive  types.Bool   `tfsdk:"is_active"`
}

func NewCiviRulesConditionDataSource() datasource.DataSource {
	return &CiviRulesConditionDataSource{}
}

func (d *CiviRulesConditionDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_civirules_condition"
}

func (d *CiviRulesConditionDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Looks up a CiviRules Condition type by ID or name. Use the returned `id` as condition_id in civicrm_civirules_rule_condition.",
		Attributes: map[string]schema.Attribute{
			"id": schema.Int64Attribute{
				Description: "Unique ID of the condition type. Specify either id or name.",
				Optional:    true,
				Computed:    true,
			},
			"name": schema.StringAttribute{
				Description: "Machine name of the condition type (e.g. 'activity_type'). Specify either id or name.",
				Optional:    true,
				Computed:    true,
			},
			"label":      schema.StringAttribute{Description: "Display label shown in the CiviRules UI.", Computed: true},
			"class_name": schema.StringAttribute{Description: "PHP class implementing the condition.", Computed: true},
			"is_active":  schema.BoolAttribute{Description: "Whether active.", Computed: true},
		},
	}
}

func (d *CiviRulesConditionDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *CiviRulesConditionDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config CiviRulesConditionDataSourceModel
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

	tflog.Debug(ctx, "Reading CiviRulesCondition data source", map[string]any{"where": where})

	results, err := d.client.Get("CiviRulesCondition", where, nil)
	if err != nil {
		resp.Diagnostics.AddError("Error reading CiviRulesCondition", err.Error())
		return
	}
	if len(results) == 0 {
		resp.Diagnostics.AddError("CiviRulesCondition not found", "No condition type found matching the specified criteria.")
		return
	}

	result := results[0]

	if id, ok := GetInt64(result, "id"); ok {
		config.ID = types.Int64Value(id)
	}
	if v, ok := GetString(result, "name"); ok {
		config.Name = types.StringValue(v)
	}
	if v, ok := GetString(result, "label"); ok {
		config.Label = types.StringValue(v)
	}
	if v, ok := GetString(result, "class_name"); ok && v != "" {
		config.ClassName = types.StringValue(v)
	} else {
		config.ClassName = types.StringNull()
	}
	if v, ok := GetBool(result, "is_active"); ok {
		config.IsActive = types.BoolValue(v)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, config)...)
}
