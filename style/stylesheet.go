package style

import "image/color"

// StyleEntry holds the complete visual properties that a named style
// can override on a report object. It is the Go equivalent of
// FastReport.Style (the named-style definition used in the StyleSheet).
type StyleEntry struct {
	// Name is the style's unique identifier.
	Name string
	// Font overrides the component font when FontChanged is true.
	Font        Font
	FontChanged bool
	// TextColor overrides the text colour when TextColorChanged is true.
	TextColor        color.RGBA
	TextColorChanged bool
	// FillColor overrides the solid-fill background colour when FillColorChanged is true.
	FillColor        color.RGBA
	FillColorChanged bool
	// BorderColor overrides all border-line colours when BorderColorChanged is true.
	BorderColor        color.RGBA
	BorderColorChanged bool
}

// StyleSheet is a named-style registry. It maps style names to StyleEntry
// definitions and applies them to objects that implement Styleable.
// It is the Go equivalent of FastReport's StyleCollection (Report.Styles).
type StyleSheet struct {
	entries map[string]*StyleEntry
	order   []string
}

// NewStyleSheet creates an empty StyleSheet.
func NewStyleSheet() *StyleSheet {
	return &StyleSheet{
		entries: make(map[string]*StyleEntry),
	}
}

// Add registers a StyleEntry. If a style with the same name already exists
// it is replaced.
func (ss *StyleSheet) Add(e *StyleEntry) {
	if _, exists := ss.entries[e.Name]; !exists {
		ss.order = append(ss.order, e.Name)
	}
	ss.entries[e.Name] = e
}

// Find returns the StyleEntry with the given name, or nil if not registered.
func (ss *StyleSheet) Find(name string) *StyleEntry {
	return ss.entries[name]
}

// Len returns the number of registered styles.
func (ss *StyleSheet) Len() int { return len(ss.entries) }

// All returns all registered StyleEntries in registration order.
func (ss *StyleSheet) All() []*StyleEntry {
	result := make([]*StyleEntry, 0, len(ss.order))
	for _, name := range ss.order {
		result = append(result, ss.entries[name])
	}
	return result
}

// Styleable is the interface that report components must implement to receive
// style overrides from a StyleSheet. ReportComponentBase in the report package
// satisfies this interface.
type Styleable interface {
	// StyleName returns the name of the style applied to the component.
	StyleName() string
	// ApplyStyle applies the given StyleEntry's overrides to the component.
	ApplyStyle(entry *StyleEntry)
}

// ApplyToObject looks up obj.StyleName() in the stylesheet and, if found,
// calls obj.ApplyStyle with the matching entry. It is a no-op when the
// component has no style name set or the style is not registered.
func (ss *StyleSheet) ApplyToObject(obj Styleable) {
	name := obj.StyleName()
	if name == "" {
		return
	}
	entry := ss.Find(name)
	if entry == nil {
		return
	}
	obj.ApplyStyle(entry)
}
