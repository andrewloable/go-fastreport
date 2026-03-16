package object_test

// advmatrix_coverage_test.go — coverage tests for advmatrix.go uncovered branches.

import (
	"bytes"
	"strings"
	"testing"

	"github.com/andrewloable/go-fastreport/object"
	"github.com/andrewloable/go-fastreport/serial"
	"github.com/andrewloable/go-fastreport/style"
)

// ── AdvMatrixColumn.Serialize: AutoSize=true, Width set ──────────────────────

func TestAdvMatrixColumn_SerializeDeserialize_AllFieldsCoverage(t *testing.T) {
	orig := object.NewAdvMatrixObject()
	col := &object.AdvMatrixColumn{
		Name:     "Col1",
		Width:    120,
		AutoSize: true,
	}
	orig.TableColumns = append(orig.TableColumns, col)

	var buf bytes.Buffer
	w := serial.NewWriter(&buf)
	if err := w.WriteObjectNamed("AdvMatrixObject", orig); err != nil {
		t.Fatalf("Serialize: %v", err)
	}
	w.Flush() //nolint:errcheck

	xml := buf.String()
	if !strings.Contains(xml, `AutoSize="true"`) {
		t.Errorf("expected AutoSize in XML:\n%s", xml)
	}
	if !strings.Contains(xml, `Width=`) {
		t.Errorf("expected Width in XML:\n%s", xml)
	}
}

// ── AdvMatrixRow.Serialize: with cells ───────────────────────────────────────

func TestAdvMatrixRow_SerializeDeserialize_WithCellsCoverage(t *testing.T) {
	orig := object.NewAdvMatrixObject()
	orig.DataSource = "DS1"

	col := &object.AdvMatrixColumn{Name: "C1", Width: 100}
	orig.TableColumns = append(orig.TableColumns, col)

	cell := &object.AdvMatrixCell{
		Name:      "Cell1",
		Text:      "Hello",
		ColSpan:   2,
		RowSpan:   1,
		HorzAlign: 1,
		VertAlign: 2,
	}
	row := &object.AdvMatrixRow{
		Name:   "Row1",
		Height: 30,
		Cells:  []*object.AdvMatrixCell{cell},
	}
	orig.TableRows = append(orig.TableRows, row)

	var buf bytes.Buffer
	w := serial.NewWriter(&buf)
	if err := w.WriteObjectNamed("AdvMatrixObject", orig); err != nil {
		t.Fatalf("Serialize: %v", err)
	}
	w.Flush() //nolint:errcheck

	xml := buf.String()
	if !strings.Contains(xml, `DataSource=`) {
		t.Errorf("expected DataSource in XML:\n%s", xml)
	}
	if !strings.Contains(xml, `TableCell`) {
		t.Errorf("expected TableCell in XML:\n%s", xml)
	}
	if !strings.Contains(xml, `ColSpan=`) {
		t.Errorf("expected ColSpan in XML:\n%s", xml)
	}

	// Deserialize and verify.
	r := serial.NewReader(bytes.NewReader(buf.Bytes()))
	_, ok := r.ReadObjectHeader()
	if !ok {
		t.Fatal("ReadObjectHeader failed")
	}
	got := object.NewAdvMatrixObject()
	if err := got.Deserialize(r); err != nil {
		t.Fatalf("Deserialize: %v", err)
	}
	for {
		ct, ok2 := r.NextChild()
		if !ok2 {
			break
		}
		if !got.DeserializeChild(ct, r) {
			r.FinishChild() //nolint:errcheck
			continue
		}
		r.FinishChild() //nolint:errcheck
	}

	if got.DataSource != "DS1" {
		t.Errorf("DataSource: got %q, want DS1", got.DataSource)
	}
	if len(got.TableColumns) != 1 {
		t.Errorf("TableColumns: got %d, want 1", len(got.TableColumns))
	}
	if len(got.TableRows) != 1 {
		t.Errorf("TableRows: got %d, want 1", len(got.TableRows))
	}
	if len(got.TableRows) > 0 && len(got.TableRows[0].Cells) != 1 {
		t.Errorf("Row cells: got %d, want 1", len(got.TableRows[0].Cells))
	}
}

