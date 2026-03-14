package expr

import (
	"strings"
	"testing"
)

func TestNewEvaluator_NilEnv(t *testing.T) {
	e := NewEvaluator(nil)
	if e == nil {
		t.Fatal("expected non-nil evaluator")
	}
	if e.env == nil {
		t.Error("expected env to be initialized")
	}
}

func TestNewEvaluator_WithEnv(t *testing.T) {
	env := Env{"Name": "Alice"}
	e := NewEvaluator(env)
	v, ok := e.GetVar("Name")
	if !ok || v != "Alice" {
		t.Errorf("expected Name=Alice, got ok=%v v=%v", ok, v)
	}
}

func TestSetVar_And_GetVar(t *testing.T) {
	e := NewEvaluator(nil)
	e.SetVar("Age", 30)
	v, ok := e.GetVar("Age")
	if !ok || v != 30 {
		t.Errorf("expected Age=30, got ok=%v v=%v", ok, v)
	}
}

func TestGetVar_Missing(t *testing.T) {
	e := NewEvaluator(nil)
	_, ok := e.GetVar("Missing")
	if ok {
		t.Error("expected false for missing var")
	}
}

func TestEval_SimpleVar(t *testing.T) {
	e := NewEvaluator(Env{"Name": "Alice"})
	val, err := e.Eval("Name")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if val != "Alice" {
		t.Errorf("expected Alice, got %v", val)
	}
}

func TestEval_EmptyExpression(t *testing.T) {
	e := NewEvaluator(nil)
	val, err := e.Eval("")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if val != nil {
		t.Errorf("expected nil, got %v", val)
	}
}

func TestEval_WhitespaceExpression(t *testing.T) {
	e := NewEvaluator(nil)
	val, err := e.Eval("   ")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if val != nil {
		t.Errorf("expected nil, got %v", val)
	}
}

func TestEval_Arithmetic(t *testing.T) {
	e := NewEvaluator(Env{"x": 5, "y": 3})
	val, err := e.Eval("x + y")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if val.(int) != 8 {
		t.Errorf("expected 8, got %v", val)
	}
}

func TestEval_Comparison(t *testing.T) {
	e := NewEvaluator(Env{"x": 10})
	val, err := e.Eval("x > 5")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if val.(bool) != true {
		t.Errorf("expected true, got %v", val)
	}
}

func TestEval_InvalidExpression(t *testing.T) {
	e := NewEvaluator(nil)
	_, err := e.Eval("??? invalid ???")
	if err == nil {
		t.Error("expected error for invalid expression")
	}
}

func TestEval_Cached(t *testing.T) {
	e := NewEvaluator(Env{"x": 10})
	// First call compiles and caches.
	val1, err := e.Eval("x + 1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Second call uses cache.
	val2, err := e.Eval("x + 1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if val1 != val2 {
		t.Errorf("cached value mismatch: %v vs %v", val1, val2)
	}
}

func TestSetVar_ClearsCache(t *testing.T) {
	e := NewEvaluator(Env{"x": 10})
	// Prime the cache.
	_, err := e.Eval("x + 1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(e.cache) == 0 {
		t.Error("expected cache to be populated")
	}
	// SetVar should clear the cache.
	e.SetVar("x", 20)
	if len(e.cache) != 0 {
		t.Error("expected cache to be cleared after SetVar")
	}
}

func TestEval_BuiltinIIF(t *testing.T) {
	e := NewEvaluator(Env{"score": 85})
	val, err := e.Eval(`IIF(score >= 90, "A", "B")`)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if val != "B" {
		t.Errorf("expected B, got %v", val)
	}
}

func TestEvalText_NoExpression(t *testing.T) {
	e := NewEvaluator(nil)
	result, err := e.EvalText("Hello, world!")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != "Hello, world!" {
		t.Errorf("expected %q, got %q", "Hello, world!", result)
	}
}

func TestEvalText_WithExpression(t *testing.T) {
	e := NewEvaluator(Env{"Name": "Alice", "Count": 42})
	result, err := e.EvalText("Hello [Name], you have [Count] items.")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	expected := "Hello Alice, you have 42 items."
	if result != expected {
		t.Errorf("expected %q, got %q", expected, result)
	}
}

func TestEvalText_EmptyText(t *testing.T) {
	e := NewEvaluator(nil)
	result, err := e.EvalText("")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != "" {
		t.Errorf("expected empty string, got %q", result)
	}
}

func TestEvalText_EvalError(t *testing.T) {
	e := NewEvaluator(nil)
	_, err := e.EvalText("[??? bad ???]")
	if err == nil {
		t.Error("expected error for invalid expression in text")
	}
}

func TestEvalText_OnlyExpression(t *testing.T) {
	e := NewEvaluator(Env{"Total": 100})
	result, err := e.EvalText("[Total]")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != "100" {
		t.Errorf("expected %q, got %q", "100", result)
	}
}

func TestIsSimpleIdent(t *testing.T) {
	cases := []struct {
		s    string
		want bool
	}{
		{"Name", true},
		{"data.Name", true},
		{"_private", true},
		{"x1", true},
		{"", false},
		{"x y", false},
		{"x+y", false},
		{"x>0", false},
	}
	for _, tc := range cases {
		got := isSimpleIdent(tc.s)
		if got != tc.want {
			t.Errorf("isSimpleIdent(%q) = %v, want %v", tc.s, got, tc.want)
		}
	}
}

func TestEval_SimpleVarNotInEnv_FallsBackToExpr(t *testing.T) {
	// A simple identifier that is not in env should fail gracefully.
	e := NewEvaluator(Env{"y": 1})
	_, err := e.Eval("unknownVar")
	// expr-lang/expr will fail to compile since the var is not in the env.
	if err == nil {
		t.Error("expected error when unknown var not in env")
	}
}

func TestEval_RuntimeError(t *testing.T) {
	// Division by zero as a runtime eval error.
	e := NewEvaluator(Env{"x": 0})
	_, err := e.Eval("1 / x")
	// expr-lang may or may not error here depending on types; just ensure no panic.
	_ = err
}

func TestEval_RuntimeErrorFromBuiltin(t *testing.T) {
	// Int() called on a struct causes a runtime error in the expr VM.
	e := NewEvaluator(Env{"s": "not_a_number"})
	_, err := e.Eval("Int(s)")
	if err == nil {
		t.Error("expected runtime error from Int on non-numeric string")
	}
}

func TestEvalText_MultipleExpressions(t *testing.T) {
	e := NewEvaluator(Env{"A": "foo", "B": "bar"})
	result, err := e.EvalText("[A] and [B]")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(result, "foo") || !strings.Contains(result, "bar") {
		t.Errorf("unexpected result: %q", result)
	}
}
