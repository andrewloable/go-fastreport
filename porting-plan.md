# FastReport .NET to Go Porting Plan

## Table of Contents

1. [Codebase Overview](#1-codebase-overview)
2. [Architecture Analysis](#2-architecture-analysis)
3. [Go Package Structure](#3-go-package-structure)
4. [C# to Go Translation Patterns](#4-c-to-go-translation-patterns)
5. [Phased Porting Plan](#5-phased-porting-plan)
6. [Detailed Module Breakdown](#6-detailed-module-breakdown)
7. [Risk Assessment & Gaps](#7-risk-assessment--gaps)
8. [Testing Strategy](#8-testing-strategy)
9. [Dependencies & Third-Party Libraries](#9-dependencies--third-party-libraries)
10. [Out of Scope](#10-out-of-scope)

---

## 1. Codebase Overview

### Source Statistics

| Project | .cs Files | Purpose |
|---------|-----------|---------|
| **FastReport.Base** | 397 | Core reporting engine, objects, data, exports |
| **FastReport.OpenSource** | 39 | Open-source distribution layer |
| **FastReport.Compat** | 27 | .NET Framework backward compatibility |
| **FastReport.Core.Web** | 108 | ASP.NET Core web components |
| **FastReport.Web.Base** | 3 | Web base classes |
| **Extras/** | 156 | Database connectors, PDF export, plugins |
| **Demos/** | 42 | Example applications |
| **Total** | **786** | |

### FastReport.Base Breakdown (primary porting target)

| Directory | Files | Description |
|-----------|-------|-------------|
| Root (objects) | 80 | Bands, text, shapes, images, containers |
| Utils/ | 57 | Serialization, JSON, fonts, compression, helpers |
| Data/ | 50 | Dictionary, data sources, connections, parameters |
| Barcode/ | 57 | QR, Aztec, Code128, EAN, DataMatrix, PDF417, etc. |
| Engine/ | 18 | Report rendering engine (partial classes) |
| Import/ | 15 | RDL, DevExpress, JasperReports importers |
| Table/ | 13 | Table object with cells, spans, dynamic rows |
| Code/ | 12 | Expression parsing, script compilation |
| Matrix/ | 11 | Pivot table / cross-tab component |
| Format/ | 10 | Number, date, currency, percent formatting |
| Export/ | 9 | HTML export, ~~image export~~, base classes (porting HTML + PDF only) |
| Gauge/ | 19 | Linear, radial, simple gauge visualizations |
| CrossView/ | 8 | Cross-tabulation view objects |
| TypeConverters/ | 12 | Property type converters (designer support) |
| Preview/ | 8 | Prepared pages, page cache, bookmarks, outline |
| Functions/ | 18 | Built-in report functions (NumToWords, Roman, StdFunctions) |

---

## 2. Architecture Analysis

### 2.1 Object Model Hierarchy

```
Base (abstract, IFRSerializable)
├── ComponentBase (abstract) — positioning: Top, Left, Width, Height
│   ├── PageBase (abstract)
│   │   └── ReportPage : IParent — page settings, bands container
│   └── ReportComponentBase (abstract) — styling, borders, fills, visibility
│       ├── BreakableComponent — CanGrow, CanShrink, CanBreak
│       │   ├── BandBase (abstract, IParent) — all band types
│       │   │   ├── DataBand : IContainDataSource
│       │   │   ├── ChildBand
│       │   │   ├── HeaderFooterBandBase (abstract)
│       │   │   │   ├── DataHeaderBand, DataFooterBand
│       │   │   │   ├── PageHeaderBand, PageFooterBand
│       │   │   │   ├── ColumnHeaderBand, ColumnFooterBand
│       │   │   │   └── GroupHeaderBand
│       │   │   ├── GroupFooterBand
│       │   │   ├── ReportTitleBand, ReportSummaryBand
│       │   │   └── OverlayBand
│       │   └── TextObjectBase (abstract)
│       │       ├── TextObject — rich text with expressions
│       │       ├── HtmlObject — HTML content
│       │       └── CellularTextObject — grid text
│       ├── LineObject, ShapeObject
│       ├── PolyLineObject, PolygonObject
│       ├── CheckBoxObject
│       ├── SubreportObject
│       ├── ContainerObject : IParent
│       ├── PictureObjectBase → PictureObject
│       ├── RFIDLabel
│       └── ZipCodeObject
│
└── Report : IParent — top-level container
    ├── Pages (PageCollection of ReportPage)
    ├── Dictionary — data sources, connections, parameters, totals
    ├── Styles (StyleCollection)
    └── ReportInfo — metadata
```

### 2.2 Core Interfaces

| Interface | Methods | Go Equivalent |
|-----------|---------|---------------|
| `IFRSerializable` | `Serialize(FRWriter)`, `Deserialize(FRReader)` | `Serializable` interface |
| `IParent` | `CanContain`, `AddChild`, `RemoveChild`, `GetChildObjects`, `GetChildOrder`, `SetChildOrder`, `UpdateLayout` | `Parent` interface |
| `IContainDataSource` | Data source binding | `DataSourceBinder` interface |
| `ITranslatable` | Localization support | `Translatable` interface |

### 2.3 Report Engine Flow

```
Report.Prepare()
  ├── Phase 1: RunPhase1(resetDataState)
  │   ├── Initialize data sources
  │   └── Fire OnStartReport
  ├── Phase 2: RunPhase2(append)
  │   ├── First Pass:
  │   │   ├── PrepareToFirstPass()
  │   │   └── RunReportPages() → for each ReportPage:
  │   │       ├── Show ReportTitle
  │   │       ├── Show PageHeader
  │   │       ├── Show ColumnHeader
  │   │       ├── RunBands(page.Bands) → for each band:
  │   │       │   ├── DataBand → RunDataBand():
  │   │       │   │   ├── Init data source, move to first row
  │   │       │   │   ├── For each row:
  │   │       │   │   │   ├── Show DataHeader (first row)
  │   │       │   │   │   ├── PrepareBand → GetData, CalcHeight
  │   │       │   │   │   ├── ShowBand → AddToPreparedPages
  │   │       │   │   │   ├── Handle keep-together
  │   │       │   │   │   └── Run nested bands
  │   │       │   │   └── Show DataFooter
  │   │       │   └── GroupHeaderBand → RunGroup():
  │   │       │       ├── Build group tree
  │   │       │       └── Show group tree recursively
  │   │       ├── Show ReportSummary
  │   │       ├── Show ColumnFooter
  │   │       └── Show PageFooter
  │   └── Second Pass (if DoublePass):
  │       ├── Re-init data, set TotalPages
  │       └── RunReportPages() again
  └── Phase 3: RunFinished()
      └── Fire OnFinishReport
```

### 2.4 Data System Architecture

```
Dictionary (Central Registry)
├── Connections (ConnectionCollection)
│   └── DataConnectionBase → database-specific implementations
│       └── TableDataSource (per table/view)
│           ├── Columns (ColumnCollection)
│           └── CommandParameters
├── DataSources (DataSourceCollection)
│   ├── TableDataSource — SQL table/view
│   ├── BusinessObjectDataSource — IEnumerable/Go slices
│   ├── ViewDataSource — filtered views
│   ├── VirtualDataSource — custom data
│   └── ProcedureDataSource — stored procedures
├── Relations (RelationCollection)
│   └── Relation — master-detail links (ParentColumns ↔ ChildColumns)
├── Parameters (ParameterCollection)
│   └── Parameter — name, type, value/expression, nested params
├── Totals (TotalCollection)
│   └── Total — Sum, Avg, Min, Max, Count, CountDistinct
└── SystemVariables
    └── PageNumber, TotalPages, Date, Time, Row#, etc.
```

### 2.5 Serialization (FRX Format)

- XML-based, serialized via `FRWriter` / deserialized via `FRReader`
- **Delta serialization**: only properties differing from defaults are written
- Supports nested objects, collections, references
- Compression and password protection support
- All objects implement `IFRSerializable`

### 2.6 Expression System

- Bracket syntax: `[DataSource.FieldName]`, `[Parameter]`, `[PageNumber]`
- Evaluated via `Report.Calc(expression)` in current data context
- Used in: text content, filters, sorts, visibility, conditional formatting
- C# expressions compiled to .NET assemblies
- **Go replacement**: Expression parser/evaluator using `expr` or custom parser

### 2.7 Export System (HTML + PDF only)

| Export | Files | Output | Status |
|--------|-------|--------|--------|
| ExportBase | 2 | Abstract base with page range, stream output | Port |
| HTMLExport | 7 | Full HTML with CSS, layers, WYSIWYG | Port |
| PdfSimpleExport | 22 (Extras) | PDF generation with text, images, lines | Port |
| ~~ImageExport~~ | ~~1~~ | ~~PNG, JPEG, BMP, GIF, TIFF~~ | **Excluded** |

### 2.8 Band Types (13)

| Band | Trigger | Key Properties |
|------|---------|----------------|
| ReportTitleBand | Once at report start | — |
| ReportSummaryBand | Once at report end | — |
| PageHeaderBand | Top of each page | PrintOn (first/last/odd/even) |
| PageFooterBand | Bottom of each page | PrintOn |
| ColumnHeaderBand | Top of each column | — |
| ColumnFooterBand | Bottom of each column | — |
| DataHeaderBand | Before data rows | RepeatOnEveryPage, KeepWithData |
| DataBand | Per data source row | DataSource, Filter, Sort, Columns, Relation |
| DataFooterBand | After data rows | KeepWithData |
| GroupHeaderBand | Group value change | Condition, SortOrder, NestedGroup |
| GroupFooterBand | Group end | — |
| ChildBand | After parent band | FillUnusedSpace, CompleteToNRows |
| OverlayBand | On top of content | — |

---

## 3. Go Package Structure

```
go-fastreport/
├── go.mod
├── report/                     # Core report types
│   ├── report.go              # Report struct (top-level container)
│   ├── page.go                # ReportPage
│   ├── base.go                # Base struct (common properties)
│   ├── component.go           # ComponentBase (positioning)
│   ├── reportcomponent.go     # ReportComponentBase (styling)
│   ├── breakable.go           # BreakableComponent
│   ├── interfaces.go          # Serializable, Parent, DataSourceBinder
│   ├── settings.go            # ReportSettings, ReportInfo
│   └── collections.go         # ObjectCollection, PageCollection
│
├── band/                       # Band types
│   ├── band.go                # BandBase
│   ├── data.go                # DataBand
│   ├── group.go               # GroupHeaderBand, GroupFooterBand
│   ├── header_footer.go       # HeaderFooterBandBase
│   ├── page_bands.go          # PageHeaderBand, PageFooterBand
│   ├── column_bands.go        # ColumnHeaderBand, ColumnFooterBand
│   ├── report_bands.go        # ReportTitleBand, ReportSummaryBand
│   ├── child.go               # ChildBand
│   ├── overlay.go             # OverlayBand
│   └── columns.go             # BandColumns (multi-column layout)
│
├── object/                     # Report objects
│   ├── text.go                # TextObjectBase, TextObject
│   ├── html.go                # HtmlObject
│   ├── picture.go             # PictureObjectBase, PictureObject
│   ├── line.go                # LineObject
│   ├── shape.go               # ShapeObject
│   ├── polyline.go            # PolyLineObject, PolygonObject
│   ├── checkbox.go            # CheckBoxObject
│   ├── subreport.go           # SubreportObject
│   ├── container.go           # ContainerObject
│   ├── cellular_text.go       # CellularTextObject
│   └── zipcode.go             # ZipCodeObject
│
├── style/                      # Styling system
│   ├── border.go              # Border, BorderLine
│   ├── fill.go                # Fills (Solid, Linear, Glass, Hatch, Path)
│   ├── style.go               # Style, StyleCollection
│   ├── highlight.go           # HighlightCondition
│   ├── watermark.go           # Watermark
│   └── textoutline.go         # TextOutline
│
├── data/                       # Data system
│   ├── dictionary.go          # Dictionary (central registry)
│   ├── datasource.go          # DataSourceBase interface and base
│   ├── table_source.go        # TableDataSource
│   ├── business_source.go     # BusinessObjectDataSource (Go slices/maps)
│   ├── connection.go          # DataConnectionBase
│   ├── column.go              # Column, ColumnCollection
│   ├── relation.go            # Relation, RelationCollection
│   ├── parameter.go           # Parameter, ParameterCollection
│   ├── total.go               # Total, TotalCollection
│   ├── systemvars.go          # SystemVariables
│   ├── filter.go              # DataSourceFilter
│   └── helper.go              # DataHelper
│
├── data/csv/                   # CSV data connection
│   └── csv.go
│
├── data/json/                  # JSON data connection
│   └── json.go
│
├── data/xml/                   # XML data connection
│   └── xml.go
│
├── data/sql/                   # SQL database connections
│   ├── connection.go          # Generic SQL connection
│   ├── postgres.go            # PostgreSQL
│   ├── mysql.go               # MySQL
│   ├── sqlite.go              # SQLite
│   └── mssql.go               # MS SQL Server
│
├── engine/                     # Report rendering engine
│   ├── engine.go              # ReportEngine main struct and properties
│   ├── run.go                 # Run, RunPhase1/2/3
│   ├── pages.go               # Page management
│   ├── bands.go               # Band processing (PrepareBand, ShowBand)
│   ├── databands.go           # Data band iteration
│   ├── groups.go              # Group processing
│   ├── breaks.go              # Page break logic
│   ├── subreports.go          # Subreport rendering
│   ├── keep.go                # Keep-together mechanism
│   ├── keepwithdata.go        # Footer keep-with-data
│   ├── processat.go           # Deferred text processing
│   ├── outline.go             # PDF outline/bookmarks
│   ├── pagenumbers.go         # Logical page numbering
│   └── reprint.go             # Repeating headers on new pages
│
├── expr/                       # Expression system
│   ├── parser.go              # Expression parser ([DataSource.Field])
│   ├── evaluator.go           # Expression evaluator
│   ├── functions.go           # Built-in functions
│   └── context.go             # Evaluation context
│
├── format/                     # Value formatting
│   ├── format.go              # FormatBase interface
│   ├── number.go              # NumberFormat
│   ├── currency.go            # CurrencyFormat
│   ├── date.go                # DateFormat
│   ├── time.go                # TimeFormat
│   ├── percent.go             # PercentFormat
│   ├── boolean.go             # BooleanFormat
│   ├── custom.go              # CustomFormat
│   └── general.go             # GeneralFormat
│
├── serial/                     # FRX serialization
│   ├── reader.go              # FRReader (XML deserialization)
│   ├── writer.go              # FRWriter (XML serialization)
│   └── registry.go            # Object type registry for deserialization
│
├── preview/                    # Prepared/rendered pages
│   ├── prepared_pages.go      # PreparedPages collection
│   ├── prepared_page.go       # Single prepared page
│   ├── page_cache.go          # Page caching
│   ├── bookmarks.go           # Bookmarks
│   ├── outline.go             # Document outline
│   └── source_pages.go        # Source page tracking
│
├── export/                     # Export system
│   ├── base.go                # ExportBase
│   ├── utils.go               # ExportUtils
│   ├── html/                  # HTML export
│   │   ├── html.go            # Main HTML exporter
│   │   ├── draw.go            # Object drawing
│   │   ├── layers.go          # Layer handling
│   │   ├── styles.go          # CSS generation
│   │   └── templates.go       # HTML templates
│   └── pdf/                   # PDF export (ported from PdfSimple)
│       ├── export.go          # PDFSimpleExport main
│       ├── config.go          # PDF configuration
│       ├── images.go          # Image embedding in PDF
│       ├── writer.go          # PdfWriter
│       ├── page.go            # PdfPage, PdfPages
│       ├── contents.go        # PdfContents
│       ├── catalog.go         # PdfCatalog, PdfInfo
│       ├── image.go           # PdfImage, PdfMask
│       └── core/              # PDF primitives
│           ├── object.go      # PdfObjectBase, PdfIndirectObject, PdfDirectObject
│           ├── dict.go        # PdfDictionary
│           ├── array.go       # PdfArray
│           ├── stream.go      # PdfStream
│           ├── string.go      # PdfString, PdfName
│           ├── numeric.go     # PdfNumeric, PdfBoolean
│           └── trailid.go     # PdfTrailerId
│
├── barcode/                    # Barcode generation
│   ├── barcode.go             # BarcodeObject report component
│   ├── base.go                # BarcodeBase
│   ├── linear.go              # LinearBarcodeBase
│   ├── code128.go             # Code128
│   ├── code39.go              # Code39
│   ├── ean.go                 # EAN-8, EAN-13
│   ├── upc.go                 # UPC-A, UPC-E
│   ├── qr/                    # QR Code
│   │   ├── encoder.go
│   │   ├── version.go
│   │   └── reed_solomon.go
│   ├── aztec/                 # Aztec Code
│   │   ├── encoder.go
│   │   └── reed_solomon.go
│   ├── datamatrix.go          # DataMatrix
│   └── pdf417.go              # PDF417
│
├── table/                      # Table component
│   ├── table.go               # TableBase, TableObject
│   ├── cell.go                # TableCell
│   ├── row.go                 # TableRow
│   ├── column.go              # TableColumn
│   └── result.go              # TableResult
│
├── matrix/                     # Matrix/Pivot table
│   ├── matrix.go              # MatrixObject
│   ├── descriptor.go          # MatrixDescriptor, HeaderDescriptor, CellDescriptor
│   ├── data.go                # MatrixData
│   ├── header.go              # MatrixHeader, MatrixHeaderItem
│   ├── cells.go               # MatrixCells
│   └── helper.go              # MatrixHelper
│
├── gauge/                      # Gauge objects (lower priority)
│   ├── gauge.go               # GaugeObject
│   ├── linear.go              # Linear gauge
│   ├── radial.go              # Radial gauge
│   └── simple.go              # Simple gauge
│
├── crossview/                  # CrossView objects (lower priority)
│   └── crossview.go
│
├── units/                      # Units and measurements
│   └── units.go               # Pixels, mm, inches, cm conversion
│
├── utils/                      # Shared utilities
│   ├── faststring.go          # Optimized string builder
│   ├── converter.go           # Type conversion utilities
│   ├── variant.go             # Flexible value container
│   ├── color.go               # Color helpers
│   ├── font.go                # Font management (FontManager)
│   ├── text.go                # Text measurement (TextRenderer)
│   ├── htmltext.go            # HTML text rendering
│   ├── image.go               # Image loading helpers
│   ├── graphics.go            # Graphics cache and drawing utils
│   ├── zip.go                 # ZIP compression
│   ├── compressor.go          # Data compression
│   ├── crc32.go               # CRC32
│   ├── crypto.go              # String encryption
│   ├── blobstore.go           # Binary data storage
│   ├── xml.go                 # XML utilities
│   ├── namecreator.go         # Unique name generation
│   ├── collection.go          # FRCollectionBase
│   ├── config.go              # Global configuration
│   └── errors.go              # Custom error types
│
└── functions/                  # Built-in report functions
    ├── standard.go            # Standard functions (StdFunctions)
    ├── numtowords.go          # Number to words (English)
    ├── numtowords_base.go     # Base class for localized number-to-words
    └── roman.go               # Roman numeral conversion
```

---

## 4. C# to Go Translation Patterns

### 4.1 Class Hierarchy → Struct Embedding + Interfaces

**C# (inheritance):**
```csharp
public abstract class ReportComponentBase : ComponentBase { }
public class TextObject : TextObjectBase { }
```

**Go (composition + interfaces):**
```go
type ReportComponentBase struct {
    ComponentBase
    Border    Border
    Fill      Fill
    // ...
}

type TextObject struct {
    TextObjectBase
    // ...
}
```

### 4.2 C# Properties → Go Exported Fields or Methods

Simple properties become exported fields. Computed properties become methods:

```go
// Simple property
type ComponentBase struct {
    Top    float32
    Left   float32
    Width  float32
    Height float32
}

// Computed property
func (e *ReportEngine) FreeSpace() float32 {
    return e.pageHeight - e.curY - e.pageFooterHeight
}
```

### 4.3 C# Events → Go Callbacks

```go
type BandBase struct {
    BreakableComponent
    OnBeforePrint  func(sender interface{}, args EventArgs)
    OnAfterPrint   func(sender interface{}, args EventArgs)
    OnBeforeLayout func(sender interface{}, args EventArgs)
    OnAfterLayout  func(sender interface{}, args EventArgs)
}
```

### 4.4 C# Partial Classes → Go Files in Same Package

The `ReportEngine` split across 19 partial .cs files maps naturally to multiple `.go` files in the `engine/` package, all operating on the same `ReportEngine` struct.

### 4.5 C# Enums → Go Constants

```go
type HorzAlign int

const (
    HorzAlignLeft   HorzAlign = iota
    HorzAlignCenter
    HorzAlignRight
    HorzAlignJustify
)
```

### 4.6 C# Collections → Go Slices with Helper Methods

```go
type ObjectCollection struct {
    items []Base
}

func (c *ObjectCollection) Add(item Base)    { c.items = append(c.items, item) }
func (c *ObjectCollection) Count() int       { return len(c.items) }
func (c *ObjectCollection) Get(i int) Base   { return c.items[i] }
```

### 4.7 C# Generics → Go Generics (Go 1.18+)

```go
type Collection[T any] struct {
    items []T
}
```

### 4.8 C# using/IDisposable → Go defer

```go
func (r *Report) Prepare() error {
    defer r.cleanup()
    // ...
}
```

### 4.9 C# async/await → Go goroutines (where needed)

Most report processing is sequential. Async patterns in the .NET code are mainly for UI responsiveness and can be omitted in Go. Use goroutines only where true concurrency is beneficial (e.g., parallel exports).

### 4.10 Expression Engine Replacement

C# compiles expressions to .NET assemblies. Go needs a different approach:

**Option A: Use `github.com/expr-lang/expr`** (recommended)
- Compile-once, evaluate-many
- Type-safe, sandboxed
- Supports custom functions

**Option B: Use `text/template`**
- Go-native but limited expression support

**Option C: Custom parser**
- Full control, more work
- Parse `[DataSource.Field]` bracket syntax directly

### 4.11 Graphics/Drawing

C# uses `System.Drawing` / `SkiaSharp`. Go approach:
- **No rasterization needed** — HTML export generates CSS/HTML directly; PDF export generates PDF primitives directly
- Font metrics for text measurement via `golang.org/x/image/font` (pure Go)
- Image handling (embedded pictures) via Go standard `image` package
- No CGo, no System.Drawing equivalent needed

---

## 5. Phased Porting Plan

### Phase 1: Foundation (Weeks 1-4)

**Goal**: Core types, serialization, basic report loading

| Task | Source Files | Go Package | Priority |
|------|-------------|------------|----------|
| Base types & interfaces | Base.cs, IFRSerializable.cs, IParent.cs, IContainDataSource.cs, ITranslatable.cs | `report/` | P0 |
| Component positioning | ComponentBase.cs | `report/` | P0 |
| Report component styling | ReportComponentBase.cs, Border.cs, Fills.cs | `report/`, `style/` | P0 |
| Units system | Utils/Units.cs | `units/` | P0 |
| FRX Reader (deserialize) | Utils/FRReader.cs | `serial/` | P0 |
| FRX Writer (serialize) | Utils/FRWriter.cs | `serial/` | P0 |
| **Object type registry** | **Utils/RegisteredObjects.cs** (1251 lines) | `serial/` | **P0** |
| Report container | Report.cs, ReportPage.cs | `report/` | P0 |
| Page settings | PageBase.cs, ReportPage.cs | `report/` | P0 |
| Collections | ObjectCollection.cs, PageCollection.cs, BandCollection.cs, ReportComponentCollection.cs | `report/` | P0 |
| Base collection class | Utils/FRCollectionBase.cs | `utils/` | P0 |
| Style system | Style.cs, StyleBase.cs, StyleCollection.cs, StyleSheet.cs | `style/` | P1 |
| Utility classes | FastString.cs, Converter.cs, Variant.cs, Config.cs, Xml.cs | `utils/` | P1 |
| Error types | Exceptions.cs | `utils/` | P1 |
| Compression | Utils/Compressor.cs, Utils/Zip.cs, Utils/Crc32.cs | `utils/` | P1 |
| Unique naming | Utils/FastNameCreator.cs | `utils/` | P1 |
| Encryption | Utils/Crypter.cs | `utils/` | P2 |
| Color helpers | Utils/ColorHelper.cs | `utils/` | P1 |

**Milestone**: Can load and save `.frx` report files with structural fidelity.

---

### Phase 2: Data System (Weeks 5-8)

**Goal**: Dictionary, data sources, parameters, totals

| Task | Source Files | Go Package | Priority |
|------|-------------|------------|----------|
| Dictionary | Data/Dictionary.cs | `data/` | P0 |
| DataComponentBase | Data/DataComponentBase.cs | `data/` | P0 |
| DataSourceBase | Data/DataSourceBase.cs | `data/` | P0 |
| Column & ColumnCollection | Data/Column.cs, ColumnCollection.cs | `data/` | P0 |
| BusinessObjectDataSource | Data/BusinessObjectDataSource.cs | `data/` | P0 |
| Parameter & ParameterCollection | Data/Parameter.cs, ParameterCollection.cs | `data/` | P0 |
| Total & TotalCollection | Data/Total.cs, TotalCollection.cs | `data/` | P0 |
| SystemVariables | Data/SystemVariables.cs | `data/` | P0 |
| Relation & RelationCollection | Data/Relation.cs, RelationCollection.cs | `data/` | P0 |
| DataSourceFilter | Data/DataSourceFilter.cs | `data/` | P1 |
| DataConnectionBase | Data/DataConnectionBase.cs | `data/` | P1 |
| ConnectionCollection | Data/ConnectionCollection.cs | `data/` | P1 |
| DataSourceCollection | Data/DataSourceCollection.cs | `data/` | P1 |
| TableDataSource | Data/TableDataSource.cs | `data/` | P1 |
| TableCollection | Data/TableCollection.cs | `data/` | P1 |
| CommandParameter | Data/CommandParameter.cs, CommandParameterCollection.cs | `data/` | P1 |
| CSV connection | Data/CsvDataConnection.cs, CsvConnectionStringBuilder.cs, CsvUtils.cs | `data/csv/` | P1 |
| JSON connection | Data/JsonConnection/*.cs (5 files) | `data/json/` | P1 |
| XML connection | Data/XmlDataConnection.cs, XmlConnectionStringBuilder.cs | `data/xml/` | P2 |
| DataHelper & DictionaryHelper | Data/DataHelper.cs, DictionaryHelper.cs | `data/` | P1 |
| BusinessObjectConverter | Data/BusinessObjectConverter.cs | `data/` | P1 |
| SQL connections | Extras (Postgres, MySQL, SQLite, MSSQL) | `data/sql/` | P2 |
| ProcedureDataSource | Data/ProcedureDataSource.cs, ProcedureParameter.cs | `data/` | P2 |
| ViewDataSource | Data/ViewDataSource.cs | `data/` | P3 |
| VirtualDataSource | Data/VirtualDataSource.cs | `data/` | P3 |

**Milestone**: Can bind Go data (slices, maps, structs) to report data sources.

---

### Phase 3: Report Objects (Weeks 9-12)

**Goal**: All visual report objects

| Task | Source Files | Go Package | Priority |
|------|-------------|------------|----------|
| All band types | BandBase.cs + 13 band type files, HeaderFooterBandBase.cs | `band/` | P0 |
| BreakableComponent | BreakableComponent.cs | `report/` | P0 |
| TextObjectBase + TextObject | TextObjectBase.cs, TextObject.cs | `object/` | P0 |
| **Font management** | **Utils/FontManager.cs, FontManager.Internals.cs, FontManager.Gdi.cs** | **`utils/`** | **P0** |
| **Text measurement** | **Utils/TextRenderer.cs** | **`utils/`** | **P0** |
| **HTML text rendering** | **Utils/HtmlTextRenderer.cs** | **`utils/`** | **P1** |
| **Image helpers** | **Utils/ImageHelper.cs, GraphicCache.cs, DrawUtils.cs** | **`utils/`** | **P1** |
| HtmlObject | HtmlObject.cs | `object/` | P1 |
| PictureObject | PictureObjectBase.cs, PictureObject.cs | `object/` | P0 |
| LineObject | LineObject.cs | `object/` | P0 |
| ShapeObject | ShapeObject.cs | `object/` | P0 |
| PolyLineObject, PolygonObject | PolyLineObject.cs, PolygonObject.cs | `object/` | P1 |
| CheckBoxObject | CheckBoxObject.cs | `object/` | P1 |
| SubreportObject | SubreportObject.cs | `object/` | P1 |
| ContainerObject | ContainerObject.cs | `object/` | P1 |
| ZipCodeObject | ZipCodeObject.cs | `object/` | P2 |
| CellularTextObject | CellularTextObject.cs | `object/` | P2 |
| RFIDLabel | RFIDLabel.cs | `object/` | P3 |
| Sort & SortCollection | Sort.cs, SortCollection.cs | `report/` | P0 |
| HighlightCondition & ConditionCollection | HighlightCondition.cs, ConditionCollection.cs | `style/` | P1 |
| Hyperlink | Hyperlink.cs | `report/` | P2 |
| Watermark | Watermark.cs | `style/` | P2 |
| CapSettings | CapSettings.cs | `style/` | P2 |
| TextOutline | TextOutline.cs | `style/` | P2 |
| Format system | Format/*.cs (10 files) | `format/` | P0 |
| Private font collection | Utils/FRPrivateFontCollection.cs | `utils/` | P2 |

**Milestone**: All report objects can be loaded, configured, and serialized.

---

### Phase 4: Expression Engine (Weeks 13-15)

**Goal**: Parse and evaluate `[expressions]` in report context

| Task | Source Files | Go Package | Priority |
|------|-------------|------------|----------|
| Expression parser | Code/CodeUtils.cs | `expr/` | P0 |
| Expression evaluator | Code/ExpressionDescriptor.cs | `expr/` | P0 |
| Built-in functions | Functions/StdFunctions.cs | `expr/` | P0 |
| Evaluation context | Code/CodeProvider.cs | `expr/` | P0 |
| Bracket syntax `[Field]` | (custom for Go) | `expr/` | P0 |
| NumToWords (English) | Functions/NumToWordsEn.cs | `expr/` | P2 |
| Roman numerals | Functions/Roman.cs | `expr/` | P2 |

**Note**: The C# codebase compiles expressions as C#/VB.NET code via Roslyn. The Go port must use a different approach — either `expr-lang/expr` library or a custom expression evaluator. The bracket syntax `[DataSource.Field]` must be parsed and resolved against the current data context.

**Milestone**: Expressions in TextObject, filters, and visibility conditions evaluate correctly.

---

### Phase 5: Report Engine (Weeks 16-22)

**Goal**: Full report rendering pipeline

| Task | Source Files | Go Package | Priority |
|------|-------------|------------|----------|
| ReportEngine core | Engine/ReportEngine.cs | `engine/` | P0 |
| Page management | Engine/ReportEngine.Pages.cs | `engine/` | P0 |
| Band processing | Engine/ReportEngine.Bands.cs | `engine/` | P0 |
| Data band iteration | Engine/ReportEngine.DataBands.cs | `engine/` | P0 |
| Group processing | Engine/ReportEngine.Groups.cs | `engine/` | P0 |
| Page breaks | Engine/ReportEngine.Break.cs | `engine/` | P0 |
| Subreport rendering | Engine/ReportEngine.Subreports.cs | `engine/` | P1 |
| Keep-together | Engine/ReportEngine.Keep.cs | `engine/` | P1 |
| Keep-with-data | Engine/ReportEngine.KeepWithData.cs | `engine/` | P1 |
| Deferred processing | Engine/ReportEngine.ProcessAt.cs | `engine/` | P1 |
| Page numbering | Engine/ReportEngine.PageNumbers.cs | `engine/` | P1 |
| Outline/bookmarks | Engine/ReportEngine.Outline.cs | `engine/` | P2 |
| Reprint headers | Engine/ReportEngine.Reprint.cs | `engine/` | P2 |
| PreparedPages | Preview/PreparedPages.cs, PreparedPage.cs | `preview/` | P0 |
| Preview Dictionary | Preview/Dictionary.cs | `preview/` | P0 |
| Source pages | Preview/SourcePages.cs | `preview/` | P1 |
| Page postprocessor | Preview/PreparedPagePostprocessor.cs | `preview/` | P1 |
| Page cache | Preview/PageCache.cs | `preview/` | P1 |
| BlobStore | Utils/BlobStore.cs | `preview/` | P1 |
| ShortProperties | Utils/ShortProperties.cs | `preview/` | P2 |
| Bookmarks | Preview/Bookmarks.cs | `preview/` | P2 |
| Outline | Preview/Outline.cs | `preview/` | P2 |
| Two-pass rendering | (part of engine flow) | `engine/` | P1 |
| Multi-column layout | BandColumns.cs, engine multi-column | `engine/` | P2 |
| Hierarchical data | IdColumn/ParentIdColumn in DataBand | `engine/` | P2 |

**Milestone**: Can prepare reports with data, producing PreparedPages output.

---

### Phase 6: Export System (Weeks 23-28)

**Goal**: HTML and PDF export only (no image export)

| Task | Source Files | Go Package | Priority |
|------|-------------|------------|----------|
| ExportBase | Export/ExportBase.cs | `export/` | P0 |
| ExportUtils | Export/ExportUtils.cs | `export/` | P0 |
| ExportsOptions | Utils/ExportsOptions.cs | `export/` | P1 |
| HTML Export | Export/Html/*.cs (7 files: HTMLExport, Draw, Layers, Styles, Templates, Utils) | `export/html/` | P0 |
| PDF Export - core | Extras/PdfSimple/PdfCore/*.cs (10 files: PdfArray, PdfDictionary, PdfStream, etc.) | `export/pdf/` | P0 |
| PDF Export - objects | Extras/PdfSimple/PdfObjects/*.cs (9 files: PdfPage, PdfContents, PdfImage, etc.) | `export/pdf/` | P0 |
| PDF Export - main | Extras/PdfSimple/PDFSimpleExport.cs, .Config.cs, .Images.cs | `export/pdf/` | P0 |

**Milestone**: Reports can be exported to HTML and PDF.

**Note**: Image export (PNG/JPEG/BMP/TIFF) is excluded to avoid heavy graphics dependencies. HTML and PDF cover the primary use cases. PDF is ported from PdfSimple which is self-contained pure code with no external dependencies.

---

### Phase 7: Advanced Components (Weeks 29-36)

**Goal**: Tables, matrices, barcodes

| Task | Source Files | Go Package | Priority |
|------|-------------|------------|----------|
| TableBase | Table/TableBase.cs | `table/` | P1 |
| TableObject | Table/TableObject.cs | `table/` | P1 |
| TableCell | Table/TableCell.cs, TableCellData.cs | `table/` | P1 |
| TableRow & TableColumn | Table/TableRow.cs, TableColumn.cs | `table/` | P1 |
| Table collections | Table/TableRowCollection.cs, TableColumnCollection.cs | `table/` | P1 |
| Table helpers | Table/TableHelper.cs, TableResult.cs, TableStyleCollection.cs | `table/` | P1 |
| MatrixObject | Matrix/MatrixObject.cs | `matrix/` | P2 |
| Matrix descriptors | Matrix/Matrix*.cs | `matrix/` | P2 |
| BarcodeObject | Barcode/BarcodeObject.cs | `barcode/` | P1 |
| BarcodeBase, LinearBarcodeBase | Barcode/BarcodeBase.cs, LinearBarcodeBase.cs | `barcode/` | P1 |
| Barcode2DBase | Barcode/Barcode2DBase.cs | `barcode/` | P1 |
| GS1Helper | Barcode/GS1Helper.cs | `barcode/` | P1 |
| Code128, Code39 | Barcode/Barcode128.cs, Barcode39.cs | `barcode/` | P1 |
| EAN, UPC | Barcode/BarcodeEAN.cs, BarcodeUPC.cs | `barcode/` | P2 |
| QR Code | Barcode/QRCode/*.cs (13 files) | `barcode/qr/` | P1 |
| Aztec Code | Barcode/Aztec/*.cs (15 files) | `barcode/aztec/` | P2 |
| DataMatrix | Barcode/BarcodeDatamatrix.cs | `barcode/` | P2 |
| PDF417 | Barcode/BarcodePDF417.cs | `barcode/` | P2 |
| Swiss QR Code | Barcode/SwissQRCode.cs | `barcode/` | P2 |
| Code93 | Barcode/Barcode93.cs | `barcode/` | P3 |
| Interleaved 2of5 | Barcode/Barcode2of5.cs | `barcode/` | P3 |
| Codabar | Barcode/BarcodeCodabar.cs | `barcode/` | P3 |
| PostNet | Barcode/BarcodePostNet.cs | `barcode/` | P3 |
| MSI/Plessey | Barcode/BarcodeMSI.cs, BarcodePlessey.cs | `barcode/` | P3 |
| Pharmacode | Barcode/BarcodePharmacode.cs | `barcode/` | P3 |
| MaxiCode | Barcode/BarcodeMaxiCode.cs | `barcode/` | P3 |
| Intelligent Mail | Barcode/BarcodeIntelligentMail.cs | `barcode/` | P3 |
| CrossView | CrossView/*.cs | `crossview/` | P3 |
| Gauge objects | Gauge/*.cs (19 files) | `gauge/` | P3 |

**Milestone**: Tables, matrices, and common barcodes render correctly.

---

### Phase 8: Polish & Extras (Weeks 37+)

| Task | Priority |
|------|----------|
| Report inheritance | P2 |
| Localization framework | P2 |
| Web serving (HTTP handlers for Go) | P2 |
| Additional SQL drivers | P2 |
| Performance optimization | P2 |
| Documentation & examples | P1 |
| Report encryption/password | P3 |
| NumToWords additional languages | P3 |

---

## 6. Detailed Module Breakdown

### 6.1 FRX Serialization (Critical Path)

**Source**: `Utils/FRReader.cs` (~500 lines), `Utils/FRWriter.cs` (~400 lines), `Utils/RegisteredObjects.cs` (~1250 lines)

The FRX format is XML-based. Key behaviors to replicate:

1. **Delta serialization**: Compare object properties against a default instance; only write differing values
2. **Type resolution**: Object types referenced by class name in XML, resolved via `RegisteredObjects.FindType(typeName)`
3. **Reference handling**: Objects can reference other objects by name (e.g., DataBand.DataSource)
4. **Property types**: Handles primitives, enums, fonts, colors, images (base64), nested objects
5. **Collections**: Band.Objects, Report.Pages, etc. serialized as nested XML elements
6. **Compression**: Optional GZIP compression of FRX content
7. **Password protection**: Optional AES encryption

**RegisteredObjects pattern (critical for deserialization)**:

The .NET codebase uses a static `RegisteredObjects` class with a `Hashtable` mapping type names to `Type` objects. During `AssemblyInitializer`, ~150+ types are registered. When `FRReader` encounters an XML element name, it calls `RegisteredObjects.FindType(name)` to instantiate the object via `Activator.CreateInstance(type)`.

```go
// Go equivalent: type registry
var registry = map[string]func() Serializable{
    "Report":          func() Serializable { return NewReport() },
    "ReportPage":      func() Serializable { return NewReportPage() },
    "DataBand":        func() Serializable { return NewDataBand() },
    "TextObject":      func() Serializable { return NewTextObject() },
    "PictureObject":   func() Serializable { return NewPictureObject() },
    // ... ~150 more types
}

// FRReader uses it during deserialization:
func (r *FRReader) Read() Serializable {
    factory, ok := registry[r.CurrentItemName()]
    if !ok {
        return nil // unknown type
    }
    obj := factory()
    obj.Deserialize(r)
    return obj
}
```

**Go implementation notes**:
- Use `encoding/xml` for basic XML parsing
- Build a type registry: `map[string]func() Serializable` for object creation
- Each type registers itself via `init()` or explicit `Register()` call
- Implement `Serialize(w *FRWriter)` and `Deserialize(r *FRReader)` on all types
- For delta serialization, each type needs a `NewDefault()` factory
- The registry also supports hierarchical categories (for designer UI) — not needed for Go, only the `FindType` lookup is critical

### 6.2 Expression Engine (Critical Path)

**Source**: `Code/CodeUtils.cs`, `Code/CodeProvider.cs`, `Code/ExpressionDescriptor.cs`

The expression system is deeply embedded throughout the codebase. Every `TextObject.Text` can contain `[expressions]`.

**Expression types to support**:
- Data field: `[DataSource.FieldName]`
- Parameter: `[ParameterName]`
- System variable: `[PageNumber]`, `[TotalPages]`, `[Date]`, `[Time]`
- Totals: `[Total.TotalName]`
- Arithmetic: `[Price * Quantity]`
- String concat: `[FirstName] + " " + [LastName]`
- Conditional: `[IIF(Condition, TrueVal, FalseVal)]`
- Functions: `[Format("{0:C}", Price)]`, `[DateDiff(...)]`

**Recommended Go approach**:
```go
// Using expr-lang/expr
program, err := expr.Compile("[Price] * [Quantity]", expr.Env(env))
result, err := expr.Run(program, env)
```

### 6.3 Report Engine (Most Complex)

The engine is 19 partial class files (~3000 lines total). Key state:

```go
type ReportEngine struct {
    report          *Report
    curX, curY      float32      // Current render position
    curColumn       int          // Multi-column index
    curBand         BandBase     // Currently processing band
    page            *ReportPage  // Current source page
    preparedPages   *PreparedPages
    pageNo          int          // Logical page number
    totalPages      int          // Total pages (2nd pass)
    rowNo           int          // Row within current band
    absRowNo        int          // Absolute row number
    finalPass       bool         // Second pass flag
    keeping         bool         // Keep-together active
    hierarchyLevel  int          // Hierarchy depth
    // ...
}
```

**Critical algorithms**:
1. Band height calculation with CanGrow/CanShrink
2. Page break detection and band splitting
3. Keep-together buffering (start/end keep)
4. Group tree building and traversal
5. Multi-column rendering (AcrossThenDown, DownThenAcross)
6. Two-pass rendering for TotalPages
7. Deferred text processing (ProcessAt = ReportFinished)

### 6.4 Data Binding

**BusinessObjectDataSource** (most relevant for Go) binds Go data:

```go
// Go-native data binding
type BusinessObjectSource struct {
    DataSourceBase
    data     interface{} // slice, map, or struct
    current  int
    columns  []Column
}

// Usage
report.RegisterData("Products", products) // []Product slice
```

The C# version uses reflection to enumerate `IEnumerable`. The Go version should use `reflect` to iterate slices/arrays and extract struct fields as columns.

---

## 7. Risk Assessment & Gaps

### 7.1 High Risk Areas

| Area | Risk | Mitigation |
|------|------|------------|
| **Expression engine** | C# compiles real code; Go needs interpreter | Use expr-lang/expr (pure Go); limit to data expressions, not full scripting |
| **Font metrics** | Text wrapping depends on font measurement | Use pure Go font libraries (`golang.org/x/image/font`); may have subtle differences |
| **Text layout** | CanGrow requires knowing rendered text height | Implement text measurement with Go font libraries (pure Go, no CGo) |
| **PDF rendering** | PdfSimple is basic; complex layouts may need work | Port PdfSimple as-is (self-contained, no external deps); enhance incrementally |
| **FRX compatibility** | Must load existing .frx files created by .NET version | Extensive testing with sample .frx files |
| **Floating-point precision** | Layout calculations are pixel-precise | Use float32 consistently; test against reference output |

### 7.2 Missing from Plan (Gaps Identified)

| Gap | Source Location | Impact | Action |
|-----|----------------|--------|--------|
| **Report inheritance** | Report.BaseReport property | Medium | Phase 8 — load base report and merge |
| **Dialog pages** | DialogPage.cs (OpenSource has stub) | Low | Not needed for headless Go |
| **Print system** | PrintSettings, PrinterSetup | Low | Not applicable — Go targets export |
| **Designer integration** | TypeConverters/, design-time attributes | None | Not porting — Go has no visual designer |
| **SVG objects** | Not present in OpenSource | None | N/A |
| **Map objects** | Not present in OpenSource | None | N/A |
| **Chart objects** | Not present in OpenSource (separate plugin) | Medium | Could add via Go charting library later |
| **Rich text (RTF)** | RichObject in commercial version | None | Not in OpenSource |
| **.NET Script events** | BeforePrint/AfterPrint as C# script code | High | Replace with Go callback registration |
| **Report compilation** | Compiling report into .NET assembly | Medium | Use interpreted expressions instead |
| **Encryption/password** | Report password protection | Low | Phase 8 — use Go crypto |
| **OLAP/Cube** | CubeSourceBase, SliceCubeSource | Low | Phase 8 or skip |
| **BackPage** | Interleaving back pages | Low | Phase 8 |
| **BandColumns detail** | Complex multi-column layout edge cases | Medium | Test thoroughly in Phase 5 |
| **Extras database connectors** | 17 database connector plugins | Medium | Port top 4 (Postgres, MySQL, SQLite, MSSQL) in Phase 2 |
| **NumToWords localization** | 15 language files for number-to-words | Low | Port English first, add others as needed |
| **Report builder** | Programmatic report construction API | Medium | Natural in Go — just struct initialization |
| **Prepared page format** | Binary format for .fpx files | Medium | Phase 6 — needed for large reports |
| **Localization (28 languages)** | Localization/*.frl files | Low | Phase 8 |
| **Web components** | FastReport.Core.Web (108 files) | Medium | Phase 8 — HTTP handlers for Go |

### 7.3 C# Features Without Direct Go Equivalent

| C# Feature | Used In | Go Strategy |
|-------------|---------|-------------|
| `virtual`/`override` methods | Throughout hierarchy | Interface dispatch or callback fields |
| `abstract` classes | Base, ComponentBase, etc. | Embedded structs + interfaces |
| C# `event` delegates | BeforePrint, AfterPrint, etc. | `func` callback fields |
| `dynamic` dispatch | Expression evaluation | `interface{}` / `any` with type switches |
| Attributes `[Category]` | Designer metadata | Not needed |
| Partial classes | ReportEngine (19 files) | Multiple .go files, same package |
| Nullable types | Various properties | Pointer types or `*T` |
| LINQ | Data filtering/sorting | Go loops or `slices` package |
| Reflection for serialization | FRWriter/FRReader | Go `reflect` package |
| Async/await | Throughout | Not needed (Go is already concurrent) |

---

## 8. Testing Strategy

### 8.1 Unit Tests

Each Go package should have comprehensive unit tests:

```
report/     → base type creation, property get/set, hierarchy
serial/     → round-trip FRX load/save
data/       → data source iteration, filtering, sorting
expr/       → expression parsing and evaluation
engine/     → band processing, page breaks, groups
export/html → HTML output verification
export/pdf  → PDF output verification
barcode/    → barcode encoding correctness
format/     → number/date formatting
```

### 8.2 Integration Tests

- **FRX compatibility**: Load sample `.frx` files from the .NET version, verify all properties parsed
- **Engine output**: Compare prepared pages output against .NET reference output
- **HTML export**: Compare HTML export output against .NET reference
- **PDF export**: Verify PDF structure and content against .NET PdfSimple output

### 8.3 Sample Reports

Create a test suite of `.frx` files covering:
- Simple list report (DataBand + TextObject)
- Master-detail report (nested DataBands)
- Grouped report (GroupHeader/Footer)
- Multi-page report with headers/footers
- Report with expressions, totals, parameters
- Table report
- Matrix/pivot report
- Barcode report
- Report with subreports
- Multi-column report
- Report with conditional formatting

### 8.4 Benchmark Tests

- Report loading time (FRX parsing)
- Report preparation time (engine)
- Export time (HTML, image)
- Memory usage for large reports

---

## 9. Dependencies & Third-Party Libraries

### 9.1 .NET Dependencies (what the original uses)

The FastReport .NET codebase relies on these external packages:

**Core (FastReport.Base + FastReport.Compat):**

| .NET Package | Version | Purpose |
|-------------|---------|---------|
| `System.Drawing.Common` | >= 4.7.3 | 2D graphics, fonts, colors, images |
| `Microsoft.CodeAnalysis.CSharp` | >= 4.0.1 | Runtime C# expression compilation |
| `Microsoft.CodeAnalysis.VisualBasic` | >= 4.0.1 | Runtime VB.NET compilation |

**Database Connectors (Extras/):**

| .NET Package | Version | Purpose |
|-------------|---------|---------|
| `System.Data.SqlClient` | 4.8.6 | MS SQL Server |
| `MySqlConnector` | >= 2.4.0 | MySQL |
| `Npgsql` | >= 8.0.3 | PostgreSQL |
| `System.Data.SQLite.Core` | >= 1.0.115.5 | SQLite |
| `Oracle.ManagedDataAccess.Core` | >= 2.19.240 | Oracle |
| `FirebirdSql.Data.FirebirdClient` | 10.0.0 | Firebird |
| `System.Data.Odbc` | 6.0.0 | ODBC |
| `MongoDB.Driver` | >= 2.20.0 | MongoDB |
| `RavenDB.Client` | 4.0.6 | RavenDB |
| `CouchbaseNetClient` | 2.7.15 | Couchbase |
| `Apache.Ignite` | 2.17.0 | Apache Ignite |
| `ClickHouse.Client` | 3.2.0 | ClickHouse |
| `CassandraCSharpDriver` | 3.17.1 | Cassandra |
| `Newtonsoft.Json` | >= 13.0.3 | JSON data handling |
| `DocumentFormat.OpenXml` | 3.0.2 | Excel data source |
| `SpreadsheetLight` | 3.5.0 | Excel manipulation |
| `Google.Apis.Sheets.v4` | 1.68.0 | Google Sheets |

**Plugins:**

| .NET Package | Version | Purpose |
|-------------|---------|---------|
| `SkiaSharp` | >= 2.88.9 | WebP image support |
| `HarfBuzz` | >= 2.88.9 | Text shaping (SkiaSharp dependency) |

### 9.2 Pure Go Constraint

**All dependencies must be pure Go (no CGo) to minimize compilation complexity.** This means:
- No `github.com/mattn/go-sqlite3` (requires CGo + C compiler)
- No `SkiaSharp` or `System.Drawing` equivalents
- No image export (removed from scope — HTML and PDF only)
- PDF is ported from PdfSimple source (pure Go, no external PDF library needed)
- Barcodes are ported from .NET source (pure Go, no external barcode library needed)

### 9.3 Go Standard Library (no external dependency needed)

| Go Package | Purpose | Replaces |
|-----------|---------|----------|
| `encoding/xml` | FRX XML parsing | Custom XML |
| `encoding/json` | JSON data source | Newtonsoft.Json |
| `encoding/csv` | CSV data source | CsvUtils |
| `compress/gzip` | FRX compression | System.IO.Compression |
| `compress/flate` | ZIP support | System.IO.Compression |
| `crypto/aes` | Report encryption | System.Security.Cryptography |
| `hash/crc32` | CRC32 checksum | Custom Crc32.cs |
| `image`, `image/png`, `image/jpeg` | Image I/O (for embedded images in HTML/PDF) | System.Drawing imaging |
| `database/sql` | SQL data connections | System.Data |
| `html/template` | HTML export templates | String-based HTML |
| `fmt` | Value formatting | String.Format |
| `reflect` | Business object reflection | System.Reflection |
| `sort` | Data sorting | LINQ OrderBy |
| `strings`, `strconv`, `unicode` | String utilities | System.String |
| `math` | Math functions | System.Math |
| `io`, `bufio`, `bytes` | Stream/buffer I/O | System.IO |
| `net/url` | URL handling for hyperlinks | System.Uri |

### 9.4 Required Go Third-Party Dependencies (all pure Go)

| Library | Pure Go | Purpose | Replaces | Used By |
|---------|---------|---------|----------|---------|
| `github.com/expr-lang/expr` | Yes | Expression evaluation (compile-once, evaluate-many, sandboxed) | Microsoft.CodeAnalysis.CSharp (Roslyn) | `expr/` |
| `golang.org/x/image/font` | Yes | Font face interface, text measurement | System.Drawing.Graphics.MeasureString | `utils/`, `engine/` |
| `golang.org/x/image/font/opentype` | Yes | OpenType/TrueType font loading | System.Drawing.FontFamily | `utils/` |
| `golang.org/x/image/math/fixed` | Yes | Fixed-point math for font metrics | System.Drawing internals | `utils/` |

**Total external dependencies: 2** (`expr-lang/expr` + `golang.org/x/image`)

### 9.5 Optional Go Dependencies (database drivers — all pure Go)

| Library | Pure Go | Purpose | Replaces |
|---------|---------|---------|----------|
| `github.com/jackc/pgx/v5` | Yes | PostgreSQL | Npgsql |
| `github.com/go-sql-driver/mysql` | Yes | MySQL | MySqlConnector |
| `modernc.org/sqlite` | Yes | SQLite (pure Go, no CGo) | System.Data.SQLite.Core |
| `github.com/microsoft/go-mssqldb` | Yes | MS SQL Server | System.Data.SqlClient |

### 9.6 Key Dependency Decisions

| Decision | Chosen | Rationale |
|----------|--------|-----------|
| **Expression engine** | `expr-lang/expr` (pure Go) | Mature, fast, sandboxed; supports custom functions; single dependency |
| **Font metrics** | `golang.org/x/image/font` (pure Go) | Standard Go font interface; part of the Go project; no CGo |
| **PDF generation** | Port PdfSimple from .NET source | Self-contained (22 files); clean PDF primitives; no external deps; pure Go |
| **Barcode generation** | Port from .NET source | Exact rendering compatibility with .NET; pure Go; no external deps |
| **QR Code** | Port from .NET source (13 files) | .NET implementation is well-structured; ensures FRX compatibility; pure Go |
| **Image export** | **Excluded** | Requires heavy graphics deps (CGo or large libs); HTML+PDF covers use cases |
| **SQLite driver** | `modernc.org/sqlite` (pure Go) | No CGo; no C compiler needed; simpler cross-compilation |
| **HTML export** | Port from .NET source | String-based HTML generation; no external deps; pure Go |

### 9.7 Dependency Summary

```
go-fastreport/
├── go.mod
│   require (
│       github.com/expr-lang/expr v1.x    // expression evaluation
│       golang.org/x/image v0.x            // font metrics
│   )
│   // Optional (imported by user's code, not by core library):
│   // github.com/jackc/pgx/v5           // PostgreSQL driver
│   // github.com/go-sql-driver/mysql    // MySQL driver
│   // modernc.org/sqlite                // SQLite driver
│   // github.com/microsoft/go-mssqldb   // MSSQL driver
```

**Zero CGo. Two required dependencies. Cross-compiles to any GOOS/GOARCH.**

---

## 10. Out of Scope

The following are **not** included in the Go port:

### 10.1 Excluded Features

1. **Visual Report Designer** — The .NET version has a WinForms/WPF designer. Go will be code-first / FRX-file-based only.
2. **Dialog pages** — Interactive forms before report execution (WinForms-specific).
3. **Print system** — Direct printer output. Go targets export formats only (HTML + PDF).
4. **TypeConverters** — Designer-specific property editors (12 files).
5. **WinForms/WPF preview** — Desktop GUI preview components.
6. **VB.NET script support** — Only C#-style expressions were used; Go uses its own expression engine.
7. **Full .NET scripting** — No arbitrary code execution in reports; only expressions via `expr-lang/expr`.
8. **FastReport Commercial features** — Only OpenSource features are ported.
9. **Import from other report formats** — RDL, DevExpress, JasperReports, ListAndLabel, StimulSoft importers (15 files) — permanently excluded.
10. **Backward compatibility layer** — FastReport.Compat (27 files) is .NET-specific.

### 10.2 Excluded Export Formats

11. **Image export** (PNG, JPEG, BMP, GIF, TIFF, EMF) — Requires heavy graphics rendering dependencies; HTML and PDF cover primary use cases.
12. **Excel export** — Not present in OpenSource.
13. **Word export** — Not present in OpenSource.

### 10.3 Excluded Data Sources

14. **OLAP/Cube sources** — CubeSourceBase, SliceCubeSource, CubeHelper, BaseCubeLink (5 files) — niche feature.
15. **Google Sheets** — Requires external API deps.
16. **Excel data source** — Requires OpenXML/SpreadsheetLight deps.
17. **NoSQL databases** — MongoDB, RavenDB, Couchbase, Cassandra, Elasticsearch, Ignite, ClickHouse — can be added as separate modules later.
18. **ODBC** — Platform-specific.

### 10.4 Excluded Infrastructure

19. **Web components** — FastReport.Core.Web (108 files) — ASP.NET-specific; Go users can serve HTML export directly.
20. **Async partial classes** — 12+ `.Async.cs` files — Go goroutines handle concurrency natively.
21. **FastReport.OpenSource partial overrides** — 39 files — .NET distribution layer, not applicable.
22. **Fakes.cs, AssemblyInitializer.cs** — .NET-specific bootstrapping.
23. **CGo dependencies** — Any library requiring CGo compilation is excluded.

### 10.5 Excluded .NET Files by Directory

| Directory | Files Excluded | Reason |
|-----------|---------------|--------|
| FastReport.Compat/ | 27 | .NET Framework compatibility |
| FastReport.Core.Web/ | 108 | ASP.NET Core web UI |
| FastReport.Web.Base/ | 3 | Web base classes |
| FastReport.OpenSource/ | 39 | .NET partial class overrides |
| Extras/Core/FastReport.Data/ (NoSQL) | ~60 | NoSQL database connectors |
| Extras/Core/FastReport.Plugin/ | ~5 | WebP/SkiaSharp plugins |
| Import/ | 15 | Report format importers |
| TypeConverters/ | 12 | Designer-only |
| Demos/ | 42 | .NET demo apps |
| Pack/ | 10 | .NET packaging |
| Tools/ | 4 | .NET test tools |
| **Total excluded** | **~325** | |

---

## Appendix A: File Count Summary

| Category | .NET Files | Go Files (est.) | Notes |
|----------|-----------|----------------|-------|
| Core objects | 84 | ~40 | Simpler without designer/async |
| Data system | 54 | ~30 | Core sources + SQL connectors |
| Engine | 19 | ~14 | Direct port, no async variants |
| Serialization + Registry | 2 + 1 | ~4 | FRX reader/writer + type registry |
| Expressions | 12 | ~4 | Simplified with expr library |
| Exports (HTML + PDF only) | 7 + 22 | ~16 | No image export |
| Barcodes | 57 | ~35 | All formats including QR/Aztec |
| Table | 13 | ~8 | Including all supporting files |
| Matrix | 11 | ~6 | |
| Format | 10 | ~9 | |
| Preview | 8 | ~7 | Including postprocessor, blobstore |
| Utils | 57 | ~20 | Font mgmt, text renderer (pure Go) |
| Functions | 18 | ~5 | English + core functions |
| Gauge | 19 | ~4 | Simplified |
| CrossView | 8 | ~2 | |
| Style | 8 | ~7 | |
| **Total** | **~400** | **~211** | Pure Go, zero CGo |

## Appendix B: Key Source Files to Read First

When starting implementation, read these files in order:

1. `FastReport.Base/Base.cs` — Foundation of all objects
2. `FastReport.Base/IFRSerializable.cs` — Serialization interface
3. `FastReport.Base/IParent.cs` — Tree structure interface
4. `FastReport.Base/ComponentBase.cs` — Positioning
5. `FastReport.Base/ReportComponentBase.cs` — Styling
6. `FastReport.Base/Report.cs` — Top-level container
7. `FastReport.Base/ReportPage.cs` — Page configuration
8. `FastReport.Base/BandBase.cs` — Band foundation
9. `FastReport.Base/DataBand.cs` — Data-driven bands
10. `FastReport.Base/TextObject.cs` — Most-used object
11. `FastReport.Base/Utils/FRReader.cs` — FRX loading
12. `FastReport.Base/Utils/FRWriter.cs` — FRX saving
13. `FastReport.Base/Data/Dictionary.cs` — Data registry
14. `FastReport.Base/Data/DataSourceBase.cs` — Data abstraction
15. `FastReport.Base/Engine/ReportEngine.cs` — Engine entry point
16. `FastReport.Base/Engine/ReportEngine.Pages.cs` — Page flow
17. `FastReport.Base/Engine/ReportEngine.Bands.cs` — Band rendering
18. `FastReport.Base/Engine/ReportEngine.DataBands.cs` — Data iteration
19. `FastReport.Base/Engine/ReportEngine.Groups.cs` — Grouping
20. `FastReport.Base/Export/ExportBase.cs` — Export framework

## Appendix C: Enums to Port

Key enumerations found across the codebase:

| Enum | Values | Used By |
|------|--------|---------|
| `HorzAlign` | Left, Center, Right, Justify | TextObject |
| `VertAlign` | Top, Center, Bottom | TextObject |
| `ImageSizeMode` | Normal, AutoSize, StretchImage, Zoom, CenterImage | PictureObject |
| `Duplicates` | Show, Merge, Clear, Hide | TextObject |
| `ProcessAt` | Default, DataFinished, ReportFinished, PageFinished, ColumnFinished, GroupFinished, Custom | TextObject |
| `PrintOn` | FirstPage, LastPage, OddPages, EvenPages, SinglePage, RepeatedBand | HeaderFooterBandBase |
| `SortOrder` | None, Ascending, Descending | Sort, GroupHeaderBand |
| `TotalType` | Sum, Min, Max, Avg, Count, CountDistinct | Total |
| `FillType` | Solid, LinearGradient, PathGradient, Hatch, Glass | Fills |
| `LineStyle` | Solid, Dash, Dot, DashDot, DashDotDot, Double | BorderLine |
| `ShapeKind` | Rectangle, RoundRectangle, Ellipse, Triangle, Diamond | ShapeObject |
| `BarcodeType` | Code128, Code39, EAN13, QRCode, Aztec, DataMatrix, PDF417, ... | BarcodeObject |
| `PageUnits` | Millimeters, Centimeters, Inches, HundrethsOfInch | Units |
| `PaperSize` | Various standard paper sizes | ReportPage |
| `PageOrientation` | Portrait, Landscape | ReportPage |

---

## Appendix D: Validation Gap Analysis

A thorough comparison of this porting plan against every file in the actual codebase identified the following gaps. Each is categorized as either addressed (already in plan), deferred (Phase 8+), or out of scope.

### D.1 Data System Gaps (18 files not explicitly listed)

| File | Description | Action |
|------|-------------|--------|
| `Data/DataComponentBase.cs` | Base class for connections/datasources | Phase 2 — needed for data hierarchy |
| `Data/ProcedureDataSource.cs` | Stored procedure data source | Phase 2 P2 — port with SQL connections |
| `Data/ProcedureParameter.cs` | Stored procedure parameters | Phase 2 P2 |
| `Data/ViewDataSource.cs` | DataView-based data source | Phase 8 — low priority |
| `Data/VirtualDataSource.cs` | Custom virtual data source | Phase 8 — low priority |
| `Data/XmlDataConnection.cs` | XML file data source | Phase 2 P2 — add `data/xml/` package |
| `Data/XmlConnectionStringBuilder.cs` | XML connection config | Phase 2 P2 |
| `Data/CommandParameter.cs` | SQL command parameters | Phase 2 P1 — needed for parameterized queries |
| `Data/CommandParameterCollection.cs` | Collection for command params | Phase 2 P1 |
| `Data/ConnectionCollection.cs` | Multi-connection management | Phase 2 P0 — needed for Dictionary |
| `Data/DataSourceCollection.cs` | Collection type | Phase 2 P0 |
| `Data/TableCollection.cs` | Tables in a connection | Phase 2 P1 |
| `Data/CsvConnectionStringBuilder.cs` | CSV config | Phase 2 P1 |
| `Data/CsvUtils.cs` | CSV utilities | Phase 2 P1 |
| `Data/DictionaryHelper.cs` | Registration helper | Phase 2 P1 |
| `Data/DbConnectionExtensions.cs` | DB extensions | Phase 2 P2 |
| `Data/BusinessObjectConverter.cs` | Object schema conversion | Phase 2 P1 |
| `Data/IJsonProviderSourceConnection.cs` | JSON provider interface | Phase 2 P1 |

### D.2 Barcode Types Not Listed (12 additional)

The plan listed 8 barcode types. The codebase contains 20+ total:

| Barcode | File | Priority |
|---------|------|----------|
| Barcode2DBase | `Barcode/Barcode2DBase.cs` | P1 — base class, required |
| Code 93 | `Barcode/Barcode93.cs` | P3 |
| Interleaved 2of5 | `Barcode/Barcode2of5.cs` | P3 |
| Codabar | `Barcode/BarcodeCodabar.cs` | P3 |
| GS1 DataBar | `Barcode/BarcodeGS1.cs` | P3 |
| Intelligent Mail | `Barcode/BarcodeIntelligentMail.cs` | P3 |
| MSI/Plessey | `Barcode/BarcodeMSI.cs` | P3 |
| MaxiCode | `Barcode/BarcodeMaxiCode.cs` | P3 |
| Pharmacode | `Barcode/BarcodePharmacode.cs` | P3 |
| Plessey | `Barcode/BarcodePlessey.cs` | P3 |
| PostNet | `Barcode/BarcodePostNet.cs` | P3 |
| Swiss QR Code | `Barcode/SwissQRCode.cs` | P2 |
| GS1Helper | `Barcode/GS1Helper.cs` | P2 |

### D.3 Table Supporting Files (6 files not listed)

| File | Description | Action |
|------|-------------|--------|
| `Table/TableCellData.cs` | Cell data container | Phase 7 — needed for table rendering |
| `Table/TableColumnCollection.cs` | Column collection | Phase 7 |
| `Table/TableRowCollection.cs` | Row collection | Phase 7 |
| `Table/TableHelper.cs` | Table utility functions | Phase 7 |
| `Table/TableResult.cs` | Result type | Phase 7 |
| `Table/TableStyleCollection.cs` | Style collection | Phase 7 |

### D.4 Utils Infrastructure Gaps (27 files)

**Critical for Phase 1:**

| File | Description | Action |
|------|-------------|--------|
| `Utils/RegisteredObjects.cs` | Object type registry for deserialization | Phase 1 P0 — critical for FRX loading |
| `Utils/FRCollectionBase.cs` | Base collection class | Phase 1 P0 |
| `Utils/BlobStore.cs` | Binary data storage in prepared pages | Phase 5 P1 |
| `Utils/Compressor.cs` | Data compression | Phase 1 P1 |
| `Utils/Zip.cs` | ZIP format support | Phase 1 P1 |
| `Utils/Crypter.cs` | String encryption/decryption | Phase 8 P2 |
| `Utils/Crc32.cs` | CRC32 checksum | Phase 1 P1 |

**Font/Text Rendering (Critical for Engine):**

| File | Description | Action |
|------|-------------|--------|
| `Utils/FontManager.cs` | Font management system | Phase 3 P0 — needed for text layout |
| `Utils/FontManager.Gdi.cs` | GDI font measurement | Phase 3 P0 |
| `Utils/FontManager.Internals.cs` | Font internals | Phase 3 P0 |
| `Utils/FRPrivateFontCollection.cs` | Custom font loading | Phase 3 P1 |
| `Utils/TextRenderer.cs` | Text rendering/measurement | Phase 3 P0 |
| `Utils/HtmlTextRenderer.cs` | HTML text rendering | Phase 3 P1 |
| `Utils/GraphicCache.cs` | Graphics object caching | Phase 5 P1 |
| `Utils/ImageHelper.cs` | Image loading utilities | Phase 3 P1 |

**Other Utils:**

| File | Description | Action |
|------|-------------|--------|
| `Utils/StorageService.cs` | Storage abstraction | Phase 8 |
| `Utils/FastNameCreator.cs` | Unique name generation | Phase 1 P1 |
| `Utils/ShortProperties.cs` | Property name compression | Phase 5 P2 |
| `Utils/Res.cs` | Localized resources | Phase 8 |
| `Utils/ResourceLoader.cs` | Resource loading | Phase 8 |
| `Utils/CompilerSettings.cs` | Compiler config | Out of scope |
| `Utils/ScriptSecurityEventArgs.cs` | Script security | Out of scope |
| `Utils/ExportsOptions.cs` | Export settings | Phase 6 P1 |
| `Utils/Validator.cs` | Validation | Phase 1 P2 |
| `Utils/FRPaintEventArgs.cs` | Paint events | Phase 5 P2 |
| `Utils/FRRandom.cs` | Random numbers | Phase 1 P2 |
| `Utils/MyEncodingInfo.cs` | Encoding info | Phase 1 P2 |

### D.5 Preview/PreparedPages Gaps (3 files)

| File | Description | Action |
|------|-------------|--------|
| `Preview/Dictionary.cs` | Preview-specific dictionary | Phase 5 P1 |
| `Preview/PreparedPagePostprocessor.cs` | Post-processing of rendered pages | Phase 5 P1 |
| `Preview/SourcePages.cs` | Source page tracking | Phase 5 P1 |

### D.6 NumToWords Localization (15 language files)

Only English is in scope for initial port. Additional languages deferred:

| Language | File | Phase |
|----------|------|-------|
| German | `Functions/NumToWordsDe.cs` | Phase 8 |
| British English | `Functions/NumToWordsEnGb.cs` | Phase 8 |
| Spanish | `Functions/NumToWordsEs.cs` | Phase 8 |
| French | `Functions/NumToWordsFr.cs` | Phase 8 |
| Indian | `Functions/NumToWordsIn.cs` | Phase 8 |
| Dutch | `Functions/NumToWordsNl.cs` | Phase 8 |
| Persian | `Functions/NumToWordsPersian.cs` | Phase 8 |
| Polish | `Functions/NumToWordsPl.cs` | Phase 8 |
| Russian | `Functions/NumToWordsRu.cs` | Phase 8 |
| Ukrainian | `Functions/NumToWordsUkr.cs` | Phase 8 |
| Spanish variant | `Functions/NumToWordSp.cs` | Phase 8 |
| Base classes | `Functions/NumToWordsBase.cs`, `NumToLettersBase.cs` | Phase 4 P2 |

### D.7 Gauge Subdirectories (not enumerated in plan)

| Directory | Files | Action |
|-----------|-------|--------|
| `Gauge/Linear/` | LinearGauge, LinearScale, LinearPointer, LinearLabel | Phase 7 P3 |
| `Gauge/Radial/` | RadialGauge, RadialScale, RadialPointer, RadialLabel, RadialUtils | Phase 7 P3 |
| `Gauge/Simple/` | SimpleGauge, SimpleScale, SimplePointer, SimpleProgressGauge, etc. | Phase 7 P3 |

### D.8 Async Partial Classes (12+ files)

All `.Async.cs` files throughout the codebase are **out of scope** for Go. Go's goroutine model makes async patterns unnecessary. These files include:
- `Report.Async.cs`, `BandBase.Async.cs`, `DataBand.Async.cs`
- `GroupHeaderBand.Async.cs`, `TextObject.Async.cs`, `HtmlObject.Async.cs`
- `PictureObject.Async.cs`, `ContainerObject.Async.cs`, `CheckBoxObject.Async.cs`
- `ReportEngine.Async.cs` and all engine async variants
- `MatrixObject.Async.cs`, `TableObject.Async.cs`, `TableCell.Async.cs`

### D.9 Import System (deferred)

All 5 importers (RDL, DevExpress, JasperReports, ListAndLabel, StimulSoft) plus supporting infrastructure (ImportBase, ComponentsFactory, SizeUnits, UnitsConverter) are deferred to Phase 8 or beyond. The core `ImportBase.cs` pattern could be useful if import support is added later.

### D.10 FastReport.OpenSource Overrides

The OpenSource project adds open-source-specific code via partial classes:
- `Base.OpenSource.cs`, `Report.OpenSource.cs` — license/restriction handling
- `ExportBase.OpenSource.cs` — export stubs for commercial features
- `NetRepository.Core.cs` — repository pattern

These are .NET distribution concerns and **not applicable** to the Go port.

### D.11 CrossView/OLAP Gap

| File | Description | Action |
|------|-------------|--------|
| `CrossView/BaseCubeLink.cs` | OLAP cube integration | Phase 8 or out of scope |
| `Data/CubeSourceBase.cs` | OLAP cube base | Phase 8 or out of scope |
| `Data/SliceCubeSource.cs` | OLAP slice cube | Phase 8 or out of scope |
| `Data/CubeSourceCollection.cs` | Collection | Phase 8 or out of scope |
| `Data/CubeHelper.cs` | OLAP utilities | Phase 8 or out of scope |

---

## Appendix E: Summary of Validation Results

| Category | Total Files | Covered in Plan | Gaps Found | Gap Severity |
|----------|-------------|----------------|------------|--------------|
| Core objects (root) | 84 | 70+ | ~14 (async + collections) | Low — async not needed, collections implicit |
| Data system | 54 | 15 | 18 | Medium — add to Phase 2 |
| Engine | 19 | 19 | 0 (+ 6 async) | None |
| Export | 9 | 9 | 0 | None |
| Barcode | 57 | ~30 | 12 extra types | Low — P3 additions |
| Table | 13 | 5 | 6 supporting | Medium — add to Phase 7 |
| Matrix | 11 | 6 | 2 | Low |
| Gauge | 19 | 4 | subdirectory detail | Low — P3 |
| Utils | 57 | ~15 | 27 | **High** — font/text critical |
| Functions | 18 | 3 | 15 languages | Low — defer |
| Preview | 8 | 5 | 3 | Medium |
| Code | 12 | 5 | 5 (MS-specific) | None — out of scope |
| Import | 15 | 0 | 15 | None — deferred |
| CrossView | 8 | 1 | OLAP link | Low |
| Format | 10 | 10 | 0 | None |
| TypeConverters | 12 | 12 | 0 | None — out of scope |

**Critical action items from validation — ALL RESOLVED:**
1. ~~Add `Utils/RegisteredObjects.cs` to Phase 1 P0 (object type registry)~~ **FIXED** — Added to Phase 1 + Section 6.1
2. ~~Add `Utils/FontManager*.cs` and `Utils/TextRenderer.cs` to Phase 3 P0 (text measurement)~~ **FIXED** — Added to Phase 3
3. ~~Add `Data/CommandParameter*.cs` and `Data/DataComponentBase.cs` to Phase 2~~ **FIXED** — Added to Phase 2
4. ~~Add `Table/TableCellData.cs` and supporting files to Phase 7~~ **FIXED** — Added to Phase 7
5. ~~Add `Preview/PreparedPagePostprocessor.cs` and `Preview/SourcePages.cs` to Phase 5~~ **FIXED** — Added to Phase 5
6. ~~Document .NET third-party dependencies and Go equivalents~~ **FIXED** — Section 9 fully rewritten
7. ~~Add missing barcode types~~ **FIXED** — All 20+ types listed in Phase 7
8. ~~Add Utils infrastructure (compression, BlobStore, collections)~~ **FIXED** — Added across Phases 1, 3, 5
