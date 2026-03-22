package band

import (
	"strings"

	"github.com/andrewloable/go-fastreport/report"
)

// ── Sort support ─────────────────────────────────────────────────────────────

// SortSpec describes one sort column for a DataBand.
// The Order field uses the same SortOrder enum as GroupHeaderBand:
// SortOrderAscending (0), SortOrderDescending (1), SortOrderNone (2).
type SortSpec struct {
	// Column is the data-source column name.
	Column string
	// Order controls the direction of the sort (SortOrderAscending / SortOrderDescending).
	Order SortOrder
	// Expression is an optional expression (overrides Column when non-empty).
	Expression string
}

// sortSpecItem is a single sort item used for serialization.
// It is the Go equivalent of FastReport.Sort.
type sortSpecItem struct {
	Expression string
	Descending bool
}

func (s *sortSpecItem) TypeName() string { return "Sort" }

func (s *sortSpecItem) Serialize(w report.Writer) error {
	w.WriteStr("Expression", s.Expression)
	if s.Descending {
		w.WriteBool("Descending", true)
	}
	return nil
}

func (s *sortSpecItem) Deserialize(r report.Reader) error {
	s.Expression = r.ReadStr("Expression", "")
	s.Descending = r.ReadBool("Descending", false)
	return nil
}

// sortCollection is the serializable wrapper for a slice of SortSpec.
// It is the Go equivalent of FastReport.SortCollection.
type sortCollection struct {
	items []SortSpec
}

func (sc *sortCollection) TypeName() string { return "Sort" }

func (sc *sortCollection) Serialize(w report.Writer) error {
	for _, s := range sc.items {
		expr := s.Expression
		if expr == "" {
			expr = s.Column
		}
		item := &sortSpecItem{
			Expression: expr,
			Descending: s.Order == SortOrderDescending,
		}
		if err := w.WriteObject(item); err != nil {
			return err
		}
	}
	return nil
}

func (sc *sortCollection) Deserialize(r report.Reader) error {
	sc.items = sc.items[:0]
	for {
		ct, ok := r.NextChild()
		if !ok {
			break
		}
		if ct == "Sort" {
			var item sortSpecItem
			if err := item.Deserialize(r); err != nil {
				if ferr := r.FinishChild(); ferr != nil {
					return ferr
				}
				return err
			}
			if item.Expression != "" {
				spec := SortSpec{Column: item.Expression, Expression: item.Expression}
				if item.Descending {
					spec.Order = SortOrderDescending
				}
				sc.items = append(sc.items, spec)
			}
		}
		if err := r.FinishChild(); err != nil {
			return err
		}
	}
	return nil
}


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

// TypeName returns the FRX element name for this band.
func (*ReportTitleBand) TypeName() string { return "ReportTitle" }

// ReportSummaryBand prints once at the end of a report.
// It is the Go equivalent of FastReport.ReportSummaryBand.
type ReportSummaryBand struct{ HeaderFooterBandBase }

// NewReportSummaryBand creates a ReportSummaryBand with defaults.
func NewReportSummaryBand() *ReportSummaryBand {
	hf := NewHeaderFooterBandBase()
	hf.FlagIsColumnDependent = true // C# BandBase.IsColumnDependentBand (BandBase.cs line 589)
	return &ReportSummaryBand{HeaderFooterBandBase: *hf}
}

// TypeName returns the FRX element name for this band.
func (*ReportSummaryBand) TypeName() string { return "ReportSummary" }

// PageHeaderBand prints at the top of each report page.
// It is the Go equivalent of FastReport.PageHeaderBand.
type PageHeaderBand struct{ BandBase }

// NewPageHeaderBand creates a PageHeaderBand with defaults.
// C# PageHeaderBand() sets FlagUseStartNewPage=false — page-level bands never
// trigger StartNewPage logic (PageHeaderBand.cs constructor).
func NewPageHeaderBand() *PageHeaderBand {
	b := NewBandBase()
	b.FlagUseStartNewPage = false
	return &PageHeaderBand{BandBase: *b}
}

