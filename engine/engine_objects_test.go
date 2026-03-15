package engine_test

import (
	"testing"

	"github.com/andrewloable/go-fastreport/band"
	"github.com/andrewloable/go-fastreport/barcode"
	"github.com/andrewloable/go-fastreport/engine"
	"github.com/andrewloable/go-fastreport/gauge"
	"github.com/andrewloable/go-fastreport/object"
	"github.com/andrewloable/go-fastreport/reportpkg"
	"github.com/andrewloable/go-fastreport/style"
	"github.com/andrewloable/go-fastreport/table"
)

func newObjectTestEngine(t *testing.T) *engine.ReportEngine {
	t.Helper()
	r := reportpkg.NewReport()
	pg := reportpkg.NewReportPage()
	r.AddPage(pg)
	e := engine.New(r)
	if err := e.Run(engine.DefaultRunOptions()); err != nil {
		t.Fatalf("Run: %v", err)
	}
	return e
}

// ── populateContainerChildren ─────────────────────────────────────────────────

// TestPopulateContainerChildren_ViaEngine adds a ContainerObject to a page's
// static band and runs the engine, which triggers populateContainerChildren.
func TestPopulateContainerChildren_ViaEngine(t *testing.T) {
	r := reportpkg.NewReport()
	pg := reportpkg.NewReportPage()

	// Create a page header band with a ContainerObject child.
	hdr := band.NewPageHeaderBand()
	hdr.SetName("PH")
	hdr.SetHeight(40)
	hdr.SetVisible(true)

	cont := object.NewContainerObject()
	cont.SetName("Cont")
	cont.SetLeft(0)
	cont.SetTop(0)
	cont.SetWidth(100)
	cont.SetHeight(40)
	cont.SetVisible(true)

	// Add a TextObject child so populateContainerChildren iterates the loop.
	childTxt := object.NewTextObject()
	childTxt.SetName("ChildTxt")
	childTxt.SetText("Hello")
	childTxt.SetLeft(0)
	childTxt.SetTop(0)
	childTxt.SetWidth(80)
	childTxt.SetHeight(15)
	childTxt.SetVisible(true)
	childTxt.SetFont(style.Font{Size: 10})
	cont.AddChild(childTxt)

	// Add a nested ContainerObject to trigger the recursive branch.
	nested := object.NewContainerObject()
	nested.SetName("Nested")
	nested.SetLeft(0)
	nested.SetTop(15)
	nested.SetWidth(80)
	nested.SetHeight(15)
	nested.SetVisible(true)
	cont.AddChild(nested)

	hdr.Objects().Add(cont)

	pg.SetPageHeader(hdr)
	r.AddPage(pg)

	e := engine.New(r)
	if err := e.Run(engine.DefaultRunOptions()); err != nil {
		t.Fatalf("Run with ContainerObject: %v", err)
	}
	if e.PreparedPages().Count() == 0 {
		t.Error("expected at least 1 prepared page after ContainerObject run")
	}
}

// ── populateTableObjects ──────────────────────────────────────────────────────

// TestPopulateTableObjects_ViaEngine adds a TableObject (with rows and columns)
// to a page header band, which triggers populateTableObjects during engine run.
func TestPopulateTableObjects_ViaEngine(t *testing.T) {
	r := reportpkg.NewReport()
	pg := reportpkg.NewReportPage()

	hdr := band.NewPageHeaderBand()
	hdr.SetName("PH")
	hdr.SetHeight(60)
	hdr.SetVisible(true)

	tbl := table.NewTableObject()
	tbl.SetName("Tbl")
	tbl.SetLeft(0)
	tbl.SetTop(0)
	tbl.SetVisible(true)

	col := table.NewTableColumn()
	col.SetWidth(80)
	tbl.AddColumn(col)

	row := table.NewTableRow()
	row.SetHeight(20)
	cell := table.NewTableCell()
	cell.SetName("Cell1")
	cell.SetText("Hello")
	row.AddCell(cell)
	tbl.AddRow(row)

	hdr.Objects().Add(tbl)
	pg.SetPageHeader(hdr)
	r.AddPage(pg)

	e := engine.New(r)
	if err := e.Run(engine.DefaultRunOptions()); err != nil {
		t.Fatalf("Run with TableObject: %v", err)
	}
	if e.PreparedPages().Count() == 0 {
		t.Error("expected at least 1 prepared page after TableObject run")
	}
}

