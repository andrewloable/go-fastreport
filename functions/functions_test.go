package functions_test

import (
	"testing"
	"time"

	"github.com/andrewloable/go-fastreport/functions"
)

// ── Math ──────────────────────────────────────────────────────────────────────

func TestMaxInt(t *testing.T) {
	cases := [][3]int{{1, 2, 2}, {5, 3, 5}, {-1, -1, -1}}
	for _, c := range cases {
		if got := functions.MaxInt(c[0], c[1]); got != c[2] {
			t.Errorf("MaxInt(%d,%d)=%d want %d", c[0], c[1], got, c[2])
		}
	}
}

func TestMinInt(t *testing.T) {
	cases := [][3]int{{1, 2, 1}, {5, 3, 3}, {-1, -1, -1}}
	for _, c := range cases {
		if got := functions.MinInt(c[0], c[1]); got != c[2] {
			t.Errorf("MinInt(%d,%d)=%d want %d", c[0], c[1], got, c[2])
		}
	}
}

func TestMaxFloat(t *testing.T) {
	if got := functions.MaxFloat(1.5, 2.5); got != 2.5 {
		t.Errorf("MaxFloat: got %v, want 2.5", got)
	}
}

func TestMinFloat(t *testing.T) {
	if got := functions.MinFloat(1.5, 2.5); got != 1.5 {
		t.Errorf("MinFloat: got %v, want 1.5", got)
	}
}

func TestAbs(t *testing.T) {
	if got := functions.Abs(-5.5); got != 5.5 {
		t.Errorf("Abs(-5.5)=%v want 5.5", got)
	}
	if got := functions.Abs(3.0); got != 3.0 {
		t.Errorf("Abs(3)=%v want 3", got)
	}
}

func TestRound(t *testing.T) {
	cases := []struct{ v, want float64 }{
		{1.4, 1}, {1.5, 2}, {-1.5, -2}, {2.0, 2},
	}
	for _, c := range cases {
		if got := functions.Round(c.v); got != c.want {
			t.Errorf("Round(%v)=%v want %v", c.v, got, c.want)
		}
	}
}

func TestRoundTo(t *testing.T) {
	if got := functions.RoundTo(3.14159, 2); got != 3.14 {
		t.Errorf("RoundTo(3.14159,2)=%v want 3.14", got)
	}
}

func TestCeiling(t *testing.T) {
	if got := functions.Ceiling(1.1); got != 2.0 {
		t.Errorf("Ceiling(1.1)=%v want 2", got)
	}
}

func TestFloor(t *testing.T) {
	if got := functions.Floor(1.9); got != 1.0 {
		t.Errorf("Floor(1.9)=%v want 1", got)
	}
}

// ── String ────────────────────────────────────────────────────────────────────

func TestLength(t *testing.T) {
	if functions.Length("hello") != 5 {
		t.Error("Length(hello) != 5")
	}
	if functions.Length("") != 0 {
		t.Error("Length('') != 0")
	}
	// Unicode rune count.
	if functions.Length("héllo") != 5 {
		t.Errorf("Length(héllo) = %d, want 5", functions.Length("héllo"))
	}
}

func TestLowerCase(t *testing.T) {
	if got := functions.LowerCase("HELLO"); got != "hello" {
		t.Errorf("LowerCase: got %q", got)
	}
}

func TestUpperCase(t *testing.T) {
	if got := functions.UpperCase("hello"); got != "HELLO" {
		t.Errorf("UpperCase: got %q", got)
	}
}

func TestTitleCase(t *testing.T) {
	if got := functions.TitleCase("hello world"); got != "Hello World" {
		t.Errorf("TitleCase: got %q, want 'Hello World'", got)
	}
}

func TestTrim(t *testing.T) {
	if got := functions.Trim("  hi  "); got != "hi" {
		t.Errorf("Trim: got %q", got)
	}
}

func TestPadLeft(t *testing.T) {
	if got := functions.PadLeft("hi", 5); got != "   hi" {
		t.Errorf("PadLeft: got %q, want '   hi'", got)
	}
	if got := functions.PadLeft("hello", 3); got != "hello" {
		t.Errorf("PadLeft no-op: got %q", got)
	}
}

func TestPadLeftChar(t *testing.T) {
	if got := functions.PadLeftChar("5", 4, '0'); got != "0005" {
		t.Errorf("PadLeftChar: got %q, want 0005", got)
	}
}

func TestPadRight(t *testing.T) {
	if got := functions.PadRight("hi", 5); got != "hi   " {
		t.Errorf("PadRight: got %q", got)
	}
}

