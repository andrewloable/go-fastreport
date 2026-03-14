package reportpkg_test

// FRX file integration tests.
// These tests load sample .frx files from test-reports/, verify the deserialized
// object graph, round-trip through SaveToString/LoadFromString, and check that
// all properties survive intact.

import (
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/andrewloable/go-fastreport/band"
	"github.com/andrewloable/go-fastreport/object"
	"github.com/andrewloable/go-fastreport/reportpkg"
)

// testReportsDir returns the absolute path to the test-reports/ directory.
func testReportsDir() string {
	_, thisFile, _, _ := runtime.Caller(0)
	return filepath.Join(filepath.Dir(thisFile), "..", "test-reports")
}

// loadFRX loads an FRX file from test-reports/ and returns the Report.
func loadFRX(t *testing.T, name string) *reportpkg.Report {
	t.Helper()
	path := filepath.Join(testReportsDir(), name)
	r := reportpkg.NewReport()
	if err := r.Load(path); err != nil {
		t.Fatalf("Load(%q): %v", name, err)
	}
	return r
}

// roundTripFRX serializes r to a string and re-loads it.
func roundTripFRX(t *testing.T, r *reportpkg.Report) *reportpkg.Report {
	t.Helper()
	xmlStr, err := r.SaveToString()
	if err != nil {
		t.Fatalf("SaveToString: %v", err)
	}
	r2 := reportpkg.NewReport()
	if err := r2.LoadFromString(xmlStr); err != nil {
		t.Fatalf("LoadFromString: %v\nXML:\n%s", err, xmlStr)
	}
	return r2
}

// ── simple-list.frx ──────────────────────────────────────────────────────────

func TestFRXFile_SimpleList_Load(t *testing.T) {
	r := loadFRX(t, "simple-list.frx")

	if r.Info.Name != "SimpleList" {
		t.Errorf("ReportName: got %q, want SimpleList", r.Info.Name)
	}
	if r.Info.Author != "Test" {
		t.Errorf("ReportAuthor: got %q, want Test", r.Info.Author)
	}

	pages := r.Pages()
	if len(pages) != 1 {
		t.Fatalf("Pages: got %d, want 1", len(pages))
	}
	pg := pages[0]
	if pg.Name() != "Page1" {
		t.Errorf("Page Name: got %q, want Page1", pg.Name())
	}

	// PageHeader must be loaded.
	if pg.PageHeader() == nil {
		t.Fatal("PageHeader should not be nil")
	}
	if pg.PageHeader().Name() != "PageHeader1" {
		t.Errorf("PageHeader Name: got %q", pg.PageHeader().Name())
	}

	// Check PageHeader has its TextObject.
	if pg.PageHeader().Objects().Len() != 1 {
		t.Errorf("PageHeader objects: got %d, want 1", pg.PageHeader().Objects().Len())
	} else {
		txt, ok := pg.PageHeader().Objects().Get(0).(*object.TextObject)
		if !ok {
			t.Errorf("PageHeader child type: got %T, want *object.TextObject", pg.PageHeader().Objects().Get(0))
		} else if txt.Text() != "Simple List Report" {
			t.Errorf("PageHeader text: got %q", txt.Text())
		}
	}

	// DataBand.
	if len(pg.Bands()) != 1 {
		t.Fatalf("Dynamic bands: got %d, want 1", len(pg.Bands()))
	}
	db, ok := pg.Bands()[0].(*band.DataBand)
	if !ok {
		t.Fatalf("Band type: got %T, want *band.DataBand", pg.Bands()[0])
	}
	if db.Name() != "DataBand1" {
		t.Errorf("DataBand Name: got %q", db.Name())
	}
	if db.Filter() != `[Value] != ""` {
		t.Errorf("DataBand Filter: got %q", db.Filter())
	}

	// PageFooter.
	if pg.PageFooter() == nil {
		t.Fatal("PageFooter should not be nil")
	}
}

func TestFRXFile_SimpleList_RoundTrip(t *testing.T) {
	r := loadFRX(t, "simple-list.frx")
	r2 := roundTripFRX(t, r)

	pages := r2.Pages()
	if len(pages) != 1 {
		t.Fatalf("Pages after round-trip: got %d, want 1", len(pages))
	}
	pg := pages[0]

	// PageHeader must survive.
	if pg.PageHeader() == nil {
		t.Fatal("PageHeader nil after round-trip")
	}

	// DataBand and Filter must survive.
	if len(pg.Bands()) != 1 {
		t.Fatalf("Bands after round-trip: got %d, want 1", len(pg.Bands()))
	}
	db, ok := pg.Bands()[0].(*band.DataBand)
	if !ok {
		t.Fatalf("Band type after round-trip: %T", pg.Bands()[0])
	}
	if db.Filter() != `[Value] != ""` {
		t.Errorf("DataBand Filter after round-trip: got %q", db.Filter())
	}

	// PageFooter must survive.
	if pg.PageFooter() == nil {
		t.Fatal("PageFooter nil after round-trip")
	}
}

// ── grouped.frx ──────────────────────────────────────────────────────────────

