// Package rdlc provides an importer for Microsoft RDL/RDLC report definitions.
//
// It is the Go equivalent of FastReport.Import.RDL (original-dotnet/
// FastReport.Base/Import/RDL/RDLImport.cs, UnitsConverter.cs, SizeUnits.cs,
// and ImportTable.cs).
//
// Usage:
//
//	rpt := reportpkg.NewReport()
//	imp := rdlc.New()
//	if err := imp.LoadReportFromFile(rpt, "report.rdlc"); err != nil {
//	    log.Fatal(err)
//	}
package rdlc

import (
	"encoding/xml"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/andrewloable/go-fastreport/importpkg"
	"github.com/andrewloable/go-fastreport/object"
	"github.com/andrewloable/go-fastreport/report"
	"github.com/andrewloable/go-fastreport/reportpkg"
	"github.com/andrewloable/go-fastreport/style"
	"github.com/andrewloable/go-fastreport/table"
)

// RDLImport is the RDL/RDLC report importer.
// It converts a Microsoft RDL or RDLC report definition into a go-fastreport
// Report, using the same component hierarchy as FastReport .NET.
//
// It is the Go equivalent of FastReport.Import.RDL.RDLImport
// (original-dotnet/FastReport.Base/Import/RDL/RDLImport.cs).
type RDLImport struct {
	importpkg.ImportBase

	// page is the current ReportPage being built.
	page *reportpkg.ReportPage

	// component is the most-recently created object during traversal.
	// C# ref: RDLImport.component (line 19).
	component interface{}

	// parent is the current parent container.
	parent report.Parent

	// defaultFontFamily is populated from the df:DefaultFontFamily element.
	// C# ref: RDLImport.defaultFontFamily.
	defaultFontFamily string

	// dataSetName holds the first DataSet name found in the RDL.
	// C# ref: RDLImport.dataSetName.
	dataSetName string

	// firstRun tracks whether we are processing the very first TextRun in a
	// Paragraphs block.
	// C# ref: RDLImport.firstRun.
	firstRun bool

	// filename is the path of the .rdl / .rdlc file being imported.
	// C# ref: RDLImport.filename.
	filename string

	// sectionWidth is the width of the ReportSection (RDLC only).
	// C# ref: RDLImport.sectionWidth.
	sectionWidth float32

	// reportNode is the parsed root XML element.
	// C# ref: RDLImport.reportNode.
	reportNode *xmlNode
}

// New creates a new RDLImport with its plugin name set.
func New() *RDLImport {
	imp := &RDLImport{}
	imp.SetName("RDL/RDLC Importer")
	return imp
}

// LoadReportFromFile opens path (.rdl or .rdlc), parses it, and populates
// report.
// C# ref: RDLImport.LoadReport(Report, string)
func (imp *RDLImport) LoadReportFromFile(rpt *reportpkg.Report, path string) error {
	imp.filename = path
	f, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("rdlc: open %q: %w", path, err)
	}
	defer f.Close()
	return imp.LoadReportFromStream(rpt, f)
}

// LoadReportFromStream parses the RDL/RDLC XML from r and populates report.
// C# ref: RDLImport.LoadReport(Report, Stream)
func (imp *RDLImport) LoadReportFromStream(rpt *reportpkg.Report, r io.Reader) error {
	imp.SetReport(rpt)
	imp.page = nil
	imp.defaultFontFamily = ""
	imp.dataSetName = ""
	imp.sectionWidth = 0

	root, err := parseXML(r)
	if err != nil {
		return fmt.Errorf("rdlc: parse XML: %w", err)
	}
	imp.reportNode = root
	imp.loadReport(root)
	return nil
}

// ── Internal helpers ──────────────────────────────────────────────────────────

