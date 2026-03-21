package reportpkg

import (
	"fmt"
	"strconv"
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

	// If the whole expression is a single [bracketed] token, unwrap it first so
	// we can check for special-character names before building the full env.
	unwrapped := unwrapBrackets(expression)
	wasSingleBracket := unwrapped != expression

	// Special handling for names that contain '#' (FastReport macro variables
	// like CopyName#, Page#, TotalPages#, Row#, AbsRow#).  The expr-lang library
	// treats '#' as invalid syntax, so we must resolve these directly from the
	// dictionary system variables rather than passing them to the evaluator.
	// This applies both when called with the raw name (from CalcText token) and
	// when called with a bracketed form like "[CopyName#]".
	if strings.Contains(unwrapped, "#") {
		if r.dictionary != nil {
			for _, sv := range r.dictionary.SystemVariables() {
				if strings.EqualFold(sv.Name, unwrapped) {
					return sv.Value, nil
				}
			}
		}
		// Unknown #-name: return nil with an error so CalcText can preserve [name].
		return nil, fmt.Errorf("unknown macro variable %q", unwrapped)
	}

	env := r.buildCalcEnv()
	ev := expr.NewEvaluator(env)

	// Replace [Token] patterns with safe identifier forms.
	goExpr := translateExpression(unwrapped)

	// If the expression is a simple dotted identifier (e.g. "Report.ReportInfo.Description"),
	// sanitize dots to underscores so it matches keys in the environment.
	if isSimpleDottedIdent(goExpr) {
		sanitized := sanitizeIdent(goExpr)
		if _, exists := env[sanitized]; exists {
			goExpr = sanitized
		}
	} else if wasSingleBracket {
		// The unwrapped token may contain spaces (e.g. "Order Details.Products.ProductName").
		// Sanitize it and check the env before passing to the evaluator.
		sanitized := sanitizeIdent(unwrapped)
		if _, exists := env[sanitized]; exists {
			goExpr = sanitized
		}
	}

	val, err := ev.Eval(goExpr)
	if err != nil {
		return val, err
	}
	// Fire OnCustomCalc hook if set — mirrors C# Report.CustomCalc event which
	// allows callers to override the resolved value after expression evaluation.
	// C# ref: FastReport.Base/Report.cs, Calc() → CustomCalc event firing.
	if r.OnCustomCalc != nil {
		val = r.OnCustomCalc(expression, val)
	}
	return val, nil
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
		// Pass the bracketed form so that Calc's unwrapBrackets + wasSingleBracket
		// logic can sanitize names containing spaces (e.g. "Order Details.Orders.ShipName").
		val, err := r.Calc("[" + tok.Value + "]")
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
			// Also store under sanitized key so names with hyphens/spaces resolve.
			if key := sanitizeIdent(p.Name); key != p.Name {
				env[key] = p.Value
			}
		}
		// System variables.
		for _, sv := range r.dictionary.SystemVariables() {
			env[sv.Name] = sv.Value
			if key := sanitizeIdent(sv.Name); key != sv.Name {
				env[key] = sv.Value
			}
		}
		// Totals (current accumulated value).
		for _, t := range r.dictionary.Totals() {
			env[t.Name] = t.Value
			// Also store under sanitized key (e.g. "Sub-Total" → "Sub_Total").
			if key := sanitizeIdent(t.Name); key != t.Name {
				env[key] = t.Value
			}
		}
	}

	// Current data source row values (injected per-row by the engine).
	// The engine injects a Columnar data source that also satisfies ColumnarDataSource.
	if r.calcDS != nil {
		if cds, ok := r.calcDS.(columnarDataSource); ok {
			for _, col := range cds.Columns() {
				val, _ := r.calcDS.GetValue(col.Name)
				val = coerceCalcValue(val)
				key := sanitizeIdent(r.calcDS.Alias() + "_" + col.Name)
				env[key] = val
				// Also expose the bare column name for convenience.
				env[sanitizeIdent(col.Name)] = val
			}
		}

		// Traverse relations: for relations where the current data source is the
		// child, find the matching parent row and inject parent fields using the
		// naming convention "CurrentAlias_ParentAlias_ColumnName".
		// This enables expressions like [Order Details.Products.ProductName] to
		// resolve when iterating Order Details rows.
		if r.dictionary != nil {
			r.injectRelatedFields(env, r.calcDS)
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

// isSimpleDottedIdent reports whether s is a dotted identifier chain like
// "Report.ReportInfo.Description" (letters, digits, underscores, and dots only).
func isSimpleDottedIdent(s string) bool {
	if s == "" || !strings.Contains(s, ".") {
		return false
	}
	for _, r := range s {
		if !(r >= 'a' && r <= 'z') && !(r >= 'A' && r <= 'Z') && !(r >= '0' && r <= '9') && r != '_' && r != '.' {
			return false
		}
	}
	return true
}

// sanitizeIdent converts a token like "DataSource.Field" into a valid Go
// identifier by replacing dots, spaces, and hyphens with underscores.
// e.g. "Order Details.Orders.ShipName" → "Order_Details_Orders_ShipName"
//      "Sub-Total" → "Sub_Total"
func sanitizeIdent(s string) string {
	s = strings.ReplaceAll(s, ".", "_")
	s = strings.ReplaceAll(s, " ", "_")
	s = strings.ReplaceAll(s, "-", "_")
	return s
}

// columnarDataSource is satisfied by data sources that expose column metadata.
// BaseDataSource satisfies this interface.
type columnarDataSource interface {
	Columns() []data.Column
}

// coerceCalcValue converts string values to their natural numeric type so
// that arithmetic expressions like [UnitPrice] * [Quantity] evaluate correctly
// when the underlying datasource stores all values as strings (e.g. XML).
func coerceCalcValue(v any) any {
	s, ok := v.(string)
	if !ok {
		return v
	}
	if i, err := strconv.ParseInt(s, 10, 64); err == nil {
		return i
	}
	if f, err := strconv.ParseFloat(s, 64); err == nil {
		return f
	}
	return s
}

// injectRelatedFields traverses all relations in the dictionary where currentDS
// is the child data source. For each such relation it seeks the parent data
// source to the first row whose join-key columns match the current child row
// values, then injects the parent row's columns into env under the key
// "ChildAlias_ParentAlias_ColumnName" (dots replaced with underscores).
//
// This enables expressions like [Order Details.Products.ProductName] to resolve
// when iterating Order Details rows, given a relation Products→Order Details on
// ProductID.
func (r *Report) injectRelatedFields(env expr.Env, currentDS data.DataSource) {
	dict := r.dictionary
	currentAlias := currentDS.Alias()
	currentName := currentDS.Name()

	for _, rel := range dict.Relations() {
		// Determine the child and parent datasources for this relation.
		// Relations loaded from FRX use only ChildSourceName/ParentSourceName;
		// the resolved pointer fields (ChildDataSource/ParentDataSource) may be
		// nil. We resolve both by falling back to dictionary lookup.

		var childAlias, parentAlias string
		var parentDS data.DataSource

		if rel.ChildDataSource != nil && rel.ParentDataSource != nil {
			childAlias = rel.ChildDataSource.Alias()
			parentDS = rel.ParentDataSource
			parentAlias = parentDS.Alias()
		} else if rel.ChildSourceName != "" && rel.ParentSourceName != "" {
			// FRX-loaded relations: resolve by name.
			childAlias = rel.ChildSourceName
			parentDS = dict.FindDataSourceByAlias(rel.ParentSourceName)
			if parentDS == nil {
				continue
			}
			parentAlias = rel.ParentSourceName
		} else {
			continue
		}

		// Use column name lists — fall back to source names when not set.
		parentCols := rel.ParentColumns
		childCols := rel.ChildColumns
		if len(parentCols) == 0 {
			parentCols = rel.ParentColumnNames
		}
		if len(childCols) == 0 {
			childCols = rel.ChildColumnNames
		}
		if len(parentCols) == 0 || len(childCols) == 0 {
			continue
		}

		// We handle the case where currentDS is the child (detail) data source.
		if !strings.EqualFold(childAlias, currentAlias) &&
			!strings.EqualFold(childAlias, currentName) {
			continue
		}

		// Read the child join-key values from the current row.
		childVals := make([]string, len(childCols))
		for i, col := range childCols {
			v, _ := currentDS.GetValue(col)
			childVals[i] = fmt.Sprintf("%v", v)
		}

		// Seek parentDS to the row whose parent join-key columns match childVals.
		if err := parentDS.First(); err != nil {
			continue
		}
		found := false
		for !parentDS.EOF() {
			match := true
			for i, col := range parentCols {
				v, _ := parentDS.GetValue(col)
				if fmt.Sprintf("%v", v) != childVals[i] {
					match = false
					break
				}
			}
			if match {
				found = true
				break
			}
			if err := parentDS.Next(); err != nil {
				break
			}
		}

		if !found {
			continue
		}

		// Inject parent columns under "CurrentAlias_ParentAlias_ColumnName".
		if pcds, ok := parentDS.(columnarDataSource); ok {
			for _, col := range pcds.Columns() {
				val, _ := parentDS.GetValue(col.Name)
				val = coerceCalcValue(val)
				key := sanitizeIdent(currentAlias + "_" + parentAlias + "_" + col.Name)
				env[key] = val
			}
		}

		// Recursively inject grand-parent fields by treating parentDS as the new
		// current datasource and looking for its own parent relations.  This
		// enables 3-hop expressions like [Order Details.Orders.Shippers.Phone].
		if len(env) > 0 {
			r.injectRelatedFieldsFrom(env, parentDS, parentAlias, currentAlias)
		}
	}
}

// injectRelatedFieldsFrom is a recursive helper used by injectRelatedFields to
// traverse relation chains beyond one hop (e.g. grandparent fields).
// It looks for relations where grandparentDS is the child and injects
// grandparent-of-grandparent fields under the key
// "childAncestorAlias_parentAlias_grandparentAlias_ColumnName".
func (r *Report) injectRelatedFieldsFrom(env expr.Env, parentDS data.DataSource, parentAlias, origChildAlias string) {
	dict := r.dictionary
	parentName := parentDS.Name()

	for _, rel := range dict.Relations() {
		var childAlias2, grandParentAlias string
		var grandParentDS data.DataSource

		if rel.ChildDataSource != nil && rel.ParentDataSource != nil {
			childAlias2 = rel.ChildDataSource.Alias()
			grandParentDS = rel.ParentDataSource
			grandParentAlias = grandParentDS.Alias()
		} else if rel.ChildSourceName != "" && rel.ParentSourceName != "" {
			childAlias2 = rel.ChildSourceName
			grandParentDS = dict.FindDataSourceByAlias(rel.ParentSourceName)
			if grandParentDS == nil {
				continue
			}
			grandParentAlias = rel.ParentSourceName
		} else {
			continue
		}

		if !strings.EqualFold(childAlias2, parentAlias) &&
			!strings.EqualFold(childAlias2, parentName) {
			continue
		}

		parentCols := rel.ParentColumns
		childCols := rel.ChildColumns
		if len(parentCols) == 0 {
			parentCols = rel.ParentColumnNames
		}
		if len(childCols) == 0 {
			childCols = rel.ChildColumnNames
		}
		if len(parentCols) == 0 || len(childCols) == 0 {
			continue
		}

		// Read join-key values from parentDS (now acting as the child in this hop).
		childVals := make([]string, len(childCols))
		for i, col := range childCols {
			v, _ := parentDS.GetValue(col)
			childVals[i] = fmt.Sprintf("%v", v)
		}

		if err := grandParentDS.First(); err != nil {
			continue
		}
		found := false
		for !grandParentDS.EOF() {
			match := true
			for i, col := range parentCols {
				v, _ := grandParentDS.GetValue(col)
				if fmt.Sprintf("%v", v) != childVals[i] {
					match = false
					break
				}
			}
			if match {
				found = true
				break
			}
			if err := grandParentDS.Next(); err != nil {
				break
			}
		}

		if !found {
			continue
		}

		// Inject as "origChild_parent_grandparent_ColumnName".
		if gcds, ok := grandParentDS.(columnarDataSource); ok {
			for _, col := range gcds.Columns() {
				val, _ := grandParentDS.GetValue(col.Name)
				val = coerceCalcValue(val)
				key := sanitizeIdent(origChildAlias + "_" + parentAlias + "_" + grandParentAlias + "_" + col.Name)
				env[key] = val
			}
		}
	}
}
