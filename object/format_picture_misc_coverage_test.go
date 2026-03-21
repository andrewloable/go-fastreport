package object_test

// format_picture_misc_coverage_test.go — coverage tests for:
//   - format_serial.go: serializeTextFormat all branches with non-default values
//   - picture.go / rich.go / sparkline.go / rfid.go: full serialize/deserialize
//   - cellular_text.go / digital_signature.go / html.go / svg.go: serialize/deserialize

import (
	"bytes"
	"strings"
	"testing"

	"github.com/andrewloable/go-fastreport/format"
	"github.com/andrewloable/go-fastreport/object"
	"github.com/andrewloable/go-fastreport/serial"
)

// ══════════════════════════════════════════════════════════════════════════════
// format_serial.go — serializeTextFormat: non-default field branches
// ══════════════════════════════════════════════════════════════════════════════

// TestSerializeTextFormat_NumberNonDefaults exercises all non-default branches
// in the NumberFormat case of serializeTextFormat:
// DecimalDigits != 2, UseLocaleSettings=false, custom separators & NegativePattern.
func TestSerializeTextFormat_NumberNonDefaults(t *testing.T) {
	orig := object.NewTextObject()
	nf := &format.NumberFormat{
		UseLocaleSettings: false,
		DecimalDigits:     4,           // non-default (default=2)
		DecimalSeparator:  ",",         // non-default (default=".")
		GroupSeparator:    ".",         // non-default (default=",")
		NegativePattern:   2,           // non-default (default=1)
	}
	orig.SetFormat(nf)

	var buf bytes.Buffer
	w := serial.NewWriter(&buf)
	if err := w.WriteObjectNamed("TextObject", orig); err != nil {
		t.Fatalf("Serialize: %v", err)
	}
	_ = w.Flush()

	xml := buf.String()
	checks := []string{
		`Format="Number"`,
		`Format.DecimalDigits=`,
		`Format.DecimalSeparator=`,
		`Format.GroupSeparator=`,
		`Format.NegativePattern=`,
		`Format.UseLocale="false"`,
	}
	for _, c := range checks {
		if !strings.Contains(xml, c) {
			t.Errorf("expected %q in XML:\n%s", c, xml)
		}
	}

	// Round-trip
	r := serial.NewReader(bytes.NewReader(buf.Bytes()))
	_, ok := r.ReadObjectHeader()
	if !ok {
		t.Fatal("ReadObjectHeader failed")
	}
	got := object.NewTextObject()
	if err := got.Deserialize(r); err != nil {
		t.Fatalf("Deserialize: %v", err)
	}
	if got.Format() == nil {
		t.Fatal("Format should not be nil after round-trip")
	}
	nfGot, ok2 := got.Format().(*format.NumberFormat)
	if !ok2 {
		t.Fatal("Format should be *format.NumberFormat")
	}
	if nfGot.DecimalDigits != 4 {
		t.Errorf("DecimalDigits: got %d, want 4", nfGot.DecimalDigits)
	}
	if nfGot.DecimalSeparator != "," {
		t.Errorf("DecimalSeparator: got %q, want ','", nfGot.DecimalSeparator)
	}
}

// TestSerializeTextFormat_NumberWithLocale exercises NumberFormat with UseLocaleSettings=true
// (default); in this case the non-locale-dependent separator branches are skipped.
func TestSerializeTextFormat_NumberWithLocale(t *testing.T) {
	orig := object.NewTextObject()
	nf := &format.NumberFormat{
		UseLocaleSettings: true,
		DecimalDigits:     0, // non-default
	}
	orig.SetFormat(nf)

	var buf bytes.Buffer
	w := serial.NewWriter(&buf)
	if err := w.WriteObjectNamed("TextObject", orig); err != nil {
		t.Fatalf("Serialize: %v", err)
	}
	_ = w.Flush()

	xml := buf.String()
	if !strings.Contains(xml, `Format="Number"`) {
		t.Errorf("expected Format=Number in XML:\n%s", xml)
	}
	// When UseLocaleSettings=true, UseLocale="false" is NOT written (it's the default)
	if strings.Contains(xml, `Format.UseLocale="false"`) {
		t.Errorf("UseLocale=false should NOT appear when locale is true:\n%s", xml)
	}
}

