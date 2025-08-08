terraform {
  required_providers {
    oneprovider = {
      source = "registry.terraform.io/MadJlzz/oneprovider"
    }
  }
}

provider "oneprovider" {}

data "oneprovider_ssh_key" "by_id" {
  id = "id"
}

data "oneprovider_ssh_key" "by_name" {
  name = "keyname"
}