package reportpkg_test

import (
	"errors"
	"testing"

	"github.com/andrewloable/go-fastreport/band"
	"github.com/andrewloable/go-fastreport/reportpkg"
)

// ── helpers ────────────────────────────────────────────────────────────────

func makeLoader(reports map[string]*reportpkg.Report) reportpkg.ReportLoaderFunc {
	return func(path string) (*reportpkg.Report, error) {
		r, ok := reports[path]
		if !ok {
			return nil, errors.New("report not found: " + path)
		}
		return r, nil
	}
}

func namedPage(name string) *reportpkg.ReportPage {
	p := reportpkg.NewReportPage()
	p.SetName(name)
	return p
}

// ── ApplyBase – scalar Info merge ─────────────────────────────────────────

func TestApplyBase_Info_ChildWins(t *testing.T) {
	base := reportpkg.NewReport()
	base.Info.Name = "Base Report"
	base.Info.Author = "Base Author"

	child := reportpkg.NewReport()
	child.Info.Name = "Child Report"

	child.ApplyBase(base)

	if child.Info.Name != "Child Report" {
		t.Errorf("Name = %q, want Child Report", child.Info.Name)
	}
	// Author not set in child → inherited from base.
	if child.Info.Author != "Base Author" {
		t.Errorf("Author = %q, want Base Author", child.Info.Author)
	}
}

func TestApplyBase_Info_Version(t *testing.T) {
	base := reportpkg.NewReport()
	base.Info.Version = "2.0"

	child := reportpkg.NewReport()
	child.ApplyBase(base)

	if child.Info.Version != "2.0" {
		t.Errorf("Version = %q, want 2.0", child.Info.Version)
	}
}

func TestApplyBase_Info_Description(t *testing.T) {
	base := reportpkg.NewReport()
	base.Info.Description = "A base description"

	child := reportpkg.NewReport()
	child.ApplyBase(base)

	if child.Info.Description != "A base description" {
		t.Errorf("Description = %q", child.Info.Description)
	}
}

// ── ApplyBase – flag inheritance ────────────────────────────────────────

func TestApplyBase_Flags_BaseDoublePass(t *testing.T) {
	base := reportpkg.NewReport()
	base.DoublePass = true

	child := reportpkg.NewReport()
	child.ApplyBase(base)

	if !child.DoublePass {
		t.Error("DoublePass should be inherited from base")
	}
}

func TestApplyBase_Flags_ChildDoublePassWins(t *testing.T) {
	base := reportpkg.NewReport()
	base.DoublePass = false

	child := reportpkg.NewReport()
	child.DoublePass = true
	child.ApplyBase(base)

	if !child.DoublePass {
		t.Error("Child's DoublePass=true should win")
	}
}

func TestApplyBase_Flags_MaxPages(t *testing.T) {
	base := reportpkg.NewReport()
	base.MaxPages = 5

	child := reportpkg.NewReport()
	child.ApplyBase(base)

	if child.MaxPages != 5 {
		t.Errorf("MaxPages = %d, want 5", child.MaxPages)
	}
}

func TestApplyBase_Events(t *testing.T) {
	base := reportpkg.NewReport()
	base.StartReportEvent = "OnStart"
	base.FinishReportEvent = "OnFinish"

	child := reportpkg.NewReport()
	child.ApplyBase(base)

	if child.StartReportEvent != "OnStart" {
		t.Errorf("StartReportEvent = %q", child.StartReportEvent)
	}
	if child.FinishReportEvent != "OnFinish" {
		t.Errorf("FinishReportEvent = %q", child.FinishReportEvent)
	}
}

// ── ApplyBase – page merging ────────────────────────────────────────────

func TestApplyBase_Pages_BasePagesInserted(t *testing.T) {
	base := reportpkg.NewReport()
	base.AddPage(namedPage("cover"))
	base.AddPage(namedPage("content"))

	child := reportpkg.NewReport()
	child.ApplyBase(base)

	if child.PageCount() != 2 {
		t.Errorf("PageCount = %d, want 2", child.PageCount())
	}
	if child.Page(0).Name() != "cover" {
		t.Errorf("Page(0) name = %q, want cover", child.Page(0).Name())
	}
	if !child.Page(0).Inherited() {
		t.Error("base-only pages should be marked inherited")
	}
}

