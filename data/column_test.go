package data_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/andrewloable/go-fastreport/data"
	"github.com/andrewloable/go-fastreport/serial"
)

func TestNewDataColumn(t *testing.T) {
	col := data.NewDataColumn("OrderID")
	if col.Name != "OrderID" {
		t.Errorf("Name = %q, want OrderID", col.Name)
	}
	if col.Alias != "OrderID" {
		t.Errorf("Alias default should equal Name, got %q", col.Alias)
	}
	if col.PropName != "OrderID" {
		t.Errorf("PropName default should equal Name, got %q", col.PropName)
	}
	if !col.Enabled {
		t.Error("Enabled should default to true")
	}
	if col.Calculated {
		t.Error("Calculated should default to false")
	}
	if col.Tag != nil {
		t.Error("Tag should default to nil")
	}
}

func TestDataColumnHasColumns(t *testing.T) {
	col := data.NewDataColumn("Customer")
	if col.HasColumns() {
		t.Error("new column should have no nested columns")
	}
	col.Columns().Add(data.NewDataColumn("Name"))
	if !col.HasColumns() {
		t.Error("column should have nested columns after add")
	}
}

func TestColumnCollectionAdd(t *testing.T) {
	cc := data.NewColumnCollection()
	cc.Add(data.NewDataColumn("A"))
	cc.Add(data.NewDataColumn("B"))
	if cc.Len() != 2 {
		t.Errorf("Len = %d, want 2", cc.Len())
	}
}

func TestColumnCollectionGet(t *testing.T) {
	cc := data.NewColumnCollection()
	cc.Add(data.NewDataColumn("X"))
	if cc.Get(0).Name != "X" {
		t.Errorf("Get(0).Name = %q, want X", cc.Get(0).Name)
	}
}

func TestColumnCollectionRemove(t *testing.T) {
	cc := data.NewColumnCollection()
	cc.Add(data.NewDataColumn("A"))
	cc.Add(data.NewDataColumn("B"))
	ok := cc.Remove("A")
	if !ok {
		t.Error("Remove should return true for existing column")
	}
	if cc.Len() != 1 {
		t.Errorf("Len = %d, want 1", cc.Len())
	}
	if cc.Get(0).Name != "B" {
		t.Errorf("remaining column = %q, want B", cc.Get(0).Name)
	}
	ok = cc.Remove("nonexistent")
	if ok {
		t.Error("Remove of nonexistent should return false")
	}
}

func TestColumnCollectionClear(t *testing.T) {
	cc := data.NewColumnCollection()
	cc.Add(data.NewDataColumn("A"))
	cc.Clear()
	if cc.Len() != 0 {
		t.Errorf("after Clear Len = %d, want 0", cc.Len())
	}
}

func TestColumnCollectionFindByName(t *testing.T) {
	cc := data.NewColumnCollection()
	cc.Add(data.NewDataColumn("OrderID"))
	cc.Add(data.NewDataColumn("CustomerName"))

	col := cc.FindByName("CustomerName")
	if col == nil {
		t.Fatal("FindByName returned nil for existing column")
	}
	if col.Name != "CustomerName" {
		t.Errorf("found %q, want CustomerName", col.Name)
	}
	if cc.FindByName("missing") != nil {
		t.Error("FindByName should return nil for missing column")
	}
}

func TestColumnCollectionFindByNameNested(t *testing.T) {
	cc := data.NewColumnCollection()
	parent := data.NewDataColumn("Address")
	parent.Columns().Add(data.NewDataColumn("City"))
	parent.Columns().Add(data.NewDataColumn("ZipCode"))
	cc.Add(parent)

	col := cc.FindByName("City")
	if col == nil {
		t.Fatal("FindByName should find nested column")
	}
	if col.Name != "City" {
		t.Errorf("found %q, want City", col.Name)
	}
}

func TestColumnCollectionFindByAlias(t *testing.T) {
	cc := data.NewColumnCollection()
	col := data.NewDataColumn("cust_name")
	col.Alias = "Customer Name"
	cc.Add(col)

	found := cc.FindByAlias("Customer Name")
	if found == nil {
		t.Fatal("FindByAlias returned nil")
	}
	if found.Name != "cust_name" {
		t.Errorf("found %q, want cust_name", found.Name)
	}
	if cc.FindByAlias("nonexistent") != nil {
		t.Error("FindByAlias should return nil for missing alias")
	}
}

