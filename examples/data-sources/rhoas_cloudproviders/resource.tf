terraform {
  required_providers {
    rhoas = {
      source  = "registry.terraform.io/redhat-developer/rhoas"
      version = "0.3"
    }
  }
}

provider "rhoas" {}

data "rhoas_cloud_providers" "example" {
}

output "all_cloud_providers" {
  value = data.rhoas_cloud_providers.example
}


# Print the available regions for the first cloud provider
data "rhoas_cloud_provider_regions" "example" {
  id = "aws"
}

output "cloud_provider_regions_aws" {
  value = data.rhoas_cloud_provider_regions.example
}
