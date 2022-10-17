## Create ACL
## Requirements
 - Offline token is accessible to terraform
 - Kafka created with a topic named my-topic
 - Service account created and id is known

## Cases

### ACL creation success
```
resource "resource_acl" "acl" {
  kafka_id = rhoas_kafka.instance.id
  resource_type = "TOPIC"
  resource_name = "my-topic"
  pattern_type = "LITERAL"
  principal = rhoas_service_account.srvcaccnt.client_id
  operation = "ALL"
  permission = "ALLOW"  
}
```