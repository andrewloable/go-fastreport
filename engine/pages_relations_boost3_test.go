package engine

// pages_relations_boost3_test.go — internal tests (package engine) targeting
// the remaining uncovered branches in pages.go (attachWatermark, showBand)
// and relations.go (applyRelationFilters alias-resolution path).
//
// Uses package engine (not engine_test) so we can access unexported fields.

import (
	"testing"

	"github.com/andrewloable/go-fastreport/band"
	"github.com/andrewloable/go-fastreport/data"
	"github.com/andrewloable/go-fastreport/report"
	"github.com/andrewloable/go-fastreport/reportpkg"
)

// ── pages.go: showBand — nil interface guard ──────────────────────────────────

// TestShowBand_NilInterface covers the `if b == nil { return }` guard in showBand.
// This is triggered by passing an untyped nil (nil interface value), which is
// distinct from a typed nil pointer (covered by TestShowBand_TypedNilPointer).
func TestShowBand_NilInterface(t *testing.T) {
	r := reportpkg.NewReport()
	pg := reportpkg.NewReportPage()
	r.AddPage(pg)

	e := New(r)
	if err := e.Run(DefaultRunOptions()); err != nil {
		t.Fatalf("Run: %v", err)
	}

	// Pass an untyped nil interface (report.Base = nil).
	// This hits the `if b == nil { return }` at line 199.
	e.showBand(nil)
	// Must not panic.
}

// ── pages.go: attachWatermark — preparedPages nil guard ───────────────────────

// TestAttachWatermark_NilPreparedPages covers the `e.preparedPages == nil` early
// return path in attachWatermark. We set preparedPages to nil and call directly.
func TestAttachWatermark_NilPreparedPages(t *testing.T) {
	r := reportpkg.NewReport()
	pg := reportpkg.NewReportPage()
	pg.SetName("WMPage")
	pg.Watermark = reportpkg.NewWatermark()
	pg.Watermark.Enabled = true
	pg.Watermark.Text = "TEST"
	r.AddPage(pg)

	e := New(r)
	// Explicitly set preparedPages to nil to trigger the nil guard.
	e.preparedPages = nil

	// Should return immediately without panic.
	e.attachWatermark(pg)
}

// ── pages.go: showBand — preparedPages nil path ───────────────────────────────

// TestShowBand_PreparedPagesNil covers the false branch of
// `if e.preparedPages != nil` in showBand. When preparedPages is nil but the
// band is valid with height > 0, the inner block is skipped and AdvanceY is
// still called.
func TestShowBand_PreparedPagesNil(t *testing.T) {
	r := reportpkg.NewReport()
	pg := reportpkg.NewReportPage()
	r.AddPage(pg)

	e := New(r)
	// Set preparedPages to nil so the inner `if e.preparedPages != nil` block
	// in showBand is skipped.
	e.preparedPages = nil

	b := band.NewPageHeaderBand()
	b.SetName("PHBand")
	b.SetHeight(20)
	b.SetVisible(true)

	beforeY := e.curY
	e.showBand(b)
	// AdvanceY is still called even when preparedPages is nil.
	if e.curY != beforeY+20 {
		t.Errorf("showBand with nil preparedPages: curY = %v, want %v", e.curY, beforeY+20)
	}
}

// ── pages.go: showBand — hasObjects false branch ──────────────────────────────

// TestShowBand_NoObjectsInterface covers the `if ho, ok := b.(hasObjects); ok`
// false branch in showBand, where the band satisfies report.Base and has Height()
// but does NOT implement Objects() *report.ObjectCollection.
// geomOnlyObj is defined in objects_coverage_gaps_test.go (same package).
func TestShowBand_NoObjectsInterface(t *testing.T) {
	r := reportpkg.NewReport()
	pg := reportpkg.NewReportPage()
	r.AddPage(pg)

	e := New(r)
	if err := e.Run(DefaultRunOptions()); err != nil {
		t.Fatalf("Run: %v", err)
	}

	// Add a page so preparedPages.CurrentPage() is non-nil.
	e.preparedPages.AddPage(200, 800, 2)

	// geomOnlyObj embeds report.BaseObject and adds geometry but no Objects().
	b := &geomOnlyObj{}
	// geomOnlyObj.Height() returns 20, so height > 0.
	// geomOnlyObj does not implement hasObjects, so the inner if is false.
	beforeY := e.curY
	e.showBand(b)
	// AdvanceY should still be called (height = 20).
	if e.curY != beforeY+20 {
		t.Errorf("showBand noObjects: curY = %v, want %v", e.curY, beforeY+20)
	}
}

// ── relations.go: applyRelationFilters — nil report early return ──────────────

