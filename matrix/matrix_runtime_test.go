package matrix_test

// matrix_runtime_test.go tests the runtime pipeline gaps implemented in this
// iteration:
//
//  1. HeaderItem rich fields: ItemParent, ItemIndex, ItemIsTotal, ItemDataRowNo,
//     ItemPageBreak, ItemIsSplitted, ItemAnyValue, ItemValues, ItemSpan,
//     ItemClearChildren, ItemGetTerminalItems, ItemFind.
//
//  2. MatrixHeader tree: Find (create/lookup), FindIndex, FindOrCreate,
//     RemoveItem, GetTerminalIndices, AddTotalItems, Reset.
//
//  3. CellStore: AddValues, GetValue, GetValues, SetValues, IsEmpty.
//
//  4. MatrixData runtime API: AddValue, AddValueAt, GetValue, SetValue,
//     SetValues, IsEmpty, RuntimeReset, ColumnHeader, RowHeader.
//
//  5. MatrixObject lifecycle: SaveState/RestoreState, OnManualBuild,
//     OnModifyResult, OnAfterTotals, AddValue shortcut, Value(index),
//     ColumnValues/RowValues2/CellValues/ColumnIndex/RowIndex fields.

import (
	"testing"

	"github.com/andrewloable/go-fastreport/matrix"
)

// ── HeaderItem rich fields ────────────────────────────────────────────────────

func TestItemParent_SetGet(t *testing.T) {
	parent := matrix.NewHeaderDescriptor("[p]")
	_ = parent
	// Use the exported NewMatrixHeader to create items through Find.
	h := matrix.NewMatrixHeader()
	h.Descriptors = []*matrix.HeaderDescriptor{matrix.NewHeaderDescriptor("[col]")}
	item := h.Find([]any{"A"}, true, 0)
	if item == nil {
		t.Fatal("Find should create item")
	}
	// Root of h should be the parent of item.
	if matrix.ItemParent(item) != h.Root {
		t.Error("parent should be root")
	}
}

func TestItemIndex_Assigned(t *testing.T) {
	h := matrix.NewMatrixHeader()
	h.Descriptors = []*matrix.HeaderDescriptor{matrix.NewHeaderDescriptor("[col]")}

	i1 := h.Find([]any{"A"}, true, 0)
	i2 := h.Find([]any{"B"}, true, 0)

	if matrix.ItemIndex(i1) != 0 {
		t.Errorf("A index = %d, want 0", matrix.ItemIndex(i1))
	}
	if matrix.ItemIndex(i2) != 1 {
		t.Errorf("B index = %d, want 1", matrix.ItemIndex(i2))
	}
}

func TestItemAnyValue_Typed(t *testing.T) {
	h := matrix.NewMatrixHeader()
	h.Descriptors = []*matrix.HeaderDescriptor{matrix.NewHeaderDescriptor("[year]")}

	item := h.Find([]any{2024}, true, 0)
	if matrix.ItemAnyValue(item) != 2024 {
		t.Errorf("AnyValue = %v, want 2024", matrix.ItemAnyValue(item))
	}
}

func TestItemIsTotal_Default(t *testing.T) {
	h := matrix.NewMatrixHeader()
	h.Descriptors = []*matrix.HeaderDescriptor{matrix.NewHeaderDescriptor("[v]")}
	item := h.Find([]any{"x"}, true, 0)
	if matrix.ItemIsTotal(item) {
		t.Error("new item should not be IsTotal")
	}
}

func TestItemDataRowNo(t *testing.T) {
	h := matrix.NewMatrixHeader()
	h.Descriptors = []*matrix.HeaderDescriptor{matrix.NewHeaderDescriptor("[v]")}
	item := h.Find([]any{"x"}, true, 42)
	if matrix.ItemDataRowNo(item) != 42 {
		t.Errorf("DataRowNo = %d, want 42", matrix.ItemDataRowNo(item))
	}
}

func TestItemPageBreak(t *testing.T) {
	descr := matrix.NewHeaderDescriptor("[v]")
	descr.PageBreak = true
	h := matrix.NewMatrixHeader()
	h.Descriptors = []*matrix.HeaderDescriptor{descr}
	item := h.Find([]any{"x"}, true, 0)
	if !matrix.ItemPageBreak(item) {
		t.Error("PageBreak should be true when descriptor has PageBreak=true")
	}
}

