package matrix

// header_store.go implements the MatrixHeader — the typed-value header tree
// that drives the AddValue/GetValue/SetValue pipeline.
//
// C# source: FastReport.Base/Matrix/MatrixHeader.cs
//            FastReport.Base/Matrix/MatrixHeaderItem.cs
//
// Gaps implemented (issues go-fastreport-919wy and go-fastreport-7ose3):
//   - MatrixHeader type with RootItem, nextIndex
//   - Find(address, create, dataRowNo) — core tree navigation with binary search
//   - FindOrCreate(address) public wrapper
//   - RemoveItem(address)
//   - GetTerminalIndices()
//   - AddTotalItems(isTemplate)
//   - Reset()

import "github.com/andrewloable/go-fastreport/table"

// MatrixHeader is the runtime header tree used by the AddValue pipeline.
// It holds a root *HeaderItem plus the collection of *HeaderDescriptors that
// define the sort order and totals behaviour for each level.
//
// The MatrixData type uses two MatrixHeader instances (Columns and Rows).
// C# source: FastReport.Base/Matrix/MatrixHeader.cs.
type MatrixHeader struct {
	// Root is the virtual root node; its children are the first-level header values.
	Root *HeaderItem

	// Descriptors is the ordered list of header descriptors for this axis
	// (one per level). Populated by MatrixData when descriptors are added.
	Descriptors []*HeaderDescriptor

	// nextIndex is the monotonically increasing flat cell-array index
	// assigned to each terminal (leaf) header item.
	// C# source: MatrixHeader.nextIndex field.
	nextIndex int
}

// NewMatrixHeader creates an empty MatrixHeader.
func NewMatrixHeader() *MatrixHeader {
	return &MatrixHeader{
		Root: newHeaderItem(""),
	}
}

// Reset clears all runtime header tree state and resets the index counter.
// C# source: MatrixHeader.Reset().
func (h *MatrixHeader) Reset() {
	ItemClearChildren(h.Root)
	h.nextIndex = 0
}

// Find locates (and optionally creates) the terminal HeaderItem addressed by
// address. address must have the same length as h.Descriptors.
//
// If create is false and the address is not found, Find returns nil.
// If create is true a new item is inserted in sort order at each missing level;
// the terminal item receives the next available Index.
// dataRowNo is recorded on new items (used for DataSource.CurrentRowNo in the engine).
//
// C# source: MatrixHeader.Find(object[] address, bool create, int dataRowNo).
func (h *MatrixHeader) Find(address []any, create bool, dataRowNo int) *HeaderItem {
	cur := h.Root
	for i, val := range address {
		sort := SortOrderAscending
		var descrExt *HeaderDescriptorExt
		if i < len(h.Descriptors) {
			sort = h.Descriptors[i].Sort
			descrExt = h.Descriptors[i].HeaderExt()
		}

		idx := ItemFind(cur, val, sort)
		if idx >= 0 {
			cur = cur.Children[idx]
			continue
		}

		if !create {
			return nil
		}

		// Create new item at the insertion point (~idx = ^idx).
		insertPos := ^idx
		newItem := newHeaderItem(anyToString(val))
		SetItemParent(newItem, cur)
		SetItemAnyValue(newItem, val)
		SetItemDataRowNo(newItem, dataRowNo)
		if descrExt != nil {
			// Copy template bindings from the descriptor.
			newItem.CellSize = 0 // will be computed later
			// Store TemplateColumn/Row/Cell references via descriptor ext.
			// These are used by the engine to copy visual properties.
			// We attach them through the item's DescriptorExt side-map entry.
			// (actual rendering is out of scope for this iteration)
			_ = descrExt
		}
		if i < len(h.Descriptors) {
			SetItemPageBreak(newItem, h.Descriptors[i].PageBreak)
		}

		// Only terminal items (last level) get an index.
		if i == len(address)-1 {
			SetItemIndex(newItem, h.nextIndex)
			h.nextIndex++
		}

		insertChildAt(cur, insertPos, newItem)
		cur = newItem
	}
	return cur
}

// FindIndex returns the flat index of the terminal item at address, or -1.
// C# source: MatrixHeader.Find(object[] address) — the public overload.
func (h *MatrixHeader) FindIndex(address []any) int {
	item := h.Find(address, false, 0)
	if item != nil {
		return ItemIndex(item)
	}
	return -1
}

// FindOrCreate returns the flat index of the terminal item at address, creating
// it if absent. C# source: MatrixHeader.FindOrCreate(object[] address).
func (h *MatrixHeader) FindOrCreate(address []any) int {
	item := h.Find(address, true, 0)
	if item != nil {
		return ItemIndex(item)
	}
	return -1
}

