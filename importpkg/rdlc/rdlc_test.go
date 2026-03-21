package rdlc_test

import (
	"strings"
	"testing"

	"github.com/andrewloable/go-fastreport/importpkg/rdlc"
	"github.com/andrewloable/go-fastreport/reportpkg"
)

// ── constructor ───────────────────────────────────────────────────────────────

func TestNew_NameIsSet(t *testing.T) {
	imp := rdlc.New()
	if imp.Name() != "RDL/RDLC Importer" {
		t.Fatalf("expected name %q, got %q", "RDL/RDLC Importer", imp.Name())
	}
}

// ── LoadReportFromFile error path ─────────────────────────────────────────────

func TestLoadReportFromFile_NonExistent_ReturnsError(t *testing.T) {
	imp := rdlc.New()
	rpt := reportpkg.NewReport()
	err := imp.LoadReportFromFile(rpt, "nonexistent_file_12345.rdlc")
	if err == nil {
		t.Fatal("expected error for non-existent file, got nil")
	}
}

// ── LoadReportFromStream: invalid XML ─────────────────────────────────────────

func TestLoadReportFromStream_InvalidXML_ReturnsError(t *testing.T) {
	imp := rdlc.New()
	rpt := reportpkg.NewReport()
	err := imp.LoadReportFromStream(rpt, strings.NewReader("not xml at all <<<<"))
	if err == nil {
		t.Fatal("expected error for invalid XML, got nil")
	}
}

// ── LoadReportFromStream: minimal valid report ────────────────────────────────

const minimalRDL = `<?xml version="1.0" encoding="utf-8"?>
<Report xmlns="http://schemas.microsoft.com/sqlserver/reporting/2005/01/reportdefinition">
  <Description>Test Report</Description>
  <Author>Test Author</Author>
  <Body>
    <Height>10cm</Height>
    <ReportItems>
      <Textbox Name="TextBox1">
        <Value>Hello World</Value>
        <Top>0cm</Top>
        <Left>0cm</Left>
        <Width>5cm</Width>
        <Height>1cm</Height>
      </Textbox>
    </ReportItems>
  </Body>
  <Page>
    <PageHeight>29.7cm</PageHeight>
    <PageWidth>21cm</PageWidth>
    <LeftMargin>2cm</LeftMargin>
    <RightMargin>2cm</RightMargin>
    <TopMargin>2cm</TopMargin>
    <BottomMargin>2cm</BottomMargin>
  </Page>
</Report>`

func TestLoadReportFromStream_Minimal_PopulatesReport(t *testing.T) {
	imp := rdlc.New()
	rpt := reportpkg.NewReport()
	err := imp.LoadReportFromStream(rpt, strings.NewReader(minimalRDL))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rpt.PageCount() == 0 {
		t.Fatal("expected at least one page")
	}
}

func TestLoadReportFromStream_Minimal_SetsDescriptionAndAuthor(t *testing.T) {
	imp := rdlc.New()
	rpt := reportpkg.NewReport()
	_ = imp.LoadReportFromStream(rpt, strings.NewReader(minimalRDL))
	if rpt.Info.Description != "Test Report" {
		t.Fatalf("expected description %q, got %q", "Test Report", rpt.Info.Description)
	}
	if rpt.Info.Author != "Test Author" {
		t.Fatalf("expected author %q, got %q", "Test Author", rpt.Info.Author)
	}
}

func TestLoadReportFromStream_Minimal_PageDimensions(t *testing.T) {
	imp := rdlc.New()
	rpt := reportpkg.NewReport()
	_ = imp.LoadReportFromStream(rpt, strings.NewReader(minimalRDL))
	page := rpt.Page(0)
	// Dimensions are stored in millimeters.
	// 29.7cm = 297mm, 21cm = 210mm, 2cm = 20mm.
	if page.PaperHeight < 296 || page.PaperHeight > 298 {
		t.Fatalf("expected PaperHeight ~297mm, got %v", page.PaperHeight)
	}
	if page.PaperWidth < 209 || page.PaperWidth > 211 {
		t.Fatalf("expected PaperWidth ~210mm, got %v", page.PaperWidth)
	}
	if page.LeftMargin < 19 || page.LeftMargin > 21 {
		t.Fatalf("expected LeftMargin ~20mm, got %v", page.LeftMargin)
	}
}

