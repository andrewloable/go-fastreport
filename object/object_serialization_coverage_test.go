package object_test

// object_serialization_coverage_test.go — additional serialization coverage for:
//   - picture.go  : PictureObjectBase.Serialize/Deserialize, PictureObject.Serialize/Deserialize
//   - rich.go     : RichObject.Serialize/Deserialize
//   - svg.go      : SVGObject.Serialize/Deserialize
//   - sparkline.go: SparklineObject.Serialize/Deserialize
//   - rfid.go     : RFIDLabel.Serialize/Deserialize
//
// Each test exercises code paths that were previously uncovered, particularly:
//   - Serialize with all fields at their non-default values (branches taken)
//   - Deserialize round-trips confirming values survive write → read
//   - Default-value paths (branches skipped) to confirm return nil path

import (
	"bytes"
	"strings"
	"testing"

	"github.com/andrewloable/go-fastreport/object"
	"github.com/andrewloable/go-fastreport/serial"
)

// ═══════════════════════════════════════════════════════════════════════════
// PictureObjectBase — default-value path (all branches skipped)
// ═══════════════════════════════════════════════════════════════════════════

// TestPictureObjectBase_Serialize_DefaultValues verifies that when all
// PictureObjectBase fields are at defaults, no optional attributes are written.
func TestPictureObjectBase_Serialize_DefaultValues(t *testing.T) {
	orig := object.NewPictureObject()
	// SizeModeZoom is the default, so SizeMode branch is skipped.

	var buf bytes.Buffer
	w := serial.NewWriter(&buf)
	if err := w.WriteObjectNamed("PictureObject", orig); err != nil {
		t.Fatalf("Serialize: %v", err)
	}
	_ = w.Flush()

	xml := buf.String()
	for _, attr := range []string{
		`Angle=`, `DataColumn=`, `Grayscale=`, `ImageLocation=`,
		`ImageSourceExpression=`, `MaxHeight=`, `MaxWidth=`, `Padding=`,
		`SizeMode=`, `ImageAlign=`, `ShowErrorImage=`,
		`Tile=`, `Transparency=`,
	} {
		if strings.Contains(xml, attr) {
			t.Errorf("unexpected attribute %q in default XML:\n%s", attr, xml)
		}
	}

	r := serial.NewReader(bytes.NewReader(buf.Bytes()))
	_, ok := r.ReadObjectHeader()
	if !ok {
		t.Fatal("ReadObjectHeader failed")
	}
	got := object.NewPictureObject()
	if err := got.Deserialize(r); err != nil {
		t.Fatalf("Deserialize: %v", err)
	}
	// Verify defaults are preserved after round-trip.
	if got.Angle() != 0 {
		t.Errorf("Angle: got %d, want 0", got.Angle())
	}
	if got.Grayscale() {
		t.Error("Grayscale should be false")
	}
	if got.SizeMode() != object.SizeModeZoom {
		t.Errorf("SizeMode: got %d, want Zoom", got.SizeMode())
	}
	if got.Tile() {
		t.Error("Tile should be false")
	}
	if got.Transparency() != 0 {
		t.Errorf("Transparency: got %v, want 0", got.Transparency())
	}
}

// TestPictureObjectBase_Serialize_SizeModeNormal verifies the SizeMode branch
// is taken when SizeMode differs from the default (SizeModeZoom).
func TestPictureObjectBase_Serialize_SizeModeNormal(t *testing.T) {
	orig := object.NewPictureObject()
	orig.SetSizeMode(object.SizeModeNormal) // not SizeModeZoom → branch taken

	var buf bytes.Buffer
	w := serial.NewWriter(&buf)
	if err := w.WriteObjectNamed("PictureObject", orig); err != nil {
		t.Fatalf("Serialize: %v", err)
	}
	_ = w.Flush()

	xml := buf.String()
	if !strings.Contains(xml, `SizeMode=`) {
		t.Errorf("expected SizeMode attribute in XML:\n%s", xml)
	}

	r := serial.NewReader(bytes.NewReader(buf.Bytes()))
	_, ok := r.ReadObjectHeader()
	if !ok {
		t.Fatal("ReadObjectHeader failed")
	}
	got := object.NewPictureObject()
	if err := got.Deserialize(r); err != nil {
		t.Fatalf("Deserialize: %v", err)
	}
	if got.SizeMode() != object.SizeModeNormal {
		t.Errorf("SizeMode: got %d, want Normal", got.SizeMode())
	}
}

// TestPictureObjectBase_Serialize_AutoSize verifies the SizeMode=AutoSize branch.
func TestPictureObjectBase_Serialize_AutoSize(t *testing.T) {
	orig := object.NewPictureObject()
	orig.SetSizeMode(object.SizeModeAutoSize)

	var buf bytes.Buffer
	w := serial.NewWriter(&buf)
	if err := w.WriteObjectNamed("PictureObject", orig); err != nil {
		t.Fatalf("Serialize: %v", err)
	}
	_ = w.Flush()

	r := serial.NewReader(bytes.NewReader(buf.Bytes()))
	_, ok := r.ReadObjectHeader()
	if !ok {
		t.Fatal("ReadObjectHeader failed")
	}
	got := object.NewPictureObject()
	if err := got.Deserialize(r); err != nil {
		t.Fatalf("Deserialize: %v", err)
	}
	if got.SizeMode() != object.SizeModeAutoSize {
		t.Errorf("SizeMode: got %d, want AutoSize", got.SizeMode())
	}
}

