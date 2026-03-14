package data

import "time"

// SystemVariableNames lists the built-in system variable names.
// These mirror FastReport's built-in system variables.
const (
	SysVarPageNumber  = "PageNumber"
	SysVarTotalPages  = "TotalPages"
	SysVarPageCount   = "PageCount" // alias for TotalPages
	SysVarDate        = "Date"
	SysVarTime        = "Time"
	SysVarRow         = "Row#"
	SysVarAbsRow      = "AbsRow#"
	SysVarHierarchyLevel = "HierarchyLevel"
	SysVarHierarchyRow   = "HierarchyRow#"
)

// SystemVariables holds the built-in engine variables updated during report
// preparation. It is the Go equivalent of FastReport.Data.SystemVariables.
type SystemVariables struct {
	// PageNumber is the current page number (1-based).
	PageNumber int
	// TotalPages is the total number of pages (available after double-pass or report finish).
	TotalPages int
	// Date is the report run date.
	Date time.Time
	// Time is the report run time.
	Time time.Time
	// Row is the current data row number within the current band (1-based).
	Row int
	// AbsRow is the absolute row number across all groups (1-based).
	AbsRow int
	// HierarchyLevel is the nesting level in a hierarchical data band.
	HierarchyLevel int
	// HierarchyRow is the row number in the hierarchical band.
	HierarchyRow int
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
func (sv *SystemVariables) ToParameters() []*Parameter {
	return []*Parameter{
		{Name: SysVarPageNumber, Value: sv.PageNumber},
		{Name: SysVarTotalPages, Value: sv.TotalPages},
		{Name: SysVarPageCount, Value: sv.TotalPages},
		{Name: SysVarDate, Value: sv.Date},
		{Name: SysVarTime, Value: sv.Time},
		{Name: SysVarRow, Value: sv.Row},
		{Name: SysVarAbsRow, Value: sv.AbsRow},
		{Name: SysVarHierarchyLevel, Value: sv.HierarchyLevel},
		{Name: SysVarHierarchyRow, Value: sv.HierarchyRow},
	}
}

// Get returns the value of the named system variable.
// Returns nil when the name is not recognised.
func (sv *SystemVariables) Get(name string) any {
	switch name {
	case SysVarPageNumber:
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
// Unknown names are ignored.
func (sv *SystemVariables) Set(name string, value any) {
	switch name {
	case SysVarPageNumber:
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
		if v, ok := value.(int); ok {
			sv.HierarchyRow = v
		}
	}
}
