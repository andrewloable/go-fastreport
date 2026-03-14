// Package report provides the core types and interfaces for go-fastreport.
// It is the Go equivalent of the FastReport.Base root namespace.
package report

// Serializable is the Go equivalent of IFRSerializable.
// All report objects that can be saved/loaded in FRX format implement this interface.
type Serializable interface {
	// Serialize writes the object's state to w.
	Serialize(w Writer) error
	// Deserialize reads the object's state from r.
	Deserialize(r Reader) error
}

// Writer is a minimal write-side interface used during FRX serialization.
// The full implementation lives in the serial package; this forward declaration
// breaks the import cycle between report/ and serial/.
type Writer interface {
	// WriteStr writes a named string property.
	WriteStr(name, value string)
	// WriteInt writes a named int property.
	WriteInt(name string, value int)
	// WriteBool writes a named bool property.
	WriteBool(name string, value bool)
	// WriteFloat writes a named float32 property.
	WriteFloat(name string, value float32)
	// WriteObject writes a child object.
	WriteObject(obj Serializable) error
}

// Reader is a minimal read-side interface used during FRX deserialization.
type Reader interface {
	// ReadStr reads a named string property, returning def if absent.
	ReadStr(name, def string) string
	// ReadInt reads a named int property, returning def if absent.
	ReadInt(name string, def int) int
	// ReadBool reads a named bool property, returning def if absent.
	ReadBool(name string, def bool) bool
	// ReadFloat reads a named float32 property, returning def if absent.
	ReadFloat(name string, def float32) float32
	// NextChild advances to the next child element and returns its type name.
	// Returns ("", false) when there are no more children.
	NextChild() (typeName string, ok bool)
}

// Base is the minimum interface implemented by every report object.
// It is the Go equivalent of FastReport.Base.
type Base interface {
	Serializable
	// Name returns the object's name.
	Name() string
	// SetName sets the object's name.
	SetName(name string)
	// BaseName returns the base name prefix used when auto-generating names
	// (e.g. "Text" for a TextObject).
	BaseName() string
	// Parent returns the containing parent, or nil if this is the root.
	Parent() Parent
	// SetParent sets the containing parent.
	SetParent(p Parent)
}

// Parent is the Go equivalent of IParent.
// Objects that can contain child objects implement this interface.
type Parent interface {
	// CanContain returns true if this parent can accept child as a child.
	CanContain(child Base) bool
	// GetChildObjects fills list with all child objects.
	GetChildObjects(list *[]Base)
	// AddChild adds child to this parent's children.
	AddChild(child Base)
	// RemoveChild removes child from this parent's children.
	RemoveChild(child Base)
	// GetChildOrder returns the z-order (index) of child in the internal list.
	GetChildOrder(child Base) int
	// SetChildOrder moves child to the specified z-order position.
	SetChildOrder(child Base, order int)
	// UpdateLayout updates child positions/sizes after the parent is resized by dx, dy.
	UpdateLayout(dx, dy float32)
}

// DataSourceBinder is the Go equivalent of IContainDataSource (internal interface).
// Objects that bind to a data source implement this.
type DataSourceBinder interface {
	// DataSourceName returns the name of the bound data source.
	DataSourceName() string
}

// Translatable is the Go equivalent of ITranslatable (internal interface).
// Objects that can be converted to/from native report objects implement this.
type Translatable interface {
	// ConvertToReportObjects performs any necessary translation.
	ConvertToReportObjects()
}
