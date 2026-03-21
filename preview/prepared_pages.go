// Package preview implements the prepared-pages collection for go-fastreport.
// A PreparedPages holds the rendered output of a report run.
// It is the Go equivalent of FastReport.Preview.PreparedPages.
package preview

import (
	"fmt"
	"image/color"

	"github.com/andrewloable/go-fastreport/style"
)

// AddPageAction controls behaviour when a new page is added.
type AddPageAction int

const (
	// AddPageActionWriteOver reuses the current slot if possible (for double-pass).
	AddPageActionWriteOver AddPageAction = iota
	// AddPageActionAdd always appends a new page.
	AddPageActionAdd
)

// ── Bookmark ──────────────────────────────────────────────────────────────────

// Bookmark is a named navigation point within the prepared pages.
type Bookmark struct {
	Name    string
	PageIdx int // zero-based page index
	OffsetY float32
}

// Bookmarks is a collection of Bookmark entries.
type Bookmarks struct {
	items          []*Bookmark
	index          map[string]int // name → slice index
	firstPassItems []*Bookmark    // saved by ClearFirstPass for GetPageNo fallback
	firstPassIndex map[string]int // index for firstPassItems
}

// NewBookmarks creates an empty Bookmarks collection.
func NewBookmarks() *Bookmarks {
	return &Bookmarks{index: make(map[string]int)}
}

// Add adds a bookmark. If a bookmark with the same name exists, it is overwritten.
func (bk *Bookmarks) Add(b *Bookmark) {
	if idx, ok := bk.index[b.Name]; ok {
		bk.items[idx] = b
		return
	}
	bk.index[b.Name] = len(bk.items)
	bk.items = append(bk.items, b)
}

// Find returns the bookmark with the given name, or nil if not found.
func (bk *Bookmarks) Find(name string) *Bookmark {
	if idx, ok := bk.index[name]; ok {
		return bk.items[idx]
	}
	return nil
}

// Count returns the number of bookmarks.
func (bk *Bookmarks) Count() int { return len(bk.items) }

// All returns all bookmarks in insertion order.
func (bk *Bookmarks) All() []*Bookmark { return bk.items }

// CurPosition returns the current insertion position (length of items slice).
// Used by the engine to save a restore point for keep-together and double-pass.
// C# equivalent: Bookmarks.CurPosition → items.Count
func (bk *Bookmarks) CurPosition() int { return len(bk.items) }

// Shift adjusts all bookmarks from fromIndex onwards for a keep-together page break.
// Each affected bookmark has its PageIdx incremented by 1, and its OffsetY adjusted
// by (newY - items[fromIndex].OffsetY).
// C# equivalent: Bookmarks.Shift(int index, float newY)
func (bk *Bookmarks) Shift(fromIndex int, newY float32) {
	if fromIndex < 0 || fromIndex >= len(bk.items) {
		return
	}
	topY := bk.items[fromIndex].OffsetY
	shift := newY - topY
	for i := fromIndex; i < len(bk.items); i++ {
		bk.items[i].PageIdx++
		bk.items[i].OffsetY += shift
	}
}

// GetPageNo returns the 1-based page number for the named bookmark, or 0 if not found.
// If not found in the active items, falls back to the saved first-pass items.
// C# equivalent: Bookmarks.GetPageNo — searches items then firstPassItems.
func (bk *Bookmarks) GetPageNo(name string) int {
	b := bk.Find(name)
	if b == nil {
		// Fall back to first-pass items saved by ClearFirstPass.
		if idx, ok := bk.firstPassIndex[name]; ok && idx < len(bk.firstPassItems) {
			return bk.firstPassItems[idx].PageIdx + 1
		}
		return 0
	}
	return b.PageIdx + 1
}

// Clear removes all active bookmarks.
// The first-pass items (saved by ClearFirstPass) are preserved as the fallback.
// C# equivalent: Bookmarks.Clear() → items.Clear()
func (bk *Bookmarks) Clear() {
	bk.items = bk.items[:0]
	bk.index = make(map[string]int)
}

