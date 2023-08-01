package provider

import (
	"context"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"net/http"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
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
	Endpoint types.String `tfsdk:"endpoint"`
}

func (p *OneProvider) Metadata(ctx context.Context, req provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "oneprovider"
	resp.Version = p.version
}

func (p *OneProvider) Schema(ctx context.Context, req provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"api_key": schema.StringAttribute{
				MarkdownDescription: "API key required by OneProvider to run authenticated requests",
				Sensitive:           true,
				Optional:            true,
			},
		},
	}
}

func (p *OneProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var data OneProviderModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Info(ctx, "hello there")

	// Configuration values are now available.
	// if data.Endpoint.IsNull() { /* ... */ }

	// Example client configuration for data sources and resources
	client := http.DefaultClient
	resp.DataSourceData = client
	resp.ResourceData = client
}

func (p *OneProvider) Resources(ctx context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		//NewExampleResource,
	}
}

func (p *OneProvider) DataSources(ctx context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		//NewExampleDataSource,
	}
}

func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &OneProvider{
			version: version,
		}
	}
}
