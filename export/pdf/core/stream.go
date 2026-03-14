package core

import (
	"bytes"
	"compress/zlib"
	"fmt"
	"io"
)

// Stream is a PDF stream object: a Dictionary followed by a block of binary
// data delimited by "stream" and "endstream" keywords.
//
// When Compressed is true the Data payload is deflated with zlib before
// writing and the /Filter /FlateDecode and /Length entries are set
// automatically.  When Compressed is false the raw Data bytes are written
// and only /Length is set.
//
// PDF representation (compressed):
//
//	<< /Filter /FlateDecode /Length N >>
//	stream
//	…compressed bytes…
//	endstream
type Stream struct {
	// Dict holds auxiliary dictionary entries (e.g. /Type, /Subtype).
	// /Length and, when Compressed is true, /Filter are managed automatically.
	Dict *Dictionary
	// Data is the raw (uncompressed) stream content.
	Data []byte
	// Compressed controls whether the data is zlib-compressed on output.
	Compressed bool
}

// NewStream returns a Stream with an empty Dictionary and Compressed set to
// true, matching the default behaviour of the original C# implementation.
func NewStream() *Stream {
	return &Stream{
		Dict:       NewDictionary(),
		Compressed: true,
	}
}

// Type implements Object.
func (s *Stream) Type() ObjectType { return TypeStream }

// WriteTo writes the complete stream object (dictionary + data) to w.
func (s *Stream) WriteTo(w io.Writer) (int64, error) {
	data := s.Data
	if data == nil {
		data = []byte{}
	}

	if s.Compressed {
		compressed, err := zlibCompress(data)
		if err != nil {
			return 0, err
		}
		data = compressed
		s.Dict.Add("Filter", NewName("FlateDecode"))
	}

	s.Dict.Add("Length", NewInt(len(data)))

	cw := &countWriter{w: w}

	if _, err := s.Dict.WriteTo(cw); err != nil {
		return cw.n, err
	}
	if _, err := fmt.Fprint(cw, "\nstream\n"); err != nil {
		return cw.n, err
	}
	if _, err := cw.Write(data); err != nil {
		return cw.n, err
	}
	_, err := fmt.Fprint(cw, "\nendstream\n")
	return cw.n, err
}

// zlibCompress compresses src using zlib (RFC 1950) at the default level and
// returns the compressed bytes.
func zlibCompress(src []byte) ([]byte, error) {
	var buf bytes.Buffer
	if err := zlibWrite(&buf, src); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// zlibWrite deflates src into dst using the zlib format.  It is split out
// from zlibCompress so that tests can inject a failing writer to exercise
// the error-handling branches.
func zlibWrite(dst io.Writer, src []byte) error {
	zw, err := zlib.NewWriterLevel(dst, zlib.DefaultCompression)
	if err != nil {
		return err
	}
	if _, err = zw.Write(src); err != nil {
		_ = zw.Close()
		return err
	}
	return zw.Close()
}
