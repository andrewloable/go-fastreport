package data_test

// data_extra2_coverage_test.go — targeted tests to cover remaining uncovered branches
// in datasource.go (compareAny: int64/float64/nil cases, SortRows multi-spec),
// connection.go (tableName path, rows.Err path),
// filter.go (time-range default fallthrough, unknown FilterOperation final return),
// view.go (rebuildIndex inner.Next non-EOF error break).

import (
	"database/sql"
	"database/sql/driver"
	"errors"
	"io"
	"testing"
	"time"

	"github.com/andrewloable/go-fastreport/data"
)

// ── compareAny: int64 case ─────────────────────────────────────────────────────

func TestSortRows_Int64(t *testing.T) {
	ds := data.NewBaseDataSource("test")
	ds.AddRow(map[string]any{"v": int64(30)})
	ds.AddRow(map[string]any{"v": int64(10)})
	ds.AddRow(map[string]any{"v": int64(20)})
	_ = ds.Init()

	ds.SortRows([]data.SortSpec{{Column: "v"}})
	_ = ds.First()
	v, _ := ds.GetValue("v")
	if v.(int64) != 10 {
		t.Errorf("first after int64 sort = %v, want 10", v)
	}
}

func TestSortRows_Float64(t *testing.T) {
	ds := data.NewBaseDataSource("test")
	ds.AddRow(map[string]any{"v": float64(3.14)})
	ds.AddRow(map[string]any{"v": float64(1.41)})
	ds.AddRow(map[string]any{"v": float64(2.71)})
	_ = ds.Init()

	ds.SortRows([]data.SortSpec{{Column: "v"}})
	_ = ds.First()
	v, _ := ds.GetValue("v")
	if v.(float64) != 1.41 {
		t.Errorf("first after float64 sort = %v, want 1.41", v)
	}
}

// compareAny default branch: type not handled by any case.
func TestSortRows_DefaultType(t *testing.T) {
	type MyStruct struct{ N int }
	ds := data.NewBaseDataSource("test")
	ds.AddRow(map[string]any{"v": MyStruct{2}})
	ds.AddRow(map[string]any{"v": MyStruct{1}})
	_ = ds.Init()
	// Should not panic — default fallback uses fmt.Sprintf.
	ds.SortRows([]data.SortSpec{{Column: "v"}})
}

// compareAny bool: equal branch (returns 0, triggers multi-spec fallthrough).
func TestSortRows_MultiSpec(t *testing.T) {
	ds := data.NewBaseDataSource("test")
	ds.AddRow(map[string]any{"group": "B", "rank": 2})
	ds.AddRow(map[string]any{"group": "A", "rank": 3})
	ds.AddRow(map[string]any{"group": "A", "rank": 1})
	_ = ds.Init()

	ds.SortRows([]data.SortSpec{
		{Column: "group"},
		{Column: "rank"},
	})
	_ = ds.First()
	g, _ := ds.GetValue("group")
	r, _ := ds.GetValue("rank")
	if g != "A" || r.(int) != 1 {
		t.Errorf("first row = group=%v rank=%v, want A 1", g, r)
	}
	_ = ds.Next()
	g2, _ := ds.GetValue("group")
	r2, _ := ds.GetValue("rank")
	if g2 != "A" || r2.(int) != 3 {
		t.Errorf("second row = group=%v rank=%v, want A 3", g2, r2)
	}
}

// compareAny: nil key (missing from map) hits default case via fmt.Sprintf.
func TestSortRows_NilValues(t *testing.T) {
	ds := data.NewBaseDataSource("test")
	ds.AddRow(map[string]any{"v": "B"})
	ds.AddRow(map[string]any{}) // no "v" key → nil
	ds.AddRow(map[string]any{"v": "A"})
	_ = ds.Init()
	ds.SortRows([]data.SortSpec{{Column: "v"}}) // should not panic
}

