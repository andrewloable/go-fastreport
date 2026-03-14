package reportpkg_test

// Smoke tests for the "Simple/Basic" category of FastReport sample FRX files.
// Each test verifies that the file loads without error and has at least one page.

import (
	"testing"

	"github.com/andrewloable/go-fastreport/reportpkg"
)

// loadFRXSmoke loads an FRX file from test-reports/ and verifies basic structure.
// It returns the loaded report or calls t.Fatal on failure.
func loadFRXSmoke(t *testing.T, filename string) *reportpkg.Report {
	t.Helper()
	path := testReportsDir() + "/" + filename
	r := reportpkg.NewReport()
	if err := r.Load(path); err != nil {
		t.Fatalf("Load(%q): %v", filename, err)
	}
	if len(r.Pages()) == 0 {
		t.Fatalf("Load(%q): no pages deserialized", filename)
	}
	return r
}

func TestFRXSmoke_SimpleList(t *testing.T) {
	r := loadFRXSmoke(t, "Simple List.frx")
	if len(r.Pages()) < 1 {
		t.Errorf("expected at least 1 page")
	}
}

func TestFRXSmoke_HelloFastReport(t *testing.T) {
	r := loadFRXSmoke(t, "Hello, FastReport!.frx")
	if len(r.Pages()) < 1 {
		t.Errorf("expected at least 1 page")
	}
}

func TestFRXSmoke_Text(t *testing.T) {
	r := loadFRXSmoke(t, "Text.frx")
	if len(r.Pages()) < 1 {
		t.Errorf("expected at least 1 page")
	}
}

func TestFRXSmoke_MailMerge(t *testing.T) {
	r := loadFRXSmoke(t, "Mail Merge.frx")
	if len(r.Pages()) < 1 {
		t.Errorf("expected at least 1 page")
	}
}

func TestFRXSmoke_Unicode(t *testing.T) {
	r := loadFRXSmoke(t, "Unicode.frx")
	if len(r.Pages()) < 1 {
		t.Errorf("expected at least 1 page")
	}
}

func TestFRXSmoke_BusinessObjects(t *testing.T) {
	r := loadFRXSmoke(t, "Business Objects.frx")
	if len(r.Pages()) < 1 {
		t.Errorf("expected at least 1 page")
	}
}

func TestFRXSmoke_PrintDataTable(t *testing.T) {
	r := loadFRXSmoke(t, "Print DataTable.frx")
	if len(r.Pages()) < 1 {
		t.Errorf("expected at least 1 page")
	}
}
