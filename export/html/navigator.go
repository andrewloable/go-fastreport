package html

// navigator.go — Multi-page and navigator mode export for the HTML exporter.
//
// C# source reference: FastReport.Base/Export/Html/HTMLExport.cs
//   - ExportHTMLNavigator (line 634-645): navigator HTML page template
//   - ExportHTMLIndex     (line 626-632): frameset index page template
//   - Finish()            (line 942-1016): navigator/multi-page Finish logic
//
// C# source reference: FastReport.Base/Export/Html/HTMLExportTemplates.cs
//   - HtmlTemplates.NavigatorTemplate (lines 103-139)
//   - HtmlTemplates.IndexTemplate     (lines 147-161)
//
// Design:
//   ExportModeSingleFile  (default) — no change, streams to io.Writer.
//   ExportModeMultiPage   — each page rendered as a separate complete HTML
//                           file (<BaseName>N.html) written to OutputDir.
//   ExportModeNavigator   — same per-page files + index.html (frameset) +
//                           <BaseName>.nav.html (JS navigator bar).
//
// The ExportToDir helper is the entry point for directory-based modes.

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/andrewloable/go-fastreport/preview"
)

// ExportToDir exports the PreparedPages to dir using the current Mode.
//
// For ExportModeMultiPage it writes one HTML file per page:
//
//	<dir>/<BaseName>1.html, <dir>/<BaseName>2.html, ...
//
// For ExportModeNavigator it additionally writes:
//
//	<dir>/index.html              — frameset entry point
//	<dir>/<BaseName>.nav.html     — JavaScript navigation bar
//
// In ExportModeSingleFile ExportToDir writes the single-page document to
// <dir>/index.html (same as calling Export with a file writer).
//
// C# reference: HTMLExport.cs Start()/Finish() with Navigator=true,
// SinglePage=false.
func (e *Exporter) ExportToDir(pp *preview.PreparedPages, dir string) error {
	if pp == nil {
		return fmt.Errorf("export/html: prepared pages is nil")
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("export/html: create output dir: %w", err)
	}
	e.OutputDir = dir
	e.pp = pp

	base := e.BaseName
	if base == "" {
		base = "page"
	}

	switch e.Mode {
	case ExportModeMultiPage, ExportModeNavigator:
		return e.exportMultiPage(pp, dir, base)
	default: // ExportModeSingleFile
		path := filepath.Join(dir, "index.html")
		f, err := os.Create(path)
		if err != nil {
			return fmt.Errorf("export/html: create index.html: %w", err)
		}
		defer f.Close()
		if err := e.Export(pp, f); err != nil {
			return err
		}
		// AddGeneratedFile after Export because Export resets generatedFiles at startup.
		e.AddGeneratedFile(path)
		return nil
	}
}

// exportMultiPage drives the multi-page export: one HTML file per page plus,
// in navigator mode, the index and nav pages.
//
// C# reference: HTMLExport.cs Finish() — the non-navigator multi-page branch
// (lines 997-1012) iterates GeneratedFiles and writes <a href=...> links.
// For navigator mode it calls ExportHTMLIndex + ExportHTMLNavigator.
func (e *Exporter) exportMultiPage(pp *preview.PreparedPages, dir, base string) error {
	// Resolve page indices via ExportBase.
	if err := e.ExportBase.Export(pp, nil, &multiPageAdapter{exp: e, dir: dir, base: base}); err != nil {
		return err
	}
	return nil
}

// multiPageAdapter implements export.Exporter so that ExportBase.Export drives
// per-page rendering.  Each page is rendered into a complete standalone HTML
// file.  After all pages, Finish writes the navigator/index files if needed.
type multiPageAdapter struct {
	exp      *Exporter
	dir      string
	base     string
	pageNums []int // 1-based page numbers in export order (for navigator)
}

func (a *multiPageAdapter) Start() error {
	a.exp.sb.Reset()
	a.exp.pageIdx = 0
	a.exp.css = newCSSRegistry()
	a.pageNums = a.pageNums[:0]
	return nil
}

func (a *multiPageAdapter) ExportPageBegin(pg *preview.PreparedPage) error {
	return a.exp.ExportPageBegin(pg)
}

func (a *multiPageAdapter) ExportBand(b *preview.PreparedBand) error {
	return a.exp.ExportBand(b)
}

