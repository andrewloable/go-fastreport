package pdf

import (
	"bytes"

	"github.com/andrewloable/go-fastreport/export/pdf/core"
)

// Contents is the page content stream.
// It accumulates PDF graphics operators in an internal buffer and, when
// Finalize is called, commits the data to the underlying PDF Stream object.
type Contents struct {
	obj    *core.IndirectObject
	stream *core.Stream
	buf    bytes.Buffer
}

// NewContents creates an uncompressed content stream and registers it with the
// writer.  The stream is not compressed because PDF readers typically decompress
// content streams on the fly and compression is optional for conformance.
func NewContents(w *Writer) *Contents {
	stream := core.NewStream()
	stream.Compressed = false

	obj := w.NewObject(stream)
	return &Contents{obj: obj, stream: stream}
}

// Write appends raw PDF graphics operators to the content stream buffer.
func (c *Contents) Write(data []byte) {
	c.buf.Write(data)
}

// WriteString appends a string of PDF operators to the content stream buffer.
func (c *Contents) WriteString(s string) {
	c.buf.WriteString(s)
}

// Finalize commits the buffered content to the PDF stream object.
// Call this after all graphics operators have been appended, before writing
// the document with Writer.Write.
func (c *Contents) Finalize() {
	c.stream.Data = make([]byte, c.buf.Len())
	copy(c.stream.Data, c.buf.Bytes())
}