// loadReport processes the root <Report> element.
// C# ref: RDLImport.LoadReport(XmlNode) private method.
func (imp *RDLImport) loadReport(node *xmlNode) {
	pageNbr := 0
	for _, child := range node.Children {
		switch child.Name {
		case "Description":
			imp.Report().Info.Description = child.Text
		case "Author":
			imp.Report().Info.Author = child.Text
		case "Body":
			if imp.page == nil {
				imp.page = importpkg.CreateReportPage(imp.Report())
			}
			imp.loadBody(child)
		case "Page":
			if pageNbr > 0 {
				imp.page = importpkg.CreateReportPage(imp.Report())
			}
			imp.loadPage(child)
			pageNbr++
		case "ReportSections":
			// RDLC format wraps Body/Page in a ReportSection element.
			section := child.FirstChildNamed("ReportSection")
			if section == nil {
				break
			}
			for _, sectionItem := range section.Children {
				switch sectionItem.Name {
				case "Body":
					if imp.page == nil {
						imp.page = importpkg.CreateReportPage(imp.Report())
					}
					imp.loadBody(sectionItem)
				case "Page":
					if pageNbr > 0 {
						imp.page = importpkg.CreateReportPage(imp.Report())
					}
					imp.loadPage(sectionItem)
					pageNbr++
				case "Width":
					imp.sectionWidth = sizeToMillimeters(sectionItem.Text)
				}
			}
		case "df:DefaultFontFamily":
			imp.defaultFontFamily = child.Text
		case "DataSets":
			if first := child.FirstChild(); first != nil {
				imp.dataSetName = first.AttrOrDefault("Name", "")
			}
		case "ReportParameters":
			if child.HasChildren() {
				imp.loadParameters(child)
			}
		}
	}
	if imp.page == nil {
		imp.page = importpkg.CreateReportPage(imp.Report())
		imp.loadPage(node)
	}
}

// loadBody processes the <Body> element.
// C# ref: RDLImport.LoadBody
func (imp *RDLImport) loadBody(bodyNode *xmlNode) {
	db := importpkg.CreateDataBand(imp.page)
	imp.parent = db
	for _, node := range bodyNode.Children {
		switch node.Name {
		case "ReportItems":
			imp.loadReportItems(node)
		case "Height":
			db.SetHeight(sizeToPixels(node.Text))
		}
	}
}

// loadPage processes a <Page> element.
// C# ref: RDLImport.LoadPage
func (imp *RDLImport) loadPage(pageNode *xmlNode) {
	pageWidthLoaded := false
	for _, node := range pageNode.Children {
		switch node.Name {
		case "PageHeader", "PageFooter":
			imp.loadPageSection(node)
		case "PageHeight":
			imp.page.PaperHeight = sizeToMillimeters(node.Text)
		case "PageWidth":
			imp.page.PaperWidth = sizeToMillimeters(node.Text)
			pageWidthLoaded = true
		case "LeftMargin":
			imp.page.LeftMargin = sizeToMillimeters(node.Text)
		case "RightMargin":
			imp.page.RightMargin = sizeToMillimeters(node.Text)
		case "TopMargin":
			imp.page.TopMargin = sizeToMillimeters(node.Text)
		case "BottomMargin":
			imp.page.BottomMargin = sizeToMillimeters(node.Text)
		}
	}
	if !pageWidthLoaded && imp.sectionWidth > 0 {
		imp.page.PaperWidth = imp.page.LeftMargin + imp.sectionWidth + imp.page.RightMargin
	}
}

// loadPageSection processes <PageHeader> or <PageFooter>.
// C# ref: RDLImport.LoadPageSection
func (imp *RDLImport) loadPageSection(node *xmlNode) {
	switch node.Name {
	case "PageHeader":
		ph := importpkg.CreatePageHeaderBand(imp.page)
		ph.SetPrintOn(report.PrintOnEvenPages | report.PrintOnOddPages | report.PrintOnRepeatedBand)
		for _, child := range node.Children {
			switch child.Name {
			case "Height":
				ph.SetHeight(sizeToPixels(child.Text))
			case "PrintOnFirstPage":
				if booleanToBool(child.Text) {
					ph.SetPrintOn(ph.PrintOn() | report.PrintOnFirstPage)
				}
			case "PrintOnLastPage":
				if booleanToBool(child.Text) {
					ph.SetPrintOn(ph.PrintOn() | report.PrintOnLastPage)
				}
			case "ReportItems":
				imp.parent = ph
				imp.loadReportItems(child)
			}
		}
	case "PageFooter":
		pf := importpkg.CreatePageFooterBand(imp.page)
		pf.SetPrintOn(report.PrintOnEvenPages | report.PrintOnOddPages | report.PrintOnRepeatedBand)
		for _, child := range node.Children {
			switch child.Name {
			case "Height":
				pf.SetHeight(sizeToPixels(child.Text))
			case "PrintOnFirstPage":
				if booleanToBool(child.Text) {
					pf.SetPrintOn(pf.PrintOn() | report.PrintOnFirstPage)
				}
			case "PrintOnLastPage":
				if booleanToBool(child.Text) {
					pf.SetPrintOn(pf.PrintOn() | report.PrintOnLastPage)
				}
			case "ReportItems":
				imp.parent = pf
				imp.loadReportItems(child)
			}
		}
	}
}