// compareAny: bool equal case (av == b.(bool) → return 0).
func TestSortRows_Bool_Equal(t *testing.T) {
	ds := data.NewBaseDataSource("test")
	ds.AddRow(map[string]any{"v": true, "id": 1})
	ds.AddRow(map[string]any{"v": true, "id": 2}) // same bool → compareAny returns 0
	_ = ds.Init()
	// No panic; stable sort preserves order for equal keys.
	ds.SortRows([]data.SortSpec{{Column: "v"}})
	_ = ds.First()
	v, _ := ds.GetValue("id")
	if v.(int) != 1 {
		t.Errorf("stable sort: first id = %v, want 1", v)
	}
}

// SortRows: all specs produce 0 → outer loop final return 0.
func TestSortRows_AllSpecsTie(t *testing.T) {
	ds := data.NewBaseDataSource("test")
	ds.AddRow(map[string]any{"a": "same", "b": "same", "id": 1})
	ds.AddRow(map[string]any{"a": "same", "b": "same", "id": 2})
	_ = ds.Init()
	// All sort keys are equal → inner loop exhausted → SortStableFunc returns 0.
	ds.SortRows([]data.SortSpec{
		{Column: "a"},
		{Column: "b"},
	})
	// Stable sort preserves original order.
	_ = ds.First()
	v, _ := ds.GetValue("id")
	if v.(int) != 1 {
		t.Errorf("all-tie sort: first id = %v, want 1", v)
	}
}

// ── filter.go: matches — time-range default (fallthrough to general compare) ──
// When value is time.Time AND fe.Value is [2]time.Time AND operation is
// something other than Equal/NotEqual/Contains/NotContains, the default branch
// is hit and execution falls through to the general compare section.

func TestFilter_TimeRange_LessThan_Fallthrough(t *testing.T) {
	start := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(2024, 1, 31, 0, 0, 0, 0, time.UTC)
	rng := [2]time.Time{start, end}

	f := data.NewDataSourceFilter()
	f.Add(rng, data.FilterLessThan) // not Equal/NotEqual/Contains/NotContains → default

	testTime := time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC)
	// Falls through to compare(time.Time, [2]time.Time) which is incomparable → false.
	if f.ValueMatch(testTime) {
		t.Error("time-range with FilterLessThan fallthrough should return false")
	}
}

func TestFilter_TimeRange_GreaterThanOrEqual_Fallthrough(t *testing.T) {
	start := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(2024, 1, 31, 0, 0, 0, 0, time.UTC)
	rng := [2]time.Time{start, end}

	f := data.NewDataSourceFilter()
	f.Add(rng, data.FilterGreaterThanOrEqual)

	testTime := time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC)
	if f.ValueMatch(testTime) {
		t.Error("time-range with FilterGreaterThanOrEqual fallthrough should return false")
	}
}

// ── filter.go: matches — final return false (unknown FilterOperation) ─────────
// The general-compare switch only handles ops 0-5 (Equal through GreaterThanOrEqual).
// Any other value falls through to `return false` at the end of matches.

func TestFilter_Matches_UnknownOperation_ReturnsFalse(t *testing.T) {
	f := data.NewDataSourceFilter()
	f.Add(42, data.FilterOperation(99)) // not one of the defined operations
	// compare(42, 42) returns (0, true), but switch has no case 99 → return false.
	if f.ValueMatch(42) {
		t.Error("unknown FilterOperation should return false via final return false")
	}
}

// ── filter.go: compare — int64 with incompatible b ───────────────────────────
// compare(a=int64, b=string) → toInt64("abc") fails → (0, false).

func TestFilter_Int64_IncompatibleB(t *testing.T) {
	f := data.NewDataSourceFilter()
	// fe.Value = "notanumber": compare(int64(value), "notanumber") → error → false.
	f.Add("notanumber", data.FilterGreaterThan)
	if f.ValueMatch(int64(10)) {
		t.Error("int64 vs string FilterGreaterThan should not match")
	}
}

// ── filtered.go: seekInner — loop path (target row > 0) ──────────────────────
// seekInner calls First() then Next() N times to reach target row N.

