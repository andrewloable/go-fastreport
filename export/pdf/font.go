package pdf

import (
	"encoding/binary"
	"fmt"
	"sort"
	"strings"

	"github.com/andrewloable/go-fastreport/export/pdf/core"
	"golang.org/x/image/font/gofont/gobold"
	"golang.org/x/image/font/gofont/gobolditalic"
	"golang.org/x/image/font/gofont/goitalic"
	"golang.org/x/image/font/gofont/gomono"
	"golang.org/x/image/font/gofont/gomonobold"
	"golang.org/x/image/font/gofont/gomonobolditalic"
	"golang.org/x/image/font/gofont/gomonoitalic"
	"golang.org/x/image/font/gofont/goregular"
	"golang.org/x/image/font/sfnt"
	xfont "golang.org/x/image/font"
	"golang.org/x/image/math/fixed"
)

// fontKey identifies a (family, bold, italic) combination.
type fontKey struct {
	family string // "sans" or "mono"
	bold   bool
	italic bool
}

// glyphRune maps a glyph index to its Unicode code point for ToUnicode CMap.
type glyphRune struct {
	gi  sfnt.GlyphIndex
	r   rune
}

// embeddedFont holds all state for a single embedded TTF font variant.
type embeddedFont struct {
	alias      string              // "EF0", "EF1", …
	ttfData    []byte              // raw TTF bytes
	parsed     *sfnt.Font          // parsed font for metrics and glyph lookup
	baseName   string              // PDF base font name e.g. "GoRegular"
	usedGlyphs map[sfnt.GlyphIndex]bool // which glyphs have been referenced
	glyphRunes []glyphRune              // ordered glyph→rune mappings for CMap
	glyphRuneIdx map[sfnt.GlyphIndex]rune  // quick lookup glyph→rune
	// PDF indirect objects – populated by Finalize()
	fontObj    *core.IndirectObject
}

// pdfFontManager manages document-level TrueType font embedding.
// It is created once per document and shared across all pages.
type pdfFontManager struct {
	writer    *Writer
	fonts     map[fontKey]*embeddedFont
	aliasIdx  int
	// ordered list for deterministic output
	order     []fontKey
	// alias → key for fast lookup
	aliasMap  map[string]fontKey
}

// NewPDFFontManager creates a font manager bound to the given PDF writer.
func NewPDFFontManager(w *Writer) *pdfFontManager {
	return &pdfFontManager{
		writer:   w,
		fonts:    make(map[fontKey]*embeddedFont),
		aliasMap: make(map[string]fontKey),
	}
}

// familyKeywordPDF returns "mono" or "sans" from a font family name.
// Mirrors the logic from export/image/fonts.go.
func familyKeywordPDF(name string) string {
	lower := strings.ToLower(name)
	if strings.Contains(lower, "courier") || strings.Contains(lower, "consolas") ||
		strings.Contains(lower, "mono") || strings.Contains(lower, "monaco") ||
		strings.Contains(lower, "lucida console") || strings.Contains(lower, "cascadia") {
		return "mono"
	}
	return "sans"
}

// ttfDataFor returns the raw TTF bytes for a given family + bold + italic.
func ttfDataFor(family string, bold, italic bool) ([]byte, string) {
	switch family {
	case "mono":
		switch {
		case bold && italic:
			return gomonobolditalic.TTF, "GoMonoBoldItalic"
		case bold:
			return gomonobold.TTF, "GoMonoBold"
		case italic:
			return gomonoitalic.TTF, "GoMonoItalic"
		default:
			return gomono.TTF, "GoMono"
		}
	default: // "sans"
		switch {
		case bold && italic:
			return gobolditalic.TTF, "GoBoldItalic"
		case bold:
			return gobold.TTF, "GoBold"
		case italic:
			return goitalic.TTF, "GoItalic"
		default:
			return goregular.TTF, "GoRegular"
		}
	}
}

