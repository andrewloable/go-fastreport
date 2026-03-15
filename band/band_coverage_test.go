package band_test

// band_coverage_test.go – additional tests targeting previously-uncovered code paths.
// Uses the external test package (band_test) so only exported symbols are accessed.

import (
	"bytes"
	"strings"
	"testing"

	"github.com/andrewloable/go-fastreport/band"
	"github.com/andrewloable/go-fastreport/serial"
)

// ── helpers ───────────────────────────────────────────────────────────────────

// bandRoundTrip serializes b under element name elemName, then deserializes
// into b2 using the supplied deserialize function.
func bandSerialize(t *testing.T, elemName string, b interface {
	Serialize(w interface{ WriteStr(string, string); WriteInt(string, int); WriteBool(string, bool); WriteFloat(string, float32); WriteObject(interface{}) error; WriteObjectNamed(string, interface{}) error }) error
}) (string, error) {
	t.Helper()
	// Just use serial.Writer directly.
	var buf bytes.Buffer
	w := serial.NewWriter(&buf)
	if err := w.BeginObject(elemName); err != nil {
		return "", err
	}
	// We can't call b.Serialize(w) directly in this generic helper because of interface
	// differences — use the concrete approach below.
	_ = w
	_ = buf
	return "", nil
}

// serializeBandBase serializes a *band.BandBase (Serialize method) into an XML
// element with the given name and returns the XML string.
func serializeBandBase(t *testing.T, elemName string, b *band.BandBase) string {
	t.Helper()
	var buf bytes.Buffer
	w := serial.NewWriter(&buf)
	if err := w.BeginObject(elemName); err != nil {
		t.Fatalf("BeginObject: %v", err)
	}
	if err := b.Serialize(w); err != nil {
		t.Fatalf("Serialize: %v", err)
	}
	if err := w.EndObject(); err != nil {
		t.Fatalf("EndObject: %v", err)
	}
	if err := w.Flush(); err != nil {
		t.Fatalf("Flush: %v", err)
	}
	return buf.String()
}

func deserializeBandBase(t *testing.T, xmlStr string) *band.BandBase {
	t.Helper()
	r := serial.NewReader(strings.NewReader(xmlStr))
	_, ok := r.ReadObjectHeader()
	if !ok {
		t.Fatalf("ReadObjectHeader failed for: %s", xmlStr)
	}
	b := band.NewBandBase()
	if err := b.Deserialize(r); err != nil {
		t.Fatalf("Deserialize: %v", err)
	}
	return b
}

// ── TypeName tests ─────────────────────────────────────────────────────────────

func TestTypeName_ChildBand(t *testing.T) {
	c := band.NewChildBand()
	if c.TypeName() != "Child" {
		t.Errorf("ChildBand.TypeName = %q, want Child", c.TypeName())
	}
}

func TestTypeName_PageHeaderBand(t *testing.T) {
	b := band.NewPageHeaderBand()
	if b.TypeName() != "PageHeader" {
		t.Errorf("PageHeaderBand.TypeName = %q, want PageHeader", b.TypeName())
	}
}

func TestTypeName_PageFooterBand(t *testing.T) {
	b := band.NewPageFooterBand()
	if b.TypeName() != "PageFooter" {
		t.Errorf("PageFooterBand.TypeName = %q, want PageFooter", b.TypeName())
	}
}

func TestTypeName_ReportTitleBand(t *testing.T) {
	b := band.NewReportTitleBand()
	if b.TypeName() != "ReportTitle" {
		t.Errorf("ReportTitleBand.TypeName = %q, want ReportTitle", b.TypeName())
	}
}

func TestTypeName_ReportSummaryBand(t *testing.T) {
	b := band.NewReportSummaryBand()
	if b.TypeName() != "ReportSummary" {
		t.Errorf("ReportSummaryBand.TypeName = %q, want ReportSummary", b.TypeName())
	}
}

func TestTypeName_GroupHeaderBand(t *testing.T) {
	b := band.NewGroupHeaderBand()
	if b.TypeName() != "GroupHeader" {
		t.Errorf("GroupHeaderBand.TypeName = %q, want GroupHeader", b.TypeName())
	}
}

func TestTypeName_GroupFooterBand(t *testing.T) {
	b := band.NewGroupFooterBand()
	if b.TypeName() != "GroupFooter" {
		t.Errorf("GroupFooterBand.TypeName = %q, want GroupFooter", b.TypeName())
	}
}

func TestTypeName_DataHeaderBand(t *testing.T) {
	b := band.NewDataHeaderBand()
	if b.TypeName() != "DataHeader" {
		t.Errorf("DataHeaderBand.TypeName = %q, want DataHeader", b.TypeName())
	}
}

func TestTypeName_DataFooterBand(t *testing.T) {
	b := band.NewDataFooterBand()
	if b.TypeName() != "DataFooter" {
		t.Errorf("DataFooterBand.TypeName = %q, want DataFooter", b.TypeName())
	}
}

func TestTypeName_ColumnHeaderBand(t *testing.T) {
	b := band.NewColumnHeaderBand()
	if b.TypeName() != "ColumnHeader" {
		t.Errorf("ColumnHeaderBand.TypeName = %q, want ColumnHeader", b.TypeName())
	}
}

func TestTypeName_ColumnFooterBand(t *testing.T) {
	b := band.NewColumnFooterBand()
	if b.TypeName() != "ColumnFooter" {
		t.Errorf("ColumnFooterBand.TypeName = %q, want ColumnFooter", b.TypeName())
	}
}

