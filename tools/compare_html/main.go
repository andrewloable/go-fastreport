// compare_html compares Go HTML output against C# (ground-truth) HTML output.
//
// For each .html file present in csharp-html-output/ that also exists in html-output/,
// the program parses both files, extracts structured data, and writes a per-report
// <filename>.md delta document into html-delta/.
//
// Usage:
//
//	go run ./tools/compare_html
//	go run ./tools/compare_html --csharp csharp-html-output --go html-output --out html-delta
//	go run ./tools/compare_html --report "Simple List"
package main

import (
	"flag"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"golang.org/x/net/html"
)

// ---------------------------------------------------------------------------
// HTML parsing helpers
// ---------------------------------------------------------------------------

var pageClassRE = regexp.MustCompile(`\bfrpage\d+\b`)

// parseHTML reads an HTML file and returns per-page text token lists.
func parseHTML(path string) (int, [][]string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return 0, nil, err
	}

	doc, err := html.Parse(strings.NewReader(string(data)))
	if err != nil {
		return 0, nil, err
	}

	var pages [][]string
	var currentPage *[]string

	// skipDepth tracks nesting depth when inside a skip zone (style/script/head).
	// 0 means not skipping.
	skipDepth := 0
	depth := 0

	var walk func(*html.Node)
	walk = func(n *html.Node) {
		switch n.Type {
		case html.ElementNode:
			depth++
			tag := n.Data

			cls := attrVal(n, "class")

			// Detect page container.
			if pageClassRE.MatchString(cls) {
				page := make([]string, 0)
				pages = append(pages, page)
				currentPage = &pages[len(pages)-1]
			}

			// Skip style, script, head blocks.
			if (tag == "style" || tag == "script" || tag == "head") && skipDepth == 0 {
				skipDepth = depth
			}

			for c := n.FirstChild; c != nil; c = c.NextSibling {
				walk(c)
			}

			if skipDepth != 0 && depth == skipDepth {
				skipDepth = 0
			}
			depth--

		case html.TextNode:
			if skipDepth != 0 {
				return
			}
			text := strings.TrimSpace(n.Data)
			if text != "" && currentPage != nil {
				*currentPage = append(*currentPage, text)
			}

		default:
			for c := n.FirstChild; c != nil; c = c.NextSibling {
				walk(c)
			}
		}
	}

	walk(doc)
	return len(pages), pages, nil
}

func attrVal(n *html.Node, key string) string {
	for _, a := range n.Attr {
		if a.Key == key {
			return a.Val
		}
	}
	return ""
}

func flattenTokens(pages [][]string) []string {
	var result []string
	for _, page := range pages {
		result = append(result, page...)
	}
	return result
}

// ---------------------------------------------------------------------------
// Sequence matching (simplified SequenceMatcher)
// ---------------------------------------------------------------------------

// opTag represents the type of edit operation.
type opTag int

const (
	opEqual opTag = iota
	opReplace
	opDelete
	opInsert
)

// opCode represents a single edit operation.
type opCode struct {
	tag        opTag
	i1, i2     int // range in a
	j1, j2     int // range in b
}

