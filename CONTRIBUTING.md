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

## Debuging in VSCode
Run the lanuch configuration called *Debug* in VsCode, this will output a value for an enviroment variable 
called `TF_REATTACH_PROVIDERS` in the Debug Console window, copy this output.

When running a terraform commmand set the enviroment variable `TF_REATTACH_PROVIDERS` to the value given
either through export or setting it inline before the command.
```shell
TF_REATTACH_PROVIDERS='{"provider": ... }' terraform apply
```

```shell
export TF_REATTACH_PROVIDERS='{"provider": ... }' 

terraform apply
```

## Debuging in GoLand
Create a configuration that matched the following, add the path to the root of the project for `Directory` and `Working Directory`.
And add the `--debug` flag to the program arguments.

Then run the project in debug mode this will output a value for an enviroment variable  called `TF_REATTACH_PROVIDERS` in the Debug Console window, copy this output.
When running a terraform commmand set the enviroment variable `TF_REATTACH_PROVIDERS` to the value given
either through export or setting it inline before the command.
```shell
TF_REATTACH_PROVIDERS='{"provider": ... }' terraform apply
```

```shell
export TF_REATTACH_PROVIDERS='{"provider": ... }' 

terraform apply
```

## Linting

1. Install [golangci-lint](https://golangci-lint.run/)
2. Run `make lint`

## Status

* All data providers are working
* the rhoas_kafka resource is working
