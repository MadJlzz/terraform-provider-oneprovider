package provider

import (
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

func TestAccSshKeyResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccSshKeyResource,
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"oneprovider_ssh_key.key",
						tfjsonpath.New("id"),
						knownvalue.NotNull(),
					),
				},
			},
		},
	})
}
