package functions_test

import (
	"math"
	"testing"

	"github.com/andrewloable/go-fastreport/functions"
)

// ── Truncate ──────────────────────────────────────────────────────────────────

func TestTruncate_Positive(t *testing.T) {
	if got := functions.Truncate(3.7); got != 3.0 {
		t.Errorf("Truncate(3.7) = %v, want 3.0", got)
	}
}

func TestTruncate_Negative(t *testing.T) {
	if got := functions.Truncate(-3.7); got != -3.0 {
		t.Errorf("Truncate(-3.7) = %v, want -3.0", got)
	}
}

func TestTruncate_Zero(t *testing.T) {
	if got := functions.Truncate(0.9); got != 0.0 {
		t.Errorf("Truncate(0.9) = %v, want 0.0", got)
	}
}

func TestTruncate_Exact(t *testing.T) {
	if got := functions.Truncate(5.0); got != 5.0 {
		t.Errorf("Truncate(5.0) = %v, want 5.0", got)
	}
}

// ── Sign ──────────────────────────────────────────────────────────────────────

func TestSign_Negative(t *testing.T) {
	if got := functions.Sign(-5.0); got != -1.0 {
		t.Errorf("Sign(-5) = %v, want -1", got)
	}
}

func TestSign_Zero(t *testing.T) {
	if got := functions.Sign(0.0); got != 0.0 {
		t.Errorf("Sign(0) = %v, want 0", got)
	}
}

func TestSign_Positive(t *testing.T) {
	if got := functions.Sign(3.0); got != 1.0 {
		t.Errorf("Sign(3) = %v, want 1", got)
	}
}

func TestSign_NegativeFraction(t *testing.T) {
	if got := functions.Sign(-0.001); got != -1.0 {
		t.Errorf("Sign(-0.001) = %v, want -1", got)
	}
}

func TestSign_PositiveFraction(t *testing.T) {
	if got := functions.Sign(0.001); got != 1.0 {
		t.Errorf("Sign(0.001) = %v, want 1", got)
	}
}

// ── Log10 ─────────────────────────────────────────────────────────────────────

func TestLog10_Hundred(t *testing.T) {
	if got := functions.Log10(100.0); got != 2.0 {
		t.Errorf("Log10(100) = %v, want 2", got)
	}
}

func TestLog10_One(t *testing.T) {
	if got := functions.Log10(1.0); got != 0.0 {
		t.Errorf("Log10(1) = %v, want 0", got)
	}
}

func TestLog10_Ten(t *testing.T) {
	if got := functions.Log10(10.0); got != 1.0 {
		t.Errorf("Log10(10) = %v, want 1", got)
	}
}

func TestLog10_ReturnType(t *testing.T) {
	var got float64 = functions.Log10(1000.0)
	if got != 3.0 {
		t.Errorf("Log10(1000) = %v, want 3", got)
	}
}

// ── Exp ───────────────────────────────────────────────────────────────────────

func TestExp_Zero(t *testing.T) {
	if got := functions.Exp(0.0); got != 1.0 {
		t.Errorf("Exp(0) = %v, want 1", got)
	}
}

func TestExp_One(t *testing.T) {
	got := functions.Exp(1.0)
	if math.Abs(got-math.E) > 1e-10 {
		t.Errorf("Exp(1) = %v, want e (%v)", got, math.E)
	}
}

func TestExp_Negative(t *testing.T) {
	got := functions.Exp(-1.0)
	want := 1.0 / math.E
	if math.Abs(got-want) > 1e-10 {
		t.Errorf("Exp(-1) = %v, want 1/e (%v)", got, want)
	}
}

// ── Pow ───────────────────────────────────────────────────────────────────────

func TestPow_TwoToTen(t *testing.T) {
	if got := functions.Pow(2.0, 10.0); got != 1024.0 {
		t.Errorf("Pow(2,10) = %v, want 1024", got)
	}
}

