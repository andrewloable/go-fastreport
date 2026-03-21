// calcbounds.go provides CalcBounds() implementations for concrete barcode types.
// Ported from C# BarcodeBase.CalcBounds() and LinearBarcodeBase.CalcBounds().
package barcode

// patternGetter is a local interface matching the actual GetPattern signature
// used by all linear barcode concrete types (returns string + error).
type patternGetter interface {
	GetPattern() (string, error)
	GetWideBarRatio() float32
}

// ── 2D barcode textAdd ──────────────────────────────────────────────────────
// C# Barcode2DBase.CalcBounds() adds FontHeight to height when showText is true.
// FontHeight for the default Arial 8pt: Font.Height * ScreenDpiFX * 18/13 = 18.
const defaultFontHeight2D float32 = 18

// textAdd2D returns the height to add for show-text on 2D barcodes.
func textAdd2D(showText bool) float32 {
	if showText {
		return defaultFontHeight2D
	}
	return 0
}

// ── Linear barcode CalcBounds ────────────────────────────────────────────────
// C# LinearBarcodeBase.CalcBounds() (LinearBarcodeBase.cs:424-466):
//
//	float barWidth = GetWidth(Code);
//	drawArea.Width = barWidth + extra1 + extra2;
//	return new SizeF(drawArea.Width * 1.25f, 0);
//
// In Go we derive the width from GetPatternWidth * 1.25.
func linearPatternCalcBounds(pp patternGetter) (float32, float32) {
	pattern, err := pp.GetPattern()
	if err != nil || pattern == "" {
		return 0, 0
	}
	// Use the type's default WideBarRatio, but allow FRX deserialization to override.
	// C# LinearBarcodeBase: WideBarRatio is a settable property; FRX stores it as
	// Barcode.WideBarRatio="2.25" etc. In Go we check for a WBROverride from
	// BarcodeObject.Deserialize().
	wbr := pp.GetWideBarRatio()
	type wbrOverrider interface{ WBROverride() float32 }
	if ov, ok := pp.(wbrOverrider); ok {
		if v := ov.WBROverride(); v != 0 {
			wbr = v
		}
	}
	modules := MakeModules(wbr)
	return GetPatternWidth(pattern, modules) * 1.25, 0
}

// ── Code128Barcode — CalcBounds ───────────────────────────────────────────────

// CalcBounds returns the natural width of the encoded Code128 symbol.
func (c *Code128Barcode) CalcBounds() (float32, float32) {
	if c.encodedText == "" {
		return 0, 0
	}
	return linearPatternCalcBounds(c)
}

// ── Code128ABarcode — CalcBounds ─────────────────────────────────────────────

// CalcBounds returns the natural width of the encoded Code128A symbol.
func (c *Code128ABarcode) CalcBounds() (float32, float32) {
	if c.encodedText == "" {
		return 0, 0
	}
	return linearPatternCalcBounds(c)
}

// ── Code128BBarcode — CalcBounds ─────────────────────────────────────────────

// CalcBounds returns the natural width of the encoded Code128B symbol.
func (c *Code128BBarcode) CalcBounds() (float32, float32) {
	if c.encodedText == "" {
		return 0, 0
	}
	return linearPatternCalcBounds(c)
}

// ── Code128CBarcode — CalcBounds ─────────────────────────────────────────────

// CalcBounds returns the natural width of the encoded Code128C symbol.
func (c *Code128CBarcode) CalcBounds() (float32, float32) {
	if c.encodedText == "" {
		return 0, 0
	}
	return linearPatternCalcBounds(c)
}

// ── Code39Barcode — CalcBounds ───────────────────────────────────────────────

// CalcBounds returns the natural width of the encoded Code39 symbol.
func (c *Code39Barcode) CalcBounds() (float32, float32) {
	if c.encodedText == "" {
		return 0, 0
	}
	return linearPatternCalcBounds(c)
}

// ── Code39ExtendedBarcode — CalcBounds ───────────────────────────────────────

// CalcBounds returns the natural width of the encoded Code39Extended symbol.
func (c *Code39ExtendedBarcode) CalcBounds() (float32, float32) {
	if c.encodedText == "" {
		return 0, 0
	}
	return linearPatternCalcBounds(c)
}

// ── Code93Barcode — CalcBounds ───────────────────────────────────────────────

// CalcBounds returns the natural width of the encoded Code93 symbol.
func (c *Code93Barcode) CalcBounds() (float32, float32) {
	if c.encodedText == "" {
		return 0, 0
	}
	return linearPatternCalcBounds(c)
}

// ── Code93ExtendedBarcode — CalcBounds ───────────────────────────────────────

