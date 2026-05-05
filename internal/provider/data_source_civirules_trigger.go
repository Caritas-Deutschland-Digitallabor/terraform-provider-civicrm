package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var _ datasource.DataSource = &CiviRulesTriggerDataSource{}
var _ datasource.DataSourceWithConfigure = &CiviRulesTriggerDataSource{}

type CiviRulesTriggerDataSource struct {
	client *Client
}

type CiviRulesTriggerDataSourceModel struct {
	ID         types.Int64  `tfsdk:"id"`
	Name       types.String `tfsdk:"name"`
	Label      types.String `tfsdk:"label"`
	ObjectName types.String `tfsdk:"object_name"`
	Op         types.String `tfsdk:"op"`
	Cron       types.Bool   `tfsdk:"cron"`
	ClassName  types.String `tfsdk:"class_name"`
	IsActive   types.Bool   `tfsdk:"is_active"`
}

func NewCiviRulesTriggerDataSource() datasource.DataSource {
	return &CiviRulesTriggerDataSource{}
}

func (d *CiviRulesTriggerDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_civirules_trigger"
}

func (d *CiviRulesTriggerDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Fetches a CiviRules Trigger by ID or name. Use this to look up the trigger_id for civicrm_civirules_rule without hardcoding numeric IDs.",
		Attributes: map[string]schema.Attribute{
			"id": schema.Int64Attribute{
				Description: "Unique ID of the trigger. Specify either id or name.",
				Optional:    true,
				Computed:    true,
			},
			"name": schema.StringAttribute{
				Description: "Machine name of the trigger (e.g. 'activity_is_changed'). Specify either id or name.",
				Optional:    true,
				Computed:    true,
			},
			"label":       schema.StringAttribute{Description: "Display label shown in the CiviRules UI.", Computed: true},
			"object_name": schema.StringAttribute{Description: "CiviCRM entity this trigger fires on.", Computed: true},
			"op":          schema.StringAttribute{Description: "Operation: create, edit, or delete.", Computed: true},
			"cron":        schema.BoolAttribute{Description: "Whether this is a cron trigger.", Computed: true},
			"class_name":  schema.StringAttribute{Description: "PHP class implementing the trigger.", Computed: true},
			"is_active":   schema.BoolAttribute{Description: "Whether active.", Computed: true},
		},
	}
}

func (d *CiviRulesTriggerDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *CiviRulesTriggerDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config CiviRulesTriggerDataSourceModel
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

	tflog.Debug(ctx, "Reading CiviRulesTrigger data source", map[string]any{"where": where})

	results, err := d.client.Get("CiviRulesTrigger", where, nil)
	if err != nil {
		resp.Diagnostics.AddError("Error reading CiviRulesTrigger", err.Error())
		return
	}
	if len(results) == 0 {
		resp.Diagnostics.AddError("CiviRulesTrigger not found", "No trigger found matching the specified criteria.")
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
	if v, ok := GetString(result, "object_name"); ok && v != "" {
		config.ObjectName = types.StringValue(v)
	} else {
		config.ObjectName = types.StringNull()
	}
	if v, ok := GetString(result, "op"); ok && v != "" {
		config.Op = types.StringValue(v)
	} else {
		config.Op = types.StringNull()
	}
	if v, ok := GetBool(result, "cron"); ok {
		config.Cron = types.BoolValue(v)
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