// TestApplyRelationFilters_NilReport covers the `e.report == nil` early return
// path in applyRelationFilters. We construct an engine with nil report directly.
func TestApplyRelationFilters_NilReport(t *testing.T) {
	e := &ReportEngine{} // report is nil

	masterBand := band.NewDataBand()
	masterBand.SetName("NilReportMaster")
	masterBand.SetHeight(10)

	subBands := []report.Base{}
	restore := e.applyRelationFilters(masterBand, subBands)
	// Should return no-op restore without panicking.
	restore()
}

// TestApplyRelationFilters_SubBandNotDataBandDirect covers the `if !ok { continue }`
// branch at line 39-40. This is only reachable by calling applyRelationFilters
// directly with a subBands slice containing a non-DataBand element, since
// dataBandSubBands filters to DataBand types only.
func TestApplyRelationFilters_SubBandNotDataBandDirect(t *testing.T) {
	masterDS := data.NewBaseDataSource("MasterDirect")
	masterDS.SetAlias("MasterDirect")
	masterDS.AddColumn(data.Column{Name: "ID"})
	masterDS.AddRow(map[string]any{"ID": "1"})
	if err := masterDS.Init(); err != nil {
		t.Fatalf("masterDS.Init: %v", err)
	}

	// Add a dummy relation so the dict.Relations() check passes.
	rel := &data.Relation{
		Name:             "DummyRel",
		ParentDataSource: masterDS,
		ChildDataSource:  masterDS,
		ParentColumns:    []string{"ID"},
		ChildColumns:     []string{"ID"},
	}

	r := reportpkg.NewReport()
	dict := r.Dictionary()
	dict.AddDataSource(masterDS)
	dict.AddRelation(rel)

	masterBand := band.NewDataBand()
	masterBand.SetName("DirectMasterBand")
	masterBand.SetHeight(10)
	masterBand.SetDataSource(masterDS)

	r.AddPage(reportpkg.NewReportPage())
	e := New(r)
	if err := e.Run(DefaultRunOptions()); err != nil {
		t.Fatalf("Run: %v", err)
	}

	// Call applyRelationFilters directly with a non-DataBand sub-band.
	// This exercises the `if !ok { continue }` path (line 39-40).
	nonDB := band.NewGroupFooterBand() // not a *band.DataBand → !ok → continue
	nonDB.SetName("NonDataBand")
	subBands := []report.Base{nonDB}

	restore := e.applyRelationFilters(masterBand, subBands)
	restore()
}

// TestApplyRelationFilters_RelNilForChild covers the `if rel == nil { continue }`
// branch. Requires: dict has a relation (so early return is NOT taken), but the
// specific parent→child pair has no matching relation.
func TestApplyRelationFilters_RelNilForChild(t *testing.T) {
	masterDS := data.NewBaseDataSource("MasterRelNil")
	masterDS.SetAlias("MasterRelNil")
	masterDS.AddColumn(data.Column{Name: "ID"})
	masterDS.AddRow(map[string]any{"ID": "1"})
	if err := masterDS.Init(); err != nil {
		t.Fatalf("masterDS.Init: %v", err)
	}

	childDS := data.NewBaseDataSource("ChildRelNil")
	childDS.SetAlias("ChildRelNil")
	childDS.AddColumn(data.Column{Name: "ParentID"})
	childDS.AddRow(map[string]any{"ParentID": "1"})
	if err := childDS.Init(); err != nil {
		t.Fatalf("childDS.Init: %v", err)
	}

	// A third DS for the relation — ensures dict.Relations() is non-empty,
	// but FindRelation(masterDS, childDS) returns nil.
	thirdDS := data.NewBaseDataSource("ThirdRelNil")
	thirdDS.SetAlias("ThirdRelNil")
	thirdDS.AddColumn(data.Column{Name: "X"})
	if err := thirdDS.Init(); err != nil {
		t.Fatalf("thirdDS.Init: %v", err)
	}

	// Relation between masterDS and thirdDS — NOT between masterDS and childDS.
	rel := &data.Relation{
		Name:             "MasterThirdRel",
		ParentDataSource: masterDS,
		ChildDataSource:  thirdDS,
		ParentColumns:    []string{"ID"},
		ChildColumns:     []string{"X"},
	}

	r := reportpkg.NewReport()
	dict := r.Dictionary()
	dict.AddDataSource(masterDS)
	dict.AddDataSource(childDS)
	dict.AddDataSource(thirdDS)
	dict.AddRelation(rel)

	pg := reportpkg.NewReportPage()

	masterBand := band.NewDataBand()
	masterBand.SetName("MasterRelNilBand")
	masterBand.SetHeight(10)
	masterBand.SetVisible(true)
	masterBand.SetDataSource(masterDS)

	// Child band uses childDS — no relation exists for masterDS→childDS pair.
	// FindRelation returns nil → `if rel == nil { continue }` is hit.
	childBand := band.NewDataBand()
	childBand.SetName("ChildRelNilBand")
	childBand.SetHeight(10)
	childBand.SetVisible(true)
	childBand.SetDataSource(childDS)

	masterBand.Objects().Add(childBand)
	pg.AddBand(masterBand)
	r.AddPage(pg)

	e := New(r)
	if err := e.Run(DefaultRunOptions()); err != nil {
		t.Fatalf("Run: %v", err)
	}
}

