package image_test

// image_serial_test.go — tests for ImageExport Serialize / Deserialize.
//
// Mirrors C# ImageExport.Serialize (ImageExport.cs line 694).

import (
	"testing"

	imgexport "github.com/andrewloable/go-fastreport/export/image"
	"github.com/andrewloable/go-fastreport/report"
)

// ── test doubles for report.Writer / report.Reader ────────────────────────────

type serialWriter struct {
	strs  map[string]string
	ints  map[string]int
	bools map[string]bool
}

func newSerialWriter() *serialWriter {
	return &serialWriter{
		strs:  make(map[string]string),
		ints:  make(map[string]int),
		bools: make(map[string]bool),
	}
}

func (w *serialWriter) WriteStr(name, value string)              { w.strs[name] = value }
func (w *serialWriter) WriteInt(name string, value int)           { w.ints[name] = value }
func (w *serialWriter) WriteBool(name string, value bool)         { w.bools[name] = value }
func (w *serialWriter) WriteFloat(name string, value float32)     {}
func (w *serialWriter) WriteObject(obj report.Serializable) error { return nil }
func (w *serialWriter) WriteObjectNamed(_ string, _ report.Serializable) error { return nil }

type serialReader struct {
	strs  map[string]string
	ints  map[string]int
	bools map[string]bool
}

func newSerialReader(strs map[string]string, ints map[string]int, bools map[string]bool) *serialReader {
	if strs == nil {
		strs = map[string]string{}
	}
	if ints == nil {
		ints = map[string]int{}
	}
	if bools == nil {
		bools = map[string]bool{}
	}
	return &serialReader{strs: strs, ints: ints, bools: bools}
}

func (r *serialReader) ReadStr(name, def string) string {
	if v, ok := r.strs[name]; ok {
		return v
	}
	return def
}
func (r *serialReader) ReadInt(name string, def int) int {
	if v, ok := r.ints[name]; ok {
		return v
	}
	return def
}
func (r *serialReader) ReadBool(name string, def bool) bool {
	if v, ok := r.bools[name]; ok {
		return v
	}
	return def
}
func (r *serialReader) ReadFloat(name string, def float32) float32 { return def }
func (r *serialReader) NextChild() (string, bool)                  { return "", false }
func (r *serialReader) FinishChild() error                         { return nil }

// ── Serialize ─────────────────────────────────────────────────────────────────

func TestImageExport_Serialize_DefaultValues(t *testing.T) {
	exp := imgexport.NewExporter()
	w := newSerialWriter()
	exp.Serialize(w)

	// Check image-specific fields — defaults should be written (unconditionally).
	if w.ints["ImageFormat"] != int(imgexport.ImageFormatJPEG) {
		t.Errorf("ImageFormat: got %d, want %d (JPEG)", w.ints["ImageFormat"], int(imgexport.ImageFormatJPEG))
	}
	if !w.bools["SeparateFiles"] {
		t.Error("SeparateFiles should be written as true (default)")
	}
	if w.ints["ResolutionX"] != imgexport.DefaultDPI {
		t.Errorf("ResolutionX: got %d, want %d", w.ints["ResolutionX"], imgexport.DefaultDPI)
	}
	if w.ints["ResolutionY"] != imgexport.DefaultDPI {
		t.Errorf("ResolutionY: got %d, want %d", w.ints["ResolutionY"], imgexport.DefaultDPI)
	}
	if w.ints["JpegQuality"] != 100 {
		t.Errorf("JpegQuality: got %d, want 100", w.ints["JpegQuality"])
	}
	if w.bools["MultiFrameTiff"] {
		t.Error("MultiFrameTiff should be written as false (default)")
	}
	if w.bools["MonochromeTiff"] {
		t.Error("MonochromeTiff should be written as false (default)")
	}
}

func TestImageExport_Serialize_NonDefaults(t *testing.T) {
	exp := imgexport.NewExporter()
	exp.Format = imgexport.ImageFormatTIFF
	exp.SeparateFiles = false
	exp.ResolutionX = 300
	exp.ResolutionY = 300
	exp.JpegQuality = 85
	exp.MultiFrameTiff = true
	exp.MonochromeTiff = true

	w := newSerialWriter()
	exp.Serialize(w)

	if w.ints["ImageFormat"] != int(imgexport.ImageFormatTIFF) {
		t.Errorf("ImageFormat: got %d, want TIFF (%d)", w.ints["ImageFormat"], int(imgexport.ImageFormatTIFF))
	}
	if w.bools["SeparateFiles"] {
		t.Error("SeparateFiles should be false")
	}
	if w.ints["ResolutionX"] != 300 {
		t.Errorf("ResolutionX: got %d, want 300", w.ints["ResolutionX"])
	}
	if w.ints["ResolutionY"] != 300 {
		t.Errorf("ResolutionY: got %d, want 300", w.ints["ResolutionY"])
	}
	if w.ints["JpegQuality"] != 85 {
		t.Errorf("JpegQuality: got %d, want 85", w.ints["JpegQuality"])
	}
	if !w.bools["MultiFrameTiff"] {
		t.Error("MultiFrameTiff should be true")
	}
	if !w.bools["MonochromeTiff"] {
		t.Error("MonochromeTiff should be true")
	}
}

