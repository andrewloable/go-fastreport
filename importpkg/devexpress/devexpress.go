// Package devexpress provides a stub importer for DevExpress report files
// (.repx).
//
// The full C# implementation lives in:
//
//	original-dotnet/FastReport.Base/Import/DevExpress/DevExpressImport.cs
//	original-dotnet/FastReport.Base/Import/DevExpress/DevExpressImport.XmlSource.cs
//	original-dotnet/FastReport.Base/Import/DevExpress/UnitsConverter.cs
//
// That implementation (FastReport.Import.DevExpress.DevExpressImport) parses
// DevExpress XtraReports XML definitions and maps their band/control hierarchy
// to FastReport bands and objects using ComponentsFactory helpers. The Go port
// is not yet implemented; this stub provides the correct interface so that
// callers can wire up the importer and receive a clear "not implemented" error
// rather than a panic.
package devexpress

import (
	"fmt"
	"io"

	"github.com/andrewloable/go-fastreport/importpkg"
	"github.com/andrewloable/go-fastreport/reportpkg"
)

// Converter imports DevExpress XtraReports definitions into go-fastreport
// reports.
//
// It embeds importpkg.ImportBase which provides Name/Report accessors and the
// base LoadReportFromFile / LoadReportFromStream stubs that match the C#
// ImportBase contract.
//
// C# ref: FastReport.Import.DevExpress.DevExpressImport (DevExpressImport.cs)
type Converter struct {
	importpkg.ImportBase
}

// New returns a new Converter with its plugin name set.
func New() *Converter {
	c := &Converter{}
	c.SetName("DevExpress Importer")
	return c
}

// Import reads a DevExpress report from r and returns a populated Report.
//
// This method is not yet implemented. It returns an error so that callers can
// detect the unimplemented state without a panic.
//
// C# ref: DevExpressImport.LoadReport(Report, Stream) — DevExpressImport.cs
func (c *Converter) Import(r io.Reader) (*reportpkg.Report, error) {
	return nil, fmt.Errorf("devexpress: Import not implemented")
}

// LoadReportFromFile loads a DevExpress report from the named file into rpt.
//
// Not yet implemented.
//
// C# ref: DevExpressImport.LoadReport(Report, string) — DevExpressImport.cs
func (c *Converter) LoadReportFromFile(rpt *reportpkg.Report, filename string) error {
	return fmt.Errorf("devexpress: LoadReportFromFile not implemented")
}

// LoadReportFromStream loads a DevExpress report from r into rpt.
//
// Not yet implemented.
//
// C# ref: DevExpressImport.LoadReport(Report, Stream) — DevExpressImport.cs
func (c *Converter) LoadReportFromStream(rpt *reportpkg.Report, r io.Reader) error {
	return fmt.Errorf("devexpress: LoadReportFromStream not implemented")
}
