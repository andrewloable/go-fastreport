package data_test

import (
	"database/sql"
	"database/sql/driver"
	"io"
	"testing"

	"github.com/andrewloable/go-fastreport/data"
)

// -----------------------------------------------------------------------
// In-process stub SQL driver
// -----------------------------------------------------------------------

func init() {
	sql.Register("stub", &stubDriver{})
}

type stubDriver struct{}

func (d *stubDriver) Open(name string) (driver.Conn, error) {
	return &stubConn{}, nil
}

type stubConn struct{}

func (c *stubConn) Prepare(query string) (driver.Stmt, error) {
	return &stubStmt{query: query}, nil
}
func (c *stubConn) Close() error                       { return nil }
func (c *stubConn) Begin() (driver.Tx, error)          { return &stubTx{}, nil }

type stubTx struct{}

func (t *stubTx) Commit() error   { return nil }
func (t *stubTx) Rollback() error { return nil }

type stubStmt struct{ query string }

func (s *stubStmt) Close() error                               { return nil }
func (s *stubStmt) NumInput() int                              { return -1 }
func (s *stubStmt) Exec(args []driver.Value) (driver.Result, error) {
	return nil, nil
}
func (s *stubStmt) Query(args []driver.Value) (driver.Rows, error) {
	return &stubRows{pos: -1}, nil
}

type stubRows struct{ pos int }

var stubColumns = []string{"id", "name"}
var stubData = [][]driver.Value{
	{int64(1), "Alice"},
	{int64(2), "Bob"},
}

func (r *stubRows) Columns() []string { return stubColumns }
func (r *stubRows) Close() error      { return nil }
func (r *stubRows) Next(dest []driver.Value) error {
	r.pos++
	if r.pos >= len(stubData) {
		return io.EOF
	}
	copy(dest, stubData[r.pos])
	return nil
}

// -----------------------------------------------------------------------
// CommandParameter tests
// -----------------------------------------------------------------------

func TestNewCommandParameter_Defaults(t *testing.T) {
	p := data.NewCommandParameter("@id")
	if p.Name != "@id" {
		t.Errorf("Name = %q, want @id", p.Name)
	}
	if p.Direction != data.ParamDirectionInput {
		t.Errorf("Direction default = %d, want Input", p.Direction)
	}
	if p.Value != nil {
		t.Error("Value should default to nil")
	}
}

func TestCommandParameter_Fields(t *testing.T) {
	p := data.NewCommandParameter("@name")
	p.DataType = "varchar"
	p.Size = 100
	p.Expression = "[CustomerName]"
	p.DefaultValue = "Unknown"
	p.Value = "Alice"
	p.Direction = data.ParamDirectionOutput

	if p.DataType != "varchar" {
		t.Errorf("DataType = %q", p.DataType)
	}
	if p.Size != 100 {
		t.Errorf("Size = %d, want 100", p.Size)
	}
	if p.Expression != "[CustomerName]" {
		t.Errorf("Expression = %q", p.Expression)
	}
	if p.Direction != data.ParamDirectionOutput {
		t.Errorf("Direction = %d, want Output", p.Direction)
	}
	if p.Value != "Alice" {
		t.Errorf("Value = %v, want Alice", p.Value)
	}
}

func TestParameterDirection_Constants(t *testing.T) {
	dirs := []data.ParameterDirection{
		data.ParamDirectionInput,
		data.ParamDirectionOutput,
		data.ParamDirectionInputOutput,
		data.ParamDirectionReturnValue,
	}
	seen := map[data.ParameterDirection]bool{}
	for _, d := range dirs {
		if seen[d] {
			t.Errorf("duplicate ParameterDirection %d", d)
		}
		seen[d] = true
	}
}

// -----------------------------------------------------------------------
// DataConnectionBase tests
// -----------------------------------------------------------------------

func TestNewDataConnectionBase_Fields(t *testing.T) {
	c := data.NewDataConnectionBase("stub")
	if c.DriverName() != "stub" {
		t.Errorf("DriverName = %q, want stub", c.DriverName())
	}
	if c.DB() != nil {
		t.Error("DB should be nil before Open")
	}
	// CommandTimeout default is 30 to match C# DataConnectionBase constructor.
	// C# ref: FastReport.Data.DataConnectionBase constructor — commandTimeout = 30
	if c.CommandTimeout != 30 {
		t.Errorf("CommandTimeout default = %d, want 30", c.CommandTimeout)
	}
}

func TestDataConnectionBase_Open_Close(t *testing.T) {
	c := data.NewDataConnectionBase("stub")
	c.ConnectionString = "stub://test"
	if err := c.Open(); err != nil {
		t.Fatalf("Open error: %v", err)
	}
	if c.DB() == nil {
		t.Error("DB should be non-nil after Open")
	}
	if err := c.Close(); err != nil {
		t.Fatalf("Close error: %v", err)
	}
	if c.DB() != nil {
		t.Error("DB should be nil after Close")
	}
}

