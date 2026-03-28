package table

import "github.com/andrewloable/go-fastreport/report"

// DeserializeChild handles table-level child elements during FRX deserialization.
// It intercepts TableColumn and TableRow children and appends them to the
// internal slices, preventing them from being added to an Objects() collection
// (which TableBase does not have).
func (t *TableBase) DeserializeChild(childType string, r report.Reader) bool {
	switch childType {
	case "TableColumn":
		col := NewTableColumn()
		_ = col.Deserialize(r)
		// TableColumn has no sub-children in FRX.
		t.columns = append(t.columns, col)
		return true
	case "TableRow":
		row := NewTableRow()
		_ = row.Deserialize(r)
		// Process TableCell children of this row.
		for {
			ct, ok := r.NextChild()
			if !ok {
				break
			}
			if ct == "TableCell" {
				cell := NewTableCell()
				_ = cell.Deserialize(r)
				// Deserialize sub-children of the cell (e.g. PictureObject).
				for {
					childType, moreKids := r.NextChild()
					if !moreKids {
						break
					}
					cell.DeserializeChild(childType, r)
					if r.FinishChild() != nil { break }
				}
				row.AddCell(cell)
			}
			if r.FinishChild() != nil { break }
		}
		t.rows = append(t.rows, row)
		return true
	}
	return false
}
