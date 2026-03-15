package export_test

import (
	"bytes"
	"fmt"
	"strings"
	"testing"

	"github.com/andrewloable/go-fastreport/export"
	"github.com/andrewloable/go-fastreport/preview"
	"github.com/andrewloable/go-fastreport/report"
)

// ── ParsePageNumbers ───────────────────────────────────────────────────────────

func TestParsePageNumbers_Empty(t *testing.T) {
	result, err := export.ParsePageNumbers("", 10)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != nil {
		t.Errorf("empty string should return nil, got %v", result)
	}
}

func TestParsePageNumbers_Single(t *testing.T) {
	result, err := export.ParsePageNumbers("3", 10)
	if err != nil {
		t.Fatal(err)
	}
	if len(result) != 1 || result[0] != 2 {
		t.Errorf("got %v, want [2]", result)
	}
}

func TestParsePageNumbers_Range(t *testing.T) {
	result, err := export.ParsePageNumbers("2-4", 10)
	if err != nil {
		t.Fatal(err)
	}
	if fmt.Sprint(result) != "[1 2 3]" {
		t.Errorf("got %v, want [1 2 3]", result)
	}
}

func TestParsePageNumbers_Mixed(t *testing.T) {
	result, err := export.ParsePageNumbers("1,3-5,12", 15)
	if err != nil {
		t.Fatal(err)
	}
	want := "[0 2 3 4 11]"
	if fmt.Sprint(result) != want {
		t.Errorf("got %v, want %v", result, want)
	}
}

func TestParsePageNumbers_TrailingDash(t *testing.T) {
	result, err := export.ParsePageNumbers("8-", 10)
	if err != nil {
		t.Fatal(err)
	}
	if fmt.Sprint(result) != "[7 8 9]" {
		t.Errorf("got %v, want [7 8 9]", result)
	}
}

func TestParsePageNumbers_Spaces(t *testing.T) {
	result, err := export.ParsePageNumbers(" 1 , 2 ", 10)
	if err != nil {
		t.Fatal(err)
	}
	if fmt.Sprint(result) != "[0 1]" {
		t.Errorf("got %v, want [0 1]", result)
	}
}

func TestParsePageNumbers_Invalid(t *testing.T) {
	_, err := export.ParsePageNumbers("1,foo,3", 10)
	if err == nil {
		t.Error("expected error for invalid page number")
	}
}

func TestParsePageNumbers_NoDuplicates(t *testing.T) {
	result, err := export.ParsePageNumbers("1,1,2-3,2", 10)
	if err != nil {
		t.Fatal(err)
	}
	if len(result) != 3 {
		t.Errorf("got %d pages, want 3 (de-duplicated): %v", len(result), result)
	}
}

// ── ExportBase.Export ─────────────────────────────────────────────────────────

// recorder records calls to the export hooks.
type recorder struct {
	export.NoopExporter
	started   bool
	finished  bool
	pageBegin []int // pageNo values
	pageEnd   []int
	bandNames []string
	w         *bytes.Buffer
}

func newRecorder(w *bytes.Buffer) *recorder {
	return &recorder{w: w}
}

func (r *recorder) Start() error {
	r.started = true
	return nil
}

func (r *recorder) Finish() error {
	r.finished = true
	return nil
}

func (r *recorder) ExportPageBegin(pg *preview.PreparedPage) error {
	r.pageBegin = append(r.pageBegin, pg.PageNo)
	return nil
}

func (r *recorder) ExportPageEnd(pg *preview.PreparedPage) error {
	r.pageEnd = append(r.pageEnd, pg.PageNo)
	return nil
}

func (r *recorder) ExportBand(b *preview.PreparedBand) error {
	r.bandNames = append(r.bandNames, b.Name)
	return nil
}

func buildPreparedPages(pages int, bandsPerPage []string) *preview.PreparedPages {
	pp := preview.New()
	for i := 0; i < pages; i++ {
		pp.AddPage(595, 842, i+1)
		for _, name := range bandsPerPage {
			_ = pp.AddBand(&preview.PreparedBand{Name: fmt.Sprintf("%s_p%d", name, i+1), Top: 0, Height: 30})
		}
	}
	return pp
}

