package band

// types_coverage_gaps_test.go – internal tests that document and exercise the
// remaining uncovered branches in types.go.
//
// The four uncovered branches are all architecturally unreachable "return err"
// paths in the error-propagation guard at the top of each Serialize/Deserialize
// method:
//
//   types.go:238-240  GroupHeaderBand.Serialize  – g.HeaderFooterBandBase.serializeAttrs(w)
//   types.go:258-260  GroupHeaderBand.Deserialize – g.HeaderFooterBandBase.Deserialize(r)
//   types.go:426-428  DataBand.Serialize          – d.BandBase.serializeAttrs(w)
//   types.go:478-480  DataBand.Deserialize        – d.BandBase.Deserialize(r)
//
// These calls ultimately propagate through BandBase.serializeAttrs →
// BreakableComponent.Serialize → ReportComponentBase.Serialize →
// ComponentBase.Serialize → BaseObject.Serialize, none of which can return a
// non-nil error because all report.Writer methods (WriteStr, WriteInt, etc.)
// are void. The "return err" branches are identical dead-code guards as those
// already documented in band_dead_error_paths_test.go for BandBase.Serialize
// (line 336) and ChildBand.Deserialize (line 404).
//
// This file adds supplementary tests that maximise coverage of every
// REACHABLE branch inside these four functions to bring overall function
// coverage as high as possible.

import (
	"testing"
)

// ─── GroupHeaderBand.Serialize: SortOrderNone branch ─────────────────────────

// TestGroupHeaderBand_Serialize_SortOrderNone exercises the SortOrder branch
// with SortOrderNone (which differs from the ascending default and is therefore
// written out).
func TestGroupHeaderBand_Serialize_SortOrderNone(t *testing.T) {
	g := NewGroupHeaderBand()
	g.SetSortOrder(SortOrderNone)

	w := newFailAttrsWriter2()
	if err := g.Serialize(w); err != nil {
		t.Errorf("GroupHeaderBand.Serialize error: %v", err)
	}
	if _, ok := w.written["SortOrder"]; !ok {
		t.Error("SortOrder should be written when set to SortOrderNone")
	}
}

// TestGroupHeaderBand_Serialize_OnlyCondition exercises only the Condition
// branch (SortOrder stays at the default Ascending, KeepTogether and
// ResetPageNumber stay false).
func TestGroupHeaderBand_Serialize_OnlyCondition(t *testing.T) {
	g := NewGroupHeaderBand()
	g.SetCondition("[Customer]")

	w := newFailAttrsWriter2()
	if err := g.Serialize(w); err != nil {
		t.Errorf("GroupHeaderBand.Serialize error: %v", err)
	}
	if _, ok := w.written["Condition"]; !ok {
		t.Error("Condition should be written")
	}
	for _, key := range []string{"SortOrder", "KeepTogether", "ResetPageNumber"} {
		if _, ok := w.written[key]; ok {
			t.Errorf("attribute %q should NOT be written at default", key)
		}
	}
}

// TestGroupHeaderBand_Serialize_OnlyResetPageNumber exercises only the
// ResetPageNumber branch, leaving all other optional attrs at defaults.
func TestGroupHeaderBand_Serialize_OnlyResetPageNumber(t *testing.T) {
	g := NewGroupHeaderBand()
	g.SetResetPageNumber(true)

	w := newFailAttrsWriter2()
	if err := g.Serialize(w); err != nil {
		t.Errorf("GroupHeaderBand.Serialize error: %v", err)
	}
	if _, ok := w.written["ResetPageNumber"]; !ok {
		t.Error("ResetPageNumber should be written")
	}
	for _, key := range []string{"Condition", "SortOrder", "KeepTogether"} {
		if _, ok := w.written[key]; ok {
			t.Errorf("attribute %q should NOT be written at default", key)
		}
	}
}

// TestGroupHeaderBand_Serialize_OnlyKeepTogether exercises only the
// KeepTogether branch.
func TestGroupHeaderBand_Serialize_OnlyKeepTogether(t *testing.T) {
	g := NewGroupHeaderBand()
	g.SetKeepTogether(true)

	w := newFailAttrsWriter2()
	if err := g.Serialize(w); err != nil {
		t.Errorf("GroupHeaderBand.Serialize error: %v", err)
	}
	if _, ok := w.written["KeepTogether"]; !ok {
		t.Error("KeepTogether should be written")
	}
}

