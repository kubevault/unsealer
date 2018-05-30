#!/usr/bin/env bash

pushd $GOPATH/src/github.com/kube-vault/unsealer/hack/gendocs
go run main.go
popd
