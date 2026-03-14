package engine

import "github.com/andrewloable/go-fastreport/band"

// reprintEntry holds a band and its originX offset at the time it was
// registered for reprinting. originX supports subreport column offsets.
type reprintEntry struct {
	b       *band.BandBase
	originX float32
}

// initReprint initialises the reprint header/footer lists.
func (e *ReportEngine) initReprint() {
	e.reprintHeaders = nil
	e.reprintFooters = nil
	e.keepReprintHeaders = nil
	e.keepReprintFooters = nil
}

// AddReprint registers a BandBase for reprinting on each new page.
// DataHeader and GroupHeader bands are treated as "headers" (printed at the
// top of the new page); everything else is a footer (printed before the page break).
func (e *ReportEngine) AddReprint(b *band.BandBase) {
	entry := reprintEntry{b: b, originX: 0} // originX support can be extended
	if e.keeping {
		switch b.Name() {
		default:
			e.keepReprintFooters = append(e.keepReprintFooters, entry)
		}
		// A full implementation would type-switch on DataHeaderBand / GroupHeaderBand.
		// For now all bands go into keep footers.
		_ = entry
		return
	}
	e.reprintFooters = append(e.reprintFooters, entry)
}

// RemoveReprint unregisters a band from both header and footer reprint lists.
func (e *ReportEngine) RemoveReprint(b *band.BandBase) {
	if b == nil {
		return
	}
	e.reprintHeaders = removeReprintEntry(e.reprintHeaders, b)
	e.reprintFooters = removeReprintEntry(e.reprintFooters, b)
	e.keepReprintHeaders = removeReprintEntry(e.keepReprintHeaders, b)
	e.keepReprintFooters = removeReprintEntry(e.keepReprintFooters, b)
}

func removeReprintEntry(list []reprintEntry, b *band.BandBase) []reprintEntry {
	out := list[:0]
	for _, entry := range list {
		if entry.b != b {
			out = append(out, entry)
		}
	}
	return out
}

// ShowReprintHeaders renders the registered header bands at the top of the new page.
func (e *ReportEngine) ShowReprintHeaders() {
	for _, entry := range e.reprintHeaders {
		e.ShowFullBand(entry.b)
	}
}

// ShowReprintFooters renders the registered footer bands (in reverse order)
// before a page break.
func (e *ReportEngine) ShowReprintFooters() {
	for i := len(e.reprintFooters) - 1; i >= 0; i-- {
		e.ShowFullBand(e.reprintFooters[i].b)
	}
}

// startKeepReprint clears the keep-scoped reprint lists.
// Called by StartKeep.
func (e *ReportEngine) startKeepReprint() {
	e.keepReprintHeaders = nil
	e.keepReprintFooters = nil
}

// endKeepReprint merges keep-scoped lists into the main reprint lists.
// Called by EndKeep.
func (e *ReportEngine) endKeepReprint() {
	e.reprintHeaders = append(e.reprintHeaders, e.keepReprintHeaders...)
	e.reprintFooters = append(e.reprintFooters, e.keepReprintFooters...)
	e.keepReprintHeaders = nil
	e.keepReprintFooters = nil
}

// ReprintHeaderCount returns the number of reprint header entries (for testing).
func (e *ReportEngine) ReprintHeaderCount() int { return len(e.reprintHeaders) }

// ReprintFooterCount returns the number of reprint footer entries (for testing).
func (e *ReportEngine) ReprintFooterCount() int { return len(e.reprintFooters) }
