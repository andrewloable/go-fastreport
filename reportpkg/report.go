package reportpkg

import (
	"strings"

	"github.com/andrewloable/go-fastreport/data"
	"github.com/andrewloable/go-fastreport/export"
	"github.com/andrewloable/go-fastreport/preview"
	"github.com/andrewloable/go-fastreport/report"
	"github.com/andrewloable/go-fastreport/style"
)

// SaveMode specifies the save permissions for a designed report.
// C# ref: FastReport.Base/ReportInfo.cs enum SaveMode.
type SaveMode int

const (
	// SaveModeAll allows saving to all locations (default).
	SaveModeAll SaveMode = iota
	// SaveModeOriginal allows saving only to the original location.
	SaveModeOriginal
	// SaveModeUser allows saving for the current user.
	SaveModeUser
	// SaveModeRole allows saving for the current role/group.
	SaveModeRole
	// SaveModeSecurity allows saving with other security permissions.
	SaveModeSecurity
	// SaveModeDeny prohibits saving.
	SaveModeDeny
	// SaveModeCustom uses custom saving rules.
	SaveModeCustom
)

// String returns the C#-compatible enum name used in FRX serialization.
func (s SaveMode) String() string {
	switch s {
	case SaveModeAll:
		return "All"
	case SaveModeOriginal:
		return "Original"
	case SaveModeUser:
		return "User"
	case SaveModeRole:
		return "Role"
	case SaveModeSecurity:
		return "Security"
	case SaveModeDeny:
		return "Deny"
	case SaveModeCustom:
		return "Custom"
	default:
		return "All"
	}
}

// parseSaveMode converts a C# enum name string to a SaveMode value.
func parseSaveMode(s string) SaveMode {
	switch s {
	case "Original":
		return SaveModeOriginal
	case "User":
		return SaveModeUser
	case "Role":
		return SaveModeRole
	case "Security":
		return SaveModeSecurity
	case "Deny":
		return SaveModeDeny
	case "Custom":
		return SaveModeCustom
	default:
		return SaveModeAll
	}
}

// TextQuality specifies the quality of text rendering.
// C# ref: FastReport.Base/Report.cs enum TextQuality.
type TextQuality int

const (
	// TextQualityDefault uses system default text rendering.
	TextQualityDefault TextQuality = iota
	// TextQualityRegular uses regular quality.
	TextQualityRegular
	// TextQualityClearType uses ClearType rendering.
	TextQualityClearType
	// TextQualityAntiAlias uses anti-alias rendering (WYSIWYG).
	TextQualityAntiAlias
	// TextQualitySingleBPP uses single-bit-per-pixel rendering.
	TextQualitySingleBPP
	// TextQualitySingleBPPGridFit uses single-bit-per-pixel with grid fit.
	TextQualitySingleBPPGridFit
)

// String returns the C#-compatible enum name for FRX serialization.
func (t TextQuality) String() string {
	switch t {
	case TextQualityRegular:
		return "Regular"
	case TextQualityClearType:
		return "ClearType"
	case TextQualityAntiAlias:
		return "AntiAlias"
	case TextQualitySingleBPP:
		return "SingleBPP"
	case TextQualitySingleBPPGridFit:
		return "SingleBPPGridFit"
	default:
		return "Default"
	}
}

// parseTextQuality converts a C# enum name string to a TextQuality value.
func parseTextQuality(s string) TextQuality {
	switch s {
	case "Regular":
		return TextQualityRegular
	case "ClearType":
		return TextQualityClearType
	case "AntiAlias":
		return TextQualityAntiAlias
	case "SingleBPP":
		return TextQualitySingleBPP
	case "SingleBPPGridFit":
		return TextQualitySingleBPPGridFit
	default:
		return TextQualityDefault
	}
}