// loadParameters creates report-level parameters from <ReportParameters>.
// C# ref: RDLImport.LoadParameters
func (imp *RDLImport) loadParameters(parametersNode *xmlNode) {
	for _, node := range parametersNode.Children {
		name := node.AttrOrDefault("Name", "")
		if name == "" {
			continue
		}
		p := importpkg.CreateParameter(name, imp.Report())
		for _, sub := range node.Children {
			switch sub.Name {
			case "Prompt":
				p.Description = sub.Text
			case "DataType":
				p.DataType = sub.Text
			}
		}
	}
}

// loadReportItems iterates a <ReportItems> node and dispatches each child.
// C# ref: RDLImport.LoadReportItems
func (imp *RDLImport) loadReportItems(node *xmlNode) {
	for _, child := range node.Children {
		switch child.Name {
		case "Line":
			imp.loadLine(child)
		case "Rectangle":
			imp.loadRectangle(child)
		case "Textbox":
			imp.loadTextbox(child)
		case "Image":
			imp.loadImage(child)
		case "Subreport":
			imp.loadSubreport(child)
		case "Table":
			imp.loadTable(child)
		case "Tablix":
			if isTablixMatrix(child) {
				imp.loadMatrix(child)
			} else {
				imp.loadTable(child)
			}
		case "Matrix":
			imp.loadMatrix(child)
		}
	}
}

// ── Component-level loaders ───────────────────────────────────────────────────

// loadLine creates a LineObject.
// C# ref: RDLImport.LoadLine
func (imp *RDLImport) loadLine(node *xmlNode) {
	name := node.AttrOrDefault("Name", "")
	obj := importpkg.CreateLineObject(name, imp.parent)
	imp.component = obj
	imp.loadReportItem(node.Children)
}

// loadRectangle creates either a ShapeObject or a ContainerObject.
// C# ref: RDLImport.LoadRectangle / LoadContainerRectangle
func (imp *RDLImport) loadRectangle(node *xmlNode) {
	if rectangleHasReportItems(node.Children) {
		imp.loadContainerRectangle(node)
		return
	}
	name := node.AttrOrDefault("Name", "")
	shape := importpkg.CreateShapeObject(name, imp.parent)
	shape.SetShape(object.ShapeKindRectangle)
	imp.component = shape
	imp.loadReportItem(node.Children)
	for _, child := range node.Children {
		if child.Name == "ReportItems" {
			imp.loadReportItems(child)
		}
	}
}

// loadContainerRectangle creates a ContainerObject that holds nested items.
// C# ref: RDLImport.LoadContainerRectangle
func (imp *RDLImport) loadContainerRectangle(node *xmlNode) {
	savedParent := imp.parent
	name := node.AttrOrDefault("Name", "")
	cont := importpkg.CreateContainerObject(name, imp.parent)
	b := cont.Border()
	b.VisibleLines = style.BorderLinesAll
	b.SetColor(style.ColorBlack)
	cont.SetBorder(b)
	imp.component = cont
	imp.parent = cont
	imp.loadReportItem(node.Children)
	for _, child := range node.Children {
		if child.Name == "ReportItems" {
			imp.loadReportItems(child)
		}
	}
	imp.parent = savedParent
}

// rectangleHasReportItems returns true when children contain a ReportItems
// node with at least one item.
// C# ref: RDLImport.RectangleExistReportItem
func rectangleHasReportItems(children []*xmlNode) bool {
	for _, node := range children {
		if node.Name != "ReportItems" {
			continue
		}
		for _, item := range node.Children {
			switch item.Name {
			case "Line", "Rectangle", "Textbox", "Image",
				"Subreport", "Chart", "Table", "Matrix":
				return true
			}
		}
	}
	return false
}

// tableCellParent is a thin wrapper that exposes a *table.TableCell as a
// report.Parent. TableCell embeds TextObject which is not itself a report.Parent.
// We use this sentinel to detect "currently inside a table cell" in loadReportItem.
type tableCellParent struct {
	cell *table.TableCell
}

// CanContain accepts TextObject (we use the cell's own TextObject directly).
func (tcp *tableCellParent) CanContain(obj report.Base) bool {
	_, ok := obj.(*object.TextObject)
	return ok
}

// AddChild is a no-op; the cell's embedded TextObject is used directly.
func (tcp *tableCellParent) AddChild(_ report.Base) {}

// RemoveChild is a no-op.
func (tcp *tableCellParent) RemoveChild(_ report.Base) {}

// GetChildObjects adds no children (the cell's own TextObject is not exposed
// via this path).
func (tcp *tableCellParent) GetChildObjects(_ *[]report.Base) {}

// GetChildOrder always returns 0.
func (tcp *tableCellParent) GetChildOrder(_ report.Base) int { return 0 }

