package object_test

// mschart_coverage_test.go — additional coverage for mschart.go:
// parseFloatMSC edge cases, msChartTypeStr area branches,
// MSChartObject/Series serialize with all fields, DeserializeChild unknown type.

import (
	"bytes"
	"encoding/base64"
	"image/color"
	"strings"
	"testing"

	"github.com/andrewloable/go-fastreport/object"
	"github.com/andrewloable/go-fastreport/serial"
)

// ── parseFloatMSC via decodeAllSeries ────────────────────────────────────────
// parseFloatMSC is package-private; we exercise it through RenderToImage
// (which calls decodeAllSeries → parseFloatMSC).

func TestParseFloatMSC_NegativeValue(t *testing.T) {
	xmlData := `<Chart><Series><Series Name="S" ChartType="Line">` +
		`<Points><DataPoint YValues="-42.5"/></Points>` +
		`</Series></Series></Chart>`
	encoded := base64.StdEncoding.EncodeToString([]byte(xmlData))

	m := object.NewMSChartObject()
	m.ChartData = encoded
	m.ChartType = "Line"
	img := m.RenderToImage(200, 100)
	if img == nil {
		t.Fatal("RenderToImage with negative value should return non-nil image")
	}
}

func TestParseFloatMSC_EmptyYValues(t *testing.T) {
	// Empty YValues → parseFloatMSC("") → 0
	xmlData := `<Chart><Series><Series Name="S" ChartType="Bar">` +
		`<Points><DataPoint YValues=""/><DataPoint YValues="10"/></Points>` +
		`</Series></Series></Chart>`
	encoded := base64.StdEncoding.EncodeToString([]byte(xmlData))

	m := object.NewMSChartObject()
	m.ChartData = encoded
	img := m.RenderToImage(200, 100)
	// May be nil or non-nil depending on rendering, but should not panic.
	_ = img
}

func TestParseFloatMSC_IntegerOnly(t *testing.T) {
	xmlData := `<Chart><Series><Series Name="S" ChartType="Bar">` +
		`<Points><DataPoint YValues="123"/></Points>` +
		`</Series></Series></Chart>`
	encoded := base64.StdEncoding.EncodeToString([]byte(xmlData))

	m := object.NewMSChartObject()
	m.ChartData = encoded
	img := m.RenderToImage(200, 100)
	_ = img
}

func TestParseFloatMSC_FractionalOnly(t *testing.T) {
	// "0.75" — integer part is 0, fractional part is .75
	xmlData := `<Chart><Series><Series Name="S" ChartType="Line">` +
		`<Points><DataPoint YValues="0.75"/><DataPoint YValues="1.25"/></Points>` +
		`</Series></Series></Chart>`
	encoded := base64.StdEncoding.EncodeToString([]byte(xmlData))

	m := object.NewMSChartObject()
	m.ChartData = encoded
	img := m.RenderToImage(200, 100)
	if img == nil {
		t.Fatal("fractional values should render successfully")
	}
}

func TestParseFloatMSC_NegativeFractional(t *testing.T) {
	xmlData := `<Chart><Series><Series Name="S" ChartType="Line">` +
		`<Points><DataPoint YValues="-3.14"/><DataPoint YValues="2.72"/></Points>` +
		`</Series></Series></Chart>`
	encoded := base64.StdEncoding.EncodeToString([]byte(xmlData))

	m := object.NewMSChartObject()
	m.ChartData = encoded
	img := m.RenderToImage(200, 100)
	if img == nil {
		t.Fatal("negative fractional values should render successfully")
	}
}

func TestParseFloatMSC_NonDigitChars(t *testing.T) {
	// Non-digit chars in integer and fractional parts are skipped.
	xmlData := `<Chart><Series><Series Name="S" ChartType="Bar">` +
		`<Points><DataPoint YValues="  5 "/><DataPoint YValues="10"/></Points>` +
		`</Series></Series></Chart>`
	encoded := base64.StdEncoding.EncodeToString([]byte(xmlData))

	m := object.NewMSChartObject()
	m.ChartData = encoded
	img := m.RenderToImage(200, 100)
	_ = img
}