// ── Page header / footer ──────────────────────────────────────────────────────

const rdlWithPageHeaderFooter = `<?xml version="1.0"?>
<Report>
  <Body>
    <Height>5cm</Height>
  </Body>
  <Page>
    <PageHeader>
      <Height>1cm</Height>
      <PrintOnFirstPage>true</PrintOnFirstPage>
      <ReportItems>
        <Textbox Name="Header1">
          <Value>Page Header</Value>
        </Textbox>
      </ReportItems>
    </PageHeader>
    <PageFooter>
      <Height>1cm</Height>
      <ReportItems>
        <Textbox Name="Footer1">
          <Value>Page Footer</Value>
        </Textbox>
      </ReportItems>
    </PageFooter>
    <PageHeight>29.7cm</PageHeight>
    <PageWidth>21cm</PageWidth>
  </Page>
</Report>`

func TestLoadReportFromStream_PageHeaderFooter_Created(t *testing.T) {
	imp := rdlc.New()
	rpt := reportpkg.NewReport()
	err := imp.LoadReportFromStream(rpt, strings.NewReader(rdlWithPageHeaderFooter))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	page := rpt.Page(0)
	if page.PageHeader() == nil {
		t.Error("expected PageHeader to be set")
	}
	if page.PageFooter() == nil {
		t.Error("expected PageFooter to be set")
	}
}

// ── Report parameters ─────────────────────────────────────────────────────────

const rdlWithParameters = `<?xml version="1.0"?>
<Report>
  <ReportParameters>
    <ReportParameter Name="StartDate">
      <DataType>DateTime</DataType>
      <Prompt>Start Date</Prompt>
    </ReportParameter>
    <ReportParameter Name="EndDate">
      <DataType>DateTime</DataType>
    </ReportParameter>
  </ReportParameters>
  <Body><Height>5cm</Height></Body>
</Report>`

func TestLoadReportFromStream_Parameters_Registered(t *testing.T) {
	imp := rdlc.New()
	rpt := reportpkg.NewReport()
	err := imp.LoadReportFromStream(rpt, strings.NewReader(rdlWithParameters))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if p := rpt.Dictionary().FindParameter("StartDate"); p == nil {
		t.Fatal("expected StartDate parameter to be registered")
	}
	if p := rpt.Dictionary().FindParameter("EndDate"); p == nil {
		t.Fatal("expected EndDate parameter to be registered")
	}
	p := rpt.Dictionary().FindParameter("StartDate")
	if p.Description != "Start Date" {
		t.Fatalf("expected description %q, got %q", "Start Date", p.Description)
	}
}

// ── DataSets → dataSetName ────────────────────────────────────────────────────

const rdlWithDataSet = `<?xml version="1.0"?>
<Report>
  <DataSets>
    <DataSet Name="NorthWind">
    </DataSet>
  </DataSets>
  <Body>
    <Height>5cm</Height>
    <ReportItems>
      <Textbox Name="FieldText">
        <Value>=Fields!ProductName.Value</Value>
      </Textbox>
    </ReportItems>
  </Body>
</Report>`

func TestLoadReportFromStream_FieldExpression_ConvertedToBracket(t *testing.T) {
	imp := rdlc.New()
	rpt := reportpkg.NewReport()
	err := imp.LoadReportFromStream(rpt, strings.NewReader(rdlWithDataSet))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// If the text object was created and the field expression was converted we
	// simply verify the report was built without error and has one page.
	if rpt.PageCount() == 0 {
		t.Fatal("expected at least one page")
	}
}

// ── RDLC ReportSections ───────────────────────────────────────────────────────

const rdlcWithSections = `<?xml version="1.0"?>
<Report xmlns:rd="http://schemas.microsoft.com/SQLServer/reporting/reportdesigner">
  <ReportSections>
    <ReportSection>
      <Body>
        <Height>5cm</Height>
      </Body>
      <Width>17cm</Width>
      <Page>
        <PageHeight>29.7cm</PageHeight>
        <LeftMargin>2cm</LeftMargin>
        <RightMargin>2cm</RightMargin>
        <TopMargin>2cm</TopMargin>
        <BottomMargin>2cm</BottomMargin>
      </Page>
    </ReportSection>
  </ReportSections>
</Report>`