// TestSerializeTextFormat_CurrencyNonDefaults exercises all non-default branches
// in the CurrencyFormat case: custom symbol, PositivePattern, NegativePattern.
func TestSerializeTextFormat_CurrencyNonDefaults(t *testing.T) {
	orig := object.NewTextObject()
	cf := &format.CurrencyFormat{
		UseLocaleSettings: false,
		DecimalDigits:     0,   // non-default
		DecimalSeparator:  ",", // non-default
		GroupSeparator:    " ", // non-default
		CurrencySymbol:    "€", // non-default
		PositivePattern:   3,   // non-default
		NegativePattern:   5,   // non-default
	}
	orig.SetFormat(cf)

	var buf bytes.Buffer
	w := serial.NewWriter(&buf)
	if err := w.WriteObjectNamed("TextObject", orig); err != nil {
		t.Fatalf("Serialize: %v", err)
	}
	_ = w.Flush()

	xml := buf.String()
	checks := []string{
		`Format="Currency"`,
		`Format.CurrencySymbol=`,
		`Format.PositivePattern=`,
		`Format.NegativePattern=`,
		`Format.UseLocale="false"`,
	}
	for _, c := range checks {
		if !strings.Contains(xml, c) {
			t.Errorf("expected %q in XML:\n%s", c, xml)
		}
	}

	// Round-trip
	r := serial.NewReader(bytes.NewReader(buf.Bytes()))
	_, ok := r.ReadObjectHeader()
	if !ok {
		t.Fatal("ReadObjectHeader failed")
	}
	got := object.NewTextObject()
	if err := got.Deserialize(r); err != nil {
		t.Fatalf("Deserialize: %v", err)
	}
	cfGot, ok2 := got.Format().(*format.CurrencyFormat)
	if !ok2 {
		t.Fatal("Format should be *format.CurrencyFormat")
	}
	if cfGot.CurrencySymbol != "€" {
		t.Errorf("CurrencySymbol: got %q, want '€'", cfGot.CurrencySymbol)
	}
	if cfGot.PositivePattern != 3 {
		t.Errorf("PositivePattern: got %d, want 3", cfGot.PositivePattern)
	}
}

// TestSerializeTextFormat_CurrencyWithLocale exercises CurrencyFormat with locale.
func TestSerializeTextFormat_CurrencyWithLocale(t *testing.T) {
	orig := object.NewTextObject()
	cf := &format.CurrencyFormat{
		UseLocaleSettings: true,
		DecimalDigits:     3, // non-default
	}
	orig.SetFormat(cf)

	var buf bytes.Buffer
	w := serial.NewWriter(&buf)
	if err := w.WriteObjectNamed("TextObject", orig); err != nil {
		t.Fatalf("Serialize: %v", err)
	}
	_ = w.Flush()
	xml := buf.String()
	if !strings.Contains(xml, `Format="Currency"`) {
		t.Errorf("expected Format=Currency in XML:\n%s", xml)
	}
}

// TestSerializeTextFormat_DateNonDefaults exercises DateFormat with non-default Format string.
// C# DateFormat inherits CustomFormat.Serialize() which only writes the "Format" pattern
// attribute — UseLocale is NOT serialized for Date/Time formats.
func TestSerializeTextFormat_DateNonDefaults(t *testing.T) {
	orig := object.NewTextObject()
	df := &format.DateFormat{
		Format:            "dd-MM-yyyy", // non-default
		UseLocaleSettings: true,
	}
	orig.SetFormat(df)

	var buf bytes.Buffer
	w := serial.NewWriter(&buf)
	if err := w.WriteObjectNamed("TextObject", orig); err != nil {
		t.Fatalf("Serialize: %v", err)
	}
	_ = w.Flush()

	xml := buf.String()
	checks := []string{
		`Format="Date"`,
		`Format.Format=`,
	}
	for _, c := range checks {
		if !strings.Contains(xml, c) {
			t.Errorf("expected %q in XML:\n%s", c, xml)
		}
	}
	// C# does not serialize UseLocale for DateFormat.
	if strings.Contains(xml, "UseLocale") {
		t.Errorf("unexpected UseLocale in XML:\n%s", xml)
	}

	r := serial.NewReader(bytes.NewReader(buf.Bytes()))
	_, ok := r.ReadObjectHeader()
	if !ok {
		t.Fatal("ReadObjectHeader failed")
	}
	got := object.NewTextObject()
	if err := got.Deserialize(r); err != nil {
		t.Fatalf("Deserialize: %v", err)
	}
	dfGot, ok2 := got.Format().(*format.DateFormat)
	if !ok2 {
		t.Fatal("Format should be *format.DateFormat")
	}
	if dfGot.Format != "dd-MM-yyyy" {
		t.Errorf("DateFormat.Format: got %q, want 'dd-MM-yyyy'", dfGot.Format)
	}
}