// ClearFirstPass saves the current active items as the first-pass fallback, then
// resets the active items to a new empty list.
// After this call, GetPageNo will search the new active list first, then fall back
// to the saved first-pass items.
// C# equivalent: Bookmarks.ClearFirstPass() → firstPassItems = items; items = new List()
func (bk *Bookmarks) ClearFirstPass() {
	bk.firstPassItems = bk.items
	bk.firstPassIndex = bk.index
	bk.items = nil
	bk.index = make(map[string]int)
}

// ── OutlineItem ───────────────────────────────────────────────────────────────

// OutlineItem is a node in the report outline (PDF bookmarks tree).
type OutlineItem struct {
	Text     string
	PageIdx  int
	OffsetY  float32
	Children []*OutlineItem
}

// AddChild appends a child item.
func (o *OutlineItem) AddChild(child *OutlineItem) {
	o.Children = append(o.Children, child)
}

// Outline is the top-level outline tree with a current-position cursor.
type Outline struct {
	Root         *OutlineItem
	cur          *OutlineItem   // current insertion point
	stack        []*OutlineItem // ancestors of cur
	firstPassPos int            // saved len(Root.Children) for double-pass
}

// NewOutline creates an Outline with an empty root node.
func NewOutline() *Outline {
	root := &OutlineItem{Text: ""}
	return &Outline{Root: root, cur: root}
}

// Add adds an outline item as a child of the current position, then descends into it.
func (o *Outline) Add(text string, pageIdx int, offsetY float32) {
	item := &OutlineItem{Text: text, PageIdx: pageIdx, OffsetY: offsetY}
	o.cur.AddChild(item)
	o.stack = append(o.stack, o.cur)
	o.cur = item
}

// LevelUp moves the current position one level up toward the root.
func (o *Outline) LevelUp() {
	if len(o.stack) > 0 {
		o.cur = o.stack[len(o.stack)-1]
		o.stack = o.stack[:len(o.stack)-1]
	}
}

// LevelRoot resets the current position to the root.
func (o *Outline) LevelRoot() {
	o.stack = o.stack[:0]
	o.cur = o.Root
}

// CurPosition returns the last child added at the current insertion point,
// or nil if the current node has no children.
// C# equivalent: Outline.CurPosition → curItem[curItem.Count - 1]
func (o *Outline) CurPosition() *OutlineItem {
	if len(o.cur.Children) == 0 {
		return nil
	}
	return o.cur.Children[len(o.cur.Children)-1]
}

// shiftItem recursively increments PageIdx and adds deltaY to OffsetY for
// the given item and all its descendants.
func shiftItem(item *OutlineItem, deltaY float32) {
	item.PageIdx++
	item.OffsetY += deltaY
	for _, child := range item.Children {
		shiftItem(child, deltaY)
	}
}

// Shift repositions outline items after a keep-together page break.
// Starting from the item that is the next sibling of `from` (in from's parent),
// it increments PageIdx by 1 and adjusts OffsetY by (newY - nextSibling.OffsetY)
// recursively for that sibling and all its descendants.
//
// C# equivalent: Outline.Shift(XmlItem from, float newY)
func (o *Outline) Shift(from *OutlineItem, newY float32) {
	if from == nil {
		return
	}
	// Find from's parent by searching the tree.
	parent := o.findParent(o.Root, from)
	if parent == nil {
		return
	}
	idx := -1
	for i, child := range parent.Children {
		if child == from {
			idx = i
			break
		}
	}
	if idx < 0 || idx+1 >= len(parent.Children) {
		return
	}
	next := parent.Children[idx+1]
	deltaY := newY - next.OffsetY
	shiftItem(next, deltaY)
}

// findParent searches the tree rooted at node for the parent of target.
func (o *Outline) findParent(node, target *OutlineItem) *OutlineItem {
	for _, child := range node.Children {
		if child == target {
			return node
		}
		if found := o.findParent(child, target); found != nil {
			return found
		}
	}
	return nil
}

