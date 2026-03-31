package expr

import "strings"

// Token represents a piece of parsed text.
type Token struct {
	// IsExpr is true when this token contains an expression (bracketed).
	IsExpr bool
	// Value is the literal text or expression string (without brackets).
	Value string
}

// Parse splits text into literal and expression tokens.
// Default brackets are "[" and "]".
// Example: "Hello [Name]!" → [{false,"Hello "}, {true,"Name"}, {false,"!"}]
func Parse(text string) []Token {
	return ParseWithBrackets(text, "[", "]")
}

// ParseWithBrackets is like Parse but uses custom bracket characters.
// It handles nested brackets by tracking bracket depth.
// Escaped sequences "[[" and "]]" (double-open / double-close) are treated as
// literal single bracket characters in the output.
func ParseWithBrackets(text, open, close string) []Token {
	if text == "" {
		return nil
	}

	var tokens []Token
	pos := 0
	n := len(text)

	for pos < n {
		// Look for the next open bracket.
		openIdx := strings.Index(text[pos:], open)
		if openIdx == -1 {
			// No more open brackets; the rest is a literal.
			lit := UnescapeBrackets(text[pos:])
			if lit != "" {
				tokens = append(tokens, Token{IsExpr: false, Value: lit})
			}
			break
		}
		openIdx += pos // absolute index

		// Note: "[[" is NOT an escape sequence in FastReport template text.
		// In FastReport .NET, "[[expr1] op [expr2]]" is a compound expression where
		// the outer "[...]" marks the expression and the inner "[...]" are field
		// references within it (mirrors C# CodeUtils.FindMatchingBrackets depth
		// tracking — no special handling for "[["). Pure depth tracking below
		// correctly finds the outermost matching "]" for any nesting level.

		// Emit any literal text before this open bracket.
		if openIdx > pos {
			lit := UnescapeBrackets(text[pos:openIdx])
			if lit != "" {
				tokens = append(tokens, Token{IsExpr: false, Value: lit})
			}
		}

		// Find the matching close bracket, respecting nesting.
		depth := 0
		scanPos := openIdx
		exprStart := openIdx + len(open)
		closeIdx := -1

		for scanPos < n {
			if strings.HasPrefix(text[scanPos:], open) {
				depth++
				scanPos += len(open)
			} else if strings.HasPrefix(text[scanPos:], close) {
				depth--
				if depth == 0 {
					closeIdx = scanPos
					break
				}
				scanPos += len(close)
			} else {
				scanPos++
			}
		}

		if closeIdx == -1 {
			// No matching close bracket; treat the rest as a literal.
			lit := UnescapeBrackets(text[openIdx:])
			if lit != "" {
				tokens = append(tokens, Token{IsExpr: false, Value: lit})
			}
			break
		}

		// The expression content (inside the outermost brackets).
		exprText := text[exprStart:closeIdx]
		tokens = append(tokens, Token{IsExpr: true, Value: exprText})
		pos = closeIdx + len(close)
	}

	return tokens
}

// ExtractExpressions returns only the expression strings from text.
func ExtractExpressions(text string) []string {
	tokens := Parse(text)
	var exprs []string
	for _, t := range tokens {
		if t.IsExpr {
			exprs = append(exprs, t.Value)
		}
	}
	return exprs
}

// ContainsExpression returns true if text contains at least one [expression]
// with non-empty content between the brackets.
func ContainsExpression(text string) bool {
	open := strings.Index(text, "[")
	if open == -1 {
		return false
	}
	// close > 1 ensures there is at least one character between "[" and "]".
	close := strings.Index(text[open:], "]")
	return close > 1
}

// UnescapeBrackets converts FastReport escape sequences in literal text.
// "[[" → "[" and "]]" → "]" per the .NET FastReport convention.
func UnescapeBrackets(text string) string {
	if !strings.Contains(text, "[[") && !strings.Contains(text, "]]") {
		return text
	}
	text = strings.ReplaceAll(text, "[[", "[")
	text = strings.ReplaceAll(text, "]]", "]")
	return text
}
