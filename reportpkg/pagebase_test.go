package reportpkg

// pagebase_test.go — tests for the PageBase abstract-class port.
//
// C# source: original-dotnet/FastReport.Base/PageBase.cs
// C# source: original-dotnet/FastReport.Base/ReportPage.cs
// C# source: original-dotnet/FastReport.Base/PageColumns.cs
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
//   - HeightInPixels() / WidthInPixels() — computed pixel dimensions
//   - PageColumns serialization round-trip (Columns.Count/Width/Positions)

import (
	"strings"
	"testing"

	"github.com/andrewloable/go-fastreport/serial"
	"github.com/andrewloable/go-fastreport/units"
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
		`OtherPageSource="3"`, // C# uses "OtherPageSource" (singular) not "OtherPagesSource"
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
		"OtherPageSource", // C# uses "OtherPageSource" (singular)
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

// ── HeightInPixels / WidthInPixels ────────────────────────────────────────────

// TestHeightInPixels_Normal verifies that HeightInPixels returns PaperHeight*Millimeters
// when UnlimitedHeight is false.
// C# source: original-dotnet/FastReport.Base/ReportPage.cs:374-379.
func TestHeightInPixels_Normal(t *testing.T) {
	pg := NewReportPage() // PaperHeight=297 by default
	want := float32(297) * units.Millimeters
	if got := pg.HeightInPixels(); got != want {
		t.Errorf("HeightInPixels() = %v, want %v", got, want)
	}
}

// TestHeightInPixels_Unlimited verifies that HeightInPixels returns UnlimitedHeightValue
// when UnlimitedHeight is true.
// C# source: original-dotnet/FastReport.Base/ReportPage.cs:377.
func TestHeightInPixels_Unlimited(t *testing.T) {
	pg := NewReportPage()
	pg.UnlimitedHeight = true
	pg.UnlimitedHeightValue = 1500
	if got := pg.HeightInPixels(); got != 1500 {
		t.Errorf("HeightInPixels() = %v, want 1500", got)
	}
}

// TestHeightInPixels_UnlimitedZeroValue verifies that HeightInPixels returns 0
// when UnlimitedHeight is true but UnlimitedHeightValue has not yet been set.
func TestHeightInPixels_UnlimitedZeroValue(t *testing.T) {
	pg := NewReportPage()
	pg.UnlimitedHeight = true
	// UnlimitedHeightValue is 0 (engine hasn't run yet).
	if got := pg.HeightInPixels(); got != 0 {
		t.Errorf("HeightInPixels() = %v, want 0 (engine not yet run)", got)
	}
}

// TestWidthInPixels_Normal verifies that WidthInPixels returns PaperWidth*Millimeters
// when UnlimitedWidth is false.
// C# source: original-dotnet/FastReport.Base/ReportPage.cs:385-398.
func TestWidthInPixels_Normal(t *testing.T) {
	pg := NewReportPage() // PaperWidth=210 by default
	want := float32(210) * units.Millimeters
	if got := pg.WidthInPixels(); got != want {
		t.Errorf("WidthInPixels() = %v, want %v", got, want)
	}
}

// TestWidthInPixels_UnlimitedWithValue verifies that WidthInPixels returns
// UnlimitedWidthValue when UnlimitedWidth is true and the value has been set.
// C#: !IsDesigning path returns UnlimitedWidthValue (ReportPage.cs:390-393).
func TestWidthInPixels_UnlimitedWithValue(t *testing.T) {
	pg := NewReportPage()
	pg.UnlimitedWidth = true
	pg.UnlimitedWidthValue = 2000
	if got := pg.WidthInPixels(); got != 2000 {
		t.Errorf("WidthInPixels() = %v, want 2000", got)
	}
}

// TestWidthInPixels_UnlimitedZeroValue verifies that WidthInPixels falls back to
// PaperWidth*Millimeters when UnlimitedWidth is true but UnlimitedWidthValue is 0.
// This handles the "before engine run" case.
func TestWidthInPixels_UnlimitedZeroValue(t *testing.T) {
	pg := NewReportPage() // PaperWidth=210
	pg.UnlimitedWidth = true
	// UnlimitedWidthValue is 0 — fall back to paper width.
	want := float32(210) * units.Millimeters
	if got := pg.WidthInPixels(); got != want {
		t.Errorf("WidthInPixels() = %v, want %v (paper fallback when value=0)", got, want)
	}
}

// TestWidthInPixels_CustomPaperSize verifies that WidthInPixels uses the actual
// PaperWidth field when it differs from the A4 default.
func TestWidthInPixels_CustomPaperSize(t *testing.T) {
	pg := NewReportPage()
	pg.PaperWidth = 420 // A3 width
	want := float32(420) * units.Millimeters
	if got := pg.WidthInPixels(); got != want {
		t.Errorf("WidthInPixels() = %v, want %v (A3 width)", got, want)
	}
}

// ── PageColumns serialization round-trip ─────────────────────────────────────

// TestPageColumns_SerializeWritesAttributes verifies that Columns.Count > 1
// causes Columns.Count, Columns.Width, and Columns.Positions to be written
// to the FRX XML output.
// C# source: original-dotnet/FastReport.Base/PageColumns.cs:101-111.
func TestPageColumns_SerializeWritesAttributes(t *testing.T) {
	pg := NewReportPage()
	pg.SetName("ColPage")
	pg.Columns.Count = 2
	pg.Columns.Width = 90
	pg.Columns.Positions = []float32{0, 90}

	r := NewReport()
	r.AddPage(pg)
	xml, err := r.SaveToString()
	if err != nil {
		t.Fatalf("SaveToString: %v", err)
	}

	for _, want := range []string{
		`Columns.Count="2"`,
		`Columns.Width="90"`,
		`Columns.Positions="0,90"`,
	} {
		if !strings.Contains(xml, want) {
			t.Errorf("expected %s in XML output\nfull XML:\n%s", want, xml)
		}
	}
}

// TestPageColumns_SingleColumn_NotSerialized verifies that the default single-column
// configuration (Count=0 or 1) is omitted from the FRX XML.
// Mirrors PageColumns.Serialize skipping when Count<=1.
func TestPageColumns_SingleColumn_NotSerialized(t *testing.T) {
	pg := NewReportPage()
	pg.SetName("SingleColPage")
	// Columns.Count defaults to 0 (single column) — nothing to serialize.

	r := NewReport()
	r.AddPage(pg)
	xml, err := r.SaveToString()
	if err != nil {
		t.Fatalf("SaveToString: %v", err)
	}

	for _, absent := range []string{"Columns.Count", "Columns.Width", "Columns.Positions"} {
		if strings.Contains(xml, absent) {
			t.Errorf("field %s should be omitted when Count<=1\nfull XML:\n%s", absent, xml)
		}
	}
}

// TestPageColumns_RoundTrip verifies that a page with multi-column settings
// round-trips correctly through serialize → deserialize (save/load cycle).
// This covers the porting gap where PageColumns serialization was missing.
func TestPageColumns_RoundTrip(t *testing.T) {
	r1 := NewReport()
	pg := NewReportPage()
	pg.SetName("ColRTPg")
	pg.Columns.Count = 3
	pg.Columns.Width = 60
	pg.Columns.Positions = []float32{0, 60, 120}
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
	if pg2.Columns.Count != 3 {
		t.Errorf("Columns.Count = %d, want 3", pg2.Columns.Count)
	}
	if pg2.Columns.Width != 60 {
		t.Errorf("Columns.Width = %v, want 60", pg2.Columns.Width)
	}
	if len(pg2.Columns.Positions) != 3 {
		t.Fatalf("Columns.Positions length = %d, want 3", len(pg2.Columns.Positions))
	}
	if pg2.Columns.Positions[0] != 0 || pg2.Columns.Positions[1] != 60 || pg2.Columns.Positions[2] != 120 {
		t.Errorf("Columns.Positions = %v, want [0 60 120]", pg2.Columns.Positions)
	}
}

// TestPageColumns_BadgesStyleRoundTrip verifies the column config seen in
// Badges.frx: 2 columns, width 90, positions "0,90". This matches the actual
// FRX attribute pattern from test-reports/Badges.frx.
func TestPageColumns_BadgesStyleRoundTrip(t *testing.T) {
	// Deserialize from the exact FRX attribute pattern from Badges.frx.
	xmlDoc := `<ReportPage Name="Page1" RawPaperSize="9" Columns.Count="2" Columns.Width="90" Columns.Positions="0,90"/>`

	rdr := serial.NewReader(strings.NewReader(xmlDoc))
	_, ok := rdr.ReadObjectHeader()
	if !ok {
		t.Fatal("ReadObjectHeader failed")
	}

	pg := NewReportPage()
	if err := pg.Deserialize(rdr); err != nil {
		t.Fatalf("Deserialize: %v", err)
	}
	if pg.Columns.Count != 2 {
		t.Errorf("Columns.Count = %d, want 2", pg.Columns.Count)
	}
	if pg.Columns.Width != 90 {
		t.Errorf("Columns.Width = %v, want 90", pg.Columns.Width)
	}
	if len(pg.Columns.Positions) != 2 {
		t.Fatalf("Columns.Positions length = %d, want 2", len(pg.Columns.Positions))
	}
	if pg.Columns.Positions[0] != 0 || pg.Columns.Positions[1] != 90 {
		t.Errorf("Columns.Positions = %v, want [0 90]", pg.Columns.Positions)
	}

	// Re-serialize and verify attributes are preserved.
	r := NewReport()
	r.AddPage(pg)
	xml, err := r.SaveToString()
	if err != nil {
		t.Fatalf("SaveToString: %v", err)
	}
	for _, want := range []string{`Columns.Count="2"`, `Columns.Width="90"`, `Columns.Positions="0,90"`} {
		if !strings.Contains(xml, want) {
			t.Errorf("expected %s in re-serialized XML\nfull XML:\n%s", want, xml)
		}
	}
}
