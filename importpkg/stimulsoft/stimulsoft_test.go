package stimulsoft_test

import (
	"strings"
	"testing"

	"github.com/andrewloable/go-fastreport/importpkg/stimulsoft"
	"github.com/andrewloable/go-fastreport/reportpkg"
)

func TestNew_Name(t *testing.T) {
	c := stimulsoft.New()
	if c.Name() != "StimulSoft Importer" {
		t.Fatalf("expected name %q, got %q", "StimulSoft Importer", c.Name())
	}
}

func TestImport_NotImplemented(t *testing.T) {
	c := stimulsoft.New()
	rpt, err := c.Import(strings.NewReader("<report/>"))
	if err == nil {
		t.Fatal("expected error from Import, got nil")
	}
	if rpt != nil {
		t.Fatal("expected nil report from unimplemented Import")
	}
}

func TestLoadReportFromFile_NotImplemented(t *testing.T) {
	c := stimulsoft.New()
	rpt := reportpkg.NewReport()
	err := c.LoadReportFromFile(rpt, "report.mrt")
	if err == nil {
		t.Fatal("expected error from LoadReportFromFile, got nil")
	}
}

func TestLoadReportFromStream_NotImplemented(t *testing.T) {
	c := stimulsoft.New()
	rpt := reportpkg.NewReport()
	err := c.LoadReportFromStream(rpt, strings.NewReader("<report/>"))
	if err == nil {
		t.Fatal("expected error from LoadReportFromStream, got nil")
	}
}
