package band_test

// band_porting_gaps_test.go – tests for methods implemented as part of the
// BandBase.cs / BreakableComponent.cs / ChildBand.cs porting-gaps review.
//
// Covers:
//   - BandBase.IsEmpty()
//   - BandBase.GetExpressions()
//   - BandBase.Assign()
//   - BandBase.IsColumnDependentBand() (base returns false)
//   - IsColumnDependentBand() overrides on concrete column-dependent types
//   - ChildBand.GetTopParentBand()
//   - ChildBand.IsColumnDependentBand()
//   - ChildBand.Assign()
//   - report.BreakableComponent.Assign()

import (
	"testing"

	"github.com/andrewloable/go-fastreport/band"
	"github.com/andrewloable/go-fastreport/report"
)

// ── BandBase.IsEmpty ──────────────────────────────────────────────────────────

// TestBandBase_IsEmpty_NoObjects verifies that a band with no objects is empty.
// Mirrors C# BandBase.IsEmpty() default (BandBase.cs line 926-929).
func TestBandBase_IsEmpty_NoObjects(t *testing.T) {
	b := band.NewBandBase()
	if !b.IsEmpty() {
		t.Error("BandBase with no objects should be empty")
	}
}

// TestBandBase_IsEmpty_WithObject verifies that adding an object makes the band non-empty.
func TestBandBase_IsEmpty_WithObject(t *testing.T) {
	b := band.NewBandBase()
	// Add a minimal child object.
	child := newMinimalBase("obj1")
	b.AddChild(child)
	if b.IsEmpty() {
		t.Error("BandBase with one object should NOT be empty")
	}
}

// TestChildBand_IsEmpty_NoObjects confirms ChildBand inherits IsEmpty correctly.
func TestChildBand_IsEmpty_NoObjects(t *testing.T) {
	c := band.NewChildBand()
	if !c.IsEmpty() {
		t.Error("ChildBand with no objects should be empty")
	}
}

// TestDataBand_IsEmpty_NoObjects confirms DataBand IsEmpty on empty band.
func TestDataBand_IsEmpty_NoObjects(t *testing.T) {
	d := band.NewDataBand()
	if !d.IsEmpty() {
		t.Error("DataBand with no objects should be empty")
	}
}

// TestDataBand_IsEmpty_WithObject confirms DataBand IsEmpty false when objects present.
func TestDataBand_IsEmpty_WithObject(t *testing.T) {
	d := band.NewDataBand()
	child := newMinimalBase("obj1")
	d.AddChild(child)
	if d.IsEmpty() {
		t.Error("DataBand with one object should NOT be empty")
	}
}

// ── BandBase.GetExpressions ───────────────────────────────────────────────────

// TestBandBase_GetExpressions_Empty verifies that a band with no outline
// expression returns an empty slice.
// Mirrors C# BandBase.GetExpressions() (BandBase.cs line 606-615).
func TestBandBase_GetExpressions_Empty(t *testing.T) {
	b := band.NewBandBase()
	exprs := b.GetExpressions()
	if len(exprs) != 0 {
		t.Errorf("GetExpressions with empty OutlineExpression: got %d items, want 0", len(exprs))
	}
}

// TestBandBase_GetExpressions_WithOutline verifies that OutlineExpression is included.
func TestBandBase_GetExpressions_WithOutline(t *testing.T) {
	b := band.NewBandBase()
	b.SetOutlineExpression("[Orders.CustomerName]")
	exprs := b.GetExpressions()
	if len(exprs) != 1 {
		t.Fatalf("GetExpressions: got %d items, want 1", len(exprs))
	}
	if exprs[0] != "[Orders.CustomerName]" {
		t.Errorf("GetExpressions[0] = %q, want [Orders.CustomerName]", exprs[0])
	}
}

// TestBandBase_GetExpressions_MultipleCallsConsistent confirms idempotency.
func TestBandBase_GetExpressions_MultipleCallsConsistent(t *testing.T) {
	b := band.NewBandBase()
	b.SetOutlineExpression("[Expr]")
	e1 := b.GetExpressions()
	e2 := b.GetExpressions()
	if len(e1) != len(e2) || (len(e1) > 0 && e1[0] != e2[0]) {
		t.Error("GetExpressions should return consistent results across multiple calls")
	}
}

