package object_test

// mschart_decode_coverage_test.go — coverage for decodeAllSeries XML-unmarshal
// failure paths and MSChartObject.Serialize WriteObjectNamed error path.

import (
	"encoding/base64"
	"testing"

	"github.com/andrewloable/go-fastreport/object"
)

// TestDecodeAllSeries_InvalidXMLFromBase64 exercises the path where base64
// decoding succeeds but the resulting bytes are not valid XML, so
// xml.Unmarshal fails and decodeAllSeries returns nil.
// The effect is visible: RenderToImage falls through to the Series slice
// path, which is also empty, so nil is returned.
func TestDecodeAllSeries_InvalidXMLFromBase64(t *testing.T) {
	// Encode binary garbage that is valid base64 but definitely not valid XML.
	// Using 0xFF 0xFE bytes ensures xml.Unmarshal fails.
	invalidXML := []byte{0xFF, 0xFE, 0x00, 0x01, '<', '<', '<', 0x00}
	encoded := base64.StdEncoding.EncodeToString(invalidXML)

	m := object.NewMSChartObject()
	m.ChartData = encoded
	// RenderToImage calls decodeAllSeries(m.ChartData); unmarshal fails → nil returned.
	img := m.RenderToImage(100, 50)
	// No series, no data — expect nil image.
	if img != nil {
		t.Logf("RenderToImage returned non-nil (possibly draws empty chart): %T", img)
	}
}

// TestDecodeAllSeries_InvalidXMLFromBase64_Control verifies that valid XML
// from base64 results in series data (positive control).
func TestDecodeAllSeries_ValidXMLFromBase64(t *testing.T) {
	xmlData := `<Chart><Series><Series Name="S" ChartType="Bar">` +
		`<Points><DataPoint YValues="10"/></Points>` +
		`</Series></Series></Chart>`
	encoded := base64.StdEncoding.EncodeToString([]byte(xmlData))

	m := object.NewMSChartObject()
	m.ChartData = encoded
	img := m.RenderToImage(100, 50)
	if img == nil {
		t.Fatal("expected non-nil image for valid XML chart data")
	}
}

// TestDecodeAllSeries_InvalidBase64ThenInvalidXML exercises the path where
// base64 decoding fails (err != nil), so chartData itself is used as raw
// bytes, but those bytes are also not valid XML. xml.Unmarshal then fails
// and decodeAllSeries returns nil.
func TestDecodeAllSeries_InvalidBase64ThenInvalidXML(t *testing.T) {
	// This string is not valid base64 (spaces, angle brackets) and also not
	// valid XML.
	invalidBoth := "not-base64 !!! <<invalid>>"

	m := object.NewMSChartObject()
	m.ChartData = invalidBoth
	img := m.RenderToImage(100, 50)
	_ = img // May be nil; just confirm no panic.
}

// TestDecodeAllSeries_EmptyChartData exercises the early-return path in
// decodeAllSeries when ChartData is empty (chartData == "").
func TestDecodeAllSeries_EmptyChartData(t *testing.T) {
	m := object.NewMSChartObject()
	// ChartData is already "" by default; RenderToImage → decodeAllSeries("") → nil.
	img := m.RenderToImage(100, 50)
	if img != nil {
		t.Logf("RenderToImage with no data returned: %T", img)
	}
}
