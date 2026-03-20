package reportpkg_test

// Tests for injectRelatedFields and injectRelatedFieldsFrom in calc.go.
// These functions inject parent (and grandparent) data source column values
// into the calc environment so that expressions like
// [Child.Parent.ColumnName] and [GrandChild.Child.GrandParent.ColumnName]
// can be evaluated.

import (
	"fmt"
	"testing"

	"github.com/andrewloable/go-fastreport/data"
	"github.com/andrewloable/go-fastreport/reportpkg"
)

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

// makeDS creates and initialises a BaseDataSource with the given name/alias,
// column names, and a slice of row maps.
func makeDS(name string, cols []string, rows []map[string]any) *data.BaseDataSource {
	ds := data.NewBaseDataSource(name)
	ds.SetAlias(name)
	for _, c := range cols {
		ds.AddColumn(data.Column{Name: c})
	}
	for _, r := range rows {
		ds.AddRow(r)
	}
	_ = ds.Init()
	return ds
}

// ---------------------------------------------------------------------------
// injectRelatedFields – single-hop via direct pointer references
// ---------------------------------------------------------------------------

// TestInjectRelatedFields_BasicPointerRelation is a direct pointer variant that
// verifies [Child.Parent.Column] resolves when the relation uses DS pointers.
// The existing TestReport_Calc_RelatedDataSourceField covers this, but we add
// an explicit Orders→OrderDetails scenario to exercise more column paths.
func TestInjectRelatedFields_OrdersDetails_PointerRelation(t *testing.T) {
	r := reportpkg.NewReport()

	ordersDS := makeDS("Orders", []string{"OrderID", "CustomerID", "ShipName"}, []map[string]any{
		{"OrderID": "10248", "CustomerID": "VINET", "ShipName": "Vins et alcools Chevalier"},
		{"OrderID": "10249", "CustomerID": "TOMSP", "ShipName": "Toms Spezialitaeten"},
	})

	orderDetailsDS := makeDS("Order Details", []string{"OrderID", "ProductID", "UnitPrice", "Quantity"}, []map[string]any{
		{"OrderID": "10248", "ProductID": "11", "UnitPrice": "14.0", "Quantity": "12"},
		{"OrderID": "10248", "ProductID": "42", "UnitPrice": "9.8", "Quantity": "10"},
		{"OrderID": "10249", "ProductID": "14", "UnitPrice": "18.6", "Quantity": "9"},
	})

	dict := r.Dictionary()
	dict.AddDataSource(ordersDS)
	dict.AddDataSource(orderDetailsDS)
	dict.AddRelation(&data.Relation{
		Name:             "OrdersOrderDetails",
		ParentDataSource: ordersDS,
		ChildDataSource:  orderDetailsDS,
		ParentColumns:    []string{"OrderID"},
		ChildColumns:     []string{"OrderID"},
	})

	// First row of Order Details (OrderID=10248) should find Orders row 10248.
	_ = orderDetailsDS.First()
	r.SetCalcContext(orderDetailsDS)

	val, err := r.Calc("[Order Details.Orders.CustomerID]")
	if err != nil {
		t.Fatalf("row 0: Calc [Order Details.Orders.CustomerID]: %v", err)
	}
	if val != "VINET" {
		t.Errorf("row 0: got %v, want VINET", val)
	}

	val2, err := r.Calc("[Order Details.Orders.ShipName]")
	if err != nil {
		t.Fatalf("row 0: Calc [Order Details.Orders.ShipName]: %v", err)
	}
	if val2 != "Vins et alcools Chevalier" {
		t.Errorf("row 0: got %v, want Vins et alcools Chevalier", val2)
	}

	// Third row of Order Details (OrderID=10249) should find Orders row 10249.
	_ = orderDetailsDS.Next()
	_ = orderDetailsDS.Next()
	r.SetCalcContext(orderDetailsDS)

	val3, err := r.Calc("[Order Details.Orders.CustomerID]")
	if err != nil {
		t.Fatalf("row 2: Calc [Order Details.Orders.CustomerID]: %v", err)
	}
	if val3 != "TOMSP" {
		t.Errorf("row 2: got %v, want TOMSP", val3)
	}
}

