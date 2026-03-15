package engine

import (
	"github.com/andrewloable/go-fastreport/band"
)

// keepwithdata.go implements the KeepWithData logic for DataHeaderBand and
// DataFooterBand. This mirrors FastReport's ReportEngine.KeepWithData.cs.
//
// Rules:
//   - DataHeaderBand.KeepWithData=true → the header must stay on the same page
//     as the FIRST data row. If the header was just printed and the first row
//     overflows to a new page, move the header to the new page too.
//   - DataFooterBand.KeepWithData=true → the footer must stay on the same page
//     as the LAST data row. If the footer does not fit after the last row, move
//     the last row (and any child bands) plus the footer to the next page.

// checkKeepFooterWithData checks whether a DataFooterBand fits on the current
// page together with the last rendered data row block. If it doesn't fit and
// keepWithData is set, the last row block is cut from the page and moved to a
// new page before the footer is rendered.
//
// lastRowPosition is the CurPosition() snapshot taken just before the last
// data row was rendered. It is used to identify the "last row block".
func (e *ReportEngine) checkKeepFooterWithData(ftr *band.DataFooterBand, lastRowPosition int) {
	if ftr == nil {
		return
	}
	if !ftr.KeepWithData() {
		return
	}

	footerH := e.CalcBandHeight(&ftr.HeaderFooterBandBase.BandBase)
	if footerH <= e.freeSpace {
		// Footer fits on the current page alongside the last row — nothing to do.
		return
	}

	// Footer does not fit. Cut the last row block + footer together onto a new page.
	pp := e.preparedPages
	if pp == nil {
		return
	}
	// Cut everything from lastRowPosition onwards (the last data row + any child bands).
	keepDeltaY := e.curY - e.keepCurY
	pp.CutObjects(lastRowPosition)
	oldCurY := e.curY
	e.curY = e.keepCurY
	e.freeSpace += oldCurY - e.keepCurY

	// Start a new page.
	e.startNewPageForCurrent()

	// Paste the cut bands at the new position.
	pp.PasteObjects(0, e.curY-e.keepCurY)
	e.curY += keepDeltaY
	if e.freeSpace > keepDeltaY {
		e.freeSpace -= keepDeltaY
	} else {
		e.freeSpace = 0
	}
}

// checkKeepHeaderWithData checks whether the DataHeaderBand needs to be
// moved to the new page along with the first data row after a page break.
//
// It should be called after the header has been shown but before the first
// data row is printed. If the header + first row don't fit on the current page,
// the header is relocated to the next page.
//
// headerPosition is the CurPosition() snapshot taken just before the header
// was rendered. db is the DataBand that follows the header.
func (e *ReportEngine) checkKeepHeaderWithData(hdr *band.DataHeaderBand, db *band.DataBand, headerPosition int) {
	if hdr == nil || db == nil {
		return
	}
	if !hdr.KeepWithData() {
		return
	}

	rowH := e.CalcBandHeight(&db.BandBase)
	headerH := e.CalcBandHeight(&hdr.HeaderFooterBandBase.BandBase)
	if headerH+rowH <= e.freeSpace {
		// Both header and first row fit on the current page.
		return
	}

	// Not enough space: cut the header and move it to the next page.
	pp := e.preparedPages
	if pp == nil {
		return
	}
	oldCurY := e.curY
	keepDeltaY := oldCurY - e.keepCurY
	pp.CutObjects(headerPosition)
	e.curY = e.keepCurY
	e.freeSpace += keepDeltaY

	e.startNewPageForCurrent()

	pp.PasteObjects(0, e.curY-e.keepCurY)
	e.curY += keepDeltaY
	if e.freeSpace > keepDeltaY {
		e.freeSpace -= keepDeltaY
	} else {
		e.freeSpace = 0
	}
}
