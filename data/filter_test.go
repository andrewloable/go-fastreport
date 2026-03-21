package data_test

import (
	"testing"
	"time"

	"github.com/andrewloable/go-fastreport/data"
)

func TestDataSourceFilter_Empty(t *testing.T) {
	f := data.NewDataSourceFilter()
	// Empty filter always matches.
	if !f.ValueMatch(42) {
		t.Error("empty filter should match any value")
	}
	if !f.ValueMatch(nil) {
		t.Error("empty filter should match nil")
	}
}

func TestDataSourceFilter_Len(t *testing.T) {
	f := data.NewDataSourceFilter()
	f.Add(10, data.FilterEqual)
	f.Add(20, data.FilterLessThan)
	if f.Len() != 2 {
		t.Errorf("Len = %d, want 2", f.Len())
	}
}

func TestDataSourceFilter_Clear(t *testing.T) {
	f := data.NewDataSourceFilter()
	f.Add(1, data.FilterEqual)
	f.Clear()
	if f.Len() != 0 {
		t.Errorf("after Clear, Len = %d, want 0", f.Len())
	}
	// Should still match after clear.
	if !f.ValueMatch(99) {
		t.Error("cleared filter should match any value")
	}
}

func TestDataSourceFilter_Remove(t *testing.T) {
	f := data.NewDataSourceFilter()
	fe1 := f.Add(1, data.FilterEqual)
	f.Add(2, data.FilterEqual)
	f.Remove(fe1)
	if f.Len() != 1 {
		t.Errorf("after Remove, Len = %d, want 1", f.Len())
	}
}

func TestDataSourceFilter_IntEqual(t *testing.T) {
	f := data.NewDataSourceFilter()
	f.Add(42, data.FilterEqual)
	if !f.ValueMatch(42) {
		t.Error("42 == 42 should match")
	}
	if f.ValueMatch(41) {
		t.Error("41 == 42 should not match")
	}
}

func TestDataSourceFilter_IntNotEqual(t *testing.T) {
	f := data.NewDataSourceFilter()
	f.Add(10, data.FilterNotEqual)
	if !f.ValueMatch(11) {
		t.Error("11 != 10 should match")
	}
	if f.ValueMatch(10) {
		t.Error("10 != 10 should not match")
	}
}

func TestDataSourceFilter_IntLessThan(t *testing.T) {
	f := data.NewDataSourceFilter()
	f.Add(10, data.FilterLessThan)
	if !f.ValueMatch(9) {
		t.Error("9 < 10 should match")
	}
	if f.ValueMatch(10) {
		t.Error("10 < 10 should not match")
	}
	if f.ValueMatch(11) {
		t.Error("11 < 10 should not match")
	}
}

func TestDataSourceFilter_IntLessThanOrEqual(t *testing.T) {
	f := data.NewDataSourceFilter()
	f.Add(10, data.FilterLessThanOrEqual)
	if !f.ValueMatch(10) {
		t.Error("10 <= 10 should match")
	}
	if !f.ValueMatch(9) {
		t.Error("9 <= 10 should match")
	}
	if f.ValueMatch(11) {
		t.Error("11 <= 10 should not match")
	}
}

func TestDataSourceFilter_IntGreaterThan(t *testing.T) {
	f := data.NewDataSourceFilter()
	f.Add(5, data.FilterGreaterThan)
	if !f.ValueMatch(6) {
		t.Error("6 > 5 should match")
	}
	if f.ValueMatch(5) {
		t.Error("5 > 5 should not match")
	}
}

func TestDataSourceFilter_IntGreaterThanOrEqual(t *testing.T) {
	f := data.NewDataSourceFilter()
	f.Add(5, data.FilterGreaterThanOrEqual)
	if !f.ValueMatch(5) {
		t.Error("5 >= 5 should match")
	}
	if !f.ValueMatch(6) {
		t.Error("6 >= 5 should match")
	}
	if f.ValueMatch(4) {
		t.Error("4 >= 5 should not match")
	}
}

