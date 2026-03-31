package reportpkg

import (
	"bytes"
	"compress/gzip"
	"encoding/base64"
	"encoding/xml"
	"fmt"
	"io"
	"os"
	"path/filepath"
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
// If the file uses an <inherited> root element, the base report is loaded
// from the same directory before overlaying the inherited changes.
func (r *Report) Load(filename string) error {
	f, err := os.Open(filename)
	if err != nil {
		return fmt.Errorf("report.Load: %w", err)
	}
	defer f.Close()

	// Buffer enough bytes to detect gzip magic without consuming the stream.
	var peek [2]byte
	buf := &bytes.Buffer{}
	if _, err2 := io.ReadFull(f, peek[:]); err2 != nil {
		return fmt.Errorf("report.Load: read header: %w", err2)
	}
	buf.Write(peek[:])

	var src io.Reader = io.MultiReader(buf, f)
	if peek[0] == 0x1f && peek[1] == 0x8b {
		gz, err2 := gzip.NewReader(src)
		if err2 != nil {
			return fmt.Errorf("report.Load: gzip open: %w", err2)
		}
		defer gz.Close()
		src = gz
	}

	rdr := serial.NewReader(src)
	return r.loadFromSerialReader(rdr, filepath.Dir(filename))
}

// LoadWithPassword reads a password-protected FRX report file from filename.
// If the file is not encrypted the password is ignored.
func (r *Report) LoadWithPassword(filename, password string) error {
	f, err := os.Open(filename)
	if err != nil {
		return fmt.Errorf("report.LoadWithPassword: %w", err)
	}
	defer f.Close()
	return r.LoadFromWithPassword(f, password)
}

// LoadFromString deserializes a Report from an FRX XML string.
func (r *Report) LoadFromString(xml string) error {
	return r.LoadFrom(strings.NewReader(xml))
}

// LoadFromStringWithPassword deserializes a password-protected FRX XML string.
// If the content is not encrypted the password is ignored.
func (r *Report) LoadFromStringWithPassword(xml, password string) error {
	return r.LoadFromWithPassword(strings.NewReader(xml), password)
}

// LoadFromWithPassword deserializes a Report from an io.Reader that may contain
// a FastReport-encrypted or gzip-compressed FRX stream. If the stream is
// encrypted the password is used to decrypt it; if not encrypted the password
// is silently ignored. Gzip decompression is applied after decryption.
// Note: inherited reports loaded via this method will not resolve relative
// base-report paths; use Load(filename) for full inheritance support.
func (r *Report) LoadFromWithPassword(rd io.Reader, password string) error {
	xmlRdr, _, err := serial.NewReaderWithPassword(rd, password)
	if err != nil {
		return fmt.Errorf("report.LoadFromWithPassword: %w", err)
	}
	return r.loadFromSerialReader(xmlRdr, "")
}

// LoadFrom deserializes a Report from an io.Reader containing FRX XML.
// If the stream starts with the gzip magic bytes (0x1f 0x8b), it is
// transparently decompressed before parsing.
// Note: inherited reports loaded via this method will not resolve relative
// base-report paths; use Load(filename) for full inheritance support.
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
	return r.loadFromSerialReader(rdr, "")
}

// loadFromSerialReader is the internal shared deserializer.
// baseDir is the directory of the FRX file being loaded; it is used to resolve
// the base-report path when the root element is <inherited>. Pass "" when
// loading from an io.Reader without a known file path.
func (r *Report) loadFromSerialReader(rdr *serial.Reader, baseDir string) error {
	// Read the root element — may be "Report" or "inherited".
	typeName, ok := rdr.ReadObjectHeader()
	if !ok {
		return fmt.Errorf("report.LoadFrom: empty or invalid FRX document")
	}

	switch typeName {
	case "Report":
		return r.deserializeReportBody(rdr, baseDir)
	case "inherited":
		return r.loadInherited(rdr, baseDir)
	default:
		return fmt.Errorf("report.LoadFrom: expected root element 'Report' or 'inherited', got %q", typeName)
	}
}