func TestFilteredDataSource_SeekInner_Loop(t *testing.T) {
	items := data.NewBaseDataSource("items")
	items.AddColumn(data.Column{Name: "id"})
	items.AddRow(map[string]any{"id": "skip1"}) // idx 0
	items.AddRow(map[string]any{"id": "skip2"}) // idx 1
	items.AddRow(map[string]any{"id": "match"}) // idx 2
	_ = items.Init()

	// Filter to rows where id == "match" — only row at index 2.
	fds, err := data.NewFilteredDataSource(items, []string{"id"}, []string{"match"})
	if err != nil {
		t.Fatalf("NewFilteredDataSource: %v", err)
	}
	if fds.RowCount() != 1 {
		t.Fatalf("expected 1 row, got %d", fds.RowCount())
	}

	// First() → cursor=0 → seekInner seeks to inner row 2 (Next called twice).
	if err := fds.First(); err != nil {
		t.Fatalf("First: %v", err)
	}
	v, err := fds.GetValue("id")
	if err != nil {
		t.Fatalf("GetValue: %v", err)
	}
	if v != "match" {
		t.Errorf("GetValue(id) = %v, want match", v)
	}
}

// ── connection.go Init: tableName path (no selectCommand, uses tableName) ─────

func TestTableDataSource_Init_WithTableName(t *testing.T) {
	// The already-registered "stub" driver returns 2 rows for any query.
	// Using tableName without selectCommand → Init builds "SELECT * FROM <tableName>".
	c := data.NewDataConnectionBase("stub")
	c.ConnectionString = "stub://test"
	ts := c.CreateTable("products")
	ts.SetTableName("products_table") // no SelectCommand set

	if err := ts.Init(); err != nil {
		t.Fatalf("Init with TableName: %v", err)
	}
	if ts.RowCount() != 2 {
		t.Errorf("RowCount = %d, want 2", ts.RowCount())
	}
	_ = c.Close()
}

// ── connection.go Init: rows.Err() error path ─────────────────────────────────
// Register a driver whose rows.Next returns a non-io.EOF error so that
// database/sql propagates it via rows.Err().

type rowsErrDriver2 struct{}
type rowsErrConn2 struct{}
type rowsErrTx2 struct{}
type rowsErrStmt2 struct{}
type rowsErrRows2 struct{ pos int }

func (d *rowsErrDriver2) Open(name string) (driver.Conn, error) { return &rowsErrConn2{}, nil }
func (c *rowsErrConn2) Prepare(q string) (driver.Stmt, error)   { return &rowsErrStmt2{}, nil }
func (c *rowsErrConn2) Close() error                             { return nil }
func (c *rowsErrConn2) Begin() (driver.Tx, error)               { return &rowsErrTx2{}, nil }
func (t *rowsErrTx2) Commit() error                              { return nil }
func (t *rowsErrTx2) Rollback() error                            { return nil }
func (s *rowsErrStmt2) Close() error                             { return nil }
func (s *rowsErrStmt2) NumInput() int                            { return -1 }
func (s *rowsErrStmt2) Exec(args []driver.Value) (driver.Result, error) {
	return nil, nil
}
func (s *rowsErrStmt2) Query(args []driver.Value) (driver.Rows, error) {
	return &rowsErrRows2{}, nil
}
func (r *rowsErrRows2) Columns() []string { return []string{"id"} }
func (r *rowsErrRows2) Close() error      { return nil }
func (r *rowsErrRows2) Next(dest []driver.Value) error {
	r.pos++
	if r.pos == 1 {
		// Return a non-io.EOF error — database/sql will set rows.Err() to this.
		return errors.New("driver: simulated row iteration error")
	}
	return io.EOF
}

func init() {
	sql.Register("stub-rows-err2", &rowsErrDriver2{})
}

func TestTableDataSource_Init_RowsErrPath(t *testing.T) {
	c := data.NewDataConnectionBase("stub-rows-err2")
	c.ConnectionString = "stub-rows-err2://test"
	ts := c.CreateTable("t")
	ts.SetSelectCommand("SELECT id FROM t")

	err := ts.Init()
	if err == nil {
		t.Error("Init should return error when rows.Err() is non-nil")
	}
	_ = c.Close()
}

