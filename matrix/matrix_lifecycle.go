package matrix

import (
	"encoding/base64"
	"fmt"

	"github.com/andrewloable/go-fastreport/format"
	"github.com/andrewloable/go-fastreport/object"
	"github.com/andrewloable/go-fastreport/table"
)

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
	// Collect DataColumn image bindings from header template cells.
	// C# ref: PrintColumnHeader/PrintData call templateCell.GetData() which
	// evaluates DataColumn per data row, caching per unique header value.
	m.collectHeaderImages(colVals, rowVals, calc)
}

// collectHeaderImages scans column- and row-header template cells for
// PictureObjects with DataColumn bindings. For each unique header key seen
// during data iteration, it evaluates the DataColumn expression once and
// caches the decoded image bytes.
// C# ref: PrintColumnHeader → templateCell.GetData() → PictureObject.GetData()
//
//	evaluates DataColumn = Report.GetColumnValue(col) → image bytes.
func (m *MatrixObject) collectHeaderImages(colVals, rowVals []any, calc func(string) any) {
	if calc == nil {
		return
	}
	if m.colHeaderImgByVal == nil {
		m.colHeaderImgByVal = make(map[string][]byte)
	}
	if m.rowHeaderImgByVal == nil {
		m.rowHeaderImgByVal = make(map[string][]byte)
	}

	titleOffset := 0
	if m.ShowTitle {
		titleOffset = 1
	}
	nColDims := len(m.Data.Columns)
	if nColDims < 1 {
		nColDims = 1
	}
	nRowHdrs := len(m.Data.Rows)
	if nRowHdrs < 1 {
		nRowHdrs = 1
	}

	templateRows := m.TableBase.Rows()

	// Column header template: row at titleOffset, cells at columns nRowHdrs..nRowHdrs+ci
	if titleOffset < len(templateRows) && len(colVals) > 0 {
		row := templateRows[titleOffset]
		if row != nil {
			cells := row.Cells()
			for ci := range m.Data.Columns {
				cellIdx := nRowHdrs + ci
				if cellIdx >= len(cells) {
					break
				}
				cell := cells[cellIdx]
				if cell == nil {
					continue
				}
				for _, obj := range cell.Objects() {
					pic, ok := obj.(*object.PictureObject)
					if !ok || pic.DataColumn() == "" {
						continue
					}
					if ci >= len(colVals) {
						continue
					}
					key := fmt.Sprintf("%v", colVals[ci])
					if _, exists := m.colHeaderImgByVal[key]; exists {
						continue // already cached for this header value
					}
					if data := m.evalPicDataColumn(calc, pic.DataColumn()); len(data) > 0 {
						m.colHeaderImgByVal[key] = data
					}
				}
			}
		}
	}

	// Row header template: row at titleOffset+nColDims, cells at columns 0..nRowHdrs-1
	rowHdrTemplate := titleOffset + nColDims
	if rowHdrTemplate < len(templateRows) && len(rowVals) > 0 {
		row := templateRows[rowHdrTemplate]
		if row != nil {
			cells := row.Cells()
			for ri := range m.Data.Rows {
				if ri >= len(cells) {
					break
				}
				cell := cells[ri]
				if cell == nil {
					continue
				}
				for _, obj := range cell.Objects() {
					pic, ok := obj.(*object.PictureObject)
					if !ok || pic.DataColumn() == "" {
						continue
					}
					if ri >= len(rowVals) {
						continue
					}
					key := fmt.Sprintf("%v", rowVals[ri])
					if _, exists := m.rowHeaderImgByVal[key]; exists {
						continue // already cached
					}
					if data := m.evalPicDataColumn(calc, pic.DataColumn()); len(data) > 0 {
						m.rowHeaderImgByVal[key] = data
					}
				}
			}
		}
	}
}

// evalPicDataColumn evaluates a DataColumn expression via the calc function and
// decodes the result as a base64-encoded image. Returns nil if not decodable.
// C# ref: PictureObject.GetData() → Report.GetColumnValueNullable(DataColumn).
// applyHeaderImages replaces shared PictureObject references in a result header
// cell with clones that carry the cached image bytes for the given header key.
// C# ref: RunTimeAssign (called from PrintColumnHeader / PrintRowHeader) clones
// the template cell including PictureObject children; GetData() then loads the
// image from the DataColumn.
func (m *MatrixObject) applyHeaderImages(cell *table.TableCell, imgByVal map[string][]byte, key string) {
	data, ok := imgByVal[key]
	if !ok || len(data) == 0 {
		return
	}
	objs := cell.Objects()
	for i, obj := range objs {
		pic, isPic := obj.(*object.PictureObject)
		if !isPic || pic.DataColumn() == "" {
			continue
		}
		// Clone so each result cell gets its own independent PictureObject.
		newPic := object.NewPictureObject()
		newPic.SetName(pic.Name())
		newPic.SetLeft(pic.Left())
		newPic.SetTop(pic.Top())
		newPic.SetWidth(pic.Width())
		newPic.SetHeight(pic.Height())
		newPic.SetDataColumn(pic.DataColumn())
		newPic.SetImageData(data)
		cell.ReplaceObject(i, newPic)
	}
}

func (m *MatrixObject) evalPicDataColumn(calc func(string) any, colName string) []byte {
	val := calc("[" + colName + "]")
	if val == nil {
		return nil
	}
	// Mirror the pattern in engine/objects.go lines 3015-3027: handle both
	// pre-decoded []byte (returned by some datasources) and base64 strings.
	switch d := val.(type) {
	case []byte:
		if len(d) > 0 {
			return d
		}
	case string:
		if len(d) == 0 {
			return nil
		}
		// Try standard base64 then raw (no padding).
		decoded, err := base64.StdEncoding.DecodeString(d)
		if err != nil {
			decoded, err = base64.RawStdEncoding.DecodeString(d)
		}
		if err == nil && len(decoded) > 0 {
			return decoded
		}
	}
	return nil
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