func TestPow_ZeroExponent(t *testing.T) {
	if got := functions.Pow(5.0, 0.0); got != 1.0 {
		t.Errorf("Pow(5,0) = %v, want 1", got)
	}
}

func TestPow_OneBase(t *testing.T) {
	if got := functions.Pow(1.0, 100.0); got != 1.0 {
		t.Errorf("Pow(1,100) = %v, want 1", got)
	}
}

func TestPow_Square(t *testing.T) {
	if got := functions.Pow(4.0, 2.0); got != 16.0 {
		t.Errorf("Pow(4,2) = %v, want 16", got)
	}
}

func TestPow_ReturnType(t *testing.T) {
	var got float64 = functions.Pow(3.0, 3.0)
	if got != 27.0 {
		t.Errorf("Pow(3,3) = %v, want 27", got)
	}
}

// ── Sqrt ──────────────────────────────────────────────────────────────────────

func TestSqrt_Nine(t *testing.T) {
	if got := functions.Sqrt(9.0); got != 3.0 {
		t.Errorf("Sqrt(9) = %v, want 3", got)
	}
}

func TestSqrt_Zero(t *testing.T) {
	if got := functions.Sqrt(0.0); got != 0.0 {
		t.Errorf("Sqrt(0) = %v, want 0", got)
	}
}

func TestSqrt_One(t *testing.T) {
	if got := functions.Sqrt(1.0); got != 1.0 {
		t.Errorf("Sqrt(1) = %v, want 1", got)
	}
}

func TestSqrt_Four(t *testing.T) {
	if got := functions.Sqrt(4.0); got != 2.0 {
		t.Errorf("Sqrt(4) = %v, want 2", got)
	}
}

func TestSqrt_ReturnType(t *testing.T) {
	var got float64 = functions.Sqrt(25.0)
	if got != 5.0 {
		t.Errorf("Sqrt(25) = %v, want 5", got)
	}
}

// ── Sin ───────────────────────────────────────────────────────────────────────

func TestSin_Zero(t *testing.T) {
	if got := functions.Sin(0.0); got != 0.0 {
		t.Errorf("Sin(0) = %v, want 0", got)
	}
}

func TestSin_PiOverTwo(t *testing.T) {
	got := functions.Sin(math.Pi / 2)
	if math.Abs(got-1.0) > 1e-10 {
		t.Errorf("Sin(pi/2) = %v, want 1", got)
	}
}

func TestSin_Pi(t *testing.T) {
	got := functions.Sin(math.Pi)
	if math.Abs(got) > 1e-10 {
		t.Errorf("Sin(pi) = %v, want ~0", got)
	}
}

func TestSin_ReturnType(t *testing.T) {
	var got float64 = functions.Sin(0.0)
	_ = got
}

// ── Cos ───────────────────────────────────────────────────────────────────────

func TestCos_Zero(t *testing.T) {
	if got := functions.Cos(0.0); got != 1.0 {
		t.Errorf("Cos(0) = %v, want 1", got)
	}
}

func TestCos_PiOverTwo(t *testing.T) {
	got := functions.Cos(math.Pi / 2)
	if math.Abs(got) > 1e-10 {
		t.Errorf("Cos(pi/2) = %v, want ~0", got)
	}
}

func TestCos_Pi(t *testing.T) {
	got := functions.Cos(math.Pi)
	if math.Abs(got-(-1.0)) > 1e-10 {
		t.Errorf("Cos(pi) = %v, want -1", got)
	}
}

func TestCos_ReturnType(t *testing.T) {
	var got float64 = functions.Cos(0.0)
	_ = got
}

// ── Tan ───────────────────────────────────────────────────────────────────────

func TestTan_Zero(t *testing.T) {
	if got := functions.Tan(0.0); got != 0.0 {
		t.Errorf("Tan(0) = %v, want 0", got)
	}
}

func TestTan_PiOverFour(t *testing.T) {
	got := functions.Tan(math.Pi / 4)
	if math.Abs(got-1.0) > 1e-10 {
		t.Errorf("Tan(pi/4) = %v, want ~1", got)
	}
}

