package data_test

// tabledatasource_coverage_test.go — tests for TableDataSource and BaseDataSource
// property methods that had 0% coverage.

import (
	"testing"

	"github.com/andrewloable/go-fastreport/data"
	"github.com/andrewloable/go-fastreport/report"
)

// ── TableDataSource property methods ─────────────────────────────────────────

func TestTableDataSource_Enabled_DefaultTrue(t *testing.T) {
	ts := data.NewTableDataSource("test")
	if !ts.Enabled() {
		t.Error("Enabled should default to true")
	}
}

func TestTableDataSource_SetEnabled(t *testing.T) {
	ts := data.NewTableDataSource("test")
	ts.SetEnabled(false)
	if ts.Enabled() {
		t.Error("Enabled should be false after SetEnabled(false)")
	}
	ts.SetEnabled(true)
	if !ts.Enabled() {
		t.Error("Enabled should be true after SetEnabled(true)")
	}
}

func TestTableDataSource_IgnoreConnection(t *testing.T) {
	ts := data.NewTableDataSource("test")
	if ts.IgnoreConnection() {
		t.Error("IgnoreConnection should default to false")
	}
	ts.SetIgnoreConnection(true)
	if !ts.IgnoreConnection() {
		t.Error("IgnoreConnection should be true after SetIgnoreConnection(true)")
	}
	// When IgnoreConnection is true, Connection() returns nil.
	if ts.Connection() != nil {
		t.Error("Connection should be nil when IgnoreConnection is true")
	}
}

func TestTableDataSource_ForceLoadData(t *testing.T) {
	ts := data.NewTableDataSource("test")
	if ts.ForceLoadData() {
		t.Error("ForceLoadData should default to false")
	}
	ts.SetForceLoadData(true)
	if !ts.ForceLoadData() {
		t.Error("ForceLoadData should be true after SetForceLoadData(true)")
	}
}

func TestTableDataSource_QbSchema(t *testing.T) {
	ts := data.NewTableDataSource("test")
	if ts.QbSchema() != "" {
		t.Errorf("QbSchema should default to empty, got %q", ts.QbSchema())
	}
	ts.SetQbSchema("<schema>some xml</schema>")
	if ts.QbSchema() != "<schema>some xml</schema>" {
		t.Errorf("QbSchema = %q, want schema string", ts.QbSchema())
	}
}

// ── TableDataSource.Serialize / Deserialize ───────────────────────────────────

func TestTableDataSource_Serialize_Basic(t *testing.T) {
	ts := data.NewTableDataSource("Orders")
	ts.SetTableName("dbo.Orders")
	ts.SetSelectCommand("SELECT * FROM dbo.Orders")
	ts.SetEnabled(false)
	ts.SetQbSchema("<QB/>")
	ts.SetStoreData(true)

	w := newTDSWriter()
	if err := ts.Serialize(w); err != nil {
		t.Fatalf("Serialize: %v", err)
	}
	if w.strs["Name"] != "Orders" {
		t.Errorf("Name = %q, want 'Orders'", w.strs["Name"])
	}
	if w.strs["TableName"] != "dbo.Orders" {
		t.Errorf("TableName = %q, want 'dbo.Orders'", w.strs["TableName"])
	}
	if w.strs["SelectCommand"] != "SELECT * FROM dbo.Orders" {
		t.Errorf("SelectCommand = %q, want query", w.strs["SelectCommand"])
	}
	if v, ok := w.bools["Enabled"]; !ok || v {
		t.Error("Enabled should be written as false")
	}
	if w.strs["QbSchema"] != "<QB/>" {
		t.Errorf("QbSchema = %q, want '<QB/>'", w.strs["QbSchema"])
	}
	if v, ok := w.bools["StoreData"]; !ok || !v {
		t.Error("StoreData should be written as true")
	}
}

func TestTableDataSource_Serialize_EnabledNotWrittenWhenTrue(t *testing.T) {
	ts := data.NewTableDataSource("t")
	w := newTDSWriter()
	_ = ts.Serialize(w)
	if _, ok := w.bools["Enabled"]; ok {
		t.Error("Enabled should not be written when true (default)")
	}
}

func TestTableDataSource_Deserialize_Basic(t *testing.T) {
	ts := data.NewTableDataSource("")
	r := &tdsReader{
		strs: map[string]string{
			"Name":          "Products",
			"TableName":     "products",
			"SelectCommand": "SELECT * FROM products",
			"QbSchema":      "<schema/>",
		},
		bools: map[string]bool{
			"Enabled":   false,
			"StoreData": true,
		},
	}
	if err := ts.Deserialize(r); err != nil {
		t.Fatalf("Deserialize: %v", err)
	}
	if ts.Name() != "Products" {
		t.Errorf("Name = %q, want 'Products'", ts.Name())
	}
	if ts.TableName() != "products" {
		t.Errorf("TableName = %q, want 'products'", ts.TableName())
	}
	if ts.SelectCommand() != "SELECT * FROM products" {
		t.Errorf("SelectCommand = %q, want query", ts.SelectCommand())
	}
	if ts.Enabled() {
		t.Error("Enabled should be false after Deserialize")
	}
	if ts.QbSchema() != "<schema/>" {
		t.Errorf("QbSchema = %q, want '<schema/>'", ts.QbSchema())
	}
	if !ts.StoreData() {
		t.Error("StoreData should be true after Deserialize")
	}
}