// TestSerializeTextFormat_TimeNonDefaults exercises TimeFormat with non-default Format.
// C# TimeFormat inherits CustomFormat.Serialize() which only writes the "Format" pattern
// attribute — UseLocale is NOT serialized for Date/Time formats.
func TestSerializeTextFormat_TimeNonDefaults(t *testing.T) {
	orig := object.NewTextObject()
	tf := &format.TimeFormat{
		Format:            "HH:mm:ss.fff", // non-default
		UseLocaleSettings: true,
	}
	orig.SetFormat(tf)

	var buf bytes.Buffer
	w := serial.NewWriter(&buf)
	if err := w.WriteObjectNamed("TextObject", orig); err != nil {
		t.Fatalf("Serialize: %v", err)
	}
	_ = w.Flush()

	xml := buf.String()
	checks := []string{
		`Format="Time"`,
		`Format.Format=`,
	}
	for _, c := range checks {
		if !strings.Contains(xml, c) {
			t.Errorf("expected %q in XML:\n%s", c, xml)
		}
	}
	// C# does not serialize UseLocale for TimeFormat.
	if strings.Contains(xml, "UseLocale") {
		t.Errorf("unexpected UseLocale in XML:\n%s", xml)
	}

	r := serial.NewReader(bytes.NewReader(buf.Bytes()))
	_, ok := r.ReadObjectHeader()
	if !ok {
		t.Fatal("ReadObjectHeader failed")
	}
	got := object.NewTextObject()
	if err := got.Deserialize(r); err != nil {
		t.Fatalf("Deserialize: %v", err)
	}
	tfGot, ok2 := got.Format().(*format.TimeFormat)
	if !ok2 {
		t.Fatal("Format should be *format.TimeFormat")
	}
	if tfGot.Format != "HH:mm:ss.fff" {
		t.Errorf("TimeFormat.Format: got %q", tfGot.Format)
	}
}

// TestSerializeTextFormat_PercentNonDefaults exercises PercentFormat with all
// non-default fields (UseLocaleSettings=false, custom separators, symbol, patterns).
func TestSerializeTextFormat_PercentNonDefaults(t *testing.T) {
	orig := object.NewTextObject()
	pf := &format.PercentFormat{
		UseLocaleSettings: false,
		DecimalDigits:     1,    // non-default
		DecimalSeparator:  ",",  // non-default
		GroupSeparator:    ".",  // non-default
		PercentSymbol:     "pct", // non-default
		PositivePattern:   2,    // non-default
		NegativePattern:   3,    // non-default
	}
	orig.SetFormat(pf)

	var buf bytes.Buffer
	w := serial.NewWriter(&buf)
	if err := w.WriteObjectNamed("TextObject", orig); err != nil {
		t.Fatalf("Serialize: %v", err)
	}
	_ = w.Flush()

	xml := buf.String()
	checks := []string{
		`Format="Percent"`,
		`Format.DecimalSeparator=`,
		`Format.GroupSeparator=`,
		`Format.PercentSymbol=`,
		`Format.PositivePattern=`,
		`Format.NegativePattern=`,
		`Format.UseLocale="false"`,
	}
	for _, c := range checks {
		if !strings.Contains(xml, c) {
			t.Errorf("expected %q in XML:\n%s", c, xml)
		}
	}

	r := serial.NewReader(bytes.NewReader(buf.Bytes()))
	_, ok := r.ReadObjectHeader()
	if !ok {
		t.Fatal("ReadObjectHeader failed")
	}
	got := object.NewTextObject()
	if err := got.Deserialize(r); err != nil {
		t.Fatalf("Deserialize: %v", err)
	}
	pfGot, ok2 := got.Format().(*format.PercentFormat)
	if !ok2 {
		t.Fatal("Format should be *format.PercentFormat")
	}
	if pfGot.PercentSymbol != "pct" {
		t.Errorf("PercentSymbol: got %q, want 'pct'", pfGot.PercentSymbol)
	}
}

// TestSerializeTextFormat_PercentWithLocale exercises PercentFormat with locale=true.
func TestSerializeTextFormat_PercentWithLocale(t *testing.T) {
	orig := object.NewTextObject()
	pf := &format.PercentFormat{
		UseLocaleSettings: true,
		DecimalDigits:     3, // non-default
	}
	orig.SetFormat(pf)

	var buf bytes.Buffer
	w := serial.NewWriter(&buf)
	if err := w.WriteObjectNamed("TextObject", orig); err != nil {
		t.Fatalf("Serialize: %v", err)
	}
	_ = w.Flush()
	xml := buf.String()
	if !strings.Contains(xml, `Format="Percent"`) {
		t.Errorf("expected Format=Percent in XML:\n%s", xml)
	}
}