// ---------------------------------------------------------------------------
// injectRelatedFields – single-hop via FRX-style name references
// ---------------------------------------------------------------------------

// TestInjectRelatedFields_NameBasedRelation exercises the
// rel.ChildSourceName/rel.ParentSourceName code path in injectRelatedFields
// (the "else if" branch used when DS pointers are not set, as is the case for
// FRX-loaded reports).
func TestInjectRelatedFields_NameBasedRelation(t *testing.T) {
	r := reportpkg.NewReport()

	customersDS := makeDS("Customers", []string{"CustomerID", "CompanyName", "Country"}, []map[string]any{
		{"CustomerID": "ALFKI", "CompanyName": "Alfreds Futterkiste", "Country": "Germany"},
		{"CustomerID": "BONAP", "CompanyName": "Bon app", "Country": "France"},
	})

	ordersDS := makeDS("Orders", []string{"OrderID", "CustomerID", "Freight"}, []map[string]any{
		{"OrderID": "10643", "CustomerID": "ALFKI", "Freight": "29.46"},
		{"OrderID": "10835", "CustomerID": "ALFKI", "Freight": "69.53"},
		{"OrderID": "10730", "CustomerID": "BONAP", "Freight": "20.12"},
	})

	dict := r.Dictionary()
	dict.AddDataSource(customersDS)
	dict.AddDataSource(ordersDS)

	// Name-based relation (no DS pointer set).
	dict.AddRelation(&data.Relation{
		Name:            "CustomersOrders",
		ChildSourceName: "Orders",
		ParentSourceName: "Customers",
		ChildColumnNames: []string{"CustomerID"},
		ParentColumnNames: []string{"CustomerID"},
	})

	// First Orders row → CustomerID=ALFKI → Alfreds Futterkiste
	_ = ordersDS.First()
	r.SetCalcContext(ordersDS)

	val, err := r.Calc("[Orders.Customers.CompanyName]")
	if err != nil {
		t.Fatalf("row 0: Calc [Orders.Customers.CompanyName]: %v", err)
	}
	if val != "Alfreds Futterkiste" {
		t.Errorf("row 0: got %v, want Alfreds Futterkiste", val)
	}

	// Third Orders row → CustomerID=BONAP → Bon app
	_ = ordersDS.Next()
	_ = ordersDS.Next()
	r.SetCalcContext(ordersDS)

	val2, err := r.Calc("[Orders.Customers.CompanyName]")
	if err != nil {
		t.Fatalf("row 2: Calc [Orders.Customers.CompanyName]: %v", err)
	}
	if val2 != "Bon app" {
		t.Errorf("row 2: got %v, want Bon app", val2)
	}
}

// TestInjectRelatedFields_NameBasedRelation_MissingParentDS exercises the
// code path where FindDataSourceByAlias returns nil for the parent (rel skipped).
func TestInjectRelatedFields_NameBasedRelation_MissingParentDS(t *testing.T) {
	r := reportpkg.NewReport()

	ordersDS := makeDS("Orders", []string{"OrderID", "CustomerID"}, []map[string]any{
		{"OrderID": "1", "CustomerID": "ALFKI"},
	})

	dict := r.Dictionary()
	dict.AddDataSource(ordersDS)

	// Parent DS "Customers" is NOT registered in the dictionary.
	dict.AddRelation(&data.Relation{
		Name:             "BadRelation",
		ChildSourceName:  "Orders",
		ParentSourceName: "Customers", // not in dict
		ChildColumnNames: []string{"CustomerID"},
		ParentColumnNames: []string{"CustomerID"},
	})

	_ = ordersDS.First()
	r.SetCalcContext(ordersDS)

	// Should not panic; expression just evaluates as unknown.
	_, err := r.Calc("[Orders.Customers.CompanyName]")
	if err == nil {
		t.Error("expected error for unknown related field, got nil")
	}
}