// ── TableDataSource.RefreshColumns ───────────────────────────────────────────

func TestTableDataSource_RefreshColumns_Empty(t *testing.T) {
	// When no columns exist, RefreshColumns is a no-op.
	ts := data.NewTableDataSource("test")
	ts.RefreshColumns(true) // must not panic
}

func TestTableDataSource_RefreshColumns_WithColumns(t *testing.T) {
	c := data.NewDataConnectionBase("stub")
	c.ConnectionString = "stub://test"
	ts := c.CreateTable("t")
	ts.SetSelectCommand("SELECT id, name FROM users")
	// Init populates columns.
	if err := ts.Init(); err != nil {
		t.Fatalf("Init: %v", err)
	}
	// RefreshColumns should work without panic when columns exist.
	ts.RefreshColumns(true)
	ts.RefreshColumns(false)
	_ = c.Close()
}

// ── TableDataSource.InitSchema ────────────────────────────────────────────────

func TestTableDataSource_InitSchema_NoConnection(t *testing.T) {
	ts := data.NewTableDataSource("standalone")
	ts.SetSelectCommand("SELECT 1")
	// No connection — InitSchema should silently succeed.
	if err := ts.InitSchema(); err != nil {
		t.Errorf("InitSchema with no connection: %v", err)
	}
}

func TestTableDataSource_InitSchema_WithConnection(t *testing.T) {
	c := data.NewDataConnectionBase("stub")
	c.ConnectionString = "stub://test"
	if err := c.Open(); err != nil {
		t.Fatalf("Open: %v", err)
	}
	ts := c.CreateTable("t")
	ts.SetSelectCommand("SELECT id, name FROM users")
	// InitSchema should populate the column list.
	if err := ts.InitSchema(); err != nil {
		t.Fatalf("InitSchema: %v", err)
	}
	_ = c.Close()
}

// ── TableDataSource.GetCommandBuilder ────────────────────────────────────────

func TestDataConnectionBase_GetCommandBuilder_ReturnsNil(t *testing.T) {
	c := data.NewDataConnectionBase("stub")
	if c.GetCommandBuilder() != nil {
		t.Error("GetCommandBuilder should return nil for base connection")
	}
}

// ── BaseDataSource.BOF / HasMoreRows / Prior / SetCurrentRowNo / EnsureInit / GetDisplayName ──

func TestBaseDataSource_BOF_BeforeFirst(t *testing.T) {
	c := data.NewDataConnectionBase("stub")
	c.ConnectionString = "stub://test"
	ts := c.CreateTable("t")
	ts.SetSelectCommand("SELECT id, name FROM users")
	if err := ts.Init(); err != nil {
		t.Fatalf("Init: %v", err)
	}
	// After Init, currentRow == 0 → BOF should be false.
	if ts.BOF() {
		t.Error("BOF should be false after Init")
	}
	// Prior() decrements currentRow to -1 → BOF should be true.
	ts.Prior()
	if !ts.BOF() {
		t.Error("BOF should be true after Prior() from row 0")
	}
	_ = c.Close()
}

func TestBaseDataSource_BOF_AfterFirst(t *testing.T) {
	c := data.NewDataConnectionBase("stub")
	c.ConnectionString = "stub://test"
	ts := c.CreateTable("t")
	ts.SetSelectCommand("SELECT id, name FROM users")
	if err := ts.Init(); err != nil {
		t.Fatalf("Init: %v", err)
	}
	_ = ts.First()
	if ts.BOF() {
		t.Error("BOF should be false after First()")
	}
	_ = c.Close()
}

func TestBaseDataSource_HasMoreRows(t *testing.T) {
	c := data.NewDataConnectionBase("stub")
	c.ConnectionString = "stub://test"
	ts := c.CreateTable("t")
	ts.SetSelectCommand("SELECT id, name FROM users")
	if err := ts.Init(); err != nil {
		t.Fatalf("Init: %v", err)
	}
	// After Init, currentRow == 0 and rows exist → HasMoreRows should be true.
	if !ts.HasMoreRows() {
		t.Error("HasMoreRows should be true after Init() when rows exist")
	}
	// Advance past all rows.
	for ts.HasMoreRows() {
		if err := ts.Next(); err != nil {
			break
		}
	}
	// At EOF, HasMoreRows should be false.
	if ts.HasMoreRows() {
		t.Error("HasMoreRows should be false at EOF")
	}
	_ = c.Close()
}

