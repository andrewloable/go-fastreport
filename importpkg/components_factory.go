package importpkg

// ComponentsFactory provides factory functions for creating report components
// during import from external report formats (RDL, StimulSoft, etc.).
//
// Each function creates the component, wires it into the correct parent slot,
// and returns the new instance. Name validation mirrors the C# implementation:
// names that are not valid Go/language-independent identifiers are discarded
// and replaced with an auto-generated unique name via the report's name creator.
//
// It is the Go equivalent of FastReport.Import.ComponentsFactory
// (original-dotnet/FastReport.Base/Import/ComponentsFactory.cs).
//
// Usage pattern:
//
//	page := importpkg.CreateReportPage(report)
//	title := importpkg.CreateReportTitleBand(page)
//	text := importpkg.CreateTextObject("Text1", title)

import (
	"unicode"

	"github.com/andrewloable/go-fastreport/band"
	barcodeobj "github.com/andrewloable/go-fastreport/barcode"
	"github.com/andrewloable/go-fastreport/data"
	"github.com/andrewloable/go-fastreport/gauge"
	"github.com/andrewloable/go-fastreport/matrix"
	"github.com/andrewloable/go-fastreport/object"
	"github.com/andrewloable/go-fastreport/report"
	"github.com/andrewloable/go-fastreport/reportpkg"
	"github.com/andrewloable/go-fastreport/style"
	"github.com/andrewloable/go-fastreport/table"
	"github.com/andrewloable/go-fastreport/utils"
)

// isValidIdentifier reports whether name is a valid language-independent
// identifier (non-empty, starts with a letter or underscore, contains only
// letters, digits, or underscores).
//
// This mirrors FastReport's ComponentsFactory.IsValidIdentifier which calls
// CodeGenerator.IsValidLanguageIndependentIdentifier in the C# version.
// C# ref: original-dotnet/FastReport.Base/Import/ComponentsFactory.cs line 22.
func isValidIdentifier(name string) bool {
	if name == "" {
		return false
	}
	for i, r := range name {
		if i == 0 {
			if !unicode.IsLetter(r) && r != '_' {
				return false
			}
		} else {
			if !unicode.IsLetter(r) && !unicode.IsDigit(r) && r != '_' {
				return false
			}
		}
	}
	return true
}

// assignName sets name on obj. If name is not a valid identifier the name is
// left empty so the caller can generate a unique one.
// C# ref: ComponentsFactory — all Create* methods call obj.Name = name then
// call CreateUniqueName() when !IsValidIdentifier(name).
func assignName(obj report.Base, name string) {
	if isValidIdentifier(name) {
		obj.SetName(name)
	}
	// When the name is invalid the object gets no name; the caller is
	// expected to use utils.FastNameCreator.CreateUniqueName to assign one.
}

// ── Pages ────────────────────────────────────────────────────────────────────

// CreateReportPage creates a new ReportPage and appends it to report.
// The page is given a unique name (e.g. "ReportPage1", "ReportPage2", …).
//
// C# ref: ComponentsFactory.CreateReportPage(Report).
func CreateReportPage(rpt *reportpkg.Report) *reportpkg.ReportPage {
	page := reportpkg.NewReportPage()
	rpt.AddPage(page)
	ensureBaseName(page)
	nc := utils.NewFastNameCreator(collectPageNamers(rpt))
	nc.CreateUniqueName(page)
	return page
}

// CreateReportPageNamed creates a new ReportPage with the given name and
// appends it to report. If name is not a valid identifier, a unique name is
// generated instead.
//
// C# ref: ComponentsFactory.CreateReportPage(string, Report).
func CreateReportPageNamed(name string, rpt *reportpkg.Report) *reportpkg.ReportPage {
	page := reportpkg.NewReportPage()
	rpt.AddPage(page)
	if isValidIdentifier(name) {
		page.SetName(name)
	} else {
		ensureBaseName(page)
		nc := utils.NewFastNameCreator(collectPageNamers(rpt))
		nc.CreateUniqueName(page)
	}
	return page
}