func TestDataSourceFilter_Float(t *testing.T) {
	f := data.NewDataSourceFilter()
	f.Add(3.14, data.FilterGreaterThan)
	if !f.ValueMatch(3.15) {
		t.Error("3.15 > 3.14 should match")
	}
	if f.ValueMatch(3.14) {
		t.Error("3.14 > 3.14 should not match")
	}
}

func TestDataSourceFilter_StringEqual(t *testing.T) {
	f := data.NewDataSourceFilter()
	f.Add("hello", data.FilterEqual)
	if !f.ValueMatch("hello") {
		t.Error("'hello' == 'hello' should match")
	}
	if f.ValueMatch("world") {
		t.Error("'world' == 'hello' should not match")
	}
}

func TestDataSourceFilter_StringContains(t *testing.T) {
	f := data.NewDataSourceFilter()
	f.Add("ell", data.FilterContains)
	if !f.ValueMatch("hello") {
		t.Error("'hello' contains 'ell' should match")
	}
	if f.ValueMatch("world") {
		t.Error("'world' contains 'ell' should not match")
	}
}

func TestDataSourceFilter_StringNotContains(t *testing.T) {
	f := data.NewDataSourceFilter()
	f.Add("xyz", data.FilterNotContains)
	if !f.ValueMatch("hello") {
		t.Error("'hello' not-contains 'xyz' should match")
	}
	if f.ValueMatch("xyz123") {
		t.Error("'xyz123' not-contains 'xyz' should not match")
	}
}

func TestDataSourceFilter_StringStartsWith(t *testing.T) {
	f := data.NewDataSourceFilter()
	f.Add("he", data.FilterStartsWith)
	if !f.ValueMatch("hello") {
		t.Error("'hello' starts-with 'he' should match")
	}
	if f.ValueMatch("world") {
		t.Error("'world' starts-with 'he' should not match")
	}
}

func TestDataSourceFilter_StringNotStartsWith(t *testing.T) {
	f := data.NewDataSourceFilter()
	f.Add("he", data.FilterNotStartsWith)
	if f.ValueMatch("hello") {
		t.Error("'hello' not-starts-with 'he' should not match")
	}
	if !f.ValueMatch("world") {
		t.Error("'world' not-starts-with 'he' should match")
	}
}

func TestDataSourceFilter_StringEndsWith(t *testing.T) {
	f := data.NewDataSourceFilter()
	f.Add("lo", data.FilterEndsWith)
	if !f.ValueMatch("hello") {
		t.Error("'hello' ends-with 'lo' should match")
	}
	if f.ValueMatch("world") {
		t.Error("'world' ends-with 'lo' should not match")
	}
}

func TestDataSourceFilter_StringNotEndsWith(t *testing.T) {
	f := data.NewDataSourceFilter()
	f.Add("lo", data.FilterNotEndsWith)
	if f.ValueMatch("hello") {
		t.Error("'hello' not-ends-with 'lo' should not match")
	}
	if !f.ValueMatch("world") {
		t.Error("'world' not-ends-with 'lo' should match")
	}
}

func TestDataSourceFilter_StringSet_Equal(t *testing.T) {
	// []string value triggers set-membership logic.
	f := data.NewDataSourceFilter()
	f.Add([]string{"Alice", "Bob", "Charlie"}, data.FilterEqual)
	if !f.ValueMatch("Alice") {
		t.Error("'Alice' in set should match Equal")
	}
	if !f.ValueMatch("Bob") {
		t.Error("'Bob' in set should match Equal")
	}
	if f.ValueMatch("Dave") {
		t.Error("'Dave' not in set should not match Equal")
	}
}

func TestDataSourceFilter_StringSet_NotEqual(t *testing.T) {
	f := data.NewDataSourceFilter()
	f.Add([]string{"Alice", "Bob"}, data.FilterNotEqual)
	if f.ValueMatch("Alice") {
		t.Error("'Alice' in set should not match NotEqual")
	}
	if !f.ValueMatch("Charlie") {
		t.Error("'Charlie' not in set should match NotEqual")
	}
}

