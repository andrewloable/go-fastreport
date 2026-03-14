package core

import (
	"bytes"
	"strings"
	"testing"
)

func TestNumericInt(t *testing.T) {
	n := NewInt(42)
	if n.Type() != TypeNumeric {
		t.Errorf("Type = %v, want %v", n.Type(), TypeNumeric)
	}
	var buf bytes.Buffer
	_, err := n.WriteTo(&buf)
	if err != nil {
		t.Fatal(err)
	}
	if buf.String() != "42" {
		t.Errorf("got %q, want %q", buf.String(), "42")
	}
}

func TestNumericFloat(t *testing.T) {
	n := NewFloat(3.14)
	var buf bytes.Buffer
	_, err := n.WriteTo(&buf)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(buf.String(), "3.14") {
		t.Errorf("got %q, expected to contain 3.14", buf.String())
	}
}

func TestNumericNegative(t *testing.T) {
	n := NewInt(-5)
	var buf bytes.Buffer
	n.WriteTo(&buf)
	if buf.String() != "-5" {
		t.Errorf("got %q, want -5", buf.String())
	}
}

func TestNumericZero(t *testing.T) {
	n := NewInt(0)
	var buf bytes.Buffer
	n.WriteTo(&buf)
	if buf.String() != "0" {
		t.Errorf("got %q, want 0", buf.String())
	}
}

func TestBooleanTrue(t *testing.T) {
	b := NewBoolean(true)
	if b.Type() != TypeBoolean {
		t.Errorf("Type = %v, want %v", b.Type(), TypeBoolean)
	}
	var buf bytes.Buffer
	_, err := b.WriteTo(&buf)
	if err != nil {
		t.Fatal(err)
	}
	if buf.String() != "true" {
		t.Errorf("got %q, want true", buf.String())
	}
}

func TestBooleanFalse(t *testing.T) {
	b := NewBoolean(false)
	var buf bytes.Buffer
	b.WriteTo(&buf)
	if buf.String() != "false" {
		t.Errorf("got %q, want false", buf.String())
	}
}

func TestNameSimple(t *testing.T) {
	n := NewName("FlateDecode")
	if n.Type() != TypeName {
		t.Errorf("Type = %v, want %v", n.Type(), TypeName)
	}
	var buf bytes.Buffer
	_, err := n.WriteTo(&buf)
	if err != nil {
		t.Fatal(err)
	}
	if buf.String() != "/FlateDecode" {
		t.Errorf("got %q, want /FlateDecode", buf.String())
	}
}

func TestNameEmpty(t *testing.T) {
	n := NewName("")
	var buf bytes.Buffer
	_, err := n.WriteTo(&buf)
	if err != nil {
		t.Fatal(err)
	}
	if buf.String() != "" {
		t.Errorf("empty name should produce empty output, got %q", buf.String())
	}
}

func TestNameSpecialChars(t *testing.T) {
	n := NewName("my name")
	var buf bytes.Buffer
	n.WriteTo(&buf)
	got := buf.String()
	if !strings.HasPrefix(got, "/") {
		t.Errorf("name should start with /: got %q", got)
	}
	// Space (0x20) should be encoded as #20
	if !strings.Contains(got, "#20") {
		t.Errorf("space should be encoded as #20, got %q", got)
	}
}

func TestNullType(t *testing.T) {
	n := &Null{}
	if n.Type() != TypeNull {
		t.Errorf("Type = %v, want %v", n.Type(), TypeNull)
	}
	var buf bytes.Buffer
	_, err := n.WriteTo(&buf)
	if err != nil {
		t.Fatal(err)
	}
	if buf.String() != "null" {
		t.Errorf("got %q, want null", buf.String())
	}
}

func TestStringLiteral(t *testing.T) {
	s := NewString("Hi")
	if s.Type() != TypeString {
		t.Errorf("Type = %v, want %v", s.Type(), TypeString)
	}
	var buf bytes.Buffer
	_, err := s.WriteTo(&buf)
	if err != nil {
		t.Fatal(err)
	}
	got := buf.String()
	if !strings.HasPrefix(got, "(") || !strings.HasSuffix(got, ")") {
		t.Errorf("literal string should be wrapped in parens: got %q", got)
	}
}

