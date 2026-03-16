// Package preview internal tests — covers nil-guard branches in
// encodeOutlineItem and decodeOutlineItem that are unreachable via the public
// Save/Load API (gob rejects nil pointer elements in slices).
package preview

import "testing"

// TestEncodeOutlineItem_Nil calls encodeOutlineItem(nil) directly to cover the
// early-return nil guard on line 264.
func TestEncodeOutlineItem_Nil(t *testing.T) {
	result := encodeOutlineItem(nil)
	if result != nil {
		t.Errorf("encodeOutlineItem(nil) = %v, want nil", result)
	}
}

// TestEncodeOutlineItem_WithNilChild calls encodeOutlineItem with an
// OutlineItem that has a nil pointer in its Children slice, causing the
// recursive call encodeOutlineItem(nil) to be exercised again inside the loop.
func TestEncodeOutlineItem_WithNilChild(t *testing.T) {
	item := &OutlineItem{
		Text:     "parent",
		Children: []*OutlineItem{nil},
	}
	result := encodeOutlineItem(item)
	if result == nil {
		t.Fatal("encodeOutlineItem(non-nil item) = nil, want non-nil")
	}
	if result.Text != "parent" {
		t.Errorf("result.Text = %q, want parent", result.Text)
	}
	// The nil child should produce a nil fpxOutlineItem in Children.
	if len(result.Children) != 1 {
		t.Fatalf("result.Children len = %d, want 1", len(result.Children))
	}
	if result.Children[0] != nil {
		t.Errorf("result.Children[0] = %v, want nil", result.Children[0])
	}
}

// TestDecodeOutlineItem_NilSrc calls decodeOutlineItem with a nil src to cover
// the nil guard (src == nil) on line 398.
func TestDecodeOutlineItem_NilSrc(t *testing.T) {
	dst := &OutlineItem{}
	decodeOutlineItem(nil, dst) // must not panic
	if dst.Text != "" {
		t.Errorf("dst.Text = %q after nil src, want empty", dst.Text)
	}
}

// TestDecodeOutlineItem_NilDst calls decodeOutlineItem with a nil dst to cover
// the nil guard (dst == nil) on line 398.
func TestDecodeOutlineItem_NilDst(t *testing.T) {
	src := &fpxOutlineItem{Text: "hello"}
	decodeOutlineItem(src, nil) // must not panic
}

// TestDecodeOutlineItem_BothNil calls decodeOutlineItem with both nil arguments
// to cover the combined nil guard.
func TestDecodeOutlineItem_BothNil(t *testing.T) {
	decodeOutlineItem(nil, nil) // must not panic
}

// TestDecodeOutlineItem_WithNilChild decodes an fpxOutlineItem that has a nil
// child pointer in its Children slice, exercising the recursive
// decodeOutlineItem(nil, child) call path.
func TestDecodeOutlineItem_WithNilChild(t *testing.T) {
	src := &fpxOutlineItem{
		Text:     "root",
		Children: []*fpxOutlineItem{nil},
	}
	dst := &OutlineItem{}
	decodeOutlineItem(src, dst)

	if dst.Text != "root" {
		t.Errorf("dst.Text = %q, want root", dst.Text)
	}
	// The nil child should add one entry to dst.Children (an empty OutlineItem
	// that was allocated but then skipped by the nil src guard in the recursive call).
	if len(dst.Children) != 1 {
		t.Fatalf("dst.Children len = %d, want 1", len(dst.Children))
	}
}