func TestItemIsSplitted_SetGet(t *testing.T) {
	h := matrix.NewMatrixHeader()
	h.Descriptors = []*matrix.HeaderDescriptor{matrix.NewHeaderDescriptor("[v]")}
	item := h.Find([]any{"x"}, true, 0)

	matrix.SetItemIsSplitted(item, true)
	if !matrix.ItemIsSplitted(item) {
		t.Error("IsSplitted should be true after SetItemIsSplitted(true)")
	}
	matrix.SetItemIsSplitted(item, false)
	if matrix.ItemIsSplitted(item) {
		t.Error("IsSplitted should be false after SetItemIsSplitted(false)")
	}
}

func TestItemValues_TwoLevels(t *testing.T) {
	h := matrix.NewMatrixHeader()
	h.Descriptors = []*matrix.HeaderDescriptor{
		matrix.NewHeaderDescriptor("[year]"),
		matrix.NewHeaderDescriptor("[month]"),
	}
	item := h.Find([]any{2024, "Jan"}, true, 0)
	vals := matrix.ItemValues(item)
	if len(vals) != 2 {
		t.Fatalf("Values len = %d, want 2", len(vals))
	}
	if vals[0] != 2024 || vals[1] != "Jan" {
		t.Errorf("Values = %v, want [2024 Jan]", vals)
	}
}

func TestItemClearChildren(t *testing.T) {
	h := matrix.NewMatrixHeader()
	h.Descriptors = []*matrix.HeaderDescriptor{matrix.NewHeaderDescriptor("[v]")}
	h.Find([]any{"A"}, true, 0)
	h.Find([]any{"B"}, true, 0)
	if len(h.Root.Children) != 2 {
		t.Fatalf("expected 2 children, got %d", len(h.Root.Children))
	}
	matrix.ItemClearChildren(h.Root)
	if len(h.Root.Children) != 0 {
		t.Errorf("children not cleared: %d", len(h.Root.Children))
	}
}

func TestItemGetTerminalItems_IgnoresSplitted(t *testing.T) {
	h := matrix.NewMatrixHeader()
	h.Descriptors = []*matrix.HeaderDescriptor{matrix.NewHeaderDescriptor("[v]")}
	itemA := h.Find([]any{"A"}, true, 0)
	h.Find([]any{"B"}, true, 0)

	// Mark A as splitted — it should be excluded from terminal items.
	matrix.SetItemIsSplitted(itemA, true)
	terminals := matrix.ItemGetTerminalItems(h.Root)
	for _, t2 := range terminals {
		if t2 == itemA {
			t.Error("splitted item should not appear in terminal items")
		}
	}
}

func TestItemFind_AscendingSort(t *testing.T) {
	h := matrix.NewMatrixHeader()
	descr := matrix.NewHeaderDescriptor("[v]")
	descr.Sort = matrix.SortOrderAscending
	h.Descriptors = []*matrix.HeaderDescriptor{descr}

	// Insert A, B, C in order.
	h.Find([]any{"A"}, true, 0)
	h.Find([]any{"B"}, true, 0)
	h.Find([]any{"C"}, true, 0)

	// Existing: Find returns non-negative index.
	idx := matrix.ItemFind(h.Root, "B", matrix.SortOrderAscending)
	if idx < 0 {
		t.Errorf("Find B = %d, want >= 0", idx)
	}
	// Missing: Find returns negative (bitwise complement of insertion point).
	idx = matrix.ItemFind(h.Root, "Z", matrix.SortOrderAscending)
	if idx >= 0 {
		t.Errorf("Find Z = %d, want negative", idx)
	}
}

func TestItemFind_NoSort(t *testing.T) {
	h := matrix.NewMatrixHeader()
	descr := matrix.NewHeaderDescriptor("[v]")
	descr.Sort = matrix.SortOrderNone
	h.Descriptors = []*matrix.HeaderDescriptor{descr}

	h.Find([]any{"X"}, true, 0)
	h.Find([]any{"Y"}, true, 0)

	// Linear scan should find "X".
	idx := matrix.ItemFind(h.Root, "X", matrix.SortOrderNone)
	if idx < 0 {
		t.Errorf("Find X (none sort) = %d, want >= 0", idx)
	}
}

