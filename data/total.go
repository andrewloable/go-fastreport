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

	// internal state
	sum            float64
	count          int
	minVal         float64
	maxVal         float64
	distinctValues map[any]struct{}
	initialized    bool
}

// NewAggregateTotal creates an AggregateTotal with defaults (TotalTypeSum).
func NewAggregateTotal(name string) *AggregateTotal {
	return &AggregateTotal{
		Name:           name,
		TotalType:      TotalTypeSum,
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

