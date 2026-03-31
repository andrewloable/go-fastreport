package engine

import (
	"image/color"

	"github.com/andrewloable/go-fastreport/object"
	"github.com/andrewloable/go-fastreport/preview"
	"github.com/andrewloable/go-fastreport/style"
	"github.com/andrewloable/go-fastreport/table"
)

// renderManualBuildAcrossThenDown renders a ManualBuild table using the
// AcrossThenDown layout. For each row group that fits on one page height,
// every column group gets its own page.
//
// Page order: (rg0,cg0), (rg0,cg1), ..., (rg1,cg0), (rg1,cg1), ...
// This matches C# TableResult.GeneratePagesAcrossThenDown (TableResult.cs line 411).
//
// After this function returns the engine is positioned on the last rendered page
// with curY at the bottom of the last slice. showBand will add the template band
// height as a final gap (acceptable).
func (e *ReportEngine) renderManualBuildAcrossThenDown(
	base *table.TableBase,
	originX, originY float32,
	bandName string,
) {
	cols := base.Columns()
	rows := base.Rows()
	nCols := len(cols)
	nRows := len(rows)
	fixedRows := base.FixedRows()
	fixedCols := base.FixedColumns()

	if nCols == 0 || nRows == 0 {
		return
	}

	// Pre-compute cumulative column X offsets.
	colX := make([]float32, nCols+1)
	for i, col := range cols {
		if col.Visible() {
			colX[i+1] = colX[i] + col.Width()
		} else {
			colX[i+1] = colX[i]
		}
	}

	// Pre-compute cumulative row Y offsets.
	rowY := make([]float32, nRows+1)
	for i, row := range rows {
		rowY[i+1] = rowY[i] + row.Height()
	}

	fixedColW := colX[fixedCols]
	fixedRowH := float32(0)
	if fixedRows > 0 && fixedRows <= nRows {
		fixedRowH = rowY[fixedRows]
	}

	availWidth := e.pageWidth - originX

	// Compute column groups: each group holds columns that fit horizontally.
	// Fixed columns are repeated in every group and are NOT counted in colGroups.
	type colGroup struct {
		startCol int
		endCol   int
	}
	var colGroups []colGroup
	startCol := fixedCols
	for startCol < nCols {
		endCol := startCol
		usedW := fixedColW
		for endCol < nCols {
			cw := colX[endCol+1] - colX[endCol]
			if usedW+cw > availWidth+0.1 && endCol > startCol {
				break
			}
			usedW += cw
			endCol++
		}
		if endCol == startCol {
			endCol = startCol + 1 // safety: at least 1 column per group
		}
		colGroups = append(colGroups, colGroup{startCol, endCol})
		startCol = endCol
	}
	if len(colGroups) == 0 {
		colGroups = append(colGroups, colGroup{fixedCols, nCols})
	}

	// Iterate row groups. Each row group is a set of consecutive data rows that
	// fit within one page's available height (including the fixed header rows).
	//
	// IMPORTANT: a new page must be started BEFORE computing avail so that
	// FreeSpace() reflects the fresh page height, not the tail of the previous
	// col group's page.
	firstRowGroup := true
	currentRow := fixedRows // first data row index (fixed rows are repeated in each group)
	if currentRow >= nRows {
		currentRow = 0 // no fixed rows
	}

	for currentRow < nRows {
		// For all row groups after the first, start a new page so that
		// FreeSpace() returns the full fresh-page height.
		if !firstRowGroup {
			e.startNewPageForCurrent()
		}

		avail := e.FreeSpace()
		if avail <= 0 {
			avail = e.pageHeight - e.PageFooterHeight()
		}
		if avail <= 0 {
			avail = 1
		}

		// Count how many data rows fit within avail (fixed header always included).
		rowsFit := 0
		accH := fixedRowH
		for ri := currentRow; ri < nRows; ri++ {
			rowH := rowY[ri+1] - rowY[ri]
			if rowsFit > 0 && accH+rowH > avail+0.1 {
				break
			}
			accH += rowH
			rowsFit++
		}
		if rowsFit == 0 {
			rowsFit = 1 // safety: render at least one row
		}

		endRow := currentRow + rowsFit
		if endRow > nRows {
			endRow = nRows
		}

		// Height of this row group slice (fixed header + data rows).
		sliceH := fixedRowH
		for ri := currentRow; ri < endRow; ri++ {
			sliceH += rowY[ri+1] - rowY[ri]
		}

		// Render each column group for this row group, each on its own page.
		// Column group 0 uses the already-started page; subsequent groups each
		// get their own fresh page.
		for cgIdx, cg := range colGroups {
			if cgIdx > 0 {
				e.startNewPageForCurrent()
			}

			pb := &preview.PreparedBand{
				Name:   bandName,
				Left:   0,
				Top:    e.curY,
				Height: originY + sliceH,
				Width:  e.pageWidth,
			}

			e.populateAcrossThenDownSlice(
				base, pb,
				originX, originY,
				cols, rows, colX, rowY,
				fixedRows, fixedCols,
				fixedRowH, fixedColW,
				currentRow, endRow,
				cg.startCol, cg.endCol,
			)

			_ = e.preparedPages.AddBand(pb)
			e.AdvanceY(originY + sliceH)
		}

		firstRowGroup = false
		currentRow = endRow
	}
}