// ─── GroupHeaderBand.Deserialize: all attribute combinations ─────────────────

// TestGroupHeaderBand_Deserialize_SortOrderNone exercises the SortOrderNone
// value during deserialization.
func TestGroupHeaderBand_Deserialize_SortOrderNone(t *testing.T) {
	r := &groupHeaderDeserMock{
		mockReader:      newMockReader(),
		condition:       "[Val]",
		sortOrder:       "None",
		keepTogether:    false,
		resetPageNumber: false,
	}
	g := NewGroupHeaderBand()
	if err := g.Deserialize(r); err != nil {
		t.Fatalf("GroupHeaderBand.Deserialize error: %v", err)
	}
	if g.SortOrder() != SortOrderNone {
		t.Errorf("SortOrder = %v, want SortOrderNone", g.SortOrder())
	}
}

// TestGroupHeaderBand_Deserialize_KeepTogetherOnly exercises only the
// KeepTogether=true branch.
func TestGroupHeaderBand_Deserialize_KeepTogetherOnly(t *testing.T) {
	r := &groupHeaderDeserMock{
		mockReader:      newMockReader(),
		sortOrder:       "Ascending",
		keepTogether:    true,
		resetPageNumber: false,
	}
	g := NewGroupHeaderBand()
	if err := g.Deserialize(r); err != nil {
		t.Fatalf("GroupHeaderBand.Deserialize error: %v", err)
	}
	if !g.KeepTogether() {
		t.Error("KeepTogether should be true")
	}
	if g.ResetPageNumber() {
		t.Error("ResetPageNumber should be false")
	}
}

// TestGroupHeaderBand_Deserialize_ResetPageNumberOnly exercises only the
// ResetPageNumber=true branch.
func TestGroupHeaderBand_Deserialize_ResetPageNumberOnly(t *testing.T) {
	r := &groupHeaderDeserMock{
		mockReader:      newMockReader(),
		sortOrder:       "Ascending",
		keepTogether:    false,
		resetPageNumber: true,
	}
	g := NewGroupHeaderBand()
	if err := g.Deserialize(r); err != nil {
		t.Fatalf("GroupHeaderBand.Deserialize error: %v", err)
	}
	if g.KeepTogether() {
		t.Error("KeepTogether should be false")
	}
	if !g.ResetPageNumber() {
		t.Error("ResetPageNumber should be true")
	}
}

// ─── DataBand.Serialize: Sort child-element branch ───────────────────────────

// TestDataBand_Serialize_SortASCOnly exercises the sort child-element path with
// a single ascending sort spec (no Expression). Sort is now written as a child
// <Sort> element via WriteObjectNamed, not as an attribute string.
func TestDataBand_Serialize_SortASCOnly(t *testing.T) {
	d := NewDataBand()
	d.AddSort(SortSpec{Column: "Price", Order: SortOrderAscending})

	w := newFailAttrsWriter2()
	if err := d.Serialize(w); err != nil {
		t.Errorf("DataBand.Serialize error: %v", err)
	}
	if _, ok := w.written["Sort"]; ok {
		t.Error("Sort should NOT be written as an attribute; it is a child element")
	}
}

// TestDataBand_Serialize_SortExpressionASC exercises the Expression path inside
// sortCollection.Serialize when Expression is non-empty (overrides Column).
func TestDataBand_Serialize_SortExpressionASC(t *testing.T) {
	d := NewDataBand()
	d.AddSort(SortSpec{Expression: "[Price]+[Tax]", Order: SortOrderAscending})

	w := newFailAttrsWriter2()
	if err := d.Serialize(w); err != nil {
		t.Errorf("DataBand.Serialize error: %v", err)
	}
	if _, ok := w.written["Sort"]; ok {
		t.Error("Sort should NOT be written as an attribute; it is a child element")
	}
}

// ─── DataBand.Deserialize: DataSource alias and filter ───────────────────────

