package band_test

// band_porting_gaps2_test.go – tests for DataBand and GroupHeaderBand
// methods added in go-fastreport-mdnt4 (porting-gaps review 2026-03-21):
//
//   DataBand:
//     - GetExpressions() returns sort-expression list + filter
//     - Assign() deep-copies all property fields
//
//   GroupHeaderBand:
//     - Header / Footer accessors and AddChild routing
//     - GroupDataBand() traversal through nested groups
//     - DataSource() delegates to GroupDataBand
//     - ResetGroupValue() / GroupValueChanged() via injected calc function
//     - GetExpressions() returns the condition string
//     - Assign() copies scalar properties
//     - Serialize writes header/footer child bands

import (
	"testing"

	"github.com/andrewloable/go-fastreport/band"
)

// ── DataBand.GetExpressions ───────────────────────────────────────────────────

func TestDataBand_GetExpressions_FilterOnly(t *testing.T) {
	d := band.NewDataBand()
	d.SetFilter("[Amount] > 0")

	exprs := d.GetExpressions()
	// Expect exactly one sort expression (none) plus the filter.
	// GetExpressions appends filter unconditionally, so len should be 1.
	if len(exprs) != 1 {
		t.Fatalf("GetExpressions len = %d, want 1", len(exprs))
	}
	if exprs[0] != "[Amount] > 0" {
		t.Errorf("GetExpressions[0] = %q, want [Amount] > 0", exprs[0])
	}
}

func TestDataBand_GetExpressions_SortAndFilter(t *testing.T) {
	d := band.NewDataBand()
	d.AddSort(band.SortSpec{Expression: "[Orders.Date]"})
	d.AddSort(band.SortSpec{Column: "Price"})
	d.SetFilter("[Total] > 100")

	exprs := d.GetExpressions()
	// [Orders.Date], Price (column falls back to Column), [Total] > 100
	if len(exprs) != 3 {
		t.Fatalf("GetExpressions len = %d, want 3; exprs=%v", len(exprs), exprs)
	}
	if exprs[0] != "[Orders.Date]" {
		t.Errorf("exprs[0] = %q, want [Orders.Date]", exprs[0])
	}
	if exprs[1] != "Price" {
		t.Errorf("exprs[1] = %q, want Price", exprs[1])
	}
	if exprs[2] != "[Total] > 100" {
		t.Errorf("exprs[2] = %q, want [Total] > 100", exprs[2])
	}
}

func TestDataBand_GetExpressions_SortExpressionOverridesColumn(t *testing.T) {
	d := band.NewDataBand()
	// When both Expression and Column are set, Expression wins.
	d.AddSort(band.SortSpec{Column: "Col", Expression: "[Computed]"})

	exprs := d.GetExpressions()
	if len(exprs) != 2 {
		t.Fatalf("GetExpressions len = %d, want 2", len(exprs))
	}
	if exprs[0] != "[Computed]" {
		t.Errorf("exprs[0] = %q, want [Computed] (expression takes priority)", exprs[0])
	}
}

func TestDataBand_GetExpressions_NoSortNoFilter(t *testing.T) {
	d := band.NewDataBand()
	// No sort, empty filter: expect exactly one empty-string entry.
	exprs := d.GetExpressions()
	if len(exprs) != 1 {
		t.Fatalf("GetExpressions len = %d, want 1", len(exprs))
	}
	if exprs[0] != "" {
		t.Errorf("exprs[0] = %q, want empty string (empty filter)", exprs[0])
	}
}

func TestDataBand_GetExpressions_EmptySortSpec(t *testing.T) {
	// A SortSpec with both Expression="" and Column="" is silently omitted.
	d := band.NewDataBand()
	d.AddSort(band.SortSpec{})
	d.SetFilter("f")

	exprs := d.GetExpressions()
	// Empty SortSpec: Expression="" and Column="" → skipped, only filter.
	if len(exprs) != 1 {
		t.Fatalf("GetExpressions len = %d, want 1 (empty sort skipped)", len(exprs))
	}
	if exprs[0] != "f" {
		t.Errorf("exprs[0] = %q, want f", exprs[0])
	}
}

// ── DataBand.Assign ────────────────────────────────────────────────────────────

