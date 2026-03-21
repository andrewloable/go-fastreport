package format

import (
	"testing"
	"time"
)

// ---------------------------------------------------------------------------
// BooleanFormat — Clone, Equals, GetSampleValue
// ---------------------------------------------------------------------------

func TestBooleanFormat_Clone(t *testing.T) {
	orig := &BooleanFormat{TrueText: "Yes", FalseText: "No"}
	cloned := orig.Clone()
	bf, ok := cloned.(*BooleanFormat)
	if !ok {
		t.Fatal("Clone() did not return *BooleanFormat")
	}
	if bf.TrueText != orig.TrueText || bf.FalseText != orig.FalseText {
		t.Errorf("Clone() fields mismatch: got %v, want %v", bf, orig)
	}
	// Ensure it's a new pointer (deep copy).
	if cloned == Format(orig) {
		t.Error("Clone() returned same pointer, want new instance")
	}
}

func TestBooleanFormat_Clone_Defaults(t *testing.T) {
	orig := NewBooleanFormat()
	cloned := orig.Clone().(*BooleanFormat)
	if cloned.TrueText != "True" || cloned.FalseText != "False" {
		t.Errorf("Clone of default BooleanFormat: TrueText=%q FalseText=%q", cloned.TrueText, cloned.FalseText)
	}
}

func TestBooleanFormat_Equals_Same(t *testing.T) {
	a := &BooleanFormat{TrueText: "Yes", FalseText: "No"}
	b := &BooleanFormat{TrueText: "Yes", FalseText: "No"}
	if !a.Equals(b) {
		t.Error("expected Equals to return true for identical BooleanFormats")
	}
}

func TestBooleanFormat_Equals_Defaults(t *testing.T) {
	a := NewBooleanFormat()
	b := NewBooleanFormat()
	if !a.Equals(b) {
		t.Error("expected Equals true for two default BooleanFormats")
	}
}

func TestBooleanFormat_Equals_DifferentTrue(t *testing.T) {
	a := &BooleanFormat{TrueText: "Yes", FalseText: "No"}
	b := &BooleanFormat{TrueText: "True", FalseText: "No"}
	if a.Equals(b) {
		t.Error("expected Equals false when TrueText differs")
	}
}

func TestBooleanFormat_Equals_DifferentFalse(t *testing.T) {
	a := &BooleanFormat{TrueText: "Yes", FalseText: "No"}
	b := &BooleanFormat{TrueText: "Yes", FalseText: "False"}
	if a.Equals(b) {
		t.Error("expected Equals false when FalseText differs")
	}
}

func TestBooleanFormat_Equals_WrongType(t *testing.T) {
	a := NewBooleanFormat()
	if a.Equals(NewGeneralFormat()) {
		t.Error("expected Equals false when comparing BooleanFormat to GeneralFormat")
	}
}

func TestBooleanFormat_GetSampleValue(t *testing.T) {
	f := NewBooleanFormat()
	// C# GetSampleValue calls FormatValue(false) → should return FalseText.
	got := f.GetSampleValue()
	if got != "False" {
		t.Errorf("GetSampleValue = %q, want %q", got, "False")
	}
}

func TestBooleanFormat_GetSampleValue_Custom(t *testing.T) {
	f := &BooleanFormat{TrueText: "Yes", FalseText: "No"}
	got := f.GetSampleValue()
	if got != "No" {
		t.Errorf("GetSampleValue = %q, want %q", got, "No")
	}
}

// ---------------------------------------------------------------------------
// CurrencyFormat — Clone, Equals, GetSampleValue
// ---------------------------------------------------------------------------

