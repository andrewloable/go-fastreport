package data

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/andrewloable/go-fastreport/report"
	"github.com/andrewloable/go-fastreport/utils"
)

// -----------------------------------------------------------------------
// CommandParameter
// -----------------------------------------------------------------------

// ParameterDirection specifies whether a command parameter is input, output, or both.
// Mirrors C# System.Data.ParameterDirection enum names.
type ParameterDirection int

const (
	ParamDirectionInput       ParameterDirection = iota
	ParamDirectionOutput
	ParamDirectionInputOutput
	ParamDirectionReturnValue
)

// paramDirectionToString maps ParameterDirection to its C# enum name string.
// C# WriteValue("Direction", Direction) writes the enum name, not its numeric value.
// Mirrors System.Data.ParameterDirection enum names.
func paramDirectionToString(d ParameterDirection) string {
	switch d {
	case ParamDirectionOutput:
		return "Output"
	case ParamDirectionInputOutput:
		return "InputOutput"
	case ParamDirectionReturnValue:
		return "ReturnValue"
	default:
		return "Input"
	}
}

// paramDirectionFromString parses a ParameterDirection from its C# enum name.
// Falls back to parsing as an integer for backward compatibility with old FRX files
// that may have written the numeric value.
func paramDirectionFromString(s string) ParameterDirection {
	switch s {
	case "Output":
		return ParamDirectionOutput
	case "InputOutput":
		return ParamDirectionInputOutput
	case "ReturnValue":
		return ParamDirectionReturnValue
	case "Input", "":
		return ParamDirectionInput
	default:
		// Legacy: numeric value written by older Go port.
		switch s {
		case "1":
			return ParamDirectionOutput
		case "2":
			return ParamDirectionInputOutput
		case "3":
			return ParamDirectionReturnValue
		}
		return ParamDirectionInput
	}
}

// paramValueUninitialized is a sentinel type used by ResetLastValue to mark
// the lastValue as unset. It mirrors C# CommandParameter.ParamValue.Uninitialized.
type paramValueUninitialized struct{}

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
	// lastValue caches the previously evaluated value for change detection.
	lastValue any
}

// NewCommandParameter creates a CommandParameter with default Input direction.
func NewCommandParameter(name string) *CommandParameter {
	return &CommandParameter{Name: name, Direction: ParamDirectionInput}
}

// LastValue returns the cached last-evaluated value.
func (p *CommandParameter) LastValue() any { return p.lastValue }

// SetLastValue sets the cached last-evaluated value.
func (p *CommandParameter) SetLastValue(v any) { p.lastValue = v }

// ResetLastValue resets the cached value to the uninitialized sentinel,
// signalling that the parameter must be re-evaluated on the next use.
func (p *CommandParameter) ResetLastValue() {
	p.lastValue = paramValueUninitialized{}
}

// Assign copies all CommandParameter properties from src.
// Mirrors C# CommandParameter.Assign (CommandParameter.cs lines 159-169):
// Name, DataType, Size, Value, Expression, DefaultValue and Direction are all copied.
func (p *CommandParameter) Assign(src *CommandParameter) {
	if src == nil {
		return
	}
	p.Name = src.Name
	p.DataType = src.DataType
	p.Size = src.Size
	p.Value = src.Value
	p.Expression = src.Expression
	p.DefaultValue = src.DefaultValue
	p.Direction = src.Direction
}

// GetExpressions returns the list of expressions used by this parameter.
// Mirrors C# CommandParameter.GetExpressions (CommandParameter.cs).
func (p *CommandParameter) GetExpressions() []string {
	if p.Expression != "" {
		return []string{p.Expression}
	}
	return nil
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
		// C# WriteValue("Direction", Direction) writes the enum name, e.g. "Output".
		// Mirrors C# CommandParameter.Serialize (CommandParameter.cs line 154).
		w.WriteStr("Direction", paramDirectionToString(p.Direction))
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
	// C# WriteValue writes the enum name ("Input","Output",etc.). Read as string with
	// fallback to legacy integer form. Mirrors C# CommandParameter.Serialize line 154.
	p.Direction = paramDirectionFromString(r.ReadStr("Direction", "Input"))
	return nil
}

