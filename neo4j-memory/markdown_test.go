package main

import (
	"slices"
	"strings"
	"testing"
)

func TestParseVariantKeepsHeadingsInsideLongerFence(t *testing.T) {
	text := strings.Join([]string{
		"# agy", "", "## peer-cli", "", "```text", "Antigravity CLI", "```", "",
		"## peer-setup", "", "````text", "## 사전 점검", "", "```bash", "agy --version", "```", "````", "",
		"## empty", "", "```text", "```",
	}, "\n")
	got, err := parseVariant(text)
	if err != nil {
		t.Fatal(err)
	}
	want := []VariantSection{
		{Name: "peer-cli", Value: "Antigravity CLI"},
		{Name: "peer-setup", Value: "## 사전 점검\n\n```bash\nagy --version\n```"},
		{Name: "empty", Value: ""},
	}
	if !slices.Equal(got, want) {
		t.Fatalf("got %q\nwant %q", got, want)
	}
}

func TestParseVariantRejectsMalformedSections(t *testing.T) {
	for name, text := range map[string]string{
		"no block":       "## a\ntext\n## b\n```\nx\n```",
		"unclosed block": "## a\n```text\nx",
	} {
		if _, err := parseVariant(text); err == nil {
			t.Errorf("%s: expected an error", name)
		}
	}
}

func TestStripFrontmatter(t *testing.T) {
	if got := stripFrontmatter("---\nname: x\n---\n\nbody\n"); got != "body\n" {
		t.Errorf("got %q", got)
	}
	if got := stripFrontmatter("---\nunterminated\nbody"); got != "---\nunterminated\nbody" {
		t.Errorf("got %q", got)
	}
}

func TestNormalizeTextComposesHangul(t *testing.T) {
	decomposed := "가\r\n"
	if got := normalizeText(decomposed); got != "가\n" {
		t.Errorf("got %q", got)
	}
}
