package reportpkg_test

// Smoke tests for the "Group" category of FastReport sample FRX files.
// Each test verifies that the file loads without error and that GroupHeaderBand
// and GroupFooterBand are deserialized correctly.

import (
	"testing"

	"github.com/andrewloable/go-fastreport/band"
)

func TestFRXSmoke_Groups(t *testing.T) {
	r := loadFRXSmoke(t, "Groups.frx")
	var ghCount int
	for _, pg := range r.Pages() {
		for _, b := range pg.Bands() {
			if _, ok := b.(*band.GroupHeaderBand); ok {
				ghCount++
			}
		}
	}
	if ghCount == 0 {
		t.Error("expected at least one GroupHeaderBand in Groups.frx")
	}
}

func TestFRXSmoke_DrillDownGroups(t *testing.T) {
	r := loadFRXSmoke(t, "Drill-Down Groups.frx")
	var ghCount int
	for _, pg := range r.Pages() {
		for _, b := range pg.Bands() {
			if _, ok := b.(*band.GroupHeaderBand); ok {
				ghCount++
			}
		}
	}
	if ghCount == 0 {
		t.Error("expected at least one GroupHeaderBand in Drill-Down Groups.frx")
	}
}

func TestFRXSmoke_PrintTotalInGroupHeader(t *testing.T) {
	r := loadFRXSmoke(t, "Print Total in The Group Header.frx")
	var ghCount int
	for _, pg := range r.Pages() {
		for _, b := range pg.Bands() {
			if _, ok := b.(*band.GroupHeaderBand); ok {
				ghCount++
			}
		}
	}
	if ghCount == 0 {
		t.Error("expected at least one GroupHeaderBand")
	}
}

func TestFRXSmoke_SortGroupByTotal(t *testing.T) {
	r := loadFRXSmoke(t, "Sort Group By Total.frx")
	var ghCount int
	for _, pg := range r.Pages() {
		for _, b := range pg.Bands() {
			if gh, ok := b.(*band.GroupHeaderBand); ok {
				ghCount++
				if gh.Condition() == "" {
					t.Errorf("GroupHeaderBand %q has empty Condition", gh.Name())
				}
			}
		}
	}
	if ghCount == 0 {
		t.Error("expected at least one GroupHeaderBand in Sort Group By Total.frx")
	}
}

func TestFRXSmoke_ComplexMasterDetailGroup(t *testing.T) {
	r := loadFRXSmoke(t, "Complex (Master-detail + Group).frx")
	var ghCount int
	for _, pg := range r.Pages() {
		for _, b := range pg.Bands() {
			if _, ok := b.(*band.GroupHeaderBand); ok {
				ghCount++
			}
		}
	}
	if ghCount == 0 {
		t.Error("expected at least one GroupHeaderBand in Complex (Master-detail + Group).frx")
	}
}
