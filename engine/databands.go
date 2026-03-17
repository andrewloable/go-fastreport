package engine

import (
	"fmt"

	"github.com/andrewloable/go-fastreport/band"
	"github.com/andrewloable/go-fastreport/data"
	"github.com/andrewloable/go-fastreport/report"
)

// ── Data band iteration ───────────────────────────────────────────────────────

// RunDataBandRows iterates over rows for a data band with default keep settings.
// Called from showGroupTree for leaf group nodes.
func (e *ReportEngine) RunDataBandRows(db *band.DataBand, rows int) {
	e.RunDataBandRowsKeep(db, rows, e.NeedKeepFirstRow(db), e.NeedKeepLastRow(db))
}

// RunDataBandRowsKeep is the inner iteration loop that mirrors
// C# RunDataBand(dataBand, rowCount, keepFirstRow, keepLastRow).
func (e *ReportEngine) RunDataBandRowsKeep(db *band.DataBand, rows int, keepFirstRow, keepLastRow bool) {
	if rows == 0 {
		// Nothing to print: show child if PrintIfDatabandEmpty.
		if child := db.Child(); child != nil && child.PrintIfDatabandEmpty {
			e.ShowFullBand(&child.BandBase)
		}
		return
	}

	// Set up multi-column state (nil when Columns.Count <= 1).
	cs := newDataBandColumnState(e, db)

	// Obtain the data source reference for per-row advancement and calc context.
	ds := db.DataSourceRef()

	isFirstRow := true
	someRowsPrinted := false
	db.SetRowNo(0)
	db.SetIsFirstRow(false)
	db.SetIsLastRow(false)

	// check if we have only one data row that should be kept with both header and footer
	oneRow := rows == 1 && keepFirstRow && keepLastRow

	for rowIdx := 0; rowIdx < rows; rowIdx++ {
		isLastRow := rowIdx == rows-1

		// Inject current DS row into the expression evaluator so text
		// expressions like [Orders.CustomerID] resolve correctly.
		if ds != nil && e.report != nil {
			if fullDS, ok := ds.(data.DataSource); ok {
				e.report.SetCalcContext(fullDS)
			}
		}

		db.SetRowNo(db.RowNo() + 1)
		db.SetAbsRowNo(e.absRowNo)
		e.absRowNo++
		db.SetIsFirstRow(isFirstRow)
		db.SetIsLastRow(isLastRow)

		// Accumulate aggregate totals for this row.
		e.accumulateTotals()

		// keep header
		if isFirstRow && keepFirstRow {
			e.startKeepBand(&db.BandBase)
		}

		// keep together
		if isFirstRow && db.KeepTogether() {
			e.startKeepBand(&db.BandBase)
		}

		// keep detail
		if db.KeepDetail() {
			e.startKeepBand(&db.BandBase)
		}

		// show header
		if isFirstRow {
			e.showDataBandHeader(db)
			if cs != nil {
				cs.rowY = e.curY
			}
		}

		// keep footer
		if isLastRow && keepLastRow && db.IsDeepmostDataBand() {
			e.startKeepBand(&db.BandBase)
		}

		// start block event
		if isFirstRow {
			e.OnStateChanged(db, EngineStateBlockStarted)
		}

		// StartNewPage: start a new page before non-first rows.
		if db.StartNewPage() && db.FlagUseStartNewPage && db.RowNo() > 1 {
			e.startNewPageForCurrent()
			if cs != nil {
				cs.rowY = e.curY
			}
		}

		// Show the data band itself (C# ShowDataBand).
		e.showDataBandBody(db, rows, cs)

		// end keep header
		if isFirstRow && keepFirstRow && !oneRow {
			e.EndKeep()
		}

		// end keep footer
		if isLastRow && keepLastRow && db.IsDeepmostDataBand() {
			e.CheckKeepFooter(db)
		}

		// Run nested sub-bands (with relation filtering for master-detail).
		subBands := dataBandSubBands(db)
		restore := e.applyRelationFilters(db, subBands)
		if err := e.runBands(subBands); err != nil {
			restore()
			e.flushColumnRow(cs)
			if db.KeepDetail() {
				e.EndKeep()
			}
			return
		}
		restore()

		// up the outline
		e.OutlineUp()

		// end keep detail
		if db.KeepDetail() {
			e.EndKeep()
		}

		isFirstRow = false
		someRowsPrinted = true

		// Advance the data source to the next row for the next iteration.
		if ds != nil {
			_ = ds.Next()
		}
		if e.aborted {
			break
		}
	}

	// Flush any partially-filled column row.
	e.flushColumnRow(cs)

	// CompleteToNRows: fill missing child rows (C# child.CompleteToNRows > rowCount).
	if child := db.Child(); child != nil && child.CompleteToNRows > rows {
		for i := 0; i < child.CompleteToNRows-rows; i++ {
			child.SetRowNo(rows + i + 1)
			child.SetAbsRowNo(rows + i + 1)
			e.ShowFullBand(&child.BandBase)
		}
	}

	// Print child if databand is empty.
	if child := db.Child(); child != nil && child.PrintIfDatabandEmpty && db.IsDatasourceEmpty() {
		e.ShowFullBand(&child.BandBase)
	}

	if someRowsPrinted {
		// Finish block event.
		e.OnStateChanged(db, EngineStateBlockFinished)

		// Show footer.
		e.showDataBandFooter(db)

		// end KeepTogether
		if db.KeepTogether() {
			e.EndKeep()
		}

		// end KeepLastRow
		if keepLastRow {
			e.EndKeep()
		}
	}
}