// collectPageNamers builds an ObjectNamer slice from all pages in the report
// so that FastNameCreator can avoid name collisions.
func collectPageNamers(rpt *reportpkg.Report) []utils.ObjectNamer {
	pages := rpt.Pages()
	namers := make([]utils.ObjectNamer, 0, len(pages))
	for _, p := range pages {
		namers = append(namers, p)
	}
	return namers
}

// ── Bands ────────────────────────────────────────────────────────────────────

// CreateReportTitleBand creates a ReportTitleBand and assigns it to page.ReportTitle.
//
// C# ref: ComponentsFactory.CreateReportTitleBand(ReportPage).
func CreateReportTitleBand(page *reportpkg.ReportPage) *band.ReportTitleBand {
	b := band.NewReportTitleBand()
	page.SetReportTitle(b)
	genUniqueBandName(b, page)
	return b
}

// CreateReportSummaryBand creates a ReportSummaryBand and assigns it to page.ReportSummary.
//
// C# ref: ComponentsFactory.CreateReportSummaryBand(ReportPage).
func CreateReportSummaryBand(page *reportpkg.ReportPage) *band.ReportSummaryBand {
	b := band.NewReportSummaryBand()
	page.SetReportSummary(b)
	genUniqueBandName(b, page)
	return b
}

// CreatePageHeaderBand creates a PageHeaderBand and assigns it to page.PageHeader.
//
// C# ref: ComponentsFactory.CreatePageHeaderBand(ReportPage).
func CreatePageHeaderBand(page *reportpkg.ReportPage) *band.PageHeaderBand {
	b := band.NewPageHeaderBand()
	page.SetPageHeader(b)
	genUniqueBandName(b, page)
	return b
}

// CreatePageFooterBand creates a PageFooterBand and assigns it to page.PageFooter.
//
// C# ref: ComponentsFactory.CreatePageFooterBand(ReportPage).
func CreatePageFooterBand(page *reportpkg.ReportPage) *band.PageFooterBand {
	b := band.NewPageFooterBand()
	page.SetPageFooter(b)
	genUniqueBandName(b, page)
	return b
}

// CreateColumnHeaderBand creates a ColumnHeaderBand and assigns it to page.ColumnHeader.
//
// C# ref: ComponentsFactory.CreateColumnHeaderBand(ReportPage).
func CreateColumnHeaderBand(page *reportpkg.ReportPage) *band.ColumnHeaderBand {
	b := band.NewColumnHeaderBand()
	page.SetColumnHeader(b)
	genUniqueBandName(b, page)
	return b
}

// CreateColumnFooterBand creates a ColumnFooterBand and assigns it to page.ColumnFooter.
//
// C# ref: ComponentsFactory.CreateColumnFooterBand(ReportPage).
func CreateColumnFooterBand(page *reportpkg.ReportPage) *band.ColumnFooterBand {
	b := band.NewColumnFooterBand()
	page.SetColumnFooter(b)
	genUniqueBandName(b, page)
	return b
}

// CreateDataBand creates a DataBand and appends it to page.Bands.
//
// C# ref: ComponentsFactory.CreateDataBand(ReportPage).
func CreateDataBand(page *reportpkg.ReportPage) *band.DataBand {
	b := band.NewDataBand()
	page.AddBand(b)
	genUniqueBandName(b, page)
	return b
}

// CreateDataHeaderBand creates a DataHeaderBand and sets it as the header of data.
//
// C# ref: ComponentsFactory.CreateDataHeaderBand(DataBand).
func CreateDataHeaderBand(data *band.DataBand) *band.DataHeaderBand {
	b := band.NewDataHeaderBand()
	data.SetHeader(b)
	genUniqueNameFromBase(b)
	return b
}

