# AGENTS.md

This file provides guidance to coding agents (e.g. Claude Code, claude.ai/code) when working with code in this repository.

## Repository purpose

Go module `kubevault.dev/unsealer` — automates [Vault initialization](https://www.vaultproject.io/docs/commands/operator/init.html) and [unsealing](https://www.vaultproject.io/docs/concepts/seal.html#unsealing) for HashiCorp Vault instances running on Kubernetes. KubeVault deploys this as a sidecar (or DaemonSet) alongside the Vault server; on startup, it initializes Vault if needed, persists the resulting unseal keys + root token to a configured key-value store, and unseals the Vault process whenever it comes up sealed.

Supported key stores:
- AWS KMS-encrypted SSM Parameter Store (`pkg/kv/aws_kms`, `pkg/kv/aws_ssm`).
- Azure Key Vault (`pkg/kv/azure`).
- Google Cloud KMS-encrypted GCS objects (`pkg/kv/cloudkms`, `pkg/kv/gcs`).
- Kubernetes Secrets (`pkg/kv/kubernetes`).

The produced binary is `vault-unsealer`.

## Architecture

- `cmd/vault-unsealer/main.go` — entry point.
- `pkg/cmds/`:
  - `root.go` — top-level Cobra command.
  - `run.go` — long-running unsealer process.
- `pkg/worker/`:
  - `options.go` — runtime options (which key store, check interval, init thresholds).
  - `worker.go` — main control loop. Polls Vault status; if uninitialized → init + write keys; if sealed → read keys + submit unseal shares.
- `pkg/kv/` — **pluggable key-value store backends** for unseal keys + root token:
  - One subdirectory per backend (`aws_kms`, `aws_ssm`, `azure`, `cloudkms`, `gcs`, `kubernetes`).
  - `storage.go` — shared interface.
  - `util/` — shared helpers.
- `pkg/vault/`:
  - `vault.go` — Vault HTTP client wrapper.
  - `auth/`, `policy/`, `unseal/`, `util/` — auth setup, policy seed, unseal submission, helpers.
- `bin/` — local build artifacts (gitignored).
- `Dockerfile.in` (PROD, distroless), `Dockerfile.dbg` (debian), `Dockerfile.ubi` (Red Hat certified) — three image variants.
- `hack/`, `Makefile` — AppsCode build harness (everything runs inside `ghcr.io/appscode/golang-dev`).
- `vendor/` — checked-in deps.

API types come from `kubevault.dev/apimachinery`.

## Common commands

All Make targets run inside `ghcr.io/appscode/golang-dev` — Docker must be running.

- `make ci` — CI pipeline.
- `make build` / `make all-build` — build host or all-platform binaries.
- `make fmt`, `make lint`, `make unit-tests` / `make test` — standard.
- `make verify` — `verify-gen verify-modules`; `go mod tidy && go mod vendor` must leave the tree clean.
- `make container` — build PROD, DBG, and UBI images.
- `make push` — push all three; `make docker-manifest` writes multi-arch manifests; `make release` is the full publish flow.
- `make push-to-kind` / `make deploy-to-kind` — load into Kind and Helm-install.
- `make add-license` / `make check-license` — manage license headers.

Run a single Go test (requires a local Go toolchain):

```
go test ./pkg/worker/... -run TestName -v
```

## Conventions

- Module path is `kubevault.dev/unsealer` (vanity URL). Imports must use that.
- License: `LICENSE.md` (AppsCode). New files need the standard "Copyright AppsCode Inc. and Contributors" header (`make add-license`).
- Sign off commits (`git commit -s`); contributions follow the DCO.
- Vendor directory is checked in — `go mod tidy && go mod vendor` must leave the tree clean (enforced by `verify-modules`).
- **Adding a new key store backend**: drop a new directory under `pkg/kv/<name>/` implementing the interface declared in `pkg/kv/storage.go`. Don't branch on backend type inside `pkg/worker/`; the worker only sees the interface.
- All Vault HTTP interaction goes through `pkg/vault/`. Don't import the Vault SDK directly from `pkg/worker/` or `pkg/kv/`.
- The worker reconciles Vault state continuously — keep `pkg/worker/worker.go`'s loop idempotent. Init must only happen once (guarded by Vault's own `Initialized` check); unseal can happen many times.
- Three Dockerfiles, one binary — keep `Dockerfile.in`, `Dockerfile.dbg`, and `Dockerfile.ubi` in sync.
