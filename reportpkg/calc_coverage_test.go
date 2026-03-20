package reportpkg_test

// calc_coverage_test.go — additional tests to increase coverage for Calc,
// buildCalcEnv, and coerceCalcValue in reportpkg/calc.go.

import (
	"testing"

	"github.com/andrewloable/go-fastreport/data"
	"github.com/andrewloable/go-fastreport/reportpkg"
)

// ── Calc — nil dictionary edge cases ─────────────────────────────────────

// TestCalc_NilDictionary exercises the code path where r.dictionary is nil.
// buildCalcEnv must not panic, and a simple literal expression still works.
func TestCalc_NilDictionary(t *testing.T) {
	r := reportpkg.NewReport()
	r.SetDictionary(nil)

	// A numeric literal expression does not require the dictionary.
	val, err := r.Calc("1 + 1")
	if err != nil {
		t.Fatalf("Calc with nil dictionary: %v", err)
	}
	if val != 2 {
		t.Errorf("got %v, want 2", val)
	}
}

// TestCalc_NilDictionary_HashVariable tests the nil-dictionary branch inside
// the '#'-variable special-case handling in Calc (lines 60-68 of calc.go).
func TestCalc_NilDictionary_HashVariable(t *testing.T) {
	r := reportpkg.NewReport()
	r.SetDictionary(nil)

	// With a nil dictionary, a '#'-name should return an error (unknown macro).
	_, err := r.Calc("[PageNumber#]")
	if err == nil {
		t.Error("expected error for unknown macro variable with nil dictionary")
	}
}

// ── Calc — literal (no-bracket) expression ────────────────────────────────

// TestCalc_LiteralString confirms that a plain string literal (no brackets)
// evaluates to its value unchanged.
func TestCalc_LiteralString(t *testing.T) {
	r := reportpkg.NewReport()

	val, err := r.Calc(`"hello"`)
	if err != nil {
		t.Fatalf("Calc literal string: %v", err)
	}
	if val != "hello" {
		t.Errorf("got %v, want hello", val)
	}
}

// TestCalc_LiteralNumber confirms that a bare numeric literal evaluates correctly.
func TestCalc_LiteralNumber(t *testing.T) {
	r := reportpkg.NewReport()

	val, err := r.Calc("42")
	if err != nil {
		t.Fatalf("Calc literal number: %v", err)
	}
	if val != 42 {
		t.Errorf("got %v (type %T), want 42", val, val)
	}
}

// ── Calc — boolean expression ─────────────────────────────────────────────

// TestCalc_BooleanExpression exercises boolean evaluation in Calc.
func TestCalc_BooleanExpression(t *testing.T) {
	r := reportpkg.NewReport()
	r.Dictionary().AddParameter(&data.Parameter{Name: "X", Value: 5})
	r.Dictionary().AddParameter(&data.Parameter{Name: "Y", Value: 3})

	val, err := r.Calc("[X] > [Y]")
	if err != nil {
		t.Fatalf("Calc bool expr: %v", err)
	}
	b, ok := val.(bool)
	if !ok {
		t.Fatalf("expected bool, got %T (%v)", val, val)
	}
	if !b {
		t.Errorf("got false, want true for 5 > 3")
	}
}

// TestCalc_BooleanFalse exercises a boolean expression that evaluates to false.
func TestCalc_BooleanFalse(t *testing.T) {
	r := reportpkg.NewReport()
	r.Dictionary().AddParameter(&data.Parameter{Name: "A", Value: 1})
	r.Dictionary().AddParameter(&data.Parameter{Name: "B", Value: 10})

	val, err := r.Calc("[A] > [B]")
	if err != nil {
		t.Fatalf("Calc bool false: %v", err)
	}
	b, ok := val.(bool)
	if !ok {
		t.Fatalf("expected bool, got %T (%v)", val, val)
	}
	if b {
		t.Errorf("got true, want false for 1 > 10")
	}
}

// ── Calc — isSimpleDottedIdent path ───────────────────────────────────────