// CreateDataFooterBand creates a DataFooterBand and sets it as the footer of data.
//
// C# ref: ComponentsFactory.CreateDataFooterBand(DataBand).
func CreateDataFooterBand(data *band.DataBand) *band.DataFooterBand {
	b := band.NewDataFooterBand()
	data.SetFooter(b)
	genUniqueNameFromBase(b)
	return b
}

// CreateGroupHeaderBand creates a GroupHeaderBand and appends it to page.Bands.
//
// C# ref: ComponentsFactory.CreateGroupHeaderBand(ReportPage).
func CreateGroupHeaderBand(page *reportpkg.ReportPage) *band.GroupHeaderBand {
	b := band.NewGroupHeaderBand()
	page.AddBand(b)
	genUniqueBandName(b, page)
	return b
}

// CreateGroupFooterBandOnPage creates a GroupFooterBand and appends it to page.Bands.
//
// C# ref: ComponentsFactory.CreateGroupFooterBand(ReportPage).
func CreateGroupFooterBandOnPage(page *reportpkg.ReportPage) *band.GroupFooterBand {
	b := band.NewGroupFooterBand()
	page.AddBand(b)
	genUniqueBandName(b, page)
	return b
}

// CreateGroupFooterBand creates a GroupFooterBand and assigns it as the footer
// of the given GroupHeaderBand.
//
// C# ref: ComponentsFactory.CreateGroupFooterBand(GroupHeaderBand).
func CreateGroupFooterBand(groupHeader *band.GroupHeaderBand) *band.GroupFooterBand {
	b := band.NewGroupFooterBand()
	groupHeader.SetGroupFooter(b)
	genUniqueNameFromBase(b)
	return b
}

// CreateChildBand creates a ChildBand and adds it as a child of parent.
// BandBase.AddChild handles ChildBand specially by storing it in b.child.
//
// C# ref: ComponentsFactory.CreateChildBand(BandBase).
// The C# signature accepts BandBase; in Go, any report.Parent is accepted
// so that callers do not need to extract the embedded BandBase field.
func CreateChildBand(parent report.Parent) *band.ChildBand {
	b := band.NewChildBand()
	parent.AddChild(b)
	genUniqueNameFromBase(b)
	return b
}

// CreateOverlayBand creates an OverlayBand and assigns it to page.Overlay.
//
// C# ref: ComponentsFactory.CreateOverlayBand(ReportPage).
func CreateOverlayBand(page *reportpkg.ReportPage) *band.OverlayBand {
	b := band.NewOverlayBand()
	page.SetOverlay(b)
	genUniqueBandName(b, page)
	return b
}

// ── Objects ──────────────────────────────────────────────────────────────────

// CreateStyle creates a named StyleEntry and adds it to the report's style sheet.
//
// C# ref: ComponentsFactory.CreateStyle(string, Report).
func CreateStyle(name string, rpt *reportpkg.Report) *style.StyleEntry {
	e := &style.StyleEntry{Name: name}
	rpt.Styles().Add(e)
	return e
}

// CreateTextObject creates a TextObject with name and adds it to parent if
// parent.CanContain allows it. If name is not a valid identifier, a unique
// name is generated.
//
// C# ref: ComponentsFactory.CreateTextObject(string, Base).
func CreateTextObject(name string, parent report.Parent) *object.TextObject {
	obj := object.NewTextObject()
	assignName(obj, name)
	if parent.CanContain(obj) {
		parent.AddChild(obj)
	}
	if !isValidIdentifier(obj.Name()) {
		genUniqueNameFromBase(obj)
	}
	return obj
}

// CreatePictureObject creates a PictureObject with name and adds it to parent.
//
// C# ref: ComponentsFactory.CreatePictureObject(string, Base).
func CreatePictureObject(name string, parent report.Parent) *object.PictureObject {
	obj := object.NewPictureObject()
	assignName(obj, name)
	if parent.CanContain(obj) {
		parent.AddChild(obj)
	}
	if !isValidIdentifier(obj.Name()) {
		genUniqueNameFromBase(obj)
	}
	return obj
}

