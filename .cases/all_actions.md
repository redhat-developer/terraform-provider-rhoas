## Perform All Actions

```
terraform {
  required_providers {
    rhoas = {
        source  = "registry.terraform.io/redhat-developer/rhoas"
        version = "0.1.0"
    }
  }
}

provider "rhoas" {
    offline_token = "..."
}

resource "rhoas_service_account" "srvcaccnt" {
  name = "service_account"
}

resource "rhoas_kafka" "instance" {
  name = "instance"
  plan = "developer.x1"
  billing_model = "standard"
  acl = [
    {
      principal = rhoas_service_account.srvcaccnt.client_id,
      resource_type = "TOPIC",
      resource_name = "topic-1",
      pattern_type = "LITERAL",
      operation_type = "ALL",
      permission_type = "ALLOW",
    },
  ]
}

resource "rhoas_topic" "topic-1" {
  kafka_id = rhoas_kafka.instance.id
  name = "topic-1"
  partitions = 1
}

resource "rhoas_topic" "topic-2" {
  kafka_id = rhoas_kafka.instance.id
  name = "topic-2"
  partitions = 1
}

resource "rhoas_acl" "acl" {
  kafka_id = rhoas_kafka.instance.id
  principal = rhoas_service_account.srvcaccnt.client_id
  resource_type = "TOPIC"
  resource_name = "topic-2"
  pattern_type = "LITERAL"
  operation_type = "ALL"
  permission_type = "ALLOW"
}

data "rhoas_kafka" "instance_data" {
  id = rhoas_kafka.instance.id
}

data "rhoas_service_account" "srvcaccnt_data" {
  id = rhoas_service_account.srvcaccnt.id
}
```