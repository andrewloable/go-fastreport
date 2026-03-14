package data

import (
	"fmt"

	"github.com/andrewloable/go-fastreport/expr"
)

// Evaluate returns the effective value of the parameter identified by
// complexName. For simple parameters the raw Value is returned. When the
// parameter's Expression field is non-empty, it is evaluated using the
// expr evaluator with all top-level parameters and system variables
// available in scope.
//
// complexName may be dot-separated (e.g. "Filters.MinDate") to address nested
// parameters — the same resolution rules as FindParameter apply.
//
// Returns an error when the parameter does not exist or the expression fails
// to compile or evaluate.
func (d *Dictionary) Evaluate(complexName string) (any, error) {
	p := d.FindParameter(complexName)
	if p == nil {
		return nil, fmt.Errorf("dictionary: parameter %q not found", complexName)
	}

	if p.Expression == "" {
		return p.Value, nil
	}

	// Build evaluation environment from all top-level parameters and system
	// variables so that inter-parameter references work.
	env := make(expr.Env, len(d.parameters)+len(d.systemVariables))
	for _, param := range d.parameters {
		env[param.Name] = param.Value
	}
	for _, sv := range d.systemVariables {
		env[sv.Name] = sv.Value
	}

	ev := expr.NewEvaluator(env)
	return ev.Eval(p.Expression)
}

// EvaluateAll evaluates all top-level parameters that have an Expression set
// and stores the result back into their Value field. This is typically called
// once before the engine runs so that expression-based parameters are resolved.
func (d *Dictionary) EvaluateAll() error {
	for _, p := range d.parameters {
		if p.Expression == "" {
			continue
		}
		val, err := d.Evaluate(p.Name)
		if err != nil {
			return fmt.Errorf("dictionary: evaluating parameter %q: %w", p.Name, err)
		}
		p.Value = val
	}
	return nil
}
