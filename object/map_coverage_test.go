package object_test

// map_coverage_test.go — additional coverage for map.go:
// MapLayer Serialize/Deserialize with all fields,
// MapObject Serialize/Deserialize zero-offset (no-write) branch,
// MapObject DeserializeChild with unknown type.

import (
	"bytes"
	"strings"
	"testing"

	"github.com/andrewloable/go-fastreport/object"
	"github.com/andrewloable/go-fastreport/serial"
)

// ── MapLayer: Serialize with all fields ──────────────────────────────────────

func TestMapLayer_Serialize_AllFields(t *testing.T) {
	orig := object.NewMapLayer()
	orig.Shapefile = "world.shp"
	orig.Type = "Choropleth"
	orig.DataSource = "GeoDS"
	orig.Filter = "[Country] = 'US'"
	orig.SpatialColumn = "ISO"
	orig.SpatialValue = "[ISO]"
	orig.AnalyticalValue = "[GDP]"
	orig.LabelColumn = "Name"
	orig.BoxAsString = "0,0,100,100"
	orig.Palette = "Blues"

	var buf bytes.Buffer
	w := serial.NewWriter(&buf)
	if err := w.WriteObjectNamed("MapLayer", orig); err != nil {
		t.Fatalf("Serialize: %v", err)
	}
	w.Flush() //nolint:errcheck

	xml := buf.String()
	for _, want := range []string{
		"Shapefile=", "Type=", "DataSource=", "Filter=",
		"SpatialColumn=", "SpatialValue=", "AnalyticalValue=",
		"LabelColumn=", "BoxAsString=", "Palette=",
	} {
		if !strings.Contains(xml, want) {
			t.Errorf("expected %q in XML:\n%s", want, xml)
		}
	}

	r := serial.NewReader(bytes.NewReader(buf.Bytes()))
	_, ok := r.ReadObjectHeader()
	if !ok {
		t.Fatal("ReadObjectHeader failed")
	}
	got := object.NewMapLayer()
	if err := got.Deserialize(r); err != nil {
		t.Fatalf("Deserialize: %v", err)
	}
	if got.Shapefile != "world.shp" {
		t.Errorf("Shapefile: got %q", got.Shapefile)
	}
	if got.Type != "Choropleth" {
		t.Errorf("Type: got %q", got.Type)
	}
	if got.DataSource != "GeoDS" {
		t.Errorf("DataSource: got %q", got.DataSource)
	}
	if got.Filter != "[Country] = 'US'" {
		t.Errorf("Filter: got %q", got.Filter)
	}
	if got.SpatialColumn != "ISO" {
		t.Errorf("SpatialColumn: got %q", got.SpatialColumn)
	}
	if got.SpatialValue != "[ISO]" {
		t.Errorf("SpatialValue: got %q", got.SpatialValue)
	}
	if got.AnalyticalValue != "[GDP]" {
		t.Errorf("AnalyticalValue: got %q", got.AnalyticalValue)
	}
	if got.LabelColumn != "Name" {
		t.Errorf("LabelColumn: got %q", got.LabelColumn)
	}
	if got.BoxAsString != "0,0,100,100" {
		t.Errorf("BoxAsString: got %q", got.BoxAsString)
	}
	if got.Palette != "Blues" {
		t.Errorf("Palette: got %q", got.Palette)
	}
}

// ── MapLayer: Serialize with zero-value fields (skip branches) ───────────────

func TestMapLayer_Serialize_DefaultFields(t *testing.T) {
	// All fields empty — all if-branches in Serialize are skipped.
	orig := object.NewMapLayer()

	var buf bytes.Buffer
	w := serial.NewWriter(&buf)
	if err := w.WriteObjectNamed("MapLayer", orig); err != nil {
		t.Fatalf("Serialize: %v", err)
	}
	w.Flush() //nolint:errcheck

	// Should produce a valid (but attribute-free) element.
	xml := buf.String()
	if !strings.Contains(xml, "MapLayer") {
		t.Errorf("expected MapLayer element in XML:\n%s", xml)
	}

	r := serial.NewReader(bytes.NewReader(buf.Bytes()))
	_, ok := r.ReadObjectHeader()
	if !ok {
		t.Fatal("ReadObjectHeader failed")
	}
	got := object.NewMapLayer()
	if err := got.Deserialize(r); err != nil {
		t.Fatalf("Deserialize: %v", err)
	}
	if got.Shapefile != "" || got.Palette != "" {
		t.Error("all fields should be empty after default round-trip")
	}
}

