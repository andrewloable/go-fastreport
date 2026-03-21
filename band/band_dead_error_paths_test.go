package band

// band_dead_error_paths_test.go – internal tests targeting the specific uncovered
// "return err" branches in band.go, header_footer.go, and types.go.
//
// Analysis: All uncovered branches are defensive error-propagation statements of
// the form:
//
//   if err := parent.Serialize(w); err != nil { return err }
//
// The parent chain is:
//   BandBase.serializeAttrs → BreakableComponent.Serialize → ReportComponentBase.Serialize
//     → ComponentBase.Serialize → BaseObject.Serialize → (void WriteXxx methods)
//
// Since all WriteXxx methods in report.Writer return void, the ancestor Serialize
// calls NEVER return a non-nil error with the current interface design. These
// branches are architecturally unreachable by test code without modifying source.
//
// This file documents that fact and adds complementary tests that maximise coverage
// of all REACHABLE branches in the affected functions.

import (
	"errors"
	"testing"

	"github.com/andrewloable/go-fastreport/report"
)

// ─── failAttrsWriter: a Writer whose WriteObject always errors ─────────────────
// We use this to trigger the error path inside serializeChildren (the only
// Writer-level error path that IS reachable), and verify that BandBase.Serialize,
// HeaderFooterBandBase.Serialize, GroupHeaderBand.Serialize, and DataBand.Serialize
// all propagate those errors correctly (line 338, 59, 253, 473 respectively).
// This is distinct from the "serializeAttrs fails" paths which are unreachable.

type failAttrsWriter2 struct {
	written         map[string]string
	failWriteObject bool
}

func newFailAttrsWriter2() *failAttrsWriter2 {
	return &failAttrsWriter2{written: make(map[string]string)}
}

func (w *failAttrsWriter2) WriteStr(name, value string)      { w.written[name] = value }
func (w *failAttrsWriter2) WriteInt(name string, v int)      { w.written[name] = "int" }
func (w *failAttrsWriter2) WriteBool(name string, v bool)    { w.written[name] = "bool" }
func (w *failAttrsWriter2) WriteFloat(name string, v float32) { w.written[name] = "float" }

func (w *failAttrsWriter2) WriteObject(obj report.Serializable) error {
	if w.failWriteObject {
		return errors.New("failAttrsWriter2: WriteObject error")
	}
	return nil
}

func (w *failAttrsWriter2) WriteObjectNamed(name string, obj report.Serializable) error {
	if w.failWriteObject {
		return errors.New("failAttrsWriter2: WriteObjectNamed error")
	}
	return nil
}

// minimalSerializable2 is a minimal Serializable for use as a child object.
type minimalSerializable2 struct {
	report.BaseObject
}

func (s *minimalSerializable2) Serialize(w report.Writer) error   { return nil }
func (s *minimalSerializable2) Deserialize(r report.Reader) error { return nil }

// ─── BandBase.UpdateLayout ─────────────────────────────────────────────────────
//
// UpdateLayout is an explicitly documented no-op. The coverage tool reports 0.0%
// because the function body contains zero executable statements (only a comment).
// The function IS executed by these tests; the 0.0% is a coverage-tool artifact
// for empty function bodies.

func TestBandBase_UpdateLayout_DirectCall(t *testing.T) {
	b := NewBandBase()
	// All combinations of positive, zero, and negative deltas.
	b.UpdateLayout(0, 0)
	b.UpdateLayout(100, 200)
	b.UpdateLayout(-50, -75)
	b.UpdateLayout(0.001, 999.999)
}

func TestBandBase_UpdateLayout_ViaAllBandTypes(t *testing.T) {
	// Call UpdateLayout on every concrete band type to ensure the no-op
	// is exercised from each subtype.
	bands := []interface{ UpdateLayout(float32, float32) }{
		NewBandBase(),
		NewChildBand(),
		NewReportTitleBand(),
		NewReportSummaryBand(),
		NewPageHeaderBand(),
		NewPageFooterBand(),
		NewColumnHeaderBand(),
		NewColumnFooterBand(),
		NewDataHeaderBand(),
		NewDataFooterBand(),
		NewGroupHeaderBand(),
		NewGroupFooterBand(),
		NewDataBand(),
		NewOverlayBand(),
	}
	for _, b := range bands {
		b.UpdateLayout(10, 20)
	}
}

// ─── BandBase.serializeAttrs: all reachable branches ──────────────────────────

