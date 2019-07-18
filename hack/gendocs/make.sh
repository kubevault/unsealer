#!/usr/bin/env bash

pushd $GOPATH/src/kubevault.dev/unsealer/hack/gendocs
go run main.go
popd
