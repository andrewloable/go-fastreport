package reportpkg_test

// Smoke tests for subreport FRX reports.

import (
	"testing"

	"github.com/andrewloable/go-fastreport/object"
)

func TestFRXSmoke_Subreport(t *testing.T) {
	r := loadFRXSmoke(t, "Subreport.frx")
	n := countObjectsOfType[*object.SubreportObject](r)
	if n == 0 {
		t.Error("expected at least one SubreportObject in Subreport.frx")
	}
}

func TestFRXSmoke_SideBySideSubreports(t *testing.T) {
	r := loadFRXSmoke(t, "Side-by-Side Subreports.frx")
	n := countObjectsOfType[*object.SubreportObject](r)
	if n == 0 {
		t.Error("expected at least one SubreportObject in Side-by-Side Subreports.frx")
	}
}
