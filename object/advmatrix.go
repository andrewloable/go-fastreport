package object

import (
	"image/color"

	"github.com/andrewloable/go-fastreport/report"
	"github.com/andrewloable/go-fastreport/style"
	"github.com/andrewloable/go-fastreport/utils"
)


// AdvMatrixDescriptor holds the definition of a row or column dimension
// in an AdvMatrixObject.
type AdvMatrixDescriptor struct {
	// Expression is the data binding expression for this dimension.
	Expression string
	// DisplayText is the header text (may contain expressions).
	DisplayText string
	// Sort is the sort direction ("Ascending", "Descending", or "").
	Sort string
	// Children are nested descriptor levels.
	Children []*AdvMatrixDescriptor
}

// AdvMatrixColumn holds the physical column definition from the FRX layout.
type AdvMatrixColumn struct {
	Name     string
	Width    float32
	AutoSize bool
}

// AdvMatrixCell holds a single cell within a physical row of the AdvMatrix table.
type AdvMatrixCell struct {
	Name      string
	Width     float32
	Height    float32
	ColSpan   int
	RowSpan   int
	Text      string
	HorzAlign int
	VertAlign int
	// Style fields parsed from FRX attributes.
	Border    *style.Border // nil = no border
	FillColor *color.RGBA   // nil = no fill
	Font      *style.Font   // nil = inherit default
	// Buttons contains any MatrixCollapseButton/MatrixSortButton children.
	Buttons []*MatrixButton
}

// MatrixButton holds the minimal properties of a MatrixCollapseButton or
// MatrixSortButton embedded in a TableCell.
type MatrixButton struct {
	// TypeName is "MatrixCollapseButton" or "MatrixSortButton".
	TypeName string
	Name     string
	Left     float32
	Width    float32
	Height   float32
	Dock     string
	SymbolSize float32
	Symbol   string
	ShowCollapseExpandMenu bool
}

// AdvMatrixRow holds a physical row definition and its cells.
type AdvMatrixRow struct {
	Name     string
	Height   float32
	AutoSize bool
	Cells    []*AdvMatrixCell
}

// AdvMatrixObject is an advanced pivot-matrix object that generates
// cross-tab reports with collapsible/sortable rows and columns.
//
// It is the Go equivalent of FastReport.AdvMatrix.AdvMatrixObject.
// This implementation supports FRX loading (deserialization) and serialization
// for round-trip fidelity. The physical template cells (TableRows/TableColumns)
// are rendered by populateAdvMatrixCells in the report engine, with ColSpan,
// RowSpan, fill color, border, and font all resolved from the deserialized FRX.
type AdvMatrixObject struct {
	report.ReportComponentBase

	// DataSource is the name of the bound data source.
	DataSource string
	// Columns holds the column dimension descriptors (pivot axes).
	Columns []*AdvMatrixDescriptor
	// Rows holds the row dimension descriptors (pivot axes).
	Rows []*AdvMatrixDescriptor

	// TableColumns holds the physical column width definitions from the FRX.
	TableColumns []*AdvMatrixColumn
	// TableRows holds the physical row definitions (with their cells) from the FRX.
	TableRows []*AdvMatrixRow
}

// NewAdvMatrixObject creates an AdvMatrixObject with defaults.
func NewAdvMatrixObject() *AdvMatrixObject {
	return &AdvMatrixObject{
		ReportComponentBase: *report.NewReportComponentBase(),
	}
}

// BaseName returns the base name prefix for auto-generated names.
func (a *AdvMatrixObject) BaseName() string { return "AdvMatrix" }

// TypeName returns "AdvMatrixObject".
func (a *AdvMatrixObject) TypeName() string { return "AdvMatrixObject" }

// Serialize writes AdvMatrixObject properties that differ from defaults.
func (a *AdvMatrixObject) Serialize(w report.Writer) error {
	if err := a.ReportComponentBase.Serialize(w); err != nil {
		return err
	}
	if a.DataSource != "" {
		w.WriteStr("DataSource", a.DataSource)
	}
	// Write physical column definitions as "TableColumn" child elements.
	for _, col := range a.TableColumns {
		if err := w.WriteObjectNamed("TableColumn", col); err != nil {
			return err
		}
	}
	// Write physical row definitions (with their cells) as "TableRow" children.
	for _, row := range a.TableRows {
		if err := w.WriteObjectNamed("TableRow", row); err != nil {
			return err
		}
	}
	return nil
}