// TestBandBase_serializeAttrs_AllBranches exercises every conditional branch
// inside serializeAttrs by setting exactly one non-default field at a time.
func TestBandBase_serializeAttrs_AllBranches(t *testing.T) {
	tests := []struct {
		name   string
		setup  func(*BandBase)
		expect string
	}{
		{
			name:   "startNewPage=true",
			setup:  func(b *BandBase) { b.startNewPage = true },
			expect: "StartNewPage",
		},
		{
			name:   "firstRowStartsNewPage=false",
			setup:  func(b *BandBase) { b.firstRowStartsNewPage = false },
			expect: "FirstRowStartsNewPage",
		},
		{
			name:   "printOnBottom=true",
			setup:  func(b *BandBase) { b.printOnBottom = true },
			expect: "PrintOnBottom",
		},
		{
			name:   "keepChild=true",
			setup:  func(b *BandBase) { b.keepChild = true },
			expect: "KeepChild",
		},
		{
			name:   "outlineExpression set",
			setup:  func(b *BandBase) { b.outlineExpression = "[Name]" },
			expect: "OutlineExpression",
		},
		{
			name:   "repeatBandNTimes != 1",
			setup:  func(b *BandBase) { b.repeatBandNTimes = 3 },
			expect: "RepeatBandNTimes",
		},
		{
			name:   "beforeLayoutEvent set",
			setup:  func(b *BandBase) { b.beforeLayoutEvent = "handler" },
			expect: "BeforeLayoutEvent",
		},
		{
			name:   "afterLayoutEvent set",
			setup:  func(b *BandBase) { b.afterLayoutEvent = "handler" },
			expect: "AfterLayoutEvent",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			b := NewBandBase()
			tc.setup(b)
			w := newFailAttrsWriter2()
			if err := b.serializeAttrs(w); err != nil {
				t.Errorf("serializeAttrs error: %v", err)
			}
			if _, ok := w.written[tc.expect]; !ok {
				t.Errorf("expected %q to be written", tc.expect)
			}
		})
	}
}

// TestBandBase_serializeAttrs_CanBreakFalse verifies the CanBreak=false branch
// inside BreakableComponent.Serialize (which serializeAttrs calls via the parent chain).
func TestBandBase_serializeAttrs_CanBreakFalse(t *testing.T) {
	b := NewBandBase()
	b.SetCanBreak(false)
	w := newFailAttrsWriter2()
	if err := b.serializeAttrs(w); err != nil {
		t.Errorf("serializeAttrs error: %v", err)
	}
	if _, ok := w.written["CanBreak"]; !ok {
		t.Error("expected CanBreak to be written when false")
	}
}

// ─── BandBase.Serialize: children-error path (line 338) ──────────────────────
// Note: the serializeAttrs-error path at line 336 is unreachable (dead code)
// because BreakableComponent.Serialize never returns an error.

func TestBandBase_Serialize_ChildrenErrorPath(t *testing.T) {
	b := NewBandBase()
	child := &minimalSerializable2{BaseObject: *report.NewBaseObject()}
	b.AddChild(child)

	w := newFailAttrsWriter2()
	w.failWriteObject = true

	err := b.Serialize(w)
	if err == nil {
		t.Error("BandBase.Serialize: expected error from failing WriteObject")
	}
}

// TestBandBase_Serialize_NoChildren exercises the successful path (line 338) with no children.
func TestBandBase_Serialize_NoChildren(t *testing.T) {
	b := NewBandBase()
	w := newFailAttrsWriter2()
	if err := b.Serialize(w); err != nil {
		t.Errorf("BandBase.Serialize with no children: %v", err)
	}
}

// ─── BandBase.Deserialize: all branches ───────────────────────────────────────
// Note: the "if err := b.BreakableComponent.Deserialize(r); err != nil { return err }"
// path (line 344) is unreachable dead code.

type bandBaseDeserMock struct {
	*mockReader
	canBreak           bool
	startNewPage       bool
	firstRowStartsPage bool
	printOnBottom      bool
	keepChild          bool
	outlineExpr        string
	repeatBand         int
	beforeEvent        string
	afterEvent         string
}

