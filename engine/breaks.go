package engine

import (
	"github.com/andrewloable/go-fastreport/band"
	"github.com/andrewloable/go-fastreport/preview"
	"github.com/andrewloable/go-fastreport/report"
)

// ── Soft page breaks ──────────────────────────────────────────────────────────

// BreakBand handles a soft page break for b.
// It renders as much of the band as fits in FreeSpace on the current page,
// then starts a new page and renders the remainder.
//
// When CanBreak=true the band's objects are partitioned at the break line:
//   - Objects entirely above the line stay on the current page.
//   - Objects that cross the line are clipped (top fragment stays, bottom moves).
//   - Objects entirely below the line move to the new page with adjusted tops.
//
// When CanBreak=false or the band fits, the whole band moves to the next page.
func (e *ReportEngine) BreakBand(b *band.BandBase) {
	height := e.CalcBandHeight(b)
	free := e.freeSpace

	if b.CanBreak() && free > 0 && height > free {
		// ── determine break line ──────────────────────────────────────────
		// The break line starts at freeSpace. We then lower it to avoid
		// cutting through non-breakable objects.
		breakLine := free
		// Iterate objects: if a non-breakable object crosses breakLine, pull
		// breakLine down to that object's top.
		changed := true
		for changed {
			changed = false
			for i := 0; i < b.Objects().Len(); i++ {
				obj := b.Objects().Get(i)
				top, bottom := objTopBottom(obj)
				if top < breakLine && bottom > breakLine {
					// Object crosses the line.
					if !objCanBreak(obj) {
						if top < breakLine {
							breakLine = top
							changed = true
							break
						}
					}
				}
			}
		}

		// ── build top PreparedBand ────────────────────────────────────────
		if e.preparedPages != nil {
			pbTop := &preview.PreparedBand{
				Name:   b.Name(),
				Top:    e.curY,
				Height: breakLine,
			}
			e.splitPopulateTop(b.Objects(), pbTop, breakLine)
			_ = e.preparedPages.AddBand(pbTop)
		}
		e.AdvanceY(breakLine)

		// ── start new page ────────────────────────────────────────────────
		e.startNewPageForCurrent()

		// ── build remainder PreparedBand ──────────────────────────────────
		remainder := height - breakLine
		if remainder > 0 && e.preparedPages != nil {
			pbRem := &preview.PreparedBand{
				Name:   b.Name(),
				Top:    e.curY,
				Height: remainder,
			}
			e.splitPopulateBottom(b.Objects(), pbRem, breakLine)
			_ = e.preparedPages.AddBand(pbRem)
			e.AdvanceY(remainder)
		}
	} else {
		// Cannot break or already fits — move whole band to the next page.
		e.startNewPageForCurrent()
		if e.preparedPages != nil {
			pb := &preview.PreparedBand{
				Name:   b.Name(),
				Top:    e.curY,
				Height: height,
			}
			e.populateBandObjects(b, pb)
			_ = e.preparedPages.AddBand(pb)
		}
		e.AdvanceY(height)
	}
}

// objTopBottom returns the Top and Bottom (Top+Height) of any report object
// that exposes Top() and Height() via the ComponentBase interface.
func objTopBottom(obj report.Base) (top, bottom float32) {
	type positioned interface {
		Top() float32
		Height() float32
	}
	if p, ok := obj.(positioned); ok {
		t := p.Top()
		return t, t + p.Height()
	}
	return 0, 0
}

// objCanBreak returns true if obj is a BreakableComponent with CanBreak=true.
func objCanBreak(obj report.Base) bool {
	type breakable interface {
		CanBreak() bool
	}
	if b, ok := obj.(breakable); ok {
		return b.CanBreak()
	}
	return false
}