// TestCalc_SimpleDottedIdent tests the isSimpleDottedIdent → sanitized lookup
// path in Calc (lines 79-83 of calc.go). A system variable stored with a
// dot-qualified name should be resolved via that branch.
func TestCalc_SimpleDottedIdent_InEnv(t *testing.T) {
	r := reportpkg.NewReport()
	// Add a system variable whose sanitized form matches the env key.
	// Report.ReportInfo.Description → Report_ReportInfo_Description
	r.Dictionary().SetSystemVariable("Report_ReportInfo_Description", "TestDesc")

	// Use the dot form to trigger isSimpleDottedIdent.
	val, err := r.Calc("Report_ReportInfo_Description")
	if err != nil {
		t.Fatalf("Calc dotted ident: %v", err)
	}
	if val != "TestDesc" {
		t.Errorf("got %v, want TestDesc", val)
	}
}

// ── Calc — complex expression with multiple fields ────────────────────────

// TestCalc_ComplexExpression tests "[Field1] + [Field2]" where both fields
// come from an injected data source.
func TestCalc_ComplexExpression_TwoFields(t *testing.T) {
	r := reportpkg.NewReport()

	ds := data.NewBaseDataSource("Items")
	ds.SetAlias("Items")
	ds.AddColumn(data.Column{Name: "Price"})
	ds.AddColumn(data.Column{Name: "Quantity"})
	ds.AddRow(map[string]any{"Price": "10", "Quantity": "3"})
	_ = ds.Init()
	_ = ds.First()

	r.SetCalcContext(ds)

	// Both "10" and "3" are string values; coerceCalcValue converts them to
	// int64, so the expression evaluates to 10 + 3 = 13.
	val, err := r.Calc("[Price] + [Quantity]")
	if err != nil {
		t.Fatalf("Calc complex expr: %v", err)
	}
	// The result of int64(10) + int64(3) is int64(13).
	var got int64
	switch v := val.(type) {
	case int64:
		got = v
	case int:
		got = int64(v)
	default:
		t.Fatalf("unexpected type %T: %v", val, val)
	}
	if got != 13 {
		t.Errorf("got %v, want 13", got)
	}
}

// TestCalc_ComplexExpression_FloatArithmetic tests "[UnitPrice] * [Qty]" where
// UnitPrice is a float string, exercising the string→float64 coercion path.
func TestCalc_ComplexExpression_FloatArithmetic(t *testing.T) {
	r := reportpkg.NewReport()

	ds := data.NewBaseDataSource("Orders")
	ds.SetAlias("Orders")
	ds.AddColumn(data.Column{Name: "UnitPrice"})
	ds.AddColumn(data.Column{Name: "Qty"})
	ds.AddRow(map[string]any{"UnitPrice": "2.5", "Qty": "4"})
	_ = ds.Init()
	_ = ds.First()

	r.SetCalcContext(ds)

	val, err := r.Calc("[UnitPrice] * [Qty]")
	if err != nil {
		t.Fatalf("Calc float arithmetic: %v", err)
	}
	var got float64
	switch v := val.(type) {
	case float64:
		got = v
	case int64:
		got = float64(v)
	default:
		t.Fatalf("unexpected type %T: %v", val, val)
	}
	if got != 10.0 {
		t.Errorf("got %v, want 10.0", got)
	}
}

// ── buildCalcEnv — nil dictionary path ───────────────────────────────────

// TestBuildCalcEnv_NilDictionary confirms buildCalcEnv does not panic and
// returns a usable (possibly empty) environment when the dictionary is nil.
// This exercises the `if r.dictionary != nil` false-branch in buildCalcEnv.
func TestBuildCalcEnv_NilDictionary(t *testing.T) {
	r := reportpkg.NewReport()
	r.SetDictionary(nil)

	// A literal expression works even with no dictionary.
	val, err := r.Calc("7 * 6")
	if err != nil {
		t.Fatalf("Calc with nil dictionary env: %v", err)
	}
	if val != 42 {
		t.Errorf("got %v, want 42", val)
	}
}