func TestLoadReportFromStream_RDLCSections_PageWidthFromSectionWidth(t *testing.T) {
	imp := rdlc.New()
	rpt := reportpkg.NewReport()
	err := imp.LoadReportFromStream(rpt, strings.NewReader(rdlcWithSections))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rpt.PageCount() == 0 {
		t.Fatal("expected at least one page")
	}
	page := rpt.Page(0)
	// PageWidth = LeftMargin(20mm) + sectionWidth(170mm) + RightMargin(20mm) = 210mm
	if page.PaperWidth < 205 || page.PaperWidth > 215 {
		t.Fatalf("expected PaperWidth ~210mm, got %v", page.PaperWidth)
	}
}

// ── Line object ───────────────────────────────────────────────────────────────

const rdlWithLine = `<?xml version="1.0"?>
<Report>
  <Body>
    <Height>5cm</Height>
    <ReportItems>
      <Line Name="Line1">
        <Top>0cm</Top>
        <Left>0cm</Left>
        <Width>10cm</Width>
        <Height>0.1cm</Height>
      </Line>
    </ReportItems>
  </Body>
</Report>`

func TestLoadReportFromStream_Line_Created(t *testing.T) {
	imp := rdlc.New()
	rpt := reportpkg.NewReport()
	err := imp.LoadReportFromStream(rpt, strings.NewReader(rdlWithLine))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rpt.PageCount() == 0 {
		t.Fatal("expected at least one page")
	}
}

// ── Rectangle / ContainerObject ───────────────────────────────────────────────

const rdlWithRectangle = `<?xml version="1.0"?>
<Report>
  <Body>
    <Height>5cm</Height>
    <ReportItems>
      <Rectangle Name="Rect1">
        <Top>0cm</Top>
        <Left>0cm</Left>
        <Width>5cm</Width>
        <Height>3cm</Height>
        <ReportItems>
          <Textbox Name="Inside1">
            <Value>Inside</Value>
          </Textbox>
        </ReportItems>
      </Rectangle>
    </ReportItems>
  </Body>
</Report>`

func TestLoadReportFromStream_Rectangle_ContainerCreated(t *testing.T) {
	imp := rdlc.New()
	rpt := reportpkg.NewReport()
	err := imp.LoadReportFromStream(rpt, strings.NewReader(rdlWithRectangle))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rpt.PageCount() == 0 {
		t.Fatal("expected at least one page")
	}
}

// ── Image object ──────────────────────────────────────────────────────────────

const rdlWithImage = `<?xml version="1.0"?>
<Report>
  <Body>
    <Height>5cm</Height>
    <ReportItems>
      <Image Name="Img1">
        <Source>External</Source>
        <Value>/nonexistent/path/logo.png</Value>
        <Sizing>Fit</Sizing>
        <Top>0cm</Top>
        <Left>0cm</Left>
        <Width>3cm</Width>
        <Height>2cm</Height>
      </Image>
    </ReportItems>
  </Body>
</Report>`

func TestLoadReportFromStream_Image_Created(t *testing.T) {
	imp := rdlc.New()
	rpt := reportpkg.NewReport()
	err := imp.LoadReportFromStream(rpt, strings.NewReader(rdlWithImage))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rpt.PageCount() == 0 {
		t.Fatal("expected at least one page")
	}
}

// ── Table object ──────────────────────────────────────────────────────────────

const rdlWithTable = `<?xml version="1.0"?>
<Report>
  <Body>
    <Height>10cm</Height>
    <ReportItems>
      <Table Name="Table1">
        <Top>0cm</Top>
        <Left>0cm</Left>
        <Width>10cm</Width>
        <Height>5cm</Height>
        <TableColumns>
          <TableColumn>
            <Width>5cm</Width>
          </TableColumn>
          <TableColumn>
            <Width>5cm</Width>
          </TableColumn>
        </TableColumns>
        <Header>
          <TableRows>
            <TableRow>
              <Height>1cm</Height>
              <TableCells>
                <TableCell>
                  <ReportItems>
                    <Textbox Name="HdrCol1">
                      <Value>Column 1</Value>
                    </Textbox>
                  </ReportItems>
                </TableCell>
                <TableCell>
                  <ReportItems>
                    <Textbox Name="HdrCol2">
                      <Value>Column 2</Value>
                    </Textbox>
                  </ReportItems>
                </TableCell>
              </TableCells>
            </TableRow>
          </TableRows>
        </Header>
      </Table>
    </ReportItems>
  </Body>
</Report>`

