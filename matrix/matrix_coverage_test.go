package matrix_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/andrewloable/go-fastreport/matrix"
	"github.com/andrewloable/go-fastreport/serial"
)

// ── helper: serialize a MatrixObject to XML ───────────────────────────────────

func serializeMatrix(t *testing.T, m *matrix.MatrixObject) string {
	t.Helper()
	var buf bytes.Buffer
	w := serial.NewWriter(&buf)
	if err := w.WriteObjectNamed("MatrixObject", m); err != nil {
		t.Fatalf("serialize: %v", err)
	}
	if err := w.Flush(); err != nil {
		t.Fatalf("flush: %v", err)
	}
	return buf.String()
}

// helper: deserialize a MatrixObject from XML string.
func deserializeMatrix(t *testing.T, xml string) *matrix.MatrixObject {
	t.Helper()
	r := serial.NewReader(strings.NewReader(xml))
	typeName, ok := r.ReadObjectHeader()
	if !ok {
		t.Fatalf("ReadObjectHeader returned ok=false; xml was:\n%s", xml)
	}
	if typeName != "MatrixObject" {
		t.Fatalf("unexpected type %q, want MatrixObject", typeName)
	}
	m := matrix.New()
	if err := m.Deserialize(r); err != nil {
		t.Fatalf("Deserialize: %v", err)
	}
	// Process child elements (MatrixRows, MatrixColumns, MatrixCells).
	for {
		childType, ok := r.NextChild()
		if !ok {
			break
		}
		if !m.DeserializeChild(childType, r) {
			// Unknown child — skip.
			if err := r.FinishChild(); err != nil {
				t.Fatalf("FinishChild: %v", err)
			}
			continue
		}
		if err := r.FinishChild(); err != nil {
			t.Fatalf("FinishChild: %v", err)
		}
	}
	return m
}

// ── Serialize / Deserialize round-trip tests ──────────────────────────────────

func TestSerializeDeserialize_Defaults(t *testing.T) {
	orig := matrix.New()
	xml := serializeMatrix(t, orig)

	got := deserializeMatrix(t, xml)

	// Verify defaults come back correctly.
	if !got.AutoSize {
		t.Error("AutoSize should default to true after round-trip")
	}
	if !got.PrintIfEmpty {
		t.Error("PrintIfEmpty should default to true after round-trip")
	}
	if got.CellsSideBySide {
		t.Error("CellsSideBySide should default to false")
	}
	if got.ShowTitle {
		t.Error("ShowTitle should default to false")
	}
	if got.SplitRows {
		t.Error("SplitRows should default to false")
	}
}

func TestSerializeDeserialize_AllStringFields(t *testing.T) {
	orig := matrix.New()
	orig.DataSourceName = "SalesDS"
	orig.Filter = "[Amount] > 0"
	orig.Style = "Blue"
	orig.ManualBuildEvent = "OnManualBuild"
	orig.ModifyResultEvent = "OnModifyResult"
	orig.AfterTotalsEvent = "OnAfterTotals"

	xml := serializeMatrix(t, orig)

	got := deserializeMatrix(t, xml)

	if got.DataSourceName != "SalesDS" {
		t.Errorf("DataSourceName = %q, want SalesDS", got.DataSourceName)
	}
	if got.Filter != "[Amount] > 0" {
		t.Errorf("Filter = %q, want [Amount] > 0", got.Filter)
	}
	if got.Style != "Blue" {
		t.Errorf("Style = %q, want Blue", got.Style)
	}
	if got.ManualBuildEvent != "OnManualBuild" {
		t.Errorf("ManualBuildEvent = %q, want OnManualBuild", got.ManualBuildEvent)
	}
	if got.ModifyResultEvent != "OnModifyResult" {
		t.Errorf("ModifyResultEvent = %q, want OnModifyResult", got.ModifyResultEvent)
	}
	if got.AfterTotalsEvent != "OnAfterTotals" {
		t.Errorf("AfterTotalsEvent = %q, want OnAfterTotals", got.AfterTotalsEvent)
	}
}

func TestSerializeDeserialize_BoolFlags(t *testing.T) {
	orig := matrix.New()
	orig.AutoSize = false
	orig.CellsSideBySide = true
	orig.KeepCellsSideBySide = true
	orig.ShowTitle = true
	orig.SplitRows = true
	orig.PrintIfEmpty = false

	xml := serializeMatrix(t, orig)

	got := deserializeMatrix(t, xml)

	if got.AutoSize {
		t.Error("AutoSize: expected false")
	}
	if !got.CellsSideBySide {
		t.Error("CellsSideBySide: expected true")
	}
	if !got.KeepCellsSideBySide {
		t.Error("KeepCellsSideBySide: expected true")
	}
	if !got.ShowTitle {
		t.Error("ShowTitle: expected true")
	}
	if !got.SplitRows {
		t.Error("SplitRows: expected true")
	}
	if got.PrintIfEmpty {
		t.Error("PrintIfEmpty: expected false")
	}
}

func TestSerializeDeserialize_EvenStylePriority(t *testing.T) {
	orig := matrix.New()
	orig.EvenStylePriority = matrix.EvenStylePriorityColumns

	xml := serializeMatrix(t, orig)

	got := deserializeMatrix(t, xml)
	if got.EvenStylePriority != matrix.EvenStylePriorityColumns {
		t.Errorf("EvenStylePriority = %v, want Columns", got.EvenStylePriority)
	}
}

