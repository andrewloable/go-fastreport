package reportpkg_test

// frx_serialization_fixes_test.go — Integration tests verifying that
// serialization bugs fixed in this project produce correct output when
// loading real .frx files from test-reports/.
//
// Bugs verified:
//   1. LineObject StartCap/EndCap dot-qualified attribute format
//   2. GroupHeaderBand.SortOrder string deserialization
//   3. Format UseLocale attribute name round-trip
//   4. PDF /Info dictionary presence

import (
	"bytes"
	"strings"
	"testing"

	"github.com/andrewloable/go-fastreport/band"
	"github.com/andrewloable/go-fastreport/export/pdf"
	"github.com/andrewloable/go-fastreport/format"
	"github.com/andrewloable/go-fastreport/object"
	"github.com/andrewloable/go-fastreport/preview"
	"github.com/andrewloable/go-fastreport/report"
	"github.com/andrewloable/go-fastreport/reportpkg"
)

// ── Bug 1: LineObject StartCap/EndCap dot-qualified attribute names ─────────────
//
// Old code serialised caps as "StartCap=8,8,4" (CSV) instead of
// "StartCap.Width=8 StartCap.Height=8 StartCap.Style=Arrow" (dot-qualified).
// Loading the C#-generated FRX file verifies the fix.

func TestFRXIntegration_LinesAndShapes_CapStylesRead(t *testing.T) {
	r := loadFRXSmoke(t, "Lines and Shapes.frx")

	type hasObjects interface {
		Objects() *report.ObjectCollection
	}

	var lines []*object.LineObject
	for _, pg := range r.Pages() {
		for _, b := range pg.AllBands() {
			if h, ok := b.(hasObjects); ok {
				objs := h.Objects()
				for i := 0; i < objs.Len(); i++ {
					if lo, ok := objs.Get(i).(*object.LineObject); ok {
						lines = append(lines, lo)
					}
				}
			}
		}
	}
	if len(lines) == 0 {
		t.Fatal("expected LineObjects in Lines and Shapes.frx")
	}

	// The FRX contains at least one LineObject with StartCap.Style="Arrow".
	// Verify at least one LineObject has a non-None StartCap or EndCap style.
	var foundNonNoneCap bool
	for _, lo := range lines {
		if lo.StartCap.Style != object.CapStyleNone || lo.EndCap.Style != object.CapStyleNone {
			foundNonNoneCap = true
			break
		}
	}
	if !foundNonNoneCap {
		t.Error("expected at least one LineObject with non-None cap style; caps may not be deserializing from dot-qualified attributes")
	}
}

// TestFRXIntegration_LinesAndShapes_CapStyleRoundTrip verifies that a round-trip
// (load → save → load) preserves LineObject cap settings.
func TestFRXIntegration_LinesAndShapes_CapStyleRoundTrip(t *testing.T) {
	r := loadFRXSmoke(t, "Lines and Shapes.frx")
	r2 := roundTripFRX(t, r)

	type hasObjects interface {
		Objects() *report.ObjectCollection
	}

	// Find all line objects in r2 with non-None caps.
	var foundNonNoneCap bool
	for _, pg := range r2.Pages() {
		for _, b := range pg.AllBands() {
			if h, ok := b.(hasObjects); ok {
				objs := h.Objects()
				for i := 0; i < objs.Len(); i++ {
					if lo, ok := objs.Get(i).(*object.LineObject); ok {
						if lo.StartCap.Style != object.CapStyleNone || lo.EndCap.Style != object.CapStyleNone {
							foundNonNoneCap = true
						}
					}
				}
			}
		}
	}
	if !foundNonNoneCap {
		t.Error("cap styles not preserved after round-trip; dot-qualified serialization may be broken")
	}
}

// ── Bug 2: GroupHeaderBand.SortOrder string deserialization ─────────────────
//
// Old code wrote SortOrder as integer (e.g. 2) instead of string name ("None").
// Loading Groups.frx verifies that the C#-generated SortOrder="None" is
// correctly read as SortOrderNone.

