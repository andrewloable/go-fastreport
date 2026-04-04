package engine

import (
	"math"
	"reflect"

	"github.com/andrewloable/go-fastreport/band"
	"github.com/andrewloable/go-fastreport/preview"
	"github.com/andrewloable/go-fastreport/report"
	"github.com/andrewloable/go-fastreport/reportpkg"
	"github.com/andrewloable/go-fastreport/style"
)

// ── PreparedPages integration ─────────────────────────────────────────────────

// PreparedPages returns the engine's prepared-pages collection.
func (e *ReportEngine) PreparedPages() *preview.PreparedPages {
	return e.preparedPages
}

// ── Page lifecycle ────────────────────────────────────────────────────────────

// startPage initialises a new output page from the given ReportPage template.
// It sets up page dimensions, adds a page to PreparedPages, and shows the
// Overlay, ReportTitle (if first page), and PageHeader bands.
//
// Matches C# StartFirstPage (isFirst=true) and StartPage (isFirst=false).
func (e *ReportEngine) startPage(pg *reportpkg.ReportPage, isFirst bool) {
	const mmPerPx = 3.78
	e.pageWidth = (pg.PaperWidth - pg.LeftMargin - pg.RightMargin) * mmPerPx
	e.pageHeight = (pg.PaperHeight - pg.TopMargin - pg.BottomMargin) * mmPerPx
	e.freeSpace = e.pageHeight
	if pg.UnlimitedHeight {
		e.freeSpace = math.MaxFloat32 / 2 // effectively unlimited
	}
	e.curX = 0
	e.curY = 0
	e.curColumn = 0

	// C# StartFirstPageShared: if the page has ResetPageNumber, reset the
	// logical page counter before adding the page.
	if isFirst && pg.ResetPageNumber {
		e.ResetLogicalPageNumber()
	}

	// Add page: increment counters, add to PreparedPages, call IncLogicalPageNumber.
	// Mirrors C# PreparedPages.AddPage() which calls engine.IncLogicalPageNumber().
	e.totalPages++
	e.pageNo++
	if e.preparedPages != nil {
		e.preparedPages.AddPage(e.pageWidth, e.pageHeight, e.pageNo)
		if cp := e.preparedPages.CurrentPage(); cp != nil {
			cp.Landscape = pg.Landscape
		}
	}
	e.curPage++
	e.IncLogicalPageNumber()

	// C# StartFirstPageShared: StartOnOddPage — if true and the current page
	// index is odd (i.e. the page landed on an even page number), add a blank
	// filler page so the report content starts on an odd page.
	// C# check: if (page.StartOnOddPage && (CurPage % 2) == 1) AddPage(page)
	// CurPage is 0-based in C#, so (CurPage % 2) == 1 means the index is odd
	// = the page number is even = we need a blank to push to an odd page.
	if isFirst && pg.StartOnOddPage && (e.curPage%2) == 1 {
		e.totalPages++
		e.pageNo++
		if e.preparedPages != nil {
			e.preparedPages.AddPage(e.pageWidth, e.pageHeight, e.pageNo)
		}
		e.curPage++
		e.IncLogicalPageNumber()
	}

	// C# StartFirstPageShared: track firstReportPage for page numbering.
	if e.isFirstReportPage {
		e.firstReportPage = e.curPage
	}
	e.isFirstReportPage = false

	// Keep system variables in sync (after page numbering is updated).
	e.syncPageVariables()

	if e.preparedPages != nil {
		// Attach watermark metadata to the prepared page.
		if pg.Watermark != nil && pg.Watermark.Enabled {
			e.attachWatermark(pg)
		}
	}

	// Page-level outline entry.
	if pg.OutlineExpression != "" {
		text := e.evalText(pg.OutlineExpression)
		if text != "" {
			e.AddOutline(text)
		}
	}

	// Back page: render referenced page's bands before the current page content.
	e.applyBackPage(pg)

	if isFirst {
		// C# StartFirstPage: show overlay, title, page header.
		e.showBand(pg.Overlay())
		// Title and PageHeader order depends on TitleBeforeHeader flag.
		if pg.TitleBeforeHeader {
			e.showBand(pg.ReportTitle())
			e.showBand(pg.PageHeader())
		} else {
			e.showBand(pg.PageHeader())
			e.showBand(pg.ReportTitle())
		}

		// Column header below the page header.
		e.columnStartY = e.curY
		e.showBand(pg.ColumnHeader())

		// C# fires ColumnStarted event at end of StartFirstPage.
		e.OnStateChanged(pg, EngineStateColumnStarted)
	} else {
		// C# StartPage (non-first): show overlay, page header, fire PageStarted,
		// then call StartColumn (ColumnHeader + reprint headers + ColumnStarted).
		e.showBand(pg.Overlay())
		e.showBand(pg.PageHeader())
		e.OnStateChanged(pg, EngineStatePageStarted)

		e.columnStartY = e.curY
		e.startColumn(pg)
	}
}

