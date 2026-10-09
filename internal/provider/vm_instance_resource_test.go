package provider

import (
	"context"
	"errors"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/MadJlzz/terraform-provider-oneprovider/pkg/oneprovider"
	"github.com/MadJlzz/terraform-provider-oneprovider/pkg/oneprovider/client"
	"github.com/MadJlzz/terraform-provider-oneprovider/pkg/oneprovider/vm"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
)

func testAccVmInstanceResourceConfig(hostname string) string {
	return fmt.Sprintf(`
data "oneprovider_vm_size" "small" {name = "devd50c2"}
data "oneprovider_vm_location" "warsaw" {city = "Warsaw"}
resource "oneprovider_vm_instance" "ubuntu" {
	location_id      = data.oneprovider_vm_location.warsaw.id
	instance_size_id = data.oneprovider_vm_size.small.id
	template_id      = "1194"
	hostname         = %q
}
`, hostname)
}

func TestAccVmInstanceResource_removedOutOfBand(t *testing.T) {
	var vmID string

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccVmInstanceResourceConfig(testAccRandomName()),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrWith("oneprovider_vm_instance.ubuntu", "id", func(value string) error {
						vmID = value
						return nil
					}),
				),
			},
			{
				PreConfig: func() {
					svc, err := oneprovider.NewService(
						DefaultEndpoint,
						os.Getenv(ApiKeyEnvVar),
						os.Getenv(ClientKeyEnvVar),
					)
					if err != nil {
						t.Fatalf("failed to create service for out-of-band deletion: %v", err)
					}
					err = svc.VM.DestroyInstance(context.Background(), &vm.InstanceDestroyRequest{
						VmId:         vmID,
						ConfirmClose: true,
					})
					if err != nil {
						t.Fatalf("failed to destroy VM out-of-band: %v", err)
					}

					// Poll until the API confirms the VM is gone (error 810).
					deadline := time.Now().Add(2 * time.Minute)
					for time.Now().Before(deadline) {
						_, err = svc.VM.GetInstanceByID(context.Background(), vmID)
						if err != nil {
							var apiErr *client.APIError
							if errors.As(err, &apiErr) && apiErr.Code == 810 {
								break
							}
						}
						time.Sleep(5 * time.Second)
					}
					if err == nil {
						t.Fatalf("VM %s still exists after waiting for deletion", vmID)
					}
				},
				RefreshState:       true,
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccVmInstanceResource(t *testing.T) {
	hostname := testAccRandomName()
	updatedHostname := testAccRandomName()

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccVmInstanceResourceConfig(hostname),
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
						knownvalue.StringExact("47"),
					),
					statecheck.ExpectKnownValue(
						"oneprovider_vm_instance.ubuntu",
						tfjsonpath.New("instance_size_id"),
						knownvalue.StringExact("87"),
					),
					statecheck.ExpectKnownValue(
						"oneprovider_vm_instance.ubuntu",
						tfjsonpath.New("template_id"),
						knownvalue.StringExact("1194"),
					),
					statecheck.ExpectKnownValue(
						"oneprovider_vm_instance.ubuntu",
						tfjsonpath.New("hostname"),
						knownvalue.StringExact(hostname),
					),
				},
			},
			{
				Config: testAccVmInstanceResourceConfig(updatedHostname),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"oneprovider_vm_instance.ubuntu",
						tfjsonpath.New("hostname"),
						knownvalue.StringExact(updatedHostname),
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
						knownvalue.StringExact("47"),
					),
					statecheck.ExpectKnownValue(
						"oneprovider_vm_instance.ubuntu",
						tfjsonpath.New("instance_size_id"),
						knownvalue.StringExact("87"),
					),
					statecheck.ExpectKnownValue(
						"oneprovider_vm_instance.ubuntu",
						tfjsonpath.New("template_id"),
						knownvalue.StringExact("1194"),
					),
				},
			},

			// Delete testing automatically occurs in TestCase
		},
	})
}
