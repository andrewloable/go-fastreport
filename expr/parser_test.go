package expr

import (
	"testing"
)

func TestParse_Empty(t *testing.T) {
	tokens := Parse("")
	if tokens != nil {
		t.Errorf("expected nil, got %v", tokens)
	}
}

func TestParse_NoExpression(t *testing.T) {
	tokens := Parse("Hello, world!")
	if len(tokens) != 1 {
		t.Fatalf("expected 1 token, got %d", len(tokens))
	}
	if tokens[0].IsExpr || tokens[0].Value != "Hello, world!" {
		t.Errorf("unexpected token: %+v", tokens[0])
	}
}

func TestParse_SingleExpression(t *testing.T) {
	tokens := Parse("[Name]")
	if len(tokens) != 1 {
		t.Fatalf("expected 1 token, got %d", len(tokens))
	}
	if !tokens[0].IsExpr || tokens[0].Value != "Name" {
		t.Errorf("unexpected token: %+v", tokens[0])
	}
}

func TestParse_MixedTokens(t *testing.T) {
	tokens := Parse("Hello [Name]!")
	if len(tokens) != 3 {
		t.Fatalf("expected 3 tokens, got %d: %v", len(tokens), tokens)
	}
	if tokens[0].IsExpr || tokens[0].Value != "Hello " {
		t.Errorf("token[0] unexpected: %+v", tokens[0])
	}
	if !tokens[1].IsExpr || tokens[1].Value != "Name" {
		t.Errorf("token[1] unexpected: %+v", tokens[1])
	}
	if tokens[2].IsExpr || tokens[2].Value != "!" {
		t.Errorf("token[2] unexpected: %+v", tokens[2])
	}
}

func TestParse_MultipleExpressions(t *testing.T) {
	tokens := Parse("Dear [FirstName] [LastName], total: [Sum]")
	if len(tokens) != 6 {
		t.Fatalf("expected 6 tokens, got %d: %v", len(tokens), tokens)
	}
	exprs := []string{"FirstName", "LastName", "Sum"}
	exprIdx := 0
	for _, tok := range tokens {
		if tok.IsExpr {
			if tok.Value != exprs[exprIdx] {
				t.Errorf("expr[%d]: expected %q, got %q", exprIdx, exprs[exprIdx], tok.Value)
			}
			exprIdx++
		}
	}
}

func TestParse_UnclosedBracket(t *testing.T) {
	// Unclosed bracket: rest should be literal.
	tokens := Parse("Hello [unclosed")
	// Should have "Hello " literal and then "[unclosed" as literal (no matching close).
	for _, tok := range tokens {
		if tok.IsExpr {
			t.Errorf("expected no expression tokens, got %+v", tok)
		}
	}
}

func TestParse_DoubleOpenBracketIsCompoundExpr(t *testing.T) {
	// In FastReport, "[[" is NOT an escape for a literal "[". It introduces a
	// compound expression where the outer "[...]" marks the expression and the
	// inner "[...]" are field references within it. This mirrors C# FastReport's
	// FindMatchingBrackets which uses pure depth tracking (no special "[[" handling).
	//
	// "Price: [[10]]" is parsed using depth tracking:
	//   - Depth 1 at first "[", depth 2 at second "[", depth 1 at first "]",
	//     depth 0 at second "]" → expression token "[10]".
	// So one literal "Price: " and one expression "[10]" are produced.
	tokens := Parse("Price: [[10]]")
	var exprTokens []Token
	var litTokens []Token
	for _, tok := range tokens {
		if tok.IsExpr {
			exprTokens = append(exprTokens, tok)
		} else {
			litTokens = append(litTokens, tok)
		}
	}
	// Must have exactly one expression token with value "[10]".
	if len(exprTokens) != 1 {
		t.Fatalf("expected 1 expression token, got %d: %+v", len(exprTokens), exprTokens)
	}
	if exprTokens[0].Value != "[10]" {
		t.Errorf("expr value: expected %q, got %q", "[10]", exprTokens[0].Value)
	}
	// Literal part must be "Price: ".
	litCombined := ""
	for _, tok := range litTokens {
		litCombined += tok.Value
	}
	if litCombined != "Price: " {
		t.Errorf("literal combined: expected %q, got %q", "Price: ", litCombined)
	}

	// Real-world compound expression: [[Products.ProductName].Substring(0,1)]
	tokens2 := Parse("[[Products.ProductName].Substring(0,1)]")
	if len(tokens2) != 1 || !tokens2[0].IsExpr {
		t.Fatalf("expected exactly one expression token, got %+v", tokens2)
	}
	if tokens2[0].Value != "[Products.ProductName].Substring(0,1)" {
		t.Errorf("compound expr: expected %q, got %q",
			"[Products.ProductName].Substring(0,1)", tokens2[0].Value)
	}
}

