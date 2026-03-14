// Example: json_datasource demonstrates using the JSON data source to bind
// report data from a JSON string. Run with:
//
//	go run ./examples/json_datasource/
package main

import (
	"fmt"
	"os"

	"github.com/andrewloable/go-fastreport/band"
	"github.com/andrewloable/go-fastreport/engine"
	"github.com/andrewloable/go-fastreport/export/html"
	jsondata "github.com/andrewloable/go-fastreport/data/json"
	"github.com/andrewloable/go-fastreport/reportpkg"
)

const sampleJSON = `[
  {"Name": "Alice",   "Department": "Engineering", "Salary": 95000},
  {"Name": "Bob",     "Department": "Marketing",   "Salary": 72000},
  {"Name": "Carol",   "Department": "Engineering", "Salary": 105000},
  {"Name": "Dave",    "Department": "HR",          "Salary": 65000},
  {"Name": "Eve",     "Department": "Engineering", "Salary": 88000}
]`

// jsonBandDS adapts the JSON data source to the band.DataSource interface.
type jsonBandDS struct {
	ds *jsondata.JSONDataSource
}

func (d *jsonBandDS) RowCount() int { return d.ds.RowCount() }
func (d *jsonBandDS) First() error  { return d.ds.First() }
func (d *jsonBandDS) Next() error   { return d.ds.Next() }
func (d *jsonBandDS) EOF() bool     { return d.ds.EOF() }
func (d *jsonBandDS) GetValue(col string) (any, error) { return d.ds.GetValue(col) }

func main() {
	// 1. Load and initialise the JSON data source.
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
	db.SetDataSource(&jsonBandDS{ds: ds})
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