// TestInjectRelatedFields_EmptyColumns exercises the guard that skips relations
// when neither ParentColumns nor ParentColumnNames are populated.
func TestInjectRelatedFields_EmptyColumns(t *testing.T) {
	r := reportpkg.NewReport()

	parentDS := makeDS("Parent", []string{"ID", "Label"}, []map[string]any{
		{"ID": "1", "Label": "Alpha"},
	})
	childDS := makeDS("Child", []string{"ID"}, []map[string]any{
		{"ID": "1"},
	})

	dict := r.Dictionary()
	dict.AddDataSource(parentDS)
	dict.AddDataSource(childDS)

	// Relation with no column info at all — must be skipped without panic.
	dict.AddRelation(&data.Relation{
		Name:             "EmptyColsRelation",
		ParentDataSource: parentDS,
		ChildDataSource:  childDS,
		// No ParentColumns/ChildColumns set.
	})

	_ = childDS.First()
	r.SetCalcContext(childDS)

	// Expression should be unresolvable (relation skipped).
	_, err := r.Calc("[Child.Parent.Label]")
	if err == nil {
		t.Error("expected error when relation has no columns, got nil")
	}
}

// TestInjectRelatedFields_RelationWithNoPointerAndNoName covers the else branch
// (both pointer and source-name fields are absent) so it is simply skipped.
func TestInjectRelatedFields_RelationWithNoPointerAndNoName(t *testing.T) {
	r := reportpkg.NewReport()

	childDS := makeDS("Child", []string{"ID"}, []map[string]any{
		{"ID": "1"},
	})

	dict := r.Dictionary()
	dict.AddDataSource(childDS)

	// Completely empty relation (no pointer, no name) — must be skipped.
	dict.AddRelation(&data.Relation{Name: "Ghost"})

	_ = childDS.First()
	r.SetCalcContext(childDS)

	// No panic expected; expression evaluates to an error.
	_, err := r.Calc("[Child.Parent.Label]")
	if err == nil {
		t.Error("expected error for ghost relation, got nil")
	}
}

// ---------------------------------------------------------------------------
// injectRelatedFieldsFrom – grandparent traversal (the main coverage target)
// ---------------------------------------------------------------------------

// TestInjectRelatedFieldsFrom_ThreeHop is the primary test for
// injectRelatedFieldsFrom. It sets up a 3-level chain:
//
//	Shippers ← Orders ← OrderDetails
//
// and expects that an expression referencing
// [Order Details.Orders.Shippers.Phone] resolves correctly.
func TestInjectRelatedFieldsFrom_ThreeHop_PointerRelation(t *testing.T) {
	r := reportpkg.NewReport()

	shippersDS := makeDS("Shippers", []string{"ShipperID", "CompanyName", "Phone"}, []map[string]any{
		{"ShipperID": "1", "CompanyName": "Speedy Express", "Phone": "(503) 555-9831"},
		{"ShipperID": "2", "CompanyName": "United Package", "Phone": "(503) 555-3199"},
		{"ShipperID": "3", "CompanyName": "Federal Shipping", "Phone": "(503) 555-9931"},
	})

	ordersDS := makeDS("Orders", []string{"OrderID", "CustomerID", "ShipVia"}, []map[string]any{
		{"OrderID": "10248", "CustomerID": "VINET", "ShipVia": "3"},
		{"OrderID": "10249", "CustomerID": "TOMSP", "ShipVia": "1"},
		{"OrderID": "10250", "CustomerID": "HANAR", "ShipVia": "2"},
	})

	orderDetailsDS := makeDS("Order Details", []string{"OrderID", "ProductID", "UnitPrice"}, []map[string]any{
		{"OrderID": "10248", "ProductID": "11", "UnitPrice": "14.0"},
		{"OrderID": "10249", "ProductID": "14", "UnitPrice": "18.6"},
		{"OrderID": "10250", "ProductID": "65", "UnitPrice": "16.8"},
	})

	dict := r.Dictionary()
	dict.AddDataSource(shippersDS)
	dict.AddDataSource(ordersDS)
	dict.AddDataSource(orderDetailsDS)

	// Relation 1: Orders is child of Shippers (ShipVia → ShipperID).
	dict.AddRelation(&data.Relation{
		Name:             "ShippersOrders",
		ParentDataSource: shippersDS,
		ChildDataSource:  ordersDS,
		ParentColumns:    []string{"ShipperID"},
		ChildColumns:     []string{"ShipVia"},
	})

	// Relation 2: Order Details is child of Orders (OrderID → OrderID).
	dict.AddRelation(&data.Relation{
		Name:             "OrdersOrderDetails",
		ParentDataSource: ordersDS,
		ChildDataSource:  orderDetailsDS,
		ParentColumns:    []string{"OrderID"},
		ChildColumns:     []string{"OrderID"},
	})

	// ---- Row 0: OrderID=10248, ShipVia=3 → Federal Shipping ----
	_ = orderDetailsDS.First()
	r.SetCalcContext(orderDetailsDS)

	val, err := r.Calc("[Order Details.Orders.Shippers.Phone]")
	if err != nil {
		t.Fatalf("row 0: Calc [Order Details.Orders.Shippers.Phone]: %v", err)
	}
	if val != "(503) 555-9931" {
		t.Errorf("row 0: got %v, want (503) 555-9931", val)
	}

	// Also check that the intermediate Orders field is accessible (1-hop).
	val2, err := r.Calc("[Order Details.Orders.CustomerID]")
	if err != nil {
		t.Fatalf("row 0: Calc [Order Details.Orders.CustomerID]: %v", err)
	}
	if val2 != "VINET" {
		t.Errorf("row 0 Orders.CustomerID: got %v, want VINET", val2)
	}

	// ---- Row 1: OrderID=10249, ShipVia=1 → Speedy Express ----
	_ = orderDetailsDS.Next()
	r.SetCalcContext(orderDetailsDS)

	val3, err := r.Calc("[Order Details.Orders.Shippers.Phone]")
	if err != nil {
		t.Fatalf("row 1: Calc [Order Details.Orders.Shippers.Phone]: %v", err)
	}
	if val3 != "(503) 555-9831" {
		t.Errorf("row 1: got %v, want (503) 555-9831", val3)
	}

	// ---- Row 2: OrderID=10250, ShipVia=2 → United Package ----
	_ = orderDetailsDS.Next()
	r.SetCalcContext(orderDetailsDS)

	val4, err := r.Calc("[Order Details.Orders.Shippers.Phone]")
	if err != nil {
		t.Fatalf("row 2: Calc [Order Details.Orders.Shippers.Phone]: %v", err)
	}
	if val4 != "(503) 555-3199" {
		t.Errorf("row 2: got %v, want (503) 555-3199", val4)
	}
}