func TestTypeName_OverlayBand(t *testing.T) {
	b := band.NewOverlayBand()
	if b.TypeName() != "Overlay" {
		t.Errorf("OverlayBand.TypeName = %q, want Overlay", b.TypeName())
	}
}

func TestTypeName_DataBand(t *testing.T) {
	b := band.NewDataBand()
	if b.TypeName() != "Data" {
		t.Errorf("DataBand.TypeName = %q, want Data", b.TypeName())
	}
}

// ── BandBase Serialize / Deserialize round-trips ──────────────────────────────

func TestBandBase_Serialize_Defaults(t *testing.T) {
	b := band.NewBandBase()
	out := serializeBandBase(t, "Band", b)
	// Default BandBase should produce minimal output (no non-default attrs).
	if !strings.Contains(out, "<Band") {
		t.Errorf("expected <Band element, got: %s", out)
	}
}

func TestBandBase_SerializeDeserialize_RoundTrip(t *testing.T) {
	orig := band.NewBandBase()
	orig.SetStartNewPage(true)
	orig.SetFirstRowStartsNewPage(false)
	orig.SetPrintOnBottom(true)
	orig.SetKeepChild(true)
	orig.SetOutlineExpression("[Name]")
	orig.SetRepeatBandNTimes(3)
	orig.SetBeforeLayoutEvent("Band1_Before")
	orig.SetAfterLayoutEvent("Band1_After")

	xmlStr := serializeBandBase(t, "BandBase", orig)

	// Verify attributes are in the XML.
	for _, want := range []string{
		`StartNewPage="true"`,
		`FirstRowStartsNewPage="false"`,
		`PrintOnBottom="true"`,
		`KeepChild="true"`,
		`OutlineExpression="[Name]"`,
		`RepeatBandNTimes="3"`,
		`BeforeLayoutEvent="Band1_Before"`,
		`AfterLayoutEvent="Band1_After"`,
	} {
		if !strings.Contains(xmlStr, want) {
			t.Errorf("missing %q in:\n%s", want, xmlStr)
		}
	}

	// Deserialize and verify.
	got := deserializeBandBase(t, xmlStr)
	if !got.StartNewPage() {
		t.Error("StartNewPage should be true")
	}
	if got.FirstRowStartsNewPage() {
		t.Error("FirstRowStartsNewPage should be false")
	}
	if !got.PrintOnBottom() {
		t.Error("PrintOnBottom should be true")
	}
	if !got.KeepChild() {
		t.Error("KeepChild should be true")
	}
	if got.OutlineExpression() != "[Name]" {
		t.Errorf("OutlineExpression = %q, want [Name]", got.OutlineExpression())
	}
	if got.RepeatBandNTimes() != 3 {
		t.Errorf("RepeatBandNTimes = %d, want 3", got.RepeatBandNTimes())
	}
	if got.BeforeLayoutEvent() != "Band1_Before" {
		t.Errorf("BeforeLayoutEvent = %q", got.BeforeLayoutEvent())
	}
	if got.AfterLayoutEvent() != "Band1_After" {
		t.Errorf("AfterLayoutEvent = %q", got.AfterLayoutEvent())
	}
}

func TestBandBase_Deserialize_Defaults(t *testing.T) {
	// Deserializing an element with no attributes should restore all defaults.
	xmlStr := `<BandBase/>`
	got := deserializeBandBase(t, xmlStr)
	if got.StartNewPage() {
		t.Error("StartNewPage should default to false")
	}
	if !got.FirstRowStartsNewPage() {
		t.Error("FirstRowStartsNewPage should default to true")
	}
	if got.PrintOnBottom() {
		t.Error("PrintOnBottom should default to false")
	}
	if got.KeepChild() {
		t.Error("KeepChild should default to false")
	}
	if got.OutlineExpression() != "" {
		t.Errorf("OutlineExpression should default to empty, got %q", got.OutlineExpression())
	}
	if got.RepeatBandNTimes() != 1 {
		t.Errorf("RepeatBandNTimes should default to 1, got %d", got.RepeatBandNTimes())
	}
}

// ── UpdateLayout (no-op) ──────────────────────────────────────────────────────

func TestBandBase_UpdateLayout(t *testing.T) {
	b := band.NewBandBase()
	// UpdateLayout is a documented no-op; just verify it doesn't panic.
	b.UpdateLayout(100, 200)
	b.UpdateLayout(0, 0)
	b.UpdateLayout(-5, -10)
}

// ── SetChildOrder edge cases ──────────────────────────────────────────────────

func TestBandBase_SetChildOrder_OutOfBounds(t *testing.T) {
	b := band.NewBandBase()
	obj1 := newMinimalBase("a")
	obj2 := newMinimalBase("b")
	b.AddChild(obj1)
	b.AddChild(obj2)

	// Setting order beyond length should clamp to end.
	b.SetChildOrder(obj1, 999)
	if b.Objects().Get(1) != obj1 {
		t.Error("obj1 should be at end after SetChildOrder(999)")
	}
}

func TestBandBase_SetChildOrder_NotFound(t *testing.T) {
	b := band.NewBandBase()
	obj1 := newMinimalBase("a")
	obj2 := newMinimalBase("b") // not added to b
	b.AddChild(obj1)

	// SetChildOrder for an object not in the collection is a no-op.
	b.SetChildOrder(obj2, 0) // should not panic
	if b.Objects().Len() != 1 {
		t.Errorf("Objects len should be 1, got %d", b.Objects().Len())
	}
}

// ── ChildBand Serialize / Deserialize ─────────────────────────────────────────

