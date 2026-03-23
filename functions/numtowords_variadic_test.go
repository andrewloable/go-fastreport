package functions_test

import (
	"strings"
	"testing"

	"github.com/andrewloable/go-fastreport/functions"
)

// ── ToWordsVariadic (English / USD) ──────────────────────────────────────────

func TestToWordsVariadic_NoArgs_DefaultCurrency(t *testing.T) {
	// (value) → ConvertCurrencyEn(v, "USD", false)
	got := functions.ToWordsVariadic(1)
	if !strings.Contains(got, "dollar") {
		t.Errorf("ToWordsVariadic(1) = %q, want to contain 'dollar'", got)
	}
}

func TestToWordsVariadic_BoolArg_DecimalWords(t *testing.T) {
	// (value, bool=true) → ConvertCurrencyEn(v, "USD", true)
	got := functions.ToWordsVariadic(1.50, true)
	if !strings.Contains(strings.ToLower(got), "fifty") {
		t.Errorf("ToWordsVariadic(1.50, true) = %q, want cents as words", got)
	}
}

func TestToWordsVariadic_BoolArg_False(t *testing.T) {
	// (value, bool=false) → ConvertCurrencyEn(v, "USD", false)
	got := functions.ToWordsVariadic(1.50, false)
	if !strings.Contains(got, "50") {
		t.Errorf("ToWordsVariadic(1.50, false) = %q, want numeric cents '50'", got)
	}
}

func TestToWordsVariadic_StringArg_Currency(t *testing.T) {
	// (value, string) → ConvertCurrencyEn(v, string, false)
	got := functions.ToWordsVariadic(5, "EUR")
	if !strings.Contains(strings.ToLower(got), "euro") {
		t.Errorf("ToWordsVariadic(5, 'EUR') = %q, want 'euro'", got)
	}
}

func TestToWordsVariadic_StringBool_CurrencyDecimal(t *testing.T) {
	// (value, string, bool) → ConvertCurrencyEn(v, string, bool)
	got := functions.ToWordsVariadic(1.25, "USD", true)
	if !strings.Contains(strings.ToLower(got), "twenty") {
		t.Errorf("ToWordsVariadic(1.25, 'USD', true) = %q, want cents as words", got)
	}
}

func TestToWordsVariadic_TwoStrings_CustomUnits(t *testing.T) {
	// (value, string one, string many) → convertNumberEn(v, one, many)
	got := functions.ToWordsVariadic(3, "page", "pages")
	if !strings.Contains(strings.ToLower(got), "three") || !strings.Contains(got, "pages") {
		t.Errorf("ToWordsVariadic(3, 'page', 'pages') = %q, want 'Three pages'", got)
	}
}

func TestToWordsVariadic_TwoStrings_Singular(t *testing.T) {
	got := functions.ToWordsVariadic(1, "page", "pages")
	if !strings.Contains(strings.ToLower(got), "one") || !strings.Contains(got, "page") {
		t.Errorf("ToWordsVariadic(1, 'page', 'pages') = %q, want singular form", got)
	}
}

func TestToWordsVariadic_ThreeArgs_CustomUnitsWithBool(t *testing.T) {
	// (value, string, string, bool) — decimalPart ignored (junior=nil)
	got := functions.ToWordsVariadic(5, "item", "items", false)
	if !strings.Contains(strings.ToLower(got), "five") || !strings.Contains(got, "items") {
		t.Errorf("ToWordsVariadic(5, 'item', 'items', false) = %q", got)
	}
}

func TestToWordsVariadic_Zero_CustomUnits(t *testing.T) {
	got := functions.ToWordsVariadic(0, "page", "pages")
	if !strings.Contains(strings.ToLower(got), "zero") || !strings.Contains(got, "pages") {
		t.Errorf("ToWordsVariadic(0, 'page', 'pages') = %q, want 'Zero pages'", got)
	}
}

func TestToWordsVariadic_FloatValue(t *testing.T) {
	got := functions.ToWordsVariadic(float64(2.0))
	if !strings.Contains(strings.ToLower(got), "dollar") {
		t.Errorf("ToWordsVariadic(float64(2.0)) = %q, want 'dollars'", got)
	}
}

