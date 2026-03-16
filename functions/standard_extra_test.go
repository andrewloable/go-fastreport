package functions_test

import (
	"testing"
	"time"

	"github.com/andrewloable/go-fastreport/functions"
)

// ── AddHours ──────────────────────────────────────────────────────────────────

func TestAddHours(t *testing.T) {
	base := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	cases := []struct {
		hours float64
		wantH int
	}{
		{1, 1},
		{6, 6},
		{-2, 22}, // wraps to previous day's hour
		{0.5, 0}, // 30 minutes → still hour 0
	}
	for _, c := range cases {
		got := functions.AddHours(base, c.hours)
		if got.Hour() != c.wantH {
			t.Errorf("AddHours(%v).Hour() = %d, want %d", c.hours, got.Hour(), c.wantH)
		}
	}
}

// ── AddMinutes ────────────────────────────────────────────────────────────────

func TestAddMinutes(t *testing.T) {
	base := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	cases := []struct {
		mins  float64
		wantM int
	}{
		{30, 30},
		{60, 0},  // wraps to next hour
		{-1, 59}, // wraps to previous hour
		{90, 30}, // 1h30m
	}
	for _, c := range cases {
		got := functions.AddMinutes(base, c.mins)
		if got.Minute() != c.wantM {
			t.Errorf("AddMinutes(%v).Minute() = %d, want %d", c.mins, got.Minute(), c.wantM)
		}
	}
}

// ── AddSeconds ────────────────────────────────────────────────────────────────

func TestAddSeconds(t *testing.T) {
	base := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	cases := []struct {
		secs  float64
		wantS int
	}{
		{30, 30},
		{60, 0},  // wraps to next minute
		{-1, 59}, // wraps to previous minute
		{90, 30},
	}
	for _, c := range cases {
		got := functions.AddSeconds(base, c.secs)
		if got.Second() != c.wantS {
			t.Errorf("AddSeconds(%v).Second() = %d, want %d", c.secs, got.Second(), c.wantS)
		}
	}
}

// ── DayOfYear ─────────────────────────────────────────────────────────────────

func TestDayOfYear(t *testing.T) {
	cases := []struct {
		date time.Time
		want int
	}{
		{time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC), 1},
		{time.Date(2024, 12, 31, 0, 0, 0, 0, time.UTC), 366}, // 2024 is leap year
		{time.Date(2023, 12, 31, 0, 0, 0, 0, time.UTC), 365}, // non-leap year
		{time.Date(2024, 3, 1, 0, 0, 0, 0, time.UTC), 61},    // after Feb 29 in leap year
		{time.Date(2024, 7, 4, 0, 0, 0, 0, time.UTC), 186},
	}
	for _, c := range cases {
		got := functions.DayOfYear(c.date)
		if got != c.want {
			t.Errorf("DayOfYear(%v) = %d, want %d", c.date, got, c.want)
		}
	}
}

// ── WeekOfYear ────────────────────────────────────────────────────────────────

func TestWeekOfYear(t *testing.T) {
	cases := []struct {
		date time.Time
		want int
	}{
		{time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC), 1},
		{time.Date(2024, 1, 7, 0, 0, 0, 0, time.UTC), 1},
		{time.Date(2024, 1, 8, 0, 0, 0, 0, time.UTC), 2},
		{time.Date(2024, 12, 30, 0, 0, 0, 0, time.UTC), 1}, // ISO: last days may be week 1 of next year
		{time.Date(2024, 7, 4, 0, 0, 0, 0, time.UTC), 27},
		{time.Date(2023, 12, 28, 0, 0, 0, 0, time.UTC), 52},
	}
	for _, c := range cases {
		_, isoWeek := c.date.ISOWeek()
		got := functions.WeekOfYear(c.date)
		if got != isoWeek {
			t.Errorf("WeekOfYear(%v) = %d, want %d (ISO)", c.date, got, isoWeek)
		}
	}
}

