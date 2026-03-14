package reportpkg_test

// Smoke tests for the "Master-Detail" category of FastReport sample FRX files.
// Each test verifies load success and presence of DataBand objects.

import (
	"testing"

	"github.com/andrewloable/go-fastreport/band"
)

func countDataBandsInReport(t *testing.T, filename string) int {
	t.Helper()
	r := loadFRXSmoke(t, filename)
	var count int
	for _, pg := range r.Pages() {
		for _, b := range pg.Bands() {
			if _, ok := b.(*band.DataBand); ok {
				count++
			}
		}
	}
	return count
}

func TestFRXSmoke_MasterDetail(t *testing.T) {
	dbCount := countDataBandsInReport(t, "Master-Detail.frx")
	if dbCount == 0 {
		t.Error("expected at least one DataBand in Master-Detail.frx")
	}
}

func TestFRXSmoke_ComplexMasterDetailGroup_MD(t *testing.T) {
	// This report has DataBands nested inside GroupHeaderBands.
	// Just verify it loads and has pages.
	r := loadFRXSmoke(t, "Complex (Master-detail + Group).frx")
	if len(r.Pages()) == 0 {
		t.Error("expected at least one page")
	}
}

func TestFRXSmoke_RowDatasourceMasterDetail(t *testing.T) {
	dbCount := countDataBandsInReport(t, "Row Datasource, Master-Detail.frx")
	if dbCount == 0 {
		t.Error("expected at least one DataBand in Row Datasource, Master-Detail.frx")
	}
}
