#!/bin/bash

# Check if the correct number of arguments are provided
if [ $# -ne 2 ]; then
  echo "Usage: $0 <yaml-file> <release-type>"
  exit 1
fi

YAML_FILE=$1
RELEASE_TYPE=$2

# Check if the YAML file exists
if [ ! -f "$YAML_FILE" ]; then
  echo "Error: File '$YAML_FILE' not found!"
  exit 1
fi

# Check if yq is installed
if ! command -v yq &> /dev/null; then
  echo "yq is required but it's not installed. Please install it first."
  exit 1
fi


if [ "$RELEASE_TYPE" = "" ]; then
  # Production release
  yq eval ".annotations.\"artifacthub.io/prerelease\" = \"false\"" deploy/helm/Chart.yaml -i

else
  yq eval ".annotations.\"artifacthub.io/prerelease\" = \"true\"" deploy/helm/Chart.yaml -i

fi


