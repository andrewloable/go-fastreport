package preview

// SourcePages tracks the correspondence between source report page indices
// and the ranges of prepared (rendered) page indices they produced.
//
// It is the Go equivalent of FastReport.Preview.SourcePages, used by the
// double-pass rendering engine to:
//  1. Record, during the first pass, which source page generated each output page.
//  2. Look up, during the second pass, which output page range belongs to each
//     source page so the engine can overwrite them with correct values
//     (e.g. TotalPages expressions resolved after the first pass).
//
// Example double-pass flow:
//
//	sp := preview.NewSourcePages()
//
//	// First pass
//	for _, srcPage := range report.Pages() {
//	    firstOutput := preparedPages.Count()
//	    engine.RenderPage(srcPage)          // appends prepared pages
//	    lastOutput := preparedPages.Count() - 1
//	    sp.Record(srcPageIndex, firstOutput, lastOutput)
//	}
//
//	// Second pass — rewrite each source page's output range
//	for srcIdx := range report.Pages() {
//	    start, end, ok := sp.Range(srcIdx)
//	    if ok {
//	        engine.RewriteRange(srcIdx, start, end)
//	    }
//	}
type SourcePages struct {
	// entries maps source page index → the range of prepared page indices it generated.
	entries []sourcePageEntry
}

// sourcePageEntry records the prepared-page range for one source page.
type sourcePageEntry struct {
	// sourceIdx is the index of the source ReportPage.
	sourceIdx int
	// firstOutput is the zero-based index of the first PreparedPage generated.
	firstOutput int
	// lastOutput is the zero-based index of the last PreparedPage generated.
	lastOutput int
}

// NewSourcePages creates an empty SourcePages tracker.
func NewSourcePages() *SourcePages {
	return &SourcePages{}
}

// Record registers that source page at srcIdx generated prepared pages in the
// inclusive range [firstOutput, lastOutput]. Calling Record with the same
// srcIdx a second time replaces the previous entry.
func (sp *SourcePages) Record(srcIdx, firstOutput, lastOutput int) {
	for i := range sp.entries {
		if sp.entries[i].sourceIdx == srcIdx {
			sp.entries[i].firstOutput = firstOutput
			sp.entries[i].lastOutput = lastOutput
			return
		}
	}
	sp.entries = append(sp.entries, sourcePageEntry{
		sourceIdx:   srcIdx,
		firstOutput: firstOutput,
		lastOutput:  lastOutput,
	})
}

// Range returns the prepared-page range [first, last] for the given source page
// index. Returns (0, 0, false) if the source page has not been recorded.
func (sp *SourcePages) Range(srcIdx int) (first, last int, ok bool) {
	for _, e := range sp.entries {
		if e.sourceIdx == srcIdx {
			return e.firstOutput, e.lastOutput, true
		}
	}
	return 0, 0, false
}

// Count returns the number of recorded source page entries.
func (sp *SourcePages) Count() int { return len(sp.entries) }

// Clear removes all entries.
func (sp *SourcePages) Clear() { sp.entries = sp.entries[:0] }

// SourceIndices returns all source page indices in insertion order.
func (sp *SourcePages) SourceIndices() []int {
	out := make([]int, len(sp.entries))
	for i, e := range sp.entries {
		out[i] = e.sourceIdx
	}
	return out
}

// RemoveLast removes the most recently recorded entry, mirroring the C#
// SourcePages.RemoveLast() used by the engine when a page is rolled back.
func (sp *SourcePages) RemoveLast() {
	if len(sp.entries) > 0 {
		sp.entries = sp.entries[:len(sp.entries)-1]
	}
}
