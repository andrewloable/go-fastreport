// compare_html_semantic — Semantic comparison of Go vs C# HTML output.
//
// Parses both HTML files, extracts per-page elements (position, size, text, style),
// normalises number formatting, and compares structurally.
//
// The C# output is ground-truth (expected). The Go output is the actual.
//
// Usage:
//
//	go run ./tools/compare_html_semantic
//	go run ./tools/compare_html_semantic --report "Simple List"
package main

import (
	"flag"
	"fmt"
	"io"
	"math"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"golang.org/x/net/html"
)

// posTolerance is the maximum pixel difference that is considered a rounding
// error and therefore ignored during position/size comparison.
var posTolerance = 0.1

// ---------------------------------------------------------------------------
// HTML parsing
// ---------------------------------------------------------------------------

var (
	pageRE = regexp.MustCompile(`\bfrpage(\d+)\b`)
	numRE  = regexp.MustCompile(`-?\d+(?:\.\d+)?`)
)

func parseStyle(style string) map[string]string {
	result := map[string]string{}
	for _, part := range strings.Split(style, ";") {
		part = strings.TrimSpace(part)
		idx := strings.Index(part, ":")
		if idx < 0 {
			continue
		}
		k := strings.TrimSpace(strings.ToLower(part[:idx]))
		v := strings.TrimSpace(part[idx+1:])
		result[k] = v
	}
	return result
}

func normNum(s string) string {
	return numRE.ReplaceAllStringFunc(s, func(n string) string {
		if strings.Contains(n, ".") {
			n = strings.TrimRight(n, "0")
			n = strings.TrimRight(n, ".")
		}
		return n
	})
}

func normStyle(sd map[string]string) map[string]string {
	result := make(map[string]string, len(sd))
	for k, v := range sd {
		result[k] = normNum(v)
	}
	return result
}

// Element is a positioned element extracted from an HTML page.
type Element struct {
	Tag       string
	Cls       string
	RawStyle  string
	Text      string
	Left      string
	Top       string
	Width     string
	Height    string
	BgColor   string
	Color     string
	Font      string
	Border    string
	TextAlign string
	VertAlign string
	HasSVG    bool // true if this element contains an <svg> child
}

func newElement(tag, cls, styleStr string) *Element {
	sd := parseStyle(styleStr)
	nsd := normStyle(sd)
	font := nsd["font"]
	if font == "" {
		font = nsd["font-family"]
	}
	return &Element{
		Tag:       tag,
		Cls:       cls,
		RawStyle:  styleStr,
		Left:      nsd["left"],
		Top:       nsd["top"],
		Width:     nsd["width"],
		Height:    nsd["height"],
		BgColor:   nsd["background-color"],
		Color:     nsd["color"],
		Font:      font,
		Border:    nsd["border"],
		TextAlign: nsd["text-align"],
		VertAlign: nsd["vertical-align"],
	}
}

// parsePx extracts a float64 from a CSS value like "130.08px" or "130.08".
func parsePx(s string) float64 {
	s = strings.TrimSuffix(strings.TrimSpace(s), "px")
	v, _ := strconv.ParseFloat(s, 64)
	return v
}

// pxClose returns true if two CSS pixel values are within posTolerance.
func pxClose(a, b string) bool {
	return math.Abs(parsePx(a)-parsePx(b)) <= posTolerance
}

// roundedPosKey returns a position key with values rounded to 1 decimal place
// so that 0.01px rounding differences map to the same bucket.
func roundedPosKey(left, top string) string {
	l := math.Round(parsePx(left)*10) / 10
	t := math.Round(parsePx(top)*10) / 10
	return fmt.Sprintf("%.1f,%.1f", l, t)
}

func (e *Element) posKey() string {
	return roundedPosKey(e.Left, e.Top)
}

// sig returns a short identity tag for the element: "tag" or "tag.cssClass".
// Uses only the first CSS class to keep it concise.
func (e *Element) sig() string {
	if e.Cls != "" {
		cls := strings.Fields(e.Cls)[0]
		return e.Tag + "." + cls
	}
	return e.Tag
}

func (e *Element) String() string {
	t := e.Text
	if len(t) > 30 {
		t = t[:30]
	}
	return fmt.Sprintf("%s(%s,%s %sx%s %q)", e.sig(), e.Left, e.Top, e.Width, e.Height, t)
}

// Page is a parsed page from HTML output.
type Page struct {
	Idx      int
	Cls      string
	Width    string
	Height   string
	BgColor  string
	Elements []*Element
	Texts    []string
	SVGCount int // number of <svg> elements on this page
}

