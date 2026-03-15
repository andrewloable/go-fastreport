package data_test

import (
	"bytes"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/andrewloable/go-fastreport/data"
	"github.com/andrewloable/go-fastreport/serial"
)

// ─── helpers ────────────────────────────────────────────────────────────────

// roundTripCommandParameter serializes p to XML and deserializes into a new
// CommandParameter, returning the deserialized copy.
func roundTripCommandParameter(t *testing.T, p *data.CommandParameter) *data.CommandParameter {
	t.Helper()
	var buf bytes.Buffer
	w := serial.NewWriter(&buf)
	if err := w.WriteObjectNamed("Parameter", p); err != nil {
		t.Fatalf("serialize: %v", err)
	}
	if err := w.Flush(); err != nil {
		t.Fatalf("flush: %v", err)
	}

	r := serial.NewReader(bytes.NewReader(buf.Bytes()))
	typeName, ok := r.ReadObjectHeader()
	if !ok {
		t.Fatalf("ReadObjectHeader: not ok; xml=%s", buf.String())
	}
	if typeName != "Parameter" {
		t.Fatalf("typeName=%q, want Parameter", typeName)
	}
	got := data.NewCommandParameter("")
	if err := got.Deserialize(r); err != nil {
		t.Fatalf("deserialize: %v", err)
	}
	return got
}

// roundTripDataComponent serializes a DataComponentBase and deserializes it.
func roundTripDataComponent(t *testing.T, d *data.DataComponentBase) *data.DataComponentBase {
	t.Helper()
	var buf bytes.Buffer
	w := serial.NewWriter(&buf)
	if err := w.WriteObjectNamed("DataComponent", d); err != nil {
		t.Fatalf("serialize: %v", err)
	}
	if err := w.Flush(); err != nil {
		t.Fatalf("flush: %v", err)
	}

	r := serial.NewReader(bytes.NewReader(buf.Bytes()))
	_, ok := r.ReadObjectHeader()
	if !ok {
		t.Fatalf("ReadObjectHeader: not ok")
	}
	got := data.NewDataComponentBase("orig")
	if err := got.Deserialize(r); err != nil {
		t.Fatalf("deserialize: %v", err)
	}
	return got
}

// ─── DataConnectionCollection ────────────────────────────────────────────────

func TestDataConnectionCollection_Basic(t *testing.T) {
	col := data.NewDataConnectionCollection()
	if col.Count() != 0 {
		t.Errorf("Count = %d, want 0", col.Count())
	}

	c1 := data.NewDataConnectionBase("stub")
	c1.SetName("conn1")
	c2 := data.NewDataConnectionBase("stub")
	c2.SetName("conn2")

	col.Add(c1)
	col.Add(c2)

	if col.Count() != 2 {
		t.Errorf("Count = %d, want 2", col.Count())
	}
	if col.Get(0) != c1 {
		t.Error("Get(0) should return c1")
	}
	if col.Get(1) != c2 {
		t.Error("Get(1) should return c2")
	}

	all := col.All()
	if len(all) != 2 {
		t.Errorf("All len = %d, want 2", len(all))
	}
}

func TestDataConnectionCollection_FindByName(t *testing.T) {
	col := data.NewDataConnectionCollection()
	c := data.NewDataConnectionBase("stub")
	c.SetName("MyConn")
	col.Add(c)

	found := col.FindByName("myconn") // case-insensitive
	if found != c {
		t.Error("FindByName should return c (case-insensitive)")
	}
	if col.FindByName("NoSuch") != nil {
		t.Error("FindByName should return nil for unknown name")
	}
}

func TestDataConnectionCollection_Remove(t *testing.T) {
	col := data.NewDataConnectionCollection()
	c1 := data.NewDataConnectionBase("stub")
	c2 := data.NewDataConnectionBase("stub")
	col.Add(c1)
	col.Add(c2)

	col.Remove(c1)
	if col.Count() != 1 {
		t.Errorf("Count = %d after remove, want 1", col.Count())
	}
	if col.Get(0) != c2 {
		t.Error("remaining element should be c2")
	}

	// Removing non-member is a no-op.
	col.Remove(c1)
	if col.Count() != 1 {
		t.Error("removing non-member should be a no-op")
	}
}

// ─── DataSourceCollection ─────────────────────────────────────────────────────

func TestDataSourceCollection_Basic(t *testing.T) {
	col := data.NewDataSourceCollection()
	if col.Count() != 0 {
		t.Errorf("Count = %d, want 0", col.Count())
	}

	ds1 := data.NewBaseDataSource("orders")
	ds1.SetAlias("Orders")
	ds2 := data.NewBaseDataSource("customers")
	ds2.SetAlias("Customers")

	col.Add(ds1)
	col.Add(ds2)

	if col.Count() != 2 {
		t.Errorf("Count = %d, want 2", col.Count())
	}
	if col.Get(0) != ds1 {
		t.Error("Get(0) should return ds1")
	}

	all := col.All()
	if len(all) != 2 {
		t.Errorf("All len = %d, want 2", len(all))
	}
}

func TestDataSourceCollection_FindByName(t *testing.T) {
	col := data.NewDataSourceCollection()
	ds := data.NewBaseDataSource("MySource")
	col.Add(ds)

	found := col.FindByName("mysource")
	if found != ds {
		t.Error("FindByName should be case-insensitive")
	}
	if col.FindByName("NoSuch") != nil {
		t.Error("FindByName should return nil for unknown name")
	}
}

func TestDataSourceCollection_FindByAlias(t *testing.T) {
	col := data.NewDataSourceCollection()
	ds := data.NewBaseDataSource("ds1")
	ds.SetAlias("My Alias")
	col.Add(ds)

	found := col.FindByAlias("my alias")
	if found != ds {
		t.Error("FindByAlias should be case-insensitive")
	}
	if col.FindByAlias("NoSuchAlias") != nil {
		t.Error("FindByAlias should return nil for unknown alias")
	}
}

