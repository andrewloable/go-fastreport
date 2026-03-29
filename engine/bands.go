package engine

import (
	"image/color"
	"sort"
	"strings"

	"github.com/andrewloable/go-fastreport/band"
	"github.com/andrewloable/go-fastreport/matrix"
	"github.com/andrewloable/go-fastreport/object"
	"github.com/andrewloable/go-fastreport/preview"
	"github.com/andrewloable/go-fastreport/report"
	"github.com/andrewloable/go-fastreport/table"
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
	// Extract BandBase from any band type (ReportTitleBand, PageHeaderBand, etc.)
	bb := extractBandBase(b)
	if bb == nil {
		type hasHeight interface{ Height() float32 }
		if h, ok := b.(hasHeight); ok {
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
	// effectiveH[i] is the effective height for each object after CanGrow/CanShrink.
	// nil when no grow/shrink objects are present.
	effectiveH []float32
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

		// For horizontal LineObjects with Height=0, set effective height to the
		// border line width (default 1) so the line is visible. This matches
		// C# where LineObject.Height is used in rendering calculations.
		if line, isLine := obj.(*object.LineObject); isLine {
			if effectiveH[i] == 0 && d.(interface{ Width() float32 }).Width() > 0 && !line.Diagonal() {
				lw := float32(1)
				if line.Border().Lines[0] != nil && line.Border().Lines[0].Width > 0 {
					lw = line.Border().Lines[0].Width
				}
				effectiveH[i] = lw
			}
			continue
		}

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
		// Trim trailing newlines — C# MeasureString ignores trailing whitespace/newlines
		// so "text\n" measures the same as "text".
		textToMeasure = strings.TrimRight(textToMeasure, "\n")

		var textH float32
		switch txt.TextRenderType() {
		case object.TextRenderTypeHtmlTags, object.TextRenderTypeHtmlParagraph:
			renderer := utils.NewHtmlTextRenderer(textToMeasure, txt.Font(), defaultTextColor)
			textH = renderer.MeasureHeight(objWidth)
		default:
			_, textH = utils.MeasureText(textToMeasure, txt.Font(), objWidth)
			// C# cross-platform build (CROSSPLATFORM && !SKIA, !IsWindows) always
			// uses AdvancedTextRenderer.CalcHeight() which returns pure line-height
			// without GDI+ MeasureString padding. Go matches this path.
			// C# ref: TextObject.cs line 821 — IsAdvancedRendererNeeded || !Config.IsWindows.
		}
		// Add top+bottom padding back into the measured height.
		// C# CalcSize adds Padding.Vertical + 1 (a 1px tolerance).
		if p := txt.Padding(); p.Top+p.Bottom > 0 {
			textH += p.Top + p.Bottom
		}
		textH += 1 // C# adds 1px tolerance in CalcSize()

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
	// C# sorts objects by Top before this loop so that cascading shifts
	// propagate correctly from top to bottom (BandBase.cs CalcHeight).
	shifts := make([]float32, n)
	if hasGrowShrink {
		// Build a sorted index array by Top position (ascending).
		sortedIdx := make([]int, 0, n)
		for i := 0; i < n; i++ {
			if _, hasDim := objs.Get(i).(hasDims); hasDim {
				sortedIdx = append(sortedIdx, i)
			}
		}
		sort.Slice(sortedIdx, func(a, b int) bool {
			return tops[sortedIdx[a]] < tops[sortedIdx[b]]
		})

		for _, i := range sortedIdx {
			d := objs.Get(i).(hasDims)
			delta := effectiveH[i] - d.Height()
			if delta == 0 {
				continue
			}
			parentBottom := tops[i] + d.Height()
			for _, j := range sortedIdx {
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
		return bandLayout{height: baseHeight, shifts: shifts, effectiveH: effectiveH}
	}
	return bandLayout{height: maxBottom, shifts: shifts, effectiveH: effectiveH}
}

// bandNotExportable returns true when the band's Exportable flag is false.
// Returns false (exportable) for types that do not implement the interface.
func bandNotExportable(b report.Base) bool {
	type hasExportable interface{ Exportable() bool }
	if ex, ok := b.(hasExportable); ok {
		return !ex.Exportable()
	}
	return false // default: exportable
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
		Name:          b.Name(),
		Left:          e.curX,
		Top:           e.curY,
		Height:        height,
		NotExportable: !b.Exportable(),
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
	// Evaluate VisibleExpression on the band itself, mirroring C# CanPrint()
	// (ReportEngine.Bands.cs line 259) which is called on the band by
	// PreparedPage.DoAdd() and GetBandHeightWithChildren(). When a
	// VisibleExpression is set it overrides the static Visible flag.
	// C# behaviour: TotalPages-based expressions are true on first pass,
	// then re-evaluated on the final pass.
	if expr := b.VisibleExpression(); expr != "" {
		if e.report != nil {
			visible := b.CalcVisibleExpression(expr, func(s string) (any, error) {
				return e.report.Calc(s)
			})
			if !visible {
				return
			}
		}
	} else if !b.Visible() {
		return
	}

	// Determine whether the child band should be shown after this band.
	// Mirrors the C# ShowBand(band, getData) child-filtering logic:
	//   showChild = child != null
	//     && !(band is DataBand && child.CompleteToNRows > 0)
	//     && !child.FillUnusedSpace
	//     && !(band is DataBand && child.PrintIfDatabandEmpty)
	child := b.Child()
	showChild := child != nil &&
		!child.FillUnusedSpace &&
		!(b.FlagIsDataBand && child.CompleteToNRows > 0) &&
		!(b.FlagIsDataBand && child.PrintIfDatabandEmpty)

	// KeepChild: start a keep scope so parent + child stay on the same page.
	if showChild && b.KeepChild() {
		e.startKeepBand(b)
	}

	// Fire BeforePrint before any rendering takes place, mirroring the
	// C# ReportEngine.ShowBand which calls band.OnBeforePrint first.
	b.FireBeforePrint()
	b.FireBeforeLayout()

	// Apply EvenStyle for alternating data rows.
	// C# ref: BandBase.SaveState() (BandBase.cs:618-645) — after calling
	// SaveState and OnBeforePrint, applies the even style when RowNo is even.
	// The style modifies the band's Fill (e.g. OldLace background) and is
	// restored after rendering via RestoreState.
	evenStyleApplied := e.applyBandEvenStyle(b)
	if evenStyleApplied {
		defer e.restoreBandEvenStyle(b)
	}

	// Populate outline entry if band has an OutlineExpression and is not a
	// reprinted/repeated band. Mirrors C# AddBandOutline: !band.Repeated check.
	// C# ref: ReportEngine.Outline.cs AddBandOutline (line 29):
	//   if (band.Visible && !IsNullOrEmpty(band.OutlineExpression) && !band.Repeated)
	addedOutline := false
	if expr := b.OutlineExpression(); expr != "" && !b.Repeated() {
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
	// Use effectiveFreeSpace to reserve footer area, matching C# behaviour.
	if e.outputBand == nil && b.FlagCheckFreeSpace && e.effectiveFreeSpace() < height {
		e.startNewPageForCurrent()
	}

	// FillUnusedSpace: if the child band has FillUnusedSpace, repeatedly show
	// it to fill any remaining space before rendering this band. Mirrors C#
	// AddToPreparedPages lines 422-441.
	if child != nil && child.FillUnusedSpace && e.outputBand == nil {
		bandHeight := height
		if b.Repeated() {
			bandHeight = 0
		}
		for e.effectiveFreeSpace()-bandHeight-e.CalcBandHeight(&child.BandBase) >= 0 {
			saveCurY := e.curY
			e.ShowFullBand(&child.BandBase)
			// Nothing was printed — break to avoid an endless loop.
			if e.curY == saveCurY {
				break
			}
		}
	}

	// PrintOnBottom: snap band to the bottom of the page (above footer area).
	// Mirrors C# AddToPreparedPages lines 452-465.
	// For a ChildBand, only the band itself is subtracted; for other bands,
	// the full height including child bands is subtracted.
	if b.PrintOnBottom() {
		e.curY = e.pageHeight - e.PageFooterHeight() - e.ColumnFooterHeight()
		// showChild being false while child != nil indicates this is the child
		// scenario (FillUnusedSpace or DataBand exclusion). But for PrintOnBottom,
		// C# checks "band is ChildBand" — which means the band itself is a child.
		// We approximate this: if the band has no child of its own, subtract only
		// its height; otherwise subtract the full chain.
		if b.Child() == nil {
			e.curY -= height
		} else {
			e.curY -= e.GetBandHeightWithChildren(b)
		}
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
			Name:          b.Name(),
			Left:          e.curX,
			Top:           e.curY,
			Height:        height,
			Width:         b.Width(),
			NotExportable: !b.Exportable(),
		}
		// Populate fill/border for band background rendering.
		populateBandProps(b, pb)
		e.populateBandObjects(b, pb)
		applyAnchorAdjustments(b, pb, 0, height-b.Height(), 0)
		// Apply grow/shrink adjustments to PreparedObject positions and sizes.
		// When a CanGrow/CanShrink object changes size, its PreparedObject height
		// must match, and objects below it shift vertically.
		layout := calcBandLayout(b, b.Height(), e.evalText)
		if layout.shifts != nil {
			applyBandObjectShifts(b, pb, layout.shifts)
		}
		if layout.effectiveH != nil {
			applyBandObjectHeights(b, pb, layout.effectiveH)
		}
		// Apply page-level column X offset so data bands in multi-column mode
		// are rendered in their respective column positions.
		if e.curX != 0 {
			for i := range pb.Objects {
				pb.Objects[i].Left += e.curX
			}
		}
		// Apply hierarchy indent: shift the band right and narrow its width.
		// Mirrors C# ReportEngine.Bands.cs lines 469-476:
		//   band.Left += hierarchyIndent; band.Width -= hierarchyIndent;
		// Only non-zero during hierarchical DataBand rendering.
		if e.hierarchyIndent > 0 {
			pb.Left += e.hierarchyIndent
			pb.Width -= e.hierarchyIndent
			for i := range pb.Objects {
				pb.Objects[i].Left += e.hierarchyIndent
			}
		}

		// If objects extend beyond the declared band height (e.g. matrix
		// table cells), grow the band and split across pages.
		// C# ref: Band grows to fit expanded objects; BreakBand splits.
		maxBottom := pb.Height
		for _, po := range pb.Objects {
			if bot := po.Top + po.Height; bot > maxBottom {
				maxBottom = bot
			}
		}
		if maxBottom > pb.Height {
			pb.Height = maxBottom
		}

		// Split the band across pages if it exceeds available space.
		// Use FreeSpace() which deducts page footer height, matching C#.
		if pb.Height > e.FreeSpace() && e.FreeSpace() > 0 && e.pageHeight > 0 {
			e.splitPreparedBandAcrossPages(pb)
		} else {
			_ = e.preparedPages.AddBand(pb)
		}
		// Render inner subreports (PrintOnParent=true) into this prepared band.
		// Mirrors C# PrepareBandShared → RenderInnerSubreports (Bands.cs line 31).
		// Must be called AFTER AddBand so the PreparedBand exists in pg.Bands
		// and can be found by RenderInnerSubreport's backward search.
		e.RenderInnerSubreports(b)

		// Horizontal page splitting for wide matrices.
		// C# ref: TableResult.GeneratePages splits columns across pages.
		if e.pendingHSplit != nil {
			e.splitBandHorizontallyForMatrix(pb)
			e.pendingHSplit = nil
		}
	}
	e.AdvanceY(height)
	b.FireAfterLayout()
	// Fire AfterPrint after the band has been added to the prepared pages,
	// mirroring the C# ReportEngine.ShowBand which calls band.OnAfterPrint last.
	b.FireAfterPrint()

	// Render outer subreports (PrintOnParent=false) before the child band.
	// Mirrors C# ShowBand lines 229-232:
	//   if (band.Visible) RenderOuterSubreports(band);
	// Only when not already in PrintOnParent mode (outputBand != nil would
	// indicate we are inside a subreport rendering).
	if e.outputBand == nil {
		e.RenderOuterSubreports(b)
	}

	// Show child band. Skip if child is used to fill empty space (processed above)
	// or excluded by the DataBand/ChildBand filtering conditions.
	if showChild {
		e.ShowFullBand(&child.BandBase)
		if b.KeepChild() {
			e.EndKeep()
		}
	}

	if addedOutline {
		e.OutlineUp()
	}
}

// ── Band height helpers ──────────────────────────────────────────────────────

// GetBandHeightWithChildren returns the total height of band plus all its
// child bands, mirroring the C# ReportEngine.GetBandHeightWithChildren method.
// It walks the chain: band -> band.Child -> child.Child -> ... and sums heights.
// The walk stops if a child has FillUnusedSpace or CompleteToNRows != 0.
//
// C# special case (ReportEngine.cs GetBandHeightWithChildren lines 376-393):
// When a band's VisibleExpression references "TotalPages" and we are in the
// FinalPass (second pass of a double-pass report), the band's height is still
// included in the calculation even if it is currently not visible. This ensures
// correct footer-area reservation when band visibility depends on total pages.
func (e *ReportEngine) GetBandHeightWithChildren(b *band.BandBase) float32 {
	if b == nil {
		return 0
	}
	result := float32(0)
	cur := b
	for cur != nil {
		include := cur.Visible()
		if !include {
			// C# special case: include height if VisibleExpression contains "TotalPages"
			// and we are in the final pass. This reserves correct space for bands
			// whose visibility depends on the total page count.
			if e.finalPass {
				if expr := cur.VisibleExpression(); expr != "" && strings.Contains(expr, "TotalPages") {
					include = true
				}
			}
		}
		if include {
			if cur.CanGrow() || cur.CanShrink() {
				result += e.CalcBandHeight(cur)
			} else {
				result += cur.Height()
			}
		}
		child := cur.Child()
		if child == nil {
			break
		}
		if child.FillUnusedSpace || child.CompleteToNRows != 0 {
			break
		}
		cur = &child.BandBase
	}
	return result
}

// PageFooterHeight returns the height of the page footer band including its
// child bands, mirroring the C# PageFooterHeight property.
func (e *ReportEngine) PageFooterHeight() float32 {
	if e.currentPage == nil {
		return 0
	}
	pf := e.currentPage.PageFooter()
	if pf == nil {
		return 0
	}
	return e.GetBandHeightWithChildren(&pf.BandBase)
}

// ColumnFooterHeight returns the height of the column footer band including
// its child bands, mirroring the C# ColumnFooterHeight property.
func (e *ReportEngine) ColumnFooterHeight() float32 {
	if e.currentPage == nil {
		return 0
	}
	cf := e.currentPage.ColumnFooter()
	if cf == nil {
		return 0
	}
	return e.GetBandHeightWithChildren(&cf.BandBase)
}

// getReprintFootersHeight returns the total height of all registered reprint
// footer bands, mirroring the C# GetFootersHeight() method.
func (e *ReportEngine) getReprintFootersHeight() float32 {
	result := float32(0)
	for _, entry := range e.reprintFooters {
		saveRepeated := entry.b.Repeated()
		entry.b.SetRepeated(true)
		result += e.GetBandHeightWithChildren(entry.b)
		entry.b.SetRepeated(saveRepeated)
	}
	for _, entry := range e.keepReprintFooters {
		saveRepeated := entry.b.Repeated()
		entry.b.SetRepeated(true)
		result += e.GetBandHeightWithChildren(entry.b)
		entry.b.SetRepeated(saveRepeated)
	}
	return result
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

// applyBandObjectHeights adjusts PreparedObject Height values to match the
// effective heights computed by calcBandLayout (CanGrow/CanShrink expansion).
//
// effectiveH[i] is the computed height for source object i. When it differs
// from the declared height, the PreparedObject is updated so the HTML exporter
// renders the object at the correct size.
func applyBandObjectHeights(bb *band.BandBase, pb *preview.PreparedBand, effectiveH []float32) {
	if len(effectiveH) == 0 {
		return
	}
	objs := bb.Objects()
	if objs == nil || objs.Len() == 0 {
		return
	}

	type hasVisible interface{ Visible() bool }
	type hasGeom interface {
		Top() float32
		Height() float32
	}

	poIdx := 0
	for i := 0; i < objs.Len() && i < len(effectiveH); i++ {
		obj := objs.Get(i)
		if poIdx >= len(pb.Objects) {
			break
		}
		if v, ok := obj.(hasVisible); ok && !v.Visible() {
			continue
		}
		if _, ok := obj.(hasGeom); !ok {
			continue
		}

		// CellularTextObject: skip height adjustment for the container (its
		// dimensions are computed from the table grid, not from FRX Height)
		// and skip over all inserted cell PreparedObjects.
		if _, ok := obj.(*object.CellularTextObject); ok {
			poIdx++
			for poIdx < len(pb.Objects) && strings.Contains(pb.Objects[poIdx].Name, "_r") {
				poIdx++
			}
			continue
		}
		// TableObject inserts cell PreparedObjects after its anchor.
		// Skip over them so the FRX→PreparedObject index mapping stays aligned.
		if _, ok := obj.(*table.TableObject); ok {
			// TableObject inserts cell PreparedObjects after its anchor.
			// Skip the anchor and ALL cells. We don't modify table cell heights —
			// they're computed by the table engine, not by band CanGrow/CanShrink.
			// poIdx advances past all objects in pb.Objects that belong to this table.
			poIdx++ // skip anchor
			// Skip cell PreparedObjects. The count between adjacent FRX object
			// anchors is the number of cells from this table.
			nextAnchor := poIdx
			// Find the next FRX object's anchor position: skip i to i+1 in the
			// FRX loop, peek at the next visible/geom FRX object index.
			nextFRXIdx := i + 1
			for nextFRXIdx < objs.Len() {
				nobj := objs.Get(nextFRXIdx)
				if v, ok2 := nobj.(hasVisible); ok2 && !v.Visible() {
					nextFRXIdx++
					continue
				}
				if _, ok2 := nobj.(hasGeom); !ok2 {
					nextFRXIdx++
					continue
				}
				break
			}
			if nextFRXIdx >= objs.Len() {
				// Table is the last FRX object: skip all remaining pb.Objects.
				nextAnchor = len(pb.Objects)
			} else {
				// The next FRX object starts at this poIdx + cells.
				// Since we don't know the exact cell count, scan for a
				// pb.Object that matches the next FRX object's name.
				nextName := objs.Get(nextFRXIdx).(interface{ Name() string }).Name()
				for nextAnchor < len(pb.Objects) && pb.Objects[nextAnchor].Name != nextName {
					nextAnchor++
				}
			}
			poIdx = nextAnchor
			continue
		}
		// MatrixObject: same skip pattern as TableObject.
		if _, ok := obj.(*matrix.MatrixObject); ok {
			poIdx++
			nextAnchor := poIdx
			nextFRXIdx := i + 1
			for nextFRXIdx < objs.Len() {
				nobj := objs.Get(nextFRXIdx)
				if v, ok2 := nobj.(hasVisible); ok2 && !v.Visible() {
					nextFRXIdx++
					continue
				}
				if _, ok2 := nobj.(hasGeom); !ok2 {
					nextFRXIdx++
					continue
				}
				break
			}
			if nextFRXIdx >= objs.Len() {
				nextAnchor = len(pb.Objects)
			} else {
				nextName := objs.Get(nextFRXIdx).(interface{ Name() string }).Name()
				for nextAnchor < len(pb.Objects) && pb.Objects[nextAnchor].Name != nextName {
					nextAnchor++
				}
			}
			poIdx = nextAnchor
			continue
		}
		if effectiveH[i] != pb.Objects[poIdx].Height {
			pb.Objects[poIdx].Height = effectiveH[i]
		}
		poIdx++
	}
}

// applyBandEvenStyle applies the EvenStyle to the band and its child objects
// when the current row number is even (2, 4, 6, ...). Returns true if the
// style was applied (and thus restoreBandEvenStyle must be called later).
//
// C# ref: BandBase.SaveState() (BandBase.cs:618-645):
//
//	if (RowNo % 2 == 0) { ApplyEvenStyle(); foreach obj: obj.ApplyEvenStyle(); }
func (e *ReportEngine) applyBandEvenStyle(b *band.BandBase) bool {
	if e.report == nil || b.EvenStyleName() == "" || b.RowNo()%2 != 0 {
		return false
	}
	// Save band state.
	b.SaveState()
	b.ApplyEvenStyle(e.report)
	// Save and apply even style on each child object.
	objs := b.Objects()
	if objs != nil {
		for i := 0; i < objs.Len(); i++ {
			type evenStyleable interface {
				SaveState()
				ApplyEvenStyle(report.StyleLookup)
			}
			if es, ok := objs.Get(i).(evenStyleable); ok {
				es.SaveState()
				es.ApplyEvenStyle(e.report)
			}
		}
	}
	return true
}

// restoreBandEvenStyle restores the band and child objects to their pre-EvenStyle
// state. Must be called after rendering when applyBandEvenStyle returned true.
func (e *ReportEngine) restoreBandEvenStyle(b *band.BandBase) {
	// Restore child objects first (reverse of save order).
	objs := b.Objects()
	if objs != nil {
		for i := 0; i < objs.Len(); i++ {
			type restoreable interface {
				RestoreState()
			}
			if rs, ok := objs.Get(i).(restoreable); ok {
				rs.RestoreState()
			}
		}
	}
	b.RestoreState()
}
