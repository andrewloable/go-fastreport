package report_test

import (
	"testing"

	"github.com/loabletech/go-fastreport/report"
)

// --- helpers ---

func newBO() *report.BaseObject {
	return report.NewBaseObject()
}

// parentStub is a minimal Parent that tracks children.
type parentStub struct {
	children []report.Base
}

func (p *parentStub) CanContain(report.Base) bool { return true }
func (p *parentStub) GetChildObjects(list *[]report.Base) {
	*list = append(*list, p.children...)
}
func (p *parentStub) AddChild(child report.Base) {
	p.children = append(p.children, child)
	child.SetParent(p)
}
func (p *parentStub) RemoveChild(child report.Base) {
	for i, c := range p.children {
		if c == child {
			p.children = append(p.children[:i], p.children[i+1:]...)
			return
		}
	}
}
func (p *parentStub) GetChildOrder(child report.Base) int {
	for i, c := range p.children {
		if c == child {
			return i
		}
	}
	return -1
}
func (p *parentStub) SetChildOrder(child report.Base, order int) {}
func (p *parentStub) UpdateLayout(dx, dy float32)                {}

// baseChild is a BaseObject that also implements Parent so we can nest.
type baseChild struct {
	*report.BaseObject
	parentStub
}

func newBaseChild(name string) *baseChild {
	bc := &baseChild{BaseObject: report.NewBaseObject()}
	bc.SetName(name)
	return bc
}

// Disambiguate Parent() and SetParent() between embedded BaseObject and parentStub.
func (bc *baseChild) Parent() report.Parent    { return bc.BaseObject.Parent() }
func (bc *baseChild) SetParent(p report.Parent) { bc.BaseObject.SetParent(p) }
func (bc *baseChild) CanContain(report.Base) bool { return true }
func (bc *baseChild) GetChildObjects(list *[]report.Base) {
	bc.parentStub.GetChildObjects(list)
}

// --- Tests ---

func TestNewBaseObject_DefaultFlags(t *testing.T) {
	b := newBO()
	if b.Flags() != report.DefaultFlags {
		t.Errorf("default flags = %d, want %d", b.Flags(), report.DefaultFlags)
	}
}

func TestName(t *testing.T) {
	b := newBO()
	if b.Name() != "" {
		t.Error("initial name should be empty")
	}
	b.SetName("MyObject")
	if b.Name() != "MyObject" {
		t.Errorf("Name() = %q, want MyObject", b.Name())
	}
}

func TestBaseName(t *testing.T) {
	b := newBO()
	if b.BaseName() != "" {
		t.Error("initial base name should be empty")
	}
	b.SetBaseName("Text")
	if b.BaseName() != "Text" {
		t.Errorf("BaseName() = %q, want Text", b.BaseName())
	}
}

func TestParent(t *testing.T) {
	b := newBO()
	if b.Parent() != nil {
		t.Error("initial parent should be nil")
	}
	p := &parentStub{}
	b.SetParent(p)
	if b.Parent() != p {
		t.Error("Parent() should be the set parent")
	}
	b.SetParent(nil)
	if b.Parent() != nil {
		t.Error("Parent() should be nil after reset")
	}
}

func TestSetFlag(t *testing.T) {
	b := newBO()
	// Clear CanMove
	b.SetFlag(report.CanMove, false)
	if b.HasFlag(report.CanMove) {
		t.Error("CanMove should be cleared")
	}
	// Set it back
	b.SetFlag(report.CanMove, true)
	if !b.HasFlag(report.CanMove) {
		t.Error("CanMove should be set")
	}
}