// TestDataBand_Deserialize_DataSourceAlias exercises the DataSource alias read.
func TestDataBand_Deserialize_DataSourceAlias(t *testing.T) {
	r := &dataBandDeserMock{
		mockReader:      newMockReader(),
		dataSourceAlias: "Orders",
	}
	d := NewDataBand()
	if err := d.Deserialize(r); err != nil {
		t.Fatalf("DataBand.Deserialize error: %v", err)
	}
	if d.DataSourceAlias() != "Orders" {
		t.Errorf("DataSourceAlias = %q, want Orders", d.DataSourceAlias())
	}
}

// TestDataBand_Deserialize_FilterOnly exercises only the filter read path.
func TestDataBand_Deserialize_FilterOnly(t *testing.T) {
	r := &dataBandDeserMock{
		mockReader: newMockReader(),
		filter:     "[Qty] > 0",
	}
	d := NewDataBand()
	if err := d.Deserialize(r); err != nil {
		t.Fatalf("DataBand.Deserialize error: %v", err)
	}
	if d.Filter() != "[Qty] > 0" {
		t.Errorf("Filter = %q, want [Qty] > 0", d.Filter())
	}
}

// TestDataBand_Deserialize_SortSingleFieldNoDir exercises the sort parser with
// a single token (no direction token, so the default ASC is assumed).
func TestDataBand_Deserialize_SortSingleFieldNoDir(t *testing.T) {
	r := &dataBandDeserMock{
		mockReader: newMockReader(),
		sortStr:    "Name",
	}
	d := NewDataBand()
	if err := d.Deserialize(r); err != nil {
		t.Fatalf("DataBand.Deserialize error: %v", err)
	}
	if len(d.Sort()) != 1 {
		t.Fatalf("Sort len = %d, want 1", len(d.Sort()))
	}
	if d.Sort()[0].Column != "Name" {
		t.Errorf("Sort[0].Column = %q, want Name", d.Sort()[0].Column)
	}
	if d.Sort()[0].Order != SortOrderAscending {
		t.Errorf("Sort[0].Order = %v, want Ascending (default)", d.Sort()[0].Order)
	}
}

// TestDataBand_Deserialize_ColumnsCountZero exercises the branch where
// Columns.Count is zero (the `if n > 0` branch is NOT taken).
func TestDataBand_Deserialize_ColumnsCountZero(t *testing.T) {
	r := &dataBandDeserMock{
		mockReader:   newMockReader(),
		columnsCount: 0, // not > 0, so SetCount should NOT be called
	}
	d := NewDataBand()
	if err := d.Deserialize(r); err != nil {
		t.Fatalf("DataBand.Deserialize error: %v", err)
	}
	// NewBandColumns defaults to count=0; Deserialize should leave it at 0.
	if d.Columns().Count() != 0 {
		t.Errorf("Columns.Count = %d, want 0 (default)", d.Columns().Count())
	}
}

// TestDataBand_Deserialize_IndentNonZero exercises the Indent read path with a
// non-zero value.
func TestDataBand_Deserialize_IndentNonZero(t *testing.T) {
	r := &dataBandDeserMock{
		mockReader: newMockReader(),
		indent:     12.5,
	}
	d := NewDataBand()
	if err := d.Deserialize(r); err != nil {
		t.Fatalf("DataBand.Deserialize error: %v", err)
	}
	if d.Indent() != 12.5 {
		t.Errorf("Indent = %v, want 12.5", d.Indent())
	}
}

// TestDataBand_Deserialize_KeepSummaryOnly exercises the KeepSummary=true read.
func TestDataBand_Deserialize_KeepSummaryOnly(t *testing.T) {
	r := &dataBandDeserMock{
		mockReader:  newMockReader(),
		keepSummary: true,
	}
	d := NewDataBand()
	if err := d.Deserialize(r); err != nil {
		t.Fatalf("DataBand.Deserialize error: %v", err)
	}
	if !d.KeepSummary() {
		t.Error("KeepSummary should be true")
	}
}

// ─── sortOrderToString / sortOrderFromString round-trips ──────────────────────
//
// C# FastReport serialises GroupHeaderBand.SortOrder as an enum name string
// ("None", "Ascending", "Descending") via Converter.ToString (format "G").
// Real FRX files contain e.g. SortOrder="None" — verified in test-reports/.
// These tests cover every branch of the two helpers and confirm the fix for
// the previous integer-based serialisation bug.

