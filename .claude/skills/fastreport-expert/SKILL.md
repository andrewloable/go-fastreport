---
name: fastreport-expert
description: Expert on the FastReport .NET codebase. Use when asking questions about FastReport architecture, object model, report engine internals, data binding, exports, bands, serialization, or when porting FastReport features to Go.
user_invocable: true
---

You are an expert on the FastReport Open Source .NET codebase located at `original-dotnet/`. Your role is to answer questions about the FastReport architecture, help port features to Go, and provide deep knowledge of how the report engine works internally.

## Reference Codebase Location

All source code is in `original-dotnet/` with these key projects:
- `FastReport.Base/` - Core foundation: all report objects, engine, utilities
- `FastReport.OpenSource/` - Open source distribution additions
- `FastReport.Compat/` - Backward compatibility layer
- `FastReport.Core.Web/` and `FastReport.Web.Base/` - Web components

## Architecture Overview

### Object Model Hierarchy

```
Base (IFRSerializable)
  └── ComponentBase (positioning: Top, Left, Width, Height)
       └── ReportComponentBase (styling, borders, fills)
            ├── TextObject, PictureObject, LineObject, ShapeObject
            ├── CheckBoxObject, BarcodeObject, ZipCodeObject
            ├── ContainerObject, SubreportObject
            ├── MatrixObject, TableObject, CrossViewObject
            └── [Other visual objects]
       └── BreakableComponent (grow/shrink, shift, page breaking)
            └── BandBase (IParent)
                 ├── ReportTitleBand, ReportSummaryBand
                 ├── PageHeaderBand, PageFooterBand
                 ├── ColumnHeaderBand, ColumnFooterBand
                 ├── DataHeaderBand, DataBand, DataFooterBand
                 ├── GroupHeaderBand, GroupFooterBand
                 ├── ChildBand, OverlayBand
                 └── TextObjectBase → TextObject, HtmlObject
  └── Report (IParent) - top-level container with Pages, Dictionary, Styles
```

### Key Source File Locations

| Area | Path |
|------|------|
| Core objects | `FastReport.Base/*.cs` (Report.cs, ReportPage.cs, BandBase.cs, TextObject.cs, etc.) |
| Data system | `FastReport.Base/Data/` (Dictionary.cs, DataSourceBase.cs, DataConnectionBase.cs) |
| Report engine | `FastReport.Base/Engine/ReportEngine*.cs` (partial classes for Bands, DataBands, Groups, Pages, Break, Subreports) |
| Exports | `FastReport.Base/Export/` and `FastReport.OpenSource/Export/` |
| Script/expressions | `FastReport.Base/Code/` (CodeProvider, ExpressionDescriptor, AssemblyDescriptor) |
| Serialization (FRX) | `FastReport.Base/Utils/FRReader.cs`, `FRWriter.cs` |
| Utilities | `FastReport.Base/Utils/` |
| Preview/prepared pages | `FastReport.Base/Preview/` |
| Barcodes | `FastReport.Base/Barcode/` |
| Matrix/pivot | `FastReport.Base/Matrix/` |
| Table | `FastReport.Base/Table/` |
| Gauge | `FastReport.Base/Gauge/` |
| CrossView | `FastReport.Base/CrossView/` |

### Report Execution Flow

1. `Report.Prepare()` - Entry point: compiles script, initializes data, runs engine
2. Engine iterates through pages and bands in order
3. For each band: fires BeforeLayout → renders objects → fires AfterLayout
4. DataBand iterates data source rows, applying filters and sorts
5. GroupHeaderBand detects group value changes, prints headers/footers
6. Creates `PreparedPages` collection with rendered page snapshots
7. PreparedPages can be previewed or exported

### Band Types (13 total)

| Band | Purpose |
|------|---------|
| ReportTitleBand | Once at report start |
| ReportSummaryBand | Once at report end |
| PageHeaderBand | Top of each page |
| PageFooterBand | Bottom of each page |
| ColumnHeaderBand | Top of each column (multi-column) |
| ColumnFooterBand | Bottom of each column |
| DataHeaderBand | Before data rows |
| DataBand | Repeats per data source row |
| DataFooterBand | After data rows |
| GroupHeaderBand | At group value change (nestable) |
| GroupFooterBand | Group summary |
| ChildBand | After another band |
| OverlayBand | On top of other content |

### Data System

**Dictionary** (Report.Dictionary) contains:
- **Connections** - Database connections (SQL Server, MySQL, PostgreSQL, SQLite, etc.)
- **DataSources** - TableDataSource, ProcedureDataSource, BusinessObjectDataSource
- **Relations** - Master-detail relationships between data sources
- **Parameters** - Report parameters with expressions and nested support
- **SystemVariables** - PageNumber, TotalPages, Date, Time, etc.
- **Totals** - Aggregate calculations (Sum, Avg, Min, Max, Count)

**Data binding in bands:**
- DataBand.DataSource - connects to a data source
- Filter property - boolean expression for row filtering
- Sort collection - multiple sort conditions
- Relation property - master-detail relationships
- Hierarchical data via IdColumn/ParentIdColumn

### Serialization (FRX Format)

- XML-based format serialized via `FRWriter` / deserialized via `FRReader`
- All objects implement `IFRSerializable` with `Serialize(FRWriter)` and `Deserialize(FRReader)`
- Delta serialization: only properties differing from defaults are written
- Supports compression and password protection
- Methods: `Report.Load()`, `Report.Save()`, `Report.SaveToString()`, `Report.LoadFromString()`

### Expression and Script Engine

- Supports C# and VB.NET via `CodeProvider` abstraction
- Expressions use bracket syntax: `[DataSource.FieldName]`, `[ParameterName]`, `[PageNumber]`
- `Report.Calc(expression)` evaluates in current context
- Expressions used in: filters, sorts, visibility, printability, text content, conditional styles
- Script compiled to .NET assembly via Roslyn

### Export System

- **ExportBase** - Abstract base with PageRange, PageNumbers, OpenAfterExport
- Open source includes: HtmlExport, ImageExport (PNG, JPEG, BMP, GIF, TIFF, EMF)
- PDF via separate plugin (FastReport.OpenSource.Export.PdfSimple)
- Per-object exportability control via Exportable property

### Key Design Patterns

- **Composite**: IParent interface, tree hierarchy (Report → Pages → Bands → Objects)
- **Serialization**: IFRSerializable with delta serialization
- **Observer**: Events (BeforePrint, AfterPrint, AfterData, Click)
- **Template Method**: ReportEngine partial classes for different engine aspects
- **Factory**: DataSource/Export creation based on type

### Units System

- Internal measurements in screen pixels
- Conversion utilities for mm, inches, cm via `Units` class

## How To Use This Skill

When answering questions:
1. **Always read the actual source files** in `original-dotnet/` to provide accurate answers
2. For architecture questions, reference the specific `.cs` files
3. For porting to Go, explain the C# patterns and suggest idiomatic Go equivalents
4. For "how does X work" questions, trace through the execution flow in the engine code
5. Pay attention to async variants (`.Async.cs` suffix files)

When helping port to Go:
- Map C# class hierarchies to Go interfaces and struct embedding
- Replace C# events with Go callback functions or channels
- Replace C# properties with Go getter/setter methods or exported fields
- Map IFRSerializable to a Go Serializable interface
- Consider Go's lack of inheritance — use composition and interfaces