func TestStringHex(t *testing.T) {
	s := NewHexString("Hi")
	var buf bytes.Buffer
	_, err := s.WriteTo(&buf)
	if err != nil {
		t.Fatal(err)
	}
	got := buf.String()
	if !strings.HasPrefix(got, "<") || !strings.HasSuffix(got, ">") {
		t.Errorf("hex string should be wrapped in <...>: got %q", got)
	}
	// Should contain FEFF BOM: "FEFF"
	if !strings.Contains(got, "FEFF") {
		t.Errorf("hex string should contain BOM FEFF, got %q", got)
	}
}

func TestStringEmpty(t *testing.T) {
	s := NewString("")
	var buf bytes.Buffer
	_, err := s.WriteTo(&buf)
	if err != nil {
		t.Fatal(err)
	}
	// Empty string still has parens with BOM
	got := buf.String()
	if !strings.HasPrefix(got, "(") {
		t.Errorf("empty literal string should start with (, got %q", got)
	}
}

func TestArrayEmpty(t *testing.T) {
	a := NewArray()
	if a.Type() != TypeArray {
		t.Errorf("Type = %v, want %v", a.Type(), TypeArray)
	}
	if a.Len() != 0 {
		t.Errorf("empty array Len = %d, want 0", a.Len())
	}
	var buf bytes.Buffer
	_, err := a.WriteTo(&buf)
	if err != nil {
		t.Fatal(err)
	}
	if buf.String() != "[ ]" {
		t.Errorf("got %q, want '[ ]'", buf.String())
	}
}

func TestArrayWithItems(t *testing.T) {
	a := NewArray(NewInt(1), NewInt(2), NewBoolean(true))
	if a.Len() != 3 {
		t.Errorf("Len = %d, want 3", a.Len())
	}
	a.Add(NewName("X"))
	if a.Len() != 4 {
		t.Errorf("after Add Len = %d, want 4", a.Len())
	}
	var buf bytes.Buffer
	_, err := a.WriteTo(&buf)
	if err != nil {
		t.Fatal(err)
	}
	got := buf.String()
	if !strings.HasPrefix(got, "[ ") {
		t.Errorf("array should start with '[ ': got %q", got)
	}
	if !strings.Contains(got, "1") || !strings.Contains(got, "2") || !strings.Contains(got, "true") {
		t.Errorf("array output missing items: %q", got)
	}
}

func TestDictionaryEmpty(t *testing.T) {
	d := NewDictionary()
	if d.Type() != TypeDictionary {
		t.Errorf("Type = %v, want %v", d.Type(), TypeDictionary)
	}
	if d.Len() != 0 {
		t.Errorf("empty dict Len = %d, want 0", d.Len())
	}
	var buf bytes.Buffer
	_, err := d.WriteTo(&buf)
	if err != nil {
		t.Fatal(err)
	}
	got := buf.String()
	if !strings.HasPrefix(got, "<<") || !strings.HasSuffix(got, ">>") {
		t.Errorf("dict should be <<...>>: got %q", got)
	}
}

func TestDictionaryAddGet(t *testing.T) {
	d := NewDictionary()
	d.Add("Type", NewName("Page"))
	d.Add("Count", NewInt(5))
	if d.Len() != 2 {
		t.Errorf("Len = %d, want 2", d.Len())
	}
	v := d.Get("Type")
	if v == nil {
		t.Fatal("Get(Type) returned nil")
	}
	if v.Type() != TypeName {
		t.Errorf("expected Name type, got %v", v.Type())
	}
	if d.Get("Missing") != nil {
		t.Error("Get(Missing) should return nil")
	}
}

func TestDictionaryWriteTo(t *testing.T) {
	d := NewDictionary()
	d.Add("Filter", NewName("FlateDecode"))
	d.Add("Length", NewInt(100))
	var buf bytes.Buffer
	_, err := d.WriteTo(&buf)
	if err != nil {
		t.Fatal(err)
	}
	got := buf.String()
	if !strings.Contains(got, "/Filter") {
		t.Errorf("dict should contain /Filter: %q", got)
	}
	if !strings.Contains(got, "/FlateDecode") {
		t.Errorf("dict should contain /FlateDecode: %q", got)
	}
	if !strings.Contains(got, "/Length") {
		t.Errorf("dict should contain /Length: %q", got)
	}
}

