terraform {
  required_providers {
    oneprovider = {
      source = "registry.terraform.io/MadJlzz/oneprovider"
    }
  }
}

provider "oneprovider" {}

data "oneprovider_vm_size" "dev" {
  name = "01d20c1-2"
}
