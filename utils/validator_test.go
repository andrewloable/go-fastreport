package utils

import (
	"strings"
	"testing"
)

// ── mockReport ────────────────────────────────────────────────────────────────

type mockReport struct {
	pageCount       int
	bandNames       []string
	dataSourceNames []string
	textExpressions []string
	parameterNames  []string
}

func (m *mockReport) PageCount() int           { return m.pageCount }
func (m *mockReport) BandNames() []string      { return m.bandNames }
func (m *mockReport) DataSourceNames() []string { return m.dataSourceNames }
func (m *mockReport) TextExpressions() []string { return m.textExpressions }
func (m *mockReport) ParameterNames() []string  { return m.parameterNames }

// ── ValidationSeverity.String ─────────────────────────────────────────────────

func TestValidationSeverity_String(t *testing.T) {
	cases := []struct {
		sev  ValidationSeverity
		want string
	}{
		{ValidationError, "Error"},
		{ValidationWarning, "Warning"},
		{ValidationInfo, "Info"},
		{ValidationSeverity(99), "Info"}, // default branch
	}
	for _, tc := range cases {
		if got := tc.sev.String(); got != tc.want {
			t.Errorf("ValidationSeverity(%d).String() = %q, want %q", tc.sev, got, tc.want)
		}
	}
}

// ── ValidationIssue.Error ─────────────────────────────────────────────────────

func TestValidationIssue_Error_WithObjectName(t *testing.T) {
	v := ValidationIssue{
		Severity:   ValidationError,
		ObjectName: "DataBand1",
		Message:    "missing data source",
	}
	s := v.Error()
	if !strings.Contains(s, "DataBand1") || !strings.Contains(s, "missing data source") {
		t.Errorf("Error() = %q — expected object name and message", s)
	}
}

func TestValidationIssue_Error_WithoutObjectName(t *testing.T) {
	v := ValidationIssue{
		Severity: ValidationWarning,
		Message:  "report has no pages",
	}
	s := v.Error()
	if !strings.Contains(s, "report has no pages") {
		t.Errorf("Error() = %q — expected message", s)
	}
	if strings.Contains(s, "[]") {
		t.Errorf("Error() should not include empty object name bracket, got %q", s)
	}
}

// ── NewReportValidator ────────────────────────────────────────────────────────

func TestNewReportValidator_NotNil(t *testing.T) {
	v := NewReportValidator()
	if v == nil {
		t.Fatal("NewReportValidator returned nil")
	}
}

func TestNewReportValidator_DefaultRules(t *testing.T) {
	v := NewReportValidator()
	// ruleNoPages fires when PageCount = 0
	issues := v.Validate(&mockReport{pageCount: 0})
	if len(issues) == 0 {
		t.Error("expected at least one issue for empty report")
	}
}

// ── AddRule ───────────────────────────────────────────────────────────────────

func TestReportValidator_AddRule(t *testing.T) {
	v := NewReportValidator()
	called := false
	v.AddRule(func(r ValidatableReport) []ValidationIssue {
		called = true
		return nil
	})
	v.Validate(&mockReport{pageCount: 1})
	if !called {
		t.Error("custom rule was not called")
	}
}

// ── ruleNoPages ───────────────────────────────────────────────────────────────

func TestRuleNoPages_FiresWhenEmpty(t *testing.T) {
	v := NewReportValidator()
	issues := v.Validate(&mockReport{pageCount: 0})
	found := false
	for _, iss := range issues {
		if strings.Contains(iss.Message, "no pages") {
			found = true
		}
	}
	if !found {
		t.Error("expected 'no pages' warning")
	}
}

func TestRuleNoPages_SilentWhenPages(t *testing.T) {
	v := NewReportValidator()
	issues := v.Validate(&mockReport{pageCount: 1, textExpressions: []string{"[x]"}})
	for _, iss := range issues {
		if strings.Contains(iss.Message, "no pages") {
			t.Errorf("unexpected 'no pages' issue: %v", iss)
		}
	}
}

// ── ruleBracketsBalanced ──────────────────────────────────────────────────────

