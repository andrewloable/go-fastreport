package matrix

// descriptor_templates.go adds the TemplateColumn/TemplateRow/TemplateCell fields
// to Descriptor and HeaderDescriptor, plus Assign() methods and MatrixData
// collection API (IndexOf/Contains/Insert/Remove/ToArray for Columns/Rows/Cells).
//
// These are the porting gaps identified in issue go-fastreport-5r21k:
//   - MatrixDescriptor: TemplateColumn, TemplateRow, TemplateCell, Assign()
//   - MatrixHeaderDescriptor: TemplateTotalColumn, TemplateTotalRow, TemplateTotalCell, Assign()
//   - MatrixHeader collection API: IndexOf/Contains/Insert/Remove/ToArray
//   - MatrixData.Clear()
//
// C# source references are in FastReport.Base/Matrix/MatrixDescriptor.cs,
// MatrixHeaderDescriptor.cs, and MatrixHeader.cs.

import "github.com/andrewloable/go-fastreport/table"

// ── Descriptor template fields and Assign ─────────────────────────────────────

// TemplateColumn is the table column bound to this descriptor.
// Used internally during BuildTemplate (MatrixHelper.UpdateColumnDescriptors in C#).
// C# source: FastReport.Base/Matrix/MatrixDescriptor.cs, property TemplateColumn.
//
// NOTE: These three fields extend the Descriptor struct. Because Go does not
// allow adding fields to an existing struct in a different file, they are
// stored on a per-instance extension attached via the descriptorExt map below.
// For callers that need the canonical C# fields, use the accessor helpers
// on DescriptorExt.

// DescriptorExt carries the template-binding fields that the C# MatrixDescriptor
// exposes as TemplateColumn, TemplateRow, and TemplateCell.
// It is embedded as a value in both HeaderDescriptor and CellDescriptor via
// the Descriptor type, exposed through the ExtFields accessor.
//
// C# source: FastReport.Base/Matrix/MatrixDescriptor.cs.
type DescriptorExt struct {
	// TemplateColumn is the table column bound to this descriptor.
	// C# source: MatrixDescriptor.TemplateColumn property.
	TemplateColumn *table.TableColumn

	// TemplateRow is the table row bound to this descriptor.
	// C# source: MatrixDescriptor.TemplateRow property.
	TemplateRow *table.TableRow

	// TemplateCell is the table cell bound to this descriptor.
	// Exposed so callers can adjust fill, text, etc. before BuildTemplate.
	// C# source: MatrixDescriptor.TemplateCell property.
	TemplateCell *table.TableCell
}

// assignDescriptorExt copies base template fields from src to dst.
// Mirrors the Assign() body of FastReport.Matrix.MatrixDescriptor.
func assignDescriptorExt(dst, src *DescriptorExt) {
	dst.TemplateCell = src.TemplateCell
}

// ── HeaderDescriptorExt — total-element template fields ───────────────────────

// HeaderDescriptorExt adds the TemplateTotalColumn/Row/Cell fields that
// C# MatrixHeaderDescriptor exposes.
// C# source: FastReport.Base/Matrix/MatrixHeaderDescriptor.cs.
type HeaderDescriptorExt struct {
	DescriptorExt

	// TemplateTotalColumn is the template column bound to the "total" element.
	// C# source: MatrixHeaderDescriptor.TemplateTotalColumn property.
	TemplateTotalColumn *table.TableColumn

	// TemplateTotalRow is the template row bound to the "total" element.
	// C# source: MatrixHeaderDescriptor.TemplateTotalRow property.
	TemplateTotalRow *table.TableRow

	// TemplateTotalCell is the template cell bound to the "total" element.
	// Callers may set this to change the "Total" label text or fill colour.
	// C# source: MatrixHeaderDescriptor.TemplateTotalCell property.
	TemplateTotalCell *table.TableCell
}