// ExportPageEnd renders the page content and writes it to a standalone file.
//
// C# reference: HTMLExport.cs ExportHTMLPageFinal() with !singlePage branch —
// writes pageFileName = targetIndexPath + targetFileName + pageNumber + ".html".
func (a *multiPageAdapter) ExportPageEnd(pg *preview.PreparedPage) error {
	// Record the current length of the main output buffer before this page's
	// CSS + content is appended by ExportPageEnd.
	priorLen := a.exp.sb.Len()

	if err := a.exp.ExportPageEnd(pg); err != nil {
		return err
	}
	// pageIdx was already incremented in ExportPageBegin.
	pageNum := a.exp.pageIdx // 1-based page number
	a.pageNums = append(a.pageNums, pageNum)

	// Extract only this page's contribution (everything appended after priorLen).
	// For page 1 this includes the frpage-container div open; for subsequent pages
	// it is just CSS + page content.  We wrap it in a fresh standalone document.
	allSoFar := a.exp.sb.String()
	thisPageFragment := allSoFar[priorLen:]

	// Reset both output buffers so that the next page starts from a clean slate.
	// This prevents accumulation across pages and mirrors C#'s per-file stream.
	// Both sb and pageBuf must be cleared because ExportPageEnd swaps them and
	// ExportPageBegin will swap again; stale content in either buffer corrupts the next page.
	a.exp.sb.Reset()
	a.exp.pageBuf.Reset()

	pageContent := buildStandalonePage(a.exp.Title, thisPageFragment)
	path := filepath.Join(a.dir, fmt.Sprintf("%s%d.html", a.base, pageNum))
	if err := writeFile(path, pageContent); err != nil {
		return fmt.Errorf("export/html: write page %d: %w", pageNum, err)
	}
	a.exp.AddGeneratedFile(path)
	return nil
}

// Finish writes the navigator/index files for ExportModeNavigator, or a simple
// link-list index for ExportModeMultiPage, matching C# behavior.
//
// C# reference:
//   - Navigator mode:  HTMLExport.cs Finish() lines 947-976
//                      (ExportHTMLIndex + ExportHTMLNavigator)
//   - Multi-page mode: HTMLExport.cs Finish() lines 996-1012
//                      (simple <a href=...> link list written to Stream)
func (a *multiPageAdapter) Finish() error {
	switch a.exp.Mode {
	case ExportModeNavigator:
		return a.finishNavigator()
	default: // ExportModeMultiPage
		return a.finishMultiPage()
	}
}

// finishMultiPage writes a simple index.html with links to each page file.
// C# reference: HTMLExport.cs Finish() lines 996-1012 — non-navigator, non-singlePage
// branch writes <a href="pageN.html">Page N</a><br /> for each generated html.
func (a *multiPageAdapter) finishMultiPage() error {
	var sb strings.Builder
	sb.WriteString("<!DOCTYPE HTML PUBLIC \"-//W3C//DTD HTML 4.01 Transitional//EN\">\n")
	sb.WriteString("<html><head>\n")
	sb.WriteString("<meta http-equiv=\"Content-Type\" content=\"text/html; charset=utf-8\">\n")
	sb.WriteString("<meta name=Generator content=\"FastReport http://www.fast-report.com\">\n")
	sb.WriteString(fmt.Sprintf("<title>%s</title>\n", a.exp.Title))
	sb.WriteString("</head>\r\n<body bgcolor=\"#FFFFFF\" text=\"#000000\">\r\n")
	for i, n := range a.pageNums {
		fname := fmt.Sprintf("%s%d.html", a.base, n)
		sb.WriteString(fmt.Sprintf("<a href=\"%s\">Page %d</a><br />\n", fname, i+1))
	}
	sb.WriteString("</body>\r\n</html>\n")

	path := filepath.Join(a.dir, "index.html")
	if err := writeFile(path, sb.String()); err != nil {
		return fmt.Errorf("export/html: write index: %w", err)
	}
	a.exp.AddGeneratedFile(path)
	return nil
}

