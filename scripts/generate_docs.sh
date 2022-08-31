#!/bin/bash

# run from the root directory of the project

docker build --platform linux/amd64 -t rhoas_provider $(pwd)/scripts
docker run --platform linux/amd64 --mount type=bind,source="$(pwd)",target=/terraform-provider-rhoas -d rhoas_provider