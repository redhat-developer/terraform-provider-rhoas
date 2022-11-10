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
      source  = "registry.terraform.io/redhat-developer/rhoas"
      version = "0.3"
    }
  }
}

provider "rhoas" {}

resource "rhoas_kafka" "foo" {
  name = "foo"
}

resource "rhoas_service_account" "bar" {
  name        = "bar"
  description = "a service account with permissions to use the new topic"

  depends_on = [
    rhoas_kafka.foo
  ]
}

resource "rhoas_topic" "baz" {
  name       = "baz"
  partitions = 5
  kafka_id   = rhoas_kafka.foo.id

  depends_on = [
    rhoas_kafka.foo
  ]
}

output "bootstrap_server_foo" {
  value = rhoas_kafka.foo.bootstrap_server_host
}

output "client_id" {
  value = rhoas_service_account.bar.client_id
}

output "client_secret" {
  value     = rhoas_service_account.bar.client_secret
  sensitive = true
}
```