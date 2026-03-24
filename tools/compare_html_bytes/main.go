// compare_html_bytes — Byte-level comparison of Go vs C# HTML output.
//
// For each .html file in csharp-html-output/ that also exists in html-output/,
// produce a detailed character-level diff and write it to html-delta/<filename>.md.
//
// The C# output is the ground-truth (expected). The Go output is the actual.
//
// Usage:
//
//	go run ./tools/compare_html_bytes/
//	go run ./tools/compare_html_bytes/ --csharp csharp-html-output --go html-output --out html-delta
//	go run ./tools/compare_html_bytes/ --report "Simple List"
package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// byteStats holds the result of comparing two byte sequences.
type byteStats struct {
	identical    bool
	expectedLen  int
	actualLen    int
	lenDiff      int
	firstDiff    int // -1 if not applicable
	hasFirstDiff bool
}

// computeByteStats compares two byte slices and returns stats.
func computeByteStats(expected, actual []byte) byteStats {
	if string(expected) == string(actual) {
		return byteStats{
			identical:   true,
			expectedLen: len(expected),
			actualLen:   len(actual),
		}
	}

	firstDiff := -1
	minLen := len(expected)
	if len(actual) < minLen {
		minLen = len(actual)
	}
	for i := 0; i < minLen; i++ {
		if expected[i] != actual[i] {
			firstDiff = i
			break
		}
	}
	if firstDiff == -1 {
		// One is a prefix of the other.
		firstDiff = minLen
	}

	return byteStats{
		identical:    false,
		expectedLen:  len(expected),
		actualLen:    len(actual),
		lenDiff:      len(actual) - len(expected),
		firstDiff:    firstDiff,
		hasFirstDiff: true,
	}
}

// contextAround returns the text before and after the given byte position.
func contextAround(text string, bytePos, radius int) (before, after string) {
	start := bytePos - radius
	if start < 0 {
		start = 0
	}
	end := bytePos + radius
	if end > len(text) {
		end = len(text)
	}
	snippet := text[start:end]
	markerPos := bytePos - start
	return snippet[:markerPos], snippet[markerPos:]
}

// unifiedDiff produces a simple unified diff between two texts.
// It returns the diff lines and whether the output was truncated.
func unifiedDiff(expectedText, actualText string, maxLines int) ([]string, bool) {
	expLines := splitLines(expectedText)
	actLines := splitLines(actualText)

	hunks := computeHunks(expLines, actLines, 3)

	var out []string
	out = append(out, "--- expected (C#)")
	out = append(out, "+++ actual (Go)")

	for _, h := range hunks {
		out = append(out, h...)
	}

	truncated := false
	if len(out) > maxLines {
		out = out[:maxLines]
		truncated = true
	}
	return out, truncated
}

// splitLines splits text into lines, preserving line content without trailing newlines.
func splitLines(s string) []string {
	if s == "" {
		return nil
	}
	lines := strings.Split(s, "\n")
	return lines
}

// computeHunks generates unified-diff hunks with context lines.
func computeHunks(a, b []string, contextLines int) [][]string {
	// Use a simple LCS-based diff to find matching blocks.
	opcodes := sequenceMatcherOpcodes(a, b)

	// Group opcodes into hunks.
	groups := groupOpcodes(opcodes, contextLines)

	var hunks [][]string
	for _, group := range groups {
		var hunk []string

		// Calculate hunk header ranges.
		firstOp := group[0]
		lastOp := group[len(group)-1]
		aStart := firstOp.aStart + 1 // 1-based
		aCount := lastOp.aEnd - firstOp.aStart
		bStart := firstOp.bStart + 1 // 1-based
		bCount := lastOp.bEnd - firstOp.bStart

		hunk = append(hunk, fmt.Sprintf("@@ -%d,%d +%d,%d @@", aStart, aCount, bStart, bCount))

		for _, op := range group {
			switch op.tag {
			case opEqual:
				for i := op.aStart; i < op.aEnd; i++ {
					hunk = append(hunk, " "+a[i])
				}
			case opDelete:
				for i := op.aStart; i < op.aEnd; i++ {
					hunk = append(hunk, "-"+a[i])
				}
			case opInsert:
				for i := op.bStart; i < op.bEnd; i++ {
					hunk = append(hunk, "+"+b[i])
				}
			case opReplace:
				for i := op.aStart; i < op.aEnd; i++ {
					hunk = append(hunk, "-"+a[i])
				}
				for i := op.bStart; i < op.bEnd; i++ {
					hunk = append(hunk, "+"+b[i])
				}
			}
		}
		hunks = append(hunks, hunk)
	}
	return hunks
}

