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
	// ClassState is a reference to the Script's shared class-level variable map.
	// Reads and writes to class fields (e.g. "total", "totals") go here so that
	// their values persist across method invocations (first pass → second pass).
	ClassState map[string]interface{}
	// IsFirstPass is true during the first pass of a double-pass report.
	// Mirrors C# Engine.FirstPass property.
	IsFirstPass bool
	// IsFinalPass is true during the second pass of a double-pass report.
	// Mirrors C# Engine.FinalPass property.
	IsFinalPass bool
	// GetVariableValue resolves a named report variable (system var or parameter).
	// Mirrors C# Report.GetVariableValue(name).
	GetVariableValue func(string) interface{}
	// GetTotalValue resolves the current accumulated value of a named total.
	// Mirrors C# Report.GetTotalValue(name).
	GetTotalValue func(string) interface{}
	// GetColumnValue retrieves the current value of a data column by dotted path
	// (e.g. "Products.ProductName"). Mirrors C# Report.GetColumnValue(name).
	GetColumnValue func(string) interface{}
	// ReturnValue holds the value returned by a script method that executes a
	// "return <expr>;" statement. Callers read this after invoking the method.
	ReturnValue interface{}
}

// CompiledMethod is an executable script method.
type CompiledMethod func(ctx *Context)

// Script holds the compiled methods and shared class-level state for a C# script.
// ClassState persists across all method invocations during a report run, allowing
// first-pass handlers to accumulate data that second-pass handlers can read.
type Script struct {
	// Methods maps event method names (e.g. "Cell4_BeforePrint") to closures.
	Methods map[string]CompiledMethod
	// ClassState is the shared mutable state for class-level fields (e.g. int total,
	// List<int> totals). Both methods in the same script share the same map.
	ClassState map[string]interface{}
}

// ParseScript parses a C# script text and returns a Script containing compiled
// methods and shared class-level state. Unrecognised statements are silently skipped.
func ParseScript(scriptText string) (*Script, error) {
	// Build the shared class state from class-level field declarations.
	classState := parseClassFields(scriptText)

	// Merge string array fields into class state.
	for k, v := range parseClassStringArrays(scriptText) {
		classState[k] = v
	}

	result := make(map[string]CompiledMethod)

	// Split into lines for line-by-line processing.
	lines := strings.Split(scriptText, "\n")

	// Find all private void/object/int/string/… Method_Name(…) declarations.
	// Non-void methods support "return <expr>;" and store their return value
	// in Context.ReturnValue after the method is invoked.
	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		var methodPrefix string
		for _, pfx := range []string{
			"private void ",
			"private object ",
			"private int ",
			"private double ",
			"private decimal ",
			"private string ",
			"private bool ",
		} {
			if strings.HasPrefix(trimmed, pfx) {
				methodPrefix = pfx
				break
			}
		}
		if methodPrefix == "" {
			continue
		}
		// Re-check that it isn't a field declaration (no '(' means it's a field).
		{
			rest := strings.TrimPrefix(trimmed, methodPrefix)
			if !strings.Contains(rest, "(") {
				continue
			}
		}
		// Extract method name: "private void Cell4_BeforePrint(object sender, EventArgs e)"
		rest := strings.TrimPrefix(trimmed, methodPrefix)
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

		// Compile the body lines into a closure that references the shared classState.
		compiled := compileMethod(senderName, bodyLines, classState)
		result[methodName] = compiled
	}

	return &Script{Methods: result, ClassState: classState}, nil
}

