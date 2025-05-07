#!/bin/bash

# Set namespace from TEST_CLUSTER_NAMESPACE env var or command line parameter with "test-namespace" as default
NAMESPACE=${TEST_CLUSTER_NAMESPACE:-${1:-"test-namespace"}}

# Ensure the .tmp directory exists
mkdir -p .tmp

# File to store the last Skaffold run ID
LAST_RUN_ID_FILE=".tmp/last_run_id"

# Check if the current Skaffold run ID matches the last run ID
if [ -f "$LAST_RUN_ID_FILE" ] && [ "$(cat $LAST_RUN_ID_FILE)" == "$SKAFFOLD_RUN_ID" ]; then
    echo "Skaffold run ID matches the last run ID. Skipping cleanup."
    exit 0
fi

# Save the current Skaffold run ID
echo "$SKAFFOLD_RUN_ID" > "$LAST_RUN_ID_FILE"

# Get the current Kubernetes context
current_context=$(kubectl config current-context)

# Check if the current context is 'docker-desktop' or 'default'
if [ "$current_context" != "docker-desktop" ] && [ "$current_context" != "default" ]; then
    echo "This script can only be run in the 'docker-desktop' or 'default' context. Current context is '$current_context'. Exiting gracefully."
    exit 0
fi

# Patch the CRD to remove finalizers
echo "Patching the CRD to remove finalizers"
kubectl patch crd/opentelemetrycollectors.opentelemetry.io -p '{"metadata":{"finalizers":[]}}' --type=merge

# Delete all resources in the specified namespace
echo "Deleting all resources in the $NAMESPACE namespace"
kubectl delete all --all -n $NAMESPACE

# Those resources are not deleted by the previous command
kubectl delete secrets --all -n $NAMESPACE
kubectl delete configmaps --all -n $NAMESPACE
kubectl delete persistentvolumeclaims --all -n $NAMESPACE
kubectl delete serviceaccounts --all -n $NAMESPACE
kubectl delete roles --all -n $NAMESPACE
kubectl delete rolebindings --all -n $NAMESPACE
kubectl delete networkpolicies --all -n $NAMESPACE

# Delete all CRDs from monitoring.coreos.com group
echo "Deleting all CRDs from monitoring.coreos.com group"
crds=$(kubectl get crd -o jsonpath='{range .items[?(@.spec.group=="monitoring.coreos.com")]}{.metadata.name}{"\n"}{end}')
if [ -n "$crds" ]; then
    kubectl delete crd $crds
else
    echo "No CRDs found in monitoring.coreos.com group"
fi

# Delete all CRDs from cert-manager.io group
echo "Deleting all CRDs from cert-manager.io group"
crds=$(kubectl get crd -o jsonpath='{range .items[?(@.spec.group=="cert-manager.io")]}{.metadata.name}{"\n"}{end}')
if [ -n "$crds" ]; then
    kubectl delete crd $crds
else
    echo "No CRDs found in cert-manager.io group"
fi

# Delete all CRDs from acme.cert-manager.io group
echo "Deleting all CRDs from acme.cert-manager.io group"
crds=$(kubectl get crd -o jsonpath='{range .items[?(@.spec.group=="acme.cert-manager.io")]}{.metadata.name}{"\n"}{end}')
if [ -n "$crds" ]; then
    kubectl delete crd $crds
else
    echo "No CRDs found in acme.cert-manager.io group"
fi

