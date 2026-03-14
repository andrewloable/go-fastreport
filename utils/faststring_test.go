package utils

import "testing"

func TestNewFastString(t *testing.T) {
	fs := NewFastString()
	if fs == nil {
		t.Fatal("NewFastString returned nil")
	}
	if fs.Len() != 0 {
		t.Errorf("expected Len 0, got %d", fs.Len())
	}
	if fs.String() != "" {
		t.Errorf("expected empty string, got %q", fs.String())
	}
}

func TestFastString_Append(t *testing.T) {
	fs := NewFastString()
	result := fs.Append("hello").Append(", ").Append("world")
	if result != fs {
		t.Error("Append should return receiver for chaining")
	}
	if fs.String() != "hello, world" {
		t.Errorf("expected %q, got %q", "hello, world", fs.String())
	}
}

func TestFastString_AppendEmpty(t *testing.T) {
	fs := NewFastString()
	fs.Append("")
	if fs.Len() != 0 {
		t.Errorf("expected Len 0 after appending empty string, got %d", fs.Len())
	}
}

func TestFastString_AppendLine(t *testing.T) {
	fs := NewFastString()
	fs.AppendLine("line1").AppendLine("line2")
	expected := "line1\nline2\n"
	if fs.String() != expected {
		t.Errorf("expected %q, got %q", expected, fs.String())
	}
}

func TestFastString_AppendLine_Empty(t *testing.T) {
	fs := NewFastString()
	fs.AppendLine("")
	if fs.String() != "\n" {
		t.Errorf("expected newline only, got %q", fs.String())
	}
}

func TestFastString_AppendRune(t *testing.T) {
	fs := NewFastString()
	fs.AppendRune('A').AppendRune('€').AppendRune('Z')
	expected := "A€Z"
	if fs.String() != expected {
		t.Errorf("expected %q, got %q", expected, fs.String())
	}
}

func TestFastString_Len(t *testing.T) {
	fs := NewFastString()
	fs.Append("abc")
	if fs.Len() != 3 {
		t.Errorf("expected Len 3, got %d", fs.Len())
	}
}

func TestFastString_Reset(t *testing.T) {
	fs := NewFastString()
	fs.Append("some content")
	fs.Reset()
	if fs.Len() != 0 {
		t.Errorf("expected Len 0 after Reset, got %d", fs.Len())
	}
	if fs.String() != "" {
		t.Errorf("expected empty string after Reset, got %q", fs.String())
	}
	// Should be reusable after Reset.
	fs.Append("new")
	if fs.String() != "new" {
		t.Errorf("expected %q after reuse, got %q", "new", fs.String())
	}
}

func TestFastString_IsEmpty(t *testing.T) {
	fs := NewFastString()
	if !fs.IsEmpty() {
		t.Error("expected IsEmpty true for new FastString")
	}
	fs.Append("x")
	if fs.IsEmpty() {
		t.Error("expected IsEmpty false after Append")
	}
	fs.Reset()
	if !fs.IsEmpty() {
		t.Error("expected IsEmpty true after Reset")
	}
}

func TestFastString_Chaining(t *testing.T) {
	fs := NewFastString()
	got := fs.Append("a").AppendRune('b').AppendLine("c").String()
	expected := "abc\n"
	if got != expected {
		t.Errorf("chaining: expected %q, got %q", expected, got)
	}
}
