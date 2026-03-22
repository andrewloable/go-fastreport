package utils_test

import (
	"os"
	"strings"
	"testing"

	"github.com/andrewloable/go-fastreport/utils"
)

func TestConfigDefaults(t *testing.T) {
	cfg := &utils.Config{PreparedCompressed: true}
	if !cfg.GetPreparedCompressed() {
		t.Error("PreparedCompressed should default to true")
	}
	if cfg.GetForbidLocalData() {
		t.Error("ForbidLocalData should default to false")
	}
	// GetTempFolder() falls back to os.TempDir() when TempFolder is unset.
	// Mirrors C# Config.GetTempFolder() which returns GetTempPath() when TempFolder==null
	// (Config.cs lines 291-293).
	if cfg.GetTempFolder() != os.TempDir() {
		t.Errorf("GetTempFolder() = %q, want os.TempDir() = %q", cfg.GetTempFolder(), os.TempDir())
	}
	// GetConfiguredTempFolder() exposes the raw field (empty when unset).
	if cfg.GetConfiguredTempFolder() != "" {
		t.Error("GetConfiguredTempFolder should return empty when TempFolder not set")
	}
	if cfg.GetRightToLeft() {
		t.Error("RightToLeft should default to false")
	}
}

func TestConfigSetGet(t *testing.T) {
	cfg := &utils.Config{}

	cfg.SetPreparedCompressed(true)
	if !cfg.GetPreparedCompressed() {
		t.Error("PreparedCompressed not set")
	}
	cfg.SetPreparedCompressed(false)
	if cfg.GetPreparedCompressed() {
		t.Error("PreparedCompressed not cleared")
	}

	cfg.SetForbidLocalData(true)
	if !cfg.GetForbidLocalData() {
		t.Error("ForbidLocalData not set")
	}

	cfg.SetTempFolder("/tmp/test")
	if cfg.GetTempFolder() != "/tmp/test" {
		t.Errorf("TempFolder = %q, want /tmp/test", cfg.GetTempFolder())
	}

	cfg.SetRightToLeft(true)
	if !cfg.GetRightToLeft() {
		t.Error("RightToLeft not set")
	}
}

func TestDefaultConfigExists(t *testing.T) {
	if utils.DefaultConfig == nil {
		t.Fatal("DefaultConfig should not be nil")
	}
	if !utils.DefaultConfig.GetPreparedCompressed() {
		t.Error("DefaultConfig.PreparedCompressed should be true")
	}
}

func TestConfigConcurrency(t *testing.T) {
	cfg := &utils.Config{}
	done := make(chan struct{})
	for i := 0; i < 10; i++ {
		go func(i int) {
			cfg.SetPreparedCompressed(i%2 == 0)
			cfg.GetPreparedCompressed()
			cfg.SetForbidLocalData(i%3 == 0)
			cfg.GetForbidLocalData()
			done <- struct{}{}
		}(i)
	}
	for i := 0; i < 10; i++ {
		<-done
	}
}

func TestVersion(t *testing.T) {
	// Version must be a non-empty semver-like string.
	// Mirrors C# Config.Version (Config.cs line 199).
	if utils.Version == "" {
		t.Error("Version should not be empty")
	}
	if !strings.Contains(utils.Version, ".") {
		t.Errorf("Version %q does not look like a semver string", utils.Version)
	}
}

func TestGetTempFolderFallback(t *testing.T) {
	// GetTempFolder() on an unconfigured Config returns os.TempDir().
	// Mirrors C# Config.GetTempFolder(): return TempFolder == null ? GetTempPath() : TempFolder.
	cfg := &utils.Config{}
	if got := cfg.GetTempFolder(); got != os.TempDir() {
		t.Errorf("GetTempFolder() = %q, want os.TempDir() = %q", got, os.TempDir())
	}

	cfg.SetTempFolder("/custom/tmp")
	if got := cfg.GetTempFolder(); got != "/custom/tmp" {
		t.Errorf("GetTempFolder() after set = %q, want /custom/tmp", got)
	}
	if got := cfg.GetConfiguredTempFolder(); got != "/custom/tmp" {
		t.Errorf("GetConfiguredTempFolder() = %q, want /custom/tmp", got)
	}

	// Reset to empty reverts to os.TempDir() fallback.
	cfg.SetTempFolder("")
	if got := cfg.GetTempFolder(); got != os.TempDir() {
		t.Errorf("GetTempFolder() after reset = %q, want os.TempDir()", got)
	}
	if got := cfg.GetConfiguredTempFolder(); got != "" {
		t.Errorf("GetConfiguredTempFolder() after reset = %q, want empty", got)
	}
}

func TestGetEffectiveTempFolder(t *testing.T) {
	// Package-level helper delegates to DefaultConfig.GetTempFolder().
	if got := utils.GetEffectiveTempFolder(); got == "" {
		t.Error("GetEffectiveTempFolder() should not return empty string")
	}
}

func TestCreateTempFile_NoDir(t *testing.T) {
	// CreateTempFile("") creates a file in the effective temp dir.
	// Mirrors C# Config.CreateTempFile("") -> GetTempFileName() (Config.cs lines 284-289, 411-414).
	cfg := &utils.Config{}
	path, err := cfg.CreateTempFile("")
	if err != nil {
		t.Fatalf("CreateTempFile(\"\") error: %v", err)
	}
	defer os.Remove(path)
	if _, statErr := os.Stat(path); os.IsNotExist(statErr) {
		t.Errorf("CreateTempFile path %q does not exist", path)
	}
}

func TestCreateTempFile_WithDir(t *testing.T) {
	// CreateTempFile(dir) creates a file inside the given directory.
	// Mirrors C# Config.CreateTempFile(dir) -> Path.Combine(dir, Path.GetRandomFileName()).
	dir := t.TempDir()
	cfg := &utils.Config{}
	path, err := cfg.CreateTempFile(dir)
	if err != nil {
		t.Fatalf("CreateTempFile(%q) error: %v", dir, err)
	}
	defer os.Remove(path)
	if !strings.HasPrefix(path, dir) {
		t.Errorf("CreateTempFile path %q is not inside dir %q", path, dir)
	}
	if _, statErr := os.Stat(path); os.IsNotExist(statErr) {
		t.Errorf("CreateTempFile path %q does not exist", path)
	}
}

func TestTempFilePath(t *testing.T) {
	// TempFilePath() creates a timestamped temp file and returns its path.
	// Mirrors C# Config.GetTempFileName() (Config.cs lines 411-414).
	cfg := &utils.Config{}
	path, err := cfg.TempFilePath()
	if err != nil {
		t.Fatalf("TempFilePath() error: %v", err)
	}
	defer os.Remove(path)
	if path == "" {
		t.Error("TempFilePath() returned empty string")
	}
	if _, statErr := os.Stat(path); os.IsNotExist(statErr) {
		t.Errorf("TempFilePath path %q does not exist", path)
	}
}

func TestCreateTempFileInDir(t *testing.T) {
	// Package-level helper delegates to DefaultConfig.CreateTempFile(dir).
	dir := t.TempDir()
	path, err := utils.CreateTempFileInDir(dir)
	if err != nil {
		t.Fatalf("CreateTempFileInDir(%q) error: %v", dir, err)
	}
	defer os.Remove(path)
	if !strings.HasPrefix(path, dir) {
		t.Errorf("CreateTempFileInDir path %q is not inside dir %q", path, dir)
	}
}
