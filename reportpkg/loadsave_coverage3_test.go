package reportpkg

// loadsave_coverage3_test.go — additional internal tests to cover the remaining
// uncovered branches in loadsave.go.
//
// Target: SaveToString (line 651) error return path at 75.0% → 100%
//
// Strategy: SaveToString calls r.SaveTo(&buf) where buf is a bytes.Buffer.
// bytes.Buffer.Write never returns an error, so the only way for SaveTo to fail
// is for the XML serialization itself to return an error. We accomplish this by
// adding a band whose Serialize method returns a non-nil error. When
// Report.Serialize iterates pages and calls serializeBands, which calls
// w.WriteObject(failBand), the serial.Writer propagates the Serialize error
// through WriteObjectNamed → SaveTo → SaveToString, hitting the "return '', err"
// branch on line 654.

import (
	"fmt"
	"strings"
	"testing"

	"github.com/andrewloable/go-fastreport/report"
)

// ── failSerializeBand — a band whose Serialize always returns an error ────────

// failSerializeBand is a minimal band implementation whose Serialize always
// returns an error, allowing us to drive an error through the SaveToString
// → SaveTo → serial.Writer.WriteObjectNamed → Report.Serialize chain.
type failSerializeBand struct {
	report.BaseObject
}

func (f *failSerializeBand) TypeName() string { return "FailSerializeBand" }

func (f *failSerializeBand) Serialize(w report.Writer) error {
	return fmt.Errorf("failSerializeBand: intentional serialize error")
}

func (f *failSerializeBand) Deserialize(r report.Reader) error { return nil }

// ── TestSaveToString_SaveToInternalError_Internal ─────────────────────────────

// TestSaveToString_SaveToInternalError_Internal exercises the error-return path
// of SaveToString (loadsave.go:653-655):
//
//	if err := r.SaveTo(&buf); err != nil {
//	    return "", err           // ← previously uncovered
//	}
//
// The test creates a Report with a page that contains a band whose Serialize
// method returns a non-nil error. When SaveToString calls SaveTo → serial.Writer
// → Report.Serialize → page.Serialize → serializeBands → WriteObject(failBand)
// → failBand.Serialize(), the error propagates all the way back and SaveToString
// returns it rather than the empty-string success value.
func TestSaveToString_SaveToInternalError_Internal(t *testing.T) {
	r := NewReport()
	pg := NewReportPage()
	pg.SetName("FailPage")

	// Inject the failing band directly into the page's ordered band list.
	fb := &failSerializeBand{}
	fb.SetName("FailBand")
	pg.AddBand(fb)

	r.AddPage(pg)

	s, err := r.SaveToString()
	if err == nil {
		t.Fatal("SaveToString should return an error when a band's Serialize fails, got nil")
	}
	if s != "" {
		t.Errorf("SaveToString should return empty string on error, got: %q", s[:min3(len(s), 80)])
	}
	if !strings.Contains(err.Error(), "intentional serialize error") &&
		!strings.Contains(err.Error(), "write report") {
		t.Errorf("unexpected error message: %v", err)
	}
}

// min3 is a local helper to avoid importing math or depending on Go 1.21's min builtin.
func min3(a, b int) int {
	if a < b {
		return a
	}
	return b
}