// RegisterFont registers a TrueType font variant for embedding and returns its
// alias string (e.g. "EF0"). If the same variant is registered twice the
// existing alias is returned.
func (fm *pdfFontManager) RegisterFont(name string, bold, italic bool) string {
	family := familyKeywordPDF(name)
	key := fontKey{family: family, bold: bold, italic: italic}
	if ef, ok := fm.fonts[key]; ok {
		return ef.alias
	}

	ttfData, baseName := ttfDataFor(family, bold, italic)
	parsed, err := sfnt.Parse(ttfData)
	if err != nil {
		// Fallback: return standard font alias if TTF parsing fails.
		return pdfFontAlias(name, bold, italic)
	}

	alias := fmt.Sprintf("EF%d", fm.aliasIdx)
	fm.aliasIdx++

	ef := &embeddedFont{
		alias:        alias,
		ttfData:      ttfData,
		parsed:       parsed,
		baseName:     baseName,
		usedGlyphs:   make(map[sfnt.GlyphIndex]bool),
		glyphRuneIdx: make(map[sfnt.GlyphIndex]rune),
	}
	fm.fonts[key] = ef
	fm.order = append(fm.order, key)
	fm.aliasMap[alias] = key
	return alias
}

// lookupFont returns the embeddedFont for the given alias, or nil if unknown.
func (fm *pdfFontManager) lookupFont(alias string) *embeddedFont {
	key, ok := fm.aliasMap[alias]
	if !ok {
		return nil
	}
	return fm.fonts[key]
}

// EncodeText converts a UTF-8 string to a PDF hex string for Identity-H
// encoding. Each character is looked up in the font's cmap to obtain its
// glyph index; the result is returned as <HHHH…> (big-endian uint16 per glyph).
func (fm *pdfFontManager) EncodeText(alias, text string) string {
	ef := fm.lookupFont(alias)
	if ef == nil || ef.parsed == nil {
		return fmt.Sprintf("(%s)", pdfEscape(text))
	}

	buf := &sfnt.Buffer{}
	var hexBytes []byte
	for _, r := range text {
		gi, err := ef.parsed.GlyphIndex(buf, r)
		if err != nil || gi == 0 {
			gi = 0
		}
		// Track glyph usage for ToUnicode CMap.
		if gi != 0 && !ef.usedGlyphs[gi] {
			ef.usedGlyphs[gi] = true
			ef.glyphRunes = append(ef.glyphRunes, glyphRune{gi: gi, r: r})
			ef.glyphRuneIdx[gi] = r
		}
		var b [2]byte
		binary.BigEndian.PutUint16(b[:], uint16(gi))
		hexBytes = append(hexBytes, b[:]...)
	}

	var sb strings.Builder
	sb.WriteByte('<')
	for _, b := range hexBytes {
		fmt.Fprintf(&sb, "%02X", b)
	}
	sb.WriteByte('>')
	return sb.String()
}

// MeasureText returns the width of text rendered with the named embedded font
// at sizePt points, using actual glyph advance widths from the sfnt font.
func (fm *pdfFontManager) MeasureText(alias, text string, sizePt float64) float64 {
	ef := fm.lookupFont(alias)
	if ef == nil || ef.parsed == nil {
		return pdfEstimateTextWidth(text, sizePt)
	}

	buf := &sfnt.Buffer{}
	unitsPerEm := ef.parsed.UnitsPerEm()
	// Use ppem = unitsPerEm so advance values are in font design units (26.6 fixed).
	ppem := fixed.Int26_6(unitsPerEm) << 6
	var totalWidth float64
	for _, r := range text {
		gi, err := ef.parsed.GlyphIndex(buf, r)
		if err != nil || gi == 0 {
			// Use a rough estimate for missing glyphs.
			totalWidth += sizePt * 0.6
			continue
		}
		advance, err := ef.parsed.GlyphAdvance(buf, gi, ppem, xfont.HintingNone)
		if err != nil {
			totalWidth += sizePt * 0.6
			continue
		}
		// advance is in 26.6 fixed-point at ppem=unitsPerEm, so advance>>6 == design units.
		designUnits := float64(advance >> 6)
		totalWidth += designUnits * sizePt / float64(unitsPerEm)
	}
	return totalWidth
}

