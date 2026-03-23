package crossview

// CrossViewHelper holds geometry state used when building the cross-tab template
// or result layout. It is the Go equivalent of FastReport.CrossView.CrossViewHelper
// (CrossViewHelper.cs).
//
// The C# helper is heavily UI-dependent (it writes directly into TableResult /
// TableColumn / TableRow / TableCell objects backed by GDI+). The Go port
// focuses on the geometry calculation parts that are independent of UI:
// header/body width and height, and descriptor index resolution.
// Table-cell content building (PrintHeaders, PrintXAxisTemplate, etc.) is
// intentionally skipped because it depends on GDI+ drawing primitives that
// are not ported.
type CrossViewHelper struct {
	crossView *CrossViewObject

	// Geometry — populated by UpdateTemplateSizes / BuildTemplate.
	headerWidth  int
	headerHeight int
	bodyWidth    int
	bodyHeight   int

	// Result body dimensions — populated by StartPrint.
	resultBodyWidth  int
	resultBodyHeight int
}

// NewCrossViewHelper creates a CrossViewHelper associated with cv.
// Mirrors the CrossViewHelper constructor (CrossViewHelper.cs line 767-771).
func NewCrossViewHelper(cv *CrossViewObject) *CrossViewHelper {
	return &CrossViewHelper{crossView: cv}
}

// CrossView returns the associated CrossViewObject.
func (h *CrossViewHelper) CrossView() *CrossViewObject { return h.crossView }

// HeaderWidth returns the number of header columns (Y-axis field count, or 1
// when no source is assigned).
// Mirrors CrossViewHelper.HeaderWidth (CrossViewHelper.cs lines 46-49).
func (h *CrossViewHelper) HeaderWidth() int { return h.headerWidth }

// HeaderHeight returns the number of header rows including any title / caption
// rows.
// Mirrors CrossViewHelper.HeaderHeight (CrossViewHelper.cs lines 41-44).
func (h *CrossViewHelper) HeaderHeight() int { return h.headerHeight }

// TemplateBodyWidth returns the template body column count (one per terminal
// column descriptor, accounting for measures).
// Mirrors CrossViewHelper.TemplateBodyWidth (CrossViewHelper.cs lines 51-54).
func (h *CrossViewHelper) TemplateBodyWidth() int { return h.bodyWidth }

// TemplateBodyHeight returns the template body row count.
// Mirrors CrossViewHelper.TemplateBodyHeight (CrossViewHelper.cs lines 56-59).
func (h *CrossViewHelper) TemplateBodyHeight() int { return h.bodyHeight }

// ResultBodyWidth returns the result body column count computed during StartPrint.
// Mirrors CrossViewHelper.ResultBodyWidth (CrossViewHelper.cs lines 61-64).
func (h *CrossViewHelper) ResultBodyWidth() int { return h.resultBodyWidth }

// ResultBodyHeight returns the result body row count computed during StartPrint.
// Mirrors CrossViewHelper.ResultBodyHeight (CrossViewHelper.cs lines 66-69).
func (h *CrossViewHelper) ResultBodyHeight() int { return h.resultBodyHeight }

