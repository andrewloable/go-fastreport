package functions

import (
	"math"
	"strings"
	"unicode"
)

// toWordsVal converts any value to float64 for the ToWords family of functions.
// Mirrors C# Convert.ToDecimal(value) used in StdFunctions overloads.
func toWordsVal(v any) float64 {
	return ToFloat(v)
}

// ── internal helpers ──────────────────────────────────────────────────────────

// convertNumberWestern formats a float64 value as locale words + custom unit names.
// Used by western-locale ConvertNumber overloads (C# junior=null, so decimal ignored).
// C# StdFunctions.cs: ToWords(value, string one, string many[, bool])
//   → new NumToWordsXxx().ConvertNumber(value, true, one, many, many, _)
func convertNumberWestern(
	value float64,
	one, many string,
	positiveFn func(int64) string,
	caseFn func(int64, string, string) string,
	zero, minus string,
) string {
	n := int64(math.Abs(value))
	negative := value < 0
	var result string
	if n == 0 {
		result = zero + " " + many
	} else {
		result = strings.TrimSpace(positiveFn(n)) + " " + caseFn(n, one, many)
	}
	if negative {
		result = minus + " " + result
	}
	r := []rune(result)
	if len(r) > 0 {
		r[0] = unicode.ToUpper(r[0])
	}
	return string(r)
}

// convertNumberGendered formats a float64 value using a gendered 3-form locale
// (Russian/Ukrainian style: one/two/many).
// C# StdFunctions.cs: ToWordsRu(value, bool male, one, two, many[, bool])
//   → new NumToWordsRu().ConvertNumber(value, male, one, two, many, _)
func convertNumberGendered(
	value float64,
	male bool,
	one, two, many string,
	positiveFn func(int64, bool) string,
	scaleWordFn func(int64, string, string, string) string,
	zero, minus string,
) string {
	n := int64(math.Abs(value))
	negative := value < 0
	var result string
	if n == 0 {
		result = zero + " " + many
	} else {
		result = strings.TrimSpace(positiveFn(n, !male)) + " " + scaleWordFn(n, one, two, many)
	}
	if negative {
		result = minus + " " + result
	}
	r := []rune(result)
	if len(r) > 0 {
		r[0] = unicode.ToUpper(r[0])
	}
	return string(r)
}

// dispatchToWordsWestern is the shared dispatcher for all western-locale ToWordsXxx
// variadic functions. It implements the C# overload resolution pattern:
//
//	(value)                   → ConvertCurrencyXxx(v, defaultCurrency, false)
//	(value, bool)             → ConvertCurrencyXxx(v, defaultCurrency, bool)
//	(value, string)           → ConvertCurrencyXxx(v, string, false)
//	(value, string, bool)     → ConvertCurrencyXxx(v, string, bool)
//	(value, string, string)   → convertNumberWestern(v, one=args[0], many=args[1])
//	(value, string, string, bool) → convertNumberWestern (decimalPart ignored, junior=nil)
func dispatchToWordsWestern(
	value any,
	args []any,
	defaultCurrency string,
	currencyFn func(float64, string, bool) (string, error),
	positiveFn func(int64) string,
	caseFn func(int64, string, string) string,
	zero, minus string,
) string {
	v := toWordsVal(value)
	switch len(args) {
	case 0:
		s, _ := currencyFn(v, defaultCurrency, false)
		return s
	case 1:
		if b, ok := args[0].(bool); ok {
			s, _ := currencyFn(v, defaultCurrency, b)
			return s
		}
		if cur, ok := args[0].(string); ok {
			s, _ := currencyFn(v, cur, false)
			return s
		}
	case 2:
		if cur, ok0 := args[0].(string); ok0 {
			if b, ok1 := args[1].(bool); ok1 {
				s, _ := currencyFn(v, cur, b)
				return s
			}
			if many, ok1 := args[1].(string); ok1 {
				// two string args → custom unit: one=args[0], many=args[1]
				return convertNumberWestern(v, cur, many, positiveFn, caseFn, zero, minus)
			}
		}
	case 3:
		if one, ok0 := args[0].(string); ok0 {
			if many, ok1 := args[1].(string); ok1 {
				// decimalPartToWord (args[2]) is ignored: C# ConvertNumber passes junior=nil
				_ = args[2]
				return convertNumberWestern(v, one, many, positiveFn, caseFn, zero, minus)
			}
		}
	}
	s, _ := currencyFn(v, defaultCurrency, false)
	return s
}