// ReportInfo holds descriptive metadata about a report.
// It is the Go equivalent of FastReport.ReportInfo embedded in ReportSettings.
// C# ref: FastReport.Base/ReportInfo.cs.
type ReportInfo struct {
	// Name is the display name of the report.
	Name string
	// Author is the report author.
	Author string
	// Description provides a short description.
	Description string
	// Version is the report version string (free-form, e.g. "1.0").
	Version string
	// Tag is an arbitrary string tag associated with the report.
	// C# ref: ReportInfo.Tag property.
	Tag string
	// Created is the ISO-8601 creation timestamp stored in the FRX.
	Created string
	// Modified is the ISO-8601 last-modified timestamp stored in the FRX.
	Modified string
	// CreatorVersion is the FastReport version that created the file (e.g. "2023.1.0").
	CreatorVersion string
	// SavePreviewPicture indicates the FRX should embed a preview thumbnail.
	SavePreviewPicture bool
	// PreviewPictureRatio is the scale ratio used when generating the preview picture.
	// Values <= 0 are clamped to 0.05. Default is 0.1.
	// C# ref: ReportInfo.PreviewPictureRatio property.
	PreviewPictureRatio float32
	// SaveMode controls who is allowed to save the report.
	// C# ref: ReportInfo.SaveMode property.
	SaveMode SaveMode
	// Picture holds the raw bytes of the embedded preview image (PNG/JPEG).
	Picture []byte
}

// Clear resets all ReportInfo fields to their default values.
// Mirrors C# ReportInfo.Clear() (ReportInfo.cs lines 201–217).
func (ri *ReportInfo) Clear() {
	ri.Name = ""
	ri.Author = ""
	ri.Version = ""
	ri.Description = ""
	ri.Tag = ""
	ri.Picture = nil
	ri.Created = ""
	ri.Modified = ""
	ri.SavePreviewPicture = false
	ri.PreviewPictureRatio = 0.1
	ri.CreatorVersion = ""
	ri.SaveMode = SaveModeAll
}

// Serialize writes ReportInfo fields as dot-qualified attributes.
// Mirrors C# ReportInfo.Serialize (ReportInfo.cs line 219).
func (ri *ReportInfo) Serialize(w report.Writer) {
	if ri.Name != "" {
		w.WriteStr("ReportInfo.Name", ri.Name)
	}
	if ri.Author != "" {
		w.WriteStr("ReportInfo.Author", ri.Author)
	}
	if ri.Description != "" {
		w.WriteStr("ReportInfo.Description", ri.Description)
	}
	if ri.Version != "" {
		w.WriteStr("ReportInfo.Version", ri.Version)
	}
	if ri.Tag != "" {
		w.WriteStr("ReportInfo.Tag", ri.Tag)
	}
	if ri.Created != "" {
		w.WriteStr("ReportInfo.Created", ri.Created)
	}
	if ri.Modified != "" {
		w.WriteStr("ReportInfo.Modified", ri.Modified)
	}
	if ri.CreatorVersion != "" {
		w.WriteStr("ReportInfo.CreatorVersion", ri.CreatorVersion)
	}
	if ri.SavePreviewPicture {
		w.WriteBool("ReportInfo.SavePreviewPicture", true)
	}
	if ri.PreviewPictureRatio != 0 && ri.PreviewPictureRatio != 0.1 {
		w.WriteFloat("ReportInfo.PreviewPictureRatio", ri.PreviewPictureRatio)
	}
	if ri.SaveMode != SaveModeAll {
		w.WriteStr("ReportInfo.SaveMode", ri.SaveMode.String())
	}
}

