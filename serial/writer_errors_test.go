package serial

// writer_errors_test.go — internal tests for the remaining uncovered error branches
// in writer.go. Uses package serial (not serial_test) to access unexported types.

import (
	"encoding/xml"
	"errors"
	"strings"
	"testing"

	"github.com/andrewloable/go-fastreport/report"
)

// ── alwaysFailWriter: io.Writer that always fails ─────────────────────────────

type alwaysFailWriter struct {
	err error
}

func (f *alwaysFailWriter) Write(p []byte) (int, error) {
	return 0, f.err
}

// ── EndObject: empty stack → error ────────────────────────────────────────────

func TestEndObject_EmptyStack_ReturnsError(t *testing.T) {
	var buf strings.Builder
	w := NewWriter(&buf)
	err := w.EndObject() // no preceding BeginObject
	if err == nil {
		t.Fatal("EndObject on empty stack: expected error, got nil")
	}
	if !strings.Contains(err.Error(), "empty stack") {
		t.Errorf("error should mention empty stack: %v", err)
	}
}

// ── flushPending: EncodeToken error with large attribute ─────────────────────
//
// xml.Encoder buffers internally (bufio, default 4096 B). Writing a StartElement
// with an attribute value larger than the buffer forces an immediate flush of the
// bufio.Writer to the underlying io.Writer.  If that writer fails, EncodeToken
// returns the error, which flushPending propagates.

func TestFlushPending_EncodeTokenError(t *testing.T) {
	fw := &alwaysFailWriter{err: errors.New("forced write failure")}
	w := NewWriter(fw)

	// Manually push an un-flushed element with a large attribute so that
	// EncodeToken must write >4096 bytes and flush to the failing writer.
	w.stack = append(w.stack, elementState{
		name: "Elem",
		attrs: []xml.Attr{{
			Name:  xml.Name{Local: "Data"},
			Value: strings.Repeat("X", 8000),
		}},
	})

	err := w.flushPending()
	if err == nil {
		t.Fatal("flushPending: expected error from EncodeToken with failing writer, got nil")
	}
}

// ── BeginObject: flushPending error propagation ───────────────────────────────

func TestBeginObject_FlushPendingError(t *testing.T) {
	fw := &alwaysFailWriter{err: errors.New("forced write failure")}
	w := NewWriter(fw)

	// Push a large pending element; the next BeginObject will call flushPending
	// which must flush the >4096-byte start tag to the failing writer.
	w.stack = append(w.stack, elementState{
		name: "Parent",
		attrs: []xml.Attr{{
			Name:  xml.Name{Local: "A"},
			Value: strings.Repeat("Y", 8000),
		}},
	})

	err := w.BeginObject("Child")
	if err == nil {
		t.Fatal("BeginObject: expected error from flushPending, got nil")
	}
}

// ── WriteObject: BeginObject error propagation ────────────────────────────────

// simpleSerializable is a minimal Serializable for use in error-path tests.
type simpleSerializable struct{}

func (s *simpleSerializable) TypeName() string                  { return "Simple" }
func (s *simpleSerializable) Serialize(w report.Writer) error   { return nil }
func (s *simpleSerializable) Deserialize(r report.Reader) error { return nil }

func TestWriteObject_BeginObjectError(t *testing.T) {
	fw := &alwaysFailWriter{err: errors.New("forced write failure")}
	w := NewWriter(fw)

	// Push a large pending element so that WriteObject's BeginObject call
	// triggers a flushPending failure.
	w.stack = append(w.stack, elementState{
		name: "Parent",
		attrs: []xml.Attr{{
			Name:  xml.Name{Local: "A"},
			Value: strings.Repeat("Z", 8000),
		}},
	})

	err := w.WriteObject(&simpleSerializable{})
	if err == nil {
		t.Fatal("WriteObject: expected error from BeginObject, got nil")
	}
}

// ── WriteObjectNamed: BeginObject error propagation ──────────────────────────

func TestWriteObjectNamed_BeginObjectError(t *testing.T) {
	fw := &alwaysFailWriter{err: errors.New("forced write failure")}
	w := NewWriter(fw)

	w.stack = append(w.stack, elementState{
		name: "Parent",
		attrs: []xml.Attr{{
			Name:  xml.Name{Local: "A"},
			Value: strings.Repeat("W", 8000),
		}},
	})

	err := w.WriteObjectNamed("Child", &simpleSerializable{})
	if err == nil {
		t.Fatal("WriteObjectNamed: expected error from BeginObject, got nil")
	}
}