func TestBaseDataSource_Prior(t *testing.T) {
	c := data.NewDataConnectionBase("stub")
	c.ConnectionString = "stub://test"
	ts := c.CreateTable("t")
	ts.SetSelectCommand("SELECT id, name FROM users")
	if err := ts.Init(); err != nil {
		t.Fatalf("Init: %v", err)
	}
	// Move to second row, then Prior.
	_ = ts.First()
	_ = ts.Next()
	rowBefore := ts.CurrentRowNo()
	ts.Prior()
	if ts.CurrentRowNo() != rowBefore-1 {
		t.Errorf("Prior: CurrentRowNo = %d, want %d", ts.CurrentRowNo(), rowBefore-1)
	}
	_ = c.Close()
}

func TestBaseDataSource_SetCurrentRowNo(t *testing.T) {
	c := data.NewDataConnectionBase("stub")
	c.ConnectionString = "stub://test"
	ts := c.CreateTable("t")
	ts.SetSelectCommand("SELECT id, name FROM users")
	if err := ts.Init(); err != nil {
		t.Fatalf("Init: %v", err)
	}
	ts.SetCurrentRowNo(1)
	if ts.CurrentRowNo() != 1 {
		t.Errorf("SetCurrentRowNo: CurrentRowNo = %d, want 1", ts.CurrentRowNo())
	}
	_ = c.Close()
}

func TestBaseDataSource_EnsureInit_CallsInit(t *testing.T) {
	ts := data.NewTableDataSource("standalone")
	// Not yet initialized — First() should return ErrNotInitialized.
	if err := ts.First(); err != data.ErrNotInitialized {
		t.Fatalf("First() before EnsureInit: want ErrNotInitialized, got %v", err)
	}
	// EnsureInit calls BaseDataSource.Init() via embedding, marking initialized=true.
	if err := ts.EnsureInit(); err != nil {
		t.Fatalf("EnsureInit: %v", err)
	}
	// After EnsureInit, First() must not return ErrNotInitialized (returns ErrEOF instead).
	if err := ts.First(); err == data.ErrNotInitialized {
		t.Error("First() after EnsureInit must not return ErrNotInitialized")
	}
}

func TestBaseDataSource_EnsureInit_Idempotent(t *testing.T) {
	c := data.NewDataConnectionBase("stub")
	c.ConnectionString = "stub://test"
	ts := c.CreateTable("t")
	ts.SetSelectCommand("SELECT id, name FROM users")
	if err := ts.Init(); err != nil {
		t.Fatalf("Init: %v", err)
	}
	// Already initialized — EnsureInit should be a no-op.
	if err := ts.EnsureInit(); err != nil {
		t.Fatalf("EnsureInit (already init): %v", err)
	}
	_ = c.Close()
}

func TestBaseDataSource_GetDisplayName_UsesAlias(t *testing.T) {
	ts := data.NewTableDataSource("Orders")
	ts.SetAlias("MyOrders")
	if got := ts.GetDisplayName(); got != "MyOrders" {
		t.Errorf("GetDisplayName = %q, want 'MyOrders'", got)
	}
}

func TestBaseDataSource_GetDisplayName_FallsBackToName(t *testing.T) {
	ts := data.NewTableDataSource("Products")
	// No alias set.
	if got := ts.GetDisplayName(); got != "Products" {
		t.Errorf("GetDisplayName = %q, want 'Products'", got)
	}
}

// ── test doubles ─────────────────────────────────────────────────────────────

type tdsWriter struct {
	strs  map[string]string
	bools map[string]bool
}

func newTDSWriter() *tdsWriter {
	return &tdsWriter{strs: make(map[string]string), bools: make(map[string]bool)}
}

func (w *tdsWriter) WriteStr(name, value string) { w.strs[name] = value }
func (w *tdsWriter) WriteInt(name string, value int) {}
func (w *tdsWriter) WriteBool(name string, value bool) { w.bools[name] = value }
func (w *tdsWriter) WriteFloat(name string, value float32) {}
func (w *tdsWriter) WriteObject(obj report.Serializable) error { return nil }
func (w *tdsWriter) WriteObjectNamed(name string, obj report.Serializable) error { return nil }

type tdsReader struct {
	strs  map[string]string
	bools map[string]bool
}

func (r *tdsReader) ReadStr(name, def string) string {
	if v, ok := r.strs[name]; ok {
		return v
	}
	return def
}
func (r *tdsReader) ReadInt(name string, def int) int { return def }
func (r *tdsReader) ReadBool(name string, def bool) bool {
	if v, ok := r.bools[name]; ok {
		return v
	}
	return def
}
func (r *tdsReader) ReadFloat(name string, def float32) float32 { return def }
func (r *tdsReader) NextChild() (string, bool)                  { return "", false }
func (r *tdsReader) FinishChild() error                         { return nil }
