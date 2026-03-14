package reportpkg_test

// Smoke tests for Chart FRX reports.
// MSChartObject is registered in the serial registry.
// These tests verify that Chart FRX files load without panic and contain chart objects.

import (
	"testing"

	"github.com/andrewloable/go-fastreport/object"
)

func TestFRXSmoke_Chart(t *testing.T) {
	r := loadFRXSmoke(t, "Chart.frx")
	if n := countObjectsOfType[*object.MSChartObject](r); n == 0 {
		t.Error("expected at least one MSChartObject in Chart.frx")
	}
}

func TestFRXSmoke_MicrosoftChartSample(t *testing.T) {
	loadFRXSmoke(t, "Microsoft Chart Sample.frx")
}

func TestFRXSmoke_SeriesOfChart(t *testing.T) {
	loadFRXSmoke(t, "Series of Chart.frx")
}

func TestFRXSmoke_InteractiveChart(t *testing.T) {
	loadFRXSmoke(t, "Interactive Chart.frx")
}

func TestFRXSmoke_InteractiveMatrixWithChart(t *testing.T) {
	loadFRXSmoke(t, "Interactive Matrix With Chart.frx")
}
