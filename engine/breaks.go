package engine

import (
	"github.com/andrewloable/go-fastreport/band"
	"github.com/andrewloable/go-fastreport/preview"
)

// ── Soft page breaks ──────────────────────────────────────────────────────────

// BreakBand handles a soft page break for b.
// It renders as much of the band as fits in FreeSpace, then starts a new
// page and renders the remainder.
//
// The Go implementation is a simplified version: when a band overflows, it
// either places the whole band on the next page (if CanBreak=false) or splits
// it at FreeSpace using Break().
func (e *ReportEngine) BreakBand(b *band.BandBase) {
	height := e.CalcBandHeight(b)
	free := e.freeSpace

	if b.CanBreak() && free > 0 && height > free {
		// Attempt to split the band at FreeSpace height.
		// In a full implementation, Break() would split text objects etc.
		// Here we render the top portion and defer the rest.

		// Render top portion (up to free space).
		if e.preparedPages != nil {
			pb := &preview.PreparedBand{
				Name:   b.Name(),
				Top:    e.curY,
				Height: free,
			}
			_ = e.preparedPages.AddBand(pb)
		}
		e.AdvanceY(free)

		// Start new page.
		e.startNewPageForCurrent()

		// Render remainder.
		remainder := height - free
		if remainder > 0 && e.preparedPages != nil {
			pb := &preview.PreparedBand{
				Name:   b.Name(),
				Top:    e.curY,
				Height: remainder,
			}
			_ = e.preparedPages.AddBand(pb)
			e.AdvanceY(remainder)
		}
	} else {
		// Cannot break or already on new page — put the whole band on the next page.
		e.startNewPageForCurrent()
		if e.preparedPages != nil {
			pb := &preview.PreparedBand{
				Name:   b.Name(),
				Top:    e.curY,
				Height: height,
			}
			_ = e.preparedPages.AddBand(pb)
		}
		e.AdvanceY(height)
	}
}

// ── Hard page breaks ──────────────────────────────────────────────────────────

// pageBreaker is a local interface matching objects that expose PageBreak().
type pageBreaker interface {
	PageBreak() bool
	Top() float32
}

// BandHasHardPageBreaks returns true when any object in b has PageBreak=true.
func (e *ReportEngine) BandHasHardPageBreaks(b *band.BandBase) bool {
	for i := 0; i < b.Objects().Len(); i++ {
		obj := b.Objects().Get(i)
		if pb, ok := obj.(pageBreaker); ok && pb.PageBreak() {
			return true
		}
	}
	return false
}

// SplitHardPageBreaks splits b at the positions of objects with PageBreak=true
// and returns the resulting band parts. Each part after the first has
// StartNewPage=true, triggering a page break when rendered.
//
// If there are no hard breaks, a single-element slice containing the original
// band is returned.
func (e *ReportEngine) SplitHardPageBreaks(b *band.BandBase) []*band.BandBase {
	// Collect break positions from objects that have PageBreak=true.
	type breakPart struct {
		name    string
		topY    float32
		height  float32
		newPage bool
	}

	var breaks []float32
	for i := 0; i < b.Objects().Len(); i++ {
		obj := b.Objects().Get(i)
		if pb, ok := obj.(pageBreaker); ok && pb.PageBreak() {
			breaks = append(breaks, pb.Top())
		}
	}

	if len(breaks) == 0 {
		return []*band.BandBase{b}
	}

	var parts []*band.BandBase
	offsetY := float32(0)
	for idx, breakY := range breaks {
		part := band.NewBandBase()
		part.SetName(b.Name())
		part.SetHeight(breakY - offsetY)
		part.SetVisible(true)
		if idx > 0 {
			part.SetStartNewPage(true)
		}
		parts = append(parts, part)
		offsetY = breakY
	}
	// Trailing part after last break.
	last := band.NewBandBase()
	last.SetName(b.Name())
	last.SetHeight(b.Height() - offsetY)
	last.SetVisible(true)
	last.SetStartNewPage(true)
	parts = append(parts, last)

	return parts
}

// ShowBandWithPageBreaks renders band b, handling any hard page breaks.
// This wraps showBand (in pages.go) to intercept bands that contain
// objects with PageBreak=true.
func (e *ReportEngine) ShowBandWithPageBreaks(b *band.BandBase) {
	if e.BandHasHardPageBreaks(b) {
		for _, part := range e.SplitHardPageBreaks(b) {
			if part.StartNewPage() {
				e.startNewPageForCurrent()
			}
			e.showBand(part)
		}
	} else {
		e.showBand(b)
	}
}
