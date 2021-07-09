terraform {
  required_providers {
    rhoas = {
      source  = "pmuir/rhoas"
    }
  }
}

provider "rhoas" {}

/*resource "rhoas_kafka" "foo" {
  kafka {
    name = "foo"
  }
}

output "bootstrap_server_foo" {
  value = rhoas_kafka.foo.kafka[0].bootstrap_server_host
}*/
