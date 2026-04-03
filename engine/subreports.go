package engine

import (
	"github.com/andrewloable/go-fastreport/band"
	"github.com/andrewloable/go-fastreport/object"
	"github.com/andrewloable/go-fastreport/report"
)

// subreports.go provides inner and outer subreport rendering.
//
// Inner subreport (PrintOnParent=true): the subreport's bands are drawn inside
// the parent band at the subreport object's Left/Top offset.
//
// Outer subreport (PrintOnParent=false): the subreport runs after the parent
// band, starting at CurY and advancing CurY when done.

// renderSubreport runs the bands of a subreport's linked ReportPage.
//
// When a master DataBand is active (e.masterDataBand != nil), it applies
// master-detail relation filters to the subreport page's bands before running
// them. This mirrors C# BandBase.ParentDataBand which climbs through
// ReportPage.Subreport back to the enclosing DataBand, and then
// DataBand.InitDataSource which calls DataSource.Init(parentDataSource) to
// filter the child datasource to the current parent row.
// C# ref: BandBase.cs ParentDataBand lines 311-323, DataBand.cs line 567.
func (e *ReportEngine) renderSubreport(sr *object.SubreportObject) {
	pgName := sr.ReportPageName()
	if pgName == "" || e.report == nil {
		return
	}
	pg := e.report.FindPage(pgName)
	if pg == nil {
		return
	}
	bands := pg.Bands()

	// Apply master-detail relation filters so subreport data bands are
	// filtered to the current parent row. Save and restore the masterDataBand
	// to handle nested subreports correctly.
	saveMaster := e.masterDataBand
	var restore func()
	if saveMaster != nil {
		restore = e.applyRelationFilters(saveMaster, bands)
		// Clear masterDataBand inside subreport so nested sub-bands within
		// the subreport don't incorrectly re-apply the outer relation.
		// If the subreport itself has data bands with sub-bands, those will
		// set their own masterDataBand via RunDataBandFull.
		e.masterDataBand = nil
	} else {
		restore = func() {}
	}

	_ = e.runBands(bands)
	restore()
	e.masterDataBand = saveMaster
}

// RenderInnerSubreport renders a subreport inside the coordinate space of
// parentBand. When a PreparedBand for the parent exists on the current page,
// subreport objects are merged directly into it at the SubreportObject's
// Left/Top offset (PrintOnParent semantics). CurX, CurY, and outputBand are
// always saved and restored.
func (e *ReportEngine) RenderInnerSubreport(parentBand *band.BandBase, sr *object.SubreportObject) {
	saveCurX := e.curX
	saveCurY := e.curY
	saveOutputBand := e.outputBand
	saveOutputBandOffsetX := e.outputBandOffsetX
	saveOutputBandOffsetY := e.outputBandOffsetY

	// Locate the PreparedBand for parentBand on the current page so that
	// subreport objects can be merged into it (PrintOnParent semantics).
	// We search from the end because the parent band was just rendered.
	if e.preparedPages != nil {
		if pg := e.preparedPages.CurrentPage(); pg != nil {
			for i := len(pg.Bands) - 1; i >= 0; i-- {
				if pg.Bands[i].Name == parentBand.Name() {
					e.outputBand = pg.Bands[i]
					break
				}
			}
		}
	}

	e.outputBandOffsetX = float32(sr.Left())
	e.outputBandOffsetY = float32(sr.Top())
	// Reset local cursor to 0,0 so subreport band Y positions are relative
	// to the SubreportObject's Top within the parent band.
	e.curX = 0
	e.curY = 0

	e.renderSubreport(sr)

	e.curX = saveCurX
	e.curY = saveCurY
	e.outputBand = saveOutputBand
	e.outputBandOffsetX = saveOutputBandOffsetX
	e.outputBandOffsetY = saveOutputBandOffsetY
}

// evalSubreportVisible evaluates VisibleExpression (if set) and returns
// whether the subreport should be rendered. Matches C# RenderInnerSubreports
// / RenderOuterSubreports lines 46-49 / 73-76 in ReportEngine.Subreports.cs:
//   if (!String.IsNullOrEmpty(subreport.VisibleExpression))
//     subreport.Visible = subreport.CalcVisibleExpression(subreport.VisibleExpression);
//   if (subreport.Visible && ...) RenderXxxSubreport(...)
func (e *ReportEngine) evalSubreportVisible(sr *object.SubreportObject) bool {
	if expr := sr.VisibleExpression(); expr != "" && e.report != nil {
		visible := sr.CalcVisibleExpression(expr, func(s string) (any, error) {
			return e.report.Calc(s)
		})
		sr.SetVisible(visible)
	}
	return sr.Visible()
}

