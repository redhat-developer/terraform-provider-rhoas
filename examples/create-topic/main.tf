terraform {
  required_providers {
    rhoas = {
      version = "0.1"
      source  = "pmuir/rhoas"
    }
    kafka = {
      source = "Mongey/kafka"
    }
  }
}

provider "rhoas" {}

provider "kafka" {
  bootstrap_servers = [ rhoas_kafka.foo.kafka[0].bootstrap_server ]
  tls_enabled = true
  sasl_username = rhoas_service_account.foo.service_account[0].client_id
  sasl_password = rhoas_service_account.foo.service_account[0].client_secret
}

resource "rhoas_kafka" "foo" {
  kafka {
    name = "terraform-create-topic-1"
  }
}

resource "rhoas_service_account" "foo" {
  service_account {
    name = "create-topic"
    description = "blah blah blah"
  }
}

resource "kafka_topic" "prices" {
  name = "prices"
  partitions = 5
  replication_factor = 3
  config = {
    "cleanup.policy" = "delete"
  }
}

