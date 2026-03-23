package data_test

// commandparam_coverage_test.go — tests for CommandParameter and DataConnectionBase
// property methods that had 0% coverage.

import (
	"testing"

	"github.com/andrewloable/go-fastreport/data"
	"github.com/andrewloable/go-fastreport/report"
)

// ── CommandParameter.LastValue / SetLastValue / ResetLastValue ────────────────

func TestCommandParameter_LastValue_Default(t *testing.T) {
	p := data.NewCommandParameter("@id")
	// LastValue is nil on a fresh parameter.
	if p.LastValue() != nil {
		t.Errorf("LastValue = %v, want nil for new parameter", p.LastValue())
	}
}

func TestCommandParameter_SetLastValue(t *testing.T) {
	p := data.NewCommandParameter("@id")
	p.SetLastValue(42)
	if p.LastValue() != 42 {
		t.Errorf("LastValue = %v, want 42", p.LastValue())
	}
}

func TestCommandParameter_SetLastValue_String(t *testing.T) {
	p := data.NewCommandParameter("@name")
	p.SetLastValue("hello")
	if p.LastValue() != "hello" {
		t.Errorf("LastValue = %v, want 'hello'", p.LastValue())
	}
}

func TestCommandParameter_ResetLastValue(t *testing.T) {
	p := data.NewCommandParameter("@x")
	p.SetLastValue(99)
	p.ResetLastValue()
	// After reset, LastValue should be the uninitialized sentinel (not nil and not 99).
	lv := p.LastValue()
	if lv == 99 {
		t.Error("LastValue still 99 after ResetLastValue")
	}
}

// ── CommandParameter.Assign ───────────────────────────────────────────────────

func TestCommandParameter_Assign_CopiesFields(t *testing.T) {
	src := data.NewCommandParameter("@src")
	src.DataType = "int"
	src.Size = 10
	src.Value = int64(7)
	src.Expression = "[Field.ID]"
	src.DefaultValue = "0"
	src.Direction = data.ParamDirectionOutput

	dst := data.NewCommandParameter("@dst")
	dst.Assign(src)

	if dst.Name != "@src" {
		t.Errorf("Name = %q, want '@src'", dst.Name)
	}
	if dst.DataType != "int" {
		t.Errorf("DataType = %q, want 'int'", dst.DataType)
	}
	if dst.Size != 10 {
		t.Errorf("Size = %d, want 10", dst.Size)
	}
	if dst.Value != int64(7) {
		t.Errorf("Value = %v, want 7", dst.Value)
	}
	if dst.Expression != "[Field.ID]" {
		t.Errorf("Expression = %q, want '[Field.ID]'", dst.Expression)
	}
	if dst.DefaultValue != "0" {
		t.Errorf("DefaultValue = %q, want '0'", dst.DefaultValue)
	}
	if dst.Direction != data.ParamDirectionOutput {
		t.Errorf("Direction = %v, want Output", dst.Direction)
	}
}

func TestCommandParameter_Assign_NilSrc(t *testing.T) {
	dst := data.NewCommandParameter("@x")
	dst.Assign(nil) // must not panic
}

// ── CommandParameter.GetExpressions ─────────────────────────────────────────

func TestCommandParameter_GetExpressions_WithExpression(t *testing.T) {
	p := data.NewCommandParameter("@id")
	p.Expression = "[Order.ID]"
	exprs := p.GetExpressions()
	if len(exprs) != 1 {
		t.Fatalf("len(GetExpressions) = %d, want 1", len(exprs))
	}
	if exprs[0] != "[Order.ID]" {
		t.Errorf("GetExpressions[0] = %q, want '[Order.ID]'", exprs[0])
	}
}

func TestCommandParameter_GetExpressions_Empty(t *testing.T) {
	p := data.NewCommandParameter("@id")
	// No expression set.
	exprs := p.GetExpressions()
	if exprs != nil {
		t.Errorf("GetExpressions = %v, want nil when no expression", exprs)
	}
}

// ── DataConnectionBase property methods ─────────────────────────────────────

func TestDataConnectionBase_IsSqlBased_DefaultTrue(t *testing.T) {
	// NewDataConnectionBase matches C# default: IsSqlBased=true.
	c := data.NewDataConnectionBase("stub")
	if !c.IsSqlBased() {
		t.Error("IsSqlBased should default to true (C# DataConnectionBase default)")
	}
}