// ── relations.go: applyRelationFilters — alias resolution success ─────────────

// TestApplyRelationFilters_ChildAliasResolved exercises the alias-resolution
// branch where childDB.DataSourceRef() is nil but DataSourceAlias() resolves
// to a valid band.DataSource in the dictionary:
//
//	if alias := childDB.DataSourceAlias(); alias != "" {
//	    if resolved := dict.FindDataSourceByAlias(alias); resolved != nil {
//	        if bds, ok2 := resolved.(band.DataSource); ok2 {
//	            childDS = bds   ← this assignment was previously uncovered
//	        }
//	    }
//	}
func TestApplyRelationFilters_ChildAliasResolved(t *testing.T) {
	masterDS := data.NewBaseDataSource("MasterAlias")
	masterDS.SetAlias("MasterAlias")
	masterDS.AddColumn(data.Column{Name: "ID"})
	masterDS.AddRow(map[string]any{"ID": "1"})
	if err := masterDS.Init(); err != nil {
		t.Fatalf("masterDS.Init: %v", err)
	}

	childDS := data.NewBaseDataSource("ChildAlias")
	childDS.SetAlias("ChildAlias")
	childDS.AddColumn(data.Column{Name: "ParentID"})
	childDS.AddRow(map[string]any{"ParentID": "1"})
	if err := childDS.Init(); err != nil {
		t.Fatalf("childDS.Init: %v", err)
	}

	// Relation linking master → child data sources.
	rel := &data.Relation{
		Name:             "AliasRel",
		ParentDataSource: masterDS,
		ChildDataSource:  childDS,
		ParentColumns:    []string{"ID"},
		ChildColumns:     []string{"ParentID"},
	}

	r := reportpkg.NewReport()
	dict := r.Dictionary()
	dict.AddDataSource(masterDS)
	dict.AddDataSource(childDS)
	dict.AddRelation(rel)

	pg := reportpkg.NewReportPage()

	masterBand := band.NewDataBand()
	masterBand.SetName("MasterAliasBand")
	masterBand.SetHeight(15)
	masterBand.SetVisible(true)
	masterBand.SetDataSource(masterDS)

	// Child band: no direct data source, but has an alias that resolves in dict.
	// This forces the alias-resolution branch inside applyRelationFilters.
	childBand := band.NewDataBand()
	childBand.SetName("ChildAliasBand")
	childBand.SetHeight(10)
	childBand.SetVisible(true)
	// No SetDataSource — DataSourceRef() returns nil, triggering alias lookup.
	childBand.SetDataSourceAlias("ChildAlias")

	// Add as a sub-band (DataBand inside DataBand's Objects) so dataBandSubBands
	// picks it up.
	masterBand.Objects().Add(childBand)
	pg.AddBand(masterBand)
	r.AddPage(pg)

	e := New(r)
	if err := e.Run(DefaultRunOptions()); err != nil {
		t.Fatalf("Run with alias-resolved child: %v", err)
	}
	if e.preparedPages.Count() == 0 {
		t.Error("expected at least 1 prepared page")
	}
}