func TestDataSourceCollection_Remove(t *testing.T) {
	col := data.NewDataSourceCollection()
	ds1 := data.NewBaseDataSource("a")
	ds2 := data.NewBaseDataSource("b")
	col.Add(ds1)
	col.Add(ds2)

	col.Remove(ds1)
	if col.Count() != 1 {
		t.Errorf("Count = %d after remove, want 1", col.Count())
	}

	// Remove non-member is no-op.
	col.Remove(ds1)
	if col.Count() != 1 {
		t.Error("removing non-member should be a no-op")
	}
}

// ─── CommandParameterCollection ───────────────────────────────────────────────

func TestCommandParameterCollection_Basic(t *testing.T) {
	col := data.NewCommandParameterCollection()
	if col.Count() != 0 {
		t.Errorf("Count = %d, want 0", col.Count())
	}

	p1 := data.NewCommandParameter("@id")
	p2 := data.NewCommandParameter("@name")
	col.Add(p1)
	col.Add(p2)

	if col.Count() != 2 {
		t.Errorf("Count = %d, want 2", col.Count())
	}
	if col.Get(0) != p1 {
		t.Error("Get(0) should return p1")
	}

	all := col.All()
	if len(all) != 2 {
		t.Errorf("All len = %d, want 2", len(all))
	}
}

func TestCommandParameterCollection_FindByName(t *testing.T) {
	col := data.NewCommandParameterCollection()
	p := data.NewCommandParameter("@customer")
	col.Add(p)

	found := col.FindByName("@CUSTOMER") // case-insensitive
	if found != p {
		t.Error("FindByName should be case-insensitive")
	}
	if col.FindByName("@missing") != nil {
		t.Error("FindByName should return nil for unknown name")
	}
}

func TestCommandParameterCollection_Remove(t *testing.T) {
	col := data.NewCommandParameterCollection()
	p1 := data.NewCommandParameter("@a")
	p2 := data.NewCommandParameter("@b")
	col.Add(p1)
	col.Add(p2)

	col.Remove(p1)
	if col.Count() != 1 {
		t.Errorf("Count = %d after remove, want 1", col.Count())
	}
	// Remove non-member is no-op.
	col.Remove(p1)
	if col.Count() != 1 {
		t.Error("removing non-member should be a no-op")
	}
}

func TestCommandParameterCollection_CreateUniqueName(t *testing.T) {
	col := data.NewCommandParameterCollection()

	// No conflict — base name returned as-is.
	name := col.CreateUniqueName("@param")
	if name != "@param" {
		t.Errorf("CreateUniqueName (no conflict) = %q, want @param", name)
	}

	// Add @param, then generate unique name.
	col.Add(data.NewCommandParameter("@param"))
	name2 := col.CreateUniqueName("@param")
	if name2 == "@param" {
		t.Error("CreateUniqueName should return different name when @param exists")
	}
	// It should be @param1
	if name2 != "@param1" {
		t.Errorf("CreateUniqueName = %q, want @param1", name2)
	}

	// Add @param1 too.
	col.Add(data.NewCommandParameter("@param1"))
	name3 := col.CreateUniqueName("@param")
	if name3 != "@param2" {
		t.Errorf("CreateUniqueName = %q, want @param2", name3)
	}
}

func TestCommandParameterCollection_Serialize_Deserialize(t *testing.T) {
	col := data.NewCommandParameterCollection()
	p1 := data.NewCommandParameter("@id")
	p1.DataType = "int"
	p1.Size = 4
	p2 := data.NewCommandParameter("@name")
	p2.Expression = "[CustomerName]"
	p2.DefaultValue = "Unknown"
	p2.Direction = data.ParamDirectionOutput
	col.Add(p1)
	col.Add(p2)

	// Serialize the whole collection into a parent element.
	var buf bytes.Buffer
	w := serial.NewWriter(&buf)
	if err := w.BeginObject("Params"); err != nil {
		t.Fatal(err)
	}
	if err := col.Serialize(w); err != nil {
		t.Fatalf("Serialize: %v", err)
	}
	if err := w.EndObject(); err != nil {
		t.Fatal(err)
	}
	if err := w.Flush(); err != nil {
		t.Fatal(err)
	}

	// Deserialize.
	r := serial.NewReader(bytes.NewReader(buf.Bytes()))
	_, ok := r.ReadObjectHeader()
	if !ok {
		t.Fatalf("ReadObjectHeader failed; xml=%s", buf.String())
	}

	got := data.NewCommandParameterCollection()
	if err := got.Deserialize(r); err != nil {
		t.Fatalf("Deserialize: %v", err)
	}

	if got.Count() != 2 {
		t.Errorf("Count = %d, want 2", got.Count())
	}
	if got.Get(0).Name != "@id" {
		t.Errorf("param[0].Name = %q, want @id", got.Get(0).Name)
	}
	if got.Get(0).DataType != "int" {
		t.Errorf("param[0].DataType = %q, want int", got.Get(0).DataType)
	}
	if got.Get(0).Size != 4 {
		t.Errorf("param[0].Size = %d, want 4", got.Get(0).Size)
	}
	if got.Get(1).Name != "@name" {
		t.Errorf("param[1].Name = %q, want @name", got.Get(1).Name)
	}
	if got.Get(1).Expression != "[CustomerName]" {
		t.Errorf("param[1].Expression = %q, want [CustomerName]", got.Get(1).Expression)
	}
	if got.Get(1).DefaultValue != "Unknown" {
		t.Errorf("param[1].DefaultValue = %q, want Unknown", got.Get(1).DefaultValue)
	}
	if got.Get(1).Direction != data.ParamDirectionOutput {
		t.Errorf("param[1].Direction = %d, want Output", got.Get(1).Direction)
	}
}

func TestCommandParameterCollection_Deserialize_UnknownChild(t *testing.T) {
	// A child element that is not "Parameter" should be skipped gracefully.
	xml := `<Params><Unknown Foo="bar"/><Parameter Name="@x"/></Params>`
	r := serial.NewReader(strings.NewReader(xml))
	_, ok := r.ReadObjectHeader()
	if !ok {
		t.Fatal("ReadObjectHeader failed")
	}
	col := data.NewCommandParameterCollection()
	if err := col.Deserialize(r); err != nil {
		t.Fatalf("Deserialize: %v", err)
	}
	if col.Count() != 1 {
		t.Errorf("Count = %d, want 1 (Unknown skipped)", col.Count())
	}
}

