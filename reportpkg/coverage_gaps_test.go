package reportpkg

// coverage_gaps_test.go — additional internal tests to reach 100% coverage.
// All of these use package reportpkg (not _test) so we can access unexported
// functions and types.

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/andrewloable/go-fastreport/band"
	"github.com/andrewloable/go-fastreport/data"
	preview "github.com/andrewloable/go-fastreport/preview"
	"github.com/andrewloable/go-fastreport/report"
	"github.com/andrewloable/go-fastreport/serial"
)

// ── Serial registry — uncovered factory functions ─────────────────────────

// TestSerialRegistrations_UncoveredFactories exercises factory functions in
// serial_registrations.go whose closures were not invoked by earlier tests.
func TestSerialRegistrations_UncoveredFactories(t *testing.T) {
	names := []string{
		// Report-level containers (serial_registrations.go lines 19-20)
		"Report",
		"ReportPage",
		// Object types not yet exercised by existing tests
		"ContainerObject",
		"HtmlObject",
		"SparklineObject",
		"AdvMatrixObject",
		"MSChartSeries",
		"RFIDLabel",
		"MapLayer",
	}
	for _, name := range names {
		obj, err := serial.DefaultRegistry.Create(name)
		if err != nil {
			t.Errorf("Create(%q) returned error: %v", name, err)
			continue
		}
		if obj == nil {
			t.Errorf("Create(%q) returned nil", name)
		}
	}
}

// ── Prepare — EvaluateAll error path (prepare.go:52-54) ──────────────────

func TestPrepare_EvaluateAll_Error(t *testing.T) {
	savedFunc := globalPrepareFunc
	globalPrepareFunc = func(r *Report) (*preview.PreparedPages, error) {
		return preview.New(), nil
	}
	defer func() { globalPrepareFunc = savedFunc }()

	r := NewReport()
	// A parameter with an invalid expression causes EvaluateAll to fail.
	r.Dictionary().AddParameter(&data.Parameter{
		Name:       "BadParam",
		Expression: "!!invalid_expr!!",
	})
	err := r.Prepare()
	if err == nil {
		t.Error("expected error from EvaluateAll with invalid expression")
	}
	if !strings.Contains(err.Error(), "parameter evaluation") {
		t.Errorf("unexpected error message: %v", err)
	}
}

// ── PrepareWithContext — EvaluateAll error path (prepare.go:75-77) ────────

func TestPrepareWithContext_EvaluateAll_Error(t *testing.T) {
	savedCtx := globalPrepareFuncCtx
	globalPrepareFuncCtx = func(ctx context.Context, r *Report) (*preview.PreparedPages, error) {
		return preview.New(), nil
	}
	defer func() { globalPrepareFuncCtx = savedCtx }()

	r := NewReport()
	r.Dictionary().AddParameter(&data.Parameter{
		Name:       "BadParam",
		Expression: "!!invalid!!",
	})
	err := r.PrepareWithContext(context.Background())
	if err == nil {
		t.Error("expected error from EvaluateAll in PrepareWithContext")
	}
	if !strings.Contains(err.Error(), "parameter evaluation") {
		t.Errorf("unexpected error message: %v", err)
	}
}

// ── errWriter — io.Writer that fails after N bytes ────────────────────────

type errWriter struct {
	failAfter int
	written   int
}

func (e *errWriter) Write(p []byte) (int, error) {
	if e.written >= e.failAfter {
		return 0, fmt.Errorf("errWriter: write error")
	}
	n := len(p)
	if e.written+n > e.failAfter {
		n = e.failAfter - e.written
	}
	e.written += n
	return n, nil
}

// ── countingFailWriter — report.Writer that succeeds on first N WriteObject
// calls then fails on subsequent ones.

type countingFailWriter struct {
	succeedCount int
	called       int
}