// SetChildOrder is a no-op.
func (tcp *tableCellParent) SetChildOrder(_ report.Base, _ int) {}

// UpdateLayout is a no-op.
func (tcp *tableCellParent) UpdateLayout(_, _ float32) {}

// loadTextbox creates a TextObject (or reuses the cell's embedded one).
// C# ref: RDLImport.LoadTextbox
func (imp *RDLImport) loadTextbox(node *xmlNode) {
	var txt *object.TextObject
	if tcp, ok := imp.parent.(*tableCellParent); ok {
		txt = &tcp.cell.TextObject
	} else {
		name := node.AttrOrDefault("Name", "")
		txt = importpkg.CreateTextObject(name, imp.parent)
	}
	imp.component = txt
	imp.loadReportItem(node.Children)
	for _, child := range node.Children {
		switch child.Name {
		case "CanGrow":
			txt.SetCanGrow(booleanToBool(child.Text))
		case "CanShrink":
			txt.SetCanShrink(booleanToBool(child.Text))
		case "HideDuplicates":
			txt.SetDuplicates(object.DuplicatesHide)
		case "Value":
			txt.SetText(child.Text)
		case "Paragraphs":
			imp.loadParagraphs(child, txt)
		}
	}
}

// loadImage creates a PictureObject.
// C# ref: RDLImport.LoadImage
func (imp *RDLImport) loadImage(node *xmlNode) {
	name := node.AttrOrDefault("Name", "")
	pic := importpkg.CreatePictureObject(name, imp.parent)
	imp.component = pic
	imp.loadReportItem(node.Children)
	for _, child := range node.Children {
		switch child.Name {
		case "Value":
			if _, err := os.Stat(child.Text); err == nil {
				pic.SetImageLocation(child.Text)
			}
		case "Sizing":
			pic.SetSizeMode(convertSizing(child.Text))
		}
	}
}

// loadSubreport creates a SubreportObject and recursively imports nested files.
// C# ref: RDLImport.LoadSubreport
func (imp *RDLImport) loadSubreport(node *xmlNode) {
	name := node.AttrOrDefault("Name", "")
	sub := importpkg.CreateSubreportObject(name, imp.parent)
	subPage := importpkg.CreateReportPage(imp.Report())
	sub.SetReportPageName(subPage.Name())
	imp.component = sub

	reportName := ""
	for _, child := range node.Children {
		if child.Name == "ReportName" {
			reportName = child.Text
		}
	}

	dir := ""
	if imp.filename != "" {
		dir = filepath.Dir(imp.filename)
	}

	subFile := ""
	if reportName != "" && dir != "" {
		candidate := filepath.Join(dir, reportName+".rdl")
		if _, err := os.Stat(candidate); err == nil {
			subFile = candidate
		} else {
			candidate = filepath.Join(dir, reportName+".rdlc")
			if _, err := os.Stat(candidate); err == nil {
				subFile = candidate
			}
		}
	}

	if subFile != "" {
		subImp := &RDLImport{filename: subFile}
		subImp.SetReport(imp.Report())
		subImp.page = subPage
		f, err := os.Open(subFile)
		if err == nil {
			defer f.Close()
			root, err2 := parseXML(f)
			if err2 == nil {
				subImp.reportNode = root
				subImp.loadReport(root)
			}
		}
	} else {
		db := importpkg.CreateDataBand(subPage)
		db.SetHeight(2.0 * 37.8) // 2 cm in pixels
		imp.loadReportItem(node.Children)
	}
}

// ── Style / appearance loaders ────────────────────────────────────────────────

// loadReportItem applies position/size/visibility/style to the current component.
// C# ref: RDLImport.LoadReportItem
func (imp *RDLImport) loadReportItem(children []*xmlNode) {
	for _, node := range children {
		switch node.Name {
		case "Top":
			if c, ok := imp.component.(interface{ SetTop(float32) }); ok {
				c.SetTop(sizeToPixels(node.Text))
			}
		case "Left":
			if c, ok := imp.component.(interface{ SetLeft(float32) }); ok {
				c.SetLeft(sizeToPixels(node.Text))
			}
		case "Height":
			if c, ok := imp.component.(interface{ SetHeight(float32) }); ok {
				c.SetHeight(sizeToPixels(node.Text))
			}
		case "Width":
			if c, ok := imp.component.(interface{ SetWidth(float32) }); ok {
				c.SetWidth(sizeToPixels(node.Text))
			}
		case "Visibility":
			imp.loadVisibility(node)
		case "Style":
			imp.loadStyle(node)
		}
	}
	// When inside a table cell, override dimensions with the cell's own.
	// C# ref: "if (parent is TableCell) { component.Width/Height = cell.W/H }"
	if tcp, ok := imp.parent.(*tableCellParent); ok {
		if c, ok := imp.component.(interface {
			SetWidth(float32)
			SetHeight(float32)
		}); ok {
			c.SetWidth(tcp.cell.Width())
			c.SetHeight(tcp.cell.Height())
		}
	}
}

