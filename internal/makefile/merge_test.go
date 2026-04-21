package makefile

import (
	"strings"
	"testing"
)

func TestHasSection(t *testing.T) {
	content := "## Base\nfoo:\n\techo foo\n\n## CLI Targets\nbar:\n\techo bar\n"
	if !HasSection(content, "## CLI Targets") {
		t.Error("HasSection should find '## CLI Targets'")
	}
	if HasSection(content, "## Controller Targets") {
		t.Error("HasSection should not find '## Controller Targets'")
	}
}

func TestAppendSection(t *testing.T) {
	base := "## Base\nfoo:\n\techo foo\n"
	section := "\n## CLI Targets\nbar:\n\techo bar\n"
	result := AppendSection(base, section)
	if !strings.Contains(result, "## CLI Targets") {
		t.Error("result should contain CLI Targets section")
	}
	if !strings.Contains(result, "## Base") {
		t.Error("result should still contain Base section")
	}
}

func TestAppendSection_BeforeToolsBlock(t *testing.T) {
	base := "## Base\nfoo:\n\techo foo\n\n## Tools\n\nGOLANGCI_LINT = bin/golangci-lint\n"
	section := "\n## CLI Targets\nbar:\n\techo bar\n"
	result := AppendSection(base, section)
	cliIdx := strings.Index(result, "## CLI Targets")
	toolsIdx := strings.Index(result, "## Tools")
	if cliIdx > toolsIdx {
		t.Error("CLI Targets should appear before ## Tools block")
	}
}
