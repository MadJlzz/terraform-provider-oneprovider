---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "oneprovider_vm_location Data Source - oneprovider"
subcategory: ""
description: |-
  Retrieve location information given its city
---

# oneprovider_vm_location (Data Source)

Retrieve location information given its city

## Example Usage

```terraform
terraform {
  required_providers {
    oneprovider = {
      source = "registry.terraform.io/MadJlzz/oneprovider"
    }
  }
}

provider "oneprovider" {}

data "oneprovider_vm_location" "fez" {
  city = "Fez"
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `city` (String) Filter by city

### Read-Only

- `available_sizes` (List of Number) List of available VM sizes
- `available_types` (List of String) List of available VM types
- `country` (String) Location country
- `id` (String) Location ID
- `ipv4` (String) Location IPv4 address
- `ipv6` (String) Location IPv6 address
- `region` (String) Location region