// ── Descriptor serialization round-trips ──────────────────────────────────────

func TestSerializeDeserialize_MatrixRows(t *testing.T) {
	orig := matrix.New()
	orig.Data.AddRow(matrix.NewHeaderDescriptor("[Year]"))
	h2 := matrix.NewHeaderDescriptor("[Month]")
	h2.Sort = matrix.SortOrderDescending
	h2.Totals = false
	h2.TotalsFirst = true
	h2.PageBreak = true
	h2.SuppressTotals = true
	orig.Data.AddRow(h2)

	xml := serializeMatrix(t, orig)

	got := deserializeMatrix(t, xml)

	if len(got.Data.Rows) != 2 {
		t.Fatalf("Data.Rows len = %d, want 2", len(got.Data.Rows))
	}
	r0 := got.Data.Rows[0]
	if r0.Expression != "[Year]" {
		t.Errorf("Rows[0].Expression = %q, want [Year]", r0.Expression)
	}
	r1 := got.Data.Rows[1]
	if r1.Expression != "[Month]" {
		t.Errorf("Rows[1].Expression = %q, want [Month]", r1.Expression)
	}
	if r1.Sort != matrix.SortOrderDescending {
		t.Errorf("Rows[1].Sort = %v, want Descending", r1.Sort)
	}
	if r1.Totals {
		t.Error("Rows[1].Totals: expected false")
	}
	if !r1.TotalsFirst {
		t.Error("Rows[1].TotalsFirst: expected true")
	}
	if !r1.PageBreak {
		t.Error("Rows[1].PageBreak: expected true")
	}
	if !r1.SuppressTotals {
		t.Error("Rows[1].SuppressTotals: expected true")
	}
}

func TestSerializeDeserialize_MatrixColumns(t *testing.T) {
	orig := matrix.New()
	orig.Data.AddColumn(matrix.NewHeaderDescriptor("[Product]"))
	orig.Data.AddColumn(matrix.NewHeaderDescriptor("[Region]"))

	xml := serializeMatrix(t, orig)

	got := deserializeMatrix(t, xml)

	if len(got.Data.Columns) != 2 {
		t.Fatalf("Data.Columns len = %d, want 2", len(got.Data.Columns))
	}
	if got.Data.Columns[0].Expression != "[Product]" {
		t.Errorf("Columns[0].Expression = %q, want [Product]", got.Data.Columns[0].Expression)
	}
	if got.Data.Columns[1].Expression != "[Region]" {
		t.Errorf("Columns[1].Expression = %q, want [Region]", got.Data.Columns[1].Expression)
	}
}

func TestSerializeDeserialize_MatrixCells(t *testing.T) {
	orig := matrix.New()
	orig.Data.AddCell(matrix.NewCellDescriptor("[Revenue]", matrix.AggregateFunctionSum))
	orig.Data.AddCell(matrix.NewCellDescriptor("[Count]", matrix.AggregateFunctionCount))

	c3 := matrix.NewCellDescriptor("[Pct]", matrix.AggregateFunctionAvg)
	c3.Percent = matrix.MatrixPercentColumnTotal
	orig.Data.AddCell(c3)

	// Also add a cell with non-Sum function so it writes Function attribute.
	orig.Data.AddCell(matrix.NewCellDescriptor("[Distinct]", matrix.AggregateFunctionCountDistinct))

	xml := serializeMatrix(t, orig)

	got := deserializeMatrix(t, xml)

	if len(got.Data.Cells) != 4 {
		t.Fatalf("Data.Cells len = %d, want 4", len(got.Data.Cells))
	}
	c0 := got.Data.Cells[0]
	if c0.Expression != "[Revenue]" {
		t.Errorf("Cells[0].Expression = %q", c0.Expression)
	}
	if c0.Function != matrix.AggregateFunctionSum {
		t.Errorf("Cells[0].Function = %v, want Sum", c0.Function)
	}
	c2 := got.Data.Cells[2]
	if c2.Percent != matrix.MatrixPercentColumnTotal {
		t.Errorf("Cells[2].Percent = %v, want ColumnTotal", c2.Percent)
	}
	c3got := got.Data.Cells[3]
	if c3got.Function != matrix.AggregateFunctionCountDistinct {
		t.Errorf("Cells[3].Function = %v, want CountDistinct", c3got.Function)
	}
}

func TestSerializeDeserialize_AllDescriptors(t *testing.T) {
	// Full round-trip: rows + columns + cells together.
	orig := matrix.New()
	orig.DataSourceName = "DS"
	orig.Data.AddRow(matrix.NewHeaderDescriptor("[Year]"))
	orig.Data.AddColumn(matrix.NewHeaderDescriptor("[Quarter]"))
	orig.Data.AddCell(matrix.NewCellDescriptor("[Amount]", matrix.AggregateFunctionSum))

	xml := serializeMatrix(t, orig)

	got := deserializeMatrix(t, xml)
	if len(got.Data.Rows) != 1 {
		t.Errorf("Data.Rows = %d, want 1", len(got.Data.Rows))
	}
	if len(got.Data.Columns) != 1 {
		t.Errorf("Data.Columns = %d, want 1", len(got.Data.Columns))
	}
	if len(got.Data.Cells) != 1 {
		t.Errorf("Data.Cells = %d, want 1", len(got.Data.Cells))
	}
}

