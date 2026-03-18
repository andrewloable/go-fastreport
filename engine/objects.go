package engine

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"math"
	"strings"

	"github.com/andrewloable/go-fastreport/band"
	"github.com/andrewloable/go-fastreport/barcode"
	"github.com/andrewloable/go-fastreport/format"
	"github.com/andrewloable/go-fastreport/gauge"
	"github.com/andrewloable/go-fastreport/maprender"
	"github.com/andrewloable/go-fastreport/object"
	"github.com/andrewloable/go-fastreport/preview"
	"github.com/andrewloable/go-fastreport/report"
	"github.com/andrewloable/go-fastreport/sparkline"
	"github.com/andrewloable/go-fastreport/style"
	"github.com/andrewloable/go-fastreport/table"
)


// populateBandObjects converts the child report objects of a BandBase into
// preview.PreparedObject snapshots and appends them to pb.
// It evaluates [bracket] expressions in TextObject text via Report.Calc().
func (e *ReportEngine) populateBandObjects(bb *band.BandBase, pb *preview.PreparedBand) {
	if bb == nil {
		return
	}
	// Apply dock layout using the band's own width/height as the container.
	applyDockLayout(bb.Objects(), bb.Width(), bb.Height())
	e.populateBandObjects2(bb.Objects(), pb)
}

