package expr

import (
	"strings"
	"testing"
	"time"
)

func TestBuiltinFunctions_Keys(t *testing.T) {
	fns := BuiltinFunctions()
	expected := []string{"IIF", "Format", "DateDiff", "Str", "Int", "Float", "Len", "Upper", "Lower"}
	for _, k := range expected {
		if _, ok := fns[k]; !ok {
			t.Errorf("missing builtin function: %s", k)
		}
	}
}

// --- IIF ---

func TestIIF_True(t *testing.T) {
	result := iif(true, "yes", "no")
	if result != "yes" {
		t.Errorf("expected yes, got %v", result)
	}
}

func TestIIF_False(t *testing.T) {
	result := iif(false, "yes", "no")
	if result != "no" {
		t.Errorf("expected no, got %v", result)
	}
}

func TestIIF_NilValues(t *testing.T) {
	result := iif(true, nil, "fallback")
	if result != nil {
		t.Errorf("expected nil, got %v", result)
	}
}

// --- Format ---

func TestFormat_WithPercent(t *testing.T) {
	result := formatValue(3.14159, "%.2f")
	if result != "3.14" {
		t.Errorf("expected 3.14, got %q", result)
	}
}

func TestFormat_WithoutPercent(t *testing.T) {
	result := formatValue(42, "d")
	if !strings.Contains(result, "42") {
		t.Errorf("expected result to contain 42, got %q", result)
	}
}

func TestFormat_EmptyFormat(t *testing.T) {
	result := formatValue("hello", "")
	if result != "hello" {
		t.Errorf("expected hello, got %q", result)
	}
}

// --- DateDiff ---

func TestDateDiff_Days(t *testing.T) {
	d1 := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	d2 := time.Date(2024, 1, 11, 0, 0, 0, 0, time.UTC)
	result, err := dateDiff(d1, d2, "days")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != 10.0 {
		t.Errorf("expected 10, got %v", result)
	}
}

func TestDateDiff_Hours(t *testing.T) {
	d1 := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	d2 := time.Date(2024, 1, 1, 5, 0, 0, 0, time.UTC)
	result, err := dateDiff(d1, d2, "hours")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != 5.0 {
		t.Errorf("expected 5, got %v", result)
	}
}

func TestDateDiff_Minutes(t *testing.T) {
	d1 := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	d2 := time.Date(2024, 1, 1, 0, 30, 0, 0, time.UTC)
	result, err := dateDiff(d1, d2, "minutes")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != 30.0 {
		t.Errorf("expected 30, got %v", result)
	}
}

func TestDateDiff_Seconds(t *testing.T) {
	d1 := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	d2 := time.Date(2024, 1, 1, 0, 0, 45, 0, time.UTC)
	result, err := dateDiff(d1, d2, "seconds")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != 45.0 {
		t.Errorf("expected 45, got %v", result)
	}
}

func TestDateDiff_InvalidUnit(t *testing.T) {
	d1 := time.Now()
	d2 := time.Now()
	_, err := dateDiff(d1, d2, "years")
	if err == nil {
		t.Error("expected error for unknown unit")
	}
}

// --- Str ---

func TestStr_String(t *testing.T) {
	if str("hello") != "hello" {
		t.Error("expected hello")
	}
}

func TestStr_Int(t *testing.T) {
	if str(42) != "42" {
		t.Error("expected 42")
	}
}

func TestStr_Nil(t *testing.T) {
	if str(nil) != "" {
		t.Error("expected empty string for nil")
	}
}

// --- Int ---

func TestToInt_Int(t *testing.T) {
	v, err := toInt(42)
	if err != nil || v != 42 {
		t.Errorf("expected 42, got %v %v", v, err)
	}
}

func TestToInt_Int8(t *testing.T) {
	v, err := toInt(int8(10))
	if err != nil || v != 10 {
		t.Errorf("expected 10, got %v %v", v, err)
	}
}

func TestToInt_Int16(t *testing.T) {
	v, err := toInt(int16(200))
	if err != nil || v != 200 {
		t.Errorf("expected 200, got %v %v", v, err)
	}
}

func TestToInt_Int32(t *testing.T) {
	v, err := toInt(int32(300))
	if err != nil || v != 300 {
		t.Errorf("expected 300, got %v %v", v, err)
	}
}

