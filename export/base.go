package export

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/andrewloable/go-fastreport/preview"
	"github.com/andrewloable/go-fastreport/report"
)

// PageRange controls which pages are exported.
type PageRange int

const (
	// PageRangeAll exports all pages (default).
	PageRangeAll PageRange = iota
	// PageRangeCurrent exports only the current page (set via CurPage).
	PageRangeCurrent
	// PageRangeCustom exports the pages specified by PageNumbers.
	PageRangeCustom
)

// ExportBase is the base for all export filters.
// Concrete exporters embed ExportBase and override the hook methods.
//
// Usage:
//
//	type MyExport struct { export.ExportBase }
//	func (m *MyExport) ExportBand(band *preview.PreparedBand) { ... }
//	func (m *MyExport) Export(pages *preview.PreparedPages, w io.Writer) error {
//	    return m.ExportBase.Export(pages, w, m)
//	}
type ExportBase struct {
	// PageRange controls which pages are included.
	PageRange PageRange

	// PageNumbers is a comma/range string like "1,3-5,12".
	// Used when PageRange == PageRangeCustom. Empty means all pages.
	PageNumbers string

	// CurPage is the 1-based page number to export when PageRange == PageRangeCurrent.
	CurPage int

	// Zoom is a scaling factor applied to the exported content (default 1.0).
	// Matches C# ExportBase.Zoom.
	Zoom float32

	// OnProgress is an optional callback invoked once per page during export.
	// It receives the current 1-based page number and the total page count.
	// This is the Go equivalent of C# ReportSettings.OnProgress (called from
	// ExportBase.Export for each page when ShowProgress is true).
	// C# ref: FastReport.Base/Export/ExportBase.cs, Export() method.
	OnProgress func(page, total int)

	// HasMultipleFiles indicates that this exporter produces multiple output files
	// (e.g. one per page). Mirrors C# ExportBase.HasMultipleFiles (line 149).
	HasMultipleFiles bool

	// ShiftNonExportable indicates that non-exportable bands should shift
	// subsequent bands up. Mirrors C# ExportBase.ShiftNonExportable (line 159).
	ShiftNonExportable bool

	// pages holds the resolved zero-based page indices to export.
	pages []int

	// generatedFiles holds the paths of output files produced by this export.
	generatedFiles []string

	// tempFiles tracks temporary files created via CreateTempFile.
	tempFiles []*os.File
}

// NewExportBase creates an ExportBase with sensible defaults.
func NewExportBase() ExportBase {
	return ExportBase{
		PageRange: PageRangeAll,
		CurPage:   1,
		Zoom:      1,
	}
}

// GeneratedFiles returns the list of output file paths produced by this export.
// Matches C# ExportBase.GeneratedFiles.
func (e *ExportBase) GeneratedFiles() []string { return e.generatedFiles }

// Serialize writes non-default ExportBase settings to w.
// Mirrors C# ExportBase.Serialize (ExportBase.cs line 378).
func (e *ExportBase) Serialize(w report.Writer) {
	if e.PageRange != PageRangeAll {
		w.WriteInt("PageRange", int(e.PageRange))
	}
	if e.PageNumbers != "" {
		w.WriteStr("PageNumbers", e.PageNumbers)
	}
	if e.ShiftNonExportable {
		w.WriteBool("ShiftNonExportable", true)
	}
	if e.HasMultipleFiles {
		w.WriteBool("HasMultipleFiles", true)
	}
}

// Deserialize reads ExportBase settings from r.
// Mirrors C# ExportBase.Deserialize (ExportBase.cs).
func (e *ExportBase) Deserialize(r report.Reader) {
	e.PageRange = PageRange(r.ReadInt("PageRange", int(PageRangeAll)))
	e.PageNumbers = r.ReadStr("PageNumbers", "")
	e.ShiftNonExportable = r.ReadBool("ShiftNonExportable", false)
	e.HasMultipleFiles = r.ReadBool("HasMultipleFiles", false)
}

// AddGeneratedFile appends path to the generated files list.
func (e *ExportBase) AddGeneratedFile(path string) { e.generatedFiles = append(e.generatedFiles, path) }

// CreateTempFile creates a new temporary file in the OS temp directory under a
// "TempExport" subdirectory.  The file is tracked and deleted by DeleteTempFiles.
// Matches C# ExportBase.CreateTempFile.
func (e *ExportBase) CreateTempFile() (*os.File, error) {
	dir := filepath.Join(os.TempDir(), "TempExport")
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return nil, fmt.Errorf("export: create temp dir: %w", err)
	}
	f, err := os.CreateTemp(dir, "")
	if err != nil {
		return nil, fmt.Errorf("export: create temp file: %w", err)
	}
	e.tempFiles = append(e.tempFiles, f)
	return f, nil
}

// DeleteTempFiles closes and removes all temporary files created via CreateTempFile.
// Matches C# ExportBase.DeleteTempFiles.
func (e *ExportBase) DeleteTempFiles() {
	for _, f := range e.tempFiles {
		name := f.Name()
		_ = f.Close()
		_ = os.Remove(name)
	}
	e.tempFiles = e.tempFiles[:0]
}

// ── Page-number parsing ───────────────────────────────────────────────────────

