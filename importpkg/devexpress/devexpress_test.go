package devexpress_test

import (
	"strings"
	"testing"

	"github.com/andrewloable/go-fastreport/importpkg/devexpress"
	"github.com/andrewloable/go-fastreport/reportpkg"
)

func TestNew_Name(t *testing.T) {
	c := devexpress.New()
	if c.Name() != "DevExpress Importer" {
		t.Fatalf("expected name %q, got %q", "DevExpress Importer", c.Name())
	}
}

func TestImport_NotImplemented(t *testing.T) {
	c := devexpress.New()
	rpt, err := c.Import(strings.NewReader("<XtraReportsLayoutSerializer/>"))
	if err == nil {
		t.Fatal("expected error from Import, got nil")
	}
	if rpt != nil {
		t.Fatal("expected nil report from unimplemented Import")
	}
}

func TestLoadReportFromFile_NotImplemented(t *testing.T) {
	c := devexpress.New()
	rpt := reportpkg.NewReport()
	err := c.LoadReportFromFile(rpt, "report.repx")
	if err == nil {
		t.Fatal("expected error from LoadReportFromFile, got nil")
	}
}

func TestLoadReportFromStream_NotImplemented(t *testing.T) {
	c := devexpress.New()
	rpt := reportpkg.NewReport()
	err := c.LoadReportFromStream(rpt, strings.NewReader("<XtraReportsLayoutSerializer/>"))
	if err == nil {
		t.Fatal("expected error from LoadReportFromStream, got nil")
	}
}
