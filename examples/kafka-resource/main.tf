terraform {
  required_providers {
    rhoas = {
      version = "0.1"
      source  = "pmuir/rhoas"
    }
  }
}

provider "rhoas" {}

resource "rhoas_kafka" "foo" {
  kafka {
    name = "foo"
  }
}

output "bootstrap_server_foo" {
  value = rhoas_kafka.foo.kafka[0].bootstrap_server
}
