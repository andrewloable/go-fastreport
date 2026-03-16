package maprender_test

import (
	"image"
	"testing"

	"github.com/andrewloable/go-fastreport/maprender"
)

// Ensure image is used.
var _ image.Image

// ── clampInt edge cases ───────────────────────────────────────────────────────

// clampInt is exercised via drawLegend when the legend is rendered.
// We need to hit the v < 0 and v >= max branches directly through the legend.

func TestRender_ClampInt_ViaLegend_LargeImage(t *testing.T) {
	// drawLegend uses clampInt for border drawing.
	// With a large image, the clamp values will use the normal path (v in range).
	features := []maprender.GeoFeature{
		{Value: 10, Polygons: [][][2]float64{{{-10, 35}, {40, 35}, {40, 70}, {-10, 70}, {-10, 35}}}},
	}
	img := maprender.Render(maprender.Options{
		Width:  800,
		Height: 400,
		Layers: []maprender.Layer{
			{
				Type:        "Choropleth",
				Palette:     "Green",
				GeoFeatures: features,
				ShowLegend:  true,
			},
		},
	})
	if img == nil {
		t.Fatal("Render with legend returned nil")
	}
	b := img.Bounds()
	if b.Dx() != 800 || b.Dy() != 400 {
		t.Errorf("size = %dx%d, want 800x400", b.Dx(), b.Dy())
	}
}

func TestRender_ClampInt_NegativeV_ViaSmallLegend(t *testing.T) {
	// Render with a tiny image so legend coordinates may go negative.
	features := []maprender.GeoFeature{
		{Value: 5, Polygons: [][][2]float64{{{-10, 35}, {40, 35}, {40, 70}, {-10, 70}, {-10, 35}}}},
		{Value: 50, Polygons: [][][2]float64{{{20, -35}, {55, -35}, {55, 35}, {20, 35}, {20, -35}}}},
	}
	img := maprender.Render(maprender.Options{
		Width:  50, // Small enough that legend x = width-60 is negative → clampInt(v<0) → 0
		Height: 80,
		Layers: []maprender.Layer{
			{
				Type:        "Choropleth",
				Palette:     "Heat",
				GeoFeatures: features,
				ShowLegend:  true,
			},
		},
	})
	if img == nil {
		t.Fatal("Render small with legend returned nil")
	}
}

// ── charBitmap coverage ───────────────────────────────────────────────────────

// charBitmap is called via drawBanner which is called from Render.
// We exercise all uncovered rune cases by triggering labels with those characters.

func TestRender_CharBitmap_AllChars(t *testing.T) {
	// Each render produces a label like "Map (N layer)" with specific chars.
	// Force palette label to cover '·' and more chars.
	img := maprender.Render(maprender.Options{
		Width:  400,
		Height: 200,
		Layers: []maprender.Layer{
			{Palette: "Spectral"},
			{Palette: "Blue"},
			{Palette: "Green"},
			{Palette: "Heat"},
		},
	})
	if img == nil {
		t.Fatal("multi-layer render returned nil")
	}
}

func TestRender_CharBitmap_SingleLayer_Palette(t *testing.T) {
	// Label: "Map (1 layer) · Spectral" — covers 'S', 'p', 'e', 'c', 't', 'r', 'a', 'l', '·'
	img := maprender.Render(maprender.Options{
		Width:  400,
		Height: 200,
		Layers: []maprender.Layer{
			{Palette: "Spectral"},
		},
	})
	if img == nil {
		t.Fatal("render returned nil")
	}
}

func TestRender_CharBitmap_SingleLayer_NoPalette(t *testing.T) {
	// No palette → label is "Map (1 layer)" without the "· palette" suffix.
	img := maprender.Render(maprender.Options{
		Width:  400,
		Height: 200,
		Layers: []maprender.Layer{
			{Type: "Choropleth"},
		},
	})
	if img == nil {
		t.Fatal("render without palette returned nil")
	}
}

func TestRender_CharBitmap_DefaultBitmap(t *testing.T) {
	// Using a palette name with characters not in the switch default case.
	// The default charBitmap returns the "unknown" glyph (0b111, 0b101, …).
	// Characters like 'z', 'x', 'q', 'u' etc. not explicitly listed fall through.
	img := maprender.Render(maprender.Options{
		Width:  400,
		Height: 200,
		Layers: []maprender.Layer{
			{Palette: "Custom-xyz-palett"},
		},
	})
	if img == nil {
		t.Fatal("render with custom palette returned nil")
	}
}