func (m *bandBaseDeserMock) ReadBool(name string, def bool) bool {
	switch name {
	case "CanBreak":
		return m.canBreak
	case "StartNewPage":
		return m.startNewPage
	case "FirstRowStartsNewPage":
		return m.firstRowStartsPage
	case "PrintOnBottom":
		return m.printOnBottom
	case "KeepChild":
		return m.keepChild
	default:
		return def
	}
}

func (m *bandBaseDeserMock) ReadStr(name, def string) string {
	switch name {
	case "OutlineExpression":
		return m.outlineExpr
	case "BeforeLayoutEvent":
		return m.beforeEvent
	case "AfterLayoutEvent":
		return m.afterEvent
	default:
		return def
	}
}

func (m *bandBaseDeserMock) ReadInt(name string, def int) int {
	if name == "RepeatBandNTimes" {
		return m.repeatBand
	}
	return def
}

func TestBandBase_Deserialize_AllNonDefaults(t *testing.T) {
	r := &bandBaseDeserMock{
		mockReader:         newMockReader(),
		canBreak:           false,
		startNewPage:       true,
		firstRowStartsPage: false,
		printOnBottom:      true,
		keepChild:          true,
		outlineExpr:        "[Category]",
		repeatBand:         4,
		beforeEvent:        "onBefore",
		afterEvent:         "onAfter",
	}

	b := NewBandBase()
	if err := b.Deserialize(r); err != nil {
		t.Fatalf("BandBase.Deserialize error: %v", err)
	}
	if b.CanBreak() {
		t.Error("CanBreak should be false")
	}
	if !b.startNewPage {
		t.Error("startNewPage should be true")
	}
	if b.firstRowStartsNewPage {
		t.Error("firstRowStartsNewPage should be false")
	}
	if !b.printOnBottom {
		t.Error("printOnBottom should be true")
	}
	if !b.keepChild {
		t.Error("keepChild should be true")
	}
	if b.outlineExpression != "[Category]" {
		t.Errorf("outlineExpression = %q, want [Category]", b.outlineExpression)
	}
	if b.repeatBandNTimes != 4 {
		t.Errorf("repeatBandNTimes = %d, want 4", b.repeatBandNTimes)
	}
	if b.beforeLayoutEvent != "onBefore" {
		t.Errorf("beforeLayoutEvent = %q, want onBefore", b.beforeLayoutEvent)
	}
	if b.afterLayoutEvent != "onAfter" {
		t.Errorf("afterLayoutEvent = %q, want onAfter", b.afterLayoutEvent)
	}
}

// ─── ChildBand.Deserialize: all branches ──────────────────────────────────────
// Note: the "if err := c.BandBase.Deserialize(r); err != nil { return err }"
// path (line 404) is unreachable dead code.

type childBandDeserMock struct {
	*mockReader
	fillUnused  bool
	completeN   int
	printIfEmpty bool
}

func (m *childBandDeserMock) ReadBool(name string, def bool) bool {
	switch name {
	case "FillUnusedSpace":
		return m.fillUnused
	case "PrintIfDatabandEmpty":
		return m.printIfEmpty
	case "FirstRowStartsNewPage":
		return true // keep default
	case "CanBreak":
		return true // keep default
	default:
		return def
	}
}

func (m *childBandDeserMock) ReadInt(name string, def int) int {
	if name == "CompleteToNRows" {
		return m.completeN
	}
	return def
}

func TestChildBand_Deserialize_AllNonDefaults_v2(t *testing.T) {
	r := &childBandDeserMock{
		mockReader:   newMockReader(),
		fillUnused:   true,
		completeN:    7,
		printIfEmpty: true,
	}

	c := NewChildBand()
	if err := c.Deserialize(r); err != nil {
		t.Fatalf("ChildBand.Deserialize error: %v", err)
	}
	if !c.FillUnusedSpace {
		t.Error("FillUnusedSpace should be true")
	}
	if c.CompleteToNRows != 7 {
		t.Errorf("CompleteToNRows = %d, want 7", c.CompleteToNRows)
	}
	if !c.PrintIfDatabandEmpty {
		t.Error("PrintIfDatabandEmpty should be true")
	}
}

func TestChildBand_Deserialize_Defaults(t *testing.T) {
	r := newMockReader()
	c := NewChildBand()
	if err := c.Deserialize(r); err != nil {
		t.Fatalf("ChildBand.Deserialize error: %v", err)
	}
	if c.FillUnusedSpace {
		t.Error("FillUnusedSpace should default to false")
	}
	if c.CompleteToNRows != 0 {
		t.Errorf("CompleteToNRows should default to 0, got %d", c.CompleteToNRows)
	}
	if c.PrintIfDatabandEmpty {
		t.Error("PrintIfDatabandEmpty should default to false")
	}
}

