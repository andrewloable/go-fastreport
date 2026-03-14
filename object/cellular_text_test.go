package object_test

import (
	"testing"

	"github.com/andrewloable/go-fastreport/object"
)

func TestNewCellularTextObject_Defaults(t *testing.T) {
	c := object.NewCellularTextObject()
	if c == nil {
		t.Fatal("NewCellularTextObject returned nil")
	}
	if !c.WordWrap() {
		t.Error("WordWrap should default to true")
	}
	if c.CellWidth() != 0 {
		t.Errorf("CellWidth = %v, want 0", c.CellWidth())
	}
	if c.CellHeight() != 0 {
		t.Errorf("CellHeight = %v, want 0", c.CellHeight())
	}
	if c.HorzSpacing() != 0 {
		t.Errorf("HorzSpacing = %v, want 0", c.HorzSpacing())
	}
	if c.TypeName() != "CellularTextObject" {
		t.Errorf("TypeName = %q, want CellularTextObject", c.TypeName())
	}
}

func TestCellularTextObject_SetFields(t *testing.T) {
	c := object.NewCellularTextObject()
	c.SetCellWidth(28.35)
	c.SetCellHeight(28.35)
	c.SetHorzSpacing(7.56)
	c.SetWordWrap(false)

	if c.CellWidth() != 28.35 {
		t.Errorf("CellWidth = %v, want 28.35", c.CellWidth())
	}
	if c.CellHeight() != 28.35 {
		t.Errorf("CellHeight = %v, want 28.35", c.CellHeight())
	}
	if c.HorzSpacing() != 7.56 {
		t.Errorf("HorzSpacing = %v, want 7.56", c.HorzSpacing())
	}
	if c.WordWrap() {
		t.Error("WordWrap should be false")
	}
}

func TestCellularTextObject_ImplementsReportBase(t *testing.T) {
	c := object.NewCellularTextObject()
	c.SetName("ct1")
	if c.Name() != "ct1" {
		t.Errorf("Name = %q, want ct1", c.Name())
	}
	if c.BaseName() != "CellularText" {
		t.Errorf("BaseName = %q, want CellularText", c.BaseName())
	}
}
