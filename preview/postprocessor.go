package preview

import (
	"fmt"
	"math"
	"sort"
	"strings"
)

// Postprocessor scans all PreparedPages after the engine run completes and
// performs deferred replacements:
//   - [TotalPages] / [PageCount] / [TOTALPAGES#] → actual page count string
//   - [Page] / [PAGE#] → current page number string
//   - Text object merging (MergeMode Vertical/Horizontal) within each band
//   - Text duplicates (Clear/Hide/Merge) per-object-name grouping
//
// This is the Go equivalent of FastReport.Preview.PreparedPagePostprocessor.
// C# source: FastReport.Base/Preview/PreparedPagePostprocessor.cs
type Postprocessor struct {
	pp *PreparedPages
	// unlimitedBandIdx is the sequential counter for PostProcessBandUnlimited.
	// Mirrors C# PreparedPagePostprocessor.iBand field (PostprocessUnlimited path).
	unlimitedBandIdx int
}

// NewPostprocessor creates a Postprocessor that will operate on pp.
func NewPostprocessor(pp *PreparedPages) *Postprocessor {
	return &Postprocessor{pp: pp}
}

// Process applies all deferred post-processing steps to the prepared pages.
// Call this once after the engine has finished its final pass.
func (p *Postprocessor) Process() {
	total := p.pp.Count()
	totalStr := fmt.Sprintf("%d", total)

	// Pass 1: macro substitution (TotalPages, Page).
	for i := range p.pp.pages {
		pg := p.pp.pages[i]
		pageStr := fmt.Sprintf("%d", pg.PageNo)
		for j := range pg.Bands {
			band := pg.Bands[j]
			for k := range band.Objects {
				obj := &band.Objects[k]
				if obj.Text == "" {
					continue
				}
				t := obj.Text
				t = strings.ReplaceAll(t, "[TotalPages]", totalStr)
				t = strings.ReplaceAll(t, "[PageCount]", totalStr)
				t = strings.ReplaceAll(t, "[Page]", pageStr)
				t = strings.ReplaceAll(t, "[PAGE#]", pageStr)
				t = strings.ReplaceAll(t, "[TOTALPAGES#]", totalStr)
				obj.Text = t
			}
		}
	}

	// Pass 2: text object merging (MergeMode Vertical/Horizontal).
	// C# source: PreparedPagePostprocessor.cs MergeTextObjects / CollectMergedTextObjects.
	p.mergeTextObjects()

	// Pass 3: text duplicates (Clear / Hide / Merge).
	// C# source: PreparedPagePostprocessor.cs ProcessDuplicates / CloseDuplicates.
	p.processDuplicates()
}

// ── MergeTextObjects ──────────────────────────────────────────────────────────

// mergeKey identifies a band within the page collection for grouping merge candidates.
type mergeKey struct {
	pageIdx int
	bandIdx int
}

// mergeEntry holds a reference to a mergeable text object and its absolute position.
// C# equivalent: TextObject references in mergedTextObjects dictionary.
type mergeEntry struct {
	pageIdx int
	bandIdx int
	objIdx  int
	// absLeft / absTop are absolute positions on the page.
	absLeft float32
	absTop  float32
}

// mergeTextObjects collects all text objects with MergeMode != MergeModeNone,
// groups them by band, sorts each group by (absLeft, absTop), then performs
// vertical and horizontal merging within each group.
// C# source: PreparedPagePostprocessor.cs MergeTextObjects() and CollectMergedTextObjects().
func (p *Postprocessor) mergeTextObjects() {
	// groups maps band identity → list of merge candidates.
	groups := make(map[mergeKey][]mergeEntry)

	for pi := range p.pp.pages {
		pg := p.pp.pages[pi]
		for bi := range pg.Bands {
			band := pg.Bands[bi]
			for oi := range band.Objects {
				obj := &band.Objects[oi]
				if obj.Kind != ObjectTypeText {
					continue
				}
				if obj.MergeMode == MergeModeNone {
					continue
				}
				key := mergeKey{pageIdx: pi, bandIdx: bi}
				entry := mergeEntry{
					pageIdx: pi,
					bandIdx: bi,
					objIdx:  oi,
					absLeft: band.Left + obj.Left,
					absTop:  band.Top + obj.Top,
				}
				groups[key] = append(groups[key], entry)
			}
		}
	}

	for key, entries := range groups {
		// Sort by absLeft then absTop, matching C# sort delegate.
		sort.Slice(entries, func(i, j int) bool {
			ai := entries[i]
			aj := entries[j]
			if ai.absLeft != aj.absLeft {
				return ai.absLeft < aj.absLeft
			}
			return ai.absTop < aj.absTop
		})

		// Vertical merge pass.
		entries = p.mergeTextObjectsInBand(key, entries)
		// Horizontal merge pass (C# calls MergeTextObjectsInBand twice).
		p.mergeTextObjectsInBand(key, entries)
	}
}

