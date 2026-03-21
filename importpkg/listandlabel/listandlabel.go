// Package listandlabel provides a stub importer for List & Label report files.
//
// The full C# implementation lives in:
//
//	original-dotnet/FastReport.Base/Import/ListAndLabel/ListAndLabelImport.cs
//	original-dotnet/FastReport.Base/Import/ListAndLabel/UnitsConverter.cs
//
// That implementation (FastReport.Import.ListAndLabel.ListAndLabelImport)
// parses List & Label report definitions and maps their component tree to
// FastReport bands and objects using ComponentsFactory helpers. The Go port
// is not yet implemented; this stub provides the correct interface so that
// callers can wire up the importer and receive a clear "not implemented" error
// rather than a panic.
package listandlabel

import (
	"fmt"
	"io"

	"github.com/andrewloable/go-fastreport/importpkg"
	"github.com/andrewloable/go-fastreport/reportpkg"
)

// Converter imports List & Label report definitions into go-fastreport reports.
//
// It embeds importpkg.ImportBase which provides Name/Report accessors and the
// base LoadReportFromFile / LoadReportFromStream stubs that match the C#
// ImportBase contract.
//
// C# ref: FastReport.Import.ListAndLabel.ListAndLabelImport
//
//	(ListAndLabelImport.cs)
type Converter struct {
	importpkg.ImportBase
}

// New returns a new Converter with its plugin name set.
func New() *Converter {
	c := &Converter{}
	c.SetName("List & Label Importer")
	return c
}

// Import reads a List & Label report from r and returns a populated Report.
//
// This method is not yet implemented. It returns an error so that callers can
// detect the unimplemented state without a panic.
//
// C# ref: ListAndLabelImport.LoadReport(Report, Stream) —
// ListAndLabelImport.cs
func (c *Converter) Import(r io.Reader) (*reportpkg.Report, error) {
	return nil, fmt.Errorf("listandlabel: Import not implemented")
}

// LoadReportFromFile loads a List & Label report from the named file into rpt.
//
// Not yet implemented.
//
// C# ref: ListAndLabelImport.LoadReport(Report, string) —
// ListAndLabelImport.cs
func (c *Converter) LoadReportFromFile(rpt *reportpkg.Report, filename string) error {
	return fmt.Errorf("listandlabel: LoadReportFromFile not implemented")
}

// LoadReportFromStream loads a List & Label report from r into rpt.
//
// Not yet implemented.
//
// C# ref: ListAndLabelImport.LoadReport(Report, Stream) —
// ListAndLabelImport.cs
func (c *Converter) LoadReportFromStream(rpt *reportpkg.Report, r io.Reader) error {
	return fmt.Errorf("listandlabel: LoadReportFromStream not implemented")
}
