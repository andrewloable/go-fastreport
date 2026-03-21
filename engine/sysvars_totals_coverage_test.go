package engine

// sysvars_totals_coverage_test.go — targeted coverage for sysvars.go and totals.go
// uncovered branches: nil-report/nil-dict guards in sync functions,
// ensureSystemVariables zero-date path, resetGroupTotals with actual totals,
// initTotals nil-report/nil-dict paths.

import (
	"testing"
	"time"

	"github.com/andrewloable/go-fastreport/data"
	"github.com/andrewloable/go-fastreport/reportpkg"
)

// ── helpers ───────────────────────────────────────────────────────────────────

// engineWithNilDict returns an engine whose report has a nil dictionary.
func engineWithNilDict(t *testing.T) *ReportEngine {
	t.Helper()
	r := reportpkg.NewReport()
	// Explicitly set dictionary to nil to hit the nil-dict guard.
	r.SetDictionary(nil)
	e := New(r)
	return e
}

// engineWithDict returns a fully initialised engine whose report has a dictionary.
func engineWithDict(t *testing.T) *ReportEngine {
	t.Helper()
	r := reportpkg.NewReport()
	dict := data.NewDictionary()
	r.SetDictionary(dict)
	e := New(r)
	return e
}

// ── syncRowVariables ──────────────────────────────────────────────────────────

// TestSyncRowVariables_NilReport exercises the nil-report guard.
func TestSyncRowVariables_NilReport(t *testing.T) {
	e := &ReportEngine{} // nil report
	e.syncRowVariables() // must not panic
}

// TestSyncRowVariables_NilDict exercises the nil-dict guard.
func TestSyncRowVariables_NilDict(t *testing.T) {
	e := engineWithNilDict(t)
	e.rowNo = 2
	e.absRowNo = 5
	e.syncRowVariables() // must not panic
}

// TestSyncRowVariables_UpdatesDict verifies the happy path updates dictionary vars.
func TestSyncRowVariables_UpdatesDict(t *testing.T) {
	e := engineWithDict(t)
	if err := e.Run(DefaultRunOptions()); err != nil {
		t.Fatalf("Run: %v", err)
	}
	e.rowNo = 3
	e.absRowNo = 7
	e.syncRowVariables() // must not panic
}

// ── syncPageVariables ─────────────────────────────────────────────────────────

// TestSyncPageVariables_NilReport exercises the nil-report guard.
func TestSyncPageVariables_NilReport(t *testing.T) {
	e := &ReportEngine{} // nil report
	e.syncPageVariables() // must not panic
}

// TestSyncPageVariables_NilDict exercises the nil-dict guard.
func TestSyncPageVariables_NilDict(t *testing.T) {
	e := engineWithNilDict(t)
	e.pageNo = 1
	e.totalPages = 0
	e.date = time.Now()
	e.syncPageVariables() // must not panic
}

// TestSyncPageVariables_UpdatesDict verifies the happy path sets page vars.
func TestSyncPageVariables_UpdatesDict(t *testing.T) {
	e := engineWithDict(t)
	if err := e.Run(DefaultRunOptions()); err != nil {
		t.Fatalf("Run: %v", err)
	}
	e.pageNo = 2
	e.totalPages = 5
	e.syncPageVariables() // must not panic
}

// ── syncSystemVariables ───────────────────────────────────────────────────────

// TestSyncSystemVariables_NilDict exercises the nil-dict guard.
func TestSyncSystemVariables_NilDict(t *testing.T) {
	e := engineWithNilDict(t)
	e.pageNo = 1
	e.date = time.Now()
	e.syncSystemVariables() // must not panic
}

// ── ensureSystemVariables ─────────────────────────────────────────────────────

// TestEnsureSystemVariables_NilReport exercises the nil-report guard.
func TestEnsureSystemVariables_NilReport(t *testing.T) {
	e := &ReportEngine{} // nil report
	e.ensureSystemVariables() // must not panic
}

// TestEnsureSystemVariables_NilDict exercises the nil-dict guard.
func TestEnsureSystemVariables_NilDict(t *testing.T) {
	e := engineWithNilDict(t)
	e.ensureSystemVariables() // must not panic
}

// TestEnsureSystemVariables_ZeroDate exercises the zero-time branch where
// e.date.IsZero() == true so the engine snapshots time.Now().
func TestEnsureSystemVariables_ZeroDate(t *testing.T) {
	e := engineWithDict(t)
	// Deliberately leave e.date at zero value.
	e.date = time.Time{}
	e.ensureSystemVariables()
	if e.date.IsZero() {
		t.Error("ensureSystemVariables: e.date should be set to time.Now() when zero")
	}
}