func newPage(idx int, cls, styleStr string) *Page {
	sd := parseStyle(styleStr)
	nsd := normStyle(sd)
	return &Page{
		Idx:     idx,
		Cls:     cls,
		Width:   nsd["width"],
		Height:  nsd["height"],
		BgColor: nsd["background-color"],
	}
}

// ---------------------------------------------------------------------------
// HTML extractor using golang.org/x/net/html tokenizer
// ---------------------------------------------------------------------------

func extractPages(r io.Reader) []*Page {
	tokenizer := html.NewTokenizer(r)

	var (
		pages     []*Page
		curPage   *Page
		curElem   *Element
		textParts []string
		skip      int
		tagStack  []string
	)

	flushText := func() {
		if curElem != nil && len(textParts) > 0 {
			curElem.Text = strings.Join(textParts, " ")
		}
		textParts = textParts[:0]
	}

	for {
		tt := tokenizer.Next()
		switch tt {
		case html.ErrorToken:
			// EOF or error — done.
			return pages

		case html.StartTagToken:
			tn, hasAttr := tokenizer.TagName()
			tag := string(tn)
			tagStack = append(tagStack, tag)

			if tag == "style" || tag == "script" || tag == "head" {
				skip++
				continue
			}
			if skip > 0 {
				continue
			}

			// Track SVG elements inside pages.
			if tag == "svg" && curPage != nil {
				curPage.SVGCount++
				if curElem != nil {
					curElem.HasSVG = true
				}
			}

			cls := ""
			style := ""
			if hasAttr {
				for {
					key, val, more := tokenizer.TagAttr()
					k := string(key)
					if k == "class" {
						cls = string(val)
					} else if k == "style" {
						style = string(val)
					}
					if !more {
						break
					}
				}
			}

			m := pageRE.FindStringSubmatch(cls)
			if m != nil {
				idx := 0
				for _, c := range m[1] {
					idx = idx*10 + int(c-'0')
				}
				curPage = newPage(idx, cls, style)
				pages = append(pages, curPage)
				continue
			}

			if curPage != nil && style != "" && strings.Contains(strings.ToLower(style), "left:") {
				flushText()
				curElem = newElement(tag, cls, style)
				curPage.Elements = append(curPage.Elements, curElem)
			}

		case html.EndTagToken:
			tn, _ := tokenizer.TagName()
			tag := string(tn)
			// Pop from stack if matching.
			if len(tagStack) > 0 && tagStack[len(tagStack)-1] == tag {
				tagStack = tagStack[:len(tagStack)-1]
			}
			if (tag == "style" || tag == "script" || tag == "head") && skip > 0 {
				skip--
			}
			if curElem != nil && tag == curElem.Tag {
				flushText()
				curElem = nil
			}

		case html.SelfClosingTagToken:
			tn, hasAttr := tokenizer.TagName()
			tag := string(tn)

			if skip > 0 {
				continue
			}

			cls := ""
			style := ""
			if hasAttr {
				for {
					key, val, more := tokenizer.TagAttr()
					k := string(key)
					if k == "class" {
						cls = string(val)
					} else if k == "style" {
						style = string(val)
					}
					if !more {
						break
					}
				}
			}

			m := pageRE.FindStringSubmatch(cls)
			if m != nil {
				idx := 0
				for _, c := range m[1] {
					idx = idx*10 + int(c-'0')
				}
				curPage = newPage(idx, cls, style)
				pages = append(pages, curPage)
				continue
			}

			if curPage != nil && style != "" && strings.Contains(strings.ToLower(style), "left:") {
				flushText()
				elem := newElement(tag, cls, style)
				curPage.Elements = append(curPage.Elements, elem)
			}

		case html.TextToken:
			if skip > 0 {
				continue
			}
			text := strings.TrimSpace(string(tokenizer.Text()))
			if text == "" {
				continue
			}
			if curPage != nil {
				textParts = append(textParts, text)
				curPage.Texts = append(curPage.Texts, text)
			}
		}
	}
}

func parseHTMLFile(path string) ([]*Page, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return extractPages(f), nil
}

// ---------------------------------------------------------------------------
// CSS class extraction (for style comparison)
// ---------------------------------------------------------------------------

var cssClassRE = regexp.MustCompile(`\.(s\d+)\s*\{([^}]+)\}`)

func extractCSSClasses(htmlText string) map[string]map[string]string {
	result := map[string]map[string]string{}
	for _, m := range cssClassRE.FindAllStringSubmatch(htmlText, -1) {
		clsName := m[1]
		props := parseStyle(m[2])
		result[clsName] = normStyle(props)
	}
	return result
}

