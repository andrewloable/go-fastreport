package band_test

import (
	"testing"

	"github.com/andrewloable/go-fastreport/band"
)

// -----------------------------------------------------------------------
// Thin band types (just instantiation + inheritance checks)
// -----------------------------------------------------------------------

func TestNewReportTitleBand(t *testing.T) {
	b := band.NewReportTitleBand()
	if b == nil {
		t.Fatal("NewReportTitleBand returned nil")
	}
	if !b.FirstRowStartsNewPage() {
		t.Error("ReportTitleBand should inherit FirstRowStartsNewPage=true")
	}
}

func TestNewReportSummaryBand(t *testing.T) {
	b := band.NewReportSummaryBand()
	if b == nil {
		t.Fatal("NewReportSummaryBand returned nil")
	}
	if b.KeepWithData() {
		t.Error("ReportSummaryBand.KeepWithData should default to false")
	}
	if b.RepeatOnEveryPage() {
		t.Error("ReportSummaryBand.RepeatOnEveryPage should default to false")
	}
}

func TestNewPageHeaderBand(t *testing.T) {
	b := band.NewPageHeaderBand()
	if b == nil {
		t.Fatal("NewPageHeaderBand returned nil")
	}
}

func TestNewPageFooterBand(t *testing.T) {
	b := band.NewPageFooterBand()
	if b == nil {
		t.Fatal("NewPageFooterBand returned nil")
	}
}

func TestNewColumnHeaderBand(t *testing.T) {
	b := band.NewColumnHeaderBand()
	if b == nil {
		t.Fatal("NewColumnHeaderBand returned nil")
	}
}

func TestNewColumnFooterBand(t *testing.T) {
	b := band.NewColumnFooterBand()
	if b == nil {
		t.Fatal("NewColumnFooterBand returned nil")
	}
}

func TestNewDataHeaderBand(t *testing.T) {
	b := band.NewDataHeaderBand()
	if b == nil {
		t.Fatal("NewDataHeaderBand returned nil")
	}
	b.SetKeepWithData(true)
	if !b.KeepWithData() {
		t.Error("DataHeaderBand.KeepWithData should be true")
	}
}

func TestNewDataFooterBand(t *testing.T) {
	b := band.NewDataFooterBand()
	if b == nil {
		t.Fatal("NewDataFooterBand returned nil")
	}
	b.SetRepeatOnEveryPage(true)
	if !b.RepeatOnEveryPage() {
		t.Error("DataFooterBand.RepeatOnEveryPage should be true")
	}
}

func TestNewGroupFooterBand(t *testing.T) {
	b := band.NewGroupFooterBand()
	if b == nil {
		t.Fatal("NewGroupFooterBand returned nil")
	}
}

func TestNewOverlayBand(t *testing.T) {
	b := band.NewOverlayBand()
	if b == nil {
		t.Fatal("NewOverlayBand returned nil")
	}
}

// -----------------------------------------------------------------------
// GroupHeaderBand
// -----------------------------------------------------------------------

func TestNewGroupHeaderBand_Defaults(t *testing.T) {
	g := band.NewGroupHeaderBand()
	if g == nil {
		t.Fatal("NewGroupHeaderBand returned nil")
	}
	if g.SortOrder() != band.SortOrderAscending {
		t.Errorf("SortOrder default = %d, want Ascending", g.SortOrder())
	}
	if g.KeepTogether() {
		t.Error("KeepTogether should default to false")
	}
	if g.ResetPageNumber() {
		t.Error("ResetPageNumber should default to false")
	}
	if g.Condition() != "" {
		t.Errorf("Condition default = %q, want empty", g.Condition())
	}
	if g.NestedGroup() != nil {
		t.Error("NestedGroup should default to nil")
	}
	if g.Data() != nil {
		t.Error("Data should default to nil")
	}
}

func TestGroupHeaderBand_Condition(t *testing.T) {
	g := band.NewGroupHeaderBand()
	g.SetCondition("[Orders.CustomerName]")
	if g.Condition() != "[Orders.CustomerName]" {
		t.Errorf("Condition = %q", g.Condition())
	}
}

func TestGroupHeaderBand_SortOrder(t *testing.T) {
	g := band.NewGroupHeaderBand()
	g.SetSortOrder(band.SortOrderDescending)
	if g.SortOrder() != band.SortOrderDescending {
		t.Error("SortOrder should be Descending")
	}
}

func TestGroupHeaderBand_KeepTogether(t *testing.T) {
	g := band.NewGroupHeaderBand()
	g.SetKeepTogether(true)
	if !g.KeepTogether() {
		t.Error("KeepTogether should be true")
	}
}

func TestGroupHeaderBand_ResetPageNumber(t *testing.T) {
	g := band.NewGroupHeaderBand()
	g.SetResetPageNumber(true)
	if !g.ResetPageNumber() {
		t.Error("ResetPageNumber should be true")
	}
}

