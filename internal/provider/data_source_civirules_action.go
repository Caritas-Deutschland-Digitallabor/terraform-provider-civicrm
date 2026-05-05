package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var _ datasource.DataSource = &CiviRulesActionDataSource{}
var _ datasource.DataSourceWithConfigure = &CiviRulesActionDataSource{}

// CiviRulesActionDataSource looks up a CiviRulesAction type definition.
// Use this to resolve action_id for civicrm_civirules_rule_action
// without hardcoding numeric IDs.
type CiviRulesActionDataSource struct {
	client *Client
}

type CiviRulesActionDataSourceModel struct {
	ID        types.Int64  `tfsdk:"id"`
	Name      types.String `tfsdk:"name"`
	Label     types.String `tfsdk:"label"`
	ClassName types.String `tfsdk:"class_name"`
	IsActive  types.Bool   `tfsdk:"is_active"`
}

func NewCiviRulesActionDataSource() datasource.DataSource {
	return &CiviRulesActionDataSource{}
}

func (d *CiviRulesActionDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_civirules_action"
}

func (d *CiviRulesActionDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Looks up a CiviRules Action type by ID or name. Use the returned `id` as action_id in civicrm_civirules_rule_action.",
		Attributes: map[string]schema.Attribute{
			"id": schema.Int64Attribute{
				Description: "Unique ID of the action type. Specify either id or name.",
				Optional:    true,
				Computed:    true,
			},
			"name": schema.StringAttribute{
				Description: "Machine name of the action type (e.g. 'change_case_status'). Specify either id or name.",
				Optional:    true,
				Computed:    true,
			},
			"label":      schema.StringAttribute{Description: "Display label shown in the CiviRules UI.", Computed: true},
			"class_name": schema.StringAttribute{Description: "PHP class implementing the action.", Computed: true},
			"is_active":  schema.BoolAttribute{Description: "Whether active.", Computed: true},
		},
	}
}

func (d *CiviRulesActionDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *CiviRulesActionDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config CiviRulesActionDataSourceModel
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

	tflog.Debug(ctx, "Reading CiviRulesAction data source", map[string]any{"where": where})

	results, err := d.client.Get("CiviRulesAction", where, nil)
	if err != nil {
		resp.Diagnostics.AddError("Error reading CiviRulesAction", err.Error())
		return
	}
	if len(results) == 0 {
		resp.Diagnostics.AddError("CiviRulesAction not found", "No action type found matching the specified criteria.")
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
