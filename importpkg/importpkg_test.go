package importpkg_test

import (
	"io"
	"strings"
	"testing"

	"github.com/andrewloable/go-fastreport/importpkg"
	"github.com/andrewloable/go-fastreport/reportpkg"
)

// ── ImportBase ───────────────────────────────────────────────────────────────

func TestImportBase_NameGetSet(t *testing.T) {
	var b importpkg.ImportBase
	if b.Name() != "" {
		t.Fatalf("expected empty name, got %q", b.Name())
	}
	b.SetName("RDL Importer")
	if b.Name() != "RDL Importer" {
		t.Fatalf("expected %q, got %q", "RDL Importer", b.Name())
	}
}

func TestImportBase_ReportGetSet(t *testing.T) {
	var b importpkg.ImportBase
	if b.Report() != nil {
		t.Fatal("expected nil report initially")
	}
	rpt := reportpkg.NewReport()
	b.SetReport(rpt)
	if b.Report() != rpt {
		t.Fatal("SetReport/Report round-trip failed")
	}
}

func TestImportBase_LoadReportFromFile_SetsReport(t *testing.T) {
	var b importpkg.ImportBase
	rpt := reportpkg.NewReport()
	err := b.LoadReportFromFile(rpt, "nonexistent.rdl")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if b.Report() != rpt {
		t.Fatal("LoadReportFromFile should store the report reference")
	}
}

func TestImportBase_LoadReportFromStream_SetsReport(t *testing.T) {
	var b importpkg.ImportBase
	rpt := reportpkg.NewReport()
	r := strings.NewReader("<empty/>")
	err := b.LoadReportFromStream(rpt, r)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if b.Report() != rpt {
		t.Fatal("LoadReportFromStream should store the report reference")
	}
}

// Verify that ImportBase implements the expected Importer behaviour when
// embedded by a concrete type.
type testImporter struct {
	importpkg.ImportBase
	callCount int
}

func (ti *testImporter) LoadReportFromFile(report *reportpkg.Report, filename string) error {
	ti.callCount++
	ti.SetReport(report)
	return nil
}

func (ti *testImporter) LoadReportFromStream(report *reportpkg.Report, r io.Reader) error {
	ti.callCount++
	ti.SetReport(report)
	return nil
}

func TestImportBase_EmbeddingPattern(t *testing.T) {
	imp := &testImporter{}
	imp.SetName("Test Importer")

	rpt := reportpkg.NewReport()
	_ = imp.LoadReportFromFile(rpt, "file.xml")
	_ = imp.LoadReportFromStream(rpt, strings.NewReader(""))

	if imp.callCount != 2 {
		t.Fatalf("expected 2 calls, got %d", imp.callCount)
	}
	if imp.Name() != "Test Importer" {
		t.Fatalf("name not preserved via embedding")
	}
}

// ── isValidIdentifier (tested via CreateTextObject behaviour) ─────────────────

func TestIsValidIdentifier_ViaCreateTextObject(t *testing.T) {
	rpt := reportpkg.NewReport()
	page := importpkg.CreateReportPage(rpt)
	db := importpkg.CreateDataBand(page)

	// Valid name: object keeps the name.
	obj := importpkg.CreateTextObject("MyText", db)
	if obj.Name() != "MyText" {
		t.Fatalf("expected name %q, got %q", "MyText", obj.Name())
	}

	// Invalid name (starts with digit): object gets auto-generated name.
	obj2 := importpkg.CreateTextObject("123invalid", db)
	if obj2.Name() == "123invalid" {
		t.Fatal("invalid identifier should trigger auto-name generation")
	}
	if obj2.Name() == "" {
		t.Fatal("auto-generated name must not be empty")
	}

	// Empty name: object gets auto-generated name.
	obj3 := importpkg.CreateTextObject("", db)
	if obj3.Name() == "" {
		t.Fatal("empty name should trigger auto-name generation")
	}
}

// ── Pages ─────────────────────────────────────────────────────────────────────

func TestCreateReportPage_AddsPageToReport(t *testing.T) {
	rpt := reportpkg.NewReport()
	page := importpkg.CreateReportPage(rpt)

	if page == nil {
		t.Fatal("CreateReportPage returned nil")
	}
	if rpt.PageCount() != 1 {
		t.Fatalf("expected 1 page, got %d", rpt.PageCount())
	}
	if rpt.Page(0) != page {
		t.Fatal("page not registered in report")
	}
	if page.Name() == "" {
		t.Fatal("page must have a non-empty auto-generated name")
	}
}

