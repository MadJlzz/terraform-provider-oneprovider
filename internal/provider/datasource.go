package provider

import (
	"context"
	"fmt"
	"github.com/MadJlzz/terraform-provider-oneprovider/pkg/oneprovider"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

type baseDatasource struct {
	svc *oneprovider.OneProvider
}

func (bd *baseDatasource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	// Always perform a nil check when handling ProviderData because Terraform
	// sets that data after it calls the ConfigureProvider RPC.
	if req.ProviderData == nil {
		return
	}
	tflog.Info(ctx, "configuring datasource dependencies")
	svc, ok := req.ProviderData.(*oneprovider.OneProvider)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected oneprovider.API, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}
	bd.svc = svc
}