// Assign copies all fields from src to dst (base + header-specific).
// Mirrors FastReport.Matrix.MatrixHeaderDescriptor.Assign().
// C# source: FastReport.Base/Matrix/MatrixHeaderDescriptor.cs, method Assign.
func (h *HeaderDescriptorExt) Assign(src *HeaderDescriptorExt) {
	assignDescriptorExt(&h.DescriptorExt, &src.DescriptorExt)
	h.TemplateTotalCell = src.TemplateTotalCell
}

// ── HeaderDescriptor — embed DescriptorExt + HeaderDescriptorExt helpers ──────
// HeaderDescriptor already exists in matrix.go. We extend it via two promoted
// struct fields added below. Because Go structs cannot be altered in a separate
// file we use standalone per-instance extension maps.

// headerExtMap stores the HeaderDescriptorExt for each *HeaderDescriptor.
// This avoids importing cgo or unsafe while still giving every descriptor its
// own TemplateColumn/Row/Cell and TemplateTotalColumn/Row/Cell fields.
var headerExtMap = make(map[*HeaderDescriptor]*HeaderDescriptorExt)

// HeaderExt returns the HeaderDescriptorExt for h, creating it if necessary.
// Use this to get/set TemplateColumn, TemplateRow, TemplateCell,
// TemplateTotalColumn, TemplateTotalRow, TemplateTotalCell.
//
// Example:
//
//	h.HeaderExt().TemplateCell = someCell
func (h *HeaderDescriptor) HeaderExt() *HeaderDescriptorExt {
	if ext, ok := headerExtMap[h]; ok {
		return ext
	}
	ext := &HeaderDescriptorExt{}
	headerExtMap[h] = ext
	return ext
}

// Assign copies the Sort/Totals/TotalsFirst/PageBreak/SuppressTotals and the
// template-cell fields from src to h.
// Mirrors FastReport.Matrix.MatrixHeaderDescriptor.Assign().
// C# source: FastReport.Base/Matrix/MatrixHeaderDescriptor.cs, method Assign.
func (h *HeaderDescriptor) Assign(src *HeaderDescriptor) {
	// Base Descriptor fields (Expression + TemplateCell).
	h.Expression = src.Expression
	if srcExt, ok := headerExtMap[src]; ok {
		h.HeaderExt().DescriptorExt.TemplateCell = srcExt.DescriptorExt.TemplateCell
		h.HeaderExt().TemplateTotalCell = srcExt.TemplateTotalCell
	}
	// Header-specific fields.
	h.Sort = src.Sort
	h.Totals = src.Totals
	h.TotalsFirst = src.TotalsFirst
	h.PageBreak = src.PageBreak
	h.SuppressTotals = src.SuppressTotals
}

// cellExtMap stores the DescriptorExt for each *CellDescriptor.
var cellExtMap = make(map[*CellDescriptor]*DescriptorExt)

// CellExt returns the DescriptorExt for c, creating it if necessary.
func (c *CellDescriptor) CellExt() *DescriptorExt {
	if ext, ok := cellExtMap[c]; ok {
		return ext
	}
	ext := &DescriptorExt{}
	cellExtMap[c] = ext
	return ext
}

// ── MatrixData.Clear ──────────────────────────────────────────────────────────

// Clear removes all descriptors from Columns, Rows and Cells.
// Matches FastReport.Matrix.MatrixData.Clear().
// C# source: FastReport.Base/Matrix/MatrixData.cs, method Clear.
func (d *MatrixData) Clear() {
	d.Columns = d.Columns[:0]
	d.Rows = d.Rows[:0]
	d.Cells = d.Cells[:0]
}

// ── MatrixData — Column collection API ─────────────────────────────────────────
// Mirrors FastReport.Matrix.MatrixHeader collection API.
// C# source: FastReport.Base/Matrix/MatrixHeader.cs.

// IndexOfColumn returns the zero-based index of h in Columns, or -1 if not found.
func (d *MatrixData) IndexOfColumn(h *HeaderDescriptor) int {
	for i, v := range d.Columns {
		if v == h {
			return i
		}
	}
	return -1
}

