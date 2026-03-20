package band_test

// band_databand_coverage_test.go – tests for DataBand methods that were at 0% coverage:
//   - CollectChildRows / SetCollectChildRows
//   - ResetPageNumber / SetResetPageNumber (DataBand version)
//   - VirtualRowCount / SetVirtualRowCount
//   - IsDatasourceEmpty (nil datasource and non-nil datasource)
//   - IsDeepmostDataBand (no nested bands vs nested DataBand / GroupHeaderBand)
//   - UpdateLayout (no-op – verify it does not panic with children present)

import (
	"testing"

	"github.com/andrewloable/go-fastreport/band"
)

// ─── CollectChildRows / SetCollectChildRows ────────────────────────────────────

func TestDataBand_CollectChildRows_DefaultFalse(t *testing.T) {
	d := band.NewDataBand()
	if d.CollectChildRows() {
		t.Error("CollectChildRows should default to false")
	}
}

func TestDataBand_CollectChildRows_SetTrue(t *testing.T) {
	d := band.NewDataBand()
	d.SetCollectChildRows(true)
	if !d.CollectChildRows() {
		t.Error("CollectChildRows should be true after SetCollectChildRows(true)")
	}
}

func TestDataBand_CollectChildRows_SetFalse(t *testing.T) {
	d := band.NewDataBand()
	d.SetCollectChildRows(true)
	d.SetCollectChildRows(false)
	if d.CollectChildRows() {
		t.Error("CollectChildRows should be false after SetCollectChildRows(false)")
	}
}

// ─── ResetPageNumber / SetResetPageNumber (DataBand) ──────────────────────────

func TestDataBand_ResetPageNumber_DefaultFalse(t *testing.T) {
	d := band.NewDataBand()
	if d.ResetPageNumber() {
		t.Error("DataBand.ResetPageNumber should default to false")
	}
}

func TestDataBand_ResetPageNumber_SetTrue(t *testing.T) {
	d := band.NewDataBand()
	d.SetResetPageNumber(true)
	if !d.ResetPageNumber() {
		t.Error("DataBand.ResetPageNumber should be true after SetResetPageNumber(true)")
	}
}

func TestDataBand_ResetPageNumber_SetFalse(t *testing.T) {
	d := band.NewDataBand()
	d.SetResetPageNumber(true)
	d.SetResetPageNumber(false)
	if d.ResetPageNumber() {
		t.Error("DataBand.ResetPageNumber should be false after SetResetPageNumber(false)")
	}
}

// ─── VirtualRowCount / SetVirtualRowCount ────────────────────────────────────

func TestDataBand_VirtualRowCount_DefaultZero(t *testing.T) {
	d := band.NewDataBand()
	// The field default is 0 (zero value); the FRX default of 1 is only applied
	// during deserialization via ReadInt("RowCount", 1).
	if d.VirtualRowCount() != 0 {
		t.Errorf("VirtualRowCount default = %d, want 0", d.VirtualRowCount())
	}
}

func TestDataBand_VirtualRowCount_Set(t *testing.T) {
	d := band.NewDataBand()
	d.SetVirtualRowCount(5)
	if d.VirtualRowCount() != 5 {
		t.Errorf("VirtualRowCount = %d, want 5", d.VirtualRowCount())
	}
}

func TestDataBand_VirtualRowCount_SetZero(t *testing.T) {
	d := band.NewDataBand()
	d.SetVirtualRowCount(10)
	d.SetVirtualRowCount(0)
	if d.VirtualRowCount() != 0 {
		t.Errorf("VirtualRowCount = %d, want 0 after reset", d.VirtualRowCount())
	}
}

// ─── IsDatasourceEmpty ────────────────────────────────────────────────────────

// stubDataSourceEmpty is a stub DataSource with zero rows.
type stubDataSourceEmpty struct{}

