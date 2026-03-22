package utils_test

import (
	"bytes"
	"compress/gzip"
	"io"
	"strings"
	"testing"

	"github.com/andrewloable/go-fastreport/utils"
)

// gzipBytes compresses data using gzip and returns the result.
func gzipBytes(t *testing.T, data []byte) []byte {
	t.Helper()
	var buf bytes.Buffer
	w := gzip.NewWriter(&buf)
	if _, err := w.Write(data); err != nil {
		t.Fatalf("gzip write: %v", err)
	}
	if err := w.Close(); err != nil {
		t.Fatalf("gzip close: %v", err)
	}
	return buf.Bytes()
}

// ── RegisterResource / GetStream ──────────────────────────────────────────────

func TestGetStream_Unregistered(t *testing.T) {
	rc, err := utils.GetStream("NonExistentAssembly", "no-such-resource.xml")
	if err != nil {
		t.Fatalf("expected nil error for unregistered resource, got: %v", err)
	}
	if rc != nil {
		rc.Close()
		t.Fatal("expected nil reader for unregistered resource")
	}
}

func TestGetStreamFR_Unregistered(t *testing.T) {
	rc, err := utils.GetStreamFR("absolutely-not-registered.bin")
	if err != nil {
		t.Fatalf("expected nil error, got: %v", err)
	}
	if rc != nil {
		rc.Close()
		t.Fatal("expected nil reader")
	}
}

func TestRegisterResourceBytes_RoundTrip(t *testing.T) {
	const assembly = "TestAssembly"
	const name = "test-resource.txt"
	payload := []byte("hello from resource loader")

	utils.RegisterResourceBytes(assembly, name, payload)

	rc, err := utils.GetStream(assembly, name)
	if err != nil {
		t.Fatalf("GetStream: %v", err)
	}
	if rc == nil {
		t.Fatal("GetStream returned nil for registered resource")
	}
	defer rc.Close()

	got, err := io.ReadAll(rc)
	if err != nil {
		t.Fatalf("ReadAll: %v", err)
	}
	if !bytes.Equal(got, payload) {
		t.Errorf("content = %q, want %q", got, payload)
	}
}

func TestGetStreamFR_RegisteredInFastReport(t *testing.T) {
	const name = "fr-test-resource.dat"
	payload := []byte("FastReport resource content")

	utils.RegisterResourceBytes("FastReport", name, payload)

	rc, err := utils.GetStreamFR(name)
	if err != nil {
		t.Fatalf("GetStreamFR: %v", err)
	}
	if rc == nil {
		t.Fatal("GetStreamFR returned nil")
	}
	defer rc.Close()

	got, err := io.ReadAll(rc)
	if err != nil {
		t.Fatalf("ReadAll: %v", err)
	}
	if string(got) != string(payload) {
		t.Errorf("got %q, want %q", got, payload)
	}
}

// Each call to GetStream must return an independent reader (fresh copy).
func TestGetStream_IndependentReaders(t *testing.T) {
	const assembly = "TestAssembly"
	const name = "independent-resource.bin"
	payload := []byte("shared payload data")

	utils.RegisterResourceBytes(assembly, name, payload)

	rc1, err := utils.GetStream(assembly, name)
	if err != nil || rc1 == nil {
		t.Fatal("first GetStream failed")
	}
	defer rc1.Close()

	rc2, err := utils.GetStream(assembly, name)
	if err != nil || rc2 == nil {
		t.Fatal("second GetStream failed")
	}
	defer rc2.Close()

	// Read from rc1 fully.
	got1, _ := io.ReadAll(rc1)
	// rc2 should still be at position 0.
	got2, _ := io.ReadAll(rc2)

	if !bytes.Equal(got1, payload) {
		t.Errorf("rc1 content = %q, want %q", got1, payload)
	}
	if !bytes.Equal(got2, payload) {
		t.Errorf("rc2 content = %q, want %q", got2, payload)
	}
}

