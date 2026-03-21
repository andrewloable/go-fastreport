package data

import (
	"database/sql"
	"fmt"

	"github.com/andrewloable/go-fastreport/report"
)

// -----------------------------------------------------------------------
// CommandParameter
// -----------------------------------------------------------------------

// ParameterDirection specifies whether a command parameter is input, output, or both.
type ParameterDirection int

const (
	ParamDirectionInput       ParameterDirection = iota
	ParamDirectionOutput
	ParamDirectionInputOutput
	ParamDirectionReturnValue
)

// CommandParameter represents a parameter passed to a SQL command.
// It is the Go equivalent of FastReport.Data.CommandParameter.
type CommandParameter struct {
	// Name is the parameter name (e.g. "@customerId").
	Name string
	// DataType is the SQL data type hint string (e.g. "int", "varchar").
	DataType string
	// Size is the maximum data size (e.g. for varchar columns).
	Size int
	// Expression is a report expression whose result is used as the value.
	Expression string
	// DefaultValue is used when Expression evaluates to nil.
	DefaultValue string
	// Direction indicates input/output direction.
	Direction ParameterDirection
	// Value holds the resolved run-time value.
	Value any
}

// NewCommandParameter creates a CommandParameter with default Input direction.
func NewCommandParameter(name string) *CommandParameter {
	return &CommandParameter{Name: name, Direction: ParamDirectionInput}
}

// Serialize writes CommandParameter properties to w.
func (p *CommandParameter) Serialize(w report.Writer) error {
	if p.Name != "" {
		w.WriteStr("Name", p.Name)
	}
	if p.DataType != "" {
		w.WriteStr("DataType", p.DataType)
	}
	if p.Size != 0 {
		w.WriteInt("Size", p.Size)
	}
	if p.Expression != "" {
		w.WriteStr("Expression", p.Expression)
	}
	if p.DefaultValue != "" {
		w.WriteStr("DefaultValue", p.DefaultValue)
	}
	if p.Direction != ParamDirectionInput {
		w.WriteInt("Direction", int(p.Direction))
	}
	return nil
}

// Deserialize reads CommandParameter properties from r.
func (p *CommandParameter) Deserialize(r report.Reader) error {
	p.Name = r.ReadStr("Name", "")
	p.DataType = r.ReadStr("DataType", "")
	p.Size = r.ReadInt("Size", 0)
	p.Expression = r.ReadStr("Expression", "")
	p.DefaultValue = r.ReadStr("DefaultValue", "")
	p.Direction = ParameterDirection(r.ReadInt("Direction", int(ParamDirectionInput)))
	return nil
}

// -----------------------------------------------------------------------
// DataConnectionBase
// -----------------------------------------------------------------------

// DataConnectionBase is the abstract base for all database connections.
// It is the Go equivalent of FastReport.Data.DataConnectionBase.
//
// Concrete connection types (Postgres, MySQL, SQLite …) embed this struct
// and provide a sql.DB via the Open() method.
type DataConnectionBase struct {
	DataComponentBase

	// ConnectionString is the DSN or connection string.
	ConnectionString string
	// ConnectionStringExpression is a report expression that resolves to the DSN.
	ConnectionStringExpression string
	// LoginPrompt causes the engine to prompt for credentials at run time.
	LoginPrompt bool
	// CommandTimeout is the per-query timeout in seconds (0 = driver default).
	CommandTimeout int

	// db is the underlying *sql.DB once opened.
	db *sql.DB
	// tables is the set of known TableDataSources loaded from this connection.
	tables []*TableDataSource
	// driverName is the database/sql driver name (e.g. "postgres", "sqlite3").
	driverName string
}

// NewDataConnectionBase creates a DataConnectionBase for the given sql driver.
func NewDataConnectionBase(driverName string) *DataConnectionBase {
	return &DataConnectionBase{
		DataComponentBase: *NewDataComponentBase(""),
		driverName:        driverName,
	}
}

// DriverName returns the database/sql driver name.
func (c *DataConnectionBase) DriverName() string { return c.driverName }

