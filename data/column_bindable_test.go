package data_test

// column_bindable_test.go — tests for ColumnBindableControl enum and
// related DataColumn methods (SetBindableControlType, Serialize/Deserialize
// round-trips for BindableControl/CustomBindableControl).
// C# ref: FastReport.Base/Data/Column.cs

import (
	"bytes"
	"strings"
	"testing"

	"github.com/andrewloable/go-fastreport/data"
	"github.com/andrewloable/go-fastreport/serial"
)

// ── ColumnBindableControl constants ──────────────────────────────────────────

func TestColumnBindableControl_Constants_Unique(t *testing.T) {
	controls := []data.ColumnBindableControl{
		data.ColumnBindableControlText,
		data.ColumnBindableControlRichText,
		data.ColumnBindableControlPicture,
		data.ColumnBindableControlCheckBox,
		data.ColumnBindableControlCustom,
	}
	seen := make(map[data.ColumnBindableControl]bool)
	for _, c := range controls {
		if seen[c] {
			t.Errorf("duplicate ColumnBindableControl value %d", c)
		}
		seen[c] = true
	}
}

func TestNewDataColumn_DefaultBindableControl(t *testing.T) {
	col := data.NewDataColumn("X")
	if col.BindableControl != data.ColumnBindableControlText {
		t.Errorf("default BindableControl = %d, want ColumnBindableControlText (%d)",
			col.BindableControl, data.ColumnBindableControlText)
	}
	if col.CustomBindableControl != "" {
		t.Errorf("default CustomBindableControl = %q, want empty", col.CustomBindableControl)
	}
}

// ── SetBindableControlType ───────────────────────────────────────────────────

func TestSetBindableControlType_ByteSlice(t *testing.T) {
	col := data.NewDataColumn("Photo")
	col.SetBindableControlType("[]byte")
	if col.BindableControl != data.ColumnBindableControlPicture {
		t.Errorf("[]byte → BindableControl = %d, want Picture", col.BindableControl)
	}
}

func TestSetBindableControlType_Image(t *testing.T) {
	col := data.NewDataColumn("Img")
	col.SetBindableControlType("image.Image")
	if col.BindableControl != data.ColumnBindableControlPicture {
		t.Errorf("image.Image → BindableControl = %d, want Picture", col.BindableControl)
	}
}

func TestSetBindableControlType_Bool(t *testing.T) {
	col := data.NewDataColumn("Active")
	col.SetBindableControlType("bool")
	if col.BindableControl != data.ColumnBindableControlCheckBox {
		t.Errorf("bool → BindableControl = %d, want CheckBox", col.BindableControl)
	}
}

func TestSetBindableControlType_String(t *testing.T) {
	col := data.NewDataColumn("Name")
	col.SetBindableControlType("string")
	if col.BindableControl != data.ColumnBindableControlText {
		t.Errorf("string → BindableControl = %d, want Text", col.BindableControl)
	}
}

func TestSetBindableControlType_Int(t *testing.T) {
	col := data.NewDataColumn("Count")
	col.SetBindableControlType("int")
	if col.BindableControl != data.ColumnBindableControlText {
		t.Errorf("int → BindableControl = %d, want Text", col.BindableControl)
	}
}

func TestSetBindableControlType_Unknown(t *testing.T) {
	col := data.NewDataColumn("X")
	col.SetBindableControlType("some.CustomType")
	if col.BindableControl != data.ColumnBindableControlText {
		t.Errorf("unknown type → BindableControl = %d, want Text", col.BindableControl)
	}
}

// ── Serialize BindableControl ────────────────────────────────────────────────

func TestDataColumn_Serialize_BindableControlSkippedWhenText(t *testing.T) {
	col := data.NewDataColumn("X")
	col.BindableControl = data.ColumnBindableControlText // default
	var buf bytes.Buffer
	w := serial.NewWriter(&buf)
	_ = w.BeginObject("Column")
	_ = col.Serialize(w)
	_ = w.EndObject()
	_ = w.Flush()
	if strings.Contains(buf.String(), "BindableControl") {
		t.Errorf("BindableControl should not be serialized for default Text; xml=%s", buf.String())
	}
}

func TestDataColumn_Serialize_BindableControlPicture(t *testing.T) {
	col := data.NewDataColumn("Photo")
	col.BindableControl = data.ColumnBindableControlPicture
	var buf bytes.Buffer
	w := serial.NewWriter(&buf)
	_ = w.BeginObject("Column")
	_ = col.Serialize(w)
	_ = w.EndObject()
	_ = w.Flush()
	if !strings.Contains(buf.String(), "BindableControl") {
		t.Errorf("BindableControl should be serialized for Picture; xml=%s", buf.String())
	}
	if !strings.Contains(buf.String(), "Picture") {
		t.Errorf("BindableControl value 'Picture' not found in xml=%s", buf.String())
	}
}

func TestDataColumn_Serialize_CustomBindableControl(t *testing.T) {
	col := data.NewDataColumn("X")
	col.BindableControl = data.ColumnBindableControlCustom
	col.CustomBindableControl = "MySpecialControl"
	var buf bytes.Buffer
	w := serial.NewWriter(&buf)
	_ = w.BeginObject("Column")
	_ = col.Serialize(w)
	_ = w.EndObject()
	_ = w.Flush()
	xmlStr := buf.String()
	if !strings.Contains(xmlStr, "CustomBindableControl") {
		t.Errorf("CustomBindableControl not serialized; xml=%s", xmlStr)
	}
	if !strings.Contains(xmlStr, "MySpecialControl") {
		t.Errorf("CustomBindableControl value not found; xml=%s", xmlStr)
	}
}

