package export_test

// exportfile_zip_test.go — tests for ExportToFile and ExportAndZip methods on ExportBase.
//
// C# reference: FastReport.Base/Export/ExportBase.cs
//   - Export(Report, string fileName)   lines 577-585
//   - ExportAndZip(Report, Stream)      lines 598-614

import (
	"archive/zip"
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/andrewloable/go-fastreport/export"
	"github.com/andrewloable/go-fastreport/preview"
)

// ── helpers ───────────────────────────────────────────────────────────────────

// buildOnePage builds a PreparedPages with one page and one band.
func buildOnePage() *preview.PreparedPages {
	pp := preview.New()
	pp.AddPage(595, 842, 1)
	_ = pp.AddBand(&preview.PreparedBand{Name: "B", Top: 0, Height: 30})
	return pp
}

// contentRecorder is a minimal Exporter that records the io.Writer it was given
// and writes a fixed string to it so ExportToFile creates a non-empty file.
type contentRecorder struct {
	export.NoopExporter
	w   *bytes.Buffer
	err error
}

func (c *contentRecorder) Start() error {
	if c.err != nil {
		return c.err
	}
	return nil
}

// writeRecorder is an Exporter that writes a known payload via the ExportBase
// pipeline so the output file (or stream) contains checkable bytes.
type writeRecorder struct {
	export.NoopExporter
	payload []byte
	out     *bytes.Buffer
}

func (w *writeRecorder) Start() error {
	if w.out != nil {
		_, err := w.out.Write(w.payload)
		return err
	}
	return nil
}

// ── ExportToFile ──────────────────────────────────────────────────────────────

// TestExportBase_ExportToFile_CreatesFile verifies that ExportToFile creates a
// file at the given path.
func TestExportBase_ExportToFile_CreatesFile(t *testing.T) {
	pp := buildOnePage()
	base := export.NewExportBase()
	rec := newRecorder(new(bytes.Buffer))

	dir := t.TempDir()
	path := filepath.Join(dir, "out.html")

	if err := base.ExportToFile(pp, path, rec); err != nil {
		t.Fatalf("ExportToFile: %v", err)
	}
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("output file not created: %v", err)
	}
}

// TestExportBase_ExportToFile_CallsHooks verifies Start/ExportPageBegin/ExportBand/
// ExportPageEnd/Finish are all invoked.
func TestExportBase_ExportToFile_CallsHooks(t *testing.T) {
	pp := buildOnePage()
	base := export.NewExportBase()
	rec := newRecorder(new(bytes.Buffer))

	dir := t.TempDir()
	path := filepath.Join(dir, "out.html")

	if err := base.ExportToFile(pp, path, rec); err != nil {
		t.Fatalf("ExportToFile: %v", err)
	}
	if !rec.started {
		t.Error("Start not called")
	}
	if !rec.finished {
		t.Error("Finish not called")
	}
	if len(rec.pageBegin) != 1 {
		t.Errorf("ExportPageBegin calls = %d, want 1", len(rec.pageBegin))
	}
}

// TestExportBase_ExportToFile_BadPath verifies that an error is returned when
// the destination path cannot be created (e.g. directory does not exist).
func TestExportBase_ExportToFile_BadPath(t *testing.T) {
	pp := buildOnePage()
	base := export.NewExportBase()
	rec := newRecorder(new(bytes.Buffer))

	// A path whose parent does not exist.
	path := filepath.Join(t.TempDir(), "nonexistent", "subdir", "out.html")
	err := base.ExportToFile(pp, path, rec)
	if err == nil {
		t.Fatal("expected error for bad path, got nil")
	}
}

// TestExportBase_ExportToFile_ExporterError propagates an error from the Exporter.
func TestExportBase_ExportToFile_ExporterError(t *testing.T) {
	pp := buildOnePage()
	base := export.NewExportBase()
	exp := &errExporter{failOn: "start"}

	dir := t.TempDir()
	path := filepath.Join(dir, "out.html")

	err := base.ExportToFile(pp, path, exp)
	if err == nil {
		t.Fatal("expected error from exporter, got nil")
	}
}

// ── ExportAndZip ──────────────────────────────────────────────────────────────

// TestExportBase_ExportAndZip_ProducesZIP verifies that the output stream
// contains a valid ZIP archive.
func TestExportBase_ExportAndZip_ProducesZIP(t *testing.T) {
	pp := buildOnePage()
	base := export.NewExportBase()
	rec := newRecorder(new(bytes.Buffer))

	var buf bytes.Buffer
	if err := base.ExportAndZip(pp, "report.html", rec, &buf); err != nil {
		t.Fatalf("ExportAndZip: %v", err)
	}

	// buf should now contain a valid ZIP archive.
	r, err := zip.NewReader(bytes.NewReader(buf.Bytes()), int64(buf.Len()))
	if err != nil {
		t.Fatalf("output is not a valid ZIP: %v", err)
	}
	if len(r.File) != 1 {
		t.Errorf("ZIP contains %d files, want 1", len(r.File))
	}
}

// TestExportBase_ExportAndZip_EntryName verifies the entry name inside the ZIP.
func TestExportBase_ExportAndZip_EntryName(t *testing.T) {
	pp := buildOnePage()
	base := export.NewExportBase()
	rec := newRecorder(new(bytes.Buffer))

	var buf bytes.Buffer
	const wantName = "myreport.html"
	if err := base.ExportAndZip(pp, wantName, rec, &buf); err != nil {
		t.Fatalf("ExportAndZip: %v", err)
	}

	r, err := zip.NewReader(bytes.NewReader(buf.Bytes()), int64(buf.Len()))
	if err != nil {
		t.Fatalf("invalid ZIP: %v", err)
	}
	if len(r.File) == 0 {
		t.Fatal("ZIP is empty")
	}
	if r.File[0].Name != wantName {
		t.Errorf("ZIP entry name = %q, want %q", r.File[0].Name, wantName)
	}
}