// deserializeReportBody reads the attributes and child elements of a <Report>
// root after the element header has already been consumed by loadFromSerialReader.
// baseDir is the directory of the FRX file (may be empty when loading from an
// io.Reader without a known file path).
func (r *Report) deserializeReportBody(rdr *serial.Reader, baseDir string) error {
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
		case "DialogPage":
			// DialogPage elements contain UI form controls used by the .NET
			// designer. Deserialize them so the FRX loads without error, then
			// discard — the engine never renders dialog pages.
			dp := NewDialogPage()
			_ = dp.Deserialize(rdr)
		case "Styles":
			deserializeStyles(rdr, r.Styles())
		case "Dictionary":
			deserializeDictionary(rdr, r.Dictionary(), baseDir)
		case "ScriptText":
			r.ScriptText = rdr.ReadInnerText()
		default:
			// Unknown top-level child — skip.
			_ = rdr.SkipElement()
		}
		_ = rdr.FinishChild()
	}
	return nil
}

// loadInherited handles an <inherited> root element. It reads the BaseReport
// attribute to locate the base FRX file, loads it, then deserializes the
// inherited overlay on top via ApplyBase.
func (r *Report) loadInherited(rdr *serial.Reader, baseDir string) error {
	// Read BaseReport attribute from the <inherited> root element.
	baseReportAttr := rdr.ReadStr("BaseReport", "")

	// Also read any other top-level report attributes present on the root.
	_ = r.Deserialize(rdr)

	if baseReportAttr == "" {
		return fmt.Errorf("report.LoadFrom: <inherited> root element has no BaseReport attribute")
	}

	// Resolve the base report path relative to the directory of the current file.
	basePath := baseReportAttr
	if baseDir != "" && !filepath.IsAbs(baseReportAttr) {
		basePath = filepath.Join(baseDir, baseReportAttr)
	}
	if basePath == "" || baseDir == "" {
		return fmt.Errorf("report.LoadFrom: cannot resolve base report path %q (no directory context; use Load(filename) instead of LoadFrom)", baseReportAttr)
	}

	// Load the base report. If the caller registered an OnLoadBaseReport hook,
	// delegate to it — this mirrors the C# Report.LoadBaseReport event which
	// allows the host application to intercept inherited-report loading.
	// C# ref: FastReport.Base/Report.cs, set_FileName → LoadBaseReport event.
	var base *Report
	if r.OnLoadBaseReport != nil {
		loaded, err := r.OnLoadBaseReport(basePath, r)
		if err != nil {
			return fmt.Errorf("report.LoadFrom: load base report %q: %w", basePath, err)
		}
		base = loaded
	} else {
		base = NewReport()
		if err := base.Load(basePath); err != nil {
			return fmt.Errorf("report.LoadFrom: load base report %q: %w", basePath, err)
		}
	}

	// Deserialize the inherited file's child elements (pages, dictionary, etc.)
	for {
		childType, ok := rdr.NextChild()
		if !ok {
			break
		}
		switch childType {
		case "inherited":
			// An <inherited Name="PageName"> child represents an overlay of a base page.
			pg, err := deserializeInheritedPage(rdr)
			if err != nil {
				_ = rdr.FinishChild()
				return fmt.Errorf("report.LoadFrom: deserialize inherited page: %w", err)
			}
			r.AddPage(pg)
		case "ReportPage":
			pg, err := deserializePage(rdr)
			if err != nil {
				_ = rdr.FinishChild()
				return fmt.Errorf("report.LoadFrom: deserialize page: %w", err)
			}
			r.AddPage(pg)
		case "DialogPage":
			// Deserialize and discard — dialog pages are not rendered.
			dp := NewDialogPage()
			_ = dp.Deserialize(rdr)
		case "Styles":
			deserializeStyles(rdr, r.Styles())
		case "Dictionary":
			deserializeDictionary(rdr, r.Dictionary(), baseDir)
		default:
			_ = rdr.SkipElement()
		}
		_ = rdr.FinishChild()
	}

	// Merge base report into this (child) report.
	r.ApplyBase(base)
	return nil
}

