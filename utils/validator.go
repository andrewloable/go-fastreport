// Package utils provides report validation for go-fastreport.
// The Validator checks report definitions for common structural errors.
package utils

import (
	"fmt"
	"strings"
)

// ValidationSeverity indicates the severity of a validation finding.
type ValidationSeverity int

const (
	// ValidationError is a problem that will prevent correct report output.
	ValidationError ValidationSeverity = iota
	// ValidationWarning is a condition that may produce unexpected output.
	ValidationWarning
	// ValidationInfo is an informational hint (not necessarily a problem).
	ValidationInfo
)

func (s ValidationSeverity) String() string {
	switch s {
	case ValidationError:
		return "Error"
	case ValidationWarning:
		return "Warning"
	default:
		return "Info"
	}
}

// ValidationIssue describes a single validation finding.
type ValidationIssue struct {
	// Severity is the classification of the finding.
	Severity ValidationSeverity
	// ObjectName is the name of the report object involved (may be empty).
	ObjectName string
	// Message describes the problem.
	Message string
}

func (v ValidationIssue) Error() string {
	if v.ObjectName != "" {
		return fmt.Sprintf("[%s] %s: %s", v.Severity, v.ObjectName, v.Message)
	}
	return fmt.Sprintf("[%s] %s", v.Severity, v.Message)
}

// Validatable is implemented by report objects that can report their own issues.
type Validatable interface {
	Validate() []ValidationIssue
}

// ReportValidator validates a report definition by applying a set of rules.
// Rules are functions that receive a generic "report snapshot" via the
// ValidatableReport interface and return any issues found.
type ReportValidator struct {
	rules []ValidationRule
}

// ValidationRule is a single validation check.
type ValidationRule func(r ValidatableReport) []ValidationIssue

// ValidatableReport is the interface that a report must implement to be
// validated. Using an interface instead of a concrete type avoids an import
// cycle between utils and reportpkg.
type ValidatableReport interface {
	// PageCount returns the number of report pages.
	PageCount() int
	// BandNames returns the names of all bands across all pages.
	BandNames() []string
	// DataSourceNames returns the registered data source names.
	DataSourceNames() []string
	// TextExpressions returns all [bracket] expressions found in text objects.
	TextExpressions() []string
	// ParameterNames returns the registered parameter names.
	ParameterNames() []string
	// ObjectNames returns the names of all named report objects (components,
	// bands, pages). Used for duplicate-name detection, matching the C#
	// Validator.ValidateReport duplicate-name loop (Validator.cs lines 127–145).
	ObjectNames() []string
}

// NewReportValidator creates a ReportValidator with the default rule set.
func NewReportValidator() *ReportValidator {
	v := &ReportValidator{}
	v.rules = []ValidationRule{
		ruleNoPages,
		ruleBracketsBalanced,
		ruleDuplicateNames,
	}
	return v
}

// AddRule appends a custom validation rule.
func (rv *ReportValidator) AddRule(rule ValidationRule) {
	rv.rules = append(rv.rules, rule)
}

// Validate runs all rules against r and returns the collected issues.
func (rv *ReportValidator) Validate(r ValidatableReport) []ValidationIssue {
	var issues []ValidationIssue
	for _, rule := range rv.rules {
		issues = append(issues, rule(r)...)
	}
	return issues
}

// ── Built-in rules ────────────────────────────────────────────────────────────

// ruleNoPages warns when the report has no pages.
func ruleNoPages(r ValidatableReport) []ValidationIssue {
	if r.PageCount() == 0 {
		return []ValidationIssue{{
			Severity: ValidationWarning,
			Message:  "report has no pages",
		}}
	}
	return nil
}

// ruleBracketsBalanced checks that every [bracket] expression in text objects
// has balanced opening and closing brackets.
func ruleBracketsBalanced(r ValidatableReport) []ValidationIssue {
	var issues []ValidationIssue
	for _, expr := range r.TextExpressions() {
		if err := checkBracketsBalanced(expr); err != nil {
			issues = append(issues, ValidationIssue{
				Severity: ValidationError,
				Message:  fmt.Sprintf("unbalanced brackets in expression %q: %v", expr, err),
			})
		}
	}
	return issues
}