// Finalize creates all required PDF objects (FontFile2, FontDescriptor, CIDFont,
// ToUnicode, Type0) for every registered font.  This must be called after all
// content has been written and before writer.Write().
func (fm *pdfFontManager) Finalize() {
	for _, key := range fm.order {
		ef := fm.fonts[key]
		fm.finalizeFont(ef)
	}
}

// finalizeFont builds the PDF object graph for a single embedded font.
func (fm *pdfFontManager) finalizeFont(ef *embeddedFont) {
	// ── FontFile2 ────────────────────────────────────────────────────────────
	ffStream := core.NewStream()
	ffStream.Compressed = false
	ffStream.Dict.Add("Type", core.NewName("EmbeddedFile"))
	ffStream.Dict.Add("Length1", core.NewInt(len(ef.ttfData)))
	ffStream.Data = ef.ttfData
	ffObj := fm.writer.NewObject(ffStream)

	// ── FontDescriptor ───────────────────────────────────────────────────────
	// Extract actual metrics from the parsed font for proper PDF rendering.
	ascent, descent, capHeight := 800, -200, 716
	if m, err := ef.parsed.Metrics(&sfnt.Buffer{}, fixed.Int26_6(ef.parsed.UnitsPerEm())<<6, xfont.HintingNone); err == nil {
		upm := float64(ef.parsed.UnitsPerEm())
		ascent = int(float64(m.Ascent>>6) * 1000.0 / upm)
		descent = -int(float64(m.Descent>>6) * 1000.0 / upm)
		capHeight = int(float64(m.CapHeight>>6) * 1000.0 / upm)
	}
	fdDict := core.NewDictionary()
	fdDict.Add("Type", core.NewName("FontDescriptor"))
	fdDict.Add("FontName", core.NewName(ef.baseName))
	fdDict.Add("Flags", core.NewInt(4))
	fdDict.Add("FontBBox", core.NewArray(
		core.NewInt(-166), core.NewInt(descent),
		core.NewInt(1000), core.NewInt(ascent),
	))
	fdDict.Add("ItalicAngle", core.NewInt(0))
	fdDict.Add("Ascent", core.NewInt(ascent))
	fdDict.Add("Descent", core.NewInt(descent))
	fdDict.Add("CapHeight", core.NewInt(capHeight))
	fdDict.Add("StemV", core.NewInt(80))
	fdDict.Add("FontFile2", core.NewRef(ffObj))
	fdObj := fm.writer.NewObject(fdDict)

	// ── CIDFont ──────────────────────────────────────────────────────────────
	cidSysInfo := core.NewDictionary()
	cidSysInfo.Add("Registry", core.NewString("Adobe"))
	cidSysInfo.Add("Ordering", core.NewString("Identity"))
	cidSysInfo.Add("Supplement", core.NewInt(0))

	cidDict := core.NewDictionary()
	cidDict.Add("Type", core.NewName("Font"))
	cidDict.Add("Subtype", core.NewName("CIDFontType2"))
	cidDict.Add("BaseFont", core.NewName(ef.baseName))
	cidDict.Add("CIDSystemInfo", cidSysInfo)
	cidDict.Add("FontDescriptor", core.NewRef(fdObj))
	cidDict.Add("DW", core.NewInt(1000))

	// Build W (glyph width) array so PDF readers know exact character widths.
	// Format: [CID [w] CID [w] ...] — one entry per used glyph.
	wArr := fm.buildGlyphWidths(ef)
	if wArr != nil {
		cidDict.Add("W", wArr)
	}

	cidObj := fm.writer.NewObject(cidDict)

	// ── ToUnicode CMap ───────────────────────────────────────────────────────
	toUnicodeObj := fm.buildToUnicode(ef)

	// ── Type0 (main font object) ─────────────────────────────────────────────
	descendants := core.NewArray(core.NewRef(cidObj))
	t0Dict := core.NewDictionary()
	t0Dict.Add("Type", core.NewName("Font"))
	t0Dict.Add("Subtype", core.NewName("Type0"))
	t0Dict.Add("BaseFont", core.NewName(ef.baseName+"-Identity-H"))
	t0Dict.Add("Encoding", core.NewName("Identity-H"))
	t0Dict.Add("DescendantFonts", descendants)
	t0Dict.Add("ToUnicode", core.NewRef(toUnicodeObj))
	t0Obj := fm.writer.NewObject(t0Dict)

	ef.fontObj = t0Obj
}