// endPage finalises the current page.
//
// Matches C# EndLastPage / EndPage flow:
//   - EndLastPage (isLast=true): fires ColumnFinished event, shows
//     ReportSummary (which internally triggers ColumnFooter in
//     AddToPreparedPages) OR shows ColumnFooter if no ReportSummary,
//     then shows PageFooter.
//   - EndPage (isLast=false): fires PageFinished event, shows PageFooter,
//     then optionally starts the next page.
//
// PageFooter and ReportSummary always span the full page width, so curX is
// reset to 0 before rendering them to prevent the column X offset from
// being applied to their objects.
func (e *ReportEngine) endPage(pg *reportpkg.ReportPage, isLast bool) {
	if isLast {
		// C# EndLastPage: fire ColumnFinished event first.
		e.OnStateChanged(pg, EngineStateColumnFinished)

		// Full-page bands must not inherit any column X offset.
		savedX := e.curX
		e.curX = 0

		if pg.ReportSummary() != nil {
			// Do not show column footer here — C# handles it inside
			// AddToPreparedPages when processing ReportSummaryBand.
			e.showBand(pg.ReportSummary())
		} else {
			e.showBand(pg.ColumnFooter())
		}

		e.showBand(pg.PageFooter())
		e.curX = savedX
	} else {
		// C# EndPage(startPage): fire PageFinished, show footer.
		e.OnStateChanged(pg, EngineStatePageFinished)

		savedX := e.curX
		e.curX = 0
		// C# ShowPageFooter(startPage=true) — mirrors ReportEngine.Pages.cs lines 119-131.
		// In double-pass reports on the last page, if the footer is LastPage-only,
		// call ShiftLastPage() instead of showing the band (adds a virtual page entry
		// so TotalPages accounts for the extra page the footer will occupy).
		pf := pg.PageFooter()
		if !e.FirstPass() &&
			e.knownTotalPages > 0 &&
			e.curPage == e.knownTotalPages-1 &&
			pf != nil &&
			(pf.PrintOn()&report.PrintOnLastPage) != 0 &&
			(pf.PrintOn()&report.PrintOnFirstPage) == 0 {
			e.ShiftLastPage()
		} else {
			e.showBand(pf)
		}
		e.curX = savedX
	}
}

// applyBackPage renders the bands of the page referenced by pg.BackPage behind
// the current page's content. It checks BackPageOddEven to decide whether to
// apply on odd pages (1), even pages (2), or both (0).
//
// The back-page bands are printed at their natural Y positions (same as a
// normal page start) and do not advance CurY — they appear as background
// layers. After rendering, CurY is reset to 0 so the main page content
// starts from the top.
func (e *ReportEngine) applyBackPage(pg *reportpkg.ReportPage) {
	if pg.BackPage == "" {
		return
	}
	// Check odd/even constraint.
	switch pg.BackPageOddEven {
	case 1: // odd pages only
		if e.pageNo%2 == 0 {
			return
		}
	case 2: // even pages only
		if e.pageNo%2 != 0 {
			return
		}
	}

	// Find the referenced back-page template.
	backPg := e.report.FindPage(pg.BackPage)
	if backPg == nil {
		return
	}

	// Save and restore CurY so that back-page rendering does not shift the
	// main page's print position.
	savedY := e.curY
	e.curY = 0

	// Render all bands of the back page at their natural positions.
	for _, b := range backPg.AllBands() {
		e.showBandNoAdvance(b)
	}

	e.curY = savedY
}

