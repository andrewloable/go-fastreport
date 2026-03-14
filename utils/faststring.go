package utils

import "strings"

// FastString is an optimized string builder that wraps [strings.Builder].
// It is the Go equivalent of FastReport's FastString class and provides a
// fluent API for building strings incrementally without repeated allocations.
type FastString struct {
	sb strings.Builder
}

// NewFastString creates a new empty FastString.
func NewFastString() *FastString {
	return &FastString{}
}

// Append appends s to the builder and returns the receiver for chaining.
func (fs *FastString) Append(s string) *FastString {
	fs.sb.WriteString(s)
	return fs
}

// AppendLine appends s followed by a newline character ("\n") and returns
// the receiver for chaining.
func (fs *FastString) AppendLine(s string) *FastString {
	fs.sb.WriteString(s)
	fs.sb.WriteByte('\n')
	return fs
}

// AppendRune appends the Unicode code point r and returns the receiver for
// chaining.
func (fs *FastString) AppendRune(r rune) *FastString {
	fs.sb.WriteRune(r)
	return fs
}

// String returns the accumulated string.
func (fs *FastString) String() string {
	return fs.sb.String()
}

// Len returns the number of bytes currently accumulated.
func (fs *FastString) Len() int {
	return fs.sb.Len()
}

// Reset clears the builder so it can be reused without allocating a new one.
func (fs *FastString) Reset() {
	fs.sb.Reset()
}

// IsEmpty reports whether the builder contains no characters.
func (fs *FastString) IsEmpty() bool {
	return fs.sb.Len() == 0
}
