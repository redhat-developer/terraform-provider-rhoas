terraform {
  required_providers {
    rhoas = {
      source  = "registry.terraform.io/redhat-developer/rhoas"
      version = "0.1"
    }
  }
}

provider "rhoas" {}

resource "rhoas_service_account" "foo" {
  name        = "foo"
  description = "blah blah blah"
}

output "client_id" {
  value = rhoas_service_account.foo.client_id
}

output "client_secret" {
  value     = rhoas_service_account.foo.client_secret
  sensitive = true
}