// dispatchToWordsRuUk is the shared dispatcher for Russian/Ukrainian-locale ToWordsXxx
// variadic functions. Handles 3-form gendered pluralisation.
//
//	(value)                                     → ConvertCurrencyXxx(v, defaultCurrency, false)
//	(value, bool)                               → ConvertCurrencyXxx(v, defaultCurrency, bool)
//	(value, string)                             → ConvertCurrencyXxx(v, string, false)
//	(value, string, bool)                       → ConvertCurrencyXxx(v, string, bool)
//	(value, bool, string, string, string)       → convertNumberGendered(v, male, one, two, many)
//	(value, bool, string, string, string, bool) → convertNumberGendered (decimalPart ignored)
func dispatchToWordsRuUk(
	value any,
	args []any,
	defaultCurrency string,
	currencyFn func(float64, string, bool) (string, error),
	positiveFn func(int64, bool) string,
	scaleWordFn func(int64, string, string, string) string,
	zero, minus string,
) string {
	v := toWordsVal(value)
	switch len(args) {
	case 0:
		s, _ := currencyFn(v, defaultCurrency, false)
		return s
	case 1:
		if b, ok := args[0].(bool); ok {
			s, _ := currencyFn(v, defaultCurrency, b)
			return s
		}
		if cur, ok := args[0].(string); ok {
			s, _ := currencyFn(v, cur, false)
			return s
		}
	case 2:
		if cur, ok0 := args[0].(string); ok0 {
			if b, ok1 := args[1].(bool); ok1 {
				s, _ := currencyFn(v, cur, b)
				return s
			}
		}
	case 4:
		// (bool male, string one, string two, string many)
		if male, ok0 := args[0].(bool); ok0 {
			if one, ok1 := args[1].(string); ok1 {
				if two, ok2 := args[2].(string); ok2 {
					if many, ok3 := args[3].(string); ok3 {
						return convertNumberGendered(v, male, one, two, many, positiveFn, scaleWordFn, zero, minus)
					}
				}
			}
		}
	case 5:
		// (bool male, string one, string two, string many, bool decimalPartToWord)
		if male, ok0 := args[0].(bool); ok0 {
			if one, ok1 := args[1].(string); ok1 {
				if two, ok2 := args[2].(string); ok2 {
					if many, ok3 := args[3].(string); ok3 {
						// decimalPartToWord (args[4]) is ignored: C# ConvertNumber passes junior=nil
						_ = args[4]
						return convertNumberGendered(v, male, one, two, many, positiveFn, scaleWordFn, zero, minus)
					}
				}
			}
		}
	}
	s, _ := currencyFn(v, defaultCurrency, false)
	return s
}

// ── Variadic ToWords dispatchers ──────────────────────────────────────────────
// Each function maps to a C# StdFunctions overload group.
// Registered in standard.go All() replacing the integer-only variants.

// ToWordsVariadic is the variadic dispatcher for "ToWords"/"NumToWords" (English/USD).
// C# StdFunctions.cs lines 711-776.
func ToWordsVariadic(value any, args ...any) string {
	return dispatchToWordsWestern(value, args, "USD",
		ConvertCurrencyEn,
		numToWordsPositive,
		enSimpleCase,
		"zero", "negative")
}

// ToWordsEnGbVariadic is the variadic dispatcher for "ToWordsEnGb" (English GB/GBP).
// C# StdFunctions.cs lines 782-845.
func ToWordsEnGbVariadic(value any, args ...any) string {
	return dispatchToWordsWestern(value, args, "GBP",
		ConvertCurrencyEnGb,
		enGbPositive,
		enSimpleCase, // British English uses same n==1 rule as US English
		"zero", "negative")
}