// CalcBounds returns the natural width of the encoded Code93Extended symbol.
func (c *Code93ExtendedBarcode) CalcBounds() (float32, float32) {
	if c.encodedText == "" {
		return 0, 0
	}
	return linearPatternCalcBounds(c)
}

// ── Code2of5Barcode — CalcBounds ─────────────────────────────────────────────

// CalcBounds returns the natural width of the encoded 2-of-5 symbol.
func (c *Code2of5Barcode) CalcBounds() (float32, float32) {
	if c.encodedText == "" {
		return 0, 0
	}
	return linearPatternCalcBounds(c)
}

// ── Code2of5IndustrialBarcode — CalcBounds ───────────────────────────────────

// CalcBounds returns the natural width of the encoded Standard 2-of-5 symbol.
func (c *Code2of5IndustrialBarcode) CalcBounds() (float32, float32) {
	if c.encodedText == "" {
		return 0, 0
	}
	return linearPatternCalcBounds(c)
}

// ── Code2of5MatrixBarcode — CalcBounds ───────────────────────────────────────

// CalcBounds returns the natural width of the encoded Matrix 2-of-5 symbol.
func (c *Code2of5MatrixBarcode) CalcBounds() (float32, float32) {
	if c.encodedText == "" {
		return 0, 0
	}
	return linearPatternCalcBounds(c)
}

// ── CodabarBarcode — CalcBounds ──────────────────────────────────────────────

// CalcBounds returns the natural width of the encoded Codabar symbol.
func (c *CodabarBarcode) CalcBounds() (float32, float32) {
	if c.encodedText == "" {
		return 0, 0
	}
	return linearPatternCalcBounds(c)
}

// ── EAN8Barcode — CalcBounds ─────────────────────────────────────────────────

// CalcBounds returns the natural width of the encoded EAN-8 symbol.
func (e *EAN8Barcode) CalcBounds() (float32, float32) {
	if e.encodedText == "" {
		return 0, 0
	}
	return linearPatternCalcBounds(e)
}

// ── EAN13Barcode — CalcBounds ────────────────────────────────────────────────

// CalcBounds returns the natural width of the encoded EAN-13 symbol.
// C# BarcodeEAN13 constructor sets extra1=8 (quiet zone on left).
// C# LinearBarcodeBase.CalcBounds(): drawArea.Width = barWidth + extra1 + extra2.
// EAN-13: extra1=8, extra2=0 → barWidth+8 modules × 1.25.
func (e *EAN13Barcode) CalcBounds() (float32, float32) {
	if e.encodedText == "" {
		return 0, 0
	}
	w, h := linearPatternCalcBounds(e)
	if w == 0 {
		return 0, 0
	}
	// Add extra1=8 modules quiet zone (C# BarcodeEAN13 constructor: extra1=8).
	return w + 8*1.25, h
}

// ── UPCABarcode — CalcBounds ─────────────────────────────────────────────────

// CalcBounds returns the natural width of the encoded UPC-A symbol.
// C# BarcodeUPC_A extends BarcodeUPC_E0 which sets extra1=8, extra2=8.
// UPC-A: extra1=8, extra2=8 → barWidth+16 modules × 1.25.
func (u *UPCABarcode) CalcBounds() (float32, float32) {
	if u.encodedText == "" {
		return 0, 0
	}
	w, h := linearPatternCalcBounds(u)
	if w == 0 {
		return 0, 0
	}
	// Add extra1=8 + extra2=8 = 16 modules quiet zones (C# BarcodeUPC_E0: extra1=8, extra2=8).
	return w + 16*1.25, h
}

// ── UPCEBarcode — CalcBounds ─────────────────────────────────────────────────

// CalcBounds returns the natural width of the encoded UPC-E symbol.
// C# BarcodeUPC_E0 sets extra1=8, extra2=8 → barWidth+16 modules × 1.25.
func (u *UPCEBarcode) CalcBounds() (float32, float32) {
	if u.encodedText == "" {
		return 0, 0
	}
	w, h := linearPatternCalcBounds(u)
	if w == 0 {
		return 0, 0
	}
	// Add extra1=8 + extra2=8 = 16 modules quiet zones (C# BarcodeUPC_E0: extra1=8, extra2=8).
	return w + 16*1.25, h
}

// ── UPCE0Barcode — CalcBounds ────────────────────────────────────────────────

