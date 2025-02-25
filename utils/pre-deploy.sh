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

# Those resources are not deleted by the previous command
kubectl delete secrets --all -n test-namespace
kubectl delete configmaps --all -n test-namespace
kubectl delete persistentvolumeclaims --all -n test-namespace
kubectl delete serviceaccounts --all -n test-namespace
kubectl delete roles --all -n test-namespace
kubectl delete rolebindings --all -n test-namespace
kubectl delete networkpolicies --all -n test-namespace

# Delete all CRDs from monitoring.coreos.com group
echo "Deleting all CRDs from monitoring.coreos.com group"
kubectl delete crd $(kubectl get crd -o jsonpath='{range .items[?(@.spec.group=="monitoring.coreos.com")]}{.metadata.name}{"\n"}{end}')

# Delete all CRDs from cert-manager.io group
echo "Deleting all CRDs from cert-manager.io group"
kubectl delete crd $(kubectl get crd -o jsonpath='{range .items[?(@.spec.group=="cert-manager.io")]}{.metadata.name}{"\n"}{end}')

# Delete all CRDs from cert-manager.io group
echo "Deleting all CRDs from acme.cert-manager.io group"
kubectl delete crd $(kubectl get crd -o jsonpath='{range .items[?(@.spec.group=="acme.cert-manager.io")]}{.metadata.name}{"\n"}{end}')

