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

// ── BlobStore ─────────────────────────────────────────────────────────────────

// BlobStore stores binary blobs (e.g. images) referenced by prepared pages.
type BlobStore struct {
	blobs [][]byte
	index map[string]int // name → index
}

// NewBlobStore creates an empty BlobStore.
func NewBlobStore() *BlobStore {
	return &BlobStore{index: make(map[string]int)}
}

// Add stores blob data under name and returns its integer index.
func (b *BlobStore) Add(name string, data []byte) int {
	if idx, ok := b.index[name]; ok {
		return idx
	}
	idx := len(b.blobs)
	b.blobs = append(b.blobs, data)
	b.index[name] = idx
	return idx
}

// Get returns the blob at the given index, or nil if out of range.
func (b *BlobStore) Get(idx int) []byte {
	if idx < 0 || idx >= len(b.blobs) {
		return nil
	}
	return b.blobs[idx]
}

// Count returns the number of stored blobs.
func (b *BlobStore) Count() int { return len(b.blobs) }

// ── Bookmark ──────────────────────────────────────────────────────────────────

// Bookmark is a named navigation point within the prepared pages.
type Bookmark struct {
	Name    string
	PageIdx int // zero-based page index
	OffsetY float32
}

// Bookmarks is a collection of Bookmark entries.
type Bookmarks struct {
	items []*Bookmark
	index map[string]int // name → slice index
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

// GetPageNo returns the 1-based page number for the named bookmark, or 0 if not found.
func (bk *Bookmarks) GetPageNo(name string) int {
	b := bk.Find(name)
	if b == nil {
		return 0
	}
	return b.PageIdx + 1
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
	Root  *OutlineItem
	cur   *OutlineItem   // current insertion point
	stack []*OutlineItem // ancestors of cur
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

// ── PreparedBand ──────────────────────────────────────────────────────────────

// PreparedBand is a rendered band snapshot stored inside a PreparedPage.
type PreparedBand struct {
	// Name is the source band's name.
	Name string
	// Top is the Y position on the page in pixels.
	Top float32
	// Height is the rendered height in pixels.
	Height float32
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

	// cut bands held between CutObjects and PasteObjects (keep-together).
	cutBands []*PreparedBand
}

// New creates an empty PreparedPages collection.
func New() *PreparedPages {
	return &PreparedPages{
		curPage:       -1,
		addPageAction: AddPageActionAdd,
		Bookmarks:     NewBookmarks(),
		Outline:       NewOutline(),
		BlobStore:     NewBlobStore(),
	}
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

// Clear removes all pages.
func (pp *PreparedPages) Clear() {
	pp.pages = pp.pages[:0]
	pp.curPage = -1
}

// AddBookmark adds a navigation bookmark.
func (pp *PreparedPages) AddBookmark(name string, offsetY float32) {
	pp.Bookmarks.Add(&Bookmark{
		Name:    name,
		PageIdx: pp.curPage,
		OffsetY: offsetY,
	})
}
