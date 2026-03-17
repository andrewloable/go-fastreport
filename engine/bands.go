package engine

import (
	"image/color"
	"strings"

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
	requiredHeight := calcBandLayout(bb, baseHeight, e.evalText).height

	if canGrow && requiredHeight > baseHeight {
		return requiredHeight
	}
	if canShrink && requiredHeight < baseHeight {
		return requiredHeight
	}
	return baseHeight
}

// bandLayout holds the results of a band layout calculation.
type bandLayout struct {
	// height is the required band height (may differ from declared height when
	// CanGrow or CanShrink is active).
	height float32
	// shifts[i] is the Y offset to add to object i's declared Top position due
	// to grow/shrink propagation from objects above it. Negative = shift up.
	// shifts is nil when no grow/shrink objects are present.
	shifts []float32
}

// calcBandLayout measures all TextObject children of the band and
// returns the required height together with per-object Y shifts.
//
// The algorithm mirrors FastReport.BandBase.CalcHeight() from the .NET source:
//  1. For each object that has CanGrow or CanShrink set, measure its content
//     height and decide the effective height (respecting individual object flags).
//  2. Compute downward shifts for objects that sit below a growing/shrinking peer.
//  3. Return the maximum bottom edge across all visible objects, plus the shifts.
func calcBandLayout(bb *band.BandBase, baseHeight float32, evalFn func(string) string) bandLayout {
	objs := bb.Objects()
	if objs == nil || objs.Len() == 0 {
		return bandLayout{height: baseHeight}
	}

	n := objs.Len()

	// hasDims provides the geometry of any component object.
	type hasDims interface {
		Top() float32
		Height() float32
	}
	type hasCanGrowShrink interface {
		CanGrow() bool
		CanShrink() bool
	}
	type hasVisible interface{ Visible() bool }

	// Step 1: compute effective height for each object.
	tops := make([]float32, n)
	effectiveH := make([]float32, n)
	hasGrowShrink := false
	for i := 0; i < n; i++ {
		obj := objs.Get(i)
		d, hasDim := obj.(hasDims)
		if !hasDim {
			continue
		}
		tops[i] = d.Top()
		effectiveH[i] = d.Height()

		txt, isTxt := obj.(*object.TextObject)
		if !isTxt {
			continue // non-text objects keep their declared height
		}

		// Check whether this TextObject individually permits grow/shrink.
		gs, hasGS := obj.(hasCanGrowShrink)
		canGrowObj := hasGS && gs.CanGrow()
		canShrinkObj := hasGS && gs.CanShrink()
		if !canGrowObj && !canShrinkObj {
			continue // object is fixed size
		}

		// Measure the rendered text height.
		objWidth := txt.Width()
		if objWidth <= 0 {
			objWidth = bb.Width()
		}
		// Account for left+right padding so measured text fits within the
		// actual content area.
		if p := txt.Padding(); p.Left+p.Right > 0 {
			objWidth -= p.Left + p.Right
			if objWidth < 1 {
				objWidth = 1
			}
		}

		// Evaluate the text so we measure the actual content, not the template.
		textToMeasure := txt.Text()
		if evalFn != nil {
			textToMeasure = evalFn(textToMeasure)
		}
		// Normalize line endings for consistent line counting.
		textToMeasure = strings.ReplaceAll(textToMeasure, "\r\n", "\n")
		textToMeasure = strings.ReplaceAll(textToMeasure, "\r", "\n")

		var textH float32
		switch txt.TextRenderType() {
		case object.TextRenderTypeHtmlTags, object.TextRenderTypeHtmlParagraph:
			renderer := utils.NewHtmlTextRenderer(textToMeasure, txt.Font(), defaultTextColor)
			textH = renderer.MeasureHeight(objWidth)
		default:
			_, textH = utils.MeasureText(textToMeasure, txt.Font(), objWidth)
		}
		// Add top+bottom padding back into the measured height.
		if p := txt.Padding(); p.Top+p.Bottom > 0 {
			textH += p.Top + p.Bottom
		}

		declaredH := txt.Height()
		if canGrowObj && textH > declaredH {
			effectiveH[i] = textH
			hasGrowShrink = true
		} else if canShrinkObj && textH < declaredH {
			effectiveH[i] = textH
			hasGrowShrink = true
		}
		// If neither condition applies, the declared height is unchanged.
	}

	// Step 2: propagate downward shifts for objects that sit below a
	// growing or shrinking peer. Object j is "below" object i when j's top
	// (after accumulated shifts) is at or below i's original bottom edge.
	shifts := make([]float32, n)
	if hasGrowShrink {
		for i := 0; i < n; i++ {
			obj := objs.Get(i)
			d, hasDim := obj.(hasDims)
			if !hasDim {
				continue
			}
			delta := effectiveH[i] - d.Height()
			if delta == 0 {
				continue
			}
			parentBottom := tops[i] + d.Height()
			for j := 0; j < n; j++ {
				if j == i {
					continue
				}
				if tops[j]+shifts[j] >= parentBottom-1e-4 {
					proposed := delta + shifts[i]
					if delta > 0 {
						if proposed > shifts[j] {
							shifts[j] = proposed
						}
					} else {
						if proposed < shifts[j] {
							shifts[j] = proposed
						}
					}
				}
			}
		}
	}

	// Step 3: compute max bottom edge across all visible objects.
	maxBottom := float32(0)
	for i := 0; i < n; i++ {
		obj := objs.Get(i)
		if v, ok := obj.(hasVisible); ok && !v.Visible() {
			continue
		}
		if _, hasDim := obj.(hasDims); !hasDim {
			continue
		}
		bottom := tops[i] + shifts[i] + effectiveH[i]
		if bottom > maxBottom {
			maxBottom = bottom
		}
	}

	if maxBottom <= 0 {
		return bandLayout{height: baseHeight, shifts: shifts}
	}
	return bandLayout{height: maxBottom, shifts: shifts}
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
	applyAnchorAdjustments(b, pb, 0, height-b.Height(), 0)
	if e.curX != 0 {
		for i := range pb.Objects {
			pb.Objects[i].Left += e.curX
		}
	}
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

	// Fire BeforePrint before any rendering takes place, mirroring the
	// C# ReportEngine.ShowBand which calls band.OnBeforePrint first.
	b.FireBeforePrint()
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
		b.FireAfterPrint()
		if addedOutline {
			e.OutlineUp()
		}
		return
	}

	// Check free space (skipped when rendering into a parent band via PrintOnParent).
	if e.outputBand == nil && b.FlagCheckFreeSpace && e.freeSpace < height {
		e.startNewPageForCurrent()
	}

	if e.outputBand != nil {
		// PrintOnParent mode: merge this subreport band's objects directly into
		// the parent band's PreparedBand, offset by the subreport position.
		tmp := &preview.PreparedBand{}
		e.populateBandObjects(b, tmp)
		yOff := e.outputBandOffsetY + e.curY
		for _, po := range tmp.Objects {
			po.Left += e.outputBandOffsetX
			po.Top += yOff
			e.outputBand.Objects = append(e.outputBand.Objects, po)
		}
		// Grow the parent band height to accommodate the subreport content.
		newBottom := yOff + height
		if newBottom > e.outputBand.Height {
			e.outputBand.Height = newBottom
		}
	} else if e.preparedPages != nil {
		pb := &preview.PreparedBand{
			Name:   b.Name(),
			Top:    e.curY,
			Height: height,
		}
		e.populateBandObjects(b, pb)
		applyAnchorAdjustments(b, pb, 0, height-b.Height(), 0)
		// Apply grow/shrink propagation shifts to PreparedObject Y positions.
		// When a CanShrink/CanGrow object changes size, objects below it shift
		// vertically. This mirrors FastReport's CalcHeight shift propagation.
		if layout := calcBandLayout(b, b.Height(), e.evalText); layout.shifts != nil {
			applyBandObjectShifts(b, pb, layout.shifts)
		}
		// Apply page-level column X offset so data bands in multi-column mode
		// are rendered in their respective column positions.
		if e.curX != 0 {
			for i := range pb.Objects {
				pb.Objects[i].Left += e.curX
			}
		}
		_ = e.preparedPages.AddBand(pb)
	}
	e.AdvanceY(height)
	b.FireAfterLayout()
	// Fire AfterPrint after the band has been added to the prepared pages,
	// mirroring the C# ReportEngine.ShowBand which calls band.OnAfterPrint last.
	b.FireAfterPrint()

	// Show child band if present.
	if child := b.Child(); child != nil {
		e.ShowFullBand(&child.BandBase)
	}

	if addedOutline {
		e.OutlineUp()
	}
}