func (c *countingFailWriter) WriteStr(name, value string)          {}
func (c *countingFailWriter) WriteInt(name string, value int)       {}
func (c *countingFailWriter) WriteBool(name string, value bool)     {}
func (c *countingFailWriter) WriteFloat(name string, value float32) {}
func (c *countingFailWriter) WriteObject(obj report.Serializable) error {
	c.called++
	if c.called > c.succeedCount {
		return fmt.Errorf("countingFailWriter: WriteObject error on call %d", c.called)
	}
	return nil
}
func (c *countingFailWriter) WriteObjectNamed(name string, obj report.Serializable) error {
	c.called++
	if c.called > c.succeedCount {
		return fmt.Errorf("countingFailWriter: WriteObjectNamed error on call %d", c.called)
	}
	return nil
}

// ── noopWriter — report.Writer that never fails ───────────────────────────

type noopWriter struct{}

func (n *noopWriter) WriteStr(name, value string)          {}
func (n *noopWriter) WriteInt(name string, value int)       {}
func (n *noopWriter) WriteBool(name string, value bool)     {}
func (n *noopWriter) WriteFloat(name string, value float32) {}
func (n *noopWriter) WriteObject(obj report.Serializable) error {
	return nil
}
func (n *noopWriter) WriteObjectNamed(name string, obj report.Serializable) error {
	return nil
}

// ── Report.Serialize — Styles write error (report.go:196-198) ─────────────

func TestReport_Serialize_StylesWriteError(t *testing.T) {
	// Load a report with a style so that the styles serializer is invoked.
	r := NewReport()
	frx := `<?xml version="1.0" encoding="utf-8"?><Report>
		<Styles>
			<Style Name="S1"/>
		</Styles>
		<ReportPage Name="Page1"/>
	</Report>`
	if err := r.LoadFromString(frx); err != nil {
		t.Fatalf("LoadFromString: %v", err)
	}
	// mockPageWriter.failWriteObject=true causes WriteObject to fail, which is
	// called when writing the Styles child element.
	w := &mockPageWriter{failWriteObject: true}
	err := r.Serialize(w)
	if err == nil {
		t.Error("expected error when writing Styles fails")
	}
}

// ── Report.Serialize — Pages write error (report.go:203-205) ─────────────

func TestReport_Serialize_PagesWriteError(t *testing.T) {
	r := NewReport()
	pg := NewReportPage()
	pg.SetName("Page1")
	r.AddPage(pg)

	w := &mockPageWriter{failWriteObject: true}
	err := r.Serialize(w)
	if err == nil {
		t.Error("expected error when writing pages fails")
	}
}

// ── Report.Deserialize — extended attribute coverage (report.go:212-214) ──

func TestReport_Deserialize_AllAttributes(t *testing.T) {
	frx := `<?xml version="1.0" encoding="utf-8"?><Report ReportName="R1" ReportAuthor="Alice" ReportDescription="Desc" ReportVersion="2.0" Created="2024-01-01" Modified="2024-01-02" CreatorVersion="2023.1" SavePreviewPicture="false" ConvertNulls="true" DoublePass="true" InitialPageNumber="3" MaxPages="10" StartReportEvent="start" FinishReportEvent="finish">
		<ReportPage Name="Page1"/>
	</Report>`
	r := NewReport()
	if err := r.LoadFromString(frx); err != nil {
		t.Fatalf("LoadFromString: %v", err)
	}
	if r.Info.Name != "R1" {
		t.Errorf("Info.Name: %q", r.Info.Name)
	}
	if r.InitialPageNumber != 3 {
		t.Errorf("InitialPageNumber: %d", r.InitialPageNumber)
	}
	if !r.ConvertNulls {
		t.Error("ConvertNulls should be true")
	}
}

// ── Page.Serialize — BaseObject.Serialize path (page.go:306-308) ──────────

// BaseObject.Serialize cannot return an error (WriteStr is void), so the
// error-check at page.go:306 is exercised simply by calling Serialize with
// a writer that propagates errors from WriteObject.
func TestPage_Serialize_NoError(t *testing.T) {
	pg := NewReportPage()
	pg.SetName("P1")
	w := &noopWriter{}
	err := pg.Serialize(w)
	if err != nil {
		t.Errorf("Page.Serialize unexpectedly failed: %v", err)
	}
}

