package json_test

import (
	"os"
	"path/filepath"
	"testing"

	jsondata "github.com/andrewloable/go-fastreport/data/json"
)

func TestNew(t *testing.T) {
	ds := jsondata.New("test")
	if ds == nil {
		t.Fatal("New returned nil")
	}
	if ds.Name() != "test" {
		t.Errorf("Name = %q, want test", ds.Name())
	}
}

func TestSetFilePath(t *testing.T) {
	ds := jsondata.New("x")
	ds.SetFilePath("/tmp/data.json")
	if ds.FilePath() != "/tmp/data.json" {
		t.Errorf("FilePath = %q", ds.FilePath())
	}
}

func TestSetJSON(t *testing.T) {
	ds := jsondata.New("x")
	ds.SetJSON(`[{"a":1}]`)
	if ds.JSON() != `[{"a":1}]` {
		t.Errorf("JSON = %q", ds.JSON())
	}
}

func TestSetRootPath(t *testing.T) {
	ds := jsondata.New("x")
	ds.SetRootPath("data.items")
	if ds.RootPath() != "data.items" {
		t.Errorf("RootPath = %q", ds.RootPath())
	}
}

const flatJSON = `[
  {"id":1,"name":"Alice","score":9.5},
  {"id":2,"name":"Bob","score":7.0},
  {"id":3,"name":"Charlie","score":8.3}
]`

func TestInit_FlatArray(t *testing.T) {
	ds := jsondata.New("people")
	ds.SetJSON(flatJSON)
	if err := ds.Init(); err != nil {
		t.Fatalf("Init error: %v", err)
	}
	if ds.RowCount() != 3 {
		t.Errorf("RowCount = %d, want 3", ds.RowCount())
	}
}

func TestInit_Navigation(t *testing.T) {
	js := `{"id":1,"name":"Alice","score":9.5,"active":true}`
	ds := jsondata.New("person")
	ds.SetJSON(js)
	if err := ds.Init(); err != nil {
		t.Fatalf("Init: %v", err)
	}
	if ds.RowCount() != 1 {
		t.Errorf("RowCount = %d, want 1", ds.RowCount())
	}
}

func TestGetValue_StringField(t *testing.T) {
	ds := jsondata.New("people")
	ds.SetJSON(flatJSON)
	_ = ds.Init()
	_ = ds.First()

	v, err := ds.GetValue("name")
	if err != nil {
		t.Fatalf("GetValue error: %v", err)
	}
	if v != "Alice" {
		t.Errorf("name = %v, want Alice", v)
	}
}

func TestGetValue_NumberField(t *testing.T) {
	ds := jsondata.New("people")
	ds.SetJSON(flatJSON)
	_ = ds.Init()
	_ = ds.First()

	v, err := ds.GetValue("score")
	if err != nil {
		t.Fatalf("GetValue error: %v", err)
	}
	if v.(float64) != 9.5 {
		t.Errorf("score = %v, want 9.5", v)
	}
}

func TestGetValue_MissingColumn(t *testing.T) {
	ds := jsondata.New("people")
	ds.SetJSON(flatJSON)
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

func TestNext_EOF(t *testing.T) {
	ds := jsondata.New("people")
	ds.SetJSON(flatJSON)
	_ = ds.Init()
	_ = ds.First()
	_ = ds.Next()
	_ = ds.Next()
	err := ds.Next()
	if err == nil {
		t.Error("expected EOF error after last row")
	}
	if !ds.EOF() {
		t.Error("EOF should be true")
	}
}

func TestInit_RootPath(t *testing.T) {
	js := `{"data":{"items":[{"x":1},{"x":2}]}}`
	ds := jsondata.New("items")
	ds.SetJSON(js)
	ds.SetRootPath("data.items")
	if err := ds.Init(); err != nil {
		t.Fatalf("Init: %v", err)
	}
	if ds.RowCount() != 2 {
		t.Errorf("RowCount = %d, want 2", ds.RowCount())
	}
	_ = ds.First()
	v, _ := ds.GetValue("x")
	if v.(float64) != 1 {
		t.Errorf("x = %v, want 1", v)
	}
}

func TestInit_NestedObject(t *testing.T) {
	js := `[{"id":1,"address":{"city":"NYC","zip":"10001"}}]`
	ds := jsondata.New("people")
	ds.SetJSON(js)
	if err := ds.Init(); err != nil {
		t.Fatalf("Init: %v", err)
	}
	_ = ds.First()
	v, _ := ds.GetValue("address")
	if v == nil {
		t.Error("address should not be nil")
	}
}

func TestInit_NullValues(t *testing.T) {
	js := `[{"id":1,"name":null},{"id":2,"name":"Bob"}]`
	ds := jsondata.New("people")
	ds.SetJSON(js)
	if err := ds.Init(); err != nil {
		t.Fatalf("Init: %v", err)
	}
	_ = ds.First()
	v, err := ds.GetValue("name")
	if err != nil {
		t.Fatalf("GetValue: %v", err)
	}
	if v != nil {
		t.Errorf("expected nil for null JSON value, got %v", v)
	}
}

func TestInit_FileSource(t *testing.T) {
	tmp := t.TempDir()
	fpath := filepath.Join(tmp, "data.json")
	if err := os.WriteFile(fpath, []byte(flatJSON), 0644); err != nil {
		t.Fatal(err)
	}
	ds := jsondata.New("file")
	ds.SetFilePath(fpath)
	if err := ds.Init(); err != nil {
		t.Fatalf("Init: %v", err)
	}
	if ds.RowCount() != 3 {
		t.Errorf("RowCount = %d, want 3", ds.RowCount())
	}
}

func TestInit_NoSource(t *testing.T) {
	ds := jsondata.New("empty")
	err := ds.Init()
	if err == nil {
		t.Error("expected error for no source configured")
	}
}

func TestInit_InvalidJSON(t *testing.T) {
	ds := jsondata.New("bad")
	ds.SetJSON(`not valid json`)
	err := ds.Init()
	if err == nil {
		t.Error("expected error for invalid JSON")
	}
}

func TestInit_InvalidRootPath(t *testing.T) {
	ds := jsondata.New("bad")
	ds.SetJSON(`{"a":1}`)
	ds.SetRootPath("missing.key")
	err := ds.Init()
	if err == nil {
		t.Error("expected error for missing root path")
	}
}

func TestInit_ScalarArray(t *testing.T) {
	js := `[1, 2, 3]`
	ds := jsondata.New("nums")
	ds.SetJSON(js)
	if err := ds.Init(); err != nil {
		t.Fatalf("Init: %v", err)
	}
	if ds.RowCount() != 3 {
		t.Errorf("RowCount = %d, want 3", ds.RowCount())
	}
	_ = ds.First()
	v, _ := ds.GetValue("_value")
	if v.(float64) != 1 {
		t.Errorf("_value = %v, want 1", v)
	}
}

func TestClose(t *testing.T) {
	ds := jsondata.New("people")
	ds.SetJSON(flatJSON)
	_ = ds.Init()
	if err := ds.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}
}

func TestColumns(t *testing.T) {
	ds := jsondata.New("people")
	ds.SetJSON(flatJSON)
	_ = ds.Init()
	cols := ds.Columns()
	if len(cols) != 3 {
		t.Errorf("Columns len = %d, want 3", len(cols))
	}
}
