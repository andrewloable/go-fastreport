package preview

import (
	"encoding/gob"
	"fmt"
	"image/color"
	"io"

	"github.com/andrewloable/go-fastreport/style"
)

// fpxMagic is the 4-byte magic number at the start of every FPX file.
const fpxMagic = "FPX1"

// ── gob-encodable mirror structs ─────────────────────────────────────────────
// These are separate from the runtime structs to allow safe schema evolution.

type fpxFile struct {
	Magic     string
	BlobStore fpxBlobStore
	Pages     []fpxPage
	Bookmarks []fpxBookmark
	Outline   *fpxOutlineItem
}

type fpxBlobStore struct {
	Index map[string]int
	Blobs [][]byte
}

type fpxPage struct {
	PageNo    int
	Width     float32
	Height    float32
	Bands     []fpxBand
	Watermark *fpxWatermark
}

type fpxBand struct {
	Name    string
	Kind    int // PreparedBandKind: 0=Normal, 1=PageFooter, 2=Overlay
	Top     float32
	Height  float32
	Objects []fpxObject
}

type fpxBorderLine struct {
	Color color.RGBA
	Style int
	Width float32
}

type fpxBorder struct {
	Lines        [4]fpxBorderLine
	VisibleLines int
	Shadow       bool
	ShadowColor  color.RGBA
	ShadowWidth  float32
}

type fpxObject struct {
	Name           string
	Kind           int
	Left           float32
	Top            float32
	Width          float32
	Height         float32
	Text           string
	BlobIdx        int
	ShapeKind      int
	ShapeCurve     float32
	LineDiagonal   bool
	Points         [][2]float32
	Font           style.Font
	TextColor      color.RGBA
	FillColor      color.RGBA
	HorzAlign      int
	VertAlign      int
	Border         fpxBorder
	WordWrap       bool
	Checked         bool
	CheckedSymbol   int
	UncheckedSymbol int
	CheckColor      color.RGBA
	CheckWidthRatio float32
	Duplicates      int
	IsBarcode       bool
	BarcodeModules [][]bool
	HyperlinkKind  int
	HyperlinkValue string
}

type fpxBookmark struct {
	Name    string
	PageIdx int
	OffsetY float32
}

type fpxOutlineItem struct {
	Text     string
	PageIdx  int
	OffsetY  float32
	Children []*fpxOutlineItem
}

type fpxWatermark struct {
	Enabled           bool
	Text              string
	Font              style.Font
	TextColor         color.RGBA
	TextRotation      int
	ShowTextOnTop     bool
	ImageBlobIdx      int
	ImageSize         int
	ImageTransparency float32
	ShowImageOnTop    bool
}

// ── Save ──────────────────────────────────────────────────────────────────────

// Save encodes the PreparedPages (including all blobs, bookmarks, and outline)
// into the binary FPX format and writes it to w.
func (pp *PreparedPages) Save(w io.Writer) error {
	f := fpxFile{
		Magic:     fpxMagic,
		BlobStore: encodeBlobStore(pp.BlobStore),
		Pages:     encodePages(pp.pages),
		Bookmarks: encodeBookmarks(pp.Bookmarks),
		Outline:   encodeOutlineItem(pp.Outline.Root),
	}
	if err := gob.NewEncoder(w).Encode(f); err != nil {
		return fmt.Errorf("fpx save: %w", err)
	}
	return nil
}

func encodeBlobStore(bs *BlobStore) fpxBlobStore {
	n := bs.Count()
	out := fpxBlobStore{
		Index: make(map[string]int, n),
		Blobs: make([][]byte, n),
	}
	for i := 0; i < n; i++ {
		src := bs.Get(i)
		cp := make([]byte, len(src))
		copy(cp, src)
		out.Blobs[i] = cp
		if key := bs.GetSource(i); key != "" {
			out.Index[key] = i
		}
	}
	return out
}

func encodePages(pages []*PreparedPage) []fpxPage {
	out := make([]fpxPage, len(pages))
	for i, pg := range pages {
		out[i] = fpxPage{
			PageNo: pg.PageNo,
			Width:  pg.Width,
			Height: pg.Height,
			Bands:  encodeBands(pg.Bands),
		}
		if pg.Watermark != nil {
			wm := encodeWatermark(pg.Watermark)
			out[i].Watermark = &wm
		}
	}
	return out
}

func encodeBands(bands []*PreparedBand) []fpxBand {
	out := make([]fpxBand, len(bands))
	for i, b := range bands {
		out[i] = fpxBand{
			Name:    b.Name,
			Kind:    int(b.Kind),
			Top:     b.Top,
			Height:  b.Height,
			Objects: encodeObjects(b.Objects),
		}
	}
	return out
}

