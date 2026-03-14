// Example: frx_json demonstrates loading a report from an FRX file, setting
// report variables (parameters), and registering a JSON data source in the
// Dictionary so the engine resolves it automatically — matching the FastReport
// .NET run-time binding model.
//
// Run with:
//
//	go run ./examples/frx_json/
package main

import (
	"fmt"
	"os"

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
	// The FRX DataBand element carries DataSource="employees" — the engine will
	// resolve that alias from the Dictionary at run time.
	r := reportpkg.NewReport()
	if err := r.Load("examples/frx_json/report.frx"); err != nil {
		fmt.Fprintf(os.Stderr, "load FRX: %v\n", err)
		os.Exit(1)
	}
	fmt.Fprintf(os.Stderr, "Loaded report %q (%d page(s))\n", r.Info.Name, len(r.Pages()))

	// 2. Populate the Dictionary with parameters and the data source.
	// Parameters resolve [ReportTitle] / [FilterDept] bracket expressions.
	// The data source is looked up by alias when the engine encounters a
	// DataBand whose DataSource attribute matches — no manual band traversal needed.
	ds := jsondata.New("employees") // alias matches DataSource="employees" in FRX
	ds.SetJSON(employeeJSON)
	if err := ds.Init(); err != nil {
		fmt.Fprintf(os.Stderr, "data source init: %v\n", err)
		os.Exit(1)
	}
	fmt.Fprintf(os.Stderr, "JSON data source: %d rows\n", ds.RowCount())

	dict := r.Dictionary()
	dict.AddParameter(&data.Parameter{Name: "ReportTitle", Value: "Employee Directory"})
	dict.AddParameter(&data.Parameter{Name: "FilterDept", Value: "All Departments"})
	dict.AddDataSource(ds)

	// 3. Run the engine.
	e := engine.New(r)
	if err := e.Run(engine.DefaultRunOptions()); err != nil {
		fmt.Fprintf(os.Stderr, "engine run: %v\n", err)
		os.Exit(1)
	}

	// 4. Export to HTML (stdout).
	exp := html.NewExporter()
	exp.Title = r.Info.Name
	if err := exp.Export(e.PreparedPages(), os.Stdout); err != nil {
		fmt.Fprintf(os.Stderr, "export: %v\n", err)
		os.Exit(1)
	}
}