// mergeTextObjectsInBand performs a single O(n²) merge pass over the entries
// list. For each pair (j, i) where j > i, if obj[j] can be merged into obj[i],
// obj[j] is disposed (zeroed) and removed from the list.
// Returns the updated entries list.
// C# source: PreparedPagePostprocessor.cs MergeTextObjectsInBand().
func (p *Postprocessor) mergeTextObjectsInBand(key mergeKey, entries []mergeEntry) []mergeEntry {
	band := p.pp.pages[key.pageIdx].Bands[key.bandIdx]
	for i := 0; i < len(entries); i++ {
		j := i + 1
		for j < len(entries) {
			obj := &band.Objects[entries[j].objIdx]
			obj2 := &band.Objects[entries[i].objIdx]
			if p.mergeObjects(entries[j], obj, entries[i], obj2, band) {
				// Dispose obj (j): zero its dimensions so it is invisible.
				obj.Width = 0
				obj.Height = 0
				obj.Text = ""
				// Remove j from the list.
				entries = append(entries[:j], entries[j+1:]...)
			} else {
				j++
			}
		}
	}
	return entries
}

// mergeObjects tries to merge obj into obj2 (extending obj2 to cover obj).
// Returns true and mutates obj2 if merging is possible.
// C# source: PreparedPagePostprocessor.cs Merge(TextObject obj, TextObject obj2).
func (p *Postprocessor) mergeObjects(eObj mergeEntry, obj *PreparedObject, eObj2 mergeEntry, obj2 *PreparedObject, band *PreparedBand) bool {
	if obj2.Text != obj.Text {
		return false
	}

	// Absolute bounds of obj.
	objLeft := band.Left + obj.Left
	objTop := band.Top + obj.Top
	objRight := objLeft + obj.Width
	objBottom := objTop + obj.Height
	if obj.Width < 0 {
		objLeft, objRight = objRight, objLeft
	}
	if obj.Height < 0 {
		objTop, objBottom = objBottom, objTop
	}

	// Absolute bounds of obj2.
	obj2Left := band.Left + obj2.Left
	obj2Top := band.Top + obj2.Top
	obj2Right := obj2Left + obj2.Width
	obj2Bottom := obj2Top + obj2.Height
	if obj2.Width < 0 {
		obj2Left, obj2Right = obj2Right, obj2Left
	}
	if obj2.Height < 0 {
		obj2Top, obj2Bottom = obj2Bottom, obj2Top
	}

	// Check vertical merge: same width and left, vertically adjacent.
	// C# source: PreparedPagePostprocessor.cs lines 183–196.
	if obj.MergeMode&MergeModeVertical != 0 && obj2.MergeMode&MergeModeVertical != 0 &&
		isEqualWithInaccuracy(obj2Right-obj2Left, objRight-objLeft) &&
		isEqualWithInaccuracy(obj2Left, objLeft) {
		if isEqualWithInaccuracy(obj2Bottom, objTop) {
			obj2.Height += obj.Height
			return true
		} else if isEqualWithInaccuracy(obj2Top, objBottom) {
			obj2.Height += obj.Height
			obj2.Top -= obj.Height
			return true
		}
	}

	// Check horizontal merge: same height and top, horizontally adjacent.
	// C# source: PreparedPagePostprocessor.cs lines 198–212.
	if obj.MergeMode&MergeModeHorizontal != 0 && obj2.MergeMode&MergeModeHorizontal != 0 &&
		isEqualWithInaccuracy(obj2Bottom-obj2Top, objBottom-objTop) &&
		isEqualWithInaccuracy(obj2Top, objTop) {
		if isEqualWithInaccuracy(obj2Right, objLeft) {
			obj2.Width += obj.Width
			return true
		} else if isEqualWithInaccuracy(obj2Left, objRight) {
			obj2.Width += obj.Width
			obj2.Left -= obj.Width
			return true
		}
	}

	return false
}

// isEqualWithInaccuracy returns true if the two float32 values differ by less
// than 0.01. C# source: PreparedPagePostprocessor.cs IsEqualWithInaccuracy().
func isEqualWithInaccuracy(a, b float32) bool {
	return math.Abs(float64(a-b)) < 0.01
}

// dupEntry holds a reference to an object within the page/band hierarchy for
// duplicate processing.
type dupEntry struct {
	pageIdx int
	bandIdx int
	objIdx  int
	// absTop is the absolute vertical position of the object on the page (pixels).
	absTop float32
	// absBottom is absTop + Height.
	absBottom float32
}