// Deserialize reads ReportInfo fields.  Reads both the C# dot-qualified
// "ReportInfo.*" form and legacy short-form names for backward compatibility.
// Mirrors C# ReportInfo.Deserialize (ReportInfo.cs).
func (ri *ReportInfo) Deserialize(r report.Reader) {
	ri.Name = r.ReadStr("ReportInfo.Name", "")
	if ri.Name == "" {
		ri.Name = r.ReadStr("ReportName", "")
	}
	ri.Author = r.ReadStr("ReportInfo.Author", "")
	if ri.Author == "" {
		ri.Author = r.ReadStr("ReportAuthor", "")
	}
	ri.Description = r.ReadStr("ReportInfo.Description", "")
	if ri.Description == "" {
		ri.Description = r.ReadStr("ReportDescription", "")
	}
	ri.Version = r.ReadStr("ReportInfo.Version", "")
	if ri.Version == "" {
		ri.Version = r.ReadStr("ReportVersion", "")
	}
	ri.Tag = r.ReadStr("ReportInfo.Tag", "")
	ri.Created = r.ReadStr("ReportInfo.Created", "")
	if ri.Created == "" {
		ri.Created = r.ReadStr("Created", "")
	}
	ri.Modified = r.ReadStr("ReportInfo.Modified", "")
	if ri.Modified == "" {
		ri.Modified = r.ReadStr("Modified", "")
	}
	ri.CreatorVersion = r.ReadStr("ReportInfo.CreatorVersion", "")
	if ri.CreatorVersion == "" {
		ri.CreatorVersion = r.ReadStr("CreatorVersion", "")
	}
	ri.SavePreviewPicture = r.ReadBool("ReportInfo.SavePreviewPicture", false)
	if !ri.SavePreviewPicture {
		ri.SavePreviewPicture = r.ReadBool("SavePreviewPicture", false)
	}
	ri.PreviewPictureRatio = r.ReadFloat("ReportInfo.PreviewPictureRatio", 0.1)
	if ri.PreviewPictureRatio <= 0 {
		ri.PreviewPictureRatio = 0.05
	}
	ri.SaveMode = parseSaveMode(r.ReadStr("ReportInfo.SaveMode", "All"))
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

	// ScriptLanguage records the script language stored in the FRX attribute.
	// The Go port does not execute C# or VB scripts; this field is preserved
	// for round-trip fidelity only.
	// C# ref: FastReport.Base/Report.cs ScriptLanguage property.
	ScriptLanguage string

	// Serialization options.
	Compressed bool

	// TextQuality specifies the text rendering quality hint stored in the FRX.
	// Affects on-screen preview only; PDF/image export uses its own rendering.
	// C# ref: FastReport.Base/Report.cs TextQuality property.
	TextQuality TextQuality

	// SmoothGraphics indicates whether graphic objects (bitmaps, shapes) should
	// be displayed smoothly. Stored in FRX for round-trip fidelity.
	// C# ref: FastReport.Base/Report.cs SmoothGraphics property.
	SmoothGraphics bool

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

	// OnCustomCalc is an optional hook called after Calc resolves a value from
	// the expression environment. It receives the original expression string and
	// the resolved value, and returns the (potentially overridden) value.
	// When nil, the resolved value is used as-is.
	//
	// This is the Go equivalent of the C# Report.CustomCalc event
	// (CustomCalcEventArgs.Expression / CalculatedObject).
	// C# ref: FastReport.Base/Report.cs Calc() method.
	OnCustomCalc func(expression string, value any) any

	// BaseReportPath is the path to the base (parent) report file.
	// When non-empty, the base report is loaded and merged into this report
	// before the engine runs.
	BaseReportPath string

	// OnLoadBaseReport is an optional callback invoked when the report needs to
	// load a base (parent) report from a file path. The callback receives the
	// resolved file path and the current report, and returns the loaded base
	// report or an error. When nil, the default file-system loader is used.
	//
	// This is the Go equivalent of the C# Report.LoadBaseReport event
	// (CustomLoadEventArgs.FileName / Report).
	// C# ref: FastReport.Base/Report.cs line ~1065.
	OnLoadBaseReport func(fileName string, r *Report) (*Report, error)

	// Settings holds global runtime settings for the report.
	Settings *ReportSettings

	// ExportsOptions holds report-level export UI configuration.
	ExportsOptions *export.ExportsOptions

	// calcDS is the current-row data source set by SetCalcContext.
	calcDS data.DataSource

	// preparedPages holds the output from the last Prepare() call.
	preparedPages *preview.PreparedPages

	// customFunctions is the registry of user-defined callback functions.
	// Keys are function names as they appear in report expressions.
	customFunctions map[string]func(args []any) (any, error)

	// aborted is set to true when Abort() is called; the engine checks this
	// to stop processing.
	// Mirrors C# Report.aborted field (Report.cs line 247).
	aborted bool

	// isParameterChanged tracks whether SetParameterValue has been called
	// since the last run, which affects IsPrepared.
	// Mirrors C# Report.isParameterChanged (Report.cs line 254).
	isParameterChanged bool
}

