package utils

import "sync"

// Config holds global configuration settings for the go-fastreport library.
// Access via the package-level DefaultConfig variable.
type Config struct {
	mu sync.RWMutex

	// PreparedCompressed enables compression in prepared report files (fpx).
	PreparedCompressed bool

	// ForbidLocalData prevents local file paths in XML/CSV data sources.
	ForbidLocalData bool

	// TempFolder overrides the system temporary directory.
	TempFolder string

	// RightToLeft enables right-to-left text direction.
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

// SetTempFolder sets the temporary folder path.
func (c *Config) SetTempFolder(path string) {
	c.mu.Lock()
	c.TempFolder = path
	c.mu.Unlock()
}

// GetTempFolder returns the configured temporary folder path.
func (c *Config) GetTempFolder() string {
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