// Open opens the underlying *sql.DB using the ConnectionString.
// Returns an error if the connection cannot be established.
func (c *DataConnectionBase) Open() error {
	if c.db != nil {
		return nil // already open
	}
	db, err := sql.Open(c.driverName, c.ConnectionString)
	if err != nil {
		return fmt.Errorf("DataConnection(%s): %w", c.driverName, err)
	}
	c.db = db
	return nil
}

// Close closes the underlying *sql.DB.
func (c *DataConnectionBase) Close() error {
	if c.db == nil {
		return nil
	}
	err := c.db.Close()
	c.db = nil
	return err
}

// DB returns the open *sql.DB, or nil if not yet opened.
func (c *DataConnectionBase) DB() *sql.DB { return c.db }

// Tables returns the TableDataSources registered with this connection.
func (c *DataConnectionBase) Tables() []*TableDataSource { return c.tables }

// AddTable registers a TableDataSource with this connection.
func (c *DataConnectionBase) AddTable(t *TableDataSource) {
	t.connection = c
	c.tables = append(c.tables, t)
}

// CreateTable creates a TableDataSource with the given name and registers it.
func (c *DataConnectionBase) CreateTable(name string) *TableDataSource {
	t := NewTableDataSource(name)
	c.AddTable(t)
	return t
}

// -----------------------------------------------------------------------
// TableDataSource
// -----------------------------------------------------------------------

// TableDataSource is a data source backed by a SQL query executed against a
// DataConnectionBase. It is the Go equivalent of FastReport.Data.TableDataSource.
type TableDataSource struct {
	BaseDataSource

	// tableName is the underlying table/view name.
	tableName string
	// selectCommand is the full SQL SELECT statement.
	selectCommand string
	// parameters are the bound SQL command parameters.
	parameters []*CommandParameter
	// storeData caches results so they can be replayed without re-querying.
	storeData bool
	// connection is the owning DataConnectionBase.
	connection *DataConnectionBase
	// ignoreConnection, when true, causes the data source to use externally-provided
	// data instead of querying the database through the connection.
	// C# ref: FastReport.Data.TableDataSource.ignoreConnection
	ignoreConnection bool
	// forceLoadData forces data reload even if already cached.
	// C# ref: FastReport.Data.DataSourceBase.ForceLoadData
	forceLoadData bool
}

// NewTableDataSource creates a TableDataSource with the given name.
func NewTableDataSource(name string) *TableDataSource {
	return &TableDataSource{
		BaseDataSource: BaseDataSource{name: name, alias: name},
	}
}

// TableName returns the source table or view name.
func (t *TableDataSource) TableName() string { return t.tableName }

// SetTableName sets the table name.
func (t *TableDataSource) SetTableName(s string) { t.tableName = s }

// SelectCommand returns the SQL SELECT statement.
func (t *TableDataSource) SelectCommand() string { return t.selectCommand }

// SetSelectCommand sets the SQL SELECT statement.
func (t *TableDataSource) SetSelectCommand(s string) { t.selectCommand = s }

// Parameters returns the command parameters.
func (t *TableDataSource) Parameters() []*CommandParameter { return t.parameters }

// AddParameter appends a command parameter.
func (t *TableDataSource) AddParameter(p *CommandParameter) {
	t.parameters = append(t.parameters, p)
}

// StoreData returns whether query results are cached for replay.
func (t *TableDataSource) StoreData() bool { return t.storeData }

// SetStoreData sets the store-data flag.
func (t *TableDataSource) SetStoreData(v bool) { t.storeData = v }

// IgnoreConnection returns whether the data source ignores its connection and
// uses externally-provided data instead of querying the database.
// C# ref: FastReport.Data.TableDataSource.IgnoreConnection
func (t *TableDataSource) IgnoreConnection() bool { return t.ignoreConnection }

// SetIgnoreConnection sets the ignore-connection flag.
func (t *TableDataSource) SetIgnoreConnection(v bool) { t.ignoreConnection = v }

// ForceLoadData returns whether data is force-reloaded even if already cached.
// C# ref: FastReport.Data.DataSourceBase.ForceLoadData
func (t *TableDataSource) ForceLoadData() bool { return t.forceLoadData }

// SetForceLoadData sets the force-load-data flag.
func (t *TableDataSource) SetForceLoadData(v bool) { t.forceLoadData = v }

