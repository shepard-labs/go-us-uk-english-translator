package dict

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestLoadDictionary(t *testing.T) {
	t.Run("American direction", func(t *testing.T) {
		d, err := LoadDictionary(DirectionAmerican, "")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(d) == 0 {
			t.Fatal("dictionary is empty")
		}
		if d["colour"] != "color" {
			t.Errorf("expected 'colour' to map to 'color', got '%s'", d["colour"])
		}
	})

	t.Run("British direction", func(t *testing.T) {
		d, err := LoadDictionary(DirectionBritish, "")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(d) == 0 {
			t.Fatal("dictionary is empty")
		}
		if d["color"] != "colour" {
			t.Errorf("expected 'color' to map to 'colour', got '%s'", d["color"])
		}
	})

	t.Run("User dictionary merge", func(t *testing.T) {
		tempDir := t.TempDir()
		userDictPath := filepath.Join(tempDir, "user_dict.json")
		userDict := map[string]string{
			"customuk": "customus",
			"colour":   "customcolor", // Override existing
		}
		userData, _ := json.Marshal(userDict)
		if err := os.WriteFile(userDictPath, userData, 0644); err != nil {
			t.Fatalf("failed to write user dict: %v", err)
		}

		d, err := LoadDictionary(DirectionAmerican, userDictPath)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if d["customuk"] != "customus" {
			t.Errorf("expected user dict entry 'customuk' -> 'customus'")
		}
		if d["colour"] != "customcolor" {
			t.Errorf("expected overridden entry 'colour' -> 'customcolor', got '%s'", d["colour"])
		}
	})

	t.Run("Unknown direction", func(t *testing.T) {
		_, err := LoadDictionary("invalid", "")
		if err == nil {
			t.Fatal("expected error for unknown direction, got nil")
		}
	})
}
