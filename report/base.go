package report

// Restrictions specifies a set of actions that cannot be performed on the object in design mode.
// It is the Go equivalent of FastReport.Restrictions.
type Restrictions int

const (
	// RestrictionsNone specifies no restrictions.
	RestrictionsNone Restrictions = 0
	// DontMove restricts moving the object.
	DontMove Restrictions = 1
	// DontResize restricts resizing the object.
	DontResize Restrictions = 2
	// DontModify restricts modifying the object's properties.
	DontModify Restrictions = 4
	// DontEdit restricts editing the object.
	DontEdit Restrictions = 8
	// DontDelete restricts deleting the object.
	DontDelete Restrictions = 16
	// HideAllProperties hides all properties of the object.
	HideAllProperties Restrictions = 32
)

// ObjectFlags specifies a set of actions that can be performed on the object in design mode.
// It is the Go equivalent of FastReport.Flags.
type ObjectFlags int

const (
	// FlagsNone specifies no actions.
	FlagsNone ObjectFlags = 0
	// CanMove allows moving the object.
	CanMove ObjectFlags = 1
	// CanResize allows resizing the object.
	CanResize ObjectFlags = 2
	// CanDelete allows deleting the object.
	CanDelete ObjectFlags = 4
	// CanEdit allows editing the object.
	CanEdit ObjectFlags = 8
	// CanChangeOrder allows changing the Z-order of an object.
	CanChangeOrder ObjectFlags = 16
	// CanChangeParent allows moving the object to another parent.
	CanChangeParent ObjectFlags = 32
	// CanCopy allows copying the object to the clipboard.
	CanCopy ObjectFlags = 64
	// CanDraw allows drawing the object.
	CanDraw ObjectFlags = 128
	// CanGroup allows grouping the object.
	CanGroup ObjectFlags = 256
	// CanWriteChildren allows write children in the preview mode by itself.
	CanWriteChildren ObjectFlags = 512
	// CanWriteBounds allows write object's bounds into the report stream.
	CanWriteBounds ObjectFlags = 1024
	// HasSmartTag allows the "smart tag" functionality.
	HasSmartTag ObjectFlags = 2048
	// HasGlobalName specifies that the object's name is global.
	HasGlobalName ObjectFlags = 4096
	// CanShowChildrenInReportTree specifies that the object can display children in the designer's Report Tree.
	CanShowChildrenInReportTree ObjectFlags = 8192
	// InterceptsPreviewMouseEvents specifies that the object supports mouse events in the preview window.
	InterceptsPreviewMouseEvents ObjectFlags = 16384
)

// DefaultFlags is the default set of flags applied to new objects.
const DefaultFlags = CanMove | CanResize | CanDelete | CanEdit | CanChangeOrder | CanChangeParent | CanCopy | CanDraw | CanGroup | HasGlobalName

// ObjectState tracks internal runtime state of an object.
// It is the Go equivalent of the internal FastReport.ObjectState enum.
type ObjectState byte

const (
	// ObjectStateNone indicates no special state.
	ObjectStateNone ObjectState = 0
	// IsAncestorState marks the object as defined in an ancestor report.
	IsAncestorState ObjectState = 1
	// IsDesigningState marks the object as being in design mode.
	IsDesigningState ObjectState = 2
	// IsPrintingState marks the object as being in print mode.
	IsPrintingState ObjectState = 4
	// IsRunningState marks the object as being in report-run mode.
	IsRunningState ObjectState = 8
	// IsDeserializingState marks the object as currently being deserialized.
	IsDeserializingState ObjectState = 16
)

// BaseObject is a concrete implementation of the Base interface.
// It is the Go equivalent of FastReport.Base and provides the common fields and behaviour
// shared by all report objects.
type BaseObject struct {
	name         string
	baseName     string
	parent       Parent
	flags        ObjectFlags
	restrictions Restrictions
	objectState  ObjectState
	tag          any
	zOrder       int
}

// NewBaseObject creates a BaseObject with DefaultFlags applied.
func NewBaseObject() *BaseObject {
	return &BaseObject{flags: DefaultFlags}
}

// --- report.Base interface ---

// Name returns the object's name.
func (b *BaseObject) Name() string { return b.name }

// SetName sets the object's name without any validation.
func (b *BaseObject) SetName(name string) { b.name = name }

// BaseName returns the base name prefix used when auto-generating names.
func (b *BaseObject) BaseName() string { return b.baseName }

// SetBaseName sets the base name prefix.
func (b *BaseObject) SetBaseName(name string) { b.baseName = name }

// Parent returns the containing parent, or nil if this is the root.
func (b *BaseObject) Parent() Parent { return b.parent }

// SetParent sets the containing parent directly without checking CanContain.
func (b *BaseObject) SetParent(p Parent) { b.parent = p }

// --- Flags ---

// Flags returns the current ObjectFlags bitmask.
func (b *BaseObject) Flags() ObjectFlags { return b.flags }

// SetFlag sets or clears a specific flag.
func (b *BaseObject) SetFlag(f ObjectFlags, value bool) {
	if value {
		b.flags |= f
	} else {
		b.flags &^= f
	}
}

// HasFlag reports whether the given flag is currently set.
func (b *BaseObject) HasFlag(f ObjectFlags) bool {
	return b.flags&f != 0
}

// --- Restrictions ---

