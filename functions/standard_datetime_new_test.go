package functions_test

import (
	"testing"
	"time"

	"github.com/andrewloable/go-fastreport/functions"
)

// ── IsLeapYear ────────────────────────────────────────────────────────────────

func TestIsLeapYear_LeapYears(t *testing.T) {
	leapYears := []int{1600, 2000, 2004, 2024, 2400}
	for _, y := range leapYears {
		if !functions.IsLeapYear(y) {
			t.Errorf("IsLeapYear(%d) = false, want true", y)
		}
	}
}

func TestIsLeapYear_NonLeapYears(t *testing.T) {
	nonLeapYears := []int{1700, 1800, 1900, 2001, 2100, 2023, 2025}
	for _, y := range nonLeapYears {
		if functions.IsLeapYear(y) {
			t.Errorf("IsLeapYear(%d) = true, want false", y)
		}
	}
}

func TestIsLeapYear_CenturyDivisibleBy400(t *testing.T) {
	// 2000 is a leap year (divisible by 400)
	if !functions.IsLeapYear(2000) {
		t.Errorf("IsLeapYear(2000) = false, want true")
	}
	// 1900 is not a leap year (divisible by 100 but not 400)
	if functions.IsLeapYear(1900) {
		t.Errorf("IsLeapYear(1900) = true, want false")
	}
}

func TestIsLeapYear_RegisteredInAll(t *testing.T) {
	all := functions.All()
	if _, ok := all["IsLeapYear"]; !ok {
		t.Error("IsLeapYear not registered in All()")
	}
}

// ── TimeSerial ────────────────────────────────────────────────────────────────

func TestTimeSerial_Basic(t *testing.T) {
	cases := []struct {
		h, m, s    int
		wantH      int
		wantM      int
		wantS      int
	}{
		{0, 0, 0, 0, 0, 0},
		{12, 30, 45, 12, 30, 45},
		{23, 59, 59, 23, 59, 59},
		{9, 5, 1, 9, 5, 1},
	}
	for _, c := range cases {
		got := functions.TimeSerial(c.h, c.m, c.s)
		if got.Hour() != c.wantH || got.Minute() != c.wantM || got.Second() != c.wantS {
			t.Errorf("TimeSerial(%d,%d,%d) = %v, want %02d:%02d:%02d",
				c.h, c.m, c.s, got, c.wantH, c.wantM, c.wantS)
		}
	}
}

func TestTimeSerial_ZeroDate(t *testing.T) {
	// Date portion should be the zero date: year 1, January 1
	got := functions.TimeSerial(10, 20, 30)
	want := time.Date(1, time.January, 1, 10, 20, 30, 0, time.UTC)
	if !got.Equal(want) {
		t.Errorf("TimeSerial(10,20,30) = %v, want %v", got, want)
	}
}

func TestTimeSerial_UTC(t *testing.T) {
	got := functions.TimeSerial(8, 0, 0)
	if got.Location() != time.UTC {
		t.Errorf("TimeSerial location = %v, want UTC", got.Location())
	}
}

func TestTimeSerial_RegisteredInAll(t *testing.T) {
	all := functions.All()
	if _, ok := all["TimeSerial"]; !ok {
		t.Error("TimeSerial not registered in All()")
	}
}
