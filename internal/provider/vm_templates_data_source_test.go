package provider

import (
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"testing"
)

const testAccVmTemplatesDataSourceConfig = `
data "oneprovider_vm_templates" "test" {}
`

func TestAccVmTemplatesDataSource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Read testing
			{
				Config: testAccVmTemplatesDataSourceConfig,
				ConfigStateChecks: []statecheck.StateCheck{
					// Verify the placeholder ID is set correctly
					statecheck.ExpectKnownValue(
						"data.oneprovider_vm_templates.test",
						tfjsonpath.New("id"),
						knownvalue.StringExact("placeholder"),
					),
					// Verify templates list exists (may be empty)
					//statecheck.ExpectKnownValue(
					//	"data.oneprovider_vm_templates.test",
					//	tfjsonpath.New("templates"),
					//	knownvalue.NotNull(),
					//),
				},
			},
		},
	})
}
