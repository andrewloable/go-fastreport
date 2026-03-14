package report_test

import (
	"testing"

	"github.com/andrewloable/go-fastreport/report"
)

// --- Mock implementations ---

// mockWriter implements report.Writer for testing.
type mockWriter struct {
	strings map[string]string
	ints    map[string]int
	bools   map[string]bool
	floats  map[string]float32
	objects []report.Serializable
}

func newMockWriter() *mockWriter {
	return &mockWriter{
		strings: make(map[string]string),
		ints:    make(map[string]int),
		bools:   make(map[string]bool),
		floats:  make(map[string]float32),
	}
}

func (w *mockWriter) WriteStr(name, value string)         { w.strings[name] = value }
func (w *mockWriter) WriteInt(name string, value int)     { w.ints[name] = value }
func (w *mockWriter) WriteBool(name string, value bool)   { w.bools[name] = value }
func (w *mockWriter) WriteFloat(name string, value float32) { w.floats[name] = value }
func (w *mockWriter) WriteObject(obj report.Serializable) error {
	w.objects = append(w.objects, obj)
	return nil
}

// mockReader implements report.Reader for testing.
type mockReader struct {
	strings  map[string]string
	ints     map[string]int
	bools    map[string]bool
	floats   map[string]float32
	children []string
	childIdx int
}

func newMockReader() *mockReader {
	return &mockReader{
		strings: make(map[string]string),
		ints:    make(map[string]int),
		bools:   make(map[string]bool),
		floats:  make(map[string]float32),
	}
}

func (r *mockReader) ReadStr(name, def string) string {
	if v, ok := r.strings[name]; ok {
		return v
	}
	return def
}
func (r *mockReader) ReadInt(name string, def int) int {
	if v, ok := r.ints[name]; ok {
		return v
	}
	return def
}
func (r *mockReader) ReadBool(name string, def bool) bool {
	if v, ok := r.bools[name]; ok {
		return v
	}
	return def
}
func (r *mockReader) ReadFloat(name string, def float32) float32 {
	if v, ok := r.floats[name]; ok {
		return v
	}
	return def
}
func (r *mockReader) NextChild() (string, bool) {
	if r.childIdx >= len(r.children) {
		return "", false
	}
	t := r.children[r.childIdx]
	r.childIdx++
	return t, true
}
func (r *mockReader) FinishChild() error { return nil }

// simpleObject implements report.Base for testing.
type simpleObject struct {
	name   string
	parent report.Parent
}

func (o *simpleObject) Name() string          { return o.name }
func (o *simpleObject) SetName(n string)      { o.name = n }
func (o *simpleObject) BaseName() string      { return "Simple" }
func (o *simpleObject) Parent() report.Parent { return o.parent }
func (o *simpleObject) SetParent(p report.Parent) { o.parent = p }
func (o *simpleObject) Serialize(w report.Writer) error {
	w.WriteStr("Name", o.name)
	return nil
}
func (o *simpleObject) Deserialize(r report.Reader) error {
	o.name = r.ReadStr("Name", "")
	return nil
}

// simpleParent implements report.Parent for testing.
type simpleParent struct {
	children []report.Base
}

func (p *simpleParent) CanContain(child report.Base) bool { return true }
func (p *simpleParent) GetChildObjects(list *[]report.Base) {
	*list = append(*list, p.children...)
}
func (p *simpleParent) AddChild(child report.Base) {
	p.children = append(p.children, child)
	child.SetParent(p)
}
func (p *simpleParent) RemoveChild(child report.Base) {
	for i, c := range p.children {
		if c == child {
			p.children = append(p.children[:i], p.children[i+1:]...)
			return
		}
	}
}
func (p *simpleParent) GetChildOrder(child report.Base) int {
	for i, c := range p.children {
		if c == child {
			return i
		}
	}
	return -1
}
func (p *simpleParent) SetChildOrder(child report.Base, order int) {
	idx := p.GetChildOrder(child)
	if idx < 0 {
		return
	}
	p.children = append(p.children[:idx], p.children[idx+1:]...)
	if order > len(p.children) {
		order = len(p.children)
	}
	newChildren := make([]report.Base, order+1+len(p.children)-order)
	copy(newChildren, p.children[:order])
	newChildren[order] = child
	copy(newChildren[order+1:], p.children[order:])
	p.children = newChildren
}
func (p *simpleParent) UpdateLayout(dx, dy float32) {}

// --- Tests ---

