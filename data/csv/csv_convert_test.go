package csv_test

// csv_convert_test.go covers ConvertFieldTypes type inference, NewFromConnectionString,
// and ConnectionStringBuilder.String() — ported from C# CsvUtils.DetermineTypes and
// CsvDataConnection property-setter pattern via CheckForChangeConnection.

import (
	"testing"
	"time"

	csvdata "github.com/andrewloable/go-fastreport/data/csv"
)

// ─── ConvertFieldTypes ────────────────────────────────────────────────────────

func TestConvertFieldTypes_IntColumn(t *testing.T) {
	raw := "id,name\n1,Alice\n2,Bob\n3,Charlie"
	ds := csvdata.New("people")
	ds.SetCSV(raw)
	ds.SetConvertFieldTypes(true)
	if err := ds.Init(); err != nil {
		t.Fatalf("Init: %v", err)
	}

	// id column should be inferred as int.
	cols := ds.Columns()
	if len(cols) != 2 {
		t.Fatalf("cols len = %d, want 2", len(cols))
	}
	if cols[0].DataType != "int" {
		t.Errorf("id DataType = %q, want int", cols[0].DataType)
	}
	if cols[1].DataType != "string" {
		t.Errorf("name DataType = %q, want string", cols[1].DataType)
	}

	_ = ds.First()
	v, err := ds.GetValue("id")
	if err != nil {
		t.Fatalf("GetValue: %v", err)
	}
	if v != 1 {
		t.Errorf("id[0] = %v (%T), want 1 (int)", v, v)
	}
}

func TestConvertFieldTypes_FloatColumn(t *testing.T) {
	raw := "score\n9.5\n7.0\n8.3"
	ds := csvdata.New("scores")
	ds.SetCSV(raw)
	ds.SetConvertFieldTypes(true)
	if err := ds.Init(); err != nil {
		t.Fatalf("Init: %v", err)
	}

	cols := ds.Columns()
	if cols[0].DataType != "float64" {
		t.Errorf("score DataType = %q, want float64", cols[0].DataType)
	}

	_ = ds.First()
	v, _ := ds.GetValue("score")
	fv, ok := v.(float64)
	if !ok {
		t.Fatalf("score type = %T, want float64", v)
	}
	if fv != 9.5 {
		t.Errorf("score[0] = %v, want 9.5", fv)
	}
}

func TestConvertFieldTypes_MixedIntAndFloat(t *testing.T) {
	// Mix of int-parseable and float-only → promoted to float64.
	// Mirrors C# DetermineTypes: int+double → double.
	raw := "val\n1\n2.5\n3"
	ds := csvdata.New("vals")
	ds.SetCSV(raw)
	ds.SetConvertFieldTypes(true)
	if err := ds.Init(); err != nil {
		t.Fatalf("Init: %v", err)
	}
	cols := ds.Columns()
	if cols[0].DataType != "float64" {
		t.Errorf("mixed int+float DataType = %q, want float64", cols[0].DataType)
	}
}

func TestConvertFieldTypes_DateTimeColumn(t *testing.T) {
	raw := "ts\n2024-01-15\n2024-06-30\n2024-12-31"
	ds := csvdata.New("times")
	ds.SetCSV(raw)
	ds.SetConvertFieldTypes(true)
	if err := ds.Init(); err != nil {
		t.Fatalf("Init: %v", err)
	}
	cols := ds.Columns()
	if cols[0].DataType != "time.Time" {
		t.Errorf("ts DataType = %q, want time.Time", cols[0].DataType)
	}

	_ = ds.First()
	v, _ := ds.GetValue("ts")
	tv, ok := v.(time.Time)
	if !ok {
		t.Fatalf("ts type = %T, want time.Time", v)
	}
	if tv.Year() != 2024 || tv.Month() != 1 || tv.Day() != 15 {
		t.Errorf("ts[0] = %v, want 2024-01-15", tv)
	}
}