// showBandNoAdvance renders a band into the prepared pages at the current CurY
// position without advancing CurY. Used for back-page rendering.
func (e *ReportEngine) showBandNoAdvance(b report.Base) {
	if b == nil {
		return
	}
	if v := reflect.ValueOf(b); v.Kind() == reflect.Ptr && v.IsNil() {
		return
	}
	type hasVisible interface{ Visible() bool }
	if vis, ok := b.(hasVisible); ok && !vis.Visible() {
		return
	}
	height := e.bandHeight(b)
	if height <= 0 {
		return
	}
	if e.preparedPages != nil {
		pb := &preview.PreparedBand{
			Name:          b.Name() + "_back",
			Top:           e.curY,
			Height:        height,
			NotExportable: bandNotExportable(b),
		}
		type hasObjects interface{ Objects() *report.ObjectCollection }
		if ho, ok := b.(hasObjects); ok {
			e.populateBandObjects2(nil, ho.Objects(), pb)
		}
		_ = e.preparedPages.AddBand(pb)
	}
	// Advance local cursor within back-page rendering so consecutive bands stack.
	e.curY += height
}

// startColumn initialises the current column.
// freeSpace is reset to the remaining vertical height from columnStartY,
// so that multi-column pages don't run out of space immediately after
// advancing from the previous column.
//
// Matches C# StartColumn: reset curY, show ColumnHeader, show reprint headers,
// fire ColumnStarted event.
func (e *ReportEngine) startColumn(pg *reportpkg.ReportPage) {
	e.curY = e.columnStartY
	e.freeSpace = e.pageHeight - e.columnStartY
	e.showBand(pg.ColumnHeader())
	e.ShowReprintHeaders()
	e.OnStateChanged(pg, EngineStateColumnStarted)
}

// endColumn finalises the current column. If there are more columns, advance to the
// next column; otherwise end the page and start a new one.
//
// Matches C# EndColumn flow:
//  1. Fire ColumnFinished event
//  2. If keeping, CutObjects
//  3. ShowReprintFooters
//  4. Show ColumnFooter
//  5. Increment curColumn; wrap to 0 if past last
//  6. Calculate curX from column positions
//  7. If curColumn == 0 -> EndPage else StartColumn
//  8. If keeping, PasteObjects
func (e *ReportEngine) endColumn(pg *reportpkg.ReportPage) bool {
	cols := pg.Columns.Count
	if cols <= 1 {
		return false
	}
	// Step 1: fire ColumnFinished event.
	e.OnStateChanged(pg, EngineStateColumnFinished)

	// Step 2: check keep.
	if e.keeping {
		e.cutObjects()
	}

	// Step 3: show reprint footers.
	e.ShowReprintFooters()

	// Step 4: show column footer.
	e.showBand(pg.ColumnFooter())

	// Step 5: increment column.
	e.curColumn++
	if e.curColumn >= cols {
		e.curColumn = 0
	}

	// Step 6: calculate curX from Columns.Positions (C# pattern).
	// C#: curX = page.Columns.Positions[curColumn] * Units.Millimeters
	const mmPerPxCol = 3.78
	if e.curColumn < len(pg.Columns.Positions) {
		e.curX = pg.Columns.Positions[e.curColumn] * mmPerPxCol
	} else {
		// Fallback: divide page evenly.
		e.curX = float32(e.curColumn) * (e.pageWidth / float32(cols))
	}

	// Step 7: if wrapped to column 0, signal caller to end page.
	if e.curColumn == 0 {
		// Step 8: paste kept objects after new page starts (handled by caller).
		if e.keeping {
			e.pasteObjects()
		}
		return false // caller should end page and start new one
	}

	// Still on the same page, start the next column.
	e.startColumn(pg)

	// Step 8: paste kept objects.
	if e.keeping {
		e.pasteObjects()
	}
	return true // advanced to next column, same page
}