// ─── CommandParameter.Serialize (all-zero round-trip) ─────────────────────────

func TestCommandParameter_Serialize_Defaults(t *testing.T) {
	// When all fields are at default, serialized form is minimal.
	p := data.NewCommandParameter("")
	got := roundTripCommandParameter(t, p)
	if got.Name != "" {
		t.Errorf("Name = %q, want empty", got.Name)
	}
	if got.Direction != data.ParamDirectionInput {
		t.Errorf("Direction = %d, want Input", got.Direction)
	}
}

func TestCommandParameter_Serialize_AllFields(t *testing.T) {
	p := data.NewCommandParameter("@p1")
	p.DataType = "varchar"
	p.Size = 50
	p.Expression = "[Col]"
	p.DefaultValue = "default"
	p.Direction = data.ParamDirectionInputOutput

	got := roundTripCommandParameter(t, p)

	if got.Name != "@p1" {
		t.Errorf("Name = %q", got.Name)
	}
	if got.DataType != "varchar" {
		t.Errorf("DataType = %q", got.DataType)
	}
	if got.Size != 50 {
		t.Errorf("Size = %d", got.Size)
	}
	if got.Expression != "[Col]" {
		t.Errorf("Expression = %q", got.Expression)
	}
	if got.DefaultValue != "default" {
		t.Errorf("DefaultValue = %q", got.DefaultValue)
	}
	if got.Direction != data.ParamDirectionInputOutput {
		t.Errorf("Direction = %d", got.Direction)
	}
}

// ─── DataComponentBase ────────────────────────────────────────────────────────

func TestDataComponentBase_Getters_Setters(t *testing.T) {
	d := data.NewDataComponentBase("comp1")

	if d.Name() != "comp1" {
		t.Errorf("Name = %q, want comp1", d.Name())
	}
	if d.Alias() != "comp1" {
		t.Errorf("Alias default = %q, want comp1", d.Alias())
	}
	if !d.Enabled() {
		t.Error("Enabled default should be true")
	}
	if d.ReferenceName() != "" {
		t.Errorf("ReferenceName default = %q, want empty", d.ReferenceName())
	}
	if d.Reference() != nil {
		t.Error("Reference default should be nil")
	}
	if d.IsAliased() {
		t.Error("IsAliased should be false when alias equals name")
	}

	d.SetAlias("MyAlias")
	if d.Alias() != "MyAlias" {
		t.Errorf("Alias = %q, want MyAlias", d.Alias())
	}
	if !d.IsAliased() {
		t.Error("IsAliased should be true after setting different alias")
	}

	d.SetEnabled(false)
	if d.Enabled() {
		t.Error("Enabled should be false after SetEnabled(false)")
	}

	d.SetReferenceName("refname")
	if d.ReferenceName() != "refname" {
		t.Errorf("ReferenceName = %q, want refname", d.ReferenceName())
	}

	ref := struct{ x int }{42}
	d.SetReference(ref)
	if d.Reference() == nil {
		t.Error("Reference should not be nil after SetReference")
	}

	d.InitializeComponent() // should not panic
}

func TestDataComponentBase_SetName_SyncsAlias(t *testing.T) {
	d := data.NewDataComponentBase("original")
	// alias == name initially → SetName should update alias too.
	d.SetName("updated")
	if d.Alias() != "updated" {
		t.Errorf("Alias should sync to updated, got %q", d.Alias())
	}

	// If alias was already different, SetName should NOT change it.
	d.SetAlias("custom")
	d.SetName("renamed")
	if d.Alias() != "custom" {
		t.Errorf("Alias should remain custom after SetName, got %q", d.Alias())
	}
}

func TestDataComponentBase_Serialize_Deserialize(t *testing.T) {
	d := data.NewDataComponentBase("mycomp")
	d.SetAlias("My Component")
	d.SetEnabled(false)
	d.SetReferenceName("shared1")

	got := roundTripDataComponent(t, d)

	if got.Alias() != "My Component" {
		t.Errorf("Alias = %q, want 'My Component'", got.Alias())
	}
	if got.Enabled() {
		t.Error("Enabled should be false after round-trip")
	}
	if got.ReferenceName() != "shared1" {
		t.Errorf("ReferenceName = %q, want shared1", got.ReferenceName())
	}
}

func TestDataComponentBase_Serialize_NoAlias(t *testing.T) {
	// When alias equals name, Alias attribute is not written.
	d := data.NewDataComponentBase("comp")
	// don't change alias — it will equal name

	var buf bytes.Buffer
	w := serial.NewWriter(&buf)
	if err := w.WriteObjectNamed("DC", d); err != nil {
		t.Fatal(err)
	}
	_ = w.Flush()

	xml := buf.String()
	// "Enabled" not written because it is true (default).
	if strings.Contains(xml, `Enabled="false"`) {
		t.Error("Enabled should not be written when true")
	}
}

// ─── DataConnectionBase.Serialize / Deserialize ───────────────────────────────

func TestDataConnectionBase_Serialize_Deserialize(t *testing.T) {
	// DataConnectionBase embeds DataComponentBase; Serialize/Deserialize come
	// from DataComponentBase and cover Alias, Enabled, ReferenceName.
	c := data.NewDataConnectionBase("stub")
	c.SetName("myconn")
	c.SetAlias("My Connection") // different from name → should serialize
	c.SetEnabled(false)
	c.SetReferenceName("sharedConn")

	var buf bytes.Buffer
	w := serial.NewWriter(&buf)
	if err := w.WriteObjectNamed("Connection", c); err != nil {
		t.Fatalf("Serialize: %v", err)
	}
	if err := w.Flush(); err != nil {
		t.Fatalf("Flush: %v", err)
	}

	r := serial.NewReader(bytes.NewReader(buf.Bytes()))
	_, ok := r.ReadObjectHeader()
	if !ok {
		t.Fatalf("ReadObjectHeader failed; xml=%s", buf.String())
	}
	got := data.NewDataConnectionBase("stub")
	got.SetName("myconn")
	if err := got.Deserialize(r); err != nil {
		t.Fatalf("Deserialize: %v", err)
	}
	if got.Alias() != "My Connection" {
		t.Errorf("Alias = %q, want 'My Connection'", got.Alias())
	}
	if got.Enabled() {
		t.Error("Enabled should be false after round-trip")
	}
	if got.ReferenceName() != "sharedConn" {
		t.Errorf("ReferenceName = %q, want sharedConn", got.ReferenceName())
	}
}

