package pdf

import (
	"github.com/andrewloable/go-fastreport/export/pdf/core"
)

// Catalog represents the PDF document catalog (root object).
// It holds the /Type /Catalog dictionary and a reference to the page tree.
type Catalog struct {
	obj   *core.IndirectObject
	dict  *core.Dictionary
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
	dict.Add("Pages", core.NewRef(pages.obj))

	obj := w.NewObject(dict)
	w.setCatalog(obj)

	return &Catalog{obj: obj, dict: dict, pages: pages}
}

// SetOutlines registers the outline root object with the catalog and sets
// /PageMode to /UseOutlines so that PDF viewers open the bookmarks panel
// automatically when the document contains an outline tree.
func (c *Catalog) SetOutlines(outlineRef *core.IndirectObject) {
	c.dict.Add("Outlines", core.NewRef(outlineRef))
	c.dict.Add("PageMode", core.NewName("UseOutlines"))
}

// SetNamedDests registers the /Names /Dests name tree with the catalog.
func (c *Catalog) SetNamedDests(namesRef *core.IndirectObject) {
	names := core.NewDictionary()
	names.Add("Dests", core.NewRef(namesRef))
	c.dict.Add("Names", names)
}

// SetAcroForm registers an /AcroForm dictionary with the catalog.
// Required for PDF interactive form fields (including digital signatures).
func (c *Catalog) SetAcroForm(acroForm *core.Dictionary) {
	c.dict.Add("AcroForm", acroForm)
}

// Info holds PDF document metadata stored in the document information dictionary.
// Matches C# PdfInfo (PDFSimpleExport.Config.cs / PdfInfo.cs).
type Info struct {
	obj      *core.IndirectObject
	dict     *core.Dictionary
	Title    string
	Author   string
	Subject  string
	Keywords string
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

// SetSubject sets the /Subject entry in the info dictionary.
// Matches C# PdfInfo.Subject (PDFSimpleExport.Config.cs).
func (info *Info) SetSubject(subject string) {
	info.Subject = subject
	if subject == "" {
		return
	}
	info.dict.Add("Subject", core.NewHexString(subject))
}

// SetKeywords sets the /Keywords entry in the info dictionary.
// Matches C# PdfInfo.Keywords (PDFSimpleExport.Config.cs).
func (info *Info) SetKeywords(keywords string) {
	info.Keywords = keywords
	if keywords == "" {
		return
	}
	info.dict.Add("Keywords", core.NewHexString(keywords))
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