func TestPadRightChar(t *testing.T) {
	if got := functions.PadRightChar("A", 3, '-'); got != "A--" {
		t.Errorf("PadRightChar: got %q", got)
	}
}

func TestInsert(t *testing.T) {
	if got := functions.Insert("helo", 3, "l"); got != "hello" {
		t.Errorf("Insert: got %q, want hello", got)
	}
}

func TestInsert_AtStart(t *testing.T) {
	if got := functions.Insert("world", 0, "hello "); got != "hello world" {
		t.Errorf("Insert at start: got %q", got)
	}
}

func TestInsert_AtEnd(t *testing.T) {
	if got := functions.Insert("hello", 5, "!"); got != "hello!" {
		t.Errorf("Insert at end: got %q", got)
	}
}

func TestRemove(t *testing.T) {
	if got := functions.Remove("hello world", 5); got != "hello" {
		t.Errorf("Remove: got %q, want hello", got)
	}
}

func TestRemoveCount(t *testing.T) {
	if got := functions.RemoveCount("hello world", 5, 6); got != "hello" {
		t.Errorf("RemoveCount: got %q, want hello", got)
	}
}

func TestReplace(t *testing.T) {
	if got := functions.Replace("hello world", "world", "Go"); got != "hello Go" {
		t.Errorf("Replace: got %q", got)
	}
}

func TestSubstring(t *testing.T) {
	if got := functions.Substring("hello", 2); got != "llo" {
		t.Errorf("Substring: got %q, want llo", got)
	}
}

func TestSubstringLen(t *testing.T) {
	if got := functions.SubstringLen("hello world", 6, 5); got != "world" {
		t.Errorf("SubstringLen: got %q, want world", got)
	}
}

func TestContains(t *testing.T) {
	if !functions.Contains("hello", "ell") {
		t.Error("Contains: expected true")
	}
	if functions.Contains("hello", "xyz") {
		t.Error("Contains: expected false")
	}
}

func TestIndexOf(t *testing.T) {
	if got := functions.IndexOf("hello", "ll"); got != 2 {
		t.Errorf("IndexOf: got %d, want 2", got)
	}
	if got := functions.IndexOf("hello", "xyz"); got != -1 {
		t.Errorf("IndexOf not found: got %d, want -1", got)
	}
}

func TestAsc(t *testing.T) {
	if got := functions.Asc("A"); got != 65 {
		t.Errorf("Asc(A) = %d, want 65", got)
	}
	if got := functions.Asc(""); got != 0 {
		t.Errorf("Asc('') = %d, want 0", got)
	}
}

func TestChr(t *testing.T) {
	if got := functions.Chr(65); got != "A" {
		t.Errorf("Chr(65) = %q, want A", got)
	}
}

// ── Date / time ───────────────────────────────────────────────────────────────

func TestAddDays(t *testing.T) {
	base := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	got := functions.AddDays(base, 5)
	if got.Day() != 6 {
		t.Errorf("AddDays: got day %d, want 6", got.Day())
	}
}

func TestAddMonths(t *testing.T) {
	// Jan 1 + 1 month = Feb 1 (no day overflow).
	base := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	got := functions.AddMonths(base, 1)
	if got.Month() != time.February {
		t.Errorf("AddMonths: got %v, want February", got.Month())
	}
}

func TestAddYears(t *testing.T) {
	base := time.Date(2020, 3, 1, 0, 0, 0, 0, time.UTC)
	got := functions.AddYears(base, 4)
	if got.Year() != 2024 {
		t.Errorf("AddYears: got %d, want 2024", got.Year())
	}
}

func TestDateSerial(t *testing.T) {
	d := functions.DateSerial(2024, 6, 15)
	if d.Year() != 2024 || d.Month() != 6 || d.Day() != 15 {
		t.Errorf("DateSerial: got %v", d)
	}
}

func TestDayMonthYear(t *testing.T) {
	d := time.Date(2024, 7, 20, 0, 0, 0, 0, time.UTC)
	if functions.Day(d) != 20 {
		t.Errorf("Day: got %d", functions.Day(d))
	}
	if functions.Month(d) != 7 {
		t.Errorf("Month: got %d", functions.Month(d))
	}
	if functions.Year(d) != 2024 {
		t.Errorf("Year: got %d", functions.Year(d))
	}
}

