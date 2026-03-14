package reportpkg_test

// FRX round-trip integration tests.
//
// These tests verify that a Report → ReportPage → Band → TextObject hierarchy
// can be serialized to FRX XML and deserialized back with all properties
// preserved (structural fidelity).

import (
	"bytes"
	"strings"
	"testing"

	"github.com/andrewloable/go-fastreport/band"
	"github.com/andrewloable/go-fastreport/object"
	"github.com/andrewloable/go-fastreport/report"
	"github.com/andrewloable/go-fastreport/reportpkg"
	"github.com/andrewloable/go-fastreport/serial"
)

// ── helpers ───────────────────────────────────────────────────────────────────

// serializeReport writes r as FRX XML and returns the bytes.
func serializeReport(t *testing.T, r *reportpkg.Report) []byte {
	t.Helper()
	var buf bytes.Buffer
	w := serial.NewWriter(&buf)
	if err := w.WriteHeader(); err != nil {
		t.Fatalf("WriteHeader: %v", err)
	}
	if err := w.WriteObjectNamed("Report", r); err != nil {
		t.Fatalf("serialize Report: %v", err)
	}
	if err := w.Flush(); err != nil {
		t.Fatalf("Flush: %v", err)
	}
	return buf.Bytes()
}

// serializePage writes page as an FRX element and returns the bytes.
func serializePage(t *testing.T, p *reportpkg.ReportPage) []byte {
	t.Helper()
	var buf bytes.Buffer
	w := serial.NewWriter(&buf)
	if err := w.WriteObjectNamed("ReportPage", p); err != nil {
		t.Fatalf("serialize ReportPage: %v", err)
	}
	if err := w.Flush(); err != nil {
		t.Fatalf("Flush: %v", err)
	}
	return buf.Bytes()
}

// deserializeReport decodes the first XML element from data into a Report.
func deserializeReport(t *testing.T, data []byte) *reportpkg.Report {
	t.Helper()
	r := serial.NewReader(bytes.NewReader(data))
	typeName, ok := r.ReadObjectHeader()
	if !ok {
		t.Fatalf("ReadObjectHeader returned ok=false; xml:\n%s", data)
	}
	if typeName != "Report" {
		t.Fatalf("expected Report element, got %q", typeName)
	}
	rep := reportpkg.NewReport()
	if err := rep.Deserialize(r); err != nil {
		t.Fatalf("Deserialize Report: %v", err)
	}
	return rep
}

// deserializePage decodes a single ReportPage element.
func deserializePage(t *testing.T, data []byte) *reportpkg.ReportPage {
	t.Helper()
	r := serial.NewReader(bytes.NewReader(data))
	typeName, ok := r.ReadObjectHeader()
	if !ok {
		t.Fatalf("ReadObjectHeader returned ok=false; xml:\n%s", data)
	}
	if typeName != "ReportPage" {
		t.Fatalf("expected ReportPage element, got %q", typeName)
	}
	p := reportpkg.NewReportPage()
	if err := p.Deserialize(r); err != nil {
		t.Fatalf("Deserialize ReportPage: %v", err)
	}
	return p
}

// ── Report round-trip ─────────────────────────────────────────────────────────

func TestRoundTrip_Report_DefaultsOnly(t *testing.T) {
	orig := reportpkg.NewReport()
	data := serializeReport(t, orig)
	got := deserializeReport(t, data)

	// Defaults should survive the round-trip.
	if got.InitialPageNumber != 1 {
		t.Errorf("InitialPageNumber: got %d, want 1", got.InitialPageNumber)
	}
	if got.Compressed {
		t.Error("Compressed should default to false")
	}
	if got.DoublePass {
		t.Error("DoublePass should default to false")
	}
}