// loadVisibility reads <Visibility> and sets component.Visible.
// C# ref: RDLImport.LoadVisibility
func (imp *RDLImport) loadVisibility(node *xmlNode) {
	for _, child := range node.Children {
		if child.Name == "Hidden" {
			if c, ok := imp.component.(interface{ SetVisible(bool) }); ok {
				c.SetVisible(!booleanToBool(child.Text))
			}
		}
	}
}

// loadStyle applies font, alignment, border and padding properties.
// C# ref: RDLImport.LoadStyle
func (imp *RDLImport) loadStyle(styleNode *xmlNode) {
	fontStyle := style.FontStyleRegular
	fontFamily := "Arial"
	fontSize := float32(10)
	paddingTop := float32(0)
	paddingLeft := float32(2)
	paddingRight := float32(2)
	paddingBottom := float32(0)

	for _, node := range styleNode.Children {
		switch {
		case strings.HasSuffix(node.Name, "Border"):
			imp.loadBorder(node)
		case node.Name == "BackgroundColor":
			col := convertColor(node.Text)
			if shape, ok := imp.component.(*object.ShapeObject); ok {
				shape.SetFill(&style.SolidFill{Color: col})
			} else if tbl, ok := imp.component.(*table.TableObject); ok {
				tbl.SetFill(&style.SolidFill{Color: col})
			}
		case node.Name == "FontStyle":
			fontStyle = convertFontStyle(node.Text)
		case node.Name == "FontFamily":
			fontFamily = node.Text
		case node.Name == "FontSize":
			fontSize = convertFontSize(node.Text)
		case node.Name == "TextAlign":
			if txt, ok := imp.component.(*object.TextObject); ok {
				txt.SetHorzAlign(convertTextAlign(node.Text))
			}
		case node.Name == "VerticalAlign":
			if txt, ok := imp.component.(*object.TextObject); ok {
				txt.SetVertAlign(convertVerticalAlign(node.Text))
			}
		case node.Name == "WritingMode":
			if txt, ok := imp.component.(*object.TextObject); ok {
				txt.SetAngle(convertWritingMode(node.Text))
			}
		case node.Name == "Color":
			if txt, ok := imp.component.(*object.TextObject); ok {
				txt.SetTextColor(convertColor(node.Text))
			}
		case node.Name == "PaddingLeft":
			paddingLeft = float32(sizeToInt(node.Text, sizeUnitPt))
		case node.Name == "PaddingRight":
			paddingRight = float32(sizeToInt(node.Text, sizeUnitPt))
		case node.Name == "PaddingTop":
			paddingTop = float32(sizeToInt(node.Text, sizeUnitPt))
		case node.Name == "PaddingBottom":
			paddingBottom = float32(sizeToInt(node.Text, sizeUnitPt))
		}
	}

	if txt, ok := imp.component.(*object.TextObject); ok {
		txt.SetFont(style.Font{Name: fontFamily, Size: fontSize, Style: fontStyle})
		txt.SetPadding(object.Padding{
			Left:   paddingLeft,
			Top:    paddingTop,
			Right:  paddingRight,
			Bottom: paddingBottom,
		})
	} else if pic, ok := imp.component.(*object.PictureObject); ok {
		pic.SetPadding(object.Padding{
			Left:   paddingLeft,
			Top:    paddingTop,
			Right:  paddingRight,
			Bottom: paddingBottom,
		})
	}
}

// borderOwner is the interface for objects that have a settable border.
type borderOwner interface {
	Border() style.Border
	SetBorder(style.Border)
}

// loadBorder processes a border element.
// C# ref: RDLImport.LoadBorder
func (imp *RDLImport) loadBorder(borderNode *xmlNode) {
	side := "default"
	switch borderNode.Name {
	case "TopBorder":
		side = "Top"
	case "BottomBorder":
		side = "Bottom"
	case "LeftBorder":
		side = "Left"
	case "RightBorder":
		side = "Right"
	}
	for _, node := range borderNode.Children {
		switch node.Name {
		case "Color":
			imp.loadBorderColor(node, side)
		case "Style":
			imp.loadBorderStyle(node, side)
		case "Width":
			imp.loadBorderWidth(node, side)
		}
	}
}

