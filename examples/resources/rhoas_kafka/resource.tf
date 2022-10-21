terraform {
  required_providers {
    rhoas = {
      source  = "registry.terraform.io/redhat-developer/rhoas"
      version = "0.1"
    }
  }
}

provider "rhoas" {}

resource "rhoas_kafka" "foo" {
  name = "foo"
  plan = "developer.x1"
  billing_model = "standard"
}

output "bootstrap_server_foo" {
  value = rhoas_kafka.foo.bootstrap_server_host
}
