package band

// band_internal_coverage_test.go – internal tests for unexported functionality.
// Uses package band (not band_test) to access unexported fields.

import (
	"errors"
	"testing"

	"github.com/andrewloable/go-fastreport/report"
)

// ── mockWriter is a report.Writer that can fail on WriteObject ────────────────

type mockWriter struct {
	failWriteObject bool
	written         map[string]string
}

func newMockWriter() *mockWriter {
	return &mockWriter{written: make(map[string]string)}
}

func (m *mockWriter) WriteStr(name, value string)   { m.written[name] = value }
func (m *mockWriter) WriteInt(name string, v int)    { m.written[name] = "int" }
func (m *mockWriter) WriteBool(name string, v bool)  { m.written[name] = "bool" }
func (m *mockWriter) WriteFloat(name string, v float32) { m.written[name] = "float" }

func (m *mockWriter) WriteObject(obj report.Serializable) error {
	if m.failWriteObject {
		return errors.New("mock WriteObject error")
	}
	return nil
}

func (m *mockWriter) WriteObjectNamed(name string, obj report.Serializable) error {
	if m.failWriteObject {
		return errors.New("mock WriteObjectNamed error")
	}
	return nil
}

// ── BandColumns.ActualWidth with pageWidthFn set ──────────────────────────────

func TestBandColumns_ActualWidth_WithPageWidthFn_MultiColumn(t *testing.T) {
	bc := NewBandColumns()
	bc.Width = 0
	_ = bc.SetCount(4)
	bc.pageWidthFn = func() float32 { return 800 }

	got := bc.ActualWidth()
	want := float32(800) / float32(4) // 200
	if got != want {
		t.Errorf("ActualWidth = %v, want %v", got, want)
	}
}

func TestBandColumns_ActualWidth_WithPageWidthFn_CountLEOne(t *testing.T) {
	bc := NewBandColumns()
	bc.Width = 0
	_ = bc.SetCount(0) // count <= 1 → treated as 1
	bc.pageWidthFn = func() float32 { return 600 }

	got := bc.ActualWidth()
	want := float32(600) // 600 / 1
	if got != want {
		t.Errorf("ActualWidth with count=0 and pageWidthFn = %v, want %v", got, want)
	}
}

func TestBandColumns_ActualWidth_WithPageWidthFn_CountOne(t *testing.T) {
	bc := NewBandColumns()
	bc.Width = 0
	_ = bc.SetCount(1)
	bc.pageWidthFn = func() float32 { return 500 }

	got := bc.ActualWidth()
	want := float32(500) // 500 / 1
	if got != want {
		t.Errorf("ActualWidth with count=1 and pageWidthFn = %v, want %v", got, want)
	}
}

// ── serializeChildren error path ──────────────────────────────────────────────

// minimalSerializable satisfies report.Base (and report.Serializable).
type minimalSerializable struct {
	report.BaseObject
}

func (s *minimalSerializable) Serialize(w report.Writer) error   { return nil }
func (s *minimalSerializable) Deserialize(r report.Reader) error { return nil }

func TestBandBase_serializeChildren_ErrorPropagation(t *testing.T) {
	b := NewBandBase()

	// Add a child object.
	child := &minimalSerializable{BaseObject: *report.NewBaseObject()}
	child.SetName("child1")
	b.AddChild(child)

	// Use a writer that fails on WriteObject.
	w := newMockWriter()
	w.failWriteObject = true

	err := b.serializeChildren(w)
	if err == nil {
		t.Error("serializeChildren should propagate WriteObject error")
	}
}

func TestBandBase_serializeChildren_NoError(t *testing.T) {
	b := NewBandBase()

	child := &minimalSerializable{BaseObject: *report.NewBaseObject()}
	child.SetName("c1")
	b.AddChild(child)

	w := newMockWriter()
	w.failWriteObject = false

	err := b.serializeChildren(w)
	if err != nil {
		t.Errorf("serializeChildren should not error, got: %v", err)
	}
}

// ── BandBase.Serialize error path (via serializeAttrs or serializeChildren) ───