// TypeName returns the FRX element name for this band.
func (*PageHeaderBand) TypeName() string { return "PageHeader" }

// PageFooterBand prints at the bottom of each report page.
// It is the Go equivalent of FastReport.PageFooterBand.
type PageFooterBand struct{ BandBase }

// NewPageFooterBand creates a PageFooterBand with defaults.
// C# PageFooterBand() sets FlagUseStartNewPage=false — page-level bands never
// trigger StartNewPage logic (PageFooterBand.cs constructor).
func NewPageFooterBand() *PageFooterBand {
	b := NewBandBase()
	b.FlagUseStartNewPage = false
	return &PageFooterBand{BandBase: *b}
}

// TypeName returns the FRX element name for this band.
func (*PageFooterBand) TypeName() string { return "PageFooter" }

// ColumnHeaderBand prints at the top of each column in a multi-column layout.
// It is the Go equivalent of FastReport.ColumnHeaderBand.
type ColumnHeaderBand struct{ BandBase }

// NewColumnHeaderBand creates a ColumnHeaderBand with defaults.
// C# ColumnHeaderBand() sets FlagUseStartNewPage=false — column-level bands
// never trigger StartNewPage logic (ColumnHeaderBand.cs constructor).
func NewColumnHeaderBand() *ColumnHeaderBand {
	b := NewBandBase()
	b.FlagUseStartNewPage = false
	b.FlagIsColumnDependent = true // C# BandBase.IsColumnDependentBand (BandBase.cs line 590)
	return &ColumnHeaderBand{BandBase: *b}
}

// TypeName returns the FRX element name for this band.
func (*ColumnHeaderBand) TypeName() string { return "ColumnHeader" }

// ColumnFooterBand prints at the bottom of each column.
// It is the Go equivalent of FastReport.ColumnFooterBand.
type ColumnFooterBand struct{ BandBase }

// NewColumnFooterBand creates a ColumnFooterBand with defaults.
// C# ColumnFooterBand() sets FlagUseStartNewPage=false — column-level bands
// never trigger StartNewPage logic (ColumnFooterBand.cs constructor).
func NewColumnFooterBand() *ColumnFooterBand {
	b := NewBandBase()
	b.FlagUseStartNewPage = false
	b.FlagIsColumnDependent = true // C# BandBase.IsColumnDependentBand (BandBase.cs line 590)
	return &ColumnFooterBand{BandBase: *b}
}

// TypeName returns the FRX element name for this band.
func (*ColumnFooterBand) TypeName() string { return "ColumnFooter" }

// DataHeaderBand prints before the data rows of a DataBand.
// It is the Go equivalent of FastReport.DataHeaderBand.
type DataHeaderBand struct{ HeaderFooterBandBase }

// NewDataHeaderBand creates a DataHeaderBand with defaults.
func NewDataHeaderBand() *DataHeaderBand {
	hf := NewHeaderFooterBandBase()
	hf.FlagIsColumnDependent = true // C# BandBase.IsColumnDependentBand (BandBase.cs line 589)
	return &DataHeaderBand{HeaderFooterBandBase: *hf}
}

// TypeName returns the FRX element name for this band.
func (*DataHeaderBand) TypeName() string { return "DataHeader" }

// DataFooterBand prints after the data rows of a DataBand.
// It is the Go equivalent of FastReport.DataFooterBand.
type DataFooterBand struct{ HeaderFooterBandBase }

// NewDataFooterBand creates a DataFooterBand with defaults.
func NewDataFooterBand() *DataFooterBand {
	hf := NewHeaderFooterBandBase()
	hf.FlagIsColumnDependent = true // C# BandBase.IsColumnDependentBand (BandBase.cs line 589)
	return &DataFooterBand{HeaderFooterBandBase: *hf}
}

// TypeName returns the FRX element name for this band.
func (*DataFooterBand) TypeName() string { return "DataFooter" }

// GroupFooterBand prints at the end of a group.
// It is the Go equivalent of FastReport.GroupFooterBand.
type GroupFooterBand struct{ HeaderFooterBandBase }

