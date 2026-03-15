package utils

import (
	"encoding/xml"
	"fmt"
	"io"
	"os"
	"strings"
	"sync"
)

// ── Res — localization framework ──────────────────────────────────────────────
//
// Res provides string lookup from locale files. It is the Go equivalent of
// FastReport.Utils.Res.
//
// Usage:
//
//	// At startup (optional — defaults to English built-in strings):
//	utils.LoadLocale("/path/to/en.frl")
//
//	// Look up a string by comma-separated key:
//	label := utils.ResGet("Messages,SaveChanges")
//
// Key format: "Category,SubCategory,Item" — maps to a nested XML element path.
//
// .frl files are XML files with the structure:
//
//	<Locale Name="en">
//	  <Category Name="Messages">
//	    <Item Name="SaveChanges" Text="Save changes?" />
//	  </Category>
//	</Locale>

// localeNode is a node in the parsed locale XML tree.
type localeNode struct {
	Name     string
	Text     string
	Children []*localeNode
}

// localeState holds the active locale and the built-in English fallback.
var localeState struct {
	mu      sync.RWMutex
	current *localeNode
	builtin *localeNode
}

func init() {
	// Initialise with the built-in English strings.
	root := builtinEnglish()
	localeState.current = root
	localeState.builtin = root
}

// ResGet returns the locale string for the given comma-separated key path.
// Falls back to the built-in English locale, then returns "<key> NOT LOCALIZED!".
func ResGet(key string) string {
	localeState.mu.RLock()
	cur := localeState.current
	blt := localeState.builtin
	localeState.mu.RUnlock()

	if v := lookupLocale(cur, key); v != "" {
		return v
	}
	if cur != blt {
		if v := lookupLocale(blt, key); v != "" {
			return v
		}
	}
	return key + " NOT LOCALIZED!"
}

// ResSet overrides a locale string for the given comma-separated key path.
// The value is applied to the current locale. If intermediate nodes are missing
// they are created.
func ResSet(key, value string) {
	localeState.mu.Lock()
	defer localeState.mu.Unlock()
	setLocale(localeState.current, key, value)
}

// LocaleName returns the name attribute of the current locale root (e.g. "en").
func LocaleName() string {
	localeState.mu.RLock()
	defer localeState.mu.RUnlock()
	if localeState.current != nil {
		return localeState.current.Name
	}
	return "en"
}

// LoadLocale loads a .frl (XML) locale file and sets it as the active locale.
// Falls back to the built-in English locale if the file cannot be read.
func LoadLocale(filename string) error {
	f, err := os.Open(filename)
	if err != nil {
		return fmt.Errorf("locale: open %q: %w", filename, err)
	}
	defer f.Close()
	return LoadLocaleReader(f)
}

// LoadLocaleReader parses a locale XML from r and sets it as the active locale.
func LoadLocaleReader(r io.Reader) error {
	root, err := parseLocaleXML(r)
	if err != nil {
		return err
	}
	localeState.mu.Lock()
	localeState.current = root
	localeState.mu.Unlock()
	return nil
}

// ResetToBuiltin resets the active locale to the built-in English strings.
func ResetToBuiltin() {
	localeState.mu.Lock()
	localeState.current = localeState.builtin
	localeState.mu.Unlock()
}

// ── Internal helpers ──────────────────────────────────────────────────────────

// lookupLocale traverses the node tree for a comma-separated key path and
// returns the Text attribute of the terminal node.
func lookupLocale(root *localeNode, key string) string {
	if root == nil {
		return ""
	}
	parts := strings.Split(key, ",")
	node := root
	for _, part := range parts {
		part = strings.TrimSpace(part)
		child := findChild(node, part)
		if child == nil {
			return ""
		}
		node = child
	}
	return node.Text
}

// setLocale sets the Text for the node at the given key path, creating nodes
// as needed.
func setLocale(root *localeNode, key, value string) {
	if root == nil {
		return
	}
	parts := strings.Split(key, ",")
	node := root
	for _, part := range parts {
		part = strings.TrimSpace(part)
		child := findChild(node, part)
		if child == nil {
			child = &localeNode{Name: part}
			node.Children = append(node.Children, child)
		}
		node = child
	}
	node.Text = value
}

// findChild returns the first child of node whose Name matches name (case-insensitive).
func findChild(node *localeNode, name string) *localeNode {
	for _, c := range node.Children {
		if strings.EqualFold(c.Name, name) {
			return c
		}
	}
	return nil
}