func TestGroupHeaderBand_NestedGroup(t *testing.T) {
	outer := band.NewGroupHeaderBand()
	inner := band.NewGroupHeaderBand()
	outer.SetNestedGroup(inner)
	if outer.NestedGroup() != inner {
		t.Error("NestedGroup should be the inner band")
	}
}

func TestGroupHeaderBand_DataAndFooter(t *testing.T) {
	g := band.NewGroupHeaderBand()
	d := band.NewDataBand()
	f := band.NewGroupFooterBand()
	g.SetData(d)
	g.SetGroupFooter(f)
	if g.Data() != d {
		t.Error("Data should be set")
	}
	if g.GroupFooter() != f {
		t.Error("GroupFooter should be set")
	}
}

func TestGroupHeaderBand_InheritsKeepWithData(t *testing.T) {
	g := band.NewGroupHeaderBand()
	g.SetKeepWithData(true)
	if !g.KeepWithData() {
		t.Error("GroupHeaderBand.KeepWithData should be settable via HeaderFooterBandBase")
	}
}

// -----------------------------------------------------------------------
// DataBand
// -----------------------------------------------------------------------

func TestNewDataBand_Defaults(t *testing.T) {
	d := band.NewDataBand()
	if d == nil {
		t.Fatal("NewDataBand returned nil")
	}
	if d.Filter() != "" {
		t.Errorf("Filter default = %q, want empty", d.Filter())
	}
	if d.PrintIfDetailEmpty() {
		t.Error("PrintIfDetailEmpty should default to false")
	}
	if d.PrintIfDSEmpty() {
		t.Error("PrintIfDSEmpty should default to false")
	}
	if d.KeepTogether() {
		t.Error("KeepTogether should default to false")
	}
	if d.KeepDetail() {
		t.Error("KeepDetail should default to false")
	}
	if d.Columns() == nil {
		t.Error("Columns should not be nil")
	}
	if d.IDColumn() != "" {
		t.Errorf("IDColumn default = %q, want empty", d.IDColumn())
	}
}

func TestDataBand_Filter(t *testing.T) {
	d := band.NewDataBand()
	d.SetFilter("[Amount] > 0")
	if d.Filter() != "[Amount] > 0" {
		t.Errorf("Filter = %q", d.Filter())
	}
}

func TestDataBand_PrintIfDetailEmpty(t *testing.T) {
	d := band.NewDataBand()
	d.SetPrintIfDetailEmpty(true)
	if !d.PrintIfDetailEmpty() {
		t.Error("PrintIfDetailEmpty should be true")
	}
}

func TestDataBand_PrintIfDSEmpty(t *testing.T) {
	d := band.NewDataBand()
	d.SetPrintIfDSEmpty(true)
	if !d.PrintIfDSEmpty() {
		t.Error("PrintIfDSEmpty should be true")
	}
}

func TestDataBand_KeepTogether(t *testing.T) {
	d := band.NewDataBand()
	d.SetKeepTogether(true)
	if !d.KeepTogether() {
		t.Error("KeepTogether should be true")
	}
}

func TestDataBand_KeepDetail(t *testing.T) {
	d := band.NewDataBand()
	d.SetKeepDetail(true)
	if !d.KeepDetail() {
		t.Error("KeepDetail should be true")
	}
}

func TestDataBand_HierarchyColumns(t *testing.T) {
	d := band.NewDataBand()
	d.SetIDColumn("ID")
	d.SetParentIDColumn("ParentID")
	if d.IDColumn() != "ID" {
		t.Errorf("IDColumn = %q, want ID", d.IDColumn())
	}
	if d.ParentIDColumn() != "ParentID" {
		t.Errorf("ParentIDColumn = %q, want ParentID", d.ParentIDColumn())
	}
}

func TestDataBand_Indent(t *testing.T) {
	d := band.NewDataBand()
	d.SetIndent(20)
	if d.Indent() != 20 {
		t.Errorf("Indent = %v, want 20", d.Indent())
	}
}

func TestDataBand_KeepSummary(t *testing.T) {
	d := band.NewDataBand()
	d.SetKeepSummary(true)
	if !d.KeepSummary() {
		t.Error("KeepSummary should be true")
	}
}

func TestDataBand_HeaderFooter(t *testing.T) {
	d := band.NewDataBand()
	h := band.NewDataHeaderBand()
	f := band.NewDataFooterBand()
	d.SetHeader(h)
	d.SetFooter(f)
	if d.Header() != h {
		t.Error("Header should be set")
	}
	if d.Footer() != f {
		t.Error("Footer should be set")
	}
}

func TestDataBand_InheritsCanBreak(t *testing.T) {
	d := band.NewDataBand()
	if !d.CanBreak() {
		t.Error("DataBand should inherit CanBreak=true from BandBase")
	}
}
