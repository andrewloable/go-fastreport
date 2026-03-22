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
	objectNames     []string
}

func (m *mockReport) PageCount() int            { return m.pageCount }
func (m *mockReport) BandNames() []string       { return m.bandNames }
func (m *mockReport) DataSourceNames() []string { return m.dataSourceNames }
func (m *mockReport) TextExpressions() []string { return m.textExpressions }
func (m *mockReport) ParameterNames() []string  { return m.parameterNames }
func (m *mockReport) ObjectNames() []string     { return m.objectNames }

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

// ── NormalizeBoundsF ──────────────────────────────────────────────────────────

func TestNormalizeBoundsF_PositiveIsUnchanged(t *testing.T) {
	l, to, w, h := NormalizeBoundsF(10, 20, 30, 40)
	if l != 10 || to != 20 || w != 30 || h != 40 {
		t.Errorf("expected (10,20,30,40) got (%v,%v,%v,%v)", l, to, w, h)
	}
}

func TestNormalizeBoundsF_NegativeWidth(t *testing.T) {
	// left=50, width=-20: right = 50+(-20) = 30; normalized: left=30, width=20
	l, to, w, h := NormalizeBoundsF(50, 5, -20, 10)
	if l != 30 || to != 5 || w != 20 || h != 10 {
		t.Errorf("got (%v,%v,%v,%v)", l, to, w, h)
	}
}

func TestNormalizeBoundsF_NegativeHeight(t *testing.T) {
	// top=50, height=-15: bottom = 50+(-15) = 35; normalized: top=35, height=15
	l, to, w, h := NormalizeBoundsF(5, 50, 10, -15)
	if l != 5 || to != 35 || w != 10 || h != 15 {
		t.Errorf("got (%v,%v,%v,%v)", l, to, w, h)
	}
}

func TestNormalizeBoundsF_BothNegative(t *testing.T) {
	l, to, w, h := NormalizeBoundsF(100, 200, -40, -60)
	if l != 60 || to != 140 || w != 40 || h != 60 {
		t.Errorf("got (%v,%v,%v,%v)", l, to, w, h)
	}
}

// ── RectsIntersectF ───────────────────────────────────────────────────────────

func TestRectsIntersectF_Overlapping(t *testing.T) {
	if !RectsIntersectF(0, 0, 10, 10, 5, 5, 10, 10) {
		t.Error("overlapping rects should intersect")
	}
}

func TestRectsIntersectF_Adjacent(t *testing.T) {
	// Edge-touching rects use open interval — should NOT intersect.
	if RectsIntersectF(0, 0, 10, 10, 10, 0, 10, 10) {
		t.Error("edge-touching rects should not intersect")
	}
}

func TestRectsIntersectF_Separated(t *testing.T) {
	if RectsIntersectF(0, 0, 5, 5, 10, 10, 5, 5) {
		t.Error("separated rects should not intersect")
	}
}

func TestRectsIntersectF_Contained(t *testing.T) {
	if !RectsIntersectF(0, 0, 100, 100, 10, 10, 20, 20) {
		t.Error("contained rect should intersect with parent")
	}
}

// ── RectContainInOtherF ───────────────────────────────────────────────────────

func TestRectContainInOtherF_Contained(t *testing.T) {
	if !RectContainInOtherF(0, 0, 100, 100, 10, 10, 50, 50) {
		t.Error("inner should be contained in outer")
	}
}

func TestRectContainInOtherF_AtEdge(t *testing.T) {
	// Inner at exact same bounds as outer — 0.01 shrink keeps it inside.
	if !RectContainInOtherF(0, 0, 100, 100, 0, 0, 100, 100) {
		t.Error("inner at exact edge should be contained after 0.01 shrink")
	}
}

func TestRectContainInOtherF_OutOfBounds(t *testing.T) {
	// Inner extends past outer on the right.
	if RectContainInOtherF(0, 0, 50, 50, 40, 0, 30, 20) {
		t.Error("inner extending past outer should not be contained")
	}
}

func TestRectContainInOtherF_NegativeSize(t *testing.T) {
	// outer: left=100, width=-100 → normalized left=0, width=100.
	if !RectContainInOtherF(100, 0, -100, 50, 10, 10, 20, 20) {
		t.Error("normalized outer should contain inner")
	}
}

// ── ruleDuplicateNames ────────────────────────────────────────────────────────

func TestRuleDuplicateNames_NoDuplicates(t *testing.T) {
	v := NewReportValidator()
	issues := v.Validate(&mockReport{
		pageCount:   1,
		objectNames: []string{"Text1", "Text2", "DataBand1"},
	})
	for _, iss := range issues {
		if strings.Contains(iss.Message, "duplicate") {
			t.Errorf("unexpected duplicate-name issue: %v", iss)
		}
	}
}

func TestRuleDuplicateNames_HasDuplicate(t *testing.T) {
	v := NewReportValidator()
	issues := v.Validate(&mockReport{
		pageCount:   1,
		objectNames: []string{"Text1", "Text2", "Text1"},
	})
	found := false
	for _, iss := range issues {
		if iss.Severity == ValidationError && strings.Contains(iss.Message, "duplicate") {
			found = true
		}
	}
	if !found {
		t.Error("expected duplicate-name error for 'Text1'")
	}
}

func TestRuleDuplicateNames_EmptyNameIgnored(t *testing.T) {
	v := NewReportValidator()
	issues := v.Validate(&mockReport{
		pageCount:   1,
		objectNames: []string{"", "", "Text1"},
	})
	for _, iss := range issues {
		if strings.Contains(iss.Message, "duplicate") {
			t.Errorf("empty names should not be flagged as duplicates: %v", iss)
		}
	}
}

func TestRuleDuplicateNames_OnlyReportedOnce(t *testing.T) {
	// Even if a name appears three times, only one error per unique duplicate name.
	v := NewReportValidator()
	issues := v.Validate(&mockReport{
		pageCount:   1,
		objectNames: []string{"X", "X", "X"},
	})
	count := 0
	for _, iss := range issues {
		if strings.Contains(iss.Message, "duplicate") && iss.ObjectName == "X" {
			count++
		}
	}
	if count != 1 {
		t.Errorf("expected exactly 1 duplicate error for 'X', got %d", count)
	}
}
