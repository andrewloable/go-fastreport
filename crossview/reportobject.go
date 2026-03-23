package crossview

import (
	"github.com/andrewloable/go-fastreport/report"
)

// CrossViewReportObject wraps CrossViewObject and implements report.Base so that
// CrossView objects can be registered in the serial registry and loaded from FRX.
//
// C# ref: FastReport.CrossView.CrossViewObject (CrossViewObject.cs) which inherits
// from TableBase → BreakableComponent → ReportComponentBase.
type CrossViewReportObject struct {
	report.ReportComponentBase

	// CrossView is the embedded computation/data object.
	CrossView CrossViewObject

	// saveVisible holds the pre-print visible state for SaveState/RestoreState.
	// Mirrors C# CrossViewObject.saveVisible (CrossViewObject.cs line 26).
	saveVisible bool
}

// NewCrossViewReportObject creates a CrossViewReportObject with defaults.
// Mirrors CrossViewObject constructor (CrossViewObject.cs lines 495–509).
func NewCrossViewReportObject() *CrossViewReportObject {
	ro := &CrossViewReportObject{}
	ro.SetBaseName("CrossViewObject")
	ro.ReportComponentBase.SetVisible(true)
	ro.CrossView.ShowTitle = true
	ro.CrossView.ShowXAxisFieldsCaption = true
	ro.CrossView.ShowYAxisFieldsCaption = true
	helper := NewCrossViewHelper(&ro.CrossView)
	ro.CrossView.helper = helper
	return ro
}

// TypeName returns the FRX type name.
func (ro *CrossViewReportObject) TypeName() string { return "CrossViewObject" }

// BaseName returns the base name prefix used for auto-naming.
func (ro *CrossViewReportObject) BaseName() string { return "CrossViewObject" }

// Serialize writes the CrossViewReportObject to w.
// Mirrors C# CrossViewObject.Serialize (CrossViewObject.cs lines 356–392).
func (ro *CrossViewReportObject) Serialize(w report.Writer) error {
	// Write CrossViewData child collections as FRX child elements.
	// C# writes Data.Columns as "<CrossViewColumns>", Data.Rows as "<CrossViewRows>",
	// Data.Cells as "<CrossViewCells>". Each collection serializes its items as
	// "<Header ...>" or "<Cell ...>" child elements.
	// Mirrors CrossViewObject.Serialize lines 360-363 (C# writer.Write calls).
	colHeader := NewCrossViewHeader("CrossViewColumns")
	colHeader.Items = ro.CrossView.Data.Columns
	if err := w.WriteObjectNamed("CrossViewColumns", colHeader); err != nil {
		return err
	}
	rowHeader := NewCrossViewHeader("CrossViewRows")
	rowHeader.Items = ro.CrossView.Data.Rows
	if err := w.WriteObjectNamed("CrossViewRows", rowHeader); err != nil {
		return err
	}
	cells := NewCrossViewCells("CrossViewCells")
	cells.Items = ro.CrossView.Data.Cells
	if err := w.WriteObjectNamed("CrossViewCells", cells); err != nil {
		return err
	}

	// Write base component properties (bounds, fill, border, etc.).
	if err := ro.ReportComponentBase.Serialize(w); err != nil {
		return err
	}

	// Write CrossView-specific properties.
	if s := ro.CrossView.Data.ColumnDescriptorsIndexesStr(); s != "" {
		w.WriteStr("ColumnDescriptorsIndexes", s)
	}
	if s := ro.CrossView.Data.RowDescriptorsIndexesStr(); s != "" {
		w.WriteStr("RowDescriptorsIndexes", s)
	}
	if s := ro.CrossView.Data.ColumnTerminalIndexesStr(); s != "" {
		w.WriteStr("ColumnTerminalIndexes", s)
	}
	if s := ro.CrossView.Data.RowTerminalIndexesStr(); s != "" {
		w.WriteStr("RowTerminalIndexes", s)
	}
	if ro.CrossView.ShowTitle {
		w.WriteBool("ShowTitle", true)
	}
	if ro.CrossView.ShowXAxisFieldsCaption {
		w.WriteBool("ShowXAxisFieldsCaption", true)
	}
	if ro.CrossView.ShowYAxisFieldsCaption {
		w.WriteBool("ShowYAxisFieldsCaption", true)
	}
	if ro.CrossView.Style != "" {
		w.WriteStr("Style", ro.CrossView.Style)
	}
	if ro.CrossView.ModifyResultEvent != "" {
		w.WriteStr("ModifyResultEvent", ro.CrossView.ModifyResultEvent)
	}
	return nil
}

