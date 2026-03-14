package utils_test

import (
	"testing"

	"github.com/andrewloable/go-fastreport/utils"
)

// mockObj implements ObjectNamer for testing.
type mockObj struct {
	baseName string
	name     string
}

func (m *mockObj) BaseName() string    { return m.baseName }
func (m *mockObj) Name() string        { return m.name }
func (m *mockObj) SetName(n string)    { m.name = n }

func newMock(baseName, name string) *mockObj {
	return &mockObj{baseName: baseName, name: name}
}

func TestNewFastNameCreatorEmpty(t *testing.T) {
	nc := utils.NewFastNameCreator(nil)
	obj := newMock("Text", "")
	nc.CreateUniqueName(obj)
	if obj.Name() != "Text1" {
		t.Errorf("expected Text1, got %q", obj.Name())
	}
}

func TestCreateUniqueNameSequential(t *testing.T) {
	nc := utils.NewFastNameCreator(nil)
	obj1 := newMock("Band", "")
	obj2 := newMock("Band", "")
	obj3 := newMock("Band", "")
	nc.CreateUniqueName(obj1)
	nc.CreateUniqueName(obj2)
	nc.CreateUniqueName(obj3)
	if obj1.Name() != "Band1" {
		t.Errorf("expected Band1, got %q", obj1.Name())
	}
	if obj2.Name() != "Band2" {
		t.Errorf("expected Band2, got %q", obj2.Name())
	}
	if obj3.Name() != "Band3" {
		t.Errorf("expected Band3, got %q", obj3.Name())
	}
}

func TestNewFastNameCreatorWithExisting(t *testing.T) {
	existing := []utils.ObjectNamer{
		newMock("Text", "Text1"),
		newMock("Text", "Text2"),
		newMock("Band", "Band5"),
	}
	nc := utils.NewFastNameCreator(existing)

	textObj := newMock("Text", "")
	nc.CreateUniqueName(textObj)
	if textObj.Name() != "Text3" {
		t.Errorf("expected Text3, got %q", textObj.Name())
	}

	bandObj := newMock("Band", "")
	nc.CreateUniqueName(bandObj)
	if bandObj.Name() != "Band6" {
		t.Errorf("expected Band6, got %q", bandObj.Name())
	}
}

func TestNewFastNameCreatorEmptyNames(t *testing.T) {
	existing := []utils.ObjectNamer{
		newMock("Text", ""), // empty name — should be ignored
	}
	nc := utils.NewFastNameCreator(existing)
	obj := newMock("Text", "")
	nc.CreateUniqueName(obj)
	if obj.Name() != "Text1" {
		t.Errorf("expected Text1, got %q", obj.Name())
	}
}

func TestNewFastNameCreatorNoNumericSuffix(t *testing.T) {
	existing := []utils.ObjectNamer{
		newMock("Text", "TextObject"), // no numeric suffix
	}
	nc := utils.NewFastNameCreator(existing)
	obj := newMock("Text", "")
	nc.CreateUniqueName(obj)
	// Should still start from 1 since TextObject base="TextObject" not "Text"
	if obj.Name() != "Text1" {
		t.Errorf("expected Text1, got %q", obj.Name())
	}
}

func TestDifferentBaseNames(t *testing.T) {
	nc := utils.NewFastNameCreator(nil)
	a := newMock("TextObject", "")
	b := newMock("PictureObject", "")
	nc.CreateUniqueName(a)
	nc.CreateUniqueName(b)
	if a.Name() != "TextObject1" {
		t.Errorf("expected TextObject1, got %q", a.Name())
	}
	if b.Name() != "PictureObject1" {
		t.Errorf("expected PictureObject1, got %q", b.Name())
	}
}

func TestSplitBaseNameViaCreator(t *testing.T) {
	// Test that names like "Text10" correctly parse base="Text", num=10
	existing := []utils.ObjectNamer{
		newMock("Text", "Text10"),
	}
	nc := utils.NewFastNameCreator(existing)
	obj := newMock("Text", "")
	nc.CreateUniqueName(obj)
	if obj.Name() != "Text11" {
		t.Errorf("expected Text11, got %q", obj.Name())
	}
}
