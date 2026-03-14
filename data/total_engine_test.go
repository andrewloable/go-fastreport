package data_test

import (
	"testing"

	"github.com/andrewloable/go-fastreport/data"
)

// ── Construction ──────────────────────────────────────────────────────────────

func TestNewTotalEngine_Empty(t *testing.T) {
	te := data.NewTotalEngine()
	if te == nil {
		t.Fatal("NewTotalEngine returned nil")
	}
	if te.Len() != 0 {
		t.Errorf("expected 0 totals, got %d", te.Len())
	}
}

// ── Register ──────────────────────────────────────────────────────────────────

func TestTotalEngine_Register(t *testing.T) {
	te := data.NewTotalEngine()
	tot := data.NewAggregateTotal("Grand")
	te.Register(tot)

	if te.Len() != 1 {
		t.Errorf("expected 1 total, got %d", te.Len())
	}
	if te.Find("Grand") == nil {
		t.Error("Find('Grand') should return registered total")
	}
}

func TestTotalEngine_Register_Replace(t *testing.T) {
	te := data.NewTotalEngine()
	t1 := data.NewAggregateTotal("T")
	t1.TotalType = data.TotalTypeSum
	te.Register(t1)

	t2 := data.NewAggregateTotal("T")
	t2.TotalType = data.TotalTypeCount
	te.Register(t2) // replaces t1

	if te.Len() != 1 {
		t.Errorf("expected 1 total after replace, got %d", te.Len())
	}
	if te.Find("T").TotalType != data.TotalTypeCount {
		t.Error("replacement should override the original total")
	}
}

// ── Accumulate ────────────────────────────────────────────────────────────────

func TestTotalEngine_Accumulate_Sum(t *testing.T) {
	te := data.NewTotalEngine()
	tot := data.NewAggregateTotal("S")
	tot.TotalType = data.TotalTypeSum
	te.Register(tot)

	for _, v := range []any{1, 2, 3} {
		if err := te.Accumulate("S", v); err != nil {
			t.Fatalf("Accumulate error: %v", err)
		}
	}

	if te.Value("S") != 6.0 {
		t.Errorf("Sum = %v, want 6", te.Value("S"))
	}
}

func TestTotalEngine_Accumulate_NotRegistered(t *testing.T) {
	te := data.NewTotalEngine()
	err := te.Accumulate("unknown", 42)
	if err == nil {
		t.Error("expected error for unknown total name")
	}
}

// ── Value ──────────────────────────────────────────────────────────────────────

func TestTotalEngine_Value_NotRegistered(t *testing.T) {
	te := data.NewTotalEngine()
	if v := te.Value("nonexistent"); v != nil {
		t.Errorf("Value for unknown name = %v, want nil", v)
	}
}

func TestTotalEngine_Value_Count(t *testing.T) {
	te := data.NewTotalEngine()
	tot := data.NewAggregateTotal("C")
	tot.TotalType = data.TotalTypeCount
	te.Register(tot)
	_ = te.Accumulate("C", "a")
	_ = te.Accumulate("C", "b")
	_ = te.Accumulate("C", "c")
	if te.Value("C") != 3 {
		t.Errorf("Count = %v, want 3", te.Value("C"))
	}
}

// ── Reset ──────────────────────────────────────────────────────────────────────

func TestTotalEngine_Reset(t *testing.T) {
	te := data.NewTotalEngine()
	tot := data.NewAggregateTotal("R")
	te.Register(tot)
	_ = te.Accumulate("R", 100)

	if err := te.Reset("R"); err != nil {
		t.Fatalf("Reset error: %v", err)
	}
	if te.Value("R") != 0.0 {
		t.Errorf("Sum after reset = %v, want 0", te.Value("R"))
	}
}

func TestTotalEngine_Reset_NotRegistered(t *testing.T) {
	te := data.NewTotalEngine()
	if err := te.Reset("ghost"); err == nil {
		t.Error("expected error resetting unknown total")
	}
}

func TestTotalEngine_ResetAll(t *testing.T) {
	te := data.NewTotalEngine()
	for _, name := range []string{"A", "B", "C"} {
		tot := data.NewAggregateTotal(name)
		te.Register(tot)
		_ = te.Accumulate(name, 10)
	}

	te.ResetAll()

	for _, name := range []string{"A", "B", "C"} {
		if te.Value(name) != 0.0 {
			t.Errorf("Value(%q) after ResetAll = %v, want 0", name, te.Value(name))
		}
	}
}

// ── All ────────────────────────────────────────────────────────────────────────

func TestTotalEngine_All_Order(t *testing.T) {
	te := data.NewTotalEngine()
	names := []string{"First", "Second", "Third"}
	for _, n := range names {
		te.Register(data.NewAggregateTotal(n))
	}

	all := te.All()
	if len(all) != 3 {
		t.Fatalf("All() returned %d totals, want 3", len(all))
	}
	for i, name := range names {
		if all[i].Name != name {
			t.Errorf("All()[%d].Name = %q, want %q", i, all[i].Name, name)
		}
	}
}

// ── Multiple total types ───────────────────────────────────────────────────────

func TestTotalEngine_MultipleTypes(t *testing.T) {
	te := data.NewTotalEngine()

	sum := data.NewAggregateTotal("Sum")
	sum.TotalType = data.TotalTypeSum
	te.Register(sum)

	cnt := data.NewAggregateTotal("Count")
	cnt.TotalType = data.TotalTypeCount
	te.Register(cnt)

	for _, v := range []any{5, 10, 15} {
		_ = te.Accumulate("Sum", v)
		_ = te.Accumulate("Count", v)
	}

	if te.Value("Sum") != 30.0 {
		t.Errorf("Sum = %v, want 30", te.Value("Sum"))
	}
	if te.Value("Count") != 3 {
		t.Errorf("Count = %v, want 3", te.Value("Count"))
	}
}

// ── Accumulate + Reset cycle ───────────────────────────────────────────────────

func TestTotalEngine_AccumulateThenReset_ThenAccumulate(t *testing.T) {
	te := data.NewTotalEngine()
	tot := data.NewAggregateTotal("T")
	tot.TotalType = data.TotalTypeMax
	te.Register(tot)

	_ = te.Accumulate("T", 100)
	_ = te.Accumulate("T", 200)
	_ = te.Reset("T")
	_ = te.Accumulate("T", 50)

	if te.Value("T") != 50.0 {
		t.Errorf("Max after reset+accumulate = %v, want 50", te.Value("T"))
	}
}
