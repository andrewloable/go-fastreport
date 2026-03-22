package csv

// connection_string_setters.go — setters and String() serialisation for
// ConnectionStringBuilder, mirroring C# CsvDataConnection property setters
// and CsvConnectionStringBuilder.ToString() / CheckForChangeConnection.
// (FastReport.Base/Data/CsvDataConnection.cs + FastReport.OpenSource/Data/CsvDataConnection.Core.cs)

import (
	"fmt"
	"strconv"
	"strings"
)

// canonicalKeys maps lowercase key → canonical key name as used in C# connection strings.
// This ensures String() emits proper casing that FastReport .NET will recognise.
var canonicalKeys = map[string]string{
	"csvfile":                 "CsvFile",
	"codepage":                "Codepage",
	"separator":               "Separator",
	"fieldnamesinfirststring": "FieldNamesInFirstString",
	"removequotationmarks":    "RemoveQuotationMarks",
	"convertfieldtypes":       "ConvertFieldTypes",
	"numberformat":            "NumberFormat",
	"currencyformat":          "CurrencyFormat",
	"datetimeformat":          "DateTimeFormat",
}

// set stores a key-value pair using a lowercase key.
func (b *ConnectionStringBuilder) set(key, value string) {
	b.vals[strings.ToLower(key)] = value
}

// SetCsvFile sets the CsvFile property.
func (b *ConnectionStringBuilder) SetCsvFile(v string) { b.set("csvfile", v) }

// SetCodepage sets the Codepage property.
func (b *ConnectionStringBuilder) SetCodepage(v int) { b.set("codepage", strconv.Itoa(v)) }

// SetSeparator sets the Separator property.
func (b *ConnectionStringBuilder) SetSeparator(v string) { b.set("separator", v) }

// SetFieldNamesInFirstString sets the FieldNamesInFirstString property.
func (b *ConnectionStringBuilder) SetFieldNamesInFirstString(v bool) {
	b.set("fieldnamesinfirststring", fmt.Sprintf("%t", v))
}

// SetRemoveQuotationMarks sets the RemoveQuotationMarks property.
func (b *ConnectionStringBuilder) SetRemoveQuotationMarks(v bool) {
	b.set("removequotationmarks", fmt.Sprintf("%t", v))
}

// SetConvertFieldTypes sets the ConvertFieldTypes property.
func (b *ConnectionStringBuilder) SetConvertFieldTypes(v bool) {
	b.set("convertfieldtypes", fmt.Sprintf("%t", v))
}

// SetNumberFormat sets the NumberFormat locale property.
func (b *ConnectionStringBuilder) SetNumberFormat(v string) { b.set("numberformat", v) }

// SetCurrencyFormat sets the CurrencyFormat locale property.
func (b *ConnectionStringBuilder) SetCurrencyFormat(v string) { b.set("currencyformat", v) }

// SetDateTimeFormat sets the DateTimeFormat locale property.
func (b *ConnectionStringBuilder) SetDateTimeFormat(v string) { b.set("datetimeformat", v) }

// String serialises the builder back to a FastReport connection string.
// Mirrors C# CsvConnectionStringBuilder.ToString() used by CheckForChangeConnection.
// Only keys that have been explicitly set are included.
func (b *ConnectionStringBuilder) String() string {
	// Emit keys in a stable order matching C# property declaration order.
	order := []string{
		"csvfile",
		"codepage",
		"separator",
		"fieldnamesinfirststring",
		"removequotationmarks",
		"convertfieldtypes",
		"numberformat",
		"currencyformat",
		"datetimeformat",
	}
	var parts []string
	for _, lk := range order {
		if v, ok := b.vals[lk]; ok {
			canonical := canonicalKeys[lk]
			if canonical == "" {
				canonical = lk
			}
			parts = append(parts, canonical+"="+v)
		}
	}
	return strings.Join(parts, ";")
}