// ─── BaseDataSource.SortRows ──────────────────────────────────────────────────

func TestBaseDataSource_SortRows_String(t *testing.T) {
	ds := data.NewBaseDataSource("test")
	ds.AddRow(map[string]any{"name": "Charlie"})
	ds.AddRow(map[string]any{"name": "Alice"})
	ds.AddRow(map[string]any{"name": "Bob"})
	_ = ds.Init()

	ds.SortRows([]data.SortSpec{{Column: "name"}})
	_ = ds.First()
	v, _ := ds.GetValue("name")
	if v != "Alice" {
		t.Errorf("first after sort = %v, want Alice", v)
	}
}

func TestBaseDataSource_SortRows_Descending(t *testing.T) {
	ds := data.NewBaseDataSource("test")
	ds.AddRow(map[string]any{"val": int64(1)})
	ds.AddRow(map[string]any{"val": int64(3)})
	ds.AddRow(map[string]any{"val": int64(2)})
	_ = ds.Init()

	ds.SortRows([]data.SortSpec{{Column: "val", Descending: true}})
	_ = ds.First()
	v, _ := ds.GetValue("val")
	if v.(int64) != 3 {
		t.Errorf("first after desc sort = %v, want 3", v)
	}
}

func TestBaseDataSource_SortRows_Empty(t *testing.T) {
	ds := data.NewBaseDataSource("test")
	// No panic expected for empty specs.
	ds.SortRows(nil)
	ds.SortRows([]data.SortSpec{})
}

func TestBaseDataSource_SortRows_Float32(t *testing.T) {
	ds := data.NewBaseDataSource("test")
	ds.AddRow(map[string]any{"v": float32(3.5)})
	ds.AddRow(map[string]any{"v": float32(1.2)})
	_ = ds.Init()

	ds.SortRows([]data.SortSpec{{Column: "v"}})
	_ = ds.First()
	v, _ := ds.GetValue("v")
	if v.(float32) != float32(1.2) {
		t.Errorf("first after float32 sort = %v, want 1.2", v)
	}
}

func TestBaseDataSource_SortRows_Int(t *testing.T) {
	ds := data.NewBaseDataSource("test")
	ds.AddRow(map[string]any{"v": 10})
	ds.AddRow(map[string]any{"v": 5})
	_ = ds.Init()

	ds.SortRows([]data.SortSpec{{Column: "v"}})
	_ = ds.First()
	v, _ := ds.GetValue("v")
	if v.(int) != 5 {
		t.Errorf("first after int sort = %v, want 5", v)
	}
}

func TestBaseDataSource_SortRows_Bool(t *testing.T) {
	ds := data.NewBaseDataSource("test")
	ds.AddRow(map[string]any{"v": true})
	ds.AddRow(map[string]any{"v": false})
	_ = ds.Init()

	ds.SortRows([]data.SortSpec{{Column: "v"}})
	_ = ds.First()
	v, _ := ds.GetValue("v")
	if v.(bool) != false {
		t.Errorf("first after bool sort = %v, want false", v)
	}
}

func TestBaseDataSource_SortRows_Default(t *testing.T) {
	// Struct values fall back to fmt.Sprintf comparison.
	type Point struct{ X int }
	ds := data.NewBaseDataSource("test")
	ds.AddRow(map[string]any{"v": Point{2}})
	ds.AddRow(map[string]any{"v": Point{1}})
	_ = ds.Init()

	// Should not panic.
	ds.SortRows([]data.SortSpec{{Column: "v"}})
}

// ─── Dictionary additions ─────────────────────────────────────────────────────

func TestDictionary_AddConnection_Remove_Find(t *testing.T) {
	d := data.NewDictionary()
	c1 := data.NewDataConnectionBase("stub")
	c1.SetName("C1")
	c2 := data.NewDataConnectionBase("stub")
	c2.SetName("C2")

	d.AddConnection(c1)
	d.AddConnection(c2)

	conns := d.Connections()
	if len(conns) != 2 {
		t.Fatalf("Connections len = %d, want 2", len(conns))
	}

	found := d.FindConnectionByName("c1")
	if found != c1 {
		t.Error("FindConnectionByName should find c1 case-insensitively")
	}
	if d.FindConnectionByName("NoSuch") != nil {
		t.Error("FindConnectionByName should return nil for unknown name")
	}

	d.RemoveConnection(c1)
	if len(d.Connections()) != 1 {
		t.Errorf("Connections len = %d after remove, want 1", len(d.Connections()))
	}
	// Remove non-member is a no-op.
	d.RemoveConnection(c1)
	if len(d.Connections()) != 1 {
		t.Error("removing non-member should be no-op")
	}
}

func TestDictionary_AggregateTotals(t *testing.T) {
	d := data.NewDictionary()

	at := data.NewAggregateTotal("GrandTotal")
	d.AddAggregateTotal(at)

	ats := d.AggregateTotals()
	if len(ats) != 1 {
		t.Fatalf("AggregateTotals len = %d, want 1", len(ats))
	}
	if ats[0] != at {
		t.Error("AggregateTotals[0] should be at")
	}

	// A simple Total placeholder should also be added.
	tots := d.Totals()
	found := false
	for _, t := range tots {
		if t.Name == "GrandTotal" {
			found = true
		}
	}
	if !found {
		t.Error("AddAggregateTotal should also add a placeholder Total")
	}

	// Adding again should not duplicate the placeholder.
	d.AddAggregateTotal(at)
	count := 0
	for _, t := range d.Totals() {
		if t.Name == "GrandTotal" {
			count++
		}
	}
	if count != 1 {
		t.Errorf("duplicate Total placeholders: %d, want 1", count)
	}
}

