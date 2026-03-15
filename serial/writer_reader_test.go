package serial

import (
	"bytes"
	"fmt"
	"io"
	"strings"
	"testing"

	"github.com/andrewloable/go-fastreport/report"
)

// ── test object definitions ───────────────────────────────────────────────────

// textObject is a minimal Serializable that round-trips a rich set of
// property types (string, int, bool, float32) and a nested child.
type textObject struct {
	Name    string
	Left    int
	Top     int
	Width   float32
	Height  float32
	Visible bool
	Text    string
	Child   *childObject // optional nested child
}

func (t *textObject) TypeName() string { return "TextObject" }

func (t *textObject) Serialize(w report.Writer) error {
	w.WriteStr("Name", t.Name)
	w.WriteInt("Left", t.Left)
	w.WriteInt("Top", t.Top)
	w.WriteFloat("Width", t.Width)
	w.WriteFloat("Height", t.Height)
	w.WriteBool("Visible", t.Visible)
	w.WriteStr("Text", t.Text)
	if t.Child != nil {
		if err := w.WriteObject(t.Child); err != nil {
			return err
		}
	}
	return nil
}

func (t *textObject) Deserialize(r report.Reader) error {
	t.Name = r.ReadStr("Name", "")
	t.Left = r.ReadInt("Left", 0)
	t.Top = r.ReadInt("Top", 0)
	t.Width = r.ReadFloat("Width", 0)
	t.Height = r.ReadFloat("Height", 0)
	t.Visible = r.ReadBool("Visible", false)
	t.Text = r.ReadStr("Text", "")

	typeName, ok := r.NextChild()
	if ok && typeName == "ChildObject" {
		t.Child = &childObject{}
		if err := t.Child.Deserialize(r); err != nil {
			return err
		}
		if err := r.(*Reader).FinishChild(); err != nil {
			return err
		}
	} else if ok {
		// Unknown child — skip it.
		if err := r.(*Reader).SkipElement(); err != nil {
			return err
		}
		if err := r.(*Reader).FinishChild(); err != nil {
			return err
		}
	}
	return nil
}

type childObject struct {
	Label string
	Value int
}

func (c *childObject) TypeName() string { return "ChildObject" }

func (c *childObject) Serialize(w report.Writer) error {
	w.WriteStr("Label", c.Label)
	w.WriteInt("Value", c.Value)
	return nil
}

func (c *childObject) Deserialize(r report.Reader) error {
	c.Label = r.ReadStr("Label", "")
	c.Value = r.ReadInt("Value", 0)
	return nil
}

// ── Writer tests ──────────────────────────────────────────────────────────────

func TestWriterImplementsInterface(t *testing.T) {
	var _ report.Writer = (*Writer)(nil)
}

func TestReaderImplementsInterface(t *testing.T) {
	var _ report.Reader = (*Reader)(nil)
}

func TestWriteHeader(t *testing.T) {
	var buf bytes.Buffer
	w := NewWriter(&buf)
	if err := w.WriteHeader(); err != nil {
		t.Fatalf("WriteHeader: %v", err)
	}
	got := buf.String()
	if !strings.HasPrefix(got, `<?xml version="1.0" encoding="utf-8"?>`) {
		t.Errorf("unexpected header: %q", got)
	}
}

func TestWriteSimpleObject(t *testing.T) {
	obj := &textObject{
		Name:    "Text1",
		Left:    10,
		Top:     20,
		Width:   200.5,
		Height:  30.0,
		Visible: true,
		Text:    "Hello, World!",
	}

	var buf bytes.Buffer
	w := NewWriter(&buf)
	if err := w.WriteObjectNamed("TextObject", obj); err != nil {
		t.Fatalf("WriteObjectNamed: %v", err)
	}
	if err := w.Flush(); err != nil {
		t.Fatalf("Flush: %v", err)
	}

	out := buf.String()
	for _, want := range []string{
		`TextObject`,
		`Name="Text1"`,
		`Left="10"`,
		`Top="20"`,
		`Width="200.5"`,
		`Height="30"`,
		`Visible="true"`,
		`Text="Hello, World!"`,
	} {
		if !strings.Contains(out, want) {
			t.Errorf("output missing %q\nfull output:\n%s", want, out)
		}
	}
}

