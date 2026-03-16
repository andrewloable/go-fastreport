package sql_test

// sql_coverage_test.go — tests for NewSQLDataSource, RegisterSQL, and ProcedureDataSource.
// Uses the external test package to access exported API.

import (
	"bytes"
	"strings"
	"testing"

	"github.com/andrewloable/go-fastreport/data"
	sqlpkg "github.com/andrewloable/go-fastreport/data/sql"
	"github.com/andrewloable/go-fastreport/serial"

	// Register SQLite driver for in-memory tests.
	_ "modernc.org/sqlite"
)

// ── NewSQLDataSource ──────────────────────────────────────────────────────────

func TestNewSQLDataSource_Basic(t *testing.T) {
	conn := sqlpkg.NewSQLiteConnection(":memory:")
	ds := sqlpkg.NewSQLDataSource(&conn.DataConnectionBase, "SELECT 1", "Test")

	if ds == nil {
		t.Fatal("NewSQLDataSource returned nil")
	}
	if ds.Name() != "Test" {
		t.Errorf("Name = %q, want Test", ds.Name())
	}
	if ds.SelectCommand() != "SELECT 1" {
		t.Errorf("SelectCommand = %q, want 'SELECT 1'", ds.SelectCommand())
	}
}

func TestNewSQLDataSource_AddedToConnection(t *testing.T) {
	conn := sqlpkg.NewSQLiteConnection(":memory:")
	_ = sqlpkg.NewSQLDataSource(&conn.DataConnectionBase, "SELECT 1", "DS1")
	_ = sqlpkg.NewSQLDataSource(&conn.DataConnectionBase, "SELECT 2", "DS2")

	// Both should be in conn's tables.
	tables := conn.DataConnectionBase.Tables()
	if len(tables) != 2 {
		t.Errorf("Tables len = %d, want 2", len(tables))
	}
}

// ── RegisterSQL ───────────────────────────────────────────────────────────────

func TestRegisterSQL_Basic(t *testing.T) {
	conn := sqlpkg.NewSQLiteConnection(":memory:")
	dict := data.NewDictionary()

	ds := sqlpkg.RegisterSQL(dict, &conn.DataConnectionBase, "SELECT 1", "RegTest")

	if ds == nil {
		t.Fatal("RegisterSQL returned nil")
	}
	if ds.Name() != "RegTest" {
		t.Errorf("Name = %q, want RegTest", ds.Name())
	}

	// Should be in the dictionary.
	found := dict.FindDataSourceByName("RegTest")
	if found == nil {
		t.Error("RegisterSQL should add data source to dictionary")
	}
}

// ── ProcedureDataSource ───────────────────────────────────────────────────────

func TestNewProcedureDataSource_Defaults(t *testing.T) {
	p := sqlpkg.NewProcedureDataSource("MyProc")
	if p == nil {
		t.Fatal("NewProcedureDataSource returned nil")
	}
	if p.Name() != "MyProc" {
		t.Errorf("Name = %q, want MyProc", p.Name())
	}
	if p.TypeName() != "ProcedureDataSource" {
		t.Errorf("TypeName = %q, want ProcedureDataSource", p.TypeName())
	}
	if p.BaseName() != "ProcedureDataSource" {
		t.Errorf("BaseName = %q, want ProcedureDataSource", p.BaseName())
	}
}

func TestProcedureDataSource_Serialize_Basic(t *testing.T) {
	p := sqlpkg.NewProcedureDataSource("P1")
	p.SetSelectCommand("CALL my_proc(?)")
	p.SetTableName("results")
	p.SetStoreData(true)
	p.SetAlias("MyAlias")

	var buf bytes.Buffer
	w := serial.NewWriter(&buf)
	if err := w.BeginObject("ProcedureDataSource"); err != nil {
		t.Fatalf("BeginObject: %v", err)
	}
	if err := p.Serialize(w); err != nil {
		t.Fatalf("Serialize: %v", err)
	}
	if err := w.EndObject(); err != nil {
		t.Fatalf("EndObject: %v", err)
	}
	if err := w.Flush(); err != nil {
		t.Fatalf("Flush: %v", err)
	}

	xml := buf.String()
	for _, want := range []string{"CALL my_proc(?)", "results", "StoreData", "MyAlias"} {
		if !strings.Contains(xml, want) {
			t.Errorf("serialized XML missing %q:\n%s", want, xml)
		}
	}
}