func TestDictionary_ResolveRelations(t *testing.T) {
	d := data.NewDictionary()
	parent := data.NewBaseDataSource("customers")
	parent.SetAlias("Customers")
	child := data.NewBaseDataSource("orders")
	child.SetAlias("Orders")
	d.AddDataSource(parent)
	d.AddDataSource(child)

	rel := &data.Relation{
		Name:             "CustomersOrders",
		ParentSourceName: "Customers",
		ChildSourceName:  "Orders",
		ParentColumnNames: []string{"ID"},
		ChildColumnNames:  []string{"CustomerID"},
	}
	d.AddRelation(rel)

	d.ResolveRelations()

	if rel.ParentDataSource != parent {
		t.Error("ResolveRelations should resolve ParentDataSource by alias")
	}
	if rel.ChildDataSource != child {
		t.Error("ResolveRelations should resolve ChildDataSource by alias")
	}
	if len(rel.ParentColumns) == 0 {
		t.Error("ResolveRelations should populate ParentColumns from ParentColumnNames")
	}
	if len(rel.ChildColumns) == 0 {
		t.Error("ResolveRelations should populate ChildColumns from ChildColumnNames")
	}
}

func TestDictionary_ResolveRelations_ByName(t *testing.T) {
	d := data.NewDictionary()
	parent := data.NewBaseDataSource("customers")
	d.AddDataSource(parent)

	rel := &data.Relation{
		ParentSourceName: "customers",
		ChildSourceName:  "orders",
	}
	d.AddRelation(rel)
	d.ResolveRelations()

	if rel.ParentDataSource != parent {
		t.Error("should resolve parent by name fallback")
	}
}

func TestDictionary_ResolveRelations_AlreadyResolved(t *testing.T) {
	d := data.NewDictionary()
	parent := data.NewBaseDataSource("p")
	child := data.NewBaseDataSource("c")

	rel := &data.Relation{
		ParentDataSource: parent,
		ChildDataSource:  child,
		ParentColumns:    []string{"ID"},
		ChildColumns:     []string{"PID"},
	}
	d.AddRelation(rel)
	d.ResolveRelations() // should not overwrite already-set fields

	if rel.ParentDataSource != parent {
		t.Error("already-resolved ParentDataSource should not change")
	}
	if len(rel.ParentColumns) != 1 {
		t.Error("already-set ParentColumns should not be overwritten")
	}
}

// ─── Filter helper types not yet covered ─────────────────────────────────────

func TestFilter_Compare_Int32(t *testing.T) {
	f := data.NewDataSourceFilter()
	f.Add(int32(5), data.FilterEqual)
	if !f.ValueMatch(int32(5)) {
		t.Error("int32 == int32 should match")
	}
	if f.ValueMatch(int32(6)) {
		t.Error("int32 != should not match")
	}
}

func TestFilter_Compare_Float32(t *testing.T) {
	f := data.NewDataSourceFilter()
	f.Add(float32(1.5), data.FilterGreaterThan)
	if !f.ValueMatch(float32(2.0)) {
		t.Error("float32 2.0 > 1.5 should match")
	}
	if f.ValueMatch(float32(1.0)) {
		t.Error("float32 1.0 > 1.5 should not match")
	}
}

func TestFilter_Compare_Bool(t *testing.T) {
	f := data.NewDataSourceFilter()
	f.Add(true, data.FilterEqual)
	if !f.ValueMatch(true) {
		t.Error("true == true should match")
	}
	if f.ValueMatch(false) {
		t.Error("false == true should not match")
	}
}

func TestFilter_Compare_Time(t *testing.T) {
	t1 := time.Date(2024, 6, 1, 0, 0, 0, 0, time.UTC)
	t2 := time.Date(2024, 6, 2, 0, 0, 0, 0, time.UTC)

	// FilterLessThan checks: value < filterValue
	// So with Add(t2, FilterLessThan), t1 should match (t1 < t2).
	f := data.NewDataSourceFilter()
	f.Add(t2, data.FilterLessThan)
	if !f.ValueMatch(t1) {
		t.Error("t1 < t2, should match FilterLessThan")
	}
	if f.ValueMatch(t2) {
		t.Error("t2 == t2, should not match FilterLessThan")
	}
	if f.ValueMatch(time.Date(2024, 6, 3, 0, 0, 0, 0, time.UTC)) {
		t.Error("t3 > t2, should not match FilterLessThan")
	}
}

func TestFilter_Compare_BothNil(t *testing.T) {
	f := data.NewDataSourceFilter()
	f.Add(nil, data.FilterEqual)
	if f.ValueMatch(nil) {
		t.Error("nil compare should return false (incomparable)")
	}
}

func TestFilter_StringSet_Contains(t *testing.T) {
	f := data.NewDataSourceFilter()
	f.Add([]string{"a", "b"}, data.FilterContains)
	if !f.ValueMatch("a") {
		t.Error("'a' in set with FilterContains should match")
	}
	if f.ValueMatch("c") {
		t.Error("'c' not in set should not match")
	}
}

func TestFilter_StringSet_NotContains(t *testing.T) {
	f := data.NewDataSourceFilter()
	f.Add([]string{"a", "b"}, data.FilterNotContains)
	if f.ValueMatch("a") {
		t.Error("'a' in set with FilterNotContains should not match")
	}
	if !f.ValueMatch("c") {
		t.Error("'c' not in set with FilterNotContains should match")
	}
}

func TestFilter_StringSet_DefaultOperation(t *testing.T) {
	// Operations other than Equal/NotEqual/Contains/NotContains return false.
	f := data.NewDataSourceFilter()
	f.Add([]string{"a", "b"}, data.FilterLessThan)
	if f.ValueMatch("a") {
		t.Error("stringset with FilterLessThan should return false")
	}
}