// TestInjectRelatedFieldsFrom_ThreeHop_NameBasedRelation exercises the
// name-based (FRX-style) path inside injectRelatedFieldsFrom.
func TestInjectRelatedFieldsFrom_ThreeHop_NameBasedRelation(t *testing.T) {
	r := reportpkg.NewReport()

	countriesDS := makeDS("Countries", []string{"CountryCode", "CountryName"}, []map[string]any{
		{"CountryCode": "DE", "CountryName": "Germany"},
		{"CountryCode": "FR", "CountryName": "France"},
	})

	customersDS := makeDS("Customers", []string{"CustomerID", "CompanyName", "Country"}, []map[string]any{
		{"CustomerID": "ALFKI", "CompanyName": "Alfreds Futterkiste", "Country": "DE"},
		{"CustomerID": "BONAP", "CompanyName": "Bon app", "Country": "FR"},
	})

	ordersDS := makeDS("Orders", []string{"OrderID", "CustomerID", "Freight"}, []map[string]any{
		{"OrderID": "10643", "CustomerID": "ALFKI", "Freight": "29.46"},
		{"OrderID": "10730", "CustomerID": "BONAP", "Freight": "20.12"},
	})

	dict := r.Dictionary()
	dict.AddDataSource(countriesDS)
	dict.AddDataSource(customersDS)
	dict.AddDataSource(ordersDS)

	// Use name-based relations to exercise the ChildSourceName/ParentSourceName paths.
	// Customers is child of Countries (Country → CountryCode).
	dict.AddRelation(&data.Relation{
		Name:              "CountriesCustomers",
		ChildSourceName:   "Customers",
		ParentSourceName:  "Countries",
		ChildColumnNames:  []string{"Country"},
		ParentColumnNames: []string{"CountryCode"},
	})

	// Orders is child of Customers (CustomerID → CustomerID).
	dict.AddRelation(&data.Relation{
		Name:              "CustomersOrders",
		ChildSourceName:   "Orders",
		ParentSourceName:  "Customers",
		ChildColumnNames:  []string{"CustomerID"},
		ParentColumnNames: []string{"CustomerID"},
	})

	// First Orders row → ALFKI → Germany.
	_ = ordersDS.First()
	r.SetCalcContext(ordersDS)

	val, err := r.Calc("[Orders.Customers.Countries.CountryName]")
	if err != nil {
		t.Fatalf("row 0: Calc [Orders.Customers.Countries.CountryName]: %v", err)
	}
	if val != "Germany" {
		t.Errorf("row 0: got %v, want Germany", val)
	}

	// Second Orders row → BONAP → France.
	_ = ordersDS.Next()
	r.SetCalcContext(ordersDS)

	val2, err := r.Calc("[Orders.Customers.Countries.CountryName]")
	if err != nil {
		t.Fatalf("row 1: Calc [Orders.Customers.Countries.CountryName]: %v", err)
	}
	if val2 != "France" {
		t.Errorf("row 1: got %v, want France", val2)
	}
}

