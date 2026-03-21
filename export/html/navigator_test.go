package html_test

// navigator_test.go — Tests for multi-page and navigator HTML export modes.
//
// Tests cover:
//   - ExportModeMultiPage: one file per page written to a temp directory
//   - ExportModeNavigator: per-page files + index.html + <base>.nav.html
//   - ExportModeSingleFile via ExportToDir: produces index.html
//   - ExportMode constants existence
//   - Navigator HTML content (JS vars, buttons, frameset)
//   - Index HTML content (frameset, frame names)
//   - MultiPage index.html (link list)

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/andrewloable/go-fastreport/export/html"
	"github.com/andrewloable/go-fastreport/preview"
)

// buildNavPages builds a PreparedPages with n pages and one band per page.
func buildNavPages(n int) *preview.PreparedPages {
	pp := preview.New()
	for i := 0; i < n; i++ {
		pp.AddPage(794, 1123, i+1)
		_ = pp.AddBand(&preview.PreparedBand{
			Name:   "DataBand",
			Top:    0,
			Height: 40,
		})
	}
	return pp
}

// ── ExportMode constants ──────────────────────────────────────────────────────

func TestExportMode_Constants(t *testing.T) {
	// Verify the three mode constants have distinct values.
	modes := []html.ExportMode{
		html.ExportModeSingleFile,
		html.ExportModeMultiPage,
		html.ExportModeNavigator,
	}
	seen := make(map[html.ExportMode]bool)
	for _, m := range modes {
		if seen[m] {
			t.Errorf("duplicate ExportMode value: %d", m)
		}
		seen[m] = true
	}
}

func TestExportMode_DefaultIsSingleFile(t *testing.T) {
	exp := html.NewExporter()
	if exp.Mode != html.ExportModeSingleFile {
		t.Errorf("default Mode should be ExportModeSingleFile (0), got %d", exp.Mode)
	}
}

// ── ExportToDir — SingleFile mode ─────────────────────────────────────────────

func TestExportToDir_SingleFile_WritesIndexHTML(t *testing.T) {
	dir := t.TempDir()
	pp := buildNavPages(2)

	exp := html.NewExporter()
	exp.Mode = html.ExportModeSingleFile
	if err := exp.ExportToDir(pp, dir); err != nil {
		t.Fatalf("ExportToDir SingleFile: %v", err)
	}

	indexPath := filepath.Join(dir, "index.html")
	if _, err := os.Stat(indexPath); err != nil {
		t.Fatalf("ExportToDir SingleFile: expected index.html, got err: %v", err)
	}
	data, _ := os.ReadFile(indexPath)
	out := string(data)
	if !strings.Contains(out, "<!DOCTYPE HTML") {
		t.Error("index.html should contain DOCTYPE")
	}
	// Both pages present in single file.
	if !strings.Contains(out, `class="frpage0"`) || !strings.Contains(out, `class="frpage1"`) {
		t.Error("SingleFile index.html should contain both frpage0 and frpage1")
	}
}

func TestExportToDir_SingleFile_InGeneratedFiles(t *testing.T) {
	dir := t.TempDir()
	pp := buildNavPages(1)

	exp := html.NewExporter()
	exp.Mode = html.ExportModeSingleFile
	if err := exp.ExportToDir(pp, dir); err != nil {
		t.Fatalf("ExportToDir SingleFile: %v", err)
	}

	gen := exp.GeneratedFiles()
	if len(gen) == 0 {
		t.Error("GeneratedFiles should not be empty after SingleFile export")
	}
}

// ── ExportToDir — MultiPage mode ──────────────────────────────────────────────

