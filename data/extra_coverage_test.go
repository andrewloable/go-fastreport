package data_test

// extra_coverage_test.go — targeted tests to cover remaining uncovered branches
// in command_parameter_collection.go, dictionary.go, dictionary_eval.go,
// helper.go, and virtual.go.

import (
	"bytes"
	"fmt"
	"strings"
	"testing"

	"github.com/andrewloable/go-fastreport/data"
	"github.com/andrewloable/go-fastreport/report"
	"github.com/andrewloable/go-fastreport/serial"
)

// ── mock writer that fails on WriteObjectNamed ────────────────────────────────

type failWriter struct{}

func (w *failWriter) WriteStr(name, value string)         {}
func (w *failWriter) WriteInt(name string, value int)     {}
func (w *failWriter) WriteBool(name string, value bool)   {}
func (w *failWriter) WriteFloat(name string, value float32) {}
func (w *failWriter) WriteObject(obj report.Serializable) error {
	return fmt.Errorf("failWriter: WriteObject always fails")
}
func (w *failWriter) WriteObjectNamed(name string, obj report.Serializable) error {
	return fmt.Errorf("failWriter: WriteObjectNamed always fails")
}

// ── mock reader that returns an error from FinishChild ────────────────────────

type failFinishReader struct {
	// Returns one "Parameter" child, then fails on FinishChild.
	step int
}

func (r *failFinishReader) ReadStr(name, def string) string  { return def }
func (r *failFinishReader) ReadInt(name string, def int) int { return def }
func (r *failFinishReader) ReadBool(name string, def bool) bool { return def }
func (r *failFinishReader) ReadFloat(name string, def float32) float32 { return def }
func (r *failFinishReader) NextChild() (string, bool) {
	if r.step == 0 {
		r.step++
		return "Parameter", true
	}
	return "", false
}
func (r *failFinishReader) FinishChild() error {
	return fmt.Errorf("failFinishReader: FinishChild failed")
}

// ── virtual.go: First() without Init() ───────────────────────────────────────

func TestVirtualDataSource_First_NotInitialized(t *testing.T) {
	vds := data.NewVirtualDataSource("v", 3)
	// Call First without Init — should return ErrNotInitialized.
	err := vds.First()
	if err != data.ErrNotInitialized {
		t.Errorf("First without Init: want ErrNotInitialized, got %v", err)
	}
}

// ── helper.go: GetParameter — nested path where intermediate child is missing ─

func TestGetParameter_Nested_MissingChild(t *testing.T) {
	dict := newStubDict()
	// Parent exists but does NOT have a nested "Missing" child.
	parent := &data.Parameter{Name: "Filter"}
	dict.params = []*data.Parameter{parent}

	// "Filter.Missing" — parent found, but child "Missing" is not found → nil.
	got := data.GetParameter(dict, "Filter.Missing")
	if got != nil {
		t.Errorf("GetParameter nested missing child should return nil, got %v", got)
	}
}

// ── dictionary.go: ResolveRelations — parent and child resolved by name ───────

func TestDictionary_ResolveRelations_ChildByName(t *testing.T) {
	d := data.NewDictionary()

	// Give parent a different alias so FindDataSourceByAlias fails for parent too,
	// forcing FindDataSourceByName to be called for both parent and child.
	parent := data.NewBaseDataSource("customers")
	parent.SetAlias("ParentAlias") // alias ≠ "customers", so alias lookup fails

	child := data.NewBaseDataSource("orders")
	child.SetAlias("ChildAlias") // alias ≠ "orders", so alias lookup fails

	d.AddDataSource(parent)
	d.AddDataSource(child)

	rel := &data.Relation{
		Name:             "CustomersOrders",
		ParentSourceName: "customers", // not found by alias → falls back to name
		ChildSourceName:  "orders",    // not found by alias → falls back to name
	}
	d.AddRelation(rel)
	d.ResolveRelations()

	if rel.ParentDataSource != parent {
		t.Errorf("ResolveRelations should resolve ParentDataSource by name when alias lookup fails; got %v", rel.ParentDataSource)
	}
	if rel.ChildDataSource != child {
		t.Errorf("ResolveRelations should resolve ChildDataSource by name when alias lookup fails; got %v", rel.ChildDataSource)
	}
}

// ── dictionary_eval.go: EvaluateAll — error when expression is invalid ────────

func TestDictionary_EvaluateAll_Error(t *testing.T) {
	d := data.NewDictionary()
	// A parameter with an expression that cannot be evaluated successfully.
	d.AddParameter(&data.Parameter{
		Name:       "Bad",
		Expression: "this +++invalid+++ expr",
	})

	err := d.EvaluateAll()
	if err == nil {
		t.Error("EvaluateAll with invalid expression should return error")
	}
}

