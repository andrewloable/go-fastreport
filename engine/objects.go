package engine

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"math"
	"regexp"
	"strings"

	"github.com/andrewloable/go-fastreport/band"
	"github.com/andrewloable/go-fastreport/barcode"
	"github.com/andrewloable/go-fastreport/crossview"
	"github.com/andrewloable/go-fastreport/expr"
	"github.com/andrewloable/go-fastreport/format"
	"github.com/andrewloable/go-fastreport/gauge"
	"github.com/andrewloable/go-fastreport/maprender"
	"github.com/andrewloable/go-fastreport/matrix"
	"github.com/andrewloable/go-fastreport/object"
	"github.com/andrewloable/go-fastreport/preview"
	"github.com/andrewloable/go-fastreport/report"
	"github.com/andrewloable/go-fastreport/sparkline"
	"github.com/andrewloable/go-fastreport/style"
	"github.com/andrewloable/go-fastreport/table"
	"github.com/andrewloable/go-fastreport/utils"
)


// countRE matches table aggregate expressions like [Count(Cell2)], [Sum(Cell5)].
// Used in autoManualBuild to substitute data row counts in trailer columns.
var countRE = regexp.MustCompile(`\[Count\([^)]+\)\]`)

// populateBandObjects converts the child report objects of a BandBase into
// preview.PreparedObject snapshots and appends them to pb.
// It evaluates [bracket] expressions in TextObject text via Report.Calc().
func (e *ReportEngine) populateBandObjects(bb *band.BandBase, pb *preview.PreparedBand) {
	if bb == nil {
		return
	}
	// Apply dock layout using the band's own width/height as the container.
	applyDockLayout(bb.Objects(), bb.Width(), bb.Height())
	e.populateBandObjects2(bb, bb.Objects(), pb)
}

// populateBandObjects2 converts objects from any ObjectCollection into PreparedObjects.
// parentBand is the BandBase that owns objs; it is used to build sender predicates
// for ProcessAtDataFinished and ProcessAtGroupFinished deferred handlers.
func (e *ReportEngine) populateBandObjects2(parentBand *band.BandBase, objs *report.ObjectCollection, pb *preview.PreparedBand) {
	if objs == nil {
		return
	}
	ss := e.report.Styles()
	for i := 0; i < objs.Len(); i++ {
		obj := objs.Get(i)
		// Apply named style overrides before snapshotting object properties.
		if ss != nil {
			if styleable, ok := obj.(style.Styleable); ok {
				ss.ApplyToObject(styleable)
			}
		}
		if po := e.buildPreparedObject(obj); po != nil {
			idx := len(pb.Objects)
			pb.Objects = append(pb.Objects, *po)
			// Recursively add ContainerObject children with coordinate offsets.
			if cont, ok := obj.(*object.ContainerObject); ok {
				e.populateContainerChildren(cont, cont.Left(), cont.Top(), pb)
			}
			// Render TableObject cells as individual PreparedObjects.
			if tbl, ok := obj.(*table.TableObject); ok {
				base := &tbl.TableBase
				if tbl.IsManualBuild() {
					if built := tbl.InvokeManualBuild(); built != nil {
						base = built
					} else if built := e.autoManualBuild(tbl); built != nil {
						base = built
					}
				}
				// Apply AutoSize, auto-spans, and CellDuplicates before rendering.
				// Mirrors C# TableResult.GeneratePages() calling CalcWidth/CalcHeight,
				// CalcSpans, and ProcessDuplicates (TableBase.cs, TableResult.cs).
				base.CalcWidth()
				base.CalcHeight()
				base.ProcessDuplicates()
				e.populateTableObjects(base, tbl.Left(), tbl.Top(), pb)
				// For ManualBuild/wrapped tables, shrink the band height to the
				// table's top offset. The table cells extend beyond the band.
				// Mirrors C# where GeneratePages adjusts the band height.
				if tbl.IsManualBuild() && tbl.Top() > 0 {
					pb.Height = tbl.Top()
				}
			}
			// Render AdvMatrixObject physical cells as individual PreparedObjects.
			if adv, ok := obj.(*object.AdvMatrixObject); ok {
				e.populateAdvMatrixCells(adv, pb)
			}
			// Render MatrixObject (classic cross-tab matrix).
			// C# source: MatrixObject.GetData → MatrixHelper.StartPrint/AddDataRow/FinishPrint.
			if mx, ok := obj.(*matrix.MatrixObject); ok {
				// Resolve data source if not already set.
				if mx.DataSource == nil && mx.DataSourceName != "" && e.report != nil {
					dict := e.report.Dictionary()
					if dict != nil {
						ds := dict.FindDataSourceByAlias(mx.DataSourceName)
						if ds == nil {
							ds = dict.FindDataSourceByName(mx.DataSourceName)
						}
						if ds == nil {
							// Try case-insensitive search through all data sources.
							for _, existing := range dict.DataSources() {
								if strings.EqualFold(existing.Name(), mx.DataSourceName) ||
									strings.EqualFold(existing.Alias(), mx.DataSourceName) {
									ds = existing
									break
								}
							}
						}
						mx.DataSource = ds
					}
				}
				// Extract cell format from template cells (must be after FRX load).
				mx.ExtractCellFormat()
				// Collect data from the bound data source.
				calc := func(expr string) any {
					val, err := e.report.Calc(expr)
					if err != nil {
						return nil
					}
					return val
				}
				e.report.SetCalcContext(mx.DataSource)
				mx.GetDataWithCalc(calc)
				// Bridge runtime store to multi-level trees for BuildTemplateMultiLevel.
				mx.SyncRuntimeToMultiLevel()
				// Build the result table from collected data.
				mx.BuildTemplateMultiLevel()
				// Apply AutoSize and render cells.
				mx.TableBase.CalcWidth()
				mx.TableBase.CalcHeight()
				e.populateTableObjects(&mx.TableBase, mx.Left(), mx.Top(), pb)
				// Shrink band height to the matrix's Top offset (space above matrix).
				// The matrix cells extend beyond the band. Mirrors C# GeneratePages.
				if mx.Top() > 0 {
					pb.Height = mx.Top()
				}
			}
			// Render CrossViewReportObject cells as a grid of PreparedObjects.
			// Mirrors C# CrossViewObject.GetData + engine rendering in CrossViewObject.cs.
			if cv, ok := obj.(*crossview.CrossViewReportObject); ok {
				cv.GetData()
				e.populateCrossViewGrid(cv, pb)
			}
			// Render CellularTextObject as a character-grid of PreparedObjects.
			if cellular, ok := obj.(*object.CellularTextObject); ok {
				text := e.evalText(cellular.Text())
				e.populateCellularTextCells(cellular, text, pb)
			}
			// Register deferred text evaluation for non-default ProcessAt values.
			// C# ProcessAt.cs AddObjectToProcess (line 200-205): Custom objects are
			// registered separately for manual processing via Engine.ProcessObject(obj).
			// All other non-default values are queued as one-shot deferred handlers.
			if txt, ok := obj.(*object.TextObject); ok && txt.ProcessAt() != object.ProcessAtDefault {
				capturedPb := pb
				capturedIdx := idx
				capturedTxt := txt
				capturedFmt := txt.Format()
				pb.Objects[idx].Text = "" // placeholder: blank until deferred state fires
				evalFn := func() {
					// Mirrors C# ProcessInfo.Process() (ReportEngine.ProcessAt.cs lines 88-108):
					// re-evaluate text expression and re-apply highlight conditions with the
					// current data context, then update text, FillColor, TextColor, Font on
					// the already-placed PreparedObject.
					po := &capturedPb.Objects[capturedIdx]
					po.Text = e.evalTextWithFormat(capturedTxt.Text(), capturedFmt)
					// Restore design-time base fill/text-color/font.
					if f, ok2 := capturedTxt.Fill().(*style.SolidFill); ok2 && f.Color.A > 0 {
						po.FillColor = f.Color
					} else {
						po.FillColor = color.RGBA{}
					}
					po.TextColor = capturedTxt.TextColor()
					po.Font = capturedTxt.Font()
					// Suppress invisible text (alpha=0 text color).
					if po.TextColor.A == 0 && po.Text != "" {
						po.Text = ""
					}
					// Re-apply highlight conditions with the new data context.
					if e.report != nil {
						for _, cond := range capturedTxt.Highlights() {
							result, err := e.report.Calc(cond.Expression)
							if err != nil {
								continue
							}
							matched, _ := result.(bool)
							if !matched {
								continue
							}
							if !cond.Visible {
								po.Text = ""
								break
							}
							if cond.ApplyFill {
								po.FillColor = cond.FillColor
							}
							if cond.ApplyTextFill {
								po.TextColor = cond.TextFillColor
							}
							if cond.ApplyFont {
								po.Font = cond.Font
							}
							break
						}
					}
				}
				if txt.ProcessAt() == object.ProcessAtCustom {
					// Custom: register for manual ProcessObject(obj) call.
					// Mirrors C# AddObjectToProcess (ProcessAt.cs lines 200-205).
					e.RegisterCustomObject(txt, evalFn)
				} else {
					// All other values use one-shot handlers. C# removes ProcessInfo
					// from objectsToProcess after it fires (ProcessAt.cs lines 180-184).
					// Each band render registers a new one-shot handler, so page/column
					// footers naturally re-register on each page/column they are rendered on.
					state := processAtToEngineState(txt.ProcessAt())
					switch txt.ProcessAt() {
					case object.ProcessAtDataFinished:
						// Only fire when the relevant DataBand finishes.
						// Mirrors C# ProcessInfo.Process lines 123-131.
						e.AddSenderDeferredHandler(state, makeDataFinishedPred(parentBand), evalFn)
					case object.ProcessAtGroupFinished:
						// Only fire when the owning GroupHeaderBand finishes.
						// Mirrors C# ProcessInfo.Process lines 133-137.
						e.AddSenderDeferredHandler(state, makeGroupFinishedPred(parentBand), evalFn)
					default:
						e.AddDeferredHandler(state, evalFn)
					}
				}
			}
		}
	}
}

