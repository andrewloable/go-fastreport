// Package xml provides an XML data source for go-fastreport.
// It is the Go equivalent of FastReport.Data.XmlDataConnection.
package xml

import (
	"encoding/xml"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/andrewloable/go-fastreport/data"
)

// XMLDataSource is a DataSource backed by an XML file or string.
// It reads a list of repeating child elements and exposes their attributes
// and child text content as columns.
//
// Configuration:
//   - SetFilePath / SetXML — source of XML input
//   - SetRootPath — slash-separated path to the repeating element container
//     e.g. "Customers" finds <Customers> under the root element
//   - SetRowElement — local name of each row element (default: first child tag)
//
// Column values come from:
//  1. Element attributes: <Item Name="Alice" Age="30"/>
//  2. Child element text: <Item><Name>Alice</Name><Age>30</Age></Item>
type XMLDataSource struct {
	data.BaseDataSource

	sourcePath   string // file path
	sourceString string // raw XML string
	rootPath     string // slash-separated path to container element
	rowElement   string // local name of each row element (optional)
}

// New creates an XMLDataSource with the given name.
func New(name string) *XMLDataSource {
	return &XMLDataSource{
		BaseDataSource: *data.NewBaseDataSource(name),
	}
}

// SetFilePath sets the path to an XML file as the data source.
func (x *XMLDataSource) SetFilePath(path string) { x.sourcePath = path }

// FilePath returns the XML file path.
func (x *XMLDataSource) FilePath() string { return x.sourcePath }

// SetXML sets a raw XML string as the data source.
func (x *XMLDataSource) SetXML(s string) { x.sourceString = s }

// XML returns the raw XML string.
func (x *XMLDataSource) XML() string { return x.sourceString }

// SetRootPath sets a slash-separated path to the container element.
// e.g. "Customers" finds the first <Customers> child of the root.
// e.g. "Orders/Items" finds <Items> inside <Orders>.
func (x *XMLDataSource) SetRootPath(path string) { x.rootPath = path }

// RootPath returns the root path.
func (x *XMLDataSource) RootPath() string { return x.rootPath }

// SetRowElement sets the local element name used to identify row elements.
// If empty, the first child element name under the container is used.
func (x *XMLDataSource) SetRowElement(name string) { x.rowElement = name }

// RowElement returns the configured row element name.
func (x *XMLDataSource) RowElement() string { return x.rowElement }

// Init reads and parses the XML, populating the row store.
func (x *XMLDataSource) Init() error {
	raw, err := x.readSource()
	if err != nil {
		return fmt.Errorf("XMLDataSource %q: %w", x.Name(), err)
	}

	rows, cols, err := parseXML(raw, x.rootPath, x.rowElement)
	if err != nil {
		return fmt.Errorf("XMLDataSource %q: %w", x.Name(), err)
	}

	// Reset base and load columns + rows.
	x.BaseDataSource = *data.NewBaseDataSource(x.Name())
	for _, col := range cols {
		x.AddColumn(data.Column{Name: col, Alias: col, DataType: "string"})
	}
	for _, row := range rows {
		x.AddRow(row)
	}

	return x.BaseDataSource.Init()
}

// ── helpers ───────────────────────────────────────────────────────────────────

func (x *XMLDataSource) readSource() ([]byte, error) {
	if x.sourceString != "" {
		return []byte(x.sourceString), nil
	}
	if x.sourcePath != "" {
		f, err := os.Open(x.sourcePath)
		if err != nil {
			return nil, err
		}
		defer f.Close()
		return io.ReadAll(f)
	}
	return nil, fmt.Errorf("no source configured (set FilePath or XML)")
}

