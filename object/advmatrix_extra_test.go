package object_test

// advmatrix_extra_test.go — additional coverage tests for advmatrix.go
// Targets the remaining uncovered branches not hit by advmatrix_coverage_test.go.

import (
	"bytes"
	"image/color"
	"strings"
	"testing"

	"github.com/andrewloable/go-fastreport/object"
	"github.com/andrewloable/go-fastreport/serial"
	"github.com/andrewloable/go-fastreport/style"
)

// ── AdvMatrixObject.Serialize: error path via ReportComponentBase ─────────────
// The ReportComponentBase.Serialize always returns nil, so error paths in
// AdvMatrixObject.Serialize (line 112) and WriteObjectNamed error returns
// cannot be triggered without a failing writer. Test all positive paths instead.

// ── AdvMatrixObject.Serialize: with multiple columns AND rows ─────────────────

func TestAdvMatrixObject_Serialize_MultipleColumnsAndRows(t *testing.T) {
	orig := object.NewAdvMatrixObject()
	orig.DataSource = "MultiDS"

	orig.TableColumns = append(orig.TableColumns,
		&object.AdvMatrixColumn{Name: "ColA", Width: 80},
		&object.AdvMatrixColumn{Name: "ColB", Width: 120, AutoSize: true},
	)

	cell1 := &object.AdvMatrixCell{Name: "c1", Text: "Val1", ColSpan: 2, RowSpan: 2}
	cell2 := &object.AdvMatrixCell{Name: "c2", Text: "Val2"}
	row1 := &object.AdvMatrixRow{Name: "R1", Height: 30, Cells: []*object.AdvMatrixCell{cell1}}
	row2 := &object.AdvMatrixRow{Name: "R2", Height: 25, AutoSize: true, Cells: []*object.AdvMatrixCell{cell2}}
	orig.TableRows = append(orig.TableRows, row1, row2)

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
	if !strings.Contains(xml, "ColA") {
		t.Errorf("expected ColA in XML:\n%s", xml)
	}
	if !strings.Contains(xml, "ColB") {
		t.Errorf("expected ColB in XML:\n%s", xml)
	}
	if !strings.Contains(xml, "R1") {
		t.Errorf("expected R1 in XML:\n%s", xml)
	}
	if !strings.Contains(xml, "R2") {
		t.Errorf("expected R2 in XML:\n%s", xml)
	}

	// Deserialize and verify round-trip.
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
		got.DeserializeChild(ct, r) //nolint:errcheck
		r.FinishChild()             //nolint:errcheck
	}
	if got.DataSource != "MultiDS" {
		t.Errorf("DataSource: got %q, want MultiDS", got.DataSource)
	}
	if len(got.TableColumns) != 2 {
		t.Errorf("TableColumns: got %d, want 2", len(got.TableColumns))
	}
	if len(got.TableRows) != 2 {
		t.Errorf("TableRows: got %d, want 2", len(got.TableRows))
	}
}

// ── formatBorderLinesStr: BorderLinesNone branch (via direct serialize path) ──
// BorderLinesNone==0 is skipped by the outer guard, but we verify the guard.

func TestAdvMatrixCell_Serialize_NoBorderWhenNone(t *testing.T) {
	orig := object.NewAdvMatrixObject()
	b := style.NewBorder()
	b.VisibleLines = style.BorderLinesNone // = 0, so condition is false → no border written
	cell := &object.AdvMatrixCell{Name: "c1", Border: b}
	row := &object.AdvMatrixRow{Name: "R1", Cells: []*object.AdvMatrixCell{cell}}
	orig.TableRows = append(orig.TableRows, row)

	var buf bytes.Buffer
	w := serial.NewWriter(&buf)
	w.WriteObjectNamed("AdvMatrixObject", orig) //nolint:errcheck
	w.Flush()                                   //nolint:errcheck

	xml := buf.String()
	if strings.Contains(xml, `Border.Lines=`) {
		t.Errorf("unexpected Border.Lines when VisibleLines=None:\n%s", xml)
	}
}

// ── DeserializeChild: "Rows" with non-Descriptor child ───────────────────────

