package data_test

import (
	"errors"
	"testing"

	"github.com/andrewloable/go-fastreport/data"
)

func makeDS(rows ...map[string]any) *data.BaseDataSource {
	ds := data.NewBaseDataSource("test")
	for _, r := range rows {
		ds.AddRow(r)
	}
	return ds
}

func TestBaseDataSourceName(t *testing.T) {
	ds := data.NewBaseDataSource("Orders")
	if ds.Name() != "Orders" {
		t.Errorf("Name() = %q, want Orders", ds.Name())
	}
	if ds.Alias() != "Orders" {
		t.Errorf("Alias() default should equal Name, got %q", ds.Alias())
	}
	ds.SetName("NewOrders")
	if ds.Name() != "NewOrders" {
		t.Errorf("SetName: got %q", ds.Name())
	}
	ds.SetAlias("My Orders")
	if ds.Alias() != "My Orders" {
		t.Errorf("SetAlias: got %q", ds.Alias())
	}
}

func TestBaseDataSourceColumns(t *testing.T) {
	ds := data.NewBaseDataSource("test")
	ds.AddColumn(data.Column{Name: "ID", Alias: "ID", DataType: "int"})
	ds.AddColumn(data.Column{Name: "Name", Alias: "Name", DataType: "string"})
	cols := ds.Columns()
	if len(cols) != 2 {
		t.Fatalf("expected 2 columns, got %d", len(cols))
	}
	if cols[0].Name != "ID" {
		t.Errorf("cols[0].Name = %q, want ID", cols[0].Name)
	}
}

func TestBaseDataSourceInitFirst(t *testing.T) {
	ds := makeDS(
		map[string]any{"Name": "Alice"},
		map[string]any{"Name": "Bob"},
	)

	if err := ds.Init(); err != nil {
		t.Fatalf("Init error: %v", err)
	}
	if ds.RowCount() != 2 {
		t.Errorf("RowCount = %d, want 2", ds.RowCount())
	}

	if err := ds.First(); err != nil {
		t.Fatalf("First error: %v", err)
	}
	if ds.CurrentRowNo() != 0 {
		t.Errorf("CurrentRowNo after First = %d, want 0", ds.CurrentRowNo())
	}
}

func TestBaseDataSourceNextEOF(t *testing.T) {
	ds := makeDS(
		map[string]any{"x": 1},
		map[string]any{"x": 2},
	)
	ds.Init()
	ds.First()

	if ds.EOF() {
		t.Error("EOF should be false on first row")
	}

	if err := ds.Next(); err != nil {
		t.Fatalf("Next error: %v", err)
	}
	if ds.CurrentRowNo() != 1 {
		t.Errorf("CurrentRowNo = %d, want 1", ds.CurrentRowNo())
	}

	err := ds.Next()
	if !errors.Is(err, data.ErrEOF) {
		t.Errorf("Next past end: expected ErrEOF, got %v", err)
	}
	if !ds.EOF() {
		t.Error("EOF should be true after exhausting rows")
	}
}

func TestBaseDataSourceGetValue(t *testing.T) {
	ds := makeDS(map[string]any{"Name": "Alice", "Age": 30})
	ds.Init()
	ds.First()

	v, err := ds.GetValue("Name")
	if err != nil {
		t.Fatalf("GetValue error: %v", err)
	}
	if v != "Alice" {
		t.Errorf("GetValue(Name) = %v, want Alice", v)
	}

	v, err = ds.GetValue("Age")
	if err != nil {
		t.Fatalf("GetValue(Age) error: %v", err)
	}
	if v != 30 {
		t.Errorf("GetValue(Age) = %v, want 30", v)
	}

	// Missing column returns nil
	v, err = ds.GetValue("Missing")
	if err != nil {
		t.Fatalf("GetValue(Missing) error: %v", err)
	}
	if v != nil {
		t.Errorf("GetValue(Missing) = %v, want nil", v)
	}
}

func TestBaseDataSourceGetValueNotInitialized(t *testing.T) {
	ds := data.NewBaseDataSource("test")
	_, err := ds.GetValue("col")
	if !errors.Is(err, data.ErrNotInitialized) {
		t.Errorf("expected ErrNotInitialized, got %v", err)
	}
}

func TestBaseDataSourceFirstNotInitialized(t *testing.T) {
	ds := data.NewBaseDataSource("test")
	err := ds.First()
	if !errors.Is(err, data.ErrNotInitialized) {
		t.Errorf("expected ErrNotInitialized, got %v", err)
	}
}

func TestBaseDataSourceNextNotInitialized(t *testing.T) {
	ds := data.NewBaseDataSource("test")
	err := ds.Next()
	if !errors.Is(err, data.ErrNotInitialized) {
		t.Errorf("expected ErrNotInitialized, got %v", err)
	}
}

func TestBaseDataSourceEmptyFirst(t *testing.T) {
	ds := data.NewBaseDataSource("test")
	ds.Init()
	err := ds.First()
	if !errors.Is(err, data.ErrEOF) {
		t.Errorf("First on empty datasource: expected ErrEOF, got %v", err)
	}
	if ds.CurrentRowNo() != -1 {
		t.Errorf("CurrentRowNo on empty = %d, want -1", ds.CurrentRowNo())
	}
}

func TestBaseDataSourceClose(t *testing.T) {
	ds := makeDS(map[string]any{"x": 1})
	ds.Init()
	ds.First()

	if err := ds.Close(); err != nil {
		t.Fatalf("Close error: %v", err)
	}
	// After close, datasource is no longer initialized
	err := ds.Next()
	if !errors.Is(err, data.ErrNotInitialized) {
		t.Errorf("after Close, Next should return ErrNotInitialized, got %v", err)
	}
}

func TestBaseDataSourceGetValueNoCurrentRow(t *testing.T) {
	ds := makeDS(map[string]any{"x": 1})
	ds.Init()
	// No First() called, so currentRow = -1
	_, err := ds.GetValue("x")
	if err == nil {
		t.Error("expected error when no current row, got nil")
	}
}

func TestBaseDataSourceGetValuePastEnd(t *testing.T) {
	ds := makeDS(map[string]any{"x": 1})
	ds.Init()
	ds.First()
	ds.Next() // past end
	_, err := ds.GetValue("x")
	if err == nil {
		t.Error("expected error when past end, got nil")
	}
}

func TestBaseDataSourceImplementsInterface(t *testing.T) {
	var _ data.DataSource = (*data.BaseDataSource)(nil)
}
