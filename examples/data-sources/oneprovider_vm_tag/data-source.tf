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

data "oneprovider_vm_tag" "managed" {
  name    = "terraform managed"
  project = data.oneprovider_vm_project.production.id
}

output "tag_id" {
  value = data.oneprovider_vm_tag.managed.id
}