// splitPopulateTop builds PreparedObjects for the portion of the band's objects
// that falls above (or at) the breakLine.
func (e *ReportEngine) splitPopulateTop(objs *report.ObjectCollection, pb *preview.PreparedBand, breakLine float32) {
	if objs == nil {
		return
	}
	for i := 0; i < objs.Len(); i++ {
		obj := objs.Get(i)
		top, bottom := objTopBottom(obj)
		if top >= breakLine {
			continue // entirely below breakLine — skip
		}
		po := e.buildPreparedObject(obj)
		if po == nil {
			continue
		}
		if bottom > breakLine {
			// Object crosses breakLine — clip height to what fits.
			po.Height = breakLine - top
		}
		pb.Objects = append(pb.Objects, *po)
	}
}

// splitPopulateBottom builds PreparedObjects for the portion of the band's objects
// that falls below the breakLine, adjusting their Top coordinates to start at 0.
func (e *ReportEngine) splitPopulateBottom(objs *report.ObjectCollection, pb *preview.PreparedBand, breakLine float32) {
	if objs == nil {
		return
	}
	for i := 0; i < objs.Len(); i++ {
		obj := objs.Get(i)
		top, bottom := objTopBottom(obj)
		if bottom <= breakLine {
			continue // entirely above breakLine — skip
		}
		po := e.buildPreparedObject(obj)
		if po == nil {
			continue
		}
		if top < breakLine {
			// Object straddles the line — its remainder starts at 0.
			po.Height = bottom - breakLine
			po.Top = 0
		} else {
			// Entirely below breakLine — shift top up by breakLine.
			po.Top = top - breakLine
		}
		pb.Objects = append(pb.Objects, *po)
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
// and returns the resulting band parts.
//
// The algorithm mirrors C# ReportEngine.SplitHardPageBreaks exactly:
// iterate objects in order; when a PageBreak object is encountered, finish
// the current part (if any) and start a new one with StartNewPage=true and
// FirstRowStartsNewPage=true. Every object (including the PageBreak one) is
// added to the current part with its Top adjusted relative to the part offset.
//
// If there are no hard breaks, a single-element slice containing the original
// band is returned.
func (e *ReportEngine) SplitHardPageBreaks(b *band.BandBase) []*band.BandBase {
	// Quick check: any PageBreak objects at all?
	hasBreak := false
	for i := 0; i < b.Objects().Len(); i++ {
		if pb, ok := b.Objects().Get(i).(pageBreaker); ok && pb.PageBreak() {
			hasBreak = true
			break
		}
	}
	if !hasBreak {
		return []*band.BandBase{b}
	}

	var parts []*band.BandBase
	var currentPart *band.BandBase
	offsetY := float32(0)

	for i := 0; i < b.Objects().Len(); i++ {
		obj := b.Objects().Get(i)

		// Check if this object triggers a page break.
		pb, isPageBreak := obj.(pageBreaker)
		if isPageBreak && pb.PageBreak() {
			// End the current part.
			if currentPart != nil {
				currentPart.SetHeight(pb.Top() - offsetY)
			}
			currentPart = nil
			offsetY = pb.Top()
		}

		// Start a new part if needed.
		if currentPart == nil {
			currentPart = band.NewBandBase()
			currentPart.SetName(b.Name())
			currentPart.SetVisible(true)
			currentPart.SetWidth(b.Width())
			if isPageBreak && pb.PageBreak() {
				currentPart.SetStartNewPage(true)
				currentPart.SetFirstRowStartsNewPage(true)
			}
			parts = append(parts, currentPart)
		}

		// Clone the object into the current part with adjusted Top.
		// We add it to the part's Objects collection so showBand renders it.
		type settableTop interface {
			Top() float32
			SetTop(float32)
		}
		if st, ok := obj.(settableTop); ok {
			st.SetTop(st.Top() - offsetY)
		}
		currentPart.Objects().Add(obj)
	}

	// Set height of the last part.
	if currentPart != nil {
		currentPart.SetHeight(b.Height() - offsetY)
	}
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
