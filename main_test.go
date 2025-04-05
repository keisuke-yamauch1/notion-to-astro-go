package main

import (
	"regexp"
	"strings"
	"testing"
)

func TestProcessEmptyLines(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name: "Single empty line between paragraphs",
			input: `---
title: Test
---

First paragraph.

Second paragraph.

Third paragraph.`,
			expected: `---
title: Test
---

First paragraph.
Second paragraph.
Third paragraph.`,
		},
		{
			name: "Multiple empty lines between paragraphs",
			input: `---
title: Test
---

First paragraph.


Second paragraph.



Third paragraph.`,
			expected: `---
title: Test
---

First paragraph.

Second paragraph.

Third paragraph.`,
		},
		{
			name: "Mixed single and multiple empty lines",
			input: `---
title: Test
---

First paragraph.

Second paragraph.


Third paragraph.



Fourth paragraph.

Fifth paragraph.`,
			expected: `---
title: Test
---

First paragraph.
Second paragraph.

Third paragraph.

Fourth paragraph.
Fifth paragraph.`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := processEmptyLines(tt.input)
			if result != tt.expected {
				t.Errorf("processEmptyLines() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestBlogDescriptionGeneration(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		expected string
	}{
		{
			name:     "Short content without newlines",
			content:  "This is a short blog post content.",
			expected: "This is a short blog post content.",
		},
		{
			name:     "Content with newlines",
			content:  "This is a blog post\nwith newlines\nin the content.",
			expected: "This is a blog post with newlines in the content.",
		},
		{
			name:     "Long content with newlines",
			content:  "This is a very long blog post content that exceeds 70 characters\nand has newlines\nin it. The description should be limited to 70 characters and newlines should be converted to spaces.",
			expected: "This is a very long blog post content that exceeds 70 characters and h",
		},
		{
			name:     "Content with multiple consecutive spaces",
			content:  "This is a blog post  with   multiple    consecutive     spaces.",
			expected: "This is a blog post with multiple consecutive spaces.",
		},
		{
			name:     "Content with multiple newlines",
			content:  "This is a blog post\n\nwith\n\n\nmultiple\n\nnewlines.",
			expected: "This is a blog post with multiple newlines.",
		},
		{
			name:     "Short Japanese content",
			content:  "これは短い日本語のブログ記事です。",
			expected: "これは短い日本語のブログ記事です。",
		},
		{
			name:     "Japanese content with newlines",
			content:  "これは日本語の\nブログ記事\nです。",
			expected: "これは日本語の ブログ記事 です。",
		},
		{
			name:     "Long Japanese content",
			content:  "これは70文字を超える長い日本語のブログ記事です。日本語は1文字が複数バイトで表現されるため、バイト数ではなく文字数でカウントする必要があります。このテストでは、70文字を超える部分が正しく切り取られることを確認します。",
			expected: "これは70文字を超える長い日本語のブログ記事です。日本語は1文字が複数バイトで表現されるため、バイト数ではなく文字数でカウントする必要があり",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Replace newlines with spaces
			descriptionText := strings.ReplaceAll(tt.content, "\n", " ")
			// Remove extra spaces
			descriptionText = regexp.MustCompile(`\s+`).ReplaceAllString(descriptionText, " ")
			// Trim spaces
			descriptionText = strings.TrimSpace(descriptionText)
			// Get first 70 characters or less if content is shorter
			// Use runes to correctly handle multi-byte characters like Japanese
			runes := []rune(descriptionText)
			if len(runes) > 70 {
				descriptionText = string(runes[:70])
			}

			if descriptionText != tt.expected {
				t.Errorf("Blog description generation failed. Got: %q, Want: %q", descriptionText, tt.expected)
			}
		})
	}
}

func TestConvertMarkdownLinksToPlainText(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "No markdown links",
			input:    "This is a text without markdown links.",
			expected: "This is a text without markdown links.",
		},
		{
			name:     "Single markdown link",
			input:    "[aaa](https://www.kechiiiiin.com/)は〇〇だ",
			expected: "aaaは〇〇だ",
		},
		{
			name:     "Multiple markdown links",
			input:    "[aaa](https://www.kechiiiiin.com/)は[bbb](https://example.com)だ",
			expected: "aaaはbbbだ",
		},
		{
			name:     "Markdown link with Japanese text",
			input:    "[日本語](https://example.jp/)のテキスト",
			expected: "日本語のテキスト",
		},
		{
			name:     "Text with brackets but not a markdown link",
			input:    "This [is] not a markdown link.",
			expected: "This [is] not a markdown link.",
		},
		{
			name:     "Text with parentheses but not a markdown link",
			input:    "This (is) not a markdown link.",
			expected: "This (is) not a markdown link.",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := convertMarkdownLinksToPlainText(tt.input)
			if result != tt.expected {
				t.Errorf("convertMarkdownLinksToPlainText() = %v, want %v", result, tt.expected)
			}
		})
	}
}