// PrepareToFirstPass saves the current root children count for later trimming
// during double-pass report generation. Resets the cursor to root.
// C# equivalent: Outline.PrepareToFirstPass()
func (o *Outline) PrepareToFirstPass() {
	if len(o.Root.Children) == 0 {
		o.firstPassPos = -1
	} else {
		o.firstPassPos = len(o.Root.Children)
	}
	o.LevelRoot()
}

// ClearFirstPass trims Root.Children back to the position saved by
// PrepareToFirstPass, then resets the cursor to root.
// C# equivalent: Outline.ClearFirstPass() → Clear(firstPassPosition)
func (o *Outline) ClearFirstPass() {
	if o.firstPassPos == -1 {
		o.Root.Children = nil
	} else if o.firstPassPos < len(o.Root.Children) {
		o.Root.Children = o.Root.Children[:o.firstPassPos]
	}
	o.LevelRoot()
}

// Clear removes all children from the root and resets the cursor to root.
// C# equivalent: Outline.Clear() → rootItem.Clear(); LevelRoot()
func (o *Outline) Clear() {
	o.Root.Children = nil
	o.LevelRoot()
}

// IsEmpty reports whether the outline has no top-level items.
// C# equivalent: Outline.IsEmpty → rootItem.Count == 0
func (o *Outline) IsEmpty() bool {
	return len(o.Root.Children) == 0
}

// ── PreparedBand ──────────────────────────────────────────────────────────────

// PreparedBandKind classifies the source band type stored in PreparedBand.
// Used by GetLastY to exclude PageFooter and Overlay bands, mirroring the C#
// PreparedPage.GetLastY() check: !(obj is PageFooterBand) && !(obj is OverlayBand).
type PreparedBandKind int

const (
	// PreparedBandKindNormal is the default for all regular bands (data, header, footer sub-types, etc.).
	PreparedBandKindNormal PreparedBandKind = iota
	// PreparedBandKindPageFooter marks a PageFooterBand; excluded from GetLastY.
	PreparedBandKindPageFooter
	// PreparedBandKindOverlay marks an OverlayBand; excluded from GetLastY.
	PreparedBandKindOverlay
)

// PreparedBand is a rendered band snapshot stored inside a PreparedPage.
type PreparedBand struct {
	// Name is the source band's name.
	Name string
	// Kind classifies the source band type (Normal, PageFooter, Overlay).
	// Set by the engine when building the PreparedBand; used by GetLastY.
	Kind PreparedBandKind
	// Left is the X position on the page in pixels (column offset).
	Left float32
	// Top is the Y position on the page in pixels.
	Top float32
	// Height is the rendered height in pixels.
	Height float32
	// Width is the band width in pixels.
	Width float32
	// FillColor is the band's background fill color.
	// A=0 means transparent (C# outputs "background-color:transparent").
	FillColor color.RGBA
	// Border is the band's border definition.
	Border style.Border
	// Objects holds rendered object snapshots (text, images, etc.).
	Objects []PreparedObject
}

// ObjectType distinguishes the kind of report object stored in PreparedObject.
type ObjectType int

const (
	// ObjectTypeText is a TextObject or HtmlObject.
	ObjectTypeText ObjectType = iota
	// ObjectTypePicture is a PictureObject.
	ObjectTypePicture
	// ObjectTypeLine is a LineObject.
	ObjectTypeLine
	// ObjectTypeShape is a ShapeObject.
	ObjectTypeShape
	// ObjectTypeCheckBox is a CheckBoxObject.
	ObjectTypeCheckBox
	// ObjectTypeBarcode is a BarcodeObject.
	ObjectTypeBarcode
	// ObjectTypePolyLine is an open polyline (PolyLineObject).
	ObjectTypePolyLine
	// ObjectTypePolygon is a closed polygon (PolygonObject).
	ObjectTypePolygon
	// ObjectTypeSVG is a raw SVG document stored as UTF-8 bytes in BlobStore.
	// HTML exporters may emit it inline; PDF/image exporters draw a placeholder.
	ObjectTypeSVG
	// ObjectTypeDigitalSignature is a DigitalSignatureObject.
	// PDF exporters render it as a /Widget /Sig form field annotation.
	// Other exporters render a styled placeholder box.
	ObjectTypeDigitalSignature
	// ObjectTypeHtml is an HtmlObject whose Text field contains raw HTML markup.
	// HTML exporters must emit the text verbatim (not HTML-escaped).
	// Other exporters strip tags and render plain text.
	ObjectTypeHtml
	// ObjectTypeRTF is a RichObject whose Text field contains raw RTF content.
	// HTML exporters convert RTF to HTML; PDF/image exporters render plain text
	// after stripping RTF control words via utils.StripRTF.
	ObjectTypeRTF
)

