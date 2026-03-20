package reportpkg

// inherited_loadsave_test.go — tests for inherited report loading coverage.
//
// Targets:
//   loadsave.go:29   Load                  66.7% → cover gzip path, empty-file path
//   loadsave.go:190  loadInherited         60.0% → cover missing BaseReport, no-dir context,
//                                                   base-load error, DialogPage, ReportPage,
//                                                   Styles, unknown child in inherited
//   loadsave.go:261  deserializeInheritedPage 75% → cover obj.Deserialize error path
//   dialogpage.go:17 DialogPage.TypeName    0.0% → direct call
//   dialogpage.go:60 DialogPage.Serialize   0.0% → direct call
//   dialogpage.go:47 drainChildren         71.4% → cover nested recursive drain

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/andrewloable/go-fastreport/report"
	"github.com/andrewloable/go-fastreport/serial"
)

// ── helpers ──────────────────────────────────────────────────────────────────

// testReportsDirInternal returns the absolute path to test-reports/.
func testReportsDirInternal() string {
	_, thisFile, _, _ := runtime.Caller(0)
	return filepath.Join(filepath.Dir(thisFile), "..", "test-reports")
}

// ── Load: non-existent file ───────────────────────────────────────────────────

func TestLoad_NonExistentFile(t *testing.T) {
	r := NewReport()
	err := r.Load("/no/such/file/totally_missing_12345.frx")
	if err == nil {
		t.Fatal("Load with non-existent file should return an error")
	}
	if !strings.Contains(err.Error(), "report.Load") {
		t.Errorf("error should mention 'report.Load', got: %v", err)
	}
}

// ── Load: empty file (io.ReadFull error) ─────────────────────────────────────

func TestLoad_EmptyFile(t *testing.T) {
	// Create a temp empty file; io.ReadFull will return io.ErrUnexpectedEOF.
	tmp, err := os.CreateTemp("", "empty_*.frx")
	if err != nil {
		t.Fatalf("create temp file: %v", err)
	}
	tmp.Close()
	defer os.Remove(tmp.Name())

	r := NewReport()
	loadErr := r.Load(tmp.Name())
	if loadErr == nil {
		t.Fatal("Load of empty file should return an error")
	}
	if !strings.Contains(loadErr.Error(), "report.Load") {
		t.Errorf("error should mention 'report.Load', got: %v", loadErr)
	}
}

// ── Load: gzip-compressed FRX file ───────────────────────────────────────────

func TestLoad_GzipCompressedFile(t *testing.T) {
	// Build a minimal report, save it compressed, write to a temp file, then Load.
	r := NewReport()
	r.Info.Name = "GzipLoadTest"
	pg := NewReportPage()
	pg.SetName("Page1")
	r.AddPage(pg)
	r.Compressed = true

	var buf bytes.Buffer
	if err := r.SaveTo(&buf); err != nil {
		t.Fatalf("SaveTo compressed: %v", err)
	}
	// Verify it really is gzip.
	b := buf.Bytes()
	if len(b) < 2 || b[0] != 0x1f || b[1] != 0x8b {
		t.Fatalf("expected gzip magic, got: %x %x", b[0], b[1])
	}

	tmp, err := os.CreateTemp("", "gzip_*.frx")
	if err != nil {
		t.Fatalf("create temp file: %v", err)
	}
	if _, err2 := tmp.Write(b); err2 != nil {
		tmp.Close()
		os.Remove(tmp.Name())
		t.Fatalf("write temp file: %v", err2)
	}
	tmp.Close()
	defer os.Remove(tmp.Name())

	r2 := NewReport()
	if err3 := r2.Load(tmp.Name()); err3 != nil {
		t.Fatalf("Load compressed file: %v", err3)
	}
	if r2.Info.Name != "GzipLoadTest" {
		t.Errorf("Info.Name = %q, want GzipLoadTest", r2.Info.Name)
	}
}

// ── Load: file with bad gzip magic (valid XML after 2 non-gzip bytes) ─────────