func TestToInt_Int64(t *testing.T) {
	v, err := toInt(int64(999))
	if err != nil || v != 999 {
		t.Errorf("expected 999, got %v %v", v, err)
	}
}

func TestToInt_Uint(t *testing.T) {
	v, err := toInt(uint(5))
	if err != nil || v != 5 {
		t.Errorf("expected 5, got %v %v", v, err)
	}
}

func TestToInt_Uint8(t *testing.T) {
	v, err := toInt(uint8(8))
	if err != nil || v != 8 {
		t.Errorf("expected 8, got %v %v", v, err)
	}
}

func TestToInt_Uint16(t *testing.T) {
	v, err := toInt(uint16(16))
	if err != nil || v != 16 {
		t.Errorf("expected 16, got %v %v", v, err)
	}
}

func TestToInt_Uint32(t *testing.T) {
	v, err := toInt(uint32(32))
	if err != nil || v != 32 {
		t.Errorf("expected 32, got %v %v", v, err)
	}
}

func TestToInt_Uint64(t *testing.T) {
	v, err := toInt(uint64(64))
	if err != nil || v != 64 {
		t.Errorf("expected 64, got %v %v", v, err)
	}
}

func TestToInt_Float32(t *testing.T) {
	v, err := toInt(float32(3.7))
	if err != nil || v != 4 {
		t.Errorf("expected 4 (rounded), got %v %v", v, err)
	}
}

func TestToInt_Float64(t *testing.T) {
	v, err := toInt(float64(2.5))
	if err != nil || v != 3 {
		t.Errorf("expected 3 (rounded), got %v %v", v, err)
	}
}

func TestToInt_BoolTrue(t *testing.T) {
	v, err := toInt(true)
	if err != nil || v != 1 {
		t.Errorf("expected 1, got %v %v", v, err)
	}
}

func TestToInt_BoolFalse(t *testing.T) {
	v, err := toInt(false)
	if err != nil || v != 0 {
		t.Errorf("expected 0, got %v %v", v, err)
	}
}

func TestToInt_String(t *testing.T) {
	v, err := toInt("42")
	if err != nil || v != 42 {
		t.Errorf("expected 42, got %v %v", v, err)
	}
}

func TestToInt_StringInvalid(t *testing.T) {
	_, err := toInt("abc")
	if err == nil {
		t.Error("expected error for non-numeric string")
	}
}

func TestToInt_Unsupported(t *testing.T) {
	_, err := toInt(struct{}{})
	if err == nil {
		t.Error("expected error for unsupported type")
	}
}

// --- Float ---

func TestToFloat_Float64(t *testing.T) {
	v, err := toFloat(3.14)
	if err != nil || v != 3.14 {
		t.Errorf("expected 3.14, got %v %v", v, err)
	}
}

func TestToFloat_Float32(t *testing.T) {
	v, err := toFloat(float32(1.5))
	if err != nil || v != 1.5 {
		t.Errorf("expected 1.5, got %v %v", v, err)
	}
}

func TestToFloat_Int(t *testing.T) {
	v, err := toFloat(10)
	if err != nil || v != 10.0 {
		t.Errorf("expected 10.0, got %v %v", v, err)
	}
}

func TestToFloat_Int8(t *testing.T) {
	v, err := toFloat(int8(2))
	if err != nil || v != 2.0 {
		t.Errorf("expected 2.0, got %v %v", v, err)
	}
}

func TestToFloat_Int16(t *testing.T) {
	v, err := toFloat(int16(3))
	if err != nil || v != 3.0 {
		t.Errorf("expected 3.0, got %v %v", v, err)
	}
}

func TestToFloat_Int32(t *testing.T) {
	v, err := toFloat(int32(4))
	if err != nil || v != 4.0 {
		t.Errorf("expected 4.0, got %v %v", v, err)
	}
}

func TestToFloat_Int64(t *testing.T) {
	v, err := toFloat(int64(5))
	if err != nil || v != 5.0 {
		t.Errorf("expected 5.0, got %v %v", v, err)
	}
}

func TestToFloat_Uint(t *testing.T) {
	v, err := toFloat(uint(6))
	if err != nil || v != 6.0 {
		t.Errorf("expected 6.0, got %v %v", v, err)
	}
}