// buildToUnicode creates the ToUnicode CMap stream for the embedded font.
func (fm *pdfFontManager) buildToUnicode(ef *embeddedFont) *core.IndirectObject {
	var sb strings.Builder
	sb.WriteString("/CIDInit /ProcSet findresource begin\n")
	sb.WriteString("12 dict begin\n")
	sb.WriteString("begincmap\n")
	sb.WriteString("/CIDSystemInfo << /Registry (Adobe) /Ordering (UCS) /Supplement 0 >> def\n")
	sb.WriteString("/CMapName /Adobe-Identity-UCS def\n")
	sb.WriteString("/CMapType 2 def\n")
	sb.WriteString("1 begincodespacerange\n")
	sb.WriteString("<0000> <FFFF>\n")
	sb.WriteString("endcodespacerange\n")

	mappings := ef.glyphRunes
	// Write in batches of up to 100 (PDF spec limit per bfchar section).
	batchSize := 100
	for i := 0; i < len(mappings); i += batchSize {
		end := i + batchSize
		if end > len(mappings) {
			end = len(mappings)
		}
		batch := mappings[i:end]
		fmt.Fprintf(&sb, "%d beginbfchar\n", len(batch))
		for _, m := range batch {
			fmt.Fprintf(&sb, "<%04X> <%04X>\n", uint16(m.gi), uint16(m.r))
		}
		sb.WriteString("endbfchar\n")
	}

	sb.WriteString("endcmap\n")
	sb.WriteString("CMapName currentdict /CMap defineresource pop\n")
	sb.WriteString("end\n")
	sb.WriteString("end\n")

	cmapStream := core.NewStream()
	cmapStream.Compressed = true
	cmapStream.Data = []byte(sb.String())
	return fm.writer.NewObject(cmapStream)
}

// buildGlyphWidths builds the PDF W (width) array for a CIDFont.
// Each used glyph gets an entry: [CID [width]] where width is in 1/1000 em units.
// This tells PDF readers the exact advance width of each character so text
// is spaced correctly instead of falling back to the DW default.
func (fm *pdfFontManager) buildGlyphWidths(ef *embeddedFont) *core.Array {
	if ef.parsed == nil || len(ef.usedGlyphs) == 0 {
		return nil
	}

	buf := &sfnt.Buffer{}
	unitsPerEm := float64(ef.parsed.UnitsPerEm())
	ppem := fixed.Int26_6(ef.parsed.UnitsPerEm()) << 6

	// Collect glyph indices sorted for deterministic output.
	type gw struct {
		gi    sfnt.GlyphIndex
		width int // in 1/1000 em units
	}
	var glyphs []gw
	for gi := range ef.usedGlyphs {
		advance, err := ef.parsed.GlyphAdvance(buf, gi, ppem, xfont.HintingNone)
		if err != nil {
			continue
		}
		// advance is in 26.6 fixed-point at ppem=unitsPerEm, so advance>>6 == design units.
		designUnits := float64(advance >> 6)
		// Convert to 1/1000 em units (PDF CIDFont standard).
		w1000 := int(designUnits * 1000.0 / unitsPerEm)
		glyphs = append(glyphs, gw{gi: gi, width: w1000})
	}

	sort.Slice(glyphs, func(i, j int) bool {
		return glyphs[i].gi < glyphs[j].gi
	})

	// Build W array: [CID [width] CID [width] ...]
	wArr := core.NewArray()
	for _, g := range glyphs {
		wArr.Add(core.NewInt(int(g.gi)))
		wArr.Add(core.NewArray(core.NewInt(g.width)))
	}
	return wArr
}

// AddToPage inserts all finalized embedded font references into a page's
// /Font resource dictionary.  Call after Finalize().
func (fm *pdfFontManager) AddToPage(fontDict *core.Dictionary) {
	for _, key := range fm.order {
		ef := fm.fonts[key]
		if ef.fontObj != nil {
			fontDict.Add(ef.alias, core.NewRef(ef.fontObj))
		}
	}
}
