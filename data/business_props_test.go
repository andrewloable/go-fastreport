package data_test

// business_props_test.go — tests for BusinessObjectDataSource property methods
// that had 0% coverage: Enabled, SetEnabled, SetName, PropName, SetPropName,
// ReferenceName, SetReferenceName, Serialize, Deserialize.

import (
	"testing"

	"github.com/andrewloable/go-fastreport/data"
	"github.com/andrewloable/go-fastreport/report"
)

// ── Enabled / SetEnabled ──────────────────────────────────────────────────────

func TestBusinessObjectDataSource_Enabled_DefaultTrue(t *testing.T) {
	ds := data.NewBusinessObjectDataSource("test", nil)
	if !ds.Enabled() {
		t.Error("Enabled should default to true")
	}
}

func TestBusinessObjectDataSource_SetEnabled(t *testing.T) {
	ds := data.NewBusinessObjectDataSource("test", nil)
	ds.SetEnabled(false)
	if ds.Enabled() {
		t.Error("Enabled should be false after SetEnabled(false)")
	}
	ds.SetEnabled(true)
	if !ds.Enabled() {
		t.Error("Enabled should be true after SetEnabled(true)")
	}
}

// ── SetName ───────────────────────────────────────────────────────────────────

func TestBusinessObjectDataSource_SetName_SyncsAlias(t *testing.T) {
	ds := data.NewBusinessObjectDataSource("OldName", nil)
	// When alias was equal to name, SetName should sync alias to new name.
	ds.SetName("NewName")
	if ds.Name() != "NewName" {
		t.Errorf("Name = %q, want 'NewName'", ds.Name())
	}
	// Alias should also be updated (it was "OldName" == name).
	if ds.Alias() != "NewName" {
		t.Errorf("Alias = %q, want 'NewName' (synced with name)", ds.Alias())
	}
}

func TestBusinessObjectDataSource_SetName_PreservesCustomAlias(t *testing.T) {
	ds := data.NewBusinessObjectDataSource("Orders", nil)
	ds.SetAlias("OrdersAlias") // explicitly set a different alias
	ds.SetName("Invoices")
	if ds.Name() != "Invoices" {
		t.Errorf("Name = %q, want 'Invoices'", ds.Name())
	}
	// Alias must NOT be synced since it was different from name.
	if ds.Alias() != "OrdersAlias" {
		t.Errorf("Alias = %q, want 'OrdersAlias' (custom alias preserved)", ds.Alias())
	}
}

// ── PropName / SetPropName ────────────────────────────────────────────────────

func TestBusinessObjectDataSource_PropName_Default(t *testing.T) {
	ds := data.NewBusinessObjectDataSource("test", nil)
	if ds.PropName() != "" {
		t.Errorf("PropName = %q, want empty default", ds.PropName())
	}
}

func TestBusinessObjectDataSource_SetPropName(t *testing.T) {
	ds := data.NewBusinessObjectDataSource("test", nil)
	ds.SetPropName("Items")
	if ds.PropName() != "Items" {
		t.Errorf("PropName = %q, want 'Items'", ds.PropName())
	}
}

// ── ReferenceName / SetReferenceName ─────────────────────────────────────────

func TestBusinessObjectDataSource_ReferenceName_Default(t *testing.T) {
	ds := data.NewBusinessObjectDataSource("test", nil)
	if ds.ReferenceName() != "" {
		t.Errorf("ReferenceName = %q, want empty default", ds.ReferenceName())
	}
}

func TestBusinessObjectDataSource_SetReferenceName(t *testing.T) {
	ds := data.NewBusinessObjectDataSource("test", nil)
	ds.SetReferenceName("Orders.Items")
	if ds.ReferenceName() != "Orders.Items" {
		t.Errorf("ReferenceName = %q, want 'Orders.Items'", ds.ReferenceName())
	}
}

// ── Serialize ────────────────────────────────────────────────────────────────

func TestBusinessObjectDataSource_Serialize_BasicProperties(t *testing.T) {
	ds := data.NewBusinessObjectDataSource("Orders", nil)
	ds.SetPropName("Items")
	ds.SetReferenceName("Root.Orders")
	ds.SetEnabled(false)

	w := &bizWriter{}
	if err := ds.Serialize(w); err != nil {
		t.Fatalf("Serialize: %v", err)
	}
	if w.strs["Name"] != "Orders" {
		t.Errorf("Name = %q, want 'Orders'", w.strs["Name"])
	}
	if w.strs["PropName"] != "Items" {
		t.Errorf("PropName = %q, want 'Items'", w.strs["PropName"])
	}
	if w.strs["ReferenceName"] != "Root.Orders" {
		t.Errorf("ReferenceName = %q, want 'Root.Orders'", w.strs["ReferenceName"])
	}
	if v, ok := w.bools["Enabled"]; !ok || v {
		t.Error("Enabled should be written as false")
	}
}