func TestExport_AllPages(t *testing.T) {
	pp := buildPreparedPages(3, []string{"Header", "Data"})
	base := export.NewExportBase()
	rec := newRecorder(new(bytes.Buffer))

	if err := base.Export(pp, rec.w, rec); err != nil {
		t.Fatalf("Export: %v", err)
	}

	if !rec.started {
		t.Error("Start() not called")
	}
	if !rec.finished {
		t.Error("Finish() not called")
	}
	if len(rec.pageBegin) != 3 {
		t.Errorf("ExportPageBegin calls = %d, want 3", len(rec.pageBegin))
	}
	if len(rec.bandNames) != 6 { // 2 bands × 3 pages
		t.Errorf("ExportBand calls = %d, want 6", len(rec.bandNames))
	}
}

func TestExport_CurrentPage(t *testing.T) {
	pp := buildPreparedPages(5, []string{"Band"})
	base := export.NewExportBase()
	base.PageRange = export.PageRangeCurrent
	base.CurPage = 3 // 1-based
	rec := newRecorder(new(bytes.Buffer))

	if err := base.Export(pp, rec.w, rec); err != nil {
		t.Fatalf("Export: %v", err)
	}

	if len(rec.pageBegin) != 1 {
		t.Errorf("should export 1 page, got %d", len(rec.pageBegin))
	}
	if rec.pageBegin[0] != 3 {
		t.Errorf("exported page %d, want 3", rec.pageBegin[0])
	}
}

func TestExport_CustomPageNumbers(t *testing.T) {
	pp := buildPreparedPages(10, []string{"B"})
	base := export.NewExportBase()
	base.PageRange = export.PageRangeCustom
	base.PageNumbers = "1,3-5"
	rec := newRecorder(new(bytes.Buffer))

	if err := base.Export(pp, rec.w, rec); err != nil {
		t.Fatalf("Export: %v", err)
	}

	if len(rec.pageBegin) != 4 {
		t.Errorf("should export 4 pages (1,3,4,5), got %d: %v", len(rec.pageBegin), rec.pageBegin)
	}
}

func TestExport_OutOfRangePageIgnored(t *testing.T) {
	pp := buildPreparedPages(3, []string{"B"})
	base := export.NewExportBase()
	base.PageRange = export.PageRangeCustom
	base.PageNumbers = "1,99" // 99 is out of range
	rec := newRecorder(new(bytes.Buffer))

	if err := base.Export(pp, rec.w, rec); err != nil {
		t.Fatalf("Export: %v", err)
	}

	if len(rec.pageBegin) != 1 {
		t.Errorf("only 1 valid page, got %d", len(rec.pageBegin))
	}
}

func TestExport_NilPages(t *testing.T) {
	base := export.NewExportBase()
	rec := newRecorder(new(bytes.Buffer))
	err := base.Export(nil, rec.w, rec)
	if err == nil {
		t.Error("expected error for nil PreparedPages")
	}
}

func TestExport_EmptyPages(t *testing.T) {
	pp := preview.New()
	base := export.NewExportBase()
	rec := newRecorder(new(bytes.Buffer))

	if err := base.Export(pp, rec.w, rec); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(rec.pageBegin) != 0 {
		t.Errorf("expected 0 pages, got %d", len(rec.pageBegin))
	}
}

// ── Error propagation ─────────────────────────────────────────────────────────

type errExporter struct {
	export.NoopExporter
	failOn string
}

func (e *errExporter) Start() error {
	if e.failOn == "start" {
		return fmt.Errorf("start error")
	}
	return nil
}

func (e *errExporter) ExportBand(b *preview.PreparedBand) error {
	if e.failOn == "band" {
		return fmt.Errorf("band error")
	}
	return nil
}

func (e *errExporter) Finish() error {
	if e.failOn == "finish" {
		return fmt.Errorf("finish error")
	}
	return nil
}

