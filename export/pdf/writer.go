package pdf

import (
	"bytes"
	"fmt"
	"io"

	"github.com/andrewloable/go-fastreport/export/pdf/core"
)

// Writer manages PDF file generation.
// It tracks all PDF indirect objects, builds the cross-reference table,
// and writes the complete PDF binary to an io.Writer.
type Writer struct {
	nextObjNum int
	objects    []*core.IndirectObject
	catalog    *core.IndirectObject
	info       *core.IndirectObject
}

// NewWriter creates a new Writer with object numbering starting at 1.
func NewWriter() *Writer {
	return &Writer{nextObjNum: 1}
}

// NewObject creates a new numbered IndirectObject and registers it with the writer.
func (w *Writer) NewObject(value core.Object) *core.IndirectObject {
	obj := &core.IndirectObject{
		Number:     w.nextObjNum,
		Generation: 0,
		Value:      value,
	}
	w.nextObjNum++
	w.objects = append(w.objects, obj)
	return obj
}

// setCatalog marks an indirect object as the PDF catalog (Root).
func (w *Writer) setCatalog(obj *core.IndirectObject) {
	w.catalog = obj
}

// setInfo marks an indirect object as the PDF info dictionary.
func (w *Writer) setInfo(obj *core.IndirectObject) {
	w.info = obj
}

// Write generates the complete PDF file to out.
// Format: %PDF-1.4\n + binary comment + objects + xref table + trailer
func (w *Writer) Write(out io.Writer) error {
	var buf bytes.Buffer

	// PDF header
	fmt.Fprint(&buf, "%PDF-1.4\n")
	// Binary comment hint (bytes > 127 tell tools this is a binary file)
	buf.Write([]byte{0x25, 0xE2, 0xE3, 0xCF, 0xD3, 0x0D, 0x0A})

	// Write each object and record its byte offset
	offsets := make([]int64, len(w.objects))
	for i, obj := range w.objects {
		offsets[i] = int64(buf.Len())
		if _, err := obj.WriteTo(&buf); err != nil {
			return fmt.Errorf("writing object %d: %w", obj.Number, err)
		}
	}

	// Cross-reference table
	xrefOffset := int64(buf.Len())
	fmt.Fprint(&buf, "xref\n")
	fmt.Fprintf(&buf, "0 %d\n", len(w.objects)+1)
	// Free entry for object 0
	fmt.Fprint(&buf, "0000000000 65535 f\r\n")
	for _, offset := range offsets {
		fmt.Fprintf(&buf, "%010d 00000 n\r\n", offset)
	}

	// Trailer dictionary — write inline using raw PDF syntax
	size := len(w.objects) + 1
	fmt.Fprint(&buf, "trailer\n<< ")
	fmt.Fprintf(&buf, "/Size %d ", size)
	if w.catalog != nil {
		fmt.Fprintf(&buf, "/Root %s ", w.catalog.Reference())
	}
	if w.info != nil {
		fmt.Fprintf(&buf, "/Info %s ", w.info.Reference())
	}
	fmt.Fprint(&buf, ">>")

	fmt.Fprintf(&buf, "\nstartxref\n%d\n%%%%EOF\n", xrefOffset)

	_, err := out.Write(buf.Bytes())
	return err
}