// NewGroupFooterBand creates a GroupFooterBand with defaults.
func NewGroupFooterBand() *GroupFooterBand {
	hf := NewHeaderFooterBandBase()
	hf.FlagIsColumnDependent = true // C# BandBase.IsColumnDependentBand (BandBase.cs line 589)
	return &GroupFooterBand{HeaderFooterBandBase: *hf}
}

// TypeName returns the FRX element name for this band.
func (*GroupFooterBand) TypeName() string { return "GroupFooter" }

// OverlayBand prints on top of all other band content on a page.
// It is the Go equivalent of FastReport.OverlayBand.
type OverlayBand struct{ BandBase }

// NewOverlayBand creates an OverlayBand with defaults.
// C# OverlayBand() sets FlagUseStartNewPage=false — overlay bands never
// trigger StartNewPage logic (OverlayBand.cs constructor).
func NewOverlayBand() *OverlayBand {
	b := NewBandBase()
	b.FlagUseStartNewPage = false
	return &OverlayBand{BandBase: *b}
}

// TypeName returns the FRX element name for this band.
func (*OverlayBand) TypeName() string { return "Overlay" }

// -----------------------------------------------------------------------
// SortOrder enum (used by GroupHeaderBand)
// -----------------------------------------------------------------------

// SortOrder controls how group data rows are sorted.
type SortOrder int

const (
	// SortOrderAscending sorts rows in ascending order (default).
	// Equivalent to FastReport.SortOrder.Ascending.
	SortOrderAscending SortOrder = iota
	// SortOrderDescending sorts rows in descending order.
	// Equivalent to FastReport.SortOrder.Descending.
	SortOrderDescending
	// SortOrderNone leaves the row order unchanged.
	// Equivalent to FastReport.SortOrder.None.
	SortOrderNone
)

// sortOrderToString converts a SortOrder to the FRX attribute string name used
// by C# FastReport (Converter.ToString writes the enum name, e.g. "Ascending").
// C# SortOrder enum: None=0, Ascending=1, Descending=2.
// Go uses the same names but different iota values; only names matter for FRX.
func sortOrderToString(o SortOrder) string {
	switch o {
	case SortOrderDescending:
		return "Descending"
	case SortOrderNone:
		return "None"
	default:
		return "Ascending"
	}
}

// sortOrderFromString parses an FRX SortOrder attribute string name.
// Returns SortOrderAscending for an unrecognised value (matches C# default).
func sortOrderFromString(s string) SortOrder {
	switch s {
	case "Descending":
		return SortOrderDescending
	case "None":
		return SortOrderNone
	default:
		return SortOrderAscending
	}
}

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
	// header/footer are data-level header/footer bands attached to this group.
	// Mirrors C# GroupHeaderBand fields header/footer (GroupHeaderBand.cs lines 80-81).
	header          *DataHeaderBand
	footer          *DataFooterBand
	condition       string
	sortOrder       SortOrder // default SortOrderAscending
	keepTogether    bool
	resetPageNumber bool
	// groupValue holds the last evaluated condition value.
	// Used by ResetGroupValue and GroupValueChanged (GroupHeaderBand.cs lines 415-445).
	groupValue any
}

// NewGroupHeaderBand creates a GroupHeaderBand with defaults.
func NewGroupHeaderBand() *GroupHeaderBand {
	hf := NewHeaderFooterBandBase()
	hf.FlagIsColumnDependent = true // C# BandBase.IsColumnDependentBand (BandBase.cs line 589)
	return &GroupHeaderBand{
		HeaderFooterBandBase: *hf,
		sortOrder:            SortOrderAscending,
	}
}

// TypeName returns the FRX element name for this band.
func (*GroupHeaderBand) TypeName() string { return "GroupHeader" }

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

// Header returns the DataHeaderBand attached to this group.
// Mirrors C# GroupHeaderBand.Header property (GroupHeaderBand.cs lines 163-172).
func (g *GroupHeaderBand) Header() *DataHeaderBand { return g.header }

