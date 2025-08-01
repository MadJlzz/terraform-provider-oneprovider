terraform {
  required_providers {
    oneprovider = {
      source = "registry.terraform.io/MadJlzz/oneprovider"
    }
  }
}

provider "oneprovider" {}

data "oneprovider_vm_template" "ubuntu" {
  name = "Ubuntu 24.04 64bits"
}

data "oneprovider_vm_location" "paris" {
  city = "Paris"
}

resource "oneprovider_vm_instance" "vm" {
  location_id      = data.oneprovider_vm_location.paris.id
  instance_size_id = "45"
  template_id      = data.oneprovider_vm_template.ubuntu.id
  hostname         = "theOneRing"
}