// ── msChartTypeStr: area branches ────────────────────────────────────────────

func TestMSChartTypeStr_AreaAndStackedArea(t *testing.T) {
	// Exercise "area" and "stackedarea" branches via RenderToImage.
	for _, ct := range []string{"Area", "StackedArea"} {
		xmlData := `<Chart><Series><Series Name="S" ChartType="` + ct + `">` +
			`<Points><DataPoint YValues="10"/><DataPoint YValues="20"/></Points>` +
			`</Series></Series></Chart>`
		encoded := base64.StdEncoding.EncodeToString([]byte(xmlData))

		m := object.NewMSChartObject()
		m.ChartData = encoded
		m.ChartType = ct
		img := m.RenderToImage(200, 100)
		if img == nil {
			t.Errorf("ChartType=%q: RenderToImage returned nil", ct)
		}
	}
}

// ── MSChartSeries.Serialize: all fields ──────────────────────────────────────

func TestMSChartSeries_Serialize_AllFields(t *testing.T) {
	orig := object.NewMSChartSeries()
	orig.ChartType = "Bar"
	orig.ValuesSource = "[Revenue]"
	orig.ArgumentSource = "[Quarter]"
	orig.LegendText = "Revenue"
	orig.Color = color.RGBA{R: 0, G: 128, B: 255, A: 255}

	var buf bytes.Buffer
	w := serial.NewWriter(&buf)
	if err := w.WriteObjectNamed("MSChartSeries", orig); err != nil {
		t.Fatalf("Serialize: %v", err)
	}
	w.Flush() //nolint:errcheck

	xml := buf.String()
	for _, want := range []string{"ChartType=", "ValuesSource=", "ArgumentSource=", "LegendText=", "Color="} {
		if !strings.Contains(xml, want) {
			t.Errorf("expected %q in XML:\n%s", want, xml)
		}
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
	if got.ValuesSource != "[Revenue]" {
		t.Errorf("ValuesSource: got %q", got.ValuesSource)
	}
	if got.ArgumentSource != "[Quarter]" {
		t.Errorf("ArgumentSource: got %q", got.ArgumentSource)
	}
	if got.LegendText != "Revenue" {
		t.Errorf("LegendText: got %q", got.LegendText)
	}
	if got.Color.G != 128 {
		t.Errorf("Color.G: got %d, want 128", got.Color.G)
	}
}

// ── MSChartObject.Serialize: with Series children ────────────────────────────

func TestMSChartObject_Serialize_WithSeries(t *testing.T) {
	orig := object.NewMSChartObject()
	orig.ChartType = "Bar"
	orig.DataSource = "myDS"
	orig.ChartData = "abc=="

	s1 := object.NewMSChartSeries()
	s1.ChartType = "Bar"
	s1.LegendText = "Foo"
	orig.Series = append(orig.Series, s1)

	s2 := object.NewMSChartSeries()
	s2.ChartType = "Line"
	s2.ValuesSource = "[Val]"
	orig.Series = append(orig.Series, s2)

	var buf bytes.Buffer
	w := serial.NewWriter(&buf)
	if err := w.WriteObjectNamed("MSChartObject", orig); err != nil {
		t.Fatalf("Serialize: %v", err)
	}
	w.Flush() //nolint:errcheck

	xml := buf.String()
	if !strings.Contains(xml, "MSChartSeries") {
		t.Errorf("expected MSChartSeries elements in XML:\n%s", xml)
	}
	if !strings.Contains(xml, "LegendText=") {
		t.Errorf("expected LegendText in XML:\n%s", xml)
	}
}

// ── MSChartObject.DeserializeChild: unknown type ─────────────────────────────

func TestMSChartObject_DeserializeChild_UnknownType(t *testing.T) {
	// Build XML with an unknown child type — DeserializeChild should return false.
	xmlStr := `<MSChartObject ChartType="Bar"><UnknownChild Foo="bar"/></MSChartObject>`

	r := serial.NewReader(strings.NewReader(xmlStr))
	typeName, ok := r.ReadObjectHeader()
	if !ok {
		t.Fatal("ReadObjectHeader failed")
	}
	if typeName != "MSChartObject" {
		t.Fatalf("expected MSChartObject, got %q", typeName)
	}

	m := object.NewMSChartObject()
	if err := m.Deserialize(r); err != nil {
		t.Fatalf("Deserialize: %v", err)
	}

	// Drain children manually — DeserializeChild("UnknownChild") → false
	for {
		childType, ok := r.NextChild()
		if !ok {
			break
		}
		handled := m.DeserializeChild(childType, r)
		if handled {
			t.Errorf("DeserializeChild(%q) should return false", childType)
		}
		r.FinishChild() //nolint:errcheck
	}

	if len(m.Series) != 0 {
		t.Errorf("Series should be empty, got %d", len(m.Series))
	}
}

// ── MSChartObject.DeserializeChild: MSChartSeries round-trip ─────────────────

func TestMSChartObject_DeserializeChild_MSChartSeries(t *testing.T) {
	orig := object.NewMSChartObject()
	orig.ChartType = "Pie"

	s := object.NewMSChartSeries()
	s.LegendText = "Slice"
	s.ValuesSource = "[Amount]"
	orig.Series = append(orig.Series, s)

	var buf bytes.Buffer
	w := serial.NewWriter(&buf)
	if err := w.WriteObjectNamed("MSChartObject", orig); err != nil {
		t.Fatalf("Serialize: %v", err)
	}
	w.Flush() //nolint:errcheck

	r := serial.NewReader(bytes.NewReader(buf.Bytes()))
	_, ok := r.ReadObjectHeader()
	if !ok {
		t.Fatal("ReadObjectHeader failed")
	}
	got := object.NewMSChartObject()
	if err := got.Deserialize(r); err != nil {
		t.Fatalf("Deserialize: %v", err)
	}
	for {
		childType, ok := r.NextChild()
		if !ok {
			break
		}
		got.DeserializeChild(childType, r)
		r.FinishChild() //nolint:errcheck
	}

	if len(got.Series) != 1 {
		t.Fatalf("Series count: got %d, want 1", len(got.Series))
	}
	if got.Series[0].LegendText != "Slice" {
		t.Errorf("LegendText: got %q", got.Series[0].LegendText)
	}
}

// ── RenderToImage: Series color from MSChartSeries ───────────────────────────

func TestMSChartObject_RenderToImage_SeriesColorOverride(t *testing.T) {
	xmlData := `<Chart><Series><Series Name="S1" ChartType="Bar">` +
		`<Points><DataPoint YValues="10"/><DataPoint YValues="20"/></Points>` +
		`</Series></Series></Chart>`
	encoded := base64.StdEncoding.EncodeToString([]byte(xmlData))

	m := object.NewMSChartObject()
	m.ChartData = encoded
	m.ChartType = "Bar"

	// Add a series with an explicit color — this exercises the i < len(m.Series) branch
	s := object.NewMSChartSeries()
	s.Color = color.RGBA{R: 200, G: 50, B: 50, A: 255}
	m.Series = append(m.Series, s)

	img := m.RenderToImage(300, 200)
	if img == nil {
		t.Fatal("RenderToImage with color override should return non-nil")
	}
}

// ── MSChartSeries.Deserialize: invalid color string ──────────────────────────

func TestMSChartSeries_Deserialize_InvalidColor(t *testing.T) {
	// "Color" present but not a valid color string — ParseColor returns error,
	// so Color stays zero. Branch: if c, err := utils.ParseColor(cs); err == nil
	xmlStr := `<MSChartSeries Color="notacolor" LegendText="test"/>`

	r := serial.NewReader(strings.NewReader(xmlStr))
	_, ok := r.ReadObjectHeader()
	if !ok {
		t.Fatal("ReadObjectHeader failed")
	}
	s := object.NewMSChartSeries()
	if err := s.Deserialize(r); err != nil {
		t.Fatalf("Deserialize: %v", err)
	}
	// Color should remain zero (invalid color not applied)
	if s.Color != (color.RGBA{}) {
		t.Errorf("Color should be zero for invalid input, got %v", s.Color)
	}
	if s.LegendText != "test" {
		t.Errorf("LegendText: got %q, want test", s.LegendText)
	}
}
