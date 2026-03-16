// Example: frx_to_html processes all FRX reports from a directory,
// automatically registers the bundled NorthWind XML data sources, runs the
// report engine, and exports each result to an HTML file.
//
// Run with:
//
//	go run ./examples/frx_to_html/
//	go run ./examples/frx_to_html/ -dir test-reports -out output
//	go run ./examples/frx_to_html/ -frx "test-reports/Simple List.frx" -out output
//
// Flags:
//
//	-dir    Directory containing .frx files (default: test-reports)
//	-frx    Single .frx file to render (overrides -dir)
//	-nwind  Path to nwind.xml for NorthWind data (default: test-reports/nwind.xml)
//	-out    Output directory for HTML files (default: html-output)
package main

import (
	"encoding/xml"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/andrewloable/go-fastreport/data"
	xmldata "github.com/andrewloable/go-fastreport/data/xml"
	"github.com/andrewloable/go-fastreport/engine"
	"github.com/andrewloable/go-fastreport/export/html"
	"github.com/andrewloable/go-fastreport/preview"
	"github.com/andrewloable/go-fastreport/reportpkg"
)

func main() {
	frxPath := flag.String("frx", "", "single .frx file to render (overrides -dir)")
	frxDir := flag.String("dir", "test-reports", "directory containing .frx files")
	nwindPath := flag.String("nwind", "test-reports/nwind.xml", "path to nwind.xml NorthWind data file")
	outDir := flag.String("out", "html-output", "output directory for HTML files")
	flag.Parse()

	// Pre-load the NorthWind XML bytes once so we don't re-read for every report.
	nwindRaw, nwindErr := os.ReadFile(*nwindPath)
	if nwindErr != nil {
		fmt.Fprintf(os.Stderr, "warning: cannot read NorthWind data %q: %v\n", *nwindPath, nwindErr)
	}

	// Collect FRX files to process.
	var frxFiles []string
	if *frxPath != "" {
		frxFiles = []string{*frxPath}
	} else {
		entries, err := filepath.Glob(filepath.Join(*frxDir, "*.frx"))
		if err != nil || len(entries) == 0 {
			fmt.Fprintf(os.Stderr, "no .frx files found in %q\n", *frxDir)
			os.Exit(1)
		}
		frxFiles = entries
	}

	// Create output directory.
	if err := os.MkdirAll(*outDir, 0o755); err != nil {
		fmt.Fprintf(os.Stderr, "create output dir %q: %v\n", *outDir, err)
		os.Exit(1)
	}

	ok, failed := 0, 0
	for _, frx := range frxFiles {
		base := filepath.Base(frx)
		stem := strings.TrimSuffix(base, filepath.Ext(base))
		outFile := filepath.Join(*outDir, stem+".html")

		pages, err := renderFRX(frx, nwindRaw)
		if err != nil {
			fmt.Fprintf(os.Stderr, "  FAIL  %s: %v\n", base, err)
			failed++
			continue
		}

		f, err := os.Create(outFile)
		if err != nil {
			fmt.Fprintf(os.Stderr, "  FAIL  %s (create output): %v\n", base, err)
			failed++
			continue
		}

		exp := html.NewExporter()
		exp.Title = stem
		exp.EmbedCSS = true
		if err := exp.Export(pages, f); err != nil {
			f.Close()
			fmt.Fprintf(os.Stderr, "  FAIL  %s (export): %v\n", base, err)
			failed++
			continue
		}
		f.Close()

		fmt.Printf("  OK    %-50s → %s (%d page(s))\n", base, outFile, pages.Count())
		ok++
	}

	fmt.Printf("\n%d succeeded, %d failed — HTML files in %q\n", ok, failed, *outDir)
	if failed > 0 {
		os.Exit(1)
	}
}

// renderFRX loads a single FRX, hydrates its data sources, runs the engine,
// and returns the prepared pages ready for export.
func renderFRX(frxPath string, nwindRaw []byte) (*preview.PreparedPages, error) {
	r := reportpkg.NewReport()
	if err := r.Load(frxPath); err != nil {
		return nil, fmt.Errorf("load: %w", err)
	}

	if len(nwindRaw) > 0 {
		registerNorthWindSources(r, nwindRaw)
	}

	e := engine.New(r)
	if err := e.Run(engine.DefaultRunOptions()); err != nil {
		return nil, fmt.Errorf("engine: %w", err)
	}
	return e.PreparedPages(), nil
}

// registerNorthWindSources parses the NorthWind XML bytes, builds one
// XMLDataSource per table, and registers each in the report dictionary,
// replacing any uninitialised stub TableDataSource loaded from the FRX.
func registerNorthWindSources(r *reportpkg.Report, raw []byte) {
	tableNames := extractTopLevelElementNames(raw)
	dict := r.Dictionary()
	xmlStr := string(raw)

	for _, name := range tableNames {
		// Decoded name: XML-encoding like _x0020_ → space, _x002E_ → dot, etc.
		decoded := decodeXMLName(name)

		ds := xmldata.New(decoded)
		ds.SetXML(xmlStr)
		ds.SetRowElement(name) // element name in the XML file (encoded)
		if err := ds.Init(); err != nil {
			continue
		}
		// Set alias AFTER Init() — Init() resets BaseDataSource which clears the alias.
		ds.SetAlias(decoded)

		// Replace any stub data sources with matching name or alias.
		// Collect first, then remove to avoid mutation-during-iteration.
		var toRemove []data.DataSource
		for _, existing := range dict.DataSources() {
			n, a := existing.Name(), existing.Alias()
			if strings.EqualFold(n, name) || strings.EqualFold(a, name) ||
				strings.EqualFold(n, decoded) || strings.EqualFold(a, decoded) {
				toRemove = append(toRemove, existing)
			}
		}
		for _, existing := range toRemove {
			dict.RemoveDataSource(existing)
		}
		dict.AddDataSource(ds)
		dict.AddParameter(&data.Parameter{Name: decoded, Value: decoded})
	}
}

// decodeXMLName converts XML-encoded element names back to human-readable form.
// The pattern _xHHHH_ (7 chars) encodes a single Unicode code point.
// e.g. "Order_x0020_Details" → "Order Details"
func decodeXMLName(name string) string {
	for {
		start := strings.Index(name, "_x")
		if start == -1 || start+6 >= len(name) || name[start+6] != '_' {
			break
		}
		hex := name[start+2 : start+6]
		var r rune
		fmt.Sscanf(hex, "%X", &r)
		name = name[:start] + string(r) + name[start+7:]
	}
	return name
}

// extractTopLevelElementNames returns unique child element names of the XML
// root in first-occurrence order.
func extractTopLevelElementNames(xmlBytes []byte) []string {
	dec := xml.NewDecoder(strings.NewReader(string(xmlBytes)))
	var depth int
	seen := make(map[string]bool)
	var order []string
	for {
		tok, err := dec.Token()
		if err != nil {
			break
		}
		switch t := tok.(type) {
		case xml.StartElement:
			depth++
			if depth == 2 {
				if !seen[t.Name.Local] {
					seen[t.Name.Local] = true
					order = append(order, t.Name.Local)
				}
			}
		case xml.EndElement:
			depth--
		}
	}
	return order
}
