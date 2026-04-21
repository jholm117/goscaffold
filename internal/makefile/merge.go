package makefile

import "strings"

func HasSection(content, sentinel string) bool {
	return strings.Contains(content, sentinel)
}

func AppendSection(content, section string) string {
	toolsMarker := "## Tools"
	if idx := strings.Index(content, toolsMarker); idx != -1 {
		return content[:idx] + section + "\n" + content[idx:]
	}
	return content + section
}
