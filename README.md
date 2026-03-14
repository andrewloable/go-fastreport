# go-fastreport

**go-fastreport** is an open source reporting engine written in Go that allows applications to generate document-like reports using existing FastReport `.frx` templates.

The project provides a lightweight runtime capable of loading and rendering FastReport report definitions using data from Go applications. It enables backend services, APIs, and microservices to generate reports without depending on the .NET runtime.

---

## Features

* Load and parse FastReport `.frx` templates
* Generate document-style reports
* Native Go implementation
* High performance and lightweight
* Band-based report layout engine
* Data binding from Go structures and maps
* Export reports to PDF
* Designed for backend services and cloud environments

---

## Motivation

FastReport is a powerful reporting engine widely used in .NET applications. However, Go applications currently lack a native engine capable of rendering FastReport templates.

**go-fastreport** aims to bridge that gap by providing:

* Compatibility with existing FastReport `.frx` report files
* A native Go reporting engine
* High performance suitable for backend systems
* Easy integration with Go applications

This allows teams to reuse existing report templates without rewriting them.

---

## Example

```go
package main

import (
    "github.com/go-fastreport/go-fastreport"
)

func main() {

    report := fastreport.Load("invoice.frx")

    report.SetData(map[string]interface{}{
        "customer": "John Doe",
        "total": 1200,
    })

    report.ExportPDF("invoice.pdf")
}
```

---

## Supported Report Elements

The engine aims to support common FastReport components including:

* ReportPage
* PageHeaderBand
* PageFooterBand
* DataBand
* GroupHeaderBand
* TextObject
* TableObject
* PictureObject
* SubReport

Support will expand over time.

---

## Architecture

The project is composed of several modules:

```
parser/
FRX template parser

engine/
report rendering engine

layout/
band layout engine

components/
text, table, image components

datasource/
data binding system

export/
PDF and HTML exporters
```

---

## Roadmap

### Phase 1

* FRX XML parser
* Basic report rendering
* Text objects
* Data binding
* PDF export

### Phase 2

* Tables
* Images
* Page layout improvements
* Group bands

### Phase 3

* Charts
* Barcodes
* Subreports
* HTML export

---

## Installation

```bash
go get github.com/go-fastreport/go-fastreport
```

---

## Project Status

Early development.

The goal of the project is to build a fully functional FastReport-compatible reporting engine for Go.

---

## License

MIT License

---

## Disclaimer

This project is inspired by FastReport but is an independent implementation written in Go.

It is not affiliated with or endorsed by the FastReport project.

```
```