// RunDataBandFull runs a DataBand with a data source.
// It calls ds.First(), iterates all rows (up to db.MaxRows()), shows header/
// data/footer bands, and leaves the data source positioned before the last row.
//
// This is the primary entry point used by RunBands when a DataBand has a DataSource.
func (e *ReportEngine) RunDataBandFull(db *band.DataBand) error {
	// Resolve data source from Dictionary by alias if not already bound directly.
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

	// C# RunDataBand (outer): dataBand.InitDataSource(); dataBand.DataSource.First();
	if err := ds.First(); err != nil {
		if err != data.ErrEOF {
			return err
		}
		// Empty data source.
		if db.PrintIfDSEmpty() {
			db.SetRowNo(1)
			db.SetIsFirstRow(true)
			db.SetIsLastRow(true)
			e.ShowFullBand(&db.BandBase)
		} else if child := db.Child(); child != nil && child.PrintIfDatabandEmpty {
			e.ShowFullBand(&child.BandBase)
		}
		return nil
	}

	// If IdColumn and ParentIdColumn are both set, use hierarchical rendering.
	if db.IDColumn() != "" && db.ParentIDColumn() != "" {
		return e.runDataBandHierarchical(db, ds)
	}

	// C# RunDataBand (outer): compute rowCount.
	rowCount := ds.RowCount()
	if db.IsDatasourceEmpty() && db.PrintIfDSEmpty() {
		rowCount = 1
	}
	if db.CollectChildRows() && rowCount > 1 {
		rowCount = 1
	}
	maxRows := db.MaxRows()
	if maxRows > 0 && rowCount > maxRows {
		rowCount = maxRows
	}

	keepFirstRow := e.NeedKeepFirstRow(db)
	keepLastRow := e.NeedKeepLastRow(db)

	// C# RunDataBand (inner): the main iteration loop.
	cs := newDataBandColumnState(e, db)

	isFirstRow := true
	someRowsPrinted := false
	db.SetRowNo(0)
	db.SetIsFirstRow(false)
	db.SetIsLastRow(false)

	// check if we have only one data row that should be kept with both header and footer
	oneRow := rowCount == 1 && keepFirstRow && keepLastRow

	rowNo := 0
	for rowNo < rowCount && !ds.EOF() {
		// Evaluate filter expression; skip row when it evaluates to false.
		if !e.evalBandFilter(db) {
			if err := ds.Next(); err != nil {
				break
			}
			continue
		}

		isLastRow := rowNo == rowCount-1

		db.SetRowNo(db.RowNo() + 1)
		db.SetAbsRowNo(e.absRowNo)
		e.absRowNo++
		db.SetIsFirstRow(isFirstRow)
		db.SetIsLastRow(isLastRow)
		rowNo++
		e.rowNo = db.RowNo()
		e.syncRowVariables()

		// Inject the current data source row into the report's Calc evaluator.
		if e.report != nil {
			if fullDS, ok := ds.(data.DataSource); ok {
				e.report.SetCalcContext(fullDS)
			}
		}
		// Accumulate aggregate totals for this row.
		e.accumulateTotals()

		// keep header
		if isFirstRow && keepFirstRow {
			e.startKeepBand(&db.BandBase)
		}

		// keep together
		if isFirstRow && db.KeepTogether() {
			e.startKeepBand(&db.BandBase)
		}

		// keep detail
		if db.KeepDetail() {
			e.startKeepBand(&db.BandBase)
		}

		// show header
		if isFirstRow {
			e.showDataBandHeader(db)
			if cs != nil {
				cs.rowY = e.curY
			}
		}

		// keep footer
		if isLastRow && keepLastRow && db.IsDeepmostDataBand() {
			e.startKeepBand(&db.BandBase)
		}

		// start block event
		if isFirstRow {
			e.OnStateChanged(db, EngineStateBlockStarted)
		}

		if db.StartNewPage() && db.FlagUseStartNewPage && db.RowNo() > 1 {
			e.startNewPageForCurrent()
			if cs != nil {
				cs.rowY = e.curY
			}
		}

		// Show the data band (C# ShowDataBand).
		e.showDataBandBody(db, rowCount, cs)

		// end keep header
		if isFirstRow && keepFirstRow && !oneRow {
			e.EndKeep()
		}

		// end keep footer
		if isLastRow && keepLastRow && db.IsDeepmostDataBand() {
			e.CheckKeepFooter(db)
		}

		// Run nested sub-bands (C# RunBands(dataBand.Bands)).
		subBands := dataBandSubBands(db)
		restore := e.applyRelationFilters(db, subBands)
		if err := e.runBands(subBands); err != nil {
			restore()
			e.flushColumnRow(cs)
			if db.KeepDetail() {
				e.EndKeep()
			}
			return err
		}
		restore()

		// Up the outline.
		e.OutlineUp()

		// End keep detail.
		if db.KeepDetail() {
			e.EndKeep()
		}

		isFirstRow = false
		someRowsPrinted = true

		// Multi-column: break after first row.
		if db.Columns().Count() > 1 {
			break
		}

		if err := ds.Next(); err != nil {
			break // EOF
		}
		if e.aborted {
			break
		}
	}

	// Flush any partially-filled column row.
	e.flushColumnRow(cs)

	// CompleteToNRows: fill missing child rows (C# child.CompleteToNRows > rowCount).
	if child := db.Child(); child != nil && child.CompleteToNRows > rowCount {
		for i := 0; i < child.CompleteToNRows-rowCount; i++ {
			child.SetRowNo(rowCount + i + 1)
			child.SetAbsRowNo(rowCount + i + 1)
			e.ShowFullBand(&child.BandBase)
		}
	}

	// Print child if databand is empty.
	if child := db.Child(); child != nil && child.PrintIfDatabandEmpty && db.IsDatasourceEmpty() {
		e.ShowFullBand(&child.BandBase)
	}

	if someRowsPrinted {
		// Finish block event.
		e.OnStateChanged(db, EngineStateBlockFinished)

		// Show footer (C# ShowDataFooter).
		e.showDataBandFooter(db)

		// End KeepTogether.
		if db.KeepTogether() {
			e.EndKeep()
		}

		// End KeepLastRow.
		if keepLastRow {
			e.EndKeep()
		}
	} else {
		if db.PrintIfDSEmpty() {
			db.SetRowNo(1)
			db.SetIsFirstRow(true)
			db.SetIsLastRow(true)
			e.ShowFullBand(&db.BandBase)
		}
	}

	// FillUnusedSpace: repeat the child band to fill remaining page space.
	if child := db.Child(); child != nil && child.FillUnusedSpace {
		for e.freeSpace > 0 && child.Height() > 0 && e.freeSpace >= child.Height() {
			e.ShowFullBand(&child.BandBase)
		}
	}

	// C# RunDataBand (outer): do not leave the datasource in EOF state.
	type hasPrior interface{ Prior() }
	if p, ok := ds.(hasPrior); ok {
		p.Prior()
	}

	return nil
}