// SetHeader sets the DataHeaderBand for this group.
func (g *GroupHeaderBand) SetHeader(h *DataHeaderBand) { g.header = h }

// Footer returns the DataFooterBand attached to this group.
// Mirrors C# GroupHeaderBand.Footer property (GroupHeaderBand.cs lines 180-189).
func (g *GroupHeaderBand) Footer() *DataFooterBand { return g.footer }

// SetFooter sets the DataFooterBand for this group.
func (g *GroupHeaderBand) SetFooter(f *DataFooterBand) { g.footer = f }

// GroupDataBand traverses nested groups to find the DataBand.
// Only the last nested group may have a Data band; this method walks the
// chain g → g.NestedGroup → ... until it finds a non-nil Data.
// Mirrors C# GroupHeaderBand.GroupDataBand (GroupHeaderBand.cs lines 254-267).
func (g *GroupHeaderBand) GroupDataBand() *DataBand {
	group := g
	for group != nil {
		if group.data != nil {
			return group.data
		}
		group = group.nestedGroup
	}
	return nil
}

// DataSource returns the DataSource from the associated DataBand.
// Returns nil when GroupDataBand returns nil or the band has no data source set.
// Mirrors C# GroupHeaderBand.DataSource (GroupHeaderBand.cs lines 245-252).
func (g *GroupHeaderBand) DataSource() DataSource {
	db := g.GroupDataBand()
	if db == nil {
		return nil
	}
	return db.dataSource
}

// ResetGroupValue evaluates the group Condition and stores the result as the
// current group value baseline.  The caller supplies a calc function because
// GroupHeaderBand has no direct reference to the Report object.
// Mirrors C# GroupHeaderBand.ResetGroupValue (GroupHeaderBand.cs lines 415-425).
func (g *GroupHeaderBand) ResetGroupValue(calc func(string) (any, error)) error {
	if g.condition == "" {
		return nil
	}
	v, err := calc(g.condition)
	if err != nil {
		return err
	}
	g.groupValue = v
	return nil
}

// GroupValueChanged evaluates the group Condition and returns true when the
// result differs from the stored baseline.  It does NOT update the baseline —
// call ResetGroupValue to record the new value.
// Mirrors C# GroupHeaderBand.GroupValueChanged (GroupHeaderBand.cs lines 427-445).
func (g *GroupHeaderBand) GroupValueChanged(calc func(string) (any, error)) (bool, error) {
	if g.condition == "" {
		return false, nil
	}
	v, err := calc(g.condition)
	if err != nil {
		return false, err
	}
	if g.groupValue == nil {
		return v != nil, nil
	}
	return !objectsEqual(g.groupValue, v), nil
}

// objectsEqual performs a value-equality check equivalent to C# object.Equals.
// For comparable types the == operator is used; non-comparable types fall back
// to a pointer-identity check.
func objectsEqual(a, b any) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}
	defer func() { recover() }() // guard non-comparable types
	return a == b
}

// GetExpressions returns the group condition expression for use by the
// expression validator / walker.
// Mirrors C# GroupHeaderBand.GetExpressions (GroupHeaderBand.cs lines 369-371).
func (g *GroupHeaderBand) GetExpressions() []string {
	return []string{g.condition}
}

// Assign copies all GroupHeaderBand properties from src into g.
// Child band references (nestedGroup, data, groupFooter, header, footer) are
// NOT copied because they are structural, not property-level.
// Mirrors C# GroupHeaderBand.Assign (GroupHeaderBand.cs lines 339-348).
func (g *GroupHeaderBand) Assign(src *GroupHeaderBand) {
	if src == nil {
		return
	}
	g.condition = src.condition
	g.sortOrder = src.sortOrder
	g.keepTogether = src.keepTogether
	g.resetPageNumber = src.resetPageNumber
}

