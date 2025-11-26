package provider

import (
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
)

const testAccVmSizeDataSourceThatExistsConfig = `
data "oneprovider_vm_size" "small" {name = "01d20c1"}
`
const testAccVmSizeDataSourceThatDontExistsConfig = `
data "oneprovider_vm_size" "unknown" {name = "aRandomSize"}
`

func TestAccVmSizeDataSource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Read testing with an existing location.
			{
				Config: testAccVmSizeDataSourceThatExistsConfig,
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"data.oneprovider_vm_size.small",
						tfjsonpath.New("id"),
						knownvalue.StringExact("44"),
					),

					statecheck.ExpectKnownValue(
						"data.oneprovider_vm_size.small",
						tfjsonpath.New("name"),
						knownvalue.StringExact("01d20c1"),
					),
				},
			},
			// Read testing with a location that does not exist.
			{
				Config:      testAccVmSizeDataSourceThatDontExistsConfig,
				ExpectError: regexp.MustCompile("location not found for city fez"),
			},
		},
	})
}