// ─── HeaderFooterBandBase.serializeAttrs ──────────────────────────────────────
// Note: the "if err := h.BandBase.serializeAttrs(w); err != nil { return err }"
// path (line 44) is unreachable dead code.

func TestHeaderFooterBandBase_serializeAttrs_BothFlags(t *testing.T) {
	h := NewHeaderFooterBandBase()
	h.SetKeepWithData(true)
	h.SetRepeatOnEveryPage(true)

	w := newFailAttrsWriter2()
	if err := h.serializeAttrs(w); err != nil {
		t.Errorf("serializeAttrs error: %v", err)
	}
	if _, ok := w.written["KeepWithData"]; !ok {
		t.Error("KeepWithData should be written")
	}
	if _, ok := w.written["RepeatOnEveryPage"]; !ok {
		t.Error("RepeatOnEveryPage should be written")
	}
}

func TestHeaderFooterBandBase_serializeAttrs_KeepWithDataOnly(t *testing.T) {
	h := NewHeaderFooterBandBase()
	h.SetKeepWithData(true)
	h.SetRepeatOnEveryPage(false)

	w := newFailAttrsWriter2()
	if err := h.serializeAttrs(w); err != nil {
		t.Errorf("serializeAttrs error: %v", err)
	}
	if _, ok := w.written["KeepWithData"]; !ok {
		t.Error("KeepWithData should be written")
	}
	if _, ok := w.written["RepeatOnEveryPage"]; ok {
		t.Error("RepeatOnEveryPage should not be written (false=default)")
	}
}

func TestHeaderFooterBandBase_serializeAttrs_RepeatOnEveryPageOnly(t *testing.T) {
	h := NewHeaderFooterBandBase()
	h.SetKeepWithData(false)
	h.SetRepeatOnEveryPage(true)

	w := newFailAttrsWriter2()
	if err := h.serializeAttrs(w); err != nil {
		t.Errorf("serializeAttrs error: %v", err)
	}
	if _, ok := w.written["RepeatOnEveryPage"]; !ok {
		t.Error("RepeatOnEveryPage should be written")
	}
	if _, ok := w.written["KeepWithData"]; ok {
		t.Error("KeepWithData should not be written (false=default)")
	}
}

func TestHeaderFooterBandBase_serializeAttrs_NeitherFlag(t *testing.T) {
	h := NewHeaderFooterBandBase()
	// Both defaults: neither should be written.
	w := newFailAttrsWriter2()
	if err := h.serializeAttrs(w); err != nil {
		t.Errorf("serializeAttrs error: %v", err)
	}
	if _, ok := w.written["KeepWithData"]; ok {
		t.Error("KeepWithData should not be written (false=default)")
	}
	if _, ok := w.written["RepeatOnEveryPage"]; ok {
		t.Error("RepeatOnEveryPage should not be written (false=default)")
	}
}

// ─── HeaderFooterBandBase.Serialize: children-error path ─────────────────────
// Note: the serializeAttrs-error path (line 57) is unreachable dead code.

func TestHeaderFooterBandBase_Serialize_ChildrenErrorPath(t *testing.T) {
	h := NewHeaderFooterBandBase()
	child := &minimalSerializable2{BaseObject: *report.NewBaseObject()}
	h.AddChild(child)

	w := newFailAttrsWriter2()
	w.failWriteObject = true

	err := h.Serialize(w)
	if err == nil {
		t.Error("HeaderFooterBandBase.Serialize: expected error from failing WriteObject")
	}
}

func TestHeaderFooterBandBase_Serialize_Success(t *testing.T) {
	h := NewHeaderFooterBandBase()
	h.SetKeepWithData(true)
	h.SetRepeatOnEveryPage(true)
	child := &minimalSerializable2{BaseObject: *report.NewBaseObject()}
	h.AddChild(child)

	w := newFailAttrsWriter2()
	if err := h.Serialize(w); err != nil {
		t.Errorf("HeaderFooterBandBase.Serialize: unexpected error: %v", err)
	}
}

// ─── HeaderFooterBandBase.Deserialize: all branches ──────────────────────────
// Note: the "if err := h.BandBase.Deserialize(r); err != nil { return err }"
// path (line 65) is unreachable dead code.

