package sql_test

// procedure_extra_test.go — additional branch coverage for ProcedureDataSource
// to push Serialize, Init, and GetValue closer to 100%.

import (
	"fmt"
	"testing"

	"github.com/andrewloable/go-fastreport/data"
	sqlpkg "github.com/andrewloable/go-fastreport/data/sql"
	"github.com/andrewloable/go-fastreport/report"

	_ "modernc.org/sqlite"
)

// failSQLWriter is a report.Writer that always fails on WriteObjectNamed.
type failSQLWriter struct{}

func (w *failSQLWriter) WriteStr(name, value string)            {}
func (w *failSQLWriter) WriteInt(name string, value int)        {}
func (w *failSQLWriter) WriteBool(name string, value bool)      {}
func (w *failSQLWriter) WriteFloat(name string, value float32)  {}
func (w *failSQLWriter) WriteObject(obj report.Serializable) error {
	return fmt.Errorf("failSQLWriter: WriteObject fails")
}
func (w *failSQLWriter) WriteObjectNamed(name string, obj report.Serializable) error {
	return fmt.Errorf("failSQLWriter: WriteObjectNamed fails")
}

// ── Serialize: error return when WriteObjectNamed fails ───────────────────────

func TestProcedureDataSource_Serialize_WriteObjectNamedError(t *testing.T) {
	p := sqlpkg.NewProcedureDataSource("ProcErr")
	// Add a parameter so the loop body is entered.
	p.AddParameter(data.NewCommandParameter("@x"))

	fw := &failSQLWriter{}
	err := p.Serialize(fw)
	if err == nil {
		t.Error("Serialize should propagate error from WriteObjectNamed")
	}
}

// ── Init: input-direction parameters must be skipped (continue branch) ────────

func TestProcedureDataSource_Init_InputParamSkipped(t *testing.T) {
	conn := sqlpkg.NewSQLiteConnection(":memory:")
	if err := conn.Open(); err != nil {
		t.Fatalf("Open: %v", err)
	}
	defer conn.Close()

	_, err := conn.DB().Exec("CREATE TABLE vals (n INTEGER)")
	if err != nil {
		t.Fatalf("CREATE TABLE: %v", err)
	}
	_, err = conn.DB().Exec("INSERT INTO vals VALUES (7)")
	if err != nil {
		t.Fatalf("INSERT: %v", err)
	}

	p := sqlpkg.NewProcedureDataSourceSQL(&conn.DataConnectionBase, "SELECT n FROM vals", "ProcInput")

	// Add an input parameter — Init should skip it (the "continue" branch) and
	// must NOT create a synthetic column for it.
	inParam := data.NewCommandParameter("myInputParam")
	inParam.Direction = data.ParamDirectionInput
	inParam.Value = int64(0)
	p.AddParameter(inParam)

	if err := p.Init(); err != nil {
		t.Fatalf("Init: %v", err)
	}

	// "myInputParam" is an input-only param — it must NOT appear as a column.
	for _, col := range p.Columns() {
		if col.Name == "myInputParam" {
			t.Error("Init should skip input-direction parameters; found synthetic column for input param")
		}
	}
}

// ── GetValue: success path through TableDataSource ────────────────────────────

func TestProcedureDataSource_GetValue_SuccessPath(t *testing.T) {
	conn := sqlpkg.NewSQLiteConnection(":memory:")
	if err := conn.Open(); err != nil {
		t.Fatalf("Open: %v", err)
	}
	defer conn.Close()

	_, err := conn.DB().Exec("CREATE TABLE score (pts INTEGER)")
	if err != nil {
		t.Fatalf("CREATE TABLE: %v", err)
	}
	_, err = conn.DB().Exec("INSERT INTO score VALUES (99)")
	if err != nil {
		t.Fatalf("INSERT: %v", err)
	}

	p := sqlpkg.NewProcedureDataSourceSQL(&conn.DataConnectionBase, "SELECT pts FROM score", "ProcScore")

	if err := p.Init(); err != nil {
		t.Fatalf("Init: %v", err)
	}
	if err := p.First(); err != nil {
		t.Fatalf("First: %v", err)
	}

	// GetValue for a regular result-set column exercises the success return path:
	// the output-param loop produces no match, so control falls through to
	// TableDataSource.GetValue which succeeds and returns a value.
	val, err := p.GetValue("pts")
	if err != nil {
		t.Fatalf("GetValue(pts): %v", err)
	}
	if val == nil {
		t.Error("GetValue(pts) returned nil, want a non-nil value")
	}
}

// ── GetValue: InputOutput parameter is returned directly ─────────────────────

func TestProcedureDataSource_GetValue_InputOutputParam(t *testing.T) {
	p := sqlpkg.NewProcedureDataSource("TestProc2")
	// InputOutput direction is treated as an output — should be returned directly.
	ioParam := data.NewCommandParameter("ioCol")
	ioParam.Direction = data.ParamDirectionInputOutput
	ioParam.Value = "resultValue"
	p.AddParameter(ioParam)

	val, err := p.GetValue("ioCol")
	if err != nil {
		t.Fatalf("GetValue ioCol: %v", err)
	}
	if val != "resultValue" {
		t.Errorf("GetValue ioCol = %v, want resultValue", val)
	}
}

// ── Init: mix of input and output params ─────────────────────────────────────

func TestProcedureDataSource_Init_MixedParams(t *testing.T) {
	conn := sqlpkg.NewSQLiteConnection(":memory:")
	if err := conn.Open(); err != nil {
		t.Fatalf("Open: %v", err)
	}
	defer conn.Close()

	_, err := conn.DB().Exec("CREATE TABLE mixed (x INTEGER)")
	if err != nil {
		t.Fatalf("CREATE TABLE: %v", err)
	}
	_, err = conn.DB().Exec("INSERT INTO mixed VALUES (1)")
	if err != nil {
		t.Fatalf("INSERT: %v", err)
	}

	p := sqlpkg.NewProcedureDataSourceSQL(&conn.DataConnectionBase, "SELECT x FROM mixed", "ProcMixed")

	// Input param — should be skipped by Init (continue branch).
	inParam := data.NewCommandParameter("filterVal")
	inParam.Direction = data.ParamDirectionInput
	p.AddParameter(inParam)

	// Output param not in result set — should get a synthetic column.
	outParam := data.NewCommandParameter("status")
	outParam.Direction = data.ParamDirectionOutput
	outParam.Value = "ok"
	p.AddParameter(outParam)

	if err := p.Init(); err != nil {
		t.Fatalf("Init: %v", err)
	}

	foundInput := false
	foundOutput := false
	for _, col := range p.Columns() {
		if col.Name == "filterVal" {
			foundInput = true
		}
		if col.Name == "status" {
			foundOutput = true
		}
	}
	if foundInput {
		t.Error("input param should NOT appear as a synthetic column")
	}
	if !foundOutput {
		t.Error("output param not in result set SHOULD appear as a synthetic column")
	}
}