// Registering the same key again replaces the provider.
func TestRegisterResource_Replacement(t *testing.T) {
	const assembly = "TestAssembly"
	const name = "replaceable.txt"

	utils.RegisterResourceBytes(assembly, name, []byte("first"))
	utils.RegisterResourceBytes(assembly, name, []byte("second"))

	rc, err := utils.GetStream(assembly, name)
	if err != nil || rc == nil {
		t.Fatal("GetStream failed after replacement")
	}
	defer rc.Close()

	got, _ := io.ReadAll(rc)
	if string(got) != "second" {
		t.Errorf("after replacement got %q, want %q", got, "second")
	}
}

// ── UnpackStream ──────────────────────────────────────────────────────────────

func TestUnpackStream_Unregistered(t *testing.T) {
	rc, err := utils.UnpackStream("NoAssembly", "no-resource.gz")
	if err != nil {
		t.Fatalf("expected nil error for unregistered resource, got: %v", err)
	}
	if rc != nil {
		rc.Close()
		t.Fatal("expected nil reader for unregistered resource")
	}
}

func TestUnpackStreamFR_Unregistered(t *testing.T) {
	rc, err := utils.UnpackStreamFR("no-gz-resource.gz")
	if err != nil {
		t.Fatalf("expected nil error, got: %v", err)
	}
	if rc != nil {
		rc.Close()
		t.Fatal("expected nil reader")
	}
}

func TestUnpackStream_GzipRoundTrip(t *testing.T) {
	const assembly = "TestAssembly"
	const name = "packed-resource.gz"
	original := []byte("this content was gzip-compressed before registration")

	compressed := gzipBytes(t, original)
	utils.RegisterResourceBytes(assembly, name, compressed)

	rc, err := utils.UnpackStream(assembly, name)
	if err != nil {
		t.Fatalf("UnpackStream: %v", err)
	}
	if rc == nil {
		t.Fatal("UnpackStream returned nil")
	}
	defer rc.Close()

	got, err := io.ReadAll(rc)
	if err != nil {
		t.Fatalf("ReadAll: %v", err)
	}
	if !bytes.Equal(got, original) {
		t.Errorf("decompressed = %q, want %q", got, original)
	}
}

func TestUnpackStreamFR_GzipRoundTrip(t *testing.T) {
	const name = "fr-packed.gz"
	original := []byte("FastReport compressed resource")

	utils.RegisterResourceBytes("FastReport", name, gzipBytes(t, original))

	rc, err := utils.UnpackStreamFR(name)
	if err != nil {
		t.Fatalf("UnpackStreamFR: %v", err)
	}
	if rc == nil {
		t.Fatal("UnpackStreamFR returned nil")
	}
	defer rc.Close()

	got, _ := io.ReadAll(rc)
	if !bytes.Equal(got, original) {
		t.Errorf("decompressed = %q, want %q", got, original)
	}
}

func TestUnpackStream_InvalidGzip(t *testing.T) {
	const assembly = "TestAssembly"
	const name = "bad-gzip.gz"

	utils.RegisterResourceBytes(assembly, name, []byte("not gzip data at all"))

	_, err := utils.UnpackStream(assembly, name)
	if err == nil {
		t.Fatal("expected error for invalid gzip data, got nil")
	}
	if !strings.Contains(err.Error(), "gzip.NewReader") {
		t.Errorf("error should mention gzip.NewReader, got: %v", err)
	}
}

// ── RegisterResource with custom provider ─────────────────────────────────────

func TestRegisterResource_CustomProvider(t *testing.T) {
	const assembly = "TestAssembly"
	const name = "custom-provider.txt"
	callCount := 0

	utils.RegisterResource(assembly, name, func() (io.ReadCloser, error) {
		callCount++
		return io.NopCloser(strings.NewReader("from custom provider")), nil
	})

	// Call twice to verify the provider is called each time.
	for i := 0; i < 2; i++ {
		rc, err := utils.GetStream(assembly, name)
		if err != nil || rc == nil {
			t.Fatalf("GetStream call %d: err=%v, rc=%v", i+1, err, rc)
		}
		rc.Close()
	}

	if callCount != 2 {
		t.Errorf("provider called %d times, want 2", callCount)
	}
}
