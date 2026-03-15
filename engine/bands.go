package engine

import (
	"image/color"

	"github.com/andrewloable/go-fastreport/band"
	"github.com/andrewloable/go-fastreport/object"
	"github.com/andrewloable/go-fastreport/preview"
	"github.com/andrewloable/go-fastreport/report"
	"github.com/andrewloable/go-fastreport/utils"
)

// defaultTextColor is used when a TextObject does not expose a text color.
var defaultTextColor = color.RGBA{A: 255}

// ── CanPrint ──────────────────────────────────────────────────────────────────

// CanPrint returns true if obj should be rendered on the current page.
// It evaluates the PrintOn bitmask against the current page position.
// pageIndex is the zero-based index of the current page in PreparedPages.
// totalPages is the total number of prepared pages.
//
// Each bit in PrintOn is treated as an independent "allow" condition;
// the component prints if ANY enabled condition matches the current page.
// PrintOnAllPages (0) always prints.
func (e *ReportEngine) CanPrint(obj *report.ReportComponentBase, pageIndex, totalPages int) bool {
	if !obj.Visible() {
		return false
	}

	printOn := obj.PrintOn()

	// PrintOnAllPages (zero value) — always print.
	if printOn == report.PrintOnAllPages {
		return true
	}

	isFirstPage := pageIndex == 0
	isLastPage := pageIndex == totalPages-1
	isSinglePage := isFirstPage && isLastPage
	pageNumber := pageIndex + 1 // 1-based page number for odd/even check

	canPrint := false
	if (printOn&report.PrintOnOddPages) != 0 && pageNumber%2 == 1 {
		canPrint = true
	}
	if (printOn&report.PrintOnEvenPages) != 0 && pageNumber%2 == 0 {
		canPrint = true
	}
	if (printOn&report.PrintOnFirstPage) != 0 && isFirstPage {
		canPrint = true
	}
	if (printOn&report.PrintOnLastPage) != 0 && isLastPage {
		canPrint = true
	}
	if (printOn&report.PrintOnSinglePage) != 0 && isSinglePage {
		canPrint = true
	}
	return canPrint
}

// ── CalcBandHeight ────────────────────────────────────────────────────────────

// CalcBandHeight returns the rendered height of a band in pixels.
// If the band has CanGrow or CanShrink set, child TextObjects are measured
// and the height is adjusted to fit their content.
func (e *ReportEngine) CalcBandHeight(b report.Base) float32 {
	bb, ok := b.(*band.BandBase)
	if !ok {
		type hasHeight interface{ Height() float32 }
		if h, ok2 := b.(hasHeight); ok2 {
			return h.Height()
		}
		return 0
	}

	baseHeight := bb.Height()
	canGrow := bb.CanGrow()
	canShrink := bb.CanShrink()

	if !canGrow && !canShrink {
		return baseHeight
	}

	// Measure TextObject children to compute the required height.
	requiredHeight := calcBandRequiredHeight(bb, baseHeight)

	if canGrow && requiredHeight > baseHeight {
		return requiredHeight
	}
	if canShrink && requiredHeight < baseHeight {
		return requiredHeight
	}
	return baseHeight
}

// calcBandRequiredHeight measures all TextObject children of the band and
// returns the minimum height needed to display all content without clipping.
// baseHeight is used as the starting lower bound.
func calcBandRequiredHeight(bb *band.BandBase, baseHeight float32) float32 {
	objs := bb.Objects()
	if objs == nil || objs.Len() == 0 {
		return baseHeight
	}

	// We need the maximum bottom edge of all expanded text objects.
	maxBottom := float32(0)

	for i := 0; i < objs.Len(); i++ {
		obj := objs.Get(i)
		txt, ok := obj.(*object.TextObject)
		if !ok {
			// Non-text objects use their declared bottom edge.
			type hasDims interface {
				Top() float32
				Height() float32
			}
			if d, ok2 := obj.(hasDims); ok2 {
				bottom := d.Top() + d.Height()
				if bottom > maxBottom {
					maxBottom = bottom
				}
			}
			continue
		}

		// Measure the text content.
		objWidth := txt.Width()
		if objWidth <= 0 {
			objWidth = bb.Width()
		}

		var textH float32
		switch txt.TextRenderType() {
		case object.TextRenderTypeHtmlTags, object.TextRenderTypeHtmlParagraph:
			// Use HtmlTextRenderer for HTML-formatted text.
			renderer := utils.NewHtmlTextRenderer(txt.Text(), txt.Font(), defaultTextColor)
			textH = renderer.MeasureHeight(objWidth)
		default:
			_, textH = utils.MeasureText(txt.Text(), txt.Font(), objWidth)
		}

		declaredH := txt.Height()
		usedH := textH
		if usedH < declaredH {
			usedH = declaredH
		}

		bottom := txt.Top() + usedH
		if bottom > maxBottom {
			maxBottom = bottom
		}
	}

	if maxBottom < baseHeight {
		return baseHeight
	}
	return maxBottom
}