func TestFRXFile_Grouped_Load(t *testing.T) {
	r := loadFRX(t, "grouped.frx")

	if r.Info.Name != "GroupedReport" {
		t.Errorf("ReportName: got %q", r.Info.Name)
	}

	pages := r.Pages()
	if len(pages) != 1 {
		t.Fatalf("Pages: got %d, want 1", len(pages))
	}
	pg := pages[0]

	// Expect: GroupHeader, Data, GroupFooter, ReportSummary in dynamic bands.
	// PageHeader and PageFooter in their slots.
	if pg.PageHeader() == nil {
		t.Fatal("PageHeader nil")
	}
	if pg.PageFooter() == nil {
		t.Fatal("PageFooter nil")
	}
	if pg.ReportSummary() == nil {
		t.Fatal("ReportSummary nil")
	}

	bands := pg.Bands()
	// Expect GroupHeader + DataBand + GroupFooter = 3 dynamic bands.
	if len(bands) != 3 {
		t.Fatalf("Dynamic bands: got %d, want 3", len(bands))
	}

	gh, ok := bands[0].(*band.GroupHeaderBand)
	if !ok {
		t.Fatalf("bands[0] type: %T, want *band.GroupHeaderBand", bands[0])
	}
	if gh.Condition() != "[Category]" {
		t.Errorf("GroupHeader Condition: got %q", gh.Condition())
	}

	if _, ok := bands[1].(*band.DataBand); !ok {
		t.Fatalf("bands[1] type: %T, want *band.DataBand", bands[1])
	}

	if _, ok := bands[2].(*band.GroupFooterBand); !ok {
		t.Fatalf("bands[2] type: %T, want *band.GroupFooterBand", bands[2])
	}
}

func TestFRXFile_Grouped_RoundTrip(t *testing.T) {
	r := loadFRX(t, "grouped.frx")
	r2 := roundTripFRX(t, r)

	pages := r2.Pages()
	if len(pages) != 1 {
		t.Fatalf("Pages: got %d, want 1", len(pages))
	}
	pg := pages[0]

	if pg.ReportSummary() == nil {
		t.Fatal("ReportSummary nil after round-trip")
	}
	if len(pg.Bands()) != 3 {
		t.Fatalf("Dynamic bands after round-trip: got %d, want 3", len(pg.Bands()))
	}

	gh, ok := pg.Bands()[0].(*band.GroupHeaderBand)
	if !ok {
		t.Fatalf("bands[0] type: %T", pg.Bands()[0])
	}
	if gh.Condition() != "[Category]" {
		t.Errorf("GroupHeader Condition: got %q", gh.Condition())
	}
}

// ── landscape-a3.frx ─────────────────────────────────────────────────────────

func TestFRXFile_LandscapeA3_Load(t *testing.T) {
	r := loadFRX(t, "landscape-a3.frx")

	if !r.DoublePass {
		t.Error("DoublePass should be true")
	}

	pages := r.Pages()
	if len(pages) != 1 {
		t.Fatalf("Pages: got %d, want 1", len(pages))
	}
	pg := pages[0]

	if pg.PaperWidth != 420 {
		t.Errorf("PaperWidth: got %v, want 420", pg.PaperWidth)
	}
	if !pg.Landscape {
		t.Error("Landscape should be true")
	}
	if pg.LeftMargin != 15 {
		t.Errorf("LeftMargin: got %v, want 15", pg.LeftMargin)
	}

	// ReportTitle, ColumnHeader, DataBand, ColumnFooter.
	if pg.ReportTitle() == nil {
		t.Fatal("ReportTitle nil")
	}
	if pg.ColumnHeader() == nil {
		t.Fatal("ColumnHeader nil")
	}
	if pg.ColumnFooter() == nil {
		t.Fatal("ColumnFooter nil")
	}
	if len(pg.Bands()) != 1 {
		t.Fatalf("Dynamic bands: got %d, want 1", len(pg.Bands()))
	}

	// ColumnHeader should have 2 TextObjects.
	ch := pg.ColumnHeader()
	if ch.Objects().Len() != 2 {
		t.Errorf("ColumnHeader objects: got %d, want 2", ch.Objects().Len())
	}
}

func TestFRXFile_LandscapeA3_RoundTrip(t *testing.T) {
	r := loadFRX(t, "landscape-a3.frx")
	xmlStr, err := r.SaveToString()
	if err != nil {
		t.Fatalf("SaveToString: %v", err)
	}

	// Verify key elements appear in the saved XML.
	for _, want := range []string{
		`PaperWidth="420"`, `Landscape="true"`, `LeftMargin="15"`,
		`<ReportTitle `, `<ColumnHeader `, `<ColumnFooter `,
		`<Data `,
	} {
		if !strings.Contains(xmlStr, want) {
			t.Errorf("SaveToString: XML missing %q", want)
		}
	}

	// Re-load and verify.
	r2 := reportpkg.NewReport()
	if err := r2.LoadFromString(xmlStr); err != nil {
		t.Fatalf("LoadFromString: %v", err)
	}
	pages := r2.Pages()
	if len(pages) != 1 {
		t.Fatalf("Pages after round-trip: %d", len(pages))
	}
	pg := pages[0]
	if !pg.Landscape {
		t.Error("Landscape should be true after round-trip")
	}
	if pg.ReportTitle() == nil {
		t.Fatal("ReportTitle nil after round-trip")
	}
	if pg.ColumnHeader() == nil {
		t.Fatal("ColumnHeader nil after round-trip")
	}
	if ch := pg.ColumnHeader(); ch.Objects().Len() != 2 {
		t.Errorf("ColumnHeader objects after round-trip: got %d, want 2", ch.Objects().Len())
	}
}
