package core

import (
	"bytes"
	"testing"
)

func TestNumeric_Type(t *testing.T) {
	n := NewInt(0)
	if n.Type() != TypeNumeric {
		t.Fatalf("expected TypeNumeric, got %q", n.Type())
	}
}

func TestNewInt(t *testing.T) {
	n := NewInt(42)
	if !n.IsInt {
		t.Fatal("NewInt should set IsInt=true")
	}
	if n.Value != 42 {
		t.Fatalf("expected Value=42, got %v", n.Value)
	}
}

func TestNewFloat(t *testing.T) {
	n := NewFloat(3.14)
	if n.IsInt {
		t.Fatal("NewFloat should set IsInt=false")
	}
	if n.Value != 3.14 {
		t.Fatalf("expected Value=3.14, got %v", n.Value)
	}
}

func TestNumeric_WriteTo_Int_Zero(t *testing.T) {
	n := NewInt(0)
	var buf bytes.Buffer
	_, err := n.WriteTo(&buf)
	if err != nil {
		t.Fatal(err)
	}
	if buf.String() != "0" {
		t.Fatalf("got %q want %q", buf.String(), "0")
	}
}

func TestNumeric_WriteTo_Int_Positive(t *testing.T) {
	n := NewInt(1024)
	var buf bytes.Buffer
	_, err := n.WriteTo(&buf)
	if err != nil {
		t.Fatal(err)
	}
	if buf.String() != "1024" {
		t.Fatalf("got %q want %q", buf.String(), "1024")
	}
}

func TestNumeric_WriteTo_Int_Negative(t *testing.T) {
	n := NewInt(-7)
	var buf bytes.Buffer
	_, err := n.WriteTo(&buf)
	if err != nil {
		t.Fatal(err)
	}
	if buf.String() != "-7" {
		t.Fatalf("got %q want %q", buf.String(), "-7")
	}
}

func TestNumeric_WriteTo_Float_FourDecimalPlaces(t *testing.T) {
	n := NewFloat(3.14159265)
	var buf bytes.Buffer
	_, err := n.WriteTo(&buf)
	if err != nil {
		t.Fatal(err)
	}
	// strconv.FormatFloat with 'f',4 gives "3.1416" (rounded)
	got := buf.String()
	want := "3.1416"
	if got != want {
		t.Fatalf("got %q want %q", got, want)
	}
}

func TestNumeric_WriteTo_Float_Zero(t *testing.T) {
	n := NewFloat(0)
	var buf bytes.Buffer
	_, err := n.WriteTo(&buf)
	if err != nil {
		t.Fatal(err)
	}
	if buf.String() != "0.0000" {
		t.Fatalf("got %q want %q", buf.String(), "0.0000")
	}
}

func TestNumeric_WriteTo_Float_Negative(t *testing.T) {
	n := NewFloat(-1.5)
	var buf bytes.Buffer
	_, err := n.WriteTo(&buf)
	if err != nil {
		t.Fatal(err)
	}
	if buf.String() != "-1.5000" {
		t.Fatalf("got %q want %q", buf.String(), "-1.5000")
	}
}

func TestNumeric_WriteTo_ByteCountMatches(t *testing.T) {
	n := NewFloat(123.456)
	var buf bytes.Buffer
	nn, _ := n.WriteTo(&buf)
	if nn != int64(buf.Len()) {
		t.Fatalf("byte count mismatch: WriteTo=%d buf.Len=%d", nn, buf.Len())
	}
}

func TestNumeric_WriteTo_Int_ByteCountMatches(t *testing.T) {
	n := NewInt(999)
	var buf bytes.Buffer
	nn, _ := n.WriteTo(&buf)
	if nn != int64(buf.Len()) {
		t.Fatalf("byte count mismatch: WriteTo=%d buf.Len=%d", nn, buf.Len())
	}
}
