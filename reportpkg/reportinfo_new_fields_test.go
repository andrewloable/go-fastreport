package reportpkg

// reportinfo_new_fields_test.go — tests for the newly ported ReportInfo fields
// (Tag, SaveMode, PreviewPictureRatio) and Report fields (TextQuality,
// SmoothGraphics, ScriptLanguage) added in go-fastreport-u7abq.
//
// All tests are in package reportpkg (internal) so they can access unexported
// helpers and call Deserialize/Serialize directly via serial.NewReader.

import (
	"strings"
	"testing"

	"github.com/andrewloable/go-fastreport/serial"
)

// ── SaveMode enum ────────────────────────────────────────────────────────────

func TestSaveMode_String(t *testing.T) {
	cases := []struct {
		mode SaveMode
		want string
	}{
		{SaveModeAll, "All"},
		{SaveModeOriginal, "Original"},
		{SaveModeUser, "User"},
		{SaveModeRole, "Role"},
		{SaveModeSecurity, "Security"},
		{SaveModeDeny, "Deny"},
		{SaveModeCustom, "Custom"},
		{SaveMode(99), "All"}, // unknown defaults to All
	}
	for _, tc := range cases {
		if got := tc.mode.String(); got != tc.want {
			t.Errorf("SaveMode(%d).String() = %q, want %q", int(tc.mode), got, tc.want)
		}
	}
}

func TestParseSaveMode(t *testing.T) {
	cases := []struct {
		in   string
		want SaveMode
	}{
		{"All", SaveModeAll},
		{"Original", SaveModeOriginal},
		{"User", SaveModeUser},
		{"Role", SaveModeRole},
		{"Security", SaveModeSecurity},
		{"Deny", SaveModeDeny},
		{"Custom", SaveModeCustom},
		{"unknown", SaveModeAll}, // unknown defaults to All
		{"", SaveModeAll},
	}
	for _, tc := range cases {
		if got := parseSaveMode(tc.in); got != tc.want {
			t.Errorf("parseSaveMode(%q) = %v, want %v", tc.in, got, tc.want)
		}
	}
}

// ── TextQuality enum ─────────────────────────────────────────────────────────

func TestTextQuality_String(t *testing.T) {
	cases := []struct {
		q    TextQuality
		want string
	}{
		{TextQualityDefault, "Default"},
		{TextQualityRegular, "Regular"},
		{TextQualityClearType, "ClearType"},
		{TextQualityAntiAlias, "AntiAlias"},
		{TextQualitySingleBPP, "SingleBPP"},
		{TextQualitySingleBPPGridFit, "SingleBPPGridFit"},
		{TextQuality(99), "Default"}, // unknown defaults to Default
	}
	for _, tc := range cases {
		if got := tc.q.String(); got != tc.want {
			t.Errorf("TextQuality(%d).String() = %q, want %q", int(tc.q), got, tc.want)
		}
	}
}

func TestParseTextQuality(t *testing.T) {
	cases := []struct {
		in   string
		want TextQuality
	}{
		{"Default", TextQualityDefault},
		{"Regular", TextQualityRegular},
		{"ClearType", TextQualityClearType},
		{"AntiAlias", TextQualityAntiAlias},
		{"SingleBPP", TextQualitySingleBPP},
		{"SingleBPPGridFit", TextQualitySingleBPPGridFit},
		{"unknown", TextQualityDefault},
		{"", TextQualityDefault},
	}
	for _, tc := range cases {
		if got := parseTextQuality(tc.in); got != tc.want {
			t.Errorf("parseTextQuality(%q) = %v, want %v", tc.in, got, tc.want)
		}
	}
}

// ── ReportInfo.Tag serialization round-trip ──────────────────────────────────

func TestReportInfo_Tag_Serialize(t *testing.T) {
	r := NewReport()
	r.Info.Tag = "my-tag-value"

	xml, err := r.SaveToString()
	if err != nil {
		t.Fatalf("SaveToString: %v", err)
	}
	if !strings.Contains(xml, `ReportInfo.Tag="my-tag-value"`) {
		t.Errorf("expected ReportInfo.Tag in XML; got:\n%s", xml)
	}

	// Round-trip.
	r2 := NewReport()
	if err := r2.LoadFromString(xml); err != nil {
		t.Fatalf("LoadFromString: %v", err)
	}
	if r2.Info.Tag != "my-tag-value" {
		t.Errorf("Info.Tag after round-trip = %q, want %q", r2.Info.Tag, "my-tag-value")
	}
}

func TestReportInfo_Tag_EmptyNotSerialized(t *testing.T) {
	r := NewReport()
	r.Info.Tag = "" // empty — should not appear

	xml, err := r.SaveToString()
	if err != nil {
		t.Fatalf("SaveToString: %v", err)
	}
	if strings.Contains(xml, "ReportInfo.Tag") {
		t.Errorf("empty Tag should not appear in XML; got:\n%s", xml)
	}
}