// applyAnchorAdjustments adjusts the positions and sizes of PreparedObjects
// within pb according to the Anchor property of their corresponding band objects.
//
// When a band grows (or shrinks) due to CanGrow/CanShrink, objects anchored
// only to the bottom edge need their Y coordinate shifted down by the delta;
// objects anchored to both top and bottom need their Height increased by delta.
// The equivalent rules apply horizontally for left/right anchors using deltaW.
//
// startIdx is the index of the first PreparedObject in pb.Objects that belongs
// to this band (used when the band is one of several contributing to a page).
// deltaH is (effectiveHeight - declaredHeight) for the vertical axis.
// deltaW is (effectiveWidth  - declaredWidth)  for the horizontal axis (usually 0).
//
// The function iterates band objects in the same order as populateBandObjects2
// and matches each source object to its primary PreparedObject slot by index.
// Extra PreparedObjects added for container children, table cells, and
// AdvMatrix cells are skipped when iterating the source objects.
func applyAnchorAdjustments(bb *band.BandBase, pb *preview.PreparedBand, startIdx int, deltaH, deltaW float32) {
	if (deltaH == 0 && deltaW == 0) || bb == nil {
		return
	}
	objs := bb.Objects()
	if objs == nil || objs.Len() == 0 {
		return
	}

	// hasAnchor is satisfied by any object that exposes its Anchor property.
	type hasAnchor interface {
		Anchor() report.AnchorStyle
	}

	// poIdx tracks our position in pb.Objects. We start at startIdx and advance
	// by 1 per successfully built source object (buildPreparedObject returned
	// non-nil). Extra entries added for children/cells are skipped because we
	// only care about the primary slot for each source object.
	poIdx := startIdx
	for i := 0; i < objs.Len(); i++ {
		obj := objs.Get(i)
		if poIdx >= len(pb.Objects) {
			break
		}

		// Skip invisible objects — buildPreparedObject returned nil for these.
		type hasVisible interface{ Visible() bool }
		if v, ok := obj.(hasVisible); ok && !v.Visible() {
			continue
		}

		// Skip objects with no geometry — buildPreparedObject returned nil.
		type hasGeom interface {
			Left() float32
			Top() float32
			Width() float32
			Height() float32
		}
		if _, ok := obj.(hasGeom); !ok {
			continue
		}

		// poIdx now points to the primary PreparedObject for this source object.
		if anc, ok := obj.(hasAnchor); ok {
			a := anc.Anchor()
			po := &pb.Objects[poIdx]

			// Vertical anchor adjustments (deltaH).
			if deltaH != 0 {
				ancTop := (a & report.AnchorTop) != 0
				ancBottom := (a & report.AnchorBottom) != 0
				switch {
				case ancBottom && !ancTop:
					// Only bottom-anchored: move down by delta.
					po.Top += deltaH
				case ancTop && ancBottom:
					// Anchored to both: stretch vertically.
					po.Height += deltaH
				// AnchorTop only (default) or AnchorNone: no vertical change.
				}
			}

			// Horizontal anchor adjustments (deltaW).
			if deltaW != 0 {
				ancLeft := (a & report.AnchorLeft) != 0
				ancRight := (a & report.AnchorRight) != 0
				switch {
				case ancRight && !ancLeft:
					// Only right-anchored: move right by delta.
					po.Left += deltaW
				case ancLeft && ancRight:
					// Anchored to both: stretch horizontally.
					po.Width += deltaW
				// AnchorLeft only (default) or AnchorNone: no horizontal change.
				}
			}
		}
		poIdx++
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

// applyBandObjectShifts adjusts PreparedObject Top positions to account for
// vertical grow/shrink propagation from sibling objects in the same band.
//
// shifts[i] is the Y offset for source object i (computed by calcBandLayout).
// When object i grows or shrinks, all objects below it shift by the same delta.
// This mirrors FastReport's CalcHeight shift propagation for visual correctness.
//
// The function iterates source objects and PreparedObjects in parallel (same
// order as populateBandObjects2), applying shifts[i] to the primary
// PreparedObject for each source object.
func applyBandObjectShifts(bb *band.BandBase, pb *preview.PreparedBand, shifts []float32) {
	if len(shifts) == 0 {
		return
	}
	objs := bb.Objects()
	if objs == nil || objs.Len() == 0 {
		return
	}

	// Check if any shift is non-zero.
	hasNonZero := false
	for _, s := range shifts {
		if s != 0 {
			hasNonZero = true
			break
		}
	}
	if !hasNonZero {
		return
	}

	type hasVisible interface{ Visible() bool }
	type hasGeom interface {
		Top() float32
		Height() float32
	}

	poIdx := 0
	for i := 0; i < objs.Len() && i < len(shifts); i++ {
		obj := objs.Get(i)
		if poIdx >= len(pb.Objects) {
			break
		}

		// Skip invisible objects.
		if v, ok := obj.(hasVisible); ok && !v.Visible() {
			continue
		}

		// Skip objects with no geometry.
		if _, ok := obj.(hasGeom); !ok {
			continue
		}

		if shifts[i] != 0 {
			pb.Objects[poIdx].Top += shifts[i]
		}
		poIdx++
	}
}