// ── ToWordsRuVariadic (Russian / RUR) ────────────────────────────────────────

func TestToWordsRuVariadic_NoArgs_DefaultCurrency(t *testing.T) {
	got := functions.ToWordsRuVariadic(5)
	if got == "" {
		t.Error("ToWordsRuVariadic(5) returned empty")
	}
}

func TestToWordsRuVariadic_BoolArg(t *testing.T) {
	got := functions.ToWordsRuVariadic(5.50, true)
	if got == "" {
		t.Error("ToWordsRuVariadic(5.50, true) returned empty")
	}
}

func TestToWordsRuVariadic_CurrencyArg(t *testing.T) {
	got := functions.ToWordsRuVariadic(3, "USD")
	if got == "" {
		t.Error("ToWordsRuVariadic(3, 'USD') returned empty")
	}
}

func TestToWordsRuVariadic_CurrencyBool(t *testing.T) {
	got := functions.ToWordsRuVariadic(3, "USD", true)
	if got == "" {
		t.Error("ToWordsRuVariadic(3, 'USD', true) returned empty")
	}
}

func TestToWordsRuVariadic_MaleCustomUnits(t *testing.T) {
	// (value, bool male, one, two, many) — masculine custom unit
	got := functions.ToWordsRuVariadic(5, true, "рубль", "рубля", "рублей")
	// 5 → many form
	if !strings.Contains(got, "рублей") {
		t.Errorf("ToWordsRuVariadic(5, true, ...) = %q, want 'рублей' (many form)", got)
	}
}

func TestToWordsRuVariadic_FeminineCustomUnits(t *testing.T) {
	// 1 with feminine gender → одна (not один)
	got := functions.ToWordsRuVariadic(1, false, "страница", "страницы", "страниц")
	if !strings.Contains(got, "страница") {
		t.Errorf("ToWordsRuVariadic(1, false, ...) = %q, want 'страница'", got)
	}
}

func TestToWordsRuVariadic_CustomUnitsWithDecimalBool(t *testing.T) {
	// (value, bool, one, two, many, bool decimalPart) — decimalPart ignored
	got := functions.ToWordsRuVariadic(3, true, "рубль", "рубля", "рублей", false)
	if got == "" {
		t.Error("ToWordsRuVariadic with 5 args returned empty")
	}
}

func TestToWordsRuVariadic_Zero(t *testing.T) {
	got := functions.ToWordsRuVariadic(0, true, "рубль", "рубля", "рублей")
	// 0 → many form "ноль рублей"
	if !strings.Contains(got, "рублей") {
		t.Errorf("ToWordsRuVariadic(0, ...) = %q, want 'рублей' (many form for zero)", got)
	}
}

// ── ToWordsUkrVariadic (Ukrainian / UAH) ─────────────────────────────────────

func TestToWordsUkrVariadic_NoArgs(t *testing.T) {
	got := functions.ToWordsUkrVariadic(10)
	if got == "" {
		t.Error("ToWordsUkrVariadic(10) returned empty")
	}
}

func TestToWordsUkrVariadic_CurrencyArg(t *testing.T) {
	got := functions.ToWordsUkrVariadic(10, "USD")
	if got == "" {
		t.Error("ToWordsUkrVariadic(10, 'USD') returned empty")
	}
}

// ── ToWordsDeVariadic (German / EUR) ─────────────────────────────────────────

func TestToWordsDeVariadic_NoArgs(t *testing.T) {
	got := functions.ToWordsDeVariadic(42)
	if !strings.Contains(strings.ToLower(got), "euro") {
		t.Errorf("ToWordsDeVariadic(42) = %q, want EUR", got)
	}
}

func TestToWordsDeVariadic_CustomUnits(t *testing.T) {
	got := functions.ToWordsDeVariadic(2, "Buch", "Bücher")
	if !strings.Contains(got, "Bücher") {
		t.Errorf("ToWordsDeVariadic(2, 'Buch', 'Bücher') = %q, want 'Bücher'", got)
	}
}