// UpdateTemplateSizes recomputes the header/body geometry from the current
// CrossViewObject state.
// Mirrors CrossViewHelper.UpdateTemplateSizes() (CrossViewHelper.cs lines 78-103).
func (h *CrossViewHelper) UpdateTemplateSizes() {
	cv := h.crossView
	data := &cv.Data

	if data.SourceAssigned() {
		h.headerWidth = data.YAxisFieldsCount()
		h.headerHeight = data.XAxisFieldsCount()
	} else {
		h.headerWidth = 1
		h.headerHeight = 1
	}

	if cv.ShowXAxisFieldsCaption {
		h.headerHeight++
	}
	if cv.ShowTitle {
		h.headerHeight++
	}
	// If headerHeight is still 0 and ShowYAxisFieldsCaption is set, ensure
	// at least 1 header row (matches C# line 94-96).
	if h.headerHeight == 0 && cv.ShowYAxisFieldsCaption {
		h.headerHeight = 1
	}

	// Template body dimensions: one cell per terminal descriptor, with an
	// extra multiplier when measures appear on the corresponding axis.
	h.bodyWidth = 1 + data.XAxisFieldsCount()
	if data.MeasuresInXAxis() {
		h.bodyWidth = (h.bodyWidth - 1) * data.MeasuresCount()
	}

	h.bodyHeight = 1 + data.YAxisFieldsCount()
	if data.MeasuresInYAxis() {
		h.bodyHeight = (h.bodyHeight - 1) * data.MeasuresCount()
	}
}

// UpdateDescriptors recomputes template geometry without rebuilding the full
// table. In the C# implementation this updates TemplateColumn / TemplateRow /
// TemplateCell back-references on each descriptor; in Go those references are
// not maintained (no GDI+ table), so this method only refreshes the size
// properties.
// Mirrors CrossViewHelper.UpdateDescriptors() (CrossViewHelper.cs lines 563-597).
func (h *CrossViewHelper) UpdateDescriptors() {
	h.UpdateTemplateSizes()
}

// BuildTemplate rebuilds the geometry for the cross-tab template from the
// current CrossViewObject state. It calls UpdateTemplateSizes and then
// UpdateDescriptors to synchronise all derived fields.
//
// The C# implementation additionally creates a TableResult, writes cell
// content, and copies it back into the CrossView table — that part is
// GDI+/UI-dependent and is not ported here.
// Mirrors CrossViewHelper.BuildTemplate() (CrossViewHelper.cs lines 483-561).
func (h *CrossViewHelper) BuildTemplate() {
	h.UpdateTemplateSizes()
	h.UpdateDescriptors()
}

// StartPrint initialises the result-body dimensions from the data source and
// resets the design-time flag. In the C# code this also creates the
// TableResult; in Go we only cache the result dimensions.
// Mirrors CrossViewHelper.StartPrint() (CrossViewHelper.cs lines 611-636).
func (h *CrossViewHelper) StartPrint() {
	data := &h.crossView.Data
	h.resultBodyHeight = data.DataRowCount()
	h.resultBodyWidth = data.DataColumnCount()
}

// FinishPrint synchronises descriptors after data has been loaded and fires
// any post-processing. In the C# implementation it also calls InitResultTable
// and the print-helper methods; those are UI-dependent and skipped here.
// Mirrors CrossViewHelper.FinishPrint() (CrossViewHelper.cs lines 646-671).
func (h *CrossViewHelper) FinishPrint() {
	h.UpdateDescriptors()
}

// AddData adds data from the bound cube source to the result.
// In the C# implementation this copies cube cell values into the TableResult;
// in Go, result building is deferred to Build() / buildGrid(). This stub
// satisfies the engine lifecycle contract.
// Mirrors CrossViewHelper.AddData() (CrossViewHelper.cs lines 638-644).
func (h *CrossViewHelper) AddData() {
	// No-op in the Go port: the result grid is built on demand via CrossViewObject.Build().
}

// CreateOtherDescriptor (re)initialises the internal "no source" placeholder
// descriptors. In the full C# implementation these are CrossViewDescriptor
// objects used in the template table; in Go we only reset the geometry state.
// Mirrors CrossViewHelper.CreateOtherDescriptor() (CrossViewHelper.cs lines 748-764).
func (h *CrossViewHelper) CreateOtherDescriptor() {
	// In the Go port, placeholder descriptors are not needed because we do
	// not maintain a backing TableResult. Reset geometry state so the next
	// BuildTemplate call starts from a clean slate.
	h.headerWidth = 0
	h.headerHeight = 0
	h.bodyWidth = 0
	h.bodyHeight = 0
}
