PROJECT_NAME ?= goscaffold
LOCALBIN ?= $(shell pwd)/bin

## Tool Versions
GOLANGCI_LINT_VERSION ?= v2.8.0
GOVULNCHECK_VERSION ?= v1.1.4
GORELEASER_VERSION ?= v2.8.2

.PHONY: build
build: ## Build binary.
	go build -o bin/$(PROJECT_NAME) ./cmd/$(PROJECT_NAME)

.PHONY: install
install: ## Install to $GOPATH/bin.
	go install ./cmd/$(PROJECT_NAME)

.PHONY: test
test: ## Run tests.
	go test ./... -coverprofile cover.out

.PHONY: lint
lint: golangci-lint ## Run linter.
	"$(GOLANGCI_LINT)" run

.PHONY: lint-fix
lint-fix: golangci-lint ## Run linter with auto-fixes.
	"$(GOLANGCI_LINT)" run --fix

.PHONY: lint-config
lint-config: golangci-lint ## Verify linter config.
	"$(GOLANGCI_LINT)" config verify

.PHONY: fmt
fmt: ## Run go fmt.
	go fmt ./...

.PHONY: vet
vet: ## Run go vet.
	go vet ./...

.PHONY: govulncheck
govulncheck: $(GOVULNCHECK) ## Run vulnerability check.
	"$(GOVULNCHECK)" ./...

.PHONY: release-snapshot
release-snapshot: goreleaser ## Build release snapshot (no publish).
	"$(GORELEASER)" release --snapshot --clean

.PHONY: setup-hooks
setup-hooks: ## Install git hooks.
	@git config core.hooksPath .githooks
	@echo "Installed hooks via core.hooksPath -> .githooks/"

.PHONY: clean
clean: ## Remove build artifacts.
	rm -rf bin/ dist/ cover.out

.PHONY: help
help: ## Show this help.
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-20s\033[0m %s\n", $$1, $$2}'

## Tools

GOLANGCI_LINT = $(LOCALBIN)/golangci-lint
GOVULNCHECK = $(LOCALBIN)/govulncheck
GORELEASER = $(LOCALBIN)/goreleaser

.PHONY: golangci-lint
golangci-lint: $(GOLANGCI_LINT)
$(GOLANGCI_LINT): $(LOCALBIN)
	$(call go-install-tool,$(GOLANGCI_LINT),github.com/golangci/golangci-lint/v2/cmd/golangci-lint,$(GOLANGCI_LINT_VERSION))

$(GOVULNCHECK): $(LOCALBIN)
	$(call go-install-tool,$(GOVULNCHECK),golang.org/x/vuln/cmd/govulncheck,$(GOVULNCHECK_VERSION))

.PHONY: goreleaser
goreleaser: $(GORELEASER)
$(GORELEASER): $(LOCALBIN)
	$(call go-install-tool,$(GORELEASER),github.com/goreleaser/goreleaser/v2,$(GORELEASER_VERSION))

$(LOCALBIN):
	mkdir -p "$(LOCALBIN)"

define go-install-tool
@[ -f "$(1)-$(3)" ] && [ "$$(readlink -- "$(1)" 2>/dev/null)" = "$(1)-$(3)" ] || { \
set -e; \
package=$(2)@$(3) ;\
echo "Downloading $${package}" ;\
rm -f "$(1)" ;\
GOOS=$(shell go env GOHOSTOS) GOARCH=$(shell go env GOHOSTARCH) GOBIN="$(LOCALBIN)" go install $${package} ;\
mv "$(LOCALBIN)/$$(basename "$(1)")" "$(1)-$(3)" ;\
} ;\
ln -sf "$$(realpath "$(1)-$(3)")" "$(1)"
endef