// ── Page.Serialize — serializeBands error (page.go:374-376) ───────────────

func TestPage_Serialize_SerializeBandsError(t *testing.T) {
	pg := NewReportPage()
	ph := band.NewPageHeaderBand()
	ph.SetName("PH1")
	pg.SetPageHeader(ph)

	w := &mockPageWriter{failWriteObject: true}
	err := pg.Serialize(w)
	if err == nil {
		t.Error("Page.Serialize should propagate serializeBands error")
	}
}

// ── serializeBands — TitleBeforeHeader+PageHeader error (page.go:402-404) ─

func TestSerializeBands_TitleBeforeHeader_PageHeaderError(t *testing.T) {
	pg := NewReportPage()
	pg.TitleBeforeHeader = true
	// Only set PageHeader (no ReportTitle) so the ReportTitle branch is skipped
	// and the first write attempt is PageHeader — which fails.
	ph := band.NewPageHeaderBand()
	ph.SetName("PH1")
	pg.SetPageHeader(ph)

	w := &mockPageWriter{failWriteObject: true}
	err := pg.serializeBands(w)
	if err == nil {
		t.Error("expected error when WriteObject fails for PageHeader in TitleBeforeHeader path")
	}
}

// ── serializeBands — nil band in dynamic bands (page.go:415-416) ─────────

func TestSerializeBands_NilDynamicBand(t *testing.T) {
	pg := NewReportPage()
	// Inject a nil band directly (not possible via public API).
	pg.bands = append(pg.bands, nil)
	db := band.NewDataBand()
	db.SetName("DB1")
	pg.bands = append(pg.bands, db)

	r := NewReport()
	r.AddPage(pg)
	xml, err := r.SaveToString()
	if err != nil {
		t.Fatalf("SaveToString with nil band: %v", err)
	}
	if !strings.Contains(xml, "<Data") {
		t.Error("expected Data band in XML output")
	}
}

// ── serializeBands — default order ReportTitle write error (page.go:402-404) ─
// TitleBeforeHeader=false (default). PageHeader succeeds, but ReportTitle fails.

func TestSerializeBands_DefaultOrder_ReportTitle_WriteObjectError(t *testing.T) {
	pg := NewReportPage()
	// TitleBeforeHeader defaults to false.
	ph := band.NewPageHeaderBand()
	ph.SetName("PH1")
	pg.SetPageHeader(ph)
	rt := band.NewReportTitleBand()
	rt.SetName("RT1")
	pg.SetReportTitle(rt)

	// Succeed on the first WriteObject (PageHeader), fail on second (ReportTitle).
	w := &countingFailWriter{succeedCount: 1}
	err := pg.serializeBands(w)
	if err == nil {
		t.Error("expected error when WriteObject fails for ReportTitle in default order")
	}
}

// ── serializeBands — ColumnFooter write error (page.go:435-437) ──────────

func TestSerializeBands_ColumnFooter_WriteObjectError(t *testing.T) {
	pg := NewReportPage()
	cf := band.NewColumnFooterBand()
	cf.SetName("CF1")
	pg.SetColumnFooter(cf)

	w := &mockPageWriter{failWriteObject: true}
	err := pg.serializeBands(w)
	if err == nil {
		t.Error("expected error when WriteObject fails for ColumnFooter")
	}
}

// ── serializeBands — PageFooter write error (page.go:440-442) ─────────────

func TestSerializeBands_PageFooter_WriteObjectError(t *testing.T) {
	pg := NewReportPage()
	pf := band.NewPageFooterBand()
	pf.SetName("PF1")
	pg.SetPageFooter(pf)

	w := &mockPageWriter{failWriteObject: true}
	err := pg.serializeBands(w)
	if err == nil {
		t.Error("expected error when WriteObject fails for PageFooter")
	}
}

// ── serializeBands — Overlay write error ─────────────────────────────────

