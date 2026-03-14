package utils_test

import (
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
	if cfg.GetTempFolder() != "" {
		t.Error("TempFolder should default to empty")
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
