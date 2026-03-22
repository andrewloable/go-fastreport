package matrix

// header_item.go adds the fields and methods missing from the original Go
// HeaderItem (header_tree.go). These are ported from:
//   original-dotnet/FastReport.Base/Matrix/MatrixHeaderItem.cs
//
// Gaps implemented:
//   - Parent pointer (MatrixHeaderItem.Parent)
//   - Index field (the flat-array cell address; MatrixHeaderItem.Index)
//   - IsTotal flag (MatrixHeaderItem.IsTotal)
//   - DataRowNo field (MatrixHeaderItem.DataRowNo)
//   - PageBreak flag (MatrixHeaderItem.PageBreak)
//   - IsSplitted flag (MatrixHeaderItem.IsSplitted)
//   - Values property (returns ancestor path as []any)
//   - Find binary-search method (MatrixHeaderItem.Find)
//   - GetTerminalItems (returns leaf items ignoring IsSplitted nodes)
//
// Note: HeaderItem in header_tree.go uses string Values for its existing
// multi-level API (AddDataMultiLevel). This file adds the typed-value fields
// used by the AddValue/Find pipeline (MatrixHeader.Find in header_store.go).

import (
	"fmt"
	"sort"
)

// itemParent stores the parent pointer for each *HeaderItem.
// Stored externally to avoid modifying the existing HeaderItem struct in
// header_tree.go. Use ItemParent() and SetItemParent() to access.
var itemParentMap = make(map[*HeaderItem]*HeaderItem)

// ItemParent returns the parent of item, or nil if it is the root.
// C# source: MatrixHeaderItem.Parent property.
func ItemParent(item *HeaderItem) *HeaderItem {
	return itemParentMap[item]
}

// SetItemParent sets the parent pointer for item.
func SetItemParent(item *HeaderItem, parent *HeaderItem) {
	if parent == nil {
		delete(itemParentMap, item)
	} else {
		itemParentMap[item] = parent
	}
}

// itemIndexMap stores the flat cell-array index for each *HeaderItem.
// C# source: MatrixHeaderItem.Index property.
var itemIndexMap = make(map[*HeaderItem]int)

// ItemIndex returns the flat cell-array index for item (-1 if unset).
// C# source: MatrixHeaderItem.Index property.
func ItemIndex(item *HeaderItem) int {
	if v, ok := itemIndexMap[item]; ok {
		return v
	}
	return -1
}

// SetItemIndex sets the flat cell-array index for item.
func SetItemIndex(item *HeaderItem, index int) {
	itemIndexMap[item] = index
}

// itemIsTotalMap stores the IsTotal flag for each *HeaderItem.
// C# source: MatrixHeaderItem.IsTotal property.
var itemIsTotalMap = make(map[*HeaderItem]bool)

// ItemIsTotal reports whether item is a "total" summary row/column.
// C# source: MatrixHeaderItem.IsTotal property.
func ItemIsTotal(item *HeaderItem) bool {
	return itemIsTotalMap[item]
}

// SetItemIsTotal sets the IsTotal flag for item.
func SetItemIsTotal(item *HeaderItem, v bool) {
	if v {
		itemIsTotalMap[item] = true
	} else {
		delete(itemIsTotalMap, item)
	}
}

// itemDataRowNoMap stores the data source row number for each *HeaderItem.
// C# source: MatrixHeaderItem.DataRowNo property.
var itemDataRowNoMap = make(map[*HeaderItem]int)

// ItemDataRowNo returns the data source row number for item.
// C# source: MatrixHeaderItem.DataRowNo property.
func ItemDataRowNo(item *HeaderItem) int {
	return itemDataRowNoMap[item]
}

// SetItemDataRowNo sets the data source row number for item.
func SetItemDataRowNo(item *HeaderItem, n int) {
	itemDataRowNoMap[item] = n
}

// itemPageBreakMap stores the PageBreak flag for each *HeaderItem.
// C# source: MatrixHeaderItem.PageBreak property.
var itemPageBreakMap = make(map[*HeaderItem]bool)

