package engine

import (
	"fmt"

	"github.com/andrewloable/go-fastreport/band"
	"github.com/andrewloable/go-fastreport/data"
)

// groupTreeItem represents one node in the group tree built before rendering.
// Each node corresponds to one instance of a GroupHeaderBand with its sub-items
// (nested groups or data rows).
type groupTreeItem struct {
	band     *band.GroupHeaderBand // nil for the synthetic root
	items    []*groupTreeItem
	rowNo    int // index of first data row in this group (0-based)
	rowCount int // number of data rows belonging to this group node
}

func (g *groupTreeItem) firstItem() *groupTreeItem {
	if len(g.items) == 0 {
		return nil
	}
	return g.items[0]
}

func (g *groupTreeItem) lastItem() *groupTreeItem {
	if len(g.items) == 0 {
		return nil
	}
	return g.items[len(g.items)-1]
}

func (g *groupTreeItem) addItem(item *groupTreeItem) *groupTreeItem {
	g.items = append(g.items, item)
	return item
}

// ── Group value tracking ──────────────────────────────────────────────────────

// groupConditionValue evaluates the group condition expression for the given
// GroupHeaderBand. Mirrors C# GetGroupConditionValue which calls Report.Calc,
// supporting complex expressions like [Year([Orders.OrderDate])].
// Falls back to direct ds.GetValue for simple column references when the
// data source doesn't expose columns to the expression environment.
func (e *ReportEngine) groupConditionValue(b *band.GroupHeaderBand, ds band.DataSource) string {
	cond := b.Condition()
	if cond == "" {
		return ""
	}
	// Primary: use Report.Calc for full expression support (mirrors C#).
	if e.report != nil {
		v, err := e.report.Calc(cond)
		if err == nil && v != nil {
			return fmt.Sprintf("%v", v)
		}
	}
	// Fallback: strip brackets and query the data source directly.
	col := cond
	if len(col) > 2 && col[0] == '[' && col[len(col)-1] == ']' {
		col = col[1 : len(col)-1]
	}
	v, err := ds.GetValue(col)
	if err != nil || v == nil {
		return ""
	}
	return fmt.Sprintf("%v", v)
}

// groupValueState tracks the last seen condition value for each GroupHeaderBand.
type groupValueState struct {
	lastValue map[*band.GroupHeaderBand]string
	hasValue  map[*band.GroupHeaderBand]bool
}

func newGroupValueState() *groupValueState {
	return &groupValueState{
		lastValue: make(map[*band.GroupHeaderBand]string),
		hasValue:  make(map[*band.GroupHeaderBand]bool),
	}
}

func (gs *groupValueState) reset(b *band.GroupHeaderBand) {
	gs.hasValue[b] = false
	gs.lastValue[b] = ""
}

// changed returns true if the current value for b differs from the last seen
// value, and updates the stored value. calcFn is e.groupConditionValue.
func (gs *groupValueState) changed(b *band.GroupHeaderBand, calcFn func(*band.GroupHeaderBand) string) bool {
	v := calcFn(b)
	if !gs.hasValue[b] || gs.lastValue[b] != v {
		gs.lastValue[b] = v
		gs.hasValue[b] = true
		return true
	}
	return false
}

// ── Tree construction ─────────────────────────────────────────────────────────

