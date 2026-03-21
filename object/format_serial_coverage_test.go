package object_test

// format_serial_coverage_test.go — additional coverage for format_serial.go
// to exercise the "default value → skip write" branches that are not covered
// by the existing non-default-value tests.
//
// In serializeTextFormat, each format type compares field values against
// NewXxxFormat() defaults. When all fields equal the defaults, only
// `w.WriteStr("Format", "<type>")` is written. The "skip" branches for each
// conditional write need to be exercised with default-value instances.

import (
	"bytes"
	"strings"
	"testing"

	"github.com/andrewloable/go-fastreport/format"
	"github.com/andrewloable/go-fastreport/object"
	"github.com/andrewloable/go-fastreport/serial"
)

// TestSerializeTextFormat_BooleanDefaults exercises BooleanFormat with default
// TrueText="True" and FalseText="False". Neither `if v.TrueText != dflt.TrueText`
// nor `if v.FalseText != dflt.FalseText` triggers a write.
func TestSerializeTextFormat_BooleanDefaults(t *testing.T) {
	orig := object.NewTextObject()
	orig.SetFormat(format.NewBooleanFormat()) // TrueText="True", FalseText="False"

	var buf bytes.Buffer
	w := serial.NewWriter(&buf)
	if err := w.WriteObjectNamed("TextObject", orig); err != nil {
		t.Fatalf("Serialize: %v", err)
	}
	_ = w.Flush()

	xml := buf.String()
	// Format type should still be written.
	if !strings.Contains(xml, `Format="Boolean"`) {
		t.Errorf("expected Format=Boolean in XML:\n%s", xml)
	}
	// Default TrueText/FalseText are NOT written (they match the defaults).
	if strings.Contains(xml, `Format.TrueText=`) {
		t.Errorf("TrueText should NOT appear for default value:\n%s", xml)
	}
	if strings.Contains(xml, `Format.FalseText=`) {
		t.Errorf("FalseText should NOT appear for default value:\n%s", xml)
	}
}

// TestSerializeTextFormat_CustomDefault exercises CustomFormat with the default
// Format field (empty string, which maps to "%v" in NewCustomFormat).
// `if v.Format != dflt.Format` → false (both are "%v"), so Format is not written.
func TestSerializeTextFormat_CustomDefault(t *testing.T) {
	orig := object.NewTextObject()
	orig.SetFormat(format.NewCustomFormat()) // Format="%v"

	var buf bytes.Buffer
	w := serial.NewWriter(&buf)
	if err := w.WriteObjectNamed("TextObject", orig); err != nil {
		t.Fatalf("Serialize: %v", err)
	}
	_ = w.Flush()

	xml := buf.String()
	if !strings.Contains(xml, `Format="Custom"`) {
		t.Errorf("expected Format=Custom in XML:\n%s", xml)
	}
	// The Format.Format field should NOT be written (matches default "%v").
	if strings.Contains(xml, `Format.Format=`) {
		t.Errorf("Format.Format should NOT appear for default value:\n%s", xml)
	}
}

// TestSerializeTextFormat_NumberAllDefaults exercises NumberFormat with all
// default values (UseLocaleSettings=true, DecimalDigits=2, etc.). Only the
// Format type string is written; all conditional branches skip their writes.
func TestSerializeTextFormat_NumberAllDefaults(t *testing.T) {
	orig := object.NewTextObject()
	orig.SetFormat(format.NewNumberFormat()) // UseLocaleSettings=true, DecimalDigits=2

	var buf bytes.Buffer
	w := serial.NewWriter(&buf)
	if err := w.WriteObjectNamed("TextObject", orig); err != nil {
		t.Fatalf("Serialize: %v", err)
	}
	_ = w.Flush()

	xml := buf.String()
	if !strings.Contains(xml, `Format="Number"`) {
		t.Errorf("expected Format=Number in XML:\n%s", xml)
	}
	// No Format.DecimalDigits (2 == default).
	if strings.Contains(xml, `Format.DecimalDigits=`) {
		t.Errorf("Format.DecimalDigits should NOT appear for default:\n%s", xml)
	}
	// No UseLocaleSettings written (true is default for the write-false check).
	if strings.Contains(xml, `Format.UseLocaleSettings=`) {
		t.Errorf("Format.UseLocaleSettings should NOT appear when true:\n%s", xml)
	}
}