// ── MatrixHeader tree navigation ──────────────────────────────────────────────

func TestMatrixHeader_FindAndLookup(t *testing.T) {
	h := matrix.NewMatrixHeader()
	h.Descriptors = []*matrix.HeaderDescriptor{matrix.NewHeaderDescriptor("[v]")}

	// Create item.
	item := h.Find([]any{"A"}, true, 0)
	if item == nil {
		t.Fatal("Find(create=true) returned nil")
	}

	// Lookup should find it.
	found := h.Find([]any{"A"}, false, 0)
	if found != item {
		t.Error("Find(create=false) should return the same item")
	}

	// Not-found lookup should return nil.
	notFound := h.Find([]any{"Z"}, false, 0)
	if notFound != nil {
		t.Error("Find(create=false) for absent address should return nil")
	}
}

func TestMatrixHeader_FindIndex(t *testing.T) {
	h := matrix.NewMatrixHeader()
	h.Descriptors = []*matrix.HeaderDescriptor{matrix.NewHeaderDescriptor("[v]")}

	h.Find([]any{"A"}, true, 0)
	h.Find([]any{"B"}, true, 0)

	if got := h.FindIndex([]any{"A"}); got != 0 {
		t.Errorf("FindIndex A = %d, want 0", got)
	}
	if got := h.FindIndex([]any{"B"}); got != 1 {
		t.Errorf("FindIndex B = %d, want 1", got)
	}
	if got := h.FindIndex([]any{"Z"}); got != -1 {
		t.Errorf("FindIndex Z = %d, want -1", got)
	}
}

func TestMatrixHeader_FindOrCreate(t *testing.T) {
	h := matrix.NewMatrixHeader()
	h.Descriptors = []*matrix.HeaderDescriptor{matrix.NewHeaderDescriptor("[v]")}

	idx := h.FindOrCreate([]any{"new"})
	if idx < 0 {
		t.Errorf("FindOrCreate returned %d, want >= 0", idx)
	}
	// Second call should return the same index.
	idx2 := h.FindOrCreate([]any{"new"})
	if idx != idx2 {
		t.Errorf("FindOrCreate second call: %d != %d", idx, idx2)
	}
}

func TestMatrixHeader_RemoveItem(t *testing.T) {
	h := matrix.NewMatrixHeader()
	h.Descriptors = []*matrix.HeaderDescriptor{matrix.NewHeaderDescriptor("[v]")}

	h.Find([]any{"A"}, true, 0)
	h.Find([]any{"B"}, true, 0)
	if len(h.Root.Children) != 2 {
		t.Fatalf("expected 2 children before remove")
	}

	h.RemoveItem([]any{"A"})
	if len(h.Root.Children) != 1 {
		t.Errorf("expected 1 child after RemoveItem, got %d", len(h.Root.Children))
	}

	// Removing an absent item should be a no-op.
	h.RemoveItem([]any{"Z"})
	if len(h.Root.Children) != 1 {
		t.Errorf("RemoveItem absent should be noop: got %d children", len(h.Root.Children))
	}
}

func TestMatrixHeader_GetTerminalIndices(t *testing.T) {
	h := matrix.NewMatrixHeader()
	h.Descriptors = []*matrix.HeaderDescriptor{matrix.NewHeaderDescriptor("[v]")}

	h.Find([]any{"A"}, true, 0) // index 0
	h.Find([]any{"B"}, true, 0) // index 1
	h.Find([]any{"C"}, true, 0) // index 2

	indices := h.GetTerminalIndices()
	if len(indices) != 3 {
		t.Fatalf("GetTerminalIndices len = %d, want 3", len(indices))
	}
}

func TestMatrixHeader_Reset(t *testing.T) {
	h := matrix.NewMatrixHeader()
	h.Descriptors = []*matrix.HeaderDescriptor{matrix.NewHeaderDescriptor("[v]")}

	h.Find([]any{"A"}, true, 0)
	h.Find([]any{"B"}, true, 0)
	h.Reset()

	if len(h.Root.Children) != 0 {
		t.Errorf("after Reset, children len = %d, want 0", len(h.Root.Children))
	}

	// nextIndex should be reset: next created item should be index 0.
	item := h.Find([]any{"New"}, true, 0)
	if matrix.ItemIndex(item) != 0 {
		t.Errorf("after Reset, first new item index = %d, want 0", matrix.ItemIndex(item))
	}
}