// showBand is the central band-output method.
// It advances CurY by the band height, adds a PreparedBand to the current page,
// and fires the band's BeforePrint/AfterPrint hooks.
// Nil bands and invisible bands are silently skipped.
//
// C# behaviour implemented here:
//   - PageFooterBand snaps to page bottom (CurY = PageHeight - footerH) when !UnlimitedHeight
//   - OverlayBand does not advance CurY
func (e *ReportEngine) showBand(b report.Base) {
	if b == nil {
		return
	}
	// Guard against typed nil pointers passed as interface (e.g. (*band.OverlayBand)(nil)).
	if v := reflect.ValueOf(b); v.Kind() == reflect.Ptr && v.IsNil() {
		return
	}

	// Visibility check: evaluate VisibleExpression first (overrides static Visible),
	// mirroring C# CanPrint() in ReportEngine.Bands.cs (line 259) which calls
	// engine.CanPrint(band) inside PreparedPage.DoAdd and GetBandHeightWithChildren.
	type hasVisibleExprBand interface {
		Visible() bool
		VisibleExpression() string
		CalcVisibleExpression(expression string, calc func(string) (any, error)) bool
	}
	if v, ok := b.(hasVisibleExprBand); ok {
		expr := v.VisibleExpression()
		if expr != "" {
			if e.report != nil {
				visible := v.CalcVisibleExpression(expr, func(s string) (any, error) {
					return e.report.Calc(s)
				})
				if !visible {
					return
				}
			}
		} else if !v.Visible() {
			return
		}
	}

	// Use CalcBandHeight so CanGrow/CanShrink bands expand or contract based
	// on their actual text content. This handles all BandBase-derived types
	// (ReportTitle, PageHeader, etc.) via the GetBandBase() interface.
	height := e.CalcBandHeight(b)
	if height <= 0 {
		return
	}

	// C# pattern: snap PageFooterBand to the bottom of the page.
	// In AddToPreparedPages: if (band is PageFooterBand && !UnlimitedHeight)
	//     CurY = PageHeight - GetBandHeightWithChildren(band);
	if pf, isPageFooter := b.(*band.PageFooterBand); isPageFooter {
		if e.currentPage != nil && !e.currentPage.UnlimitedHeight {
			footerH := e.GetBandHeightWithChildren(&pf.BandBase)
			e.curY = e.pageHeight - footerH
			// Recalculate freeSpace so AdvanceY doesn't go negative incorrectly.
			e.freeSpace = footerH
		}
	}

	// C# pattern: OverlayBand does not advance CurY.
	_, isOverlay := b.(*band.OverlayBand)

	if e.outputBand != nil {
		// PrintOnParent mode: merge objects directly into the parent PreparedBand.
		type hasObjects interface {
			Objects() *report.ObjectCollection
		}
		if ho, ok := b.(hasObjects); ok {
			tmp := &preview.PreparedBand{}
			e.populateBandObjects2(nil, ho.Objects(), tmp)
			yOff := e.outputBandOffsetY + e.curY
			for _, po := range tmp.Objects {
				po.Left += e.outputBandOffsetX
				po.Top += yOff
				e.outputBand.Objects = append(e.outputBand.Objects, po)
			}
			newBottom := yOff + height
			if newBottom > e.outputBand.Height {
				e.outputBand.Height = newBottom
			}
		}
	} else if e.preparedPages != nil {
		pb := &preview.PreparedBand{
			Name:          b.Name(),
			Left:          e.curX,
			Top:           e.curY,
			Height:        height,
			NotExportable: bandNotExportable(b),
		}
		// Set band kind so GetLastY can exclude PageFooter and Overlay bands,
		// mirroring C# PreparedPage.GetLastY() checks.
		if _, ok := b.(*band.PageFooterBand); ok {
			pb.Kind = preview.PreparedBandKindPageFooter
		} else if _, ok := b.(*band.OverlayBand); ok {
			pb.Kind = preview.PreparedBandKindOverlay
		}
		// Populate band-level properties (width, fill, border) for background rendering.
		populateBandProps(b, pb)
		// Populate child objects from any band type that exposes Objects().
		type hasObjects interface {
			Objects() *report.ObjectCollection
		}
		if ho, ok := b.(hasObjects); ok {
			e.populateBandObjects2(extractBandBase(b), ho.Objects(), pb)
		}
		// Apply CanGrow/CanShrink adjustments to object positions and sizes.
		// This mirrors the same logic in showFullBandOnce for data bands.
		// Extract BandBase from any band type that embeds it.
		if bb := extractBandBase(b); bb != nil {
			layout := calcBandLayout(bb, bb.Height(), e.evalText)
			if layout.shifts != nil {
				applyBandObjectShifts(bb, pb, layout.shifts)
			}
			if layout.effectiveH != nil {
				applyBandObjectHeights(bb, pb, layout.effectiveH)
			}
		}
		// Apply page-level column X offset to all objects in this band.
		if e.curX != 0 {
			for i := range pb.Objects {
				pb.Objects[i].Left += e.curX
			}
		}
		// Apply subreport X offset (originX) for non-page bands.
		// Mirrors C# ReportEngine.Bands.cs AddToPreparedPages line 450:
		//   if (!isPageBand) band.Left += originX + CurX;
		// isPageBand is true for PageHeaderBand, PageFooterBand, OverlayBand.
		if e.originX != 0 {
			_, isPageBand := b.(*band.PageHeaderBand)
			if !isPageBand {
				_, isPageBand = b.(*band.PageFooterBand)
			}
			if !isPageBand {
				_, isPageBand = b.(*band.OverlayBand)
			}
			if !isPageBand {
				pb.Left += e.originX
				for i := range pb.Objects {
					pb.Objects[i].Left += e.originX
				}
			}
		}
		// Do not put page bands twice — this may happen when rendering a subreport
		// or appending one report to another. Mirrors C# AddToPreparedPages lines 484-499
		// (ReportEngine.Bands.cs):
		//   if (isPageBand) bandAlreadyExists = PreparedPages.ContainsBand(band.GetType())
		//   if (!bandAlreadyExists) PreparedPages.AddBand(band)
		bandAlreadyExists := false
		switch b.(type) {
		case *band.PageHeaderBand:
			bandAlreadyExists = e.preparedPages.ContainsBandPrefix("PageHeader")
		case *band.PageFooterBand:
			bandAlreadyExists = e.preparedPages.ContainsBandPrefix("PageFooter")
		case *band.OverlayBand:
			bandAlreadyExists = e.preparedPages.ContainsBandPrefix("Overlay")
		}
		if !bandAlreadyExists {
			_ = e.preparedPages.AddBand(pb)
		}
	}

	// C# pattern: OverlayBand does not advance CurY (it sits behind other bands).
	if !isOverlay {
		e.AdvanceY(height)
	}

	// C# ShowBand calls ShowChildBand after the parent band has been rendered.
	// This ensures the ChildBand (e.g. ReportTitleBand's Child1) appears
	// immediately after its parent at the current CurY position.
	// C# ref: ReportEngine.Bands.cs ShowBand → ShowChildBand(band).
	if bb := extractBandBase(b); bb != nil {
		if child := bb.Child(); child != nil {
			e.ShowFullBand(&child.BandBase)
		}
	}
}