type opTag int

const (
	opEqual   opTag = iota
	opReplace
	opInsert
	opDelete
)

type opcode struct {
	tag    opTag
	aStart int
	aEnd   int
	bStart int
	bEnd   int
}

// sequenceMatcherOpcodes computes the opcodes to transform a into b,
// similar to Python's difflib.SequenceMatcher.get_opcodes().
func sequenceMatcherOpcodes(a, b []string) []opcode {
	matches := longestCommonSubsequence(a, b)

	var ops []opcode
	ai, bi := 0, 0

	for _, m := range matches {
		if ai < m[0] || bi < m[1] {
			if ai < m[0] && bi < m[1] {
				ops = append(ops, opcode{opReplace, ai, m[0], bi, m[1]})
			} else if ai < m[0] {
				ops = append(ops, opcode{opDelete, ai, m[0], bi, bi})
			} else {
				ops = append(ops, opcode{opInsert, ai, ai, bi, m[1]})
			}
		}
		if m[2] > 0 {
			ops = append(ops, opcode{opEqual, m[0], m[0] + m[2], m[1], m[1] + m[2]})
		}
		ai = m[0] + m[2]
		bi = m[1] + m[2]
	}

	if ai < len(a) || bi < len(b) {
		if ai < len(a) && bi < len(b) {
			ops = append(ops, opcode{opReplace, ai, len(a), bi, len(b)})
		} else if ai < len(a) {
			ops = append(ops, opcode{opDelete, ai, len(a), bi, bi})
		} else {
			ops = append(ops, opcode{opInsert, ai, ai, bi, len(b)})
		}
	}

	return ops
}

// longestCommonSubsequence returns matching blocks as [aStart, bStart, size] triples.
func longestCommonSubsequence(a, b []string) [][3]int {
	// Build index of b lines for faster matching.
	bIndex := make(map[string][]int)
	for i, line := range b {
		bIndex[line] = append(bIndex[line], i)
	}

	var blocks [][3]int
	findMatchingBlocks(a, b, bIndex, 0, len(a), 0, len(b), &blocks)
	sort.Slice(blocks, func(i, j int) bool {
		if blocks[i][0] != blocks[j][0] {
			return blocks[i][0] < blocks[j][0]
		}
		return blocks[i][1] < blocks[j][1]
	})
	return blocks
}

// findMatchingBlocks recursively finds the longest matching block and recurses.
func findMatchingBlocks(a, b []string, bIndex map[string][]int,
	aLo, aHi, bLo, bHi int, blocks *[][3]int) {

	bestA, bestB, bestSize := aLo, bLo, 0

	// For each line in a[aLo:aHi], find occurrences in b[bLo:bHi] and track
	// how long a run of matches extends.
	// j2len maps b-index to current match length.
	j2len := make(map[int]int)
	for i := aLo; i < aHi; i++ {
		newJ2len := make(map[int]int)
		for _, j := range bIndex[a[i]] {
			if j < bLo {
				continue
			}
			if j >= bHi {
				break
			}
			k := j2len[j-1] + 1
			newJ2len[j] = k
			if k > bestSize {
				bestA = i - k + 1
				bestB = j - k + 1
				bestSize = k
			}
		}
		j2len = newJ2len
	}

	if bestSize > 0 {
		if bestA > aLo && bestB > bLo {
			findMatchingBlocks(a, b, bIndex, aLo, bestA, bLo, bestB, blocks)
		}
		*blocks = append(*blocks, [3]int{bestA, bestB, bestSize})
		if bestA+bestSize < aHi && bestB+bestSize < bHi {
			findMatchingBlocks(a, b, bIndex, bestA+bestSize, aHi, bestB+bestSize, bHi, blocks)
		}
	}
}