// loadBorderColor sets the color of a border side.
// C# ref: RDLImport.LoadBorderColor
func (imp *RDLImport) loadBorderColor(node *xmlNode, side string) {
	col := convertColor(node.Text)
	bo, ok := imp.component.(borderOwner)
	if !ok {
		return
	}
	b := bo.Border()
	switch side {
	case "Default", "default":
		b.SetColor(col)
	case "Top":
		b.Top().Color = col
	case "Left":
		b.Left().Color = col
	case "Right":
		b.Right().Color = col
	case "Bottom":
		b.Bottom().Color = col
	}
	bo.SetBorder(b)
}

// loadBorderStyle sets the line style for a border side.
// C# ref: RDLImport.LoadBorderStyle
func (imp *RDLImport) loadBorderStyle(node *xmlNode, side string) {
	if node.Text == "None" {
		return
	}
	ls := convertBorderStyle(node.Text)
	bo, ok := imp.component.(borderOwner)
	if !ok {
		return
	}
	b := bo.Border()
	switch side {
	case "Default", "default":
		b.VisibleLines = style.BorderLinesAll
		b.SetLineStyle(ls)
	case "Top":
		b.VisibleLines |= style.BorderLinesTop
		b.Top().Style = ls
	case "Left":
		b.VisibleLines |= style.BorderLinesLeft
		b.Left().Style = ls
	case "Right":
		b.VisibleLines |= style.BorderLinesRight
		b.Right().Style = ls
	case "Bottom":
		b.VisibleLines |= style.BorderLinesBottom
		b.Bottom().Style = ls
	}
	bo.SetBorder(b)
}

// loadBorderWidth sets the line width for a border side.
// C# ref: RDLImport.LoadBorderWidth
func (imp *RDLImport) loadBorderWidth(node *xmlNode, side string) {
	w := sizeToPixels(node.Text)
	bo, ok := imp.component.(borderOwner)
	if !ok {
		return
	}
	b := bo.Border()
	switch side {
	case "Default", "default":
		b.SetWidth(w)
	case "Top":
		b.Top().Width = w
	case "Left":
		b.Left().Width = w
	case "Right":
		b.Right().Width = w
	case "Bottom":
		b.Bottom().Width = w
	}
	bo.SetBorder(b)
}

// ── Paragraphs / TextRuns ─────────────────────────────────────────────────────

// loadParagraphs assembles rich-text runs into the TextObject.
// C# ref: RDLImport.LoadParagraphs
func (imp *RDLImport) loadParagraphs(node *xmlNode, txt *object.TextObject) {
	imp.firstRun = true
	for _, para := range node.Children {
		if para.Name != "Paragraph" {
			continue
		}
		for _, paraChild := range para.Children {
			if paraChild.Name != "TextRuns" {
				continue
			}
			for _, run := range paraChild.Children {
				imp.parseTextRun(run, txt)
			}
		}
		if imp.firstRun {
			imp.firstRun = false
		}
	}
}

// parseTextRun handles a single <TextRun>.
// C# ref: RDLImport.ParseTextRun
func (imp *RDLImport) parseTextRun(node *xmlNode, txt *object.TextObject) {
	if node.Name != "TextRun" {
		return
	}
	for _, child := range node.Children {
		switch child.Name {
		case "Value":
			imp.parseTextBoxValue(child, txt)
		case "Style":
			imp.parseTextBoxStyle(child, txt)
		}
	}
}

// parseTextBoxValue appends a run value to the TextObject.
// C# ref: RDLImport.ParseTextBoxValue
func (imp *RDLImport) parseTextBoxValue(node *xmlNode, txt *object.TextObject) {
	if !imp.firstRun {
		txt.SetText(txt.Text() + "\r\n")
	}
	txt.SetText(txt.Text() + imp.getValue(node.Text))
}

// parseTextBoxStyle applies per-run font attributes.
// C# ref: RDLImport.ParseTextBoxStyle
func (imp *RDLImport) parseTextBoxStyle(node *xmlNode, txt *object.TextObject) {
	fs := txt.Font().Style
	family := ""
	size := float32(0)

	for _, child := range node.Children {
		switch child.Name {
		case "FontFamily":
			family = child.Text
		case "FontSize":
			s := strings.TrimSpace(strings.Replace(child.Text, "pt", "", 1))
			if f, err := strconv.ParseFloat(s, 32); err == nil {
				size = float32(f)
			}
		case "FontWeight":
			if child.Text == "Bold" {
				fs |= style.FontStyleBold
			}
		case "FontStyle":
			if child.Text == "Italic" {
				fs |= style.FontStyleItalic
			}
		case "TextDecoration":
			if child.Text == "Underline" {
				fs |= style.FontStyleUnderline
			}
		case "Color":
			txt.SetTextColor(convertColor(child.Text))
		}
	}

	if family == "" {
		family = imp.defaultFontFamily
	}
	current := txt.Font()
	switch {
	case family == "" && size == 0:
		current.Style = fs
	case family == "" && size != 0:
		current.Size = size
		current.Style = fs
	case family != "" && size == 0:
		current.Name = family
		current.Style = fs
	default:
		current.Name = family
		current.Size = size
		current.Style = fs
	}
	txt.SetFont(current)
}

