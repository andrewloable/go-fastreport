package object_test

// zipcode_test.go — external tests for ZipCodeObject.
//
// Verifies:
//   - Constructor defaults match FastReport .NET (ZipCodeObject.cs line 362-378)
//   - All getters/setters work correctly
//   - Serialize/Deserialize round-trips all fields
//   - Diff-based Serialize: default values are NOT written to XML
//   - Non-default values ARE written and recovered correctly
//   - GetExpressions returns DataColumn and Expression when set
//   - Serial registry creates a ZipCodeObject via "ZipCodeObject" name

import (
	"bytes"
	"strings"
	"testing"

	"github.com/andrewloable/go-fastreport/object"
	"github.com/andrewloable/go-fastreport/reportpkg"
	"github.com/andrewloable/go-fastreport/serial"
)

// ── Constructor defaults ──────────────────────────────────────────────────────

// TestZipCodeObject_Constructor_MatchesCSharpDefaults verifies that
// NewZipCodeObject initialises all fields to the C# defaults from
// ZipCodeObject.cs lines 362-378.
func TestZipCodeObject_Constructor_MatchesCSharpDefaults(t *testing.T) {
	z := object.NewZipCodeObject()
	if z == nil {
		t.Fatal("NewZipCodeObject returned nil")
	}

	// text = "123456"  (ZipCodeObject.cs:371)
	if z.Text() != "123456" {
		t.Errorf("Text default = %q, want 123456", z.Text())
	}
	// dataColumn = ""  (ZipCodeObject.cs:372)
	if z.DataColumn() != "" {
		t.Errorf("DataColumn default = %q, want empty", z.DataColumn())
	}
	// expression = ""  (ZipCodeObject.cs:373)
	if z.Expression() != "" {
		t.Errorf("Expression default = %q, want empty", z.Expression())
	}
	// segmentWidth = Units.Centimeters * 0.5f = 18.9  (ZipCodeObject.cs:364)
	if z.SegmentWidth() != 18.9 {
		t.Errorf("SegmentWidth default = %v, want 18.9", z.SegmentWidth())
	}
	// segmentHeight = Units.Centimeters * 1 = 37.8  (ZipCodeObject.cs:365)
	if z.SegmentHeight() != 37.8 {
		t.Errorf("SegmentHeight default = %v, want 37.8", z.SegmentHeight())
	}
	// spacing = Units.Centimeters * 0.9f = 34.02  (ZipCodeObject.cs:366)
	if z.Spacing() != 34.02 {
		t.Errorf("Spacing default = %v, want 34.02", z.Spacing())
	}
	// segmentCount = 6  (ZipCodeObject.cs:367)
	if z.SegmentCount() != 6 {
		t.Errorf("SegmentCount default = %d, want 6", z.SegmentCount())
	}
	// showMarkers = true  (ZipCodeObject.cs:368)
	if !z.ShowMarkers() {
		t.Error("ShowMarkers default should be true")
	}
	// showGrid = true  (ZipCodeObject.cs:369)
	if !z.ShowGrid() {
		t.Error("ShowGrid default should be true")
	}
}

// ── Getters / Setters ─────────────────────────────────────────────────────────

func TestZipCodeObject_AllGettersSetters(t *testing.T) {
	z := object.NewZipCodeObject()

	z.SetText("99999")
	if z.Text() != "99999" {
		t.Errorf("Text = %q, want 99999", z.Text())
	}

	z.SetDataColumn("Orders.ZipCode")
	if z.DataColumn() != "Orders.ZipCode" {
		t.Errorf("DataColumn = %q, want Orders.ZipCode", z.DataColumn())
	}

	z.SetExpression("[Customer.Zip]")
	if z.Expression() != "[Customer.Zip]" {
		t.Errorf("Expression = %q, want [Customer.Zip]", z.Expression())
	}

	z.SetSegmentWidth(9.45)
	if z.SegmentWidth() != 9.45 {
		t.Errorf("SegmentWidth = %v, want 9.45", z.SegmentWidth())
	}

	z.SetSegmentHeight(18.9)
	if z.SegmentHeight() != 18.9 {
		t.Errorf("SegmentHeight = %v, want 18.9", z.SegmentHeight())
	}

	z.SetSpacing(17.01)
	if z.Spacing() != 17.01 {
		t.Errorf("Spacing = %v, want 17.01", z.Spacing())
	}

	z.SetSegmentCount(5)
	if z.SegmentCount() != 5 {
		t.Errorf("SegmentCount = %d, want 5", z.SegmentCount())
	}

	z.SetShowMarkers(false)
	if z.ShowMarkers() {
		t.Error("ShowMarkers should be false after SetShowMarkers(false)")
	}
	z.SetShowMarkers(true)
	if !z.ShowMarkers() {
		t.Error("ShowMarkers should be true after SetShowMarkers(true)")
	}

	z.SetShowGrid(false)
	if z.ShowGrid() {
		t.Error("ShowGrid should be false after SetShowGrid(false)")
	}
	z.SetShowGrid(true)
	if !z.ShowGrid() {
		t.Error("ShowGrid should be true after SetShowGrid(true)")
	}
}