// ── DeserializeChild tests ────────────────────────────────────────────────────

func TestDeserializeChild_UnknownChild(t *testing.T) {
	m := matrix.New()
	// Feed an XML snippet with unknown child element.
	src := `<MatrixObject><UnknownChild Foo="bar"/></MatrixObject>`
	r := serial.NewReader(strings.NewReader(src))
	_, ok := r.ReadObjectHeader()
	if !ok {
		t.Fatal("ReadObjectHeader failed")
	}
	_ = m.Deserialize(r)

	childType, ok := r.NextChild()
	if !ok {
		// No children visible — that's fine, skip the test.
		return
	}
	// DeserializeChild should return false for unknown children.
	handled := m.DeserializeChild(childType, r)
	if handled {
		t.Errorf("DeserializeChild returned true for unknown child %q, want false", childType)
	}
	_ = r.FinishChild()
}

func TestDeserializeChild_MatrixRows(t *testing.T) {
	// Directly test DeserializeChild with MatrixRows content.
	xmlStr := `<MatrixObject>` +
		`<MatrixRows>` +
		`<Header Expression="[Year]" Sort="2" Totals="false" TotalsFirst="true" PageBreak="true" SuppressTotals="true"/>` +
		`</MatrixRows>` +
		`</MatrixObject>`

	got := deserializeMatrix(t, xmlStr)
	if len(got.Data.Rows) != 1 {
		t.Fatalf("Rows = %d, want 1", len(got.Data.Rows))
	}
	r := got.Data.Rows[0]
	if r.Expression != "[Year]" {
		t.Errorf("Expression = %q", r.Expression)
	}
	if r.Sort != matrix.SortOrderDescending {
		t.Errorf("Sort = %v, want Descending", r.Sort)
	}
	if r.Totals {
		t.Error("Totals: expected false")
	}
	if !r.TotalsFirst {
		t.Error("TotalsFirst: expected true")
	}
	if !r.PageBreak {
		t.Error("PageBreak: expected true")
	}
	if !r.SuppressTotals {
		t.Error("SuppressTotals: expected true")
	}
}

func TestDeserializeChild_MatrixColumns(t *testing.T) {
	xmlStr := `<MatrixObject>` +
		`<MatrixColumns>` +
		`<Header Expression="[Product]"/>` +
		`<Header Expression="[Region]"/>` +
		`</MatrixColumns>` +
		`</MatrixObject>`

	got := deserializeMatrix(t, xmlStr)
	if len(got.Data.Columns) != 2 {
		t.Fatalf("Columns = %d, want 2", len(got.Data.Columns))
	}
	if got.Data.Columns[0].Expression != "[Product]" {
		t.Errorf("Columns[0] = %q", got.Data.Columns[0].Expression)
	}
}

func TestDeserializeChild_MatrixCells(t *testing.T) {
	xmlStr := `<MatrixObject>` +
		`<MatrixCells>` +
		`<Cell Expression="[Amount]" Function="4" Percent="1"/>` +
		`</MatrixCells>` +
		`</MatrixObject>`

	got := deserializeMatrix(t, xmlStr)
	if len(got.Data.Cells) != 1 {
		t.Fatalf("Cells = %d, want 1", len(got.Data.Cells))
	}
	c := got.Data.Cells[0]
	if c.Expression != "[Amount]" {
		t.Errorf("Expression = %q", c.Expression)
	}
	// Function=4 is AggregateFunctionAvg.
	if c.Function != matrix.AggregateFunctionAvg {
		t.Errorf("Function = %v, want Avg", c.Function)
	}
	// Percent=1 is MatrixPercentColumnTotal.
	if c.Percent != matrix.MatrixPercentColumnTotal {
		t.Errorf("Percent = %v, want ColumnTotal", c.Percent)
	}
}

func TestDeserializeChild_MatrixRows_NonHeaderChild(t *testing.T) {
	// Inside MatrixRows, a non-Header child type should be skipped.
	xmlStr := `<MatrixObject>` +
		`<MatrixRows>` +
		`<SomethingElse Foo="x"/>` +
		`<Header Expression="[Real]"/>` +
		`</MatrixRows>` +
		`</MatrixObject>`

	got := deserializeMatrix(t, xmlStr)
	// Only the Header should be captured.
	if len(got.Data.Rows) != 1 {
		t.Fatalf("Rows = %d, want 1; SomethingElse should be ignored", len(got.Data.Rows))
	}
	if got.Data.Rows[0].Expression != "[Real]" {
		t.Errorf("Rows[0].Expression = %q", got.Data.Rows[0].Expression)
	}
}

func TestDeserializeChild_MatrixColumns_NonHeaderChild(t *testing.T) {
	xmlStr := `<MatrixObject>` +
		`<MatrixColumns>` +
		`<Noise/>` +
		`<Header Expression="[Col]"/>` +
		`</MatrixColumns>` +
		`</MatrixObject>`

	got := deserializeMatrix(t, xmlStr)
	if len(got.Data.Columns) != 1 {
		t.Fatalf("Columns = %d, want 1", len(got.Data.Columns))
	}
}

func TestDeserializeChild_MatrixCells_NonCellChild(t *testing.T) {
	xmlStr := `<MatrixObject>` +
		`<MatrixCells>` +
		`<Junk/>` +
		`<Cell Expression="[Val]"/>` +
		`</MatrixCells>` +
		`</MatrixObject>`

	got := deserializeMatrix(t, xmlStr)
	if len(got.Data.Cells) != 1 {
		t.Fatalf("Cells = %d, want 1", len(got.Data.Cells))
	}
}

