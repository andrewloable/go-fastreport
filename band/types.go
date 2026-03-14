package band

import "github.com/andrewloable/go-fastreport/report"

// -----------------------------------------------------------------------
// Thin band types that add no new fields
// -----------------------------------------------------------------------

// ReportTitleBand prints once at the start of a report.
// It is the Go equivalent of FastReport.ReportTitleBand.
type ReportTitleBand struct{ BandBase }

// NewReportTitleBand creates a ReportTitleBand with defaults.
func NewReportTitleBand() *ReportTitleBand {
	return &ReportTitleBand{BandBase: *NewBandBase()}
}

// ReportSummaryBand prints once at the end of a report.
// It is the Go equivalent of FastReport.ReportSummaryBand.
type ReportSummaryBand struct{ HeaderFooterBandBase }

// NewReportSummaryBand creates a ReportSummaryBand with defaults.
func NewReportSummaryBand() *ReportSummaryBand {
	return &ReportSummaryBand{HeaderFooterBandBase: *NewHeaderFooterBandBase()}
}

// PageHeaderBand prints at the top of each report page.
// It is the Go equivalent of FastReport.PageHeaderBand.
type PageHeaderBand struct{ BandBase }

// NewPageHeaderBand creates a PageHeaderBand with defaults.
func NewPageHeaderBand() *PageHeaderBand {
	return &PageHeaderBand{BandBase: *NewBandBase()}
}

// PageFooterBand prints at the bottom of each report page.
// It is the Go equivalent of FastReport.PageFooterBand.
type PageFooterBand struct{ BandBase }

// NewPageFooterBand creates a PageFooterBand with defaults.
func NewPageFooterBand() *PageFooterBand {
	return &PageFooterBand{BandBase: *NewBandBase()}
}

// ColumnHeaderBand prints at the top of each column in a multi-column layout.
// It is the Go equivalent of FastReport.ColumnHeaderBand.
type ColumnHeaderBand struct{ BandBase }

// NewColumnHeaderBand creates a ColumnHeaderBand with defaults.
func NewColumnHeaderBand() *ColumnHeaderBand {
	return &ColumnHeaderBand{BandBase: *NewBandBase()}
}

// ColumnFooterBand prints at the bottom of each column.
// It is the Go equivalent of FastReport.ColumnFooterBand.
type ColumnFooterBand struct{ BandBase }

// NewColumnFooterBand creates a ColumnFooterBand with defaults.
func NewColumnFooterBand() *ColumnFooterBand {
	return &ColumnFooterBand{BandBase: *NewBandBase()}
}

// DataHeaderBand prints before the data rows of a DataBand.
// It is the Go equivalent of FastReport.DataHeaderBand.
type DataHeaderBand struct{ HeaderFooterBandBase }

// NewDataHeaderBand creates a DataHeaderBand with defaults.
func NewDataHeaderBand() *DataHeaderBand {
	return &DataHeaderBand{HeaderFooterBandBase: *NewHeaderFooterBandBase()}
}

// DataFooterBand prints after the data rows of a DataBand.
// It is the Go equivalent of FastReport.DataFooterBand.
type DataFooterBand struct{ HeaderFooterBandBase }

// NewDataFooterBand creates a DataFooterBand with defaults.
func NewDataFooterBand() *DataFooterBand {
	return &DataFooterBand{HeaderFooterBandBase: *NewHeaderFooterBandBase()}
}

// GroupFooterBand prints at the end of a group.
// It is the Go equivalent of FastReport.GroupFooterBand.
type GroupFooterBand struct{ HeaderFooterBandBase }

// NewGroupFooterBand creates a GroupFooterBand with defaults.
func NewGroupFooterBand() *GroupFooterBand {
	return &GroupFooterBand{HeaderFooterBandBase: *NewHeaderFooterBandBase()}
}

// OverlayBand prints on top of all other band content on a page.
// It is the Go equivalent of FastReport.OverlayBand.
type OverlayBand struct{ BandBase }

// NewOverlayBand creates an OverlayBand with defaults.
func NewOverlayBand() *OverlayBand {
	return &OverlayBand{BandBase: *NewBandBase()}
}

// -----------------------------------------------------------------------
// SortOrder enum (used by GroupHeaderBand)
// -----------------------------------------------------------------------

// SortOrder controls how group data rows are sorted.
type SortOrder int

const (
	// SortOrderAscending sorts rows in ascending order (default).
	SortOrderAscending SortOrder = iota
	// SortOrderDescending sorts rows in descending order.
	SortOrderDescending
	// SortOrderNone leaves the row order unchanged.
	SortOrderNone
)

// -----------------------------------------------------------------------
// GroupHeaderBand
// -----------------------------------------------------------------------

