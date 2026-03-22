package serial

import (
	"strings"
	"testing"
)

// TestReadDouble_present verifies that ReadDouble returns the parsed float64
// value when the attribute exists.
func TestReadDouble_present(t *testing.T) {
	src := `<Item attr="3.14159265358979" />`
	r := NewReader(strings.NewReader(src))
	_, ok := r.ReadObjectHeader()
	if !ok {
		t.Fatal("ReadObjectHeader failed")
	}
	got := r.ReadDouble("attr", 0)
	want := 3.14159265358979
	if got != want {
		t.Errorf("ReadDouble: got %v, want %v", got, want)
	}
}

// TestReadDouble_absent verifies that ReadDouble returns the default value
// when the attribute is missing.
func TestReadDouble_absent(t *testing.T) {
	src := `<Item />`
	r := NewReader(strings.NewReader(src))
	_, ok := r.ReadObjectHeader()
	if !ok {
		t.Fatal("ReadObjectHeader failed")
	}
	got := r.ReadDouble("missing", 99.5)
	if got != 99.5 {
		t.Errorf("ReadDouble absent: got %v, want 99.5", got)
	}
}

// TestReadDouble_invalid verifies that ReadDouble returns the default value
// when the attribute value cannot be parsed as a float.
func TestReadDouble_invalid(t *testing.T) {
	src := `<Item attr="notanumber" />`
	r := NewReader(strings.NewReader(src))
	_, ok := r.ReadObjectHeader()
	if !ok {
		t.Fatal("ReadObjectHeader failed")
	}
	got := r.ReadDouble("attr", 42.0)
	if got != 42.0 {
		t.Errorf("ReadDouble invalid: got %v, want 42.0", got)
	}
}

// TestHasProperty_present verifies that HasProperty returns true when the
// attribute exists in the current element.
func TestHasProperty_present(t *testing.T) {
	src := `<Item name="hello" />`
	r := NewReader(strings.NewReader(src))
	_, ok := r.ReadObjectHeader()
	if !ok {
		t.Fatal("ReadObjectHeader failed")
	}
	if !r.HasProperty("name") {
		t.Error("HasProperty(\"name\") = false, want true")
	}
}

// TestHasProperty_absent verifies that HasProperty returns false when the
// attribute does not exist in the current element.
func TestHasProperty_absent(t *testing.T) {
	src := `<Item name="hello" />`
	r := NewReader(strings.NewReader(src))
	_, ok := r.ReadObjectHeader()
	if !ok {
		t.Fatal("ReadObjectHeader failed")
	}
	if r.HasProperty("nonexistent") {
		t.Error("HasProperty(\"nonexistent\") = true, want false")
	}
}
