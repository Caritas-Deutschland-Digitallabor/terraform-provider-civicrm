package provider

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var _ datasource.DataSource = &CaseTypeDataSource{}
var _ datasource.DataSourceWithConfigure = &CaseTypeDataSource{}

type CaseTypeDataSource struct {
	client *Client
}

type CaseTypeDataSourceModel struct {
	ID          types.Int64  `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	Title       types.String `tfsdk:"title"`
	Description types.String `tfsdk:"description"`
	IsActive    types.Bool   `tfsdk:"is_active"`
	IsReserved  types.Bool   `tfsdk:"is_reserved"`
	Weight      types.Int64  `tfsdk:"weight"`
	Definition  types.String `tfsdk:"definition"`
}

func NewCaseTypeDataSource() datasource.DataSource {
	return &CaseTypeDataSource{}
}

func (d *CaseTypeDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_case_type"
}

func (d *CaseTypeDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Fetches a CiviCRM Case Type by ID or name.",
		Attributes: map[string]schema.Attribute{
			"id": schema.Int64Attribute{
				Description: "Unique ID of the case type. Specify either id or name.",
				Optional:    true,
				Computed:    true,
			},
			"name": schema.StringAttribute{
				Description: "Machine name of the case type. Specify either id or name.",
				Optional:    true,
				Computed:    true,
			},
			"title":       schema.StringAttribute{Description: "Display title.", Computed: true},
			"description": schema.StringAttribute{Description: "Description.", Computed: true},
			"is_active":   schema.BoolAttribute{Description: "Whether the case type is active.", Computed: true},
			"is_reserved": schema.BoolAttribute{Description: "Whether reserved by system.", Computed: true},
			"weight":      schema.Int64Attribute{Description: "Sort weight.", Computed: true},
			"definition":  schema.StringAttribute{Description: "Case type definition as JSON string.", Computed: true},
		},
	}
}

func (d *CaseTypeDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *CaseTypeDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config CaseTypeDataSourceModel
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

	tflog.Debug(ctx, "Reading CaseType data source", map[string]any{"where": where})

	results, err := d.client.Get("CaseType", where, nil)
	if err != nil {
		resp.Diagnostics.AddError("Error reading CaseType", err.Error())
		return
	}
	if len(results) == 0 {
		resp.Diagnostics.AddError("CaseType not found", "No case type found matching the specified criteria.")
		return
	}

	result := results[0]

	if id, ok := GetInt64(result, "id"); ok {
		config.ID = types.Int64Value(id)
	}
	if v, ok := GetString(result, "name"); ok {
		config.Name = types.StringValue(v)
	}
	if v, ok := GetString(result, "title"); ok {
		config.Title = types.StringValue(v)
	}
	if v, ok := GetString(result, "description"); ok && v != "" {
		config.Description = types.StringValue(v)
	} else {
		config.Description = types.StringNull()
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
	if def, ok := result["definition"]; ok && def != nil {
		switch v := def.(type) {
		case string:
			if v != "" {
				config.Definition = types.StringValue(v)
			} else {
				config.Definition = types.StringNull()
			}
		default:
			if b, err := json.Marshal(v); err == nil {
				config.Definition = types.StringValue(string(b))
			} else {
				config.Definition = types.StringNull()
			}
		}
	} else {
		config.Definition = types.StringNull()
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, config)...)
}
