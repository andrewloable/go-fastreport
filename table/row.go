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

	// savedHeight / savedVisible store state for SaveState/RestoreState.
	savedHeight  float32
	savedVisible bool
}

// NewTableRow creates a TableRow with defaults matching the C# constructor.
func NewTableRow() *TableRow {
	r := &TableRow{
		ComponentBase: *report.NewComponentBase(),
		maxHeight:     1000, // matches C# DefaultValue(1000)
	}
	// C# DefaultHeight: (int)Math.Round(18 / (0.25f * Units.Centimeters)) * (0.25f * Units.Centimeters)
	// = round(18 / 9.45) * 9.45 = 2 * 9.45 = 18.9
	r.SetHeight(18.9)
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

// SetHeight overrides ComponentBase.SetHeight to enforce min/max bounds.
// If height < minHeight (and minHeight > 0), it is clamped up.
// If height > maxHeight (and maxHeight > 0 and canBreak is false), it is clamped down.
func (r *TableRow) SetHeight(v float32) {
	if r.minHeight > 0 && v < r.minHeight {
		v = r.minHeight
	}
	if !r.canBreak && r.maxHeight > 0 && v > r.maxHeight {
		v = r.maxHeight
	}
	r.ComponentBase.SetHeight(v)
}

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

// Assign copies all TableRow properties from src.
// Mirrors C# TableRow.Assign (TableRow.cs:288-297).
func (r *TableRow) Assign(src *TableRow) {
	if src == nil {
		return
	}
	r.minHeight = src.minHeight
	r.maxHeight = src.maxHeight
	r.autoSize = src.autoSize
	r.keepRows = src.keepRows
	r.canBreak = src.canBreak
	r.pageBreak = src.pageBreak
	r.ComponentBase.Assign(&src.ComponentBase)
	// Note: cells are not copied — structural children managed by the table.
}

// Clear resets the row and clears its cells.
// Mirrors C# TableRow.Clear (TableRow.cs:361-368).
func (r *TableRow) Clear() {
	r.cells = nil
	r.SetHeight(30)
}

// SaveState saves the current Height and Visible values so they can be
// restored later with RestoreState.
// Mirrors C# TableRow.SaveState (TableRow.cs).
func (r *TableRow) SaveState() {
	r.savedHeight = r.Height()
	r.savedVisible = r.Visible()
}

// RestoreState restores Height and Visible to the values saved by SaveState.
// Mirrors C# TableRow.RestoreState (TableRow.cs).
func (r *TableRow) RestoreState() {
	r.SetHeight(r.savedHeight)
	r.SetVisible(r.savedVisible)
}

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
	// ComponentBase.Deserialize defaults Height to 0; re-read with the C# default
	// of 18.9 so that rows without an explicit Height attribute match C# output.
	r.SetHeight(rd.ReadFloat("Height", 18.9))
	r.minHeight = rd.ReadFloat("MinHeight", 0)
	r.maxHeight = rd.ReadFloat("MaxHeight", 1000)
	r.autoSize = rd.ReadBool("AutoSize", false)
	r.canBreak = rd.ReadBool("CanBreak", false)
	r.pageBreak = rd.ReadBool("PageBreak", false)
	r.keepRows = rd.ReadInt("KeepRows", 0)
	return nil
}
