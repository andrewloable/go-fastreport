package crossview_test

import (
	"testing"

	"github.com/andrewloable/go-fastreport/crossview"
)

func TestNewCrossViewHelper_NotNil(t *testing.T) {
	cv := crossview.NewCrossViewObject()
	h := crossview.NewCrossViewHelper(cv)
	if h == nil {
		t.Fatal("NewCrossViewHelper returned nil")
	}
	if h.CrossView() != cv {
		t.Error("CrossView() should return the associated CrossViewObject")
	}
}

func TestCrossViewHelper_Defaults(t *testing.T) {
	cv := crossview.NewCrossViewObject()
	h := crossview.NewCrossViewHelper(cv)
	if h.HeaderWidth() != 0 || h.HeaderHeight() != 0 {
		t.Errorf("fresh helper: HeaderWidth=%d HeaderHeight=%d, want 0,0", h.HeaderWidth(), h.HeaderHeight())
	}
	if h.TemplateBodyWidth() != 0 || h.TemplateBodyHeight() != 0 {
		t.Errorf("fresh helper: BodyWidth=%d BodyHeight=%d, want 0,0", h.TemplateBodyWidth(), h.TemplateBodyHeight())
	}
}

func TestCrossViewHelper_UpdateTemplateSizes_NoSource(t *testing.T) {
	cv := crossview.NewCrossViewObject()
	// Disable optional rows so we can test the pure no-source base of 1,1.
	cv.ShowTitle = false
	cv.ShowXAxisFieldsCaption = false
	h := crossview.NewCrossViewHelper(cv)
	h.UpdateTemplateSizes()
	// No data source assigned and no caption rows: header{w,h}=1
	if h.HeaderWidth() != 1 {
		t.Errorf("no-source HeaderWidth = %d, want 1", h.HeaderWidth())
	}
	if h.HeaderHeight() != 1 {
		t.Errorf("no-source HeaderHeight = %d, want 1", h.HeaderHeight())
	}
}

func TestCrossViewHelper_BuildTemplate_CallsUpdate(t *testing.T) {
	cv := crossview.NewCrossViewObject()
	cv.ShowTitle = false
	cv.ShowXAxisFieldsCaption = false
	h := crossview.NewCrossViewHelper(cv)
	// Before: all zeros
	if h.HeaderWidth() != 0 {
		t.Fatal("expected zero before BuildTemplate")
	}
	h.BuildTemplate()
	// After: no-source defaults with captions disabled → header 1,1
	if h.HeaderWidth() != 1 || h.HeaderHeight() != 1 {
		t.Errorf("after BuildTemplate: HeaderWidth=%d HeaderHeight=%d, want 1,1", h.HeaderWidth(), h.HeaderHeight())
	}
}

func TestCrossViewHelper_StartPrint_ResetsDimensions(t *testing.T) {
	cv := crossview.NewCrossViewObject()
	h := crossview.NewCrossViewHelper(cv)
	// Before StartPrint: result dimensions are 0
	if h.ResultBodyWidth() != 0 || h.ResultBodyHeight() != 0 {
		t.Error("expected zero result dimensions before StartPrint")
	}
	h.StartPrint()
	// After StartPrint with no data source: still 0 (no rows/columns)
	if h.ResultBodyWidth() != 0 || h.ResultBodyHeight() != 0 {
		t.Errorf("StartPrint with no data: ResultBodyWidth=%d ResultBodyHeight=%d, want 0,0",
			h.ResultBodyWidth(), h.ResultBodyHeight())
	}
}

func TestCrossViewHelper_CreateOtherDescriptor_ResetsGeometry(t *testing.T) {
	cv := crossview.NewCrossViewObject()
	cv.ShowTitle = false
	cv.ShowXAxisFieldsCaption = false
	h := crossview.NewCrossViewHelper(cv)
	h.BuildTemplate() // sets header to 1,1 (no-source, no captions)
	h.CreateOtherDescriptor()
	// After reset: geometry goes back to 0
	if h.HeaderWidth() != 0 || h.HeaderHeight() != 0 {
		t.Errorf("after CreateOtherDescriptor: HeaderWidth=%d HeaderHeight=%d, want 0,0",
			h.HeaderWidth(), h.HeaderHeight())
	}
}

func TestCrossViewHelper_FinishPrint_CallsDescriptors(t *testing.T) {
	cv := crossview.NewCrossViewObject()
	cv.ShowTitle = false
	cv.ShowXAxisFieldsCaption = false
	h := crossview.NewCrossViewHelper(cv)
	// FinishPrint should not panic even with empty state
	h.FinishPrint()
	// After FinishPrint, geometry is valid (no-source defaults from UpdateDescriptors)
	if h.HeaderWidth() != 1 {
		t.Errorf("after FinishPrint: HeaderWidth = %d, want 1", h.HeaderWidth())
	}
}
