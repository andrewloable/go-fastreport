package sql

import (
	"github.com/andrewloable/go-fastreport/data"
)

// NewSQLDataSource creates a data.TableDataSource bound to conn that will
// execute query when Init() is called. The data source is ready to pass to
// data.Dictionary.AddDataSource().
//
// Example:
//
//	conn := sql.NewSQLiteConnection("mydb.sqlite")
//	ds := sql.NewSQLDataSource(conn.DataConnectionBase, "SELECT * FROM orders", "Orders")
//	dict.AddDataSource(ds)
func NewSQLDataSource(conn *data.DataConnectionBase, query, name string) *data.TableDataSource {
	ds := data.NewTableDataSource(name)
	ds.SetSelectCommand(query)
	conn.AddTable(ds)
	return ds
}

// RegisterSQL is a convenience function that creates a connection, builds its DSN,
// creates a TableDataSource, registers it in the dictionary, and returns the data
// source. The caller must supply the DSN builder (typically Connection.BuildDSN).
//
// For simple use cases, prefer calling dict.AddDataSource(sql.NewSQLDataSource(...)).
func RegisterSQL(dict *data.Dictionary, conn *data.DataConnectionBase, query, name string) *data.TableDataSource {
	ds := NewSQLDataSource(conn, query, name)
	dict.AddDataSource(ds)
	return ds
}
