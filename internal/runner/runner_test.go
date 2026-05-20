package runner

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"testing"
)

func TestRun_Stdin(t *testing.T) {
	opts := Options{
		Direction: "american",
		Stdin:     true,
	}

	// Mock stdin
	oldStdin := os.Stdin
	r, w, _ := os.Pipe()
	os.Stdin = r

	// Mock stdout
	oldStdout := os.Stdout
	outR, outW, _ := os.Pipe()
	os.Stdout = outW

	w.Write([]byte("The colour is red."))
	w.Close()

	exitCode, err := Run(opts, []string{})

	outW.Close()
	os.Stdin = oldStdin
	os.Stdout = oldStdout

	if err != nil {
		t.Fatalf("run failed: %v", err)
	}
	if exitCode != 1 {
		t.Errorf("expected exit code 1 for changes on stdin, got %d", exitCode)
	}

	var buf bytes.Buffer
	io.Copy(&buf, outR)
	if buf.String() != "The color is red." {
		t.Errorf("expected 'The color is red.', got %q", buf.String())
	}
}

func TestPrintSummary(t *testing.T) {
	stats := &Stats{
		FilesProcessed:    5,
		FilesChanged:      2,
		TotalReplacements: 10,
		FileReplacements: map[string]int{
			"file1.ts": 7,
			"file2.md": 3,
		},
	}

	// Capture stderr
	oldStderr := os.Stderr
	r, w, _ := os.Pipe()
	os.Stderr = w

	printSummary(stats)

	w.Close()
	os.Stderr = oldStderr

	var buf bytes.Buffer
	io.Copy(&buf, r)

	output := buf.String()
	if !bytes.Contains([]byte(output), []byte("Files processed: 5")) {
		t.Errorf("missing processed count in summary: %s", output)
	}
	if !bytes.Contains([]byte(output), []byte("file1.ts (7 replacements)")) {
		t.Errorf("missing file info in summary: %s", output)
	}
}

func TestRun_Files(t *testing.T) {
	tempDir := t.TempDir()

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"mixed", "mixed.input.ts", "mixed.expected.ts"},
		{"markdown", "markdown.input.md", "markdown.expected.md"},
		{"html", "test.input.html", "test.expected.html"},
		{"yaml", "test.input.yaml", "test.expected.yaml"},
		{"json", "test.input.json", "test.expected.json"},
		{"js", "test.input.js", "test.expected.js"},
		{"txt", "test.input.txt", "test.expected.txt"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			inputPath := filepath.Join("../../testdata", tt.input)
			expectedPath := filepath.Join("../../testdata", tt.expected)

			inputBytes, err := os.ReadFile(inputPath)
			if err != nil {
				t.Fatalf("failed to read input file: %v", err)
			}
			expectedBytes, err := os.ReadFile(expectedPath)
			if err != nil {
				t.Fatalf("failed to read expected file: %v", err)
			}

			// Copy input file to temp dir so it can be modified
			tempInputPath := filepath.Join(tempDir, tt.input)
			if err := os.WriteFile(tempInputPath, inputBytes, 0644); err != nil {
				t.Fatalf("failed to write temp input: %v", err)
			}

			opts := Options{
				Direction:  "american",
				DryRun:     false,
				Recursive:  false,
				Stdin:      false,
				Summary:    false,
				NoCode:     true,
				Extensions: []string{".ts", ".md", ".html", ".yaml", ".json", ".js", ".txt"},
			}

			exitCode, err := Run(opts, []string{tempInputPath})
			if err != nil {
				t.Fatalf("run failed: %v", err)
			}
			if exitCode != 1 { // Should return 1 when changes are made
				t.Errorf("expected exit code 1 (changes made), got %d", exitCode)
			}

			resultBytes, err := os.ReadFile(tempInputPath)
			if err != nil {
				t.Fatalf("failed to read resulting file: %v", err)
			}

			if !bytes.Equal(resultBytes, expectedBytes) {
				t.Errorf("expected:\n%s\n\ngot:\n%s", string(expectedBytes), string(resultBytes))
			}

			// Test idempotency
			exitCode2, err := Run(opts, []string{tempInputPath})
			if err != nil {
				t.Fatalf("second run failed: %v", err)
			}
			if exitCode2 != 0 {
				t.Errorf("expected exit code 0 on second run (idempotency), got %d", exitCode2)
			}
		})
	}
}

func TestRun_DryRun(t *testing.T) {
	tempDir := t.TempDir()
	inputPath := filepath.Join("../../testdata", "mixed.input.ts")

	inputBytes, err := os.ReadFile(inputPath)
	if err != nil {
		t.Fatalf("failed to read input file: %v", err)
	}

	tempInputPath := filepath.Join(tempDir, "mixed.input.ts")
	if err := os.WriteFile(tempInputPath, inputBytes, 0644); err != nil {
		t.Fatalf("failed to write temp input: %v", err)
	}

	opts := Options{
		Direction:  "american",
		DryRun:     true,
		Recursive:  false,
		Stdin:      false,
		Summary:    false,
		NoCode:     true,
		Extensions: []string{".ts"},
	}

	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	exitCode, err := Run(opts, []string{tempInputPath})

	w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	io.Copy(&buf, r)

	if err != nil {
		t.Fatalf("run failed: %v", err)
	}
	if exitCode != 1 {
		t.Errorf("expected exit code 1 for changes in dry-run, got %d", exitCode)
	}

	if buf.Len() == 0 {
		t.Errorf("expected diff output on stdout during dry-run")
	}

	// Verify file was NOT modified
	resultBytes, _ := os.ReadFile(tempInputPath)
	if !bytes.Equal(resultBytes, inputBytes) {
		t.Errorf("expected file to be unmodified in dry-run")
	}
}

func TestRun_Recursive(t *testing.T) {
	tempDir := t.TempDir()

	// Create nested structure
	nestedDir := filepath.Join(tempDir, "src", "components")
	if err := os.MkdirAll(nestedDir, 0755); err != nil {
		t.Fatalf("failed to create dirs: %v", err)
	}

	// Create excluded directory
	nodeModulesDir := filepath.Join(tempDir, "node_modules")
	if err := os.MkdirAll(nodeModulesDir, 0755); err != nil {
		t.Fatalf("failed to create node_modules: %v", err)
	}

	file1 := filepath.Join(nestedDir, "test1.ts")
	file2 := filepath.Join(nodeModulesDir, "test2.ts")

	os.WriteFile(file1, []byte("colour is red"), 0644)
	os.WriteFile(file2, []byte("colour is red"), 0644)

	opts := Options{
		Direction:  "american",
		DryRun:     false,
		Recursive:  true,
		Stdin:      false,
		Summary:    false,
		NoCode:     true,
		Extensions: []string{".ts"},
		Exclude:    []string{"node_modules"},
	}

	exitCode, err := Run(opts, []string{tempDir})
	if err != nil {
		t.Fatalf("run failed: %v", err)
	}
	if exitCode != 1 {
		t.Errorf("expected exit code 1, got %d", exitCode)
	}

	res1, _ := os.ReadFile(file1)
	if string(res1) != "color is red" {
		t.Errorf("expected file1 to be converted")
	}

	res2, _ := os.ReadFile(file2)
	if string(res2) != "colour is red" {
		t.Errorf("expected file2 to be skipped (excluded directory)")
	}
}
