package format

import (
	"testing"
	"time"
)

// ---------------------------------------------------------------------------
// GeneralFormat
// ---------------------------------------------------------------------------

func TestGeneralFormat_FormatType(t *testing.T) {
	f := NewGeneralFormat()
	if got := f.FormatType(); got != "General" {
		t.Fatalf("FormatType = %q, want %q", got, "General")
	}
}

func TestGeneralFormat_Nil(t *testing.T) {
	f := NewGeneralFormat()
	if got := f.FormatValue(nil); got != "" {
		t.Fatalf("got %q, want %q", got, "")
	}
}

func TestGeneralFormat_String(t *testing.T) {
	f := NewGeneralFormat()
	if got := f.FormatValue("hello"); got != "hello" {
		t.Fatalf("got %q, want %q", got, "hello")
	}
}

func TestGeneralFormat_Int(t *testing.T) {
	f := NewGeneralFormat()
	if got := f.FormatValue(42); got != "42" {
		t.Fatalf("got %q, want %q", got, "42")
	}
}

func TestGeneralFormat_Float(t *testing.T) {
	f := NewGeneralFormat()
	// fmt.Sprint(3.14) == "3.14"
	if got := f.FormatValue(3.14); got != "3.14" {
		t.Fatalf("got %q, want %q", got, "3.14")
	}
}

func TestGeneralFormat_ImplementsFormat(t *testing.T) {
	var _ Format = NewGeneralFormat()
}

// ---------------------------------------------------------------------------
// BooleanFormat
// ---------------------------------------------------------------------------

func TestBooleanFormat_FormatType(t *testing.T) {
	f := NewBooleanFormat()
	if got := f.FormatType(); got != "Boolean" {
		t.Fatalf("FormatType = %q, want %q", got, "Boolean")
	}
}

func TestBooleanFormat_Defaults(t *testing.T) {
	f := NewBooleanFormat()
	if f.TrueText != "True" {
		t.Fatalf("TrueText = %q, want %q", f.TrueText, "True")
	}
	if f.FalseText != "False" {
		t.Fatalf("FalseText = %q, want %q", f.FalseText, "False")
	}
}

func TestBooleanFormat_True(t *testing.T) {
	f := NewBooleanFormat()
	if got := f.FormatValue(true); got != "True" {
		t.Fatalf("got %q, want %q", got, "True")
	}
}

func TestBooleanFormat_False(t *testing.T) {
	f := NewBooleanFormat()
	if got := f.FormatValue(false); got != "False" {
		t.Fatalf("got %q, want %q", got, "False")
	}
}

func TestBooleanFormat_CustomText(t *testing.T) {
	f := &BooleanFormat{TrueText: "Yes", FalseText: "No"}
	if got := f.FormatValue(true); got != "Yes" {
		t.Fatalf("got %q, want %q", got, "Yes")
	}
	if got := f.FormatValue(false); got != "No" {
		t.Fatalf("got %q, want %q", got, "No")
	}
}

func TestBooleanFormat_Nil(t *testing.T) {
	f := NewBooleanFormat()
	if got := f.FormatValue(nil); got != "" {
		t.Fatalf("got %q, want %q", got, "")
	}
}

func TestBooleanFormat_NonBool(t *testing.T) {
	f := NewBooleanFormat()
	if got := f.FormatValue(42); got != "42" {
		t.Fatalf("got %q, want %q", got, "42")
	}
}

func TestBooleanFormat_ImplementsFormat(t *testing.T) {
	var _ Format = NewBooleanFormat()
}

// ---------------------------------------------------------------------------
// CustomFormat
// ---------------------------------------------------------------------------

func TestCustomFormat_FormatType(t *testing.T) {
	f := NewCustomFormat()
	if got := f.FormatType(); got != "Custom" {
		t.Fatalf("FormatType = %q, want %q", got, "Custom")
	}
}

func TestCustomFormat_Default(t *testing.T) {
	f := NewCustomFormat()
	if got := f.FormatValue(42); got != "42" {
		t.Fatalf("got %q, want %q", got, "42")
	}
}

