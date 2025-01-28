#!/usr/bin/env bash

set -x
set -o errexit
set -o pipefail 

SOURCE=$(dirname "$0")/..

DOMAIN="solarwinds.com"
GROUP="apps"

if [ -z "$DOCKERHUB_IMAGE" ]; then
  DOCKERHUB_IMAGE="solarwinds/solarwinds-otel-operator"
fi
if [ -z "$VERSION" ]; then
  VERSION="0.0.0"
fi

IMG=$DOCKERHUB_IMAGE:$VERSION
BUNDLE_IMG=$DOCKERHUB_IMAGE:$VERSION-bundle


rm $SOURCE/operator/swi-otel-operator -d -r || true
mkdir $SOURCE/operator/swi-otel-operator
cd $SOURCE/operator/swi-otel-operator

# Initialize the Helm operator project
operator-sdk init --plugins=helm --domain=$DOMAIN

# Create api
operator-sdk create api --helm-chart=../../deploy/helm               

# Build the operator image
make docker-build IMG=$IMG

##
## Create bundle requires CVS file, template is prepared and used
mkdir ./config/manifests/bases
cp ../swi-otel-operator.clusterserviceversion.yaml ./config/manifests/bases/swi-otel-operator.clusterserviceversion.yaml

# Generate the operator bundle
make bundle VERSION=$VERSION IMG=$IMG

# Build the bundle image
make bundle-build BUNDLE_IMG=$BUNDLE_IMG IMG=$IMG

# Validate the bundle
operator-sdk bundle validate ./bundle

cd -


