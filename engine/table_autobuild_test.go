package engine

// table_autobuild_test.go — internal tests for autoManualBuild and tableExtractDSAlias.

import (
	"testing"

	"github.com/andrewloable/go-fastreport/data"
	"github.com/andrewloable/go-fastreport/reportpkg"
	"github.com/andrewloable/go-fastreport/table"
)

// ── tableExtractDSAlias ───────────────────────────────────────────────────────

func TestTableExtractDSAlias_Simple(t *testing.T) {
	got := tableExtractDSAlias("[Employees.FirstName]")
	if got != "Employees" {
		t.Errorf("got %q, want Employees", got)
	}
}

func TestTableExtractDSAlias_MultipleExpressions(t *testing.T) {
	got := tableExtractDSAlias("Hello [Employees.FirstName] [Employees.LastName]")
	if got != "Employees" {
		t.Errorf("got %q, want Employees", got)
	}
}

func TestTableExtractDSAlias_NoExpression(t *testing.T) {
	got := tableExtractDSAlias("Static text")
	if got != "" {
		t.Errorf("got %q, want empty", got)
	}
}

func TestTableExtractDSAlias_NoDot(t *testing.T) {
	got := tableExtractDSAlias("[NoField]")
	if got != "" {
		t.Errorf("got %q, want empty", got)
	}
}

func TestTableExtractDSAlias_EmptyString(t *testing.T) {
	got := tableExtractDSAlias("")
	if got != "" {
		t.Errorf("got %q, want empty", got)
	}
}

// ── autoManualBuild ───────────────────────────────────────────────────────────

// buildAutoManualEngine creates an engine with a data source "Employees"
// containing 2 rows. The template table has FixedColumns=1 and
// ManualBuildEvent set (simulating an FRX table with a C# script).
func buildAutoManualEngine(t *testing.T) (*ReportEngine, *table.TableObject) {
	t.Helper()

	r := reportpkg.NewReport()
	pg := reportpkg.NewReportPage()
	r.AddPage(pg)

	// Set up the Employees data source with 2 rows.
	dict := data.NewDictionary()
	ds := data.NewBaseDataSource("Employees")
	ds.SetAlias("Employees")
	ds.AddColumn(data.Column{Name: "FirstName", DataType: "string"})
	ds.AddColumn(data.Column{Name: "Title", DataType: "string"})
	ds.AddRow(map[string]any{"FirstName": "Anne", "Title": "Sales Rep"})
	ds.AddRow(map[string]any{"FirstName": "Bob", "Title": "Manager"})
	dict.AddDataSource(ds)
	r.SetDictionary(dict)

	e := New(r)
	if err := e.Run(DefaultRunOptions()); err != nil {
		t.Fatalf("Run: %v", err)
	}

	// Build template: 2 rows × 3 columns (header, data, footer).
	// FixedColumns=1 means column 0 is the header column.
	tbl := table.NewTableObject()
	tbl.SetFixedColumns(1)
	tbl.ManualBuildEvent = "Table1_ManualBuild"
	tbl.NewColumn() // col 0: header
	tbl.NewColumn() // col 1: data
	tbl.NewColumn() // col 2: footer
	row0 := tbl.NewRow()
	row0.Cell(0).SetText("Name")
	row0.Cell(1).SetText("[Employees.FirstName]")
	row0.Cell(2).SetText("Count")
	row1 := tbl.NewRow()
	row1.Cell(0).SetText("Title")
	row1.Cell(1).SetText("[Employees.Title]")
	row1.Cell(2).SetText("2")

	return e, tbl
}

// TestAutoManualBuild_ColumnFirst verifies that autoManualBuild generates the
// correct result table for the standard column-first ManualBuild pattern:
// header col + one data col per DS row + footer col.
func TestAutoManualBuild_ColumnFirst(t *testing.T) {
	e, tbl := buildAutoManualEngine(t)

	result := e.autoManualBuild(tbl)
	if result == nil {
		t.Fatal("autoManualBuild returned nil")
	}

	// Expected: 2 template rows, 4 result columns (header + 2 data + footer).
	if result.RowCount() != 2 {
		t.Errorf("RowCount: got %d, want 2", result.RowCount())
	}
	if result.ColumnCount() != 4 {
		t.Errorf("ColumnCount: got %d, want 4", result.ColumnCount())
	}

	// Header column (col 0): static text from template.
	c := result.Cell(0, 0)
	if c == nil || c.Text() != "Name" {
		t.Errorf("Cell(0,0) = %q, want Name", c.Text())
	}

	// First data column (col 1): evaluated with DS row 0 (Anne).
	c = result.Cell(0, 1)
	if c == nil || c.Text() != "Anne" {
		t.Errorf("Cell(0,1) = %q, want Anne", c.Text())
	}

	// Second data column (col 2): evaluated with DS row 1 (Bob).
	c = result.Cell(0, 2)
	if c == nil || c.Text() != "Bob" {
		t.Errorf("Cell(0,2) = %q, want Bob", c.Text())
	}

	// Footer column (col 3): static text from template.
	c = result.Cell(0, 3)
	if c == nil || c.Text() != "Count" {
		t.Errorf("Cell(0,3) = %q, want Count", c.Text())
	}

	// Row 1: title row.
	c = result.Cell(1, 1)
	if c == nil || c.Text() != "Sales Rep" {
		t.Errorf("Cell(1,1) = %q, want Sales Rep", c.Text())
	}
	c = result.Cell(1, 2)
	if c == nil || c.Text() != "Manager" {
		t.Errorf("Cell(1,2) = %q, want Manager", c.Text())
	}
}

