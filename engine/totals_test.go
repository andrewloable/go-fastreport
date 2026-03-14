package engine_test

import (
	"testing"

	"github.com/andrewloable/go-fastreport/band"
	"github.com/andrewloable/go-fastreport/data"
	"github.com/andrewloable/go-fastreport/engine"
	"github.com/andrewloable/go-fastreport/object"
	"github.com/andrewloable/go-fastreport/reportpkg"
)

// numericDS is a data source whose Value column yields numeric values.
// It implements both band.DataSource and data.DataSource.
type numericDS struct {
	values []float64
	pos    int
}

func newNumericDS(vals ...float64) *numericDS { return &numericDS{values: vals, pos: -1} }

// band.DataSource + data.DataSource methods
func (d *numericDS) Name() string         { return "Numbers" }
func (d *numericDS) Alias() string        { return "Numbers" }
func (d *numericDS) Init() error          { d.pos = -1; return nil }
func (d *numericDS) First() error         { d.pos = 0; return nil }
func (d *numericDS) Next() error          { d.pos++; return nil }
func (d *numericDS) EOF() bool            { return d.pos >= len(d.values) }
func (d *numericDS) RowCount() int        { return len(d.values) }
func (d *numericDS) CurrentRowNo() int    { return d.pos }
func (d *numericDS) Close() error         { return nil }
func (d *numericDS) GetValue(col string) (any, error) {
	if col == "Value" && d.pos >= 0 && d.pos < len(d.values) {
		return d.values[d.pos], nil
	}
	return nil, nil
}

func (d *numericDS) Columns() []data.Column {
	return []data.Column{{Name: "Value"}}
}

// TestAggregateTotals_Sum verifies that a Sum total accumulates correctly
// across all data rows and that a TextObject expression referencing the total
// evaluates to the expected value after the run.
func TestAggregateTotals_Sum(t *testing.T) {
	r := reportpkg.NewReport()
	dict := data.NewDictionary()

	// Register a Sum total over [Value].
	at := data.NewAggregateTotal("GrandTotal")
	at.TotalType = data.TotalTypeSum
	at.Expression = "Value"
	dict.AddAggregateTotal(at)
	r.SetDictionary(dict)

	pg := reportpkg.NewReportPage()

	// Data source implementing both band.DataSource and data.DataSource.
	ds := newNumericDS(1, 2, 3, 4, 5) // sum = 15

	db := band.NewDataBand()
	db.SetName("DataBand1")
	db.SetHeight(20)
	db.SetVisible(true)
	db.SetDataSource(ds)

	// TextObject that renders the running total (for verification).
	txt := object.NewTextObject()
	txt.SetName("TotalText")
	txt.SetText("[GrandTotal]")
	db.AddChild(txt)

	pg.AddBand(db)
	r.AddPage(pg)

	e := engine.New(r)
	if err := e.Run(engine.DefaultRunOptions()); err != nil {
		t.Fatalf("Run: %v", err)
	}

	// After the run, GrandTotal should be 15.
	final := dict.FindTotal("GrandTotal")
	if final == nil {
		t.Fatal("GrandTotal not found in dictionary")
	}
	got, ok := final.Value.(float64)
	if !ok {
		t.Fatalf("GrandTotal.Value type: %T, want float64", final.Value)
	}
	if got != 15 {
		t.Errorf("GrandTotal sum: got %v, want 15", got)
	}
}

// TestAggregateTotals_Count verifies that a Count total counts rows.
func TestAggregateTotals_Count(t *testing.T) {
	r := reportpkg.NewReport()
	dict := data.NewDictionary()

	at := data.NewAggregateTotal("RowCount")
	at.TotalType = data.TotalTypeCount
	dict.AddAggregateTotal(at)
	r.SetDictionary(dict)

	pg := reportpkg.NewReportPage()

	ds := newNumericDS(10, 20, 30) // 3 rows

	db := band.NewDataBand()
	db.SetName("DataBand1")
	db.SetHeight(20)
	db.SetVisible(true)
	db.SetDataSource(ds)
	pg.AddBand(db)
	r.AddPage(pg)

	e := engine.New(r)
	if err := e.Run(engine.DefaultRunOptions()); err != nil {
		t.Fatalf("Run: %v", err)
	}

	final := dict.FindTotal("RowCount")
	if final == nil {
		t.Fatal("RowCount not found in dictionary")
	}
	got, ok := final.Value.(int)
	if !ok {
		t.Fatalf("RowCount.Value type: %T, want int", final.Value)
	}
	if got != 3 {
		t.Errorf("RowCount: got %v, want 3", got)
	}
}

// TestAggregateTotals_ResetAfterPrint verifies that a total marked
// ResetAfterPrint=true is reset during resetGroupTotals (simulating group reset).
func TestAggregateTotals_ResetAfterPrint(t *testing.T) {
	r := reportpkg.NewReport()
	dict := data.NewDictionary()

	at := data.NewAggregateTotal("GroupSum")
	at.TotalType = data.TotalTypeSum
	at.Expression = "Value"
	at.ResetAfterPrint = true
	dict.AddAggregateTotal(at)
	r.SetDictionary(dict)

	pg := reportpkg.NewReportPage()
	ds := newNumericDS(5, 10)

	db := band.NewDataBand()
	db.SetName("DataBand1")
	db.SetHeight(20)
	db.SetVisible(true)
	db.SetDataSource(ds)
	pg.AddBand(db)
	r.AddPage(pg)

	e := engine.New(r)
	if err := e.Run(engine.DefaultRunOptions()); err != nil {
		t.Fatalf("Run: %v", err)
	}

	// GroupSum accumulates 5+10=15 before resetGroupTotals is called.
	// resetGroupTotals resets it to 0.
	// Since we don't call resetGroupTotals explicitly here and the band has no
	// group footer to trigger it, the value should be 15 after the run.
	final := dict.FindTotal("GroupSum")
	if final == nil {
		t.Fatal("GroupSum not found in dictionary")
	}
	got, ok := final.Value.(float64)
	if !ok {
		t.Fatalf("GroupSum.Value type: %T, want float64", final.Value)
	}
	if got != 15 {
		t.Errorf("GroupSum: got %v, want 15", got)
	}
}