func TestWriteObjectWithChild(t *testing.T) {
	obj := &textObject{
		Name:    "Parent",
		Text:    "outer",
		Child:   &childObject{Label: "lbl", Value: 42},
	}

	var buf bytes.Buffer
	w := NewWriter(&buf)
	if err := w.WriteObjectNamed("TextObject", obj); err != nil {
		t.Fatalf("WriteObjectNamed: %v", err)
	}
	if err := w.Flush(); err != nil {
		t.Fatalf("Flush: %v", err)
	}

	out := buf.String()
	if !strings.Contains(out, `ChildObject`) {
		t.Errorf("expected ChildObject element in:\n%s", out)
	}
	if !strings.Contains(out, `Label="lbl"`) {
		t.Errorf("expected Label attr in:\n%s", out)
	}
	if !strings.Contains(out, `Value="42"`) {
		t.Errorf("expected Value attr in:\n%s", out)
	}
}

func TestWriteBoolFalse(t *testing.T) {
	var buf bytes.Buffer
	w := NewWriter(&buf)
	if err := w.BeginObject("Obj"); err != nil {
		t.Fatal(err)
	}
	w.WriteBool("Active", false)
	if err := w.EndObject(); err != nil {
		t.Fatal(err)
	}
	if err := w.Flush(); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(buf.String(), `Active="false"`) {
		t.Errorf("unexpected output: %s", buf.String())
	}
}

func TestWriteFloatNegative(t *testing.T) {
	var buf bytes.Buffer
	w := NewWriter(&buf)
	if err := w.BeginObject("Obj"); err != nil {
		t.Fatal(err)
	}
	w.WriteFloat("X", -3.14)
	if err := w.EndObject(); err != nil {
		t.Fatal(err)
	}
	if err := w.Flush(); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(buf.String(), `X="-3.14"`) {
		t.Errorf("unexpected output: %s", buf.String())
	}
}

func TestEndObjectEmptyStack(t *testing.T) {
	var buf bytes.Buffer
	w := NewWriter(&buf)
	if err := w.EndObject(); err == nil {
		t.Error("expected error for EndObject on empty stack")
	}
}

// ── Reader tests ──────────────────────────────────────────────────────────────

func TestReadObjectHeader(t *testing.T) {
	src := `<Report Name="r1"></Report>`
	r := NewReader(strings.NewReader(src))

	typeName, ok := r.ReadObjectHeader()
	if !ok {
		t.Fatal("ReadObjectHeader returned ok=false")
	}
	if typeName != "Report" {
		t.Errorf("got typeName=%q, want Report", typeName)
	}
	if got := r.ReadStr("Name", ""); got != "r1" {
		t.Errorf("got Name=%q, want r1", got)
	}
}

func TestReadObjectHeaderEOF(t *testing.T) {
	r := NewReader(strings.NewReader(""))
	_, ok := r.ReadObjectHeader()
	if ok {
		t.Error("expected ok=false at EOF")
	}
}

func TestReadObjectHeaderEndElement(t *testing.T) {
	// A stray end element should return ok=false.
	src := `</Foo>`
	r := NewReader(strings.NewReader(src))
	_, ok := r.ReadObjectHeader()
	if ok {
		t.Error("expected ok=false on stray end element")
	}
}

func TestReadStrDefault(t *testing.T) {
	src := `<Obj/>`
	r := NewReader(strings.NewReader(src))
	r.ReadObjectHeader()
	if got := r.ReadStr("Missing", "default"); got != "default" {
		t.Errorf("got %q, want default", got)
	}
}

func TestReadIntDefault(t *testing.T) {
	src := `<Obj/>`
	r := NewReader(strings.NewReader(src))
	r.ReadObjectHeader()
	if got := r.ReadInt("Missing", 99); got != 99 {
		t.Errorf("got %d, want 99", got)
	}
}

func TestReadIntBadValue(t *testing.T) {
	src := `<Obj Count="abc"/>`
	r := NewReader(strings.NewReader(src))
	r.ReadObjectHeader()
	if got := r.ReadInt("Count", 7); got != 7 {
		t.Errorf("got %d, want 7", got)
	}
}

func TestReadBoolDefault(t *testing.T) {
	src := `<Obj/>`
	r := NewReader(strings.NewReader(src))
	r.ReadObjectHeader()
	if got := r.ReadBool("Missing", true); got != true {
		t.Errorf("got %v, want true", got)
	}
}

func TestReadBoolVariants(t *testing.T) {
	tests := []struct {
		val  string
		want bool
	}{
		{"true", true},
		{"True", true},
		{"TRUE", true},
		{"1", true},
		{"false", false},
		{"0", false},
		{"yes", false},
	}
	for _, tt := range tests {
		src := `<Obj F="` + tt.val + `"/>`
		r := NewReader(strings.NewReader(src))
		r.ReadObjectHeader()
		got := r.ReadBool("F", !tt.want)
		if got != tt.want {
			t.Errorf("val=%q: got %v, want %v", tt.val, got, tt.want)
		}
	}
}

