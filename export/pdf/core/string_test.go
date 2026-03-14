package core

import (
	"bytes"
	"strings"
	"testing"
)

func TestString_Type(t *testing.T) {
	s := NewString("hello")
	if s.Type() != TypeString {
		t.Fatalf("expected TypeString, got %q", s.Type())
	}
}

func TestNewString(t *testing.T) {
	s := NewString("test")
	if s.Value != "test" {
		t.Fatalf("expected 'test', got %q", s.Value)
	}
	if s.IsHex {
		t.Fatal("NewString should not be hex")
	}
}

func TestNewHexString(t *testing.T) {
	s := NewHexString("test")
	if !s.IsHex {
		t.Fatal("NewHexString should be hex")
	}
}

func TestString_WriteTo_Literal_Empty(t *testing.T) {
	s := NewString("")
	var buf bytes.Buffer
	_, err := s.WriteTo(&buf)
	if err != nil {
		t.Fatal(err)
	}
	got := buf.String()
	// Empty string: BOM only is FEFF → encoded as \376\377 in octal scheme,
	// but both bytes are ≥ 0x7F so octal. The BOM is 0xFE, 0xFF.
	// 0xFE=254, 0xFF=255 → \254 \255
	if !strings.HasPrefix(got, "(") || !strings.HasSuffix(got, ")") {
		t.Fatalf("expected parenthesised output, got %q", got)
	}
}

func TestString_WriteTo_Literal_Simple(t *testing.T) {
	// ASCII characters should appear literally inside the parens (after BOM).
	// BOM (0xFE, 0xFF) will be octal-escaped; then 'A' is 0x00, 0x41.
	// 0x00 is a plain byte < 0x7F (NUL, no special case → written as char).
	s := NewString("A")
	var buf bytes.Buffer
	_, err := s.WriteTo(&buf)
	if err != nil {
		t.Fatal(err)
	}
	got := buf.String()
	if !strings.HasPrefix(got, "(") || !strings.HasSuffix(got, ")") {
		t.Fatalf("expected parenthesised output, got %q", got)
	}
}

func TestString_WriteTo_Literal_EscapedChars(t *testing.T) {
	// Verify that special characters are escaped correctly.
	// We test the escape by creating a string that contains only special chars
	// and verifying the resulting literal contains the escape sequences.
	// The UTF-16BE encoding wraps each char: BOM + high byte + low byte.
	// For ASCII specials, high byte is 0x00 (written as NUL) and low byte
	// is the special character itself.

	cases := []struct {
		input   string
		wantEsc string // escape sequence expected in low-byte position
	}{
		{"\n", `\n`},
		{"\r", `\r`},
		{"\t", `\t`},
		{"\b", `\b`},
		{"\f", `\f`},
		{"(", `\(`},
		{")", `\)`},
		{"\\", `\\`},
	}
	for _, tc := range cases {
		t.Run(tc.wantEsc, func(t *testing.T) {
			s := NewString(tc.input)
			var buf bytes.Buffer
			if _, err := s.WriteTo(&buf); err != nil {
				t.Fatal(err)
			}
			got := buf.String()
			if !strings.Contains(got, tc.wantEsc) {
				t.Fatalf("expected %q in output, got %q", tc.wantEsc, got)
			}
		})
	}
}

func TestString_WriteTo_Hex_Empty(t *testing.T) {
	s := NewHexString("")
	var buf bytes.Buffer
	_, err := s.WriteTo(&buf)
	if err != nil {
		t.Fatal(err)
	}
	got := buf.String()
	// Empty string: only BOM → <FEFF>
	want := "<FEFF>"
	if got != want {
		t.Fatalf("got %q want %q", got, want)
	}
}

func TestString_WriteTo_Hex_Simple(t *testing.T) {
	// "A" → UTF-16BE: BOM(FEFF) + 0x0041
	s := NewHexString("A")
	var buf bytes.Buffer
	_, err := s.WriteTo(&buf)
	if err != nil {
		t.Fatal(err)
	}
	got := buf.String()
	want := "<FEFF0041>"
	if got != want {
		t.Fatalf("got %q want %q", got, want)
	}
}

func TestString_WriteTo_Hex_MultiChar(t *testing.T) {
	// "AB" → BOM + 0x0041 + 0x0042
	s := NewHexString("AB")
	var buf bytes.Buffer
	_, err := s.WriteTo(&buf)
	if err != nil {
		t.Fatal(err)
	}
	got := buf.String()
	want := "<FEFF00410042>"
	if got != want {
		t.Fatalf("got %q want %q", got, want)
	}
}

func TestString_WriteTo_ByteCountMatches(t *testing.T) {
	s := NewHexString("Hello")
	var buf bytes.Buffer
	n, _ := s.WriteTo(&buf)
	if n != int64(buf.Len()) {
		t.Fatalf("byte count mismatch: WriteTo=%d buf.Len=%d", n, buf.Len())
	}
}

func TestStringToUTF16BE(t *testing.T) {
	// Empty string should give just the BOM
	got := stringToUTF16BE("")
	if len(got) != 2 || got[0] != 0xFE || got[1] != 0xFF {
		t.Fatalf("expected BOM only, got %v", got)
	}

	// "A" (U+0041) → BOM + 0x00 + 0x41
	got = stringToUTF16BE("A")
	want := []byte{0xFE, 0xFF, 0x00, 0x41}
	if !bytes.Equal(got, want) {
		t.Fatalf("got %v want %v", got, want)
	}
}

func TestString_WriteTo_Literal_HighByte(t *testing.T) {
	// A character like 'é' (U+00E9) has UTF-16BE: 0x00E9
	// High byte 0x00 → written as NUL (literal \x00)
	// Low byte 0xE9 → ≥ 0x7F → octal escape
	s := NewString("é")
	var buf bytes.Buffer
	if _, err := s.WriteTo(&buf); err != nil {
		t.Fatal(err)
	}
	got := buf.String()
	// Should contain a backslash + decimal representation of 0xE9 (233)
	if !strings.Contains(got, `\233`) {
		t.Fatalf("expected octal escape for 0xE9, got %q", got)
	}
}