// ── populateAdvMatrixCells ────────────────────────────────────────────────────

// TestPopulateAdvMatrixCells_ViaEngine adds an AdvMatrixObject with TableRows
// to a band, which triggers populateAdvMatrixCells during engine run.
func TestPopulateAdvMatrixCells_ViaEngine(t *testing.T) {
	r := reportpkg.NewReport()
	pg := reportpkg.NewReportPage()

	hdr := band.NewPageHeaderBand()
	hdr.SetName("PH")
	hdr.SetHeight(60)
	hdr.SetVisible(true)

	adv := object.NewAdvMatrixObject()
	adv.SetName("Adv")
	adv.SetLeft(0)
	adv.SetTop(0)
	adv.SetWidth(200)
	adv.SetHeight(40)
	adv.SetVisible(true)

	// Set up physical table structure.
	adv.TableColumns = []*object.AdvMatrixColumn{
		{Name: "Col1", Width: 100},
		{Name: "Col2", Width: 100},
	}
	adv.TableRows = []*object.AdvMatrixRow{
		{
			Name:   "Row1",
			Height: 20,
			Cells: []*object.AdvMatrixCell{
				{Name: "C1", Text: "A", ColSpan: 1, RowSpan: 1},
				{Name: "C2", Text: "B", ColSpan: 1, RowSpan: 1},
			},
		},
	}

	hdr.Objects().Add(adv)
	pg.SetPageHeader(hdr)
	r.AddPage(pg)

	e := engine.New(r)
	if err := e.Run(engine.DefaultRunOptions()); err != nil {
		t.Fatalf("Run with AdvMatrixObject: %v", err)
	}
	if e.PreparedPages().Count() == 0 {
		t.Error("expected at least 1 prepared page after AdvMatrixObject run")
	}
}

// ── evalGaugeValue / renderGaugeBlob (via gauge in band) ─────────────────────

// TestEvalGaugeValue_ViaLinearGaugeInBand adds a LinearGauge to a page header
// band, triggering evalGaugeValue and renderGaugeBlob via buildPreparedObject.
func TestEvalGaugeValue_ViaLinearGaugeInBand(t *testing.T) {
	r := reportpkg.NewReport()
	pg := reportpkg.NewReportPage()

	hdr := band.NewPageHeaderBand()
	hdr.SetName("PH")
	hdr.SetHeight(60)
	hdr.SetVisible(true)

	g := gauge.NewLinearGauge()
	g.SetName("LinearG")
	g.SetLeft(0)
	g.SetTop(0)
	g.SetWidth(120)
	g.SetHeight(40)
	g.SetVisible(true)
	g.GaugeObject.Minimum = 0
	g.GaugeObject.Maximum = 100
	g.GaugeObject.SetValue(50)
	// Set an expression that evaluates to PageNumber (int) to cover evalGaugeValue's type switch.
	g.GaugeObject.Expression = "PageNumber"

	hdr.Objects().Add(g)
	pg.SetPageHeader(hdr)
	r.AddPage(pg)

	e := engine.New(r)
	if err := e.Run(engine.DefaultRunOptions()); err != nil {
		t.Fatalf("Run with LinearGauge: %v", err)
	}
	if e.PreparedPages().Count() == 0 {
		t.Error("expected at least 1 prepared page after LinearGauge run")
	}
}

// ── renderBarcode / extractBarcodeModules (via BarcodeObject in band) ─────────

// TestRenderBarcode_ViaEngineRun adds a barcode.BarcodeObject (Code128) to a
// band and runs the engine, triggering renderBarcode and extractBarcodeModules.
func TestRenderBarcode_ViaEngineRun(t *testing.T) {
	r := reportpkg.NewReport()
	pg := reportpkg.NewReportPage()

	hdr := band.NewPageHeaderBand()
	hdr.SetName("PH")
	hdr.SetHeight(80)
	hdr.SetVisible(true)

	bc := barcode.NewBarcodeObject()
	bc.SetName("BC")
	bc.SetLeft(0)
	bc.SetTop(0)
	bc.SetWidth(200)
	bc.SetHeight(60)
	bc.SetVisible(true)
	bc.SetText("CODE128")
	// Barcode is pre-set to Code128 by NewBarcodeObject.

	hdr.Objects().Add(bc)
	pg.SetPageHeader(hdr)
	r.AddPage(pg)

	e := engine.New(r)
	if err := e.Run(engine.DefaultRunOptions()); err != nil {
		t.Fatalf("Run with BarcodeObject: %v", err)
	}
	if e.PreparedPages().Count() == 0 {
		t.Error("expected at least 1 prepared page after BarcodeObject run")
	}
}