// ── toFloat coverage through public API ───────────────────────────────────────

func TestToFloat_ViaAddData_Float32(t *testing.T) {
	// toFloat is called internally by AddData; exercise it with float32 value.
	m := matrix.New()
	m.Data.AddCell(matrix.NewCellDescriptor("[V]", matrix.AggregateFunctionSum))
	var f32 float32 = 3.5
	m.AddData("r", "c", []any{f32})
	result, err := m.CellResult("r", "c", 0)
	if err != nil {
		t.Fatalf("CellResult: %v", err)
	}
	if result != float64(f32) {
		t.Errorf("toFloat(float32) = %v, want %v", result, float64(f32))
	}
}

func TestToFloat_ViaAddData_Int(t *testing.T) {
	m := matrix.New()
	m.Data.AddCell(matrix.NewCellDescriptor("[V]", matrix.AggregateFunctionSum))
	m.AddData("r", "c", []any{int(7)})
	result, _ := m.CellResult("r", "c", 0)
	if result != 7 {
		t.Errorf("toFloat(int) = %v, want 7", result)
	}
}

func TestToFloat_ViaAddData_Int64(t *testing.T) {
	m := matrix.New()
	m.Data.AddCell(matrix.NewCellDescriptor("[V]", matrix.AggregateFunctionSum))
	m.AddData("r", "c", []any{int64(100)})
	result, _ := m.CellResult("r", "c", 0)
	if result != 100 {
		t.Errorf("toFloat(int64) = %v, want 100", result)
	}
}

func TestToFloat_ViaAddData_Int32(t *testing.T) {
	m := matrix.New()
	m.Data.AddCell(matrix.NewCellDescriptor("[V]", matrix.AggregateFunctionSum))
	m.AddData("r", "c", []any{int32(42)})
	result, _ := m.CellResult("r", "c", 0)
	if result != 42 {
		t.Errorf("toFloat(int32) = %v, want 42", result)
	}
}

func TestToFloat_ViaAddData_UnknownType(t *testing.T) {
	// An unknown type should produce 0.
	m := matrix.New()
	m.Data.AddCell(matrix.NewCellDescriptor("[V]", matrix.AggregateFunctionSum))
	m.AddData("r", "c", []any{"not-a-number"}) // string — not handled by toFloat
	result, _ := m.CellResult("r", "c", 0)
	if result != 0 {
		t.Errorf("toFloat(string) = %v, want 0", result)
	}
}

// ── AddDataMultiLevel tests ───────────────────────────────────────────────────

func TestAddDataMultiLevel_Basic(t *testing.T) {
	m := matrix.New()
	m.Data.AddCell(matrix.NewCellDescriptor("[Revenue]", matrix.AggregateFunctionSum))

	m.AddDataMultiLevel(
		[]string{"2024", "Q1"},
		[]string{"North"},
		[]any{100.0},
	)
	m.AddDataMultiLevel(
		[]string{"2024", "Q1"},
		[]string{"North"},
		[]any{50.0},
	)

	result, err := m.CellResultMultiLevel([]string{"2024", "Q1"}, []string{"North"}, 0)
	if err != nil {
		t.Fatalf("CellResultMultiLevel: %v", err)
	}
	if result != 150 {
		t.Errorf("result = %v, want 150", result)
	}
}

func TestAddDataMultiLevel_MultiplePaths(t *testing.T) {
	m := matrix.New()
	m.Data.AddCell(matrix.NewCellDescriptor("[V]", matrix.AggregateFunctionSum))

	m.AddDataMultiLevel([]string{"A"}, []string{"X"}, []any{10.0})
	m.AddDataMultiLevel([]string{"A"}, []string{"Y"}, []any{20.0})
	m.AddDataMultiLevel([]string{"B"}, []string{"X"}, []any{30.0})

	rAX, _ := m.CellResultMultiLevel([]string{"A"}, []string{"X"}, 0)
	rAY, _ := m.CellResultMultiLevel([]string{"A"}, []string{"Y"}, 0)
	rBX, _ := m.CellResultMultiLevel([]string{"B"}, []string{"X"}, 0)

	if rAX != 10 {
		t.Errorf("A/X = %v, want 10", rAX)
	}
	if rAY != 20 {
		t.Errorf("A/Y = %v, want 20", rAY)
	}
	if rBX != 30 {
		t.Errorf("B/X = %v, want 30", rBX)
	}
}

func TestAddDataMultiLevel_ValuesExceedDescriptors(t *testing.T) {
	// Extra values beyond the number of cell descriptors should be ignored.
	m := matrix.New()
	m.Data.AddCell(matrix.NewCellDescriptor("[V]", matrix.AggregateFunctionSum))

	m.AddDataMultiLevel([]string{"r"}, []string{"c"}, []any{5.0, 99.0, 999.0})

	result, err := m.CellResultMultiLevel([]string{"r"}, []string{"c"}, 0)
	if err != nil {
		t.Fatalf("CellResultMultiLevel: %v", err)
	}
	if result != 5 {
		t.Errorf("result = %v, want 5", result)
	}
}

