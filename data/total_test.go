package data_test

import (
	"math"
	"testing"

	"github.com/andrewloable/go-fastreport/data"
)

// -----------------------------------------------------------------------
// AggregateTotal — construction
// -----------------------------------------------------------------------

func TestNewAggregateTotal_Defaults(t *testing.T) {
	tot := data.NewAggregateTotal("Grand")
	if tot == nil {
		t.Fatal("NewAggregateTotal returned nil")
	}
	if tot.Name != "Grand" {
		t.Errorf("Name = %q, want Grand", tot.Name)
	}
	if tot.TotalType != data.TotalTypeSum {
		t.Errorf("TotalType default = %d, want Sum", tot.TotalType)
	}
	if !tot.ResetOnReprint {
		t.Error("ResetOnReprint default should be true (matching C# default)")
	}
}

// -----------------------------------------------------------------------
// Sum
// -----------------------------------------------------------------------

func TestAggregateTotal_Sum(t *testing.T) {
	tot := data.NewAggregateTotal("S")
	tot.TotalType = data.TotalTypeSum
	for _, v := range []any{1, 2, 3, 4} {
		if err := tot.Add(v); err != nil {
			t.Fatalf("Add error: %v", err)
		}
	}
	if tot.Value() != 10.0 {
		t.Errorf("Sum = %v, want 10", tot.Value())
	}
}

func TestAggregateTotal_Sum_FloatValues(t *testing.T) {
	tot := data.NewAggregateTotal("S")
	_ = tot.Add(1.5)
	_ = tot.Add(2.5)
	if tot.Value() != 4.0 {
		t.Errorf("Sum = %v, want 4", tot.Value())
	}
}

func TestAggregateTotal_Sum_NilSkipped(t *testing.T) {
	tot := data.NewAggregateTotal("S")
	_ = tot.Add(nil)
	_ = tot.Add(5)
	if tot.Value() != 5.0 {
		t.Errorf("Sum with nil = %v, want 5", tot.Value())
	}
}

// -----------------------------------------------------------------------
// Min / Max
// -----------------------------------------------------------------------

func TestAggregateTotal_Min(t *testing.T) {
	tot := data.NewAggregateTotal("M")
	tot.TotalType = data.TotalTypeMin
	for _, v := range []any{7, 3, 9, 1, 5} {
		_ = tot.Add(v)
	}
	if tot.Value() != 1.0 {
		t.Errorf("Min = %v, want 1", tot.Value())
	}
}

func TestAggregateTotal_Min_Empty(t *testing.T) {
	tot := data.NewAggregateTotal("M")
	tot.TotalType = data.TotalTypeMin
	if tot.Value() != nil {
		t.Errorf("Min of empty = %v, want nil", tot.Value())
	}
}

func TestAggregateTotal_Max(t *testing.T) {
	tot := data.NewAggregateTotal("M")
	tot.TotalType = data.TotalTypeMax
	for _, v := range []any{3, 9, 1, 7} {
		_ = tot.Add(v)
	}
	if tot.Value() != 9.0 {
		t.Errorf("Max = %v, want 9", tot.Value())
	}
}

func TestAggregateTotal_Max_Empty(t *testing.T) {
	tot := data.NewAggregateTotal("M")
	tot.TotalType = data.TotalTypeMax
	if tot.Value() != nil {
		t.Errorf("Max of empty = %v, want nil", tot.Value())
	}
}

// -----------------------------------------------------------------------
// Avg
// -----------------------------------------------------------------------

func TestAggregateTotal_Avg(t *testing.T) {
	tot := data.NewAggregateTotal("A")
	tot.TotalType = data.TotalTypeAvg
	for _, v := range []any{2, 4, 6} {
		_ = tot.Add(v)
	}
	if tot.Value() != 4.0 {
		t.Errorf("Avg = %v, want 4", tot.Value())
	}
}

func TestAggregateTotal_Avg_Empty(t *testing.T) {
	tot := data.NewAggregateTotal("A")
	tot.TotalType = data.TotalTypeAvg
	if tot.Value() != nil {
		t.Errorf("Avg of empty = %v, want nil", tot.Value())
	}
}

// -----------------------------------------------------------------------
// Count / CountDistinct
// -----------------------------------------------------------------------

func TestAggregateTotal_Count(t *testing.T) {
	tot := data.NewAggregateTotal("C")
	tot.TotalType = data.TotalTypeCount
	for _, v := range []any{"a", "b", "a", nil} {
		_ = tot.Add(v)
	}
	// nil is skipped, so count = 3
	if tot.Value() != 3 {
		t.Errorf("Count = %v, want 3", tot.Value())
	}
}

func TestAggregateTotal_CountDistinct(t *testing.T) {
	tot := data.NewAggregateTotal("CD")
	tot.TotalType = data.TotalTypeCountDistinct
	for _, v := range []any{"x", "y", "x", "z", "y"} {
		_ = tot.Add(v)
	}
	if tot.Value() != 3 {
		t.Errorf("CountDistinct = %v, want 3 (x,y,z)", tot.Value())
	}
}

