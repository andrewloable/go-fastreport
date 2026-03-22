// Package csv provides a CSV data source for go-fastreport.
// It is the Go equivalent of FastReport.Data.CsvDataConnection.
package csv

import (
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/andrewloable/go-fastreport/data"
)

// CSVDataSource is a DataSource backed by a CSV file or string.
// The first row is treated as a header row by default (HasHeader = true).
type CSVDataSource struct {
	data.BaseDataSource

	// sourcePath is the file path to read CSV from.
	sourcePath string
	// sourceString is a raw CSV string.
	sourceString string
	// sourceStringSet indicates SetCSV was called (allows empty string).
	sourceStringSet bool
	// separator is the column delimiter (default: comma ',').
	separator rune
	// hasHeader indicates the first row contains column names (default: true).
	hasHeader bool
	// comment is the character that starts a comment line (0 = disabled).
	comment rune
	// lazyQuotes enables relaxed quote parsing.
	lazyQuotes bool
	// convertFieldTypes enables automatic type inference per column.
	// Mirrors C# CsvConnectionStringBuilder.ConvertFieldTypes / CsvUtils.DetermineTypes.
	convertFieldTypes bool
}

// New creates a CSVDataSource with the given name.
func New(name string) *CSVDataSource {
	return &CSVDataSource{
		BaseDataSource: *data.NewBaseDataSource(name),
		separator:      ',',
		hasHeader:      true,
	}
}

// SetFilePath sets the CSV file path.
func (c *CSVDataSource) SetFilePath(path string) { c.sourcePath = path }

// FilePath returns the CSV file path.
func (c *CSVDataSource) FilePath() string { return c.sourcePath }

// SetCSV sets a raw CSV string as the data source.
func (c *CSVDataSource) SetCSV(s string) { c.sourceString = s; c.sourceStringSet = true }

// CSV returns the raw CSV string.
func (c *CSVDataSource) CSV() string { return c.sourceString }

// SetSeparator sets the column delimiter rune (default: ',').
func (c *CSVDataSource) SetSeparator(r rune) { c.separator = r }

// Separator returns the column delimiter.
func (c *CSVDataSource) Separator() rune { return c.separator }

// SetHasHeader sets whether the first row is a header row (default: true).
func (c *CSVDataSource) SetHasHeader(v bool) { c.hasHeader = v }

// HasHeader returns whether the first row is treated as a header.
func (c *CSVDataSource) HasHeader() bool { return c.hasHeader }

// SetComment sets the comment character (0 = disabled).
func (c *CSVDataSource) SetComment(r rune) { c.comment = r }

// Comment returns the comment character.
func (c *CSVDataSource) Comment() rune { return c.comment }

// SetLazyQuotes enables relaxed quote parsing.
func (c *CSVDataSource) SetLazyQuotes(v bool) { c.lazyQuotes = v }

// LazyQuotes returns whether relaxed quote parsing is enabled.
func (c *CSVDataSource) LazyQuotes() bool { return c.lazyQuotes }

// Init loads and parses the CSV, populating the in-memory row store.
func (c *CSVDataSource) Init() error {
	r, closer, err := c.openReader()
	if err != nil {
		return fmt.Errorf("CSVDataSource %q: %w", c.Name(), err)
	}
	if closer != nil {
		defer closer()
	}

	cr := csv.NewReader(r)
	cr.Comma = c.separator
	cr.Comment = c.comment
	cr.LazyQuotes = c.lazyQuotes
	cr.TrimLeadingSpace = true
	cr.FieldsPerRecord = -1 // allow variable field counts per row

	allRecords, err := cr.ReadAll()
	if err != nil {
		return fmt.Errorf("CSVDataSource %q: parse error: %w", c.Name(), err)
	}

	// Reset base.
	c.BaseDataSource = *data.NewBaseDataSource(c.Name())

	if len(allRecords) == 0 {
		return c.BaseDataSource.Init()
	}

	var headers []string
	dataStart := 0
	if c.hasHeader {
		headers = allRecords[0]
		dataStart = 1
	} else {
		// Generate synthetic column names: col0, col1, …
		if len(allRecords) > 0 {
			for i := range allRecords[0] {
				headers = append(headers, fmt.Sprintf("col%d", i))
			}
		}
	}

	// Infer column types from data rows when ConvertFieldTypes is enabled.
	// Mirrors C# CsvUtils.DetermineTypes: int > float64 > time.Time > string.
	dataRecords := allRecords[dataStart:]
	colTypes := make([]string, len(headers))
	for i := range colTypes {
		colTypes[i] = "string"
	}
	if c.convertFieldTypes {
		colTypes = determineColumnTypes(headers, dataRecords)
	}

	for i, h := range headers {
		c.AddColumn(data.Column{Name: h, Alias: h, DataType: colTypes[i]})
	}

	for _, record := range dataRecords {
		row := make(map[string]any, len(headers))
		for i, h := range headers {
			var raw string
			if i < len(record) {
				raw = record[i]
			}
			row[h] = convertValue(raw, colTypes[i])
		}
		c.AddRow(row)
	}

	return c.BaseDataSource.Init()
}

// ─── helpers ──────────────────────────────────────────────────────────────────

func (c *CSVDataSource) openReader() (io.Reader, func(), error) {
	if c.sourceStringSet {
		return strings.NewReader(c.sourceString), nil, nil
	}
	if c.sourcePath != "" {
		f, err := os.Open(c.sourcePath)
		if err != nil {
			return nil, nil, err
		}
		return f, func() { f.Close() }, nil
	}
	return nil, nil, fmt.Errorf("no source configured (set FilePath or CSV)")
}