func TestCreateReportPage_MultiplePagesHaveUniqueNames(t *testing.T) {
	rpt := reportpkg.NewReport()
	p1 := importpkg.CreateReportPage(rpt)
	p2 := importpkg.CreateReportPage(rpt)

	if p1.Name() == p2.Name() {
		t.Fatalf("pages must have unique names, both got %q", p1.Name())
	}
}

func TestCreateReportPageNamed_ValidName(t *testing.T) {
	rpt := reportpkg.NewReport()
	page := importpkg.CreateReportPageNamed("SalesPage", rpt)

	if page.Name() != "SalesPage" {
		t.Fatalf("expected name %q, got %q", "SalesPage", page.Name())
	}
}

func TestCreateReportPageNamed_InvalidName_GetsAutoName(t *testing.T) {
	rpt := reportpkg.NewReport()
	page := importpkg.CreateReportPageNamed("123bad", rpt)

	if page.Name() == "123bad" {
		t.Fatal("invalid page name should be replaced by auto-generated name")
	}
	if page.Name() == "" {
		t.Fatal("auto-generated page name must not be empty")
	}
}

// ── Bands ────────────────────────────────────────────────────────────────────

func TestCreateReportTitleBand(t *testing.T) {
	rpt := reportpkg.NewReport()
	page := importpkg.CreateReportPage(rpt)

	b := importpkg.CreateReportTitleBand(page)
	if b == nil {
		t.Fatal("CreateReportTitleBand returned nil")
	}
	if page.ReportTitle() != b {
		t.Fatal("band not assigned to page.ReportTitle")
	}
	if b.Name() == "" {
		t.Fatal("band must have an auto-generated name")
	}
}

func TestCreateReportSummaryBand(t *testing.T) {
	rpt := reportpkg.NewReport()
	page := importpkg.CreateReportPage(rpt)

	b := importpkg.CreateReportSummaryBand(page)
	if page.ReportSummary() != b {
		t.Fatal("band not assigned to page.ReportSummary")
	}
}

func TestCreatePageHeaderBand(t *testing.T) {
	rpt := reportpkg.NewReport()
	page := importpkg.CreateReportPage(rpt)

	b := importpkg.CreatePageHeaderBand(page)
	if page.PageHeader() != b {
		t.Fatal("band not assigned to page.PageHeader")
	}
}

func TestCreatePageFooterBand(t *testing.T) {
	rpt := reportpkg.NewReport()
	page := importpkg.CreateReportPage(rpt)

	b := importpkg.CreatePageFooterBand(page)
	if page.PageFooter() != b {
		t.Fatal("band not assigned to page.PageFooter")
	}
}

func TestCreateColumnHeaderBand(t *testing.T) {
	rpt := reportpkg.NewReport()
	page := importpkg.CreateReportPage(rpt)

	b := importpkg.CreateColumnHeaderBand(page)
	if page.ColumnHeader() != b {
		t.Fatal("band not assigned to page.ColumnHeader")
	}
}

func TestCreateColumnFooterBand(t *testing.T) {
	rpt := reportpkg.NewReport()
	page := importpkg.CreateReportPage(rpt)

	b := importpkg.CreateColumnFooterBand(page)
	if page.ColumnFooter() != b {
		t.Fatal("band not assigned to page.ColumnFooter")
	}
}

func TestCreateDataBand(t *testing.T) {
	rpt := reportpkg.NewReport()
	page := importpkg.CreateReportPage(rpt)

	db := importpkg.CreateDataBand(page)
	if db == nil {
		t.Fatal("CreateDataBand returned nil")
	}
	bands := page.Bands()
	if len(bands) != 1 || bands[0] != db {
		t.Fatal("DataBand not appended to page.Bands")
	}
}

func TestCreateDataHeaderBand(t *testing.T) {
	rpt := reportpkg.NewReport()
	page := importpkg.CreateReportPage(rpt)
	db := importpkg.CreateDataBand(page)

	h := importpkg.CreateDataHeaderBand(db)
	if db.Header() != h {
		t.Fatal("DataHeaderBand not set as DataBand.Header")
	}
}

func TestCreateDataFooterBand(t *testing.T) {
	rpt := reportpkg.NewReport()
	page := importpkg.CreateReportPage(rpt)
	db := importpkg.CreateDataBand(page)

	f := importpkg.CreateDataFooterBand(db)
	if db.Footer() != f {
		t.Fatal("DataFooterBand not set as DataBand.Footer")
	}
}

func TestCreateGroupHeaderBand(t *testing.T) {
	rpt := reportpkg.NewReport()
	page := importpkg.CreateReportPage(rpt)

	gh := importpkg.CreateGroupHeaderBand(page)
	if gh == nil {
		t.Fatal("CreateGroupHeaderBand returned nil")
	}
	bands := page.Bands()
	if len(bands) != 1 || bands[0] != gh {
		t.Fatal("GroupHeaderBand not appended to page.Bands")
	}
}

