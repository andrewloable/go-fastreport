package style_test

import (
	"image/color"
	"testing"

	"github.com/andrewloable/go-fastreport/style"
)

// stubStyleable is a minimal Styleable implementation for testing.
type stubStyleable struct {
	name         string
	appliedEntry *style.StyleEntry
}

func (s *stubStyleable) StyleName() string { return s.name }
func (s *stubStyleable) ApplyStyle(e *style.StyleEntry) {
	s.appliedEntry = e
}

// ── Construction ──────────────────────────────────────────────────────────────

func TestNewStyleSheet_Empty(t *testing.T) {
	ss := style.NewStyleSheet()
	if ss == nil {
		t.Fatal("NewStyleSheet returned nil")
	}
	if ss.Len() != 0 {
		t.Errorf("expected 0 entries, got %d", ss.Len())
	}
}

// ── Add ───────────────────────────────────────────────────────────────────────

func TestStyleSheet_Add(t *testing.T) {
	ss := style.NewStyleSheet()
	e := &style.StyleEntry{Name: "Red"}
	ss.Add(e)

	if ss.Len() != 1 {
		t.Errorf("expected 1 entry, got %d", ss.Len())
	}
}

func TestStyleSheet_Add_Replace(t *testing.T) {
	ss := style.NewStyleSheet()
	e1 := &style.StyleEntry{Name: "S", FontChanged: false}
	e2 := &style.StyleEntry{Name: "S", FontChanged: true}
	ss.Add(e1)
	ss.Add(e2) // replaces e1

	if ss.Len() != 1 {
		t.Errorf("expected 1 entry after replace, got %d", ss.Len())
	}
	found := ss.Find("S")
	if !found.FontChanged {
		t.Error("replacement entry should have FontChanged=true")
	}
}

// ── Find ──────────────────────────────────────────────────────────────────────

func TestStyleSheet_Find(t *testing.T) {
	ss := style.NewStyleSheet()
	ss.Add(&style.StyleEntry{Name: "Bold"})

	if ss.Find("Bold") == nil {
		t.Error("Find should return the registered entry")
	}
	if ss.Find("Unknown") != nil {
		t.Error("Find for unknown name should return nil")
	}
}

// C# StyleCollection.IndexOf(string) uses String.Compare with ignoreCase=true.
// Verify that Go Find() matches this case-insensitive behaviour.
func TestStyleSheet_Find_CaseInsensitive(t *testing.T) {
	ss := style.NewStyleSheet()
	ss.Add(&style.StyleEntry{Name: "HeaderStyle"})

	variants := []string{"HeaderStyle", "headerstyle", "HEADERSTYLE", "headerSTYLE"}
	for _, v := range variants {
		if ss.Find(v) == nil {
			t.Errorf("Find(%q) should find entry registered as HeaderStyle", v)
		}
	}
}

func TestStyleSheet_Add_Replace_DifferentCase_PreservesOrder(t *testing.T) {
	ss := style.NewStyleSheet()
	ss.Add(&style.StyleEntry{Name: "Foo", FontChanged: false})
	ss.Add(&style.StyleEntry{Name: "Bar", FontChanged: false})
	// Replace "foo" (same key, different casing) — should not grow the count.
	ss.Add(&style.StyleEntry{Name: "FOO", FontChanged: true})

	if ss.Len() != 2 {
		t.Errorf("Len = %d, want 2 after case-insensitive replace", ss.Len())
	}
	found := ss.Find("foo")
	if found == nil {
		t.Fatal("Find(\"foo\") should find the replaced entry")
	}
	if !found.FontChanged {
		t.Error("replacement entry should have FontChanged=true")
	}
	// All() must return both entries in insertion order.
	all := ss.All()
	if len(all) != 2 {
		t.Fatalf("All() len = %d, want 2", len(all))
	}
}

// ── All / order ───────────────────────────────────────────────────────────────

func TestStyleSheet_All_Order(t *testing.T) {
	ss := style.NewStyleSheet()
	names := []string{"Alpha", "Beta", "Gamma"}
	for _, n := range names {
		ss.Add(&style.StyleEntry{Name: n})
	}

	all := ss.All()
	if len(all) != 3 {
		t.Fatalf("All() len = %d, want 3", len(all))
	}
	for i, name := range names {
		if all[i].Name != name {
			t.Errorf("All()[%d].Name = %q, want %q", i, all[i].Name, name)
		}
	}
}

// ── ApplyToObject ─────────────────────────────────────────────────────────────

func TestStyleSheet_ApplyToObject_Applies(t *testing.T) {
	ss := style.NewStyleSheet()
	entry := &style.StyleEntry{
		Name:            "Header",
		FillColorChanged: true,
		FillColor:       color.RGBA{R: 200, G: 0, B: 0, A: 255},
	}
	ss.Add(entry)

	obj := &stubStyleable{name: "Header"}
	ss.ApplyToObject(obj)

	if obj.appliedEntry == nil {
		t.Fatal("ApplyToObject should have called ApplyStyle")
	}
	if obj.appliedEntry.Name != "Header" {
		t.Errorf("applied entry name = %q, want Header", obj.appliedEntry.Name)
	}
}

func TestStyleSheet_ApplyToObject_NoStyleName_NoOp(t *testing.T) {
	ss := style.NewStyleSheet()
	ss.Add(&style.StyleEntry{Name: "S"})

	obj := &stubStyleable{name: ""} // no style name
	ss.ApplyToObject(obj)

	if obj.appliedEntry != nil {
		t.Error("should not apply style when component has no style name")
	}
}

func TestStyleSheet_ApplyToObject_UnknownStyle_NoOp(t *testing.T) {
	ss := style.NewStyleSheet()
	obj := &stubStyleable{name: "NoSuchStyle"}
	ss.ApplyToObject(obj)

	if obj.appliedEntry != nil {
		t.Error("should not apply style when style name is not registered")
	}
}

// ── StyleEntry fields ─────────────────────────────────────────────────────────

func TestStyleEntry_Fields(t *testing.T) {
	e := &style.StyleEntry{
		Name:               "Test",
		FontChanged:        true,
		Font:               style.Font{Name: "Arial", Size: 12},
		TextColorChanged:   true,
		TextColor:          color.RGBA{R: 0, G: 0, B: 255, A: 255},
		FillColorChanged:   true,
		FillColor:          color.RGBA{R: 255, G: 255, B: 0, A: 255},
		BorderColorChanged: true,
		BorderColor:        color.RGBA{R: 0, G: 0, B: 0, A: 255},
	}

	if e.Name != "Test" {
		t.Errorf("Name = %q, want Test", e.Name)
	}
	if !e.FontChanged || e.Font.Name != "Arial" {
		t.Error("Font not set correctly")
	}
	if !e.TextColorChanged || e.TextColor.B != 255 {
		t.Error("TextColor not set correctly")
	}
}
