package reportpkg_test

// Tests for FRX Sort deserialization in DataBand.

import (
	"testing"

	"github.com/andrewloable/go-fastreport/band"
	"github.com/andrewloable/go-fastreport/report"
	"github.com/andrewloable/go-fastreport/reportpkg"
)

// findDataBands collects all DataBand instances from the report.
func findDataBands(r *reportpkg.Report) []*band.DataBand {
	var result []*band.DataBand
	var collectBands func(b report.Base)
	collectBands = func(b report.Base) {
		if db, ok := b.(*band.DataBand); ok {
			result = append(result, db)
		}
		// Recurse into GroupHeaderBand's nested DataBand and NestedGroup.
		if gh, ok := b.(*band.GroupHeaderBand); ok {
			if gh.Data() != nil {
				collectBands(gh.Data())
			}
			if gh.NestedGroup() != nil {
				collectBands(gh.NestedGroup())
			}
		}
	}
	for _, pg := range r.Pages() {
		for _, b := range pg.AllBands() {
			collectBands(b)
		}
	}
	return result
}

func TestDataBandSort_Deserializes(t *testing.T) {
	// Badges.frx has a DataBand with two Sort items (FirstName, LastName).
	r := loadFRXSmoke(t, "Badges.frx")
	dbs := findDataBands(r)
	var sorted *band.DataBand
	for _, db := range dbs {
		if len(db.Sort()) > 0 {
			sorted = db
			break
		}
	}
	if sorted == nil {
		t.Fatal("expected at least one DataBand with sort specs in Badges.frx")
	}
	if len(sorted.Sort()) < 2 {
		t.Errorf("expected 2 sort specs, got %d", len(sorted.Sort()))
	}
}

func TestDataBandSort_SingleSort(t *testing.T) {
	// Groups.frx has a DataBand with one Sort item.
	r := loadFRXSmoke(t, "Groups.frx")
	dbs := findDataBands(r)
	var sorted *band.DataBand
	for _, db := range dbs {
		if len(db.Sort()) > 0 {
			sorted = db
			break
		}
	}
	if sorted == nil {
		t.Fatal("expected at least one DataBand with sort specs in Groups.frx")
	}
	spec := sorted.Sort()[0]
	if spec.Column == "" {
		t.Error("sort spec Column should be non-empty")
	}
}
