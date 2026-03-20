# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

**go-fastreport** is a pure Go port of [FastReport .NET](https://github.com/FastReports/FastReport). It loads `.frx` report definitions, binds data sources, runs a band-based layout engine, and exports to HTML, PDF, PNG, CSV, XLSX, RTF, and SVG.

## Source of Truth

**The original C# source code in `original-dotnet/` is the single source of truth for all implementation decisions.** When porting features, fixing bugs, or implementing new functionality:

1. **Always read the corresponding C# code first** before writing Go code
2. **Follow the same processes and algorithms** as the C# implementation — do not invent alternative approaches
3. **Match C# behavior exactly** — the C# HTML/PDF output is the ground-truth for correctness
4. **Reference specific C# files and line numbers** in comments when porting non-obvious logic
5. **Use `tools/compare_html_semantic/`** to verify Go HTML output matches C# output

## Build & Test

```bash
go build ./...                                  # Build all packages
go test ./...                                   # Run full test suite (~98% coverage)
go test ./engine/...                            # Test a specific package
go test -v ./reportpkg/... -run TestFRXSmoke_   # Run a single test by name
go test -bench=. ./engine/...                   # Run benchmarks
go run ./examples/frx_to_html/                  # Run from repo root (needs test-reports/)
```

No Makefile, CI pipeline, or linter config exists. All builds use the standard `go` toolchain.

## Issue Tracking

This project uses **bd (beads)** for issue tracking, not markdown TODOs. See `AGENTS.md` for workflow commands (`bd ready`, `bd show`, `bd close`).

## Architecture

### Core Pipeline

```
.frx file
  → serial.Reader (XML decoder, attribute maps)
  → serial.DefaultRegistry.Create(typeName) (factory pattern)
  → obj.Deserialize(reader) (populate fields from attributes)
  → reportpkg.Report (pages, bands, objects, dictionary)
  → engine.ReportEngine.Run() (3-phase: init → execute → finish)
  → preview.PreparedPages (rendered snapshots: PreparedPage → PreparedBand → PreparedObject)
  → export/html, export/pdf, etc. (iterate PreparedPages via ExportBase interface)
```

### Key Packages

| Package | Role |
|---------|------|
| `report/` | Core interfaces: `Base`, `Serializable`, `Reader`, `Writer`, `Parent`, `ChildDeserializer` |
| `serial/` | FRX XML parser and **DefaultRegistry** — factory map of type name → constructor |
| `reportpkg/` | `Report` struct, FRX load/save, `Calc()`/`CalcText()` expression evaluation, Dictionary |
| `engine/` | `ReportEngine` — runs bands, populates `PreparedPages`, manages page breaks, system variables |
| `preview/` | `PreparedPages`, `PreparedBand`, `PreparedObject`, `BlobStore` — engine output consumed by exporters |
| `band/` | 13 band types: DataBand, GroupHeaderBand, PageHeader, PageFooter, ReportTitle, etc. |
| `object/` | Report objects: TextObject, PictureObject, LineObject, ShapeObject, CheckBoxObject, MSChartObject, AdvMatrixObject, etc. |
| `data/` | `DataSource` interface, `Dictionary` (parameters, system variables, totals). Subpackages: `json/`, `xml/`, `csv/`, `sql/` |
| `expr/` | Expression evaluator wrapping `expr-lang/expr`. `Env` map of variable bindings, compiled program cache |
| `style/` | `Border`, `Fill` (SolidFill, LinearGradientFill, etc.), `Font`, `StyleSheet` |
| `format/` | Value formatters: NumberFormat, DateFormat, CurrencyFormat, etc. |
| `functions/` | Built-in expression functions (math, string, date, conversion) |
| `export/` | `ExportBase` + subpackages: `html/`, `pdf/`, `image/`, `csv/`, `xlsx/`, `rtf/`, `svg/` |

### Key Patterns

**Serial Registry** (`serial/registry.go`): All deserializable types register a factory function at init time (e.g. `serial.DefaultRegistry.MustRegister("TextObject", func() report.Base { return object.NewTextObject() })`). During FRX load, `deserializeChildren()` calls `registry.Create(typeName)` for each XML element.

**ChildDeserializer** (`report/interfaces.go`): Objects that need to handle custom child XML elements (e.g. TextObject handles `<Highlight>`, `<Formats>`) implement `DeserializeChild(childType string, r Reader) bool`. Called by `deserializeChildren()` before falling back to the registry.

**Expression Evaluation** (`reportpkg/calc.go`): Bracket expressions like `[DataSource.Field]` are evaluated via `Report.Calc()` / `Report.CalcText()`. The environment is built from dictionary parameters, system variables, totals, current data row values, and custom functions. Dots in identifiers are sanitized to underscores (e.g. `Report.ReportInfo.Description` → `Report_ReportInfo_Description`).

**Engine → PreparedPages** (`engine/objects.go`): `buildPreparedObject()` is the central switch that converts each object type (TextObject, PictureObject, BarcodeObject, etc.) into a `preview.PreparedObject` snapshot. Text expressions are evaluated here via `evalTextWithFormat()`. Images are stored in `BlobStore` and referenced by index.

**FRX Attribute Names**: FRX .NET uses both short (`ReportDescription`) and dot-qualified (`ReportInfo.Description`) attribute names. The Go port reads both forms as fallbacks.

### C# → Go Translation Conventions

- C# partial classes → single Go file per type (e.g. `band/databand.go`)
- C# properties → Go getter/setter methods (e.g. `Visible()` / `SetVisible()`)
- C# inheritance → Go embedding (e.g. `TextObject` embeds `ReportComponentBase`)
- C# events → Go callback fields or engine state handlers
- C# `System.Drawing` → `image/color`, `style.Font`, `style.Border`
- C# `CodeDom` expression compilation → `expr-lang/expr` library with environment injection

### Test Reports

The `test-reports/` directory contains 100+ `.frx` files and `nwind.xml` (NorthWind sample data) from the original FastReport .NET distribution. These are used for FRX smoke tests and the `examples/frx_to_html` example.
