package pdf

import (
	"github.com/andrewloable/go-fastreport/export/pdf/core"
)

// Catalog represents the PDF document catalog (root object).
// It holds the /Type /Catalog dictionary and a reference to the page tree.
type Catalog struct {
	obj   *core.IndirectObject
	pages *Pages
}

// NewCatalog creates a Catalog, registers it with the writer, and links the
// given Pages object.  The catalog is automatically marked as the document
// Root so the writer can include it in the trailer.
func NewCatalog(w *Writer, pages *Pages) *Catalog {
	dict := core.NewDictionary()
	dict.Add("Type", core.NewName("Catalog"))
	dict.Add("Version", core.NewName("1.5"))

	markInfo := core.NewDictionary()
	markInfo.Add("Marked", core.NewBoolean(true))
	dict.Add("MarkInfo", markInfo)

	// Link to the Pages tree using an indirect reference
	dict.Add("Pages", core.NewName(pages.obj.Reference()))

	obj := w.NewObject(dict)
	w.setCatalog(obj)

	return &Catalog{obj: obj, pages: pages}
}

// Info holds PDF document metadata stored in the document information dictionary.
type Info struct {
	obj      *core.IndirectObject
	dict     *core.Dictionary
	Title    string
	Author   string
	Creator  string
	Producer string
}

// NewInfo creates an Info dictionary, registers it with the writer, and marks
// it as the document Info entry in the trailer.
func NewInfo(w *Writer) *Info {
	dict := core.NewDictionary()
	dict.Add("Creator", core.NewHexString("go-fastreport"))
	dict.Add("Producer", core.NewHexString("go-fastreport"))

	obj := w.NewObject(dict)
	w.setInfo(obj)

	return &Info{
		obj:      obj,
		dict:     dict,
		Creator:  "go-fastreport",
		Producer: "go-fastreport",
	}
}

// SetTitle sets the /Title entry in the info dictionary.
func (info *Info) SetTitle(title string) {
	info.Title = title
	if title == "" {
		return
	}
	info.dict.Add("Title", core.NewHexString(title))
}

// SetAuthor sets the /Author entry in the info dictionary.
func (info *Info) SetAuthor(author string) {
	info.Author = author
	if author == "" {
		return
	}
	info.dict.Add("Author", core.NewHexString(author))
}

// SetCreator sets the /Creator entry in the info dictionary.
func (info *Info) SetCreator(creator string) {
	info.Creator = creator
	info.dict.Add("Creator", core.NewHexString(creator))
}

// SetProducer sets the /Producer entry in the info dictionary.
func (info *Info) SetProducer(producer string) {
	info.Producer = producer
	info.dict.Add("Producer", core.NewHexString(producer))
}
