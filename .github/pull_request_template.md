## Verify
### Terraform config (main.tf)
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
```
### Commands
- `make clean`
- `make install`
- `terraform init`
- `terraform apply`

### Expected Results