// TestAutoManualBuild_NoCallback_Required verifies that autoManualBuild only
// triggers when ManualBuild callback is nil (not already set by user code).
func TestAutoManualBuild_NoCallback_Required(t *testing.T) {
	e, tbl := buildAutoManualEngine(t)
	// Set a Go callback — autoManualBuild should return nil (let InvokeManualBuild handle it).
	tbl.ManualBuild = func(h *table.TableHelper) {}
	result := e.autoManualBuild(tbl)
	if result != nil {
		t.Error("autoManualBuild should return nil when ManualBuild callback is set")
	}
}

// TestAutoManualBuild_NoManualBuildEvent_Returns_Nil verifies autoManualBuild
// returns nil when no ManualBuildEvent is set.
func TestAutoManualBuild_NoManualBuildEvent_Returns_Nil(t *testing.T) {
	e, tbl := buildAutoManualEngine(t)
	tbl.ManualBuildEvent = "" // clear event
	result := e.autoManualBuild(tbl)
	if result != nil {
		t.Error("autoManualBuild should return nil when ManualBuildEvent is empty")
	}
}

// TestAutoManualBuild_SimpleRowFirst verifies that autoManualBuild generates a
// valid result for the simple row-first pattern (no FixedRows, no FixedColumns,
// DS expressions in row 1):  header row printed once + one row per DS record.
func TestAutoManualBuild_SimpleRowFirst(t *testing.T) {
	e, tbl := buildAutoManualEngine(t)
	tbl.SetFixedColumns(0)
	result := e.autoManualBuild(tbl)
	if result == nil {
		t.Fatal("autoManualBuild should return non-nil for simple row-first pattern")
	}
	// 1 header row + 2 data rows (one per Employees record).
	if result.RowCount() != 3 {
		t.Errorf("expected 3 rows, got %d", result.RowCount())
	}
}

// TestAutoManualBuild_NoFixedColumns_NoDS_Returns_Nil verifies that
// autoManualBuild returns nil when FixedColumns=0 and no DS alias can be
// detected in the data row (row 1).
func TestAutoManualBuild_NoFixedColumns_NoDS_Returns_Nil(t *testing.T) {
	e, tbl := buildAutoManualEngine(t)
	tbl.SetFixedColumns(0)
	// Clear DS expressions from row 1 so detection fails.
	tbl.Cell(1, 1).SetText("Static value")
	tbl.Cell(1, 0).SetText("Static 2")
	result := e.autoManualBuild(tbl)
	if result != nil {
		t.Error("autoManualBuild should return nil when FixedColumns=0 and no DS detected")
	}
}

// TestAutoManualBuild_UnknownDS_Returns_Nil verifies that autoManualBuild
// returns nil when no matching data source is found in the dictionary.
func TestAutoManualBuild_UnknownDS_Returns_Nil(t *testing.T) {
	e, tbl := buildAutoManualEngine(t)
	// Replace data cell expression with unknown DS alias.
	tbl.Cell(0, 1).SetText("[Unknown.Field]")
	tbl.Cell(1, 1).SetText("[Unknown.Field2]")
	result := e.autoManualBuild(tbl)
	if result != nil {
		t.Error("autoManualBuild should return nil when DS alias not found")
	}
}

// TestAutoManualBuild_StaticDataColumn_Returns_Nil verifies that autoManualBuild
// returns nil when the data column has no [DS.Field] expressions.
func TestAutoManualBuild_StaticDataColumn_Returns_Nil(t *testing.T) {
	e, tbl := buildAutoManualEngine(t)
	tbl.Cell(0, 1).SetText("Static")
	tbl.Cell(1, 1).SetText("Static")
	result := e.autoManualBuild(tbl)
	if result != nil {
		t.Error("autoManualBuild should return nil when data column has no DS expressions")
	}
}
