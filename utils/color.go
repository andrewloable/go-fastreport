package utils

import (
	"fmt"
	"image/color"
	"strconv"
	"strings"
)

// Predefined colors for common use.
var (
	// ColorTransparent matches C# Color.Transparent = Color.FromArgb(0, 255, 255, 255):
	// white with alpha=0. The RGB values are 255 to match the .NET definition.
	ColorTransparent = color.RGBA{R: 255, G: 255, B: 255, A: 0}
	// ColorBlack is fully opaque black.
	ColorBlack = color.RGBA{R: 0, G: 0, B: 0, A: 255}
	// ColorWhite is fully opaque white.
	ColorWhite = color.RGBA{R: 255, G: 255, B: 255, A: 255}
	// ColorRed is fully opaque red.
	ColorRed = color.RGBA{R: 255, G: 0, B: 0, A: 255}
	// ColorGreen is fully opaque green (HTML green, not lime).
	ColorGreen = color.RGBA{R: 0, G: 128, B: 0, A: 255}
	// ColorBlue is fully opaque blue.
	ColorBlue = color.RGBA{R: 0, G: 0, B: 255, A: 255}
)

// namedColors maps .NET/CSS color names (case-insensitive) to color.RGBA.
// These are the standard System.Drawing.Color named colors used in FRX files.
var namedColors = map[string]color.RGBA{
	// C# Color.Transparent = Color.FromArgb(0, 255, 255, 255): white with alpha=0.
	"transparent":          {255, 255, 255, 0},
	"aliceblue":            {240, 248, 255, 255},
	"antiquewhite":         {250, 235, 215, 255},
	"aqua":                 {0, 255, 255, 255},
	"aquamarine":           {127, 255, 212, 255},
	"azure":                {240, 255, 255, 255},
	"beige":                {245, 245, 220, 255},
	"bisque":               {255, 228, 196, 255},
	"black":                {0, 0, 0, 255},
	"blanchedalmond":       {255, 235, 205, 255},
	"blue":                 {0, 0, 255, 255},
	"blueviolet":           {138, 43, 226, 255},
	"brown":                {165, 42, 42, 255},
	"burlywood":            {222, 184, 135, 255},
	"cadetblue":            {95, 158, 160, 255},
	"chartreuse":           {127, 255, 0, 255},
	"chocolate":            {210, 105, 30, 255},
	"coral":                {255, 127, 80, 255},
	"cornflowerblue":       {100, 149, 237, 255},
	"cornsilk":             {255, 248, 220, 255},
	"crimson":              {220, 20, 60, 255},
	"cyan":                 {0, 255, 255, 255},
	"darkblue":             {0, 0, 139, 255},
	"darkcyan":             {0, 139, 139, 255},
	"darkgoldenrod":        {184, 134, 11, 255},
	"darkgray":             {169, 169, 169, 255},
	"darkgreen":            {0, 100, 0, 255},
	"darkkhaki":            {189, 183, 107, 255},
	"darkmagenta":          {139, 0, 139, 255},
	"darkolivegreen":       {85, 107, 47, 255},
	"darkorange":           {255, 140, 0, 255},
	"darkorchid":           {153, 50, 204, 255},
	"darkred":              {139, 0, 0, 255},
	"darksalmon":           {233, 150, 122, 255},
	"darkseagreen":         {143, 188, 143, 255},
	"darkslateblue":        {72, 61, 139, 255},
	"darkslategray":        {47, 79, 79, 255},
	"darkturquoise":        {0, 206, 209, 255},
	"darkviolet":           {148, 0, 211, 255},
	"deeppink":             {255, 20, 147, 255},
	"deepskyblue":          {0, 191, 255, 255},
	"dimgray":              {105, 105, 105, 255},
	"dodgerblue":           {30, 144, 255, 255},
	"firebrick":            {178, 34, 34, 255},
	"floralwhite":          {255, 250, 240, 255},
	"forestgreen":          {34, 139, 34, 255},
	"fuchsia":              {255, 0, 255, 255},
	"gainsboro":            {220, 220, 220, 255},
	"ghostwhite":           {248, 248, 255, 255},
	"gold":                 {255, 215, 0, 255},
	"goldenrod":            {218, 165, 32, 255},
	"gray":                 {128, 128, 128, 255},
	"green":                {0, 128, 0, 255},
	"greenyellow":          {173, 255, 47, 255},
	"honeydew":             {240, 255, 240, 255},
	"hotpink":              {255, 105, 180, 255},
	"indianred":            {205, 92, 92, 255},
	"indigo":               {75, 0, 130, 255},
	"ivory":                {255, 255, 240, 255},
	"khaki":                {240, 230, 140, 255},
	"lavender":             {230, 230, 250, 255},
	"lavenderblush":        {255, 240, 245, 255},
	"lawngreen":            {124, 252, 0, 255},
	"lemonchiffon":         {255, 250, 205, 255},
	"lightblue":            {173, 216, 230, 255},
	"lightcoral":           {240, 128, 128, 255},
	"lightcyan":            {224, 255, 255, 255},
	"lightgoldenrodyellow": {250, 250, 210, 255},
	"lightgray":            {211, 211, 211, 255},
	"lightgreen":           {144, 238, 144, 255},
	"lightpink":            {255, 182, 193, 255},
	"lightsalmon":          {255, 160, 122, 255},
	"lightseagreen":        {32, 178, 170, 255},
	"lightskyblue":         {135, 206, 250, 255},
	"lightslategray":       {119, 136, 153, 255},
	"lightsteelblue":       {176, 196, 222, 255},
	"lightyellow":          {255, 255, 224, 255},
	"lime":                 {0, 255, 0, 255},
	"limegreen":            {50, 205, 50, 255},
	"linen":                {250, 240, 230, 255},
	"magenta":              {255, 0, 255, 255},
	"maroon":               {128, 0, 0, 255},
	"mediumaquamarine":     {102, 205, 170, 255},
	"mediumblue":           {0, 0, 205, 255},
	"mediumorchid":         {186, 85, 211, 255},
	"mediumpurple":         {147, 112, 219, 255},
	"mediumseagreen":       {60, 179, 113, 255},
	"mediumslateblue":      {123, 104, 238, 255},
	"mediumspringgreen":    {0, 250, 154, 255},
	"mediumturquoise":      {72, 209, 204, 255},
	"mediumvioletred":      {199, 21, 133, 255},
	"midnightblue":         {25, 25, 112, 255},
	"mintcream":            {245, 255, 250, 255},
	"mistyrose":            {255, 228, 225, 255},
	"moccasin":             {255, 228, 181, 255},
	"navajowhite":          {255, 222, 173, 255},
	"navy":                 {0, 0, 128, 255},
	"oldlace":              {253, 245, 230, 255},
	"olive":                {128, 128, 0, 255},
	"olivedrab":            {107, 142, 35, 255},
	"orange":               {255, 165, 0, 255},
	"orangered":            {255, 69, 0, 255},
	"orchid":               {218, 112, 214, 255},
	"palegoldenrod":        {238, 232, 170, 255},
	"palegreen":            {152, 251, 152, 255},
	"paleturquoise":        {175, 238, 238, 255},
	"palevioletred":        {219, 112, 147, 255},
	"papayawhip":           {255, 239, 213, 255},
	"peachpuff":            {255, 218, 185, 255},
	"peru":                 {205, 133, 63, 255},
	"pink":                 {255, 192, 203, 255},
	"plum":                 {221, 160, 221, 255},
	"powderblue":           {176, 224, 230, 255},
	"purple":               {128, 0, 128, 255},
	"red":                  {255, 0, 0, 255},
	"rosybrown":            {188, 143, 143, 255},
	"royalblue":            {65, 105, 225, 255},
	"saddlebrown":          {139, 69, 19, 255},
	"salmon":               {250, 128, 114, 255},
	"sandybrown":           {244, 164, 96, 255},
	"seagreen":             {46, 139, 87, 255},
	"seashell":             {255, 245, 238, 255},
	"sienna":               {160, 82, 45, 255},
	"silver":               {192, 192, 192, 255},
	"skyblue":              {135, 206, 235, 255},
	"slateblue":            {106, 90, 205, 255},
	"slategray":            {112, 128, 144, 255},
	"snow":                 {255, 250, 250, 255},
	"springgreen":          {0, 255, 127, 255},
	"steelblue":            {70, 130, 180, 255},
	"tan":                  {210, 180, 140, 255},
	"teal":                 {0, 128, 128, 255},
	"thistle":              {216, 191, 216, 255},
	"tomato":               {255, 99, 71, 255},
	"turquoise":            {64, 224, 208, 255},
	"violet":               {238, 130, 238, 255},
	"wheat":                {245, 222, 179, 255},
	"white":                {255, 255, 255, 255},
	"whitesmoke":           {245, 245, 245, 255},
	"yellow":               {255, 255, 0, 255},
	"yellowgreen":          {154, 205, 50, 255},

	// Windows system colors — mapped to Windows 10 defaults so FRX files that
	// serialize system-color names (e.g. Fill.Color="Highlight") render with
	// a reasonable colour rather than transparent.
	"highlight":            {0, 120, 215, 255}, // Windows 10 accent / selection blue
	"highlighttext":        {255, 255, 255, 255},
	"windowtext":           {0, 0, 0, 255},
	"window":               {255, 255, 255, 255},
	"btnface":              {240, 240, 240, 255},
	"btntext":              {0, 0, 0, 255},
	"desktop":              {0, 0, 0, 255},
	"activecaption":        {0, 120, 215, 255},
	"inactivecaption":      {191, 205, 219, 255},
	"menu":                 {240, 240, 240, 255},
	"menutext":             {0, 0, 0, 255},
	"scrollbar":            {200, 200, 200, 255},
	"graytext":             {109, 109, 109, 255},
	"infotext":             {0, 0, 0, 255},
	"infobk":               {255, 255, 225, 255},
}

