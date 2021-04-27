terraform {
  required_providers {
    rhoas = {
      version = "0.1"
      source  = "pmuir/rhoas"
    }
  }
}

provider "rhoas" {}

data "rhoas_kafkas" "all" {
}

output "all_kafkas" {
  value = data.rhoas_kafkas.all
}
