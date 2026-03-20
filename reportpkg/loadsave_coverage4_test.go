package reportpkg

// loadsave_coverage4_test.go — internal tests targeting the remaining uncovered
// branches in loadsave.go:
//
//   deserializeChildren  line 367-369  — else-if isContainer (Objects() without AddChild)
//   deserializeCsvConnection line 552-554 — filePath join when baseDir != ""
//   parseCsvInlineDataSet line 582-584   — base64 decode error
//   parseCsvInlineDataSet line 592-594   — XML root-element not found (EOF)
//   parseCsvInlineDataSet lines 626,631  — sibling element skip (name mismatch)
//   deserializeBusinessObjectDataSource line 790 — drain grandchildren of Column child

import (
	"encoding/base64"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/andrewloable/go-fastreport/report"
	"github.com/andrewloable/go-fastreport/serial"
)

// ─── containerOnlyObject ─────────────────────────────────────────────────────
//
// containerOnlyObject implements report.Base and has an Objects() method but
// does NOT implement the full report.Parent interface (no CanContain, no
// AddChild, etc.). This is enough to exercise the `else if isContainer` branch
// in deserializeChildren (loadsave.go:367-369).

type containerOnlyObject struct {
	report.BaseObject
	objects *report.ObjectCollection
}

func (c *containerOnlyObject) TypeName() string { return "ContainerOnlyObject" }
func (c *containerOnlyObject) Serialize(w report.Writer) error { return nil }
func (c *containerOnlyObject) Deserialize(r report.Reader) error { return nil }
func (c *containerOnlyObject) Objects() *report.ObjectCollection { return c.objects }

// TestDeserializeChildren_IsContainerFallback exercises the
// `else if isContainer` branch in deserializeChildren (loadsave.go:367-369).
//
// The parent implements Objects() (satisfying the local hasObjects interface)
// but does NOT implement report.Parent (no AddChild), so the else-if branch
// is the only way to attach the child.
func TestDeserializeChildren_IsContainerFallback(t *testing.T) {
	// XML: a TextObject child inside an otherwise empty element.
	// serial.Reader is positioned at the outer element; deserializeChildren
	// is called to drain its children.
	xmlDoc := `<Outer>
		<TextObject Name="InnerText"/>
	</Outer>`

	rdr := serial.NewReader(strings.NewReader(xmlDoc))
	_, ok := rdr.ReadObjectHeader()
	if !ok {
		t.Fatal("ReadObjectHeader failed")
	}

	parent := &containerOnlyObject{
		objects: report.NewObjectCollection(),
	}

	deserializeChildren(rdr, parent)

	if parent.objects.Len() == 0 {
		t.Error("expected child to be added to container.Objects() via isContainer branch")
	}
}

// ─── CSV baseDir join (loadsave.go:552-554) ──────────────────────────────────