// TestSerializeTextFormat_BooleanNonDefaults exercises BooleanFormat with non-default
// TrueText and FalseText.
func TestSerializeTextFormat_BooleanNonDefaults(t *testing.T) {
	orig := object.NewTextObject()
	bf := &format.BooleanFormat{
		TrueText:  "Yes",  // non-default
		FalseText: "No",   // non-default
	}
	orig.SetFormat(bf)

	var buf bytes.Buffer
	w := serial.NewWriter(&buf)
	if err := w.WriteObjectNamed("TextObject", orig); err != nil {
		t.Fatalf("Serialize: %v", err)
	}
	_ = w.Flush()

	xml := buf.String()
	checks := []string{`Format="Boolean"`, `Format.TrueText=`, `Format.FalseText=`}
	for _, c := range checks {
		if !strings.Contains(xml, c) {
			t.Errorf("expected %q in XML:\n%s", c, xml)
		}
	}

	r := serial.NewReader(bytes.NewReader(buf.Bytes()))
	_, ok := r.ReadObjectHeader()
	if !ok {
		t.Fatal("ReadObjectHeader failed")
	}
	got := object.NewTextObject()
	if err := got.Deserialize(r); err != nil {
		t.Fatalf("Deserialize: %v", err)
	}
	bfGot, ok2 := got.Format().(*format.BooleanFormat)
	if !ok2 {
		t.Fatal("Format should be *format.BooleanFormat")
	}
	if bfGot.TrueText != "Yes" {
		t.Errorf("TrueText: got %q, want 'Yes'", bfGot.TrueText)
	}
	if bfGot.FalseText != "No" {
		t.Errorf("FalseText: got %q, want 'No'", bfGot.FalseText)
	}
}

// TestSerializeTextFormat_CustomNonDefaults exercises CustomFormat with non-default
// Format string.
func TestSerializeTextFormat_CustomNonDefaults(t *testing.T) {
	orig := object.NewTextObject()
	cf := &format.CustomFormat{
		Format: "##,###.00", // non-default (default="")
	}
	orig.SetFormat(cf)

	var buf bytes.Buffer
	w := serial.NewWriter(&buf)
	if err := w.WriteObjectNamed("TextObject", orig); err != nil {
		t.Fatalf("Serialize: %v", err)
	}
	_ = w.Flush()

	xml := buf.String()
	checks := []string{`Format="Custom"`, `Format.Format=`}
	for _, c := range checks {
		if !strings.Contains(xml, c) {
			t.Errorf("expected %q in XML:\n%s", c, xml)
		}
	}

	r := serial.NewReader(bytes.NewReader(buf.Bytes()))
	_, ok := r.ReadObjectHeader()
	if !ok {
		t.Fatal("ReadObjectHeader failed")
	}
	got := object.NewTextObject()
	if err := got.Deserialize(r); err != nil {
		t.Fatalf("Deserialize: %v", err)
	}
	cfGot, ok2 := got.Format().(*format.CustomFormat)
	if !ok2 {
		t.Fatal("Format should be *format.CustomFormat")
	}
	if cfGot.Format != "##,###.00" {
		t.Errorf("CustomFormat.Format: got %q, want '##,###.00'", cfGot.Format)
	}
}

// TestSerializeTextFormat_GeneralFormat exercises the GeneralFormat case in
// serializeTextFormat (writes nothing — the nil check would have already caught nil,
// but a *GeneralFormat is a non-nil value that still writes nothing).
func TestSerializeTextFormat_GeneralFormat(t *testing.T) {
	orig := object.NewTextObject()
	orig.SetFormat(format.NewGeneralFormat())

	var buf bytes.Buffer
	w := serial.NewWriter(&buf)
	if err := w.WriteObjectNamed("TextObject", orig); err != nil {
		t.Fatalf("Serialize: %v", err)
	}
	_ = w.Flush()

	xml := buf.String()
	// GeneralFormat should produce no Format= attribute.
	if strings.Contains(xml, `Format="`) {
		t.Errorf("GeneralFormat should not write Format attribute; got:\n%s", xml)
	}
}

// ══════════════════════════════════════════════════════════════════════════════
// picture.go — PictureObjectBase full serialize/deserialize
// ══════════════════════════════════════════════════════════════════════════════

