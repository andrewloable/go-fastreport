package reportpkg

// pagebase_test.go — tests for the PageBase abstract-class port.
//
// C# source: original-dotnet/FastReport.Base/PageBase.cs
// C# source: original-dotnet/FastReport.Base/ReportPage.cs
//
// PageBase provides:
//   - PageName() — preview-navigator display name (falls back to Name())
//   - SetPageName() — explicit override for PageName
//   - NeedRefresh() / NeedModify() — preview-refresh flags
//   - Refresh() — sets NeedRefresh = true
//   - Modify() — sets NeedRefresh = true AND NeedModify = true
//
// ReportPage adds (ported in this pass):
//   - ExportAlias — page name override for exporters
//   - RawPaperSize — raw printer paper-size index
//   - ExtraDesignWidth — designer-only extra width flag
//   - PrintOnRollPaper — roll-paper flag for unlimited-height pages
//   - UnlimitedWidth / UnlimitedHeightValue / UnlimitedWidthValue
//   - FirstPageSource / OtherPagesSource / LastPageSource
//   - Duplex

import (
	"strings"
	"testing"

	"github.com/andrewloable/go-fastreport/serial"
)

// ── PageBase.PageName ─────────────────────────────────────────────────────────

// TestPageName_FallsBackToName verifies that PageName() returns Name() when no
// explicit override has been set. Mirrors PageBase.PageName C# getter fallback.
func TestPageName_FallsBackToName(t *testing.T) {
	pg := NewReportPage()
	pg.SetName("Page1")
	if got := pg.PageName(); got != "Page1" {
		t.Errorf("PageName() = %q, want %q", got, "Page1")
	}
}

// TestPageName_ExplicitOverride verifies that SetPageName stores an override
// that PageName() then returns instead of Name().
func TestPageName_ExplicitOverride(t *testing.T) {
	pg := NewReportPage()
	pg.SetName("Page1")
	pg.SetPageName("My Custom Page")
	if got := pg.PageName(); got != "My Custom Page" {
		t.Errorf("PageName() = %q, want %q", got, "My Custom Page")
	}
}

// TestPageName_ClearOverride verifies that setting pageName to "" reverts to Name().
func TestPageName_ClearOverride(t *testing.T) {
	pg := NewReportPage()
	pg.SetName("Page1")
	pg.SetPageName("Override")
	pg.SetPageName("") // clear
	if got := pg.PageName(); got != "Page1" {
		t.Errorf("PageName() after clear = %q, want %q", got, "Page1")
	}
}

// ── PageBase.NeedRefresh / NeedModify ─────────────────────────────────────────

// TestRefresh_SetsNeedRefresh verifies that Refresh() sets NeedRefresh to true
// without touching NeedModify. Mirrors PageBase.Refresh() C#.
func TestRefresh_SetsNeedRefresh(t *testing.T) {
	pg := NewReportPage()
	if pg.NeedRefresh() {
		t.Error("NeedRefresh should be false on a new page")
	}
	if pg.NeedModify() {
		t.Error("NeedModify should be false on a new page")
	}
	pg.Refresh()
	if !pg.NeedRefresh() {
		t.Error("NeedRefresh should be true after Refresh()")
	}
	if pg.NeedModify() {
		t.Error("NeedModify should still be false after Refresh()")
	}
}

// TestModify_SetsBothFlags verifies that Modify() sets both NeedRefresh and
// NeedModify to true. Mirrors PageBase.Modify() C#.
func TestModify_SetsBothFlags(t *testing.T) {
	pg := NewReportPage()
	pg.Modify()
	if !pg.NeedRefresh() {
		t.Error("NeedRefresh should be true after Modify()")
	}
	if !pg.NeedModify() {
		t.Error("NeedModify should be true after Modify()")
	}
}

// ── New ReportPage properties — defaults ─────────────────────────────────────

// TestNewReportPage_NewDefaults checks that all newly-added fields have their
// correct default values as specified by FastReport's ReportPage constructor.
func TestNewReportPage_NewDefaults(t *testing.T) {
	pg := NewReportPage()

	if pg.ExportAlias != "" {
		t.Errorf("ExportAlias default = %q, want empty", pg.ExportAlias)
	}
	if pg.RawPaperSize != 0 {
		t.Errorf("RawPaperSize default = %d, want 0", pg.RawPaperSize)
	}
	if pg.ExtraDesignWidth {
		t.Error("ExtraDesignWidth default should be false")
	}
	if pg.PrintOnRollPaper {
		t.Error("PrintOnRollPaper default should be false")
	}
	if pg.UnlimitedWidth {
		t.Error("UnlimitedWidth default should be false")
	}
	if pg.UnlimitedHeightValue != 0 {
		t.Errorf("UnlimitedHeightValue default = %v, want 0", pg.UnlimitedHeightValue)
	}
	if pg.UnlimitedWidthValue != 0 {
		t.Errorf("UnlimitedWidthValue default = %v, want 0", pg.UnlimitedWidthValue)
	}
	// Paper sources default to 7 (System.Drawing.Printing.PaperSourceKind.AutomaticFeed).
	if pg.FirstPageSource != 7 {
		t.Errorf("FirstPageSource default = %d, want 7", pg.FirstPageSource)
	}
	if pg.OtherPagesSource != 7 {
		t.Errorf("OtherPagesSource default = %d, want 7", pg.OtherPagesSource)
	}
	if pg.LastPageSource != 7 {
		t.Errorf("LastPageSource default = %d, want 7", pg.LastPageSource)
	}
	if pg.Duplex != "" {
		t.Errorf("Duplex default = %q, want empty", pg.Duplex)
	}
}

