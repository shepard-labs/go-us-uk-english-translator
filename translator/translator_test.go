package translator

import (
	"testing"
)

func TestNewConverter(t *testing.T) {
	t.Run("American direction default", func(t *testing.T) {
		conv, err := NewConverter(DirectionAmerican)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !conv.skipExclusionZones {
			t.Error("expected skipExclusionZones to be true by default")
		}
		
		out, count := conv.Convert("colour is nice", "test")
		if out != "color is nice" {
			t.Errorf("expected 'color is nice', got '%s'", out)
		}
		if count != 1 {
			t.Errorf("expected count 1, got %d", count)
		}
	})

	t.Run("British direction default", func(t *testing.T) {
		conv, err := NewConverter(DirectionBritish)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		
		out, count := conv.Convert("color is nice", "test")
		if out != "colour is nice" {
			t.Errorf("expected 'colour is nice', got '%s'", out)
		}
		if count != 1 {
			t.Errorf("expected count 1, got %d", count)
		}
	})

	t.Run("Options with no-code=false", func(t *testing.T) {
		conv, err := NewConverterWithOptions(ConverterOptions{
			Direction:          DirectionAmerican,
			SkipExclusionZones: false,
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		
		input := "```colour```"
		out, count := conv.Convert(input, "test")
		// Should replace inside code block because SkipExclusionZones is false
		if out != "```color```" {
			t.Errorf("expected replacement inside code block when skip is false, got '%s'", out)
		}
		if count != 1 {
			t.Errorf("expected count 1, got %d", count)
		}
	})
}