func TestCustomFormat_Nil(t *testing.T) {
	f := NewCustomFormat()
	if got := f.FormatValue(nil); got != "" {
		t.Fatalf("got %q, want %q", got, "")
	}
}

func TestCustomFormat_Printf(t *testing.T) {
	f := &CustomFormat{Format: "%.2f"}
	if got := f.FormatValue(3.14159); got != "3.14" {
		t.Fatalf("got %q, want %q", got, "3.14")
	}
}

func TestCustomFormat_EmptyFormat(t *testing.T) {
	f := &CustomFormat{Format: ""}
	if got := f.FormatValue("hello"); got != "hello" {
		t.Fatalf("got %q, want %q", got, "hello")
	}
}

func TestCustomFormat_StringVerb(t *testing.T) {
	f := &CustomFormat{Format: "%s!"}
	if got := f.FormatValue("world"); got != "world!" {
		t.Fatalf("got %q, want %q", got, "world!")
	}
}

func TestCustomFormat_ImplementsFormat(t *testing.T) {
	var _ Format = NewCustomFormat()
}

// ---------------------------------------------------------------------------
// NumberFormat
// ---------------------------------------------------------------------------

func TestNumberFormat_FormatType(t *testing.T) {
	f := NewNumberFormat()
	if got := f.FormatType(); got != "Number" {
		t.Fatalf("FormatType = %q, want %q", got, "Number")
	}
}

func TestNumberFormat_Defaults(t *testing.T) {
	f := NewNumberFormat()
	if f.DecimalDigits != 2 {
		t.Fatalf("DecimalDigits = %d, want 2", f.DecimalDigits)
	}
	if f.DecimalSeparator != "." {
		t.Fatalf("DecimalSeparator = %q, want %q", f.DecimalSeparator, ".")
	}
	if f.GroupSeparator != "," {
		t.Fatalf("GroupSeparator = %q, want %q", f.GroupSeparator, ",")
	}
}

func TestNumberFormat_Nil(t *testing.T) {
	f := NewNumberFormat()
	if got := f.FormatValue(nil); got != "" {
		t.Fatalf("got %q, want %q", got, "")
	}
}

func TestNumberFormat_NonNumeric(t *testing.T) {
	f := NewNumberFormat()
	if got := f.FormatValue("abc"); got != "abc" {
		t.Fatalf("got %q, want %q", got, "abc")
	}
}

func TestNumberFormat_BasicPositive(t *testing.T) {
	f := &NumberFormat{
		UseLocaleSettings: false,
		DecimalDigits:     2,
		DecimalSeparator:  ".",
		GroupSeparator:    ",",
		NegativePattern:   1,
	}
	if got := f.FormatValue(1234.5); got != "1,234.50" {
		t.Fatalf("got %q, want %q", got, "1,234.50")
	}
}

func TestNumberFormat_NegativePattern0(t *testing.T) {
	f := &NumberFormat{
		UseLocaleSettings: false,
		DecimalDigits:     2,
		DecimalSeparator:  ".",
		GroupSeparator:    ",",
		NegativePattern:   0,
	}
	if got := f.FormatValue(-1234.5); got != "(1,234.50)" {
		t.Fatalf("got %q, want %q", got, "(1,234.50)")
	}
}

func TestNumberFormat_NegativePattern1(t *testing.T) {
	f := &NumberFormat{
		UseLocaleSettings: false,
		DecimalDigits:     2,
		DecimalSeparator:  ".",
		GroupSeparator:    ",",
		NegativePattern:   1,
	}
	if got := f.FormatValue(-1234.5); got != "-1,234.50" {
		t.Fatalf("got %q, want %q", got, "-1,234.50")
	}
}

func TestNumberFormat_NegativePattern2(t *testing.T) {
	f := &NumberFormat{
		UseLocaleSettings: false,
		DecimalDigits:     2,
		DecimalSeparator:  ".",
		GroupSeparator:    ",",
		NegativePattern:   2,
	}
	if got := f.FormatValue(-1234.5); got != "- 1,234.50" {
		t.Fatalf("got %q, want %q", got, "- 1,234.50")
	}
}

