package reportpkg_test

// calc_buildenv_test.go — tests targeting the uncovered branches in
// buildCalcEnv (reportpkg/calc.go:128).  The branches involve:
//   1. Parameter/SystemVariable/Total names that contain special characters
//      (hyphen, space, dot) so that sanitizeIdent(name) != name, causing the
//      sanitized key to also be stored in the env.
//   2. r.calcDS set to a DataSource that does NOT implement columnarDataSource
//      (no Columns() method) — the type-assertion false branch at line 160.
//   3. r.calcDS set while r.dictionary is nil — the inner nil-dict guard at
//      line 176 is not taken, but still must not panic.

import (
	"testing"

	"github.com/andrewloable/go-fastreport/data"
	"github.com/andrewloable/go-fastreport/reportpkg"
)

// ---------------------------------------------------------------------------
// 1. Sanitized parameter name (hyphen/space triggers key != p.Name branch)
// ---------------------------------------------------------------------------

// TestBuildCalcEnv_ParameterWithHyphen verifies that a parameter whose name
// contains a hyphen (e.g. "Sub-Total") is stored under both the original name
// and the sanitized name ("Sub_Total") so that both forms can be used in
// expressions.
func TestBuildCalcEnv_ParameterWithHyphen(t *testing.T) {
	r := reportpkg.NewReport()
	r.Dictionary().AddParameter(&data.Parameter{Name: "Sub-Total", Value: int64(42)})

	// Evaluate using the sanitized key (Sub_Total) — this works because
	// buildCalcEnv stores env["Sub_Total"] = 42 when key != p.Name.
	val, err := r.Calc("[Sub-Total]")
	if err != nil {
		t.Fatalf("Calc [Sub-Total]: %v", err)
	}
	if val != int64(42) {
		t.Errorf("got %v (%T), want int64(42)", val, val)
	}
}

// TestBuildCalcEnv_ParameterWithSpace verifies that a parameter name with a
// space is stored under both the original and sanitized (underscore) forms.
func TestBuildCalcEnv_ParameterWithSpace(t *testing.T) {
	r := reportpkg.NewReport()
	r.Dictionary().AddParameter(&data.Parameter{Name: "First Name", Value: "Alice"})

	// Expression uses the sanitized key First_Name.
	val, err := r.Calc("First_Name")
	if err != nil {
		t.Fatalf("Calc First_Name: %v", err)
	}
	if val != "Alice" {
		t.Errorf("got %v, want Alice", val)
	}
}

// ---------------------------------------------------------------------------
// 2. Sanitized system variable name (hyphen triggers key != sv.Name branch)
// ---------------------------------------------------------------------------

// TestBuildCalcEnv_SystemVariableWithHyphen verifies that a system variable
// name containing a hyphen is stored under both original and sanitized keys.
func TestBuildCalcEnv_SystemVariableWithHyphen(t *testing.T) {
	r := reportpkg.NewReport()
	// AddSystemVariable with a hyphenated name.
	r.Dictionary().AddSystemVariable(&data.Parameter{Name: "Page-Count", Value: int64(5)})

	// Access it via the sanitized key Page_Count.
	val, err := r.Calc("Page_Count")
	if err != nil {
		t.Fatalf("Calc Page_Count: %v", err)
	}
	if val != int64(5) {
		t.Errorf("got %v (%T), want int64(5)", val, val)
	}
}

// TestBuildCalcEnv_SystemVariableWithSpace verifies the sanitized-key branch
// for a system variable name that contains a space.
func TestBuildCalcEnv_SystemVariableWithSpace(t *testing.T) {
	r := reportpkg.NewReport()
	r.Dictionary().AddSystemVariable(&data.Parameter{Name: "Report Date", Value: "2024-01-15"})

	val, err := r.Calc("Report_Date")
	if err != nil {
		t.Fatalf("Calc Report_Date: %v", err)
	}
	if val != "2024-01-15" {
		t.Errorf("got %v, want 2024-01-15", val)
	}
}

// ---------------------------------------------------------------------------
// 3. Sanitized total name (hyphen triggers key != t.Name branch)
// ---------------------------------------------------------------------------

// TestBuildCalcEnv_TotalWithHyphen verifies that a total whose name contains a
// hyphen is stored under the sanitized key in the env (lines 151-153).
func TestBuildCalcEnv_TotalWithHyphen(t *testing.T) {
	r := reportpkg.NewReport()
	r.Dictionary().AddTotal(&data.Total{Name: "Grand-Total", Value: float64(999.99)})

	// Evaluate using the sanitized key Grand_Total.
	val, err := r.Calc("Grand_Total")
	if err != nil {
		t.Fatalf("Calc Grand_Total: %v", err)
	}
	if val != float64(999.99) {
		t.Errorf("got %v, want 999.99", val)
	}
}