func TestCreateGroupFooterBandOnPage(t *testing.T) {
	rpt := reportpkg.NewReport()
	page := importpkg.CreateReportPage(rpt)

	gf := importpkg.CreateGroupFooterBandOnPage(page)
	if gf == nil {
		t.Fatal("CreateGroupFooterBandOnPage returned nil")
	}
	bands := page.Bands()
	if len(bands) != 1 || bands[0] != gf {
		t.Fatal("GroupFooterBand not appended to page.Bands")
	}
}

func TestCreateGroupFooterBand_AssignedToGroupHeader(t *testing.T) {
	rpt := reportpkg.NewReport()
	page := importpkg.CreateReportPage(rpt)
	gh := importpkg.CreateGroupHeaderBand(page)

	gf := importpkg.CreateGroupFooterBand(gh)
	if gh.GroupFooter() != gf {
		t.Fatal("GroupFooterBand not assigned to GroupHeader.GroupFooter")
	}
}

func TestCreateChildBand(t *testing.T) {
	rpt := reportpkg.NewReport()
	page := importpkg.CreateReportPage(rpt)
	db := importpkg.CreateDataBand(page)

	// Pass the embedded BandBase pointer via the DataBand.
	cb := importpkg.CreateChildBand(db)
	if cb == nil {
		t.Fatal("CreateChildBand returned nil")
	}
	if db.Child() != cb {
		t.Fatal("ChildBand not stored in DataBand.Child")
	}
}

func TestCreateOverlayBand(t *testing.T) {
	rpt := reportpkg.NewReport()
	page := importpkg.CreateReportPage(rpt)

	ob := importpkg.CreateOverlayBand(page)
	if page.Overlay() != ob {
		t.Fatal("OverlayBand not assigned to page.Overlay")
	}
}

// ── Band name uniqueness across a page ───────────────────────────────────────

func TestBandNames_UniqueAcrossPage(t *testing.T) {
	rpt := reportpkg.NewReport()
	page := importpkg.CreateReportPage(rpt)

	ph := importpkg.CreatePageHeaderBand(page)
	pf := importpkg.CreatePageFooterBand(page)
	rt := importpkg.CreateReportTitleBand(page)
	db := importpkg.CreateDataBand(page)

	names := map[string]bool{
		ph.Name(): true,
		pf.Name(): false, // initialised below
		rt.Name(): false,
		db.Name(): false,
	}
	names[pf.Name()] = true
	names[rt.Name()] = true
	names[db.Name()] = true

	if len(names) != 4 {
		t.Fatalf("expected 4 unique band names, got %d: %v",
			len(names), names)
	}
}

// ── Objects ──────────────────────────────────────────────────────────────────

func TestCreateTextObject_ParentCanContain(t *testing.T) {
	rpt := reportpkg.NewReport()
	page := importpkg.CreateReportPage(rpt)
	db := importpkg.CreateDataBand(page)

	obj := importpkg.CreateTextObject("Label1", db)
	if obj == nil {
		t.Fatal("CreateTextObject returned nil")
	}
	if obj.Name() != "Label1" {
		t.Fatalf("expected name %q, got %q", "Label1", obj.Name())
	}
}

func TestCreatePictureObject(t *testing.T) {
	rpt := reportpkg.NewReport()
	page := importpkg.CreateReportPage(rpt)
	db := importpkg.CreateDataBand(page)

	obj := importpkg.CreatePictureObject("Pic1", db)
	if obj == nil {
		t.Fatal("CreatePictureObject returned nil")
	}
	if obj.Name() != "Pic1" {
		t.Fatalf("expected name %q, got %q", "Pic1", obj.Name())
	}
}

func TestCreateLineObject(t *testing.T) {
	rpt := reportpkg.NewReport()
	page := importpkg.CreateReportPage(rpt)
	db := importpkg.CreateDataBand(page)

	obj := importpkg.CreateLineObject("Line1", db)
	if obj.Name() != "Line1" {
		t.Fatalf("expected %q, got %q", "Line1", obj.Name())
	}
}

func TestCreateShapeObject(t *testing.T) {
	rpt := reportpkg.NewReport()
	page := importpkg.CreateReportPage(rpt)
	db := importpkg.CreateDataBand(page)

	obj := importpkg.CreateShapeObject("Shape1", db)
	if obj.Name() != "Shape1" {
		t.Fatalf("expected %q, got %q", "Shape1", obj.Name())
	}
}

