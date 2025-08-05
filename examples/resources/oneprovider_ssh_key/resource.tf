terraform {
  required_providers {
    oneprovider = {
      source = "registry.terraform.io/MadJlzz/oneprovider"
    }
  }
}

provider "oneprovider" {}

resource "oneprovider_ssh_key" "ubuntu" {
  name       = "example"
  public_key = "ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIMbAn3/YgZhhmsQIiGjOPOhODxpKXUo+LF3rFBvOOnYl"
}