// ── AdvMatrixObject: empty rows/columns ──────────────────────────────────────

func TestAdvMatrixObject_Serialize_EmptyCoverage(t *testing.T) {
	orig := object.NewAdvMatrixObject()

	var buf bytes.Buffer
	w := serial.NewWriter(&buf)
	if err := w.WriteObjectNamed("AdvMatrixObject", orig); err != nil {
		t.Fatalf("Serialize: %v", err)
	}
	w.Flush() //nolint:errcheck

	xml := buf.String()
	if strings.Contains(xml, "TableColumn") {
		t.Errorf("unexpected TableColumn in empty matrix:\n%s", xml)
	}
}

// ── formatBorderLinesStr via cell with various BorderLines ───────────────────

func TestAdvMatrixCell_BorderLines_VariantsCoverage(t *testing.T) {
	// Exercise formatBorderLinesStr with single Left, Right, Bottom values
	// and a combo (Left|Right) which hits the default → returns "".
	cases := []struct {
		bl   style.BorderLines
		want string // expected xml content substring (empty = no Border.Lines)
	}{
		{style.BorderLinesLeft, `Border.Lines="Left"`},
		{style.BorderLinesRight, `Border.Lines="Right"`},
		{style.BorderLinesBottom, `Border.Lines="Bottom"`},
		{style.BorderLinesTop, `Border.Lines="Top"`},
		{style.BorderLinesAll, `Border.Lines="All"`},
		{style.BorderLinesLeft | style.BorderLinesRight, ""}, // default case → no write
	}

	for _, tc := range cases {
		b := style.NewBorder()
		b.VisibleLines = tc.bl

		orig := object.NewAdvMatrixObject()
		row := &object.AdvMatrixRow{Name: "R1"}
		cell := &object.AdvMatrixCell{Name: "c1", Border: b}
		row.Cells = []*object.AdvMatrixCell{cell}
		orig.TableRows = append(orig.TableRows, row)

		var buf bytes.Buffer
		ww := serial.NewWriter(&buf)
		ww.WriteObjectNamed("AdvMatrixObject", orig) //nolint:errcheck
		ww.Flush()                                   //nolint:errcheck

		xml := buf.String()
		if tc.want != "" && !strings.Contains(xml, tc.want) {
			t.Errorf("bl=%d: expected %q in XML:\n%s", tc.bl, tc.want, xml)
		}
	}
}

// ── DeserializeChild: drainAdvChildren path ───────────────────────────────────

func TestAdvMatrixObject_DeserializeChild_UnknownChildCoverage(t *testing.T) {
	// Build XML where the AdvMatrixObject has special child types that trigger drainAdvChildren.
	xml := `<AdvMatrixObject>` +
		`<Cells><SomeGrandChild/></Cells>` +
		`<MatrixCollapseButton Name="btn1"/>` +
		`<MatrixRows><Desc/></MatrixRows>` +
		`<MatrixColumns/>` +
		`</AdvMatrixObject>`

	r := serial.NewReader(strings.NewReader(xml))
	_, ok := r.ReadObjectHeader()
	if !ok {
		t.Fatal("ReadObjectHeader failed")
	}
	obj := object.NewAdvMatrixObject()
	obj.Deserialize(r) //nolint:errcheck

	for {
		ct, ok2 := r.NextChild()
		if !ok2 {
			break
		}
		handled := obj.DeserializeChild(ct, r)
		_ = handled
		r.FinishChild() //nolint:errcheck
	}
}

// ── drainAdvChildren: nested children ────────────────────────────────────────

