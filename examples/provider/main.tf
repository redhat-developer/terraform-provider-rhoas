terraform {
  required_providers {
    rhoas = {
      source  = "registry.terraform.io/redhat-developer/rhoas"
      version = "0.3"
    }
  }
}

provider "rhoas" {}