// groupOpcodes groups opcodes into hunks separated by enough equal lines.
func groupOpcodes(ops []opcode, n int) [][]opcode {
	if len(ops) == 0 {
		return nil
	}

	// If the first opcode is equal, trim it to at most n context lines.
	// If the last opcode is equal, trim it similarly.
	var groups [][]opcode
	var currentGroup []opcode

	for i, op := range ops {
		if op.tag == opEqual {
			equalLen := op.aEnd - op.aStart
			if equalLen > 2*n {
				// Split: end of one hunk, start of another.
				if i > 0 || op.aStart < op.aEnd-n {
					// Add trailing context to current group.
					if currentGroup != nil || i == 0 {
						end := op.aStart + n
						if i == 0 {
							// First opcode: only trailing context.
							end = op.aStart + n
							if end > op.aEnd {
								end = op.aEnd
							}
						}
						endB := op.bStart + (end - op.aStart)
						if end > op.aEnd {
							end = op.aEnd
							endB = op.bEnd
						}
						trimmed := opcode{opEqual, op.aStart, end, op.bStart, endB}
						currentGroup = append(currentGroup, trimmed)
						groups = append(groups, currentGroup)
						currentGroup = nil
					}
				}
				// Leading context for next hunk.
				if i < len(ops)-1 {
					start := op.aEnd - n
					if start < op.aStart {
						start = op.aStart
					}
					startB := op.bEnd - (op.aEnd - start)
					currentGroup = []opcode{{opEqual, start, op.aEnd, startB, op.bEnd}}
				}
			} else {
				currentGroup = append(currentGroup, op)
			}
		} else {
			if currentGroup == nil {
				currentGroup = []opcode{}
			}
			currentGroup = append(currentGroup, op)
		}
	}

	if currentGroup != nil {
		groups = append(groups, currentGroup)
	}

	return groups
}

// lineSimilarity computes a simple similarity ratio between two string slices.
// This is an approximation of Python difflib.SequenceMatcher.ratio().
func lineSimilarity(a, b []string) float64 {
	if len(a) == 0 && len(b) == 0 {
		return 1.0
	}

	matches := longestCommonSubsequence(a, b)
	totalMatching := 0
	for _, m := range matches {
		totalMatching += m[2]
	}
	return 2.0 * float64(totalMatching) / float64(len(a)+len(b))
}

// formatComma formats an integer with comma separators.
func formatComma(n int) string {
	sign := ""
	if n < 0 {
		sign = "-"
		n = -n
	}
	s := fmt.Sprintf("%d", n)
	if len(s) <= 3 {
		return sign + s
	}
	var parts []string
	for len(s) > 3 {
		parts = append([]string{s[len(s)-3:]}, parts...)
		s = s[:len(s)-3]
	}
	if len(s) > 0 {
		parts = append([]string{s}, parts...)
	}
	return sign + strings.Join(parts, ",")
}

// formatSignedComma formats with a leading + or - sign and comma separators.
func formatSignedComma(n int) string {
	if n >= 0 {
		return "+" + formatComma(n)
	}
	return formatComma(n)
}

