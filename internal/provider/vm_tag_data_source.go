package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ datasource.DataSourceWithConfigure = &vmTagDataSource{}

type vmTagDataSource struct {
	datasourceServiceInjector
}

type vmTagDataSourceModel struct {
	Name    types.String `tfsdk:"name"`
	Project types.String `tfsdk:"project"`
	ID      types.String `tfsdk:"id"`
}

func NewVmTagDataSource() datasource.DataSource { return &vmTagDataSource{} }

func (ds *vmTagDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_vm_tag"
}

func (ds *vmTagDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Fetches an existing tag by label within a VM project.",
		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{
				Description: "Exact tag label. If labels are duplicated, the first match is returned.",
				Required:    true,
				Validators:  []validator.String{stringvalidator.LengthAtLeast(1)},
			},
			"project": schema.StringAttribute{
				Description: "Project UUID. If omitted, searches the account's default project.",
				Optional:    true,
				Validators:  []validator.String{stringvalidator.LengthAtLeast(1)},
			},
			"id": schema.StringAttribute{
				Description: "Tag ID to use in the tags attribute of a VM instance.",
				Computed:    true,
			},
		},
	}
}

func (ds *vmTagDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data vmTagDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	tag, err := ds.svc.Tags.GetByName(ctx, data.Project.ValueString(), data.Name.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Unable to read VM tag", err.Error())
		return
	}
	data.ID = types.StringValue(tag.Id.String())
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
