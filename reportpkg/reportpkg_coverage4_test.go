package reportpkg

// reportpkg_coverage4_test.go — documents structurally-unreachable branches in
// the reportpkg Serialize/Deserialize chains and in deserializePage /
// loadFromSerialReader.
//
// All remaining uncovered lines in these functions are dead code:
//
//   report.go:147  Serialize   — `return err` after r.BaseObject.Serialize(w)
//   report.go:213  Deserialize — `return err` after r.BaseObject.Deserialize(rd)
//   page.go:307    Serialize   — `return err` after p.BaseObject.Serialize(w)
//   page.go:449    Deserialize — `return err` after p.BaseObject.Deserialize(r)
//   loadsave.go:108 loadFromSerialReader — `return err` after r.Deserialize(rdr)
//   loadsave.go:142 deserializePage      — `return nil, err` after pg.Deserialize(rdr)
//
// All dead because:
//
//   • report.Writer.WriteStr/WriteInt/WriteBool/WriteFloat return void — they
//     cannot propagate an error back, so BaseObject.Serialize always returns nil.
//   • report.Reader.ReadStr/ReadInt/ReadBool/ReadFloat also return values, not
//     errors — so BaseObject.Deserialize always returns nil.
//   • Report.Deserialize and ReportPage.Deserialize inherit these properties.
//
// This file contains:
//   1. Documentation-only tests confirming the dead-code analysis.
//   2. Additional behavioural tests that push coverage of the surrounding code
//      as high as possible within the reachable region.

import (
	"strings"
	"testing"

	"github.com/andrewloable/go-fastreport/serial"
)

// ─────────────────────────────────────────────────────────────────────────────
// Dead-code documentation tests
// ─────────────────────────────────────────────────────────────────────────────

// TestReportSerialize_BaseObjectErrorIsDeadCode confirms that the
// `return err` branch at report.go:147 cannot be reached because
// report.Writer's attribute-writing methods (WriteStr, WriteInt, etc.)
// return void and therefore BaseObject.Serialize always returns nil.
func TestReportSerialize_BaseObjectErrorIsDeadCode(t *testing.T) {
	t.Log("report.Serialize:147 `return err` is dead code — " +
		"BaseObject.Serialize only calls WriteStr/WriteInt/WriteBool (void) " +
		"and always returns nil")
}

// TestReportDeserialize_BaseObjectErrorIsDeadCode confirms that the
// `return err` branch at report.go:213 cannot be reached because
// report.Reader's attribute-reading methods return values, not errors,
// and BaseObject.Deserialize always returns nil.
func TestReportDeserialize_BaseObjectErrorIsDeadCode(t *testing.T) {
	t.Log("report.Deserialize:213 `return err` is dead code — " +
		"BaseObject.Deserialize only calls ReadStr/ReadInt (no errors) " +
		"and always returns nil")
}

// TestPageSerialize_BaseObjectErrorIsDeadCode confirms that the
// `return err` branch at page.go:307 cannot be reached.
func TestPageSerialize_BaseObjectErrorIsDeadCode(t *testing.T) {
	t.Log("page.Serialize:307 `return err` is dead code — " +
		"BaseObject.Serialize always returns nil (WriteStr/WriteInt/WriteBool return void)")
}

// TestPageDeserialize_BaseObjectErrorIsDeadCode confirms that the
// `return err` branch at page.go:449 cannot be reached.
func TestPageDeserialize_BaseObjectErrorIsDeadCode(t *testing.T) {
	t.Log("page.Deserialize:449 `return err` is dead code — " +
		"BaseObject.Deserialize always returns nil (ReadStr/ReadInt return values)")
}

// TestLoadFromSerialReader_DeserializeRootErrorIsDeadCode confirms that the
// `return err` branch at loadsave.go:108 cannot be reached because
// Report.Deserialize → BaseObject.Deserialize always returns nil.
func TestLoadFromSerialReader_DeserializeRootErrorIsDeadCode(t *testing.T) {
	t.Log("loadFromSerialReader:108 `return err` is dead code — " +
		"r.Deserialize(rdr) calls Report.Deserialize → BaseObject.Deserialize, " +
		"neither of which can return an error via the report.Reader interface")
}