func TestSerializeBands_Overlay_WriteObjectError(t *testing.T) {
	pg := NewReportPage()
	ov := band.NewOverlayBand()
	ov.SetName("OV1")
	pg.SetOverlay(ov)

	w := &mockPageWriter{failWriteObject: true}
	err := pg.serializeBands(w)
	if err == nil {
		t.Error("expected error when WriteObject fails for Overlay")
	}
}

// ── Page.Deserialize — full attribute coverage (page.go:448+) ─────────────

func TestPage_Deserialize_AllAttributes(t *testing.T) {
	frx := `<?xml version="1.0" encoding="utf-8"?><Report>
		<ReportPage Name="TestPg" PaperWidth="100" PaperHeight="200"
			Landscape="true" LeftMargin="5" TopMargin="6" RightMargin="7" BottomMargin="8"
			MirrorMargins="true" TitleBeforeHeader="true" PrintOnPreviousPage="true"
			ResetPageNumber="true" StartOnOddPage="true"
			OutlineExpression="[Page]" CreatePageEvent="pg_create"
			StartPageEvent="pg_start" FinishPageEvent="pg_finish"
			ManualBuildEvent="pg_build" BackPage="OtherPage" BackPageOddEven="1"
			Columns.Count="2" Columns.Width="95"/>
	</Report>`
	r := NewReport()
	if err := r.LoadFromString(frx); err != nil {
		t.Fatalf("LoadFromString: %v", err)
	}
	if r.PageCount() == 0 {
		t.Fatal("no pages")
	}
	pg := r.Page(0)
	if pg.PaperWidth != 100 {
		t.Errorf("PaperWidth: %v", pg.PaperWidth)
	}
	if !pg.Landscape {
		t.Error("expected Landscape=true")
	}
	if pg.LeftMargin != 5 {
		t.Errorf("LeftMargin: %v", pg.LeftMargin)
	}
	if pg.Columns.Count != 2 {
		t.Errorf("Columns.Count: %d", pg.Columns.Count)
	}
	if pg.BackPage != "OtherPage" {
		t.Errorf("BackPage: %q", pg.BackPage)
	}
}

// ── Page.Deserialize — watermark nil path (page.go:473-475) ─────────────

func TestPage_Deserialize_WatermarkNilGuard(t *testing.T) {
	// Create a ReportPage and manually set Watermark to nil to exercise
	// the nil-guard in Deserialize (page.go:473-475).
	pg := NewReportPage()
	pg.Watermark = nil // force nil

	frx := `<?xml version="1.0" encoding="utf-8"?><Report>
		<ReportPage Name="WMNil" Watermark.Enabled="true" Watermark.Text="TEST"/>
	</Report>`
	r := NewReport()
	// Replace the page's Watermark after loading so we control the nil state.
	// Actually we need to exercise the nil guard during Deserialize itself.
	// We do this by calling pg.Deserialize directly with a reader that has
	// Watermark attributes, after setting pg.Watermark = nil.
	if err := r.LoadFromString(frx); err != nil {
		t.Fatalf("LoadFromString: %v", err)
	}
	// The test above already exercises the normal path. To hit the nil guard,
	// we need Watermark to be nil at the time Deserialize is called.
	// Manually test via serial.NewReader:
	xmlDoc := `<ReportPage Name="WMNil2" Watermark.Enabled="true"/>`
	rdr := serial.NewReader(strings.NewReader(xmlDoc))
	typeName, ok := rdr.ReadObjectHeader()
	if !ok || typeName != "ReportPage" {
		t.Fatalf("ReadObjectHeader: got %q, ok=%v", typeName, ok)
	}
	pg2 := NewReportPage()
	pg2.Watermark = nil // force nil before Deserialize
	if err := pg2.Deserialize(rdr); err != nil {
		t.Fatalf("Deserialize: %v", err)
	}
	if pg2.Watermark == nil {
		t.Error("Watermark should be initialized by Deserialize when nil")
	}
	if !pg2.Watermark.Enabled {
		t.Error("Watermark.Enabled should be true")
	}
}