func TestColumnCollectionFindByAliasNested(t *testing.T) {
	cc := data.NewColumnCollection()
	parent := data.NewDataColumn("Location")
	child := data.NewDataColumn("zip")
	child.Alias = "Zip Code"
	parent.Columns().Add(child)
	cc.Add(parent)

	found := cc.FindByAlias("Zip Code")
	if found == nil {
		t.Fatal("FindByAlias should find nested column by alias")
	}
	if found.Name != "zip" {
		t.Errorf("found %q, want zip", found.Name)
	}
}

func TestColumnCollectionAll(t *testing.T) {
	cc := data.NewColumnCollection()
	cc.Add(data.NewDataColumn("A"))
	cc.Add(data.NewDataColumn("B"))
	cc.Add(data.NewDataColumn("C"))

	sum := 0
	for i := range cc.All() {
		sum += i
	}
	if sum != 3 { // 0+1+2
		t.Errorf("All indices sum = %d, want 3", sum)
	}
}

func TestColumnCollectionAllEarlyStop(t *testing.T) {
	cc := data.NewColumnCollection()
	for i := 0; i < 5; i++ {
		cc.Add(data.NewDataColumn("col"))
	}
	count := 0
	for range cc.All() {
		count++
		if count == 2 {
			break
		}
	}
	if count != 2 {
		t.Errorf("early stop count = %d, want 2", count)
	}
}

func TestColumnCollectionSlice(t *testing.T) {
	cc := data.NewColumnCollection()
	cc.Add(data.NewDataColumn("A"))
	cc.Add(data.NewDataColumn("B"))
	s := cc.Slice()
	if len(s) != 2 {
		t.Fatalf("Slice len = %d, want 2", len(s))
	}
	// Modifying the slice should not affect the collection
	s[0] = data.NewDataColumn("Z")
	if cc.Get(0).Name != "A" {
		t.Error("Slice is not a copy")
	}
}

func TestColumnFormatConstants(t *testing.T) {
	formats := []data.ColumnFormat{
		data.ColumnFormatAuto,
		data.ColumnFormatGeneral,
		data.ColumnFormatNumber,
		data.ColumnFormatCurrency,
		data.ColumnFormatDate,
		data.ColumnFormatTime,
		data.ColumnFormatPercent,
		data.ColumnFormatBoolean,
	}
	seen := make(map[data.ColumnFormat]bool)
	for _, f := range formats {
		if seen[f] {
			t.Errorf("duplicate ColumnFormat value %d", f)
		}
		seen[f] = true
	}
}

// ---- SetName ---------------------------------------------------------------

func TestDataColumnSetName_SyncsPropName(t *testing.T) {
	col := data.NewDataColumn("OldName")
	// PropName should initially match Name.
	if col.PropName != "OldName" {
		t.Fatalf("PropName = %q, want OldName", col.PropName)
	}
	col.SetName("NewName")
	if col.Name != "NewName" {
		t.Errorf("Name = %q, want NewName", col.Name)
	}
	if col.PropName != "NewName" {
		t.Errorf("PropName should sync: got %q, want NewName", col.PropName)
	}
	if col.Alias != "NewName" {
		t.Errorf("Alias should sync: got %q, want NewName", col.Alias)
	}
}

func TestDataColumnSetName_KeepsCustomPropName(t *testing.T) {
	col := data.NewDataColumn("OrigName")
	col.PropName = "CustomProp"
	col.SetName("ChangedName")
	if col.PropName != "CustomProp" {
		t.Errorf("PropName should remain %q when custom, got %q", "CustomProp", col.PropName)
	}
}

func TestDataColumnSetName_KeepsCustomAlias(t *testing.T) {
	col := data.NewDataColumn("OrigName")
	col.Alias = "Custom Alias"
	col.SetName("ChangedName")
	if col.Alias != "Custom Alias" {
		t.Errorf("Alias should remain %q when custom, got %q", "Custom Alias", col.Alias)
	}
}

// ---- FullName --------------------------------------------------------------

func TestDataColumnFullName_TopLevel(t *testing.T) {
	col := data.NewDataColumn("OrderID")
	if fn := col.FullName(); fn != "OrderID" {
		t.Errorf("FullName = %q, want OrderID", fn)
	}
}

func TestDataColumnFullName_Nested(t *testing.T) {
	parent := data.NewDataColumn("Address")
	child := data.NewDataColumn("City")
	parent.Columns().Add(child)

	if fn := child.FullName(); fn != "Address.City" {
		t.Errorf("FullName = %q, want Address.City", fn)
	}
}

func TestDataColumnFullName_DeepNested(t *testing.T) {
	root := data.NewDataColumn("Root")
	mid := data.NewDataColumn("Mid")
	leaf := data.NewDataColumn("Leaf")
	root.Columns().Add(mid)
	mid.Columns().Add(leaf)

	if fn := leaf.FullName(); fn != "Root.Mid.Leaf" {
		t.Errorf("FullName = %q, want Root.Mid.Leaf", fn)
	}
}

