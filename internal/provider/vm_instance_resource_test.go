package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
)

const testAccVmInstanceResource = `
resource "oneprovider_vm_instance" "ubuntu" {
	location_id      = "34"
	instance_size_id = "45"
	template_id      = "100030"
	hostname         = "ubuntu-test"
}
`

const testAccVmInstanceResourceUpdate = `
resource "oneprovider_vm_instance" "ubuntu" {
	location_id      = "34"
	instance_size_id = "45"
	template_id      = "100030"
	hostname         = "ubuntu-test-updated"
}
`

func TestAccVmInstanceResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccVmInstanceResource,
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"oneprovider_vm_instance.ubuntu",
						tfjsonpath.New("id"),
						knownvalue.NotNull(),
					),
					statecheck.ExpectKnownValue(
						"oneprovider_vm_instance.ubuntu",
						tfjsonpath.New("ip_address"),
						knownvalue.NotNull(),
					),
					statecheck.ExpectKnownValue(
						"oneprovider_vm_instance.ubuntu",
						tfjsonpath.New("password"),
						knownvalue.NotNull(),
					),
					statecheck.ExpectKnownValue(
						"oneprovider_vm_instance.ubuntu",
						tfjsonpath.New("location_id"),
						knownvalue.StringExact("34"),
					),
					statecheck.ExpectKnownValue(
						"oneprovider_vm_instance.ubuntu",
						tfjsonpath.New("instance_size_id"),
						knownvalue.StringExact("45"),
					),
					statecheck.ExpectKnownValue(
						"oneprovider_vm_instance.ubuntu",
						tfjsonpath.New("template_id"),
						knownvalue.StringExact("100030"),
					),
					statecheck.ExpectKnownValue(
						"oneprovider_vm_instance.ubuntu",
						tfjsonpath.New("hostname"),
						knownvalue.StringExact("ubuntu-test"),
					),
				},
			},
			{
				Config: testAccVmInstanceResourceUpdate,
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"oneprovider_vm_instance.ubuntu",
						tfjsonpath.New("hostname"),
						knownvalue.StringExact("ubuntu-test-updated"),
					),
					statecheck.ExpectKnownValue(
						"oneprovider_vm_instance.ubuntu",
						tfjsonpath.New("id"),
						knownvalue.NotNull(),
					),
					statecheck.ExpectKnownValue(
						"oneprovider_vm_instance.ubuntu",
						tfjsonpath.New("ip_address"),
						knownvalue.NotNull(),
					),
					statecheck.ExpectKnownValue(
						"oneprovider_vm_instance.ubuntu",
						tfjsonpath.New("password"),
						knownvalue.NotNull(),
					),
					statecheck.ExpectKnownValue(
						"oneprovider_vm_instance.ubuntu",
						tfjsonpath.New("location_id"),
						knownvalue.StringExact("34"),
					),
					statecheck.ExpectKnownValue(
						"oneprovider_vm_instance.ubuntu",
						tfjsonpath.New("instance_size_id"),
						knownvalue.StringExact("45"),
					),
					statecheck.ExpectKnownValue(
						"oneprovider_vm_instance.ubuntu",
						tfjsonpath.New("template_id"),
						knownvalue.StringExact("100030"),
					),
				},
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}
