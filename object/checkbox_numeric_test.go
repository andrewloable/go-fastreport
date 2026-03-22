package object_test

import (
	"fmt"
	"testing"

	"github.com/andrewloable/go-fastreport/object"
)

// TestCheckBoxObject_GetData_NumericEval verifies that non-zero numeric values
// returned from the calc function cause Checked to be true, and zero values
// cause Checked to be false.
// Mirrors C# CheckBoxObject.GetDataShared (CheckBoxObject.cs line 346-361)
// which uses Variant semantics: numeric non-zero → checked, zero → unchecked.
func TestCheckBoxObject_GetData_NumericEval(t *testing.T) {
	tests := []struct {
		name        string
		returnVal   any
		returnErr   error
		wantChecked bool
	}{
		// bool cases
		{"bool true via DataColumn", true, nil, true},
		{"bool false via DataColumn", false, nil, false},

		// int cases
		{"int 1", int(1), nil, true},
		{"int 0", int(0), nil, false},
		{"int -1", int(-1), nil, true},
		{"int 42", int(42), nil, true},

		// int8 cases
		{"int8 non-zero", int8(5), nil, true},
		{"int8 zero", int8(0), nil, false},

		// int16 cases
		{"int16 non-zero", int16(100), nil, true},
		{"int16 zero", int16(0), nil, false},

		// int32 cases
		{"int32 non-zero", int32(7), nil, true},
		{"int32 zero", int32(0), nil, false},

		// int64 cases
		{"int64 non-zero", int64(999), nil, true},
		{"int64 zero", int64(0), nil, false},

		// uint cases
		{"uint non-zero", uint(3), nil, true},
		{"uint zero", uint(0), nil, false},

		// float32 cases
		{"float32 3.14", float32(3.14), nil, true},
		{"float32 0.0", float32(0.0), nil, false},

		// float64 cases
		{"float64 3.14", float64(3.14), nil, true},
		{"float64 0.0", float64(0.0), nil, false},
		{"float64 -1.5", float64(-1.5), nil, true},

		// string cases (legacy support)
		{"string true", "true", nil, true},
		{"string True", "True", nil, true},
		{"string 1", "1", nil, true},
		{"string false", "false", nil, false},
		{"string 0", "0", nil, false},
		{"string other", "yes", nil, false},

		// error / nil cases
		{"nil value", nil, nil, false},
		{"calc error", nil, fmt.Errorf("no column"), false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			c := object.NewCheckBoxObject()
			c.SetDataColumn("SomeColumn")

			calc := func(expr string) (any, error) {
				return tc.returnVal, tc.returnErr
			}
			c.GetData(calc)

			if c.Checked() != tc.wantChecked {
				t.Errorf("Checked() = %v, want %v (input: %T %v)",
					c.Checked(), tc.wantChecked, tc.returnVal, tc.returnVal)
			}
		})
	}
}

// TestCheckBoxObject_GetData_Expression verifies the expression path strictly
// uses bool semantics only (per C# CheckBoxObject.cs line 356-358).
func TestCheckBoxObject_GetData_Expression(t *testing.T) {
	tests := []struct {
		name        string
		returnVal   any
		wantChecked bool
	}{
		{"bool true", true, true},
		{"bool false", false, false},
		// Numeric values must NOT affect Expression path (C# strictly bool).
		{"int 1 ignored", int(1), false},
		{"float64 3.14 ignored", float64(3.14), false},
		{"string true ignored", "true", false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			c := object.NewCheckBoxObject()
			c.SetExpression("[SomeExpr]")

			calc := func(expr string) (any, error) {
				return tc.returnVal, nil
			}
			c.GetData(calc)

			if c.Checked() != tc.wantChecked {
				t.Errorf("Checked() = %v, want %v (input: %T %v)",
					c.Checked(), tc.wantChecked, tc.returnVal, tc.returnVal)
			}
		})
	}
}

// TestCheckBoxObject_GetData_HideIfUnchecked verifies that when HideIfUnchecked
// is set and the data evaluates to false, Visible is set to false.
func TestCheckBoxObject_GetData_HideIfUnchecked(t *testing.T) {
	c := object.NewCheckBoxObject()
	c.SetDataColumn("Col")
	c.SetHideIfUnchecked(true)

	// Value evaluates to 0 → unchecked → should hide.
	c.GetData(func(expr string) (any, error) { return int(0), nil })
	if c.Visible() {
		t.Error("Visible should be false when Checked=false and HideIfUnchecked=true")
	}

	// Reset visibility and test non-zero → checked → should stay visible.
	c2 := object.NewCheckBoxObject()
	c2.SetDataColumn("Col")
	c2.SetHideIfUnchecked(true)
	c2.GetData(func(expr string) (any, error) { return int(1), nil })
	if !c2.Visible() {
		t.Error("Visible should be true when Checked=true")
	}
	if !c2.Checked() {
		t.Error("Checked should be true for int(1)")
	}
}