func TestTan_Pi(t *testing.T) {
	got := functions.Tan(math.Pi)
	if math.Abs(got) > 1e-10 {
		t.Errorf("Tan(pi) = %v, want ~0", got)
	}
}

func TestTan_ReturnType(t *testing.T) {
	var got float64 = functions.Tan(0.0)
	_ = got
}

// ── Asin ──────────────────────────────────────────────────────────────────────

func TestAsin_Zero(t *testing.T) {
	if got := functions.Asin(0.0); got != 0.0 {
		t.Errorf("Asin(0) = %v, want 0", got)
	}
}

func TestAsin_One(t *testing.T) {
	got := functions.Asin(1.0)
	if math.Abs(got-math.Pi/2) > 1e-10 {
		t.Errorf("Asin(1) = %v, want pi/2", got)
	}
}

// ── Acos ──────────────────────────────────────────────────────────────────────

func TestAcos_One(t *testing.T) {
	if got := functions.Acos(1.0); got != 0.0 {
		t.Errorf("Acos(1) = %v, want 0", got)
	}
}

func TestAcos_Zero(t *testing.T) {
	got := functions.Acos(0.0)
	if math.Abs(got-math.Pi/2) > 1e-10 {
		t.Errorf("Acos(0) = %v, want pi/2", got)
	}
}

// ── Atan ──────────────────────────────────────────────────────────────────────

func TestAtan_Zero(t *testing.T) {
	if got := functions.Atan(0.0); got != 0.0 {
		t.Errorf("Atan(0) = %v, want 0", got)
	}
}

func TestAtan_One(t *testing.T) {
	got := functions.Atan(1.0)
	if math.Abs(got-math.Pi/4) > 1e-10 {
		t.Errorf("Atan(1) = %v, want pi/4", got)
	}
}

// ── Atan2 ─────────────────────────────────────────────────────────────────────

func TestAtan2_OneOne(t *testing.T) {
	got := functions.Atan2(1.0, 1.0)
	if math.Abs(got-math.Pi/4) > 1e-10 {
		t.Errorf("Atan2(1,1) = %v, want pi/4", got)
	}
}

func TestAtan2_ZeroZero(t *testing.T) {
	// Atan2(0,0) is implementation-defined but math.Atan2 returns 0
	if got := functions.Atan2(0.0, 1.0); got != 0.0 {
		t.Errorf("Atan2(0,1) = %v, want 0", got)
	}
}

func TestAtan2_NegativeX(t *testing.T) {
	got := functions.Atan2(0.0, -1.0)
	if math.Abs(got-math.Pi) > 1e-10 {
		t.Errorf("Atan2(0,-1) = %v, want pi", got)
	}
}

// ── String functions previously uncovered ─────────────────────────────────────

func TestTrimStart(t *testing.T) {
	if got := functions.TrimStart("  hello"); got != "hello" {
		t.Errorf("TrimStart: got %q, want 'hello'", got)
	}
	// Should not trim trailing spaces.
	if got := functions.TrimStart("  hello  "); got != "hello  " {
		t.Errorf("TrimStart trailing: got %q, want 'hello  '", got)
	}
}

func TestTrimEnd(t *testing.T) {
	if got := functions.TrimEnd("hello  "); got != "hello" {
		t.Errorf("TrimEnd: got %q, want 'hello'", got)
	}
	// Should not trim leading spaces.
	if got := functions.TrimEnd("  hello  "); got != "  hello" {
		t.Errorf("TrimEnd leading: got %q, want '  hello'", got)
	}
}

func TestLastIndexOf_Found(t *testing.T) {
	if got := functions.LastIndexOf("hello world hello", "hello"); got != 12 {
		t.Errorf("LastIndexOf: got %d, want 12", got)
	}
}

func TestLastIndexOf_NotFound(t *testing.T) {
	if got := functions.LastIndexOf("hello", "xyz"); got != -1 {
		t.Errorf("LastIndexOf not found: got %d, want -1", got)
	}
}

