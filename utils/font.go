package utils

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"

	"golang.org/x/image/font"
	"golang.org/x/image/font/basicfont"
	"golang.org/x/image/font/opentype"
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
	mu       sync.RWMutex
	cache    map[FontDescriptor]font.Face
	fontData map[string]*opentype.Font // family+bold key → parsed font
}

// DefaultFontManager is the global font manager.
var DefaultFontManager = NewFontManager()

// NewFontManager creates a new FontManager with an empty cache.
func NewFontManager() *FontManager {
	return &FontManager{
		cache:    make(map[FontDescriptor]font.Face),
		fontData: make(map[string]*opentype.Font),
	}
}

// GetFace returns a font.Face for the given descriptor.
// It attempts to load the actual TrueType font from the system to match
// C# GDI+ MeasureString metrics. Falls back to basicfont.Face7x13 if
// the font cannot be loaded.
func (fm *FontManager) GetFace(desc FontDescriptor) font.Face {
	fm.mu.RLock()
	if face, ok := fm.cache[desc]; ok {
		fm.mu.RUnlock()
		return face
	}
	fm.mu.RUnlock()

	// Try to load the actual font.
	face := fm.loadSystemFont(desc)
	if face != nil {
		fm.mu.Lock()
		fm.cache[desc] = face
		fm.mu.Unlock()
		return face
	}

	return basicfont.Face7x13
}

// loadSystemFont attempts to find and load a TrueType font from the OS.
func (fm *FontManager) loadSystemFont(desc FontDescriptor) font.Face {
	isBold := desc.Style&FontStyle(FontStyleBold) != 0
	isItalic := desc.Style&FontStyle(FontStyleItalic) != 0

	// Build a cache key for the parsed font data (family + style).
	fontKey := strings.ToLower(desc.Family)
	if isBold && isItalic {
		fontKey += "-bolditalic"
	} else if isBold {
		fontKey += "-bold"
	} else if isItalic {
		fontKey += "-italic"
	}

	fm.mu.RLock()
	otFont, exists := fm.fontData[fontKey]
	fm.mu.RUnlock()

	if !exists {
		// Try to find the font file on the system.
		data := findFontFile(desc.Family, isBold, isItalic)
		if data == nil {
			// Cache nil to avoid repeated lookups.
			fm.mu.Lock()
			fm.fontData[fontKey] = nil
			fm.mu.Unlock()
			return nil
		}
		var err error
		otFont, err = opentype.Parse(data)
		if err != nil {
			fm.mu.Lock()
			fm.fontData[fontKey] = nil
			fm.mu.Unlock()
			return nil
		}
		fm.mu.Lock()
		fm.fontData[fontKey] = otFont
		fm.mu.Unlock()
	}

	if otFont == nil {
		return nil
	}

	// Create a face at the requested size.
	// Use 96 DPI to match C# GDI+ default (screen DPI).
	face, err := opentype.NewFace(otFont, &opentype.FaceOptions{
		Size:    float64(desc.Size),
		DPI:     96,
		Hinting: font.HintingFull,
	})
	if err != nil {
		return nil
	}
	return face
}

// findFontFile searches system font directories for the named font.
func findFontFile(family string, bold, italic bool) []byte {
	candidates := buildFontCandidates(family, bold, italic)
	dirs := systemFontDirs()

	for _, dir := range dirs {
		for _, candidate := range candidates {
			path := filepath.Join(dir, candidate)
			data, err := os.ReadFile(path)
			if err == nil {
				return data
			}
		}
	}
	return nil
}

// buildFontCandidates returns possible font file names for a family+style.
// Covers common naming conventions across Windows, macOS, and Linux.
func buildFontCandidates(family string, bold, italic bool) []string {
	name := strings.ReplaceAll(family, " ", "")
	var candidates []string

	// Style-specific variants first (most specific).
	switch {
	case bold && italic:
		candidates = append(candidates,
			family+" Bold Italic.ttf",
			name+"bi.ttf",
			name+"-BoldItalic.ttf",
			family+"BoldItalic.ttf",
		)
	case bold:
		candidates = append(candidates,
			family+" Bold.ttf",
			name+"bd.ttf",
			name+"b.ttf",
			name+"-Bold.ttf",
		)
	case italic:
		candidates = append(candidates,
			family+" Italic.ttf",
			name+"i.ttf",
			name+"-Italic.ttf",
			family+"Italic.ttf",
		)
	}

	// Fallback to the base font file.
	candidates = append(candidates,
		family+".ttf",
		name+".ttf",
		family+".ttc",
		name+".ttc",
	)
	return candidates
}

// systemFontDirs returns OS-specific font directories.
func systemFontDirs() []string {
	switch runtime.GOOS {
	case "darwin":
		return []string{
			"/System/Library/Fonts/Supplemental",
			"/System/Library/Fonts",
			"/Library/Fonts",
			filepath.Join(os.Getenv("HOME"), "Library/Fonts"),
		}
	case "linux":
		return []string{
			"/usr/share/fonts/truetype",
			"/usr/share/fonts/truetype/msttcorefonts",
			"/usr/share/fonts/truetype/dejavu",
			"/usr/share/fonts",
			"/usr/local/share/fonts",
			filepath.Join(os.Getenv("HOME"), ".fonts"),
			filepath.Join(os.Getenv("HOME"), ".local/share/fonts"),
		}
	case "windows":
		windir := os.Getenv("WINDIR")
		if windir == "" {
			windir = `C:\Windows`
		}
		return []string{
			filepath.Join(windir, "Fonts"),
		}
	default:
		return nil
	}
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
