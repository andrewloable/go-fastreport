package utils_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/loabletech/go-fastreport/utils"
)

func TestDuplicateNameError(t *testing.T) {
	err := &utils.DuplicateNameError{Name: "Text1"}
	if !strings.Contains(err.Error(), "Text1") {
		t.Errorf("expected name in error: %v", err)
	}
	if !strings.Contains(err.Error(), "duplicate") {
		t.Errorf("expected 'duplicate' in error: %v", err)
	}
}

func TestAncestorError(t *testing.T) {
	err := &utils.AncestorError{Name: "Band1"}
	if !strings.Contains(err.Error(), "Band1") {
		t.Errorf("expected name in error: %v", err)
	}
}

func TestFileFormatError(t *testing.T) {
	t.Run("with detail", func(t *testing.T) {
		err := &utils.FileFormatError{Detail: "missing root element"}
		if !strings.Contains(err.Error(), "missing root element") {
			t.Errorf("expected detail in error: %v", err)
		}
	})
	t.Run("without detail", func(t *testing.T) {
		err := &utils.FileFormatError{}
		if err.Error() == "" {
			t.Error("expected non-empty error message")
		}
	})
}

func TestDecryptError(t *testing.T) {
	err := &utils.DecryptError{}
	if err.Error() == "" {
		t.Error("expected non-empty error message")
	}
}

func TestCompilerError(t *testing.T) {
	infos := []utils.CompilerErrorInfo{
		{Line: 1, Column: 5, ReportObject: "Text1", Message: "syntax error"},
	}
	err := &utils.CompilerError{Msg: "compilation failed", Errors: infos}
	if err.Error() != "compilation failed" {
		t.Errorf("unexpected error message: %v", err)
	}
	if len(err.Errors) != 1 {
		t.Errorf("expected 1 error info, got %d", len(err.Errors))
	}
	if err.Errors[0].Line != 1 || err.Errors[0].Column != 5 {
		t.Errorf("unexpected error info: %+v", err.Errors[0])
	}
}

func TestParentError(t *testing.T) {
	err := &utils.ParentError{ParentType: "DataBand", ChildType: "ReportPage"}
	if !strings.Contains(err.Error(), "DataBand") || !strings.Contains(err.Error(), "ReportPage") {
		t.Errorf("expected both types in error: %v", err)
	}
}

func TestClassError(t *testing.T) {
	err := &utils.ClassError{Name: "UnknownWidget"}
	if !strings.Contains(err.Error(), "UnknownWidget") {
		t.Errorf("expected class name in error: %v", err)
	}
}

func TestDataTableError(t *testing.T) {
	err := &utils.DataTableError{Alias: "Orders"}
	if !strings.Contains(err.Error(), "Orders") {
		t.Errorf("expected alias in error: %v", err)
	}
}

func TestDataNotInitializedError(t *testing.T) {
	err := &utils.DataNotInitializedError{Alias: "Products"}
	if !strings.Contains(err.Error(), "Products") {
		t.Errorf("expected alias in error: %v", err)
	}
}

func TestNotValidIdentifierError(t *testing.T) {
	err := &utils.NotValidIdentifierError{Value: "123bad"}
	if !strings.Contains(err.Error(), "123bad") {
		t.Errorf("expected value in error: %v", err)
	}
}

func TestUnknownNameError(t *testing.T) {
	err := &utils.UnknownNameError{Value: "MissingField"}
	if !strings.Contains(err.Error(), "MissingField") {
		t.Errorf("expected value in error: %v", err)
	}
}

func TestGroupHeaderNoConditionError(t *testing.T) {
	err := &utils.GroupHeaderNoConditionError{Name: "GroupHeader1"}
	if !strings.Contains(err.Error(), "GroupHeader1") {
		t.Errorf("expected name in error: %v", err)
	}
}

func TestImageLoadError(t *testing.T) {
	cause := errors.New("file not found")
	err := &utils.ImageLoadError{Cause: cause}
	if !strings.Contains(err.Error(), "file not found") {
		t.Errorf("expected cause in error: %v", err)
	}
	if !errors.Is(err, cause) {
		t.Error("errors.Is should find wrapped cause")
	}
}

func TestErrorsAreErrors(t *testing.T) {
	errs := []error{
		&utils.DuplicateNameError{Name: "x"},
		&utils.AncestorError{Name: "x"},
		&utils.FileFormatError{},
		&utils.DecryptError{},
		&utils.CompilerError{Msg: "err"},
		&utils.ParentError{ParentType: "a", ChildType: "b"},
		&utils.ClassError{Name: "x"},
		&utils.DataTableError{Alias: "x"},
		&utils.DataNotInitializedError{Alias: "x"},
		&utils.NotValidIdentifierError{Value: "x"},
		&utils.UnknownNameError{Value: "x"},
		&utils.GroupHeaderNoConditionError{Name: "x"},
		&utils.ImageLoadError{Cause: errors.New("x")},
	}
	for _, err := range errs {
		if err.Error() == "" {
			t.Errorf("%T.Error() returned empty string", err)
		}
	}
}
