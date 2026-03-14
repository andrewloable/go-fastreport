package reportpkg_test

// Smoke tests for expression/conditional formatting FRX reports.

import (
	"testing"
)

func TestFRXSmoke_UsingExpressions(t *testing.T) {
	loadFRXSmoke(t, "Using Expressions.frx")
}

func TestFRXSmoke_Highlight(t *testing.T) {
	loadFRXSmoke(t, "Highlight.frx")
}

func TestFRXSmoke_HighlightBasedOnRowColumn(t *testing.T) {
	loadFRXSmoke(t, "Highlight Based on Row-Column.frx")
}

func TestFRXSmoke_AlternateColorEachRow(t *testing.T) {
	loadFRXSmoke(t, "Alternate Color Each Row.frx")
}

func TestFRXSmoke_OddEvenRows(t *testing.T) {
	loadFRXSmoke(t, "Odd-Even Rows.frx")
}

func TestFRXSmoke_OddEvenPagesMirrorMargins(t *testing.T) {
	r := loadFRXSmoke(t, "Odd-Even Pages, Mirror Margins.frx")
	// Verify mirror margins was deserialized.
	for _, pg := range r.Pages() {
		if pg.MirrorMargins {
			return
		}
	}
	t.Error("expected MirrorMargins=true in Odd-Even Pages, Mirror Margins.frx")
}

func TestFRXSmoke_DuplicateValues(t *testing.T) {
	loadFRXSmoke(t, "Duplicate Values.frx")
}