// GroupHeaderBand prints at the start of each data group.
// It is the Go equivalent of FastReport.GroupHeaderBand.
type GroupHeaderBand struct {
	HeaderFooterBandBase

	nestedGroup     *GroupHeaderBand
	data            *DataBand
	groupFooter     *GroupFooterBand
	condition       string
	sortOrder       SortOrder // default SortOrderAscending
	keepTogether    bool
	resetPageNumber bool
}

// NewGroupHeaderBand creates a GroupHeaderBand with defaults.
func NewGroupHeaderBand() *GroupHeaderBand {
	return &GroupHeaderBand{
		HeaderFooterBandBase: *NewHeaderFooterBandBase(),
		sortOrder:            SortOrderAscending,
	}
}

// NestedGroup returns the inner (nested) group header band.
func (g *GroupHeaderBand) NestedGroup() *GroupHeaderBand { return g.nestedGroup }

// SetNestedGroup sets the nested group header.
func (g *GroupHeaderBand) SetNestedGroup(n *GroupHeaderBand) { g.nestedGroup = n }

// Data returns the DataBand associated with this group.
func (g *GroupHeaderBand) Data() *DataBand { return g.data }

// SetData sets the data band for this group.
func (g *GroupHeaderBand) SetData(d *DataBand) { g.data = d }

// GroupFooter returns the footer band for this group.
func (g *GroupHeaderBand) GroupFooter() *GroupFooterBand { return g.groupFooter }

// SetGroupFooter sets the group footer band.
func (g *GroupHeaderBand) SetGroupFooter(f *GroupFooterBand) { g.groupFooter = f }

// Condition returns the grouping expression (e.g. "[Orders.CustomerName]").
func (g *GroupHeaderBand) Condition() string { return g.condition }

// SetCondition sets the grouping expression.
func (g *GroupHeaderBand) SetCondition(s string) { g.condition = s }

// SortOrder returns how the group data rows are sorted.
func (g *GroupHeaderBand) SortOrder() SortOrder { return g.sortOrder }

// SetSortOrder sets the sort order.
func (g *GroupHeaderBand) SetSortOrder(o SortOrder) { g.sortOrder = o }

// KeepTogether returns whether the group header stays with its data on the same page.
func (g *GroupHeaderBand) KeepTogether() bool { return g.keepTogether }

// SetKeepTogether sets keep-together.
func (g *GroupHeaderBand) SetKeepTogether(v bool) { g.keepTogether = v }

// ResetPageNumber returns whether the page counter resets at each new group.
func (g *GroupHeaderBand) ResetPageNumber() bool { return g.resetPageNumber }

// SetResetPageNumber sets the reset-page-number flag.
func (g *GroupHeaderBand) SetResetPageNumber(v bool) { g.resetPageNumber = v }

// Serialize writes GroupHeaderBand properties that differ from defaults.
func (g *GroupHeaderBand) Serialize(w report.Writer) error {
	if err := g.HeaderFooterBandBase.Serialize(w); err != nil {
		return err
	}
	if g.condition != "" {
		w.WriteStr("Condition", g.condition)
	}
	if g.sortOrder != SortOrderAscending {
		w.WriteInt("SortOrder", int(g.sortOrder))
	}
	if g.keepTogether {
		w.WriteBool("KeepTogether", true)
	}
	if g.resetPageNumber {
		w.WriteBool("ResetPageNumber", true)
	}
	return nil
}

// Deserialize reads GroupHeaderBand properties.
func (g *GroupHeaderBand) Deserialize(r report.Reader) error {
	if err := g.HeaderFooterBandBase.Deserialize(r); err != nil {
		return err
	}
	g.condition = r.ReadStr("Condition", "")
	g.sortOrder = SortOrder(r.ReadInt("SortOrder", int(SortOrderAscending)))
	g.keepTogether = r.ReadBool("KeepTogether", false)
	g.resetPageNumber = r.ReadBool("ResetPageNumber", false)
	return nil
}

// -----------------------------------------------------------------------
// DataBand
// -----------------------------------------------------------------------

// DataBand repeats once per row of a connected data source.
// It is the Go equivalent of FastReport.DataBand.
type DataBand struct {
	BandBase

	header               *DataHeaderBand
	footer               *DataFooterBand
	columns              *BandColumns
	filter               string
	printIfDetailEmpty   bool
	printIfDSEmpty       bool
	keepTogether         bool
	keepDetail           bool
	idColumn             string
	parentIDColumn       string
	indent               float32
	keepSummary          bool
}

// NewDataBand creates a DataBand with defaults.
func NewDataBand() *DataBand {
	return &DataBand{
		BandBase: *NewBandBase(),
		columns:  NewBandColumns(),
	}
}

// Header returns the DataHeaderBand for this data band.
func (d *DataBand) Header() *DataHeaderBand { return d.header }

// SetHeader sets the header band.
func (d *DataBand) SetHeader(h *DataHeaderBand) { d.header = h }