// ── Serialize: defaults produce empty (no-attribute) XML ─────────────────────

// TestZipCodeObject_Serialize_DefaultsNotWritten verifies that a freshly
// constructed ZipCodeObject produces XML with no ZipCode-specific attributes,
// mirroring the C# diff-based Serialize (ZipCodeObject.cs:295-320).
func TestZipCodeObject_Serialize_DefaultsNotWritten(t *testing.T) {
	z := object.NewZipCodeObject()

	var buf bytes.Buffer
	w := serial.NewWriter(&buf)
	if err := w.WriteObjectNamed("ZipCodeObject", z); err != nil {
		t.Fatalf("Serialize: %v", err)
	}
	_ = w.Flush()

	xml := buf.String()
	for _, attr := range []string{
		"Text=", "SegmentWidth=", "SegmentHeight=", "Spacing=",
		"SegmentCount=", "ShowMarkers=", "ShowGrid=",
	} {
		if strings.Contains(xml, attr) {
			t.Errorf("default value %q should not appear in XML:\n%s", attr, xml)
		}
	}
}

// ── Serialize + Deserialize: round-trip all non-default fields ────────────────

// TestZipCodeObject_RoundTrip_NonDefaultFields verifies that all non-default
// field values survive a Serialize → Deserialize cycle.
func TestZipCodeObject_RoundTrip_NonDefaultFields(t *testing.T) {
	orig := object.NewZipCodeObject()
	orig.SetText("90210")           // non-default
	orig.SetDataColumn("Addr.Zip")  // non-default
	orig.SetExpression("[Zip]")     // non-default
	orig.SetSegmentWidth(9.45)      // non-default (default 18.9)
	orig.SetSegmentHeight(18.9)     // non-default (default 37.8)
	orig.SetSpacing(17.01)          // non-default (default 34.02)
	orig.SetSegmentCount(5)         // non-default (default 6)
	orig.SetShowMarkers(false)      // non-default (default true)
	orig.SetShowGrid(false)         // non-default (default true)

	var buf bytes.Buffer
	w := serial.NewWriter(&buf)
	if err := w.WriteObjectNamed("ZipCodeObject", orig); err != nil {
		t.Fatalf("Serialize: %v", err)
	}
	_ = w.Flush()

	xml := buf.String()
	for _, attr := range []string{
		"Text=", "DataColumn=", "Expression=", "SegmentWidth=",
		"SegmentHeight=", "Spacing=", "SegmentCount=", "ShowMarkers=", "ShowGrid=",
	} {
		if !strings.Contains(xml, attr) {
			t.Errorf("expected %q in XML:\n%s", attr, xml)
		}
	}

	r := serial.NewReader(bytes.NewReader(buf.Bytes()))
	_, ok := r.ReadObjectHeader()
	if !ok {
		t.Fatal("ReadObjectHeader failed")
	}
	got := object.NewZipCodeObject()
	if err := got.Deserialize(r); err != nil {
		t.Fatalf("Deserialize: %v", err)
	}

	if got.Text() != "90210" {
		t.Errorf("Text: got %q, want 90210", got.Text())
	}
	if got.DataColumn() != "Addr.Zip" {
		t.Errorf("DataColumn: got %q, want Addr.Zip", got.DataColumn())
	}
	if got.Expression() != "[Zip]" {
		t.Errorf("Expression: got %q, want [Zip]", got.Expression())
	}
	if got.SegmentWidth() != 9.45 {
		t.Errorf("SegmentWidth: got %v, want 9.45", got.SegmentWidth())
	}
	if got.SegmentHeight() != 18.9 {
		t.Errorf("SegmentHeight: got %v, want 18.9", got.SegmentHeight())
	}
	if got.Spacing() != 17.01 {
		t.Errorf("Spacing: got %v, want 17.01", got.Spacing())
	}
	if got.SegmentCount() != 5 {
		t.Errorf("SegmentCount: got %d, want 5", got.SegmentCount())
	}
	if got.ShowMarkers() {
		t.Error("ShowMarkers: want false")
	}
	if got.ShowGrid() {
		t.Error("ShowGrid: want false")
	}
}

