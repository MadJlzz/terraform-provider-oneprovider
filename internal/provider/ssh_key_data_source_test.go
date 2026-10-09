package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
)

func testAccSshKeyDataSourceConfig(name, publicKey string) string {
	return fmt.Sprintf(`
resource "oneprovider_ssh_key" "random" {
	name       = %q
	public_key = %q
}

data "oneprovider_ssh_key" "by_name" {
	name = oneprovider_ssh_key.random.name
}
`, name, publicKey)
}

func TestAccSshKeyDataSource(t *testing.T) {
	name := testAccRandomName()
	publicKey := testAccRandomSSHPublicKey(t)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccSshKeyDataSourceConfig(name, publicKey),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"data.oneprovider_ssh_key.by_name",
						tfjsonpath.New("id"),
						knownvalue.NotNull(),
					),
					statecheck.ExpectKnownValue(
						"data.oneprovider_ssh_key.by_name",
						tfjsonpath.New("name"),
						knownvalue.StringExact(name),
					),
					statecheck.ExpectKnownValue(
						"data.oneprovider_ssh_key.by_name",
						tfjsonpath.New("public_key"),
						knownvalue.StringExact(publicKey),
					),
				},
			},
		},
	})
}