func TestHourMinuteSecond(t *testing.T) {
	d := time.Date(2024, 1, 1, 14, 30, 45, 0, time.UTC)
	if functions.Hour(d) != 14 {
		t.Errorf("Hour: got %d", functions.Hour(d))
	}
	if functions.Minute(d) != 30 {
		t.Errorf("Minute: got %d", functions.Minute(d))
	}
	if functions.Second(d) != 45 {
		t.Errorf("Second: got %d", functions.Second(d))
	}
}

func TestDayOfWeek(t *testing.T) {
	// 2024-01-01 is a Monday.
	d := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	if got := functions.DayOfWeek(d); got != "Monday" {
		t.Errorf("DayOfWeek: got %q, want Monday", got)
	}
}

func TestDaysInMonth(t *testing.T) {
	if got := functions.DaysInMonth(2024, 2); got != 29 { // leap year
		t.Errorf("DaysInMonth(2024,2) = %d, want 29", got)
	}
	if got := functions.DaysInMonth(2023, 2); got != 28 {
		t.Errorf("DaysInMonth(2023,2) = %d, want 28", got)
	}
}

func TestMonthName(t *testing.T) {
	if got := functions.MonthName(1); got != "January" {
		t.Errorf("MonthName(1) = %q, want January", got)
	}
	if got := functions.MonthName(0); got != "" {
		t.Errorf("MonthName(0) should be empty, got %q", got)
	}
}

// ── Formatting ────────────────────────────────────────────────────────────────

func TestFormatNumber(t *testing.T) {
	if got := functions.FormatNumber(1234.5678, 2); got != "1234.57" {
		t.Errorf("FormatNumber: got %q, want 1234.57", got)
	}
}

func TestFormatCurrency(t *testing.T) {
	if got := functions.FormatCurrency(9.5); got != "$9.50" {
		t.Errorf("FormatCurrency: got %q, want $9.50", got)
	}
}

func TestFormatPercent(t *testing.T) {
	if got := functions.FormatPercent(0.1234); got != "12.34%" {
		t.Errorf("FormatPercent: got %q, want 12.34%%", got)
	}
}

func TestFormatDateTime(t *testing.T) {
	d := time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC)
	got := functions.FormatDateTime(d, "2006-01-02")
	if got != "2024-01-15" {
		t.Errorf("FormatDateTime: got %q, want 2024-01-15", got)
	}
}

func TestFormatDateTime_EmptyLayout(t *testing.T) {
	d := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)
	got := functions.FormatDateTime(d, "")
	if got == "" {
		t.Error("FormatDateTime with empty layout should not return empty")
	}
}

// ── Control flow ──────────────────────────────────────────────────────────────

func TestIIF_True(t *testing.T) {
	if got := functions.IIF(true, "yes", "no"); got != "yes" {
		t.Errorf("IIF(true): got %v", got)
	}
}

func TestIIF_False(t *testing.T) {
	if got := functions.IIF(false, "yes", "no"); got != "no" {
		t.Errorf("IIF(false): got %v", got)
	}
}

func TestChoose(t *testing.T) {
	if got := functions.Choose(2, "a", "b", "c"); got != "b" {
		t.Errorf("Choose(2,...): got %v", got)
	}
}

func TestChoose_OutOfRange(t *testing.T) {
	if got := functions.Choose(5, "a", "b"); got != nil {
		t.Errorf("Choose out of range: got %v, want nil", got)
	}
	if got := functions.Choose(0, "a"); got != nil {
		t.Errorf("Choose(0): got %v, want nil", got)
	}
}

func TestSwitch_FirstMatch(t *testing.T) {
	x := 3
	got := functions.Switch(x == 1, "one", x == 2, "two", x == 3, "three")
	if got != "three" {
		t.Errorf("Switch: got %v, want three", got)
	}
}

func TestSwitch_NoMatch(t *testing.T) {
	if got := functions.Switch(false, "a", false, "b"); got != nil {
		t.Errorf("Switch no match: got %v, want nil", got)
	}
}

// ── NumToWords ────────────────────────────────────────────────────────────────

func TestNumToWords(t *testing.T) {
	cases := []struct {
		n    int64
		want string
	}{
		{0, "zero"},
		{1, "one"},
		{10, "ten"},
		{11, "eleven"},
		{15, "fifteen"},
		{20, "twenty"},
		{21, "twenty-one"},
		{99, "ninety-nine"},
		{100, "one hundred"},
		{101, "one hundred one"},
		{123, "one hundred twenty-three"},
		{1000, "one thousand"},
		{1001, "one thousand one"},
		{1000000, "one million"},
		{-5, "negative five"},
	}
	for _, c := range cases {
		got := functions.NumToWords(c.n)
		if got != c.want {
			t.Errorf("NumToWords(%d) = %q, want %q", c.n, got, c.want)
		}
	}
}