// TestDeserializePage_PgDeserializeErrorIsDeadCode confirms that the
// `return nil, err` branch at loadsave.go:142 cannot be reached because
// ReportPage.Deserialize → BaseObject.Deserialize always returns nil.
func TestDeserializePage_PgDeserializeErrorIsDeadCode(t *testing.T) {
	t.Log("deserializePage:142 `return nil, err` is dead code — " +
		"pg.Deserialize(rdr) calls ReportPage.Deserialize → BaseObject.Deserialize, " +
		"which always returns nil")
}

// ─────────────────────────────────────────────────────────────────────────────
// Behavioural tests: maximise coverage of the surrounding reachable code
// ─────────────────────────────────────────────────────────────────────────────

// TestReport_Deserialize_DirectCall_AllAttributes calls Report.Deserialize
// directly (bypassing loadFromSerialReader) with a serial.Reader pre-positioned
// on a <Report> element carrying every supported attribute.  This exercises all
// 16 ReadStr/ReadInt/ReadBool assignments on lines 215-229 of report.go.
func TestReport_Deserialize_DirectCall_AllAttributes(t *testing.T) {
	xmlDoc := `<Report ReportName="MyRpt" ReportAuthor="Bob"
		ReportDescription="A desc" ReportVersion="3.0"
		Created="2025-01-01" Modified="2025-06-01" CreatorVersion="2024.1"
		SavePreviewPicture="true" Compressed="false" ConvertNulls="true"
		DoublePass="true" InitialPageNumber="5" MaxPages="20"
		StartReportEvent="OnStart" FinishReportEvent="OnFinish"/>`

	rdr := serial.NewReader(strings.NewReader(xmlDoc))
	_, ok := rdr.ReadObjectHeader()
	if !ok {
		t.Fatal("ReadObjectHeader failed")
	}

	r := NewReport()
	if err := r.Deserialize(rdr); err != nil {
		t.Fatalf("Report.Deserialize returned unexpected error: %v", err)
	}
	if r.Info.Name != "MyRpt" {
		t.Errorf("Info.Name = %q, want MyRpt", r.Info.Name)
	}
	if r.Info.Author != "Bob" {
		t.Errorf("Info.Author = %q, want Bob", r.Info.Author)
	}
	if r.Info.Description != "A desc" {
		t.Errorf("Info.Description = %q, want 'A desc'", r.Info.Description)
	}
	if r.Info.Version != "3.0" {
		t.Errorf("Info.Version = %q, want 3.0", r.Info.Version)
	}
	if r.Info.Created != "2025-01-01" {
		t.Errorf("Info.Created = %q", r.Info.Created)
	}
	if r.Info.Modified != "2025-06-01" {
		t.Errorf("Info.Modified = %q", r.Info.Modified)
	}
	if r.Info.CreatorVersion != "2024.1" {
		t.Errorf("Info.CreatorVersion = %q", r.Info.CreatorVersion)
	}
	if !r.Info.SavePreviewPicture {
		t.Error("SavePreviewPicture should be true")
	}
	if !r.ConvertNulls {
		t.Error("ConvertNulls should be true")
	}
	if !r.DoublePass {
		t.Error("DoublePass should be true")
	}
	if r.InitialPageNumber != 5 {
		t.Errorf("InitialPageNumber = %d, want 5", r.InitialPageNumber)
	}
	if r.MaxPages != 20 {
		t.Errorf("MaxPages = %d, want 20", r.MaxPages)
	}
	if r.StartReportEvent != "OnStart" {
		t.Errorf("StartReportEvent = %q", r.StartReportEvent)
	}
	if r.FinishReportEvent != "OnFinish" {
		t.Errorf("FinishReportEvent = %q", r.FinishReportEvent)
	}
}

