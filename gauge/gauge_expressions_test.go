package gauge_test

import (
	"testing"

	"github.com/andrewloable/go-fastreport/gauge"
)

func TestGaugeObject_GetExpressions_Empty(t *testing.T) {
	g := gauge.NewGaugeObject()
	exprs := g.GetExpressions()
	// No Expression set — only base component expressions (all empty by default)
	for _, e := range exprs {
		if e == "" {
			t.Error("GetExpressions should not return empty strings")
		}
	}
}

func TestGaugeObject_GetExpressions_WithExpression(t *testing.T) {
	g := gauge.NewGaugeObject()
	g.Expression = "[Orders.Amount]"
	exprs := g.GetExpressions()
	found := false
	for _, e := range exprs {
		if e == "[Orders.Amount]" {
			found = true
		}
	}
	if !found {
		t.Errorf("GetExpressions = %v, want to contain %q", exprs, "[Orders.Amount]")
	}
}

func TestGaugeObject_Clone_IndependentFields(t *testing.T) {
	src := gauge.NewGaugeObject()
	src.Minimum = 10
	src.Maximum = 200
	src.Expression = "[Gauge.Value]"

	dst := src.Clone()
	if dst == nil {
		t.Fatal("Clone returned nil")
	}
	if dst.Minimum != 10 || dst.Maximum != 200 {
		t.Errorf("Clone Minimum=%v Maximum=%v, want 10,200", dst.Minimum, dst.Maximum)
	}
	if dst.Expression != "[Gauge.Value]" {
		t.Errorf("Clone Expression = %q, want %q", dst.Expression, "[Gauge.Value]")
	}

	// Mutating dst should not affect src
	dst.Minimum = 0
	dst.Expression = "other"
	if src.Minimum != 10 || src.Expression != "[Gauge.Value]" {
		t.Error("Clone is not independent from source")
	}
}

func TestGaugeObject_Clone_ScaleIndependent(t *testing.T) {
	src := gauge.NewGaugeObject()
	src.Scale.MajorStep = 25

	dst := src.Clone()
	dst.Scale.MajorStep = 50
	if src.Scale.MajorStep != 25 {
		t.Error("Clone Scale should be independent from source")
	}
}

func TestLinearGauge_GetExpressions_WithExpression(t *testing.T) {
	g := gauge.NewLinearGauge()
	g.Expression = "[Total]"
	exprs := g.GetExpressions()
	found := false
	for _, e := range exprs {
		if e == "[Total]" {
			found = true
		}
	}
	if !found {
		t.Errorf("LinearGauge.GetExpressions = %v, want to contain %q", exprs, "[Total]")
	}
}