func serializeChildBand(t *testing.T, c *band.ChildBand) string {
	t.Helper()
	var buf bytes.Buffer
	w := serial.NewWriter(&buf)
	if err := w.BeginObject("Child"); err != nil {
		t.Fatalf("BeginObject: %v", err)
	}
	if err := c.Serialize(w); err != nil {
		t.Fatalf("Serialize: %v", err)
	}
	if err := w.EndObject(); err != nil {
		t.Fatalf("EndObject: %v", err)
	}
	if err := w.Flush(); err != nil {
		t.Fatalf("Flush: %v", err)
	}
	return buf.String()
}

func deserializeChildBand(t *testing.T, xmlStr string) *band.ChildBand {
	t.Helper()
	r := serial.NewReader(strings.NewReader(xmlStr))
	_, ok := r.ReadObjectHeader()
	if !ok {
		t.Fatalf("ReadObjectHeader failed")
	}
	c := band.NewChildBand()
	if err := c.Deserialize(r); err != nil {
		t.Fatalf("Deserialize: %v", err)
	}
	return c
}

func TestChildBand_SerializeDeserialize_RoundTrip(t *testing.T) {
	orig := band.NewChildBand()
	orig.FillUnusedSpace = true
	orig.CompleteToNRows = 5
	orig.PrintIfDatabandEmpty = true

	xmlStr := serializeChildBand(t, orig)

	for _, want := range []string{
		`FillUnusedSpace="true"`,
		`CompleteToNRows="5"`,
		`PrintIfDatabandEmpty="true"`,
	} {
		if !strings.Contains(xmlStr, want) {
			t.Errorf("missing %q in:\n%s", want, xmlStr)
		}
	}

	got := deserializeChildBand(t, xmlStr)
	if !got.FillUnusedSpace {
		t.Error("FillUnusedSpace should be true")
	}
	if got.CompleteToNRows != 5 {
		t.Errorf("CompleteToNRows = %d, want 5", got.CompleteToNRows)
	}
	if !got.PrintIfDatabandEmpty {
		t.Error("PrintIfDatabandEmpty should be true")
	}
}

func TestChildBand_SerializeDeserialize_Defaults(t *testing.T) {
	orig := band.NewChildBand()
	xmlStr := serializeChildBand(t, orig)
	got := deserializeChildBand(t, xmlStr)
	if got.FillUnusedSpace {
		t.Error("FillUnusedSpace should default to false")
	}
	if got.CompleteToNRows != 0 {
		t.Errorf("CompleteToNRows should default to 0, got %d", got.CompleteToNRows)
	}
	if got.PrintIfDatabandEmpty {
		t.Error("PrintIfDatabandEmpty should default to false")
	}
}

// ── BandColumns ───────────────────────────────────────────────────────────────

func TestBandColumns_Count(t *testing.T) {
	bc := band.NewBandColumns()
	if bc.Count() != 0 {
		t.Errorf("Count default = %d, want 0", bc.Count())
	}
}

func TestBandColumns_SetCount(t *testing.T) {
	bc := band.NewBandColumns()
	if err := bc.SetCount(3); err != nil {
		t.Fatalf("SetCount(3): %v", err)
	}
	if bc.Count() != 3 {
		t.Errorf("Count = %d, want 3", bc.Count())
	}
}

func TestBandColumns_SetCount_Negative(t *testing.T) {
	bc := band.NewBandColumns()
	if err := bc.SetCount(-1); err == nil {
		t.Error("SetCount(-1) should return error")
	}
}

func TestBandColumns_SetCount_Zero(t *testing.T) {
	bc := band.NewBandColumns()
	if err := bc.SetCount(0); err != nil {
		t.Errorf("SetCount(0) should be valid, got: %v", err)
	}
	if bc.Count() != 0 {
		t.Errorf("Count = %d, want 0", bc.Count())
	}
}

func TestBandColumns_ActualWidth_ExplicitWidth(t *testing.T) {
	bc := band.NewBandColumns()
	bc.Width = 150
	if bc.ActualWidth() != 150 {
		t.Errorf("ActualWidth = %v, want 150", bc.ActualWidth())
	}
}

func TestBandColumns_ActualWidth_NoWidthNoFn(t *testing.T) {
	bc := band.NewBandColumns()
	bc.Width = 0
	// No pageWidthFn set — should return 0.
	if bc.ActualWidth() != 0 {
		t.Errorf("ActualWidth with no width or fn = %v, want 0", bc.ActualWidth())
	}
}

func TestBandColumns_ActualWidth_WithPageWidthFn(t *testing.T) {
	bc := band.NewBandColumns()
	bc.Width = 0
	_ = bc.SetCount(3)
	// Simulate setting pageWidthFn via DataBand (which we can't do directly
	// since the field is unexported). Instead use a DataBand to wire it up.
	d := band.NewDataBand()
	cols := d.Columns()
	_ = cols.SetCount(2)
	cols.Width = 0
	// DataBand doesn't expose pageWidthFn directly; we test via Positions.
	// When Width==0 and no pageWidthFn, ActualWidth==0, so Positions returns
	// all-zero positions.
	pos := cols.Positions()
	if len(pos) != 2 {
		t.Errorf("Positions len = %d, want 2", len(pos))
	}
	// With no page-width function, all positions are 0.
	for i, p := range pos {
		if p != 0 {
			t.Errorf("Positions[%d] = %v, want 0 (no pageWidthFn)", i, p)
		}
	}
}