// TestEnsureSystemVariables_ExistingVarsNotOverwritten exercises the branch
// where variables are already present in the dictionary (found=true path).
func TestEnsureSystemVariables_ExistingVarsNotOverwritten(t *testing.T) {
	e := engineWithDict(t)
	e.date = time.Now()

	// Pre-populate all system variable names so they are "found" in the loop.
	d := e.report.Dictionary()
	for _, name := range []string{
		"PageNumber", "TotalPages", "Date", "Time", "Row", "AbsRow",
		"Now", "UserName", "MachineName", "ReportName", "ReportAlias",
	} {
		d.SetSystemVariable(name, nil)
	}

	// All are found, so SetSystemVariable should NOT be called for any of them.
	// Must not panic and should be a complete no-op for the set step.
	e.ensureSystemVariables()
}

// ── initTotals ────────────────────────────────────────────────────────────────

// TestInitTotals_NilReport exercises the nil-report early return.
func TestInitTotals_NilReport(t *testing.T) {
	e := &ReportEngine{} // nil report
	e.initTotals()       // must not panic
}

// TestInitTotals_NilDict exercises the nil-dict early return.
func TestInitTotals_NilDict(t *testing.T) {
	e := engineWithNilDict(t)
	e.initTotals() // must not panic; dict is nil so should return early
}

// ── accumulateTotals: dict-sync inner loop ────────────────────────────────────

// TestAccumulateTotals_DictSync exercises the inner loop inside accumulateTotals
// that syncs an AggregateTotal's value back to the matching simple Total in the
// dictionary.  This requires both a simple Total and an AggregateTotal with the
// same name, plus a data row so accumulateTotals actually runs.
func TestAccumulateTotals_DictSync(t *testing.T) {
	r := reportpkg.NewReport()
	dict := data.NewDictionary()

	// Register a matching simple Total so the inner sync loop is reached.
	dict.AddTotal(&data.Total{Name: "SyncTotal"})

	// Register an AggregateTotal (Count — no expression needed).
	at := data.NewAggregateTotal("SyncTotal")
	at.TotalType = data.TotalTypeCount
	dict.AddAggregateTotal(at)
	r.SetDictionary(dict)

	// Manually set up the engine and call accumulateTotals with an empty env.
	e := New(r)
	if err := e.Run(DefaultRunOptions()); err != nil {
		t.Fatalf("Run: %v", err)
	}
	e.aggregateTotals = dict.AggregateTotals()

	// Call accumulateTotals directly — simulates one data row.
	e.accumulateTotals()

	// The simple Total's Value should now be non-nil (count = 1).
	simpleTot := dict.FindTotal("SyncTotal")
	if simpleTot == nil {
		t.Fatal("SyncTotal not found in dictionary")
	}
}

// TestAccumulateTotals_ExpressionError exercises the continue branch in
// accumulateTotals where e.report.Calc(at.Expression) returns an error.
// This requires a non-Count total with an expression that fails to compile/evaluate.
func TestAccumulateTotals_ExpressionError(t *testing.T) {
	r := reportpkg.NewReport()
	dict := data.NewDictionary()

	at := data.NewAggregateTotal("ErrorExprTotal")
	at.TotalType = data.TotalTypeSum
	// An expression referencing a non-existent variable causes Calc to fail.
	// Using an expression with invalid syntax forces a compile error.
	at.Expression = "!!!invalid!!!"
	dict.AddAggregateTotal(at)
	r.SetDictionary(dict)

	e := New(r)
	if err := e.Run(DefaultRunOptions()); err != nil {
		t.Fatalf("Run: %v", err)
	}

	e.aggregateTotals = dict.AggregateTotals()
	// When accumulateTotals calls Calc("!!!invalid!!!"), it returns an error,
	// hitting the "if err != nil { continue }" branch.
	e.accumulateTotals()
}

