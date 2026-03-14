package reportpkg_test

// Smoke tests for Table FRX reports.
// TableObject is a complex container; these tests verify FRX files load
// without panic and produce at least one page.

import (
	"testing"
)

func TestFRXSmoke_Table(t *testing.T) {
	loadFRXSmoke(t, "Table.frx")
}

func TestFRXSmoke_FitDynamicTableToPage(t *testing.T) {
	loadFRXSmoke(t, "Fit Dynamic Table To Page.frx")
}

func TestFRXSmoke_MultiplicationTable(t *testing.T) {
	loadFRXSmoke(t, "Multiplication Table.frx")
}
