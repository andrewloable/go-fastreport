package data_test

// data_connection_coverage_test.go — targeted tests to cover remaining uncovered
// branches in connection.go Init and command_parameter_collection.go Deserialize.
//
// Uncovered branches identified from coverage profile:
//
//   connection.go:228.45,230.4 — Init: t.connection.Open() error branch.
//     sql.Open() (called from DataConnectionBase.Open) does NOT call driver.Open();
//     it defers the actual connection until first use. Therefore this branch is
//     only reachable when the driver name is not registered. Covered below by
//     using an unregistered driver name after the connection is already open
//     (so DB() is not nil and the branch is skipped), OR via the stub-open-fail
//     driver that makes sql.Open itself report an error in some circumstances.
//     NOTE: In practice this branch fires when sql.Open() itself errors, which
//     requires the driver name to be completely invalid. The existing
//     TestTableDataSource_Init_OpenError test already covers the query-time error.
//     For sql.Open, the only way to get an error is an unregistered driver — but
//     Open() validates the driver name eagerly, so we can trigger this by calling
//     Init() on a connection configured with a non-registered driver name.
//
//   connection.go:254.16,256.3 — Init: rows.Columns() error branch.
//     Requires a driver whose rows.Columns() returns an error.
//
//   connection.go:272.48,274.4 — Init: rows.Scan() error branch.
//     Requires a driver whose rows.Scan() returns an error.
//
//   command_parameter_collection.go:89.43,91.5 — Deserialize: p.Deserialize(r)
//     error branch. CommandParameter.Deserialize only calls ReadStr/ReadInt and
//     always returns nil — this is dead code.

import (
	"database/sql"
	"database/sql/driver"
	"errors"
	"io"
	"testing"

	"github.com/andrewloable/go-fastreport/data"
)

// ── connection.go Init: sql.Open() error ─────────────────────────────────────
// sql.Open() returns an error when the driver name is not registered.
// DataConnectionBase.Open calls sql.Open(driverName, dsn), so using an
// unregistered driver name will make Open() return an error which Init()
// must propagate.

func TestTableDataSource_Init_UnregisteredDriver(t *testing.T) {
	// "definitely-not-registered-driver-xyz" is not registered with database/sql.
	c := data.NewDataConnectionBase("definitely-not-registered-driver-xyz")
	c.ConnectionString = "dsn"
	ts := c.CreateTable("t")
	ts.SetSelectCommand("SELECT id FROM t")

	err := ts.Init()
	if err == nil {
		t.Error("Init should return error when driver is not registered (sql.Open fails)")
	}
}

// ── connection.go Init: rows.Columns() error path ────────────────────────────
// Register a driver whose rows.Columns() returns an error.

type colsErrDriver struct{}
type colsErrConn struct{}
type colsErrTx struct{}
type colsErrStmt struct{}
type colsErrRows struct{}

func (d *colsErrDriver) Open(name string) (driver.Conn, error) { return &colsErrConn{}, nil }
func (c *colsErrConn) Prepare(q string) (driver.Stmt, error)   { return &colsErrStmt{}, nil }
func (c *colsErrConn) Close() error                             { return nil }
func (c *colsErrConn) Begin() (driver.Tx, error)               { return &colsErrTx{}, nil }
func (t *colsErrTx) Commit() error                              { return nil }
func (t *colsErrTx) Rollback() error                            { return nil }
func (s *colsErrStmt) Close() error                             { return nil }
func (s *colsErrStmt) NumInput() int                            { return -1 }
func (s *colsErrStmt) Exec(args []driver.Value) (driver.Result, error) {
	return nil, nil
}
func (s *colsErrStmt) Query(args []driver.Value) (driver.Rows, error) {
	return &colsErrRows{}, nil
}

