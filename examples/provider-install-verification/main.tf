terraform {
  required_providers {
    oneprovider = {
      source = "registry.terraform.io/MadJlzz/oneprovider"
    }
  }
}

provider "oneprovider" {
  api_key    = "fakeApiKey"
  client_key = "fakeClientKey"
}

data "oneprovider_example" "example" {}
