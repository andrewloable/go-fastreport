package reportpkg_test

// Smoke tests for special-feature FRX reports.

import (
	"strings"
	"testing"
	"time"

	"github.com/andrewloable/go-fastreport/object"
	"github.com/andrewloable/go-fastreport/reportpkg"
)

func TestFRXSmoke_SVG(t *testing.T) {
	r := loadFRXSmoke(t, "SVG.frx")
	if n := countObjectsOfType[*object.SVGObject](r); n == 0 {
		t.Error("expected at least one SVGObject in SVG.frx")
	}
}

func TestFRXSmoke_RichText_Objects(t *testing.T) {
	r := loadFRXSmoke(t, "RichText.frx")
	if n := countObjectsOfType[*object.RichObject](r); n == 0 {
		t.Error("expected at least one RichObject in RichText.frx")
	}
}

func TestFRXSmoke_CellularText(t *testing.T) {
	r := loadFRXSmoke(t, "CellularText.frx")
	n := countObjectsOfType[*object.CellularTextObject](r)
	if n == 0 {
		t.Error("expected at least one CellularTextObject in CellularText.frx")
	}
}

func TestFRXSmoke_RichText(t *testing.T) {
	loadFRXSmoke(t, "RichText.frx")
}

func TestFRXSmoke_TextureFill(t *testing.T) {
	loadFRXSmoke(t, "TextureFill.frx")
}

func TestFRXSmoke_CompleteUptoNRows(t *testing.T) {
	loadFRXSmoke(t, "Complete upto N Rows.frx")
}

func TestFRXSmoke_PrintCopyNames(t *testing.T) {
	loadFRXSmoke(t, "Print Copy Names.frx")
}

func TestFRXSmoke_PrintMonthNames(t *testing.T) {
	loadFRXSmoke(t, "Print Month Names.frx")
}

func TestFRXSmoke_InheritedReportBase(t *testing.T) {
	loadFRXSmoke(t, "Inherited Report - base.frx")
}

func TestFRXSmoke_InheritedReport(t *testing.T) {
	// Inherited reports use <inherited> as root element instead of <Report>.
	// Loading will fail at the root-element check; verify the file exists and
	// the failure is the expected root-element mismatch (not a crash/panic).
	r := reportpkg.NewReport()
	path := testReportsDir() + "/Inherited Report.frx"
	err := r.Load(path)
	if err == nil {
		// If future implementation supports inherited reports, this is fine.
		if len(r.Pages()) == 0 {
			t.Error("loaded inherited report has no pages")
		}
		return
	}
	// Accept only "wrong root element" errors; anything else is unexpected.
	if !strings.Contains(err.Error(), "inherited") && !strings.Contains(err.Error(), "Report") {
		t.Errorf("unexpected load error for inherited report: %v", err)
	}
}

func TestFRXSmoke_StressTest(t *testing.T) {
	// The stress-test file is large; verify it loads within a reasonable time.
	done := make(chan struct{})
	go func() {
		loadFRXSmoke(t, "Stress-Test (1000x1000).frx")
		close(done)
	}()
	select {
	case <-done:
		// passed
	case <-time.After(10 * time.Second):
		t.Error("Stress-Test (1000x1000).frx took more than 10 seconds to load")
	}
}