func TestConvertFieldTypes_StringFallback(t *testing.T) {
	// A column with mixed text → string.
	raw := "mixed\nhello\n42\nworld"
	ds := csvdata.New("mixed")
	ds.SetCSV(raw)
	ds.SetConvertFieldTypes(true)
	if err := ds.Init(); err != nil {
		t.Fatalf("Init: %v", err)
	}
	cols := ds.Columns()
	if cols[0].DataType != "string" {
		t.Errorf("mixed DataType = %q, want string", cols[0].DataType)
	}
}

func TestConvertFieldTypes_EmptyCellsIgnored(t *testing.T) {
	// Empty cells should not affect type inference.
	raw := "id\n1\n\n3"
	ds := csvdata.New("sparse")
	ds.SetCSV(raw)
	ds.SetConvertFieldTypes(true)
	if err := ds.Init(); err != nil {
		t.Fatalf("Init: %v", err)
	}
	cols := ds.Columns()
	if cols[0].DataType != "int" {
		t.Errorf("id DataType = %q, want int (empty cells should be ignored)", cols[0].DataType)
	}
}

func TestConvertFieldTypes_Disabled_AllStrings(t *testing.T) {
	// With ConvertFieldTypes=false (default), all columns remain string.
	raw := "id,score\n1,9.5\n2,7.0"
	ds := csvdata.New("data")
	ds.SetCSV(raw)
	// do NOT call SetConvertFieldTypes(true)
	if err := ds.Init(); err != nil {
		t.Fatalf("Init: %v", err)
	}
	for _, col := range ds.Columns() {
		if col.DataType != "string" {
			t.Errorf("col %q DataType = %q, want string when ConvertFieldTypes=false", col.Name, col.DataType)
		}
	}
}

func TestConvertFieldTypes_GetterSetter(t *testing.T) {
	ds := csvdata.New("x")
	if ds.ConvertFieldTypes() {
		t.Error("default ConvertFieldTypes should be false")
	}
	ds.SetConvertFieldTypes(true)
	if !ds.ConvertFieldTypes() {
		t.Error("ConvertFieldTypes should be true after SetConvertFieldTypes(true)")
	}
}

// ─── NewFromConnectionString ──────────────────────────────────────────────────

func TestNewFromConnectionString_BasicProperties(t *testing.T) {
	cs := "CsvFile=/tmp/test.csv;Separator=,;FieldNamesInFirstString=true;ConvertFieldTypes=true"
	ds := csvdata.NewFromConnectionString("ds1", cs)
	if ds == nil {
		t.Fatal("NewFromConnectionString returned nil")
	}
	if ds.Name() != "ds1" {
		t.Errorf("Name = %q, want ds1", ds.Name())
	}
	if ds.FilePath() != "/tmp/test.csv" {
		t.Errorf("FilePath = %q, want /tmp/test.csv", ds.FilePath())
	}
	if ds.Separator() != ',' {
		t.Errorf("Separator = %q, want ','", ds.Separator())
	}
	if !ds.HasHeader() {
		t.Error("HasHeader should be true (FieldNamesInFirstString=true)")
	}
	if !ds.ConvertFieldTypes() {
		t.Error("ConvertFieldTypes should be true")
	}
}

func TestNewFromConnectionString_DefaultSeparator(t *testing.T) {
	// When Separator is omitted, ConnectionStringBuilder returns ";".
	// NewFromConnectionString should use the first rune of that default → ';'.
	cs := "CsvFile=/tmp/test.csv"
	ds := csvdata.NewFromConnectionString("ds", cs)
	if ds.Separator() != ';' {
		t.Errorf("Separator = %q, want ';' (ConnectionStringBuilder default)", ds.Separator())
	}
}