// showDataBandHeader mirrors C# ShowDataHeader.
func (e *ReportEngine) showDataBandHeader(db *band.DataBand) {
	hdr := db.Header()
	if hdr != nil {
		e.ShowFullBand(&hdr.HeaderFooterBandBase.BandBase)
		if hdr.RepeatOnEveryPage() {
			e.AddReprintDataHeader(hdr)
		}
	}
	// C# also registers the footer for reprint when it has RepeatOnEveryPage.
	ftr := db.Footer()
	if ftr != nil && ftr.RepeatOnEveryPage() {
		e.AddReprintDataFooter(ftr)
	}
}

// showDataBandFooter mirrors C# ShowDataFooter.
func (e *ReportEngine) showDataBandFooter(db *band.DataBand) {
	ds := db.DataSourceRef()

	// C# ShowDataFooter: dataBand.DataSource.Prior();
	type hasPrior interface{ Prior() }
	if ds != nil {
		if p, ok := ds.(hasPrior); ok {
			p.Prior()
		}
	}

	ftr := db.Footer()
	e.RemoveReprint(footerBand(ftr))
	if ftr != nil {
		e.ShowFullBand(&ftr.HeaderFooterBandBase.BandBase)
	}
	if hdr := db.Header(); hdr != nil {
		e.RemoveReprint(&hdr.HeaderFooterBandBase.BandBase)
	}

	// C# ShowDataFooter: dataBand.DataSource.Next();
	if ds != nil {
		_ = ds.Next()
	}
}