// sequenceOpcodes computes edit opcodes between two string slices using a
// simple LCS-based algorithm. For very long sequences it uses a bounded
// approach to avoid excessive runtime.
func sequenceOpcodes(a, b []string) []opCode {
	// Build index of b elements for quick matching.
	bIndex := make(map[string][]int)
	for j, s := range b {
		bIndex[s] = append(bIndex[s], j)
	}

	// Find matching blocks using a patience-like approach.
	type matchBlock struct {
		i, j, size int
	}

	var findMatchingBlocks func(alo, ahi, blo, bhi int) []matchBlock
	findMatchingBlocks = func(alo, ahi, blo, bhi int) []matchBlock {
		var blocks []matchBlock

		// Find longest matching block in the given ranges.
		bestI, bestJ, bestSize := alo, blo, 0

		// j2len tracks the length of the longest match ending at each j.
		j2len := make(map[int]int)
		for i := alo; i < ahi; i++ {
			newJ2len := make(map[int]int)
			for _, j := range bIndex[a[i]] {
				if j < blo {
					continue
				}
				if j >= bhi {
					break
				}
				k := j2len[j-1] + 1
				newJ2len[j] = k
				if k > bestSize {
					bestI = i - k + 1
					bestJ = j - k + 1
					bestSize = k
				}
			}
			j2len = newJ2len
		}

		if bestSize > 0 {
			if bestI > alo && bestJ > blo {
				blocks = append(blocks, findMatchingBlocks(alo, bestI, blo, bestJ)...)
			}
			blocks = append(blocks, matchBlock{bestI, bestJ, bestSize})
			if bestI+bestSize < ahi && bestJ+bestSize < bhi {
				blocks = append(blocks, findMatchingBlocks(bestI+bestSize, ahi, bestJ+bestSize, bhi)...)
			}
		}

		return blocks
	}

	matches := findMatchingBlocks(0, len(a), 0, len(b))
	// Add sentinel.
	matches = append(matches, matchBlock{len(a), len(b), 0})

	var codes []opCode
	i, j := 0, 0
	for _, m := range matches {
		if i < m.i || j < m.j {
			if i < m.i && j < m.j {
				codes = append(codes, opCode{opReplace, i, m.i, j, m.j})
			} else if i < m.i {
				codes = append(codes, opCode{opDelete, i, m.i, j, j})
			} else {
				codes = append(codes, opCode{opInsert, i, i, j, m.j})
			}
		}
		if m.size > 0 {
			codes = append(codes, opCode{opEqual, m.i, m.i + m.size, m.j, m.j + m.size})
		}
		i = m.i + m.size
		j = m.j + m.size
	}

	return codes
}

// sequenceRatio computes the similarity ratio between two string slices.
func sequenceRatio(a, b []string, codes []opCode) float64 {
	if len(a) == 0 && len(b) == 0 {
		return 1.0
	}
	matches := 0
	for _, c := range codes {
		if c.tag == opEqual {
			matches += c.i2 - c.i1
		}
	}
	return 2.0 * float64(matches) / float64(len(a)+len(b))
}

// ---------------------------------------------------------------------------
// Delta analysis
// ---------------------------------------------------------------------------

var unresolvedRE = regexp.MustCompile(`\[[A-Za-z_][A-Za-z0-9_.]*\]`)

func isUnresolved(token string) bool {
	return unresolvedRE.MatchString(token)
}

type delta struct {
	name          string
	csPages       int
	goPages       int
	pageMatch     bool
	missingFromGo []string
	spuriousInGo  []string
	unresolved    []string
	diffHunks     []string
	similarity    float64
}

func analyse(name string, csPages, goPages [][]string) *delta {
	csFlat := flattenTokens(csPages)
	goFlat := flattenTokens(goPages)

	csSet := make(map[string]bool)
	for _, t := range csFlat {
		csSet[t] = true
	}
	goSet := make(map[string]bool)
	for _, t := range goFlat {
		goSet[t] = true
	}

	var missingFromGo []string
	for t := range csSet {
		if !goSet[t] {
			missingFromGo = append(missingFromGo, t)
		}
	}
	sort.Strings(missingFromGo)

	var spuriousInGo []string
	for t := range goSet {
		if !csSet[t] {
			spuriousInGo = append(spuriousInGo, t)
		}
	}
	sort.Strings(spuriousInGo)

	var unresolved []string
	for _, t := range spuriousInGo {
		if isUnresolved(t) {
			unresolved = append(unresolved, t)
		}
	}

	// Sequence diff.
	codes := sequenceOpcodes(csFlat, goFlat)

	var diffHunks []string
	for _, c := range codes {
		if c.tag == opEqual {
			continue
		}
		csChunk := csFlat[c.i1:c.i2]
		goChunk := goFlat[c.j1:c.j2]
		switch c.tag {
		case opReplace:
			diffHunks = append(diffHunks, fmt.Sprintf("replace  CS=%v  ->  Go=%v", csChunk, goChunk))
		case opDelete:
			diffHunks = append(diffHunks, fmt.Sprintf("missing  CS=%v", csChunk))
		case opInsert:
			diffHunks = append(diffHunks, fmt.Sprintf("extra    Go=%v", goChunk))
		}
		if len(diffHunks) >= 30 {
			diffHunks = append(diffHunks, "... (truncated, more differences exist)")
			break
		}
	}

	similarity := sequenceRatio(csFlat, goFlat, codes)

	return &delta{
		name:          name,
		csPages:       len(csPages),
		goPages:       len(goPages),
		pageMatch:     len(csPages) == len(goPages),
		missingFromGo: missingFromGo,
		spuriousInGo:  spuriousInGo,
		unresolved:    unresolved,
		diffHunks:     diffHunks,
		similarity:    similarity,
	}
}