// NewReport creates a Report with defaults matching C# ClearReportProperties().
// C# ref: FastReport.Base/Report.cs ClearReportProperties() (~line 1115).
func NewReport() *Report {
	return &Report{
		BaseObject:        *report.NewBaseObject(),
		dictionary:        data.NewDictionary(),
		styles:            style.NewStyleSheet(),
		Settings:          NewReportSettings(),
		InitialPageNumber: 1,
		// ConvertNulls default is true — matches C# ClearReportProperties().
		ConvertNulls:   true,
		ExportsOptions: export.NewExportsOptions(),
		Info: ReportInfo{
			PreviewPictureRatio: 0.1,
		},
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
// Attribute names match C# FastReport serialization:
//   - ReportInfo.* dot-qualified form for metadata fields
//   - ScriptLanguage, TextQuality, SmoothGraphics for rendering hints
//
// C# ref: FastReport.Base/Report.cs Report.Serialize() (~line 1872)
// C# ref: FastReport.Base/ReportInfo.cs ReportInfo.Serialize() (~line 219)
func (r *Report) Serialize(w report.Writer) error {
	if err := r.BaseObject.Serialize(w); err != nil {
		return err
	}
	// ScriptLanguage is always written (C# always serializes it).
	// C# ref: Report.cs line 1887 — "always serialize ScriptLanguage"
	if r.ScriptLanguage != "" {
		w.WriteStr("ScriptLanguage", r.ScriptLanguage)
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
	if r.TextQuality != TextQualityDefault {
		w.WriteStr("TextQuality", r.TextQuality.String())
	}
	if r.SmoothGraphics {
		w.WriteBool("SmoothGraphics", true)
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
	// ReportInfo fields — delegate to ReportInfo.Serialize.
	// C# ref: ReportInfo.cs Serialize() method (~line 219).
	r.Info.Serialize(w)
	// Write Styles child element when the stylesheet has entries.
	if r.styles.Len() > 0 {
		if err := w.WriteObject(&stylesSerializer{r.styles}); err != nil {
			return err
		}
	}
	// Write Dictionary child element when the dictionary has any entries.
	// C# ref: FastReport.Report.Serialize — Dictionary.Serialize(Writer)
	if hasDictionaryContent(r.dictionary) {
		if err := w.WriteObject(&dictionarySerializer{r.dictionary}); err != nil {
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
// Reads both C# dot-qualified "ReportInfo.*" form and legacy short-form
// attributes for backwards compatibility.
// C# ref: FastReport.Base/Report.cs Report.Deserialize() (~line 1929)
func (r *Report) Deserialize(rd report.Reader) error {
	if err := r.BaseObject.Deserialize(rd); err != nil {
		return err
	}
	// ScriptLanguage — preserved for round-trip only; Go does not execute scripts.
	r.ScriptLanguage = rd.ReadStr("ScriptLanguage", "")
	// Core report flags.
	r.Compressed = rd.ReadBool("Compressed", false)
	// ConvertNulls default is true in C# (ClearReportProperties sets true).
	// C# ref: Report.cs line 1132.
	r.ConvertNulls = rd.ReadBool("ConvertNulls", true)
	r.DoublePass = rd.ReadBool("DoublePass", false)
	// TextQuality and SmoothGraphics — rendering hints, preserved for round-trip.
	r.TextQuality = parseTextQuality(rd.ReadStr("TextQuality", "Default"))
	r.SmoothGraphics = rd.ReadBool("SmoothGraphics", false)
	r.InitialPageNumber = rd.ReadInt("InitialPageNumber", 1)
	r.MaxPages = rd.ReadInt("MaxPages", 0)
	r.StartReportEvent = rd.ReadStr("StartReportEvent", "")
	r.FinishReportEvent = rd.ReadStr("FinishReportEvent", "")

	// ReportInfo fields — delegate to ReportInfo.Deserialize.
	// C# ref: ReportInfo.cs Serialize() writes "ReportInfo.*" attribute names.
	r.Info.Deserialize(rd)
	return nil
}

// ── ValidatableReport interface ───────────────────────────────────────────────
// Report satisfies the utils.ValidatableReport interface (import-cycle-free:
// only the method signatures need to match; no import of utils/ is needed).

// textHolder is a local interface for objects that expose a Text() method.
type textHolder interface {
	Text() string
}

// allObjects collects all report.Base objects from all pages recursively.
func (r *Report) allObjects() []report.Base {
	var result []report.Base
	var collect func(obj report.Base)
	collect = func(obj report.Base) {
		result = append(result, obj)
		if p, ok := obj.(report.Parent); ok {
			var children []report.Base
			p.GetChildObjects(&children)
			for _, child := range children {
				collect(child)
			}
		}
	}
	for _, pg := range r.pages {
		for _, b := range pg.AllBands() {
			collect(b)
		}
	}
	return result
}

// BandNames returns the names of all bands across all pages.
// Implements utils.ValidatableReport.
func (r *Report) BandNames() []string {
	var names []string
	for _, pg := range r.pages {
		for _, b := range pg.AllBands() {
			if n := b.Name(); n != "" {
				names = append(names, n)
			}
		}
	}
	return names
}

// DataSourceNames returns registered data source names.
// Implements utils.ValidatableReport.
func (r *Report) DataSourceNames() []string {
	if r.dictionary == nil {
		return nil
	}
	dss := r.dictionary.DataSources()
	names := make([]string, 0, len(dss))
	for _, ds := range dss {
		names = append(names, ds.Name())
	}
	return names
}

// TextExpressions returns all text values (including [bracket] expressions) from
// text-bearing report objects. Implements utils.ValidatableReport.
func (r *Report) TextExpressions() []string {
	var exprs []string
	for _, obj := range r.allObjects() {
		if th, ok := obj.(textHolder); ok {
			if t := th.Text(); t != "" {
				exprs = append(exprs, t)
			}
		}
	}
	return exprs
}

// ParameterNames returns the names of all report parameters.
// Implements utils.ValidatableReport.
func (r *Report) ParameterNames() []string {
	if r.dictionary == nil {
		return nil
	}
	params := r.dictionary.Parameters()
	names := make([]string, 0, len(params))
	for _, p := range params {
		names = append(names, p.Name)
	}
	return names
}

// ObjectNames returns the names of all named objects (components, bands, pages).
// Implements utils.ValidatableReport.
func (r *Report) ObjectNames() []string {
	var names []string
	for _, pg := range r.pages {
		if n := pg.Name(); n != "" {
			names = append(names, n)
		}
	}
	for _, obj := range r.allObjects() {
		if n := obj.Name(); n != "" {
			names = append(names, n)
		}
	}
	return names
}

// ── Report.cs additional methods ──────────────────────────────────────────────

// Abort signals the report to stop processing.
// Mirrors C# Report.Abort() (Report.cs line 1745).
func (r *Report) Abort() { r.aborted = true }

// Aborted returns true if Abort() has been called.
// Mirrors C# Report.Aborted property (Report.cs line 705).
func (r *Report) Aborted() bool { return r.aborted }

// SetAborted sets the aborted flag directly (used by the engine).
// Mirrors C# Report.SetAborted (Report.cs line 1794).
func (r *Report) SetAborted(v bool) { r.aborted = v }

// IsPrepared returns true when prepared pages are available and no parameter
// change has occurred since the last run.
// Mirrors C# Report.IsPrepared (Report.cs line 402–404):
//
//	get { return !isParameterChanged && PreparedPages != null && PreparedPages.Count != 0; }
func (r *Report) IsPrepared() bool {
	return !r.isParameterChanged && r.preparedPages != nil && r.preparedPages.Count() > 0
}

// FindObject returns the first named object (page, band, or component) whose
// Name equals name (case-insensitive), or nil if not found.
// Mirrors C# Report.FindObject (Report.cs line 1751–1758).
func (r *Report) FindObject(name string) report.Base {
	// Check pages first.
	for _, pg := range r.pages {
		if strings.EqualFold(pg.Name(), name) {
			return pg
		}
	}
	// Walk all objects in all pages.
	for _, obj := range r.allObjects() {
		if strings.EqualFold(obj.Name(), name) {
			return obj
		}
	}
	return nil
}

// ApplyStyles re-applies named styles to every report component that has a
// StyleName set. Call this after modifying the Styles collection.
// Mirrors C# Report.ApplyStyles (Report.cs line 1774–1781).
func (r *Report) ApplyStyles() {
	ss := r.styles
	if ss == nil {
		return
	}
	for _, obj := range r.allObjects() {
		if s, ok := obj.(style.Styleable); ok {
			ss.ApplyToObject(s)
		}
	}
}

// Clear resets all pages and report properties to defaults.
// Mirrors C# Report.Clear() (Report.cs line 1762–1766).
func (r *Report) Clear() {
	r.pages = nil
	r.Info.Clear()
	r.dictionary = data.NewDictionary()
	r.styles = style.NewStyleSheet()
	r.ScriptText = ""
	r.ScriptLanguage = ""
	r.Compressed = false
	r.ConvertNulls = true
	r.DoublePass = false
	r.TextQuality = TextQualityDefault
	r.SmoothGraphics = false
	r.InitialPageNumber = 1
	r.MaxPages = 0
	r.StartReportEvent = ""
	r.FinishReportEvent = ""
	r.BaseReportPath = ""
	r.aborted = false
	r.isParameterChanged = false
}

// GetParameterValue returns the value of the parameter identified by
// complexName (dot-separated path for nested parameters).
// Returns nil when the parameter is not found.
// Mirrors C# Report.GetParameterValue (Report.cs line 1607–1623).
func (r *Report) GetParameterValue(complexName string) any {
	par := data.GetParameter(r.dictionary, complexName)
	if par == nil {
		return nil
	}
	return par.Value
}

// SetParameterValue sets the value of the parameter identified by complexName.
// When the parameter does not exist it is created.
// Also clears any expression on the parameter and marks isParameterChanged.
// Mirrors C# Report.SetParameterValue (Report.cs line 1651–1663).
func (r *Report) SetParameterValue(complexName string, value any) {
	par := data.GetParameter(r.dictionary, complexName)
	if par == nil {
		par = data.CreateParameter(r.dictionary, complexName)
	}
	if par != nil {
		par.Value = value
		par.Expression = ""
	}
	r.isParameterChanged = true
}

// SetPreparedPages sets the prepared pages (called by the engine after a run).
// Mirrors C# Report.SetPreparedPages (Report.cs line 1787–1792).
func (r *Report) SetPreparedPages(pages *preview.PreparedPages) {
	r.preparedPages = pages
}
