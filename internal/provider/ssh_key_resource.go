package provider

import (
	"context"
	"fmt"
	"github.com/MadJlzz/terraform-provider-oneprovider/pkg/oneprovider"
	"github.com/MadJlzz/terraform-provider-oneprovider/pkg/oneprovider/ssh"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"regexp"
	"time"
)

var (
	_ resource.ResourceWithConfigure = &sshKeyResource{}
)

type sshKeyResource struct {
	svc *oneprovider.Service
}

type sshKeyResourceModel struct {
	Id        types.String `tfsdk:"id"`
	Name      types.String `tfsdk:"name"`
	PublicKey types.String `tfsdk:"public_key"`
}

func NewSSHKeyResource() resource.Resource {
	return &sshKeyResource{}
}

// TODO: refactor this since it's duplicated with configure from vm_instance_resource.
func (r *sshKeyResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	// Always perform a nil check when handling ProviderData because Terraform
	// sets that data after it calls the ConfigureProvider RPC.
	if req.ProviderData == nil {
		return
	}
	tflog.Info(ctx, "configuring datasource dependencies")
	svc, ok := req.ProviderData.(*oneprovider.Service)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected service type",
			fmt.Sprintf("Expected oneprovider.Service, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}
	r.svc = svc
}

func (r *sshKeyResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_ssh_key"
}

func (r *sshKeyResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description:         "Create a new public SSH key",
		MarkdownDescription: "Create a new public SSH key",
		Attributes: map[string]schema.Attribute{
			// Inputs
			"name": schema.StringAttribute{
				Description: "Name of the SSH key resource.",
				Required:    true,
				Validators: []validator.String{
					stringvalidator.RegexMatches(
						regexp.MustCompile(`^[a-z0-9]*$`),
						"must only contain only alphabetic characters",
					),
				},
			},
			"public_key": schema.StringAttribute{
				Description: "Public key value.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			// Outputs
			"id": schema.StringAttribute{
				Description: "UUID of the SSH key resource.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

func (r *sshKeyResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data sshKeyResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	createReq := &ssh.SshKeyCreateRequest{
		Name:      data.Name.ValueString(),
		PublicKey: data.PublicKey.ValueString(),
	}

	sshKey, err := r.svc.SSH.Create(ctx, createReq)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to create resource",
			"An unexpected error occurred while attempting to create the resource. "+
				"Please retry the operation or report this issue to the provider developers.\n\n"+
				err.Error(),
		)
		return
	}

	data.Id = types.StringValue(sshKey.Response.Key.Uuid)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *sshKeyResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data *sshKeyResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	info, err := r.svc.SSH.GetByID(ctx, data.Id.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to refresh resource",
			"An unexpected error occurred while attempting to refresh the resource."+
				"Please retry the operation or report this issue to the provider developers.\n\n"+
				err.Error(),
		)
		return
	}

	data.Name = types.StringValue(info.Name)
	data.PublicKey = types.StringValue(info.Value)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *sshKeyResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data sshKeyResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	updateReq := &ssh.SshKeyUpdateRequest{
		Uuid:      data.Id.ValueString(),
		Name:      data.Name.ValueString(),
		PublicKey: data.PublicKey.ValueString(),
	}

	_, err := r.svc.SSH.Update(ctx, updateReq)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to update resource",
			"An unexpected error occurred while attempting to update the resource."+
				"Please retry the operation or report this issue to the provider developers.\n\n"+
				err.Error(),
		)
		return
	}

	// OneProvider backend needs time to apply changes, so we retry with backoff.
	err = retry.RetryContext(ctx, time.Duration(30)*time.Second, func() *retry.RetryError {
		info, infoErr := r.svc.SSH.GetByID(ctx, data.Id.ValueString())
		if infoErr != nil {
			return retry.NonRetryableError(infoErr)
		}
		if info.Name != data.Name.ValueString() {
			return retry.RetryableError(fmt.Errorf("SSH key name is not updated yet"))
		}
		return nil
	})
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to refresh resource after update",
			"The update succeeded but failed to refresh the resource state."+
				"Please retry the operation or report this issue to the provider developers.\n\n"+
				err.Error(),
		)
		return
	}

	// PublicKey update is not supported at the moment.
	data.Name = types.StringValue(updateReq.Name)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *sshKeyResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data sshKeyResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.svc.SSH.Destroy(ctx, data.Id.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to destroy resource",
			"An unexpected error occurred while attempting to destroy the resource. "+
				"Please retry the operation or report this issue to the provider developers.\n\n"+
				err.Error(),
		)
		return
	}
}