// ---------------------------------------------------------------------------
// Markdown report generation
// ---------------------------------------------------------------------------

const (
	statusPass  = "\u2705 PASS"
	statusMinor = "\u26a0\ufe0f  MINOR"
	statusFail  = "\u274c FAIL"
)

func deltaStatus(d *delta) string {
	if !d.pageMatch {
		return statusFail
	}
	if len(d.unresolved) > 0 {
		return statusFail
	}
	if len(d.missingFromGo) > 0 {
		return statusFail
	}
	if d.similarity < 0.85 {
		return statusMinor
	}
	if len(d.spuriousInGo) > 0 {
		return statusMinor
	}
	return statusPass
}

func renderMD(d *delta) string {
	st := deltaStatus(d)
	var b strings.Builder

	fmt.Fprintf(&b, "# %s\n\n", d.name)
	fmt.Fprintf(&b, "**Status:** %s  \n", st)
	fmt.Fprintf(&b, "**Similarity:** %.0f%%  \n", d.similarity*100)
	pageLine := fmt.Sprintf("**Pages:** C# = %d, Go = %d", d.csPages, d.goPages)
	if !d.pageMatch {
		pageLine += "  \u26a0\ufe0f mismatch"
	}
	fmt.Fprintf(&b, "%s\n\n", pageLine)

	// Unresolved expressions.
	if len(d.unresolved) > 0 {
		b.WriteString("## Unresolved Expressions in Go Output\n\n")
		b.WriteString("These tokens appear to be unevaluated bracket expressions that should have been replaced with data values:\n\n")
		limit := 50
		if len(d.unresolved) < limit {
			limit = len(d.unresolved)
		}
		for _, t := range d.unresolved[:limit] {
			fmt.Fprintf(&b, "- `%s`\n", t)
		}
		if len(d.unresolved) > 50 {
			fmt.Fprintf(&b, "- ... and %d more\n", len(d.unresolved)-50)
		}
		b.WriteString("\n")
	}

	// Missing text.
	if len(d.missingFromGo) > 0 {
		b.WriteString("## Text Present in C# Output but Missing from Go\n\n")
		b.WriteString("These text tokens appear in the ground-truth C# output but are absent from the Go output:\n\n")
		var shown []string
		for _, t := range d.missingFromGo {
			if !isUnresolved(t) {
				shown = append(shown, t)
				if len(shown) >= 60 {
					break
				}
			}
		}
		for _, t := range shown {
			fmt.Fprintf(&b, "- `%s`\n", t)
		}
		if len(d.missingFromGo) > 60 {
			fmt.Fprintf(&b, "- ... and %d more\n", len(d.missingFromGo)-60)
		}
		b.WriteString("\n")
	}

	// Spurious text (not unresolved).
	var extra []string
	for _, t := range d.spuriousInGo {
		if !isUnresolved(t) {
			extra = append(extra, t)
		}
	}
	if len(extra) > 0 {
		b.WriteString("## Text Present in Go Output but Absent from C#\n\n")
		b.WriteString("These tokens appear only in the Go output (may be extra labels, formatting artefacts, or duplicates):\n\n")
		limit := 40
		if len(extra) < limit {
			limit = len(extra)
		}
		for _, t := range extra[:limit] {
			fmt.Fprintf(&b, "- `%s`\n", t)
		}
		if len(extra) > 40 {
			fmt.Fprintf(&b, "- ... and %d more\n", len(extra)-40)
		}
		b.WriteString("\n")
	}

	// Sequence diff hunks.
	if len(d.diffHunks) > 0 {
		b.WriteString("## Sequence Diff Hunks\n\n")
		b.WriteString("Ordered differences between C# and Go text token sequences (up to 30 hunks):\n\n")
		b.WriteString("```\n")
		for _, h := range d.diffHunks {
			b.WriteString(h)
			b.WriteString("\n")
		}
		b.WriteString("```\n\n")
	}

	if st == statusPass {
		b.WriteString("_No significant differences detected._\n\n")
	}

	return b.String()
}

