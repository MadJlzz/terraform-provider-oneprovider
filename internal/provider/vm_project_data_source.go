package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ datasource.DataSourceWithConfigure = &vmProjectDataSource{}

type vmProjectDataSource struct {
	datasourceServiceInjector
}

type vmProjectDataSourceModel struct {
	ID   types.String `tfsdk:"id"`
	Name types.String `tfsdk:"name"`
}

func NewVmProjectDataSource() datasource.DataSource {
	return &vmProjectDataSource{}
}

func (ds *vmProjectDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_vm_project"
}

func (ds *vmProjectDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Fetches a VM project by its name.",
		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{
				Description: "Exact project name. If multiple projects share this name, the first match returned by the API is used.",
				Required:    true,
			},
			"id": schema.StringAttribute{
				Description: "Project UUID to use in the project attribute of a VM instance.",
				Computed:    true,
			},
		},
	}
}

func (ds *vmProjectDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data vmProjectDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	name := data.Name.ValueString()
	if name == "" {
		resp.Diagnostics.AddAttributeError(path.Root("name"), "Missing attribute configuration", "Please provide a non-empty project name.")
		return
	}

	project, err := ds.svc.Projects.GetByName(ctx, name)
	if err != nil {
		resp.Diagnostics.AddError("Unable to read VM project", err.Error())
		return
	}
	if project.Uuid == "" {
		resp.Diagnostics.AddAttributeError(path.Root("id"), "Invalid project ID", "The API returned an empty project UUID; refusing to write an empty ID to state.")
		return
	}

	data.ID = types.StringValue(project.Uuid)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