// TestDeserializeCsvConnection_BaseDirJoin exercises the file-path join at
// loadsave.go:552-554: `filePath = filepath.Join(baseDir, filePath)`.
//
// This branch is only reachable when:
//   - The CSV TableDataSource has a non-absolute TableName (the csv file name)
//   - baseDir is non-empty (i.e., the FRX was loaded via Load(filename), not
//     LoadFromString)
//   - StoreData is false or TableData is empty (so the inline path is skipped)
//
// Strategy: write a minimal FRX to a temp file and call r.Load(path).
func TestDeserializeCsvConnection_BaseDirJoin(t *testing.T) {
	frx := `<?xml version="1.0" encoding="utf-8"?>
<Report>
  <Dictionary>
    <CsvDataConnection Name="CC" ConnectionString="">
      <TableDataSource Name="DS1" TableName="relative/data.csv"/>
    </CsvDataConnection>
  </Dictionary>
  <ReportPage Name="Page1"/>
</Report>`

	tmp := t.TempDir()
	frxPath := filepath.Join(tmp, "test.frx")
	if err := os.WriteFile(frxPath, []byte(frx), 0o644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	r := NewReport()
	if err := r.Load(frxPath); err != nil {
		t.Fatalf("Load: %v", err)
	}

	// Verify the data source was registered.
	sources := r.Dictionary().DataSources()
	if len(sources) == 0 {
		t.Error("expected at least one DataSource from CsvDataConnection")
	}
}

// ─── parseCsvInlineDataSet — base64 decode error (loadsave.go:582-584) ───────

// TestParseCsvInlineDataSet_InvalidBase64 exercises the `return nil` path at
// loadsave.go:582-584 when base64.StdEncoding.DecodeString returns an error.
//
// When parseCsvInlineDataSet returns nil, deserializeCsvConnection falls through
// to the file-based path, so the report still loads successfully.
func TestParseCsvInlineDataSet_InvalidBase64(t *testing.T) {
	frx := `<?xml version="1.0" encoding="utf-8"?>
<Report>
  <Dictionary>
    <CsvDataConnection Name="CC" ConnectionString="">
      <TableDataSource Name="DS1" TableName="data.csv"
        StoreData="true" TableData="!!NOT_VALID_BASE64!!"/>
    </CsvDataConnection>
  </Dictionary>
  <ReportPage Name="Page1"/>
</Report>`

	r := NewReport()
	// Must not panic or return an error — parseCsvInlineDataSet returns nil,
	// and the code falls back to a file-based source.
	if err := r.LoadFromString(frx); err != nil {
		t.Fatalf("LoadFromString with invalid base64 TableData: %v", err)
	}

	// A file-based data source is added instead.
	sources := r.Dictionary().DataSources()
	if len(sources) == 0 {
		t.Error("expected fallback file-based DataSource when base64 decode fails")
	}
}

// ─── parseCsvInlineDataSet — XML root not found (loadsave.go:592-594) ────────

// TestParseCsvInlineDataSet_NoXMLRoot exercises the `return nil` path at
// loadsave.go:592-594 when the decoded bytes contain no XML StartElement token
// (e.g., valid base64 of plain text that the XML decoder cannot tokenise).
//
// "hello" → base64: "aGVsbG8=" — the XML decoder returns an error immediately.
func TestParseCsvInlineDataSet_NoXMLRoot(t *testing.T) {
	notXML := base64.StdEncoding.EncodeToString([]byte("this is not xml at all"))

	frx := `<?xml version="1.0" encoding="utf-8"?>
<Report>
  <Dictionary>
    <CsvDataConnection Name="CC" ConnectionString="">
      <TableDataSource Name="DS2" TableName="data.csv"
        StoreData="true" TableData="` + notXML + `"/>
    </CsvDataConnection>
  </Dictionary>
  <ReportPage Name="Page1"/>
</Report>`

	r := NewReport()
	if err := r.LoadFromString(frx); err != nil {
		t.Fatalf("LoadFromString with non-XML base64 TableData: %v", err)
	}

	// Fallback file-based source should be present.
	sources := r.Dictionary().DataSources()
	if len(sources) == 0 {
		t.Error("expected fallback file-based DataSource when XML root not found")
	}
}

// ─── parseCsvInlineDataSet — sibling element name mismatch (loadsave.go:626,631) ─

// TestParseCsvInlineDataSet_SiblingElementMismatch exercises the
// `se.Name.Local != rowElemName` branch in parseCsvInlineDataSet
// (loadsave.go:626-631).
//
// The inline XML dataset has a root element followed by rows whose local name
// does NOT match the TableName. The function skips them (line 628 dec.Skip(),
// line 631 continue) and returns a data source with zero rows.
func TestParseCsvInlineDataSet_SiblingElementMismatch(t *testing.T) {
	// tableName in FRX: "myrows"
	// Inline XML uses <otherrows> elements — name mismatch.
	inlineXML := `<NewDataSet>
  <otherrows><col1>val1</col1></otherrows>
  <otherrows><col1>val2</col1></otherrows>
</NewDataSet>`
	b64 := base64.StdEncoding.EncodeToString([]byte(inlineXML))

	frx := `<?xml version="1.0" encoding="utf-8"?>
<Report>
  <Dictionary>
    <CsvDataConnection Name="CC" ConnectionString="">
      <TableDataSource Name="DS3" TableName="myrows"
        StoreData="true" TableData="` + b64 + `"/>
    </CsvDataConnection>
  </Dictionary>
  <ReportPage Name="Page1"/>
</Report>`

	r := NewReport()
	if err := r.LoadFromString(frx); err != nil {
		t.Fatalf("LoadFromString with sibling-mismatch TableData: %v", err)
	}

	// Inline path was taken (parseCsvInlineDataSet returns a 0-row BaseDataSource).
	sources := r.Dictionary().DataSources()
	if len(sources) == 0 {
		t.Error("expected at least one DataSource (0-row inline) when sibling names mismatch")
	}
}

// ─── parseCsvInlineDataSet — happy path with namespace-qualified child ────────

// TestParseCsvInlineDataSet_NamespaceSkip exercises the namespace-qualified
// element skip at loadsave.go:620-624 (the xs:schema handling).
//
// The inline DataSet XML has an xs:schema element (namespace-qualified) that
// must be skipped, followed by real row data.
func TestParseCsvInlineDataSet_NamespaceSkip(t *testing.T) {
	// An XSD-style schema element followed by row data.
	// The namespace prefix is "xs" mapped to a dummy URI.
	inlineXML := `<NewDataSet xmlns:xs="http://www.w3.org/2001/XMLSchema">
  <xs:schema><xs:element name="col1" type="xs:string"/></xs:schema>
  <myrows><col1>hello</col1></myrows>
</NewDataSet>`
	b64 := base64.StdEncoding.EncodeToString([]byte(inlineXML))

	frx := `<?xml version="1.0" encoding="utf-8"?>
<Report>
  <Dictionary>
    <CsvDataConnection Name="CC2" ConnectionString="">
      <TableDataSource Name="DS4" TableName="myrows"
        StoreData="true" TableData="` + b64 + `"/>
    </CsvDataConnection>
  </Dictionary>
  <ReportPage Name="Page1"/>
</Report>`

	r := NewReport()
	if err := r.LoadFromString(frx); err != nil {
		t.Fatalf("LoadFromString with namespace-prefixed schema: %v", err)
	}

	sources := r.Dictionary().DataSources()
	if len(sources) == 0 {
		t.Error("expected at least one DataSource")
	}
}

// ─── deserializeBusinessObjectDataSource — Column with grandchildren (line 790) ─

// TestDeserializeBusinessObjectDataSource_ColumnWithGrandchildren exercises
// the grandchild drain loop at loadsave.go:785-791 (line 790 specifically).
//
// The BusinessObjectDataSource has a Column child that itself has grandchildren.
// The inner drain loop at line 785-791 calls rdr.FinishChild() for each
// grandchild (line 790).
func TestDeserializeBusinessObjectDataSource_ColumnWithGrandchildren(t *testing.T) {
	frx := `<?xml version="1.0" encoding="utf-8"?>
<Report>
  <Dictionary>
    <BusinessObjectDataSource Name="Orders" Alias="Orders">
      <Column Name="ID">
        <Metadata Key="type" Value="int"/>
        <Metadata Key="nullable" Value="false"/>
      </Column>
    </BusinessObjectDataSource>
  </Dictionary>
  <ReportPage Name="Page1"/>
</Report>`

	r := NewReport()
	if err := r.LoadFromString(frx); err != nil {
		t.Fatalf("LoadFromString with BusinessObject Column with grandchildren: %v", err)
	}

	sources := r.Dictionary().DataSources()
	if len(sources) == 0 {
		t.Error("expected at least one BusinessObjectDataSource")
	}
}

// ─── deserializeReportBody — DialogPage child (loadsave.go:168-173) ──────────

// TestDeserializeReportBody_DialogPage exercises the DialogPage branch in
// deserializeReportBody (loadsave.go:168-173), which deserializes and discards
// dialog pages used only in the .NET designer.
func TestDeserializeReportBody_DialogPage(t *testing.T) {
	frx := `<?xml version="1.0" encoding="utf-8"?>
<Report>
  <DialogPage Name="Dialog1"/>
  <ReportPage Name="Page1"/>
</Report>`

	r := NewReport()
	if err := r.LoadFromString(frx); err != nil {
		t.Fatalf("LoadFromString with DialogPage: %v", err)
	}

	// The dialog page is discarded; only the ReportPage is kept.
	if r.PageCount() != 1 {
		t.Errorf("expected 1 ReportPage (DialogPage discarded), got %d", r.PageCount())
	}
}

// ─── deserializePage — unknown band type (loadsave.go:309-314) ───────────────

// TestDeserializePage_UnknownBandType exercises the unknown-type skip at
// loadsave.go:309-314, where a page child element with an unregistered type
// name is skipped gracefully.
func TestDeserializePage_UnknownBandType(t *testing.T) {
	frx := `<?xml version="1.0" encoding="utf-8"?>
<Report>
  <ReportPage Name="Page1">
    <UnknownBand_XYZ123 Name="BadBand"/>
    <PageHeaderBand Name="PH"/>
  </ReportPage>
</Report>`

	r := NewReport()
	if err := r.LoadFromString(frx); err != nil {
		t.Fatalf("LoadFromString with unknown band type: %v", err)
	}

	if r.PageCount() != 1 {
		t.Fatalf("expected 1 page, got %d", r.PageCount())
	}
}

// ─── deserializeChildren — unknown child type (loadsave.go:355-359) ──────────

// TestDeserializeChildren_UnknownChildType exercises the unknown-type skip at
// loadsave.go:355-359 in deserializeChildren. An unregistered child element
// inside a band should be silently skipped.
func TestDeserializeChildren_UnknownChildType(t *testing.T) {
	frx := `<?xml version="1.0" encoding="utf-8"?>
<Report>
  <ReportPage Name="Page1">
    <PageHeaderBand Name="PH">
      <UnknownObject_ZZZZ Name="Bad"/>
      <TextObject Name="T1"/>
    </PageHeaderBand>
  </ReportPage>
</Report>`

	r := NewReport()
	if err := r.LoadFromString(frx); err != nil {
		t.Fatalf("LoadFromString with unknown child type: %v", err)
	}

	if r.PageCount() != 1 {
		t.Fatalf("expected 1 page, got %d", r.PageCount())
	}
}

// ─── parseCsvInlineDataSet — successful full parse (happy path) ───────────────

// TestParseCsvInlineDataSet_FullParse exercises the full happy path of
// parseCsvInlineDataSet: valid base64 → valid XML → rows parsed → BaseDataSource
// returned with correct columns and rows.
func TestParseCsvInlineDataSet_FullParse(t *testing.T) {
	inlineXML := `<NewDataSet>
  <myrows><Name>Alice</Name><Age>30</Age></myrows>
  <myrows><Name>Bob</Name><Age>25</Age></myrows>
</NewDataSet>`
	b64 := base64.StdEncoding.EncodeToString([]byte(inlineXML))

	frx := `<?xml version="1.0" encoding="utf-8"?>
<Report>
  <Dictionary>
    <CsvDataConnection Name="CC3" ConnectionString="">
      <TableDataSource Name="DS5" Alias="People" TableName="myrows"
        StoreData="true" TableData="` + b64 + `"/>
    </CsvDataConnection>
  </Dictionary>
  <ReportPage Name="Page1"/>
</Report>`

	r := NewReport()
	if err := r.LoadFromString(frx); err != nil {
		t.Fatalf("LoadFromString with valid inline CSV dataset: %v", err)
	}

	sources := r.Dictionary().DataSources()
	if len(sources) == 0 {
		t.Fatal("expected at least one inline DataSource")
	}

	// The inline source should have 2 rows.
	ds := sources[0]
	if ds.RowCount() != 2 {
		t.Errorf("expected 2 rows, got %d", ds.RowCount())
	}
}
