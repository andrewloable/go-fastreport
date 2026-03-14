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
	_ = e.runBands(bands)
}

// RenderInnerSubreport renders a subreport inside the coordinate space of
// parentBand. It saves/restores CurX, CurY and any outputBand state.
func (e *ReportEngine) RenderInnerSubreport(parentBand *band.BandBase, sr *object.SubreportObject) {
	saveCurX := e.curX
	saveCurY := e.curY

	e.curX = float32(sr.Left())
	e.curY = float32(sr.Top())

	e.renderSubreport(sr)

	e.curX = saveCurX
	e.curY = saveCurY
}

// RenderInnerSubreports finds all inner SubreportObjects in parentBand's
// object collection and renders each one in turn.
func (e *ReportEngine) RenderInnerSubreports(parentBand *band.BandBase) {
	objs := parentBand.Objects()
	for i := 0; i < objs.Len(); i++ {
		obj := objs.Get(i)
		sr, ok := obj.(*object.SubreportObject)
		if !ok {
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
func (e *ReportEngine) RenderOuterSubreports(parentBand *band.BandBase) {
	saveCurY := e.curY
	saveCurX := e.curX

	maxY := e.curY
	hasSubreports := false

	objs := parentBand.Objects()
	for i := 0; i < objs.Len(); i++ {
		obj := objs.Get(i)
		sr, ok := obj.(*object.SubreportObject)
		if !ok {
			continue
		}
		if sr.PrintOnParent() {
			continue
		}
		hasSubreports = true

		// Restore start Y and set X offset for this subreport column.
		e.curY = saveCurY
		e.curX = saveCurX + float32(sr.Left())

		e.renderSubreport(sr)

		if e.curY > maxY {
			maxY = e.curY
		}
	}

	e.curX = saveCurX
	if hasSubreports {
		e.curY = maxY
	}
}

// runBandsSlice is an alias used by renderSubreport to run a []report.Base.
// It delegates to the existing runBands method.
func (e *ReportEngine) runBandsFromBase(bands []report.Base) error {
	return e.runBands(bands)
}