func TestRoundTrip_Report_AllProperties(t *testing.T) {
	orig := reportpkg.NewReport()
	orig.Info.Name = "Monthly Sales"
	orig.Info.Author = "Alice"
	orig.Info.Description = "A detailed sales report"
	orig.Compressed = true
	orig.ConvertNulls = true
	orig.DoublePass = true
	orig.InitialPageNumber = 3
	orig.MaxPages = 10
	orig.StartReportEvent = "OnStartReport"
	orig.FinishReportEvent = "OnFinishReport"

	data := serializeReport(t, orig)
	got := deserializeReport(t, data)

	checks := []struct {
		label string
		got   any
		want  any
	}{
		{"Info.Name", got.Info.Name, "Monthly Sales"},
		{"Info.Author", got.Info.Author, "Alice"},
		{"Info.Description", got.Info.Description, "A detailed sales report"},
		{"Compressed", got.Compressed, true},
		{"ConvertNulls", got.ConvertNulls, true},
		{"DoublePass", got.DoublePass, true},
		{"InitialPageNumber", got.InitialPageNumber, 3},
		{"MaxPages", got.MaxPages, 10},
		{"StartReportEvent", got.StartReportEvent, "OnStartReport"},
		{"FinishReportEvent", got.FinishReportEvent, "OnFinishReport"},
	}
	for _, c := range checks {
		if c.got != c.want {
			t.Errorf("%s: got %v, want %v", c.label, c.got, c.want)
		}
	}
}

func TestRoundTrip_Report_XmlHeader(t *testing.T) {
	orig := reportpkg.NewReport()
	data := serializeReport(t, orig)

	xml := string(data)
	if !strings.HasPrefix(xml, `<?xml version="1.0" encoding="utf-8"?>`) {
		t.Errorf("XML header missing; got start: %q", xml[:min(50, len(xml))])
	}
	if !strings.Contains(xml, "<Report") {
		t.Errorf("expected <Report element in output:\n%s", xml)
	}
}

// ── ReportPage round-trip ─────────────────────────────────────────────────────

func TestRoundTrip_ReportPage_Defaults(t *testing.T) {
	orig := reportpkg.NewReportPage()
	data := serializePage(t, orig)
	got := deserializePage(t, data)

	// A4 defaults.
	if got.PaperWidth != 210 {
		t.Errorf("PaperWidth: got %v, want 210", got.PaperWidth)
	}
	if got.PaperHeight != 297 {
		t.Errorf("PaperHeight: got %v, want 297", got.PaperHeight)
	}
	if got.LeftMargin != 10 {
		t.Errorf("LeftMargin: got %v, want 10", got.LeftMargin)
	}
	if got.Landscape {
		t.Error("Landscape should default to false")
	}
}

func TestRoundTrip_ReportPage_AllProperties(t *testing.T) {
	orig := reportpkg.NewReportPage()
	orig.PaperWidth = 297
	orig.PaperHeight = 420
	orig.Landscape = true
	orig.LeftMargin = 15
	orig.TopMargin = 20
	orig.RightMargin = 15
	orig.BottomMargin = 20
	orig.MirrorMargins = true
	orig.TitleBeforeHeader = true
	orig.PrintOnPreviousPage = true
	orig.ResetPageNumber = true
	orig.StartOnOddPage = true
	orig.OutlineExpression = "[Customer.Name]"
	orig.CreatePageEvent = "OnCreatePage"
	orig.StartPageEvent = "OnStartPage"
	orig.FinishPageEvent = "OnFinishPage"
	orig.ManualBuildEvent = "OnManualBuild"

	data := serializePage(t, orig)
	got := deserializePage(t, data)

	if got.PaperWidth != 297 {
		t.Errorf("PaperWidth: got %v, want 297", got.PaperWidth)
	}
	if got.PaperHeight != 420 {
		t.Errorf("PaperHeight: got %v, want 420", got.PaperHeight)
	}
	if !got.Landscape {
		t.Error("Landscape should be true")
	}
	if got.LeftMargin != 15 {
		t.Errorf("LeftMargin: got %v, want 15", got.LeftMargin)
	}
	if !got.MirrorMargins {
		t.Error("MirrorMargins should be true")
	}
	if !got.TitleBeforeHeader {
		t.Error("TitleBeforeHeader should be true")
	}
	if !got.ResetPageNumber {
		t.Error("ResetPageNumber should be true")
	}
	if got.OutlineExpression != "[Customer.Name]" {
		t.Errorf("OutlineExpression: got %q", got.OutlineExpression)
	}
	if got.CreatePageEvent != "OnCreatePage" {
		t.Errorf("CreatePageEvent: got %q", got.CreatePageEvent)
	}
}