func TestCreatePolyLineObject(t *testing.T) {
	rpt := reportpkg.NewReport()
	page := importpkg.CreateReportPage(rpt)
	db := importpkg.CreateDataBand(page)

	obj := importpkg.CreatePolyLineObject("Poly1", db)
	if obj.Name() != "Poly1" {
		t.Fatalf("expected %q, got %q", "Poly1", obj.Name())
	}
}

func TestCreatePolygonObject(t *testing.T) {
	rpt := reportpkg.NewReport()
	page := importpkg.CreateReportPage(rpt)
	db := importpkg.CreateDataBand(page)

	obj := importpkg.CreatePolygonObject("Polygon1", db)
	if obj.Name() != "Polygon1" {
		t.Fatalf("expected %q, got %q", "Polygon1", obj.Name())
	}
}

func TestCreateSubreportObject(t *testing.T) {
	rpt := reportpkg.NewReport()
	page := importpkg.CreateReportPage(rpt)
	db := importpkg.CreateDataBand(page)

	obj := importpkg.CreateSubreportObject("Sub1", db)
	if obj.Name() != "Sub1" {
		t.Fatalf("expected %q, got %q", "Sub1", obj.Name())
	}
}

func TestCreateContainerObject(t *testing.T) {
	rpt := reportpkg.NewReport()
	page := importpkg.CreateReportPage(rpt)
	db := importpkg.CreateDataBand(page)

	obj := importpkg.CreateContainerObject("Container1", db)
	if obj.Name() != "Container1" {
		t.Fatalf("expected %q, got %q", "Container1", obj.Name())
	}
}

func TestCreateCheckBoxObject(t *testing.T) {
	rpt := reportpkg.NewReport()
	page := importpkg.CreateReportPage(rpt)
	db := importpkg.CreateDataBand(page)

	obj := importpkg.CreateCheckBoxObject("Check1", db)
	if obj.Name() != "Check1" {
		t.Fatalf("expected %q, got %q", "Check1", obj.Name())
	}
}

func TestCreateHtmlObject(t *testing.T) {
	rpt := reportpkg.NewReport()
	page := importpkg.CreateReportPage(rpt)
	db := importpkg.CreateDataBand(page)

	obj := importpkg.CreateHtmlObject("Html1", db)
	if obj.Name() != "Html1" {
		t.Fatalf("expected %q, got %q", "Html1", obj.Name())
	}
}

func TestCreateTableObject(t *testing.T) {
	rpt := reportpkg.NewReport()
	page := importpkg.CreateReportPage(rpt)
	db := importpkg.CreateDataBand(page)

	obj := importpkg.CreateTableObject("Table1", db)
	if obj.Name() != "Table1" {
		t.Fatalf("expected %q, got %q", "Table1", obj.Name())
	}
}

func TestCreateMatrixObject(t *testing.T) {
	rpt := reportpkg.NewReport()
	page := importpkg.CreateReportPage(rpt)
	db := importpkg.CreateDataBand(page)

	obj := importpkg.CreateMatrixObject("Matrix1", db)
	if obj.Name() != "Matrix1" {
		t.Fatalf("expected %q, got %q", "Matrix1", obj.Name())
	}
}

func TestCreateBarcodeObject(t *testing.T) {
	rpt := reportpkg.NewReport()
	page := importpkg.CreateReportPage(rpt)
	db := importpkg.CreateDataBand(page)

	obj := importpkg.CreateBarcodeObject("Barcode1", db)
	if obj.Name() != "Barcode1" {
		t.Fatalf("expected %q, got %q", "Barcode1", obj.Name())
	}
}

func TestCreateZipCodeObject(t *testing.T) {
	rpt := reportpkg.NewReport()
	page := importpkg.CreateReportPage(rpt)
	db := importpkg.CreateDataBand(page)

	obj := importpkg.CreateZipCodeObject("Zip1", db)
	if obj.Name() != "Zip1" {
		t.Fatalf("expected %q, got %q", "Zip1", obj.Name())
	}
}

func TestCreateCellularTextObject(t *testing.T) {
	rpt := reportpkg.NewReport()
	page := importpkg.CreateReportPage(rpt)
	db := importpkg.CreateDataBand(page)

	obj := importpkg.CreateCellularTextObject("Cell1", db)
	if obj.Name() != "Cell1" {
		t.Fatalf("expected %q, got %q", "Cell1", obj.Name())
	}
}

