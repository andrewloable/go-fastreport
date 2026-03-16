package band

// band_headerfooter_coverage_test.go — supplementary internal tests for
// header_footer.go that document and exercise the remaining coverage gaps.
//
// Remaining uncovered branches (confirmed from coverage profile):
//
//   header_footer.go:43-45  serializeAttrs: if err := h.BandBase.serializeAttrs(w); err != nil { return err }
//   header_footer.go:56-58  Serialize:      if err := h.serializeAttrs(w); err != nil { return err }
//   header_footer.go:64-66  Deserialize:    if err := h.BandBase.Deserialize(r); err != nil { return err }
//
// All three are architecturally dead: BandBase.serializeAttrs never returns
// non-nil (all writer methods are void), and BandBase.Deserialize never returns
// non-nil without breakableDeserialize injection. The breakableSerialize /
// breakableDeserialize hooks from band_error_hooks_test.go DO fire these paths
// for BandBase itself, but they do NOT reach HeaderFooterBandBase because the
// hook is at the BandBase level and BandBase.serializeAttrs calls the hook
// directly — causing BandBase.serializeAttrs to return err immediately, before
// HeaderFooterBandBase.serializeAttrs's own call to BandBase.serializeAttrs can
// return err. That is, injecting a breakableSerialize error causes:
//
//   BandBase.serializeAttrs → breakableSerialize → errInjected → returns errInjected
//
// so HeaderFooterBandBase.serializeAttrs's `if err := h.BandBase.serializeAttrs(w)`
// check DOES receive errInjected and DOES return it — covering line 44.
//
// These tests use the existing hooks from band_error_hooks_test.go.

import (
	"testing"
)

// TestHeaderFooterBandBase_serializeAttrs_BandBaseError covers the
// "return err" branch at header_footer.go:44 by injecting a
// breakableSerialize error so that h.BandBase.serializeAttrs(w) returns err.
func TestHeaderFooterBandBase_serializeAttrs_BandBaseSerializeError(t *testing.T) {
	h := NewHeaderFooterBandBase()
	w := newMockWriter()

	var got error
	withBreakableSerializeError(func() {
		got = h.serializeAttrs(w)
	})

	if got == nil {
		t.Fatal("serializeAttrs: expected error from BandBase.serializeAttrs via breakableSerialize hook, got nil")
	}
	if got != errInjected {
		t.Errorf("serializeAttrs: got %v, want errInjected", got)
	}
}

// TestHeaderFooterBandBase_Serialize_serializeAttrsError covers the
// "return err" branch at header_footer.go:57 by injecting a
// breakableSerialize error so that h.serializeAttrs(w) returns err.
func TestHeaderFooterBandBase_Serialize_SerializeAttrsError(t *testing.T) {
	h := NewHeaderFooterBandBase()
	w := newMockWriter()

	var got error
	withBreakableSerializeError(func() {
		got = h.Serialize(w)
	})

	if got == nil {
		t.Fatal("Serialize: expected error from serializeAttrs via breakableSerialize hook, got nil")
	}
	if got != errInjected {
		t.Errorf("Serialize: got %v, want errInjected", got)
	}
}

// TestHeaderFooterBandBase_Deserialize_BandBaseDeserializeError covers the
// "return err" branch at header_footer.go:65 by injecting a
// breakableDeserialize error so that h.BandBase.Deserialize(r) returns err.
func TestHeaderFooterBandBase_Deserialize_BandBaseDeserializeError(t *testing.T) {
	h := NewHeaderFooterBandBase()
	r := newMockReader()

	var got error
	withBreakableDeserializeError(func() {
		got = h.Deserialize(r)
	})

	if got == nil {
		t.Fatal("Deserialize: expected error from BandBase.Deserialize via breakableDeserialize hook, got nil")
	}
	if got != errInjected {
		t.Errorf("Deserialize: got %v, want errInjected", got)
	}
}

// TestGroupHeaderBand_Serialize_HeaderFooterBaseError covers the
// "return err" branch at types.go:239 (GroupHeaderBand.Serialize calls
// g.HeaderFooterBandBase.serializeAttrs which calls BandBase.serializeAttrs).
func TestGroupHeaderBand_Serialize_HeaderFooterBaseError(t *testing.T) {
	g := NewGroupHeaderBand()
	g.SetCondition("[Name]")
	w := newMockWriter()

	var got error
	withBreakableSerializeError(func() {
		got = g.Serialize(w)
	})

	if got == nil {
		t.Fatal("GroupHeaderBand.Serialize: expected error via breakableSerialize hook, got nil")
	}
	if got != errInjected {
		t.Errorf("GroupHeaderBand.Serialize: got %v, want errInjected", got)
	}
}

// TestGroupHeaderBand_Deserialize_HeaderFooterBaseError covers the
// "return err" branch at types.go:259 (GroupHeaderBand.Deserialize calls
// g.HeaderFooterBandBase.Deserialize which calls BandBase.Deserialize).
func TestGroupHeaderBand_Deserialize_HeaderFooterBaseError(t *testing.T) {
	g := NewGroupHeaderBand()
	r := newMockReader()

	var got error
	withBreakableDeserializeError(func() {
		got = g.Deserialize(r)
	})

	if got == nil {
		t.Fatal("GroupHeaderBand.Deserialize: expected error via breakableDeserialize hook, got nil")
	}
	if got != errInjected {
		t.Errorf("GroupHeaderBand.Deserialize: got %v, want errInjected", got)
	}
}

// TestDataBand_Serialize_BandBaseError covers the "return err" branch at
// types.go:427 (DataBand.Serialize calls d.BandBase.serializeAttrs).
func TestDataBand_Serialize_BandBaseError(t *testing.T) {
	d := NewDataBand()
	w := newMockWriter()

	var got error
	withBreakableSerializeError(func() {
		got = d.Serialize(w)
	})

	if got == nil {
		t.Fatal("DataBand.Serialize: expected error via breakableSerialize hook, got nil")
	}
	if got != errInjected {
		t.Errorf("DataBand.Serialize: got %v, want errInjected", got)
	}
}

// TestDataBand_Deserialize_BandBaseError covers the "return err" branch at
// types.go:479 (DataBand.Deserialize calls d.BandBase.Deserialize).
func TestDataBand_Deserialize_BandBaseError(t *testing.T) {
	d := NewDataBand()
	r := newMockReader()

	var got error
	withBreakableDeserializeError(func() {
		got = d.Deserialize(r)
	})

	if got == nil {
		t.Fatal("DataBand.Deserialize: expected error via breakableDeserialize hook, got nil")
	}
	if got != errInjected {
		t.Errorf("DataBand.Deserialize: got %v, want errInjected", got)
	}
}
