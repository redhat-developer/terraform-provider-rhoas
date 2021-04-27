terraform {
  required_providers {
    rhoas = {
      source  = "pmuir/rhoas"
    }
  }
}

provider "rhoas" {}

resource "rhoas_service_account" "foo" {
  service_account {
    name = "foo"
    description = "blah blah blah"
  }
}

output "client_id" {
  value = rhoas_service_account.foo.service_account[0].client_id
}

output "client_secret" {
  value = rhoas_service_account.foo.service_account[0].client_secret
  sensitive = true
}