// ruleDuplicateNames reports an error for every object name that appears more
// than once in the report. This mirrors the duplicate-name loop in C#
// Validator.ValidateReport (Validator.cs lines 127–145).
func ruleDuplicateNames(r ValidatableReport) []ValidationIssue {
	names := r.ObjectNames()
	seen := make(map[string]bool, len(names))
	reported := make(map[string]bool)
	var issues []ValidationIssue
	for _, name := range names {
		if name == "" {
			continue
		}
		if seen[name] && !reported[name] {
			issues = append(issues, ValidationIssue{
				Severity:   ValidationError,
				ObjectName: name,
				Message:    "duplicate object name",
			})
			reported[name] = true
		}
		seen[name] = true
	}
	return issues
}

// checkBracketsBalanced verifies that square brackets in s are balanced.
func checkBracketsBalanced(s string) error {
	depth := 0
	for i, ch := range s {
		switch ch {
		case '[':
			depth++
		case ']':
			depth--
			if depth < 0 {
				return fmt.Errorf("unexpected ']' at position %d", i)
			}
		}
	}
	if depth != 0 {
		return fmt.Errorf("missing %d closing ']'", depth)
	}
	return nil
}

// ── Geometry helpers (mirror of C# Validator internal methods) ─────────────────
//
// These helpers work with axis-aligned rectangles represented as four float32
// values (left, top, width, height), matching the C# RectangleF convention.
// They are used by both the validator rules and by report.ReportComponentBase.Validate().
// C# reference: FastReport.Base/Utils/Validator.cs lines 38–88.

// NormalizeBoundsF ensures the width and height of a rectangle are non-negative.
// If width < 0 the left edge is moved to the right edge and width is negated.
// If height < 0 the top edge is moved to the bottom edge and height is negated.
// Equivalent to the C# internal Validator.NormalizeBounds(ref RectangleF).
func NormalizeBoundsF(left, top, width, height float32) (float32, float32, float32, float32) {
	if width < 0 {
		left = left + width // left = original Right
		width = -width
	}
	if height < 0 {
		top = top + height // top = original Bottom
		height = -height
	}
	return left, top, width, height
}

// RectsIntersectF reports whether two (normalized) rectangles overlap.
// The 0.01-unit inset applied by GetIntersectingObjects in C# is not applied
// here; callers that need it should shrink r1 by -0.01 before calling.
// C# reference: Validator.cs line 70 — bounds.IntersectsWith(bounds1).
func RectsIntersectF(l1, t1, w1, h1, l2, t2, w2, h2 float32) bool {
	return l1 < l2+w2 && l1+w1 > l2 && t1 < t2+h2 && t1+h1 > t2
}

// RectContainInOtherF returns true when the inner rectangle is fully contained
// within the outer rectangle. Both rectangles are normalized first, and the
// inner rectangle is shrunk by 0.01 units on all sides to compensate for
// designer grid-fit inaccuracy, matching the C# implementation exactly.
// C# reference: Validator.cs lines 79–88.
func RectContainInOtherF(
	outerL, outerT, outerW, outerH float32,
	innerL, innerT, innerW, innerH float32,
) bool {
	outerL, outerT, outerW, outerH = NormalizeBoundsF(outerL, outerT, outerW, outerH)
	innerL, innerT, innerW, innerH = NormalizeBoundsF(innerL, innerT, innerW, innerH)
	// Shrink inner by 0.01 to compensate for grid-fit inaccuracy.
	innerL += 0.01
	innerT += 0.01
	innerW -= 0.02
	innerH -= 0.02
	return innerL >= outerL && innerT >= outerT &&
		innerL+innerW <= outerL+outerW && innerT+innerH <= outerT+outerH
}

// ── Convenience helpers ───────────────────────────────────────────────────────

// ExtractBracketExpressions returns all [bracket] sub-strings found in text.
// Nested brackets are returned as the outermost span.
func ExtractBracketExpressions(text string) []string {
	var results []string
	depth := 0
	start := -1
	for i, ch := range text {
		switch ch {
		case '[':
			if depth == 0 {
				start = i
			}
			depth++
		case ']':
			if depth > 0 {
				depth--
				if depth == 0 && start >= 0 {
					results = append(results, text[start:i+1])
					start = -1
				}
			}
		}
	}
	return results
}

// HasUnresolvedExpression returns true if text contains at least one [bracket]
// expression whose content is not in the knownNames set.
func HasUnresolvedExpression(text string, knownNames map[string]bool) bool {
	for _, expr := range ExtractBracketExpressions(text) {
		inner := strings.TrimSpace(expr[1 : len(expr)-1])
		// Simple check: if the first token (before '.') is not a known name, flag it.
		token := inner
		if dot := strings.IndexByte(inner, '.'); dot >= 0 {
			token = inner[:dot]
		}
		if !knownNames[token] {
			return true
		}
	}
	return false
}

