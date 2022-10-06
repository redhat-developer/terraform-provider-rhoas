## Create Service Account 
## Requirements
 - Offline token is accessible to terraform

## Cases

### Service account creation success
```json
resource "resource_service_account" "srvcaccnt" {
  name = "my-service-account"
}
```