// ToWordsDeVariadic is the variadic dispatcher for "ToWordsDe" (German/EUR).
// C# StdFunctions.cs lines 1000-1063.
func ToWordsDeVariadic(value any, args ...any) string {
	return dispatchToWordsWestern(value, args, "EUR",
		ConvertCurrencyDe,
		func(n int64) string { return dePositive(n, false) },
		deSimpleCase,
		"null", "minus")
}

// ToWordsEsVariadic is the variadic dispatcher for "ToWordsEs" (Spanish/EUR).
// C# StdFunctions.cs lines 853-917.
func ToWordsEsVariadic(value any, args ...any) string {
	return dispatchToWordsWestern(value, args, "EUR",
		ConvertCurrencyEs,
		esPositive,
		esSimpleCase,
		"cero", "minus")
}

// ToWordsFrVariadic is the variadic dispatcher for "ToWordsFr" (French/EUR).
// C# StdFunctions.cs lines 1071-1134.
func ToWordsFrVariadic(value any, args ...any) string {
	return dispatchToWordsWestern(value, args, "EUR",
		ConvertCurrencyFr,
		func(n int64) string { return frPositive(n, false, false) },
		frSimpleCase,
		"zéro", "moins")
}

// ToWordsNlVariadic is the variadic dispatcher for "ToWordsNl" (Dutch/EUR).
// C# StdFunctions.cs lines 1142-1180.
func ToWordsNlVariadic(value any, args ...any) string {
	return dispatchToWordsWestern(value, args, "EUR",
		ConvertCurrencyNl,
		nlPositive,
		nlSimpleCase,
		"nul", "min")
}

// ToWordsSpVariadic is the variadic dispatcher for "ToWordsSp" (Spanish-Sp/EUR).
// C# StdFunctions.cs lines 1357-1416.
func ToWordsSpVariadic(value any, args ...any) string {
	return dispatchToWordsWestern(value, args, "EUR",
		ConvertCurrencySp,
		spPositive,
		spSimpleCase,
		"cero", "menos")
}

// ToWordsPlVariadic is the variadic dispatcher for "ToWordsPl" (Polish/PLN).
// C# StdFunctions.cs lines 1495-1556.
// Note: Polish ConvertNumber uses plScaleWord(n, one, many, many) (two=many hardcoded).
func ToWordsPlVariadic(value any, args ...any) string {
	return dispatchToWordsWestern(value, args, "PLN",
		ConvertCurrencyPl,
		func(n int64) string { return plPositive(n, false) },
		func(n int64, one, many string) string { return plScaleWord(n, one, many, many) },
		"zero", "minus")
}

// ToWordsInVariadic is the variadic dispatcher for "ToWordsIn" (Indian/INR).
// C# StdFunctions.cs lines 1213-1274.
func ToWordsInVariadic(value any, args ...any) string {
	return dispatchToWordsWestern(value, args, "INR",
		ConvertCurrencyIn,
		inPositive,
		inSimpleCase,
		"zero", "minus")
}

// ToWordsFaVariadic is the variadic dispatcher for "ToWordsPersian" (Persian/EUR).
// C# StdFunctions.cs lines 1426-1485.
func ToWordsFaVariadic(value any, args ...any) string {
	return dispatchToWordsWestern(value, args, "EUR",
		ConvertCurrencyFa,
		faPositive,
		faSimpleCase,
		"\u0635\u0641\u0631", "منفی")
}

// ToWordsRuVariadic is the variadic dispatcher for "ToWordsRu" (Russian/RUR).
// C# StdFunctions.cs lines 925-992.
func ToWordsRuVariadic(value any, args ...any) string {
	return dispatchToWordsRuUk(value, args, "RUR",
		ConvertCurrencyRu,
		ruPositive,
		ruScaleWord,
		"ноль", "минус")
}

// ToWordsUkrVariadic is the variadic dispatcher for "ToWordsUkr" (Ukrainian/UAH).
// C# StdFunctions.cs lines 1282-1347.
func ToWordsUkrVariadic(value any, args ...any) string {
	return dispatchToWordsRuUk(value, args, "UAH",
		ConvertCurrencyUk,
		ukPositive,
		ukScaleWord,
		"нуль", "мінус")
}