// TestPictureObjectBase_Serialize_AllOptional tests each optional field branch
// in isolation to confirm the conditional write is taken.
func TestPictureObjectBase_Serialize_AllOptional(t *testing.T) {
	orig := object.NewPictureObject()
	orig.SetAngle(180)
	orig.SetDataColumn("col")
	orig.SetGrayscale(true)
	orig.SetImageLocation("http://host/img.jpg")
	orig.SetImageSourceExpression("[Img]")
	orig.SetMaxHeight(100)
	orig.SetMaxWidth(200)
	orig.SetPadding(object.Padding{Left: 1, Top: 2, Right: 3, Bottom: 4})
	orig.SetSizeMode(object.SizeModeStretchImage)
	orig.SetImageAlign(object.ImageAlignBottomRight)
	orig.SetShowErrorImage(true)
	orig.SetTile(true)
	orig.SetTransparency(0.25)

	var buf bytes.Buffer
	w := serial.NewWriter(&buf)
	if err := w.WriteObjectNamed("PictureObject", orig); err != nil {
		t.Fatalf("Serialize: %v", err)
	}
	_ = w.Flush()

	xml := buf.String()
	for _, attr := range []string{
		`Angle=`, `DataColumn=`, `Grayscale="true"`, `ImageLocation=`,
		`ImageSourceExpression=`, `MaxHeight=`, `MaxWidth=`, `Padding=`,
		`SizeMode=`, `ImageAlign=`, `ShowErrorImage="true"`,
		`Tile="true"`, `Transparency=`,
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
	got := object.NewPictureObject()
	if err := got.Deserialize(r); err != nil {
		t.Fatalf("Deserialize: %v", err)
	}

	if got.Angle() != 180 {
		t.Errorf("Angle: got %d, want 180", got.Angle())
	}
	if got.DataColumn() != "col" {
		t.Errorf("DataColumn: got %q", got.DataColumn())
	}
	if !got.Grayscale() {
		t.Error("Grayscale should be true")
	}
	if got.ImageLocation() != "http://host/img.jpg" {
		t.Errorf("ImageLocation: got %q", got.ImageLocation())
	}
	if got.ImageSourceExpression() != "[Img]" {
		t.Errorf("ImageSourceExpression: got %q", got.ImageSourceExpression())
	}
	if got.MaxHeight() != 100 {
		t.Errorf("MaxHeight: got %v", got.MaxHeight())
	}
	if got.MaxWidth() != 200 {
		t.Errorf("MaxWidth: got %v", got.MaxWidth())
	}
	if got.SizeMode() != object.SizeModeStretchImage {
		t.Errorf("SizeMode: got %d", got.SizeMode())
	}
	if got.ImageAlign() != object.ImageAlignBottomRight {
		t.Errorf("ImageAlign: got %d", got.ImageAlign())
	}
	if !got.ShowErrorImage() {
		t.Error("ShowErrorImage should be true")
	}
	if !got.Tile() {
		t.Error("Tile should be true")
	}
	if got.Transparency() != 0.25 {
		t.Errorf("Transparency: got %v, want 0.25", got.Transparency())
	}
}

// TestPictureObject_Serialize_DefaultValues verifies that PictureObject with
// tile=false and transparency=0 skips those branches.
func TestPictureObject_Serialize_DefaultValues(t *testing.T) {
	orig := object.NewPictureObject()
	// tile=false, transparency=0 → branches skipped

	var buf bytes.Buffer
	w := serial.NewWriter(&buf)
	if err := w.WriteObjectNamed("PictureObject", orig); err != nil {
		t.Fatalf("Serialize: %v", err)
	}
	_ = w.Flush()

	xml := buf.String()
	if strings.Contains(xml, `Tile=`) {
		t.Errorf("unexpected Tile in default XML:\n%s", xml)
	}
	if strings.Contains(xml, `Transparency=`) {
		t.Errorf("unexpected Transparency in default XML:\n%s", xml)
	}

	r := serial.NewReader(bytes.NewReader(buf.Bytes()))
	_, ok := r.ReadObjectHeader()
	if !ok {
		t.Fatal("ReadObjectHeader failed")
	}
	got := object.NewPictureObject()
	if err := got.Deserialize(r); err != nil {
		t.Fatalf("Deserialize: %v", err)
	}
	if got.Tile() {
		t.Error("Tile should be false after round-trip")
	}
	if got.Transparency() != 0 {
		t.Errorf("Transparency: got %v, want 0", got.Transparency())
	}
}

// TestPictureObject_Serialize_WithTileAndTransparency verifies the tile and
// transparency branches are taken when set to non-default values.
func TestPictureObject_Serialize_WithTileAndTransparency(t *testing.T) {
	orig := object.NewPictureObject()
	orig.SetTile(true)
	orig.SetTransparency(0.75)

	var buf bytes.Buffer
	w := serial.NewWriter(&buf)
	if err := w.WriteObjectNamed("PictureObject", orig); err != nil {
		t.Fatalf("Serialize: %v", err)
	}
	_ = w.Flush()

	xml := buf.String()
	if !strings.Contains(xml, `Tile="true"`) {
		t.Errorf("expected Tile=true in XML:\n%s", xml)
	}
	if !strings.Contains(xml, `Transparency=`) {
		t.Errorf("expected Transparency in XML:\n%s", xml)
	}

	r := serial.NewReader(bytes.NewReader(buf.Bytes()))
	_, ok := r.ReadObjectHeader()
	if !ok {
		t.Fatal("ReadObjectHeader failed")
	}
	got := object.NewPictureObject()
	if err := got.Deserialize(r); err != nil {
		t.Fatalf("Deserialize: %v", err)
	}
	if !got.Tile() {
		t.Error("Tile should be true")
	}
	if got.Transparency() != 0.75 {
		t.Errorf("Transparency: got %v, want 0.75", got.Transparency())
	}
}

// TestPictureObjectBase_Deserialize_WithPadding verifies that a Padding string
// is correctly parsed during deserialization.
func TestPictureObjectBase_Deserialize_WithPadding(t *testing.T) {
	orig := object.NewPictureObject()
	orig.SetPadding(object.Padding{Left: 10, Top: 20, Right: 30, Bottom: 40})

	var buf bytes.Buffer
	w := serial.NewWriter(&buf)
	if err := w.WriteObjectNamed("PictureObject", orig); err != nil {
		t.Fatalf("Serialize: %v", err)
	}
	_ = w.Flush()

	r := serial.NewReader(bytes.NewReader(buf.Bytes()))
	_, ok := r.ReadObjectHeader()
	if !ok {
		t.Fatal("ReadObjectHeader failed")
	}
	got := object.NewPictureObject()
	if err := got.Deserialize(r); err != nil {
		t.Fatalf("Deserialize: %v", err)
	}
	want := object.Padding{Left: 10, Top: 20, Right: 30, Bottom: 40}
	if got.Padding() != want {
		t.Errorf("Padding: got %+v, want %+v", got.Padding(), want)
	}
}

// ═══════════════════════════════════════════════════════════════════════════
// RichObject — additional coverage
// ═══════════════════════════════════════════════════════════════════════════

// TestRichObject_Serialize_OnlyText tests the path where only text is set (canGrow=false).
func TestRichObject_Serialize_OnlyText(t *testing.T) {
	orig := object.NewRichObject()
	orig.SetText("{\\rtf1 content}")
	// canGrow remains false → branch for canGrow is skipped

	var buf bytes.Buffer
	w := serial.NewWriter(&buf)
	if err := w.WriteObjectNamed("RichObject", orig); err != nil {
		t.Fatalf("Serialize: %v", err)
	}
	_ = w.Flush()

	xml := buf.String()
	if !strings.Contains(xml, `Text=`) {
		t.Errorf("expected Text in XML:\n%s", xml)
	}
	if strings.Contains(xml, `CanGrow=`) {
		t.Errorf("unexpected CanGrow when false:\n%s", xml)
	}

	r := serial.NewReader(bytes.NewReader(buf.Bytes()))
	_, ok := r.ReadObjectHeader()
	if !ok {
		t.Fatal("ReadObjectHeader failed")
	}
	got := object.NewRichObject()
	if err := got.Deserialize(r); err != nil {
		t.Fatalf("Deserialize: %v", err)
	}
	if got.Text() != "{\\rtf1 content}" {
		t.Errorf("Text: got %q", got.Text())
	}
	if got.CanGrow() {
		t.Error("CanGrow should be false")
	}
}

// TestRichObject_Serialize_OnlyCanGrow tests the path where text is empty but canGrow=true.
func TestRichObject_Serialize_OnlyCanGrow(t *testing.T) {
	orig := object.NewRichObject()
	orig.SetCanGrow(true)
	// text remains "" → text branch is skipped

	var buf bytes.Buffer
	w := serial.NewWriter(&buf)
	if err := w.WriteObjectNamed("RichObject", orig); err != nil {
		t.Fatalf("Serialize: %v", err)
	}
	_ = w.Flush()

	xml := buf.String()
	if strings.Contains(xml, `Text=`) {
		t.Errorf("unexpected Text when empty:\n%s", xml)
	}
	if !strings.Contains(xml, `CanGrow="true"`) {
		t.Errorf("expected CanGrow=true in XML:\n%s", xml)
	}

	r := serial.NewReader(bytes.NewReader(buf.Bytes()))
	_, ok := r.ReadObjectHeader()
	if !ok {
		t.Fatal("ReadObjectHeader failed")
	}
	got := object.NewRichObject()
	if err := got.Deserialize(r); err != nil {
		t.Fatalf("Deserialize: %v", err)
	}
	if got.Text() != "" {
		t.Errorf("Text: got %q, want empty", got.Text())
	}
	if !got.CanGrow() {
		t.Error("CanGrow should be true")
	}
}

// TestRichObject_Serialize_BothTextAndCanGrow tests the path where both fields are set.
func TestRichObject_Serialize_BothTextAndCanGrow(t *testing.T) {
	orig := object.NewRichObject()
	orig.SetText("{\\rtf1 hello}")
	orig.SetCanGrow(true)

	var buf bytes.Buffer
	w := serial.NewWriter(&buf)
	if err := w.WriteObjectNamed("RichObject", orig); err != nil {
		t.Fatalf("Serialize: %v", err)
	}
	_ = w.Flush()

	xml := buf.String()
	if !strings.Contains(xml, `Text=`) {
		t.Errorf("expected Text in XML:\n%s", xml)
	}
	if !strings.Contains(xml, `CanGrow="true"`) {
		t.Errorf("expected CanGrow=true in XML:\n%s", xml)
	}

	r := serial.NewReader(bytes.NewReader(buf.Bytes()))
	_, ok := r.ReadObjectHeader()
	if !ok {
		t.Fatal("ReadObjectHeader failed")
	}
	got := object.NewRichObject()
	if err := got.Deserialize(r); err != nil {
		t.Fatalf("Deserialize: %v", err)
	}
	if got.Text() != "{\\rtf1 hello}" {
		t.Errorf("Text: got %q", got.Text())
	}
	if !got.CanGrow() {
		t.Error("CanGrow should be true")
	}
}

// TestRichObject_Deserialize_EmptyXML verifies deserialization from an XML that
// has no Text or CanGrow attributes (both default).
func TestRichObject_Deserialize_EmptyXML(t *testing.T) {
	xml := `<RichObject Name="r1"/>`
	r := serial.NewReader(strings.NewReader(xml))
	_, ok := r.ReadObjectHeader()
	if !ok {
		t.Fatal("ReadObjectHeader failed")
	}
	got := object.NewRichObject()
	if err := got.Deserialize(r); err != nil {
		t.Fatalf("Deserialize: %v", err)
	}
	if got.Text() != "" {
		t.Errorf("Text: got %q, want empty", got.Text())
	}
	if got.CanGrow() {
		t.Error("CanGrow should be false")
	}
}

// ═══════════════════════════════════════════════════════════════════════════
// SVGObject — additional coverage
// ═══════════════════════════════════════════════════════════════════════════

// TestSVGObject_Serialize_WithSvgData verifies the SvgData branch is taken
// when SvgData is non-empty.
func TestSVGObject_Serialize_WithSvgData(t *testing.T) {
	orig := object.NewSVGObject()
	orig.SvgData = "PHN2ZyB4bWxucz0idXJuOmFiYyI+PC9zdmc+"

	var buf bytes.Buffer
	w := serial.NewWriter(&buf)
	if err := w.WriteObjectNamed("SVGObject", orig); err != nil {
		t.Fatalf("Serialize: %v", err)
	}
	_ = w.Flush()

	xml := buf.String()
	if !strings.Contains(xml, `SvgData=`) {
		t.Errorf("expected SvgData in XML:\n%s", xml)
	}

	r := serial.NewReader(bytes.NewReader(buf.Bytes()))
	_, ok := r.ReadObjectHeader()
	if !ok {
		t.Fatal("ReadObjectHeader failed")
	}
	got := object.NewSVGObject()
	if err := got.Deserialize(r); err != nil {
		t.Fatalf("Deserialize: %v", err)
	}
	if got.SvgData != "PHN2ZyB4bWxucz0idXJuOmFiYyI+PC9zdmc+" {
		t.Errorf("SvgData: got %q", got.SvgData)
	}
}

// TestSVGObject_Serialize_EmptySvgData verifies the SvgData branch is skipped
// when SvgData is empty.
func TestSVGObject_Serialize_EmptySvgData(t *testing.T) {
	orig := object.NewSVGObject()
	// SvgData="" → branch skipped

	var buf bytes.Buffer
	w := serial.NewWriter(&buf)
	if err := w.WriteObjectNamed("SVGObject", orig); err != nil {
		t.Fatalf("Serialize: %v", err)
	}
	_ = w.Flush()

	xml := buf.String()
	if strings.Contains(xml, `SvgData=`) {
		t.Errorf("unexpected SvgData in empty XML:\n%s", xml)
	}

	r := serial.NewReader(bytes.NewReader(buf.Bytes()))
	_, ok := r.ReadObjectHeader()
	if !ok {
		t.Fatal("ReadObjectHeader failed")
	}
	got := object.NewSVGObject()
	if err := got.Deserialize(r); err != nil {
		t.Fatalf("Deserialize: %v", err)
	}
	if got.SvgData != "" {
		t.Errorf("SvgData: got %q, want empty", got.SvgData)
	}
}

// TestSVGObject_Deserialize_FromRawXML verifies deserialization from raw XML
// with SvgData attribute present.
func TestSVGObject_Deserialize_FromRawXML(t *testing.T) {
	rawXML := `<SVGObject Name="svg1" SvgData="aHR0cHM6Ly9leGFtcGxlLmNvbQ=="/>`
	r := serial.NewReader(strings.NewReader(rawXML))
	_, ok := r.ReadObjectHeader()
	if !ok {
		t.Fatal("ReadObjectHeader failed")
	}
	got := object.NewSVGObject()
	if err := got.Deserialize(r); err != nil {
		t.Fatalf("Deserialize: %v", err)
	}
	if got.SvgData != "aHR0cHM6Ly9leGFtcGxlLmNvbQ==" {
		t.Errorf("SvgData: got %q", got.SvgData)
	}
	if got.Name() != "svg1" {
		t.Errorf("Name: got %q, want svg1", got.Name())
	}
}

// ═══════════════════════════════════════════════════════════════════════════
// SparklineObject — additional coverage
// ═══════════════════════════════════════════════════════════════════════════

// TestSparklineObject_Serialize_OnlyChartData verifies the ChartData branch is
// taken but Dock branch is skipped when Dock is empty.
func TestSparklineObject_Serialize_OnlyChartData(t *testing.T) {
	orig := object.NewSparklineObject()
	orig.ChartData = "Y2hhcnREYXRh"
	// Dock="" → Dock branch skipped

	var buf bytes.Buffer
	w := serial.NewWriter(&buf)
	if err := w.WriteObjectNamed("SparklineObject", orig); err != nil {
		t.Fatalf("Serialize: %v", err)
	}
	_ = w.Flush()

	xml := buf.String()
	if !strings.Contains(xml, `ChartData=`) {
		t.Errorf("expected ChartData in XML:\n%s", xml)
	}
	if strings.Contains(xml, `Dock=`) {
		t.Errorf("unexpected Dock when empty:\n%s", xml)
	}

	r := serial.NewReader(bytes.NewReader(buf.Bytes()))
	_, ok := r.ReadObjectHeader()
	if !ok {
		t.Fatal("ReadObjectHeader failed")
	}
	got := object.NewSparklineObject()
	if err := got.Deserialize(r); err != nil {
		t.Fatalf("Deserialize: %v", err)
	}
	if got.ChartData != "Y2hhcnREYXRh" {
		t.Errorf("ChartData: got %q", got.ChartData)
	}
	if got.Dock != "" {
		t.Errorf("Dock: got %q, want empty", got.Dock)
	}
}

// TestSparklineObject_Serialize_OnlyDock verifies the Dock branch is taken but
// ChartData branch is skipped when ChartData is empty.
func TestSparklineObject_Serialize_OnlyDock(t *testing.T) {
	orig := object.NewSparklineObject()
	orig.Dock = "Left"
	// ChartData="" → ChartData branch skipped

	var buf bytes.Buffer
	w := serial.NewWriter(&buf)
	if err := w.WriteObjectNamed("SparklineObject", orig); err != nil {
		t.Fatalf("Serialize: %v", err)
	}
	_ = w.Flush()

	xml := buf.String()
	if strings.Contains(xml, `ChartData=`) {
		t.Errorf("unexpected ChartData when empty:\n%s", xml)
	}
	if !strings.Contains(xml, `Dock=`) {
		t.Errorf("expected Dock in XML:\n%s", xml)
	}

	r := serial.NewReader(bytes.NewReader(buf.Bytes()))
	_, ok := r.ReadObjectHeader()
	if !ok {
		t.Fatal("ReadObjectHeader failed")
	}
	got := object.NewSparklineObject()
	if err := got.Deserialize(r); err != nil {
		t.Fatalf("Deserialize: %v", err)
	}
	if got.ChartData != "" {
		t.Errorf("ChartData: got %q, want empty", got.ChartData)
	}
	if got.Dock != "Left" {
		t.Errorf("Dock: got %q, want Left", got.Dock)
	}
}

// TestSparklineObject_Serialize_BothFields verifies both ChartData and Dock
// branches are taken together.
func TestSparklineObject_Serialize_BothFields(t *testing.T) {
	orig := object.NewSparklineObject()
	orig.ChartData = "ZGF0YQ=="
	orig.Dock = "Fill"

	var buf bytes.Buffer
	w := serial.NewWriter(&buf)
	if err := w.WriteObjectNamed("SparklineObject", orig); err != nil {
		t.Fatalf("Serialize: %v", err)
	}
	_ = w.Flush()

	xml := buf.String()
	if !strings.Contains(xml, `ChartData=`) {
		t.Errorf("expected ChartData in XML:\n%s", xml)
	}
	if !strings.Contains(xml, `Dock=`) {
		t.Errorf("expected Dock in XML:\n%s", xml)
	}

	r := serial.NewReader(bytes.NewReader(buf.Bytes()))
	_, ok := r.ReadObjectHeader()
	if !ok {
		t.Fatal("ReadObjectHeader failed")
	}
	got := object.NewSparklineObject()
	if err := got.Deserialize(r); err != nil {
		t.Fatalf("Deserialize: %v", err)
	}
	if got.ChartData != "ZGF0YQ==" {
		t.Errorf("ChartData: got %q", got.ChartData)
	}
	if got.Dock != "Fill" {
		t.Errorf("Dock: got %q, want Fill", got.Dock)
	}
}

// TestSparklineObject_Deserialize_FromRawXML verifies deserialization from raw
// XML covering both ChartData and Dock attributes.
func TestSparklineObject_Deserialize_FromRawXML(t *testing.T) {
	rawXML := `<SparklineObject Name="sp1" ChartData="abc123" Dock="Right"/>`
	r := serial.NewReader(strings.NewReader(rawXML))
	_, ok := r.ReadObjectHeader()
	if !ok {
		t.Fatal("ReadObjectHeader failed")
	}
	got := object.NewSparklineObject()
	if err := got.Deserialize(r); err != nil {
		t.Fatalf("Deserialize: %v", err)
	}
	if got.ChartData != "abc123" {
		t.Errorf("ChartData: got %q, want abc123", got.ChartData)
	}
	if got.Dock != "Right" {
		t.Errorf("Dock: got %q, want Right", got.Dock)
	}
}

// ═══════════════════════════════════════════════════════════════════════════
// RFIDLabel — additional coverage
// ═══════════════════════════════════════════════════════════════════════════

// TestRFIDLabel_Serialize_DefaultValues verifies that default-value fields are
// not written (all branches skipped).
func TestRFIDLabel_Serialize_DefaultValues(t *testing.T) {
	orig := object.NewRFIDLabel()
	// All banks empty, passwords empty, locks=0, bools=false, ErrorHandle=Skip(0)

	var buf bytes.Buffer
	w := serial.NewWriter(&buf)
	if err := w.WriteObjectNamed("RFIDLabel", orig); err != nil {
		t.Fatalf("Serialize: %v", err)
	}
	_ = w.Flush()

	xml := buf.String()
	for _, attr := range []string{
		`EPCBank.Data=`, `EPCBank.DataColumn=`, `EPCBank.Offset=`, `EPCBank.DataFormat=`,
		`TIDBank.Data=`, `UserBank.Data=`,
		`AccessPassword=`, `KillPassword=`,
		`LockKillPassword=`, `LockAccessPassword=`, `LockEPCBank=`, `LockUserBank=`,
		`UseAdjustForEPC=`, `RewriteEPCBank=`, `ErrorHandle=`,
	} {
		if strings.Contains(xml, attr) {
			t.Errorf("unexpected attribute %q in default XML:\n%s", attr, xml)
		}
	}

	r := serial.NewReader(bytes.NewReader(buf.Bytes()))
	_, ok := r.ReadObjectHeader()
	if !ok {
		t.Fatal("ReadObjectHeader failed")
	}
	got := object.NewRFIDLabel()
	if err := got.Deserialize(r); err != nil {
		t.Fatalf("Deserialize: %v", err)
	}
	if got.EPCBank.Data != "" {
		t.Errorf("EPCBank.Data: got %q, want empty", got.EPCBank.Data)
	}
	if got.UseAdjustForEPC {
		t.Error("UseAdjustForEPC should be false")
	}
	if got.RewriteEPCBank {
		t.Error("RewriteEPCBank should be false")
	}
}

// TestRFIDLabel_Serialize_EPCBankOnly tests serialization with only EPCBank set.
func TestRFIDLabel_Serialize_EPCBankOnly(t *testing.T) {
	orig := object.NewRFIDLabel()
	orig.EPCBank = object.RFIDBank{
		Data:       "DEADBEEF",
		DataColumn: "epc_col",
		Offset:     8,
		DataFormat: object.RFIDBankFormatASCII,
	}

	var buf bytes.Buffer
	w := serial.NewWriter(&buf)
	if err := w.WriteObjectNamed("RFIDLabel", orig); err != nil {
		t.Fatalf("Serialize: %v", err)
	}
	_ = w.Flush()

	xml := buf.String()
	for _, attr := range []string{
		`EPCBank.Data=`, `EPCBank.DataColumn=`, `EPCBank.Offset=`, `EPCBank.DataFormat=`,
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
	got := object.NewRFIDLabel()
	if err := got.Deserialize(r); err != nil {
		t.Fatalf("Deserialize: %v", err)
	}
	if got.EPCBank.Data != "DEADBEEF" {
		t.Errorf("EPCBank.Data: got %q", got.EPCBank.Data)
	}
	if got.EPCBank.DataColumn != "epc_col" {
		t.Errorf("EPCBank.DataColumn: got %q", got.EPCBank.DataColumn)
	}
	if got.EPCBank.Offset != 8 {
		t.Errorf("EPCBank.Offset: got %d", got.EPCBank.Offset)
	}
	if got.EPCBank.DataFormat != object.RFIDBankFormatASCII {
		t.Errorf("EPCBank.DataFormat: got %v", got.EPCBank.DataFormat)
	}
}

// TestRFIDLabel_Serialize_LockTypes exercises all non-zero lock type branches.
func TestRFIDLabel_Serialize_LockTypes(t *testing.T) {
	orig := object.NewRFIDLabel()
	orig.LockKillPassword = object.RFIDLockTypeLock
	orig.LockAccessPassword = object.RFIDLockTypePermanentUnlock
	orig.LockEPCBank = object.RFIDLockTypePermanentLock
	orig.LockUserBank = object.RFIDLockTypeLock

	var buf bytes.Buffer
	w := serial.NewWriter(&buf)
	if err := w.WriteObjectNamed("RFIDLabel", orig); err != nil {
		t.Fatalf("Serialize: %v", err)
	}
	_ = w.Flush()

	xml := buf.String()
	for _, attr := range []string{
		`LockKillPassword=`, `LockAccessPassword=`, `LockEPCBank=`, `LockUserBank=`,
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
	got := object.NewRFIDLabel()
	if err := got.Deserialize(r); err != nil {
		t.Fatalf("Deserialize: %v", err)
	}
	if got.LockKillPassword != object.RFIDLockTypeLock {
		t.Errorf("LockKillPassword: got %v", got.LockKillPassword)
	}
	if got.LockAccessPassword != object.RFIDLockTypePermanentUnlock {
		t.Errorf("LockAccessPassword: got %v", got.LockAccessPassword)
	}
	if got.LockEPCBank != object.RFIDLockTypePermanentLock {
		t.Errorf("LockEPCBank: got %v", got.LockEPCBank)
	}
	if got.LockUserBank != object.RFIDLockTypeLock {
		t.Errorf("LockUserBank: got %v", got.LockUserBank)
	}
}

// TestRFIDLabel_Serialize_Passwords exercises password and password data column branches.
func TestRFIDLabel_Serialize_Passwords(t *testing.T) {
	orig := object.NewRFIDLabel()
	orig.AccessPassword = "ACPASS"
	orig.AccessPasswordDataColumn = "acc_col"
	orig.KillPassword = "KLPASS"
	orig.KillPasswordDataColumn = "kill_col"

	var buf bytes.Buffer
	w := serial.NewWriter(&buf)
	if err := w.WriteObjectNamed("RFIDLabel", orig); err != nil {
		t.Fatalf("Serialize: %v", err)
	}
	_ = w.Flush()

	xml := buf.String()
	for _, attr := range []string{
		`AccessPassword=`, `AccessPasswordDataColumn=`,
		`KillPassword=`, `KillPasswordDataColumn=`,
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
	got := object.NewRFIDLabel()
	if err := got.Deserialize(r); err != nil {
		t.Fatalf("Deserialize: %v", err)
	}
	if got.AccessPassword != "ACPASS" {
		t.Errorf("AccessPassword: got %q", got.AccessPassword)
	}
	if got.AccessPasswordDataColumn != "acc_col" {
		t.Errorf("AccessPasswordDataColumn: got %q", got.AccessPasswordDataColumn)
	}
	if got.KillPassword != "KLPASS" {
		t.Errorf("KillPassword: got %q", got.KillPassword)
	}
	if got.KillPasswordDataColumn != "kill_col" {
		t.Errorf("KillPasswordDataColumn: got %q", got.KillPasswordDataColumn)
	}
}

// TestRFIDLabel_Serialize_BoolFlagsAndErrorHandle exercises the UseAdjustForEPC,
// RewriteEPCBank, and ErrorHandle branches.
func TestRFIDLabel_Serialize_BoolFlagsAndErrorHandle(t *testing.T) {
	orig := object.NewRFIDLabel()
	orig.UseAdjustForEPC = true
	orig.RewriteEPCBank = true
	orig.ErrorHandle = object.RFIDErrorHandlePause

	var buf bytes.Buffer
	w := serial.NewWriter(&buf)
	if err := w.WriteObjectNamed("RFIDLabel", orig); err != nil {
		t.Fatalf("Serialize: %v", err)
	}
	_ = w.Flush()

	xml := buf.String()
	if !strings.Contains(xml, `UseAdjustForEPC="true"`) {
		t.Errorf("expected UseAdjustForEPC=true in XML:\n%s", xml)
	}
	if !strings.Contains(xml, `RewriteEPCBank="true"`) {
		t.Errorf("expected RewriteEPCBank=true in XML:\n%s", xml)
	}
	if !strings.Contains(xml, `ErrorHandle=`) {
		t.Errorf("expected ErrorHandle in XML:\n%s", xml)
	}

	r := serial.NewReader(bytes.NewReader(buf.Bytes()))
	_, ok := r.ReadObjectHeader()
	if !ok {
		t.Fatal("ReadObjectHeader failed")
	}
	got := object.NewRFIDLabel()
	if err := got.Deserialize(r); err != nil {
		t.Fatalf("Deserialize: %v", err)
	}
	if !got.UseAdjustForEPC {
		t.Error("UseAdjustForEPC should be true")
	}
	if !got.RewriteEPCBank {
		t.Error("RewriteEPCBank should be true")
	}
	if got.ErrorHandle != object.RFIDErrorHandlePause {
		t.Errorf("ErrorHandle: got %v, want Pause", got.ErrorHandle)
	}
}

// TestRFIDLabel_Serialize_UserBankWithDecimalFormat exercises UserBank with
// Decimal format and TIDBank without any data.
func TestRFIDLabel_Serialize_UserBankDecimalFormat(t *testing.T) {
	orig := object.NewRFIDLabel()
	orig.UserBank = object.RFIDBank{
		Data:       "12345",
		DataFormat: object.RFIDBankFormatDecimal,
		Offset:     2,
	}

	var buf bytes.Buffer
	w := serial.NewWriter(&buf)
	if err := w.WriteObjectNamed("RFIDLabel", orig); err != nil {
		t.Fatalf("Serialize: %v", err)
	}
	_ = w.Flush()

	xml := buf.String()
	if !strings.Contains(xml, `UserBank.Data=`) {
		t.Errorf("expected UserBank.Data in XML:\n%s", xml)
	}
	if !strings.Contains(xml, `UserBank.DataFormat=`) {
		t.Errorf("expected UserBank.DataFormat in XML:\n%s", xml)
	}

	r := serial.NewReader(bytes.NewReader(buf.Bytes()))
	_, ok := r.ReadObjectHeader()
	if !ok {
		t.Fatal("ReadObjectHeader failed")
	}
	got := object.NewRFIDLabel()
	if err := got.Deserialize(r); err != nil {
		t.Fatalf("Deserialize: %v", err)
	}
	if got.UserBank.Data != "12345" {
		t.Errorf("UserBank.Data: got %q", got.UserBank.Data)
	}
	if got.UserBank.DataFormat != object.RFIDBankFormatDecimal {
		t.Errorf("UserBank.DataFormat: got %v", got.UserBank.DataFormat)
	}
	if got.UserBank.Offset != 2 {
		t.Errorf("UserBank.Offset: got %d", got.UserBank.Offset)
	}
}

// TestRFIDLabel_Deserialize_FromRawXML verifies deserialization from a raw XML
// string covering all field paths.
func TestRFIDLabel_Deserialize_FromRawXML(t *testing.T) {
	rawXML := `<RFIDLabel Name="rfid1" ` +
		`EPCBank.Data="EPC1" EPCBank.DataColumn="e_col" EPCBank.Offset="4" EPCBank.DataFormat="1" ` +
		`TIDBank.Data="TID1" ` +
		`UserBank.Data="USER" UserBank.DataColumn="u_col" ` +
		`AccessPassword="AP" KillPassword="KP" ` +
		`LockEPCBank="3" UseAdjustForEPC="true" RewriteEPCBank="true" ErrorHandle="2"/>`

	r := serial.NewReader(strings.NewReader(rawXML))
	_, ok := r.ReadObjectHeader()
	if !ok {
		t.Fatal("ReadObjectHeader failed")
	}
	got := object.NewRFIDLabel()
	if err := got.Deserialize(r); err != nil {
		t.Fatalf("Deserialize: %v", err)
	}

	if got.EPCBank.Data != "EPC1" {
		t.Errorf("EPCBank.Data: got %q", got.EPCBank.Data)
	}
	if got.EPCBank.DataColumn != "e_col" {
		t.Errorf("EPCBank.DataColumn: got %q", got.EPCBank.DataColumn)
	}
	if got.EPCBank.Offset != 4 {
		t.Errorf("EPCBank.Offset: got %d", got.EPCBank.Offset)
	}
	if got.EPCBank.DataFormat != object.RFIDBankFormatASCII {
		t.Errorf("EPCBank.DataFormat: got %v, want ASCII", got.EPCBank.DataFormat)
	}
	if got.TIDBank.Data != "TID1" {
		t.Errorf("TIDBank.Data: got %q", got.TIDBank.Data)
	}
	if got.UserBank.Data != "USER" {
		t.Errorf("UserBank.Data: got %q", got.UserBank.Data)
	}
	if got.UserBank.DataColumn != "u_col" {
		t.Errorf("UserBank.DataColumn: got %q", got.UserBank.DataColumn)
	}
	if got.AccessPassword != "AP" {
		t.Errorf("AccessPassword: got %q", got.AccessPassword)
	}
	if got.KillPassword != "KP" {
		t.Errorf("KillPassword: got %q", got.KillPassword)
	}
	if got.LockEPCBank != object.RFIDLockTypePermanentLock {
		t.Errorf("LockEPCBank: got %v", got.LockEPCBank)
	}
	if !got.UseAdjustForEPC {
		t.Error("UseAdjustForEPC should be true")
	}
	if !got.RewriteEPCBank {
		t.Error("RewriteEPCBank should be true")
	}
	if got.ErrorHandle != object.RFIDErrorHandleError {
		t.Errorf("ErrorHandle: got %v, want Error", got.ErrorHandle)
	}
}