// ── ReportInfo.SaveMode serialization round-trip ─────────────────────────────

func TestReportInfo_SaveMode_Serialize(t *testing.T) {
	cases := []struct {
		mode    SaveMode
		inXML   bool   // whether it appears in XML (All is default, not written)
		wantStr string // attribute value in XML
	}{
		{SaveModeAll, false, ""},
		{SaveModeOriginal, true, "Original"},
		{SaveModeUser, true, "User"},
		{SaveModeRole, true, "Role"},
		{SaveModeSecurity, true, "Security"},
		{SaveModeDeny, true, "Deny"},
		{SaveModeCustom, true, "Custom"},
	}
	for _, tc := range cases {
		r := NewReport()
		r.Info.SaveMode = tc.mode

		xml, err := r.SaveToString()
		if err != nil {
			t.Fatalf("SaveMode %v: SaveToString: %v", tc.mode, err)
		}

		if tc.inXML {
			want := `ReportInfo.SaveMode="` + tc.wantStr + `"`
			if !strings.Contains(xml, want) {
				t.Errorf("SaveMode %v: expected %q in XML; got:\n%s", tc.mode, want, xml)
			}
			// Round-trip.
			r2 := NewReport()
			if err := r2.LoadFromString(xml); err != nil {
				t.Fatalf("SaveMode %v: LoadFromString: %v", tc.mode, err)
			}
			if r2.Info.SaveMode != tc.mode {
				t.Errorf("SaveMode %v round-trip: got %v", tc.mode, r2.Info.SaveMode)
			}
		} else {
			if strings.Contains(xml, "ReportInfo.SaveMode") {
				t.Errorf("SaveMode All should not appear in XML; got:\n%s", xml)
			}
		}
	}
}

// ── ReportInfo.PreviewPictureRatio serialization and validation ──────────────

func TestReportInfo_PreviewPictureRatio_NonDefault(t *testing.T) {
	r := NewReport()
	r.Info.PreviewPictureRatio = 0.5

	xml, err := r.SaveToString()
	if err != nil {
		t.Fatalf("SaveToString: %v", err)
	}
	if !strings.Contains(xml, "ReportInfo.PreviewPictureRatio") {
		t.Errorf("expected ReportInfo.PreviewPictureRatio in XML; got:\n%s", xml)
	}

	// Round-trip.
	r2 := NewReport()
	if err := r2.LoadFromString(xml); err != nil {
		t.Fatalf("LoadFromString: %v", err)
	}
	if r2.Info.PreviewPictureRatio != 0.5 {
		t.Errorf("PreviewPictureRatio after round-trip = %v, want 0.5", r2.Info.PreviewPictureRatio)
	}
}

func TestReportInfo_PreviewPictureRatio_DefaultNotSerialized(t *testing.T) {
	r := NewReport()
	// Default is 0.1 — should not be serialized (matches C# diff-only serialization).

	xml, err := r.SaveToString()
	if err != nil {
		t.Fatalf("SaveToString: %v", err)
	}
	if strings.Contains(xml, "ReportInfo.PreviewPictureRatio") {
		t.Errorf("default PreviewPictureRatio should not appear in XML; got:\n%s", xml)
	}
}

func TestReportInfo_PreviewPictureRatio_ZeroOrNegativeClamped(t *testing.T) {
	// When loading an FRX with <= 0 value, it should be clamped to 0.05.
	// C# ref: ReportInfo.cs PreviewPictureRatio setter.
	xmlDoc := `<?xml version="1.0" encoding="utf-8"?><Report ReportInfo.PreviewPictureRatio="-0.5"><ReportPage Name="P1"/></Report>`
	r := NewReport()
	if err := r.LoadFromString(xmlDoc); err != nil {
		t.Fatalf("LoadFromString: %v", err)
	}
	if r.Info.PreviewPictureRatio != 0.05 {
		t.Errorf("expected clamped value 0.05, got %v", r.Info.PreviewPictureRatio)
	}
}

// ── Report.TextQuality serialization round-trip ──────────────────────────────

func TestReport_TextQuality_Serialize(t *testing.T) {
	r := NewReport()
	r.TextQuality = TextQualityClearType

	xml, err := r.SaveToString()
	if err != nil {
		t.Fatalf("SaveToString: %v", err)
	}
	if !strings.Contains(xml, `TextQuality="ClearType"`) {
		t.Errorf("expected TextQuality=\"ClearType\" in XML; got:\n%s", xml)
	}

	// Round-trip.
	r2 := NewReport()
	if err := r2.LoadFromString(xml); err != nil {
		t.Fatalf("LoadFromString: %v", err)
	}
	if r2.TextQuality != TextQualityClearType {
		t.Errorf("TextQuality after round-trip = %v, want ClearType", r2.TextQuality)
	}
}

