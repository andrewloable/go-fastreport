package data

import (
	"testing"
)

func TestVirtualDataSource_Basic(t *testing.T) {
	ds := NewVirtualDataSource("test", 3)

	if ds.Name() != "test" {
		t.Fatalf("Name: got %q, want %q", ds.Name(), "test")
	}
	if ds.RowCount() != 3 {
		t.Fatalf("RowCount: got %d, want 3", ds.RowCount())
	}

	if err := ds.Init(); err != nil {
		t.Fatalf("Init: %v", err)
	}

	if err := ds.First(); err != nil {
		t.Fatalf("First: %v", err)
	}
	if ds.EOF() {
		t.Fatal("EOF after First on non-empty source")
	}

	rowsSeen := 0
	for !ds.EOF() {
		rowsSeen++
		val, err := ds.GetValue("anything")
		if err != nil {
			t.Fatalf("GetValue: %v", err)
		}
		if val != nil {
			t.Errorf("row %d: expected nil value, got %v", rowsSeen, val)
		}
		_ = ds.Next()
	}

	if rowsSeen != 3 {
		t.Errorf("rows seen: got %d, want 3", rowsSeen)
	}
}

func TestVirtualDataSource_ZeroRows(t *testing.T) {
	ds := NewVirtualDataSource("empty", 0)
	if err := ds.Init(); err != nil {
		t.Fatal(err)
	}
	err := ds.First()
	if err != ErrEOF {
		t.Fatalf("First on 0-row source: got %v, want ErrEOF", err)
	}
	if !ds.EOF() {
		t.Error("EOF should be true for 0-row source")
	}
}

func TestVirtualDataSource_SetRowsCount(t *testing.T) {
	ds := NewVirtualDataSource("v", 5)
	ds.SetRowsCount(2)
	if ds.RowsCount() != 2 {
		t.Errorf("RowsCount after set: got %d, want 2", ds.RowsCount())
	}
}

func TestVirtualDataSource_ImplementsInterface(t *testing.T) {
	var _ DataSource = (*VirtualDataSource)(nil)
}
