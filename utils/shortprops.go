package utils

// shortprops.go provides a property-name abbreviation map used for compact
// serialisation of prepared pages. This mirrors FastReport.Utils.ShortProperties,
// which maps full property names to short codes to reduce FPX file size.
//
// Usage:
//
//	code, ok := ShortPropCode("FillColor")    // → "fc", true
//	name, ok := ShortPropName("fc")           // → "FillColor", true

var shortToFull = map[string]string{
	// Geometry
	"l":  "Left",
	"t":  "Top",
	"w":  "Width",
	"h":  "Height",
	// Text
	"tx": "Text",
	"ha": "HorzAlign",
	"va": "VertAlign",
	"ww": "WordWrap",
	// Colour / fill
	"fc": "FillColor",
	"tc": "TextColor",
	"bc": "BackColor",
	// Font
	"fn": "Font.Name",
	"fs": "Font.Size",
	"fb": "Font.Bold",
	"fi": "Font.Italic",
	"fu": "Font.Underline",
	// Border
	"bw": "Border.Width",
	"bl": "Border.Lines",
	// Blob / picture
	"bi": "BlobIdx",
	// Misc
	"nm": "Name",
	"vi": "Visible",
	"ck": "Checked",
}

// fullToShort is the inverse of shortToFull, built at init time.
var fullToShort map[string]string

func init() {
	fullToShort = make(map[string]string, len(shortToFull))
	for k, v := range shortToFull {
		fullToShort[v] = k
	}
}

// ShortPropCode returns the short code for a full property name.
// Returns ("", false) if no abbreviation is registered.
func ShortPropCode(fullName string) (code string, ok bool) {
	code, ok = fullToShort[fullName]
	return
}

// ShortPropName returns the full property name for a short code.
// Returns ("", false) if the code is not recognised.
func ShortPropName(code string) (name string, ok bool) {
	name, ok = shortToFull[code]
	return
}

// ExpandPropName returns the full property name for s, or s unchanged if
// s is not a registered short code.
func ExpandPropName(s string) string {
	if full, ok := shortToFull[s]; ok {
		return full
	}
	return s
}

// AbbrevPropName returns the short code for s, or s unchanged if no
// abbreviation is registered.
func AbbrevPropName(s string) string {
	if code, ok := fullToShort[s]; ok {
		return code
	}
	return s
}