func TestNumberFormat_NegativePattern3(t *testing.T) {
	f := &NumberFormat{
		UseLocaleSettings: false,
		DecimalDigits:     2,
		DecimalSeparator:  ".",
		GroupSeparator:    ",",
		NegativePattern:   3,
	}
	if got := f.FormatValue(-1234.5); got != "1,234.50-" {
		t.Fatalf("got %q, want %q", got, "1,234.50-")
	}
}

func TestNumberFormat_NegativePattern4(t *testing.T) {
	f := &NumberFormat{
		UseLocaleSettings: false,
		DecimalDigits:     2,
		DecimalSeparator:  ".",
		GroupSeparator:    ",",
		NegativePattern:   4,
	}
	if got := f.FormatValue(-1234.5); got != "1,234.50 -" {
		t.Fatalf("got %q, want %q", got, "1,234.50 -")
	}
}

func TestNumberFormat_NegativePatternDefault(t *testing.T) {
	// Unknown pattern should fall back to -n.
	f := &NumberFormat{
		UseLocaleSettings: false,
		DecimalDigits:     2,
		DecimalSeparator:  ".",
		GroupSeparator:    ",",
		NegativePattern:   99,
	}
	if got := f.FormatValue(-1.0); got != "-1.00" {
		t.Fatalf("got %q, want %q", got, "-1.00")
	}
}

func TestNumberFormat_ZeroDecimals(t *testing.T) {
	f := &NumberFormat{
		UseLocaleSettings: false,
		DecimalDigits:     0,
		DecimalSeparator:  ".",
		GroupSeparator:    ",",
		NegativePattern:   1,
	}
	if got := f.FormatValue(1234.0); got != "1,234" {
		t.Fatalf("got %q, want %q", got, "1,234")
	}
}

func TestNumberFormat_NoGroupSeparator(t *testing.T) {
	f := &NumberFormat{
		UseLocaleSettings: false,
		DecimalDigits:     2,
		DecimalSeparator:  ".",
		GroupSeparator:    "",
		NegativePattern:   1,
	}
	if got := f.FormatValue(1234567.89); got != "1234567.89" {
		t.Fatalf("got %q, want %q", got, "1234567.89")
	}
}

func TestNumberFormat_SmallNumber(t *testing.T) {
	f := &NumberFormat{
		UseLocaleSettings: false,
		DecimalDigits:     2,
		DecimalSeparator:  ".",
		GroupSeparator:    ",",
		NegativePattern:   1,
	}
	if got := f.FormatValue(5.0); got != "5.00" {
		t.Fatalf("got %q, want %q", got, "5.00")
	}
}

func TestNumberFormat_UseLocale(t *testing.T) {
	// With UseLocaleSettings=true we still produce a valid number string.
	f := NewNumberFormat()
	result := f.FormatValue(1234.5)
	if result == "" {
		t.Fatal("expected non-empty result")
	}
}

func TestNumberFormat_StringNumeric(t *testing.T) {
	f := &NumberFormat{
		UseLocaleSettings: false,
		DecimalDigits:     2,
		DecimalSeparator:  ".",
		GroupSeparator:    ",",
		NegativePattern:   1,
	}
	if got := f.FormatValue("42"); got != "42.00" {
		t.Fatalf("got %q, want %q", got, "42.00")
	}
}

func TestNumberFormat_AllIntTypes(t *testing.T) {
	f := &NumberFormat{
		UseLocaleSettings: false,
		DecimalDigits:     0,
		DecimalSeparator:  ".",
		GroupSeparator:    "",
		NegativePattern:   1,
	}
	cases := []struct {
		v    any
		want string
	}{
		{int8(5), "5"},
		{int16(5), "5"},
		{int32(5), "5"},
		{int64(5), "5"},
		{uint(5), "5"},
		{uint8(5), "5"},
		{uint16(5), "5"},
		{uint32(5), "5"},
		{uint64(5), "5"},
		{float32(5), "5"},
	}
	for _, tc := range cases {
		if got := f.FormatValue(tc.v); got != tc.want {
			t.Fatalf("FormatValue(%T(%v)) = %q, want %q", tc.v, tc.v, got, tc.want)
		}
	}
}

