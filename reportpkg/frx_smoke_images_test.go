package reportpkg_test

// Smoke tests for Image and Picture FRX reports.

import (
	"testing"

	"github.com/andrewloable/go-fastreport/object"
)

func TestFRXSmoke_Image(t *testing.T) {
	r := loadFRXSmoke(t, "Image.frx")
	n := countObjectsOfType[*object.PictureObject](r)
	if n == 0 {
		t.Error("expected at least one PictureObject in Image.frx")
	}
}

func TestFRXSmoke_PicturesInsideTheMatrix(t *testing.T) {
	// PictureObjects are nested inside a MatrixObject; verify the file loads
	// without panic (the matrix traversal is not yet in countObjectsOfType).
	loadFRXSmoke(t, "Pictures Inside The Matrix.frx")
}

func TestFRXSmoke_TheUSAMap(t *testing.T) {
	// The USA Map uses MapObject which is not yet implemented;
	// verify the file loads without panic (unknown objects are skipped gracefully).
	loadFRXSmoke(t, "The USA Map.frx")
}