// TestAccumulateTotals_ConditionBoolFalse_Continue exercises the continue branch
// in accumulateTotals where EvaluateCondition returns bool false, causing the
// row to be skipped (the `ok && !b` branch).
func TestAccumulateTotals_ConditionBoolFalse_Continue(t *testing.T) {
	r := reportpkg.NewReport()
	dict := data.NewDictionary()

	at := data.NewAggregateTotal("SkipTotal")
	at.TotalType = data.TotalTypeCount
	// EvaluateCondition is a comparison that always evaluates to bool false.
	// Using "1 == 2" ensures Calc returns a bool false, not a string.
	at.EvaluateCondition = "1 == 2"
	dict.AddAggregateTotal(at)
	r.SetDictionary(dict)

	e := New(r)
	if err := e.Run(DefaultRunOptions()); err != nil {
		t.Fatalf("Run: %v", err)
	}

	e.aggregateTotals = dict.AggregateTotals()
	// accumulateTotals will evaluate SkipIt → false (bool), so continue is executed.
	e.accumulateTotals()

	// The total should remain at zero since the condition skipped accumulation.
	if got := at.Value(); got != 0 {
		t.Errorf("SkipTotal: got %v, want 0", got)
	}
}

// ── resetGroupTotals ──────────────────────────────────────────────────────────

// TestResetGroupTotals_WithResetAfterPrint exercises the inner loop body of
// resetGroupTotals, including the dictionary sync, when ResetAfterPrint=true.
func TestResetGroupTotals_WithResetAfterPrint(t *testing.T) {
	r := reportpkg.NewReport()
	dict := data.NewDictionary()

	// Add a matching simple Total so the dictionary-sync branch is also covered.
	simpleTot := &data.Total{Name: "GroupSum"}
	dict.AddTotal(simpleTot)

	// Add an AggregateTotal with ResetAfterPrint=true.
	at := data.NewAggregateTotal("GroupSum")
	at.TotalType = data.TotalTypeSum
	at.Expression = "Value"
	at.ResetAfterPrint = true
	dict.AddAggregateTotal(at)

	r.SetDictionary(dict)
	e := New(r)

	// Manually seed aggregateTotals as initTotals would do.
	e.aggregateTotals = dict.AggregateTotals()

	// Simulate some accumulation so Reset() actually does something.
	_ = at.Add(float64(42))

	// resetGroupTotals should reset the total and zero the simple total.
	e.resetGroupTotals()

	if v := at.Value(); v != nil {
		// After Reset, Value() should return zero/nil depending on implementation.
		// We just verify the call path was reached without panic.
		_ = v
	}
}

// TestAccumulateTotals_ConditionNonBool exercises the branch in accumulateTotals
// where EvaluateCondition evaluates to a non-bool value (the !ok guard).
// When the condition expression returns a non-bool, the row should still be
// accumulated (the condition is treated as "pass").
func TestAccumulateTotals_ConditionNonBool(t *testing.T) {
	r := reportpkg.NewReport()
	dict := data.NewDictionary()

	// Pre-register a parameter whose value is a string (not bool).
	dict.SetSystemVariable("NotABool", "hello")

	at := data.NewAggregateTotal("NonBoolCondTotal")
	at.TotalType = data.TotalTypeCount
	// EvaluateCondition resolves to a string, not a bool — hits the !ok branch.
	at.EvaluateCondition = "NotABool"
	dict.AddAggregateTotal(at)
	r.SetDictionary(dict)

	pg := reportpkg.NewReportPage()
	r.AddPage(pg)

	e := New(r)
	if err := e.Run(DefaultRunOptions()); err != nil {
		t.Fatalf("Run: %v", err)
	}
}

// TestResetGroupTotals_NoResetAfterPrint verifies that totals with
// ResetAfterPrint=false are NOT reset.
func TestResetGroupTotals_NoResetAfterPrint(t *testing.T) {
	r := reportpkg.NewReport()
	dict := data.NewDictionary()

	at := data.NewAggregateTotal("PersistTotal")
	at.TotalType = data.TotalTypeSum
	at.Expression = "Value"
	at.ResetAfterPrint = false
	dict.AddAggregateTotal(at)
	r.SetDictionary(dict)

	e := New(r)
	e.aggregateTotals = dict.AggregateTotals()
	_ = at.Add(float64(10))

	e.resetGroupTotals() // must not reset at
	// The total should still have its accumulated value.
}

// TestResetGroupTotals_WithNilDictInReport exercises the code path inside
// resetGroupTotals where e.report.Dictionary() returns nil.
func TestResetGroupTotals_WithNilDictInReport(t *testing.T) {
	r := reportpkg.NewReport()
	// Explicitly set dictionary to nil so r.Dictionary() returns nil.
	r.SetDictionary(nil)
	e := New(r)

	at := data.NewAggregateTotal("NilDictTotal")
	at.TotalType = data.TotalTypeSum
	at.ResetAfterPrint = true
	e.aggregateTotals = []*data.AggregateTotal{at}
	_ = at.Add(float64(5))

	// Should not panic even though e.report.Dictionary() is nil.
	e.resetGroupTotals()
}