// ── Serialize/Deserialize for sub-types ───────────────────────────────────────

// Serialize writes AdvMatrixColumn properties.
func (c *AdvMatrixColumn) Serialize(w report.Writer) error {
	if c.Name != "" {
		w.WriteStr("Name", c.Name)
	}
	if c.Width != 0 {
		w.WriteFloat("Width", c.Width)
	}
	if c.AutoSize {
		w.WriteBool("AutoSize", true)
	}
	return nil
}

// Deserialize reads AdvMatrixColumn properties.
func (c *AdvMatrixColumn) Deserialize(r report.Reader) error {
	c.Name = r.ReadStr("Name", "")
	c.Width = r.ReadFloat("Width", 0)
	c.AutoSize = r.ReadBool("AutoSize", false)
	return nil
}

// Serialize writes AdvMatrixRow properties and its cell children.
func (row *AdvMatrixRow) Serialize(w report.Writer) error {
	if row.Name != "" {
		w.WriteStr("Name", row.Name)
	}
	if row.Height != 0 {
		w.WriteFloat("Height", row.Height)
	}
	if row.AutoSize {
		w.WriteBool("AutoSize", true)
	}
	for _, cell := range row.Cells {
		if err := w.WriteObjectNamed("TableCell", cell); err != nil {
			return err
		}
	}
	return nil
}

// Deserialize reads AdvMatrixRow properties (cells deserialized via DeserializeChild).
func (row *AdvMatrixRow) Deserialize(r report.Reader) error {
	row.Name = r.ReadStr("Name", "")
	row.Height = r.ReadFloat("Height", 0)
	row.AutoSize = r.ReadBool("AutoSize", false)
	return nil
}

// Serialize writes AdvMatrixCell properties and button children.
func (cell *AdvMatrixCell) Serialize(w report.Writer) error {
	if cell.Name != "" {
		w.WriteStr("Name", cell.Name)
	}
	if cell.Width != 0 {
		w.WriteFloat("Width", cell.Width)
	}
	if cell.Height != 0 {
		w.WriteFloat("Height", cell.Height)
	}
	if cell.ColSpan != 1 && cell.ColSpan != 0 {
		w.WriteInt("ColSpan", cell.ColSpan)
	}
	if cell.RowSpan != 1 && cell.RowSpan != 0 {
		w.WriteInt("RowSpan", cell.RowSpan)
	}
	if cell.Text != "" {
		w.WriteStr("Text", cell.Text)
	}
	if cell.HorzAlign != 0 {
		w.WriteInt("HorzAlign", cell.HorzAlign)
	}
	if cell.VertAlign != 0 {
		w.WriteInt("VertAlign", cell.VertAlign)
	}
	if cell.Border != nil && cell.Border.VisibleLines != 0 {
		bl := formatBorderLinesStr(cell.Border.VisibleLines)
		if bl != "" {
			w.WriteStr("Border.Lines", bl)
		}
		if len(cell.Border.Lines) > 0 && cell.Border.Lines[0] != nil {
			c := cell.Border.Lines[0].Color
			w.WriteStr("Border.Color", utils.FormatColor(c))
		}
	}
	if cell.FillColor != nil {
		w.WriteStr("Fill.Color", utils.FormatColor(*cell.FillColor))
	}
	if cell.Font != nil {
		w.WriteStr("Font", style.FontToStr(*cell.Font))
	}
	for _, btn := range cell.Buttons {
		if err := w.WriteObjectNamed(btn.TypeName, btn); err != nil {
			return err
		}
	}
	return nil
}

// formatBorderLinesStr converts a BorderLines bitmask to the FRX string form.
func formatBorderLinesStr(bl style.BorderLines) string {
	switch bl {
	case style.BorderLinesAll:
		return "All"
	case style.BorderLinesNone:
		return "None"
	case style.BorderLinesLeft:
		return "Left"
	case style.BorderLinesRight:
		return "Right"
	case style.BorderLinesTop:
		return "Top"
	case style.BorderLinesBottom:
		return "Bottom"
	default:
		return ""
	}
}

