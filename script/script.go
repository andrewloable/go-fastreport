// Package script provides a limited C# script evaluator for go-fastreport.
// It parses C# method bodies found in FRX <ScriptText> and executes them
// at report render time to apply BeforePrint event logic.
//
// Only the patterns needed to support cell-indicator scripts are handled;
// unrecognised statements are silently skipped.
package script

import (
	"image/color"
	"strconv"
	"strings"

	"github.com/andrewloable/go-fastreport/style"
)

// ContextObject allows getting and setting properties on a report object
// from within an executing script method.
type ContextObject interface {
	ScriptGetProperty(name string) interface{}
	ScriptSetProperty(name string, value interface{})
}

// Context is the runtime state during script execution.
type Context struct {
	// SenderName is the name of the "sender" object (e.g. "Cell4").
	SenderName string
	// SenderValue is the current value of the sender object (e.g. Cell4.Value).
	SenderValue interface{}
	// Objects is a map of named child objects that can be read/written by the script.
	Objects map[string]ContextObject
	// Vars holds local variables declared inside the script method.
	Vars map[string]interface{}
}

// CompiledMethod is an executable script method.
type CompiledMethod func(ctx *Context)

// ParseScript parses a C# script text and returns a map of method name to
// CompiledMethod. Unrecognised statements are silently skipped.
func ParseScript(scriptText string) (map[string]CompiledMethod, error) {
	result := make(map[string]CompiledMethod)

	// Split into lines for line-by-line processing.
	lines := strings.Split(scriptText, "\n")

	// Find all private void Method_Name(…) declarations.
	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		if !strings.HasPrefix(trimmed, "private void ") {
			continue
		}
		// Extract method name: "private void Cell4_BeforePrint(object sender, EventArgs e)"
		rest := strings.TrimPrefix(trimmed, "private void ")
		parenIdx := strings.Index(rest, "(")
		if parenIdx < 0 {
			continue
		}
		methodName := strings.TrimSpace(rest[:parenIdx])
		if methodName == "" {
			continue
		}

		// Derive senderName from method name: "Cell4_BeforePrint" → "Cell4".
		senderName := ""
		if underIdx := strings.Index(methodName, "_"); underIdx >= 0 {
			senderName = methodName[:underIdx]
		}

		// Collect body lines between the matching braces.
		bodyLines := collectBody(lines, i)

		// Compile the body lines into a closure.
		compiled := compileMethod(senderName, bodyLines)
		result[methodName] = compiled
	}

	return result, nil
}

// collectBody finds the opening '{' after line i and collects all lines until
// the matching closing '}', returning the inner lines (excluding the braces).
func collectBody(lines []string, methodLine int) []string {
	depth := 0
	inBody := false
	var body []string

	for i := methodLine; i < len(lines); i++ {
		line := lines[i]
		for _, ch := range line {
			if ch == '{' {
				depth++
				if depth == 1 {
					inBody = true
				}
			} else if ch == '}' {
				depth--
				if depth == 0 && inBody {
					return body
				}
			}
		}
		if inBody && depth > 0 {
			body = append(body, line)
		}
	}
	return body
}

// compileMethod turns a slice of C# body lines into a CompiledMethod closure.
func compileMethod(senderName string, bodyLines []string) CompiledMethod {
	// Pre-parse the lines into a list of statements.
	stmts := parseStatements(senderName, bodyLines)

	return func(ctx *Context) {
		// Make sure senderName is set in context.
		if ctx.SenderName == "" {
			ctx.SenderName = senderName
		}
		executeStatements(ctx, stmts)
	}
}

// ── Statement representation ─────────────────────────────────────────────────

type stmtKind int

const (
	stmtAssign   stmtKind = iota // ObjName.Prop = expr  or  varName = expr
	stmtVarDecl                  // type varName = expr
	stmtIfSimple                 // if (cond) stmt
	stmtIfBlock                  // if (cond) { stmts }
)

type statement struct {
	kind stmtKind

	// For stmtAssign / stmtVarDecl.
	targetObj  string // "" for local var assignment
	targetProp string // property name or var name
	valueExpr  string // raw expression text

	// For stmtIfSimple / stmtIfBlock.
	condExpr string
	thenBody []statement
}

// ── Statement parser ─────────────────────────────────────────────────────────

