package reportpkg_test

import (
	"testing"

	"github.com/andrewloable/go-fastreport/data"
	"github.com/andrewloable/go-fastreport/reportpkg"
)

func TestReport_Calc_Parameter(t *testing.T) {
	r := reportpkg.NewReport()
	param := &data.Parameter{Name: "Greeting", Value: "Hello"}
	r.Dictionary().AddParameter(param)

	val, err := r.Calc("[Greeting]")
	if err != nil {
		t.Fatalf("Calc: %v", err)
	}
	if val != "Hello" {
		t.Errorf("got %v, want Hello", val)
	}
}

func TestReport_Calc_SystemVariable(t *testing.T) {
	r := reportpkg.NewReport()
	r.Dictionary().SetSystemVariable("PageNumber", 3)

	val, err := r.Calc("[PageNumber]")
	if err != nil {
		t.Fatalf("Calc: %v", err)
	}
	if val != 3 {
		t.Errorf("got %v, want 3", val)
	}
}

func TestReport_Calc_ArithmeticExpression(t *testing.T) {
	r := reportpkg.NewReport()
	param := &data.Parameter{Name: "Price", Value: 10.0}
	r.Dictionary().AddParameter(param)

	val, err := r.Calc("[Price] * 2")
	if err != nil {
		t.Fatalf("Calc: %v", err)
	}
	if val != 20.0 {
		t.Errorf("got %v, want 20.0", val)
	}
}

func TestReport_CalcText_Template(t *testing.T) {
	r := reportpkg.NewReport()
	param := &data.Parameter{Name: "Name", Value: "World"}
	r.Dictionary().AddParameter(param)

	text, err := r.CalcText("Hello [Name]!")
	if err != nil {
		t.Fatalf("CalcText: %v", err)
	}
	if text != "Hello World!" {
		t.Errorf("got %q, want %q", text, "Hello World!")
	}
}

func TestReport_CalcText_NoExpressions(t *testing.T) {
	r := reportpkg.NewReport()
	text, err := r.CalcText("Static text")
	if err != nil {
		t.Fatalf("CalcText: %v", err)
	}
	if text != "Static text" {
		t.Errorf("got %q, want %q", text, "Static text")
	}
}

func TestReport_Calc_DataSourceColumn(t *testing.T) {
	r := reportpkg.NewReport()

	ds := data.NewBaseDataSource("Customers")
	ds.SetAlias("Customers")
	ds.AddColumn(data.Column{Name: "Name"})
	ds.AddRow(map[string]any{"Name": "Alice"})
	_ = ds.Init()
	_ = ds.First()

	r.SetCalcContext(ds)

	val, err := r.Calc("[Name]")
	if err != nil {
		t.Fatalf("Calc: %v", err)
	}
	if val != "Alice" {
		t.Errorf("got %v, want Alice", val)
	}
}

// TestReport_Calc_RelatedDataSourceField verifies that expressions like
// [Order Details.Products.ProductName] are resolved when the current calc
// context is "Order Details" and a Products→Order Details relation exists.
// This mirrors the Cascaded Data Filtering report pattern.
func TestReport_Calc_RelatedDataSourceField(t *testing.T) {
	r := reportpkg.NewReport()

	// Products (parent data source).
	productsDS := data.NewBaseDataSource("Products")
	productsDS.SetAlias("Products")
	productsDS.AddColumn(data.Column{Name: "ProductID"})
	productsDS.AddColumn(data.Column{Name: "ProductName"})
	productsDS.AddRow(map[string]any{"ProductID": "1", "ProductName": "Chai"})
	productsDS.AddRow(map[string]any{"ProductID": "2", "ProductName": "Chang"})
	_ = productsDS.Init()

	// Order Details (child data source).
	orderDetailsDS := data.NewBaseDataSource("Order Details")
	orderDetailsDS.SetAlias("Order Details")
	orderDetailsDS.AddColumn(data.Column{Name: "OrderID"})
	orderDetailsDS.AddColumn(data.Column{Name: "ProductID"})
	orderDetailsDS.AddColumn(data.Column{Name: "UnitPrice"})
	orderDetailsDS.AddRow(map[string]any{"OrderID": "10248", "ProductID": "1", "UnitPrice": 14.0})
	orderDetailsDS.AddRow(map[string]any{"OrderID": "10248", "ProductID": "2", "UnitPrice": 9.8})
	_ = orderDetailsDS.Init()

	// Register data sources and relation in the dictionary.
	dict := r.Dictionary()
	dict.AddDataSource(productsDS)
	dict.AddDataSource(orderDetailsDS)
	rel := &data.Relation{
		Name:             "ProductsOrderDetails",
		ParentDataSource: productsDS,
		ChildDataSource:  orderDetailsDS,
		ParentColumns:    []string{"ProductID"},
		ChildColumns:     []string{"ProductID"},
	}
	dict.AddRelation(rel)

	// Position Order Details on the first row (ProductID=1 → Chai).
	_ = orderDetailsDS.First()
	r.SetCalcContext(orderDetailsDS)

	val, err := r.Calc("[Order Details.Products.ProductName]")
	if err != nil {
		t.Fatalf("Calc [Order Details.Products.ProductName]: %v", err)
	}
	if val != "Chai" {
		t.Errorf("got %v, want Chai", val)
	}

	// Advance to second row (ProductID=2 → Chang) and verify again.
	_ = orderDetailsDS.Next()
	r.SetCalcContext(orderDetailsDS)

	val2, err := r.Calc("[Order Details.Products.ProductName]")
	if err != nil {
		t.Fatalf("Calc [Order Details.Products.ProductName] row2: %v", err)
	}
	if val2 != "Chang" {
		t.Errorf("got %v, want Chang", val2)
	}
}

// TestReport_Calc_RelatedDataSourceField_NoMatch verifies that when no parent
// row matches the child join key, the related field is not injected and the
// expression evaluates to an error (the key is absent from the env), not a panic.
func TestReport_Calc_RelatedDataSourceField_NoMatch(t *testing.T) {
	r := reportpkg.NewReport()

	productsDS := data.NewBaseDataSource("Products")
	productsDS.SetAlias("Products")
	productsDS.AddColumn(data.Column{Name: "ProductID"})
	productsDS.AddColumn(data.Column{Name: "ProductName"})
	productsDS.AddRow(map[string]any{"ProductID": "99", "ProductName": "Unknown"})
	_ = productsDS.Init()

	orderDetailsDS := data.NewBaseDataSource("Order Details")
	orderDetailsDS.SetAlias("Order Details")
	orderDetailsDS.AddColumn(data.Column{Name: "ProductID"})
	orderDetailsDS.AddRow(map[string]any{"ProductID": "1"}) // no matching Products row
	_ = orderDetailsDS.Init()
	_ = orderDetailsDS.First()

	dict := r.Dictionary()
	dict.AddDataSource(productsDS)
	dict.AddDataSource(orderDetailsDS)
	dict.AddRelation(&data.Relation{
		Name:             "ProductsOrderDetails",
		ParentDataSource: productsDS,
		ChildDataSource:  orderDetailsDS,
		ParentColumns:    []string{"ProductID"},
		ChildColumns:     []string{"ProductID"},
	})

	r.SetCalcContext(orderDetailsDS)

	// When no parent row matches, Calc should return an error (unknown identifier)
	// rather than panicking or returning a wrong value.
	_, err := r.Calc("[Order Details.Products.ProductName]")
	if err == nil {
		t.Error("expected error when no matching parent row, got nil")
	}
}
