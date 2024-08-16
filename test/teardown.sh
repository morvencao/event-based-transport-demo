#!/bin/bash -ex

# get current directory and root directory
CURRENT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"
ROOT_DIR=$(dirname $CURRENT_DIR)

if ! command -v kind >/dev/null 2>&1; then 
    echo "This script will install kind (https://kind.sigs.k8s.io/) on your machine."
    curl -Lo ./kind-amd64 "https://kind.sigs.k8s.io/dl/v0.12.0/kind-$(uname)-amd64"
    chmod +x ./kind-amd64
    sudo mv ./kind-amd64 /usr/local/bin/kind
fi

kind delete cluster --name edge
rm -rf ${ROOT_DIR}/test/.kubeconfig
