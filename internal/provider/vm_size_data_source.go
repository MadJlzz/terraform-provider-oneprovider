package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ datasource.DataSourceWithConfigure = &vmSizeDataSource{}
)

type vmSizeDataSource struct {
	datasourceServiceInjector
}

type vmSizeDataSourceModel struct {
	// Input attributes (Optional/Required for filtering)
	Name types.String `tfsdk:"name"`

	// Output attributes (Computed)
	ID    types.String `tfsdk:"id"`
	Type  types.String `tfsdk:"type"`
	Cores types.String `tfsdk:"cores"`
	RAM   types.String `tfsdk:"ram"`
	Disk  types.String `tfsdk:"disk"`
}

func NewVmSizeDataSource() datasource.DataSource {
	return &vmSizeDataSource{}
}

func (ds *vmSizeDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_vm_size"
}

func (ds *vmSizeDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description:         "Retrieve VM size information",
		MarkdownDescription: "Retrieve VM size information",
		Attributes: map[string]schema.Attribute{
			// Input attributes (Optional/Required for filtering)
			"name": schema.StringAttribute{
				Description: "Filter by name",
				Required:    true,
			},
			// Output attributes (Computed)
			"id": schema.StringAttribute{
				Description: "Size ID",
				Computed:    true,
			},
			"type": schema.StringAttribute{
				Description: "Type definition",
				Computed:    true,
			},
			"cores": schema.StringAttribute{
				Description: "Number of CPU core available on the VM",
				Computed:    true,
			},
			"ram": schema.StringAttribute{
				Description: "RAM available in MB",
				Computed:    true,
			},
			"disk": schema.StringAttribute{
				Description: "Disk storage size in GB",
				Computed:    true,
			},
		},
	}
}

func (ds *vmSizeDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data *vmSizeDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	name := data.Name.ValueString()
	if name == "" {
		resp.Diagnostics.AddAttributeError(
			path.Root("name"),
			"Missing attribute configuration",
			"Name is required to filter api results. Please provide a valid size name.")
		return
	}

	s, err := ds.svc.VM.GetSizeByName(ctx, name)
	if err != nil {
		resp.Diagnostics.Append()
		resp.Diagnostics.AddError(
			"Unable to refresh datasource",
			"An unexpected error occurred while creating the datasource read request."+
				"Please report this issue to the provider developers.\n\n"+
				err.Error(),
		)
		return
	}

	data.ID = types.StringValue(s.Id)
	data.Type = types.StringValue(s.Type)
	data.Cores = types.StringValue(s.Cores)
	data.RAM = types.StringValue(s.RAM)
	data.Disk = types.StringValue(s.Disk)
	
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