// MergeMode controls how adjacent text objects with the same content are
// merged. Mirrors FastReport.MergeMode flags (bitfield: Vertical=2, Horizontal=1).
// C# source: FastReport.Base/Object/TextObject.cs MergeMode enum.
type MergeMode int

const (
	// MergeModeNone disables merging (default).
	MergeModeNone MergeMode = 0
	// MergeModeHorizontal merges horizontally adjacent objects.
	MergeModeHorizontal MergeMode = 1
	// MergeModeVertical merges vertically adjacent objects.
	MergeModeVertical MergeMode = 2
)

// DuplicatesMode controls how consecutive objects with the same name and text
// are rendered. Mirrors FastReport.Duplicates enum.
type DuplicatesMode int

const (
	// DuplicatesShow shows all duplicate values (default).
	DuplicatesShow DuplicatesMode = iota
	// DuplicatesClear keeps the first occurrence and clears text in duplicates.
	DuplicatesClear
	// DuplicatesHide keeps the first occurrence and hides (removes) duplicates.
	DuplicatesHide
	// DuplicatesMerge stretches the first occurrence to cover all duplicates.
	DuplicatesMerge
)

// PreparedObject is a rendered report component snapshot.
type PreparedObject struct {
	// Name is the component name.
	Name string
	// ObjectType distinguishes text, picture, line, shape, checkbox, etc.
	Kind ObjectType
	// Left, Top, Width, Height are position and size in pixels.
	Left, Top, Width, Height float32
	// Text is the rendered text content (for text objects).
	Text string
	// BlobIdx is the blob store index for image objects (-1 = no blob).
	BlobIdx int
	// ShapeKind identifies the geometric shape (for ObjectTypeShape).
	// 0=Rectangle, 1=RoundRectangle, 2=Ellipse, 3=Triangle, 4=Diamond.
	ShapeKind int
	// ShapeCurve is the corner radius in pixels for RoundRectangle shapes.
	ShapeCurve float32
	// LineDiagonal indicates a diagonal line (for ObjectTypeLine).
	LineDiagonal bool
	// Points holds vertex coordinates for ObjectTypePolyLine and ObjectTypePolygon.
	// Each element is [x, y] in pixels relative to the object's Left/Top origin.
	Points [][2]float32

	// ── Style fields ────────────────────────────────────────────────────────
	// Font describes the text font.
	Font style.Font
	// TextColor is the foreground colour for text.
	TextColor color.RGBA
	// FillColor is the background fill colour.
	FillColor color.RGBA
	// HorzAlign is the horizontal text alignment (0=Left,1=Center,2=Right,3=Justify).
	HorzAlign int
	// VertAlign is the vertical text alignment (0=Top,1=Center,2=Bottom).
	VertAlign int
	// Border holds the rendered border lines.
	Border style.Border
	// WordWrap indicates whether text wraps within the bounds.
	WordWrap bool
	// Checked is the checked state for ObjectTypeCheckBox.
	Checked bool
	// CheckedSymbol is the symbol drawn when checked (0=Check, 1=Cross, 2=Plus, 3=Fill).
	CheckedSymbol int
	// UncheckedSymbol is the symbol drawn when unchecked (0=None, 1=Cross, 2=Minus, 3=Slash, 4=BackSlash).
	UncheckedSymbol int
	// CheckColor is the color of the check symbol.
	CheckColor color.RGBA
	// Duplicates controls how repeated values with the same object name are handled.
	Duplicates DuplicatesMode

	// ── Barcode-specific fields (ObjectTypePicture only) ─────────────────────
	// IsBarcode indicates that this picture object was rendered from a BarcodeObject.
	// When true and BarcodeModules is non-nil, exporters may use vector rendering.
	IsBarcode bool
	// BarcodeModules is the raw module bit-matrix of the encoded barcode
	// (true = dark module). Rows are indexed [y][x]. Only set when IsBarcode is true.
	// The dimensions correspond to the minimum symbol size (1px per module);
	// exporters scale the modules to fit the object bounds.
	BarcodeModules [][]bool
	// HyperlinkKind indicates the type of hyperlink (0=None, 1=URL, 2=PageNumber, 3=Bookmark).
	HyperlinkKind int
	// HyperlinkValue is the resolved hyperlink target (URL, page number string, bookmark name).
	HyperlinkValue string
	// HyperlinkTarget is the anchor target attribute (e.g. "_blank" for new tab).
	// Corresponds to C# Hyperlink.Target / OpenLinkInNewTab.
	HyperlinkTarget string
	// Bookmark is the anchor name for this object. When non-empty the HTML exporter
	// emits <a name="..."></a> before the object div, matching C# ExportObject behaviour.
	// C# reference: HTMLExportLayers.cs ExportObject → obj.Bookmark → <a name="...">.
	Bookmark string

	// TextRenderType mirrors object.TextRenderType and controls how the Text
	// field is interpreted by exporters.
	// 0=Default (plain text), 1=HtmlTags, 2=HtmlParagraph, 3=Inline.
	TextRenderType int

	// ── Padding / layout fields ──────────────────────────────────────────────
	// PaddingLeft, PaddingTop, PaddingRight, PaddingBottom are interior spacing
	// in pixels, sourced from TextObject.Padding.
	PaddingLeft, PaddingTop, PaddingRight, PaddingBottom float32
	// ParagraphOffset is the first-line text indent in pixels.
	ParagraphOffset float32
	// LineHeight is the explicit line height in pixels (0 = use default).
	LineHeight float32
	// RTL indicates right-to-left text direction.
	RTL bool
	// Clip indicates whether the object should clip its content.
	Clip bool
	// MergeMode controls merging of adjacent objects with equal text.
	// MergeModeNone (0) means no merging (default).
	// C# source: TextObject.MergeMode (TextObject.cs)
	MergeMode MergeMode
}

