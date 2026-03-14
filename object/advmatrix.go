package object

import "github.com/andrewloable/go-fastreport/report"

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

// AdvMatrixObject is an advanced pivot-matrix object that generates
// cross-tab reports with collapsible/sortable rows and columns.
//
// It is the Go equivalent of FastReport.AdvMatrix.AdvMatrixObject.
// This stub supports FRX loading (deserialization) only; rendering is
// not yet implemented.
type AdvMatrixObject struct {
	report.ReportComponentBase

	// DataSource is the name of the bound data source.
	DataSource string
	// Columns holds the column descriptors.
	Columns []*AdvMatrixDescriptor
	// Rows holds the row descriptors.
	Rows []*AdvMatrixDescriptor
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
// It reads Columns and Rows descriptor blocks, and silently accepts
// all other child types (TableColumn, TableRow, MatrixCollapseButton, etc.)
// that appear in AdvMatrix FRX files.
func (a *AdvMatrixObject) DeserializeChild(childType string, r report.Reader) bool {
	switch childType {
	case "Columns":
		for {
			ct, ok := r.NextChild()
			if !ok {
				break
			}
			if ct == "Descriptor" {
				d := readAdvDescriptor(r)
				a.Columns = append(a.Columns, d)
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
			}
			_ = r.FinishChild()
		}
		return true
	case "TableColumn", "TableRow", "MatrixCollapseButton", "MatrixSortButton",
		"Cells", "MatrixRows", "MatrixColumns":
		// Drain all grandchildren silently.
		for {
			_, ok := r.NextChild()
			if !ok {
				break
			}
			_ = r.FinishChild()
		}
		return true
	}
	return false
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
			// Drain unknown grandchildren.
			for {
				_, ok2 := r.NextChild()
				if !ok2 {
					break
				}
				_ = r.FinishChild()
			}
		}
		_ = r.FinishChild()
	}
	return d
}