// RenderInnerSubreports finds all inner SubreportObjects in parentBand's
// object collection and renders each one in turn.
// Mirrors C# ReportEngine.Subreports.cs RenderInnerSubreports (lines 38-54):
// VisibleExpression is evaluated before checking Visible, then PrintOnParent.
func (e *ReportEngine) RenderInnerSubreports(parentBand *band.BandBase) {
	objs := parentBand.Objects()
	for i := 0; i < objs.Len(); i++ {
		obj := objs.Get(i)
		sr, ok := obj.(*object.SubreportObject)
		if !ok {
			continue
		}
		// Apply VisibleExpression before checking Visible and PrintOnParent.
		// C# ref: RenderInnerSubreports lines 46-52.
		if !e.evalSubreportVisible(sr) {
			continue
		}
		if !sr.PrintOnParent() {
			continue
		}
		e.RenderInnerSubreport(parentBand, sr)
	}
}

// RenderOuterSubreports finds all outer SubreportObjects in parentBand's
// object collection and renders each in its own column-like area, then
// advances CurY to the maximum Y reached by any subreport.
// Mirrors C# ReportEngine.Subreports.cs RenderOuterSubreports (lines 56-114):
//
//	float saveCurY = CurY;
//	float saveOriginX = originX;
//	// For each outer subreport:
//	//   CurY = saveCurY - subreport.Height;  (line 83)
//	//   originX = saveOriginX + subreport.Left;
func (e *ReportEngine) RenderOuterSubreports(parentBand *band.BandBase) {
	saveCurY := e.curY
	saveOriginX := e.originX
	saveCurPage := e.curPage

	maxY := e.curY
	maxPage := e.curPage
	hasSubreports := false

	objs := parentBand.Objects()
	for i := 0; i < objs.Len(); i++ {
		obj := objs.Get(i)
		sr, ok := obj.(*object.SubreportObject)
		if !ok {
			continue
		}
		// Apply VisibleExpression before checking Visible and PrintOnParent.
		// C# ref: RenderOuterSubreports lines 73-78.
		if !e.evalSubreportVisible(sr) {
			continue
		}
		if sr.PrintOnParent() {
			continue
		}
		hasSubreports = true

		// Restore start position for each subreport (C# lines 82-84):
		//   CurPage = saveCurPage;
		//   CurY = saveCurY - subreport.Height;
		//   originX = saveOriginX + subreport.Left;
		e.curPage = saveCurPage
		e.curY = saveCurY - sr.Height()
		e.originX = saveOriginX + sr.Left()

		e.renderSubreport(sr)

		// Track the furthest page/Y reached by any subreport (C# lines 91-100).
		if e.curPage == maxPage {
			if e.curY > maxY {
				maxY = e.curY
			}
		} else if e.curPage > maxPage {
			maxPage = e.curPage
			maxY = e.curY
		}
	}

	// Restore originX (C# finally block line 111).
	e.originX = saveOriginX
	if hasSubreports {
		e.curPage = maxPage
		e.curY = maxY
	}
}

// runBandsSlice is an alias used by renderSubreport to run a []report.Base.
// It delegates to the existing runBands method.
func (e *ReportEngine) runBandsFromBase(bands []report.Base) error {
	return e.runBands(bands)
}

// collectAllSubreportPageNames scans all pages and bands for SubreportObjects
// and returns a set of the page names they reference. Pages in this set should
// not be rendered as top-level report pages.
//
// Mirrors the C# RunReportPages condition: page.Subreport == null (line 92 in
// ReportEngine.Pages.cs). In C#, this is a back-reference set on ReportPage when
// a SubreportObject links to it. In Go, we achieve the same by forward-scanning
// all SubreportObjects and collecting their ReportPageName values.
func (e *ReportEngine) collectAllSubreportPageNames() map[string]bool {
	result := make(map[string]bool)
	if e.report == nil {
		return result
	}
	for _, pg := range e.report.Pages() {
		for _, b := range pg.AllBands() {
			e.scanForSubreportPageNames(b, result)
		}
	}
	return result
}

// scanForSubreportPageNames recursively walks b and its ObjectCollection
// children, looking for SubreportObjects.
func (e *ReportEngine) scanForSubreportPageNames(b report.Base, result map[string]bool) {
	if sr, ok := b.(*object.SubreportObject); ok {
		if name := sr.ReportPageName(); name != "" {
			result[name] = true
		}
		return // SubreportObjects don't contain other bands/objects
	}
	// Recurse into object collections exposed by bands.
	type hasObjects interface {
		Objects() *report.ObjectCollection
	}
	if ho, ok := b.(hasObjects); ok {
		objs := ho.Objects()
		for i := 0; i < objs.Len(); i++ {
			e.scanForSubreportPageNames(objs.Get(i), result)
		}
	}
}