// ── PreparedWatermark ─────────────────────────────────────────────────────────

// WatermarkTextRotation mirrors reportpkg.WatermarkTextRotation.
type WatermarkTextRotation int

const (
	WatermarkTextRotationHorizontal      WatermarkTextRotation = iota
	WatermarkTextRotationVertical
	WatermarkTextRotationForwardDiagonal
	WatermarkTextRotationBackwardDiagonal
)

// WatermarkImageSize mirrors reportpkg.WatermarkImageSize.
type WatermarkImageSize int

const (
	WatermarkImageSizeNormal  WatermarkImageSize = iota
	WatermarkImageSizeCenter
	WatermarkImageSizeStretch
	WatermarkImageSizeZoom
	WatermarkImageSizeTile
)

// PreparedWatermark holds the watermark properties for a single prepared page.
// It is populated by the engine from the ReportPage.Watermark and attached
// to PreparedPage so that exporters can render it without needing reportpkg.
type PreparedWatermark struct {
	Enabled bool

	// Text watermark
	Text          string
	Font          style.Font
	TextColor     color.RGBA
	TextRotation  WatermarkTextRotation
	ShowTextOnTop bool

	// Image watermark (-1 = no image)
	ImageBlobIdx     int
	ImageSize        WatermarkImageSize
	ImageTransparency float32
	ShowImageOnTop   bool
}

