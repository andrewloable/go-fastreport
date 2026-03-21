package jasperreports_test

import (
	"strings"
	"testing"

	"github.com/andrewloable/go-fastreport/importpkg/jasperreports"
	"github.com/andrewloable/go-fastreport/reportpkg"
)

func TestNew_Name(t *testing.T) {
	c := jasperreports.New()
	if c.Name() != "JasperReports Importer" {
		t.Fatalf("expected name %q, got %q", "JasperReports Importer", c.Name())
	}
}

func TestImport_NotImplemented(t *testing.T) {
	c := jasperreports.New()
	rpt, err := c.Import(strings.NewReader("<jasperReport/>"))
	if err == nil {
		t.Fatal("expected error from Import, got nil")
	}
	if rpt != nil {
		t.Fatal("expected nil report from unimplemented Import")
	}
}

func TestLoadReportFromFile_NotImplemented(t *testing.T) {
	c := jasperreports.New()
	rpt := reportpkg.NewReport()
	err := c.LoadReportFromFile(rpt, "report.jrxml")
	if err == nil {
		t.Fatal("expected error from LoadReportFromFile, got nil")
	}
}

func TestLoadReportFromStream_NotImplemented(t *testing.T) {
	c := jasperreports.New()
	rpt := reportpkg.NewReport()
	err := c.LoadReportFromStream(rpt, strings.NewReader("<jasperReport/>"))
	if err == nil {
		t.Fatal("expected error from LoadReportFromStream, got nil")
	}
}