// Deserialize reads AdvMatrixCell properties.
func (cell *AdvMatrixCell) Deserialize(r report.Reader) error {
	cell.Name = r.ReadStr("Name", "")
	cell.Width = r.ReadFloat("Width", 0)
	cell.Height = r.ReadFloat("Height", 0)
	cell.ColSpan = r.ReadInt("ColSpan", 1)
	cell.RowSpan = r.ReadInt("RowSpan", 1)
	cell.Text = r.ReadStr("Text", "")
	cell.HorzAlign = int(ParseHorzAlign(r.ReadStr("HorzAlign", "Left")))
	cell.VertAlign = int(ParseVertAlign(r.ReadStr("VertAlign", "Top")))
	if cell.ColSpan < 1 {
		cell.ColSpan = 1
	}
	if cell.RowSpan < 1 {
		cell.RowSpan = 1
	}
	// Parse border.
	borderLines := r.ReadStr("Border.Lines", "")
	borderColor := r.ReadStr("Border.Color", "")
	if borderLines != "" || borderColor != "" {
		b := style.NewBorder()
		if borderLines != "" {
			bl := parseBorderLinesStr(borderLines)
			b.VisibleLines = bl
		}
		if borderColor != "" {
			if c, err := utils.ParseColor(borderColor); err == nil {
				for i := range b.Lines {
					if b.Lines[i] != nil {
						b.Lines[i].Color = c
					}
				}
			}
		}
		cell.Border = b
	}
	// Parse fill color.
	if fc := r.ReadStr("Fill.Color", ""); fc != "" {
		if c, err := utils.ParseColor(fc); err == nil {
			cell.FillColor = &c
		}
	}
	// Parse font.
	if fs := r.ReadStr("Font", ""); fs != "" {
		f := style.FontFromStr(fs)
		cell.Font = &f
	}
	return nil
}

// parseBorderLinesStr converts the FRX Border.Lines string to a BorderLines bitmask.
func parseBorderLinesStr(s string) style.BorderLines {
	switch s {
	case "All":
		return style.BorderLinesLeft | style.BorderLinesRight | style.BorderLinesTop | style.BorderLinesBottom
	case "None":
		return 0
	case "Left":
		return style.BorderLinesLeft
	case "Right":
		return style.BorderLinesRight
	case "Top":
		return style.BorderLinesTop
	case "Bottom":
		return style.BorderLinesBottom
	default:
		return 0
	}
}

// Serialize writes MatrixButton properties.
func (btn *MatrixButton) Serialize(w report.Writer) error {
	if btn.Name != "" {
		w.WriteStr("Name", btn.Name)
	}
	if btn.Left != 0 {
		w.WriteFloat("Left", btn.Left)
	}
	if btn.Width != 0 {
		w.WriteFloat("Width", btn.Width)
	}
	if btn.Height != 0 {
		w.WriteFloat("Height", btn.Height)
	}
	if btn.Dock != "" {
		w.WriteStr("Dock", btn.Dock)
	}
	if btn.SymbolSize != 0 {
		w.WriteFloat("SymbolSize", btn.SymbolSize)
	}
	if btn.Symbol != "" {
		w.WriteStr("Symbol", btn.Symbol)
	}
	if btn.ShowCollapseExpandMenu {
		w.WriteBool("ShowCollapseExpandMenu", true)
	}
	return nil
}

// Deserialize reads MatrixButton properties.
func (btn *MatrixButton) Deserialize(r report.Reader) error {
	btn.Name = r.ReadStr("Name", "")
	btn.Left = r.ReadFloat("Left", 0)
	btn.Width = r.ReadFloat("Width", 0)
	btn.Height = r.ReadFloat("Height", 0)
	btn.Dock = r.ReadStr("Dock", "")
	btn.SymbolSize = r.ReadFloat("SymbolSize", 0)
	btn.Symbol = r.ReadStr("Symbol", "")
	btn.ShowCollapseExpandMenu = r.ReadBool("ShowCollapseExpandMenu", false)
	return nil
}

// Deserialize reads AdvMatrixObject properties from an FRX reader.
func (a *AdvMatrixObject) Deserialize(r report.Reader) error {
	if err := a.ReportComponentBase.Deserialize(r); err != nil {
		return err
	}
	a.DataSource = r.ReadStr("DataSource", "")
	return nil
}