// ---------------------------------------------------------------------------
// Summary index
// ---------------------------------------------------------------------------

type result struct {
	name  string
	delta *delta // nil means missing
}

func renderSummary(results []result) string {
	var b strings.Builder

	b.WriteString("# HTML Delta Summary\n\n")
	b.WriteString("Comparison of Go HTML output (`html-output/`) against C# ground-truth (`csharp-html-output/`).  \n\n")

	var passList, minorList, failList []result
	var missingList []string

	for _, r := range results {
		if r.delta == nil {
			missingList = append(missingList, r.name)
		} else {
			st := deltaStatus(r.delta)
			switch st {
			case statusPass:
				passList = append(passList, r)
			case statusMinor:
				minorList = append(minorList, r)
			default:
				failList = append(failList, r)
			}
		}
	}

	total := len(results)
	b.WriteString("| Category | Count |\n")
	b.WriteString("|---|---|\n")
	fmt.Fprintf(&b, "| \u2705 Pass | %d |\n", len(passList))
	fmt.Fprintf(&b, "| \u26a0\ufe0f  Minor differences | %d |\n", len(minorList))
	fmt.Fprintf(&b, "| \u274c Fail | %d |\n", len(failList))
	fmt.Fprintf(&b, "| \u2796 Go output missing | %d |\n", len(missingList))
	fmt.Fprintf(&b, "| **Total** | **%d** |\n\n", total)

	tableRows := func(entries []result) {
		for _, r := range entries {
			sim := fmt.Sprintf("%.0f%%", r.delta.similarity*100)
			var pages string
			if r.delta.pageMatch {
				pages = fmt.Sprintf("%d", r.delta.csPages)
			} else {
				pages = fmt.Sprintf("%d -> %d \u26a0\ufe0f", r.delta.csPages, r.delta.goPages)
			}
			link := url.PathEscape(r.name)
			fmt.Fprintf(&b, "| [%s](%s.md) | %s | %s |\n", r.name, link, sim, pages)
		}
	}

	if len(failList) > 0 {
		b.WriteString("## \u274c Failures\n\n")
		b.WriteString("| Report | Similarity | Pages |\n")
		b.WriteString("|---|---|---|\n")
		tableRows(failList)
		b.WriteString("\n")
	}

	if len(minorList) > 0 {
		b.WriteString("## \u26a0\ufe0f  Minor Differences\n\n")
		b.WriteString("| Report | Similarity | Pages |\n")
		b.WriteString("|---|---|---|\n")
		tableRows(minorList)
		b.WriteString("\n")
	}

	if len(passList) > 0 {
		b.WriteString("## \u2705 Passing\n\n")
		b.WriteString("| Report | Similarity | Pages |\n")
		b.WriteString("|---|---|---|\n")
		tableRows(passList)
		b.WriteString("\n")
	}

	if len(missingList) > 0 {
		b.WriteString("## \u2796 Go Output Missing\n\n")
		b.WriteString("These reports were rendered by C# but have no matching Go output file:\n\n")
		for _, name := range missingList {
			fmt.Fprintf(&b, "- %s\n", name)
		}
		b.WriteString("\n")
	}

	return b.String()
}

// ---------------------------------------------------------------------------
// Main
// ---------------------------------------------------------------------------

