package csv

// csv_fieldtypes.go — ConvertFieldTypes support and NewFromConnectionString
// constructor for CSVDataSource.
//
// Ports:
//   - C# CsvDataConnection.ConvertFieldTypes property (CsvDataConnection.cs)
//   - C# CsvDataConnection constructor + CheckForChangeConnection (CsvDataConnection.Core.cs)
//   - C# CsvUtils.CreateDataTable type-conversion path (CsvUtils.cs)

import (
	"github.com/andrewloable/go-fastreport/data"
)

// NewFromConnectionString creates a CSVDataSource from a FastReport connection
// string (e.g. "CsvFile=/data.csv;Separator=,;FieldNamesInFirstString=true").
// This mirrors the C# CsvDataConnection constructor + property setters pattern,
// where each property setter calls CheckForChangeConnection(builder) to
// normalise and persist the connection string.
func NewFromConnectionString(name, connectionString string) *CSVDataSource {
	b := NewConnectionStringBuilder(connectionString)
	sep := rune(';') // ConnectionStringBuilder default separator
	if s := b.Separator(); s != "" {
		runes := []rune(s)
		sep = runes[0]
	}
	return &CSVDataSource{
		BaseDataSource:    *data.NewBaseDataSource(name),
		sourcePath:        b.CsvFile(),
		separator:         sep,
		hasHeader:         b.FieldNamesInFirstString(),
		convertFieldTypes: b.ConvertFieldTypes(),
	}
}

// SetConvertFieldTypes enables automatic type inference per column.
// When true, Init() will attempt to convert each column to int, float64,
// or time.Time — falling back to string — matching C# DetermineTypes logic.
// See csv_convert.go for the inference algorithm.
func (c *CSVDataSource) SetConvertFieldTypes(v bool) { c.convertFieldTypes = v }

// ConvertFieldTypes returns whether automatic type conversion is enabled.
func (c *CSVDataSource) ConvertFieldTypes() bool { return c.convertFieldTypes }