func TestHasFlag_AllFlags(t *testing.T) {
	flags := []report.ObjectFlags{
		report.FlagsNone,
		report.CanMove,
		report.CanResize,
		report.CanDelete,
		report.CanEdit,
		report.CanChangeOrder,
		report.CanChangeParent,
		report.CanCopy,
		report.CanDraw,
		report.CanGroup,
		report.CanWriteChildren,
		report.CanWriteBounds,
		report.HasSmartTag,
		report.HasGlobalName,
		report.CanShowChildrenInReportTree,
		report.InterceptsPreviewMouseEvents,
	}

	b := newBO()
	b.SetFlag(report.FlagsNone, true) // no-op but covers the branch
	for _, f := range flags {
		if f == report.FlagsNone {
			continue
		}
		b.SetFlag(f, true)
		if !b.HasFlag(f) {
			t.Errorf("HasFlag(%d) should be true after SetFlag(true)", f)
		}
		b.SetFlag(f, false)
		if b.HasFlag(f) {
			t.Errorf("HasFlag(%d) should be false after SetFlag(false)", f)
		}
	}
}

func TestRestrictions(t *testing.T) {
	b := newBO()
	if b.Restrictions() != report.RestrictionsNone {
		t.Error("initial restrictions should be None")
	}
	r := report.DontMove | report.DontDelete
	b.SetRestrictions(r)
	if b.Restrictions() != r {
		t.Errorf("Restrictions() = %d, want %d", b.Restrictions(), r)
	}
}

func TestTag(t *testing.T) {
	b := newBO()
	if b.Tag() != nil {
		t.Error("initial tag should be nil")
	}
	b.SetTag("hello")
	if b.Tag() != "hello" {
		t.Errorf("Tag() = %v, want hello", b.Tag())
	}
	b.SetTag(42)
	if b.Tag() != 42 {
		t.Errorf("Tag() = %v, want 42", b.Tag())
	}
}

func TestZOrder(t *testing.T) {
	b := newBO()
	if b.ZOrder() != 0 {
		t.Error("initial zOrder should be 0")
	}
}

func TestObjectState(t *testing.T) {
	b := newBO()

	if b.IsAncestor() {
		t.Error("IsAncestor should be false initially")
	}
	b.SetObjectState(report.IsAncestorState, true)
	if !b.IsAncestor() {
		t.Error("IsAncestor should be true")
	}
	b.SetObjectState(report.IsAncestorState, false)
	if b.IsAncestor() {
		t.Error("IsAncestor should be false after clear")
	}

	b.SetObjectState(report.IsDesigningState, true)
	if !b.IsDesigning() {
		t.Error("IsDesigning should be true")
	}
	b.SetObjectState(report.IsDesigningState, false)

	b.SetObjectState(report.IsPrintingState, true)
	if !b.IsPrinting() {
		t.Error("IsPrinting should be true")
	}
	b.SetObjectState(report.IsPrintingState, false)

	b.SetObjectState(report.IsRunningState, true)
	if !b.IsRunning() {
		t.Error("IsRunning should be true")
	}
	b.SetObjectState(report.IsRunningState, false)

	// IsDeserializingState just exercises the flag path
	b.SetObjectState(report.IsDeserializingState, true)
	if !b.GetObjectState(report.IsDeserializingState) {
		t.Error("IsDeserializingState should be true")
	}
	b.SetObjectState(report.IsDeserializingState, false)
}

func TestChildObjects(t *testing.T) {
	p := &parentStub{}
	c1 := newBO()
	c1.SetName("c1")
	c2 := newBO()
	c2.SetName("c2")
	p.AddChild(c1)
	p.AddChild(c2)

	got := report.ChildObjects(p)
	if len(got) != 2 {
		t.Fatalf("ChildObjects returned %d, want 2", len(got))
	}
}

func TestFindObject_DirectHit(t *testing.T) {
	a := newBO()
	a.SetName("Alpha")
	b := newBO()
	b.SetName("Beta")

	found := report.FindObject("Beta", []report.Base{a, b})
	if found != b {
		t.Error("FindObject did not find Beta")
	}
}

func TestFindObject_NotFound(t *testing.T) {
	a := newBO()
	a.SetName("Alpha")
	found := report.FindObject("Missing", []report.Base{a})
	if found != nil {
		t.Error("FindObject should return nil when not found")
	}
}