// TestApplyRelationFilters_ThreeLevelNestedMasterDetail verifies that a 3-level
// master-detail hierarchy (grandparent → parent → child) does NOT produce a
// cartesian product when the parent DataBand's data source has been wrapped in a
// FilteredDataSource by the grandparent's applyRelationFilters call.
//
// Bug: when the parent band's DS is a FilteredDataSource, FindRelation could not
// match it against the relation's ParentDataSource pointer (which is the unwrapped
// base DS), so the grandchild band never got filtered, iterating ALL grandchild
// rows for every parent row → O(N*M) instead of O(M).
//
// Setup:
//   - Employees: 2 rows (IDs 1, 2)
//   - Orders: 3 rows (2 for employee 1, 1 for employee 2)
//   - OrderDetails: 4 rows (2 for order 10, 1 for order 11, 1 for order 12)
//
// Relations: Employees→Orders (EmployeeID), Orders→OrderDetails (OrderID)
//
// Expected rendered DataBand rows:
//   - Employee 1 → Orders 10, 11 → OrderDetails: 2+1 = 3 detail rows
//   - Employee 2 → Order 12 → OrderDetails: 1 detail row
//
// Total detail (Data3/OrderDetails) band invocations: 4 (NOT 2*4=8 cartesian).
func TestApplyRelationFilters_ThreeLevelNestedMasterDetail(t *testing.T) {
	// ── Employees ──────────────────────────────────────────────────────────────
	empDS := data.NewBaseDataSource("Employees")
	empDS.SetAlias("Employees")
	empDS.AddColumn(data.Column{Name: "EmployeeID"})
	empDS.AddRow(map[string]any{"EmployeeID": "1"})
	empDS.AddRow(map[string]any{"EmployeeID": "2"})
	if err := empDS.Init(); err != nil {
		t.Fatalf("empDS.Init: %v", err)
	}

	// ── Orders ─────────────────────────────────────────────────────────────────
	ordDS := data.NewBaseDataSource("Orders")
	ordDS.SetAlias("Orders")
	ordDS.AddColumn(data.Column{Name: "OrderID"})
	ordDS.AddColumn(data.Column{Name: "EmployeeID"})
	ordDS.AddRow(map[string]any{"OrderID": "10", "EmployeeID": "1"})
	ordDS.AddRow(map[string]any{"OrderID": "11", "EmployeeID": "1"})
	ordDS.AddRow(map[string]any{"OrderID": "12", "EmployeeID": "2"})
	if err := ordDS.Init(); err != nil {
		t.Fatalf("ordDS.Init: %v", err)
	}

	// ── OrderDetails ───────────────────────────────────────────────────────────
	detDS := data.NewBaseDataSource("OrderDetails")
	detDS.SetAlias("OrderDetails")
	detDS.AddColumn(data.Column{Name: "OrderID"})
	detDS.AddRow(map[string]any{"OrderID": "10"})
	detDS.AddRow(map[string]any{"OrderID": "10"})
	detDS.AddRow(map[string]any{"OrderID": "11"})
	detDS.AddRow(map[string]any{"OrderID": "12"})
	if err := detDS.Init(); err != nil {
		t.Fatalf("detDS.Init: %v", err)
	}

	// ── Relations ──────────────────────────────────────────────────────────────
	relEmpOrd := &data.Relation{
		Name:             "EmpOrd",
		ParentDataSource: empDS,
		ChildDataSource:  ordDS,
		ParentColumns:    []string{"EmployeeID"},
		ChildColumns:     []string{"EmployeeID"},
	}
	relOrdDet := &data.Relation{
		Name:             "OrdDet",
		ParentDataSource: ordDS,
		ChildDataSource:  detDS,
		ParentColumns:    []string{"OrderID"},
		ChildColumns:     []string{"OrderID"},
	}

	// ── Report & bands ─────────────────────────────────────────────────────────
	r := reportpkg.NewReport()
	dict := r.Dictionary()
	dict.AddDataSource(empDS)
	dict.AddDataSource(ordDS)
	dict.AddDataSource(detDS)
	dict.AddRelation(relEmpOrd)
	dict.AddRelation(relOrdDet)

	pg := reportpkg.NewReportPage()

	// Employee band (Data1)
	empBand := band.NewDataBand()
	empBand.SetName("Data1")
	empBand.SetHeight(10)
	empBand.SetVisible(true)
	empBand.SetDataSource(empDS)

	// Orders band (Data2) — nested inside empBand
	ordBand := band.NewDataBand()
	ordBand.SetName("Data2")
	ordBand.SetHeight(10)
	ordBand.SetVisible(true)
	ordBand.SetDataSource(ordDS)

	// OrderDetails band (Data3) — nested inside ordBand
	detBand := band.NewDataBand()
	detBand.SetName("Data3")
	detBand.SetHeight(10)
	detBand.SetVisible(true)
	detBand.SetDataSource(detDS)

	ordBand.Objects().Add(detBand)
	empBand.Objects().Add(ordBand)
	pg.AddBand(empBand)
	r.AddPage(pg)

	e := New(r)
	if err := e.Run(DefaultRunOptions()); err != nil {
		t.Fatalf("Run: %v", err)
	}

	// Count how many Data3 (OrderDetails) PreparedBands were rendered.
	var detCount int
	for i := 0; i < e.preparedPages.Count(); i++ {
		pp := e.preparedPages.GetPage(i)
		for _, pb := range pp.Bands {
			if pb.Name == "Data3" {
				detCount++
			}
		}
	}

	// Expect exactly 4 detail rows (2 for order 10, 1 for order 11, 1 for order 12).
	// Without the fix, this would be 8 (2 employees × 4 detail rows = cartesian product).
	if detCount != 4 {
		t.Errorf("expected 4 OrderDetails band invocations, got %d (cartesian product bug?)", detCount)
	}
}
