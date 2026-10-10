package provider

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/MadJlzz/terraform-provider-oneprovider/pkg/oneprovider/client"
	"github.com/MadJlzz/terraform-provider-oneprovider/pkg/oneprovider/vm"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
)

var (
	_ resource.Resource                = &vmInstanceResource{}
	_ resource.ResourceWithConfigure   = &vmInstanceResource{}
	_ resource.ResourceWithImportState = &vmInstanceResource{}
)

type vmInstanceResource struct {
	resourceServiceInjector
}

type vmInstanceResourceModel struct {
	ID             types.String   `tfsdk:"id"`
	LocationId     types.String   `tfsdk:"location_id"`
	InstanceSizeId types.String   `tfsdk:"instance_size_id"`
	TemplateId     types.String   `tfsdk:"template_id"`
	Hostname       types.String   `tfsdk:"hostname"`
	IPAddress      types.String   `tfsdk:"ip_address"`
	Password       types.String   `tfsdk:"password"`
	SshKeys        types.List     `tfsdk:"ssh_keys"`
	Timeouts       timeouts.Value `tfsdk:"timeouts"`
}

func parseRequiredNumericID(value string) (int, error) {
	if value == "" {
		return 0, fmt.Errorf("ID must not be empty")
	}
	for _, digit := range value {
		if digit < '0' || digit > '9' {
			return 0, fmt.Errorf("ID must be a non-negative decimal integer")
		}
	}
	parsed, err := strconv.ParseInt(value, 10, strconv.IntSize)
	if err != nil {
		return 0, fmt.Errorf("ID is out of range for int: %w", err)
	}
	return int(parsed), nil
}

func requiredResponseID(id vm.APIID) (string, error) {
	value := id.String()
	if value == "" {
		return "", fmt.Errorf("response ID is empty")
	}
	return value, nil
}

func requiredCreateResponseID(id vm.APIID) (string, error) {
	return requiredResponseID(id)
}

type instanceInfoGetter func(context.Context, string) (*vm.InstanceReadResponse, error)

