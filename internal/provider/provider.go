package provider

import (
	"context"
	"github.com/MadJlzz/terraform-provider-oneprovider/pkg/oneprovider"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"os"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
)

const (
	ApiKeyEnvVar    = "ONEPROVIDER_API_KEY"
	ClientKeyEnvVar = "ONEPROVIDER_CLIENT_KEY"
	DefaultEndpoint = "https://api.oneprovider.com"
)

// Ensure OneProvider satisfies various provider interfaces.
var _ provider.Provider = &OneProvider{}

// OneProvider defines the provider implementation.
type OneProvider struct {
	// version is set to the provider version on release, "dev" when the
	// provider is built and ran locally, and "test" when running acceptance
	// testing.
	version string
}

// OneProviderModel describes the provider data model.
type OneProviderModel struct {
	ApiKey    types.String `tfsdk:"api_key"`
	ClientKey types.String `tfsdk:"client_key"`
	Endpoint  types.String `tfsdk:"endpoint"`
}

func (p *OneProvider) Metadata(ctx context.Context, req provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "oneprovider"
	resp.Version = p.version
}

func (p *OneProvider) Schema(ctx context.Context, req provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"api_key": schema.StringAttribute{
				Description: "Api Key for OneProvider API. May also be provided via ONEPROVIDER_API_KEY environment variable.",
				Optional:    true,
				Sensitive:   true,
			},
			"client_key": schema.StringAttribute{
				Description: "Client key for OneProvider API. May also be provided via ONEPROVIDER_CLIENT_KEY environment variable.",
				Optional:    true,
				Sensitive:   true,
			},
			"endpoint": schema.StringAttribute{
				Description: "URI for OneProvider API. Defaults to https://api.oneprovider.com",
				Optional:    true,
			},
		},
	}
}

func (p *OneProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	tflog.Info(ctx, "Configuring OneProvider client")

	var providerConfiguration OneProviderModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &providerConfiguration)...)
	if resp.Diagnostics.HasError() {
		return
	}

	endpoint := DefaultEndpoint
	if !providerConfiguration.Endpoint.IsNull() {
		endpoint = providerConfiguration.Endpoint.ValueString()
	}

	apiKey := os.Getenv(ApiKeyEnvVar)
	if !providerConfiguration.ApiKey.IsNull() {
		apiKey = providerConfiguration.ApiKey.ValueString()
	}

	clientKey := os.Getenv(ClientKeyEnvVar)
	if !providerConfiguration.ClientKey.IsNull() {
		clientKey = providerConfiguration.ClientKey.ValueString()
	}

	if endpoint == "" {
		resp.Diagnostics.AddAttributeError(
			path.Root("endpoint"),
			"Missing OneProvider API endpoint",
			"Missing or empty value for endpoint property. Cannot create OneProvider API client. Set the endpoint value in the configuration.",
		)
	}

	if clientKey == "" {
		resp.Diagnostics.AddAttributeError(
			path.Root("client_key"),
			"Missing OneProvider API client key",
			"Missing or empty value for client key property. Cannot create OneProvider API client. Set the client key value in the configuration or use the "+ClientKeyEnvVar+" environment variable.",
		)
	}

	if apiKey == "" {
		resp.Diagnostics.AddAttributeError(
			path.Root("api_key"),
			"Missing OneProvider API api key",
			"Missing or empty value for api key property. Cannot create OneProvider API client. Set the api key value in the configuration or use the "+ApiKeyEnvVar+" environment variable.",
		)
	}

	if resp.Diagnostics.HasError() {
		return
	}

	svc, err := oneprovider.NewService(endpoint, clientKey, apiKey)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to create OneProvider API client",
			"An unexpected error occurred when creating the OneProvider API client. "+
				"If the error is not clear, please contact the provider developers.\n\n"+
				"OneProvider client error: "+err.Error(),
		)
		return
	}
	resp.DataSourceData = svc
	resp.ResourceData = svc
}

func (p *OneProvider) Resources(ctx context.Context) []func() resource.Resource {
	return []func() resource.Resource{}
}

func (p *OneProvider) DataSources(ctx context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		NewVmTemplatesDataSource,
		NewLocationsDataSource,
	}
}

func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &OneProvider{
			version: version,
		}
	}
}