// ── BandBase.Assign ───────────────────────────────────────────────────────────

// TestBandBase_Assign_CopiesAllProperties verifies all BandBase fields are copied.
// Mirrors C# BandBase.Assign(Base source) (BandBase.cs line 514-529).
func TestBandBase_Assign_CopiesAllProperties(t *testing.T) {
	src := band.NewBandBase()
	src.SetStartNewPage(true)
	src.SetFirstRowStartsNewPage(false)
	src.SetPrintOnBottom(true)
	src.SetKeepChild(true)
	src.SetOutlineExpression("[Outline]")
	src.SetBeforeLayoutEvent("OnBefore")
	src.SetAfterLayoutEvent("OnAfter")
	src.SetRepeatBandNTimes(3)
	src.SetIsLastRow(true)
	src.AddGuide(10.0)
	src.AddGuide(20.0)

	dst := band.NewBandBase()
	dst.Assign(src)

	if !dst.StartNewPage() {
		t.Error("Assign: StartNewPage not copied")
	}
	if dst.FirstRowStartsNewPage() {
		t.Error("Assign: FirstRowStartsNewPage not copied (expected false)")
	}
	if !dst.PrintOnBottom() {
		t.Error("Assign: PrintOnBottom not copied")
	}
	if !dst.KeepChild() {
		t.Error("Assign: KeepChild not copied")
	}
	if dst.OutlineExpression() != "[Outline]" {
		t.Errorf("Assign: OutlineExpression = %q, want [Outline]", dst.OutlineExpression())
	}
	if dst.BeforeLayoutEvent() != "OnBefore" {
		t.Errorf("Assign: BeforeLayoutEvent = %q, want OnBefore", dst.BeforeLayoutEvent())
	}
	if dst.AfterLayoutEvent() != "OnAfter" {
		t.Errorf("Assign: AfterLayoutEvent = %q, want OnAfter", dst.AfterLayoutEvent())
	}
	if dst.RepeatBandNTimes() != 3 {
		t.Errorf("Assign: RepeatBandNTimes = %d, want 3", dst.RepeatBandNTimes())
	}
	if !dst.IsLastRow() {
		t.Error("Assign: IsLastRow not copied")
	}
	if len(dst.Guides()) != 2 {
		t.Errorf("Assign: Guides len = %d, want 2", len(dst.Guides()))
	}
	if dst.Guides()[0] != 10.0 || dst.Guides()[1] != 20.0 {
		t.Errorf("Assign: Guides = %v, want [10 20]", dst.Guides())
	}
}

// TestBandBase_Assign_DeepCopiesGuides verifies that modifying src.Guides after
// Assign does not affect dst.Guides.
func TestBandBase_Assign_DeepCopiesGuides(t *testing.T) {
	src := band.NewBandBase()
	src.AddGuide(5.0)

	dst := band.NewBandBase()
	dst.Assign(src)

	// Mutate src by appending a new guide.
	src.AddGuide(99.0)

	if len(dst.Guides()) != 1 {
		t.Errorf("Assign: Guides should have been deep-copied; dst has %d guides", len(dst.Guides()))
	}
}

// TestBandBase_Assign_NilSource verifies that Assign(nil) is a no-op.
func TestBandBase_Assign_NilSource(t *testing.T) {
	dst := band.NewBandBase()
	dst.SetStartNewPage(true)
	dst.Assign(nil) // must not panic
	if !dst.StartNewPage() {
		t.Error("Assign(nil) should not modify dst")
	}
}

// ── IsColumnDependentBand on concrete types ───────────────────────────────────

// TestIsColumnDependentBand_BandBase_False verifies BandBase returns false.
func TestIsColumnDependentBand_BandBase_False(t *testing.T) {
	b := band.NewBandBase()
	if b.IsColumnDependentBand() {
		t.Error("BandBase.IsColumnDependentBand should return false (not column-dependent)")
	}
}