// ── Serialize — new fields written only when non-default ─────────────────────

// TestSerialize_NewFields verifies that newly-added fields appear in the
// serialized XML only when they differ from their defaults.
func TestSerialize_NewFields(t *testing.T) {
	pg := NewReportPage()
	pg.SetName("P1")
	pg.ExportAlias = "Alias1"
	pg.RawPaperSize = 9
	pg.ExtraDesignWidth = true
	pg.UnlimitedHeight = true
	pg.PrintOnRollPaper = true
	pg.UnlimitedWidth = true
	pg.UnlimitedHeightValue = 1200
	pg.UnlimitedWidthValue = 800
	pg.FirstPageSource = 2
	pg.OtherPagesSource = 3
	pg.LastPageSource = 4
	pg.Duplex = "Vertical"

	r := NewReport()
	r.AddPage(pg)
	xml, err := r.SaveToString()
	if err != nil {
		t.Fatalf("SaveToString: %v", err)
	}

	for _, want := range []string{
		`ExportAlias="Alias1"`,
		`RawPaperSize="9"`,
		`ExtraDesignWidth="true"`,
		`PrintOnRollPaper="true"`,
		`UnlimitedWidth="true"`,
		`UnlimitedHeightValue="1200"`,
		`UnlimitedWidthValue="800"`,
		`FirstPageSource="2"`,
		`OtherPagesSource="3"`,
		`LastPageSource="4"`,
		`Duplex="Vertical"`,
	} {
		if !strings.Contains(xml, want) {
			t.Errorf("expected %s in XML output\nfull XML:\n%s", want, xml)
		}
	}
}

// TestSerialize_DefaultsOmitted verifies that fields at their default values
// are NOT written to the XML (omit-if-default behaviour mirrors C# serializer).
func TestSerialize_DefaultsOmitted(t *testing.T) {
	pg := NewReportPage()
	pg.SetName("P1")
	// All new fields left at defaults.

	r := NewReport()
	r.AddPage(pg)
	xml, err := r.SaveToString()
	if err != nil {
		t.Fatalf("SaveToString: %v", err)
	}

	for _, absent := range []string{
		"ExportAlias",
		"RawPaperSize",
		"ExtraDesignWidth",
		"PrintOnRollPaper",
		"UnlimitedWidth",
		"UnlimitedHeightValue",
		"UnlimitedWidthValue",
		"FirstPageSource",
		"OtherPagesSource",
		"LastPageSource",
		"Duplex",
	} {
		if strings.Contains(xml, absent) {
			t.Errorf("field %s should be omitted from XML when at default value\nfull XML:\n%s", absent, xml)
		}
	}
}

// ── Deserialize — round-trip new fields ───────────────────────────────────────

