package expr

import (
	"fmt"
	"strings"

	exprlib "github.com/expr-lang/expr"
	"github.com/expr-lang/expr/vm"

	"github.com/andrewloable/go-fastreport/functions"
)

// Env is the evaluation environment (variable bindings).
type Env map[string]any

// EvalResult holds the result of expression evaluation.
type EvalResult struct {
	Value any
	Err   error
}

// Evaluator compiles and evaluates FastReport expressions.
type Evaluator struct {
	env   Env
	cache map[string]*vm.Program
}

// NewEvaluator creates a new Evaluator with the given environment.
func NewEvaluator(env Env) *Evaluator {
	if env == nil {
		env = make(Env)
	}
	return &Evaluator{
		env:   env,
		cache: make(map[string]*vm.Program),
	}
}

// SetVar sets a variable in the evaluation environment.
func (e *Evaluator) SetVar(name string, value any) {
	e.env[name] = value
	// Invalidate cached programs because env shape may have changed.
	e.cache = make(map[string]*vm.Program)
}

// GetVar gets a variable from the environment.
func (e *Evaluator) GetVar(name string) (any, bool) {
	v, ok := e.env[name]
	return v, ok
}

// Eval evaluates an expression string and returns the result.
// For simple variable references (identifiers without operators), it first
// looks up env directly for performance. For complex expressions it uses
// expr-lang/expr with a merged environment that includes built-in functions.
func (e *Evaluator) Eval(expression string) (any, error) {
	expression = strings.TrimSpace(expression)
	if expression == "" {
		return nil, nil
	}

	// Fast path: simple identifier lookup (no spaces, no operators).
	if isSimpleIdent(expression) {
		if v, ok := e.env[expression]; ok {
			return v, nil
		}
	}

	// Build a merged environment: standard functions (base) + built-in
	// expr-specific wrappers (override) + user variables (highest priority).
	stdFns := functions.All()
	merged := make(map[string]any, len(e.env)+len(stdFns)+len(BuiltinFunctions()))
	for k, v := range stdFns {
		merged[k] = v
	}
	// BuiltinFunctions overrides entries from functions.All() where signatures
	// differ (e.g. DateDiff argument order, error-returning Int/Float).
	for k, v := range BuiltinFunctions() {
		merged[k] = v
	}
	for k, v := range e.env {
		merged[k] = v
	}

	// Try the cache first.
	prog, ok := e.cache[expression]
	if !ok {
		var err error
		prog, err = exprlib.Compile(expression, exprlib.Env(merged))
		if err != nil {
			return nil, fmt.Errorf("expr compile %q: %w", expression, err)
		}
		e.cache[expression] = prog
	}

	result, err := exprlib.Run(prog, merged)
	if err != nil {
		return nil, fmt.Errorf("expr eval %q: %w", expression, err)
	}
	return result, nil
}

// EvalText evaluates all [expressions] in text, replacing them with their
// string representations and returning the full evaluated string.
func (e *Evaluator) EvalText(text string) (string, error) {
	tokens := Parse(text)
	if len(tokens) == 0 {
		return text, nil
	}

	var sb strings.Builder
	for _, tok := range tokens {
		if !tok.IsExpr {
			sb.WriteString(tok.Value)
			continue
		}
		val, err := e.Eval(tok.Value)
		if err != nil {
			return "", err
		}
		sb.WriteString(fmt.Sprintf("%v", val))
	}
	return sb.String(), nil
}

// isSimpleIdent returns true when s looks like a plain Go/FastReport identifier
// (letters, digits, underscores, dots — no operators or spaces).
func isSimpleIdent(s string) bool {
	if s == "" {
		return false
	}
	for _, r := range s {
		if !isIdentRune(r) {
			return false
		}
	}
	return true
}

func isIdentRune(r rune) bool {
	return (r >= 'a' && r <= 'z') ||
		(r >= 'A' && r <= 'Z') ||
		(r >= '0' && r <= '9') ||
		r == '_' || r == '.'
}
