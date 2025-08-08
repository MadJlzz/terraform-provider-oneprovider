package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
)

const testAccSshKeyDataSourceConfig = `
resource "oneprovider_ssh_key" "random" {
	name       = "akey"
	public_key = "ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIMbAn3/YgZhhmsQIiGjOPOhODxpKXUo+LF3rFBvOOnYz"
}

data "oneprovider_ssh_key" "by_name" {
	name = oneprovider_ssh_key.random.name
}
`

func TestAccSshKeyDataSource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccSshKeyDataSourceConfig,
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"data.oneprovider_ssh_key.by_name",
						tfjsonpath.New("id"),
						knownvalue.NotNull(),
					),
					statecheck.ExpectKnownValue(
						"data.oneprovider_ssh_key.by_name",
						tfjsonpath.New("name"),
						knownvalue.StringExact("akey"),
					),
					statecheck.ExpectKnownValue(
						"data.oneprovider_ssh_key.by_name",
						tfjsonpath.New("public_key"),
						knownvalue.StringExact("ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIMbAn3/YgZhhmsQIiGjOPOhODxpKXUo+LF3rFBvOOnYz"),
					),
				},
			},
		},
	})
}
