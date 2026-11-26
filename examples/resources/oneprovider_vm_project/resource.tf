terraform {
  required_providers {
    oneprovider = {
      source = "registry.terraform.io/MadJlzz/oneprovider"
    }
  }
}

provider "oneprovider" {}

resource "oneprovider_vm_project" "production" {
  name       = "Production"
  is_default = false
}

# Use oneprovider_vm_project.production.id in the project attribute of a VM or tag.
