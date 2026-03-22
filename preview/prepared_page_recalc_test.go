package preview

import (
	"testing"
)

// TestReCalcSizes verifies that ReCalcSizes sets Width and Height to the
// maximum right-edge and bottom-edge of all bands on the page.
func TestReCalcSizes(t *testing.T) {
	p := &PreparedPage{PageNo: 1}
	p.Bands = []*PreparedBand{
		{Left: 10, Top: 5, Width: 100, Height: 30},  // right=110, bottom=35
		{Left: 20, Top: 40, Width: 200, Height: 50}, // right=220, bottom=90
	}

	p.ReCalcSizes()

	if p.Width != 220 {
		t.Errorf("ReCalcSizes: Width = %v, want 220", p.Width)
	}
	if p.Height != 90 {
		t.Errorf("ReCalcSizes: Height = %v, want 90", p.Height)
	}
}

// TestReCalcSizes_NoBands verifies that ReCalcSizes with no bands produces
// zero Width and Height.
func TestReCalcSizes_NoBands(t *testing.T) {
	p := &PreparedPage{PageNo: 1, Width: 500, Height: 700}
	p.ReCalcSizes()
	if p.Width != 0 {
		t.Errorf("ReCalcSizes (no bands): Width = %v, want 0", p.Width)
	}
	if p.Height != 0 {
		t.Errorf("ReCalcSizes (no bands): Height = %v, want 0", p.Height)
	}
}

// TestMirrorMargins_even verifies that bands on an even-numbered page are
// shifted right by (rightMargin - leftMargin).
func TestMirrorMargins_even(t *testing.T) {
	leftMargin := float32(20)
	rightMargin := float32(30)
	expectedShift := rightMargin - leftMargin // 10

	p := &PreparedPage{PageNo: 2} // even page
	p.Bands = []*PreparedBand{
		{Left: 50},
		{Left: 100},
	}

	p.MirrorMargins(leftMargin, rightMargin)

	if p.Bands[0].Left != 50+expectedShift {
		t.Errorf("MirrorMargins even: band[0].Left = %v, want %v", p.Bands[0].Left, 50+expectedShift)
	}
	if p.Bands[1].Left != 100+expectedShift {
		t.Errorf("MirrorMargins even: band[1].Left = %v, want %v", p.Bands[1].Left, 100+expectedShift)
	}
}

// TestMirrorMargins_odd verifies that bands on an odd-numbered page are not
// shifted.
func TestMirrorMargins_odd(t *testing.T) {
	p := &PreparedPage{PageNo: 1} // odd page
	p.Bands = []*PreparedBand{
		{Left: 50},
		{Left: 100},
	}

	p.MirrorMargins(20, 30)

	if p.Bands[0].Left != 50 {
		t.Errorf("MirrorMargins odd: band[0].Left = %v, want 50 (unchanged)", p.Bands[0].Left)
	}
	if p.Bands[1].Left != 100 {
		t.Errorf("MirrorMargins odd: band[1].Left = %v, want 100 (unchanged)", p.Bands[1].Left)
	}
}

// TestMirrorMargins_negativeShift verifies that when leftMargin > rightMargin
// the bands are shifted left (negative offset) on even pages.
func TestMirrorMargins_negativeShift(t *testing.T) {
	p := &PreparedPage{PageNo: 4} // even page
	p.Bands = []*PreparedBand{
		{Left: 80},
	}

	// leftMargin 40, rightMargin 10 → shift = 10-40 = -30
	p.MirrorMargins(40, 10)

	if p.Bands[0].Left != 80-30 {
		t.Errorf("MirrorMargins negativeShift: band[0].Left = %v, want 50", p.Bands[0].Left)
	}
}