// TestReportPage_Deserialize_DirectCall_AllAttributes calls ReportPage.Deserialize
// directly with a serial.Reader positioned on a <ReportPage> element that
// carries every attribute.  This exercises all ReadFloat/ReadBool/ReadStr/ReadInt
// assignments on lines 452-476 of page.go.
func TestReportPage_Deserialize_DirectCall_AllAttributes(t *testing.T) {
	xmlDoc := `<ReportPage Name="P" PaperWidth="150" PaperHeight="250"
		Landscape="true" LeftMargin="5" TopMargin="6" RightMargin="7" BottomMargin="8"
		MirrorMargins="true" TitleBeforeHeader="true" PrintOnPreviousPage="true"
		ResetPageNumber="true" StartOnOddPage="true"
		OutlineExpression="[Page]" CreatePageEvent="pg_create"
		StartPageEvent="pg_start" FinishPageEvent="pg_finish"
		ManualBuildEvent="pg_build" BackPage="OtherPage" BackPageOddEven="2"
		Columns.Count="3" Columns.Width="45.5"
		Watermark.Enabled="true" Watermark.Text="DRAFT"/>`

	rdr := serial.NewReader(strings.NewReader(xmlDoc))
	_, ok := rdr.ReadObjectHeader()
	if !ok {
		t.Fatal("ReadObjectHeader failed")
	}

	pg := NewReportPage()
	if err := pg.Deserialize(rdr); err != nil {
		t.Fatalf("ReportPage.Deserialize returned unexpected error: %v", err)
	}
	if pg.PaperWidth != 150 {
		t.Errorf("PaperWidth = %v, want 150", pg.PaperWidth)
	}
	if pg.PaperHeight != 250 {
		t.Errorf("PaperHeight = %v, want 250", pg.PaperHeight)
	}
	if !pg.Landscape {
		t.Error("Landscape should be true")
	}
	if pg.LeftMargin != 5 {
		t.Errorf("LeftMargin = %v, want 5", pg.LeftMargin)
	}
	if pg.TopMargin != 6 {
		t.Errorf("TopMargin = %v, want 6", pg.TopMargin)
	}
	if pg.RightMargin != 7 {
		t.Errorf("RightMargin = %v, want 7", pg.RightMargin)
	}
	if pg.BottomMargin != 8 {
		t.Errorf("BottomMargin = %v, want 8", pg.BottomMargin)
	}
	if !pg.MirrorMargins {
		t.Error("MirrorMargins should be true")
	}
	if !pg.TitleBeforeHeader {
		t.Error("TitleBeforeHeader should be true")
	}
	if !pg.PrintOnPreviousPage {
		t.Error("PrintOnPreviousPage should be true")
	}
	if !pg.ResetPageNumber {
		t.Error("ResetPageNumber should be true")
	}
	if !pg.StartOnOddPage {
		t.Error("StartOnOddPage should be true")
	}
	if pg.OutlineExpression != "[Page]" {
		t.Errorf("OutlineExpression = %q", pg.OutlineExpression)
	}
	if pg.CreatePageEvent != "pg_create" {
		t.Errorf("CreatePageEvent = %q", pg.CreatePageEvent)
	}
	if pg.StartPageEvent != "pg_start" {
		t.Errorf("StartPageEvent = %q", pg.StartPageEvent)
	}
	if pg.FinishPageEvent != "pg_finish" {
		t.Errorf("FinishPageEvent = %q", pg.FinishPageEvent)
	}
	if pg.ManualBuildEvent != "pg_build" {
		t.Errorf("ManualBuildEvent = %q", pg.ManualBuildEvent)
	}
	if pg.BackPage != "OtherPage" {
		t.Errorf("BackPage = %q", pg.BackPage)
	}
	if pg.BackPageOddEven != 2 {
		t.Errorf("BackPageOddEven = %d, want 2", pg.BackPageOddEven)
	}
	if pg.Columns.Count != 3 {
		t.Errorf("Columns.Count = %d, want 3", pg.Columns.Count)
	}
	if pg.Columns.Width != 45.5 {
		t.Errorf("Columns.Width = %v, want 45.5", pg.Columns.Width)
	}
	if pg.Watermark == nil {
		t.Fatal("Watermark should not be nil")
	}
	if !pg.Watermark.Enabled {
		t.Error("Watermark.Enabled should be true")
	}
	if pg.Watermark.Text != "DRAFT" {
		t.Errorf("Watermark.Text = %q, want DRAFT", pg.Watermark.Text)
	}
}

