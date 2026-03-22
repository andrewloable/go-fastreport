package crossview

import (
	"strconv"
	"strings"

	"github.com/andrewloable/go-fastreport/report"
)

// ── HeaderDescriptor serialization ───────────────────────────────────────────

// Serialize writes header descriptor properties to FRX (element name = "Header").
func (h *HeaderDescriptor) Serialize(w report.Writer) error {
	if h.IsTotal {
		w.WriteBool("IsTotal", true)
	}
	if h.IsGrandTotal {
		w.WriteBool("IsGrandTotal", true)
	}
	if h.FieldName != "" {
		w.WriteStr("FieldName", h.FieldName)
	}
	if h.MeasureName != "" {
		w.WriteStr("MeasureName", h.MeasureName)
	}
	if h.IsMeasure {
		w.WriteBool("IsMeasure", true)
	}
	if h.Cell != 0 {
		w.WriteInt("Cell", h.Cell)
	}
	if h.CellSize != 0 {
		w.WriteInt("CellSize", h.CellSize)
	}
	if h.Level != 0 {
		w.WriteInt("Level", h.Level)
	}
	if h.LevelSize != 0 {
		w.WriteInt("LevelSize", h.LevelSize)
	}
	if h.Expression != "" {
		w.WriteStr("Expression", h.Expression)
	}
	return nil
}

// Deserialize reads header descriptor properties from FRX.
func (h *HeaderDescriptor) Deserialize(r report.Reader) error {
	h.IsTotal = r.ReadBool("IsTotal", false)
	h.IsGrandTotal = r.ReadBool("IsGrandTotal", false)
	h.FieldName = r.ReadStr("FieldName", "")
	h.MeasureName = r.ReadStr("MeasureName", "")
	h.IsMeasure = r.ReadBool("IsMeasure", false)
	h.Cell = r.ReadInt("Cell", 0)
	h.CellSize = r.ReadInt("CellSize", 0)
	h.Level = r.ReadInt("Level", 0)
	h.LevelSize = r.ReadInt("LevelSize", 0)
	h.Expression = r.ReadStr("Expression", "")
	return nil
}

// ── CellDescriptor serialization ─────────────────────────────────────────────

// Serialize writes cell descriptor properties to FRX (element name = "Cell").
func (c *CellDescriptor) Serialize(w report.Writer) error {
	if c.IsXTotal {
		w.WriteBool("IsXTotal", true)
	}
	if c.IsYTotal {
		w.WriteBool("IsYTotal", true)
	}
	if c.IsXGrandTotal {
		w.WriteBool("IsXGrandTotal", true)
	}
	if c.IsYGrandTotal {
		w.WriteBool("IsYGrandTotal", true)
	}
	xfn := c.XFieldName
	if c.IsXGrandTotal {
		xfn = ""
	}
	if xfn != "" {
		w.WriteStr("XFieldName", xfn)
	}
	yfn := c.YFieldName
	if c.IsYGrandTotal {
		yfn = ""
	}
	if yfn != "" {
		w.WriteStr("YFieldName", yfn)
	}
	if c.MeasureName != "" {
		w.WriteStr("MeasureName", c.MeasureName)
	}
	if c.X != 0 {
		w.WriteInt("X", c.X)
	}
	if c.Y != 0 {
		w.WriteInt("Y", c.Y)
	}
	if c.Expression != "" {
		w.WriteStr("Expression", c.Expression)
	}
	return nil
}

// Deserialize reads cell descriptor properties from FRX.
func (c *CellDescriptor) Deserialize(r report.Reader) error {
	c.IsXTotal = r.ReadBool("IsXTotal", false)
	c.IsYTotal = r.ReadBool("IsYTotal", false)
	c.IsXGrandTotal = r.ReadBool("IsXGrandTotal", false)
	c.IsYGrandTotal = r.ReadBool("IsYGrandTotal", false)
	c.XFieldName = r.ReadStr("XFieldName", "")
	c.YFieldName = r.ReadStr("YFieldName", "")
	c.MeasureName = r.ReadStr("MeasureName", "")
	c.X = r.ReadInt("X", 0)
	c.Y = r.ReadInt("Y", 0)
	c.Expression = r.ReadStr("Expression", "")
	return nil
}

// ── CrossViewHeader collection ────────────────────────────────────────────────

// CrossViewHeader is a named, serializable collection of HeaderDescriptor items.
// It is the Go equivalent of FastReport.CrossView.CrossViewHeader.
type CrossViewHeader struct {
	// Name is the FRX element name used when serializing (e.g. "Columns" or "Rows").
	Name  string
	Items []*HeaderDescriptor
}

// NewCrossViewHeader creates an empty CrossViewHeader with the given name.
func NewCrossViewHeader(name string) *CrossViewHeader {
	return &CrossViewHeader{Name: name}
}

// Add appends a descriptor.
func (h *CrossViewHeader) Add(d *HeaderDescriptor) { h.Items = append(h.Items, d) }

