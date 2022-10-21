## Create kafka 
## Requirements
 - Offline token is accessible to terraform 

## Cases

### Missing required fields
The fields `plan` and `billing_model` are not defined and do not have default values.
```
resource "rhoas_kafka" "instance" {
  name = "my-instance"
}
```

### Entering computed fields 
The fields `owner` and `id` are computed by the provider and are not allowed to be defined in the terraform config.
```
resource "rhoas_kafka" "instance" {
  name = "my-instance"
  plan = "developer.x1"
  billing_model = "standard"
  owner = "my-username"
  id = "1234567890"
}
```

### Kafka creation success
```
resource "rhoas_kafka" "instance" {
  name = "my-instance"
  plan = "developer.x1"
  billing_model = "standard"
}
```