type hfDeserMock struct {
	*mockReader
	keepWithData      bool
	repeatOnEveryPage bool
}

func (m *hfDeserMock) ReadBool(name string, def bool) bool {
	switch name {
	case "KeepWithData":
		return m.keepWithData
	case "RepeatOnEveryPage":
		return m.repeatOnEveryPage
	case "FirstRowStartsNewPage":
		return true
	case "CanBreak":
		return true
	default:
		return def
	}
}

func TestHeaderFooterBandBase_Deserialize_BothTrue(t *testing.T) {
	r := &hfDeserMock{
		mockReader:        newMockReader(),
		keepWithData:      true,
		repeatOnEveryPage: true,
	}
	h := NewHeaderFooterBandBase()
	if err := h.Deserialize(r); err != nil {
		t.Fatalf("Deserialize error: %v", err)
	}
	if !h.KeepWithData() {
		t.Error("KeepWithData should be true")
	}
	if !h.RepeatOnEveryPage() {
		t.Error("RepeatOnEveryPage should be true")
	}
}

func TestHeaderFooterBandBase_Deserialize_BothFalse(t *testing.T) {
	r := &hfDeserMock{
		mockReader:        newMockReader(),
		keepWithData:      false,
		repeatOnEveryPage: false,
	}
	h := NewHeaderFooterBandBase()
	if err := h.Deserialize(r); err != nil {
		t.Fatalf("Deserialize error: %v", err)
	}
	if h.KeepWithData() {
		t.Error("KeepWithData should be false")
	}
	if h.RepeatOnEveryPage() {
		t.Error("RepeatOnEveryPage should be false")
	}
}

func TestHeaderFooterBandBase_Deserialize_OnlyKeepWithData(t *testing.T) {
	r := &hfDeserMock{
		mockReader:        newMockReader(),
		keepWithData:      true,
		repeatOnEveryPage: false,
	}
	h := NewHeaderFooterBandBase()
	if err := h.Deserialize(r); err != nil {
		t.Fatalf("Deserialize error: %v", err)
	}
	if !h.KeepWithData() {
		t.Error("KeepWithData should be true")
	}
	if h.RepeatOnEveryPage() {
		t.Error("RepeatOnEveryPage should be false")
	}
}

func TestHeaderFooterBandBase_Deserialize_OnlyRepeatOnEveryPage(t *testing.T) {
	r := &hfDeserMock{
		mockReader:        newMockReader(),
		keepWithData:      false,
		repeatOnEveryPage: true,
	}
	h := NewHeaderFooterBandBase()
	if err := h.Deserialize(r); err != nil {
		t.Fatalf("Deserialize error: %v", err)
	}
	if h.KeepWithData() {
		t.Error("KeepWithData should be false")
	}
	if !h.RepeatOnEveryPage() {
		t.Error("RepeatOnEveryPage should be true")
	}
}

// ─── GroupHeaderBand.Serialize: all branches ──────────────────────────────────
// Note: the "if err := g.HeaderFooterBandBase.serializeAttrs(w); err != nil { return err }"
// path (line 239) is unreachable dead code.

func TestGroupHeaderBand_Serialize_AllBranches(t *testing.T) {
	g := NewGroupHeaderBand()
	g.SetCondition("[Department]")
	g.SetSortOrder(SortOrderDescending)
	g.SetKeepTogether(true)
	g.SetResetPageNumber(true)
	g.SetKeepWithData(true)
	g.SetRepeatOnEveryPage(true)

	w := newFailAttrsWriter2()
	if err := g.Serialize(w); err != nil {
		t.Errorf("GroupHeaderBand.Serialize error: %v", err)
	}
	if _, ok := w.written["Condition"]; !ok {
		t.Error("expected Condition to be written")
	}
	if _, ok := w.written["SortOrder"]; !ok {
		t.Error("expected SortOrder to be written")
	}
	if _, ok := w.written["KeepTogether"]; !ok {
		t.Error("expected KeepTogether to be written")
	}
	if _, ok := w.written["ResetPageNumber"]; !ok {
		t.Error("expected ResetPageNumber to be written")
	}
	if _, ok := w.written["KeepWithData"]; !ok {
		t.Error("expected KeepWithData to be written")
	}
	if _, ok := w.written["RepeatOnEveryPage"]; !ok {
		t.Error("expected RepeatOnEveryPage to be written")
	}
}