// ── view.go: rebuildIndex — inner.Next returns non-ErrEOF error → break ───────

// nextErrorSource returns an error from Next() after errAfter successful calls.
type nextErrorSource struct {
	rows      []map[string]any
	cursor    int
	errAfter  int
	nextCount int
}

func newNextErrorSource(errAfter int, rows ...map[string]any) *nextErrorSource {
	return &nextErrorSource{rows: rows, cursor: -1, errAfter: errAfter}
}

func (s *nextErrorSource) Name() string  { return "nextErr" }
func (s *nextErrorSource) Alias() string { return "nextErr" }
func (s *nextErrorSource) Init() error {
	s.cursor = -1
	s.nextCount = 0
	return nil
}
func (s *nextErrorSource) First() error {
	s.cursor = 0
	s.nextCount = 0
	if len(s.rows) == 0 {
		return data.ErrEOF
	}
	return nil
}
func (s *nextErrorSource) Next() error {
	s.nextCount++
	if s.nextCount > s.errAfter {
		return errors.New("intentional Next error")
	}
	s.cursor++
	if s.cursor >= len(s.rows) {
		return data.ErrEOF
	}
	return nil
}
func (s *nextErrorSource) EOF() bool         { return s.cursor >= len(s.rows) }
func (s *nextErrorSource) RowCount() int     { return len(s.rows) }
func (s *nextErrorSource) CurrentRowNo() int { return s.cursor }
func (s *nextErrorSource) GetValue(col string) (any, error) {
	if s.cursor < 0 || s.cursor >= len(s.rows) {
		return nil, errors.New("out of range")
	}
	v, ok := s.rows[s.cursor][col]
	if !ok {
		return nil, nil
	}
	return v, nil
}
func (s *nextErrorSource) Close() error { return nil }

func TestViewDataSource_RebuildIndex_NextError(t *testing.T) {
	// 3 rows, but Next() fails after 1 successful call.
	// rebuildIndex: First() → cursor=0. Loop:
	//   iter 1: EOF=false → include row0 → Next() #1 (ok, cursor=1).
	//   iter 2: EOF=false → include row1 → Next() #2 (error, non-ErrEOF) → break.
	// Result: 2 rows indexed.
	inner := newNextErrorSource(1,
		map[string]any{"x": 1},
		map[string]any{"x": 2},
		map[string]any{"x": 3},
	)

	vds := data.NewViewDataSource(inner, "v", "v", "", nil)
	if err := vds.Init(); err != nil {
		t.Fatalf("Init: %v", err)
	}
	if vds.RowCount() != 2 {
		t.Errorf("RowCount = %d, want 2 (break after Next error)", vds.RowCount())
	}
}

func TestViewDataSource_RebuildIndex_NextErrorImmediate(t *testing.T) {
	// Next() fails on the first call (errAfter=0).
	// rebuildIndex: First() → cursor=0. Loop:
	//   iter 1: EOF=false → include row0 → Next() #1 (error immediately) → break.
	// Result: 1 row indexed.
	inner := newNextErrorSource(0,
		map[string]any{"x": 1},
		map[string]any{"x": 2},
	)

	vds := data.NewViewDataSource(inner, "v", "v", "", nil)
	if err := vds.Init(); err != nil {
		t.Fatalf("Init: %v", err)
	}
	if vds.RowCount() != 1 {
		t.Errorf("RowCount = %d, want 1 (break on first Next error)", vds.RowCount())
	}
}

// ── filter.go: matches — time scalar with unsupported operation (line 193) ────
// When value is time.Time, fe.Value is time.Time, but the operation is not one
// of the six comparison ops handled in the switch, the final `return false` at
// line 193 is reached.

func TestFilter_TimeScalar_UnsupportedOp_ReturnsFalse(t *testing.T) {
	tv := time.Date(2024, 6, 15, 0, 0, 0, 0, time.UTC)
	f := data.NewDataSourceFilter()
	f.Add(tv, data.FilterOperation(99)) // undefined op — falls through switch to return false
	if f.ValueMatch(tv) {
		t.Error("time.Time with unknown FilterOperation should return false")
	}
}

