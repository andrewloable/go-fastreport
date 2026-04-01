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

			// Trim objects that extend past the split boundary.
			// Exception: full-page section background objects (ObjectTypePicture,
			// BlobIdx<0, IgnoreForRowSnap) must keep their original width (= page width)
			// so the background fills the whole page, matching C# behavior.
			if po.Left+po.Width > firstPageRight {
				if !(po.IgnoreForRowSnap && po.Kind == preview.ObjectTypePicture && po.BlobIdx < 0) {
					po.Width = firstPageRight - po.Left
				}
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
			Name:         pb.Name,
			Left:         pb.Left,
			Top:          pb.Top,
			Width:        pb.Width,
			Height:       pb.Height,
			NoBackground: true,
		}

		// Duplicate fixed column objects on the continuation page.
		// IgnoreForRowSnap border containers (sectionBorder) that span beyond the
		// fixed columns must be trimmed to the visible width of this page group.
		// Full-page background objects (ObjectTypePicture, BlobIdx<0) keep their width.
		// C# ref: each continuation section shows only the columns visible on that page.
		for _, fo := range fixedObjs {
			adj := fo
			if fo.IgnoreForRowSnap && fo.Kind != preview.ObjectTypePicture && fo.Left+fo.Width > fixedRight {
				adj.Width = usedW
			}
			contPb.Objects = append(contPb.Objects, adj)
		}

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