func TestNewFromConnectionString_EmptyConnectionString(t *testing.T) {
	ds := csvdata.NewFromConnectionString("empty", "")
	if ds == nil {
		t.Fatal("NewFromConnectionString returned nil for empty connection string")
	}
	if ds.Name() != "empty" {
		t.Errorf("Name = %q", ds.Name())
	}
	// ConvertFieldTypes default is true from ConnectionStringBuilder.
	if !ds.ConvertFieldTypes() {
		t.Error("ConvertFieldTypes should default to true via ConnectionStringBuilder")
	}
}

// ─── ConnectionStringBuilder.String() ────────────────────────────────────────

func TestConnectionStringBuilder_String_RoundTrip(t *testing.T) {
	original := "CsvFile=/data/test.csv;Separator=,;FieldNamesInFirstString=true;ConvertFieldTypes=false"
	b := csvdata.NewConnectionStringBuilder(original)
	result := b.String()

	// Parse the result again.
	b2 := csvdata.NewConnectionStringBuilder(result)
	if b2.CsvFile() != "/data/test.csv" {
		t.Errorf("round-trip CsvFile = %q", b2.CsvFile())
	}
	if b2.Separator() != "," {
		t.Errorf("round-trip Separator = %q", b2.Separator())
	}
	if !b2.FieldNamesInFirstString() {
		t.Error("round-trip FieldNamesInFirstString should be true")
	}
	if b2.ConvertFieldTypes() {
		t.Error("round-trip ConvertFieldTypes should be false")
	}
}

func TestConnectionStringBuilder_String_Empty(t *testing.T) {
	b := csvdata.NewConnectionStringBuilder("")
	s := b.String()
	// An empty builder has no stored keys → empty string.
	if s != "" {
		t.Errorf("String() for empty builder = %q, want empty", s)
	}
}

func TestConnectionStringBuilder_Setters(t *testing.T) {
	b := csvdata.NewConnectionStringBuilder("")
	b.SetCsvFile("/tmp/x.csv")
	b.SetCodepage(65001)
	b.SetSeparator(",")
	b.SetFieldNamesInFirstString(true)
	b.SetRemoveQuotationMarks(false)
	b.SetConvertFieldTypes(true)
	b.SetNumberFormat("en-US")
	b.SetCurrencyFormat("en-GB")
	b.SetDateTimeFormat("de-DE")

	if b.CsvFile() != "/tmp/x.csv" {
		t.Errorf("CsvFile = %q", b.CsvFile())
	}
	if b.Codepage() != 65001 {
		t.Errorf("Codepage = %d", b.Codepage())
	}
	if b.Separator() != "," {
		t.Errorf("Separator = %q", b.Separator())
	}
	if !b.FieldNamesInFirstString() {
		t.Error("FieldNamesInFirstString should be true")
	}
	if b.RemoveQuotationMarks() {
		t.Error("RemoveQuotationMarks should be false")
	}
	if !b.ConvertFieldTypes() {
		t.Error("ConvertFieldTypes should be true")
	}
	if b.NumberFormat() != "en-US" {
		t.Errorf("NumberFormat = %q", b.NumberFormat())
	}
	if b.CurrencyFormat() != "en-GB" {
		t.Errorf("CurrencyFormat = %q", b.CurrencyFormat())
	}
	if b.DateTimeFormat() != "de-DE" {
		t.Errorf("DateTimeFormat = %q", b.DateTimeFormat())
	}

	// Verify String() emits canonical names.
	s := b.String()
	b3 := csvdata.NewConnectionStringBuilder(s)
	if b3.CsvFile() != "/tmp/x.csv" {
		t.Errorf("after String() round-trip CsvFile = %q", b3.CsvFile())
	}
	if b3.Codepage() != 65001 {
		t.Errorf("after String() round-trip Codepage = %d", b3.Codepage())
	}
	if b3.NumberFormat() != "en-US" {
		t.Errorf("after String() round-trip NumberFormat = %q", b3.NumberFormat())
	}
}