// ---------------------------------------------------------------------------
// Comparison
// ---------------------------------------------------------------------------

// Delta holds all differences between expected and actual for one report.
type Delta struct {
	Name             string
	PageCountExp     int
	PageCountAct     int
	PageDimDiffs     []string
	MissingTexts     []textEntry // (page, text) present in C# but not Go
	ExtraTexts       []textEntry // (page, text) present in Go but not C#
	PositionDiffs    []string
	ElementCountDiffs []string
	TextContentDiffs  []string
	StyleDiffs        []string
	TextOrderDiffs    []string
	CSSClassDiffs     []string
	SVGDiffs          []string // SVG count mismatches per page
	ElemClassDiffs    []string // CSS class mismatches on matched elements
}

type textEntry struct {
	Page int
	Text string
}

func (d *Delta) isPass() bool {
	return d.PageCountExp == d.PageCountAct &&
		len(d.PageDimDiffs) == 0 &&
		len(d.MissingTexts) == 0 &&
		len(d.ExtraTexts) == 0 &&
		len(d.PositionDiffs) == 0 &&
		len(d.TextContentDiffs) == 0 &&
		len(d.StyleDiffs) == 0 &&
		len(d.ElementCountDiffs) == 0 &&
		len(d.TextOrderDiffs) == 0 &&
		len(d.CSSClassDiffs) == 0 &&
		len(d.SVGDiffs) == 0 &&
		len(d.ElemClassDiffs) == 0
}

func (d *Delta) severity() string {
	if d.PageCountExp != d.PageCountAct {
		return "PAGE_COUNT"
	}
	if len(d.MissingTexts) > 0 || len(d.ExtraTexts) > 0 {
		return "TEXT_CONTENT"
	}
	if len(d.TextContentDiffs) > 0 {
		return "TEXT_VALUES"
	}
	if len(d.SVGDiffs) > 0 {
		return "SVG_STRUCTURE"
	}
	if len(d.PositionDiffs) > 0 {
		return "POSITIONING"
	}
	if len(d.StyleDiffs) > 0 || len(d.ElemClassDiffs) > 0 {
		return "STYLING"
	}
	if len(d.ElementCountDiffs) > 0 {
		return "ELEMENTS"
	}
	if len(d.TextOrderDiffs) > 0 {
		return "TEXT_ORDER"
	}
	return "PASS"
}

// ---------------------------------------------------------------------------
// Sequence matching (port of difflib.SequenceMatcher)
// ---------------------------------------------------------------------------

type opCode struct {
	Tag        string // "equal", "replace", "delete", "insert"
	I1, I2     int    // range in a
	J1, J2     int    // range in b
}

// sequenceOpcodes returns opcodes describing how to turn a into b,
// matching Python's difflib.SequenceMatcher(autojunk=False).get_opcodes().
func sequenceOpcodes(a, b []string) []opCode {
	// Find longest common subsequence matching blocks using a simple approach.
	matches := longestCommonSubsequenceBlocks(a, b)
	// Add sentinel.
	matches = append(matches, [3]int{len(a), len(b), 0})

	var ops []opCode
	i, j := 0, 0
	for _, m := range matches {
		ai, bj, size := m[0], m[1], m[2]
		tag := ""
		if i < ai && j < bj {
			tag = "replace"
		} else if i < ai {
			tag = "delete"
		} else if j < bj {
			tag = "insert"
		}
		if tag != "" {
			ops = append(ops, opCode{tag, i, ai, j, bj})
		}
		if size > 0 {
			ops = append(ops, opCode{"equal", ai, ai + size, bj, bj + size})
		}
		i = ai + size
		j = bj + size
	}
	return ops
}

// longestCommonSubsequenceBlocks returns matching blocks [(i, j, size), ...].
func longestCommonSubsequenceBlocks(a, b []string) [][3]int {
	// Build index of b values for faster lookup.
	b2j := map[string][]int{}
	for j, s := range b {
		b2j[s] = append(b2j[s], j)
	}

	var blocks [][3]int
	findBlocks(a, b, b2j, 0, len(a), 0, len(b), &blocks)
	sort.Slice(blocks, func(i, j int) bool {
		if blocks[i][0] != blocks[j][0] {
			return blocks[i][0] < blocks[j][0]
		}
		return blocks[i][1] < blocks[j][1]
	})
	// Collapse adjacent equal blocks.
	var collapsed [][3]int
	for _, blk := range blocks {
		if len(collapsed) > 0 {
			last := &collapsed[len(collapsed)-1]
			if last[0]+last[2] == blk[0] && last[1]+last[2] == blk[1] {
				last[2] += blk[2]
				continue
			}
		}
		collapsed = append(collapsed, blk)
	}
	return collapsed
}

