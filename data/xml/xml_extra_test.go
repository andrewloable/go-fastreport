package xml_test

import (
	"os"
	"path/filepath"
	"testing"

	xmlds "github.com/andrewloable/go-fastreport/data/xml"
)

// TestXMLDataSource_FileNotFound covers the readSource file open error path.
func TestXMLDataSource_FileNotFound(t *testing.T) {
	ds := xmlds.New("bad")
	ds.SetFilePath("/nonexistent/path/data.xml")
	err := ds.Init()
	if err == nil {
		t.Error("expected error for non-existent file")
	}
}

// TestXMLDataSource_EmptyRootPath covers parsing with no rootPath (auto-detect
// from root element children).
func TestXMLDataSource_EmptyRootPath(t *testing.T) {
	xml := `<?xml version="1.0"?>
<Root>
  <Item Name="A"/>
  <Item Name="B"/>
</Root>`
	ds := xmlds.New("items")
	ds.SetXML(xml)
	// No rootPath — root element is <Root>, children are the row elements.
	if err := ds.Init(); err != nil {
		t.Fatalf("Init: %v", err)
	}
	if ds.RowCount() != 2 {
		t.Errorf("RowCount = %d, want 2", ds.RowCount())
	}
}

// TestXMLDataSource_SkipToChildElement_NotFound covers the EndElement error
// path in skipToChildElement: when the parent ends before the target child.
func TestXMLDataSource_SkipToChildElement_NotFound(t *testing.T) {
	xml := `<Root><Other/></Root>`
	ds := xmlds.New("missing")
	ds.SetXML(xml)
	ds.SetRootPath("NonExistent")
	err := ds.Init()
	if err == nil {
		t.Error("expected error when rootPath child element is not found")
	}
}

// TestXMLDataSource_SkipToChildElement_Skip covers the skip-subtree branch.
func TestXMLDataSource_SkipToChildElement_Skip(t *testing.T) {
	// <Orders> is nested inside <Root>; <Other> subtree is skipped.
	xml := `<?xml version="1.0"?>
<Root>
  <Other><Deep/></Other>
  <Orders>
    <Item Product="Widget"/>
    <Item Product="Gadget"/>
  </Orders>
</Root>`
	ds := xmlds.New("orders")
	ds.SetXML(xml)
	ds.SetRootPath("Orders")
	if err := ds.Init(); err != nil {
		t.Fatalf("Init: %v", err)
	}
	if ds.RowCount() != 2 {
		t.Errorf("RowCount = %d, want 2", ds.RowCount())
	}
}

// TestXMLDataSource_ReadChildren_NestedElement covers the nested-element skip
// inside readElementText.
func TestXMLDataSource_ReadChildren_NestedElement(t *testing.T) {
	// Child element contains a nested element that should be skipped.
	xml := `<?xml version="1.0"?>
<Root>
  <Row>
    <Name>Alice<Extra/></Name>
    <Age>30</Age>
  </Row>
</Root>`
	ds := xmlds.New("rows")
	ds.SetXML(xml)
	if err := ds.Init(); err != nil {
		t.Fatalf("Init: %v", err)
	}
	if ds.RowCount() != 1 {
		t.Errorf("RowCount = %d, want 1", ds.RowCount())
	}
	_ = ds.First()
	name, _ := ds.GetValue("Name")
	if name != "Alice" {
		t.Errorf("Name = %q, want Alice", name)
	}
}

// TestXMLDataSource_ReadChildren_MismatchedEndElement covers the end-element
// that does not match the parent local name (just ignored/continued).
func TestXMLDataSource_ReadChildren_MismatchedEndElement(t *testing.T) {
	// Normally well-formed XML; this tests the EndElement != parentLocal path
	// by having a child element that has the same name as the outer container
	// but at a different nesting level.
	xml := `<?xml version="1.0"?>
<Outer>
  <Row>
    <Data>value</Data>
  </Row>
  <Row>
    <Data>other</Data>
  </Row>
</Outer>`
	ds := xmlds.New("rows")
	ds.SetXML(xml)
	if err := ds.Init(); err != nil {
		t.Fatalf("Init: %v", err)
	}
	if ds.RowCount() != 2 {
		t.Errorf("RowCount = %d, want 2", ds.RowCount())
	}
}

// TestXMLDataSource_RowElement covers the SetRowElement configuration.
func TestXMLDataSource_RowElement(t *testing.T) {
	xml := `<?xml version="1.0"?>
<Root>
  <Product Name="Apple" Price="1.50"/>
  <Product Name="Banana" Price="0.99"/>
  <Category Name="Fruit"/>
</Root>`
	ds := xmlds.New("products")
	ds.SetXML(xml)
	ds.SetRowElement("Product")
	if err := ds.Init(); err != nil {
		t.Fatalf("Init: %v", err)
	}
	// Only Product rows should be included; Category should be skipped.
	if ds.RowCount() != 2 {
		t.Errorf("RowCount = %d, want 2", ds.RowCount())
	}
	_ = ds.First()
	name, _ := ds.GetValue("Name")
	if name != "Apple" {
		t.Errorf("Name = %q, want Apple", name)
	}
}