// ── Deserialize: missing attributes use C# defaults ──────────────────────────

// TestZipCodeObject_Deserialize_MissingAttrsUseCSharpDefaults loads a minimal
// FRX snippet with no ZipCode attributes and verifies C# defaults are restored.
func TestZipCodeObject_Deserialize_MissingAttrsUseCSharpDefaults(t *testing.T) {
	xmlStr := `<ZipCodeObject name="z1"/>`

	r := serial.NewReader(strings.NewReader(xmlStr))
	_, ok := r.ReadObjectHeader()
	if !ok {
		t.Fatal("ReadObjectHeader failed")
	}
	z := object.NewZipCodeObject()
	if err := z.Deserialize(r); err != nil {
		t.Fatalf("Deserialize: %v", err)
	}

	if z.Text() != "123456" {
		t.Errorf("Text default after Deserialize = %q, want 123456", z.Text())
	}
	if z.SegmentWidth() != 18.9 {
		t.Errorf("SegmentWidth default after Deserialize = %v, want 18.9", z.SegmentWidth())
	}
	if z.SegmentHeight() != 37.8 {
		t.Errorf("SegmentHeight default after Deserialize = %v, want 37.8", z.SegmentHeight())
	}
	if z.Spacing() != 34.02 {
		t.Errorf("Spacing default after Deserialize = %v, want 34.02", z.Spacing())
	}
	if z.SegmentCount() != 6 {
		t.Errorf("SegmentCount default after Deserialize = %d, want 6", z.SegmentCount())
	}
	if !z.ShowMarkers() {
		t.Error("ShowMarkers default after Deserialize should be true")
	}
	if !z.ShowGrid() {
		t.Error("ShowGrid default after Deserialize should be true")
	}
}

// ── GetExpressions ────────────────────────────────────────────────────────────

// TestZipCodeObject_GetExpressions_Empty verifies an empty slice when neither
// DataColumn nor Expression is set.
func TestZipCodeObject_GetExpressions_Empty(t *testing.T) {
	z := object.NewZipCodeObject()
	exprs := z.GetExpressions()
	if len(exprs) != 0 {
		t.Errorf("GetExpressions (no binding): got %v, want empty", exprs)
	}
}

// TestZipCodeObject_GetExpressions_DataColumnOnly verifies only DataColumn
// is returned when only DataColumn is set.
func TestZipCodeObject_GetExpressions_DataColumnOnly(t *testing.T) {
	z := object.NewZipCodeObject()
	z.SetDataColumn("Orders.ZipCode")

	exprs := z.GetExpressions()
	if len(exprs) != 1 || exprs[0] != "Orders.ZipCode" {
		t.Errorf("GetExpressions (DataColumn only): got %v, want [Orders.ZipCode]", exprs)
	}
}

// TestZipCodeObject_GetExpressions_ExpressionOnly verifies only Expression
// is returned when only Expression is set.
func TestZipCodeObject_GetExpressions_ExpressionOnly(t *testing.T) {
	z := object.NewZipCodeObject()
	z.SetExpression("[Customer.PostalCode]")

	exprs := z.GetExpressions()
	if len(exprs) != 1 || exprs[0] != "[Customer.PostalCode]" {
		t.Errorf("GetExpressions (Expression only): got %v, want [[Customer.PostalCode]]", exprs)
	}
}

// TestZipCodeObject_GetExpressions_Both verifies both DataColumn and Expression
// are returned when both are set, mirroring C# ZipCodeObject.GetExpressions()
// (ZipCodeObject.cs:325-335) which adds both if non-empty.
func TestZipCodeObject_GetExpressions_Both(t *testing.T) {
	z := object.NewZipCodeObject()
	z.SetDataColumn("T.Zip")
	z.SetExpression("[Zip]")

	exprs := z.GetExpressions()
	if len(exprs) != 2 {
		t.Fatalf("GetExpressions (both): got %d expressions, want 2", len(exprs))
	}
	if exprs[0] != "T.Zip" {
		t.Errorf("GetExpressions[0]: got %q, want T.Zip", exprs[0])
	}
	if exprs[1] != "[Zip]" {
		t.Errorf("GetExpressions[1]: got %q, want [Zip]", exprs[1])
	}
}

