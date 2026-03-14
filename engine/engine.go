// Package engine implements the go-fastreport report engine.
// It drives report execution: initialising data sources, iterating bands,
// and producing a stream of prepared pages ready for export.
package engine

import (
	"fmt"
	"time"

	"github.com/andrewloable/go-fastreport/data"
	"github.com/andrewloable/go-fastreport/reportpkg"
)

// RunOptions controls optional aspects of a report run.
type RunOptions struct {
	// Append appends output to existing prepared pages instead of replacing them.
	Append bool
	// ResetDataState re-initialises data sources before running.
	ResetDataState bool
	// MaxPages limits the number of output pages (0 = unlimited).
	MaxPages int
}

// DefaultRunOptions returns RunOptions with the most common defaults.
func DefaultRunOptions() RunOptions {
	return RunOptions{ResetDataState: true}
}

// ReportEngine drives the execution of a Report.
// It is the Go equivalent of FastReport.Engine.ReportEngine.
type ReportEngine struct {
	// report is the report definition being executed.
	report *reportpkg.Report

	// curX / curY track the current print position on the page (pixels).
	curX float32
	curY float32

	// curColumn is the zero-based index of the current multi-column slot.
	curColumn int

	// date is the snapshot of "now" captured at the start of Phase 1.
	date time.Time

	// finalPass is true during the optional second (double-pass) run.
	finalPass bool

	// pageNo is the logical (displayed) page number of the current page.
	pageNo int

	// totalPages is the total number of logical pages in the prepared output.
	totalPages int

	// rowNo is the current data row number within the current data band (1-based).
	rowNo int

	// absRowNo is the absolute row number across all data bands (1-based).
	absRowNo int

	// hierarchyLevel is the nesting depth in hierarchical reports.
	hierarchyLevel int

	// hierarchyRowNo is a dot-separated string like "1.2.3" for hierarchical reports.
	hierarchyRowNo string

	// pagesLimit is the maximum number of prepared pages (0 = unlimited).
	pagesLimit int

	// pageWidth / pageHeight are the usable dimensions of the current page (pixels).
	// These are set from the current ReportPage's paper size minus margins.
	pageWidth  float32
	pageHeight float32

	// freeSpace is the remaining vertical space on the current page (pixels).
	freeSpace float32

	// dataSources is the flat list of data sources registered for this run.
	dataSources []*data.BaseDataSource

	// startReportFired tracks whether the OnStartReport event has already been fired.
	startReportFired bool

	// aborted is set to true if the run is cancelled.
	aborted bool
}

// New creates a ReportEngine for the given report.
func New(r *reportpkg.Report) *ReportEngine {
	return &ReportEngine{
		report:  r,
		pageNo:  1,
		rowNo:   1,
		absRowNo: 1,
	}
}

// ── Properties ────────────────────────────────────────────────────────────────

// CurX returns the current horizontal print position in pixels.
func (e *ReportEngine) CurX() float32 { return e.curX }

// SetCurX sets the current horizontal print position.
func (e *ReportEngine) SetCurX(v float32) { e.curX = v }

// CurY returns the current vertical print position in pixels.
func (e *ReportEngine) CurY() float32 { return e.curY }

// SetCurY sets the current vertical print position.
func (e *ReportEngine) SetCurY(v float32) { e.curY = v }

// CurColumn returns the zero-based index of the current print column.
func (e *ReportEngine) CurColumn() int { return e.curColumn }

// PageWidth returns the usable page width (paper width minus left+right margins) in pixels.
func (e *ReportEngine) PageWidth() float32 { return e.pageWidth }

// PageHeight returns the usable page height (paper height minus top+bottom margins) in pixels.
func (e *ReportEngine) PageHeight() float32 { return e.pageHeight }

// FreeSpace returns the remaining vertical space on the current page in pixels.
func (e *ReportEngine) FreeSpace() float32 { return e.freeSpace }

// PageNo returns the current logical page number (1-based).
func (e *ReportEngine) PageNo() int { return e.pageNo }

// TotalPages returns the total number of logical pages prepared so far.
func (e *ReportEngine) TotalPages() int { return e.totalPages }

// RowNo returns the current data row number within the current data band (1-based).
func (e *ReportEngine) RowNo() int { return e.rowNo }

// AbsRowNo returns the absolute data row number across all bands (1-based).
func (e *ReportEngine) AbsRowNo() int { return e.absRowNo }

// FinalPass returns true when the engine is in the second (double) pass.
func (e *ReportEngine) FinalPass() bool { return e.finalPass }

// FirstPass returns true when the engine is in the first (or only) pass.
// In single-pass mode this is always true.
func (e *ReportEngine) FirstPass() bool {
	return !(e.report.DoublePass && e.finalPass)
}

// HierarchyLevel returns the nesting depth in hierarchical reports (0-based).
func (e *ReportEngine) HierarchyLevel() int { return e.hierarchyLevel }

// HierarchyRowNo returns the dot-separated hierarchy row identifier (e.g. "1.2.3").
func (e *ReportEngine) HierarchyRowNo() string { return e.hierarchyRowNo }

// Date returns the date/time snapshot captured at the start of the run.
func (e *ReportEngine) Date() time.Time { return e.date }

// Aborted returns true if the run has been cancelled.
func (e *ReportEngine) Aborted() bool { return e.aborted }

// Abort signals the engine to stop processing after the current band.
func (e *ReportEngine) Abort() { e.aborted = true }

// ── Run ───────────────────────────────────────────────────────────────────────

