terraform {
  required_providers {
    oneprovider = {
      source = "registry.terraform.io/MadJlzz/oneprovider"
    }
  }
}

provider "oneprovider" {
  endpoint   = "http://localhost:3000"
  api_key    = "fakeApiKey"
  client_key = "fakeClientKey"
}

data "oneprovider_vm_templates" "example" {}

output "test" {
  value = data.oneprovider_vm_templates.example
}