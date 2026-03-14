// Example: frx_json demonstrates loading a report from an FRX file, setting
// report variables (parameters), and binding a JSON data source at runtime.
//
// Run with:
//
//	go run ./examples/frx_json/
package main

import (
	"fmt"
	"os"

	"github.com/andrewloable/go-fastreport/band"
	"github.com/andrewloable/go-fastreport/data"
	jsondata "github.com/andrewloable/go-fastreport/data/json"
	"github.com/andrewloable/go-fastreport/engine"
	"github.com/andrewloable/go-fastreport/export/html"
	"github.com/andrewloable/go-fastreport/reportpkg"
)

const employeeJSON = `[
  {"Name": "Alice",   "Department": "Engineering", "Salary": 95000},
  {"Name": "Bob",     "Department": "Marketing",   "Salary": 72000},
  {"Name": "Carol",   "Department": "Engineering", "Salary": 105000},
  {"Name": "Dave",    "Department": "HR",          "Salary": 65000},
  {"Name": "Eve",     "Department": "Engineering", "Salary": 88000}
]`

func main() {
	// 1. Load the FRX report definition.
	r := reportpkg.NewReport()
	if err := r.Load("examples/frx_json/report.frx"); err != nil {
		fmt.Fprintf(os.Stderr, "load FRX: %v\n", err)
		os.Exit(1)
	}
	fmt.Fprintf(os.Stderr, "Loaded report %q (%d page(s))\n", r.Info.Name, len(r.Pages()))

	// 2. Set report variables (parameters).
	// Parameters are referenced in the FRX with bracket syntax: [ParamName].
	// They appear in text objects just like data source fields.
	dict := data.NewDictionary()
	dict.AddParameter(&data.Parameter{Name: "ReportTitle", Value: "Employee Directory"})
	dict.AddParameter(&data.Parameter{Name: "FilterDept", Value: "All Departments"})
	r.SetDictionary(dict)

	// 3. Create and initialise the JSON data source.
	// JSONDataSource embeds data.BaseDataSource so it satisfies both
	// band.DataSource (for iteration) and data.DataSource (for expression
	// evaluation — the engine calls SetCalcContext(ds) per row so that
	// [Name], [Department], [Salary] resolve to the current row's values).
	ds := jsondata.New("employees")
	ds.SetJSON(employeeJSON)
	if err := ds.Init(); err != nil {
		fmt.Fprintf(os.Stderr, "data source init: %v\n", err)
		os.Exit(1)
	}
	fmt.Fprintf(os.Stderr, "JSON data source: %d rows\n", ds.RowCount())

	// 4. Bind the data source to every DataBand in the report.
	for _, pg := range r.Pages() {
		for _, b := range pg.Bands() {
			if db, ok := b.(*band.DataBand); ok {
				db.SetDataSource(ds)
			}
		}
	}

	// 5. Run the engine.
	e := engine.New(r)
	if err := e.Run(engine.DefaultRunOptions()); err != nil {
		fmt.Fprintf(os.Stderr, "engine run: %v\n", err)
		os.Exit(1)
	}

	// 6. Export to HTML (stdout).
	exp := html.NewExporter()
	exp.Title = r.Info.Name
	if err := exp.Export(e.PreparedPages(), os.Stdout); err != nil {
		fmt.Fprintf(os.Stderr, "export: %v\n", err)
		os.Exit(1)
	}
}