func TestRuleBracketsBalanced_Valid(t *testing.T) {
	v := NewReportValidator()
	issues := v.Validate(&mockReport{
		pageCount:       1,
		textExpressions: []string{"[DataSource.Field]", "[Param]"},
	})
	for _, iss := range issues {
		if strings.Contains(iss.Message, "unbalanced") {
			t.Errorf("unexpected unbalanced-bracket issue: %v", iss)
		}
	}
}

func TestRuleBracketsBalanced_ExtraClose(t *testing.T) {
	v := NewReportValidator()
	issues := v.Validate(&mockReport{
		pageCount:       1,
		textExpressions: []string{"[foo]]"},
	})
	found := false
	for _, iss := range issues {
		if strings.Contains(iss.Message, "unbalanced") {
			found = true
		}
	}
	if !found {
		t.Error("expected unbalanced-bracket error for '[foo]]'")
	}
}

func TestRuleBracketsBalanced_MissingClose(t *testing.T) {
	v := NewReportValidator()
	issues := v.Validate(&mockReport{
		pageCount:       1,
		textExpressions: []string{"[foo"},
	})
	found := false
	for _, iss := range issues {
		if strings.Contains(iss.Message, "unbalanced") {
			found = true
		}
	}
	if !found {
		t.Error("expected unbalanced-bracket error for '[foo'")
	}
}

// ── checkBracketsBalanced ─────────────────────────────────────────────────────

func TestCheckBracketsBalanced_OK(t *testing.T) {
	if err := checkBracketsBalanced("[hello][world]"); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestCheckBracketsBalanced_ExtraClose(t *testing.T) {
	if err := checkBracketsBalanced("]extra"); err == nil {
		t.Error("expected error for early ']'")
	}
}

func TestCheckBracketsBalanced_MissingClose(t *testing.T) {
	if err := checkBracketsBalanced("[open"); err == nil {
		t.Error("expected error for unclosed '['")
	}
}

func TestCheckBracketsBalanced_Empty(t *testing.T) {
	if err := checkBracketsBalanced(""); err != nil {
		t.Errorf("empty string should have no error: %v", err)
	}
}

// ── ExtractBracketExpressions ─────────────────────────────────────────────────

func TestExtractBracketExpressions_None(t *testing.T) {
	result := ExtractBracketExpressions("no brackets here")
	if len(result) != 0 {
		t.Errorf("expected empty, got %v", result)
	}
}

func TestExtractBracketExpressions_Single(t *testing.T) {
	result := ExtractBracketExpressions("Total: [Sum(Field)]")
	if len(result) != 1 || result[0] != "[Sum(Field)]" {
		t.Errorf("got %v", result)
	}
}

func TestExtractBracketExpressions_Multiple(t *testing.T) {
	result := ExtractBracketExpressions("[A] text [B]")
	if len(result) != 2 {
		t.Fatalf("expected 2, got %d: %v", len(result), result)
	}
	if result[0] != "[A]" || result[1] != "[B]" {
		t.Errorf("got %v", result)
	}
}

func TestExtractBracketExpressions_Nested(t *testing.T) {
	// Outer bracket wins
	result := ExtractBracketExpressions("[outer[inner]]")
	if len(result) != 1 {
		t.Fatalf("expected 1, got %d: %v", len(result), result)
	}
}

// ── HasUnresolvedExpression ───────────────────────────────────────────────────

func TestHasUnresolvedExpression_Known(t *testing.T) {
	known := map[string]bool{"DataSource": true}
	if HasUnresolvedExpression("[DataSource.Field]", known) {
		t.Error("known source should not flag as unresolved")
	}
}

func TestHasUnresolvedExpression_Unknown(t *testing.T) {
	known := map[string]bool{"DataSource": true}
	if !HasUnresolvedExpression("[Unknown.Field]", known) {
		t.Error("unknown source should flag as unresolved")
	}
}

func TestHasUnresolvedExpression_NoExpressions(t *testing.T) {
	if HasUnresolvedExpression("plain text", map[string]bool{}) {
		t.Error("no expressions should not be unresolved")
	}
}

func TestHasUnresolvedExpression_NoDotSuffix(t *testing.T) {
	known := map[string]bool{"MyParam": true}
	if HasUnresolvedExpression("[MyParam]", known) {
		t.Error("known param without dot should not flag as unresolved")
	}
}