// ---- Parent ----------------------------------------------------------------

func TestDataColumnParent(t *testing.T) {
	parent := data.NewDataColumn("Parent")
	child := data.NewDataColumn("Child")
	if child.Parent() != nil {
		t.Error("new column should have nil parent")
	}
	parent.Columns().Add(child)
	if child.Parent() != parent {
		t.Error("after Add, child.Parent() should equal parent")
	}
}

func TestDataColumnSetParent(t *testing.T) {
	col := data.NewDataColumn("X")
	p := data.NewDataColumn("P")
	col.SetParent(p)
	if col.Parent() != p {
		t.Error("SetParent did not set the parent")
	}
	col.SetParent(nil)
	if col.Parent() != nil {
		t.Error("SetParent(nil) did not clear the parent")
	}
}

// ---- Tag -------------------------------------------------------------------

func TestDataColumnTag(t *testing.T) {
	col := data.NewDataColumn("X")
	col.Tag = "hello"
	if col.Tag != "hello" {
		t.Errorf("Tag = %v, want hello", col.Tag)
	}
	col.Tag = 42
	if col.Tag != 42 {
		t.Errorf("Tag = %v, want 42", col.Tag)
	}
}

// ---- GetExpressions --------------------------------------------------------

func TestDataColumnGetExpressions_NotCalculated(t *testing.T) {
	col := data.NewDataColumn("OrderID")
	exprs := col.GetExpressions()
	if exprs != nil {
		t.Errorf("GetExpressions should return nil for non-calculated column, got %v", exprs)
	}
}

func TestDataColumnGetExpressions_Calculated(t *testing.T) {
	col := data.NewDataColumn("Total")
	col.Calculated = true
	col.Expression = "[Price] * [Qty]"
	exprs := col.GetExpressions()
	if len(exprs) != 1 || exprs[0] != "[Price] * [Qty]" {
		t.Errorf("GetExpressions = %v, want [\"[Price] * [Qty]\"]", exprs)
	}
}

func TestDataColumnGetExpressions_CalculatedEmptyExpression(t *testing.T) {
	col := data.NewDataColumn("Empty")
	col.Calculated = true
	col.Expression = ""
	exprs := col.GetExpressions()
	if exprs != nil {
		t.Errorf("GetExpressions should return nil when expression is empty, got %v", exprs)
	}
}

// ---- FindByPropName --------------------------------------------------------

func TestColumnCollectionFindByPropName(t *testing.T) {
	cc := data.NewColumnCollection()
	col1 := data.NewDataColumn("col1")
	col1.PropName = "PropertyA"
	col2 := data.NewDataColumn("col2")
	col2.PropName = "PropertyB"
	cc.Add(col1)
	cc.Add(col2)

	found := cc.FindByPropName("PropertyB")
	if found == nil {
		t.Fatal("FindByPropName returned nil for existing column")
	}
	if found.Name != "col2" {
		t.Errorf("found %q, want col2", found.Name)
	}
	if cc.FindByPropName("missing") != nil {
		t.Error("FindByPropName should return nil for missing PropName")
	}
}

// ---- Serialize / Deserialize round-trip ------------------------------------

// roundTripDataColumn serializes col to XML and deserializes into a new
// DataColumn, returning the deserialized copy.
func roundTripDataColumn(t *testing.T, col *data.DataColumn) *data.DataColumn {
	t.Helper()
	var buf bytes.Buffer
	w := serial.NewWriter(&buf)
	if err := w.BeginObject("Column"); err != nil {
		t.Fatalf("BeginObject: %v", err)
	}
	w.WriteStr("Name", col.Name)
	if err := col.Serialize(w); err != nil {
		t.Fatalf("Serialize: %v", err)
	}
	if err := w.EndObject(); err != nil {
		t.Fatalf("EndObject: %v", err)
	}
	if err := w.Flush(); err != nil {
		t.Fatalf("Flush: %v", err)
	}

	r := serial.NewReader(bytes.NewReader(buf.Bytes()))
	typeName, ok := r.ReadObjectHeader()
	if !ok {
		t.Fatalf("ReadObjectHeader: not ok; xml=%s", buf.String())
	}
	if typeName != "Column" {
		t.Fatalf("typeName=%q, want Column", typeName)
	}
	got := data.NewDataColumn(r.ReadStr("Name", ""))
	if err := got.Deserialize(r); err != nil {
		t.Fatalf("Deserialize: %v", err)
	}
	return got
}

