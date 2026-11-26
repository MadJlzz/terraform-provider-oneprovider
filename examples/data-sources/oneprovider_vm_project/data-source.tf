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

output "project_id" {
  value = data.oneprovider_vm_project.production.id
}
