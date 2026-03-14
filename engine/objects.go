package engine

import (
	"fmt"
	"image/color"

	"github.com/andrewloable/go-fastreport/band"
	"github.com/andrewloable/go-fastreport/gauge"
	"github.com/andrewloable/go-fastreport/object"
	"github.com/andrewloable/go-fastreport/preview"
	"github.com/andrewloable/go-fastreport/report"
	"github.com/andrewloable/go-fastreport/style"
)


// populateBandObjects converts the child report objects of a BandBase into
// preview.PreparedObject snapshots and appends them to pb.
// It evaluates [bracket] expressions in TextObject text via Report.Calc().
func (e *ReportEngine) populateBandObjects(bb *band.BandBase, pb *preview.PreparedBand) {
	if bb == nil {
		return
	}
	e.populateBandObjects2(bb.Objects(), pb)
}

// populateBandObjects2 converts objects from any ObjectCollection into PreparedObjects.
func (e *ReportEngine) populateBandObjects2(objs *report.ObjectCollection, pb *preview.PreparedBand) {
	if objs == nil {
		return
	}
	for i := 0; i < objs.Len(); i++ {
		obj := objs.Get(i)
		if po := e.buildPreparedObject(obj); po != nil {
			pb.Objects = append(pb.Objects, *po)
		}
	}
}

// buildPreparedObject converts a single report.Base object into a PreparedObject,
// or returns nil if the object type is not renderable (e.g. a nested band).
func (e *ReportEngine) buildPreparedObject(obj report.Base) *preview.PreparedObject {
	// Skip invisible and band types.
	type hasVisible interface{ Visible() bool }
	if v, ok := obj.(hasVisible); ok && !v.Visible() {
		return nil
	}

	// Geometry accessors common to all component objects.
	type hasGeom interface {
		Left() float32
		Top() float32
		Width() float32
		Height() float32
	}

	geom, ok := obj.(hasGeom)
	if !ok {
		return nil // no geometry = not a renderable object
	}

	po := &preview.PreparedObject{
		Name:    obj.Name(),
		Left:    geom.Left(),
		Top:     geom.Top(),
		Width:   geom.Width(),
		Height:  geom.Height(),
		BlobIdx: -1,
		Font:    style.DefaultFont(),
	}

	switch v := obj.(type) {
	case *object.HtmlObject:
		po.Kind = preview.ObjectTypeText
		po.TextColor = color.RGBA{A: 255}
		po.Text = e.evalText(v.Text())

	case *object.TextObject:
		po.Kind = preview.ObjectTypeText
		po.Font = v.Font()
		po.TextColor = color.RGBA{A: 255} // default black
		po.HorzAlign = int(v.HorzAlign())
		po.VertAlign = int(v.VertAlign())
		po.WordWrap = v.WordWrap()
		po.Text = e.evalText(v.Text())
		// Apply highlight conditions — first matching condition wins.
		if e.report != nil {
			for _, cond := range v.Highlights() {
				result, err := e.report.Calc(cond.Expression)
				if err != nil {
					continue
				}
				matched, _ := result.(bool)
				if !matched {
					continue
				}
				if !cond.Visible {
					return nil
				}
				if cond.ApplyFill {
					po.FillColor = cond.FillColor
				}
				if cond.ApplyTextFill {
					po.TextColor = cond.TextFillColor
				}
				if cond.ApplyFont {
					po.Font = cond.Font
				}
				break
			}
		}

	case *object.LineObject:
		po.Kind = preview.ObjectTypeLine
		po.LineDiagonal = v.Diagonal()
		po.Border = v.Border()
		if f, ok := v.Fill().(*style.SolidFill); ok {
			po.FillColor = f.Color
		}

	case *object.ShapeObject:
		po.Kind = preview.ObjectTypeShape
		po.ShapeKind = int(v.Shape())
		po.ShapeCurve = v.Curve()
		po.Border = v.Border()
		if f, ok := v.Fill().(*style.SolidFill); ok {
			po.FillColor = f.Color
		}

	case *object.PictureObject:
		po.Kind = preview.ObjectTypePicture
		if data := v.ImageData(); len(data) > 0 {
			po.BlobIdx = e.preparedPages.BlobStore.Add(v.Name(), data)
		}

	case *object.CellularTextObject:
		// CellularTextObject renders like a TextObject; the cellular grid is a
		// visual presentation detail handled at export time.
		po.Kind = preview.ObjectTypeText
		po.Font = v.Font()
		po.TextColor = color.RGBA{A: 255}
		po.HorzAlign = int(v.HorzAlign())
		po.VertAlign = int(v.VertAlign())
		po.WordWrap = v.WordWrap()
		po.Text = e.evalText(v.Text())

	case *object.CheckBoxObject:
		po.Kind = preview.ObjectTypeCheckBox
		// Represent checkbox state as "true" / "false" text for now.
		po.Text = fmt.Sprintf("%v", v.Checked())

	case *gauge.LinearGauge:
		// Render as a text label showing the evaluated expression value.
		po.Kind = preview.ObjectTypeText
		po.TextColor = color.RGBA{A: 255}
		po.Text = e.evalGaugeText(v.Expression, v.Value())

	case *gauge.RadialGauge:
		po.Kind = preview.ObjectTypeText
		po.TextColor = color.RGBA{A: 255}
		po.Text = e.evalGaugeText(v.Expression, v.Value())

	case *gauge.SimpleGauge:
		po.Kind = preview.ObjectTypeText
		po.TextColor = color.RGBA{A: 255}
		po.Text = e.evalGaugeText(v.Expression, v.Value())

	case *gauge.SimpleProgressGauge:
		po.Kind = preview.ObjectTypeText
		po.TextColor = color.RGBA{A: 255}
		po.Text = e.evalGaugeText(v.Expression, v.Value())

	case *object.RichObject:
		// Render RTF text as plain text; full RTF rendering is handled at export time.
		po.Kind = preview.ObjectTypeText
		po.TextColor = color.RGBA{A: 255}
		po.Text = e.evalText(v.Text())

	case *object.SVGObject:
		// SVG is rendered as an embedded picture blob; the export layer decodes SvgData.
		po.Kind = preview.ObjectTypePicture

	case *object.SparklineObject:
		// Sparkline chart rendered as a picture blob at export time.
		po.Kind = preview.ObjectTypePicture

	case *object.AdvMatrixObject:
		// AdvMatrix rendering is not yet implemented; emit an empty picture placeholder.
		po.Kind = preview.ObjectTypePicture

	case *object.MSChartObject:
		// MSChart rendering is not yet implemented; emit an empty picture placeholder.
		po.Kind = preview.ObjectTypePicture

	default:
		// Not a known renderable type (could be a nested band etc.)
		return nil
	}

	return po
}

// evalGaugeText evaluates a gauge expression and formats the result.
// If the expression evaluates successfully the result is returned as a string;
// otherwise the raw default value is shown.
func (e *ReportEngine) evalGaugeText(expr string, defaultVal float64) string {
	if expr == "" {
		return fmt.Sprintf("%g", defaultVal)
	}
	if e.report != nil {
		if result, err := e.report.Calc(expr); err == nil {
			return fmt.Sprintf("%v", result)
		}
	}
	return fmt.Sprintf("%g", defaultVal)
}

// evalText evaluates a text template with [bracket] expressions.
// Returns the raw text on error.
func (e *ReportEngine) evalText(text string) string {
	if e.report == nil || text == "" {
		return text
	}
	result, err := e.report.CalcText(text)
	if err != nil {
		return text
	}
	return result
}