// populateContainerChildren renders the children of a ContainerObject into pb,
// offsetting each child's Left/Top by the accumulated container origin (offsetX, offsetY).
func (e *ReportEngine) populateContainerChildren(c *object.ContainerObject, offsetX, offsetY float32, pb *preview.PreparedBand) {
	objs := c.Objects()
	if objs == nil {
		return
	}
	// Apply dock layout within the container's own bounds.
	applyDockLayout(objs, c.Width(), c.Height())
	for i := 0; i < objs.Len(); i++ {
		child := objs.Get(i)
		if po := e.buildPreparedObject(child); po != nil {
			po.Left += offsetX
			po.Top += offsetY
			pb.Objects = append(pb.Objects, *po)
			// Recurse into nested containers.
			if nested, ok := child.(*object.ContainerObject); ok {
				e.populateContainerChildren(nested, offsetX+nested.Left(), offsetY+nested.Top(), pb)
			}
		}
	}
}

// populateTableObjects flattens a TableBase's cells into PreparedObjects.
// Each cell is rendered at its computed absolute position (originX + colOffset,
// originY + rowOffset). ColSpan and RowSpan determine cell size.
// When Layout is Wrapped, columns are split into chunks that fit the page width,
// with FixedColumns repeated for each chunk (mirroring C# TableBase.GeneratePages).
func (e *ReportEngine) populateTableObjects(tbl *table.TableBase, originX, originY float32, pb *preview.PreparedBand) {
	cols := tbl.Columns()
	nCols := len(cols)
	if nCols == 0 {
		return
	}

	// Pre-compute cumulative column X offsets.
	colX := make([]float32, nCols+1)
	for i, col := range cols {
		colX[i+1] = colX[i] + col.Width()
	}

	// Compute total table height (sum of all row heights).
	tableH := float32(0)
	for _, row := range tbl.Rows() {
		tableH += row.Height()
	}

	// Determine column ranges to render. For Wrapped layout, split into sections.
	type colRange struct {
		startCol int // first column index (inclusive)
		endCol   int // last column index (exclusive)
		xOffset  float32
		yOffset  float32
	}
	var sections []colRange

	if tbl.Layout() == table.TableLayoutWrapped && tbl.FixedColumns() > 0 {
		fixedCols := tbl.FixedColumns()
		fixedWidth := colX[fixedCols]
		// Available width for data columns per section.
		pageWidth := pb.Width
		if pageWidth <= 0 {
			pageWidth = 718.2 // A4 default
		}
		availWidth := pageWidth - originX - fixedWidth

		// First pass: compute section boundaries (data column chunks).
		type sectionInfo struct {
			dataStart, dataEnd int
			chunkW             float32
			wrapY              float32
		}
		var sectionInfos []sectionInfo
		wrapY := float32(0)
		dataCol := fixedCols
		for dataCol < nCols {
			chunkEnd := dataCol
			chunkW := float32(0)
			for chunkEnd < nCols {
				cw := cols[chunkEnd].Width()
				if chunkW+cw > availWidth && chunkEnd > dataCol {
					break
				}
				chunkW += cw
				chunkEnd++
			}
			if chunkEnd == dataCol {
				chunkEnd = dataCol + 1
			}
			sectionInfos = append(sectionInfos, sectionInfo{dataCol, chunkEnd, chunkW, wrapY})
			dataCol = chunkEnd
			wrapY += tableH + tbl.WrappedGap()
		}

		// Total height of all wrapped sections.
		nSec := len(sectionInfos)
		totalWrappedH := float32(nSec)*tableH + float32(nSec-1)*tbl.WrappedGap()

		// Second pass: emit section backgrounds, border containers, and column ranges.
		for _, sec := range sectionInfos {
			// Section background: cascading height from section top to the page footer.
			// Mirrors C# where each section PreparedBand spans to the page footer.
			// bandContentArea = pageHeight - bandAbsoluteTop - footerHeight.
			bandContentArea := totalWrappedH + originY
			if e.pageHeight > 0 {
				footerH := e.PageFooterHeight()
				bandContentArea = e.pageHeight - pb.Top - footerH
			}
			remainingH := bandContentArea - (originY + sec.wrapY)
			// Use ObjectTypePicture for section bg — its LayerBack CSS matches
			// C#'s band background: text-align:center, color:white, border:none.
			sectionBg := preview.PreparedObject{
				Kind:    preview.ObjectTypePicture,
				Left:    0,
				Top:     originY + sec.wrapY,
				Width:   pageWidth,
				Height:  remainingH,
				BlobIdx: -1,
			}
			pb.Objects = append(pb.Objects, sectionBg)

			// Table border container per section.
			sectionW := fixedWidth + sec.chunkW
			sectionBorder := preview.PreparedObject{
				Kind:     preview.ObjectTypeText,
				Left:     originX,
				Top:      originY + sec.wrapY,
				Width:    sectionW,
				Height:   tableH,
				BlobIdx:  -1,
				Font:     style.DefaultFont(), // C#: TextObjectBase default = Arial 10pt
				WordWrap: true,
				Clip:     true,
				Border:   tbl.Border(),
			}
			pb.Objects = append(pb.Objects, sectionBorder)

			// Emit fixed columns + data columns for this section.
			sections = append(sections, colRange{0, fixedCols, 0, sec.wrapY})
			sections = append(sections, colRange{sec.dataStart, sec.dataEnd, fixedWidth - colX[sec.dataStart], sec.wrapY})
		}
	} else {
		// Non-wrapped: single section with all columns.
		sections = append(sections, colRange{0, nCols, 0, 0})
	}

	// Render cells for each section.
	for _, sec := range sections {
		rowY := float32(0)
		for ri, row := range tbl.Rows() {
			rowH := row.Height()
			for ci := sec.startCol; ci < sec.endCol; ci++ {
				if ci >= len(row.Cells()) {
					continue
				}
				cell := row.Cells()[ci]
				if cell == nil {
					continue
				}
				if tbl.IsInsideSpan(ci, ri) {
					continue
				}
				colSpan := cell.ColSpan()
				if colSpan < 1 {
					colSpan = 1
				}
				endCol := ci + colSpan
				if endCol > sec.endCol {
					endCol = sec.endCol
				}
				if endCol > nCols {
					endCol = nCols
				}
				cellW := colX[endCol] - colX[ci]

				rowSpan := cell.RowSpan()
				if rowSpan < 1 {
					rowSpan = 1
				}
				cellH := float32(0)
				for si := ri; si < ri+rowSpan && si < len(tbl.Rows()); si++ {
					cellH += tbl.Rows()[si].Height()
				}

				absLeft := originX + sec.xOffset + colX[ci]
				absTop := originY + sec.yOffset + rowY

				cellText := e.evalTextWithFormat(cell.Text(), cell.Format())
				cellFill := color.RGBA{R: 255, G: 255, B: 255, A: 0}
				if f, ok := cell.Fill().(*style.SolidFill); ok {
					cellFill = f.Color
				}
				po := preview.PreparedObject{
					Name:      cell.Name(),
					Kind:      preview.ObjectTypeText,
					Left:      absLeft,
					Top:       absTop,
					Width:     cellW,
					Height:    cellH,
					Text:      cellText,
					BlobIdx:   -1,
					Font:      cell.Font(),
					TextColor: color.RGBA{A: 255},
					FillColor: cellFill,
					HorzAlign: int(cell.HorzAlign()),
					VertAlign: int(cell.VertAlign()),
					WordWrap:  cell.WordWrap(),
					Clip:      true, // C# TableCell defaults to Clip=true
					Border:    cell.Border(),
				}
				pb.Objects = append(pb.Objects, po)

				// Render embedded PictureObjects inside the cell.
				// C# table cells can contain child PictureObjects (e.g. Photo column
				// with BindableControl="Picture"). These are rendered as separate
				// LayerBack+LayerPicture div pairs.
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
						if data := pic.ImageData(); len(data) > 0 {
							if e.preparedPages != nil {
								picPO.BlobIdx = e.preparedPages.BlobStore.Add("", data)
							}
						}
						pb.Objects = append(pb.Objects, picPO)
					}
				}
			}
			rowY += rowH
		}
	}
}