// TestPictureObjectBase_SerializeDeserialize_AllFields exercises all non-default
// fields of PictureObjectBase through a full round-trip.
func TestPictureObjectBase_SerializeDeserialize_AllFields(t *testing.T) {
	orig := object.NewPictureObject()
	orig.SetAngle(45)
	orig.SetDataColumn("Products.Photo")
	orig.SetGrayscale(true)
	orig.SetImageLocation("http://example.com/logo.png")
	orig.SetImageSourceExpression("[ImagePath]")
	orig.SetMaxHeight(200)
	orig.SetMaxWidth(300)
	orig.SetPadding(object.Padding{Left: 2, Top: 3, Right: 4, Bottom: 5})
	orig.SetSizeMode(object.SizeModeStretchImage)
	orig.SetImageAlign(object.ImageAlignCenterCenter)
	orig.SetShowErrorImage(true)
	orig.SetTile(true)
	orig.SetTransparency(0.3)

	var buf bytes.Buffer
	w := serial.NewWriter(&buf)
	if err := w.WriteObjectNamed("PictureObject", orig); err != nil {
		t.Fatalf("Serialize: %v", err)
	}
	_ = w.Flush()

	xml := buf.String()
	checks := []string{
		`Angle=`, `DataColumn=`, `Grayscale="true"`, `ImageLocation=`,
		`ImageSourceExpression=`, `MaxHeight=`, `MaxWidth=`, `Padding=`,
		`SizeMode=`, `ImageAlign=`, `ShowErrorImage="true"`,
		`Tile="true"`, `Transparency=`,
	}
	for _, c := range checks {
		if !strings.Contains(xml, c) {
			t.Errorf("expected %q in XML:\n%s", c, xml)
		}
	}

	r := serial.NewReader(bytes.NewReader(buf.Bytes()))
	_, ok := r.ReadObjectHeader()
	if !ok {
		t.Fatal("ReadObjectHeader failed")
	}
	got := object.NewPictureObject()
	if err := got.Deserialize(r); err != nil {
		t.Fatalf("Deserialize: %v", err)
	}
	if got.Angle() != 45 {
		t.Errorf("Angle: got %d, want 45", got.Angle())
	}
	if got.DataColumn() != "Products.Photo" {
		t.Errorf("DataColumn: got %q", got.DataColumn())
	}
	if !got.Grayscale() {
		t.Error("Grayscale should be true")
	}
	if got.ImageLocation() != "http://example.com/logo.png" {
		t.Errorf("ImageLocation: got %q", got.ImageLocation())
	}
	if got.ImageSourceExpression() != "[ImagePath]" {
		t.Errorf("ImageSourceExpression: got %q", got.ImageSourceExpression())
	}
	if got.MaxHeight() != 200 {
		t.Errorf("MaxHeight: got %v, want 200", got.MaxHeight())
	}
	if got.MaxWidth() != 300 {
		t.Errorf("MaxWidth: got %v, want 300", got.MaxWidth())
	}
	if got.SizeMode() != object.SizeModeStretchImage {
		t.Errorf("SizeMode: got %d", got.SizeMode())
	}
	if got.ImageAlign() != object.ImageAlignCenterCenter {
		t.Errorf("ImageAlign: got %d", got.ImageAlign())
	}
	if !got.ShowErrorImage() {
		t.Error("ShowErrorImage should be true")
	}
	if !got.Tile() {
		t.Error("Tile should be true")
	}
	if got.Transparency() != 0.3 {
		t.Errorf("Transparency: got %v, want 0.3", got.Transparency())
	}
}

// ══════════════════════════════════════════════════════════════════════════════
// rich.go — RichObject full serialize/deserialize
// ══════════════════════════════════════════════════════════════════════════════

// TestRichObject_SerializeDeserialize_DefaultFields exercises the empty/default-value
// path of RichObject (text="", canGrow=false), where branches are not taken.
func TestRichObject_SerializeDeserialize_DefaultFields(t *testing.T) {
	orig := object.NewRichObject()
	// Defaults: text="" and canGrow=false — neither branch in Serialize is taken.

	var buf bytes.Buffer
	w := serial.NewWriter(&buf)
	if err := w.WriteObjectNamed("RichObject", orig); err != nil {
		t.Fatalf("Serialize: %v", err)
	}
	_ = w.Flush()

	r := serial.NewReader(bytes.NewReader(buf.Bytes()))
	_, ok := r.ReadObjectHeader()
	if !ok {
		t.Fatal("ReadObjectHeader failed")
	}
	got := object.NewRichObject()
	if err := got.Deserialize(r); err != nil {
		t.Fatalf("Deserialize: %v", err)
	}
	if got.Text() != "" {
		t.Errorf("Text: got %q, want empty", got.Text())
	}
	if got.CanGrow() {
		t.Error("CanGrow should be false")
	}
}

// ══════════════════════════════════════════════════════════════════════════════
// sparkline.go — SparklineObject full serialize/deserialize
// ══════════════════════════════════════════════════════════════════════════════

// TestSparklineObject_SerializeDeserialize_DefaultFields exercises the empty path
// of SparklineObject (ChartData="", Dock="").
func TestSparklineObject_SerializeDeserialize_DefaultFields(t *testing.T) {
	orig := object.NewSparklineObject()
	// Defaults: both fields empty — neither branch in Serialize is taken.

	var buf bytes.Buffer
	w := serial.NewWriter(&buf)
	if err := w.WriteObjectNamed("SparklineObject", orig); err != nil {
		t.Fatalf("Serialize: %v", err)
	}
	_ = w.Flush()

	r := serial.NewReader(bytes.NewReader(buf.Bytes()))
	_, ok := r.ReadObjectHeader()
	if !ok {
		t.Fatal("ReadObjectHeader failed")
	}
	got := object.NewSparklineObject()
	if err := got.Deserialize(r); err != nil {
		t.Fatalf("Deserialize: %v", err)
	}
	if got.ChartData != "" {
		t.Errorf("ChartData: got %q, want empty", got.ChartData)
	}
	if got.Dock != "" {
		t.Errorf("Dock: got %q, want empty", got.Dock)
	}
}