// Count returns the number of items.
func (h *CrossViewHeader) Count() int { return len(h.Items) }

// Get returns the descriptor at index, or nil if out of range.
func (h *CrossViewHeader) Get(i int) *HeaderDescriptor {
	if i < 0 || i >= len(h.Items) {
		return nil
	}
	return h.Items[i]
}

// Clear removes all items.
func (h *CrossViewHeader) Clear() { h.Items = h.Items[:0] }

// IndexOf returns the index of d in the collection, or -1 if not found.
// Mirrors C# CrossViewHeader.IndexOf (CrossViewHeader.cs line 82–85).
func (h *CrossViewHeader) IndexOf(d *HeaderDescriptor) int {
	for i, item := range h.Items {
		if item == d {
			return i
		}
	}
	return -1
}

// Contains reports whether d is a member of the collection (pointer equality).
// Mirrors C# CrossViewHeader.Contains (CrossViewHeader.cs line 92–95).
func (h *CrossViewHeader) Contains(d *HeaderDescriptor) bool {
	return h.IndexOf(d) >= 0
}

// Insert inserts d at the given index.  If index >= Count, d is appended.
// Mirrors C# CrossViewHeader.Insert (CrossViewHeader.cs line 60–63).
func (h *CrossViewHeader) Insert(index int, d *HeaderDescriptor) {
	if index >= len(h.Items) {
		h.Items = append(h.Items, d)
		return
	}
	h.Items = append(h.Items, nil)
	copy(h.Items[index+1:], h.Items[index:])
	h.Items[index] = d
}

// Remove removes the first occurrence of d from the collection (by pointer equality).
// If d is not found, Remove is a no-op.
// Mirrors C# CrossViewHeader.Remove (CrossViewHeader.cs line 69–74).
func (h *CrossViewHeader) Remove(d *HeaderDescriptor) {
	idx := h.IndexOf(d)
	if idx < 0 {
		return
	}
	h.Items = append(h.Items[:idx], h.Items[idx+1:]...)
}

// ToArray returns a shallow copy of the items slice.
// Mirrors C# CrossViewHeader.ToArray (CrossViewHeader.cs line 101–109).
func (h *CrossViewHeader) ToArray() []*HeaderDescriptor {
	cp := make([]*HeaderDescriptor, len(h.Items))
	copy(cp, h.Items)
	return cp
}

// AddRange appends all descriptors in items to the collection.
// Mirrors C# CrossViewHeader.AddRange (CrossViewHeader.cs line 37–43).
func (h *CrossViewHeader) AddRange(items []*HeaderDescriptor) {
	h.Items = append(h.Items, items...)
}

// Serialize writes each descriptor as a "Header" child element.
func (h *CrossViewHeader) Serialize(w report.Writer) error {
	for _, d := range h.Items {
		if err := w.WriteObjectNamed("Header", d); err != nil {
			return err
		}
	}
	return nil
}

// Deserialize reads "Header" child elements into the collection.
func (h *CrossViewHeader) Deserialize(r report.Reader) error {
	h.Clear()
	for {
		childType, ok := r.NextChild()
		if !ok {
			break
		}
		if childType == "Header" {
			d := &HeaderDescriptor{}
			if err := d.Deserialize(r); err != nil {
				if err2 := r.FinishChild(); err2 != nil {
					break
				}
				continue
			}
			h.Add(d)
		}
		if err := r.FinishChild(); err != nil {
			break
		}
	}
	return nil
}

// ── CrossViewCells collection ─────────────────────────────────────────────────

// CrossViewCells is a named, serializable collection of CellDescriptor items.
// It is the Go equivalent of FastReport.CrossView.CrossViewCells.
type CrossViewCells struct {
	// Name is the FRX element name used when serializing (e.g. "Cells").
	Name  string
	Items []*CellDescriptor
}

// NewCrossViewCells creates an empty CrossViewCells with the given name.
func NewCrossViewCells(name string) *CrossViewCells {
	return &CrossViewCells{Name: name}
}

// Add appends a descriptor.
func (c *CrossViewCells) Add(d *CellDescriptor) { c.Items = append(c.Items, d) }

// Count returns the number of items.
func (c *CrossViewCells) Count() int { return len(c.Items) }

// Get returns the descriptor at index, or nil if out of range.
func (c *CrossViewCells) Get(i int) *CellDescriptor {
	if i < 0 || i >= len(c.Items) {
		return nil
	}
	return c.Items[i]
}

// Clear removes all items.
func (c *CrossViewCells) Clear() { c.Items = c.Items[:0] }

// Serialize writes each descriptor as a "Cell" child element.
func (c *CrossViewCells) Serialize(w report.Writer) error {
	for _, d := range c.Items {
		if err := w.WriteObjectNamed("Cell", d); err != nil {
			return err
		}
	}
	return nil
}