func TestDataColumnSerializeDeserialize_Defaults(t *testing.T) {
	col := data.NewDataColumn("OrderID")
	got := roundTripDataColumn(t, col)
	if got.Name != "OrderID" {
		t.Errorf("Name = %q, want OrderID", got.Name)
	}
	if got.Alias != "OrderID" {
		t.Errorf("Alias = %q, want OrderID", got.Alias)
	}
	if got.PropName != "OrderID" {
		t.Errorf("PropName = %q, want OrderID", got.PropName)
	}
	if got.Format != data.ColumnFormatAuto {
		t.Errorf("Format = %d, want ColumnFormatAuto", got.Format)
	}
	if got.Calculated {
		t.Error("Calculated should be false")
	}
	if !got.Enabled {
		t.Error("Enabled should be true")
	}
}

func TestDataColumnSerializeDeserialize_AllFields(t *testing.T) {
	col := data.NewDataColumn("calc_total")
	col.Alias = "Total Amount"
	col.DataType = "System.Decimal"
	col.PropName = "TotalProp"
	col.Format = data.ColumnFormatCurrency
	col.Calculated = true
	col.Expression = "[Price] * [Qty]"
	col.Enabled = false

	got := roundTripDataColumn(t, col)
	if got.Name != "calc_total" {
		t.Errorf("Name = %q", got.Name)
	}
	if got.Alias != "Total Amount" {
		t.Errorf("Alias = %q, want 'Total Amount'", got.Alias)
	}
	if got.DataType != "System.Decimal" {
		t.Errorf("DataType = %q, want System.Decimal", got.DataType)
	}
	if got.PropName != "TotalProp" {
		t.Errorf("PropName = %q, want TotalProp", got.PropName)
	}
	if got.Format != data.ColumnFormatCurrency {
		t.Errorf("Format = %d, want ColumnFormatCurrency (%d)", got.Format, data.ColumnFormatCurrency)
	}
	if !got.Calculated {
		t.Error("Calculated should be true")
	}
	if got.Expression != "[Price] * [Qty]" {
		t.Errorf("Expression = %q", got.Expression)
	}
	if got.Enabled {
		t.Error("Enabled should be false")
	}
}

func TestDataColumnSerializeDeserialize_CalculatedColumn(t *testing.T) {
	col := data.NewDataColumn("FullName")
	col.Calculated = true
	col.Expression = "[FirstName] + \" \" + [LastName]"

	got := roundTripDataColumn(t, col)
	if !got.Calculated {
		t.Error("Calculated should be true")
	}
	if got.Expression != "[FirstName] + \" \" + [LastName]" {
		t.Errorf("Expression = %q", got.Expression)
	}
}

func TestDataColumnDeserialize_FromXML(t *testing.T) {
	xmlData := `<Column Name="Amount" Alias="Order Amount" DataType="System.Decimal" PropName="amt" Format="Number" Calculated="true" Expression="[Price]*[Qty]" Enabled="false"/>`
	r := serial.NewReader(strings.NewReader(xmlData))
	typeName, ok := r.ReadObjectHeader()
	if !ok {
		t.Fatal("ReadObjectHeader: not ok")
	}
	if typeName != "Column" {
		t.Fatalf("typeName = %q, want Column", typeName)
	}
	col := data.NewDataColumn(r.ReadStr("Name", ""))
	if err := col.Deserialize(r); err != nil {
		t.Fatalf("Deserialize: %v", err)
	}
	if col.Name != "Amount" {
		t.Errorf("Name = %q", col.Name)
	}
	if col.Alias != "Order Amount" {
		t.Errorf("Alias = %q", col.Alias)
	}
	if col.DataType != "System.Decimal" {
		t.Errorf("DataType = %q", col.DataType)
	}
	if col.PropName != "amt" {
		t.Errorf("PropName = %q", col.PropName)
	}
	if col.Format != data.ColumnFormatNumber {
		t.Errorf("Format = %d, want ColumnFormatNumber", col.Format)
	}
	if !col.Calculated {
		t.Error("Calculated should be true")
	}
	if col.Expression != "[Price]*[Qty]" {
		t.Errorf("Expression = %q", col.Expression)
	}
	if col.Enabled {
		t.Error("Enabled should be false")
	}
}