// finishNavigator writes the frameset index.html and the JS navigator bar.
//
// C# reference:
//   - ExportHTMLIndex: HTMLExport.cs lines 626-632 — frameset with topFrame
//     (navigator) and mainFrame (first page or .main.html).
//   - ExportHTMLNavigator: HTMLExport.cs lines 634-645 — JS navigator bar
//     with First/Prev/Next/Last buttons using DoPage().
//   - HtmlTemplates: HTMLExportTemplates.cs IndexTemplate + NavigatorTemplate.
func (a *multiPageAdapter) finishNavigator() error {
	pageCount := len(a.pageNums)
	navFile := a.base + ".nav.html"
	firstPageFile := fmt.Sprintf("%s1.html", a.base)

	// index.html — frameset
	indexHTML := buildIndexHTML(a.exp.Title, navFile, firstPageFile)
	indexPath := filepath.Join(a.dir, "index.html")
	if err := writeFile(indexPath, indexHTML); err != nil {
		return fmt.Errorf("export/html: write index: %w", err)
	}
	a.exp.AddGeneratedFile(indexPath)

	// <base>.nav.html — JavaScript navigator bar
	navHTML := buildNavigatorHTML(a.exp.Title, a.base, pageCount)
	navPath := filepath.Join(a.dir, navFile)
	if err := writeFile(navPath, navHTML); err != nil {
		return fmt.Errorf("export/html: write navigator: %w", err)
	}
	a.exp.AddGeneratedFile(navPath)
	return nil
}

// ── HTML template builders ────────────────────────────────────────────────────

// buildStandalonePage wraps a page's CSS+content fragment into a complete HTML
// document.  This is the per-page file written in multi-page/navigator modes.
//
// C# reference: HTMLExport.cs DoPage() + DoPageStart() + DoPageEnd() sequence,
// which writes PageTemplateTitle, the CSS/content, then PageTemplateFooter.
func buildStandalonePage(title, content string) string {
	var sb strings.Builder
	sb.WriteString("<!DOCTYPE HTML PUBLIC \"-//W3C//DTD HTML 4.01 Transitional//EN\">\n")
	sb.WriteString("<html><head>\n")
	sb.WriteString("<meta http-equiv=\"Content-Type\" content=\"text/html; charset=utf-8\">\n")
	sb.WriteString("<meta name=Generator content=\"FastReport http://www.fast-report.com\">\n")
	sb.WriteString(fmt.Sprintf("<title>%s</title>\n", title))
	sb.WriteString("</head>\r\n<body bgcolor=\"#FFFFFF\" text=\"#000000\">\r\n")
	sb.WriteString(content)
	sb.WriteString("</body>\r\n</html>\n")
	return sb.String()
}

// buildIndexHTML generates the frameset index page.
//
// C# reference: HTMLExportTemplates.cs IndexTemplate (lines 150-160):
//
//	{0} = title, {1} = navigator frame src, {2} = main frame src
//
// The frameset uses rows="36,*": 36px top row for the navigator bar, rest for content.
func buildIndexHTML(title, navFile, mainFile string) string {
	var sb strings.Builder
	sb.WriteString("<!DOCTYPE HTML PUBLIC \"-//W3C//DTD HTML 4.01 Frameset//EN\" \"http://www.w3.org/TR/html4/frameset.dtd\"\n")
	sb.WriteString("<html><head>\n")
	sb.WriteString("<meta http-equiv=\"Content-Type\" content=\"text/html; charset=utf-8\">\n")
	sb.WriteString("<meta name=Generator content=\"FastReport http://www.fast-report.com\">\n")
	sb.WriteString(fmt.Sprintf("<title>%s</title>\n", title))
	sb.WriteString("<frameset rows=\"36,*\" cols=\"*\">\n")
	sb.WriteString(fmt.Sprintf("<frame name=\"topFrame\" src=\"%s\" noresize frameborder=\"0\" scrolling=\"no\">\n", navFile))
	sb.WriteString(fmt.Sprintf("<frame name=\"mainFrame\" src=\"%s\" frameborder=\"0\">\n", mainFile))
	sb.WriteString("</frameset>\n")
	sb.WriteString("</html>\n")
	return sb.String()
}

