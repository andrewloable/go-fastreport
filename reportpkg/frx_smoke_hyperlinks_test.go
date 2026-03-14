package reportpkg_test

// Smoke tests for Hyperlinks, Bookmarks and Outline FRX reports.

import (
	"testing"
)

func TestFRXSmoke_HyperlinksBookmarks(t *testing.T) {
	loadFRXSmoke(t, "Hyperlinks, Bookmarks.frx")
}

func TestFRXSmoke_ComplexHyperlinksOutlineTOC(t *testing.T) {
	loadFRXSmoke(t, "Complex (Hyperlinks, Outline, TOC).frx")
}

func TestFRXSmoke_Guidelines(t *testing.T) {
	loadFRXSmoke(t, "Guidelines.frx")
}
