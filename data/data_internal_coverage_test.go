package data

// data_internal_coverage_test.go — internal tests for unexported functions.
// Uses package data (not data_test) to access compare, toInt64, toFloat64, etc.

import (
	"testing"
	"time"
)

// ── toInt64 ───────────────────────────────────────────────────────────────────

func TestToInt64_Int(t *testing.T) {
	v, err := toInt64(int(42))
	if err != nil || v != 42 {
		t.Errorf("toInt64(int=42) = %v, %v; want 42, nil", v, err)
	}
}

func TestToInt64_Int32(t *testing.T) {
	v, err := toInt64(int32(100))
	if err != nil || v != 100 {
		t.Errorf("toInt64(int32=100) = %v, %v", v, err)
	}
}

func TestToInt64_Int64(t *testing.T) {
	v, err := toInt64(int64(9999))
	if err != nil || v != 9999 {
		t.Errorf("toInt64(int64=9999) = %v, %v", v, err)
	}
}

func TestToInt64_Float32(t *testing.T) {
	v, err := toInt64(float32(3.7))
	if err != nil {
		t.Errorf("toInt64(float32=3.7) unexpected error: %v", err)
	}
	if v != 3 { // truncated
		t.Errorf("toInt64(float32=3.7) = %d, want 3", v)
	}
}

func TestToInt64_Float64(t *testing.T) {
	v, err := toInt64(float64(5.9))
	if err != nil {
		t.Errorf("toInt64(float64=5.9) unexpected error: %v", err)
	}
	if v != 5 {
		t.Errorf("toInt64(float64=5.9) = %d, want 5", v)
	}
}

func TestToInt64_String_Error(t *testing.T) {
	_, err := toInt64("hello")
	if err == nil {
		t.Error("toInt64(string) should return error")
	}
}

// ── toFloat64 ─────────────────────────────────────────────────────────────────

func TestToFloat64_Float64(t *testing.T) {
	v, err := toFloat64(float64(3.14))
	if err != nil || v != 3.14 {
		t.Errorf("toFloat64(float64=3.14) = %v, %v", v, err)
	}
}

func TestToFloat64_Float32(t *testing.T) {
	v, err := toFloat64(float32(1.5))
	if err != nil {
		t.Errorf("toFloat64(float32=1.5) error: %v", err)
	}
	if v != float64(float32(1.5)) {
		t.Errorf("toFloat64(float32=1.5) = %v", v)
	}
}

func TestToFloat64_Int(t *testing.T) {
	v, err := toFloat64(int(10))
	if err != nil || v != 10.0 {
		t.Errorf("toFloat64(int=10) = %v, %v", v, err)
	}
}

func TestToFloat64_Int8(t *testing.T) {
	v, err := toFloat64(int8(8))
	if err != nil || v != 8.0 {
		t.Errorf("toFloat64(int8=8) = %v, %v", v, err)
	}
}

func TestToFloat64_Int16(t *testing.T) {
	v, err := toFloat64(int16(16))
	if err != nil || v != 16.0 {
		t.Errorf("toFloat64(int16=16) = %v, %v", v, err)
	}
}

func TestToFloat64_Int32(t *testing.T) {
	v, err := toFloat64(int32(32))
	if err != nil || v != 32.0 {
		t.Errorf("toFloat64(int32=32) = %v, %v", v, err)
	}
}

func TestToFloat64_Int64(t *testing.T) {
	v, err := toFloat64(int64(64))
	if err != nil || v != 64.0 {
		t.Errorf("toFloat64(int64=64) = %v, %v", v, err)
	}
}

func TestToFloat64_Uint(t *testing.T) {
	v, err := toFloat64(uint(7))
	if err != nil || v != 7.0 {
		t.Errorf("toFloat64(uint=7) = %v, %v", v, err)
	}
}

func TestToFloat64_Uint8(t *testing.T) {
	v, err := toFloat64(uint8(8))
	if err != nil || v != 8.0 {
		t.Errorf("toFloat64(uint8=8) = %v, %v", v, err)
	}
}

func TestToFloat64_Uint16(t *testing.T) {
	v, err := toFloat64(uint16(16))
	if err != nil || v != 16.0 {
		t.Errorf("toFloat64(uint16=16) = %v, %v", v, err)
	}
}

func TestToFloat64_Uint32(t *testing.T) {
	v, err := toFloat64(uint32(32))
	if err != nil || v != 32.0 {
		t.Errorf("toFloat64(uint32=32) = %v, %v", v, err)
	}
}

func TestToFloat64_Uint64(t *testing.T) {
	v, err := toFloat64(uint64(64))
	if err != nil || v != 64.0 {
		t.Errorf("toFloat64(uint64=64) = %v, %v", v, err)
	}
}

func TestToFloat64_String_Error(t *testing.T) {
	_, err := toFloat64("hello")
	if err == nil {
		t.Error("toFloat64(string) should return error")
	}
}

// ── compare ───────────────────────────────────────────────────────────────────

func TestCompare_NilA(t *testing.T) {
	_, ok := compare(nil, 42)
	if ok {
		t.Error("compare(nil, 42) should return ok=false")
	}
}

func TestCompare_NilB(t *testing.T) {
	_, ok := compare(42, nil)
	if ok {
		t.Error("compare(42, nil) should return ok=false")
	}
}