func findBlocks(a, b []string, b2j map[string][]int, alo, ahi, blo, bhi int, blocks *[][3]int) {
	// Find longest matching block in a[alo:ahi], b[blo:bhi].
	bestI, bestJ, bestSize := alo, blo, 0

	// j2len maps j -> length of longest match ending at a[i], b[j].
	j2len := map[int]int{}
	for i := alo; i < ahi; i++ {
		newJ2Len := map[int]int{}
		for _, j := range b2j[a[i]] {
			if j < blo {
				continue
			}
			if j >= bhi {
				break // b2j[s] indices are in order
			}
			k := j2len[j-1] + 1
			newJ2Len[j] = k
			if k > bestSize {
				bestI = i - k + 1
				bestJ = j - k + 1
				bestSize = k
			}
		}
		j2len = newJ2Len
	}

	if bestSize > 0 {
		if alo < bestI && blo < bestJ {
			findBlocks(a, b, b2j, alo, bestI, blo, bestJ, blocks)
		}
		*blocks = append(*blocks, [3]int{bestI, bestJ, bestSize})
		if bestI+bestSize < ahi && bestJ+bestSize < bhi {
			findBlocks(a, b, b2j, bestI+bestSize, ahi, bestJ+bestSize, bhi, blocks)
		}
	}
}

// ---------------------------------------------------------------------------
// Compare
// ---------------------------------------------------------------------------