// TestInjectRelatedFieldsFrom_NoGrandParentMatch verifies that when no
// grandparent row matches the parent's join key, injectRelatedFieldsFrom does
// not inject any key and the expression returns an error (not a panic).
func TestInjectRelatedFieldsFrom_NoGrandParentMatch(t *testing.T) {
	r := reportpkg.NewReport()

	shippersDS := makeDS("Shippers", []string{"ShipperID", "Phone"}, []map[string]any{
		{"ShipperID": "99", "Phone": "555-0000"}, // only shipper 99
	})

	ordersDS := makeDS("Orders", []string{"OrderID", "ShipVia"}, []map[string]any{
		{"OrderID": "10248", "ShipVia": "3"}, // ShipVia=3 has no matching shipper
	})

	orderDetailsDS := makeDS("Order Details", []string{"OrderID"}, []map[string]any{
		{"OrderID": "10248"},
	})

	dict := r.Dictionary()
	dict.AddDataSource(shippersDS)
	dict.AddDataSource(ordersDS)
	dict.AddDataSource(orderDetailsDS)

	dict.AddRelation(&data.Relation{
		Name:             "ShippersOrders",
		ParentDataSource: shippersDS,
		ChildDataSource:  ordersDS,
		ParentColumns:    []string{"ShipperID"},
		ChildColumns:     []string{"ShipVia"},
	})

	dict.AddRelation(&data.Relation{
		Name:             "OrdersOrderDetails",
		ParentDataSource: ordersDS,
		ChildDataSource:  orderDetailsDS,
		ParentColumns:    []string{"OrderID"},
		ChildColumns:     []string{"OrderID"},
	})

	_ = orderDetailsDS.First()
	r.SetCalcContext(orderDetailsDS)

	_, err := r.Calc("[Order Details.Orders.Shippers.Phone]")
	if err == nil {
		t.Error("expected error when grandparent row not found, got nil")
	}
}

// TestInjectRelatedFieldsFrom_MissingGrandParentDS exercises the code path
// inside injectRelatedFieldsFrom where FindDataSourceByAlias returns nil for
// the grandparent DS (name-based relation, grandparent not registered).
func TestInjectRelatedFieldsFrom_MissingGrandParentDS(t *testing.T) {
	r := reportpkg.NewReport()

	customersDS := makeDS("Customers", []string{"CustomerID", "Country"}, []map[string]any{
		{"CustomerID": "ALFKI", "Country": "DE"},
	})

	ordersDS := makeDS("Orders", []string{"OrderID", "CustomerID"}, []map[string]any{
		{"OrderID": "1", "CustomerID": "ALFKI"},
	})

	dict := r.Dictionary()
	dict.AddDataSource(customersDS)
	dict.AddDataSource(ordersDS)

	// Name-based: grandparent "Countries" is NOT in dict.
	dict.AddRelation(&data.Relation{
		Name:              "CountriesCustomers",
		ChildSourceName:   "Customers",
		ParentSourceName:  "Countries", // not registered
		ChildColumnNames:  []string{"Country"},
		ParentColumnNames: []string{"CountryCode"},
	})

	dict.AddRelation(&data.Relation{
		Name:              "CustomersOrders",
		ChildSourceName:   "Orders",
		ParentSourceName:  "Customers",
		ChildColumnNames:  []string{"CustomerID"},
		ParentColumnNames: []string{"CustomerID"},
	})

	_ = ordersDS.First()
	r.SetCalcContext(ordersDS)

	// grandparent Countries not in dict → expression unresolvable.
	_, err := r.Calc("[Orders.Customers.Countries.CountryName]")
	if err == nil {
		t.Error("expected error when grandparent DS missing from dict, got nil")
	}
}

