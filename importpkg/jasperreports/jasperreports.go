// Package jasperreports provides a stub importer for JasperReports report
// files (.jrxml).
//
// The full C# implementation lives in:
//
//	original-dotnet/FastReport.Base/Import/JasperReports/JasperReportsImport.cs
//	original-dotnet/FastReport.Base/Import/JasperReports/UnitsConverter.cs
//
// That implementation (FastReport.Import.JasperReports.JasperReportsImport)
// parses JRXML XML definitions and maps their band/element hierarchy to
// FastReport bands and objects using ComponentsFactory helpers. The Go port
// is not yet implemented; this stub provides the correct interface so that
// callers can wire up the importer and receive a clear "not implemented" error
// rather than a panic.
package jasperreports

import (
	"fmt"
	"io"

	"github.com/andrewloable/go-fastreport/importpkg"
	"github.com/andrewloable/go-fastreport/reportpkg"
)

// Converter imports JasperReports JRXML definitions into go-fastreport reports.
//
// It embeds importpkg.ImportBase which provides Name/Report accessors and the
// base LoadReportFromFile / LoadReportFromStream stubs that match the C#
// ImportBase contract.
//
// C# ref: FastReport.Import.JasperReports.JasperReportsImport
//
//	(JasperReportsImport.cs)
type Converter struct {
	importpkg.ImportBase
}

// New returns a new Converter with its plugin name set.
func New() *Converter {
	c := &Converter{}
	c.SetName("JasperReports Importer")
	return c
}

// Import reads a JasperReports JRXML report from r and returns a populated
// Report.
//
// This method is not yet implemented. It returns an error so that callers can
// detect the unimplemented state without a panic.
//
// C# ref: JasperReportsImport.LoadReport(Report, Stream) —
// JasperReportsImport.cs
func (c *Converter) Import(r io.Reader) (*reportpkg.Report, error) {
	return nil, fmt.Errorf("jasperreports: Import not implemented")
}

// LoadReportFromFile loads a JasperReports JRXML report from the named file
// into rpt.
//
// Not yet implemented.
//
// C# ref: JasperReportsImport.LoadReport(Report, string) —
// JasperReportsImport.cs
func (c *Converter) LoadReportFromFile(rpt *reportpkg.Report, filename string) error {
	return fmt.Errorf("jasperreports: LoadReportFromFile not implemented")
}

// LoadReportFromStream loads a JasperReports JRXML report from r into rpt.
//
// Not yet implemented.
//
// C# ref: JasperReportsImport.LoadReport(Report, Stream) —
// JasperReportsImport.cs
func (c *Converter) LoadReportFromStream(rpt *reportpkg.Report, r io.Reader) error {
	return fmt.Errorf("jasperreports: LoadReportFromStream not implemented")
}