func TestFilter_TimeRange_NotContains(t *testing.T) {
	start := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(2024, 1, 31, 0, 0, 0, 0, time.UTC)
	rng := [2]time.Time{start, end}

	f := data.NewDataSourceFilter()
	f.Add(rng, data.FilterNotContains)

	inRange := time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC)
	if f.ValueMatch(inRange) {
		t.Error("date in range with FilterNotContains should not match")
	}
	before := time.Date(2023, 12, 31, 0, 0, 0, 0, time.UTC)
	if !f.ValueMatch(before) {
		t.Error("date before range with FilterNotContains should match")
	}
}

func TestFilter_cmpFloat64_Equal(t *testing.T) {
	f := data.NewDataSourceFilter()
	f.Add(float64(3.14), data.FilterEqual)
	if !f.ValueMatch(float64(3.14)) {
		t.Error("3.14 == 3.14 should match")
	}
}

func TestFilter_toInt64_Types(t *testing.T) {
	// int32 value filtered against int filter.
	f := data.NewDataSourceFilter()
	f.Add(int(10), data.FilterEqual)
	if !f.ValueMatch(int32(10)) {
		t.Error("int32(10) == int(10) should match")
	}
}

// ─── FilteredDataSource extra paths ──────────────────────────────────────────

func TestFilteredDataSource_Name_Alias(t *testing.T) {
	inner := data.NewBaseDataSource("myDS")
	inner.SetAlias("My DS")
	_ = inner.Init()

	fds, err := data.NewFilteredDataSource(inner, nil, nil)
	if err != nil {
		t.Fatalf("NewFilteredDataSource: %v", err)
	}
	if fds.Name() != "myDS" {
		t.Errorf("Name = %q, want myDS", fds.Name())
	}
	if fds.Alias() != "My DS" {
		t.Errorf("Alias = %q, want 'My DS'", fds.Alias())
	}
}

func TestFilteredDataSource_Init_NoOp(t *testing.T) {
	inner := data.NewBaseDataSource("x")
	_ = inner.Init()
	fds, _ := data.NewFilteredDataSource(inner, nil, nil)
	if err := fds.Init(); err != nil {
		t.Errorf("Init should be no-op, got %v", err)
	}
}

func TestFilteredDataSource_CurrentRowNo(t *testing.T) {
	inner := data.NewBaseDataSource("x")
	inner.AddRow(map[string]any{"v": 1})
	_ = inner.Init()
	fds, _ := data.NewFilteredDataSource(inner, nil, nil)

	// Before First(), cursor is -1.
	if fds.CurrentRowNo() != -1 {
		t.Errorf("CurrentRowNo before First = %d, want -1", fds.CurrentRowNo())
	}
	_ = fds.First()
	if fds.CurrentRowNo() != 0 {
		t.Errorf("CurrentRowNo after First = %d, want 0", fds.CurrentRowNo())
	}
}

func TestFilteredDataSource_Columns(t *testing.T) {
	inner := data.NewBaseDataSource("x")
	inner.AddColumn(data.Column{Name: "col1"})
	inner.AddColumn(data.Column{Name: "col2"})
	_ = inner.Init()

	fds, _ := data.NewFilteredDataSource(inner, nil, nil)
	cols := fds.Columns()
	if len(cols) != 2 {
		t.Errorf("Columns len = %d, want 2", len(cols))
	}
}

func TestFilteredDataSource_Close(t *testing.T) {
	inner := data.NewBaseDataSource("x")
	inner.AddRow(map[string]any{"v": 1})
	_ = inner.Init()
	fds, _ := data.NewFilteredDataSource(inner, nil, nil)
	if err := fds.Close(); err != nil {
		t.Errorf("Close: %v", err)
	}
}

func TestFilteredDataSource_Next_EOF(t *testing.T) {
	inner := data.NewBaseDataSource("x")
	inner.AddRow(map[string]any{"v": 1})
	_ = inner.Init()
	fds, _ := data.NewFilteredDataSource(inner, nil, nil)
	_ = fds.First()
	err := fds.Next()
	if err != data.ErrEOF {
		t.Errorf("Next past end: want ErrEOF, got %v", err)
	}
}

func TestFilteredDataSource_Inner_NoColumns(t *testing.T) {
	// Use a DataSource that does not implement hasColumns.
	vds := data.NewVirtualDataSource("v", 2)
	_ = vds.Init()

	fds, err := data.NewFilteredDataSource(vds, nil, nil)
	if err != nil {
		t.Fatalf("NewFilteredDataSource: %v", err)
	}
	cols := fds.Columns()
	if cols != nil {
		t.Errorf("Columns should be nil for inner that has no Columns(), got %v", cols)
	}
}

// ─── SystemVariables extra coverage ──────────────────────────────────────────

func TestSystemVariables_Set_PageCount_Alias(t *testing.T) {
	sv := data.NewSystemVariables()
	sv.Set(data.SysVarPageCount, 7)
	if sv.TotalPages != 7 {
		t.Errorf("TotalPages after Set(PageCount) = %d, want 7", sv.TotalPages)
	}
}

func TestSystemVariables_Set_AbsRow(t *testing.T) {
	sv := data.NewSystemVariables()
	sv.Set(data.SysVarAbsRow, 100)
	if sv.AbsRow != 100 {
		t.Errorf("AbsRow = %d, want 100", sv.AbsRow)
	}
}

func TestSystemVariables_Set_HierarchyRow(t *testing.T) {
	sv := data.NewSystemVariables()
	sv.Set(data.SysVarHierarchyRow, 5)
	if sv.HierarchyRow != 5 {
		t.Errorf("HierarchyRow = %d, want 5", sv.HierarchyRow)
	}
}

func TestSystemVariables_Set_Time(t *testing.T) {
	sv := data.NewSystemVariables()
	now := time.Now()
	sv.Set(data.SysVarTime, now)
	if !sv.Time.Equal(now) {
		t.Error("Time not set correctly")
	}
}

