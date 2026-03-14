package reportpkg_test

// Smoke tests for Advanced Matrix FRX reports.
// AdvMatrix reports use MatrixObject which is not yet registered in the serial
// registry; these tests verify the files load without panic and have at least
// one page.

import (
	"testing"
)

func TestFRXSmoke_AdvMatrixCollapseSort(t *testing.T) {
	loadFRXSmoke(t, "AdvMatrix - Collapse + Sort.frx")
}

func TestFRXSmoke_AdvMatrixItemsComparison(t *testing.T) {
	loadFRXSmoke(t, "AdvMatrix - Items comparison.frx")
}

func TestFRXSmoke_AdvMatrixOrderDetails(t *testing.T) {
	loadFRXSmoke(t, "AdvMatrix - Order details.frx")
}

func TestFRXSmoke_AdvMatrixSortYearByTotal(t *testing.T) {
	loadFRXSmoke(t, "AdvMatrix - Sort Year by total.frx")
}

func TestFRXSmoke_AdvMatrixSteppedLayout(t *testing.T) {
	loadFRXSmoke(t, "AdvMatrix - Stepped layout.frx")
}

func TestFRXSmoke_AdvMatrixTwoDataCells(t *testing.T) {
	loadFRXSmoke(t, "AdvMatrix - Two data cells.frx")
}

func TestFRXSmoke_AdvMatrixUserFunction(t *testing.T) {
	loadFRXSmoke(t, "AdvMatrix - User function.frx")
}