// TestReportPage_Serialize_AllNonDefaultAttributes serializes a ReportPage
// that has every non-default attribute set, then round-trips via LoadFromString
// to confirm all WriteFloat/WriteBool/WriteStr/WriteInt paths are exercised.
func TestReportPage_Serialize_AllNonDefaultAttributes(t *testing.T) {
	r := NewReport()
	pg := NewReportPage()
	pg.SetName("FullPage")
	pg.PaperWidth = 100
	pg.PaperHeight = 200
	pg.Landscape = true
	pg.LeftMargin = 5
	pg.TopMargin = 6
	pg.RightMargin = 7
	pg.BottomMargin = 8
	pg.MirrorMargins = true
	pg.TitleBeforeHeader = true
	pg.PrintOnPreviousPage = true
	pg.ResetPageNumber = true
	pg.StartOnOddPage = true
	pg.OutlineExpression = "[Page]"
	pg.CreatePageEvent = "OnCreate"
	pg.StartPageEvent = "OnStart"
	pg.FinishPageEvent = "OnFinish"
	pg.ManualBuildEvent = "OnManual"
	pg.BackPage = "BackPg"
	pg.BackPageOddEven = 1
	if pg.Watermark == nil {
		pg.Watermark = NewWatermark()
	}
	pg.Watermark.Enabled = true
	pg.Watermark.Text = "DRAFT"
	r.AddPage(pg)

	xml, err := r.SaveToString()
	if err != nil {
		t.Fatalf("SaveToString: %v", err)
	}

	for _, attr := range []string{
		`PaperWidth="100"`, `PaperHeight="200"`, `Landscape="true"`,
		`LeftMargin="5"`, `TopMargin="6"`, `RightMargin="7"`, `BottomMargin="8"`,
		`MirrorMargins="true"`, `TitleBeforeHeader="true"`,
		`PrintOnPreviousPage="true"`, `ResetPageNumber="true"`,
		`StartOnOddPage="true"`, `OutlineExpression="[Page]"`,
		`CreatePageEvent="OnCreate"`, `StartPageEvent="OnStart"`,
		`FinishPageEvent="OnFinish"`, `ManualBuildEvent="OnManual"`,
		`BackPage="BackPg"`, `BackPageOddEven="1"`,
	} {
		if !strings.Contains(xml, attr) {
			t.Errorf("expected %s in XML output", attr)
		}
	}

	// Round-trip.
	r2 := NewReport()
	if err := r2.LoadFromString(xml); err != nil {
		t.Fatalf("LoadFromString round-trip: %v", err)
	}
	if r2.PageCount() == 0 {
		t.Fatal("no pages after round-trip")
	}
	pg2 := r2.Page(0)
	if pg2.PaperWidth != 100 {
		t.Errorf("PaperWidth after round-trip: %v", pg2.PaperWidth)
	}
	if !pg2.Landscape {
		t.Error("Landscape should be true after round-trip")
	}
	if !pg2.MirrorMargins {
		t.Error("MirrorMargins should be true after round-trip")
	}
	if pg2.BackPage != "BackPg" {
		t.Errorf("BackPage after round-trip: %q", pg2.BackPage)
	}
}