func TestNumberFormat_ImplementsFormat(t *testing.T) {
	var _ Format = NewNumberFormat()
}

// ---------------------------------------------------------------------------
// CurrencyFormat
// ---------------------------------------------------------------------------

func TestCurrencyFormat_FormatType(t *testing.T) {
	f := NewCurrencyFormat()
	if got := f.FormatType(); got != "Currency" {
		t.Fatalf("FormatType = %q, want %q", got, "Currency")
	}
}

func TestCurrencyFormat_Nil(t *testing.T) {
	f := NewCurrencyFormat()
	if got := f.FormatValue(nil); got != "" {
		t.Fatalf("got %q, want %q", got, "")
	}
}

func TestCurrencyFormat_NonNumeric(t *testing.T) {
	f := NewCurrencyFormat()
	if got := f.FormatValue("x"); got != "x" {
		t.Fatalf("got %q, want %q", got, "x")
	}
}

func TestCurrencyFormat_PositivePatterns(t *testing.T) {
	base := &CurrencyFormat{
		UseLocaleSettings: false,
		DecimalDigits:     2,
		DecimalSeparator:  ".",
		GroupSeparator:    ",",
		CurrencySymbol:    "$",
	}
	cases := []struct {
		pattern int
		want    string
	}{
		{0, "$1,234.50"},
		{1, "1,234.50$"},
		{2, "$ 1,234.50"},
		{3, "1,234.50 $"},
		{99, "$1,234.50"}, // default
	}
	for _, tc := range cases {
		f := *base
		f.PositivePattern = tc.pattern
		if got := f.FormatValue(1234.5); got != tc.want {
			t.Fatalf("PositivePattern=%d: got %q, want %q", tc.pattern, got, tc.want)
		}
	}
}

func TestCurrencyFormat_NegativePatterns(t *testing.T) {
	base := &CurrencyFormat{
		UseLocaleSettings: false,
		DecimalDigits:     2,
		DecimalSeparator:  ".",
		GroupSeparator:    ",",
		CurrencySymbol:    "$",
	}
	cases := []struct {
		pattern int
		want    string
	}{
		{0, "($1,234.50)"},
		{1, "-$1,234.50"},
		{2, "$-1,234.50"},
		{3, "$1,234.50-"},
		{4, "(1,234.50$)"},
		{5, "-1,234.50$"},
		{6, "1,234.50-$"},
		{7, "1,234.50$-"},
		{8, "-1,234.50 $"},
		{9, "-$ 1,234.50"},
		{10, "1,234.50 $-"},
		{11, "$ 1,234.50-"},
		{12, "$ -1,234.50"},
		{13, "1,234.50- $"},
		{14, "($ 1,234.50)"},
		{15, "(1,234.50 $)"},
		{99, "-$1,234.50"}, // default
	}
	for _, tc := range cases {
		f := *base
		f.NegativePattern = tc.pattern
		if got := f.FormatValue(-1234.5); got != tc.want {
			t.Fatalf("NegativePattern=%d: got %q, want %q", tc.pattern, got, tc.want)
		}
	}
}

func TestCurrencyFormat_UseLocale(t *testing.T) {
	f := NewCurrencyFormat()
	result := f.FormatValue(100.0)
	if result == "" {
		t.Fatal("expected non-empty result")
	}
}

func TestCurrencyFormat_ImplementsFormat(t *testing.T) {
	var _ Format = NewCurrencyFormat()
}

// ---------------------------------------------------------------------------
// PercentFormat
// ---------------------------------------------------------------------------

func TestPercentFormat_FormatType(t *testing.T) {
	f := NewPercentFormat()
	if got := f.FormatType(); got != "Percent" {
		t.Fatalf("FormatType = %q, want %q", got, "Percent")
	}
}

func TestPercentFormat_Nil(t *testing.T) {
	f := NewPercentFormat()
	if got := f.FormatValue(nil); got != "" {
		t.Fatalf("got %q, want %q", got, "")
	}
}

