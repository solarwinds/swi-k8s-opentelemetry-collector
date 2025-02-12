#!/usr/bin/env bash

set -x
set -o errexit
set -o pipefail 

SOURCE=$(dirname "$0")/..

DOMAIN="com"
GROUP="solarwinds"

if [ -z "$DOCKERHUB_IMAGE" ]; then
  DOCKERHUB_IMAGE="solarwinds/solarwinds-otel-operator"
fi
if [ -z "$VERSION" ]; then
  VERSION="1.2.3"
fi

IMG=$DOCKERHUB_IMAGE:$VERSION
BUNDLE_IMG=$DOCKERHUB_IMAGE:$VERSION-bundle


rm $SOURCE/deploy/helm-openshift -d -r || true
mkdir $SOURCE/deploy/helm-openshift
cp -r $SOURCE/deploy/helm/* $SOURCE/deploy/helm-openshift/

cd $SOURCE/deploy/helm-openshift


### adjust the bundle metadata
yq eval -i ".annotations.\"charts.openshift.io/name\" = \"swo-k8s-collector\"" ./Chart.yaml
yq eval -i ".openshift.enabled = true" ./Values.yaml
yq eval -i ".ebpfNetworkMonitoring.enabled = false" ./Values.yaml

chart-verifier verify .


cd -

