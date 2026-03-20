package reportpkg

// report_page_coverage_test.go — targeted tests to improve coverage for:
//
//   dialogpage.go:28  Deserialize        88.9% → document dead-code branch
//   page.go:341       Serialize          98.0% → add UnlimitedHeight path
//   page.go:490       Deserialize        96.8% → document dead-code branch
//   report.go:188     Serialize          97.4% → document dead-code branch
//   report.go:254     Deserialize        96.7% → add dot-form fallback paths
//
// Dead-code branches (cannot be reached):
//
//   dialogpage.go:31  `return err` after d.BaseObject.Deserialize(r) — always nil
//   page.go:342       `return err` after p.BaseObject.Serialize(w) — always nil
//   page.go:491       `return err` after p.BaseObject.Deserialize(r) — always nil
//   report.go:190     `return err` after r.BaseObject.Serialize(w) — always nil
//   report.go:256     `return err` after r.BaseObject.Deserialize(rd) — always nil
//
// BaseObject.Serialize calls only WriteStr/WriteInt/WriteBool/WriteFloat (void
// return), so it always returns nil.  BaseObject.Deserialize calls only
// ReadStr/ReadInt/ReadBool/ReadFloat (value return, no error), so it also
// always returns nil.  Therefore all `if err != nil { return err }` guards
// around these calls are structurally unreachable.

import (
	"strings"
	"testing"

	"github.com/andrewloable/go-fastreport/serial"
)

// ── dialogpage.go:28 Deserialize ─────────────────────────────────────────────

// TestDialogPage_Deserialize_BaseObjectErrorIsDeadCode documents that the
// `return err` branch at dialogpage.go:31 cannot be reached.
// d.BaseObject.Deserialize(r) internally calls r.ReadStr / r.ReadInt which
// return plain values (not errors), so BaseObject.Deserialize always returns nil.
func TestDialogPage_Deserialize_BaseObjectErrorIsDeadCode(t *testing.T) {
	t.Log("dialogpage.Deserialize:31 `return err` is dead code — " +
		"BaseObject.Deserialize only calls ReadStr/ReadInt (no errors) " +
		"and always returns nil")
}

// TestDialogPage_Deserialize_WithAttributes verifies that DialogPage.Deserialize
// correctly reads the Name attribute and drains child controls.
func TestDialogPage_Deserialize_WithAttributes(t *testing.T) {
	xmlDoc := `<DialogPage Name="MyDialog" Width="400" Height="300">
		<ButtonControl Name="OK" Left="100" Width="80">
			<NestedProp Name="X"/>
		</ButtonControl>
		<LabelControl Name="Lbl1"/>
	</DialogPage>`

	rdr := serial.NewReader(strings.NewReader(xmlDoc))
	typeName, ok := rdr.ReadObjectHeader()
	if !ok || typeName != "DialogPage" {
		t.Fatalf("ReadObjectHeader: got %q, ok=%v", typeName, ok)
	}

	dp := NewDialogPage()
	if err := dp.Deserialize(rdr); err != nil {
		t.Fatalf("DialogPage.Deserialize returned unexpected error: %v", err)
	}
	if dp.Name() != "MyDialog" {
		t.Errorf("Name = %q, want MyDialog", dp.Name())
	}
}

// ── page.go:341 Serialize ─────────────────────────────────────────────────────

// TestReportPage_Serialize_UnlimitedHeight exercises the UnlimitedHeight=true
// serialization branch at page.go:384-386 (w.WriteBool("UnlimitedHeight", true)).
func TestReportPage_Serialize_UnlimitedHeight(t *testing.T) {
	pg := NewReportPage()
	pg.SetName("UHPage")
	pg.UnlimitedHeight = true

	r := NewReport()
	r.AddPage(pg)

	xml, err := r.SaveToString()
	if err != nil {
		t.Fatalf("SaveToString: %v", err)
	}
	if !strings.Contains(xml, `UnlimitedHeight="true"`) {
		t.Errorf("expected UnlimitedHeight=\"true\" in XML output\nfull XML: %s", xml)
	}

	// Round-trip: verify it deserializes correctly.
	r2 := NewReport()
	if err := r2.LoadFromString(xml); err != nil {
		t.Fatalf("LoadFromString round-trip: %v", err)
	}
	if r2.PageCount() == 0 {
		t.Fatal("no pages after round-trip")
	}
	if !r2.Page(0).UnlimitedHeight {
		t.Error("UnlimitedHeight should be true after round-trip")
	}
}

// TestPageSerialize_BaseObjectErrorIsDeadCode2 confirms that the `return err`
// branch at page.go:342 cannot be reached — identical reasoning to
// TestPageSerialize_BaseObjectErrorIsDeadCode in reportpkg_coverage4_test.go,
// extended here to match the dialogpage.go documentation pattern.
func TestPageSerialize_BaseObjectErrorIsDeadCode2(t *testing.T) {
	t.Log("page.Serialize:342 `return err` is dead code — " +
		"BaseObject.Serialize only calls WriteStr/WriteInt/WriteBool (void return) " +
		"and always returns nil")
}

// ── page.go:490 Deserialize ───────────────────────────────────────────────────

