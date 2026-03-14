// Package serial provides FRX serialization and deserialization utilities for go-fastreport.
package serial

import (
	"encoding/xml"
	"fmt"
	"io"
	"strconv"

	"github.com/andrewloable/go-fastreport/report"
)

// TypeNamer is an optional interface that report objects may implement to
// provide their FRX element name during serialization.  If an object does not
// implement TypeNamer, WriteObject falls back to the Go type name (without the
// package prefix).
type TypeNamer interface {
	TypeName() string
}

// Writer implements report.Writer and encodes objects as FRX XML.
//
// FRX format: each object is an XML element whose tag is the object's type
// name; properties are XML attributes on that element; child objects are
// nested elements.
//
// Example output:
//
//	<?xml version="1.0" encoding="utf-8"?>
//	<Report>
//	  <ReportPage Name="Page1" Width="850" Height="1169">
//	    <DataBand Name="Data1" Height="50">
//	      <TextObject Name="Text1" Left="0" Top="0" Width="200" Height="30" Text="Hello"/>
//	    </DataBand>
//	  </ReportPage>
//	</Report>
//
// The writer satisfies report.Writer.
type Writer struct {
	enc   *xml.Encoder
	w     io.Writer
	stack []elementState // open element stack
}

// elementState tracks an open XML element so that attributes can be written
// before any child elements are emitted and the element can be properly closed.
type elementState struct {
	name   string
	attrs  []xml.Attr
	flushed bool // true once the start element has been written to enc
}

// NewWriter creates a Writer that writes FRX XML to w.
func NewWriter(w io.Writer) *Writer {
	enc := xml.NewEncoder(w)
	enc.Indent("", "  ")
	return &Writer{enc: enc, w: w}
}

// WriteHeader writes the XML processing instruction header.
// Call this once before the first WriteObject / BeginObject.
func (w *Writer) WriteHeader() error {
	_, err := io.WriteString(w.w, `<?xml version="1.0" encoding="utf-8"?>`)
	return err
}

// BeginObject opens a new XML element with the given type name.
// All subsequent WriteStr/WriteInt/WriteBool/WriteFloat calls accumulate
// as attributes on this element until EndObject or a nested BeginObject is
// called, at which point the start tag is flushed.
func (w *Writer) BeginObject(name string) error {
	// Flush any un-flushed ancestor start tag before we can nest.
	if err := w.flushPending(); err != nil {
		return err
	}
	w.stack = append(w.stack, elementState{name: name})
	return nil
}

// EndObject closes the innermost open element.
func (w *Writer) EndObject() error {
	if len(w.stack) == 0 {
		return fmt.Errorf("serial: EndObject called with empty stack")
	}
	top := &w.stack[len(w.stack)-1]
	if !top.flushed {
		// Element has no children — emit a self-closing start tag then end.
		if err := w.enc.EncodeToken(xml.StartElement{
			Name:  xml.Name{Local: top.name},
			Attr:  top.attrs,
		}); err != nil {
			return err
		}
	}
	if err := w.enc.EncodeToken(xml.EndElement{Name: xml.Name{Local: top.name}}); err != nil {
		return err
	}
	w.stack = w.stack[:len(w.stack)-1]
	return nil
}

// WriteObject writes obj as a child XML element.
// The element name is determined by:
//  1. obj implementing TypeNamer, or
//  2. the bare Go type name (last component after any ".").
//
// WriteObject satisfies report.Writer.
func (w *Writer) WriteObject(obj report.Serializable) error {
	name := typeNameOf(obj)
	if err := w.BeginObject(name); err != nil {
		return err
	}
	if err := obj.Serialize(w); err != nil {
		return err
	}
	return w.EndObject()
}

// WriteObjectNamed writes obj as a child element with an explicit element name.
// Use this when the caller knows the element name independently of the object
// (e.g., the top-level "Report" element).
func (w *Writer) WriteObjectNamed(name string, obj report.Serializable) error {
	if err := w.BeginObject(name); err != nil {
		return err
	}
	if err := obj.Serialize(w); err != nil {
		return err
	}
	return w.EndObject()
}

// WriteStr writes a named string attribute on the current element.
func (w *Writer) WriteStr(name, value string) {
	w.addAttr(name, value)
}

// WriteInt writes a named int attribute on the current element.
func (w *Writer) WriteInt(name string, value int) {
	w.addAttr(name, strconv.Itoa(value))
}

// WriteBool writes a named bool attribute on the current element.
func (w *Writer) WriteBool(name string, value bool) {
	if value {
		w.addAttr(name, "true")
	} else {
		w.addAttr(name, "false")
	}
}

// WriteFloat writes a named float32 attribute on the current element.
// The value is formatted with the invariant (English) decimal point, matching
// the C# CultureInfo.InvariantCulture formatting used in FRWriter.
func (w *Writer) WriteFloat(name string, value float32) {
	w.addAttr(name, strconv.FormatFloat(float64(value), 'f', -1, 32))
}

// Flush flushes any buffered XML output to the underlying writer.
func (w *Writer) Flush() error {
	return w.enc.Flush()
}

// ── internal helpers ──────────────────────────────────────────────────────────

// addAttr appends an XML attribute to the current open element.
// If there is no open element the call is a no-op (defensive).
func (w *Writer) addAttr(name, value string) {
	if len(w.stack) == 0 {
		return
	}
	top := &w.stack[len(w.stack)-1]
	top.attrs = append(top.attrs, xml.Attr{
		Name:  xml.Name{Local: name},
		Value: value,
	})
}

// flushPending writes the StartElement token for the top-of-stack element if
// it has not been written yet.  This is called whenever a child element is
// about to be opened so that the parent's start tag is emitted first.
func (w *Writer) flushPending() error {
	if len(w.stack) == 0 {
		return nil
	}
	top := &w.stack[len(w.stack)-1]
	if top.flushed {
		return nil
	}
	if err := w.enc.EncodeToken(xml.StartElement{
		Name: xml.Name{Local: top.name},
		Attr: top.attrs,
	}); err != nil {
		return err
	}
	top.flushed = true
	return nil
}

// typeNameOf returns the FRX element name for obj.
func typeNameOf(obj report.Serializable) string {
	if tn, ok := obj.(TypeNamer); ok {
		return tn.TypeName()
	}
	// Fall back to bare Go type name.
	name := fmt.Sprintf("%T", obj)
	// Strip pointer and package prefix: "*mypkg.Foo" → "Foo".
	for i := len(name) - 1; i >= 0; i-- {
		if name[i] == '.' {
			return name[i+1:]
		}
	}
	return name
}
