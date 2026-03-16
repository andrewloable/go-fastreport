package maprender_test

// maprender_coverage2_test.go — targeted tests to reach 100% coverage on
// Render (remaining branch) and charBitmap (unknown character / missing cases).

import (
	"testing"

	"github.com/andrewloable/go-fastreport/maprender"
)

// TestRender_MixedLayers exercises the branch at line ~163:
//
//	if len(layer.GeoFeatures) == 0 { continue }
//
// inside the `if hasUserFeatures` block.  When hasUserFeatures is true (at
// least one layer has GeoFeatures) but another layer in the same Options has
// no GeoFeatures, the `continue` is taken for that layer.
func TestRender_MixedLayers_SomeWithGeoFeatures(t *testing.T) {
	img := maprender.Render(maprender.Options{
		Width:  400,
		Height: 200,
		Layers: []maprender.Layer{
			{
				// This layer HAS GeoFeatures → sets hasUserFeatures = true.
				Type:    "Choropleth",
				Palette: "Heat",
				GeoFeatures: []maprender.GeoFeature{
					{
						Value: 75,
						Polygons: [][][2]float64{
							{{-10, 35}, {40, 35}, {40, 70}, {-10, 70}, {-10, 35}},
						},
					},
				},
			},
			{
				// This layer has NO GeoFeatures → the inner `continue` is taken.
				Type:    "Choropleth",
				Palette: "Blue",
			},
		},
	})
	if img == nil {
		t.Fatal("Render with mixed layers (some with GeoFeatures) returned nil")
	}
}

// TestRender_CharBitmap_GI triggers the 'g' and 'i' cases in charBitmap.
// Palette name "gi" contains both characters, which appear in the label
// "Map (1 layer) · gi" — hitting the otherwise-uncovered cases.
func TestRender_CharBitmap_GI(t *testing.T) {
	// 'g' case: charBitmap('g')
	// 'i' case: charBitmap('i')
	// Both appear in the palette label suffix " · gi".
	img := maprender.Render(maprender.Options{
		Width:  400,
		Height: 200,
		Layers: []maprender.Layer{
			{Palette: "gi"},
		},
	})
	if img == nil {
		t.Fatal("Render with palette 'gi' returned nil")
	}
}

// TestRender_CharBitmap_Gi exercises lowercase 'g' and 'i' via a longer
// palette name to ensure both switch arms are taken.
func TestRender_CharBitmap_LongGIPalette(t *testing.T) {
	// Label: "Map (1 layer) · giving" — contains both 'g' and 'i'.
	img := maprender.Render(maprender.Options{
		Width:  400,
		Height: 200,
		Layers: []maprender.Layer{
			{Palette: "giving"},
		},
	})
	if img == nil {
		t.Fatal("Render with palette 'giving' returned nil")
	}
}

// TestRender_DrawPoly_TooFewPoints exercises the early-return branch inside
// the drawPoly closure when a GeoFeature polygon has fewer than 3 points.
// The closure returns immediately without drawing, hitting the uncovered branch.
func TestRender_DrawPoly_TooFewPoints(t *testing.T) {
	img := maprender.Render(maprender.Options{
		Width:  400,
		Height: 200,
		Layers: []maprender.Layer{
			{
				Type:    "Choropleth",
				Palette: "Heat",
				GeoFeatures: []maprender.GeoFeature{
					{
						// Only 2 points — fewer than 3 → drawPoly early return.
						Value: 50,
						Polygons: [][][2]float64{
							{{0, 0}, {10, 10}},
						},
					},
					{
						// Valid polygon with 4+ points.
						Value: 80,
						Polygons: [][][2]float64{
							{{-10, 35}, {40, 35}, {40, 70}, {-10, 70}, {-10, 35}},
						},
					},
				},
			},
		},
	})
	if img == nil {
		t.Fatal("Render with too-few-point polygon returned nil")
	}
}