// getValue converts an RDL field expression to a FastReport bracket expression.
// C# ref: RDLImport.GetValue
func (imp *RDLImport) getValue(rdlValue string) string {
	if rdlValue == "" || rdlValue[0] != '=' {
		return rdlValue
	}
	if strings.HasPrefix(rdlValue[1:], "Fields") {
		excl := strings.Index(rdlValue, "!")
		if excl < 0 {
			return rdlValue
		}
		remainder := rdlValue[excl+1:]
		dot := strings.Index(remainder, ".")
		if dot < 0 {
			return rdlValue
		}
		return "[" + imp.dataSetName + "." + remainder[:dot] + "]"
	}
	return rdlValue
}

// ── Table/Matrix loaders (from ImportTable.cs) ────────────────────────────────

// loadTable processes <Table> or non-matrix <Tablix>.
// C# ref: RDLImport.LoadTable
func (imp *RDLImport) loadTable(node *xmlNode) {
	name := node.AttrOrDefault("Name", "")
	tbl := importpkg.CreateTableObject(name, imp.parent)
	imp.component = tbl

	imp.loadReportItem(node.Children)

	var tableColumnsNode, headerNode, detailsNode, footerNode, tableRowsNode *xmlNode
	for _, child := range node.Children {
		switch child.Name {
		case "TableColumns":
			tableColumnsNode = child
		case "Header":
			headerNode = child
		case "Details":
			detailsNode = child
		case "Footer":
			footerNode = child
		case "TablixBody":
			for _, bodyChild := range child.Children {
				switch bodyChild.Name {
				case "TablixColumns":
					tableColumnsNode = bodyChild
				case "TablixRows":
					tableRowsNode = child
				}
			}
		}
	}

	imp.loadTableColumns(tableColumnsNode, tbl)
	if headerNode != nil {
		imp.loadTableHeader(headerNode, tbl)
	} else if tableRowsNode != nil {
		imp.loadTableHeader(tableRowsNode, tbl)
	}
	imp.loadTableDetails(detailsNode, tbl)
	imp.loadTableFooter(footerNode, tbl)
}

// loadTableColumns processes <TableColumns> / <TablixColumns>.
// C# ref: RDLImport.LoadTableColumns
func (imp *RDLImport) loadTableColumns(node *xmlNode, tbl *table.TableObject) {
	if node == nil {
		return
	}
	for _, child := range node.Children {
		if child.Name != "TableColumn" && child.Name != "TablixColumn" {
			continue
		}
		col := table.NewTableColumn()
		tbl.AddColumn(col)
		for _, item := range child.Children {
			if item.Name == "Width" {
				col.SetWidth(sizeToPixels(item.Text))
			}
		}
	}
}

// loadTableRows processes <TableRows> / <TablixRows>.
// C# ref: RDLImport.LoadTableRows
func (imp *RDLImport) loadTableRows(node *xmlNode, tbl *table.TableObject) {
	if node == nil {
		return
	}
	for _, child := range node.Children {
		if child.Name != "TableRow" && child.Name != "TablixRow" {
			continue
		}
		row := table.NewTableRow()
		tbl.AddRow(row)
		for _, item := range child.Children {
			switch item.Name {
			case "Height":
				row.SetHeight(sizeToPixels(item.Text))
			case "TableCells", "TablixCells":
				imp.loadTableCells(item, tbl)
			}
		}
	}
}

// loadTableCells processes cells for the most-recently added row.
// C# ref: RDLImport.LoadTableCells
func (imp *RDLImport) loadTableCells(node *xmlNode, tbl *table.TableObject) {
	col := 0
	rowIdx := len(tbl.Rows()) - 1
	if rowIdx < 0 {
		return
	}
	for _, child := range node.Children {
		if child.Name != "TableCell" && child.Name != "TablixCell" {
			continue
		}
		for _, item := range child.Children {
			switch item.Name {
			case "ReportItems", "CellContents":
				savedParent := imp.parent
				savedComp := imp.component
				cell := tbl.Cell(rowIdx, col)
				if cell != nil {
					imp.parent = &tableCellParent{cell: cell}
				}
				imp.loadReportItems(item)
				imp.component = savedComp
				imp.parent = savedParent
			case "ColSpan":
				if n, err := strconv.Atoi(strings.TrimSpace(item.Text)); err == nil && n > 1 {
					cell := tbl.Cell(rowIdx, col)
					if cell != nil {
						cell.SetColSpan(n)
					}
					col += n - 1
				}
			}
		}
		col++
	}
}