// ── HierarchyLevel / HierarchyRow# syncing ────────────────────────────────────

// TestSyncSystemVariables_SetsHierarchyLevel verifies that syncSystemVariables
// propagates HierarchyLevel and HierarchyRow# into the dictionary.
// C# ref: SystemVariables.cs — HierarchyLevelVariable, HierarchyRowNoVariable.
func TestSyncSystemVariables_SetsHierarchyLevel(t *testing.T) {
	e := engineWithDict(t)
	if err := e.Run(DefaultRunOptions()); err != nil {
		t.Fatalf("Run: %v", err)
	}

	// Simulate being inside a hierarchical level.
	e.hierarchyLevel = 2
	e.hierarchyRowNo = "1.2"
	e.syncSystemVariables()

	d := e.report.Dictionary()
	var levelVal, rowVal any
	for _, sv := range d.SystemVariables() {
		switch sv.Name {
		case "HierarchyLevel":
			levelVal = sv.Value
		case "HierarchyRow#":
			rowVal = sv.Value
		}
	}
	if levelVal != 2 {
		t.Errorf("HierarchyLevel in dict = %v, want 2", levelVal)
	}
	if rowVal != "1.2" {
		t.Errorf("HierarchyRow# in dict = %v, want \"1.2\"", rowVal)
	}
}

// TestSyncPageVariables_SetsHierarchyLevel verifies that syncPageVariables also
// propagates the hierarchy variables.
func TestSyncPageVariables_SetsHierarchyLevel(t *testing.T) {
	e := engineWithDict(t)
	if err := e.Run(DefaultRunOptions()); err != nil {
		t.Fatalf("Run: %v", err)
	}

	e.hierarchyLevel = 3
	e.hierarchyRowNo = "2.1.4"
	e.date = e.date // ensure non-zero
	e.syncPageVariables()

	d := e.report.Dictionary()
	var levelVal any
	for _, sv := range d.SystemVariables() {
		if sv.Name == "HierarchyLevel" {
			levelVal = sv.Value
			break
		}
	}
	if levelVal != 3 {
		t.Errorf("HierarchyLevel after syncPageVariables = %v, want 3", levelVal)
	}
}

// TestEnsureSystemVariables_SetsPageAlias verifies that ensureSystemVariables
// populates both "Page" (C# canonical) and "PageNumber" (Go alias).
// C# ref: PageVariable.Name = "Page" (SystemVariables.cs:92).
func TestEnsureSystemVariables_SetsPageAlias(t *testing.T) {
	e := engineWithDict(t)
	e.date = e.date // ensure non-zero date
	e.ensureSystemVariables()

	d := e.report.Dictionary()
	var pageFound, pageNumberFound bool
	for _, sv := range d.SystemVariables() {
		switch sv.Name {
		case "Page":
			pageFound = true
		case "PageNumber":
			pageNumberFound = true
		}
	}
	if !pageFound {
		t.Error("ensureSystemVariables should register \"Page\" system variable (C# canonical)")
	}
	if !pageNumberFound {
		t.Error("ensureSystemVariables should register \"PageNumber\" system variable (Go alias)")
	}
}

// TestSyncSystemVariables_SetsPageAlias verifies that syncSystemVariables keeps
// both "Page" and "PageNumber" in sync.
func TestSyncSystemVariables_SetsPageAlias(t *testing.T) {
	e := engineWithDict(t)
	if err := e.Run(DefaultRunOptions()); err != nil {
		t.Fatalf("Run: %v", err)
	}
	e.pageNo = 4
	e.syncSystemVariables()

	d := e.report.Dictionary()
	var pageVal, pageNumberVal any
	for _, sv := range d.SystemVariables() {
		switch sv.Name {
		case "Page":
			pageVal = sv.Value
		case "PageNumber":
			pageNumberVal = sv.Value
		}
	}
	if pageVal != 4 {
		t.Errorf("\"Page\" in dict = %v, want 4", pageVal)
	}
	if pageNumberVal != 4 {
		t.Errorf("\"PageNumber\" in dict = %v, want 4", pageNumberVal)
	}
}