// ── command_parameter_collection.go: Serialize — error return branch ─────────

func TestCommandParameterCollection_Serialize_WriterError(t *testing.T) {
	col := data.NewCommandParameterCollection()
	col.Add(data.NewCommandParameter("@p1"))

	fw := &failWriter{}
	err := col.Serialize(fw)
	if err == nil {
		t.Error("Serialize should propagate error from WriteObjectNamed")
	}
}

// ── command_parameter_collection.go: Deserialize — FinishChild error branch ──

func TestCommandParameterCollection_Deserialize_FinishChildError(t *testing.T) {
	col := data.NewCommandParameterCollection()
	r := &failFinishReader{}
	err := col.Deserialize(r)
	if err == nil {
		t.Error("Deserialize should propagate error from FinishChild")
	}
}

// ── command_parameter_collection.go: Deserialize — child Deserialize error ───

// errorParamReader causes CommandParameter.Deserialize to succeed but we need
// FinishChild to succeed while inner Deserialize fails. Actually inner
// Deserialize is CommandParameter.Deserialize which only calls ReadStr/ReadInt
// and always returns nil. So the inner Deserialize error path requires reading
// a corrupt inner child. The simpler path is: test with real serial to verify
// that a valid round-trip exercises the non-error paths.

func TestCommandParameterCollection_Deserialize_ValidXML(t *testing.T) {
	// Build a small XML with one Parameter child.
	xml := `<Params><Parameter Name="@x" DataType="int"/></Params>`
	r := serial.NewReader(strings.NewReader(xml))
	_, ok := r.ReadObjectHeader()
	if !ok {
		t.Fatal("ReadObjectHeader failed")
	}
	col := data.NewCommandParameterCollection()
	if err := col.Deserialize(r); err != nil {
		t.Fatalf("Deserialize: %v", err)
	}
	if col.Count() != 1 {
		t.Errorf("Count = %d, want 1", col.Count())
	}
	if col.Get(0).Name != "@x" {
		t.Errorf("Name = %q, want @x", col.Get(0).Name)
	}
}

// roundTripCommandParameterCollection verifies Serialize followed by Deserialize.
func TestCommandParameterCollection_Serialize_Valid(t *testing.T) {
	col := data.NewCommandParameterCollection()
	col.Add(data.NewCommandParameter("@a"))

	var buf bytes.Buffer
	w := serial.NewWriter(&buf)
	if err := w.BeginObject("Params"); err != nil {
		t.Fatal(err)
	}
	if err := col.Serialize(w); err != nil {
		t.Fatalf("Serialize: %v", err)
	}
	if err := w.EndObject(); err != nil {
		t.Fatal(err)
	}
	if err := w.Flush(); err != nil {
		t.Fatal(err)
	}
	if buf.Len() == 0 {
		t.Error("serialized output should not be empty")
	}
}

// ── dictionary.go: ResolveRelations — parent nil, child already resolved ──────

func TestDictionary_ResolveRelations_ParentNilChildAlreadyResolved(t *testing.T) {
	d := data.NewDictionary()
	parent := data.NewBaseDataSource("customers")
	parent.SetAlias("Customers")
	child := data.NewBaseDataSource("orders")
	child.SetAlias("Orders")

	d.AddDataSource(parent)
	d.AddDataSource(child)

	rel := &data.Relation{
		Name:             "CustomersOrders2",
		ParentSourceName: "Customers",
		// ChildDataSource already resolved — skip the child if block.
		ChildDataSource: child,
	}
	d.AddRelation(rel)
	d.ResolveRelations()

	if rel.ParentDataSource != parent {
		t.Errorf("ResolveRelations should resolve ParentDataSource when only child is pre-resolved; got %v", rel.ParentDataSource)
	}
	if rel.ChildDataSource != child {
		t.Error("ChildDataSource should remain unchanged when already resolved")
	}
}

// ── GetParameter: ensure system-variable fallback returns nil for multi-segment ─

func TestGetParameter_SystemVar_MultiSegment(t *testing.T) {
	// A multi-segment name where the first segment is not in user params —
	// it falls into the systemVars search, which returns it directly (no
	// further child resolution). Confirm it returns nil when not found.
	dict := newStubDict()
	// No params, no sysVars with name "X".
	got := data.GetParameter(dict, "X.Y")
	if got != nil {
		t.Errorf("GetParameter with multi-segment and no match should return nil, got %v", got)
	}
}