func TestReadFloatDefault(t *testing.T) {
	src := `<Obj/>`
	r := NewReader(strings.NewReader(src))
	r.ReadObjectHeader()
	if got := r.ReadFloat("Missing", 1.5); got != 1.5 {
		t.Errorf("got %v, want 1.5", got)
	}
}

func TestReadFloatBadValue(t *testing.T) {
	src := `<Obj X="notafloat"/>`
	r := NewReader(strings.NewReader(src))
	r.ReadObjectHeader()
	if got := r.ReadFloat("X", 9.9); got != float32(9.9) {
		t.Errorf("got %v, want 9.9", got)
	}
}

func TestNextChildNoChildren(t *testing.T) {
	src := `<Obj/>`
	r := NewReader(strings.NewReader(src))
	r.ReadObjectHeader()
	_, ok := r.NextChild()
	if ok {
		t.Error("expected no children for self-closing element")
	}
}

func TestNextChildMultiple(t *testing.T) {
	src := `<Parent><Child1/><Child2/></Parent>`
	r := NewReader(strings.NewReader(src))
	r.ReadObjectHeader() // Parent

	names := []string{}
	for {
		name, ok := r.NextChild()
		if !ok {
			break
		}
		names = append(names, name)
		if err := r.FinishChild(); err != nil {
			t.Fatalf("FinishChild: %v", err)
		}
	}
	if len(names) != 2 || names[0] != "Child1" || names[1] != "Child2" {
		t.Errorf("got children %v, want [Child1 Child2]", names)
	}
}

func TestSkipElement(t *testing.T) {
	src := `<Parent><Skip><Deep/></Skip><Keep Name="k"/></Parent>`
	r := NewReader(strings.NewReader(src))
	r.ReadObjectHeader() // Parent

	name, ok := r.NextChild()
	if !ok || name != "Skip" {
		t.Fatalf("expected Skip child, got %q ok=%v", name, ok)
	}
	if err := r.SkipElement(); err != nil {
		t.Fatalf("SkipElement: %v", err)
	}
	if err := r.FinishChild(); err != nil {
		t.Fatalf("FinishChild after skip: %v", err)
	}

	name2, ok2 := r.NextChild()
	if !ok2 || name2 != "Keep" {
		t.Fatalf("expected Keep child, got %q ok=%v", name2, ok2)
	}
	if got := r.ReadStr("Name", ""); got != "k" {
		t.Errorf("got Name=%q, want k", got)
	}
	if err := r.FinishChild(); err != nil {
		t.Fatalf("FinishChild: %v", err)
	}
}

func TestFinishChildNoMatch(t *testing.T) {
	r := NewReader(strings.NewReader(""))
	if err := r.FinishChild(); err == nil {
		t.Error("expected error for FinishChild without NextChild")
	}
}

func TestCurrentName(t *testing.T) {
	src := `<MyElement/>`
	r := NewReader(strings.NewReader(src))
	r.ReadObjectHeader()
	if got := r.CurrentName(); got != "MyElement" {
		t.Errorf("got %q, want MyElement", got)
	}
}

// ── round-trip tests ──────────────────────────────────────────────────────────

// roundTrip serializes orig, then deserializes it into a fresh object and
// returns it.
func roundTrip(t *testing.T, orig *textObject) *textObject {
	t.Helper()

	// --- Serialize ---
	var buf bytes.Buffer
	w := NewWriter(&buf)
	if err := w.WriteObjectNamed("TextObject", orig); err != nil {
		t.Fatalf("serialize: %v", err)
	}
	if err := w.Flush(); err != nil {
		t.Fatalf("flush: %v", err)
	}

	// --- Deserialize ---
	r := NewReader(bytes.NewReader(buf.Bytes()))
	typeName, ok := r.ReadObjectHeader()
	if !ok {
		t.Fatalf("ReadObjectHeader returned ok=false; xml was:\n%s", buf.String())
	}
	if typeName != "TextObject" {
		t.Fatalf("got typeName=%q, want TextObject", typeName)
	}

	got := &textObject{}
	if err := got.Deserialize(r); err != nil {
		t.Fatalf("deserialize: %v", err)
	}
	return got
}

