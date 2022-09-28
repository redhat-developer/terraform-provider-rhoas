terraform {
  required_providers {
    rhoas = {
      source  = "registry.terraform.io/redhat-developer/rhoas"
      version = "0.1"
    }
  }
}

provider "rhoas" {}

data "rhoas_kafkas" "all" {
}

output "all_kafkas" {
  value = data.rhoas_kafkas.all
}