func TestDataColumnSerialize_PropNameSkippedWhenSameAsName(t *testing.T) {
	col := data.NewDataColumn("OrderID")
	// PropName defaults to same as Name, so it should not be serialized.
	var buf bytes.Buffer
	w := serial.NewWriter(&buf)
	if err := w.BeginObject("Column"); err != nil {
		t.Fatal(err)
	}
	_ = col.Serialize(w)
	_ = w.EndObject()
	_ = w.Flush()

	xml := buf.String()
	if strings.Contains(xml, "PropName") {
		t.Errorf("PropName should not be serialized when it matches Name; xml=%s", xml)
	}
}

func TestDataColumnSerialize_AliasSkippedWhenSameAsName(t *testing.T) {
	col := data.NewDataColumn("OrderID")
	var buf bytes.Buffer
	w := serial.NewWriter(&buf)
	if err := w.BeginObject("Column"); err != nil {
		t.Fatal(err)
	}
	_ = col.Serialize(w)
	_ = w.EndObject()
	_ = w.Flush()

	xml := buf.String()
	if strings.Contains(xml, "Alias") {
		t.Errorf("Alias should not be serialized when it matches Name; xml=%s", xml)
	}
}

func TestDataColumnSerialize_FormatSkippedWhenAuto(t *testing.T) {
	col := data.NewDataColumn("OrderID")
	var buf bytes.Buffer
	w := serial.NewWriter(&buf)
	if err := w.BeginObject("Column"); err != nil {
		t.Fatal(err)
	}
	_ = col.Serialize(w)
	_ = w.EndObject()
	_ = w.Flush()

	xml := buf.String()
	if strings.Contains(xml, "Format") {
		t.Errorf("Format should not be serialized when Auto; xml=%s", xml)
	}
}

func TestColumnFormatString(t *testing.T) {
	tests := []struct {
		f    data.ColumnFormat
		want string
	}{
		{data.ColumnFormatAuto, "Auto"},
		{data.ColumnFormatGeneral, "General"},
		{data.ColumnFormatNumber, "Number"},
		{data.ColumnFormatCurrency, "Currency"},
		{data.ColumnFormatDate, "Date"},
		{data.ColumnFormatTime, "Time"},
		{data.ColumnFormatPercent, "Percent"},
		{data.ColumnFormatBoolean, "Boolean"},
		{data.ColumnFormat(99), "99"}, // out of range
	}
	for _, tc := range tests {
		got := data.ColumnFormatString(tc.f)
		if got != tc.want {
			t.Errorf("ColumnFormatString(%d) = %q, want %q", tc.f, got, tc.want)
		}
	}
}

func TestDataColumnDeserialize_AllFormats(t *testing.T) {
	formats := []struct {
		name string
		want data.ColumnFormat
	}{
		{"Auto", data.ColumnFormatAuto},
		{"General", data.ColumnFormatGeneral},
		{"Number", data.ColumnFormatNumber},
		{"Currency", data.ColumnFormatCurrency},
		{"Date", data.ColumnFormatDate},
		{"Time", data.ColumnFormatTime},
		{"Percent", data.ColumnFormatPercent},
		{"Boolean", data.ColumnFormatBoolean},
	}
	for _, tc := range formats {
		xmlData := `<Column Name="X" Format="` + tc.name + `"/>`
		r := serial.NewReader(strings.NewReader(xmlData))
		_, _ = r.ReadObjectHeader()
		col := data.NewDataColumn(r.ReadStr("Name", ""))
		_ = col.Deserialize(r)
		if col.Format != tc.want {
			t.Errorf("Format %q: got %d, want %d", tc.name, col.Format, tc.want)
		}
	}
}

func TestDataColumnDeserialize_UnknownFormatDefaultsToAuto(t *testing.T) {
	xmlData := `<Column Name="X" Format="SomeUnknownFormat"/>`
	r := serial.NewReader(strings.NewReader(xmlData))
	_, _ = r.ReadObjectHeader()
	col := data.NewDataColumn(r.ReadStr("Name", ""))
	_ = col.Deserialize(r)
	if col.Format != data.ColumnFormatAuto {
		t.Errorf("unknown format should default to Auto, got %d", col.Format)
	}
}

func TestDataColumnDeserialize_ExpressionWithoutCalculated(t *testing.T) {
	// Expression attribute present but Calculated=false -- should still read it.
	xmlData := `<Column Name="X" Expression="[Foo]+[Bar]"/>`
	r := serial.NewReader(strings.NewReader(xmlData))
	_, _ = r.ReadObjectHeader()
	col := data.NewDataColumn(r.ReadStr("Name", ""))
	_ = col.Deserialize(r)
	if col.Expression != "[Foo]+[Bar]" {
		t.Errorf("Expression = %q, want [Foo]+[Bar]", col.Expression)
	}
	if col.Calculated {
		t.Error("Calculated should be false")
	}
}
