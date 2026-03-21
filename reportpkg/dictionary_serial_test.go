package reportpkg

// dictionary_serial_test.go — internal tests for Dictionary FRX serialization.
//
// Tests verify that Report.SaveTo writes a <Dictionary> element containing
// Parameters, Totals, Relations, and data connections, and that a round-trip
// (save → load) preserves the dictionary content.

import (
	"strings"
	"testing"

	"github.com/andrewloable/go-fastreport/data"
)

// ─── hasDictionaryContent ─────────────────────────────────────────────────────

func TestHasDictionaryContent_Empty(t *testing.T) {
	d := data.NewDictionary()
	if hasDictionaryContent(d) {
		t.Error("empty dictionary should not have content")
	}
}

func TestHasDictionaryContent_WithParameter(t *testing.T) {
	d := data.NewDictionary()
	d.AddParameter(&data.Parameter{Name: "P1"})
	if !hasDictionaryContent(d) {
		t.Error("dictionary with a parameter should have content")
	}
}

func TestHasDictionaryContent_WithTotal(t *testing.T) {
	d := data.NewDictionary()
	d.AddTotal(&data.Total{Name: "T1"})
	if !hasDictionaryContent(d) {
		t.Error("dictionary with a total should have content")
	}
}

func TestHasDictionaryContent_WithRelation(t *testing.T) {
	d := data.NewDictionary()
	d.AddRelation(&data.Relation{Name: "R1"})
	if !hasDictionaryContent(d) {
		t.Error("dictionary with a relation should have content")
	}
}

func TestHasDictionaryContent_WithDataSource(t *testing.T) {
	d := data.NewDictionary()
	d.AddDataSource(data.NewBusinessObjectDataSource("DS", nil))
	if !hasDictionaryContent(d) {
		t.Error("dictionary with a data source should have content")
	}
}

// ─── save round-trip: empty dictionary ───────────────────────────────────────

func TestSaveTo_EmptyDictionary_NoDictionaryElement(t *testing.T) {
	// A report with an empty dictionary should NOT emit a <Dictionary/> element.
	r := NewReport()
	xml, err := r.SaveToString()
	if err != nil {
		t.Fatalf("SaveToString: %v", err)
	}
	if strings.Contains(xml, "<Dictionary") {
		t.Errorf("empty dictionary should not produce <Dictionary> element; xml:\n%s", xml)
	}
}

// ─── save: parameters ─────────────────────────────────────────────────────────

func TestSaveTo_DictionaryParameters_Roundtrip(t *testing.T) {
	r := NewReport()
	dict := r.Dictionary()
	p := &data.Parameter{Name: "StartDate", DataType: "System.DateTime", Expression: "[Today]"}
	dict.AddParameter(p)

	// Save to string.
	xml, err := r.SaveToString()
	if err != nil {
		t.Fatalf("SaveToString: %v", err)
	}
	if !strings.Contains(xml, `<Dictionary`) {
		t.Fatal("expected <Dictionary> element in saved XML")
	}
	if !strings.Contains(xml, `Name="StartDate"`) {
		t.Errorf("expected parameter Name=StartDate in saved XML; got:\n%s", xml)
	}
	if !strings.Contains(xml, `DataType="System.DateTime"`) {
		t.Errorf("expected DataType in saved XML; got:\n%s", xml)
	}
	if !strings.Contains(xml, `Expression="[Today]"`) {
		t.Errorf("expected Expression in saved XML; got:\n%s", xml)
	}

	// Reload and verify.
	r2 := NewReport()
	if err := r2.LoadFromString(xml); err != nil {
		t.Fatalf("LoadFromString: %v", err)
	}
	params := r2.Dictionary().Parameters()
	if len(params) == 0 {
		t.Fatal("expected parameter after round-trip load")
	}
	found := false
	for _, q := range params {
		if q.Name == "StartDate" {
			found = true
			if q.DataType != "System.DateTime" {
				t.Errorf("DataType = %q, want System.DateTime", q.DataType)
			}
			if q.Expression != "[Today]" {
				t.Errorf("Expression = %q, want [Today]", q.Expression)
			}
		}
	}
	if !found {
		t.Error("parameter StartDate not found after round-trip")
	}
}