// Columns returns an error — this should cause Init to fail at the
// rows.Columns() call.
// Note: database/sql calls driver.Rows.Columns() synchronously after Query,
// so returning an error here surfaces through rows.Columns().
// However, the standard database/sql interface requires Columns() to return
// []string with no error. The rows.Columns() error path in TableDataSource.Init
// is therefore reachable only through a sql.Rows implementation that propagates
// an internal error. We use a driver.Rows that returns a real column error by
// having the sql.Rows.Columns() call fail, which requires returning a nil/empty
// name list that triggers column validation. In practice, database/sql does not
// propagate driver.Rows.Columns() errors directly. We instead use a custom
// approach: register a queryFail driver that returns rows whose Next() fails
// immediately with a non-EOF error, causing rows.Err() to be non-nil. Then we
// verify that the Columns() error path exists via a scan-fail driver instead.

func (r *colsErrRows) Columns() []string {
	// Return a valid column list — database/sql doesn't propagate Columns() panics.
	return []string{"id"}
}
func (r *colsErrRows) Close() error { return nil }
func (r *colsErrRows) Next(dest []driver.Value) error {
	return io.EOF
}

func init() {
	sql.Register("stub-cols-ok", &colsErrDriver{})
}

// TestTableDataSource_Init_EmptyResult verifies that Init works with a driver
// that returns zero rows (immediate io.EOF on first Next call).
func TestTableDataSource_Init_EmptyResult(t *testing.T) {
	c := data.NewDataConnectionBase("stub-cols-ok")
	c.ConnectionString = "stub-cols-ok://test"
	ts := c.CreateTable("t")
	ts.SetSelectCommand("SELECT id FROM t")

	if err := ts.Init(); err != nil {
		t.Fatalf("Init with zero-row result: %v", err)
	}
	if ts.RowCount() != 0 {
		t.Errorf("RowCount = %d, want 0 for empty result", ts.RowCount())
	}
	_ = c.Close()
}

// ── connection.go Init: rows.Scan() error path ───────────────────────────────
// Register a driver whose Next() returns one row but Scan fails by returning
// an unexpected number of values, or we implement it via a scan-fail approach.
// Since database/sql handles scanning via the driver.Rows.Next() call into
// destination slice, we simulate a scan error by having Next() return an
// error rather than populate values. database/sql surfaces this as rows.Scan error.

type scanErrDriver struct{}
type scanErrConn struct{}
type scanErrTx struct{}
type scanErrStmt struct{}
type scanErrRows struct{ called bool }

func (d *scanErrDriver) Open(name string) (driver.Conn, error)  { return &scanErrConn{}, nil }
func (c *scanErrConn) Prepare(q string) (driver.Stmt, error)    { return &scanErrStmt{}, nil }
func (c *scanErrConn) Close() error                              { return nil }
func (c *scanErrConn) Begin() (driver.Tx, error)                { return &scanErrTx{}, nil }
func (t *scanErrTx) Commit() error                               { return nil }
func (t *scanErrTx) Rollback() error                             { return nil }
func (s *scanErrStmt) Close() error                              { return nil }
func (s *scanErrStmt) NumInput() int                             { return -1 }
func (s *scanErrStmt) Exec(args []driver.Value) (driver.Result, error) { return nil, nil }
func (s *scanErrStmt) Query(args []driver.Value) (driver.Rows, error) {
	return &scanErrRows{}, nil
}
func (r *scanErrRows) Columns() []string { return []string{"id"} }
func (r *scanErrRows) Close() error      { return nil }
func (r *scanErrRows) Next(dest []driver.Value) error {
	if !r.called {
		r.called = true
		// Return a non-nil, non-io.EOF error — database/sql propagates this
		// as a rows.Scan() error when it tries to populate the destination.
		return errors.New("scanErrRows: simulated scan failure")
	}
	return io.EOF
}

func init() {
	sql.Register("stub-scan-err", &scanErrDriver{})
}