// -----------------------------------------------------------------------
// Reset
// -----------------------------------------------------------------------

func TestAggregateTotal_Reset(t *testing.T) {
	tot := data.NewAggregateTotal("R")
	_ = tot.Add(10)
	_ = tot.Add(20)
	tot.Reset()
	if tot.Value() != 0.0 {
		t.Errorf("Sum after reset = %v, want 0", tot.Value())
	}
}

func TestAggregateTotal_Reset_ThenAccumulate(t *testing.T) {
	tot := data.NewAggregateTotal("R")
	tot.TotalType = data.TotalTypeMin
	_ = tot.Add(5)
	tot.Reset()
	_ = tot.Add(8)
	if tot.Value() != 8.0 {
		t.Errorf("Min after reset+add = %v, want 8", tot.Value())
	}
}

// -----------------------------------------------------------------------
// Add — non-numeric type error
// -----------------------------------------------------------------------

func TestAggregateTotal_Add_NonNumericError(t *testing.T) {
	tot := data.NewAggregateTotal("E")
	tot.TotalType = data.TotalTypeSum
	err := tot.Add("not a number")
	if err == nil {
		t.Error("expected error for non-numeric Sum value, got nil")
	}
}

// -----------------------------------------------------------------------
// Various numeric input types
// -----------------------------------------------------------------------

func TestAggregateTotal_IntTypes(t *testing.T) {
	tot := data.NewAggregateTotal("I")
	_ = tot.Add(int8(1))
	_ = tot.Add(int16(2))
	_ = tot.Add(int32(3))
	_ = tot.Add(int64(4))
	_ = tot.Add(uint(5))
	_ = tot.Add(float32(0.5))
	if math.Abs(tot.Value().(float64)-15.5) > 1e-6 {
		t.Errorf("Sum of mixed int types = %v, want 15.5", tot.Value())
	}
}

// -----------------------------------------------------------------------
// TotalType constants
// -----------------------------------------------------------------------

func TestTotalTypeConstants(t *testing.T) {
	types := []data.TotalType{
		data.TotalTypeSum,
		data.TotalTypeMin,
		data.TotalTypeMax,
		data.TotalTypeAvg,
		data.TotalTypeCount,
		data.TotalTypeCountDistinct,
	}
	seen := map[data.TotalType]bool{}
	for _, tt := range types {
		if seen[tt] {
			t.Errorf("duplicate TotalType value %d", tt)
		}
		seen[tt] = true
	}
}

// -----------------------------------------------------------------------
// StartKeep / EndKeep
// -----------------------------------------------------------------------

func TestAggregateTotal_StartKeep_EndKeep_Sum(t *testing.T) {
	tot := data.NewAggregateTotal("K")
	tot.TotalType = data.TotalTypeSum
	tot.IsPageFooter = true // StartKeep/EndKeep only act on page-footer totals
	_ = tot.Add(10)
	_ = tot.Add(20)
	// sum=30 at this point

	tot.StartKeep()

	// accumulate more after the snapshot
	_ = tot.Add(100)
	_ = tot.Add(200)
	// sum=330

	if tot.Value() != 330.0 {
		t.Errorf("Sum before EndKeep = %v, want 330", tot.Value())
	}

	tot.EndKeep()

	// should be restored to 30
	if tot.Value() != 30.0 {
		t.Errorf("Sum after EndKeep = %v, want 30", tot.Value())
	}
}

func TestAggregateTotal_StartKeep_EndKeep_MinMax(t *testing.T) {
	tot := data.NewAggregateTotal("K")
	tot.TotalType = data.TotalTypeMin
	tot.IsPageFooter = true // StartKeep/EndKeep only act on page-footer totals
	_ = tot.Add(5)
	_ = tot.Add(3)
	// min=3

	tot.StartKeep()

	_ = tot.Add(1)
	// min=1
	if tot.Value() != 1.0 {
		t.Errorf("Min before EndKeep = %v, want 1", tot.Value())
	}

	tot.EndKeep()
	if tot.Value() != 3.0 {
		t.Errorf("Min after EndKeep = %v, want 3", tot.Value())
	}
}

func TestAggregateTotal_StartKeep_EndKeep_Count(t *testing.T) {
	tot := data.NewAggregateTotal("K")
	tot.TotalType = data.TotalTypeCount
	tot.IsPageFooter = true // StartKeep/EndKeep only act on page-footer totals
	_ = tot.Add("a")
	_ = tot.Add("b")
	// count=2

	tot.StartKeep()

	_ = tot.Add("c")
	_ = tot.Add("d")
	_ = tot.Add("e")
	// count=5
	if tot.Value() != 5 {
		t.Errorf("Count before EndKeep = %v, want 5", tot.Value())
	}

	tot.EndKeep()
	if tot.Value() != 2 {
		t.Errorf("Count after EndKeep = %v, want 2", tot.Value())
	}
}

