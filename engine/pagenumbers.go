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
// Called at the start of each run phase.
func (e *ReportEngine) initPageNumbers() {
	e.pageNumbers = nil
	e.logicalPageNo = 0
}

// IncLogicalPageNumber advances the logical page counter and records an entry
// for the current physical page.
func (e *ReportEngine) IncLogicalPageNumber() {
	e.logicalPageNo++
	e.pageNumbers = append(e.pageNumbers, pageNumberInfo{pageNo: e.logicalPageNo})
}

// ResetLogicalPageNumber resets the logical page counter (e.g. at a new group
// that has ResetPageNumber=true). It also back-fills the totalPages field for
// all entries since the last reset so that "[TotalPages]" resolves correctly.
func (e *ReportEngine) ResetLogicalPageNumber() {
	// Back-fill totalPages for the current group.
	for i := len(e.pageNumbers) - 1; i >= 0; i-- {
		e.pageNumbers[i].totalPages = e.logicalPageNo
		if e.pageNumbers[i].pageNo == 1 {
			break
		}
	}
	e.logicalPageNo = 0
}

// LogicalPageNo returns the current logical page number (1-based).
func (e *ReportEngine) LogicalPageNo() int { return e.logicalPageNo }

// LogicalPageCount returns the total number of logical page entries recorded.
func (e *ReportEngine) LogicalPageCount() int { return len(e.pageNumbers) }