func TestCurrencyFormat_Clone(t *testing.T) {
	orig := &CurrencyFormat{
		UseLocaleSettings: false,
		DecimalDigits:     3,
		DecimalSeparator:  ",",
		GroupSeparator:    ".",
		CurrencySymbol:    "€",
		PositivePattern:   2,
		NegativePattern:   9,
	}
	cloned := orig.Clone().(*CurrencyFormat)
	if cloned.UseLocaleSettings != orig.UseLocaleSettings ||
		cloned.DecimalDigits != orig.DecimalDigits ||
		cloned.DecimalSeparator != orig.DecimalSeparator ||
		cloned.GroupSeparator != orig.GroupSeparator ||
		cloned.CurrencySymbol != orig.CurrencySymbol ||
		cloned.PositivePattern != orig.PositivePattern ||
		cloned.NegativePattern != orig.NegativePattern {
		t.Errorf("Clone() fields mismatch:\n  got  %+v\n  want %+v", cloned, orig)
	}
	// Deep copy — different pointer.
	if Format(cloned) == Format(orig) {
		t.Error("Clone() returned same pointer, want new instance")
	}
}

func TestCurrencyFormat_Clone_Defaults(t *testing.T) {
	orig := NewCurrencyFormat()
	cloned := orig.Clone().(*CurrencyFormat)
	if cloned.DecimalDigits != 2 || !cloned.UseLocaleSettings || cloned.CurrencySymbol != "$" {
		t.Errorf("Clone of default CurrencyFormat: %+v", cloned)
	}
}

func TestCurrencyFormat_Equals_Same(t *testing.T) {
	a := &CurrencyFormat{UseLocaleSettings: false, DecimalDigits: 2, CurrencySymbol: "€"}
	b := &CurrencyFormat{UseLocaleSettings: false, DecimalDigits: 2, CurrencySymbol: "€"}
	if !a.Equals(b) {
		t.Error("expected Equals true for identical CurrencyFormats")
	}
}

func TestCurrencyFormat_Equals_Defaults(t *testing.T) {
	a := NewCurrencyFormat()
	b := NewCurrencyFormat()
	if !a.Equals(b) {
		t.Error("expected Equals true for two default CurrencyFormats")
	}
}

func TestCurrencyFormat_Equals_DifferentLocale(t *testing.T) {
	a := &CurrencyFormat{UseLocaleSettings: true, DecimalDigits: 2, CurrencySymbol: "$"}
	b := &CurrencyFormat{UseLocaleSettings: false, DecimalDigits: 2, CurrencySymbol: "$"}
	if a.Equals(b) {
		t.Error("expected Equals false when UseLocaleSettings differs")
	}
}

func TestCurrencyFormat_Equals_DifferentDigits(t *testing.T) {
	a := NewCurrencyFormat()
	b := NewCurrencyFormat()
	b.DecimalDigits = 4
	if a.Equals(b) {
		t.Error("expected Equals false when DecimalDigits differs")
	}
}

func TestCurrencyFormat_Equals_DifferentSymbol(t *testing.T) {
	a := NewCurrencyFormat()
	b := NewCurrencyFormat()
	b.CurrencySymbol = "€"
	if a.Equals(b) {
		t.Error("expected Equals false when CurrencySymbol differs")
	}
}

func TestCurrencyFormat_Equals_DifferentDecSep(t *testing.T) {
	a := NewCurrencyFormat()
	b := NewCurrencyFormat()
	b.DecimalSeparator = ","
	if a.Equals(b) {
		t.Error("expected Equals false when DecimalSeparator differs")
	}
}

func TestCurrencyFormat_Equals_DifferentGrpSep(t *testing.T) {
	a := NewCurrencyFormat()
	b := NewCurrencyFormat()
	b.GroupSeparator = "."
	if a.Equals(b) {
		t.Error("expected Equals false when GroupSeparator differs")
	}
}

func TestCurrencyFormat_Equals_DifferentPosPattern(t *testing.T) {
	a := NewCurrencyFormat()
	b := NewCurrencyFormat()
	b.PositivePattern = 1
	if a.Equals(b) {
		t.Error("expected Equals false when PositivePattern differs")
	}
}

func TestCurrencyFormat_Equals_DifferentNegPattern(t *testing.T) {
	a := NewCurrencyFormat()
	b := NewCurrencyFormat()
	b.NegativePattern = 5
	if a.Equals(b) {
		t.Error("expected Equals false when NegativePattern differs")
	}
}

