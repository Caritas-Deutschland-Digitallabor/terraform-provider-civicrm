package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var _ datasource.DataSource = &CiviRulesRuleDataSource{}
var _ datasource.DataSourceWithConfigure = &CiviRulesRuleDataSource{}

type CiviRulesRuleDataSource struct {
	client *Client
}

type CiviRulesRuleDataSourceModel struct {
	ID            types.Int64  `tfsdk:"id"`
	Name          types.String `tfsdk:"name"`
	Label         types.String `tfsdk:"label"`
	TriggerID     types.Int64  `tfsdk:"trigger_id"`
	TriggerParams types.String `tfsdk:"trigger_params"`
	Description   types.String `tfsdk:"description"`
	HelpText      types.String `tfsdk:"help_text"`
	IsActive      types.Bool   `tfsdk:"is_active"`
	IsDebug       types.Bool   `tfsdk:"is_debug"`
}

func NewCiviRulesRuleDataSource() datasource.DataSource {
	return &CiviRulesRuleDataSource{}
}

func (d *CiviRulesRuleDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_civirules_rule"
}

func (d *CiviRulesRuleDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Fetches a CiviRules Rule by ID or name.",
		Attributes: map[string]schema.Attribute{
			"id": schema.Int64Attribute{
				Description: "Unique ID of the rule. Specify either id or name.",
				Optional:    true,
				Computed:    true,
			},
			"name": schema.StringAttribute{
				Description: "Machine name of the rule. Specify either id or name.",
				Optional:    true,
				Computed:    true,
			},
			"label":          schema.StringAttribute{Description: "Display label.", Computed: true},
			"trigger_id":     schema.Int64Attribute{Description: "ID of the trigger.", Computed: true},
			"trigger_params": schema.StringAttribute{Description: "Trigger parameters as JSON.", Computed: true},
			"description":    schema.StringAttribute{Description: "Description.", Computed: true},
			"help_text":      schema.StringAttribute{Description: "Help text.", Computed: true},
			"is_active":      schema.BoolAttribute{Description: "Whether active.", Computed: true},
			"is_debug":       schema.BoolAttribute{Description: "Whether debug mode is on.", Computed: true},
		},
	}
}

func (d *CiviRulesRuleDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *CiviRulesRuleDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config CiviRulesRuleDataSourceModel
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

	tflog.Debug(ctx, "Reading CiviRulesRule data source", map[string]any{"where": where})

	results, err := d.client.Get("CiviRulesRule", where, nil)
	if err != nil {
		resp.Diagnostics.AddError("Error reading CiviRulesRule", err.Error())
		return
	}
	if len(results) == 0 {
		resp.Diagnostics.AddError("CiviRulesRule not found", "No rule found matching the specified criteria.")
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
	if v, ok := GetInt64(result, "trigger_id"); ok {
		config.TriggerID = types.Int64Value(v)
	}
	if v, ok := GetString(result, "trigger_params"); ok && v != "" {
		config.TriggerParams = types.StringValue(v)
	} else {
		config.TriggerParams = types.StringNull()
	}
	if v, ok := GetString(result, "description"); ok && v != "" {
		config.Description = types.StringValue(v)
	} else {
		config.Description = types.StringNull()
	}
	if v, ok := GetString(result, "help_text"); ok && v != "" {
		config.HelpText = types.StringValue(v)
	} else {
		config.HelpText = types.StringNull()
	}
	if v, ok := GetBool(result, "is_active"); ok {
		config.IsActive = types.BoolValue(v)
	}
	if v, ok := GetBool(result, "is_debug"); ok {
		config.IsDebug = types.BoolValue(v)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, config)...)
}
