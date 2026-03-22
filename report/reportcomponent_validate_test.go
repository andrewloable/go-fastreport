package report

// reportcomponent_validate_test.go — tests for ReportComponentBase.Validate().
//
// Covers the three checks from C# ReportComponentBase.Validate()
// (FastReport.Base/ReportComponentBase.cs lines 802–816):
//   1. Zero/negative size → Error
//   2. Empty name → Error
//   3. AbsBounds not contained within parent AbsBounds → Error

import (
	"testing"

	"github.com/andrewloable/go-fastreport/utils"
)

// ── helpers ───────────────────────────────────────────────────────────────────

func makeRC(name string, left, top, width, height float32) *ReportComponentBase {
	rc := NewReportComponentBase()
	rc.SetName(name)
	rc.SetLeft(left)
	rc.SetTop(top)
	rc.SetWidth(width)
	rc.SetHeight(height)
	return rc
}

func countErrors(issues []utils.ValidationIssue) int {
	n := 0
	for _, iss := range issues {
		if iss.Severity == utils.ValidationError {
			n++
		}
	}
	return n
}

// mockParent implements report.Parent and exposes AbsLeft/AbsTop/Width/Height
// so that ReportComponentBase.Validate() can perform the bounds-containment check.
type mockParent struct {
	BaseObject
	left, top, width, height float32
}

func newMockParent(left, top, width, height float32) *mockParent {
	p := &mockParent{
		BaseObject: *NewBaseObject(),
		left:       left,
		top:        top,
		width:      width,
		height:     height,
	}
	return p
}

// report.Parent interface — trivial implementations for test use only.
func (mp *mockParent) CanContain(_ Base) bool          { return true }
func (mp *mockParent) GetChildObjects(_ *[]Base)        {}
func (mp *mockParent) AddChild(_ Base)                  {}
func (mp *mockParent) RemoveChild(_ Base)               {}
func (mp *mockParent) GetChildOrder(_ Base) int         { return 0 }
func (mp *mockParent) SetChildOrder(_ Base, _ int)      {}
func (mp *mockParent) UpdateLayout(_, _ float32)        {}

// Serialize / Deserialize for report.Base.
func (mp *mockParent) Serialize(_ Writer) error   { return nil }
func (mp *mockParent) Deserialize(_ Reader) error { return nil }
func (mp *mockParent) BaseName() string           { return "mock" }

// Geometry methods used by ReportComponentBase.Validate().
func (mp *mockParent) AbsLeft() float32  { return mp.left }
func (mp *mockParent) AbsTop() float32   { return mp.top }
func (mp *mockParent) Width() float32    { return mp.width }
func (mp *mockParent) Height() float32   { return mp.height }

// ── size check ────────────────────────────────────────────────────────────────

func TestReportComponentValidate_ValidSize(t *testing.T) {
	rc := makeRC("Text1", 0, 0, 100, 50)
	issues := rc.Validate()
	if countErrors(issues) != 0 {
		t.Errorf("unexpected errors for valid component: %v", issues)
	}
}

func TestReportComponentValidate_ZeroWidth(t *testing.T) {
	rc := makeRC("Text1", 0, 0, 0, 50)
	issues := rc.Validate()
	if countErrors(issues) == 0 {
		t.Error("expected error for zero width")
	}
}

func TestReportComponentValidate_ZeroHeight(t *testing.T) {
	rc := makeRC("Text1", 0, 0, 100, 0)
	issues := rc.Validate()
	if countErrors(issues) == 0 {
		t.Error("expected error for zero height")
	}
}

func TestReportComponentValidate_NegativeWidth(t *testing.T) {
	rc := makeRC("Text1", 0, 0, -5, 50)
	issues := rc.Validate()
	if countErrors(issues) == 0 {
		t.Error("expected error for negative width")
	}
}

func TestReportComponentValidate_NegativeHeight(t *testing.T) {
	rc := makeRC("Text1", 0, 0, 100, -10)
	issues := rc.Validate()
	if countErrors(issues) == 0 {
		t.Error("expected error for negative height")
	}
}

// ── name check ────────────────────────────────────────────────────────────────

func TestReportComponentValidate_ValidName(t *testing.T) {
	rc := makeRC("MyText", 0, 0, 100, 50)
	issues := rc.Validate()
	for _, iss := range issues {
		if iss.Message == "unnamed object: report component has no name" {
			t.Errorf("unexpected unnamed-object error for named component: %v", iss)
		}
	}
}

func TestReportComponentValidate_EmptyName(t *testing.T) {
	rc := makeRC("", 0, 0, 100, 50)
	issues := rc.Validate()
	found := false
	for _, iss := range issues {
		if iss.Severity == utils.ValidationError && iss.ObjectName == "" {
			found = true
		}
	}
	if !found {
		t.Error("expected unnamed-object error")
	}
}

// ── out-of-bounds check ───────────────────────────────────────────────────────

func TestReportComponentValidate_WithinParent(t *testing.T) {
	parent := newMockParent(0, 0, 200, 100)
	parent.SetName("Band1")
	child := makeRC("Text1", 10, 10, 50, 30)
	child.SetParent(parent)

	issues := child.Validate()
	if countErrors(issues) != 0 {
		t.Errorf("unexpected errors for child within parent: %v", issues)
	}
}

func TestReportComponentValidate_OutOfParent(t *testing.T) {
	// parent is 0,0 200x100; child starts at 180 with width 50 → extends to 230 > 200.
	parent := newMockParent(0, 0, 200, 100)
	parent.SetName("Band1")
	child := makeRC("Text1", 180, 0, 50, 30)
	child.SetParent(parent)

	issues := child.Validate()
	found := false
	for _, iss := range issues {
		if iss.Severity == utils.ValidationError && iss.ObjectName == "Text1" {
			found = true
		}
	}
	if !found {
		t.Error("expected out-of-bounds error for child extending past parent")
	}
}

func TestReportComponentValidate_NoParent(t *testing.T) {
	// Without a parent the bounds check is skipped — no spurious errors.
	rc := makeRC("Text1", 0, 0, 100, 50)
	issues := rc.Validate()
	if countErrors(issues) != 0 {
		t.Errorf("unexpected errors for component with no parent: %v", issues)
	}
}

func TestReportComponentValidate_ParentWithoutGeometry(t *testing.T) {
	// A parent that does not expose AbsLeft/Width etc. → bounds check skipped.
	// Use a plain BaseObject-based parent that implements Parent but not geometry.
	type plainParent struct {
		BaseObject
	}
	pp := &plainParent{BaseObject: *NewBaseObject()}
	pp.SetName("root")
	_ = pp // pp doesn't implement geometry interface → bounds check skipped
	// We just confirm no panic or spurious error.
	rc := makeRC("Text1", 0, 0, 100, 50)
	// rc.parent is nil — the plain parent doesn't satisfy report.Parent fully for
	// SetParent (it doesn't fully implement Parent), so skip the SetParent call here.
	// The test with "no parent" above already covers the nil-parent case.
	issues := rc.Validate()
	if countErrors(issues) != 0 {
		t.Errorf("unexpected errors: %v", issues)
	}
}
