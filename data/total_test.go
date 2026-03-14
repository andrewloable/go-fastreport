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
