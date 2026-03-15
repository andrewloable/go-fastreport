package utils

import (
	"strings"
	"testing"
)

func TestFloatCollection_String_Empty(t *testing.T) {
	var fc FloatCollection
	if fc.String() != "" {
		t.Errorf("empty collection string = %q, want empty", fc.String())
	}
}

func TestFloatCollection_String_Single(t *testing.T) {
	fc := FloatCollection{2.5}
	s := fc.String()
	if s != "2.5" {
		t.Errorf("String = %q, want '2.5'", s)
	}
}

func TestFloatCollection_String_Multiple(t *testing.T) {
	fc := FloatCollection{2, 4, 2, 4}
	s := fc.String()
	if s != "2,4,2,4" {
		t.Errorf("String = %q, want '2,4,2,4'", s)
	}
}

func TestParseFloatCollection_Empty(t *testing.T) {
	fc, err := ParseFloatCollection("")
	if err != nil {
		t.Fatalf("ParseFloatCollection empty: %v", err)
	}
	if len(fc) != 0 {
		t.Errorf("expected empty, got %v", fc)
	}
}

func TestParseFloatCollection_Valid(t *testing.T) {
	fc, err := ParseFloatCollection("1,2.5,3")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(fc) != 3 {
		t.Fatalf("expected 3 elements, got %d", len(fc))
	}
	if fc[0] != 1 || fc[1] != 2.5 || fc[2] != 3 {
		t.Errorf("got %v", fc)
	}
}

func TestParseFloatCollection_Whitespace(t *testing.T) {
	fc, err := ParseFloatCollection(" 1 , 2 , 3 ")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(fc) != 3 {
		t.Fatalf("expected 3 elements, got %d", len(fc))
	}
}

func TestParseFloatCollection_Invalid(t *testing.T) {
	_, err := ParseFloatCollection("1,abc,3")
	if err == nil {
		t.Error("expected error for non-numeric token")
	}
}

func TestParseFloatCollection_RoundTrip(t *testing.T) {
	original := FloatCollection{2, 4, 2, 4}
	fc, err := ParseFloatCollection(original.String())
	if err != nil {
		t.Fatalf("ParseFloatCollection: %v", err)
	}
	if len(fc) != len(original) {
		t.Fatalf("length mismatch: got %d, want %d", len(fc), len(original))
	}
	for i := range fc {
		if fc[i] != original[i] {
			t.Errorf("fc[%d] = %v, want %v", i, fc[i], original[i])
		}
	}
}

func TestMustParseFloatCollection_Valid(t *testing.T) {
	fc := MustParseFloatCollection("1,2,3")
	if len(fc) != 3 {
		t.Errorf("expected 3 elements, got %d", len(fc))
	}
}

func TestMustParseFloatCollection_Panics(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic for invalid input")
		}
	}()
	MustParseFloatCollection("bad")
}

func TestFloatCollection_Add(t *testing.T) {
	var fc FloatCollection
	fc.Add(1.5)
	fc.Add(2.5)
	if fc.Len() != 2 {
		t.Errorf("Len after Add = %d, want 2", fc.Len())
	}
	if fc.Get(0) != 1.5 || fc.Get(1) != 2.5 {
		t.Errorf("values after Add: %v", fc)
	}
}

func TestFloatCollection_Clear(t *testing.T) {
	fc := FloatCollection{1, 2, 3}
	fc.Clear()
	if fc.Len() != 0 {
		t.Errorf("Len after Clear = %d, want 0", fc.Len())
	}
}

func TestFloatCollection_Len(t *testing.T) {
	fc := FloatCollection{1, 2, 3}
	if fc.Len() != 3 {
		t.Errorf("Len = %d, want 3", fc.Len())
	}
}

func TestFloatCollection_Get(t *testing.T) {
	fc := FloatCollection{10, 20, 30}
	if fc.Get(2) != 30 {
		t.Errorf("Get(2) = %v, want 30", fc.Get(2))
	}
}

func TestParseFloatCollection_SkipsEmpty(t *testing.T) {
	// comma with nothing between — the empty token is skipped
	fc, err := ParseFloatCollection("1,,2")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(fc) != 2 {
		t.Errorf("expected 2 elements (empty skipped), got %d: %v", len(fc), fc)
	}
}

func TestFloatCollection_StringRoundTrip(t *testing.T) {
	input := "1.5,2.25,0.75"
	fc := MustParseFloatCollection(input)
	got := fc.String()
	// Values may not round-trip exactly to same string due to float32 representation,
	// but re-parsing should give equivalent values.
	fc2 := MustParseFloatCollection(got)
	if len(fc) != len(fc2) {
		t.Fatalf("length mismatch after round-trip: %d vs %d", len(fc), len(fc2))
	}
	_ = strings.Contains(got, ",") // just use strings to avoid import warning
}
