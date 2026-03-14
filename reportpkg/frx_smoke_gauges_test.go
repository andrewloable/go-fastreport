package reportpkg_test

// Smoke tests for Gauge FRX reports.
// Gauge element types (LinearGauge, RadialGauge, SimpleGauge, SimpleProgressGauge)
// are registered in the serial registry and deserialized from FRX files.
// These tests verify that the FRX files load without panic and have at least one page.

import (
	"testing"

	"github.com/andrewloable/go-fastreport/gauge"
)

func TestFRXSmoke_Gauge(t *testing.T) {
	r := loadFRXSmoke(t, "Gauge.frx")
	// Verify at least one LinearGauge was deserialized.
	n := countObjectsOfType[*gauge.LinearGauge](r)
	if n == 0 {
		t.Error("expected at least one LinearGauge in Gauge.frx")
	}
}

func TestFRXSmoke_AdvMatrixSparklineGauge(t *testing.T) {
	loadFRXSmoke(t, "AdvMatrix - Sparkline, Gauge.frx")
}