// TestBuildCalcEnv_EmptyDictionary confirms buildCalcEnv handles a dictionary
// with no entries without error.
func TestBuildCalcEnv_EmptyDictionary(t *testing.T) {
	r := reportpkg.NewReport()
	// Dictionary is non-nil but empty (default from NewReport).

	val, err := r.Calc("100")
	if err != nil {
		t.Fatalf("Calc with empty dictionary: %v", err)
	}
	if val != 100 {
		t.Errorf("got %v, want 100", val)
	}
}

// TestBuildCalcEnv_WithCustomFunction confirms that a registered custom function
// is available in the expression environment, exercising the customFunctions
// injection loop in buildCalcEnv.
func TestBuildCalcEnv_WithCustomFunction(t *testing.T) {
	r := reportpkg.NewReport()
	r.RegisterFunction("Double", func(args []any) (any, error) {
		v := args[0].(int)
		return v * 2, nil
	})

	val, err := r.Calc("Double(5)")
	if err != nil {
		t.Fatalf("Calc with custom function: %v", err)
	}
	if val != 10 {
		t.Errorf("got %v, want 10", val)
	}
}

// ── coerceCalcValue — all type branches ──────────────────────────────────

// TestCoerceCalcValue_StringToInt64 exercises the string→int64 conversion
// branch of coerceCalcValue (via a data source column value "123").
func TestCoerceCalcValue_StringToInt64(t *testing.T) {
	r := reportpkg.NewReport()

	ds := data.NewBaseDataSource("DS")
	ds.SetAlias("DS")
	ds.AddColumn(data.Column{Name: "Val"})
	ds.AddRow(map[string]any{"Val": "123"})
	_ = ds.Init()
	_ = ds.First()

	r.SetCalcContext(ds)

	val, err := r.Calc("[Val]")
	if err != nil {
		t.Fatalf("Calc string→int64: %v", err)
	}
	if val != int64(123) {
		t.Errorf("got %v (type %T), want int64(123)", val, val)
	}
}

// TestCoerceCalcValue_StringToFloat64 exercises the string→float64 conversion
// branch of coerceCalcValue (via a data source column value "3.14").
func TestCoerceCalcValue_StringToFloat64(t *testing.T) {
	r := reportpkg.NewReport()

	ds := data.NewBaseDataSource("DS")
	ds.SetAlias("DS")
	ds.AddColumn(data.Column{Name: "Pi"})
	ds.AddRow(map[string]any{"Pi": "3.14"})
	_ = ds.Init()
	_ = ds.First()

	r.SetCalcContext(ds)

	val, err := r.Calc("[Pi]")
	if err != nil {
		t.Fatalf("Calc string→float64: %v", err)
	}
	f, ok := val.(float64)
	if !ok {
		t.Fatalf("expected float64, got %T (%v)", val, val)
	}
	if f != 3.14 {
		t.Errorf("got %v, want 3.14", f)
	}
}

// TestCoerceCalcValue_StringStaysString exercises the pass-through branch of
// coerceCalcValue where the string cannot be parsed as int64 or float64.
func TestCoerceCalcValue_StringStaysString(t *testing.T) {
	r := reportpkg.NewReport()

	ds := data.NewBaseDataSource("DS")
	ds.SetAlias("DS")
	ds.AddColumn(data.Column{Name: "Label"})
	ds.AddRow(map[string]any{"Label": "abc"})
	_ = ds.Init()
	_ = ds.First()

	r.SetCalcContext(ds)

	val, err := r.Calc("[Label]")
	if err != nil {
		t.Fatalf("Calc string stays string: %v", err)
	}
	if val != "abc" {
		t.Errorf("got %v, want abc", val)
	}
}

