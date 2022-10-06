## Create kafka 
## Requirements
 - Offline token is accessible to terraform 

## Cases

### Missing required fields
The fields `cloud_provider` and `region` are not defined and do not have default values.
```json 
resource "rhoas_kafka" "instance" {
  name = "my-instance"
}
```

### Entering computed fields 
The fields `owner` and `id` are computed by the provider and are not allowed to be defined in the terraform config.
```json
resource "rhoas_kafka" "instance" {
  name = "my-instance"
  cloud_provider = "aws"
  region = "us-east-1"
  owner = "my-username"
  id = "1234567890"
}
```

### Kafka creation success
```json
resource "rhoas_kafka" "instance" {
  name = "my-instance"
  cloud_provider = "aws"
  region = "us-east-1"
}
```