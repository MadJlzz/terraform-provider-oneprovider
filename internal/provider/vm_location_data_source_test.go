package provider

import (
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
)

const testAccVmLocationDataSourceThatExistsConfig = `
data "oneprovider_vm_location" "fez" {city = "Fez"}
`
const testAccVmLocationDataSourceThatDontExistsConfig = `
data "oneprovider_vm_location" "fez" {city = "fez"}
`

func TestAccVmLocationDataSource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Read testing with an existing location.
			{
				Config: testAccVmLocationDataSourceThatExistsConfig,
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"data.oneprovider_vm_location.fez",
						tfjsonpath.New("id"),
						knownvalue.StringExact("198"),
					),
					statecheck.ExpectKnownValue(
						"data.oneprovider_vm_location.fez",
						tfjsonpath.New("region"),
						knownvalue.StringExact("Africa"),
					),
					statecheck.ExpectKnownValue(
						"data.oneprovider_vm_location.fez",
						tfjsonpath.New("country"),
						knownvalue.StringExact("MA"),
					),
					statecheck.ExpectKnownValue(
						"data.oneprovider_vm_location.fez",
						tfjsonpath.New("city"),
						knownvalue.StringExact("Fez"),
					),
				},
			},
			// Read testing with a location that does not exist.
			{
				Config:      testAccVmLocationDataSourceThatDontExistsConfig,
				ExpectError: regexp.MustCompile("location not found for city fez"),
			},
		},
	})
}