func TestExportToDir_MultiPage_CreatesPageFiles(t *testing.T) {
	dir := t.TempDir()
	pp := buildNavPages(3)

	exp := html.NewExporter()
	exp.Mode = html.ExportModeMultiPage
	if err := exp.ExportToDir(pp, dir); err != nil {
		t.Fatalf("ExportToDir MultiPage: %v", err)
	}

	for i := 1; i <= 3; i++ {
		path := filepath.Join(dir, "page"+strings.Repeat("", 0)+string(rune('0'+i))+".html")
		// Use a simpler pattern:
		path = filepath.Join(dir, "page"+itoa(i)+".html")
		if _, err := os.Stat(path); err != nil {
			t.Errorf("expected page file %s, got err: %v", path, err)
		}
	}
}

func TestExportToDir_MultiPage_EachPageIsStandalone(t *testing.T) {
	dir := t.TempDir()
	pp := buildNavPages(2)

	exp := html.NewExporter()
	exp.Mode = html.ExportModeMultiPage
	if err := exp.ExportToDir(pp, dir); err != nil {
		t.Fatalf("ExportToDir MultiPage: %v", err)
	}

	for i := 1; i <= 2; i++ {
		path := filepath.Join(dir, "page"+itoa(i)+".html")
		data, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("read page%d.html: %v", i, err)
		}
		out := string(data)
		if !strings.Contains(out, "<!DOCTYPE HTML") {
			t.Errorf("page%d.html should be standalone (has DOCTYPE)", i)
		}
		if !strings.Contains(out, "<html>") {
			t.Errorf("page%d.html should have <html> tag", i)
		}
		if !strings.Contains(out, "</html>") {
			t.Errorf("page%d.html should have </html> closing tag", i)
		}
	}
}

func TestExportToDir_MultiPage_CreatesIndexHTML(t *testing.T) {
	dir := t.TempDir()
	pp := buildNavPages(2)

	exp := html.NewExporter()
	exp.Mode = html.ExportModeMultiPage
	if err := exp.ExportToDir(pp, dir); err != nil {
		t.Fatalf("ExportToDir MultiPage: %v", err)
	}

	indexPath := filepath.Join(dir, "index.html")
	data, err := os.ReadFile(indexPath)
	if err != nil {
		t.Fatalf("MultiPage index.html not found: %v", err)
	}
	out := string(data)
	// Index should contain links to page files.
	if !strings.Contains(out, "page1.html") {
		t.Error("MultiPage index.html should link to page1.html")
	}
	if !strings.Contains(out, "page2.html") {
		t.Error("MultiPage index.html should link to page2.html")
	}
}

func TestExportToDir_MultiPage_IndexContainsLinks(t *testing.T) {
	dir := t.TempDir()
	pp := buildNavPages(3)

	exp := html.NewExporter()
	exp.Mode = html.ExportModeMultiPage
	if err := exp.ExportToDir(pp, dir); err != nil {
		t.Fatalf("ExportToDir MultiPage: %v", err)
	}

	data, _ := os.ReadFile(filepath.Join(dir, "index.html"))
	out := string(data)
	// C# multi-page non-navigator index: links with "Page N" text.
	if !strings.Contains(out, "Page 1") {
		t.Error("index.html should contain 'Page 1'")
	}
	if !strings.Contains(out, "Page 2") {
		t.Error("index.html should contain 'Page 2'")
	}
	if !strings.Contains(out, "Page 3") {
		t.Error("index.html should contain 'Page 3'")
	}
}

func TestExportToDir_MultiPage_GeneratedFilesCount(t *testing.T) {
	dir := t.TempDir()
	pp := buildNavPages(3)

	exp := html.NewExporter()
	exp.Mode = html.ExportModeMultiPage
	if err := exp.ExportToDir(pp, dir); err != nil {
		t.Fatalf("ExportToDir MultiPage: %v", err)
	}

	gen := exp.GeneratedFiles()
	// 3 page files + 1 index.html = 4 generated files.
	if len(gen) != 4 {
		t.Errorf("expected 4 generated files (3 pages + index), got %d: %v", len(gen), gen)
	}
}

