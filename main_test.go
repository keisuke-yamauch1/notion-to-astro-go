package main

import (
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
