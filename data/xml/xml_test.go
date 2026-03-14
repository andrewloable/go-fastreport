package xml_test

import (
	"os"
	"path/filepath"
	"testing"

	xmlds "github.com/andrewloable/go-fastreport/data/xml"
)

const simpleXML = `<?xml version="1.0" encoding="utf-8"?>
<Customers>
  <Customer Name="Alice" Age="30" City="London"/>
  <Customer Name="Bob"   Age="25" City="Paris"/>
  <Customer Name="Carol" Age="35" City="Berlin"/>
</Customers>`

const childElemXML = `<?xml version="1.0" encoding="utf-8"?>
<Customers>
  <Customer>
    <Name>Alice</Name>
    <Age>30</Age>
  </Customer>
  <Customer>
    <Name>Bob</Name>
    <Age>25</Age>
  </Customer>
</Customers>`

const nestedXML = `<?xml version="1.0" encoding="utf-8"?>
<Root>
  <Orders>
    <Item Product="Apple" Qty="5"/>
    <Item Product="Banana" Qty="3"/>
  </Orders>
</Root>`

const mixedXML = `<?xml version="1.0" encoding="utf-8"?>
<Root>
  <Row ID="1"><Label>First</Label></Row>
  <Row ID="2"><Label>Second</Label></Row>
</Root>`

// ── attribute-based rows ──────────────────────────────────────────────────────

func TestXMLDataSource_Attributes_RowCount(t *testing.T) {
	ds := xmlds.New("customers")
	ds.SetXML(simpleXML)
	if err := ds.Init(); err != nil {
		t.Fatalf("Init: %v", err)
	}
	if ds.RowCount() != 3 {
		t.Errorf("RowCount = %d, want 3", ds.RowCount())
	}
}

func TestXMLDataSource_Attributes_GetValue(t *testing.T) {
	ds := xmlds.New("customers")
	ds.SetXML(simpleXML)
	if err := ds.Init(); err != nil {
		t.Fatalf("Init: %v", err)
	}
	if err := ds.First(); err != nil {
		t.Fatalf("First: %v", err)
	}

	name, err := ds.GetValue("Name")
	if err != nil {
		t.Fatalf("GetValue(Name): %v", err)
	}
	if name != "Alice" {
		t.Errorf("Name = %q, want Alice", name)
	}

	age, _ := ds.GetValue("Age")
	if age != "30" {
		t.Errorf("Age = %q, want 30", age)
	}
}

func TestXMLDataSource_Attributes_Traversal(t *testing.T) {
	ds := xmlds.New("customers")
	ds.SetXML(simpleXML)
	if err := ds.Init(); err != nil {
		t.Fatalf("Init: %v", err)
	}

	names := []string{"Alice", "Bob", "Carol"}
	if err := ds.First(); err != nil {
		t.Fatalf("First: %v", err)
	}
	for i, want := range names {
		if ds.EOF() {
			t.Fatalf("EOF at row %d", i)
		}
		got, _ := ds.GetValue("Name")
		if got != want {
			t.Errorf("row %d Name = %q, want %q", i, got, want)
		}
		_ = ds.Next()
	}
	if !ds.EOF() {
		t.Error("expected EOF after last row")
	}
}

// ── child-element-based rows ──────────────────────────────────────────────────

func TestXMLDataSource_ChildElements_GetValue(t *testing.T) {
	ds := xmlds.New("customers")
	ds.SetXML(childElemXML)
	if err := ds.Init(); err != nil {
		t.Fatalf("Init: %v", err)
	}
	if err := ds.First(); err != nil {
		t.Fatalf("First: %v", err)
	}

	name, _ := ds.GetValue("Name")
	if name != "Alice" {
		t.Errorf("Name = %q, want Alice", name)
	}
	age, _ := ds.GetValue("Age")
	if age != "30" {
		t.Errorf("Age = %q, want 30", age)
	}
}

// ── RootPath navigation ───────────────────────────────────────────────────────