// populateAcrossThenDownSlice fills a PreparedBand with the table cells for
// the specified row group [startDataRow, endDataRow) and column group
// [startCol, endCol). Fixed header rows and fixed columns are always included.
//
// Cell positions are remapped: fixed header rows start at originY, data rows
// follow immediately after the header. Column group starts after fixed columns.
func (e *ReportEngine) populateAcrossThenDownSlice(
	tbl *table.TableBase,
	pb *preview.PreparedBand,
	originX, originY float32,
	cols []*table.TableColumn,
	rows []*table.TableRow,
	colX, rowY []float32,
	fixedRows, fixedCols int,
	fixedRowH, fixedColW float32,
	startDataRow, endDataRow int,
	startCol, endCol int,
) {
	nCols := len(cols)
	nRows := len(rows)

	// Total width and height of this slice.
	cgW := colX[endCol] - colX[startCol]
	totalW := fixedColW + cgW
	sliceH := fixedRowH
	for ri := startDataRow; ri < endDataRow; ri++ {
		sliceH += rowY[ri+1] - rowY[ri]
	}

	// Section background (page-width background, covers the table area).
	sectionBg := preview.PreparedObject{
		Kind:             preview.ObjectTypePicture,
		Left:             0,
		Top:              originY,
		Width:            pb.Width,
		Height:           sliceH,
		BlobIdx:          -1,
		IgnoreForRowSnap: true,
	}
	pb.Objects = append(pb.Objects, sectionBg)

	// Table outer border container.
	sectionBorder := preview.PreparedObject{
		Kind:             preview.ObjectTypeText,
		Left:             originX,
		Top:              originY,
		Width:            totalW,
		Height:           sliceH,
		BlobIdx:          -1,
		Font:             style.DefaultFont(),
		WordWrap:         true,
		Clip:             true,
		Border:           tbl.Border(),
		IgnoreForRowSnap: true,
	}
	pb.Objects = append(pb.Objects, sectionBorder)

	// rowTopInSlice returns the Top coordinate of row ri within the PreparedBand.
	rowTopInSlice := func(ri int) float32 {
		if ri < fixedRows {
			// Fixed header row: positioned relative to originY.
			return originY + rowY[ri]
		}
		// Data row: remapped to follow directly after the fixed header.
		return originY + fixedRowH + (rowY[ri] - rowY[startDataRow])
	}

	// Build the list of rows to render: fixed header rows + data rows.
	type rowRange struct {
		start, end int
	}
	renderRanges := []rowRange{}
	if fixedRows > 0 {
		renderRanges = append(renderRanges, rowRange{0, fixedRows})
	}
	renderRanges = append(renderRanges, rowRange{startDataRow, endDataRow})

	// Column ranges to render: fixed columns (repeated) + data column group.
	type colRenderGroup struct {
		startC, endC int
		xShift       float32 // shift to apply so this group starts after fixed cols
	}
	var renderCols []colRenderGroup
	if fixedCols > 0 {
		renderCols = append(renderCols, colRenderGroup{0, fixedCols, 0})
	}
	// Data column group: shift left by colX[startCol] then offset by fixedColW.
	renderCols = append(renderCols, colRenderGroup{
		startC: startCol,
		endC:   endCol,
		xShift: fixedColW - colX[startCol],
	})

	for _, rr := range renderRanges {
		for ri := rr.start; ri < rr.end && ri < nRows; ri++ {
			row := rows[ri]
			for _, cg := range renderCols {
				for ci := cg.startC; ci < cg.endC && ci < nCols; ci++ {
					if !cols[ci].Visible() {
						continue
					}
					if tbl.IsInsideSpan(ci, ri) {
						continue
					}
					if ci >= len(row.Cells()) {
						continue
					}
					cell := row.Cells()[ci]
					if cell == nil {
						continue
					}

					colSpan := cell.ColSpan()
					if colSpan < 1 {
						colSpan = 1
					}
					cellEndCol := ci + colSpan
					if cellEndCol > cg.endC {
						cellEndCol = cg.endC
					}
					if cellEndCol > nCols {
						cellEndCol = nCols
					}
					cellW := colX[cellEndCol] - colX[ci]

					rowSpan := cell.RowSpan()
					if rowSpan < 1 {
						rowSpan = 1
					}
					cellEndRow := ri + rowSpan
					if cellEndRow > nRows {
						cellEndRow = nRows
					}
					// Cell height based on raw row span.
					cellH := rowY[cellEndRow] - rowY[ri]

					absLeft := originX + cg.xShift + colX[ci]
					absTop := rowTopInSlice(ri)

					cellText := e.evalTextWithFormat(cell.Text(), cell.Format())
					cellFill := colorFromFill(cell.Fill())
					textColor := cell.TextColor()
					if textColor.A == 0 {
						textColor = color.RGBA{A: 255}
					}
					pad := cell.Padding()
					po := preview.PreparedObject{
						Name:          cell.Name(),
						Kind:          preview.ObjectTypeText,
						Left:          absLeft,
						Top:           absTop,
						Width:         cellW,
						Height:        cellH,
						Text:          cellText,
						BlobIdx:       -1,
						Font:          cell.Font(),
						TextColor:     textColor,
						FillColor:     cellFill,
						HorzAlign:     int(cell.HorzAlign()),
						VertAlign:     int(cell.VertAlign()),
						WordWrap:      cell.WordWrap(),
						Clip:          true,
						Border:        cell.Border(),
						PaddingLeft:   pad.Left,
						PaddingTop:    pad.Top,
						PaddingRight:  pad.Right,
						PaddingBottom: pad.Bottom,
					}
					pb.Objects = append(pb.Objects, po)

					// Render embedded PictureObjects inside the cell.
					// Mirrors populateTableObjects: cells can contain child PictureObjects
					// (e.g. Photos with BindableControl="Picture") that are rendered as
					// separate LayerBack+LayerPicture div pairs.
					for _, childObj := range cell.Objects() {
						if pic, ok := childObj.(*object.PictureObject); ok {
							picPO := preview.PreparedObject{
								Name:    pic.Name(),
								Kind:    preview.ObjectTypePicture,
								Left:    absLeft + pic.Left(),
								Top:     absTop + pic.Top(),
								Width:   pic.Width(),
								Height:  pic.Height(),
								BlobIdx: -1,
							}
							if imgData := pic.ImageData(); len(imgData) > 0 {
								if e.preparedPages != nil {
									picPO.BlobIdx = e.preparedPages.BlobStore.Add("", imgData)
								}
							}
							pb.Objects = append(pb.Objects, picPO)
						}
					}
				}
			}
		}
	}
}