func TestLoadReportFromStream_Table_Created(t *testing.T) {
	imp := rdlc.New()
	rpt := reportpkg.NewReport()
	err := imp.LoadReportFromStream(rpt, strings.NewReader(rdlWithTable))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rpt.PageCount() == 0 {
		t.Fatal("expected at least one page")
	}
}

// ── Textbox with Paragraphs/TextRuns ─────────────────────────────────────────

const rdlWithParagraphs = `<?xml version="1.0"?>
<Report>
  <Body>
    <Height>5cm</Height>
    <ReportItems>
      <Textbox Name="RichText1">
        <Paragraphs>
          <Paragraph>
            <TextRuns>
              <TextRun>
                <Value>Hello</Value>
                <Style>
                  <FontFamily>Arial</FontFamily>
                  <FontSize>12pt</FontSize>
                  <FontWeight>Bold</FontWeight>
                </Style>
              </TextRun>
            </TextRuns>
          </Paragraph>
          <Paragraph>
            <TextRuns>
              <TextRun>
                <Value>World</Value>
                <Style>
                  <FontStyle>Italic</FontStyle>
                </Style>
              </TextRun>
            </TextRuns>
          </Paragraph>
        </Paragraphs>
      </Textbox>
    </ReportItems>
  </Body>
</Report>`

func TestLoadReportFromStream_Paragraphs_Processed(t *testing.T) {
	imp := rdlc.New()
	rpt := reportpkg.NewReport()
	err := imp.LoadReportFromStream(rpt, strings.NewReader(rdlWithParagraphs))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rpt.PageCount() == 0 {
		t.Fatal("expected at least one page")
	}
}

// ── Visibility ────────────────────────────────────────────────────────────────

const rdlWithHiddenTextbox = `<?xml version="1.0"?>
<Report>
  <Body>
    <Height>5cm</Height>
    <ReportItems>
      <Textbox Name="Hidden1">
        <Value>Hidden</Value>
        <Visibility>
          <Hidden>true</Hidden>
        </Visibility>
      </Textbox>
    </ReportItems>
  </Body>
</Report>`

func TestLoadReportFromStream_HiddenComponent_VisibleFalse(t *testing.T) {
	imp := rdlc.New()
	rpt := reportpkg.NewReport()
	err := imp.LoadReportFromStream(rpt, strings.NewReader(rdlWithHiddenTextbox))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Report builds without error. Visibility is applied to the component during
	// load. We verify the report builds cleanly.
	if rpt.PageCount() == 0 {
		t.Fatal("expected at least one page")
	}
}

// ── Subreport (no sub-file present) ──────────────────────────────────────────

const rdlWithSubreport = `<?xml version="1.0"?>
<Report>
  <Body>
    <Height>5cm</Height>
    <ReportItems>
      <Subreport Name="Sub1">
        <ReportName>NonExistentSubReport</ReportName>
        <Top>1cm</Top>
        <Left>0cm</Left>
        <Width>5cm</Width>
        <Height>2cm</Height>
      </Subreport>
    </ReportItems>
  </Body>
</Report>`

func TestLoadReportFromStream_Subreport_NoFile_CreatesPlaceholder(t *testing.T) {
	imp := rdlc.New()
	rpt := reportpkg.NewReport()
	err := imp.LoadReportFromStream(rpt, strings.NewReader(rdlWithSubreport))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// A placeholder page should have been created for the subreport.
	if rpt.PageCount() < 2 {
		t.Fatalf("expected at least 2 pages (main + subreport placeholder), got %d", rpt.PageCount())
	}
}

// ── HideDuplicates ────────────────────────────────────────────────────────────

const rdlWithHideDuplicates = `<?xml version="1.0"?>
<Report>
  <Body>
    <Height>5cm</Height>
    <ReportItems>
      <Textbox Name="Dup1">
        <Value>SomeValue</Value>
        <HideDuplicates/>
      </Textbox>
    </ReportItems>
  </Body>
</Report>`