// autoManualBuild attempts to automatically build a ManualBuild table when the
// FRX specifies a ManualBuildEvent script but no Go callback is registered.
//
// It supports the standard column-first pattern (FixedColumns > 0):
//   - Prints FixedColumns header columns once.
//   - Detects the data source by scanning the first data column's cells for
//     [DataSourceAlias.Column] expressions.
//   - Iterates the detected data source and prints the data column once per row,
//     evaluating expressions at each row's context.
//   - Prints any remaining trailer columns once.
//
// Returns nil when the pattern cannot be detected or the data source is absent.
func (e *ReportEngine) autoManualBuild(tbl *table.TableObject) *table.TableBase {
	if tbl.ManualBuild != nil || tbl.ManualBuildEvent == "" {
		return nil
	}

	nCols := tbl.ColumnCount()
	if nCols < 2 {
		return nil
	}

	fixedCols := tbl.FixedColumns()
	if fixedCols <= 0 {
		return nil // only handle column-first pattern
	}

	dataColIdx := fixedCols // first non-fixed (data) column

	// Detect data source alias from expressions in the data column.
	dsAlias := ""
	for ri := 0; ri < tbl.RowCount(); ri++ {
		cell := tbl.Cell(ri, dataColIdx)
		if cell != nil {
			if a := tableExtractDSAlias(cell.Text()); a != "" {
				dsAlias = a
				break
			}
		}
	}
	if dsAlias == "" {
		return nil
	}

	dict := e.report.Dictionary()
	if dict == nil {
		return nil
	}
	ds := dict.FindDataSourceByAlias(dsAlias)
	if ds == nil {
		ds = dict.FindDataSourceByName(dsAlias)
	}
	if ds == nil {
		return nil
	}

	if err := ds.Init(); err != nil {
		return nil
	}
	if err := ds.First(); err != nil {
		// Empty data source: fall back to static template.
		return nil
	}

	h := table.NewTableHelper(tbl)

	// Print header columns (0 .. fixedCols-1).
	for i := 0; i < fixedCols; i++ {
		h.PrintColumn(i)
		h.PrintRows() // column-first: PrintRows fills rows for current column
	}

	// Print one data column per data source row.
	dataRowCount := 0
	for !ds.EOF() {
		e.report.SetCalcContext(ds)
		eval := func(text string) string { return e.evalText(text) }
		h.CellTextEval = eval
		// Evaluate PictureObject DataColumn bindings for embedded images.
		// For each PictureObject in the cell, create a NEW PictureObject with
		// the captured image data (since the template's PictureObject is shared).
		h.CellObjectEval = func(cell *table.TableCell) {
			objs := cell.Objects()
			for i, childObj := range objs {
				if pic, ok := childObj.(*object.PictureObject); ok {
					col := pic.DataColumn()
					if col == "" {
						continue
					}
					val, err := e.report.Calc("[" + col + "]")
					if err != nil {
						continue
					}
					s, ok := val.(string)
					if !ok || len(s) == 0 {
						continue
					}
					// Try standard base64 decoding first, then raw (no padding).
					decoded, decErr := base64.StdEncoding.DecodeString(s)
					if decErr != nil {
						decoded, decErr = base64.RawStdEncoding.DecodeString(s)
					}
					if decErr != nil || len(decoded) == 0 {
						continue
					}
					newPic := object.NewPictureObject()
					newPic.SetName(pic.Name())
					newPic.SetLeft(pic.Left())
					newPic.SetTop(pic.Top())
					newPic.SetWidth(pic.Width())
					newPic.SetHeight(pic.Height())
					newPic.SetImageData(decoded)
					cell.ReplaceObject(i, newPic)
				}
			}
		}
		h.PrintColumn(dataColIdx)
		h.PrintRows()
		h.CellTextEval = nil
		h.CellObjectEval = nil
		_ = ds.Next()
		dataRowCount++
	}

	// Print trailer columns (dataColIdx+1 .. nCols-1).
	// Substitute Count(CellName) aggregate with the data row count.
	countStr := fmt.Sprintf("%d", dataRowCount)
	for i := dataColIdx + 1; i < nCols; i++ {
		h.CellTextEval = func(text string) string {
			return countRE.ReplaceAllString(text, countStr)
		}
		h.PrintColumn(i)
		h.PrintRows()
		h.CellTextEval = nil
	}

	return h.Result()
}

// populateCrossViewGrid renders a CrossViewReportObject's result grid as
// individual PreparedObjects (one per grid cell). The cell positions are
// computed by dividing the object's Width/Height evenly across the grid
// columns/rows. When no CubeSource is bound the object has already been
// rendered as a bounding-box shape and this function is a no-op.
//
// Mirrors the C# CrossViewObject print pipeline which writes result cells into
// a TableResult and then calls GeneratePages (CrossViewObject.cs lines 444-463
// and TableBase.cs GeneratePages).
func (e *ReportEngine) populateCrossViewGrid(cv *crossview.CrossViewReportObject, pb *preview.PreparedBand) {
	if cv.CrossView.Source == nil {
		return
	}
	grid, err := cv.CrossView.Build()
	if err != nil || grid == nil || grid.RowCount == 0 || grid.ColCount == 0 {
		return
	}
	originX := cv.Left()
	originY := cv.Top()
	totalW := cv.Width()
	totalH := cv.Height()
	cellW := totalW / float32(grid.ColCount)
	cellH := totalH / float32(grid.RowCount)
	black := color.RGBA{A: 255}

	for r := 0; r < grid.RowCount; r++ {
		for c := 0; c < grid.ColCount; c++ {
			cell := grid.Cell(r, c)
			cs := cell.ColSpan
			if cs < 1 {
				cs = 1
			}
			rs := cell.RowSpan
			if rs < 1 {
				rs = 1
			}
			po := preview.PreparedObject{
				Name:      cv.Name(),
				Kind:      preview.ObjectTypeText,
				Left:      originX + float32(c)*cellW,
				Top:       originY + float32(r)*cellH,
				Width:     cellW * float32(cs),
				Height:    cellH * float32(rs),
				Text:      cell.Text,
				BlobIdx:   -1,
				Font:      style.DefaultFont(),
				TextColor: black,
				FillColor: color.RGBA{},
				Border:    cv.Border(),
			}
			pb.Objects = append(pb.Objects, po)
		}
	}
}

// tableExtractDSAlias extracts the data source alias from the first
// [Alias.Column] bracket expression found in text. Returns "" if not found.
func tableExtractDSAlias(text string) string {
	// Quick scan without regexp: find first '[', then look for '.' before ']'.
	for i := 0; i < len(text); i++ {
		if text[i] != '[' {
			continue
		}
		start := i + 1
		for j := start; j < len(text); j++ {
			if text[j] == ']' {
				break
			}
			if text[j] == '.' && j > start {
				alias := text[start:j]
				if strings.ContainsAny(alias, " \t") {
					break // skip qualified names with spaces for now
				}
				return alias
			}
		}
	}
	return ""
}

// populateAdvMatrixCells renders the physical table cells of an AdvMatrixObject
// as individual PreparedObjects. Each AdvMatrixRow holds AdvMatrixCell entries;
// column widths come from TableColumns.  Absolute positions are computed from
// the object's own Left/Top.
//
// ColSpan and RowSpan are resolved to pixel widths/heights using cumulative
// column/row offset arrays, matching the logic in populateTableObjects.
func (e *ReportEngine) populateAdvMatrixCells(adv *object.AdvMatrixObject, pb *preview.PreparedBand) {
	if len(adv.TableRows) == 0 {
		return
	}

	// Build cumulative column X offsets.
	colX := make([]float32, len(adv.TableColumns)+1)
	for i, col := range adv.TableColumns {
		colX[i+1] = colX[i] + col.Width
	}

	// Build cumulative row Y offsets for RowSpan height computation.
	rowYOff := make([]float32, len(adv.TableRows)+1)
	for ri, row := range adv.TableRows {
		h := row.Height
		if h <= 0 {
			h = 20
		}
		rowYOff[ri+1] = rowYOff[ri] + h
	}

	originX := adv.Left()
	originY := adv.Top()
	black := color.RGBA{A: 255}
	transparent := color.RGBA{R: 255, G: 255, B: 255, A: 0}

	for ri, row := range adv.TableRows {
		for ci, cell := range row.Cells {
			if cell == nil {
				continue
			}

			// Cell left: cumulative column offset.
			var cellX float32
			if ci < len(colX) {
				cellX = colX[ci]
			}

			// Cell width: span ColSpan columns.
			colSpan := cell.ColSpan
			if colSpan < 1 {
				colSpan = 1
			}
			endCol := ci + colSpan
			if endCol > len(colX)-1 {
				endCol = len(colX) - 1
			}
			cellW := colX[endCol] - colX[ci]

			// Cell height: span RowSpan rows.
			rowSpan := cell.RowSpan
			if rowSpan < 1 {
				rowSpan = 1
			}
			endRow := ri + rowSpan
			if endRow > len(rowYOff)-1 {
				endRow = len(rowYOff) - 1
			}
			cellH := rowYOff[endRow] - rowYOff[ri]

			text := e.evalText(cell.Text)
			fillColor := transparent
			if cell.FillColor != nil {
				fillColor = *cell.FillColor
			}
			po := preview.PreparedObject{
				Name:      cell.Name,
				Kind:      preview.ObjectTypeText,
				Left:      originX + cellX,
				Top:       originY + rowYOff[ri],
				Width:     cellW,
				Height:    cellH,
				Text:      text,
				BlobIdx:   -1,
				TextColor: black,
				FillColor: fillColor,
				HorzAlign: cell.HorzAlign,
				VertAlign: cell.VertAlign,
				WordWrap:  true,
			}
			if cell.Font != nil {
				po.Font = *cell.Font
			}
			if cell.Border != nil {
				po.Border = *cell.Border
			}
			pb.Objects = append(pb.Objects, po)
		}
	}
}

