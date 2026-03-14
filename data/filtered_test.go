package data_test

import (
	"testing"

	"github.com/andrewloable/go-fastreport/data"
)

// buildOrders creates a BaseDataSource with 3 orders.
func buildOrders() *data.BaseDataSource {
	ds := data.NewBaseDataSource("Orders")
	ds.AddColumn(data.Column{Name: "OrderID"})
	ds.AddColumn(data.Column{Name: "Customer"})
	ds.AddRow(map[string]any{"OrderID": "1", "Customer": "Alice"})
	ds.AddRow(map[string]any{"OrderID": "2", "Customer": "Bob"})
	ds.AddRow(map[string]any{"OrderID": "3", "Customer": "Carol"})
	_ = ds.Init()
	return ds
}

// buildOrderItems creates a BaseDataSource with items belonging to orders 1 and 2.
func buildOrderItems() *data.BaseDataSource {
	ds := data.NewBaseDataSource("OrderItems")
	ds.AddColumn(data.Column{Name: "ItemID"})
	ds.AddColumn(data.Column{Name: "OrderID"})
	ds.AddColumn(data.Column{Name: "Product"})
	ds.AddRow(map[string]any{"ItemID": "1", "OrderID": "1", "Product": "Widget"})
	ds.AddRow(map[string]any{"ItemID": "2", "OrderID": "1", "Product": "Gadget"})
	ds.AddRow(map[string]any{"ItemID": "3", "OrderID": "2", "Product": "Doohickey"})
	ds.AddRow(map[string]any{"ItemID": "4", "OrderID": "3", "Product": "Thingamajig"})
	_ = ds.Init()
	return ds
}

func TestFilteredDataSource_BasicFilter(t *testing.T) {
	items := buildOrderItems()

	// Filter to only items where OrderID == "1"
	fds, err := data.NewFilteredDataSource(items, []string{"OrderID"}, []string{"1"})
	if err != nil {
		t.Fatalf("NewFilteredDataSource: %v", err)
	}

	if fds.RowCount() != 2 {
		t.Errorf("expected 2 filtered rows for OrderID=1, got %d", fds.RowCount())
	}

	if err := fds.First(); err != nil {
		t.Fatalf("First: %v", err)
	}

	products := []string{}
	for !fds.EOF() {
		v, _ := fds.GetValue("Product")
		products = append(products, v.(string))
		_ = fds.Next()
	}
	if len(products) != 2 {
		t.Errorf("expected 2 products, got %v", products)
	}
	if products[0] != "Widget" || products[1] != "Gadget" {
		t.Errorf("unexpected products: %v", products)
	}
}

func TestFilteredDataSource_NoMatch(t *testing.T) {
	items := buildOrderItems()

	// Filter to items for a non-existent order.
	fds, err := data.NewFilteredDataSource(items, []string{"OrderID"}, []string{"99"})
	if err != nil {
		t.Fatalf("NewFilteredDataSource: %v", err)
	}

	if fds.RowCount() != 0 {
		t.Errorf("expected 0 rows, got %d", fds.RowCount())
	}

	_ = fds.First()
	if !fds.EOF() {
		t.Error("expected EOF for empty filter result")
	}
}

func TestFilteredDataSource_MultipleFilters(t *testing.T) {
	items := buildOrderItems()

	// Filter on two parent values — only row 3 matches OrderID=2.
	fds, err := data.NewFilteredDataSource(items, []string{"OrderID"}, []string{"2"})
	if err != nil {
		t.Fatalf("NewFilteredDataSource: %v", err)
	}

	if fds.RowCount() != 1 {
		t.Errorf("expected 1 row for OrderID=2, got %d", fds.RowCount())
	}

	_ = fds.First()
	v, _ := fds.GetValue("Product")
	if v.(string) != "Doohickey" {
		t.Errorf("expected 'Doohickey', got %v", v)
	}
}
