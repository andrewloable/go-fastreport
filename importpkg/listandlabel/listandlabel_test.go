package listandlabel_test

import (
	"strings"
	"testing"

	"github.com/andrewloable/go-fastreport/importpkg/listandlabel"
	"github.com/andrewloable/go-fastreport/reportpkg"
)

func TestNew_Name(t *testing.T) {
	c := listandlabel.New()
	if c.Name() != "List & Label Importer" {
		t.Fatalf("expected name %q, got %q", "List & Label Importer", c.Name())
	}
}

func TestImport_NotImplemented(t *testing.T) {
	c := listandlabel.New()
	rpt, err := c.Import(strings.NewReader("<report/>"))
	if err == nil {
		t.Fatal("expected error from Import, got nil")
	}
	if rpt != nil {
		t.Fatal("expected nil report from unimplemented Import")
	}
}

func TestLoadReportFromFile_NotImplemented(t *testing.T) {
	c := listandlabel.New()
	rpt := reportpkg.NewReport()
	err := c.LoadReportFromFile(rpt, "report.lst")
	if err == nil {
		t.Fatal("expected error from LoadReportFromFile, got nil")
	}
}

func TestLoadReportFromStream_NotImplemented(t *testing.T) {
	c := listandlabel.New()
	rpt := reportpkg.NewReport()
	err := c.LoadReportFromStream(rpt, strings.NewReader("<report/>"))
	if err == nil {
		t.Fatal("expected error from LoadReportFromStream, got nil")
	}
}
