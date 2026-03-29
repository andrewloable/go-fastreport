package engine

import "github.com/andrewloable/go-fastreport/preview"

// matrixHSplitInfo holds the metadata needed to split a wide matrix table
// horizontally across pages. Set during matrix processing in objects.go and
// consumed by splitBandHorizontallyForMatrix in bands.go.
type matrixHSplitInfo struct {
	colBounds  []float32 // cumulative column X offsets (relative to table left = 0)
	originX    float32   // absolute X of the matrix table (renderLeft)
	fixedCols  int       // number of fixed (row-header) columns
	fixedWidth float32   // total width of fixed columns
}

// splitBandHorizontallyForMatrix splits a wide matrix band that was already
// added to the current page. It trims the current band to only show the first
// column group, then creates new pages for remaining column groups.
//
// C# ref: TableResult.GeneratePagesDownThenAcross / GetColumnsFit.
func (e *ReportEngine) splitBandHorizontallyForMatrix(pb *preview.PreparedBand) {
	info := e.pendingHSplit
	if info == nil {
		return
	}

	availWidth := e.pageWidth - info.originX
	fixedRight := info.fixedWidth + info.originX
	nCols := len(info.colBounds) - 1

	// Determine which data columns fit on the first page.
	firstEnd := info.fixedCols
	for firstEnd < nCols {
		colRight := info.colBounds[firstEnd+1]
		if info.fixedWidth+colRight-info.colBounds[info.fixedCols] > availWidth+0.1 && firstEnd > info.fixedCols {
			break
		}
		firstEnd++
	}
	firstPageRight := info.colBounds[firstEnd] + info.originX

	// Partition objects: first page vs overflow.
	// Use a negative tolerance so objects at the exact boundary (the start of
	// the first excluded column) go to overflow, not to the first page.
	// Column widths are > 50px, so -1.0 safely excludes boundary objects
	// while keeping all objects in prior columns.
	splitX := firstPageRight - 1.0

	// Use midpoint between fixed columns and first data column to avoid
	// including data column objects that start at the exact same x as fixedRight.
	fixedMid := fixedRight
	if info.fixedCols < nCols {
		firstDataX := info.colBounds[info.fixedCols] + info.originX
		fixedMid = (fixedRight + firstDataX) / 2
	}

	var firstObjs, overflowObjs, fixedObjs []preview.PreparedObject
	// straddleObjs: objects on page 1 that span into the overflow area
	// (e.g. ColSpan=2 "Total" header). Detected from ORIGINAL widths
	// before trimming.
	var straddleObjs []preview.PreparedObject

	for _, po := range pb.Objects {
		if po.Left < splitX {
			// Collect fixed-column objects.
			if po.Left < fixedMid {
				fixedObjs = append(fixedObjs, po)
			}

			// Detect straddling objects BEFORE trimming. Only include objects
			// whose right edge extends well past firstPageRight (at least 1px
			// into the overflow area). Single-column cells at the boundary
			// end exactly at firstPageRight and must NOT be duplicated.
			origRight := po.Left + po.Width
			if po.Left >= fixedMid && origRight > firstPageRight+1.0 {
				straddleObjs = append(straddleObjs, po) // save with original width
			}

			// Trim objects that extend past the split boundary (e.g. ColSpan
			// headers or table-wide border containers).
			if po.Left+po.Width > firstPageRight {
				po.Width = firstPageRight - po.Left
			}
			firstObjs = append(firstObjs, po)
		} else {
			overflowObjs = append(overflowObjs, po)
		}
	}

	// Replace the already-added band's objects with trimmed first-page objects.
	pb.Objects = firstObjs

	// Build continuation pages for remaining column groups.
	startCol := firstEnd
	for startCol < nCols {
		endCol := startCol
		usedW := info.fixedWidth
		for endCol < nCols {
			cw := info.colBounds[endCol+1] - info.colBounds[endCol]
			if usedW+cw > availWidth+0.1 && endCol > startCol {
				break
			}
			usedW += cw
			endCol++
		}
		if endCol == startCol {
			endCol = startCol + 1
		}

		groupLeft := info.colBounds[startCol] + info.originX
		groupRight := info.colBounds[endCol] + info.originX
		xShift := fixedRight - groupLeft // shift group objects to start after fixed cols

		// Start a new page (shows page footer on current, page header on new).
		e.startNewPageForCurrent()

		contPb := &preview.PreparedBand{
			Name:   pb.Name,
			Left:   pb.Left,
			Top:    e.curY,
			Width:  pb.Width,
			Height: pb.Height,
		}

		// Duplicate fixed column objects on the continuation page.
		contPb.Objects = append(contPb.Objects, fixedObjs...)

		// Include straddling objects (e.g. ColSpan "Total" header) trimmed
		// to show only the portion within this continuation group.
		for _, po := range straddleObjs {
			origRight := po.Left + po.Width
			trimmed := po
			trimmed.Left = fixedRight
			trimmed.Width = origRight + xShift - fixedRight
			if trimmed.Width > 0 {
				contPb.Objects = append(contPb.Objects, trimmed)
			}
		}

		// Add this group's objects, shifted left.
		for _, po := range overflowObjs {
			if po.Left >= groupLeft-0.1 && po.Left < groupRight+0.1 {
				shifted := po
				shifted.Left += xShift
				contPb.Objects = append(contPb.Objects, shifted)
			}
		}

		_ = e.preparedPages.AddBand(contPb)
		e.AdvanceY(contPb.Height)

		startCol = endCol
	}
}