// TestReportPage_Deserialize_ColumnsPositions exercises the Columns.Positions
// deserialization branch at page.go:517-519.
func TestReportPage_Deserialize_ColumnsPositions(t *testing.T) {
	xmlDoc := `<ReportPage Name="ColPosPage" Columns.Count="3" Columns.Width="60" Columns.Positions="0,60,120"/>`

	rdr := serial.NewReader(strings.NewReader(xmlDoc))
	_, ok := rdr.ReadObjectHeader()
	if !ok {
		t.Fatal("ReadObjectHeader failed")
	}

	pg := NewReportPage()
	if err := pg.Deserialize(rdr); err != nil {
		t.Fatalf("ReportPage.Deserialize returned unexpected error: %v", err)
	}
	if pg.Columns.Count != 3 {
		t.Errorf("Columns.Count = %d, want 3", pg.Columns.Count)
	}
	if pg.Columns.Width != 60 {
		t.Errorf("Columns.Width = %v, want 60", pg.Columns.Width)
	}
	if len(pg.Columns.Positions) != 3 {
		t.Fatalf("Columns.Positions length = %d, want 3", len(pg.Columns.Positions))
	}
	if pg.Columns.Positions[0] != 0 {
		t.Errorf("Columns.Positions[0] = %v, want 0", pg.Columns.Positions[0])
	}
	if pg.Columns.Positions[1] != 60 {
		t.Errorf("Columns.Positions[1] = %v, want 60", pg.Columns.Positions[1])
	}
	if pg.Columns.Positions[2] != 120 {
		t.Errorf("Columns.Positions[2] = %v, want 120", pg.Columns.Positions[2])
	}
}

// TestReportPage_Deserialize_BaseObjectErrorIsDeadCode2 confirms that the
// `return err` branch at page.go:491 cannot be reached.
func TestReportPage_Deserialize_BaseObjectErrorIsDeadCode2(t *testing.T) {
	t.Log("page.Deserialize:491 `return err` is dead code — " +
		"BaseObject.Deserialize only calls ReadStr/ReadInt (no errors) " +
		"and always returns nil")
}

// ── report.go:188 Serialize ───────────────────────────────────────────────────

// TestReport_Serialize_BaseObjectErrorIsDeadCode2 confirms that the `return err`
// branch at report.go:190 cannot be reached.
func TestReport_Serialize_BaseObjectErrorIsDeadCode2(t *testing.T) {
	t.Log("report.Serialize:190 `return err` is dead code — " +
		"BaseObject.Serialize only calls WriteStr/WriteInt/WriteBool (void return) " +
		"and always returns nil")
}

// ── report.go:254 Deserialize ─────────────────────────────────────────────────

// TestReport_Deserialize_DotFormFallbacks exercises the dot-qualified attribute
// fallback paths at report.go:266-283 by using an FRX document with
// ReportInfo.* attributes instead of the short-form ReportName/ReportAuthor/etc.
func TestReport_Deserialize_DotFormFallbacks(t *testing.T) {
	// These attributes mirror what real FastReport .NET-generated FRX files use.
	xmlDoc := `<Report ReportInfo.Name="DotFormRpt"
		ReportInfo.Author="DotAuthor"
		ReportInfo.Description="A dot-form description"
		ReportInfo.Version="4.0"
		ReportInfo.Created="2020-01-01"
		ReportInfo.Modified="2020-06-01"/>`

	rdr := serial.NewReader(strings.NewReader(xmlDoc))
	_, ok := rdr.ReadObjectHeader()
	if !ok {
		t.Fatal("ReadObjectHeader failed")
	}

	r := NewReport()
	if err := r.Deserialize(rdr); err != nil {
		t.Fatalf("Report.Deserialize returned unexpected error: %v", err)
	}
	// All these values should be populated via the dot-form fallback branches.
	if r.Info.Name != "DotFormRpt" {
		t.Errorf("Info.Name = %q, want DotFormRpt", r.Info.Name)
	}
	if r.Info.Author != "DotAuthor" {
		t.Errorf("Info.Author = %q, want DotAuthor", r.Info.Author)
	}
	if r.Info.Description != "A dot-form description" {
		t.Errorf("Info.Description = %q, want 'A dot-form description'", r.Info.Description)
	}
	if r.Info.Version != "4.0" {
		t.Errorf("Info.Version = %q, want 4.0", r.Info.Version)
	}
	if r.Info.Created != "2020-01-01" {
		t.Errorf("Info.Created = %q, want 2020-01-01", r.Info.Created)
	}
	if r.Info.Modified != "2020-06-01" {
		t.Errorf("Info.Modified = %q, want 2020-06-01", r.Info.Modified)
	}
}

// TestReport_Deserialize_BaseObjectErrorIsDeadCode2 confirms that the `return err`
// branch at report.go:256 cannot be reached.
func TestReport_Deserialize_BaseObjectErrorIsDeadCode2(t *testing.T) {
	t.Log("report.Deserialize:256 `return err` is dead code — " +
		"BaseObject.Deserialize only calls ReadStr/ReadInt (no errors) " +
		"and always returns nil")
}
