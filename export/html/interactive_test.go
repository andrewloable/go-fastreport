// Tests for HTML export interactive features: bookmarks, hyperlink targets.
// C# reference: HTMLExportLayers.cs GetHref, ExportObject.
package html_test

import (
	"strings"
	"testing"

	"github.com/andrewloable/go-fastreport/preview"
)

// ── Bookmark anchor emission ──────────────────────────────────────────────────

// TestBookmarkAnchor_EmittedBeforeObject verifies that an object with a non-empty
// Bookmark field causes a <a name="..."> anchor to be emitted before the object div.
// C# reference: HTMLExportLayers.cs ExportObject →
//
//	if (!String.IsNullOrEmpty(obj.Bookmark)) htmlPage.Append("<a name=\"...\">");
func TestBookmarkAnchor_EmittedBeforeObject(t *testing.T) {
	pp := buildPage([]preview.PreparedObject{
		{
			Kind:     preview.ObjectTypeText,
			Left:     0, Top: 0, Width: 100, Height: 20,
			Text:     "Section start",
			Bookmark: "SectionA",
		},
	})
	out := exportHTML(t, pp)
	if !strings.Contains(out, `<a name="SectionA">`) {
		t.Errorf("bookmark anchor: expected <a name=\"SectionA\">, got %q", out)
	}
}

// TestBookmarkAnchor_AppearsBeforeObjectDiv checks that the anchor precedes the
// object's content div.
func TestBookmarkAnchor_AppearsBeforeObjectDiv(t *testing.T) {
	pp := buildPage([]preview.PreparedObject{
		{
			Kind:     preview.ObjectTypeText,
			Left:     0, Top: 0, Width: 100, Height: 20,
			Text:     "Section start",
			Bookmark: "SectionA",
		},
	})
	out := exportHTML(t, pp)
	anchorIdx := strings.Index(out, `<a name="SectionA">`)
	divIdx := strings.Index(out, "Section start")
	if anchorIdx < 0 {
		t.Fatal("bookmark anchor: <a name> not found")
	}
	if divIdx < 0 {
		t.Fatal("bookmark anchor: object text not found")
	}
	if anchorIdx > divIdx {
		t.Errorf("bookmark anchor: <a name> appears after object text (anchor=%d, text=%d)", anchorIdx, divIdx)
	}
}

// TestBookmarkAnchor_MultipleObjects verifies that only the object with a Bookmark
// produces an anchor; objects without one do not.
func TestBookmarkAnchor_MultipleObjects(t *testing.T) {
	pp := buildPage([]preview.PreparedObject{
		{
			Kind:     preview.ObjectTypeText,
			Left:     0, Top: 0, Width: 100, Height: 20,
			Text:     "Anchored",
			Bookmark: "AnchorHere",
		},
		{
			Kind:   preview.ObjectTypeText,
			Left:   0, Top: 25, Width: 100, Height: 20,
			Text:   "No anchor",
			// no Bookmark field
		},
	})
	out := exportHTML(t, pp)
	// One anchor for the first object.
	count := strings.Count(out, `<a name="`)
	// Page N anchors are also present (e.g. PageN1); count only object bookmarks.
	anchorCount := strings.Count(out, `<a name="AnchorHere">`)
	if anchorCount != 1 {
		t.Errorf("bookmark anchor: expected 1 AnchorHere anchor, got %d (full count=%d)", anchorCount, count)
	}
	// No spurious anchor for the second object.
	if strings.Contains(out, `<a name="No anchor">`) {
		t.Errorf("bookmark anchor: unexpected anchor for non-bookmark object")
	}
}

// TestBookmarkAnchor_EmptyBookmark verifies no spurious anchor is emitted when
// the Bookmark field is empty.
func TestBookmarkAnchor_EmptyBookmark(t *testing.T) {
	pp := buildPage([]preview.PreparedObject{
		{
			Kind:     preview.ObjectTypeText,
			Left:     0, Top: 0, Width: 100, Height: 20,
			Text:     "Normal text",
			Bookmark: "",
		},
	})
	out := exportHTML(t, pp)
	// No object bookmark anchor — only the page PageN1 anchor should exist.
	if strings.Contains(out, `<a name=""`) {
		t.Errorf("bookmark anchor: unexpected <a name=\"\"> in output: %q", out)
	}
}

// TestBookmarkAnchor_HtmlEscaped verifies that special HTML characters in bookmark
// names are escaped in the anchor.
func TestBookmarkAnchor_HtmlEscaped(t *testing.T) {
	pp := buildPage([]preview.PreparedObject{
		{
			Kind:     preview.ObjectTypeText,
			Left:     0, Top: 0, Width: 100, Height: 20,
			Text:     "Escaped",
			Bookmark: "a&b",
		},
	})
	out := exportHTML(t, pp)
	if !strings.Contains(out, `<a name="a&amp;b">`) {
		t.Errorf("bookmark anchor: expected HTML-escaped anchor, got %q", out)
	}
}