// ── extractPolygons coverage ──────────────────────────────────────────────────

func TestParseGeoJSON_PointGeometry(t *testing.T) {
	// Point geometry → extractPolygons returns nil (no polygon).
	geoJSON := []byte(`{
		"type": "FeatureCollection",
		"features": [{
			"type": "Feature",
			"properties": {"name": "Capital"},
			"geometry": {
				"type": "Point",
				"coordinates": [0.0, 51.5]
			}
		}]
	}`)

	features, err := maprender.ParseGeoJSON(geoJSON)
	if err != nil {
		t.Fatalf("ParseGeoJSON Point: %v", err)
	}
	if len(features) != 1 {
		t.Fatalf("count = %d, want 1", len(features))
	}
	if len(features[0].Polygons) != 0 {
		t.Errorf("Point should produce 0 polygons, got %d", len(features[0].Polygons))
	}
}

func TestParseGeoJSON_UnknownGeometryType(t *testing.T) {
	// Unknown geometry type → returns nil.
	geoJSON := []byte(`{
		"type": "FeatureCollection",
		"features": [{
			"type": "Feature",
			"properties": {"name": "Line"},
			"geometry": {
				"type": "LineString",
				"coordinates": [[0,0],[1,1]]
			}
		}]
	}`)

	features, err := maprender.ParseGeoJSON(geoJSON)
	if err != nil {
		t.Fatalf("ParseGeoJSON LineString: %v", err)
	}
	if len(features) != 1 {
		t.Fatalf("count = %d, want 1", len(features))
	}
	if len(features[0].Polygons) != 0 {
		t.Errorf("LineString should produce 0 polygons, got %d", len(features[0].Polygons))
	}
}

func TestParseGeoJSON_InvalidJSON(t *testing.T) {
	_, err := maprender.ParseGeoJSON([]byte(`not json`))
	if err == nil {
		t.Error("expected error for invalid JSON")
	}
}

func TestParseGeoJSON_EmptyPolygonCoordinates(t *testing.T) {
	// Polygon with empty coordinates array → unmarshal succeeds but len(rings)==0.
	geoJSON := []byte(`{
		"type": "FeatureCollection",
		"features": [{
			"type": "Feature",
			"properties": {"name": "Empty"},
			"geometry": {
				"type": "Polygon",
				"coordinates": []
			}
		}]
	}`)

	features, err := maprender.ParseGeoJSON(geoJSON)
	if err != nil {
		t.Fatalf("ParseGeoJSON empty coords: %v", err)
	}
	// Feature should still be extracted but with 0 polygons.
	if len(features) != 1 {
		t.Fatalf("count = %d, want 1", len(features))
	}
	if len(features[0].Polygons) != 0 {
		t.Errorf("Empty polygon should produce 0 polygons, got %d", len(features[0].Polygons))
	}
}

// ── centroid edge cases ───────────────────────────────────────────────────────

func TestParseGeoJSON_CentroidOfEmptyRing(t *testing.T) {
	// If a MultiPolygon ring is empty, centroid is not called (no polygon[0]).
	// We need to test centroid(empty) which is called when Polygons[0] exists but is empty.
	// Actually centroid(ring) with len(ring)==0 returns (0,0) — verify via empty polygon rings.
	geoJSON := []byte(`{
		"type": "FeatureCollection",
		"features": [{
			"type": "Feature",
			"properties": {"name": "Test"},
			"geometry": {
				"type": "Polygon",
				"coordinates": [[[0,0],[10,0],[10,10],[0,10],[0,0]]]
			}
		}]
	}`)

	features, err := maprender.ParseGeoJSON(geoJSON)
	if err != nil {
		t.Fatalf("ParseGeoJSON: %v", err)
	}
	if len(features) == 0 {
		t.Fatal("expected feature")
	}
	// BubbleLon and BubbleLat should be set to centroid.
	f := features[0]
	if f.BubbleLon == 0 && f.BubbleLat == 0 {
		// Centroid of the above polygon is (4,4) approximately.
		t.Logf("BubbleLon=%v, BubbleLat=%v (may vary)", f.BubbleLon, f.BubbleLat)
	}
}