func TestCompare_Int_Equal(t *testing.T) {
	r, ok := compare(int(5), int(5))
	if !ok || r != 0 {
		t.Errorf("compare(5, 5) = %d, %v; want 0, true", r, ok)
	}
}

func TestCompare_Int_Less(t *testing.T) {
	r, ok := compare(int(3), int(5))
	if !ok || r >= 0 {
		t.Errorf("compare(3, 5) = %d, %v; want <0, true", r, ok)
	}
}

func TestCompare_Int_Greater(t *testing.T) {
	r, ok := compare(int(7), int(3))
	if !ok || r <= 0 {
		t.Errorf("compare(7, 3) = %d, %v; want >0, true", r, ok)
	}
}

func TestCompare_Int_IncompatibleB(t *testing.T) {
	_, ok := compare(int(5), "not a number")
	if ok {
		t.Error("compare(int, string) should return ok=false")
	}
}

func TestCompare_Int64_WithInt64(t *testing.T) {
	r, ok := compare(int64(100), int64(50))
	if !ok || r <= 0 {
		t.Errorf("compare(int64=100, int64=50) = %d, %v; want >0, true", r, ok)
	}
}

func TestCompare_Int32_WithInt(t *testing.T) {
	r, ok := compare(int32(5), int(5))
	if !ok || r != 0 {
		t.Errorf("compare(int32=5, int=5) = %d, %v; want 0, true", r, ok)
	}
}

func TestCompare_Int32_IncompatibleB(t *testing.T) {
	_, ok := compare(int32(5), "abc")
	if ok {
		t.Error("compare(int32, string) should return ok=false")
	}
}

func TestCompare_Float64_Equal(t *testing.T) {
	r, ok := compare(float64(3.14), float64(3.14))
	if !ok || r != 0 {
		t.Errorf("compare(3.14, 3.14) = %d, %v; want 0, true", r, ok)
	}
}

func TestCompare_Float64_IncompatibleB(t *testing.T) {
	_, ok := compare(float64(1.0), "abc")
	if ok {
		t.Error("compare(float64, string) should return ok=false")
	}
}

func TestCompare_Float32_Equal(t *testing.T) {
	r, ok := compare(float32(2.5), float32(2.5))
	if !ok || r != 0 {
		t.Errorf("compare(float32=2.5, float32=2.5) = %d, %v", r, ok)
	}
}

func TestCompare_Float32_IncompatibleB(t *testing.T) {
	_, ok := compare(float32(1.0), "abc")
	if ok {
		t.Error("compare(float32, string) should return ok=false")
	}
}

func TestCompare_String_Equal(t *testing.T) {
	r, ok := compare("hello", "hello")
	if !ok || r != 0 {
		t.Errorf("compare(hello, hello) = %d, %v; want 0, true", r, ok)
	}
}

func TestCompare_String_IncompatibleB(t *testing.T) {
	_, ok := compare("hello", 42)
	if ok {
		t.Error("compare(string, int) should return ok=false")
	}
}

func TestCompare_Time_Equal(t *testing.T) {
	now := time.Now()
	r, ok := compare(now, now)
	if !ok || r != 0 {
		t.Errorf("compare(time, time) equal: got %d, %v; want 0, true", r, ok)
	}
}

func TestCompare_Time_Before(t *testing.T) {
	t1 := time.Now()
	t2 := t1.Add(time.Hour)
	r, ok := compare(t1, t2)
	if !ok || r >= 0 {
		t.Errorf("compare(t1, t2) where t1<t2: got %d, %v; want <0, true", r, ok)
	}
}

func TestCompare_Time_After(t *testing.T) {
	t1 := time.Now()
	t2 := t1.Add(time.Hour)
	r, ok := compare(t2, t1)
	if !ok || r <= 0 {
		t.Errorf("compare(t2, t1) where t2>t1: got %d, %v; want >0, true", r, ok)
	}
}

func TestCompare_Time_IncompatibleB(t *testing.T) {
	_, ok := compare(time.Now(), "not a time")
	if ok {
		t.Error("compare(time.Time, string) should return ok=false")
	}
}

func TestCompare_Bool_True_True(t *testing.T) {
	r, ok := compare(true, true)
	if !ok || r != 0 {
		t.Errorf("compare(true, true) = %d, %v; want 0, true", r, ok)
	}
}

func TestCompare_Bool_True_False(t *testing.T) {
	r, ok := compare(true, false)
	if !ok || r <= 0 {
		t.Errorf("compare(true, false) = %d, %v; want >0, true", r, ok)
	}
}

func TestCompare_Bool_False_True(t *testing.T) {
	r, ok := compare(false, true)
	if !ok || r >= 0 {
		t.Errorf("compare(false, true) = %d, %v; want <0, true", r, ok)
	}
}

func TestCompare_Bool_IncompatibleB(t *testing.T) {
	_, ok := compare(true, "not a bool")
	if ok {
		t.Error("compare(bool, string) should return ok=false")
	}
}

func TestCompare_UnknownType(t *testing.T) {
	// struct type — falls through to default return (0, false)
	type myStruct struct{}
	_, ok := compare(myStruct{}, myStruct{})
	if ok {
		t.Error("compare(struct, struct) should return ok=false")
	}
}

// ── InitializeComponent ────────────────────────────────────────────────────────

func TestDataComponentBase_InitializeComponent(t *testing.T) {
	d := &DataComponentBase{enabled: true}
	// InitializeComponent is a no-op; just verify it doesn't panic.
	d.InitializeComponent()
}