// ── AddBandToPreparedPages ────────────────────────────────────────────────────

// AddBandToPreparedPages adds b to the current PreparedPage.
// If there is not enough free space and the band cannot break, it starts a new
// page first. Returns true if the band was added, false if it was skipped.
//
// b may be any *band.BandBase (or an embedding struct cast to *band.BandBase).
func (e *ReportEngine) AddBandToPreparedPages(b *band.BandBase) bool {
	if e.preparedPages == nil {
		return false
	}

	height := e.CalcBandHeight(b)
	if height <= 0 {
		return false
	}

	// Check free space for regular bands (not page service bands).
	if b.FlagCheckFreeSpace && e.freeSpace < height {
		if b.CanBreak() || b.FlagMustBreak {
			// Band can break — break it (simplified: just start new page for now).
			e.startNewPageForCurrent()
		} else {
			// No break: start a new page and retry.
			e.startNewPageForCurrent()
			b.FlagMustBreak = true
			result := e.AddBandToPreparedPages(b)
			b.FlagMustBreak = false
			return result
		}
	}

	pb := &preview.PreparedBand{
		Name:   b.Name(),
		Top:    e.curY,
		Height: height,
	}
	e.populateBandObjects(b, pb)
	_ = e.preparedPages.AddBand(pb)
	e.AdvanceY(height)
	return true
}

// startNewPageForCurrent ends the current column (or page) and starts the
// next one. When the current page template has multiple columns and there is
// a next column available, the engine advances to that column rather than
// starting a new page.
func (e *ReportEngine) startNewPageForCurrent() {
	if e.currentPage == nil {
		return
	}
	// Cut any kept bands before breaking to a new column/page.
	e.CheckKeepTogether()

	// For multi-column layouts, try advancing to the next column first.
	if e.currentPage.Columns.Count > 1 {
		if e.endColumn(e.currentPage) {
			// Successfully advanced to the next column on the same page.
			e.startColumn(e.currentPage)
			e.FinishKeepTogether()
			return
		}
	}
	// No more columns (or single-column page): start a new page.
	e.endPage(e.currentPage, false)
	e.startPage(e.currentPage, false)
	e.FinishKeepTogether()
}

// ── ShowFullBand ──────────────────────────────────────────────────────────────

// ShowFullBand shows band b, repeating it RepeatBandNTimes times.
// It fires BeforeLayout/AfterLayout events and recursively shows any ChildBand.
func (e *ReportEngine) ShowFullBand(b *band.BandBase) {
	if b == nil {
		return
	}
	n := b.RepeatBandNTimes()
	if n <= 0 {
		n = 1
	}
	for i := 0; i < n; i++ {
		e.showFullBandOnce(b)
	}
}

func (e *ReportEngine) showFullBandOnce(b *band.BandBase) {
	if !b.Visible() {
		return
	}

	b.FireBeforeLayout()

	// Populate outline entry if band has an OutlineExpression.
	addedOutline := false
	if expr := b.OutlineExpression(); expr != "" {
		text := e.evalText(expr)
		if text != "" {
			e.AddOutline(text)
			addedOutline = true
		}
	}

	height := e.CalcBandHeight(b)
	if height <= 0 {
		b.FireAfterLayout()
		if addedOutline {
			e.OutlineUp()
		}
		return
	}

	// Check free space.
	if b.FlagCheckFreeSpace && e.freeSpace < height {
		e.startNewPageForCurrent()
	}

	if e.preparedPages != nil {
		pb := &preview.PreparedBand{
			Name:   b.Name(),
			Top:    e.curY,
			Height: height,
		}
		e.populateBandObjects(b, pb)
		_ = e.preparedPages.AddBand(pb)
	}
	e.AdvanceY(height)
	b.FireAfterLayout()

	// Show child band if present.
	if child := b.Child(); child != nil {
		e.ShowFullBand(&child.BandBase)
	}

	if addedOutline {
		e.OutlineUp()
	}
}

// ShowDataBandRow shows a data band for a single data row, handling
// StartNewPage, row tracking, and child bands.
func (e *ReportEngine) ShowDataBandRow(db *band.DataBand, rowNo, absRowNo int) {
	db.SetRowNo(rowNo)
	db.SetAbsRowNo(absRowNo)

	// StartNewPage: skip on the first row (avoids empty first page).
	if db.StartNewPage() && db.FlagUseStartNewPage && rowNo != 1 {
		e.startNewPageForCurrent()
	}

	e.ShowFullBand(&db.BandBase)
}