func TestGroupHeaderBand_Serialize_Defaults(t *testing.T) {
	g := NewGroupHeaderBand()
	// No fields set — only SortOrderAscending is default, all booleans false.
	w := newFailAttrsWriter2()
	if err := g.Serialize(w); err != nil {
		t.Errorf("GroupHeaderBand.Serialize error: %v", err)
	}
	// None of the optional attributes should be written at defaults.
	for _, key := range []string{"Condition", "SortOrder", "KeepTogether", "ResetPageNumber"} {
		if _, ok := w.written[key]; ok {
			t.Errorf("attribute %q should not be written at default", key)
		}
	}
}

func TestGroupHeaderBand_Serialize_ChildrenErrorPath(t *testing.T) {
	g := NewGroupHeaderBand()
	child := &minimalSerializable2{BaseObject: *report.NewBaseObject()}
	g.AddChild(child)

	w := newFailAttrsWriter2()
	w.failWriteObject = true

	err := g.Serialize(w)
	if err == nil {
		t.Error("GroupHeaderBand.Serialize: expected error from failing WriteObject")
	}
}

// ─── GroupHeaderBand.Deserialize: all branches ────────────────────────────────
// Note: the "if err := g.HeaderFooterBandBase.Deserialize(r); err != nil { return err }"
// path (line 259) is unreachable dead code.

type groupHeaderDeserMock struct {
	*mockReader
	condition       string
	// sortOrder holds the FRX string name ("Ascending", "Descending", "None").
	// C# serialises SortOrder via Converter.ToString (enum name format "G").
	sortOrder       string
	keepTogether    bool
	resetPageNumber bool
}

func (m *groupHeaderDeserMock) ReadStr(name, def string) string {
	switch name {
	case "Condition":
		return m.condition
	case "SortOrder":
		if m.sortOrder != "" {
			return m.sortOrder
		}
		return def
	}
	return def
}

func (m *groupHeaderDeserMock) ReadBool(name string, def bool) bool {
	switch name {
	case "KeepTogether":
		return m.keepTogether
	case "ResetPageNumber":
		return m.resetPageNumber
	case "FirstRowStartsNewPage":
		return true
	case "CanBreak":
		return true
	default:
		return def
	}
}

func (m *groupHeaderDeserMock) ReadInt(name string, def int) int {
	return def
}

func TestGroupHeaderBand_Deserialize_AllNonDefaults(t *testing.T) {
	r := &groupHeaderDeserMock{
		mockReader:      newMockReader(),
		condition:       "[Region]",
		sortOrder:       "Descending",
		keepTogether:    true,
		resetPageNumber: true,
	}

	g := NewGroupHeaderBand()
	if err := g.Deserialize(r); err != nil {
		t.Fatalf("GroupHeaderBand.Deserialize error: %v", err)
	}
	if g.Condition() != "[Region]" {
		t.Errorf("Condition = %q, want [Region]", g.Condition())
	}
	if g.SortOrder() != SortOrderDescending {
		t.Errorf("SortOrder = %v, want Descending", g.SortOrder())
	}
	if !g.KeepTogether() {
		t.Error("KeepTogether should be true")
	}
	if !g.ResetPageNumber() {
		t.Error("ResetPageNumber should be true")
	}
}

func TestGroupHeaderBand_Deserialize_Defaults(t *testing.T) {
	r := &groupHeaderDeserMock{
		mockReader:      newMockReader(),
		condition:       "",
		sortOrder:       "Ascending",
		keepTogether:    false,
		resetPageNumber: false,
	}

	g := NewGroupHeaderBand()
	if err := g.Deserialize(r); err != nil {
		t.Fatalf("GroupHeaderBand.Deserialize error: %v", err)
	}
	if g.Condition() != "" {
		t.Errorf("Condition should be empty, got %q", g.Condition())
	}
	if g.SortOrder() != SortOrderAscending {
		t.Errorf("SortOrder = %v, want Ascending", g.SortOrder())
	}
	if g.KeepTogether() {
		t.Error("KeepTogether should default to false")
	}
	if g.ResetPageNumber() {
		t.Error("ResetPageNumber should default to false")
	}
}

// ─── DataBand.Serialize: all branches ─────────────────────────────────────────
// Note: the "if err := d.BandBase.serializeAttrs(w); err != nil { return err }"
// path (line 427) is unreachable dead code.

