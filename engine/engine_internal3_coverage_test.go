package engine

// engine_internal3_coverage_test.go — internal tests (package engine) targeting
// unexported functions that cannot be reached from external test packages:
//
//   keep.go startKeepBand:  `if b != nil && b.AbsRowNo() == 1 && !b.StartNewPage() { return }`

import (
	"testing"

	"github.com/andrewloable/go-fastreport/band"
	"github.com/andrewloable/go-fastreport/reportpkg"
)

// ── startKeepBand: b != nil, AbsRowNo==1, !StartNewPage → early return ────────

// TestStartKeepBand_NonNilBand_FirstRow exercises the guard:
//   `if b != nil && b.AbsRowNo() == 1 && !b.StartNewPage() { return }`
// startKeepBand is only called via StartKeep (which passes nil), so this
// branch is unreachable through the public API.  We call it directly.
func TestStartKeepBand_NonNilBand_FirstRow(t *testing.T) {
	r := reportpkg.NewReport()
	pg := reportpkg.NewReportPage()
	r.AddPage(pg)
	e := New(r)
	if err := e.Run(DefaultRunOptions()); err != nil {
		t.Fatalf("Run: %v", err)
	}

	db := band.NewDataBand()
	db.SetAbsRowNo(1)          // first absolute row
	db.SetStartNewPage(false)  // !StartNewPage → guard fires

	// Keeping must be false before the call.
	if e.keeping {
		t.Fatal("expected keeping=false before test")
	}

	// Call the unexported function directly.
	// Because AbsRowNo()==1 and !StartNewPage(), keeping should remain false.
	e.startKeepBand(&db.BandBase)

	if e.keeping {
		t.Error("startKeepBand with AbsRowNo==1 and !StartNewPage should NOT set keeping=true")
	}
}

// TestStartKeepBand_AlreadyKeeping exercises the `if e.keeping { return }` guard.
// This is covered by TestStartKeep_Idempotent but we add an explicit internal
// test for clarity and to ensure direct startKeepBand coverage.
func TestStartKeepBand_AlreadyKeeping(t *testing.T) {
	r := reportpkg.NewReport()
	pg := reportpkg.NewReportPage()
	r.AddPage(pg)
	e := New(r)
	if err := e.Run(DefaultRunOptions()); err != nil {
		t.Fatalf("Run: %v", err)
	}

	// Set keeping=true directly.
	e.keeping = true
	savedY := e.keepCurY

	// Call again — should return immediately without updating keepCurY.
	e.AdvanceY(50)
	e.startKeepBand(nil)

	if e.keepCurY != savedY {
		t.Errorf("startKeepBand while keeping: keepCurY changed from %v to %v", savedY, e.keepCurY)
	}
}