// ══════════════════════════════════════════════════════════════════════════════
// rfid.go — RFIDLabel serialize/deserialize full paths
// ══════════════════════════════════════════════════════════════════════════════

// TestRFIDLabel_SerializeDeserialize_AllBanksAndPasswords exercises all non-default
// fields in RFIDLabel.Serialize and Deserialize including all three banks,
// passwords, lock types, bool flags, and ErrorHandle.
func TestRFIDLabel_SerializeDeserialize_AllBanksAndPasswords(t *testing.T) {
	orig := object.NewRFIDLabel()
	orig.EPCBank = object.RFIDBank{
		Data:       "AABBCCDD",
		DataColumn: "EPCCol",
		Offset:     4,
		DataFormat: object.RFIDBankFormatASCII,
	}
	orig.TIDBank = object.RFIDBank{
		Data:       "T1D2",
		DataColumn: "TIDCol",
		Offset:     2,
		DataFormat: object.RFIDBankFormatDecimal,
	}
	orig.UserBank = object.RFIDBank{
		Data:       "USERDATA",
		DataColumn: "UserCol",
		Offset:     0,
		DataFormat: object.RFIDBankFormatHex,
	}
	orig.AccessPassword = "PASS1"
	orig.AccessPasswordDataColumn = "AccCol"
	orig.KillPassword = "KILL1"
	orig.KillPasswordDataColumn = "KillCol"
	orig.LockKillPassword = object.RFIDLockTypePermanentLock
	orig.LockAccessPassword = object.RFIDLockTypeLock
	orig.LockEPCBank = object.RFIDLockTypeLock // non-default to ensure LockEPCBlock= appears in XML
	orig.LockUserBank = object.RFIDLockTypeLock
	orig.UseAdjustForEPC = true
	orig.RewriteEPCBank = true
	orig.ErrorHandle = object.RFIDErrorHandleError

	var buf bytes.Buffer
	w := serial.NewWriter(&buf)
	if err := w.WriteObjectNamed("RFIDLabel", orig); err != nil {
		t.Fatalf("Serialize: %v", err)
	}
	_ = w.Flush()

	xml := buf.String()
	// C# key names: "EpcBank", "TidBank", "LockEPCBlock", "LockUserBlock", "RewriteEPCbank"
	// (RFIDLabel.cs:463-498)
	checks := []string{
		`EpcBank.Data=`, `EpcBank.DataColumn=`, `EpcBank.Offset=`, `EpcBank.DataFormat=`,
		`TidBank.Data=`, `TidBank.DataColumn=`, `TidBank.Offset=`, `TidBank.DataFormat=`,
		`UserBank.Data=`, `UserBank.DataColumn=`,
		`AccessPassword=`, `AccessPasswordDataColumn=`,
		`KillPassword=`, `KillPasswordDataColumn=`,
		`LockKillPassword=`, `LockAccessPassword=`, `LockEPCBlock=`, `LockUserBlock=`,
		`UseAdjustForEPC="true"`, `RewriteEPCbank="true"`, `ErrorHandle=`,
	}
	for _, c := range checks {
		if !strings.Contains(xml, c) {
			t.Errorf("expected %q in XML:\n%s", c, xml)
		}
	}

	r := serial.NewReader(bytes.NewReader(buf.Bytes()))
	_, ok := r.ReadObjectHeader()
	if !ok {
		t.Fatal("ReadObjectHeader failed")
	}
	got := object.NewRFIDLabel()
	if err := got.Deserialize(r); err != nil {
		t.Fatalf("Deserialize: %v", err)
	}
	if got.EPCBank.Data != "AABBCCDD" {
		t.Errorf("EPCBank.Data: got %q", got.EPCBank.Data)
	}
	if got.EPCBank.DataColumn != "EPCCol" {
		t.Errorf("EPCBank.DataColumn: got %q", got.EPCBank.DataColumn)
	}
	if got.EPCBank.Offset != 4 {
		t.Errorf("EPCBank.Offset: got %d", got.EPCBank.Offset)
	}
	if got.EPCBank.DataFormat != object.RFIDBankFormatASCII {
		t.Errorf("EPCBank.DataFormat: got %v", got.EPCBank.DataFormat)
	}
	if got.TIDBank.DataFormat != object.RFIDBankFormatDecimal {
		t.Errorf("TIDBank.DataFormat: got %v", got.TIDBank.DataFormat)
	}
	if got.AccessPassword != "PASS1" {
		t.Errorf("AccessPassword: got %q", got.AccessPassword)
	}
	if got.KillPassword != "KILL1" {
		t.Errorf("KillPassword: got %q", got.KillPassword)
	}
	if got.LockKillPassword != object.RFIDLockTypePermanentLock {
		t.Errorf("LockKillPassword: got %v", got.LockKillPassword)
	}
	if got.LockAccessPassword != object.RFIDLockTypeLock {
		t.Errorf("LockAccessPassword: got %v", got.LockAccessPassword)
	}
	if got.LockEPCBank != object.RFIDLockTypeLock {
		t.Errorf("LockEPCBank: got %v", got.LockEPCBank)
	}
	if got.LockUserBank != object.RFIDLockTypeLock {
		t.Errorf("LockUserBank: got %v", got.LockUserBank)
	}
	if !got.UseAdjustForEPC {
		t.Error("UseAdjustForEPC should be true")
	}
	if !got.RewriteEPCBank {
		t.Error("RewriteEPCBank should be true")
	}
	if got.ErrorHandle != object.RFIDErrorHandleError {
		t.Errorf("ErrorHandle: got %v", got.ErrorHandle)
	}
}