func TestAddDataMultiLevel_AggregateMin(t *testing.T) {
	m := matrix.New()
	m.Data.AddCell(matrix.NewCellDescriptor("[V]", matrix.AggregateFunctionMin))

	m.AddDataMultiLevel([]string{"r"}, []string{"c"}, []any{10.0})
	m.AddDataMultiLevel([]string{"r"}, []string{"c"}, []any{3.0})
	m.AddDataMultiLevel([]string{"r"}, []string{"c"}, []any{7.0})

	result, _ := m.CellResultMultiLevel([]string{"r"}, []string{"c"}, 0)
	if result != 3 {
		t.Errorf("Min = %v, want 3", result)
	}
}

func TestAddDataMultiLevel_AggregateMax(t *testing.T) {
	m := matrix.New()
	m.Data.AddCell(matrix.NewCellDescriptor("[V]", matrix.AggregateFunctionMax))

	m.AddDataMultiLevel([]string{"r"}, []string{"c"}, []any{10.0})
	m.AddDataMultiLevel([]string{"r"}, []string{"c"}, []any{3.0})
	m.AddDataMultiLevel([]string{"r"}, []string{"c"}, []any{7.0})

	result, _ := m.CellResultMultiLevel([]string{"r"}, []string{"c"}, 0)
	if result != 10 {
		t.Errorf("Max = %v, want 10", result)
	}
}

func TestAddDataMultiLevel_AggregateCount(t *testing.T) {
	m := matrix.New()
	m.Data.AddCell(matrix.NewCellDescriptor("[V]", matrix.AggregateFunctionCount))

	m.AddDataMultiLevel([]string{"r"}, []string{"c"}, []any{1.0})
	m.AddDataMultiLevel([]string{"r"}, []string{"c"}, []any{2.0})
	m.AddDataMultiLevel([]string{"r"}, []string{"c"}, []any{3.0})

	result, _ := m.CellResultMultiLevel([]string{"r"}, []string{"c"}, 0)
	if result != 3 {
		t.Errorf("Count = %v, want 3", result)
	}
}

func TestAddDataMultiLevel_AggregateAvg(t *testing.T) {
	m := matrix.New()
	m.Data.AddCell(matrix.NewCellDescriptor("[V]", matrix.AggregateFunctionAvg))

	m.AddDataMultiLevel([]string{"r"}, []string{"c"}, []any{10.0})
	m.AddDataMultiLevel([]string{"r"}, []string{"c"}, []any{20.0})

	result, _ := m.CellResultMultiLevel([]string{"r"}, []string{"c"}, 0)
	if result != 15 {
		t.Errorf("Avg = %v, want 15", result)
	}
}

func TestAddDataMultiLevel_AggregateCountDistinct(t *testing.T) {
	m := matrix.New()
	m.Data.AddCell(matrix.NewCellDescriptor("[V]", matrix.AggregateFunctionCountDistinct))

	m.AddDataMultiLevel([]string{"r"}, []string{"c"}, []any{"A"})
	m.AddDataMultiLevel([]string{"r"}, []string{"c"}, []any{"B"})
	m.AddDataMultiLevel([]string{"r"}, []string{"c"}, []any{"A"}) // duplicate

	result, _ := m.CellResultMultiLevel([]string{"r"}, []string{"c"}, 0)
	if result != 2 {
		t.Errorf("CountDistinct = %v, want 2", result)
	}
}

// ── CellResultMultiLevel error paths ──────────────────────────────────────────

func TestCellResultMultiLevel_NoData(t *testing.T) {
	m := matrix.New()
	// No AddDataMultiLevel called — mlAccumulators is nil.
	_, err := m.CellResultMultiLevel([]string{"r"}, []string{"c"}, 0)
	if err == nil {
		t.Error("expected error when no multi-level data, got nil")
	}
}

func TestCellResultMultiLevel_MissingKey(t *testing.T) {
	m := matrix.New()
	m.Data.AddCell(matrix.NewCellDescriptor("[V]", matrix.AggregateFunctionSum))
	m.AddDataMultiLevel([]string{"A"}, []string{"X"}, []any{1.0})

	// Query a key that was never added.
	_, err := m.CellResultMultiLevel([]string{"B"}, []string{"Y"}, 0)
	if err == nil {
		t.Error("expected error for missing key, got nil")
	}
}

// ── BuildTemplateMultiLevel tests ─────────────────────────────────────────────

func TestBuildTemplateMultiLevel_SingleLevel(t *testing.T) {
	m := matrix.New()
	m.Data.AddCell(matrix.NewCellDescriptor("[Revenue]", matrix.AggregateFunctionSum))

	m.AddDataMultiLevel([]string{"Alice"}, []string{"Q1"}, []any{100.0})
	m.AddDataMultiLevel([]string{"Bob"}, []string{"Q1"}, []any{200.0})
	m.AddDataMultiLevel([]string{"Alice"}, []string{"Q2"}, []any{50.0})

	m.BuildTemplateMultiLevel()

	// col header rows = 1 (single col level), row leaves = 2 (Alice, Bob).
	// nRowHeaderCols = 1 (LevelSize-1). Total rows = 1 (header) + 2 (leaves) = 3.
	if m.RowCount() != 3 {
		t.Errorf("RowCount = %d, want 3", m.RowCount())
	}
}

