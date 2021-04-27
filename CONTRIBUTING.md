# Terraform Provider for RHOAS

Run the following command to build the provider

```shell
make install
```

## Test sample configuration

First, build and install the provider.

```shell
make install
```

Get your offline token from https://cloud.redhat.com/openshift/token.

Then, run the following command to initialize the workspace and apply the sample configuration.

```shell
export OFFLINE_TOKEN=<offline token>
terraform init && terraform apply
```

## Status

* All data providers are working
* the rhoas_kafka resource is working
* the rhoas_service_account resource isn't working quite
