package data_test

// coverage_gaps_test.go — targeted tests to cover remaining uncovered branches:
//   - business.go Init: non-nil pointer dereference (rv = rv.Elem())
//   - connection.go Init: Open error, parameter iteration, Query error, Scan error

import (
	"database/sql"
	"database/sql/driver"
	"errors"
	"testing"

	"github.com/andrewloable/go-fastreport/data"
)

// ── business.go Init: non-nil pointer to struct ───────────────────────────────
//
// Passing a *product (non-nil pointer) exercises the `rv = rv.Elem()` branch
// inside the `for rv.Kind() == reflect.Ptr` loop.

type bizProduct struct {
	ID    int
	Name  string
	Price float64
}

func TestBusinessObjectDataSource_Init_NonNilPointerToStruct(t *testing.T) {
	p := &bizProduct{ID: 1, Name: "Widget", Price: 9.99}
	ds := data.NewBusinessObjectDataSource("PtrStruct", p)
	if err := ds.Init(); err != nil {
		t.Fatalf("Init(*struct): %v", err)
	}
	// A single struct pointer → treated as a one-row source.
	if ds.RowCount() != 1 {
		t.Errorf("RowCount = %d, want 1 for pointer-to-struct", ds.RowCount())
	}
	if err := ds.First(); err != nil {
		t.Fatalf("First: %v", err)
	}
	v, err := ds.GetValue("Name")
	if err != nil {
		t.Fatalf("GetValue(Name): %v", err)
	}
	if v != "Widget" {
		t.Errorf("GetValue(Name) = %v, want Widget", v)
	}
}

func TestBusinessObjectDataSource_Init_NonNilPointerToSlice(t *testing.T) {
	rows := []bizProduct{
		{ID: 1, Name: "Alpha", Price: 1.0},
		{ID: 2, Name: "Beta", Price: 2.0},
	}
	ds := data.NewBusinessObjectDataSource("PtrSlice", &rows)
	if err := ds.Init(); err != nil {
		t.Fatalf("Init(*[]struct): %v", err)
	}
	if ds.RowCount() != 2 {
		t.Errorf("RowCount = %d, want 2 for pointer-to-slice", ds.RowCount())
	}
	if err := ds.First(); err != nil {
		t.Fatalf("First: %v", err)
	}
	v, err := ds.GetValue("Name")
	if err != nil {
		t.Fatalf("GetValue(Name): %v", err)
	}
	if v != "Alpha" {
		t.Errorf("GetValue(Name) = %v, want Alpha", v)
	}
}

// ── connection.go Init: Open() error path ─────────────────────────────────────
//
// Register a driver whose Open() always returns an error.
// TableDataSource.Init calls connection.Open() when db is nil — this error
// should be propagated back to the caller.

type openFailDriver struct{}

func (d *openFailDriver) Open(name string) (driver.Conn, error) {
	return nil, errors.New("openFailDriver: simulated Open failure")
}

func init() {
	sql.Register("stub-open-fail", &openFailDriver{})
}

func TestTableDataSource_Init_OpenError(t *testing.T) {
	c := data.NewDataConnectionBase("stub-open-fail")
	c.ConnectionString = "stub-open-fail://test"
	ts := c.CreateTable("t")
	ts.SetSelectCommand("SELECT id FROM t")

	err := ts.Init()
	if err == nil {
		t.Error("Init should return error when connection.Open() fails")
	}
}

// ── connection.go Init: parameter iteration (line 243-245) ───────────────────
//
// When parameters are set on a TableDataSource, Init iterates them to build
// the args slice. The "stub" driver (registered in connection_test.go) accepts
// any query, so this exercises the parameter loop without needing a special driver.

func TestTableDataSource_Init_WithParameter(t *testing.T) {
	c := data.NewDataConnectionBase("stub")
	c.ConnectionString = "stub://test"
	ts := c.CreateTable("p")
	ts.SetSelectCommand("SELECT id, name FROM users WHERE id = ?")

	// Add a parameter — this exercises the `for i, p := range t.parameters` loop.
	param := data.NewCommandParameter("@id")
	param.Value = int64(1)
	ts.AddParameter(param)

	if err := ts.Init(); err != nil {
		t.Fatalf("Init with parameter: %v", err)
	}
	// The stub driver ignores query args and returns 2 rows.
	if ts.RowCount() != 2 {
		t.Errorf("RowCount = %d, want 2", ts.RowCount())
	}
	_ = c.Close()
}

// ── connection.go Init: Query error path (line 248-250) ──────────────────────
//
// Register a driver whose stmt.Query always errors.

type queryFailDriver struct{}
type queryFailConn struct{}
type queryFailTx struct{}
type queryFailStmt struct{}

func (d *queryFailDriver) Open(name string) (driver.Conn, error) { return &queryFailConn{}, nil }
func (c *queryFailConn) Prepare(q string) (driver.Stmt, error)   { return &queryFailStmt{}, nil }
func (c *queryFailConn) Close() error                             { return nil }
func (c *queryFailConn) Begin() (driver.Tx, error)               { return &queryFailTx{}, nil }
func (t *queryFailTx) Commit() error                              { return nil }
func (t *queryFailTx) Rollback() error                            { return nil }
func (s *queryFailStmt) Close() error                             { return nil }
func (s *queryFailStmt) NumInput() int                            { return -1 }
func (s *queryFailStmt) Exec(args []driver.Value) (driver.Result, error) {
	return nil, nil
}
func (s *queryFailStmt) Query(args []driver.Value) (driver.Rows, error) {
	return nil, errors.New("queryFailStmt: simulated Query failure")
}

func init() {
	sql.Register("stub-query-fail", &queryFailDriver{})
}

func TestTableDataSource_Init_QueryError(t *testing.T) {
	c := data.NewDataConnectionBase("stub-query-fail")
	c.ConnectionString = "stub-query-fail://test"
	ts := c.CreateTable("t")
	ts.SetSelectCommand("SELECT id FROM t")

	err := ts.Init()
	if err == nil {
		t.Error("Init should return error when DB.Query() fails")
	}
	_ = c.Close()
}


// TestDataComponentBase_InitializeComponent ensures the InitializeComponent
// method is exercised (it has an empty body but the coverage tool still
// reports 0% if it is never called).
func TestDataComponentBase_InitializeComponent(t *testing.T) {
	d := data.NewDataComponentBase("test")
	// InitializeComponent is a no-op hook for subclasses.
	d.InitializeComponent() // must not panic
}