// populateCellularTextCells renders a CellularTextObject as a character grid of
// individual PreparedObjects. The algorithm mirrors FastReport .NET's GetTable()
// method in CellularTextObject.cs:
//
//   - If CellWidth/CellHeight are 0, auto-size from the font height.
//   - Compute colCount = (Width + HorzSpacing + 1) / (cellWidth + HorzSpacing).
//   - Layout text with optional word-wrap, one character per column per row.
//   - Apply HorzAlign offset per row (right/center push the characters).
//   - Emit one PreparedObject per cell with the character centered inside.
func (e *ReportEngine) populateCellularTextCells(v *object.CellularTextObject, text string, pb *preview.PreparedBand) {
	cellW := v.CellWidth()
	cellH := v.CellHeight()
	horzSpacing := v.HorzSpacing()
	vertSpacing := v.VertSpacing()

	// Auto-calculate cell size from font height when not explicitly set.
	// Mirrors GetCellWidthInternal: round (fontHeight + 10) up to nearest 0.25 cm.
	if cellW == 0 || cellH == 0 {
		// Font size in pt → pixels at 96 DPI: px = pt * 96/72
		fontPx := v.Font().Size * 96.0 / 72.0
		// 0.25 cm in pixels = 37.8 * 0.25 = 9.45
		quarterCm := float32(9.45)
		raw := fontPx + 10
		auto := float32(math.Round(float64(raw)/float64(quarterCm))) * quarterCm
		if auto <= 0 {
			auto = quarterCm
		}
		if cellW == 0 {
			cellW = auto
		}
		if cellH == 0 {
			cellH = auto
		}
	}

	// Compute column count: how many cells fit in the total width.
	totalW := v.Width()
	colCount := int((totalW + horzSpacing + 1) / (cellW + horzSpacing))
	if colCount < 1 {
		colCount = 1
	}

	// Compute row count: how many rows fit in the total height.
	totalH := v.Height()
	rowCount := int((totalH + vertSpacing + 1) / (cellH + vertSpacing))
	if rowCount < 1 {
		rowCount = 1
	}

	// Build grid: grid[row][col] holds the character for that cell (empty string = blank).
	grid := make([][]string, rowCount)
	for i := range grid {
		grid[i] = make([]string, colCount)
	}

	// Fill grid with text, handling word-wrap and CRLF.
	normalized := strings.ReplaceAll(text, "\r\n", "\n")
	row := 0
	lineBegin := 0
	lastSpace := 0
	runes := []rune(normalized)
	wordWrap := v.WordWrap()
	horzAlign := v.HorzAlign()

	fillRow := func(rowIdx int, line []rune) {
		if rowIdx >= rowCount {
			return
		}
		// Trim trailing spaces.
		for len(line) > 0 && line[len(line)-1] == ' ' {
			line = line[:len(line)-1]
		}
		if len(line) > colCount {
			line = line[:colCount]
		}
		// Compute HorzAlign offset.
		offset := 0
		switch horzAlign {
		case object.HorzAlignRight:
			offset = colCount - len(line)
		case object.HorzAlignCenter:
			offset = (colCount - len(line)) / 2
		}
		if offset < 0 {
			offset = 0
		}
		for i, ch := range line {
			col := i + offset
			if col >= 0 && col < colCount {
				grid[rowIdx][col] = string(ch)
			}
		}
	}

	for i := 0; i < len(runes); i++ {
		isCRLF := runes[i] == '\n'
		if runes[i] == ' ' || isCRLF {
			lastSpace = i
		}

		if i-lineBegin+1 > colCount || isCRLF {
			if wordWrap && lastSpace > lineBegin {
				fillRow(row, runes[lineBegin:lastSpace])
				lineBegin = lastSpace + 1
			} else if i-lineBegin > 0 {
				fillRow(row, runes[lineBegin:i])
				lineBegin = i
			} else {
				lineBegin = i + 1
			}
			lastSpace = lineBegin
			row++
		}
	}
	// Finish the last line.
	if lineBegin < len(runes) {
		fillRow(row, runes[lineBegin:])
	}

	// Emit PreparedObjects for each cell.
	originX := v.Left()
	originY := v.Top()
	border := v.Border()
	font := v.Font()
	textColor := v.TextColor()
	var fillColor color.RGBA
	if f, ok := v.Fill().(*style.SolidFill); ok {
		fillColor = f.Color
	}

	// Default border for spacing cells: no visible lines (matches C# default TableCell).
	spacerBorder := *style.NewBorder()
	// Spacer cells in C# inherit the TABLE's default font (TextObjectBase default = 10pt).
	// They also have WordWrap=true and Clip=true (C# TableCell defaults).
	spacerFont := style.DefaultFont()

	for ri := 0; ri < rowCount; ri++ {
		cellTop := originY + float32(ri)*(cellH+vertSpacing)
		for ci := 0; ci < colCount; ci++ {
			cellLeft := originX + float32(ci)*(cellW+horzSpacing)
			po := preview.PreparedObject{
				Name:      fmt.Sprintf("%s_r%dc%d", v.Name(), ri, ci),
				Kind:      preview.ObjectTypeText,
				Left:      cellLeft,
				Top:       cellTop,
				Width:     cellW,
				Height:    cellH,
				Text:      grid[ri][ci],
				BlobIdx:   -1,
				Font:      font,
				TextColor: textColor,
				FillColor: fillColor,
				HorzAlign: 1, // Center — each character is always centered in its cell
				VertAlign: 1, // Center
				WordWrap:  true, // C# TableCell defaults to WordWrap=true
				Clip:      true, // C# TableCell defaults to Clip=true
				Border:    border,
			}
			pb.Objects = append(pb.Objects, po)

			// Emit horizontal spacing cell to the right of this character cell.
			// C# inserts spacing columns at odd indices after setting cell properties.
			if horzSpacing > 0 && ci < colCount-1 {
				spacer := preview.PreparedObject{
					Name:     fmt.Sprintf("%s_r%dhs%d", v.Name(), ri, ci),
					Kind:     preview.ObjectTypeText,
					Left:     cellLeft + cellW,
					Top:      cellTop,
					Width:    horzSpacing,
					Height:   cellH,
					BlobIdx:  -1,
					Border:   spacerBorder,
					Font:     spacerFont,
					WordWrap: true,
					Clip:     true,
				}
				pb.Objects = append(pb.Objects, spacer)
			}
		}

		// Emit vertical spacing row below this character row.
		if vertSpacing > 0 && ri < rowCount-1 {
			spacerTop := cellTop + cellH
			for ci := 0; ci < colCount; ci++ {
				cellLeft := originX + float32(ci)*(cellW+horzSpacing)
				spacer := preview.PreparedObject{
					Name:     fmt.Sprintf("%s_vs%dc%d", v.Name(), ri, ci),
					Kind:     preview.ObjectTypeText,
					Left:     cellLeft,
					Top:      spacerTop,
					Width:    cellW,
					Height:   vertSpacing,
					BlobIdx:  -1,
					Border:   spacerBorder,
					Font:     spacerFont,
					WordWrap: true,
					Clip:     true,
				}
				pb.Objects = append(pb.Objects, spacer)

				// Intersection spacer (bottom-right corner between cells).
				if horzSpacing > 0 && ci < colCount-1 {
					spacer := preview.PreparedObject{
						Name:     fmt.Sprintf("%s_vs%dhs%d", v.Name(), ri, ci),
						Kind:     preview.ObjectTypeText,
						Left:     cellLeft + cellW,
						Top:      spacerTop,
						Width:    horzSpacing,
						Height:   vertSpacing,
						BlobIdx:  -1,
						Border:   spacerBorder,
						Font:     spacerFont,
						WordWrap: true,
						Clip:     true,
					}
					pb.Objects = append(pb.Objects, spacer)
				}
			}
		}
	}
}

// makeDataFinishedPred returns a sender predicate for ProcessAtDataFinished.
// If our band's parent is a DataBand (e.g. we are a DataHeaderBand), the
// handler should only fire when that specific DataBand sends BlockFinished.
// Mirrors C# ProcessInfo.Process lines 123-131 (sender-check for DataFinished).
func makeDataFinishedPred(parentBand *band.BandBase) func(any) bool {
	return func(sender any) bool {
		senderDB, ok := sender.(*band.DataBand)
		if !ok {
			return false
		}
		if parentBand == nil {
			return true
		}
		// If our band's parent is a DataBand (we're a child header/footer band),
		// only process when the sender is exactly that DataBand.
		if parentDB, ok2 := parentBand.Parent().(*band.DataBand); ok2 {
			return senderDB == parentDB
		}
		return true
	}
}

