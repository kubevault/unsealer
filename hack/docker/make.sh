#!/bin/bash

set -o errexit
set -o nounset
set -o pipefail

GOPATH=$(go env GOPATH)
SRC=$GOPATH/src
BIN=$GOPATH/bin
ROOT=$GOPATH
REPO_ROOT=$GOPATH/src/github.com/kubevault/unsealer

source "$REPO_ROOT/hack/libbuild/common/lib.sh"
source "$REPO_ROOT/hack/libbuild/common/kubevault_image.sh"

APPSCODE_ENV=${APPSCODE_ENV:-dev}
IMG=unsealer

DIST=$GOPATH/src/github.com/kubevault/unsealer/dist
mkdir -p $DIST
if [ -f "$DIST/.tag" ]; then
    export $(cat $DIST/.tag | xargs)
fi

clean() {
    pushd $GOPATH/src/github.com/kubevault/unsealer/hack/docker
    rm vault-unsealer Dockerfile
    popd
}

build_binary() {
    pushd $GOPATH/src/github.com/kubevault/unsealer
    ./hack/builddeps.sh
    ./hack/make.py build vault-unsealer
    detect_tag $DIST/.tag
    popd
}

build_docker() {
    pushd $GOPATH/src/github.com/kubevault/unsealer/hack/docker
    cp $DIST/vault-unsealer/vault-unsealer-alpine-amd64 vault-unsealer
    chmod 755 vault-unsealer

    cat >Dockerfile <<EOL
FROM alpine

RUN set -x \
  && apk add --update --no-cache ca-certificates

COPY vault-unsealer /usr/bin/vault-unsealer

USER nobody:nobody
ENTRYPOINT ["vault-unsealer"]
EOL
    local cmd="docker build -t $DOCKER_REGISTRY/$IMG:$TAG ."
    echo $cmd; $cmd

    rm vault-unsealer Dockerfile
    popd
}

build() {
    build_binary
    build_docker
}

docker_push() {
    if [ "$APPSCODE_ENV" = "prod" ]; then
        echo "Nothing to do in prod env. Are you trying to 'release' binaries to prod?"
        exit 0
    fi
    if [ "$TAG_STRATEGY" = "git_tag" ]; then
        echo "Are you trying to 'release' binaries to prod?"
        exit 1
    fi
    hub_canary
}

docker_release() {
    if [ "$APPSCODE_ENV" != "prod" ]; then
        echo "'release' only works in PROD env."
        exit 1
    fi
    if [ "$TAG_STRATEGY" != "git_tag" ]; then
        echo "'apply_tag' to release binaries and/or docker images."
        exit 1
    fi
    hub_up
}

source_repo $@
