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

## Debugging in VSCode
Run the launch configuration called *Debug* in VsCode, this will output a value for an environment variable 
called `TF_REATTACH_PROVIDERS` in the Debug Console window, copy this output.

When running a terraform command set the environment variable `TF_REATTACH_PROVIDERS` to the value given
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

Then run the project in debug mode this will output a value for an environment variable  called `TF_REATTACH_PROVIDERS` in the Debug Console window, copy this output.
When running a terraform command set the environment variable `TF_REATTACH_PROVIDERS` to the value given
either through export or setting it inline before the command.
```shell
TF_REATTACH_PROVIDERS='{"provider": ... }' terraform apply
```

```shell
export TF_REATTACH_PROVIDERS='{"provider": ... }' 

terraform apply
```

## Internationalization

All text strings are placed in `./rhoas/localize/locales/en` directory containing `.toml` files.
These files are used in:

- Provider itself - all printed messages/strings
- generation of the documentation

This directory contains number of `toml` files that are used for:

1. Data source and resource definitions for later use in generated documentation
2. Provider output and error messages that aren't included in the generated documentation.

Each time we change any strings in data source and resource definitions we should regenerate markdown documentation files and push them with the PR.

## Using Provider with Mock RHOAS API

RHOAS SDK provides mock for all supported APIs.
To use mock you need to have NPM installed on your system and have free port 8000
To work and test provider locally please follow the [mock readme](https://github.com/redhat-developer/app-services-sdk-js/tree/main/packages/api-mock).

Define the LOCAL_DEV enviroment variable to use the locally running mock server.
```shell
LOCAL_DEV=http://localhost:8000 terraform apply
```

### Logging in

To log in to the mock API, run `rhoas login` against the local server with your authentication token:

```shell
rhoas login --api-gateway=http://localhost:8000
```
    
## Running Tests
1. To run the unit tests, run `make test` in the root of the project
2. To run the acceptance tests, run `make testacc` in the root of the project.

## Linting

1. Install [golangci-lint](https://golangci-lint.run/)
2. Run `make lint`