func TestDataBand_Serialize_AllOptionalAttrs(t *testing.T) {
	d := NewDataBand()
	d.SetFilter("[Amount] > 100")
	d.AddSort(SortSpec{Column: "Name", Order: SortOrderAscending})
	d.AddSort(SortSpec{Column: "Date", Order: SortOrderDescending})
	d.AddSort(SortSpec{Expression: "[Total]", Order: SortOrderDescending})
	d.SetPrintIfDetailEmpty(true)
	d.SetPrintIfDSEmpty(true)
	d.SetKeepTogether(true)
	d.SetKeepDetail(true)
	d.SetIDColumn("ID")
	d.SetParentIDColumn("ParentID")
	d.SetIndent(20)
	d.SetKeepSummary(true)

	w := newFailAttrsWriter2()
	if err := d.Serialize(w); err != nil {
		t.Errorf("DataBand.Serialize error: %v", err)
	}
	for _, key := range []string{
		"Filter",
		"PrintIfDetailEmpty", "PrintIfDatasourceEmpty",
		"KeepTogether", "KeepDetail",
		"IdColumn", "ParentIdColumn",
		"Indent", "KeepSummary",
	} {
		if _, ok := w.written[key]; !ok {
			t.Errorf("expected attribute %q to be written", key)
		}
	}
	// Sort is written as a child element via WriteObjectNamed, not as an attribute.
	if _, ok := w.written["Sort"]; ok {
		t.Error("Sort should NOT be written as an attribute; it is a child element now")
	}
}

func TestDataBand_Serialize_Defaults(t *testing.T) {
	d := NewDataBand()
	w := newFailAttrsWriter2()
	if err := d.Serialize(w); err != nil {
		t.Errorf("DataBand.Serialize error: %v", err)
	}
	// No optional attributes should be written at defaults.
	for _, key := range []string{
		"Filter",
		"PrintIfDetailEmpty", "PrintIfDatasourceEmpty",
		"KeepTogether", "KeepDetail",
		"IdColumn", "ParentIdColumn",
		"Indent", "KeepSummary",
	} {
		if _, ok := w.written[key]; ok {
			t.Errorf("attribute %q should not be written at default", key)
		}
	}
}

func TestDataBand_Serialize_ChildrenErrorPath(t *testing.T) {
	d := NewDataBand()
	child := &minimalSerializable2{BaseObject: *report.NewBaseObject()}
	d.AddChild(child)

	w := newFailAttrsWriter2()
	w.failWriteObject = true

	err := d.Serialize(w)
	if err == nil {
		t.Error("DataBand.Serialize: expected error from failing WriteObject")
	}
}

// ─── DataBand.Deserialize: all branches ───────────────────────────────────────
// Note: the "if err := d.BandBase.Deserialize(r); err != nil { return err }"
// path (line 479) is unreachable dead code.

type dataBandDeserMock struct {
	*mockReader
	filter           string
	sortStr          string
	printIfDetail    bool
	printIfDSEmpty   bool
	keepTogether     bool
	keepDetail       bool
	idColumn         string
	parentIDColumn   string
	indent           float32
	keepSummary      bool
	columnsCount     int
	columnsWidth     float32
	dataSourceAlias  string
}

func (m *dataBandDeserMock) ReadStr(name, def string) string {
	switch name {
	case "DataSource":
		return m.dataSourceAlias
	case "Filter":
		return m.filter
	case "Sort":
		return m.sortStr
	case "IdColumn":
		return m.idColumn
	case "ParentIdColumn":
		return m.parentIDColumn
	default:
		return def
	}
}

func (m *dataBandDeserMock) ReadBool(name string, def bool) bool {
	switch name {
	case "PrintIfDetailEmpty":
		return m.printIfDetail
	case "PrintIfDatasourceEmpty":
		return m.printIfDSEmpty
	case "KeepTogether":
		return m.keepTogether
	case "KeepDetail":
		return m.keepDetail
	case "KeepSummary":
		return m.keepSummary
	case "FirstRowStartsNewPage":
		return true
	case "CanBreak":
		return true
	default:
		return def
	}
}

func (m *dataBandDeserMock) ReadInt(name string, def int) int {
	if name == "Columns.Count" {
		return m.columnsCount
	}
	return def
}