func TestDataBand_Assign_CopiesAllFields(t *testing.T) {
	src := band.NewDataBand()
	src.SetFilter("[X] > 1")
	src.SetVirtualRowCount(5)
	src.SetMaxRows(10)
	src.SetPrintIfDetailEmpty(true)
	src.SetPrintIfDSEmpty(true)
	src.SetKeepTogether(true)
	src.SetKeepDetail(true)
	src.SetIDColumn("id")
	src.SetParentIDColumn("parentId")
	src.SetIndent(3.5)
	src.SetCollectChildRows(true)
	src.SetResetPageNumber(true)
	src.AddSort(band.SortSpec{Column: "Name", Order: band.SortOrderDescending})

	dst := band.NewDataBand()
	dst.Assign(src)

	if dst.Filter() != "[X] > 1" {
		t.Errorf("Filter = %q, want [X] > 1", dst.Filter())
	}
	if dst.VirtualRowCount() != 5 {
		t.Errorf("VirtualRowCount = %d, want 5", dst.VirtualRowCount())
	}
	if dst.MaxRows() != 10 {
		t.Errorf("MaxRows = %d, want 10", dst.MaxRows())
	}
	if !dst.PrintIfDetailEmpty() {
		t.Error("PrintIfDetailEmpty should be true")
	}
	if !dst.PrintIfDSEmpty() {
		t.Error("PrintIfDSEmpty should be true")
	}
	if !dst.KeepTogether() {
		t.Error("KeepTogether should be true")
	}
	if !dst.KeepDetail() {
		t.Error("KeepDetail should be true")
	}
	if dst.IDColumn() != "id" {
		t.Errorf("IDColumn = %q, want id", dst.IDColumn())
	}
	if dst.ParentIDColumn() != "parentId" {
		t.Errorf("ParentIDColumn = %q, want parentId", dst.ParentIDColumn())
	}
	if dst.Indent() != 3.5 {
		t.Errorf("Indent = %v, want 3.5", dst.Indent())
	}
	if !dst.CollectChildRows() {
		t.Error("CollectChildRows should be true")
	}
	if !dst.ResetPageNumber() {
		t.Error("ResetPageNumber should be true")
	}
	if len(dst.Sort()) != 1 {
		t.Fatalf("Sort len = %d, want 1", len(dst.Sort()))
	}
	if dst.Sort()[0].Column != "Name" {
		t.Errorf("Sort[0].Column = %q, want Name", dst.Sort()[0].Column)
	}
	if dst.Sort()[0].Order != band.SortOrderDescending {
		t.Errorf("Sort[0].Order = %v, want Descending", dst.Sort()[0].Order)
	}
}

func TestDataBand_Assign_NilSrcIsNoop(t *testing.T) {
	d := band.NewDataBand()
	d.SetFilter("x")
	d.Assign(nil)
	// Should not panic; filter unchanged.
	if d.Filter() != "x" {
		t.Errorf("Filter = %q, want x", d.Filter())
	}
}

func TestDataBand_Assign_SortDeepCopy(t *testing.T) {
	src := band.NewDataBand()
	src.AddSort(band.SortSpec{Column: "Col1"})

	dst := band.NewDataBand()
	dst.Assign(src)

	// Modifying src sort after Assign should not affect dst.
	src.AddSort(band.SortSpec{Column: "Col2"})
	if len(dst.Sort()) != 1 {
		t.Errorf("Sort len = %d after src mutation, want 1 (deep copy)", len(dst.Sort()))
	}
}

// ── GroupHeaderBand.Header / Footer ──────────────────────────────────────────

func TestGroupHeaderBand_HeaderFooter_DefaultNil(t *testing.T) {
	g := band.NewGroupHeaderBand()
	if g.Header() != nil {
		t.Error("Header should default to nil")
	}
	if g.Footer() != nil {
		t.Error("Footer should default to nil")
	}
}

func TestGroupHeaderBand_SetHeader(t *testing.T) {
	g := band.NewGroupHeaderBand()
	h := band.NewDataHeaderBand()
	g.SetHeader(h)
	if g.Header() != h {
		t.Error("Header should return the set band")
	}
}

func TestGroupHeaderBand_SetFooter(t *testing.T) {
	g := band.NewGroupHeaderBand()
	f := band.NewDataFooterBand()
	g.SetFooter(f)
	if g.Footer() != f {
		t.Error("Footer should return the set band")
	}
}

// ── GroupHeaderBand.AddChild routes DataHeaderBand / DataFooterBand ──────────

func TestGroupHeaderBand_AddChild_DataHeaderBand(t *testing.T) {
	g := band.NewGroupHeaderBand()
	h := band.NewDataHeaderBand()
	g.AddChild(h)
	if g.Header() != h {
		t.Error("AddChild(DataHeaderBand) should set Header")
	}
}

func TestGroupHeaderBand_AddChild_DataFooterBand(t *testing.T) {
	g := band.NewGroupHeaderBand()
	f := band.NewDataFooterBand()
	g.AddChild(f)
	if g.Footer() != f {
		t.Error("AddChild(DataFooterBand) should set Footer")
	}
}