// makeGroupFinishedPred returns a sender predicate for ProcessAtGroupFinished.
// The handler fires only when the GroupHeaderBand that owns our band fires GroupFinished.
// Mirrors C# ProcessInfo.Process lines 133-137 (sender == topParentBand).
func makeGroupFinishedPred(parentBand *band.BandBase) func(any) bool {
	return func(sender any) bool {
		senderGH, ok := sender.(*band.GroupHeaderBand)
		if !ok {
			return false
		}
		if parentBand == nil {
			return true
		}
		// If our band's parent is a GroupHeaderBand, only process when sender matches.
		if parentGH, ok2 := parentBand.Parent().(*band.GroupHeaderBand); ok2 {
			return senderGH == parentGH
		}
		return true
	}
}

// processAtToEngineState maps a ProcessAt value to its corresponding EngineState.
func processAtToEngineState(pa object.ProcessAt) EngineState {
	switch pa {
	case object.ProcessAtReportFinished:
		return EngineStateReportFinished
	case object.ProcessAtReportPageFinished:
		return EngineStateReportPageFinished
	case object.ProcessAtPageFinished:
		return EngineStatePageFinished
	case object.ProcessAtColumnFinished:
		return EngineStateColumnFinished
	case object.ProcessAtDataFinished:
		return EngineStateBlockFinished
	case object.ProcessAtGroupFinished:
		return EngineStateGroupFinished
	default:
		return EngineStateReportFinished
	}
}