func TestPercentFormat_NonNumeric(t *testing.T) {
	f := NewPercentFormat()
	if got := f.FormatValue("abc"); got != "abc" {
		t.Fatalf("got %q, want %q", got, "abc")
	}
}

func TestPercentFormat_BasicPositive(t *testing.T) {
	f := &PercentFormat{
		UseLocaleSettings: false,
		DecimalDigits:     2,
		DecimalSeparator:  ".",
		GroupSeparator:    ",",
		PercentSymbol:     "%",
		PositivePattern:   1, // n%
	}
	// 0.25 * 100 = 25.00%
	if got := f.FormatValue(0.25); got != "25.00%" {
		t.Fatalf("got %q, want %q", got, "25.00%")
	}
}

func TestPercentFormat_PositivePatterns(t *testing.T) {
	base := &PercentFormat{
		UseLocaleSettings: false,
		DecimalDigits:     0,
		DecimalSeparator:  ".",
		GroupSeparator:    ",",
		PercentSymbol:     "%",
	}
	cases := []struct {
		pattern int
		want    string
	}{
		{0, "25 %"},
		{1, "25%"},
		{2, "%25"},
		{3, "% 25"},
		{99, "25 %"}, // default
	}
	for _, tc := range cases {
		f := *base
		f.PositivePattern = tc.pattern
		if got := f.FormatValue(0.25); got != tc.want {
			t.Fatalf("PositivePattern=%d: got %q, want %q", tc.pattern, got, tc.want)
		}
	}
}

func TestPercentFormat_NegativePatterns(t *testing.T) {
	base := &PercentFormat{
		UseLocaleSettings: false,
		DecimalDigits:     0,
		DecimalSeparator:  ".",
		GroupSeparator:    ",",
		PercentSymbol:     "%",
	}
	cases := []struct {
		pattern int
		want    string
	}{
		{0, "-25 %"},
		{1, "-25%"},
		{2, "-%25"},
		{3, "%-25"},
		{4, "%25-"},
		{5, "25-%"},
		{6, "25%-"},
		{7, "-%25"},
		{8, "25 %-"},
		{9, "% 25-"},
		{10, "% -25"},
		{11, "25- %"},
		{99, "-25 %"}, // default
	}
	for _, tc := range cases {
		f := *base
		f.NegativePattern = tc.pattern
		if got := f.FormatValue(-0.25); got != tc.want {
			t.Fatalf("NegativePattern=%d: got %q, want %q", tc.pattern, got, tc.want)
		}
	}
}

func TestPercentFormat_UseLocale(t *testing.T) {
	f := NewPercentFormat()
	result := f.FormatValue(0.5)
	if result == "" {
		t.Fatal("expected non-empty result")
	}
}

func TestPercentFormat_ImplementsFormat(t *testing.T) {
	var _ Format = NewPercentFormat()
}

// ---------------------------------------------------------------------------
// DateFormat
// ---------------------------------------------------------------------------

func TestDateFormat_FormatType(t *testing.T) {
	f := NewDateFormat()
	if got := f.FormatType(); got != "Date" {
		t.Fatalf("FormatType = %q, want %q", got, "Date")
	}
}

func TestDateFormat_Nil(t *testing.T) {
	f := NewDateFormat()
	if got := f.FormatValue(nil); got != "" {
		t.Fatalf("got %q, want %q", got, "")
	}
}

func TestDateFormat_TimeTime(t *testing.T) {
	// Default format is C# "d" (short date = M/d/yyyy).
	f := NewDateFormat()
	tm := time.Date(2024, 6, 15, 0, 0, 0, 0, time.UTC)
	if got := f.FormatValue(tm); got != "6/15/2024" {
		t.Fatalf("got %q, want %q", got, "6/15/2024")
	}
}

func TestDateFormat_PointerToTime(t *testing.T) {
	f := NewDateFormat()
	tm := time.Date(2024, 6, 15, 0, 0, 0, 0, time.UTC)
	if got := f.FormatValue(&tm); got != "6/15/2024" {
		t.Fatalf("got %q, want %q", got, "6/15/2024")
	}
}