// Serialize writes GroupHeaderBand properties that differ from defaults.
// Attributes are written before child objects.
func (g *GroupHeaderBand) Serialize(w report.Writer) error {
	if err := g.HeaderFooterBandBase.serializeAttrs(w); err != nil {
		return err
	}
	if g.condition != "" {
		w.WriteStr("Condition", g.condition)
	}
	if g.sortOrder != SortOrderAscending {
		// C# writes the enum name via Converter.ToString (format "G"), e.g. "None".
		// GroupHeaderBand.cs:361 — writer.WriteValue("SortOrder", SortOrder)
		w.WriteStr("SortOrder", sortOrderToString(g.sortOrder))
	}
	if g.keepTogether {
		w.WriteBool("KeepTogether", true)
	}
	if g.resetPageNumber {
		w.WriteBool("ResetPageNumber", true)
	}
	if err := g.BandBase.serializeChildren(w); err != nil {
		return err
	}
	// Write special child bands not in the objects collection.
	// Mirrors C# GroupHeaderBand.GetChildObjects (GroupHeaderBand.cs:272).
	if g.header != nil {
		if err := w.WriteObject(g.header); err != nil {
			return err
		}
	}
	if g.nestedGroup != nil {
		if err := w.WriteObject(g.nestedGroup); err != nil {
			return err
		}
	}
	if g.data != nil {
		if err := w.WriteObject(g.data); err != nil {
			return err
		}
	}
	if g.groupFooter != nil {
		if err := w.WriteObject(g.groupFooter); err != nil {
			return err
		}
	}
	if g.footer != nil {
		if err := w.WriteObject(g.footer); err != nil {
			return err
		}
	}
	return nil
}

// AddChild routes deserialized child bands to the correct GroupHeaderBand field.
// Mirrors C# GroupHeaderBand.AddChild (GroupHeaderBand.cs:295).
func (g *GroupHeaderBand) AddChild(child report.Base) {
	switch c := child.(type) {
	case *DataBand:
		g.data = c
		c.SetParent(g)
	case *GroupHeaderBand:
		g.nestedGroup = c
		c.SetParent(g)
	case *GroupFooterBand:
		g.groupFooter = c
		c.SetParent(g)
	case *DataHeaderBand:
		g.header = c
		c.SetParent(g)
	case *DataFooterBand:
		g.footer = c
		c.SetParent(g)
	default:
		g.BandBase.AddChild(child)
	}
}

// Deserialize reads GroupHeaderBand properties.
func (g *GroupHeaderBand) Deserialize(r report.Reader) error {
	if err := g.HeaderFooterBandBase.Deserialize(r); err != nil {
		return err
	}
	g.condition = r.ReadStr("Condition", "")
	// C# serialises SortOrder as an enum name string ("None", "Ascending", "Descending").
	// GroupHeaderBand.cs:361 — writer.WriteValue("SortOrder", SortOrder)
	// where Converter.ToString writes Enum.Format(type, value, "G").
	g.sortOrder = sortOrderFromString(r.ReadStr("SortOrder", "Ascending"))
	g.keepTogether = r.ReadBool("KeepTogether", false)
	g.resetPageNumber = r.ReadBool("ResetPageNumber", false)
	return nil
}

// -----------------------------------------------------------------------
// DataBand
// -----------------------------------------------------------------------

// DataSource is a minimal interface for data iteration used by DataBand.
// Concrete implementations (data.BaseDataSource, json/csv sources, etc.) satisfy this.
type DataSource interface {
	RowCount() int
	First() error
	Next() error
	EOF() bool
	GetValue(column string) (any, error)
}

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
	collectChildRows     bool
	resetPageNumber      bool
	rowCount             int // virtual row count when no DataSource is set (default 1)

	// dataSourceAlias is the name/alias from the FRX DataSource attribute.
	// The engine resolves this to a live DataSource from the Dictionary at run time.
	dataSourceAlias string
	// dataSource is the data provider bound at runtime.
	dataSource DataSource
	// maxRows limits printed rows (0 = unlimited).
	maxRows int
	// sort holds the ordered list of sort specifications.
	sort []SortSpec
	// relationName is the name of the master-detail Relation that links this
	// DataBand to its parent DataBand.  Mirrors C# DataBand.Relation (DataBand.cs line 36).
	relationName string
}

