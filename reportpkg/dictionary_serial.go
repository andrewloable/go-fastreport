package reportpkg

// dictionary_serial.go — serialization of the report Dictionary to FRX XML.
//
// The C# counterpart is FastReport.Data.Dictionary.Serialize (Dictionary.cs:649-661):
//
//	foreach (Base c in childObjects)
//	{
//	    if (c is Parameter || c is Total || c is CubeSourceBase ||
//	        (c is DataComponentBase && (c as DataComponentBase).Enabled))
//	        writer.Write(c);
//	}
//
// We write: Parameters (always), Totals (always), enabled Connections (with nested
// TableDataSources), enabled standalone DataSources (BusinessObjectDataSource,
// TableDataSource without a connection), and Relations.

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/andrewloable/go-fastreport/data"
	"github.com/andrewloable/go-fastreport/report"
)

// dictionarySerializer wraps a data.Dictionary for writing as a <Dictionary>
// child element via report.Writer.WriteObject. It follows the same adapter
// pattern as stylesSerializer in styles_serial.go.
type dictionarySerializer struct {
	dict *data.Dictionary
}

// TypeName implements serial.TypeNamer; returns "Dictionary" as the XML element name.
func (d *dictionarySerializer) TypeName() string { return "Dictionary" }

// Serialize writes dictionary child elements. The order mirrors C#:
// connections first (with their table data sources), then standalone data
// sources, relations, parameters, and totals.
//
// C# ref: FastReport.Data.Dictionary.Serialize (Dictionary.cs:649-661) and
//
//	FastReport.Data.Dictionary.GetChildObjects (Dictionary.cs:791-817)
func (d *dictionarySerializer) Serialize(w report.Writer) error {
	dict := d.dict

	// 1. Connections (DataComponentBase, written only when Enabled).
	for _, conn := range dict.Connections() {
		if !conn.Enabled() {
			continue
		}
		s := &connectionSerializer{conn: conn}
		if err := w.WriteObject(s); err != nil {
			return err
		}
	}

	// 2. Standalone data sources (not associated with a connection).
	// These are DataSources that the Dictionary tracks directly (e.g. from
	// RegisterData, or standalone TableDataSource / BusinessObjectDataSource).
	// Skip data sources that belong to a connection (they are serialized above).
	connSet := make(map[string]struct{}, len(dict.Connections()))
	for _, conn := range dict.Connections() {
		for _, t := range conn.Tables() {
			connSet[t.Name()] = struct{}{}
		}
	}
	for _, ds := range dict.DataSources() {
		if _, owned := connSet[ds.Name()]; owned {
			continue
		}
		// Only serialize enabled data components. BusinessObjectDataSource and
		// other Go-native sources implement Enabled() via DataComponentBase.
		if ec, ok := ds.(interface{ Enabled() bool }); ok && !ec.Enabled() {
			continue
		}
		s := dataSourceSerializer(ds)
		if s == nil {
			continue
		}
		if err := w.WriteObject(s); err != nil {
			return err
		}
	}

	// 3. Relations.
	for _, rel := range dict.Relations() {
		s := &relationSerializer{rel: rel}
		if err := w.WriteObject(s); err != nil {
			return err
		}
	}

	// 4. Parameters (always written).
	for _, p := range dict.Parameters() {
		s := &parameterSerializer{param: p}
		if err := w.WriteObject(s); err != nil {
			return err
		}
	}

	// 5. Totals (always written).
	for _, t := range dict.Totals() {
		s := &totalSerializer{total: t}
		if err := w.WriteObject(s); err != nil {
			return err
		}
	}

	return nil
}

// Deserialize is a no-op (loading is handled separately in loadsave.go).
func (d *dictionarySerializer) Deserialize(_ report.Reader) error { return nil }

// ── connectionSerializer ─────────────────────────────────────────────────────

// connectionSerializer serializes a DataConnectionBase as the appropriate
// connection element (e.g. <JsonDataConnection>, <CsvDataConnection>, etc.).
type connectionSerializer struct {
	conn *data.DataConnectionBase
}

