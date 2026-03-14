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

	if e.preparedPages != nil {
		e.preparedPages.AddPage(e.pageWidth, e.pageHeight, e.pageNo)
	}

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
	} else {
		e.showBand(pg.Overlay())
		e.showBand(pg.PageHeader())
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

	if e.preparedPages != nil {
		pb := &preview.PreparedBand{
			Name:   b.Name(),
			Top:    e.curY,
			Height: height,
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