func TestNumToWordsFloat(t *testing.T) {
	cases := []struct {
		v    float64
		want string
	}{
		{0, "zero"},
		{1.0, "one"},
		{1.50, "one and 50/100"},
		{-2.25, "negative two and 25/100"},
	}
	for _, c := range cases {
		got := functions.NumToWordsFloat(c.v)
		if got != c.want {
			t.Errorf("NumToWordsFloat(%v) = %q, want %q", c.v, got, c.want)
		}
	}
}

// ── Roman ─────────────────────────────────────────────────────────────────────

func TestToRoman(t *testing.T) {
	cases := []struct {
		n    int
		want string
	}{
		{1, "I"}, {4, "IV"}, {5, "V"}, {9, "IX"},
		{14, "XIV"}, {40, "XL"}, {90, "XC"},
		{400, "CD"}, {900, "CM"},
		{1994, "MCMXCIV"}, {3999, "MMMCMXCIX"},
	}
	for _, c := range cases {
		got, err := functions.ToRoman(c.n)
		if err != nil {
			t.Errorf("ToRoman(%d): unexpected error: %v", c.n, err)
			continue
		}
		if got != c.want {
			t.Errorf("ToRoman(%d) = %q, want %q", c.n, got, c.want)
		}
	}
}

func TestToRoman_OutOfRange(t *testing.T) {
	if _, err := functions.ToRoman(0); err == nil {
		t.Error("expected error for 0")
	}
	if _, err := functions.ToRoman(4000); err == nil {
		t.Error("expected error for 4000")
	}
	if _, err := functions.ToRoman(-1); err == nil {
		t.Error("expected error for -1")
	}
}

func TestMustToRoman(t *testing.T) {
	if got := functions.MustToRoman(42); got != "XLII" {
		t.Errorf("MustToRoman(42) = %q, want XLII", got)
	}
}

func TestMustToRoman_Panics(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic for out-of-range value")
		}
	}()
	functions.MustToRoman(0) // should panic
}

func TestFromRoman(t *testing.T) {
	cases := []struct {
		s    string
		want int
	}{
		{"I", 1}, {"IV", 4}, {"IX", 9}, {"XIV", 14},
		{"MCMXCIV", 1994}, {"MMMCMXCIX", 3999},
	}
	for _, c := range cases {
		got, err := functions.FromRoman(c.s)
		if err != nil {
			t.Errorf("FromRoman(%q): unexpected error: %v", c.s, err)
			continue
		}
		if got != c.want {
			t.Errorf("FromRoman(%q) = %d, want %d", c.s, got, c.want)
		}
	}
}

func TestFromRoman_InvalidChar(t *testing.T) {
	if _, err := functions.FromRoman("ABC"); err == nil {
		t.Error("expected error for invalid roman numeral")
	}
}

func TestFromRoman_Empty(t *testing.T) {
	if _, err := functions.FromRoman(""); err == nil {
		t.Error("expected error for empty string")
	}
}

func TestFromRoman_LowerCase(t *testing.T) {
	// FromRoman should handle lowercase by converting to upper.
	got, err := functions.FromRoman("xiv")
	if err != nil {
		t.Fatalf("FromRoman(xiv) error: %v", err)
	}
	if got != 14 {
		t.Errorf("FromRoman(xiv) = %d, want 14", got)
	}
}

// ── All() map ─────────────────────────────────────────────────────────────────

func TestAll_ContainsExpectedFunctions(t *testing.T) {
	all := functions.All()
	required := []string{
		"MaxInt", "MinInt", "Abs", "Round", "Ceiling", "Floor",
		"Length", "LowerCase", "UpperCase", "TitleCase", "Trim",
		"PadLeft", "PadRight", "Insert", "Remove", "Replace",
		"Substring", "Contains", "IndexOf", "Asc", "Chr",
		"AddDays", "AddMonths", "AddYears", "DateSerial",
		"Day", "Month", "Year", "Hour", "Minute", "Second",
		"DayOfWeek", "DaysInMonth", "MonthName",
		"FormatNumber", "FormatCurrency", "FormatPercent",
		"IIF", "Choose", "Switch",
		"NumToWords", "Roman",
	}
	for _, name := range required {
		if _, ok := all[name]; !ok {
			t.Errorf("All() missing function %q", name)
		}
	}
}
