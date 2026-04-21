package markdown

import (
	"strings"
	"testing"
)

func TestReplaceSection(t *testing.T) {
	content := `# myproject

## Build & Test

- old build command
- old test command

## Architecture

Custom architecture docs here.

## Other

More custom stuff.
`

	newSection := `## Build & Test

- ` + "`make build`" + ` — build binary
- ` + "`make test`" + ` — run tests
- ` + "`make lint`" + ` — run golangci-lint
`

	result := ReplaceSection(content, "## Build & Test", newSection)

	if !strings.Contains(result, "make lint") {
		t.Error("new section content should be present")
	}
	if strings.Contains(result, "old build command") {
		t.Error("old section content should be gone")
	}
	if !strings.Contains(result, "Custom architecture docs here.") {
		t.Error("other sections should be preserved")
	}
	if !strings.Contains(result, "More custom stuff.") {
		t.Error("subsequent sections should be preserved")
	}
}

func TestReplaceSection_NotFound(t *testing.T) {
	content := "# myproject\n\n## Other\n\nstuff\n"
	result := ReplaceSection(content, "## Build & Test", "## Build & Test\n\nnew\n")
	if result != content {
		t.Error("should return content unchanged when section not found")
	}
}

func TestReplaceSection_LastSection(t *testing.T) {
	content := "# myproject\n\n## Build & Test\n\nold content\n"
	newSection := "## Build & Test\n\nnew content\n"

	result := ReplaceSection(content, "## Build & Test", newSection)

	if !strings.Contains(result, "new content") {
		t.Error("section should be replaced even when it's the last section")
	}
	if strings.Contains(result, "old content") {
		t.Error("old content should be gone")
	}
}
