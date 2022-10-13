---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "rhoas_topic Data Source - terraform-provider-rhoas"
subcategory: ""
description: |-
  rhoas_topic provides a Topic accessible to your organization in Red Hat OpenShift Streams for Apache Kafka.
---

# rhoas_topic (Data Source)

`rhoas_topic` provides a Topic accessible to your organization in Red Hat OpenShift Streams for Apache Kafka.



<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `kafka_id` (String) The unique identifier for the Kafka instance
- `name` (String) The name of the Kafka topic

### Read-Only

- `id` (String) The ID of this resource.
- `partitions` (Number) The number of partitions in the topic

