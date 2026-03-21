package data

import "time"

// System variable name constants mirror FastReport.Data.SystemVariables (SystemVariables.cs).
// Each constant holds the exact Name string that the C# variable class uses.
const (
	// SysVarPage is the current page number ("Page" — C# PageVariable.Name).
	SysVarPage = "Page"
	// SysVarPageNumber is a Go-idiomatic alias for SysVarPage ("PageNumber").
	SysVarPageNumber = "PageNumber"
	// SysVarTotalPages is the total number of pages ("TotalPages" — C# TotalPagesVariable.Name).
	SysVarTotalPages = "TotalPages"
	// SysVarPageCount is an alternative alias for TotalPages kept for backward compatibility.
	SysVarPageCount = "PageCount"
	// SysVarPageN is the "Page N" string ("PageN" — C# PageNVariable.Name).
	SysVarPageN = "PageN"
	// SysVarPageNofM is the "Page N of M" string ("PageNofM" — C# PageNofMVariable.Name).
	SysVarPageNofM = "PageNofM"
	// SysVarDate is the report-run date ("Date" — C# DateVariable.Name).
	SysVarDate = "Date"
	// SysVarTime is the report-run time ("Time" — not directly in C# but
	// used by the Go port for the time component of Date).
	SysVarTime = "Time"
	// SysVarRow is the in-group data row counter ("Row#" — C# RowVariable.Name).
	SysVarRow = "Row#"
	// SysVarAbsRow is the absolute data row counter ("AbsRow#" — C# AbsRowVariable.Name).
	SysVarAbsRow = "AbsRow#"
	// SysVarPageMacro is the page-number print macro ("Page#" — C# PageMacroVariable.Name).
	SysVarPageMacro = "Page#"
	// SysVarTotalPagesMacro is the total-pages print macro ("TotalPages#").
	SysVarTotalPagesMacro = "TotalPages#"
	// SysVarCopyNameMacro is the copy-name print macro ("CopyName#").
	SysVarCopyNameMacro = "CopyName#"
	// SysVarHierarchyLevel is the hierarchy nesting depth ("HierarchyLevel").
	SysVarHierarchyLevel = "HierarchyLevel"
	// SysVarHierarchyRow is the dot-separated hierarchy row identifier ("HierarchyRow#").
	SysVarHierarchyRow = "HierarchyRow#"
)

// SystemVariables holds the built-in engine variables updated during report
// preparation. It is the Go equivalent of FastReport.Data.SystemVariables.
// C# ref: FastReport.Data.SystemVariables (SystemVariables.cs)
type SystemVariables struct {
	// PageNumber is the current page number (1-based).
	// Corresponds to C# PageVariable (Name="Page").
	PageNumber int
	// TotalPages is the total number of pages (available after double-pass or report finish).
	// Corresponds to C# TotalPagesVariable (Name="TotalPages").
	TotalPages int
	// Date is the report run date/time.
	// Corresponds to C# DateVariable (Name="Date").
	Date time.Time
	// Time is the report run time (same instant as Date).
	Time time.Time
	// Row is the current data row number within the current band (1-based).
	// Corresponds to C# RowVariable (Name="Row#").
	Row int
	// AbsRow is the absolute row number across all groups (1-based).
	// Corresponds to C# AbsRowVariable (Name="AbsRow#").
	AbsRow int
	// HierarchyLevel is the nesting level in a hierarchical data band.
	// Corresponds to C# HierarchyLevelVariable (Name="HierarchyLevel").
	HierarchyLevel int
	// HierarchyRow is the dot-separated row identifier in hierarchical reports (e.g. "1.2.3").
	// Corresponds to C# HierarchyRowNoVariable (Name="HierarchyRow#").
	HierarchyRow string
}

// NewSystemVariables creates a SystemVariables initialised to the current time.
func NewSystemVariables() *SystemVariables {
	now := time.Now()
	return &SystemVariables{
		PageNumber: 1,
		Date:       now,
		Time:       now,
		Row:        1,
		AbsRow:     1,
	}
}

// ToParameters converts the system variables to a []*Parameter slice suitable
// for use as Dictionary.SystemVariables().
// The returned slice includes both the canonical C# names ("Page", "Row#", etc.)
// and the Go-idiomatic aliases ("PageNumber", "AbsRow#", etc.) for compatibility.
func (sv *SystemVariables) ToParameters() []*Parameter {
	return []*Parameter{
		// C# canonical names
		{Name: SysVarPage, Value: sv.PageNumber},
		{Name: SysVarTotalPages, Value: sv.TotalPages},
		{Name: SysVarDate, Value: sv.Date},
		{Name: SysVarTime, Value: sv.Time},
		{Name: SysVarRow, Value: sv.Row},
		{Name: SysVarAbsRow, Value: sv.AbsRow},
		{Name: SysVarHierarchyLevel, Value: sv.HierarchyLevel},
		{Name: SysVarHierarchyRow, Value: sv.HierarchyRow},
		// Go aliases / extra names
		{Name: SysVarPageNumber, Value: sv.PageNumber},
		{Name: SysVarPageCount, Value: sv.TotalPages},
	}
}

// Get returns the value of the named system variable.
// Accepts both the canonical C# names and the Go alias names.
// Returns nil when the name is not recognised.
func (sv *SystemVariables) Get(name string) any {
	switch name {
	case SysVarPage, SysVarPageNumber:
		return sv.PageNumber
	case SysVarTotalPages, SysVarPageCount:
		return sv.TotalPages
	case SysVarDate:
		return sv.Date
	case SysVarTime:
		return sv.Time
	case SysVarRow:
		return sv.Row
	case SysVarAbsRow:
		return sv.AbsRow
	case SysVarHierarchyLevel:
		return sv.HierarchyLevel
	case SysVarHierarchyRow:
		return sv.HierarchyRow
	}
	return nil
}

// Set updates a system variable by name.
// Accepts both the canonical C# names and the Go alias names.
// Unknown names are ignored.
func (sv *SystemVariables) Set(name string, value any) {
	switch name {
	case SysVarPage, SysVarPageNumber:
		if v, ok := value.(int); ok {
			sv.PageNumber = v
		}
	case SysVarTotalPages, SysVarPageCount:
		if v, ok := value.(int); ok {
			sv.TotalPages = v
		}
	case SysVarDate:
		if v, ok := value.(time.Time); ok {
			sv.Date = v
		}
	case SysVarTime:
		if v, ok := value.(time.Time); ok {
			sv.Time = v
		}
	case SysVarRow:
		if v, ok := value.(int); ok {
			sv.Row = v
		}
	case SysVarAbsRow:
		if v, ok := value.(int); ok {
			sv.AbsRow = v
		}
	case SysVarHierarchyLevel:
		if v, ok := value.(int); ok {
			sv.HierarchyLevel = v
		}
	case SysVarHierarchyRow:
		if v, ok := value.(string); ok {
			sv.HierarchyRow = v
		}
	}
}