// ── MapObject: Serialize with no offsets (skip branches) ─────────────────────

func TestMapObject_Serialize_ZeroOffsets(t *testing.T) {
	// OffsetX and OffsetY are both 0 — the if-branches are skipped.
	orig := object.NewMapObject()

	var buf bytes.Buffer
	w := serial.NewWriter(&buf)
	if err := w.WriteObjectNamed("MapObject", orig); err != nil {
		t.Fatalf("Serialize: %v", err)
	}
	w.Flush() //nolint:errcheck

	xml := buf.String()
	if strings.Contains(xml, "OffsetX=") {
		t.Errorf("OffsetX should not be written when zero:\n%s", xml)
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
	if got.OffsetX != 0 || got.OffsetY != 0 {
		t.Errorf("offsets should be 0, got (%v,%v)", got.OffsetX, got.OffsetY)
	}
}

// ── MapObject.DeserializeChild: unknown type returns false ───────────────────

func TestMapObject_DeserializeChild_UnknownType(t *testing.T) {
	xmlStr := `<MapObject OffsetX="5"><UnknownLayer Foo="bar"/></MapObject>`

	r := serial.NewReader(strings.NewReader(xmlStr))
	typeName, ok := r.ReadObjectHeader()
	if !ok {
		t.Fatal("ReadObjectHeader failed")
	}
	if typeName != "MapObject" {
		t.Fatalf("expected MapObject, got %q", typeName)
	}

	m := object.NewMapObject()
	if err := m.Deserialize(r); err != nil {
		t.Fatalf("Deserialize: %v", err)
	}

	for {
		childType, ok := r.NextChild()
		if !ok {
			break
		}
		handled := m.DeserializeChild(childType, r)
		if handled {
			t.Errorf("DeserializeChild(%q) should return false for unknown type", childType)
		}
		r.FinishChild() //nolint:errcheck
	}

	if len(m.Layers) != 0 {
		t.Errorf("Layers should be empty, got %d", len(m.Layers))
	}
}

// ── MapObject.DeserializeChild: MapLayer round-trip ──────────────────────────

func TestMapObject_DeserializeChild_MapLayer_RoundTrip(t *testing.T) {
	orig := object.NewMapObject()
	orig.OffsetX = 3.0
	orig.OffsetY = 7.0

	layer := object.NewMapLayer()
	layer.Shapefile = "states.shp"
	layer.Palette = "Reds"
	orig.Layers = append(orig.Layers, layer)

	var buf bytes.Buffer
	w := serial.NewWriter(&buf)
	if err := w.WriteObjectNamed("MapObject", orig); err != nil {
		t.Fatalf("Serialize: %v", err)
	}
	// Manually write MapLayer as child since MapObject.Serialize doesn't write layers.
	// Re-serialize with explicit layer child using raw XML.
	buf.Reset()

	// Build the XML by hand including the MapLayer child.
	xmlStr := `<MapObject OffsetX="3" OffsetY="7"><MapLayer Shapefile="states.shp" Palette="Reds"/></MapObject>`

	r := serial.NewReader(strings.NewReader(xmlStr))
	_, ok := r.ReadObjectHeader()
	if !ok {
		t.Fatal("ReadObjectHeader failed")
	}
	m := object.NewMapObject()
	if err := m.Deserialize(r); err != nil {
		t.Fatalf("Deserialize: %v", err)
	}
	for {
		childType, ok2 := r.NextChild()
		if !ok2 {
			break
		}
		m.DeserializeChild(childType, r)
		r.FinishChild() //nolint:errcheck
	}

	if len(m.Layers) != 1 {
		t.Fatalf("Layers count: got %d, want 1", len(m.Layers))
	}
	if m.Layers[0].Shapefile != "states.shp" {
		t.Errorf("Shapefile: got %q", m.Layers[0].Shapefile)
	}
	if m.Layers[0].Palette != "Reds" {
		t.Errorf("Palette: got %q", m.Layers[0].Palette)
	}
}
