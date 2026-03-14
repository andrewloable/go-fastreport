package pdf

import (
	"github.com/andrewloable/go-fastreport/export/pdf/core"
)

// Pages is the PDF page tree root (/Type /Pages).
// It maintains the ordered list of pages and is registered as an indirect
// object so individual Page objects can reference it as their /Parent.
type Pages struct {
	obj      *core.IndirectObject
	dict     *core.Dictionary
	kids     *core.Array
	pageList []*Page
}

// NewPages creates the page tree root and registers it with the writer.
func NewPages(w *Writer) *Pages {
	kids := core.NewArray()
	dict := core.NewDictionary()
	dict.Add("Type", core.NewName("Pages"))
	dict.Add("Kids", kids)
	dict.Add("Count", core.NewInt(0))

	obj := w.NewObject(dict)
	return &Pages{obj: obj, dict: dict, kids: kids}
}

// AddPage appends a page to the tree and updates the /Count entry.
func (p *Pages) AddPage(page *Page) {
	p.pageList = append(p.pageList, page)
	p.kids.Add(core.NewName(page.obj.Reference()))
	p.dict.Add("Count", core.NewInt(len(p.pageList)))
}

// Count returns the number of pages in the tree.
func (p *Pages) Count() int { return len(p.pageList) }

// Page represents a single PDF page (/Type /Page).
// Width and Height are in PDF user units (points, 1/72 inch).
type Page struct {
	obj      *core.IndirectObject
	contents *Contents
	xObjects *core.Dictionary // /Resources /XObject sub-dictionary
	Width    float64
	Height   float64
}

// NewPage creates a new page, links it to the given Pages tree, and registers
// an empty Contents stream with the writer.
func NewPage(w *Writer, pages *Pages, width, height float64) *Page {
	contents := NewContents(w)

	// Build the MediaBox array: [ 0 0 width height ]
	mediaBox := core.NewArray(
		core.NewInt(0),
		core.NewInt(0),
		core.NewFloat(width),
		core.NewFloat(height),
	)

	// Resources dictionary with common ProcSet
	resources := core.NewDictionary()
	xObject := core.NewDictionary()
	resources.Add("XObject", xObject)
	resources.Add("ProcSet", core.NewArray(
		core.NewName("PDF"),
		core.NewName("Text"),
		core.NewName("ImageC"),
	))

	dict := core.NewDictionary()
	dict.Add("Type", core.NewName("Page"))
	dict.Add("Parent", core.NewName(pages.obj.Reference()))
	dict.Add("MediaBox", mediaBox)
	dict.Add("Resources", resources)
	dict.Add("Contents", core.NewName(contents.obj.Reference()))

	obj := w.NewObject(dict)
	page := &Page{
		obj:      obj,
		contents: contents,
		xObjects: xObject,
		Width:    width,
		Height:   height,
	}

	pages.AddPage(page)
	return page
}

// Contents returns the page's content stream object.
func (p *Page) Contents() *Contents { return p.contents }

// AddXObject registers an indirect object as an XObject resource under the
// given name (e.g. "Im0").  The name is used in content streams as /Im0.
func (p *Page) AddXObject(name string, obj *core.IndirectObject) {
	p.xObjects.Add(name, core.NewName(obj.Reference()))
}