// renderMD generates a markdown report for one file comparison.
func renderMD(name string, stats byteStats, expectedText, actualText string) string {
	var b strings.Builder

	fmt.Fprintf(&b, "# %s\n\n", name)

	if stats.identical {
		b.WriteString("**Status:** PASS (identical)\n")
		fmt.Fprintf(&b, "**Size:** %s bytes\n\n", formatComma(stats.expectedLen))
		return b.String()
	}

	b.WriteString("**Status:** FAIL (differences found)\n")
	fmt.Fprintf(&b, "**Expected size (C#):** %s bytes\n", formatComma(stats.expectedLen))
	fmt.Fprintf(&b, "**Actual size (Go):** %s bytes\n", formatComma(stats.actualLen))
	fmt.Fprintf(&b, "**Size difference:** %s bytes\n\n", formatSignedComma(stats.lenDiff))

	if stats.hasFirstDiff {
		pos := stats.firstDiff
		fmt.Fprintf(&b, "**First difference at byte:** %s\n\n", formatComma(pos))

		b.WriteString("## Context Around First Difference\n\n")

		if pos < len(expectedText) {
			beforeExp, afterExp := contextAround(expectedText, pos, 120)
			b.WriteString("**Expected (C#):**\n```\n")
			fmt.Fprintf(&b, "%s>>>HERE<<<%s\n", beforeExp, afterExp)
			b.WriteString("```\n\n")
		}

		if pos < len(actualText) {
			beforeAct, afterAct := contextAround(actualText, pos, 120)
			b.WriteString("**Actual (Go):**\n```\n")
			fmt.Fprintf(&b, "%s>>>HERE<<<%s\n", beforeAct, afterAct)
			b.WriteString("```\n\n")
		}
	}

	// Line-level diff summary.
	expLines := strings.Split(expectedText, "\n")
	actLines := strings.Split(actualText, "\n")
	fmt.Fprintf(&b, "**Expected lines:** %s\n", formatComma(len(expLines)))
	fmt.Fprintf(&b, "**Actual lines:** %s\n\n", formatComma(len(actLines)))

	ratio := lineSimilarity(expLines, actLines)
	fmt.Fprintf(&b, "**Line similarity:** %.1f%%\n\n", ratio*100)

	// Unified diff.
	diff, truncated := unifiedDiff(expectedText, actualText, 500)
	if len(diff) > 0 {
		b.WriteString("## Unified Diff\n\n")
		if truncated {
			b.WriteString("_(Truncated to first 500 lines of diff)_\n\n")
		}
		b.WriteString("```diff\n")
		for _, d := range diff {
			fmt.Fprintf(&b, "%s\n", d)
		}
		b.WriteString("```\n\n")
	}

	return b.String()
}

type result struct {
	name  string
	stats *byteStats // nil means missing
}

// renderSummary generates the summary README.md.
func renderSummary(results []result) string {
	var b strings.Builder

	b.WriteString("# HTML Byte-Level Comparison Summary\n\n")
	b.WriteString("Character-level comparison of Go HTML output (`html-output/`) against C# ground-truth (`csharp-html-output/`).\n\n")

	var passList []result
	var failList []result
	var missingList []result
	for _, r := range results {
		if r.stats == nil {
			missingList = append(missingList, r)
		} else if r.stats.identical {
			passList = append(passList, r)
		} else {
			failList = append(failList, r)
		}
	}

	total := len(results)
	b.WriteString("| Category | Count |\n")
	b.WriteString("|---|---|\n")
	fmt.Fprintf(&b, "| Identical | %d |\n", len(passList))
	fmt.Fprintf(&b, "| Different | %d |\n", len(failList))
	fmt.Fprintf(&b, "| Go output missing | %d |\n", len(missingList))
	fmt.Fprintf(&b, "| **Total** | **%d** |\n\n", total)

	if len(failList) > 0 {
		b.WriteString("## Different\n\n")
		b.WriteString("| Report | C# Size | Go Size | Size Diff | First Diff Byte |\n")
		b.WriteString("|---|---|---|---|---|\n")
		for _, r := range failList {
			link := strings.ReplaceAll(r.name, " ", "%20")
			csSz := formatComma(r.stats.expectedLen)
			goSz := formatComma(r.stats.actualLen)
			szDiff := formatSignedComma(r.stats.lenDiff)
			first := "?"
			if r.stats.hasFirstDiff {
				first = formatComma(r.stats.firstDiff)
			}
			fmt.Fprintf(&b, "| [%s](%s.md) | %s | %s | %s | %s |\n", r.name, link, csSz, goSz, szDiff, first)
		}
		b.WriteString("\n")
	}

	if len(passList) > 0 {
		b.WriteString("## Identical\n\n")
		b.WriteString("| Report | Size |\n")
		b.WriteString("|---|---|\n")
		for _, r := range passList {
			fmt.Fprintf(&b, "| %s | %s |\n", r.name, formatComma(r.stats.expectedLen))
		}
		b.WriteString("\n")
	}

	if len(missingList) > 0 {
		b.WriteString("## Go Output Missing\n\n")
		for _, r := range missingList {
			fmt.Fprintf(&b, "- %s\n", r.name)
		}
		b.WriteString("\n")
	}

	return b.String()
}

