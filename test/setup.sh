#!/bin/bash -ex

# get current directory and root directory
CURRENT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"
ROOT_DIR=$(dirname $CURRENT_DIR)

kind_version=0.12.0

if ! command -v kind >/dev/null 2>&1; then
    echo "This script will install kind (https://kind.sigs.k8s.io/) on your machine."
    curl -Lo ./kind-amd64 "https://kind.sigs.k8s.io/dl/v${kind_version}/kind-$(uname)-amd64"
    chmod +x ./kind-amd64
    sudo mv ./kind-amd64 /usr/local/bin/kind
fi

cat << EOF | kind create cluster --name edge --kubeconfig ${ROOT_DIR}/test/.kubeconfig --config=-
kind: Cluster
apiVersion: kind.x-k8s.io/v1alpha4
nodes:
- role: control-plane
  extraPortMappings:
  - containerPort: 30080
    hostPort: 30080
  - containerPort: 31883
    hostPort: 31883
EOF
export KUBECONFIG=${ROOT_DIR}/test/.kubeconfig

export image_repository=quay.io/morvencao/event-based-transport-demo
export image_tag=latest
make image

kind load docker-image --name edge eclipse-mosquitto:2.0.18
kind load docker-image --name edge ${image_repository}:$image_tag

kubectl create ns mqtt || true
kubectl apply -f ${ROOT_DIR}/deploy/mqtt.yaml
kubectl create ns agent || true
kubectl apply -f ${ROOT_DIR}/deploy/agent.yaml
