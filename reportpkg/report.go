package reportpkg

import (
	"github.com/andrewloable/go-fastreport/data"
	"github.com/andrewloable/go-fastreport/export"
	"github.com/andrewloable/go-fastreport/preview"
	"github.com/andrewloable/go-fastreport/report"
	"github.com/andrewloable/go-fastreport/style"
)

// ReportInfo holds descriptive metadata about a report.
// It is the Go equivalent of FastReport.ReportInfo embedded in ReportSettings.
type ReportInfo struct {
	// Name is the display name of the report.
	Name string
	// Author is the report author.
	Author string
	// Description provides a short description.
	Description string
	// Version is the report version string (free-form, e.g. "1.0").
	Version string
	// Created is the ISO-8601 creation timestamp stored in the FRX.
	Created string
	// Modified is the ISO-8601 last-modified timestamp stored in the FRX.
	Modified string
	// CreatorVersion is the FastReport version that created the file (e.g. "2023.1.0").
	CreatorVersion string
	// SavePreviewPicture indicates the FRX should embed a preview thumbnail.
	SavePreviewPicture bool
	// Picture holds the raw bytes of the embedded preview image (PNG/JPEG).
	Picture []byte
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

	// OnStartReport is called by the engine at the very beginning of a run,
	// after data sources are initialised. It is the Go equivalent of the
	// FastReport OnStartReport script event.
	OnStartReport func()

	// OnFinishReport is called by the engine at the very end of a run,
	// after all pages have been generated. It is the Go equivalent of the
	// FastReport OnFinishReport script event.
	OnFinishReport func()

	// BaseReportPath is the path to the base (parent) report file.
	// When non-empty, the base report is loaded and merged into this report
	// before the engine runs.
	BaseReportPath string

	// ExportsOptions holds report-level export UI configuration.
	ExportsOptions *export.ExportsOptions

	// calcDS is the current-row data source set by SetCalcContext.
	calcDS data.DataSource

	// preparedPages holds the output from the last Prepare() call.
	preparedPages *preview.PreparedPages

	// customFunctions is the registry of user-defined callback functions.
	// Keys are function names as they appear in report expressions.
	customFunctions map[string]func(args []any) (any, error)
}

// NewReport creates a Report with defaults.
func NewReport() *Report {
	return &Report{
		BaseObject:        *report.NewBaseObject(),
		dictionary:        data.NewDictionary(),
		styles:            style.NewStyleSheet(),
		InitialPageNumber: 1,
		ExportsOptions:    export.NewExportsOptions(),
	}
}

// RegisterFunction registers a named Go callback function that can be called
// from report expressions using the syntax [FuncName(arg1, arg2, ...)].
// The fn receives all arguments as []any and returns a single value or an error.
// Registering a name that already exists overwrites the previous entry.
//
// Example:
//
//	r.RegisterFunction("DoubleValue", func(args []any) (any, error) {
//	    v := args[0].(int)
//	    return v * 2, nil
//	})
//	// In the report template: [DoubleValue(5)] → "10"
func (r *Report) RegisterFunction(name string, fn func(args []any) (any, error)) {
	if r.customFunctions == nil {
		r.customFunctions = make(map[string]func(args []any) (any, error))
	}
	r.customFunctions[name] = fn
}

// CustomFunctions returns a copy of the registered custom function map.
// The returned map is safe to inspect but mutations do not affect the report.
func (r *Report) CustomFunctions() map[string]func(args []any) (any, error) {
	out := make(map[string]func(args []any) (any, error), len(r.customFunctions))
	for k, v := range r.customFunctions {
		out[k] = v
	}
	return out
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
	if r.Info.Version != "" {
		w.WriteStr("ReportVersion", r.Info.Version)
	}
	if r.Info.Created != "" {
		w.WriteStr("Created", r.Info.Created)
	}
	if r.Info.Modified != "" {
		w.WriteStr("Modified", r.Info.Modified)
	}
	if r.Info.CreatorVersion != "" {
		w.WriteStr("CreatorVersion", r.Info.CreatorVersion)
	}
	if r.Info.SavePreviewPicture {
		w.WriteBool("SavePreviewPicture", true)
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
	// Write Styles child element when the stylesheet has entries.
	if r.styles.Len() > 0 {
		if err := w.WriteObject(&stylesSerializer{r.styles}); err != nil {
			return err
		}
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
	r.Info.Version = rd.ReadStr("ReportVersion", "")
	r.Info.Created = rd.ReadStr("Created", "")
	r.Info.Modified = rd.ReadStr("Modified", "")
	// FRX .NET uses dot-qualified attribute names (e.g. "ReportInfo.Description").
	// Read these as fallbacks when the short forms are empty.
	if r.Info.Name == "" {
		r.Info.Name = rd.ReadStr("ReportInfo.Name", "")
	}
	if r.Info.Author == "" {
		r.Info.Author = rd.ReadStr("ReportInfo.Author", "")
	}
	if r.Info.Description == "" {
		r.Info.Description = rd.ReadStr("ReportInfo.Description", "")
	}
	if r.Info.Version == "" {
		r.Info.Version = rd.ReadStr("ReportInfo.Version", "")
	}
	if r.Info.Created == "" {
		r.Info.Created = rd.ReadStr("ReportInfo.Created", "")
	}
	if r.Info.Modified == "" {
		r.Info.Modified = rd.ReadStr("ReportInfo.Modified", "")
	}
	r.Info.CreatorVersion = rd.ReadStr("CreatorVersion", "")
	r.Info.SavePreviewPicture = rd.ReadBool("SavePreviewPicture", false)
	r.Compressed = rd.ReadBool("Compressed", false)
	r.ConvertNulls = rd.ReadBool("ConvertNulls", false)
	r.DoublePass = rd.ReadBool("DoublePass", false)
	r.InitialPageNumber = rd.ReadInt("InitialPageNumber", 1)
	r.MaxPages = rd.ReadInt("MaxPages", 0)
	r.StartReportEvent = rd.ReadStr("StartReportEvent", "")
	r.FinishReportEvent = rd.ReadStr("FinishReportEvent", "")
	return nil
}
