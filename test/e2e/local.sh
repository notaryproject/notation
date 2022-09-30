#!/bin/sh
testdata_dir=`pwd`/test/e2e/testdata
export NOTATION_E2E_BINARY_PATH=$1
if [ ! -f "$NOTATION_E2E_BINARY_PATH" ];then
    echo "notation binary $NOTATION_E2E_BINARY_PATH not found"
    exit 1
fi

# load artifact to be signed
docker load -i ${testdata_dir}/images/net-monitor.tar.gz

# start distribution for testing purpose
docker run -d -p 5000:5000 --rm --name notation-e2e-registry \
    --mount type=bind,source=`pwd`/test/e2e/testdata/config/config-example-with-extensions.yml,target=/etc/docker/registry/config.yml \
    --mount type=bind,source=`pwd`/test/e2e/testdata/config/passwd,target=/etc/docker/registry/passwd \
    ghcr.io/oras-project/registry:latest 

# build notation image in case some test specs using docker
docker build -t notation-e2e -f Dockerfile `pwd`

# run tests
ginkgo -r -p `pwd`/test/e2e -v
