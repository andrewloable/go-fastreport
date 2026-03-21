// Package stimulsoft provides a stub importer for Stimulsoft report files
// (.mrt, .mrz, .mrx).
//
// The full C# implementation lives in:
//
//	original-dotnet/FastReport.Base/Import/StimulSoft/StimulSoftImport.cs
//	original-dotnet/FastReport.Base/Import/StimulSoft/UnitsConverter.cs
//
// That implementation (FastReport.Import.StimulSoft.StimulSoftImport) parses
// Stimulsoft XML/JSON report definitions and maps their component tree to
// FastReport bands and objects using ComponentsFactory helpers. The Go port
// is not yet implemented; this stub provides the correct interface so that
// callers can wire up the importer and receive a clear "not implemented" error
// rather than a panic.
package stimulsoft

import (
	"fmt"
	"io"

	"github.com/andrewloable/go-fastreport/importpkg"
	"github.com/andrewloable/go-fastreport/reportpkg"
)

// Converter imports Stimulsoft report definitions into go-fastreport reports.
//
// It embeds importpkg.ImportBase which provides Name/Report accessors and the
// base LoadReportFromFile / LoadReportFromStream stubs that match the C#
// ImportBase contract.
//
// C# ref: FastReport.Import.StimulSoft.StimulSoftImport (StimulSoftImport.cs)
type Converter struct {
	importpkg.ImportBase
}

// New returns a new Converter with its plugin name set.
func New() *Converter {
	c := &Converter{}
	c.SetName("StimulSoft Importer")
	return c
}

// Import reads a Stimulsoft report from r and returns a populated Report.
//
// This method is not yet implemented. It returns an error so that callers can
// detect the unimplemented state without a panic.
//
// C# ref: StimulSoftImport.LoadReport(Report, Stream) — StimulSoftImport.cs
func (c *Converter) Import(r io.Reader) (*reportpkg.Report, error) {
	return nil, fmt.Errorf("stimulsoft: Import not implemented")
}

// LoadReportFromFile loads a Stimulsoft report from the named file into rpt.
//
// Not yet implemented.
//
// C# ref: StimulSoftImport.LoadReport(Report, string) — StimulSoftImport.cs
func (c *Converter) LoadReportFromFile(rpt *reportpkg.Report, filename string) error {
	return fmt.Errorf("stimulsoft: LoadReportFromFile not implemented")
}

// LoadReportFromStream loads a Stimulsoft report from r into rpt.
//
// Not yet implemented.
//
// C# ref: StimulSoftImport.LoadReport(Report, Stream) — StimulSoftImport.cs
func (c *Converter) LoadReportFromStream(rpt *reportpkg.Report, r io.Reader) error {
	return fmt.Errorf("stimulsoft: LoadReportFromStream not implemented")
}
