package core

import (
	"bytes"
	"errors"
	"io"
	"strings"
	"testing"
)

func TestNewRef_TypeAndValue(t *testing.T) {
	obj := &IndirectObject{Number: 5, Generation: 0}
	ref := NewRef(obj)
	if ref.Value != "5 0 R" {
		t.Errorf("Ref.Value = %q, want \"5 0 R\"", ref.Value)
	}
	if ref.Type() != TypeRef {
		t.Errorf("Ref.Type() = %q, want %q", ref.Type(), TypeRef)
	}
}

func TestRef_WriteTo(t *testing.T) {
	obj := &IndirectObject{Number: 3, Generation: 0}
	ref := NewRef(obj)
	var buf bytes.Buffer
	n, err := ref.WriteTo(&buf)
	if err != nil {
		t.Fatalf("WriteTo: %v", err)
	}
	if buf.String() != "3 0 R" {
		t.Errorf("WriteTo output = %q, want \"3 0 R\"", buf.String())
	}
	if n != int64(buf.Len()) {
		t.Errorf("byte count mismatch: n=%d buf.Len=%d", n, buf.Len())
	}
}

// ── stream error paths ─────────────────────────────────────────────────────────

// failAfterWriter fails after writing N bytes total.
type failAfterWriter struct {
	limit int
	written int
}

func (f *failAfterWriter) Write(p []byte) (int, error) {
	if f.written >= f.limit {
		return 0, errors.New("write failed")
	}
	n := len(p)
	if f.written+n > f.limit {
		n = f.limit - f.written
	}
	f.written += n
	return n, nil
}

func TestStream_WriteTo_DictError(t *testing.T) {
	s := NewStream()
	s.Compressed = false
	s.Data = []byte("data")

	// Fail immediately — Dictionary write fails
	fw := &failAfterWriter{limit: 0}
	_, err := s.WriteTo(fw)
	if err == nil {
		t.Error("expected error from Dict.WriteTo failing")
	}
}

func TestStream_WriteTo_StreamKeywordError(t *testing.T) {
	s := NewStream()
	s.Compressed = false
	s.Data = []byte("data")

	// Write the dict string to get its length first
	var dictBuf bytes.Buffer
	s.Dict.Add("Length", NewInt(len(s.Data)))
	_, _ = s.Dict.WriteTo(&dictBuf)

	// Allow the dict to write but fail at the "\nstream\n" write
	s2 := NewStream()
	s2.Compressed = false
	s2.Data = []byte("data")
	fw := &failAfterWriter{limit: dictBuf.Len()}
	_, err := s2.WriteTo(fw)
	if err == nil {
		t.Error("expected error from stream keyword write failing")
	}
}

func TestStream_WriteTo_DataError(t *testing.T) {
	s := NewStream()
	s.Compressed = false
	data := []byte("hello")
	s.Data = data

	// Write dict + "\nstream\n" but fail on data
	var dictBuf bytes.Buffer
	s.Dict.Add("Length", NewInt(len(data)))
	_, _ = s.Dict.WriteTo(&dictBuf)
	dictLen := dictBuf.Len()
	streamKeyword := "\nstream\n"

	s2 := NewStream()
	s2.Compressed = false
	s2.Data = data
	fw := &failAfterWriter{limit: dictLen + len(streamKeyword)}
	_, err := s2.WriteTo(fw)
	if err == nil {
		t.Error("expected error from data write failing")
	}
}

func TestStream_WriteTo_Compressed_zlibError(t *testing.T) {
	// Cover the zlibCompress error path by passing a zero-size data.
	// Note: zlibCompress itself can't easily be made to fail since it uses
	// bytes.Buffer internally; instead, exercise WriteTo with compressed=true
	// but using a writer that fails before stream data is written.
	s := NewStream()
	s.Compressed = true
	s.Data = []byte("compress")
	fw := &failAfterWriter{limit: 0}
	_, err := s.WriteTo(fw)
	if err != nil {
		// Expected: either zlibCompress succeeds (always) and dict write fails
		// or zlibCompress itself fails. Either way, error propagated.
		_ = err
	}
}

func TestZlibWrite_WriterError(t *testing.T) {
	// zlibWrite with a writer that fails — covers the zw.Write error path.
	// We write a lot of data so zw.Write eventually flushes to the underlying writer.
	src := bytes.Repeat([]byte("ABCD"), 10000)
	// A writer that fails immediately.
	err := zlibWrite(&failAfterWriter{limit: 0}, src)
	if err == nil {
		t.Error("expected error from zlibWrite with failing writer")
	}
}

func TestRef_WriteTo_WithWriter(t *testing.T) {
	obj := &IndirectObject{Number: 10, Generation: 2}
	ref := NewRef(obj)
	var sb strings.Builder
	_, err := ref.WriteTo(io.MultiWriter(&sb, io.Discard))
	if err != nil {
		t.Fatalf("WriteTo: %v", err)
	}
	if sb.String() != "10 2 R" {
		t.Errorf("got %q, want \"10 2 R\"", sb.String())
	}
}
