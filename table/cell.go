package table

import (
	"github.com/andrewloable/go-fastreport/object"
	"github.com/andrewloable/go-fastreport/report"
	"github.com/andrewloable/go-fastreport/serial"
	"github.com/andrewloable/go-fastreport/style"
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

	// savedText / savedObjectCount for SaveState/RestoreState.
	savedText        string
	savedObjectCount int
}

// TypeName returns the FRX element name.
func (c *TableCell) TypeName() string { return "TableCell" }

// NewTableCell creates a TableCell with default spans of 1.
// Padding defaults to (2,1,2,1) matching C# TableCell (TableCell.cs line 508).
func NewTableCell() *TableCell {
	c := &TableCell{
		TextObject: *object.NewTextObject(),
		colSpan:    1,
		rowSpan:    1,
	}
	c.SetPadding(object.Padding{Left: 2, Top: 1, Right: 2, Bottom: 1})
	return c
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

// ReplaceObject replaces the embedded object at index i.
func (c *TableCell) ReplaceObject(i int, obj report.Serializable) {
	if i >= 0 && i < len(c.objects) {
		c.objects[i] = obj
	}
}

// ObjectCount returns the number of embedded objects.
func (c *TableCell) ObjectCount() int { return len(c.objects) }

// DeserializeChild handles child XML elements inside a TableCell during FRX
// deserialization. Embedded report objects (e.g. PictureObject) are created
// via the serial registry and added to the cell's objects collection.
// Mirrors C# TableCell which inherits IParent.AddChild (ReportComponentBase.cs).
func (c *TableCell) DeserializeChild(childType string, r report.Reader) bool {
	obj, err := serial.DefaultRegistry.Create(childType)
	if err != nil || obj == nil {
		return false
	}
	_ = obj.Deserialize(r)
	if s, ok := obj.(report.Serializable); ok {
		c.AddObject(s)
	}
	return true
}

// CalcWidth returns the current width of the cell.
// Mirrors C# TableCell.CalcWidth (TableCell.cs).
func (c *TableCell) CalcWidth() float32 { return c.Width() }

// CalcHeight returns the current height of the cell.
// Mirrors C# TableCell.CalcHeight (TableCell.cs).
func (c *TableCell) CalcHeight() float32 { return c.Height() }

// SaveState saves the current Text and embedded object count so they can be
// restored later with RestoreState.
// Mirrors C# TableCell.SaveState (TableCell.cs).
func (c *TableCell) SaveState() {
	c.savedText = c.Text()
	c.savedObjectCount = len(c.objects)
}

// RestoreState restores Text and trims the objects slice back to the count
// saved by SaveState (discarding any objects added after the save).
// Mirrors C# TableCell.RestoreState (TableCell.cs).
func (c *TableCell) RestoreState() {
	c.SetText(c.savedText)
	if c.savedObjectCount <= len(c.objects) {
		c.objects = c.objects[:c.savedObjectCount]
	}
}

// Assign copies all TableCell properties from src.
// Mirrors C# TableCell.Assign (TableCell.cs:221-228).
func (c *TableCell) Assign(src *TableCell) {
	if src == nil {
		return
	}
	// Copy the embedded TextObject by value (copies all scalar fields), then
	// deep-copy reference types to ensure independence.
	c.TextObject = src.TextObject
	// Deep-copy highlights slice.
	srcH := src.TextObject.Highlights()
	if len(srcH) > 0 {
		h := make([]style.HighlightCondition, len(srcH))
		copy(h, srcH)
		c.TextObject.SetHighlights(h)
	} else {
		c.TextObject.SetHighlights(nil)
	}
	// Copy table-cell-specific fields.
	c.colSpan = src.colSpan
	c.rowSpan = src.rowSpan
	c.duplicates = src.duplicates
	// objects not copied (structural; managed by table).
}

// Clone creates an exact copy of this cell.
// Mirrors C# TableCell.Clone (TableCell.cs:235-239).
func (c *TableCell) Clone() *TableCell {
	cell := NewTableCell()
	cell.Assign(c)
	return cell
}

// EqualStyle returns true when two cells have identical visual style settings.
// This is used for style deduplication (mirrors C# TableCell.Equals —
// TableCell.cs:247-283, but named EqualStyle to avoid shadowing built-in).
func (c *TableCell) EqualStyle(other *TableCell) bool {
	if other == nil {
		return false
	}
	return c.HorzAlign() == other.HorzAlign() &&
		c.VertAlign() == other.VertAlign() &&
		c.Angle() == other.Angle() &&
		c.RightToLeft() == other.RightToLeft() &&
		c.WordWrap() == other.WordWrap() &&
		c.Underlines() == other.Underlines() &&
		c.Clip() == other.Clip() &&
		c.Wysiwyg() == other.Wysiwyg() &&
		c.Font() == other.Font() &&
		c.TextColor() == other.TextColor() &&
		c.FontWidthRatio() == other.FontWidthRatio() &&
		c.FirstTabOffset() == other.FirstTabOffset() &&
		c.TabWidth() == other.TabWidth() &&
		c.LineHeight() == other.LineHeight() &&
		c.ParagraphOffset() == other.ParagraphOffset() &&
		c.duplicates == other.duplicates
}

// GetData populates the cell with its data value. When insideSpan is true
// (the cell is covered by a spanning neighbour), its text is cleared so the
// spanning cell's content shows through.
// Mirrors C# TableCell.GetData (TableCell.cs).
func (c *TableCell) GetData(insideSpan bool) {
	if insideSpan {
		c.SetText("")
	}
}

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
