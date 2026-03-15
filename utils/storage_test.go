package utils_test

import (
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/andrewloable/go-fastreport/utils"
)

// ── MemoryStorageService tests ─────────────────────────────────────────────────

func TestMemoryStorageService_SaveLoad(t *testing.T) {
	m := utils.NewMemoryStorageService()
	data := []byte("hello report")

	if err := m.Save("report.frx", data); err != nil {
		t.Fatalf("Save: %v", err)
	}
	got, err := m.Load("report.frx")
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if string(got) != "hello report" {
		t.Errorf("Load = %q, want %q", got, "hello report")
	}
}

func TestMemoryStorageService_LoadMissing(t *testing.T) {
	m := utils.NewMemoryStorageService()
	_, err := m.Load("missing.frx")
	if !os.IsNotExist(err) {
		t.Errorf("Load missing: expected ErrNotExist, got %v", err)
	}
}

func TestMemoryStorageService_Exists(t *testing.T) {
	m := utils.NewMemoryStorageService()
	if m.Exists("x") {
		t.Error("Exists on empty store should be false")
	}
	m.Put("x", []byte("data"))
	if !m.Exists("x") {
		t.Error("Exists after Put should be true")
	}
}

func TestMemoryStorageService_Reader(t *testing.T) {
	m := utils.NewMemoryStorageService()
	m.Put("r.txt", []byte("streaming"))

	rc, err := m.Reader("r.txt")
	if err != nil {
		t.Fatalf("Reader: %v", err)
	}
	defer rc.Close()

	got, err := io.ReadAll(rc)
	if err != nil {
		t.Fatalf("ReadAll: %v", err)
	}
	if string(got) != "streaming" {
		t.Errorf("Reader content = %q", got)
	}
}

func TestMemoryStorageService_Writer(t *testing.T) {
	m := utils.NewMemoryStorageService()

	wc, err := m.Writer("w.txt")
	if err != nil {
		t.Fatalf("Writer: %v", err)
	}
	if _, err := wc.Write([]byte("written")); err != nil {
		t.Fatalf("Write: %v", err)
	}
	if err := wc.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}

	got, err := m.Load("w.txt")
	if err != nil {
		t.Fatalf("Load after Writer: %v", err)
	}
	if string(got) != "written" {
		t.Errorf("Writer content = %q, want %q", got, "written")
	}
}

func TestMemoryStorageService_IsolatesData(t *testing.T) {
	// Mutating returned bytes should not affect the stored value.
	m := utils.NewMemoryStorageService()
	m.Put("iso.txt", []byte("original"))

	got, _ := m.Load("iso.txt")
	got[0] = 'X'

	got2, _ := m.Load("iso.txt")
	if string(got2) != "original" {
		t.Error("Load should return isolated copy, not mutable reference")
	}
}

func TestMemoryStorageService_ImplementsInterface(t *testing.T) {
	var _ utils.StorageService = (*utils.MemoryStorageService)(nil)
}

// ── FileStorageService tests ───────────────────────────────────────────────────

func TestFileStorageService_SaveLoad(t *testing.T) {
	dir := t.TempDir()
	fs := utils.NewFileStorageService(dir)

	data := []byte("file content")
	if err := fs.Save("test.frx", data); err != nil {
		t.Fatalf("Save: %v", err)
	}
	got, err := fs.Load("test.frx")
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if string(got) != "file content" {
		t.Errorf("Load = %q", got)
	}
}

func TestFileStorageService_Exists(t *testing.T) {
	dir := t.TempDir()
	fs := utils.NewFileStorageService(dir)

	if fs.Exists("nope.frx") {
		t.Error("Exists on missing file should be false")
	}
	_ = fs.Save("yes.frx", []byte("y"))
	if !fs.Exists("yes.frx") {
		t.Error("Exists after Save should be true")
	}
}

func TestFileStorageService_SubDirectory(t *testing.T) {
	dir := t.TempDir()
	fs := utils.NewFileStorageService(dir)

	if err := fs.Save(filepath.Join("sub", "report.frx"), []byte("sub")); err != nil {
		t.Fatalf("Save in subdir: %v", err)
	}
	got, err := fs.Load(filepath.Join("sub", "report.frx"))
	if err != nil {
		t.Fatalf("Load from subdir: %v", err)
	}
	if string(got) != "sub" {
		t.Errorf("Load = %q", got)
	}
}

func TestFileStorageService_Reader(t *testing.T) {
	dir := t.TempDir()
	fs := utils.NewFileStorageService(dir)
	_ = fs.Save("stream.txt", []byte("stream data"))

	rc, err := fs.Reader("stream.txt")
	if err != nil {
		t.Fatalf("Reader: %v", err)
	}
	defer rc.Close()

	got, _ := io.ReadAll(rc)
	if string(got) != "stream data" {
		t.Errorf("Reader = %q", got)
	}
}

func TestFileStorageService_Writer(t *testing.T) {
	dir := t.TempDir()
	fs := utils.NewFileStorageService(dir)

	wc, err := fs.Writer("out.txt")
	if err != nil {
		t.Fatalf("Writer: %v", err)
	}
	_, _ = wc.Write([]byte("out content"))
	_ = wc.Close()

	got, err := fs.Load("out.txt")
	if err != nil {
		t.Fatalf("Load after Writer: %v", err)
	}
	if string(got) != "out content" {
		t.Errorf("Writer = %q", got)
	}
}

func TestFileStorageService_AbsolutePathIgnoresBase(t *testing.T) {
	dir := t.TempDir()
	fs := utils.NewFileStorageService(dir)

	// Write to an absolute path (different from base dir).
	absPath := filepath.Join(t.TempDir(), "abs.txt")
	if err := fs.Save(absPath, []byte("abs")); err != nil {
		t.Fatalf("Save abs: %v", err)
	}
	got, err := fs.Load(absPath)
	if err != nil {
		t.Fatalf("Load abs: %v", err)
	}
	if string(got) != "abs" {
		t.Errorf("abs = %q", got)
	}
}

func TestFileStorageService_ImplementsInterface(t *testing.T) {
	var _ utils.StorageService = (*utils.FileStorageService)(nil)
}
