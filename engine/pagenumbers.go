package engine

// pageNumberInfo tracks the logical page number and total-pages value for
// each physical page in the prepared output.  The total-pages field is filled
// in at the end of the run (or group reset) so that "[TotalPages]" placeholders
// can be resolved in a second pass.
type pageNumberInfo struct {
	pageNo     int
	totalPages int
}

// initPageNumbers initialises the page-number tracking state.
// Called at the start of each run phase (mirrors C# InitPageNumbers).
func (e *ReportEngine) initPageNumbers() {
	e.pageNumbers = nil
	e.logicalPageNo = 0
}

// IncLogicalPageNumber advances the logical page counter and records an entry
// for the current physical page.
//
// C# behaviour: during the first pass (or single pass) a new entry is always
// appended. During the second pass of a double-pass report, existing entries
// are reused unless the second pass produces more pages than the first (which
// happens when content grows). In the latter case a new entry is appended.
func (e *ReportEngine) IncLogicalPageNumber() {
	e.logicalPageNo++
	index := e.curPage - e.firstReportPage
	if e.FirstPass() || index >= len(e.pageNumbers) {
		e.pageNumbers = append(e.pageNumbers, pageNumberInfo{pageNo: e.logicalPageNo})
	}
	// During the second pass with index in range, we reuse the existing entry
	// (its totalPages was already back-filled during the first pass).
}

// ResetLogicalPageNumber resets the logical page counter (e.g. at a new group
// that has ResetPageNumber=true, or at the end of a report page with
// ResetPageNumber). It also back-fills the totalPages field for all entries
// since the last reset so that "[TotalPages]" resolves correctly.
//
// C# behaviour: this only runs during the first pass. During the second pass
// the pageNumbers list already has correct totalPages values from the first
// pass, so re-running the back-fill would be incorrect.
func (e *ReportEngine) ResetLogicalPageNumber() {
	if !e.FirstPass() {
		return
	}
	// Back-fill totalPages for the current group.
	for i := len(e.pageNumbers) - 1; i >= 0; i-- {
		e.pageNumbers[i].totalPages = e.logicalPageNo
		if e.pageNumbers[i].pageNo == 1 {
			break
		}
	}
	e.logicalPageNo = 0
}

// GetLogicalPageNumber returns the logical page number for the current page,
// adjusted by the report's InitialPageNumber setting.
// This mirrors C# ReportEngine.PageNo which calls GetLogicalPageNumber().
func (e *ReportEngine) GetLogicalPageNumber() int {
	index := e.curPage - e.firstReportPage
	if index < 0 || index >= len(e.pageNumbers) {
		// Fallback: use the raw pageNo when the pageNumbers list hasn't been
		// populated yet (e.g. before the first AddPage call).
		return e.pageNo
	}
	return e.pageNumbers[index].pageNo + e.report.InitialPageNumber - 1
}

// GetLogicalTotalPages returns the total number of logical pages for the
// current page's group, adjusted by InitialPageNumber.
// This mirrors C# ReportEngine.TotalPages which calls GetLogicalTotalPages().
func (e *ReportEngine) GetLogicalTotalPages() int {
	index := e.curPage - e.firstReportPage
	if index < 0 || index >= len(e.pageNumbers) {
		// Fallback: use the raw totalPages or knownTotalPages.
		tp := e.totalPages
		if e.knownTotalPages > 0 {
			tp = e.knownTotalPages
		}
		return tp
	}
	return e.pageNumbers[index].totalPages + e.report.InitialPageNumber - 1
}

// ShiftLastPage is called when the number of pages increased during the second
// pass of a DoublePass report (compared to the first pass). It appends a new
// entry and recalculates totalPages for all entries.
//
// C# equivalent: ReportEngine.ShiftLastPage()
func (e *ReportEngine) ShiftLastPage() {
	info := pageNumberInfo{pageNo: len(e.pageNumbers) + 1}
	e.pageNumbers = append(e.pageNumbers, info)

	// Recalculate totalPages for all entries to reflect the new count.
	for i := range e.pageNumbers {
		e.pageNumbers[i].totalPages = len(e.pageNumbers)
	}
}

// LogicalPageNo returns the current logical page number (1-based).
func (e *ReportEngine) LogicalPageNo() int { return e.logicalPageNo }

// LogicalPageCount returns the total number of logical page entries recorded.
func (e *ReportEngine) LogicalPageCount() int { return len(e.pageNumbers) }

// CurPageIndex returns the zero-based index of the current prepared page.
func (e *ReportEngine) CurPageIndex() int { return e.curPage }

// SetCurPageIndex sets the current prepared page index (used by subreports/tables).
func (e *ReportEngine) SetCurPageIndex(v int) { e.curPage = v }