// makeGroupTree iterates the data source attached to groupBand and builds a
// groupTreeItem hierarchy. The root item's band is nil (synthetic root).
func (e *ReportEngine) makeGroupTree(groupBand *band.GroupHeaderBand) *groupTreeItem {
	root := &groupTreeItem{}

	db := groupBand.Data()
	if db == nil {
		return root
	}
	ds := db.DataSourceRef()
	if ds == nil {
		return root
	}

	if err := ds.First(); err != nil {
		return root
	}

	// C# MakeGroupTree: iterate ALL rows with no MaxRows limit.
	// MaxRows is applied later in showGroupTree when rendering each leaf's rows.
	total := ds.RowCount()

	gs := newGroupValueState()
	rowIdx := 0

	// calcFn captures ds so groupConditionValue can fall back to direct
	// column access when the data source doesn't implement data.DataSource.
	calcFn := func(b *band.GroupHeaderBand) string {
		return e.groupConditionValue(b, ds)
	}

	for rowIdx < total && !ds.EOF() {
		// Update calc context per-row so Report.Calc evaluates complex
		// expressions like [Year([Orders.OrderDate])] against the current row.
		if e.report != nil {
			if fullDS, ok := ds.(data.DataSource); ok {
				e.report.SetCalcContext(fullDS)
			}
		}
		if rowIdx == 0 {
			initGroupItem(groupBand, root, rowIdx, gs, calcFn)
		} else {
			checkGroupItem(groupBand, root, rowIdx, gs, calcFn)
		}
		if err := ds.Next(); err != nil {
			break
		}
		rowIdx++
		if e.aborted {
			break
		}
	}

	return root
}

// initGroupItem adds new tree items for all nested group levels starting from
// header, anchored at curItem. Used on the first data row or after an outermost
// group value change.
// Also resets AbsRowNo and RowNo to 0 on each header band, matching C# InitGroupItem
// lines 172-173: header.AbsRowNo = 0; header.RowNo = 0.
func initGroupItem(header *band.GroupHeaderBand, curItem *groupTreeItem, rowIdx int, gs *groupValueState, calcFn func(*band.GroupHeaderBand) string) {
	for header != nil {
		gs.reset(header)
		gs.changed(header, calcFn) // consume current value as the baseline

		// C# InitGroupItem lines 172-173: reset per-group counters.
		header.SetAbsRowNo(0)
		header.SetRowNo(0)

		child := &groupTreeItem{band: header, rowNo: rowIdx, rowCount: 1}
		curItem = curItem.addItem(child)
		header = header.NestedGroup()
	}
}

// checkGroupItem walks the nested group chain and finds the outermost level
// where the group condition value changed, then re-initialises from that level.
func checkGroupItem(header *band.GroupHeaderBand, curItem *groupTreeItem, rowIdx int, gs *groupValueState, calcFn func(*band.GroupHeaderBand) string) {
	for header != nil {
		if gs.changed(header, calcFn) {
			// Re-initialise this level and all nested levels.
			initGroupItem(header, curItem, rowIdx, gs, calcFn)
			return
		}
		// Same group: increment row count on the current last child.
		last := curItem.lastItem()
		if last != nil {
			last.rowCount++
		}
		header = header.NestedGroup()
		curItem = last
	}
}

// ── Tree rendering ────────────────────────────────────────────────────────────