// TestCoerceCalcValue_Int32 exercises the non-string pass-through branch of
// coerceCalcValue with an int32 value (stored directly in the data source).
func TestCoerceCalcValue_Int32(t *testing.T) {
	r := reportpkg.NewReport()

	ds := data.NewBaseDataSource("DS")
	ds.SetAlias("DS")
	ds.AddColumn(data.Column{Name: "Count"})
	ds.AddRow(map[string]any{"Count": int32(42)})
	_ = ds.Init()
	_ = ds.First()

	r.SetCalcContext(ds)

	val, err := r.Calc("[Count]")
	if err != nil {
		t.Fatalf("Calc int32 passthrough: %v", err)
	}
	if val != int32(42) {
		t.Errorf("got %v (type %T), want int32(42)", val, val)
	}
}

// TestCoerceCalcValue_Uint exercises the non-string pass-through branch of
// coerceCalcValue with a uint value.
func TestCoerceCalcValue_Uint(t *testing.T) {
	r := reportpkg.NewReport()

	ds := data.NewBaseDataSource("DS")
	ds.SetAlias("DS")
	ds.AddColumn(data.Column{Name: "Flags"})
	ds.AddRow(map[string]any{"Flags": uint(255)})
	_ = ds.Init()
	_ = ds.First()

	r.SetCalcContext(ds)

	val, err := r.Calc("[Flags]")
	if err != nil {
		t.Fatalf("Calc uint passthrough: %v", err)
	}
	if val != uint(255) {
		t.Errorf("got %v (type %T), want uint(255)", val, val)
	}
}

// TestCoerceCalcValue_Bool exercises the non-string pass-through branch of
// coerceCalcValue with a bool value.
func TestCoerceCalcValue_Bool(t *testing.T) {
	r := reportpkg.NewReport()

	ds := data.NewBaseDataSource("DS")
	ds.SetAlias("DS")
	ds.AddColumn(data.Column{Name: "Active"})
	ds.AddRow(map[string]any{"Active": true})
	_ = ds.Init()
	_ = ds.First()

	r.SetCalcContext(ds)

	val, err := r.Calc("[Active]")
	if err != nil {
		t.Fatalf("Calc bool passthrough: %v", err)
	}
	if val != true {
		t.Errorf("got %v, want true", val)
	}
}

// ── Calc — wasSingleBracket sanitized path ────────────────────────────────

// TestCalc_SingleBracketWithSpaces exercises the wasSingleBracket branch in
// Calc where the unwrapped token contains spaces and is sanitized before
// being looked up in the environment (lines 84-90 of calc.go).
func TestCalc_SingleBracketWithSpaces(t *testing.T) {
	r := reportpkg.NewReport()

	// Register "Order_Details_Orders_ShipName" in the dictionary — this is
	// the sanitized form of "Order Details.Orders.ShipName".
	r.Dictionary().AddParameter(&data.Parameter{
		Name:  "Order_Details_Orders_ShipName",
		Value: "Berlin",
	})

	// The bracketed form with spaces triggers the wasSingleBracket path.
	val, err := r.Calc("[Order Details.Orders.ShipName]")
	if err != nil {
		t.Fatalf("Calc bracketed with spaces: %v", err)
	}
	if val != "Berlin" {
		t.Errorf("got %v, want Berlin", val)
	}
}

// ── Calc — hash variable with dictionary present ──────────────────────────

// TestCalc_HashVariable_Found tests the '#'-variable lookup path when the
// dictionary has a matching system variable.
func TestCalc_HashVariable_Found(t *testing.T) {
	r := reportpkg.NewReport()
	r.Dictionary().SetSystemVariable("Row#", int64(5))

	val, err := r.Calc("[Row#]")
	if err != nil {
		t.Fatalf("Calc Row#: %v", err)
	}
	if val != int64(5) {
		t.Errorf("got %v, want int64(5)", val)
	}
}

// TestCalc_HashVariable_NotFound tests the '#'-variable path when the dictionary
// does not contain the variable (returns error, not panic).
func TestCalc_HashVariable_NotFound(t *testing.T) {
	r := reportpkg.NewReport()
	// Dictionary is present but doesn't have "Unknown#".

	_, err := r.Calc("[Unknown#]")
	if err == nil {
		t.Error("expected error for unknown '#' variable")
	}
}