func TestBaseInterface(t *testing.T) {
	obj := &simpleObject{name: "Test1"}
	if obj.Name() != "Test1" {
		t.Errorf("Name() = %q, want Test1", obj.Name())
	}
	obj.SetName("Text2")
	if obj.Name() != "Text2" {
		t.Errorf("SetName: Name() = %q, want Text2", obj.Name())
	}
	if obj.BaseName() != "Simple" {
		t.Errorf("BaseName() = %q, want Simple", obj.BaseName())
	}
	if obj.Parent() != nil {
		t.Error("Parent() should be nil initially")
	}
}

func TestSerializable(t *testing.T) {
	obj := &simpleObject{name: "MyObj"}
	w := newMockWriter()
	if err := obj.Serialize(w); err != nil {
		t.Fatalf("Serialize error: %v", err)
	}
	if w.strings["Name"] != "MyObj" {
		t.Errorf("serialized Name = %q, want MyObj", w.strings["Name"])
	}

	obj2 := &simpleObject{}
	r := newMockReader()
	r.strings["Name"] = "MyObj"
	if err := obj2.Deserialize(r); err != nil {
		t.Fatalf("Deserialize error: %v", err)
	}
	if obj2.Name() != "MyObj" {
		t.Errorf("deserialized Name = %q, want MyObj", obj2.Name())
	}
}

func TestParentInterface(t *testing.T) {
	parent := &simpleParent{}
	child1 := &simpleObject{name: "c1"}
	child2 := &simpleObject{name: "c2"}
	child3 := &simpleObject{name: "c3"}

	parent.AddChild(child1)
	parent.AddChild(child2)
	parent.AddChild(child3)

	var children []report.Base
	parent.GetChildObjects(&children)
	if len(children) != 3 {
		t.Fatalf("expected 3 children, got %d", len(children))
	}

	if parent.GetChildOrder(child1) != 0 {
		t.Errorf("child1 order = %d, want 0", parent.GetChildOrder(child1))
	}
	if parent.GetChildOrder(child2) != 1 {
		t.Errorf("child2 order = %d, want 1", parent.GetChildOrder(child2))
	}

	parent.RemoveChild(child2)
	children = nil
	parent.GetChildObjects(&children)
	if len(children) != 2 {
		t.Fatalf("expected 2 children after remove, got %d", len(children))
	}

	if !parent.CanContain(child1) {
		t.Error("CanContain should return true")
	}

	if child1.Parent() != parent {
		t.Error("child1.Parent() should be the parent")
	}
}

func TestWriterInterface(t *testing.T) {
	w := newMockWriter()
	w.WriteStr("key", "value")
	w.WriteInt("n", 42)
	w.WriteBool("flag", true)
	w.WriteFloat("f", 3.14)

	if w.strings["key"] != "value" {
		t.Errorf("WriteStr: got %q", w.strings["key"])
	}
	if w.ints["n"] != 42 {
		t.Errorf("WriteInt: got %d", w.ints["n"])
	}
	if !w.bools["flag"] {
		t.Error("WriteBool: expected true")
	}
	if w.floats["f"] != 3.14 {
		t.Errorf("WriteFloat: got %v", w.floats["f"])
	}
}

func TestReaderDefaults(t *testing.T) {
	r := newMockReader()
	if r.ReadStr("missing", "default") != "default" {
		t.Error("ReadStr should return default when key absent")
	}
	if r.ReadInt("missing", 99) != 99 {
		t.Error("ReadInt should return default when key absent")
	}
	if r.ReadBool("missing", true) != true {
		t.Error("ReadBool should return default when key absent")
	}
	if r.ReadFloat("missing", 1.5) != 1.5 {
		t.Error("ReadFloat should return default when key absent")
	}
}

func TestReaderNextChild(t *testing.T) {
	r := newMockReader()
	r.children = []string{"TextObject", "BandBase"}

	typeName, ok := r.NextChild()
	if !ok || typeName != "TextObject" {
		t.Errorf("NextChild #1: got (%q, %v)", typeName, ok)
	}
	typeName, ok = r.NextChild()
	if !ok || typeName != "BandBase" {
		t.Errorf("NextChild #2: got (%q, %v)", typeName, ok)
	}
	typeName, ok = r.NextChild()
	if ok || typeName != "" {
		t.Errorf("NextChild exhausted: got (%q, %v)", typeName, ok)
	}
}

// Verify that the interface types are usable as variables (compile-time checks).
var _ report.Serializable = (*simpleObject)(nil)
var _ report.Base = (*simpleObject)(nil)
var _ report.Parent = (*simpleParent)(nil)
var _ report.Writer = (*mockWriter)(nil)
var _ report.Reader = (*mockReader)(nil)