func TestMatrixHeader_AddTotalItems(t *testing.T) {
	descr := matrix.NewHeaderDescriptor("[v]")
	descr.Totals = true
	descr.TotalsFirst = false

	h := matrix.NewMatrixHeader()
	h.Descriptors = []*matrix.HeaderDescriptor{descr}

	h.Find([]any{"A"}, true, 0)
	h.Find([]any{"B"}, true, 0)

	// Before AddTotalItems: 2 children.
	if len(h.Root.Children) != 2 {
		t.Fatalf("expected 2 children before AddTotalItems")
	}

	h.AddTotalItems(false)

	// After: 3 children (A, B, total).
	if len(h.Root.Children) != 3 {
		t.Errorf("after AddTotalItems, children = %d, want 3", len(h.Root.Children))
	}
	// The total item should be last (TotalsFirst=false).
	total := h.Root.Children[2]
	if !matrix.ItemIsTotal(total) {
		t.Error("last child should be a total item")
	}
}

func TestMatrixHeader_AddTotalItems_SuppressTotals_OneValue(t *testing.T) {
	descr := matrix.NewHeaderDescriptor("[v]")
	descr.Totals = true
	descr.SuppressTotals = true

	h := matrix.NewMatrixHeader()
	h.Descriptors = []*matrix.HeaderDescriptor{descr}
	h.Find([]any{"OnlyOne"}, true, 0)

	// SuppressTotals with only 1 child → total should NOT be added.
	h.AddTotalItems(false)
	if len(h.Root.Children) != 1 {
		t.Errorf("SuppressTotals with 1 child: children = %d, want 1", len(h.Root.Children))
	}
}

func TestMatrixHeader_AddTotalItems_TotalsFirst(t *testing.T) {
	descr := matrix.NewHeaderDescriptor("[v]")
	descr.Totals = true
	descr.TotalsFirst = true

	h := matrix.NewMatrixHeader()
	h.Descriptors = []*matrix.HeaderDescriptor{descr}
	h.Find([]any{"A"}, true, 0)
	h.Find([]any{"B"}, true, 0)
	h.AddTotalItems(false)

	// TotalsFirst=true → total at position 0.
	first := h.Root.Children[0]
	if !matrix.ItemIsTotal(first) {
		t.Error("with TotalsFirst=true, first child should be total")
	}
}

func TestMatrixHeader_GetTerminalIndicesAt(t *testing.T) {
	h := matrix.NewMatrixHeader()
	h.Descriptors = []*matrix.HeaderDescriptor{
		matrix.NewHeaderDescriptor("[year]"),
		matrix.NewHeaderDescriptor("[month]"),
	}
	h.Find([]any{"2024", "Jan"}, true, 0)
	h.Find([]any{"2024", "Feb"}, true, 0)
	h.Find([]any{"2025", "Jan"}, true, 0)

	// Terminal indices under year=2024.
	indices := h.GetTerminalIndicesAt([]any{"2024"})
	if len(indices) != 2 {
		t.Errorf("GetTerminalIndicesAt 2024: len = %d, want 2", len(indices))
	}
}

// ── MatrixData AddValue / GetValue / SetValue ─────────────────────────────────

func TestMatrixData_AddValueGetValue(t *testing.T) {
	var d matrix.MatrixData
	d.AddColumn(matrix.NewHeaderDescriptor("[year]"))
	d.AddRow(matrix.NewHeaderDescriptor("[name]"))
	d.AddCell(matrix.NewCellDescriptor("[rev]", matrix.AggregateFunctionSum))

	d.AddValue([]any{2024}, []any{"Alice"}, []any{100.0})
	d.AddValue([]any{2024}, []any{"Alice"}, []any{200.0})
	d.AddValue([]any{2025}, []any{"Alice"}, []any{50.0})

	// colIdx for 2024 is 0, rowIdx for Alice is 0.
	vals := d.GetValues(0, 0, 0) // col=0, row=0, cell=0
	if len(vals) != 2 {
		t.Fatalf("GetValues: len=%d, want 2 (two AddValue calls)", len(vals))
	}

	// col=1 (2025), row=0 (Alice), cell=0 should have 1 value.
	vals2 := d.GetValues(1, 0, 0)
	if len(vals2) != 1 {
		t.Fatalf("GetValues 2025/Alice: len=%d, want 1", len(vals2))
	}
}

