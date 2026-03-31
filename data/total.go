package data

import (
	"fmt"
	"math"
)

// TotalType identifies the aggregate function a Total computes.
type TotalType int

const (
	// TotalTypeSum computes the sum of values.
	TotalTypeSum TotalType = iota
	// TotalTypeMin computes the minimum value.
	TotalTypeMin
	// TotalTypeMax computes the maximum value.
	TotalTypeMax
	// TotalTypeAvg computes the average value.
	TotalTypeAvg
	// TotalTypeCount counts non-nil values.
	TotalTypeCount
	// TotalTypeCountDistinct counts distinct non-nil values.
	TotalTypeCountDistinct
)

// AggregateTotal is a richer Total that tracks aggregate state.
// It is the Go equivalent of FastReport.Data.Total.
//
// The simple Total struct in helper.go acts as a name/value pair; this type
// performs the actual accumulation used by the engine.
type AggregateTotal struct {
	// Name is the total's unique identifier (used in expressions like [GrandTotal]).
	Name string
	// TotalType is the aggregate function.
	TotalType TotalType
	// Expression is the value expression evaluated per row (empty for Count).
	Expression string
	// EvaluateCondition is an optional filter expression; empty means always evaluate.
	EvaluateCondition string
	// IncludeInvisibleRows includes rows whose band is hidden.
	IncludeInvisibleRows bool
	// ResetAfterPrint resets the total after it is printed.
	ResetAfterPrint bool
	// ResetOnReprint controls whether the total resets when a band is reprinted
	// (e.g. RepeatOnEveryPage). C# default is true.
	ResetOnReprint bool
	// Evaluator is the name of the DataBand that drives this total.
	Evaluator string
	// PrintOn is the name of the band where the total is printed/reset.
	PrintOn string

	// IsPageFooter mirrors C# Total.IsPageFooter: true when PrintOn is a
	// PageFooterBand, ColumnFooterBand, or a HeaderFooterBandBase with
	// RepeatOnEveryPage.  StartKeep/EndKeep only take effect for page-footer
	// totals; all other totals are unaffected by the keep-together mechanism.
	// Set by the engine during initTotals().
	IsPageFooter bool

	// internal state
	sum            float64
	count          int
	minVal         float64
	maxVal         float64
	distinctValues map[any]struct{}
	initialized    bool

	// keep-together snapshot fields (used by StartKeep/EndKeep)
	keepSum         float64
	keepCount       int
	keepMinVal      float64
	keepMaxVal      float64
	keepInitialized bool
}

// NewAggregateTotal creates an AggregateTotal with defaults (TotalTypeSum,
// ResetOnReprint=true matching C# default).
func NewAggregateTotal(name string) *AggregateTotal {
	return &AggregateTotal{
		Name:           name,
		TotalType:      TotalTypeSum,
		ResetOnReprint: true,
		distinctValues: make(map[any]struct{}),
	}
}

// Reset clears the accumulated state.
func (t *AggregateTotal) Reset() {
	t.sum = 0
	t.count = 0
	t.minVal = math.MaxFloat64
	t.maxVal = -math.MaxFloat64
	t.distinctValues = make(map[any]struct{})
	t.initialized = false
}

// Add accumulates a value into the aggregate.
// value must be convertible to float64 for numeric aggregates.
func (t *AggregateTotal) Add(value any) error {
	if value == nil {
		return nil
	}
	switch t.TotalType {
	case TotalTypeCount:
		t.count++
		return nil
	case TotalTypeCountDistinct:
		t.distinctValues[value] = struct{}{}
		return nil
	}
	f, err := toFloat64(value)
	if err != nil {
		return fmt.Errorf("AggregateTotal %q: %w", t.Name, err)
	}
	t.sum += f
	t.count++
	if !t.initialized || f < t.minVal {
		t.minVal = f
	}
	if !t.initialized || f > t.maxVal {
		t.maxVal = f
	}
	t.initialized = true
	return nil
}

// Value returns the computed aggregate result.
func (t *AggregateTotal) Value() any {
	switch t.TotalType {
	case TotalTypeSum:
		return t.sum
	case TotalTypeMin:
		if !t.initialized {
			return nil
		}
		return t.minVal
	case TotalTypeMax:
		if !t.initialized {
			return nil
		}
		return t.maxVal
	case TotalTypeAvg:
		if t.count == 0 {
			return nil
		}
		return t.sum / float64(t.count)
	case TotalTypeCount:
		return t.count
	case TotalTypeCountDistinct:
		return len(t.distinctValues)
	default:
		return nil
	}
}

// StartKeep snapshots the current accumulator state into internal keep* fields.
// Only page-footer totals (IsPageFooter=true) participate in the keep mechanism.
// Mirrors C# Total.StartKeep() which early-returns when !IsPageFooter.
func (t *AggregateTotal) StartKeep() {
	if !t.IsPageFooter {
		return
	}
	t.keepSum = t.sum
	t.keepCount = t.count
	t.keepMinVal = t.minVal
	t.keepMaxVal = t.maxVal
	t.keepInitialized = t.initialized
}

// EndKeep restores the accumulator state from the snapshot taken by StartKeep.
// Only page-footer totals (IsPageFooter=true) are affected.
// Mirrors C# Total.EndKeep() which early-returns when !IsPageFooter.
func (t *AggregateTotal) EndKeep() {
	if !t.IsPageFooter {
		return
	}
	t.sum = t.keepSum
	t.count = t.keepCount
	t.minVal = t.keepMinVal
	t.maxVal = t.keepMaxVal
	t.initialized = t.keepInitialized
}

// Clone creates a copy of the AggregateTotal with the same configuration
// but fresh (zero) accumulator state. The keep-together snapshot fields are
// also reset.
func (t *AggregateTotal) Clone() *AggregateTotal {
	return &AggregateTotal{
		Name:                 t.Name,
		TotalType:            t.TotalType,
		Expression:           t.Expression,
		EvaluateCondition:    t.EvaluateCondition,
		IncludeInvisibleRows: t.IncludeInvisibleRows,
		ResetAfterPrint:      t.ResetAfterPrint,
		ResetOnReprint:       t.ResetOnReprint,
		Evaluator:            t.Evaluator,
		PrintOn:              t.PrintOn,
		distinctValues:       make(map[any]struct{}),
	}
}