func comparePages(name string, expPages, actPages []*Page, expCSS, actCSS map[string]map[string]string) *Delta {
	d := &Delta{
		Name:         name,
		PageCountExp: len(expPages),
		PageCountAct: len(actPages),
	}

	n := len(expPages)
	if len(actPages) < n {
		n = len(actPages)
	}

	for i := 0; i < n; i++ {
		ep := expPages[i]
		ap := actPages[i]

		// Page dimensions (with rounding tolerance).
		if !pxClose(ep.Width, ap.Width) || !pxClose(ep.Height, ap.Height) {
			d.PageDimDiffs = append(d.PageDimDiffs,
				fmt.Sprintf("Page %d: dim C#=%sx%s Go=%sx%s", i, ep.Width, ep.Height, ap.Width, ap.Height))
		}

		// Element counts.
		if len(ep.Elements) != len(ap.Elements) {
			d.ElementCountDiffs = append(d.ElementCountDiffs,
				fmt.Sprintf("Page %d: C#=%d elements, Go=%d elements", i, len(ep.Elements), len(ap.Elements)))
		}

		// Text content (set comparison).
		expTexts := map[string]struct{}{}
		for _, t := range ep.Texts {
			expTexts[t] = struct{}{}
		}
		actTexts := map[string]struct{}{}
		for _, t := range ap.Texts {
			actTexts[t] = struct{}{}
		}
		// Missing in Go.
		{
			var missing []string
			for t := range expTexts {
				if _, ok := actTexts[t]; !ok {
					missing = append(missing, t)
				}
			}
			sort.Strings(missing)
			for _, t := range missing {
				d.MissingTexts = append(d.MissingTexts, textEntry{i, t})
			}
		}
		// Extra in Go.
		{
			var extra []string
			for t := range actTexts {
				if _, ok := expTexts[t]; !ok {
					extra = append(extra, t)
				}
			}
			sort.Strings(extra)
			for _, t := range extra {
				d.ExtraTexts = append(d.ExtraTexts, textEntry{i, t})
			}
		}

		// Text sequence comparison.
		if !strSliceEqual(ep.Texts, ap.Texts) {
			ops := sequenceOpcodes(ep.Texts, ap.Texts)
			for _, op := range ops {
				if op.Tag == "equal" {
					continue
				}
				csChunk := ep.Texts[op.I1:op.I2]
				goChunk := ap.Texts[op.J1:op.J2]
				switch op.Tag {
				case "replace":
					d.TextOrderDiffs = append(d.TextOrderDiffs,
						fmt.Sprintf("Page %d: replace C#=%s -> Go=%s", i, fmtChunk(csChunk, 3), fmtChunk(goChunk, 3)))
				case "delete":
					d.TextOrderDiffs = append(d.TextOrderDiffs,
						fmt.Sprintf("Page %d: missing C#=%s", i, fmtChunk(csChunk, 3)))
				case "insert":
					d.TextOrderDiffs = append(d.TextOrderDiffs,
						fmt.Sprintf("Page %d: extra Go=%s", i, fmtChunk(goChunk, 3)))
				}
				if len(d.TextOrderDiffs) > 30 {
					d.TextOrderDiffs = append(d.TextOrderDiffs, "... (truncated)")
					break
				}
			}
		}

		// Element-by-element comparison (match by position).
		expByPos := map[string][]*Element{}
		for _, e := range ep.Elements {
			key := e.posKey()
			expByPos[key] = append(expByPos[key], e)
		}
		actByPos := map[string][]*Element{}
		for _, e := range ap.Elements {
			key := e.posKey()
			actByPos[key] = append(actByPos[key], e)
		}

		allPositions := map[string]struct{}{}
		for k := range expByPos {
			allPositions[k] = struct{}{}
		}
		for k := range actByPos {
			allPositions[k] = struct{}{}
		}
		sortedPositions := make([]string, 0, len(allPositions))
		for k := range allPositions {
			sortedPositions = append(sortedPositions, k)
		}
		sort.Strings(sortedPositions)

		for _, pos := range sortedPositions {
			expElems := expByPos[pos]
			actElems := actByPos[pos]
			minLen := len(expElems)
			if len(actElems) < minLen {
				minLen = len(actElems)
			}
			for j := 0; j < minLen; j++ {
				ee := expElems[j]
				ae := actElems[j]
				// Size comparison (with rounding tolerance).
				if !pxClose(ee.Width, ae.Width) || !pxClose(ee.Height, ae.Height) {
					d.PositionDiffs = append(d.PositionDiffs,
						fmt.Sprintf("Page %d @%s [%s]: size C#=%sx%s Go=%sx%s", i, pos, ee.sig(), ee.Width, ee.Height, ae.Width, ae.Height))
				}
				// Text at same position.
				if ee.Text != "" && ae.Text != "" && ee.Text != ae.Text {
					eeText := ee.Text
					if len(eeText) > 60 {
						eeText = eeText[:60]
					}
					aeText := ae.Text
					if len(aeText) > 60 {
						aeText = aeText[:60]
					}
					d.TextContentDiffs = append(d.TextContentDiffs,
						fmt.Sprintf("Page %d @%s [%s]: C#=%q Go=%q", i, pos, ee.sig(), eeText, aeText))
				}
				// Background color.
				if ee.BgColor != "" && ae.BgColor != "" && ee.BgColor != ae.BgColor {
					d.StyleDiffs = append(d.StyleDiffs,
						fmt.Sprintf("Page %d @%s [%s]: bg C#=%s Go=%s", i, pos, ee.sig(), ee.BgColor, ae.BgColor))
				}
			}

			// SVG structure comparison.
			for j := 0; j < minLen; j++ {
				ee := expElems[j]
				ae := actElems[j]
				if ee.HasSVG != ae.HasSVG {
					who := "Go has SVG, C# does not"
					if ee.HasSVG {
						who = "C# has SVG, Go does not"
					}
					d.SVGDiffs = append(d.SVGDiffs,
						fmt.Sprintf("Page %d @%s [%s]: %s", i, pos, ae.sig(), who))
				}
				// CSS class differences on matched elements.
				if ee.Cls != ae.Cls && len(d.ElemClassDiffs) < 50 {
					d.ElemClassDiffs = append(d.ElemClassDiffs,
						fmt.Sprintf("Page %d @%s [%s]: resolved styles differ (C# classes=%q Go classes=%q)", i, pos, ee.sig(), ee.Cls, ae.Cls))
				}
			}

			// Elements at a position in C# but not in Go.
			if len(expElems) > len(actElems) && len(d.PositionDiffs) < 50 {
				for j := len(actElems); j < len(expElems); j++ {
					d.PositionDiffs = append(d.PositionDiffs,
						fmt.Sprintf("Page %d @%s: missing in Go: %s", i, pos, expElems[j]))
				}
			}
			if len(actElems) > len(expElems) && len(d.PositionDiffs) < 50 {
				for j := len(expElems); j < len(actElems); j++ {
					d.PositionDiffs = append(d.PositionDiffs,
						fmt.Sprintf("Page %d @%s: extra in Go: %s", i, pos, actElems[j]))
				}
			}
		}

		// SVG count per page.
		if ep.SVGCount != ap.SVGCount {
			d.SVGDiffs = append(d.SVGDiffs,
				fmt.Sprintf("Page %d: SVG count C#=%d Go=%d", i, ep.SVGCount, ap.SVGCount))
		}
	}

	// Truncate large diffs.
	truncateStrSlice(&d.PositionDiffs, 50)
	truncateStrSlice(&d.TextContentDiffs, 50)
	truncateStrSlice(&d.StyleDiffs, 50)
	truncateStrSlice(&d.TextOrderDiffs, 50)
	truncateStrSlice(&d.ElementCountDiffs, 50)
	truncateTextEntries(&d.MissingTexts, 50)
	truncateTextEntries(&d.ExtraTexts, 50)
	truncateStrSlice(&d.SVGDiffs, 50)
	truncateStrSlice(&d.ElemClassDiffs, 50)

	return d
}

