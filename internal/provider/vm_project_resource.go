package provider

import (
	"context"
	"errors"

	"github.com/MadJlzz/terraform-provider-oneprovider/pkg/oneprovider/projects"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ resource.ResourceWithConfigure   = &vmProjectResource{}
	_ resource.ResourceWithImportState = &vmProjectResource{}
)

type vmProjectResource struct {
	resourceServiceInjector
}

type vmProjectResourceModel struct {
	ID        types.String `tfsdk:"id"`
	Name      types.String `tfsdk:"name"`
	IsDefault types.Bool   `tfsdk:"is_default"`
}

func NewVmProjectResource() resource.Resource { return &vmProjectResource{} }

func (r *vmProjectResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_vm_project"
}

func (r *vmProjectResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a OneProvider VM project by UUID.",
		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{
				Description: "Name of the project. Changing the name does not change the UUID.",
				Required:    true,
				Validators:  []validator.String{stringvalidator.LengthAtLeast(1)},
			},
			"is_default": schema.BoolAttribute{
				Description: "Whether this is the account's default project. Defaults to false. Only one project should be configured as the default.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
			},
			"id": schema.StringAttribute{
				Description:   "Project UUID to use in the project attribute of VM instances and tags.",
				Computed:      true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
		},
	}
}

func (r *vmProjectResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data vmProjectResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	project, err := r.svc.Projects.Create(ctx, data.Name.ValueString(), data.IsDefault.ValueBool())
	if err != nil {
		resp.Diagnostics.AddError("Unable to create VM project", err.Error())
		return
	}
	data.ID = types.StringValue(project.Uuid)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *vmProjectResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data vmProjectResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	project, err := r.svc.Projects.GetByID(ctx, data.ID.ValueString())
	if errors.Is(err, projects.ErrNotFound) {
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		resp.Diagnostics.AddError("Unable to read VM project", err.Error())
		return
	}
	data.Name = types.StringValue(project.Name)
	data.IsDefault = types.BoolValue(project.IsDefault == 1)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *vmProjectResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data vmProjectResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.svc.Projects.Update(ctx, data.ID.ValueString(), data.Name.ValueString(), data.IsDefault.ValueBool()); err != nil {
		resp.Diagnostics.AddError("Unable to update VM project", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *vmProjectResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data vmProjectResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.svc.Projects.Delete(ctx, data.ID.ValueString()); err != nil {
		_, readErr := r.svc.Projects.GetByID(ctx, data.ID.ValueString())
		if !errors.Is(readErr, projects.ErrNotFound) {
			resp.Diagnostics.AddError("Unable to delete VM project", err.Error())
		}
	}
}

func (r *vmProjectResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
