package object

import (
	"github.com/andrewloable/go-fastreport/report"
	"github.com/andrewloable/go-fastreport/style"
)

// CellularTextObject renders text in a grid of fixed-size cells — one character
// per cell. It is the Go equivalent of FastReport.CellularTextObject.
//
// It extends TextObject with cell-size and spacing controls. The engine renders
// it as a text object; the cellular grid visualization is handled at export time.
type CellularTextObject struct {
	TextObject

	// cellWidth is the width of each character cell in pixels (0 = auto from font).
	cellWidth float32
	// cellHeight is the height of each character cell in pixels (0 = auto from font).
	cellHeight float32
	// horzSpacing is the horizontal gap between cells in pixels.
	horzSpacing float32
	// vertSpacing is the vertical gap between cells in pixels.
	vertSpacing float32
	// wordWrap controls whether text wraps to a new line when it reaches the end.
	wordWrap bool
}

// NewCellularTextObject creates a CellularTextObject with defaults.
//
// The C# constructor (CellularTextObject.cs) explicitly sets:
//   - CanBreak = false  (text cells must not split across pages)
//   - Border.Lines = BorderLines.All  (all four cell borders visible by default)
//
// WordWrap defaults to true (inherited from TextObject).
func NewCellularTextObject() *CellularTextObject {
	c := &CellularTextObject{
		TextObject: *NewTextObject(),
		wordWrap:   true,
	}
	// Match C# CellularTextObject() constructor: CanBreak = false.
	c.SetCanBreak(false)
	// Match C# CellularTextObject() constructor: Border.Lines = BorderLines.All.
	border := c.Border()
	border.VisibleLines = style.BorderLinesAll
	c.SetBorder(border)
	return c
}

// BaseName returns the base name prefix for auto-generated names.
func (c *CellularTextObject) BaseName() string { return "CellularText" }

// TypeName returns "CellularTextObject".
func (c *CellularTextObject) TypeName() string { return "CellularTextObject" }

// CellWidth returns the character cell width in pixels (0 = auto).
func (c *CellularTextObject) CellWidth() float32 { return c.cellWidth }

// SetCellWidth sets the character cell width.
func (c *CellularTextObject) SetCellWidth(v float32) { c.cellWidth = v }

// CellHeight returns the character cell height in pixels (0 = auto).
func (c *CellularTextObject) CellHeight() float32 { return c.cellHeight }

// SetCellHeight sets the character cell height.
func (c *CellularTextObject) SetCellHeight(v float32) { c.cellHeight = v }

// HorzSpacing returns the horizontal gap between cells in pixels.
func (c *CellularTextObject) HorzSpacing() float32 { return c.horzSpacing }

// SetHorzSpacing sets the horizontal cell gap.
func (c *CellularTextObject) SetHorzSpacing(v float32) { c.horzSpacing = v }

// VertSpacing returns the vertical gap between cells in pixels.
func (c *CellularTextObject) VertSpacing() float32 { return c.vertSpacing }

// SetVertSpacing sets the vertical cell gap.
func (c *CellularTextObject) SetVertSpacing(v float32) { c.vertSpacing = v }

// WordWrap returns whether text wraps within the object bounds.
func (c *CellularTextObject) WordWrap() bool { return c.wordWrap }

// SetWordWrap sets the word-wrap flag.
func (c *CellularTextObject) SetWordWrap(v bool) { c.wordWrap = v }

// Serialize writes CellularTextObject properties that differ from defaults.
func (c *CellularTextObject) Serialize(w report.Writer) error {
	if err := c.TextObject.Serialize(w); err != nil {
		return err
	}
	if c.cellWidth != 0 {
		w.WriteFloat("CellWidth", c.cellWidth)
	}
	if c.cellHeight != 0 {
		w.WriteFloat("CellHeight", c.cellHeight)
	}
	if c.horzSpacing != 0 {
		w.WriteFloat("HorzSpacing", c.horzSpacing)
	}
	if c.vertSpacing != 0 {
		w.WriteFloat("VertSpacing", c.vertSpacing)
	}
	if !c.wordWrap {
		w.WriteBool("WordWrap", false)
	}
	return nil
}

// Deserialize reads CellularTextObject properties.
func (c *CellularTextObject) Deserialize(r report.Reader) error {
	if err := c.TextObject.Deserialize(r); err != nil {
		return err
	}
	c.cellWidth = r.ReadFloat("CellWidth", 0)
	c.cellHeight = r.ReadFloat("CellHeight", 0)
	c.horzSpacing = r.ReadFloat("HorzSpacing", 0)
	c.vertSpacing = r.ReadFloat("VertSpacing", 0)
	c.wordWrap = r.ReadBool("WordWrap", true)
	return nil
}