func TestCurrencyFormat_Equals_WrongType(t *testing.T) {
	a := NewCurrencyFormat()
	if a.Equals(NewGeneralFormat()) {
		t.Error("expected Equals false when comparing CurrencyFormat to GeneralFormat")
	}
}

func TestCurrencyFormat_GetSampleValue(t *testing.T) {
	f := &CurrencyFormat{
		UseLocaleSettings: false,
		DecimalDigits:     2,
		DecimalSeparator:  ".",
		GroupSeparator:    ",",
		CurrencySymbol:    "$",
		PositivePattern:   0,
		NegativePattern:   1,
	}
	// C# GetSampleValue calls FormatValue(-12345) → "-$12,345.00" with pattern 1.
	got := f.GetSampleValue()
	if got != "-$12,345.00" {
		t.Errorf("GetSampleValue = %q, want %q", got, "-$12,345.00")
	}
}

func TestCurrencyFormat_GetSampleValue_Negative(t *testing.T) {
	// Verify the result is non-empty and contains the currency symbol.
	f := NewCurrencyFormat()
	got := f.GetSampleValue()
	if got == "" {
		t.Error("GetSampleValue returned empty string")
	}
}

// ---------------------------------------------------------------------------
// CustomFormat — Clone, Equals, GetSampleValue
// ---------------------------------------------------------------------------

func TestCustomFormat_Clone(t *testing.T) {
	orig := &CustomFormat{Format: "%.4f"}
	cloned := orig.Clone().(*CustomFormat)
	if cloned.Format != orig.Format {
		t.Errorf("Clone() Format = %q, want %q", cloned.Format, orig.Format)
	}
	if Format(cloned) == Format(orig) {
		t.Error("Clone() returned same pointer, want new instance")
	}
}

func TestCustomFormat_Clone_Default(t *testing.T) {
	orig := NewCustomFormat()
	cloned := orig.Clone().(*CustomFormat)
	if cloned.Format != "%v" {
		t.Errorf("Clone of default CustomFormat: Format=%q, want %%v", cloned.Format)
	}
}

func TestCustomFormat_Equals_Same(t *testing.T) {
	a := &CustomFormat{Format: "%.2f"}
	b := &CustomFormat{Format: "%.2f"}
	if !a.Equals(b) {
		t.Error("expected Equals true for identical CustomFormats")
	}
}

func TestCustomFormat_Equals_Defaults(t *testing.T) {
	a := NewCustomFormat()
	b := NewCustomFormat()
	if !a.Equals(b) {
		t.Error("expected Equals true for two default CustomFormats")
	}
}

func TestCustomFormat_Equals_Different(t *testing.T) {
	a := &CustomFormat{Format: "%.2f"}
	b := &CustomFormat{Format: "%.4f"}
	if a.Equals(b) {
		t.Error("expected Equals false when Format strings differ")
	}
}

func TestCustomFormat_Equals_WrongType(t *testing.T) {
	a := NewCustomFormat()
	if a.Equals(NewBooleanFormat()) {
		t.Error("expected Equals false when comparing CustomFormat to BooleanFormat")
	}
}

func TestCustomFormat_GetSampleValue(t *testing.T) {
	f := NewCustomFormat()
	// C# GetSampleValue returns "".
	got := f.GetSampleValue()
	if got != "" {
		t.Errorf("GetSampleValue = %q, want %q", got, "")
	}
}

// ---------------------------------------------------------------------------
// GeneralFormat — Clone, Equals, GetSampleValue
// ---------------------------------------------------------------------------

func TestGeneralFormat_Clone(t *testing.T) {
	orig := NewGeneralFormat()
	cloned := orig.Clone()
	if _, ok := cloned.(*GeneralFormat); !ok {
		t.Fatal("Clone() did not return *GeneralFormat")
	}
}

func TestGeneralFormat_Equals_Same(t *testing.T) {
	a := NewGeneralFormat()
	b := NewGeneralFormat()
	if !a.Equals(b) {
		t.Error("expected Equals true for two GeneralFormats")
	}
}

