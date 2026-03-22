package csv

// csv_convert.go — automatic field-type inference for CSVDataSource.
// Ports C# CsvUtils.DetermineTypes (FastReport.Base/Data/CsvUtils.cs).
//
// C# DetermineTypes iterates all cells per column and infers a single type using
// priority: Int32 > Decimal (currency) > Double > DateTime > String.
// Go simplification: int > float64 > time.Time > string (no currency symbol check,
// no locale-aware parsing; those depend on C# CultureInfo which has no direct Go
// equivalent outside of the golang.org/x/text subtree).

import (
	"strconv"
	"time"
)

// dateTimeLayouts are tried in order when detecting datetime columns.
// Mirrors C# DateTime.TryParse which tests multiple culture-aware date formats.
var dateTimeLayouts = []string{
	time.RFC3339,
	"2006-01-02T15:04:05",
	"2006-01-02 15:04:05",
	"2006-01-02",
	"01/02/2006",
	"01/02/2006 15:04:05",
	"2006/01/02",
	"02-01-2006",
}

// colKind enumerates the inferred type of a column in narrowest-to-widest order.
// The order matches C# DetermineTypes priority: int < float64 < datetime < string.
type colKind int

const (
	kindInt colKind = iota
	kindFloat
	kindDateTime
	kindString
)

// determineColumnTypes scans all data rows to infer the narrowest Go type for
// each column, matching C# CsvUtils.DetermineTypes behaviour:
//   - All cells parse as int64  → "int"
//   - Mix of int64 and float64  → "float64"  (C# int+double → double)
//   - All cells parse as float  → "float64"
//   - All cells are dates       → "time.Time"
//   - Any cell is unparseable   → "string"
//
// Empty cells are skipped (they don't affect type inference).
func determineColumnTypes(headers []string, rows [][]string) []string {
	n := len(headers)
	kinds := make([]colKind, n) // all start at kindInt

	for _, record := range rows {
		for i := range headers {
			if kinds[i] == kindString {
				continue // already widened to string — nothing to do
			}
			var raw string
			if i < len(record) {
				raw = record[i]
			}
			if raw == "" {
				continue // empty cells don't affect type inference
			}
			kinds[i] = widenKind(kinds[i], raw)
		}
	}

	result := make([]string, n)
	for i, k := range kinds {
		switch k {
		case kindInt:
			result[i] = "int"
		case kindFloat:
			result[i] = "float64"
		case kindDateTime:
			result[i] = "time.Time"
		default:
			result[i] = "string"
		}
	}
	return result
}

// widenKind returns the narrowest kind >= current that can parse raw.
// If no numeric/datetime format matches, returns kindString.
func widenKind(current colKind, raw string) colKind {
	for k := current; k <= kindString; k++ {
		if canParse(raw, k) {
			return k
		}
	}
	return kindString
}

// canParse returns true if raw can be parsed as the given kind.
func canParse(raw string, k colKind) bool {
	switch k {
	case kindInt:
		_, err := strconv.ParseInt(raw, 10, 64)
		return err == nil
	case kindFloat:
		_, err := strconv.ParseFloat(raw, 64)
		return err == nil
	case kindDateTime:
		for _, layout := range dateTimeLayouts {
			if _, err := time.Parse(layout, raw); err == nil {
				return true
			}
		}
		return false
	default:
		return true
	}
}

// convertValue parses raw into the Go type implied by dataType.
// On parse failure it falls back to the raw string, preserving data integrity.
func convertValue(raw, dataType string) any {
	if raw == "" {
		return raw
	}
	switch dataType {
	case "int":
		if v, err := strconv.ParseInt(raw, 10, 64); err == nil {
			return int(v)
		}
	case "float64":
		if v, err := strconv.ParseFloat(raw, 64); err == nil {
			return v
		}
	case "time.Time":
		for _, layout := range dateTimeLayouts {
			if v, err := time.Parse(layout, raw); err == nil {
				return v
			}
		}
	}
	return raw
}