// TypeName returns the FRX element name for the connection based on its driver.
// C# mapping: driver "json" → "JsonDataConnection", "csv" → "CsvDataConnection",
// "xml" → "XmlDataConnection", "sql" → "TableDataSource", etc.
func (s *connectionSerializer) TypeName() string {
	switch strings.ToLower(s.conn.DriverName()) {
	case "json":
		return "JsonDataConnection"
	case "csv":
		return "CsvDataConnection"
	case "xml":
		return "XmlDataConnection"
	default:
		return "DataConnection"
	}
}

// Serialize writes the connection attributes and nested TableDataSources.
func (s *connectionSerializer) Serialize(w report.Writer) error {
	conn := s.conn
	if conn.Name() != "" {
		w.WriteStr("Name", conn.Name())
	}
	if conn.ConnectionString != "" {
		w.WriteStr("ConnectionString", conn.ConnectionString)
	}
	if !conn.Enabled() {
		w.WriteBool("Enabled", false)
	}
	// Nested table data sources.
	for _, t := range conn.Tables() {
		ts := &tableDataSourceSerializer{ts: t}
		if err := w.WriteObject(ts); err != nil {
			return err
		}
	}
	return nil
}

// Deserialize is a no-op.
func (s *connectionSerializer) Deserialize(_ report.Reader) error { return nil }

// ── tableDataSourceSerializer ─────────────────────────────────────────────────

type tableDataSourceSerializer struct {
	ts *data.TableDataSource
}

func (s *tableDataSourceSerializer) TypeName() string { return "TableDataSource" }

func (s *tableDataSourceSerializer) Serialize(w report.Writer) error {
	ts := s.ts
	if ts.Name() != "" {
		w.WriteStr("Name", ts.Name())
	}
	if ts.Alias() != ts.Name() && ts.Alias() != "" {
		w.WriteStr("Alias", ts.Alias())
	}
	if ts.TableName() != "" {
		w.WriteStr("TableName", ts.TableName())
	}
	if ts.SelectCommand() != "" {
		w.WriteStr("SelectCommand", ts.SelectCommand())
	}
	if ts.StoreData() {
		w.WriteBool("StoreData", true)
	}
	return nil
}

func (s *tableDataSourceSerializer) Deserialize(_ report.Reader) error { return nil }

// ── businessObjectDataSourceSerializer ──────────────────────────────────────

type businessObjectDataSourceSerializer struct {
	ds *data.BusinessObjectDataSource
}

func (s *businessObjectDataSourceSerializer) TypeName() string {
	return "BusinessObjectDataSource"
}

func (s *businessObjectDataSourceSerializer) Serialize(w report.Writer) error {
	ds := s.ds
	if ds.Name() != "" {
		w.WriteStr("Name", ds.Name())
	}
	if ds.Alias() != ds.Name() && ds.Alias() != "" {
		w.WriteStr("Alias", ds.Alias())
	}
	return nil
}

func (s *businessObjectDataSourceSerializer) Deserialize(_ report.Reader) error { return nil }

// ── dataSourceSerializer factory ─────────────────────────────────────────────

// dataSourceSerializer returns the appropriate serializer for a DataSource,
// or nil when the data source type is not serializable (e.g. system data sources).
func dataSourceSerializer(ds data.DataSource) report.Serializable {
	switch v := ds.(type) {
	case *data.TableDataSource:
		return &tableDataSourceSerializer{ts: v}
	case *data.BusinessObjectDataSource:
		return &businessObjectDataSourceSerializer{ds: v}
	default:
		return nil
	}
}

// ── relationSerializer ────────────────────────────────────────────────────────

type relationSerializer struct {
	rel *data.Relation
}

func (s *relationSerializer) TypeName() string { return "Relation" }

func (s *relationSerializer) Serialize(w report.Writer) error {
	rel := s.rel
	if rel.Name != "" {
		w.WriteStr("Name", rel.Name)
	}
	// ParentDataSource and ChildDataSource are written by alias/name.
	// C# ref: FastReport.Data.Relation.Serialize
	parentName := rel.ParentSourceName
	if rel.ParentDataSource != nil {
		parentName = rel.ParentDataSource.Name()
	}
	if parentName != "" {
		w.WriteStr("ParentDataSource", parentName)
	}
	childName := rel.ChildSourceName
	if rel.ChildDataSource != nil {
		childName = rel.ChildDataSource.Name()
	}
	if childName != "" {
		w.WriteStr("ChildDataSource", childName)
	}
	// Column names.
	parentCols := rel.ParentColumns
	if len(parentCols) == 0 {
		parentCols = rel.ParentColumnNames
	}
	if len(parentCols) > 0 {
		w.WriteStr("ParentColumns", strings.Join(parentCols, ","))
	}
	childCols := rel.ChildColumns
	if len(childCols) == 0 {
		childCols = rel.ChildColumnNames
	}
	if len(childCols) > 0 {
		w.WriteStr("ChildColumns", strings.Join(childCols, ","))
	}
	return nil
}