func TestReport_TextQuality_DefaultNotSerialized(t *testing.T) {
	r := NewReport()
	// Default is TextQualityDefault — should not appear in XML.

	xml, err := r.SaveToString()
	if err != nil {
		t.Fatalf("SaveToString: %v", err)
	}
	if strings.Contains(xml, "TextQuality") {
		t.Errorf("default TextQuality should not appear in XML; got:\n%s", xml)
	}
}

// ── Report.SmoothGraphics serialization round-trip ───────────────────────────

func TestReport_SmoothGraphics_Serialize(t *testing.T) {
	r := NewReport()
	r.SmoothGraphics = true

	xml, err := r.SaveToString()
	if err != nil {
		t.Fatalf("SaveToString: %v", err)
	}
	if !strings.Contains(xml, `SmoothGraphics="true"`) {
		t.Errorf("expected SmoothGraphics=\"true\" in XML; got:\n%s", xml)
	}

	// Round-trip.
	r2 := NewReport()
	if err := r2.LoadFromString(xml); err != nil {
		t.Fatalf("LoadFromString: %v", err)
	}
	if !r2.SmoothGraphics {
		t.Error("SmoothGraphics should be true after round-trip")
	}
}

func TestReport_SmoothGraphics_FalseNotSerialized(t *testing.T) {
	r := NewReport()
	r.SmoothGraphics = false // default

	xml, err := r.SaveToString()
	if err != nil {
		t.Fatalf("SaveToString: %v", err)
	}
	if strings.Contains(xml, "SmoothGraphics") {
		t.Errorf("default SmoothGraphics=false should not appear in XML; got:\n%s", xml)
	}
}

// ── Report.ScriptLanguage round-trip ────────────────────────────────────────

func TestReport_ScriptLanguage_RoundTrip(t *testing.T) {
	r := NewReport()
	r.ScriptLanguage = "CSharp"

	xml, err := r.SaveToString()
	if err != nil {
		t.Fatalf("SaveToString: %v", err)
	}
	if !strings.Contains(xml, `ScriptLanguage="CSharp"`) {
		t.Errorf("expected ScriptLanguage=\"CSharp\" in XML; got:\n%s", xml)
	}

	// Round-trip.
	r2 := NewReport()
	if err := r2.LoadFromString(xml); err != nil {
		t.Fatalf("LoadFromString: %v", err)
	}
	if r2.ScriptLanguage != "CSharp" {
		t.Errorf("ScriptLanguage after round-trip = %q, want CSharp", r2.ScriptLanguage)
	}
}

func TestReport_ScriptLanguage_EmptyNotSerialized(t *testing.T) {
	r := NewReport()
	// Default is empty — should not appear.

	xml, err := r.SaveToString()
	if err != nil {
		t.Fatalf("SaveToString: %v", err)
	}
	if strings.Contains(xml, "ScriptLanguage") {
		t.Errorf("empty ScriptLanguage should not appear in XML; got:\n%s", xml)
	}
}

// ── Deserialize: dot-qualified ReportInfo.Tag from real C# FRX ───────────────

func TestReport_Deserialize_ReportInfoTag_DotForm(t *testing.T) {
	xmlDoc := `<Report ReportInfo.Tag="report-tag-123"/>`

	rdr := serial.NewReader(strings.NewReader(xmlDoc))
	_, ok := rdr.ReadObjectHeader()
	if !ok {
		t.Fatal("ReadObjectHeader failed")
	}

	r := NewReport()
	if err := r.Deserialize(rdr); err != nil {
		t.Fatalf("Deserialize: %v", err)
	}
	if r.Info.Tag != "report-tag-123" {
		t.Errorf("Info.Tag = %q, want report-tag-123", r.Info.Tag)
	}
}

// ── Deserialize: dot-qualified ReportInfo.SaveMode from real C# FRX ──────────

func TestReport_Deserialize_SaveMode_DotForm(t *testing.T) {
	xmlDoc := `<Report ReportInfo.SaveMode="Deny"/>`

	rdr := serial.NewReader(strings.NewReader(xmlDoc))
	_, ok := rdr.ReadObjectHeader()
	if !ok {
		t.Fatal("ReadObjectHeader failed")
	}

	r := NewReport()
	if err := r.Deserialize(rdr); err != nil {
		t.Fatalf("Deserialize: %v", err)
	}
	if r.Info.SaveMode != SaveModeDeny {
		t.Errorf("Info.SaveMode = %v, want SaveModeDeny", r.Info.SaveMode)
	}
}

// ── Deserialize: ReportInfo.CreatorVersion dot-form ──────────────────────────

