#!/bin/bash -e

export NOTATION_E2E_BINARY_PATH=$(realpath $1)
if [ ! -f "$NOTATION_E2E_BINARY_PATH" ];then
    echo "run.sh <notation-binary-path>"
    exit 1
fi

# install dependency
go install -mod=mod github.com/onsi/ginkgo/v2/ginkgo

# set environment variable for E2E testing
REG_HOST=localhost
REG_PORT=5000
ZOT_CONTAINER_NAME=zot

export NOTATION_E2E_REGISTRY_HOST=$REG_HOST:$REG_PORT
export NOTATION_E2E_REGISTRY_USERNAME=testuser
export NOTATION_E2E_REGISTRY_PASSWORD=testpassword
export NOTATION_E2E_KEY_PATH=`pwd`/testdata/config/localkeys/e2e.key
export NOTATION_E2E_CERT_PATH=`pwd`/testdata/config/localkeys/e2e.crt
export NOTATION_E2E_CONFIG_PATH=`pwd`/testdata/config
export NOTATION_E2E_OCI_LAYOUT_PATH=`pwd`/testdata/registry/oci_layout
export NOTATION_E2E_TEST_REPO=e2e
export NOTATION_E2E_TEST_TAG=v1
export REGISTRY_STORAGE_PATH=/tmp/zot-registry

# create temperory directory for Zot storage
mkdir -p /tmp/zot-registry && echo "Zot storage path: $REGISTRY_STORAGE_PATH created"

# start zot
docker run -d -p $REG_PORT:$REG_PORT -it --name $ZOT_CONTAINER_NAME \
    --mount type=bind,source=`pwd`/testdata/registry/zot/,target=/etc/zot \
    --mount type=bind,source=$REGISTRY_STORAGE_PATH,target=/var/lib/registry \
    --rm ghcr.io/project-zot/zot-minimal-linux-amd64:latest

# stop container and clean zot storage directory when exit
function cleanup {
    docker container stop $ZOT_CONTAINER_NAME 1>/dev/null && echo "Zot stopped"
    rm -rf $REGISTRY_STORAGE_PATH && echo "Zot storage path: $REGISTRY_STORAGE_PATH deleted"
}
trap cleanup EXIT

# run tests
ginkgo -r -p -v
