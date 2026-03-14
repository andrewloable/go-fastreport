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

## Quick Start

### Simple list report (in-memory data)

```go
package main

import (
    "os"

    "github.com/andrewloable/go-fastreport/band"
    "github.com/andrewloable/go-fastreport/engine"
    "github.com/andrewloable/go-fastreport/export/html"
    "github.com/andrewloable/go-fastreport/reportpkg"
)

// salesDS is a minimal in-memory DataSource.
type salesDS struct {
    rows []map[string]any
    pos  int
}

func (d *salesDS) RowCount() int { return len(d.rows) }
func (d *salesDS) First() error  { d.pos = 0; return nil }
func (d *salesDS) Next() error   { d.pos++; return nil }
func (d *salesDS) EOF() bool     { return d.pos >= len(d.rows) }
func (d *salesDS) GetValue(col string) (any, error) {
    if d.pos < 0 || d.pos >= len(d.rows) {
        return nil, nil
    }
    return d.rows[d.pos][col], nil
}

func main() {
    // 1. Build report definition.
    r := reportpkg.NewReport()
    pg := reportpkg.NewReportPage()
    r.AddPage(pg)

    // Page header.
    hdr := band.NewPageHeaderBand()
    hdr.SetName("PageHeader")
    hdr.SetHeight(40)
    hdr.SetVisible(true)
    pg.SetPageHeader(hdr)

    // Data band.
    db := band.NewDataBand()
    db.SetName("DataBand")
    db.SetHeight(20)
    db.SetVisible(true)
    db.SetDataSource(&salesDS{rows: []map[string]any{
        {"Product": "Apple",  "Qty": 10},
        {"Product": "Banana", "Qty": 5},
        {"Product": "Cherry", "Qty": 20},
    }})
    pg.AddBand(db)

    // Page footer.
    ftr := band.NewPageFooterBand()
    ftr.SetName("PageFooter")
    ftr.SetHeight(30)
    ftr.SetVisible(true)
    pg.SetPageFooter(ftr)

    // 2. Run the engine.
    e := engine.New(r)
    if err := e.Run(engine.DefaultRunOptions()); err != nil {
        panic(err)
    }

    // 3. Export to HTML.
    exp := html.NewExporter()
    exp.Title = "Sales Report"
    if err := exp.Export(e.PreparedPages(), os.Stdout); err != nil {
        panic(err)
    }
}
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

### JSON

```go
import "github.com/andrewloable/go-fastreport/data/json"

ds := json.New("customers")
ds.SetJSON(`[{"Name":"Alice","Age":30},{"Name":"Bob","Age":25}]`)
if err := ds.Init(); err != nil { ... }

ds.First()
for !ds.EOF() {
    name, _ := ds.GetValue("Name")
    fmt.Println(name)
    ds.Next()
}
```

### XML

```go
import "github.com/andrewloable/go-fastreport/data/xml"

ds := xml.New("orders")
ds.SetXML(`<Orders><Item Product="Apple" Qty="5"/><Item Product="Banana" Qty="3"/></Orders>`)
ds.SetRootPath("") // root element is the container
if err := ds.Init(); err != nil { ... }
```

### CSV

```go
import "github.com/andrewloable/go-fastreport/data/csv"

ds := csv.New("sales")
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

This is an active port of FastReport .NET Open Source. The core engine, data binding, serialization, and export layers are functional. Advanced features (full expression evaluation with data context, conditional formatting, auto-grow text) are under development.

See [porting-plan.md](porting-plan.md) for the detailed implementation roadmap.

---

## License

MIT License — see [LICENSE](LICENSE).

---

## Disclaimer

go-fastreport is an independent Go implementation inspired by [FastReport Open Source](https://github.com/FastReports/FastReport). It is not affiliated with or endorsed by Fast Reports Inc.
