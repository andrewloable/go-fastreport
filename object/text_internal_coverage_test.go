package object

// text_internal_coverage_test.go — internal tests for text.go error-path branches
// that cannot be reached via the public serial.Reader/serial.Writer.
//
// Uncovered lines targeted:
//   text.go:169-171   TextObjectBase.Serialize: return err from BreakableComponent.Serialize
//   text.go:210-212   TextObjectBase.Deserialize: return err from BreakableComponent.Deserialize
//   text.go:575-576   DeserializeChild/Formats: if r.FinishChild()!=nil { break } (inner drain)
//   text.go:579-580   DeserializeChild/Formats: if r.FinishChild()!=nil { break } (outer)
//   text.go:617       DeserializeChild/Highlight: if r.FinishChild()!=nil { break }
//   text.go:626-628   TextObject.Serialize: return err from TextObjectBase.Serialize
//   text.go:705-707   TextObject.Deserialize: return err from TextObjectBase.Deserialize

import (
	"errors"
	"testing"

	"github.com/andrewloable/go-fastreport/report"
)

// ── mock Writer that errors on all base calls ─────────────────────────────────

// errSerializeWriter returns an error from WriteObject and WriteObjectNamed
// but is a no-op for all attribute writes. This lets us simulate the chain
// where BreakableComponent.Serialize → ReportComponentBase.Serialize calls
// eventually propagate an error. However, since the chain is pure in-memory
// no-ops, we need a writer whose Serialize chain returns an error.
//
// To trigger the error paths in TextObjectBase.Serialize and TextObject.Serialize,
// we need to make BreakableComponent.Serialize itself return a non-nil error.
// Since BreakableComponent.Serialize calls ReportComponentBase.Serialize, and
// that returns nil unconditionally, we use a wrapper: embed a real TextObject
// but swap out the BreakableComponent to one that short-circuits. The only
// practical way is to call Serialize with a custom writer that also implements
// the error return — but Serialize(w Writer) does not use the writer return value
// for attribute writes.
//
// Actually: the error comes from t.BreakableComponent.Serialize(w) which calls
// bc.ReportComponentBase.Serialize(w) which calls bc.ComponentBase.Serialize(w)
// which calls bc.BaseObject.Serialize(w). All of these return nil. The only way
// to get a non-nil error is to force the top-level writer to error.
//
// The ONLY way to cover the "return err" dead-code is with an internal test that
// directly calls the Serialize method on a subtype with a writer that errors.
// Since the error propagates from the embedded call, we need to override it.
//
// Strategy: create a minimal internal mock that wraps the call tree and injects
// an error. We use a testable sub-struct that overrides Serialize so that
// the parent call returns an error.

// errWriter is a writer that is no-op for attributes but errors on WriteObject.
type errOnSerializeWriter struct{ sentinel error }

func (w *errOnSerializeWriter) WriteStr(name, value string)            {}
func (w *errOnSerializeWriter) WriteInt(name string, value int)        {}
func (w *errOnSerializeWriter) WriteBool(name string, value bool)      {}
func (w *errOnSerializeWriter) WriteFloat(name string, value float32)  {}
func (w *errOnSerializeWriter) WriteObject(obj report.Serializable) error {
	return w.sentinel
}
func (w *errOnSerializeWriter) WriteObjectNamed(name string, obj report.Serializable) error {
	return w.sentinel
}

// ── errOnDeserializeReader: mock Reader that errors on FinishChild ────────────

// errFinishReader is a mock Reader whose FinishChild always returns an error.
// NextChild returns a synthetic child on the first call, then returns false.
type errFinishReader struct {
	callCount    int
	childName    string
	returnChildN int // how many times to return a valid child before returning false
}

func (r *errFinishReader) ReadStr(name, def string) string  { return def }
func (r *errFinishReader) ReadInt(name string, def int) int { return def }
func (r *errFinishReader) ReadBool(name string, def bool) bool { return def }
func (r *errFinishReader) ReadFloat(name string, def float32) float32 { return def }
func (r *errFinishReader) NextChild() (string, bool) {
	if r.callCount < r.returnChildN {
		r.callCount++
		return r.childName, true
	}
	return "", false
}
func (r *errFinishReader) FinishChild() error {
	return errors.New("mock FinishChild error")
}

// ── errDeserializeReader: mock Reader whose Deserialize will propagate an error
// The trick: we can't make BreakableComponent.Deserialize return an error
// through the real chain. Instead we need a mock reader that doesn't work.
// But actually the error paths in TextObjectBase.Deserialize:210 and
// TextObject.Deserialize:705 are dead code (BreakableComponent.Deserialize
// always returns nil). We still want to exercise them for coverage.
//
// We can achieve this by creating a custom sub-type in the internal test
// that overrides the embedded Deserialize. But Go's composition doesn't allow
// that easily. Instead we directly call the Deserialize method with a
// custom reader. The error path requires BreakableComponent.Deserialize
// to return non-nil, which it won't with any standard reader.
//
// These paths are truly dead code under the current design. We accept that
// they remain at 0 for the error-return statements themselves and focus on
// the reachable branches that CAN be covered.

// ── TestTextObject_DeserializeChild_Formats_FinishChildError ─────────────────
// Cover text.go:579-580: the outer `if r.FinishChild() != nil { break }` in
// the Formats loop. After NextChild returns a child and we process the format,
// FinishChild errors → we break out of the loop.