// TestXMLDataSource_DeepRootPath covers a two-segment rootPath.
func TestXMLDataSource_DeepRootPath(t *testing.T) {
	xml := `<?xml version="1.0"?>
<Root>
  <Level1>
    <Items>
      <Item ID="1"/>
      <Item ID="2"/>
    </Items>
  </Level1>
</Root>`
	ds := xmlds.New("items")
	ds.SetXML(xml)
	ds.SetRootPath("Level1/Items")
	if err := ds.Init(); err != nil {
		t.Fatalf("Init: %v", err)
	}
	if ds.RowCount() != 2 {
		t.Errorf("RowCount = %d, want 2", ds.RowCount())
	}
}

// TestXMLDataSource_NoRootElement covers the error when XML has no root element.
func TestXMLDataSource_NoRootElement(t *testing.T) {
	ds := xmlds.New("noroot")
	ds.SetXML(`<!-- just a comment -->`)
	err := ds.Init()
	if err == nil {
		t.Error("expected error when XML has no root element")
	}
}

// TestXMLDataSource_FileSource_ReadAll covers reading from a file.
func TestXMLDataSource_FileSource_ReadAll(t *testing.T) {
	xml := `<?xml version="1.0"?>
<Customers>
  <Customer Name="X" Age="1"/>
  <Customer Name="Y" Age="2"/>
</Customers>`
	tmp := t.TempDir()
	fpath := filepath.Join(tmp, "data.xml")
	if err := os.WriteFile(fpath, []byte(xml), 0o644); err != nil {
		t.Fatal(err)
	}
	ds := xmlds.New("customers")
	ds.SetFilePath(fpath)
	if err := ds.Init(); err != nil {
		t.Fatalf("Init: %v", err)
	}
	if ds.RowCount() != 2 {
		t.Errorf("RowCount = %d, want 2", ds.RowCount())
	}
}

// TestXMLDataSource_SkipSiblingWithDifferentName covers the path where
// a sibling element with a different name from rowElem is skipped.
func TestXMLDataSource_SkipSiblingWithDifferentName(t *testing.T) {
	xml := `<?xml version="1.0"?>
<Root>
  <Row Name="Alice"/>
  <Header Title="ignored"/>
  <Row Name="Bob"/>
</Root>`
	ds := xmlds.New("rows")
	ds.SetXML(xml)
	ds.SetRowElement("Row")
	if err := ds.Init(); err != nil {
		t.Fatalf("Init: %v", err)
	}
	if ds.RowCount() != 2 {
		t.Errorf("RowCount = %d, want 2", ds.RowCount())
	}
	_ = ds.First()
	name, _ := ds.GetValue("Name")
	if name != "Alice" {
		t.Errorf("Name = %q, want Alice", name)
	}
}

// TestXMLDataSource_readElementText_Whitespace covers TrimSpace in readElementText.
func TestXMLDataSource_readElementText_Whitespace(t *testing.T) {
	xml := `<?xml version="1.0"?>
<Root>
  <Row>
    <Name>
      Alice
    </Name>
  </Row>
</Root>`
	ds := xmlds.New("rows")
	ds.SetXML(xml)
	if err := ds.Init(); err != nil {
		t.Fatalf("Init: %v", err)
	}
	_ = ds.First()
	name, _ := ds.GetValue("Name")
	if name != "Alice" {
		t.Errorf("Name = %q, want Alice (trimmed)", name)
	}
}

// TestXMLDataSource_readElementText_NestedSkip covers skipToChildElement skip path
// inside readElementText when a nested element is encountered.
func TestXMLDataSource_readElementText_NestedSkip(t *testing.T) {
	// Name element contains a nested element — should be skipped, returning
	// only the text content.
	xml := `<?xml version="1.0"?>
<Root>
  <Row>
    <Name>Hello <b>world</b> text</Name>
  </Row>
</Root>`
	ds := xmlds.New("rows")
	ds.SetXML(xml)
	if err := ds.Init(); err != nil {
		t.Fatalf("Init: %v", err)
	}
	_ = ds.First()
	name, _ := ds.GetValue("Name")
	// Text content before and after the nested element should be concatenated.
	if name == "" {
		t.Error("Name should not be empty")
	}
}

// TestXMLDataSource_RootPath_Slash covers path trimming of empty slash segments.
func TestXMLDataSource_RootPath_Slash(t *testing.T) {
	xml := `<?xml version="1.0"?>
<Root>
  <Orders>
    <Item Product="Apple"/>
  </Orders>
</Root>`
	ds := xmlds.New("orders")
	ds.SetXML(xml)
	// Slash prefix/suffix — empty segments are trimmed.
	ds.SetRootPath("/Orders/")
	if err := ds.Init(); err != nil {
		t.Fatalf("Init: %v", err)
	}
	if ds.RowCount() != 1 {
		t.Errorf("RowCount = %d, want 1", ds.RowCount())
	}
}