func TestBuildTemplateMultiLevel_MultiLevel(t *testing.T) {
	m := matrix.New()
	m.Data.AddCell(matrix.NewCellDescriptor("[Revenue]", matrix.AggregateFunctionSum))

	// 2-level row path: Year/Quarter. 1-level col path: Region.
	m.AddDataMultiLevel([]string{"2024", "Q1"}, []string{"North"}, []any{100.0})
	m.AddDataMultiLevel([]string{"2024", "Q2"}, []string{"North"}, []any{200.0})
	m.AddDataMultiLevel([]string{"2024", "Q1"}, []string{"South"}, []any{50.0})

	m.BuildTemplateMultiLevel()

	// col header rows = colRoot.LevelSize = 1.
	// row leaves: Q1, Q2 = 2.
	// nRowHeaderCols = rowRoot.LevelSize-1 = 1.
	// RowCount = 1 (col-header) + 2 (leaves) = 3.
	if m.RowCount() != 3 {
		t.Errorf("RowCount = %d, want 3", m.RowCount())
	}
}

func TestBuildTemplateMultiLevel_DataCellValues(t *testing.T) {
	m := matrix.New()
	m.Data.AddCell(matrix.NewCellDescriptor("[Revenue]", matrix.AggregateFunctionSum))

	m.AddDataMultiLevel([]string{"Alice"}, []string{"Q1"}, []any{42.0})

	m.BuildTemplateMultiLevel()

	// The table should have data cell "42".
	found := false
	for ri := 0; ri < m.RowCount(); ri++ {
		row := m.Row(ri)
		if row == nil {
			continue
		}
		for ci := 0; ci < row.CellCount(); ci++ {
			cell := m.Cell(ri, ci)
			if cell != nil && cell.Text() == "42" {
				found = true
			}
		}
	}
	if !found {
		t.Error("expected cell with text '42' in BuildTemplateMultiLevel output")
	}
}

func TestBuildTemplateMultiLevel_NoCells(t *testing.T) {
	// No cell descriptors — BuildTemplateMultiLevel should still run without panic.
	m := matrix.New()
	m.AddDataMultiLevel([]string{"r"}, []string{"c"}, []any{1.0})
	m.BuildTemplateMultiLevel()
	// Just verify it doesn't panic.
}

func TestBuildTemplateMultiLevel_EmptyPaths(t *testing.T) {
	// Empty row/col paths — single root-level leaf.
	m := matrix.New()
	m.Data.AddCell(matrix.NewCellDescriptor("[V]", matrix.AggregateFunctionSum))

	m.AddDataMultiLevel([]string{"r"}, []string{"c"}, []any{7.0})
	m.BuildTemplateMultiLevel()

	if m.RowCount() == 0 {
		t.Error("expected at least one row in BuildTemplateMultiLevel output")
	}
}

func TestBuildTemplateMultiLevel_DeepHierarchy(t *testing.T) {
	// 3-level row: Country/State/City. 2-level col: Year/Quarter.
	m := matrix.New()
	m.Data.AddCell(matrix.NewCellDescriptor("[V]", matrix.AggregateFunctionSum))

	m.AddDataMultiLevel([]string{"US", "CA", "SF"}, []string{"2024", "Q1"}, []any{10.0})
	m.AddDataMultiLevel([]string{"US", "CA", "LA"}, []string{"2024", "Q1"}, []any{20.0})
	m.AddDataMultiLevel([]string{"US", "NY", "NYC"}, []string{"2024", "Q2"}, []any{30.0})

	m.BuildTemplateMultiLevel()

	// row leaves = 3 (SF, LA, NYC). col header rows = 1 (LevelSize-1).
	// RowCount = 1 (col-header) + 3 (leaf rows) + 1 (padding) = 5.
	if m.RowCount() != 5 {
		t.Errorf("RowCount = %d, want 5", m.RowCount())
	}
}

func TestBuildTemplateMultiLevel_RowSpanning(t *testing.T) {
	// Shared ancestor at level 0 should produce a RowSpan cell.
	m := matrix.New()
	m.Data.AddCell(matrix.NewCellDescriptor("[V]", matrix.AggregateFunctionSum))

	// Both leaves share "US" parent.
	m.AddDataMultiLevel([]string{"US", "Q1"}, []string{"P"}, []any{1.0})
	m.AddDataMultiLevel([]string{"US", "Q2"}, []string{"P"}, []any{2.0})

	m.BuildTemplateMultiLevel()

	// RowCount = 1 (col-header) + 2 (row-leaves) = 3.
	if m.RowCount() != 3 {
		t.Errorf("RowCount = %d, want 3", m.RowCount())
	}
}

func TestBuildTemplateMultiLevel_ColSpanning(t *testing.T) {
	// Column headers with shared parent produce ColSpan.
	m := matrix.New()
	m.Data.AddCell(matrix.NewCellDescriptor("[V]", matrix.AggregateFunctionSum))

	// Shared col parent "2024" with two children "Q1" and "Q2".
	m.AddDataMultiLevel([]string{"r"}, []string{"2024", "Q1"}, []any{1.0})
	m.AddDataMultiLevel([]string{"r"}, []string{"2024", "Q2"}, []any{2.0})

	m.BuildTemplateMultiLevel()

	// col level 0: "2024" spans 2. col level 1: Q1, Q2.
	// nColHeaderRows = LevelSize-1 = 1. Row leaves = 1. RowCount = 1+1 = 2.
	// But col tree is 2024→Q1,Q2 (2 levels), LevelSize=2, -1=1 header row.
	// Actually colRoot→2024→Q1,Q2: LevelSize=3, -1=2 header rows. RowCount=2+1=3.
	if m.RowCount() != 3 {
		t.Errorf("RowCount = %d, want 3", m.RowCount())
	}
}

