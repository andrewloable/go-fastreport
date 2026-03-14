package reportpkg_test

// Smoke tests for page-feature FRX reports (page breaks, headers, double-pass, etc.)

import (
	"testing"
)

func TestFRXSmoke_HandlePageBreaks(t *testing.T) {
	loadFRXSmoke(t, "Handle Page Breaks.frx")
}

func TestFRXSmoke_KeepTogether(t *testing.T) {
	loadFRXSmoke(t, "Keep Together.frx")
}

func TestFRXSmoke_RepeatHeaders(t *testing.T) {
	loadFRXSmoke(t, "Repeat Headers.frx")
}

func TestFRXSmoke_DoublePassTotalPages(t *testing.T) {
	r := loadFRXSmoke(t, "Double Pass, Total Pages.frx")
	if !r.DoublePass {
		t.Error("expected DoublePass=true in Double Pass, Total Pages.frx")
	}
}

func TestFRXSmoke_StartNewPageResetPageNumbers(t *testing.T) {
	loadFRXSmoke(t, "Start New Page, Reset Page Numbers.frx")
}

func TestFRXSmoke_ReportWithCoverPage(t *testing.T) {
	loadFRXSmoke(t, "Report With Cover Page.frx")
}

func TestFRXSmoke_PrintOnPreviousPage(t *testing.T) {
	loadFRXSmoke(t, "Print on Previous Page.frx")
}

func TestFRXSmoke_ShiftObjectPosition(t *testing.T) {
	loadFRXSmoke(t, "Shift Object Position.frx")
}

func TestFRXSmoke_Watermark(t *testing.T) {
	loadFRXSmoke(t, "Watermark.frx")
}

func TestFRXSmoke_Outline(t *testing.T) {
	loadFRXSmoke(t, "Outline.frx")
}