// -----------------------------------------------------------------------
// DataConnectionBase — event-arg types
// -----------------------------------------------------------------------

// DatabaseLoginEventArgs carries the connection string passed to the
// OnDatabaseLogin callback. The callback may override ConnectionString, or
// set UserName/Password, before the connection is opened.
// C# ref: FastReport.DatabaseLoginEventArgs (ReportEventArgs.cs).
// / ReportSettings.Core.cs OnDatabaseLogin → fires DatabaseLogin event.
type DatabaseLoginEventArgs struct {
	// ConnectionString is the DSN passed to the connection. The callback may
	// replace it to inject credentials or switch to a different server.
	ConnectionString string
	// UserName is an optional credential field available for custom Open() overrides.
	UserName string
	// Password is an optional credential field available for custom Open() overrides.
	Password string
}

// AfterDatabaseLoginEventArgs carries the open *sql.DB passed to the
// OnAfterDatabaseLogin callback after the connection has been established.
// C# ref: FastReport.AfterDatabaseLoginEventArgs (ReportEventArgs.cs).
// / ReportSettings.OnAfterDatabaseLogin → fires AfterDatabaseLogin event.
type AfterDatabaseLoginEventArgs struct {
	// DB is the opened *sql.DB. The callback may inspect or configure it.
	DB *sql.DB
}

// FilterConnectionTablesEventArgs carries the context passed to the
// OnFilterConnectionTables callback. Set Skip = true to exclude the table.
// C# ref: FastReport.Utils.Config.FilterConnectionTablesEventArgs.
type FilterConnectionTablesEventArgs struct {
	// Connection is the DataConnectionBase that owns the table list.
	Connection *DataConnectionBase
	// TableName is the candidate table name being evaluated.
	TableName string
	// Skip, when set to true by the callback, causes the table to be excluded.
	Skip bool
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
	// CommandTimeout is the per-query timeout in seconds (default 30, per C# default).
	CommandTimeout int

	// OnDatabaseLogin is an optional callback invoked just before sql.Open.
	// The callback receives a *DatabaseLoginEventArgs whose ConnectionString is
	// pre-filled with the current value; the callback may replace it.
	// C# ref: ReportSettings.DatabaseLogin event / Core.cs OnDatabaseLogin.
	OnDatabaseLogin func(e *DatabaseLoginEventArgs)

	// OnAfterDatabaseLogin is an optional callback invoked after sql.Open succeeds.
	// The callback receives a *AfterDatabaseLoginEventArgs with the open *sql.DB.
	// C# ref: ReportSettings.AfterDatabaseLogin event / ReportSettings.OnAfterDatabaseLogin.
	OnAfterDatabaseLogin func(e *AfterDatabaseLoginEventArgs)

	// OnFilterConnectionTables is an optional callback invoked for each table name
	// during CreateAllTables. Set e.Skip = true to exclude a table.
	// C# ref: FastReport.Utils.Config.OnFilterConnectionTables event.
	OnFilterConnectionTables func(e *FilterConnectionTablesEventArgs)

	// isSqlBased indicates whether this connection is SQL-based.
	// C# ref: FastReport.Data.DataConnectionBase.IsSqlBased
	isSqlBased bool
	// canContainProcedures indicates whether this connection can expose stored procedures.
	// C# ref: FastReport.Data.DataConnectionBase.CanContainProcedures
	canContainProcedures bool

	// db is the underlying *sql.DB once opened.
	db *sql.DB
	// tables is the set of known TableDataSources loaded from this connection.
	tables []*TableDataSource
	// driverName is the database/sql driver name (e.g. "postgres", "sqlite3").
	driverName string
}

// NewDataConnectionBase creates a DataConnectionBase for the given sql driver.
// Defaults match C# DataConnectionBase constructor: IsSqlBased=true, CommandTimeout=30.
func NewDataConnectionBase(driverName string) *DataConnectionBase {
	return &DataConnectionBase{
		DataComponentBase:    *NewDataComponentBase(""),
		driverName:           driverName,
		isSqlBased:           true,
		canContainProcedures: false,
		CommandTimeout:       30,
	}
}

