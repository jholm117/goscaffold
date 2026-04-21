package makefile

import (
	"strings"
	"testing"
)

func TestReplaceVariable(t *testing.T) {
	content := "PROJECT_NAME ?= oldname\nLOCALBIN ?= $(shell pwd)/bin\nFOO ?= bar\n"

	result := ReplaceVariable(content, "PROJECT_NAME", "PROJECT_NAME ?= newname")
	if result != "PROJECT_NAME ?= newname\nLOCALBIN ?= $(shell pwd)/bin\nFOO ?= bar\n" {
		t.Errorf("ReplaceVariable wrong:\n%s", result)
	}
}

func TestReplaceVariable_NotFound(t *testing.T) {
	content := "FOO ?= bar\n"
	result := ReplaceVariable(content, "MISSING", "MISSING ?= val")
	if result != content {
		t.Error("ReplaceVariable should return content unchanged when variable not found")
	}
}

func TestReplaceVariable_EqualsSign(t *testing.T) {
	content := "GOLANGCI_LINT = $(LOCALBIN)/golangci-lint\nGOVULNCHECK = $(LOCALBIN)/govulncheck\n"
	result := ReplaceVariable(content, "GOLANGCI_LINT", "GOLANGCI_LINT = $(LOCALBIN)/new-lint")
	if result != "GOLANGCI_LINT = $(LOCALBIN)/new-lint\nGOVULNCHECK = $(LOCALBIN)/govulncheck\n" {
		t.Errorf("ReplaceVariable with = wrong:\n%s", result)
	}
}

func TestReplaceTarget(t *testing.T) {
	content := `.PHONY: build
build: ## Build binary.
	go build -o bin/old ./cmd/old

.PHONY: test
test: ## Run tests.
	go test ./...

custom-target:
	echo custom
`
	replacement := `.PHONY: build
build: ## Build binary.
	go build -o bin/new ./cmd/new
`
	result := ReplaceTarget(content, "build", replacement)

	if !strings.Contains(result, "go build -o bin/new ./cmd/new") {
		t.Error("build target should be replaced")
	}
	if !strings.Contains(result, "go test ./...") {
		t.Error("test target should be preserved")
	}
	if !strings.Contains(result, "echo custom") {
		t.Error("custom target should be preserved")
	}
}

func TestReplaceTarget_NotFound(t *testing.T) {
	content := ".PHONY: build\nbuild:\n\techo build\n"
	result := ReplaceTarget(content, "missing", ".PHONY: missing\nmissing:\n\techo hi\n")
	if result != content {
		t.Error("ReplaceTarget should return content unchanged when target not found")
	}
}

func TestReplaceTarget_CustomBetween(t *testing.T) {
	content := `.PHONY: build
build: ## Build.
	go build ./...

my-custom:
	echo custom

.PHONY: test
test: ## Test.
	go test ./...
`
	replacement := `.PHONY: build
build: ## Build.
	go build -o bin/app ./cmd/app
`
	result := ReplaceTarget(content, "build", replacement)

	if !strings.Contains(result, "go build -o bin/app ./cmd/app") {
		t.Error("build target should be replaced")
	}
	if !strings.Contains(result, "echo custom") {
		t.Error("custom target between managed targets should survive")
	}
}

func TestReplaceDefine(t *testing.T) {
	content := "stuff before\n\ndefine go-install-tool\nold content\nold line 2\nendef\n\nstuff after\n"
	replacement := "define go-install-tool\nnew content\nendef"

	result := ReplaceDefine(content, "go-install-tool", replacement)

	if !strings.Contains(result, "new content") {
		t.Error("define block should be replaced")
	}
	if !strings.Contains(result, "stuff before") {
		t.Error("content before should be preserved")
	}
	if !strings.Contains(result, "stuff after") {
		t.Error("content after should be preserved")
	}
	if strings.Contains(result, "old content") {
		t.Error("old define content should be gone")
	}
}

func TestReplaceSpecialTarget(t *testing.T) {
	content := "$(GOLANGCI_LINT): $(LOCALBIN)\n\t$(call go-install-tool,old)\n\n$(GOVULNCHECK): $(LOCALBIN)\n\t$(call go-install-tool,other)\n"

	replacement := "$(GOLANGCI_LINT): $(LOCALBIN)\n\t$(call go-install-tool,new)"

	result := ReplaceSpecialTarget(content, "$(GOLANGCI_LINT):", replacement)

	if !strings.Contains(result, "go-install-tool,new") {
		t.Error("special target should be replaced")
	}
	if !strings.Contains(result, "go-install-tool,other") {
		t.Error("other special target should be preserved")
	}
}