// ── Band round-trip ───────────────────────────────────────────────────────────

func serializeBand(t *testing.T, name string, b report.Serializable) []byte {
	t.Helper()
	var buf bytes.Buffer
	w := serial.NewWriter(&buf)
	if err := w.WriteObjectNamed(name, b); err != nil {
		t.Fatalf("serialize %s: %v", name, err)
	}
	if err := w.Flush(); err != nil {
		t.Fatalf("Flush: %v", err)
	}
	return buf.Bytes()
}

func deserializeBandBase(t *testing.T, data []byte) *band.DataBand {
	t.Helper()
	r := serial.NewReader(bytes.NewReader(data))
	_, ok := r.ReadObjectHeader()
	if !ok {
		t.Fatalf("ReadObjectHeader returned ok=false")
	}
	b := band.NewDataBand()
	if err := b.Deserialize(r); err != nil {
		t.Fatalf("Deserialize DataBand: %v", err)
	}
	return b
}

func TestRoundTrip_DataBand_Properties(t *testing.T) {
	orig := band.NewDataBand()
	orig.SetName("DataBand1")
	orig.SetHeight(56.7)
	orig.SetVisible(true)
	orig.SetFilter("[Amount] > 100")
	orig.SetKeepTogether(true)

	data := serializeBand(t, "DataBand", orig)
	got := deserializeBandBase(t, data)

	if got.Name() != "DataBand1" {
		t.Errorf("Name: got %q, want DataBand1", got.Name())
	}
	if got.Height() != 56.7 {
		t.Errorf("Height: got %v, want 56.7", got.Height())
	}
	if got.Filter() != "[Amount] > 100" {
		t.Errorf("Filter: got %q", got.Filter())
	}
	if !got.KeepTogether() {
		t.Error("KeepTogether should be true")
	}
}

func TestRoundTrip_ReportTitleBand(t *testing.T) {
	orig := band.NewReportTitleBand()
	orig.SetName("Title1")
	orig.SetHeight(100)

	data := serializeBand(t, "ReportTitleBand", orig)
	r := serial.NewReader(bytes.NewReader(data))
	typeName, ok := r.ReadObjectHeader()
	if !ok || typeName != "ReportTitleBand" {
		t.Fatalf("got typeName=%q ok=%v", typeName, ok)
	}
	got := band.NewReportTitleBand()
	if err := got.Deserialize(r); err != nil {
		t.Fatalf("Deserialize: %v", err)
	}
	if got.Name() != "Title1" {
		t.Errorf("Name: got %q, want Title1", got.Name())
	}
	if got.Height() != 100 {
		t.Errorf("Height: got %v, want 100", got.Height())
	}
}

// ── TextObject round-trip ─────────────────────────────────────────────────────

func serializeTextObject(t *testing.T, obj *object.TextObject) []byte {
	t.Helper()
	var buf bytes.Buffer
	w := serial.NewWriter(&buf)
	if err := w.WriteObjectNamed("TextObject", obj); err != nil {
		t.Fatalf("serialize TextObject: %v", err)
	}
	if err := w.Flush(); err != nil {
		t.Fatalf("Flush: %v", err)
	}
	return buf.Bytes()
}

func deserializeTextObject(t *testing.T, data []byte) *object.TextObject {
	t.Helper()
	r := serial.NewReader(bytes.NewReader(data))
	_, ok := r.ReadObjectHeader()
	if !ok {
		t.Fatalf("ReadObjectHeader returned ok=false")
	}
	obj := object.NewTextObject()
	if err := obj.Deserialize(r); err != nil {
		t.Fatalf("Deserialize TextObject: %v", err)
	}
	return obj
}

