package report

// BreakableComponent extends ReportComponentBase for objects that can split
// their content across page boundaries. It is the Go equivalent of
// FastReport.BreakableComponent.
//
// Both BandBase and TextObjectBase embed BreakableComponent.
type BreakableComponent struct {
	ReportComponentBase

	// canBreak controls whether the engine may split this object at a page boundary.
	// Defaults to true.
	canBreak bool

	// breakTo is a reference to another BreakableComponent of the same type that
	// receives the overflowing content when this object is split. Callers must
	// set this to nil when the referenced object is destroyed.
	breakTo *BreakableComponent

	// FlagMustBreak is set by the report engine to force a break on the next pass.
	FlagMustBreak bool
}

// NewBreakableComponent creates a BreakableComponent with CanBreak=true.
func NewBreakableComponent() *BreakableComponent {
	bc := &BreakableComponent{
		ReportComponentBase: *NewReportComponentBase(),
		canBreak:            true,
	}
	return bc
}

// CanBreak returns whether this component may be broken across page boundaries.
func (bc *BreakableComponent) CanBreak() bool { return bc.canBreak }

// SetCanBreak sets the canBreak flag.
func (bc *BreakableComponent) SetCanBreak(v bool) { bc.canBreak = v }

// BreakTo returns the companion object that receives overflowing content, or nil.
func (bc *BreakableComponent) BreakTo() *BreakableComponent { return bc.breakTo }

// SetBreakTo sets the companion overflow object. Pass nil to clear.
func (bc *BreakableComponent) SetBreakTo(target *BreakableComponent) { bc.breakTo = target }

// Break splits the component content at the current height boundary.
// The part that fits stays in this object; the overflow is placed in breakTo.
// Returns true when at least one unit of content fits in this object.
//
// The base implementation always returns false (no content to split).
// Concrete types such as TextObject override this to perform real splitting.
func (bc *BreakableComponent) Break(breakTo *BreakableComponent) bool {
	return false
}

// CalcHeight returns the minimum height needed to display the component's
// content. The base implementation returns the current Height. Subclasses
// override this to compute the required height from their content.
func (bc *BreakableComponent) CalcHeight() float32 {
	return bc.Height()
}

// Assign copies all BreakableComponent properties from src into this component.
// Note: the breakTo reference is copied as a pointer (shallow). Callers that
// need lifetime-independent copies must clear BreakTo separately.
//
// Mirrors C# BreakableComponent.Assign(Base source) (BreakableComponent.cs line 64-71).
func (bc *BreakableComponent) Assign(src *BreakableComponent) {
	if src == nil {
		return
	}
	bc.ReportComponentBase = src.ReportComponentBase
	bc.canBreak = src.canBreak
	bc.breakTo = src.breakTo
}

// Serialize writes BreakableComponent properties that differ from defaults.
func (bc *BreakableComponent) Serialize(w Writer) error {
	if err := bc.ReportComponentBase.Serialize(w); err != nil {
		return err
	}
	if !bc.canBreak {
		w.WriteBool("CanBreak", false)
	}
	return nil
}

// Deserialize reads BreakableComponent properties.
func (bc *BreakableComponent) Deserialize(r Reader) error {
	if err := bc.ReportComponentBase.Deserialize(r); err != nil {
		return err
	}
	bc.canBreak = r.ReadBool("CanBreak", true)
	return nil
}
