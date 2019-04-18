[![Go Report Card](https://goreportcard.com/badge/github.com/kubevault/unsealer)](https://goreportcard.com/report/github.com/kubevault/unsealer)
[![Build Status](https://travis-ci.org/kubevault/unsealer.svg?branch=master)](https://travis-ci.org/kubevault/unsealer)
[![codecov](https://codecov.io/gh/kubevault/unsealer/branch/master/graph/badge.svg)](https://codecov.io/gh/kubevault/unsealer)
[![Docker Pulls](https://img.shields.io/docker/pulls/kubevault/vault-unsealer.svg)](https://hub.docker.com/r/kubevault/vault-unsealer/)
[![Slack](https://slack.appscode.com/badge.svg)](https://slack.appscode.com)
[![Twitter](https://img.shields.io/twitter/follow/kubevault.svg?style=social&logo=twitter&label=Follow)](https://twitter.com/intent/follow?screen_name=KubeVault)

# Vault Unsealer

This project automates the process of [initializing](https://www.vaultproject.io/docs/commands/operator/init.html) and [unsealing](https://www.vaultproject.io/docs/concepts/seal.html#unsealing) HashiCorp Vault instances running.

## Installation
To install Vault operator & CSI driver, please follow the guide [here](https://github.com/kubevault/docs/blob/master/docs/setup/README.md).

## Using KubeVault
Want to learn how to use KubeVault? Please start [here](https://github.com/kubevault/docs/blob/master/docs/guides/README.md).

## Contribution guidelines
Want to help improve KubeVault? Please start [here](https://github.com/kubevault/docs/blob/master/docs/CONTRIBUTING.md).

---

**KubeVault binaries collects anonymous usage statistics to help us learn how the software is being used and how we can improve it. To disable stats collection, run the operator with the flag** `--enable-analytics=false`.

---

## Acknowledgement
This project started as a fork of [jetstack/vault-unsealer](https://github.com/jetstack/vault-unsealer).

## Support
We use Slack for public discussions. To chit chat with us or the rest of the community, join us in the [AppsCode Slack team](https://appscode.slack.com/messages/kubevault/) channel `#kubevault`. To sign up, use our [Slack inviter](https://slack.appscode.com/).

If you have found a bug with KubeVault or want to request for new features, please [file an issue](https://github.com/kubevault/project/issues/new).