func TestRoundTripSimple(t *testing.T) {
	orig := &textObject{
		Name:    "Text1",
		Left:    10,
		Top:     20,
		Width:   200.5,
		Height:  30,
		Visible: true,
		Text:    "Hello, World!",
	}
	got := roundTrip(t, orig)

	if got.Name != orig.Name {
		t.Errorf("Name: got %q, want %q", got.Name, orig.Name)
	}
	if got.Left != orig.Left {
		t.Errorf("Left: got %d, want %d", got.Left, orig.Left)
	}
	if got.Top != orig.Top {
		t.Errorf("Top: got %d, want %d", got.Top, orig.Top)
	}
	if got.Width != orig.Width {
		t.Errorf("Width: got %v, want %v", got.Width, orig.Width)
	}
	if got.Height != orig.Height {
		t.Errorf("Height: got %v, want %v", got.Height, orig.Height)
	}
	if got.Visible != orig.Visible {
		t.Errorf("Visible: got %v, want %v", got.Visible, orig.Visible)
	}
	if got.Text != orig.Text {
		t.Errorf("Text: got %q, want %q", got.Text, orig.Text)
	}
	if got.Child != nil {
		t.Errorf("expected no child, got %+v", got.Child)
	}
}

func TestRoundTripWithChild(t *testing.T) {
	orig := &textObject{
		Name:    "Parent",
		Left:    0,
		Width:   100,
		Visible: false,
		Text:    "outer",
		Child:   &childObject{Label: "inner label", Value: 99},
	}
	got := roundTrip(t, orig)

	if got.Name != orig.Name {
		t.Errorf("Name: got %q, want %q", got.Name, orig.Name)
	}
	if got.Visible != orig.Visible {
		t.Errorf("Visible: got %v, want %v", got.Visible, orig.Visible)
	}
	if got.Child == nil {
		t.Fatal("expected child, got nil")
	}
	if got.Child.Label != orig.Child.Label {
		t.Errorf("Child.Label: got %q, want %q", got.Child.Label, orig.Child.Label)
	}
	if got.Child.Value != orig.Child.Value {
		t.Errorf("Child.Value: got %d, want %d", got.Child.Value, orig.Child.Value)
	}
}

func TestRoundTripBoolFalse(t *testing.T) {
	orig := &textObject{Name: "X", Visible: false}
	got := roundTrip(t, orig)
	if got.Visible != false {
		t.Errorf("Visible: got %v, want false", got.Visible)
	}
}

func TestRoundTripZeroValues(t *testing.T) {
	orig := &textObject{}
	got := roundTrip(t, orig)
	if got.Name != "" || got.Left != 0 || got.Width != 0 || got.Visible != false || got.Text != "" {
		t.Errorf("zero-value round-trip failed: got %+v", got)
	}
}

func TestRoundTripSpecialChars(t *testing.T) {
	orig := &textObject{
		Name: "special",
		Text: `Hello <World> & "friends"`,
	}
	got := roundTrip(t, orig)
	if got.Text != orig.Text {
		t.Errorf("Text: got %q, want %q", got.Text, orig.Text)
	}
}

// ── typeNameOf tests ──────────────────────────────────────────────────────────

func TestTypeNameOfWithNamer(t *testing.T) {
	obj := &textObject{}
	if got := typeNameOf(obj); got != "TextObject" {
		t.Errorf("got %q, want TextObject", got)
	}
}

func TestTypeNameOfFallback(t *testing.T) {
	// anonymousObj does NOT implement TypeNamer.
	obj := &anonymousObj{}
	got := typeNameOf(obj)
	if got != "anonymousObj" {
		t.Errorf("got %q, want anonymousObj", got)
	}
}

type anonymousObj struct{}

func (a *anonymousObj) Serialize(w report.Writer) error   { return nil }
func (a *anonymousObj) Deserialize(r report.Reader) error { return nil }

// ── WriteObject via TypeNamer interface ───────────────────────────────────────

func TestWriteObjectUsesTypeName(t *testing.T) {
	child := &childObject{Label: "lbl", Value: 7}

	var buf bytes.Buffer
	w := NewWriter(&buf)
	if err := w.BeginObject("Parent"); err != nil {
		t.Fatal(err)
	}
	if err := w.WriteObject(child); err != nil {
		t.Fatal(err)
	}
	if err := w.EndObject(); err != nil {
		t.Fatal(err)
	}
	if err := w.Flush(); err != nil {
		t.Fatal(err)
	}
	out := buf.String()
	if !strings.Contains(out, `<ChildObject`) {
		t.Errorf("expected <ChildObject in output:\n%s", out)
	}
}

// ── addAttr no-op when stack is empty ────────────────────────────────────────

