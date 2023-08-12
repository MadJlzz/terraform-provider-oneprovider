![](./assets/logo.png)

# Terraform Provider OneProvider

Terraform provider for OneProvider. Provider documentation is available [here]().

## Contribute

### Requirements

- [Terraform](https://developer.hashicorp.com/terraform/downloads) >= 1.0
- [Go](https://golang.org/doc/install) >= 1.20

### Developing the provider

If you wish to work on the provider, you'll first need [Go](http://www.golang.org) installed on your machine (see [Requirements](#requirements) above).

To compile the provider, run `go install`. This will build the provider and put the provider binary in the `$GOPATH/bin` directory.

To generate or update documentation, run `go generate`.

In order to run the full suite of Acceptance tests, run `make testacc`.

*Note:* Acceptance tests create real resources, and often cost money to run.

```shell
make testacc
```

If you want to test your provider locally first, you'll have to create a `.terraformrc` file. Provider needs to be
compiled as well. (`go install`)
```text
provider_installation {

  dev_overrides {
      "registry.terraform.io/MadJlzz/oneprovider" = "<PATH TO YOUR GO BINARIES>"
  }

  # For all other providers, install them directly from their origin provider
  # registries as normal. If you omit this, Terraform will _only_ use
  # the dev_overrides block, and so no other providers will be available.
  direct {}
}
```

Then, simply run `terraform plan` on code that uses this provider.

# Roadmap

- [x] Datasource to list available OSes (templates): GET /vm/templates
- [ ] Datasource to list available locations: GET /vm/locations
- [ ] Resource to create a VM: POST /vm/create
```text
location_id	        Integer	Virtual server's location ID.
instance_size	        Integer	Instance's size ID. [A list of available sizes can be returned with the /vm/sizes/ call]
template	        String	ID of the OS' template or the UUID of the image to install [A list of available templates can be returned with the /vm/templates call and images with /vm/images/list]
hostname	        String	New hostname of the server/VM
sshKeys (optional)	Array	SSH Keys
```

- [ ] Resource to delete a VM: POST /vm/destroy
```text
vm_id	        Integer	Virtual server ID
confirm_close	Boolean	Parameter to confirm you want to pay the bandwidth overage
```