func waitForVMReady(ctx context.Context, timeout time.Duration, id vm.APIID, get instanceInfoGetter) error {
	instanceID, err := requiredCreateResponseID(id)
	if err != nil {
		return err
	}

	return retry.RetryContext(ctx, timeout, func() *retry.RetryError {
		info, infoErr := get(ctx, instanceID)
		if infoErr != nil {
			// The VM was just created with the same credentials, so a failing /vm/info call is most
			// likely a transient API or network issue. Keep polling until the timeout expires.
			return retry.RetryableError(fmt.Errorf("unable to read vm instance %s: %w", instanceID, infoErr))
		}
		if info.Response.ServerInstall || strings.ToLower(info.Response.ServerState.State) == "offline" {
			return retry.RetryableError(fmt.Errorf("vm instance not ready yet"))
		}
		if info.Response.ServerInfo.IpAddress == "" {
			// This should never been happening because when I do the create - I get an IP back.
			// The fact that from the GET endpoint, there is some cases where ServerInfo.* is filled with empty
			// values means that something is wrong in their backend.
			return retry.RetryableError(fmt.Errorf("getInstance returned empty informations"))
		}
		return nil
	})
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
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"instance_size_id": schema.StringAttribute{
				Description: "Instance size ID referencing the hardware specs of the VM instance",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"template_id": schema.StringAttribute{
				Description: "Template ID referencing the OS to use for that VM instance",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"hostname": schema.StringAttribute{
				Description: "Hostname of the VM instance",
				Required:    true,
			},
			"ssh_keys": schema.ListAttribute{
				Description: "List of SSH keys UUID to add to the VM instance. Note: The OneProvider API does not return SSH key information, so the state reflects the configured values rather than the actual server state.",
				ElementType: types.StringType,
				// Schema Using Attribute Default must be computed when using default.
				Computed: true,
				Optional: true,
				Default: listdefault.StaticValue(
					types.ListValueMust(
						types.StringType,
						[]attr.Value{},
					),
				),
			},
			// Outputs
			"id": schema.StringAttribute{
				Description: "ID of the VM instance. Generated by the provider.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"ip_address": schema.StringAttribute{
				Description: "IP address of the VM instance",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"password": schema.StringAttribute{
				Description: "Password of the root user",
				Computed:    true,
				Sensitive:   true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
		Blocks: map[string]schema.Block{
			"timeouts": timeouts.Block(ctx, timeouts.Opts{
				Create: true,
				Update: true,
			}),
		},
	}
}

func (r *vmInstanceResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data *vmInstanceResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	locationId, err := parseRequiredNumericID(data.LocationId.ValueString())
	if err != nil {
		resp.Diagnostics.AddAttributeError(
			path.Root("location_id"),
			"Invalid location_id",
			"location_id must be a non-negative decimal integer: "+err.Error(),
		)
		return
	}

	instanceSizeId, err := parseRequiredNumericID(data.InstanceSizeId.ValueString())
	if err != nil {
		resp.Diagnostics.AddAttributeError(
			path.Root("instance_size_id"),
			"Invalid instance_size_id",
			"instance_size_id must be a non-negative decimal integer: "+err.Error(),
		)
		return
	}

	if data.TemplateId.ValueString() == "" {
		resp.Diagnostics.AddAttributeError(
			path.Root("template_id"),
			"Invalid template_id",
			"template_id must not be empty.",
		)
		return
	}

	var sshKeys []string
	resp.Diagnostics.Append(data.SshKeys.ElementsAs(ctx, &sshKeys, false)...)
	if resp.Diagnostics.HasError() {
		return
	}

	createRequest := &vm.InstanceCreateRequest{
		LocationId:     locationId,
		InstanceSizeId: instanceSizeId,
		TemplateId:     data.TemplateId.ValueString(),
		Hostname:       data.Hostname.ValueString(),
		SshKeys:        sshKeys,
	}
	vmInstance, err := r.svc.VM.CreateInstance(ctx, createRequest)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to create resource",
			"An unexpected error occurred while attempting to create the resource."+
				"Please retry the operation or report this issue to the provider developers.\n\n"+
				err.Error(),
		)
		return
	}

	responseID, err := requiredCreateResponseID(vmInstance.Response.Id)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to create resource",
			fmt.Sprintf(
				"The API accepted the create request but returned an empty VM ID, so Terraform cannot track the VM. "+
					"A VM may still have been created (hostname %q, IP address %q); check the OneProvider panel "+
					"and delete or import it manually.",
				data.Hostname.ValueString(), vmInstance.Response.IpAddress,
			),
		)
		return
	}

	data.ID = types.StringValue(responseID)
	data.IPAddress = types.StringValue(vmInstance.Response.IpAddress)
	data.Password = types.StringValue(vmInstance.Response.Password)

	// Persist the VM in state before waiting for it to be ready.
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	createTimeout, diags := data.Timeouts.Create(ctx, 5*time.Minute)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	err = waitForVMReady(ctx, createTimeout, vmInstance.Response.Id, r.svc.VM.GetInstanceByID)
	if err != nil {
		resp.Diagnostics.AddError(
			"VM instance did not become ready",
			fmt.Sprintf(
				"VM instance %s was created but did not become ready within %s. "+
					"It has been saved to the state and marked as tainted, so it will be replaced on the next apply. "+
					"Consider increasing the create timeout using the timeouts block.\n\n%s",
				responseID, createTimeout, err.Error(),
			),
		)
		return
	}
}

