package replace

import (
	"bytes"
	"fmt"
	"os"
	"regexp"
	"sort"
	"strings"
	"unicode"

	"github.com/rivo/uniseg"
)

var (
	exclusionPatterns = []*regexp.Regexp{
		regexp.MustCompile(`(?s)` + "```.*?```"),
		regexp.MustCompile("`[^`\n]+`"),
		regexp.MustCompile(`\b(?:https?|ftp)://[^\s"'>]+`),
		regexp.MustCompile(`(?:from|require\s*\()\s*["']([^"']+)["']`),
		regexp.MustCompile(`(?:href|src|class|id)="[^"]+"`),
	}
)

type interval struct {
	start int
	end   int
}

// Convert processes the input text, applies dictionary replacements,
// and skips exclusion zones. It returns the converted text and the number of replacements.
// It also prints a warning to os.Stderr if an unconverted dictionary key is found
// in the output as a safety check (per the spec).
func Convert(input string, dict map[string]string, skipExclusionZones bool, filename string) (string, int) {
	var exclusions []interval

	if skipExclusionZones {
		for _, re := range exclusionPatterns {
			matches := re.FindAllStringIndex(input, -1)
			for _, m := range matches {
				exclusions = append(exclusions, interval{start: m[0], end: m[1]})
			}
		}

		// Sort and merge intervals
		sort.Slice(exclusions, func(i, j int) bool {
			return exclusions[i].start < exclusions[j].start
		})

		var merged []interval
		for _, exc := range exclusions {
			if len(merged) == 0 {
				merged = append(merged, exc)
			} else {
				last := &merged[len(merged)-1]
				if exc.start <= last.end {
					if exc.end > last.end {
						last.end = exc.end
					}
				} else {
					merged = append(merged, exc)
				}
			}
		}
		exclusions = merged
	}

	isExcluded := func(start, end int) bool {
		// Binary search could be used, but since the number of intervals
		// might not be huge, a linear scan is fine.
		for _, exc := range exclusions {
			if start < exc.end && end > exc.start {
				return true
			}
		}
		return false
	}

	var buf bytes.Buffer
	buf.Grow(len(input))
	replacementsCount := 0

	currentByteIndex := 0
	state := -1
	rest := input
	for len(rest) > 0 {
		var wordStr string
		wordStr, rest, state = uniseg.FirstWordInString(rest, state)

		startIdx := currentByteIndex
		endIdx := currentByteIndex + len(wordStr)
		currentByteIndex = endIdx

		// Check if the segment is a word consisting of letters
		// We only want to replace actual words.
		isWord := false
		for _, r := range wordStr {
			if unicode.IsLetter(r) {
				isWord = true
				break
			}
		}

		if !isWord || isExcluded(startIdx, endIdx) {
			buf.WriteString(wordStr)
			// Don't warn for excluded words!
			continue
		}

		// It's a valid word outside exclusion zones.
		lowerWord := strings.ToLower(wordStr)
		if replacement, exists := dict[lowerWord]; exists {
			// Apply casing rules
			replacementStr := applyCasing(wordStr, replacement)
			buf.WriteString(replacementStr)
			replacementsCount++
		} else {
			buf.WriteString(wordStr)
		}
	}

	result := buf.String()

	// Perform the self-check pass on the original string's non-excluded words?
	// The spec says "After replacement, scan the output...".
	// Let's do it by tokenizing the result, but since we want to avoid false positives in exclusion zones,
	// and tracking exclusion zones in the output is complex (offsets change due to replacements),
	// a practical way is to check the non-excluded words during the FIRST pass to see if they
	// accidentally remained (or if the replacement string itself is a dictionary key, which is what the self-check is likely catching).
	// Let's implement the self-check during the first pass, right after writing to `buf`.
	// Wait, the spec explicitly says: "scan the output for any remaining dictionary key at a word boundary".
	// Let's calculate the line number for the *output*.
	
	// Actually, if we just check the non-excluded tokens:
	// We can't just do that because the replacement itself might introduce a dictionary key!
	// So we need to check the newly written string.
	// Let's calculate the line numbers of `result` and check.
	// But how do we avoid false positives in exclusion zones?
	// We can map `result` offsets back to `input` offsets, or simply apply `exclusionPatterns` on `result`!
	
	var resultExclusions []interval
	if skipExclusionZones {
		for _, re := range exclusionPatterns {
			matches := re.FindAllStringIndex(result, -1)
			for _, m := range matches {
				resultExclusions = append(resultExclusions, interval{start: m[0], end: m[1]})
			}
		}

		sort.Slice(resultExclusions, func(i, j int) bool {
			return resultExclusions[i].start < resultExclusions[j].start
		})

		var merged []interval
		for _, exc := range resultExclusions {
			if len(merged) == 0 {
				merged = append(merged, exc)
			} else {
				last := &merged[len(merged)-1]
				if exc.start <= last.end {
					if exc.end > last.end {
						last.end = exc.end
					}
				} else {
					merged = append(merged, exc)
				}
			}
		}
		resultExclusions = merged
	}

	isResultExcluded := func(start, end int) bool {
		for _, exc := range resultExclusions {
			if start < exc.end && end > exc.start {
				return true
			}
		}
		return false
	}

	checkState := -1
	checkRest := result
	currentCheckByteIndex := 0
	lineNumber := 1

	for len(checkRest) > 0 {
		var wordStr string
		wordStr, checkRest, checkState = uniseg.FirstWordInString(checkRest, checkState)

		startIdx := currentCheckByteIndex
		endIdx := currentCheckByteIndex + len(wordStr)
		currentCheckByteIndex = endIdx

		// Update line number
		lineNumber += strings.Count(wordStr, "\n")

		// Only check actual words outside exclusion zones
		isWord := false
		for _, r := range wordStr {
			if unicode.IsLetter(r) {
				isWord = true
				break
			}
		}

		if !isWord || isResultExcluded(startIdx, endIdx) {
			continue
		}

		lowerWord := strings.ToLower(wordStr)
		if _, exists := dict[lowerWord]; exists {
			fmt.Fprintf(os.Stderr, "WARNING: %s:%d: possible unconverted UK spelling: '%s'\n", filename, lineNumber, wordStr)
		}
	}

	return result, replacementsCount
}

func applyCasing(original, replacement string) string {
	allUpper := true
	allLower := true

	runes := []rune(original)
	for _, r := range runes {
		if unicode.IsLower(r) {
			allUpper = false
		}
		if unicode.IsUpper(r) {
			allLower = false
		}
	}

	// Mixed case if neither allUpper, allLower, nor purely Title Case
	// Actually, if it's title case, the first is upper, rest are lower.
	isTitle := false
	if len(runes) > 0 && unicode.IsUpper(runes[0]) {
		restLower := true
		for _, r := range runes[1:] {
			if unicode.IsUpper(r) {
				restLower = false
				break
			}
		}
		isTitle = restLower
	}

	if allUpper {
		return strings.ToUpper(replacement)
	}
	if isTitle {
		replRunes := []rune(replacement)
		if len(replRunes) > 0 {
			first := unicode.ToUpper(replRunes[0])
			rest := strings.ToLower(string(replRunes[1:]))
			return string(first) + rest
		}
	}
	if allLower {
		return strings.ToLower(replacement)
	}
	// Mixed or other forms -> lowercase replacement
	return strings.ToLower(replacement)
}