// ── Render edge cases ─────────────────────────────────────────────────────────

func TestRender_ZoomSet(t *testing.T) {
	// Non-default zoom.
	img := maprender.Render(maprender.Options{
		Width:  400,
		Height: 200,
		Zoom:   2.0,
	})
	if img == nil {
		t.Fatal("Render with Zoom=2.0 returned nil")
	}
}

func TestRender_BubbleZeroValue(t *testing.T) {
	// All bubbles have zero value → maxVal stays 1.
	img := maprender.Render(maprender.Options{
		Width:  400,
		Height: 200,
		Layers: []maprender.Layer{
			{
				Type: "Bubble",
				Bubbles: []maprender.Bubble{
					{Lon: 0, Lat: 0, Value: 0},
				},
			},
		},
	})
	if img == nil {
		t.Fatal("Render bubble zero value returned nil")
	}
}

func TestRender_BubbleZeroColor(t *testing.T) {
	// Bubble with zero Alpha → uses default color.
	img := maprender.Render(maprender.Options{
		Width:  400,
		Height: 200,
		Layers: []maprender.Layer{
			{
				Type: "Bubble",
				Bubbles: []maprender.Bubble{
					{Lon: 10, Lat: 20, Value: 50}, // Color.A == 0 → default red
				},
			},
		},
	})
	if img == nil {
		t.Fatal("Render bubble zero color returned nil")
	}
}

func TestRender_SingleLayerLabel(t *testing.T) {
	// 1 layer → "Map (1 layer)" (no 's').
	img := maprender.Render(maprender.Options{
		Width:  400,
		Height: 200,
		Layers: []maprender.Layer{{Type: "Choropleth"}},
	})
	if img == nil {
		t.Fatal("render nil")
	}
}

func TestRender_MultiLayerLabel(t *testing.T) {
	// >1 layers → "Map (N layers)" (with 's').
	img := maprender.Render(maprender.Options{
		Width:  400,
		Height: 200,
		Layers: []maprender.Layer{{}, {}},
	})
	if img == nil {
		t.Fatal("render nil")
	}
}

func TestRender_TwoLayers(t *testing.T) {
	// "Map (2 layers)" — exercises '2' in charBitmap.
	img := maprender.Render(maprender.Options{
		Width:  400,
		Height: 200,
		Layers: []maprender.Layer{{}, {}},
	})
	if img == nil {
		t.Fatal("render nil")
	}
}

func TestRender_ThreeLayers(t *testing.T) {
	// "Map (3 layers)" — exercises '3' in charBitmap.
	img := maprender.Render(maprender.Options{
		Width:  400,
		Height: 200,
		Layers: []maprender.Layer{{}, {}, {}},
	})
	if img == nil {
		t.Fatal("render nil")
	}
}

func TestRender_FourLayers(t *testing.T) {
	// "Map (4 layers)" — exercises '4' in charBitmap.
	img := maprender.Render(maprender.Options{
		Width:  400,
		Height: 200,
		Layers: []maprender.Layer{{}, {}, {}, {}},
	})
	if img == nil {
		t.Fatal("render nil")
	}
}

func TestRender_FiveLayers(t *testing.T) {
	// "Map (5 layers)" — exercises '5' in charBitmap.
	img := maprender.Render(maprender.Options{
		Width:  400,
		Height: 200,
		Layers: []maprender.Layer{{}, {}, {}, {}, {}},
	})
	if img == nil {
		t.Fatal("render nil")
	}
}

func TestRender_DigitsInLabel(t *testing.T) {
	// Palette "Green" → label "Map (1 layer) · Green" — exercises 'g', 'n'.
	img := maprender.Render(maprender.Options{
		Width:  400,
		Height: 200,
		Layers: []maprender.Layer{{Palette: "Green"}},
	})
	if img == nil {
		t.Fatal("render nil")
	}
}

func TestRender_PaletteOcean(t *testing.T) {
	// "Ocean" → exercises 'o', 'c'.
	img := maprender.Render(maprender.Options{
		Width:  400,
		Height: 200,
		Layers: []maprender.Layer{{Palette: "Ocean"}},
	})
	if img == nil {
		t.Fatal("render nil")
	}
}

// ── StepPalette coverage ──────────────────────────────────────────────────────

