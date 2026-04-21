package markdown

import "strings"

// ReplaceSection finds a markdown header (e.g. "## Build & Test") and replaces
// everything from that header to the next same-level header (## ).
// If the header is the last section, replaces to end of content.
// If the header is not found, returns content unchanged.
func ReplaceSection(content, header, replacement string) string {
	start := strings.Index(content, header)
	if start == -1 {
		return content
	}

	afterHeader := start + len(header)
	rest := content[afterHeader:]

	end := -1
	lines := strings.Split(rest, "\n")
	pos := 0
	for i, line := range lines {
		if i == 0 {
			pos += len(line) + 1
			continue
		}
		if strings.HasPrefix(line, "## ") {
			end = afterHeader + pos
			break
		}
		pos += len(line) + 1
	}

	if end == -1 {
		return content[:start] + strings.TrimRight(replacement, "\n") + "\n"
	}

	return content[:start] + strings.TrimRight(replacement, "\n") + "\n\n" + content[end:]
}