// ItemPageBreak reports whether item should trigger a page break.
// C# source: MatrixHeaderItem.PageBreak property.
func ItemPageBreak(item *HeaderItem) bool {
	return itemPageBreakMap[item]
}

// SetItemPageBreak sets the PageBreak flag for item.
func SetItemPageBreak(item *HeaderItem, v bool) {
	if v {
		itemPageBreakMap[item] = true
	} else {
		delete(itemPageBreakMap, item)
	}
}

// itemIsSplittedMap stores the IsSplitted flag for each *HeaderItem.
// C# source: MatrixHeaderItem.IsSplitted property.
var itemIsSplittedMap = make(map[*HeaderItem]bool)

// ItemIsSplitted reports whether item is a duplicate inserted for row-splitting.
// C# source: MatrixHeaderItem.IsSplitted property (internal).
func ItemIsSplitted(item *HeaderItem) bool {
	return itemIsSplittedMap[item]
}

// SetItemIsSplitted sets the IsSplitted flag for item.
func SetItemIsSplitted(item *HeaderItem, v bool) {
	if v {
		itemIsSplittedMap[item] = true
	} else {
		delete(itemIsSplittedMap, item)
	}
}

// itemAnyValueMap stores the typed (any) value for header items used in the
// AddValue pipeline. The existing HeaderItem.Value holds the string display
// value; this map holds the original typed value for aggregation key matching.
// C# source: MatrixHeaderItem.Value property (object type in C#).
var itemAnyValueMap = make(map[*HeaderItem]any)

// ItemAnyValue returns the typed value for item (used in Find/binary-search).
// C# source: MatrixHeaderItem.Value (object).
func ItemAnyValue(item *HeaderItem) any {
	return itemAnyValueMap[item]
}

// SetItemAnyValue sets the typed value for item.
func SetItemAnyValue(item *HeaderItem, v any) {
	itemAnyValueMap[item] = v
}

// ItemValues returns the path of ancestor values from outermost to this item.
// C# source: MatrixHeaderItem.Values property.
func ItemValues(item *HeaderItem) []any {
	count := 0
	cur := item
	for ItemParent(cur) != nil {
		count++
		cur = ItemParent(cur)
	}
	if count == 0 {
		return nil
	}
	result := make([]any, count)
	cur = item
	idx := count - 1
	for ItemParent(cur) != nil {
		result[idx] = ItemAnyValue(cur)
		idx--
		cur = ItemParent(cur)
	}
	return result
}

// ItemSpan returns the number of leaf columns/rows this item spans.
// C# source: MatrixHeaderItem.Span property.
func ItemSpan(item *HeaderItem) int {
	terminals := ItemGetTerminalItems(item)
	return len(terminals)
}

// ItemClearChildren removes all children of item.
// C# source: MatrixHeaderItem.Clear().
func ItemClearChildren(item *HeaderItem) {
	item.Children = item.Children[:0]
	item.childIndex = make(map[string]int)
}

// ItemGetTerminalItems returns all leaf items (items with no children, and not
// IsSplitted). Mirrors the private GetTerminalItems helper in C#.
// C# source: MatrixHeaderItem.GetTerminalItems().
func ItemGetTerminalItems(item *HeaderItem) []*HeaderItem {
	var result []*HeaderItem
	itemGetTerminalItems(item, &result)
	return result
}

func itemGetTerminalItems(item *HeaderItem, out *[]*HeaderItem) {
	if len(item.Children) == 0 && !ItemIsSplitted(item) {
		*out = append(*out, item)
		return
	}
	for _, child := range item.Children {
		itemGetTerminalItems(child, out)
	}
}

