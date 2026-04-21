# /tmp/goscaffold-smoke-test

## Build & Test

- `make build` — build binary to `bin//tmp/goscaffold-smoke-test`
- `make test` — run all tests
- `make lint` — run golangci-lint
- `make setup-hooks` — install pre-push hook

## Project Layout

- `cmd//tmp/goscaffold-smoke-test/` — CLI entry point (cobra)
- `internal/` — private packages
- `pkg/` — public packages
- `hack/ci-checks.sh` — shared CI checks (hook + CI parity)
