package data

import (
	"strings"

	"github.com/andrewloable/go-fastreport/report"
)

// DataComponentBase is the base for all data components: connections, data sources,
// and columns. It is the Go equivalent of FastReport.Data.DataComponentBase.
//
// DataComponentBase adds Alias, Enabled, ReferenceName, and Reference fields on
// top of the minimal report.Base contract. Both DataConnectionBase and
// BaseDataSource should embed it to gain consistent FRX serialization.
type DataComponentBase struct {
	// name is the internal component name used in expressions and FRX files.
	name string
	// alias is the human-friendly display name shown in the data dictionary.
	// Defaults to the same value as name.
	alias string
	// enabled controls whether this component is active during report execution.
	// Disabled components are hidden from the dictionary but still accessible.
	enabled bool
	// referenceName stores the name of a shared object reference (infrastructure use).
	referenceName string
	// reference holds a run-time object reference (infrastructure use).
	reference any
}

// NewDataComponentBase creates a DataComponentBase with Enabled=true.
func NewDataComponentBase(name string) *DataComponentBase {
	return &DataComponentBase{
		name:    name,
		alias:   name,
		enabled: true,
	}
}

// Name returns the component's internal name.
func (d *DataComponentBase) Name() string { return d.name }

// SetName sets the internal name. When alias was empty or case-insensitively
// equal to the old name it is also updated to stay in sync, matching the C#
// behaviour:
//
//	bool changeAlias = String.IsNullOrEmpty(Alias) ||
//	                   String.Compare(Alias, Name, true) == 0;
func (d *DataComponentBase) SetName(name string) {
	if d.alias == "" || strings.EqualFold(d.alias, d.name) {
		d.alias = name
	}
	d.name = name
}

// Alias returns the human-friendly display name.
func (d *DataComponentBase) Alias() string { return d.alias }

// SetAlias sets the display alias.
func (d *DataComponentBase) SetAlias(a string) { d.alias = a }

// IsAliased returns true when the alias differs from the name.
func (d *DataComponentBase) IsAliased() bool { return d.name != d.alias }

// Enabled returns whether this component is active.
func (d *DataComponentBase) Enabled() bool { return d.enabled }

// SetEnabled enables or disables this component.
func (d *DataComponentBase) SetEnabled(v bool) { d.enabled = v }

// ReferenceName returns the infrastructure reference name.
func (d *DataComponentBase) ReferenceName() string { return d.referenceName }

// SetReferenceName sets the infrastructure reference name.
func (d *DataComponentBase) SetReferenceName(n string) { d.referenceName = n }

// Reference returns the run-time reference object.
func (d *DataComponentBase) Reference() any { return d.reference }

// SetReference sets the run-time reference object.
func (d *DataComponentBase) SetReference(ref any) { d.reference = ref }

// InitializeComponent is called by the engine before running a report.
// Subclasses override this to perform late initialization.
// The base implementation is intentionally a no-op; types that embed
// DataComponentBase override this method to add their own setup logic.
func (d *DataComponentBase) InitializeComponent() {
	_ = d // no-op base implementation; subclasses re-implement as needed
}

// Serialize writes the component's non-default properties to w.
func (d *DataComponentBase) Serialize(w report.Writer) error {
	if d.IsAliased() {
		w.WriteStr("Alias", d.alias)
	}
	if !d.enabled {
		w.WriteBool("Enabled", false)
	}
	if d.referenceName != "" {
		w.WriteStr("ReferenceName", d.referenceName)
	}
	return nil
}

// Deserialize reads the component's properties from r.
func (d *DataComponentBase) Deserialize(r report.Reader) error {
	d.alias = r.ReadStr("Alias", d.name)
	d.enabled = r.ReadBool("Enabled", true)
	d.referenceName = r.ReadStr("ReferenceName", "")
	return nil
}