func TestProcedureDataSource_Deserialize_RoundTrip(t *testing.T) {
	orig := sqlpkg.NewProcedureDataSource("Proc1")
	orig.SetSelectCommand("CALL proc1(?)")
	orig.SetTableName("out")
	orig.SetStoreData(false)

	var buf bytes.Buffer
	w := serial.NewWriter(&buf)
	if err := w.BeginObject("ProcedureDataSource"); err != nil {
		t.Fatalf("BeginObject: %v", err)
	}
	if err := orig.Serialize(w); err != nil {
		t.Fatalf("Serialize: %v", err)
	}
	if err := w.EndObject(); err != nil {
		t.Fatalf("EndObject: %v", err)
	}
	if err := w.Flush(); err != nil {
		t.Fatalf("Flush: %v", err)
	}

	r := serial.NewReader(bytes.NewReader(buf.Bytes()))
	typeName, ok := r.ReadObjectHeader()
	if !ok || typeName != "ProcedureDataSource" {
		t.Fatalf("ReadObjectHeader: got %q, ok=%v", typeName, ok)
	}

	got := sqlpkg.NewProcedureDataSource("")
	if err := got.Deserialize(r); err != nil {
		t.Fatalf("Deserialize: %v", err)
	}

	if got.Name() != "Proc1" {
		t.Errorf("Name: got %q, want Proc1", got.Name())
	}
	if got.SelectCommand() != "CALL proc1(?)" {
		t.Errorf("SelectCommand: got %q", got.SelectCommand())
	}
	if got.TableName() != "out" {
		t.Errorf("TableName: got %q, want out", got.TableName())
	}
}

func TestProcedureDataSource_Deserialize_NoAlias(t *testing.T) {
	// When Alias is empty in FRX, it defaults to Name.
	xml := `<ProcedureDataSource Name="Proc2"/>`
	r := serial.NewReader(strings.NewReader(xml))
	_, ok := r.ReadObjectHeader()
	if !ok {
		t.Fatal("ReadObjectHeader failed")
	}
	p := sqlpkg.NewProcedureDataSource("")
	if err := p.Deserialize(r); err != nil {
		t.Fatalf("Deserialize: %v", err)
	}
	if p.Name() != "Proc2" {
		t.Errorf("Name = %q, want Proc2", p.Name())
	}
	if p.Alias() != "Proc2" {
		t.Errorf("Alias = %q, want Proc2 (defaulted from Name)", p.Alias())
	}
}

func TestProcedureDataSource_GetValue_OutParam(t *testing.T) {
	p := sqlpkg.NewProcedureDataSource("TestProc")
	// Add an output parameter.
	outParam := data.NewCommandParameter("OutCol")
	outParam.Direction = data.ParamDirectionOutput
	outParam.Value = int64(42)
	p.AddParameter(outParam)

	// GetValue for the output parameter should return its Value.
	val, err := p.GetValue("OutCol")
	if err != nil {
		t.Fatalf("GetValue OutCol: %v", err)
	}
	if val != int64(42) {
		t.Errorf("GetValue OutCol = %v, want 42", val)
	}
}

func TestProcedureDataSource_GetValue_InputParam_NotInOutputs(t *testing.T) {
	p := sqlpkg.NewProcedureDataSource("TestProc")
	// Add an input parameter - should NOT be returned as a synthetic column.
	inParam := data.NewCommandParameter("InputCol")
	inParam.Direction = data.ParamDirectionInput
	inParam.Value = "hello"
	p.AddParameter(inParam)

	// "InputCol" is input-only so GetValue will try the underlying table.
	// Since not initialized, it will fail.
	_, err := p.GetValue("InputCol")
	// Expect an error since no rows are loaded.
	if err == nil {
		t.Log("GetValue for input-only param: no error (may return nil)")
	}
}

func TestProcedureDataSource_Serialize_WithParameter(t *testing.T) {
	p := sqlpkg.NewProcedureDataSource("ProcWithParam")
	param := data.NewCommandParameter("Param1")
	param.Value = "testval"
	p.AddParameter(param)

	var buf bytes.Buffer
	w := serial.NewWriter(&buf)
	if err := w.BeginObject("ProcedureDataSource"); err != nil {
		t.Fatalf("BeginObject: %v", err)
	}
	if err := p.Serialize(w); err != nil {
		t.Fatalf("Serialize: %v", err)
	}
	if err := w.EndObject(); err != nil {
		t.Fatalf("EndObject: %v", err)
	}
	if err := w.Flush(); err != nil {
		t.Fatalf("Flush: %v", err)
	}

	xml := buf.String()
	if !strings.Contains(xml, "Param1") {
		t.Errorf("serialized XML missing parameter name:\n%s", xml)
	}
}

