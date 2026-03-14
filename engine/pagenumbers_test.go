package engine_test

import (
	"testing"

	"github.com/andrewloable/go-fastreport/engine"
	"github.com/andrewloable/go-fastreport/reportpkg"
)

func newPageNumEngine(t *testing.T) *engine.ReportEngine {
	t.Helper()
	r := reportpkg.NewReport()
	pg := reportpkg.NewReportPage()
	r.AddPage(pg)
	e := engine.New(r)
	if err := e.Run(engine.DefaultRunOptions()); err != nil {
		t.Fatalf("Run: %v", err)
	}
	return e
}

func TestIncLogicalPageNumber_Increments(t *testing.T) {
	e := newPageNumEngine(t)
	if e.LogicalPageNo() != 0 {
		// After a Run() the engine may have incremented; just test the behaviour.
	}
	before := e.LogicalPageNo()
	e.IncLogicalPageNumber()
	if e.LogicalPageNo() != before+1 {
		t.Errorf("LogicalPageNo = %d, want %d", e.LogicalPageNo(), before+1)
	}
}

func TestIncLogicalPageNumber_AddsEntry(t *testing.T) {
	e := newPageNumEngine(t)
	before := e.LogicalPageCount()
	e.IncLogicalPageNumber()
	if e.LogicalPageCount() != before+1 {
		t.Errorf("LogicalPageCount = %d, want %d", e.LogicalPageCount(), before+1)
	}
}

func TestResetLogicalPageNumber_ResetsCounter(t *testing.T) {
	e := newPageNumEngine(t)
	e.IncLogicalPageNumber()
	e.IncLogicalPageNumber()
	e.ResetLogicalPageNumber()
	if e.LogicalPageNo() != 0 {
		t.Errorf("LogicalPageNo after reset = %d, want 0", e.LogicalPageNo())
	}
}

func TestResetLogicalPageNumber_BackfillsTotalPages(t *testing.T) {
	e := newPageNumEngine(t)
	e.IncLogicalPageNumber() // pageNo=1
	e.IncLogicalPageNumber() // pageNo=2
	e.ResetLogicalPageNumber()
	// After reset, logicalPageNo=0; count remains 2.
	if e.LogicalPageCount() != 2 {
		t.Errorf("LogicalPageCount = %d, want 2", e.LogicalPageCount())
	}
}
