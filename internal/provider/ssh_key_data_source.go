package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/MadJlzz/terraform-provider-oneprovider/pkg/oneprovider/ssh"
)

var (
	_ datasource.DataSourceWithConfigure = &sshKeyDataSource{}
)

type sshKeyDataSource struct {
	datasourceServiceInjector
}

type sshKeyDataSourceModel struct {
	// Input attributes (Optional/Required for filtering)
	Id   types.String `tfsdk:"id"`
	Name types.String `tfsdk:"name"`

	// Output attributes (Computed)
	PublicKey types.String `tfsdk:"public_key"`
}

func NewSSHKeyDataSource() datasource.DataSource {
	return &sshKeyDataSource{}
}

func (ds *sshKeyDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_ssh_key"
}

func (ds *sshKeyDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description:         "Retrieve SSH key by its ID or name.",
		MarkdownDescription: "Retrieve SSH key by its ID or name.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Filter by ID.",
				Optional:            true,
				Validators: []validator.String{
					stringvalidator.ExactlyOneOf(
						path.Expressions{path.MatchRoot("name")}...,
					),
				},
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Filter by name.",
				Optional:            true,
			},
			"public_key": schema.StringAttribute{
				MarkdownDescription: "SSH public key.",
				Computed:            true,
			},
		},
	}
}

func (ds *sshKeyDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data *sshKeyDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var info *ssh.SshKeyReadResponse
	var err error

	if !data.Id.IsNull() && !data.Id.IsUnknown() {
		info, err = ds.svc.SSH.GetByID(ctx, data.Id.ValueString())
		if err == nil {
			data.Name = types.StringValue(info.Name)
		}
	} else {
		info, err = ds.svc.SSH.GetByName(ctx, data.Name.ValueString())
		if err == nil {
			data.Id = types.StringValue(info.Uuid)
		}
	}

	if err != nil {
		resp.Diagnostics.AddError("", "")
		return
	}

	data.PublicKey = types.StringValue(info.Value)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
