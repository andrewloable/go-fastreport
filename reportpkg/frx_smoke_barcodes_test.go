package reportpkg_test

// Smoke tests for barcode FRX reports.

import (
	"testing"

	"github.com/andrewloable/go-fastreport/object"
)

func TestFRXSmoke_Barcode(t *testing.T) {
	r := loadFRXSmoke(t, "Barcode.frx")
	n := countObjectsOfType[*object.BarcodeObject](r)
	if n == 0 {
		t.Error("expected at least one BarcodeObject in Barcode.frx")
	}
}

func TestFRXSmoke_QRCodes(t *testing.T) {
	r := loadFRXSmoke(t, "QR-Codes.frx")
	n := countObjectsOfType[*object.BarcodeObject](r)
	if n == 0 {
		t.Error("expected at least one BarcodeObject in QR-Codes.frx")
	}
}

func TestFRXSmoke_Pharmacode(t *testing.T) {
	// BarcodeObjects are nested inside TableObjects; just verify the file loads.
	loadFRXSmoke(t, "Pharmacode.frx")
}

func TestFRXSmoke_ZipCode(t *testing.T) {
	r := loadFRXSmoke(t, "ZipCode.frx")
	n := countObjectsOfType[*object.ZipCodeObject](r)
	if n == 0 {
		t.Error("expected at least one ZipCodeObject in ZipCode.frx")
	}
}