// deserializeInheritedPage reads an <inherited Name="…"> child element that
// represents a page overlay in an inherited report. It captures any new bands
// added in the inherited page; <inherited> sub-elements (references to base
// objects) are skipped since the merge is handled by ApplyBase at report level.
func deserializeInheritedPage(rdr *serial.Reader) (*ReportPage, error) {
	pg := NewReportPage()
	if err := pg.Deserialize(rdr); err != nil {
		return nil, err
	}
	pg.SetInherited(true)

	for {
		childType, ok := rdr.NextChild()
		if !ok {
			break
		}
		// "inherited" sub-elements reference base objects; skip their sub-tree.
		if childType == "inherited" {
			_ = rdr.SkipElement()
			_ = rdr.FinishChild()
			continue
		}
		obj, err := serial.DefaultRegistry.Create(childType)
		if err != nil {
			_ = rdr.SkipElement()
			_ = rdr.FinishChild()
			continue
		}
		if err2 := obj.Deserialize(rdr); err2 != nil {
			_ = rdr.FinishChild()
			return nil, fmt.Errorf("deserialize %s: %w", childType, err2)
		}
		deserializeChildren(rdr, obj)
		_ = rdr.FinishChild()
		pg.AddBandByTypeName(childType, obj)
	}
	return pg, nil
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
	// Prefer AddChild (from report.Parent) when available. AddChild is aware of
	// special children like ChildBand (sets b.child) vs regular objects
	// (appended to b.objects).  Direct Objects().Add() would bypass this
	// routing and always append to the objects slice.
	parentAdder, hasAddChild := parent.(report.Parent)

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

		if hasAddChild {
			parentAdder.AddChild(child)
		} else if isContainer {
			container.Objects().Add(child)
		}
	}
}


