package utils

import "fmt"

// Value is a flexible value container analogous to the C# object/any used for
// dynamic data binding values in data sources and expressions.
// A zero Value holds nil.
type Value struct {
	v any
}

// NewValue creates a Value that wraps v.
func NewValue(v any) Value {
	return Value{v: v}
}

// Raw returns the underlying value without any conversion.
func (val Value) Raw() any {
	return val.v
}

// IsNil reports whether the Value holds nil.
func (val Value) IsNil() bool {
	return val.v == nil
}

// String returns a human-readable representation of the value using
// fmt.Sprintf("%v", v). Returns an empty string when IsNil is true.
func (val Value) String() string {
	if val.v == nil {
		return ""
	}
	return fmt.Sprintf("%v", val.v)
}

// Int attempts to extract an integer from the value.
// It handles int, int8, int16, int32, int64, uint, uint8, uint16, uint32,
// uint64, float32, and float64, truncating fractional parts as needed.
// The second return value is false when conversion is not possible.
func (val Value) Int() (int, bool) {
	switch v := val.v.(type) {
	case int:
		return v, true
	case int8:
		return int(v), true
	case int16:
		return int(v), true
	case int32:
		return int(v), true
	case int64:
		return int(v), true
	case uint:
		return int(v), true
	case uint8:
		return int(v), true
	case uint16:
		return int(v), true
	case uint32:
		return int(v), true
	case uint64:
		return int(v), true
	case float32:
		return int(v), true
	case float64:
		return int(v), true
	}
	return 0, false
}

// Float64 attempts to extract a float64 from the value.
// It handles all integer and float types.
// The second return value is false when conversion is not possible.
func (val Value) Float64() (float64, bool) {
	switch v := val.v.(type) {
	case float64:
		return v, true
	case float32:
		return float64(v), true
	case int:
		return float64(v), true
	case int8:
		return float64(v), true
	case int16:
		return float64(v), true
	case int32:
		return float64(v), true
	case int64:
		return float64(v), true
	case uint:
		return float64(v), true
	case uint8:
		return float64(v), true
	case uint16:
		return float64(v), true
	case uint32:
		return float64(v), true
	case uint64:
		return float64(v), true
	}
	return 0, false
}

// Bool attempts to extract a bool from the value.
// The second return value is false when the underlying type is not bool.
func (val Value) Bool() (bool, bool) {
	b, ok := val.v.(bool)
	return b, ok
}

// Equals reports whether val and other contain equal values using Go's ==
// operator on the underlying any values. For slice/map types that are not
// comparable, Equals returns false rather than panicking.
func (val Value) Equals(other Value) (result bool) {
	defer func() {
		if r := recover(); r != nil {
			result = false
		}
	}()
	result = val.v == other.v
	return result
}