// DeserializeChild handles child elements of AdvMatrixObject.
// It reads TableColumn/TableRow (with TableCell and button children),
// and dimension Columns/Rows descriptor blocks.
func (a *AdvMatrixObject) DeserializeChild(childType string, r report.Reader) bool {
	switch childType {
	case "TableColumn":
		col := &AdvMatrixColumn{
			Name:     r.ReadStr("Name", ""),
			Width:    r.ReadFloat("Width", 0),
			AutoSize: r.ReadBool("AutoSize", false),
		}
		a.TableColumns = append(a.TableColumns, col)
		// TableColumn has no children.
		return true

	case "TableRow":
		row := &AdvMatrixRow{
			Name:     r.ReadStr("Name", ""),
			Height:   r.ReadFloat("Height", 0),
			AutoSize: r.ReadBool("AutoSize", false),
		}
		// Iterate TableCell children of this row.
		for {
			ct, ok := r.NextChild()
			if !ok {
				break
			}
			if ct == "TableCell" {
				cell := &AdvMatrixCell{}
				_ = cell.Deserialize(r)
				// Iterate children of the cell (MatrixCollapseButton, MatrixSortButton).
				for {
					btnType, ok2 := r.NextChild()
					if !ok2 {
						break
					}
					if btnType == "MatrixCollapseButton" || btnType == "MatrixSortButton" {
						btn := &MatrixButton{
							TypeName:               btnType,
							Name:                   r.ReadStr("Name", ""),
							Left:                   r.ReadFloat("Left", 0),
							Width:                  r.ReadFloat("Width", 0),
							Height:                 r.ReadFloat("Height", 0),
							Dock:                   r.ReadStr("Dock", ""),
							SymbolSize:             r.ReadFloat("SymbolSize", 0),
							Symbol:                 r.ReadStr("Symbol", ""),
							ShowCollapseExpandMenu: r.ReadBool("ShowCollapseExpandMenu", false),
						}
						cell.Buttons = append(cell.Buttons, btn)
					}
					// Drain any grandchildren of the button (none expected in practice).
					drainAdvChildren(r)
					if r.FinishChild() != nil { break }
				}
				row.Cells = append(row.Cells, cell)
			} else {
				// Drain unexpected row children.
				drainAdvChildren(r)
			}
			if r.FinishChild() != nil { break }
		}
		a.TableRows = append(a.TableRows, row)
		return true

	case "Columns":
		for {
			ct, ok := r.NextChild()
			if !ok {
				break
			}
			if ct == "Descriptor" {
				d := readAdvDescriptor(r)
				a.Columns = append(a.Columns, d)
			} else {
				drainAdvChildren(r)
			}
			if r.FinishChild() != nil { break }
		}
		return true

	case "Rows":
		for {
			ct, ok := r.NextChild()
			if !ok {
				break
			}
			if ct == "Descriptor" {
				d := readAdvDescriptor(r)
				a.Rows = append(a.Rows, d)
			} else {
				drainAdvChildren(r)
			}
			if r.FinishChild() != nil { break }
		}
		return true

	case "MatrixCollapseButton", "MatrixSortButton", "Cells", "MatrixRows", "MatrixColumns":
		// These appear at the AdvMatrix level in some FRX variants; drain them.
		drainAdvChildren(r)
		return true
	}
	return false
}

// drainAdvChildren discards all children of the current element (recursive).
func drainAdvChildren(r report.Reader) {
	for {
		_, ok := r.NextChild()
		if !ok {
			break
		}
		drainAdvChildren(r)
		if r.FinishChild() != nil { break }
	}
}

// readAdvDescriptor reads a single Descriptor element and its children.
func readAdvDescriptor(r report.Reader) *AdvMatrixDescriptor {
	d := &AdvMatrixDescriptor{
		Expression:  r.ReadStr("Expression", ""),
		DisplayText: r.ReadStr("DisplayText", ""),
		Sort:        r.ReadStr("Sort", ""),
	}
	for {
		ct, ok := r.NextChild()
		if !ok {
			break
		}
		if ct == "Descriptor" {
			child := readAdvDescriptor(r)
			d.Children = append(d.Children, child)
		} else {
			drainAdvChildren(r)
		}
		if r.FinishChild() != nil { break }
	}
	return d
}
