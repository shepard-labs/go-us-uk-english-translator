package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/shepard-labs/go-us-uk-english-translator/internal/runner"
	"github.com/shepard-labs/go-us-uk-english-translator/translator"
)

func main() {
	var extFlag string
	var excludeFlag string

	american := flag.Bool("american", true, "Translate to American English (US). Mutually exclusive with --british.")
	british := flag.Bool("british", false, "Translate to British English (UK). Mutually exclusive with --american.")
	dryRun := flag.Bool("dry-run", false, "Print a unified diff to stdout; do not modify files.")
	flag.BoolVar(dryRun, "n", false, "Shorthand for --dry-run")
	recursive := flag.Bool("recursive", false, "Walk directories recursively.")
	flag.BoolVar(recursive, "r", false, "Shorthand for --recursive")
	flag.StringVar(&extFlag, "ext", ".ts,.tsx,.js,.jsx,.md,.txt,.yaml,.yml,.json,.html,.css", "Comma-separated file extensions to process when given a directory.")
	flag.StringVar(&excludeFlag, "exclude", "node_modules,.git,vendor,dist,build", "Comma-separated directory names to skip during recursive walk.")
	stdin := flag.Bool("stdin", false, "Read from stdin, write converted text to stdout.")
	summary := flag.Bool("summary", true, "After processing, print a table of file path and replacement count.")
	flag.BoolVar(summary, "s", true, "Shorthand for --summary")
	noCode := flag.Bool("no-code", true, "Skip replacement inside fenced code blocks, inline code spans, and import/require paths.")
	dictionary := flag.String("dictionary", "", "Path to an additional JSON dictionary file to merge.")
	flag.StringVar(dictionary, "d", "", "Shorthand for --dictionary")
	ignoreWords := flag.String("ignore-words", "", "Path to a plain-text file whose entries are excluded from replacement.")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s [flags] [file|dir ...]\n", os.Args[0])
		flag.PrintDefaults()
	}

	flag.Parse()

	var dir translator.Direction
	if *british && !*american {
		dir = translator.DirectionBritish
	} else if *american && !*british {
		dir = translator.DirectionAmerican
	} else if *american && *british {
		fmt.Fprintln(os.Stderr, "Error: --american and --british are mutually exclusive.")
		os.Exit(2)
	} else {
		// Both false, default to American per spec defaults
		dir = translator.DirectionAmerican
	}

	extensions := strings.Split(extFlag, ",")
	for i, e := range extensions {
		extensions[i] = strings.TrimSpace(e)
	}

	excludes := strings.Split(excludeFlag, ",")
	for i, e := range excludes {
		excludes[i] = strings.TrimSpace(e)
	}

	opts := runner.Options{
		Direction:       dir,
		DryRun:          *dryRun,
		Recursive:       *recursive,
		Extensions:      extensions,
		Exclude:         excludes,
		Stdin:           *stdin,
		Summary:         *summary,
		NoCode:          *noCode,
		DictionaryPath:  *dictionary,
		IgnoreWordsPath: *ignoreWords,
	}

	exitCode, err := runner.Run(opts, flag.Args())
	if err != nil {
		fmt.Fprintln(os.Stderr, "Fatal error:", err)
		os.Exit(2)
	}

	os.Exit(exitCode)
}
