package band

import "github.com/andrewloable/go-fastreport/report"

// HeaderFooterBandBase is the base for band types that support
// "Keep With Data" and "Repeat on Every Page". It is the Go equivalent of
// FastReport.HeaderFooterBandBase.
//
// Used by DataHeaderBand, DataFooterBand, GroupHeaderBand, GroupFooterBand,
// ColumnHeaderBand, ColumnFooterBand.
type HeaderFooterBandBase struct {
	BandBase

	// keepWithData causes this band to stay on the same page as the next data band.
	keepWithData bool

	// repeatOnEveryPage reprints this band at the top of each new page.
	repeatOnEveryPage bool
}

// NewHeaderFooterBandBase creates a HeaderFooterBandBase with defaults.
func NewHeaderFooterBandBase() *HeaderFooterBandBase {
	return &HeaderFooterBandBase{
		BandBase: *NewBandBase(),
	}
}

// KeepWithData returns whether the band must stay with the next data band.
func (h *HeaderFooterBandBase) KeepWithData() bool { return h.keepWithData }

// SetKeepWithData sets the keep-with-data flag.
func (h *HeaderFooterBandBase) SetKeepWithData(v bool) { h.keepWithData = v }

// RepeatOnEveryPage returns whether the band is reprinted on every new page.
func (h *HeaderFooterBandBase) RepeatOnEveryPage() bool { return h.repeatOnEveryPage }

// SetRepeatOnEveryPage sets the repeat-on-every-page flag.
func (h *HeaderFooterBandBase) SetRepeatOnEveryPage(v bool) { h.repeatOnEveryPage = v }

// Serialize writes HeaderFooterBandBase properties that differ from defaults.
// serializeAttrs writes HeaderFooterBandBase attributes only (no children).
func (h *HeaderFooterBandBase) serializeAttrs(w report.Writer) error {
	if err := h.BandBase.serializeAttrs(w); err != nil {
		return err
	}
	if h.keepWithData {
		w.WriteBool("KeepWithData", true)
	}
	if h.repeatOnEveryPage {
		w.WriteBool("RepeatOnEveryPage", true)
	}
	return nil
}

func (h *HeaderFooterBandBase) Serialize(w report.Writer) error {
	if err := h.serializeAttrs(w); err != nil {
		return err
	}
	return h.BandBase.serializeChildren(w)
}

// Deserialize reads HeaderFooterBandBase properties.
func (h *HeaderFooterBandBase) Deserialize(r report.Reader) error {
	if err := h.BandBase.Deserialize(r); err != nil {
		return err
	}
	h.keepWithData = r.ReadBool("KeepWithData", false)
	h.repeatOnEveryPage = r.ReadBool("RepeatOnEveryPage", false)
	return nil
}