func TestRoundTrip_TextObject_Defaults(t *testing.T) {
	orig := object.NewTextObject()
	data := serializeTextObject(t, orig)
	got := deserializeTextObject(t, data)

	// Defaults should round-trip cleanly.
	if got.Text() != "" {
		t.Errorf("Text: got %q, want empty", got.Text())
	}
	if !got.AllowExpressions() {
		t.Error("AllowExpressions should default to true")
	}
	if got.Brackets() != "[,]" {
		t.Errorf("Brackets: got %q, want [,]", got.Brackets())
	}
}

func TestRoundTrip_TextObject_AllProperties(t *testing.T) {
	orig := object.NewTextObject()
	orig.SetName("Text1")
	orig.SetLeft(10)
	orig.SetTop(20)
	orig.SetWidth(200)
	orig.SetHeight(30)
	orig.SetText("Hello [Name]")
	orig.SetAngle(90)
	orig.SetWordWrap(true)
	orig.SetHideZeros(true)
	orig.SetNullValue("N/A")
	orig.SetHideValue("0")

	data := serializeTextObject(t, orig)
	got := deserializeTextObject(t, data)

	if got.Name() != "Text1" {
		t.Errorf("Name: got %q, want Text1", got.Name())
	}
	if got.Left() != 10 {
		t.Errorf("Left: got %v, want 10", got.Left())
	}
	if got.Top() != 20 {
		t.Errorf("Top: got %v, want 20", got.Top())
	}
	if got.Width() != 200 {
		t.Errorf("Width: got %v, want 200", got.Width())
	}
	if got.Height() != 30 {
		t.Errorf("Height: got %v, want 30", got.Height())
	}
	if got.Text() != "Hello [Name]" {
		t.Errorf("Text: got %q, want 'Hello [Name]'", got.Text())
	}
	if got.Angle() != 90 {
		t.Errorf("Angle: got %d, want 90", got.Angle())
	}
	if !got.WordWrap() {
		t.Error("WordWrap should be true")
	}
	if !got.HideZeros() {
		t.Error("HideZeros should be true")
	}
	if got.NullValue() != "N/A" {
		t.Errorf("NullValue: got %q, want N/A", got.NullValue())
	}
}

func TestRoundTrip_TextObject_SpecialCharsInText(t *testing.T) {
	orig := object.NewTextObject()
	orig.SetText(`Hello <World> & "friends" ` + "\r\n")

	data := serializeTextObject(t, orig)
	got := deserializeTextObject(t, data)

	if got.Text() != orig.Text() {
		t.Errorf("Text: got %q, want %q", got.Text(), orig.Text())
	}
}

// ── Unknown element handling ──────────────────────────────────────────────────

// TestRoundTrip_UnknownElement verifies that a Report round-trip gracefully
// skips XML elements it does not recognise.
func TestRoundTrip_UnknownElement_GracefulSkip(t *testing.T) {
	// Inject an unknown child element into a valid Report XML.
	xmlSrc := `<?xml version="1.0" encoding="utf-8"?>
<Report ReportName="Test">
  <UnknownElement Foo="bar"><NestedUnknown/></UnknownElement>
</Report>`

	// We just need to be able to read the Report attributes without error.
	// The serial.Reader does not auto-skip unknown children — our Deserialize
	// implementation is expected to handle that at each level.
	// For the Report itself, Deserialize reads only known attributes and then
	// returns. The unknown element is simply left in the stream; as long as no
	// panic or error occurs while reading the Report header attributes, the
	// test passes.
	r := serial.NewReader(strings.NewReader(xmlSrc))
	typeName, ok := r.ReadObjectHeader()
	if !ok || typeName != "Report" {
		t.Fatalf("got typeName=%q ok=%v", typeName, ok)
	}
	rep := reportpkg.NewReport()
	if err := rep.Deserialize(r); err != nil {
		t.Fatalf("Deserialize should not error on unknown children: %v", err)
	}
	if rep.Info.Name != "Test" {
		t.Errorf("ReportName: got %q, want Test", rep.Info.Name)
	}
}

