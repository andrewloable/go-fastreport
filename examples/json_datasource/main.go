// Example: json_datasource demonstrates using the JSON data source to bind
// report data from a JSON string. Run with:
//
//	go run ./examples/json_datasource/
package main

import (
	"fmt"
	"os"

	"github.com/andrewloable/go-fastreport/band"
	jsondata "github.com/andrewloable/go-fastreport/data/json"
	"github.com/andrewloable/go-fastreport/engine"
	"github.com/andrewloable/go-fastreport/export/html"
	"github.com/andrewloable/go-fastreport/reportpkg"
)

const sampleJSON = `[
  {"Name": "Alice",   "Department": "Engineering", "Salary": 95000},
  {"Name": "Bob",     "Department": "Marketing",   "Salary": 72000},
  {"Name": "Carol",   "Department": "Engineering", "Salary": 105000},
  {"Name": "Dave",    "Department": "HR",          "Salary": 65000},
  {"Name": "Eve",     "Department": "Engineering", "Salary": 88000}
]`

func main() {
	// 1. Load and initialise the JSON data source.
	// JSONDataSource embeds data.BaseDataSource so it satisfies band.DataSource
	// directly — no adapter wrapper needed.
	ds := jsondata.New("employees")
	ds.SetJSON(sampleJSON)
	if err := ds.Init(); err != nil {
		fmt.Fprintf(os.Stderr, "data source error: %v\n", err)
		os.Exit(1)
	}
	fmt.Fprintf(os.Stderr, "Loaded %d employees from JSON\n", ds.RowCount())

	// 2. Build report.
	r := reportpkg.NewReport()
	r.Info.Name = "Employee Directory"

	pg := reportpkg.NewReportPage()
	r.AddPage(pg)

	phdr := band.NewPageHeaderBand()
	phdr.SetName("PageHeader")
	phdr.SetHeight(40)
	phdr.SetVisible(true)
	pg.SetPageHeader(phdr)

	db := band.NewDataBand()
	db.SetName("EmployeeBand")
	db.SetHeight(20)
	db.SetVisible(true)
	db.SetDataSource(ds)
	pg.AddBand(db)

	pftr := band.NewPageFooterBand()
	pftr.SetName("PageFooter")
	pftr.SetHeight(25)
	pftr.SetVisible(true)
	pg.SetPageFooter(pftr)

	// 3. Run engine.
	e := engine.New(r)
	if err := e.Run(engine.DefaultRunOptions()); err != nil {
		fmt.Fprintf(os.Stderr, "engine error: %v\n", err)
		os.Exit(1)
	}

	// 4. Export.
	exp := html.NewExporter()
	exp.Title = "Employee Directory"
	if err := exp.Export(e.PreparedPages(), os.Stdout); err != nil {
		fmt.Fprintf(os.Stderr, "export error: %v\n", err)
		os.Exit(1)
	}
}