// ══════════════════════════════════════════════════════════════════════════════
// cellular_text.go — CellularTextObject serialize/deserialize
// ══════════════════════════════════════════════════════════════════════════════

// TestCellularTextObject_SerializeDeserialize_DefaultFields exercises the default
// path (all fields zero / wordWrap=true so no branches written).
func TestCellularTextObject_SerializeDeserialize_DefaultFields(t *testing.T) {
	orig := object.NewCellularTextObject()
	// All zeros + wordWrap=true → none of the non-default branches are taken.

	var buf bytes.Buffer
	w := serial.NewWriter(&buf)
	if err := w.WriteObjectNamed("CellularTextObject", orig); err != nil {
		t.Fatalf("Serialize: %v", err)
	}
	_ = w.Flush()

	xml := buf.String()
	// None of the optional cellular fields should appear.
	for _, c := range []string{`CellWidth=`, `CellHeight=`, `HorzSpacing=`, `VertSpacing=`, `WordWrap=`} {
		if strings.Contains(xml, c) {
			t.Errorf("unexpected %q in default XML:\n%s", c, xml)
		}
	}

	r := serial.NewReader(bytes.NewReader(buf.Bytes()))
	_, ok := r.ReadObjectHeader()
	if !ok {
		t.Fatal("ReadObjectHeader failed")
	}
	got := object.NewCellularTextObject()
	if err := got.Deserialize(r); err != nil {
		t.Fatalf("Deserialize: %v", err)
	}
	if got.CellWidth() != 0 {
		t.Errorf("CellWidth: got %v, want 0", got.CellWidth())
	}
	if !got.WordWrap() {
		t.Error("WordWrap should be true (default)")
	}
}

// TestCellularTextObject_SerializeDeserialize_AllNonDefaults exercises all
// non-default branches of CellularTextObject serialization.
func TestCellularTextObject_SerializeDeserialize_AllNonDefaults(t *testing.T) {
	orig := object.NewCellularTextObject()
	orig.SetCellWidth(28)
	orig.SetCellHeight(32)
	orig.SetHorzSpacing(4)
	orig.SetVertSpacing(6)
	orig.SetWordWrap(false)

	var buf bytes.Buffer
	w := serial.NewWriter(&buf)
	if err := w.WriteObjectNamed("CellularTextObject", orig); err != nil {
		t.Fatalf("Serialize: %v", err)
	}
	_ = w.Flush()

	xml := buf.String()
	checks := []string{`CellWidth=`, `CellHeight=`, `HorzSpacing=`, `VertSpacing=`, `WordWrap="false"`}
	for _, c := range checks {
		if !strings.Contains(xml, c) {
			t.Errorf("expected %q in XML:\n%s", c, xml)
		}
	}

	r := serial.NewReader(bytes.NewReader(buf.Bytes()))
	_, ok := r.ReadObjectHeader()
	if !ok {
		t.Fatal("ReadObjectHeader failed")
	}
	got := object.NewCellularTextObject()
	if err := got.Deserialize(r); err != nil {
		t.Fatalf("Deserialize: %v", err)
	}
	if got.CellWidth() != 28 {
		t.Errorf("CellWidth: got %v, want 28", got.CellWidth())
	}
	if got.CellHeight() != 32 {
		t.Errorf("CellHeight: got %v, want 32", got.CellHeight())
	}
	if got.HorzSpacing() != 4 {
		t.Errorf("HorzSpacing: got %v, want 4", got.HorzSpacing())
	}
	if got.VertSpacing() != 6 {
		t.Errorf("VertSpacing: got %v, want 6", got.VertSpacing())
	}
	if got.WordWrap() {
		t.Error("WordWrap should be false after round-trip")
	}
}

// ══════════════════════════════════════════════════════════════════════════════
// digital_signature.go — DigitalSignatureObject serialize/deserialize
// ══════════════════════════════════════════════════════════════════════════════