// ── Deserialize BindableControl ──────────────────────────────────────────────

func TestDataColumn_Deserialize_BindableControlPicture(t *testing.T) {
	xmlData := `<Column Name="Photo" BindableControl="Picture"/>`
	r := serial.NewReader(strings.NewReader(xmlData))
	_, _ = r.ReadObjectHeader()
	col := data.NewDataColumn(r.ReadStr("Name", ""))
	_ = col.Deserialize(r)
	if col.BindableControl != data.ColumnBindableControlPicture {
		t.Errorf("BindableControl = %d, want Picture", col.BindableControl)
	}
}

func TestDataColumn_Deserialize_BindableControlCheckBox(t *testing.T) {
	xmlData := `<Column Name="Active" BindableControl="CheckBox"/>`
	r := serial.NewReader(strings.NewReader(xmlData))
	_, _ = r.ReadObjectHeader()
	col := data.NewDataColumn(r.ReadStr("Name", ""))
	_ = col.Deserialize(r)
	if col.BindableControl != data.ColumnBindableControlCheckBox {
		t.Errorf("BindableControl = %d, want CheckBox", col.BindableControl)
	}
}

func TestDataColumn_Deserialize_BindableControlRichText(t *testing.T) {
	xmlData := `<Column Name="Notes" BindableControl="RichText"/>`
	r := serial.NewReader(strings.NewReader(xmlData))
	_, _ = r.ReadObjectHeader()
	col := data.NewDataColumn(r.ReadStr("Name", ""))
	_ = col.Deserialize(r)
	if col.BindableControl != data.ColumnBindableControlRichText {
		t.Errorf("BindableControl = %d, want RichText", col.BindableControl)
	}
}

func TestDataColumn_Deserialize_CustomBindableControl(t *testing.T) {
	xmlData := `<Column Name="X" CustomBindableControl="MyControl"/>`
	r := serial.NewReader(strings.NewReader(xmlData))
	_, _ = r.ReadObjectHeader()
	col := data.NewDataColumn(r.ReadStr("Name", ""))
	_ = col.Deserialize(r)
	if col.CustomBindableControl != "MyControl" {
		t.Errorf("CustomBindableControl = %q, want MyControl", col.CustomBindableControl)
	}
	// When CustomBindableControl is set, BindableControl should be Custom.
	if col.BindableControl != data.ColumnBindableControlCustom {
		t.Errorf("BindableControl = %d, want Custom when CustomBindableControl set", col.BindableControl)
	}
}

func TestDataColumn_Deserialize_UnknownBindableControl_DefaultsToText(t *testing.T) {
	xmlData := `<Column Name="X" BindableControl="UnknownType"/>`
	r := serial.NewReader(strings.NewReader(xmlData))
	_, _ = r.ReadObjectHeader()
	col := data.NewDataColumn(r.ReadStr("Name", ""))
	_ = col.Deserialize(r)
	if col.BindableControl != data.ColumnBindableControlText {
		t.Errorf("unknown BindableControl should default to Text, got %d", col.BindableControl)
	}
}

func TestDataColumn_Deserialize_BindableControlAbsent_DefaultsToText(t *testing.T) {
	// No BindableControl attribute → should remain Text.
	xmlData := `<Column Name="X"/>`
	r := serial.NewReader(strings.NewReader(xmlData))
	_, _ = r.ReadObjectHeader()
	col := data.NewDataColumn(r.ReadStr("Name", ""))
	_ = col.Deserialize(r)
	if col.BindableControl != data.ColumnBindableControlText {
		t.Errorf("absent BindableControl should default to Text, got %d", col.BindableControl)
	}
}

// ── Round-trip ──────────────────────────────────────────────────────────────

func TestDataColumn_RoundTrip_BindableControlPicture(t *testing.T) {
	col := data.NewDataColumn("Photo")
	col.BindableControl = data.ColumnBindableControlPicture

	var buf bytes.Buffer
	w := serial.NewWriter(&buf)
	_ = w.BeginObject("Column")
	w.WriteStr("Name", col.Name)
	_ = col.Serialize(w)
	_ = w.EndObject()
	_ = w.Flush()

	r := serial.NewReader(bytes.NewReader(buf.Bytes()))
	_, _ = r.ReadObjectHeader()
	got := data.NewDataColumn(r.ReadStr("Name", ""))
	_ = got.Deserialize(r)

	if got.BindableControl != data.ColumnBindableControlPicture {
		t.Errorf("round-trip BindableControl = %d, want Picture", got.BindableControl)
	}
}

func TestDataColumn_RoundTrip_CustomBindableControl(t *testing.T) {
	col := data.NewDataColumn("Special")
	col.BindableControl = data.ColumnBindableControlCustom
	col.CustomBindableControl = "SpecialWidget"

	var buf bytes.Buffer
	w := serial.NewWriter(&buf)
	_ = w.BeginObject("Column")
	w.WriteStr("Name", col.Name)
	_ = col.Serialize(w)
	_ = w.EndObject()
	_ = w.Flush()

	r := serial.NewReader(bytes.NewReader(buf.Bytes()))
	_, _ = r.ReadObjectHeader()
	got := data.NewDataColumn(r.ReadStr("Name", ""))
	_ = got.Deserialize(r)

	if got.CustomBindableControl != "SpecialWidget" {
		t.Errorf("round-trip CustomBindableControl = %q, want SpecialWidget", got.CustomBindableControl)
	}
	if got.BindableControl != data.ColumnBindableControlCustom {
		t.Errorf("round-trip BindableControl = %d, want Custom", got.BindableControl)
	}
}
