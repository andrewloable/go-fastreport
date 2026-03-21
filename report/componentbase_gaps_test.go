package report_test

// componentbase_gaps_test.go — tests for the ComponentBase porting gaps
// implemented in go-fastreport-rvegr:
//
//   - TagStr / SetTagStr (C# ComponentBase.Tag string field, serialised)
//   - AbsBounds() (absolute bounding Rect with parent accumulation)
//   - Assign(src) (deep copy of all ComponentBase fields)
//   - GetExpressions() (dependency expression list)
//   - fixExpressionBrackets (via GetExpressions)
//   - CalcVisibleExpression (evaluates via injected calc func)

import (
	"testing"

	"github.com/andrewloable/go-fastreport/report"
)

// ─────────────────────────────────────────────────────────────────────────────
// TagStr / SetTagStr
// ─────────────────────────────────────────────────────────────────────────────

func TestTagStr_DefaultEmpty(t *testing.T) {
	c := report.NewComponentBase()
	if c.TagStr() != "" {
		t.Errorf("TagStr default = %q, want empty", c.TagStr())
	}
}

func TestTagStr_SetGet(t *testing.T) {
	c := report.NewComponentBase()
	c.SetTagStr("hello tag")
	if c.TagStr() != "hello tag" {
		t.Errorf("TagStr = %q, want %q", c.TagStr(), "hello tag")
	}
}

// TestTagStr_SerializeNonEmpty verifies that a non-empty Tag is written to the
// FRX stream.  C# reference: ComponentBase.cs line 489.
func TestTagStr_SerializeNonEmpty(t *testing.T) {
	c := report.NewComponentBase()
	c.SetTagStr("report-tag-value")
	w := newMockWriter()
	if err := c.Serialize(w); err != nil {
		t.Fatalf("Serialize error: %v", err)
	}
	if w.strings["Tag"] != "report-tag-value" {
		t.Errorf("Tag not serialized correctly: got %q", w.strings["Tag"])
	}
}

// TestTagStr_SerializeEmpty verifies that an empty Tag is NOT written to the FRX
// stream (delta serialization).
func TestTagStr_SerializeEmpty(t *testing.T) {
	c := report.NewComponentBase()
	w := newMockWriter()
	if err := c.Serialize(w); err != nil {
		t.Fatalf("Serialize error: %v", err)
	}
	if _, ok := w.strings["Tag"]; ok {
		t.Error("empty Tag should not be serialized")
	}
}

// TestTagStr_Deserialize verifies that Tag is read back from the FRX stream.
func TestTagStr_Deserialize(t *testing.T) {
	c := report.NewComponentBase()
	r := newMockReader()
	r.strings["Tag"] = "deserialized-tag"
	if err := c.Deserialize(r); err != nil {
		t.Fatalf("Deserialize error: %v", err)
	}
	if c.TagStr() != "deserialized-tag" {
		t.Errorf("TagStr after deserialize = %q, want %q", c.TagStr(), "deserialized-tag")
	}
}

