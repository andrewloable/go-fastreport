package image

import (
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

// selectFace returns a font.Face for the given style properties.
// sizePt is the font size in points; dpi is the rendering resolution.
// Falls back to basicfont.Face7x13 if loading fails.
func selectFace(f style.Font, sizePt, dpi float64) xfont.Face {
	bold := f.Style&style.FontStyleBold != 0
	italic := f.Style&style.FontStyleItalic != 0
	family := familyKeyword(f.Name)
	key := fontCacheKey{family: family, bold: bold, italic: italic, sizePt: sizePt, dpi: dpi}

	fontCacheMu.Lock()
	defer fontCacheMu.Unlock()

	if face, ok := fontCache[key]; ok {
		return face
	}

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
