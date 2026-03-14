# go-fastreport

**go-fastreport** is a pure Go reporting library that ports the core functionality of [FastReport .NET Open Source](https://github.com/FastReports/FastReport) to Go. It provides a band-based report engine, multiple data source adapters, and export to HTML, PDF, and PNG.

---

## Features

- **Band-based layout engine** — ReportTitle, PageHeader, PageFooter, DataBand, GroupHeader, GroupFooter, ChildBand, OverlayBand, and more
- **Data sources** — JSON, XML, CSV, SQL (`database/sql`), and custom in-memory adapters
- **Expression evaluator** — bracket-expression syntax `[DataSource.Field]` with built-in functions
- **FRX serialization** — read/write FastReport XML (`.frx`) report definitions
- **Export targets** — HTML, PDF (structural), PNG image
- **CrossView (pivot table)** — via `crossview` package
- **Barcodes** — QR, Code128, Code39, EAN, DataMatrix, Aztec, PDF417, and more
- **Pure Go** — no CGo dependencies; runs on any platform

---

## Installation

```bash
go get github.com/andrewloable/go-fastreport
```

Requires **Go 1.23+**.

---

## Build & Test

```bash
# Clone the repository
git clone https://github.com/andrewloable/go-fastreport.git
cd go-fastreport

# Build all packages
go build ./...

# Run the full test suite
go test ./...

# Run tests for a specific package
go test ./engine/...
go test ./reportpkg/...

# Run with verbose output
go test -v ./reportpkg/... -run TestFRXSmoke_

# Run benchmarks
go test -bench=. ./engine/...
```

---

## Quick Start

The most common workflow is to design a report in FastReport .NET, save it as an `.frx`
file, then load it at runtime and bind a JSON data source. Here is a complete example:

**`report.frx`** — report template with parameters and data fields:

```xml
<?xml version="1.0" encoding="utf-8"?>
<Report ReportName="EmployeeList">
  <ReportPage Name="Page1" PaperWidth="210" PaperHeight="297"
              LeftMargin="10" TopMargin="10" RightMargin="10" BottomMargin="10">

    <PageHeader Name="PageHeader1" Height="40" Visible="true">
      <!-- [ReportTitle] and [FilterDept] are report variables (parameters) -->
      <TextObject Name="Title" Left="0" Top="2" Width="190" Height="14"
                  Text="[ReportTitle]" Font.Bold="true" HorzAlign="Center"/>
      <TextObject Name="Dept" Left="0" Top="20" Width="190" Height="12"
                  Text="Department: [FilterDept]" HorzAlign="Center"/>
    </PageHeader>

    <!-- DataSource="employees" matches the alias given to jsondata.New("employees") -->
    <Data Name="DataBand1" Height="15" Visible="true" DataSource="employees">
      <!-- [Name], [Department], [Salary] come from the JSON data source -->
      <TextObject Name="NameText"   Left="0"   Top="2" Width="80"  Height="11" Text="[Name]"/>
      <TextObject Name="DeptText"   Left="85"  Top="2" Width="60"  Height="11" Text="[Department]"/>
      <TextObject Name="SalaryText" Left="150" Top="2" Width="40"  Height="11" Text="[Salary]" HorzAlign="Right"/>
    </Data>

    <PageFooter Name="PageFooter1" Height="20" Visible="true">
      <TextObject Name="PageNo" Left="0" Top="5" Width="190" Height="11"
                  Text="Page [PageNumber]" HorzAlign="Right"/>
    </PageFooter>
  </ReportPage>
</Report>
```

**`main.go`**:

```go
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
  {"Name": "Carol",   "Department": "Engineering", "Salary": 105000}
]`

func main() {
    // 1. Load the FRX report definition.
    r := reportpkg.NewReport()
    if err := r.Load("report.frx"); err != nil {
        fmt.Fprintf(os.Stderr, "load FRX: %v\n", err)
        os.Exit(1)
    }

    // 2. Build the Dictionary: parameters (report variables) + data source.
    // The engine resolves [ParamName] bracket expressions from parameters and
    // resolves the DataBand's DataSource attribute by alias from the Dictionary —
    // matching how FastReport .NET wires data at run time automatically.
    ds := jsondata.New("employees") // alias must match DataSource="employees" in FRX
    ds.SetJSON(employeeJSON)
    if err := ds.Init(); err != nil {
        fmt.Fprintf(os.Stderr, "data source init: %v\n", err)
        os.Exit(1)
    }

    dict := r.Dictionary()
    dict.AddParameter(&data.Parameter{Name: "ReportTitle", Value: "Employee Directory"})
    dict.AddParameter(&data.Parameter{Name: "FilterDept", Value: "All Departments"})
    dict.AddDataSource(ds) // engine binds this to DataBand by alias at run time

    // 3. Run the engine.
    e := engine.New(r)
    if err := e.Run(engine.DefaultRunOptions()); err != nil {
        fmt.Fprintf(os.Stderr, "engine run: %v\n", err)
        os.Exit(1)
    }

    // 4. Export to HTML.
    exp := html.NewExporter()
    exp.Title = r.Info.Name
    if err := exp.Export(e.PreparedPages(), os.Stdout); err != nil {
        fmt.Fprintf(os.Stderr, "export: %v\n", err)
        os.Exit(1)
    }
}
```

A runnable version of this example lives in [`examples/frx_json/`](examples/frx_json).
Other working examples:

| Example | Description |
|---------|-------------|
| [`examples/frx_json`](examples/frx_json) | Load FRX + JSON data source + report variables |
| [`examples/simple_html_report`](examples/simple_html_report) | Build a report entirely in code |
| [`examples/json_datasource`](examples/json_datasource) | JSON data source standalone usage |
| [`examples/xml_datasource`](examples/xml_datasource) | XML data source standalone usage |

```bash
go run ./examples/frx_json/
```

---

## Package Overview

| Package | Purpose |
|---------|---------|
| `reportpkg` | `Report`, `ReportPage` — the top-level report definition |
| `band` | All 13 band types: `DataBand`, `GroupHeaderBand`, `PageHeaderBand`, etc. |
| `object` | `TextObject`, `SubreportObject`, `PictureObject`, etc. |
| `engine` | Report execution engine (`ReportEngine.Run`) |
| `expr` | Expression parser and evaluator for `[bracket]` expressions |
| `data` | `DataSource` interface and `BaseDataSource` |
| `data/json` | JSON file/string data source |
| `data/xml` | XML file/string data source |
| `data/csv` | CSV file/reader data source |
| `data/sql` | SQL database data source (`database/sql`) |
| `export/html` | HTML export |
| `export/pdf` | PDF export (structural) |
| `export/image` | PNG image export |
| `crossview` | Pivot table (CrossView) object |
| `barcode` | Barcode rendering (QR, Code128, EAN, etc.) |
| `serial` | FRX XML serialization (`Writer` / `Reader`) |
| `preview` | `PreparedPages` — rendered page output |
| `style` | `Border`, `Fill`, `Style`, font helpers |
| `units` | Unit conversion (mm, cm, inches ↔ pixels) |
| `format` | Number and date formatting |
| `functions` | Built-in expression functions (IIF, Format, etc.) |

---

## Data Sources

All data source types (`JSONDataSource`, `XMLDataSource`, `CSVDataSource`, `SQLDataSource`) embed `data.BaseDataSource` and satisfy `band.DataSource` directly — no adapter wrapper is needed.

### Binding to a report (FRX workflow)

Register the data source in the report Dictionary. The engine matches it to DataBand elements by alias at run time:

```go
ds := jsondata.New("products") // alias matches DataSource="products" in the FRX DataBand
ds.SetJSON(productsJSON)
ds.Init()

r.Dictionary().AddDataSource(ds)
// engine resolves DataBand.DataSource="products" → ds automatically
```

### Binding programmatically (code-built reports)

When building a report in code, assign the data source directly to the DataBand:

```go
db := band.NewDataBand()
db.SetDataSource(ds) // ds is any type satisfying band.DataSource
```

### JSON

```go
import jsondata "github.com/andrewloable/go-fastreport/data/json"

ds := jsondata.New("customers")
ds.SetJSON(`[{"Name":"Alice","Age":30},{"Name":"Bob","Age":25}]`)
if err := ds.Init(); err != nil { ... }
```

### XML

```go
import xmldata "github.com/andrewloable/go-fastreport/data/xml"

ds := xmldata.New("orders")
ds.SetXML(`<Orders><Item Product="Apple" Qty="5"/><Item Product="Banana" Qty="3"/></Orders>`)
if err := ds.Init(); err != nil { ... }
```

### CSV

```go
import csvdata "github.com/andrewloable/go-fastreport/data/csv"

ds := csvdata.New("sales")
ds.SetFilePath("sales.csv")
ds.HasHeader = true
if err := ds.Init(); err != nil { ... }
```

### SQL

```go
import (
    "database/sql"
    _ "github.com/lib/pq"
    sqlds "github.com/andrewloable/go-fastreport/data/sql"
)

db, _ := sql.Open("postgres", "...")
ds := sqlds.New("employees", db, "SELECT id, name FROM employees WHERE active = $1", true)
if err := ds.Init(); err != nil { ... }
```

---

## Export

### HTML

```go
import "github.com/andrewloable/go-fastreport/export/html"

exp := html.NewExporter()
exp.Title = "My Report"
exp.EmbedCSS = true
exp.Scale = 1.0

var buf bytes.Buffer
if err := exp.Export(preparedPages, &buf); err != nil { ... }
```

### PDF

```go
import "github.com/andrewloable/go-fastreport/export/pdf"

exp := pdf.NewExporter()
if err := exp.Export(preparedPages, outputWriter); err != nil { ... }
```

### PNG Image

```go
import "github.com/andrewloable/go-fastreport/export/image"

exp := image.NewExporter()
exp.Scale = 2.0 // 2× for high-DPI
if err := exp.Export(preparedPages, outputWriter); err != nil { ... }
```

### Page range selection

All exporters support `PageRange`:

```go
exp.PageRange = export.PageRangeCustom
exp.PageNumbers = "1,3-5" // export pages 1, 3, 4, 5
```

---

## FRX Serialization

```go
import (
    "github.com/andrewloable/go-fastreport/reportpkg"
    "github.com/andrewloable/go-fastreport/serial"
)

// Write
var buf bytes.Buffer
w := serial.NewWriter(&buf)
w.WriteHeader()
w.WriteObjectNamed("Report", report)
w.Flush()

// Read
r := serial.NewReader(&buf)
typeName, _ := r.ReadObjectHeader()
rep := reportpkg.NewReport()
rep.Deserialize(r)
```

---

## Barcodes

```go
import "github.com/andrewloable/go-fastreport/barcode/qr"

bc := qr.New()
bc.SetData("https://example.com")
img, err := bc.Encode() // returns image.Image
```

---

## CrossView (Pivot Table)

```go
import "github.com/andrewloable/go-fastreport/crossview"

cv := crossview.NewCrossViewObject()
cv.SetSource(myCubeSource) // implements crossview.CubeSourceBase
grid, err := cv.Build()
// grid.Cell(row, col) returns a ResultCell
```

---

## Architecture

```
go-fastreport/
├── reportpkg/       Report, ReportPage
├── band/            13 band types + BandBase
├── object/          TextObject, PictureObject, SubreportObject, ...
├── engine/          ReportEngine (execution pipeline)
├── expr/            Expression parser + evaluator
├── data/            DataSource interface + BaseDataSource
│   ├── json/        JSON data source
│   ├── xml/         XML data source
│   ├── csv/         CSV data source
│   └── sql/         SQL data source
├── export/          ExportBase, page range, utilities
│   ├── html/        HTML exporter
│   ├── pdf/         PDF exporter (+ core PDF objects)
│   └── image/       PNG image exporter
├── crossview/       CrossView pivot table
├── barcode/         Barcode types (QR, Code128, EAN, ...)
├── serial/          FRX XML reader/writer
├── preview/         PreparedPages, PreparedPage, PreparedBand
├── style/           Border, Fill, Font, Style
├── units/           Unit conversions
├── format/          Number/date formatting
├── functions/       Built-in expression functions
├── gauge/           Gauge objects
├── matrix/          Matrix (table-style pivot) object
└── table/           TableObject
```

---

## Band Types

| Band | When Printed |
|------|-------------|
| `ReportTitleBand` | Once at report start |
| `ReportSummaryBand` | Once at report end |
| `PageHeaderBand` | Top of each page |
| `PageFooterBand` | Bottom of each page |
| `ColumnHeaderBand` | Top of each column (multi-column layout) |
| `ColumnFooterBand` | Bottom of each column |
| `DataHeaderBand` | Before first data row |
| `DataBand` | Once per data source row |
| `DataFooterBand` | After last data row |
| `GroupHeaderBand` | At group value change |
| `GroupFooterBand` | At group end |
| `ChildBand` | After its parent band |
| `OverlayBand` | On top of page content |

---

## Status

This is an active port of FastReport .NET Open Source. The following are functional:

- Core engine: data iteration, band rendering, page breaks, multi-column layouts, groups
- Data binding: JSON, XML, CSV, SQL, and custom in-memory data sources
- Dictionary-based DataSource resolution: `DataSource="alias"` in FRX DataBand elements resolved automatically from the report Dictionary at run time — matching the FastReport .NET model
- FRX serialization: read and write `.frx` files including real FastReport sample files
- Export: HTML, PDF (structural), PNG image
- Aggregate totals: Sum, Count, Average, Min, Max with per-group reset
- CanGrow / CanShrink: dynamic band height based on text content
- Expression evaluation: `[DataSource.Field]`, `[Parameter]`, `[SystemVariable]` syntax
- Smoke tested against 50+ real FastReport `.frx` sample files

Features under development:
- Full conditional formatting (HighlightCondition evaluation)
- Master-detail relation traversal at engine runtime
- FRX compression (gzip)
- HTML export of images and vector shapes

See [porting-plan.md](porting-plan.md) for the detailed implementation roadmap.

---

## License

MIT License — see [LICENSE](LICENSE).

---

## Disclaimer

go-fastreport is an independent Go implementation inspired by [FastReport Open Source](https://github.com/FastReports/FastReport). It is not affiliated with or endorsed by Fast Reports Inc.