func TestFRXIntegration_Groups_SortOrderNone(t *testing.T) {
	r := loadFRXSmoke(t, "Groups.frx")

	var ghBands []*band.GroupHeaderBand
	for _, pg := range r.Pages() {
		for _, b := range pg.Bands() {
			if gh, ok := b.(*band.GroupHeaderBand); ok {
				ghBands = append(ghBands, gh)
			}
		}
	}
	if len(ghBands) == 0 {
		t.Fatal("expected at least one GroupHeaderBand in Groups.frx")
	}
	// Groups.frx: <GroupHeaderBand ... SortOrder="None" ...>
	found := false
	for _, gh := range ghBands {
		if gh.SortOrder() == band.SortOrderNone {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected GroupHeaderBand.SortOrder() == SortOrderNone; got %v", ghBands[0].SortOrder())
	}
}

// ── Bug 3: Format.UseLocale attribute name round-trip ───────────────────────
//
// CurrencyFormat with UseLocaleSettings=true is loaded from the C#-generated FRX
// and must round-trip correctly. The attribute name is "Format.UseLocale"
// (not "Format.UseLocaleSettings").
//
// Groups.frx has Text4 inside Data1 (DataBand nested in GroupHeaderBand1).
// AllBands() only returns top-level page bands, so we use collectAllBands
// to recurse into GroupHeaderBand.Data().

// collectAllBands recursively gathers bands including nested DataBands inside
// GroupHeaderBand. This is needed because AllBands() only returns top-level bands.
func collectAllBands(top []report.Base) []report.Base {
	var result []report.Base
	for _, b := range top {
		result = append(result, b)
		if gh, ok := b.(*band.GroupHeaderBand); ok {
			if gh.Data() != nil {
				result = append(result, collectAllBands([]report.Base{gh.Data()})...)
			}
			if gh.NestedGroup() != nil {
				result = append(result, collectAllBands([]report.Base{gh.NestedGroup()})...)
			}
		}
	}
	return result
}

// hasCurrencyWithLocale searches all bands (recursively) for a TextObject
// with CurrencyFormat.UseLocaleSettings=true.
func hasCurrencyWithLocale(r *reportpkg.Report) bool {
	type hasObjects interface {
		Objects() *report.ObjectCollection
	}
	for _, pg := range r.Pages() {
		for _, b := range collectAllBands(pg.AllBands()) {
			if h, ok := b.(hasObjects); ok {
				objs := h.Objects()
				for i := 0; i < objs.Len(); i++ {
					if to, ok := objs.Get(i).(*object.TextObject); ok {
						if cf, ok := to.Format().(*format.CurrencyFormat); ok {
							if cf.UseLocaleSettings {
								return true
							}
						}
					}
				}
			}
		}
	}
	return false
}

func TestFRXIntegration_Groups_CurrencyFormatUseLocale(t *testing.T) {
	// Groups.frx: Text4 has Format="Currency" Format.UseLocale="true"
	// inside Data1 (DataBand nested in GroupHeaderBand1).
	r := loadFRXSmoke(t, "Groups.frx")
	if !hasCurrencyWithLocale(r) {
		t.Error("expected at least one TextObject with CurrencyFormat.UseLocaleSettings=true in Groups.frx")
	}
}

// TestFRXIntegration_Groups_FormatRoundTrip verifies CurrencyFormat.UseLocale
// survives a save/load round-trip.
func TestFRXIntegration_Groups_FormatRoundTrip(t *testing.T) {
	r := loadFRXSmoke(t, "Groups.frx")
	r2 := roundTripFRX(t, r)
	if !hasCurrencyWithLocale(r2) {
		t.Error("CurrencyFormat.UseLocaleSettings not preserved after round-trip in Groups.frx")
	}
}

// ── Bug 4: PDF /Info dictionary ──────────────────────────────────────────────
//
// Verify that the PDF export writes an /Info dictionary entry and no '#20'
// encoding artifacts in reference tokens. This is tested via PreparedPages
// constructed directly (avoiding data source requirements).

func TestFRXIntegration_PDF_InfoDictionary(t *testing.T) {
	pp := preview.New()
	pp.AddPage(595, 842, 1)
	b := &preview.PreparedBand{
		Name:   "ReportTitle",
		Top:    0,
		Height: 30,
	}
	b.Objects = []preview.PreparedObject{
		{
			Name:   "TitleText",
			Kind:   preview.ObjectTypeText,
			Left:   0,
			Top:    0,
			Width:  400,
			Height: 25,
			Text:   "Integration Test Report",
		},
	}
	_ = pp.AddBand(b)

	exp := pdf.NewExporter()
	exp.Author = "Integration Test"
	exp.Title = "Serialization Fix Verification"
	var buf bytes.Buffer
	if err := exp.Export(pp, &buf); err != nil {
		t.Fatalf("PDF export failed: %v", err)
	}

	pdfOut := buf.String()
	// /Info must appear in PDF trailer.
	if !strings.Contains(pdfOut, "/Info") {
		t.Error("PDF output does not contain /Info dictionary reference")
	}
	// No '#20' URL-encoded whitespace in object reference tokens (regression for
	// the indirect reference encoding bug fixed in go-fastreport-qtecl).
	if strings.Contains(pdfOut, "#20") {
		t.Error("PDF output contains '#20' artifacts in reference tokens (indirect ref encoding regression)")
	}
	// Verify basic PDF structure.
	if !strings.HasPrefix(pdfOut, "%PDF-") {
		t.Error("PDF output does not start with %PDF- header")
	}
}
