package image

import (
	"os"
	"path/filepath"
	"strings"
	"sync"

	xfont "golang.org/x/image/font"
	"golang.org/x/image/font/basicfont"
	"golang.org/x/image/font/gofont/gobold"
	"golang.org/x/image/font/gofont/gobolditalic"
	"golang.org/x/image/font/gofont/goitalic"
	"golang.org/x/image/font/gofont/gomono"
	"golang.org/x/image/font/gofont/gomonobold"
	"golang.org/x/image/font/gofont/gomonobolditalic"
	"golang.org/x/image/font/gofont/gomonoitalic"
	"golang.org/x/image/font/gofont/goregular"
	"golang.org/x/image/font/opentype"

	"github.com/andrewloable/go-fastreport/style"
)

// fontCacheKey identifies a unique (family, bold, italic, sizePt, dpi) combination.
type fontCacheKey struct {
	family string
	bold   bool
	italic bool
	sizePt float64
	dpi    float64
}

// fontCache holds parsed and sized font faces keyed by fontCacheKey.
var (
	fontCacheMu sync.Mutex
	fontCache   = map[fontCacheKey]xfont.Face{}
)

// goFontData holds the raw TTF bytes for the four variants of each family.
type goFontData struct {
	regular    []byte
	bold       []byte
	italic     []byte
	boldItalic []byte
}

// fontFamilies maps lowercase family keywords to Go font TTF data.
var fontFamilies = map[string]goFontData{
	"sans": {
		regular:    goregular.TTF,
		bold:       gobold.TTF,
		italic:     goitalic.TTF,
		boldItalic: gobolditalic.TTF,
	},
	"mono": {
		regular:    gomono.TTF,
		bold:       gomonobold.TTF,
		italic:     gomonoitalic.TTF,
		boldItalic: gomonobolditalic.TTF,
	},
}

// familyKeyword returns the font family keyword ("sans" or "mono") by
// matching common Windows/CSS font names.
func familyKeyword(name string) string {
	lower := strings.ToLower(name)
	if strings.Contains(lower, "courier") || strings.Contains(lower, "consolas") ||
		strings.Contains(lower, "mono") || strings.Contains(lower, "monaco") ||
		strings.Contains(lower, "lucida console") || strings.Contains(lower, "cascadia") {
		return "mono"
	}
	return "sans"
}

// systemFontDirs lists OS-level font directories to search for TTF/OTF files.
var systemFontDirs = []string{
	// macOS
	"/Library/Fonts",
	"/System/Library/Fonts/Supplemental",
	"/System/Library/Fonts",
	// Linux
	"/usr/share/fonts",
	"/usr/local/share/fonts",
	// Windows (for cross-platform completeness)
	`C:\Windows\Fonts`,
}

// tryLoadSystemFont attempts to locate and load a TTF/OTF file for the given
// font name and style. Returns nil if the font cannot be found.
// Searches systemFontDirs with several filename conventions.
func tryLoadSystemFont(name string, bold, italic bool, sizePt, dpi float64) xfont.Face {
	// Build candidate filenames for this family + style, from most to least specific.
	var candidates []string
	switch {
	case bold && italic:
		candidates = []string{
			name + " Bold Italic.ttf",
			name + "BoldItalic.ttf",
			name + "-BoldItalic.ttf",
			name + "_BoldItalic.ttf",
			name + " Bold Italic.otf",
			name + "-BoldItalic.otf",
			// Fallback: try bold-only, then italic-only, then regular
			name + " Bold.ttf",
			name + " Italic.ttf",
			name + ".ttf",
		}
	case bold:
		candidates = []string{
			name + " Bold.ttf",
			name + "Bold.ttf",
			name + "-Bold.ttf",
			name + "_Bold.ttf",
			name + " Bold.otf",
			name + "-Bold.otf",
			name + ".ttf",
		}
	case italic:
		candidates = []string{
			name + " Italic.ttf",
			name + "Italic.ttf",
			name + "-Italic.ttf",
			name + "_Italic.ttf",
			name + " Italic.otf",
			name + "-Italic.otf",
			name + ".ttf",
		}
	default:
		candidates = []string{
			name + ".ttf",
			name + ".otf",
			name + " Regular.ttf",
			name + "-Regular.ttf",
		}
	}

	home, _ := os.UserHomeDir()
	dirs := append([]string{filepath.Join(home, "Library", "Fonts")}, systemFontDirs...)

	for _, dir := range dirs {
		for _, fname := range candidates {
			path := filepath.Join(dir, fname)
			data, err := os.ReadFile(path)
			if err != nil {
				continue
			}
			parsed, err := opentype.Parse(data)
			if err != nil {
				continue
			}
			face, err := opentype.NewFace(parsed, &opentype.FaceOptions{
				Size: sizePt,
				DPI:  dpi,
			})
			if err != nil {
				continue
			}
			return face
		}
	}
	return nil
}

// selectFace returns a font.Face for the given style properties.
// sizePt is the font size in points; dpi is the rendering resolution.
// First tries to load the named font from the system, then falls back to Go
// built-in fonts (gobold, goitalic, etc.), then to basicfont.Face7x13.
func selectFace(f style.Font, sizePt, dpi float64) xfont.Face {
	bold := f.Style&style.FontStyleBold != 0
	italic := f.Style&style.FontStyleItalic != 0
	family := familyKeyword(f.Name)
	key := fontCacheKey{family: f.Name, bold: bold, italic: italic, sizePt: sizePt, dpi: dpi}

	fontCacheMu.Lock()
	defer fontCacheMu.Unlock()

	if face, ok := fontCache[key]; ok {
		return face
	}

	// Attempt to load the exact named system font first.
	if face := tryLoadSystemFont(f.Name, bold, italic, sizePt, dpi); face != nil {
		fontCache[key] = face
		return face
	}

	// Fall back to Go built-in fonts by family keyword.
	fam := fontFamilies[family]
	var ttf []byte
	switch {
	case bold && italic:
		ttf = fam.boldItalic
	case bold:
		ttf = fam.bold
	case italic:
		ttf = fam.italic
	default:
		ttf = fam.regular
	}

	parsed, err := opentype.Parse(ttf)
	if err != nil {
		fontCache[key] = basicfont.Face7x13
		return basicfont.Face7x13
	}
	face, err := opentype.NewFace(parsed, &opentype.FaceOptions{
		Size: sizePt,
		DPI:  dpi,
	})
	if err != nil {
		fontCache[key] = basicfont.Face7x13
		return basicfont.Face7x13
	}
	fontCache[key] = face
	return face
}
