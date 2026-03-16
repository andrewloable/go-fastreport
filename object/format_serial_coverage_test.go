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
