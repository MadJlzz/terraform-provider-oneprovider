package provider

import (
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
)

const testAccVMTemplateDataSourceConfig = `data "oneprovider_vm_template" "ubuntu" { name = "Ubuntu 24.04" }`

const testAccVMTemplateDontExistDataSourceConfig = `data "oneprovider_vm_template" "ubuntu" { name = "random-name-that-does-not-exist" }`

func TestAccVMTemplateDataSource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccVMTemplateDataSourceConfig,
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"data.oneprovider_vm_template.ubuntu",
						tfjsonpath.New("name"),
						knownvalue.StringExact("Ubuntu 24.04"),
					),
					statecheck.ExpectKnownValue(
						"data.oneprovider_vm_template.ubuntu",
						tfjsonpath.New("id"),
						knownvalue.StringExact("100030"),
					),
					statecheck.ExpectKnownValue(
						"data.oneprovider_vm_template.ubuntu",
						tfjsonpath.New("size"),
						knownvalue.StringExact("5368709120")),
				},
			},
			{
				Config:      testAccVMTemplateDontExistDataSourceConfig,
				ExpectError: regexp.MustCompile("template not found for name random-name-that-does-not-exist"),
			},
		},
	})
}