// footerBand returns the BandBase pointer for a DataFooterBand, or nil.
func footerBand(ftr *band.DataFooterBand) *band.BandBase {
	if ftr == nil {
		return nil
	}
	return &ftr.HeaderFooterBandBase.BandBase
}

// showDataBandBody mirrors C# ShowDataBand.
func (e *ReportEngine) showDataBandBody(db *band.DataBand, rowCount int, cs *dataBandColumnState) {
	if db.Columns().Count() > 1 {
		db.SetWidth(db.Columns().ActualWidth())
		e.showBandInColumn(db, cs)
	} else {
		// ResetPageNumber handling.
		if db.ResetPageNumber() && (db.FirstRowStartsNewPage() || db.RowNo() > 1) {
			e.ResetLogicalPageNumber()
		}
		if cs != nil {
			e.showBandInColumn(db, cs)
		} else {
			e.ShowFullBand(&db.BandBase)
		}
	}
}

// runDataBandHierarchical renders a DataBand with IdColumn/ParentIdColumn set.
func (e *ReportEngine) runDataBandHierarchical(db *band.DataBand, ds band.DataSource) error {
	idCol := db.IDColumn()
	parentCol := db.ParentIDColumn()

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

	// Build parent->children map.
	children := make(map[string][]int) // parentID -> slice of row indices in rows[]
	for i := range rows {
		children[rows[i].parentID] = append(children[rows[i].parentID], i)
	}

	// Determine root rows.
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

	if err := ds.First(); err != nil {
		return err
	}

	headerShown := false
	prevLevel := e.hierarchyLevel
	prevRowNo := e.hierarchyRowNo

	var renderRows func(indices []int, level int, prefix string) error
	renderRows = func(indices []int, level int, prefix string) error {
		for nth, idx := range indices {
			row := rows[idx]

			if err := ds.First(); err != nil {
				return err
			}
			for k := 0; k < row.idx; k++ {
				if err := ds.Next(); err != nil {
					break
				}
			}

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
	case *band.DataBand, *band.GroupHeaderBand:
		return true
	}
	return false
}
