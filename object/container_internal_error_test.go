package object

// container_internal_error_test.go — internal tests to cover the error-return
// branch inside ContainerObject.Serialize when WriteObject returns an error.
//
// The branch at container.go:273-275 (return err inside the child-write loop)
// is unreachable from the public serial.Writer because WriteObject only fails
// when BeginObject fails, which in turn requires an underlying I/O failure.
// This internal test injects a mock report.Writer whose WriteObject always
// returns a sentinel error so that the loop's early-return path is executed.

import (
	"errors"
	"testing"

	"github.com/andrewloable/go-fastreport/report"
)

// writeObjectErrWriter is a minimal mock of report.Writer whose WriteObject
// always returns a sentinel error. All attribute-write methods are no-ops.
type writeObjectErrWriter struct {
	errMsg string
}

func (w *writeObjectErrWriter) WriteStr(name, value string)           {}
func (w *writeObjectErrWriter) WriteInt(name string, value int)       {}
func (w *writeObjectErrWriter) WriteBool(name string, value bool)     {}
func (w *writeObjectErrWriter) WriteFloat(name string, value float32) {}
func (w *writeObjectErrWriter) WriteObject(obj report.Serializable) error {
	return errors.New(w.errMsg)
}
func (w *writeObjectErrWriter) WriteObjectNamed(name string, obj report.Serializable) error {
	return errors.New(w.errMsg)
}

// TestContainerObject_Serialize_WriteObjectError covers the return-err branch
// inside the ContainerObject.Serialize child-write loop (container.go:273-275).
func TestContainerObject_Serialize_WriteObjectError(t *testing.T) {
	c := NewContainerObject()
	child := NewBarcodeObject()
	child.SetText("child1")
	c.AddChild(child)

	w := &writeObjectErrWriter{errMsg: "child write error"}
	err := c.Serialize(w)
	if err == nil {
		t.Fatal("expected error from Serialize when WriteObject fails for child, got nil")
	}
	if err.Error() != "child write error" {
		t.Errorf("unexpected error message: %q", err.Error())
	}
}
