package engine

import (
	"strconv"
	"strings"

	"github.com/andrewloable/go-fastreport/band"
	"github.com/andrewloable/go-fastreport/expr"
)

// evalBandFilter evaluates the filter expression on the current row of a
// DataBand's data source. Returns true when the row passes (should be
// rendered), or true when no filter is set or the expression fails to compile.
//
// Filter expressions may reference columns using [BracketedName] syntax or
// bare identifiers. Bracketed names may be qualified with a datasource alias
// (e.g. [Order Details.OrderID]); the column part after the last dot is used
// for GetValue.  All column values are injected into the environment under
// their sanitized (dots/spaces → underscores) key so they match the
// translated expression.
func (e *ReportEngine) evalBandFilter(db *band.DataBand) bool {
	filterExpr := db.Filter()
	if filterExpr == "" {
		return true
	}

	ds := db.DataSourceRef()
	if ds == nil {
		// When the report context is available, delegate to Report.Calc which
		// can resolve expressions against the current calc context.
		if e.report != nil {
			val, err := e.report.Calc(filterExpr)
			if err != nil {
				return true
			}
			if b, ok := val.(bool); ok {
				return b
			}
		}
		return true
	}

	// Build environment from current row.  Bracketed names like
	// [Order Details.OrderID] are stored under the sanitised key
	// "Order_Details_OrderID" so the translated expression can find them.
	// String values are coerced to int/float when they look numeric so that
	// filter expressions like [OrderID] == 10278 evaluate correctly even when
	// the datasource stores all values as strings (e.g. XML datasource).
	env := make(expr.Env)
	for _, qualName := range extractBracketedNames(filterExpr) {
		// Extract bare column name (part after last dot, or the whole name).
		colName := qualName
		if idx := strings.LastIndex(qualName, "."); idx >= 0 {
			colName = qualName[idx+1:]
		}
		val, _ := ds.GetValue(colName)
		val = coerceValue(val)
		safeKey := sanitizeFilterIdent(qualName)
		env[safeKey] = val
		// Also expose bare column name for simple [ColName] expressions.
		if colName != qualName {
			env[sanitizeFilterIdent(colName)] = val
		}
	}

	ev := expr.NewEvaluator(env)

	// Replace [Name] tokens with their sanitised identifier equivalents.
	evalExpr := sanitizeFilterExpr(filterExpr)

	result, err := ev.Eval(evalExpr)
	if err != nil {
		// On eval error pass the row through — don't silently drop data.
		return true
	}
	if b, ok := result.(bool); ok {
		return b
	}
	return true
}

// extractBracketedNames returns all identifiers found inside [brackets] in s.
func extractBracketedNames(s string) []string {
	var names []string
	for {
		start := strings.Index(s, "[")
		if start == -1 {
			break
		}
		end := strings.Index(s[start:], "]")
		if end == -1 {
			break
		}
		name := s[start+1 : start+end]
		if name != "" {
			names = append(names, name)
		}
		s = s[start+end+1:]
	}
	return names
}

// convertBracketExpr replaces [Name] tokens with bare Name identifiers so the
// expr-lang evaluator can resolve them as environment variables.
func convertBracketExpr(s string) string {
	var b strings.Builder
	for {
		start := strings.Index(s, "[")
		if start == -1 {
			b.WriteString(s)
			break
		}
		b.WriteString(s[:start])
		end := strings.Index(s[start:], "]")
		if end == -1 {
			b.WriteString(s[start:])
			break
		}
		b.WriteString(s[start+1 : start+end])
		s = s[start+end+1:]
	}
	return b.String()
}

// sanitizeFilterIdent converts a potentially qualified name like
// "Order Details.OrderID" to a safe identifier "Order_Details_OrderID".
func sanitizeFilterIdent(s string) string {
	s = strings.ReplaceAll(s, ".", "_")
	s = strings.ReplaceAll(s, " ", "_")
	return s
}

// coerceValue attempts to convert a string value to its natural numeric type
// so that filter expressions like [OrderID] == 10278 evaluate correctly when
// the underlying datasource stores everything as strings (e.g. XML datasource).
func coerceValue(v any) any {
	s, ok := v.(string)
	if !ok {
		return v // already a non-string type — return as-is
	}
	// Try integer first.
	if i, err := strconv.ParseInt(s, 10, 64); err == nil {
		return i
	}
	// Try float.
	if f, err := strconv.ParseFloat(s, 64); err == nil {
		return f
	}
	return s
}

// sanitizeFilterExpr replaces [Name] tokens in a filter expression with their
// sanitised identifier equivalents so the expr-lang evaluator can find them.
func sanitizeFilterExpr(s string) string {
	var b strings.Builder
	for {
		start := strings.Index(s, "[")
		if start == -1 {
			b.WriteString(s)
			break
		}
		b.WriteString(s[:start])
		end := strings.Index(s[start:], "]")
		if end == -1 {
			b.WriteString(s[start:])
			break
		}
		inner := s[start+1 : start+end]
		b.WriteString(sanitizeFilterIdent(inner))
		s = s[start+end+1:]
	}
	return b.String()
}