// NewDataBand creates a DataBand with defaults.
func NewDataBand() *DataBand {
	b := NewBandBase()
	b.FlagCheckFreeSpace = true     // DataBands respect page free-space by default.
	b.FlagIsDataBand = true         // Used by engine child-band filtering (C# "band is DataBand").
	b.FlagIsColumnDependent = true  // C# BandBase.IsColumnDependentBand (BandBase.cs line 589).
	return &DataBand{
		BandBase: *b,
		columns:  NewBandColumns(),
	}
}

// TypeName returns the FRX element name for this band.
func (*DataBand) TypeName() string { return "Data" }

// DataSourceAlias returns the alias string from the FRX DataSource attribute.
// The engine resolves this to a live DataSource from the report Dictionary.
func (d *DataBand) DataSourceAlias() string { return d.dataSourceAlias }

// SetDataSourceAlias sets the data source alias for this band.
// The engine resolves this alias to a live DataSource from the report Dictionary at run time.
func (d *DataBand) SetDataSourceAlias(alias string) { d.dataSourceAlias = alias }

// DataSourceRef returns the bound data source (nil if not set).
func (d *DataBand) DataSourceRef() DataSource { return d.dataSource }

// SetDataSource binds a data source to this band.
func (d *DataBand) SetDataSource(ds DataSource) { d.dataSource = ds }

// UpdateDataSourceRef replaces the bound data source if ds implements DataSource.
// Mirrors C# IContainDataSource.UpdateDataSourceRef(DataSourceBase).
// The parameter type is any to avoid an import cycle between band/ and data/.
func (d *DataBand) UpdateDataSourceRef(ds any) {
	if src, ok := ds.(DataSource); ok {
		d.dataSource = src
	}
}

// MaxRows returns the maximum rows to print (0 = unlimited).
func (d *DataBand) MaxRows() int { return d.maxRows }

// SetMaxRows sets the maximum row limit.
func (d *DataBand) SetMaxRows(n int) { d.maxRows = n }

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

// Sort returns the ordered list of sort specifications.
func (d *DataBand) Sort() []SortSpec { return d.sort }

// SetSort replaces the sort specification list.
func (d *DataBand) SetSort(specs []SortSpec) { d.sort = specs }

// AddSort appends a sort specification.
func (d *DataBand) AddSort(spec SortSpec) { d.sort = append(d.sort, spec) }

// KeepSummary returns whether the data band stays with its footer band.
func (d *DataBand) KeepSummary() bool { return d.keepSummary }

// SetKeepSummary sets keep-summary.
func (d *DataBand) SetKeepSummary(v bool) { d.keepSummary = v }

// CollectChildRows returns whether the master band collects all child rows
// under a single master row (flattened master-detail).
func (d *DataBand) CollectChildRows() bool { return d.collectChildRows }

// SetCollectChildRows sets the collect-child-rows flag.
func (d *DataBand) SetCollectChildRows(v bool) { d.collectChildRows = v }

// ResetPageNumber returns whether the page number resets when this band starts printing.
func (d *DataBand) ResetPageNumber() bool { return d.resetPageNumber }

// SetResetPageNumber sets the reset-page-number flag.
func (d *DataBand) SetResetPageNumber(v bool) { d.resetPageNumber = v }

// VirtualRowCount returns the number of virtual rows when no DataSource is set (default 1).
func (d *DataBand) VirtualRowCount() int { return d.rowCount }

// SetVirtualRowCount sets the virtual row count.
func (d *DataBand) SetVirtualRowCount(n int) { d.rowCount = n }

// RelationName returns the name of the master-detail Relation linking this band
// to its parent DataBand.  The engine resolves this to a live Relation object
// from the report Dictionary at run time.
// Mirrors C# DataBand.Relation (DataBand.cs line 36).
func (d *DataBand) RelationName() string { return d.relationName }

// SetRelationName sets the relation name.
func (d *DataBand) SetRelationName(name string) { d.relationName = name }

// IsDatasourceEmpty returns true when the data source is nil or has zero rows.
func (d *DataBand) IsDatasourceEmpty() bool {
	return d.dataSource == nil || d.dataSource.RowCount() == 0
}

