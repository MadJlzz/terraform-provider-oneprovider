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
	_ datasource.DataSource              = &listVMTemplatesDataSource{}
	_ datasource.DataSourceWithConfigure = &listVMTemplatesDataSource{}
)

func NewVmTemplatesDataSource() datasource.DataSource {
	return &listVMTemplatesDataSource{}
}

// listVMTemplatesDataSource defines the data source implementation.
type listVMTemplatesDataSource struct {
	svc oneprovider.API
}

// listVMTemplatesDataSourceModel describes the data source data model.
type listVMTemplatesDataSourceModel struct {
	ID          types.String      `tfsdk:"id"`
	VMTemplates []vmTemplateModel `tfsdk:"templates"`
}

type vmTemplateModel struct {
	ID      types.Int64            `tfsdk:"id"`
	Name    types.String           `tfsdk:"name"`
	Size    types.String           `tfsdk:"size"`
	Display vmTemplateDisplayModel `tfsdk:"display"`
}

type vmTemplateDisplayModel struct {
	Name        types.String `tfsdk:"name"`
	Display     types.String `tfsdk:"display"`
	Description types.String `tfsdk:"description"`
	Oca         types.Int64  `tfsdk:"oca"`
}

func (d *listVMTemplatesDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_vm_templates"
}

func (d *listVMTemplatesDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Fetches list of VM templates available.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Placeholder identifier attribute.",
				Computed:    true,
			},
			"templates": schema.ListNestedAttribute{
				Description: "List of VM templates.",
				Computed:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.Int64Attribute{
							Description: "Numeric identifier of the template.",
							Computed:    true,
						},
						"name": schema.StringAttribute{
							Description: "Name of the template.",
							Computed:    true,
						},
						"size": schema.StringAttribute{
							Description: "Size of the template in ???",
							Computed:    true,
						},
						"display": schema.SingleNestedAttribute{
							Computed: true,
							Attributes: map[string]schema.Attribute{
								"name": schema.StringAttribute{
									Description: "Display name of the template.",
									Computed:    true,
								},
								"display": schema.StringAttribute{
									Description: "Display of the template.",
									Computed:    true,
								},
								"description": schema.StringAttribute{
									Description: "Description of the display template.",
									Computed:    true,
								},
								"oca": schema.Int64Attribute{
									Description: "???",
									Computed:    true,
								},
							},
						},
					},
				},
			},
		},
	}
}

func (d *listVMTemplatesDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *listVMTemplatesDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state listVMTemplatesDataSourceModel
	tflog.Info(ctx, "retrieving oneprovider vm templates")

	templates, err := d.svc.ListTemplates(ctx)
	if err != nil {
		resp.Diagnostics.AddError(
			"unable to list vm templates",
			err.Error(),
		)
	}

	for _, template := range templates.Templates {
		templateState := vmTemplateModel{
			ID:   types.Int64Value(int64(template.Id)),
			Name: types.StringValue(template.Name),
			Size: types.StringValue(template.Size),
			Display: vmTemplateDisplayModel{
				Name:        types.StringValue(template.Display.Name),
				Display:     types.StringValue(template.Display.Display),
				Description: types.StringValue(template.Display.Description),
				Oca:         types.Int64Value(int64(template.Display.Oca)),
			},
		}
		state.VMTemplates = append(state.VMTemplates, templateState)
	}

	// This is required by terraform for running acceptance tests.
	state.ID = types.StringValue("placeholder")

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