// TestIsColumnDependentBand_PageHeaderBand_False verifies PageHeaderBand returns false.
func TestIsColumnDependentBand_PageHeaderBand_False(t *testing.T) {
	b := band.NewPageHeaderBand()
	if b.IsColumnDependentBand() {
		t.Error("PageHeaderBand.IsColumnDependentBand should return false")
	}
}

// TestIsColumnDependentBand_PageFooterBand_False verifies PageFooterBand returns false.
func TestIsColumnDependentBand_PageFooterBand_False(t *testing.T) {
	b := band.NewPageFooterBand()
	if b.IsColumnDependentBand() {
		t.Error("PageFooterBand.IsColumnDependentBand should return false")
	}
}

// TestIsColumnDependentBand_ReportTitleBand_False verifies ReportTitleBand returns false.
func TestIsColumnDependentBand_ReportTitleBand_False(t *testing.T) {
	b := band.NewReportTitleBand()
	if b.IsColumnDependentBand() {
		t.Error("ReportTitleBand.IsColumnDependentBand should return false")
	}
}

// TestIsColumnDependentBand_OverlayBand_False verifies OverlayBand returns false.
func TestIsColumnDependentBand_OverlayBand_False(t *testing.T) {
	b := band.NewOverlayBand()
	if b.IsColumnDependentBand() {
		t.Error("OverlayBand.IsColumnDependentBand should return false")
	}
}

// TestIsColumnDependentBand_DataBand_True verifies DataBand returns true.
// Mirrors C# BandBase.IsColumnDependentBand (BandBase.cs line 589).
func TestIsColumnDependentBand_DataBand_True(t *testing.T) {
	d := band.NewDataBand()
	if !d.IsColumnDependentBand() {
		t.Error("DataBand.IsColumnDependentBand should return true")
	}
}

// TestIsColumnDependentBand_DataHeaderBand_True verifies DataHeaderBand returns true.
func TestIsColumnDependentBand_DataHeaderBand_True(t *testing.T) {
	d := band.NewDataHeaderBand()
	if !d.IsColumnDependentBand() {
		t.Error("DataHeaderBand.IsColumnDependentBand should return true")
	}
}

// TestIsColumnDependentBand_DataFooterBand_True verifies DataFooterBand returns true.
func TestIsColumnDependentBand_DataFooterBand_True(t *testing.T) {
	d := band.NewDataFooterBand()
	if !d.IsColumnDependentBand() {
		t.Error("DataFooterBand.IsColumnDependentBand should return true")
	}
}

// TestIsColumnDependentBand_GroupHeaderBand_True verifies GroupHeaderBand returns true.
func TestIsColumnDependentBand_GroupHeaderBand_True(t *testing.T) {
	g := band.NewGroupHeaderBand()
	if !g.IsColumnDependentBand() {
		t.Error("GroupHeaderBand.IsColumnDependentBand should return true")
	}
}

// TestIsColumnDependentBand_GroupFooterBand_True verifies GroupFooterBand returns true.
func TestIsColumnDependentBand_GroupFooterBand_True(t *testing.T) {
	g := band.NewGroupFooterBand()
	if !g.IsColumnDependentBand() {
		t.Error("GroupFooterBand.IsColumnDependentBand should return true")
	}
}

// TestIsColumnDependentBand_ColumnHeaderBand_True verifies ColumnHeaderBand returns true.
func TestIsColumnDependentBand_ColumnHeaderBand_True(t *testing.T) {
	b := band.NewColumnHeaderBand()
	if !b.IsColumnDependentBand() {
		t.Error("ColumnHeaderBand.IsColumnDependentBand should return true")
	}
}

// TestIsColumnDependentBand_ColumnFooterBand_True verifies ColumnFooterBand returns true.
func TestIsColumnDependentBand_ColumnFooterBand_True(t *testing.T) {
	b := band.NewColumnFooterBand()
	if !b.IsColumnDependentBand() {
		t.Error("ColumnFooterBand.IsColumnDependentBand should return true")
	}
}