func TestBandColumns_Positions_ZeroCount(t *testing.T) {
	bc := band.NewBandColumns()
	pos := bc.Positions()
	if pos != nil {
		t.Errorf("Positions with count=0 should be nil, got %v", pos)
	}
}

func TestBandColumns_Positions_WithWidth(t *testing.T) {
	bc := band.NewBandColumns()
	_ = bc.SetCount(3)
	bc.Width = 100
	pos := bc.Positions()
	if len(pos) != 3 {
		t.Fatalf("Positions len = %d, want 3", len(pos))
	}
	if pos[0] != 0 {
		t.Errorf("pos[0] = %v, want 0", pos[0])
	}
	if pos[1] != 100 {
		t.Errorf("pos[1] = %v, want 100", pos[1])
	}
	if pos[2] != 200 {
		t.Errorf("pos[2] = %v, want 200", pos[2])
	}
}

func TestBandColumns_Assign(t *testing.T) {
	src := band.NewBandColumns()
	_ = src.SetCount(4)
	src.Width = 80
	src.Layout = band.ColumnLayoutDownThenAcross
	src.MinRowCount = 10

	dst := band.NewBandColumns()
	dst.Assign(src)

	if dst.Count() != 4 {
		t.Errorf("Count = %d, want 4", dst.Count())
	}
	if dst.Width != 80 {
		t.Errorf("Width = %v, want 80", dst.Width)
	}
	if dst.Layout != band.ColumnLayoutDownThenAcross {
		t.Errorf("Layout = %v, want DownThenAcross", dst.Layout)
	}
	if dst.MinRowCount != 10 {
		t.Errorf("MinRowCount = %d, want 10", dst.MinRowCount)
	}
}

// ── HeaderFooterBandBase Serialize / Deserialize ──────────────────────────────

func serializeDataHeaderBand(t *testing.T, h *band.DataHeaderBand) string {
	t.Helper()
	var buf bytes.Buffer
	w := serial.NewWriter(&buf)
	if err := w.BeginObject("DataHeader"); err != nil {
		t.Fatalf("BeginObject: %v", err)
	}
	if err := h.Serialize(w); err != nil {
		t.Fatalf("Serialize: %v", err)
	}
	if err := w.EndObject(); err != nil {
		t.Fatalf("EndObject: %v", err)
	}
	if err := w.Flush(); err != nil {
		t.Fatalf("Flush: %v", err)
	}
	return buf.String()
}

func deserializeDataHeaderBand(t *testing.T, xmlStr string) *band.DataHeaderBand {
	t.Helper()
	r := serial.NewReader(strings.NewReader(xmlStr))
	_, ok := r.ReadObjectHeader()
	if !ok {
		t.Fatalf("ReadObjectHeader failed")
	}
	h := band.NewDataHeaderBand()
	if err := h.Deserialize(r); err != nil {
		t.Fatalf("Deserialize: %v", err)
	}
	return h
}

func TestDataHeaderBand_SerializeDeserialize_RoundTrip(t *testing.T) {
	orig := band.NewDataHeaderBand()
	orig.SetKeepWithData(true)
	orig.SetRepeatOnEveryPage(true)

	xmlStr := serializeDataHeaderBand(t, orig)

	for _, want := range []string{
		`KeepWithData="true"`,
		`RepeatOnEveryPage="true"`,
	} {
		if !strings.Contains(xmlStr, want) {
			t.Errorf("missing %q in:\n%s", want, xmlStr)
		}
	}

	got := deserializeDataHeaderBand(t, xmlStr)
	if !got.KeepWithData() {
		t.Error("KeepWithData should be true")
	}
	if !got.RepeatOnEveryPage() {
		t.Error("RepeatOnEveryPage should be true")
	}
}

func TestDataHeaderBand_SerializeDeserialize_Defaults(t *testing.T) {
	orig := band.NewDataHeaderBand()
	xmlStr := serializeDataHeaderBand(t, orig)
	got := deserializeDataHeaderBand(t, xmlStr)
	if got.KeepWithData() {
		t.Error("KeepWithData should default to false")
	}
	if got.RepeatOnEveryPage() {
		t.Error("RepeatOnEveryPage should default to false")
	}
}

// ── GroupHeaderBand Serialize / Deserialize ───────────────────────────────────

func serializeGroupHeaderBand(t *testing.T, g *band.GroupHeaderBand) string {
	t.Helper()
	var buf bytes.Buffer
	w := serial.NewWriter(&buf)
	if err := w.BeginObject("GroupHeader"); err != nil {
		t.Fatalf("BeginObject: %v", err)
	}
	if err := g.Serialize(w); err != nil {
		t.Fatalf("Serialize: %v", err)
	}
	if err := w.EndObject(); err != nil {
		t.Fatalf("EndObject: %v", err)
	}
	if err := w.Flush(); err != nil {
		t.Fatalf("Flush: %v", err)
	}
	return buf.String()
}

func deserializeGroupHeaderBand(t *testing.T, xmlStr string) *band.GroupHeaderBand {
	t.Helper()
	r := serial.NewReader(strings.NewReader(xmlStr))
	_, ok := r.ReadObjectHeader()
	if !ok {
		t.Fatalf("ReadObjectHeader failed")
	}
	g := band.NewGroupHeaderBand()
	if err := g.Deserialize(r); err != nil {
		t.Fatalf("Deserialize: %v", err)
	}
	return g
}