func TestLoad_BadGzipMagic_FallsBackToXML(t *testing.T) {
	// Write a plain XML FRX; the first two bytes are '<' and '?' (not gzip magic).
	// This exercises the non-gzip path of Load and confirms the peek+MultiReader approach.
	r := NewReport()
	r.Info.Name = "PlainLoadTest"
	pg := NewReportPage()
	pg.SetName("P1")
	r.AddPage(pg)

	xmlStr, err := r.SaveToString()
	if err != nil {
		t.Fatalf("SaveToString: %v", err)
	}

	tmp, err := os.CreateTemp("", "plain_*.frx")
	if err != nil {
		t.Fatalf("create temp file: %v", err)
	}
	if _, err2 := tmp.WriteString(xmlStr); err2 != nil {
		tmp.Close()
		os.Remove(tmp.Name())
		t.Fatalf("write temp file: %v", err2)
	}
	tmp.Close()
	defer os.Remove(tmp.Name())

	r2 := NewReport()
	if err3 := r2.Load(tmp.Name()); err3 != nil {
		t.Fatalf("Load plain file: %v", err3)
	}
	if r2.Info.Name != "PlainLoadTest" {
		t.Errorf("Info.Name = %q, want PlainLoadTest", r2.Info.Name)
	}
}

// ── Load: corrupt gzip file (gzip.NewReader should fail) ─────────────────────

func TestLoad_CorruptGzipFile(t *testing.T) {
	// Write two gzip magic bytes followed by garbage — gzip.NewReader will fail.
	bad := []byte{0x1f, 0x8b, 0xde, 0xad, 0xbe, 0xef}

	tmp, err := os.CreateTemp("", "badgzip_*.frx")
	if err != nil {
		t.Fatalf("create temp file: %v", err)
	}
	if _, err2 := tmp.Write(bad); err2 != nil {
		tmp.Close()
		os.Remove(tmp.Name())
		t.Fatalf("write temp file: %v", err2)
	}
	tmp.Close()
	defer os.Remove(tmp.Name())

	r := NewReport()
	loadErr := r.Load(tmp.Name())
	if loadErr == nil {
		t.Fatal("Load with corrupt gzip file should return an error")
	}
	if !strings.Contains(loadErr.Error(), "report.Load") {
		t.Errorf("error should mention 'report.Load', got: %v", loadErr)
	}
}

// ── Load: real inherited report from test-reports/ ───────────────────────────

func TestLoad_InheritedReport_FromTestReports(t *testing.T) {
	dir := testReportsDirInternal()
	path := filepath.Join(dir, "Inherited Report.frx")
	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Skip("test-reports/Inherited Report.frx not present")
	}

	r := NewReport()
	if err := r.Load(path); err != nil {
		t.Fatalf("Load inherited report: %v", err)
	}
	// After ApplyBase, the report must have at least one page.
	if r.PageCount() == 0 {
		t.Error("expected at least 1 page after loading inherited report")
	}
}

// ── loadInherited: missing BaseReport attribute ───────────────────────────────

func TestLoadFromSerialReader_Inherited_MissingBaseReport(t *testing.T) {
	// <inherited> with no BaseReport attribute → error from loadInherited.
	frx := `<?xml version="1.0" encoding="utf-8"?><inherited>
		<ReportPage Name="Page1"/>
	</inherited>`

	r := NewReport()
	rdr := serial.NewReader(strings.NewReader(frx))
	err := r.loadFromSerialReader(rdr, "/some/dir")
	if err == nil {
		t.Fatal("loadInherited with missing BaseReport should return error")
	}
	if !strings.Contains(err.Error(), "BaseReport") && !strings.Contains(err.Error(), "inherited") {
		t.Errorf("unexpected error message: %v", err)
	}
}

// ── loadInherited: no directory context (baseDir == "") ──────────────────────

func TestLoadFromSerialReader_Inherited_NoDirContext(t *testing.T) {
	// <inherited BaseReport="something.frx"> with baseDir="" → cannot resolve path.
	frx := `<?xml version="1.0" encoding="utf-8"?><inherited BaseReport="something.frx">
	</inherited>`

	r := NewReport()
	rdr := serial.NewReader(strings.NewReader(frx))
	err := r.loadFromSerialReader(rdr, "") // empty baseDir
	if err == nil {
		t.Fatal("loadInherited with empty baseDir should return an error")
	}
	if !strings.Contains(err.Error(), "no directory context") &&
		!strings.Contains(err.Error(), "cannot resolve") {
		t.Errorf("unexpected error message: %v", err)
	}
}

// ── loadInherited: base report file not found ─────────────────────────────────

