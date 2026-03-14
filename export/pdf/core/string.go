package core

import (
	"fmt"
	"io"
	"strings"
)

// String is a PDF string object.  It can be rendered either as a
// parenthesised literal string or as an angle-bracket hex string.
//
// Literal representation:  (Hello\nWorld)
// Hex representation:      <FEFF00480065006C006C006F>
//
// In both cases the string value is first converted to a big-endian UTF-16
// sequence with a BOM (0xFEFF), matching the behaviour of the original C#
// implementation.
type String struct {
	// Value is the Go string to be encoded.
	Value string
	// IsHex selects the hex representation when true; parenthesised otherwise.
	IsHex bool
}

// NewString returns a literal (parenthesised) PDF string.
func NewString(v string) *String { return &String{Value: v} }

// NewHexString returns a hex PDF string.
func NewHexString(v string) *String { return &String{Value: v, IsHex: true} }

// Type implements Object.
func (s *String) Type() ObjectType { return TypeString }

// WriteTo writes the PDF string representation to w.
func (s *String) WriteTo(w io.Writer) (int64, error) {
	cw := &countWriter{w: w}
	var err error
	if s.IsHex {
		err = writeHexString(cw, s.Value)
	} else {
		err = writeLiteralString(cw, s.Value)
	}
	return cw.n, err
}

// stringToUTF16BE converts a Go string to a big-endian UTF-16 byte sequence
// prefixed with the BOM (0xFE, 0xFF).  Each rune occupies exactly two bytes
// (characters above U+FFFF are truncated to the low 16 bits), matching the
// original C# char-level conversion.
func stringToUTF16BE(s string) []byte {
	result := make([]byte, 0, 2+len(s)*2)
	// BOM
	result = append(result, 0xFE, 0xFF)
	for _, r := range s {
		hi := byte(r >> 8)
		lo := byte(r & 0xFF)
		result = append(result, hi, lo)
	}
	return result
}

// writeHexString writes the hex representation of s: <FEFF…>
func writeHexString(w io.Writer, s string) error {
	if _, err := fmt.Fprint(w, "<"); err != nil {
		return err
	}
	for _, b := range stringToUTF16BE(s) {
		if _, err := fmt.Fprintf(w, "%02X", b); err != nil {
			return err
		}
	}
	_, err := fmt.Fprint(w, ">")
	return err
}

// writeLiteralString writes the parenthesised representation of s.
// Bytes below 0x7F are written with special-character escaping; bytes ≥ 0x7F
// are written as octal escape sequences (\ddd).
func writeLiteralString(w io.Writer, s string) error {
	if _, err := fmt.Fprint(w, "("); err != nil {
		return err
	}
	var sb strings.Builder
	for _, b := range stringToUTF16BE(s) {
		if b < 0x7F {
			switch b {
			case '\n':
				sb.WriteString(`\n`)
			case '\r':
				sb.WriteString(`\r`)
			case '\t':
				sb.WriteString(`\t`)
			case '\b':
				sb.WriteString(`\b`)
			case '\f':
				sb.WriteString(`\f`)
			case '(':
				sb.WriteString(`\(`)
			case ')':
				sb.WriteString(`\)`)
			case '\\':
				sb.WriteString(`\\`)
			default:
				sb.WriteByte(b)
			}
		} else {
			// Octal escape for bytes ≥ 0x7F
			fmt.Fprintf(&sb, "\\%d", int(b))
		}
	}
	if _, err := fmt.Fprint(w, sb.String()); err != nil {
		return err
	}
	_, err := fmt.Fprint(w, ")")
	return err
}