func (s *relationSerializer) Deserialize(_ report.Reader) error { return nil }

// ── parameterSerializer ───────────────────────────────────────────────────────

type parameterSerializer struct {
	param *data.Parameter
}

func (s *parameterSerializer) TypeName() string { return "Parameter" }

func (s *parameterSerializer) Serialize(w report.Writer) error {
	p := s.param
	if p.Name != "" {
		w.WriteStr("Name", p.Name)
	}
	if p.DataType != "" {
		w.WriteStr("DataType", p.DataType)
	}
	// C# writes AsString when Expression is empty; Expression otherwise.
	// C# ref: FastReport.Data.Parameter.Serialize (Parameter.cs:188-198)
	if p.Expression == "" {
		if p.Value != nil {
			w.WriteStr("AsString", fmt.Sprint(p.Value))
		}
	} else {
		w.WriteStr("Expression", p.Expression)
	}
	if p.Description != "" {
		w.WriteStr("Description", p.Description)
	}
	// Nested parameters.
	for _, child := range p.Parameters() {
		cs := &parameterSerializer{param: child}
		if err := w.WriteObject(cs); err != nil {
			return err
		}
	}
	return nil
}

func (s *parameterSerializer) Deserialize(_ report.Reader) error { return nil }

// ── totalSerializer ───────────────────────────────────────────────────────────

type totalSerializer struct {
	total *data.Total
}

func (s *totalSerializer) TypeName() string { return "Total" }

func (s *totalSerializer) Serialize(w report.Writer) error {
	t := s.total
	if t.Name != "" {
		w.WriteStr("Name", t.Name)
	}
	// Write TotalType string. C# ref: FastReport.Data.Total.Serialize
	// Default is Sum; only write when non-default.
	tt := totalTypeToString(t.TotalType)
	if tt != "" && tt != "Sum" {
		w.WriteStr("TotalType", tt)
	}
	if t.Expression != "" {
		w.WriteStr("Expression", t.Expression)
	}
	if t.Evaluator != "" {
		w.WriteStr("Evaluator", t.Evaluator)
	}
	if t.PrintOn != "" {
		w.WriteStr("PrintOn", t.PrintOn)
	}
	if t.ResetAfterPrint {
		w.WriteBool("ResetAfterPrint", true)
	}
	// C# default for ResetOnReprint is true; write only when false.
	if !t.ResetOnReprint {
		w.WriteBool("ResetOnReprint", false)
	}
	if t.EvaluateCondition != "" {
		w.WriteStr("EvaluateCondition", t.EvaluateCondition)
	}
	if t.IncludeInvisibleRows {
		w.WriteBool("IncludeInvisibleRows", true)
	}
	return nil
}

func (s *totalSerializer) Deserialize(_ report.Reader) error { return nil }

// totalTypeToString converts a TotalType constant to its FRX string representation.
func totalTypeToString(tt data.TotalType) string {
	switch tt {
	case data.TotalTypeSum:
		return "Sum"
	case data.TotalTypeMin:
		return "Min"
	case data.TotalTypeMax:
		return "Max"
	case data.TotalTypeAvg:
		return "Avg"
	case data.TotalTypeCount:
		return "Count"
	case data.TotalTypeCountDistinct:
		return "CountDistinct"
	default:
		return strconv.Itoa(int(tt))
	}
}

// ── hasDictionaryContent ──────────────────────────────────────────────────────

// hasDictionaryContent returns true when the dictionary has at least one item
// that would be written to the <Dictionary> element. This avoids emitting an
// empty <Dictionary/> element for reports that have no data definitions.
func hasDictionaryContent(dict *data.Dictionary) bool {
	return len(dict.Connections()) > 0 ||
		len(dict.DataSources()) > 0 ||
		len(dict.Relations()) > 0 ||
		len(dict.Parameters()) > 0 ||
		len(dict.Totals()) > 0
}
