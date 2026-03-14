// Example: xml_datasource demonstrates using the XML data source.
// Run with:
//
//	go run ./examples/xml_datasource/
package main

import (
	"fmt"
	"os"

	"github.com/andrewloable/go-fastreport/band"
	xmldata "github.com/andrewloable/go-fastreport/data/xml"
	"github.com/andrewloable/go-fastreport/engine"
	"github.com/andrewloable/go-fastreport/export/html"
	"github.com/andrewloable/go-fastreport/reportpkg"
)

const catalogXML = `<?xml version="1.0" encoding="utf-8"?>
<Catalog>
  <Book ISBN="978-0-13-468599-1" Title="The Go Programming Language"   Author="Donovan &amp; Kernighan" Year="2015"/>
  <Book ISBN="978-1-491-98124-2" Title="Concurrency in Go"             Author="Katherine Cox-Buday"     Year="2017"/>
  <Book ISBN="978-0-13-110362-7" Title="The C Programming Language"    Author="Kernighan &amp; Ritchie"  Year="1988"/>
  <Book ISBN="978-1-593-27584-6" Title="The Linux Command Line"        Author="William Shotts"          Year="2012"/>
</Catalog>`

func main() {
	// 1. Load and initialise the XML data source.
	// XMLDataSource embeds data.BaseDataSource so it satisfies band.DataSource
	// directly — no adapter wrapper needed.
	ds := xmldata.New("books")
	ds.SetXML(catalogXML)
	// The root element is <Catalog>; row elements are <Book>.
	// No rootPath needed since <Book> elements are direct children.
	if err := ds.Init(); err != nil {
		fmt.Fprintf(os.Stderr, "data source error: %v\n", err)
		os.Exit(1)
	}
	fmt.Fprintf(os.Stderr, "Loaded %d books from XML\n", ds.RowCount())

	// 2. Build report.
	r := reportpkg.NewReport()
	r.Info.Name = "Book Catalog"

	pg := reportpkg.NewReportPage()
	r.AddPage(pg)

	title := band.NewReportTitleBand()
	title.SetName("ReportTitle")
	title.SetHeight(50)
	title.SetVisible(true)
	pg.SetReportTitle(title)

	db := band.NewDataBand()
	db.SetName("BookBand")
	db.SetHeight(20)
	db.SetVisible(true)
	db.SetDataSource(ds)
	pg.AddBand(db)

	// 3. Run engine.
	e := engine.New(r)
	if err := e.Run(engine.DefaultRunOptions()); err != nil {
		fmt.Fprintf(os.Stderr, "engine error: %v\n", err)
		os.Exit(1)
	}

	// 4. Export to HTML.
	exp := html.NewExporter()
	exp.Title = "Book Catalog"
	exp.EmbedCSS = true
	if err := exp.Export(e.PreparedPages(), os.Stdout); err != nil {
		fmt.Fprintf(os.Stderr, "export error: %v\n", err)
		os.Exit(1)
	}
}
