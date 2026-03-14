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
	entry := reprintEntry{b: b, originX: 0}
	if e.keeping {
		e.keepReprintFooters = append(e.keepReprintFooters, entry)
		return
	}
	e.reprintFooters = append(e.reprintFooters, entry)
}

// addReprintBand registers a concrete band object for reprinting, correctly
// classifying DataHeaderBand and GroupHeaderBand as headers and footers otherwise.
// This should be used instead of AddReprint when the concrete type is available.
func (e *ReportEngine) addReprintBand(b *band.BandBase, isHeader bool) {
	entry := reprintEntry{b: b, originX: 0}
	if e.keeping {
		if isHeader {
			e.keepReprintHeaders = append(e.keepReprintHeaders, entry)
		} else {
			e.keepReprintFooters = append(e.keepReprintFooters, entry)
		}
		return
	}
	if isHeader {
		e.reprintHeaders = append(e.reprintHeaders, entry)
	} else {
		e.reprintFooters = append(e.reprintFooters, entry)
	}
}

// AddReprintDataHeader registers a DataHeaderBand for reprinting as a header.
// Called when the band has RepeatOnEveryPage=true.
func (e *ReportEngine) AddReprintDataHeader(b *band.DataHeaderBand) {
	if b == nil {
		return
	}
	e.addReprintBand(&b.HeaderFooterBandBase.BandBase, true)
}

// AddReprintGroupHeader registers a GroupHeaderBand for reprinting as a header.
// Called when the band has RepeatOnEveryPage=true.
func (e *ReportEngine) AddReprintGroupHeader(b *band.GroupHeaderBand) {
	if b == nil {
		return
	}
	e.addReprintBand(&b.HeaderFooterBandBase.BandBase, true)
}

// AddReprintDataFooter registers a DataFooterBand for reprinting as a footer.
// Called when the band has RepeatOnEveryPage=true.
func (e *ReportEngine) AddReprintDataFooter(b *band.DataFooterBand) {
	if b == nil {
		return
	}
	e.addReprintBand(&b.HeaderFooterBandBase.BandBase, false)
}

// AddReprintGroupFooter registers a GroupFooterBand for reprinting as a footer.
// Called when the band has RepeatOnEveryPage=true.
func (e *ReportEngine) AddReprintGroupFooter(b *band.GroupFooterBand) {
	if b == nil {
		return
	}
	e.addReprintBand(&b.HeaderFooterBandBase.BandBase, false)
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
