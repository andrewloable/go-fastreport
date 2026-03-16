package serial

// writer_extra_coverage_test.go — additional internal tests targeting the
// remaining uncovered statements in writer.go:
//
//   writer.go:92  EndObject: `return err` when EncodeToken(StartElement) fails
//                 for an un-flushed element
//   writer.go:96  EndObject: `return err` when EncodeToken(EndElement) fails
//   writer.go:212 typeNameOf: `return name` when %T has no '.' (unreachable for
//                 any real Go named type — documented below)
//
// Tests use package serial (not serial_test) to access unexported fields.

import (
	"encoding/xml"
	"errors"
	"strings"
	"testing"
)

// ── EndObject: EncodeToken(StartElement) error path (line 92) ───────────────
//
// Strategy:
//   Push an un-flushed element (flushed=false) with a large attribute onto
//   w.stack, backed by an always-failing writer.  When EndObject calls
//   EncodeToken(StartElement{...}) the xml.Encoder must flush its internal
//   bufio.Writer to the underlying io.Writer — which fails — and EncodeToken
//   returns that error, hitting the `return err` at line 92.

func TestEndObject_StartElementTokenError(t *testing.T) {
	fw := &alwaysFailWriter{err: errors.New("forced write failure")}
	w := NewWriter(fw)

	// Un-flushed element with a large attribute to force an immediate flush
	// of the bufio.Writer when EncodeToken(StartElement) is called.
	w.stack = append(w.stack, elementState{
		name:    "BigElem",
		flushed: false, // EndObject will try to emit the StartElement
		attrs: []xml.Attr{{
			Name:  xml.Name{Local: "Data"},
			Value: strings.Repeat("X", 8000), // > 4096-byte bufio buffer
		}},
	})

	err := w.EndObject()
	if err == nil {
		t.Fatal("EndObject un-flushed with failing writer: expected StartElement error, got nil")
	}
}

// ── EndObject: EncodeToken(EndElement) error path (line 96) ──────────────────
//
// Strategy:
//   Push an already-flushed element onto w.stack (flushed=true) so that
//   EndObject skips the StartElement emission branch and goes directly to
//   EncodeToken(xml.EndElement{...}).  Because the xml.Encoder's own internal
//   element stack is empty (no matching StartElement was ever written to it),
//   the encoder detects the unmatched end tag and returns an error.
//   That error flows to the `return err` statement at line 96.

func TestEndObject_EndElementTokenError(t *testing.T) {
	var buf strings.Builder
	w := NewWriter(&buf)

	// Push a flushed element — the xml.Encoder does NOT know about it,
	// so its internal element stack is empty.
	w.stack = append(w.stack, elementState{
		name:    "Phantom",
		flushed: true, // skip the StartElement branch in EndObject
	})

	err := w.EndObject()
	if err == nil {
		t.Fatal("EndObject with flushed element on empty encoder stack: expected error from unmatched EndElement, got nil")
	}
	// The xml package returns "xml: end tag </Phantom> without start tag" or similar.
	if !strings.Contains(err.Error(), "Phantom") && !strings.Contains(err.Error(), "end tag") && !strings.Contains(err.Error(), "start") {
		t.Errorf("unexpected error message: %v", err)
	}
}

// ── typeNameOf: no-dot fallback (line 212) — documented as unreachable ────────
//
// The `return name` branch at line 212 executes when fmt.Sprintf("%T", obj)
// produces a string with no '.' separator, meaning no package prefix.  In Go,
// every named type that can implement an interface (and thus report.Serializable)
// always has a package prefix in its %T string (e.g. "serial.noDotObj").
//
// Unnamed / built-in types (int, string, …) cannot satisfy report.Serializable
// because you cannot define methods on them.
//
// This branch is structurally unreachable for any valid Serializable value and
// is therefore not tested.  The statement counter in go tool cover will continue
// to show it as uncovered; this comment documents the constraint.
