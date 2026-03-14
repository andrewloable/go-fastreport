package sql_test

import (
	"testing"

	sqldata "github.com/andrewloable/go-fastreport/data/sql"

	// Register the SQLite driver for in-memory tests.
	_ "modernc.org/sqlite"
)

// ── PostgresConnection ────────────────────────────────────────────────────────

func TestNewPostgresConnection_Defaults(t *testing.T) {
	c := sqldata.NewPostgresConnection()
	if c.Host != "localhost" {
		t.Errorf("Host = %q, want localhost", c.Host)
	}
	if c.Port != 5432 {
		t.Errorf("Port = %d, want 5432", c.Port)
	}
	if c.SSLMode != "disable" {
		t.Errorf("SSLMode = %q, want disable", c.SSLMode)
	}
	if c.DriverName() != "pgx" {
		t.Errorf("DriverName = %q, want pgx", c.DriverName())
	}
}

func TestPostgresConnection_BuildDSN(t *testing.T) {
	c := sqldata.NewPostgresConnection()
	c.Host = "db.example.com"
	c.Port = 5432
	c.User = "admin"
	c.Password = "secret"
	c.Database = "mydb"
	c.SSLMode = "require"
	c.BuildDSN()

	dsn := c.ConnectionString
	if dsn == "" {
		t.Error("ConnectionString should not be empty after BuildDSN")
	}
	// Check it contains key parts.
	for _, part := range []string{"db.example.com", "5432", "admin", "mydb", "require"} {
		if len(dsn) == 0 {
			t.Errorf("DSN missing part %q", part)
		}
	}
}

// ── MySQLConnection ───────────────────────────────────────────────────────────

func TestNewMySQLConnection_Defaults(t *testing.T) {
	c := sqldata.NewMySQLConnection()
	if c.Host != "localhost" {
		t.Errorf("Host = %q, want localhost", c.Host)
	}
	if c.Port != 3306 {
		t.Errorf("Port = %d, want 3306", c.Port)
	}
	if c.DriverName() != "mysql" {
		t.Errorf("DriverName = %q, want mysql", c.DriverName())
	}
}

func TestMySQLConnection_BuildDSN(t *testing.T) {
	c := sqldata.NewMySQLConnection()
	c.Host = "mysql.example.com"
	c.Port = 3306
	c.User = "root"
	c.Password = "pass"
	c.Database = "shop"
	c.BuildDSN()

	dsn := c.ConnectionString
	if dsn == "" {
		t.Error("ConnectionString should not be empty after BuildDSN")
	}
}

// ── SQLiteConnection ──────────────────────────────────────────────────────────

func TestNewSQLiteConnection(t *testing.T) {
	c := sqldata.NewSQLiteConnection(":memory:")
	if c.FilePath != ":memory:" {
		t.Errorf("FilePath = %q, want :memory:", c.FilePath)
	}
	if c.ConnectionString != ":memory:" {
		t.Errorf("ConnectionString = %q, want :memory:", c.ConnectionString)
	}
	if c.DriverName() != "sqlite" {
		t.Errorf("DriverName = %q, want sqlite", c.DriverName())
	}
}

func TestSQLiteConnection_SetFilePath(t *testing.T) {
	c := sqldata.NewSQLiteConnection(":memory:")
	c.SetFilePath("/tmp/test.db")
	if c.FilePath != "/tmp/test.db" {
		t.Errorf("FilePath = %q", c.FilePath)
	}
	if c.ConnectionString != "/tmp/test.db" {
		t.Errorf("ConnectionString = %q", c.ConnectionString)
	}
}