// CalcBounds returns the natural width of the encoded UPC-E0 symbol.
// C# BarcodeUPC_E0 sets extra1=8, extra2=8 → barWidth+16 modules × 1.25.
func (u *UPCE0Barcode) CalcBounds() (float32, float32) {
	if u.encodedText == "" {
		return 0, 0
	}
	w, h := linearPatternCalcBounds(u)
	if w == 0 {
		return 0, 0
	}
	// Add extra1=8 + extra2=8 = 16 modules quiet zones (C# BarcodeUPC_E0: extra1=8, extra2=8).
	return w + 16*1.25, h
}

// ── UPCE1Barcode — CalcBounds ────────────────────────────────────────────────

// CalcBounds returns the natural width of the encoded UPC-E1 symbol.
// Inherits extra1=8, extra2=8 from BarcodeUPC_E0.
func (u *UPCE1Barcode) CalcBounds() (float32, float32) {
	if u.encodedText == "" {
		return 0, 0
	}
	w, h := linearPatternCalcBounds(u)
	if w == 0 {
		return 0, 0
	}
	// Add extra1=8 + extra2=8 = 16 modules quiet zones (C# BarcodeUPC_E0: extra1=8, extra2=8).
	return w + 16*1.25, h
}

// ── MSIBarcode — CalcBounds ──────────────────────────────────────────────────

// CalcBounds returns the natural width of the encoded MSI symbol.
func (m *MSIBarcode) CalcBounds() (float32, float32) {
	if m.encodedText == "" {
		return 0, 0
	}
	return linearPatternCalcBounds(m)
}

// ── GS1Barcode — CalcBounds ──────────────────────────────────────────────────

// CalcBounds returns the natural width of the encoded GS1-128 symbol.
func (g *GS1Barcode) CalcBounds() (float32, float32) {
	if g.encodedText == "" {
		return 0, 0
	}
	return linearPatternCalcBounds(g)
}

// ── GS1_128Barcode — CalcBounds ──────────────────────────────────────────────

// CalcBounds returns the natural width of the encoded GS1-128 symbol.
func (g *GS1_128Barcode) CalcBounds() (float32, float32) {
	if g.encodedText == "" {
		return 0, 0
	}
	return linearPatternCalcBounds(g)
}

// ── ITF14Barcode — CalcBounds ────────────────────────────────────────────────

// CalcBounds returns the natural width of the encoded ITF-14 symbol.
func (i *ITF14Barcode) CalcBounds() (float32, float32) {
	if i.encodedText == "" {
		return 0, 0
	}
	return linearPatternCalcBounds(i)
}

// ── DeutscheIdentcodeBarcode — CalcBounds ────────────────────────────────────

// CalcBounds returns the natural width of the encoded Deutsche Identcode symbol.
func (d *DeutscheIdentcodeBarcode) CalcBounds() (float32, float32) {
	if d.encodedText == "" {
		return 0, 0
	}
	return linearPatternCalcBounds(d)
}

// ── DeutscheLeitcodeBarcode — CalcBounds ─────────────────────────────────────

// CalcBounds returns the natural width of the encoded Deutsche Leitcode symbol.
func (d *DeutscheLeitcodeBarcode) CalcBounds() (float32, float32) {
	if d.encodedText == "" {
		return 0, 0
	}
	return linearPatternCalcBounds(d)
}

// ── Supplement2Barcode — CalcBounds ──────────────────────────────────────────

// CalcBounds returns the natural width of the encoded Supplement 2 symbol.
func (s *Supplement2Barcode) CalcBounds() (float32, float32) {
	if s.encodedText == "" {
		return 0, 0
	}
	return linearPatternCalcBounds(s)
}

// ── Supplement5Barcode — CalcBounds ──────────────────────────────────────────

// CalcBounds returns the natural width of the encoded Supplement 5 symbol.
func (s *Supplement5Barcode) CalcBounds() (float32, float32) {
	if s.encodedText == "" {
		return 0, 0
	}
	return linearPatternCalcBounds(s)
}

// ── QRBarcode — CalcBounds ───────────────────────────────────────────────────
// C# BarcodeQR.CalcBounds() (BarcodeQR.cs:296-300): matrix.Width * PixelSize, matrix.Height * PixelSize
// PixelSize = 4 for QR.

// CalcBounds returns the natural (width, height) of the encoded QR Code symbol.
func (q *QRBarcode) CalcBounds() (float32, float32) {
	if q.encodedText == "" {
		return 0, 0
	}
	_, rows, cols := q.GetMatrix()
	const pixelSize = 4
	return float32(cols) * pixelSize, float32(rows)*pixelSize + textAdd2D(q.showText)
}

// ── DataMatrixBarcode — CalcBounds ───────────────────────────────────────────
// C# BarcodeDatamatrix.CalcBounds() (BarcodeDatamatrix.cs:1096-1100): width * PixelSize, height * PixelSize
// Default PixelSize = 3 for DataMatrix.