// populateBandProps fills in the PreparedBand's Width, FillColor, and Border
// from the source band object. These are used by the HTML exporter to render
// the band background div (C# LayerBack pattern).
func populateBandProps(b report.Base, pb *preview.PreparedBand) {
	type hasFill interface {
		Fill() style.Fill
	}
	type hasWidth interface {
		Width() float32
	}
	type hasBorder interface {
		Border() style.Border
	}
	if w, ok := b.(hasWidth); ok {
		pb.Width = w.Width()
	}
	if f, ok := b.(hasFill); ok {
		if sf, ok2 := f.Fill().(*style.SolidFill); ok2 {
			pb.FillColor = sf.Color
		} else if gf, ok2 := f.Fill().(*style.GlassFill); ok2 {
			// GlassFill: C# ReportComponentBase.FillColor returns Color.Transparent
			// for non-SolidFill types, so the base div is transparent. The glass
			// sheen is rendered as a PNG image in a second CSS class.
			// C# ref: HTMLExportLayers.cs LayerBack → LayerPicture → dual-class.
			pb.BackgroundCSS = renderGlassFillCSS(gf, int(pb.Width), int(pb.Height))
		} else if lgf, ok2 := f.Fill().(*style.LinearGradientFill); ok2 {
			// LinearGradientFill: rendered as a PNG image in a second CSS class,
			// matching C# HTMLExportLayers.cs LayerBack → LayerPicture for non-SolidFill.
			pb.BackgroundCSS = renderLinearGradientFillCSS(lgf, int(pb.Width), int(pb.Height))
		} else if tf, ok2 := f.Fill().(*style.TextureFill); ok2 {
			// TextureFill: rendered using CSS background-image with appropriate repeat mode.
			// C# ref: HTMLExportLayers.cs LayerBack → LayerPicture for non-SolidFill.
			pb.BackgroundCSS = renderTextureFillCSS(tf, int(pb.Width), int(pb.Height))
		}
	}
	if br, ok := b.(hasBorder); ok {
		pb.Border = br.Border()
	}
}

