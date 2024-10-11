#!/bin/bash

# Check if the correct number of arguments are provided
if [ $# -ne 3 ]; then
  echo "Usage: $0 <yaml-file> <values-yaml-file> <release>"
  exit 1
fi

YAML_FILE=$1
VALUES_YAML_FILE=$2
RELEASE=$3


# Check if the YAML file exists
if [ ! -f "$YAML_FILE" ]; then
  echo "Error: File '$YAML_FILE' not found!"
  exit 1
fi

# Check if the YAML file exists
if [ ! -f "$VALUES_YAML_FILE" ]; then
  echo "Error: File '$VALUES_YAML_FILE' not found!"
  exit 1
fi



# Check if yq is installed
if ! command -v yq &> /dev/null; then
  echo "yq is required but it's not installed. Please install it first."
  exit 1
fi

KERNEL_COLLECTOR=$(yq eval '.ebpfNetworkMonitoring.kernelCollector.image.repository + ":" + .ebpfNetworkMonitoring.kernelCollector.image.tag' $VALUES_YAML_FILE)
K8S_COLLECTOR_WATCHER=$(yq eval '.ebpfNetworkMonitoring.k8sCollector.watcher.image.repository + ":" + .ebpfNetworkMonitoring.k8sCollector.watcher.image.tag' $VALUES_YAML_FILE)
K8S_COLLECTOR_RELAY=$(yq eval '.ebpfNetworkMonitoring.k8sCollector.relay.image.repository + ":" + .ebpfNetworkMonitoring.k8sCollector.relay.image.tag' $VALUES_YAML_FILE)
REDUCER=$(yq eval '.ebpfNetworkMonitoring.reducer.image.repository + ":" + .ebpfNetworkMonitoring.reducer.image.tag' $VALUES_YAML_FILE)

echo "Found images:"
echo $KERNEL_COLLECTOR
echo $K8S_COLLECTOR_WATCHER
echo $K8S_COLLECTOR_RELAY
echo $REDUCER


yq eval "(.entries.\"swo-k8s-collector\"[] | select(.version == \"$RELEASE\").annotations.\"artifacthub.io/images\") = 
\"- name: ebpf-kernelCollector\n  image: $KERNEL_COLLECTOR\n  whitelisted: true
- name: ebpf-k8sCollectorWatcher\n  image: $K8S_COLLECTOR_WATCHER\n  whitelisted: true
- name: ebpf-k8sCollectorRelay\n  image: $K8S_COLLECTOR_RELAY\n  whitelisted: true
- name: ebpf-reducer\n  image: $REDUCER\n  whitelisted: true\"" -i "$YAML_FILE"


echo "Annotation added to release $RELEASE in $YAML_FILE."