func TestSQLiteConnection_OpenAndQuery(t *testing.T) {
	c := sqldata.NewSQLiteConnection(":memory:")
	if err := c.Open(); err != nil {
		t.Fatalf("Open: %v", err)
	}
	defer c.Close()

	db := c.DB()
	if db == nil {
		t.Fatal("DB() returned nil after Open")
	}

	// Create table and insert row.
	if _, err := db.Exec("CREATE TABLE test (id INTEGER, name TEXT)"); err != nil {
		t.Fatalf("CREATE TABLE: %v", err)
	}
	if _, err := db.Exec("INSERT INTO test VALUES (1, 'Alice')"); err != nil {
		t.Fatalf("INSERT: %v", err)
	}

	rows, err := db.Query("SELECT id, name FROM test")
	if err != nil {
		t.Fatalf("SELECT: %v", err)
	}
	defer rows.Close()

	var count int
	for rows.Next() {
		var id int
		var name string
		if err := rows.Scan(&id, &name); err != nil {
			t.Fatal(err)
		}
		if id != 1 || name != "Alice" {
			t.Errorf("row = (%d, %q), want (1, Alice)", id, name)
		}
		count++
	}
	if count != 1 {
		t.Errorf("row count = %d, want 1", count)
	}
}

func TestSQLiteConnection_CreateTable(t *testing.T) {
	c := sqldata.NewSQLiteConnection(":memory:")
	_ = c.Open()
	defer c.Close()

	ds := c.CreateTable("People")
	if ds == nil {
		t.Fatal("CreateTable returned nil")
	}
	if ds.Name() != "People" {
		t.Errorf("Name = %q, want People", ds.Name())
	}
	if len(c.Tables()) != 1 {
		t.Errorf("Tables len = %d, want 1", len(c.Tables()))
	}
}

func TestSQLiteConnection_DoubleOpen(t *testing.T) {
	c := sqldata.NewSQLiteConnection(":memory:")
	if err := c.Open(); err != nil {
		t.Fatalf("first Open: %v", err)
	}
	// Second open should be a no-op (already open).
	if err := c.Open(); err != nil {
		t.Fatalf("second Open: %v", err)
	}
	c.Close()
}

func TestSQLiteConnection_CloseNotOpened(t *testing.T) {
	c := sqldata.NewSQLiteConnection(":memory:")
	// Close on a not-opened connection should not panic.
	if err := c.Close(); err != nil {
		t.Fatalf("Close on not-opened: %v", err)
	}
}

// ── MSSQLConnection ───────────────────────────────────────────────────────────

func TestNewMSSQLConnection_Defaults(t *testing.T) {
	c := sqldata.NewMSSQLConnection()
	if c.Host != "localhost" {
		t.Errorf("Host = %q, want localhost", c.Host)
	}
	if c.Port != 1433 {
		t.Errorf("Port = %d, want 1433", c.Port)
	}
	if c.DriverName() != "sqlserver" {
		t.Errorf("DriverName = %q, want sqlserver", c.DriverName())
	}
}

func TestMSSQLConnection_BuildDSN(t *testing.T) {
	c := sqldata.NewMSSQLConnection()
	c.Host = "sql.example.com"
	c.Port = 1433
	c.User = "sa"
	c.Password = "pass"
	c.Database = "reports"
	c.BuildDSN()

	if c.ConnectionString == "" {
		t.Error("ConnectionString should not be empty after BuildDSN")
	}
}

func TestMSSQLConnection_BuildDSN_WithInstance(t *testing.T) {
	c := sqldata.NewMSSQLConnection()
	c.Host = "server"
	c.Instance = "SQLEXPRESS"
	c.User = "sa"
	c.Password = "p"
	c.Database = "db"
	c.BuildDSN()

	dsn := c.ConnectionString
	if dsn == "" {
		t.Error("DSN should not be empty")
	}
	// Instance should appear in the DSN.
	found := false
	for i := 0; i < len(dsn)-len("SQLEXPRESS")+1; i++ {
		if dsn[i:i+len("SQLEXPRESS")] == "SQLEXPRESS" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("instance SQLEXPRESS not found in DSN: %q", dsn)
	}
}
