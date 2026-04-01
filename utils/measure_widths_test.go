package utils

import (
	"fmt"
	"math"
	"testing"

	"github.com/andrewloable/go-fastreport/style"
)

func TestMeasureColumnWidths(t *testing.T) {
	tahoma8 := style.Font{Name: "Tahoma", Size: 8}
	tahomaB8 := style.Font{Name: "Tahoma", Size: 8, Style: style.FontStyleBold}

	// Raw measurements with full precision
	type entry struct {
		text string
		font style.Font
	}
	entries := []entry{
		{"Andrew Fuller", tahoma8},
		{"Steven Buchanan", tahoma8},
		{"Total", tahomaB8},
		{"₱39,999.00", tahomaB8},
		{"₱12,099.00", tahoma8},
		{"₱7,800.00", tahoma8},
	}
	
	fmt.Printf("\n%-22s %-5s %-15s %-12s %-10s\n", "Text", "B?", "rawMeas", "textW+4+2", "ceil")
	for _, e := range entries {
		w, _ := MeasureText(e.text, e.font, 0)
		textW := ScaleWidth(w, e.font)
		sum := textW + 4 + 2
		c := float32(math.Ceil(float64(sum)))
		bold := " "
		if e.font.Style == style.FontStyleBold {
			bold = "B"
		}
		fmt.Printf("%-22s %-5s %-15.10f %-12.10f %-10.0f\n", e.text, bold, textW, sum, c)
	}
}
