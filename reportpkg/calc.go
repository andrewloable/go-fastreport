package reportpkg

import (
	"fmt"
	"strings"

	"github.com/andrewloable/go-fastreport/data"
	"github.com/andrewloable/go-fastreport/expr"
)

// calcContext holds the current-row data source injected via SetCalcContext.
type calcContext struct {
	ds data.DataSource
}

// calcContextKey is a placeholder type for the context map key.
type calcContextKey struct{}

// SetCalcContext injects a data source whose current row values are made
// available in expression evaluation as "SourceName.ColumnName".
// Pass nil to clear the context.
func (r *Report) SetCalcContext(ds data.DataSource) {
	r.calcDS = ds
}

// Calc evaluates a FastReport bracket expression and returns its value.
//
// The expression may be:
//   - A bare variable name: "PageNumber"
//   - A [bracketed] expression: "[PageNumber]"
//   - A compound expression: "[Field1] + ' ' + [Field2]"
//   - A data-source column reference: "[DataSource.Column]"
//
// The evaluation environment is built from:
//  1. Dictionary parameters (name → value)
//  2. Dictionary system variables (name → value)
//  3. Current-row data source values (if SetCalcContext was called)
//
// Column references use the form "Alias_Column" internally (dot replaced with
// underscore) so they work as valid identifiers in the expr language.
func (r *Report) Calc(expression string) (any, error) {
	expression = strings.TrimSpace(expression)
	if expression == "" {
		return nil, nil
	}

	env := r.buildCalcEnv()
	ev := expr.NewEvaluator(env)

	// If the whole expression is a single [bracketed] token, unwrap it.
	unwrapped := unwrapBrackets(expression)

	// Replace [Token] patterns with safe identifier forms.
	goExpr := translateExpression(unwrapped)

	return ev.Eval(goExpr)
}

// CalcText evaluates a text template that may contain multiple [bracket]
// expressions and concatenates the results into a string.
// Example: "Hello [Name]!" → "Hello World!"
func (r *Report) CalcText(template string) (string, error) {
	tokens := expr.Parse(template)
	if tokens == nil {
		return template, nil
	}

	var sb strings.Builder
	for _, tok := range tokens {
		if !tok.IsExpr {
			sb.WriteString(tok.Value)
			continue
		}
		val, err := r.Calc(tok.Value)
		if err != nil {
			// On error, emit the raw bracket expression.
			sb.WriteString("[")
			sb.WriteString(tok.Value)
			sb.WriteString("]")
			continue
		}
		sb.WriteString(fmt.Sprintf("%v", val))
	}
	return sb.String(), nil
}

// buildCalcEnv constructs the expression environment from the dictionary and
// the current calc context data source.
func (r *Report) buildCalcEnv() expr.Env {
	env := make(expr.Env)

	if r.dictionary != nil {
		// Parameters.
		for _, p := range r.dictionary.Parameters() {
			env[p.Name] = p.Value
		}
		// System variables.
		for _, sv := range r.dictionary.SystemVariables() {
			env[sv.Name] = sv.Value
		}
		// Totals (current accumulated value).
		for _, t := range r.dictionary.Totals() {
			env[t.Name] = t.Value
		}
	}

	// Current data source row values (injected per-row by the engine).
	// The engine injects a Columnar data source that also satisfies ColumnarDataSource.
	if r.calcDS != nil {
		if cds, ok := r.calcDS.(columnarDataSource); ok {
			for _, col := range cds.Columns() {
				val, _ := r.calcDS.GetValue(col.Name)
				key := sanitizeIdent(r.calcDS.Alias() + "_" + col.Name)
				env[key] = val
				// Also expose the bare column name for convenience.
				env[sanitizeIdent(col.Name)] = val
			}
		}
	}

	// Inject user-registered callback functions.
	// Each is wrapped as a variadic func(...any) (any, error) so that
	// expr-lang/expr can call them with any number of arguments.
	for name, fn := range r.customFunctions {
		fn := fn // capture loop variable
		env[name] = func(args ...any) (any, error) {
			return fn(args)
		}
	}

	return env
}

// unwrapBrackets removes a single enclosing [...] pair if the entire expression
// is bracketed. "[Foo]" → "Foo", "Foo" → "Foo".
func unwrapBrackets(s string) string {
	if len(s) >= 2 && s[0] == '[' && s[len(s)-1] == ']' {
		// Make sure it's a single balanced pair, not "[A] + [B]".
		depth := 0
		for i, ch := range s {
			if ch == '[' {
				depth++
			} else if ch == ']' {
				depth--
				if depth == 0 && i < len(s)-1 {
					// Closing bracket before the end → not a single pair.
					return s
				}
			}
		}
		return s[1 : len(s)-1]
	}
	return s
}

// translateExpression replaces [Token] occurrences within a Go expression with
// their sanitized identifier equivalents so the expr evaluator can handle them.
func translateExpression(s string) string {
	var out strings.Builder
	remaining := s
	for {
		start := strings.Index(remaining, "[")
		if start == -1 {
			out.WriteString(remaining)
			break
		}
		out.WriteString(remaining[:start])
		remaining = remaining[start+1:]
		end := strings.Index(remaining, "]")
		if end == -1 {
			// Malformed — just emit the rest.
			out.WriteString("[")
			out.WriteString(remaining)
			break
		}
		inner := remaining[:end]
		remaining = remaining[end+1:]
		out.WriteString(sanitizeIdent(inner))
	}
	return out.String()
}

// sanitizeIdent converts a token like "DataSource.Field" into a valid identifier
// "DataSource_Field" by replacing dots and spaces.
func sanitizeIdent(s string) string {
	s = strings.ReplaceAll(s, ".", "_")
	s = strings.ReplaceAll(s, " ", "_")
	return s
}

// columnarDataSource is satisfied by data sources that expose column metadata.
// BaseDataSource satisfies this interface.
type columnarDataSource interface {
	Columns() []data.Column
}
