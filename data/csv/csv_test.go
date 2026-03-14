package csv_test

import (
	"os"
	"path/filepath"
	"testing"

	csvdata "github.com/andrewloable/go-fastreport/data/csv"
)

func TestNew(t *testing.T) {
	ds := csvdata.New("test")
	if ds == nil {
		t.Fatal("New returned nil")
	}
	if ds.Name() != "test" {
		t.Errorf("Name = %q, want test", ds.Name())
	}
	if ds.Separator() != ',' {
		t.Errorf("default Separator = %q, want ','", ds.Separator())
	}
	if !ds.HasHeader() {
		t.Error("default HasHeader should be true")
	}
}

func TestSetters(t *testing.T) {
	ds := csvdata.New("x")
	ds.SetFilePath("/tmp/data.csv")
	ds.SetCSV("a,b\n1,2")
	ds.SetSeparator(';')
	ds.SetHasHeader(false)
	ds.SetComment('#')
	ds.SetLazyQuotes(true)

	if ds.FilePath() != "/tmp/data.csv" {
		t.Errorf("FilePath = %q", ds.FilePath())
	}
	if ds.CSV() != "a,b\n1,2" {
		t.Errorf("CSV = %q", ds.CSV())
	}
	if ds.Separator() != ';' {
		t.Errorf("Separator = %q", ds.Separator())
	}
	if ds.HasHeader() {
		t.Error("HasHeader should be false")
	}
	if ds.Comment() != '#' {
		t.Errorf("Comment = %q", ds.Comment())
	}
	if !ds.LazyQuotes() {
		t.Error("LazyQuotes should be true")
	}
}

const simpleCSV = `id,name,score
1,Alice,9.5
2,Bob,7.0
3,Charlie,8.3`

func TestInit_FlatCSV(t *testing.T) {
	ds := csvdata.New("people")
	ds.SetCSV(simpleCSV)
	if err := ds.Init(); err != nil {
		t.Fatalf("Init: %v", err)
	}
	if ds.RowCount() != 3 {
		t.Errorf("RowCount = %d, want 3", ds.RowCount())
	}
}

func TestColumns(t *testing.T) {
	ds := csvdata.New("people")
	ds.SetCSV(simpleCSV)
	_ = ds.Init()
	cols := ds.Columns()
	if len(cols) != 3 {
		t.Fatalf("cols len = %d, want 3", len(cols))
	}
	if cols[0].Name != "id" || cols[1].Name != "name" || cols[2].Name != "score" {
		t.Errorf("col names = %v", cols)
	}
}

func TestGetValue_String(t *testing.T) {
	ds := csvdata.New("people")
	ds.SetCSV(simpleCSV)
	_ = ds.Init()
	_ = ds.First()

	v, err := ds.GetValue("name")
	if err != nil {
		t.Fatalf("GetValue: %v", err)
	}
	if v != "Alice" {
		t.Errorf("name = %v, want Alice", v)
	}
}

func TestGetValue_MissingColumn(t *testing.T) {
	ds := csvdata.New("people")
	ds.SetCSV(simpleCSV)
	_ = ds.Init()
	_ = ds.First()

	v, err := ds.GetValue("nonexistent")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if v != nil {
		t.Errorf("expected nil for missing column, got %v", v)
	}
}

func TestNext_AllRows(t *testing.T) {
	ds := csvdata.New("people")
	ds.SetCSV(simpleCSV)
	_ = ds.Init()
	_ = ds.First()

	count := 1
	for ds.Next() == nil {
		count++
	}
	if count != 3 {
		t.Errorf("iterated %d rows, want 3", count)
	}
	if !ds.EOF() {
		t.Error("EOF should be true after last row")
	}
}

