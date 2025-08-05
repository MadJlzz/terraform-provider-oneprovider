package provider

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"testing"
)

const testAccSshKeyResource = `
resource "oneprovider_ssh_key" "key" {
	name       = "frodoshouse"
	public_key = "ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIMbAn3/YgZhhmsQIiGjOPOhODxpKXUo+LF3rFBvOOnYl"
}
`

const testAccSshKeyResourceNameUpdate = `
resource "oneprovider_ssh_key" "key" {
	name       = "frodosnewhouse"
	public_key = "ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIMbAn3/YgZhhmsQIiGjOPOhODxpKXUo+LF3rFBvOOnYl"
}
`

const testAccSshKeyResourceKeyUpdate = `
resource "oneprovider_ssh_key" "key" {
	name       = "frodosnewhouse"
	public_key = "ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIMbAn3/YgZhhmsQIiGjOPOhODxpKXUo+LF3rFBvOOnYw"
}
`

func TestAccSshKeyResource(t *testing.T) {
	var initialID string

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccSshKeyResource,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrWith("oneprovider_ssh_key.key", "id", func(value string) error {
						initialID = value
						return nil
					}),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"oneprovider_ssh_key.key",
						tfjsonpath.New("id"),
						knownvalue.NotNull(),
					),
					statecheck.ExpectKnownValue(
						"oneprovider_ssh_key.key",
						tfjsonpath.New("name"),
						knownvalue.StringExact("frodoshouse"),
					),
					statecheck.ExpectKnownValue(
						"oneprovider_ssh_key.key",
						tfjsonpath.New("public_key"),
						knownvalue.StringExact("ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIMbAn3/YgZhhmsQIiGjOPOhODxpKXUo+LF3rFBvOOnYl"),
					),
				},
			},
			{
				Config: testAccSshKeyResourceNameUpdate,
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"oneprovider_ssh_key.key",
						tfjsonpath.New("name"),
						knownvalue.StringExact("frodosnewhouse"),
					),
				},
			},
			{
				Config: testAccSshKeyResourceKeyUpdate,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrWith("oneprovider_ssh_key.key", "id", func(value string) error {
						if value == initialID {
							return fmt.Errorf("expected resource to be recreated, but ID remained the same: %s", value)
						}
						return nil
					}),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"oneprovider_ssh_key.key",
						tfjsonpath.New("name"),
						knownvalue.StringExact("frodosnewhouse"),
					),
					statecheck.ExpectKnownValue(
						"oneprovider_ssh_key.key",
						tfjsonpath.New("public_key"),
						knownvalue.StringExact("ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIMbAn3/YgZhhmsQIiGjOPOhODxpKXUo+LF3rFBvOOnYw"),
					),
				},
			},
		},
	})
}