// parseClassFields parses class-level primitive field declarations from a C# script.
// Handles:
//
//	private int fieldName;                          → classState["fieldName"] = 0.0
//	private List<T> fieldName = new List<T>();      → classState["fieldName"] = []interface{}{}
//
// These are initialised to their zero values so evalExpr can read/write them.
func parseClassFields(scriptText string) map[string]interface{} {
	result := make(map[string]interface{})
	lines := strings.Split(scriptText, "\n")
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		// private int fieldName;
		if strings.HasPrefix(trimmed, "private int ") {
			rest := strings.TrimPrefix(trimmed, "private int ")
			// Strip any initialiser or semicolon.
			if semi := strings.Index(rest, ";"); semi >= 0 {
				rest = rest[:semi]
			}
			if eq := strings.Index(rest, "="); eq >= 0 {
				rest = rest[:eq]
			}
			name := strings.TrimSpace(rest)
			if name != "" && !strings.Contains(name, " ") {
				result[name] = float64(0)
			}
		}
		// private Hashtable fieldName = new Hashtable();
		if strings.HasPrefix(trimmed, "private Hashtable ") {
			rest := strings.TrimPrefix(trimmed, "private Hashtable ")
			if eq := strings.Index(rest, "="); eq >= 0 {
				rest = rest[:eq]
			}
			if semi := strings.Index(rest, ";"); semi >= 0 {
				rest = rest[:semi]
			}
			name := strings.TrimSpace(rest)
			if name != "" && !strings.Contains(name, " ") {
				result[name] = map[string]interface{}{}
			}
		}
		// private List<T> fieldName = new List<T>();
		if strings.HasPrefix(trimmed, "private List<") {
			// Extract field name: after "> " to "="
			gtIdx := strings.Index(trimmed, ">")
			if gtIdx < 0 {
				continue
			}
			rest := strings.TrimSpace(trimmed[gtIdx+1:])
			if eq := strings.Index(rest, "="); eq >= 0 {
				rest = rest[:eq]
			}
			if semi := strings.Index(rest, ";"); semi >= 0 {
				rest = rest[:semi]
			}
			name := strings.TrimSpace(rest)
			if name != "" && !strings.Contains(name, " ") {
				result[name] = []interface{}{}
			}
		}
	}
	return result
}