func TestAggregateTotal_StartKeep_EndKeep_Uninitialized(t *testing.T) {
	// StartKeep on a fresh total, then add values, then EndKeep should
	// restore to uninitialized state.
	tot := data.NewAggregateTotal("K")
	tot.TotalType = data.TotalTypeMax
	tot.IsPageFooter = true // StartKeep/EndKeep only act on page-footer totals

	tot.StartKeep()

	_ = tot.Add(42)
	if tot.Value() != 42.0 {
		t.Errorf("Max before EndKeep = %v, want 42", tot.Value())
	}

	tot.EndKeep()
	if tot.Value() != nil {
		t.Errorf("Max after EndKeep = %v, want nil (uninitialized)", tot.Value())
	}
}

// -----------------------------------------------------------------------
// Clone
// -----------------------------------------------------------------------

func TestAggregateTotal_Clone_CopiesConfig(t *testing.T) {
	orig := data.NewAggregateTotal("Orig")
	orig.TotalType = data.TotalTypeAvg
	orig.Expression = "[Orders.Amount]"
	orig.EvaluateCondition = "[Orders.Status] == \"Active\""
	orig.IncludeInvisibleRows = true
	orig.ResetAfterPrint = true
	orig.ResetOnReprint = false
	orig.Evaluator = "DataBand1"
	orig.PrintOn = "GroupFooter1"

	clone := orig.Clone()

	if clone.Name != "Orig" {
		t.Errorf("Clone Name = %q, want Orig", clone.Name)
	}
	if clone.TotalType != data.TotalTypeAvg {
		t.Errorf("Clone TotalType = %d, want Avg", clone.TotalType)
	}
	if clone.Expression != "[Orders.Amount]" {
		t.Errorf("Clone Expression = %q", clone.Expression)
	}
	if clone.EvaluateCondition != "[Orders.Status] == \"Active\"" {
		t.Errorf("Clone EvaluateCondition = %q", clone.EvaluateCondition)
	}
	if !clone.IncludeInvisibleRows {
		t.Error("Clone IncludeInvisibleRows should be true")
	}
	if !clone.ResetAfterPrint {
		t.Error("Clone ResetAfterPrint should be true")
	}
	if clone.ResetOnReprint {
		t.Error("Clone ResetOnReprint should be false")
	}
	if clone.Evaluator != "DataBand1" {
		t.Errorf("Clone Evaluator = %q, want DataBand1", clone.Evaluator)
	}
	if clone.PrintOn != "GroupFooter1" {
		t.Errorf("Clone PrintOn = %q, want GroupFooter1", clone.PrintOn)
	}
}

func TestAggregateTotal_Clone_FreshState(t *testing.T) {
	orig := data.NewAggregateTotal("Orig")
	orig.TotalType = data.TotalTypeSum
	_ = orig.Add(100)
	_ = orig.Add(200)
	// orig sum = 300

	clone := orig.Clone()

	// Clone should have fresh state (sum=0)
	if clone.Value() != 0.0 {
		t.Errorf("Clone Sum = %v, want 0 (fresh state)", clone.Value())
	}

	// Original should be unaffected
	if orig.Value() != 300.0 {
		t.Errorf("Original Sum = %v, want 300", orig.Value())
	}
}

func TestAggregateTotal_Clone_Independent(t *testing.T) {
	orig := data.NewAggregateTotal("Orig")
	orig.TotalType = data.TotalTypeSum

	clone := orig.Clone()

	// Accumulate into clone only
	_ = clone.Add(50)

	// Original should remain at 0
	if orig.Value() != 0.0 {
		t.Errorf("Original after clone.Add = %v, want 0", orig.Value())
	}
	if clone.Value() != 50.0 {
		t.Errorf("Clone after Add = %v, want 50", clone.Value())
	}
}

// -----------------------------------------------------------------------
// New properties — Evaluator, PrintOn, ResetOnReprint
// -----------------------------------------------------------------------

func TestAggregateTotal_EvaluatorAndPrintOn(t *testing.T) {
	tot := data.NewAggregateTotal("T")
	tot.Evaluator = "DataBand1"
	tot.PrintOn = "GroupFooter1"
	if tot.Evaluator != "DataBand1" {
		t.Errorf("Evaluator = %q, want DataBand1", tot.Evaluator)
	}
	if tot.PrintOn != "GroupFooter1" {
		t.Errorf("PrintOn = %q, want GroupFooter1", tot.PrintOn)
	}
}

func TestAggregateTotal_ResetOnReprint_CanBeDisabled(t *testing.T) {
	tot := data.NewAggregateTotal("T")
	if !tot.ResetOnReprint {
		t.Fatal("default ResetOnReprint should be true")
	}
	tot.ResetOnReprint = false
	if tot.ResetOnReprint {
		t.Error("ResetOnReprint should be false after setting")
	}
}