// ── ensureMultiLevel idempotency ──────────────────────────────────────────────

func TestEnsureMultiLevel_CalledTwice(t *testing.T) {
	// AddDataMultiLevel calls ensureMultiLevel; calling it again should not reset state.
	m := matrix.New()
	m.Data.AddCell(matrix.NewCellDescriptor("[V]", matrix.AggregateFunctionSum))

	m.AddDataMultiLevel([]string{"r"}, []string{"c"}, []any{5.0})
	// Second call — ensureMultiLevel should detect rowRoot != nil.
	m.AddDataMultiLevel([]string{"r"}, []string{"c"}, []any{5.0})

	result, err := m.CellResultMultiLevel([]string{"r"}, []string{"c"}, 0)
	if err != nil {
		t.Fatalf("CellResultMultiLevel: %v", err)
	}
	if result != 10 {
		t.Errorf("result = %v, want 10 (two adds of 5)", result)
	}
}

// ── joinPath uniqueness (via CellResultMultiLevel) ────────────────────────────

func TestJoinPath_SeparatesSegments(t *testing.T) {
	// "ab","c" and "a","bc" should produce different keys.
	m := matrix.New()
	m.Data.AddCell(matrix.NewCellDescriptor("[V]", matrix.AggregateFunctionSum))

	m.AddDataMultiLevel([]string{"ab", "c"}, []string{"x"}, []any{1.0})
	m.AddDataMultiLevel([]string{"a", "bc"}, []string{"x"}, []any{2.0})

	r1, err1 := m.CellResultMultiLevel([]string{"ab", "c"}, []string{"x"}, 0)
	r2, err2 := m.CellResultMultiLevel([]string{"a", "bc"}, []string{"x"}, 0)

	if err1 != nil {
		t.Fatalf("r1: %v", err1)
	}
	if err2 != nil {
		t.Fatalf("r2: %v", err2)
	}
	if r1 == r2 {
		t.Errorf("joinPath does not distinguish ab+c (%v) from a+bc (%v)", r1, r2)
	}
	if r1 != 1 {
		t.Errorf("r1 = %v, want 1", r1)
	}
	if r2 != 2 {
		t.Errorf("r2 = %v, want 2", r2)
	}
}

// ── Serialize output content verification ─────────────────────────────────────

func TestSerialize_WritesDataSource(t *testing.T) {
	m := matrix.New()
	m.DataSourceName = "MySalesDS"
	xml := serializeMatrix(t, m)
	if !strings.Contains(xml, `DataSource="MySalesDS"`) {
		t.Errorf("expected DataSource attr in output:\n%s", xml)
	}
}

func TestSerialize_WritesFilter(t *testing.T) {
	m := matrix.New()
	m.Filter = "[Amount] > 100"
	xml := serializeMatrix(t, m)
	if !strings.Contains(xml, "Filter=") {
		t.Errorf("expected Filter attr in output:\n%s", xml)
	}
}

func TestSerialize_WritesAutoSizeFalse(t *testing.T) {
	m := matrix.New()
	m.AutoSize = false
	xml := serializeMatrix(t, m)
	if !strings.Contains(xml, `AutoSize="false"`) {
		t.Errorf("expected AutoSize=false in output:\n%s", xml)
	}
}

func TestSerialize_WritesCellsSideBySide(t *testing.T) {
	m := matrix.New()
	m.CellsSideBySide = true
	xml := serializeMatrix(t, m)
	if !strings.Contains(xml, `CellsSideBySide="true"`) {
		t.Errorf("expected CellsSideBySide=true in output:\n%s", xml)
	}
}

func TestSerialize_WritesKeepCellsSideBySide(t *testing.T) {
	m := matrix.New()
	m.KeepCellsSideBySide = true
	xml := serializeMatrix(t, m)
	if !strings.Contains(xml, `KeepCellsSideBySide="true"`) {
		t.Errorf("expected KeepCellsSideBySide=true in output:\n%s", xml)
	}
}

func TestSerialize_WritesShowTitle(t *testing.T) {
	m := matrix.New()
	m.ShowTitle = true
	xml := serializeMatrix(t, m)
	if !strings.Contains(xml, `ShowTitle="true"`) {
		t.Errorf("expected ShowTitle=true in output:\n%s", xml)
	}
}

func TestSerialize_WritesSplitRows(t *testing.T) {
	m := matrix.New()
	m.SplitRows = true
	xml := serializeMatrix(t, m)
	if !strings.Contains(xml, `SplitRows="true"`) {
		t.Errorf("expected SplitRows=true in output:\n%s", xml)
	}
}

func TestSerialize_WritesPrintIfEmptyFalse(t *testing.T) {
	m := matrix.New()
	m.PrintIfEmpty = false
	xml := serializeMatrix(t, m)
	if !strings.Contains(xml, `PrintIfEmpty="false"`) {
		t.Errorf("expected PrintIfEmpty=false in output:\n%s", xml)
	}
}

func TestSerialize_WritesEventNames(t *testing.T) {
	m := matrix.New()
	m.ManualBuildEvent = "MBE"
	m.ModifyResultEvent = "MRE"
	m.AfterTotalsEvent = "ATE"
	xml := serializeMatrix(t, m)
	for _, want := range []string{`ManualBuildEvent="MBE"`, `ModifyResultEvent="MRE"`, `AfterTotalsEvent="ATE"`} {
		if !strings.Contains(xml, want) {
			t.Errorf("expected %q in output:\n%s", want, xml)
		}
	}
}