// ── PreparedPage ──────────────────────────────────────────────────────────────

// PreparedPage is a single rendered page in the prepared-pages collection.
type PreparedPage struct {
	// PageNo is the 1-based logical page number.
	PageNo int
	// Width / Height are the page dimensions in pixels.
	Width, Height float32
	// Bands holds the rendered bands in print order.
	Bands []*PreparedBand
	// Watermark holds the optional page watermark, or nil if none.
	Watermark *PreparedWatermark
}

// AddBand appends a rendered band to the page.
func (p *PreparedPage) AddBand(b *PreparedBand) {
	p.Bands = append(p.Bands, b)
}

// ── PreparedPages ─────────────────────────────────────────────────────────────

// PreparedPages is the collection of rendered pages produced by the engine.
// It is the Go equivalent of FastReport.Preview.PreparedPages.
type PreparedPages struct {
	pages         []*PreparedPage
	curPage       int
	addPageAction AddPageAction
	Bookmarks     *Bookmarks
	Outline       *Outline
	BlobStore     *BlobStore

	// pageCache is the LRU page cache. It is invalidated on structural
	// changes (ModifyPage, RemovePage, CopyPage, ApplyWatermark) so that
	// callers using GetCachedPage always see up-to-date data.
	// C# equivalent: PreparedPages.pageCache field.
	pageCache *PageCache

	// cut bands held between CutObjects and PasteObjects (keep-together).
	cutBands []*PreparedBand

	// firstPassPage and firstPassPosition record the page index and band
	// count at the start of the current double-pass cycle, used by
	// ClearFirstPass to restore the collection to that checkpoint.
	// C# equivalent: firstPassPage / firstPassPosition fields.
	firstPassPage     int
	firstPassPosition int
}

// New creates an empty PreparedPages collection.
func New() *PreparedPages {
	pp := &PreparedPages{
		curPage:       -1,
		addPageAction: AddPageActionAdd,
		Bookmarks:     NewBookmarks(),
		Outline:       NewOutline(),
		BlobStore:     NewBlobStore(),
	}
	// pageCache is initialised after pp is constructed so NewPageCache can
	// hold a back-reference to pp (C# equivalent: pageCache = new PageCache(this)).
	pp.pageCache = NewPageCache(pp, 0)
	return pp
}

// Count returns the number of prepared pages.
func (pp *PreparedPages) Count() int { return len(pp.pages) }

// CurPage returns the zero-based index of the current page.
func (pp *PreparedPages) CurPage() int { return pp.curPage }

// AddPageAction returns the current add-page behaviour.
func (pp *PreparedPages) AddPageAction() AddPageAction { return pp.addPageAction }

// SetAddPageAction sets the add-page behaviour.
func (pp *PreparedPages) SetAddPageAction(a AddPageAction) { pp.addPageAction = a }

// AddPage adds (or rewrites) a page for the given page dimensions.
func (pp *PreparedPages) AddPage(width, height float32, pageNo int) {
	if pp.addPageAction == AddPageActionWriteOver &&
		pp.curPage >= 0 && pp.curPage < len(pp.pages) {
		// Reuse current slot (double-pass rewrite).
		pp.pages[pp.curPage] = &PreparedPage{PageNo: pageNo, Width: width, Height: height}
	} else {
		pp.pages = append(pp.pages, &PreparedPage{PageNo: pageNo, Width: width, Height: height})
		pp.curPage = len(pp.pages) - 1
	}
}

// NextPage advances the curPage pointer.
func (pp *PreparedPages) NextPage() {
	pp.curPage++
}

// GetPage returns the PreparedPage at index, or nil if out of range.
func (pp *PreparedPages) GetPage(index int) *PreparedPage {
	if index < 0 || index >= len(pp.pages) {
		return nil
	}
	return pp.pages[index]
}

// CurrentPage returns the current PreparedPage, or nil if no page has been added.
func (pp *PreparedPages) CurrentPage() *PreparedPage {
	return pp.GetPage(pp.curPage)
}

