package provider

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"testing"
)

func testAccSshKeyResourceConfig(name, publicKey string) string {
	return fmt.Sprintf(`
resource "oneprovider_ssh_key" "key" {
	name       = %q
	public_key = %q
}
`, name, publicKey)
}

func TestAccSshKeyResource(t *testing.T) {
	var initialID string

	name := testAccRandomName()
	updatedName := testAccRandomName()
	publicKey := testAccRandomSSHPublicKey(t)
	updatedPublicKey := testAccRandomSSHPublicKey(t)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccSshKeyResourceConfig(name, publicKey),
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
						knownvalue.StringExact(name),
					),
					statecheck.ExpectKnownValue(
						"oneprovider_ssh_key.key",
						tfjsonpath.New("public_key"),
						knownvalue.StringExact(publicKey),
					),
				},
			},
			{
				Config: testAccSshKeyResourceConfig(updatedName, publicKey),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"oneprovider_ssh_key.key",
						tfjsonpath.New("name"),
						knownvalue.StringExact(updatedName),
					),
				},
			},
			{
				Config: testAccSshKeyResourceConfig(updatedName, updatedPublicKey),
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
						knownvalue.StringExact(updatedName),
					),
					statecheck.ExpectKnownValue(
						"oneprovider_ssh_key.key",
						tfjsonpath.New("public_key"),
						knownvalue.StringExact(updatedPublicKey),
					),
				},
			},
		},
	})
}
