// Package xlsx implements an Excel XLSX export filter for go-fastreport.
// It renders prepared pages as an Excel workbook, mapping each band row
// of text objects to a spreadsheet row, with basic font and fill styling.
package xlsx

import (
	"bytes"
	"fmt"
	"io"
	"math"
	"sort"

	excelize "github.com/xuri/excelize/v2"

	"github.com/andrewloable/go-fastreport/export"
	"github.com/andrewloable/go-fastreport/preview"
	"github.com/andrewloable/go-fastreport/style"
)

// Exporter produces XLSX output from a PreparedPages collection.
//
// Strategy:
//   - All pages are written to a single worksheet (SheetName).
//   - Bands are visited in order; text objects within each band are grouped by
//     Y coordinate (same approach as the CSV exporter) to identify logical rows.
//   - Within each row, objects are sorted by X coordinate to assign columns.
//   - Cell values, font styling, fill colour, borders and column widths are
//     applied via the excelize library.
//   - Picture objects (images) are embedded using AddPictureFromBytes.
type Exporter struct {
	export.ExportBase

	// SheetName is the name of the worksheet to create (default "Report").
	SheetName string

	w      io.Writer
	pp     *preview.PreparedPages
	f      *excelize.File
	sheet  string
	rowIdx int // 1-based current row in the sheet

	// colWidths tracks the maximum estimated character width per column index.
	colWidths map[int]float64

	// styleCache maps a styleKey to an excelize style ID to avoid re-creating
	// identical styles on every cell.
	styleCache map[styleKey]int

	// imageRow accumulates the 1-based row index at which the next image
	// should be anchored (set when processing picture objects in a band).
	// We record the band's start row and use it for image anchoring.
	currentBandRow int
}

// styleKey uniquely identifies a cell style so we can cache excelize style IDs.
type styleKey struct {
	bold, italic, underline bool
	fontSize                float32
	fontName                string
	fontColorARGB           string
	fillColorARGB           string
	horzAlign               int
	vertAlign               int
	borderTop, borderRight, borderBottom, borderLeft bool
}

// NewExporter creates an Exporter with sensible defaults.
func NewExporter() *Exporter {
	return &Exporter{
		ExportBase: export.NewExportBase(),
		SheetName:  "Report",
	}
}

// Export writes the PreparedPages as an Excel workbook to w.
func (e *Exporter) Export(pages *preview.PreparedPages, w io.Writer) error {
	e.w = w
	e.pp = pages
	return e.ExportBase.Export(pages, w, e)
}

// FileExtension returns the recommended file extension.
func (e *Exporter) FileExtension() string { return ".xlsx" }

// Name returns the human-readable name of this export format.
func (e *Exporter) Name() string { return "XLSX" }

// ── Exporter interface ─────────────────────────────────────────────────────────

// Start initialises the excelize file and state.
func (e *Exporter) Start() error {
	e.f = excelize.NewFile()
	e.sheet = e.SheetName
	if e.sheet == "" {
		e.sheet = "Report"
	}

	// Rename the default "Sheet1" to our desired name.
	defaultSheet := e.f.GetSheetName(0)
	if defaultSheet != e.sheet {
		if err := e.f.SetSheetName(defaultSheet, e.sheet); err != nil {
			return fmt.Errorf("xlsx: set sheet name: %w", err)
		}
	}

	e.rowIdx = 1
	e.colWidths = make(map[int]float64)
	e.styleCache = make(map[styleKey]int)
	return nil
}

// ExportPageBegin is a no-op for XLSX (pages are transparent to the format;
// all content goes into a single worksheet).
func (e *Exporter) ExportPageBegin(_ *preview.PreparedPage) error { return nil }