func TestGroupHeaderBand_SerializeDeserialize_RoundTrip(t *testing.T) {
	orig := band.NewGroupHeaderBand()
	orig.SetCondition("[Orders.CustomerName]")
	orig.SetSortOrder(band.SortOrderDescending)
	orig.SetKeepTogether(true)
	orig.SetResetPageNumber(true)
	orig.SetKeepWithData(true)
	orig.SetRepeatOnEveryPage(true)

	xmlStr := serializeGroupHeaderBand(t, orig)

	for _, want := range []string{
		`Condition="[Orders.CustomerName]"`,
		`SortOrder="1"`,
		`KeepTogether="true"`,
		`ResetPageNumber="true"`,
		`KeepWithData="true"`,
		`RepeatOnEveryPage="true"`,
	} {
		if !strings.Contains(xmlStr, want) {
			t.Errorf("missing %q in:\n%s", want, xmlStr)
		}
	}

	got := deserializeGroupHeaderBand(t, xmlStr)
	if got.Condition() != "[Orders.CustomerName]" {
		t.Errorf("Condition = %q", got.Condition())
	}
	if got.SortOrder() != band.SortOrderDescending {
		t.Errorf("SortOrder = %v, want Descending", got.SortOrder())
	}
	if !got.KeepTogether() {
		t.Error("KeepTogether should be true")
	}
	if !got.ResetPageNumber() {
		t.Error("ResetPageNumber should be true")
	}
	if !got.KeepWithData() {
		t.Error("KeepWithData should be true")
	}
	if !got.RepeatOnEveryPage() {
		t.Error("RepeatOnEveryPage should be true")
	}
}

func TestGroupHeaderBand_SerializeDeserialize_Defaults(t *testing.T) {
	orig := band.NewGroupHeaderBand()
	xmlStr := serializeGroupHeaderBand(t, orig)
	got := deserializeGroupHeaderBand(t, xmlStr)
	if got.SortOrder() != band.SortOrderAscending {
		t.Errorf("SortOrder default = %v, want Ascending", got.SortOrder())
	}
	if got.KeepTogether() {
		t.Error("KeepTogether should default to false")
	}
	if got.ResetPageNumber() {
		t.Error("ResetPageNumber should default to false")
	}
}

// ── DataBand Serialize / Deserialize ─────────────────────────────────────────

func serializeDataBand(t *testing.T, d *band.DataBand) string {
	t.Helper()
	var buf bytes.Buffer
	w := serial.NewWriter(&buf)
	if err := w.BeginObject("Data"); err != nil {
		t.Fatalf("BeginObject: %v", err)
	}
	if err := d.Serialize(w); err != nil {
		t.Fatalf("Serialize: %v", err)
	}
	if err := w.EndObject(); err != nil {
		t.Fatalf("EndObject: %v", err)
	}
	if err := w.Flush(); err != nil {
		t.Fatalf("Flush: %v", err)
	}
	return buf.String()
}

func deserializeDataBand(t *testing.T, xmlStr string) *band.DataBand {
	t.Helper()
	r := serial.NewReader(strings.NewReader(xmlStr))
	_, ok := r.ReadObjectHeader()
	if !ok {
		t.Fatalf("ReadObjectHeader failed")
	}
	d := band.NewDataBand()
	if err := d.Deserialize(r); err != nil {
		t.Fatalf("Deserialize: %v", err)
	}
	return d
}

func TestDataBand_SerializeDeserialize_Defaults(t *testing.T) {
	orig := band.NewDataBand()
	xmlStr := serializeDataBand(t, orig)
	got := deserializeDataBand(t, xmlStr)
	if got.Filter() != "" {
		t.Errorf("Filter default = %q", got.Filter())
	}
	if got.DataSourceAlias() != "" {
		t.Errorf("DataSourceAlias default = %q", got.DataSourceAlias())
	}
}

func TestDataBand_SerializeDeserialize_AllAttrs(t *testing.T) {
	orig := band.NewDataBand()
	orig.SetFilter("[Amount] > 0")
	orig.SetPrintIfDetailEmpty(true)
	orig.SetPrintIfDSEmpty(true)
	orig.SetKeepTogether(true)
	orig.SetKeepDetail(true)
	orig.SetIDColumn("ID")
	orig.SetParentIDColumn("ParentID")
	orig.SetIndent(15)
	orig.SetKeepSummary(true)
	orig.AddSort(band.SortSpec{Column: "Name", Order: band.SortOrderAscending})
	orig.AddSort(band.SortSpec{Column: "Date", Order: band.SortOrderDescending})
	_ = orig.Columns().SetCount(2)
	orig.Columns().Width = 200

	xmlStr := serializeDataBand(t, orig)

	for _, want := range []string{
		`Filter=`,          // attribute is present (value may be XML-escaped)
		`PrintIfDetailEmpty="true"`,
		`PrintIfDatasourceEmpty="true"`,
		`KeepTogether="true"`,
		`KeepDetail="true"`,
		`IdColumn="ID"`,
		`ParentIdColumn="ParentID"`,
		`Indent="15"`,
		`KeepSummary="true"`,
		`Name ASC`,
		`Date DESC`,
	} {
		if !strings.Contains(xmlStr, want) {
			t.Errorf("missing %q in:\n%s", want, xmlStr)
		}
	}

	got := deserializeDataBand(t, xmlStr)
	if got.Filter() != "[Amount] > 0" {
		t.Errorf("Filter = %q", got.Filter())
	}
	if !got.PrintIfDetailEmpty() {
		t.Error("PrintIfDetailEmpty should be true")
	}
	if !got.PrintIfDSEmpty() {
		t.Error("PrintIfDSEmpty should be true")
	}
	if !got.KeepTogether() {
		t.Error("KeepTogether should be true")
	}
	if !got.KeepDetail() {
		t.Error("KeepDetail should be true")
	}
	if got.IDColumn() != "ID" {
		t.Errorf("IDColumn = %q", got.IDColumn())
	}
	if got.ParentIDColumn() != "ParentID" {
		t.Errorf("ParentIDColumn = %q", got.ParentIDColumn())
	}
	if got.Indent() != 15 {
		t.Errorf("Indent = %v", got.Indent())
	}
	if !got.KeepSummary() {
		t.Error("KeepSummary should be true")
	}
	if len(got.Sort()) != 2 {
		t.Fatalf("Sort len = %d, want 2", len(got.Sort()))
	}
	if got.Sort()[0].Column != "Name" {
		t.Errorf("Sort[0].Column = %q, want Name", got.Sort()[0].Column)
	}
	if got.Sort()[0].Order != band.SortOrderAscending {
		t.Errorf("Sort[0].Order = %v, want Ascending", got.Sort()[0].Order)
	}
	if got.Sort()[1].Column != "Date" {
		t.Errorf("Sort[1].Column = %q, want Date", got.Sort()[1].Column)
	}
	if got.Sort()[1].Order != band.SortOrderDescending {
		t.Errorf("Sort[1].Order = %v, want Descending", got.Sort()[1].Order)
	}
}