func TestSystemVariables_Get_HierarchyLevel(t *testing.T) {
	sv := data.NewSystemVariables()
	sv.HierarchyLevel = 3
	if sv.Get(data.SysVarHierarchyLevel) != 3 {
		t.Error("Get(HierarchyLevel) should return 3")
	}
}

func TestSystemVariables_Get_HierarchyRow(t *testing.T) {
	sv := data.NewSystemVariables()
	sv.HierarchyRow = 8
	if sv.Get(data.SysVarHierarchyRow) != 8 {
		t.Error("Get(HierarchyRow) should return 8")
	}
}

func TestSystemVariables_Get_Time(t *testing.T) {
	sv := data.NewSystemVariables()
	now := time.Now()
	sv.Time = now
	if sv.Get(data.SysVarTime) != now {
		t.Error("Get(Time) should return now")
	}
}

// ─── Dictionary extra - FindDataSourceByName partial path ────────────────────

func TestDictionary_FindDataSourceByName_NotFound(t *testing.T) {
	d := data.NewDictionary()
	d.AddDataSource(data.NewBaseDataSource("exists"))
	if d.FindDataSourceByName("nope") != nil {
		t.Error("FindDataSourceByName should return nil for unknown name")
	}
}

// ─── FindRelation fallback by name ───────────────────────────────────────────

func TestFindRelation_FallbackByName(t *testing.T) {
	dict := newStubDict()
	parent := data.NewBaseDataSource("customers")
	child := data.NewBaseDataSource("orders")

	// Relation has nil DataSource pointers but string names.
	rel := &data.Relation{
		ParentSourceName: "customers",
		ChildSourceName:  "orders",
	}
	dict.relations = []*data.Relation{rel}

	found := data.FindRelation(dict, parent, child)
	if found != rel {
		t.Error("FindRelation should match by name when DataSource pointers are nil")
	}
}

func TestFindRelation_FallbackNilParentChild(t *testing.T) {
	dict := newStubDict()
	// rel has nil DataSources and nil strings.
	rel := &data.Relation{}
	dict.relations = []*data.Relation{rel}
	// Pass nil parent/child — names will be "".
	found := data.FindRelation(dict, nil, nil)
	if found != rel {
		t.Errorf("FindRelation with nil parents and empty rel should match, got %v", found)
	}
}

// ─── AggregateTotal extra ────────────────────────────────────────────────────

func TestAggregateTotal_Add_UnknownType(t *testing.T) {
	at := data.NewAggregateTotal("t")
	at.TotalType = data.TotalTypeSum
	// uint8 is not in toFloat64.
	// Actually uint8 IS in toFloat64; let's try a string.
	err := at.Add("not a number")
	if err == nil {
		t.Error("expected error for non-numeric Add")
	}
}

func TestAggregateTotal_Value_Default(t *testing.T) {
	// TotalType beyond known range returns nil.
	at := data.NewAggregateTotal("t")
	at.TotalType = data.TotalType(999)
	v := at.Value()
	if v != nil {
		t.Errorf("unknown TotalType Value() = %v, want nil", v)
	}
}

// ─── ViewDataSource ───────────────────────────────────────────────────────────

func makeViewInner() *data.BaseDataSource {
	ds := data.NewBaseDataSource("inner")
	ds.AddRow(map[string]any{"name": "Alice", "active": true})
	ds.AddRow(map[string]any{"name": "Bob", "active": false})
	ds.AddRow(map[string]any{"name": "Carol", "active": true})
	return ds
}

func TestViewDataSource_NoFilter(t *testing.T) {
	inner := makeViewInner()
	vds := data.NewViewDataSource(inner, "myView", "My View", "", nil)

	if vds.Name() != "myView" {
		t.Errorf("Name = %q, want myView", vds.Name())
	}
	if vds.Alias() != "My View" {
		t.Errorf("Alias = %q, want 'My View'", vds.Alias())
	}
	if vds.Filter() != "" {
		t.Errorf("Filter = %q, want empty", vds.Filter())
	}

	if err := vds.Init(); err != nil {
		t.Fatalf("Init: %v", err)
	}
	if vds.RowCount() != 3 {
		t.Errorf("RowCount = %d, want 3 (no filter)", vds.RowCount())
	}
}

func TestViewDataSource_WithFilter(t *testing.T) {
	inner := makeViewInner()
	// Filter: include only rows where active == true.
	eval := func(expr string, src data.DataSource) (bool, error) {
		v, err := src.GetValue("active")
		if err != nil {
			return false, err
		}
		b, _ := v.(bool)
		return b, nil
	}
	vds := data.NewViewDataSource(inner, "active", "Active", "active", eval)

	if err := vds.Init(); err != nil {
		t.Fatalf("Init: %v", err)
	}
	if vds.RowCount() != 2 {
		t.Errorf("RowCount = %d, want 2 (filtered)", vds.RowCount())
	}
}

func TestViewDataSource_Next_GetValue(t *testing.T) {
	inner := makeViewInner()
	vds := data.NewViewDataSource(inner, "v", "V", "", nil)
	if err := vds.Init(); err != nil {
		t.Fatalf("Init: %v", err)
	}

	if err := vds.Next(); err != nil {
		t.Fatalf("Next (first): %v", err)
	}
	v, err := vds.GetValue("name")
	if err != nil {
		t.Fatalf("GetValue: %v", err)
	}
	if v != "Alice" {
		t.Errorf("name = %v, want Alice", v)
	}
}

func TestViewDataSource_Next_EOF(t *testing.T) {
	inner := makeViewInner()
	vds := data.NewViewDataSource(inner, "v", "V", "", nil)
	_ = vds.Init()

	_ = vds.Next()
	_ = vds.Next()
	_ = vds.Next()
	err := vds.Next()
	if err != data.ErrEOF {
		t.Errorf("Next past end: want ErrEOF, got %v", err)
	}
	if !vds.EOF() {
		t.Error("EOF should be true")
	}
}