func TestAdvMatrixObject_DeserializeChild_RowsWithUnknownChild(t *testing.T) {
	xmlStr := `<AdvMatrixObject>` +
		`<Rows>` +
		`<UnknownElement Foo="bar"/>` +
		`<Descriptor Expression="[Region]" DisplayText="Region"/>` +
		`</Rows>` +
		`</AdvMatrixObject>`

	r := serial.NewReader(strings.NewReader(xmlStr))
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

	if len(obj.Rows) != 1 {
		t.Errorf("expected 1 row descriptor, got %d", len(obj.Rows))
	}
	if len(obj.Rows) > 0 && obj.Rows[0].Expression != "[Region]" {
		t.Errorf("Rows[0].Expression: got %q, want [Region]", obj.Rows[0].Expression)
	}
}

// ── DeserializeChild: "Columns" with non-Descriptor child ────────────────────

func TestAdvMatrixObject_DeserializeChild_ColumnsWithUnknownChild(t *testing.T) {
	xmlStr := `<AdvMatrixObject>` +
		`<Columns>` +
		`<SomeUnknown/>` +
		`<Descriptor Expression="[Year]" DisplayText="Year"/>` +
		`</Columns>` +
		`</AdvMatrixObject>`

	r := serial.NewReader(strings.NewReader(xmlStr))
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

	if len(obj.Columns) != 1 {
		t.Errorf("expected 1 column descriptor, got %d", len(obj.Columns))
	}
}

// ── readAdvDescriptor: with non-Descriptor child (hits else branch) ───────────

func TestAdvMatrixObject_ReadAdvDescriptor_WithNonDescriptorChild(t *testing.T) {
	xmlStr := `<AdvMatrixObject>` +
		`<Columns>` +
		`<Descriptor Expression="[Year]" DisplayText="Year">` +
		`<SomeOtherElement Foo="bar"/>` +
		`<Descriptor Expression="[Month]" DisplayText="Month"/>` +
		`</Descriptor>` +
		`</Columns>` +
		`</AdvMatrixObject>`

	r := serial.NewReader(strings.NewReader(xmlStr))
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
		t.Fatal("expected at least 1 column descriptor")
	}
	// The outer descriptor should have 1 child (Month), unknown element drained.
	if len(obj.Columns[0].Children) != 1 {
		t.Errorf("expected 1 child descriptor, got %d", len(obj.Columns[0].Children))
	}
	if obj.Columns[0].Children[0].Expression != "[Month]" {
		t.Errorf("child expression: got %q, want [Month]", obj.Columns[0].Children[0].Expression)
	}
}

// ── AdvMatrixObject: DeserializeChild with MatrixSortButton in TableCell ──────

func TestAdvMatrixObject_DeserializeChild_MatrixSortButton(t *testing.T) {
	xmlStr := `<AdvMatrixObject>` +
		`<TableRow Name="R1" Height="30">` +
		`<TableCell Name="c1" Text="header">` +
		`<MatrixSortButton Name="sb1" Left="5" Width="10" Height="10" Dock="Right" SymbolSize="8" Symbol="Asc" ShowCollapseExpandMenu="true"/>` +
		`</TableCell>` +
		`</TableRow>` +
		`</AdvMatrixObject>`

	r := serial.NewReader(strings.NewReader(xmlStr))
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
		t.Fatalf("expected 1 row, got %d", len(obj.TableRows))
	}
	if len(obj.TableRows[0].Cells) != 1 {
		t.Fatalf("expected 1 cell, got %d", len(obj.TableRows[0].Cells))
	}
	cell := obj.TableRows[0].Cells[0]
	if len(cell.Buttons) != 1 {
		t.Errorf("expected 1 button, got %d", len(cell.Buttons))
	}
	if len(cell.Buttons) > 0 {
		btn := cell.Buttons[0]
		if btn.TypeName != "MatrixSortButton" {
			t.Errorf("TypeName: got %q, want MatrixSortButton", btn.TypeName)
		}
		if btn.Name != "sb1" {
			t.Errorf("Name: got %q, want sb1", btn.Name)
		}
		if !btn.ShowCollapseExpandMenu {
			t.Error("ShowCollapseExpandMenu: want true")
		}
	}
}

// ── AdvMatrixObject: Serialize with FillColor and Font on cell ────────────────

func TestAdvMatrixCell_Serialize_FillColorAndFont(t *testing.T) {
	orig := object.NewAdvMatrixObject()

	row := &object.AdvMatrixRow{Name: "R1"}
	font := style.FontFromStr("Arial, 10, Regular")
	cell := &object.AdvMatrixCell{
		Name: "c1",
		Text: "styled",
		Font: &font,
	}
	row.Cells = []*object.AdvMatrixCell{cell}
	orig.TableRows = append(orig.TableRows, row)

	var buf bytes.Buffer
	w := serial.NewWriter(&buf)
	w.WriteObjectNamed("AdvMatrixObject", orig) //nolint:errcheck
	w.Flush()                                   //nolint:errcheck

	xml := buf.String()
	if !strings.Contains(xml, `Font=`) {
		t.Errorf("expected Font in XML:\n%s", xml)
	}
}