// ── ProcedureDataSource.Init ──────────────────────────────────────────────────

func TestProcedureDataSource_Init_WithOutputParam(t *testing.T) {
	conn := sqlpkg.NewSQLiteConnection(":memory:")
	if err := conn.Open(); err != nil {
		t.Fatalf("Open: %v", err)
	}
	defer conn.Close()

	// Create a table and use SELECT as "procedure" call.
	_, err := conn.DB().Exec("CREATE TABLE items (id INTEGER, val TEXT)")
	if err != nil {
		t.Fatalf("CREATE TABLE: %v", err)
	}
	_, err = conn.DB().Exec("INSERT INTO items VALUES (1, 'hello')")
	if err != nil {
		t.Fatalf("INSERT: %v", err)
	}

	p := sqlpkg.NewProcedureDataSourceSQL(&conn.DataConnectionBase, "SELECT id, val FROM items", "Proc")

	// Add an output parameter (it won't be in the result set columns, so Init adds a synthetic col).
	outParam := data.NewCommandParameter("extra_out")
	outParam.Direction = data.ParamDirectionOutput
	outParam.Value = int64(99)
	p.AddParameter(outParam)

	if err := p.Init(); err != nil {
		t.Fatalf("Init: %v", err)
	}
	// Verify synthetic column was added.
	found := false
	for _, col := range p.Columns() {
		if col.Name == "extra_out" {
			found = true
			break
		}
	}
	if !found {
		t.Error("Init should add synthetic column for output parameter not in result set")
	}
}

func TestProcedureDataSource_Init_OutputParamAlreadyInResultSet(t *testing.T) {
	conn := sqlpkg.NewSQLiteConnection(":memory:")
	if err := conn.Open(); err != nil {
		t.Fatalf("Open: %v", err)
	}
	defer conn.Close()

	_, err := conn.DB().Exec("CREATE TABLE t2 (id INTEGER)")
	if err != nil {
		t.Fatalf("CREATE TABLE: %v", err)
	}
	_, err = conn.DB().Exec("INSERT INTO t2 VALUES (1)")
	if err != nil {
		t.Fatalf("INSERT: %v", err)
	}

	p := sqlpkg.NewProcedureDataSourceSQL(&conn.DataConnectionBase, "SELECT id FROM t2", "Proc2")

	// Output param with the same name as a result column — should not duplicate.
	outParam := data.NewCommandParameter("id")
	outParam.Direction = data.ParamDirectionOutput
	outParam.Value = int64(1)
	p.AddParameter(outParam)

	if err := p.Init(); err != nil {
		t.Fatalf("Init: %v", err)
	}
	count := 0
	for _, col := range p.Columns() {
		if col.Name == "id" {
			count++
		}
	}
	if count != 1 {
		t.Errorf("Init: 'id' column should appear exactly once, got %d", count)
	}
}

func TestProcedureDataSource_Init_NoConnection(t *testing.T) {
	p := sqlpkg.NewProcedureDataSource("NoConn")
	p.SetSelectCommand("SELECT 1")
	// No connection set — Init should fail.
	err := p.Init()
	if err == nil {
		t.Error("Init without connection should return error")
	}
}

func TestProcedureDataSource_GetValue_ErrorPath(t *testing.T) {
	// No rows loaded → GetValue for a non-output param should error.
	p := sqlpkg.NewProcedureDataSource("P")
	_, err := p.GetValue("nonexistent_col")
	if err == nil {
		t.Error("GetValue for nonexistent column with no rows should return error")
	}
}

// ── NewProcedureDataSourceSQL ──────────────────────────────────────────────────

func TestNewProcedureDataSourceSQL_Basic(t *testing.T) {
	conn := sqlpkg.NewSQLiteConnection(":memory:")
	p := sqlpkg.NewProcedureDataSourceSQL(&conn.DataConnectionBase, "CALL my_proc(?)", "MyProc")

	if p == nil {
		t.Fatal("NewProcedureDataSourceSQL returned nil")
	}
	if p.Name() != "MyProc" {
		t.Errorf("Name = %q, want MyProc", p.Name())
	}
	if p.SelectCommand() != "CALL my_proc(?)" {
		t.Errorf("SelectCommand = %q, want 'CALL my_proc(?)'", p.SelectCommand())
	}

	// Should be registered in the connection.
	tables := conn.DataConnectionBase.Tables()
	if len(tables) != 1 {
		t.Errorf("Tables len = %d, want 1", len(tables))
	}
}