func TestDataSourceFilter_TimeRange(t *testing.T) {
	start := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(2024, 1, 31, 0, 0, 0, 0, time.UTC)
	rng := [2]time.Time{start, end}

	f := data.NewDataSourceFilter()
	f.Add(rng, data.FilterEqual)

	inRange := time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC)
	if !f.ValueMatch(inRange) {
		t.Error("date in range should match")
	}

	before := time.Date(2023, 12, 31, 0, 0, 0, 0, time.UTC)
	if f.ValueMatch(before) {
		t.Error("date before range should not match")
	}

	after := time.Date(2024, 2, 1, 0, 0, 0, 0, time.UTC)
	if f.ValueMatch(after) {
		t.Error("date after range should not match")
	}
}

func TestDataSourceFilter_TimeEqual(t *testing.T) {
	t1 := time.Date(2024, 6, 15, 0, 0, 0, 0, time.UTC)
	t2 := time.Date(2024, 6, 15, 0, 0, 0, 0, time.UTC)
	f := data.NewDataSourceFilter()
	f.Add(t1, data.FilterEqual)
	if !f.ValueMatch(t2) {
		t.Error("same time should match Equal")
	}
}

func TestDataSourceFilter_IncomparableTypes(t *testing.T) {
	// Incomparable types (e.g. bool vs int) should not match.
	f := data.NewDataSourceFilter()
	f.Add(42, data.FilterEqual)
	if f.ValueMatch(true) {
		t.Error("bool value against int filter should not match")
	}
}

func TestDataSourceFilter_NilValue_Fails(t *testing.T) {
	f := data.NewDataSourceFilter()
	f.Add(10, data.FilterEqual)
	if f.ValueMatch(nil) {
		t.Error("nil should not match a non-nil filter")
	}
}

func TestDataSourceFilter_MultipleConditions_AND(t *testing.T) {
	// Both conditions must match.
	f := data.NewDataSourceFilter()
	f.Add(0, data.FilterGreaterThan)   // > 0
	f.Add(100, data.FilterLessThan)    // < 100
	if !f.ValueMatch(50) {
		t.Error("50 is > 0 and < 100, should match")
	}
	if f.ValueMatch(0) {
		t.Error("0 is not > 0, should not match")
	}
	if f.ValueMatch(100) {
		t.Error("100 is not < 100, should not match")
	}
}

// TestDataSourceFilter_StringSet_NilValue verifies that nil is treated as ""
// in the string-set branch, matching C# behaviour (value == null ? "" : value.ToString()).
func TestDataSourceFilter_StringSet_NilValue(t *testing.T) {
	// A set that contains the empty string should match nil.
	f := data.NewDataSourceFilter()
	f.Add([]string{"", "Alice"}, data.FilterEqual)
	if !f.ValueMatch(nil) {
		t.Error("nil should match string-set containing empty string")
	}
	if !f.ValueMatch("Alice") {
		t.Error("'Alice' should match string-set containing 'Alice'")
	}

	// A set that does NOT contain the empty string should not match nil.
	f2 := data.NewDataSourceFilter()
	f2.Add([]string{"Alice", "Bob"}, data.FilterEqual)
	if f2.ValueMatch(nil) {
		t.Error("nil should not match string-set that does not contain empty string")
	}
}

// TestDataSourceFilter_StringSet_NilValue_NotEqual verifies NotEqual with nil value.
func TestDataSourceFilter_StringSet_NilValue_NotEqual(t *testing.T) {
	// Set without empty string: nil maps to "", not in set → NotEqual matches.
	f := data.NewDataSourceFilter()
	f.Add([]string{"Alice", "Bob"}, data.FilterNotEqual)
	if !f.ValueMatch(nil) {
		t.Error("nil (maps to '') not in set — NotEqual should match")
	}
}

// TestDataSourceFilter_TimeEqual_DateOnly verifies C# time-stripping behaviour:
// when the filter value has no time component, the data value's time is stripped
// before comparison, so filter "Equal 2024-06-15" matches "2024-06-15 14:30:00".
func TestDataSourceFilter_TimeEqual_DateOnly(t *testing.T) {
	filterDate := time.Date(2024, 6, 15, 0, 0, 0, 0, time.UTC) // no time component
	f := data.NewDataSourceFilter()
	f.Add(filterDate, data.FilterEqual)

	// Data value with a time component — should still match because filter has no time.
	dataWithTime := time.Date(2024, 6, 15, 14, 30, 0, 0, time.UTC)
	if !f.ValueMatch(dataWithTime) {
		t.Error("filter date-only Equal should match data value with time on same day")
	}

	// Different day — should not match.
	differentDay := time.Date(2024, 6, 16, 0, 0, 0, 0, time.UTC)
	if f.ValueMatch(differentDay) {
		t.Error("filter date-only Equal should not match a different day")
	}
}

