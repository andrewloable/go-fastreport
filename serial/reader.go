package serial

import (
	"encoding/xml"
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/andrewloable/go-fastreport/utils"
)

// Reader implements report.Reader and decodes objects from FRX XML.
//
// Usage pattern:
//
//	r := serial.NewReader(src)
//	typeName, ok := r.ReadObjectHeader()
//	// typeName is the XML element name (e.g. "Report")
//	// r.ReadStr / r.ReadInt / … read attributes of that element
//	// r.NextChild() advances to each nested child element
type Reader struct {
	dec *xml.Decoder

	// current element name (set by ReadObjectHeader / NextChild)
	curName string

	// attrs holds the attributes of the current element
	attrs map[string]string

	// depth tracks the nesting level of the element whose attributes are
	// in attrs.  When NextChild is called we scan forward at depth+1.
	depth int

	// childDepth is the depth of children we are currently iterating over.
	childDepth int

	// done signals that NextChild found an end element before any child start.
	done bool

	// skipped is set to true by SkipElement so that FinishChild knows it
	// does not need to call skipRemainingContent again.
	skipped bool

	// stack of saved states for nested NextChild calls (not needed for the
	// basic interface, but kept for correctness when readers are composed).
	stateStack []readerState

	// seenFirstChild tracks whether NextChild has been called at least once
	// for the current element.
	seenFirstChild bool
}

type readerState struct {
	curName        string
	attrs          map[string]string
	depth          int
	childDepth     int
	done           bool
	skipped        bool
	seenFirstChild bool
}

// NewReader creates a Reader that reads FRX XML from r.
func NewReader(r io.Reader) *Reader {
	return &Reader{
		dec:   xml.NewDecoder(r),
		attrs: make(map[string]string),
	}
}

// NewReaderWithPassword creates a Reader that transparently decrypts the
// stream if it begins with the FastReport "rij" encryption signature.
// If the stream is not encrypted the password is ignored.
// Returns an error only if the stream cannot be read or decryption fails.
func NewReaderWithPassword(r io.Reader, password string) (*Reader, bool, error) {
	plain, encrypted, err := utils.PeekAndDecrypt(r, password)
	if err != nil {
		return nil, false, err
	}
	return NewReader(plain), encrypted, nil
}

// ReadObjectHeader reads the next XML start element and returns its type name.
// It populates the internal attribute map so that ReadStr/ReadInt/… work.
// Returns ("", false) at EOF or on error.
func (r *Reader) ReadObjectHeader() (typeName string, ok bool) {
	for {
		tok, err := r.dec.Token()
		if err != nil {
			return "", false
		}
		switch t := tok.(type) {
		case xml.StartElement:
			r.curName = t.Name.Local
			r.attrs = attrsToMap(t.Attr)
			r.depth = int(r.dec.InputOffset()) // not the real depth; use a counter instead
			// We track depth via a separate counter maintained in the struct.
			// Reset child tracking.
			r.seenFirstChild = false
			r.done = false
			return r.curName, true
		case xml.EndElement:
			return "", false
		}
		// Skip other tokens (ProcInst, CharData, etc.)
	}
}

// ReadStr reads the named attribute, returning def if absent.
func (r *Reader) ReadStr(name, def string) string {
	v, ok := r.attrs[name]
	if !ok {
		return def
	}
	return v
}

// ReadInt reads the named attribute as an int, returning def if absent or
// unparseable.
func (r *Reader) ReadInt(name string, def int) int {
	v, ok := r.attrs[name]
	if !ok {
		return def
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return def
	}
	return n
}

// ReadBool reads the named attribute as a bool, returning def if absent.
// Accepts "true" / "1" as true; anything else as false.
func (r *Reader) ReadBool(name string, def bool) bool {
	v, ok := r.attrs[name]
	if !ok {
		return def
	}
	return v == "1" || strings.EqualFold(v, "true")
}

