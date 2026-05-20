# Contributing to shepard-labs/go-us-uk-english-translator

Thank you for your interest in contributing to shepard-labs/go-us-uk-english-translator. This document provides guidelines and instructions for contributing.

---

## Table of Contents

- [Getting Started](#getting-started)
- [Development Setup](#development-setup)
- [How to Contribute](#how-to-contribute)
  - [Reporting Bugs](#reporting-bugs)
  - [Suggesting Features](#suggesting-features)
  - [Submitting Code Changes](#submitting-code-changes)
- [Code Style](#code-style)
- [Testing](#testing)
- [Updating the Dictionary](#updating-the-dictionary)
- [Pull Request Process](#pull-request-process)
- [License](#license)

---

## Getting Started

1. Fork the repository on GitHub.
2. Clone your fork locally:
   ```bash
   git clone https://github.com/shepard-labs/go-us-uk-english-translator.git
   cd go-us-uk-english-translator
   ```
3. Add the upstream remote:
   ```bash
   git remote add upstream https://github.com/shepard-labs/go-us-uk-english-translator.git
   ```
4. Create a branch for your work:
   ```bash
   git checkout -b my-feature
   ```

---

## Development Setup

### Requirements

- **Go 1.25** or later

### Install dependencies

```bash
go mod download
```

### Build

```bash
go build ./...
```

### Run tests

```bash
go test -v ./...
```

### Run the API server locally

```bash
go run ./cmd/apiserver
```

The server starts on `http://localhost:8080` by default.

---

## How to Contribute

### Reporting Bugs

Open a GitHub issue with:

- A clear, descriptive title.
- Steps to reproduce the problem.
- Expected behavior vs. actual behavior.
- Go version (`go version`) and OS.
- Minimal code or text snippet that reproduces the issue, if applicable.

### Suggesting Features

Open a GitHub issue tagged as a feature request with:

- A description of the problem the feature would solve.
- Your proposed solution or API design.
- Any alternatives you have considered.

### Submitting Code Changes

1. Ensure your change addresses an open issue or is discussed beforehand for larger features.
2. Write tests that cover your change.
3. Run the full test suite and confirm all tests pass.
4. Submit a pull request against the `main` branch.

---

## Code Style

- Follow standard Go conventions as enforced by `gofmt` and `go vet`.
- Run `gofmt` on all Go files before committing:
  ```bash
  gofmt -w .
  ```
- Run `go vet` to catch common issues:
  ```bash
  go vet ./...
  ```
- Keep functions focused and small. Each `.go` file corresponds to a single feature or concern.
- Exported types and functions should have documentation comments.
- Do not add dependencies without discussion. The library intentionally has a minimal dependency footprint (currently just `github.com/rivo/uniseg`).

---

## Testing

All changes must include tests. We have strict accuracy safeguards: every substitution must be guarded by whole-word boundaries, casing must be preserved, and exclusion zones (code fences, URLs, import paths) must be respected.

### Running tests

```bash
go test -v -count=1 ./...
```

### Writing tests

- Test function names should follow the pattern `TestFeatureName_Scenario` (e.g., `TestConvert_Idempotency`).
- Test both positive and negative cases (e.g. ensure substrings like "ise" in "precise" are NOT converted).
- When modifying the CLI `runner`, use integration tests with `.input` and `.expected` fixtures in the `testdata/` directory.
- Use table-driven tests when testing multiple inputs for the same function.

---

## Updating the Dictionary

The built-in translation dictionary lives in `internal/dict/uk_spellings.json` and is embedded into the binary at compile time. 

If you are proposing an addition or modification to the dictionary:

1. **Verify it is a canonical difference**: Ensure the term represents a widely recognized spelling difference between British and American English, rather than regional slang, jargon, or non-standard variations.
2. **Beware of false positives**: Some British words have valid, distinct American meanings (e.g., "boot" vs "trunk"). The core translation engine translates text indiscriminately as long as it matches a word boundary. Do not add words that would corrupt unrelated text.
3. **Format**: The file is a JSON object where each key is a UK spelling and each value is the US equivalent. Keys should be in lowercase.
4. **Test**: Ensure you run `go test ./...` after modifying the JSON file to catch any formatting errors and ensure the `go:embed` directive properly captures the changes.

---

## Pull Request Process

1. **One concern per PR.** Keep pull requests focused. A bug fix and a new feature should be separate PRs.
2. **Descriptive title and body.** Explain what the change does and why. Reference the relevant issue number if applicable.
3. **All tests must pass.** The full test suite (`go test ./...`) must pass before a PR will be reviewed.
4. **Code must compile cleanly.** `go build ./...` and `go vet ./...` must produce no errors or warnings.
5. **Formatting.** All Go code must be formatted with `gofmt`.
6. **Review.** At least one maintainer review is required before merging.
7. **Squash on merge.** PRs are squash-merged to keep the commit history clean.

---

## License

By contributing to shepard-labs/go-us-uk-english-translator, you agree that your contributions will be licensed under the [GNU General Public License v3.0](LICENSE).