// ── GroupHeaderBand.GroupDataBand ─────────────────────────────────────────────

func TestGroupHeaderBand_GroupDataBand_Direct(t *testing.T) {
	g := band.NewGroupHeaderBand()
	db := band.NewDataBand()
	g.SetData(db)

	if g.GroupDataBand() != db {
		t.Error("GroupDataBand should return the directly set DataBand")
	}
}

func TestGroupHeaderBand_GroupDataBand_Nested(t *testing.T) {
	outer := band.NewGroupHeaderBand()
	inner := band.NewGroupHeaderBand()
	db := band.NewDataBand()
	outer.SetNestedGroup(inner)
	inner.SetData(db)

	if outer.GroupDataBand() != db {
		t.Error("GroupDataBand should traverse to the nested group's DataBand")
	}
}

func TestGroupHeaderBand_GroupDataBand_DeeplyNested(t *testing.T) {
	g1 := band.NewGroupHeaderBand()
	g2 := band.NewGroupHeaderBand()
	g3 := band.NewGroupHeaderBand()
	db := band.NewDataBand()
	g1.SetNestedGroup(g2)
	g2.SetNestedGroup(g3)
	g3.SetData(db)

	if g1.GroupDataBand() != db {
		t.Error("GroupDataBand should traverse three levels to find DataBand")
	}
}

func TestGroupHeaderBand_GroupDataBand_NoData(t *testing.T) {
	g := band.NewGroupHeaderBand()
	if g.GroupDataBand() != nil {
		t.Error("GroupDataBand should return nil when no DataBand is set")
	}
}

// ── GroupHeaderBand.DataSource ────────────────────────────────────────────────

type mockDS struct{}

func (m *mockDS) RowCount() int              { return 5 }
func (m *mockDS) First() error               { return nil }
func (m *mockDS) Next() error                { return nil }
func (m *mockDS) EOF() bool                  { return false }
func (m *mockDS) GetValue(col string) (any, error) { return nil, nil }

func TestGroupHeaderBand_DataSource_NilWhenNoDataBand(t *testing.T) {
	g := band.NewGroupHeaderBand()
	if g.DataSource() != nil {
		t.Error("DataSource should be nil when GroupDataBand is nil")
	}
}

func TestGroupHeaderBand_DataSource_DelegatestoDataBand(t *testing.T) {
	g := band.NewGroupHeaderBand()
	db := band.NewDataBand()
	ds := &mockDS{}
	db.SetDataSource(ds)
	g.SetData(db)

	got := g.DataSource()
	if got != ds {
		t.Error("DataSource should return the DataBand's data source")
	}
}

// ── GroupHeaderBand.ResetGroupValue / GroupValueChanged ───────────────────────

func TestGroupHeaderBand_ResetGroupValue_StoresValue(t *testing.T) {
	g := band.NewGroupHeaderBand()
	g.SetCondition("[CustomerName]")

	calc := func(expr string) (any, error) { return "ACME", nil }
	if err := g.ResetGroupValue(calc); err != nil {
		t.Fatalf("ResetGroupValue error: %v", err)
	}
	// After reset, GroupValueChanged should return false (same value).
	changed, err := g.GroupValueChanged(calc)
	if err != nil {
		t.Fatalf("GroupValueChanged error: %v", err)
	}
	if changed {
		t.Error("GroupValueChanged should be false immediately after ResetGroupValue with same value")
	}
}

func TestGroupHeaderBand_GroupValueChanged_True(t *testing.T) {
	g := band.NewGroupHeaderBand()
	g.SetCondition("[Dept]")

	v := "Sales"
	calc := func(expr string) (any, error) { return v, nil }
	if err := g.ResetGroupValue(calc); err != nil {
		t.Fatalf("ResetGroupValue error: %v", err)
	}

	// Simulate value change.
	v = "Engineering"
	changed, err := g.GroupValueChanged(calc)
	if err != nil {
		t.Fatalf("GroupValueChanged error: %v", err)
	}
	if !changed {
		t.Error("GroupValueChanged should be true after value changed")
	}
}

func TestGroupHeaderBand_GroupValueChanged_NilToNonNil(t *testing.T) {
	g := band.NewGroupHeaderBand()
	g.SetCondition("[X]")

	// Reset with nil value.
	calc := func(expr string) (any, error) { return nil, nil }
	if err := g.ResetGroupValue(calc); err != nil {
		t.Fatalf("ResetGroupValue error: %v", err)
	}

	// Now return a non-nil value.
	calc2 := func(expr string) (any, error) { return "something", nil }
	changed, err := g.GroupValueChanged(calc2)
	if err != nil {
		t.Fatalf("GroupValueChanged error: %v", err)
	}
	if !changed {
		t.Error("GroupValueChanged should be true when value changes from nil to non-nil")
	}
}

