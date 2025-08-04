package provider

import (
	"context"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"strconv"
)

var (
	_ datasource.DataSource = &vmTemplateDataSource{}
)

type vmTemplateDataSource struct {
	baseDatasource
}

type vmTemplateDataSourceModel struct {
	ID   types.String `tfsdk:"id"`
	Name types.String `tfsdk:"name"`
	Size types.String `tfsdk:"size"`
}

func NewVmTemplateDataSource() datasource.DataSource {
	return &vmTemplateDataSource{}
}

func (ds *vmTemplateDataSource) Metadata(ctx context.Context, request datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = request.ProviderTypeName + "_vm_template"
}

func (ds *vmTemplateDataSource) Schema(ctx context.Context, request datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description:         "Fetches a VM template by its name.",
		MarkdownDescription: "Fetches a VM template by its name.",
		Attributes: map[string]schema.Attribute{
			// Inputs
			"name": schema.StringAttribute{
				Description: "Name of the template.",
				Required:    true,
			},
			// Outputs
			"id": schema.StringAttribute{
				Description: "Placeholder identifier attribute.",
				Computed:    true,
			},
			"size": schema.StringAttribute{
				Description: "Size of the template in ???",
				Computed:    true,
			},
		},
	}
}

func (ds *vmTemplateDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data *vmTemplateDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	name := data.Name.ValueString()
	if name == "" {
		resp.Diagnostics.AddError(
			"name is required to filter api results",
			"name is required to filter api results",
		)
		return
	}

	tpl, err := ds.svc.VM.GetTemplateByName(ctx, name)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to get template by name",
			err.Error(),
		)
		return
	}

	data.ID = types.StringValue(strconv.Itoa(tpl.Id))
	data.Size = types.StringValue(tpl.Size)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