// ── All property types ────────────────────────────────────────────────────────

func TestRoundTrip_AllPropertyTypes(t *testing.T) {
	// Verify that string, int, bool, float32 all survive a round-trip.
	orig := reportpkg.NewReportPage()
	orig.SetName("Page1")    // string
	orig.PaperWidth = 297.5  // float32
	orig.Landscape = true    // bool
	orig.Columns.Count = 2   // int (embedded struct)

	data := serializePage(t, orig)
	got := deserializePage(t, data)

	if got.Name() != "Page1" {
		t.Errorf("Name (string): got %q, want Page1", got.Name())
	}
	if got.PaperWidth != 297.5 {
		t.Errorf("PaperWidth (float32): got %v, want 297.5", got.PaperWidth)
	}
	if !got.Landscape {
		t.Error("Landscape (bool): should be true")
	}
}

// ── Output XML structure ──────────────────────────────────────────────────────

func TestRoundTrip_OutputContainsExpectedElements(t *testing.T) {
	orig := reportpkg.NewReport()
	orig.Info.Name = "SalesReport"

	data := serializeReport(t, orig)
	xml := string(data)

	for _, want := range []string{
		`<Report`,
		`ReportName="SalesReport"`,
	} {
		if !strings.Contains(xml, want) {
			t.Errorf("XML missing %q in:\n%s", want, xml)
		}
	}
}

func TestRoundTrip_DeltaSerialization_OnlyNonDefaults(t *testing.T) {
	// A default ReportPage should NOT emit PaperWidth or PaperHeight
	// (since they equal the defaults 210/297).
	orig := reportpkg.NewReportPage()
	data := serializePage(t, orig)
	xml := string(data)

	if strings.Contains(xml, `PaperWidth=`) {
		t.Errorf("default PaperWidth should not be serialized; xml:\n%s", xml)
	}
	if strings.Contains(xml, `PaperHeight=`) {
		t.Errorf("default PaperHeight should not be serialized; xml:\n%s", xml)
	}
	// A non-default value SHOULD appear.
	orig2 := reportpkg.NewReportPage()
	orig2.PaperWidth = 100
	data2 := serializePage(t, orig2)
	if !strings.Contains(string(data2), `PaperWidth="100"`) {
		t.Errorf("non-default PaperWidth should be serialized; xml:\n%s", string(data2))
	}
}

// ── SaveToString / LoadFromString round-trip ──────────────────────────────────