func TestStepPalette_Empty(t *testing.T) {
	p := &maprender.StepPalette{}
	c := p.Color(0.5)
	if c.A == 0 {
		t.Error("StepPalette empty should return fallback with alpha>0")
	}
}

func TestStepPalette_AtOne(t *testing.T) {
	// t=1.0 → idx=len(Colors), clamped to len-1.
	p := maprender.NamedPalette("Spectral")
	c := p.Color(1.0)
	if c.A == 0 {
		t.Error("StepPalette at t=1.0 should have alpha>0")
	}
}

// ── LinearPalette coverage ────────────────────────────────────────────────────

func TestLinearPalette_AtZero(t *testing.T) {
	p := maprender.NamedPalette("Blue")
	c := p.Color(0)
	if c.A == 0 {
		t.Error("Color at 0 should have alpha>0")
	}
}

func TestLinearPalette_AtOne(t *testing.T) {
	p := maprender.NamedPalette("Blue")
	c := p.Color(1)
	if c.A == 0 {
		t.Error("Color at 1 should have alpha>0")
	}
}

func TestLinearPalette_Clamp(t *testing.T) {
	p := maprender.NamedPalette("Blue")
	// Values outside [0,1] are clamped.
	c1 := p.Color(-0.5)
	c2 := p.Color(1.5)
	if c1.A == 0 || c2.A == 0 {
		t.Error("clamped values should have alpha>0")
	}
}

// ── GeoJSON name not in properties ───────────────────────────────────────────

func TestParseGeoJSON_NoNameProperty(t *testing.T) {
	geoJSON := []byte(`{
		"type": "FeatureCollection",
		"features": [{
			"type": "Feature",
			"properties": {},
			"geometry": {
				"type": "Polygon",
				"coordinates": [[[0,0],[1,0],[1,1],[0,1],[0,0]]]
			}
		}]
	}`)
	features, err := maprender.ParseGeoJSON(geoJSON)
	if err != nil {
		t.Fatalf("ParseGeoJSON: %v", err)
	}
	if len(features) == 0 {
		t.Fatal("expected feature")
	}
	if features[0].Name != "" {
		t.Errorf("Name = %q, want empty", features[0].Name)
	}
}

// ── centroid with empty ring ──────────────────────────────────────────────────

func TestParseGeoJSON_CentroidEmptyRing(t *testing.T) {
	// Polygon with a single empty ring → extractPolygons returns a polygon of 0 points.
	// Then centroid([]) → returns (0,0).
	geoJSON := []byte(`{
		"type": "FeatureCollection",
		"features": [{
			"type": "Feature",
			"properties": {"name": "Empty Ring"},
			"geometry": {
				"type": "Polygon",
				"coordinates": [[]]
			}
		}]
	}`)
	features, err := maprender.ParseGeoJSON(geoJSON)
	if err != nil {
		t.Fatalf("ParseGeoJSON empty ring: %v", err)
	}
	if len(features) == 0 {
		t.Fatal("expected 1 feature")
	}
	// BubbleLon and BubbleLat should both be 0 (centroid of empty ring).
	f := features[0]
	if f.BubbleLon != 0 || f.BubbleLat != 0 {
		t.Errorf("centroid of empty ring: got (%v,%v), want (0,0)", f.BubbleLon, f.BubbleLat)
	}
}

// ── GeoFeatures with all-same value (valRange==0) ─────────────────────────────

func TestRender_GeoFeatures_AllSameValue(t *testing.T) {
	// valRange == 0 → set to 1 to avoid division by zero.
	features := []maprender.GeoFeature{
		{Value: 42, Polygons: [][][2]float64{{{-10, 35}, {40, 35}, {40, 70}, {-10, 70}, {-10, 35}}}},
		{Value: 42, Polygons: [][][2]float64{{{-20, -35}, {55, -35}, {55, 35}, {-20, 35}, {-20, -35}}}},
	}
	img := maprender.Render(maprender.Options{
		Width:  400,
		Height: 200,
		Layers: []maprender.Layer{
			{
				Type:        "Choropleth",
				Palette:     "Heat",
				GeoFeatures: features,
			},
		},
	})
	if img == nil {
		t.Fatal("render nil")
	}
	if _, ok := img.(*image.NRGBA); !ok {
		t.Error("expected *image.NRGBA")
	}
}