// IsDeepmostDataBand returns true when this DataBand has no nested sub-bands.
func (d *DataBand) IsDeepmostDataBand() bool {
	if d.objects == nil {
		return true
	}
	for i := 0; i < d.objects.Len(); i++ {
		switch d.objects.Get(i).(type) {
		case *DataBand, *GroupHeaderBand:
			return false
		}
	}
	return true
}

// GetExpressions returns the expressions used by this DataBand: all Sort
// expressions followed by the Filter expression.  The caller (expression walker
// / validator) uses this list to pre-compile or enumerate band expressions.
// Mirrors C# DataBand.GetExpressions (DataBand.cs lines 542-551).
func (d *DataBand) GetExpressions() []string {
	exprs := make([]string, 0, len(d.sort)+1)
	for _, s := range d.sort {
		expr := s.Expression
		if expr == "" {
			expr = s.Column
		}
		if expr != "" {
			exprs = append(exprs, expr)
		}
	}
	exprs = append(exprs, d.filter)
	return exprs
}

// Assign copies all DataBand properties from src into d.
// Child band references (header, footer) and sub-band collections are NOT
// copied because they are structural, not property-level.
// The sort collection is deep-copied so the two bands remain independent.
// Mirrors C# DataBand.Assign (DataBand.cs lines 462-483).
func (d *DataBand) Assign(src *DataBand) {
	if src == nil {
		return
	}
	d.dataSourceAlias = src.dataSourceAlias
	d.dataSource = src.dataSource
	d.rowCount = src.rowCount
	d.maxRows = src.maxRows
	// Deep-copy sort specs so caller mutations don't affect src.
	d.sort = make([]SortSpec, len(src.sort))
	copy(d.sort, src.sort)
	d.filter = src.filter
	d.printIfDetailEmpty = src.printIfDetailEmpty
	d.printIfDSEmpty = src.printIfDSEmpty
	d.keepTogether = src.keepTogether
	d.keepDetail = src.keepDetail
	d.idColumn = src.idColumn
	d.parentIDColumn = src.parentIDColumn
	d.indent = src.indent
	d.collectChildRows = src.collectChildRows
	d.resetPageNumber = src.resetPageNumber
	d.relationName = src.relationName
}

// Serialize writes DataBand properties that differ from defaults.
// Attributes are written before child objects (required by the streaming XML encoder).
// AddChild routes deserialized child bands to the correct DataBand field.
// Mirrors C# DataBand.AddChild (DataBand.cs).
func (d *DataBand) AddChild(child report.Base) {
	switch c := child.(type) {
	case *DataHeaderBand:
		d.header = c
		c.SetParent(d)
	case *DataFooterBand:
		d.footer = c
		c.SetParent(d)
	default:
		d.BandBase.AddChild(child)
	}
}

func (d *DataBand) Serialize(w report.Writer) error {
	// Write BandBase + parent attrs first (no children yet).
	if err := d.BandBase.serializeAttrs(w); err != nil {
		return err
	}
	// DataBand-specific attrs.
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
	if d.collectChildRows {
		w.WriteBool("CollectChildRows", true)
	}
	if d.resetPageNumber {
		w.WriteBool("ResetPageNumber", true)
	}
	if d.rowCount != 1 {
		w.WriteInt("RowCount", d.rowCount)
	}
	if d.relationName != "" {
		w.WriteStr("Relation", d.relationName)
	}
	// Write BandColumns attributes (matches C# Columns.Serialize).
	if d.columns.count > 0 {
		w.WriteInt("Columns.Count", d.columns.count)
	}
	if d.columns.Width != 0 {
		w.WriteFloat("Columns.Width", d.columns.Width)
	}
	if d.columns.Layout != ColumnLayoutAcrossThenDown {
		w.WriteStr("Columns.Layout", formatColumnLayout(d.columns.Layout))
	}
	if d.columns.MinRowCount != 0 {
		w.WriteInt("Columns.MinRowCount", d.columns.MinRowCount)
	}
	// Write <Sort> child collection before band object children.
	if len(d.sort) > 0 {
		sc := &sortCollection{items: d.sort}
		if err := w.WriteObjectNamed("Sort", sc); err != nil {
			return err
		}
	}
	// Write child objects after all attrs.
	if err := d.BandBase.serializeChildren(w); err != nil {
		return err
	}
	// Write header/footer sub-bands not in the objects collection.
	// Mirrors C# DataBand.GetChildObjects.
	if d.header != nil {
		if err := w.WriteObject(d.header); err != nil {
			return err
		}
	}
	if d.footer != nil {
		if err := w.WriteObject(d.footer); err != nil {
			return err
		}
	}
	return nil
}