func TestSaveTo_NestedParameters_Roundtrip(t *testing.T) {
	r := NewReport()
	parent := &data.Parameter{Name: "Filter"}
	child := &data.Parameter{Name: "Region", DataType: "System.String"}
	parent.AddParameter(child)
	r.Dictionary().AddParameter(parent)

	xml, err := r.SaveToString()
	if err != nil {
		t.Fatalf("SaveToString: %v", err)
	}

	r2 := NewReport()
	if err := r2.LoadFromString(xml); err != nil {
		t.Fatalf("LoadFromString: %v", err)
	}
	params := r2.Dictionary().Parameters()
	found := false
	for _, p := range params {
		if p.Name == "Filter" {
			found = true
			nested := p.Parameters()
			if len(nested) == 0 {
				t.Error("expected nested parameter Region after round-trip")
			} else if nested[0].Name != "Region" {
				t.Errorf("nested parameter Name = %q, want Region", nested[0].Name)
			}
		}
	}
	if !found {
		t.Error("parameter Filter not found after round-trip")
	}
}

// ─── save: totals ─────────────────────────────────────────────────────────────

func TestSaveTo_DictionaryTotals_Roundtrip(t *testing.T) {
	r := NewReport()
	tot := &data.Total{
		Name:       "SalesTotal",
		Expression: "[Amount]",
		TotalType:  data.TotalTypeSum,
		Evaluator:  "DataBand1",
		PrintOn:    "ReportSummaryBand",
	}
	r.Dictionary().AddTotal(tot)

	xml, err := r.SaveToString()
	if err != nil {
		t.Fatalf("SaveToString: %v", err)
	}
	if !strings.Contains(xml, `Name="SalesTotal"`) {
		t.Errorf("expected SalesTotal in saved XML; got:\n%s", xml)
	}

	r2 := NewReport()
	if err := r2.LoadFromString(xml); err != nil {
		t.Fatalf("LoadFromString: %v", err)
	}
	totals := r2.Dictionary().Totals()
	if len(totals) == 0 {
		t.Fatal("expected total after round-trip load")
	}
	found := false
	for _, q := range totals {
		if q.Name == "SalesTotal" {
			found = true
			if q.Evaluator != "DataBand1" {
				t.Errorf("Evaluator = %q, want DataBand1", q.Evaluator)
			}
			if q.PrintOn != "ReportSummaryBand" {
				t.Errorf("PrintOn = %q, want ReportSummaryBand", q.PrintOn)
			}
		}
	}
	if !found {
		t.Error("total SalesTotal not found after round-trip")
	}
}

func TestSaveTo_TotalType_NonDefault(t *testing.T) {
	// Verify that non-Sum TotalType is serialized and reloaded correctly.
	r := NewReport()
	r.Dictionary().AddTotal(&data.Total{
		Name:      "MinPrice",
		TotalType: data.TotalTypeMin,
		Evaluator: "DataBand1",
	})

	xml, err := r.SaveToString()
	if err != nil {
		t.Fatalf("SaveToString: %v", err)
	}
	if !strings.Contains(xml, `TotalType="Min"`) {
		t.Errorf("expected TotalType=Min in saved XML; got:\n%s", xml)
	}

	r2 := NewReport()
	if err := r2.LoadFromString(xml); err != nil {
		t.Fatalf("LoadFromString: %v", err)
	}
	tots := r2.Dictionary().Totals()
	if len(tots) == 0 {
		t.Fatal("expected total after round-trip")
	}
	if tots[0].TotalType != data.TotalTypeMin {
		t.Errorf("TotalType = %v, want TotalTypeMin", tots[0].TotalType)
	}
}

// ─── save: relations ──────────────────────────────────────────────────────────