func TestBandBase_Serialize_ErrorFromChildren(t *testing.T) {
	b := NewBandBase()
	child := &minimalSerializable{BaseObject: *report.NewBaseObject()}
	b.AddChild(child)

	w := newMockWriter()
	w.failWriteObject = true

	err := b.Serialize(w)
	if err == nil {
		t.Error("BandBase.Serialize should propagate error from serializeChildren")
	}
}

// ── HeaderFooterBandBase.Serialize with children (serializeChildren branch) ───

func TestHeaderFooterBandBase_Serialize_WithChildren(t *testing.T) {
	h := NewHeaderFooterBandBase()
	h.SetKeepWithData(true)
	h.SetRepeatOnEveryPage(true)

	child := &minimalSerializable{BaseObject: *report.NewBaseObject()}
	h.AddChild(child)

	w := newMockWriter()
	w.failWriteObject = false

	err := h.Serialize(w)
	if err != nil {
		t.Errorf("HeaderFooterBandBase.Serialize with children should not error: %v", err)
	}
}

func TestHeaderFooterBandBase_Serialize_WriteObjectError(t *testing.T) {
	h := NewHeaderFooterBandBase()
	child := &minimalSerializable{BaseObject: *report.NewBaseObject()}
	h.AddChild(child)

	w := newMockWriter()
	w.failWriteObject = true

	err := h.Serialize(w)
	if err == nil {
		t.Error("HeaderFooterBandBase.Serialize should propagate WriteObject error from children")
	}
}

// ── GroupHeaderBand.Serialize with children ───────────────────────────────────

func TestGroupHeaderBand_Serialize_WithChildren(t *testing.T) {
	g := NewGroupHeaderBand()
	g.SetCondition("[Name]")

	child := &minimalSerializable{BaseObject: *report.NewBaseObject()}
	g.AddChild(child)

	w := newMockWriter()
	w.failWriteObject = false

	err := g.Serialize(w)
	if err != nil {
		t.Errorf("GroupHeaderBand.Serialize with children should not error: %v", err)
	}
}

func TestGroupHeaderBand_Serialize_WriteObjectError(t *testing.T) {
	g := NewGroupHeaderBand()
	child := &minimalSerializable{BaseObject: *report.NewBaseObject()}
	g.AddChild(child)

	w := newMockWriter()
	w.failWriteObject = true

	err := g.Serialize(w)
	if err == nil {
		t.Error("GroupHeaderBand.Serialize should propagate WriteObject error")
	}
}

// ── DataBand.Serialize with children ─────────────────────────────────────────

func TestDataBand_Serialize_WriteObjectError(t *testing.T) {
	d := NewDataBand()
	child := &minimalSerializable{BaseObject: *report.NewBaseObject()}
	d.AddChild(child)

	w := newMockWriter()
	w.failWriteObject = true

	err := d.Serialize(w)
	if err == nil {
		t.Error("DataBand.Serialize should propagate WriteObject error")
	}
}

// ── ChildBand.Serialize error from BandBase.Serialize ────────────────────────

func TestChildBand_Serialize_WriteObjectError(t *testing.T) {
	c := NewChildBand()
	child := &minimalSerializable{BaseObject: *report.NewBaseObject()}
	c.AddChild(child)

	w := newMockWriter()
	w.failWriteObject = true

	err := c.Serialize(w)
	if err == nil {
		t.Error("ChildBand.Serialize should propagate WriteObject error")
	}
}

// ── BandColumns with pageWidthFn: Positions test ─────────────────────────────

func TestBandColumns_Positions_WithPageWidthFn(t *testing.T) {
	bc := NewBandColumns()
	_ = bc.SetCount(3)
	bc.Width = 0
	bc.pageWidthFn = func() float32 { return 300 }

	pos := bc.Positions()
	if len(pos) != 3 {
		t.Fatalf("Positions len = %d, want 3", len(pos))
	}
	// Each column is 100px wide (300/3).
	if pos[0] != 0 {
		t.Errorf("pos[0] = %v, want 0", pos[0])
	}
	if pos[1] != 100 {
		t.Errorf("pos[1] = %v, want 100", pos[1])
	}
	if pos[2] != 200 {
		t.Errorf("pos[2] = %v, want 200", pos[2])
	}
}