func TestLastIndexOf_FirstOccurrence(t *testing.T) {
	// Only one occurrence — should return index 0.
	if got := functions.LastIndexOf("hello", "h"); got != 0 {
		t.Errorf("LastIndexOf single: got %d, want 0", got)
	}
}

func TestSplit(t *testing.T) {
	got := functions.Split("a,b,c", ",")
	if len(got) != 3 {
		t.Errorf("Split: got %v (len=%d), want 3 parts", got, len(got))
	}
	if got[0] != "a" || got[1] != "b" || got[2] != "c" {
		t.Errorf("Split: got %v, want [a b c]", got)
	}
}

func TestSplit_NoSep(t *testing.T) {
	got := functions.Split("hello", ",")
	if len(got) != 1 || got[0] != "hello" {
		t.Errorf("Split no sep: got %v, want [hello]", got)
	}
}

func TestJoin(t *testing.T) {
	if got := functions.Join(", ", []string{"a", "b", "c"}); got != "a, b, c" {
		t.Errorf("Join: got %q, want 'a, b, c'", got)
	}
}

func TestJoin_Empty(t *testing.T) {
	if got := functions.Join(",", []string{}); got != "" {
		t.Errorf("Join empty: got %q, want ''", got)
	}
}

func TestIsNullOrEmpty_Empty(t *testing.T) {
	if !functions.IsNullOrEmpty("") {
		t.Error("IsNullOrEmpty('') should be true")
	}
}

func TestIsNullOrEmpty_NonEmpty(t *testing.T) {
	if functions.IsNullOrEmpty("hello") {
		t.Error("IsNullOrEmpty('hello') should be false")
	}
}

func TestIsNullOrWhiteSpace_Whitespace(t *testing.T) {
	if !functions.IsNullOrWhiteSpace("   ") {
		t.Error("IsNullOrWhiteSpace('   ') should be true")
	}
}

func TestIsNullOrWhiteSpace_Empty(t *testing.T) {
	if !functions.IsNullOrWhiteSpace("") {
		t.Error("IsNullOrWhiteSpace('') should be true")
	}
}

func TestIsNullOrWhiteSpace_NonEmpty(t *testing.T) {
	if functions.IsNullOrWhiteSpace("hello") {
		t.Error("IsNullOrWhiteSpace('hello') should be false")
	}
}

func TestConcat(t *testing.T) {
	if got := functions.Concat("hello", " ", "world"); got != "hello world" {
		t.Errorf("Concat: got %q, want 'hello world'", got)
	}
}

func TestConcat_Single(t *testing.T) {
	if got := functions.Concat("only"); got != "only" {
		t.Errorf("Concat single: got %q, want 'only'", got)
	}
}

func TestConcat_Empty(t *testing.T) {
	if got := functions.Concat(); got != "" {
		t.Errorf("Concat empty: got %q, want ''", got)
	}
}

// ── Type conversion functions previously uncovered ────────────────────────────

func TestToBoolean_Bool(t *testing.T) {
	if !functions.ToBoolean(true) {
		t.Error("ToBoolean(true) should be true")
	}
	if functions.ToBoolean(false) {
		t.Error("ToBoolean(false) should be false")
	}
}

func TestToBoolean_Nil(t *testing.T) {
	if functions.ToBoolean(nil) {
		t.Error("ToBoolean(nil) should be false")
	}
}

func TestToBoolean_Int(t *testing.T) {
	if !functions.ToBoolean(1) {
		t.Error("ToBoolean(1) should be true")
	}
	if functions.ToBoolean(0) {
		t.Error("ToBoolean(0) should be false")
	}
}

func TestToBoolean_String(t *testing.T) {
	if !functions.ToBoolean("true") {
		t.Error("ToBoolean('true') should be true")
	}
	if functions.ToBoolean("false") {
		t.Error("ToBoolean('false') should be false")
	}
	// invalid string → false
	if functions.ToBoolean("invalid") {
		t.Error("ToBoolean('invalid') should be false")
	}
}