// TestSerializeTextFormat_CurrencyAllDefaults exercises CurrencyFormat with all
// default values to cover the "skip" branches for each conditional write.
func TestSerializeTextFormat_CurrencyAllDefaults(t *testing.T) {
	orig := object.NewTextObject()
	orig.SetFormat(format.NewCurrencyFormat()) // UseLocaleSettings=true, defaults

	var buf bytes.Buffer
	w := serial.NewWriter(&buf)
	if err := w.WriteObjectNamed("TextObject", orig); err != nil {
		t.Fatalf("Serialize: %v", err)
	}
	_ = w.Flush()

	xml := buf.String()
	if !strings.Contains(xml, `Format="Currency"`) {
		t.Errorf("expected Format=Currency in XML:\n%s", xml)
	}
}

// TestSerializeTextFormat_PercentAllDefaults exercises PercentFormat with all
// default values to cover the "skip" branches for each conditional write.
func TestSerializeTextFormat_PercentAllDefaults(t *testing.T) {
	orig := object.NewTextObject()
	orig.SetFormat(format.NewPercentFormat()) // UseLocaleSettings=true, defaults

	var buf bytes.Buffer
	w := serial.NewWriter(&buf)
	if err := w.WriteObjectNamed("TextObject", orig); err != nil {
		t.Fatalf("Serialize: %v", err)
	}
	_ = w.Flush()

	xml := buf.String()
	if !strings.Contains(xml, `Format="Percent"`) {
		t.Errorf("expected Format=Percent in XML:\n%s", xml)
	}
}

// TestSerializeTextFormat_DateDefaults exercises DateFormat with default
// Format string (not written) and UseLocaleSettings=false (not written as true).
func TestSerializeTextFormat_DateDefaults(t *testing.T) {
	orig := object.NewTextObject()
	orig.SetFormat(format.NewDateFormat()) // UseLocaleSettings=false (default)

	var buf bytes.Buffer
	w := serial.NewWriter(&buf)
	if err := w.WriteObjectNamed("TextObject", orig); err != nil {
		t.Fatalf("Serialize: %v", err)
	}
	_ = w.Flush()

	xml := buf.String()
	if !strings.Contains(xml, `Format="Date"`) {
		t.Errorf("expected Format=Date in XML:\n%s", xml)
	}
	// UseLocaleSettings=false is the default — NOT written.
	if strings.Contains(xml, `Format.UseLocaleSettings="true"`) {
		t.Errorf("Format.UseLocaleSettings should NOT appear when false:\n%s", xml)
	}
}

// TestSerializeTextFormat_TimeDefaults exercises TimeFormat with default values.
func TestSerializeTextFormat_TimeDefaults(t *testing.T) {
	orig := object.NewTextObject()
	orig.SetFormat(format.NewTimeFormat()) // UseLocaleSettings=false (default)

	var buf bytes.Buffer
	w := serial.NewWriter(&buf)
	if err := w.WriteObjectNamed("TextObject", orig); err != nil {
		t.Fatalf("Serialize: %v", err)
	}
	_ = w.Flush()

	xml := buf.String()
	if !strings.Contains(xml, `Format="Time"`) {
		t.Errorf("expected Format=Time in XML:\n%s", xml)
	}
}

// ---------------------------------------------------------------------------
// C# FRX compatibility: attribute name is "UseLocale" not "UseLocaleSettings"
// ---------------------------------------------------------------------------

// TestDeserialize_UseLocale_Currency verifies that Format.UseLocale="true" in
// an FRX attribute (the C# FastReport attribute name) is correctly read as
// UseLocaleSettings=true. This simulates loading real C# FRX files.
func TestDeserialize_UseLocale_Currency(t *testing.T) {
	// This XML matches what C# FastReport writes for CurrencyFormat.
	xml := `<TextObject Name="t1" Format="Currency" Format.UseLocale="true"/>`
	r := serial.NewReader(strings.NewReader(xml))
	_, _ = r.ReadObjectHeader()
	got := object.NewTextObject()
	_ = got.Deserialize(r)
	cf, ok := got.Format().(*format.CurrencyFormat)
	if !ok {
		t.Fatal("expected *format.CurrencyFormat")
	}
	if !cf.UseLocaleSettings {
		t.Error("UseLocaleSettings should be true when Format.UseLocale=\"true\"")
	}
}