// RemoveItem removes the terminal item at address from its parent's children list.
// C# source: MatrixHeader.RemoveItem(object[] address).
func (h *MatrixHeader) RemoveItem(address []any) {
	item := h.Find(address, false, 0)
	if item == nil {
		return
	}
	parent := ItemParent(item)
	if parent == nil {
		return
	}
	// Remove from parent.Children.
	for i, ch := range parent.Children {
		if ch == item {
			parent.Children = append(parent.Children[:i], parent.Children[i+1:]...)
			// Rebuild childIndex.
			parent.childIndex = make(map[string]int, len(parent.Children))
			for j, c := range parent.Children {
				parent.childIndex[c.Value] = j
			}
			break
		}
	}
	SetItemParent(item, nil)
}

// GetTerminalIndices returns the flat indices of all terminal items under root.
// C# source: MatrixHeader.GetTerminalIndices() — public overload.
func (h *MatrixHeader) GetTerminalIndices() []int {
	return terminalIndicesOf(h.Root)
}

// GetTerminalIndicesAt returns the flat indices of all terminal items under the
// node addressed by address. C# source: MatrixHeader.GetTerminalIndices(object[] address).
func (h *MatrixHeader) GetTerminalIndicesAt(address []any) []int {
	item := h.Find(address, false, 0)
	if item == nil {
		return nil
	}
	return terminalIndicesOf(item)
}

func terminalIndicesOf(root *HeaderItem) []int {
	terminals := ItemGetTerminalItems(root)
	result := make([]int, len(terminals))
	for i, t := range terminals {
		result[i] = ItemIndex(t)
	}
	return result
}

// AddTotalItems adds total summary items to the tree using the descriptor
// configuration (Totals, TotalsFirst, SuppressTotals, TemplateTotalColumn/Row/Cell).
// isTemplate true means we are building the design-time template (SuppressTotals ignored).
// C# source: MatrixHeader.AddTotalItems(bool isTemplate).
func (h *MatrixHeader) AddTotalItems(isTemplate bool) {
	h.addTotalItemsAt(h.Root, 0, isTemplate)
}

func (h *MatrixHeader) addTotalItemsAt(root *HeaderItem, descriptorIndex int, isTemplate bool) {
	if descriptorIndex >= len(h.Descriptors) {
		return
	}
	for _, child := range root.Children {
		h.addTotalItemsAt(child, descriptorIndex+1, isTemplate)
	}

	descr := h.Descriptors[descriptorIndex]
	if !descr.Totals {
		return
	}
	if !isTemplate && descr.SuppressTotals && len(root.Children) <= 1 {
		return
	}

	totalItem := newHeaderItem(root.Value)
	SetItemParent(totalItem, root)
	SetItemIsTotal(totalItem, true)
	SetItemAnyValue(totalItem, ItemAnyValue(root))
	SetItemDataRowNo(totalItem, ItemDataRowNo(root))
	SetItemIndex(totalItem, h.nextIndex)
	h.nextIndex++

	// Bind template fields if available.
	ext := descr.HeaderExt()
	_ = ext // TemplateTotal* fields are engine-side; recorded but not rendered here.

	if descr.TotalsFirst && !isTemplate {
		// Insert at front.
		root.Children = append([]*HeaderItem{totalItem}, root.Children...)
	} else {
		root.Children = append(root.Children, totalItem)
	}
	// Rebuild childIndex (totalItem may share a Value with siblings).
	root.childIndex = make(map[string]int, len(root.Children))
	for i, ch := range root.Children {
		root.childIndex[ch.Value] = i
	}
}

// ── DescriptorExt binding to table cells ─────────────────────────────────────

// SetDescriptorTemplates binds the table column/row/cell references from the
// MatrixHelper UpdateColumnDescriptors / UpdateRowDescriptors / UpdateCellDescriptors
// pass. In Go these are stored on the descriptor's HeaderDescriptorExt.
// This helper is provided so the engine (or tests) can wire them up.
func SetDescriptorTemplates(d *HeaderDescriptor, col *table.TableColumn, row *table.TableRow, cell *table.TableCell) {
	ext := d.HeaderExt()
	ext.TemplateColumn = col
	ext.TemplateRow = row
	ext.TemplateCell = cell
}

// SetDescriptorTotalTemplates binds the "total" table references.
func SetDescriptorTotalTemplates(d *HeaderDescriptor, col *table.TableColumn, row *table.TableRow, cell *table.TableCell) {
	ext := d.HeaderExt()
	ext.TemplateTotalColumn = col
	ext.TemplateTotalRow = row
	ext.TemplateTotalCell = cell
}