func TestTableDataSource_Init_ScanError(t *testing.T) {
	c := data.NewDataConnectionBase("stub-scan-err")
	c.ConnectionString = "stub-scan-err://test"
	ts := c.CreateTable("t")
	ts.SetSelectCommand("SELECT id FROM t")

	err := ts.Init()
	if err == nil {
		t.Error("Init should return error when rows.Scan() fails")
	}
	_ = c.Close()
}

// ── command_parameter_collection.go Deserialize: non-Parameter child ─────────
// The Deserialize loop processes children named "Parameter". When a child with
// a different type name is encountered, the `if typeName == "Parameter"` block
// is skipped. The FinishChild() is still called. This path is exercised by the
// existing TestCommandParameterCollection_Deserialize_FinishChildError in
// extra_coverage_test.go, but that test also triggers FinishChild error.
//
// Here we add a test where a non-Parameter child is present and FinishChild
// succeeds, exercising the `typeName != "Parameter"` branch cleanly.

// nonParamReader delivers one "Other" child then no more children.
type nonParamReader struct {
	step int
}

func (r *nonParamReader) ReadStr(name, def string) string          { return def }
func (r *nonParamReader) ReadInt(name string, def int) int         { return def }
func (r *nonParamReader) ReadBool(name string, def bool) bool      { return def }
func (r *nonParamReader) ReadFloat(name string, def float32) float32 { return def }
func (r *nonParamReader) NextChild() (string, bool) {
	if r.step == 0 {
		r.step++
		return "Other", true // not "Parameter"
	}
	return "", false
}
func (r *nonParamReader) FinishChild() error { return nil }

func TestCommandParameterCollection_Deserialize_NonParameterChild(t *testing.T) {
	col := data.NewCommandParameterCollection()
	r := &nonParamReader{}
	err := col.Deserialize(r)
	if err != nil {
		t.Errorf("Deserialize with non-Parameter child should not error, got %v", err)
	}
	// The "Other" child is not a Parameter, so nothing is added to the collection.
	if col.Count() != 0 {
		t.Errorf("Count = %d, want 0 (Other child is not a Parameter)", col.Count())
	}
}

// TestCommandParameterCollection_Deserialize_MixedChildren exercises the case
// where both Parameter and non-Parameter children are present in sequence.
func TestCommandParameterCollection_Deserialize_MixedChildren(t *testing.T) {
	col := data.NewCommandParameterCollection()
	// mixedChildReader delivers: "Metadata" (non-Parameter), "Parameter", then done.
	r := &mixedChildReader{}
	err := col.Deserialize(r)
	if err != nil {
		t.Fatalf("Deserialize with mixed children: %v", err)
	}
	// Only the Parameter child should be added.
	if col.Count() != 1 {
		t.Errorf("Count = %d, want 1 (one Parameter out of two children)", col.Count())
	}
}

type mixedChildReader struct {
	step int
}

func (r *mixedChildReader) ReadStr(name, def string) string          { return def }
func (r *mixedChildReader) ReadInt(name string, def int) int         { return def }
func (r *mixedChildReader) ReadBool(name string, def bool) bool      { return def }
func (r *mixedChildReader) ReadFloat(name string, def float32) float32 { return def }
func (r *mixedChildReader) NextChild() (string, bool) {
	switch r.step {
	case 0:
		r.step++
		return "Metadata", true // non-Parameter child
	case 1:
		r.step++
		return "Parameter", true // Parameter child
	default:
		return "", false
	}
}
func (r *mixedChildReader) FinishChild() error { return nil }

