// Package utils provides shared utility types and helpers used throughout go-fastreport.
package utils

import "fmt"

// DuplicateNameError is returned when an object with the same name already exists.
type DuplicateNameError struct {
	Name string
}

func (e *DuplicateNameError) Error() string {
	return fmt.Sprintf("duplicate name: %q", e.Name)
}

// AncestorError is returned when trying to rename an object from an ancestor report.
type AncestorError struct {
	Name string
}

func (e *AncestorError) Error() string {
	return fmt.Sprintf("cannot rename ancestor object: %q", e.Name)
}

// FileFormatError is returned when loading a malformed FRX report file.
type FileFormatError struct {
	Detail string
}

func (e *FileFormatError) Error() string {
	if e.Detail != "" {
		return "wrong file format: " + e.Detail
	}
	return "wrong file format"
}

// DecryptError is returned when loading an encrypted report with the wrong password.
type DecryptError struct{}

func (e *DecryptError) Error() string {
	return "decrypt error: wrong password"
}

// CompilerErrorInfo holds a single expression compiler error location.
type CompilerErrorInfo struct {
	Line         int
	Column       int
	ReportObject string
	Message      string
}

// CompilerError is returned when an expression cannot be compiled.
type CompilerError struct {
	Msg    string
	Errors []CompilerErrorInfo
}

func (e *CompilerError) Error() string {
	return e.Msg
}

// ParentError is returned when a child object cannot be added to a parent.
type ParentError struct {
	ParentType string
	ChildType  string
}

func (e *ParentError) Error() string {
	return fmt.Sprintf("cannot add %q to parent %q", e.ChildType, e.ParentType)
}

// ClassError is returned when deserializing an unknown object type.
type ClassError struct {
	Name string
}

func (e *ClassError) Error() string {
	return fmt.Sprintf("cannot find class: %q", e.Name)
}

// DataTableError is returned when a table data source is not properly configured.
type DataTableError struct {
	Alias string
}

func (e *DataTableError) Error() string {
	return fmt.Sprintf("%s: table is null or not configured", e.Alias)
}

// DataNotInitializedError is returned when accessing a data source that has not been initialized.
type DataNotInitializedError struct {
	Alias string
}

func (e *DataNotInitializedError) Error() string {
	return fmt.Sprintf("%s: data source is not initialized", e.Alias)
}

// NotValidIdentifierError is returned when an object name is not a valid identifier.
type NotValidIdentifierError struct {
	Value string
}

func (e *NotValidIdentifierError) Error() string {
	return fmt.Sprintf("%q is not a valid identifier name", e.Value)
}

// UnknownNameError is returned when an unknown name is supplied.
type UnknownNameError struct {
	Value string
}

func (e *UnknownNameError) Error() string {
	return fmt.Sprintf("unknown name: %q", e.Value)
}

// GroupHeaderNoConditionError is returned when a GroupHeader has no group condition.
type GroupHeaderNoConditionError struct {
	Name string
}

func (e *GroupHeaderNoConditionError) Error() string {
	return fmt.Sprintf("group header %q has no group condition", e.Name)
}

// ImageLoadError is returned when an image cannot be loaded.
type ImageLoadError struct {
	Cause error
}

func (e *ImageLoadError) Error() string {
	return fmt.Sprintf("image load error: %v", e.Cause)
}

func (e *ImageLoadError) Unwrap() error {
	return e.Cause
}
