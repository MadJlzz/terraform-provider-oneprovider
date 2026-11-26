package provider

import (
	"context"
	"errors"
	"strings"

	"github.com/MadJlzz/terraform-provider-oneprovider/pkg/oneprovider/tags"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ resource.ResourceWithConfigure   = &vmTagResource{}
	_ resource.ResourceWithImportState = &vmTagResource{}
)

type vmTagResource struct {
	resourceServiceInjector
}

type vmTagResourceModel struct {
	ID          types.String `tfsdk:"id"`
	Project     types.String `tfsdk:"project"`
	Label       types.String `tfsdk:"label"`
	Description types.String `tfsdk:"description"`
}

func NewVmTagResource() resource.Resource { return &vmTagResource{} }

func (r *vmTagResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_vm_tag"
}

func (r *vmTagResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a tag in a VM project. Destroying the tag also removes its assignments from all services. To detach a tag from just one VM, edit that VM's tags instead.",
		Attributes: map[string]schema.Attribute{
			"label": schema.StringAttribute{
				Description: "Label of the tag.",
				Required:    true,
				Validators:  []validator.String{stringvalidator.LengthAtLeast(1)},
			},
			"description": schema.StringAttribute{
				Description: "Description of the tag. Defaults to an empty string.",
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString(""),
			},
			"project": schema.StringAttribute{
				Description:   "UUID of the project that owns the tag. Changing it replaces the tag.",
				Required:      true,
				Validators:    []validator.String{stringvalidator.LengthAtLeast(1)},
				PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"id": schema.StringAttribute{
				Description: "Tag ID to use in the tags attribute of a VM instance.",
				Computed:    true,
			},
		},
	}
}

func (r *vmTagResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data vmTagResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	tag, err := r.svc.Tags.Create(ctx, data.Project.ValueString(), data.Label.ValueString(), data.Description.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Unable to create VM tag", err.Error())
		return
	}
	data.ID = types.StringValue(tag.Id.String())
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *vmTagResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data vmTagResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	tag, err := r.svc.Tags.GetByID(ctx, data.Project.ValueString(), data.ID.ValueString())
	if errors.Is(err, tags.ErrNotFound) {
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		resp.Diagnostics.AddError("Unable to read VM tag", err.Error())
		return
	}
	data.Label = types.StringValue(tag.Label)
	data.Description = types.StringValue(tag.Description)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *vmTagResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data vmTagResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	resp.Diagnostics.Append(req.State.GetAttribute(ctx, path.Root("id"), &data.ID)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.svc.Tags.Update(ctx, data.ID.ValueString(), data.Label.ValueString(), data.Description.ValueString()); err != nil {
		resp.Diagnostics.AddError("Unable to update VM tag", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *vmTagResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data vmTagResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.svc.Tags.Delete(ctx, data.ID.ValueString()); err != nil {
		_, readErr := r.svc.Tags.GetByID(ctx, data.Project.ValueString(), data.ID.ValueString())
		if !errors.Is(readErr, tags.ErrNotFound) {
			resp.Diagnostics.AddError("Unable to delete VM tag", err.Error())
		}
	}
}

func (r *vmTagResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	project, id, ok := strings.Cut(req.ID, "/")
	if !ok || project == "" || id == "" || strings.Contains(id, "/") {
		resp.Diagnostics.AddError("Invalid VM tag import ID", "Expected project_uuid/tag_id.")
		return
	}
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("project"), project)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), id)...)
}
