#!/bin/bash

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

# Delete all resources in the test-namespace
echo "Deleting all resources in the test-namespace"
kubectl delete all --all -n test-namespace