func TestExportToDir_MultiPage_CustomBaseName(t *testing.T) {
	dir := t.TempDir()
	pp := buildNavPages(2)

	exp := html.NewExporter()
	exp.Mode = html.ExportModeMultiPage
	exp.BaseName = "report"
	if err := exp.ExportToDir(pp, dir); err != nil {
		t.Fatalf("ExportToDir MultiPage CustomBaseName: %v", err)
	}

	for i := 1; i <= 2; i++ {
		path := filepath.Join(dir, "report"+itoa(i)+".html")
		if _, err := os.Stat(path); err != nil {
			t.Errorf("expected %s, got: %v", path, err)
		}
	}
}

func TestExportToDir_MultiPage_DefaultBaseName(t *testing.T) {
	dir := t.TempDir()
	pp := buildNavPages(1)

	exp := html.NewExporter()
	exp.Mode = html.ExportModeMultiPage
	// BaseName not set — should default to "page".
	if err := exp.ExportToDir(pp, dir); err != nil {
		t.Fatalf("ExportToDir MultiPage DefaultBaseName: %v", err)
	}

	path := filepath.Join(dir, "page1.html")
	if _, err := os.Stat(path); err != nil {
		t.Errorf("default BaseName should produce page1.html: %v", err)
	}
}

func TestExportToDir_MultiPage_EachPageHasCorrectContent(t *testing.T) {
	// Each page should have its own frpage div (frpage0 in its own file).
	dir := t.TempDir()
	pp := buildNavPages(2)

	exp := html.NewExporter()
	exp.Mode = html.ExportModeMultiPage
	if err := exp.ExportToDir(pp, dir); err != nil {
		t.Fatalf("ExportToDir MultiPage: %v", err)
	}

	for i := 1; i <= 2; i++ {
		data, err := os.ReadFile(filepath.Join(dir, "page"+itoa(i)+".html"))
		if err != nil {
			t.Fatalf("read page%d.html: %v", i, err)
		}
		out := string(data)
		// Each page file should contain exactly one frpage div.
		if !strings.Contains(out, "frpage") {
			t.Errorf("page%d.html should contain frpage div", i)
		}
	}
}

// ── ExportToDir — Navigator mode ──────────────────────────────────────────────

func TestExportToDir_Navigator_CreatesPageFiles(t *testing.T) {
	dir := t.TempDir()
	pp := buildNavPages(3)

	exp := html.NewExporter()
	exp.Mode = html.ExportModeNavigator
	if err := exp.ExportToDir(pp, dir); err != nil {
		t.Fatalf("ExportToDir Navigator: %v", err)
	}

	for i := 1; i <= 3; i++ {
		path := filepath.Join(dir, "page"+itoa(i)+".html")
		if _, err := os.Stat(path); err != nil {
			t.Errorf("Navigator: expected %s, got err: %v", path, err)
		}
	}
}

func TestExportToDir_Navigator_CreatesIndexHTML(t *testing.T) {
	dir := t.TempDir()
	pp := buildNavPages(2)

	exp := html.NewExporter()
	exp.Mode = html.ExportModeNavigator
	if err := exp.ExportToDir(pp, dir); err != nil {
		t.Fatalf("ExportToDir Navigator: %v", err)
	}

	indexPath := filepath.Join(dir, "index.html")
	if _, err := os.Stat(indexPath); err != nil {
		t.Fatalf("Navigator: expected index.html, got err: %v", err)
	}
}

func TestExportToDir_Navigator_CreatesNavHTML(t *testing.T) {
	dir := t.TempDir()
	pp := buildNavPages(2)

	exp := html.NewExporter()
	exp.Mode = html.ExportModeNavigator
	if err := exp.ExportToDir(pp, dir); err != nil {
		t.Fatalf("ExportToDir Navigator: %v", err)
	}

	navPath := filepath.Join(dir, "page.nav.html")
	if _, err := os.Stat(navPath); err != nil {
		t.Fatalf("Navigator: expected page.nav.html, got err: %v", err)
	}
}

