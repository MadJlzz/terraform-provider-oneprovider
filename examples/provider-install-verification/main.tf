terraform {
  required_providers {
    oneprovider = {
      source = "registry.terraform.io/MadJlzz/oneprovider"
    }
  }
}

provider "oneprovider" {
  host = "test"
}

data "oneprovider_example" "example" {}

output "test" {
  value = data.oneprovider_example.example.id
}