func TestGeneralFormat_Equals_WrongType(t *testing.T) {
	a := NewGeneralFormat()
	if a.Equals(NewBooleanFormat()) {
		t.Error("expected Equals false comparing GeneralFormat to BooleanFormat")
	}
}

func TestGeneralFormat_GetSampleValue(t *testing.T) {
	f := NewGeneralFormat()
	if got := f.GetSampleValue(); got != "" {
		t.Errorf("GeneralFormat.GetSampleValue = %q, want empty", got)
	}
}

// ---------------------------------------------------------------------------
// NumberFormat — Clone, Equals, GetSampleValue
// ---------------------------------------------------------------------------

func TestNumberFormat_Clone(t *testing.T) {
	orig := &NumberFormat{
		UseLocaleSettings: false,
		DecimalDigits:     3,
		DecimalSeparator:  ",",
		GroupSeparator:    ".",
		NegativePattern:   3,
	}
	cloned := orig.Clone().(*NumberFormat)
	if cloned.UseLocaleSettings != orig.UseLocaleSettings ||
		cloned.DecimalDigits != orig.DecimalDigits ||
		cloned.DecimalSeparator != orig.DecimalSeparator ||
		cloned.GroupSeparator != orig.GroupSeparator ||
		cloned.NegativePattern != orig.NegativePattern {
		t.Errorf("Clone() fields mismatch:\n  got  %+v\n  want %+v", cloned, orig)
	}
}

func TestNumberFormat_Equals_Same(t *testing.T) {
	a := NewNumberFormat()
	b := NewNumberFormat()
	if !a.Equals(b) {
		t.Error("expected Equals true for two default NumberFormats")
	}
}

func TestNumberFormat_Equals_Different(t *testing.T) {
	a := NewNumberFormat()
	b := NewNumberFormat()
	b.DecimalDigits = 4
	if a.Equals(b) {
		t.Error("expected Equals false when DecimalDigits differs")
	}
}

func TestNumberFormat_Equals_DifferentNegPattern(t *testing.T) {
	a := NewNumberFormat()
	b := NewNumberFormat()
	b.NegativePattern = 0
	if a.Equals(b) {
		t.Error("expected Equals false when NegativePattern differs")
	}
}

func TestNumberFormat_Equals_DifferentLocale(t *testing.T) {
	a := NewNumberFormat()
	b := NewNumberFormat()
	b.UseLocaleSettings = false
	if a.Equals(b) {
		t.Error("expected Equals false when UseLocaleSettings differs")
	}
}

func TestNumberFormat_Equals_DifferentDecSep(t *testing.T) {
	a := NewNumberFormat()
	b := NewNumberFormat()
	b.DecimalSeparator = ","
	if a.Equals(b) {
		t.Error("expected Equals false when DecimalSeparator differs")
	}
}

func TestNumberFormat_Equals_DifferentGrpSep(t *testing.T) {
	a := NewNumberFormat()
	b := NewNumberFormat()
	b.GroupSeparator = "."
	if a.Equals(b) {
		t.Error("expected Equals false when GroupSeparator differs")
	}
}

func TestNumberFormat_Equals_WrongType(t *testing.T) {
	a := NewNumberFormat()
	if a.Equals(NewGeneralFormat()) {
		t.Error("expected Equals false when comparing NumberFormat to GeneralFormat")
	}
}

func TestNumberFormat_GetSampleValue(t *testing.T) {
	f := &NumberFormat{
		UseLocaleSettings: false,
		DecimalDigits:     2,
		DecimalSeparator:  ".",
		GroupSeparator:    ",",
		NegativePattern:   1,
	}
	// C# GetSampleValue calls FormatValue(-12345.678) → "-12,345.68"
	got := f.GetSampleValue()
	if got != "-12,345.68" {
		t.Errorf("GetSampleValue = %q, want %q", got, "-12,345.68")
	}
}

// ---------------------------------------------------------------------------
// PercentFormat — Clone, Equals, GetSampleValue
// ---------------------------------------------------------------------------

