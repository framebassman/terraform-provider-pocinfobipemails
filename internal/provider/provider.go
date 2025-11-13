// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"
	"os"

	"github.com/framebassman/infobip-api-go-client/v3/pkg/infobip"
	"github.com/framebassman/infobip-api-go-client/v3/pkg/infobip/api"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ provider.Provider = &pocinfobipemailsProvider{}
)

// New is a helper function to simplify provider server and testing implementation.
func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &pocinfobipemailsProvider{
			version: version,
		}
	}
}

// pocinfobipemailsProvider is the provider implementation.
type pocinfobipemailsProvider struct {
	// version is set to the provider version on release, "dev" when the
	// provider is built and ran locally, and "test" when running acceptance
	// testing.
	version string
}

// Metadata returns the provider type name.
func (p *pocinfobipemailsProvider) Metadata(_ context.Context, _ provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "pocinfobipemails"
	resp.Version = p.version
}

// pocInfobipEmailsProviderModel maps provider schema data to a Go type.
type pocInfobipEmailsProviderModel struct {
	BaseUrl types.String `tfsdk:"base_url"`
	ApiKey  types.String `tfsdk:"api_key"`
}

type providerClient struct {
	client *api.APIClient
	apiKey string
}

// Schema defines the provider-level schema for configuration data.
func (p *pocinfobipemailsProvider) Schema(_ context.Context, _ provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"base_url": schema.StringAttribute{
				Optional: false,
				Required: true,
			},
			"api_key": schema.StringAttribute{
				Optional: false,
				Required: true,
			},
		},
	}
}

func (p *pocinfobipemailsProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	tflog.Info(ctx, "Configuring Infobip client")
	// Retrieve provider data from configuration
	var config pocInfobipEmailsProviderModel
	diags := req.Config.Get(ctx, &config)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// If practitioner provided a configuration value for any of the
	// attributes, it must be a known value.

	if config.BaseUrl.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root("base_url"),
			"Unknown Infobip API base url",
			"The provider cannot create the Infobip API client as there is an unknown configuration value for the Infobip API base url. "+
				"Either target apply the source of the value first, set the value statically in the configuration, or use the POCINFOBIPEMAILS_BASE_URL environment variable.",
		)
	}

	if config.ApiKey.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root("api_key"),
			"Unknown Infobip API key",
			"The provider cannot create the Infobip API client as there is an unknown configuration value for the Infobip API key. "+
				"Either target apply the source of the value first, set the value statically in the configuration, or use the POCINFOBIPEMAILS_API_KEY environment variable.",
		)
	}

	if resp.Diagnostics.HasError() {
		return
	}

	// Default values to environment variables, but override
	// with Terraform configuration value if set.

	base_url := os.Getenv("POCINFOBIPEMAILS_BASE_URL")
	api_key := os.Getenv("POCINFOBIPEMAILS_API_KEY")

	if !config.BaseUrl.IsNull() {
		base_url = config.BaseUrl.ValueString()
	}

	if !config.ApiKey.IsNull() {
		api_key = config.ApiKey.ValueString()
	}

	// If any of the expected configurations are missing, return
	// errors with provider-specific guidance.

	if base_url == "" {
		resp.Diagnostics.AddAttributeError(
			path.Root("base_url"),
			"Missing Infobip API base url",
			"The provider cannot create the Infobip API client as there is a missing or empty value for the Infobip API base url. "+
				"Set the host value in the configuration or use the POCINFOBIPEMAILS_BASE_URL environment variable. "+
				"If either is already set, ensure the value is not empty.",
		)
	}

	if api_key == "" {
		resp.Diagnostics.AddAttributeError(
			path.Root("api_key"),
			"Missing Infobip API key",
			"The provider cannot create the Infobip API client as there is a missing or empty value for the Infobip API key. "+
				"Set the api_key value in the configuration or use the POCINFOBIPEMAILS_API_KEY environment variable. "+
				"If either is already set, ensure the value is not empty.",
		)
	}

	if resp.Diagnostics.HasError() {
		return
	}

	ctx = tflog.SetField(ctx, "infobip_base_url", base_url)
	ctx = tflog.SetField(ctx, "infobip_api_key", api_key)
	ctx = tflog.MaskFieldValuesWithFieldKeys(ctx, "infobip_api_key")

	tflog.Debug(ctx, "Creating Infobip client")

	ctx = tflog.SetField(ctx, "infobip_base_url", base_url)
	ctx = tflog.SetField(ctx, "infobip_api_key", api_key)
	ctx = tflog.MaskFieldValuesWithFieldKeys(ctx, "infobip_api_key")

	configuration := infobip.NewConfiguration()
	configuration.Host = base_url

	infobipClient := api.NewAPIClient(configuration)

	auth := context.WithValue(
		context.Background(),
		infobip.ContextAPIKeys,
		map[string]infobip.APIKey{"APIKeyHeader": {Key: api_key, Prefix: "App"}},
	)

	apiResponse, httpResponse, err := infobipClient.
		EmailAPI.
		GetAllEmailTemplates(auth).
		Execute()

	// Check for errors
	if err != nil {
		resp.Diagnostics.AddError("Failed to fetch all email templates", err.Error()) // Fail the test with the error message
		return
	}

	// Output response details for debugging
	tflog.Info(ctx, "Response: "+fmt.Sprintf("%+v", apiResponse))
	tflog.Info(ctx, "HTTP Response Details: "+fmt.Sprintf("%+v", httpResponse))

	// Validate response
	if apiResponse == nil || apiResponse.Results == nil {
		resp.Diagnostics.AddError("Invalid response", "Expected messages, but got: "+fmt.Sprintf("%+v", apiResponse))
	}

	// Make the HashiCups client available during DataSource and Resource
	// type Configure methods.
	// Build provider payload containing both client and apiKey
	provData := &providerClient{
		client: infobipClient,
		apiKey: api_key,
	}
	resp.DataSourceData = provData
	resp.ResourceData = provData
	tflog.Info(ctx, "Configured Infobip client", map[string]any{"success": true})
}

// DataSources defines the data sources implemented in the provider.
func (p *pocinfobipemailsProvider) DataSources(_ context.Context) []func() datasource.DataSource {
	return nil
}

// Resources defines the resources implemented in the provider.
func (p *pocinfobipemailsProvider) Resources(_ context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewEmailTemplateResource,
	}
}