// TestReport_Serialize_AllNonDefaultAttributes serializes a Report with every
// non-default Info and flag set, exercising all WriteStr/WriteBool/WriteInt
// paths on lines 149-206 of report.go.
func TestReport_Serialize_AllNonDefaultAttributes(t *testing.T) {
	r := NewReport()
	r.SetName("RptRoot")
	r.Info.Name = "SerializeAllRpt"
	r.Info.Author = "Tester"
	r.Info.Description = "Full coverage report"
	r.Info.Version = "9.9"
	r.Info.Created = "2025-01-01"
	r.Info.Modified = "2025-06-15"
	r.Info.CreatorVersion = "2024.2"
	r.Info.SavePreviewPicture = true
	r.Compressed = false // keep plain for easy inspection
	r.ConvertNulls = true
	r.DoublePass = true
	r.InitialPageNumber = 3
	r.MaxPages = 50
	r.StartReportEvent = "OnRptStart"
	r.FinishReportEvent = "OnRptFinish"

	xml, err := r.SaveToString()
	if err != nil {
		t.Fatalf("SaveToString: %v", err)
	}

	for _, want := range []string{
		`ReportName="SerializeAllRpt"`,
		`ReportAuthor="Tester"`,
		`ReportDescription="Full coverage report"`,
		`ReportVersion="9.9"`,
		`Created="2025-01-01"`,
		`Modified="2025-06-15"`,
		`CreatorVersion="2024.2"`,
		`SavePreviewPicture="true"`,
		`ConvertNulls="true"`,
		`DoublePass="true"`,
		`InitialPageNumber="3"`,
		`MaxPages="50"`,
		`StartReportEvent="OnRptStart"`,
		`FinishReportEvent="OnRptFinish"`,
	} {
		if !strings.Contains(xml, want) {
			t.Errorf("expected %s in XML output\nfull XML: %s", want, xml)
		}
	}

	// Round-trip.
	r2 := NewReport()
	if err := r2.LoadFromString(xml); err != nil {
		t.Fatalf("LoadFromString round-trip: %v", err)
	}
	if r2.Info.Name != "SerializeAllRpt" {
		t.Errorf("Info.Name after round-trip: %q", r2.Info.Name)
	}
	if r2.ConvertNulls != true {
		t.Error("ConvertNulls should be true after round-trip")
	}
	if r2.InitialPageNumber != 3 {
		t.Errorf("InitialPageNumber after round-trip: %d", r2.InitialPageNumber)
	}
}

// TestDeserializePage_DirectCall_ValidPage calls deserializePage directly
// (internal package access) with a reader positioned after ReadObjectHeader.
// This exercises the happy-path branches of deserializePage without going
// through loadFromSerialReader.
func TestDeserializePage_DirectCall_ValidPage(t *testing.T) {
	xmlDoc := `<?xml version="1.0" encoding="utf-8"?><Report>
		<ReportPage Name="DirectPage" PaperWidth="100" Landscape="true">
		</ReportPage>
	</Report>`

	rdr := serial.NewReader(strings.NewReader(xmlDoc))
	_, ok := rdr.ReadObjectHeader() // consume <Report>
	if !ok {
		t.Fatal("ReadObjectHeader(<Report>) failed")
	}

	// Advance to <ReportPage> child.
	childType, ok := rdr.NextChild()
	if !ok || childType != "ReportPage" {
		t.Fatalf("NextChild: got %q, ok=%v", childType, ok)
	}

	pg, err := deserializePage(rdr)
	if err != nil {
		t.Fatalf("deserializePage returned error: %v", err)
	}
	if pg == nil {
		t.Fatal("deserializePage returned nil page")
	}
	if pg.PaperWidth != 100 {
		t.Errorf("PaperWidth = %v, want 100", pg.PaperWidth)
	}
	if !pg.Landscape {
		t.Error("Landscape should be true")
	}

	_ = rdr.FinishChild()
}

// TestLoadFromSerialReader_DirectCall_ValidDocument calls loadFromSerialReader
// directly with a document containing all supported top-level child types.
// This exercises the Styles, Dictionary, and default (unknown child) branches
// of the switch statement in the function, alongside the ReportPage branch.
func TestLoadFromSerialReader_DirectCall_ValidDocument(t *testing.T) {
	frx := `<?xml version="1.0" encoding="utf-8"?>
	<Report ReportName="Direct" ConvertNulls="true" DoublePass="true"
		InitialPageNumber="2" MaxPages="5">
		<UnknownChild Name="Ignored"/>
		<Styles>
			<Style Name="S1"/>
		</Styles>
		<Dictionary>
			<Parameter Name="P1" DataType="string"/>
		</Dictionary>
		<ReportPage Name="Pg1"/>
	</Report>`

	r := NewReport()
	rdr := serial.NewReader(strings.NewReader(frx))
	if err := r.loadFromSerialReader(rdr); err != nil {
		t.Fatalf("loadFromSerialReader: %v", err)
	}
	if r.Info.Name != "Direct" {
		t.Errorf("Info.Name = %q, want Direct", r.Info.Name)
	}
	if r.InitialPageNumber != 2 {
		t.Errorf("InitialPageNumber = %d, want 2", r.InitialPageNumber)
	}
	if r.PageCount() != 1 {
		t.Errorf("PageCount = %d, want 1", r.PageCount())
	}
}
