package table_test

import (
	"testing"

	"github.com/andrewloable/go-fastreport/object"
	"github.com/andrewloable/go-fastreport/table"
)

// ── TableBase.Styles() ────────────────────────────────────────────────────────

// TestTableBase_Styles_NotNil verifies that a newly created TableBase always
// owns a non-nil TableStyleCollection, matching C# TableBase constructor
// (TableBase.cs line 1388).
func TestTableBase_Styles_NotNil(t *testing.T) {
	tb := table.NewTableBase()
	if tb.Styles() == nil {
		t.Fatal("NewTableBase().Styles() returned nil")
	}
}

// TestTableBase_Styles_InitiallyEmpty verifies that the style collection starts
// empty with a valid default style.
func TestTableBase_Styles_InitiallyEmpty(t *testing.T) {
	tb := table.NewTableBase()
	sc := tb.Styles()
	if sc.Count() != 0 {
		t.Errorf("initial Count: got %d, want 0", sc.Count())
	}
	if sc.DefaultStyle() == nil {
		t.Error("DefaultStyle() should not be nil")
	}
}

// TestTableObject_Styles_NotNil verifies that TableObject (which embeds
// TableBase) also owns a valid style collection.
func TestTableObject_Styles_NotNil(t *testing.T) {
	to := table.NewTableObject()
	if to.Styles() == nil {
		t.Fatal("NewTableObject().Styles() returned nil")
	}
}

// ── TableCell.SetStyle() ──────────────────────────────────────────────────────

// TestTableCell_SetStyle_NilTable verifies that SetStyle with a nil table
// returns the style argument unchanged.
func TestTableCell_SetStyle_NilTable(t *testing.T) {
	c := table.NewTableCell()
	style := table.NewTableCell()
	got := c.SetStyle(nil, style)
	if got != style {
		t.Error("SetStyle(nil, style) should return the style argument unchanged")
	}
}

// TestTableCell_SetStyle_AddsToCollection verifies that calling SetStyle
// with a table causes the style to be added to the table's Styles collection,
// matching C# TableCellData.SetStyle → Table.Styles.Add (TableCellData.cs
// line 328).
func TestTableCell_SetStyle_AddsToCollection(t *testing.T) {
	tb := table.NewTableBase()
	c := table.NewTableCell()
	style := table.NewTableCell()

	c.SetStyle(tb, style)

	if tb.Styles().Count() != 1 {
		t.Errorf("Styles().Count() after SetStyle: got %d, want 1", tb.Styles().Count())
	}
}

// TestTableCell_SetStyle_Deduplication verifies that calling SetStyle with
// two cells that have the same visual style results in only one entry in the
// collection (deduplication).
func TestTableCell_SetStyle_Deduplication(t *testing.T) {
	tb := table.NewTableBase()
	c1 := table.NewTableCell()
	c2 := table.NewTableCell()

	style1 := table.NewTableCell()
	style2 := table.NewTableCell()
	// Both styles use default appearance → they should be deduplicated.

	ref1 := c1.SetStyle(tb, style1)
	ref2 := c2.SetStyle(tb, style2)

	if tb.Styles().Count() != 1 {
		t.Errorf("Count after two identical styles: got %d, want 1", tb.Styles().Count())
	}
	if ref1 != ref2 {
		t.Error("SetStyle should return the same canonical reference for equal styles")
	}
}

// TestTableCell_SetStyle_DifferentStylesNotDeduplicated verifies that two
// cells with different visual styles each get their own entry in the
// collection.
func TestTableCell_SetStyle_DifferentStylesNotDeduplicated(t *testing.T) {
	tb := table.NewTableBase()

	style1 := table.NewTableCell()
	style1.SetHorzAlign(object.HorzAlignLeft)

	style2 := table.NewTableCell()
	style2.SetHorzAlign(object.HorzAlignCenter) // different alignment

	c := table.NewTableCell()
	c.SetStyle(tb, style1)
	c.SetStyle(tb, style2)

	if tb.Styles().Count() != 2 {
		t.Errorf("Count after two different styles: got %d, want 2", tb.Styles().Count())
	}
}

// TestTableCell_SetStyle_ReturnsCanonical verifies that the return value of
// SetStyle is the canonical deduplicated instance from the collection.
func TestTableCell_SetStyle_ReturnsCanonical(t *testing.T) {
	tb := table.NewTableBase()
	style := table.NewTableCell()

	canonical := table.NewTableCell().SetStyle(tb, style)
	// The canonical is from the collection — Get(0) must return the same pointer.
	if tb.Styles().Get(0) != canonical {
		t.Error("SetStyle return value should equal Styles().Get(0)")
	}
}

// TestTableCell_SetStyle_MultipleCallsSameTable verifies that multiple
// SetStyle calls with different styles on the same table accumulate
// independently.
func TestTableCell_SetStyle_MultipleCallsSameTable(t *testing.T) {
	tb := table.NewTableBase()
	c := table.NewTableCell()

	aligns := []object.HorzAlign{
		object.HorzAlignLeft,
		object.HorzAlignCenter,
		object.HorzAlignRight,
	}
	for _, align := range aligns {
		style := table.NewTableCell()
		style.SetHorzAlign(align)
		c.SetStyle(tb, style)
	}

	if tb.Styles().Count() != 3 {
		t.Errorf("Styles().Count(): got %d, want 3", tb.Styles().Count())
	}
}
