package core

import (
	"fmt"
	"io"
)

// IndirectObject wraps an Object with a PDF object number and generation
// number, producing the "N G obj … endobj" block used in PDF cross-reference
// tables.
//
// Example output for Number=3, Generation=0:
//
//	3 0 obj
//	…value…
//	endobj
type IndirectObject struct {
	// Number is the 1-based object number assigned by the PDF writer.
	Number int
	// Generation is almost always 0 for newly created objects.
	Generation int
	// Value is the PDF object contained in this indirect wrapper.
	Value Object
}

// Type implements Object.
func (o *IndirectObject) Type() ObjectType { return TypeIndirect }

// WriteTo writes the complete indirect-object block to w.
// It satisfies io.WriterTo so that IndirectObject can be used wherever an
// Object is expected.
func (o *IndirectObject) WriteTo(w io.Writer) (int64, error) {
	cw := &countWriter{w: w}
	_, err := fmt.Fprintf(cw, "%d %d obj\n", o.Number, o.Generation)
	if err != nil {
		return cw.n, err
	}
	if o.Value != nil {
		if _, err = o.Value.WriteTo(cw); err != nil {
			return cw.n, err
		}
	}
	_, err = fmt.Fprint(cw, "\nendobj\n")
	return cw.n, err
}

// Reference writes the indirect reference token "N G R" for this object.
// This is a convenience helper for callers that need to refer to the object
// from elsewhere in the PDF without embedding the full indirect-object block.
func (o *IndirectObject) Reference() string {
	return fmt.Sprintf("%d %d R", o.Number, o.Generation)
}

// countWriter is a thin wrapper around an io.Writer that accumulates the
// total number of bytes written so that WriteTo can return an int64.
type countWriter struct {
	w io.Writer
	n int64
}

func (cw *countWriter) Write(p []byte) (int, error) {
	n, err := cw.w.Write(p)
	cw.n += int64(n)
	return n, err
}
