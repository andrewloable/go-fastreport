package reportpkg_test

// Smoke tests for shape, line, and checkbox FRX reports.

import (
	"testing"

	"github.com/andrewloable/go-fastreport/object"
	"github.com/andrewloable/go-fastreport/report"
	"github.com/andrewloable/go-fastreport/reportpkg"
)

// countObjectsOfType counts objects of a given type across all band objects in r.
func countObjectsOfType[T any](r *reportpkg.Report) int {
	var count int
	type hasObjects interface {
		Objects() *report.ObjectCollection
	}
	for _, pg := range r.Pages() {
		// AllBands includes singleton bands (ReportTitle, PageHeader, ReportSummary, etc.)
		// in addition to the ordered data/group bands.
		for _, b := range pg.AllBands() {
			if h, ok := b.(hasObjects); ok {
				objs := h.Objects()
				for i := 0; i < objs.Len(); i++ {
					if _, ok := objs.Get(i).(T); ok {
						count++
					}
				}
			}
		}
	}
	return count
}

func TestFRXSmoke_LinesAndShapes(t *testing.T) {
	r := loadFRXSmoke(t, "Lines and Shapes.frx")
	if n := countObjectsOfType[*object.LineObject](r); n == 0 {
		t.Error("expected at least one LineObject in Lines and Shapes.frx")
	}
	if n := countObjectsOfType[*object.ShapeObject](r); n == 0 {
		t.Error("expected at least one ShapeObject in Lines and Shapes.frx")
	}
}

func TestFRXSmoke_Box(t *testing.T) {
	r := loadFRXSmoke(t, "Box.frx")
	if len(r.Pages()) == 0 {
		t.Fatal("expected at least one page")
	}
	// Box.frx uses ShapeObjects and PolygonObjects.
	shapes := countObjectsOfType[*object.ShapeObject](r)
	polys := countObjectsOfType[*object.PolygonObject](r)
	if shapes+polys == 0 {
		t.Error("expected ShapeObject or PolygonObject in Box.frx")
	}
}

func TestFRXSmoke_Polygon(t *testing.T) {
	r := loadFRXSmoke(t, "Polygon.frx")
	if n := countObjectsOfType[*object.PolygonObject](r); n == 0 {
		t.Error("expected at least one PolygonObject in Polygon.frx")
	}
}

func TestFRXSmoke_CheckBox(t *testing.T) {
	r := loadFRXSmoke(t, "CheckBox.frx")
	if n := countObjectsOfType[*object.CheckBoxObject](r); n == 0 {
		t.Error("expected at least one CheckBoxObject in CheckBox.frx")
	}
}