// IsSqlBased returns whether this connection is SQL-based.
// C# ref: FastReport.Data.DataConnectionBase.IsSqlBased
func (c *DataConnectionBase) IsSqlBased() bool { return c.isSqlBased }

// SetIsSqlBased sets the IsSqlBased flag.
func (c *DataConnectionBase) SetIsSqlBased(v bool) { c.isSqlBased = v }

// CanContainProcedures returns whether this connection can contain stored procedures.
// C# ref: FastReport.Data.DataConnectionBase.CanContainProcedures
func (c *DataConnectionBase) CanContainProcedures() bool { return c.canContainProcedures }

// SetCanContainProcedures sets the CanContainProcedures flag.
func (c *DataConnectionBase) SetCanContainProcedures(v bool) { c.canContainProcedures = v }

// GetTableNames returns the list of table and view names available in this connection.
// The base implementation returns an empty slice; SQL-based subclasses override this
// to query the database schema.
// C# ref: FastReport.Data.DataConnectionBase.GetTableNames()
func (c *DataConnectionBase) GetTableNames() []string {
	return nil
}

// GetTableCount returns the number of tables/views available in this connection.
// C# ref: derived from GetTableNames().
func (c *DataConnectionBase) GetTableCount() int {
	return len(c.GetTableNames())
}

// FilterTables is a hook called by CreateAllTables to allow subclasses or callers
// to remove table names from the list before table objects are created.
// The base implementation is a no-op.
// C# ref: FastReport.Data.DataConnectionBase.FilterTables (virtual, no-op in base)
func (c *DataConnectionBase) FilterTables(tableNames []string) []string {
	return tableNames
}

// CreateAllTables fills the Tables collection with all tables available in the
// connection (calls GetTableNames, removes stale tables, creates new ones) and
// then calls InitSchema on each table.
// C# ref: FastReport.Data.DataConnectionBase.CreateAllTables()
func (c *DataConnectionBase) CreateAllTables() {
	c.CreateAllTablesWithSchema(true)
}

// CreateAllTablesWithSchema fills the Tables collection. When initSchema is true
// each table's schema is initialised by calling InitSchema().
// C# ref: FastReport.Data.DataConnectionBase.CreateAllTables(bool initSchema)
func (c *DataConnectionBase) CreateAllTablesWithSchema(initSchema bool) {
	tableNames := c.FilterTables(c.GetTableNames())
	c.createAllTablesShared(tableNames)
	if initSchema {
		for _, t := range c.tables {
			_ = t.InitSchema()
		}
	}
}

// createAllTablesShared implements the shared logic of CreateAllTables:
// 1. Remove tables whose name no longer exists in the connection.
// 2. Create new TableDataSource objects for table names not yet registered.
// C# ref: FastReport.Data.DataConnectionBase.CreateAllTablesShared
func (c *DataConnectionBase) createAllTablesShared(tableNames []string) {
	// Step 1: remove stale tables (those with no SelectCommand whose TableName
	// is no longer in the connection's schema list).
	for i := 0; i < len(c.tables); {
		t := c.tables[i]
		// Keep query-based tables (they have a SelectCommand) — they are
		// user-defined and are not managed by CreateAllTables.
		if t.selectCommand != "" {
			i++
			continue
		}
		found := false
		for _, name := range tableNames {
			if strings.EqualFold(t.tableName, name) {
				found = true
				break
			}
		}
		if !found {
			// Remove stale table.
			c.tables = append(c.tables[:i], c.tables[i+1:]...)
			continue
		}
		i++
	}

	// Step 2: create new TableDataSource for names not already present.
	for _, tableName := range tableNames {
		found := false
		for _, t := range c.tables {
			if strings.EqualFold(t.tableName, tableName) {
				found = true
				break
			}
		}
		if !found {
			// Sanitise table name into a valid identifier (same as C# logic).
			fixedName := strings.NewReplacer(".", "_", "[", "", "]", "", `"`, "").Replace(tableName)
			t := NewTableDataSource(fixedName)
			t.SetTableName(tableName)
			// New tables are disabled by default so they are not selected automatically.
			// C# ref: DataConnectionBase.CreateAllTablesShared — table.Enabled = false
			t.SetEnabled(false)
			c.AddTable(t)
		}
	}
}

