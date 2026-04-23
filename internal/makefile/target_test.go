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
	if !strings.Contains(result, "go build -o bin/new ./cmd/new\n\n.PHONY: test") {
		t.Error("should have blank line between replaced target and next target")
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

func TestReplaceTarget_PrefixMatch(t *testing.T) {
	content := `.PHONY: e2e-up
e2e-up: ## Bootstrap kind cluster.
	./hack/e2e-up.sh

.PHONY: e2e-down
e2e-down: ## Delete kind cluster.
	./hack/e2e-down.sh

.PHONY: e2e
e2e: ## Run E2E tests.
	go test -tags e2e ./test/e2e/...

.PHONY: e2e-all
e2e-all: e2e-up e2e ## Run all E2E.
	echo done
`
	replacement := ".PHONY: e2e\ne2e: e2e-up ## Run E2E tests.\n" +
		"\tgo test ./test/e2e/ -v -tags e2e -timeout 10m\n"

	result := ReplaceTarget(content, "e2e", replacement)

	if !strings.Contains(result, "go test ./test/e2e/ -v -tags e2e -timeout 10m") {
		t.Error("e2e target should be replaced")
	}
	if !strings.Contains(result, "./hack/e2e-up.sh") {
		t.Error("e2e-up target should be preserved")
	}
	if !strings.Contains(result, "./hack/e2e-down.sh") {
		t.Error("e2e-down target should be preserved")
	}
	if !strings.Contains(result, "e2e-all") {
		t.Error("e2e-all target should be preserved")
	}
	if strings.Contains(result, "go test -tags e2e ./test/e2e/...") {
		t.Error("old e2e recipe should be gone")
	}
}

func TestHasTarget(t *testing.T) {
	content := ".PHONY: build\nbuild: ## Build.\n\tgo build\n\n.PHONY: e2e-up\ne2e-up:\n\techo up\n"

	if !HasTarget(content, "build") {
		t.Error("should find build")
	}
	if !HasTarget(content, "e2e-up") {
		t.Error("should find e2e-up")
	}
	if HasTarget(content, "e2e") {
		t.Error("should NOT find e2e (prefix of e2e-up)")
	}
	if HasTarget(content, "missing") {
		t.Error("should NOT find missing")
	}
}

func TestHasTarget_BareTarget(t *testing.T) {
	content := "build:\n\tgo build ./...\n\ntest:\n\tgo test ./...\n"

	if !HasTarget(content, "build") {
		t.Error("should find bare build target")
	}
	if !HasTarget(content, "test") {
		t.Error("should find bare test target")
	}
	if HasTarget(content, "testing") {
		t.Error("should NOT find testing (prefix of test line content)")
	}
}

func TestInsertTarget(t *testing.T) {
	content := `.PHONY: build
build: ## Build.
	go build ./...

.PHONY: clean
clean: ## Clean.
	rm -rf bin/

.PHONY: golangci-lint
golangci-lint: $(GOLANGCI_LINT)
`
	target := ".PHONY: tools\ntools: $(GOLANGCI_LINT) $(GOVULNCHECK) ## Install all tool dependencies.\n"

	result := InsertTarget(content, target, ".PHONY: golangci-lint")
	if !strings.Contains(result, ".PHONY: tools") {
		t.Error("tools target should be inserted")
	}
	toolsIdx := strings.Index(result, ".PHONY: tools")
	lintIdx := strings.Index(result, ".PHONY: golangci-lint")
	if toolsIdx > lintIdx {
		t.Error("tools should be inserted before golangci-lint")
	}
}

func TestInsertTarget_NoMarker(t *testing.T) {
	content := ".PHONY: build\nbuild:\n\techo build\n"
	target := ".PHONY: tools\ntools:\n\techo tools\n"

	result := InsertTarget(content, target, ".PHONY: golangci-lint")
	if !strings.Contains(result, ".PHONY: tools") {
		t.Error("should append target when marker not found")
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
	content := "$(GOLANGCI_LINT): $(LOCALBIN)\n\t$(call go-install-tool,old)\n\n" +
		"$(GOVULNCHECK): $(LOCALBIN)\n\t$(call go-install-tool,other)\n"

	replacement := "$(GOLANGCI_LINT): $(LOCALBIN)\n\t$(call go-install-tool,new)"

	result := ReplaceSpecialTarget(content, "$(GOLANGCI_LINT):", replacement)

	if !strings.Contains(result, "go-install-tool,new") {
		t.Error("special target should be replaced")
	}
	if !strings.Contains(result, "go-install-tool,other") {
		t.Error("other special target should be preserved")
	}
}