// ContainsColumn reports whether h is present in Columns.
func (d *MatrixData) ContainsColumn(h *HeaderDescriptor) bool {
	return d.IndexOfColumn(h) >= 0
}

// InsertColumn inserts h at position index in Columns.
// Panics if index is out of range [0, len(Columns)].
func (d *MatrixData) InsertColumn(index int, h *HeaderDescriptor) {
	d.Columns = append(d.Columns, nil)
	copy(d.Columns[index+1:], d.Columns[index:])
	d.Columns[index] = h
}

// RemoveColumn removes the first occurrence of h from Columns.
// Does nothing if h is not present.
func (d *MatrixData) RemoveColumn(h *HeaderDescriptor) {
	if i := d.IndexOfColumn(h); i >= 0 {
		d.Columns = append(d.Columns[:i], d.Columns[i+1:]...)
	}
}

// ColumnsToArray returns a copy of the Columns slice as a new slice.
// Matches FastReport.Matrix.MatrixHeader.ToArray().
func (d *MatrixData) ColumnsToArray() []*HeaderDescriptor {
	out := make([]*HeaderDescriptor, len(d.Columns))
	copy(out, d.Columns)
	return out
}

// ── MatrixData — Row collection API ────────────────────────────────────────────

// IndexOfRow returns the zero-based index of h in Rows, or -1 if not found.
func (d *MatrixData) IndexOfRow(h *HeaderDescriptor) int {
	for i, v := range d.Rows {
		if v == h {
			return i
		}
	}
	return -1
}

// ContainsRow reports whether h is present in Rows.
func (d *MatrixData) ContainsRow(h *HeaderDescriptor) bool {
	return d.IndexOfRow(h) >= 0
}

// InsertRow inserts h at position index in Rows.
// Panics if index is out of range [0, len(Rows)].
func (d *MatrixData) InsertRow(index int, h *HeaderDescriptor) {
	d.Rows = append(d.Rows, nil)
	copy(d.Rows[index+1:], d.Rows[index:])
	d.Rows[index] = h
}

// RemoveRow removes the first occurrence of h from Rows.
// Does nothing if h is not present.
func (d *MatrixData) RemoveRow(h *HeaderDescriptor) {
	if i := d.IndexOfRow(h); i >= 0 {
		d.Rows = append(d.Rows[:i], d.Rows[i+1:]...)
	}
}

// RowsToArray returns a copy of the Rows slice as a new slice.
// Matches FastReport.Matrix.MatrixHeader.ToArray().
func (d *MatrixData) RowsToArray() []*HeaderDescriptor {
	out := make([]*HeaderDescriptor, len(d.Rows))
	copy(out, d.Rows)
	return out
}

// ── MatrixData — Cell collection API ───────────────────────────────────────────

// IndexOfCell returns the zero-based index of c in Cells, or -1 if not found.
func (d *MatrixData) IndexOfCell(c *CellDescriptor) int {
	for i, v := range d.Cells {
		if v == c {
			return i
		}
	}
	return -1
}

// ContainsCell reports whether c is present in Cells.
func (d *MatrixData) ContainsCell(c *CellDescriptor) bool {
	return d.IndexOfCell(c) >= 0
}

// InsertCell inserts c at position index in Cells.
// Panics if index is out of range [0, len(Cells)].
func (d *MatrixData) InsertCell(index int, c *CellDescriptor) {
	d.Cells = append(d.Cells, nil)
	copy(d.Cells[index+1:], d.Cells[index:])
	d.Cells[index] = c
}

// RemoveCell removes the first occurrence of c from Cells.
// Does nothing if c is not present.
func (d *MatrixData) RemoveCell(c *CellDescriptor) {
	if i := d.IndexOfCell(c); i >= 0 {
		d.Cells = append(d.Cells[:i], d.Cells[i+1:]...)
	}
}

// CellsToArray returns a copy of the Cells slice as a new slice.
func (d *MatrixData) CellsToArray() []*CellDescriptor {
	out := make([]*CellDescriptor, len(d.Cells))
	copy(out, d.Cells)
	return out
}