// CreateLineObject creates a LineObject with name and adds it to parent.
//
// C# ref: ComponentsFactory.CreateLineObject(string, Base).
func CreateLineObject(name string, parent report.Parent) *object.LineObject {
	obj := object.NewLineObject()
	assignName(obj, name)
	if parent.CanContain(obj) {
		parent.AddChild(obj)
	}
	if !isValidIdentifier(obj.Name()) {
		genUniqueNameFromBase(obj)
	}
	return obj
}

// CreateShapeObject creates a ShapeObject with name and adds it to parent.
//
// C# ref: ComponentsFactory.CreateShapeObject(string, Base).
func CreateShapeObject(name string, parent report.Parent) *object.ShapeObject {
	obj := object.NewShapeObject()
	assignName(obj, name)
	if parent.CanContain(obj) {
		parent.AddChild(obj)
	}
	if !isValidIdentifier(obj.Name()) {
		genUniqueNameFromBase(obj)
	}
	return obj
}

// CreatePolyLineObject creates a PolyLineObject with name and adds it to parent.
//
// C# ref: ComponentsFactory.CreatePolyLineObject(string, Base).
func CreatePolyLineObject(name string, parent report.Parent) *object.PolyLineObject {
	obj := object.NewPolyLineObject()
	assignName(obj, name)
	if parent.CanContain(obj) {
		parent.AddChild(obj)
	}
	if !isValidIdentifier(obj.Name()) {
		genUniqueNameFromBase(obj)
	}
	return obj
}

// CreatePolygonObject creates a PolygonObject with name and adds it to parent.
//
// C# ref: ComponentsFactory.CreatePolygonObject(string, Base).
func CreatePolygonObject(name string, parent report.Parent) *object.PolygonObject {
	obj := object.NewPolygonObject()
	assignName(obj, name)
	if parent.CanContain(obj) {
		parent.AddChild(obj)
	}
	if !isValidIdentifier(obj.Name()) {
		genUniqueNameFromBase(obj)
	}
	return obj
}

// CreateSubreportObject creates a SubreportObject with name and adds it to parent.
//
// C# ref: ComponentsFactory.CreateSubreportObject(string, Base).
func CreateSubreportObject(name string, parent report.Parent) *object.SubreportObject {
	obj := object.NewSubreportObject()
	assignName(obj, name)
	if parent.CanContain(obj) {
		parent.AddChild(obj)
	}
	if !isValidIdentifier(obj.Name()) {
		genUniqueNameFromBase(obj)
	}
	return obj
}

// CreateContainerObject creates a ContainerObject with name and adds it to parent.
//
// C# ref: ComponentsFactory.CreateContainerObject(string, Base).
func CreateContainerObject(name string, parent report.Parent) *object.ContainerObject {
	obj := object.NewContainerObject()
	assignName(obj, name)
	if parent.CanContain(obj) {
		parent.AddChild(obj)
	}
	if !isValidIdentifier(obj.Name()) {
		genUniqueNameFromBase(obj)
	}
	return obj
}

// CreateCheckBoxObject creates a CheckBoxObject with name and adds it to parent.
//
// C# ref: ComponentsFactory.CreateCheckBoxObject(string, Base).
func CreateCheckBoxObject(name string, parent report.Parent) *object.CheckBoxObject {
	obj := object.NewCheckBoxObject()
	assignName(obj, name)
	if parent.CanContain(obj) {
		parent.AddChild(obj)
	}
	if !isValidIdentifier(obj.Name()) {
		genUniqueNameFromBase(obj)
	}
	return obj
}

