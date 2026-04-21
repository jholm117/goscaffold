package markdown

import (
	"strings"
	"testing"
)

func TestPatchBadges_InsertWhenNone(t *testing.T) {
	content := "# myproject\n\nSome description.\n\n## Getting Started\n"
	badges := "[![CI](https://example.com/ci.svg)](https://example.com/ci)\n" +
		"[![Release](https://example.com/release.svg)](https://example.com/release)"

	result := PatchBadges(content, badges)

	if !strings.Contains(result, "[![CI]") {
		t.Error("badges should be inserted")
	}
	if !strings.Contains(result, "Some description.") {
		t.Error("description should be preserved")
	}
	ciIdx := strings.Index(result, "[![CI]")
	descIdx := strings.Index(result, "Some description.")
	if ciIdx > descIdx {
		t.Error("badges should appear before description")
	}
}

func TestPatchBadges_ReplaceExisting(t *testing.T) {
	content := "# myproject\n\n[![OldBadge](https://old.svg)](https://old)\n\nSome description.\n"
	badges := "[![CI](https://new.svg)](https://new)"

	result := PatchBadges(content, badges)

	if strings.Contains(result, "OldBadge") {
		t.Error("old badges should be removed")
	}
	if !strings.Contains(result, "[![CI]") {
		t.Error("new badges should be present")
	}
	if !strings.Contains(result, "Some description.") {
		t.Error("description should be preserved")
	}
}

func TestPatchBadges_PreserveContentBelow(t *testing.T) {
	content := "# myproject\n\n" +
		"[![Old](https://old.svg)](https://old)\n" +
		"[![Old2](https://old2.svg)](https://old2)\n\n" +
		"Description here.\n\n## Section\n\nMore content.\n"
	badges := "[![New](https://new.svg)](https://new)"

	result := PatchBadges(content, badges)

	if strings.Contains(result, "Old") {
		t.Error("old badges should be removed")
	}
	if !strings.Contains(result, "Description here.") {
		t.Error("description should be preserved")
	}
	if !strings.Contains(result, "## Section") {
		t.Error("sections should be preserved")
	}
}

func TestPatchBadges_NoTitle(t *testing.T) {
	content := "No title here.\n"
	badges := "[![CI](https://ci.svg)](https://ci)"

	result := PatchBadges(content, badges)

	if result != content {
		t.Error("should return content unchanged when no # title found")
	}
}