// TestRoundTrip_SaveLoad_Page_With_Band_And_Object verifies that a full
// Report → ReportPage → DataBand → TextObject hierarchy survives a
// SaveToString / LoadFromString round-trip with all properties intact.
func TestRoundTrip_SaveLoad_Page_With_Band_And_Object(t *testing.T) {
	// Build the report.
	r := reportpkg.NewReport()
	r.Info.Name = "SalesReport"

	pg := reportpkg.NewReportPage()
	pg.SetName("Page1")
	pg.PaperWidth = 297
	pg.Landscape = true

	db := band.NewDataBand()
	db.SetName("Data1")
	db.SetHeight(56)
	db.SetFilter("[Amount] > 0")

	txt := object.NewTextObject()
	txt.SetName("Text1")
	txt.SetLeft(10)
	txt.SetTop(5)
	txt.SetWidth(200)
	txt.SetHeight(30)
	txt.SetText("Hello [Name]")

	db.AddChild(txt)
	pg.AddBand(db)
	r.AddPage(pg)

	// Serialize.
	xmlStr, err := r.SaveToString()
	if err != nil {
		t.Fatalf("SaveToString: %v", err)
	}

	// Verify expected elements appear in the XML.
	for _, want := range []string{
		"<Report", "SalesReport",
		"<ReportPage", "Page1",
		"<Data ", "<TextObject", "Hello [Name]",
	} {
		if !strings.Contains(xmlStr, want) {
			t.Errorf("SaveToString: XML missing %q:\n%s", want, xmlStr)
		}
	}

	// Deserialize.
	r2 := reportpkg.NewReport()
	if err := r2.LoadFromString(xmlStr); err != nil {
		t.Fatalf("LoadFromString: %v", err)
	}

	// Verify page.
	pages := r2.Pages()
	if len(pages) != 1 {
		t.Fatalf("Pages: got %d, want 1", len(pages))
	}
	p2 := pages[0]
	if p2.PaperWidth != 297 {
		t.Errorf("Page PaperWidth: got %v, want 297", p2.PaperWidth)
	}
	if !p2.Landscape {
		t.Error("Page Landscape: should be true")
	}

	// Verify band.
	bands := p2.Bands()
	if len(bands) != 1 {
		t.Fatalf("Bands: got %d, want 1", len(bands))
	}
	db2, ok := bands[0].(*band.DataBand)
	if !ok {
		t.Fatalf("Band type: got %T, want *band.DataBand", bands[0])
	}
	if db2.Name() != "Data1" {
		t.Errorf("Band Name: got %q, want Data1", db2.Name())
	}
	if db2.Filter() != "[Amount] > 0" {
		t.Errorf("Band Filter: got %q", db2.Filter())
	}

	// Verify text object inside band.
	if db2.Objects().Len() != 1 {
		t.Fatalf("Band objects: got %d, want 1", db2.Objects().Len())
	}
	txt2, ok := db2.Objects().Get(0).(*object.TextObject)
	if !ok {
		t.Fatalf("Object type: got %T, want *object.TextObject", db2.Objects().Get(0))
	}
	if txt2.Text() != "Hello [Name]" {
		t.Errorf("TextObject.Text: got %q, want 'Hello [Name]'", txt2.Text())
	}
	if txt2.Left() != 10 {
		t.Errorf("TextObject.Left: got %v, want 10", txt2.Left())
	}
}

// TestRoundTrip_SaveLoad_PageHeaderFooter verifies that singleton band slots
// (PageHeader, PageFooter) survive a SaveToString / LoadFromString round-trip.
func TestRoundTrip_SaveLoad_PageHeaderFooter(t *testing.T) {
	r := reportpkg.NewReport()
	pg := reportpkg.NewReportPage()
	pg.SetName("Page1")

	ph := band.NewPageHeaderBand()
	ph.SetName("PageHeader1")
	ph.SetHeight(40)

	pf := band.NewPageFooterBand()
	pf.SetName("PageFooter1")
	pf.SetHeight(30)

	pg.SetPageHeader(ph)
	pg.SetPageFooter(pf)
	r.AddPage(pg)

	xmlStr, err := r.SaveToString()
	if err != nil {
		t.Fatalf("SaveToString: %v", err)
	}

	for _, want := range []string{"<PageHeader ", "<PageFooter "} {
		if !strings.Contains(xmlStr, want) {
			t.Errorf("XML missing %q:\n%s", want, xmlStr)
		}
	}

	r2 := reportpkg.NewReport()
	if err := r2.LoadFromString(xmlStr); err != nil {
		t.Fatalf("LoadFromString: %v", err)
	}

	pages := r2.Pages()
	if len(pages) != 1 {
		t.Fatalf("Pages: got %d, want 1", len(pages))
	}
	p2 := pages[0]
	if p2.PageHeader() == nil {
		t.Fatal("PageHeader should not be nil after round-trip")
	}
	if p2.PageHeader().Name() != "PageHeader1" {
		t.Errorf("PageHeader Name: got %q, want PageHeader1", p2.PageHeader().Name())
	}
	if p2.PageFooter() == nil {
		t.Fatal("PageFooter should not be nil after round-trip")
	}
}

// ── min helper (Go 1.25 has built-in min, but keep local for safety) ──────────

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