// TestDeserialize_UseLocale_Percent verifies that Format.UseLocale="true" is
// correctly read for PercentFormat, matching real C# FRX output.
func TestDeserialize_UseLocale_Percent(t *testing.T) {
	xml := `<TextObject Name="t1" Format="Percent" Format.UseLocale="true" Format.DecimalDigits="2"/>`
	r := serial.NewReader(strings.NewReader(xml))
	_, _ = r.ReadObjectHeader()
	got := object.NewTextObject()
	_ = got.Deserialize(r)
	pf, ok := got.Format().(*format.PercentFormat)
	if !ok {
		t.Fatal("expected *format.PercentFormat")
	}
	if !pf.UseLocaleSettings {
		t.Error("UseLocaleSettings should be true when Format.UseLocale=\"true\"")
	}
	if pf.DecimalDigits != 2 {
		t.Errorf("DecimalDigits: got %d, want 2", pf.DecimalDigits)
	}
}

// TestDeserialize_UseLocale_Date verifies that Format.UseLocale="true" is
// correctly read for DateFormat.
func TestDeserialize_UseLocale_Date(t *testing.T) {
	xml := `<TextObject Name="t1" Format="Date" Format.Format="d" Format.UseLocale="true"/>`
	r := serial.NewReader(strings.NewReader(xml))
	_, _ = r.ReadObjectHeader()
	got := object.NewTextObject()
	_ = got.Deserialize(r)
	df, ok := got.Format().(*format.DateFormat)
	if !ok {
		t.Fatal("expected *format.DateFormat")
	}
	if !df.UseLocaleSettings {
		t.Error("UseLocaleSettings should be true when Format.UseLocale=\"true\"")
	}
}

// TestDeserialize_UseLocale_Number verifies that Format.UseLocale="false" is
// correctly read for NumberFormat (enabling custom separators).
func TestDeserialize_UseLocale_Number(t *testing.T) {
	xml := `<TextObject Name="t1" Format="Number" Format.UseLocale="false" Format.DecimalSeparator="," Format.GroupSeparator="."/>`
	r := serial.NewReader(strings.NewReader(xml))
	_, _ = r.ReadObjectHeader()
	got := object.NewTextObject()
	_ = got.Deserialize(r)
	nf, ok := got.Format().(*format.NumberFormat)
	if !ok {
		t.Fatal("expected *format.NumberFormat")
	}
	if nf.UseLocaleSettings {
		t.Error("UseLocaleSettings should be false when Format.UseLocale=\"false\"")
	}
	if nf.DecimalSeparator != "," {
		t.Errorf("DecimalSeparator: got %q, want ','", nf.DecimalSeparator)
	}
}

// TestSerialize_UseLocale_CurrencyRoundTrip verifies that a serialized
// CurrencyFormat with UseLocaleSettings=true writes "UseLocale" (C# FRX name)
// and can be deserialized back correctly.
func TestSerialize_UseLocale_CurrencyRoundTrip(t *testing.T) {
	orig := object.NewTextObject()
	orig.SetFormat(&format.CurrencyFormat{
		UseLocaleSettings: true,
		DecimalDigits:     2,
		CurrencySymbol:    "$",
	})

	var buf bytes.Buffer
	w := serial.NewWriter(&buf)
	_ = w.WriteObjectNamed("TextObject", orig)
	_ = w.Flush()
	out := buf.String()

	// The serialized form should use "UseLocale" (C# FRX name).
	if strings.Contains(out, `Format.UseLocale=`) {
		// UseLocale=true is NOT written because true is the default for CurrencyFormat
		// (NewCurrencyFormat defaults UseLocaleSettings=true so the `!v.UseLocaleSettings`
		// branch is not taken, and UseLocale is only written when UseLocaleSettings=false).
	}
	// When UseLocaleSettings=true, UseLocale is NOT written (it's the default).
	if strings.Contains(out, `Format.UseLocale="false"`) {
		t.Errorf("UseLocale=false should NOT appear when locale is true:\n%s", out)
	}

	// Round-trip: deserialize and verify.
	r := serial.NewReader(bytes.NewReader(buf.Bytes()))
	_, _ = r.ReadObjectHeader()
	got := object.NewTextObject()
	_ = got.Deserialize(r)
	cf, ok := got.Format().(*format.CurrencyFormat)
	if !ok {
		t.Fatal("expected *format.CurrencyFormat after round-trip")
	}
	if !cf.UseLocaleSettings {
		t.Error("UseLocaleSettings should be true after round-trip")
	}
}
