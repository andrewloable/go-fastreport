package engine

import (
	"fmt"

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
	var lastRowPosition int
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
				headerPos := 0
				if e.preparedPages != nil {
					headerPos = e.preparedPages.CurPosition()
				}
				e.keepCurY = e.curY
				e.ShowFullBand(&hdr.HeaderFooterBandBase.BandBase)
				// KeepWithData: ensure header stays with first data row.
				e.checkKeepHeaderWithData(hdr, db, headerPos)
				if hdr.RepeatOnEveryPage() {
					e.AddReprintDataHeader(hdr)
				}
			}
			headerShown = true
		}

		// Start new page if configured (not on first row).
		if db.StartNewPage() && db.FlagUseStartNewPage && rowIdx > 0 {
			e.startNewPageForCurrent()
		}

		// Snapshot position before last row for KeepWithData footer check.
		if isLast && e.preparedPages != nil {
			lastRowPosition = e.preparedPages.CurPosition()
			e.keepCurY = e.curY
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
		if hdr := db.Header(); hdr != nil && hdr.RepeatOnEveryPage() {
			e.RemoveReprint(&hdr.HeaderFooterBandBase.BandBase)
		}
		if ftr := db.Footer(); ftr != nil {
			// KeepWithData: move last row + footer to new page if footer doesn't fit.
			e.checkKeepFooterWithData(ftr, lastRowPosition)
			e.ShowFullBand(&ftr.HeaderFooterBandBase.BandBase)
			if ftr.RepeatOnEveryPage() {
				e.AddReprintDataFooter(ftr)
			}
		}
	}
}

// RunDataBandFull runs a DataBand with a data source.
// It calls ds.First(), iterates all rows (up to db.MaxRows()), shows header/
// data/footer bands, and leaves the data source positioned after the last row.
//
// This is the primary entry point used by RunBands when a DataBand has a DataSource.
func (e *ReportEngine) RunDataBandFull(db *band.DataBand) error {
	// Resolve data source from Dictionary by alias if not already bound directly.
	// In FastReport .NET the FRX stores the DataSource attribute as an alias name
	// and the engine resolves it from the Dictionary at run time automatically.
	ds := db.DataSourceRef()
	if ds == nil {
		if alias := db.DataSourceAlias(); alias != "" {
			if dict := e.report.Dictionary(); dict != nil {
				if resolved := dict.FindDataSourceByAlias(alias); resolved != nil {
					if bandDS, ok := resolved.(band.DataSource); ok {
						db.SetDataSource(bandDS)
						ds = bandDS
					}
				}
			}
		}
	}
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

	// If IdColumn and ParentIdColumn are both set, use hierarchical rendering.
	if db.IDColumn() != "" && db.ParentIDColumn() != "" {
		return e.runDataBandHierarchical(db, ds)
	}

	maxRows := db.MaxRows()
	total := ds.RowCount()
	if maxRows > 0 && total > maxRows {
		total = maxRows
	}

	if db.KeepTogether() {
		e.StartKeep()
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
		// Accumulate aggregate totals for this row.
		e.accumulateTotals()

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

		subBands := dataBandSubBands(db)
		restore := e.applyRelationFilters(db, subBands)
		if err := e.runBands(subBands); err != nil {
			restore()
			return err
		}
		restore()

		if err := ds.Next(); err != nil {
			break // EOF
		}
		if e.aborted {
			break
		}
	}

	if db.KeepTogether() {
		e.EndKeep()
	}

	// Notify deferred objects waiting for DataFinished.
	e.OnStateChanged(db, EngineStateBlockFinished)

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

// runDataBandHierarchical renders a DataBand with IdColumn/ParentIdColumn set.
// It builds a parent→children map and renders rows recursively from the roots.
func (e *ReportEngine) runDataBandHierarchical(db *band.DataBand, ds band.DataSource) error {
	idCol := db.IDColumn()
	parentCol := db.ParentIDColumn()

	// Snapshot all rows so we can build the tree.
	type rowSnapshot struct {
		idx      int
		id       string
		parentID string
	}
	var rows []rowSnapshot

	fullDS, hasFull := ds.(data.DataSource)
	if !hasFull {
		return nil
	}

	for i := 0; !ds.EOF(); i++ {
		idVal, _ := fullDS.GetValue(idCol)
		parentVal, _ := fullDS.GetValue(parentCol)
		rows = append(rows, rowSnapshot{
			idx:      i,
			id:       fmt.Sprintf("%v", idVal),
			parentID: fmt.Sprintf("%v", parentVal),
		})
		if err := ds.Next(); err != nil {
			break
		}
	}

	// Build parent→children map.
	children := make(map[string][]int) // parentID → slice of row indices
	for i, row := range rows {
		children[row.parentID] = append(children[row.parentID], i)
	}

	// Determine root rows: those whose parentID is "", "0", or not found in idSet.
	idSet := make(map[string]bool, len(rows))
	for _, row := range rows {
		idSet[row.id] = true
	}
	var roots []int
	for i, row := range rows {
		if row.parentID == "" || row.parentID == "0" || !idSet[row.parentID] {
			roots = append(roots, i)
		}
	}

	// Reset data source to start.
	if err := ds.First(); err != nil {
		return err
	}

	headerShown := false
	prevLevel := e.hierarchyLevel
	prevRowNo := e.hierarchyRowNo

	// Recursive renderer.
	var renderRows func(indices []int, level int, prefix string) error
	renderRows = func(indices []int, level int, prefix string) error {
		for nth, idx := range indices {
			row := rows[idx]

			// Seek data source to this row.
			if err := ds.First(); err != nil {
				return err
			}
			for k := 0; k < row.idx; k++ {
				if err := ds.Next(); err != nil {
					break
				}
			}

			// Set hierarchy state.
			e.hierarchyLevel = level
			rowLabel := fmt.Sprintf("%d", nth+1)
			if prefix != "" {
				rowLabel = prefix + "." + rowLabel
			}
			e.hierarchyRowNo = rowLabel

			db.SetRowNo(row.idx + 1)
			db.SetAbsRowNo(e.absRowNo)
			e.absRowNo++

			if e.report != nil {
				e.report.SetCalcContext(fullDS)
			}
			e.accumulateTotals()

			if !headerShown {
				if hdr := db.Header(); hdr != nil {
					e.ShowFullBand(&hdr.HeaderFooterBandBase.BandBase)
				}
				headerShown = true
			}

			e.ShowFullBand(&db.BandBase)

			// Render children.
			if kidIndices, ok := children[row.id]; ok {
				if err := renderRows(kidIndices, level+1, rowLabel); err != nil {
					return err
				}
			}
		}
		return nil
	}

	if err := renderRows(roots, 0, ""); err != nil {
		return err
	}

	// Restore hierarchy state.
	e.hierarchyLevel = prevLevel
	e.hierarchyRowNo = prevRowNo

	if headerShown {
		if ftr := db.Footer(); ftr != nil {
			e.ShowFullBand(&ftr.HeaderFooterBandBase.BandBase)
		}
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
