package maprender

import (
	"encoding/json"
	"image/color"
	"math"
)

// ── GeoJSON types ─────────────────────────────────────────────────────────────

// GeoFeature is a single GeoJSON Feature.
type GeoFeature struct {
	// ID is the feature identifier (optional — from properties or "id" field).
	ID string
	// Name is a human-readable name (e.g. from properties.name).
	Name string
	// Polygons holds all polygon rings for this feature.
	// Each polygon is a slice of [lon, lat] pairs (exterior ring only).
	Polygons [][][2]float64
	// Value is a numeric data value for choropleth/bubble sizing (set by caller).
	Value float64
	// BubbleLon/BubbleLat is the bubble centre (centroid if not specified).
	BubbleLon, BubbleLat float64
}

// ParseGeoJSON decodes a GeoJSON FeatureCollection or Geometry string
// and returns the extracted features.
func ParseGeoJSON(data []byte) ([]GeoFeature, error) {
	var raw struct {
		Type     string          `json:"type"`
		Features []rawFeature    `json:"features"`
		// If type is directly Geometry, handle as single feature.
		Geometries []rawGeometry `json:"geometries"`
	}
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, err
	}

	var features []GeoFeature
	for _, f := range raw.Features {
		feat := GeoFeature{}
		feat.ID = f.ID
		if n, ok := f.Properties["name"]; ok {
			if ns, ok2 := n.(string); ok2 {
				feat.Name = ns
			}
		}
		feat.Polygons = extractPolygons(f.Geometry)
		if len(feat.Polygons) > 0 {
			feat.BubbleLon, feat.BubbleLat = centroid(feat.Polygons[0])
		}
		features = append(features, feat)
	}
	return features, nil
}

type rawFeature struct {
	ID         string                 `json:"id"`
	Type       string                 `json:"type"`
	Properties map[string]interface{} `json:"properties"`
	Geometry   rawGeometry            `json:"geometry"`
}

type rawGeometry struct {
	Type        string          `json:"type"`
	Coordinates json.RawMessage `json:"coordinates"`
}

func extractPolygons(g rawGeometry) [][][2]float64 {
	switch g.Type {
	case "Polygon":
		var rings [][][2]float64
		if err := json.Unmarshal(g.Coordinates, &rings); err == nil && len(rings) > 0 {
			// Return only the exterior ring (first ring) as a polygon.
			return rings[:1]
		}
	case "MultiPolygon":
		var multiRings [][][][2]float64
		if err := json.Unmarshal(g.Coordinates, &multiRings); err == nil {
			var result [][][2]float64
			for _, polygon := range multiRings {
				if len(polygon) > 0 {
					result = append(result, polygon[0])
				}
			}
			return result
		}
	case "Point":
		// Points are handled as bubble centres — no polygon.
		return nil
	}
	return nil
}

// centroid returns the arithmetic centroid of a polygon ring.
func centroid(ring [][2]float64) (float64, float64) {
	if len(ring) == 0 {
		return 0, 0
	}
	var sumX, sumY float64
	for _, pt := range ring {
		sumX += pt[0]
		sumY += pt[1]
	}
	n := float64(len(ring))
	return sumX / n, sumY / n
}

// ── Choropleth palette ────────────────────────────────────────────────────────

// ChoroplethPalette maps a normalised value [0,1] to a color.
type ChoroplethPalette interface {
	Color(t float64) color.NRGBA
}

// LinearPalette interpolates linearly between two colors.
type LinearPalette struct {
	Low, High color.NRGBA
}

func (p *LinearPalette) Color(t float64) color.NRGBA {
	t = clamp01(t)
	return color.NRGBA{
		R: uint8(float64(p.Low.R) + t*float64(int(p.High.R)-int(p.Low.R))),
		G: uint8(float64(p.Low.G) + t*float64(int(p.High.G)-int(p.Low.G))),
		B: uint8(float64(p.Low.B) + t*float64(int(p.High.B)-int(p.Low.B))),
		A: 255,
	}
}

// StepPalette maps discrete values to colors via a list.
type StepPalette struct {
	Colors []color.NRGBA
}

func (p *StepPalette) Color(t float64) color.NRGBA {
	if len(p.Colors) == 0 {
		return color.NRGBA{200, 200, 200, 255}
	}
	idx := int(t * float64(len(p.Colors)))
	if idx >= len(p.Colors) {
		idx = len(p.Colors) - 1
	}
	return p.Colors[idx]
}

// NamedPalette returns a ChoroplethPalette for common palette names.
func NamedPalette(name string) ChoroplethPalette {
	switch name {
	case "Heat", "Red":
		return &LinearPalette{
			Low:  color.NRGBA{255, 240, 200, 255},
			High: color.NRGBA{180, 0, 0, 255},
		}
	case "Blue", "Ocean":
		return &LinearPalette{
			Low:  color.NRGBA{210, 235, 255, 255},
			High: color.NRGBA{0, 50, 150, 255},
		}
	case "Green":
		return &LinearPalette{
			Low:  color.NRGBA{210, 245, 210, 255},
			High: color.NRGBA{0, 100, 0, 255},
		}
	case "Spectral":
		return &StepPalette{Colors: []color.NRGBA{
			{215, 48, 39, 255},
			{252, 141, 89, 255},
			{254, 224, 139, 255},
			{255, 255, 191, 255},
			{217, 239, 139, 255},
			{145, 207, 96, 255},
			{26, 152, 80, 255},
		}}
	default: // "Default"
		return &LinearPalette{
			Low:  color.NRGBA{200, 220, 160, 255},
			High: color.NRGBA{50, 120, 50, 255},
		}
	}
}

// ── Bubble ────────────────────────────────────────────────────────────────────

// Bubble represents a circle overlay at a geographic position.
type Bubble struct {
	Lon, Lat float64 // centre in degrees
	Value    float64 // controls radius scaling
	Color    color.NRGBA
	Label    string
}

func clamp01(v float64) float64 {
	return math.Max(0, math.Min(1, v))
}