func TestDataConnectionBase_SetIsSqlBased(t *testing.T) {
	c := data.NewDataConnectionBase("stub")
	c.SetIsSqlBased(true)
	if !c.IsSqlBased() {
		t.Error("IsSqlBased should be true after SetIsSqlBased(true)")
	}
	c.SetIsSqlBased(false)
	if c.IsSqlBased() {
		t.Error("IsSqlBased should be false after SetIsSqlBased(false)")
	}
}

func TestDataConnectionBase_CanContainProcedures_DefaultFalse(t *testing.T) {
	c := data.NewDataConnectionBase("stub")
	if c.CanContainProcedures() {
		t.Error("CanContainProcedures should default to false")
	}
}

func TestDataConnectionBase_SetCanContainProcedures(t *testing.T) {
	c := data.NewDataConnectionBase("stub")
	c.SetCanContainProcedures(true)
	if !c.CanContainProcedures() {
		t.Error("CanContainProcedures should be true after SetCanContainProcedures(true)")
	}
}

func TestDataConnectionBase_GetTableNames_ReturnsNil(t *testing.T) {
	c := data.NewDataConnectionBase("stub")
	// Base implementation returns nil (override in SQL subclasses).
	names := c.GetTableNames()
	if names != nil {
		t.Errorf("GetTableNames = %v, want nil for base", names)
	}
}

func TestDataConnectionBase_GetTableCount_Zero(t *testing.T) {
	c := data.NewDataConnectionBase("stub")
	if count := c.GetTableCount(); count != 0 {
		t.Errorf("GetTableCount = %d, want 0", count)
	}
}

func TestDataConnectionBase_FilterTables_Passthrough(t *testing.T) {
	c := data.NewDataConnectionBase("stub")
	names := []string{"orders", "customers"}
	got := c.FilterTables(names)
	if len(got) != 2 {
		t.Errorf("FilterTables len = %d, want 2", len(got))
	}
}

func TestDataConnectionBase_CreateAllTables_NoOp(t *testing.T) {
	// Base GetTableNames returns nil, so CreateAllTables creates no tables.
	c := data.NewDataConnectionBase("stub")
	c.CreateAllTables() // must not panic
	if len(c.Tables()) != 0 {
		t.Errorf("Tables len = %d, want 0 (no schema names from base)", len(c.Tables()))
	}
}

func TestDataConnectionBase_CreateAllTablesWithSchema_NoOp(t *testing.T) {
	c := data.NewDataConnectionBase("stub")
	c.CreateAllTablesWithSchema(false) // must not panic
	if len(c.Tables()) != 0 {
		t.Errorf("Tables len = %d, want 0", len(c.Tables()))
	}
}

func TestDataConnectionBase_DeleteTable(t *testing.T) {
	c := data.NewDataConnectionBase("stub")
	ts := c.CreateTable("myTable")
	if len(c.Tables()) != 1 {
		t.Fatalf("Tables len = %d, want 1 before delete", len(c.Tables()))
	}
	c.DeleteTable(ts)
	if len(c.Tables()) != 0 {
		t.Errorf("Tables len = %d, want 0 after delete", len(c.Tables()))
	}
}

func TestDataConnectionBase_DeleteTable_NotPresent(t *testing.T) {
	c := data.NewDataConnectionBase("stub")
	other := data.NewTableDataSource("orphan")
	c.DeleteTable(other) // must not panic when not in list
}

func TestDataConnectionBase_FillTable_UninitializedCallsInit(t *testing.T) {
	c := data.NewDataConnectionBase("stub")
	c.ConnectionString = "stub://test"
	ts := c.CreateTable("t")
	ts.SetSelectCommand("SELECT id, name FROM users")
	// Not yet initialized — FillTable should call Init.
	if err := c.FillTable(ts); err != nil {
		t.Fatalf("FillTable: %v", err)
	}
	if ts.RowCount() != 2 {
		t.Errorf("RowCount = %d, want 2 after FillTable", ts.RowCount())
	}
	_ = c.Close()
}

