package provider

import (
	"context"
	"fmt"
	"math/big"

	"github.com/MadJlzz/terraform-provider-oneprovider/pkg/oneprovider/vm"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ datasource.DataSourceWithConfigure = &vmLocationDataSource{}
)

type vmLocationDataSource struct {
	datasourceServiceInjector
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

func availableSizesToTerraform(_ context.Context, sizes []vm.StringOrNumber) (types.List, diag.Diagnostics) {
	if sizes == nil {
		return types.ListNull(types.NumberType), nil
	}

	elements := make([]attr.Value, len(sizes))
	var diags diag.Diagnostics
	for index, size := range sizes {
		value := size.String()
		if value == "" {
			diags.AddAttributeError(
				path.Root("available_sizes").AtListIndex(index),
				"Invalid available size ID",
				"available_sizes elements must be non-empty decimal integer IDs.",
			)
			continue
		}
		validDigits := true
		for _, digit := range value {
			if digit < '0' || digit > '9' {
				validDigits = false
				break
			}
		}
		if !validDigits {
			diags.AddAttributeError(
				path.Root("available_sizes").AtListIndex(index),
				"Invalid available size ID",
				fmt.Sprintf("available_sizes[%d] must be a non-negative decimal integer, got %q.", index, value),
			)
			continue
		}
		integer, ok := new(big.Int).SetString(value, 10)
		if !ok {
			diags.AddAttributeError(
				path.Root("available_sizes").AtListIndex(index),
				"Invalid available size ID",
				fmt.Sprintf("available_sizes[%d] must be a non-negative decimal integer, got %q.", index, value),
			)
			continue
		}
		elements[index] = types.NumberValue(new(big.Float).SetPrec(uint(integer.BitLen() + 1)).SetInt(integer))
	}
	if diags.HasError() {
		return types.ListNull(types.NumberType), diags
	}
	return types.ListValue(types.NumberType, elements)
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

	if l.Id.String() == "" {
		resp.Diagnostics.AddAttributeError(
			path.Root("id"),
			"Invalid location ID",
			"The API returned an empty location ID; refusing to write an empty ID to state.",
		)
		return
	}

	data.ID = types.StringValue(l.Id.String())
	data.Region = types.StringValue(l.Region)
	data.Country = types.StringValue(l.Country)
	var availableTypesDiags diag.Diagnostics
	data.AvailableTypes, availableTypesDiags = types.ListValueFrom(ctx, types.StringType, l.AvailableTypes)
	resp.Diagnostics.Append(availableTypesDiags...)
	if resp.Diagnostics.HasError() {
		return
	}
	var availableSizesDiags diag.Diagnostics
	data.AvailableSizes, availableSizesDiags = availableSizesToTerraform(ctx, l.AvailableSizes)
	resp.Diagnostics.Append(availableSizesDiags...)
	if resp.Diagnostics.HasError() {
		return
	}
	data.Ipv4 = types.StringValue(l.AvailableIPs.IPv4)
	data.Ipv6 = types.StringValue(l.AvailableIPs.IPv6)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