func TestWriteStrNoStack(t *testing.T) {
	// Writing attributes with no open element should not panic.
	var buf bytes.Buffer
	w := NewWriter(&buf)
	// No BeginObject called.
	w.WriteStr("X", "y")
	w.WriteInt("N", 1)
	w.WriteBool("B", true)
	w.WriteFloat("F", 1.0)
	// Nothing written — just verify no panic and Flush is fine.
	if err := w.Flush(); err != nil {
		t.Fatalf("Flush: %v", err)
	}
}

// ── flushPending on already-flushed element ───────────────────────────────────

func TestFlushPendingAlreadyFlushed(t *testing.T) {
	// Open a parent, then open two children in sequence — the parent will be
	// flushed when the first child is opened, and flushPending on an already-
	// flushed element should be a no-op.
	var buf bytes.Buffer
	w := NewWriter(&buf)
	if err := w.BeginObject("Parent"); err != nil {
		t.Fatal(err)
	}
	// First child: causes parent's start tag to be flushed.
	if err := w.BeginObject("Child1"); err != nil {
		t.Fatal(err)
	}
	if err := w.EndObject(); err != nil {
		t.Fatal(err)
	}
	// Second child: flushPending on already-flushed Parent is a no-op.
	if err := w.BeginObject("Child2"); err != nil {
		t.Fatal(err)
	}
	if err := w.EndObject(); err != nil {
		t.Fatal(err)
	}
	if err := w.EndObject(); err != nil {
		t.Fatal(err)
	}
	if err := w.Flush(); err != nil {
		t.Fatal(err)
	}
	out := buf.String()
	if !strings.Contains(out, `Child1`) || !strings.Contains(out, `Child2`) {
		t.Errorf("expected both children in output:\n%s", out)
	}
}

// ── ReadObjectHeader skips processing instructions / char data ────────────────

func TestReadObjectHeaderSkipsNonElements(t *testing.T) {
	// A ProcInst before the root element should be skipped.
	src := `<?xml version="1.0" encoding="utf-8"?><Root Attr="val"/>`
	r := NewReader(strings.NewReader(src))
	typeName, ok := r.ReadObjectHeader()
	if !ok || typeName != "Root" {
		t.Fatalf("got typeName=%q ok=%v, want Root true", typeName, ok)
	}
	if got := r.ReadStr("Attr", ""); got != "val" {
		t.Errorf("got Attr=%q, want val", got)
	}
}

// ── NextChild error path ──────────────────────────────────────────────────────

func TestNextChildEOFReturnsNotOk(t *testing.T) {
	// Feed a document where we've already consumed the element — the next
	// Token call inside NextChild will return io.EOF.
	src := `<Parent/>`
	r := NewReader(strings.NewReader(src))
	r.ReadObjectHeader() // consumes <Parent/> start + its self-closing end
	// NextChild will get EOF from the decoder.
	_, ok := r.NextChild()
	if ok {
		t.Error("expected ok=false when decoder returns EOF in NextChild")
	}
}

// ── WriteObjectNamed serialise error propagation ──────────────────────────────

// errObj always returns an error from Serialize.
type errObj struct{}

func (e *errObj) TypeName() string                        { return "ErrObj" }
func (e *errObj) Serialize(w report.Writer) error         { return fmt.Errorf("serialize error") }
func (e *errObj) Deserialize(r report.Reader) error       { return nil }

func TestWriteObjectPropagatesSerializeError(t *testing.T) {
	var buf bytes.Buffer
	w := NewWriter(&buf)
	err := w.WriteObjectNamed("ErrObj", &errObj{})
	if err == nil || err.Error() != "serialize error" {
		t.Errorf("expected serialize error, got %v", err)
	}
}

func TestWriteObjectPropagatesSerializeErrorViaInterface(t *testing.T) {
	var buf bytes.Buffer
	w := NewWriter(&buf)
	if err := w.BeginObject("Root"); err != nil {
		t.Fatal(err)
	}
	err := w.WriteObject(&errObj{})
	if err == nil {
		t.Error("expected error propagated from WriteObject")
	}
}

// ── NewReaderWithPassword ──────────────────────────────────────────────────────

func TestNewReaderWithPassword_PlainXML(t *testing.T) {
	src := `<Report Name="test"/>`
	r, encrypted, err := NewReaderWithPassword(strings.NewReader(src), "anypassword")
	if err != nil {
		t.Fatalf("NewReaderWithPassword: %v", err)
	}
	if encrypted {
		t.Error("expected encrypted=false for plain XML")
	}
	typeName, ok := r.ReadObjectHeader()
	if !ok || typeName != "Report" {
		t.Errorf("ReadObjectHeader: got %q ok=%v, want Report true", typeName, ok)
	}
	if got := r.ReadStr("Name", ""); got != "test" {
		t.Errorf("Name = %q, want test", got)
	}
}