func TestInit_NoHeader(t *testing.T) {
	raw := "1,Alice\n2,Bob"
	ds := csvdata.New("people")
	ds.SetCSV(raw)
	ds.SetHasHeader(false)
	if err := ds.Init(); err != nil {
		t.Fatalf("Init: %v", err)
	}
	if ds.RowCount() != 2 {
		t.Errorf("RowCount = %d, want 2", ds.RowCount())
	}
	cols := ds.Columns()
	if len(cols) != 2 {
		t.Fatalf("cols len = %d, want 2", len(cols))
	}
	if cols[0].Name != "col0" || cols[1].Name != "col1" {
		t.Errorf("col names = %v", cols)
	}
	_ = ds.First()
	v, _ := ds.GetValue("col1")
	if v != "Alice" {
		t.Errorf("col1 = %v, want Alice", v)
	}
}

func TestInit_SemicolonSeparator(t *testing.T) {
	raw := "id;name\n1;Alice\n2;Bob"
	ds := csvdata.New("people")
	ds.SetCSV(raw)
	ds.SetSeparator(';')
	if err := ds.Init(); err != nil {
		t.Fatalf("Init: %v", err)
	}
	if ds.RowCount() != 2 {
		t.Errorf("RowCount = %d, want 2", ds.RowCount())
	}
}

func TestInit_QuotedFields(t *testing.T) {
	raw := `name,bio
"Alice","She said ""hello"""
"Bob","Enjoys, hiking"`
	ds := csvdata.New("people")
	ds.SetCSV(raw)
	if err := ds.Init(); err != nil {
		t.Fatalf("Init: %v", err)
	}
	_ = ds.First()
	v, _ := ds.GetValue("bio")
	if v != `She said "hello"` {
		t.Errorf("bio = %q", v)
	}
}

func TestInit_EmptyCSV(t *testing.T) {
	ds := csvdata.New("empty")
	ds.SetCSV("")
	if err := ds.Init(); err != nil {
		t.Fatalf("Init: %v", err)
	}
	if ds.RowCount() != 0 {
		t.Errorf("RowCount = %d, want 0", ds.RowCount())
	}
}

func TestInit_HeaderOnly(t *testing.T) {
	ds := csvdata.New("hdr")
	ds.SetCSV("id,name")
	if err := ds.Init(); err != nil {
		t.Fatalf("Init: %v", err)
	}
	if ds.RowCount() != 0 {
		t.Errorf("RowCount = %d, want 0", ds.RowCount())
	}
	if len(ds.Columns()) != 2 {
		t.Errorf("Columns = %d, want 2", len(ds.Columns()))
	}
}

func TestInit_FileSource(t *testing.T) {
	tmp := t.TempDir()
	fpath := filepath.Join(tmp, "data.csv")
	if err := os.WriteFile(fpath, []byte(simpleCSV), 0644); err != nil {
		t.Fatal(err)
	}
	ds := csvdata.New("file")
	ds.SetFilePath(fpath)
	if err := ds.Init(); err != nil {
		t.Fatalf("Init: %v", err)
	}
	if ds.RowCount() != 3 {
		t.Errorf("RowCount = %d, want 3", ds.RowCount())
	}
}

func TestInit_NoSource(t *testing.T) {
	ds := csvdata.New("empty")
	if err := ds.Init(); err == nil {
		t.Error("expected error for no source")
	}
}

func TestInit_CommentLines(t *testing.T) {
	raw := "id,name\n# this is a comment\n1,Alice"
	ds := csvdata.New("people")
	ds.SetCSV(raw)
	ds.SetComment('#')
	if err := ds.Init(); err != nil {
		t.Fatalf("Init: %v", err)
	}
	if ds.RowCount() != 1 {
		t.Errorf("RowCount = %d, want 1", ds.RowCount())
	}
}

func TestInit_ShortRow(t *testing.T) {
	// Row with fewer fields than header — missing fields should be empty string.
	raw := "id,name,score\n1,Alice"
	ds := csvdata.New("people")
	ds.SetCSV(raw)
	if err := ds.Init(); err != nil {
		t.Fatalf("Init: %v", err)
	}
	_ = ds.First()
	v, _ := ds.GetValue("score")
	if v != "" {
		t.Errorf("score = %v, want empty string", v)
	}
}

func TestClose(t *testing.T) {
	ds := csvdata.New("people")
	ds.SetCSV(simpleCSV)
	_ = ds.Init()
	if err := ds.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}
}
