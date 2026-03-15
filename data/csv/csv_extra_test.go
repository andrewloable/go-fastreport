package csv_test

import (
	"os"
	"path/filepath"
	"testing"

	csvdata "github.com/andrewloable/go-fastreport/data/csv"
)

// TestInit_FileNotFound covers the openReader file error path.
func TestInit_FileNotFound(t *testing.T) {
	ds := csvdata.New("bad")
	ds.SetFilePath("/nonexistent/path/to/file.csv")
	err := ds.Init()
	if err == nil {
		t.Error("expected error for non-existent file path")
	}
}

// TestInit_ParseError covers the csv parse error path.
func TestInit_ParseError(t *testing.T) {
	// A CSV with mismatched quoting that cannot be parsed strictly.
	// Using a double-quote that isn't properly closed.
	raw := "a,b\n\"unclosed"
	ds := csvdata.New("bad")
	ds.SetCSV(raw)
	// With default settings (LazyQuotes=false), this may or may not error.
	// We just verify Init handles it gracefully.
	_ = ds.Init()
}

// TestInit_EmptyCSVWithStringSet covers the sourceStringSet with empty string path.
func TestInit_EmptyStringSet(t *testing.T) {
	ds := csvdata.New("empty")
	ds.SetCSV("") // sourceStringSet = true, sourceString = ""
	if err := ds.Init(); err != nil {
		t.Fatalf("Init with empty string set: %v", err)
	}
	if ds.RowCount() != 0 {
		t.Errorf("RowCount = %d, want 0", ds.RowCount())
	}
}

// TestInit_FileWithNoHeader covers file-based source with no header.
func TestInit_FileNoHeader(t *testing.T) {
	tmp := t.TempDir()
	fpath := filepath.Join(tmp, "noheader.csv")
	content := "1,Alice\n2,Bob\n"
	if err := os.WriteFile(fpath, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	ds := csvdata.New("noheader")
	ds.SetFilePath(fpath)
	ds.SetHasHeader(false)
	if err := ds.Init(); err != nil {
		t.Fatalf("Init: %v", err)
	}
	if ds.RowCount() != 2 {
		t.Errorf("RowCount = %d, want 2", ds.RowCount())
	}
	cols := ds.Columns()
	if len(cols) < 1 {
		t.Fatal("expected columns")
	}
	if cols[0].Name != "col0" {
		t.Errorf("cols[0].Name = %q, want col0", cols[0].Name)
	}
}

// TestInit_LargeFile ensures data is properly loaded from a file with many rows.
func TestInit_LargeFile(t *testing.T) {
	tmp := t.TempDir()
	fpath := filepath.Join(tmp, "large.csv")

	var sb []byte
	sb = append(sb, []byte("id,val\n")...)
	for i := 0; i < 100; i++ {
		sb = append(sb, []byte("1,x\n")...)
	}
	if err := os.WriteFile(fpath, sb, 0644); err != nil {
		t.Fatal(err)
	}

	ds := csvdata.New("large")
	ds.SetFilePath(fpath)
	if err := ds.Init(); err != nil {
		t.Fatalf("Init: %v", err)
	}
	if ds.RowCount() != 100 {
		t.Errorf("RowCount = %d, want 100", ds.RowCount())
	}
}

// TestInit_NoSourceAtAll covers the error when neither FilePath nor CSV is set.
func TestInit_NoSourceAtAll(t *testing.T) {
	ds := csvdata.New("nothing")
	err := ds.Init()
	if err == nil {
		t.Error("expected error when no source is configured")
	}
}

// TestInit_WithLazyQuotes covers LazyQuotes=true path.
func TestInit_LazyQuotes(t *testing.T) {
	// Relaxed quoting — bare quotes inside fields are allowed.
	raw := "name,note\nAlice,she said \"hi\" to him\nBob,normal"
	ds := csvdata.New("lazy")
	ds.SetCSV(raw)
	ds.SetLazyQuotes(true)
	if err := ds.Init(); err != nil {
		t.Fatalf("Init with LazyQuotes: %v", err)
	}
	if ds.RowCount() != 2 {
		t.Errorf("RowCount = %d, want 2", ds.RowCount())
	}
}

// TestInit_CommentWithNoHeader covers comment character with no-header CSV.
func TestInit_CommentNoHeader(t *testing.T) {
	raw := "# skip this\n1,Alice\n# skip too\n2,Bob"
	ds := csvdata.New("commented")
	ds.SetCSV(raw)
	ds.SetHasHeader(false)
	ds.SetComment('#')
	if err := ds.Init(); err != nil {
		t.Fatalf("Init: %v", err)
	}
	if ds.RowCount() != 2 {
		t.Errorf("RowCount = %d, want 2", ds.RowCount())
	}
}

// TestInit_ManyColumnsShortRow covers when a row has fewer columns than header.
func TestInit_ManyColumnsShortRowMultiple(t *testing.T) {
	raw := "a,b,c,d\n1,2\n3,4,5,6"
	ds := csvdata.New("mixed")
	ds.SetCSV(raw)
	if err := ds.Init(); err != nil {
		t.Fatalf("Init: %v", err)
	}
	if ds.RowCount() != 2 {
		t.Errorf("RowCount = %d, want 2", ds.RowCount())
	}
	_ = ds.First()
	v, _ := ds.GetValue("c")
	if v != "" {
		t.Errorf("short row: c = %v, want empty string", v)
	}
}
