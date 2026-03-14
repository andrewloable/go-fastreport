package style

import "image/color"

// Common colours used as defaults throughout go-fastreport.
var (
	// ColorWhite is opaque white.
	ColorWhite = color.RGBA{R: 255, G: 255, B: 255, A: 255}
	// ColorBlack is opaque black.
	ColorBlack = color.RGBA{R: 0, G: 0, B: 0, A: 255}
	// ColorTransparent is fully transparent.
	ColorTransparent = color.RGBA{R: 0, G: 0, B: 0, A: 0}
)