// Run executes the report with the provided options.
// It runs Phase 1 (data initialisation), Phase 2 (page generation), and
// Phase 3 (finish events and page limiting).
// Returns an error if any phase fails.
func (e *ReportEngine) Run(opts RunOptions) error {
	e.pagesLimit = opts.MaxPages

	if err := e.runPhase1(opts.ResetDataState); err != nil {
		e.runFinished()
		return fmt.Errorf("engine phase 1: %w", err)
	}
	if err := e.runPhase2(opts.Append); err != nil {
		e.runFinished()
		return fmt.Errorf("engine phase 2: %w", err)
	}
	e.runFinished()
	return nil
}

// runPhase1 initialises data sources and fires the OnStartReport event.
func (e *ReportEngine) runPhase1(resetDataState bool) error {
	e.date = time.Now()
	e.aborted = false
	e.finalPass = false
	e.startReportFired = false

	if resetDataState {
		if err := e.initializeData(); err != nil {
			return err
		}
	}

	// Fire OnStartReport event.
	if e.report.StartReportEvent != "" {
		// Event firing is handled by the script engine in a full implementation.
		// Here we record that it was fired.
	}
	e.startReportFired = true
	return nil
}

// runPhase2 runs one or two passes of page generation.
func (e *ReportEngine) runPhase2(append bool) error {
	if err := e.prepareToFirstPass(append); err != nil {
		return err
	}
	if err := e.runReportPages(); err != nil {
		return err
	}

	// Double-pass: run a second (final) pass to resolve TotalPages references.
	if e.report.DoublePass && !e.aborted {
		e.finalPass = true
		e.resetPageNumber()
		if err := e.prepareToSecondPass(); err != nil {
			return err
		}
		if err := e.runReportPages(); err != nil {
			return err
		}
	}
	return nil
}

// runFinished fires the OnFinishReport event and enforces the page limit.
func (e *ReportEngine) runFinished() {
	// Fire OnFinishReport event (script engine in full implementation).
	if e.report.FinishReportEvent != "" {
		// event would be fired here
	}
	e.limitPreparedPages()
}

// initializeData opens all registered data sources.
func (e *ReportEngine) initializeData() error {
	for _, ds := range e.dataSources {
		if err := ds.Init(); err != nil {
			return fmt.Errorf("data source %q: %w", ds.Name(), err)
		}
	}
	return nil
}

// prepareToFirstPass resets engine state for the first (or only) pass.
func (e *ReportEngine) prepareToFirstPass(append bool) error {
	if !append {
		e.pageNo = e.report.InitialPageNumber
		e.totalPages = 0
		e.curY = 0
		e.curX = 0
		e.curColumn = 0
		e.rowNo = 1
		e.absRowNo = 1
	}
	e.freeSpace = e.pageHeight
	return nil
}

// prepareToSecondPass resets positioning for the second pass.
func (e *ReportEngine) prepareToSecondPass() error {
	e.curY = 0
	e.curX = 0
	e.curColumn = 0
	e.rowNo = 1
	e.absRowNo = 1
	e.freeSpace = e.pageHeight
	return nil
}

// runReportPages iterates through the report's ReportPages and generates output.
// In the full engine implementation this method drives band printing; here it
// establishes page dimensions and counts pages.
func (e *ReportEngine) runReportPages() error {
	for _, pg := range e.report.Pages() {
		if e.aborted {
			break
		}
		if e.pagesLimit > 0 && e.totalPages >= e.pagesLimit {
			break
		}
		if err := e.processPage(pg); err != nil {
			return err
		}
	}
	return nil
}

// processPage sets up the current page dimensions and increments counters.
func (e *ReportEngine) processPage(pg *reportpkg.ReportPage) error {
	// Compute usable dimensions (paper minus margins, converted mm→px at 96 dpi).
	const mmPerPx = 96.0 / 25.4
	e.pageWidth = (pg.PaperWidth - pg.LeftMargin - pg.RightMargin) * mmPerPx
	e.pageHeight = (pg.PaperHeight - pg.TopMargin - pg.BottomMargin) * mmPerPx
	e.freeSpace = e.pageHeight

	e.totalPages++
	e.pageNo++
	return nil
}

// resetPageNumber resets the logical page counter for the second pass.
func (e *ReportEngine) resetPageNumber() {
	e.pageNo = e.report.InitialPageNumber
}

// limitPreparedPages enforces the MaxPages limit.
func (e *ReportEngine) limitPreparedPages() {
	if e.pagesLimit > 0 && e.totalPages > e.pagesLimit {
		e.totalPages = e.pagesLimit
	}
}

// ── Data source registration ──────────────────────────────────────────────────

// RegisterDataSource adds a data source to be initialised during Phase 1.
func (e *ReportEngine) RegisterDataSource(ds *data.BaseDataSource) {
	e.dataSources = append(e.dataSources, ds)
}

// DataSources returns the registered data sources.
func (e *ReportEngine) DataSources() []*data.BaseDataSource {
	return e.dataSources
}

// ── Position helpers ──────────────────────────────────────────────────────────

// AdvanceY moves the vertical position down by delta pixels and reduces free space.
func (e *ReportEngine) AdvanceY(delta float32) {
	e.curY += delta
	e.freeSpace -= delta
	if e.freeSpace < 0 {
		e.freeSpace = 0
	}
}

// NewPage resets the vertical position for a new page and increments counters.
func (e *ReportEngine) NewPage() {
	e.curY = 0
	e.curX = 0
	e.curColumn = 0
	e.freeSpace = e.pageHeight
	e.totalPages++
	e.pageNo++
}