func TestLoadFromSerialReader_Inherited_BaseReportNotFound(t *testing.T) {
	frx := `<?xml version="1.0" encoding="utf-8"?><inherited BaseReport="nonexistent_base.frx">
	</inherited>`

	r := NewReport()
	rdr := serial.NewReader(strings.NewReader(frx))
	// baseDir must be non-empty so it tries to load the (non-existent) base file.
	err := r.loadFromSerialReader(rdr, "/tmp")
	if err == nil {
		t.Fatal("loadInherited with non-existent base report should return an error")
	}
	if !strings.Contains(err.Error(), "load base report") &&
		!strings.Contains(err.Error(), "report.LoadFrom") {
		t.Errorf("unexpected error message: %v", err)
	}
}

// ── loadInherited: via a temp base file — exercises ReportPage + Dictionary + Styles + unknown child ─

func TestLoadInherited_FullInheritance_TempFiles(t *testing.T) {
	// Create a temp directory to hold both the base and inherited reports.
	dir, err := os.MkdirTemp("", "inherit_test_*")
	if err != nil {
		t.Fatalf("MkdirTemp: %v", err)
	}
	defer os.RemoveAll(dir)

	// Write the base report.
	baseXML := `<?xml version="1.0" encoding="utf-8"?>
<Report ReportName="BaseReport">
  <Styles>
    <Style Name="S1"/>
  </Styles>
  <ReportPage Name="Page1">
    <PageHeaderBand Name="PageHeader1" Width="718.2" Height="18.9">
      <TextObject Name="T1" Width="100" Height="20" Text="Header"/>
    </PageHeaderBand>
    <PageFooterBand Name="PageFooter1" Top="200" Width="718.2" Height="18.9">
      <TextObject Name="T2" Width="100" Height="20" Text="Footer"/>
    </PageFooterBand>
  </ReportPage>
</Report>`
	baseFile := filepath.Join(dir, "base.frx")
	if err2 := os.WriteFile(baseFile, []byte(baseXML), 0644); err2 != nil {
		t.Fatalf("write base.frx: %v", err2)
	}

	// Write the inherited report, exercising:
	//   - <inherited Name="Page1"> page overlay
	//   - <ReportPage Name="NewPage"> new page
	//   - <DialogPage Name="DP1"> dialog page branch
	//   - <Styles> branch
	//   - <Dictionary> branch
	//   - unknown child (SkipElement branch)
	inheritedXML := `<?xml version="1.0" encoding="utf-8"?>
<inherited BaseReport="base.frx">
  <Styles>
    <Style Name="S2"/>
  </Styles>
  <Dictionary>
    <Parameter Name="P1" DataType="string"/>
  </Dictionary>
  <UnknownTopLevel Name="Ignored"/>
  <inherited Name="Page1">
    <DataBand Name="Data1" Width="718.2" Height="18.9" DataSource="">
      <TextObject Name="Text1" Width="100" Height="18.9" Text="[P1]"/>
    </DataBand>
  </inherited>
  <ReportPage Name="NewPage" PaperWidth="210" PaperHeight="297">
  </ReportPage>
  <DialogPage Name="DP1">
    <ButtonControl Name="Btn1"/>
  </DialogPage>
</inherited>`
	inheritedFile := filepath.Join(dir, "child.frx")
	if err3 := os.WriteFile(inheritedFile, []byte(inheritedXML), 0644); err3 != nil {
		t.Fatalf("write child.frx: %v", err3)
	}

	r := NewReport()
	if err4 := r.Load(inheritedFile); err4 != nil {
		t.Fatalf("Load inherited report: %v", err4)
	}

	// The merged report should have pages (at minimum the base page + new page).
	if r.PageCount() == 0 {
		t.Error("expected at least 1 page after loading inherited report")
	}
}

// ── deserializeInheritedPage: obj.Deserialize error path ─────────────────────

// failingInheritedBand is a registered type whose Deserialize always fails.
// Named uniquely to avoid conflicts with the failingBase type in loadsave_coverage2_test.go.
type failingInheritedBand struct {
	report.BaseObject
}

func (f *failingInheritedBand) TypeName() string { return "FailingInheritedBand" }
func (f *failingInheritedBand) Serialize(w report.Writer) error { return nil }
func (f *failingInheritedBand) Deserialize(r report.Reader) error {
	return fmt.Errorf("failingInheritedBand: intentional error")
}

func init() {
	_ = serial.DefaultRegistry.Register("FailingInheritedBand", func() report.Base {
		return &failingInheritedBand{}
	})
}