// TestInjectRelatedFieldsFrom_EmptyGrandParentColumns verifies that when the
// grandparent relation has no column info, it is skipped without panic.
func TestInjectRelatedFieldsFrom_EmptyGrandParentColumns(t *testing.T) {
	r := reportpkg.NewReport()

	grandParentDS := makeDS("GrandParent", []string{"ID", "Val"}, []map[string]any{
		{"ID": "1", "Val": "top"},
	})

	parentDS := makeDS("Parent", []string{"ID", "GrandParentID"}, []map[string]any{
		{"ID": "10", "GrandParentID": "1"},
	})

	childDS := makeDS("Child", []string{"ID", "ParentID"}, []map[string]any{
		{"ID": "100", "ParentID": "10"},
	})

	dict := r.Dictionary()
	dict.AddDataSource(grandParentDS)
	dict.AddDataSource(parentDS)
	dict.AddDataSource(childDS)

	// Child→Parent relation (has columns).
	dict.AddRelation(&data.Relation{
		Name:             "ParentChild",
		ParentDataSource: parentDS,
		ChildDataSource:  childDS,
		ParentColumns:    []string{"ID"},
		ChildColumns:     []string{"ParentID"},
	})

	// Parent→GrandParent relation with NO column info — should be skipped.
	dict.AddRelation(&data.Relation{
		Name:             "GrandParentParent",
		ParentDataSource: grandParentDS,
		ChildDataSource:  parentDS,
		// No ParentColumns / ChildColumns.
	})

	_ = childDS.First()
	r.SetCalcContext(childDS)

	// 1-hop (Child→Parent) should still work.
	// Note: coerceCalcValue converts string "1" → int64(1), so compare via Sprintf.
	val, err := r.Calc("[Child.Parent.GrandParentID]")
	if err != nil {
		t.Fatalf("1-hop Calc: %v", err)
	}
	if fmt.Sprintf("%v", val) != "1" {
		t.Errorf("1-hop: got %v, want 1", val)
	}

	// 3-hop should fail because the grandparent relation has no columns.
	_, err2 := r.Calc("[Child.Parent.GrandParent.Val]")
	if err2 == nil {
		t.Error("expected error when grandparent relation has no columns, got nil")
	}
}

// TestInjectRelatedFieldsFrom_RelationWithNoPointerAndNoName_GrandParent covers
// the else-branch in injectRelatedFieldsFrom where both pointer and source-name
// fields are absent (the relation is skipped entirely).
func TestInjectRelatedFieldsFrom_RelationWithNoPointerAndNoName_GrandParent(t *testing.T) {
	r := reportpkg.NewReport()

	parentDS := makeDS("Parent", []string{"ID", "Val"}, []map[string]any{
		{"ID": "1", "Val": "hello"},
	})

	childDS := makeDS("Child", []string{"ParentID"}, []map[string]any{
		{"ParentID": "1"},
	})

	dict := r.Dictionary()
	dict.AddDataSource(parentDS)
	dict.AddDataSource(childDS)

	// Valid child→parent relation.
	dict.AddRelation(&data.Relation{
		Name:             "ParentChild",
		ParentDataSource: parentDS,
		ChildDataSource:  childDS,
		ParentColumns:    []string{"ID"},
		ChildColumns:     []string{"ParentID"},
	})

	// Ghost relation — no pointers, no names; present only to exercise the else path.
	dict.AddRelation(&data.Relation{Name: "Ghost"})

	_ = childDS.First()
	r.SetCalcContext(childDS)

	// 1-hop should still work (Ghost is skipped).
	val, err := r.Calc("[Child.Parent.Val]")
	if err != nil {
		t.Fatalf("Calc [Child.Parent.Val]: %v", err)
	}
	if val != "hello" {
		t.Errorf("got %v, want hello", val)
	}
}