// TestIsColumnDependentBand_ReportSummaryBand_True verifies ReportSummaryBand returns true.
func TestIsColumnDependentBand_ReportSummaryBand_True(t *testing.T) {
	b := band.NewReportSummaryBand()
	if !b.IsColumnDependentBand() {
		t.Error("ReportSummaryBand.IsColumnDependentBand should return true")
	}
}

// ── ChildBand.GetTopParentBand ────────────────────────────────────────────────

// TestChildBand_GetTopParentBand_NoParent verifies nil is returned when the
// ChildBand has no parent.
// Mirrors C# ChildBand.GetTopParentBand (ChildBand.cs line 67-79).
func TestChildBand_GetTopParentBand_NoParent(t *testing.T) {
	c := band.NewChildBand()
	top := c.GetTopParentBand()
	if top != nil {
		t.Error("GetTopParentBand with no parent should return nil")
	}
}

// TestChildBand_GetTopParentBand_DirectDataBandParent verifies that a ChildBand
// with a DataBand parent returns that DataBand (via the columnDependentChecker
// interface).  Since GetTopParentBand returns the interface, we verify the
// parent is identified as column-dependent.
func TestChildBand_GetTopParentBand_DirectDataBandParent(t *testing.T) {
	db := band.NewDataBand()
	c := band.NewChildBand()
	db.AddChild(c)

	top := c.GetTopParentBand()
	if top == nil {
		t.Fatal("GetTopParentBand should find DataBand parent, got nil")
	}
	if !top.IsColumnDependentBand() {
		t.Error("GetTopParentBand: parent DataBand should be column-dependent")
	}
}

// TestChildBand_GetTopParentBand_NestedChildBands verifies that a doubly-nested
// ChildBand correctly traverses two levels to reach the real parent.
func TestChildBand_GetTopParentBand_NestedChildBands(t *testing.T) {
	db := band.NewDataBand()
	c1 := band.NewChildBand()
	c2 := band.NewChildBand()
	// c1 is directly under db; c2 is the child of c1.
	db.AddChild(c1)
	c1.SetChild(c2)
	c2.SetParent(c1)

	top := c2.GetTopParentBand()
	if top == nil {
		t.Fatal("GetTopParentBand should traverse two ChildBand levels, got nil")
	}
	if !top.IsColumnDependentBand() {
		t.Error("GetTopParentBand: eventual parent DataBand should be column-dependent")
	}
}

// TestChildBand_GetTopParentBand_PageHeaderParent verifies that a ChildBand
// under a PageHeaderBand (not column-dependent) correctly returns that band.
func TestChildBand_GetTopParentBand_PageHeaderParent(t *testing.T) {
	ph := band.NewPageHeaderBand()
	c := band.NewChildBand()
	ph.AddChild(c)

	top := c.GetTopParentBand()
	if top == nil {
		t.Fatal("GetTopParentBand should find PageHeaderBand parent, got nil")
	}
	if top.IsColumnDependentBand() {
		t.Error("GetTopParentBand: parent PageHeaderBand should NOT be column-dependent")
	}
}

// ── ChildBand.IsColumnDependentBand ──────────────────────────────────────────

// TestChildBand_IsColumnDependentBand_NoParent verifies false when no parent.
func TestChildBand_IsColumnDependentBand_NoParent(t *testing.T) {
	c := band.NewChildBand()
	if c.IsColumnDependentBand() {
		t.Error("ChildBand with no parent should return false for IsColumnDependentBand")
	}
}

// TestChildBand_IsColumnDependentBand_UnderDataBand verifies true when under DataBand.
// Mirrors C# BandBase.IsColumnDependentBand for ChildBand (BandBase.cs line 582-586).
func TestChildBand_IsColumnDependentBand_UnderDataBand(t *testing.T) {
	db := band.NewDataBand()
	c := band.NewChildBand()
	db.AddChild(c)

	if !c.IsColumnDependentBand() {
		t.Error("ChildBand under DataBand should be column-dependent")
	}
}

