package reportpkg

import (
	"github.com/andrewloable/go-fastreport/report"
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
}

// NewReport creates a Report with defaults.
func NewReport() *Report {
	return &Report{
		BaseObject:        *report.NewBaseObject(),
		InitialPageNumber: 1,
	}
}

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
