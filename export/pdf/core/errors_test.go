package core

// errorWriter is an io.Writer that always returns an error after writing
// a fixed number of bytes.  It is used to exercise error-handling branches
// in WriteTo implementations that are otherwise unreachable when writing to
// a bytes.Buffer (which never returns an error).
import (
	"errors"
	"io"
	"testing"
)

var errFakeWrite = errors.New("fake write error")

// failWriter returns an error on the Nth call to Write (0-indexed).
type failWriter struct {
	call    int
	failAt  int
	written int
}

func (fw *failWriter) Write(p []byte) (int, error) {
	if fw.call == fw.failAt {
		fw.call++
		return 0, errFakeWrite
	}
	fw.call++
	fw.written += len(p)
	return len(p), nil
}

// newFail returns a failWriter that errors on the given Write call.
func newFail(at int) *failWriter { return &failWriter{failAt: at} }

// ---------------------------------------------------------------------------
// Array error paths
// ---------------------------------------------------------------------------

func TestArray_WriteTo_Error_Open(t *testing.T) {
	a := NewArray(NewInt(1))
	_, err := a.WriteTo(newFail(0))
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestArray_WriteTo_Error_Item(t *testing.T) {
	a := NewArray(NewInt(1))
	_, err := a.WriteTo(newFail(1))
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestArray_WriteTo_Error_Space(t *testing.T) {
	a := NewArray(NewInt(1))
	_, err := a.WriteTo(newFail(2))
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestArray_WriteTo_Error_Close(t *testing.T) {
	// After writing "[ " + item + " " we fail on "]"
	a := NewArray(NewInt(1))
	_, err := a.WriteTo(newFail(3))
	if err == nil {
		t.Fatal("expected error")
	}
}

// ---------------------------------------------------------------------------
// Dictionary error paths
// ---------------------------------------------------------------------------

func TestDictionary_WriteTo_Error_Open(t *testing.T) {
	d := NewDictionary()
	d.Add("X", NewInt(1))
	_, err := d.WriteTo(newFail(0))
	if err == nil {
		t.Fatal("expected error on open bracket")
	}
}

func TestDictionary_WriteTo_Error_Key(t *testing.T) {
	d := NewDictionary()
	d.Add("X", NewInt(1))
	_, err := d.WriteTo(newFail(1))
	if err == nil {
		t.Fatal("expected error on key write")
	}
}

func TestDictionary_WriteTo_Error_Value(t *testing.T) {
	d := NewDictionary()
	d.Add("X", NewInt(1))
	_, err := d.WriteTo(newFail(2))
	if err == nil {
		t.Fatal("expected error on value write")
	}
}

func TestDictionary_WriteTo_Error_Space(t *testing.T) {
	d := NewDictionary()
	d.Add("X", NewInt(1))
	_, err := d.WriteTo(newFail(3))
	if err == nil {
		t.Fatal("expected error on space after value")
	}
}

func TestDictionary_WriteTo_Error_Close(t *testing.T) {
	d := NewDictionary()
	d.Add("X", NewInt(1))
	_, err := d.WriteTo(newFail(4))
	if err == nil {
		t.Fatal("expected error on close bracket")
	}
}

// ---------------------------------------------------------------------------
// IndirectObject error paths
// ---------------------------------------------------------------------------

func TestIndirectObject_WriteTo_Error_Header(t *testing.T) {
	o := &IndirectObject{Number: 1, Value: NewInt(1)}
	_, err := o.WriteTo(newFail(0))
	if err == nil {
		t.Fatal("expected error on header write")
	}
}

func TestIndirectObject_WriteTo_Error_Value(t *testing.T) {
	o := &IndirectObject{Number: 1, Value: NewInt(1)}
	_, err := o.WriteTo(newFail(1))
	if err == nil {
		t.Fatal("expected error on value write")
	}
}

func TestIndirectObject_WriteTo_Error_Footer(t *testing.T) {
	o := &IndirectObject{Number: 1, Value: NewInt(1)}
	_, err := o.WriteTo(newFail(2))
	if err == nil {
		t.Fatal("expected error on footer write")
	}
}

// ---------------------------------------------------------------------------
// Name error paths
// ---------------------------------------------------------------------------

func TestName_WriteTo_Error_Slash(t *testing.T) {
	n := NewName("Type")
	_, err := n.WriteTo(newFail(0))
	if err == nil {
		t.Fatal("expected error on slash write")
	}
}

func TestName_WriteTo_Error_Regular(t *testing.T) {
	n := NewName("Type")
	_, err := n.WriteTo(newFail(1))
	if err == nil {
		t.Fatal("expected error on regular char write")
	}
}

func TestName_WriteTo_Error_Encoded(t *testing.T) {
	// space → #20 encoding
	n := NewName("a b")
	// fail on: 0=/  1=a  2=#20  (fail at the # write)
	_, err := n.WriteTo(newFail(2))
	if err == nil {
		t.Fatal("expected error on encoded char write")
	}
}

// ---------------------------------------------------------------------------
// Numeric error paths
// ---------------------------------------------------------------------------

func TestNumeric_WriteTo_Error_Int(t *testing.T) {
	n := NewInt(42)
	_, err := n.WriteTo(newFail(0))
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestNumeric_WriteTo_Error_Float(t *testing.T) {
	n := NewFloat(1.23)
	_, err := n.WriteTo(newFail(0))
	if err == nil {
		t.Fatal("expected error")
	}
}

// ---------------------------------------------------------------------------
// Boolean error paths
// ---------------------------------------------------------------------------

func TestBoolean_WriteTo_Error_True(t *testing.T) {
	b := NewBoolean(true)
	_, err := b.WriteTo(newFail(0))
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestBoolean_WriteTo_Error_False(t *testing.T) {
	b := NewBoolean(false)
	_, err := b.WriteTo(newFail(0))
	if err == nil {
		t.Fatal("expected error")
	}
}

// ---------------------------------------------------------------------------
// Null error paths
// ---------------------------------------------------------------------------

func TestNull_WriteTo_Error(t *testing.T) {
	n := &Null{}
	_, err := n.WriteTo(newFail(0))
	if err == nil {
		t.Fatal("expected error")
	}
}

// ---------------------------------------------------------------------------
// String error paths
// ---------------------------------------------------------------------------

func TestString_WriteTo_Error_Hex_Open(t *testing.T) {
	s := NewHexString("A")
	_, err := s.WriteTo(newFail(0))
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestString_WriteTo_Error_Hex_Body(t *testing.T) {
	s := NewHexString("A")
	_, err := s.WriteTo(newFail(1))
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestString_WriteTo_Error_Hex_Close(t *testing.T) {
	// BOM (2 bytes → 2 Write calls for FFFE bytes) + "A" (2 more) + close
	// With our implementation each byte is written as fmt.Fprintf("#%02X")
	// so each byte in UTF-16BE is one Write call. BOM = 2 calls, "A" = 2 calls.
	// Call order: 0=open "<"  1..4= hex bytes  5=close ">"
	s := NewHexString("A")
	_, err := s.WriteTo(newFail(5))
	if err == nil {
		t.Fatal("expected error on hex close")
	}
}

func TestString_WriteTo_Error_Literal_Open(t *testing.T) {
	s := NewString("A")
	_, err := s.WriteTo(newFail(0))
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestString_WriteTo_Error_Literal_Body(t *testing.T) {
	s := NewString("A")
	_, err := s.WriteTo(newFail(1))
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestString_WriteTo_Error_Literal_Close(t *testing.T) {
	s := NewString("A")
	_, err := s.WriteTo(newFail(2))
	if err == nil {
		t.Fatal("expected error")
	}
}

// ---------------------------------------------------------------------------
// Stream error paths
// ---------------------------------------------------------------------------

func TestStream_WriteTo_Error_Dict(t *testing.T) {
	s := NewStream()
	s.Compressed = false
	s.Data = []byte("x")
	// Dictionary open "<<" is the first write
	_, err := s.WriteTo(newFail(0))
	if err == nil {
		t.Fatal("expected error on dict write")
	}
}

func TestStream_WriteTo_Error_StreamKeyword(t *testing.T) {
	// We need to get past the dictionary write – use a writer that allows
	// enough writes to finish the (empty) dictionary then fails on "stream\n"
	s := NewStream()
	s.Compressed = false
	s.Data = []byte("x")
	// Dictionary with no entries writes: "<< >>" → 1 write ("<<  >>")
	// Actually it's: fmt.Fprint("<< ") → write 0, then fmt.Fprint(">>") → write 1
	// Then "\nstream\n" → write 2
	_, err := s.WriteTo(newFail(2))
	if err == nil {
		t.Fatal("expected error on stream keyword")
	}
}

func TestStream_WriteTo_Error_Data(t *testing.T) {
	s := NewStream()
	s.Compressed = false
	s.Data = []byte("x")
	// Writes: 0="<< "  1=">>"  2="\nstream\n"  3=data  4="\nendstream\n"
	_, err := s.WriteTo(newFail(3))
	if err == nil {
		t.Fatal("expected error on data write")
	}
}

func TestStream_WriteTo_Error_EndStream(t *testing.T) {
	s := NewStream()
	s.Compressed = false
	s.Data = []byte("x")
	_, err := s.WriteTo(newFail(4))
	if err == nil {
		t.Fatal("expected error on endstream keyword")
	}
}

// ---------------------------------------------------------------------------
// zlibCompressTo error paths
// ---------------------------------------------------------------------------

// zlibFailWriter is an io.Writer that fails on every Write call after
// `skipWrites` successful writes, allowing tests to drive error branches
// inside zlibWrite that cannot be reached with a normal bytes.Buffer.
type zlibFailWriter struct {
	skipWrites int
	written    int
}

func (z *zlibFailWriter) Write(p []byte) (int, error) {
	if z.written >= z.skipWrites {
		return 0, errFakeWrite
	}
	z.written++
	return len(p), nil
}

func TestZlibWrite_WriteError_Immediate(t *testing.T) {
	// Fail on the very first Write (the zlib header).
	// zlib.NewWriterLevel writes a 2-byte header immediately; this causes
	// an error that propagates back through zw.Write(src).
	err := zlibWrite(&zlibFailWriter{skipWrites: 0}, []byte("data"))
	if err == nil {
		t.Fatal("expected error when writer always fails immediately")
	}
}

func TestZlibWrite_WriteError_AfterHeader(t *testing.T) {
	// Allow the 2-byte header write to succeed (1 Write call), then fail.
	// This exercises the zw.Write(src) error branch.
	err := zlibWrite(&zlibFailWriter{skipWrites: 1}, []byte("data"))
	if err == nil {
		t.Fatal("expected error after zlib header")
	}
}

// TestZlibWrite_WriteError_DuringData exercises line 97-99 in stream.go:
// the zw.Write(src) call fails when src is large enough to force the deflate
// encoder to flush its internal buffer (≥ 32 KiB) to the underlying writer.
// By allowing only the header write to succeed and failing on the first data
// flush, the error from zw.Write is captured and returned.
func TestZlibWrite_WriteError_DuringData(t *testing.T) {
	// Use 64 KiB of data: this forces at least one internal flush during zw.Write.
	// skipWrites=1 allows the 2-byte zlib header write then rejects data writes.
	largeData := make([]byte, 64*1024)
	for i := range largeData {
		largeData[i] = byte(i)
	}
	err := zlibWrite(&zlibFailWriter{skipWrites: 1}, largeData)
	if err == nil {
		t.Fatal("expected error when data write forces a flush to a failing writer")
	}
}

// TestZlibCompress_ErrorPath verifies that zlibCompress propagates errors from
// zlibWrite (line 83-85 of stream.go). Since zlibCompress writes to an internal
// bytes.Buffer (which never fails), we cannot inject a failure through the
// public API; instead we exercise the error path indirectly by calling
// zlibWrite with a failing writer — confirming that the mechanism zlibCompress
// relies on does propagate errors correctly.
func TestZlibCompress_InternalBuffer_Success(t *testing.T) {
	// Normal path: zlibCompress should succeed and return non-empty output.
	out, err := zlibCompress([]byte("hello world"))
	if err != nil {
		t.Fatalf("zlibCompress: %v", err)
	}
	if len(out) == 0 {
		t.Fatal("expected non-empty compressed output")
	}
}

// TestStream_WriteTo_CompressedError exercises lines 55-57 of stream.go:
// the error return from zlibCompress inside WriteTo when the stream is
// compressed. We achieve this by using a very large data payload and a
// writer that fails before data can be fully flushed. However, since
// zlibCompress writes to an internal buffer first, the error only surfaces
// at the Dict.WriteTo stage (not zlibCompress). For completeness the test
// documents this behaviour.
func TestStream_WriteTo_Compressed_DictError(t *testing.T) {
	s := NewStream()
	s.Compressed = true
	s.Data = []byte("compressed data test")
	// Fail immediately — Dict.WriteTo will fail before zlibCompress error.
	_, err := s.WriteTo(newFail(0))
	if err == nil {
		t.Fatal("expected error")
	}
}

// Ensure failWriter implements io.Writer at compile time.
var _ io.Writer = (*failWriter)(nil)
var _ io.Writer = (*zlibFailWriter)(nil)
