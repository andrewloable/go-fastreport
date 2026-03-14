package reportpkg

import (
	"github.com/andrewloable/go-fastreport/data"
	"github.com/andrewloable/go-fastreport/preview"
	"github.com/andrewloable/go-fastreport/report"
	"github.com/andrewloable/go-fastreport/style"
)

// ReportInfo holds descriptive metadata about a report.
type ReportInfo struct {
	// Name is the display name of the report.
	Name string
	// Author is the report author.
	Author string
	// Description provides a short description.
	Description string
	// Version is the report version string.
	Version string
}

// Report is the top-level container for a report definition.
// It holds pages, a data dictionary, styles, and run-time settings.
// It is the Go equivalent of FastReport.Report.
type Report struct {
	report.BaseObject

	// Pages is the ordered list of report page templates.
	pages []*ReportPage

	// Dictionary is the central registry for data sources, parameters, and totals.
	dictionary *data.Dictionary

	// Styles is the named-style registry for the report.
	styles *style.StyleSheet

	// Info holds descriptive metadata.
	Info ReportInfo

	// Script settings.
	ScriptText string

	// Serialization options.
	Compressed bool

	// Behaviour flags.
	ConvertNulls bool
	DoublePass   bool

	// Page numbering.
	InitialPageNumber int // default 1
	MaxPages          int // 0 = unlimited

	// Script event names.
	StartReportEvent  string
	FinishReportEvent string

	// BaseReportPath is the path to the base (parent) report file.
	// When non-empty, the base report is loaded and merged into this report
	// before the engine runs.
	BaseReportPath string

	// calcDS is the current-row data source set by SetCalcContext.
	calcDS data.DataSource

	// preparedPages holds the output from the last Prepare() call.
	preparedPages *preview.PreparedPages
}

// NewReport creates a Report with defaults.
func NewReport() *Report {
	return &Report{
		BaseObject:        *report.NewBaseObject(),
		dictionary:        data.NewDictionary(),
		styles:            style.NewStyleSheet(),
		InitialPageNumber: 1,
	}
}

// Dictionary returns the report's data dictionary.
func (r *Report) Dictionary() *data.Dictionary { return r.dictionary }

// SetDictionary replaces the report's data dictionary (useful for tests).
func (r *Report) SetDictionary(d *data.Dictionary) { r.dictionary = d }

// Styles returns the report's style sheet.
func (r *Report) Styles() *style.StyleSheet { return r.styles }

// SetStyles replaces the report's style sheet.
func (r *Report) SetStyles(ss *style.StyleSheet) { r.styles = ss }

// --- Pages ---

// Pages returns the ordered list of ReportPage templates.
func (r *Report) Pages() []*ReportPage { return r.pages }

// PageCount returns the number of pages.
func (r *Report) PageCount() int { return len(r.pages) }

// Page returns the page at index i.
func (r *Report) Page(i int) *ReportPage { return r.pages[i] }

// AddPage appends a page to the report.
func (r *Report) AddPage(p *ReportPage) { r.pages = append(r.pages, p) }

// FindPage returns the ReportPage with the given name, or nil if not found.
func (r *Report) FindPage(name string) *ReportPage {
	for _, p := range r.pages {
		if p.Name() == name {
			return p
		}
	}
	return nil
}

// RemovePage removes a page by reference.
func (r *Report) RemovePage(p *ReportPage) {
	for i, pg := range r.pages {
		if pg == p {
			r.pages = append(r.pages[:i], r.pages[i+1:]...)
			return
		}
	}
}

// --- Serialization ---

// Serialize writes Report properties that differ from defaults.
func (r *Report) Serialize(w report.Writer) error {
	if err := r.BaseObject.Serialize(w); err != nil {
		return err
	}
	if r.Info.Name != "" {
		w.WriteStr("ReportName", r.Info.Name)
	}
	if r.Info.Author != "" {
		w.WriteStr("ReportAuthor", r.Info.Author)
	}
	if r.Info.Description != "" {
		w.WriteStr("ReportDescription", r.Info.Description)
	}
	if r.Compressed {
		w.WriteBool("Compressed", true)
	}
	if r.ConvertNulls {
		w.WriteBool("ConvertNulls", true)
	}
	if r.DoublePass {
		w.WriteBool("DoublePass", true)
	}
	if r.InitialPageNumber != 1 {
		w.WriteInt("InitialPageNumber", r.InitialPageNumber)
	}
	if r.MaxPages != 0 {
		w.WriteInt("MaxPages", r.MaxPages)
	}
	if r.StartReportEvent != "" {
		w.WriteStr("StartReportEvent", r.StartReportEvent)
	}
	if r.FinishReportEvent != "" {
		w.WriteStr("FinishReportEvent", r.FinishReportEvent)
	}
	// Write pages as child elements — mirrors FastReport's Base.Serialize()
	// iterating ChildObjects and calling writer.Write(child) for each.
	for _, pg := range r.pages {
		if err := w.WriteObject(pg); err != nil {
			return err
		}
	}
	return nil
}

// Deserialize reads Report properties.
func (r *Report) Deserialize(rd report.Reader) error {
	if err := r.BaseObject.Deserialize(rd); err != nil {
		return err
	}
	r.Info.Name = rd.ReadStr("ReportName", "")
	r.Info.Author = rd.ReadStr("ReportAuthor", "")
	r.Info.Description = rd.ReadStr("ReportDescription", "")
	r.Compressed = rd.ReadBool("Compressed", false)
	r.ConvertNulls = rd.ReadBool("ConvertNulls", false)
	r.DoublePass = rd.ReadBool("DoublePass", false)
	r.InitialPageNumber = rd.ReadInt("InitialPageNumber", 1)
	r.MaxPages = rd.ReadInt("MaxPages", 0)
	r.StartReportEvent = rd.ReadStr("StartReportEvent", "")
	r.FinishReportEvent = rd.ReadStr("FinishReportEvent", "")
	return nil
}
