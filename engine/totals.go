package engine

import (
	"strings"

	"github.com/andrewloable/go-fastreport/band"
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
	// Build a set of page-footer band names.
	// Mirrors C# Total.IsPageFooter: true when PrintOn is PageFooterBand,
	// ColumnFooterBand, or a HeaderFooterBandBase with RepeatOnEveryPage.
	// Only page-footer totals participate in the StartKeep/EndKeep mechanism;
	// group-footer totals must not be rolled back by keep-together (C# Total.cs).
	pageFooterBands := e.collectPageFooterBandNames()
	for _, at := range e.aggregateTotals {
		at.Reset()
		if at.PrintOn != "" {
			at.IsPageFooter = pageFooterBands[strings.ToLower(at.PrintOn)]
		}
	}
}

// collectPageFooterBandNames walks the report's page band tree and returns a
// lower-cased set of band names that qualify as "page footer" bands:
//   - PageFooterBand
//   - ColumnFooterBand
//   - HeaderFooterBandBase with RepeatOnEveryPage=true
//
// Mirrors C# Total.IsPageFooter (Total.cs).
func (e *ReportEngine) collectPageFooterBandNames() map[string]bool {
	result := make(map[string]bool)
	if e.report == nil {
		return result
	}
	for _, page := range e.report.Pages() {
		for _, b := range page.AllBands() {
			switch v := b.(type) {
			case *band.PageFooterBand:
				result[strings.ToLower(v.Name())] = true
			case *band.ColumnFooterBand:
				result[strings.ToLower(v.Name())] = true
			default:
				// Check for HeaderFooterBandBase with RepeatOnEveryPage.
				type repeatChecker interface {
					RepeatOnEveryPage() bool
					Name() string
				}
				if rc, ok := b.(repeatChecker); ok && rc.RepeatOnEveryPage() {
					result[strings.ToLower(rc.Name())] = true
				}
			}
		}
	}
	return result
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

// accumulateTotalsForBand is like accumulateTotals but only accumulates totals
// whose Evaluator matches bandName.  Totals with an empty Evaluator always
// accumulate (grand-total behaviour).
//
// C# ref: TotalCollection.ProcessBand → if (total.Evaluator == band) total.AddValue()
func (e *ReportEngine) accumulateTotalsForBand(bandName string) {
	if e.report == nil || len(e.aggregateTotals) == 0 {
		return
	}
	dict := e.report.Dictionary()

	for _, at := range e.aggregateTotals {
		// Skip if Evaluator is set and doesn't match the current band.
		if at.Evaluator != "" && !strings.EqualFold(at.Evaluator, bandName) {
			continue
		}

		var val any

		switch at.TotalType {
		case data.TotalTypeCount, data.TotalTypeCountDistinct:
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

		if at.EvaluateCondition != "" {
			condVal, err := e.report.Calc(at.EvaluateCondition)
			if err == nil {
				if b, ok := condVal.(bool); ok && !b {
					continue
				}
			}
		}

		_ = at.Add(val)

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

// resetTotalsForBand resets AggregateTotals whose PrintOn matches bandName and
// whose ResetAfterPrint flag is true.  repeated controls whether the band is a
// reprint; when repeated=true, only totals with ResetOnReprint=true are reset.
//
// C# ref: TotalCollection.ProcessBand →
//
//	else if (total.PrintOn == band && total.ResetAfterPrint)
//	    if (!band.Repeated || total.ResetOnReprint) total.ResetValue()
func (e *ReportEngine) resetTotalsForBand(bandName string, repeated bool) {
	if e.report == nil || len(e.aggregateTotals) == 0 || bandName == "" {
		return
	}
	for _, at := range e.aggregateTotals {
		if !strings.EqualFold(at.PrintOn, bandName) {
			continue
		}
		if !at.ResetAfterPrint {
			continue
		}
		if repeated && !at.ResetOnReprint {
			continue
		}
		at.Reset()
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