// TestDataSourceFilter_TimeEqual_WithTime verifies that when the filter value DOES
// have a time component, comparison is exact (time is not stripped).
func TestDataSourceFilter_TimeEqual_WithTime(t *testing.T) {
	filterDateTime := time.Date(2024, 6, 15, 14, 30, 0, 0, time.UTC) // has time
	f := data.NewDataSourceFilter()
	f.Add(filterDateTime, data.FilterEqual)

	// Exact match should work.
	if !f.ValueMatch(filterDateTime) {
		t.Error("exact DateTime should match")
	}

	// Same date but different time — should NOT match (element has time component).
	differentTime := time.Date(2024, 6, 15, 15, 0, 0, 0, time.UTC)
	if f.ValueMatch(differentTime) {
		t.Error("different time should not match when filter has time component")
	}
}

// TestDataSourceFilter_TimeNotEqual_DateOnly verifies NotEqual with date-only filter.
func TestDataSourceFilter_TimeNotEqual_DateOnly(t *testing.T) {
	filterDate := time.Date(2024, 6, 15, 0, 0, 0, 0, time.UTC)
	f := data.NewDataSourceFilter()
	f.Add(filterDate, data.FilterNotEqual)

	// Same day with time — should NOT match NotEqual (stripped time makes them equal).
	dataWithTime := time.Date(2024, 6, 15, 14, 30, 0, 0, time.UTC)
	if f.ValueMatch(dataWithTime) {
		t.Error("date-only NotEqual: same day with time should not match (equal after strip)")
	}

	// Different day — should match NotEqual.
	differentDay := time.Date(2024, 6, 16, 0, 0, 0, 0, time.UTC)
	if !f.ValueMatch(differentDay) {
		t.Error("date-only NotEqual: different day should match")
	}
}

// TestDataSourceFilter_TimeLessThan_DateOnly verifies LessThan with date-only filter.
func TestDataSourceFilter_TimeLessThan_DateOnly(t *testing.T) {
	filterDate := time.Date(2024, 6, 15, 0, 0, 0, 0, time.UTC)
	f := data.NewDataSourceFilter()
	f.Add(filterDate, data.FilterLessThan)

	earlier := time.Date(2024, 6, 14, 23, 59, 59, 0, time.UTC)
	if !f.ValueMatch(earlier) {
		t.Error("date before filter date should match LessThan (date-only)")
	}

	// Same day with time — after stripping becomes equal, so NOT less than.
	sameDay := time.Date(2024, 6, 15, 10, 0, 0, 0, time.UTC)
	if f.ValueMatch(sameDay) {
		t.Error("same day (after time strip) should not match LessThan")
	}
}

// TestDataSourceFilter_TimeRange_SliceForm verifies that []time.Time{start,end} also
// works as a date range, in addition to [2]time.Time.
func TestDataSourceFilter_TimeRange_SliceForm(t *testing.T) {
	start := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(2024, 1, 31, 0, 0, 0, 0, time.UTC)
	rng := []time.Time{start, end}

	f := data.NewDataSourceFilter()
	f.Add(rng, data.FilterEqual)

	inRange := time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC)
	if !f.ValueMatch(inRange) {
		t.Error("date in range should match (slice form)")
	}

	before := time.Date(2023, 12, 31, 0, 0, 0, 0, time.UTC)
	if f.ValueMatch(before) {
		t.Error("date before range should not match (slice form)")
	}

	// End-of-day on the last day of the range must match (AddDate(0,0,1) makes end exclusive).
	endOfRange := time.Date(2024, 1, 31, 23, 59, 59, 0, time.UTC)
	if !f.ValueMatch(endOfRange) {
		t.Error("end-of-day on last day should match (slice form)")
	}

	after := time.Date(2024, 2, 1, 0, 0, 0, 0, time.UTC)
	if f.ValueMatch(after) {
		t.Error("date after range should not match (slice form)")
	}
}