func TestToFloat_Uint8(t *testing.T) {
	v, err := toFloat(uint8(7))
	if err != nil || v != 7.0 {
		t.Errorf("expected 7.0, got %v %v", v, err)
	}
}

func TestToFloat_Uint16(t *testing.T) {
	v, err := toFloat(uint16(8))
	if err != nil || v != 8.0 {
		t.Errorf("expected 8.0, got %v %v", v, err)
	}
}

func TestToFloat_Uint32(t *testing.T) {
	v, err := toFloat(uint32(9))
	if err != nil || v != 9.0 {
		t.Errorf("expected 9.0, got %v %v", v, err)
	}
}

func TestToFloat_Uint64(t *testing.T) {
	v, err := toFloat(uint64(11))
	if err != nil || v != 11.0 {
		t.Errorf("expected 11.0, got %v %v", v, err)
	}
}

func TestToFloat_BoolTrue(t *testing.T) {
	v, err := toFloat(true)
	if err != nil || v != 1.0 {
		t.Errorf("expected 1.0, got %v %v", v, err)
	}
}

func TestToFloat_BoolFalse(t *testing.T) {
	v, err := toFloat(false)
	if err != nil || v != 0.0 {
		t.Errorf("expected 0.0, got %v %v", v, err)
	}
}

func TestToFloat_String(t *testing.T) {
	v, err := toFloat("2.718")
	if err != nil || v != 2.718 {
		t.Errorf("expected 2.718, got %v %v", v, err)
	}
}

func TestToFloat_StringInvalid(t *testing.T) {
	_, err := toFloat("not_a_float")
	if err == nil {
		t.Error("expected error for non-numeric string")
	}
}

func TestToFloat_Unsupported(t *testing.T) {
	_, err := toFloat(struct{}{})
	if err == nil {
		t.Error("expected error for unsupported type")
	}
}

// --- Len ---

func TestStrLen_ASCII(t *testing.T) {
	if strLen("hello") != 5 {
		t.Error("expected 5")
	}
}

func TestStrLen_Empty(t *testing.T) {
	if strLen("") != 0 {
		t.Error("expected 0")
	}
}

func TestStrLen_Unicode(t *testing.T) {
	// "こんにちは" is 5 runes.
	if strLen("こんにちは") != 5 {
		t.Errorf("expected 5 runes, got %d", strLen("こんにちは"))
	}
}

// --- Upper ---

func TestUpper_Basic(t *testing.T) {
	if upper("hello") != "HELLO" {
		t.Errorf("expected HELLO, got %q", upper("hello"))
	}
}

func TestUpper_Mixed(t *testing.T) {
	if upper("Hello World") != "HELLO WORLD" {
		t.Errorf("expected HELLO WORLD, got %q", upper("Hello World"))
	}
}

func TestUpper_AlreadyUpper(t *testing.T) {
	if upper("ABC") != "ABC" {
		t.Error("expected ABC")
	}
}

// --- Lower ---

func TestLower_Basic(t *testing.T) {
	if lower("HELLO") != "hello" {
		t.Errorf("expected hello, got %q", lower("HELLO"))
	}
}

func TestLower_Mixed(t *testing.T) {
	if lower("Hello World") != "hello world" {
		t.Errorf("expected hello world, got %q", lower("Hello World"))
	}
}

func TestLower_AlreadyLower(t *testing.T) {
	if lower("abc") != "abc" {
		t.Error("expected abc")
	}
}

// --- Integration with Evaluator ---

func TestEval_Str(t *testing.T) {
	e := NewEvaluator(Env{"val": 123})
	result, err := e.Eval("Str(val)")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != "123" {
		t.Errorf("expected \"123\", got %v", result)
	}
}

func TestEval_Upper(t *testing.T) {
	e := NewEvaluator(Env{"s": "hello"})
	result, err := e.Eval("Upper(s)")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != "HELLO" {
		t.Errorf("expected HELLO, got %v", result)
	}
}

func TestEval_Lower(t *testing.T) {
	e := NewEvaluator(Env{"s": "WORLD"})
	result, err := e.Eval("Lower(s)")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != "world" {
		t.Errorf("expected world, got %v", result)
	}
}

func TestEval_Len(t *testing.T) {
	e := NewEvaluator(Env{"s": "hello"})
	result, err := e.Eval("Len(s)")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.(int) != 5 {
		t.Errorf("expected 5, got %v", result)
	}
}
