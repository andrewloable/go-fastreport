package object_test

// object_misc_coverage_test.go — coverage tests for barcode, lines, map,
// mschart, picture, rich, sparkline, and cellular_text objects.

import (
	"bytes"
	"image/color"
	"strings"
	"testing"

	"github.com/andrewloable/go-fastreport/object"
	"github.com/andrewloable/go-fastreport/serial"
)

// ── RichObject ────────────────────────────────────────────────────────────────

func TestRichObject_SerializeDeserialize_WithText(t *testing.T) {
	orig := object.NewRichObject()
	orig.SetText("{\\rtf1 Hello}")
	orig.SetCanGrow(true)

	var buf bytes.Buffer
	w := serial.NewWriter(&buf)
	if err := w.WriteObjectNamed("RichObject", orig); err != nil {
		t.Fatalf("Serialize: %v", err)
	}
	w.Flush() //nolint:errcheck

	xml := buf.String()
	if !strings.Contains(xml, `Text=`) {
		t.Errorf("expected Text attribute in XML:\n%s", xml)
	}
	if !strings.Contains(xml, `CanGrow="true"`) {
		t.Errorf("expected CanGrow in XML:\n%s", xml)
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
	if got.Text() != "{\\rtf1 Hello}" {
		t.Errorf("Text: got %q, want {\\rtf1 Hello}", got.Text())
	}
	if !got.CanGrow() {
		t.Error("CanGrow should be true")
	}
}

// ── SparklineObject ───────────────────────────────────────────────────────────

func TestSparklineObject_SerializeDeserialize_WithData(t *testing.T) {
	orig := object.NewSparklineObject()
	orig.ChartData = "base64chartdata=="
	orig.Dock = "Fill"

	var buf bytes.Buffer
	w := serial.NewWriter(&buf)
	if err := w.WriteObjectNamed("SparklineObject", orig); err != nil {
		t.Fatalf("Serialize: %v", err)
	}
	w.Flush() //nolint:errcheck

	xml := buf.String()
	if !strings.Contains(xml, `ChartData=`) {
		t.Errorf("expected ChartData attribute in XML:\n%s", xml)
	}
	if !strings.Contains(xml, `Dock=`) {
		t.Errorf("expected Dock attribute in XML:\n%s", xml)
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
	if got.ChartData != "base64chartdata==" {
		t.Errorf("ChartData: got %q", got.ChartData)
	}
	if got.Dock != "Fill" {
		t.Errorf("Dock: got %q", got.Dock)
	}
}

// ── LineObject: diagonal ──────────────────────────────────────────────────────

func TestLineObject_SerializeDeserialize_Diagonal(t *testing.T) {
	orig := object.NewLineObject()
	orig.SetDiagonal(true)

	var buf bytes.Buffer
	w := serial.NewWriter(&buf)
	if err := w.WriteObjectNamed("LineObject", orig); err != nil {
		t.Fatalf("Serialize: %v", err)
	}
	w.Flush() //nolint:errcheck

	xml := buf.String()
	if !strings.Contains(xml, `Diagonal="true"`) {
		t.Errorf("expected Diagonal in XML:\n%s", xml)
	}

	r := serial.NewReader(bytes.NewReader(buf.Bytes()))
	_, ok := r.ReadObjectHeader()
	if !ok {
		t.Fatal("ReadObjectHeader failed")
	}
	got := object.NewLineObject()
	if err := got.Deserialize(r); err != nil {
		t.Fatalf("Deserialize: %v", err)
	}
	if !got.Diagonal() {
		t.Error("Diagonal should be true after round-trip")
	}
}

// ── ShapeObject: non-default ShapeType ───────────────────────────────────────

func TestShapeObject_SerializeDeserialize_NonDefaultShape(t *testing.T) {
	orig := object.NewShapeObject()
	orig.SetShape(object.ShapeKindEllipse)
	orig.SetCurve(5.0)

	var buf bytes.Buffer
	w := serial.NewWriter(&buf)
	if err := w.WriteObjectNamed("ShapeObject", orig); err != nil {
		t.Fatalf("Serialize: %v", err)
	}
	w.Flush() //nolint:errcheck

	xml := buf.String()
	if !strings.Contains(xml, `Shape=`) {
		t.Errorf("expected Shape attribute in XML:\n%s", xml)
	}
	if !strings.Contains(xml, `Curve=`) {
		t.Errorf("expected Curve attribute in XML:\n%s", xml)
	}

	r := serial.NewReader(bytes.NewReader(buf.Bytes()))
	_, ok := r.ReadObjectHeader()
	if !ok {
		t.Fatal("ReadObjectHeader failed")
	}
	got := object.NewShapeObject()
	if err := got.Deserialize(r); err != nil {
		t.Fatalf("Deserialize: %v", err)
	}
	if got.Shape() != object.ShapeKindEllipse {
		t.Errorf("Shape: got %d, want ShapeKindEllipse", got.Shape())
	}
	if got.Curve() != 5.0 {
		t.Errorf("Curve: got %v, want 5.0", got.Curve())
	}
}

// ── PictureObject: tile + transparency ───────────────────────────────────────

func TestPictureObject_SerializeDeserialize_TileTransparency(t *testing.T) {
	orig := object.NewPictureObject()
	orig.SetTile(true)
	orig.SetTransparency(0.5)

	var buf bytes.Buffer
	w := serial.NewWriter(&buf)
	if err := w.WriteObjectNamed("PictureObject", orig); err != nil {
		t.Fatalf("Serialize: %v", err)
	}
	w.Flush() //nolint:errcheck

	xml := buf.String()
	if !strings.Contains(xml, `Tile="true"`) {
		t.Errorf("expected Tile in XML:\n%s", xml)
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
		t.Error("Tile should be true after round-trip")
	}
	if got.Transparency() != 0.5 {
		t.Errorf("Transparency: got %v, want 0.5", got.Transparency())
	}
}

// ── MapObject: OffsetX/OffsetY ───────────────────────────────────────────────

func TestMapObject_SerializeDeserialize_Offsets(t *testing.T) {
	orig := object.NewMapObject()
	orig.OffsetX = 10.5
	orig.OffsetY = 20.5

	var buf bytes.Buffer
	w := serial.NewWriter(&buf)
	if err := w.WriteObjectNamed("MapObject", orig); err != nil {
		t.Fatalf("Serialize: %v", err)
	}
	w.Flush() //nolint:errcheck

	xml := buf.String()
	if !strings.Contains(xml, `OffsetX=`) {
		t.Errorf("expected OffsetX in XML:\n%s", xml)
	}

	r := serial.NewReader(bytes.NewReader(buf.Bytes()))
	_, ok := r.ReadObjectHeader()
	if !ok {
		t.Fatal("ReadObjectHeader failed")
	}
	got := object.NewMapObject()
	if err := got.Deserialize(r); err != nil {
		t.Fatalf("Deserialize: %v", err)
	}
	if got.OffsetX != 10.5 {
		t.Errorf("OffsetX: got %v, want 10.5", got.OffsetX)
	}
	if got.OffsetY != 20.5 {
		t.Errorf("OffsetY: got %v, want 20.5", got.OffsetY)
	}
}

// ── MSChartObject: ChartType + DataSource ────────────────────────────────────

func TestMSChartObject_SerializeDeserialize_NonDefaults(t *testing.T) {
	orig := object.NewMSChartObject()
	orig.ChartType = "Bar"
	orig.DataSource = "DS1"
	orig.ChartData = "somedata"

	var buf bytes.Buffer
	w := serial.NewWriter(&buf)
	if err := w.WriteObjectNamed("MSChartObject", orig); err != nil {
		t.Fatalf("Serialize: %v", err)
	}
	w.Flush() //nolint:errcheck

	xml := buf.String()
	if !strings.Contains(xml, `ChartType=`) {
		t.Errorf("expected ChartType in XML:\n%s", xml)
	}

	r := serial.NewReader(bytes.NewReader(buf.Bytes()))
	_, ok := r.ReadObjectHeader()
	if !ok {
		t.Fatal("ReadObjectHeader failed")
	}
	got := object.NewMSChartObject()
	if err := got.Deserialize(r); err != nil {
		t.Fatalf("Deserialize: %v", err)
	}
	if got.ChartType != "Bar" {
		t.Errorf("ChartType: got %q, want Bar", got.ChartType)
	}
	if got.DataSource != "DS1" {
		t.Errorf("DataSource: got %q, want DS1", got.DataSource)
	}
}

// ── MSChartSeries: ValuesSource + Color ──────────────────────────────────────

func TestMSChartSeries_SerializeDeserialize_ValuesColor(t *testing.T) {
	orig := object.NewMSChartSeries()
	orig.ChartType = "Line"
	orig.ValuesSource = "[Sales]"
	orig.ArgumentSource = "[Month]"
	orig.LegendText = "Sales"
	orig.Color = color.RGBA{R: 255, G: 0, B: 0, A: 255}

	var buf bytes.Buffer
	w := serial.NewWriter(&buf)
	if err := w.WriteObjectNamed("MSChartSeries", orig); err != nil {
		t.Fatalf("Serialize: %v", err)
	}
	w.Flush() //nolint:errcheck

	xml := buf.String()
	if !strings.Contains(xml, `ValuesSource=`) {
		t.Errorf("expected ValuesSource in XML:\n%s", xml)
	}
	if !strings.Contains(xml, `Color=`) {
		t.Errorf("expected Color in XML:\n%s", xml)
	}

	r := serial.NewReader(bytes.NewReader(buf.Bytes()))
	_, ok := r.ReadObjectHeader()
	if !ok {
		t.Fatal("ReadObjectHeader failed")
	}
	got := object.NewMSChartSeries()
	if err := got.Deserialize(r); err != nil {
		t.Fatalf("Deserialize: %v", err)
	}
	if got.ValuesSource != "[Sales]" {
		t.Errorf("ValuesSource: got %q", got.ValuesSource)
	}
	if got.Color.R != 255 {
		t.Errorf("Color.R: got %d, want 255", got.Color.R)
	}
}

// ── MSChartTypeStr: all chart type constants ──────────────────────────────────

func TestMSChartTypeStr_AllTypes(t *testing.T) {
	// Exercise msChartTypeStr indirectly through MSChartObject serialization
	// by setting ChartType to each known type and verifying round-trip.
	types := []string{"Bar", "Column", "StackedBar", "StackedColumn",
		"Area", "StackedArea", "Pie", "Doughnut", "Line", "unknown"}
	for _, ct := range types {
		orig := object.NewMSChartObject()
		orig.ChartType = ct

		var buf bytes.Buffer
		w := serial.NewWriter(&buf)
		w.WriteObjectNamed("MSChartObject", orig) //nolint:errcheck
		w.Flush()                                 //nolint:errcheck

		r := serial.NewReader(bytes.NewReader(buf.Bytes()))
		r.ReadObjectHeader()
		got := object.NewMSChartObject()
		got.Deserialize(r) //nolint:errcheck

		if got.ChartType != ct {
			t.Errorf("ChartType %q: round-trip got %q", ct, got.ChartType)
		}
	}
}

// ── CellularTextObject: non-zero spacing ─────────────────────────────────────

func TestCellularTextObject_SerializeDeserialize_Spacing(t *testing.T) {
	orig := object.NewCellularTextObject()
	orig.SetCellWidth(10)
	orig.SetCellHeight(20)
	orig.SetHorzSpacing(5)
	orig.SetVertSpacing(3)
	orig.SetWordWrap(false)

	var buf bytes.Buffer
	w := serial.NewWriter(&buf)
	if err := w.WriteObjectNamed("CellularTextObject", orig); err != nil {
		t.Fatalf("Serialize: %v", err)
	}
	w.Flush() //nolint:errcheck

	xml := buf.String()
	if !strings.Contains(xml, `CellWidth=`) {
		t.Errorf("expected CellWidth in XML:\n%s", xml)
	}
	if !strings.Contains(xml, `WordWrap="false"`) {
		t.Errorf("expected WordWrap=false in XML:\n%s", xml)
	}

	r := serial.NewReader(bytes.NewReader(buf.Bytes()))
	_, ok := r.ReadObjectHeader()
	if !ok {
		t.Fatal("ReadObjectHeader failed")
	}
	got := object.NewCellularTextObject()
	if err := got.Deserialize(r); err != nil {
		t.Fatalf("Deserialize: %v", err)
	}
	if got.CellWidth() != 10 {
		t.Errorf("CellWidth: got %v, want 10", got.CellWidth())
	}
	if got.CellHeight() != 20 {
		t.Errorf("CellHeight: got %v, want 20", got.CellHeight())
	}
	if got.HorzSpacing() != 5 {
		t.Errorf("HorzSpacing: got %v, want 5", got.HorzSpacing())
	}
	if got.VertSpacing() != 3 {
		t.Errorf("VertSpacing: got %v, want 3", got.VertSpacing())
	}
	if got.WordWrap() {
		t.Error("WordWrap should be false after round-trip")
	}
}

// ── BarcodeObject: TextPosition above ────────────────────────────────────────

func TestBarcodeObject_SerializeDeserialize_ShowTextFalse(t *testing.T) {
	orig := object.NewBarcodeObject()
	orig.SetShowText(false)
	orig.SetAutoSize(false)
	orig.SetText("12345")
	orig.SetBarcodeType("EAN-13")
	orig.SetAllowExpressions(true)

	var buf bytes.Buffer
	w := serial.NewWriter(&buf)
	if err := w.WriteObjectNamed("BarcodeObject", orig); err != nil {
		t.Fatalf("Serialize: %v", err)
	}
	w.Flush() //nolint:errcheck

	xml := buf.String()
	if !strings.Contains(xml, `ShowText="false"`) {
		t.Errorf("expected ShowText=false in XML:\n%s", xml)
	}
	if !strings.Contains(xml, `AutoSize="false"`) {
		t.Errorf("expected AutoSize=false in XML:\n%s", xml)
	}

	r := serial.NewReader(bytes.NewReader(buf.Bytes()))
	_, ok := r.ReadObjectHeader()
	if !ok {
		t.Fatal("ReadObjectHeader failed")
	}
	got := object.NewBarcodeObject()
	if err := got.Deserialize(r); err != nil {
		t.Fatalf("Deserialize: %v", err)
	}
	if got.ShowText() {
		t.Error("ShowText should be false")
	}
	if got.AllowExpressions() != true {
		t.Error("AllowExpressions should be true")
	}
}