// processDuplicates scans all text objects and groups consecutive same-text,
// same-name, vertically adjacent objects, then applies the configured mode.
func (p *Postprocessor) processDuplicates() {
	// groups maps object name → ordered slice of duplicate candidate entries.
	groups := make(map[string][]dupEntry)

	for pi := range p.pp.pages {
		pg := p.pp.pages[pi]
		for bi := range pg.Bands {
			band := pg.Bands[bi]
			for oi := range band.Objects {
				obj := &band.Objects[oi]
				if obj.Kind != ObjectTypeText || obj.Name == "" {
					continue
				}
				if obj.Duplicates == DuplicatesShow {
					continue
				}
				absTop := band.Top + obj.Top
				entry := dupEntry{
					pageIdx:   pi,
					bandIdx:   bi,
					objIdx:    oi,
					absTop:    absTop,
					absBottom: absTop + obj.Height,
				}
				groups[obj.Name] = append(groups[obj.Name], entry)
			}
		}
	}

	// For each name group, find runs of duplicates and apply the mode.
	// Every value in groups has len >= 1 by construction (built via append),
	// so no empty-slice guard is needed here.
	for _, entries := range groups {
		p.processGroup(entries)
	}
}

// processGroup finds runs of consecutive duplicate objects (same text, adjacent
// vertically) within entries and applies the duplicate mode.
func (p *Postprocessor) processGroup(entries []dupEntry) {
	// Run-length encode into consecutive groups.
	runStart := 0
	for i := 1; i <= len(entries); i++ {
		endRun := i == len(entries)
		if !endRun {
			prev := entries[i-1]
			curr := entries[i]
			prevObj := p.obj(prev)
			currObj := p.obj(curr)
			// Same text and vertically adjacent (within 0.5 px).
			adjacent := math.Abs(float64(curr.absTop-prev.absBottom)) <= 0.5
			sameText := currObj.Text == prevObj.Text
			if adjacent && sameText {
				continue // extend the current run
			}
		}

		// Close the run [runStart, i).
		run := entries[runStart:i]
		if len(run) > 1 {
			mode := p.obj(run[0]).Duplicates
			p.applyDuplicateMode(run, mode)
		}
		runStart = i
	}
}

// applyDuplicateMode applies Clear/Hide/Merge to a run of duplicate objects.
func (p *Postprocessor) applyDuplicateMode(run []dupEntry, mode DuplicatesMode) {
	switch mode {
	case DuplicatesClear:
		// Keep first, clear text in all others.
		for i := 1; i < len(run); i++ {
			p.obj(run[i]).Text = ""
		}

	case DuplicatesHide:
		// Keep first, mark others as hidden by clearing text and zeroing size.
		for i := 1; i < len(run); i++ {
			o := p.obj(run[i])
			o.Text = ""
			o.Width = 0
			o.Height = 0
		}

	case DuplicatesMerge:
		// Stretch the first object to cover the full span, hide the rest.
		first := p.obj(run[0])
		last := run[len(run)-1]
		lastObj := p.obj(last)
		// New height = distance from first top to last bottom.
		newBottom := last.absTop + lastObj.Height
		first.Height = newBottom - run[0].absTop
		// Hide all others.
		for i := 1; i < len(run); i++ {
			o := p.obj(run[i])
			o.Text = ""
			o.Width = 0
			o.Height = 0
		}
	}
}

// obj is a convenience helper that returns a pointer to the PreparedObject
// identified by a dupEntry.
func (p *Postprocessor) obj(e dupEntry) *PreparedObject {
	return &p.pp.pages[e.pageIdx].Bands[e.bandIdx].Objects[e.objIdx]
}

// ProcessUnlimited runs the duplicates-clear pass restricted to a single
// unlimited-height page. Mirrors C# PreparedPagePostprocessor.PostprocessUnlimited.
// Resets the unlimitedBandIdx counter so PostProcessBandUnlimited starts from 0.
func (p *Postprocessor) ProcessUnlimited(pageIdx int) {
	if pageIdx < 0 || pageIdx >= p.pp.Count() {
		return
	}
	pg := p.pp.pages[pageIdx]

	// Build duplicates groups for this page only.
	groups := make(map[string][]dupEntry)
	for bi := range pg.Bands {
		band := pg.Bands[bi]
		for oi := range band.Objects {
			obj := &band.Objects[oi]
			if obj.Kind != ObjectTypeText || obj.Name == "" {
				continue
			}
			if obj.Duplicates == DuplicatesShow {
				continue
			}
			absTop := band.Top + obj.Top
			entry := dupEntry{
				pageIdx:   pageIdx,
				bandIdx:   bi,
				objIdx:    oi,
				absTop:    absTop,
				absBottom: absTop + obj.Height,
			}
			groups[obj.Name] = append(groups[obj.Name], entry)
		}
	}

	for _, entries := range groups {
		p.processGroup(entries)
	}

	// Reset sequential counter for PostProcessBandUnlimited calls that follow.
	p.unlimitedBandIdx = 0
}

// PostProcessBandUnlimited advances the sequential band counter and returns
// the band pointer unchanged.
// Mirrors C# PreparedPagePostprocessor.PostProcessBandUnlimitedPage.
func (p *Postprocessor) PostProcessBandUnlimited(band *PreparedBand) *PreparedBand {
	p.unlimitedBandIdx++
	return band
}