// ── LoadFromWithPassword — NewReaderWithPassword error (loadsave.go:61-63) ─

func TestLoadFromWithPassword_ReadError(t *testing.T) {
	// Use an io.Pipe where the writer end is closed with an error immediately.
	// This causes PeekAndDecrypt (called by NewReaderWithPassword) to receive
	// a non-EOF, non-ErrUnexpectedEOF error, triggering the error return.
	pr, pw := newErrPipe(fmt.Errorf("simulated read error"))
	_ = pw
	r := NewReport()
	err := r.LoadFromWithPassword(pr, "password")
	if err == nil {
		t.Error("LoadFromWithPassword should return error when reader fails")
	}
}

// newErrPipe creates an io.PipeReader that immediately returns the given error.
func newErrPipe(readErr error) (*errPipeReader, struct{}) {
	return &errPipeReader{err: readErr}, struct{}{}
}

type errPipeReader struct {
	err error
}

func (e *errPipeReader) Read(p []byte) (int, error) {
	return 0, e.err
}

// ── loadFromSerialReader — extra attributes (loadsave.go:107+) ───────────

func TestLoadFromSerialReader_AllReportAttributes(t *testing.T) {
	// Exercise all Deserialize attribute paths and the unknown-child default path.
	r := NewReport()
	frx := `<?xml version="1.0" encoding="utf-8"?><Report ReportName="Test" ReportAuthor="Alice" ReportDescription="Desc" ReportVersion="1.0" Created="2024-01-01" Modified="2024-01-02" CreatorVersion="2023.1" SavePreviewPicture="false" ConvertNulls="true" DoublePass="true" InitialPageNumber="2" MaxPages="10" StartReportEvent="start" FinishReportEvent="finish">
		<UnknownTopLevel Name="Skip"/>
		<ReportPage Name="Page1"/>
	</Report>`
	if err := r.LoadFromString(frx); err != nil {
		t.Fatalf("LoadFromString: %v", err)
	}
	if r.Info.Author != "Alice" {
		t.Errorf("Author: %q", r.Info.Author)
	}
	if r.InitialPageNumber != 2 {
		t.Errorf("InitialPageNumber: %d", r.InitialPageNumber)
	}
}

// ── deserializeJsonConnection — non-TableDataSource child (loadsave.go:281-283) ─

func TestDeserializeJsonConnection_NonTableDataSourceChild(t *testing.T) {
	frx := `<?xml version="1.0" encoding="utf-8"?><Report>
		<Dictionary>
			<JsonDataConnection Name="JC" ConnectionString="/tmp/data.json">
				<OtherChild Name="X"/>
				<TableDataSource Name="T1" TableName="$.items"/>
			</JsonDataConnection>
		</Dictionary>
		<ReportPage Name="Page1"/>
	</Report>`
	r := NewReport()
	if err := r.LoadFromString(frx); err != nil {
		t.Fatalf("LoadFromString: %v", err)
	}
	sources := r.Dictionary().DataSources()
	if len(sources) == 0 {
		t.Error("expected at least 1 data source from JsonDataConnection")
	}
}

// ── deserializeJsonTableDataSource — drain nested children (loadsave.go:301) ─

func TestDeserializeJsonTableDataSource_WithChildren(t *testing.T) {
	frx := `<?xml version="1.0" encoding="utf-8"?><Report>
		<Dictionary>
			<JsonDataConnection Name="JC" ConnectionString="/tmp/data.json">
				<TableDataSource Name="T1" TableName="$.items">
					<Column Name="id"/>
					<Column Name="name"/>
				</TableDataSource>
			</JsonDataConnection>
		</Dictionary>
		<ReportPage Name="Page1"/>
	</Report>`
	r := NewReport()
	if err := r.LoadFromString(frx); err != nil {
		t.Fatalf("LoadFromString: %v", err)
	}
	sources := r.Dictionary().DataSources()
	if len(sources) == 0 {
		t.Error("expected at least 1 data source")
	}
}