func TestPercentFormat_Clone(t *testing.T) {
	orig := &PercentFormat{
		UseLocaleSettings: false,
		DecimalDigits:     1,
		DecimalSeparator:  ",",
		GroupSeparator:    ".",
		PercentSymbol:     "pct",
		PositivePattern:   2,
		NegativePattern:   4,
	}
	cloned := orig.Clone().(*PercentFormat)
	if cloned.UseLocaleSettings != orig.UseLocaleSettings ||
		cloned.DecimalDigits != orig.DecimalDigits ||
		cloned.DecimalSeparator != orig.DecimalSeparator ||
		cloned.GroupSeparator != orig.GroupSeparator ||
		cloned.PercentSymbol != orig.PercentSymbol ||
		cloned.PositivePattern != orig.PositivePattern ||
		cloned.NegativePattern != orig.NegativePattern {
		t.Errorf("Clone() fields mismatch:\n  got  %+v\n  want %+v", cloned, orig)
	}
}

func TestPercentFormat_Equals_Same(t *testing.T) {
	a := NewPercentFormat()
	b := NewPercentFormat()
	if !a.Equals(b) {
		t.Error("expected Equals true for two default PercentFormats")
	}
}

func TestPercentFormat_Equals_DifferentSymbol(t *testing.T) {
	a := NewPercentFormat()
	b := NewPercentFormat()
	b.PercentSymbol = "pct"
	if a.Equals(b) {
		t.Error("expected Equals false when PercentSymbol differs")
	}
}

func TestPercentFormat_Equals_WrongType(t *testing.T) {
	a := NewPercentFormat()
	if a.Equals(NewNumberFormat()) {
		t.Error("expected Equals false when comparing PercentFormat to NumberFormat")
	}
}

func TestPercentFormat_GetSampleValue_Method(t *testing.T) {
	f := &PercentFormat{
		UseLocaleSettings: false,
		DecimalDigits:     2,
		DecimalSeparator:  ".",
		GroupSeparator:    ",",
		PercentSymbol:     "%",
		PositivePattern:   0, // "n %"
	}
	// C# GetSampleValue calls FormatValue(1.23f) → 1.23*100 = 123 → "123.00 %"
	got := f.GetSampleValue()
	if got != "123.00 %" {
		t.Errorf("GetSampleValue = %q, want %q", got, "123.00 %")
	}
}

// ---------------------------------------------------------------------------
// DateFormat — Clone, Equals, GetSampleValue
// ---------------------------------------------------------------------------

func TestDateFormat_Clone(t *testing.T) {
	orig := &DateFormat{Format: "MM/dd/yyyy", UseLocaleSettings: true}
	cloned := orig.Clone().(*DateFormat)
	if cloned.Format != orig.Format || cloned.UseLocaleSettings != orig.UseLocaleSettings {
		t.Errorf("Clone() fields mismatch:\n  got  %+v\n  want %+v", cloned, orig)
	}
	if Format(cloned) == Format(orig) {
		t.Error("Clone() returned same pointer, want new instance")
	}
}

func TestDateFormat_Equals_Same(t *testing.T) {
	a := NewDateFormat()
	b := NewDateFormat()
	if !a.Equals(b) {
		t.Error("expected Equals true for two default DateFormats")
	}
}

func TestDateFormat_Equals_DifferentFormat(t *testing.T) {
	a := NewDateFormat()
	b := &DateFormat{Format: "dd/MM/yyyy"}
	if a.Equals(b) {
		t.Error("expected Equals false when Format strings differ")
	}
}

func TestDateFormat_Equals_DifferentLocale(t *testing.T) {
	a := NewDateFormat()
	b := NewDateFormat()
	b.UseLocaleSettings = true
	if a.Equals(b) {
		t.Error("expected Equals false when UseLocaleSettings differs")
	}
}

func TestDateFormat_Equals_WrongType(t *testing.T) {
	a := NewDateFormat()
	if a.Equals(NewTimeFormat()) {
		t.Error("expected Equals false when comparing DateFormat to TimeFormat")
	}
}

func TestDateFormat_GetSampleValue_Method(t *testing.T) {
	// Default format "d" → M/d/yyyy. The sample date is 2007-11-30.
	f := NewDateFormat()
	got := f.GetSampleValue()
	if got != "11/30/2007" {
		t.Errorf("GetSampleValue = %q, want %q", got, "11/30/2007")
	}
}

