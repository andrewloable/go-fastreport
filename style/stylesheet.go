package style

import (
	"image/color"
	"strings"
)

// StyleEntry holds the complete visual properties that a named style
// can override on a report object. It is the Go equivalent of
// FastReport.StyleBase / FastReport.Style.
type StyleEntry struct {
	// Name is the style's unique identifier.
	Name string

	// ApplyBorder controls whether the Border is applied to the object.
	// Defaults to true.
	ApplyBorder bool
	// Border holds the border overrides to apply when ApplyBorder is true.
	Border Border

	// ApplyFill controls whether the Fill is applied to the object.
	// Defaults to true.
	ApplyFill bool
	// Fill is the fill override when ApplyFill is true.
	// When non-nil it takes precedence over the legacy FillColor field.
	// It is the Go equivalent of FastReport.StyleBase.Fill (FillBase interface).
	Fill Fill
	// FillColor is the solid fill colour override when ApplyFill is true and
	// Fill is nil. Kept for backward compatibility with existing serialisation code.
	FillColor color.RGBA

	// ApplyTextFill controls whether the text fill is applied to the object.
	// Defaults to true.
	ApplyTextFill bool
	// TextFill is the text fill override when ApplyTextFill is true.
	// When non-nil it takes precedence over the legacy TextColor field.
	// It is the Go equivalent of FastReport.StyleBase.TextFill (FillBase interface).
	TextFill Fill
	// TextColor is the text fill colour override when ApplyTextFill is true and
	// TextFill is nil. Kept for backward compatibility with existing serialisation code.
	TextColor color.RGBA

	// ApplyFont controls whether the Font is applied to the object.
	// Defaults to true.
	ApplyFont bool
	// Font overrides the component font when ApplyFont is true.
	Font Font

	// Legacy "Changed" fields kept for backward compatibility with existing code.
	// They map to the corresponding Apply* flags.
	FontChanged        bool
	TextColorChanged   bool
	FillColorChanged   bool
	BorderColorChanged bool
	// BorderColor overrides all border-line colours (legacy; prefer Border).
	BorderColor color.RGBA
}

// Assign copies all fields from src into e, performing deep copies of
// Border, Fill, and TextFill. It is the Go equivalent of
// FastReport.StyleBase.Assign(StyleBase source).
func (e *StyleEntry) Assign(src *StyleEntry) {
	if src == nil {
		return
	}
	e.Name = src.Name
	e.ApplyBorder = src.ApplyBorder
	e.Border = src.Border // Border is a value type with a pointer slice; use Clone below
	if src.Border.Lines[0] != nil {
		cloned := *NewBorder()
		for i := range src.Border.Lines {
			if src.Border.Lines[i] != nil {
				*cloned.Lines[i] = *src.Border.Lines[i]
			}
		}
		cloned.VisibleLines = src.Border.VisibleLines
		cloned.Shadow = src.Border.Shadow
		e.Border = cloned
	}
	e.ApplyFill = src.ApplyFill
	if src.Fill != nil {
		e.Fill = src.Fill.Clone()
	} else {
		e.Fill = nil
	}
	e.FillColor = src.FillColor
	e.ApplyTextFill = src.ApplyTextFill
	if src.TextFill != nil {
		e.TextFill = src.TextFill.Clone()
	} else {
		e.TextFill = nil
	}
	e.TextColor = src.TextColor
	e.ApplyFont = src.ApplyFont
	e.Font = src.Font
	e.FontChanged = src.FontChanged
	e.TextColorChanged = src.TextColorChanged
	e.FillColorChanged = src.FillColorChanged
	e.BorderColorChanged = src.BorderColorChanged
	e.BorderColor = src.BorderColor
}

// Clone returns a deep copy of e. It is the Go equivalent of
// FastReport.Style.Clone().
func (e *StyleEntry) Clone() *StyleEntry {
	result := &StyleEntry{}
	result.Assign(e)
	return result
}

// EffectiveFill returns the fill to use when applying this style entry.
// If the Fill interface field is set it is returned; otherwise a SolidFill
// wrapping FillColor is returned. Returns nil when ApplyFill is false.
// This is a helper for ReportComponentBase.ApplyStyle.
func (e *StyleEntry) EffectiveFill() Fill {
	if !e.ApplyFill && !e.FillColorChanged {
		return nil
	}
	if e.Fill != nil {
		return e.Fill
	}
	return &SolidFill{Color: e.FillColor}
}

// EffectiveTextFill returns the text fill to use when applying this style
// entry. If the TextFill interface field is set it is returned; otherwise a
// SolidFill wrapping TextColor is returned. Returns nil when ApplyTextFill
// is false.
func (e *StyleEntry) EffectiveTextFill() Fill {
	if !e.ApplyTextFill && !e.TextColorChanged {
		return nil
	}
	if e.TextFill != nil {
		return e.TextFill
	}
	return &SolidFill{Color: e.TextColor}
}

// StyleSheet is a named-style registry. It maps style names to StyleEntry
// definitions and applies them to objects that implement Styleable.
// It is the Go equivalent of FastReport's StyleCollection (Report.Styles).
//
// Name lookups are case-insensitive to match C# StyleCollection.IndexOf(string)
// which uses String.Compare(s.Name, value, ignoreCase: true).
type StyleSheet struct {
	// entries is keyed by the lower-cased name for case-insensitive lookup.
	entries map[string]*StyleEntry
	// order stores the original (non-lowercased) registration names so that
	// All() returns entries in insertion order with their original names.
	order []string
}

// NewStyleSheet creates an empty StyleSheet.
func NewStyleSheet() *StyleSheet {
	return &StyleSheet{
		entries: make(map[string]*StyleEntry),
	}
}

// Add registers a StyleEntry. If a style with the same name (case-insensitive)
// already exists it is replaced in-place, preserving insertion order.
func (ss *StyleSheet) Add(e *StyleEntry) {
	key := strings.ToLower(e.Name)
	if _, exists := ss.entries[key]; !exists {
		ss.order = append(ss.order, e.Name)
	}
	ss.entries[key] = e
}

// Find returns the StyleEntry with the given name, or nil if not registered.
// The lookup is case-insensitive, matching C# StyleCollection.IndexOf(string)
// behaviour (String.Compare with ignoreCase=true).
func (ss *StyleSheet) Find(name string) *StyleEntry {
	return ss.entries[strings.ToLower(name)]
}

// Len returns the number of registered styles.
func (ss *StyleSheet) Len() int { return len(ss.entries) }

// All returns all registered StyleEntries in registration order.
func (ss *StyleSheet) All() []*StyleEntry {
	result := make([]*StyleEntry, 0, len(ss.order))
	for _, name := range ss.order {
		result = append(result, ss.entries[strings.ToLower(name)])
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
