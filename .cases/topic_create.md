## Create Topic 
## Requirements
 - Offline token is accessible to terraform
 - Kafka instance is created 

## Cases

### Missing required fields
The field `partitions` is not defined and does not have default values.
```
resource "rhoas_topic" "topic" {
  name = "my-topic"
  kafka_id = rhoas_kafka.instance.id
}
```

### Topic creation success
```
resource "rhoas_topic" "topic" {
  name = "topic"
  partitions = 1
  kafka_id = rhoas_kafka.instance.id
}
```