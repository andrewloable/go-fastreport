package reportpkg_test

// Smoke tests for Map and DigitalSignature FRX reports.
// MapObject and DigitalSignatureObject are registered in the serial registry.
// These tests verify that the FRX files load without panic and contain
// the expected object types.

import (
	"testing"

	"github.com/andrewloable/go-fastreport/object"
)

func TestFRXSmoke_USAMap(t *testing.T) {
	r := loadFRXSmoke(t, "The USA Map.frx")
	if n := countObjectsOfType[*object.MapObject](r); n == 0 {
		t.Error("expected at least one MapObject in The USA Map.frx")
	}
}

func TestFRXSmoke_PurchaseOrderDigitalSignature(t *testing.T) {
	r := loadFRXSmoke(t, "Purchase Order.frx")
	if n := countObjectsOfType[*object.DigitalSignatureObject](r); n == 0 {
		t.Error("expected at least one DigitalSignatureObject in Purchase Order.frx")
	}
}
