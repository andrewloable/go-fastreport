package report

import (
	"fmt"
	"strconv"
	"strings"
)

// FormatFloat serialises a float32 trimming unnecessary trailing zeros.
func FormatFloat(v float32) string {
	return fmt.Sprintf("%g", v)
}

// ParseFloat parses a float32 from string; returns 0 on error.
func ParseFloat(s string) float32 {
	s = strings.TrimSpace(s)
	v, err := strconv.ParseFloat(s, 32)
	if err != nil {
		return 0
	}
	return float32(v)
}

// SplitComma splits s by comma and trims each element.
func SplitComma(s string) []string {
	parts := strings.Split(s, ",")
	for i, p := range parts {
		parts[i] = strings.TrimSpace(p)
	}
	return parts
}