// ── XML parser ────────────────────────────────────────────────────────────────

// frlXMLLocale is the root XML element of a .frl locale file.
type frlXMLLocale struct {
	XMLName  xml.Name      `xml:"Locale"`
	Name     string        `xml:"Name,attr"`
	Children []frlXMLItem  `xml:",any"`
}

// frlXMLItem is a generic XML element with a Name attribute and optional Text.
type frlXMLItem struct {
	XMLName  xml.Name
	Name     string       `xml:"Name,attr"`
	Text     string       `xml:"Text,attr"`
	Children []frlXMLItem `xml:",any"`
}

// parseLocaleXML reads an XML locale file into a localeNode tree.
func parseLocaleXML(r io.Reader) (*localeNode, error) {
	var loc frlXMLLocale
	if err := xml.NewDecoder(r).Decode(&loc); err != nil {
		return nil, fmt.Errorf("locale: parse XML: %w", err)
	}
	root := &localeNode{Name: loc.Name}
	for i := range loc.Children {
		root.Children = append(root.Children, xmlToNode(&loc.Children[i]))
	}
	return root, nil
}

func xmlToNode(x *frlXMLItem) *localeNode {
	n := &localeNode{
		Name: x.Name,
		Text: x.Text,
	}
	// Use XMLName.Local as fallback when Name attr is absent.
	if n.Name == "" {
		n.Name = x.XMLName.Local
	}
	for i := range x.Children {
		n.Children = append(n.Children, xmlToNode(&x.Children[i]))
	}
	return n
}

// ── Built-in English strings ──────────────────────────────────────────────────
//
// A minimal set of English strings is embedded to provide useful defaults
// without requiring an external .frl file. This covers the most common
// UI strings referenced by the report engine and objects.

func builtinEnglish() *localeNode {
	root := &localeNode{Name: "en"}

	add := func(key, text string) {
		setLocale(root, key, text)
	}

	// ── Report objects ──────────────────────────────────────────────────────
	add("Objects,Text", "Text")
	add("Objects,Picture", "Picture")
	add("Objects,Line", "Line")
	add("Objects,Shape", "Shape")
	add("Objects,RichText", "Rich Text")
	add("Objects,Table", "Table")
	add("Objects,Matrix", "Matrix")
	add("Objects,Chart", "Chart")
	add("Objects,Barcode", "Barcode")
	add("Objects,ZipCode", "Zip Code")
	add("Objects,Gauge", "Gauge")
	add("Objects,Map", "Map")
	add("Objects,CrossView", "Cross View")
	add("Objects,CheckBox", "Check Box")
	add("Objects,SubReport", "SubReport")

	// ── Band names ──────────────────────────────────────────────────────────
	add("Bands,ReportTitle", "Report Title")
	add("Bands,ReportSummary", "Report Summary")
	add("Bands,PageHeader", "Page Header")
	add("Bands,PageFooter", "Page Footer")
	add("Bands,ColumnHeader", "Column Header")
	add("Bands,ColumnFooter", "Column Footer")
	add("Bands,DataHeader", "Data Header")
	add("Bands,Data", "Data")
	add("Bands,DataFooter", "Data Footer")
	add("Bands,GroupHeader", "Group Header")
	add("Bands,GroupFooter", "Group Footer")
	add("Bands,Child", "Child")
	add("Bands,Overlay", "Overlay")

	// ── Messages ────────────────────────────────────────────────────────────
	add("Messages,SaveChanges", "Save changes?")
	add("Messages,FileNotFound", "File not found")
	add("Messages,Error", "Error")
	add("Messages,Warning", "Warning")
	add("Messages,Information", "Information")
	add("Messages,Yes", "Yes")
	add("Messages,No", "No")
	add("Messages,Cancel", "Cancel")
	add("Messages,OK", "OK")

	// ── Export ──────────────────────────────────────────────────────────────
	add("Export,PDF,Name", "PDF Document")
	add("Export,HTML,Name", "HTML Document")
	add("Export,Image,Name", "Image File")
	add("Export,CSV,Name", "CSV File")

	// ── Data ────────────────────────────────────────────────────────────────
	add("Data,Parameter", "Parameter")
	add("Data,Total", "Total")
	add("Data,Variable", "Variable")
	add("Data,Connection", "Connection")
	add("Data,DataSource", "Data Source")

	return root
}