func TestExport_StartError(t *testing.T) {
	pp := buildPreparedPages(1, []string{"B"})
	base := export.NewExportBase()
	exp := &errExporter{failOn: "start"}
	err := base.Export(pp, new(bytes.Buffer), exp)
	if err == nil || !strings.Contains(err.Error(), "start") {
		t.Errorf("expected start error, got %v", err)
	}
}

func TestExport_BandError(t *testing.T) {
	pp := buildPreparedPages(1, []string{"B"})
	base := export.NewExportBase()
	exp := &errExporter{failOn: "band"}
	err := base.Export(pp, new(bytes.Buffer), exp)
	if err == nil || !strings.Contains(err.Error(), "band") {
		t.Errorf("expected band error, got %v", err)
	}
}

// ── Utils ─────────────────────────────────────────────────────────────────────

func TestPixelsToMM(t *testing.T) {
	mm := export.PixelsToMM(96)
	if mm < 25.3 || mm > 25.5 {
		t.Errorf("PixelsToMM(96) = %v, want ~25.4", mm)
	}
}

func TestMMToPixels(t *testing.T) {
	px := export.MMToPixels(25.4)
	if px < 95.9 || px > 96.1 {
		t.Errorf("MMToPixels(25.4) = %v, want ~96", px)
	}
}

func TestPixelsToPoints(t *testing.T) {
	pt := export.PixelsToPoints(96)
	if pt < 71.9 || pt > 72.1 {
		t.Errorf("PixelsToPoints(96) = %v, want ~72", pt)
	}
}

