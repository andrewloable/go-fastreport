package engine

import (
	"github.com/andrewloable/go-fastreport/band"
	"github.com/andrewloable/go-fastreport/data"
	"github.com/andrewloable/go-fastreport/report"
)

// ── Data band iteration ───────────────────────────────────────────────────────

// RunDataBandRows iterates over the rows provided by rows (a slice of any values,
// one per row) and renders db for each one. This is the core iteration loop for
// data bands when a data source is not available (in-memory rows).
//
// It renders:
//   - DataHeader before the first row (if present)
//   - db for each row
//   - sub-bands (via RunBands) after each row
//   - DataFooter after the last row (if present)
func (e *ReportEngine) RunDataBandRows(db *band.DataBand, rows int) {
	if rows == 0 {
		// Nothing to print: show child if PrintIfDatabandEmpty.
		if child := db.Child(); child != nil {
			e.ShowFullBand(&child.BandBase)
		}
		return
	}

	headerShown := false
	for rowIdx := 0; rowIdx < rows; rowIdx++ {
		isFirst := rowIdx == 0
		isLast := rowIdx == rows-1

		db.SetRowNo(rowIdx + 1)
		db.SetAbsRowNo(e.absRowNo)
		e.absRowNo++
		db.SetIsFirstRow(isFirst)
		db.SetIsLastRow(isLast)

		// Show DataHeader on first row.
		if isFirst && !headerShown {
			if hdr := db.Header(); hdr != nil {
				e.ShowFullBand(&hdr.HeaderFooterBandBase.BandBase)
			}
			headerShown = true
		}

		// Start new page if configured (not on first row).
		if db.StartNewPage() && db.FlagUseStartNewPage && rowIdx > 0 {
			e.startNewPageForCurrent()
		}

		// Show the data band itself.
		e.ShowFullBand(&db.BandBase)

		// Run nested sub-bands.
		if err := e.runBands(dataBandSubBands(db)); err != nil {
			return
		}
	}

	// Show DataFooter after all rows.
	if headerShown {
		if ftr := db.Footer(); ftr != nil {
			e.ShowFullBand(&ftr.HeaderFooterBandBase.BandBase)
		}
	}
}

// RunDataBandFull runs a DataBand with a data source.
// It calls ds.First(), iterates all rows (up to db.MaxRows()), shows header/
// data/footer bands, and leaves the data source positioned after the last row.
//
// This is the primary entry point used by RunBands when a DataBand has a DataSource.
func (e *ReportEngine) RunDataBandFull(db *band.DataBand) error {
	ds := db.DataSourceRef()
	if ds == nil {
		// No data source — render the band once as a static (no-iteration) band.
		e.ShowFullBand(&db.BandBase)
		return nil
	}

	// Apply sort specs to in-memory data sources before iterating.
	if sortSpecs := db.Sort(); len(sortSpecs) > 0 {
		if sortable, ok := ds.(data.Sortable); ok {
			specs := make([]data.SortSpec, len(sortSpecs))
			for i, s := range sortSpecs {
				specs[i] = data.SortSpec{
					Column:     s.Column,
					Descending: s.Order == band.SortOrderDescending,
				}
			}
			sortable.SortRows(specs)
		}
	}

	if err := ds.First(); err != nil {
		return err
	}

	maxRows := db.MaxRows()
	total := ds.RowCount()
	if maxRows > 0 && total > maxRows {
		total = maxRows
	}

	headerShown := false
	rowNo := 0

	for rowNo < total && !ds.EOF() {
		// Evaluate filter expression; skip row when it evaluates to false.
		if !e.evalBandFilter(db) {
			if err := ds.Next(); err != nil {
				break
			}
			continue
		}

		isFirst := rowNo == 0
		isLast := rowNo == total-1

		rowNo++
		e.rowNo = rowNo
		db.SetRowNo(rowNo)
		db.SetAbsRowNo(e.absRowNo)
		e.absRowNo++
		db.SetIsFirstRow(isFirst)
		db.SetIsLastRow(isLast)
		e.syncRowVariables()

		// Inject the current data source row into the report's Calc evaluator.
		// band.DataSource is a subset of data.DataSource; the concrete type
		// (*data.BaseDataSource) satisfies data.DataSource.
		if e.report != nil {
			if fullDS, ok := ds.(data.DataSource); ok {
				e.report.SetCalcContext(fullDS)
			}
		}

		if isFirst && !headerShown {
			if hdr := db.Header(); hdr != nil {
				e.ShowFullBand(&hdr.HeaderFooterBandBase.BandBase)
			}
			headerShown = true
		}

		if db.StartNewPage() && db.FlagUseStartNewPage && rowNo > 1 {
			e.startNewPageForCurrent()
		}

		e.ShowFullBand(&db.BandBase)

		if err := e.runBands(dataBandSubBands(db)); err != nil {
			return err
		}

		if err := ds.Next(); err != nil {
			break // EOF
		}
		if e.aborted {
			break
		}
	}

	if headerShown {
		if ftr := db.Footer(); ftr != nil {
			e.ShowFullBand(&ftr.HeaderFooterBandBase.BandBase)
		}
	} else if db.PrintIfDSEmpty() {
		// Show a single empty row when the data source is empty.
		db.SetRowNo(1)
		db.SetIsFirstRow(true)
		db.SetIsLastRow(true)
		e.ShowFullBand(&db.BandBase)
	}

	return nil
}

// dataBandSubBands returns the sub-bands nested inside a DataBand's object collection.
// Sub-bands are child bands that are rendered for each row of the parent data band.
func dataBandSubBands(db *band.DataBand) []report.Base {
	var result []report.Base
	for i := 0; i < db.Objects().Len(); i++ {
		obj := db.Objects().Get(i)
		if isSubBand(obj) {
			result = append(result, obj)
		}
	}
	return result
}

// isSubBand returns true if obj is a band type that can be nested in a DataBand.
func isSubBand(obj report.Base) bool {
	switch obj.(type) {
	case *band.DataBand, *band.GroupHeaderBand, *band.GroupFooterBand,
		*band.ChildBand, *band.DataHeaderBand, *band.DataFooterBand:
		return true
	}
	return false
}
