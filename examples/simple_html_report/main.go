// Example: simple_html_report demonstrates building a list report in memory
// and exporting it to HTML. Run with:
//
//	go run ./examples/simple_html_report/
package main

import (
	"fmt"
	"os"

	"github.com/andrewloable/go-fastreport/band"
	"github.com/andrewloable/go-fastreport/engine"
	"github.com/andrewloable/go-fastreport/export/html"
	"github.com/andrewloable/go-fastreport/reportpkg"
)

// ── in-memory data source ─────────────────────────────────────────────────────

type Product struct {
	Name  string
	Price float64
	Stock int
}

type productDS struct {
	rows []Product
	pos  int
}

func (d *productDS) RowCount() int { return len(d.rows) }
func (d *productDS) First() error  { d.pos = 0; return nil }
func (d *productDS) Next() error   { d.pos++; return nil }
func (d *productDS) EOF() bool     { return d.pos >= len(d.rows) }
func (d *productDS) GetValue(col string) (any, error) {
	if d.pos < 0 || d.pos >= len(d.rows) {
		return nil, nil
	}
	row := d.rows[d.pos]
	switch col {
	case "Name":
		return row.Name, nil
	case "Price":
		return row.Price, nil
	case "Stock":
		return row.Stock, nil
	}
	return nil, nil
}

// ── main ──────────────────────────────────────────────────────────────────────

func main() {
	// Sample product catalog.
	products := &productDS{
		rows: []Product{
			{Name: "Apple", Price: 0.99, Stock: 150},
			{Name: "Banana", Price: 0.49, Stock: 80},
			{Name: "Cherry", Price: 2.49, Stock: 40},
			{Name: "Date", Price: 3.99, Stock: 20},
			{Name: "Elderberry", Price: 5.99, Stock: 10},
		},
	}

	// 1. Build the report definition.
	r := reportpkg.NewReport()
	r.Info.Name = "Product Catalog"
	r.Info.Author = "go-fastreport"

	pg := reportpkg.NewReportPage()
	pg.SetName("Page1")
	r.AddPage(pg)

	// Report title band.
	title := band.NewReportTitleBand()
	title.SetName("ReportTitle")
	title.SetHeight(50)
	title.SetVisible(true)
	pg.SetReportTitle(title)

	// Page header.
	phdr := band.NewPageHeaderBand()
	phdr.SetName("PageHeader")
	phdr.SetHeight(30)
	phdr.SetVisible(true)
	pg.SetPageHeader(phdr)

	// Data header.
	dhdr := band.NewDataHeaderBand()
	dhdr.SetName("DataHeader")
	dhdr.SetHeight(25)
	dhdr.SetVisible(true)

	// Data band.
	db := band.NewDataBand()
	db.SetName("DataBand")
	db.SetHeight(20)
	db.SetVisible(true)
	db.SetHeader(dhdr)
	db.SetDataSource(products)
	pg.AddBand(db)

	// Report summary.
	summary := band.NewReportSummaryBand()
	summary.SetName("ReportSummary")
	summary.SetHeight(40)
	summary.SetVisible(true)
	pg.SetReportSummary(summary)

	// Page footer.
	pftr := band.NewPageFooterBand()
	pftr.SetName("PageFooter")
	pftr.SetHeight(25)
	pftr.SetVisible(true)
	pg.SetPageFooter(pftr)

	// 2. Run the report engine.
	e := engine.New(r)
	if err := e.Run(engine.DefaultRunOptions()); err != nil {
		fmt.Fprintf(os.Stderr, "engine error: %v\n", err)
		os.Exit(1)
	}

	pp := e.PreparedPages()
	fmt.Fprintf(os.Stderr, "Rendered %d page(s)\n", pp.Count())

	// 3. Export to HTML on stdout.
	exp := html.NewExporter()
	exp.Title = "Product Catalog"
	exp.EmbedCSS = true

	if err := exp.Export(pp, os.Stdout); err != nil {
		fmt.Fprintf(os.Stderr, "export error: %v\n", err)
		os.Exit(1)
	}
}