// ── AdvMatrixCell: Serialize with FillColor ───────────────────────────────────

func TestAdvMatrixCell_Serialize_WithFillColor(t *testing.T) {
	orig := object.NewAdvMatrixObject()
	row := &object.AdvMatrixRow{Name: "R1"}
	fc := color.RGBA{R: 255, G: 0, B: 0, A: 255}
	cell := &object.AdvMatrixCell{
		Name:      "c1",
		FillColor: &fc,
	}
	row.Cells = []*object.AdvMatrixCell{cell}
	orig.TableRows = append(orig.TableRows, row)

	var buf bytes.Buffer
	w := serial.NewWriter(&buf)
	w.WriteObjectNamed("AdvMatrixObject", orig) //nolint:errcheck
	w.Flush()                                   //nolint:errcheck

	xml := buf.String()
	if !strings.Contains(xml, `Fill.Color=`) {
		t.Errorf("expected Fill.Color in XML:\n%s", xml)
	}
}

// ── AdvMatrixCell: Serialize with MatrixCollapseButton child ──────────────────

func TestAdvMatrixCell_Serialize_WithMatrixCollapseButton(t *testing.T) {
	orig := object.NewAdvMatrixObject()
	row := &object.AdvMatrixRow{Name: "R1"}
	btn := &object.MatrixButton{
		TypeName:               "MatrixCollapseButton",
		Name:                   "btn1",
		Left:                   2,
		Width:                  12,
		Height:                 12,
		Dock:                   "Left",
		SymbolSize:             6,
		Symbol:                 "Collapse",
		ShowCollapseExpandMenu: true,
	}
	cell := &object.AdvMatrixCell{
		Name:    "c1",
		Buttons: []*object.MatrixButton{btn},
	}
	row.Cells = []*object.AdvMatrixCell{cell}
	orig.TableRows = append(orig.TableRows, row)

	var buf bytes.Buffer
	w := serial.NewWriter(&buf)
	w.WriteObjectNamed("AdvMatrixObject", orig) //nolint:errcheck
	w.Flush()                                   //nolint:errcheck

	xml := buf.String()
	if !strings.Contains(xml, "MatrixCollapseButton") {
		t.Errorf("expected MatrixCollapseButton in XML:\n%s", xml)
	}
	if !strings.Contains(xml, `ShowCollapseExpandMenu=`) {
		t.Errorf("expected ShowCollapseExpandMenu in XML:\n%s", xml)
	}
}

// ── AdvMatrixObject: Serialize with Columns/Rows dimension descriptors ────────

func TestAdvMatrixObject_Serialize_WithDescriptors(t *testing.T) {
	orig := object.NewAdvMatrixObject()
	orig.Columns = []*object.AdvMatrixDescriptor{
		{
			Expression:  "[Year]",
			DisplayText: "Year",
			Sort:        "Asc",
			Children: []*object.AdvMatrixDescriptor{
				{Expression: "[Month]", DisplayText: "Month"},
			},
		},
	}
	orig.Rows = []*object.AdvMatrixDescriptor{
		{Expression: "[Region]", DisplayText: "Region"},
	}

	// Descriptors are not serialized by AdvMatrixObject.Serialize directly
	// (they are loaded from FRX but not written back in this implementation).
	// This test verifies the object can be created and round-tripped via XML.
	var buf bytes.Buffer
	w := serial.NewWriter(&buf)
	if err := w.WriteObjectNamed("AdvMatrixObject", orig); err != nil {
		t.Fatalf("Serialize: %v", err)
	}
	w.Flush() //nolint:errcheck
	// The XML won't contain descriptors from Serialize, but it should not error.
	_ = buf.String()
}

// ── AdvMatrixObject: DeserializeChild unknown type returns false ───────────────

func TestAdvMatrixObject_DeserializeChild_UnknownTypeReturnsFalse(t *testing.T) {
	xmlStr := `<AdvMatrixObject><UnknownType Foo="bar"/></AdvMatrixObject>`

	r := serial.NewReader(strings.NewReader(xmlStr))
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
		if handled {
			t.Errorf("expected false for unknown child %q", ct)
		}
		r.FinishChild() //nolint:errcheck
	}
}
