package provider

import (
	"context"
	"fmt"
	"github.com/MadJlzz/terraform-provider-oneprovider/pkg/oneprovider"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Ensure provider defined types fully satisfy framework interfaces.
var (
	_ datasource.DataSource              = &listLocationsDataSource{}
	_ datasource.DataSourceWithConfigure = &listLocationsDataSource{}
)

func NewLocationsDataSource() datasource.DataSource {
	return &listLocationsDataSource{}
}

// listLocationsDataSource defines the data source implementation.
type listLocationsDataSource struct {
	svc oneprovider.API
}

// listLocationsDataSourceModel describes the data source data model.
type listLocationsDataSourceModel struct {
	ID types.String `tfsdk:"id"`
}

func (d *listLocationsDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_locations"
}

func (d *listLocationsDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Fetches list of available locations.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Placeholder identifier attribute.",
				Computed:    true,
			},
		},
	}
}

func (d *listLocationsDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}
	svc, ok := req.ProviderData.(oneprovider.API)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *http.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}
	d.svc = svc
}

func (d *listLocationsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state listLocationsDataSourceModel
	tflog.Info(ctx, "retrieving oneprovider locations")

	locations, err := d.svc.ListLocations(ctx)
	if err != nil {
		resp.Diagnostics.AddError(
			"unable to list vm templates",
			err.Error(),
		)
	}

	tflog.Info(ctx, "location data", map[string]interface{}{
		"locations": locations,
	})

	// This is required by terraform for running acceptance tests.
	state.ID = types.StringValue("placeholder")

	// Save data into Terraform state
	//resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
