package provider

import (
	"context"
	"fmt"

	"github.com/MadJlzz/terraform-provider-oneprovider/pkg/oneprovider"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

type datasourceServiceInjector struct {
	svc *oneprovider.Service
}

func (dsi *datasourceServiceInjector) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	// Always perform a nil check when handling ProviderData because Terraform
	// sets that data after it calls the ConfigureProvider RPC.
	if req.ProviderData == nil {
		return
	}
	tflog.Info(ctx, "configuring datasource dependencies")
	svc, ok := req.ProviderData.(*oneprovider.Service)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected oneprovider.API, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}
	dsi.svc = svc
}

type resourceServiceInjector struct {
	svc *oneprovider.Service
}

func (rsi *resourceServiceInjector) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	// Always perform a nil check when handling ProviderData because Terraform
	// sets that data after it calls the ConfigureProvider RPC.
	if req.ProviderData == nil {
		return
	}
	tflog.Info(ctx, "configuring resource dependencies")
	svc, ok := req.ProviderData.(*oneprovider.Service)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected service type",
			fmt.Sprintf("Expected oneprovider.Service, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}
	rsi.svc = svc
}