// ParsePageNumbers parses a page range string like "1,3-5,12" and returns the
// corresponding zero-based page indices (sorted, de-duplicated).
// totalPages is the total number of pages (used to resolve trailing "-" ranges).
// An empty string returns (nil, nil) — the caller should treat this as "all pages".
func ParsePageNumbers(s string, totalPages int) ([]int, error) {
	s = strings.ReplaceAll(s, " ", "")
	if s == "" {
		return nil, nil
	}

	// Allow trailing dash: "3-" → "3-N".
	if strings.HasSuffix(s, "-") {
		s += strconv.Itoa(totalPages)
	}

	seen := make(map[int]bool)
	var result []int

	for _, part := range strings.Split(s, ",") {
		if part == "" {
			continue
		}
		if idx := strings.Index(part, "-"); idx >= 0 {
			// Range.
			startStr := part[:idx]
			endStr := part[idx+1:]
			start, err := strconv.Atoi(startStr)
			if err != nil {
				return nil, fmt.Errorf("invalid page number %q in %q", startStr, s)
			}
			end, err := strconv.Atoi(endStr)
			if err != nil {
				return nil, fmt.Errorf("invalid page number %q in %q", endStr, s)
			}
			if start > end {
				start, end = end, start
			}
			for p := start; p <= end; p++ {
				idx0 := p - 1
				if !seen[idx0] {
					seen[idx0] = true
					result = append(result, idx0)
				}
			}
		} else {
			// Single page.
			p, err := strconv.Atoi(part)
			if err != nil {
				return nil, fmt.Errorf("invalid page number %q in %q", part, s)
			}
			idx0 := p - 1
			if !seen[idx0] {
				seen[idx0] = true
				result = append(result, idx0)
			}
		}
	}
	return result, nil
}

// preparePageIndices resolves the set of 0-based page indices to export.
func (e *ExportBase) preparePageIndices(totalPages int) error {
	e.pages = e.pages[:0]

	switch e.PageRange {
	case PageRangeCurrent:
		e.pages = append(e.pages, e.CurPage-1)
	case PageRangeCustom:
		indices, err := ParsePageNumbers(e.PageNumbers, totalPages)
		if err != nil {
			return err
		}
		if indices == nil {
			// Empty PageNumbers → all pages.
			for i := 0; i < totalPages; i++ {
				e.pages = append(e.pages, i)
			}
		} else {
			e.pages = indices
		}
	default: // PageRangeAll
		for i := 0; i < totalPages; i++ {
			e.pages = append(e.pages, i)
		}
	}

	// Remove out-of-range indices.
	valid := e.pages[:0]
	for _, idx := range e.pages {
		if idx >= 0 && idx < totalPages {
			valid = append(valid, idx)
		}
	}
	e.pages = valid
	return nil
}

// ── Exporter interface ────────────────────────────────────────────────────────

// Exporter is the interface that concrete export filters implement.
// ExportBase.Export calls these hooks in order for each page/band.
type Exporter interface {
	// Start is called once before any pages are exported.
	Start() error
	// ExportPageBegin is called at the start of each page.
	ExportPageBegin(page *preview.PreparedPage) error
	// ExportBand is called for each band on the page.
	ExportBand(band *preview.PreparedBand) error
	// ExportPageEnd is called at the end of each page.
	ExportPageEnd(page *preview.PreparedPage) error
	// Finish is called once after all pages are exported.
	Finish() error
}

// ── Export ────────────────────────────────────────────────────────────────────

// Export drives the export lifecycle.
// pages is the PreparedPages produced by the report engine.
// w is the output writer (available to the concrete exporter via its own field).
// exp is the concrete exporter — typically the struct that embeds ExportBase.
func (e *ExportBase) Export(pages *preview.PreparedPages, w io.Writer, exp Exporter) error {
	if pages == nil {
		return fmt.Errorf("export: prepared pages is nil")
	}
	_ = w // available to concrete exporter; stored externally

	if err := e.preparePageIndices(pages.Count()); err != nil {
		return fmt.Errorf("export: prepare page indices: %w", err)
	}

	e.generatedFiles = e.generatedFiles[:0]

	if len(e.pages) == 0 {
		return nil
	}

	if err := exp.Start(); err != nil {
		return fmt.Errorf("export: start: %w", err)
	}

	for i, idx := range e.pages {
		pg := pages.GetPage(idx)
		if pg == nil {
			continue
		}
		// Notify progress before each page — mirrors C# ExportBase.Export which
		// calls Config.ReportSettings.OnProgress(Report, message, i+1, pages.Count).
		if e.OnProgress != nil {
			e.OnProgress(i+1, len(e.pages))
		}
		if err := exp.ExportPageBegin(pg); err != nil {
			return fmt.Errorf("export: page %d begin: %w", idx+1, err)
		}
		for _, b := range pg.Bands {
			if err := exp.ExportBand(b); err != nil {
				return fmt.Errorf("export: page %d band %q: %w", idx+1, b.Name, err)
			}
		}
		if err := exp.ExportPageEnd(pg); err != nil {
			return fmt.Errorf("export: page %d end: %w", idx+1, err)
		}
	}

	if err := exp.Finish(); err != nil {
		return fmt.Errorf("export: finish: %w", err)
	}
	e.DeleteTempFiles()
	return nil
}

// Pages returns the resolved zero-based page indices (populated after Export).
func (e *ExportBase) Pages() []int { return e.pages }

// ── NoopExporter ──────────────────────────────────────────────────────────────

// NoopExporter is an Exporter with empty hook implementations.
// Embed it in concrete exporters to avoid implementing unused hooks.
type NoopExporter struct{}

func (NoopExporter) Start() error                                  { return nil }
func (NoopExporter) ExportPageBegin(*preview.PreparedPage) error   { return nil }
func (NoopExporter) ExportBand(*preview.PreparedBand) error        { return nil }
func (NoopExporter) ExportPageEnd(*preview.PreparedPage) error     { return nil }
func (NoopExporter) Finish() error                                 { return nil }