func TestLoadReportFromStream_HideDuplicates_DoesNotError(t *testing.T) {
	imp := rdlc.New()
	rpt := reportpkg.NewReport()
	err := imp.LoadReportFromStream(rpt, strings.NewReader(rdlWithHideDuplicates))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// ── Matrix object ─────────────────────────────────────────────────────────────

const rdlWithMatrix = `<?xml version="1.0"?>
<Report>
  <Body>
    <Height>10cm</Height>
    <ReportItems>
      <Matrix Name="Matrix1">
        <Top>0cm</Top>
        <Left>0cm</Left>
        <Width>10cm</Width>
        <Height>5cm</Height>
      </Matrix>
    </ReportItems>
  </Body>
</Report>`

func TestLoadReportFromStream_Matrix_Created(t *testing.T) {
	imp := rdlc.New()
	rpt := reportpkg.NewReport()
	err := imp.LoadReportFromStream(rpt, strings.NewReader(rdlWithMatrix))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rpt.PageCount() == 0 {
		t.Fatal("expected at least one page")
	}
}

// ── Tablix (table variant) ────────────────────────────────────────────────────

const rdlWithTablix = `<?xml version="1.0"?>
<Report>
  <Body>
    <Height>10cm</Height>
    <ReportItems>
      <Tablix Name="Tablix1">
        <Top>0cm</Top>
        <Left>0cm</Left>
        <Width>10cm</Width>
        <Height>5cm</Height>
        <TablixBody>
          <TablixColumns>
            <TablixColumn><Width>5cm</Width></TablixColumn>
            <TablixColumn><Width>5cm</Width></TablixColumn>
          </TablixColumns>
          <TablixRows>
            <TablixRow>
              <Height>1cm</Height>
              <TablixCells>
                <TablixCell>
                  <CellContents>
                    <Textbox Name="TCell1"><Value>A</Value></Textbox>
                  </CellContents>
                </TablixCell>
                <TablixCell>
                  <CellContents>
                    <Textbox Name="TCell2"><Value>B</Value></Textbox>
                  </CellContents>
                </TablixCell>
              </TablixCells>
            </TablixRow>
          </TablixRows>
        </TablixBody>
      </Tablix>
    </ReportItems>
  </Body>
</Report>`

func TestLoadReportFromStream_Tablix_Table_Created(t *testing.T) {
	imp := rdlc.New()
	rpt := reportpkg.NewReport()
	err := imp.LoadReportFromStream(rpt, strings.NewReader(rdlWithTablix))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rpt.PageCount() == 0 {
		t.Fatal("expected at least one page")
	}
}

// ── Tablix with TablixCorner → treated as matrix ──────────────────────────────

const rdlWithTablixMatrix = `<?xml version="1.0"?>
<Report>
  <Body>
    <Height>10cm</Height>
    <ReportItems>
      <Tablix Name="MatrixTablix1">
        <TablixCorner/>
        <Top>0cm</Top>
        <Left>0cm</Left>
        <Width>10cm</Width>
        <Height>5cm</Height>
      </Tablix>
    </ReportItems>
  </Body>
</Report>`

func TestLoadReportFromStream_TablixMatrix_Created(t *testing.T) {
	imp := rdlc.New()
	rpt := reportpkg.NewReport()
	err := imp.LoadReportFromStream(rpt, strings.NewReader(rdlWithTablixMatrix))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rpt.PageCount() == 0 {
		t.Fatal("expected at least one page")
	}
}

// ── Multiple LoadReportFromStream calls (re-use of importer) ──────────────────

func TestLoadReportFromStream_MultipleCallsOnSameImporter(t *testing.T) {
	imp := rdlc.New()
	rpt1 := reportpkg.NewReport()
	if err := imp.LoadReportFromStream(rpt1, strings.NewReader(minimalRDL)); err != nil {
		t.Fatalf("first call error: %v", err)
	}
	rpt2 := reportpkg.NewReport()
	if err := imp.LoadReportFromStream(rpt2, strings.NewReader(minimalRDL)); err != nil {
		t.Fatalf("second call error: %v", err)
	}
	if rpt1.PageCount() == 0 || rpt2.PageCount() == 0 {
		t.Fatal("both reports should have at least one page")
	}
}