// showGroupTree recursively renders a group tree node:
//  1. ShowGroupHeader (unless root)
//  2. If leaf → RunDataBandRows for its rows; else ShowDataHeader + nested items + ShowDataFooter
//  3. ShowGroupFooter (unless root)
//
// groupStack is maintained so that getAllFooters can walk ancestor
// GroupHeaderBands when computing keep-with-data footer heights.
func (e *ReportEngine) showGroupTree(root *groupTreeItem) {
	// Push this group header onto the ancestor stack (innermost at front).
	// getAllFooters walks e.groupStack to find GroupFooterBands that must be
	// kept with the last data row, mirroring C# EnumHeaders + GetAllFooters.
	if root.band != nil {
		e.groupStack = append([]*band.GroupHeaderBand{root.band}, e.groupStack...)
		defer func() {
			e.groupStack = e.groupStack[1:]
		}()
	}

	if root.band != nil {
		// C# ShowGroupTree line 226: position the data source to the correct row
		// so group header expressions evaluate against the right data.
		if db := root.band.Data(); db != nil {
			if ds := db.DataSourceRef(); ds != nil {
				type rowPositioner interface{ SetCurrentRowNo(int) }
				if rp, ok := ds.(rowPositioner); ok {
					rp.SetCurrentRowNo(root.rowNo)
				}
				// Update the calc context so expressions in the group header see
				// the correct data row.
				if e.report != nil {
					if fullDS, ok := ds.(data.DataSource); ok {
						e.report.SetCalcContext(fullDS)
					}
				}
			}
		}
		e.showGroupHeader(root.band)
	}

	if len(root.items) == 0 {
		// Leaf node: render the associated data band rows.
		// Mirrors C#: RunDataBand(root.Band.GroupDataBand, rowCount, keepFirstRow, keepLastRow)
		if root.band != nil && root.rowCount > 0 {
			db := root.band.Data()
			if db != nil {
				rowCount := root.rowCount
				maxRows := db.MaxRows()
				if maxRows > 0 && rowCount > maxRows {
					rowCount = maxRows
				}
				keepFirstRow := e.NeedKeepFirstRowGroup(root.band)
				keepLastRow := e.NeedKeepLastRow(db)
				e.RunDataBandRowsKeep(db, rowCount, keepFirstRow, keepLastRow)
			}
		}
	} else {
		// Non-leaf: show data header for first nested group, iterate items,
		// then show data footer.
		first := root.firstItem()
		if first != nil && first.band != nil {
			e.showDataHeader(first.band)
		}

		for i, item := range root.items {
			item.band.SetIsFirstRow(i == 0)
			item.band.SetIsLastRow(i == len(root.items)-1)
			e.showGroupTree(item)
			if e.aborted {
				break
			}
		}

		if first != nil && first.band != nil {
			e.showDataFooter(first.band)
		}
	}

	if root.band != nil {
		e.showGroupFooter(root.band)
	}
}

// showDataHeader renders the DataHeaderBand of a GroupHeaderBand.
// Mirrors C# ShowDataHeader(GroupHeaderBand): shows header, registers reprints.
func (e *ReportEngine) showDataHeader(header *band.GroupHeaderBand) {
	header.SetRowNo(0)
	db := header.Data()
	if db == nil {
		return
	}
	if hdr := db.Header(); hdr != nil {
		e.ShowFullBand(&hdr.HeaderFooterBandBase.BandBase)
		if hdr.RepeatOnEveryPage() {
			e.AddReprint(&hdr.HeaderFooterBandBase.BandBase)
		}
	}
	if ftr := db.Footer(); ftr != nil {
		if ftr.RepeatOnEveryPage() {
			e.AddReprint(&ftr.HeaderFooterBandBase.BandBase)
		}
	}
}

// showDataFooter renders the DataFooterBand of a GroupHeaderBand.
// Mirrors C# ShowDataFooter(GroupHeaderBand): removes reprints.
func (e *ReportEngine) showDataFooter(header *band.GroupHeaderBand) {
	db := header.Data()
	if db == nil {
		return
	}
	if ftr := db.Footer(); ftr != nil {
		e.RemoveReprint(&ftr.HeaderFooterBandBase.BandBase)
	}
	if ftr := db.Footer(); ftr != nil {
		e.ShowFullBand(&ftr.HeaderFooterBandBase.BandBase)
	}
	if hdr := db.Header(); hdr != nil {
		e.RemoveReprint(&hdr.HeaderFooterBandBase.BandBase)
	}
}

// showGroupHeader renders the group header band and increments row counters.
// Mirrors C# ShowGroupHeader: handles KeepTogether, KeepWithData, RepeatOnEveryPage,
// GroupFooter RepeatOnEveryPage, and fires GroupStarted event.
func (e *ReportEngine) showGroupHeader(header *band.GroupHeaderBand) {
	header.SetAbsRowNo(header.AbsRowNo() + 1)
	header.SetRowNo(header.RowNo() + 1)

	// C#: if (header.ResetPageNumber && (header.FirstRowStartsNewPage || header.RowNo > 1))
	//         ResetLogicalPageNumber();
	if header.ResetPageNumber() && (header.FirstRowStartsNewPage() || header.RowNo() > 1) {
		e.ResetLogicalPageNumber()
	}

	// C#: if (header.KeepTogether) StartKeep(header);
	if header.KeepTogether() {
		e.startKeepBand(&header.HeaderFooterBandBase.BandBase)
	}
	// C#: if (header.KeepWithData) StartKeep(header.GroupDataBand);
	if header.KeepWithData() {
		if db := header.Data(); db != nil {
			e.startKeepBand(&db.BandBase)
		}
	}

	// C#: OnStateChanged(header, EngineState.GroupStarted);
	e.OnStateChanged(header, EngineStateGroupStarted)

	e.ShowFullBand(&header.HeaderFooterBandBase.BandBase)
	if header.RepeatOnEveryPage() {
		e.AddReprintGroupHeader(header)
	}

	// Register group footer for RepeatOnEveryPage.
	if ftr := header.GroupFooter(); ftr != nil {
		if ftr.RepeatOnEveryPage() {
			e.AddReprintGroupFooter(ftr)
		}
	}
}

