package object

// advmatrix_internal_coverage_test.go — internal tests to cover error branches
// in AdvMatrixObject.Serialize (lines 119–129 of advmatrix.go).
//
// The WriteObjectNamed error paths cannot be reached via the public serial.Writer
// because it never returns errors.  This internal test package uses a mock
// report.Writer that always errors on WriteObjectNamed so that both the
// TableColumn loop (line 120) and the TableRow loop (line 126) return-on-error
// branches are executed and counted by the coverage tool.

import (
	"errors"
	"testing"

	"github.com/andrewloable/go-fastreport/report"
)

// errWriter is a minimal mock of report.Writer whose WriteObjectNamed always
// returns a sentinel error.  All other methods are no-ops.
type errWriter struct {
	errMsg string
}

func (e *errWriter) WriteStr(name, value string)       {}
func (e *errWriter) WriteInt(name string, value int)   {}
func (e *errWriter) WriteBool(name string, value bool) {}
func (e *errWriter) WriteFloat(name string, value float32) {}
func (e *errWriter) WriteObject(obj report.Serializable) error { return nil }
func (e *errWriter) WriteObjectNamed(name string, obj report.Serializable) error {
	return errors.New(e.errMsg)
}

// TestAdvMatrixObject_Serialize_WriteObjectNamedErrorOnTableColumn verifies that
// AdvMatrixObject.Serialize returns the error from WriteObjectNamed when iterating
// the TableColumns slice (the first WriteObjectNamed call in Serialize).
func TestAdvMatrixObject_Serialize_WriteObjectNamedErrorOnTableColumn(t *testing.T) {
	a := NewAdvMatrixObject()
	a.TableColumns = append(a.TableColumns, &AdvMatrixColumn{Name: "C1", Width: 50})

	w := &errWriter{errMsg: "column write error"}
	err := a.Serialize(w)
	if err == nil {
		t.Fatal("expected error from Serialize when WriteObjectNamed fails for TableColumn, got nil")
	}
	if err.Error() != "column write error" {
		t.Errorf("unexpected error message: %q", err.Error())
	}
}

// TestAdvMatrixObject_Serialize_WriteObjectNamedErrorOnTableRow verifies that
// AdvMatrixObject.Serialize returns the error from WriteObjectNamed when iterating
// the TableRows slice.  TableColumns must be empty so the column loop succeeds
// (no-ops) and execution reaches the row loop.
func TestAdvMatrixObject_Serialize_WriteObjectNamedErrorOnTableRow(t *testing.T) {
	a := NewAdvMatrixObject()
	// No TableColumns — the column loop completes without calling WriteObjectNamed.
	a.TableRows = append(a.TableRows, &AdvMatrixRow{Name: "R1", Height: 20})

	w := &errWriter{errMsg: "row write error"}
	err := a.Serialize(w)
	if err == nil {
		t.Fatal("expected error from Serialize when WriteObjectNamed fails for TableRow, got nil")
	}
	if err.Error() != "row write error" {
		t.Errorf("unexpected error message: %q", err.Error())
	}
}