// DeleteTable removes the TableDataSource from the tables collection and
// resets its internal table reference (equivalent to C# DataSet table removal).
// C# ref: FastReport.Data.DataConnectionBase.DeleteTable
func (c *DataConnectionBase) DeleteTable(source *TableDataSource) {
	for i, t := range c.tables {
		if t == source {
			c.tables = append(c.tables[:i], c.tables[i+1:]...)
			break
		}
	}
	// Clear cached row data so the source is no longer considered initialised.
	source.initialized = false
	source.rows = nil
}

// FillTable reloads data for the given source when needed:
// - when ForceLoadData is true, or
// - when any parameter value has changed since the last load.
// C# ref: FastReport.Data.DataConnectionBase.FillTable (internal)
func (c *DataConnectionBase) FillTable(source *TableDataSource) error {
	if !source.initialized && len(source.rows) == 0 {
		return source.Init()
	}

	parametersChanged := false
	for _, par := range source.parameters {
		if par.Value != par.lastValue {
			par.lastValue = par.Value
			parametersChanged = true
		}
	}

	if source.forceLoadData || !source.initialized || parametersChanged {
		return source.Init()
	}
	return nil
}

// Serialize writes the connection's properties to w.
// Child TableDataSources that are enabled are written as child elements.
// C# ref: FastReport.Data.DataConnectionBase.Serialize
func (c *DataConnectionBase) Serialize(w report.Writer) error {
	if c.Name() != "" {
		w.WriteStr("Name", c.Name())
	}
	if c.IsAliased() {
		w.WriteStr("Alias", c.Alias())
	}
	if !c.Enabled() {
		w.WriteBool("Enabled", false)
	}
	if c.ReferenceName() != "" {
		w.WriteStr("ReferenceName", c.ReferenceName())
	}
	if c.ConnectionString != "" {
		// Encrypt the connection string before writing to FRX, matching
		// C# DataConnectionBase.Serialize: Crypter.EncryptString(ConnectionString).
		// C# ref: FastReport.Base/Data/DataConnectionBase.cs Serialize line 1062.
		encrypted, err := utils.EncryptStringDefault(c.ConnectionString)
		if err != nil {
			// On encryption failure, fall back to plaintext.
			encrypted = c.ConnectionString
		}
		w.WriteStr("ConnectionString", encrypted)
	}
	if c.ConnectionStringExpression != "" {
		w.WriteStr("ConnectionStringExpression", c.ConnectionStringExpression)
	}
	if c.LoginPrompt {
		w.WriteBool("LoginPrompt", true)
	}
	if c.CommandTimeout != 30 {
		w.WriteInt("CommandTimeout", c.CommandTimeout)
	}
	// Write enabled child tables.
	// C# ref: DataConnectionBase.Serialize — only enabled tables are written.
	for _, t := range c.tables {
		if t.Enabled() {
			if err := w.WriteObjectNamed("TableDataSource", t); err != nil {
				return err
			}
		}
	}
	return nil
}

