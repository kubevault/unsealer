#!/usr/bin/env bash

pushd $GOPATH/src/github.com/kubevault/unsealer/hack/gendocs
go run main.go
popd