// parseXML reads all row elements from raw XML.
// rootPath is a slash-separated path to the container element.
// rowElem is the expected local name of each row element (empty = auto-detect).
func parseXML(raw []byte, rootPath, rowElem string) ([]map[string]any, []string, error) {
	// Build the path segments to navigate.
	var pathSegs []string
	for _, seg := range strings.Split(rootPath, "/") {
		if s := strings.TrimSpace(seg); s != "" {
			pathSegs = append(pathSegs, s)
		}
	}

	dec := xml.NewDecoder(strings.NewReader(string(raw)))

	// Skip to root element.
	if err := skipToElement(dec); err != nil {
		return nil, nil, fmt.Errorf("no root element: %w", err)
	}

	// Navigate to container element if rootPath is set.
	if len(pathSegs) > 0 {
		for _, seg := range pathSegs {
			if err := skipToChildElement(dec, seg); err != nil {
				return nil, nil, fmt.Errorf("path %q: element %q not found: %w", rootPath, seg, err)
			}
		}
	}

	// Now read child row elements.
	colSet := make(map[string]bool)
	var colOrder []string
	var rows []map[string]any

	for {
		tok, err := dec.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, nil, err
		}

		start, ok := tok.(xml.StartElement)
		if !ok {
			continue
		}

		// Auto-detect row element name from first child.
		if rowElem == "" {
			rowElem = start.Name.Local
		}
		if start.Name.Local != rowElem {
			// Skip unexpected sibling.
			if err := dec.Skip(); err != nil {
				return nil, nil, err
			}
			continue
		}

		// Build a row from this element.
		row := make(map[string]any)

		// Attributes.
		for _, attr := range start.Attr {
			col := attr.Name.Local
			row[col] = attr.Value
			if !colSet[col] {
				colSet[col] = true
				colOrder = append(colOrder, col)
			}
		}

		// Child elements — read text content as column values.
		if err := readChildren(dec, start.Name.Local, row, colSet, &colOrder); err != nil {
			return nil, nil, err
		}

		rows = append(rows, row)
	}

	return rows, colOrder, nil
}

// skipToElement advances the decoder to the next StartElement.
func skipToElement(dec *xml.Decoder) error {
	for {
		tok, err := dec.Token()
		if err != nil {
			return err
		}
		if _, ok := tok.(xml.StartElement); ok {
			return nil
		}
	}
}

// skipToChildElement advances from the current position into an element
// named localName, consuming its start token.
func skipToChildElement(dec *xml.Decoder, localName string) error {
	for {
		tok, err := dec.Token()
		if err != nil {
			return err
		}
		switch t := tok.(type) {
		case xml.StartElement:
			if t.Name.Local == localName {
				return nil
			}
			// Skip this subtree.
			if err := dec.Skip(); err != nil {
				return err
			}
		case xml.EndElement:
			return fmt.Errorf("element %q not found (parent ended)", localName)
		}
	}
}

// readChildren reads the text-content children of a row element.
// Each child element's text is stored as a column.
func readChildren(dec *xml.Decoder, parentLocal string, row map[string]any, colSet map[string]bool, colOrder *[]string) error {
	for {
		tok, err := dec.Token()
		if err != nil {
			return err
		}
		switch t := tok.(type) {
		case xml.StartElement:
			col := t.Name.Local
			text, err := readElementText(dec, col)
			if err != nil {
				return err
			}
			row[col] = text
			if !colSet[col] {
				colSet[col] = true
				*colOrder = append(*colOrder, col)
			}
		case xml.EndElement:
			if t.Name.Local == parentLocal {
				return nil
			}
		}
	}
}

// readElementText reads all character data inside an element and returns it.
// The closing end element is consumed.
func readElementText(dec *xml.Decoder, localName string) (string, error) {
	var buf strings.Builder
	for {
		tok, err := dec.Token()
		if err != nil {
			return "", err
		}
		switch t := tok.(type) {
		case xml.CharData:
			buf.Write(t)
		case xml.StartElement:
			// Nested element — skip.
			if err := dec.Skip(); err != nil {
				return "", err
			}
		case xml.EndElement:
			if t.Name.Local == localName {
				return strings.TrimSpace(buf.String()), nil
			}
		}
	}
}