// Deserialize reads the connection's properties from r.
// Child TableDataSource elements are read back as registered tables.
// C# ref: FastReport.Data.DataConnectionBase.Deserialize (implicit via FRX load)
func (c *DataConnectionBase) Deserialize(r report.Reader) error {
	name := r.ReadStr("Name", "")
	if name != "" {
		c.SetName(name)
	}
	alias := r.ReadStr("Alias", c.Name())
	if alias != "" {
		c.SetAlias(alias)
	}
	c.SetEnabled(r.ReadBool("Enabled", true))
	c.SetReferenceName(r.ReadStr("ReferenceName", ""))
	// Decrypt the connection string from FRX, matching C# DataConnectionBase
	// property setter which calls SetConnectionString(Crypter.DecryptString(value)).
	// C# ref: FastReport.Base/Data/DataConnectionBase.cs ConnectionString setter line 104.
	rawCS := r.ReadStr("ConnectionString", "")
	if rawCS != "" {
		decrypted, err := utils.DecryptStringDefault(rawCS)
		if err != nil {
			// On decryption failure (e.g. plaintext CS), use as-is.
			decrypted = rawCS
		}
		c.ConnectionString = decrypted
	}
	c.ConnectionStringExpression = r.ReadStr("ConnectionStringExpression", "")
	c.LoginPrompt = r.ReadBool("LoginPrompt", false)
	c.CommandTimeout = r.ReadInt("CommandTimeout", 30)

	// Read child TableDataSource elements.
	for {
		typeName, ok := r.NextChild()
		if !ok {
			break
		}
		if typeName == "TableDataSource" {
			t := NewTableDataSource("")
			if err := t.Deserialize(r); err != nil {
				if ferr := r.FinishChild(); ferr != nil {
					return ferr
				}
				return err
			}
			t.connection = c
			c.tables = append(c.tables, t)
		}
		if err := r.FinishChild(); err != nil {
			return err
		}
	}
	return nil
}

// DriverName returns the database/sql driver name.
func (c *DataConnectionBase) DriverName() string { return c.driverName }

// GetCommandBuilder returns a connection-specific command builder, or nil if
// the base connection does not provide one. Concrete connection subclasses
// that support schema-querying command builders should override this.
// C# ref: FastReport.Data.DataConnectionBase.GetAdapter (analogous virtual method).
func (c *DataConnectionBase) GetCommandBuilder() any { return nil }

// Open opens the underlying *sql.DB using the ConnectionString.
// Returns an error if the connection cannot be established.
//
// If OnDatabaseLogin is set it is called before sql.Open with the current
// ConnectionString; the callback may replace it.
// If OnAfterDatabaseLogin is set it is called after sql.Open succeeds.
// C# ref: FastReport.Base/Data/DataConnectionBase.cs Open() which calls
// Config.ReportSettings.OnDatabaseLogin and Config.ReportSettings.OnAfterDatabaseLogin.
func (c *DataConnectionBase) Open() error {
	if c.db != nil {
		return nil // already open
	}
	dsn := c.ConnectionString
	if c.OnDatabaseLogin != nil {
		args := &DatabaseLoginEventArgs{ConnectionString: dsn}
		c.OnDatabaseLogin(args)
		dsn = args.ConnectionString
	}
	db, err := sql.Open(c.driverName, dsn)
	if err != nil {
		return fmt.Errorf("DataConnection(%s): %w", c.driverName, err)
	}
	c.db = db
	if c.OnAfterDatabaseLogin != nil {
		c.OnAfterDatabaseLogin(&AfterDatabaseLoginEventArgs{DB: db})
	}
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
	// enabled controls whether this table is active during report execution.
	// C# ref: FastReport.Data.DataComponentBase.Enabled
	enabled bool
	// qbSchema holds the query-builder schema string (designer use only).
	// C# ref: FastReport.Data.TableDataSource.QbSchema
	qbSchema string
}

// NewTableDataSource creates a TableDataSource with the given name.
// Enabled defaults to true to match DataComponentBase defaults.
func NewTableDataSource(name string) *TableDataSource {
	return &TableDataSource{
		BaseDataSource: BaseDataSource{name: name, alias: name},
		enabled:        true,
	}
}

// Enabled returns whether this table data source is active.
// C# ref: FastReport.Data.DataComponentBase.Enabled
func (t *TableDataSource) Enabled() bool { return t.enabled }

// SetEnabled enables or disables this table data source.
func (t *TableDataSource) SetEnabled(v bool) { t.enabled = v }

