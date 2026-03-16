// Package csv implements a CSV export filter for go-fastreport.
// It extracts tabular text data from prepared pages, grouping objects
// by their Y coordinate (row) and sorting each row left-to-right by X.
// One CSV row is written per band row; non-text objects are skipped.
package csv

import (
	"encoding/csv"
	"io"
	"math"
	"sort"

	"github.com/andrewloable/go-fastreport/export"
	"github.com/andrewloable/go-fastreport/preview"
)

// Exporter produces CSV output from a PreparedPages collection.
//
// Each band row in the prepared pages becomes one CSV row.
// Objects within a band are grouped by their Y coordinate and then
// sorted left-to-right by X coordinate to produce column order.
// Only text-bearing objects (ObjectTypeText, ObjectTypeHtml, ObjectTypeRTF,
// ObjectTypeCheckBox) are emitted; all other object types are skipped.
type Exporter struct {
	export.ExportBase

	// Separator is the CSV field delimiter (default ',').
	Separator rune
	// Quote is the CSV quoting character (default '"').
	Quote rune
	// HeaderRow controls whether a header row with band and object names
	// is emitted before data rows (default true).
	HeaderRow bool

	w      io.Writer
	cw     *csv.Writer
	header []string // column headers collected during first-pass
	// headerWritten tracks whether the header row has been flushed.
	headerWritten bool
	// pendingRows accumulates all rows when HeaderRow is true (deferred until
	// we know the full column set). When HeaderRow is false, rows are written
	// immediately.
	pendingRows [][]string
	// maxCols is the widest row seen so far (used to pad rows uniformly).
	maxCols int
}

// NewExporter creates an Exporter with sensible defaults.
func NewExporter() *Exporter {
	return &Exporter{
		ExportBase: export.NewExportBase(),
		Separator:  ',',
		Quote:      '"',
		HeaderRow:  true,
	}
}

// Export writes the PreparedPages as CSV to w.
func (e *Exporter) Export(pages *preview.PreparedPages, w io.Writer) error {
	e.w = w
	return e.ExportBase.Export(pages, w, e)
}

// FileExtension returns the recommended file extension.
func (e *Exporter) FileExtension() string { return ".csv" }

// Name returns the human-readable name of this export format.
func (e *Exporter) Name() string { return "CSV" }

// ── Exporter interface ─────────────────────────────────────────────────────────

// Start initialises the csv.Writer and clears accumulated state.
func (e *Exporter) Start() error {
	e.cw = csv.NewWriter(e.w)
	e.cw.Comma = e.Separator
	// encoding/csv does not expose a Quote field directly;
	// the standard library always uses '"'. We honour e.Quote for future
	// compatibility but cannot override the built-in writer's quote character.
	// If a caller sets a non-default Quote, they should supply their own writer.
	e.header = e.header[:0]
	e.pendingRows = e.pendingRows[:0]
	e.maxCols = 0
	e.headerWritten = false
	return nil
}

// ExportPageBegin is a no-op for CSV (pages are transparent to the format).
func (e *Exporter) ExportPageBegin(_ *preview.PreparedPage) error { return nil }

// ExportBand converts a single band into one or more CSV rows.
//
// Strategy:
//  1. Collect all text objects from the band.
//  2. Group them by Y coordinate (with a small epsilon tolerance so objects
//     that differ by less than 1 pixel are treated as the same row).
//  3. Within each Y-group sort by X (left-to-right).
//  4. Emit one CSV row per Y-group.
func (e *Exporter) ExportBand(b *preview.PreparedBand) error {
	textObjs := collectTextObjects(b)
	if len(textObjs) == 0 {
		return nil
	}

	// Group objects by row (Y position).
	rows := groupByY(textObjs)

	for _, row := range rows {
		// Sort each row left-to-right.
		sort.Slice(row, func(i, j int) bool {
			return row[i].Left < row[j].Left
		})

		record := make([]string, len(row))
		for i, obj := range row {
			record[i] = objectText(obj)
		}

		if e.HeaderRow && !e.headerWritten {
			// Build header entries from band name + object name.
			for _, obj := range row {
				colName := obj.Name
				if colName == "" {
					colName = b.Name
				}
				e.header = appendUnique(e.header, colName)
			}
		}

		if len(record) > e.maxCols {
			e.maxCols = len(record)
		}

		if e.HeaderRow {
			e.pendingRows = append(e.pendingRows, record)
		} else {
			if err := e.cw.Write(record); err != nil {
				return err
			}
		}
	}
	return nil
}

// ExportPageEnd is a no-op for CSV.
func (e *Exporter) ExportPageEnd(_ *preview.PreparedPage) error { return nil }

// Finish flushes the CSV writer.
// When HeaderRow is true, it first writes the header then all accumulated rows.
func (e *Exporter) Finish() error {
	if e.HeaderRow && len(e.pendingRows) > 0 {
		// Pad header to maxCols.
		hdr := padRow(e.header, e.maxCols)
		if err := e.cw.Write(hdr); err != nil {
			return err
		}
		for _, row := range e.pendingRows {
			if err := e.cw.Write(padRow(row, e.maxCols)); err != nil {
				return err
			}
		}
	} else if !e.HeaderRow {
		// Rows were written eagerly; just flush.
	}
	e.cw.Flush()
	return e.cw.Error()
}

// ── helpers ───────────────────────────────────────────────────────────────────

// collectTextObjects returns all text-bearing objects from a band.
func collectTextObjects(b *preview.PreparedBand) []preview.PreparedObject {
	var out []preview.PreparedObject
	for _, obj := range b.Objects {
		switch obj.Kind {
		case preview.ObjectTypeText,
			preview.ObjectTypeHtml,
			preview.ObjectTypeRTF,
			preview.ObjectTypeCheckBox:
			out = append(out, obj)
		}
	}
	return out
}

// objectText extracts the display text from a PreparedObject.
func objectText(obj preview.PreparedObject) string {
	if obj.Kind == preview.ObjectTypeCheckBox {
		if obj.Checked || obj.Text == "true" {
			return "true"
		}
		return "false"
	}
	return obj.Text
}

// groupByY groups objects by their Y coordinate, using an epsilon of 1 pixel.
// Returns a slice of rows; each row is a slice of objects sharing the same Y.
// Rows are returned in ascending Y order.
func groupByY(objs []preview.PreparedObject) [][]preview.PreparedObject {
	const eps = 1.0 // pixel tolerance

	type yGroup struct {
		y    float32
		objs []preview.PreparedObject
	}

	var groups []yGroup

	for _, obj := range objs {
		placed := false
		for i := range groups {
			if math.Abs(float64(obj.Top-groups[i].y)) < eps {
				groups[i].objs = append(groups[i].objs, obj)
				placed = true
				break
			}
		}
		if !placed {
			groups = append(groups, yGroup{y: obj.Top, objs: []preview.PreparedObject{obj}})
		}
	}

	// Sort groups by Y ascending.
	sort.Slice(groups, func(i, j int) bool {
		return groups[i].y < groups[j].y
	})

	rows := make([][]preview.PreparedObject, len(groups))
	for i, g := range groups {
		rows[i] = g.objs
	}
	return rows
}

// appendUnique appends s to slice only if it is not already present.
func appendUnique(slice []string, s string) []string {
	for _, v := range slice {
		if v == s {
			return slice
		}
	}
	return append(slice, s)
}

// padRow returns row padded with empty strings to length n.
func padRow(row []string, n int) []string {
	if len(row) >= n {
		return row
	}
	padded := make([]string, n)
	copy(padded, row)
	return padded
}