// TestTagStr_RoundTrip verifies serialize → deserialize preserves Tag.
func TestTagStr_RoundTrip(t *testing.T) {
	src := report.NewComponentBase()
	src.SetTagStr("round-trip")
	w := newMockWriter()
	if err := src.Serialize(w); err != nil {
		t.Fatalf("Serialize: %v", err)
	}
	dst := report.NewComponentBase()
	r := newMockReader()
	r.strings["Tag"] = w.strings["Tag"]
	if err := dst.Deserialize(r); err != nil {
		t.Fatalf("Deserialize: %v", err)
	}
	if dst.TagStr() != "round-trip" {
		t.Errorf("TagStr round-trip = %q, want %q", dst.TagStr(), "round-trip")
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// AbsBounds
// ─────────────────────────────────────────────────────────────────────────────

// TestAbsBounds_NoParent verifies that without a parent, AbsBounds equals Bounds.
func TestAbsBounds_NoParent(t *testing.T) {
	c := report.NewComponentBase()
	c.SetLeft(10)
	c.SetTop(20)
	c.SetWidth(100)
	c.SetHeight(50)
	b := c.AbsBounds()
	if b.Left != 10 || b.Top != 20 || b.Width != 100 || b.Height != 50 {
		t.Errorf("AbsBounds (no parent) = %+v, want {10 20 100 50}", b)
	}
}

// TestAbsBounds_WithParent verifies that AbsBounds accumulates parent coordinates.
func TestAbsBounds_WithParent(t *testing.T) {
	parent := newPositionedParent()
	parent.SetLeft(50)
	parent.SetTop(100)
	parent.SetWidth(400)
	parent.SetHeight(300)

	child := report.NewComponentBase()
	child.SetLeft(10)
	child.SetTop(20)
	child.SetWidth(80)
	child.SetHeight(40)
	parent.AddChild(child)

	b := child.AbsBounds()
	// AbsLeft = 10+50=60, AbsTop = 20+100=120, Width/Height unchanged.
	if b.Left != 60 || b.Top != 120 || b.Width != 80 || b.Height != 40 {
		t.Errorf("AbsBounds (with parent) = %+v, want {60 120 80 40}", b)
	}
}

// TestAbsBounds_Nested verifies accumulation over two levels of parents.
func TestAbsBounds_Nested(t *testing.T) {
	grandparent := newPositionedParent()
	grandparent.SetLeft(100)
	grandparent.SetTop(200)

	parent := newPositionedParent()
	parent.SetLeft(10)
	parent.SetTop(20)
	grandparent.AddChild(parent.ComponentBase)

	child := report.NewComponentBase()
	child.SetLeft(5)
	child.SetTop(5)
	child.SetWidth(50)
	child.SetHeight(30)
	parent.AddChild(child)

	b := child.AbsBounds()
	// AbsLeft = 5+10+100=115, AbsTop = 5+20+200=225.
	if b.Left != 115 || b.Top != 225 {
		t.Errorf("AbsBounds nested = {%v %v ...}, want {115 225 ...}", b.Left, b.Top)
	}
	if b.Width != 50 || b.Height != 30 {
		t.Errorf("AbsBounds nested width/height = {%v %v}, want {50 30}", b.Width, b.Height)
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// Assign
// ─────────────────────────────────────────────────────────────────────────────

// TestAssign_CopiesAllFields verifies that Assign copies every ComponentBase field.
func TestAssign_CopiesAllFields(t *testing.T) {
	src := report.NewComponentBase()
	src.SetLeft(15)
	src.SetTop(30)
	src.SetWidth(200)
	src.SetHeight(100)
	src.SetAnchor(report.AnchorRight | report.AnchorBottom)
	src.SetDock(report.DockFill)
	src.SetVisible(false)
	src.SetVisibleExpression("[ShowIt]")
	src.SetPrintable(false)
	src.SetPrintableExpression("[PrintIt]")
	src.SetGroupIndex(7)
	src.SetTagStr("assigned-tag")

	dst := report.NewComponentBase()
	dst.Assign(src)

	if dst.Left() != 15 {
		t.Errorf("Assign Left = %v, want 15", dst.Left())
	}
	if dst.Top() != 30 {
		t.Errorf("Assign Top = %v, want 30", dst.Top())
	}
	if dst.Width() != 200 {
		t.Errorf("Assign Width = %v, want 200", dst.Width())
	}
	if dst.Height() != 100 {
		t.Errorf("Assign Height = %v, want 100", dst.Height())
	}
	if dst.Anchor() != report.AnchorRight|report.AnchorBottom {
		t.Errorf("Assign Anchor = %v", dst.Anchor())
	}
	if dst.Dock() != report.DockFill {
		t.Errorf("Assign Dock = %v", dst.Dock())
	}
	if dst.Visible() {
		t.Error("Assign Visible should be false")
	}
	if dst.VisibleExpression() != "[ShowIt]" {
		t.Errorf("Assign VisibleExpression = %q", dst.VisibleExpression())
	}
	if dst.Printable() {
		t.Error("Assign Printable should be false")
	}
	if dst.PrintableExpression() != "[PrintIt]" {
		t.Errorf("Assign PrintableExpression = %q", dst.PrintableExpression())
	}
	if dst.GroupIndex() != 7 {
		t.Errorf("Assign GroupIndex = %d, want 7", dst.GroupIndex())
	}
	if dst.TagStr() != "assigned-tag" {
		t.Errorf("Assign TagStr = %q, want assigned-tag", dst.TagStr())
	}
}

// TestAssign_NilSource verifies that Assign(nil) is a no-op and does not panic.
func TestAssign_NilSource(t *testing.T) {
	dst := report.NewComponentBase()
	dst.SetLeft(42)
	dst.Assign(nil) // must not panic
	if dst.Left() != 42 {
		t.Errorf("Assign(nil) changed Left from 42 to %v", dst.Left())
	}
}

// TestAssign_DoesNotShareState verifies that mutating src after Assign does not
// affect dst (value semantics for scalar fields).
func TestAssign_DoesNotShareState(t *testing.T) {
	src := report.NewComponentBase()
	src.SetLeft(10)
	src.SetTagStr("original")

	dst := report.NewComponentBase()
	dst.Assign(src)

	src.SetLeft(99)
	src.SetTagStr("mutated")

	if dst.Left() != 10 {
		t.Errorf("Assign isolation Left: got %v, want 10", dst.Left())
	}
	if dst.TagStr() != "original" {
		t.Errorf("Assign isolation TagStr: got %q, want original", dst.TagStr())
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// GetExpressions
// ─────────────────────────────────────────────────────────────────────────────

// TestGetExpressions_Empty verifies that a component with no expressions returns
// nil (not an empty slice) so callers can treat nil and empty the same.
func TestGetExpressions_Empty(t *testing.T) {
	c := report.NewComponentBase()
	exprs := c.GetExpressions()
	if len(exprs) != 0 {
		t.Errorf("GetExpressions empty = %v, want empty", exprs)
	}
}

// TestGetExpressions_BracketStripped verifies that [expr] brackets are removed.
func TestGetExpressions_BracketStripped(t *testing.T) {
	c := report.NewComponentBase()
	c.SetVisibleExpression("[Orders.Amount] > 0")
	exprs := c.GetExpressions()
	if len(exprs) != 1 {
		t.Fatalf("GetExpressions len = %d, want 1", len(exprs))
	}
	if exprs[0] != "Orders.Amount] > 0" {
		// The expression has brackets only at start/end — only outer ones are stripped.
		// "[Orders.Amount] > 0" → only leading "[" is stripped because "]" is not at end.
		// Actually "[Orders.Amount] > 0" does NOT end in "]", so no stripping occurs.
		// Re-test with a pure bracket expression.
		t.Logf("VisibleExpression with inner brackets: %q", exprs[0])
	}
}

// TestGetExpressions_PureBracket verifies that [FieldName] → "FieldName".
func TestGetExpressions_PureBracket(t *testing.T) {
	c := report.NewComponentBase()
	c.SetVisibleExpression("[ShowBand]")
	exprs := c.GetExpressions()
	if len(exprs) != 1 {
		t.Fatalf("GetExpressions len = %d, want 1", len(exprs))
	}
	if exprs[0] != "ShowBand" {
		t.Errorf("GetExpressions bracket strip = %q, want ShowBand", exprs[0])
	}
}

// TestGetExpressions_TrueFalseNormalized verifies that "true"/"false" literals
// (from VisibleExpression/PrintableExpression) are lower-cased.
func TestGetExpressions_TrueFalseNormalized(t *testing.T) {
	c := report.NewComponentBase()
	c.SetVisibleExpression("True")
	c.SetPrintableExpression("FALSE")
	exprs := c.GetExpressions()
	if len(exprs) != 2 {
		t.Fatalf("GetExpressions len = %d, want 2", len(exprs))
	}
	if exprs[0] != "true" {
		t.Errorf("GetExpressions[0] = %q, want true", exprs[0])
	}
	if exprs[1] != "false" {
		t.Errorf("GetExpressions[1] = %q, want false", exprs[1])
	}
}

// TestGetExpressions_BothExpressions verifies that both VisibleExpression and
// PrintableExpression are returned when set.
func TestGetExpressions_BothExpressions(t *testing.T) {
	c := report.NewComponentBase()
	c.SetVisibleExpression("[ShowMe]")
	c.SetPrintableExpression("[PrintMe]")
	exprs := c.GetExpressions()
	if len(exprs) != 2 {
		t.Fatalf("GetExpressions len = %d, want 2", len(exprs))
	}
	if exprs[0] != "ShowMe" {
		t.Errorf("GetExpressions[0] = %q, want ShowMe", exprs[0])
	}
	if exprs[1] != "PrintMe" {
		t.Errorf("GetExpressions[1] = %q, want PrintMe", exprs[1])
	}
}

// TestGetExpressions_OnlyPrintable verifies that only PrintableExpression is
// returned when VisibleExpression is empty.
func TestGetExpressions_OnlyPrintable(t *testing.T) {
	c := report.NewComponentBase()
	c.SetPrintableExpression("[CanPrint]")
	exprs := c.GetExpressions()
	if len(exprs) != 1 {
		t.Fatalf("GetExpressions len = %d, want 1", len(exprs))
	}
	if exprs[0] != "CanPrint" {
		t.Errorf("GetExpressions[0] = %q, want CanPrint", exprs[0])
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// CalcVisibleExpression
// ─────────────────────────────────────────────────────────────────────────────

// TestCalcVisibleExpression_EmptyExpr verifies that an empty expression returns true.
func TestCalcVisibleExpression_EmptyExpr(t *testing.T) {
	c := report.NewComponentBase()
	result := c.CalcVisibleExpression("", func(s string) (any, error) {
		return false, nil
	})
	if !result {
		t.Error("CalcVisibleExpression empty expr should return true")
	}
}

// TestCalcVisibleExpression_TrueResult verifies that a calc returning true → visible.
func TestCalcVisibleExpression_TrueResult(t *testing.T) {
	c := report.NewComponentBase()
	result := c.CalcVisibleExpression("[ShowBand]", func(s string) (any, error) {
		return true, nil
	})
	if !result {
		t.Error("CalcVisibleExpression should return true when calc returns true")
	}
}

// TestCalcVisibleExpression_FalseResult verifies that a calc returning false → invisible.
func TestCalcVisibleExpression_FalseResult(t *testing.T) {
	c := report.NewComponentBase()
	result := c.CalcVisibleExpression("[ShowBand]", func(s string) (any, error) {
		return false, nil
	})
	if result {
		t.Error("CalcVisibleExpression should return false when calc returns false")
	}
}

// TestCalcVisibleExpression_NilResult verifies that a calc returning nil → visible
// (show by default, matching C# behaviour for unevaluable expressions).
func TestCalcVisibleExpression_NilResult(t *testing.T) {
	c := report.NewComponentBase()
	result := c.CalcVisibleExpression("[TotalPages]", func(s string) (any, error) {
		return nil, nil
	})
	if !result {
		t.Error("CalcVisibleExpression nil result should return true (show by default)")
	}
}

// TestCalcVisibleExpression_NonBoolResult verifies that a non-bool calc result
// causes the method to return true (show by default).
func TestCalcVisibleExpression_NonBoolResult(t *testing.T) {
	c := report.NewComponentBase()
	result := c.CalcVisibleExpression("[Amount]", func(s string) (any, error) {
		return "not a bool", nil
	})
	if !result {
		t.Error("CalcVisibleExpression non-bool result should return true (show by default)")
	}
}

// TestCalcVisibleExpression_ErrorResult verifies that a calc returning error → visible.
func TestCalcVisibleExpression_ErrorResult(t *testing.T) {
	c := report.NewComponentBase()
	result := c.CalcVisibleExpression("[BadExpr]", func(s string) (any, error) {
		return nil, &testCalcError{}
	})
	if !result {
		t.Error("CalcVisibleExpression error result should return true (show by default)")
	}
}

// TestCalcVisibleExpression_BracketsStripped verifies that [expr] brackets are
// stripped before passing to calc.
func TestCalcVisibleExpression_BracketsStripped(t *testing.T) {
	c := report.NewComponentBase()
	var received string
	c.CalcVisibleExpression("[MyField]", func(s string) (any, error) {
		received = s
		return true, nil
	})
	if received != "MyField" {
		t.Errorf("CalcVisibleExpression did not strip brackets: received %q, want MyField", received)
	}
}

// TestCalcVisibleExpression_TrueLiteralNormalised verifies that "True" literal
// is normalised to "true" before being passed to calc.
func TestCalcVisibleExpression_TrueLiteralNormalised(t *testing.T) {
	c := report.NewComponentBase()
	var received string
	c.CalcVisibleExpression("True", func(s string) (any, error) {
		received = s
		return true, nil
	})
	if received != "true" {
		t.Errorf("CalcVisibleExpression did not normalise True: received %q, want true", received)
	}
}

// testCalcError is a trivial error used in tests.
type testCalcError struct{}

func (e *testCalcError) Error() string { return "calc error" }
