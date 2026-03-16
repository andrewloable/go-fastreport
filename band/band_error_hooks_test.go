package band

// band_error_hooks_test.go – internal tests that use the breakableSerialize and
// breakableDeserialize testability hooks to cover the dead error-propagation
// branches in BandBase.serializeAttrs, BandBase.Serialize, BandBase.Deserialize,
// and ChildBand.Deserialize.
//
// These branches are of the form:
//
//	if err := breakable{Serialize,Deserialize}(...); err != nil { return err }
//
// Under normal production use the wrapped functions never return non-nil errors,
// making those branches unreachable without injection.

import (
	"errors"
	"testing"

	"github.com/andrewloable/go-fastreport/report"
)

// ── helpers ───────────────────────────────────────────────────────────────────

var errInjected = errors.New("injected test error")

// withBreakableSerializeError temporarily replaces breakableSerialize so that
// it returns errInjected, runs f, then restores the original.
func withBreakableSerializeError(f func()) {
	orig := breakableSerialize
	breakableSerialize = func(bc *report.BreakableComponent, w report.Writer) error {
		return errInjected
	}
	defer func() { breakableSerialize = orig }()
	f()
}

// withBreakableDeserializeError temporarily replaces breakableDeserialize so that
// it returns errInjected, runs f, then restores the original.
func withBreakableDeserializeError(f func()) {
	orig := breakableDeserialize
	breakableDeserialize = func(bc *report.BreakableComponent, r report.Reader) error {
		return errInjected
	}
	defer func() { breakableDeserialize = orig }()
	f()
}

// ── BandBase.serializeAttrs: return err branch (line ~305) ───────────────────

func TestBandBase_serializeAttrs_BreakableSerializeError(t *testing.T) {
	b := NewBandBase()
	w := newMockWriter()

	var got error
	withBreakableSerializeError(func() {
		got = b.serializeAttrs(w)
	})

	if got == nil {
		t.Fatal("serializeAttrs: expected error from breakableSerialize hook, got nil")
	}
	if got != errInjected {
		t.Errorf("serializeAttrs: got %v, want errInjected", got)
	}
}

// ── BandBase.Serialize: return err branch from serializeAttrs (line ~349) ────

func TestBandBase_Serialize_SerializeAttrsError(t *testing.T) {
	b := NewBandBase()
	w := newMockWriter()

	var got error
	withBreakableSerializeError(func() {
		got = b.Serialize(w)
	})

	if got == nil {
		t.Fatal("Serialize: expected error from serializeAttrs, got nil")
	}
	if got != errInjected {
		t.Errorf("Serialize: got %v, want errInjected", got)
	}
}

// ── BandBase.Deserialize: return err branch (line ~357) ──────────────────────

func TestBandBase_Deserialize_BreakableDeserializeError(t *testing.T) {
	b := NewBandBase()
	r := newMockReader()

	var got error
	withBreakableDeserializeError(func() {
		got = b.Deserialize(r)
	})

	if got == nil {
		t.Fatal("BandBase.Deserialize: expected error from breakableDeserialize hook, got nil")
	}
	if got != errInjected {
		t.Errorf("BandBase.Deserialize: got %v, want errInjected", got)
	}
}

// ── ChildBand.Deserialize: return err branch (line ~417) ─────────────────────

func TestChildBand_Deserialize_BandBaseDeserializeError(t *testing.T) {
	c := NewChildBand()
	r := newMockReader()

	// ChildBand.Deserialize calls c.BandBase.Deserialize(r) which in turn calls
	// breakableDeserialize. Injecting an error at that hook causes
	// BandBase.Deserialize to return it, and ChildBand.Deserialize must propagate it.
	var got error
	withBreakableDeserializeError(func() {
		got = c.Deserialize(r)
	})

	if got == nil {
		t.Fatal("ChildBand.Deserialize: expected propagated error from BandBase.Deserialize, got nil")
	}
	if got != errInjected {
		t.Errorf("ChildBand.Deserialize: got %v, want errInjected", got)
	}
}
