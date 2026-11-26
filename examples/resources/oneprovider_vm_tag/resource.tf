terraform {
  required_providers {
    oneprovider = {
      source = "registry.terraform.io/MadJlzz/oneprovider"
    }
  }
}

provider "oneprovider" {}

data "oneprovider_vm_project" "production" {
  name = "Production"
}

resource "oneprovider_vm_tag" "managed" {
  project     = data.oneprovider_vm_project.production.id
  label       = "terraform managed"
  description = "VMs managed by Terraform"
}

# To assign the tag in a oneprovider_vm_instance resource:
# tags = [oneprovider_vm_tag.managed.id]
