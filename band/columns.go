package band

import "fmt"

// ColumnLayout controls the order in which multi-column data bands are filled.
type ColumnLayout int

const (
	// ColumnLayoutAcrossThenDown fills columns left-to-right, then wraps down.
	ColumnLayoutAcrossThenDown ColumnLayout = iota
	// ColumnLayoutDownThenAcross fills each column top-to-bottom before moving right.
	ColumnLayoutDownThenAcross
)

// BandColumns holds the multi-column layout settings for a DataBand.
// It is the Go equivalent of FastReport.BandColumns.
type BandColumns struct {
	// count is the number of columns. 0 or 1 means single-column mode.
	count int
	// Width is the column width in pixels. 0 means auto (split page width evenly).
	Width float32
	// Layout controls column fill order.
	Layout ColumnLayout
	// MinRowCount is the minimum number of rows per column (DownThenAcross only).
	// 0 means auto-calculate.
	MinRowCount int

	// pageWidthFn is an optional callback that returns the usable page width in
	// pixels (page width minus margins). Set by DataBand after construction.
	pageWidthFn func() float32
}

// NewBandColumns creates a BandColumns with defaults (count=0, AcrossThenDown).
func NewBandColumns() *BandColumns {
	return &BandColumns{}
}

// Count returns the number of columns. Values ≤ 1 mean single-column mode.
func (bc *BandColumns) Count() int { return bc.count }

// SetCount sets the column count. Returns an error when count < 0.
func (bc *BandColumns) SetCount(n int) error {
	if n < 0 {
		return fmt.Errorf("BandColumns.SetCount: count must be >= 0, got %d", n)
	}
	bc.count = n
	return nil
}

// ActualWidth returns the effective column width in pixels.
// When Width is 0, the page width (supplied via pageWidthFn) is divided evenly.
func (bc *BandColumns) ActualWidth() float32 {
	if bc.Width != 0 {
		return bc.Width
	}
	if bc.pageWidthFn != nil {
		pageW := bc.pageWidthFn()
		n := bc.count
		if n <= 1 {
			n = 1
		}
		return pageW / float32(n)
	}
	return 0
}

// Positions returns the left-edge x-offset of each column in pixels.
func (bc *BandColumns) Positions() []float32 {
	n := bc.count
	if n <= 0 {
		return nil
	}
	colW := bc.ActualWidth()
	pos := make([]float32, n)
	for i := range pos {
		pos[i] = float32(i) * colW
	}
	return pos
}

// Assign copies settings from another BandColumns.
func (bc *BandColumns) Assign(src *BandColumns) {
	bc.count = src.count
	bc.Width = src.Width
	bc.Layout = src.Layout
	bc.MinRowCount = src.MinRowCount
}