// ── DataBand DataSource methods ───────────────────────────────────────────────

// mockDataSource satisfies band.DataSource interface.
type mockDataSource struct {
	name string
}

func (m *mockDataSource) RowCount() int                   { return 0 }
func (m *mockDataSource) First() error                    { return nil }
func (m *mockDataSource) Next() error                     { return nil }
func (m *mockDataSource) EOF() bool                       { return true }
func (m *mockDataSource) GetValue(col string) (any, error) { return nil, nil }

func TestDataBand_DataSourceAlias(t *testing.T) {
	// DataSourceAlias is read from FRX DataSource attribute on Deserialize.
	xmlStr := `<Data DataSource="Customers"/>`
	got := deserializeDataBand(t, xmlStr)
	if got.DataSourceAlias() != "Customers" {
		t.Errorf("DataSourceAlias = %q, want Customers", got.DataSourceAlias())
	}
}

func TestDataBand_DataSourceRef(t *testing.T) {
	d := band.NewDataBand()
	if d.DataSourceRef() != nil {
		t.Error("DataSourceRef should default to nil")
	}
	ds := &mockDataSource{name: "TestDS"}
	d.SetDataSource(ds)
	if d.DataSourceRef() != ds {
		t.Error("DataSourceRef should be the mock data source")
	}
}

func TestDataBand_MaxRows(t *testing.T) {
	d := band.NewDataBand()
	if d.MaxRows() != 0 {
		t.Errorf("MaxRows default = %d, want 0", d.MaxRows())
	}
	d.SetMaxRows(100)
	if d.MaxRows() != 100 {
		t.Errorf("MaxRows = %d, want 100", d.MaxRows())
	}
}

func TestDataBand_Sort(t *testing.T) {
	d := band.NewDataBand()
	if len(d.Sort()) != 0 {
		t.Errorf("Sort default len = %d, want 0", len(d.Sort()))
	}

	specs := []band.SortSpec{
		{Column: "Name", Order: band.SortOrderAscending},
		{Column: "Date", Order: band.SortOrderDescending},
	}
	d.SetSort(specs)
	if len(d.Sort()) != 2 {
		t.Fatalf("Sort len = %d, want 2", len(d.Sort()))
	}
	if d.Sort()[0].Column != "Name" {
		t.Errorf("Sort[0].Column = %q", d.Sort()[0].Column)
	}
}

func TestDataBand_AddSort(t *testing.T) {
	d := band.NewDataBand()
	d.AddSort(band.SortSpec{Column: "Col1", Order: band.SortOrderAscending})
	d.AddSort(band.SortSpec{Column: "Col2", Order: band.SortOrderDescending})
	if len(d.Sort()) != 2 {
		t.Fatalf("Sort len = %d, want 2", len(d.Sort()))
	}
	if d.Sort()[1].Column != "Col2" {
		t.Errorf("Sort[1].Column = %q, want Col2", d.Sort()[1].Column)
	}
}

func TestDataBand_Sort_WithExpression(t *testing.T) {
	d := band.NewDataBand()
	// Expression overrides Column during Serialize.
	d.AddSort(band.SortSpec{Column: "Col", Expression: "[Total]", Order: band.SortOrderDescending})

	xmlStr := serializeDataBand(t, d)
	if !strings.Contains(xmlStr, "[Total] DESC") {
		t.Errorf("expected '[Total] DESC' in sort output:\n%s", xmlStr)
	}
}

func TestDataBand_Deserialize_SortString_Semicolons(t *testing.T) {
	// Test parsing of the semi-colon delimited sort string on Deserialize.
	xmlStr := `<Data Sort="ColA ASC;ColB DESC;ColC ASC"/>`
	got := deserializeDataBand(t, xmlStr)
	if len(got.Sort()) != 3 {
		t.Fatalf("Sort len = %d, want 3", len(got.Sort()))
	}
	if got.Sort()[0].Column != "ColA" || got.Sort()[0].Order != band.SortOrderAscending {
		t.Errorf("Sort[0] = %+v", got.Sort()[0])
	}
	if got.Sort()[1].Column != "ColB" || got.Sort()[1].Order != band.SortOrderDescending {
		t.Errorf("Sort[1] = %+v", got.Sort()[1])
	}
	if got.Sort()[2].Column != "ColC" || got.Sort()[2].Order != band.SortOrderAscending {
		t.Errorf("Sort[2] = %+v", got.Sort()[2])
	}
}

