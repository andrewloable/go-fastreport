package reportpkg_test

import (
	"bytes"
	"testing"

	"github.com/andrewloable/go-fastreport/reportpkg"
)

func TestCompression_RoundTrip(t *testing.T) {
	// Build a minimal report.
	r := reportpkg.NewReport()
	r.Info.Name = "CompressionTest"
	pg := reportpkg.NewReportPage()
	pg.SetName("Page1")
	r.AddPage(pg)
	r.Compressed = true

	// Save compressed.
	var buf bytes.Buffer
	if err := r.SaveTo(&buf); err != nil {
		t.Fatalf("SaveTo compressed: %v", err)
	}
	compressed := buf.Bytes()

	// First two bytes must be gzip magic.
	if len(compressed) < 2 || compressed[0] != 0x1f || compressed[1] != 0x8b {
		t.Fatalf("expected gzip magic bytes, got %x %x", compressed[0], compressed[1])
	}

	// Load back — should auto-decompress.
	r2 := reportpkg.NewReport()
	if err := r2.LoadFrom(bytes.NewReader(compressed)); err != nil {
		t.Fatalf("LoadFrom compressed: %v", err)
	}
	if r2.Info.Name != "CompressionTest" {
		t.Errorf("Info.Name = %q, want %q", r2.Info.Name, "CompressionTest")
	}
	if len(r2.Pages()) != 1 {
		t.Errorf("page count = %d, want 1", len(r2.Pages()))
	}
}

func TestCompression_UncompressedStillLoads(t *testing.T) {
	r := reportpkg.NewReport()
	r.Info.Name = "PlainTest"
	r.AddPage(reportpkg.NewReportPage())

	var buf bytes.Buffer
	if err := r.SaveTo(&buf); err != nil {
		t.Fatalf("SaveTo: %v", err)
	}

	r2 := reportpkg.NewReport()
	if err := r2.LoadFrom(bytes.NewReader(buf.Bytes())); err != nil {
		t.Fatalf("LoadFrom plain: %v", err)
	}
	if r2.Info.Name != "PlainTest" {
		t.Errorf("Info.Name = %q, want %q", r2.Info.Name, "PlainTest")
	}
}