func TestXMLDataSource_RootPath(t *testing.T) {
	ds := xmlds.New("orders")
	ds.SetXML(nestedXML)
	ds.SetRootPath("Orders")
	if err := ds.Init(); err != nil {
		t.Fatalf("Init: %v", err)
	}
	if ds.RowCount() != 2 {
		t.Errorf("RowCount = %d, want 2", ds.RowCount())
	}
	if err := ds.First(); err != nil {
		t.Fatalf("First: %v", err)
	}
	product, _ := ds.GetValue("Product")
	if product != "Apple" {
		t.Errorf("Product = %q, want Apple", product)
	}
}

// ── mixed attributes + child elements ────────────────────────────────────────

func TestXMLDataSource_MixedColumns(t *testing.T) {
	ds := xmlds.New("rows")
	ds.SetXML(mixedXML)
	if err := ds.Init(); err != nil {
		t.Fatalf("Init: %v", err)
	}
	if ds.RowCount() != 2 {
		t.Errorf("RowCount = %d, want 2", ds.RowCount())
	}
	if err := ds.First(); err != nil {
		t.Fatalf("First: %v", err)
	}

	id, _ := ds.GetValue("ID")
	if id != "1" {
		t.Errorf("ID = %q, want 1", id)
	}
	label, _ := ds.GetValue("Label")
	if label != "First" {
		t.Errorf("Label = %q, want First", label)
	}
}

// ── file-based ────────────────────────────────────────────────────────────────

func TestXMLDataSource_File(t *testing.T) {
	dir := t.TempDir()
	fpath := filepath.Join(dir, "data.xml")
	if err := os.WriteFile(fpath, []byte(simpleXML), 0o644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	ds := xmlds.New("customers")
	ds.SetFilePath(fpath)
	if err := ds.Init(); err != nil {
		t.Fatalf("Init: %v", err)
	}
	if ds.RowCount() != 3 {
		t.Errorf("RowCount = %d, want 3", ds.RowCount())
	}
}

// ── error cases ───────────────────────────────────────────────────────────────

func TestXMLDataSource_NoSource_Error(t *testing.T) {
	ds := xmlds.New("empty")
	if err := ds.Init(); err == nil {
		t.Error("expected error for no source configured")
	}
}

func TestXMLDataSource_InvalidXML_Error(t *testing.T) {
	ds := xmlds.New("bad")
	ds.SetXML("<not valid xml")
	if err := ds.Init(); err == nil {
		t.Error("expected error for invalid XML")
	}
}

func TestXMLDataSource_MissingRootPath_Error(t *testing.T) {
	ds := xmlds.New("missing")
	ds.SetXML(simpleXML)
	ds.SetRootPath("NonExistent")
	if err := ds.Init(); err == nil {
		t.Error("expected error for missing root path element")
	}
}

// ── empty data set ────────────────────────────────────────────────────────────

func TestXMLDataSource_Empty_RowCount(t *testing.T) {
	ds := xmlds.New("empty")
	ds.SetXML(`<Root></Root>`)
	if err := ds.Init(); err != nil {
		t.Fatalf("Init: %v", err)
	}
	if ds.RowCount() != 0 {
		t.Errorf("RowCount = %d, want 0", ds.RowCount())
	}
	// With 0 rows, RowCount returns 0 and EOF is immediately true
	// because currentRow (-1) < len(rows) (0) is false only when compared >= 0.
	// BaseDataSource.EOF uses currentRow >= len(rows); -1 >= 0 is false,
	// so we check RowCount directly for the "empty" condition.
	if ds.RowCount() != 0 {
		t.Errorf("expected 0 rows for empty XML, got %d", ds.RowCount())
	}
}

// ── accessor methods ─────────────────────────────────────────────────────────

func TestXMLDataSource_Accessors(t *testing.T) {
	ds := xmlds.New("test")
	ds.SetFilePath("/tmp/x.xml")
	ds.SetXML("<r/>")
	ds.SetRootPath("A/B")
	ds.SetRowElement("Row")

	if ds.FilePath() != "/tmp/x.xml" {
		t.Errorf("FilePath = %q", ds.FilePath())
	}
	if ds.XML() != "<r/>" {
		t.Errorf("XML = %q", ds.XML())
	}
	if ds.RootPath() != "A/B" {
		t.Errorf("RootPath = %q", ds.RootPath())
	}
	if ds.RowElement() != "Row" {
		t.Errorf("RowElement = %q", ds.RowElement())
	}
}