// loadTableHeader processes a <Header> rows container.
// C# ref: RDLImport.LoadHeader
func (imp *RDLImport) loadTableHeader(node *xmlNode, tbl *table.TableObject) {
	if node == nil {
		return
	}
	for _, child := range node.Children {
		if child.Name == "TableRows" || child.Name == "TablixRows" {
			imp.loadTableRows(child, tbl)
		}
	}
}

// loadTableDetails processes a <Details> element.
// C# ref: RDLImport.LoadDetails
func (imp *RDLImport) loadTableDetails(node *xmlNode, tbl *table.TableObject) {
	if node == nil {
		return
	}
	for _, child := range node.Children {
		if child.Name == "TableRows" {
			imp.loadTableRows(child, tbl)
		}
	}
}

// loadTableFooter processes a <Footer> element.
// C# ref: RDLImport.LoadFooter
func (imp *RDLImport) loadTableFooter(node *xmlNode, tbl *table.TableObject) {
	if node == nil {
		return
	}
	for _, child := range node.Children {
		if child.Name == "TableRows" {
			imp.loadTableRows(child, tbl)
		}
	}
}

// loadMatrix processes a <Matrix> or matrix-type <Tablix>.
// Full grouping is intentionally minimal — the C# source also leaves many
// matrix sections commented out. C# ref: RDLImport.LoadMatrix (ImportTable.cs).
func (imp *RDLImport) loadMatrix(node *xmlNode) {
	name := node.AttrOrDefault("Name", "")
	mx := importpkg.CreateMatrixObject(name, imp.parent)
	imp.component = mx
	imp.loadReportItem(node.Children)
}

// isTablixMatrix returns true when a <Tablix> has a <TablixCorner> child.
// C# ref: RDLImport.IsTablixMatrix
func isTablixMatrix(node *xmlNode) bool {
	for _, child := range node.Children {
		if child.Name == "TablixCorner" {
			return true
		}
	}
	return false
}

// ── Minimal XML DOM ───────────────────────────────────────────────────────────

// xmlNode is a lightweight DOM element.
type xmlNode struct {
	Name     string
	Text     string
	Attrs    map[string]string
	Children []*xmlNode
}

// AttrOrDefault returns the named attribute value or def.
func (n *xmlNode) AttrOrDefault(name, def string) string {
	if v, ok := n.Attrs[name]; ok {
		return v
	}
	return def
}

// HasChildren returns true if the node has any child elements.
func (n *xmlNode) HasChildren() bool { return len(n.Children) > 0 }

// FirstChild returns the first child, or nil.
func (n *xmlNode) FirstChild() *xmlNode {
	if len(n.Children) == 0 {
		return nil
	}
	return n.Children[0]
}

// FirstChildNamed returns the first child with the given element name.
func (n *xmlNode) FirstChildNamed(name string) *xmlNode {
	for _, c := range n.Children {
		if c.Name == name {
			return c
		}
	}
	return nil
}

// parseXML reads r and returns the document root as an xmlNode.
func parseXML(r io.Reader) (*xmlNode, error) {
	dec := xml.NewDecoder(r)
	var stack []*xmlNode
	var root *xmlNode

	for {
		tok, err := dec.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
		switch t := tok.(type) {
		case xml.StartElement:
			node := &xmlNode{
				Name:  t.Name.Local,
				Attrs: make(map[string]string, len(t.Attr)),
			}
			for _, a := range t.Attr {
				node.Attrs[a.Name.Local] = a.Value
			}
			if len(stack) > 0 {
				parent := stack[len(stack)-1]
				parent.Children = append(parent.Children, node)
			}
			stack = append(stack, node)
		case xml.EndElement:
			if len(stack) == 0 {
				break
			}
			top := stack[len(stack)-1]
			stack = stack[:len(stack)-1]
			if len(stack) == 0 {
				root = top
			}
		case xml.CharData:
			if len(stack) > 0 {
				text := strings.TrimSpace(string(t))
				if text != "" {
					stack[len(stack)-1].Text = text
				}
			}
		}
	}
	if root == nil {
		return nil, fmt.Errorf("rdlc: empty or invalid XML document")
	}
	return root, nil
}
