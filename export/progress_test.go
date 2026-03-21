package export_test

// progress_test.go — tests for the ExportBase.OnProgress callback.
//
// ExportBase.OnProgress is the Go equivalent of the C# pattern:
//   Config.ReportSettings.OnProgress(Report, message, pageNo, totalPages)
// called once per page during ExportBase.Export().
// C# ref: FastReport.Base/Export/ExportBase.cs, Export() method.
// Ported during go-fastreport-yixy (ReportEventArgs / ProgressEventArgs).

import (
	"bytes"
	"testing"

	"github.com/andrewloable/go-fastreport/export"
)

// TestExportBase_OnProgress_InvokedPerPage verifies that OnProgress is called
// once for each exported page with the correct (current, total) values.
func TestExportBase_OnProgress_InvokedPerPage(t *testing.T) {
	const numPages = 3
	pp := buildPreparedPages(numPages, []string{"Band"})

	base := export.NewExportBase()

	var calls [][2]int
	base.OnProgress = func(page, total int) {
		calls = append(calls, [2]int{page, total})
	}

	if err := base.Export(pp, new(bytes.Buffer), &export.NoopExporter{}); err != nil {
		t.Fatalf("Export: %v", err)
	}

	if len(calls) != numPages {
		t.Fatalf("OnProgress called %d times, want %d", len(calls), numPages)
	}
	for i, c := range calls {
		wantPage := i + 1
		if c[0] != wantPage {
			t.Errorf("call[%d]: page = %d, want %d", i, c[0], wantPage)
		}
		if c[1] != numPages {
			t.Errorf("call[%d]: total = %d, want %d", i, c[1], numPages)
		}
	}
}

// TestExportBase_OnProgress_NilIsNoop verifies that Export succeeds when
// OnProgress is nil (no regression).
func TestExportBase_OnProgress_NilIsNoop(t *testing.T) {
	pp := buildPreparedPages(2, []string{"Band"})

	base := export.NewExportBase()
	// OnProgress is nil by default.

	if err := base.Export(pp, new(bytes.Buffer), &export.NoopExporter{}); err != nil {
		t.Fatalf("Export with nil OnProgress: %v", err)
	}
}

// TestExportBase_OnProgress_SubsetPages verifies that OnProgress reflects the
// number of actually-exported pages (not total pages in the PreparedPages)
// when PageRange is set to PageRangeCustom.
func TestExportBase_OnProgress_SubsetPages(t *testing.T) {
	pp := buildPreparedPages(5, []string{"Band"})

	base := export.NewExportBase()
	base.PageRange = export.PageRangeCustom
	base.PageNumbers = "1,3" // export only 2 of 5 pages

	var calls [][2]int
	base.OnProgress = func(page, total int) {
		calls = append(calls, [2]int{page, total})
	}

	if err := base.Export(pp, new(bytes.Buffer), &export.NoopExporter{}); err != nil {
		t.Fatalf("Export: %v", err)
	}

	if len(calls) != 2 {
		t.Fatalf("OnProgress called %d times, want 2", len(calls))
	}
	// total should be 2 (the number of selected pages).
	for i, c := range calls {
		if c[1] != 2 {
			t.Errorf("call[%d]: total = %d, want 2", i, c[1])
		}
	}
}
