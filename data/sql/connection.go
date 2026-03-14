// Package sql provides SQL database connections for go-fastreport.
// Each connection type wraps DataConnectionBase with a driver-specific DSN format.
//
// Drivers are NOT imported here — the user must blank-import the desired driver:
//
//	import _ "modernc.org/sqlite"                    // SQLite
//	import _ "github.com/jackc/pgx/v5/stdlib"        // PostgreSQL
//	import _ "github.com/go-sql-driver/mysql"         // MySQL
//	import _ "github.com/microsoft/go-mssqldb"        // SQL Server
package sql

import (
	"fmt"

	"github.com/andrewloable/go-fastreport/data"
)

// ── PostgreSQL ────────────────────────────────────────────────────────────────

// PostgresConnection is a DataConnectionBase configured for PostgreSQL via pgx.
// Driver: "pgx" (requires "github.com/jackc/pgx/v5/stdlib").
type PostgresConnection struct {
	data.DataConnectionBase

	// Host is the server hostname or IP (default: "localhost").
	Host string
	// Port is the server port (default: 5432).
	Port int
	// Database is the target database name.
	Database string
	// User is the login username.
	User string
	// Password is the login password.
	Password string
	// SSLMode is the SSL mode (default: "disable").
	SSLMode string
}

// NewPostgresConnection creates a PostgresConnection with defaults.
func NewPostgresConnection() *PostgresConnection {
	return &PostgresConnection{
		DataConnectionBase: *data.NewDataConnectionBase("pgx"),
		Host:               "localhost",
		Port:               5432,
		SSLMode:            "disable",
	}
}

// BuildDSN builds and sets the DSN from the individual connection fields.
func (c *PostgresConnection) BuildDSN() {
	c.ConnectionString = fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		c.Host, c.Port, c.User, c.Password, c.Database, c.SSLMode,
	)
}

// ── MySQL ─────────────────────────────────────────────────────────────────────

// MySQLConnection is a DataConnectionBase configured for MySQL.
// Driver: "mysql" (requires "github.com/go-sql-driver/mysql").
type MySQLConnection struct {
	data.DataConnectionBase

	// Host is the server hostname or IP (default: "localhost").
	Host string
	// Port is the server port (default: 3306).
	Port int
	// Database is the target database name.
	Database string
	// User is the login username.
	User string
	// Password is the login password.
	Password string
}

// NewMySQLConnection creates a MySQLConnection with defaults.
func NewMySQLConnection() *MySQLConnection {
	return &MySQLConnection{
		DataConnectionBase: *data.NewDataConnectionBase("mysql"),
		Host:               "localhost",
		Port:               3306,
	}
}

// BuildDSN builds and sets the DSN from the individual connection fields.
func (c *MySQLConnection) BuildDSN() {
	c.ConnectionString = fmt.Sprintf(
		"%s:%s@tcp(%s:%d)/%s?parseTime=true",
		c.User, c.Password, c.Host, c.Port, c.Database,
	)
}

// ── SQLite ────────────────────────────────────────────────────────────────────

// SQLiteConnection is a DataConnectionBase configured for SQLite.
// Driver: "sqlite" (requires "modernc.org/sqlite").
type SQLiteConnection struct {
	data.DataConnectionBase

	// FilePath is the path to the SQLite database file.
	// Use ":memory:" for an in-memory database.
	FilePath string
}

// NewSQLiteConnection creates a SQLiteConnection for the given file path.
func NewSQLiteConnection(filePath string) *SQLiteConnection {
	c := &SQLiteConnection{
		DataConnectionBase: *data.NewDataConnectionBase("sqlite"),
		FilePath:           filePath,
	}
	c.ConnectionString = filePath
	return c
}

// SetFilePath updates the file path and the connection string.
func (c *SQLiteConnection) SetFilePath(path string) {
	c.FilePath = path
	c.ConnectionString = path
}

// ── MSSQL (SQL Server) ────────────────────────────────────────────────────────

// MSSQLConnection is a DataConnectionBase configured for Microsoft SQL Server.
// Driver: "sqlserver" (requires "github.com/microsoft/go-mssqldb").
type MSSQLConnection struct {
	data.DataConnectionBase

	// Host is the server hostname or IP (default: "localhost").
	Host string
	// Port is the server port (default: 1433).
	Port int
	// Database is the target database name.
	Database string
	// User is the login username.
	User string
	// Password is the login password.
	Password string
	// Instance is the named instance (empty = default instance).
	Instance string
}

// NewMSSQLConnection creates an MSSQLConnection with defaults.
func NewMSSQLConnection() *MSSQLConnection {
	return &MSSQLConnection{
		DataConnectionBase: *data.NewDataConnectionBase("sqlserver"),
		Host:               "localhost",
		Port:               1433,
	}
}

// BuildDSN builds and sets the DSN from the individual connection fields.
func (c *MSSQLConnection) BuildDSN() {
	instance := ""
	if c.Instance != "" {
		instance = `\` + c.Instance
	}
	c.ConnectionString = fmt.Sprintf(
		"sqlserver://%s:%s@%s%s:%d?database=%s",
		c.User, c.Password, c.Host, instance, c.Port, c.Database,
	)
}