// extractBandBase returns the *band.BandBase from any band type, or nil.
func extractBandBase(b report.Base) *band.BandBase {
	switch v := b.(type) {
	case *band.BandBase:
		return v
	case *band.ReportTitleBand:
		return &v.BandBase
	case *band.ReportSummaryBand:
		return &v.BandBase
	case *band.PageHeaderBand:
		return &v.BandBase
	case *band.PageFooterBand:
		return &v.BandBase
	case *band.ColumnHeaderBand:
		return &v.BandBase
	case *band.ColumnFooterBand:
		return &v.BandBase
	case *band.DataHeaderBand:
		return &v.BandBase
	case *band.DataFooterBand:
		return &v.BandBase
	case *band.GroupHeaderBand:
		return &v.BandBase
	case *band.GroupFooterBand:
		return &v.BandBase
	case *band.OverlayBand:
		return &v.BandBase
	case *band.DataBand:
		return &v.BandBase
	case *band.ChildBand:
		return &v.BandBase
	}
	return nil
}

// bandHeight returns the rendered height for a band in pixels.
// For bands with CanGrow, it returns the default height (full implementation
// would measure text, but this skeleton uses the declared height).
func (e *ReportEngine) bandHeight(b report.Base) float32 {
	// Type-assert to get Height from ComponentBase.
	type hasHeight interface {
		Height() float32
	}
	if h, ok := b.(hasHeight); ok {
		return h.Height()
	}
	return 0
}

// ── Report page iteration ─────────────────────────────────────────────────────

