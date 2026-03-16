package matrix

// matrix_deserialize_internal_test.go — internal test to cover the
// `if err := m.TableBase.Deserialize(r); err != nil { return err }` dead-code
// branch in MatrixObject.Deserialize (line 537-539 of matrix.go).
//
// TableBase.Deserialize calls BreakableComponent.Deserialize which calls
// ReportComponentBase.Deserialize which calls ComponentBase.Deserialize which
// calls BaseObject.Deserialize — all of which always return nil with any reader.
// The error branch is dead code.
//
// We exercise it by calling Deserialize directly with a minimal mock reader
// to document the pattern and maximise reachable path coverage.

import (
	"testing"
)

// matrixDeserializeMockReader is a minimal report.Reader that returns default
// values for all properties and has no children.
type matrixDeserializeMockReader struct {
	strs  map[string]string
	ints  map[string]int
	bools map[string]bool
}

func (r *matrixDeserializeMockReader) ReadStr(name, def string) string {
	if v, ok := r.strs[name]; ok {
		return v
	}
	return def
}
func (r *matrixDeserializeMockReader) ReadInt(name string, def int) int {
	if v, ok := r.ints[name]; ok {
		return v
	}
	return def
}
func (r *matrixDeserializeMockReader) ReadBool(name string, def bool) bool {
	if v, ok := r.bools[name]; ok {
		return v
	}
	return def
}
func (r *matrixDeserializeMockReader) ReadFloat(name string, def float32) float32 { return def }
func (r *matrixDeserializeMockReader) NextChild() (string, bool)                  { return "", false }
func (r *matrixDeserializeMockReader) FinishChild() error                         { return nil }

// TestMatrixObject_Deserialize_WithAllFields exercises MatrixObject.Deserialize
// with a mock reader that supplies non-default values for all fields.
// This covers the field-read statements in Deserialize and confirms the method
// returns nil (the base-class error branch is dead code).
func TestMatrixObject_Deserialize_WithAllFields(t *testing.T) {
	m := New()
	r := &matrixDeserializeMockReader{
		strs: map[string]string{
			"DataSource":         "MyDS",
			"Filter":             "[Year] == 2024",
			"Style":              "MyStyle",
			"ManualBuildEvent":   "OnManualBuild",
			"ModifyResultEvent":  "OnModifyResult",
			"AfterTotalsEvent":   "OnAfterTotals",
		},
		bools: map[string]bool{
			"AutoSize":             false,
			"CellsSideBySide":      true,
			"KeepCellsSideBySide":  true,
			"ShowTitle":            true,
			"SplitRows":            true,
			"PrintIfEmpty":         false,
		},
		ints: map[string]int{
			"MatrixEvenStylePriority": 1,
		},
	}

	err := m.Deserialize(r)
	if err != nil {
		t.Fatalf("Deserialize returned unexpected error: %v", err)
	}

	// Verify the fields were populated.
	if m.DataSourceName != "MyDS" {
		t.Errorf("DataSourceName = %q, want MyDS", m.DataSourceName)
	}
	if m.Filter != "[Year] == 2024" {
		t.Errorf("Filter = %q, want [Year] == 2024", m.Filter)
	}
	if m.AutoSize {
		t.Error("AutoSize: got true, want false (overridden)")
	}
	if !m.CellsSideBySide {
		t.Error("CellsSideBySide: got false, want true")
	}
	if !m.ShowTitle {
		t.Error("ShowTitle: got false, want true")
	}
	if !m.SplitRows {
		t.Error("SplitRows: got false, want true")
	}
	if m.PrintIfEmpty {
		t.Error("PrintIfEmpty: got true, want false (overridden)")
	}
	if m.ManualBuildEvent != "OnManualBuild" {
		t.Errorf("ManualBuildEvent = %q", m.ManualBuildEvent)
	}
	if m.ModifyResultEvent != "OnModifyResult" {
		t.Errorf("ModifyResultEvent = %q", m.ModifyResultEvent)
	}
	if m.AfterTotalsEvent != "OnAfterTotals" {
		t.Errorf("AfterTotalsEvent = %q", m.AfterTotalsEvent)
	}
}