func encodeBorder(b style.Border) fpxBorder {
	out := fpxBorder{
		VisibleLines: int(b.VisibleLines),
		Shadow:       b.Shadow,
		ShadowColor:  b.ShadowColor,
		ShadowWidth:  b.ShadowWidth,
	}
	for i, bl := range b.Lines {
		if bl != nil {
			out.Lines[i] = fpxBorderLine{Color: bl.Color, Style: int(bl.Style), Width: bl.Width}
		}
	}
	return out
}

func decodeBorder(f fpxBorder) style.Border {
	var b style.Border
	b.VisibleLines = style.BorderLines(f.VisibleLines)
	b.Shadow = f.Shadow
	b.ShadowColor = f.ShadowColor
	b.ShadowWidth = f.ShadowWidth
	for i, fl := range f.Lines {
		b.Lines[i] = &style.BorderLine{Color: fl.Color, Style: style.LineStyle(fl.Style), Width: fl.Width}
	}
	return b
}

func encodeObjects(objs []PreparedObject) []fpxObject {
	out := make([]fpxObject, len(objs))
	for i, o := range objs {
		pts := make([][2]float32, len(o.Points))
		copy(pts, o.Points)
		mods := cloneBoolMatrix(o.BarcodeModules)
		out[i] = fpxObject{
			Name:           o.Name,
			Kind:           int(o.Kind),
			Left:           o.Left,
			Top:            o.Top,
			Width:          o.Width,
			Height:         o.Height,
			Text:           o.Text,
			BlobIdx:        o.BlobIdx,
			ShapeKind:      o.ShapeKind,
			ShapeCurve:     o.ShapeCurve,
			LineDiagonal:   o.LineDiagonal,
			Points:         pts,
			Font:           o.Font,
			TextColor:      o.TextColor,
			FillColor:      o.FillColor,
			HorzAlign:      o.HorzAlign,
			VertAlign:      o.VertAlign,
			Border:         encodeBorder(o.Border),
			WordWrap:       o.WordWrap,
			Checked:         o.Checked,
			CheckedSymbol:   int(o.CheckedSymbol),
			UncheckedSymbol: int(o.UncheckedSymbol),
			CheckColor:      o.CheckColor,
			CheckWidthRatio: o.CheckWidthRatio,
			Duplicates:      int(o.Duplicates),
			IsBarcode:       o.IsBarcode,
			BarcodeModules: mods,
			HyperlinkKind:  o.HyperlinkKind,
			HyperlinkValue: o.HyperlinkValue,
		}
	}
	return out
}

func cloneBoolMatrix(m [][]bool) [][]bool {
	if m == nil {
		return nil
	}
	out := make([][]bool, len(m))
	for i, row := range m {
		cp := make([]bool, len(row))
		copy(cp, row)
		out[i] = cp
	}
	return out
}

func encodeBookmarks(bk *Bookmarks) []fpxBookmark {
	out := make([]fpxBookmark, len(bk.items))
	for i, b := range bk.items {
		out[i] = fpxBookmark{Name: b.Name, PageIdx: b.PageIdx, OffsetY: b.OffsetY}
	}
	return out
}

func encodeOutlineItem(item *OutlineItem) *fpxOutlineItem {
	if item == nil {
		return nil
	}
	out := &fpxOutlineItem{Text: item.Text, PageIdx: item.PageIdx, OffsetY: item.OffsetY}
	for _, c := range item.Children {
		out.Children = append(out.Children, encodeOutlineItem(c))
	}
	return out
}

func encodeWatermark(wm *PreparedWatermark) fpxWatermark {
	return fpxWatermark{
		Enabled:           wm.Enabled,
		Text:              wm.Text,
		Font:              wm.Font,
		TextColor:         wm.TextColor,
		TextRotation:      int(wm.TextRotation),
		ShowTextOnTop:     wm.ShowTextOnTop,
		ImageBlobIdx:      wm.ImageBlobIdx,
		ImageSize:         int(wm.ImageSize),
		ImageTransparency: wm.ImageTransparency,
		ShowImageOnTop:    wm.ShowImageOnTop,
	}
}

// ── Load ──────────────────────────────────────────────────────────────────────

// Load decodes a PreparedPages collection from the binary FPX format read from r.
// The returned PreparedPages is fully populated (pages, blobs, bookmarks, outline).
func Load(r io.Reader) (*PreparedPages, error) {
	var f fpxFile
	if err := gob.NewDecoder(r).Decode(&f); err != nil {
		return nil, fmt.Errorf("fpx load: %w", err)
	}
	if f.Magic != fpxMagic {
		return nil, fmt.Errorf("fpx load: bad magic %q", f.Magic)
	}

	pp := New()
	decodeBlobStore(f.BlobStore, pp.BlobStore)
	decodePages(f.Pages, pp)
	decodeBookmarks(f.Bookmarks, pp.Bookmarks)
	decodeOutlineItem(f.Outline, pp.Outline.Root)
	return pp, nil
}

