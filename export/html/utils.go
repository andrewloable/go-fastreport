package html

// utils.go — HTML-specific utility types and functions ported from
// C# FastReport.Export.Html.HTMLExportUtils.cs.
//
// C# source: FastReport.Base/Export/Html/HTMLExportUtils.cs
//
// Ported:
//   - HTMLExportFormat enum  → HTMLExportFormat / const block
//   - ImageFormat enum       → HTMLImageFormat / const block
//   - HtmlSizeUnits enum     → HtmlSizeUnits / const block
//   - Px()                   → px() helper (appends "px;" suffix, internal)
//   - SizeValue()            → SizeValue()
//   - WriteMimePart()        → WriteMimePart()
//   - WriteMHTHeader()       → WriteMHTHeader()
//   - HTMLPageData class     → HTMLPageData struct

import (
	"encoding/base64"
	"fmt"
	"io"
	"math"
	"strings"
	"time"

	"github.com/andrewloable/go-fastreport/export"
)

// ── Enums ─────────────────────────────────────────────────────────────────────

// ExportMode controls how the HTML exporter produces output files.
// Matches C# FastReport.Export.Html.HTMLExport.ExportType.
//
// C# reference: HTMLExport.cs lines 86-102 (ExportType enum).
type ExportMode int

const (
	// ExportModeSingleFile writes all pages into a single HTML document (default).
	// Matches C# ExportType.Export with SinglePage=true.
	ExportModeSingleFile ExportMode = iota
	// ExportModeMultiPage writes each report page as a separate HTML file
	// (pageN.html) in the output directory.
	// Matches C# ExportType.Export with SinglePage=false, Navigator=false.
	ExportModeMultiPage
	// ExportModeNavigator writes separate page files plus a JavaScript-based
	// frame navigator (index.html + nav.html) so users can browse pages in a
	// browser without a server.
	// Matches C# ExportType.Export with SinglePage=false, Navigator=true.
	ExportModeNavigator
)

// HTMLExportFormat specifies the output file format for HTML export.
// Matches C# FastReport.Export.Html.HTMLExportFormat.
type HTMLExportFormat int

const (
	// HTMLExportFormatMessageHTML is the MIME-encapsulated (MHT/MHTML) format.
	// C# enum value: MessageHTML = 0.
	HTMLExportFormatMessageHTML HTMLExportFormat = iota
	// HTMLExportFormatHTML is the plain HTML format.
	// C# enum value: HTML = 1.
	HTMLExportFormatHTML
)

// HTMLImageFormat specifies the image encoding used when exporting pictures in HTML.
// Matches C# FastReport.Export.Html.ImageFormat.
type HTMLImageFormat int

const (
	// HTMLImageFormatBmp specifies BMP encoding.
	// C# enum value: Bmp = 0.
	HTMLImageFormatBmp HTMLImageFormat = iota
	// HTMLImageFormatPng specifies PNG encoding (recommended).
	// C# enum value: Png = 1.
	HTMLImageFormatPng
	// HTMLImageFormatJpeg specifies JPEG encoding.
	// C# enum value: Jpeg = 2.
	HTMLImageFormatJpeg
	// HTMLImageFormatGif specifies GIF encoding.
	// C# enum value: Gif = 3.
	HTMLImageFormatGif
)

// HtmlSizeUnits specifies the CSS units used for HTML dimensions.
// Matches C# FastReport.Export.Html.HtmlSizeUnits.
type HtmlSizeUnits int

const (
	// HtmlSizeUnitsPixel specifies pixel (px) units.
	// C# enum value: Pixel = 0.
	HtmlSizeUnitsPixel HtmlSizeUnits = iota
	// HtmlSizeUnitsPercent specifies percentage (%) units.
	// C# enum value: Percent = 1.
	HtmlSizeUnitsPercent
)

// ── px helper ─────────────────────────────────────────────────────────────────

// px formats pixel as a CSS value string with a trailing "px;" suffix,
// matching C# HTMLExport.Px(double pixel) which calls ExportUtils.FloatToString.
// For example: px(28.35) → "28.35px;"
//
// C# source: HTMLExportUtils.cs line 69-72.
func px(pixel float64) string {
	s := export.FloatToString(pixel, 2)
	return s + "px;"
}

// ── SizeValue ─────────────────────────────────────────────────────────────────

// SizeValue formats value in the given units relative to maxvalue.
// When units == HtmlSizeUnitsPixel it returns "NNpx;".
// When units == HtmlSizeUnitsPercent it returns "NN%" (integer percentage).
// Otherwise it returns the plain float string.
//
// Matches C# HTMLExport.SizeValue (HTMLExportUtils.cs lines 74-84).
func SizeValue(value, maxvalue float64, units HtmlSizeUnits) string {
	switch units {
	case HtmlSizeUnitsPixel:
		return px(value)
	case HtmlSizeUnitsPercent:
		return fmt.Sprintf("%d%%", int(math.Round(value*100/maxvalue)))
	default:
		return export.FloatToString(value, 2)
	}
}

// ── WriteMHTHeader ─────────────────────────────────────────────────────────────