// CalcBounds returns the natural (width, height) of the encoded DataMatrix symbol.
func (d *DataMatrixBarcode) CalcBounds() (float32, float32) {
	if d.encodedText == "" {
		return 0, 0
	}
	_, rows, cols := d.GetMatrix()
	const pixelSize = 3
	return float32(cols) * pixelSize, float32(rows)*pixelSize + textAdd2D(d.showText)
}

// ── AztecBarcode — CalcBounds ────────────────────────────────────────────────
// C# BarcodeAztec.CalcBounds() (BarcodeAztec.cs:45-49): matrix.Width * PIXEL_SIZE, matrix.Height * PIXEL_SIZE
// PIXEL_SIZE = 4.

// CalcBounds returns the natural (width, height) of the encoded Aztec symbol.
func (a *AztecBarcode) CalcBounds() (float32, float32) {
	if a.encodedText == "" {
		return 0, 0
	}
	_, rows, cols := a.GetMatrix()
	const pixelSize = 4
	return float32(cols) * pixelSize, float32(rows)*pixelSize + textAdd2D(a.showText)
}

// ── PDF417Barcode — CalcBounds ───────────────────────────────────────────────
// C# BarcodePDF417.CalcBounds() (BarcodePDF417.cs:1515-1519): bitColumns * PixelSize.Width, codeRows * PixelSize.Height
// Default PixelSize = {1, 2} for PDF417.

// CalcBounds returns the natural (width, height) of the encoded PDF417 symbol.
// C# BarcodePDF417.cs:1517-1518: bitColumns * PixelSize.Width, codeRows * PixelSize.Height
func (p *PDF417Barcode) CalcBounds() (float32, float32) {
	if p.encodedText == "" {
		return 0, 0
	}
	_, rows, cols := p.GetMatrix()
	pixelW := p.PixelSizeWidth
	pixelH := p.PixelSizeHeight
	if pixelW <= 0 {
		pixelW = 2
	}
	if pixelH <= 0 {
		pixelH = 8
	}
	return float32(cols) * float32(pixelW), float32(rows)*float32(pixelH) + textAdd2D(p.showText)
}

// ── GS1DatamatrixBarcode — CalcBounds ────────────────────────────────────────
// GS1 DataMatrix uses the same pixel size as DataMatrix (3).

// CalcBounds returns the natural (width, height) of the encoded GS1 DataMatrix symbol.
func (g *GS1DatamatrixBarcode) CalcBounds() (float32, float32) {
	if g.encodedText == "" {
		return 0, 0
	}
	_, rows, cols := g.GetMatrix()
	const pixelSize = 3
	return float32(cols) * pixelSize, float32(rows)*pixelSize + textAdd2D(g.showText)
}

// ── SwissQRBarcode — CalcBounds ──────────────────────────────────────────────
// Swiss QR Code uses the same pixel size as QR Code (4).

// CalcBounds returns the natural (width, height) of the encoded Swiss QR symbol.
func (s *SwissQRBarcode) CalcBounds() (float32, float32) {
	if s.encodedText == "" {
		return 0, 0
	}
	_, rows, cols := s.GetMatrix()
	const pixelSize = 4
	return float32(cols) * pixelSize, float32(rows)*pixelSize + textAdd2D(s.showText)
}

// ── PharmacodeBarcode — CalcBounds ───────────────────────────────────────────

// CalcBounds returns the natural width of the encoded Pharmacode symbol.
func (b *PharmacodeBarcode) CalcBounds() (float32, float32) {
	if b.encodedText == "" {
		return 0, 0
	}
	return linearPatternCalcBounds(b)
}

// ── PlesseyBarcode — CalcBounds ──────────────────────────────────────────────

// CalcBounds returns the natural width of the encoded Plessey symbol.
func (b *PlesseyBarcode) CalcBounds() (float32, float32) {
	if b.encodedText == "" {
		return 0, 0
	}
	return linearPatternCalcBounds(b)
}

// ── PostNetBarcode — CalcBounds ──────────────────────────────────────────────

// CalcBounds returns the natural width of the encoded POSTNET symbol.
func (b *PostNetBarcode) CalcBounds() (float32, float32) {
	if b.encodedText == "" {
		return 0, 0
	}
	return linearPatternCalcBounds(b)
}

// ── JapanPost4StateBarcode — CalcBounds ──────────────────────────────────────

// CalcBounds returns the natural width of the encoded Japan Post 4-state symbol.
func (j *JapanPost4StateBarcode) CalcBounds() (float32, float32) {
	if j.encodedText == "" {
		return 0, 0
	}
	return linearPatternCalcBounds(j)
}