// TestInjectRelatedFields_MultipleChildRelations verifies that when the
// dictionary contains more than one relation where different data sources are
// the child, only the relation whose ChildAlias matches the current calc
// context DS is used (the other is ignored).
func TestInjectRelatedFields_MultipleChildRelations(t *testing.T) {
	r := reportpkg.NewReport()

	suppliersDS := makeDS("Suppliers", []string{"SupplierID", "SupplierName"}, []map[string]any{
		{"SupplierID": "1", "SupplierName": "Exotic Liquids"},
	})

	productsDS := makeDS("Products", []string{"ProductID", "ProductName", "SupplierID"}, []map[string]any{
		{"ProductID": "1", "ProductName": "Chai", "SupplierID": "1"},
	})

	categoriesDS := makeDS("Categories", []string{"CategoryID", "CategoryName"}, []map[string]any{
		{"CategoryID": "1", "CategoryName": "Beverages"},
	})

	orderDetailsDS := makeDS("Order Details", []string{"OrderID", "ProductID"}, []map[string]any{
		{"OrderID": "10248", "ProductID": "1"},
	})

	dict := r.Dictionary()
	dict.AddDataSource(suppliersDS)
	dict.AddDataSource(productsDS)
	dict.AddDataSource(categoriesDS)
	dict.AddDataSource(orderDetailsDS)

	// Order Details → Products
	dict.AddRelation(&data.Relation{
		Name:             "ProductsOrderDetails",
		ParentDataSource: productsDS,
		ChildDataSource:  orderDetailsDS,
		ParentColumns:    []string{"ProductID"},
		ChildColumns:     []string{"ProductID"},
	})

	// Products → Suppliers (unrelated to orderDetails calc context at 1-hop)
	dict.AddRelation(&data.Relation{
		Name:             "SuppliersProducts",
		ParentDataSource: suppliersDS,
		ChildDataSource:  productsDS,
		ParentColumns:    []string{"SupplierID"},
		ChildColumns:     []string{"SupplierID"},
	})

	// Categories → Products (another unrelated relation)
	dict.AddRelation(&data.Relation{
		Name:             "CategoriesProducts",
		ParentDataSource: categoriesDS,
		ChildDataSource:  productsDS,
		ParentColumns:    []string{"CategoryID"},
		ChildColumns:     []string{"CategoryID"},
	})

	_ = orderDetailsDS.First()
	r.SetCalcContext(orderDetailsDS)

	// Only the Products-related injection matters for Order Details context.
	val, err := r.Calc("[Order Details.Products.ProductName]")
	if err != nil {
		t.Fatalf("Calc [Order Details.Products.ProductName]: %v", err)
	}
	if val != "Chai" {
		t.Errorf("got %v, want Chai", val)
	}
}

// TestInjectRelatedFields_CalcText verifies that CalcText (multi-bracket
// templates) correctly resolves related fields injected by injectRelatedFields.
func TestInjectRelatedFields_CalcText(t *testing.T) {
	r := reportpkg.NewReport()

	customersDS := makeDS("Customers", []string{"CustomerID", "CompanyName"}, []map[string]any{
		{"CustomerID": "ALFKI", "CompanyName": "Alfreds Futterkiste"},
	})

	ordersDS := makeDS("Orders", []string{"OrderID", "CustomerID", "Freight"}, []map[string]any{
		{"OrderID": "10643", "CustomerID": "ALFKI", "Freight": "29.46"},
	})

	dict := r.Dictionary()
	dict.AddDataSource(customersDS)
	dict.AddDataSource(ordersDS)
	dict.AddRelation(&data.Relation{
		Name:             "CustomersOrders",
		ParentDataSource: customersDS,
		ChildDataSource:  ordersDS,
		ParentColumns:    []string{"CustomerID"},
		ChildColumns:     []string{"CustomerID"},
	})

	_ = ordersDS.First()
	r.SetCalcContext(ordersDS)

	text, err := r.CalcText("Order [Orders.OrderID] for [Orders.Customers.CompanyName]")
	if err != nil {
		t.Fatalf("CalcText: %v", err)
	}
	want := "Order 10643 for Alfreds Futterkiste"
	if text != want {
		t.Errorf("got %q, want %q", text, want)
	}
}
