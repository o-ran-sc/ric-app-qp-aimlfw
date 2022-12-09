#!/bin/bash

BASE_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null && pwd )"

DOCKER_FILE=${BASE_DIR}/Dockerfile
IMAGE_NAME=qoe-aiml-assist
IMAGE_TAG="1.0.0"
TARGET_URL=""

function install_go() {
        if [ ! "$(which go)" ]; then
                wget -c https://golang.org/dl/go1.17.linux-amd64.tar.gz -O - | sudo tar -xz -C /usr/local
                export PATH=$PATH:/usr/local/go/bin
        fi
}

pushd ${BASE_DIR}
        if [ -d $BASE_DIR/vendor ]; then
                rm -rf $BASE_DIR/vendor
        fi
        install_go
        go mod vendor
        if [ -z "${TARGET_URL}" ]; then
            docker build --no-cache -f ${DOCKER_FILE} -t ${IMAGE_NAME}:${IMAGE_TAG} ./ || exit
        elif
            docker build --no-cache -f ${DOCKER_FILE} -t ${TARGET_URL}/${IMAGE_NAME}:${IMAGE_TAG} ./ || exit
        fi
        rm -rf $BASE_DIR/vendor
popd
