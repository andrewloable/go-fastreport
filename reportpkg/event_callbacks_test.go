package reportpkg_test

// event_callbacks_test.go — tests for the Go equivalents of C# ReportEventArgs:
//   - OnCustomCalc      (C# Report.CustomCalc / CustomCalcEventArgs)
//   - OnLoadBaseReport  (C# Report.LoadBaseReport / CustomLoadEventArgs)
//
// These were added during the go-fastreport-yixy issue analysis:
//   C# src: FastReport.Base/ReportEventArgs.cs, Report.cs

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/andrewloable/go-fastreport/data"
	"github.com/andrewloable/go-fastreport/reportpkg"
)

// ── OnCustomCalc ─────────────────────────────────────────────────────────────

// TestOnCustomCalc_OverridesResolvedValue verifies that OnCustomCalc can
// replace the value returned by Calc for a parameter expression.
// Mirrors C# CustomCalcEventArgs.CalculatedObject being reassigned.
func TestOnCustomCalc_OverridesResolvedValue(t *testing.T) {
	r := reportpkg.NewReport()
	r.Dictionary().AddParameter(&data.Parameter{Name: "Price", Value: 10.0})

	// Intercept: multiply every resolved value by 3.
	r.OnCustomCalc = func(expr string, val any) any {
		if f, ok := val.(float64); ok {
			return f * 3
		}
		return val
	}

	got, err := r.Calc("[Price]")
	if err != nil {
		t.Fatalf("Calc: %v", err)
	}
	// Expect 10.0 * 3 = 30.0.
	if got != 30.0 {
		t.Errorf("OnCustomCalc override: got %v, want 30.0", got)
	}
}

// TestOnCustomCalc_NilDoesNotAffectResult verifies that when OnCustomCalc is
// nil the value is returned as-is (no regression).
func TestOnCustomCalc_NilDoesNotAffectResult(t *testing.T) {
	r := reportpkg.NewReport()
	r.Dictionary().AddParameter(&data.Parameter{Name: "X", Value: 42})

	got, err := r.Calc("[X]")
	if err != nil {
		t.Fatalf("Calc: %v", err)
	}
	if got != 42 {
		t.Errorf("nil OnCustomCalc: got %v, want 42", got)
	}
}

// TestOnCustomCalc_ReceivesExpressionString verifies the hook receives the
// original expression string passed to Calc.
func TestOnCustomCalc_ReceivesExpressionString(t *testing.T) {
	r := reportpkg.NewReport()
	r.Dictionary().AddParameter(&data.Parameter{Name: "Name", Value: "Alice"})

	var capturedExpr string
	r.OnCustomCalc = func(expr string, val any) any {
		capturedExpr = expr
		return val
	}

	if _, err := r.Calc("[Name]"); err != nil {
		t.Fatalf("Calc: %v", err)
	}
	if !strings.Contains(capturedExpr, "Name") {
		t.Errorf("expected expression containing 'Name', got %q", capturedExpr)
	}
}

// ── OnLoadBaseReport ──────────────────────────────────────────────────────────

// writeTemp writes content to a temp file with the given name suffix and
// returns its path. The caller is responsible for cleanup.
func writeTempFRX(t *testing.T, dir, name, content string) string {
	t.Helper()
	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatalf("write temp FRX %q: %v", path, err)
	}
	return path
}

// TestOnLoadBaseReport_CallbackUsed verifies that when OnLoadBaseReport is set,
// it is invoked instead of the default file-system loader when loading an
// inherited report. Mirrors C# Report.LoadBaseReport event (CustomLoadEventArgs).
func TestOnLoadBaseReport_CallbackUsed(t *testing.T) {
	dir := t.TempDir()

	// Write a minimal inherited FRX that references a base report.
	const inheritedContent = `<?xml version="1.0" encoding="utf-8"?>
<inherited BaseReport="base.frx">
</inherited>`
	childPath := writeTempFRX(t, dir, "child.frx", inheritedContent)

	// The base report returned by the callback.
	base := reportpkg.NewReport()
	base.Info.Name = "BaseFromCallback"

	var callbackPath string
	child := reportpkg.NewReport()
	child.OnLoadBaseReport = func(fileName string, r *reportpkg.Report) (*reportpkg.Report, error) {
		callbackPath = fileName
		return base, nil
	}

	if err := child.Load(childPath); err != nil {
		t.Fatalf("Load: %v", err)
	}

	// Callback should have been invoked with a path ending in "base.frx".
	if !strings.HasSuffix(callbackPath, "base.frx") {
		t.Errorf("expected callback path ending in 'base.frx', got %q", callbackPath)
	}
}

// TestOnLoadBaseReport_CallbackErrorPropagates verifies that an error returned
// by OnLoadBaseReport is wrapped and propagated.
func TestOnLoadBaseReport_CallbackErrorPropagates(t *testing.T) {
	dir := t.TempDir()

	const inheritedContent = `<?xml version="1.0" encoding="utf-8"?>
<inherited BaseReport="base.frx">
</inherited>`
	childPath := writeTempFRX(t, dir, "child.frx", inheritedContent)

	child := reportpkg.NewReport()
	child.OnLoadBaseReport = func(fileName string, r *reportpkg.Report) (*reportpkg.Report, error) {
		return nil, errors.New("simulated load failure")
	}

	err := child.Load(childPath)
	if err == nil {
		t.Fatal("expected error from OnLoadBaseReport callback, got nil")
	}
	if !strings.Contains(err.Error(), "simulated load failure") {
		t.Errorf("expected error to mention 'simulated load failure', got: %v", err)
	}
}