func (m *dataBandDeserMock) ReadFloat(name string, def float32) float32 {
	switch name {
	case "Indent":
		return m.indent
	case "Columns.Width":
		return m.columnsWidth
	default:
		return def
	}
}

func TestDataBand_Deserialize_AllNonDefaults(t *testing.T) {
	r := &dataBandDeserMock{
		mockReader:      newMockReader(),
		dataSourceAlias: "Customers",
		filter:          "[Total] > 0",
		sortStr:         "Name ASC;Date DESC",
		printIfDetail:   true,
		printIfDSEmpty:  true,
		keepTogether:    true,
		keepDetail:      true,
		idColumn:        "ID",
		parentIDColumn:  "ParentID",
		indent:          15,
		keepSummary:     true,
		columnsCount:    3,
		columnsWidth:    200,
	}

	d := NewDataBand()
	if err := d.Deserialize(r); err != nil {
		t.Fatalf("DataBand.Deserialize error: %v", err)
	}

	if d.DataSourceAlias() != "Customers" {
		t.Errorf("DataSourceAlias = %q, want Customers", d.DataSourceAlias())
	}
	if d.Filter() != "[Total] > 0" {
		t.Errorf("Filter = %q, want [Total] > 0", d.Filter())
	}
	if len(d.Sort()) != 2 {
		t.Errorf("Sort len = %d, want 2", len(d.Sort()))
	} else {
		if d.Sort()[0].Column != "Name" {
			t.Errorf("Sort[0].Column = %q, want Name", d.Sort()[0].Column)
		}
		if d.Sort()[0].Order != SortOrderAscending {
			t.Errorf("Sort[0].Order = %v, want Ascending", d.Sort()[0].Order)
		}
		if d.Sort()[1].Column != "Date" {
			t.Errorf("Sort[1].Column = %q, want Date", d.Sort()[1].Column)
		}
		if d.Sort()[1].Order != SortOrderDescending {
			t.Errorf("Sort[1].Order = %v, want Descending", d.Sort()[1].Order)
		}
	}
	if !d.PrintIfDetailEmpty() {
		t.Error("PrintIfDetailEmpty should be true")
	}
	if !d.PrintIfDSEmpty() {
		t.Error("PrintIfDSEmpty should be true")
	}
	if !d.KeepTogether() {
		t.Error("KeepTogether should be true")
	}
	if !d.KeepDetail() {
		t.Error("KeepDetail should be true")
	}
	if d.IDColumn() != "ID" {
		t.Errorf("IDColumn = %q, want ID", d.IDColumn())
	}
	if d.ParentIDColumn() != "ParentID" {
		t.Errorf("ParentIDColumn = %q, want ParentID", d.ParentIDColumn())
	}
	if d.Indent() != 15 {
		t.Errorf("Indent = %v, want 15", d.Indent())
	}
	if !d.KeepSummary() {
		t.Error("KeepSummary should be true")
	}
	if d.Columns().Count() != 3 {
		t.Errorf("Columns.Count = %d, want 3", d.Columns().Count())
	}
	if d.Columns().Width != 200 {
		t.Errorf("Columns.Width = %v, want 200", d.Columns().Width)
	}
}

func TestDataBand_Deserialize_SortAscDesc(t *testing.T) {
	// Test that sort strings with explicit ASC/DESC are parsed correctly.
	r := &dataBandDeserMock{
		mockReader: newMockReader(),
		sortStr:    "ColA ASC;ColB DESC",
	}
	d := NewDataBand()
	if err := d.Deserialize(r); err != nil {
		t.Fatalf("DataBand.Deserialize error: %v", err)
	}
	if len(d.Sort()) != 2 {
		t.Fatalf("Sort len = %d, want 2", len(d.Sort()))
	}
	if d.Sort()[1].Order != SortOrderDescending {
		t.Errorf("Sort[1].Order = %v, want Descending", d.Sort()[1].Order)
	}
}

func TestDataBand_Deserialize_SortEmptyParts(t *testing.T) {
	// Empty parts between semicolons should be skipped.
	r := &dataBandDeserMock{
		mockReader: newMockReader(),
		sortStr:    ";ColA ASC;;",
	}
	d := NewDataBand()
	if err := d.Deserialize(r); err != nil {
		t.Fatalf("DataBand.Deserialize error: %v", err)
	}
	if len(d.Sort()) != 1 {
		t.Fatalf("Sort len = %d, want 1 (empty parts skipped)", len(d.Sort()))
	}
}
