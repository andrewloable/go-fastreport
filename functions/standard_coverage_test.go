package functions_test

import (
	"math"
	"testing"
	"time"

	"github.com/andrewloable/go-fastreport/functions"
)

// ── MaxFloat additional branches ─────────────────────────────────────────────

func TestMaxFloat_Equal(t *testing.T) {
	if got := functions.MaxFloat(3.5, 3.5); got != 3.5 {
		t.Errorf("MaxFloat(3.5,3.5) = %v, want 3.5", got)
	}
}

func TestMaxFloat_Negative(t *testing.T) {
	if got := functions.MaxFloat(-1.0, -2.0); got != -1.0 {
		t.Errorf("MaxFloat(-1,-2) = %v, want -1", got)
	}
}

func TestMaxFloat_NaN(t *testing.T) {
	nan := math.NaN()
	got := functions.MaxFloat(nan, 1.0)
	// NaN comparisons: nan > 1.0 is false, so returns b (1.0)
	if got != 1.0 {
		t.Errorf("MaxFloat(NaN,1.0) = %v, want 1.0", got)
	}
}

// ── MinFloat additional branches ─────────────────────────────────────────────

func TestMinFloat_Equal(t *testing.T) {
	if got := functions.MinFloat(2.2, 2.2); got != 2.2 {
		t.Errorf("MinFloat(2.2,2.2) = %v, want 2.2", got)
	}
}

func TestMinFloat_Negative(t *testing.T) {
	if got := functions.MinFloat(-3.0, -1.0); got != -3.0 {
		t.Errorf("MinFloat(-3,-1) = %v, want -3", got)
	}
}

func TestMinFloat_NaN(t *testing.T) {
	nan := math.NaN()
	got := functions.MinFloat(nan, 1.0)
	// NaN comparisons: nan < 1.0 is false, so returns b (1.0)
	if got != 1.0 {
		t.Errorf("MinFloat(NaN,1.0) = %v, want 1.0", got)
	}
}

// ── PadRightChar — string longer than requested width ─────────────────────────

func TestPadRightChar_AlreadyLong(t *testing.T) {
	if got := functions.PadRightChar("hello", 3, '-'); got != "hello" {
		t.Errorf("PadRightChar no-op: got %q, want hello", got)
	}
}

func TestPadRightChar_ExactLength(t *testing.T) {
	if got := functions.PadRightChar("abc", 3, '-'); got != "abc" {
		t.Errorf("PadRightChar exact: got %q, want abc", got)
	}
}

// ── Insert — boundary paths ───────────────────────────────────────────────────

func TestInsert_NegativeIndex(t *testing.T) {
	// negative index clamps to 0 (insert at beginning)
	if got := functions.Insert("world", -5, "hello "); got != "hello world" {
		t.Errorf("Insert negative index: got %q, want 'hello world'", got)
	}
}

func TestInsert_IndexBeyondLength(t *testing.T) {
	// index beyond length clamps to end
	if got := functions.Insert("hello", 100, "!"); got != "hello!" {
		t.Errorf("Insert beyond length: got %q, want 'hello!'", got)
	}
}

// ── Remove — boundary paths ───────────────────────────────────────────────────

func TestRemove_NegativeIndex(t *testing.T) {
	// negative index clamps to 0 → removes everything
	if got := functions.Remove("hello", -3); got != "" {
		t.Errorf("Remove negative index: got %q, want ''", got)
	}
}

func TestRemove_AtZero(t *testing.T) {
	if got := functions.Remove("hello", 0); got != "" {
		t.Errorf("Remove at 0: got %q, want ''", got)
	}
}

func TestRemove_IndexAtLength(t *testing.T) {
	// startIndex == len → returns s unchanged
	if got := functions.Remove("hello", 5); got != "hello" {
		t.Errorf("Remove at length: got %q, want 'hello'", got)
	}
}

func TestRemove_IndexBeyondLength(t *testing.T) {
	if got := functions.Remove("hello", 100); got != "hello" {
		t.Errorf("Remove beyond length: got %q, want 'hello'", got)
	}
}

// ── RemoveCount — boundary paths ──────────────────────────────────────────────