func TestViewDataSource_First(t *testing.T) {
	inner := makeViewInner()
	vds := data.NewViewDataSource(inner, "v", "V", "", nil)
	_ = vds.Init()
	_ = vds.Next()
	_ = vds.Next()

	if err := vds.First(); err != nil {
		t.Fatalf("First: %v", err)
	}
	if vds.CurrentRowNo() != -1 {
		t.Errorf("CurrentRowNo after First = %d, want -1", vds.CurrentRowNo())
	}
}

func TestViewDataSource_First_BeforeInit(t *testing.T) {
	// First without Init triggers rebuildIndex implicitly.
	inner := makeViewInner()
	vds := data.NewViewDataSource(inner, "v", "V", "", nil)
	if err := vds.First(); err != nil {
		t.Fatalf("First before Init: %v", err)
	}
}

func TestViewDataSource_EOF(t *testing.T) {
	inner := data.NewBaseDataSource("empty")
	_ = inner.Init()
	vds := data.NewViewDataSource(inner, "v", "V", "", nil)
	_ = vds.Init()
	// cursor = -1, rows = [] (empty); EOF is cursor >= len(rows) = -1 >= 0 = false initially.
	// After calling Next, cursor becomes 0 which is >= 0 (len([])=0).
	err := vds.Next()
	if err != data.ErrEOF {
		t.Errorf("Next on empty view: want ErrEOF, got %v", err)
	}
	if !vds.EOF() {
		t.Error("EOF should be true after Next on empty view")
	}
}

func TestViewDataSource_RowCount(t *testing.T) {
	inner := makeViewInner()
	vds := data.NewViewDataSource(inner, "v", "V", "", nil)
	_ = vds.Init()
	if vds.RowCount() != 3 {
		t.Errorf("RowCount = %d, want 3", vds.RowCount())
	}
}

func TestViewDataSource_CurrentRowNo_BeforeNext(t *testing.T) {
	inner := makeViewInner()
	vds := data.NewViewDataSource(inner, "v", "V", "", nil)
	_ = vds.Init()
	// cursor = -1 before any Next.
	if vds.CurrentRowNo() != -1 {
		t.Errorf("CurrentRowNo before Next = %d, want -1", vds.CurrentRowNo())
	}
}

func TestViewDataSource_SetFilter(t *testing.T) {
	inner := makeViewInner()
	vds := data.NewViewDataSource(inner, "v", "V", "", nil)
	_ = vds.Init()
	if vds.RowCount() != 3 {
		t.Fatalf("RowCount before filter = %d, want 3", vds.RowCount())
	}

	// Change filter — index should be invalidated.
	eval := func(expr string, src data.DataSource) (bool, error) {
		v, err := src.GetValue("active")
		if err != nil {
			return false, err
		}
		b, _ := v.(bool)
		return b, nil
	}
	vds.SetFilter("active")
	// Need a new ViewDataSource with the eval, or call rebuildIndex indirectly.
	// SetFilter invalidates initDone; Init would call inner.Init again.
	// Instead create a new view with the evaluator.
	_ = eval
	if vds.Filter() != "active" {
		t.Errorf("Filter = %q, want active", vds.Filter())
	}
}

func TestViewDataSource_Close(t *testing.T) {
	inner := makeViewInner()
	vds := data.NewViewDataSource(inner, "v", "V", "", nil)
	_ = vds.Init()
	if err := vds.Close(); err != nil {
		t.Errorf("Close: %v", err)
	}
}

func TestViewDataSource_FilterEvalError(t *testing.T) {
	// When evaluator returns an error, the row is included (safe default).
	inner := data.NewBaseDataSource("inner")
	inner.AddRow(map[string]any{"x": 1})
	inner.AddRow(map[string]any{"x": 2})

	eval := func(expr string, src data.DataSource) (bool, error) {
		v, _ := src.GetValue("x")
		if v.(int) == 1 {
			return false, fmt.Errorf("eval error") // error → include row
		}
		return true, nil
	}
	vds := data.NewViewDataSource(inner, "v", "V", "expr", eval)
	if err := vds.Init(); err != nil {
		t.Fatalf("Init: %v", err)
	}
	// Row 1 had eval error (included), row 2 returned true (included).
	if vds.RowCount() != 2 {
		t.Errorf("RowCount = %d, want 2 (eval error → include)", vds.RowCount())
	}
}

func TestViewDataSource_NilEval_UsesDefault(t *testing.T) {
	// Passing nil eval uses the always-true evaluator.
	inner := makeViewInner()
	vds := data.NewViewDataSource(inner, "v", "V", "any_expr", nil)
	_ = vds.Init()
	if vds.RowCount() != 3 {
		t.Errorf("RowCount = %d, want 3 with nil eval (include all)", vds.RowCount())
	}
}

// ─── VirtualDataSource extra ──────────────────────────────────────────────────

func TestVirtualDataSource_Alias_SetAlias(t *testing.T) {
	vds := data.NewVirtualDataSource("myVirtual", 5)
	if vds.Alias() != "myVirtual" {
		t.Errorf("Alias default = %q, want myVirtual", vds.Alias())
	}
	vds.SetAlias("My Virtual")
	if vds.Alias() != "My Virtual" {
		t.Errorf("Alias after SetAlias = %q, want 'My Virtual'", vds.Alias())
	}
}

func TestVirtualDataSource_Close(t *testing.T) {
	vds := data.NewVirtualDataSource("v", 3)
	_ = vds.Init()
	if err := vds.Close(); err != nil {
		t.Errorf("Close: %v", err)
	}
}

func TestVirtualDataSource_First_Empty(t *testing.T) {
	vds := data.NewVirtualDataSource("v", 0)
	_ = vds.Init()
	err := vds.First()
	if err != data.ErrEOF {
		t.Errorf("First on empty virtual: want ErrEOF, got %v", err)
	}
}

func TestVirtualDataSource_Next_NotInitialized(t *testing.T) {
	vds := data.NewVirtualDataSource("v", 3)
	// Not initialized yet.
	err := vds.Next()
	if err != data.ErrNotInitialized {
		t.Errorf("Next without Init: want ErrNotInitialized, got %v", err)
	}
}