func TestSaveTo_DictionaryRelations_Roundtrip(t *testing.T) {
	r := NewReport()
	rel := &data.Relation{
		Name:             "Orders_Details",
		ParentSourceName: "Orders",
		ChildSourceName:  "OrderDetails",
		ParentColumnNames: []string{"OrderID"},
		ChildColumnNames:  []string{"OrderID"},
	}
	r.Dictionary().AddRelation(rel)

	xml, err := r.SaveToString()
	if err != nil {
		t.Fatalf("SaveToString: %v", err)
	}
	if !strings.Contains(xml, `Name="Orders_Details"`) {
		t.Errorf("expected relation name in saved XML; got:\n%s", xml)
	}

	r2 := NewReport()
	if err := r2.LoadFromString(xml); err != nil {
		t.Fatalf("LoadFromString: %v", err)
	}
	rels := r2.Dictionary().Relations()
	if len(rels) == 0 {
		t.Fatal("expected relation after round-trip load")
	}
	if rels[0].Name != "Orders_Details" {
		t.Errorf("Relation.Name = %q, want Orders_Details", rels[0].Name)
	}
	if rels[0].ParentSourceName != "Orders" {
		t.Errorf("ParentSourceName = %q, want Orders", rels[0].ParentSourceName)
	}
	if rels[0].ChildSourceName != "OrderDetails" {
		t.Errorf("ChildSourceName = %q, want OrderDetails", rels[0].ChildSourceName)
	}
	if len(rels[0].ParentColumnNames) == 0 || rels[0].ParentColumnNames[0] != "OrderID" {
		t.Errorf("ParentColumnNames = %v, want [OrderID]", rels[0].ParentColumnNames)
	}
}

// ─── save: business object data source ───────────────────────────────────────

func TestSaveTo_BusinessObjectDataSource_Roundtrip(t *testing.T) {
	r := NewReport()
	ds := data.NewBusinessObjectDataSource("Customers", nil)
	ds.SetAlias("Customers")
	r.Dictionary().AddDataSource(ds)

	xml, err := r.SaveToString()
	if err != nil {
		t.Fatalf("SaveToString: %v", err)
	}
	if !strings.Contains(xml, `<BusinessObjectDataSource`) {
		t.Errorf("expected <BusinessObjectDataSource> in saved XML; got:\n%s", xml)
	}
	if !strings.Contains(xml, `Name="Customers"`) {
		t.Errorf("expected Name=Customers in saved XML; got:\n%s", xml)
	}

	r2 := NewReport()
	if err := r2.LoadFromString(xml); err != nil {
		t.Fatalf("LoadFromString: %v", err)
	}
	sources := r2.Dictionary().DataSources()
	if len(sources) == 0 {
		t.Fatal("expected data source after round-trip load")
	}
	if sources[0].Name() != "Customers" {
		t.Errorf("DataSource.Name = %q, want Customers", sources[0].Name())
	}
}

// ─── totalTypeToString ────────────────────────────────────────────────────────

func TestTotalTypeToString_AllValues(t *testing.T) {
	cases := []struct {
		tt   data.TotalType
		want string
	}{
		{data.TotalTypeSum, "Sum"},
		{data.TotalTypeMin, "Min"},
		{data.TotalTypeMax, "Max"},
		{data.TotalTypeAvg, "Avg"},
		{data.TotalTypeCount, "Count"},
		{data.TotalTypeCountDistinct, "CountDistinct"},
	}
	for _, c := range cases {
		got := totalTypeToString(c.tt)
		if got != c.want {
			t.Errorf("totalTypeToString(%v) = %q, want %q", c.tt, got, c.want)
		}
	}
}

func TestTotalTypeToString_Unknown(t *testing.T) {
	got := totalTypeToString(data.TotalType(99))
	if got == "" {
		t.Error("unknown TotalType should produce non-empty string")
	}
}

// ─── dictionarySerializer: Sum TotalType not written (default) ───────────────

func TestSaveTo_TotalType_SumNotWritten(t *testing.T) {
	// TotalType "Sum" is the default and should not be written to the FRX.
	r := NewReport()
	r.Dictionary().AddTotal(&data.Total{Name: "T", TotalType: data.TotalTypeSum})

	xml, err := r.SaveToString()
	if err != nil {
		t.Fatalf("SaveToString: %v", err)
	}
	// TotalType attribute should not be present for Sum (it is the default).
	if strings.Contains(xml, `TotalType="Sum"`) {
		t.Errorf("TotalType=Sum should not be written; got:\n%s", xml)
	}
}