// ExportBand converts a single band into one or more spreadsheet rows.
//
// Strategy:
//  1. Collect text objects from the band.
//  2. Group them by Y coordinate (epsilon 1 px) — each group becomes a row.
//  3. Within each row sort objects by X (left-to-right).
//  4. Map each object to a column using its sort index.
//  5. Set cell value and apply styling.
//  6. Process picture objects as anchored images.
func (e *Exporter) ExportBand(b *preview.PreparedBand) error {
	// Remember where this band starts so image anchors are correct.
	e.currentBandRow = e.rowIdx

	// ── Text objects ──────────────────────────────────────────────────────
	textObjs := collectTextObjects(b)
	if len(textObjs) > 0 {
		rows := groupByY(textObjs)
		for _, row := range rows {
			sort.Slice(row, func(i, j int) bool {
				return row[i].Left < row[j].Left
			})
			for colIdx, obj := range row {
				col := colIdx + 1 // 1-based
				cellName, err := excelize.CoordinatesToCellName(col, e.rowIdx)
				if err != nil {
					return fmt.Errorf("xlsx: cell name (%d,%d): %w", col, e.rowIdx, err)
				}

				text := objectText(obj)
				if err := e.f.SetCellValue(e.sheet, cellName, text); err != nil {
					return fmt.Errorf("xlsx: set cell value %s: %w", cellName, err)
				}

				// Apply cell style.
				styleID, err := e.cellStyle(obj)
				if err == nil && styleID != 0 {
					_ = e.f.SetCellStyle(e.sheet, cellName, cellName, styleID)
				}

				// Track column width estimate (1 char ≈ 7 px; Excel column width ≈ char count).
				charWidth := estimateCharWidth(text, obj.Width)
				if charWidth > e.colWidths[col] {
					e.colWidths[col] = charWidth
				}
			}
			e.rowIdx++
		}
	}

	// ── Picture objects ───────────────────────────────────────────────────
	for _, obj := range b.Objects {
		if obj.Kind != preview.ObjectTypePicture {
			continue
		}
		if obj.BlobIdx < 0 || e.pp == nil {
			continue
		}
		imgData := e.pp.BlobStore.Get(obj.BlobIdx)
		if len(imgData) == 0 {
			continue
		}

		// Determine the column and row for the anchor cell.
		// We use a rough mapping: Left px → column, currentBandRow for row.
		col := int(obj.Left/60) + 1
		if col < 1 {
			col = 1
		}
		cellName, err := excelize.CoordinatesToCellName(col, e.currentBandRow)
		if err != nil {
			continue
		}

		ext := imageExtension(imgData)
		picOpts := &excelize.GraphicOptions{
			ScaleX: 1.0,
			ScaleY: 1.0,
		}
		if addErr := e.f.AddPictureFromBytes(e.sheet, cellName, &excelize.Picture{
			Extension: ext,
			File:      imgData,
			Format:    picOpts,
		}); addErr != nil {
			// Non-fatal: skip images that fail to embed.
			continue
		}
	}

	return nil
}

// ExportPageEnd is a no-op for XLSX.
func (e *Exporter) ExportPageEnd(_ *preview.PreparedPage) error { return nil }

// Finish applies column widths, then writes the workbook to the writer.
func (e *Exporter) Finish() error {
	// Apply estimated column widths (capped at 60 chars wide).
	for col, w := range e.colWidths {
		colName, err := excelize.ColumnNumberToName(col)
		if err != nil {
			continue
		}
		width := w
		if width < 8 {
			width = 8
		}
		if width > 60 {
			width = 60
		}
		_ = e.f.SetColWidth(e.sheet, colName, colName, width)
	}

	// Write to the provided writer.
	var buf bytes.Buffer
	if err := e.f.Write(&buf); err != nil {
		return fmt.Errorf("xlsx: write workbook: %w", err)
	}
	_, err := io.Copy(e.w, &buf)
	return err
}

// ── style helpers ─────────────────────────────────────────────────────────────

// cellStyle returns (or creates and caches) an excelize style ID for the object.
func (e *Exporter) cellStyle(obj preview.PreparedObject) (int, error) {
	fc := obj.FillColor
	tc := obj.TextColor
	font := obj.Font

	key := styleKey{
		bold:          font.Style&style.FontStyleBold != 0,
		italic:        font.Style&style.FontStyleItalic != 0,
		underline:     font.Style&style.FontStyleUnderline != 0,
		fontSize:      font.Size,
		fontName:      font.Name,
		fontColorARGB: rgbaToARGB(tc.A, tc.R, tc.G, tc.B),
		fillColorARGB: rgbaToARGB(fc.A, fc.R, fc.G, fc.B),
		horzAlign:     obj.HorzAlign,
		vertAlign:     obj.VertAlign,
		borderTop:     obj.Border.VisibleLines&style.BorderLinesTop != 0,
		borderRight:   obj.Border.VisibleLines&style.BorderLinesRight != 0,
		borderBottom:  obj.Border.VisibleLines&style.BorderLinesBottom != 0,
		borderLeft:    obj.Border.VisibleLines&style.BorderLinesLeft != 0,
	}

	if id, ok := e.styleCache[key]; ok {
		return id, nil
	}

	s := &excelize.Style{}

	// Font.
	xFont := &excelize.Font{
		Bold:   key.bold,
		Italic: key.italic,
		Size:   float64(font.Size),
		Family: font.Name,
		Color:  key.fontColorARGB,
	}
	if key.underline {
		xFont.Underline = "single"
	}
	s.Font = xFont

	// Fill (solid fill when alpha > 0).
	if fc.A > 0 {
		s.Fill = excelize.Fill{
			Type:    "pattern",
			Pattern: 1,
			Color:   []string{key.fillColorARGB},
		}
	}

	// Alignment.
	align := &excelize.Alignment{WrapText: obj.WordWrap}
	switch obj.HorzAlign {
	case 1:
		align.Horizontal = "center"
	case 2:
		align.Horizontal = "right"
	case 3:
		align.Horizontal = "justify"
	default:
		align.Horizontal = "left"
	}
	switch obj.VertAlign {
	case 1:
		align.Vertical = "center"
	case 2:
		align.Vertical = "bottom"
	default:
		align.Vertical = "top"
	}
	s.Alignment = align

	// Borders.
	if key.borderTop || key.borderRight || key.borderBottom || key.borderLeft {
		var borders []excelize.Border
		if key.borderTop {
			borders = append(borders, excelize.Border{Type: "top", Color: "000000", Style: 1})
		}
		if key.borderRight {
			borders = append(borders, excelize.Border{Type: "right", Color: "000000", Style: 1})
		}
		if key.borderBottom {
			borders = append(borders, excelize.Border{Type: "bottom", Color: "000000", Style: 1})
		}
		if key.borderLeft {
			borders = append(borders, excelize.Border{Type: "left", Color: "000000", Style: 1})
		}
		s.Border = borders
	}

	id, err := e.f.NewStyle(s)
	if err != nil {
		return 0, err
	}
	e.styleCache[key] = id
	return id, nil
}

