package csv

// connection_string_test.go — tests for ConnectionStringBuilder.
// Uses the csv package directly (no _test suffix) to access unexported parse/get.

import (
	"testing"
)

func TestNewConnectionStringBuilder_Empty(t *testing.T) {
	b := NewConnectionStringBuilder("")
	if b.CsvFile() != "" {
		t.Errorf("CsvFile = %q, want empty", b.CsvFile())
	}
	if b.Codepage() != 0 {
		t.Errorf("Codepage = %d, want 0", b.Codepage())
	}
	// Default separator is ";" when not in connection string.
	if b.Separator() != ";" {
		t.Errorf("Separator = %q, want ;", b.Separator())
	}
	// FieldNamesInFirstString defaults to false.
	if b.FieldNamesInFirstString() {
		t.Error("FieldNamesInFirstString should default to false")
	}
	// RemoveQuotationMarks defaults to true.
	if !b.RemoveQuotationMarks() {
		t.Error("RemoveQuotationMarks should default to true")
	}
	// ConvertFieldTypes defaults to true.
	if !b.ConvertFieldTypes() {
		t.Error("ConvertFieldTypes should default to true")
	}
	// Format strings default to empty.
	if b.NumberFormat() != "" {
		t.Errorf("NumberFormat = %q, want empty", b.NumberFormat())
	}
	if b.CurrencyFormat() != "" {
		t.Errorf("CurrencyFormat = %q, want empty", b.CurrencyFormat())
	}
	if b.DateTimeFormat() != "" {
		t.Errorf("DateTimeFormat = %q, want empty", b.DateTimeFormat())
	}
}

func TestNewConnectionStringBuilder_FullString(t *testing.T) {
	cs := "CsvFile=/data/test.csv;Codepage=65001;Separator=,;FieldNamesInFirstString=true;RemoveQuotationMarks=false;ConvertFieldTypes=false;NumberFormat=en-US;CurrencyFormat=en-US;DateTimeFormat=en-US"
	b := NewConnectionStringBuilder(cs)

	if b.CsvFile() != "/data/test.csv" {
		t.Errorf("CsvFile = %q, want /data/test.csv", b.CsvFile())
	}
	if b.Codepage() != 65001 {
		t.Errorf("Codepage = %d, want 65001", b.Codepage())
	}
	if b.Separator() != "," {
		t.Errorf("Separator = %q, want ,", b.Separator())
	}
	if !b.FieldNamesInFirstString() {
		t.Error("FieldNamesInFirstString should be true")
	}
	if b.RemoveQuotationMarks() {
		t.Error("RemoveQuotationMarks should be false")
	}
	if b.ConvertFieldTypes() {
		t.Error("ConvertFieldTypes should be false")
	}
	if b.NumberFormat() != "en-US" {
		t.Errorf("NumberFormat = %q, want en-US", b.NumberFormat())
	}
	if b.CurrencyFormat() != "en-US" {
		t.Errorf("CurrencyFormat = %q, want en-US", b.CurrencyFormat())
	}
	if b.DateTimeFormat() != "en-US" {
		t.Errorf("DateTimeFormat = %q, want en-US", b.DateTimeFormat())
	}
}

func TestConnectionStringBuilder_CaseInsensitiveKeys(t *testing.T) {
	cs := "CSVFILE=/tmp/x.csv;CODEPAGE=1252;SEPARATOR=|"
	b := NewConnectionStringBuilder(cs)
	if b.CsvFile() != "/tmp/x.csv" {
		t.Errorf("CsvFile = %q, want /tmp/x.csv", b.CsvFile())
	}
	if b.Codepage() != 1252 {
		t.Errorf("Codepage = %d, want 1252", b.Codepage())
	}
	if b.Separator() != "|" {
		t.Errorf("Separator = %q, want |", b.Separator())
	}
}

func TestConnectionStringBuilder_Codepage_NonNumeric(t *testing.T) {
	cs := "Codepage=notanumber"
	b := NewConnectionStringBuilder(cs)
	// Non-numeric codepage should return default 0.
	if b.Codepage() != 0 {
		t.Errorf("Codepage for non-numeric = %d, want 0", b.Codepage())
	}
}

func TestConnectionStringBuilder_parse_SkipsEmptyParts(t *testing.T) {
	// Extra semicolons → empty parts that should be ignored.
	cs := ";;CsvFile=/tmp/x.csv;;"
	b := NewConnectionStringBuilder(cs)
	if b.CsvFile() != "/tmp/x.csv" {
		t.Errorf("CsvFile = %q, want /tmp/x.csv", b.CsvFile())
	}
}

func TestConnectionStringBuilder_parse_SkipsNoEquals(t *testing.T) {
	// Part without '=' should be skipped silently.
	cs := "NoEqualsSign;CsvFile=/tmp/y.csv"
	b := NewConnectionStringBuilder(cs)
	if b.CsvFile() != "/tmp/y.csv" {
		t.Errorf("CsvFile = %q, want /tmp/y.csv", b.CsvFile())
	}
}

func TestConnectionStringBuilder_FieldNamesInFirstString_False(t *testing.T) {
	cs := "FieldNamesInFirstString=False"
	b := NewConnectionStringBuilder(cs)
	if b.FieldNamesInFirstString() {
		t.Error("FieldNamesInFirstString=False should return false")
	}
}

func TestConnectionStringBuilder_RemoveQuotationMarks_True(t *testing.T) {
	cs := "RemoveQuotationMarks=true"
	b := NewConnectionStringBuilder(cs)
	if !b.RemoveQuotationMarks() {
		t.Error("RemoveQuotationMarks=true should return true")
	}
}

func TestConnectionStringBuilder_ConvertFieldTypes_True(t *testing.T) {
	cs := "ConvertFieldTypes=true"
	b := NewConnectionStringBuilder(cs)
	if !b.ConvertFieldTypes() {
		t.Error("ConvertFieldTypes=true should return true")
	}
}