func (r *vmInstanceResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data *vmInstanceResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	info, err := r.svc.VM.GetInstanceByID(ctx, data.ID.ValueString())
	if err != nil {
		var apiErr *client.APIError
		if errors.As(err, &apiErr) && apiErr.Code == 810 {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError(
			"Unable to refresh resource",
			"An unexpected error occurred while attempting to refresh the resource."+
				"Please retry the operation or report this issue to the provider developers.\n\n"+
				err.Error(),
		)
		return
	}

	// During import, only the ID is set. We need to populate required attributes
	// by looking them up from the API response.
	if data.LocationId.IsNull() || data.LocationId.ValueString() == "" {
		lr, lErr := r.svc.VM.GetLocationByCity(ctx, info.Response.ServerInfo.City)
		if lErr != nil {
			resp.Diagnostics.AddError(
				"Unable to refresh resource",
				"An unexpected error occurred while attempting to refresh the resource during an Import."+
					"Please retry the operation or report this issue to the provider developers.\n\n"+
					lErr.Error(),
			)
			return
		}
		if lr.Id.String() == "" {
			resp.Diagnostics.AddAttributeError(
				path.Root("location_id"),
				"Invalid location ID",
				"The API returned an empty location ID during import; refusing to write an empty ID to state.",
			)
			return
		}
		data.LocationId = types.StringValue(lr.Id.String())
	}
	if data.InstanceSizeId.IsNull() || data.InstanceSizeId.ValueString() == "" {
		is, isErr := r.svc.VM.GetSizeByName(ctx, info.Response.ServerInfo.Plan)
		if isErr != nil {
			resp.Diagnostics.AddError(
				"Unable to refresh resource",
				"An unexpected error occurred while attempting to refresh the resource during an Import."+
					"Please retry the operation or report this issue to the provider developers.\n\n"+
					isErr.Error(),
			)
			return
		}
		if is.Id.String() == "" {
			resp.Diagnostics.AddAttributeError(
				path.Root("instance_size_id"),
				"Invalid size ID",
				"The API returned an empty size ID during import; refusing to write an empty ID to state.",
			)
			return
		}
		data.InstanceSizeId = types.StringValue(is.Id.String())
	}
	if data.TemplateId.IsNull() || data.TemplateId.ValueString() == "" {
		ti, tiErr := r.svc.VM.GetTemplateByName(ctx, info.Response.ServerInfo.Template)
		if tiErr != nil {
			resp.Diagnostics.AddError(
				"Unable to refresh resource",
				"An unexpected error occurred while attempting to refresh the resource during an Import."+
					"Please retry the operation or report this issue to the provider developers.\n\n"+
					tiErr.Error(),
			)
			return
		}
		if ti.Id.String() == "" {
			resp.Diagnostics.AddAttributeError(
				path.Root("template_id"),
				"Invalid template ID",
				"The API returned an empty template ID during import; refusing to write an empty ID to state.",
			)
			return
		}
		data.TemplateId = types.StringValue(ti.Id.String())
	}

	data.Hostname = types.StringValue(info.Response.ServerInfo.Hostname)
	data.IPAddress = types.StringValue(info.Response.ServerInfo.IpAddress)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *vmInstanceResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var state *vmInstanceResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	var plan *vmInstanceResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)

	// We are updating the hostname...
	if state.Hostname != plan.Hostname {
		updateRequest := &vm.InstanceHostnameUpdateRequest{
			VmId:     plan.ID.ValueString(),
			Hostname: plan.Hostname.ValueString(),
		}

		err := r.svc.VM.UpdateInstanceHostname(ctx, updateRequest)
		if err != nil {
			resp.Diagnostics.AddError(
				"Unable to update resource",
				"An unexpected error occurred while attempting to update the resource."+
					"Please retry the operation or report this issue to the provider developers.\n\n"+
					err.Error(),
			)
			return
		}

		updateTimeout, diags := plan.Timeouts.Update(ctx, 30*time.Second)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}

		err = retry.RetryContext(ctx, updateTimeout, func() *retry.RetryError {
			info, infoErr := r.svc.VM.GetInstanceByID(ctx, plan.ID.ValueString())
			if infoErr != nil {
				return retry.NonRetryableError(infoErr)
			}
			if info.Response.ServerInfo.Hostname != plan.Hostname.ValueString() {
				return retry.RetryableError(fmt.Errorf("vm instance hostname not updated yet"))
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
		plan.Hostname = types.StringValue(updateRequest.Hostname)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *vmInstanceResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data *vmInstanceResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	destroyRequest := &vm.InstanceDestroyRequest{
		VmId:         data.ID.ValueString(),
		ConfirmClose: true,
	}

	err := r.svc.VM.DestroyInstance(ctx, destroyRequest)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to destroy resource",
			"An unexpected error occurred while attempting to destroy the resource."+
				"Please retry the operation or report this issue to the provider developers.\n\n"+
				err.Error(),
		)
		return
	}
}

func (r *vmInstanceResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
