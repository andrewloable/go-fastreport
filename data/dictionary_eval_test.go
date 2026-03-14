package data_test

import (
	"testing"

	"github.com/andrewloable/go-fastreport/data"
)

// ── Evaluate — basic value lookup ─────────────────────────────────────────────

func TestDictionary_Evaluate_RawValue(t *testing.T) {
	d := data.NewDictionary()
	d.AddParameter(&data.Parameter{Name: "MaxRows", Value: 100})

	val, err := d.Evaluate("MaxRows")
	if err != nil {
		t.Fatalf("Evaluate error: %v", err)
	}
	if val != 100 {
		t.Errorf("Value = %v, want 100", val)
	}
}

func TestDictionary_Evaluate_StringValue(t *testing.T) {
	d := data.NewDictionary()
	d.AddParameter(&data.Parameter{Name: "Title", Value: "Sales Report"})

	val, err := d.Evaluate("Title")
	if err != nil {
		t.Fatalf("Evaluate error: %v", err)
	}
	if val != "Sales Report" {
		t.Errorf("Value = %v, want 'Sales Report'", val)
	}
}

// ── Evaluate — expression ─────────────────────────────────────────────────────

func TestDictionary_Evaluate_Expression_Arithmetic(t *testing.T) {
	d := data.NewDictionary()
	d.AddParameter(&data.Parameter{Name: "Base", Value: 10})
	d.AddParameter(&data.Parameter{
		Name:       "Double",
		Expression: "Base * 2",
	})

	val, err := d.Evaluate("Double")
	if err != nil {
		t.Fatalf("Evaluate error: %v", err)
	}
	if v, ok := val.(int); !ok || v != 20 {
		t.Errorf("Double = %v (%T), want 20 (int)", val, val)
	}
}

func TestDictionary_Evaluate_Expression_Comparison(t *testing.T) {
	d := data.NewDictionary()
	d.AddParameter(&data.Parameter{Name: "Limit", Value: 50})
	d.AddParameter(&data.Parameter{
		Name:       "IsOver",
		Expression: "Limit > 40",
	})

	val, err := d.Evaluate("IsOver")
	if err != nil {
		t.Fatalf("Evaluate error: %v", err)
	}
	if v, ok := val.(bool); !ok || !v {
		t.Errorf("IsOver = %v, want true", val)
	}
}

func TestDictionary_Evaluate_Expression_UsesSystemVars(t *testing.T) {
	d := data.NewDictionary()
	d.SetSystemVariable("PageNumber", 5)
	d.AddParameter(&data.Parameter{
		Name:       "PageLabel",
		Expression: "PageNumber + 1",
	})

	val, err := d.Evaluate("PageLabel")
	if err != nil {
		t.Fatalf("Evaluate error: %v", err)
	}
	if v, ok := val.(int); !ok || v != 6 {
		t.Errorf("PageLabel = %v, want 6", val)
	}
}

// ── Evaluate — not found ──────────────────────────────────────────────────────

func TestDictionary_Evaluate_NotFound(t *testing.T) {
	d := data.NewDictionary()
	_, err := d.Evaluate("NonExistent")
	if err == nil {
		t.Error("expected error for missing parameter")
	}
}

// ── EvaluateAll ───────────────────────────────────────────────────────────────

func TestDictionary_EvaluateAll(t *testing.T) {
	d := data.NewDictionary()
	d.AddParameter(&data.Parameter{Name: "X", Value: 3})
	d.AddParameter(&data.Parameter{Name: "Y", Expression: "X * 10"})
	d.AddParameter(&data.Parameter{Name: "Z", Value: "static"})

	if err := d.EvaluateAll(); err != nil {
		t.Fatalf("EvaluateAll error: %v", err)
	}

	yParam := d.FindParameter("Y")
	if yParam == nil {
		t.Fatal("parameter Y not found")
	}
	if v, ok := yParam.Value.(int); !ok || v != 30 {
		t.Errorf("Y.Value after EvaluateAll = %v, want 30", yParam.Value)
	}

	zParam := d.FindParameter("Z")
	if zParam.Value != "static" {
		t.Errorf("Z.Value changed unexpectedly: %v", zParam.Value)
	}
}

// ── Nested parameter lookup ───────────────────────────────────────────────────

func TestDictionary_Evaluate_NestedParameter(t *testing.T) {
	d := data.NewDictionary()
	parent := &data.Parameter{Name: "Filters"}
	child := &data.Parameter{Name: "MinDate", Value: "2024-01-01"}
	parent.AddParameter(child)
	d.AddParameter(parent)

	val, err := d.Evaluate("Filters.MinDate")
	if err != nil {
		t.Fatalf("Evaluate nested error: %v", err)
	}
	if val != "2024-01-01" {
		t.Errorf("Filters.MinDate = %v, want 2024-01-01", val)
	}
}

// ── Invalid expression ────────────────────────────────────────────────────────

func TestDictionary_Evaluate_InvalidExpression(t *testing.T) {
	d := data.NewDictionary()
	d.AddParameter(&data.Parameter{
		Name:       "Bad",
		Expression: "this is not valid +++",
	})

	_, err := d.Evaluate("Bad")
	if err == nil {
		t.Error("expected error for invalid expression")
	}
}
