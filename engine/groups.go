package engine

import (
	"fmt"

	"github.com/andrewloable/go-fastreport/band"
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

// groupConditionValue returns a string representing the current group condition
// value for the given GroupHeaderBand. The Condition field holds either a bare
// column name or a bracketed expression like "[DataSource.Field]". We strip
// brackets if present and query the data source directly.
func groupConditionValue(b *band.GroupHeaderBand, ds band.DataSource) string {
	cond := b.Condition()
	if cond == "" {
		return ""
	}
	// Strip surrounding brackets if any: [Column] → Column
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
// value, and updates the stored value.
func (gs *groupValueState) changed(b *band.GroupHeaderBand, ds band.DataSource) bool {
	v := groupConditionValue(b, ds)
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

	total := ds.RowCount()
	maxRows := db.MaxRows()
	if maxRows > 0 && total > maxRows {
		total = maxRows
	}

	gs := newGroupValueState()
	rowIdx := 0

	for rowIdx < total && !ds.EOF() {
		if rowIdx == 0 {
			initGroupItem(groupBand, root, rowIdx, gs, ds)
		} else {
			checkGroupItem(groupBand, root, rowIdx, gs, ds)
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
func initGroupItem(header *band.GroupHeaderBand, curItem *groupTreeItem, rowIdx int, gs *groupValueState, ds band.DataSource) {
	for header != nil {
		gs.reset(header)
		gs.changed(header, ds) // consume current value as the baseline

		child := &groupTreeItem{band: header, rowNo: rowIdx, rowCount: 1}
		curItem = curItem.addItem(child)
		header = header.NestedGroup()
	}
}

// checkGroupItem walks the nested group chain and finds the outermost level
// where the group condition value changed, then re-initialises from that level.
func checkGroupItem(header *band.GroupHeaderBand, curItem *groupTreeItem, rowIdx int, gs *groupValueState, ds band.DataSource) {
	for header != nil {
		if gs.changed(header, ds) {
			// Re-initialise this level and all nested levels.
			initGroupItem(header, curItem, rowIdx, gs, ds)
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
func (e *ReportEngine) showGroupTree(root *groupTreeItem) {
	if root.band != nil {
		e.showGroupHeader(root.band)
	}

	if len(root.items) == 0 {
		// Leaf node: render the associated data band rows.
		if root.band != nil && root.rowCount > 0 {
			db := root.band.Data()
			if db != nil {
				e.RunDataBandRows(db, root.rowCount)
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
func (e *ReportEngine) showDataHeader(header *band.GroupHeaderBand) {
	header.SetRowNo(0)
	db := header.Data()
	if db == nil {
		return
	}
	if hdr := db.Header(); hdr != nil {
		e.ShowFullBand(&hdr.HeaderFooterBandBase.BandBase)
	}
}

// showDataFooter renders the DataFooterBand of a GroupHeaderBand.
func (e *ReportEngine) showDataFooter(header *band.GroupHeaderBand) {
	db := header.Data()
	if db == nil {
		return
	}
	if ftr := db.Footer(); ftr != nil {
		e.ShowFullBand(&ftr.HeaderFooterBandBase.BandBase)
	}
}

// showGroupHeader renders the group header band and increments row counters.
func (e *ReportEngine) showGroupHeader(header *band.GroupHeaderBand) {
	header.SetAbsRowNo(header.AbsRowNo() + 1)
	header.SetRowNo(header.RowNo() + 1)
	e.ShowFullBand(&header.HeaderFooterBandBase.BandBase)
}

// showGroupFooter renders the group footer band.
func (e *ReportEngine) showGroupFooter(header *band.GroupHeaderBand) {
	ftr := header.GroupFooter()
	if ftr == nil {
		return
	}
	ftr.SetAbsRowNo(ftr.AbsRowNo() + 1)
	ftr.SetRowNo(ftr.RowNo() + 1)
	e.ShowFullBand(&ftr.HeaderFooterBandBase.BandBase)
}

// ── Public entry point ────────────────────────────────────────────────────────

// RunGroup is the top-level method called by the band runner when it encounters
// a GroupHeaderBand. It builds the group tree then renders it.
func (e *ReportEngine) RunGroup(groupBand *band.GroupHeaderBand) {
	db := groupBand.Data()
	if db == nil {
		return
	}
	if db.DataSourceRef() == nil {
		return
	}

	tree := e.makeGroupTree(groupBand)
	e.showGroupTree(tree)
}
