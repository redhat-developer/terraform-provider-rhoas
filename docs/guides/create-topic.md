---
subcategory: ""
page_title: "Create a Kafka Topic - RHOAS provider"
description: |-
    An example of creating a new Kafka instance and then creating a topic
---

# Create a Topic on a new Red Hat OpenShift Streams for Apache Kafka instance

To create a Kafka instance, a service account for using it, and then a topic:

```terraform
terraform {
  required_providers {
    rhoas = {
      source  = "pmuir/rhoas"
    }
    kafka = {
      source = "Mongey/kafka"
      version = "0.2.12"
    }
  }
}

provider "rhoas" {}

provider "kafka" {
  bootstrap_servers = [ rhoas_kafka.foo.kafka[0].bootstrap_server_host ]
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
```