func strSliceEqual(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func truncateStrSlice(s *[]string, limit int) {
	if len(*s) > limit {
		nMore := len(*s) - limit
		*s = (*s)[:limit]
		*s = append(*s, fmt.Sprintf("... and %d more", nMore))
	}
}

func truncateTextEntries(s *[]textEntry, limit int) {
	if len(*s) > limit {
		nMore := len(*s) - limit
		*s = (*s)[:limit]
		*s = append(*s, textEntry{-1, fmt.Sprintf("... and %d more", nMore)})
	}
}

func fmtChunk(chunk []string, maxItems int) string {
	display := chunk
	if len(display) > maxItems {
		display = display[:maxItems]
	}
	// Format as Python-style list representation.
	parts := make([]string, len(display))
	for i, s := range display {
		parts[i] = fmt.Sprintf("'%s'", s)
	}
	return "[" + strings.Join(parts, ", ") + "]"
}

// ---------------------------------------------------------------------------
// Markdown report
// ---------------------------------------------------------------------------

func renderMD(d *Delta) string {
	var lines []string
	lines = append(lines, fmt.Sprintf("# %s", d.Name), "")

	status := "PASS"
	if !d.isPass() {
		status = fmt.Sprintf("FAIL (%s)", d.severity())
	}
	lines = append(lines, fmt.Sprintf("**Status:** %s", status))
	pageLine := fmt.Sprintf("**Pages:** C# = %d, Go = %d", d.PageCountExp, d.PageCountAct)
	if d.PageCountExp != d.PageCountAct {
		pageLine += fmt.Sprintf("  (mismatch: %+d)", d.PageCountAct-d.PageCountExp)
	}
	lines = append(lines, pageLine, "")

	if len(d.PageDimDiffs) > 0 {
		lines = append(lines, "## Page Dimension Differences", "")
		for _, s := range d.PageDimDiffs {
			lines = append(lines, fmt.Sprintf("- %s", s))
		}
		lines = append(lines, "")
	}

	if len(d.ElementCountDiffs) > 0 {
		lines = append(lines, "## Element Count Differences", "")
		for _, s := range d.ElementCountDiffs {
			lines = append(lines, fmt.Sprintf("- %s", s))
		}
		lines = append(lines, "")
	}

	if len(d.MissingTexts) > 0 {
		lines = append(lines, "## Text Present in C# but Missing from Go", "")
		for _, te := range d.MissingTexts {
			t := te.Text
			if len(t) > 80 {
				t = t[:80]
			}
			lines = append(lines, fmt.Sprintf("- Page %d: `%s`", te.Page, t))
		}
		lines = append(lines, "")
	}

	if len(d.ExtraTexts) > 0 {
		lines = append(lines, "## Text Present in Go but Absent from C#", "")
		for _, te := range d.ExtraTexts {
			t := te.Text
			if len(t) > 80 {
				t = t[:80]
			}
			lines = append(lines, fmt.Sprintf("- Page %d: `%s`", te.Page, t))
		}
		lines = append(lines, "")
	}

	if len(d.TextContentDiffs) > 0 {
		lines = append(lines, "## Text Value Differences (Same Position)", "")
		for _, s := range d.TextContentDiffs {
			lines = append(lines, fmt.Sprintf("- %s", s))
		}
		lines = append(lines, "")
	}

	if len(d.PositionDiffs) > 0 {
		lines = append(lines, "## Position / Size Differences", "")
		for _, s := range d.PositionDiffs {
			lines = append(lines, fmt.Sprintf("- %s", s))
		}
		lines = append(lines, "")
	}

	if len(d.StyleDiffs) > 0 {
		lines = append(lines, "## Style Differences", "")
		for _, s := range d.StyleDiffs {
			lines = append(lines, fmt.Sprintf("- %s", s))
		}
		lines = append(lines, "")
	}

	if len(d.TextOrderDiffs) > 0 {
		lines = append(lines, "## Text Order Differences", "")
		lines = append(lines, "```")
		for _, s := range d.TextOrderDiffs {
			lines = append(lines, s)
		}
		lines = append(lines, "```", "")
	}

	if len(d.SVGDiffs) > 0 {
		lines = append(lines, "## SVG Structure Differences", "")
		for _, s := range d.SVGDiffs {
			lines = append(lines, fmt.Sprintf("- %s", s))
		}
		lines = append(lines, "")
	}

	if len(d.ElemClassDiffs) > 0 {
		lines = append(lines, "## Element CSS Class Differences", "")
		for _, s := range d.ElemClassDiffs {
			lines = append(lines, fmt.Sprintf("- %s", s))
		}
		lines = append(lines, "")
	}

	if d.isPass() {
		lines = append(lines, "_No significant differences detected._", "")
	}

	return strings.Join(lines, "\n")
}

func renderSummary(results []result) string {
	var lines []string
	lines = append(lines, "# HTML Semantic Comparison Summary", "")
	lines = append(lines, "Semantic comparison of Go HTML output vs C# ground-truth.", "")

	var passList []result
	var failList []result
	var missingList []string
	severityCounts := map[string]int{}

	for _, r := range results {
		if r.delta == nil {
			missingList = append(missingList, r.name)
		} else if r.delta.isPass() {
			passList = append(passList, r)
		} else {
			failList = append(failList, r)
			sev := r.delta.severity()
			severityCounts[sev]++
		}
	}

	total := len(results)
	lines = append(lines, "| Category | Count |")
	lines = append(lines, "|---|---|")
	lines = append(lines, fmt.Sprintf("| Pass | %d |", len(passList)))
	lines = append(lines, fmt.Sprintf("| Fail | %d |", len(failList)))
	lines = append(lines, fmt.Sprintf("| Go missing | %d |", len(missingList)))
	lines = append(lines, fmt.Sprintf("| **Total** | **%d** |", total))
	lines = append(lines, "")

	if len(severityCounts) > 0 {
		lines = append(lines, "### Failure Breakdown", "")
		lines = append(lines, "| Severity | Count |")
		lines = append(lines, "|---|---|")
		// Sort by count descending.
		type sevEntry struct {
			sev   string
			count int
		}
		var sevs []sevEntry
		for s, c := range severityCounts {
			sevs = append(sevs, sevEntry{s, c})
		}
		sort.Slice(sevs, func(i, j int) bool {
			return sevs[i].count > sevs[j].count
		})
		for _, se := range sevs {
			lines = append(lines, fmt.Sprintf("| %s | %d |", se.sev, se.count))
		}
		lines = append(lines, "")
	}

	if len(failList) > 0 {
		lines = append(lines, "## Failures", "")
		lines = append(lines, "| Report | Severity | Pages (C#/Go) | Missing Texts | Extra Texts | Pos Diffs |")
		lines = append(lines, "|---|---|---|---|---|---|")
		for _, r := range failList {
			d := r.delta
			link := strings.ReplaceAll(r.name, " ", "%20")
			pg := fmt.Sprintf("%d/%d", d.PageCountExp, d.PageCountAct)
			if d.PageCountExp != d.PageCountAct {
				pg += " !"
			}
			miss := len(d.MissingTexts)
			extra := len(d.ExtraTexts)
			pos := len(d.PositionDiffs)
			lines = append(lines, fmt.Sprintf("| [%s](%s.md) | %s | %s | %d | %d | %d |",
				r.name, link, d.severity(), pg, miss, extra, pos))
		}
		lines = append(lines, "")
	}

	if len(passList) > 0 {
		lines = append(lines, "## Passing", "")
		lines = append(lines, "| Report | Pages |")
		lines = append(lines, "|---|---|")
		for _, r := range passList {
			lines = append(lines, fmt.Sprintf("| %s | %d |", r.name, r.delta.PageCountExp))
		}
		lines = append(lines, "")
	}

	if len(missingList) > 0 {
		lines = append(lines, "## Go Output Missing", "")
		for _, name := range missingList {
			lines = append(lines, fmt.Sprintf("- %s", name))
		}
		lines = append(lines, "")
	}

	return strings.Join(lines, "\n")
}

// ---------------------------------------------------------------------------
// Main
// ---------------------------------------------------------------------------

type result struct {
	name  string
	delta *Delta
}

func main() {
	csharpDir := flag.String("csharp", "csharp-html-output", "Directory with C# HTML output")
	goDir := flag.String("go", "html-output", "Directory with Go HTML output")
	outDir := flag.String("out", "html-delta", "Directory to write delta reports")
	report := flag.String("report", "", "Process only one report (by stem name)")
	tolerance := flag.Float64("tolerance", 0.1, "Pixel tolerance for rounding errors (default 0.1px)")
	flag.Parse()
	posTolerance = *tolerance

	if _, err := os.Stat(*csharpDir); os.IsNotExist(err) {
		fmt.Fprintf(os.Stderr, "error: not found: %s\n", *csharpDir)
		os.Exit(1)
	}
	if _, err := os.Stat(*goDir); os.IsNotExist(err) {
		fmt.Fprintf(os.Stderr, "error: not found: %s\n", *goDir)
		os.Exit(1)
	}

	if err := os.MkdirAll(*outDir, 0o755); err != nil {
		fmt.Fprintf(os.Stderr, "error: cannot create %s: %v\n", *outDir, err)
		os.Exit(1)
	}

	// List C# HTML files.
	csFiles, err := filepath.Glob(filepath.Join(*csharpDir, "*.html"))
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
	sort.Strings(csFiles)

	if *report != "" {
		var filtered []string
		for _, f := range csFiles {
			stem := strings.TrimSuffix(filepath.Base(f), filepath.Ext(f))
			if strings.EqualFold(stem, *report) {
				filtered = append(filtered, f)
			}
		}
		if len(filtered) == 0 {
			fmt.Fprintf(os.Stderr, "error: no match for \"%s\"\n", *report)
			os.Exit(1)
		}
		csFiles = filtered
	}

	var results []result

	for _, csPath := range csFiles {
		name := strings.TrimSuffix(filepath.Base(csPath), filepath.Ext(csPath))
		goPath := filepath.Join(*goDir, filepath.Base(csPath))

		if _, err := os.Stat(goPath); os.IsNotExist(err) {
			fmt.Printf("  MISSING  %s\n", name)
			results = append(results, result{name, nil})
			md := fmt.Sprintf("# %s\n\n**Status:** MISSING\n", name)
			os.WriteFile(filepath.Join(*outDir, name+".md"), []byte(md), 0o644)
			continue
		}

		expPages, err := parseHTMLFile(csPath)
		if err != nil {
			fmt.Printf("  ERROR    %s: %v\n", name, err)
			results = append(results, result{name, nil})
			continue
		}
		actPages, err := parseHTMLFile(goPath)
		if err != nil {
			fmt.Printf("  ERROR    %s: %v\n", name, err)
			results = append(results, result{name, nil})
			continue
		}

		csHTML, err := os.ReadFile(csPath)
		if err != nil {
			fmt.Printf("  ERROR    %s: %v\n", name, err)
			results = append(results, result{name, nil})
			continue
		}
		goHTML, err := os.ReadFile(goPath)
		if err != nil {
			fmt.Printf("  ERROR    %s: %v\n", name, err)
			results = append(results, result{name, nil})
			continue
		}

		expCSS := extractCSSClasses(string(csHTML))
		actCSS := extractCSSClasses(string(goHTML))

		delta := comparePages(name, expPages, actPages, expCSS, actCSS)

		if delta.isPass() {
			fmt.Printf("  PASS     %-55s pages=%d\n", name, delta.PageCountExp)
			old := filepath.Join(*outDir, name+".md")
			os.Remove(old) // Ignore error if not exists.
		} else {
			sev := delta.severity()
			miss := len(delta.MissingTexts)
			extra := len(delta.ExtraTexts)
			fmt.Printf("  FAIL     %-55s %-14s pages=%d/%d  miss=%d extra=%d\n",
				name, sev, delta.PageCountExp, delta.PageCountAct, miss, extra)
			md := renderMD(delta)
			os.WriteFile(filepath.Join(*outDir, name+".md"), []byte(md), 0o644)
		}

		results = append(results, result{name, delta})
	}

	if *report == "" {
		summary := renderSummary(results)
		os.WriteFile(filepath.Join(*outDir, "README.md"), []byte(summary), 0o644)
		fmt.Printf("\nSummary written to %s/README.md\n", *outDir)
	}

	passN, failN, missN := 0, 0, 0
	for _, r := range results {
		if r.delta == nil {
			missN++
		} else if r.delta.isPass() {
			passN++
		} else {
			failN++
		}
	}
	fmt.Printf("\n%d pass, %d fail, %d missing\n", passN, failN, missN)
}