// TestDigitalSignatureObject_SerializeDeserialize_DefaultFields exercises
// DigitalSignatureObject with default empty placeholder.
func TestDigitalSignatureObject_SerializeDeserialize_DefaultFields(t *testing.T) {
	orig := object.NewDigitalSignatureObject()
	// placeholder="" → no branch taken in Serialize.

	var buf bytes.Buffer
	w := serial.NewWriter(&buf)
	if err := w.WriteObjectNamed("DigitalSignatureObject", orig); err != nil {
		t.Fatalf("Serialize: %v", err)
	}
	_ = w.Flush()

	xml := buf.String()
	if strings.Contains(xml, `Placeholder=`) {
		t.Errorf("unexpected Placeholder in default XML:\n%s", xml)
	}

	r := serial.NewReader(bytes.NewReader(buf.Bytes()))
	_, ok := r.ReadObjectHeader()
	if !ok {
		t.Fatal("ReadObjectHeader failed")
	}
	got := object.NewDigitalSignatureObject()
	if err := got.Deserialize(r); err != nil {
		t.Fatalf("Deserialize: %v", err)
	}
	if got.Placeholder() != "" {
		t.Errorf("Placeholder: got %q, want empty", got.Placeholder())
	}
}

// TestDigitalSignatureObject_TypeAndBase verifies TypeName and BaseName.
func TestDigitalSignatureObject_TypeAndBase(t *testing.T) {
	d := object.NewDigitalSignatureObject()
	if d.TypeName() != "DigitalSignatureObject" {
		t.Errorf("TypeName: got %q", d.TypeName())
	}
	if d.BaseName() != "DigitalSignature" {
		t.Errorf("BaseName: got %q", d.BaseName())
	}
}

// ══════════════════════════════════════════════════════════════════════════════
// html.go — HtmlObject serialize/deserialize
// ══════════════════════════════════════════════════════════════════════════════

// TestHtmlObject_SerializeDeserialize_DefaultFields exercises HtmlObject with
// rightToLeft=false (default) — the if-branch in Serialize is NOT taken.
func TestHtmlObject_SerializeDeserialize_DefaultFields(t *testing.T) {
	orig := object.NewHtmlObject()
	// rightToLeft=false → branch not taken.

	var buf bytes.Buffer
	w := serial.NewWriter(&buf)
	if err := w.WriteObjectNamed("HtmlObject", orig); err != nil {
		t.Fatalf("Serialize: %v", err)
	}
	_ = w.Flush()

	xml := buf.String()
	if strings.Contains(xml, `RightToLeft=`) {
		t.Errorf("unexpected RightToLeft in default XML:\n%s", xml)
	}

	r := serial.NewReader(bytes.NewReader(buf.Bytes()))
	_, ok := r.ReadObjectHeader()
	if !ok {
		t.Fatal("ReadObjectHeader failed")
	}
	got := object.NewHtmlObject()
	if err := got.Deserialize(r); err != nil {
		t.Fatalf("Deserialize: %v", err)
	}
	if got.RightToLeft() {
		t.Error("RightToLeft should be false (default)")
	}
}

// ══════════════════════════════════════════════════════════════════════════════
// svg.go — SVGObject serialize/deserialize
// ══════════════════════════════════════════════════════════════════════════════

// TestSVGObject_SerializeDeserialize_DefaultFields exercises SVGObject with
// SvgData="" (default) — the if-branch in Serialize is NOT taken.
func TestSVGObject_SerializeDeserialize_DefaultFields(t *testing.T) {
	orig := object.NewSVGObject()
	// SvgData="" → branch not taken.

	var buf bytes.Buffer
	w := serial.NewWriter(&buf)
	if err := w.WriteObjectNamed("SVGObject", orig); err != nil {
		t.Fatalf("Serialize: %v", err)
	}
	_ = w.Flush()

	xml := buf.String()
	if strings.Contains(xml, `SvgData=`) {
		t.Errorf("unexpected SvgData in default XML:\n%s", xml)
	}

	r := serial.NewReader(bytes.NewReader(buf.Bytes()))
	_, ok := r.ReadObjectHeader()
	if !ok {
		t.Fatal("ReadObjectHeader failed")
	}
	got := object.NewSVGObject()
	if err := got.Deserialize(r); err != nil {
		t.Fatalf("Deserialize: %v", err)
	}
	if got.SvgData != "" {
		t.Errorf("SvgData: got %q, want empty", got.SvgData)
	}
}

// TestSVGObject_TypeAndBase verifies TypeName and BaseName.
func TestSVGObject_TypeAndBase(t *testing.T) {
	s := object.NewSVGObject()
	if s.TypeName() != "SVGObject" {
		t.Errorf("TypeName: got %q", s.TypeName())
	}
	if s.BaseName() != "SVG" {
		t.Errorf("BaseName: got %q", s.BaseName())
	}
}