func TestExportToDir_Navigator_GeneratedFilesCount(t *testing.T) {
	dir := t.TempDir()
	pp := buildNavPages(3)

	exp := html.NewExporter()
	exp.Mode = html.ExportModeNavigator
	if err := exp.ExportToDir(pp, dir); err != nil {
		t.Fatalf("ExportToDir Navigator: %v", err)
	}

	gen := exp.GeneratedFiles()
	// 3 page files + index.html + page.nav.html = 5 files.
	if len(gen) != 5 {
		t.Errorf("Navigator: expected 5 generated files, got %d: %v", len(gen), gen)
	}
}

func TestExportToDir_Navigator_IndexIsFrameset(t *testing.T) {
	dir := t.TempDir()
	pp := buildNavPages(2)

	exp := html.NewExporter()
	exp.Mode = html.ExportModeNavigator
	if err := exp.ExportToDir(pp, dir); err != nil {
		t.Fatalf("ExportToDir Navigator: %v", err)
	}

	data, _ := os.ReadFile(filepath.Join(dir, "index.html"))
	out := string(data)
	// C# IndexTemplate: frameset with rows="36,*".
	if !strings.Contains(out, "frameset") {
		t.Error("Navigator index.html should be a frameset document")
	}
	if !strings.Contains(out, "topFrame") {
		t.Error("Navigator index.html should have topFrame")
	}
	if !strings.Contains(out, "mainFrame") {
		t.Error("Navigator index.html should have mainFrame")
	}
	if !strings.Contains(out, "36,*") {
		t.Error("Navigator index.html frameset should use rows=\"36,*\"")
	}
}

func TestExportToDir_Navigator_IndexLinksNavAndFirstPage(t *testing.T) {
	dir := t.TempDir()
	pp := buildNavPages(2)

	exp := html.NewExporter()
	exp.Mode = html.ExportModeNavigator
	exp.BaseName = "rep"
	if err := exp.ExportToDir(pp, dir); err != nil {
		t.Fatalf("ExportToDir Navigator: %v", err)
	}

	data, _ := os.ReadFile(filepath.Join(dir, "index.html"))
	out := string(data)
	// topFrame src = "rep.nav.html"
	if !strings.Contains(out, "rep.nav.html") {
		t.Error("Navigator index.html topFrame should reference rep.nav.html")
	}
	// mainFrame src = "rep1.html" (first page)
	if !strings.Contains(out, "rep1.html") {
		t.Error("Navigator index.html mainFrame should reference rep1.html")
	}
}

func TestExportToDir_Navigator_NavHTMLHasJSVariables(t *testing.T) {
	dir := t.TempDir()
	pp := buildNavPages(4)

	exp := html.NewExporter()
	exp.Mode = html.ExportModeNavigator
	exp.Title = "My Report"
	if err := exp.ExportToDir(pp, dir); err != nil {
		t.Fatalf("ExportToDir Navigator: %v", err)
	}

	data, _ := os.ReadFile(filepath.Join(dir, "page.nav.html"))
	out := string(data)

	// C# NavigatorTemplate: var frCurPage = 1; var frPgCnt = {0}; var frRepName = "{1}"; var frMultipage = {2};
	if !strings.Contains(out, "frCurPage = 1") {
		t.Error("nav.html should set frCurPage = 1")
	}
	if !strings.Contains(out, "frPgCnt = 4") {
		t.Error("nav.html should set frPgCnt = 4 (4 pages)")
	}
	if !strings.Contains(out, `frRepName = "My Report"`) {
		t.Error("nav.html should set frRepName to the report title")
	}
	if !strings.Contains(out, "frMultipage = 1") {
		t.Error("nav.html should set frMultipage = 1")
	}
}

func TestExportToDir_Navigator_NavHTMLHasNavigationButtons(t *testing.T) {
	dir := t.TempDir()
	pp := buildNavPages(2)

	exp := html.NewExporter()
	exp.Mode = html.ExportModeNavigator
	if err := exp.ExportToDir(pp, dir); err != nil {
		t.Fatalf("ExportToDir Navigator: %v", err)
	}

	data, _ := os.ReadFile(filepath.Join(dir, "page.nav.html"))
	out := string(data)

	// C# NavigatorTemplate buttons: First, Prev, Next, Last.
	for _, btn := range []string{"bFirst", "bPrev", "bNext", "bLast"} {
		if !strings.Contains(out, btn) {
			t.Errorf("nav.html should contain button %s", btn)
		}
	}
}

