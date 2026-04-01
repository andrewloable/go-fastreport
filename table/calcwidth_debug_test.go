package table

import (
	"fmt"
	"testing"

	"github.com/andrewloable/go-fastreport/object"
	"github.com/andrewloable/go-fastreport/style"
)

func TestCalcWidthDebug(t *testing.T) {
	// Simulate a simple column with "Andrew Fuller" cell
	tahoma8 := style.Font{Name: "Tahoma", Size: 8}
	
	cell := NewTableCell()
	cell.SetText("Andrew Fuller")
	cell.SetFont(tahoma8)
	
	fmt.Printf("Cell padding: %+v\n", cell.Padding())
	fmt.Printf("Cell font: %+v\n", cell.Font())
	fmt.Printf("cellMeasuredWidth: %.4f\n", cellMeasuredWidth(cell))
	
	// Also test with padding (2,0,2,0) like TextObjectBase default
	cell2 := NewTableCell()
	cell2.SetText("Andrew Fuller")
	cell2.SetFont(tahoma8)
	cell2.SetPadding(object.Padding{Left: 2, Top: 0, Right: 2, Bottom: 0})
	fmt.Printf("\nWith padding (2,0,2,0):\n")
	fmt.Printf("Cell padding: %+v\n", cell2.Padding())
	fmt.Printf("cellMeasuredWidth: %.4f\n", cellMeasuredWidth(cell2))
}
