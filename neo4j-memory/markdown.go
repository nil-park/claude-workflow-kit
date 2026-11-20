package main

import (
	"fmt"
	"regexp"
	"strings"

	"golang.org/x/text/unicode/norm"
)

// VariantSection is one `## name` section of a variant file: the placeholder
// name and the content of the code block under it.
type VariantSection struct {
	Name  string
	Value string
}

var fenceOpen = regexp.MustCompile("^(`{3,})[A-Za-z]*$")

// normalizeText unifies line endings and Hangul composition so that text read
// on any OS compares equal.
func normalizeText(text string) string {
	return norm.NFC.String(strings.ReplaceAll(text, "\r\n", "\n"))
}

func stripFrontmatter(text string) string {
	if !strings.HasPrefix(text, "---\n") {
		return text
	}
	end := strings.Index(text[len("---\n"):], "\n---\n")
	if end < 0 {
		return text
	}
	return strings.TrimLeft(text[len("---\n")+end+len("\n---\n"):], "\n")
}

// parseVariant reads a multi-session variant file. Headings inside a code
// block belong to the value, so only headings outside one start a section.
func parseVariant(text string) ([]VariantSection, error) {
	var sections []VariantSection
	lines := strings.Split(text, "\n")
	for i := 0; i < len(lines); i++ {
		name, ok := strings.CutPrefix(lines[i], "## ")
		if !ok {
			continue
		}
		name = strings.TrimSpace(name)
		open := -1
		for j := i + 1; j < len(lines) && open < 0; j++ {
			switch {
			case fenceOpen.MatchString(lines[j]):
				open = j
			case strings.HasPrefix(lines[j], "## "):
				return nil, fmt.Errorf("section %q has no code block", name)
			}
		}
		if open < 0 {
			return nil, fmt.Errorf("section %q has no code block", name)
		}
		fence := fenceOpen.FindStringSubmatch(lines[open])[1]
		closing := -1
		for j := open + 1; j < len(lines); j++ {
			if strings.TrimSpace(lines[j]) == fence {
				closing = j
				break
			}
		}
		if closing < 0 {
			return nil, fmt.Errorf("section %q has an unclosed code block", name)
		}
		sections = append(sections, VariantSection{Name: name, Value: strings.Join(lines[open+1:closing], "\n")})
		i = closing
	}
	return sections, nil
}