// TestExportBase_ExportAndZip_DefaultFileName verifies that an empty fileName
// argument defaults to "report".
func TestExportBase_ExportAndZip_DefaultFileName(t *testing.T) {
	pp := buildOnePage()
	base := export.NewExportBase()
	rec := newRecorder(new(bytes.Buffer))

	var buf bytes.Buffer
	if err := base.ExportAndZip(pp, "", rec, &buf); err != nil {
		t.Fatalf("ExportAndZip: %v", err)
	}

	r, err := zip.NewReader(bytes.NewReader(buf.Bytes()), int64(buf.Len()))
	if err != nil {
		t.Fatalf("invalid ZIP: %v", err)
	}
	if len(r.File) == 0 {
		t.Fatal("ZIP is empty")
	}
	// When fileName is empty the default is "report".
	if r.File[0].Name != "report" {
		t.Errorf("ZIP entry name = %q, want \"report\"", r.File[0].Name)
	}
}

// TestExportBase_ExportAndZip_ExporterError propagates errors from the exporter.
func TestExportBase_ExportAndZip_ExporterError(t *testing.T) {
	pp := buildOnePage()
	base := export.NewExportBase()
	exp := &errExporter{failOn: "start"}

	var buf bytes.Buffer
	err := base.ExportAndZip(pp, "report.html", exp, &buf)
	if err == nil {
		t.Fatal("expected error from exporter, got nil")
	}
}

// TestExportBase_ExportAndZip_OnProgressStillFires verifies that OnProgress
// works end-to-end when exporting through ExportAndZip.
func TestExportBase_ExportAndZip_OnProgressStillFires(t *testing.T) {
	pp := buildPreparedPages(2, []string{"Band"})
	base := export.NewExportBase()
	rec := newRecorder(new(bytes.Buffer))

	var progressCalls int
	base.OnProgress = func(page, total int) {
		progressCalls++
	}

	var buf bytes.Buffer
	if err := base.ExportAndZip(pp, "r.html", rec, &buf); err != nil {
		t.Fatalf("ExportAndZip: %v", err)
	}
	if progressCalls != 2 {
		t.Errorf("OnProgress called %d times, want 2", progressCalls)
	}
}

// TestExportBase_ExportAndZip_MultiplePageExport verifies ZIP output with
// multiple pages exported at once.
func TestExportBase_ExportAndZip_MultiplePageExport(t *testing.T) {
	pp := buildPreparedPages(3, []string{"B"})
	base := export.NewExportBase()

	// Use an exporter that writes per-page data so we know content was written.
	wrotePages := 0
	pager := &pageCountExporter{onPage: func() { wrotePages++ }}

	var buf bytes.Buffer
	if err := base.ExportAndZip(pp, "report.txt", pager, &buf); err != nil {
		t.Fatalf("ExportAndZip: %v", err)
	}
	if wrotePages != 3 {
		t.Errorf("ExportPageBegin called %d times, want 3", wrotePages)
	}
	// Still produces a valid ZIP.
	if _, err := zip.NewReader(bytes.NewReader(buf.Bytes()), int64(buf.Len())); err != nil {
		t.Fatalf("output is not a valid ZIP: %v", err)
	}
}

// pageCountExporter counts ExportPageBegin calls.
type pageCountExporter struct {
	export.NoopExporter
	onPage func()
}

func (p *pageCountExporter) ExportPageBegin(_ *preview.PreparedPage) error {
	p.onPage()
	return nil
}

// TestExportBase_ExportToFile_AddedToGeneratedFiles verifies that
// ExportToFile does not add the file to GeneratedFiles (the underlying
// Export call resets GeneratedFiles and only concrete exporters add to it).
func TestExportBase_ExportToFile_GeneratedFilesAfterExport(t *testing.T) {
	pp := buildOnePage()
	base := export.NewExportBase()

	type addingExporter struct {
		export.NoopExporter
		base *export.ExportBase
		path string
	}
	ae := &struct {
		export.NoopExporter
		base *export.ExportBase
		path string
	}{base: &base}

	dir := t.TempDir()
	ae.path = filepath.Join(dir, "out.html")

	if err := base.ExportToFile(pp, ae.path, &ae.NoopExporter); err != nil {
		t.Fatalf("ExportToFile: %v", err)
	}
	// GeneratedFiles is empty because NoopExporter doesn't add anything.
	if len(base.GeneratedFiles()) != 0 {
		t.Errorf("GeneratedFiles() = %v, want empty (NoopExporter adds nothing)", base.GeneratedFiles())
	}
}

// TestExportBase_ExportAndZip_WriterError verifies behavior when the destination
// writer fails (e.g. closed pipe).  ExportAndZip should return an error.
func TestExportBase_ExportAndZip_WriterError(t *testing.T) {
	pp := buildOnePage()
	base := export.NewExportBase()
	rec := newRecorder(new(bytes.Buffer))

	// Use a failing writer.
	fw := &failingWriter{}
	err := base.ExportAndZip(pp, "r.html", rec, fw)
	if err == nil {
		t.Fatal("expected error from failing writer, got nil")
	}
}

// failingWriter always returns an error from Write.
type failingWriter struct{}

func (f *failingWriter) Write(_ []byte) (int, error) {
	return 0, fmt.Errorf("write error injected")
}
