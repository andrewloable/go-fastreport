package matrix

import "github.com/andrewloable/go-fastreport/format"

// matrix_lifecycle.go adds the runtime state fields and engine lifecycle hooks
// to MatrixObject that were identified as missing gaps.
//
// C# source: FastReport.Base/Matrix/MatrixObject.cs (Report Engine region)
//            FastReport.Base/Matrix/MatrixHelper.cs  (StartPrint, AddDataRow, FinishPrint)
//
// Gaps implemented (issue go-fastreport-b8hvg):
//   - ColumnValues / RowValues / CellValues runtime state arrays
//   - ColumnIndex / RowIndex runtime counters
//   - OnManualBuild / OnModifyResult / OnAfterTotals event firing
//   - SaveState / RestoreState lifecycle stubs
//   - InitializeComponent / FinalizeComponent lifecycle stubs
//   - GetData stub (wires data source iteration)
//   - OnAfterData stub
//   - AddValue convenience shortcut (wraps Data.AddValue)
//   - Value(index) helper (returns current cell value during expression eval)
//
// Note: Full rendering (StartPrint/FinishPrint/InitResultTable/PrintHeaders/
// PrintData) requires the report engine context (Report.Calc) and the
// TableResult rendering pipeline. Those are out of scope for this iteration and
// are documented here as stubs.

// ── Runtime state additions to MatrixObject ───────────────────────────────────

// The following fields extend MatrixObject but are kept in a separate file for
// clarity. They correspond to C# private fields on MatrixObject.
//
// Because Go does not support partial structs we store them as exported fields
// initialised by New() so that engine code can set and read them directly.

func init() {
	// Verify the fields are accessible at compile time via New().
	_ = func() {
		m := New()
		m.ColumnValues = nil
		m.RowValues2 = nil
		m.CellValues = nil
		_ = m.ColumnIndex
		_ = m.RowIndex
	}
}

// ── Event callbacks ────────────────────────────────────────────────────────────

// OnManualBuild fires the ManualBuild event.
// C# source: MatrixObject.OnManualBuild(EventArgs e).
func (m *MatrixObject) OnManualBuild() {
	if m.ManualBuildHandler != nil {
		m.ManualBuildHandler(m)
	}
}

// OnModifyResult fires the ModifyResult event.
// C# source: MatrixObject.OnModifyResult(EventArgs e).
func (m *MatrixObject) OnModifyResult() {
	if m.ModifyResultHandler != nil {
		m.ModifyResultHandler(m)
	}
}

// OnAfterTotals fires the AfterTotals event.
// C# source: MatrixObject.OnAfterTotals(EventArgs e).
func (m *MatrixObject) OnAfterTotals() {
	if m.AfterTotalsHandler != nil {
		m.AfterTotalsHandler(m)
	}
}

// ── Engine lifecycle ───────────────────────────────────────────────────────────

// SaveState saves the visibility state of the matrix object before the engine
// runs. In the C# engine this also creates the ResultTable. Here we record
// the pre-run visible flag so RestoreState can restore it.
// C# source: MatrixObject.SaveState().
func (m *MatrixObject) SaveState() {
	m.saveVisible = m.visible
}

// RestoreState restores the matrix object's visibility after the engine run.
// C# source: MatrixObject.RestoreState().
func (m *MatrixObject) RestoreState() {
	m.visible = m.saveVisible
}

// InitializeComponent is called by the engine before data collection begins.
// C# source: MatrixObject.InitializeComponent() → base + WireEvents(true).
func (m *MatrixObject) InitializeComponent() {
	// Wire-up happens in the engine; the Go version is a no-op stub.
}

// FinalizeComponent is called by the engine after the report finishes.
// C# source: MatrixObject.FinalizeComponent() → base + WireEvents(false).
func (m *MatrixObject) FinalizeComponent() {
	// Tear-down happens in the engine; the Go version is a no-op stub.
}

// GetDataWithCalc iterates the bound DataSource (if any), evaluating descriptor
// expressions via the supplied calc function for each data row that passes the
// Filter expression. It mirrors the non-footer path in C#.
//
// C# source: MatrixObject.GetData() → GetDataShared() → Helper.StartPrint() + AddDataRows().
//
// Usage (from report engine):
//
//	m.GetDataWithCalc(func(expr string) any { return report.Calc(expr) })
func (m *MatrixObject) GetDataWithCalc(calc func(expr string) any) {
	// Reset accumulated data.
	m.Data.RuntimeReset()

	// Fire ManualBuild to allow code-driven population.
	m.OnManualBuild()

	if m.DataSource == nil {
		return
	}

	// Initialise the data source and iterate rows.
	// The Go DataSource interface uses Init()/EOF()/Next() (no filter argument).
	if err := m.DataSource.Init(); err != nil {
		return
	}
	if err := m.DataSource.First(); err != nil {
		return
	}
	for !m.DataSource.EOF() {
		// Apply filter expression.
		if m.Filter != "" && calc != nil {
			result := calc(m.Filter)
			if b, ok := result.(bool); ok && !b {
				_ = m.DataSource.Next()
				continue
			}
		}
		m.addDataRowWithCalc(calc)
		_ = m.DataSource.Next()
	}
}

