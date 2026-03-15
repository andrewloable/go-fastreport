package sql

import (
	"fmt"

	"github.com/andrewloable/go-fastreport/data"
	"github.com/andrewloable/go-fastreport/report"
)

// ProcedureDataSource is a data source that executes a stored procedure.
// It is the Go equivalent of FastReport.Data.ProcedureDataSource.
//
// The SelectCommand should be the full stored procedure call expression, e.g.:
//
//	"CALL my_procedure(?, ?)"        -- MySQL / PostgreSQL
//	"EXEC my_procedure @id=?"        -- SQL Server
//
// Input parameters are bound via AddParameter; output parameter values are
// accessible as synthetic columns after Init() is called.
type ProcedureDataSource struct {
	data.TableDataSource
}

// NewProcedureDataSource creates a ProcedureDataSource with the given name.
func NewProcedureDataSource(name string) *ProcedureDataSource {
	return &ProcedureDataSource{
		TableDataSource: *data.NewTableDataSource(name),
	}
}

// TypeName returns "ProcedureDataSource" for FRX serialization.
func (p *ProcedureDataSource) TypeName() string { return "ProcedureDataSource" }

// BaseName returns the auto-name prefix.
func (p *ProcedureDataSource) BaseName() string { return "ProcedureDataSource" }

// Serialize writes ProcedureDataSource properties for FRX round-trip.
func (p *ProcedureDataSource) Serialize(w report.Writer) error {
	name := p.Name()
	if name != "" {
		w.WriteStr("Name", name)
	}
	if alias := p.Alias(); alias != "" && alias != name {
		w.WriteStr("Alias", alias)
	}
	if sc := p.SelectCommand(); sc != "" {
		w.WriteStr("SelectCommand", sc)
	}
	if tn := p.TableName(); tn != "" {
		w.WriteStr("TableName", tn)
	}
	if p.StoreData() {
		w.WriteBool("StoreData", true)
	}
	for _, param := range p.Parameters() {
		if err := w.WriteObjectNamed("Parameter", param); err != nil {
			return err
		}
	}
	return nil
}

// Deserialize reads ProcedureDataSource properties from FRX.
func (p *ProcedureDataSource) Deserialize(r report.Reader) error {
	p.SetName(r.ReadStr("Name", ""))
	alias := r.ReadStr("Alias", "")
	if alias == "" {
		alias = p.Name()
	}
	p.SetAlias(alias)
	p.SetSelectCommand(r.ReadStr("SelectCommand", ""))
	p.SetTableName(r.ReadStr("TableName", ""))
	p.SetStoreData(r.ReadBool("StoreData", false))
	return nil
}

// Init executes the stored procedure and populates the in-memory row store.
// Output parameters are captured as synthetic columns after the result set.
func (p *ProcedureDataSource) Init() error {
	// Run the base TableDataSource init which handles the query + parameters.
	if err := p.TableDataSource.Init(); err != nil {
		return err
	}
	// Populate synthetic columns from output/inputoutput parameters.
	for _, param := range p.Parameters() {
		if param.Direction == data.ParamDirectionInput {
			continue
		}
		found := false
		for _, col := range p.Columns() {
			if col.Name == param.Name {
				found = true
				break
			}
		}
		if !found {
			p.AddColumn(data.Column{Name: param.Name, Alias: param.Name, DataType: "any"})
		}
	}
	return nil
}

// GetValue returns the value for the named column.
// For output parameters, it returns the parameter's resolved Value directly.
func (p *ProcedureDataSource) GetValue(column string) (any, error) {
	for _, param := range p.Parameters() {
		if param.Name == column && param.Direction != data.ParamDirectionInput {
			return param.Value, nil
		}
	}
	v, err := p.TableDataSource.GetValue(column)
	if err != nil {
		return nil, fmt.Errorf("ProcedureDataSource %q: column %q: %w", p.Name(), column, err)
	}
	return v, nil
}

// NewProcedureDataSourceSQL is a convenience constructor that connects a
// ProcedureDataSource to an existing DataConnectionBase.
func NewProcedureDataSourceSQL(conn *data.DataConnectionBase, call, name string) *ProcedureDataSource {
	ds := NewProcedureDataSource(name)
	ds.SetSelectCommand(call)
	conn.AddTable(&ds.TableDataSource)
	return ds
}
