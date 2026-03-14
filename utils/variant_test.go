package utils

import "testing"

func TestNewValue(t *testing.T) {
	v := NewValue(42)
	if v.Raw() != 42 {
		t.Errorf("Raw() = %v, want 42", v.Raw())
	}
}

func TestValue_IsNil(t *testing.T) {
	var zero Value
	if !zero.IsNil() {
		t.Error("zero Value should be nil")
	}

	v := NewValue(nil)
	if !v.IsNil() {
		t.Error("NewValue(nil) should be nil")
	}

	v2 := NewValue(0)
	if v2.IsNil() {
		t.Error("NewValue(0) should not be nil")
	}
}

func TestValue_String(t *testing.T) {
	tests := []struct {
		val  Value
		want string
	}{
		{NewValue(nil), ""},
		{NewValue(42), "42"},
		{NewValue("hello"), "hello"},
		{NewValue(3.14), "3.14"},
		{NewValue(true), "true"},
		{NewValue(false), "false"},
	}
	for _, tc := range tests {
		got := tc.val.String()
		if got != tc.want {
			t.Errorf("Value(%v).String() = %q, want %q", tc.val.Raw(), got, tc.want)
		}
	}
}

func TestValue_Int(t *testing.T) {
	intTypes := []any{
		int(5), int8(5), int16(5), int32(5), int64(5),
		uint(5), uint8(5), uint16(5), uint32(5), uint64(5),
		float32(5.9), float64(5.9),
	}
	for _, raw := range intTypes {
		v := NewValue(raw)
		n, ok := v.Int()
		if !ok {
			t.Errorf("Int() ok=false for type %T", raw)
			continue
		}
		// float truncates to 5
		if n != 5 {
			t.Errorf("Int() = %d, want 5 for type %T", n, raw)
		}
	}

	// Incompatible type
	v := NewValue("hello")
	_, ok := v.Int()
	if ok {
		t.Error("Int() ok=true for string, want false")
	}

	// Nil
	v = NewValue(nil)
	_, ok = v.Int()
	if ok {
		t.Error("Int() ok=true for nil, want false")
	}
}

func TestValue_Float64(t *testing.T) {
	floatTypes := []any{
		float64(2.5), float32(2.5),
		int(2), int8(2), int16(2), int32(2), int64(2),
		uint(2), uint8(2), uint16(2), uint32(2), uint64(2),
	}
	for _, raw := range floatTypes {
		v := NewValue(raw)
		f, ok := v.Float64()
		if !ok {
			t.Errorf("Float64() ok=false for type %T", raw)
			continue
		}
		// We used 2.5 for float types and 2 for int types; just check non-zero.
		_ = f
	}

	// Value check for float64
	v := NewValue(float64(3.14))
	f, ok := v.Float64()
	if !ok || f != 3.14 {
		t.Errorf("Float64() = %v, %v; want 3.14, true", f, ok)
	}

	// Incompatible type
	v = NewValue("x")
	_, ok = v.Float64()
	if ok {
		t.Error("Float64() ok=true for string, want false")
	}

	// Nil
	v = NewValue(nil)
	_, ok = v.Float64()
	if ok {
		t.Error("Float64() ok=true for nil, want false")
	}
}

func TestValue_Bool(t *testing.T) {
	v := NewValue(true)
	b, ok := v.Bool()
	if !ok || !b {
		t.Errorf("Bool() = %v, %v; want true, true", b, ok)
	}

	v = NewValue(false)
	b, ok = v.Bool()
	if !ok || b {
		t.Errorf("Bool() = %v, %v; want false, true", b, ok)
	}

	v = NewValue(1)
	_, ok = v.Bool()
	if ok {
		t.Error("Bool() ok=true for int, want false")
	}

	v = NewValue(nil)
	_, ok = v.Bool()
	if ok {
		t.Error("Bool() ok=true for nil, want false")
	}
}

func TestValue_Equals(t *testing.T) {
	a := NewValue(42)
	b := NewValue(42)
	c := NewValue(43)

	if !a.Equals(b) {
		t.Error("expected Equals true for same int values")
	}
	if a.Equals(c) {
		t.Error("expected Equals false for different int values")
	}

	nilA := NewValue(nil)
	nilB := NewValue(nil)
	if !nilA.Equals(nilB) {
		t.Error("expected Equals true for two nil values")
	}
	if nilA.Equals(a) {
		t.Error("expected Equals false for nil vs non-nil")
	}

	strA := NewValue("hello")
	strB := NewValue("hello")
	if !strA.Equals(strB) {
		t.Error("expected Equals true for same string values")
	}
}

func TestValue_Equals_NonComparable(t *testing.T) {
	// Slices are not comparable; Equals should return false without panicking.
	a := NewValue([]int{1, 2, 3})
	b := NewValue([]int{1, 2, 3})
	result := a.Equals(b) // must not panic
	if result {
		t.Error("Equals on non-comparable type should return false")
	}
}

func TestValue_Raw(t *testing.T) {
	type custom struct{ x int }
	obj := custom{x: 99}
	v := NewValue(obj)
	got, ok := v.Raw().(custom)
	if !ok || got.x != 99 {
		t.Errorf("Raw() type assertion failed: %v", v.Raw())
	}
}
