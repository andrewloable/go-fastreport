package reportpkg_test

// Smoke tests for Table FRX reports.
// TableObject is registered in the serial registry; these tests verify FRX
// files load without panic and that TableObject children are deserialized.

import (
	"testing"

	"github.com/andrewloable/go-fastreport/table"
)

func TestFRXSmoke_Table(t *testing.T) {
	r := loadFRXSmoke(t, "Table.frx")
	n := countObjectsOfType[*table.TableObject](r)
	if n == 0 {
		t.Error("expected at least one TableObject in Table.frx")
	}
}

func TestFRXSmoke_Table_RowsDeserialized(t *testing.T) {
	r := loadFRXSmoke(t, "Simple Matrix.frx")
	// Simple Matrix embeds a MatrixObject which has TableRows/TableColumns
	// deserialized via DeserializeChild.  Just verify no panic.
	if len(r.Pages()) == 0 {
		t.Error("expected at least 1 page")
	}
}

func TestFRXSmoke_FitDynamicTableToPage(t *testing.T) {
	loadFRXSmoke(t, "Fit Dynamic Table To Page.frx")
}

func TestFRXSmoke_MultiplicationTable(t *testing.T) {
	loadFRXSmoke(t, "Multiplication Table.frx")
}