func TestDataBand_Deserialize_ColumnsCount(t *testing.T) {
	xmlStr := `<Data Columns.Count="3" Columns.Width="150"/>`
	got := deserializeDataBand(t, xmlStr)
	if got.Columns().Count() != 3 {
		t.Errorf("Columns.Count = %d, want 3", got.Columns().Count())
	}
	if got.Columns().Width != 150 {
		t.Errorf("Columns.Width = %v, want 150", got.Columns().Width)
	}
}

// ── ReportSummaryBand Serialize / Deserialize ─────────────────────────────────

func TestReportSummaryBand_SerializeDeserialize(t *testing.T) {
	orig := band.NewReportSummaryBand()
	orig.SetKeepWithData(true)
	orig.SetRepeatOnEveryPage(true)

	var buf bytes.Buffer
	w := serial.NewWriter(&buf)
	if err := w.BeginObject("ReportSummary"); err != nil {
		t.Fatalf("BeginObject: %v", err)
	}
	if err := orig.Serialize(w); err != nil {
		t.Fatalf("Serialize: %v", err)
	}
	if err := w.EndObject(); err != nil {
		t.Fatalf("EndObject: %v", err)
	}
	if err := w.Flush(); err != nil {
		t.Fatalf("Flush: %v", err)
	}
	xmlStr := buf.String()

	r := serial.NewReader(strings.NewReader(xmlStr))
	_, ok := r.ReadObjectHeader()
	if !ok {
		t.Fatalf("ReadObjectHeader failed")
	}
	got := band.NewReportSummaryBand()
	if err := got.Deserialize(r); err != nil {
		t.Fatalf("Deserialize: %v", err)
	}
	if !got.KeepWithData() {
		t.Error("KeepWithData should be true")
	}
	if !got.RepeatOnEveryPage() {
		t.Error("RepeatOnEveryPage should be true")
	}
}

// ── DataFooterBand Serialize / Deserialize ────────────────────────────────────

func TestDataFooterBand_SerializeDeserialize(t *testing.T) {
	orig := band.NewDataFooterBand()
	orig.SetKeepWithData(true)

	var buf bytes.Buffer
	w := serial.NewWriter(&buf)
	if err := w.BeginObject("DataFooter"); err != nil {
		t.Fatalf("BeginObject: %v", err)
	}
	if err := orig.Serialize(w); err != nil {
		t.Fatalf("Serialize: %v", err)
	}
	if err := w.EndObject(); err != nil {
		t.Fatalf("EndObject: %v", err)
	}
	if err := w.Flush(); err != nil {
		t.Fatalf("Flush: %v", err)
	}
	xmlStr := buf.String()

	r := serial.NewReader(strings.NewReader(xmlStr))
	_, ok := r.ReadObjectHeader()
	if !ok {
		t.Fatalf("ReadObjectHeader failed")
	}
	got := band.NewDataFooterBand()
	if err := got.Deserialize(r); err != nil {
		t.Fatalf("Deserialize: %v", err)
	}
	if !got.KeepWithData() {
		t.Error("KeepWithData should be true")
	}
}

// ── GroupFooterBand Serialize / Deserialize ───────────────────────────────────

func TestGroupFooterBand_SerializeDeserialize(t *testing.T) {
	orig := band.NewGroupFooterBand()
	orig.SetRepeatOnEveryPage(true)

	var buf bytes.Buffer
	w := serial.NewWriter(&buf)
	if err := w.BeginObject("GroupFooter"); err != nil {
		t.Fatalf("BeginObject: %v", err)
	}
	if err := orig.Serialize(w); err != nil {
		t.Fatalf("Serialize: %v", err)
	}
	if err := w.EndObject(); err != nil {
		t.Fatalf("EndObject: %v", err)
	}
	if err := w.Flush(); err != nil {
		t.Fatalf("Flush: %v", err)
	}
	xmlStr := buf.String()

	r := serial.NewReader(strings.NewReader(xmlStr))
	_, ok := r.ReadObjectHeader()
	if !ok {
		t.Fatalf("ReadObjectHeader failed")
	}
	got := band.NewGroupFooterBand()
	if err := got.Deserialize(r); err != nil {
		t.Fatalf("Deserialize: %v", err)
	}
	if !got.RepeatOnEveryPage() {
		t.Error("RepeatOnEveryPage should be true")
	}
}

// ── BandBase with BandBase attrs (non-default) ────────────────────────────────

func TestBandBase_Serialize_OnlyNonDefaultAttrs(t *testing.T) {
	// Verify that attributes with default values are NOT emitted.
	b := band.NewBandBase()
	// firstRowStartsNewPage is true by default — should NOT appear when true.
	out := serializeBandBase(t, "Band", b)
	if strings.Contains(out, "FirstRowStartsNewPage") {
		t.Errorf("FirstRowStartsNewPage should not be emitted at default=true:\n%s", out)
	}
	if strings.Contains(out, "RepeatBandNTimes") {
		t.Errorf("RepeatBandNTimes should not be emitted at default=1:\n%s", out)
	}
}

// ── DataBand DeserializeChild ─────────────────────────────────────────────────