func TestSortOrderToString_Ascending(t *testing.T) {
	if got := sortOrderToString(SortOrderAscending); got != "Ascending" {
		t.Errorf("sortOrderToString(Ascending) = %q, want Ascending", got)
	}
}

func TestSortOrderToString_Descending(t *testing.T) {
	if got := sortOrderToString(SortOrderDescending); got != "Descending" {
		t.Errorf("sortOrderToString(Descending) = %q, want Descending", got)
	}
}

func TestSortOrderToString_None(t *testing.T) {
	if got := sortOrderToString(SortOrderNone); got != "None" {
		t.Errorf("sortOrderToString(None) = %q, want None", got)
	}
}

func TestSortOrderToString_Unknown_DefaultsToAscending(t *testing.T) {
	// Any unrecognised value should produce "Ascending" (matches C# default).
	if got := sortOrderToString(SortOrder(99)); got != "Ascending" {
		t.Errorf("sortOrderToString(99) = %q, want Ascending", got)
	}
}

func TestSortOrderFromString_Ascending(t *testing.T) {
	if got := sortOrderFromString("Ascending"); got != SortOrderAscending {
		t.Errorf("sortOrderFromString(Ascending) = %v, want SortOrderAscending", got)
	}
}

func TestSortOrderFromString_Descending(t *testing.T) {
	if got := sortOrderFromString("Descending"); got != SortOrderDescending {
		t.Errorf("sortOrderFromString(Descending) = %v, want SortOrderDescending", got)
	}
}

func TestSortOrderFromString_None(t *testing.T) {
	if got := sortOrderFromString("None"); got != SortOrderNone {
		t.Errorf("sortOrderFromString(None) = %v, want SortOrderNone", got)
	}
}

func TestSortOrderFromString_Unknown_DefaultsToAscending(t *testing.T) {
	// An unrecognised string (e.g. old integer "1") defaults to Ascending.
	for _, s := range []string{"", "1", "2", "bogus"} {
		if got := sortOrderFromString(s); got != SortOrderAscending {
			t.Errorf("sortOrderFromString(%q) = %v, want SortOrderAscending", s, got)
		}
	}
}

// TestGroupHeaderBand_Serialize_SortOrderName verifies that GroupHeaderBand
// serialises SortOrder as a string name ("None"), not an integer, matching
// the C# FastReport FRX format (GroupHeaderBand.cs:361).
func TestGroupHeaderBand_Serialize_SortOrderName(t *testing.T) {
	g := NewGroupHeaderBand()
	g.SetSortOrder(SortOrderNone)

	w := newMockWriter()
	if err := g.Serialize(w); err != nil {
		t.Fatalf("Serialize error: %v", err)
	}
	// SortOrder must be written as a string name, not an integer.
	got, ok := w.written["SortOrder"]
	if !ok {
		t.Fatal("SortOrder attribute not written")
	}
	if got != "None" {
		t.Errorf("SortOrder = %q, want None (not an integer)", got)
	}
}

// TestGroupHeaderBand_Deserialize_SortOrderFromFRXString verifies that the
// SortOrder string from real FRX files (e.g. SortOrder="None") is correctly
// parsed. Previously the code used ReadInt which returned 0 (=Ascending) for
// any string value, silently misreading "None" as Ascending.
func TestGroupHeaderBand_Deserialize_SortOrderFromFRXString(t *testing.T) {
	for _, tc := range []struct {
		frxValue string
		want     SortOrder
	}{
		{"None", SortOrderNone},
		{"Ascending", SortOrderAscending},
		{"Descending", SortOrderDescending},
	} {
		r := &groupHeaderDeserMock{
			mockReader: newMockReader(),
			sortOrder:  tc.frxValue,
		}
		g := NewGroupHeaderBand()
		if err := g.Deserialize(r); err != nil {
			t.Fatalf("[%s] Deserialize error: %v", tc.frxValue, err)
		}
		if g.SortOrder() != tc.want {
			t.Errorf("[%s] SortOrder = %v, want %v", tc.frxValue, g.SortOrder(), tc.want)
		}
	}
}
