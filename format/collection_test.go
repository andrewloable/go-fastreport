package format

import (
	"testing"
)

// cloneableFormat is a Format implementation that supports Clone for Assign tests.
type cloneableFormat struct {
	label string
}

func (f *cloneableFormat) FormatType() string  { return "cloneable" }
func (f *cloneableFormat) FormatValue(v any) string { return f.label }
func (f *cloneableFormat) Clone() Format        { return &cloneableFormat{label: f.label} }

// simpleFormat is a Format without Clone for the else-branch in Assign.
type simpleFormat struct{}

func (f *simpleFormat) FormatType() string  { return "simple" }
func (f *simpleFormat) FormatValue(v any) string { return "simple" }

func TestCollection_AddAndCount(t *testing.T) {
	c := NewCollection()
	if c.Count() != 0 {
		t.Fatalf("empty collection count = %d, want 0", c.Count())
	}
	c.Add(NewGeneralFormat())
	c.Add(NewBooleanFormat())
	if c.Count() != 2 {
		t.Fatalf("count = %d, want 2", c.Count())
	}
}

func TestCollection_AddNil(t *testing.T) {
	c := NewCollection()
	idx := c.Add(nil)
	if idx != -1 {
		t.Errorf("Add(nil) = %d, want -1", idx)
	}
	if c.Count() != 0 {
		t.Error("Add(nil) should not change count")
	}
}

func TestCollection_Get(t *testing.T) {
	c := NewCollection()
	g := NewGeneralFormat()
	c.Add(g)
	if c.Get(0) != g {
		t.Error("Get(0) != added format")
	}
}

func TestCollection_Insert(t *testing.T) {
	c := NewCollection()
	g := NewGeneralFormat()
	b := NewBooleanFormat()
	c.Add(g)
	c.Add(b)
	custom := NewCustomFormat()
	c.Insert(1, custom)
	if c.Count() != 3 {
		t.Fatalf("count after Insert = %d, want 3", c.Count())
	}
	if c.Get(1) != custom {
		t.Error("Get(1) != inserted format")
	}
	if c.Get(2) != b {
		t.Error("Get(2) != original second format")
	}
}

func TestCollection_InsertNil(t *testing.T) {
	c := NewCollection()
	c.Add(NewGeneralFormat())
	c.Insert(0, nil)
	if c.Count() != 1 {
		t.Errorf("Insert(nil) changed count: %d", c.Count())
	}
}

func TestCollection_Remove(t *testing.T) {
	c := NewCollection()
	g := NewGeneralFormat()
	b := NewBooleanFormat()
	c.Add(g)
	c.Add(b)
	c.Remove(g)
	if c.Count() != 1 {
		t.Fatalf("count after Remove = %d, want 1", c.Count())
	}
	if c.Get(0) != b {
		t.Error("Get(0) after Remove != remaining format")
	}
}

func TestCollection_RemoveMissing(t *testing.T) {
	c := NewCollection()
	c.Add(NewGeneralFormat())
	c.Remove(NewBooleanFormat()) // not in collection
	if c.Count() != 1 {
		t.Error("Remove of missing item should not change count")
	}
}

func TestCollection_Clear(t *testing.T) {
	c := NewCollection()
	c.Add(NewGeneralFormat())
	c.Add(NewBooleanFormat())
	c.Clear()
	if c.Count() != 0 {
		t.Errorf("after Clear, count = %d, want 0", c.Count())
	}
}

func TestCollection_Contains(t *testing.T) {
	c := NewCollection()
	g := NewGeneralFormat()
	c.Add(g)
	if !c.Contains(g) {
		t.Error("Contains should return true for added format")
	}
	if c.Contains(NewBooleanFormat()) {
		t.Error("Contains should return false for absent format")
	}
}

func TestCollection_IndexOf(t *testing.T) {
	c := NewCollection()
	g := NewGeneralFormat()
	b := NewBooleanFormat()
	c.Add(g)
	c.Add(b)
	if c.IndexOf(b) != 1 {
		t.Errorf("IndexOf = %d, want 1", c.IndexOf(b))
	}
	if c.IndexOf(NewCustomFormat()) != -1 {
		t.Error("IndexOf for absent format should be -1")
	}
}

