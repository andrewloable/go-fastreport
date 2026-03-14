package engine

import "time"

// syncSystemVariables pushes the current engine state into the report's
// Dictionary system variables so that Report.Calc() can resolve them.
//
// Variables updated:
//   - PageNumber  — current logical page number (1-based)
//   - TotalPages  — total pages produced so far (updated after first pass)
//   - Date        — date portion of the run-start timestamp
//   - Time        — time portion of the run-start timestamp
//   - Row         — current data row within the active data band (1-based)
//   - AbsRow      — absolute row number across all data bands (1-based)
func (e *ReportEngine) syncSystemVariables() {
	if e.report == nil {
		return
	}
	d := e.report.Dictionary()
	if d == nil {
		return
	}
	d.SetSystemVariable("PageNumber", e.pageNo)
	d.SetSystemVariable("TotalPages", e.totalPages)
	d.SetSystemVariable("Date", e.date.Format("2006-01-02"))
	d.SetSystemVariable("Time", e.date.Format("15:04:05"))
	d.SetSystemVariable("Row", e.rowNo)
	d.SetSystemVariable("AbsRow", e.absRowNo)
	// Expose the full time.Time for callers that need it.
	d.SetSystemVariable("Now", e.date)
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
}

// syncPageVariables updates page-related system variables when a new page starts.
func (e *ReportEngine) syncPageVariables() {
	if e.report == nil {
		return
	}
	d := e.report.Dictionary()
	if d == nil {
		return
	}
	d.SetSystemVariable("PageNumber", e.pageNo)
	d.SetSystemVariable("TotalPages", e.totalPages)
	d.SetSystemVariable("Date", e.date.Format("2006-01-02"))
	d.SetSystemVariable("Time", e.date.Format("15:04:05"))
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
	defaults := map[string]any{
		"PageNumber": 1,
		"TotalPages": 0,
		"Date":       e.date.Format("2006-01-02"),
		"Time":       e.date.Format("15:04:05"),
		"Row":        1,
		"AbsRow":     1,
		"Now":        e.date,
		"UserName":   "",
		"MachineName": "",
		"ReportName": e.report.Name(),
		"ReportAlias": e.report.Name(),
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
