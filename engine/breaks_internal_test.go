package engine

import (
	"testing"

	"github.com/andrewloable/go-fastreport/object"
	"github.com/andrewloable/go-fastreport/preview"
	"github.com/andrewloable/go-fastreport/report"
	"github.com/andrewloable/go-fastreport/reportpkg"
)

func newBreaksInternalEngine(t *testing.T) *ReportEngine {
	t.Helper()
	r := reportpkg.NewReport()
	pg := reportpkg.NewReportPage()
	r.AddPage(pg)
	e := New(r)
	if err := e.Run(DefaultRunOptions()); err != nil {
		t.Fatalf("Run: %v", err)
	}
	return e
}

func TestSplitPopulateTop_NilObjs(t *testing.T) {
	e := newBreaksInternalEngine(t)
	pb := &preview.PreparedBand{Name: "test", Top: 0, Height: 50}
	e.splitPopulateTop(nil, pb, 50)
	if len(pb.Objects) != 0 {
		t.Errorf("expected 0 objects, got %d", len(pb.Objects))
	}
}

func TestSplitPopulateTop_InvisibleObjectAboveBreakLine(t *testing.T) {
	e := newBreaksInternalEngine(t)
	objs := report.NewObjectCollection()
	txt := object.NewTextObject()
	txt.SetTop(5)
	txt.SetHeight(20)
	txt.SetWidth(100)
	txt.SetText("invisible")
	txt.SetVisible(false)
	objs.Add(txt)
	pb := &preview.PreparedBand{Name: "test", Top: 0, Height: 50}
	e.splitPopulateTop(objs, pb, 50)
	if len(pb.Objects) != 0 {
		t.Errorf("expected 0 objects, got %d", len(pb.Objects))
	}
}

func TestSplitPopulateTop_InvisibleObjectStraddlingBreakLine(t *testing.T) {
	e := newBreaksInternalEngine(t)
	objs := report.NewObjectCollection()
	txt := object.NewTextObject()
	txt.SetTop(30)
	txt.SetHeight(40)
	txt.SetWidth(100)
	txt.SetText("straddle")
	txt.SetVisible(false)
	objs.Add(txt)
	pb := &preview.PreparedBand{Name: "test", Top: 0, Height: 50}
	e.splitPopulateTop(objs, pb, 50)
	if len(pb.Objects) != 0 {
		t.Errorf("expected 0 objects, got %d", len(pb.Objects))
	}
}

func TestSplitPopulateTop_BreakLineAtZero(t *testing.T) {
	e := newBreaksInternalEngine(t)
	objs := report.NewObjectCollection()
	txt := object.NewTextObject()
	txt.SetTop(0)
	txt.SetHeight(30)
	txt.SetWidth(100)
	txt.SetText("at zero")
	txt.SetVisible(true)
	objs.Add(txt)
	pb := &preview.PreparedBand{Name: "test", Top: 0, Height: 0}
	e.splitPopulateTop(objs, pb, 0)
	if len(pb.Objects) != 0 {
		t.Errorf("expected 0 objects, got %d", len(pb.Objects))
	}
}

func TestSplitPopulateBottom_NilObjs(t *testing.T) {
	e := newBreaksInternalEngine(t)
	pb := &preview.PreparedBand{Name: "test", Top: 0, Height: 100}
	e.splitPopulateBottom(nil, pb, 50)
	if len(pb.Objects) != 0 {
		t.Errorf("expected 0 objects, got %d", len(pb.Objects))
	}
}

func TestSplitPopulateBottom_InvisibleObjectBelowBreakLine(t *testing.T) {
	e := newBreaksInternalEngine(t)
	objs := report.NewObjectCollection()
	txt := object.NewTextObject()
	txt.SetTop(70)
	txt.SetHeight(20)
	txt.SetWidth(100)
	txt.SetText("invisible below")
	txt.SetVisible(false)
	objs.Add(txt)
	pb := &preview.PreparedBand{Name: "test", Top: 0, Height: 100}
	e.splitPopulateBottom(objs, pb, 50)
	if len(pb.Objects) != 0 {
		t.Errorf("expected 0 objects, got %d", len(pb.Objects))
	}
}

func TestSplitPopulateBottom_InvisibleObjectStraddlingBreakLine(t *testing.T) {
	e := newBreaksInternalEngine(t)
	objs := report.NewObjectCollection()
	txt := object.NewTextObject()
	txt.SetTop(30)
	txt.SetHeight(50)
	txt.SetWidth(100)
	txt.SetText("straddle")
	txt.SetVisible(false)
	objs.Add(txt)
	pb := &preview.PreparedBand{Name: "test", Top: 0, Height: 100}
	e.splitPopulateBottom(objs, pb, 50)
	if len(pb.Objects) != 0 {
		t.Errorf("expected 0 objects, got %d", len(pb.Objects))
	}
}

func TestSplitPopulateBottom_BreakLineAtBandHeight(t *testing.T) {
	e := newBreaksInternalEngine(t)
	objs := report.NewObjectCollection()
	txt := object.NewTextObject()
	txt.SetTop(5)
	txt.SetHeight(20)
	txt.SetWidth(100)
	txt.SetText("above all")
	txt.SetVisible(true)
	objs.Add(txt)
	pb := &preview.PreparedBand{Name: "test", Top: 0, Height: 0}
	e.splitPopulateBottom(objs, pb, 100)
	if len(pb.Objects) != 0 {
		t.Errorf("expected 0 objects, got %d", len(pb.Objects))
	}
}