func TestBusinessObjectDataSource_Serialize_EnabledNotWrittenWhenTrue(t *testing.T) {
	ds := data.NewBusinessObjectDataSource("test", nil)
	// Enabled is true by default — Serialize should not write it.
	w := &bizWriter{}
	_ = ds.Serialize(w)
	if _, ok := w.bools["Enabled"]; ok {
		t.Error("Enabled should not be written when true (it's the default)")
	}
}

func TestBusinessObjectDataSource_Serialize_AliasWrittenWhenDifferent(t *testing.T) {
	ds := data.NewBusinessObjectDataSource("Orders", nil)
	ds.SetAlias("MyOrders")
	w := &bizWriter{}
	_ = ds.Serialize(w)
	if w.strs["Alias"] != "MyOrders" {
		t.Errorf("Alias = %q, want 'MyOrders'", w.strs["Alias"])
	}
}

// ── Deserialize ───────────────────────────────────────────────────────────────

func TestBusinessObjectDataSource_Deserialize_BasicProperties(t *testing.T) {
	ds := data.NewBusinessObjectDataSource("", nil)
	r := &bizReader{
		strs: map[string]string{
			"Name":          "Products",
			"PropName":      "Items",
			"ReferenceName": "",
		},
		bools: map[string]bool{
			"Enabled": false,
		},
	}
	if err := ds.Deserialize(r); err != nil {
		t.Fatalf("Deserialize: %v", err)
	}
	if ds.Name() != "Products" {
		t.Errorf("Name = %q, want 'Products'", ds.Name())
	}
	if ds.PropName() != "Items" {
		t.Errorf("PropName = %q, want 'Items'", ds.PropName())
	}
	if ds.Enabled() {
		t.Error("Enabled should be false after Deserialize")
	}
}

func TestBusinessObjectDataSource_Deserialize_LegacyReferenceName(t *testing.T) {
	// When ReferenceName contains a dot, the last segment becomes PropName
	// and ReferenceName is cleared.
	ds := data.NewBusinessObjectDataSource("", nil)
	r := &bizReader{
		strs: map[string]string{
			"Name":          "DS",
			"ReferenceName": "Root.SubData.Items",
		},
		bools: map[string]bool{"Enabled": true},
	}
	if err := ds.Deserialize(r); err != nil {
		t.Fatalf("Deserialize: %v", err)
	}
	if ds.PropName() != "Items" {
		t.Errorf("PropName = %q, want 'Items' (extracted from legacy ReferenceName)", ds.PropName())
	}
	if ds.ReferenceName() != "" {
		t.Errorf("ReferenceName = %q, want '' (cleared after legacy split)", ds.ReferenceName())
	}
}

// ── test doubles ─────────────────────────────────────────────────────────────

type bizWriter struct {
	strs  map[string]string
	bools map[string]bool
}

func (w *bizWriter) WriteStr(name, value string) {
	if w.strs == nil {
		w.strs = make(map[string]string)
	}
	w.strs[name] = value
}
func (w *bizWriter) WriteInt(name string, value int)                             {}
func (w *bizWriter) WriteBool(name string, value bool) {
	if w.bools == nil {
		w.bools = make(map[string]bool)
	}
	w.bools[name] = value
}
func (w *bizWriter) WriteFloat(name string, value float32)                       {}
func (w *bizWriter) WriteObject(obj report.Serializable) error                   { return nil }
func (w *bizWriter) WriteObjectNamed(name string, obj report.Serializable) error { return nil }

type bizReader struct {
	strs  map[string]string
	bools map[string]bool
}

func (r *bizReader) ReadStr(name, def string) string {
	if v, ok := r.strs[name]; ok {
		return v
	}
	return def
}
func (r *bizReader) ReadInt(name string, def int) int { return def }
func (r *bizReader) ReadBool(name string, def bool) bool {
	if v, ok := r.bools[name]; ok {
		return v
	}
	return def
}
func (r *bizReader) ReadFloat(name string, def float32) float32 { return def }
func (r *bizReader) NextChild() (string, bool)                  { return "", false }
func (r *bizReader) FinishChild() error                         { return nil }