// ── deserializeCsvConnection — non-TableDataSource child (loadsave.go:351-353) ─

func TestDeserializeCsvConnection_NonTableDataSourceChild(t *testing.T) {
	frx := `<?xml version="1.0" encoding="utf-8"?><Report>
		<Dictionary>
			<CsvDataConnection Name="CC" ConnectionString="/tmp/data.csv">
				<OtherChild Name="X"/>
				<TableDataSource Name="CT1" Separator="," HasHeader="true"/>
			</CsvDataConnection>
		</Dictionary>
		<ReportPage Name="Page1"/>
	</Report>`
	r := NewReport()
	if err := r.LoadFromString(frx); err != nil {
		t.Fatalf("LoadFromString: %v", err)
	}
	sources := r.Dictionary().DataSources()
	if len(sources) == 0 {
		t.Error("expected at least 1 CSV data source")
	}
}

// ── deserializeXmlConnection — drain children inside TableDataSource (loadsave.go:388) ─

func TestDeserializeXmlConnection_TableDataSourceWithChildren(t *testing.T) {
	frx := `<?xml version="1.0" encoding="utf-8"?><Report>
		<Dictionary>
			<XmlDataConnection Name="XC" ConnectionString="/tmp/data.xml">
				<TableDataSource Name="XT1" TableName="/Root/Item">
					<Column Name="id"/>
				</TableDataSource>
			</XmlDataConnection>
		</Dictionary>
		<ReportPage Name="Page1"/>
	</Report>`
	r := NewReport()
	if err := r.LoadFromString(frx); err != nil {
		t.Fatalf("LoadFromString: %v", err)
	}
	sources := r.Dictionary().DataSources()
	if len(sources) == 0 {
		t.Error("expected at least 1 XML data source")
	}
}

// ── deserializeXmlConnection — non-TableDataSource child (loadsave.go:395-397) ─

func TestDeserializeXmlConnection_NonTableDataSourceChild(t *testing.T) {
	frx := `<?xml version="1.0" encoding="utf-8"?><Report>
		<Dictionary>
			<XmlDataConnection Name="XC" ConnectionString="/tmp/data.xml">
				<OtherChild Name="X"/>
				<TableDataSource Name="XT1" TableName="/Root/Item"/>
			</XmlDataConnection>
		</Dictionary>
		<ReportPage Name="Page1"/>
	</Report>`
	r := NewReport()
	if err := r.LoadFromString(frx); err != nil {
		t.Fatalf("LoadFromString: %v", err)
	}
	sources := r.Dictionary().DataSources()
	if len(sources) == 0 {
		t.Error("expected at least 1 XML data source after skipping unknown child")
	}
}

// ── deserializeRelation — drain children (loadsave.go:465) ────────────────

func TestDeserializeRelation_WithChildren(t *testing.T) {
	frx := `<?xml version="1.0" encoding="utf-8"?><Report>
		<Dictionary>
			<Relation Name="R1" ParentDataSource="A" ChildDataSource="B"
				ParentColumns="ID" ChildColumns="FK">
				<SomeChild Name="x"/>
			</Relation>
		</Dictionary>
		<ReportPage Name="Page1"/>
	</Report>`
	r := NewReport()
	if err := r.LoadFromString(frx); err != nil {
		t.Fatalf("LoadFromString: %v", err)
	}
	rels := r.Dictionary().Relations()
	if len(rels) != 1 {
		t.Fatalf("expected 1 relation, got %d", len(rels))
	}
}

// ── deserializeTotal — drain children (loadsave.go:485) ──────────────────

func TestDeserializeTotal_WithChildren(t *testing.T) {
	frx := `<?xml version="1.0" encoding="utf-8"?><Report>
		<Dictionary>
			<Total Name="T1" Expression="[Val]" TotalType="Sum">
				<SomeChild Name="x"/>
			</Total>
		</Dictionary>
		<ReportPage Name="Page1"/>
	</Report>`
	r := NewReport()
	if err := r.LoadFromString(frx); err != nil {
		t.Fatalf("LoadFromString: %v", err)
	}
	totals := r.Dictionary().Totals()
	if len(totals) != 1 {
		t.Fatalf("expected 1 total, got %d", len(totals))
	}
}

