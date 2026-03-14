package utils

import (
	"sync"

	"golang.org/x/image/font"
	"golang.org/x/image/font/basicfont"
)

// FontStyle specifies the style of a font.
type FontStyle int

const (
	FontStyleRegular    FontStyle = 0
	FontStyleBold       FontStyle = 1
	FontStyleItalic     FontStyle = 2
	FontStyleBoldItalic FontStyle = 3
	FontStyleUnderline  FontStyle = 4
	FontStyleStrikeout  FontStyle = 8
)

// FontDescriptor identifies a font by name, size and style.
type FontDescriptor struct {
	Family string
	Size   float32 // in points
	Style  FontStyle
}

// FontManager manages font loading and caching.
// It is the Go equivalent of FastReport's FontManager.
type FontManager struct {
	mu    sync.RWMutex
	cache map[FontDescriptor]font.Face
}

// DefaultFontManager is the global font manager.
var DefaultFontManager = NewFontManager()

// NewFontManager creates a new FontManager with an empty cache.
func NewFontManager() *FontManager {
	return &FontManager{
		cache: make(map[FontDescriptor]font.Face),
	}
}

// GetFace returns a font.Face for the given descriptor.
// Falls back to basicfont.Face7x13 if the font cannot be loaded.
func (fm *FontManager) GetFace(desc FontDescriptor) font.Face {
	fm.mu.RLock()
	if face, ok := fm.cache[desc]; ok {
		fm.mu.RUnlock()
		return face
	}
	fm.mu.RUnlock()

	// No cached face; return the default fallback.
	return basicfont.Face7x13
}

// AddFace registers a font face for a descriptor.
func (fm *FontManager) AddFace(desc FontDescriptor, face font.Face) {
	fm.mu.Lock()
	fm.cache[desc] = face
	fm.mu.Unlock()
}

// DefaultFace returns the default font face (basicfont.Face7x13).
func DefaultFace() font.Face {
	return basicfont.Face7x13
}