// ParseColor parses a color from various string formats:
//
//   - "#RGB"        — 3-digit shorthand; each digit is doubled; alpha = 0xFF.
//   - "#RRGGBB"     — 6-digit RGB; alpha = 0xFF.
//   - "#AARRGGBB"   — 8-digit ARGB; the first two hex digits are alpha.
//   - "R, G, B"     — .NET ColorConverter 3-component format; alpha = 0xFF.
//   - "A, R, G, B"  — .NET ColorConverter 4-component format.
//   - A CSS/Windows named color (case-insensitive, e.g. "White", "LightGray").
//   - A decimal integer string representing a signed 32-bit ARGB value
//     (compatible with .NET's Color.ToArgb()).
//
// Returns an error when the string cannot be recognised as any of those formats.
func ParseColor(s string) (color.RGBA, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return color.RGBA{}, fmt.Errorf("utils.ParseColor: empty string")
	}

	if strings.HasPrefix(s, "#") {
		hex := s[1:]
		switch len(hex) {
		case 3:
			// "#RGB" → expand each nibble to a byte, alpha = 0xFF
			rr := string([]byte{hex[0], hex[0]})
			gg := string([]byte{hex[1], hex[1]})
			bb := string([]byte{hex[2], hex[2]})
			hex = "FF" + rr + gg + bb
		case 6:
			// "#RRGGBB" → prepend full alpha
			hex = "FF" + hex
		case 8:
			// "#AARRGGBB" — already complete
		default:
			return color.RGBA{}, fmt.Errorf("utils.ParseColor: invalid hex length in %q", s)
		}

		v, err := strconv.ParseUint(hex, 16, 32)
		if err != nil {
			return color.RGBA{}, fmt.Errorf("utils.ParseColor: invalid hex value %q: %w", s, err)
		}
		// hex is AARRGGBB
		return color.RGBA{
			A: uint8(v >> 24),
			R: uint8(v >> 16),
			G: uint8(v >> 8),
			B: uint8(v),
		}, nil
	}

	// Try "R, G, B" or "A, R, G, B" comma-separated format (.NET ColorConverter).
	if strings.Contains(s, ",") {
		parts := strings.Split(s, ",")
		var nums [4]int
		for i, p := range parts {
			if i >= 4 {
				break
			}
			n, err := strconv.Atoi(strings.TrimSpace(p))
			if err != nil {
				goto tryNamed
			}
			nums[i] = n
		}
		switch len(parts) {
		case 3:
			// R, G, B — alpha is 255
			return color.RGBA{R: uint8(nums[0]), G: uint8(nums[1]), B: uint8(nums[2]), A: 255}, nil
		case 4:
			// A, R, G, B
			return color.RGBA{A: uint8(nums[0]), R: uint8(nums[1]), G: uint8(nums[2]), B: uint8(nums[3])}, nil
		}
	}

tryNamed:
	// Try named color (case-insensitive).
	if c, ok := namedColors[strings.ToLower(s)]; ok {
		return c, nil
	}

	// Try decimal ARGB integer (possibly negative, as .NET Color.ToArgb() returns int32).
	if n, err := strconv.ParseInt(s, 10, 64); err == nil {
		argb := uint32(n) // reinterpret sign bit as alpha bit
		return color.RGBA{
			A: uint8(argb >> 24),
			R: uint8(argb >> 16),
			G: uint8(argb >> 8),
			B: uint8(argb),
		}, nil
	}

	return color.RGBA{}, fmt.Errorf("utils.ParseColor: unrecognised color format %q", s)
}

// FormatColor formats c as an "#AARRGGBB" uppercase hex string, which is the
// canonical FRX serialisation format.
func FormatColor(c color.RGBA) string {
	return fmt.Sprintf("#%02X%02X%02X%02X", c.A, c.R, c.G, c.B)
}

// ColorEqual reports whether a and b represent the same color.
func ColorEqual(a, b color.RGBA) bool {
	return a == b
}