func main() {
	csDir := flag.String("csharp", "csharp-html-output", "Directory with C# ground-truth HTML files")
	goDir := flag.String("go", "html-output", "Directory with Go HTML files")
	outDir := flag.String("out", "html-delta", "Output directory for delta .md files")
	reportName := flag.String("report", "", "Process only this report name (without .html)")
	flag.Parse()

	if info, err := os.Stat(*csDir); err != nil || !info.IsDir() {
		fmt.Fprintf(os.Stderr, "error: C# output directory not found: %s\n", *csDir)
		os.Exit(1)
	}
	if info, err := os.Stat(*goDir); err != nil || !info.IsDir() {
		fmt.Fprintf(os.Stderr, "error: Go output directory not found: %s\n", *goDir)
		os.Exit(1)
	}

	if err := os.MkdirAll(*outDir, 0o755); err != nil {
		fmt.Fprintf(os.Stderr, "error: cannot create output directory: %v\n", err)
		os.Exit(1)
	}

	// Collect C# HTML files.
	entries, err := os.ReadDir(*csDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: cannot read C# directory: %v\n", err)
		os.Exit(1)
	}

	var csFiles []string
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		if strings.HasSuffix(strings.ToLower(e.Name()), ".html") {
			csFiles = append(csFiles, e.Name())
		}
	}
	sort.Strings(csFiles)

	if *reportName != "" {
		var filtered []string
		target := strings.ToLower(*reportName)
		for _, f := range csFiles {
			stem := strings.TrimSuffix(f, filepath.Ext(f))
			if strings.ToLower(stem) == target {
				filtered = append(filtered, f)
			}
		}
		if len(filtered) == 0 {
			fmt.Fprintf(os.Stderr, "error: no C# HTML file matching %q\n", *reportName)
			os.Exit(1)
		}
		csFiles = filtered
	}

	var results []result

	for _, csFile := range csFiles {
		name := strings.TrimSuffix(csFile, filepath.Ext(csFile))
		csPath := filepath.Join(*csDir, csFile)
		goPath := filepath.Join(*goDir, csFile)

		if _, err := os.Stat(goPath); os.IsNotExist(err) {
			fmt.Printf("  MISSING  %s\n", name)
			results = append(results, result{name: name, delta: nil})
			// Write stub delta doc.
			md := fmt.Sprintf("# %s\n\n**Status:** \u2796 MISSING -- Go output file not found.\n", name)
			os.WriteFile(filepath.Join(*outDir, name+".md"), []byte(md), 0o644)
			continue
		}

		csPageCount, csPages, err1 := parseHTML(csPath)
		goPageCount, goPages, err2 := parseHTML(goPath)
		if err1 != nil || err2 != nil {
			errMsg := ""
			if err1 != nil {
				errMsg = err1.Error()
			} else {
				errMsg = err2.Error()
			}
			fmt.Printf("  ERROR    %s: %s\n", name, errMsg)
			results = append(results, result{name: name, delta: nil})
			continue
		}

		d := analyse(name, csPages, goPages)
		st := deltaStatus(d)
		sim := fmt.Sprintf("%.0f%%", d.similarity*100)
		fmt.Printf("  %s  %-55s similarity=%s  pages=%d/%d\n", st, name, sim, csPageCount, goPageCount)

		if deltaStatus(d) != statusPass {
			md := renderMD(d)
			os.WriteFile(filepath.Join(*outDir, name+".md"), []byte(md), 0o644)
		}
		results = append(results, result{name: name, delta: d})
	}

	// Write summary index.
	if *reportName == "" {
		summaryMD := renderSummary(results)
		summaryPath := filepath.Join(*outDir, "README.md")
		os.WriteFile(summaryPath, []byte(summaryMD), 0o644)
		fmt.Printf("\nSummary written to %s\n", summaryPath)
	}

	passN, minorN, failN, missN := 0, 0, 0, 0
	for _, r := range results {
		if r.delta == nil {
			missN++
		} else {
			switch deltaStatus(r.delta) {
			case statusPass:
				passN++
			case statusMinor:
				minorN++
			default:
				failN++
			}
		}
	}
	fmt.Printf("\n%d pass, %d minor, %d fail, %d missing -- delta docs in %q\n",
		passN, minorN, failN, missN, *outDir+"/")
}