// buildPreparedObject converts a single report.Base object into a PreparedObject,
// or returns nil if the object type is not renderable (e.g. a nested band).
func (e *ReportEngine) buildPreparedObject(obj report.Base) *preview.PreparedObject {
	// Skip invisible objects.
	// If VisibleExpression is set it overrides the static Visible flag,
	// matching C# ComponentBase.CalcVisibleExpression behaviour (ComponentBase.cs:536).
	type hasVisibleExpr interface {
		Visible() bool
		VisibleExpression() string
		CalcVisibleExpression(expression string, calc func(string) (any, error)) bool
	}
	if v, ok := obj.(hasVisibleExpr); ok {
		expr := v.VisibleExpression()
		if expr != "" {
			// Expression present: evaluate it; result overrides static Visible.
			if e.report != nil {
				visible := v.CalcVisibleExpression(expr, func(s string) (any, error) {
					return e.report.Calc(s)
				})
				if !visible {
					return nil
				}
			} else if !v.Visible() {
				return nil
			}
		} else if !v.Visible() {
			return nil
		}
	}

	// Evaluate PrintableExpression and check Printable flag.
	// Mirrors C# ReportEngine.Bands.cs CanPrint lines 299-313.
	// Non-printable objects are excluded from the prepared output (they appear
	// on screen but not in print/export).
	type hasPrintable interface {
		Printable() bool
		PrintableExpression() string
		SetPrintable(bool)
	}
	if p, ok := obj.(hasPrintable); ok {
		if expr := p.PrintableExpression(); expr != "" && e.report != nil {
			if val, err := e.report.Calc(expr); err == nil {
				if b, ok2 := val.(bool); ok2 {
					p.SetPrintable(b)
				}
			}
		}
		if !p.Printable() {
			return nil
		}
	}

	// Geometry accessors common to all component objects.
	type hasGeom interface {
		Left() float32
		Top() float32
		Width() float32
		Height() float32
	}

	geom, ok := obj.(hasGeom)
	if !ok {
		return nil // no geometry = not a renderable object
	}

	// Evaluate ExportableExpression and snapshot Exportable flag.
	// Mirrors C# ReportEngine.Bands.cs lines 287-296 (CanPrint block).
	// Unlike Printable (which excludes from PreparedPages), Exportable is
	// snapshotted so exporters can choose to skip non-exportable objects.
	type hasExportable interface {
		Exportable() bool
		ExportableExpression() string
		SetExportable(bool)
	}
	notExportable := false
	if ex, ok := obj.(hasExportable); ok {
		if expr := ex.ExportableExpression(); expr != "" && e.report != nil {
			if val, err := e.report.Calc(expr); err == nil {
				if b, ok2 := val.(bool); ok2 {
					ex.SetExportable(b)
				}
			}
		}
		notExportable = !ex.Exportable()
	}

	po := &preview.PreparedObject{
		Name:          obj.Name(),
		Left:          geom.Left(),
		Top:           geom.Top(),
		Width:         geom.Width(),
		Height:        geom.Height(),
		BlobIdx:       -1,
		Font:          style.DefaultFont(),
		NotExportable: notExportable,
	}

	// Populate hyperlink from ReportComponentBase if present.
	type hasHyperlink interface {
		Hyperlink() *report.Hyperlink
	}
	if hld, ok := obj.(hasHyperlink); ok {
		if hl := hld.Hyperlink(); hl != nil {
			switch hl.Kind {
			case "URL", "": // C# default HyperlinkKind is URL
				po.HyperlinkKind = 1
				val := hl.Value
				if val == "" && hl.Expression != "" && e.report != nil {
					if ev, err := e.report.Calc(hl.Expression); err == nil {
						val = fmt.Sprintf("%v", ev)
					}
				}
				po.HyperlinkValue = val
			case "PageNumber":
				// C# HyperlinkKind.PageNumber; Go uses 2 (preview/prepared_pages.go:440).
				// Value is the target page number (static or expression-evaluated).
				// C# ref: HTMLExportLayers.cs:167 — navigate to page N in the report.
				po.HyperlinkKind = 2
				val := hl.Value
				if val == "" && hl.Expression != "" && e.report != nil {
					if ev, err := e.report.Calc(hl.Expression); err == nil {
						val = fmt.Sprintf("%v", ev)
					}
				}
				po.HyperlinkValue = val
			case "Bookmark":
				po.HyperlinkKind = 3
				po.HyperlinkValue = hl.Expression
			}
			// Carry the anchor target (e.g. "_blank") through to the prepared object.
			// C# reference: Hyperlink.Target / OpenLinkInNewTab → target="_blank".
			po.HyperlinkTarget = hl.Target
		}
	}

	switch v := obj.(type) {
	case *object.HtmlObject:
		po.Kind = preview.ObjectTypeHtml
		po.TextColor = color.RGBA{A: 255}
		po.Border = v.Border()
		if f, ok := v.Fill().(*style.SolidFill); ok && f.Color.A > 0 {
			po.FillColor = f.Color
		}
		po.Text = e.evalText(v.Text())

	case *object.TextObject:
		po.Kind = preview.ObjectTypeText
		po.Angle = v.Angle()
		po.Font = v.Font()
		po.TextColor = v.TextColor() // style-applied or default black
		po.HorzAlign = int(v.HorzAlign())
		po.VertAlign = int(v.VertAlign())
		po.WordWrap = v.WordWrap()
		po.Border = v.Border()
		if f, ok := v.Fill().(*style.SolidFill); ok && f.Color.A > 0 {
			po.FillColor = f.Color
		}
		po.Text = e.evalTextWithFormat(v.Text(), v.Format())
		// When the text color has alpha=0 (fully transparent), the text is
		// invisible. Suppress the text content so the object renders as a
		// background-only shape, matching C# HTML export behaviour.
		if po.TextColor.A == 0 && po.Text != "" {
			po.Text = ""
		}
		po.TextRenderType = int(v.TextRenderType())
		// Padding, paragraph offset, line height, RTL, Clip.
		pad := v.Padding()
		po.PaddingLeft = pad.Left
		po.PaddingTop = pad.Top
		po.PaddingRight = pad.Right
		po.PaddingBottom = pad.Bottom
		po.Trimming = int(v.Trimming())
		po.ForceJustify = v.ForceJustify()
		po.ParagraphOffset = v.ParagraphOffset()
		po.LineHeight = v.LineHeight()
		// ParagraphFormat — mirrors C# context.paragraphFormat assignment (TextObject.cs line 1199).
		pf := v.ParagraphFormat()
		po.ParagraphFirstLineIndent = pf.FirstLineIndent
		po.ParagraphLineSpacing = pf.LineSpacing
		po.ParagraphLineSpacingType = int(pf.LineSpacingType)
		po.RTL = v.RightToLeft()
		po.Clip = v.Clip()
		// Apply highlight conditions — first matching condition wins.
		if e.report != nil {
			for _, cond := range v.Highlights() {
				result, err := e.report.Calc(cond.Expression)
				if err != nil {
					continue
				}
				matched, _ := result.(bool)
				if !matched {
					continue
				}
				if !cond.Visible {
					return nil
				}
				if cond.ApplyFill {
					po.FillColor = cond.FillColor
				}
				if cond.ApplyTextFill {
					po.TextColor = cond.TextFillColor
				}
				if cond.ApplyFont {
					po.Font = cond.Font
				}
				break
			}
		}
		// AutoWidth: measure the evaluated text and shrink the width to fit.
		// C# ref: TextObject.Serialize (TextObject.cs:1383-1392) — applied when
		// SerializeTo == Preview. CalcSize() returns measured width/height.
		if v.AutoWidth() && po.Text != "" {
			po.WordWrap = false
			textW, _ := utils.MeasureText(po.Text, po.Font, 0)
			// Scale from basicfont's monospace widths to the target font's proportional widths.
			textW = utils.ScaleWidth(textW, po.Font)
			textW += po.PaddingLeft + po.PaddingRight
			if v.HorzAlign() == object.HorzAlignRight {
				po.Left += po.Width - textW
			} else if v.HorzAlign() == object.HorzAlignCenter {
				po.Left += po.Width/2 - textW/2
			}
			po.Width = textW
		}

	case *object.ContainerObject:
		// Render the container background as a rectangle; children are added
		// separately by populateContainerChildren with coordinate offsets.
		po.Kind = preview.ObjectTypeShape
		po.ShapeKind = 0 // Rectangle
		po.Border = v.Border()
		if f, ok := v.Fill().(*style.SolidFill); ok {
			po.FillColor = f.Color
		}

	case *object.PolyLineObject:
		po.Kind = preview.ObjectTypePolyLine
		po.Border = v.Border()
		if f, ok := v.Fill().(*style.SolidFill); ok {
			po.FillColor = f.Color
		}
		pts := v.Points()
		for i := 0; i < pts.Len(); i++ {
			pt := pts.Get(i)
			po.Points = append(po.Points, [2]float32{pt.X, pt.Y})
		}

	case *object.PolygonObject:
		po.Kind = preview.ObjectTypePolygon
		po.Border = v.Border()
		if f, ok := v.Fill().(*style.SolidFill); ok {
			po.FillColor = f.Color
		}
		pts := v.Points()
		for i := 0; i < pts.Len(); i++ {
			pt := pts.Get(i)
			po.Points = append(po.Points, [2]float32{pt.X, pt.Y})
		}

	case *object.LineObject:
		po.Kind = preview.ObjectTypeLine
		po.LineDiagonal = v.Diagonal()
		po.Border = v.Border()
		// Ensure horizontal lines have at least 1px height so the border is visible.
		// In C#, LineObject with Height=0 renders with the border line width.
		// C# ref: LineObject.cs Draw() uses Height for rendering calculations.
		if !v.Diagonal() && po.Height == 0 && po.Width > 0 {
			lw := float32(1)
			if v.Border().Lines[0] != nil && v.Border().Lines[0].Width > 0 {
				lw = v.Border().Lines[0].Width
			}
			po.Height = lw
		}
		if f, ok := v.Fill().(*style.SolidFill); ok {
			po.FillColor = f.Color
		}
		// Propagate cap settings so exporters can render caps.
		// Mirrors C# LineObject.GetConvertedObjects() cap path (LineObject.OpenSource.cs).
		po.LineStartCap = preview.LineCap{
			Style:  preview.LineCapStyle(v.StartCap.Style),
			Width:  v.StartCap.Width,
			Height: v.StartCap.Height,
		}
		po.LineEndCap = preview.LineCap{
			Style:  preview.LineCapStyle(v.EndCap.Style),
			Width:  v.EndCap.Width,
			Height: v.EndCap.Height,
		}

	case *object.ShapeObject:
		po.Kind = preview.ObjectTypeShape
		po.ShapeKind = int(v.Shape())
		po.ShapeCurve = v.Curve()
		po.Border = v.Border()
		if f, ok := v.Fill().(*style.SolidFill); ok {
			po.FillColor = f.Color
		}

	case *object.PictureObject:
		po.Kind = preview.ObjectTypePicture
		po.Border = v.Border()
		if f, ok := v.Fill().(*style.SolidFill); ok && f.Color.A > 0 {
			po.FillColor = f.Color
		}
		// Evaluate DataColumn binding to load image data from the current data
		// source row. C# ref: PictureObject.GetData() (PictureObject.cs:705-736)
		// calls Report.GetColumnValueNullable(DataColumn) which returns byte[].
		if col := v.DataColumn(); col != "" && e.report != nil {
			val, err := e.report.Calc("[" + col + "]")
			if err == nil && val != nil {
				switch d := val.(type) {
				case []byte:
					v.SetImageData(d)
				case string:
					if len(d) > 0 {
						// Try base64 decode (XML data sources store binary as base64).
						if decoded, decErr := base64.StdEncoding.DecodeString(d); decErr == nil && len(decoded) > 0 {
							v.SetImageData(decoded)
						} else if decoded, decErr = base64.RawStdEncoding.DecodeString(d); decErr == nil && len(decoded) > 0 {
							v.SetImageData(decoded)
						}
					}
				}
			}
		}
		if data := v.ImageData(); len(data) > 0 {
			// Apply grayscale and/or transparency transforms before storing in BlobStore.
			// C# ref: PictureObjectBase.Draw() calls ImageHelper.GetGrayscaleBitmap /
			//         ImageHelper.GetTransparentBitmap when the flags are set.
			//         FastReport.Base/Utils/ImageHelper.cs GetGrayscaleBitmap, GetTransparentBitmap
			if v.Grayscale() || v.Transparency() > 0 {
				if img, _, err := image.Decode(bytes.NewReader(data)); err == nil {
					if v.Grayscale() {
						img = utils.ApplyGrayscale(img)
					}
					if v.Transparency() > 0 {
						img = utils.ApplyTransparency(img, v.Transparency())
					}
					var buf bytes.Buffer
					if err := png.Encode(&buf, img); err == nil {
						data = buf.Bytes()
					}
				}
			}
			// Use content hash as dedup key, matching C# PictureObject.Serialize
			// which calls BlobStore.AddOrUpdate(bytes, Murmur3.ComputeHash(bytes)).
			// Using the object name would incorrectly dedup different images from
			// data-bound PictureObjects (e.g. employee photos).
			imgHash := utils.ComputeHashBytes(data)
			po.BlobIdx = e.preparedPages.BlobStore.Add(imgHash, data)
		}

	case *object.CellularTextObject:
		// CellularTextObject is rendered as a grid of individual character cells
		// by populateCellularTextCells (called from populateBandObjects2).
		// In C#, GetTable() creates a TableObject that renders a container div
		// (table background with no border, default font) plus individual cell divs.
		// The container div must be present to maintain the 1:1 mapping between FRX
		// objects and pb.Objects (used by band layout height adjustment in calcBandLayout).
		// Container properties match C# TableObject defaults: no border, transparent fill,
		// default font (TextObjectBase default = 10pt), empty text.
		po.Kind = preview.ObjectTypeText
		po.Text = ""
		po.WordWrap = true
		po.Clip = true
		po.Border = *style.NewBorder() // no visible lines (BorderLinesNone)
		// Compute container dimensions from actual table size (matching C# GetTable):
		// table.Width = colCount*cellW + (colCount-1)*horzSpacing
		// table.Height = rowCount*cellH + (rowCount-1)*vertSpacing
		{
			cw, ch := v.CellWidth(), v.CellHeight()
			if cw == 0 || ch == 0 {
				fontPx := v.Font().Size * 96.0 / 72.0
				qcm := float32(9.45)
				auto := float32(math.Round(float64(fontPx+10)/float64(qcm))) * qcm
				if auto <= 0 {
					auto = qcm
				}
				if cw == 0 {
					cw = auto
				}
				if ch == 0 {
					ch = auto
				}
			}
			hs, vs := v.HorzSpacing(), v.VertSpacing()
			cc := int((v.Width() + hs + 1) / (cw + hs))
			if cc < 1 {
				cc = 1
			}
			rc := int((v.Height() + vs + 1) / (ch + vs))
			if rc < 1 {
				rc = 1
			}
			po.Width = float32(cc)*cw + float32(max(0, cc-1))*hs
			po.Height = float32(rc)*ch + float32(max(0, rc-1))*vs
		}

	case *object.ZipCodeObject:
		// Evaluate DataColumn / Expression to update the Text value before
		// rendering, mirroring C# ZipCodeObject.GetData / GetDataShared()
		// (ZipCodeObject.cs line 338-356):
		//   if DataColumn != "" → Report.GetColumnValue(DataColumn)
		//   else if Expression != "" → Report.Calc(Expression)
		if col := v.DataColumn(); col != "" {
			v.SetText(e.evalText("[" + col + "]"))
		} else if expr := v.Expression(); expr != "" {
			v.SetText(e.evalText(expr))
		}
		// Render as a picture (C# LayerBack + LayerPicture pattern).
		// C# Draw() recalculates dimensions (ZipCodeObject.cs line 268-269).
		po.Kind = preview.ObjectTypePicture
		brd := v.Border()
		borderColor := brd.Color()
		borderWidth := brd.Lines[0].Width
		drawBorder := brd.VisibleLines != style.BorderLinesNone
		zipW, zipH := object.ZipCodeDimensions(v, borderWidth)
		po.Width = zipW
		po.Height = zipH
		if f, ok := v.Fill().(*style.SolidFill); ok && f.Color.A > 0 {
			po.FillColor = f.Color
		}
		po.BlobIdx = e.renderGaugeBlob(v.Name(),
			object.RenderZipCode(v, borderColor, borderWidth, drawBorder, po.FillColor,
				int(zipW), int(zipH)))

	case *object.CheckBoxObject:
		po.Kind = preview.ObjectTypeCheckBox
		po.Border = v.Border()
		// Evaluate expression or data column binding to determine checked state.
		if expr := v.Expression(); expr != "" {
			result, err := e.report.Calc(expr)
			if err == nil {
				v.SetChecked(anyToBool(result))
			}
		} else if col := v.DataColumn(); col != "" {
			result, err := e.report.Calc("[" + col + "]")
			if err == nil {
				v.SetChecked(anyToBool(result))
			} else {
				s := e.evalText("[" + col + "]")
				v.SetChecked(s == "true" || s == "True" || s == "1")
			}
		}
		po.Checked = v.Checked()
		po.CheckedSymbol = int(v.CheckedSymbol())
		po.UncheckedSymbol = int(v.UncheckedSymbol())
		po.CheckColor = v.CheckColor()

	case *gauge.LinearGauge:
		e.evalGaugeValue(&v.GaugeObject)
		po.Kind = preview.ObjectTypePicture
		po.Border = v.Border()
		if f, ok := v.Fill().(*style.SolidFill); ok && f.Color.A > 0 {
			po.FillColor = f.Color
		}
		po.BlobIdx = e.renderGaugeBlob(v.Name(), gauge.RenderLinear(v, int(geom.Width()), int(geom.Height())))

	case *gauge.RadialGauge:
		e.evalGaugeValue(&v.GaugeObject)
		po.Kind = preview.ObjectTypePicture
		po.Border = v.Border()
		if f, ok := v.Fill().(*style.SolidFill); ok && f.Color.A > 0 {
			po.FillColor = f.Color
		}
		po.BlobIdx = e.renderGaugeBlob(v.Name(), gauge.RenderRadial(v, int(geom.Width()), int(geom.Height())))

	case *gauge.SimpleGauge:
		e.evalGaugeValue(&v.GaugeObject)
		po.Kind = preview.ObjectTypePicture
		po.Border = v.Border()
		if f, ok := v.Fill().(*style.SolidFill); ok && f.Color.A > 0 {
			po.FillColor = f.Color
		}
		po.BlobIdx = e.renderGaugeBlob(v.Name(), gauge.RenderSimple(v, int(geom.Width()), int(geom.Height())))

	case *gauge.SimpleProgressGauge:
		e.evalGaugeValue(&v.GaugeObject)
		po.Kind = preview.ObjectTypePicture
		po.Border = v.Border()
		if f, ok := v.Fill().(*style.SolidFill); ok && f.Color.A > 0 {
			po.FillColor = f.Color
		}
		po.BlobIdx = e.renderGaugeBlob(v.Name(), gauge.RenderSimpleProgress(v, int(geom.Width()), int(geom.Height())))

	case *object.RichObject:
		// Preserve the raw RTF content so that exporters with RTF support
		// (HTML) can render formatting. Exporters that do not support RTF
		// (PDF, image) fall back to plain text via StripRTF at export time.
		po.Kind = preview.ObjectTypeRTF
		po.TextColor = color.RGBA{A: 255}
		po.Text = e.evalText(v.Text())

	case *object.SVGObject:
		// Store the raw SVG XML in BlobStore.  HTML exporters embed it inline;
		// PDF/image exporters draw a placeholder box.
		po.Kind = preview.ObjectTypeSVG
		if v.SvgData != "" {
			svgBytes := decodeSvgData(v.SvgData)
			if len(svgBytes) > 0 && e.preparedPages != nil {
				po.BlobIdx = e.preparedPages.BlobStore.Add(v.Name(), svgBytes)
			}
		}

	case *object.SparklineObject:
		po.Kind = preview.ObjectTypePicture
		if series := sparkline.DecodeChartData(v.ChartData); series != nil {
			img := sparkline.Render(series, int(geom.Width()), int(geom.Height()))
			if img != nil {
				po.BlobIdx = e.renderGaugeBlob(v.Name(), img)
			}
		}

	case *matrix.MatrixObject:
		// MatrixObject cells are rendered by populateBandObjects2.
		// The anchor exists only for FRX→PreparedObject index mapping.
		po.Kind = preview.ObjectTypeText
		po.NotExportable = true

	case *object.AdvMatrixObject:
		// Render the physical table layout as a grid of text PreparedObjects.
		po.Kind = preview.ObjectTypeShape
		po.ShapeKind = 0
		// The individual cells will be emitted below via populateAdvMatrixCells.

	case *object.RFIDLabel:
		// RFIDLabel: rendered as a dashed-border placeholder box showing EPC data.
		// Actual RFID encoding is performed by the printer driver at print time.
		po.Kind = preview.ObjectTypeDigitalSignature // reuse the dashed placeholder style
		po.TextColor = color.RGBA{A: 255}
		po.Text = v.PlaceholderText()

	case *object.DigitalSignatureObject:
		// ObjectTypeDigitalSignature: PDF exporters create a /Sig Widget annotation;
		// other exporters render a styled placeholder box with the placeholder text.
		po.Kind = preview.ObjectTypeDigitalSignature
		po.TextColor = color.RGBA{A: 255}
		po.Text = v.Placeholder()

	case *barcode.BarcodeObject:
		po.Kind = preview.ObjectTypePicture
		po.Angle = v.Angle()
		po.Border = v.Border()
		if f, ok := v.Fill().(*style.SolidFill); ok && f.Color.A > 0 {
			po.FillColor = f.Color
		}
		// Evaluate the barcode text: DataColumn → Expression → Text.
		// C# BarcodeObject.GetDataShared() (BarcodeObject.cs:601-604):
		//   if DataColumn != "" → Report.GetColumnValue(DataColumn)
		//   else if Expression != "" → Report.Calc(Expression)
		//   else evaluate bracket expressions in Text
		var text string
		if col := v.DataColumn(); col != "" {
			text = e.evalText("[" + col + "]")
		} else if v.Expression() != "" {
			text = e.evalText(v.Expression())
		} else {
			text = e.evalText(v.Text())
		}
		// Apply Trim: C# LinearBarcodeBase.Initialize() trims whitespace when Trim=true.
		if v.Trim() {
			text = strings.TrimSpace(text)
		}
		if text == "" && !v.HideIfNoData() {
			text = v.NoDataText()
		}
		// Fall back to the barcode's built-in default value when text is still
		// empty and HideIfNoData is false (matches FastReport .NET behaviour:
		// always show a demo barcode when no data is bound).
		if text == "" && !v.HideIfNoData() && v.Barcode != nil {
			text = v.Barcode.DefaultValue()
		}
		if text != "" && v.Barcode != nil {
			if err := v.Barcode.Encode(text); err == nil {
				// Apply auto-size so CalcBounds dimensions are reflected in po.Width/Height
				v.UpdateAutoSize()
				w := int(v.Width())
				h := int(v.Height())
				if w <= 0 {
					w = 200
				}
				if h <= 0 {
					h = 60
				}
				// Update po dimensions to reflect auto-size
				po.Width = v.Width()
				po.Height = v.Height()
				img, err := renderBarcode(v.Barcode, w, h)
				if err == nil && img != nil && e.preparedPages != nil {
					var buf bytes.Buffer
					if encErr := png.Encode(&buf, img); encErr == nil {
						// Use empty source to prevent deduplication by name.
					// Barcodes are dynamically rendered per data row; reusing the
					// same blob for all rows would show every badge with the same QR code.
					po.BlobIdx = e.preparedPages.BlobStore.Add("", buf.Bytes())
					}
					// Also capture the module bit-matrix for vector rendering.
					po.IsBarcode = true
					po.BarcodeModules = extractBarcodeModules(img)
				}
			}
		}

	case *object.MapObject:
		po.Kind = preview.ObjectTypePicture
		// Build maprender.Options from MapObject layers.
		opts := maprender.Options{
			Width:   int(geom.Width()),
			Height:  int(geom.Height()),
			OffsetX: float64(v.OffsetX),
			OffsetY: float64(v.OffsetY),
		}
		for _, layer := range v.Layers {
			opts.Layers = append(opts.Layers, maprender.Layer{
				Shapefile: layer.Shapefile,
				Palette:   layer.Palette,
				Type:      layer.Type,
			})
		}
		if opts.Width <= 0 {
			opts.Width = 400
		}
		if opts.Height <= 0 {
			opts.Height = 200
		}
		img := maprender.Render(opts)
		if img != nil && e.preparedPages != nil {
			var buf bytes.Buffer
			if encErr := png.Encode(&buf, img); encErr == nil {
				po.BlobIdx = e.preparedPages.BlobStore.Add(v.Name(), buf.Bytes())
			}
		}

	case *object.MSChartObject:
		po.Kind = preview.ObjectTypePicture
		img := v.RenderToImage(int(geom.Width()), int(geom.Height()))
		if img != nil {
			po.BlobIdx = e.renderGaugeBlob(v.Name(), img)
		}

	case *crossview.CrossViewReportObject:
		// The bounding box is rendered as a shape; individual cells are added
		// by populateCrossViewGrid called from populateBandObjects2.
		// Mirrors C# CrossViewObject which inherits TableBase (a breakable table).
		po.Kind = preview.ObjectTypeShape
		po.ShapeKind = 0 // Rectangle
		po.Border = v.Border()
		if f, ok := v.Fill().(*style.SolidFill); ok {
			po.FillColor = f.Color
		}

	case *table.TableObject:
		// Individual cells are rendered by populateTableObjects (called from
		// populateBandObjects2). The anchor exists only to maintain the 1:1
		// FRX→PreparedObject mapping for band layout height adjustment.
		// In C#, the table container is generated per wrapped section by
		// GeneratePages, not as a single anchor div.
		po.Kind = preview.ObjectTypeText
		po.NotExportable = true
		po.Border = v.Border()
		if f, ok := v.Fill().(*style.SolidFill); ok {
			po.FillColor = f.Color
		}

	default:
		// Not a known renderable type (could be a nested band etc.)
		return nil
	}

	// Register bookmark if the component has one, and carry the name to the
	// PreparedObject so that HTML exporters can emit <a name="..."> anchors.
	// C# reference: HTMLExportLayers.cs ExportObject → obj.Bookmark → <a name="...">.
	type hasBookmark interface{ Bookmark() string }
	if bk, ok := obj.(hasBookmark); ok {
		if name := bk.Bookmark(); name != "" {
			resolved := e.evalText(name)
			e.AddBookmark(resolved)
			po.Bookmark = resolved
		}
	}

	return po
}