func TestDataConnectionBase_Open_Idempotent(t *testing.T) {
	c := data.NewDataConnectionBase("stub")
	c.ConnectionString = "stub://test"
	_ = c.Open()
	db1 := c.DB()
	_ = c.Open() // second open is no-op
	if c.DB() != db1 {
		t.Error("second Open should return the same *sql.DB")
	}
	_ = c.Close()
}

func TestDataConnectionBase_Close_WhenNil(t *testing.T) {
	c := data.NewDataConnectionBase("stub")
	if err := c.Close(); err != nil {
		t.Errorf("Close on nil DB should be no-op, got: %v", err)
	}
}

func TestDataConnectionBase_Tables(t *testing.T) {
	c := data.NewDataConnectionBase("stub")
	tbl := c.CreateTable("Users")
	if len(c.Tables()) != 1 {
		t.Errorf("Tables len = %d, want 1", len(c.Tables()))
	}
	if tbl.Connection() != c {
		t.Error("table.Connection should point back to the connection")
	}
}

func TestDataConnectionBase_AddTable(t *testing.T) {
	c := data.NewDataConnectionBase("stub")
	t1 := data.NewTableDataSource("A")
	t2 := data.NewTableDataSource("B")
	c.AddTable(t1)
	c.AddTable(t2)
	if len(c.Tables()) != 2 {
		t.Errorf("Tables len = %d, want 2", len(c.Tables()))
	}
}

// -----------------------------------------------------------------------
// TableDataSource tests
// -----------------------------------------------------------------------

func TestNewTableDataSource_Defaults(t *testing.T) {
	ts := data.NewTableDataSource("Orders")
	if ts.Name() != "Orders" {
		t.Errorf("Name = %q, want Orders", ts.Name())
	}
	if ts.TableName() != "" {
		t.Errorf("TableName default = %q, want empty", ts.TableName())
	}
	if ts.SelectCommand() != "" {
		t.Errorf("SelectCommand default = %q, want empty", ts.SelectCommand())
	}
	if ts.StoreData() {
		t.Error("StoreData should default to false")
	}
	if ts.Connection() != nil {
		t.Error("Connection should default to nil")
	}
}

func TestTableDataSource_SetFields(t *testing.T) {
	ts := data.NewTableDataSource("T")
	ts.SetTableName("customers")
	ts.SetSelectCommand("SELECT * FROM customers WHERE active=1")
	ts.SetStoreData(true)

	if ts.TableName() != "customers" {
		t.Errorf("TableName = %q", ts.TableName())
	}
	if ts.SelectCommand() != "SELECT * FROM customers WHERE active=1" {
		t.Errorf("SelectCommand = %q", ts.SelectCommand())
	}
	if !ts.StoreData() {
		t.Error("StoreData should be true")
	}
}

func TestTableDataSource_AddParameter(t *testing.T) {
	ts := data.NewTableDataSource("T")
	p := data.NewCommandParameter("@id")
	p.Value = 42
	ts.AddParameter(p)
	if len(ts.Parameters()) != 1 {
		t.Errorf("Parameters len = %d, want 1", len(ts.Parameters()))
	}
	if ts.Parameters()[0].Value != 42 {
		t.Errorf("Parameters[0].Value = %v, want 42", ts.Parameters()[0].Value)
	}
}

func TestTableDataSource_Init_NoConnection(t *testing.T) {
	ts := data.NewTableDataSource("T")
	err := ts.Init()
	if err == nil {
		t.Error("expected error when no connection set")
	}
}

func TestTableDataSource_Init_WithStubDB(t *testing.T) {
	c := data.NewDataConnectionBase("stub")
	c.ConnectionString = "stub://test"
	ts := c.CreateTable("users")
	ts.SetSelectCommand("SELECT id, name FROM users")

	if err := ts.Init(); err != nil {
		t.Fatalf("Init error: %v", err)
	}
	if ts.RowCount() != 2 {
		t.Errorf("RowCount = %d, want 2", ts.RowCount())
	}
	if len(ts.Columns()) != 2 {
		t.Fatalf("Columns len = %d, want 2", len(ts.Columns()))
	}

	_ = ts.First()
	v, err := ts.GetValue("name")
	if err != nil {
		t.Fatalf("GetValue error: %v", err)
	}
	if v != "Alice" {
		t.Errorf("GetValue(name) = %v, want Alice", v)
	}

	_ = ts.Next()
	v2, _ := ts.GetValue("name")
	if v2 != "Bob" {
		t.Errorf("GetValue(name) row2 = %v, want Bob", v2)
	}

	_ = c.Close()
}

func TestTableDataSource_Init_NoSelectCommand_NoTableName(t *testing.T) {
	c := data.NewDataConnectionBase("stub")
	c.ConnectionString = "stub://test"
	ts := c.CreateTable("empty")
	// Neither SelectCommand nor TableName set
	err := ts.Init()
	if err == nil {
		t.Error("expected error when no query or table name")
	}
}