func main() {
	csDir := flag.String("csharp", "csharp-html-output", "Directory with C# ground-truth HTML files")
	goDir := flag.String("go", "html-output", "Directory with Go HTML files")
	outDir := flag.String("out", "html-delta", "Output directory for delta .md files")
	report := flag.String("report", "", "Process only this report name (without .html)")
	flag.Parse()

	// Validate input directories.
	if info, err := os.Stat(*csDir); err != nil || !info.IsDir() {
		fmt.Fprintf(os.Stderr, "error: C# output directory not found: %s\n", *csDir)
		os.Exit(1)
	}
	if info, err := os.Stat(*goDir); err != nil || !info.IsDir() {
		fmt.Fprintf(os.Stderr, "error: Go output directory not found: %s\n", *goDir)
		os.Exit(1)
	}

	// Create output directory.
	if err := os.MkdirAll(*outDir, 0o755); err != nil {
		fmt.Fprintf(os.Stderr, "error: cannot create output directory: %v\n", err)
		os.Exit(1)
	}

	// Find C# HTML files.
	csFiles, err := filepath.Glob(filepath.Join(*csDir, "*.html"))
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
	sort.Strings(csFiles)

	// Filter by report name if specified.
	if *report != "" {
		reportLower := strings.ToLower(*report)
		var filtered []string
		for _, f := range csFiles {
			stem := strings.TrimSuffix(filepath.Base(f), ".html")
			if strings.ToLower(stem) == reportLower {
				filtered = append(filtered, f)
			}
		}
		if len(filtered) == 0 {
			fmt.Fprintf(os.Stderr, "error: no C# HTML file matching %q\n", *report)
			os.Exit(1)
		}
		csFiles = filtered
	}

	var results []result

	for _, csPath := range csFiles {
		baseName := filepath.Base(csPath)
		name := strings.TrimSuffix(baseName, ".html")
		goPath := filepath.Join(*goDir, baseName)

		// Check if Go output exists.
		if _, err := os.Stat(goPath); os.IsNotExist(err) {
			fmt.Printf("  MISSING  %s\n", name)
			results = append(results, result{name: name, stats: nil})
			md := fmt.Sprintf("# %s\n\n**Status:** MISSING — Go output file not found.\n", name)
			mdPath := filepath.Join(*outDir, name+".md")
			os.WriteFile(mdPath, []byte(md), 0o644)
			continue
		}

		expectedBytes, err := os.ReadFile(csPath)
		if err != nil {
			fmt.Printf("  ERROR    %s: %v\n", name, err)
			results = append(results, result{name: name, stats: nil})
			continue
		}
		actualBytes, err := os.ReadFile(goPath)
		if err != nil {
			fmt.Printf("  ERROR    %s: %v\n", name, err)
			results = append(results, result{name: name, stats: nil})
			continue
		}

		stats := computeByteStats(expectedBytes, actualBytes)

		if stats.identical {
			fmt.Printf("  PASS     %-55s %10s bytes\n", name, formatComma(stats.expectedLen))
			// Remove any old delta file for this report.
			oldMD := filepath.Join(*outDir, name+".md")
			os.Remove(oldMD) // ignore error
		} else {
			expectedText := string(expectedBytes)
			actualText := string(actualBytes)
			md := renderMD(name, stats, expectedText, actualText)
			mdPath := filepath.Join(*outDir, name+".md")
			os.WriteFile(mdPath, []byte(md), 0o644)
			firstDiff := "?"
			if stats.hasFirstDiff {
				firstDiff = fmt.Sprintf("%d", stats.firstDiff)
			}
			fmt.Printf("  FAIL     %-55s C#=%10s  Go=%10s  diff@%s\n",
				name, formatComma(stats.expectedLen), formatComma(stats.actualLen), firstDiff)
		}

		results = append(results, result{name: name, stats: &stats})
	}

	// Write summary.
	if *report == "" {
		summaryMD := renderSummary(results)
		summaryPath := filepath.Join(*outDir, "README.md")
		os.WriteFile(summaryPath, []byte(summaryMD), 0o644)
		fmt.Printf("\nSummary written to %s\n", summaryPath)
	}

	passN := 0
	failN := 0
	missN := 0
	for _, r := range results {
		if r.stats == nil {
			missN++
		} else if r.stats.identical {
			passN++
		} else {
			failN++
		}
	}
	fmt.Printf("\n%d identical, %d different, %d missing — delta docs in %q\n", passN, failN, missN, *outDir+"/")

}
