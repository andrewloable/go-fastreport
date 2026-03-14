package core

import (
	"fmt"
	"io"
)

// dictEntry stores a single key/value pair in insertion order.
type dictEntry struct {
	key   string
	value Object
}

// Dictionary is a PDF dictionary object.  Keys are PDF Name values (without
// the leading slash) and values are any PDF Object.  Entries are iterated in
// insertion order, which ensures deterministic output.
//
// PDF representation:
//
//	<< /Key1 Value1 /Key2 Value2 >>
type Dictionary struct {
	entries []dictEntry
	index   map[string]int // key → position in entries slice
}

// NewDictionary returns an empty, ready-to-use Dictionary.
func NewDictionary() *Dictionary {
	return &Dictionary{index: make(map[string]int)}
}

// Type implements Object.
func (d *Dictionary) Type() ObjectType { return TypeDictionary }

// Add inserts or replaces the entry with the given key.  The key must be
// supplied without a leading slash (e.g. "Type", not "/Type").
// Add returns the receiver so calls can be chained.
func (d *Dictionary) Add(key string, value Object) *Dictionary {
	if i, ok := d.index[key]; ok {
		d.entries[i].value = value
		return d
	}
	d.index[key] = len(d.entries)
	d.entries = append(d.entries, dictEntry{key: key, value: value})
	return d
}

// Get returns the value for key, or nil if the key is not present.
func (d *Dictionary) Get(key string) Object {
	if i, ok := d.index[key]; ok {
		return d.entries[i].value
	}
	return nil
}

// Len returns the number of entries in the dictionary.
func (d *Dictionary) Len() int { return len(d.entries) }

// WriteTo writes the PDF dictionary representation to w.
func (d *Dictionary) WriteTo(w io.Writer) (int64, error) {
	cw := &countWriter{w: w}
	if _, err := fmt.Fprint(cw, "<< "); err != nil {
		return cw.n, err
	}
	for _, e := range d.entries {
		// Write /Key
		if _, err := fmt.Fprintf(cw, "/%s ", e.key); err != nil {
			return cw.n, err
		}
		// Write value
		if _, err := e.value.WriteTo(cw); err != nil {
			return cw.n, err
		}
		if _, err := fmt.Fprint(cw, " "); err != nil {
			return cw.n, err
		}
	}
	_, err := fmt.Fprint(cw, ">>")
	return cw.n, err
}
