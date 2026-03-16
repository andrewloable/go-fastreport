package serial

// writer_coverage_test.go — internal tests for uncovered branches in writer.go.
// Uses package serial (not serial_test) to access unexported helpers.

import (
	"bytes"
	"fmt"
	"strings"
	"testing"

	"github.com/andrewloable/go-fastreport/report"
)

// ── typeNameOf: strip package prefix ─────────────────────────────────────────

// noDotObj is a type in this package so %T returns "*serial.noDotObj".
// typeNameOf will find the dot and return "noDotObj".
type noDotObj struct{}

func (n *noDotObj) Serialize(w report.Writer) error   { return nil }
func (n *noDotObj) Deserialize(r report.Reader) error { return nil }

func TestTypeNameOf_StripsDotPrefix(t *testing.T) {
	obj := &noDotObj{}
	got := typeNameOf(obj)
	if got != "noDotObj" {
		t.Errorf("typeNameOf: got %q, want noDotObj", got)
	}
}

func TestTypeNameOf_WithTyperNameInterface(t *testing.T) {
	// Objects implementing TypeNamer should use that.
	obj := &namedObj{}
	got := typeNameOf(obj)
	if got != "CustomName" {
		t.Errorf("typeNameOf: got %q, want CustomName", got)
	}
}

type namedObj struct{}

func (n *namedObj) TypeName() string                  { return "CustomName" }
func (n *namedObj) Serialize(w report.Writer) error   { return nil }
func (n *namedObj) Deserialize(r report.Reader) error { return nil }

// ── flushPending: already-flushed is a no-op ─────────────────────────────────

func TestFlushPending_AlreadyFlushedNoOp(t *testing.T) {
	var buf bytes.Buffer
	w := NewWriter(&buf)
	// Push an already-flushed element directly.
	w.stack = append(w.stack, elementState{name: "Elem", flushed: true})
	err := w.flushPending()
	if err != nil {
		t.Fatalf("flushPending on flushed element: expected nil error, got %v", err)
	}
	// Stack should still have the element (no mutation by flushPending).
	if len(w.stack) != 1 {
		t.Errorf("stack len: got %d, want 1", len(w.stack))
	}
}

func TestFlushPending_EmptyStackNoOp(t *testing.T) {
	var buf bytes.Buffer
	w := NewWriter(&buf)
	err := w.flushPending()
	if err != nil {
		t.Fatalf("flushPending on empty stack: expected nil, got %v", err)
	}
}

// ── EndObject: un-flushed element (no children) emits self-closing tag ────────

func TestEndObject_UnflushedElement(t *testing.T) {
	var buf bytes.Buffer
	w := NewWriter(&buf)
	// Manually push an un-flushed element.
	w.stack = append(w.stack, elementState{name: "SelfClose", flushed: false})
	err := w.EndObject()
	if err != nil {
		t.Fatalf("EndObject: %v", err)
	}
	w.Flush() //nolint:errcheck
	out := buf.String()
	if !strings.Contains(out, "SelfClose") {
		t.Errorf("expected SelfClose in output:\n%s", out)
	}
}

// ── EndObject: already-flushed element skips StartElement emit ────────────────

func TestEndObject_FlushedElementSkipsStart(t *testing.T) {
	var buf bytes.Buffer
	w := NewWriter(&buf)
	if err := w.BeginObject("Parent"); err != nil {
		t.Fatal(err)
	}
	// Force flush of Parent by opening a child.
	if err := w.BeginObject("Child"); err != nil {
		t.Fatal(err)
	}
	if err := w.EndObject(); err != nil { // close Child
		t.Fatal(err)
	}
	// Parent.flushed == true now — EndObject takes the already-flushed path.
	if err := w.EndObject(); err != nil {
		t.Fatalf("EndObject flushed parent: %v", err)
	}
	w.Flush() //nolint:errcheck
}

// ── BeginObject: flushPending is called when stack is non-empty ───────────────

