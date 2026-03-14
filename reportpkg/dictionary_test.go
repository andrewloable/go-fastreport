package reportpkg_test

// Tests for FRX Dictionary deserialization (Parameters, Relations, Totals).

import (
	"testing"
)

func TestDictionary_Parameters(t *testing.T) {
	// Dialog Elements.frx has one Parameter named "Parameter" with an Expression.
	r := loadFRXSmoke(t, "Dialog Elements.frx")
	dict := r.Dictionary()
	if len(dict.Parameters()) == 0 {
		t.Fatal("expected at least one parameter after loading Dialog Elements.frx")
	}
	var found bool
	for _, p := range dict.Parameters() {
		if p.Name == "Parameter" {
			found = true
			if p.DataType == "" {
				t.Error("parameter DataType should be non-empty")
			}
			break
		}
	}
	if !found {
		t.Error("parameter named 'Parameter' not found in dictionary")
	}
}

func TestDictionary_Relations(t *testing.T) {
	// Odd-Even Pages, Mirror Margins.frx has a Relation.
	r := loadFRXSmoke(t, "Odd-Even Pages, Mirror Margins.frx")
	dict := r.Dictionary()
	if len(dict.Relations()) == 0 {
		t.Fatal("expected at least one relation after loading the FRX file")
	}
	rel := dict.Relations()[0]
	if rel.Name == "" {
		t.Error("relation Name should be non-empty")
	}
	if rel.ParentSourceName == "" {
		t.Error("relation ParentSourceName should be non-empty")
	}
	if rel.ChildSourceName == "" {
		t.Error("relation ChildSourceName should be non-empty")
	}
	if len(rel.ParentColumnNames) == 0 {
		t.Error("relation ParentColumnNames should be non-empty")
	}
}

func TestDictionary_Totals(t *testing.T) {
	// Start New Page, Reset Page Numbers.frx has a Total.
	r := loadFRXSmoke(t, "Start New Page, Reset Page Numbers.frx")
	dict := r.Dictionary()
	if len(dict.Totals()) == 0 {
		t.Fatal("expected at least one total after loading the FRX file")
	}
	tot := dict.Totals()[0]
	if tot.Name == "" {
		t.Error("total Name should be non-empty")
	}
}

func TestDictionary_EmptyDictionary(t *testing.T) {
	// SVG.frx has an empty <Dictionary/> element — should load cleanly.
	r := loadFRXSmoke(t, "SVG.frx")
	dict := r.Dictionary()
	if dict == nil {
		t.Error("dictionary should never be nil")
	}
}
