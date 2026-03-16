package reportpkg_test

import (
	"fmt"
	"testing"

	"github.com/andrewloable/go-fastreport/reportpkg"
)

// TestRegisterFunction_DoubleValue verifies that a custom Go callback
// registered via RegisterFunction is callable from a report expression.
func TestRegisterFunction_DoubleValue(t *testing.T) {
	r := reportpkg.NewReport()

	// Register a custom function that doubles its argument.
	r.RegisterFunction("DoubleValue", func(args []any) (any, error) {
		if len(args) != 1 {
			return nil, fmt.Errorf("DoubleValue: expected 1 argument, got %d", len(args))
		}
		switch v := args[0].(type) {
		case int:
			return v * 2, nil
		case float64:
			return v * 2, nil
		default:
			return nil, fmt.Errorf("DoubleValue: unsupported type %T", args[0])
		}
	})

	// Evaluate [DoubleValue(5)] — should return 10.
	val, err := r.Calc("[DoubleValue(5)]")
	if err != nil {
		t.Fatalf("Calc returned error: %v", err)
	}
	if val != 10 {
		t.Errorf("got %v (%T), want 10", val, val)
	}
}

// TestRegisterFunction_MultiArg verifies that custom functions can receive
// multiple arguments.
func TestRegisterFunction_MultiArg(t *testing.T) {
	r := reportpkg.NewReport()

	r.RegisterFunction("Add", func(args []any) (any, error) {
		if len(args) != 2 {
			return nil, fmt.Errorf("Add: expected 2 arguments, got %d", len(args))
		}
		a, ok1 := args[0].(int)
		b, ok2 := args[1].(int)
		if !ok1 || !ok2 {
			return nil, fmt.Errorf("Add: expected int arguments")
		}
		return a + b, nil
	})

	val, err := r.Calc("[Add(3, 7)]")
	if err != nil {
		t.Fatalf("Calc returned error: %v", err)
	}
	if val != 10 {
		t.Errorf("got %v, want 10", val)
	}
}

// TestRegisterFunction_Overwrite verifies that re-registering a name replaces
// the previous implementation.
func TestRegisterFunction_Overwrite(t *testing.T) {
	r := reportpkg.NewReport()

	r.RegisterFunction("GetValue", func(args []any) (any, error) {
		return "first", nil
	})
	r.RegisterFunction("GetValue", func(args []any) (any, error) {
		return "second", nil
	})

	val, err := r.Calc("[GetValue()]")
	if err != nil {
		t.Fatalf("Calc returned error: %v", err)
	}
	if val != "second" {
		t.Errorf("got %v, want second", val)
	}
}

// TestCustomFunctions_ReturnsCopy verifies that CustomFunctions returns an
// independent copy of the registry.
func TestCustomFunctions_ReturnsCopy(t *testing.T) {
	r := reportpkg.NewReport()

	r.RegisterFunction("Foo", func(args []any) (any, error) { return 42, nil })

	copy1 := r.CustomFunctions()
	if len(copy1) != 1 {
		t.Fatalf("expected 1 function, got %d", len(copy1))
	}

	// Mutate the copy; the report should be unaffected.
	delete(copy1, "Foo")

	copy2 := r.CustomFunctions()
	if len(copy2) != 1 {
		t.Errorf("mutation of returned map affected the report's registry")
	}
}

// TestRegisterFunction_CalcText verifies that custom functions work inside
// CalcText templates.
func TestRegisterFunction_CalcText(t *testing.T) {
	r := reportpkg.NewReport()

	r.RegisterFunction("Greet", func(args []any) (any, error) {
		return "Hello, World!", nil
	})

	text, err := r.CalcText("Message: [Greet()]")
	if err != nil {
		t.Fatalf("CalcText returned error: %v", err)
	}
	if text != "Message: Hello, World!" {
		t.Errorf("got %q, want %q", text, "Message: Hello, World!")
	}
}
