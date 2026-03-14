package utils

import (
	"testing"
)

func TestCRC32KnownValues(t *testing.T) {
	tests := []struct {
		input    []byte
		expected uint32
	}{
		// IEEE CRC32 of empty input is 0.
		{[]byte{}, 0x00000000},
		// Well-known test vector: "123456789" → 0xCBF43926
		{[]byte("123456789"), 0xCBF43926},
		// Single byte 0x00.
		{[]byte{0x00}, 0xD202EF8D},
	}

	for _, tc := range tests {
		got := CRC32(tc.input)
		if got != tc.expected {
			t.Errorf("CRC32(%q) = 0x%08X, want 0x%08X", tc.input, got, tc.expected)
		}
	}
}

func TestCRC32StringFormat(t *testing.T) {
	// "123456789" → "cbf43926" (lowercase, 8 hex digits)
	got := CRC32String([]byte("123456789"))
	want := "cbf43926"
	if got != want {
		t.Errorf("CRC32String(%q) = %q, want %q", "123456789", got, want)
	}
}

func TestCRC32StringLeadingZeroPad(t *testing.T) {
	// Verify zero-padding: empty input produces "00000000".
	got := CRC32String([]byte{})
	want := "00000000"
	if got != want {
		t.Errorf("CRC32String(empty) = %q, want %q", got, want)
	}
}

func TestCRC32Consistency(t *testing.T) {
	data := []byte("FastReport go-fastreport CRC32 test")
	a := CRC32(data)
	b := CRC32(data)
	if a != b {
		t.Fatal("CRC32 is not consistent across calls")
	}
}

func TestCRC32StringConsistency(t *testing.T) {
	data := []byte("consistency check")
	a := CRC32String(data)
	b := CRC32String(data)
	if a != b {
		t.Fatal("CRC32String is not consistent across calls")
	}
}

func TestCRC32DifferentInputs(t *testing.T) {
	a := CRC32([]byte("hello"))
	b := CRC32([]byte("world"))
	if a == b {
		t.Fatal("CRC32 of different inputs should differ")
	}
}
