package object_test

// barcode_expressions_test.go tests BarcodeObject.GetExpressions().
//
// C# reference: BarcodeObject.cs:557–576 (BarcodeObject.GetExpressions).
// The method collects expression strings for dependency tracking:
//   1. Base: Hyperlink.Expression and Bookmark from ReportComponentBase.
//   2. DataColumn if non-empty.
//   3. Expression if non-empty; else bracket expressions extracted from Text
//      when AllowExpressions is true and Brackets is non-empty.

import (
	"testing"

	"github.com/andrewloable/go-fastreport/barcode"
	"github.com/andrewloable/go-fastreport/report"
)

// containsStr is a helper to check membership in a string slice.
func containsStr(slice []string, s string) bool {
	for _, v := range slice {
		if v == s {
			return true
		}
	}
	return false
}

// TestGetExpressions_Empty verifies that a default BarcodeObject with no
// DataColumn, Expression, or Text returns an empty slice.
func TestGetExpressions_Empty(t *testing.T) {
	bc := barcode.NewBarcodeObject()
	exprs := bc.GetExpressions()
	if len(exprs) != 0 {
		t.Errorf("expected empty slice, got %v", exprs)
	}
}

// TestGetExpressions_DataColumn verifies that a non-empty DataColumn is always
// included in the returned expressions.
// C# BarcodeObject.GetExpressions(): if DataColumn != "" → expressions.Add(DataColumn)
func TestGetExpressions_DataColumn(t *testing.T) {
	bc := barcode.NewBarcodeObject()
	bc.SetDataColumn("Products.Code")

	exprs := bc.GetExpressions()
	if !containsStr(exprs, "Products.Code") {
		t.Errorf("expected 'Products.Code' in expressions %v", exprs)
	}
}

// TestGetExpressions_Expression verifies that a non-empty Expression field
// is included and that Text bracket-extraction is skipped when Expression is set.
// C# BarcodeObject.GetExpressions(): if Expression != "" → add Expression (no Text scan).
func TestGetExpressions_Expression(t *testing.T) {
	bc := barcode.NewBarcodeObject()
	bc.SetExpression("[Order.ID]")
	bc.SetAllowExpressions(true)
	bc.SetText("[ShouldNotBeExtracted]")

	exprs := bc.GetExpressions()
	if !containsStr(exprs, "[Order.ID]") {
		t.Errorf("expected '[Order.ID]' in expressions %v", exprs)
	}
	if containsStr(exprs, "ShouldNotBeExtracted") {
		t.Errorf("did not expect 'ShouldNotBeExtracted' when Expression is set; got %v", exprs)
	}
}

// TestGetExpressions_AllowExpressions_BracketText verifies that when Expression
// is empty and AllowExpressions is true, bracket expressions are extracted from Text.
// C# BarcodeObject.GetExpressions(): CodeUtils.GetExpressions(Text, brackets[0], brackets[1])
func TestGetExpressions_AllowExpressions_BracketText(t *testing.T) {
	bc := barcode.NewBarcodeObject()
	bc.SetAllowExpressions(true)
	bc.SetBrackets("[,]")
	bc.SetText("prefix [Field1] middle [Field2] suffix")

	exprs := bc.GetExpressions()
	if !containsStr(exprs, "Field1") {
		t.Errorf("expected 'Field1' in expressions %v", exprs)
	}
	if !containsStr(exprs, "Field2") {
		t.Errorf("expected 'Field2' in expressions %v", exprs)
	}
}

// TestGetExpressions_AllowExpressionsDisabled verifies that when AllowExpressions
// is false, bracket expressions in Text are NOT extracted.
func TestGetExpressions_AllowExpressionsDisabled(t *testing.T) {
	bc := barcode.NewBarcodeObject()
	bc.SetAllowExpressions(false)
	bc.SetBrackets("[,]")
	bc.SetText("[SomeField]")

	exprs := bc.GetExpressions()
	if containsStr(exprs, "SomeField") {
		t.Errorf("expected no bracket extraction when AllowExpressions=false; got %v", exprs)
	}
}

// TestGetExpressions_DataColumnAndExpression verifies that both DataColumn and
// Expression are collected when both are non-empty.
// C# BarcodeObject.GetExpressions() adds DataColumn independently of Expression.
func TestGetExpressions_DataColumnAndExpression(t *testing.T) {
	bc := barcode.NewBarcodeObject()
	bc.SetDataColumn("Orders.ID")
	bc.SetExpression("[CustomExpr]")

	exprs := bc.GetExpressions()
	if !containsStr(exprs, "Orders.ID") {
		t.Errorf("expected 'Orders.ID' in expressions %v", exprs)
	}
	if !containsStr(exprs, "[CustomExpr]") {
		t.Errorf("expected '[CustomExpr]' in expressions %v", exprs)
	}
}

// TestGetExpressions_HyperlinkExpression verifies that a Hyperlink.Expression
// is included in the returned expressions (from base class).
// C# ReportComponentBase.GetExpressions(): adds Hyperlink.Expression if non-empty.
func TestGetExpressions_HyperlinkExpression(t *testing.T) {
	bc := barcode.NewBarcodeObject()
	bc.SetHyperlink(&report.Hyperlink{
		Expression: "[HyperlinkURL]",
	})

	exprs := bc.GetExpressions()
	if !containsStr(exprs, "[HyperlinkURL]") {
		t.Errorf("expected '[HyperlinkURL]' in expressions %v", exprs)
	}
}

// TestGetExpressions_Bookmark verifies that a non-empty Bookmark is included.
// C# ReportComponentBase.GetExpressions(): adds Bookmark if non-empty.
func TestGetExpressions_Bookmark(t *testing.T) {
	bc := barcode.NewBarcodeObject()
	bc.SetBookmark("AnchorBookmark")

	exprs := bc.GetExpressions()
	if !containsStr(exprs, "AnchorBookmark") {
		t.Errorf("expected 'AnchorBookmark' in expressions %v", exprs)
	}
}

// TestGetExpressions_NoBrackets_AllowExpressionsTrue verifies that when
// AllowExpressions is true but Brackets is empty, no extraction is attempted.
func TestGetExpressions_NoBrackets_AllowExpressionsTrue(t *testing.T) {
	bc := barcode.NewBarcodeObject()
	bc.SetAllowExpressions(true)
	bc.SetBrackets("")
	bc.SetText("[SomeField]")

	exprs := bc.GetExpressions()
	if containsStr(exprs, "SomeField") {
		t.Errorf("expected no extraction when Brackets is empty; got %v", exprs)
	}
}

// TestGetExpressions_TextNoBrackets verifies that plain text (no bracket expressions)
// with AllowExpressions=true returns an empty slice.
func TestGetExpressions_TextNoBrackets(t *testing.T) {
	bc := barcode.NewBarcodeObject()
	bc.SetAllowExpressions(true)
	bc.SetBrackets("[,]")
	bc.SetText("PLAINTEXT12345")

	exprs := bc.GetExpressions()
	if len(exprs) != 0 {
		t.Errorf("expected empty expressions for plain text, got %v", exprs)
	}
}