func TestDeserializeInheritedPage_BandDeserializeError(t *testing.T) {
	// Build the XML that loadInherited would see when processing an <inherited>
	// child page containing a band whose Deserialize fails.
	// We call deserializeInheritedPage directly with a reader positioned after
	// ReadObjectHeader of the <inherited Name="Page1"> element.
	xmlDoc := `<inherited Name="Page1">
		<FailingInheritedBand Name="Bad"/>
	</inherited>`

	rdr := serial.NewReader(strings.NewReader(xmlDoc))
	typeName, ok := rdr.ReadObjectHeader()
	if !ok || typeName != "inherited" {
		t.Fatalf("ReadObjectHeader: got %q, ok=%v", typeName, ok)
	}

	_, err := deserializeInheritedPage(rdr)
	if err == nil {
		t.Fatal("deserializeInheritedPage should return error when band Deserialize fails")
	}
	if !strings.Contains(err.Error(), "intentional error") &&
		!strings.Contains(err.Error(), "deserialize") {
		t.Errorf("unexpected error message: %v", err)
	}
}

// ── deserializeInheritedPage: unknown band type (SkipElement path) ────────────

func TestDeserializeInheritedPage_UnknownBandType(t *testing.T) {
	xmlDoc := `<inherited Name="Page1">
		<UnknownBandXYZ Name="Skip"/>
		<inherited Name="SomeBase"/>
	</inherited>`

	rdr := serial.NewReader(strings.NewReader(xmlDoc))
	typeName, ok := rdr.ReadObjectHeader()
	if !ok || typeName != "inherited" {
		t.Fatalf("ReadObjectHeader: got %q, ok=%v", typeName, ok)
	}

	pg, err := deserializeInheritedPage(rdr)
	if err != nil {
		t.Fatalf("deserializeInheritedPage should not error on unknown band: %v", err)
	}
	if pg == nil {
		t.Fatal("expected non-nil page")
	}
	if !pg.Inherited() {
		t.Error("page should be marked Inherited")
	}
}

// ── DialogPage.TypeName ───────────────────────────────────────────────────────

func TestDialogPage_TypeName(t *testing.T) {
	dp := NewDialogPage()
	if got := dp.TypeName(); got != "DialogPage" {
		t.Errorf("TypeName() = %q, want DialogPage", got)
	}
}

// ── DialogPage.Serialize ──────────────────────────────────────────────────────

func TestDialogPage_Serialize(t *testing.T) {
	dp := NewDialogPage()
	dp.SetName("DP1")

	// Serialize is a documented no-op; it should return nil regardless of writer.
	var buf bytes.Buffer
	w := serial.NewWriter(&buf)
	_ = w.WriteHeader()
	if err := dp.Serialize(w); err != nil {
		t.Errorf("DialogPage.Serialize should return nil, got: %v", err)
	}
}

// ── DialogPage.Serialize via noopWriter ───────────────────────────────────────

func TestDialogPage_Serialize_NoopWriter(t *testing.T) {
	dp := NewDialogPage()
	nw := &noopWriter{}
	if err := dp.Serialize(nw); err != nil {
		t.Errorf("DialogPage.Serialize with noopWriter: %v", err)
	}
}

// ── drainChildren: nested recursive drain ─────────────────────────────────────

func TestDrainChildren_NestedRecursive(t *testing.T) {
	// A DialogPage with nested child controls forces the recursive drainChildren call.
	// <ButtonControl> has a nested <SubControl> child, which drainChildren must recurse into.
	frx := `<?xml version="1.0" encoding="utf-8"?><Report>
		<DialogPage Name="DP2">
			<ButtonControl Name="B1">
				<SubControl Name="Inner1">
					<DeepNested Name="D1"/>
				</SubControl>
			</ButtonControl>
			<LabelControl Name="L1"/>
		</DialogPage>
		<ReportPage Name="Page1"/>
	</Report>`

	r := NewReport()
	if err := r.LoadFromString(frx); err != nil {
		t.Fatalf("LoadFromString with nested DialogPage children: %v", err)
	}
	// The report should load without error and have a page.
	if r.PageCount() == 0 {
		t.Error("expected a page in the report")
	}
}

// ── loadInherited: inherited page Deserialize error propagation ───────────────