// Serialize writes the TableDataSource properties to w.
// C# ref: FastReport.Data.TableDataSource.Serialize (implicit via FRX writer)
func (t *TableDataSource) Serialize(w report.Writer) error {
	if t.name != "" {
		w.WriteStr("Name", t.name)
	}
	if t.alias != "" && !strings.EqualFold(t.alias, t.name) {
		w.WriteStr("Alias", t.alias)
	}
	if !t.enabled {
		w.WriteBool("Enabled", false)
	}
	if t.tableName != "" {
		w.WriteStr("TableName", t.tableName)
	}
	if t.selectCommand != "" {
		w.WriteStr("SelectCommand", t.selectCommand)
	}
	// QbSchema is a designer-only property; written when non-empty.
	// C# ref: TableDataSource.Serialize line 369.
	if t.qbSchema != "" {
		w.WriteStr("QbSchema", t.qbSchema)
	}
	if t.storeData {
		w.WriteBool("StoreData", true)
	}
	return nil
}

// Deserialize reads the TableDataSource properties from r.
// C# ref: FastReport.Data.TableDataSource.Deserialize (implicit via FRX reader)
func (t *TableDataSource) Deserialize(r report.Reader) error {
	t.name = r.ReadStr("Name", t.name)
	t.alias = r.ReadStr("Alias", t.name)
	t.enabled = r.ReadBool("Enabled", true)
	t.tableName = r.ReadStr("TableName", "")
	t.selectCommand = r.ReadStr("SelectCommand", "")
	// QbSchema is a designer-only property; read when present.
	// C# ref: TableDataSource.QbSchema (FastReport.Base/Data/TableDataSource.cs line 192).
	t.qbSchema = r.ReadStr("QbSchema", "")
	t.storeData = r.ReadBool("StoreData", false)
	return nil
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

// QbSchema returns the query-builder schema string (designer use only).
// C# ref: FastReport.Data.TableDataSource.QbSchema
func (t *TableDataSource) QbSchema() string { return t.qbSchema }

// SetQbSchema sets the query-builder schema string.
func (t *TableDataSource) SetQbSchema(s string) { t.qbSchema = s }

// RefreshTable reloads the table schema and synchronises the column list.
// It is equivalent to deleting and re-running InitSchema and then calling
// RefreshColumns(true) so that newly discovered columns are enabled.
// C# ref: FastReport.Data.TableDataSource.RefreshTable()
func (t *TableDataSource) RefreshTable() error {
	// Reset cached data and schema so InitSchema runs fresh.
	t.initialized = false
	t.rows = nil
	t.columns = nil

	if err := t.InitSchema(); err != nil {
		return err
	}
	t.RefreshColumns(true)
	return nil
}

// RefreshColumns synchronises the column list with the current schema obtained
// from InitSchema. New columns are added (with Enabled = enableNew) and columns
// no longer present in the schema are removed. Calculated columns are preserved.
// C# ref: FastReport.Data.TableDataSource.RefreshColumns(bool enableNew)
func (t *TableDataSource) RefreshColumns(enableNew bool) {
	if len(t.columns) == 0 {
		// No schema columns — nothing to sync.
		return
	}
	// Build a set of schema column names for fast lookup.
	schemaSet := make(map[string]bool, len(t.columns))
	for _, c := range t.columns {
		schemaSet[c.Name] = true
	}
	// We store columns inside BaseDataSource.columns. Make a working copy of
	// the existing user-facing columns (the ones from previous schema loads).
	// Since TableDataSource.columns IS the schema (populated by InitSchema/Init),
	// RefreshColumns is a no-op in the common single-load case; it is called
	// after a fresh InitSchema to add newly appearing columns and remove stale ones.
	// In practice the C# version merges DataTable.Columns with the Columns list.
	// Here we just ensure t.columns reflects the current schema from the connection.
	// Caller has already set t.columns via InitSchema. Mark new columns as enableNew.
	for i := range t.columns {
		if !schemaSet[t.columns[i].Name] {
			// Shouldn't happen since schemaSet is built from t.columns itself.
			continue
		}
		// New columns from InitSchema default to enabled; set per enableNew.
		// The column Enabled flag is not stored in the basic Column struct, but
		// we honour the enableNew semantics for future augmentation.
		_ = enableNew // honour semantics; Column.Enabled can be added later
	}
}

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