func decodeBlobStore(src fpxBlobStore, dst *BlobStore) {
	// Build a reverse map: index → source key, so we can pass the key to AddOrUpdate.
	idxToKey := make(map[int]string, len(src.Index))
	for k, v := range src.Index {
		idxToKey[v] = k
	}
	for i, b := range src.Blobs {
		cp := make([]byte, len(b))
		copy(cp, b)
		dst.AddOrUpdate(cp, idxToKey[i])
	}
}

func decodePages(src []fpxPage, pp *PreparedPages) {
	pp.pages = make([]*PreparedPage, len(src))
	for i, fp := range src {
		pg := &PreparedPage{
			PageNo: fp.PageNo,
			Width:  fp.Width,
			Height: fp.Height,
			Bands:  decodeBands(fp.Bands),
		}
		if fp.Watermark != nil {
			wm := decodeWatermark(fp.Watermark)
			pg.Watermark = &wm
		}
		pp.pages[i] = pg
	}
	if len(pp.pages) > 0 {
		pp.curPage = len(pp.pages) - 1
	}
}

func decodeBands(src []fpxBand) []*PreparedBand {
	out := make([]*PreparedBand, len(src))
	for i, fb := range src {
		out[i] = &PreparedBand{
			Name:    fb.Name,
			Kind:    PreparedBandKind(fb.Kind),
			Top:     fb.Top,
			Height:  fb.Height,
			Objects: decodeObjects(fb.Objects),
		}
	}
	return out
}

func decodeObjects(src []fpxObject) []PreparedObject {
	out := make([]PreparedObject, len(src))
	for i, fo := range src {
		pts := make([][2]float32, len(fo.Points))
		copy(pts, fo.Points)
		out[i] = PreparedObject{
			Name:           fo.Name,
			Kind:           ObjectType(fo.Kind),
			Left:           fo.Left,
			Top:            fo.Top,
			Width:          fo.Width,
			Height:         fo.Height,
			Text:           fo.Text,
			BlobIdx:        fo.BlobIdx,
			ShapeKind:      fo.ShapeKind,
			ShapeCurve:     fo.ShapeCurve,
			LineDiagonal:   fo.LineDiagonal,
			Points:         pts,
			Font:           fo.Font,
			TextColor:      fo.TextColor,
			FillColor:      fo.FillColor,
			HorzAlign:      fo.HorzAlign,
			VertAlign:      fo.VertAlign,
			Border:         decodeBorder(fo.Border),
			WordWrap:       fo.WordWrap,
			Checked:         fo.Checked,
			CheckedSymbol:   fo.CheckedSymbol,
			UncheckedSymbol: fo.UncheckedSymbol,
			CheckColor:      fo.CheckColor,
			CheckWidthRatio: fo.CheckWidthRatio,
			Duplicates:      DuplicatesMode(fo.Duplicates),
			IsBarcode:       fo.IsBarcode,
			BarcodeModules: cloneBoolMatrix(fo.BarcodeModules),
			HyperlinkKind:  fo.HyperlinkKind,
			HyperlinkValue: fo.HyperlinkValue,
		}
	}
	return out
}

func decodeBookmarks(src []fpxBookmark, dst *Bookmarks) {
	for _, fb := range src {
		dst.Add(&Bookmark{Name: fb.Name, PageIdx: fb.PageIdx, OffsetY: fb.OffsetY})
	}
}

func decodeOutlineItem(src *fpxOutlineItem, dst *OutlineItem) {
	if src == nil || dst == nil {
		return
	}
	dst.Text = src.Text
	dst.PageIdx = src.PageIdx
	dst.OffsetY = src.OffsetY
	for _, c := range src.Children {
		child := &OutlineItem{}
		decodeOutlineItem(c, child)
		dst.Children = append(dst.Children, child)
	}
}

func decodeWatermark(src *fpxWatermark) PreparedWatermark {
	return PreparedWatermark{
		Enabled:           src.Enabled,
		Text:              src.Text,
		Font:              src.Font,
		TextColor:         src.TextColor,
		TextRotation:      WatermarkTextRotation(src.TextRotation),
		ShowTextOnTop:     src.ShowTextOnTop,
		ImageBlobIdx:      src.ImageBlobIdx,
		ImageSize:         WatermarkImageSize(src.ImageSize),
		ImageTransparency: src.ImageTransparency,
		ShowImageOnTop:    src.ShowImageOnTop,
	}
}
