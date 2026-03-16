package reportpkg

import (
	"github.com/andrewloable/go-fastreport/report"
)

// DialogPage is a minimal stub for FastReport's <DialogPage> element.
// Dialog pages contain UI form controls (buttons, text boxes, etc.) that are
// used to display input dialogs at runtime in the .NET designer. In this
// Go engine dialog pages are not rendered; they are deserialized without error
// and then silently ignored at engine time.
type DialogPage struct {
	report.BaseObject
}

// TypeName returns the FRX element name.
func (*DialogPage) TypeName() string { return "DialogPage" }

// NewDialogPage creates an empty DialogPage.
func NewDialogPage() *DialogPage {
	return &DialogPage{
		BaseObject: *report.NewBaseObject(),
	}
}

// Deserialize reads DialogPage attributes (name, size, etc.) and drains all
// child control elements without error.
func (d *DialogPage) Deserialize(r report.Reader) error {
	// Read the base Name attribute so the object has an identity.
	if err := d.BaseObject.Deserialize(r); err != nil {
		return err
	}
	// Drain all child elements (ButtonControl, LabelControl, TextBoxControl, …).
	for {
		_, ok := r.NextChild()
		if !ok {
			break
		}
		// Skip the entire sub-tree of each child control.
		_ = drainChildren(r)
		_ = r.FinishChild()
	}
	return nil
}

// drainChildren consumes and discards all nested child elements from r.
func drainChildren(r report.Reader) error {
	for {
		_, ok := r.NextChild()
		if !ok {
			break
		}
		_ = drainChildren(r)
		_ = r.FinishChild()
	}
	return nil
}

// Serialize is a no-op: dialog pages are never written back to FRX by this engine.
func (d *DialogPage) Serialize(_ report.Writer) error { return nil }
