package engine

import (
	"strings"

	"github.com/andrewloable/go-fastreport/data"
)

// initTotals copies the AggregateTotal definitions from the report Dictionary
// into the engine so it can accumulate per-row.  Called once at engine start.
func (e *ReportEngine) initTotals() {
	if e.report == nil {
		return
	}
	dict := e.report.Dictionary()
	if dict == nil {
		return
	}
	e.aggregateTotals = dict.AggregateTotals()
	// Reset all accumulators for a clean run.
	for _, at := range e.aggregateTotals {
		at.Reset()
	}
}

// accumulateTotals evaluates each aggregate total's Expression using the
// current row context (already set via SetCalcContext) and calls Add() on
// the accumulator.  After accumulation, the corresponding simple Total.Value
// in the Dictionary is updated so expression evaluation picks up the current
// running total.
func (e *ReportEngine) accumulateTotals() {
	if e.report == nil || len(e.aggregateTotals) == 0 {
		return
	}
	dict := e.report.Dictionary()

	for _, at := range e.aggregateTotals {
		var val any

		switch at.TotalType {
		case data.TotalTypeCount, data.TotalTypeCountDistinct:
			// Count totals don't need an expression — they count every row.
			val = 1
		default:
			if at.Expression == "" {
				continue
			}
			v, err := e.report.Calc(at.Expression)
			if err != nil {
				continue
			}
			val = v
		}

		// Evaluate condition if set.
		if at.EvaluateCondition != "" {
			condVal, err := e.report.Calc(at.EvaluateCondition)
			if err == nil {
				if b, ok := condVal.(bool); ok && !b {
					continue
				}
			}
		}

		_ = at.Add(val)

		// Sync to dictionary's simple Total so expressions see the new value.
		if dict != nil {
			for _, t := range dict.Totals() {
				if strings.EqualFold(t.Name, at.Name) {
					t.Value = at.Value()
					break
				}
			}
		}
	}
}

// resetGroupTotals resets AggregateTotals that are group-scoped.
// Called after each group footer is rendered.
// Currently resets totals marked ResetAfterPrint=true.
func (e *ReportEngine) resetGroupTotals() {
	for _, at := range e.aggregateTotals {
		if at.ResetAfterPrint {
			at.Reset()
			// Zero out the dictionary simple total too.
			if dict := e.report.Dictionary(); dict != nil {
				for _, t := range dict.Totals() {
					if strings.EqualFold(t.Name, at.Name) {
						t.Value = at.Value()
						break
					}
				}
			}
		}
	}
}