// CreateHtmlObject creates an HtmlObject with name and adds it to parent.
//
// C# ref: ComponentsFactory.CreateHtmlObject(string, Base).
func CreateHtmlObject(name string, parent report.Parent) *object.HtmlObject {
	obj := object.NewHtmlObject()
	assignName(obj, name)
	if parent.CanContain(obj) {
		parent.AddChild(obj)
	}
	if !isValidIdentifier(obj.Name()) {
		genUniqueNameFromBase(obj)
	}
	return obj
}

// CreateTableObject creates a TableObject with name and adds it to parent.
//
// C# ref: ComponentsFactory.CreateTableObject(string, Base).
func CreateTableObject(name string, parent report.Parent) *table.TableObject {
	obj := table.NewTableObject()
	assignName(obj, name)
	if parent.CanContain(obj) {
		parent.AddChild(obj)
	}
	if !isValidIdentifier(obj.Name()) {
		genUniqueNameFromBase(obj)
	}
	return obj
}

// CreateMatrixObject creates a MatrixObject with name and adds it to parent.
//
// C# ref: ComponentsFactory.CreateMatrixObject(string, Base).
func CreateMatrixObject(name string, parent report.Parent) *matrix.MatrixObject {
	obj := matrix.New()
	assignName(obj, name)
	if parent.CanContain(obj) {
		parent.AddChild(obj)
	}
	if !isValidIdentifier(obj.Name()) {
		genUniqueNameFromBase(obj)
	}
	return obj
}

// CreateBarcodeObject creates a BarcodeObject with name and adds it to parent.
//
// C# ref: ComponentsFactory.CreateBarcodeObject(string, Base).
func CreateBarcodeObject(name string, parent report.Parent) *barcodeobj.BarcodeObject {
	obj := barcodeobj.NewBarcodeObject()
	assignName(obj, name)
	if parent.CanContain(obj) {
		parent.AddChild(obj)
	}
	if !isValidIdentifier(obj.Name()) {
		genUniqueNameFromBase(obj)
	}
	return obj
}

// CreateZipCodeObject creates a ZipCodeObject with name and adds it to parent.
//
// C# ref: ComponentsFactory.CreateZipCodeObject(string, Base).
func CreateZipCodeObject(name string, parent report.Parent) *object.ZipCodeObject {
	obj := object.NewZipCodeObject()
	assignName(obj, name)
	if parent.CanContain(obj) {
		parent.AddChild(obj)
	}
	if !isValidIdentifier(obj.Name()) {
		genUniqueNameFromBase(obj)
	}
	return obj
}

// CreateCellularTextObject creates a CellularTextObject with name and adds it
// to parent.
//
// C# ref: ComponentsFactory.CreateCellularTextObject(string, Base).
func CreateCellularTextObject(name string, parent report.Parent) *object.CellularTextObject {
	obj := object.NewCellularTextObject()
	assignName(obj, name)
	if parent.CanContain(obj) {
		parent.AddChild(obj)
	}
	if !isValidIdentifier(obj.Name()) {
		genUniqueNameFromBase(obj)
	}
	return obj
}

// CreateLinearGauge creates a LinearGauge with name and adds it to parent.
//
// C# ref: ComponentsFactory.CreateLinearGauge(string, Base).
func CreateLinearGauge(name string, parent report.Parent) *gauge.LinearGauge {
	obj := gauge.NewLinearGauge()
	assignName(obj, name)
	if parent.CanContain(obj) {
		parent.AddChild(obj)
	}
	if !isValidIdentifier(obj.Name()) {
		genUniqueNameFromBase(obj)
	}
	return obj
}

// CreateSimpleGauge creates a SimpleGauge with name and adds it to parent.
//
// C# ref: ComponentsFactory.CreateSimpleGauge(string, Base).
func CreateSimpleGauge(name string, parent report.Parent) *gauge.SimpleGauge {
	obj := gauge.NewSimpleGauge()
	assignName(obj, name)
	if parent.CanContain(obj) {
		parent.AddChild(obj)
	}
	if !isValidIdentifier(obj.Name()) {
		genUniqueNameFromBase(obj)
	}
	return obj
}

