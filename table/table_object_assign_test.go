package table

import (
	"testing"
)

// TestTableObject_Assign verifies that Assign copies TableObject-specific
// fields from src into dst.
//
// C# reference: TableObject.Assign (TableObject.cs line 226-233).
func TestTableObject_Assign(t *testing.T) {
	src := NewTableObject()
	src.ManualBuildAutoSpans = false
	called := false
	src.ManualBuild = func(h *TableHelper) { called = true }

	dst := NewTableObject()
	dst.Assign(src)

	if dst.ManualBuildAutoSpans != false {
		t.Errorf("ManualBuildAutoSpans: got %v, want false", dst.ManualBuildAutoSpans)
	}
	if dst.ManualBuild == nil {
		t.Error("ManualBuild: got nil, want non-nil callback")
	} else {
		// Invoke to confirm it's the same func.
		dst.ManualBuild(nil)
		if !called {
			t.Error("ManualBuild callback was not the same function after Assign")
		}
	}
}

// TestTableObject_Assign_NilSrc verifies that Assign with a nil source is a no-op.
func TestTableObject_Assign_NilSrc(t *testing.T) {
	dst := NewTableObject()
	dst.ManualBuildAutoSpans = true
	dst.Assign(nil) // must not panic
	if !dst.ManualBuildAutoSpans {
		t.Error("Assign(nil) must not modify receiver fields")
	}
}

// TestTableObject_Assign_TableBaseFields verifies that underlying TableBase
// fields (e.g. ManualBuildEvent) are also copied via the TableBase.Assign call.
func TestTableObject_Assign_TableBaseFields(t *testing.T) {
	src := NewTableObject()
	src.ManualBuildEvent = "OnManualBuild"

	dst := NewTableObject()
	dst.Assign(src)

	if dst.ManualBuildEvent != "OnManualBuild" {
		t.Errorf("ManualBuildEvent: got %q, want %q", dst.ManualBuildEvent, "OnManualBuild")
	}
}

// TestTableObject_GetExpressions verifies that GetExpressions does not panic
// and returns an empty (possibly nil) slice for a freshly created TableObject
// that has no expressions set.
//
// C# reference: TableCell.GetExpressions (TableCell.cs line 336-350).
func TestTableObject_GetExpressions(t *testing.T) {
	to := NewTableObject()
	exprs := to.GetExpressions()
	// A fresh TableObject has no expressions; nil and empty are both acceptable.
	if len(exprs) != 0 {
		t.Errorf("GetExpressions on empty TableObject: got %v, want empty", exprs)
	}
}

// TestTableObject_GetExpressions_WithCells verifies that GetExpressions
// collects expressions from cells when the table has rows and cells with
// visible expressions set.
func TestTableObject_GetExpressions_WithCells(t *testing.T) {
	to := NewTableObject()

	// Add one row with two cells.
	row := NewTableRow()
	cell1 := NewTableCell()
	cell1.SetVisibleExpression("[Report.PageNo > 1]")
	cell2 := NewTableCell()
	cell2.SetVisibleExpression("[Employee.HireDate > Date(2000,1,1)]")
	row.cells = append(row.cells, cell1, cell2)
	to.rows = append(to.rows, row)

	exprs := to.GetExpressions()

	if len(exprs) < 2 {
		t.Errorf("GetExpressions: got %d expression(s), want at least 2; exprs=%v", len(exprs), exprs)
	}
}
