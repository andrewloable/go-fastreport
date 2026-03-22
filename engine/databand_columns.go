package engine

import (
	"github.com/andrewloable/go-fastreport/band"
	"github.com/andrewloable/go-fastreport/data"
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

// rowPositioner is a local interface for data sources that support direct row positioning.
type rowPositioner interface{ SetCurrentRowNo(int) }

// renderDownThenAcross renders a multi-column DataBand with DownThenAcross layout.
// Each column is filled top-to-bottom before moving to the next column.
//
// Mirrors C# RenderBandDownThenAcross (ReportEngine.DataBands.cs lines 301-457):
//   - Pre-computes heights for all rows.
//   - Determines rowsPerColumn = ceil(rowCount / colCount), bounded by MinRowCount.
//   - If max column height > FreeSpace: renders rows page-by-page with column/page breaks.
//   - Else: renders all rows in columns, advancing X each rowsPerColumn rows.
func (e *ReportEngine) renderDownThenAcross(db *band.DataBand, rowCount int) {
	ds := db.DataSourceRef()
	cols := db.Columns()
	colCount := cols.Count()
	positions := cols.Positions()

	// Get the current data source row as the starting row index.
	saveRow := 0
	if fullDS, ok := ds.(data.DataSource); ok {
		saveRow = fullDS.CurrentRowNo()
	}

	// Step 1: Pre-compute heights for all rows (C# lines 223-227).
	heights := make([]float32, rowCount)
	for i := 0; i < rowCount; i++ {
		if rp, ok := ds.(rowPositioner); ok {
			rp.SetCurrentRowNo(saveRow + i)
		}
		if fullDS, ok := ds.(data.DataSource); ok && e.report != nil {
			e.report.SetCalcContext(fullDS)
		}
		heights[i] = e.CalcBandHeight(&db.BandBase)
	}
	// Restore to start row.
	if rp, ok := ds.(rowPositioner); ok {
		rp.SetCurrentRowNo(saveRow)
	}

	// Step 2: Determine rows per column (C# lines 307-309).
	rowsPerColumn := (rowCount + colCount - 1) / colCount // ceil division
	if cols.MinRowCount > 0 && rowsPerColumn < cols.MinRowCount {
		rowsPerColumn = cols.MinRowCount
	}

	// Step 3: Compute max column height (C# lines 312-327).
	maxHeight := float32(0)
	for col := 0; col < colCount; col++ {
		var colH float32
		for row := 0; row < rowsPerColumn; row++ {
			rowIdx := row + col*rowsPerColumn
			if rowIdx >= rowCount {
				break
			}
			colH += heights[rowIdx]
		}
		if colH > maxHeight {
			maxHeight = colH
		}
	}

	saveCurX := e.curX
	startColumnY := e.curY
	colIdx := 0

	if maxHeight > e.freeSpace {
		// Not enough space: render rows down-then-across with column/page breaks.
		// Mirrors C# lines 338-390.
		for i := 0; i < rowCount; i++ {
			if rp, ok := ds.(rowPositioner); ok {
				rp.SetCurrentRowNo(saveRow + i)
			}
			if fullDS, ok := ds.(data.DataSource); ok && e.report != nil {
				e.report.SetCalcContext(fullDS)
			}

			// Set column X position.
			if colIdx < len(positions) {
				e.curX = positions[colIdx] + saveCurX
			} else {
				e.curX = float32(colIdx)*db.Width() + saveCurX
			}

			// Check if this row fits in the remaining space.
			if heights[i] > e.freeSpace && i != 0 {
				colIdx++
				if colIdx >= colCount {
					// No more columns: start a new page and render remaining rows.
					colIdx = 0
					e.curX = saveCurX
					e.startNewPageForCurrent()
					startColumnY = e.curY
					// Render remaining rows via recursion (C# calls this method again).
					if rp, ok := ds.(rowPositioner); ok {
						rp.SetCurrentRowNo(saveRow + i)
					}
					e.renderDownThenAcross(db, rowCount-i)
					// Position DS past all rows (C# line 455).
					if rp, ok := ds.(rowPositioner); ok {
						rp.SetCurrentRowNo(saveRow + rowCount)
					}
					e.curX = saveCurX
					return
				}
				// Advance to the next column, reset Y.
				e.curY = startColumnY
				i-- // retry the same row in the new column
				continue
			}

			e.ShowFullBand(&db.BandBase)
			db.SetRowNo(db.RowNo() + 1)
			db.SetAbsRowNo(db.AbsRowNo() + 1)
		}
	} else {
		// Enough space: render all rows in proper column order (C# lines 393-453).
		maxY := e.curY
		rowNo := 0

		for i := 0; i < rowCount; i++ {
			if rp, ok := ds.(rowPositioner); ok {
				rp.SetCurrentRowNo(saveRow + i)
			}
			if fullDS, ok := ds.(data.DataSource); ok && e.report != nil {
				e.report.SetCalcContext(fullDS)
			}

			// Set column X position.
			if colIdx < len(positions) {
				e.curX = positions[colIdx] + saveCurX
			} else {
				e.curX = float32(colIdx)*db.Width() + saveCurX
			}

			e.ShowFullBand(&db.BandBase)
			e.OutlineUp()
			if e.curY > maxY {
				maxY = e.curY
			}

			db.SetRowNo(db.RowNo() + 1)
			db.SetAbsRowNo(db.AbsRowNo() + 1)
			rowNo++

			// When we've filled rowsPerColumn rows, advance to the next column.
			if rowNo >= rowsPerColumn {
				colIdx++
				rowNo = 0
				e.curY = startColumnY
			}
		}

		e.curX = saveCurX
		e.curY = maxY
	}

	// Position DS past all rows (C# line 455).
	if rp, ok := ds.(rowPositioner); ok {
		rp.SetCurrentRowNo(saveRow + rowCount)
	}
}