// populateBandObjects2 converts objects from any ObjectCollection into PreparedObjects.
func (e *ReportEngine) populateBandObjects2(objs *report.ObjectCollection, pb *preview.PreparedBand) {
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
				e.populateTableObjects(base, tbl.Left(), tbl.Top(), pb)
			}
			// Render AdvMatrixObject physical cells as individual PreparedObjects.
			if adv, ok := obj.(*object.AdvMatrixObject); ok {
				e.populateAdvMatrixCells(adv, pb)
			}
			// Render CellularTextObject as a character-grid of PreparedObjects.
			if cellular, ok := obj.(*object.CellularTextObject); ok {
				text := e.evalText(cellular.Text())
				e.populateCellularTextCells(cellular, text, pb)
			}
			// Register deferred text evaluation if ProcessAt != Default.
			if txt, ok := obj.(*object.TextObject); ok && txt.ProcessAt() != object.ProcessAtDefault {
				state := processAtToEngineState(txt.ProcessAt())
				capturedPb := pb
				capturedIdx := idx
				capturedTxt := txt
				capturedFmt := txt.Format()
				pb.Objects[idx].Text = "" // placeholder: blank until deferred state fires
				// PageFinished and ColumnFinished handlers must re-fire on every
				// page/column boundary. ReportFinished and other states fire once.
				switch txt.ProcessAt() {
				case object.ProcessAtPageFinished, object.ProcessAtColumnFinished:
					e.AddRepeatingDeferredHandler(state, func() {
						capturedPb.Objects[capturedIdx].Text = e.evalTextWithFormat(capturedTxt.Text(), capturedFmt)
					})
				default:
					e.AddDeferredHandler(state, func() {
						capturedPb.Objects[capturedIdx].Text = e.evalTextWithFormat(capturedTxt.Text(), capturedFmt)
					})
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
func (e *ReportEngine) populateTableObjects(tbl *table.TableBase, originX, originY float32, pb *preview.PreparedBand) {
	// Pre-compute cumulative column X offsets.
	cols := tbl.Columns()
	colX := make([]float32, len(cols)+1)
	for i, col := range cols {
		colX[i+1] = colX[i] + col.Width()
	}

	// Iterate rows, accumulating Y offset.
	rowY := float32(0)
	for ri, row := range tbl.Rows() {
		rowH := row.Height()
		for ci, cell := range row.Cells() {
			if cell == nil {
				continue
			}
			// Compute cell width from ColSpan.
			colSpan := cell.ColSpan()
			if colSpan < 1 {
				colSpan = 1
			}
			endCol := ci + colSpan
			if endCol > len(cols) {
				endCol = len(cols)
			}
			cellW := colX[endCol] - colX[ci]

			// Compute cell height from RowSpan.
			rowSpan := cell.RowSpan()
			if rowSpan < 1 {
				rowSpan = 1
			}
			cellH := float32(0)
			for si := ri; si < ri+rowSpan && si < len(tbl.Rows()); si++ {
				cellH += tbl.Rows()[si].Height()
			}

			absLeft := originX + colX[ci]
			absTop := originY + rowY

			// Build the cell PreparedObject (cell embeds TextObject).
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
				Border:    cell.Border(),
			}
			pb.Objects = append(pb.Objects, po)
		}
		rowY += rowH
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
	for !ds.EOF() {
		e.report.SetCalcContext(ds)
		eval := func(text string) string { return e.evalText(text) }
		h.CellTextEval = eval
		h.PrintColumn(dataColIdx)
		h.PrintRows()
		h.CellTextEval = nil
		_ = ds.Next()
	}

	// Print trailer columns (dataColIdx+1 .. nCols-1).
	for i := dataColIdx + 1; i < nCols; i++ {
		h.PrintColumn(i)
		h.PrintRows()
	}

	return h.Result()
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
				WordWrap:  false,
				Border:    border,
			}
			pb.Objects = append(pb.Objects, po)
		}
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
	// Skip invisible and band types.
	type hasVisible interface{ Visible() bool }
	if v, ok := obj.(hasVisible); ok && !v.Visible() {
		return nil
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

	po := &preview.PreparedObject{
		Name:    obj.Name(),
		Left:    geom.Left(),
		Top:     geom.Top(),
		Width:   geom.Width(),
		Height:  geom.Height(),
		BlobIdx: -1,
		Font:    style.DefaultFont(),
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
			case "Bookmark":
				po.HyperlinkKind = 3
				po.HyperlinkValue = hl.Expression
			}
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
		po.ParagraphOffset = v.ParagraphOffset()
		po.LineHeight = v.LineHeight()
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
		if f, ok := v.Fill().(*style.SolidFill); ok {
			po.FillColor = f.Color
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
		if data := v.ImageData(); len(data) > 0 {
			po.BlobIdx = e.preparedPages.BlobStore.Add(v.Name(), data)
		}

	case *object.CellularTextObject:
		// CellularTextObject is rendered as a grid of individual character cells.
		// The anchor PreparedObject is a transparent shape (bounding box) that
		// carries the border and fill; individual character cells are emitted by
		// populateCellularTextCells, called from populateBandObjects2 below.
		po.Kind = preview.ObjectTypeShape
		po.ShapeKind = 0 // Rectangle
		po.Border = v.Border()
		if f, ok := v.Fill().(*style.SolidFill); ok {
			po.FillColor = f.Color
		}

	case *object.ZipCodeObject:
		// Render as text showing the evaluated zip code value.
		po.Kind = preview.ObjectTypeText
		po.TextColor = color.RGBA{A: 255}
		text := v.Expression()
		if text == "" {
			text = v.Text()
		}
		po.Text = e.evalText(text)

	case *object.CheckBoxObject:
		po.Kind = preview.ObjectTypeCheckBox
		// Evaluate expression or data column binding to determine checked state.
		// If no expression/binding, fall back to the statically deserialized value.
		if expr := v.Expression(); expr != "" {
			result, err := e.report.Calc(expr)
			if err == nil {
				switch bv := result.(type) {
				case bool:
					v.SetChecked(bv)
				case string:
					v.SetChecked(bv == "true" || bv == "True" || bv == "1")
				}
			}
		} else if col := v.DataColumn(); col != "" {
			result := e.evalText("[" + col + "]")
			v.SetChecked(result == "true" || result == "True" || result == "1")
		}
		po.Checked = v.Checked()
		po.CheckedSymbol = int(v.CheckedSymbol())
		po.UncheckedSymbol = int(v.UncheckedSymbol())
		po.CheckColor = v.CheckColor()

	case *gauge.LinearGauge:
		e.evalGaugeValue(&v.GaugeObject)
		po.Kind = preview.ObjectTypePicture
		po.BlobIdx = e.renderGaugeBlob(v.Name(), gauge.RenderLinear(v, int(geom.Width()), int(geom.Height())))

	case *gauge.RadialGauge:
		e.evalGaugeValue(&v.GaugeObject)
		po.Kind = preview.ObjectTypePicture
		po.BlobIdx = e.renderGaugeBlob(v.Name(), gauge.RenderRadial(v, int(geom.Width()), int(geom.Height())))

	case *gauge.SimpleGauge:
		e.evalGaugeValue(&v.GaugeObject)
		po.Kind = preview.ObjectTypePicture
		po.BlobIdx = e.renderGaugeBlob(v.Name(), gauge.RenderSimple(v, int(geom.Width()), int(geom.Height())))

	case *gauge.SimpleProgressGauge:
		e.evalGaugeValue(&v.GaugeObject)
		po.Kind = preview.ObjectTypePicture
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
		// Evaluate the barcode text from expression or static text.
		text := v.Expression()
		if text == "" {
			text = v.Text()
		}
		text = e.evalText(text)
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
				w := int(geom.Width())
				h := int(geom.Height())
				if w <= 0 {
					w = 200
				}
				if h <= 0 {
					h = 60
				}
				img, err := renderBarcode(v.Barcode, w, h)
				if err == nil && img != nil && e.preparedPages != nil {
					var buf bytes.Buffer
					if encErr := png.Encode(&buf, img); encErr == nil {
						po.BlobIdx = e.preparedPages.BlobStore.Add(v.Name(), buf.Bytes())
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

	case *table.TableObject:
		// The bounding box is rendered as a shape; individual cells are added
		// by populateTableObjects called from populateBandObjects2.
		po.Kind = preview.ObjectTypeShape
		po.ShapeKind = 0 // Rectangle
		po.Border = v.Border()
		if f, ok := v.Fill().(*style.SolidFill); ok {
			po.FillColor = f.Color
		}

	default:
		// Not a known renderable type (could be a nested band etc.)
		return nil
	}

	// Register bookmark if the component has one.
	type hasBookmark interface{ Bookmark() string }
	if bk, ok := obj.(hasBookmark); ok {
		if name := bk.Bookmark(); name != "" {
			e.AddBookmark(e.evalText(name))
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

// evalTextWithFormat evaluates a text template and, if f is non-nil and the
// text is a single bracket expression, applies the format to the raw value.
func (e *ReportEngine) evalTextWithFormat(text string, f format.Format) string {
	if e.report == nil || text == "" {
		return text
	}
	// When a format is set and the text is exactly one bracket expression,
	// evaluate the raw value and apply the format before converting to string.
	if f != nil {
		trimmed := strings.TrimSpace(text)
		if strings.HasPrefix(trimmed, "[") && strings.HasSuffix(trimmed, "]") && strings.Count(trimmed, "[") == 1 {
			expr := trimmed[1 : len(trimmed)-1]
			if val, err := e.report.Calc(expr); err == nil {
				return f.FormatValue(val)
			}
		}
	}
	result, err := e.report.CalcText(text)
	if err != nil {
		return text
	}
	return result
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
