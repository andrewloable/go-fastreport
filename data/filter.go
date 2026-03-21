package data

import (
	"fmt"
	"strings"
	"time"
)

// FilterOperation specifies how to compare a data value against a filter value.
type FilterOperation int

const (
	// FilterEqual matches when data value equals filter value.
	FilterEqual FilterOperation = iota
	// FilterNotEqual matches when data value does not equal filter value.
	FilterNotEqual
	// FilterLessThan matches when data value is less than filter value.
	FilterLessThan
	// FilterLessThanOrEqual matches when data value is ≤ filter value.
	FilterLessThanOrEqual
	// FilterGreaterThan matches when data value is greater than filter value.
	FilterGreaterThan
	// FilterGreaterThanOrEqual matches when data value is ≥ filter value.
	FilterGreaterThanOrEqual
	// FilterContains matches when data string contains filter string.
	FilterContains
	// FilterNotContains matches when data string does not contain filter string.
	FilterNotContains
	// FilterStartsWith matches when data string starts with filter string.
	FilterStartsWith
	// FilterNotStartsWith matches when data string does not start with filter string.
	FilterNotStartsWith
	// FilterEndsWith matches when data string ends with filter string.
	FilterEndsWith
	// FilterNotEndsWith matches when data string does not end with filter string.
	FilterNotEndsWith
)

// FilterElement represents a single filter condition.
type FilterElement struct {
	// Value is the filter comparison value.
	Value any
	// Operation is the comparison operator.
	Operation FilterOperation
	// stringSet is a fast-lookup set built when Value is []string.
	stringSet map[string]struct{}
}

// newFilterElement creates a FilterElement and pre-computes the string set
// when value is a []string, matching the C# SortedList behaviour.
func newFilterElement(value any, op FilterOperation) *FilterElement {
	fe := &FilterElement{Value: value, Operation: op}
	if ss, ok := value.([]string); ok {
		fe.stringSet = make(map[string]struct{}, len(ss))
		for _, s := range ss {
			fe.stringSet[s] = struct{}{}
		}
	}
	return fe
}

// DataSourceFilter holds an ordered list of filter conditions.
// All conditions must match for a value to pass (AND semantics).
// It is the Go equivalent of FastReport.Data.DataSourceFilter.
type DataSourceFilter struct {
	elements []*FilterElement
}

// NewDataSourceFilter creates an empty DataSourceFilter.
func NewDataSourceFilter() *DataSourceFilter {
	return &DataSourceFilter{}
}

// Add appends a filter condition and returns the new FilterElement.
func (f *DataSourceFilter) Add(value any, op FilterOperation) *FilterElement {
	fe := newFilterElement(value, op)
	f.elements = append(f.elements, fe)
	return fe
}

// Remove removes the given FilterElement from the filter.
func (f *DataSourceFilter) Remove(fe *FilterElement) {
	for i, e := range f.elements {
		if e == fe {
			f.elements = append(f.elements[:i], f.elements[i+1:]...)
			return
		}
	}
}

// Clear removes all filter conditions.
func (f *DataSourceFilter) Clear() {
	f.elements = f.elements[:0]
}

// Len returns the number of filter conditions.
func (f *DataSourceFilter) Len() int { return len(f.elements) }

// ValueMatch returns true when value satisfies all filter conditions.
// An empty filter always returns true.
func (f *DataSourceFilter) ValueMatch(value any) bool {
	for _, elem := range f.elements {
		if !elem.matches(value) {
			return false
		}
	}
	return true
}