func TestCreateLinearGauge(t *testing.T) {
	rpt := reportpkg.NewReport()
	page := importpkg.CreateReportPage(rpt)
	db := importpkg.CreateDataBand(page)

	obj := importpkg.CreateLinearGauge("LinearGauge1", db)
	if obj.Name() != "LinearGauge1" {
		t.Fatalf("expected %q, got %q", "LinearGauge1", obj.Name())
	}
}

func TestCreateSimpleGauge(t *testing.T) {
	rpt := reportpkg.NewReport()
	page := importpkg.CreateReportPage(rpt)
	db := importpkg.CreateDataBand(page)

	obj := importpkg.CreateSimpleGauge("SimpleGauge1", db)
	if obj.Name() != "SimpleGauge1" {
		t.Fatalf("expected %q, got %q", "SimpleGauge1", obj.Name())
	}
}

func TestCreateRadialGauge(t *testing.T) {
	rpt := reportpkg.NewReport()
	page := importpkg.CreateReportPage(rpt)
	db := importpkg.CreateDataBand(page)

	obj := importpkg.CreateRadialGauge("RadialGauge1", db)
	if obj.Name() != "RadialGauge1" {
		t.Fatalf("expected %q, got %q", "RadialGauge1", obj.Name())
	}
}

func TestCreateSimpleProgressGauge(t *testing.T) {
	rpt := reportpkg.NewReport()
	page := importpkg.CreateReportPage(rpt)
	db := importpkg.CreateDataBand(page)

	obj := importpkg.CreateSimpleProgressGauge("Progress1", db)
	if obj.Name() != "Progress1" {
		t.Fatalf("expected %q, got %q", "Progress1", obj.Name())
	}
}

// ── Style ────────────────────────────────────────────────────────────────────

func TestCreateStyle(t *testing.T) {
	rpt := reportpkg.NewReport()
	s := importpkg.CreateStyle("HeaderStyle", rpt)

	if s == nil {
		t.Fatal("CreateStyle returned nil")
	}
	if s.Name != "HeaderStyle" {
		t.Fatalf("expected name %q, got %q", "HeaderStyle", s.Name)
	}
	found := rpt.Styles().Find("HeaderStyle")
	if found == nil {
		t.Fatal("style not registered in report.Styles")
	}
}

// ── Dictionary Elements ──────────────────────────────────────────────────────

func TestCreateParameter(t *testing.T) {
	rpt := reportpkg.NewReport()
	p := importpkg.CreateParameter("StartDate", rpt)

	if p == nil {
		t.Fatal("CreateParameter returned nil")
	}
	if p.Name != "StartDate" {
		t.Fatalf("expected name %q, got %q", "StartDate", p.Name)
	}
	found := rpt.Dictionary().FindParameter("StartDate")
	if found == nil {
		t.Fatal("parameter not registered in report.Dictionary")
	}
}

// ── Full workflow test ────────────────────────────────────────────────────────

// TestFullImportWorkflow simulates a minimal importer that builds a report
// with a page, common bands, and a text object — the typical use-case of
// ComponentsFactory.
func TestFullImportWorkflow(t *testing.T) {
	rpt := reportpkg.NewReport()

	page := importpkg.CreateReportPageNamed("Page1", rpt)
	_ = importpkg.CreatePageHeaderBand(page)
	_ = importpkg.CreatePageFooterBand(page)
	_ = importpkg.CreateReportTitleBand(page)

	db := importpkg.CreateDataBand(page)
	_ = importpkg.CreateDataHeaderBand(db)
	_ = importpkg.CreateDataFooterBand(db)

	txt := importpkg.CreateTextObject("ProductName", db)
	_ = importpkg.CreateStyle("DataStyle", rpt)
	_ = importpkg.CreateParameter("Year", rpt)

	if rpt.PageCount() != 1 {
		t.Fatalf("expected 1 page, got %d", rpt.PageCount())
	}
	if page.PageHeader() == nil {
		t.Fatal("PageHeader not set")
	}
	if page.PageFooter() == nil {
		t.Fatal("PageFooter not set")
	}
	if page.ReportTitle() == nil {
		t.Fatal("ReportTitle not set")
	}
	if len(page.Bands()) != 1 {
		t.Fatalf("expected 1 data band, got %d", len(page.Bands()))
	}
	if db.Header() == nil {
		t.Fatal("DataHeaderBand not set")
	}
	if db.Footer() == nil {
		t.Fatal("DataFooterBand not set")
	}
	if txt.Name() != "ProductName" {
		t.Fatalf("TextObject name wrong: %q", txt.Name())
	}
	if rpt.Styles().Find("DataStyle") == nil {
		t.Fatal("style not registered")
	}
	if rpt.Dictionary().FindParameter("Year") == nil {
		t.Fatal("parameter not registered")
	}
}
