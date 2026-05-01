package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var _ datasource.DataSource = &ACLEntityRoleDataSource{}
var _ datasource.DataSourceWithConfigure = &ACLEntityRoleDataSource{}

type ACLEntityRoleDataSource struct {
	client *Client
}

type ACLEntityRoleDataSourceModel struct {
	ID          types.Int64  `tfsdk:"id"`
	ACLRoleID   types.Int64  `tfsdk:"acl_role_id"`
	EntityTable types.String `tfsdk:"entity_table"`
	EntityID    types.Int64  `tfsdk:"entity_id"`
	IsActive    types.Bool   `tfsdk:"is_active"`
}

func NewACLEntityRoleDataSource() datasource.DataSource {
	return &ACLEntityRoleDataSource{}
}

func (d *ACLEntityRoleDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_acl_entity_role"
}

func (d *ACLEntityRoleDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Fetches a CiviCRM ACL Entity Role assignment by ID or by role and entity combination.",
		Attributes: map[string]schema.Attribute{
			"id": schema.Int64Attribute{
				Description: "The unique identifier of the ACL entity role assignment.",
				Optional:    true,
				Computed:    true,
			},
			"acl_role_id": schema.Int64Attribute{
				Description: "The value field of the ACL role. Use with entity_id to look up by combination.",
				Optional:    true,
				Computed:    true,
			},
			"entity_table": schema.StringAttribute{
				Description: "The table containing the entity.",
				Optional:    true,
				Computed:    true,
			},
			"entity_id": schema.Int64Attribute{
				Description: "The ID of the entity. Use with acl_role_id to look up by combination.",
				Optional:    true,
				Computed:    true,
			},
			"is_active": schema.BoolAttribute{
				Description: "Whether this role assignment is active.",
				Computed:    true,
			},
		},
	}
}

func (d *ACLEntityRoleDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}

	d.client = client
}

func (d *ACLEntityRoleDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config ACLEntityRoleDataSourceModel
	diags := req.Config.Get(ctx, &config)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Build where clause based on provided filters
	var where [][]any
	if !config.ID.IsNull() {
		where = append(where, []any{"id", "=", config.ID.ValueInt64()})
	}
	if !config.ACLRoleID.IsNull() {
		where = append(where, []any{"acl_role_id", "=", config.ACLRoleID.ValueInt64()})
	}
	if !config.EntityTable.IsNull() {
		where = append(where, []any{"entity_table", "=", config.EntityTable.ValueString()})
	}
	if !config.EntityID.IsNull() {
		where = append(where, []any{"entity_id", "=", config.EntityID.ValueInt64()})
	}

	// Require at least id or the combination of acl_role_id and entity_id
	hasID := !config.ID.IsNull()
	hasCombination := !config.ACLRoleID.IsNull() && !config.EntityID.IsNull()

	if !hasID && !hasCombination {
		resp.Diagnostics.AddError(
			"Missing Filter",
			"Either 'id' or both 'acl_role_id' and 'entity_id' must be specified.",
		)
		return
	}

	tflog.Debug(ctx, "Reading ACL entity role data source", map[string]any{
		"filters": where,
	})

	results, err := d.client.Get("ACLEntityRole", where, nil)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading ACL entity role",
			"Could not read ACL entity role: "+err.Error(),
		)
		return
	}

	if len(results) == 0 {
		resp.Diagnostics.AddError(
			"ACL entity role not found",
			"No ACL entity role found matching the specified criteria.",
		)
		return
	}

	result := results[0]

	// Update state
	if id, ok := GetInt64(result, "id"); ok {
		config.ID = types.Int64Value(id)
	}

	if aclRoleID, ok := GetInt64(result, "acl_role_id"); ok {
		config.ACLRoleID = types.Int64Value(aclRoleID)
	}

	if entityTable, ok := GetString(result, "entity_table"); ok {
		config.EntityTable = types.StringValue(entityTable)
	}

	if entityID, ok := GetInt64(result, "entity_id"); ok {
		config.EntityID = types.Int64Value(entityID)
	}

	if active, ok := GetBool(result, "is_active"); ok {
		config.IsActive = types.BoolValue(active)
	}

	diags = resp.State.Set(ctx, config)
	resp.Diagnostics.Append(diags...)
}