// parseStatements parses a list of C# lines into statement values.
func parseStatements(senderName string, lines []string) []statement {
	var stmts []statement
	i := 0
	for i < len(lines) {
		trimmed := strings.TrimSpace(lines[i])
		if trimmed == "" || strings.HasPrefix(trimmed, "//") {
			i++
			continue
		}

		// if statement.
		if strings.HasPrefix(trimmed, "if ") || strings.HasPrefix(trimmed, "if(") {
			stmt, consumed := parseIf(senderName, lines, i)
			stmts = append(stmts, stmt)
			i += consumed
			continue
		}

		// Variable declaration: "type varName = expr;"
		if st, ok := tryParseVarDecl(trimmed); ok {
			stmts = append(stmts, st)
			i++
			continue
		}

		// Assignment: "ObjName.Prop = expr;" or "varName = expr;"
		if st, ok := tryParseAssign(trimmed); ok {
			stmts = append(stmts, st)
			i++
			continue
		}

		i++
	}
	return stmts
}

// parseIf parses "if (cond) stmt" or "if (cond) { stmts }".
// Returns the parsed statement and the number of lines consumed.
func parseIf(senderName string, lines []string, start int) (statement, int) {
	trimmed := strings.TrimSpace(lines[start])

	// Extract condition text between first '(' and matching ')'.
	condText := ""
	parenOpen := strings.Index(trimmed, "(")
	if parenOpen >= 0 {
		depth := 0
		condStart := parenOpen
		for j := parenOpen; j < len(trimmed); j++ {
			if trimmed[j] == '(' {
				depth++
			} else if trimmed[j] == ')' {
				depth--
				if depth == 0 {
					condText = trimmed[condStart+1 : j]
					break
				}
			}
		}
	}

	// Find content after the closing ')' of the condition.
	afterCond := ""
	parenClose := strings.LastIndex(trimmed, ")")
	if parenClose >= 0 && parenClose < len(trimmed)-1 {
		afterCond = strings.TrimSpace(trimmed[parenClose+1:])
	}

	// If the character after ')' is '{', this is a block if.
	if strings.HasPrefix(afterCond, "{") || afterCond == "{" {
		// Collect block body.
		var blockLines []string
		depth := 0
		consumed := 0
		for i := start; i < len(lines); i++ {
			consumed++
			for _, ch := range lines[i] {
				if ch == '{' {
					depth++
				} else if ch == '}' {
					depth--
					if depth == 0 {
						goto doneBlock
					}
				}
			}
			if depth > 0 {
				blockLines = append(blockLines, lines[i])
			}
		}
	doneBlock:
		thenStmts := parseStatements(senderName, blockLines)
		return statement{
			kind:     stmtIfBlock,
			condExpr: condText,
			thenBody: thenStmts,
		}, consumed
	}

	// Simple if: inline statement on same line or next line.
	var thenLine string
	consumed := 1
	if afterCond != "" {
		thenLine = afterCond
	} else if start+1 < len(lines) {
		thenLine = strings.TrimSpace(lines[start+1])
		consumed = 2
	}

	// Parse the then-statement.
	var thenStmts []statement
	if thenLine != "" {
		if st, ok := tryParseAssign(thenLine); ok {
			thenStmts = []statement{st}
		} else if st, ok := tryParseVarDecl(thenLine); ok {
			thenStmts = []statement{st}
		}
	}

	return statement{
		kind:     stmtIfSimple,
		condExpr: condText,
		thenBody: thenStmts,
	}, consumed
}

// tryParseVarDecl tries to parse "type varName = expr;".
// Recognised type keywords: decimal, double, int, string, Color, bool.
func tryParseVarDecl(line string) (statement, bool) {
	line = strings.TrimRight(line, ";")
	types := []string{"decimal ", "double ", "int ", "float ", "string ", "Color ", "bool "}
	for _, typ := range types {
		if strings.HasPrefix(line, typ) {
			rest := strings.TrimPrefix(line, typ)
			eqIdx := strings.Index(rest, "=")
			if eqIdx < 0 {
				continue
			}
			varName := strings.TrimSpace(rest[:eqIdx])
			expr := strings.TrimSpace(rest[eqIdx+1:])
			return statement{
				kind:       stmtVarDecl,
				targetProp: varName,
				valueExpr:  expr,
			}, true
		}
	}
	return statement{}, false
}

// tryParseAssign tries to parse "ObjName.Prop = expr;" or "varName = expr;".
func tryParseAssign(line string) (statement, bool) {
	line = strings.TrimRight(line, ";")
	// Find the first '=' not preceded by '<', '>', '!', '='.
	eqIdx := -1
	for i := 0; i < len(line); i++ {
		if line[i] == '=' {
			if i > 0 {
				prev := line[i-1]
				if prev == '<' || prev == '>' || prev == '!' || prev == '=' {
					continue
				}
			}
			// Skip '==' (look-ahead).
			if i+1 < len(line) && line[i+1] == '=' {
				i++
				continue
			}
			eqIdx = i
			break
		}
	}
	if eqIdx < 0 {
		return statement{}, false
	}
	lhs := strings.TrimSpace(line[:eqIdx])
	rhs := strings.TrimSpace(line[eqIdx+1:])
	if lhs == "" || rhs == "" {
		return statement{}, false
	}

	// Check if lhs is "ObjName.Prop".
	dotIdx := strings.Index(lhs, ".")
	if dotIdx > 0 {
		obj := lhs[:dotIdx]
		prop := lhs[dotIdx+1:]
		return statement{
			kind:       stmtAssign,
			targetObj:  obj,
			targetProp: prop,
			valueExpr:  rhs,
		}, true
	}

	// Plain variable assignment.
	return statement{
		kind:       stmtAssign,
		targetObj:  "",
		targetProp: lhs,
		valueExpr:  rhs,
	}, true
}