func TestDateFormat_StringISO(t *testing.T) {
	f := NewDateFormat()
	if got := f.FormatValue("2024-06-15"); got != "6/15/2024" {
		t.Fatalf("got %q, want %q", got, "6/15/2024")
	}
}

func TestDateFormat_StringRFC3339(t *testing.T) {
	f := NewDateFormat()
	if got := f.FormatValue("2024-06-15T00:00:00Z"); got != "6/15/2024" {
		t.Fatalf("got %q, want %q", got, "6/15/2024")
	}
}

func TestDateFormat_StringSlash(t *testing.T) {
	f := NewDateFormat()
	if got := f.FormatValue("06/15/2024"); got != "6/15/2024" {
		t.Fatalf("got %q, want %q", got, "6/15/2024")
	}
}

func TestDateFormat_StringDash(t *testing.T) {
	f := NewDateFormat()
	if got := f.FormatValue("15-06-2024"); got != "6/15/2024" {
		t.Fatalf("got %q, want %q", got, "6/15/2024")
	}
}

func TestDateFormat_StringDateTime(t *testing.T) {
	f := NewDateFormat()
	if got := f.FormatValue("2024-06-15T10:30:00"); got != "6/15/2024" {
		t.Fatalf("got %q, want %q", got, "6/15/2024")
	}
}

func TestDateFormat_StringSpaceDateTime(t *testing.T) {
	f := NewDateFormat()
	if got := f.FormatValue("2024-06-15 10:30:00"); got != "6/15/2024" {
		t.Fatalf("got %q, want %q", got, "6/15/2024")
	}
}

func TestDateFormat_UnparsableString(t *testing.T) {
	f := NewDateFormat()
	if got := f.FormatValue("not-a-date"); got != "not-a-date" {
		t.Fatalf("got %q, want %q", got, "not-a-date")
	}
}

func TestDateFormat_EmptyFormat(t *testing.T) {
	// Empty format falls back to "d" (C# short date = M/d/yyyy).
	f := &DateFormat{Format: ""}
	tm := time.Date(2024, 6, 15, 0, 0, 0, 0, time.UTC)
	if got := f.FormatValue(tm); got != "6/15/2024" {
		t.Fatalf("got %q, want %q", got, "6/15/2024")
	}
}

func TestDateFormat_CsharpShortDate(t *testing.T) {
	// "d" is C# short date pattern → M/d/yyyy.
	f := &DateFormat{Format: "d"}
	tm := time.Date(2013, 8, 12, 0, 0, 0, 0, time.UTC)
	if got := f.FormatValue(tm); got != "8/12/2013" {
		t.Fatalf("got %q, want %q", got, "8/12/2013")
	}
}

func TestDateFormat_CustomLayout(t *testing.T) {
	f := &DateFormat{Format: "01/02/2006"}
	tm := time.Date(2024, 6, 15, 0, 0, 0, 0, time.UTC)
	if got := f.FormatValue(tm); got != "06/15/2024" {
		t.Fatalf("got %q, want %q", got, "06/15/2024")
	}
}

func TestDateFormat_NilPointer(t *testing.T) {
	f := NewDateFormat()
	var tp *time.Time
	// A typed nil pointer: toTime should return false.
	if got := f.FormatValue(tp); got != "" {
		t.Fatalf("got %q, want %q", got, "")
	}
}

func TestDateFormat_ImplementsFormat(t *testing.T) {
	var _ Format = NewDateFormat()
}

// ---------------------------------------------------------------------------
// TimeFormat
// ---------------------------------------------------------------------------

func TestTimeFormat_FormatType(t *testing.T) {
	f := NewTimeFormat()
	if got := f.FormatType(); got != "Time" {
		t.Fatalf("FormatType = %q, want %q", got, "Time")
	}
}

func TestTimeFormat_Nil(t *testing.T) {
	f := NewTimeFormat()
	if got := f.FormatValue(nil); got != "" {
		t.Fatalf("got %q, want %q", got, "")
	}
}

