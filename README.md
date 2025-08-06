![](./assets/logo.png)

# Terraform Provider OneProvider

> **⚠️ Disclaimer**: This Terraform provider is currently in active development and has not yet reached v1.0.
> Until the project reaches v1, breaking changes may be introduced in minor or patch releases.
> Then, semantic versioning principles will be used to ensure backward compatibility within major versions.

Terraform provider for OneProvider. Provider documentation is available [here](https://registry.terraform.io/providers/MadJlzz/oneprovider/latest).

## Contribute

### Requirements

- [Terraform](https://developer.hashicorp.com/terraform/downloads) >= 1.12
- [Go](https://golang.org/doc/install) >= 1.23.7

### Developing the provider

If you wish to work on the provider, you'll first need [Go](http://www.golang.org) installed on your machine (see [Requirements](#requirements) above).

To compile the provider, run `make install`. This will build the provider and put the provider binary in the `$GOPATH/bin` directory.

To generate or update documentation, run `make generate`.

To run the full suite of Acceptance tests, run `make testacc`.

*Note:* Acceptance tests create real resources and often cost money to run.

```shell
# Don't forget to set API keys beforehand.
export ONEPROVIDER_API_KEY=""
export ONEPROVIDER_CLIENT_KEY=""
make testacc
```

If you want to test your provider locally first, you'll have to create a `.terraformrc` file. Provider needs to be
compiled as well. (`make install`)
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