// TestDeserialize_NewFields verifies that all newly-added fields round-trip
// correctly through serialize → deserialize.
func TestDeserialize_NewFields(t *testing.T) {
	xmlDoc := `<ReportPage Name="P1"
		ExportAlias="ExAlias"
		RawPaperSize="9"
		ExtraDesignWidth="true"
		UnlimitedHeight="true"
		PrintOnRollPaper="true"
		UnlimitedWidth="true"
		UnlimitedHeightValue="1200"
		UnlimitedWidthValue="800"
		FirstPageSource="2"
		OtherPagesSource="3"
		LastPageSource="4"
		Duplex="Simplex"/>`

	rdr := serial.NewReader(strings.NewReader(xmlDoc))
	typeName, ok := rdr.ReadObjectHeader()
	if !ok || typeName != "ReportPage" {
		t.Fatalf("ReadObjectHeader: got %q, ok=%v", typeName, ok)
	}

	pg := NewReportPage()
	if err := pg.Deserialize(rdr); err != nil {
		t.Fatalf("Deserialize: %v", err)
	}

	if pg.ExportAlias != "ExAlias" {
		t.Errorf("ExportAlias = %q, want ExAlias", pg.ExportAlias)
	}
	if pg.RawPaperSize != 9 {
		t.Errorf("RawPaperSize = %d, want 9", pg.RawPaperSize)
	}
	if !pg.ExtraDesignWidth {
		t.Error("ExtraDesignWidth should be true")
	}
	if !pg.UnlimitedHeight {
		t.Error("UnlimitedHeight should be true")
	}
	if !pg.PrintOnRollPaper {
		t.Error("PrintOnRollPaper should be true")
	}
	if !pg.UnlimitedWidth {
		t.Error("UnlimitedWidth should be true")
	}
	if pg.UnlimitedHeightValue != 1200 {
		t.Errorf("UnlimitedHeightValue = %v, want 1200", pg.UnlimitedHeightValue)
	}
	if pg.UnlimitedWidthValue != 800 {
		t.Errorf("UnlimitedWidthValue = %v, want 800", pg.UnlimitedWidthValue)
	}
	if pg.FirstPageSource != 2 {
		t.Errorf("FirstPageSource = %d, want 2", pg.FirstPageSource)
	}
	if pg.OtherPagesSource != 3 {
		t.Errorf("OtherPagesSource = %d, want 3", pg.OtherPagesSource)
	}
	if pg.LastPageSource != 4 {
		t.Errorf("LastPageSource = %d, want 4", pg.LastPageSource)
	}
	if pg.Duplex != "Simplex" {
		t.Errorf("Duplex = %q, want Simplex", pg.Duplex)
	}
}

// TestDeserialize_DefaultPageSources verifies that when FirstPageSource,
// OtherPagesSource, and LastPageSource are absent from the FRX they default to 7.
func TestDeserialize_DefaultPageSources(t *testing.T) {
	xmlDoc := `<ReportPage Name="P1"/>`
	rdr := serial.NewReader(strings.NewReader(xmlDoc))
	_, ok := rdr.ReadObjectHeader()
	if !ok {
		t.Fatal("ReadObjectHeader failed")
	}
	pg := NewReportPage()
	if err := pg.Deserialize(rdr); err != nil {
		t.Fatalf("Deserialize: %v", err)
	}
	if pg.FirstPageSource != 7 {
		t.Errorf("FirstPageSource = %d, want 7", pg.FirstPageSource)
	}
	if pg.OtherPagesSource != 7 {
		t.Errorf("OtherPagesSource = %d, want 7", pg.OtherPagesSource)
	}
	if pg.LastPageSource != 7 {
		t.Errorf("LastPageSource = %d, want 7", pg.LastPageSource)
	}
}

// ── Full round-trip via Report.SaveToString / LoadFromString ─────────────────

// TestRoundTrip_NewFields verifies a full save/load cycle for all new fields.
func TestRoundTrip_NewFields(t *testing.T) {
	r1 := NewReport()
	pg := NewReportPage()
	pg.SetName("RPg")
	pg.ExportAlias = "Alias"
	pg.RawPaperSize = 5
	pg.ExtraDesignWidth = true
	pg.UnlimitedWidth = true
	pg.FirstPageSource = 1
	pg.OtherPagesSource = 2
	pg.LastPageSource = 3
	pg.Duplex = "Horizontal"
	r1.AddPage(pg)

	xml, err := r1.SaveToString()
	if err != nil {
		t.Fatalf("SaveToString: %v", err)
	}

	r2 := NewReport()
	if err := r2.LoadFromString(xml); err != nil {
		t.Fatalf("LoadFromString: %v", err)
	}
	if r2.PageCount() == 0 {
		t.Fatal("no pages after round-trip")
	}
	pg2 := r2.Page(0)
	if pg2.ExportAlias != "Alias" {
		t.Errorf("ExportAlias = %q, want Alias", pg2.ExportAlias)
	}
	if pg2.RawPaperSize != 5 {
		t.Errorf("RawPaperSize = %d, want 5", pg2.RawPaperSize)
	}
	if !pg2.ExtraDesignWidth {
		t.Error("ExtraDesignWidth should be true after round-trip")
	}
	if !pg2.UnlimitedWidth {
		t.Error("UnlimitedWidth should be true after round-trip")
	}
	if pg2.FirstPageSource != 1 {
		t.Errorf("FirstPageSource = %d, want 1", pg2.FirstPageSource)
	}
	if pg2.OtherPagesSource != 2 {
		t.Errorf("OtherPagesSource = %d, want 2", pg2.OtherPagesSource)
	}
	if pg2.LastPageSource != 3 {
		t.Errorf("LastPageSource = %d, want 3", pg2.LastPageSource)
	}
	if pg2.Duplex != "Horizontal" {
		t.Errorf("Duplex = %q, want Horizontal", pg2.Duplex)
	}
}
