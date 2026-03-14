package reportpkg_test

// Smoke tests for business/sales FRX reports.

import (
	"testing"
)

func TestFRXSmoke_SalesInTheUSA(t *testing.T) {
	loadFRXSmoke(t, "Sales in the USA.frx")
}

func TestFRXSmoke_PrintEnteredValue(t *testing.T) {
	loadFRXSmoke(t, "Print Entered Value.frx")
}