// Deserialize reads DataBand properties.
func (d *DataBand) Deserialize(r report.Reader) error {
	if err := d.BandBase.Deserialize(r); err != nil {
		return err
	}
	d.dataSourceAlias = r.ReadStr("DataSource", "")
	d.relationName = r.ReadStr("Relation", "")
	d.filter = r.ReadStr("Filter", "")
	// Parse sort string "Col1 ASC;Col2 DESC"
	if sortStr := r.ReadStr("Sort", ""); sortStr != "" {
		for _, part := range strings.Split(sortStr, ";") {
			part = strings.TrimSpace(part)
			if part == "" {
				continue
			}
			fields := strings.Fields(part)
			spec := SortSpec{Column: fields[0]}
			if len(fields) >= 2 && strings.EqualFold(fields[1], "DESC") {
				spec.Order = SortOrderDescending
			}
			d.sort = append(d.sort, spec)
		}
	}
	d.printIfDetailEmpty = r.ReadBool("PrintIfDetailEmpty", false)
	d.printIfDSEmpty = r.ReadBool("PrintIfDatasourceEmpty", false)
	d.keepTogether = r.ReadBool("KeepTogether", false)
	d.keepDetail = r.ReadBool("KeepDetail", false)
	d.idColumn = r.ReadStr("IdColumn", "")
	d.parentIDColumn = r.ReadStr("ParentIdColumn", "")
	d.indent = r.ReadFloat("Indent", 0)
	d.keepSummary = r.ReadBool("KeepSummary", false)
	d.collectChildRows = r.ReadBool("CollectChildRows", false)
	d.resetPageNumber = r.ReadBool("ResetPageNumber", false)
	d.rowCount = r.ReadInt("RowCount", 1)
	if n := r.ReadInt("Columns.Count", 0); n > 0 {
		_ = d.columns.SetCount(n)
	}
	d.columns.Width = r.ReadFloat("Columns.Width", 0)
	d.columns.Layout = parseColumnLayout(r.ReadStr("Columns.Layout", "AcrossThenDown"))
	d.columns.MinRowCount = r.ReadInt("Columns.MinRowCount", 0)
	return nil
}

// DeserializeChild handles the <Sort> container element that holds child
// <Sort Expression="..." Descending="true"/> sort-spec items in FastReport FRX files.
func (d *DataBand) DeserializeChild(childType string, r report.Reader) bool {
	if childType != "Sort" {
		return false
	}
	// The outer <Sort> element is a list container; iterate its children.
	for {
		ct, ok := r.NextChild()
		if !ok {
			break
		}
		if ct == "Sort" {
			expr := r.ReadStr("Expression", "")
			descending := r.ReadBool("Descending", false)
			if expr != "" {
				spec := SortSpec{Column: expr}
				if descending {
					spec.Order = SortOrderDescending
				}
				d.sort = append(d.sort, spec)
			}
			// Drain any unexpected grandchildren.
			for {
				_, ok2 := r.NextChild()
				if !ok2 {
					break
				}
				if r.FinishChild() != nil { break }
			}
		} else {
			// Drain unexpected element's children then finish it.
			for {
				_, ok2 := r.NextChild()
				if !ok2 {
					break
				}
				if r.FinishChild() != nil { break }
			}
		}
		if r.FinishChild() != nil { break }
	}
	return true
}
