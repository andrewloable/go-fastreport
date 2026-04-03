package reportpkg_test

import (
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/andrewloable/go-fastreport/data"
	xmldata "github.com/andrewloable/go-fastreport/data/xml"
	"github.com/andrewloable/go-fastreport/engine"
	"github.com/andrewloable/go-fastreport/reportpkg"
)

func TestXMLDataSourceImplementsMultiSortable(t *testing.T) {
	ds := xmldata.New("Products")
	_, ok := interface{}(ds).(data.MultiSortable)
	if !ok {
		t.Error("*XMLDataSource does NOT implement data.MultiSortable")
	} else {
		t.Log("*XMLDataSource implements data.MultiSortable OK")
	}
}

func TestSortGroupByTotalWithNorthWindData(t *testing.T) {
	r := loadFRXSmoke(t, "Sort Group By Total.frx")
	if !r.DoublePass {
		t.Fatal("expected DoublePass=true")
	}
	if r.CompiledScript == nil {
		t.Fatal("no compiled script")
	}
	fmt.Printf("ClassState before run: %v\n", r.CompiledScript.ClassState)

	// Load NorthWind data.
	nwindPath := testReportsDir() + "/nwind.xml"
	nwindRaw, err := os.ReadFile(nwindPath)
	if err != nil {
		t.Skipf("cannot read nwind.xml: %v", err)
	}
	loadNorthWindForReport(r, nwindRaw)

	e := engine.New(r)
	if err := e.Run(engine.DefaultRunOptions()); err != nil {
		t.Fatalf("run: %v", err)
	}

	fmt.Printf("ClassState after run: %v\n", r.CompiledScript.ClassState)

	pp := e.PreparedPages()
	if pp.Count() == 0 {
		t.Fatal("no prepared pages")
	}
	pg0 := pp.GetPage(0)
	fmt.Printf("Page 0 bands: %d\n", len(pg0.Bands))
	for i, b := range pg0.Bands {
		if i >= 8 {
			break
		}
		for j, obj := range b.Objects {
			if j >= 3 {
				break
			}
			if obj.Text != "" {
				fmt.Printf("  band[%d] obj[%d]: %q\n", i, j, obj.Text)
			}
		}
	}
}

// loadNorthWindForReport registers NorthWind XML tables into the report dictionary,
// mirroring the approach in examples/frx_to_html/main.go.
func loadNorthWindForReport(r *reportpkg.Report, raw []byte) {
	xmlStr := string(raw)
	dict := r.Dictionary()

	// Find top-level element names (table names) in the XML.
	for _, line := range strings.Split(xmlStr, "\n") {
		line = strings.TrimSpace(line)
		if !strings.HasPrefix(line, "<") || strings.HasPrefix(line, "<?") ||
			strings.HasPrefix(line, "<!") || strings.HasPrefix(line, "</") {
			continue
		}
		end := strings.IndexAny(line[1:], " \t\n\r>") + 1
		if end <= 1 {
			continue
		}
		name := line[1:end]
		if name == "" || strings.Contains(name, "/") ||
			name == "NewDataSet" || strings.HasPrefix(name, "xs:") {
			continue
		}

		ds := xmldata.New(name)
		ds.SetXML(xmlStr)
		ds.SetRowElement(name)
		if err := ds.Init(); err != nil {
			continue
		}
		ds.SetAlias(name)

		// Remove any stub data source with matching name/alias.
		var toRemove []data.DataSource
		for _, existing := range dict.DataSources() {
			if strings.EqualFold(existing.Name(), name) || strings.EqualFold(existing.Alias(), name) {
				toRemove = append(toRemove, existing)
			}
		}
		for _, existing := range toRemove {
			dict.RemoveDataSource(existing)
		}
		dict.AddDataSource(ds)
	}
}
