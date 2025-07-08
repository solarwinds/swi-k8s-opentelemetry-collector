#!/bin/bash -i

# Copies localhost's ~/.kube/config file into the container and swap out localhost
# for kubernetes.docker.internal whenever a new shell starts to keep them in sync.
if [ "$SYNC_LOCALHOST_KUBECONFIG" = "true" ] && [ -d "/usr/local/share/kube-localhost" ]; then
    mkdir -p $HOME/.kube
    sudo cp -r /usr/local/share/kube-localhost/* $HOME/.kube
    sudo chown -R $(id -u) $HOME/.kube
    
    # Docker Desktop Kubernetes uses kubernetes.docker.internal, not host.docker.internal
    sed -i -e "s/localhost:6443/kubernetes.docker.internal:6443/g" $HOME/.kube/config
    sed -i -e "s/127.0.0.1:6443/kubernetes.docker.internal:6443/g" $HOME/.kube/config
    sed -i -e "s/host\.docker\.internal:6443/kubernetes.docker.internal:6443/g" $HOME/.kube/config

    # Docker Desktop Kubernetes typically uses embedded certificates in the kubeconfig
    # No additional certificate handling needed unlike minikube
    echo "Kubernetes config synced for Docker Desktop"
fi