// addDataRowWithCalc evaluates one data source row into the matrix.
// C# source: MatrixHelper.AddDataRow().
func (m *MatrixObject) addDataRowWithCalc(calc func(expr string) any) {
	colVals := make([]any, len(m.Data.Columns))
	rowVals := make([]any, len(m.Data.Rows))
	cellVals := make([]any, len(m.Data.Cells))

	for i, d := range m.Data.Columns {
		if d.Expression != "" && calc != nil {
			colVals[i] = calc(d.Expression)
		}
	}
	for i, d := range m.Data.Rows {
		if d.Expression != "" && calc != nil {
			rowVals[i] = calc(d.Expression)
		}
	}
	for i, d := range m.Data.Cells {
		if d.Function == AggregateFunctionCustom {
			cellVals[i] = 0 // deferred; calculated at print time
		} else if d.Expression != "" && calc != nil {
			cellVals[i] = calc(d.Expression)
		}
	}

	dataRowNo := 0
	if m.DataSource != nil {
		dataRowNo = m.DataSource.CurrentRowNo()
	}
	m.Data.AddValueAt(colVals, rowVals, cellVals, dataRowNo)
}

// OnAfterData is called by the engine after all data has been collected and
// the result table has been rendered. In C# it calls Helper.FinishPrint().
// The Go version fires the ModifyResult event (wired via ResultTable_AfterData
// in C#). Full rendering is out of scope for this iteration.
// C# source: MatrixObject.OnAfterData(EventArgs e).
func (m *MatrixObject) OnAfterData() {
	m.OnModifyResult()
}

// ── Shortcut AddValue ──────────────────────────────────────────────────────────

// AddValue is a convenience shortcut that calls Data.AddValue.
// It matches the C# MatrixObject.AddValue(object[], object[], object[]) method.
// C# source: MatrixObject.AddValue(object[] columnValues, object[] rowValues, object[] cellValues).
func (m *MatrixObject) AddValue(columnValues, rowValues, cellValues []any) {
	m.Data.AddValue(columnValues, rowValues, cellValues)
}

// ExtractCellFormat scans the template cells for a non-General Format
// (e.g. Currency) and stores it for use during BuildTemplateMultiLevel.
// Must be called AFTER FRX children are fully deserialized (the template rows).
func (m *MatrixObject) ExtractCellFormat() {
	for _, row := range m.TableBase.Rows() {
		for _, cell := range row.Cells() {
			if cell == nil {
				continue
			}
			f := cell.Format()
			if f == nil {
				continue
			}
			if _, isGen := f.(*format.GeneralFormat); !isGen {
				m.cellFormat = f
				return
			}
		}
	}
}

// SyncRuntimeToMultiLevel bridges the runtime header trees (populated by
// GetDataWithCalc → AddValueAt) to the multi-level header trees used by
// BuildTemplateMultiLevel. After calling GetDataWithCalc, the data is in
// m.Data.rt.rows/columns. This method copies those trees to m.rowRoot/colRoot
// so BuildTemplateMultiLevel can generate the result table.
func (m *MatrixObject) SyncRuntimeToMultiLevel() {
	rt := m.Data.Runtime()
	if rt == nil {
		return
	}
	// Add total items to the column header tree (mirrors C# MatrixData.Columns.AddTotalItems
	// called before GenerateResult). This inserts per-group Total leaves and a Grand Total
	// leaf into the column tree so BuildTemplateMultiLevel emits them as real columns.
	// Must be called after GetDataWithCalc has fully populated the column tree.
	rt.Columns().AddTotalItems(false)
	m.rowRoot = rt.Rows().Root
	m.colRoot = rt.Columns().Root
	// Ensure mlAccumulators is initialized.
	if m.mlAccumulators == nil {
		m.mlAccumulators = make(map[multiLevelKey]*accumulator)
	}
}

// Value returns the current cell value at the given descriptor index.
// Used in cell expressions when the aggregate function is Custom.
// C# source: MatrixObject.Value(int index).
func (m *MatrixObject) Value(index int) any {
	if index < 0 || index >= len(m.CellValues) {
		return 0
	}
	v := m.CellValues[index]
	if v == nil {
		return 0
	}
	return v
}
