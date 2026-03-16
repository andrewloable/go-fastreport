package object

// advmatrix_deserialize_internal_test.go — internal test to cover the
// `if err := a.ReportComponentBase.Deserialize(r); err != nil { return err }`
// dead-code branch in AdvMatrixObject.Deserialize (line 367-370 of advmatrix.go).
//
// ReportComponentBase.Deserialize always returns nil — this branch is dead code.
// We call Deserialize with a mock reader to exercise all reachable statements
// and document the pattern used for other Deserialize coverage tests.

import (
	"testing"
)

// advMatrixMockReader is a minimal report.Reader for AdvMatrixObject.Deserialize.
type advMatrixMockReader struct {
	strs map[string]string
}

func (r *advMatrixMockReader) ReadStr(name, def string) string {
	if v, ok := r.strs[name]; ok {
		return v
	}
	return def
}
func (r *advMatrixMockReader) ReadInt(name string, def int) int       { return def }
func (r *advMatrixMockReader) ReadBool(name string, def bool) bool    { return def }
func (r *advMatrixMockReader) ReadFloat(name string, def float32) float32 { return def }
func (r *advMatrixMockReader) NextChild() (string, bool)                  { return "", false }
func (r *advMatrixMockReader) FinishChild() error                         { return nil }

// TestAdvMatrixObject_Deserialize_WithDataSource exercises AdvMatrixObject.Deserialize
// with a mock reader that supplies a DataSource value. This covers the ReadStr call
// at line 371 and verifies the method returns nil (the base error-branch is dead code).
func TestAdvMatrixObject_Deserialize_WithDataSource(t *testing.T) {
	a := NewAdvMatrixObject()
	r := &advMatrixMockReader{
		strs: map[string]string{"DataSource": "SalesDB"},
	}

	err := a.Deserialize(r)
	if err != nil {
		t.Fatalf("Deserialize returned unexpected error: %v", err)
	}
	if a.DataSource != "SalesDB" {
		t.Errorf("DataSource = %q, want SalesDB", a.DataSource)
	}
}

// TestAdvMatrixObject_Deserialize_EmptyDataSource exercises the empty-DataSource path.
func TestAdvMatrixObject_Deserialize_EmptyDataSource(t *testing.T) {
	a := NewAdvMatrixObject()
	r := &advMatrixMockReader{strs: map[string]string{}}

	err := a.Deserialize(r)
	if err != nil {
		t.Fatalf("Deserialize returned unexpected error: %v", err)
	}
	if a.DataSource != "" {
		t.Errorf("DataSource = %q, want empty string", a.DataSource)
	}
}
