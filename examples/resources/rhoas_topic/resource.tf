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

resource "rhoas_topic" "bar" {
  name       = "bar-post"
  partitions = 4
  kafka_id   = rhoas_kafka.foo.id

  depends_on = [
    rhoas_kafka.foo
  ]
}

output "topic_bar" {
  value = rhoas_topic.bar
}
