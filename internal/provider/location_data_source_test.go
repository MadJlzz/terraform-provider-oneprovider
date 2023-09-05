package provider

//
//func TestAccLocationDataSource(t *testing.T) {
//	resource.Test(t, resource.TestCase{
//		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
//		Steps: []resource.TestStep{
//			// Read testing
//			{
//				Config: providerConfig + `data "oneprovider_vm_templates" "test" {}`,
//				Check: resource.ComposeAggregateTestCheckFunc(
//					// Verify number of coffees returned
//					resource.TestCheckResourceAttr("data.oneprovider_vm_templates.test", "templates.#", "19"),
//					// Verify the first coffee to ensure all attributes are set
//					resource.TestCheckResourceAttr("data.oneprovider_vm_templates.test", "templates.0.id", "771"),
//					resource.TestCheckResourceAttr("data.oneprovider_vm_templates.test", "templates.0.name", "Debian 9.4 64bits"),
//					resource.TestCheckResourceAttr("data.oneprovider_vm_templates.test", "templates.0.size", "5368709120"),
//					resource.TestCheckResourceAttr("data.oneprovider_vm_templates.test", "templates.0.display.name", "debian"),
//					resource.TestCheckResourceAttr("data.oneprovider_vm_templates.test", "templates.0.display.display", "Debian 9.4 64bits"),
//					resource.TestCheckResourceAttr("data.oneprovider_vm_templates.test", "templates.0.display.description", ""),
//					resource.TestCheckResourceAttr("data.oneprovider_vm_templates.test", "templates.0.display.oca", "0"),
//					// Verify placeholder id attribute
//					resource.TestCheckResourceAttr("data.oneprovider_vm_templates.test", "id", "placeholder"),
//				),
//			},
//		},
//	})
//}