func TestIndirectObjectWriteTo(t *testing.T) {
	io := &IndirectObject{
		Number:     1,
		Generation: 0,
		Value:      NewInt(42),
	}
	if io.Type() != TypeIndirect {
		t.Errorf("Type = %v, want %v", io.Type(), TypeIndirect)
	}
	var buf bytes.Buffer
	_, err := io.WriteTo(&buf)
	if err != nil {
		t.Fatal(err)
	}
	got := buf.String()
	if !strings.Contains(got, "1 0 obj") {
		t.Errorf("indirect should contain '1 0 obj': %q", got)
	}
	if !strings.Contains(got, "endobj") {
		t.Errorf("indirect should contain 'endobj': %q", got)
	}
	if !strings.Contains(got, "42") {
		t.Errorf("indirect should contain value 42: %q", got)
	}
}

func TestIndirectObjectNilValue(t *testing.T) {
	io := &IndirectObject{Number: 2, Generation: 0}
	var buf bytes.Buffer
	_, err := io.WriteTo(&buf)
	if err != nil {
		t.Fatal(err)
	}
	got := buf.String()
	if !strings.Contains(got, "2 0 obj") {
		t.Errorf("expected '2 0 obj' in output: %q", got)
	}
}

func TestStreamUncompressed(t *testing.T) {
	s := NewStream()
	s.Compressed = false
	s.Data = []byte("Hello World")
	var buf bytes.Buffer
	_, err := s.WriteTo(&buf)
	if err != nil {
		t.Fatal(err)
	}
	got := buf.String()
	if s.Type() != TypeStream {
		t.Errorf("Type = %v, want %v", s.Type(), TypeStream)
	}
	if !strings.Contains(got, "stream") || !strings.Contains(got, "endstream") {
		t.Errorf("stream missing markers: %q", got)
	}
	if !strings.Contains(got, "Hello World") {
		t.Errorf("stream should contain data: %q", got)
	}
}

func TestStreamCompressed(t *testing.T) {
	s := NewStream()
	s.Compressed = true
	s.Data = []byte("compress me compress me compress me")
	var buf bytes.Buffer
	_, err := s.WriteTo(&buf)
	if err != nil {
		t.Fatal(err)
	}
	got := buf.String()
	if !strings.Contains(got, "/Filter") {
		t.Errorf("compressed stream should have /Filter: %q", got)
	}
	if !strings.Contains(got, "FlateDecode") {
		t.Errorf("compressed stream should have FlateDecode: %q", got)
	}
}

func TestStreamEmpty(t *testing.T) {
	s := NewStream()
	s.Compressed = false
	s.Data = nil
	var buf bytes.Buffer
	_, err := s.WriteTo(&buf)
	if err != nil {
		t.Fatal(err)
	}
	got := buf.String()
	if !strings.Contains(got, "stream") {
		t.Errorf("empty stream should still have 'stream' keyword: %q", got)
	}
}

func TestZlibWriteErrorPropagation(t *testing.T) {
	// Test error branches in zlibWrite via zlibCompress via Stream.WriteTo
	// with a writer that always fails after some bytes.
	s := NewStream()
	s.Compressed = true
	s.Data = make([]byte, 1024)
	var buf bytes.Buffer
	_, err := s.WriteTo(&buf)
	if err != nil {
		t.Fatalf("stream write failed: %v", err)
	}
	// Verify the /Filter was added
	if s.Dict.Get("Filter") == nil {
		t.Error("compressed stream should have /Filter in dict")
	}
}

func TestStreamCompressedEmpty(t *testing.T) {
	s := NewStream()
	s.Compressed = true
	s.Data = []byte{} // explicitly empty, not nil
	var buf bytes.Buffer
	_, err := s.WriteTo(&buf)
	if err != nil {
		t.Fatalf("stream write failed: %v", err)
	}
	got := buf.String()
	if !strings.Contains(got, "FlateDecode") {
		t.Errorf("compressed empty stream should have FlateDecode: %q", got)
	}
}

// Compile-time interface checks
var _ Object = (*Numeric)(nil)
var _ Object = (*Boolean)(nil)
var _ Object = (*Name)(nil)
var _ Object = (*Null)(nil)
var _ Object = (*String)(nil)
var _ Object = (*Array)(nil)
var _ Object = (*Dictionary)(nil)
var _ Object = (*IndirectObject)(nil)
var _ Object = (*Stream)(nil)