func TestApplyBase_Pages_ChildPageKept(t *testing.T) {
	base := reportpkg.NewReport()
	base.AddPage(namedPage("cover"))

	child := reportpkg.NewReport()
	child.AddPage(namedPage("myPage"))
	child.ApplyBase(base)

	// Child's own page + base's page = 2.
	if child.PageCount() != 2 {
		t.Errorf("PageCount = %d, want 2", child.PageCount())
	}
}

func TestApplyBase_Pages_MatchingPageMerged(t *testing.T) {
	base := reportpkg.NewReport()
	basePage := namedPage("shared")
	basePage.SetPageHeader(band.NewPageHeaderBand())
	base.AddPage(basePage)

	child := reportpkg.NewReport()
	childPage := namedPage("shared")
	child.AddPage(childPage)
	child.ApplyBase(base)

	// Should have merged — not duplicated.
	if child.PageCount() != 1 {
		t.Errorf("PageCount = %d, want 1 (merged)", child.PageCount())
	}
	if child.Page(0).PageHeader() == nil {
		t.Error("shared page should have inherited PageHeader from base")
	}
}

func TestApplyBase_Pages_Order_BasePrependedFirst(t *testing.T) {
	base := reportpkg.NewReport()
	base.AddPage(namedPage("basePage"))

	child := reportpkg.NewReport()
	child.AddPage(namedPage("childPage"))
	child.ApplyBase(base)

	if child.Page(0).Name() != "basePage" {
		t.Errorf("expected basePage first, got %q", child.Page(0).Name())
	}
	if child.Page(1).Name() != "childPage" {
		t.Errorf("expected childPage second, got %q", child.Page(1).Name())
	}
}

// ── ReportPage.Clone ────────────────────────────────────────────────────

func TestReportPage_Clone(t *testing.T) {
	orig := namedPage("page1")
	orig.PaperWidth = 100
	clone := orig.Clone()

	if clone.Name() != "page1" {
		t.Errorf("Clone Name = %q, want page1", clone.Name())
	}
	if clone.PaperWidth != 100 {
		t.Errorf("Clone PaperWidth = %v", clone.PaperWidth)
	}
	// Modifying clone should not affect original.
	clone.PaperWidth = 200
	if orig.PaperWidth != 100 {
		t.Error("Clone should not share PaperWidth with original")
	}
}

// ── LoadBase ───────────────────────────────────────────────────────────

func TestLoadBase_NoBaseReportPath(t *testing.T) {
	child := reportpkg.NewReport()
	if err := child.LoadBase(makeLoader(nil)); err != nil {
		t.Errorf("LoadBase with empty path should succeed: %v", err)
	}
}

func TestLoadBase_LoadsAndApplies(t *testing.T) {
	base := reportpkg.NewReport()
	base.Info.Name = "Base"
	base.AddPage(namedPage("basePage"))

	loader := makeLoader(map[string]*reportpkg.Report{
		"base.frx": base,
	})

	child := reportpkg.NewReport()
	child.BaseReportPath = "base.frx"
	if err := child.LoadBase(loader); err != nil {
		t.Fatalf("LoadBase: %v", err)
	}

	if child.Info.Name != "Base" {
		t.Errorf("Info.Name = %q, want Base", child.Info.Name)
	}
	if child.PageCount() != 1 {
		t.Errorf("PageCount = %d, want 1", child.PageCount())
	}
}

func TestLoadBase_LoaderError(t *testing.T) {
	child := reportpkg.NewReport()
	child.BaseReportPath = "missing.frx"
	err := child.LoadBase(makeLoader(map[string]*reportpkg.Report{}))
	if err == nil {
		t.Error("expected error for missing base report")
	}
}

// ── ReportLoaderFunc ─────────────────────────────────────────────────────

func TestReportLoaderFunc(t *testing.T) {
	base := reportpkg.NewReport()
	var fn reportpkg.ReportLoaderFunc = func(path string) (*reportpkg.Report, error) {
		if path == "ok" {
			return base, nil
		}
		return nil, errors.New("not found")
	}
	r, err := fn.Load("ok")
	if err != nil || r != base {
		t.Error("ReportLoaderFunc.Load should return the report")
	}
	_, err = fn.Load("missing")
	if err == nil {
		t.Error("expected error for missing path")
	}
}