func TestTextObject_DeserializeChild_Formats_OuterFinishChildError(t *testing.T) {
	obj := NewTextObject()
	// Reader returns one child named "NumberFormat", then FinishChild errors.
	r := &errFinishReader{
		childName:    "NumberFormat",
		returnChildN: 1,
	}
	// Call with "Formats" child type — exercises the outer FinishChild error break.
	handled := obj.DeserializeChild("Formats", r)
	if !handled {
		t.Error("expected DeserializeChild to return true for 'Formats'")
	}
	// After the break the formats collection was created (just the outer loop broke).
	if obj.formats == nil {
		t.Error("expected formats collection to be initialized")
	}
}

// ── TestTextObject_DeserializeChild_Highlight_FinishChildError ────────────────
// Cover text.go:617: `if r.FinishChild() != nil { break }` in Highlight loop.
// After reading a Condition child, FinishChild errors → we break.

func TestTextObject_DeserializeChild_Highlight_FinishChildError(t *testing.T) {
	obj := NewTextObject()
	// Reader returns one child named "Condition", then FinishChild errors.
	r := &errFinishReader{
		childName:    "Condition",
		returnChildN: 1,
	}
	handled := obj.DeserializeChild("Highlight", r)
	if !handled {
		t.Error("expected DeserializeChild to return true for 'Highlight'")
	}
	// A Condition was appended before FinishChild was called.
	if len(obj.highlights) != 1 {
		t.Errorf("expected 1 highlight, got %d", len(obj.highlights))
	}
}

// ── TestTextObject_DeserializeChild_Formats_InnerDrainFinishChildError ────────
// Cover text.go:575-576: the inner drain loop's `if r.FinishChild()!=nil{break}`.
// This is the loop that drains sub-children of each format element.
// To reach it, NextChild for the format element must return a sub-child,
// and then FinishChild for that sub-child must error.
//
// We need a reader that:
//   1st NextChild call (for the <Formats> level) → "NumberFormat"
//   2nd NextChild call (for the inner drain of NumberFormat's children) → "SubChild"
//   FinishChild → error (covers break at line 576)
//
// We implement a sequenced reader.

type sequencedReader struct {
	childSeq    []string // sequence of child type names to return from NextChild
	childIdx    int
	finishError bool // if true FinishChild always returns error
}

func (r *sequencedReader) ReadStr(name, def string) string  { return def }
func (r *sequencedReader) ReadInt(name string, def int) int { return def }
func (r *sequencedReader) ReadBool(name string, def bool) bool { return def }
func (r *sequencedReader) ReadFloat(name string, def float32) float32 { return def }
func (r *sequencedReader) NextChild() (string, bool) {
	if r.childIdx < len(r.childSeq) {
		name := r.childSeq[r.childIdx]
		r.childIdx++
		return name, true
	}
	return "", false
}
func (r *sequencedReader) FinishChild() error {
	if r.finishError {
		return errors.New("mock FinishChild error")
	}
	return nil
}

func TestTextObject_DeserializeChild_Formats_InnerDrainFinishChildError(t *testing.T) {
	obj := NewTextObject()
	// Sequence:
	//   1st NextChild → "NumberFormat"  (outer loop gets this, f != nil)
	//   2nd NextChild → "SubChild"       (inner drain loop finds a sub-child)
	//   3rd NextChild → "" (inner drain loop ends because FinishChild errors first)
	// FinishChild always errors → inner drain breaks at line 576.
	r := &sequencedReader{
		childSeq:    []string{"NumberFormat", "SubChild"},
		finishError: true,
	}
	handled := obj.DeserializeChild("Formats", r)
	if !handled {
		t.Error("expected DeserializeChild to return true for 'Formats'")
	}
	// formats was initialized; first format was added then inner drain errored
	if obj.formats == nil {
		t.Error("expected formats to be initialized")
	}
}

// ── TestTextObject_DeserializeChild_Formats_NilFormat ─────────────────────────
// Cover the branch where deserializeFormatFromChild returns nil (unknown type).
// In that case: f == nil → skip the Add/sync block → then drain and FinishChild.

func TestTextObject_DeserializeChild_Formats_NilFormatType(t *testing.T) {
	obj := NewTextObject()
	// "UnknownFormatType" maps to nil in deserializeFormatFromChild.
	r := &sequencedReader{
		childSeq:    []string{"UnknownFormatType"},
		finishError: false,
	}
	handled := obj.DeserializeChild("Formats", r)
	if !handled {
		t.Error("expected DeserializeChild to return true for 'Formats'")
	}
	if obj.formats == nil {
		t.Error("expected formats collection to be initialized even with nil format")
	}
	// format field should remain nil since no valid format was added.
	if obj.format != nil {
		t.Error("expected format field to remain nil when all formats are unknown type")
	}
}

// ── TestTextObject_DeserializeChild_Highlight_NonConditionChild ───────────────
// Cover the branch where condType != "Condition" in the Highlight loop.
// This exercises the else-path of `if condType == "Condition"` at line 594.

func TestTextObject_DeserializeChild_Highlight_NonConditionChild(t *testing.T) {
	obj := NewTextObject()
	// Return a non-Condition child type.
	r := &sequencedReader{
		childSeq:    []string{"SomeOtherElement"},
		finishError: false,
	}
	handled := obj.DeserializeChild("Highlight", r)
	if !handled {
		t.Error("expected DeserializeChild to return true for 'Highlight'")
	}
	// No Condition was added.
	if len(obj.highlights) != 0 {
		t.Errorf("expected 0 highlights for non-Condition child, got %d", len(obj.highlights))
	}
}
