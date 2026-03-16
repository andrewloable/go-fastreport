package reportpkg

// loadsave_coverage2_test.go — internal tests to cover remaining branches in
// loadsave.go that are otherwise unreachable via the public API.
//
//   - deserializePage: obj.Deserialize error path (line 158-161)
//   - loadFromSerialReader: deserializePage error propagation (line 121-123)
//   - SaveToString: surface the happy-path through the bytes.Buffer route,
//     confirming the SaveTo(&buf) call is exercised for compressed output too.

import (
	"fmt"
	"strings"
	"testing"

	"github.com/andrewloable/go-fastreport/report"
	"github.com/andrewloable/go-fastreport/serial"
)

// ── alwaysFailWriter — an io.Writer that always returns an error ─────────────

// alwaysFailWriter satisfies io.Writer but always returns an error on Write.
// Used to force SaveTo to fail so that the error-propagation path is covered.
type alwaysFailWriter struct{}

func (e *alwaysFailWriter) Write(p []byte) (n int, err error) {
	return 0, fmt.Errorf("alwaysFailWriter: intentional write failure")
}

// TestSaveToString_SaveToFails_Internal exercises the SaveTo error path
// by calling SaveTo directly with a writer that always fails.
func TestSaveToString_SaveToFails_Internal(t *testing.T) {
	r := NewReport()
	r.Info.Name = "FailReport"

	err := r.SaveTo(&alwaysFailWriter{})
	if err == nil {
		t.Fatal("SaveTo with alwaysFailWriter should return an error")
	}
}

// ── failingDeserializeObject — a report.Base whose Deserialize always errors ──

// failingBase is a minimal report.Base implementation whose Deserialize method
// always returns an error, allowing us to exercise the error-return branch in
// deserializePage (loadsave.go:158-161) and the propagation in loadFromSerialReader
// (loadsave.go:121-123).
type failingBase struct {
	report.BaseObject
}

func (f *failingBase) TypeName() string { return "FailingDeserializeBand" }

func (f *failingBase) Serialize(w report.Writer) error { return nil }

func (f *failingBase) Deserialize(r report.Reader) error {
	return fmt.Errorf("failingBase: intentional deserialize error")
}

// registerFailingType registers "FailingDeserializeBand" in the global registry.
// It is safe to call more than once because the registry silently ignores
// double-registration (the init() in serial_registrations.go already does this).
func registerFailingType() {
	_ = serial.DefaultRegistry.Register("FailingDeserializeBand", func() report.Base {
		return &failingBase{}
	})
}

// ── deserializePage — obj.Deserialize error path (line 158-161) ──────────────

// TestDeserializePage_DeserializeError_Internal triggers the
// `if err2 := obj.Deserialize(rdr); err2 != nil` branch in deserializePage
// by including a registered band type whose Deserialize always fails.
func TestDeserializePage_DeserializeError_Internal(t *testing.T) {
	registerFailingType()

	frx := `<?xml version="1.0" encoding="utf-8"?><Report>
		<ReportPage Name="Page1">
			<FailingDeserializeBand Name="Bad"/>
		</ReportPage>
	</Report>`

	r := NewReport()
	err := r.LoadFromString(frx)
	if err == nil {
		t.Fatal("LoadFromString should return error when band Deserialize fails")
	}
	if !strings.Contains(err.Error(), "intentional deserialize error") &&
		!strings.Contains(err.Error(), "deserialize page") {
		t.Errorf("unexpected error message: %v", err)
	}
}

// TestLoadFromSerialReader_DeserializePageError_Internal verifies that the
// error from deserializePage is propagated through loadFromSerialReader
// (loadsave.go:121-123).
func TestLoadFromSerialReader_DeserializePageError_Internal(t *testing.T) {
	registerFailingType()

	// Build an XML document where the first (and only) ReportPage contains
	// a FailingDeserializeBand so that deserializePage returns an error,
	// which loadFromSerialReader then propagates.
	frx := `<?xml version="1.0" encoding="utf-8"?><Report ReportName="ErrReport">
		<ReportPage Name="ErrPage">
			<FailingDeserializeBand Name="Trigger"/>
		</ReportPage>
	</Report>`

	r := NewReport()
	rdr := serial.NewReader(strings.NewReader(frx))
	err := r.loadFromSerialReader(rdr)
	if err == nil {
		t.Fatal("loadFromSerialReader should return error when deserializePage fails")
	}
	// Confirm the error message contains the expected context.
	if !strings.Contains(err.Error(), "deserialize page") {
		t.Errorf("error should mention 'deserialize page', got: %v", err)
	}
}

// ── SaveToString — compressed round-trip (exercise the SaveTo(&buf) path) ──

// TestSaveToString_CompressedRoundTrip_Internal exercises the SaveToString path
// when r.Compressed = true so that SaveTo compresses into the bytes.Buffer.
// This adds a distinct execution path through SaveToString for the compressed case.
func TestSaveToString_CompressedRoundTrip_Internal(t *testing.T) {
	r := NewReport()
	r.Info.Name = "SavToStrComp"
	r.Compressed = true
	pg := NewReportPage()
	pg.SetName("P1")
	r.AddPage(pg)

	s, err := r.SaveToString()
	if err != nil {
		t.Fatalf("SaveToString compressed unexpected error: %v", err)
	}
	if s == "" {
		t.Error("SaveToString compressed returned empty string")
	}

	// Round-trip: load back and verify.
	r2 := NewReport()
	if err2 := r2.LoadFromString(s); err2 != nil {
		t.Fatalf("LoadFromString after SaveToString compressed: %v", err2)
	}
	if r2.Info.Name != "SavToStrComp" {
		t.Errorf("round-trip name mismatch: got %q", r2.Info.Name)
	}
}

// TestSaveToString_UncompressedRoundTrip_Internal exercises the primary happy
// path of SaveToString (r.Compressed = false) to confirm the buf.String() return
// path is covered.
func TestSaveToString_UncompressedRoundTrip_Internal(t *testing.T) {
	r := NewReport()
	r.Info.Name = "SavToStrPlain"
	r.Compressed = false

	s, err := r.SaveToString()
	if err != nil {
		t.Fatalf("SaveToString uncompressed unexpected error: %v", err)
	}
	if !strings.Contains(s, "SavToStrPlain") {
		t.Errorf("SaveToString output missing report name, got: %q", s[:min2(len(s), 200)])
	}
}

// min2 returns the smaller of two ints (local helper to avoid importing math).
func min2(a, b int) int {
	if a < b {
		return a
	}
	return b
}
