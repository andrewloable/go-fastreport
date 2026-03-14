package reportpkg_test

// Smoke tests for Matrix/Pivot table FRX reports.
// MatrixObject is a complex container not yet registered in the serial registry;
// these tests verify that FRX files load without panic and have at least one page.

import (
	"testing"
)

func TestFRXSmoke_Matrix(t *testing.T) {
	loadFRXSmoke(t, "Matrix.frx")
}

func TestFRXSmoke_SimpleMatrix(t *testing.T) {
	loadFRXSmoke(t, "Simple Matrix.frx")
}

func TestFRXSmoke_AdvancedMatrix(t *testing.T) {
	loadFRXSmoke(t, "Advanced Matrix.frx")
}

func TestFRXSmoke_MatrixWithColumnsOnly(t *testing.T) {
	loadFRXSmoke(t, "Matrix With Columns Only.frx")
}

func TestFRXSmoke_MatrixWithRowsOnly(t *testing.T) {
	loadFRXSmoke(t, "Matrix With Rows Only.frx")
}

func TestFRXSmoke_TwoColumnDimensions(t *testing.T) {
	loadFRXSmoke(t, "Two Column Dimensions.frx")
}

func TestFRXSmoke_TwoRowDimensions(t *testing.T) {
	loadFRXSmoke(t, "Two Row Dimensions.frx")
}

func TestFRXSmoke_TwoCellDimensions(t *testing.T) {
	loadFRXSmoke(t, "Two Cell Dimensions.frx")
}

func TestFRXSmoke_TwoCellDimensionsSideBySide(t *testing.T) {
	loadFRXSmoke(t, "Two Cell Dimensions, Side-by-Side.frx")
}

func TestFRXSmoke_ObjectsInsideTheMatrix(t *testing.T) {
	loadFRXSmoke(t, "Objects Inside The Matrix.frx")
}