// rgbaToARGB converts RGBA components to an ARGB hex string (no leading #)
// suitable for excelize colour values.
func rgbaToARGB(a, r, g, b uint8) string {
	return fmt.Sprintf("%02X%02X%02X%02X", a, r, g, b)
}

// estimateCharWidth returns an estimated Excel column width (in character units)
// for a cell given its text content and pixel width.
// Excel column width ≈ number of characters in the default font.
// Heuristic: 1 Excel char unit ≈ 7 px.
func estimateCharWidth(text string, widthPx float32) float64 {
	byLen := float64(len([]rune(text))) + 2 // +2 for padding
	byPx := float64(widthPx) / 7.0
	if byLen > byPx {
		return byLen
	}
	return byPx
}

// imageExtension detects the file extension from image magic bytes.
func imageExtension(data []byte) string {
	if len(data) >= 2 {
		switch {
		case data[0] == 0xFF && data[1] == 0xD8:
			return ".jpg"
		case len(data) >= 8 && data[0] == 0x89 && data[1] == 0x50:
			return ".png"
		case len(data) >= 3 && data[0] == 0x47 && data[1] == 0x49 && data[2] == 0x46:
			return ".gif"
		case len(data) >= 2 && data[0] == 0x42 && data[1] == 0x4D:
			return ".bmp"
		}
	}
	return ".png" // default
}

// ── text-object helpers (mirrors csv exporter logic) ─────────────────────────

// collectTextObjects returns all text-bearing objects from a band.
func collectTextObjects(b *preview.PreparedBand) []preview.PreparedObject {
	var out []preview.PreparedObject
	for _, obj := range b.Objects {
		switch obj.Kind {
		case preview.ObjectTypeText,
			preview.ObjectTypeHtml,
			preview.ObjectTypeRTF,
			preview.ObjectTypeCheckBox:
			out = append(out, obj)
		}
	}
	return out
}

// objectText extracts the display text from a PreparedObject.
func objectText(obj preview.PreparedObject) string {
	if obj.Kind == preview.ObjectTypeCheckBox {
		if obj.Checked || obj.Text == "true" {
			return "true"
		}
		return "false"
	}
	return obj.Text
}

// groupByY groups objects by Y coordinate with a 1 px epsilon tolerance.
// Returns rows in ascending Y order.
func groupByY(objs []preview.PreparedObject) [][]preview.PreparedObject {
	const eps = 1.0

	type yGroup struct {
		y    float32
		objs []preview.PreparedObject
	}

	var groups []yGroup
	for _, obj := range objs {
		placed := false
		for i := range groups {
			if math.Abs(float64(obj.Top-groups[i].y)) < eps {
				groups[i].objs = append(groups[i].objs, obj)
				placed = true
				break
			}
		}
		if !placed {
			groups = append(groups, yGroup{y: obj.Top, objs: []preview.PreparedObject{obj}})
		}
	}

	sort.Slice(groups, func(i, j int) bool {
		return groups[i].y < groups[j].y
	})

	rows := make([][]preview.PreparedObject, len(groups))
	for i, g := range groups {
		rows[i] = g.objs
	}
	return rows
}