// buildNavigatorHTML generates the JavaScript navigation bar page.
//
// C# reference: HTMLExportTemplates.cs NavigatorTemplate (lines 103-139):
//
//	frCurPage=1, frPgCnt={pageCount}, frRepName={title}, frMultipage=1,
//	frPrefix={base} — DoPage() navigates to frPrefix+N+".html" (multipage=1).
//
// Navigation buttons: First, Prev, [page input], Next, Last with total count.
func buildNavigatorHTML(title, base string, pageCount int) string {
	var sb strings.Builder
	sb.WriteString("<!DOCTYPE HTML PUBLIC \"-//W3C//DTD HTML 4.01 Transitional//EN\">\n")
	sb.WriteString("<html><head>\n")
	sb.WriteString("<meta http-equiv=\"Content-Type\" content=\"text/html; charset=utf-8\">\n")
	sb.WriteString("<meta name=Generator content=\"FastReport http://www.fast-report.com\">\n")
	sb.WriteString("<title></title><style type=\"text/css\"><!--\n")
	sb.WriteString("body,input,select { font-family:\"Lucida Grande\",Calibri,Arial,sans-serif; font-size: 8px; font-weight: bold; font-style: normal; text-align: center; vertical-align: middle; }\n")
	sb.WriteString("input {text-align: center}\n")
	sb.WriteString(".nav { font-size : 9pt; color : #283e66; font-weight : bold; text-decoration : none;}\n")
	sb.WriteString("--></style><script language=\"javascript\" type=\"text/javascript\"><!--\n")
	sb.WriteString(fmt.Sprintf("  var frCurPage = 1; var frPgCnt = %d; var frRepName = \"%s\"; var frMultipage = 1; var frPrefix=\"%s\";\n",
		pageCount, title, base))
	sb.WriteString("  function DoPage(PgN) {\n")
	sb.WriteString("    if ((PgN > 0) && (PgN <= frPgCnt) && (PgN != frCurPage)) {\n")
	sb.WriteString("      if (frMultipage > 0)  parent.mainFrame.location = frPrefix + PgN + \".html\";\n")
	sb.WriteString("      else parent.mainFrame.location = frPrefix + \".main.html#PageN\" + PgN;\n")
	sb.WriteString("      UpdateNav(PgN); } else document.PgForm.PgEdit.value = frCurPage; }\n")
	sb.WriteString("  function UpdateNav(PgN) {\n")
	sb.WriteString("    frCurPage = PgN; document.PgForm.PgEdit.value = PgN;\n")
	sb.WriteString("    if (PgN == 1) { document.PgForm.bFirst.disabled = 1; document.PgForm.bPrev.disabled = 1; }\n")
	sb.WriteString("    else { document.PgForm.bFirst.disabled = 0; document.PgForm.bPrev.disabled = 0; }\n")
	sb.WriteString("    if (PgN == frPgCnt) { document.PgForm.bNext.disabled = 1; document.PgForm.bLast.disabled = 1; }\n")
	sb.WriteString("    else { document.PgForm.bNext.disabled = 0; document.PgForm.bLast.disabled = 0; } }\n")
	sb.WriteString("--></script></head>\n")
	sb.WriteString("<body bgcolor=\"#DDDDDD\" text=\"#000000\" leftmargin=\"0\" topmargin=\"4\" onload=\"UpdateNav(frCurPage)\">\n")
	sb.WriteString("<form name=\"PgForm\" onsubmit=\"DoPage(document.forms[0].PgEdit.value); return false;\" action=\"\">\n")
	sb.WriteString("<table cellspacing=\"0\" align=\"left\" cellpadding=\"0\" border=\"0\" width=\"100%\">\n")
	sb.WriteString("<tr valign=\"middle\">\n")
	sb.WriteString("<td width=\"60\" align=\"center\"><button name=\"bFirst\" class=\"nav\" type=\"button\" onclick=\"DoPage(1); return false;\"><b>First</b></button></td>\n")
	sb.WriteString("<td width=\"60\" align=\"center\"><button name=\"bPrev\" class=\"nav\" type=\"button\" onclick=\"DoPage(Math.max(frCurPage - 1, 1)); return false;\"><b>Prev</b></button></td>\n")
	sb.WriteString("<td width=\"100\" align=\"center\"><input type=\"text\" class=\"nav\" name=\"PgEdit\" value=\"frCurPage\" size=\"4\"></td>\n")
	sb.WriteString("<td width=\"60\" align=\"center\"><button name=\"bNext\" class=\"nav\" type=\"button\" onclick=\"DoPage(frCurPage + 1); return false;\"><b>Next</b></button></td>\n")
	sb.WriteString("<td width=\"60\" align=\"center\"><button name=\"bLast\" class=\"nav\" type=\"button\" onclick=\"DoPage(frPgCnt); return false;\"><b>Last</b></button></td>\n")
	sb.WriteString("<td width=\"20\">&nbsp;</td>\r\n")
	sb.WriteString("<td align=\"right\">Total: <script language=\"javascript\" type=\"text/javascript\"> document.write(frPgCnt);</script></td>\n")
	sb.WriteString("<td width=\"10\">&nbsp;</td>\n")
	sb.WriteString("</tr></table></form></body></html>\n")
	return sb.String()
}

// writeFile creates or overwrites path with content.
func writeFile(path, content string) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	_, werr := io.WriteString(f, content)
	cerr := f.Close()
	if werr != nil {
		return werr
	}
	return cerr
}