// ── filter.go: matches — DateTime scalar LessThanOrEqual / GreaterThan / GreaterThanOrEqual ──
// These three operations in the time-scalar branch were not yet exercised.

func TestFilter_TimeScalar_LessThanOrEqual(t *testing.T) {
	filterDate := time.Date(2024, 6, 15, 0, 0, 0, 0, time.UTC)
	f := data.NewDataSourceFilter()
	f.Add(filterDate, data.FilterLessThanOrEqual)

	// Same day with time → stripped to same date → equal → LessThanOrEqual should match.
	sameDay := time.Date(2024, 6, 15, 10, 0, 0, 0, time.UTC)
	if !f.ValueMatch(sameDay) {
		t.Error("same day (after time strip) should match LessThanOrEqual (equal case)")
	}

	// Day before — strictly less.
	before := time.Date(2024, 6, 14, 23, 59, 59, 0, time.UTC)
	if !f.ValueMatch(before) {
		t.Error("day before filter should match LessThanOrEqual")
	}

	// Day after — should not match.
	after := time.Date(2024, 6, 16, 0, 0, 0, 0, time.UTC)
	if f.ValueMatch(after) {
		t.Error("day after filter should not match LessThanOrEqual")
	}
}

func TestFilter_TimeScalar_GreaterThan(t *testing.T) {
	filterDate := time.Date(2024, 6, 15, 0, 0, 0, 0, time.UTC)
	f := data.NewDataSourceFilter()
	f.Add(filterDate, data.FilterGreaterThan)

	// Day after → greater.
	after := time.Date(2024, 6, 16, 0, 0, 0, 0, time.UTC)
	if !f.ValueMatch(after) {
		t.Error("day after filter should match GreaterThan")
	}

	// Same day with time → stripped to same → equal → NOT greater.
	sameDay := time.Date(2024, 6, 15, 10, 0, 0, 0, time.UTC)
	if f.ValueMatch(sameDay) {
		t.Error("same day (after strip) should not match GreaterThan")
	}
}

func TestFilter_TimeScalar_GreaterThanOrEqual(t *testing.T) {
	filterDate := time.Date(2024, 6, 15, 0, 0, 0, 0, time.UTC)
	f := data.NewDataSourceFilter()
	f.Add(filterDate, data.FilterGreaterThanOrEqual)

	// Same day with time → stripped to same → equal → GreaterThanOrEqual should match.
	sameDay := time.Date(2024, 6, 15, 10, 0, 0, 0, time.UTC)
	if !f.ValueMatch(sameDay) {
		t.Error("same day (after strip) should match GreaterThanOrEqual")
	}

	// Day after → greater.
	after := time.Date(2024, 6, 16, 0, 0, 0, 0, time.UTC)
	if !f.ValueMatch(after) {
		t.Error("day after filter should match GreaterThanOrEqual")
	}

	// Day before → should not match.
	before := time.Date(2024, 6, 14, 0, 0, 0, 0, time.UTC)
	if f.ValueMatch(before) {
		t.Error("day before filter should not match GreaterThanOrEqual")
	}
}

// ── datacomponent.go: SetName — case-insensitive alias sync (C# line 96) ──────
// C# uses String.Compare(Alias, Name, true) (case-insensitive).
// When alias has the same letters as name but different case, SetName must
// still update alias to stay in sync.

func TestDataComponentBase_SetName_CaseInsensitiveAliasSync(t *testing.T) {
	d := data.NewDataComponentBase("original")
	// Force alias to be a different case than name.
	d.SetAlias("ORIGINAL")
	// SetName should detect "ORIGINAL" == "original" (case-insensitively) and
	// update alias to the new name.
	d.SetName("updated")
	if d.Alias() != "updated" {
		t.Errorf("Alias should sync to updated when alias was same name (diff case), got %q", d.Alias())
	}
}