func TestCollection_Primary(t *testing.T) {
	c := NewCollection()
	if c.Primary() != nil {
		t.Error("Primary on empty collection should be nil")
	}
	g := NewGeneralFormat()
	c.Add(g)
	if c.Primary() != g {
		t.Error("Primary should return first format")
	}
}

func TestCollection_FormatValue_Empty(t *testing.T) {
	c := NewCollection()
	if got := c.FormatValue(42); got != "42" {
		t.Errorf("FormatValue on empty = %q, want 42", got)
	}
}

func TestCollection_FormatValue_Primary(t *testing.T) {
	c := NewCollection()
	c.Add(NewBooleanFormat())
	if got := c.FormatValue(true); got != "True" {
		t.Errorf("FormatValue = %q, want True", got)
	}
}

func TestCollection_All(t *testing.T) {
	c := NewCollection()
	g := NewGeneralFormat()
	b := NewBooleanFormat()
	c.Add(g)
	c.Add(b)
	all := c.All()
	if len(all) != 2 {
		t.Fatalf("All len = %d, want 2", len(all))
	}
	// Mutating the slice should not affect the collection.
	all[0] = NewCustomFormat()
	if c.Get(0) != g {
		t.Error("All returned mutable reference instead of copy")
	}
}

func TestCollection_Assign_WithClone(t *testing.T) {
	src := NewCollection()
	src.Add(&cloneableFormat{label: "orig"})

	dst := NewCollection()
	dst.Add(NewGeneralFormat())
	dst.Assign(src)

	if dst.Count() != 1 {
		t.Fatalf("Assign count = %d, want 1", dst.Count())
	}
	// Cloned — should be a different pointer.
	if dst.Get(0) == src.Get(0) {
		t.Error("Assign should deep-copy cloneable formats")
	}
	if dst.Get(0).FormatValue(nil) != "orig" {
		t.Error("Assign cloned format should have same value")
	}
}

func TestCollection_Assign_WithoutClone(t *testing.T) {
	src := NewCollection()
	sf := &simpleFormat{}
	src.Add(sf)

	dst := NewCollection()
	dst.Assign(src)

	// No clone → same pointer.
	if dst.Get(0) != sf {
		t.Error("Assign without Clone should reuse same pointer")
	}
}

func TestCollection_Assign_NilSrc(t *testing.T) {
	dst := NewCollection()
	dst.Add(NewGeneralFormat())
	dst.Assign(nil)
	if dst.Count() != 0 {
		t.Errorf("Assign(nil) should clear, count = %d", dst.Count())
	}
}

func TestToFloat64_IntViaNumberFormat(t *testing.T) {
	// Pass an int to NumberFormat.FormatValue to cover the `case int:` branch.
	f := NewNumberFormat()
	got := f.FormatValue(int(42))
	if got == "" {
		t.Error("NumberFormat.FormatValue(int(42)) returned empty string")
	}
}

func TestToFloat64_Float32(t *testing.T) {
	// The float32 case produces "return float64(t), true".
	got, ok := toFloat64(float32(3.14))
	if !ok {
		t.Error("expected ok=true for float32")
	}
	if got == 0 {
		t.Error("expected non-zero float64 from float32")
	}
}

func TestToFloat64_StringSuccess(t *testing.T) {
	got, ok := toFloat64("42.5")
	if !ok {
		t.Error("expected ok=true for numeric string")
	}
	if got != 42.5 {
		t.Errorf("toFloat64(\"42.5\") = %v, want 42.5", got)
	}
}

func TestToFloat64_FallbackReturnsFalse(t *testing.T) {
	// An unrecognized type should return (0, false).
	_, ok := toFloat64(struct{}{})
	if ok {
		t.Error("toFloat64 with unknown type should return ok=false")
	}
}

func TestToFloat64_StringBadValue(t *testing.T) {
	_, ok := toFloat64("notanumber")
	if ok {
		t.Error("toFloat64(\"notanumber\") should return ok=false")
	}
}