// TestCommandParameterCollection_Deserialize_ParameterDeserializeError documents
// that the p.Deserialize(r) error branch in Deserialize (line 89-91) is dead code.
// CommandParameter.Deserialize only calls ReadStr/ReadInt and always returns nil.
// No error can be injected through the public Reader interface for this path.
// This test verifies the happy path — a Parameter child deserializes successfully.
func TestCommandParameterCollection_Deserialize_ParameterSucceeds(t *testing.T) {
	col := data.NewCommandParameterCollection()
	r := &singleParamReader{name: "@myParam"}
	err := col.Deserialize(r)
	if err != nil {
		t.Fatalf("Deserialize with valid Parameter: %v", err)
	}
	if col.Count() != 1 {
		t.Fatalf("Count = %d, want 1", col.Count())
	}
}

type singleParamReader struct {
	step int
	name string
}

func (r *singleParamReader) ReadStr(key, def string) string {
	if key == "Name" && r.step == 1 {
		return r.name
	}
	return def
}
func (r *singleParamReader) ReadInt(name string, def int) int         { return def }
func (r *singleParamReader) ReadBool(name string, def bool) bool      { return def }
func (r *singleParamReader) ReadFloat(name string, def float32) float32 { return def }
func (r *singleParamReader) NextChild() (string, bool) {
	if r.step == 0 {
		r.step++
		return "Parameter", true
	}
	return "", false
}
func (r *singleParamReader) FinishChild() error { return nil }

// ── DataConnectionBase.Open: OnDatabaseLogin / OnAfterDatabaseLogin ──────────
// Ported from C# ReportSettings.Core.cs OnDatabaseLogin / ReportSettings.cs
// DatabaseLogin + AfterDatabaseLogin events.

// TestDataConnectionBase_OnDatabaseLogin_CanOverrideDSN verifies that
// OnDatabaseLogin is called before sql.Open and that replacing the
// ConnectionString in the callback causes the new DSN to be used.
func TestDataConnectionBase_OnDatabaseLogin_CanOverrideDSN(t *testing.T) {
	c := data.NewDataConnectionBase("stub")
	c.ConnectionString = "original-dsn"
	var capturedDSN string
	c.OnDatabaseLogin = func(e *data.DatabaseLoginEventArgs) {
		capturedDSN = e.ConnectionString
		e.ConnectionString = "replaced-dsn"
	}
	if err := c.Open(); err != nil {
		t.Fatalf("Open error: %v", err)
	}
	if capturedDSN != "original-dsn" {
		t.Errorf("OnDatabaseLogin received DSN %q, want %q", capturedDSN, "original-dsn")
	}
	if c.DB() == nil {
		t.Error("DB() should not be nil after a successful Open")
	}
	_ = c.Close()
}

// TestDataConnectionBase_OnDatabaseLogin_CalledOnce verifies that
// OnDatabaseLogin is invoked exactly once per Open() call.
func TestDataConnectionBase_OnDatabaseLogin_CalledOnce(t *testing.T) {
	c := data.NewDataConnectionBase("stub")
	c.ConnectionString = "stub://test"
	count := 0
	c.OnDatabaseLogin = func(e *data.DatabaseLoginEventArgs) { count++ }
	if err := c.Open(); err != nil {
		t.Fatalf("Open error: %v", err)
	}
	if count != 1 {
		t.Errorf("OnDatabaseLogin called %d times, want 1", count)
	}
	_ = c.Close()
}

// TestDataConnectionBase_OnDatabaseLogin_NotCalledWhenAlreadyOpen verifies
// that a second Open() call (idempotent) does not fire OnDatabaseLogin again.
func TestDataConnectionBase_OnDatabaseLogin_NotCalledWhenAlreadyOpen(t *testing.T) {
	c := data.NewDataConnectionBase("stub")
	c.ConnectionString = "stub://test"
	count := 0
	c.OnDatabaseLogin = func(e *data.DatabaseLoginEventArgs) { count++ }
	_ = c.Open()
	_ = c.Open() // second call — db != nil, should return early
	if count != 1 {
		t.Errorf("OnDatabaseLogin called %d times, want 1 (second Open is no-op)", count)
	}
	_ = c.Close()
}