// showGroupFooter renders the group footer band.
// Mirrors C# ShowGroupFooter: fires GroupFinished, handles reprint removal,
// and ends KeepTogether/KeepWithData.
// C# lines 143-158: calls dataSource.Prior() before the footer and
// dataSource.Next() after, so footer expressions see the last group row's data.
func (e *ReportEngine) showGroupFooter(header *band.GroupHeaderBand) {
	// C#: OnStateChanged(header, EngineState.GroupFinished);
	e.OnStateChanged(header, EngineStateGroupFinished)

	// C# ShowGroupFooter lines 143-145: rollback to previous data row so that
	// footer expressions print the last row's group condition value.
	var ds band.DataSource
	if db := header.Data(); db != nil {
		ds = db.DataSourceRef()
	}
	type hasPrior interface{ Prior() }
	if p, ok := ds.(hasPrior); ok {
		p.Prior()
		// Update calc context so footer expressions evaluate the prior row.
		if e.report != nil {
			if fullDS, ok := ds.(data.DataSource); ok {
				e.report.SetCalcContext(fullDS)
			}
		}
	}

	ftr := header.GroupFooter()
	if ftr != nil {
		ftr.SetAbsRowNo(ftr.AbsRowNo() + 1)
		ftr.SetRowNo(ftr.RowNo() + 1)
	}

	// C#: RemoveReprint(footer); ShowBand(footer); RemoveReprint(header);
	if ftr != nil {
		e.RemoveReprint(&ftr.HeaderFooterBandBase.BandBase)
	}
	if ftr != nil {
		e.ShowFullBand(&ftr.HeaderFooterBandBase.BandBase)
		// C# ShowBand calls ProcessTotals(band) after rendering every band.
		// ProcessTotals resets any total whose PrintOn==band && ResetAfterPrint==true.
		// Mirrors: TotalCollection.ProcessBand (TotalCollection.cs) called from
		// ReportEngine.Bands.cs ShowBand line 252.
		e.resetTotalsForBand(ftr.Name(), ftr.Repeated())
	}
	e.RemoveReprint(&header.HeaderFooterBandBase.BandBase)

	// C# ShowGroupFooter line 158: restore current row.
	if n, ok := ds.(interface{ Next() error }); ok {
		_ = n.Next()
	}

	// C#: OutlineUp(header);
	e.OutlineUp()

	// C#: if (header.KeepTogether) EndKeep();
	if header.KeepTogether() {
		e.EndKeep()
	}
	// C#: if (footer != null && footer.KeepWithData) EndKeep();
	if ftr != nil && ftr.KeepWithData() {
		e.EndKeep()
	}
}

// ── Public entry point ────────────────────────────────────────────────────────

