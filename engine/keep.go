package engine

import "github.com/andrewloable/go-fastreport/band"

// Keep-together state fields are stored on ReportEngine (see engine.go):
//
//	keeping        bool
//	keepPosition   int     // band index in current page at start of keep
//	keepCurX       float32 // CurX at start of keep
//	keepCurY       float32 // CurY at start of keep
//	keepDeltaY     float32 // Y distance accumulated while keeping
//
// These are declared in engine.go alongside the other engine fields.

// ── Internal helpers ──────────────────────────────────────────────────────────

// startKeepBand starts the keep mechanism for a specific band.
// If keeping is already active or the band is on its very first row and
// FirstRowStartsNewPage is false, keeping is not started (to avoid generating
// an empty first page).
func (e *ReportEngine) startKeepBand(b *band.BandBase) {
	if e.keeping {
		return
	}
	if b != nil && b.AbsRowNo() == 1 && !b.StartNewPage() {
		return
	}
	e.keeping = true
	pp := e.preparedPages
	if pp != nil {
		e.keepPosition = pp.CurPosition()
	}
	e.keepCurX = e.curX
	e.keepCurY = e.curY
}

// cutObjects removes the kept bands from the current page into temporary storage
// and rewinds CurY to the keep start position.
func (e *ReportEngine) cutObjects() {
	e.keepCurX = e.curX
	e.keepDeltaY = e.curY - e.keepCurY
	if e.preparedPages != nil {
		e.preparedPages.CutObjects(e.keepPosition)
	}
	// Rewind to keep start Y so the new page starts fresh.
	e.curY = e.keepCurY
	e.freeSpace += e.keepDeltaY
}

// pasteObjects pastes the kept bands onto the current (new) page and ends keeping.
func (e *ReportEngine) pasteObjects() {
	if e.preparedPages != nil {
		dy := e.curY - e.keepCurY
		e.preparedPages.PasteObjects(e.curX-e.keepCurX, dy)
	}
	e.EndKeep()
	e.curY += e.keepDeltaY
	e.freeSpace -= e.keepDeltaY
}

// ── Public API ────────────────────────────────────────────────────────────────

// IsKeeping returns true when the keep-together mechanism is active.
func (e *ReportEngine) IsKeeping() bool { return e.keeping }

// KeepCurY returns the Y position at which the current keep block started.
func (e *ReportEngine) KeepCurY() float32 { return e.keepCurY }

// StartKeep starts the keep-together mechanism.
// Bands printed between StartKeep and EndKeep will be moved to a new page
// together if they don't fit on the current page.
func (e *ReportEngine) StartKeep() {
	e.startKeepBand(nil)
}

// EndKeep ends the keep-together mechanism without moving bands.
func (e *ReportEngine) EndKeep() {
	if e.keeping {
		e.keeping = false
	}
}

// CheckKeepTogether is called when a page break is triggered while keeping is
// active.  It cuts the kept bands from the current page and re-pastes them on
// the new page (which the caller is expected to have already started).
func (e *ReportEngine) CheckKeepTogether() {
	if !e.keeping {
		return
	}
	e.cutObjects()
}

// FinishKeepTogether is called after a new page has been started (from
// CheckKeepTogether) to paste the kept bands at the new position.
func (e *ReportEngine) FinishKeepTogether() {
	if len(e.preparedPages.CutBands()) == 0 {
		return
	}
	e.pasteObjects()
}
