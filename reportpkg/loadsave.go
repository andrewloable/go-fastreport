package reportpkg

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/andrewloable/go-fastreport/data"
	csvdata "github.com/andrewloable/go-fastreport/data/csv"
	jsondata "github.com/andrewloable/go-fastreport/data/json"
	xmldata "github.com/andrewloable/go-fastreport/data/xml"
	"github.com/andrewloable/go-fastreport/report"
	"github.com/andrewloable/go-fastreport/serial"
	"github.com/andrewloable/go-fastreport/style"
	"github.com/andrewloable/go-fastreport/utils"
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
// If the stream starts with the gzip magic bytes (0x1f 0x8b), it is
// transparently decompressed before parsing.
func (r *Report) LoadFrom(rd io.Reader) error {
	// Buffer enough bytes to detect gzip magic without consuming the stream.
	var peek [2]byte
	buf := &bytes.Buffer{}
	if _, err := io.ReadFull(rd, peek[:]); err != nil {
		return fmt.Errorf("report.LoadFrom: read header: %w", err)
	}
	buf.Write(peek[:])

	var src io.Reader = io.MultiReader(buf, rd)
	if peek[0] == 0x1f && peek[1] == 0x8b {
		// Gzip-compressed FRX.
		gz, err := gzip.NewReader(src)
		if err != nil {
			return fmt.Errorf("report.LoadFrom: gzip open: %w", err)
		}
		defer gz.Close()
		src = gz
	}

	rdr := serial.NewReader(src)

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
		case "Styles":
			deserializeStyles(rdr, r.Styles())
		case "Dictionary":
			deserializeDictionary(rdr, r.Dictionary())
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
		// Give the parent object a chance to handle the child itself (e.g.
		// TextObject handles its own <Highlight> children this way).
		if cd, isCD := parent.(report.ChildDeserializer); isCD {
			if cd.DeserializeChild(childType, rdr) {
				_ = rdr.FinishChild()
				continue
			}
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


// deserializeDictionary reads the <Dictionary> element and populates the
// report dictionary with Parameters, Relations, Totals, and data connections.
func deserializeDictionary(rdr *serial.Reader, dict *data.Dictionary) {
	for {
		childType, ok := rdr.NextChild()
		if !ok {
			break
		}
		switch childType {
		case "Parameter":
			p := deserializeParameter(rdr)
			dict.AddParameter(p)
		case "Relation":
			r := deserializeRelation(rdr)
			dict.AddRelation(r)
		case "Total":
			t := deserializeTotal(rdr)
			dict.AddTotal(t)
		case "JsonDataConnection":
			conn, sources := deserializeJsonConnection(rdr)
			dict.AddConnection(conn)
			for _, ds := range sources {
				dict.AddDataSource(ds)
			}
		case "CsvDataConnection":
			conn, sources := deserializeCsvConnection(rdr)
			dict.AddConnection(conn)
			for _, ds := range sources {
				dict.AddDataSource(ds)
			}
		case "XmlDataConnection":
			conn, sources := deserializeXmlConnection(rdr)
			dict.AddConnection(conn)
			for _, ds := range sources {
				dict.AddDataSource(ds)
			}
		case "TableDataSource":
			// Top-level TableDataSource without a connection — load as a standalone source.
			ts := deserializeTableDataSource(rdr, nil)
			dict.AddDataSource(ts)
		default:
			// Unknown children are skipped.
			_ = rdr.SkipElement()
		}
		_ = rdr.FinishChild()
	}
}

// deserializeJsonConnection reads a <JsonDataConnection> element and returns
// the connection stub plus any nested data sources it exposes.
func deserializeJsonConnection(rdr *serial.Reader) (*data.DataConnectionBase, []data.DataSource) {
	name := rdr.ReadStr("Name", "JsonConnection")
	connStr := rdr.ReadStr("ConnectionString", "")
	enabled := rdr.ReadBool("Enabled", true)

	conn := data.NewDataConnectionBase("json")
	conn.SetName(name)
	conn.SetEnabled(enabled)
	conn.ConnectionString = connStr

	// The connection string for JsonDataConnection contains the file path.
	filePath := connStr

	var sources []data.DataSource
	for {
		ct, ok := rdr.NextChild()
		if !ok {
			break
		}
		if ct == "TableDataSource" {
			ds := deserializeJsonTableDataSource(rdr, filePath)
			sources = append(sources, ds)
		} else {
			_ = rdr.SkipElement()
		}
		_ = rdr.FinishChild()
	}
	return conn, sources
}

// deserializeJsonTableDataSource reads a <TableDataSource> nested inside a
// JsonDataConnection and returns a JSONDataSource configured with the file path.
func deserializeJsonTableDataSource(rdr *serial.Reader, filePath string) data.DataSource {
	name := rdr.ReadStr("Name", "")
	alias := rdr.ReadStr("Alias", name)
	rootPath := rdr.ReadStr("TableName", "")
	// Drain children.
	for {
		_, ok := rdr.NextChild()
		if !ok {
			break
		}
		_ = rdr.FinishChild()
	}
	ds := jsondata.New(name)
	ds.SetAlias(alias)
	ds.SetFilePath(filePath)
	ds.SetRootPath(rootPath)
	return ds
}

// deserializeCsvConnection reads a <CsvDataConnection> element.
func deserializeCsvConnection(rdr *serial.Reader) (*data.DataConnectionBase, []data.DataSource) {
	name := rdr.ReadStr("Name", "CsvConnection")
	connStr := rdr.ReadStr("ConnectionString", "")
	enabled := rdr.ReadBool("Enabled", true)

	conn := data.NewDataConnectionBase("csv")
	conn.SetName(name)
	conn.SetEnabled(enabled)
	conn.ConnectionString = connStr

	// CSV connection string is the file path.
	filePath := connStr

	var sources []data.DataSource
	for {
		ct, ok := rdr.NextChild()
		if !ok {
			break
		}
		if ct == "TableDataSource" {
			tableName := rdr.ReadStr("Name", "")
			alias := rdr.ReadStr("Alias", tableName)
			separator := rdr.ReadStr("Separator", ",")
			hasHeader := rdr.ReadBool("HasHeader", true)
			// Drain children.
			for {
				_, ok2 := rdr.NextChild()
				if !ok2 {
					break
				}
				_ = rdr.FinishChild()
			}
			ds := csvdata.New(tableName)
			ds.SetFilePath(filePath)
			ds.SetAlias(alias)
			if len(separator) > 0 {
				ds.SetSeparator(rune(separator[0]))
			}
			ds.SetHasHeader(hasHeader)
			sources = append(sources, ds)
		} else {
			_ = rdr.SkipElement()
		}
		_ = rdr.FinishChild()
	}
	return conn, sources
}

// deserializeXmlConnection reads an <XmlDataConnection> element.
func deserializeXmlConnection(rdr *serial.Reader) (*data.DataConnectionBase, []data.DataSource) {
	name := rdr.ReadStr("Name", "XmlConnection")
	connStr := rdr.ReadStr("ConnectionString", "")
	enabled := rdr.ReadBool("Enabled", true)

	conn := data.NewDataConnectionBase("xml")
	conn.SetName(name)
	conn.SetEnabled(enabled)
	conn.ConnectionString = connStr

	filePath := connStr

	var sources []data.DataSource
	for {
		ct, ok := rdr.NextChild()
		if !ok {
			break
		}
		if ct == "TableDataSource" {
			tableName := rdr.ReadStr("Name", "")
			alias := rdr.ReadStr("Alias", tableName)
			xPath := rdr.ReadStr("TableName", "")
			// Drain children.
			for {
				_, ok2 := rdr.NextChild()
				if !ok2 {
					break
				}
				_ = rdr.FinishChild()
			}
			ds := xmldata.New(tableName)
			ds.SetFilePath(filePath)
			ds.SetAlias(alias)
			ds.SetRootPath(xPath)
			sources = append(sources, ds)
		} else {
			_ = rdr.SkipElement()
		}
		_ = rdr.FinishChild()
	}
	return conn, sources
}

// deserializeTableDataSource reads a standalone <TableDataSource> element.
func deserializeTableDataSource(rdr *serial.Reader, conn *data.DataConnectionBase) *data.TableDataSource {
	name := rdr.ReadStr("Name", "")
	alias := rdr.ReadStr("Alias", name)
	tableName := rdr.ReadStr("TableName", "")
	selectCmd := rdr.ReadStr("SelectCommand", "")
	// Drain children (CommandParameter etc.).
	for {
		_, ok := rdr.NextChild()
		if !ok {
			break
		}
		_ = rdr.FinishChild()
	}
	ts := data.NewTableDataSource(name)
	ts.SetAlias(alias)
	ts.SetTableName(tableName)
	ts.SetSelectCommand(selectCmd)
	return ts
}

// deserializeParameter reads a single <Parameter> element (and nested children).
func deserializeParameter(rdr *serial.Reader) *data.Parameter {
	p := &data.Parameter{
		Name:       rdr.ReadStr("Name", ""),
		DataType:   rdr.ReadStr("DataType", ""),
		Expression: rdr.ReadStr("Expression", ""),
	}
	// Read nested Parameter children.
	for {
		ct, ok := rdr.NextChild()
		if !ok {
			break
		}
		if ct == "Parameter" {
			child := deserializeParameter(rdr)
			p.AddParameter(child)
		} else {
			_ = rdr.SkipElement()
		}
		_ = rdr.FinishChild()
	}
	return p
}

// deserializeRelation reads a single <Relation> element.
// Data source references are stored by name only; the engine resolves them
// at prepare time via Dictionary.FindDataSourceByAlias.
func deserializeRelation(rdr *serial.Reader) *data.Relation {
	r := &data.Relation{
		Name:              rdr.ReadStr("Name", ""),
		ParentSourceName:  rdr.ReadStr("ParentDataSource", ""),
		ChildSourceName:   rdr.ReadStr("ChildDataSource", ""),
		ParentColumnNames: splitComma(rdr.ReadStr("ParentColumns", "")),
		ChildColumnNames:  splitComma(rdr.ReadStr("ChildColumns", "")),
	}
	// Drain children (none expected).
	for {
		_, ok := rdr.NextChild()
		if !ok {
			break
		}
		_ = rdr.FinishChild()
	}
	return r
}

// deserializeTotal reads a single <Total> element.
func deserializeTotal(rdr *serial.Reader) *data.Total {
	t := &data.Total{
		Name:       rdr.ReadStr("Name", ""),
		Expression: rdr.ReadStr("Expression", ""),
		TotalType:  parseTotalType(rdr.ReadStr("TotalType", "Sum")),
		Evaluator:  rdr.ReadStr("Evaluator", ""),
		PrintOn:    rdr.ReadStr("PrintOn", ""),
	}
	// Drain children (none expected).
	for {
		_, ok := rdr.NextChild()
		if !ok {
			break
		}
		_ = rdr.FinishChild()
	}
	return t
}

// parseTotalType maps a FastReport TotalType string to a data.TotalType constant.
func parseTotalType(s string) data.TotalType {
	switch strings.ToLower(s) {
	case "sum":
		return data.TotalTypeSum
	case "min":
		return data.TotalTypeMin
	case "max":
		return data.TotalTypeMax
	case "avg":
		return data.TotalTypeAvg
	case "count":
		return data.TotalTypeCount
	case "countdistinct":
		return data.TotalTypeCountDistinct
	default:
		return data.TotalTypeSum
	}
}

// deserializeStyles reads the <Styles> element and populates the report's
// StyleSheet with named Style entries.
func deserializeStyles(rdr *serial.Reader, ss *style.StyleSheet) {
	for {
		childType, ok := rdr.NextChild()
		if !ok {
			break
		}
		if childType == "Style" {
			entry := deserializeStyleEntry(rdr)
			if entry.Name != "" {
				ss.Add(entry)
			}
		} else {
			_ = rdr.SkipElement()
		}
		_ = rdr.FinishChild()
	}
}

// deserializeStyleEntry reads a single <Style> element into a StyleEntry.
func deserializeStyleEntry(rdr *serial.Reader) *style.StyleEntry {
	e := &style.StyleEntry{
		Name:          rdr.ReadStr("Name", ""),
		ApplyBorder:   rdr.ReadBool("ApplyBorder", true),
		ApplyFill:     rdr.ReadBool("ApplyFill", true),
		ApplyTextFill: rdr.ReadBool("ApplyTextFill", true),
		ApplyFont:     rdr.ReadBool("ApplyFont", true),
	}
	// Set legacy flags from Apply* flags.
	e.FillColorChanged = e.ApplyFill
	e.FontChanged = e.ApplyFont
	e.TextColorChanged = e.ApplyTextFill
	e.BorderColorChanged = e.ApplyBorder

	// Fill color.
	if s := rdr.ReadStr("Fill.Color", ""); s != "" {
		if c, err := utils.ParseColor(s); err == nil {
			e.FillColor = c
		}
	}
	// Text fill color.
	if s := rdr.ReadStr("TextFill.Color", ""); s != "" {
		if c, err := utils.ParseColor(s); err == nil {
			e.TextColor = c
		}
	}
	// Border color (simple, all lines).
	if s := rdr.ReadStr("Border.Color", ""); s != "" {
		if c, err := utils.ParseColor(s); err == nil {
			e.BorderColor = c
			// Populate the Border struct too.
			e.Border = *style.NewBorder()
			e.Border.SetColor(c)
		}
	}
	// Border lines visibility.
	if s := rdr.ReadStr("Border.Lines", ""); s != "" {
		if e.Border.Lines[0] == nil {
			e.Border = *style.NewBorder()
		}
		e.Border.VisibleLines = parseBorderLines(s)
	}
	// Border shadow.
	if rdr.ReadBool("Border.Shadow", false) {
		if e.Border.Lines[0] == nil {
			e.Border = *style.NewBorder()
		}
		e.Border.Shadow = true
	}
	// Font string.
	if s := rdr.ReadStr("Font", ""); s != "" {
		e.Font = style.FontFromStr(s)
		e.FontChanged = true
		e.ApplyFont = true
	}

	// Drain children (none expected for Style).
	for {
		_, ok := rdr.NextChild()
		if !ok {
			break
		}
		_ = rdr.FinishChild()
	}
	return e
}

// parseBorderLines converts a comma-separated FRX BorderLines string.
func parseBorderLines(s string) style.BorderLines {
	if s == "" {
		return style.BorderLinesNone
	}
	s = strings.TrimSpace(s)
	if strings.EqualFold(s, "None") {
		return style.BorderLinesNone
	}
	if strings.EqualFold(s, "All") {
		return style.BorderLinesAll
	}
	var result style.BorderLines
	for _, p := range strings.Split(s, ",") {
		switch strings.TrimSpace(p) {
		case "Left":
			result |= style.BorderLinesLeft
		case "Right":
			result |= style.BorderLinesRight
		case "Top":
			result |= style.BorderLinesTop
		case "Bottom":
			result |= style.BorderLinesBottom
		}
	}
	return result
}

// splitComma splits a comma-separated string into a trimmed slice.
func splitComma(s string) []string {
	if s == "" {
		return nil
	}
	parts := strings.Split(s, ",")
	for i, p := range parts {
		parts[i] = strings.TrimSpace(p)
	}
	return parts
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
// If r.Compressed is true, the output is gzip-compressed.
// Pages and their bands/objects are written as nested children via Report.Serialize().
func (r *Report) SaveTo(w io.Writer) error {
	dst := w
	if r.Compressed {
		gz := gzip.NewWriter(w)
		defer gz.Close()
		dst = gz
	}
	sw := serial.NewWriter(dst)
	if err := sw.WriteHeader(); err != nil {
		return fmt.Errorf("report.SaveTo: write header: %w", err)
	}
	if err := sw.WriteObjectNamed("Report", r); err != nil {
		return fmt.Errorf("report.SaveTo: write report: %w", err)
	}
	return sw.Flush()
}