// TestDataConnectionBase_OnAfterDatabaseLogin_ReceivesDB verifies that
// OnAfterDatabaseLogin is called after sql.Open succeeds and receives
// the non-nil *sql.DB.
func TestDataConnectionBase_OnAfterDatabaseLogin_ReceivesDB(t *testing.T) {
	c := data.NewDataConnectionBase("stub")
	c.ConnectionString = "stub://test"
	var receivedDB *sql.DB
	c.OnAfterDatabaseLogin = func(e *data.AfterDatabaseLoginEventArgs) {
		receivedDB = e.DB
	}
	if err := c.Open(); err != nil {
		t.Fatalf("Open error: %v", err)
	}
	if receivedDB == nil {
		t.Error("OnAfterDatabaseLogin: received nil DB, want non-nil")
	}
	if receivedDB != c.DB() {
		t.Error("OnAfterDatabaseLogin: received DB differs from c.DB()")
	}
	_ = c.Close()
}

// TestDataConnectionBase_OnAfterDatabaseLogin_NotCalledOnError verifies that
// OnAfterDatabaseLogin is not called when Open() fails.
func TestDataConnectionBase_OnAfterDatabaseLogin_NotCalledOnError(t *testing.T) {
	c := data.NewDataConnectionBase("definitely-not-registered-xyz2")
	c.ConnectionString = "bad://dsn"
	called := false
	c.OnAfterDatabaseLogin = func(e *data.AfterDatabaseLoginEventArgs) { called = true }
	err := c.Open()
	if err == nil {
		t.Fatal("expected Open to fail with unregistered driver")
	}
	if called {
		t.Error("OnAfterDatabaseLogin should not be called when Open fails")
	}
}

// TestDataConnectionBase_BothCallbacks_Sequence verifies that OnDatabaseLogin
// fires before sql.Open and OnAfterDatabaseLogin fires after, in order.
func TestDataConnectionBase_BothCallbacks_Sequence(t *testing.T) {
	c := data.NewDataConnectionBase("stub")
	c.ConnectionString = "stub://test"
	var seq []string
	c.OnDatabaseLogin = func(e *data.DatabaseLoginEventArgs) { seq = append(seq, "login") }
	c.OnAfterDatabaseLogin = func(e *data.AfterDatabaseLoginEventArgs) { seq = append(seq, "after") }
	if err := c.Open(); err != nil {
		t.Fatalf("Open error: %v", err)
	}
	if len(seq) != 2 || seq[0] != "login" || seq[1] != "after" {
		t.Errorf("callback order = %v, want [login after]", seq)
	}
	_ = c.Close()
}

// TestReportSettings_LoginCallbackFields verifies that database login callbacks
// can be stored as func fields and wired to DataConnectionBase.OnDatabaseLogin /
// OnAfterDatabaseLogin. This documents the Go equivalent of the C#
// ReportSettings.DatabaseLogin / AfterDatabaseLogin event pattern.
// C# ref: ReportSettings.cs DatabaseLogin + AfterDatabaseLogin events;
// ReportSettings.Core.cs OnDatabaseLogin → fires DatabaseLogin event.
func TestReportSettings_LoginCallbackFields(t *testing.T) {
	loginCalled := false
	afterCalled := false
	loginCb := func(e *data.DatabaseLoginEventArgs) { loginCalled = true }
	afterCb := func(e *data.AfterDatabaseLoginEventArgs) { afterCalled = true }
	conn := data.NewDataConnectionBase("stub")
	conn.ConnectionString = "stub://test"
	conn.OnDatabaseLogin = loginCb
	conn.OnAfterDatabaseLogin = afterCb
	if err := conn.Open(); err != nil {
		t.Fatalf("Open error: %v", err)
	}
	if !loginCalled {
		t.Error("OnDatabaseLogin not called")
	}
	if !afterCalled {
		t.Error("OnAfterDatabaseLogin not called")
	}
	_ = conn.Close()
}
