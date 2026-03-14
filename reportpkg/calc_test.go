package reportpkg_test

import (
	"testing"

	"github.com/andrewloable/go-fastreport/data"
	"github.com/andrewloable/go-fastreport/reportpkg"
)

func TestReport_Calc_Parameter(t *testing.T) {
	r := reportpkg.NewReport()
	param := &data.Parameter{Name: "Greeting", Value: "Hello"}
	r.Dictionary().AddParameter(param)

	val, err := r.Calc("[Greeting]")
	if err != nil {
		t.Fatalf("Calc: %v", err)
	}
	if val != "Hello" {
		t.Errorf("got %v, want Hello", val)
	}
}

func TestReport_Calc_SystemVariable(t *testing.T) {
	r := reportpkg.NewReport()
	r.Dictionary().SetSystemVariable("PageNumber", 3)

	val, err := r.Calc("[PageNumber]")
	if err != nil {
		t.Fatalf("Calc: %v", err)
	}
	if val != 3 {
		t.Errorf("got %v, want 3", val)
	}
}

func TestReport_Calc_ArithmeticExpression(t *testing.T) {
	r := reportpkg.NewReport()
	param := &data.Parameter{Name: "Price", Value: 10.0}
	r.Dictionary().AddParameter(param)

	val, err := r.Calc("[Price] * 2")
	if err != nil {
		t.Fatalf("Calc: %v", err)
	}
	if val != 20.0 {
		t.Errorf("got %v, want 20.0", val)
	}
}

func TestReport_CalcText_Template(t *testing.T) {
	r := reportpkg.NewReport()
	param := &data.Parameter{Name: "Name", Value: "World"}
	r.Dictionary().AddParameter(param)

	text, err := r.CalcText("Hello [Name]!")
	if err != nil {
		t.Fatalf("CalcText: %v", err)
	}
	if text != "Hello World!" {
		t.Errorf("got %q, want %q", text, "Hello World!")
	}
}

func TestReport_CalcText_NoExpressions(t *testing.T) {
	r := reportpkg.NewReport()
	text, err := r.CalcText("Static text")
	if err != nil {
		t.Fatalf("CalcText: %v", err)
	}
	if text != "Static text" {
		t.Errorf("got %q, want %q", text, "Static text")
	}
}

func TestReport_Calc_DataSourceColumn(t *testing.T) {
	r := reportpkg.NewReport()

	ds := data.NewBaseDataSource("Customers")
	ds.SetAlias("Customers")
	ds.AddColumn(data.Column{Name: "Name"})
	ds.AddRow(map[string]any{"Name": "Alice"})
	_ = ds.Init()
	_ = ds.First()

	r.SetCalcContext(ds)

	val, err := r.Calc("[Name]")
	if err != nil {
		t.Fatalf("Calc: %v", err)
	}
	if val != "Alice" {
		t.Errorf("got %v, want Alice", val)
	}
}
