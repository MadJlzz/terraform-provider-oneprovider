package provider

import (
	"context"
	"github.com/MadJlzz/terraform-provider-oneprovider/internal/datasources"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"net/http"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
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
			"client_key": schema.StringAttribute{
				Optional:            true,
				Sensitive:           true,
				MarkdownDescription: "Client key required by OneProvider to run authenticated requests",
			},
			"api_key": schema.StringAttribute{
				Optional:            true,
				Sensitive:           true,
				MarkdownDescription: "API key required by OneProvider to run authenticated requests",
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

	// Configuration values are now available.
	//if providerConfiguration.Host.IsUnknown() {
	//	resp.Diagnostics.AddAttributeError(
	//		path.Root("host"),
	//		"test",
	//		"etst",
	//	)
	//}

	//if resp.Diagnostics.HasError() {
	//	return
	//}

	// if data.Endpoint.IsNull() { /* ... */ }

	// Example client configuration for data sources and resources
	// TODO: create a OneProvider HTTP client
	client := http.DefaultClient
	resp.DataSourceData = client
	resp.ResourceData = client
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