func TestToWordsDeVariadic_ZeroCustomUnits(t *testing.T) {
	got := functions.ToWordsDeVariadic(0, "Buch", "Bücher")
	// "null Bücher"
	if !strings.Contains(strings.ToLower(got), "null") {
		t.Errorf("ToWordsDeVariadic(0, ...) = %q, want 'null'", got)
	}
}

// ── ToWordsEnGbVariadic (British English / GBP) ───────────────────────────────

func TestToWordsEnGbVariadic_NoArgs(t *testing.T) {
	got := functions.ToWordsEnGbVariadic(1)
	if !strings.Contains(strings.ToLower(got), "pound") {
		t.Errorf("ToWordsEnGbVariadic(1) = %q, want 'pound'", got)
	}
}

func TestToWordsEnGbVariadic_StringArg(t *testing.T) {
	got := functions.ToWordsEnGbVariadic(5, "GBP")
	if !strings.Contains(strings.ToLower(got), "pound") {
		t.Errorf("ToWordsEnGbVariadic(5, 'GBP') = %q, want 'pounds'", got)
	}
}

// ── ToWordsEsVariadic (Spanish / EUR) ────────────────────────────────────────

func TestToWordsEsVariadic_NoArgs(t *testing.T) {
	got := functions.ToWordsEsVariadic(10)
	if got == "" {
		t.Error("ToWordsEsVariadic(10) returned empty")
	}
}

// ── ToWordsFrVariadic (French / EUR) ─────────────────────────────────────────

func TestToWordsFrVariadic_NoArgs(t *testing.T) {
	got := functions.ToWordsFrVariadic(100)
	if got == "" {
		t.Error("ToWordsFrVariadic(100) returned empty")
	}
}

// ── ToWordsNlVariadic (Dutch / EUR) ──────────────────────────────────────────

func TestToWordsNlVariadic_NoArgs(t *testing.T) {
	got := functions.ToWordsNlVariadic(5)
	if got == "" {
		t.Error("ToWordsNlVariadic(5) returned empty")
	}
}

// ── ToWordsSpVariadic (Spanish-Sp / EUR) ─────────────────────────────────────

func TestToWordsSpVariadic_NoArgs(t *testing.T) {
	got := functions.ToWordsSpVariadic(10)
	if got == "" {
		t.Error("ToWordsSpVariadic(10) returned empty")
	}
}

// ── ToWordsPlVariadic (Polish / PLN) ─────────────────────────────────────────

func TestToWordsPlVariadic_NoArgs(t *testing.T) {
	got := functions.ToWordsPlVariadic(100)
	if got == "" {
		t.Error("ToWordsPlVariadic(100) returned empty")
	}
}

func TestToWordsPlVariadic_CustomUnits(t *testing.T) {
	// Polish: 5 pages → "stron" (many form)
	got := functions.ToWordsPlVariadic(5, "strona", "stron")
	if got == "" {
		t.Error("ToWordsPlVariadic(5, 'strona', 'stron') returned empty")
	}
}

// ── ToWordsInVariadic (Indian / INR) ─────────────────────────────────────────

func TestToWordsInVariadic_NoArgs(t *testing.T) {
	got := functions.ToWordsInVariadic(1000)
	if got == "" {
		t.Error("ToWordsInVariadic(1000) returned empty")
	}
}

// ── ToWordsFaVariadic (Persian / EUR) ────────────────────────────────────────

func TestToWordsFaVariadic_NoArgs(t *testing.T) {
	got := functions.ToWordsFaVariadic(5)
	if got == "" {
		t.Error("ToWordsFaVariadic(5) returned empty")
	}
}

// ── fallback / invalid args ───────────────────────────────────────────────────

func TestToWordsVariadic_UnknownCurrency_FallbackEmpty(t *testing.T) {
	// Unknown currency returns "" (ConvertCurrencyEn fails → fallback to empty string via error suppression)
	got := functions.ToWordsVariadic(5, "XYZ")
	// Should not panic; result can be empty string for unknown currency
	_ = got
}

func TestToWordsVariadic_InvalidArgType_FallsBack(t *testing.T) {
	// Passing an unexpected type as args[0] → falls through to default
	got := functions.ToWordsVariadic(5, 42) // int arg, not string/bool
	_ = got                                  // just verify no panic
}