// ── Serial registry ───────────────────────────────────────────────────────────

// TestZipCodeObject_SerialRegistry verifies "ZipCodeObject" is registered in the
// serial DefaultRegistry and creates a valid *ZipCodeObject.
func TestZipCodeObject_SerialRegistry(t *testing.T) {
	// Load the report package to trigger serial registration init.
	_ = reportpkg.NewReport()

	obj, err := serial.DefaultRegistry.Create("ZipCodeObject")
	if err != nil {
		t.Fatalf("DefaultRegistry.Create(ZipCodeObject): %v", err)
	}
	if obj == nil {
		t.Fatal("DefaultRegistry.Create(ZipCodeObject) returned nil")
	}
	z, ok := obj.(*object.ZipCodeObject)
	if !ok {
		t.Fatalf("DefaultRegistry.Create(ZipCodeObject) returned %T, want *object.ZipCodeObject", obj)
	}
	// Newly created via registry should have C# defaults.
	if z.Text() != "123456" {
		t.Errorf("registry-created ZipCodeObject Text = %q, want 123456", z.Text())
	}
	if z.SegmentCount() != 6 {
		t.Errorf("registry-created ZipCodeObject SegmentCount = %d, want 6", z.SegmentCount())
	}
}

// ── Text preserved when Serialize writes it ──────────────────────────────────

// TestZipCodeObject_Serialize_TextWrittenWhenNonDefault verifies that Text
// appears in the XML only when it differs from the C# default "123456".
func TestZipCodeObject_Serialize_TextWrittenWhenNonDefault(t *testing.T) {
	tests := []struct {
		text    string
		wantIn  bool
	}{
		{"123456", false}, // C# default — must NOT appear
		{"000000", true},  // non-default — must appear
		{"", true},        // empty string differs from "123456" — must appear
	}

	for _, tc := range tests {
		var buf bytes.Buffer
		w := serial.NewWriter(&buf)
		z := object.NewZipCodeObject()
		z.SetText(tc.text)
		if err := w.WriteObjectNamed("ZipCodeObject", z); err != nil {
			t.Fatalf("Serialize(%q): %v", tc.text, err)
		}
		_ = w.Flush()
		xml := buf.String()
		has := strings.Contains(xml, "Text=")
		if has != tc.wantIn {
			t.Errorf("Text=%q: wantIn=%v but strings.Contains(xml, Text=)=%v\nXML: %s",
				tc.text, tc.wantIn, has, xml)
		}
	}
}

// ── Segment dimension defaults not written ────────────────────────────────────

// TestZipCodeObject_Serialize_SegmentDimsNotWrittenAtDefaults verifies that
// the C# default segment dimensions (18.9 / 37.8 / 34.02) are NOT written,
// while changed values ARE written.
func TestZipCodeObject_Serialize_SegmentDimsNotWrittenAtDefaults(t *testing.T) {
	// All at defaults — nothing written.
	z := object.NewZipCodeObject()
	var buf bytes.Buffer
	w := serial.NewWriter(&buf)
	if err := w.WriteObjectNamed("ZipCodeObject", z); err != nil {
		t.Fatalf("Serialize: %v", err)
	}
	_ = w.Flush()
	xml := buf.String()
	for _, attr := range []string{"SegmentWidth=", "SegmentHeight=", "Spacing="} {
		if strings.Contains(xml, attr) {
			t.Errorf("default %q should not appear:\n%s", attr, xml)
		}
	}

	// Change all three — all should appear.
	z.SetSegmentWidth(9.0)
	z.SetSegmentHeight(18.0)
	z.SetSpacing(12.0)
	var buf2 bytes.Buffer
	w2 := serial.NewWriter(&buf2)
	if err := w2.WriteObjectNamed("ZipCodeObject", z); err != nil {
		t.Fatalf("Serialize (changed): %v", err)
	}
	_ = w2.Flush()
	xml2 := buf2.String()
	for _, attr := range []string{"SegmentWidth=", "SegmentHeight=", "Spacing="} {
		if !strings.Contains(xml2, attr) {
			t.Errorf("changed %q should appear:\n%s", attr, xml2)
		}
	}
}