// Deserialize reads the CrossViewReportObject from r.
func (ro *CrossViewReportObject) Deserialize(r report.Reader) error {
	if err := ro.ReportComponentBase.Deserialize(r); err != nil {
		return err
	}
	ro.CrossView.Data.SetColumnDescriptorsIndexesStr(r.ReadStr("ColumnDescriptorsIndexes", ""))
	ro.CrossView.Data.SetRowDescriptorsIndexesStr(r.ReadStr("RowDescriptorsIndexes", ""))
	ro.CrossView.Data.SetColumnTerminalIndexesStr(r.ReadStr("ColumnTerminalIndexes", ""))
	ro.CrossView.Data.SetRowTerminalIndexesStr(r.ReadStr("RowTerminalIndexes", ""))
	ro.CrossView.ShowTitle = r.ReadBool("ShowTitle", false)
	ro.CrossView.ShowXAxisFieldsCaption = r.ReadBool("ShowXAxisFieldsCaption", true)
	ro.CrossView.ShowYAxisFieldsCaption = r.ReadBool("ShowYAxisFieldsCaption", true)
	ro.CrossView.Style = r.ReadStr("Style", "")
	ro.CrossView.ModifyResultEvent = r.ReadStr("ModifyResultEvent", "")
	return nil
}

// DeserializeChild handles CrossView-specific child elements (CrossViewColumns,
// CrossViewRows, CrossViewCells).
// Mirrors C# CrossViewObject.DeserializeSubItems (CrossViewObject.cs lines 327–338).
func (ro *CrossViewReportObject) DeserializeChild(childType string, r report.Reader) bool {
	switch childType {
	case "CrossViewColumns":
		// Read the columns header collection (each child is a "Header" element).
		header := NewCrossViewHeader("CrossViewColumns")
		_ = header.Deserialize(r)
		ro.CrossView.Data.Columns = header.Items
		return true
	case "CrossViewRows":
		header := NewCrossViewHeader("CrossViewRows")
		_ = header.Deserialize(r)
		ro.CrossView.Data.Rows = header.Items
		return true
	case "CrossViewCells":
		cells := NewCrossViewCells("CrossViewCells")
		_ = cells.Deserialize(r)
		ro.CrossView.Data.Cells = cells.Items
		return true
	}
	return false
}

// SaveState saves the visible state before engine rendering.
// Mirrors C# CrossViewObject.SaveState (CrossViewObject.cs lines 423–441).
func (ro *CrossViewReportObject) SaveState() {
	ro.saveVisible = ro.ReportComponentBase.Visible()
	ro.ReportComponentBase.SaveState()
}

// RestoreState restores the visible state after engine rendering.
// Mirrors C# CrossViewObject.RestoreState (CrossViewObject.cs lines 466–477).
func (ro *CrossViewReportObject) RestoreState() {
	ro.ReportComponentBase.RestoreState()
}

// GetData triggers data loading from the cube source.
// Mirrors C# CrossViewObject.GetData + GetDataShared (CrossViewObject.cs lines 444–463).
func (ro *CrossViewReportObject) GetData() {
	if ro.CrossView.Data.SourceAssigned() {
		if ro.CrossView.helper == nil {
			ro.CrossView.helper = NewCrossViewHelper(&ro.CrossView)
		}
		ro.CrossView.helper.StartPrint()
		ro.CrossView.helper.AddData()
		ro.CrossView.helper.FinishPrint()
	}
}

// OnModifyResult fires the ModifyResultHandler callback.
// Mirrors C# CrossViewObject.OnModifyResult (CrossViewObject.cs lines 483–488).
func (ro *CrossViewReportObject) OnModifyResult() {
	if ro.CrossView.ModifyResultHandler != nil {
		ro.CrossView.ModifyResultHandler(&ro.CrossView)
	}
}