// AddBand appends a rendered band to the current page.
// Returns an error if no page has been started.
func (pp *PreparedPages) AddBand(b *PreparedBand) error {
	pg := pp.CurrentPage()
	if pg == nil {
		return fmt.Errorf("PreparedPages: no current page")
	}
	pg.AddBand(b)
	return nil
}

// AddObject appends a PreparedObject to an existing PreparedBand.
func (pp *PreparedPages) AddObject(b *PreparedBand, obj PreparedObject) error {
	if b == nil {
		return fmt.Errorf("PreparedPages: nil band")
	}
	b.Objects = append(b.Objects, obj)
	return nil
}

// CurPosition returns the number of bands on the current page.
// Used by the keep-together mechanism to mark a save point.
func (pp *PreparedPages) CurPosition() int {
	pg := pp.CurrentPage()
	if pg == nil {
		return 0
	}
	return len(pg.Bands)
}

// CutObjects removes all bands from position onwards from the current page
// and stores them for later pasting. Any previously stored cut bands are discarded.
func (pp *PreparedPages) CutObjects(position int) {
	pg := pp.CurrentPage()
	if pg == nil {
		pp.cutBands = nil
		return
	}
	if position < 0 || position >= len(pg.Bands) {
		pp.cutBands = nil
		return
	}
	pp.cutBands = make([]*PreparedBand, len(pg.Bands)-position)
	copy(pp.cutBands, pg.Bands[position:])
	pg.Bands = pg.Bands[:position]
}

// PasteObjects adds the previously cut bands to the current page, offsetting
// each band's Top by dy. dx is reserved for column support (unused here).
func (pp *PreparedPages) PasteObjects(dx, dy float32) {
	pg := pp.CurrentPage()
	if pg == nil || len(pp.cutBands) == 0 {
		pp.cutBands = nil
		return
	}
	for _, b := range pp.cutBands {
		copy := *b
		copy.Top += dy
		pg.Bands = append(pg.Bands, &copy)
	}
	pp.cutBands = nil
}

// CutBands returns the bands currently held in the cut buffer (from CutObjects).
func (pp *PreparedPages) CutBands() []*PreparedBand {
	return pp.cutBands
}

// RemoveLast removes the last page from the collection.
func (pp *PreparedPages) RemoveLast() {
	if len(pp.pages) == 0 {
		return
	}
	pp.pages = pp.pages[:len(pp.pages)-1]
	if pp.curPage >= len(pp.pages) {
		pp.curPage = len(pp.pages) - 1
	}
}

// TrimTo keeps only the first n pages, discarding the rest.
func (pp *PreparedPages) TrimTo(n int) {
	if n < 0 {
		n = 0
	}
	if n >= len(pp.pages) {
		return
	}
	pp.pages = pp.pages[:n]
	if pp.curPage >= n {
		pp.curPage = n - 1
	}
}

// Clear removes all pages, bookmarks, outline items, and blobs, and invalidates
// the page cache. This mirrors C# PreparedPages.Clear() which calls
// sourcePages.Clear(), pageCache.Clear(), bookmarks.Clear(), outline.Clear(),
// blobStore.Clear(), and resets curPage to 0 (Go uses -1 for "no current page").
func (pp *PreparedPages) Clear() {
	pp.pages = pp.pages[:0]
	pp.curPage = -1
	pp.pageCache.Clear()
	pp.Bookmarks.Clear()
	pp.Outline.Clear()
	pp.BlobStore.Clear()
}

// PrepareToFirstPass records the current page and band-count checkpoint for
// use by ClearFirstPass during double-pass report generation. It also calls
// Outline.PrepareToFirstPass() to save the outline cursor position.
// C# equivalent: PreparedPages.PrepareToFirstPass()
func (pp *PreparedPages) PrepareToFirstPass() {
	pp.firstPassPage = pp.curPage
	pp.firstPassPosition = pp.CurPosition()
	pp.Outline.PrepareToFirstPass()
}

