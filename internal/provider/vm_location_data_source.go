package provider

import (
	"context"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ datasource.DataSource              = &vmLocationDataSource{}
	_ datasource.DataSourceWithConfigure = &vmLocationDataSource{}
)

type vmLocationDataSource struct {
	baseDatasource
}

type vmLocationDataSourceModel struct {
	// Input attributes (Optional/Required for filtering)
	City types.String `tfsdk:"city"`

	// Output attributes (Computed)
	ID             types.String `tfsdk:"id"`
	Region         types.String `tfsdk:"region"`
	Country        types.String `tfsdk:"country"`
	AvailableTypes types.List   `tfsdk:"available_types"`
	AvailableSizes types.List   `tfsdk:"available_sizes"`
	Ipv4           types.String `tfsdk:"ipv4"`
	Ipv6           types.String `tfsdk:"ipv6"`
}

func NewVMLocationDataSource() datasource.DataSource {
	return &vmLocationDataSource{}
}

func (ds *vmLocationDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_vm_location"
}

func (ds *vmLocationDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description:         "Retrieve location information given its city",
		MarkdownDescription: "Retrieve location information given its city",
		Attributes: map[string]schema.Attribute{
			// Input attributes (Optional/Required for filtering)
			"city": schema.StringAttribute{
				Description: "Filter by city",
				Required:    true,
			},
			// Output attributes (Computed)
			"id": schema.StringAttribute{
				Description: "Location ID",
				Computed:    true,
			},
			"region": schema.StringAttribute{
				Description: "Location region",
				Computed:    true,
			},
			"country": schema.StringAttribute{
				Description: "Location country",
				Computed:    true,
			},
			"available_types": schema.ListAttribute{
				Description: "List of available VM types",
				Computed:    true,
				ElementType: types.StringType,
			},
			"available_sizes": schema.ListAttribute{
				Description: "List of available VM sizes",
				Computed:    true,
				ElementType: types.NumberType,
			},
			"ipv4": schema.StringAttribute{
				Description: "Location IPv4 address",
				Computed:    true,
			},
			"ipv6": schema.StringAttribute{
				Description: "Location IPv6 address",
				Computed:    true,
			},
		},
	}
}

func (ds *vmLocationDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data *vmLocationDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	city := data.City.ValueString()
	if city == "" {
		resp.Diagnostics.AddAttributeError(
			path.Root("city"),
			"Missing attribute configuration",
			"City is required to filter api results. Please provide a valid city name.",
		)
		return
	}

	l, err := ds.svc.VM.GetLocationByCity(ctx, city)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to refresh datasource",
			"An unexpected error occurred while creating the datasource read request."+
				"Please report this issue to the provider developers.\n\n"+
				err.Error(),
		)
		return
	}

	data.ID = types.StringValue(l.Id)
	data.Region = types.StringValue(l.Region)
	data.Country = types.StringValue(l.Country)
	data.AvailableTypes, _ = types.ListValueFrom(ctx, types.StringType, l.AvailableTypes)
	data.AvailableSizes, _ = types.ListValueFrom(ctx, types.NumberType, l.AvailableSizes)
	data.Ipv4 = types.StringValue(l.AvailableIPs.IPv4)
	data.Ipv6 = types.StringValue(l.AvailableIPs.IPv6)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
