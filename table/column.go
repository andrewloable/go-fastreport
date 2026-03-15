package table

import (
	"github.com/andrewloable/go-fastreport/report"
)

// TableColumn represents a column in a table object.
// It is the Go equivalent of FastReport.Table.TableColumn.
type TableColumn struct {
	report.ComponentBase

	// minWidth is the minimum column width in pixels.
	minWidth float32
	// maxWidth is the maximum column width in pixels. 0 means unlimited.
	maxWidth float32
	// autoSize makes the column calculate its width automatically.
	autoSize bool
	// pageBreak inserts a page break before this column.
	pageBreak bool
	// keepColumns specifies how many columns to keep together on a page.
	keepColumns int
}

// NewTableColumn creates a TableColumn with defaults matching the C# constructor.
func NewTableColumn() *TableColumn {
	c := &TableColumn{
		ComponentBase: *report.NewComponentBase(),
		maxWidth:      5000, // matches C# DefaultValue(5000)
	}
	c.SetWidth(100) // default column width
	return c
}

// MinWidth returns the minimum column width in pixels.
func (c *TableColumn) MinWidth() float32 { return c.minWidth }

// SetMinWidth sets the minimum column width.
func (c *TableColumn) SetMinWidth(v float32) { c.minWidth = v }

// MaxWidth returns the maximum column width in pixels (0 = unlimited).
func (c *TableColumn) MaxWidth() float32 { return c.maxWidth }

// SetMaxWidth sets the maximum column width.
func (c *TableColumn) SetMaxWidth(v float32) { c.maxWidth = v }

// AutoSize returns whether the column calculates its width automatically.
func (c *TableColumn) AutoSize() bool { return c.autoSize }

// SetAutoSize sets the auto-size flag.
func (c *TableColumn) SetAutoSize(v bool) { c.autoSize = v }

// PageBreak returns whether a page break precedes this column.
func (c *TableColumn) PageBreak() bool { return c.pageBreak }

// SetPageBreak sets the page-break flag.
func (c *TableColumn) SetPageBreak(v bool) { c.pageBreak = v }

// KeepColumns returns how many columns to keep together on a page.
func (c *TableColumn) KeepColumns() int { return c.keepColumns }

// SetKeepColumns sets the keep-columns count.
func (c *TableColumn) SetKeepColumns(v int) { c.keepColumns = v }

// Serialize writes TableColumn properties that differ from defaults.
func (c *TableColumn) Serialize(w report.Writer) error {
	if err := c.ComponentBase.Serialize(w); err != nil {
		return err
	}
	if c.minWidth != 0 {
		w.WriteFloat("MinWidth", c.minWidth)
	}
	if c.maxWidth != 5000 {
		w.WriteFloat("MaxWidth", c.maxWidth)
	}
	if c.autoSize {
		w.WriteBool("AutoSize", true)
	}
	if c.pageBreak {
		w.WriteBool("PageBreak", true)
	}
	if c.keepColumns != 0 {
		w.WriteInt("KeepColumns", c.keepColumns)
	}
	return nil
}

// Deserialize reads TableColumn properties.
func (c *TableColumn) Deserialize(r report.Reader) error {
	if err := c.ComponentBase.Deserialize(r); err != nil {
		return err
	}
	c.minWidth = r.ReadFloat("MinWidth", 0)
	c.maxWidth = r.ReadFloat("MaxWidth", 5000)
	c.autoSize = r.ReadBool("AutoSize", false)
	c.pageBreak = r.ReadBool("PageBreak", false)
	c.keepColumns = r.ReadInt("KeepColumns", 0)
	return nil
}