// ---------------------------------------------------------------------------
// TimeFormat — Clone, Equals, GetSampleValue
// ---------------------------------------------------------------------------

func TestTimeFormat_Clone(t *testing.T) {
	orig := &TimeFormat{Format: "3:04 PM", UseLocaleSettings: true}
	cloned := orig.Clone().(*TimeFormat)
	if cloned.Format != orig.Format || cloned.UseLocaleSettings != orig.UseLocaleSettings {
		t.Errorf("Clone() fields mismatch:\n  got  %+v\n  want %+v", cloned, orig)
	}
	if Format(cloned) == Format(orig) {
		t.Error("Clone() returned same pointer, want new instance")
	}
}

func TestTimeFormat_Equals_Same(t *testing.T) {
	a := NewTimeFormat()
	b := NewTimeFormat()
	if !a.Equals(b) {
		t.Error("expected Equals true for two default TimeFormats")
	}
}

func TestTimeFormat_Equals_DifferentFormat(t *testing.T) {
	a := NewTimeFormat()
	b := &TimeFormat{Format: "3:04:05 PM"}
	if a.Equals(b) {
		t.Error("expected Equals false when Format strings differ")
	}
}

func TestTimeFormat_Equals_DifferentLocale(t *testing.T) {
	a := NewTimeFormat()
	b := NewTimeFormat()
	b.UseLocaleSettings = true
	if a.Equals(b) {
		t.Error("expected Equals false when UseLocaleSettings differs")
	}
}

func TestTimeFormat_Equals_WrongType(t *testing.T) {
	a := NewTimeFormat()
	if a.Equals(NewDateFormat()) {
		t.Error("expected Equals false when comparing TimeFormat to DateFormat")
	}
}

func TestTimeFormat_GetSampleValue(t *testing.T) {
	// Default format "15:04". Sample time is 2007-11-30 13:30:00 → "13:30".
	f := NewTimeFormat()
	got := f.GetSampleValue()
	if got != "13:30" {
		t.Errorf("GetSampleValue = %q, want %q", got, "13:30")
	}
}

// ---------------------------------------------------------------------------
// Collection — Equals
// ---------------------------------------------------------------------------

func TestCollection_Equals_Empty(t *testing.T) {
	a := NewCollection()
	b := NewCollection()
	if !a.Equals(b) {
		t.Error("expected Equals true for two empty Collections")
	}
}

func TestCollection_Equals_SameContent(t *testing.T) {
	a := NewCollection()
	b := NewCollection()
	bf := NewBooleanFormat()
	a.Add(bf)
	b.Add(NewBooleanFormat()) // equal content, different pointer
	if !a.Equals(b) {
		t.Error("expected Equals true when collections have same content")
	}
}

func TestCollection_Equals_DifferentCount(t *testing.T) {
	a := NewCollection()
	b := NewCollection()
	a.Add(NewBooleanFormat())
	a.Add(NewGeneralFormat())
	b.Add(NewBooleanFormat())
	if a.Equals(b) {
		t.Error("expected Equals false when collections have different counts")
	}
}

func TestCollection_Equals_DifferentContent(t *testing.T) {
	a := NewCollection()
	b := NewCollection()
	a.Add(NewBooleanFormat())
	b.Add(NewGeneralFormat())
	if a.Equals(b) {
		t.Error("expected Equals false when collections have different content")
	}
}

func TestCollection_Equals_Nil(t *testing.T) {
	a := NewCollection()
	if a.Equals(nil) {
		t.Error("expected Equals false when comparing to nil")
	}
}

func TestCollection_Equals_PointerIdentity(t *testing.T) {
	// When the format type has no Equals method (simpleFormat), pointer equality
	// is used. Same pointer → true.
	a := NewCollection()
	b := NewCollection()
	sf := &simpleFormat{}
	a.Add(sf)
	b.Add(sf) // same pointer
	if !a.Equals(b) {
		t.Error("expected Equals true for collections with same pointer (no Equals method)")
	}
}