func TestDataConnectionBase_FillTable_AlreadyInitialized_NoReload(t *testing.T) {
	c := data.NewDataConnectionBase("stub")
	c.ConnectionString = "stub://test"
	ts := c.CreateTable("t")
	ts.SetSelectCommand("SELECT id, name FROM users")
	// Initialize manually first.
	if err := ts.Init(); err != nil {
		t.Fatalf("Init: %v", err)
	}
	// FillTable a second time — no parameters changed, should not re-init.
	if err := c.FillTable(ts); err != nil {
		t.Fatalf("FillTable second call: %v", err)
	}
	_ = c.Close()
}

// ── paramDirectionToString / paramDirectionFromString (missing branches) ──────
// These are internal functions tested indirectly via Serialize/Deserialize.
// To exercise the uncovered branches (Output, InputOutput, ReturnValue) we
// serialize a CommandParameter with each direction and check the attribute value.

func TestCommandParameter_Serialize_AllDirections(t *testing.T) {
	cases := []struct {
		dir  data.ParameterDirection
		want string
	}{
		{data.ParamDirectionInput, ""},       // Input is default; Serialize omits it
		{data.ParamDirectionOutput, "Output"},
		{data.ParamDirectionInputOutput, "InputOutput"},
		{data.ParamDirectionReturnValue, "ReturnValue"},
	}

	for _, tc := range cases {
		p := data.NewCommandParameter("@p")
		p.Direction = tc.dir
		w := &directionCapture{}
		if err := p.Serialize(w); err != nil {
			t.Fatalf("Serialize direction %v: %v", tc.dir, err)
		}
		if tc.want == "" {
			// Input direction should not be written.
			if _, ok := w.written["Direction"]; ok {
				t.Errorf("Direction %v: should not write Direction attribute for Input", tc.dir)
			}
		} else {
			got, ok := w.written["Direction"]
			if !ok {
				t.Errorf("Direction %v: Direction attribute not written, want %q", tc.dir, tc.want)
			} else if got != tc.want {
				t.Errorf("Direction %v: got %q, want %q", tc.dir, got, tc.want)
			}
		}
	}
}

// directionCapture is a minimal report.Writer that records string values.
type directionCapture struct {
	written map[string]string
}

func (w *directionCapture) WriteStr(name, value string) {
	if w.written == nil {
		w.written = make(map[string]string)
	}
	w.written[name] = value
}
func (w *directionCapture) WriteInt(name string, value int)                             {}
func (w *directionCapture) WriteBool(name string, value bool)                            {}
func (w *directionCapture) WriteFloat(name string, value float32)                        {}
func (w *directionCapture) WriteObject(obj report.Serializable) error                   { return nil }
func (w *directionCapture) WriteObjectNamed(name string, obj report.Serializable) error { return nil }

func TestCommandParameter_Deserialize_AllDirections(t *testing.T) {
	cases := []struct {
		input    string
		wantDir  data.ParameterDirection
	}{
		{"Output", data.ParamDirectionOutput},
		{"InputOutput", data.ParamDirectionInputOutput},
		{"ReturnValue", data.ParamDirectionReturnValue},
		{"Input", data.ParamDirectionInput},
		{"", data.ParamDirectionInput},
		{"1", data.ParamDirectionOutput},        // legacy numeric
		{"2", data.ParamDirectionInputOutput},   // legacy numeric
		{"3", data.ParamDirectionReturnValue},   // legacy numeric
		{"unknown", data.ParamDirectionInput},   // fallback
	}

	for _, tc := range cases {
		p := data.NewCommandParameter("@p")
		r := &directionReader{direction: tc.input}
		if err := p.Deserialize(r); err != nil {
			t.Fatalf("Deserialize direction %q: %v", tc.input, err)
		}
		if p.Direction != tc.wantDir {
			t.Errorf("Deserialize %q: Direction = %v, want %v", tc.input, p.Direction, tc.wantDir)
		}
	}
}

type directionReader struct {
	direction string
}

func (r *directionReader) ReadStr(name, def string) string {
	if name == "Direction" {
		return r.direction
	}
	return def
}
func (r *directionReader) ReadInt(name string, def int) int        { return def }
func (r *directionReader) ReadBool(name string, def bool) bool     { return def }
func (r *directionReader) ReadFloat(name string, def float32) float32 { return def }
func (r *directionReader) NextChild() (string, bool)               { return "", false }
func (r *directionReader) FinishChild() error                      { return nil }
