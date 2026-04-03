package engine

import (
	"fmt"
	"time"
)

// formatCSharpDateTime formats t as C# DateTime.ToString() does with en-US culture:
// "M/d/yyyy h:mm:ss\u202fAM" / "M/d/yyyy h:mm:ss\u202fPM".
// The narrow no-break space (\u202f) before AM/PM matches .NET 6+ en-US culture output.
// C# ref: FastReport.Data.DateVariable — Value = Report.Engine.Date (DateTime.Now).
func formatCSharpDateTime(t time.Time) string {
	ampm := "AM"
	if t.Hour() >= 12 {
		ampm = "PM"
	}
	return t.Format("1/2/2006 3:04:05") + "\u202f" + ampm
}

// syncSystemVariables pushes the current engine state into the report's
// Dictionary system variables so that Report.Calc() can resolve them.
//
// Variables updated (C# ref: FastReport.Data.SystemVariables — all variable classes):
//   - Page / PageNumber — current logical page number (1-based)
//   - TotalPages        — total pages produced so far (available after double-pass)
//   - PageN             — "Page N" string
//   - PageNofM          — "Page N of M" string
//   - Date              — run-start date as formatted string
//   - Time              — run-start time as formatted string
//   - Row / Row#        — data row within the current band (1-based)
//   - AbsRow / AbsRow#  — absolute row across all bands (1-based)
//   - HierarchyLevel    — nesting depth in hierarchical reports
//   - HierarchyRow#     — dot-separated hierarchy row identifier
func (e *ReportEngine) syncSystemVariables() {
	if e.report == nil {
		return
	}
	d := e.report.Dictionary()
	if d == nil {
		return
	}
	pageNo := e.GetLogicalPageNumber()
	totalPages := e.GetLogicalTotalPages()
	d.SetSystemVariable("Page", pageNo)       // C# PageVariable.Name = "Page"
	d.SetSystemVariable("PageNumber", pageNo)  // Go alias kept for backward compat
	d.SetSystemVariable("TotalPages", totalPages)
	d.SetSystemVariable("PageN", fmt.Sprintf("Page %d", pageNo))
	d.SetSystemVariable("PageNofM", fmt.Sprintf("Page %d of %d", pageNo, totalPages))
	d.SetSystemVariable("Date", formatCSharpDateTime(e.date))
	d.SetSystemVariable("Time", e.date.Format("15:04:05"))
	d.SetSystemVariable("Row", e.rowNo)
	d.SetSystemVariable("AbsRow", e.absRowNo)
	// Also keep the '#'-suffixed names in sync (used by [Row#] / [AbsRow#] expressions).
	d.SetSystemVariable("Row#", e.rowNo)
	d.SetSystemVariable("AbsRow#", e.absRowNo)
	// Expose the full time.Time for callers that need it.
	d.SetSystemVariable("Now", e.date)
	// Hierarchy variables: C# SystemVariables.cs — HierarchyLevel, HierarchyRow#.
	d.SetSystemVariable("HierarchyLevel", e.hierarchyLevel)
	d.SetSystemVariable("HierarchyRow#", e.hierarchyRowNo)
}

// syncRowVariables updates row-related system variables after each data row advance.
// Call this from RunDataBandFull when rowNo / absRowNo are updated.
func (e *ReportEngine) syncRowVariables() {
	if e.report == nil {
		return
	}
	d := e.report.Dictionary()
	if d == nil {
		return
	}
	d.SetSystemVariable("Row", e.rowNo)
	d.SetSystemVariable("AbsRow", e.absRowNo)
	// Also keep the '#'-suffixed names in sync (used by [Row#] / [AbsRow#] expressions).
	d.SetSystemVariable("Row#", e.rowNo)
	d.SetSystemVariable("AbsRow#", e.absRowNo)
}

// syncPageVariables updates page-related system variables when a new page starts.
// Uses GetLogicalPageNumber/GetLogicalTotalPages which read from the
// pageNumbers array (mirroring C# PageNo/TotalPages properties).
func (e *ReportEngine) syncPageVariables() {
	if e.report == nil {
		return
	}
	d := e.report.Dictionary()
	if d == nil {
		return
	}
	pageNo := e.GetLogicalPageNumber()
	totalPages := e.GetLogicalTotalPages()
	d.SetSystemVariable("Page", pageNo)       // C# PageVariable.Name = "Page"
	d.SetSystemVariable("PageNumber", pageNo)  // Go alias kept for backward compat
	d.SetSystemVariable("TotalPages", totalPages)
	d.SetSystemVariable("PageN", fmt.Sprintf("Page %d", pageNo))
	d.SetSystemVariable("PageNofM", fmt.Sprintf("Page %d of %d", pageNo, totalPages))
	d.SetSystemVariable("Date", formatCSharpDateTime(e.date))
	d.SetSystemVariable("Time", e.date.Format("15:04:05"))
	// Hierarchy variables are page-scoped in hierarchical reports.
	d.SetSystemVariable("HierarchyLevel", e.hierarchyLevel)
	d.SetSystemVariable("HierarchyRow#", e.hierarchyRowNo)
}

// ensureSystemVariables initialises system variables to their defaults at engine
// startup if they are not already registered in the dictionary.
func (e *ReportEngine) ensureSystemVariables() {
	if e.report == nil {
		return
	}
	d := e.report.Dictionary()
	if d == nil {
		return
	}
	info := e.report.Info
	defaults := map[string]any{
		// C# PageVariable.Name = "Page" (SystemVariables.cs)
		"Page":       1,
		"PageNumber": 1, // Go alias
		"TotalPages": 0,
		"PageN":      "Page 1",
		"PageNofM":   "Page 1 of 0",
		"Date":       formatCSharpDateTime(e.date),
		"Time":       e.date.Format("15:04:05"),
		"Row":        1,
		"AbsRow":     1,
		"Now":        e.date,
		"UserName":   "",
		"MachineName": "",
		"ReportName": e.report.Name(),
		"ReportAlias": e.report.Name(),
		// Macro variables: these are print-time placeholders whose values are
		// the bracket form shown in the preview. They mirror the C# behaviour:
		// CopyNameMacroVariable.Value == "[COPYNAME#]", etc.
		"Page#":       "[PAGE#]",
		"TotalPages#": "[TOTALPAGES#]",
		"CopyName#":   "[COPYNAME#]",
		// Row# and AbsRow# are runtime row counters, not macros. Initialise to 1.
		"Row#":    1,
		"AbsRow#": 1,
		// Hierarchy variables (C# SystemVariables.cs: HierarchyLevelVariable,
		// HierarchyRowNoVariable).
		"HierarchyLevel": 0,
		"HierarchyRow#":  "",
		// Report.ReportInfo.* fields (dot → underscore for expression eval).
		"Report_ReportInfo_Description":    info.Description,
		"Report_ReportInfo_Author":         info.Author,
		"Report_ReportInfo_Name":           info.Name,
		"Report_ReportInfo_Version":        info.Version,
		"Report_ReportInfo_Created":        info.Created,
		"Report_ReportInfo_Modified":       info.Modified,
		"Report_ReportInfo_CreatorVersion": info.CreatorVersion,
	}
	// Use SetSystemVariable which will add if missing.
	for k, v := range defaults {
		// Only set if not already present.
		found := false
		for _, sv := range d.SystemVariables() {
			if sv.Name == k {
				found = true
				break
			}
		}
		if !found {
			d.SetSystemVariable(k, v)
		}
	}
	// Snapshot the date for consistent use throughout the run.
	if e.date.IsZero() {
		e.date = time.Now()
	}
}
