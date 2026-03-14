package reportpkg_test

// Smoke tests for data source and filtering FRX reports.

import (
	"testing"

	"github.com/andrewloable/go-fastreport/band"
)

func TestFRXSmoke_RowDatasource(t *testing.T) {
	loadFRXSmoke(t, "Row Datasource.frx")
}

func TestFRXSmoke_RowDatasourceDetailRows(t *testing.T) {
	loadFRXSmoke(t, "Row Datasource, Detail Rows.frx")
}

func TestFRXSmoke_CascadedDataFiltering(t *testing.T) {
	loadFRXSmoke(t, "Cascaded Data Filtering.frx")
}

func TestFRXSmoke_FilteringWithRanges(t *testing.T) {
	loadFRXSmoke(t, "Filtering with Ranges.frx")
}

func TestFRXSmoke_FilteringWithDateRanges(t *testing.T) {
	loadFRXSmoke(t, "Filtering with Date Ranges.frx")
}

func TestFRXSmoke_HierarchicList(t *testing.T) {
	// Hierarchic List uses DataBand.IdColumn and DataBand.ParentIdColumn
	// to define the tree hierarchy from a flat data source.
	r := loadFRXSmoke(t, "Hierarchic List.frx")
	var found bool
	for _, pg := range r.Pages() {
		for _, b := range pg.Bands() {
			if db, ok := b.(*band.DataBand); ok {
				if db.IDColumn() != "" && db.ParentIDColumn() != "" {
					found = true
				}
			}
		}
	}
	if !found {
		t.Error("expected DataBand with IdColumn and ParentIdColumn in Hierarchic List.frx")
	}
}

func TestFRXSmoke_Badges(t *testing.T) {
	loadFRXSmoke(t, "Badges.frx")
}

func TestFRXSmoke_PurchaseOrder(t *testing.T) {
	loadFRXSmoke(t, "Purchase Order.frx")
}
