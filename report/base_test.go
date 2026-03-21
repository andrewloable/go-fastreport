package report_test

import (
	"testing"

	"github.com/andrewloable/go-fastreport/report"
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
func (p *parentStub) SetChildOrder(child report.Base, order int) {
	// Find and remove the child from its current position.
	idx := -1
	for i, c := range p.children {
		if c == child {
			idx = i
			break
		}
	}
	if idx < 0 {
		return
	}
	p.children = append(p.children[:idx], p.children[idx+1:]...)
	// Clamp order to valid range.
	if order < 0 {
		order = 0
	}
	if order > len(p.children) {
		order = len(p.children)
	}
	// Insert at the requested position.
	p.children = append(p.children[:order], append([]report.Base{child}, p.children[order:]...)...)
}
func (p *parentStub) UpdateLayout(dx, dy float32) {}

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
func (bc *baseChild) Parent() report.Parent      { return bc.BaseObject.Parent() }
func (bc *baseChild) SetParent(p report.Parent)  { bc.BaseObject.SetParent(p) }
func (bc *baseChild) CanContain(report.Base) bool { return true }
func (bc *baseChild) GetChildObjects(list *[]report.Base) {
	bc.parentStub.GetChildObjects(list)
}

// AddChild overrides parentStub.AddChild to correctly set bc (the full *baseChild)
// as the child's parent, not the embedded parentStub.
func (bc *baseChild) AddChild(child report.Base) {
	bc.parentStub.children = append(bc.parentStub.children, child)
	child.SetParent(bc)
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

// --- HasRestriction ---

func TestHasRestriction_True(t *testing.T) {
	b := newBO()
	b.SetRestrictions(report.DontMove | report.DontDelete)
	if !b.HasRestriction(report.DontMove) {
		t.Error("HasRestriction(DontMove) should be true")
	}
	if !b.HasRestriction(report.DontDelete) {
		t.Error("HasRestriction(DontDelete) should be true")
	}
}

func TestHasRestriction_False(t *testing.T) {
	b := newBO()
	b.SetRestrictions(report.DontMove)
	if b.HasRestriction(report.DontDelete) {
		t.Error("HasRestriction(DontDelete) should be false when only DontMove is set")
	}
}

func TestHasRestriction_None(t *testing.T) {
	b := newBO()
	if b.HasRestriction(report.DontMove) {
		t.Error("HasRestriction on fresh BaseObject (RestrictionsNone) should be false")
	}
}

// --- SetZOrder / ZOrder with parent ---

func TestZOrder_WithParent_DelegatesToParent(t *testing.T) {
	p := &parentStub{}
	c1 := newBO()
	c1.SetName("c1")
	c2 := newBO()
	c2.SetName("c2")
	p.AddChild(c1)
	p.AddChild(c2)

	// c1 is at index 0 as added; ZOrder should reflect parent's GetChildOrder.
	if c1.ZOrder() != 0 {
		t.Errorf("ZOrder(c1) = %d, want 0", c1.ZOrder())
	}
	if c2.ZOrder() != 1 {
		t.Errorf("ZOrder(c2) = %d, want 1", c2.ZOrder())
	}
}

func TestZOrder_WithoutParent_UsesInternalField(t *testing.T) {
	b := newBO()
	b.SetZOrder(5)
	if b.ZOrder() != 5 {
		t.Errorf("ZOrder() without parent = %d, want 5", b.ZOrder())
	}
}

func TestSetZOrder_WithParent_DelegatesToParent(t *testing.T) {
	p := &parentStub{}
	c1 := newBO()
	c1.SetName("c1")
	c2 := newBO()
	c2.SetName("c2")
	p.AddChild(c1)
	p.AddChild(c2)

	// Move c2 to index 0 via SetZOrder.
	c2.SetZOrder(0)
	if c1.ZOrder() != 1 {
		t.Errorf("after SetZOrder: c1 should move to index 1, got %d", c1.ZOrder())
	}
}

// --- HasParent ---

func TestHasParent_DirectParent(t *testing.T) {
	p := &parentStub{}
	child := newBO()
	p.AddChild(child)

	if !report.HasParent(child, p) {
		t.Error("HasParent should return true when p is direct parent")
	}
}

func TestHasParent_NoParent(t *testing.T) {
	p := &parentStub{}
	child := newBO()

	if report.HasParent(child, p) {
		t.Error("HasParent should return false when object has no parent")
	}
}

func TestHasParent_DifferentParent(t *testing.T) {
	p1 := &parentStub{}
	p2 := &parentStub{}
	child := newBO()
	p1.AddChild(child)

	if report.HasParent(child, p2) {
		t.Error("HasParent should return false when p2 is not in parent chain")
	}
}

func TestHasParent_Ancestor_TwoLevels(t *testing.T) {
	// grandparent → parent → child, all as baseChild (which implements both Base and Parent)
	grandparent := newBaseChild("grandparent")
	parent := newBaseChild("parent")
	child := newBaseChild("child")

	grandparent.AddChild(parent)
	parent.AddChild(child)

	if !report.HasParent(child, grandparent) {
		t.Error("HasParent should find grandparent as ancestor of child")
	}
	if !report.HasParent(parent, grandparent) {
		t.Error("HasParent should find grandparent as ancestor of parent")
	}
	if report.HasParent(grandparent, grandparent) {
		t.Error("HasParent should not find self as its own ancestor")
	}
}

// --- AllObjects ---

func TestAllObjects_NoChildren(t *testing.T) {
	b := newBaseChild("root")
	all := report.AllObjects(b)
	if len(all) != 0 {
		t.Errorf("AllObjects on leaf node should return empty, got %d", len(all))
	}
}

func TestAllObjects_DirectChildren(t *testing.T) {
	root := newBaseChild("root")
	c1 := newBaseChild("c1")
	c2 := newBaseChild("c2")
	root.AddChild(c1)
	root.AddChild(c2)

	all := report.AllObjects(root)
	if len(all) != 2 {
		t.Fatalf("AllObjects: expected 2, got %d", len(all))
	}
}

func TestAllObjects_Recursive(t *testing.T) {
	root := newBaseChild("root")
	child := newBaseChild("child")
	grandchild := newBaseChild("grandchild")
	root.AddChild(child)
	child.AddChild(grandchild)

	all := report.AllObjects(root)
	if len(all) != 2 {
		t.Fatalf("AllObjects recursive: expected 2, got %d", len(all))
	}

	names := make(map[string]bool)
	for _, obj := range all {
		names[obj.Name()] = true
	}
	if !names["child"] {
		t.Error("AllObjects should include child")
	}
	if !names["grandchild"] {
		t.Error("AllObjects should include grandchild")
	}
	if names["root"] {
		t.Error("AllObjects should not include root itself")
	}
}

func TestAllObjects_NonParentBase(t *testing.T) {
	// A plain BaseObject (not a Parent) should return empty.
	b := newBO()
	all := report.AllObjects(b)
	if len(all) != 0 {
		t.Errorf("AllObjects on non-Parent Base should return empty, got %d", len(all))
	}
}

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
