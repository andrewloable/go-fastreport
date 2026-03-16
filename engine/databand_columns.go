package engine

import (
	"github.com/andrewloable/go-fastreport/band"
	"github.com/andrewloable/go-fastreport/preview"
)

// dataBandColumnState tracks the current column position while rendering a
// multi-column DataBand. Rows are placed left-to-right (AcrossThenDown layout)
// before the Y cursor advances.
type dataBandColumnState struct {
	// colCount is the number of columns (>= 2 when multi-column mode is active).
	colCount int
	// colWidth is the pixel width of each column (= db.Width()).
	colWidth float32
	// colIdx is the index of the next column to fill (0-based).
	colIdx int
	// rowY is the Y position for the current row of columns.
	rowY float32
	// rowHeight is the maximum band height seen in the current column row.
	rowHeight float32
}

// newDataBandColumnState creates a dataBandColumnState for db if multi-column
// mode is active (Columns.Count > 1). Returns nil for single-column bands.
func newDataBandColumnState(e *ReportEngine, db *band.DataBand) *dataBandColumnState {
	cols := db.Columns()
	if cols == nil || cols.Count() <= 1 {
		return nil
	}
	return &dataBandColumnState{
		colCount: cols.Count(),
		colWidth: db.Width(),
		colIdx:   0,
		rowY:     e.curY,
	}
}

// showBandInColumn renders db into the prepared pages at the current column
// position without advancing curY prematurely. curY is only advanced when
// all columns in a column-row have been filled.
//
// It returns true if a page break occurred (the caller may need to handle
// header reprinting). The column state is updated in-place.
func (e *ReportEngine) showBandInColumn(db *band.DataBand, cs *dataBandColumnState) {
	bb := &db.BandBase
	if !bb.Visible() {
		return
	}

	height := e.CalcBandHeight(bb)
	if height <= 0 {
		return
	}

	// Check free space: if this is the first column of a new column-row and the
	// band doesn't fit, start a new page.
	if cs.colIdx == 0 && bb.FlagCheckFreeSpace && e.freeSpace < height {
		e.startNewPageForCurrent()
		cs.rowY = e.curY
	}

	// Track the tallest band in this column-row.
	if height > cs.rowHeight {
		cs.rowHeight = height
	}

	// Compute X offset for this column.
	xOffset := float32(cs.colIdx) * cs.colWidth

	if e.preparedPages != nil {
		pb := &preview.PreparedBand{
			Name:   bb.Name(),
			Top:    cs.rowY,
			Height: height,
		}
		// Populate objects, then shift each object's Left by the column X offset.
		e.populateBandObjects(bb, pb)
		for i := range pb.Objects {
			pb.Objects[i].Left += xOffset
		}
		_ = e.preparedPages.AddBand(pb)
	}

	// Advance to the next column. When all columns are filled, advance Y.
	cs.colIdx++
	if cs.colIdx >= cs.colCount {
		// All columns in this row are filled — advance the vertical cursor.
		e.AdvanceY(cs.rowHeight)
		cs.rowY = e.curY
		cs.colIdx = 0
		cs.rowHeight = 0
	}

	// Fire events after placing the band.
	bb.FireAfterLayout()
	bb.FireAfterPrint()
}

// flushColumnRow flushes any partially-filled column row at the end of iteration.
// If colIdx > 0, some columns were filled but the row was not complete; we
// advance curY by the recorded rowHeight.
func (e *ReportEngine) flushColumnRow(cs *dataBandColumnState) {
	if cs == nil || cs.colIdx == 0 {
		return
	}
	e.AdvanceY(cs.rowHeight)
	cs.rowY = e.curY
	cs.colIdx = 0
	cs.rowHeight = 0
}
