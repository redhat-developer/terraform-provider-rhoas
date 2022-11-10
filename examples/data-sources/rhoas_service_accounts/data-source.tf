terraform {
  required_providers {
    rhoas = {
      source  = "registry.terraform.io/redhat-developer/rhoas"
      version = "0.3"
    }
  }
}

provider "rhoas" {}

resource "rhoas_service_account" "foo" {
  name = "foo"
}

data "rhoas_service_account" "foo" {
  id = rhoas_service_account.foo.id
}

data "rhoas_service_accounts" "all" {
}

output "all_service_accounts" {
  value = data.rhoas_service_accounts.all
}