// CreateRadialGauge creates a RadialGauge with name and adds it to parent.
//
// C# ref: ComponentsFactory.CreateRadialGauge(string, Base).
func CreateRadialGauge(name string, parent report.Parent) *gauge.RadialGauge {
	obj := gauge.NewRadialGauge()
	assignName(obj, name)
	if parent.CanContain(obj) {
		parent.AddChild(obj)
	}
	if !isValidIdentifier(obj.Name()) {
		genUniqueNameFromBase(obj)
	}
	return obj
}

// CreateSimpleProgressGauge creates a SimpleProgressGauge with name and adds
// it to parent.
//
// C# ref: ComponentsFactory.CreateSimpleProgressGauge(string, Base).
func CreateSimpleProgressGauge(name string, parent report.Parent) *gauge.SimpleProgressGauge {
	obj := gauge.NewSimpleProgressGauge()
	assignName(obj, name)
	if parent.CanContain(obj) {
		parent.AddChild(obj)
	}
	if !isValidIdentifier(obj.Name()) {
		genUniqueNameFromBase(obj)
	}
	return obj
}

// ── Dictionary Elements ──────────────────────────────────────────────────────

// CreateParameter creates a Parameter with the given name and adds it to the
// report's Dictionary.
//
// C# ref: ComponentsFactory.CreateParameter(string, Report).
func CreateParameter(name string, rpt *reportpkg.Report) *data.Parameter {
	p := &data.Parameter{Name: name}
	rpt.Dictionary().AddParameter(p)
	return p
}

// ── Name generation helpers ──────────────────────────────────────────────────

// typeNamed is the interface for objects that have a TypeName() method, which
// all serializable report objects implement.
type typeNamed interface {
	TypeName() string
}

// ensureBaseName ensures that obj has a non-empty BaseName so that
// FastNameCreator generates meaningful names (e.g. "PageHeader1" instead of "1").
// If BaseName is already set it is left unchanged. Otherwise it is set from
// TypeName() (when available) or "Object".
func ensureBaseName(obj utils.ObjectNamer) {
	if obj.BaseName() != "" {
		return
	}
	if tn, ok := obj.(typeNamed); ok {
		name := tn.TypeName()
		if name != "" {
			if b, ok := obj.(interface{ SetBaseName(string) }); ok {
				b.SetBaseName(name)
			}
			return
		}
	}
	if b, ok := obj.(interface{ SetBaseName(string) }); ok {
		b.SetBaseName("Object")
	}
}

// genUniqueNameFromBase generates a unique name for obj using only its own
// BaseName (seeded from TypeName if needed). Used for objects that do not
// participate in a broader name-space (e.g. DataHeaderBand, DataFooterBand
// live inside a DataBand rather than directly on the page).
func genUniqueNameFromBase(obj utils.ObjectNamer) {
	ensureBaseName(obj)
	nc := utils.NewFastNameCreator(nil)
	nc.CreateUniqueName(obj)
}

// genUniqueBandName generates a unique name for a band, seeding the name
// creator with all existing bands on the page to avoid collisions.
func genUniqueBandName(b utils.ObjectNamer, page *reportpkg.ReportPage) {
	ensureBaseName(b)
	namers := bandNamers(page)
	nc := utils.NewFastNameCreator(namers)
	nc.CreateUniqueName(b)
}

// bandNamers collects ObjectNamer from every band on a page so that new band
// names do not collide with existing ones.
func bandNamers(page *reportpkg.ReportPage) []utils.ObjectNamer {
	all := page.AllBands()
	namers := make([]utils.ObjectNamer, 0, len(all))
	for _, b := range all {
		if n, ok := b.(utils.ObjectNamer); ok {
			namers = append(namers, n)
		}
	}
	return namers
}