func TestMatrixData_IsEmpty(t *testing.T) {
	var d matrix.MatrixData
	d.AddColumn(matrix.NewHeaderDescriptor("[v]"))
	d.AddRow(matrix.NewHeaderDescriptor("[v]"))
	d.AddCell(matrix.NewCellDescriptor("[v]", matrix.AggregateFunctionSum))

	if !d.IsEmpty() {
		t.Error("IsEmpty should be true before any AddValue")
	}
	d.AddValue([]any{"A"}, []any{"r"}, []any{1.0})
	if d.IsEmpty() {
		t.Error("IsEmpty should be false after AddValue")
	}
}

func TestMatrixData_SetValue(t *testing.T) {
	var d matrix.MatrixData
	d.AddColumn(matrix.NewHeaderDescriptor("[c]"))
	d.AddRow(matrix.NewHeaderDescriptor("[r]"))
	d.AddCell(matrix.NewCellDescriptor("[v]", matrix.AggregateFunctionSum))

	d.AddValue([]any{"col"}, []any{"row"}, []any{10.0})
	// colIdx=0, rowIdx=0 → set to new single value.
	d.SetValue(0, 0, 99.0)
	got := d.GetValue(0, 0, 0)
	if got != 99.0 {
		t.Errorf("after SetValue: GetValue = %v, want 99", got)
	}
}

func TestMatrixData_RuntimeReset(t *testing.T) {
	var d matrix.MatrixData
	d.AddColumn(matrix.NewHeaderDescriptor("[c]"))
	d.AddRow(matrix.NewHeaderDescriptor("[r]"))
	d.AddCell(matrix.NewCellDescriptor("[v]", matrix.AggregateFunctionSum))

	d.AddValue([]any{"A"}, []any{"x"}, []any{1.0})
	if d.IsEmpty() {
		t.Fatal("should not be empty after AddValue")
	}
	d.RuntimeReset()
	if !d.IsEmpty() {
		t.Error("should be empty after RuntimeReset")
	}
}

func TestMatrixData_ColumnRowHeader(t *testing.T) {
	var d matrix.MatrixData
	d.AddColumn(matrix.NewHeaderDescriptor("[c]"))
	d.AddRow(matrix.NewHeaderDescriptor("[r]"))
	d.AddCell(matrix.NewCellDescriptor("[v]", matrix.AggregateFunctionSum))

	d.AddValue([]any{"col1"}, []any{"row1"}, []any{5.0})
	d.AddValue([]any{"col2"}, []any{"row1"}, []any{3.0})

	colH := d.ColumnHeader()
	if colH == nil {
		t.Fatal("ColumnHeader should not be nil after AddValue")
	}
	rowH := d.RowHeader()
	if rowH == nil {
		t.Fatal("RowHeader should not be nil after AddValue")
	}
	// Two distinct column values → two children.
	if len(colH.Root.Children) != 2 {
		t.Errorf("column header children = %d, want 2", len(colH.Root.Children))
	}
	// One distinct row value → one child.
	if len(rowH.Root.Children) != 1 {
		t.Errorf("row header children = %d, want 1", len(rowH.Root.Children))
	}
}

// ── MatrixObject lifecycle ────────────────────────────────────────────────────

func TestMatrixObject_SaveRestoreState(t *testing.T) {
	m := matrix.New()
	// visible is true by default.
	m.SaveState()
	// After SaveState, engine would set visible=false; RestoreState brings it back.
	m.RestoreState()
	// Just verify no panic — the visible field is not exported for direct check.
}

func TestMatrixObject_OnManualBuild(t *testing.T) {
	m := matrix.New()
	called := false
	m.ManualBuildHandler = func(_ *matrix.MatrixObject) { called = true }
	m.OnManualBuild()
	if !called {
		t.Error("ManualBuildHandler should have been called")
	}
}

func TestMatrixObject_OnManualBuild_NilHandler(t *testing.T) {
	m := matrix.New()
	m.ManualBuildHandler = nil
	m.OnManualBuild() // must not panic
}