// Footer returns the DataFooterBand for this data band.
func (d *DataBand) Footer() *DataFooterBand { return d.footer }

// SetFooter sets the footer band.
func (d *DataBand) SetFooter(f *DataFooterBand) { d.footer = f }

// Columns returns the multi-column layout settings.
func (d *DataBand) Columns() *BandColumns { return d.columns }

// Filter returns the row-filter expression.
func (d *DataBand) Filter() string { return d.filter }

// SetFilter sets the filter expression.
func (d *DataBand) SetFilter(s string) { d.filter = s }

// PrintIfDetailEmpty returns whether the band prints when its detail bands produce no rows.
func (d *DataBand) PrintIfDetailEmpty() bool { return d.printIfDetailEmpty }

// SetPrintIfDetailEmpty sets the print-if-detail-empty flag.
func (d *DataBand) SetPrintIfDetailEmpty(v bool) { d.printIfDetailEmpty = v }

// PrintIfDSEmpty returns whether the band prints when the data source is empty.
func (d *DataBand) PrintIfDSEmpty() bool { return d.printIfDSEmpty }

// SetPrintIfDSEmpty sets the print-if-datasource-empty flag.
func (d *DataBand) SetPrintIfDSEmpty(v bool) { d.printIfDSEmpty = v }

// KeepTogether returns whether all rows stay on the same page.
func (d *DataBand) KeepTogether() bool { return d.keepTogether }

// SetKeepTogether sets keep-together.
func (d *DataBand) SetKeepTogether(v bool) { d.keepTogether = v }

// KeepDetail returns whether the data band stays with its detail bands.
func (d *DataBand) KeepDetail() bool { return d.keepDetail }

// SetKeepDetail sets keep-detail.
func (d *DataBand) SetKeepDetail(v bool) { d.keepDetail = v }

// IDColumn returns the hierarchy ID column name (for tree-style data).
func (d *DataBand) IDColumn() string { return d.idColumn }

// SetIDColumn sets the ID column name.
func (d *DataBand) SetIDColumn(s string) { d.idColumn = s }

// ParentIDColumn returns the parent-ID column for hierarchical data.
func (d *DataBand) ParentIDColumn() string { return d.parentIDColumn }

// SetParentIDColumn sets the parent-ID column name.
func (d *DataBand) SetParentIDColumn(s string) { d.parentIDColumn = s }

// Indent returns the horizontal indentation for hierarchical rows (pixels).
func (d *DataBand) Indent() float32 { return d.indent }

// SetIndent sets the indentation.
func (d *DataBand) SetIndent(v float32) { d.indent = v }

// KeepSummary returns whether the data band stays with its footer band.
func (d *DataBand) KeepSummary() bool { return d.keepSummary }

// SetKeepSummary sets keep-summary.
func (d *DataBand) SetKeepSummary(v bool) { d.keepSummary = v }

// Serialize writes DataBand properties that differ from defaults.
func (d *DataBand) Serialize(w report.Writer) error {
	if err := d.BandBase.Serialize(w); err != nil {
		return err
	}
	if d.filter != "" {
		w.WriteStr("Filter", d.filter)
	}
	if d.printIfDetailEmpty {
		w.WriteBool("PrintIfDetailEmpty", true)
	}
	if d.printIfDSEmpty {
		w.WriteBool("PrintIfDatasourceEmpty", true)
	}
	if d.keepTogether {
		w.WriteBool("KeepTogether", true)
	}
	if d.keepDetail {
		w.WriteBool("KeepDetail", true)
	}
	if d.idColumn != "" {
		w.WriteStr("IdColumn", d.idColumn)
	}
	if d.parentIDColumn != "" {
		w.WriteStr("ParentIdColumn", d.parentIDColumn)
	}
	if d.indent != 0 {
		w.WriteFloat("Indent", d.indent)
	}
	if d.keepSummary {
		w.WriteBool("KeepSummary", true)
	}
	return nil
}

// Deserialize reads DataBand properties.
func (d *DataBand) Deserialize(r report.Reader) error {
	if err := d.BandBase.Deserialize(r); err != nil {
		return err
	}
	d.filter = r.ReadStr("Filter", "")
	d.printIfDetailEmpty = r.ReadBool("PrintIfDetailEmpty", false)
	d.printIfDSEmpty = r.ReadBool("PrintIfDatasourceEmpty", false)
	d.keepTogether = r.ReadBool("KeepTogether", false)
	d.keepDetail = r.ReadBool("KeepDetail", false)
	d.idColumn = r.ReadStr("IdColumn", "")
	d.parentIDColumn = r.ReadStr("ParentIdColumn", "")
	d.indent = r.ReadFloat("Indent", 0)
	d.keepSummary = r.ReadBool("KeepSummary", false)
	return nil
}