// Connection returns the owning DataConnectionBase.
// When IgnoreConnection is true, Connection returns nil (matching C# behaviour).
func (t *TableDataSource) Connection() *DataConnectionBase {
	if t.ignoreConnection {
		return nil
	}
	return t.connection
}

// InitSchema discovers the table schema without loading data.
// It executes the select command (or a SELECT * FROM tableName) wrapped in a
// LIMIT 0 subquery to obtain column metadata, and populates t.columns.
// C# ref: FastReport.Data.TableDataSource.InitSchema()
func (t *TableDataSource) InitSchema() error {
	if t.connection == nil || t.connection.DB() == nil {
		return nil // no connection, skip schema discovery
	}
	query := t.selectCommand
	if query == "" {
		if t.tableName == "" {
			return nil
		}
		query = "SELECT * FROM " + t.tableName
	}
	// Use LIMIT 0 to get column metadata without loading data.
	schemaQuery := "SELECT * FROM (" + query + ") AS _schema LIMIT 0"
	rows, err := t.connection.DB().Query(schemaQuery)
	if err != nil {
		// Fallback: try the original query.
		rows, err = t.connection.DB().Query(query)
		if err != nil {
			return nil // silently skip on error
		}
	}
	defer rows.Close()
	colNames, err := rows.Columns()
	if err != nil {
		return nil
	}
	t.columns = make([]Column, len(colNames))
	for i, c := range colNames {
		t.columns[i] = Column{Name: c, Alias: c}
	}
	return nil
}

// LoadData loads data into the data source. It delegates to Init().
// C# ref: FastReport.Data.TableDataSource.LoadData()
func (t *TableDataSource) LoadData() error {
	return t.Init()
}

// Init executes the SELECT command and loads results into the in-memory row store.
// When ignoreConnection is true, Init returns nil immediately (data was provided
// externally). When storeData is true and data is already loaded (initialized is
// true and forceLoadData is false), Init skips re-querying.
func (t *TableDataSource) Init() error {
	// If ignoreConnection is set, data is provided externally — skip query.
	if t.ignoreConnection {
		t.initialized = true
		return nil
	}

	// If storeData is true and we already loaded data, skip re-query unless
	// forceLoadData is set.
	if t.storeData && t.initialized && !t.forceLoadData {
		t.currentRow = 0
		return nil
	}

	if t.connection == nil {
		return fmt.Errorf("TableDataSource %q: no connection set", t.name)
	}
	if t.connection.DB() == nil {
		if err := t.connection.Open(); err != nil {
			return err
		}
	}

	query := t.selectCommand
	if query == "" {
		if t.tableName == "" {
			return fmt.Errorf("TableDataSource %q: no SelectCommand or TableName", t.name)
		}
		query = "SELECT * FROM " + t.tableName
	}

	// Build argument list from resolved CommandParameter values.
	args := make([]any, len(t.parameters))
	for i, p := range t.parameters {
		args[i] = p.Value
	}

	rows, err := t.connection.DB().Query(query, args...)
	if err != nil {
		return fmt.Errorf("TableDataSource %q: %w", t.name, err)
	}
	defer rows.Close()

	colNames, err := rows.Columns()
	if err != nil {
		return fmt.Errorf("TableDataSource %q columns: %w", t.name, err)
	}

	// Build Column metadata.
	t.columns = make([]Column, len(colNames))
	for i, c := range colNames {
		t.columns[i] = Column{Name: c, Alias: c}
	}

	// Scan all rows into the in-memory store.
	t.rows = nil
	scanBuf := make([]any, len(colNames))
	scanPtrs := make([]any, len(colNames))
	for i := range scanBuf {
		scanPtrs[i] = &scanBuf[i]
	}
	for rows.Next() {
		if err := rows.Scan(scanPtrs...); err != nil {
			return fmt.Errorf("TableDataSource %q scan: %w", t.name, err)
		}
		row := make(map[string]any, len(colNames))
		for i, col := range colNames {
			row[col] = scanBuf[i]
		}
		t.rows = append(t.rows, row)
	}
	if err := rows.Err(); err != nil {
		return fmt.Errorf("TableDataSource %q rows: %w", t.name, err)
	}
	t.currentRow = 0
	t.initialized = true
	return nil
}
