# Examples

If you want to test your local changes instead of the latest version of the plugin, run `make install` and replace the provider with:

```tf
terraform {
  required_providers {
    rhoas = {
      source  = "registry.terraform.io/redhat-developer/rhoas"
      version = "0.1"
    }
  }
}
```

These are the `${HOSTNAME}/${NAMESPACE}/${NAME}/${VERSION}` variables from the [Makefile](../Makefile).