func TestTimeFormat_TimeTime(t *testing.T) {
	f := NewTimeFormat()
	tm := time.Date(2024, 1, 1, 13, 30, 0, 0, time.UTC)
	if got := f.FormatValue(tm); got != "13:30" {
		t.Fatalf("got %q, want %q", got, "13:30")
	}
}

func TestTimeFormat_Duration(t *testing.T) {
	f := &TimeFormat{Format: "15:04:05"}
	d := 2*time.Hour + 30*time.Minute + 15*time.Second
	if got := f.FormatValue(d); got != "02:30:15" {
		t.Fatalf("got %q, want %q", got, "02:30:15")
	}
}

func TestTimeFormat_StringRFC3339(t *testing.T) {
	f := NewTimeFormat()
	if got := f.FormatValue("2024-06-15T13:30:00Z"); got != "13:30" {
		t.Fatalf("got %q, want %q", got, "13:30")
	}
}

func TestTimeFormat_UnparsableString(t *testing.T) {
	f := NewTimeFormat()
	if got := f.FormatValue("not-a-time"); got != "not-a-time" {
		t.Fatalf("got %q, want %q", got, "not-a-time")
	}
}

func TestTimeFormat_EmptyFormat(t *testing.T) {
	f := &TimeFormat{Format: ""}
	tm := time.Date(2024, 1, 1, 9, 5, 0, 0, time.UTC)
	if got := f.FormatValue(tm); got != "09:05" {
		t.Fatalf("got %q, want %q", got, "09:05")
	}
}

func TestTimeFormat_CustomLayout(t *testing.T) {
	f := &TimeFormat{Format: "3:04 PM"}
	tm := time.Date(2024, 1, 1, 14, 5, 0, 0, time.UTC)
	if got := f.FormatValue(tm); got != "2:05 PM" {
		t.Fatalf("got %q, want %q", got, "2:05 PM")
	}
}

func TestTimeFormat_ImplementsFormat(t *testing.T) {
	var _ Format = NewTimeFormat()
}

// ---------------------------------------------------------------------------
// insertGroupSeparator helper
// ---------------------------------------------------------------------------

func TestInsertGroupSeparator_Short(t *testing.T) {
	if got := insertGroupSeparator("123", ","); got != "123" {
		t.Fatalf("got %q, want %q", got, "123")
	}
}

func TestInsertGroupSeparator_Exact3(t *testing.T) {
	if got := insertGroupSeparator("100", ","); got != "100" {
		t.Fatalf("got %q, want %q", got, "100")
	}
}

func TestInsertGroupSeparator_4Digits(t *testing.T) {
	if got := insertGroupSeparator("1234", ","); got != "1,234" {
		t.Fatalf("got %q, want %q", got, "1,234")
	}
}

func TestInsertGroupSeparator_7Digits(t *testing.T) {
	if got := insertGroupSeparator("1234567", ","); got != "1,234,567" {
		t.Fatalf("got %q, want %q", got, "1,234,567")
	}
}

func TestInsertGroupSeparator_Divisible3(t *testing.T) {
	if got := insertGroupSeparator("123456", ","); got != "123,456" {
		t.Fatalf("got %q, want %q", got, "123,456")
	}
}

// ---------------------------------------------------------------------------
// isNilPointer helper
// ---------------------------------------------------------------------------

func TestIsNilPointer_UntypedNil(t *testing.T) {
	if !isNilPointer(nil) {
		t.Fatal("expected true for untyped nil")
	}
}

func TestIsNilPointer_TypedNil(t *testing.T) {
	var p *int
	if !isNilPointer(p) {
		t.Fatal("expected true for typed nil pointer")
	}
}

func TestIsNilPointer_NonNil(t *testing.T) {
	x := 42
	if isNilPointer(&x) {
		t.Fatal("expected false for non-nil pointer")
	}
}

func TestIsNilPointer_NonPointer(t *testing.T) {
	if isNilPointer(42) {
		t.Fatal("expected false for non-pointer value")
	}
}