// ClearFirstPass rolls back the prepared pages to the checkpoint saved by
// PrepareToFirstPass. It removes all pages after firstPassPage, trims
// firstPassPage's bands back to firstPassPosition, and calls
// Bookmarks.ClearFirstPass() and Outline.ClearFirstPass().
// C# equivalent: PreparedPages.ClearFirstPass()
func (pp *PreparedPages) ClearFirstPass() {
	pp.Bookmarks.ClearFirstPass()
	pp.Outline.ClearFirstPass()

	// Remove all pages after firstPassPage.
	for pp.firstPassPage < len(pp.pages)-1 {
		pp.pages = pp.pages[:len(pp.pages)-1]
	}

	// If position is at beginning (page 0, position 0), remove page 0 entirely.
	if pp.firstPassPage == 0 && pp.firstPassPosition == 0 {
		if len(pp.pages) > 0 {
			pp.pages = pp.pages[:0]
		}
	}

	// Delete bands on firstPassPage beyond firstPassPosition.
	if pp.firstPassPage >= 0 && pp.firstPassPage < len(pp.pages) {
		pg := pp.pages[pp.firstPassPage]
		if pp.firstPassPosition < len(pg.Bands) {
			pg.Bands = pg.Bands[:pp.firstPassPosition]
		}
	}

	pp.curPage = pp.firstPassPage
	if pp.curPage >= len(pp.pages) {
		pp.curPage = len(pp.pages) - 1
	}
	pp.pageCache.Clear()
}

// GetLastY returns the maximum (Top + Height) value among all non-PageFooter,
// non-Overlay bands on the current page. Returns 0 if there is no current page.
// C# equivalent: PreparedPages.GetLastY() → preparedPages[CurPage].GetLastY()
func (pp *PreparedPages) GetLastY() float32 {
	pg := pp.CurrentPage()
	if pg == nil {
		return 0
	}
	var result float32
	for _, b := range pg.Bands {
		if b.Kind == PreparedBandKindPageFooter || b.Kind == PreparedBandKindOverlay {
			continue
		}
		if bottom := b.Top + b.Height; bottom > result {
			result = bottom
		}
	}
	return result
}

// ContainsBand reports whether any band on the current page has the given name.
// C# equivalent: PreparedPages.ContainsBand(string bandName)
func (pp *PreparedPages) ContainsBand(name string) bool {
	pg := pp.CurrentPage()
	if pg == nil {
		return false
	}
	for _, b := range pg.Bands {
		if b.Name == name {
			return true
		}
	}
	return false
}

// AddBookmark adds a navigation bookmark.
func (pp *PreparedPages) AddBookmark(name string, offsetY float32) {
	pp.Bookmarks.Add(&Bookmark{
		Name:    name,
		PageIdx: pp.curPage,
		OffsetY: offsetY,
	})
}

// GetCachedPage returns the PreparedPage at index via the LRU page cache.
// Frequently accessed pages are served from memory; less-recently-used pages
// are evicted when the cache reaches its capacity limit (default 50).
// Returns nil if index is out of range.
// C# equivalent: PreparedPages.GetCachedPage(int index) → pageCache.Get(index)
func (pp *PreparedPages) GetCachedPage(index int) *PreparedPage {
	return pp.pageCache.Get(index)
}

// ClearPageCache evicts all pages from the LRU cache. Called after structural
// changes (RemovePage, CopyPage, ApplyWatermark) that invalidate cached data.
// C# equivalent: PreparedPages.ClearPageCache() → pageCache.Clear()
func (pp *PreparedPages) ClearPageCache() {
	pp.pageCache.Clear()
}

// RemovePageCache evicts the entry for index from the LRU cache. Called after
// ModifyPage so that the next GetCachedPage call returns fresh data.
// C# equivalent: PreparedPages.RemovePageCache(int index) → pageCache.Remove(index)
func (pp *PreparedPages) RemovePageCache(index int) {
	pp.pageCache.Remove(index)
}
