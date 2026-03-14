package reportpkg_test

// Tests for Styles deserialization from FRX and engine style application.

import (
	"testing"

	"github.com/andrewloable/go-fastreport/reportpkg"
)

func TestStyles_LoadFromFRX(t *testing.T) {
	// "Advanced Matrix.frx" has:
	// <Style Name="EvenRows" Fill.Color="Honeydew" Font="Arial, 10pt"
	//        ApplyBorder="false" ApplyTextFill="false" ApplyFont="false"/>
	r := loadFRXSmoke(t, "Advanced Matrix.frx")
	ss := r.Styles()
	if ss == nil {
		t.Fatal("Styles() should never be nil")
	}
	if ss.Len() == 0 {
		t.Fatal("expected at least one style in Advanced Matrix.frx")
	}
	e := ss.Find("EvenRows")
	if e == nil {
		t.Fatal("expected style named 'EvenRows'")
	}
	if e.Name != "EvenRows" {
		t.Errorf("style Name = %q, want EvenRows", e.Name)
	}
	// ApplyBorder="false"
	if e.ApplyBorder {
		t.Error("ApplyBorder should be false for EvenRows style")
	}
	// Fill.Color="Honeydew" → non-zero color
	if e.FillColor.A == 0 && e.FillColor.R == 0 && e.FillColor.G == 0 && e.FillColor.B == 0 {
		t.Error("FillColor should be non-zero (Honeydew) after deserialization")
	}
}

func TestStyles_RoundTrip(t *testing.T) {
	r1 := loadFRXSmoke(t, "Advanced Matrix.frx")
	orig := r1.Styles().Len()
	if orig == 0 {
		t.Skip("no styles in Advanced Matrix.frx")
	}
	xml, err := r1.SaveToString()
	if err != nil {
		t.Fatalf("SaveToString: %v", err)
	}
	r2 := reportpkg.NewReport()
	if err := r2.LoadFromString(xml); err != nil {
		t.Fatalf("LoadFromString: %v", err)
	}
	if r2.Styles().Len() != orig {
		t.Errorf("styles after round-trip: got %d, want %d", r2.Styles().Len(), orig)
	}
	e := r2.Styles().Find("EvenRows")
	if e == nil {
		t.Fatal("EvenRows style not found after round-trip")
	}
	if e.ApplyBorder {
		t.Error("ApplyBorder should still be false after round-trip")
	}
}