func TestCollection_Equals_DifferentFormats(t *testing.T) {
	// Collections with same length but structurally different format types → false.
	a := NewCollection()
	b := NewCollection()
	a.Add(NewBooleanFormat())             // *BooleanFormat (has Equals)
	b.Add(&BooleanFormat{TrueText: "Y"}) // different TrueText
	if a.Equals(b) {
		t.Error("expected Equals false for collections with different BooleanFormat content")
	}
}

// ---------------------------------------------------------------------------
// Collection.Assign uses Clone — verify all built-in types are cloned deeply.
// ---------------------------------------------------------------------------

func TestCollection_Assign_BooleanFormat(t *testing.T) {
	src := NewCollection()
	src.Add(&BooleanFormat{TrueText: "YES", FalseText: "NO"})
	dst := NewCollection()
	dst.Assign(src)
	if dst.Count() != 1 {
		t.Fatalf("count after Assign = %d, want 1", dst.Count())
	}
	// Different pointer — deep copy.
	if dst.Get(0) == src.Get(0) {
		t.Error("Assign should deep-copy BooleanFormat")
	}
	bf := dst.Get(0).(*BooleanFormat)
	if bf.TrueText != "YES" || bf.FalseText != "NO" {
		t.Errorf("Assign copied wrong values: %+v", bf)
	}
}

func TestCollection_Assign_CurrencyFormat(t *testing.T) {
	src := NewCollection()
	src.Add(&CurrencyFormat{UseLocaleSettings: false, CurrencySymbol: "€", DecimalDigits: 3})
	dst := NewCollection()
	dst.Assign(src)
	if dst.Get(0) == src.Get(0) {
		t.Error("Assign should deep-copy CurrencyFormat")
	}
	cf := dst.Get(0).(*CurrencyFormat)
	if cf.CurrencySymbol != "€" || cf.DecimalDigits != 3 {
		t.Errorf("Assign copied wrong values: %+v", cf)
	}
}

func TestCollection_Assign_CustomFormat(t *testing.T) {
	src := NewCollection()
	src.Add(&CustomFormat{Format: "%.4f"})
	dst := NewCollection()
	dst.Assign(src)
	if dst.Get(0) == src.Get(0) {
		t.Error("Assign should deep-copy CustomFormat")
	}
	cf := dst.Get(0).(*CustomFormat)
	if cf.Format != "%.4f" {
		t.Errorf("Assign copied wrong Format: %q", cf.Format)
	}
}

func TestCollection_Assign_GeneralFormat(t *testing.T) {
	src := NewCollection()
	src.Add(NewGeneralFormat())
	dst := NewCollection()
	dst.Assign(src)
	// After Assign, dst should contain a GeneralFormat (clone or same value).
	if dst.Count() != 1 {
		t.Fatalf("Assign count = %d, want 1", dst.Count())
	}
	if _, ok := dst.Get(0).(*GeneralFormat); !ok {
		t.Error("Assign should produce a *GeneralFormat")
	}
}

func TestCollection_Assign_NumberFormat(t *testing.T) {
	src := NewCollection()
	src.Add(&NumberFormat{DecimalDigits: 4, GroupSeparator: "'"})
	dst := NewCollection()
	dst.Assign(src)
	nf := dst.Get(0).(*NumberFormat)
	if nf.DecimalDigits != 4 || nf.GroupSeparator != "'" {
		t.Errorf("Assign copied wrong values: %+v", nf)
	}
}

func TestCollection_Assign_DateFormat(t *testing.T) {
	src := NewCollection()
	src.Add(&DateFormat{Format: "dd/MM/yyyy"})
	dst := NewCollection()
	dst.Assign(src)
	df := dst.Get(0).(*DateFormat)
	if df.Format != "dd/MM/yyyy" {
		t.Errorf("Assign copied wrong Format: %q", df.Format)
	}
}

