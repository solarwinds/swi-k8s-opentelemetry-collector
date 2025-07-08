#!/bin/bash

echo "Setting up Helm repositories and plugins..."

# Install helm-unittest plugin
echo "Installing helm-unittest plugin..."
helm plugin install https://github.com/helm-unittest/helm-unittest.git

# Add required Helm repositories
echo "Adding Helm repositories..."
helm repo add prometheus-community https://prometheus-community.github.io/helm-charts
helm repo add open-telemetry https://open-telemetry.github.io/opentelemetry-helm-charts
helm repo add jetstack https://charts.jetstack.io

# Update repositories
echo "Updating Helm repositories..."
helm repo update

echo "Helm setup complete!"
echo "Available repositories:"
helm repo list
