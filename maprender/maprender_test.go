package maprender_test

import (
	"image/color"
	"testing"

	"github.com/andrewloable/go-fastreport/maprender"
)

// ── Basic rendering ────────────────────────────────────────────────────────────

func TestRender_DefaultOptions(t *testing.T) {
	img := maprender.Render(maprender.Options{})
	if img == nil {
		t.Fatal("Render(empty) returned nil")
	}
	b := img.Bounds()
	if b.Max.X != 400 || b.Max.Y != 200 {
		t.Errorf("default size = %v, want 400x200", b)
	}
}

func TestRender_CustomSize(t *testing.T) {
	img := maprender.Render(maprender.Options{Width: 300, Height: 150})
	b := img.Bounds()
	if b.Max.X != 300 || b.Max.Y != 150 {
		t.Errorf("size = %v, want 300x150", b)
	}
}

func TestRender_HasNonBluePixels(t *testing.T) {
	// Continental outlines should create green (land) pixels.
	img := maprender.Render(maprender.Options{Width: 400, Height: 200})
	found := false
	b := img.Bounds()
	for y := b.Min.Y; y < b.Max.Y && !found; y++ {
		for x := b.Min.X; x < b.Max.X && !found; x++ {
			r, g, bv, a := img.At(x, y).RGBA()
			// Ocean is light blue (173,216,230); land is (200,220,160)
			if a > 0 && g > r && g > bv { // greenish
				found = true
			}
		}
	}
	if !found {
		t.Error("expected land (green) pixels in default render")
	}
}

// ── GeoJSON support ───────────────────────────────────────────────────────────

func TestParseGeoJSON_SinglePolygon(t *testing.T) {
	geoJSON := []byte(`{
		"type": "FeatureCollection",
		"features": [{
			"type": "Feature",
			"id": "USA",
			"properties": {"name": "United States"},
			"geometry": {
				"type": "Polygon",
				"coordinates": [[[-125, 49], [-65, 49], [-65, 25], [-125, 25], [-125, 49]]]
			}
		}]
	}`)

	features, err := maprender.ParseGeoJSON(geoJSON)
	if err != nil {
		t.Fatalf("ParseGeoJSON: %v", err)
	}
	if len(features) != 1 {
		t.Fatalf("feature count = %d, want 1", len(features))
	}
	f := features[0]
	if f.Name != "United States" {
		t.Errorf("Name = %q, want United States", f.Name)
	}
	if len(f.Polygons) != 1 {
		t.Errorf("polygons = %d, want 1", len(f.Polygons))
	}
}

func TestParseGeoJSON_MultiPolygon(t *testing.T) {
	geoJSON := []byte(`{
		"type": "FeatureCollection",
		"features": [{
			"type": "Feature",
			"properties": {"name": "Multi"},
			"geometry": {
				"type": "MultiPolygon",
				"coordinates": [
					[[[0,0],[10,0],[10,10],[0,10],[0,0]]],
					[[[20,20],[30,20],[30,30],[20,30],[20,20]]]
				]
			}
		}]
	}`)

	features, err := maprender.ParseGeoJSON(geoJSON)
	if err != nil {
		t.Fatalf("ParseGeoJSON: %v", err)
	}
	if len(features) != 1 {
		t.Fatalf("count = %d", len(features))
	}
	if len(features[0].Polygons) != 2 {
		t.Errorf("polygons = %d, want 2 (one per MultiPolygon part)", len(features[0].Polygons))
	}
}

func TestRender_WithGeoFeatures(t *testing.T) {
	// Supply a simple square polygon over Europe.
	features := []maprender.GeoFeature{
		{
			Name:  "Region A",
			Value: 100,
			Polygons: [][][2]float64{
				{{-10, 35}, {40, 35}, {40, 70}, {-10, 70}, {-10, 35}},
			},
		},
		{
			Name:  "Region B",
			Value: 50,
			Polygons: [][][2]float64{
				{{-20, -35}, {55, -35}, {55, 35}, {-20, 35}, {-20, -35}},
			},
		},
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
		t.Fatal("Render with GeoFeatures = nil")
	}
}

// ── Bubble overlays ────────────────────────────────────────────────────────────

func TestRender_WithBubbles(t *testing.T) {
	bubbles := []maprender.Bubble{
		{Lon: 0, Lat: 51, Value: 100, Color: color.NRGBA{200, 50, 50, 200}, Label: "London"},
		{Lon: 2, Lat: 48, Value: 80, Color: color.NRGBA{50, 50, 200, 200}, Label: "Paris"},
	}

	img := maprender.Render(maprender.Options{
		Width:  400,
		Height: 200,
		Layers: []maprender.Layer{
			{Type: "Bubble", Bubbles: bubbles},
		},
	})
	if img == nil {
		t.Fatal("Render with bubbles = nil")
	}
	// Check for red pixels (London bubble).
	b := img.Bounds()
	found := false
	for y := b.Min.Y; y < b.Max.Y && !found; y++ {
		for x := b.Min.X; x < b.Max.X && !found; x++ {
			r, g, bv, a := img.At(x, y).RGBA()
			if a > 0 && r > 0x8000 && g < 0x4000 && bv < 0x4000 {
				found = true
			}
		}
	}
	if !found {
		t.Error("expected red bubble pixels from London bubble")
	}
}

// ── Legend ────────────────────────────────────────────────────────────────────

func TestRender_WithLegend(t *testing.T) {
	features := []maprender.GeoFeature{
		{Value: 10, Polygons: [][][2]float64{{{-10, 35}, {40, 35}, {40, 70}, {-10, 70}, {-10, 35}}}},
		{Value: 90, Polygons: [][][2]float64{{{-20, -35}, {55, -35}, {55, 35}, {-20, 35}, {-20, -35}}}},
	}

	img := maprender.Render(maprender.Options{
		Width:  400,
		Height: 200,
		Layers: []maprender.Layer{
			{
				Type:        "Choropleth",
				Palette:     "Blue",
				GeoFeatures: features,
				ShowLegend:  true,
			},
		},
	})
	if img == nil {
		t.Fatal("Render with legend = nil")
	}
}

// ── Choropleth palettes ────────────────────────────────────────────────────────

func TestNamedPalette_Coverage(t *testing.T) {
	names := []string{"Heat", "Blue", "Green", "Spectral", "Default", "Unknown"}
	for _, name := range names {
		p := maprender.NamedPalette(name)
		if p == nil {
			t.Errorf("NamedPalette(%q) = nil", name)
			continue
		}
		// Test 0, 0.5, 1.0.
		for _, t2 := range []float64{0, 0.5, 1.0} {
			c := p.Color(t2)
			if c.A == 0 {
				t.Errorf("palette %q at t=%v: alpha=0", name, t2)
			}
		}
	}
}

func TestLinearPalette_Interpolation(t *testing.T) {
	p := &maprender.LinearPalette{
		Low:  color.NRGBA{0, 0, 0, 255},
		High: color.NRGBA{200, 200, 200, 255},
	}
	mid := p.Color(0.5)
	if mid.R < 90 || mid.R > 110 {
		t.Errorf("midpoint R = %d, want ~100", mid.R)
	}
}