// parseClassStringArrays extracts class-level private string[] field declarations
// from a C# script text and returns them as a map of variable name → string slice.
// Handles declarations like:
//
//	private string[] monthNames = new string[] { "Jan", "Feb", ... };
func parseClassStringArrays(scriptText string) map[string][]string {
	result := make(map[string][]string)
	lines := strings.Split(scriptText, "\n")
	for i := 0; i < len(lines); i++ {
		trimmed := strings.TrimSpace(lines[i])
		if !strings.HasPrefix(trimmed, "private string[]") {
			continue
		}
		// Collect lines until the closing "};" to handle multi-line declarations.
		var buf strings.Builder
		for j := i; j < len(lines); j++ {
			buf.WriteString(lines[j])
			if strings.Contains(lines[j], "};") {
				break
			}
		}
		full := strings.TrimSpace(buf.String())
		rest := strings.TrimSpace(strings.TrimPrefix(full, "private string[]"))
		eqIdx := strings.Index(rest, "=")
		if eqIdx < 0 {
			continue
		}
		varName := strings.TrimSpace(rest[:eqIdx])
		valPart := rest[eqIdx+1:]
		braceOpen := strings.Index(valPart, "{")
		braceClose := strings.LastIndex(valPart, "}")
		if braceOpen < 0 || braceClose < 0 {
			continue
		}
		inner := valPart[braceOpen+1 : braceClose]
		var arr []string
		for _, item := range strings.Split(inner, ",") {
			item = strings.TrimSpace(strings.Join(strings.Fields(item), " "))
			if strings.HasPrefix(item, "\"") && strings.HasSuffix(item, "\"") {
				arr = append(arr, item[1:len(item)-1])
			}
		}
		if len(arr) > 0 {
			result[varName] = arr
		}
	}
	return result
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
// classState is a reference to the shared Script.ClassState map; the closure
// captures it so class-level variables persist across method calls.
func compileMethod(senderName string, bodyLines []string, classState map[string]interface{}) CompiledMethod {
	// Pre-parse the lines into a list of statements.
	stmts := parseStatements(senderName, bodyLines)

	return func(ctx *Context) {
		// Make sure senderName is set in context.
		if ctx.SenderName == "" {
			ctx.SenderName = senderName
		}
		// Wire the shared class state if the caller didn't already set it.
		// This ensures class-level variables are visible even when Context is
		// constructed by caller code that doesn't know about ClassState.
		if ctx.ClassState == nil {
			ctx.ClassState = classState
		}
		executeStatements(ctx, stmts)
	}
}

// ── Statement representation ─────────────────────────────────────────────────

type stmtKind int

const (
	stmtAssign      stmtKind = iota // ObjName.Prop = expr  or  varName = expr
	stmtVarDecl                     // type varName = expr
	stmtIfSimple                    // if (cond) stmt
	stmtIfBlock                     // if (cond) { stmts }
	stmtMethodCall                  // obj.Method(args)  — void call, no assignment
	stmtReturn                      // return <expr>
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

	// For stmtMethodCall.
	method string   // method name (e.g. "Add")
	args   []string // raw argument expressions
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

		// return statement.
		if strings.HasPrefix(trimmed, "return ") || trimmed == "return;" {
			expr := ""
			if strings.HasPrefix(trimmed, "return ") {
				expr = strings.TrimRight(strings.TrimPrefix(trimmed, "return "), ";")
				expr = strings.TrimSpace(expr)
			}
			stmts = append(stmts, statement{kind: stmtReturn, valueExpr: expr})
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

		// Method call statement: "obj.Method(args);"
		if st, ok := tryParseMethodCall(trimmed); ok {
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

	// Also check next line for Allman-style brace: "if (cond)\n{".
	if afterCond == "" && start+1 < len(lines) && strings.TrimSpace(lines[start+1]) == "{" {
		afterCond = "{"
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
		} else if st, ok := tryParseMethodCall(thenLine); ok {
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

// tryParseMethodCall tries to parse a void method-call statement: "obj.Method(args);".
// Used for patterns like "totals.Add(totalValue)".
func tryParseMethodCall(line string) (statement, bool) {
	line = strings.TrimRight(line, ";")
	// Must not contain '=' (would be caught by tryParseAssign).
	if strings.Contains(line, "=") {
		return statement{}, false
	}
	dotIdx := strings.Index(line, ".")
	if dotIdx <= 0 {
		return statement{}, false
	}
	obj := line[:dotIdx]
	rest := line[dotIdx+1:]
	parenIdx := strings.Index(rest, "(")
	if parenIdx <= 0 || !strings.HasSuffix(rest, ")") {
		return statement{}, false
	}
	method := rest[:parenIdx]
	argsStr := strings.TrimSpace(rest[parenIdx+1 : len(rest)-1])
	var args []string
	if argsStr != "" {
		args = []string{argsStr}
	}
	return statement{
		kind:      stmtMethodCall,
		targetObj: obj,
		method:    method,
		args:      args,
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
			// Write to class state if this name is a class-level field; otherwise local var.
			if ctx.ClassState != nil {
				if _, isClass := ctx.ClassState[stmt.targetProp]; isClass {
					ctx.ClassState[stmt.targetProp] = val
					break
				}
			}
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

	case stmtMethodCall:
		// "Add" on List (1 arg) or Hashtable (2 args).
		if stmt.method == "Add" && len(stmt.args) == 1 {
			parts := splitArgs(stmt.args[0])
			if len(parts) == 2 {
				// Hashtable.Add(key, value) — store in map[string]interface{}.
				keyVal := evalExpr(ctx, parts[0])
				valVal := evalExpr(ctx, parts[1])
				keyStr := ""
				if s, ok := keyVal.(string); ok {
					keyStr = s
				}
				tryAddToMap := func(m interface{}) bool {
					if hm, ok := m.(map[string]interface{}); ok {
						hm[keyStr] = valVal
						return true
					}
					return false
				}
				if ctx.ClassState != nil {
					if val, ok := ctx.ClassState[stmt.targetObj]; ok {
						tryAddToMap(val)
						break
					}
				}
				tryAddToMap(ctx.Vars[stmt.targetObj])
			} else {
				// List<T>.Add(value)
				argVal := evalExpr(ctx, stmt.args[0])
				if ctx.ClassState != nil {
					if val, ok := ctx.ClassState[stmt.targetObj]; ok {
						if slice, ok := val.([]interface{}); ok {
							ctx.ClassState[stmt.targetObj] = append(slice, argVal)
						}
						break
					}
				}
				if val, ok := ctx.Vars[stmt.targetObj]; ok {
					if slice, ok := val.([]interface{}); ok {
						ctx.Vars[stmt.targetObj] = append(slice, argVal)
					}
				}
			}
		}

	case stmtReturn:
		ctx.ReturnValue = evalExpr(ctx, stmt.valueExpr)
	}
}

// splitArgs splits a comma-separated argument list string, respecting nested
// parentheses and brackets. Returns a slice of trimmed argument strings.
func splitArgs(args string) []string {
	var result []string
	depth := 0
	start := 0
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case '(', '[':
			depth++
		case ')', ']':
			depth--
		case ',':
			if depth == 0 {
				result = append(result, strings.TrimSpace(args[start:i]))
				start = i + 1
			}
		}
	}
	if last := strings.TrimSpace(args[start:]); last != "" {
		result = append(result, last)
	}
	return result
}

// ── Expression evaluator ─────────────────────────────────────────────────────

// stripOuterParens removes a single layer of wrapping parentheses if the entire
// expression is enclosed by a matched pair: "(inner)" → "inner".
// "(a)(b)" or "(a>b) + c" are left unchanged.
func stripOuterParens(expr string) string {
	if len(expr) < 2 || expr[0] != '(' {
		return expr
	}
	depth := 0
	for i, ch := range expr {
		if ch == '(' {
			depth++
		} else if ch == ')' {
			depth--
			if depth == 0 {
				if i == len(expr)-1 {
					return expr[1:i]
				}
				return expr // closing paren before end — not a full wrap
			}
		}
	}
	return expr
}

// evalReportMethodCall handles Report.GetVariableValue(...), Report.GetTotalValue(...),
// and Report.GetColumnValue(...) when the entire expression is exactly one of these calls.
// Returns (value, true) if matched; (nil, false) otherwise.
// Uses balanced-paren matching so that "Report.GetColumnValue(...)).Substring(...)"
// is NOT matched here (instead handled by the method-call-on-expression path).
func evalReportMethodCall(ctx *Context, expr string) (interface{}, bool) {
	type reportCall struct {
		prefix  string
		handler func(string) interface{}
	}
	calls := []reportCall{
		{"Report.GetVariableValue(", func(name string) interface{} {
			if ctx.GetVariableValue != nil {
				return ctx.GetVariableValue(name)
			}
			return nil
		}},
		{"Report.GetTotalValue(", func(name string) interface{} {
			if ctx.GetTotalValue != nil {
				return ctx.GetTotalValue(name)
			}
			return nil
		}},
		{"Report.GetColumnValue(", func(name string) interface{} {
			if ctx.GetColumnValue != nil {
				return ctx.GetColumnValue(name)
			}
			return nil
		}},
	}
	for _, call := range calls {
		if !strings.HasPrefix(expr, call.prefix) {
			continue
		}
		// Find the matching ')' for the opening '(' in the prefix.
		openIdx := len(call.prefix) - 1 // index of '(' at end of prefix
		depth := 0
		closeIdx := -1
		for i := openIdx; i < len(expr); i++ {
			switch expr[i] {
			case '(':
				depth++
			case ')':
				depth--
				if depth == 0 {
					closeIdx = i
				}
			}
			if closeIdx >= 0 {
				break
			}
		}
		// Only match if the closing ')' is at the very end of the expression.
		if closeIdx != len(expr)-1 {
			continue
		}
		argExpr := strings.TrimSpace(expr[openIdx+1 : closeIdx])
		argVal := evalExpr(ctx, argExpr)
		if name, ok := argVal.(string); ok {
			return call.handler(name), true
		}
		return nil, true
	}
	return nil, false
}

// evalExpr evaluates a C# expression string within ctx.
func evalExpr(ctx *Context, expr string) interface{} {
	expr = strings.TrimSpace(expr)
	if expr == "" {
		return nil
	}

	// Strip wrapping parentheses so "((Int32)inner)" → "(Int32)inner".
	if stripped := stripOuterParens(expr); stripped != expr {
		return evalExpr(ctx, stripped)
	}

	// Logical NOT: !expr
	if strings.HasPrefix(expr, "!") {
		inner := strings.TrimSpace(expr[1:])
		val := evalExpr(ctx, inner)
		return !asBool(val)
	}

	// Ternary: expr ? thenExpr : elseExpr
	if val, ok := evalTernary(ctx, expr); ok {
		return val
	}

	// Logical: expr && expr  or  expr || expr  (must precede comparison)
	if val, ok := evalLogical(ctx, expr); ok {
		return val
	}

	// Comparison: expr >= num, expr > num, etc.
	if val, ok := evalComparison(ctx, expr); ok {
		return val
	}

	// Arithmetic: expr + expr, expr - expr, expr * expr, expr / expr
	if val, ok := evalArithmetic(ctx, expr); ok {
		return val
	}

	// Array indexing: identifier[indexExpr]
	if val, ok := evalArrayIndex(ctx, expr); ok {
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

	// Report method calls: Report.GetVariableValue("name"), GetTotalValue, GetColumnValue.
	// Use balanced-paren matching to avoid triggering when these calls appear inside
	// a larger expression (e.g., wrapped in a cast and followed by .Substring).
	if v, ok := evalReportMethodCall(ctx, expr); ok {
		return v
	}

	// Method calls on arbitrary expressions: <expr>.Method(args)
	// Handles patterns like str.Substring(0,1), map.Contains(key).
	// Must come before the cast check so that "((T)expr).Method()" is handled here
	// (the cast check would otherwise fire on the leading '(' and build a broken inner
	// expression containing the un-stripped outer-paren's closing ')').
	// Scan right-to-left for the last top-level dot before a method call.
	{
		depth := 0
		lastDot := -1
		for i := len(expr) - 1; i >= 0; i-- {
			c := expr[i]
			if c == ')' || c == ']' {
				depth++
			} else if c == '(' || c == '[' {
				depth--
			} else if c == '.' && depth == 0 {
				lastDot = i
				break
			}
		}
		if lastDot > 0 {
			objExpr := expr[:lastDot]
			methodCall := expr[lastDot+1:]
			if parenIdx := strings.Index(methodCall, "("); parenIdx > 0 && strings.HasSuffix(methodCall, ")") {
				methodName := methodCall[:parenIdx]
				argsStr := methodCall[parenIdx+1 : len(methodCall)-1]
				switch methodName {
				case "Substring":
					parts := splitArgs(argsStr)
					objVal := evalExpr(ctx, objExpr)
					if s, ok := objVal.(string); ok {
						start := int(toFloat64(evalExpr(ctx, parts[0])))
						if len(parts) == 2 {
							length := int(toFloat64(evalExpr(ctx, parts[1])))
							if start >= 0 && start+length <= len(s) {
								return s[start : start+length]
							}
						} else if len(parts) == 1 {
							if start >= 0 && start <= len(s) {
								return s[start:]
							}
						}
					}
					return ""
				case "Contains", "ContainsKey":
					keyVal := evalExpr(ctx, argsStr)
					objVal := evalExpr(ctx, objExpr)
					if m, ok := objVal.(map[string]interface{}); ok {
						if ks, ok := keyVal.(string); ok {
							_, exists := m[ks]
							return exists
						}
					}
					return false
				}
			}
		}
	}

	// Cast: (decimal)expr, (double)expr, (int)expr, (Int32)expr, etc.
	// Type name is normalised to lowercase for case-insensitive matching.
	// Note: this must come after the method-call check so that "((T)expr).Method()"
	// is handled by the method-call path (which correctly strips outer parens first).
	if strings.HasPrefix(expr, "(") {
		closeIdx := strings.Index(expr, ")")
		if closeIdx > 1 {
			typeName := strings.ToLower(expr[1:closeIdx])
			inner := strings.TrimSpace(expr[closeIdx+1:])
			switch typeName {
			case "decimal", "double", "float", "float32", "float64", "single":
				inner_val := evalExpr(ctx, inner)
				return toFloat64(inner_val)
			case "int", "int32", "int64", "uint", "uint32", "uint64", "long", "short":
				inner_val := evalExpr(ctx, inner)
				return float64(int64(toFloat64(inner_val)))
			case "string":
				inner_val := evalExpr(ctx, inner)
				if inner_val == nil {
					return ""
				}
				if s, ok := inner_val.(string); ok {
					return s
				}
				return strings.TrimSpace(inner)
			default:
				// Unknown cast — evaluate inner expression.
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

	// ObjName.Prop access — e.g. "Cell4.Value", "Shape1.Visible", "Engine.FinalPass"
	if dotIdx := strings.Index(expr, "."); dotIdx > 0 {
		obj := expr[:dotIdx]
		prop := expr[dotIdx+1:]

		// Engine properties — Engine.FinalPass, Engine.FirstPass.
		if obj == "Engine" {
			switch prop {
			case "FinalPass":
				return ctx.IsFinalPass
			case "FirstPass":
				return ctx.IsFirstPass
			}
		}

		// SenderName.Value → ctx.SenderValue
		if obj == ctx.SenderName && prop == "Value" {
			return ctx.SenderValue
		}

		// Other object property reads from Objects map.
		if ctxObj, ok := ctx.Objects[obj]; ok {
			return ctxObj.ScriptGetProperty(prop)
		}

		// .Count property on a slice variable (class state or local var).
		if prop == "Count" {
			var sliceVal interface{}
			if ctx.ClassState != nil {
				sliceVal = ctx.ClassState[obj]
			}
			if sliceVal == nil {
				sliceVal = ctx.Vars[obj]
			}
			switch v := sliceVal.(type) {
			case []interface{}:
				return float64(len(v))
			case []string:
				return float64(len(v))
			}
		}
	}

	// Class-level variable lookup (e.g. "total", "totals").
	if ctx.ClassState != nil {
		if val, ok := ctx.ClassState[expr]; ok {
			return val
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

// evalLogical evaluates top-level logical operators && and || (right-to-left scan
// ensures correct left-associativity for chains). Must be called before
// evalComparison to prevent comparison operators inside operands from confusing
// the parser.
func evalLogical(ctx *Context, expr string) (interface{}, bool) {
	depth := 0
	for i := len(expr) - 1; i >= 1; i-- {
		switch expr[i] {
		case ')', ']':
			depth++
		case '(', '[':
			depth--
		}
		if depth != 0 {
			continue
		}
		if expr[i] == '&' && expr[i-1] == '&' {
			lhs := strings.TrimSpace(expr[:i-1])
			rhs := strings.TrimSpace(expr[i+1:])
			if lhs == "" || rhs == "" {
				continue
			}
			return asBool(evalExpr(ctx, lhs)) && asBool(evalExpr(ctx, rhs)), true
		}
		if expr[i] == '|' && expr[i-1] == '|' {
			lhs := strings.TrimSpace(expr[:i-1])
			rhs := strings.TrimSpace(expr[i+1:])
			if lhs == "" || rhs == "" {
				continue
			}
			return asBool(evalExpr(ctx, lhs)) || asBool(evalExpr(ctx, rhs)), true
		}
	}
	return nil, false
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

// evalArithmetic evaluates top-level binary arithmetic (+, -, *, /) in expr.
// Scans right-to-left so the last operator is split first, giving left-to-right
// associativity for same-precedence chains.
func evalArithmetic(ctx *Context, expr string) (interface{}, bool) {
	depth := 0
	for i := len(expr) - 1; i >= 0; i-- {
		switch expr[i] {
		case ')', ']':
			depth++
		case '(', '[':
			depth--
		}
		if depth != 0 {
			continue
		}
		ch := expr[i]
		if ch != '+' && ch != '-' && ch != '*' && ch != '/' {
			continue
		}
		// Skip if part of a comparison operator (>=, <=, !=, ==, ->).
		if i+1 < len(expr) && (expr[i+1] == '=' || expr[i+1] == '>') {
			continue
		}
		if i > 0 && (expr[i-1] == '<' || expr[i-1] == '>' || expr[i-1] == '!' || expr[i-1] == '=') {
			continue
		}
		lhs := strings.TrimSpace(expr[:i])
		rhs := strings.TrimSpace(expr[i+1:])
		if lhs == "" {
			// Unary operator — skip.
			continue
		}
		lval := evalExpr(ctx, lhs)
		rval := evalExpr(ctx, rhs)
		lf := toFloat64(lval)
		rf := toFloat64(rval)
		switch ch {
		case '+':
			return lf + rf, true
		case '-':
			return lf - rf, true
		case '*':
			return lf * rf, true
		case '/':
			if rf != 0 {
				return lf / rf, true
			}
			return 0.0, true
		}
	}
	return nil, false
}

// evalArrayIndex evaluates array indexing expressions of the form name[indexExpr].
// Supports []string and []interface{} array values stored in ctx.Vars.
func evalArrayIndex(ctx *Context, expr string) (interface{}, bool) {
	if !strings.HasSuffix(expr, "]") {
		return nil, false
	}
	// Find the matching '[' from the right.
	depth := 0
	bracketIdx := -1
	for i := len(expr) - 1; i >= 0; i-- {
		switch expr[i] {
		case ']':
			depth++
		case '[':
			depth--
			if depth == 0 {
				bracketIdx = i
			}
		}
		if bracketIdx >= 0 {
			break
		}
	}
	if bracketIdx <= 0 {
		return nil, false
	}
	arrExpr := expr[:bracketIdx]
	idxExpr := expr[bracketIdx+1 : len(expr)-1]

	arrVal := evalExpr(ctx, arrExpr)
	idxVal := evalExpr(ctx, idxExpr)

	switch arr := arrVal.(type) {
	case []string:
		idx := int(toFloat64(idxVal))
		if idx >= 0 && idx < len(arr) {
			return arr[idx], true
		}
		return "", true
	case []interface{}:
		idx := int(toFloat64(idxVal))
		if idx >= 0 && idx < len(arr) {
			return arr[idx], true
		}
		return nil, true
	case map[string]interface{}:
		// Hashtable / Dictionary indexing: groupList[groupName]
		if ks, ok := idxVal.(string); ok {
			return arr[ks], true
		}
		return nil, true
	}
	return nil, false
}