func TestGroupHeaderBand_GroupValueChanged_NilToNil(t *testing.T) {
	g := band.NewGroupHeaderBand()
	g.SetCondition("[X]")

	calc := func(expr string) (any, error) { return nil, nil }
	_ = g.ResetGroupValue(calc)

	changed, err := g.GroupValueChanged(calc)
	if err != nil {
		t.Fatalf("GroupValueChanged error: %v", err)
	}
	if changed {
		t.Error("GroupValueChanged should be false when both old and new value are nil")
	}
}

func TestGroupHeaderBand_ResetGroupValue_EmptyConditionNoOp(t *testing.T) {
	g := band.NewGroupHeaderBand()
	// Empty condition: ResetGroupValue is a no-op, returns nil.
	if err := g.ResetGroupValue(func(s string) (any, error) { return "x", nil }); err != nil {
		t.Errorf("ResetGroupValue with empty condition should return nil, got %v", err)
	}
}

func TestGroupHeaderBand_GroupValueChanged_EmptyConditionReturnsFalse(t *testing.T) {
	g := band.NewGroupHeaderBand()
	// Empty condition: GroupValueChanged returns false.
	changed, err := g.GroupValueChanged(func(s string) (any, error) { return "x", nil })
	if err != nil {
		t.Errorf("GroupValueChanged with empty condition should return nil error, got %v", err)
	}
	if changed {
		t.Error("GroupValueChanged with empty condition should return false")
	}
}

// ── GroupHeaderBand.GetExpressions ────────────────────────────────────────────

func TestGroupHeaderBand_GetExpressions_ReturnsCondition(t *testing.T) {
	g := band.NewGroupHeaderBand()
	g.SetCondition("[Orders.CustomerName]")

	exprs := g.GetExpressions()
	if len(exprs) != 1 {
		t.Fatalf("GetExpressions len = %d, want 1", len(exprs))
	}
	if exprs[0] != "[Orders.CustomerName]" {
		t.Errorf("exprs[0] = %q, want [Orders.CustomerName]", exprs[0])
	}
}

func TestGroupHeaderBand_GetExpressions_EmptyCondition(t *testing.T) {
	g := band.NewGroupHeaderBand()
	exprs := g.GetExpressions()
	if len(exprs) != 1 {
		t.Fatalf("GetExpressions len = %d, want 1", len(exprs))
	}
	if exprs[0] != "" {
		t.Errorf("exprs[0] = %q, want empty string", exprs[0])
	}
}

// ── GroupHeaderBand.Assign ────────────────────────────────────────────────────

func TestGroupHeaderBand_Assign_CopiesProperties(t *testing.T) {
	src := band.NewGroupHeaderBand()
	src.SetCondition("[Region]")
	src.SetSortOrder(band.SortOrderDescending)
	src.SetKeepTogether(true)
	src.SetResetPageNumber(true)

	dst := band.NewGroupHeaderBand()
	dst.Assign(src)

	if dst.Condition() != "[Region]" {
		t.Errorf("Condition = %q, want [Region]", dst.Condition())
	}
	if dst.SortOrder() != band.SortOrderDescending {
		t.Errorf("SortOrder = %v, want Descending", dst.SortOrder())
	}
	if !dst.KeepTogether() {
		t.Error("KeepTogether should be true")
	}
	if !dst.ResetPageNumber() {
		t.Error("ResetPageNumber should be true")
	}
}

func TestGroupHeaderBand_Assign_NilSrcIsNoop(t *testing.T) {
	g := band.NewGroupHeaderBand()
	g.SetCondition("[X]")
	g.Assign(nil)
	// Should not panic; condition unchanged.
	if g.Condition() != "[X]" {
		t.Errorf("Condition = %q, want [X]", g.Condition())
	}
}

func TestGroupHeaderBand_Assign_DoesNotCopyChildBands(t *testing.T) {
	src := band.NewGroupHeaderBand()
	nested := band.NewGroupHeaderBand()
	src.SetNestedGroup(nested)
	db := band.NewDataBand()
	src.SetData(db)

	dst := band.NewGroupHeaderBand()
	dst.Assign(src)

	// Structural child-band references should NOT be copied by Assign.
	if dst.NestedGroup() != nil {
		t.Error("Assign should not copy NestedGroup reference")
	}
	if dst.Data() != nil {
		t.Error("Assign should not copy Data reference")
	}
}
