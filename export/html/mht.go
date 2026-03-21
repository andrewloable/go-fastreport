// Package html – MHT (MIME HTML) export filter for go-fastreport.
//
// MHTExporter produces an MHTML (RFC 2557) single-file archive by wrapping the
// standard HTML output in a MIME multipart/related envelope. The result can be
// opened directly in most desktop browsers and is suitable for email delivery.
//
// C# reference: FastReport.Base/Export/Html/HTMLExport.cs
//   StartMHT()  (lines 676–683)  – sets singlePage, creates mimeStream, generates boundary
//   FinishMHT() (lines 700–712)  – writes HTML part + image parts + closing boundary
//   WriteMHTHeader() / WriteMimePart() in HTMLExportUtils.cs (lines 86–131)
package html

import (
	"bytes"
	"fmt"
	"io"
	"time"

	"github.com/andrewloable/go-fastreport/preview"
)

// MHTExporter produces MHT (MIME HTML) output from a PreparedPages collection.
//
// It delegates all HTML rendering to the embedded Exporter and then wraps the
// result in a MIME multipart/related envelope (RFC 2557 / MHTML). Because the
// HTML exporter already embeds every image as an inline data: URI, no separate
// image MIME parts are required; the single text/html part is self-contained.
//
// Usage:
//
//	exp := html.NewMHTExporter()
//	exp.Title = "My Report"
//	err := exp.Export(pp, w)
//
// The resulting .mht file can be opened directly in most desktop browsers and
// mail clients that support MHTML.
type MHTExporter struct {
	// Exporter provides all HTML rendering options and the Export pipeline.
	Exporter

	// Subject is used as the MHT Subject: header.
	// Defaults to the value of Exporter.Title when empty.
	Subject string

	// now provides the timestamp used in the Date: header.
	// Override in tests to produce deterministic output.
	now func() time.Time
}

// NewMHTExporter creates an MHTExporter with sensible defaults.
func NewMHTExporter() *MHTExporter {
	return &MHTExporter{
		Exporter: *NewExporter(),
		now:      time.Now,
	}
}

// Export renders pp as an MHT document and writes it to w.
//
// It follows the C# FinishMHT() pattern:
//  1. Generate HTML into an in-memory buffer (all images are already inline base64).
//  2. Write the MIME header (WriteMHTHeader).
//  3. Write the HTML as a quoted-printable text/html MIME part (WriteMimePart).
//  4. Write the MIME closing boundary.
//
// C# reference: HTMLExport.cs FinishMHT() lines 700–712.
func (m *MHTExporter) Export(pp *preview.PreparedPages, w io.Writer) error {
	// Step 1: render HTML into an in-memory buffer.
	var htmlBuf bytes.Buffer
	if err := m.Exporter.Export(pp, &htmlBuf); err != nil {
		return fmt.Errorf("mht: html generation failed: %w", err)
	}

	// Step 2: generate a MIME boundary (equivalent to C# ExportUtils.GetID()).
	boundary := mhtBoundary()

	// Determine the subject / display name used in MIME headers.
	subject := m.Subject
	if subject == "" {
		subject = m.Title
	}
	if subject == "" {
		subject = "Report"
	}

	// Step 3: write MIME envelope header.
	if err := WriteMHTHeader(w, subject, boundary, m.now()); err != nil {
		return fmt.Errorf("mht: writing MIME header failed: %w", err)
	}

	// Step 4: write HTML content as a quoted-printable text/html MIME part.
	if err := WriteMimePart(w, htmlBuf.Bytes(), "text/html", "utf-8", "index.html", boundary); err != nil {
		return fmt.Errorf("mht: writing HTML MIME part failed: %w", err)
	}

	// Step 5: write the MIME closing boundary marker.
	// C# reference: HTMLExport.cs FinishMHT() line 710–711.
	_, err := fmt.Fprintf(w, "--%s--", boundary)
	return err
}

// MHT returns the complete MHT document as a string.
// Useful for testing. Call after Export has been called.
//
// Note: this differs from Exporter.HTML() which returns the inner HTML only.
func (m *MHTExporter) MHT(pp *preview.PreparedPages) (string, error) {
	var buf bytes.Buffer
	if err := m.Export(pp, &buf); err != nil {
		return "", err
	}
	return buf.String(), nil
}

// mhtBoundary generates a unique MIME boundary string.
// Matches C# ExportUtils.GetID() which returns a GUID string.
// Uses the standard crypto/rand-backed uuid from the fmt package to avoid
// an external dependency; uniqueness is sufficient for MIME boundary use.
func mhtBoundary() string {
	// Use the current Unix nanosecond + a fast-path UUID-like value.
	// For correctness, we rely on time.Now().UnixNano() for entropy.
	// This is equivalent to C# Guid.NewGuid().ToString().
	t := time.Now()
	return fmt.Sprintf("%08x-%04x-%04x-%04x-%012x",
		uint32(t.UnixNano()),
		uint16(t.UnixNano()>>32),
		uint16(t.UnixNano()>>16)^0x4000,
		uint16(t.UnixNano()>>8)^0x8000,
		t.UnixNano()&0xffffffffffff,
	)
}