func TestBeginObject_FlushesParentBeforeChild(t *testing.T) {
	var buf bytes.Buffer
	w := NewWriter(&buf)

	if err := w.BeginObject("Parent"); err != nil {
		t.Fatal(err)
	}
	w.WriteStr("PA", "pval")

	// BeginObject("Child") must call flushPending → Parent start tag emitted.
	if err := w.BeginObject("Child"); err != nil {
		t.Fatal(err)
	}
	if len(w.stack) != 2 {
		t.Errorf("stack len: got %d, want 2", len(w.stack))
	}
	// Parent should now be marked flushed.
	if !w.stack[0].flushed {
		t.Error("parent element should be marked flushed after child BeginObject")
	}
	w.EndObject() //nolint:errcheck
	w.EndObject() //nolint:errcheck
	w.Flush()     //nolint:errcheck
}

// ── WriteObject: Serialize error propagation ──────────────────────────────────

type serErrObj2 struct{}

func (s *serErrObj2) TypeName() string                  { return "SerErrObj2" }
func (s *serErrObj2) Serialize(w report.Writer) error   { return fmt.Errorf("serialize failed") }
func (s *serErrObj2) Deserialize(r report.Reader) error { return nil }

func TestWriteObject_SerializeError(t *testing.T) {
	var buf bytes.Buffer
	w := NewWriter(&buf)
	err := w.WriteObject(&serErrObj2{})
	if err == nil || !strings.Contains(err.Error(), "serialize failed") {
		t.Errorf("WriteObject: expected serialize error, got %v", err)
	}
}

// ── WriteObjectNamed: Serialize error propagation ─────────────────────────────

func TestWriteObjectNamed_SerializeError2(t *testing.T) {
	var buf bytes.Buffer
	w := NewWriter(&buf)
	err := w.WriteObjectNamed("SerErrObj2", &serErrObj2{})
	if err == nil || !strings.Contains(err.Error(), "serialize failed") {
		t.Errorf("WriteObjectNamed: expected serialize error, got %v", err)
	}
}

// ── typeNameOf: no-dot fallback (name has no '.' so whole name is returned) ───

// This type has no dot in its %T representation ... which never happens in Go
// for named types. Verify the loop terminates correctly for the current case.
func TestTypeNameOf_LoopBehavior(t *testing.T) {
	obj := &noDotObj{}
	name := fmt.Sprintf("%T", obj)
	got := typeNameOf(obj)
	// typeNameOf should strip the package prefix
	if strings.Contains(got, ".") || strings.Contains(got, "*") {
		t.Errorf("typeNameOf(%q): unexpected chars in %q", name, got)
	}
}

// ── addAttr: no-op on empty stack ────────────────────────────────────────────

func TestAddAttr_EmptyStack(t *testing.T) {
	var buf bytes.Buffer
	w := NewWriter(&buf)
	// addAttr with empty stack should be a no-op, no panic.
	w.addAttr("Key", "Val")
	if err := w.Flush(); err != nil {
		t.Fatalf("Flush: %v", err)
	}
	if buf.Len() != 0 {
		t.Errorf("expected empty output, got %q", buf.String())
	}
}

// ── BeginObject: stack grows correctly ───────────────────────────────────────

func TestBeginObject_StackGrowth(t *testing.T) {
	var buf bytes.Buffer
	w := NewWriter(&buf)

	// Open 3 nested levels.
	for _, name := range []string{"Level1", "Level2", "Level3"} {
		if err := w.BeginObject(name); err != nil {
			t.Fatalf("BeginObject(%q): %v", name, err)
		}
	}
	if len(w.stack) != 3 {
		t.Errorf("stack len: got %d, want 3", len(w.stack))
	}
	// Close them all.
	for range 3 {
		if err := w.EndObject(); err != nil {
			t.Fatalf("EndObject: %v", err)
		}
	}
	if err := w.Flush(); err != nil {
		t.Fatal(err)
	}
}