func TestExportToDir_Navigator_NavHTMLHasDoPageFunction(t *testing.T) {
	dir := t.TempDir()
	pp := buildNavPages(2)

	exp := html.NewExporter()
	exp.Mode = html.ExportModeNavigator
	if err := exp.ExportToDir(pp, dir); err != nil {
		t.Fatalf("ExportToDir Navigator: %v", err)
	}

	data, _ := os.ReadFile(filepath.Join(dir, "page.nav.html"))
	out := string(data)
	// C# NavigatorTemplate: function DoPage(PgN) {...}
	if !strings.Contains(out, "function DoPage") {
		t.Error("nav.html should contain DoPage function")
	}
	if !strings.Contains(out, "function UpdateNav") {
		t.Error("nav.html should contain UpdateNav function")
	}
}

func TestExportToDir_Navigator_NavHTMLPrefix(t *testing.T) {
	dir := t.TempDir()
	pp := buildNavPages(2)

	exp := html.NewExporter()
	exp.Mode = html.ExportModeNavigator
	exp.BaseName = "myreport"
	if err := exp.ExportToDir(pp, dir); err != nil {
		t.Fatalf("ExportToDir Navigator: %v", err)
	}

	navPath := filepath.Join(dir, "myreport.nav.html")
	data, err := os.ReadFile(navPath)
	if err != nil {
		t.Fatalf("expected myreport.nav.html: %v", err)
	}
	out := string(data)
	// frPrefix should be the base name (used to navigate to page files).
	if !strings.Contains(out, `frPrefix="myreport"`) {
		t.Error("nav.html should set frPrefix to the BaseName")
	}
}

// ── ExportToDir — error conditions ────────────────────────────────────────────

func TestExportToDir_NilPages_ReturnsError(t *testing.T) {
	dir := t.TempDir()
	exp := html.NewExporter()
	exp.Mode = html.ExportModeMultiPage
	if err := exp.ExportToDir(nil, dir); err == nil {
		t.Error("expected error for nil PreparedPages")
	}
}

func TestExportToDir_CreatesDir_IfNotExist(t *testing.T) {
	parent := t.TempDir()
	dir := filepath.Join(parent, "newsubdir")
	pp := buildNavPages(1)

	exp := html.NewExporter()
	exp.Mode = html.ExportModeMultiPage
	if err := exp.ExportToDir(pp, dir); err != nil {
		t.Fatalf("ExportToDir should create missing directory: %v", err)
	}
	if _, err := os.Stat(dir); err != nil {
		t.Error("ExportToDir should have created the directory")
	}
}

// ── Single-page (existing Export method) still works ─────────────────────────

func TestExportMode_SingleFile_ExportMethod_Unchanged(t *testing.T) {
	// The original Export(pp, w) method should still work as before.
	pp := buildNavPages(2)
	exp := html.NewExporter()
	// Mode is ExportModeSingleFile by default — Export to io.Writer.
	var buf strings.Builder
	if err := exp.Export(pp, &buf); err != nil {
		t.Fatalf("Export: %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, `class="frpage0"`) || !strings.Contains(out, `class="frpage1"`) {
		t.Error("single-file Export should include both pages")
	}
}

// ── helpers ───────────────────────────────────────────────────────────────────

// itoa converts an int to string (avoids importing strconv in tests).
func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	neg := n < 0
	if neg {
		n = -n
	}
	var buf [20]byte
	pos := len(buf)
	for n > 0 {
		pos--
		buf[pos] = byte('0' + n%10)
		n /= 10
	}
	if neg {
		pos--
		buf[pos] = '-'
	}
	return string(buf[pos:])
}
