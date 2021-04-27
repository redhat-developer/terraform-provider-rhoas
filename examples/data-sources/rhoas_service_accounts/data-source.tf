terraform {
  required_providers {
    rhoas = {
      version = "0.1"
      source  = "pmuir/rhoas"
    }
  }
}

provider "rhoas" {}

data "rhoas_service_accounts" "all" {
}

output "all_service_accounts" {
  value = data.rhoas_service_accounts.all
}