// deserializeDictionary reads the <Dictionary> element and populates the
// report dictionary with Parameters, Relations, Totals, and data connections.
// baseDir is the directory of the FRX file; it is used to resolve relative
// file paths for CSV (and similar) data connections.
func deserializeDictionary(rdr *serial.Reader, dict *data.Dictionary, baseDir string) {
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
			// Also register as an AggregateTotal so the engine accumulates per-row.
			at := &data.AggregateTotal{
				Name:                 t.Name,
				TotalType:            t.TotalType,
				Expression:           t.Expression,
				Evaluator:            t.Evaluator,
				PrintOn:              t.PrintOn,
				ResetAfterPrint:      t.ResetAfterPrint,
				ResetOnReprint:       t.ResetOnReprint,
				EvaluateCondition:    t.EvaluateCondition,
				IncludeInvisibleRows: t.IncludeInvisibleRows,
			}
			dict.AddAggregateTotal(at)
		case "JsonDataConnection":
			conn, sources := deserializeJsonConnection(rdr)
			dict.AddConnection(conn)
			for _, ds := range sources {
				dict.AddDataSource(ds)
			}
		case "CsvDataConnection":
			conn, sources := deserializeCsvConnection(rdr, baseDir)
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
		case "BusinessObjectDataSource":
			// BusinessObjectDataSource elements declare data shapes for Go business
			// objects. Register each one as an uninitialised BusinessObjectDataSource
			// so the engine can resolve the alias at run time. Actual data must be
			// bound via report.Dictionary().RegisterData (or equivalent) before running.
			sources := deserializeBusinessObjectDataSource(rdr)
			for _, ds := range sources {
				dict.AddDataSource(ds)
			}
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

// deserializeCsvConnection reads a <CsvDataConnection> element and returns
// the connection stub plus any nested CSVDataSource instances it exposes.
//
// FRX format for CsvDataConnection:
//
//	<CsvDataConnection Name="Connection">
//	  <TableDataSource Name="ds_name" TableName="file.csv" StoreData="true"
//	                   TableData="<base64 .NET DataSet XML>" ...>
//	    <Column Name="col1" DataType="..."/>
//	    ...
//	  </TableDataSource>
//	</CsvDataConnection>
//
// When StoreData is true and TableData is non-empty the data is loaded from
// the inline base64-encoded .NET DataSet XML. Otherwise the TableName is
// treated as a CSV file path relative to baseDir (the FRX file directory).
func deserializeCsvConnection(rdr *serial.Reader, baseDir string) (*data.DataConnectionBase, []data.DataSource) {
	name := rdr.ReadStr("Name", "CsvConnection")
	connStr := rdr.ReadStr("ConnectionString", "")
	enabled := rdr.ReadBool("Enabled", true)

	conn := data.NewDataConnectionBase("csv")
	conn.SetName(name)
	conn.SetEnabled(enabled)
	conn.ConnectionString = connStr

	var sources []data.DataSource
	for {
		ct, ok := rdr.NextChild()
		if !ok {
			break
		}
		if ct == "TableDataSource" {
			dsName := rdr.ReadStr("Name", "")
			alias := rdr.ReadStr("Alias", dsName)
			csvFileName := rdr.ReadStr("TableName", "")
			storeData := rdr.ReadBool("StoreData", false)
			tableDataB64 := rdr.ReadStr("TableData", "")

			// Drain child elements (Column entries, etc.).
			for {
				_, ok2 := rdr.NextChild()
				if !ok2 {
					break
				}
				_ = rdr.FinishChild()
			}

			// Prefer inline data when StoreData is true and TableData is present.
			if storeData && tableDataB64 != "" {
				if ds := parseCsvInlineDataSet(dsName, alias, csvFileName, tableDataB64); ds != nil {
					sources = append(sources, ds)
					continue
				}
			}

			// Fall back to file-based CSV.
			// csvFileName is relative to the FRX directory when baseDir is set.
			filePath := csvFileName
			if filePath != "" && baseDir != "" && !filepath.IsAbs(filePath) {
				filePath = filepath.Join(baseDir, filePath)
			}
			ds := csvdata.New(dsName)
			ds.SetAlias(alias)
			ds.SetFilePath(filePath)
			sources = append(sources, ds)
		} else {
			_ = rdr.SkipElement()
		}
		_ = rdr.FinishChild()
	}
	return conn, sources
}

// parseCsvInlineDataSet decodes a base64-encoded .NET XML DataSet embedded in
// the TableData attribute and returns a populated BaseDataSource.
//
// The .NET DataSet XML has the form:
//
//	<NewDataSet>
//	  <xs:schema>...</xs:schema>
//	  <tableName><col1>val</col1><col2>val</col2></tableName>
//	  ...
//	</NewDataSet>
//
// where tableName matches the TableName attribute of the enclosing
// TableDataSource (e.g. "state_table.csv").
func parseCsvInlineDataSet(dsName, alias, tableName, tableDataB64 string) data.DataSource {
	rawXML, err := base64.StdEncoding.DecodeString(tableDataB64)
	if err != nil {
		return nil
	}

	dec := xml.NewDecoder(bytes.NewReader(rawXML))

	// Skip to root element (<NewDataSet> or similar).
	var root xml.StartElement
	for {
		tok, e := dec.Token()
		if e != nil {
			return nil
		}
		if se, ok := tok.(xml.StartElement); ok {
			root = se
			_ = root
			break
		}
	}

	// Read row elements whose local name matches tableName.
	// Skip any xs:schema or other non-row elements.
	rowElemName := tableName

	var colOrder []string
	colSet := make(map[string]bool)
	var rows []map[string]any

	for {
		tok, e := dec.Token()
		if e != nil {
			break
		}
		se, ok := tok.(xml.StartElement)
		if !ok {
			continue
		}
		// Skip xs:* or other namespace-qualified elements (e.g. xs:schema).
		if se.Name.Space != "" {
			if skipErr := dec.Skip(); skipErr != nil {
				break
			}
			continue
		}
		if se.Name.Local != rowElemName {
			// Unexpected sibling — skip.
			if skipErr := dec.Skip(); skipErr != nil {
				break
			}
			continue
		}

		// Parse one row: each immediate child element's text is a column value.
		row := make(map[string]any)
		for {
			inner, e2 := dec.Token()
			if e2 != nil {
				break
			}
			switch it := inner.(type) {
			case xml.StartElement:
				col := it.Name.Local
				var textBuf strings.Builder
				for {
					t2, e3 := dec.Token()
					if e3 != nil {
						break
					}
					switch tv := t2.(type) {
					case xml.CharData:
						textBuf.Write(tv)
					case xml.EndElement:
						goto colDone
					}
				}
			colDone:
				val := strings.TrimSpace(textBuf.String())
				row[col] = val
				if !colSet[col] {
					colSet[col] = true
					colOrder = append(colOrder, col)
				}
			case xml.EndElement:
				// End of row element.
				goto rowDone
			}
		}
	rowDone:
		rows = append(rows, row)
	}

	// Build a BaseDataSource from the parsed rows.
	bs := data.NewBaseDataSource(dsName)
	bs.SetAlias(alias)
	for _, col := range colOrder {
		bs.AddColumn(data.Column{Name: col, Alias: col, DataType: "string"})
	}
	for _, row := range rows {
		bs.AddRow(row)
	}
	if err2 := bs.Init(); err2 != nil {
		return nil
	}
	return bs
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

// deserializeBusinessObjectDataSource reads a <BusinessObjectDataSource> element
// and recursively collects all nested <BusinessObjectDataSource> elements.
// Each data source is created with nil data — the caller must bind actual Go
// data via report.Dictionary().RegisterData before running the report.
// Returns a flat list of all data sources (parent first, then nested children).
func deserializeBusinessObjectDataSource(rdr *serial.Reader) []data.DataSource {
	name := rdr.ReadStr("Name", "")
	alias := rdr.ReadStr("Alias", name)

	// Create a BusinessObjectDataSource stub with no data.
	// Init() is NOT called here — it will be called by the engine (or the user
	// can call RegisterData which replaces the entry with a populated source).
	ds := data.NewBusinessObjectDataSource(name, nil)
	ds.SetAlias(alias)
	_ = ds.Init() // Initialise with nil data so RowCount() == 0 and First() returns ErrEOF.

	sources := []data.DataSource{ds}

	// Read child elements: Column entries (metadata only, ignored at runtime since
	// Go reflection is used to derive columns) and nested BusinessObjectDataSources.
	for {
		ct, ok := rdr.NextChild()
		if !ok {
			break
		}
		if ct == "BusinessObjectDataSource" {
			nested := deserializeBusinessObjectDataSource(rdr)
			sources = append(sources, nested...)
			_ = rdr.FinishChild()
			continue
		}
		// Column, or any other unknown child — drain and skip.
		for {
			_, ok2 := rdr.NextChild()
			if !ok2 {
				break
			}
			_ = rdr.FinishChild()
		}
		_ = rdr.FinishChild()
	}
	return sources
}

// deserializeParameter reads a single <Parameter> element (and nested children).
func deserializeParameter(rdr *serial.Reader) *data.Parameter {
	p := &data.Parameter{
		Name:        rdr.ReadStr("Name", ""),
		DataType:    rdr.ReadStr("DataType", ""),
		Expression:  rdr.ReadStr("Expression", ""),
		Description: rdr.ReadStr("Description", ""),
	}
	// AsString is the persistent string value when Expression is empty.
	// C# ref: FastReport.Data.Parameter.Serialize (Parameter.cs:192-193)
	if asStr := rdr.ReadStr("AsString", ""); asStr != "" && p.Expression == "" {
		p.Value = asStr
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
	// C# Total constructor: resetAfterPrint = true (Total.cs constructor).
	// C# Serialize only writes "ResetAfterPrint" when it differs from the default (true).
	// Therefore: absent attribute → true; "ResetAfterPrint=false" → false.
	// ResetOnReprint default is also true; only written when false.
	t := &data.Total{
		Name:                 rdr.ReadStr("Name", ""),
		Expression:           rdr.ReadStr("Expression", ""),
		TotalType:            parseTotalType(rdr.ReadStr("TotalType", "Sum")),
		Evaluator:            rdr.ReadStr("Evaluator", ""),
		PrintOn:              rdr.ReadStr("PrintOn", ""),
		ResetAfterPrint:      rdr.ReadBool("ResetAfterPrint", true),  // C# default is true
		ResetOnReprint:       rdr.ReadBool("ResetOnReprint", true),   // C# default is true
		EvaluateCondition:    rdr.ReadStr("EvaluateCondition", ""),
		IncludeInvisibleRows: rdr.ReadBool("IncludeInvisibleRows", false),
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

	// Fill: use DeserializeFill so gradient/hatch fills in <Style> elements
	// are loaded correctly, matching C# StyleBase.Deserialize (StyleBase.cs
	// line 143). For SolidFill results we also populate the legacy FillColor
	// scalar for backward compatibility.
	e.Fill = report.DeserializeFill(rdr, "Fill", &style.SolidFill{})
	if sf, ok := e.Fill.(*style.SolidFill); ok {
		e.FillColor = sf.Color
		e.Fill = nil // keep legacy path for solid fills
	}
	// TextFill: same pattern for text colour.
	// C# StyleBase constructor initialises TextFill to SolidFill(Color.Black),
	// so an absent TextFill.Color in the FRX means "apply black" (not transparent).
	// Ref: StyleBase.cs line 109 — TextFill = new SolidFill(Color.Black).
	e.TextFill = report.DeserializeFill(rdr, "TextFill", style.NewSolidFill(style.ColorBlack))
	if sf, ok := e.TextFill.(*style.SolidFill); ok {
		e.TextColor = sf.Color
		e.TextFill = nil // keep legacy path for solid text fills
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