// ── Executor ─────────────────────────────────────────────────────────────────

func executeStatements(ctx *Context, stmts []statement) {
	for _, stmt := range stmts {
		executeStatement(ctx, stmt)
	}
}

func executeStatement(ctx *Context, stmt statement) {
	switch stmt.kind {
	case stmtVarDecl:
		val := evalExpr(ctx, stmt.valueExpr)
		ctx.Vars[stmt.targetProp] = val

	case stmtAssign:
		val := evalExpr(ctx, stmt.valueExpr)
		if stmt.targetObj == "" {
			// Local variable assignment.
			ctx.Vars[stmt.targetProp] = val
		} else {
			// Object property assignment.
			if obj, ok := ctx.Objects[stmt.targetObj]; ok {
				obj.ScriptSetProperty(stmt.targetProp, val)
			}
		}

	case stmtIfSimple, stmtIfBlock:
		cond := evalExpr(ctx, stmt.condExpr)
		if asBool(cond) {
			executeStatements(ctx, stmt.thenBody)
		}
	}
}

// ── Expression evaluator ─────────────────────────────────────────────────────

// evalExpr evaluates a C# expression string within ctx.
func evalExpr(ctx *Context, expr string) interface{} {
	expr = strings.TrimSpace(expr)
	if expr == "" {
		return nil
	}

	// Ternary: expr ? thenExpr : elseExpr
	if val, ok := evalTernary(ctx, expr); ok {
		return val
	}

	// Comparison: expr >= num, expr > num, etc.
	if val, ok := evalComparison(ctx, expr); ok {
		return val
	}

	// new SolidFill(colorExpr)
	if strings.HasPrefix(expr, "new SolidFill(") && strings.HasSuffix(expr, ")") {
		inner := expr[len("new SolidFill(") : len(expr)-1]
		colorVal := evalExpr(ctx, inner)
		if c, ok := colorVal.(color.RGBA); ok {
			return style.NewSolidFill(c)
		}
		return nil
	}

	// Cast: (decimal)expr, (double)expr, (int)expr, (string)expr
	if strings.HasPrefix(expr, "(") {
		closeIdx := strings.Index(expr, ")")
		if closeIdx > 1 {
			typeName := expr[1:closeIdx]
			inner := strings.TrimSpace(expr[closeIdx+1:])
			switch typeName {
			case "decimal", "double", "float", "float32", "float64":
				inner_val := evalExpr(ctx, inner)
				return toFloat64(inner_val)
			case "int", "int32", "int64":
				inner_val := evalExpr(ctx, inner)
				return float64(int64(toFloat64(inner_val)))
			case "string":
				inner_val := evalExpr(ctx, inner)
				if inner_val == nil {
					return ""
				}
				return strings.TrimFunc(strings.TrimSpace(expr[closeIdx+1:]), func(r rune) bool { return false })
			default:
				// Unknown cast — evaluate inner.
				return evalExpr(ctx, inner)
			}
		}
	}

	// Convert.ToDouble, Convert.ToDecimal, etc.
	if strings.HasPrefix(expr, "Convert.To") {
		paren := strings.Index(expr, "(")
		if paren >= 0 && strings.HasSuffix(expr, ")") {
			inner := expr[paren+1 : len(expr)-1]
			inner_val := evalExpr(ctx, inner)
			return toFloat64(inner_val)
		}
	}

	// Color constants.
	switch expr {
	case "Color.Red":
		return color.RGBA{R: 255, G: 0, B: 0, A: 255}
	case "Color.Yellow":
		return color.RGBA{R: 255, G: 255, B: 0, A: 255}
	case "Color.Green":
		return color.RGBA{R: 0, G: 128, B: 0, A: 255}
	case "Color.GreenYellow":
		return color.RGBA{R: 173, G: 255, B: 47, A: 255}
	case "Color.Blue":
		return color.RGBA{R: 0, G: 0, B: 255, A: 255}
	case "Color.White":
		return color.RGBA{R: 255, G: 255, B: 255, A: 255}
	case "Color.Black":
		return color.RGBA{R: 0, G: 0, B: 0, A: 255}
	case "Color.Orange":
		return color.RGBA{R: 255, G: 165, B: 0, A: 255}
	case "Color.Pink":
		return color.RGBA{R: 255, G: 192, B: 203, A: 255}
	}

	// Boolean literals.
	if expr == "true" {
		return true
	}
	if expr == "false" {
		return false
	}

	// Null literal.
	if expr == "null" {
		return nil
	}

	// Numeric literal.
	if val, err := strconv.ParseFloat(expr, 64); err == nil {
		return val
	}

	// String literal.
	if strings.HasPrefix(expr, "\"") && strings.HasSuffix(expr, "\"") {
		return expr[1 : len(expr)-1]
	}

	// ObjName.Prop access — e.g. "Cell4.Value", "Shape1.Visible"
	if dotIdx := strings.Index(expr, "."); dotIdx > 0 {
		obj := expr[:dotIdx]
		prop := expr[dotIdx+1:]

		// SenderName.Value → ctx.SenderValue
		if obj == ctx.SenderName && prop == "Value" {
			return ctx.SenderValue
		}

		// Other object property reads.
		if ctxObj, ok := ctx.Objects[obj]; ok {
			return ctxObj.ScriptGetProperty(prop)
		}
	}

	// Local variable lookup.
	if val, ok := ctx.Vars[expr]; ok {
		return val
	}

	return nil
}

