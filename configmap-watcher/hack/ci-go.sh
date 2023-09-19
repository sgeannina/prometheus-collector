#!/usr/bin/env bash
set -euo pipefail

echo "installing apt dependencies"
apt update && apt install -y curl apt-transport-https

echo "setting up work dir"

WORKDIR="$(mktemp -d)"
pushd "$WORKDIR"

GOLANG_VERSION=${GOLANG_VERSION:-"1.17"}
echo "Downloading go"
curl -O https://dl.google.com/go/go${GOLANG_VERSION}.linux-amd64.tar.gz

echo "unpacking go"
tar -xvf go${GOLANG_VERSION}.linux-amd64.tar.gz -C /usr/local

echo "updating path to include go binaries"
export PATH="$PATH:/usr/local/go/bin"
echo "##vso[task.setvariable variable=path]${PATH}"
export GOPATH="/root/go"
echo "##vso[task.setvariable variable=gopath]${GOPATH}"

echo "Successfully installed go"