// TestBuildCalcEnv_TotalWithSpace verifies the sanitized-key branch for a
// total whose name contains a space.
func TestBuildCalcEnv_TotalWithSpace(t *testing.T) {
	r := reportpkg.NewReport()
	r.Dictionary().AddTotal(&data.Total{Name: "Sub Total", Value: int64(123)})

	val, err := r.Calc("Sub_Total")
	if err != nil {
		t.Fatalf("Calc Sub_Total: %v", err)
	}
	if val != int64(123) {
		t.Errorf("got %v (%T), want int64(123)", val, val)
	}
}

// ---------------------------------------------------------------------------
// 4. calcDS that does NOT implement columnarDataSource
// ---------------------------------------------------------------------------

// minimalDS is a DataSource implementation that intentionally does NOT have a
// Columns() method, so it does not satisfy the columnarDataSource interface.
// This exercises the type-assertion false branch at line 160 of buildCalcEnv.
type minimalDS struct {
	name  string
	alias string
}

func (m *minimalDS) Name() string              { return m.name }
func (m *minimalDS) Alias() string             { return m.alias }
func (m *minimalDS) Init() error               { return nil }
func (m *minimalDS) First() error              { return nil }
func (m *minimalDS) Next() error               { return data.ErrEOF }
func (m *minimalDS) EOF() bool                 { return true }
func (m *minimalDS) RowCount() int             { return 0 }
func (m *minimalDS) CurrentRowNo() int         { return -1 }
func (m *minimalDS) GetValue(string) (any, error) { return nil, nil }
func (m *minimalDS) Close() error              { return nil }

// TestBuildCalcEnv_CalcDS_NotColumnar verifies that setting a calcDS that does
// not implement columnarDataSource does not panic and the expression evaluator
// still works correctly for non-DS expressions.
func TestBuildCalcEnv_CalcDS_NotColumnar(t *testing.T) {
	r := reportpkg.NewReport()
	r.Dictionary().AddParameter(&data.Parameter{Name: "X", Value: int64(7)})

	// Set a calcDS that does NOT have Columns() — exercises the `!ok` branch.
	r.SetCalcContext(&minimalDS{name: "Minimal", alias: "Minimal"})

	val, err := r.Calc("[X]")
	if err != nil {
		t.Fatalf("Calc [X] with non-columnar calcDS: %v", err)
	}
	if val != int64(7) {
		t.Errorf("got %v (%T), want int64(7)", val, val)
	}
}

// TestBuildCalcEnv_CalcDS_NotColumnar_LiteralExpr confirms that literal
// expressions also evaluate correctly when calcDS is a non-columnar DataSource.
func TestBuildCalcEnv_CalcDS_NotColumnar_LiteralExpr(t *testing.T) {
	r := reportpkg.NewReport()
	r.SetCalcContext(&minimalDS{name: "Minimal", alias: "Minimal"})

	val, err := r.Calc("3 + 4")
	if err != nil {
		t.Fatalf("Calc literal with non-columnar calcDS: %v", err)
	}
	// expr-lang evaluates 3+4 as int.
	var got int64
	switch v := val.(type) {
	case int:
		got = int64(v)
	case int64:
		got = v
	default:
		t.Fatalf("unexpected type %T: %v", val, val)
	}
	if got != 7 {
		t.Errorf("got %v, want 7", got)
	}
}

// ---------------------------------------------------------------------------
// 5. calcDS set + nil dictionary (line 176 inner guard)
// ---------------------------------------------------------------------------

// TestBuildCalcEnv_CalcDS_NilDictionary verifies that when calcDS is set and
// r.dictionary is nil, the inner dictionary nil check at line 176 correctly
// prevents injectRelatedFields from being called (no panic).
func TestBuildCalcEnv_CalcDS_NilDictionary(t *testing.T) {
	r := reportpkg.NewReport()
	r.SetDictionary(nil)

	// Use a full BaseDataSource (which IS columnar) so we also exercise the
	// columnarDataSource branch with a nil dictionary.
	ds := data.NewBaseDataSource("Items")
	ds.SetAlias("Items")
	ds.AddColumn(data.Column{Name: "Amount"})
	ds.AddRow(map[string]any{"Amount": int64(55)})
	_ = ds.Init()
	_ = ds.First()

	r.SetCalcContext(ds)

	// Should not panic. The column value is available in the env.
	val, err := r.Calc("[Amount]")
	if err != nil {
		t.Fatalf("Calc [Amount] with nil dict and columnar calcDS: %v", err)
	}
	if val != int64(55) {
		t.Errorf("got %v (%T), want int64(55)", val, val)
	}
}

// TestBuildCalcEnv_CalcDS_NilDictionary_NonColumnar verifies the combination of
// non-columnar calcDS + nil dictionary does not panic.
func TestBuildCalcEnv_CalcDS_NilDictionary_NonColumnar(t *testing.T) {
	r := reportpkg.NewReport()
	r.SetDictionary(nil)
	r.SetCalcContext(&minimalDS{name: "M", alias: "M"})

	// Just a literal — should evaluate fine.
	val, err := r.Calc("100")
	if err != nil {
		t.Fatalf("Calc 100 with nil dict + non-columnar: %v", err)
	}
	if val != 100 {
		t.Errorf("got %v, want 100", val)
	}
}