func TestHTMLString(t *testing.T) {
	cases := []struct{ in, want string }{
		{"hello", "hello"},
		{"a & b", "a &amp; b"},
		{"<tag>", "&lt;tag&gt;"},
		{`"quote"`, "&quot;quote&quot;"},
	}
	for _, c := range cases {
		got := export.HTMLString(c.in)
		if got != c.want {
			t.Errorf("HTMLString(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}

func TestXMLString(t *testing.T) {
	got := export.XMLString("<a>&b</a>\n")
	if !strings.Contains(got, "&lt;") || !strings.Contains(got, "&amp;") || !strings.Contains(got, "&#xA;") {
		t.Errorf("XMLString output = %q", got)
	}
}

func TestRGBToHTMLColor(t *testing.T) {
	got := export.RGBToHTMLColor(255, 0, 128)
	if got != "#FF0080" {
		t.Errorf("got %q, want #FF0080", got)
	}
}

func TestHTMLColorToRGB(t *testing.T) {
	r, g, b, ok := export.HTMLColorToRGB("#FF0080")
	if !ok || r != 255 || g != 0 || b != 128 {
		t.Errorf("got %d %d %d %v", r, g, b, ok)
	}
}

func TestHTMLColorToRGB_Invalid(t *testing.T) {
	_, _, _, ok := export.HTMLColorToRGB("ZZZZZZ")
	if ok {
		t.Error("invalid color should return ok=false")
	}
}

func TestExcelColName(t *testing.T) {
	cases := []struct {
		col  int
		want string
	}{
		{0, "A"},
		{25, "Z"},
		{26, "AA"},
		{51, "AZ"},
		{52, "BA"},
	}
	for _, c := range cases {
		got := export.ExcelColName(c.col)
		if got != c.want {
			t.Errorf("ExcelColName(%d) = %q, want %q", c.col, got, c.want)
		}
	}
}

func TestExcelCellRef(t *testing.T) {
	got := export.ExcelCellRef(2, 5)
	if got != "C5" {
		t.Errorf("got %q, want C5", got)
	}
}

func TestFormatFloat(t *testing.T) {
	got := export.FormatFloat(3.14159, 2, false)
	if got != "3.14" {
		t.Errorf("got %q, want 3.14", got)
	}
	got = export.FormatFloat(3.10, 2, true)
	if got != "3.1" {
		t.Errorf("got %q, want 3.1 (stripped zeros)", got)
	}
}

func TestRound(t *testing.T) {
	if export.Round(3.14159, 2) != 3.14 {
		t.Errorf("Round(3.14159, 2) = %v", export.Round(3.14159, 2))
	}
	if export.Round(2.5, 0) != 3 {
		t.Errorf("Round(2.5, 0) = %v", export.Round(2.5, 0))
	}
}

// ── ParsePageNumbers edge cases ───────────────────────────────────────────────

func TestParsePageNumbers_EmptyPart_TrailingComma(t *testing.T) {
	// "1,,3" has an empty middle part → the `if part == ""` branch.
	result, err := export.ParsePageNumbers("1,,3", 10)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if fmt.Sprint(result) != "[0 2]" {
		t.Errorf("got %v, want [0 2]", result)
	}
}

func TestParsePageNumbers_ReversedRange(t *testing.T) {
	// "5-2" → reversed; should be treated as "2-5".
	result, err := export.ParsePageNumbers("5-2", 10)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if fmt.Sprint(result) != "[1 2 3 4]" {
		t.Errorf("got %v, want [1 2 3 4]", result)
	}
}

func TestParsePageNumbers_InvalidRangeStart(t *testing.T) {
	_, err := export.ParsePageNumbers("x-3", 10)
	if err == nil {
		t.Error("expected error for invalid range start")
	}
}

func TestParsePageNumbers_InvalidRangeEnd(t *testing.T) {
	_, err := export.ParsePageNumbers("1-z", 10)
	if err == nil {
		t.Error("expected error for invalid range end")
	}
}

// ── preparePageIndices edge cases ─────────────────────────────────────────────

func TestExport_CustomPageNumbers_Empty_AllPages(t *testing.T) {
	// PageRangeCustom with empty PageNumbers → ParsePageNumbers returns nil → all pages.
	pp := buildPreparedPages(3, []string{"B"})
	base := export.NewExportBase()
	base.PageRange = export.PageRangeCustom
	base.PageNumbers = "" // empty → all pages
	rec := newRecorder(new(bytes.Buffer))
	if err := base.Export(pp, rec.w, rec); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(rec.pageBegin) != 3 {
		t.Errorf("expected 3 pages, got %d", len(rec.pageBegin))
	}
}

func TestExport_CustomPageNumbers_InvalidError(t *testing.T) {
	// PageRangeCustom with invalid PageNumbers → preparePageIndices error.
	pp := buildPreparedPages(3, []string{"B"})
	base := export.NewExportBase()
	base.PageRange = export.PageRangeCustom
	base.PageNumbers = "bad"
	rec := newRecorder(new(bytes.Buffer))
	err := base.Export(pp, rec.w, rec)
	if err == nil {
		t.Error("expected error for invalid page numbers")
	}
}

// ── Export error paths ────────────────────────────────────────────────────────

type pageBeginErrExporter struct {
	export.NoopExporter
}

func (e *pageBeginErrExporter) ExportPageBegin(_ *preview.PreparedPage) error {
	return fmt.Errorf("pageBegin error")
}

type pageEndErrExporter struct {
	export.NoopExporter
}

func (e *pageEndErrExporter) ExportPageEnd(_ *preview.PreparedPage) error {
	return fmt.Errorf("pageEnd error")
}

type finishErrExporter struct {
	export.NoopExporter
}

func (e *finishErrExporter) Finish() error {
	return fmt.Errorf("finish error")
}

func TestExport_PageBeginError(t *testing.T) {
	pp := buildPreparedPages(1, []string{"B"})
	base := export.NewExportBase()
	err := base.Export(pp, new(bytes.Buffer), &pageBeginErrExporter{})
	if err == nil || !strings.Contains(err.Error(), "pageBegin") {
		t.Errorf("expected pageBegin error, got %v", err)
	}
}

func TestExport_PageEndError(t *testing.T) {
	pp := buildPreparedPages(1, []string{"B"})
	base := export.NewExportBase()
	err := base.Export(pp, new(bytes.Buffer), &pageEndErrExporter{})
	if err == nil || !strings.Contains(err.Error(), "pageEnd") {
		t.Errorf("expected pageEnd error, got %v", err)
	}
}

func TestExport_FinishError(t *testing.T) {
	pp := buildPreparedPages(1, []string{"B"})
	base := export.NewExportBase()
	err := base.Export(pp, new(bytes.Buffer), &finishErrExporter{})
	if err == nil || !strings.Contains(err.Error(), "finish") {
		t.Errorf("expected finish error, got %v", err)
	}
}

// ── Pages() ───────────────────────────────────────────────────────────────────

func TestExportBase_Pages(t *testing.T) {
	pp := buildPreparedPages(3, []string{"B"})
	base := export.NewExportBase()
	rec := newRecorder(new(bytes.Buffer))
	if err := base.Export(pp, rec.w, rec); err != nil {
		t.Fatalf("Export: %v", err)
	}
	pages := base.Pages()
	if len(pages) != 3 {
		t.Errorf("Pages() len = %d, want 3", len(pages))
	}
}

// ── NoopExporter ──────────────────────────────────────────────────────────────

func TestNoopExporter_AllMethods(t *testing.T) {
	n := export.NoopExporter{}
	if err := n.Start(); err != nil {
		t.Errorf("Start: %v", err)
	}
	if err := n.ExportPageBegin(nil); err != nil {
		t.Errorf("ExportPageBegin: %v", err)
	}
	if err := n.ExportBand(nil); err != nil {
		t.Errorf("ExportBand: %v", err)
	}
	if err := n.ExportPageEnd(nil); err != nil {
		t.Errorf("ExportPageEnd: %v", err)
	}
	if err := n.Finish(); err != nil {
		t.Errorf("Finish: %v", err)
	}
}

// ── ExportsOptions ────────────────────────────────────────────────────────────

func TestExportsOptions_IsAllowed_EmptyList(t *testing.T) {
	o := export.NewExportsOptions()
	if !o.IsAllowed(export.ExportFormatPDF) {
		t.Error("empty AllowedExports should allow all formats")
	}
}

func TestExportsOptions_IsAllowed_WithList(t *testing.T) {
	o := export.NewExportsOptions()
	o.AllowedExports = []export.ExportFormat{export.ExportFormatHTML}
	if !o.IsAllowed(export.ExportFormatHTML) {
		t.Error("HTML should be allowed")
	}
	if o.IsAllowed(export.ExportFormatPDF) {
		t.Error("PDF should not be allowed")
	}
}

func TestExportsOptions_IsHidden(t *testing.T) {
	o := export.NewExportsOptions()
	o.HideExports = []export.ExportFormat{export.ExportFormatImage}
	if !o.IsHidden(export.ExportFormatImage) {
		t.Error("Image should be hidden")
	}
	if o.IsHidden(export.ExportFormatPDF) {
		t.Error("PDF should not be hidden")
	}
}

// mockWriter implements report.Writer for testing Serialize.
type mockWriter struct {
	strs  map[string]string
	bools map[string]bool
}

func newMockWriter() *mockWriter {
	return &mockWriter{strs: make(map[string]string), bools: make(map[string]bool)}
}

func (w *mockWriter) WriteStr(name, value string)              { w.strs[name] = value }
func (w *mockWriter) WriteInt(name string, value int)           {}
func (w *mockWriter) WriteBool(name string, value bool)         { w.bools[name] = value }
func (w *mockWriter) WriteFloat(name string, value float32)     {}
func (w *mockWriter) WriteObject(obj report.Serializable) error { return nil }
func (w *mockWriter) WriteObjectNamed(_ string, _ report.Serializable) error { return nil }

// mockReader implements report.Reader for testing Deserialize.
type mockReader struct {
	strs  map[string]string
	bools map[string]bool
}

func newMockReader(strs map[string]string, bools map[string]bool) *mockReader {
	return &mockReader{strs: strs, bools: bools}
}

func (r *mockReader) ReadStr(name, def string) string {
	if v, ok := r.strs[name]; ok {
		return v
	}
	return def
}
func (r *mockReader) ReadInt(name string, def int) int       { return def }
func (r *mockReader) ReadBool(name string, def bool) bool {
	if v, ok := r.bools[name]; ok {
		return v
	}
	return def
}
func (r *mockReader) ReadFloat(name string, def float32) float32 { return def }
func (r *mockReader) NextChild() (string, bool)                  { return "", false }
func (r *mockReader) FinishChild() error                         { return nil }

func TestExportsOptions_Serialize(t *testing.T) {
	o := export.NewExportsOptions()
	o.DefaultFormat = export.ExportFormatHTML // not PDF → should be written
	o.ShowProgress = false                    // not default → should be written
	o.OpenAfterExport = true                  // not default → should be written

	w := newMockWriter()
	o.Serialize(w)

	if w.strs["ExportsOptions.DefaultFormat"] != "HTML" {
		t.Errorf("DefaultFormat not serialized, got %q", w.strs["ExportsOptions.DefaultFormat"])
	}
	if w.bools["ExportsOptions.ShowProgress"] != false {
		t.Error("ShowProgress not serialized")
	}
	if w.bools["ExportsOptions.OpenAfterExport"] != true {
		t.Error("OpenAfterExport not serialized")
	}
}

func TestExportsOptions_Deserialize(t *testing.T) {
	o := export.NewExportsOptions()
	r := newMockReader(
		map[string]string{"ExportsOptions.DefaultFormat": "Image"},
		map[string]bool{"ExportsOptions.ShowProgress": false, "ExportsOptions.OpenAfterExport": true},
	)
	o.Deserialize(r)

	if o.DefaultFormat != export.ExportFormatImage {
		t.Errorf("DefaultFormat = %q, want Image", o.DefaultFormat)
	}
	if o.ShowProgress {
		t.Error("ShowProgress should be false")
	}
	if !o.OpenAfterExport {
		t.Error("OpenAfterExport should be true")
	}
}

// ── Utils — missing branches ──────────────────────────────────────────────────

func TestPointsToPixels(t *testing.T) {
	px := export.PointsToPixels(72)
	if px < 95.9 || px > 96.1 {
		t.Errorf("PointsToPixels(72) = %v, want ~96", px)
	}
}

func TestPixelsToInches(t *testing.T) {
	in := export.PixelsToInches(96)
	if in < 0.99 || in > 1.01 {
		t.Errorf("PixelsToInches(96) = %v, want ~1.0", in)
	}
}

func TestInchesToPixels(t *testing.T) {
	px := export.InchesToPixels(1)
	if px < 95.9 || px > 96.1 {
		t.Errorf("InchesToPixels(1) = %v, want ~96", px)
	}
}

func TestHTMLString_NBSP(t *testing.T) {
	// \u00a0 non-breaking space → "&nbsp;"
	got := export.HTMLString("a\u00a0b")
	if !strings.Contains(got, "&nbsp;") {
		t.Errorf("HTMLString with NBSP: got %q, want &nbsp; entity", got)
	}
}

func TestXMLString_CR_Tab(t *testing.T) {
	got := export.XMLString("a\rb\tc")
	if !strings.Contains(got, "&#xD;") {
		t.Errorf("XMLString CR not escaped: %q", got)
	}
	if !strings.Contains(got, "&#x9;") {
		t.Errorf("XMLString TAB not escaped: %q", got)
	}
}

func TestHTMLColorToRGB_ShortForm(t *testing.T) {
	// 3-char short form "#FFF" → expands to "FFFFFF" → white
	r, g, b, ok := export.HTMLColorToRGB("#FFF")
	if !ok {
		t.Fatal("expected ok=true for #FFF")
	}
	if r != 255 || g != 255 || b != 255 {
		t.Errorf("got %d %d %d, want 255 255 255", r, g, b)
	}
}

func TestHTMLColorToRGB_ShortForm_Invalid(t *testing.T) {
	// 3-char with invalid hex → err in Sscanf
	_, _, _, ok := export.HTMLColorToRGB("#GGG")
	if ok {
		t.Error("invalid 3-char color should return ok=false")
	}
}

func TestHTMLColorToRGB_WrongLength(t *testing.T) {
	// 5-char → not case-3 or case-6 → default return (0,0,0,false)
	_, _, _, ok := export.HTMLColorToRGB("#12345")
	if ok {
		t.Error("5-char color should return ok=false")
	}
}