func TestParse_NoOpenBracketWithEscape(t *testing.T) {
	// Text with only ]] escape sequences and no open brackets.
	tokens := Parse("a]]b")
	if len(tokens) != 1 {
		t.Fatalf("expected 1 token, got %d: %v", len(tokens), tokens)
	}
	if tokens[0].IsExpr || tokens[0].Value != "a]b" {
		t.Errorf("unexpected: %+v", tokens[0])
	}
}

func TestParseWithBrackets_Custom(t *testing.T) {
	tokens := ParseWithBrackets("Hello {Name}!", "{", "}")
	if len(tokens) != 3 {
		t.Fatalf("expected 3 tokens, got %d", len(tokens))
	}
	if !tokens[1].IsExpr || tokens[1].Value != "Name" {
		t.Errorf("unexpected expr token: %+v", tokens[1])
	}
}

func TestParseWithBrackets_Empty(t *testing.T) {
	tokens := ParseWithBrackets("", "[", "]")
	if tokens != nil {
		t.Errorf("expected nil, got %v", tokens)
	}
}

func TestParseWithBrackets_NoClose(t *testing.T) {
	tokens := ParseWithBrackets("[open only", "[", "]")
	for _, tok := range tokens {
		if tok.IsExpr {
			t.Errorf("expected no expr tokens for unclosed bracket, got %+v", tok)
		}
	}
}

func TestParseWithBrackets_OnlyLiteral(t *testing.T) {
	tokens := ParseWithBrackets("just text", "[", "]")
	if len(tokens) != 1 || tokens[0].IsExpr {
		t.Errorf("expected single literal token, got %v", tokens)
	}
}

func TestExtractExpressions_Basic(t *testing.T) {
	exprs := ExtractExpressions("Hello [Name] and [Age]!")
	if len(exprs) != 2 || exprs[0] != "Name" || exprs[1] != "Age" {
		t.Errorf("unexpected: %v", exprs)
	}
}

func TestExtractExpressions_None(t *testing.T) {
	exprs := ExtractExpressions("no expressions here")
	if exprs != nil {
		t.Errorf("expected nil, got %v", exprs)
	}
}

func TestContainsExpression_True(t *testing.T) {
	if !ContainsExpression("Hello [Name]!") {
		t.Error("expected true")
	}
}

func TestContainsExpression_False_NoOpen(t *testing.T) {
	if ContainsExpression("Hello world!") {
		t.Error("expected false (no brackets)")
	}
}

func TestContainsExpression_False_EmptyBracket(t *testing.T) {
	// "[]" has open and close but no content between them.
	if ContainsExpression("[]") {
		t.Error("expected false for empty brackets")
	}
}

func TestContainsExpression_False_NoClose(t *testing.T) {
	if ContainsExpression("[noclose") {
		t.Error("expected false when no close bracket follows")
	}
}

func TestUnescapeBrackets_NoEscapes(t *testing.T) {
	s := UnescapeBrackets("hello world")
	if s != "hello world" {
		t.Errorf("unexpected: %q", s)
	}
}

func TestUnescapeBrackets_DoubleOpen(t *testing.T) {
	s := UnescapeBrackets("[[")
	if s != "[" {
		t.Errorf("expected %q, got %q", "[", s)
	}
}

func TestUnescapeBrackets_DoubleClose(t *testing.T) {
	s := UnescapeBrackets("]]")
	if s != "]" {
		t.Errorf("expected %q, got %q", "]", s)
	}
}

func TestUnescapeBrackets_Both(t *testing.T) {
	s := UnescapeBrackets("[[hello]]")
	if s != "[hello]" {
		t.Errorf("expected %q, got %q", "[hello]", s)
	}
}

func TestParse_ExpressionAtStart(t *testing.T) {
	tokens := Parse("[Name] hello")
	if len(tokens) != 2 {
		t.Fatalf("expected 2, got %d: %v", len(tokens), tokens)
	}
	if !tokens[0].IsExpr || tokens[0].Value != "Name" {
		t.Errorf("token[0]: %+v", tokens[0])
	}
	if tokens[1].IsExpr || tokens[1].Value != " hello" {
		t.Errorf("token[1]: %+v", tokens[1])
	}
}

func TestParse_ExpressionOnly(t *testing.T) {
	tokens := Parse("[Sum]")
	if len(tokens) != 1 || !tokens[0].IsExpr || tokens[0].Value != "Sum" {
		t.Errorf("unexpected: %v", tokens)
	}
}

func TestParse_NestedBrackets(t *testing.T) {
	// Nested brackets: "[IIF([x]>0, [x], 0)]" — the outer brackets form the expression.
	// The inner brackets are part of the expression content.
	tokens := Parse("[IIF([x]>0, 1, 0)]")
	if len(tokens) != 1 {
		t.Fatalf("expected 1 token, got %d: %v", len(tokens), tokens)
	}
	if !tokens[0].IsExpr || tokens[0].Value != "IIF([x]>0, 1, 0)" {
		t.Errorf("unexpected: %+v", tokens[0])
	}
}
