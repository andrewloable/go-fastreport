package reportpkg

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/andrewloable/go-fastreport/report"
	"github.com/andrewloable/go-fastreport/serial"
)

// ── Load ──────────────────────────────────────────────────────────────────────

// Load reads a FRX report file from filename and populates this Report.
func (r *Report) Load(filename string) error {
	f, err := os.Open(filename)
	if err != nil {
		return fmt.Errorf("report.Load: %w", err)
	}
	defer f.Close()
	return r.LoadFrom(f)
}

// LoadFromString deserializes a Report from an FRX XML string.
func (r *Report) LoadFromString(xml string) error {
	return r.LoadFrom(strings.NewReader(xml))
}

// LoadFrom deserializes a Report from an io.Reader containing FRX XML.
func (r *Report) LoadFrom(rd io.Reader) error {
	rdr := serial.NewReader(rd)

	// Expect the root <Report> element.
	typeName, ok := rdr.ReadObjectHeader()
	if !ok {
		return fmt.Errorf("report.LoadFrom: empty or invalid FRX document")
	}
	if typeName != "Report" {
		return fmt.Errorf("report.LoadFrom: expected root element 'Report', got %q", typeName)
	}

	// Deserialize top-level Report properties (Name, Author, etc.)
	if err := r.Deserialize(rdr); err != nil {
		return fmt.Errorf("report.LoadFrom: deserialize root: %w", err)
	}

	// Iterate child elements (pages, styles, dictionary, etc.)
	for {
		childType, ok := rdr.NextChild()
		if !ok {
			break
		}
		switch childType {
		case "ReportPage":
			pg, err := deserializePage(rdr)
			if err != nil {
				_ = rdr.FinishChild()
				return fmt.Errorf("report.LoadFrom: deserialize page: %w", err)
			}
			r.AddPage(pg)
		default:
			// Unknown top-level child — skip.
			_ = rdr.SkipElement()
		}
		_ = rdr.FinishChild()
	}
	return nil
}

// deserializePage deserializes a ReportPage from the current reader position.
func deserializePage(rdr *serial.Reader) (*ReportPage, error) {
	pg := NewReportPage()
	if err := pg.Deserialize(rdr); err != nil {
		return nil, err
	}

	// Iterate bands and objects inside the page.
	for {
		childType, ok := rdr.NextChild()
		if !ok {
			break
		}
		obj, err := serial.DefaultRegistry.Create(childType)
		if err != nil {
			// Unknown type — skip.
			_ = rdr.SkipElement()
			_ = rdr.FinishChild()
			continue
		}
		if err2 := obj.Deserialize(rdr); err2 != nil {
			_ = rdr.FinishChild()
			return nil, fmt.Errorf("deserialize %s: %w", childType, err2)
		}
		// Walk children of this band/object too.
		deserializeChildren(rdr, obj)
		_ = rdr.FinishChild()

		// Attach band to page using the FRX type name.
		pg.AddBandByTypeName(childType, obj)
	}
	return pg, nil
}

// deserializeChildren recursively reads child elements of obj using the registry.
func deserializeChildren(rdr *serial.Reader, parent report.Base) {
	type hasObjects interface {
		Objects() *report.ObjectCollection
	}
	container, isContainer := parent.(hasObjects)

	for {
		childType, ok := rdr.NextChild()
		if !ok {
			break
		}
		child, err := serial.DefaultRegistry.Create(childType)
		if err != nil {
			_ = rdr.SkipElement()
			_ = rdr.FinishChild()
			continue
		}
		_ = child.Deserialize(rdr)
		deserializeChildren(rdr, child)
		_ = rdr.FinishChild()

		if isContainer {
			container.Objects().Add(child)
		}
	}
}


// ── Save ──────────────────────────────────────────────────────────────────────

// Save serializes the Report to a FRX file at filename.
func (r *Report) Save(filename string) error {
	f, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("report.Save: %w", err)
	}
	defer f.Close()
	return r.SaveTo(f)
}

// SaveToString serializes the Report to an FRX XML string.
func (r *Report) SaveToString() (string, error) {
	var buf bytes.Buffer
	if err := r.SaveTo(&buf); err != nil {
		return "", err
	}
	return buf.String(), nil
}

// SaveTo serializes the Report to an io.Writer as FRX XML.
// Pages and their bands/objects are written as nested children via Report.Serialize().
func (r *Report) SaveTo(w io.Writer) error {
	sw := serial.NewWriter(w)
	if err := sw.WriteHeader(); err != nil {
		return fmt.Errorf("report.SaveTo: write header: %w", err)
	}
	if err := sw.WriteObjectNamed("Report", r); err != nil {
		return fmt.Errorf("report.SaveTo: write report: %w", err)
	}
	return sw.Flush()
}