// ── decodeSvgData (via SVGObject in band) ─────────────────────────────────────

// TestDecodeSvgData_ViaEngineRun adds an SVGObject with raw SVG data to a band,
// triggering decodeSvgData via buildPreparedObject.
func TestDecodeSvgData_ViaEngineRun(t *testing.T) {
	r := reportpkg.NewReport()
	pg := reportpkg.NewReportPage()

	hdr := band.NewPageHeaderBand()
	hdr.SetName("PH")
	hdr.SetHeight(60)
	hdr.SetVisible(true)

	svg := object.NewSVGObject()
	svg.SetName("SVG1")
	svg.SetLeft(0)
	svg.SetTop(0)
	svg.SetWidth(100)
	svg.SetHeight(50)
	svg.SetVisible(true)
	svg.SvgData = `<svg xmlns="http://www.w3.org/2000/svg"><rect width="100" height="50"/></svg>`

	hdr.Objects().Add(svg)
	pg.SetPageHeader(hdr)
	r.AddPage(pg)

	e := engine.New(r)
	if err := e.Run(engine.DefaultRunOptions()); err != nil {
		t.Fatalf("Run with SVGObject: %v", err)
	}
	if e.PreparedPages().Count() == 0 {
		t.Error("expected at least 1 prepared page after SVGObject run")
	}
}

// ── runDataBandHierarchical ───────────────────────────────────────────────────

// hierarchicalDS is a data source implementing both band.DataSource and
// data.DataSource, with ID and ParentID columns for hierarchical rendering.
type hierarchicalDS struct {
	pos     int
	ids     []string
	parents []string
}

func newHierarchicalDS() *hierarchicalDS {
	return &hierarchicalDS{
		pos:     -1,
		ids:     []string{"1", "2", "3"},
		parents: []string{"", "1", "1"},
	}
}

func (d *hierarchicalDS) Name() string          { return "hierarchical" }
func (d *hierarchicalDS) Alias() string         { return "hierarchical" }
func (d *hierarchicalDS) Init() error           { d.pos = -1; return nil }
func (d *hierarchicalDS) First() error          { d.pos = 0; return nil }
func (d *hierarchicalDS) Next() error           { d.pos++; return nil }
func (d *hierarchicalDS) EOF() bool             { return d.pos >= len(d.ids) }
func (d *hierarchicalDS) RowCount() int         { return len(d.ids) }
func (d *hierarchicalDS) CurrentRowNo() int     { return d.pos }
func (d *hierarchicalDS) Close() error          { return nil }
func (d *hierarchicalDS) GetValue(col string) (any, error) {
	if d.pos < 0 || d.pos >= len(d.ids) {
		return nil, nil
	}
	switch col {
	case "ID":
		return d.ids[d.pos], nil
	case "ParentID":
		return d.parents[d.pos], nil
	}
	return nil, nil
}

// TestRunDataBandHierarchical_ViaEngine sets IDColumn and ParentIDColumn on a
// DataBand and calls RunDataBandFull, triggering runDataBandHierarchical.
func TestRunDataBandHierarchical_ViaEngine(t *testing.T) {
	r := reportpkg.NewReport()
	pg := reportpkg.NewReportPage()
	r.AddPage(pg)
	e := engine.New(r)
	if err := e.Run(engine.DefaultRunOptions()); err != nil {
		t.Fatalf("Run: %v", err)
	}

	db := band.NewDataBand()
	db.SetName("HierDB")
	db.SetHeight(15)
	db.SetVisible(true)
	db.SetIDColumn("ID")
	db.SetParentIDColumn("ParentID")
	db.SetDataSource(newHierarchicalDS())

	if err := e.RunDataBandFull(db); err != nil {
		t.Fatalf("RunDataBandFull hierarchical: %v", err)
	}
}
