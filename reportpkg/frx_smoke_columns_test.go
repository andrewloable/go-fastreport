package reportpkg_test

// Smoke tests for multi-column layout FRX reports.

import (
	"testing"
)

func TestFRXSmoke_ColumnDatasource(t *testing.T) {
	loadFRXSmoke(t, "Column Datasource.frx")
}

func TestFRXSmoke_ColumnDatasourceWrapped(t *testing.T) {
	loadFRXSmoke(t, "Column Datasource, Wrapped.frx")
}

func TestFRXSmoke_DatabandColumnsRowNumbers(t *testing.T) {
	// DataBand with Columns.Count=2 is nested inside a GroupHeaderBand.
	// Verify the file loads; the Columns.Count attribute is deserialized
	// correctly but requires recursive band traversal to find it.
	loadFRXSmoke(t, "Databand Columns, Row Numbers.frx")
}

func TestFRXSmoke_Labels(t *testing.T) {
	// Labels.frx uses ReportPage.Columns.Count > 1 for address-label columns.
	r := loadFRXSmoke(t, "Labels.frx")
	var found bool
	for _, pg := range r.Pages() {
		if pg.Columns.Count > 1 {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected ReportPage with Columns.Count > 1 in Labels.frx")
	}
}

func TestFRXSmoke_ComplexColumnHeaders(t *testing.T) {
	loadFRXSmoke(t, "Complex Column Headers.frx")
}