// ── ReadObjectHeader — CharData token skipping ─────────────────────────────────

func TestReadObjectHeaderSkipsCharData(t *testing.T) {
	// Whitespace before the root element is CharData — should be skipped.
	src := "   \n  <Root Attr=\"val\"/>"
	r := NewReader(strings.NewReader(src))
	typeName, ok := r.ReadObjectHeader()
	if !ok || typeName != "Root" {
		t.Errorf("got typeName=%q ok=%v, want Root true", typeName, ok)
	}
	if got := r.ReadStr("Attr", ""); got != "val" {
		t.Errorf("Attr = %q, want val", got)
	}
}

// ── NextChild — CharData skipping ─────────────────────────────────────────────

func TestNextChildSkipsCharData(t *testing.T) {
	// Text content between elements should be skipped.
	src := "<Parent>  <Child1/>  <Child2/>  </Parent>"
	r := NewReader(strings.NewReader(src))
	r.ReadObjectHeader() // <Parent>

	names := []string{}
	for {
		name, ok := r.NextChild()
		if !ok {
			break
		}
		names = append(names, name)
		if err := r.FinishChild(); err != nil {
			t.Fatalf("FinishChild: %v", err)
		}
	}
	if len(names) != 2 {
		t.Errorf("got %d children %v, want 2", len(names), names)
	}
}

// ── SkipElement — nested elements and error path ──────────────────────────────

func TestSkipElement_NoChildren(t *testing.T) {
	// Skip a self-closing element — skipRemainingContent immediately finds end.
	src := `<Parent><Empty/><Keep Name="k"/></Parent>`
	r := NewReader(strings.NewReader(src))
	r.ReadObjectHeader() // <Parent>

	name, ok := r.NextChild()
	if !ok || name != "Empty" {
		t.Fatalf("expected Empty, got %q ok=%v", name, ok)
	}
	if err := r.SkipElement(); err != nil {
		t.Fatalf("SkipElement Empty: %v", err)
	}
	if err := r.FinishChild(); err != nil {
		t.Fatalf("FinishChild Empty: %v", err)
	}

	name2, ok2 := r.NextChild()
	if !ok2 || name2 != "Keep" {
		t.Errorf("expected Keep, got %q ok=%v", name2, ok2)
	}
	if err := r.FinishChild(); err != nil {
		t.Fatalf("FinishChild Keep: %v", err)
	}
}

func TestSkipElement_ErrorPath(t *testing.T) {
	// Use a pipe that closes with error mid-stream, so skipRemainingContent fails.
	pr, pw := io.Pipe()
	go func() {
		pw.Write([]byte(`<Parent><Child>`)) //nolint:errcheck
		pw.CloseWithError(fmt.Errorf("stream closed"))
	}()

	r := NewReader(pr)
	_, ok := r.ReadObjectHeader() // <Parent>
	if !ok {
		t.Skip("ReadObjectHeader failed on pipe")
	}
	name, ok := r.NextChild()
	if !ok || name != "Child" {
		t.Skipf("didn't get Child: %q ok=%v", name, ok)
	}
	err := r.SkipElement()
	if err == nil {
		t.Error("expected error from SkipElement on truncated stream, got nil")
	}
}

// ── FinishChild — done=true path (end-element already consumed) ────────────────

func TestFinishChild_DonePath(t *testing.T) {
	// When NextChild returns ("", false), r.done=true. Subsequent FinishChild
	// should not try to skip remaining content (already consumed end tag).
	src := `<Parent><Child Label="x"/></Parent>`
	r := NewReader(strings.NewReader(src))
	r.ReadObjectHeader() // <Parent>

	name, ok := r.NextChild()
	if !ok || name != "Child" {
		t.Fatalf("expected Child: %q ok=%v", name, ok)
	}

	// Exhaust children — this reads </Parent> and sets done=true.
	_, ok2 := r.NextChild()
	if ok2 {
		t.Fatal("expected no more children")
	}

	// FinishChild should use the done=true fast-path (no skipRemainingContent).
	if err := r.FinishChild(); err != nil {
		t.Fatalf("FinishChild done path: %v", err)
	}
}