func TestRemoveCount_CountExceedsRemaining(t *testing.T) {
	// count extends past end; should clamp to end of string
	if got := functions.RemoveCount("hello", 2, 100); got != "he" {
		t.Errorf("RemoveCount exceeds: got %q, want 'he'", got)
	}
}

func TestRemoveCount_StartIndexAtLength(t *testing.T) {
	// startIndex >= n → returns s unchanged
	if got := functions.RemoveCount("hello", 5, 2); got != "hello" {
		t.Errorf("RemoveCount startIndex at length: got %q, want 'hello'", got)
	}
}

func TestRemoveCount_NegativeStartIndex(t *testing.T) {
	// negative startIndex clamps to 0
	if got := functions.RemoveCount("hello", -1, 2); got != "llo" {
		t.Errorf("RemoveCount negative start: got %q, want 'llo'", got)
	}
}

// ── Substring — boundary paths ────────────────────────────────────────────────

func TestSubstring_StartZero(t *testing.T) {
	if got := functions.Substring("hello", 0); got != "hello" {
		t.Errorf("Substring(0): got %q, want 'hello'", got)
	}
}

func TestSubstring_StartBeyondLength(t *testing.T) {
	if got := functions.Substring("hello", 100); got != "" {
		t.Errorf("Substring beyond length: got %q, want ''", got)
	}
}

func TestSubstring_NegativeStart(t *testing.T) {
	// negative start clamps to 0
	if got := functions.Substring("hello", -5); got != "hello" {
		t.Errorf("Substring negative start: got %q, want 'hello'", got)
	}
}

// ── SubstringLen — boundary paths ─────────────────────────────────────────────

func TestSubstringLen_LengthExceedsRemaining(t *testing.T) {
	if got := functions.SubstringLen("hello", 3, 100); got != "lo" {
		t.Errorf("SubstringLen exceeds: got %q, want 'lo'", got)
	}
}

func TestSubstringLen_StartBeyondLength(t *testing.T) {
	if got := functions.SubstringLen("hello", 100, 2); got != "" {
		t.Errorf("SubstringLen start beyond: got %q, want ''", got)
	}
}

func TestSubstringLen_NegativeStart(t *testing.T) {
	// negative start clamps to 0
	if got := functions.SubstringLen("hello", -1, 3); got != "hel" {
		t.Errorf("SubstringLen negative start: got %q, want 'hel'", got)
	}
}

// ── DateDiff — all interval types ─────────────────────────────────────────────

func TestDateDiff_Hours(t *testing.T) {
	d1 := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	d2 := time.Date(2024, 1, 1, 6, 0, 0, 0, time.UTC)
	if got := functions.DateDiff("h", d1, d2); got != 6.0 {
		t.Errorf("DateDiff 'h': got %v, want 6", got)
	}
	if got := functions.DateDiff("hour", d1, d2); got != 6.0 {
		t.Errorf("DateDiff 'hour': got %v, want 6", got)
	}
	if got := functions.DateDiff("hours", d1, d2); got != 6.0 {
		t.Errorf("DateDiff 'hours': got %v, want 6", got)
	}
}

func TestDateDiff_Minutes(t *testing.T) {
	d1 := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	d2 := time.Date(2024, 1, 1, 0, 30, 0, 0, time.UTC)
	cases := []string{"m", "minute", "minutes", "n"}
	for _, interval := range cases {
		if got := functions.DateDiff(interval, d1, d2); got != 30.0 {
			t.Errorf("DateDiff %q: got %v, want 30", interval, got)
		}
	}
}

func TestDateDiff_Seconds(t *testing.T) {
	d1 := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	d2 := time.Date(2024, 1, 1, 0, 0, 45, 0, time.UTC)
	cases := []string{"s", "second", "seconds"}
	for _, interval := range cases {
		if got := functions.DateDiff(interval, d1, d2); got != 45.0 {
			t.Errorf("DateDiff %q: got %v, want 45", interval, got)
		}
	}
}

