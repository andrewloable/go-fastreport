package engine

import (
	"fmt"

	"github.com/andrewloable/go-fastreport/band"
	"github.com/andrewloable/go-fastreport/data"
	"github.com/andrewloable/go-fastreport/report"
)

// applyRelationFilters inspects the report Dictionary for master-detail
// Relations between the parent DataBand's data source and each child DataBand
// in subBands. For each matched relation it wraps the child data source in a
// FilteredDataSource that only exposes rows where the join-key columns match
// the current parent row values.
//
// It returns a restore function that callers must invoke after rendering the
// sub-bands to reinstate the original (unfiltered) data sources.
func (e *ReportEngine) applyRelationFilters(parent *band.DataBand, subBands []report.Base) func() {
	if e.report == nil {
		return func() {}
	}
	dict := e.report.Dictionary()
	if dict == nil || len(dict.Relations()) == 0 {
		return func() {}
	}

	// The parent band's data source must satisfy data.DataSource so we can
	// call GetValue() to read the current row's join-key values.
	parentFullDS, ok := parent.DataSourceRef().(data.DataSource)
	if !ok {
		return func() {}
	}

	var restores []func()

	for _, sb := range subBands {
		childDB, ok := sb.(*band.DataBand)
		if !ok {
			continue
		}

		// Resolve child data source.
		childDS := childDB.DataSourceRef()
		if childDS == nil {
			// Try resolving by alias (same path as RunDataBandFull).
			if alias := childDB.DataSourceAlias(); alias != "" {
				if resolved := dict.FindDataSourceByAlias(alias); resolved != nil {
					if bds, ok2 := resolved.(band.DataSource); ok2 {
						childDS = bds
					}
				}
			}
		}
		if childDS == nil {
			continue
		}

		childFullDS, ok := childDS.(data.DataSource)
		if !ok {
			continue
		}

		// Find a relation linking parent → child data sources.
		rel := data.FindRelation(dict, parentFullDS, childFullDS)
		if rel == nil {
			continue
		}
		if len(rel.ParentColumns) == 0 {
			continue
		}

		// Read parent join-key values from the current parent row.
		parentVals := make([]string, len(rel.ParentColumns))
		for i, col := range rel.ParentColumns {
			v, _ := parentFullDS.GetValue(col)
			parentVals[i] = fmt.Sprintf("%v", v)
		}

		// Build a filtered view of the child data source.
		filtered, err := data.NewFilteredDataSource(childFullDS, rel.ChildColumns, parentVals)
		if err != nil {
			continue
		}

		// Swap the child band's data source; capture original for restore.
		orig := childDS
		childDB.SetDataSource(filtered)

		restores = append(restores, func() {
			childDB.SetDataSource(orig)
		})
	}

	return func() {
		for _, r := range restores {
			r()
		}
	}
}
