package engine_test

import (
	"testing"

	"github.com/andrewloable/go-fastreport/band"
	"github.com/andrewloable/go-fastreport/engine"
	"github.com/andrewloable/go-fastreport/reportpkg"
)

// ── benchmark helpers ─────────────────────────────────────────────────────────

// benchDS is a fast in-memory DataSource for benchmarks.
type benchDS struct {
	total int
	pos   int
}

func newBenchDS(n int) *benchDS { return &benchDS{total: n, pos: -1} }

func (d *benchDS) RowCount() int { return d.total }
func (d *benchDS) First() error  { d.pos = 0; return nil }
func (d *benchDS) Next() error   { d.pos++; return nil }
func (d *benchDS) EOF() bool     { return d.pos >= d.total }
func (d *benchDS) GetValue(col string) (any, error) {
	return d.pos, nil
}

// buildSimpleReport creates a report with a single DataBand and n rows.
func buildSimpleReport(n int) *reportpkg.Report {
	r := reportpkg.NewReport()
	pg := reportpkg.NewReportPage()
	r.AddPage(pg)

	db := band.NewDataBand()
	db.SetName("DataBand")
	db.SetHeight(20)
	db.SetVisible(true)
	db.SetDataSource(newBenchDS(n))
	pg.AddBand(db)

	return r
}

// buildHeaderFooterReport adds page header and footer bands.
func buildHeaderFooterReport(n int) *reportpkg.Report {
	r := reportpkg.NewReport()
	pg := reportpkg.NewReportPage()
	r.AddPage(pg)

	hdr := band.NewPageHeaderBand()
	hdr.SetName("PageHeader")
	hdr.SetHeight(30)
	hdr.SetVisible(true)
	pg.SetPageHeader(hdr)

	db := band.NewDataBand()
	db.SetName("DataBand")
	db.SetHeight(10)
	db.SetVisible(true)
	db.SetDataSource(newBenchDS(n))
	pg.AddBand(db)

	ftr := band.NewPageFooterBand()
	ftr.SetName("PageFooter")
	ftr.SetHeight(30)
	ftr.SetVisible(true)
	pg.SetPageFooter(ftr)

	return r
}

// ── benchmarks ────────────────────────────────────────────────────────────────

// BenchmarkRun_10Rows measures engine throughput for a small report.
func BenchmarkRun_10Rows(b *testing.B) {
	r := buildSimpleReport(10)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		e := engine.New(r)
		if err := e.Run(engine.DefaultRunOptions()); err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkRun_100Rows measures engine throughput for a medium report.
func BenchmarkRun_100Rows(b *testing.B) {
	r := buildSimpleReport(100)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		e := engine.New(r)
		if err := e.Run(engine.DefaultRunOptions()); err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkRun_1000Rows measures engine throughput for a large report.
func BenchmarkRun_1000Rows(b *testing.B) {
	r := buildSimpleReport(1000)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		e := engine.New(r)
		if err := e.Run(engine.DefaultRunOptions()); err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkRun_WithHeaderFooter measures overhead from header/footer bands.
func BenchmarkRun_WithHeaderFooter_100Rows(b *testing.B) {
	r := buildHeaderFooterReport(100)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		e := engine.New(r)
		if err := e.Run(engine.DefaultRunOptions()); err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkRun_MultiPage measures multi-page report overhead.
// Each band is 100px tall, page height ~1100px → ~11 rows/page.
func BenchmarkRun_MultiPage_1000Rows(b *testing.B) {
	r := reportpkg.NewReport()
	pg := reportpkg.NewReportPage()
	r.AddPage(pg)

	db := band.NewDataBand()
	db.SetName("DataBand")
	db.SetHeight(100)
	db.SetVisible(true)
	db.SetDataSource(newBenchDS(1000))
	pg.AddBand(db)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		e := engine.New(r)
		if err := e.Run(engine.DefaultRunOptions()); err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkRunParallel measures parallel report generation throughput.
func BenchmarkRunParallel_100Rows(b *testing.B) {
	b.RunParallel(func(pb *testing.PB) {
		// Each goroutine creates its own Report to avoid shared state.
		r := buildSimpleReport(100)
		for pb.Next() {
			e := engine.New(r)
			if err := e.Run(engine.DefaultRunOptions()); err != nil {
				b.Fatal(err)
			}
		}
	})
}