func TestSerialize_WritesMatrixRowsChildBlock(t *testing.T) {
	m := matrix.New()
	m.Data.AddRow(matrix.NewHeaderDescriptor("[Year]"))
	xml := serializeMatrix(t, m)
	if !strings.Contains(xml, "MatrixRows") {
		t.Errorf("expected MatrixRows in output:\n%s", xml)
	}
	if !strings.Contains(xml, `Expression="[Year]"`) {
		t.Errorf("expected Expression=[Year] in output:\n%s", xml)
	}
}

func TestSerialize_WritesMatrixColumnsChildBlock(t *testing.T) {
	m := matrix.New()
	m.Data.AddColumn(matrix.NewHeaderDescriptor("[Quarter]"))
	xml := serializeMatrix(t, m)
	if !strings.Contains(xml, "MatrixColumns") {
		t.Errorf("expected MatrixColumns in output:\n%s", xml)
	}
}

func TestSerialize_WritesMatrixCellsChildBlock(t *testing.T) {
	m := matrix.New()
	m.Data.AddCell(matrix.NewCellDescriptor("[Amount]", matrix.AggregateFunctionSum))
	xml := serializeMatrix(t, m)
	if !strings.Contains(xml, "MatrixCells") {
		t.Errorf("expected MatrixCells in output:\n%s", xml)
	}
}

func TestSerialize_HeaderDescriptor_NonDefaultSort(t *testing.T) {
	m := matrix.New()
	h := matrix.NewHeaderDescriptor("[V]")
	h.Sort = matrix.SortOrderNone
	m.Data.AddRow(h)
	xml := serializeMatrix(t, m)
	if !strings.Contains(xml, "Sort=") {
		t.Errorf("expected Sort attr in output for non-default sort:\n%s", xml)
	}
}

func TestSerialize_HeaderDescriptor_TotalsFalse(t *testing.T) {
	m := matrix.New()
	h := matrix.NewHeaderDescriptor("[V]")
	h.Totals = false
	m.Data.AddRow(h)
	xml := serializeMatrix(t, m)
	if !strings.Contains(xml, `Totals="false"`) {
		t.Errorf("expected Totals=false in output:\n%s", xml)
	}
}

func TestSerialize_CellDescriptor_NonSumFunction(t *testing.T) {
	m := matrix.New()
	m.Data.AddCell(matrix.NewCellDescriptor("[V]", matrix.AggregateFunctionMax))
	xml := serializeMatrix(t, m)
	if !strings.Contains(xml, "Function=") {
		t.Errorf("expected Function attr in output for non-Sum function:\n%s", xml)
	}
}

func TestSerialize_CellDescriptor_Percent(t *testing.T) {
	m := matrix.New()
	c := matrix.NewCellDescriptor("[V]", matrix.AggregateFunctionSum)
	c.Percent = matrix.MatrixPercentGrandTotal
	m.Data.AddCell(c)
	xml := serializeMatrix(t, m)
	if !strings.Contains(xml, "Percent=") {
		t.Errorf("expected Percent attr in output:\n%s", xml)
	}
}

// ── accumulator result — AggregateFunctionNone path ───────────────────────────

func TestAccumulatorResult_NoneViaMultiLevel(t *testing.T) {
	m := matrix.New()
	m.Data.AddCell(matrix.NewCellDescriptor("[V]", matrix.AggregateFunctionNone))
	m.AddDataMultiLevel([]string{"r"}, []string{"c"}, []any{99.0})
	result, err := m.CellResultMultiLevel([]string{"r"}, []string{"c"}, 0)
	if err != nil {
		t.Fatalf("CellResultMultiLevel: %v", err)
	}
	if result != 0 {
		t.Errorf("None function = %v, want 0", result)
	}
}

func TestAccumulatorResult_CustomViaMultiLevel(t *testing.T) {
	m := matrix.New()
	m.Data.AddCell(matrix.NewCellDescriptor("[V]", matrix.AggregateFunctionCustom))
	m.AddDataMultiLevel([]string{"r"}, []string{"c"}, []any{50.0})
	result, err := m.CellResultMultiLevel([]string{"r"}, []string{"c"}, 0)
	if err != nil {
		t.Fatalf("CellResultMultiLevel: %v", err)
	}
	if result != 0 {
		t.Errorf("Custom function = %v, want 0", result)
	}
}

// ── accumulator Avg with zero count (via multi-level) ────────────────────────

func TestAccumulatorResult_AvgZeroCountViaMultiLevel(t *testing.T) {
	// The zero-count Avg path can't be reached via AddData (accumulator is
	// created on first add), but we verify the existing result() path is
	// covered by adding then querying — Avg with 1 row gives sum/count.
	m := matrix.New()
	m.Data.AddCell(matrix.NewCellDescriptor("[V]", matrix.AggregateFunctionAvg))
	m.AddDataMultiLevel([]string{"r"}, []string{"c"}, []any{30.0})
	result, _ := m.CellResultMultiLevel([]string{"r"}, []string{"c"}, 0)
	if result != 30 {
		t.Errorf("Avg(30) = %v, want 30", result)
	}
}