func TestFinishChild_SkippedPath(t *testing.T) {
	// When SkipElement is called first, r.skipped=true. FinishChild should
	// restore state without calling skipRemainingContent again.
	src := `<Parent><Child><Deep/></Child><Next/></Parent>`
	r := NewReader(strings.NewReader(src))
	r.ReadObjectHeader() // <Parent>

	name, ok := r.NextChild()
	if !ok || name != "Child" {
		t.Fatalf("expected Child: %q ok=%v", name, ok)
	}
	if err := r.SkipElement(); err != nil {
		t.Fatalf("SkipElement: %v", err)
	}
	// After SkipElement, skipped=true, so FinishChild skips skipRemainingContent.
	if err := r.FinishChild(); err != nil {
		t.Fatalf("FinishChild skipped: %v", err)
	}

	name2, ok2 := r.NextChild()
	if !ok2 || name2 != "Next" {
		t.Errorf("expected Next sibling, got %q ok=%v", name2, ok2)
	}
	if err := r.FinishChild(); err != nil {
		t.Fatalf("FinishChild Next: %v", err)
	}
}

// ── typeNameOf — no package separator ─────────────────────────────────────────

type noPackageObj struct{}

func (n *noPackageObj) Serialize(w report.Writer) error   { return nil }
func (n *noPackageObj) Deserialize(r report.Reader) error { return nil }

func TestTypeNameOf_NoPackageSeparator(t *testing.T) {
	// typeNameOf falls back to fmt.Sprintf("%T",...) which for types in the
	// same package includes "serial." prefix. The loop finds '.' and strips it.
	// For a type *without* '.', the whole name is returned.
	// Since all Go named types have package prefixes, this tests the normal path.
	obj := &noPackageObj{}
	got := typeNameOf(obj)
	if got == "" {
		t.Error("typeNameOf returned empty string")
	}
}

// ── BeginObject with pending attributes ───────────────────────────────────────

func TestBeginObject_WithPendingAttributes(t *testing.T) {
	var buf bytes.Buffer
	w := NewWriter(&buf)
	if err := w.BeginObject("Parent"); err != nil {
		t.Fatal(err)
	}
	w.WriteStr("ParentAttr", "pval")
	// BeginObject for child triggers flushPending for Parent.
	if err := w.BeginObject("Child"); err != nil {
		t.Fatal(err)
	}
	w.WriteStr("ChildAttr", "cval")
	if err := w.EndObject(); err != nil { // close Child
		t.Fatal(err)
	}
	if err := w.EndObject(); err != nil { // close Parent
		t.Fatal(err)
	}
	if err := w.Flush(); err != nil {
		t.Fatal(err)
	}
	out := buf.String()
	if !strings.Contains(out, `ParentAttr="pval"`) {
		t.Errorf("expected ParentAttr in output:\n%s", out)
	}
	if !strings.Contains(out, `ChildAttr="cval"`) {
		t.Errorf("expected ChildAttr in output:\n%s", out)
	}
}

// ── EndObject — element already flushed (has children) ───────────────────────

func TestEndObject_AlreadyFlushed(t *testing.T) {
	var buf bytes.Buffer
	w := NewWriter(&buf)
	if err := w.BeginObject("Parent"); err != nil {
		t.Fatal(err)
	}
	// Open a child to force Parent's start tag to be flushed.
	if err := w.BeginObject("Child"); err != nil {
		t.Fatal(err)
	}
	if err := w.EndObject(); err != nil { // close Child
		t.Fatal(err)
	}
	// Parent is already flushed — EndObject should go straight to EndElement.
	if err := w.EndObject(); err != nil {
		t.Fatalf("EndObject on flushed parent: %v", err)
	}
	if err := w.Flush(); err != nil {
		t.Fatal(err)
	}
	out := buf.String()
	if !strings.Contains(out, `</Parent>`) {
		t.Errorf("expected </Parent> in output:\n%s", out)
	}
}

// ── WriteObject — uses TypeNamer ───────────────────────────────────────────────

func TestWriteObject_UsesTypeNameOfFallback(t *testing.T) {
	// anonymousObj does NOT implement TypeNamer — typeNameOf falls back.
	obj := &anonymousObj{}
	var buf bytes.Buffer
	w := NewWriter(&buf)
	if err := w.WriteObject(obj); err != nil {
		t.Fatalf("WriteObject: %v", err)
	}
	if err := w.Flush(); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(buf.String(), "anonymousObj") {
		t.Errorf("expected type name in output:\n%s", buf.String())
	}
}

// ── WriteObjectNamed — covers BeginObject + Serialize + EndObject path ────────

