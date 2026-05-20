# go-us-uk-english-translator

A Go library and CLI tool that converts British English spellings to American English spellings (and vice-versa) in text files, stdin, or directory trees. Designed for **zero false positives**: every substitution is guarded by whole-word boundary matching, casing is preserved automatically, and content inside code fences, URLs, and import paths is never touched.

All operations are dictionary-driven using an embedded JSON dictionary, ensuring idempotent and highly accurate conversions.

[![CI](https://github.com/shepard-labs/go-us-uk-english-translator/actions/workflows/ci.yml/badge.svg?branch=main)](https://github.com/shepard-labs/go-us-uk-english-translator/actions/workflows/ci.yml)
[![Go Reference](https://pkg.go.dev/badge/github.com/shepard-labs/go-us-uk-english-translator.svg)](https://pkg.go.dev/github.com/shepard-labs/go-us-uk-english-translator)
[![Go Report Card](https://goreportcard.com/badge/github.com/shepard-labs/go-us-uk-english-translator)](https://goreportcard.com/report/github.com/shepard-labs/go-us-uk-english-translator)
[![license](https://img.shields.io/badge/license-%20%20GNU%20GPLv3%20-green?style=plastic)](https://img.shields.io/badge/license-%20%20GNU%20GPLv3%20-green?style=plastic)
[![Ask DeepWiki](https://deepwiki.com/badge.svg)](https://deepwiki.com/shepard-labs/go-us-uk-english-translator)

---

## Features

| Feature                        | Default   | Description                                                                                            |
|--------------------------------|-----------|--------------------------------------------------------------------------------------------------------|
| **Zero False Positives**       | Always on | Uses Unicode word boundaries. Substrings like `ise` in `precise` or `enterprise` are never touched.    |
| **Case Preservation**          | Always on | Automatically matches UPPERCASE, Title Case, or lowercase token formats during replacement.            |
| **Exclusion Zones**            | Always on | Skips replacement inside fenced code blocks (`` ` ``), URLs, HTML attributes, and import paths.        |
| **Bi-directional Translation** | American  | Supports translating UK → US (`--american`) or US → UK (`--british`).                                  |
| **Idempotent**                 | Always on | Running the tool twice on the same file produces no further changes.                                   |
| **Dry-run Mode**               | Opt-in    | Prints a unified diff to stdout instead of modifying files on disk (`--dry-run`).                      |
| **Custom Dictionaries**        | Opt-in    | Provide your own dictionary JSON or an `--ignore-words` list to override the defaults.                 |
| **Self-Hosted API Server**     | Included  | Ready-to-deploy HTTP server at `cmd/apiserver`.                                                        |

---

## Installation

### Library

```bash
go get github.com/shepard-labs/go-us-uk-english-translator
```

### CLI Tool & API Server

```bash
go install github.com/shepard-labs/go-us-uk-english-translator/cmd/translate@latest
go install github.com/shepard-labs/go-us-uk-english-translator/cmd/apiserver@latest
```

### Requirements

- Go 1.25 or later

---

## Usage

### CLI Quick Start

Run the CLI tool (`translate`) against files or directories:

```bash
# Convert a single file to American English in place
translate --american file.txt

# Preview changes as a unified diff (no modifications)
translate --american --dry-run file.txt

# Recursively convert all markdown and code files in a directory
translate --american --recursive ./src

# Read from stdin and output to stdout
cat document.txt | translate --stdin > document_us.txt
```

#### CLI Flags

| Flag | Short | Description |
|---|---|---|
| `--american` | | Translate to American English (default). |
| `--british` | | Translate to British English. |
| `--dry-run` | `-n` | Print a unified diff; do not modify files. |
| `--recursive` | `-r` | Walk directories recursively. |
| `--ext` | | File extensions to process (e.g., `.ts,.md,.txt`). |
| `--exclude` | | Directory names to skip (default: `node_modules`, `.git`, etc.). |
| `--no-code` | | Skip replacement inside code fences and import paths (default: true). |
| `--dictionary` | `-d` | Path to an additional JSON dictionary to merge. |
| `--ignore-words` | | Path to a plain-text file (one word per line) to exclude from replacements. |

### Library Quick Start

The core conversion logic is decoupled and exported for use as a Go library:

```go
package main

import (
	"fmt"
	"log"

	"github.com/shepard-labs/go-us-uk-english-translator/translator"
)

func main() {
	// Initialize the converter for UK -> US translation
	converter, err := translator.NewConverter(translator.DirectionAmerican)
	if err != nil {
		log.Fatal(err)
	}

	input := "I need to organise my colours."
	output, replacements := converter.Convert(input)
	
	fmt.Printf("Converted: %s\n", output)
	fmt.Printf("Replacements made: %d\n", replacements)
}
```

Output:

```
Converted: I need to organize my colors.
Replacements made: 2
```

### Self-Hosted API Server

The repository includes a ready-to-deploy HTTP API server for text conversion.

**Build and run:**

```bash
go build -o translate-server ./cmd/apiserver
./translate-server
```

The server listens on port `8080` by default.

**Endpoint:**

```
POST /v1/convert
```

**Example request:**

```bash
curl -X POST http://localhost:8080/v1/convert \
  -H "Content-Type: application/json" \
  -d '{
    "text": "The colour of the programme is nice.",
    "target": "american"
  }'
```

**Example response:**

```json
{
  "original_text": "The colour of the programme is nice.",
  "converted_text": "The color of the program is nice.",
  "replacements_made": 2,
  "target_direction": "american"
}
```

---

## Project Structure

```
github.com/shepard-labs/go-us-uk-english-translator/
├── cmd/
│   ├── translate/
│   │   └── main.go              # CLI entry point
│   └── apiserver/
│       └── main.go              # Self-hosted HTTP API server
├── translator/
│   └── translator.go            # Exported library interface
├── internal/
│   ├── dict/
│   │   ├── dict.go              # Dictionary loader and merger
│   │   └── uk_spellings.json    # Canonical UK->US dictionary (embedded)
│   ├── replace/
│   │   ├── replace.go           # Core replacement logic
│   │   └── replace_test.go      # Unit tests for accuracy safeguards
│   └── runner/
│       ├── runner.go            # File discovery and orchestration
│       └── runner_test.go       # Integration tests
├── testdata/                    # Fixtures for testing (.input and .expected)
├── go.mod
└── go.sum
```

---

## FAQ

### How does it prevent false positives?

The replacement engine uses `github.com/rivo/uniseg` to tokenize text according to Unicode Standard Annex #29. This guarantees that dictionary keys are matched exclusively as complete words. Substrings like "ise" in "precise" or "rise" in "enterprise" are mathematically impossible to match as distinct tokens.

### How does it avoid breaking code?

By default, the engine scans the input for exclusion zones and records their byte ranges before attempting replacements. Tokens falling within these byte ranges are skipped. Supported exclusion zones include:
- Markdown fenced code blocks (` ```...``` `)
- Inline code blocks (`` `...` ``)
- URLs (`http://`, `https://`, `ftp://`)
- Import paths in code (e.g. `import from './colour-utils'`)
- HTML attribute values (e.g. `href="colour.css"`)

### Is it safe to run multiple times?

Yes, the engine is fully idempotent. If you convert a file to American English and run it again, it will produce 0 replacements and the file content will remain identical.

### Can I add my own words?

Yes. Pass `--dictionary custom.json` to the CLI with a mapping format identical to `uk_spellings.json`, or provide an `--ignore-words list.txt` to prevent specific words from being changed even if they exist in the built-in dictionary.

---

## Credits

This project relies on the following open source packages:

- **[uniseg](https://github.com/rivo/uniseg)** (`github.com/rivo/uniseg`) — Implements Unicode Standard Annex #29 for accurate word tokenization and boundary detection.

---

## Contributing

We welcome contributions of all kinds — bug fixes, new features, documentation improvements, and dataset updates.

Please read [CONTRIBUTING.md](CONTRIBUTING.md) for detailed guidelines on how to contribute, including code style, testing requirements, and the pull request process.

---

## License

This project is licensed under the **GNU General Public License v3.0**. See [LICENSE](LICENSE) for the full license text.