// TestDataSourceFilter_TimeRange_EndDayInclusive verifies that the last day of the
// range is fully included (any time on that day matches).
func TestDataSourceFilter_TimeRange_EndDayInclusive(t *testing.T) {
	start := time.Date(2024, 3, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(2024, 3, 31, 0, 0, 0, 0, time.UTC)
	rng := [2]time.Time{start, end}

	f := data.NewDataSourceFilter()
	f.Add(rng, data.FilterEqual)

	// Last moment of the last day (23:59:59) must be included.
	lastMoment := time.Date(2024, 3, 31, 23, 59, 59, 0, time.UTC)
	if !f.ValueMatch(lastMoment) {
		t.Error("23:59:59 on end day should be within range (AddDays(1) makes it exclusive)")
	}

	// Midnight of the day AFTER end must NOT match.
	dayAfter := time.Date(2024, 4, 1, 0, 0, 0, 0, time.UTC)
	if f.ValueMatch(dayAfter) {
		t.Error("midnight of the day after end should not match")
	}
}

// TestDataSourceFilter_TimeVsNonTime verifies that a DateTime value against a
// non-DateTime (scalar) element returns false, matching C# line 163.
func TestDataSourceFilter_TimeVsNonTime(t *testing.T) {
	f := data.NewDataSourceFilter()
	f.Add(42, data.FilterEqual) // int element, not DateTime
	tv := time.Date(2024, 6, 15, 0, 0, 0, 0, time.UTC)
	if f.ValueMatch(tv) {
		t.Error("DateTime value vs non-DateTime element should not match")
	}
}

// TestDataSourceFilter_StringSet_Contains verifies Contains alias on string sets.
func TestDataSourceFilter_StringSet_Contains(t *testing.T) {
	f := data.NewDataSourceFilter()
	f.Add([]string{"red", "green", "blue"}, data.FilterContains)
	if !f.ValueMatch("red") {
		t.Error("'red' in set should match Contains")
	}
	if f.ValueMatch("yellow") {
		t.Error("'yellow' not in set should not match Contains")
	}
}

// TestDataSourceFilter_StringSet_NotContains verifies NotContains alias on string sets.
func TestDataSourceFilter_StringSet_NotContains(t *testing.T) {
	f := data.NewDataSourceFilter()
	f.Add([]string{"red", "green", "blue"}, data.FilterNotContains)
	if f.ValueMatch("red") {
		t.Error("'red' in set should not match NotContains")
	}
	if !f.ValueMatch("yellow") {
		t.Error("'yellow' not in set should match NotContains")
	}
}

// TestDataSourceFilter_StringSet_UnsupportedOp verifies that unsupported operations
// on a string set (e.g. LessThan) return false.
func TestDataSourceFilter_StringSet_UnsupportedOp(t *testing.T) {
	f := data.NewDataSourceFilter()
	f.Add([]string{"Alice"}, data.FilterLessThan)
	if f.ValueMatch("Alice") {
		t.Error("LessThan on string-set should always return false")
	}
	if f.ValueMatch("Bob") {
		t.Error("LessThan on string-set should always return false")
	}
}

func TestFilterOperationConstants(t *testing.T) {
	ops := []data.FilterOperation{
		data.FilterEqual,
		data.FilterNotEqual,
		data.FilterLessThan,
		data.FilterLessThanOrEqual,
		data.FilterGreaterThan,
		data.FilterGreaterThanOrEqual,
		data.FilterContains,
		data.FilterNotContains,
		data.FilterStartsWith,
		data.FilterNotStartsWith,
		data.FilterEndsWith,
		data.FilterNotEndsWith,
	}
	seen := make(map[data.FilterOperation]bool)
	for _, op := range ops {
		if seen[op] {
			t.Errorf("duplicate FilterOperation value %d", op)
		}
		seen[op] = true
	}
	if len(ops) != 12 {
		t.Errorf("expected 12 FilterOperation values, got %d", len(ops))
	}
}
