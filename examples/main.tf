terraform {
  required_providers {
    rhoas = {
      version = "0.1"
      source  = "redhat.com/cloud/rhoas"
    }
  }
}

provider "rhoas" {}

data "rhoas_kafkas" "all" {
}

data "rhoas_service_accounts" "all" {
}

output "all_kafkas" {
  value = data.rhoas_kafkas.all.kafkas
}