// RunReportPage processes a single ReportPage: sets up the page, runs all bands,
// and tears down the page.
//
// Matches C# RunReportPage flow:
//  1. Set current page, initReprint
//  2. StartFirstPage (startPage with isFirst=true) — or PrintOnPreviousPage path
//  3. Fire ReportPageStarted + PageStarted events
//  4. Find deepest DataBand and set KeepSummary=true
//  5. Run bands (manual build or automatic)
//  6. Fire PageFinished + ReportPageFinished events
//  7. EndLastPage (endPage with isLast=true)
func (e *ReportEngine) RunReportPage(pg *reportpkg.ReportPage) error {
	e.currentPage = pg
	e.initReprint()

	// PrintOnPreviousPage: if set and a previous prepared page exists with the
	// same dimensions, continue rendering on that page instead of creating a new one.
	// Mirrors C# StartFirstPageShared lines 200-264.
	if pg.PrintOnPreviousPage && e.preparedPages != nil && e.preparedPages.Count() > 0 {
		prevPg := e.preparedPages.GetPage(e.preparedPages.Count() - 1)
		const mmPerPx = 3.78
		curW := (pg.PaperWidth - pg.LeftMargin - pg.RightMargin) * mmPerPx
		curH := (pg.PaperHeight - pg.TopMargin - pg.BottomMargin) * mmPerPx
		sameW := pg.UnlimitedWidth || (prevPg != nil && prevPg.Width == curW)
		sameH := pg.UnlimitedHeight || (prevPg != nil && prevPg.Height == curH)
		if sameW && sameH {
			// Continue on the previous page at the last rendered Y position.
			e.curY = e.preparedPages.GetLastY()
			e.syncPageVariables()
			// Only show ReportTitle (no overlay/header) — mirrors C# StartFirstPage
			// returning previousPage=true and only calling ShowBand(page.ReportTitle).
			e.showBand(pg.ReportTitle())
			goto runBands
		}
	}

	// Start the first physical page for this template.
	e.startPage(pg, true)

runBands:

	// Fire C# events: ReportPageStarted, then PageStarted.
	e.OnStateChanged(pg, EngineStateReportPageStarted)
	e.OnStateChanged(pg, EngineStatePageStarted)

	// C#: find deepest DataBand and set KeepSummary.
	e.setKeepSummaryOnDeepestDataBand(pg)

	// Run data/group bands defined on the page.
	if err := e.runBands(pg.Bands()); err != nil {
		return err
	}

	// Fire C# events: PageFinished, then ReportPageFinished.
	e.OnStateChanged(pg, EngineStatePageFinished)
	e.OnStateChanged(pg, EngineStateReportPageFinished)

	// End last page (includes ColumnFinished event, summary, footer).
	e.endPage(pg, true)
	return nil
}

// setKeepSummaryOnDeepestDataBand finds the deepest (last) DataBand on a page
// and sets its KeepSummary flag to true. This matches C#'s FindDeepmostDataBand.
func (e *ReportEngine) setKeepSummaryOnDeepestDataBand(pg *reportpkg.ReportPage) {
	var deepest *band.DataBand
	for _, b := range pg.AllBands() {
		if db, ok := b.(*band.DataBand); ok {
			deepest = db
		}
	}
	if deepest != nil {
		deepest.SetKeepSummary(true)
	}
}

// runBands iterates a slice of bands, dispatching DataBands and GroupHeaders.
func (e *ReportEngine) runBands(bands []report.Base) error {
	for _, b := range bands {
		if e.aborted {
			break
		}
		switch v := b.(type) {
		case *band.DataBand:
			// Skip invisible DataBands (e.g. drill-down groups where the data band
			// is hidden by default and shown only on interactive click).
			if !v.Visible() {
				continue
			}
			if err := e.RunDataBandFull(v); err != nil {
				return err
			}
		case *band.GroupHeaderBand:
			e.RunGroup(v)
		default:
			e.showBand(b)
		}
	}
	return nil
}

// attachWatermark converts the ReportPage watermark into a PreparedWatermark and
// attaches it to the most recently added PreparedPage.
func (e *ReportEngine) attachWatermark(pg *reportpkg.ReportPage) {
	if e.preparedPages == nil {
		return
	}
	cur := e.preparedPages.CurrentPage()
	if cur == nil {
		return
	}
	wm := pg.Watermark
	// Evaluate bracket expressions in the watermark text, matching C# TextObject.GetData()
	// called from DrawText (Watermark.cs:275). System variables like [Page#]/[TotalPages#]
	// and data column references are expanded here at page-render time.
	wmText := e.evalText(wm.Text)
	pw := &preview.PreparedWatermark{
		Enabled:           wm.Enabled,
		Text:              wmText,
		Font:              wm.Font,
		TextColor:         wm.TextFillColor,
		TextRotation:      preview.WatermarkTextRotation(wm.TextRotation),
		ShowTextOnTop:     wm.ShowTextOnTop,
		ImageBlobIdx:      -1,
		ImageSize:         preview.WatermarkImageSize(wm.ImageSize),
		ImageTransparency: wm.ImageTransparency,
		ShowImageOnTop:    wm.ShowImageOnTop,
	}
	if len(wm.ImageData) > 0 {
		pw.ImageBlobIdx = e.preparedPages.BlobStore.Add("__wm_"+pg.Name(), wm.ImageData)
	}
	cur.Watermark = pw
}
