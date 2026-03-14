package utils

import (
	"testing"

	"golang.org/x/image/font/basicfont"
)

func TestFontStyleConstants(t *testing.T) {
	if FontStyleRegular != 0 {
		t.Errorf("FontStyleRegular: got %d, want 0", FontStyleRegular)
	}
	if FontStyleBold != 1 {
		t.Errorf("FontStyleBold: got %d, want 1", FontStyleBold)
	}
	if FontStyleItalic != 2 {
		t.Errorf("FontStyleItalic: got %d, want 2", FontStyleItalic)
	}
	if FontStyleBoldItalic != 3 {
		t.Errorf("FontStyleBoldItalic: got %d, want 3", FontStyleBoldItalic)
	}
	if FontStyleUnderline != 4 {
		t.Errorf("FontStyleUnderline: got %d, want 4", FontStyleUnderline)
	}
	if FontStyleStrikeout != 8 {
		t.Errorf("FontStyleStrikeout: got %d, want 8", FontStyleStrikeout)
	}
}

func TestNewFontManager(t *testing.T) {
	fm := NewFontManager()
	if fm == nil {
		t.Fatal("NewFontManager returned nil")
	}
	if fm.cache == nil {
		t.Fatal("cache not initialised")
	}
}

func TestDefaultFontManager(t *testing.T) {
	if DefaultFontManager == nil {
		t.Fatal("DefaultFontManager is nil")
	}
}

func TestDefaultFace(t *testing.T) {
	face := DefaultFace()
	if face != basicfont.Face7x13 {
		t.Errorf("DefaultFace() did not return basicfont.Face7x13")
	}
}

func TestFontManager_GetFace_Fallback(t *testing.T) {
	fm := NewFontManager()
	desc := FontDescriptor{Family: "NonExistent", Size: 12, Style: FontStyleRegular}
	face := fm.GetFace(desc)
	if face != basicfont.Face7x13 {
		t.Errorf("GetFace() fallback: expected basicfont.Face7x13, got %v", face)
	}
}

func TestFontManager_AddFace_And_GetFace(t *testing.T) {
	fm := NewFontManager()
	desc := FontDescriptor{Family: "TestFont", Size: 14, Style: FontStyleBold}
	custom := basicfont.Face7x13 // reuse the same face as a stand-in for a "different" face

	fm.AddFace(desc, custom)

	got := fm.GetFace(desc)
	if got != custom {
		t.Errorf("GetFace() after AddFace: expected registered face, got %v", got)
	}
}

func TestFontManager_GetFace_UnknownDescriptor(t *testing.T) {
	fm := NewFontManager()
	desc := FontDescriptor{Family: "Arial", Size: 10, Style: FontStyleItalic}
	face := fm.GetFace(desc)
	if face == nil {
		t.Fatal("GetFace() returned nil")
	}
}

func TestFontDescriptor_Equality(t *testing.T) {
	a := FontDescriptor{Family: "Courier", Size: 12.0, Style: FontStyleRegular}
	b := FontDescriptor{Family: "Courier", Size: 12.0, Style: FontStyleRegular}
	if a != b {
		t.Error("identical FontDescriptors should be equal")
	}

	c := FontDescriptor{Family: "Courier", Size: 12.0, Style: FontStyleBold}
	if a == c {
		t.Error("FontDescriptors with different styles should not be equal")
	}
}

func TestFontManager_ConcurrentAccess(t *testing.T) {
	fm := NewFontManager()
	desc := FontDescriptor{Family: "Roboto", Size: 11, Style: FontStyleRegular}

	done := make(chan struct{})
	go func() {
		fm.AddFace(desc, basicfont.Face7x13)
		close(done)
	}()

	// Concurrent reads; just ensure no race condition.
	for i := 0; i < 10; i++ {
		_ = fm.GetFace(desc)
	}
	<-done
}