func TestReport_Deserialize_CreatorVersion_DotForm(t *testing.T) {
	// C# FRX files write "ReportInfo.CreatorVersion"; old Go-generated files
	// wrote "CreatorVersion" (short form). Both should be read.
	xmlDotForm := `<Report ReportInfo.CreatorVersion="2024.5.0"/>`
	xmlShortForm := `<Report CreatorVersion="2024.3.0"/>`

	for _, tc := range []struct {
		name    string
		xml     string
		want    string
	}{
		{"dot-form", xmlDotForm, "2024.5.0"},
		{"short-form", xmlShortForm, "2024.3.0"},
	} {
		rdr := serial.NewReader(strings.NewReader(tc.xml))
		_, ok := rdr.ReadObjectHeader()
		if !ok {
			t.Fatalf("%s: ReadObjectHeader failed", tc.name)
		}
		r := NewReport()
		if err := r.Deserialize(rdr); err != nil {
			t.Fatalf("%s: Deserialize: %v", tc.name, err)
		}
		if r.Info.CreatorVersion != tc.want {
			t.Errorf("%s: CreatorVersion = %q, want %q", tc.name, r.Info.CreatorVersion, tc.want)
		}
	}
}

// ── Deserialize: ReportInfo.SavePreviewPicture fallback forms ────────────────

func TestReport_Deserialize_SavePreviewPicture_BothForms(t *testing.T) {
	dotForm := `<Report ReportInfo.SavePreviewPicture="true"/>`
	shortForm := `<Report SavePreviewPicture="true"/>`

	for _, tc := range []struct{ name, xml string }{
		{"dot-form", dotForm},
		{"short-form", shortForm},
	} {
		rdr := serial.NewReader(strings.NewReader(tc.xml))
		_, ok := rdr.ReadObjectHeader()
		if !ok {
			t.Fatalf("%s: ReadObjectHeader failed", tc.name)
		}
		r := NewReport()
		if err := r.Deserialize(rdr); err != nil {
			t.Fatalf("%s: Deserialize: %v", tc.name, err)
		}
		if !r.Info.SavePreviewPicture {
			t.Errorf("%s: SavePreviewPicture should be true", tc.name)
		}
	}
}

// ── NewReport defaults ────────────────────────────────────────────────────────

func TestNewReport_Defaults(t *testing.T) {
	r := NewReport()

	// ConvertNulls default is true matching C# ClearReportProperties().
	if !r.ConvertNulls {
		t.Error("ConvertNulls should default to true (C# default)")
	}
	// PreviewPictureRatio default is 0.1 per C# ReportInfo.Clear().
	if r.Info.PreviewPictureRatio != 0.1 {
		t.Errorf("PreviewPictureRatio should default to 0.1, got %v", r.Info.PreviewPictureRatio)
	}
	// SaveMode default is All.
	if r.Info.SaveMode != SaveModeAll {
		t.Errorf("SaveMode should default to All, got %v", r.Info.SaveMode)
	}
	// TextQuality default is Default.
	if r.TextQuality != TextQualityDefault {
		t.Errorf("TextQuality should default to Default, got %v", r.TextQuality)
	}
	// SmoothGraphics default is false.
	if r.SmoothGraphics {
		t.Error("SmoothGraphics should default to false")
	}
}

// ── All new fields round-trip together ───────────────────────────────────────

func TestReport_AllNewFields_RoundTrip(t *testing.T) {
	r := NewReport()
	r.Info.Tag = "release-1.0"
	r.Info.SaveMode = SaveModeUser
	r.Info.PreviewPictureRatio = 0.25
	r.TextQuality = TextQualityAntiAlias
	r.SmoothGraphics = true
	r.ScriptLanguage = "VbScript"

	xml, err := r.SaveToString()
	if err != nil {
		t.Fatalf("SaveToString: %v", err)
	}

	r2 := NewReport()
	if err := r2.LoadFromString(xml); err != nil {
		t.Fatalf("LoadFromString: %v", err)
	}

	if r2.Info.Tag != "release-1.0" {
		t.Errorf("Info.Tag = %q, want release-1.0", r2.Info.Tag)
	}
	if r2.Info.SaveMode != SaveModeUser {
		t.Errorf("Info.SaveMode = %v, want SaveModeUser", r2.Info.SaveMode)
	}
	if r2.Info.PreviewPictureRatio != 0.25 {
		t.Errorf("Info.PreviewPictureRatio = %v, want 0.25", r2.Info.PreviewPictureRatio)
	}
	if r2.TextQuality != TextQualityAntiAlias {
		t.Errorf("TextQuality = %v, want AntiAlias", r2.TextQuality)
	}
	if !r2.SmoothGraphics {
		t.Error("SmoothGraphics should be true after round-trip")
	}
	if r2.ScriptLanguage != "VbScript" {
		t.Errorf("ScriptLanguage = %q, want VbScript", r2.ScriptLanguage)
	}
}