// WriteMHTHeader writes the MIME multipart header required for MHT (MHTML) files.
// It encodes fileName in UTF-8 Base64 (RFC 2047 encoded-word syntax) and writes
// RFC 2822 headers to w.
//
// Matches C# HTMLExport.WriteMHTHeader (HTMLExportUtils.cs lines 117-131).
//
// Parameters:
//   - w         – output writer (e.g. the MHT file stream)
//   - fileName  – display name of the report (used in From: and Subject: headers)
//   - boundary  – MIME boundary string (must not appear in any part body)
//   - now       – timestamp used in the Date: header (pass time.Now() for live use)
func WriteMHTHeader(w io.Writer, fileName, boundary string, now time.Time) error {
	// C# encodes the filename as UTF-8 Base64 encoded-word (RFC 2047).
	encoded := base64.StdEncoding.EncodeToString([]byte(fileName))
	s := "=?utf-8?B?" + encoded + "?="

	var sb strings.Builder
	sb.WriteString("From: ")
	sb.WriteString(s)
	sb.WriteString("\r\n")
	sb.WriteString("Subject: ")
	sb.WriteString(s)
	sb.WriteString("\r\n")
	sb.WriteString("Date: ")
	sb.WriteString(export.GetRFCDate(now))
	sb.WriteString("\r\n")
	sb.WriteString("MIME-Version: 1.0\r\n")
	sb.WriteString("Content-Type: multipart/related; type=\"text/html\"; boundary=\"")
	sb.WriteString(boundary)
	sb.WriteString("\"\r\n")
	sb.WriteString("\r\n")
	sb.WriteString("This is a multi-part message in MIME format.\r\n")
	sb.WriteString("\r\n")

	_, err := io.WriteString(w, sb.String())
	return err
}

// ── WriteMimePart ─────────────────────────────────────────────────────────────

// WriteMimePart writes one MIME body part to w.
// For text/html parts the body is Quoted-Printable encoded; all other MIME types
// use Base64 encoding with line breaks every 76 characters (RFC 2045).
//
// Matches C# HTMLExport.WriteMimePart (HTMLExportUtils.cs lines 86-115).
//
// Parameters:
//   - w         – output writer (e.g. the MHT file stream)
//   - data      – raw bytes of the part body
//   - mimetype  – MIME content-type (e.g. "text/html", "image/png")
//   - charset   – charset value appended to Content-Type (empty → omitted)
//   - filename  – Content-Location URL for the part
//   - boundary  – MIME boundary string matching the one in WriteMHTHeader
func WriteMimePart(w io.Writer, data []byte, mimetype, charset, filename, boundary string) error {
	var sb strings.Builder

	// Part boundary marker.
	sb.WriteString("--")
	sb.WriteString(boundary)
	sb.WriteString("\r\n")

	// Content-Type header.
	sb.WriteString("Content-Type: ")
	sb.WriteString(mimetype)
	sb.WriteString(";")
	if charset != "" {
		sb.WriteString(" charset=\"")
		sb.WriteString(charset)
		sb.WriteString("\"\r\n")
	} else {
		sb.WriteString("\r\n")
	}

	// Encoding header and body.
	var body string
	if mimetype == "text/html" {
		sb.WriteString("Content-Transfer-Encoding: quoted-printable\r\n")
		body = export.QuotedPrintable(data)
	} else {
		sb.WriteString("Content-Transfer-Encoding: base64\r\n")
		// Base64 with line breaks every 76 characters (matching C# InsertLineBreaks).
		raw := base64.StdEncoding.EncodeToString(data)
		body = insertBase64LineBreaks(raw)
	}

	// Content-Location header.
	sb.WriteString("Content-Location: ")
	sb.WriteString(export.HtmlURL(filename))
	sb.WriteString("\r\n")

	// Blank line separates headers from body.
	sb.WriteString("\r\n")
	sb.WriteString(body)
	sb.WriteString("\r\n")
	sb.WriteString("\r\n")

	_, err := io.WriteString(w, sb.String())
	return err
}

// insertBase64LineBreaks inserts "\r\n" every 76 characters in a Base64-encoded
// string, matching .NET's Base64FormattingOptions.InsertLineBreaks behaviour.
func insertBase64LineBreaks(s string) string {
	const lineLen = 76
	if len(s) <= lineLen {
		return s
	}
	var sb strings.Builder
	for i := 0; i < len(s); i += lineLen {
		end := i + lineLen
		if end > len(s) {
			end = len(s)
		}
		sb.WriteString(s[i:end])
		if end < len(s) {
			sb.WriteString("\r\n")
		}
	}
	return sb.String()
}

// ── HTMLPageData ──────────────────────────────────────────────────────────────

// HTMLPageData holds the per-page rendered output buffers used during HTML export.
// It is the Go equivalent of C# FastReport.Export.Html.HTMLPageData (inner class).
//
// In the Go exporter, rendering is single-threaded and these fields are managed
// directly on the Exporter struct; HTMLPageData is provided as a public type for
// callers that want to assemble multi-page output programmatically.
//
// C# source: HTMLExportUtils.cs lines 136-228.
type HTMLPageData struct {
	// Width is the page width in pixels.
	Width float32
	// Height is the page height in pixels.
	Height float32
	// PageNumber is the 1-based page number.
	PageNumber int
	// CSSText holds the CSS style block for this page (rendered by the exporter).
	CSSText strings.Builder
	// PageText holds the HTML content div for this page.
	PageText strings.Builder
	// Pictures holds raw image data for any embedded pictures on this page.
	Pictures [][]byte
	// Guids holds the hash/identifier strings for each picture in Pictures.
	Guids []string
}

// NewHTMLPageData creates an empty HTMLPageData for the given page number.
// C# equivalent: new HTMLPageData() with pageNumber assigned separately.
func NewHTMLPageData(pageNumber int) *HTMLPageData {
	return &HTMLPageData{
		PageNumber: pageNumber,
	}
}

// AddPicture appends an image blob and its identifier to the page data.
// This is the Go equivalent of C# pages[n].Pictures.Add(stream) / Guids.Add(hash).
func (d *HTMLPageData) AddPicture(data []byte, guid string) {
	d.Pictures = append(d.Pictures, data)
	d.Guids = append(d.Guids, guid)
}