func TestWriteObjectNamed_PendingAttrs(t *testing.T) {
	child := &childObject{Label: "lbl2", Value: 22}
	var buf bytes.Buffer
	w := NewWriter(&buf)
	if err := w.WriteObjectNamed("ChildObject", child); err != nil {
		t.Fatalf("WriteObjectNamed: %v", err)
	}
	if err := w.Flush(); err != nil {
		t.Fatal(err)
	}
	out := buf.String()
	if !strings.Contains(out, `Label="lbl2"`) {
		t.Errorf("expected Label in output:\n%s", out)
	}
}

// ── NewReaderWithPassword — error path ────────────────────────────────────────

func TestNewReaderWithPassword_Error(t *testing.T) {
	// Close the pipe with an error immediately so PeekAndDecrypt fails.
	pr, pw := io.Pipe()
	pw.CloseWithError(fmt.Errorf("simulated read error"))
	_, _, err := NewReaderWithPassword(pr, "password")
	if err == nil {
		t.Error("expected error when underlying reader fails, got nil")
	}
}

// ── ReadObjectHeader — EndElement case ────────────────────────────────────────

func TestReadObjectHeaderEndElementCase(t *testing.T) {
	// After ReadObjectHeader reads the StartElement of a self-closing element,
	// calling it again reads the matching EndElement → returns ("", false).
	src := `<First/><Second/>`
	r := NewReader(strings.NewReader(src))

	typeName, ok := r.ReadObjectHeader() // reads StartElement{First}
	if !ok || typeName != "First" {
		t.Fatalf("first ReadObjectHeader: got %q ok=%v, want First true", typeName, ok)
	}
	// Decoder has EndElement{First} pending.
	typeName2, ok2 := r.ReadObjectHeader() // reads EndElement{First} → case xml.EndElement
	if ok2 {
		t.Errorf("second ReadObjectHeader on EndElement: expected ok=false, got true (%q)", typeName2)
	}
}

// ── NextChild — error/EOF path ─────────────────────────────────────────────────

func TestNextChild_ErrorPath(t *testing.T) {
	// Pipe closes with error mid-stream — NextChild must handle the err path.
	pr, pw := io.Pipe()
	go func() {
		pw.Write([]byte(`<Parent>`)) //nolint:errcheck
		pw.CloseWithError(fmt.Errorf("stream broken"))
	}()

	r := NewReader(pr)
	_, ok := r.ReadObjectHeader() // reads <Parent>
	if !ok {
		t.Skip("ReadObjectHeader failed; pipe may have been too slow")
	}
	// NextChild tries to read the next token — gets an error from the broken pipe.
	_, ok2 := r.NextChild()
	if ok2 {
		t.Error("expected ok=false when NextChild encounters a read error")
	}
}

// ── FinishChild — skipRemainingContent error ───────────────────────────────────

func TestFinishChild_SkipRemainingContentError(t *testing.T) {
	// Enter a child element, then break the stream before the end tag is
	// available so that FinishChild's skipRemainingContent returns an error.
	pr, pw := io.Pipe()
	go func() {
		pw.Write([]byte(`<Parent><Child>`)) //nolint:errcheck
		pw.CloseWithError(fmt.Errorf("stream closed"))
	}()

	r := NewReader(pr)
	_, ok := r.ReadObjectHeader() // <Parent>
	if !ok {
		t.Skip("ReadObjectHeader failed")
	}
	name, ok2 := r.NextChild() // <Child>
	if !ok2 || name != "Child" {
		t.Skipf("NextChild did not return Child: %q ok=%v", name, ok2)
	}
	// FinishChild must call skipRemainingContent (skipped=false, done=false),
	// which tries to read </Child> but gets an error instead.
	err := r.FinishChild()
	if err == nil {
		t.Error("expected error from FinishChild when stream breaks, got nil")
	}
}

// ── flushPending — empty stack (no-op) ────────────────────────────────────────

func TestFlushPending_EmptyStack(t *testing.T) {
	// flushPending with empty stack returns nil (no-op).
	var buf bytes.Buffer
	w := NewWriter(&buf)
	// Call BeginObject without anything on the stack first — flushPending
	// is called internally and should return nil for empty stack.
	if err := w.BeginObject("OnlyOne"); err != nil {
		t.Fatalf("BeginObject: %v", err)
	}
	if err := w.EndObject(); err != nil {
		t.Fatalf("EndObject: %v", err)
	}
	if err := w.Flush(); err != nil {
		t.Fatal(err)
	}
}
