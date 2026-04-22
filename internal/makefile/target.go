package makefile

import "strings"

// ReplaceVariable finds a variable line (NAME ?= or NAME =) and replaces it.
func ReplaceVariable(content, name, replacement string) string {
	lines := strings.Split(content, "\n")
	for i, line := range lines {
		if strings.HasPrefix(line, name+" ?=") || strings.HasPrefix(line, name+" =") {
			lines[i] = replacement
			return strings.Join(lines, "\n")
		}
	}
	return content
}

// ReplaceTarget finds a .PHONY: name block and replaces through end of recipe.
// Matches ".PHONY: name\n" exactly to avoid prefix collisions (e.g. "e2e" vs "e2e-up").
func ReplaceTarget(content, name, replacement string) string {
	marker := ".PHONY: " + name + "\n"
	start := strings.Index(content, marker)
	if start == -1 {
		return content
	}

	end := FindTargetEnd(content, start+len(marker)-1)
	return content[:start] + strings.TrimRight(replacement, "\n") + "\n" + content[end:]
}

// ReplaceDefine finds a define name ... endef block and replaces it.
func ReplaceDefine(content, name, replacement string) string {
	marker := "define " + name
	start := strings.Index(content, marker)
	if start == -1 {
		return content
	}

	endMarker := "endef"
	endIdx := strings.Index(content[start:], endMarker)
	if endIdx == -1 {
		return content
	}
	end := start + endIdx + len(endMarker)

	// consume trailing newline
	if end < len(content) && content[end] == '\n' {
		end++
	}

	return content[:start] + replacement + "\n" + content[end:]
}

// ReplaceSpecialTarget finds a target by prefix like $(LOCALBIN): and replaces it.
func ReplaceSpecialTarget(content, prefix, replacement string) string {
	start := strings.Index(content, prefix)
	if start == -1 {
		return content
	}

	end := FindTargetEnd(content, start+len(prefix))
	return content[:start] + strings.TrimRight(replacement, "\n") + "\n" + content[end:]
}

// FindTargetEnd finds where a target block ends, starting search after the given position.
// A target block ends at: next .PHONY:, next variable assignment, next ## comment,
// define, or a blank line not followed by a tab-indented line.
func FindTargetEnd(content string, after int) int {
	lines := strings.Split(content[after:], "\n")
	pos := after
	inRecipe := false

	for i, line := range lines {
		if i == 0 {
			// first line is the rest of the target/phony line itself
			pos += len(line) + 1
			continue
		}

		// tab-indented line = part of recipe
		if strings.HasPrefix(line, "\t") {
			inRecipe = true
			pos += len(line) + 1
			continue
		}

		// blank line while in recipe: include it if next line is tab-indented
		if inRecipe && line == "" {
			// look ahead
			if i+1 < len(lines) && strings.HasPrefix(lines[i+1], "\t") {
				pos += 1
				continue
			}
			// blank line after recipe with no more tab lines = end
			pos += 1
			break
		}

		if inRecipe {
			// non-tab, non-blank line after recipe = end
			break
		}

		// not yet in recipe
		if line == "" {
			pos += 1
			continue
		}

		// non-blank, non-tab line before recipe started = the target line itself
		pos += len(line) + 1
		inRecipe = true
	}

	return pos
}