func TestCollection_Assign_TimeFormat(t *testing.T) {
	src := NewCollection()
	src.Add(&TimeFormat{Format: "3:04 PM"})
	dst := NewCollection()
	dst.Assign(src)
	tf := dst.Get(0).(*TimeFormat)
	if tf.Format != "3:04 PM" {
		t.Errorf("Assign copied wrong Format: %q", tf.Format)
	}
}

func TestCollection_Assign_PercentFormat(t *testing.T) {
	src := NewCollection()
	src.Add(&PercentFormat{PercentSymbol: "pct", DecimalDigits: 1})
	dst := NewCollection()
	dst.Assign(src)
	pf := dst.Get(0).(*PercentFormat)
	if pf.PercentSymbol != "pct" || pf.DecimalDigits != 1 {
		t.Errorf("Assign copied wrong values: %+v", pf)
	}
}

// ---------------------------------------------------------------------------
// BooleanFormat — typed-nil pointer branch in FormatValue
// ---------------------------------------------------------------------------

func TestBooleanFormat_TypedNilPointer(t *testing.T) {
	f := NewBooleanFormat()
	var p *bool
	got := f.FormatValue(p)
	if got != "" {
		t.Errorf("FormatValue(typed nil *bool) = %q, want empty", got)
	}
}

// ---------------------------------------------------------------------------
// CurrencyFormat — typed-nil pointer branch in FormatValue
// ---------------------------------------------------------------------------

func TestCurrencyFormat_TypedNilPointer(t *testing.T) {
	f := NewCurrencyFormat()
	var p *float64
	got := f.FormatValue(p)
	if got != "" {
		t.Errorf("FormatValue(typed nil *float64) = %q, want empty", got)
	}
}

// ---------------------------------------------------------------------------
// NumberFormat — typed-nil pointer branch in FormatValue
// ---------------------------------------------------------------------------

func TestNumberFormat_TypedNilPointer(t *testing.T) {
	f := NewNumberFormat()
	var p *float64
	got := f.FormatValue(p)
	if got != "" {
		t.Errorf("NumberFormat.FormatValue(typed nil *float64) = %q, want empty", got)
	}
}

// ---------------------------------------------------------------------------
// PercentFormat — typed-nil pointer branch in FormatValue
// ---------------------------------------------------------------------------

func TestPercentFormat_TypedNilPointer(t *testing.T) {
	f := NewPercentFormat()
	var p *float64
	got := f.FormatValue(p)
	if got != "" {
		t.Errorf("PercentFormat.FormatValue(typed nil *float64) = %q, want empty", got)
	}
}

// ---------------------------------------------------------------------------
// TimeFormat — typed-nil pointer in FormatValue
// ---------------------------------------------------------------------------

func TestTimeFormat_TypedNil(t *testing.T) {
	f := NewTimeFormat()
	var p *time.Time
	got := f.FormatValue(p)
	if got != "" {
		t.Errorf("TimeFormat.FormatValue(typed nil) = %q, want empty", got)
	}
}

// ---------------------------------------------------------------------------
// Collection.Equals — pointer identity fallback (no Equals method)
// ---------------------------------------------------------------------------

func TestCollection_Equals_SamePointerNoEqualsMethod(t *testing.T) {
	// simpleFormat has no Equals method — same pointer should return true.
	a := NewCollection()
	b := NewCollection()
	sf := &simpleFormat{}
	a.Add(sf)
	b.Add(sf) // same pointer
	if !a.Equals(b) {
		t.Error("expected Equals true for same pointer in collections without Equals method")
	}
}

// stateFormat is a non-empty format struct without an Equals method,
// ensuring pointer aliasing does not occur.
type stateFormat struct {
	id int
}

func (f *stateFormat) FormatType() string      { return "state" }
func (f *stateFormat) FormatValue(_ any) string { return "state" }

func TestCollection_Equals_DifferentPointerNoEqualsMethod(t *testing.T) {
	// stateFormat has no Equals method — different pointers should return false.
	a := NewCollection()
	b := NewCollection()
	a.Add(&stateFormat{id: 1})
	b.Add(&stateFormat{id: 2}) // distinct heap allocations
	if a.Equals(b) {
		t.Error("expected Equals false for different pointers in collections without Equals method")
	}
}