// Deserialize reads "Cell" child elements into the collection.
func (c *CrossViewCells) Deserialize(r report.Reader) error {
	c.Clear()
	for {
		childType, ok := r.NextChild()
		if !ok {
			break
		}
		if childType == "Cell" {
			d := &CellDescriptor{}
			if err := d.Deserialize(r); err != nil {
				if err2 := r.FinishChild(); err2 != nil {
					break
				}
				continue
			}
			c.Add(d)
		}
		if err := r.FinishChild(); err != nil {
			break
		}
	}
	return nil
}

// ── CrossViewData serialization ───────────────────────────────────────────────

// CrossViewDataSerial wraps CrossViewData for FRX round-trip.  It stores the
// descriptor index arrays as comma-separated strings (matching the C# implementation).
type CrossViewDataSerial struct {
	*CrossViewData
	// Index arrays (stored as comma-separated strings in FRX).
	ColumnDescriptorsIndexes string
	RowDescriptorsIndexes    string
	ColumnTerminalIndexes    string
	RowTerminalIndexes       string

	columnHeader *CrossViewHeader
	rowHeader    *CrossViewHeader
	cells        *CrossViewCells
}

// NewCrossViewDataSerial wraps a CrossViewData for serialization.
func NewCrossViewDataSerial(d *CrossViewData) *CrossViewDataSerial {
	s := &CrossViewDataSerial{
		CrossViewData: d,
		columnHeader:  NewCrossViewHeader("Columns"),
		rowHeader:     NewCrossViewHeader("Rows"),
		cells:         NewCrossViewCells("Cells"),
	}
	// Sync collections from the flat slices.
	for _, h := range d.Columns {
		s.columnHeader.Add(h)
	}
	for _, h := range d.Rows {
		s.rowHeader.Add(h)
	}
	for _, c := range d.Cells {
		s.cells.Add(c)
	}
	return s
}

// Serialize writes CrossViewData and its children to FRX.
func (s *CrossViewDataSerial) Serialize(w report.Writer) error {
	if s.ColumnDescriptorsIndexes != "" {
		w.WriteStr("ColumnDescriptorsIndexes", s.ColumnDescriptorsIndexes)
	}
	if s.RowDescriptorsIndexes != "" {
		w.WriteStr("RowDescriptorsIndexes", s.RowDescriptorsIndexes)
	}
	if s.ColumnTerminalIndexes != "" {
		w.WriteStr("ColumnTerminalIndexes", s.ColumnTerminalIndexes)
	}
	if s.RowTerminalIndexes != "" {
		w.WriteStr("RowTerminalIndexes", s.RowTerminalIndexes)
	}
	if err := w.WriteObjectNamed("Columns", s.columnHeader); err != nil {
		return err
	}
	if err := w.WriteObjectNamed("Rows", s.rowHeader); err != nil {
		return err
	}
	if err := w.WriteObjectNamed("Cells", s.cells); err != nil {
		return err
	}
	return nil
}

// Deserialize reads CrossViewData and its children from FRX.
func (s *CrossViewDataSerial) Deserialize(r report.Reader) error {
	s.ColumnDescriptorsIndexes = r.ReadStr("ColumnDescriptorsIndexes", "")
	s.RowDescriptorsIndexes = r.ReadStr("RowDescriptorsIndexes", "")
	s.ColumnTerminalIndexes = r.ReadStr("ColumnTerminalIndexes", "")
	s.RowTerminalIndexes = r.ReadStr("RowTerminalIndexes", "")
	for {
		childType, ok := r.NextChild()
		if !ok {
			break
		}
		switch childType {
		case "Columns":
			if err := s.columnHeader.Deserialize(r); err != nil {
				if err2 := r.FinishChild(); err2 != nil {
					return nil
				}
				continue
			}
		case "Rows":
			if err := s.rowHeader.Deserialize(r); err != nil {
				if err2 := r.FinishChild(); err2 != nil {
					return nil
				}
				continue
			}
		case "Cells":
			if err := s.cells.Deserialize(r); err != nil {
				if err2 := r.FinishChild(); err2 != nil {
					return nil
				}
				continue
			}
		}
		if err := r.FinishChild(); err != nil {
			return nil
		}
	}
	// Sync deserialized items back to CrossViewData.
	s.CrossViewData.Columns = s.columnHeader.Items
	s.CrossViewData.Rows = s.rowHeader.Items
	s.CrossViewData.Cells = s.cells.Items
	return nil
}

// ── Index array helpers ───────────────────────────────────────────────────────

// ParseIndexArray parses a comma-separated string of ints (e.g. "0,1,2") into a slice.
func ParseIndexArray(s string) []int {
	if s == "" {
		return nil
	}
	parts := strings.Split(s, ",")
	out := make([]int, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if n, err := strconv.Atoi(p); err == nil {
			out = append(out, n)
		}
	}
	return out
}

// FormatIndexArray formats an int slice as a comma-separated string.
func FormatIndexArray(indices []int) string {
	if len(indices) == 0 {
		return ""
	}
	parts := make([]string, len(indices))
	for i, n := range indices {
		parts[i] = strconv.Itoa(n)
	}
	return strings.Join(parts, ",")
}