// ItemFind searches for a child with the given value under item, respecting sort order.
// Returns the index of the found child (>= 0) or the bitwise-complement of the
// insertion point (< 0, equivalent to C#'s ~index from BinarySearch) if not found.
// SortOrderNone does a linear scan. Ascending/Descending use binary search.
// C# source: MatrixHeaderItem.Find(object value, SortOrder sort).
func ItemFind(item *HeaderItem, value any, sort SortOrder) int {
	if len(item.Children) == 0 {
		return ^0 // equivalent to ~0 == -1
	}

	if sort == SortOrderNone {
		// Linear scan comparing with ItemAnyValue.
		for i, child := range item.Children {
			cv := ItemAnyValue(child)
			if compareAny(cv, value) == 0 {
				return i
			}
		}
		return ^len(item.Children)
	}

	// Binary search.
	descending := sort == SortOrderDescending
	return binarySearchChildren(item.Children, value, descending)
}

// binarySearchChildren performs a binary search on children (sorted by
// ItemAnyValue) returning the index of the matching element or the
// bitwise-complement of the insertion point.
func binarySearchChildren(children []*HeaderItem, value any, descending bool) int {
	lo, hi := 0, len(children)-1
	for lo <= hi {
		mid := (lo + hi) / 2
		cmp := compareAny(ItemAnyValue(children[mid]), value)
		if descending {
			cmp = -cmp
		}
		switch {
		case cmp == 0:
			return mid
		case cmp < 0:
			lo = mid + 1
		default:
			hi = mid - 1
		}
	}
	return ^lo // insertion point as bitwise complement
}

// insertChildAt inserts newItem at position pos in item.Children, maintaining the
// childIndex map. Used by MatrixHeader.Find when creating new items in sorted order.
func insertChildAt(parent *HeaderItem, pos int, newItem *HeaderItem) {
	parent.Children = append(parent.Children, nil)
	copy(parent.Children[pos+1:], parent.Children[pos:])
	parent.Children[pos] = newItem
	// Rebuild childIndex map after structural change.
	parent.childIndex = make(map[string]int, len(parent.Children))
	for i, ch := range parent.Children {
		key := ch.Value
		parent.childIndex[key] = i
	}
}

// compareAny compares two values that may be comparable types.
// Returns negative if a < b, zero if a == b, positive if a > b.
// Mirrors the IComparable logic in C# HeaderComparer.
func compareAny(a, b any) int {
	if a == nil && b == nil {
		return 0
	}
	if a == nil {
		return 1 // nil sorts after non-nil (matches C# behavior: i1 is null → result = -1 meaning b > a)
	}
	if b == nil {
		return -1
	}
	// Try numeric comparisons first.
	af, aok := toFloat64Any(a)
	bf, bok := toFloat64Any(b)
	if aok && bok {
		switch {
		case af < bf:
			return -1
		case af > bf:
			return 1
		default:
			return 0
		}
	}
	// Fall back to string comparison.
	as := anyToString(a)
	bs := anyToString(b)
	switch {
	case as < bs:
		return -1
	case as > bs:
		return 1
	default:
		return 0
	}
}

func toFloat64Any(v any) (float64, bool) {
	switch x := v.(type) {
	case float64:
		return x, true
	case float32:
		return float64(x), true
	case int:
		return float64(x), true
	case int64:
		return float64(x), true
	case int32:
		return float64(x), true
	case int16:
		return float64(x), true
	case int8:
		return float64(x), true
	case uint:
		return float64(x), true
	case uint64:
		return float64(x), true
	case uint32:
		return float64(x), true
	}
	return 0, false
}

func anyToString(v any) string {
	if v == nil {
		return ""
	}
	if s, ok := v.(string); ok {
		return s
	}
	return fmt.Sprintf("%v", v)
}

// sortChildren sorts item.Children by ItemAnyValue according to sort order.
// Called internally when needed to maintain sorted invariant after insertion.
func sortChildren(item *HeaderItem, order SortOrder) {
	if order == SortOrderNone {
		return
	}
	descending := order == SortOrderDescending
	sort.SliceStable(item.Children, func(i, j int) bool {
		cmp := compareAny(ItemAnyValue(item.Children[i]), ItemAnyValue(item.Children[j]))
		if descending {
			return cmp > 0
		}
		return cmp < 0
	})
	// Rebuild childIndex.
	item.childIndex = make(map[string]int, len(item.Children))
	for i, ch := range item.Children {
		item.childIndex[ch.Value] = i
	}
}
