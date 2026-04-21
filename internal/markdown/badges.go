package markdown

import "strings"

// PatchBadges replaces or inserts badge lines after the # title in markdown content.
// Badge lines are consecutive lines starting with "[![".
// If no # title is found, content is returned unchanged.
func PatchBadges(content, badges string) string {
	lines := strings.Split(content, "\n")

	titleIdx := -1
	for i, line := range lines {
		if strings.HasPrefix(line, "# ") {
			titleIdx = i
			break
		}
	}
	if titleIdx == -1 {
		return content
	}

	badgeStart := -1
	badgeEnd := -1
	for i := titleIdx + 1; i < len(lines); i++ {
		if lines[i] == "" {
			continue
		}
		if strings.HasPrefix(lines[i], "[![") {
			if badgeStart == -1 {
				badgeStart = i
			}
			badgeEnd = i + 1
		} else {
			break
		}
	}

	badgeLines := strings.Split(badges, "\n")

	if badgeStart != -1 {
		result := make([]string, 0, len(lines))
		result = append(result, lines[:badgeStart]...)
		result = append(result, badgeLines...)
		result = append(result, lines[badgeEnd:]...)
		return strings.Join(result, "\n")
	}

	result := make([]string, 0, len(lines)+len(badgeLines)+1)
	result = append(result, lines[titleIdx])
	result = append(result, "")
	result = append(result, badgeLines...)
	result = append(result, lines[titleIdx+1:]...)
	return strings.Join(result, "\n")
}
