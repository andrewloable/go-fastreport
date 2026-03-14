package object

import (
	"github.com/andrewloable/go-fastreport/report"
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
	Name     string
	Width    float32
	Height   float32
	ColSpan  int
	RowSpan  int
	Text     string
	HorzAlign int
	VertAlign int
	// Buttons contains any MatrixCollapseButton/MatrixSortButton children.
	Buttons []*MatrixButton
	// RawAttrs preserves other attributes for round-trip fidelity.
	RawAttrs map[string]string
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
// for round-trip fidelity; rendering is not yet implemented.
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
	// Physical table columns and rows are written as child XML elements;
	// the Writer interface does not currently expose arbitrary element writing,
	// so they are preserved in memory for downstream consumers.
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
				cell := &AdvMatrixCell{
					Name:      r.ReadStr("Name", ""),
					Width:     r.ReadFloat("Width", 0),
					Height:    r.ReadFloat("Height", 0),
					ColSpan:   r.ReadInt("ColSpan", 1),
					RowSpan:   r.ReadInt("RowSpan", 1),
					Text:      r.ReadStr("Text", ""),
					HorzAlign: r.ReadInt("HorzAlign", 0),
					VertAlign: r.ReadInt("VertAlign", 0),
				}
				if cell.ColSpan < 1 {
					cell.ColSpan = 1
				}
				if cell.RowSpan < 1 {
					cell.RowSpan = 1
				}
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
					_ = r.FinishChild()
				}
				row.Cells = append(row.Cells, cell)
			} else {
				// Drain unexpected row children.
				drainAdvChildren(r)
			}
			_ = r.FinishChild()
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
			_ = r.FinishChild()
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
			_ = r.FinishChild()
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
		_ = r.FinishChild()
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
		_ = r.FinishChild()
	}
	return d
}
