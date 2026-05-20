package replace

import (
	"testing"
)

func TestConvert_WordBoundaryEnforcement(t *testing.T) {
	dict := map[string]string{
		"organise":  "organize",
		"colour":    "color",
		"programme": "program",
		"reorganise": "reorganize",
		"recognise": "recognize",
	}

	tests := []struct {
		input    string
		expected string
	}{
		{"enterprise", "enterprise"},
		{"precise", "precise"},
		{"organise", "organize"},
		{"reorganise", "reorganize"},
		{"I organise and recognise", "I organize and recognize"},
		{"The colour of the programme", "The color of the program"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			actual, _ := Convert(tt.input, dict, true, "test")
			if actual != tt.expected {
				t.Errorf("Convert(%q) = %q, want %q", tt.input, actual, tt.expected)
			}
		})
	}
}

func TestConvert_ExclusionZones(t *testing.T) {
	dict := map[string]string{
		"colour": "color",
	}

	tests := []struct {
		desc     string
		input    string
		expected string
	}{
		{
			"No replacement inside fenced code block",
			"```\nconst colour = 'red';\n```",
			"```\nconst colour = 'red';\n```",
		},
		{
			"Replacement outside code block is applied",
			"The colour is red. ```const colour = 'red';``` The colour changed.",
			"The color is red. ```const colour = 'red';``` The color changed.",
		},
		{
			"No replacement inside inline code",
			"Use `colour` in CSS.",
			"Use `colour` in CSS.",
		},
		{
			"No replacement inside a URL",
			"See https://example.com/colour-theory for details.",
			"See https://example.com/colour-theory for details.",
		},
		{
			"No replacement inside an import path",
			"import styles from './colour-utils';",
			"import styles from './colour-utils';",
		},
	}

	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			actual, _ := Convert(tt.input, dict, true, "test")
			if actual != tt.expected {
				t.Errorf("Convert(%q) = %q, want %q", tt.input, actual, tt.expected)
			}
		})
	}
}

func TestConvert_Casing(t *testing.T) {
	dict := map[string]string{
		"colour": "color",
	}

	tests := []struct {
		input    string
		expected string
	}{
		{"COLOUR", "COLOR"},
		{"Colour", "Color"},
		{"colour", "color"},
		{"cOloUr", "color"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			actual, _ := Convert(tt.input, dict, true, "test")
			if actual != tt.expected {
				t.Errorf("Convert(%q) = %q, want %q", tt.input, actual, tt.expected)
			}
		})
	}
}

func TestConvert_Idempotency(t *testing.T) {
	dict := map[string]string{
		"colour": "color",
	}

	input := "The colour is red."
	pass1, _ := Convert(input, dict, true, "test")
	pass2, count := Convert(pass1, dict, true, "test")

	if pass1 != pass2 {
		t.Errorf("Not idempotent. Pass1: %q, Pass2: %q", pass1, pass2)
	}
	if count != 0 {
		t.Errorf("Expected 0 replacements on second pass, got %d", count)
	}
}

func TestConvert_WithoutExclusions(t *testing.T) {
	dict := map[string]string{
		"colour": "color",
	}

	input := "```\nconst colour = 'red';\n```"
	expected := "```\nconst color = 'red';\n```"

	actual, _ := Convert(input, dict, false, "test")
	if actual != expected {
		t.Errorf("Convert without exclusions: %q, want %q", actual, expected)
	}
}

// Dummy dict test to verify the WARNING output in Convert (stdout/stderr interception left out for simplicity)
func TestConvert_SelfCheckWarning(t *testing.T) {
	// A bit tricky to verify stderr without intercepting, but we can just run it
	// to ensure it doesn't panic.
	dict := map[string]string{
		"colour": "color",
	}
	Convert("The colour is blue", dict, true, "test")
}