func TestMatrixObject_OnModifyResult(t *testing.T) {
	m := matrix.New()
	called := false
	m.ModifyResultHandler = func(_ *matrix.MatrixObject) { called = true }
	m.OnModifyResult()
	if !called {
		t.Error("ModifyResultHandler should have been called")
	}
}

func TestMatrixObject_OnAfterTotals(t *testing.T) {
	m := matrix.New()
	called := false
	m.AfterTotalsHandler = func(_ *matrix.MatrixObject) { called = true }
	m.OnAfterTotals()
	if !called {
		t.Error("AfterTotalsHandler should have been called")
	}
}

func TestMatrixObject_OnAfterData(t *testing.T) {
	m := matrix.New()
	called := false
	m.ModifyResultHandler = func(_ *matrix.MatrixObject) { called = true }
	m.OnAfterData()
	if !called {
		t.Error("OnAfterData should fire ModifyResult")
	}
}

func TestMatrixObject_AddValueShortcut(t *testing.T) {
	m := matrix.New()
	m.Data.AddColumn(matrix.NewHeaderDescriptor("[c]"))
	m.Data.AddRow(matrix.NewHeaderDescriptor("[r]"))
	m.Data.AddCell(matrix.NewCellDescriptor("[v]", matrix.AggregateFunctionSum))

	m.AddValue([]any{"col"}, []any{"row"}, []any{42.0})

	vals := m.Data.GetValues(0, 0, 0)
	if len(vals) != 1 || vals[0] != 42.0 {
		t.Errorf("AddValue shortcut: GetValues = %v, want [42]", vals)
	}
}

func TestMatrixObject_Value(t *testing.T) {
	m := matrix.New()
	m.CellValues = []any{10.5, nil, 7.0}

	if got := m.Value(0); got != 10.5 {
		t.Errorf("Value(0) = %v, want 10.5", got)
	}
	// nil entry → returns 0.
	if got := m.Value(1); got != 0 {
		t.Errorf("Value(1) nil = %v, want 0", got)
	}
	// Out-of-range → 0.
	if got := m.Value(99); got != 0 {
		t.Errorf("Value(99) = %v, want 0", got)
	}
}

func TestMatrixObject_RuntimeStateFields(t *testing.T) {
	m := matrix.New()
	m.ColumnValues = []any{2024, "Q1"}
	m.RowValues2 = []any{"Alice"}
	m.CellValues = []any{100.0}
	m.ColumnIndex = 3
	m.RowIndex = 7

	if len(m.ColumnValues) != 2 {
		t.Errorf("ColumnValues len = %d, want 2", len(m.ColumnValues))
	}
	if m.RowValues2[0] != "Alice" {
		t.Errorf("RowValues2 = %v, want [Alice]", m.RowValues2)
	}
	if m.ColumnIndex != 3 {
		t.Errorf("ColumnIndex = %d, want 3", m.ColumnIndex)
	}
	if m.RowIndex != 7 {
		t.Errorf("RowIndex = %d, want 7", m.RowIndex)
	}
}

func TestMatrixObject_InitializeFinalizeComponent(t *testing.T) {
	m := matrix.New()
	m.InitializeComponent() // must not panic
	m.FinalizeComponent()   // must not panic
}

// ── GetDataWithCalc ───────────────────────────────────────────────────────────

func TestMatrixObject_GetDataWithCalc_NilDataSource(t *testing.T) {
	m := matrix.New()
	m.Data.AddColumn(matrix.NewHeaderDescriptor("[c]"))
	m.Data.AddRow(matrix.NewHeaderDescriptor("[r]"))
	m.Data.AddCell(matrix.NewCellDescriptor("[v]", matrix.AggregateFunctionSum))

	// ManualBuild handler adds data directly.
	m.ManualBuildHandler = func(mx *matrix.MatrixObject) {
		mx.AddValue([]any{"col"}, []any{"row"}, []any{99.0})
	}

	m.GetDataWithCalc(nil)

	vals := m.Data.GetValues(0, 0, 0)
	if len(vals) != 1 || vals[0] != 99.0 {
		t.Errorf("after GetDataWithCalc with ManualBuild: vals = %v, want [99]", vals)
	}
}
