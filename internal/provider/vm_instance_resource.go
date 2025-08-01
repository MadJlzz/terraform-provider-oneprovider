package provider

import (
	"context"
	"fmt"
	"github.com/MadJlzz/terraform-provider-oneprovider/pkg/api"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"strconv"
)

var (
	_ resource.ResourceWithConfigure = &vmInstanceResource{}
)

type vmInstanceResource struct {
	svc *api.OneProvider
}

type vmInstanceResourceModel struct {
	ID             types.String `tfsdk:"id"`
	LocationId     types.String `tfsdk:"location_id"`
	InstanceSizeId types.String `tfsdk:"instance_size_id"`
	TemplateId     types.String `tfsdk:"template_id"`
	Hostname       types.String `tfsdk:"hostname"`
}

func NewVmInstanceResource() resource.Resource {
	return &vmInstanceResource{}
}

func (r *vmInstanceResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_vm_instance"
}

func (r *vmInstanceResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description:         "Create a new VM instance.",
		MarkdownDescription: "Create a new VM instance.",
		Attributes: map[string]schema.Attribute{
			// Inputs
			"location_id": schema.StringAttribute{
				Description: "Location ID referencing where the VM instance will be created",
				Required:    true,
			},
			"instance_size_id": schema.StringAttribute{
				Description: "Instance size ID referencing the hardware specs of the VM instance",
				Required:    true,
			},
			"template_id": schema.StringAttribute{
				Description: "Template ID referencing the OS to use for that VM instance",
				Required:    true,
			},
			"hostname": schema.StringAttribute{
				Description: "Hostname of the VM instance",
				Required:    true,
			},
			// Outputs
			"id": schema.StringAttribute{
				Description: "ID of the VM instance. Generated by the provider.",
				Computed:    true,
			},
		},
	}
}

func (r *vmInstanceResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	// Always perform a nil check when handling ProviderData because Terraform
	// sets that data after it calls the ConfigureProvider RPC.
	if req.ProviderData == nil {
		return
	}
	tflog.Info(ctx, "configuring datasource dependencies")
	svc, ok := req.ProviderData.(*api.OneProvider)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected oneprovider.API, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}
	r.svc = svc
}

func (r *vmInstanceResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data *vmInstanceResourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	locationId, _ := strconv.Atoi(data.LocationId.ValueString())
	instanceSizeId, _ := strconv.Atoi(data.InstanceSizeId.ValueString())

	createRequest := &api.VMInstanceCreateRequest{
		LocationId:     locationId,
		InstanceSizeId: instanceSizeId,
		TemplateId:     data.TemplateId.ValueString(),
		Hostname:       data.Hostname.ValueString(),
	}
	vm, err := r.svc.CreateVMInstance(ctx, createRequest)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to create VM instance",
			err.Error(),
		)
		return
	}

	data.ID = types.StringValue(vm.Response.Id)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *vmInstanceResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	//TODO implement me
	panic("implement read")
}

func (r *vmInstanceResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	//TODO implement me
	panic("implement update")
}

func (r *vmInstanceResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data *vmInstanceResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	destroyRequest := &api.VMInstanceDestroyRequest{
		VMId:         data.ID.ValueString(),
		ConfirmClose: true,
	}

	_, err := r.svc.DestroyVMInstance(ctx, destroyRequest)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to destroy VM instance",
			err.Error(),
		)
		return
	}
}