// ReadFloat reads the named attribute as a float32, returning def if absent or
// unparseable.
func (r *Reader) ReadFloat(name string, def float32) float32 {
	v, ok := r.attrs[name]
	if !ok {
		return def
	}
	f, err := strconv.ParseFloat(v, 32)
	if err != nil {
		return def
	}
	return float32(f)
}

// NextChild advances to the next immediate child element of the current object.
// It returns the child's type name and true, or ("", false) when there are no
// more children (or if the end tag of the parent is reached).
//
// After NextChild returns (typeName, true) the caller should read the child's
// attributes via ReadStr/ReadInt/… and recurse into its own children via
// another NextChild loop.  Call FinishChild when done with a child.
//
// NextChild satisfies report.Reader.
func (r *Reader) NextChild() (typeName string, ok bool) {
	for {
		tok, err := r.dec.Token()
		if err != nil {
			r.done = true
			return "", false
		}
		switch t := tok.(type) {
		case xml.StartElement:
			// Save state so FinishChild (or a nested call) can restore it.
			r.stateStack = append(r.stateStack, readerState{
				curName:        r.curName,
				attrs:          r.attrs,
				depth:          r.depth,
				childDepth:     r.childDepth,
				done:           r.done,
				skipped:        r.skipped,
				seenFirstChild: r.seenFirstChild,
			})
			r.curName = t.Name.Local
			r.attrs = attrsToMap(t.Attr)
			r.seenFirstChild = false
			r.done = false
			r.skipped = false
			return r.curName, true
		case xml.EndElement:
			// End of parent element — no more children.
			r.done = true
			return "", false
		}
		// Skip whitespace / comments.
	}
}

// FinishChild must be called after processing a child returned by NextChild.
// It skips any remaining content of the child element and restores the parent
// reader state so that NextChild can continue to the next sibling.
func (r *Reader) FinishChild() error {
	if len(r.stateStack) == 0 {
		return fmt.Errorf("serial: FinishChild called without matching NextChild")
	}
	// Skip any unread content inside the child, unless:
	//   - SkipElement already consumed the entire element (skipped=true), or
	//   - NextChild already consumed the parent's end element (done=true).
	// In the done=true case the end tag was consumed by the last NextChild
	// call that returned ("", false), so there is nothing left to skip.
	if !r.skipped && !r.done {
		if err := r.skipRemainingContent(); err != nil {
			return err
		}
	}
	// Restore parent state.
	top := r.stateStack[len(r.stateStack)-1]
	r.stateStack = r.stateStack[:len(r.stateStack)-1]
	r.curName = top.curName
	r.attrs = top.attrs
	r.depth = top.depth
	r.childDepth = top.childDepth
	r.done = top.done
	r.skipped = top.skipped
	r.seenFirstChild = top.seenFirstChild
	return nil
}

// SkipElement skips all remaining tokens until the matching end element for
// the element that was most recently started.  Used when the caller does not
// want to process a child.  After SkipElement, call FinishChild to restore
// the parent state.
func (r *Reader) SkipElement() error {
	if err := r.skipRemainingContent(); err != nil {
		return err
	}
	r.skipped = true
	return nil
}

// CurrentName returns the element name of the most recently read element.
func (r *Reader) CurrentName() string {
	return r.curName
}

// ── internal helpers ──────────────────────────────────────────────────────────

func attrsToMap(attrs []xml.Attr) map[string]string {
	m := make(map[string]string, len(attrs))
	for _, a := range attrs {
		m[a.Name.Local] = a.Value
	}
	return m
}

// skipRemainingContent reads and discards tokens until the end element that
// closes the current element depth is consumed.
func (r *Reader) skipRemainingContent() error {
	depth := 1
	for depth > 0 {
		tok, err := r.dec.Token()
		if err != nil {
			return err
		}
		switch tok.(type) {
		case xml.StartElement:
			depth++
		case xml.EndElement:
			depth--
		}
	}
	return nil
}