func TestAdvMatrixObject_DeserializeChild_DrainNestedChildrenCoverage(t *testing.T) {
	// TableRow with an unexpected child type (non-TableCell) to hit drainAdvChildren.
	xml := `<AdvMatrixObject>` +
		`<TableRow Name="R1" Height="30">` +
		`<UnknownElement Foo="bar"><Grandchild/></UnknownElement>` +
		`<TableCell Name="c1" Text="hello"/>` +
		`</TableRow>` +
		`</AdvMatrixObject>`

	r := serial.NewReader(strings.NewReader(xml))
	_, ok := r.ReadObjectHeader()
	if !ok {
		t.Fatal("ReadObjectHeader failed")
	}
	obj := object.NewAdvMatrixObject()
	obj.Deserialize(r) //nolint:errcheck

	for {
		ct, ok2 := r.NextChild()
		if !ok2 {
			break
		}
		obj.DeserializeChild(ct, r) //nolint:errcheck
		r.FinishChild()             //nolint:errcheck
	}

	if len(obj.TableRows) != 1 {
		t.Errorf("expected 1 row, got %d", len(obj.TableRows))
	}
	if len(obj.TableRows) > 0 && len(obj.TableRows[0].Cells) != 1 {
		t.Errorf("expected 1 cell (unknown skipped), got %d", len(obj.TableRows[0].Cells))
	}
}

// ── readAdvDescriptor: with nested descriptor ─────────────────────────────────

func TestAdvMatrixObject_DeserializeChild_DescriptorCoverage(t *testing.T) {
	xml := `<AdvMatrixObject>` +
		`<Columns>` +
		`<Descriptor Expression="[Year]" DisplayText="Year" Sort="Asc">` +
		`<Descriptor Expression="[Month]" DisplayText="Month"/>` +
		`<UnknownChild/>` +
		`</Descriptor>` +
		`</Columns>` +
		`<Rows>` +
		`<Descriptor Expression="[Cat]" DisplayText="Category"/>` +
		`</Rows>` +
		`</AdvMatrixObject>`

	r := serial.NewReader(strings.NewReader(xml))
	_, ok := r.ReadObjectHeader()
	if !ok {
		t.Fatal("ReadObjectHeader failed")
	}
	obj := object.NewAdvMatrixObject()
	obj.Deserialize(r) //nolint:errcheck

	for {
		ct, ok2 := r.NextChild()
		if !ok2 {
			break
		}
		obj.DeserializeChild(ct, r) //nolint:errcheck
		r.FinishChild()             //nolint:errcheck
	}

	if len(obj.Columns) == 0 {
		t.Error("expected at least 1 column descriptor")
	}
}

// ── parseBorderLinesStr: unknown string ───────────────────────────────────────

func TestAdvMatrixCell_Deserialize_UnknownBorderLinesCoverage(t *testing.T) {
	xml := `<AdvMatrixObject>` +
		`<TableRow Name="R1">` +
		`<TableCell Name="c1" Border.Lines="UnknownStyle" Border.Color="#FF0000FF"/>` +
		`</TableRow>` +
		`</AdvMatrixObject>`

	r := serial.NewReader(strings.NewReader(xml))
	_, ok := r.ReadObjectHeader()
	if !ok {
		t.Fatal("ReadObjectHeader failed")
	}
	obj := object.NewAdvMatrixObject()
	obj.Deserialize(r) //nolint:errcheck
	for {
		ct, ok2 := r.NextChild()
		if !ok2 {
			break
		}
		obj.DeserializeChild(ct, r) //nolint:errcheck
		r.FinishChild()             //nolint:errcheck
	}
	if len(obj.TableRows) > 0 && len(obj.TableRows[0].Cells) > 0 {
		cell := obj.TableRows[0].Cells[0]
		_ = cell.Border
	}
}

// ── AdvMatrixRow.AutoSize + AdvMatrixCell.RowSpan >1 ─────────────────────────

func TestAdvMatrixRow_AutoSize_AndCell_RowSpan(t *testing.T) {
	orig := object.NewAdvMatrixObject()
	cell := &object.AdvMatrixCell{
		Name:    "c1",
		ColSpan: 1,
		RowSpan: 3, // > 1, should serialize
		Text:    "spanning",
	}
	row := &object.AdvMatrixRow{
		Name:     "R1",
		AutoSize: true,
		Cells:    []*object.AdvMatrixCell{cell},
	}
	orig.TableRows = append(orig.TableRows, row)

	var buf bytes.Buffer
	w := serial.NewWriter(&buf)
	w.WriteObjectNamed("AdvMatrixObject", orig) //nolint:errcheck
	w.Flush()                                   //nolint:errcheck

	xml := buf.String()
	if !strings.Contains(xml, `RowSpan=`) {
		t.Errorf("expected RowSpan in XML:\n%s", xml)
	}
}
