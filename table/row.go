package table

import (
	"github.com/andrewloable/go-fastreport/report"
)

// TableRow represents a row in a table object.
// It is the Go equivalent of FastReport.Table.TableRow.
type TableRow struct {
	report.ComponentBase

	// cells holds the TableCell instances for this row (one per column).
	cells []*TableCell

	// minHeight is the minimum row height in pixels.
	minHeight float32
	// maxHeight is the maximum row height in pixels. 0 means unlimited.
	maxHeight float32
	// autoSize makes the row calculate its height automatically.
	autoSize bool
	// canBreak allows the row to be split across pages.
	canBreak bool
	// pageBreak inserts a page break before this row.
	pageBreak bool
	// keepRows specifies how many rows to keep together on a page.
	keepRows int
}

// NewTableRow creates a TableRow with defaults matching the C# constructor.
func NewTableRow() *TableRow {
	r := &TableRow{
		ComponentBase: *report.NewComponentBase(),
		maxHeight:     1000, // matches C# DefaultValue(1000)
	}
	r.SetHeight(30) // default row height
	return r
}

// Cells returns the cells in this row.
func (r *TableRow) Cells() []*TableCell { return r.cells }

// Cell returns the cell at column index i, or nil if out of range.
func (r *TableRow) Cell(i int) *TableCell {
	if i < 0 || i >= len(r.cells) {
		return nil
	}
	return r.cells[i]
}

// AddCell appends a cell to this row.
func (r *TableRow) AddCell(c *TableCell) {
	r.cells = append(r.cells, c)
}

// CellCount returns the number of cells.
func (r *TableRow) CellCount() int { return len(r.cells) }

// MinHeight returns the minimum row height.
func (r *TableRow) MinHeight() float32 { return r.minHeight }

// SetMinHeight sets the minimum row height.
func (r *TableRow) SetMinHeight(v float32) { r.minHeight = v }

// MaxHeight returns the maximum row height (0 = unlimited).
func (r *TableRow) MaxHeight() float32 { return r.maxHeight }

// SetMaxHeight sets the maximum row height.
func (r *TableRow) SetMaxHeight(v float32) { r.maxHeight = v }

// AutoSize returns whether the row calculates its height automatically.
func (r *TableRow) AutoSize() bool { return r.autoSize }

// SetAutoSize sets the auto-size flag.
func (r *TableRow) SetAutoSize(v bool) { r.autoSize = v }

// CanBreak returns whether the row may be split across pages.
func (r *TableRow) CanBreak() bool { return r.canBreak }

// SetCanBreak sets the can-break flag.
func (r *TableRow) SetCanBreak(v bool) { r.canBreak = v }

// PageBreak returns whether a page break precedes this row.
func (r *TableRow) PageBreak() bool { return r.pageBreak }

// SetPageBreak sets the page-break flag.
func (r *TableRow) SetPageBreak(v bool) { r.pageBreak = v }

// KeepRows returns how many rows to keep together on a page.
func (r *TableRow) KeepRows() int { return r.keepRows }

// SetKeepRows sets the keep-rows count.
func (r *TableRow) SetKeepRows(v int) { r.keepRows = v }

// Serialize writes TableRow properties that differ from defaults.
func (r *TableRow) Serialize(w report.Writer) error {
	if err := r.ComponentBase.Serialize(w); err != nil {
		return err
	}
	if r.minHeight != 0 {
		w.WriteFloat("MinHeight", r.minHeight)
	}
	if r.maxHeight != 1000 {
		w.WriteFloat("MaxHeight", r.maxHeight)
	}
	if r.autoSize {
		w.WriteBool("AutoSize", true)
	}
	if r.canBreak {
		w.WriteBool("CanBreak", true)
	}
	if r.pageBreak {
		w.WriteBool("PageBreak", true)
	}
	if r.keepRows != 0 {
		w.WriteInt("KeepRows", r.keepRows)
	}
	// Serialize cells.
	for _, c := range r.cells {
		if err := w.WriteObject(c); err != nil {
			return err
		}
	}
	return nil
}

// Deserialize reads TableRow properties and child cells.
func (r *TableRow) Deserialize(rd report.Reader) error {
	if err := r.ComponentBase.Deserialize(rd); err != nil {
		return err
	}
	// ComponentBase.Deserialize defaults Height to 0; re-read with the table row
	// default of 30 so that rows without an explicit Height attribute are usable.
	r.SetHeight(rd.ReadFloat("Height", 30))
	r.minHeight = rd.ReadFloat("MinHeight", 0)
	r.maxHeight = rd.ReadFloat("MaxHeight", 1000)
	r.autoSize = rd.ReadBool("AutoSize", false)
	r.canBreak = rd.ReadBool("CanBreak", false)
	r.pageBreak = rd.ReadBool("PageBreak", false)
	r.keepRows = rd.ReadInt("KeepRows", 0)
	return nil
}
