package data

// data_compareany_internal_test.go — internal (package data) tests for compareAny
// to cover the bool-equal branch (return 0) that SortRows cannot reliably trigger.

import "testing"

func TestCompareAny_Int(t *testing.T) {
	if compareAny(1, 2) >= 0 {
		t.Error("compareAny(1, 2) should be < 0")
	}
	if compareAny(2, 1) <= 0 {
		t.Error("compareAny(2, 1) should be > 0")
	}
	if compareAny(1, 1) != 0 {
		t.Error("compareAny(1, 1) should be 0")
	}
}

func TestCompareAny_Int64(t *testing.T) {
	if compareAny(int64(10), int64(20)) >= 0 {
		t.Error("compareAny(int64 10 < 20) should be < 0")
	}
	if compareAny(int64(20), int64(10)) <= 0 {
		t.Error("compareAny(int64 20 > 10) should be > 0")
	}
	if compareAny(int64(5), int64(5)) != 0 {
		t.Error("compareAny(int64 5 == 5) should be 0")
	}
}

func TestCompareAny_Float64(t *testing.T) {
	if compareAny(float64(1.0), float64(2.0)) >= 0 {
		t.Error("compareAny(float64 1.0 < 2.0) should be < 0")
	}
	if compareAny(float64(3.14), float64(3.14)) != 0 {
		t.Error("compareAny(float64 3.14 == 3.14) should be 0")
	}
}

func TestCompareAny_Float32(t *testing.T) {
	if compareAny(float32(1.0), float32(2.0)) >= 0 {
		t.Error("compareAny(float32 1.0 < 2.0) should be < 0")
	}
}

func TestCompareAny_String(t *testing.T) {
	if compareAny("a", "b") >= 0 {
		t.Error("compareAny('a', 'b') should be < 0")
	}
	if compareAny("z", "a") <= 0 {
		t.Error("compareAny('z', 'a') should be > 0")
	}
	if compareAny("x", "x") != 0 {
		t.Error("compareAny('x', 'x') should be 0")
	}
}

func TestCompareAny_Bool_Equal(t *testing.T) {
	// av == b.(bool) → return 0
	if compareAny(true, true) != 0 {
		t.Error("compareAny(true, true) should be 0")
	}
	if compareAny(false, false) != 0 {
		t.Error("compareAny(false, false) should be 0")
	}
}

func TestCompareAny_Bool_FalseTrue(t *testing.T) {
	// !av → return -1
	if compareAny(false, true) >= 0 {
		t.Error("compareAny(false, true) should be < 0")
	}
}

func TestCompareAny_Bool_TrueFalse(t *testing.T) {
	// av is true, !av is false → return 1
	if compareAny(true, false) <= 0 {
		t.Error("compareAny(true, false) should be > 0")
	}
}

func TestCompareAny_Default(t *testing.T) {
	type point struct{ X int }
	// default case → fmt.Sprintf comparison
	r := compareAny(point{1}, point{2})
	_ = r // just check no panic
}