func TestLoadInherited_InheritedPageDeserializeError(t *testing.T) {
	dir, err := os.MkdirTemp("", "inherit_err_test_*")
	if err != nil {
		t.Fatalf("MkdirTemp: %v", err)
	}
	defer os.RemoveAll(dir)

	baseXML := `<?xml version="1.0" encoding="utf-8"?>
<Report ReportName="BaseR">
  <ReportPage Name="Page1"/>
</Report>`
	if err2 := os.WriteFile(filepath.Join(dir, "base.frx"), []byte(baseXML), 0644); err2 != nil {
		t.Fatalf("write base.frx: %v", err2)
	}

	// The inherited report's page overlay contains a FailingInheritedBand.
	inheritedXML := `<?xml version="1.0" encoding="utf-8"?>
<inherited BaseReport="base.frx">
  <inherited Name="Page1">
    <FailingInheritedBand Name="Bad"/>
  </inherited>
</inherited>`
	childFile := filepath.Join(dir, "child.frx")
	if err3 := os.WriteFile(childFile, []byte(inheritedXML), 0644); err3 != nil {
		t.Fatalf("write child.frx: %v", err3)
	}

	r := NewReport()
	loadErr := r.Load(childFile)
	if loadErr == nil {
		t.Fatal("Load should return error when inherited page band Deserialize fails")
	}
	if !strings.Contains(loadErr.Error(), "deserialize inherited page") &&
		!strings.Contains(loadErr.Error(), "intentional error") {
		t.Errorf("unexpected error message: %v", loadErr)
	}
}

// ── loadInherited: ReportPage deserialize error propagation ──────────────────

func TestLoadInherited_ReportPageDeserializeError(t *testing.T) {
	dir, err := os.MkdirTemp("", "inherit_rpgerr_*")
	if err != nil {
		t.Fatalf("MkdirTemp: %v", err)
	}
	defer os.RemoveAll(dir)

	baseXML := `<?xml version="1.0" encoding="utf-8"?>
<Report ReportName="BaseRPGErr">
  <ReportPage Name="Page1"/>
</Report>`
	if err2 := os.WriteFile(filepath.Join(dir, "base.frx"), []byte(baseXML), 0644); err2 != nil {
		t.Fatalf("write base.frx: %v", err2)
	}

	// The inherited report contains a <ReportPage> (not <inherited>) with a
	// FailingInheritedBand inside — this triggers the deserializePage error path
	// within loadInherited at lines 233-235.
	inheritedXML := `<?xml version="1.0" encoding="utf-8"?>
<inherited BaseReport="base.frx">
  <ReportPage Name="BrokenPage">
    <FailingInheritedBand Name="Bad"/>
  </ReportPage>
</inherited>`
	childFile := filepath.Join(dir, "child.frx")
	if err3 := os.WriteFile(childFile, []byte(inheritedXML), 0644); err3 != nil {
		t.Fatalf("write child.frx: %v", err3)
	}

	r := NewReport()
	loadErr := r.Load(childFile)
	if loadErr == nil {
		t.Fatal("Load should return error when ReportPage inside inherited has bad band")
	}
	if !strings.Contains(loadErr.Error(), "deserialize page") &&
		!strings.Contains(loadErr.Error(), "intentional error") {
		t.Errorf("unexpected error message: %v", loadErr)
	}
}

// ── Load: gzip file with bad body (valid magic, bad content after header) ─────

func TestLoad_GzipBadBody(t *testing.T) {
	// Create a valid gzip stream that contains malformed XML.
	// This exercises the gzip.NewReader success path but fails at XML parsing level.
	var buf bytes.Buffer
	gz := gzip.NewWriter(&buf)
	_, _ = gz.Write([]byte("not-xml-garbage-content!!!"))
	gz.Close()

	tmp, err := os.CreateTemp("", "badbody_*.frx")
	if err != nil {
		t.Fatalf("create temp: %v", err)
	}
	if _, err2 := tmp.Write(buf.Bytes()); err2 != nil {
		tmp.Close()
		os.Remove(tmp.Name())
		t.Fatalf("write temp: %v", err2)
	}
	tmp.Close()
	defer os.Remove(tmp.Name())

	r := NewReport()
	// This should error — either XML parse error or "empty or invalid FRX document".
	err = r.Load(tmp.Name())
	if err == nil {
		t.Fatal("Load of gzip file with bad XML body should return an error")
	}
}