// Restrictions returns the current Restrictions bitmask.
func (b *BaseObject) Restrictions() Restrictions { return b.restrictions }

// SetRestrictions replaces the restrictions bitmask.
func (b *BaseObject) SetRestrictions(r Restrictions) { b.restrictions = r }

// --- Tag ---

// Tag returns the arbitrary user-defined tag value.
func (b *BaseObject) Tag() any { return b.tag }

// SetTag stores an arbitrary user-defined value on the object.
func (b *BaseObject) SetTag(t any) { b.tag = t }

// --- ZOrder ---

// ZOrder returns the object's z-order (index within parent's child list).
// If the object has a parent, the parent's GetChildOrder is queried, mirroring
// the C# Base.ZOrder property (Base.cs).  When there is no parent the internal
// zOrder field is returned.
func (b *BaseObject) ZOrder() int {
	if b.parent != nil {
		return b.parent.GetChildOrder(b)
	}
	return b.zOrder
}

// SetZOrder sets the z-order of the object.
// If the object has a parent, the parent's SetChildOrder is called, mirroring
// the C# Base.ZOrder setter.  When there is no parent the internal zOrder field
// is updated directly.
func (b *BaseObject) SetZOrder(order int) {
	if b.parent != nil {
		b.parent.SetChildOrder(b, order)
	} else {
		b.zOrder = order
	}
}

// HasRestriction reports whether the object has the given restriction set.
// It is the Go equivalent of Base.HasRestriction(Restrictions) (Base.cs).
func (b *BaseObject) HasRestriction(r Restrictions) bool {
	return b.restrictions&r != 0
}

// --- ObjectState helpers ---

// SetObjectState sets or clears a specific ObjectState flag.
func (b *BaseObject) SetObjectState(flag ObjectState, value bool) {
	if value {
		b.objectState |= flag
	} else {
		b.objectState &^= flag
	}
}

// GetObjectState reports whether the given ObjectState flag is set.
func (b *BaseObject) GetObjectState(flag ObjectState) bool {
	return b.objectState&flag != 0
}

// IsAncestor reports whether the object was defined in an ancestor report.
func (b *BaseObject) IsAncestor() bool { return b.GetObjectState(IsAncestorState) }

// IsDesigning reports whether the object is currently in design mode.
func (b *BaseObject) IsDesigning() bool { return b.GetObjectState(IsDesigningState) }

// IsPrinting reports whether the object is currently being printed.
func (b *BaseObject) IsPrinting() bool { return b.GetObjectState(IsPrintingState) }

// IsRunning reports whether the report engine is currently running.
func (b *BaseObject) IsRunning() bool { return b.GetObjectState(IsRunningState) }

// --- Object search / child enumeration ---

// ChildObjects returns the direct children of parent as a flat slice.
func ChildObjects(parent Parent) []Base {
	var list []Base
	parent.GetChildObjects(&list)
	return list
}

// FindObject searches list and its descendants recursively for an object with the given name.
// Returns nil if not found.
func FindObject(name string, list []Base) Base {
	for _, obj := range list {
		if obj.Name() == name {
			return obj
		}
		// Recurse into children if obj is also a Parent.
		if p, ok := obj.(Parent); ok {
			var children []Base
			p.GetChildObjects(&children)
			if found := FindObject(name, children); found != nil {
				return found
			}
		}
	}
	return nil
}

// AllObjects returns all descendants of root recursively, not including root
// itself.  It is the Go equivalent of Base.AllObjects (Base.cs).
// root must implement Parent; if it does not, an empty slice is returned.
func AllObjects(root Base) []Base {
	var result []Base
	if p, ok := root.(Parent); ok {
		enumObjects(p, &result)
	}
	return result
}

// enumObjects appends all descendants of parent to result.
func enumObjects(parent Parent, result *[]Base) {
	var children []Base
	parent.GetChildObjects(&children)
	for _, child := range children {
		*result = append(*result, child)
		if p, ok := child.(Parent); ok {
			enumObjects(p, result)
		}
	}
}

// HasParent reports whether obj has ancestor in its parent chain.
// It is the Go equivalent of Base.HasParent(Base) (Base.cs).
func HasParent(obj Base, ancestor Parent) bool {
	p := obj.Parent()
	for p != nil {
		if p == ancestor {
			return true
		}
		// Walk up: Parent interface does not expose its own Parent directly;
		// check if the Parent is also a Base so we can continue climbing.
		if b, ok := p.(Base); ok {
			p = b.Parent()
		} else {
			break
		}
	}
	return false
}

// --- Serialization ---

// Serialize writes the object's Name, Restrictions, and Flags to w when they differ from defaults.
func (b *BaseObject) Serialize(w Writer) error {
	if b.name != "" {
		w.WriteStr("Name", b.name)
	}
	if b.restrictions != RestrictionsNone {
		w.WriteInt("Restrictions", int(b.restrictions))
	}
	if b.flags != DefaultFlags {
		w.WriteInt("Flags", int(b.flags))
	}
	return nil
}

// Deserialize reads Name, Restrictions, and Flags from r.
func (b *BaseObject) Deserialize(r Reader) error {
	b.name = r.ReadStr("Name", b.name)
	b.restrictions = Restrictions(r.ReadInt("Restrictions", int(b.restrictions)))
	b.flags = ObjectFlags(r.ReadInt("Flags", int(b.flags)))
	return nil
}