// ── Deserialize ───────────────────────────────────────────────────────────────

func TestImageExport_Deserialize_Defaults(t *testing.T) {
	// An empty reader should restore all fields to defaults.
	exp := imgexport.NewExporter()
	exp.Format = imgexport.ImageFormatPNG // pre-set to non-default
	exp.ResolutionX = 300                 // pre-set to non-default

	r := newSerialReader(nil, nil, nil)
	exp.Deserialize(r)

	if exp.Format != imgexport.ImageFormatJPEG {
		t.Errorf("Format: got %d, want ImageFormatJPEG", exp.Format)
	}
	if exp.ResolutionX != imgexport.DefaultDPI {
		t.Errorf("ResolutionX: got %d, want %d", exp.ResolutionX, imgexport.DefaultDPI)
	}
	if !exp.SeparateFiles {
		t.Error("SeparateFiles should default to true")
	}
	if exp.JpegQuality != 100 {
		t.Errorf("JpegQuality: got %d, want 100", exp.JpegQuality)
	}
}

func TestImageExport_Deserialize_NonDefaults(t *testing.T) {
	exp := imgexport.NewExporter()

	r := newSerialReader(
		nil,
		map[string]int{
			"ImageFormat": int(imgexport.ImageFormatBMP),
			"ResolutionX": 150,
			"ResolutionY": 150,
			"JpegQuality": 75,
		},
		map[string]bool{
			"SeparateFiles":  false,
			"MultiFrameTiff": true,
			"MonochromeTiff": true,
		},
	)
	exp.Deserialize(r)

	if exp.Format != imgexport.ImageFormatBMP {
		t.Errorf("Format: got %d, want ImageFormatBMP", exp.Format)
	}
	if exp.SeparateFiles {
		t.Error("SeparateFiles should be false")
	}
	if exp.ResolutionX != 150 {
		t.Errorf("ResolutionX: got %d, want 150", exp.ResolutionX)
	}
	if exp.ResolutionY != 150 {
		t.Errorf("ResolutionY: got %d, want 150", exp.ResolutionY)
	}
	if exp.JpegQuality != 75 {
		t.Errorf("JpegQuality: got %d, want 75", exp.JpegQuality)
	}
	if !exp.MultiFrameTiff {
		t.Error("MultiFrameTiff should be true")
	}
	if !exp.MonochromeTiff {
		t.Error("MonochromeTiff should be true")
	}
}

// TestImageExport_SerializeDeserialize_RoundTrip verifies that a Serialize
// followed by Deserialize restores all settings faithfully.
func TestImageExport_SerializeDeserialize_RoundTrip(t *testing.T) {
	orig := imgexport.NewExporter()
	orig.Format = imgexport.ImageFormatGIF
	orig.SeparateFiles = false
	orig.ResolutionX = 200
	orig.ResolutionY = 150
	orig.JpegQuality = 60
	orig.MultiFrameTiff = true
	orig.MonochromeTiff = false

	// Serialize.
	w := newSerialWriter()
	orig.Serialize(w)

	// Deserialize into a fresh exporter.
	dst := imgexport.NewExporter()
	r := newSerialReader(w.strs, w.ints, w.bools)
	dst.Deserialize(r)

	if dst.Format != orig.Format {
		t.Errorf("Format: got %d, want %d", dst.Format, orig.Format)
	}
	if dst.SeparateFiles != orig.SeparateFiles {
		t.Errorf("SeparateFiles: got %v, want %v", dst.SeparateFiles, orig.SeparateFiles)
	}
	if dst.ResolutionX != orig.ResolutionX {
		t.Errorf("ResolutionX: got %d, want %d", dst.ResolutionX, orig.ResolutionX)
	}
	if dst.ResolutionY != orig.ResolutionY {
		t.Errorf("ResolutionY: got %d, want %d", dst.ResolutionY, orig.ResolutionY)
	}
	if dst.JpegQuality != orig.JpegQuality {
		t.Errorf("JpegQuality: got %d, want %d", dst.JpegQuality, orig.JpegQuality)
	}
	if dst.MultiFrameTiff != orig.MultiFrameTiff {
		t.Errorf("MultiFrameTiff: got %v, want %v", dst.MultiFrameTiff, orig.MultiFrameTiff)
	}
	if dst.MonochromeTiff != orig.MonochromeTiff {
		t.Errorf("MonochromeTiff: got %v, want %v", dst.MonochromeTiff, orig.MonochromeTiff)
	}
}