// TestChildBand_IsColumnDependentBand_UnderPageHeader verifies false under PageHeaderBand.
func TestChildBand_IsColumnDependentBand_UnderPageHeader(t *testing.T) {
	ph := band.NewPageHeaderBand()
	c := band.NewChildBand()
	ph.AddChild(c)

	if c.IsColumnDependentBand() {
		t.Error("ChildBand under PageHeaderBand should NOT be column-dependent")
	}
}

// ── ChildBand.Assign ─────────────────────────────────────────────────────────

// TestChildBand_Assign_CopiesAllProperties verifies all ChildBand-specific
// fields are copied from source.
// Mirrors C# ChildBand.Assign(Base source) (ChildBand.cs line 82-89).
func TestChildBand_Assign_CopiesAllProperties(t *testing.T) {
	src := band.NewChildBand()
	src.FillUnusedSpace = true
	src.CompleteToNRows = 5
	src.PrintIfDatabandEmpty = true
	// Also set a BandBase field to confirm BandBase.Assign is called.
	src.SetStartNewPage(true)

	dst := band.NewChildBand()
	dst.Assign(src)

	if !dst.FillUnusedSpace {
		t.Error("Assign: FillUnusedSpace not copied")
	}
	if dst.CompleteToNRows != 5 {
		t.Errorf("Assign: CompleteToNRows = %d, want 5", dst.CompleteToNRows)
	}
	if !dst.PrintIfDatabandEmpty {
		t.Error("Assign: PrintIfDatabandEmpty not copied")
	}
	if !dst.StartNewPage() {
		t.Error("Assign: BandBase.StartNewPage not copied via BandBase.Assign")
	}
}

// TestChildBand_Assign_NilSource verifies Assign(nil) is a no-op.
func TestChildBand_Assign_NilSource(t *testing.T) {
	dst := band.NewChildBand()
	dst.FillUnusedSpace = true
	dst.Assign(nil) // must not panic
	if !dst.FillUnusedSpace {
		t.Error("Assign(nil) should not modify dst")
	}
}

// ── report.BreakableComponent.Assign ─────────────────────────────────────────

// TestBreakableComponent_Assign_CopiesCanBreak verifies that Assign copies
// the canBreak field.
// Mirrors C# BreakableComponent.Assign(Base source) (BreakableComponent.cs line 64-71).
func TestBreakableComponent_Assign_CopiesCanBreak(t *testing.T) {
	src := report.NewBreakableComponent()
	src.SetCanBreak(false)

	dst := report.NewBreakableComponent()
	dst.Assign(src)

	if dst.CanBreak() {
		t.Error("Assign: CanBreak should have been copied as false")
	}
}

// TestBreakableComponent_Assign_CopiesBreakTo verifies that Assign copies the
// breakTo reference.
func TestBreakableComponent_Assign_CopiesBreakTo(t *testing.T) {
	target := report.NewBreakableComponent()
	src := report.NewBreakableComponent()
	src.SetBreakTo(target)

	dst := report.NewBreakableComponent()
	dst.Assign(src)

	if dst.BreakTo() != target {
		t.Error("Assign: BreakTo reference not copied")
	}
}

// TestBreakableComponent_Assign_NilSource verifies Assign(nil) is a no-op.
func TestBreakableComponent_Assign_NilSource(t *testing.T) {
	dst := report.NewBreakableComponent()
	dst.SetCanBreak(false)
	dst.Assign(nil) // must not panic
	if dst.CanBreak() {
		t.Error("Assign(nil) should not modify dst")
	}
}

// TestBreakableComponent_Assign_ClearsBreakToWhenSourceNilBreakTo verifies
// that Assign propagates a nil BreakTo correctly.
func TestBreakableComponent_Assign_ClearsBreakToWhenSourceNilBreakTo(t *testing.T) {
	target := report.NewBreakableComponent()
	dst := report.NewBreakableComponent()
	dst.SetBreakTo(target)

	src := report.NewBreakableComponent() // BreakTo = nil
	dst.Assign(src)

	if dst.BreakTo() != nil {
		t.Error("Assign: BreakTo should be nil after assigning from source with nil BreakTo")
	}
}