// splitMatrixAcrossThenDown handles a wide matrix PreparedBand that needs BOTH
// horizontal (column groups) and vertical (row groups) page splitting.
//
// When a matrix needs only horizontal splitting, splitBandHorizontallyForMatrix
// modifies the already-added pb.Objects and creates continuation pages. But when
// it also needs vertical splitting, splitPreparedBandAcrossPages creates new slice
// PreparedBands copied from pb.Objects — those slices include ALL columns (the
// wide overflow objects). splitBandHorizontallyForMatrix then operates on the
// original pb which is no longer in any page, so the fix never takes effect.
//
// This function computes both row groups (vertical slices) and column groups
// (horizontal slices) upfront, then renders one page per (row_group, col_group)
// pair in AcrossThenDown order: RG0_CG0, RG0_CG1, ..., RG1_CG0, RG1_CG1, ...
//
// C# ref: TableResult.GeneratePagesAcrossThenDown (TableResult.cs).
func (e *ReportEngine) splitMatrixAcrossThenDown(pb *preview.PreparedBand) {
	info := e.pendingHSplit
	if info == nil {
		return
	}

	availWidth := e.pageWidth - info.originX
	fixedRight := info.fixedWidth + info.originX
	nCols := len(info.colBounds) - 1
	fixedH := pb.FixedHeaderHeight

	// Midpoint used to separate fixed-column objects from data-column objects.
	fixedMid := fixedRight
	if info.fixedCols < nCols {
		firstDataX := info.colBounds[info.fixedCols] + info.originX
		fixedMid = (fixedRight + firstDataX) / 2
	}

	// Compute column groups (same logic as splitBandHorizontallyForMatrix).
	type colGroup struct {
		startCol int
		endCol   int
		usedW    float32 // fixedWidth + data-cols width for this group
	}
	var colGroups []colGroup
	{
		sc := info.fixedCols
		for sc < nCols {
			ec := sc
			usedW := info.fixedWidth
			for ec < nCols {
				cw := info.colBounds[ec+1] - info.colBounds[ec]
				if usedW+cw > availWidth+0.1 && ec > sc {
					break
				}
				usedW += cw
				ec++
			}
			if ec == sc {
				ec = sc + 1
			}
			colGroups = append(colGroups, colGroup{sc, ec, usedW})
			sc = ec
		}
	}
	if len(colGroups) == 0 {
		return
	}

	// Compute row group boundaries (same snapping logic as splitPreparedBandAcrossPages).
	// The first row group uses the current FreeSpace(); subsequent groups use the full page.
	type rowGroup struct {
		offset    float32
		breakLine float32
		isFirst   bool
	}
	var rowGroups []rowGroup
	{
		remaining := pb.Height
		offset := float32(0)
		isFirst := true
		for remaining > 0 {
			avail := float32(0)
			if isFirst {
				avail = e.FreeSpace()
			} else {
				avail = e.pageHeight - e.PageFooterHeight()
			}
			if !isFirst && fixedH > 0 {
				avail -= fixedH
			}
			if avail <= 0 {
				avail = 1
			}

			breakLine := offset + avail
			if breakLine > pb.Height {
				breakLine = pb.Height
			}

			// Snap to row boundaries (skip IgnoreForRowSnap background/border containers).
			for _, po := range pb.Objects {
				if po.IgnoreForRowSnap {
					continue
				}
				if po.Top > offset && po.Top < breakLine && po.Top+po.Height > breakLine {
					breakLine = po.Top
				}
			}
			if breakLine <= offset {
				breakLine = offset + avail
				if breakLine > pb.Height {
					breakLine = pb.Height
				}
			}

			rowGroups = append(rowGroups, rowGroup{offset, breakLine, isFirst})
			offset = breakLine
			remaining = pb.Height - offset
			isFirst = false
		}
	}

	// Render each (row group, col group) pair: AcrossThenDown order.
	// For each row group → for each col group → one page.
	firstPage := true
	for _, rg := range rowGroups {
		headerOffset := float32(0)
		if !rg.isFirst && fixedH > 0 {
			headerOffset = fixedH
		}
		sliceH := rg.breakLine - rg.offset

		for cgIdx, cg := range colGroups {
			if !firstPage {
				e.startNewPageForCurrent()
			}
			firstPage = false

			groupLeft := info.colBounds[cg.startCol] + info.originX
			groupRight := info.colBounds[cg.endCol] + info.originX
			// xShift moves data-column objects so they start right after the fixed columns.
			// For continuation col groups (cgIdx>0) the table is anchored at Left=0 (not
			// originX) to match C# which positions continuation pages flush to the page
			// margin. Subtract originX from xShift to compensate.
			xShift := fixedRight - groupLeft
			if cgIdx > 0 {
				xShift -= info.originX
			}

			sliceBand := &preview.PreparedBand{
				Name:             pb.Name,
				Left:             pb.Left,
				Top:              e.curY,
				Width:            pb.Width,
				Height:           sliceH + headerOffset,
				// BackgroundHeight only applies to the first page (gap above the matrix).
				BackgroundHeight: 0,
				// Continuation col groups suppress the band background div; sectionBg
				// objects (in fixedObjs via Left=0) provide the visual page background.
				NoBackground: cgIdx > 0,
			}
			if rg.isFirst {
				sliceBand.BackgroundHeight = pb.BackgroundHeight
			}

			// isSectionBg returns true for the full-page background object (ObjectTypePicture,
			// BlobIdx<0, IgnoreForRowSnap). This object always stays at Left=0 regardless of
			// which col group we're rendering; only non-bg fixed objects get the leftShift.
			isSectionBg := func(po preview.PreparedObject) bool {
				return po.IgnoreForRowSnap && po.Kind == preview.ObjectTypePicture && po.BlobIdx < 0
			}

			// leftShift for continuation col groups moves fixed-column objects to start at 0
			// instead of originX, matching C# which re-anchors the table at the page left edge.
			leftShift := float32(0)
			if cgIdx > 0 {
				leftShift = -info.originX
			}

			// On continuation row groups, prepend repeated fixed-header objects
			// (Top < fixedH). These are at their original positions within [0, fixedH).
			if !rg.isFirst && fixedH > 0 {
				for _, po := range pb.Objects {
					if po.Top+po.Height > fixedH {
						continue // not a header object
					}
					isFixed := po.Left < fixedMid
					// Right boundary is EXCLUSIVE (no +0.1) to prevent boundary-column
					// objects from appearing in both the current and the next col group.
					inColGroup := po.Left >= groupLeft-0.1 && po.Left < groupRight
					if !isFixed && !inColGroup {
						continue
					}
					hpo := po
					if isFixed {
						if po.IgnoreForRowSnap && po.Kind != preview.ObjectTypePicture &&
							po.Left+po.Width > fixedRight {
							hpo.Width = cg.usedW
						}
						if !isSectionBg(po) {
							hpo.Left += leftShift
						}
					} else {
						hpo.Left += xShift
					}
					sliceBand.Objects = append(sliceBand.Objects, hpo)
				}
			}

			// Add objects visible in this row group and column group.
			for _, po := range pb.Objects {
				objTop := po.Top
				objBot := po.Top + po.Height

				// Row group visibility: skip objects entirely outside this slice.
				if objBot <= rg.offset || objTop >= rg.breakLine {
					continue
				}

				// Column group visibility. Right boundary is EXCLUSIVE (no +0.1 tolerance)
				// to prevent boundary objects from appearing in two adjacent col groups.
				isFixed := po.Left < fixedMid
				inColGroup := po.Left >= groupLeft-0.1 && po.Left < groupRight
				if !isFixed && !inColGroup {
					continue
				}

				clone := po

				// Adjust Top relative to this row group's start.
				if objTop < rg.offset {
					clone.Height -= rg.offset - objTop
					clone.Top = headerOffset
				} else {
					clone.Top = objTop - rg.offset + headerOffset
				}
				// Clip height to row group boundary.
				if objBot > rg.breakLine {
					if objTop < rg.offset {
						clone.Height = sliceH
					} else {
						clone.Height = rg.breakLine - objTop
					}
				}

				// Column-group-specific adjustments.
				if isFixed {
					// Trim sectionBorder (IgnoreForRowSnap text container) to this group's width.
					// sectionBg (ObjectTypePicture, BlobIdx<0) keeps its full page width.
					if po.IgnoreForRowSnap && po.Kind != preview.ObjectTypePicture &&
						po.Left+po.Width > fixedRight {
						clone.Width = cg.usedW
					}
					if !isSectionBg(po) {
						clone.Left += leftShift
					}
				} else {
					clone.Left += xShift
				}

				sliceBand.Objects = append(sliceBand.Objects, clone)
			}

			_ = e.preparedPages.AddBand(sliceBand)
			e.AdvanceY(sliceBand.Height)
		}
	}
}