func TestDateDiff_YyyyVariant(t *testing.T) {
	// anniversary not yet reached in year
	d1 := time.Date(2020, 6, 15, 0, 0, 0, 0, time.UTC)
	d2 := time.Date(2024, 3, 1, 0, 0, 0, 0, time.UTC)
	// 4 years apart but anniversary (Jun 15) not reached → 3
	if got := functions.DateDiff("yyyy", d1, d2); got != 3.0 {
		t.Errorf("DateDiff 'yyyy': got %v, want 3", got)
	}
}

func TestDateDiff_YyyyAnniversaryReached(t *testing.T) {
	d1 := time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)
	d2 := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	if got := functions.DateDiff("yyyy", d1, d2); got != 4.0 {
		t.Errorf("DateDiff 'yyyy' anniversary: got %v, want 4", got)
	}
}

func TestDateDiff_MmVariant(t *testing.T) {
	d1 := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	d2 := time.Date(2024, 4, 1, 0, 0, 0, 0, time.UTC)
	if got := functions.DateDiff("mm", d1, d2); got != 3.0 {
		t.Errorf("DateDiff 'mm': got %v, want 3", got)
	}
}

func TestDateDiff_UnknownInterval(t *testing.T) {
	// unknown interval defaults to days
	d1 := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	d2 := time.Date(2024, 1, 6, 0, 0, 0, 0, time.UTC)
	if got := functions.DateDiff("unknown", d1, d2); got != 5.0 {
		t.Errorf("DateDiff unknown: got %v, want 5 (days default)", got)
	}
}

func TestDateDiff_CaseInsensitive(t *testing.T) {
	d1 := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	d2 := time.Date(2024, 1, 1, 2, 0, 0, 0, time.UTC)
	if got := functions.DateDiff("H", d1, d2); got != 2.0 {
		t.Errorf("DateDiff 'H' (uppercase): got %v, want 2", got)
	}
}

// ── ToInt — all type branches ─────────────────────────────────────────────────

func TestToInt_AllTypes(t *testing.T) {
	cases := []struct {
		input any
		want  int
		name  string
	}{
		{int8(10), 10, "int8"},
		{int16(200), 200, "int16"},
		{int32(300), 300, "int32"},
		{int64(400), 400, "int64"},
		{uint(5), 5, "uint"},
		{uint8(6), 6, "uint8"},
		{uint16(7), 7, "uint16"},
		{uint32(8), 8, "uint32"},
		{uint64(9), 9, "uint64"},
		{float32(3.9), 3, "float32"},
		{float64(7.7), 7, "float64"},
		{true, 1, "bool true"},
		{false, 0, "bool false"},
		{nil, 0, "nil"},
		{"42", 42, "string"},
		{"abc", 0, "string invalid"},
		{[]byte{}, 0, "unsupported type"},
	}
	for _, c := range cases {
		got := functions.ToInt(c.input)
		if got != c.want {
			t.Errorf("ToInt(%s=%v): got %d, want %d", c.name, c.input, got, c.want)
		}
	}
}

// ── ToFloat — all type branches ───────────────────────────────────────────────

func TestToFloat_AllTypes(t *testing.T) {
	cases := []struct {
		input any
		want  float64
		name  string
	}{
		{float32(1.5), float64(float32(1.5)), "float32"},
		{float64(2.5), 2.5, "float64"},
		{int(10), 10.0, "int"},
		{int8(11), 11.0, "int8"},
		{int16(12), 12.0, "int16"},
		{int32(13), 13.0, "int32"},
		{int64(14), 14.0, "int64"},
		{uint(15), 15.0, "uint"},
		{uint8(16), 16.0, "uint8"},
		{uint16(17), 17.0, "uint16"},
		{uint32(18), 18.0, "uint32"},
		{uint64(19), 19.0, "uint64"},
		{true, 1.0, "bool true"},
		{false, 0.0, "bool false"},
		{nil, 0.0, "nil"},
		{"3.14", 3.14, "string"},
		{"bad", 0.0, "string invalid"},
		{[]byte{}, 0.0, "unsupported type"},
	}
	for _, c := range cases {
		got := functions.ToFloat(c.input)
		if got != c.want {
			t.Errorf("ToFloat(%s=%v): got %v, want %v", c.name, c.input, got, c.want)
		}
	}
}