// evalGaugeText evaluates a gauge expression and formats the result.
// If the expression evaluates successfully the result is returned as a string;
// otherwise the raw default value is shown.
func (e *ReportEngine) evalGaugeText(expr string, defaultVal float64) string {
	if expr == "" {
		return fmt.Sprintf("%g", defaultVal)
	}
	if e.report != nil {
		if result, err := e.report.Calc(expr); err == nil {
			return fmt.Sprintf("%v", result)
		}
	}
	return fmt.Sprintf("%g", defaultVal)
}

// evalGaugeValue evaluates the gauge Expression and updates SetValue if successful.
func (e *ReportEngine) evalGaugeValue(g *gauge.GaugeObject) {
	if g.Expression == "" || e.report == nil {
		return
	}
	result, err := e.report.Calc(g.Expression)
	if err != nil {
		return
	}
	switch v := result.(type) {
	case float64:
		g.SetValue(v)
	case float32:
		g.SetValue(float64(v))
	case int:
		g.SetValue(float64(v))
	case int64:
		g.SetValue(float64(v))
	}
}

// renderGaugeBlob encodes img as PNG, stores it in the BlobStore, and returns the index.
// Returns -1 if the prepared pages or image is nil.
func (e *ReportEngine) renderGaugeBlob(name string, img image.Image) int {
	if img == nil || e.preparedPages == nil {
		return -1
	}
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		return -1
	}
	return e.preparedPages.BlobStore.Add(name, buf.Bytes())
}

