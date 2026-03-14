package engine

import (
	"strings"

	"github.com/andrewloable/go-fastreport/band"
	"github.com/andrewloable/go-fastreport/expr"
)

// evalBandFilter evaluates the filter expression on the current row of a
// DataBand's data source. Returns true when the row passes (should be
// rendered), or true when no filter is set or the expression fails to compile.
//
// Filter expressions may reference columns using [BracketedName] syntax or
// bare identifiers. All columns from the current data-source row are injected
// into the evaluation environment before calling the expr evaluator.
func (e *ReportEngine) evalBandFilter(db *band.DataBand) bool {
	filterExpr := db.Filter()
	if filterExpr == "" {
		return true
	}

	ds := db.DataSourceRef()
	if ds == nil {
		return true
	}

	// Build environment from current row: extract column names from the filter
	// expression and populate the evaluator env.
	env := make(expr.Env)
	for _, col := range extractBracketedNames(filterExpr) {
		val, _ := ds.GetValue(col)
		env[col] = val
	}

	ev := expr.NewEvaluator(env)

	// Convert [ColName] tokens to bare identifiers so the expr evaluator can
	// find them in the env.
	evalExpr := convertBracketExpr(filterExpr)

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
