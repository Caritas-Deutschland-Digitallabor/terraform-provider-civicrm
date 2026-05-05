package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var _ datasource.DataSource = &CaseStatusDataSource{}
var _ datasource.DataSourceWithConfigure = &CaseStatusDataSource{}

type CaseStatusDataSource struct {
	client *Client
}

type CaseStatusDataSourceModel struct {
	ID         types.Int64  `tfsdk:"id"`
	Name       types.String `tfsdk:"name"`
	Label      types.String `tfsdk:"label"`
	Grouping   types.String `tfsdk:"grouping"`
	IsActive   types.Bool   `tfsdk:"is_active"`
	IsReserved types.Bool   `tfsdk:"is_reserved"`
	Weight     types.Int64  `tfsdk:"weight"`
	Value      types.String `tfsdk:"value"`
}

func NewCaseStatusDataSource() datasource.DataSource {
	return &CaseStatusDataSource{}
}

func (d *CaseStatusDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_case_status"
}

func (d *CaseStatusDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Fetches a CiviCRM Case Status (OptionValue in case_status group) by ID or name. Useful to look up the `value` field for use in CiviRules action_params.",
		Attributes: map[string]schema.Attribute{
			"id": schema.Int64Attribute{
				Description: "Unique ID of the OptionValue. Specify either id or name.",
				Optional:    true,
				Computed:    true,
			},
			"name": schema.StringAttribute{
				Description: "Machine name of the status. Specify either id or name.",
				Optional:    true,
				Computed:    true,
			},
			"label":       schema.StringAttribute{Description: "Display label.", Computed: true},
			"grouping":    schema.StringAttribute{Description: "Category: 'Opened' or 'Closed'.", Computed: true},
			"is_active":   schema.BoolAttribute{Description: "Whether active.", Computed: true},
			"is_reserved": schema.BoolAttribute{Description: "Whether reserved by system.", Computed: true},
			"weight":      schema.Int64Attribute{Description: "Sort weight.", Computed: true},
			"value": schema.StringAttribute{
				Description: "Internal value used by CiviCRM. Reference this in CiviRules action_params for status changes.",
				Computed:    true,
			},
		},
	}
}

func (d *CaseStatusDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *CaseStatusDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config CaseStatusDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if config.ID.IsNull() && config.Name.IsNull() {
		resp.Diagnostics.AddError("Missing Filter", "At least one of 'id' or 'name' must be specified.")
		return
	}

	where := [][]any{
		{"option_group_id:name", "=", "case_status"},
	}
	if !config.ID.IsNull() {
		where = append(where, []any{"id", "=", config.ID.ValueInt64()})
	}
	if !config.Name.IsNull() {
		where = append(where, []any{"name", "=", config.Name.ValueString()})
	}

	tflog.Debug(ctx, "Reading CaseStatus data source", map[string]any{"where": where})

	results, err := d.client.Get("OptionValue", where, nil)
	if err != nil {
		resp.Diagnostics.AddError("Error reading CaseStatus", err.Error())
		return
	}
	if len(results) == 0 {
		resp.Diagnostics.AddError("CaseStatus not found", "No case status found matching the specified criteria.")
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
	if v, ok := GetString(result, "grouping"); ok && v != "" {
		config.Grouping = types.StringValue(v)
	} else {
		config.Grouping = types.StringNull()
	}
	if v, ok := GetBool(result, "is_active"); ok {
		config.IsActive = types.BoolValue(v)
	}
	if v, ok := GetBool(result, "is_reserved"); ok {
		config.IsReserved = types.BoolValue(v)
	}
	if v, ok := GetInt64(result, "weight"); ok {
		config.Weight = types.Int64Value(v)
	}
	if v, ok := GetString(result, "value"); ok && v != "" {
		config.Value = types.StringValue(v)
	} else {
		config.Value = types.StringNull()
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, config)...)
}
