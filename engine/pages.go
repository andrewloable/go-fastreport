package engine

import (
	"reflect"

	"github.com/andrewloable/go-fastreport/band"
	"github.com/andrewloable/go-fastreport/preview"
	"github.com/andrewloable/go-fastreport/report"
	"github.com/andrewloable/go-fastreport/reportpkg"
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
func (e *ReportEngine) startPage(pg *reportpkg.ReportPage, isFirst bool) {
	const mmPerPx = 96.0 / 25.4
	e.pageWidth = (pg.PaperWidth - pg.LeftMargin - pg.RightMargin) * mmPerPx
	e.pageHeight = (pg.PaperHeight - pg.TopMargin - pg.BottomMargin) * mmPerPx
	e.freeSpace = e.pageHeight
	e.curX = 0
	e.curY = 0
	e.curColumn = 0

	e.totalPages++
	e.pageNo++

	// Keep system variables in sync.
	e.syncPageVariables()

	if e.preparedPages != nil {
		e.preparedPages.AddPage(e.pageWidth, e.pageHeight, e.pageNo)
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
		// Show overlay band if present.
		e.showBand(pg.Overlay())
		// Title and PageHeader order depends on TitleBeforeHeader flag.
		if pg.TitleBeforeHeader {
			e.showBand(pg.ReportTitle())
			e.showBand(pg.PageHeader())
		} else {
			e.showBand(pg.PageHeader())
			e.showBand(pg.ReportTitle())
		}
		// Reprint headers after the page header on every page (first page too).
		e.ShowReprintHeaders()
	} else {
		e.showBand(pg.Overlay())
		e.showBand(pg.PageHeader())
		// Reprint data/group headers that have RepeatOnEveryPage=true.
		e.ShowReprintHeaders()
	}

	// Column header below the page header.
	e.columnStartY = e.curY
	e.showBand(pg.ColumnHeader())
}

// endPage finalises the current page.
// It shows ColumnFooter, ReportSummary (on last page), and PageFooter.
func (e *ReportEngine) endPage(pg *reportpkg.ReportPage, isLast bool) {
	e.showBand(pg.ColumnFooter())
	if isLast {
		e.showBand(pg.ReportSummary())
	}
	e.showBand(pg.PageFooter())
	e.OnStateChanged(pg, EngineStatePageFinished)
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
			Name:   b.Name() + "_back",
			Top:    e.curY,
			Height: height,
		}
		type hasObjects interface{ Objects() *report.ObjectCollection }
		if ho, ok := b.(hasObjects); ok {
			e.populateBandObjects2(ho.Objects(), pb)
		}
		_ = e.preparedPages.AddBand(pb)
	}
	// Advance local cursor within back-page rendering so consecutive bands stack.
	e.curY += height
}

// startColumn initialises the current column.
func (e *ReportEngine) startColumn(pg *reportpkg.ReportPage) {
	e.curY = e.columnStartY
	e.showBand(pg.ColumnHeader())
}

// endColumn finalises the current column. If there are more columns, advance to the
// next column; otherwise end the page and start a new one.
func (e *ReportEngine) endColumn(pg *reportpkg.ReportPage) bool {
	cols := pg.Columns.Count
	if cols <= 1 {
		return false
	}
	e.OnStateChanged(pg, EngineStateColumnFinished)
	e.curColumn++
	if e.curColumn >= cols {
		e.curColumn = 0
		return false // caller should start new page
	}
	e.curX = float32(e.curColumn) * (e.pageWidth / float32(cols))
	e.curY = e.columnStartY
	return true // advanced to next column, same page
}

// showBand is the central band-output method.
// It advances CurY by the band height, adds a PreparedBand to the current page,
// and fires the band's BeforePrint/AfterPrint hooks.
// Nil bands and invisible bands are silently skipped.
func (e *ReportEngine) showBand(b report.Base) {
	if b == nil {
		return
	}
	// Guard against typed nil pointers passed as interface (e.g. (*band.OverlayBand)(nil)).
	if v := reflect.ValueOf(b); v.Kind() == reflect.Ptr && v.IsNil() {
		return
	}

	// Visibility check: skip bands where Visible() == false.
	type hasVisible interface {
		Visible() bool
	}
	if vis, ok := b.(hasVisible); ok && !vis.Visible() {
		return
	}

	height := e.bandHeight(b)
	if height <= 0 {
		return
	}

	if e.outputBand != nil {
		// PrintOnParent mode: merge objects directly into the parent PreparedBand.
		type hasObjects interface {
			Objects() *report.ObjectCollection
		}
		if ho, ok := b.(hasObjects); ok {
			tmp := &preview.PreparedBand{}
			e.populateBandObjects2(ho.Objects(), tmp)
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
			Name:   b.Name(),
			Top:    e.curY,
			Height: height,
		}
		// Populate child objects from any band type that exposes Objects().
		type hasObjects interface {
			Objects() *report.ObjectCollection
		}
		if ho, ok := b.(hasObjects); ok {
			e.populateBandObjects2(ho.Objects(), pb)
		}
		// Apply page-level column X offset to all objects in this band.
		if e.curX != 0 {
			for i := range pb.Objects {
				pb.Objects[i].Left += e.curX
			}
		}
		_ = e.preparedPages.AddBand(pb)
	}

	e.AdvanceY(height)
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
func (e *ReportEngine) RunReportPage(pg *reportpkg.ReportPage) error {
	e.currentPage = pg

	// Start the first physical page for this template.
	e.startPage(pg, true)

	// Run data/group bands defined on the page.
	if err := e.runBands(pg.Bands()); err != nil {
		return err
	}

	// End page.
	e.endPage(pg, true)
	e.OnStateChanged(pg, EngineStateReportPageFinished)
	return nil
}

// runBands iterates a slice of bands, dispatching DataBands and GroupHeaders.
func (e *ReportEngine) runBands(bands []report.Base) error {
	for _, b := range bands {
		if e.aborted {
			break
		}
		switch v := b.(type) {
		case *band.DataBand:
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
	pw := &preview.PreparedWatermark{
		Enabled:           wm.Enabled,
		Text:              wm.Text,
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
