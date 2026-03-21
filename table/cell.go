package table

import (
	"github.com/andrewloable/go-fastreport/object"
	"github.com/andrewloable/go-fastreport/report"
)

// CellDuplicates specifies how duplicate cell values are displayed.
type CellDuplicates int

const (
	// CellDuplicatesShow shows duplicate values.
	CellDuplicatesShow CellDuplicates = iota
	// CellDuplicatesClear shows the cell but with no text on duplicate values.
	CellDuplicatesClear
	// CellDuplicatesMerge merges adjacent cells with the same value.
	CellDuplicatesMerge
	// CellDuplicatesMergeNonEmpty merges adjacent cells with the same non-empty value.
	CellDuplicatesMergeNonEmpty
)

// cellDuplicatesName returns the FRX enum name string for CellDuplicates.
func cellDuplicatesName(d CellDuplicates) string {
	switch d {
	case CellDuplicatesClear:
		return "Clear"
	case CellDuplicatesMerge:
		return "Merge"
	case CellDuplicatesMergeNonEmpty:
		return "MergeNonEmpty"
	default:
		return "Show"
	}
}

// parseCellDuplicates parses the FRX enum name string into a CellDuplicates value.
func parseCellDuplicates(s string) CellDuplicates {
	switch s {
	case "Clear":
		return CellDuplicatesClear
	case "Merge":
		return CellDuplicatesMerge
	case "MergeNonEmpty":
		return CellDuplicatesMergeNonEmpty
	default:
		return CellDuplicatesShow
	}
}

// TableCell represents a single cell in a TableObject.
// It embeds TextObject for text rendering and adds ColSpan, RowSpan and a
// nested objects collection.
// It is the Go equivalent of FastReport.Table.TableCell.
type TableCell struct {
	object.TextObject

	// colSpan is the number of columns this cell spans (default 1).
	colSpan int
	// rowSpan is the number of rows this cell spans (default 1).
	rowSpan int

	// objects holds additional report components embedded in the cell
	// (e.g. PictureObjects).
	objects []report.Serializable

	// duplicates controls how duplicate values are handled.
	duplicates CellDuplicates
}

// TypeName returns the FRX element name.
func (c *TableCell) TypeName() string { return "TableCell" }

// NewTableCell creates a TableCell with default spans of 1.
func NewTableCell() *TableCell {
	return &TableCell{
		TextObject: *object.NewTextObject(),
		colSpan:    1,
		rowSpan:    1,
	}
}

// ColSpan returns the column span (number of columns this cell covers).
func (c *TableCell) ColSpan() int { return c.colSpan }

// SetColSpan sets the column span. Values < 1 are clamped to 1.
func (c *TableCell) SetColSpan(v int) {
	if v < 1 {
		v = 1
	}
	c.colSpan = v
}

// RowSpan returns the row span (number of rows this cell covers).
func (c *TableCell) RowSpan() int { return c.rowSpan }

// SetRowSpan sets the row span. Values < 1 are clamped to 1.
func (c *TableCell) SetRowSpan(v int) {
	if v < 1 {
		v = 1
	}
	c.rowSpan = v
}

// Duplicates returns the cell's duplicate-value handling mode.
func (c *TableCell) Duplicates() CellDuplicates { return c.duplicates }

// SetDuplicates sets the duplicate-value handling mode.
func (c *TableCell) SetDuplicates(d CellDuplicates) { c.duplicates = d }

// Objects returns embedded report components inside the cell.
func (c *TableCell) Objects() []report.Serializable { return c.objects }

// AddObject adds an embedded component to the cell.
func (c *TableCell) AddObject(obj report.Serializable) {
	c.objects = append(c.objects, obj)
}

// ObjectCount returns the number of embedded objects.
func (c *TableCell) ObjectCount() int { return len(c.objects) }

// SetStyle applies a style to this cell through the given table's style
// deduplication collection. The returned cell is the canonical shared style
// instance from the collection — callers that need to track the applied style
// should use the return value.
//
// This mirrors C# TableCell.SetStyle → TableCellData.SetStyle (TableCell.cs
// line 328, TableCellData.cs line 326-329): the style is deduplicated through
// Table.Styles so that cells sharing identical visual appearance reference the
// same style object.
//
// If tbl is nil the style argument is returned unchanged.
func (c *TableCell) SetStyle(tbl *TableBase, style *TableCell) *TableCell {
	if tbl == nil {
		return style
	}
	return tbl.styles.Add(style)
}

// Serialize writes TableCell properties that differ from defaults.
func (c *TableCell) Serialize(w report.Writer) error {
	if err := c.TextObject.Serialize(w); err != nil {
		return err
	}
	if c.colSpan != 1 {
		w.WriteInt("ColSpan", c.colSpan)
	}
	if c.rowSpan != 1 {
		w.WriteInt("RowSpan", c.rowSpan)
	}
	// CellDuplicates is written as an enum name string (C# WriteValue convention).
	if c.duplicates != CellDuplicatesShow {
		w.WriteStr("CellDuplicates", cellDuplicatesName(c.duplicates))
	}
	for _, obj := range c.objects {
		if err := w.WriteObject(obj); err != nil {
			return err
		}
	}
	return nil
}

// Deserialize reads TableCell properties.
func (c *TableCell) Deserialize(r report.Reader) error {
	if err := c.TextObject.Deserialize(r); err != nil {
		return err
	}
	c.colSpan = r.ReadInt("ColSpan", 1)
	c.rowSpan = r.ReadInt("RowSpan", 1)
	c.duplicates = parseCellDuplicates(r.ReadStr("CellDuplicates", "Show"))
	if c.colSpan < 1 {
		c.colSpan = 1
	}
	if c.rowSpan < 1 {
		c.rowSpan = 1
	}
	return nil
}
