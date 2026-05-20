package runner

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/pmezard/go-difflib/difflib"
	"github.com/shepard-labs/go-us-uk-english-translator/translator"
)

type Options struct {
	Direction       translator.Direction
	DryRun          bool
	Recursive       bool
	Extensions      []string
	Exclude         []string
	Stdin           bool
	Summary         bool
	NoCode          bool
	DictionaryPath  string
	IgnoreWordsPath string // Note: Not directly supported by translator.ConverterOptions right now.
}

// Stats tracks file processing outcomes.
type Stats struct {
	FilesProcessed    int
	FilesChanged      int
	TotalReplacements int
	FileReplacements  map[string]int
}

// Run executes the translation process based on provided options and arguments.
func Run(opts Options, args []string) (int, error) {
	convOpts := translator.ConverterOptions{
		Direction:          opts.Direction,
		UserDictionaryPath: opts.DictionaryPath,
		SkipExclusionZones: opts.NoCode,
	}

	conv, err := translator.NewConverterWithOptions(convOpts)
	if err != nil {
		return 2, fmt.Errorf("failed to initialize converter: %w", err)
	}

	if opts.IgnoreWordsPath != "" {
		fmt.Fprintf(os.Stderr, "WARNING: --ignore-words is not currently supported by the core engine.\n")
	}

	stats := &Stats{
		FileReplacements: make(map[string]int),
	}

	if opts.Stdin || len(args) == 0 {
		changed, err := processStdin(conv, stats)
		if err != nil {
			return 2, err
		}
		if changed {
			return 1, nil
		}
		return 0, nil
	}

	// Process explicit files and directories
	for _, arg := range args {
		info, err := os.Stat(arg)
		if err != nil {
			return 2, fmt.Errorf("failed to stat %s: %w", arg, err)
		}

		if info.IsDir() {
			if !opts.Recursive {
				fmt.Fprintf(os.Stderr, "skipping directory %s (use --recursive to process)\n", arg)
				continue
			}
			err = filepath.WalkDir(arg, func(path string, d os.DirEntry, err error) error {
				if err != nil {
					return err
				}
				if d.IsDir() {
					for _, excl := range opts.Exclude {
						if d.Name() == excl {
							return filepath.SkipDir
						}
					}
					return nil
				}

				// Check file extension
				ext := filepath.Ext(path)
				validExt := false
				for _, e := range opts.Extensions {
					if ext == e {
						validExt = true
						break
					}
				}
				if !validExt {
					return nil
				}

				return processFile(path, conv, opts, stats)
			})
			if err != nil {
				return 2, err
			}
		} else {
			// Process explicit file regardless of extension
			if err := processFile(arg, conv, opts, stats); err != nil {
				return 2, err
			}
		}
	}

	if opts.Summary && !opts.Stdin && len(args) > 0 {
		printSummary(stats)
	}

	if stats.FilesChanged > 0 {
		return 1, nil
	}
	return 0, nil
}

func processStdin(conv *translator.Converter, stats *Stats) (bool, error) {
	input, err := io.ReadAll(os.Stdin)
	if err != nil {
		return false, fmt.Errorf("failed to read stdin: %w", err)
	}

	output, replacements := conv.Convert(string(input), "stdin")
	fmt.Print(output)

	stats.FilesProcessed++
	if replacements > 0 {
		stats.FilesChanged++
		stats.TotalReplacements += replacements
		return true, nil
	}
	return false, nil
}

func processFile(path string, conv *translator.Converter, opts Options, stats *Stats) error {
	content, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("failed to read %s: %w", path, err)
	}

	input := string(content)
	output, replacements := conv.Convert(input, path)

	stats.FilesProcessed++
	if replacements > 0 {
		stats.FilesChanged++
		stats.TotalReplacements += replacements
		stats.FileReplacements[path] = replacements

		if opts.DryRun {
			diff := difflib.UnifiedDiff{
				A:        difflib.SplitLines(input),
				B:        difflib.SplitLines(output),
				FromFile: path,
				ToFile:   path + " (converted)",
				Context:  3,
			}
			text, err := difflib.GetUnifiedDiffString(diff)
			if err != nil {
				return fmt.Errorf("failed to generate diff for %s: %w", path, err)
			}
			fmt.Print(text)
		} else {
			// Write file atomically
			dir := filepath.Dir(path)
			tmpFile, err := os.CreateTemp(dir, "engconv-*")
			if err != nil {
				return fmt.Errorf("failed to create temp file: %w", err)
			}
			tmpName := tmpFile.Name()

			// Cleanup the temp file if anything fails
			defer os.Remove(tmpName)

			if _, err := tmpFile.WriteString(output); err != nil {
				tmpFile.Close()
				return fmt.Errorf("failed to write to temp file: %w", err)
			}

			// Try to preserve permissions from original file
			info, err := os.Stat(path)
			if err == nil {
				tmpFile.Chmod(info.Mode())
			}

			if err := tmpFile.Close(); err != nil {
				return fmt.Errorf("failed to close temp file: %w", err)
			}

			if err := os.Rename(tmpName, path); err != nil {
				return fmt.Errorf("failed to atomically rename %s: %w", tmpName, err)
			}
		}
	}
	return nil
}

func printSummary(stats *Stats) {
	fmt.Fprintf(os.Stderr, "\n--- Summary ---\n")
	fmt.Fprintf(os.Stderr, "Files processed: %d\n", stats.FilesProcessed)
	fmt.Fprintf(os.Stderr, "Files changed:   %d\n", stats.FilesChanged)
	fmt.Fprintf(os.Stderr, "Replacements:    %d\n\n", stats.TotalReplacements)

	if stats.FilesChanged > 0 {
		fmt.Fprintf(os.Stderr, "Changed Files:\n")
		for path, count := range stats.FileReplacements {
			fmt.Fprintf(os.Stderr, "  %s (%d replacements)\n", path, count)
		}
	}
}
