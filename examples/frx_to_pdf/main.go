// Example: frx_to_pdf processes all FRX reports from a directory,
// automatically registers the bundled NorthWind XML data sources, runs the
// report engine, and exports each result to a PDF file.
//
// Run with:
//
//	go run ./examples/frx_to_pdf/
//	go run ./examples/frx_to_pdf/ -dir test-reports -out pdf-output
//	go run ./examples/frx_to_pdf/ -frx "test-reports/Simple List.frx" -out pdf-output
//
// Flags:
//
//	-dir    Directory containing .frx files (default: test-reports)
//	-frx    Single .frx file to render (overrides -dir)
//	-nwind  Path to nwind.xml for NorthWind data (default: test-reports/nwind.xml)
//	-out    Output directory for PDF files (default: pdf-output)
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
	"github.com/andrewloable/go-fastreport/export/pdf"
	"github.com/andrewloable/go-fastreport/preview"
	"github.com/andrewloable/go-fastreport/reportpkg"
)

func main() {
	frxPath := flag.String("frx", "", "single .frx file to render (overrides -dir)")
	frxDir := flag.String("dir", "test-reports", "directory containing .frx files")
	nwindPath := flag.String("nwind", "test-reports/nwind.xml", "path to nwind.xml NorthWind data file")
	outDir := flag.String("out", "pdf-output", "output directory for PDF files")
	flag.Parse()

	nwindRaw, nwindErr := os.ReadFile(*nwindPath)
	if nwindErr != nil {
		fmt.Fprintf(os.Stderr, "warning: cannot read NorthWind data %q: %v\n", *nwindPath, nwindErr)
	}

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

	if err := os.MkdirAll(*outDir, 0o755); err != nil {
		fmt.Fprintf(os.Stderr, "create output dir %q: %v\n", *outDir, err)
		os.Exit(1)
	}

	ok, failed := 0, 0
	for _, frx := range frxFiles {
		base := filepath.Base(frx)
		stem := strings.TrimSuffix(base, filepath.Ext(base))
		outFile := filepath.Join(*outDir, stem+".pdf")

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

		exp := pdf.NewExporter()
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

	fmt.Printf("\n%d succeeded, %d failed — PDF files in %q\n", ok, failed, *outDir)
	if failed > 0 {
		os.Exit(1)
	}
}

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

func registerNorthWindSources(r *reportpkg.Report, raw []byte) {
	tableNames := extractTopLevelElementNames(raw)
	dict := r.Dictionary()
	xmlStr := string(raw)

	for _, name := range tableNames {
		decoded := decodeXMLName(name)

		ds := xmldata.New(decoded)
		ds.SetXML(xmlStr)
		ds.SetRowElement(name)
		if err := ds.Init(); err != nil {
			continue
		}
		ds.SetAlias(decoded)

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
			if depth == 2 && !seen[t.Name.Local] {
				seen[t.Name.Local] = true
				order = append(order, t.Name.Local)
			}
		case xml.EndElement:
			depth--
		}
	}
	return order
}
