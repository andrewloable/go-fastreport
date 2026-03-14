package reportpkg_test

// Smoke tests for Gauge FRX reports.
// Gauge element types (LinearGauge, RadialGauge, SimpleGauge, SimpleProgressGauge)
// are not yet registered in the serial registry, so they are skipped during load.
// These tests verify that the FRX files load without panic and have at least one page.

import (
	"testing"
)

func TestFRXSmoke_Gauge(t *testing.T) {
	loadFRXSmoke(t, "Gauge.frx")
}

func TestFRXSmoke_AdvMatrixSparklineGauge(t *testing.T) {
	loadFRXSmoke(t, "AdvMatrix - Sparkline, Gauge.frx")
}