// matches checks whether value satisfies this single FilterElement.
func (fe *FilterElement) matches(value any) bool {
	// --- string-set branch (value is []string list) ---
	// C# DataSourceFilter.cs line 110: value == null ? "" : value.ToString()
	if fe.stringSet != nil {
		var s string
		if value == nil {
			s = ""
		} else {
			s = fmt.Sprint(value)
		}
		_, inSet := fe.stringSet[s]
		switch fe.Operation {
		case FilterEqual, FilterContains:
			return inSet
		case FilterNotEqual, FilterNotContains:
			return !inSet
		default:
			return false
		}
	}

	// --- time.Time range branch (two-element [2]time.Time or []time.Time) ---
	// C# DataSourceFilter.cs: checks DateTime[] of length 2 for a date range.
	// AddDays(1) makes end exclusive so the whole end-date day is included.
	if tv, ok := value.(time.Time); ok {
		var rngStart, rngEnd time.Time
		var isRange bool
		if rng, ok := fe.Value.([2]time.Time); ok {
			rngStart, rngEnd, isRange = rng[0], rng[1], true
		} else if rng, ok := fe.Value.([]time.Time); ok && len(rng) == 2 {
			rngStart, rngEnd, isRange = rng[0], rng[1], true
		}
		if isRange {
			end := rngEnd.AddDate(0, 0, 1)
			inRange := (tv.Equal(rngStart) || tv.After(rngStart)) && tv.Before(end)
			switch fe.Operation {
			case FilterEqual, FilterContains:
				return inRange
			case FilterNotEqual, FilterNotContains:
				return !inRange
			default:
				// fall through to general comparison below
			}
		}
	}

	// --- DateTime scalar comparison with optional time-stripping ---
	// C# DataSourceFilter.cs lines 151–164: when value is DateTime and element is
	// DateTime, check if the element has a time component (TimeOfDay.Ticks != 0).
	// If the element has NO time component, strip the time portion from value before
	// comparing so that filter "Equal 2024-06-15" matches "2024-06-15 14:30:00".
	// If value is DateTime but element is NOT DateTime, return false.
	if tv, ok := value.(time.Time); ok {
		ev, isTime := fe.Value.(time.Time)
		if !isTime {
			// element is not a DateTime — incomparable, C# returns false here
			return false
		}
		// Strip time from data value when element has no time component.
		// C# TimeOfDay.Ticks != 0 means the element DOES have a time component.
		if ev.Hour() == 0 && ev.Minute() == 0 && ev.Second() == 0 && ev.Nanosecond() == 0 {
			// Element has no time — compare date only.
			tv = time.Date(tv.Year(), tv.Month(), tv.Day(), 0, 0, 0, 0, tv.Location())
		}
		cmp, ok := compare(tv, ev)
		if !ok {
			return false
		}
		switch fe.Operation {
		case FilterEqual:
			return cmp == 0
		case FilterNotEqual:
			return cmp != 0
		case FilterLessThan:
			return cmp < 0
		case FilterLessThanOrEqual:
			return cmp <= 0
		case FilterGreaterThan:
			return cmp > 0
		case FilterGreaterThanOrEqual:
			return cmp >= 0
		}
		return false
	}

	// --- string-specific operations ---
	if sv, ok := value.(string); ok {
		ev := fmt.Sprint(fe.Value)
		switch fe.Operation {
		case FilterContains:
			return strings.Contains(sv, ev)
		case FilterNotContains:
			return !strings.Contains(sv, ev)
		case FilterStartsWith:
			return strings.HasPrefix(sv, ev)
		case FilterNotStartsWith:
			return !strings.HasPrefix(sv, ev)
		case FilterEndsWith:
			return strings.HasSuffix(sv, ev)
		case FilterNotEndsWith:
			return !strings.HasSuffix(sv, ev)
		}
		// fall through to comparable for Equal/NotEqual/Less/Greater on strings
	}

	// --- general comparable branch ---
	cmp, ok := compare(value, fe.Value)
	if !ok {
		return false
	}
	switch fe.Operation {
	case FilterEqual:
		return cmp == 0
	case FilterNotEqual:
		return cmp != 0
	case FilterLessThan:
		return cmp < 0
	case FilterLessThanOrEqual:
		return cmp <= 0
	case FilterGreaterThan:
		return cmp > 0
	case FilterGreaterThanOrEqual:
		return cmp >= 0
	}
	return false
}

// compare compares a and b. Returns (result, true) where result is -1/0/1,
// or (0, false) when the types are incomparable.
func compare(a, b any) (int, bool) {
	// Both must be non-nil.
	if a == nil || b == nil {
		return 0, false
	}

	// Use fmt.Sprintf as a last resort for ordering of strings.
	switch av := a.(type) {
	case int:
		bv, err := toInt64(b)
		if err != nil {
			return 0, false
		}
		return cmpInt64(int64(av), bv), true
	case int64:
		bv, err := toInt64(b)
		if err != nil {
			return 0, false
		}
		return cmpInt64(av, bv), true
	case int32:
		bv, err := toInt64(b)
		if err != nil {
			return 0, false
		}
		return cmpInt64(int64(av), bv), true
	case float64:
		bv, err := toFloat64(b)
		if err != nil {
			return 0, false
		}
		return cmpFloat64(av, bv), true
	case float32:
		bv, err := toFloat64(b)
		if err != nil {
			return 0, false
		}
		return cmpFloat64(float64(av), bv), true
	case string:
		bv, ok := b.(string)
		if !ok {
			return 0, false
		}
		return strings.Compare(av, bv), true
	case time.Time:
		bv, ok := b.(time.Time)
		if !ok {
			return 0, false
		}
		if av.Before(bv) {
			return -1, true
		} else if av.After(bv) {
			return 1, true
		}
		return 0, true
	case bool:
		bv, ok := b.(bool)
		if !ok {
			return 0, false
		}
		ai, bi := 0, 0
		if av {
			ai = 1
		}
		if bv {
			bi = 1
		}
		return ai - bi, true
	}
	return 0, false
}

func cmpInt64(a, b int64) int {
	if a < b {
		return -1
	} else if a > b {
		return 1
	}
	return 0
}

func cmpFloat64(a, b float64) int {
	if a < b {
		return -1
	} else if a > b {
		return 1
	}
	return 0
}

func toInt64(v any) (int64, error) {
	switch x := v.(type) {
	case int:
		return int64(x), nil
	case int32:
		return int64(x), nil
	case int64:
		return x, nil
	case float32:
		return int64(x), nil
	case float64:
		return int64(x), nil
	}
	return 0, fmt.Errorf("cannot convert %T to int64", v)
}

func toFloat64(v any) (float64, error) {
	switch x := v.(type) {
	case float64:
		return x, nil
	case float32:
		return float64(x), nil
	case int:
		return float64(x), nil
	case int8:
		return float64(x), nil
	case int16:
		return float64(x), nil
	case int32:
		return float64(x), nil
	case int64:
		return float64(x), nil
	case uint:
		return float64(x), nil
	case uint8:
		return float64(x), nil
	case uint16:
		return float64(x), nil
	case uint32:
		return float64(x), nil
	case uint64:
		return float64(x), nil
	}
	return 0, fmt.Errorf("cannot convert %T to float64", v)
}
