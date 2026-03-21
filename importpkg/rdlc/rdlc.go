package rdlc

// Package rdlc provides a stub for importing Microsoft RDL/RDLC reports into go-fastreport.
//
// C# source: original-dotnet/FastReport.Base/Import/RDL/ (not present in open-source edition)
// The full RDL/RDLC import is a complex format converter that is not yet implemented.
// Use Import() to get a clear "not implemented" error rather than a panic.

import (
	"fmt"
	"io"

	"github.com/andrewloable/go-fastreport/importpkg"
	"github.com/andrewloable/go-fastreport/reportpkg"
)

// Converter imports Microsoft RDL/RDLC reports.
// Embed importpkg.ImportBase for base behaviour.
type Converter struct {
	importpkg.ImportBase
}

// New creates a new RDLC Converter.
func New() *Converter {
	c := &Converter{}
	c.SetName("RDL/RDLC Importer")
	c.SetReport(reportpkg.NewReport())
	return c
}

// Import reads an RDL/RDLC report from r and converts it to a go-fastreport Report.
// Not yet implemented — returns an error.
func (c *Converter) Import(r io.Reader) (*reportpkg.Report, error) {
	return nil, fmt.Errorf("rdlc: Import not implemented — RDL/RDLC format conversion is a future work item")
}

// LoadReportFromFile loads an RDL/RDLC report from a file path.
// Not yet implemented — returns an error.
func (c *Converter) LoadReportFromFile(path string) (*reportpkg.Report, error) {
	return nil, fmt.Errorf("rdlc: LoadReportFromFile not implemented — RDL/RDLC format conversion is a future work item")
}

// LoadReportFromStream loads an RDL/RDLC report from a reader.
// Not yet implemented — returns an error.
func (c *Converter) LoadReportFromStream(r io.Reader) (*reportpkg.Report, error) {
	return nil, fmt.Errorf("rdlc: LoadReportFromStream not implemented — RDL/RDLC format conversion is a future work item")
}
