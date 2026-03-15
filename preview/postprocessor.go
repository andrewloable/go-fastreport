package preview

import (
	"fmt"
	"math"
	"strings"
)

// Postprocessor scans all PreparedPages after the engine run completes and
// performs deferred replacements:
//   - [TotalPages] / [PageCount] → actual page count string
//   - Text duplicates (Clear/Hide/Merge) per-object-name grouping
//
// This is the Go equivalent of FastReport.Preview.PreparedPagePostprocessor.
type Postprocessor struct {
	pp *PreparedPages
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
				obj.Text = t
			}
		}
	}

	// Pass 2: text duplicates (Clear / Hide / Merge).
	p.processDuplicates()
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
	for _, entries := range groups {
		if len(entries) == 0 {
			continue
		}
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
