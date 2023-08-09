package provider

import (
	"context"
	"github.com/MadJlzz/terraform-provider-oneprovider/internal/datasources"
	"github.com/MadJlzz/terraform-provider-oneprovider/internal/oneprovider"
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
	DefaultHostname = "api.oneprovider.com"
	ApiKeyEnvVar    = "ONEPROVIDER_API_KEY"
	ClientKeyEnvVar = "ONEPROVIDER_CLIENT_KEY"
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
	Host      types.String `tfsdk:"host"`
	ClientKey types.String `tfsdk:"client_key"`
	ApiKey    types.String `tfsdk:"api_key"`
}

func (p *OneProvider) Metadata(ctx context.Context, req provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "oneprovider"
	resp.Version = p.version
}

func (p *OneProvider) Schema(ctx context.Context, req provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"host": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "The host to which requests will be sent to. Defaults to api.oneprovider.com",
			},
			"api_key": schema.StringAttribute{
				Optional:            true,
				Sensitive:           true,
				MarkdownDescription: "API key required by OneProvider to run authenticated requests",
			},
			"client_key": schema.StringAttribute{
				Optional:            true,
				Sensitive:           true,
				MarkdownDescription: "Client key required by OneProvider to run authenticated requests",
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

	host := DefaultHostname
	if !providerConfiguration.Host.IsNull() {
		host = providerConfiguration.Host.ValueString()
	}

	apiKey := os.Getenv(ApiKeyEnvVar)
	if !providerConfiguration.ApiKey.IsNull() {
		apiKey = providerConfiguration.ApiKey.ValueString()
	}

	clientKey := os.Getenv(ClientKeyEnvVar)
	if !providerConfiguration.ClientKey.IsNull() {
		clientKey = providerConfiguration.ClientKey.ValueString()
	}

	if host == "" {
		resp.Diagnostics.AddAttributeError(
			path.Root("host"),
			"Missing OneProvider API host",
			"Missing or empty value for host property. Cannot create OneProvider API client. Set the host value in the configuration.",
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

	svc, err := oneprovider.NewService(host, clientKey, apiKey)
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
		datasources.NewVmTemplatesDataSource,
	}
}

func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &OneProvider{
			version: version,
		}
	}
}
