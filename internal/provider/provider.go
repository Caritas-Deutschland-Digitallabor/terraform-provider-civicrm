package provider

import (
	"context"
	"os"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var _ provider.Provider = &CiviCRMProvider{}

type CiviCRMProvider struct {
	version string
}

type CiviCRMProviderModel struct {
	URL      types.String `tfsdk:"url"`
	APIKey   types.String `tfsdk:"api_key"`
	Insecure types.Bool   `tfsdk:"insecure"`
}

func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &CiviCRMProvider{
			version: version,
		}
	}
}

func (p *CiviCRMProvider) Metadata(ctx context.Context, req provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "civicrm"
	resp.Version = p.version
}

func (p *CiviCRMProvider) Schema(ctx context.Context, req provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Terraform provider for managing CiviCRM access control resources via API v4.",
		Attributes: map[string]schema.Attribute{
			"url": schema.StringAttribute{
				Description: "The base URL of the CiviCRM instance (e.g., https://example.org/civicrm). " +
					"Can also be set via the CIVICRM_URL environment variable.",
				Optional: true,
			},
			"api_key": schema.StringAttribute{
				Description: "The API key for authenticating with CiviCRM. " +
					"Can also be set via the CIVICRM_API_KEY environment variable.",
				Optional:  true,
				Sensitive: true,
			},
			"insecure": schema.BoolAttribute{
				Description: "Skip TLS certificate verification. Only use for development. Default: false.",
				Optional:    true,
			},
		},
	}
}

func (p *CiviCRMProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	tflog.Info(ctx, "Configuring CiviCRM provider")

	var config CiviCRMProviderModel
	diags := req.Config.Get(ctx, &config)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Check for unknown values
	if config.URL.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root("url"),
			"Unknown CiviCRM URL",
			"The provider cannot create the CiviCRM API client as there is an unknown configuration value for the URL. "+
				"Either set the value statically in the configuration, or use the CIVICRM_URL environment variable.",
		)
	}

	if config.APIKey.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root("api_key"),
			"Unknown CiviCRM API Key",
			"The provider cannot create the CiviCRM API client as there is an unknown configuration value for the API key. "+
				"Either set the value statically in the configuration, or use the CIVICRM_API_KEY environment variable.",
		)
	}

	if resp.Diagnostics.HasError() {
		return
	}

	// Get values from environment variables if not set in config
	url := os.Getenv("CIVICRM_URL")
	apiKey := os.Getenv("CIVICRM_API_KEY")

	if !config.URL.IsNull() {
		url = config.URL.ValueString()
	}

	if !config.APIKey.IsNull() {
		apiKey = config.APIKey.ValueString()
	}

	// Validate required values
	if url == "" {
		resp.Diagnostics.AddAttributeError(
			path.Root("url"),
			"Missing CiviCRM URL",
			"The provider cannot create the CiviCRM API client as there is no URL configured. "+
				"Either set the url attribute in the provider configuration, or use the CIVICRM_URL environment variable.",
		)
	}

	if apiKey == "" {
		resp.Diagnostics.AddAttributeError(
			path.Root("api_key"),
			"Missing CiviCRM API Key",
			"The provider cannot create the CiviCRM API client as there is no API key configured. "+
				"Either set the api_key attribute in the provider configuration, or use the CIVICRM_API_KEY environment variable.",
		)
	}

	if resp.Diagnostics.HasError() {
		return
	}

	// Get insecure flag
	insecure := false
	if !config.Insecure.IsNull() {
		insecure = config.Insecure.ValueBool()
	}

	tflog.Debug(ctx, "Creating CiviCRM API client", map[string]any{
		"url":      url,
		"insecure": insecure,
	})

	// Create the API client
	client, err := NewClient(url, apiKey, insecure)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to create CiviCRM API client",
			"An unexpected error occurred when creating the CiviCRM API client. "+
				"Error: "+err.Error(),
		)
		return
	}

	// Make the client available to resources and data sources
	resp.DataSourceData = client
	resp.ResourceData = client

	tflog.Info(ctx, "Configured CiviCRM provider", map[string]any{
		"url": url,
	})
}

func (p *CiviCRMProvider) Resources(ctx context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewGroupResource,
		NewACLRoleResource,
		NewACLResource,
		NewACLEntityRoleResource,
		NewMailSettingsResource,
		NewSiteEmailAddressResource,
		NewCustomGroupResource,
		NewCustomFieldResource,
		NewTagResource,
		NewContactTypeResource,
		NewRelationshipTypeResource,
		NewCaseTypeResource,
	}
}

func (p *CiviCRMProvider) DataSources(ctx context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		NewGroupDataSource,
		NewACLRoleDataSource,
		NewACLDataSource,
		NewACLEntityRoleDataSource,
	}
}