// RunGroup is the top-level method called by the band runner when it encounters
// a GroupHeaderBand. It mirrors C# RunGroup:
//  1. Resolve DataSource
//  2. Apply group sort
//  3. ShowGroupTree(MakeGroupTree(groupBand))
//
// KeepTogether/KeepWithData are handled per-group-instance in showGroupHeader/showGroupFooter
// (not at the RunGroup level), matching the C# behaviour.
func (e *ReportEngine) RunGroup(groupBand *band.GroupHeaderBand) {
	db := groupBand.Data()

	// If the DataBand was added to Objects() during deserialization instead of
	// being set directly, find it there.
	if db == nil {
		for i := 0; i < groupBand.Objects().Len(); i++ {
			if candidate, ok := groupBand.Objects().Get(i).(*band.DataBand); ok {
				db = candidate
				groupBand.SetData(db)
				break
			}
		}
	}
	if db == nil {
		return
	}

	// Resolve DataSource from alias if not already bound.
	if db.DataSourceRef() == nil {
		if alias := db.DataSourceAlias(); alias != "" {
			if dict := e.report.Dictionary(); dict != nil {
				if resolved := dict.FindDataSourceByAlias(alias); resolved != nil {
					if bandDS, ok := resolved.(band.DataSource); ok {
						db.SetDataSource(bandDS)
					}
				}
			}
		}
	}
	if db.DataSourceRef() == nil {
		return
	}

	// If the GroupFooterBand was added to Objects() instead of being set
	// directly, find it there.
	if groupBand.GroupFooter() == nil {
		for i := 0; i < groupBand.Objects().Len(); i++ {
			if ftr, ok := groupBand.Objects().Get(i).(*band.GroupFooterBand); ok {
				groupBand.SetGroupFooter(ftr)
				break
			}
		}
	}

	// Sort the data source by the group condition(s) before tree scan.
	e.applyGroupSort(groupBand, db)

	tree := e.makeGroupTree(groupBand)
	// Reset the data source to the beginning so showGroupTree can iterate
	// rows in order (makeGroupTree consumed the DS cursor during pre-scan).
	if db.DataSourceRef() != nil {
		_ = db.DataSourceRef().First()
		// Re-inject the first row into the expression evaluator so that
		// showGroupHeader for the first group sees the correct data.
		if e.report != nil {
			if fullDS, ok := db.DataSourceRef().(data.DataSource); ok {
				e.report.SetCalcContext(fullDS)
			}
		}
	}
	e.showGroupTree(tree)

	// C# RunGroup line 281: do not leave the datasource in EOF state.
	// Mirrors: dataSource.Prior();
	ds := db.DataSourceRef()
	type hasPrior interface{ Prior() }
	if p, ok := ds.(hasPrior); ok {
		p.Prior()
	}
}

// applyGroupSort sorts the DataBand's data source using a combined key list:
// group-header condition columns first (for correct grouping), then the
// DataBand's own sort specs (for ordering within each group).
func (e *ReportEngine) applyGroupSort(groupBand *band.GroupHeaderBand, db *band.DataBand) {
	ds := db.DataSourceRef()
	if ds == nil {
		return
	}
	sortable, ok := ds.(data.Sortable)
	if !ok {
		return
	}

	var specs []data.SortSpec
	g := groupBand
	for g != nil {
		if g.SortOrder() != band.SortOrderNone {
			col := groupConditionColumn(g.Condition())
			if col != "" {
				specs = append(specs, data.SortSpec{
					Column:     col,
					Descending: g.SortOrder() == band.SortOrderDescending,
				})
			}
		}
		g = g.NestedGroup()
	}

	for _, s := range db.Sort() {
		expr := s.Expression
		if expr == "" {
			expr = s.Column
		}
		col := groupConditionColumn(expr)
		if col != "" {
			specs = append(specs, data.SortSpec{
				Column:     col,
				Descending: s.Order == band.SortOrderDescending,
			})
		}
	}

	if len(specs) > 0 {
		sortable.SortRows(specs)
	}
}

// groupConditionColumn extracts the bare column name from a group condition
// expression like "[Orders.CustomerID]" -> "CustomerID".
func groupConditionColumn(cond string) string {
	if len(cond) >= 2 && cond[0] == '[' && cond[len(cond)-1] == ']' {
		cond = cond[1 : len(cond)-1]
	}
	for i := len(cond) - 1; i >= 0; i-- {
		if cond[i] == '.' {
			return cond[i+1:]
		}
	}
	return cond
}