// TestBookmarkAnchor_NonTextObject verifies that bookmark anchors also work on
// non-text objects (e.g. picture, shape).
func TestBookmarkAnchor_NonTextObject(t *testing.T) {
	pp := buildPage([]preview.PreparedObject{
		{
			Kind:      preview.ObjectTypeShape,
			Left:      0, Top: 0, Width: 100, Height: 50,
			ShapeKind: 0, // Rectangle
			Bookmark:  "ShapeAnchor",
		},
	})
	out := exportHTML(t, pp)
	if !strings.Contains(out, `<a name="ShapeAnchor">`) {
		t.Errorf("bookmark anchor on shape: expected <a name=\"ShapeAnchor\">, got %q", out)
	}
}

// ── Hyperlink target attribute ────────────────────────────────────────────────

// TestHyperlinkTarget_Blank verifies that HyperlinkTarget="_blank" produces
// target="_blank" on the anchor tag.
// C# reference: HTMLExportLayers.cs GetHref →
//
//	(obj.Hyperlink.OpenLinkInNewTab ? "target=\"_blank\"" : "")
func TestHyperlinkTarget_Blank(t *testing.T) {
	pp := buildPage([]preview.PreparedObject{
		{
			Kind:            preview.ObjectTypeText,
			Left:            0, Top: 0, Width: 100, Height: 20,
			Text:            "Open in new tab",
			HyperlinkKind:   1,
			HyperlinkValue:  "https://example.com",
			HyperlinkTarget: "_blank",
		},
	})
	out := exportHTML(t, pp)
	if !strings.Contains(out, `target="_blank"`) {
		t.Errorf("hyperlink target: expected target=\"_blank\", got %q", out)
	}
	if !strings.Contains(out, `href="https://example.com"`) {
		t.Errorf("hyperlink target: expected href, got %q", out)
	}
}

// TestHyperlinkTarget_Empty verifies that an empty HyperlinkTarget produces no
// target attribute on the anchor tag.
func TestHyperlinkTarget_Empty(t *testing.T) {
	pp := buildPage([]preview.PreparedObject{
		{
			Kind:            preview.ObjectTypeText,
			Left:            0, Top: 0, Width: 100, Height: 20,
			Text:            "No target",
			HyperlinkKind:   1,
			HyperlinkValue:  "https://example.com",
			HyperlinkTarget: "",
		},
	})
	out := exportHTML(t, pp)
	if strings.Contains(out, "target=") {
		t.Errorf("hyperlink target: unexpected target attribute in output: %q", out)
	}
}

// TestHyperlinkTarget_Self verifies that a non-blank target value is also emitted.
func TestHyperlinkTarget_Self(t *testing.T) {
	pp := buildPage([]preview.PreparedObject{
		{
			Kind:            preview.ObjectTypeText,
			Left:            0, Top: 0, Width: 100, Height: 20,
			Text:            "Self target",
			HyperlinkKind:   1,
			HyperlinkValue:  "https://example.com",
			HyperlinkTarget: "_self",
		},
	})
	out := exportHTML(t, pp)
	if !strings.Contains(out, `target="_self"`) {
		t.Errorf("hyperlink target _self: expected target=\"_self\", got %q", out)
	}
}

// TestHyperlinkTarget_BookmarkKind verifies target is emitted for bookmark-kind
// hyperlinks too (not just URL kind).
func TestHyperlinkTarget_BookmarkKind(t *testing.T) {
	pp := buildPage([]preview.PreparedObject{
		{
			Kind:            preview.ObjectTypeText,
			Left:            0, Top: 0, Width: 100, Height: 20,
			Text:            "Go to section",
			HyperlinkKind:   3,
			HyperlinkValue:  "SectionA",
			HyperlinkTarget: "_blank",
		},
	})
	out := exportHTML(t, pp)
	if !strings.Contains(out, `href="#SectionA"`) {
		t.Errorf("bookmark href: expected #SectionA, got %q", out)
	}
	if !strings.Contains(out, `target="_blank"`) {
		t.Errorf("bookmark target: expected target=\"_blank\", got %q", out)
	}
}

// ── Combined bookmark anchor + hyperlink ─────────────────────────────────────

// TestBookmarkAndHyperlink_Combined verifies that an object can simultaneously
// define a bookmark anchor (navigation target) and a hyperlink (navigation link).
func TestBookmarkAndHyperlink_Combined(t *testing.T) {
	pp := buildPage([]preview.PreparedObject{
		{
			Kind:            preview.ObjectTypeText,
			Left:            0, Top: 0, Width: 150, Height: 20,
			Text:            "Linked section",
			Bookmark:        "MySectionAnchor",
			HyperlinkKind:   1,
			HyperlinkValue:  "https://example.com",
			HyperlinkTarget: "_blank",
		},
	})
	out := exportHTML(t, pp)
	// Bookmark anchor should appear.
	if !strings.Contains(out, `<a name="MySectionAnchor">`) {
		t.Errorf("combined: expected <a name=\"MySectionAnchor\">, got %q", out)
	}
	// Hyperlink with target should appear.
	if !strings.Contains(out, `href="https://example.com"`) {
		t.Errorf("combined: expected href, got %q", out)
	}
	if !strings.Contains(out, `target="_blank"`) {
		t.Errorf("combined: expected target=\"_blank\", got %q", out)
	}
}
