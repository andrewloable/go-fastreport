// Package importpkg provides the base infrastructure for importing reports from
// external formats (RDL, StimulSoft, etc.) into go-fastreport.
//
// It is the Go equivalent of FastReport.Import (ImportBase.cs and
// ComponentsFactory.cs in original-dotnet/FastReport.Base/Import/).
package importpkg

import (
	"io"

	"github.com/andrewloable/go-fastreport/reportpkg"
)

// ImportBase is the base struct for all report import plugins.
// Concrete importers (e.g. RDL, StimulSoft) embed this struct and override
// LoadReportFromFile / LoadReportFromStream as needed.
//
// It is the Go equivalent of FastReport.Import.ImportBase.
// C# ref: original-dotnet/FastReport.Base/Import/ImportBase.cs
type ImportBase struct {
	// name is the human-readable plugin name (e.g. "RDL Importer").
	name string

	// report holds a reference to the target report populated during import.
	// Set by LoadReportFromFile / LoadReportFromStream before delegating to
	// the concrete importer.
	report *reportpkg.Report
}

// Name returns the plugin name.
func (b *ImportBase) Name() string { return b.name }

// SetName sets the plugin name. Concrete importers call this in their constructor.
func (b *ImportBase) SetName(name string) { b.name = name }

// Report returns the report instance currently being populated.
func (b *ImportBase) Report() *reportpkg.Report { return b.report }

// SetReport stores the target report. Called by LoadReportFromFile and
// LoadReportFromStream before delegating import work to the concrete importer.
func (b *ImportBase) SetReport(r *reportpkg.Report) { b.report = r }

// LoadReportFromFile stores the target report and makes it ready for import
// from the named file. Concrete importers override this method to parse the
// file format and populate the report using ComponentsFactory helpers.
//
// C# ref: FastReport.Import.ImportBase.LoadReport(Report, string) —
// the C# base implementation calls report.Clear() which resets all pages
// and the dictionary. In the Go port, callers should pass a freshly created
// report or handle resetting themselves before calling this method.
func (b *ImportBase) LoadReportFromFile(report *reportpkg.Report, filename string) error {
	b.report = report
	return nil
}

// LoadReportFromStream stores the target report and makes it ready for import
// from the provided reader. Concrete importers override this method to parse
// the stream and populate the report using ComponentsFactory helpers.
//
// C# ref: FastReport.Import.ImportBase.LoadReport(Report, Stream).
func (b *ImportBase) LoadReportFromStream(report *reportpkg.Report, r io.Reader) error {
	b.report = report
	return nil
}
