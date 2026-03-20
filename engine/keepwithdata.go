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

// NeedKeepFirstRow returns true if the first data row of the DataBand must be
// kept together with its header (DataHeaderBand.KeepWithData=true).
func (e *ReportEngine) NeedKeepFirstRow(dataBand *band.DataBand) bool {
	return dataBand.Header() != nil && dataBand.Header().KeepWithData()
}

// NeedKeepFirstRowGroup returns true if the first data row of a group must be
// kept with the group header (GroupHeaderBand.KeepWithData=true).
func (e *ReportEngine) NeedKeepFirstRowGroup(groupBand *band.GroupHeaderBand) bool {
	if groupBand == nil {
		return false
	}
	if groupBand.KeepWithData() {
		return true
	}
	db := groupBand.Data()
	if db != nil && db.Header() != nil && db.Header().KeepWithData() {
		return true
	}
	return false
}

// NeedKeepLastRow returns true if the last data row must be kept with footer
// bands that have KeepWithData=true.
func (e *ReportEngine) NeedKeepLastRow(dataBand *band.DataBand) bool {
	footers := e.getAllFooters(dataBand)
	return len(footers) > 0
}

// getAllFooters returns the list of footer bands that must be kept with the
// last data row. This mirrors FastReport's ReportEngine.GetAllFooters:
//
//  1. Start with the DataBand's own DataFooterBand (if any).
//  2. Walk the engine's groupStack (pushed by showGroupTree): for each ancestor
//     GroupHeaderBand that is on its last row, add its GroupFooterBand (if any).
//     Stop when an ancestor is not on its last row (matching the C# guard:
//     `if (band != dataBand && !band.IsLastRow) break`).
//  3. Strip trailing footers that have no KeepWithData flag.
func (e *ReportEngine) getAllFooters(dataBand *band.DataBand) []keepableFooter {
	var footers []keepableFooter

	// Add DataFooterBand for the data band itself.
	if ftr := dataBand.Footer(); ftr != nil {
		footers = append(footers, keepableFooter{
			band:              &ftr.HeaderFooterBandBase.BandBase,
			keepWithData:      ftr.KeepWithData(),
			repeatOnEveryPage: ftr.RepeatOnEveryPage(),
		})
	}

	// Walk the groupStack (innermost first) and collect GroupFooterBands from
	// ancestor GroupHeaderBands. Stop when a group is not on its last row.
	for _, gh := range e.groupStack {
		// C# guard: stop if this ancestor group is not on its last row.
		// (We only need the group footer if this group instance is finishing.)
		if !gh.IsLastRow() {
			break
		}
		if gftr := gh.GroupFooter(); gftr != nil {
			footers = append(footers, keepableFooter{
				band:              &gftr.HeaderFooterBandBase.BandBase,
				keepWithData:      gftr.KeepWithData(),
				repeatOnEveryPage: gftr.RepeatOnEveryPage(),
			})
		}
	}

	// Remove footers at the end that don't have KeepWithData.
	for i := len(footers) - 1; i >= 0; i-- {
		if !footers[i].keepWithData {
			footers = footers[:i]
		} else {
			break
		}
	}
	return footers
}

type keepableFooter struct {
	band              *band.BandBase
	keepWithData      bool
	repeatOnEveryPage bool
}

// GetFootersHeight returns the total height of footer bands that need to be
// kept with the last data row.
func (e *ReportEngine) GetFootersHeight(dataBand *band.DataBand) float32 {
	footers := e.getAllFooters(dataBand)
	var height float32
	for _, f := range footers {
		if !f.repeatOnEveryPage {
			height += e.GetBandHeightWithChildren(f.band)
		}
	}
	return height
}

// CheckKeepFooter checks if there is enough space for footer bands.
func (e *ReportEngine) CheckKeepFooter(dataBand *band.DataBand) {
	if e.FreeSpace() < e.GetFootersHeight(dataBand) {
		e.startNewPageForCurrent()
	} else {
		e.EndKeep()
	}
}

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
