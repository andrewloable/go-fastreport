# GEMINI.md - go-fastreport Project Context

This project is a pure Go port of the core functionality of **FastReport .NET Open Source**. It provides a band-based reporting engine, multiple data source adapters, and export capabilities to HTML, PDF, PNG, and other formats.

## 🚀 Project Overview
- **Purpose**: A high-performance, pure Go reporting library compatible with FastReport `.frx` templates.
- **Main Technologies**: Go 1.25+, `github.com/expr-lang/expr` for expressions, `github.com/xuri/excelize/v2` for Excel export.
- **Architecture**:
    - `report/`: Core abstractions (`Base`, `Serializable`, `Parent`).
    - `reportpkg/`: `Report` container, serialization, and expression evaluation.
    - `engine/`: `ReportEngine` execution pipeline (Init → Execute → Finish).
    - `serial/`: FRX XML parser and object factory registry.
    - `data/`: Data sources (JSON, XML, CSV, SQL), Dictionary, and Totals.
    - `preview/`: Rendered snapshots (`PreparedPages`) consumed by exporters.
    - `export/`: Multi-format exporters (HTML, PDF, PNG, CSV, XLSX, etc.).

## 🛠 Building and Running
- **Build All**: `go build ./...`
- **Run Tests**: `go test ./...` (maintains ~98% coverage).
- **Test Package**: `go test ./engine/...`
- **Run Examples**: `go run ./examples/frx_to_html/` (requires `test-reports/` directory).
- **Benchmarks**: `go test -bench=. ./engine/...`

## ⚖️ Development Conventions
- **Source of Truth**: The original C# source code in `original-dotnet/` is the single source of truth. **Read the C# code first** and port algorithms exactly.
- **Idiomatic Translation**:
    - C# Properties → Go Getters/Setters (e.g., `Visible()` / `SetVisible()`).
    - C# Inheritance → Go Struct Embedding.
    - C# Events → Go Callback Fields.
    - C# Partial Classes → Multiple Go files in the same package (e.g., `engine/`).
- **Issue Tracking**: Use the **`bd` (beads)** tool for all tasks. Avoid markdown TODOs.
    - `bd ready`: Find available work.
    - `bd show <id>`: View issue details.
    - `bd update <id> --claim`: Claim a task.
    - `bd close <id>`: Complete a task.
- **Verification**: Use `tools/compare_html_semantic/` to verify that Go HTML output matches C# ground truth.

## 🏗 Key Patterns
- **Serial Registry**: Types must register a factory function in `serial.DefaultRegistry` to support FRX deserialization.
- **Expression Evaluation**: Bracket expressions `[DataSource.Field]` are evaluated via `Report.Calc()`. Dots in identifiers are sanitized to underscores for the `expr` library.
- **Engine Execution**: The engine runs in a 3-phase pipeline. It supports a `DoublePass` mode to resolve `[TotalPages]` by running the report twice.
- **Prepared Pages**: The engine output is a `preview.PreparedPages` collection, which is an immutable snapshot of the rendered report.

## 📂 Directory Structure Highlights
- `test-reports/`: Contains 100+ `.frx` sample files and data for smoke testing.
- `original-dotnet/`: The reference C# implementation.
- `porting-plan.md`: Detailed roadmap and architectural analysis.
- `CLAUDE.md`: Specific guidance for AI agents (Claude Code, etc.).
- `AGENTS.md`: Detailed instructions for the `bd` workflow.