// evalText evaluates a text template with [bracket] expressions.
// Returns the raw text on error.
func (e *ReportEngine) evalText(text string) string {
	return e.evalTextWithFormat(text, nil)
}

// evalTextWithFormat evaluates a text template, applying the format to each
// bracket expression's typed value. Mirrors C# TextObject.GetData() which calls
// CalcAndFormatExpression(expression, expressionIndex) for each [bracket] in text.
// Each expression is evaluated via Report.Calc() to get a typed value (e.g.
// time.Time for dates), then Format.FormatValue() is applied to that value.
func (e *ReportEngine) evalTextWithFormat(text string, f format.Format) string {
	if e.report == nil || text == "" {
		return text
	}

	// If no format, fall back to plain CalcText (string substitution only).
	if f == nil {
		result, err := e.report.CalcText(text)
		if err != nil {
			return text
		}
		return result
	}

	// Fast path: when the text is a single bracket expression (including
	// compound forms like [[Field1] * [Field2]]), evaluate the raw value
	// and apply the format directly. This handles double-bracket compound
	// expressions that expr.Parse would split (since [[ is an escape sequence).
	trimmed := strings.TrimSpace(text)
	if isSingleBracketExpr(trimmed) {
		if val, err := e.report.Calc(trimmed); err == nil {
			return f.FormatValue(val)
		}
	}

	// Per-expression formatting: parse text into literal and expression tokens,
	// evaluate each expression with Report.Calc() to get a typed value, and
	// apply the format. Mirrors C# TextObject.GetData() loop (TextObject.cs:1650-1669).
	tokens := expr.Parse(text)
	if tokens == nil {
		return text
	}

	var sb strings.Builder
	for _, tok := range tokens {
		if !tok.IsExpr {
			sb.WriteString(tok.Value)
			continue
		}
		// Evaluate expression to get a typed value (e.g. time.Time, float64).
		val, err := e.report.Calc("[" + tok.Value + "]")
		if err != nil {
			sb.WriteString("[")
			sb.WriteString(tok.Value)
			sb.WriteString("]")
			continue
		}
		// Apply the format to the typed value.
		sb.WriteString(f.FormatValue(val))
	}
	return sb.String()
}

// isSingleBracketExpr reports whether s is enclosed in a single balanced
// outermost [...] pair (possibly with nested brackets inside), for example:
//
//	"[FieldName]"                                    → true
//	"[[Field1] * [Field2]]"                          → true
//	"[Field1] + [Field2]"                            → false (two separate expressions)
//	"Hello [Name]!"                                  → false (literal prefix)
func isSingleBracketExpr(s string) bool {
	if len(s) < 2 || s[0] != '[' || s[len(s)-1] != ']' {
		return false
	}
	depth := 0
	for i, ch := range s {
		if ch == '[' {
			depth++
		} else if ch == ']' {
			depth--
			if depth == 0 && i < len(s)-1 {
				return false
			}
		}
	}
	return true
}

// renderBarcode renders a BarcodeBase to an image.Image using the Render method
// exposed by BaseBarcodeImpl embedders.
func renderBarcode(bc barcode.BarcodeBase, width, height int) (image.Image, error) {
	type renderer interface {
		Render(width, height int) (image.Image, error)
	}
	if r, ok := bc.(renderer); ok {
		return r.Render(width, height)
	}
	return nil, fmt.Errorf("barcode type %T does not implement Render", bc)
}

// decodeSvgData decodes an SVG source stored as either:
//   - raw UTF-8 SVG XML (starts with '<')
//   - base64-encoded SVG XML
//
// Returns the raw SVG bytes, or nil on failure.
func decodeSvgData(s string) []byte {
	s = strings.TrimSpace(s)
	if strings.HasPrefix(s, "<") {
		return []byte(s)
	}
	// Try standard and URL-safe base64.
	if b, err := base64.StdEncoding.DecodeString(s); err == nil {
		return b
	}
	if b, err := base64.URLEncoding.DecodeString(s); err == nil {
		return b
	}
	return nil
}

// anyToBool converts any value to bool following C# Variant.CBln() coercion rules:
//   - nil / "": false
//   - bool: direct
//   - numeric (int*, uint*, float*): non-zero → true
//   - string: normalised to lowercase, then compared against known truthy/falsy sets;
//     unknown non-empty strings default to true (matches C# final "return true")
//   - other non-nil: true (matches C# CBln final "return true")
//
// Mirrors FastReport.Utils.Variant.ToBoolean() / CBln() (Variant.cs lines 338-1784).
func anyToBool(v any) bool {
	if v == nil {
		return false
	}
	switch val := v.(type) {
	case bool:
		return val
	case int:
		return val != 0
	case int8:
		return val != 0
	case int16:
		return val != 0
	case int32:
		return val != 0
	case int64:
		return val != 0
	case uint:
		return val != 0
	case uint8:
		return val != 0
	case uint16:
		return val != 0
	case uint32:
		return val != 0
	case uint64:
		return val != 0
	case float32:
		return val != 0
	case float64:
		return val != 0
	case string:
		s := strings.ToLower(strings.TrimSpace(val))
		if s == "" || s == "false" || s == "f" || s == "0" || s == "0.0" ||
			s == "no" || s == "n" || s == "off" || s == "negative" || s == "neg" ||
			s == "disabled" || s == "incorrect" || s == "wrong" || s == "left" {
			return false
		}
		// Explicitly truthy values.
		if s == "true" || s == "t" || s == "1" || s == "-1" ||
			s == "yes" || s == "y" || s == "on" || s == "positive" || s == "pos" ||
			s == "enabled" || s == "correct" || s == "right" {
			return true
		}
		// Unknown non-empty string → true (mirrors C# CBln's default return true).
		return true
	}
	// Unknown non-nil type → true (mirrors C# CBln final "return true").
	return true
}

// extractBarcodeModules converts a barcode image to a boolean module matrix.
// Each true entry represents a dark (black) module in the barcode symbol.
// The image is sampled at its native resolution — 1 pixel per module.
func extractBarcodeModules(img image.Image) [][]bool {
	if img == nil {
		return nil
	}
	bounds := img.Bounds()
	w := bounds.Max.X - bounds.Min.X
	h := bounds.Max.Y - bounds.Min.Y
	if w <= 0 || h <= 0 {
		return nil
	}
	modules := make([][]bool, h)
	for y := 0; y < h; y++ {
		modules[y] = make([]bool, w)
		for x := 0; x < w; x++ {
			c := img.At(bounds.Min.X+x, bounds.Min.Y+y)
			r, g, b, _ := c.RGBA()
			// Treat pixels with luminance < 50% as dark modules.
			lum := (r + g + b) / 3
			modules[y][x] = lum < 0x7FFF
		}
	}
	return modules
}