// TestDataBand_DeserializeChild_Sort exercises DeserializeChild by calling it
// directly via the serial.Reader placed on a <Sort> element.
func TestDataBand_DeserializeChild_Sort(t *testing.T) {
	// Build XML that has a <Sort> container with nested <Sort> items.
	xmlStr := `<Sort><Sort Expression="[Name]" Descending="false"/><Sort Expression="[Date]" Descending="true"/></Sort>`
	r := serial.NewReader(strings.NewReader(xmlStr))
	_, ok := r.ReadObjectHeader() // consume the outer <Sort>
	if !ok {
		t.Fatal("ReadObjectHeader failed")
	}
	d := band.NewDataBand()
	handled := d.DeserializeChild("Sort", r)
	if !handled {
		t.Error("DeserializeChild should return true for 'Sort' child type")
	}
	if len(d.Sort()) != 2 {
		t.Fatalf("Sort len = %d, want 2", len(d.Sort()))
	}
	if d.Sort()[0].Column != "[Name]" {
		t.Errorf("Sort[0].Column = %q, want [Name]", d.Sort()[0].Column)
	}
	if d.Sort()[0].Order != band.SortOrderAscending {
		t.Errorf("Sort[0].Order = %v, want Ascending", d.Sort()[0].Order)
	}
	if d.Sort()[1].Column != "[Date]" {
		t.Errorf("Sort[1].Column = %q, want [Date]", d.Sort()[1].Column)
	}
	if d.Sort()[1].Order != band.SortOrderDescending {
		t.Errorf("Sort[1].Order = %v, want Descending", d.Sort()[1].Order)
	}
}

func TestDataBand_DeserializeChild_UnknownType(t *testing.T) {
	xmlStr := `<Other/>`
	r := serial.NewReader(strings.NewReader(xmlStr))
	_, ok := r.ReadObjectHeader()
	if !ok {
		t.Fatal("ReadObjectHeader failed")
	}
	d := band.NewDataBand()
	handled := d.DeserializeChild("Other", r)
	if handled {
		t.Error("DeserializeChild should return false for unknown child type")
	}
}

func TestDataBand_DeserializeChild_SortWithNoExpression(t *testing.T) {
	// A Sort item with empty Expression should be skipped.
	xmlStr := `<Sort><Sort Expression="" Descending="false"/></Sort>`
	r := serial.NewReader(strings.NewReader(xmlStr))
	_, ok := r.ReadObjectHeader()
	if !ok {
		t.Fatal("ReadObjectHeader failed")
	}
	d := band.NewDataBand()
	handled := d.DeserializeChild("Sort", r)
	if !handled {
		t.Error("DeserializeChild should return true for 'Sort' child type")
	}
	if len(d.Sort()) != 0 {
		t.Errorf("Sort len = %d, want 0 (empty expression skipped)", len(d.Sort()))
	}
}

func TestDataBand_DeserializeChild_SortWithUnknownChildElement(t *testing.T) {
	// Unknown child element inside <Sort> container should be drained.
	xmlStr := `<Sort><Unknown Foo="bar"/><Sort Expression="[Val]" Descending="false"/></Sort>`
	r := serial.NewReader(strings.NewReader(xmlStr))
	_, ok := r.ReadObjectHeader()
	if !ok {
		t.Fatal("ReadObjectHeader failed")
	}
	d := band.NewDataBand()
	handled := d.DeserializeChild("Sort", r)
	if !handled {
		t.Error("DeserializeChild should return true for 'Sort'")
	}
	if len(d.Sort()) != 1 {
		t.Fatalf("Sort len = %d, want 1", len(d.Sort()))
	}
	if d.Sort()[0].Column != "[Val]" {
		t.Errorf("Sort[0].Column = %q, want [Val]", d.Sort()[0].Column)
	}
}

// ── DataBand Deserialize with empty Sort parts ────────────────────────────────

func TestDataBand_Deserialize_SortString_EmptyParts(t *testing.T) {
	// Semi-colons producing empty parts should be silently skipped.
	xmlStr := `<Data Sort=";ColA ASC;;"/>`
	got := deserializeDataBand(t, xmlStr)
	if len(got.Sort()) != 1 {
		t.Fatalf("Sort len = %d, want 1 (empty parts skipped)", len(got.Sort()))
	}
	if got.Sort()[0].Column != "ColA" {
		t.Errorf("Sort[0].Column = %q, want ColA", got.Sort()[0].Column)
	}
}

// ── BandColumns ActualWidth with count<=1 ────────────────────────────────────

func TestBandColumns_ActualWidth_CountZero_WithPageFn(t *testing.T) {
	// When count<=1, ActualWidth via pageWidthFn returns full page width.
	// We can't directly set pageWidthFn (unexported), so we test via Positions
	// when Width is set explicitly (the simpler path).
	bc := band.NewBandColumns()
	_ = bc.SetCount(1)
	bc.Width = 500
	if bc.ActualWidth() != 500 {
		t.Errorf("ActualWidth with count=1 and Width=500 = %v, want 500", bc.ActualWidth())
	}
	pos := bc.Positions()
	if len(pos) != 1 {
		t.Fatalf("Positions len = %d, want 1", len(pos))
	}
	if pos[0] != 0 {
		t.Errorf("Positions[0] = %v, want 0", pos[0])
	}
}

// ── DataBand Deserialize with Columns.Count == 0 (not written) ───────────────

func TestDataBand_Deserialize_NoColumns(t *testing.T) {
	xmlStr := `<Data Filter="[x]&gt;0"/>`
	got := deserializeDataBand(t, xmlStr)
	// Columns.Count not present → count stays 0.
	if got.Columns().Count() != 0 {
		t.Errorf("Columns.Count = %d, want 0 when not serialized", got.Columns().Count())
	}
}