func TestToBoolean_NumericTypes(t *testing.T) {
	cases := []struct {
		v    any
		want bool
		name string
	}{
		{int8(1), true, "int8 nonzero"},
		{int8(0), false, "int8 zero"},
		{int16(1), true, "int16 nonzero"},
		{int16(0), false, "int16 zero"},
		{int32(1), true, "int32 nonzero"},
		{int32(0), false, "int32 zero"},
		{int64(1), true, "int64 nonzero"},
		{int64(0), false, "int64 zero"},
		{uint(1), true, "uint nonzero"},
		{uint(0), false, "uint zero"},
		{uint8(1), true, "uint8 nonzero"},
		{uint8(0), false, "uint8 zero"},
		{uint16(1), true, "uint16 nonzero"},
		{uint16(0), false, "uint16 zero"},
		{uint32(1), true, "uint32 nonzero"},
		{uint32(0), false, "uint32 zero"},
		{uint64(1), true, "uint64 nonzero"},
		{uint64(0), false, "uint64 zero"},
		{float32(1.0), true, "float32 nonzero"},
		{float32(0.0), false, "float32 zero"},
		{float64(1.0), true, "float64 nonzero"},
		{float64(0.0), false, "float64 zero"},
		{[]byte{}, false, "unsupported type"},
	}
	for _, c := range cases {
		got := functions.ToBoolean(c.v)
		if got != c.want {
			t.Errorf("ToBoolean(%s=%v): got %v, want %v", c.name, c.v, got, c.want)
		}
	}
}

func TestToInt32(t *testing.T) {
	if got := functions.ToInt32(42); got != 42 {
		t.Errorf("ToInt32(42) = %d, want 42", got)
	}
	if got := functions.ToInt32("100"); got != 100 {
		t.Errorf("ToInt32('100') = %d, want 100", got)
	}
	if got := functions.ToInt32(nil); got != 0 {
		t.Errorf("ToInt32(nil) = %d, want 0", got)
	}
}

func TestToInt64_AllTypes(t *testing.T) {
	cases := []struct {
		v    any
		want int64
		name string
	}{
		{int64(100), int64(100), "int64"},
		{int(200), int64(200), "int"},
		{int8(10), int64(10), "int8"},
		{int16(20), int64(20), "int16"},
		{int32(30), int64(30), "int32"},
		{uint(40), int64(40), "uint"},
		{uint8(50), int64(50), "uint8"},
		{uint16(60), int64(60), "uint16"},
		{uint32(70), int64(70), "uint32"},
		{uint64(80), int64(80), "uint64"},
		{float32(3.7), int64(3), "float32"},
		{float64(9.9), int64(9), "float64"},
		{true, int64(1), "bool true"},
		{false, int64(0), "bool false"},
		{nil, int64(0), "nil"},
		{"999", int64(999), "string valid"},
		{"bad", int64(0), "string invalid"},
		{[]byte{}, int64(0), "unsupported type"},
	}
	for _, c := range cases {
		got := functions.ToInt64(c.v)
		if got != c.want {
			t.Errorf("ToInt64(%s=%v): got %d, want %d", c.name, c.v, got, c.want)
		}
	}
}

func TestToDouble(t *testing.T) {
	if got := functions.ToDouble(3.14); got != 3.14 {
		t.Errorf("ToDouble(3.14) = %v, want 3.14", got)
	}
	if got := functions.ToDouble(5); got != 5.0 {
		t.Errorf("ToDouble(5) = %v, want 5.0", got)
	}
}

func TestToDecimal(t *testing.T) {
	if got := functions.ToDecimal(2.71); got != 2.71 {
		t.Errorf("ToDecimal(2.71) = %v, want 2.71", got)
	}
	if got := functions.ToDecimal("1.5"); got != 1.5 {
		t.Errorf("ToDecimal('1.5') = %v, want 1.5", got)
	}
}