func TestFindObject_Recursive(t *testing.T) {
	parent := newBaseChild("Root")
	child := newBaseChild("Leaf")
	parent.AddChild(child)

	found := report.FindObject("Leaf", []report.Base{parent})
	if found == nil {
		t.Fatal("FindObject should find Leaf recursively")
	}
	if found.Name() != "Leaf" {
		t.Errorf("found.Name() = %q, want Leaf", found.Name())
	}
}

func TestFindObject_EmptyList(t *testing.T) {
	found := report.FindObject("X", nil)
	if found != nil {
		t.Error("FindObject on nil list should return nil")
	}
}

func TestSerialize_DefaultsNotWritten(t *testing.T) {
	b := newBO()
	// name is empty, restrictions is None, flags is DefaultFlags — nothing written except flags
	// (flags == DefaultFlags so flags not written, name=="" so not written, restrictions==None so not written)
	w := newMockWriter()
	if err := b.Serialize(w); err != nil {
		t.Fatalf("Serialize error: %v", err)
	}
	if _, ok := w.strings["Name"]; ok {
		t.Error("Name should not be written when empty")
	}
	if _, ok := w.ints["Restrictions"]; ok {
		t.Error("Restrictions should not be written when None")
	}
	if _, ok := w.ints["Flags"]; ok {
		t.Error("Flags should not be written when DefaultFlags")
	}
}

func TestSerialize_NonDefaults(t *testing.T) {
	b := newBO()
	b.SetName("Obj1")
	b.SetRestrictions(report.DontMove)
	b.SetFlag(report.HasSmartTag, true)
	b.SetFlag(report.CanMove, false) // make flags != DefaultFlags

	w := newMockWriter()
	if err := b.Serialize(w); err != nil {
		t.Fatalf("Serialize error: %v", err)
	}
	if w.strings["Name"] != "Obj1" {
		t.Errorf("Name not serialized correctly: %q", w.strings["Name"])
	}
	if w.ints["Restrictions"] != int(report.DontMove) {
		t.Errorf("Restrictions not serialized correctly: %d", w.ints["Restrictions"])
	}
	if _, ok := w.ints["Flags"]; !ok {
		t.Error("Flags should be written when not DefaultFlags")
	}
}

func TestDeserialize(t *testing.T) {
	b := newBO()
	r := newMockReader()
	r.strings["Name"] = "Loaded"
	r.ints["Restrictions"] = int(report.DontEdit | report.DontDelete)
	r.ints["Flags"] = int(report.CanMove | report.CanResize)

	if err := b.Deserialize(r); err != nil {
		t.Fatalf("Deserialize error: %v", err)
	}
	if b.Name() != "Loaded" {
		t.Errorf("Name = %q, want Loaded", b.Name())
	}
	if b.Restrictions() != report.DontEdit|report.DontDelete {
		t.Errorf("Restrictions = %d", b.Restrictions())
	}
	if b.Flags() != report.CanMove|report.CanResize {
		t.Errorf("Flags = %d", b.Flags())
	}
}

func TestDeserialize_Defaults(t *testing.T) {
	b := newBO()
	b.SetName("Prior")
	// reader returns defaults (nothing set) — existing values kept
	r := newMockReader()
	if err := b.Deserialize(r); err != nil {
		t.Fatalf("Deserialize error: %v", err)
	}
	if b.Name() != "Prior" {
		t.Errorf("Name should retain prior value when not in reader, got %q", b.Name())
	}
}

// Compile-time check: *BaseObject satisfies report.Base.
var _ report.Base = (*report.BaseObject)(nil)

// Restrictions constant coverage
func TestRestrictionConstants(t *testing.T) {
	vals := []report.Restrictions{
		report.RestrictionsNone,
		report.DontMove,
		report.DontResize,
		report.DontModify,
		report.DontEdit,
		report.DontDelete,
		report.HideAllProperties,
	}
	for i, v := range vals {
		_ = i
		_ = v
	}
}
