package export_test

// tempfile_coverage_test.go — tests for AddGeneratedFile, CreateTempFile,
// and DeleteTempFiles on ExportBase (all were 0% covered).
//
// C# ref: FastReport.Base/Export/ExportBase.cs

import (
	"os"
	"testing"

	"github.com/andrewloable/go-fastreport/export"
)

// ── AddGeneratedFile / GeneratedFiles ────────────────────────────────────────

func TestExportBase_AddGeneratedFile(t *testing.T) {
	base := export.NewExportBase()
	base.AddGeneratedFile("/tmp/report.html")
	files := base.GeneratedFiles()
	if len(files) != 1 {
		t.Fatalf("GeneratedFiles len = %d, want 1", len(files))
	}
	if files[0] != "/tmp/report.html" {
		t.Errorf("GeneratedFiles[0] = %q, want %q", files[0], "/tmp/report.html")
	}
}

func TestExportBase_AddGeneratedFile_Multiple(t *testing.T) {
	base := export.NewExportBase()
	base.AddGeneratedFile("a.html")
	base.AddGeneratedFile("b.html")
	base.AddGeneratedFile("c.html")
	files := base.GeneratedFiles()
	if len(files) != 3 {
		t.Errorf("GeneratedFiles len = %d, want 3", len(files))
	}
}

// ── CreateTempFile ────────────────────────────────────────────────────────────

func TestExportBase_CreateTempFile_CreatesFile(t *testing.T) {
	base := export.NewExportBase()
	f, err := base.CreateTempFile()
	if err != nil {
		t.Fatalf("CreateTempFile: %v", err)
	}
	if f == nil {
		t.Fatal("CreateTempFile returned nil file")
	}
	// The file should exist on disk.
	if _, statErr := os.Stat(f.Name()); statErr != nil {
		t.Errorf("temp file does not exist: %v", statErr)
	}
	// Clean up.
	base.DeleteTempFiles()
	// File should be removed after DeleteTempFiles.
	if _, statErr := os.Stat(f.Name()); !os.IsNotExist(statErr) {
		t.Errorf("temp file still exists after DeleteTempFiles: %v", f.Name())
	}
}

func TestExportBase_CreateTempFile_MultipleFiles(t *testing.T) {
	base := export.NewExportBase()
	f1, err := base.CreateTempFile()
	if err != nil {
		t.Fatalf("CreateTempFile 1: %v", err)
	}
	f2, err := base.CreateTempFile()
	if err != nil {
		t.Fatalf("CreateTempFile 2: %v", err)
	}
	if f1.Name() == f2.Name() {
		t.Error("two CreateTempFile calls should produce different file names")
	}
	base.DeleteTempFiles()
	for _, name := range []string{f1.Name(), f2.Name()} {
		if _, statErr := os.Stat(name); !os.IsNotExist(statErr) {
			t.Errorf("temp file %q still exists after DeleteTempFiles", name)
		}
	}
}

func TestExportBase_DeleteTempFiles_Empty(t *testing.T) {
	base := export.NewExportBase()
	base.DeleteTempFiles() // must not panic when no temp files exist
}

func TestExportBase_DeleteTempFiles_Idempotent(t *testing.T) {
	base := export.NewExportBase()
	_, err := base.CreateTempFile()
	if err != nil {
		t.Fatalf("CreateTempFile: %v", err)
	}
	base.DeleteTempFiles()
	base.DeleteTempFiles() // second call must not panic
}