// ── deserializeStyleEntry — drain children (loadsave.go:593) ─────────────

func TestDeserializeStyleEntry_WithChildren(t *testing.T) {
	frx := `<?xml version="1.0" encoding="utf-8"?><Report>
		<Styles>
			<Style Name="WithKids">
				<SomeChild Name="x"/>
			</Style>
		</Styles>
		<ReportPage Name="Page1"/>
	</Report>`
	r := NewReport()
	if err := r.LoadFromString(frx); err != nil {
		t.Fatalf("LoadFromString: %v", err)
	}
	e := r.Styles().Find("WithKids")
	if e == nil {
		t.Fatal("WithKids style not found")
	}
}

// ── SaveTo — WriteHeader error (loadsave.go:670-672) ────────────────────

func TestSaveTo_WriteHeaderError(t *testing.T) {
	r := NewReport()
	r.Info.Name = "TestReport"
	// Fail immediately (0 bytes allowed) so WriteHeader fails.
	ew := &errWriter{failAfter: 0}
	err := r.SaveTo(ew)
	if err == nil {
		t.Error("SaveTo should return error when WriteHeader fails")
	}
}

// ── SaveTo — WriteObjectNamed error (loadsave.go:673-675) ────────────────
// We need WriteHeader to succeed but WriteObjectNamed to fail.
// The header is 38 bytes. We allow those, then fail on the next write.
// Note: xml.Encoder may buffer internally. We use a callCount writer to fail
// on the second Write call (after the first Write for the header).

type callCountWriter struct {
	calls     int
	failAfter int // fail starting from this call number (1-indexed)
}

func (c *callCountWriter) Write(p []byte) (int, error) {
	c.calls++
	if c.calls >= c.failAfter {
		return 0, fmt.Errorf("callCountWriter: Write failed on call %d", c.calls)
	}
	return len(p), nil
}

func TestSaveTo_WriteObjectNamedError(t *testing.T) {
	// Build a Report large enough that the xml.Encoder's internal buffer
	// overflows and triggers a write to the underlying io.Writer mid-encoding.
	// This means the writer failure propagates through WriteObjectNamed at
	// loadsave.go:673-675 rather than only at Flush().
	r := NewReport()
	r.Info.Name = "TestReport"
	// Add many pages so the encoded XML exceeds the internal bufio.Writer
	// buffer (default 4096 or 8192 bytes), forcing a mid-stream write.
	for i := 0; i < 300; i++ {
		pg := NewReportPage()
		pg.SetName(fmt.Sprintf("Page%04d", i))
		r.AddPage(pg)
	}
	// Allow the header (38 bytes) through, fail on subsequent writes.
	ew := &errWriter{failAfter: 38}
	err := r.SaveTo(ew)
	if err == nil {
		t.Error("SaveTo should return error when underlying writer fails after header")
	}
}

// ── SaveTo — compressed write error ──────────────────────────────────────

func TestSaveTo_Compressed_WriteError(t *testing.T) {
	r := NewReport()
	r.Compressed = true
	// Fail immediately so the gzip writer's first flush fails.
	ew := &errWriter{failAfter: 0}
	err := r.SaveTo(ew)
	if err == nil {
		t.Error("SaveTo compressed should return error when writer fails")
	}
}

// ── FindPage — not found / found paths ───────────────────────────────────

func TestFindPage_Coverage(t *testing.T) {
	r := NewReport()
	pg := NewReportPage()
	pg.SetName("TargetPage")
	r.AddPage(pg)

	if result := r.FindPage("NonExistent"); result != nil {
		t.Error("FindPage should return nil for non-existent page")
	}
	if result := r.FindPage("TargetPage"); result == nil {
		t.Fatal("FindPage should return the page")
	}
}
