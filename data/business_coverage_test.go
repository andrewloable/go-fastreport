package data_test

// business_coverage_test.go — additional coverage for BusinessObjectDataSource
// edge cases: nil pointer, unexported fields, GetValue on EOF, columnsFor/fieldValue branches.

import (
	"testing"

	"github.com/andrewloable/go-fastreport/data"
)

// ── Init with nil pointer ─────────────────────────────────────────────────────

func TestBusinessObjectDataSource_Init_NilPointer(t *testing.T) {
	// Pass a typed nil pointer — Init should detect IsNil and return empty.
	var p *product
	ds := data.NewBusinessObjectDataSource("NilPtr", p)
	if err := ds.Init(); err != nil {
		t.Fatalf("Init(nil pointer): %v", err)
	}
	if ds.RowCount() != 0 {
		t.Errorf("RowCount = %d, want 0 for nil pointer", ds.RowCount())
	}
}

// ── Init with single struct value (not a slice) ───────────────────────────────

func TestBusinessObjectDataSource_Init_SingleStruct(t *testing.T) {
	// A single struct (not slice) is treated as a one-row source.
	single := product{ID: 7, Name: "Single", Price: 9.99}
	ds := data.NewBusinessObjectDataSource("Single", single)
	if err := ds.Init(); err != nil {
		t.Fatalf("Init(single struct): %v", err)
	}
	if ds.RowCount() != 1 {
		t.Errorf("RowCount = %d, want 1 for single struct", ds.RowCount())
	}
	if err := ds.First(); err != nil {
		t.Fatalf("First: %v", err)
	}
	v, err := ds.GetValue("Name")
	if err != nil {
		t.Fatalf("GetValue(Name): %v", err)
	}
	if v != "Single" {
		t.Errorf("GetValue(Name) = %v, want Single", v)
	}
}

// ── GetValue when EOF ─────────────────────────────────────────────────────────

func TestBusinessObjectDataSource_GetValue_EOF(t *testing.T) {
	ds := data.NewBusinessObjectDataSource("P", testProducts)
	_ = ds.Init()
	_ = ds.First()
	// Exhaust the source.
	for !ds.EOF() {
		_ = ds.Next()
	}
	_, err := ds.GetValue("Name")
	if err == nil {
		t.Error("GetValue when EOF should return error")
	}
}

// ── GetValue when rows is empty ───────────────────────────────────────────────

func TestBusinessObjectDataSource_GetValue_EmptyRows(t *testing.T) {
	ds := data.NewBusinessObjectDataSource("Empty", []product{})
	_ = ds.Init()
	_, err := ds.GetValue("Name")
	if err == nil {
		t.Error("GetValue with empty rows should return error")
	}
}

// ── columnsFor and fieldValue with unexported fields ─────────────────────────

type mixedFields struct {
	Public   string
	hidden   string //nolint:unused // unexported field for test
	AnotherPublic int
}

func TestBusinessObjectDataSource_UnexportedFields_Filtered(t *testing.T) {
	// columnsFor should skip unexported fields.
	rows := []mixedFields{
		{Public: "hello", AnotherPublic: 42},
	}
	ds := data.NewBusinessObjectDataSource("Mixed", rows)
	if err := ds.Init(); err != nil {
		t.Fatalf("Init: %v", err)
	}
	// Only Public and AnotherPublic should appear (hidden is unexported).
	cols := ds.Columns()
	for _, c := range cols {
		if c.Name == "hidden" {
			t.Error("unexported field 'hidden' should not appear in columns")
		}
	}
	if len(cols) != 2 {
		t.Errorf("Columns = %d, want 2 (Public + AnotherPublic)", len(cols))
	}
	if err := ds.First(); err != nil {
		t.Fatalf("First: %v", err)
	}
	// GetValue for unexported name should fail.
	_, err := ds.GetValue("hidden")
	if err == nil {
		t.Error("GetValue for unexported field should return error")
	}
	// GetValue for exported fields should succeed.
	v, err := ds.GetValue("Public")
	if err != nil {
		t.Fatalf("GetValue(Public): %v", err)
	}
	if v != "hello" {
		t.Errorf("GetValue(Public) = %v, want hello", v)
	}
}

// ── columnsFor with nil pointer row ──────────────────────────────────────────

func TestBusinessObjectDataSource_NilPointerSlice(t *testing.T) {
	// A slice of nil pointers → columnsFor gets a nil pointer as first row.
	var p1 *product // nil
	rows := []*product{p1}
	ds := data.NewBusinessObjectDataSource("NilSlice", rows)
	if err := ds.Init(); err != nil {
		t.Fatalf("Init: %v", err)
	}
	// columnsFor on a nil pointer returns nil (no columns).
	cols := ds.Columns()
	if cols != nil {
		t.Errorf("Columns for nil-pointer row should be nil, got %v", cols)
	}
	if err := ds.First(); err != nil {
		t.Fatalf("First: %v", err)
	}
	// fieldValue on nil pointer returns nil, nil.
	v, err := ds.GetValue("anything")
	if err != nil {
		t.Logf("GetValue on nil pointer: %v (acceptable)", err)
	} else if v != nil {
		t.Logf("GetValue on nil pointer: %v", v)
	}
}

// ── fieldValue: default case (non-struct, non-map primitive row) ──────────────

func TestBusinessObjectDataSource_GetValue_SinglePrimitive(t *testing.T) {
	// A single int row — fieldValue returns the int regardless of column name.
	ds := data.NewBusinessObjectDataSource("Num", 42)
	if err := ds.Init(); err != nil {
		t.Fatalf("Init: %v", err)
	}
	if err := ds.First(); err != nil {
		t.Fatalf("First: %v", err)
	}
	v, err := ds.GetValue("whatever")
	if err != nil {
		t.Fatalf("GetValue for single primitive: %v", err)
	}
	if v != 42 {
		t.Errorf("GetValue = %v, want 42", v)
	}
}

// ── connection.go: Open with bad driver name (error path) ─────────────────────

func TestDataConnectionBase_Open_InvalidDriver(t *testing.T) {
	c := data.NewDataConnectionBase("no_such_driver_xyz")
	err := c.Open()
	// sql.Open itself rarely errors (it's lazy), but Ping would.
	// The error at Open depends on driver registration.
	// We just verify it either errors or returns without panic.
	_ = err
}

// ── TableDataSource.Init with TableName fallback ──────────────────────────────

func TestTableDataSource_Init_TableNameFallback(t *testing.T) {
	c := data.NewDataConnectionBase("stub")
	c.ConnectionString = "stub://test"
	ts := c.CreateTable("users")
	ts.SetTableName("users") // use TableName, not SelectCommand.
	if err := ts.Init(); err != nil {
		t.Fatalf("Init with TableName: %v", err)
	}
	if ts.RowCount() != 2 {
		t.Errorf("RowCount = %d, want 2", ts.RowCount())
	}
}