// evalTernary tries to evaluate "cond ? thenExpr : elseExpr".
func evalTernary(ctx *Context, expr string) (interface{}, bool) {
	// Find the top-level '?' (not inside parens/brackets).
	depth := 0
	questionIdx := -1
	for i, ch := range expr {
		switch ch {
		case '(', '[':
			depth++
		case ')', ']':
			depth--
		case '?':
			if depth == 0 {
				questionIdx = i
			}
		}
	}
	if questionIdx < 0 {
		return nil, false
	}

	// Find matching ':' after the '?'.
	colonIdx := -1
	depth = 0
	for i := questionIdx + 1; i < len(expr); i++ {
		switch expr[i] {
		case '(', '[':
			depth++
		case ')', ']':
			depth--
		case ':':
			if depth == 0 {
				colonIdx = i
			}
		}
	}
	if colonIdx < 0 {
		return nil, false
	}

	condPart := strings.TrimSpace(expr[:questionIdx])
	thenPart := strings.TrimSpace(expr[questionIdx+1 : colonIdx])
	elsePart := strings.TrimSpace(expr[colonIdx+1:])

	cond := evalExpr(ctx, condPart)
	if asBool(cond) {
		return evalExpr(ctx, thenPart), true
	}
	return evalExpr(ctx, elsePart), true
}

// evalComparison tries to evaluate binary comparisons.
// Returns (result, true) if a recognised operator is found.
func evalComparison(ctx *Context, expr string) (interface{}, bool) {
	ops := []string{">=", "<=", "!=", "==", ">", "<"}
	for _, op := range ops {
		idx := strings.Index(expr, op)
		if idx < 0 {
			continue
		}
		lhs := strings.TrimSpace(expr[:idx])
		rhs := strings.TrimSpace(expr[idx+len(op):])
		lval := evalExpr(ctx, lhs)
		rval := evalExpr(ctx, rhs)

		// Handle null comparisons.
		if op == "==" {
			if rval == nil {
				return lval == nil, true
			}
			if lval == nil {
				return false, true
			}
		}
		if op == "!=" {
			if rval == nil {
				return lval != nil, true
			}
			if lval == nil {
				return true, true
			}
		}

		// Numeric comparisons.
		lf := toFloat64(lval)
		rf := toFloat64(rval)
		switch op {
		case ">=":
			return lf >= rf, true
		case "<=":
			return lf <= rf, true
		case ">":
			return lf > rf, true
		case "<":
			return lf < rf, true
		case "==":
			return lf == rf, true
		case "!=":
			return lf != rf, true
		}
	}
	return nil, false
}

// ── Helpers ──────────────────────────────────────────────────────────────────

// asBool converts an interface{} to bool.
func asBool(v interface{}) bool {
	if v == nil {
		return false
	}
	switch t := v.(type) {
	case bool:
		return t
	case float64:
		return t != 0
	case int:
		return t != 0
	case string:
		return t != ""
	}
	return false
}

// toFloat64 converts an interface{} to float64.
func toFloat64(v interface{}) float64 {
	if v == nil {
		return 0
	}
	switch t := v.(type) {
	case float64:
		return t
	case float32:
		return float64(t)
	case int:
		return float64(t)
	case int32:
		return float64(t)
	case int64:
		return float64(t)
	case bool:
		if t {
			return 1
		}
		return 0
	case string:
		f, err := strconv.ParseFloat(t, 64)
		if err != nil {
			return 0
		}
		return f
	}
	return 0
}
