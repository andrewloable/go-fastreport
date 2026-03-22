package utils

import (
	"os"
	"sync"
	"time"
)

// Version is the go-fastreport library version.
// Mirrors C# Config.Version (Config.cs line 199), which returns the assembly version string.
const Version = "1.0.0"

// Config holds global configuration settings for the go-fastreport library.
// Access via the package-level DefaultConfig variable.
// Ported from C# FastReport.Utils.Config (Config.cs) and its Core/OpenSource partial files.
type Config struct {
	mu sync.RWMutex

	// PreparedCompressed enables compression in prepared report files (fpx).
	// Mirrors C# Config.PreparedCompressed (Config.cs line 99-103). Default: true.
	PreparedCompressed bool

	// ForbidLocalData prevents local file paths in XML/CSV data sources.
	// Mirrors C# Config.ForbidLocalData (Config.cs line 80-84).
	ForbidLocalData bool

	// TempFolder overrides the system temporary directory.
	// When empty, GetTempFolder() returns os.TempDir().
	// Mirrors C# Config.TempFolder (Config.cs line 181-185) where null means use system temp.
	TempFolder string

	// RightToLeft enables right-to-left text direction.
	// Mirrors C# Config.RightToLeft (Config.cs line 161-165).
	RightToLeft bool
}

// DefaultConfig is the global configuration instance.
var DefaultConfig = &Config{
	PreparedCompressed: true,
}

// SetPreparedCompressed sets whether prepared reports are compressed.
func (c *Config) SetPreparedCompressed(v bool) {
	c.mu.Lock()
	c.PreparedCompressed = v
	c.mu.Unlock()
}

// GetPreparedCompressed returns whether prepared reports are compressed.
func (c *Config) GetPreparedCompressed() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.PreparedCompressed
}

// SetForbidLocalData sets whether local data file paths are forbidden.
func (c *Config) SetForbidLocalData(v bool) {
	c.mu.Lock()
	c.ForbidLocalData = v
	c.mu.Unlock()
}

// GetForbidLocalData returns whether local data file paths are forbidden.
func (c *Config) GetForbidLocalData() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.ForbidLocalData
}

// SetTempFolder sets the temporary folder path override.
// Pass an empty string to revert to the system temp directory.
func (c *Config) SetTempFolder(path string) {
	c.mu.Lock()
	c.TempFolder = path
	c.mu.Unlock()
}

// GetTempFolder returns the effective temporary folder path.
// When TempFolder has not been set (empty string), it returns os.TempDir().
// Mirrors C# Config.GetTempFolder() (Config.cs lines 291-293):
//
//	return TempFolder == null ? GetTempPath() : TempFolder;
func (c *Config) GetTempFolder() string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if c.TempFolder == "" {
		return os.TempDir()
	}
	return c.TempFolder
}

// GetConfiguredTempFolder returns the raw TempFolder field without the
// os.TempDir() fallback. An empty string means "use the system temp dir".
// Use GetTempFolder() to get the effective path.
func (c *Config) GetConfiguredTempFolder() string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.TempFolder
}

// SetRightToLeft sets whether right-to-left mode is enabled.
func (c *Config) SetRightToLeft(v bool) {
	c.mu.Lock()
	c.RightToLeft = v
	c.mu.Unlock()
}

// GetRightToLeft returns whether right-to-left mode is enabled.
func (c *Config) GetRightToLeft() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.RightToLeft
}

// CreateTempFile creates a temporary file and returns its path.
// When dir is empty, the file is created in GetTempFolder().
// Mirrors C# Config.CreateTempFile(string dir) (Config.cs lines 284-289):
//
//	if (String.IsNullOrEmpty(dir)) return GetTempFileName();
//	return Path.Combine(dir, Path.GetRandomFileName());
func (c *Config) CreateTempFile(dir string) (string, error) {
	if dir == "" {
		return c.TempFilePath()
	}
	f, err := os.CreateTemp(dir, "fr-*")
	if err != nil {
		return "", err
	}
	name := f.Name()
	f.Close()
	return name, nil
}

// TempFilePath creates a unique temporary file inside GetTempFolder() and
// returns its path. The name embeds a timestamp for human readability,
// matching C# Config.GetTempFileName() (Config.cs lines 411-414):
//
//	return Path.Combine(GetTempFolder(),
//	    SystemFake.DateTime.Now.ToString("yyyy-dd-M--HH-mm-ss-") + Path.GetRandomFileName());
func (c *Config) TempFilePath() (string, error) {
	stamp := time.Now().Format("2006-02-1--15-04-05-")
	f, err := os.CreateTemp(c.GetTempFolder(), "fr-"+stamp+"*")
	if err != nil {
		return "", err
	}
	name := f.Name()
	f.Close()
	return name, nil
}

// CreateTempFileInDir creates a temporary file in the specified directory
// using DefaultConfig. Convenience wrapper around DefaultConfig.CreateTempFile.
func CreateTempFileInDir(dir string) (string, error) {
	return DefaultConfig.CreateTempFile(dir)
}

// GetEffectiveTempFolder returns the effective temp folder from DefaultConfig.
// Convenience top-level function mirroring C# Config.GetTempFolder().
func GetEffectiveTempFolder() string {
	return DefaultConfig.GetTempFolder()
}