func (s *stubDataSourceEmpty) RowCount() int                    { return 0 }
func (s *stubDataSourceEmpty) First() error                     { return nil }
func (s *stubDataSourceEmpty) Next() error                      { return nil }
func (s *stubDataSourceEmpty) EOF() bool                        { return true }
func (s *stubDataSourceEmpty) GetValue(col string) (any, error) { return nil, nil }

// stubDataSourceNonEmpty is a stub DataSource with rows.
type stubDataSourceNonEmpty struct{}

func (s *stubDataSourceNonEmpty) RowCount() int                    { return 3 }
func (s *stubDataSourceNonEmpty) First() error                     { return nil }
func (s *stubDataSourceNonEmpty) Next() error                      { return nil }
func (s *stubDataSourceNonEmpty) EOF() bool                        { return false }
func (s *stubDataSourceNonEmpty) GetValue(col string) (any, error) { return nil, nil }

func TestDataBand_IsDatasourceEmpty_NilDS(t *testing.T) {
	d := band.NewDataBand()
	// No DataSource set → should be empty.
	if !d.IsDatasourceEmpty() {
		t.Error("IsDatasourceEmpty should be true when DataSource is nil")
	}
}

func TestDataBand_IsDatasourceEmpty_ZeroRowDS(t *testing.T) {
	d := band.NewDataBand()
	d.SetDataSource(&stubDataSourceEmpty{})
	if !d.IsDatasourceEmpty() {
		t.Error("IsDatasourceEmpty should be true when DataSource has 0 rows")
	}
}

func TestDataBand_IsDatasourceEmpty_NonEmptyDS(t *testing.T) {
	d := band.NewDataBand()
	d.SetDataSource(&stubDataSourceNonEmpty{})
	if d.IsDatasourceEmpty() {
		t.Error("IsDatasourceEmpty should be false when DataSource has rows")
	}
}

// ─── IsDeepmostDataBand ───────────────────────────────────────────────────────

func TestDataBand_IsDeepmostDataBand_NoChildren(t *testing.T) {
	d := band.NewDataBand()
	if !d.IsDeepmostDataBand() {
		t.Error("DataBand with no children should be the deepmost data band")
	}
}

func TestDataBand_IsDeepmostDataBand_NonBandChild(t *testing.T) {
	// Adding a non-DataBand child (a plain BaseObject stub) should still be deepmost.
	parent := band.NewDataBand()
	obj := newMinimalBase("obj1")
	parent.AddChild(obj)
	if !parent.IsDeepmostDataBand() {
		t.Error("DataBand with only non-DataBand children should still be deepmost")
	}
}

func TestDataBand_IsDeepmostDataBand_WithNestedDataBand(t *testing.T) {
	parent := band.NewDataBand()
	child := band.NewDataBand()
	parent.AddChild(child)
	if parent.IsDeepmostDataBand() {
		t.Error("DataBand with a nested DataBand child should NOT be deepmost")
	}
}

func TestDataBand_IsDeepmostDataBand_WithNestedGroupHeaderBand(t *testing.T) {
	parent := band.NewDataBand()
	child := band.NewGroupHeaderBand()
	parent.AddChild(child)
	if parent.IsDeepmostDataBand() {
		t.Error("DataBand with a nested GroupHeaderBand child should NOT be deepmost")
	}
}

// ─── UpdateLayout (BandBase, no-op) ──────────────────────────────────────────

// TestBandBase_UpdateLayout_WithChildren verifies that UpdateLayout does not
// panic when the band has child objects. The BandBase implementation is a
// deliberate no-op; child positions are managed by the engine during prepare.
func TestBandBase_UpdateLayout_WithChildren(t *testing.T) {
	b := band.NewBandBase()
	obj1 := newMinimalBase("child1")
	obj2 := newMinimalBase("child2")
	b.AddChild(obj1)
	b.AddChild(obj2)

	// Must not panic with dx=10, dy=20.
	b.UpdateLayout(10, 20)

	// Objects are unchanged because UpdateLayout is a no-op at this level.
	if b.Objects().Len() != 2 {
		t.Errorf("Objects().Len = %d after UpdateLayout, want 2", b.Objects().Len())
	}
}
