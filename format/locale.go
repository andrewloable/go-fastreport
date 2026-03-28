// Package format locale.go provides system locale detection and a lookup table
// mapping locale codes to currency/number/percent/date formatting settings.
// This mirrors C#'s CultureInfo.CurrentCulture.NumberFormat behaviour.
package format

import (
	"os"
	"os/exec"
	"runtime"
	"strings"
	"sync"
)

// LocaleNumberInfo holds locale-specific numeric formatting settings.
// Mirrors the relevant fields of .NET's System.Globalization.NumberFormatInfo.
type LocaleNumberInfo struct {
	// Currency
	CurrencySymbol           string
	CurrencyDecimalSeparator string
	CurrencyGroupSeparator   string
	CurrencyDecimalDigits    int
	CurrencyPositivePattern  int
	CurrencyNegativePattern  int

	// Number
	NumberDecimalSeparator string
	NumberGroupSeparator   string
	NumberDecimalDigits    int
	NumberNegativePattern  int

	// Percent
	PercentSymbol           string
	PercentDecimalSeparator string
	PercentGroupSeparator   string
	PercentDecimalDigits    int
	PercentPositivePattern  int
	PercentNegativePattern  int

	// Date — Go layout strings for C# standard format specifiers.
	ShortDatePattern string // "d"
	LongDatePattern  string // "D"
}

// localeTable maps normalized locale codes (e.g. "en-US", "fil-PH") to their
// NumberFormatInfo equivalents. Values sourced from .NET CultureInfo docs.
var localeTable = map[string]LocaleNumberInfo{
	"en-US": {
		CurrencySymbol: "$", CurrencyDecimalSeparator: ".", CurrencyGroupSeparator: ",",
		CurrencyDecimalDigits: 2, CurrencyPositivePattern: 0, CurrencyNegativePattern: 0,
		NumberDecimalSeparator: ".", NumberGroupSeparator: ",", NumberDecimalDigits: 2, NumberNegativePattern: 1,
		PercentSymbol: "%", PercentDecimalSeparator: ".", PercentGroupSeparator: ",",
		PercentDecimalDigits: 2, PercentPositivePattern: 1, PercentNegativePattern: 1,
		ShortDatePattern: "1/2/2006", LongDatePattern: "Monday, January 2, 2006",
	},
	"fil-PH": {
		CurrencySymbol: "₱", CurrencyDecimalSeparator: ".", CurrencyGroupSeparator: ",",
		CurrencyDecimalDigits: 2, CurrencyPositivePattern: 0, CurrencyNegativePattern: 0,
		NumberDecimalSeparator: ".", NumberGroupSeparator: ",", NumberDecimalDigits: 2, NumberNegativePattern: 1,
		PercentSymbol: "%", PercentDecimalSeparator: ".", PercentGroupSeparator: ",",
		PercentDecimalDigits: 2, PercentPositivePattern: 1, PercentNegativePattern: 1,
		ShortDatePattern: "1/2/2006", LongDatePattern: "Monday, January 2, 2006",
	},
	"en-PH": {
		CurrencySymbol: "₱", CurrencyDecimalSeparator: ".", CurrencyGroupSeparator: ",",
		CurrencyDecimalDigits: 2, CurrencyPositivePattern: 0, CurrencyNegativePattern: 0,
		NumberDecimalSeparator: ".", NumberGroupSeparator: ",", NumberDecimalDigits: 2, NumberNegativePattern: 1,
		PercentSymbol: "%", PercentDecimalSeparator: ".", PercentGroupSeparator: ",",
		PercentDecimalDigits: 2, PercentPositivePattern: 1, PercentNegativePattern: 1,
		ShortDatePattern: "1/2/2006", LongDatePattern: "Monday, January 2, 2006",
	},
	"en-GB": {
		CurrencySymbol: "£", CurrencyDecimalSeparator: ".", CurrencyGroupSeparator: ",",
		CurrencyDecimalDigits: 2, CurrencyPositivePattern: 0, CurrencyNegativePattern: 1,
		NumberDecimalSeparator: ".", NumberGroupSeparator: ",", NumberDecimalDigits: 2, NumberNegativePattern: 1,
		PercentSymbol: "%", PercentDecimalSeparator: ".", PercentGroupSeparator: ",",
		PercentDecimalDigits: 2, PercentPositivePattern: 1, PercentNegativePattern: 1,
		ShortDatePattern: "02/01/2006", LongDatePattern: "02 January 2006",
	},
	"de-DE": {
		CurrencySymbol: "€", CurrencyDecimalSeparator: ",", CurrencyGroupSeparator: ".",
		CurrencyDecimalDigits: 2, CurrencyPositivePattern: 3, CurrencyNegativePattern: 8,
		NumberDecimalSeparator: ",", NumberGroupSeparator: ".", NumberDecimalDigits: 2, NumberNegativePattern: 1,
		PercentSymbol: "%", PercentDecimalSeparator: ",", PercentGroupSeparator: ".",
		PercentDecimalDigits: 2, PercentPositivePattern: 1, PercentNegativePattern: 1,
		ShortDatePattern: "02.01.2006", LongDatePattern: "Monday, 2. January 2006",
	},
	"fr-FR": {
		CurrencySymbol: "€", CurrencyDecimalSeparator: ",", CurrencyGroupSeparator: "\u00a0",
		CurrencyDecimalDigits: 2, CurrencyPositivePattern: 3, CurrencyNegativePattern: 8,
		NumberDecimalSeparator: ",", NumberGroupSeparator: "\u00a0", NumberDecimalDigits: 2, NumberNegativePattern: 1,
		PercentSymbol: "%", PercentDecimalSeparator: ",", PercentGroupSeparator: "\u00a0",
		PercentDecimalDigits: 2, PercentPositivePattern: 1, PercentNegativePattern: 1,
		ShortDatePattern: "02/01/2006", LongDatePattern: "Monday 2 January 2006",
	},
	"ja-JP": {
		CurrencySymbol: "¥", CurrencyDecimalSeparator: ".", CurrencyGroupSeparator: ",",
		CurrencyDecimalDigits: 0, CurrencyPositivePattern: 0, CurrencyNegativePattern: 1,
		NumberDecimalSeparator: ".", NumberGroupSeparator: ",", NumberDecimalDigits: 2, NumberNegativePattern: 1,
		PercentSymbol: "%", PercentDecimalSeparator: ".", PercentGroupSeparator: ",",
		PercentDecimalDigits: 2, PercentPositivePattern: 1, PercentNegativePattern: 1,
		ShortDatePattern: "2006/01/02", LongDatePattern: "2006年1月2日",
	},
	"ru-RU": {
		CurrencySymbol: "₽", CurrencyDecimalSeparator: ",", CurrencyGroupSeparator: "\u00a0",
		CurrencyDecimalDigits: 2, CurrencyPositivePattern: 3, CurrencyNegativePattern: 8,
		NumberDecimalSeparator: ",", NumberGroupSeparator: "\u00a0", NumberDecimalDigits: 2, NumberNegativePattern: 1,
		PercentSymbol: "%", PercentDecimalSeparator: ",", PercentGroupSeparator: "\u00a0",
		PercentDecimalDigits: 2, PercentPositivePattern: 1, PercentNegativePattern: 1,
		ShortDatePattern: "02.01.2006", LongDatePattern: "2 January 2006 г.",
	},
	"uk-UA": {
		CurrencySymbol: "₴", CurrencyDecimalSeparator: ",", CurrencyGroupSeparator: "\u00a0",
		CurrencyDecimalDigits: 2, CurrencyPositivePattern: 3, CurrencyNegativePattern: 8,
		NumberDecimalSeparator: ",", NumberGroupSeparator: "\u00a0", NumberDecimalDigits: 2, NumberNegativePattern: 1,
		PercentSymbol: "%", PercentDecimalSeparator: ",", PercentGroupSeparator: "\u00a0",
		PercentDecimalDigits: 2, PercentPositivePattern: 1, PercentNegativePattern: 1,
		ShortDatePattern: "02.01.2006", LongDatePattern: "2 January 2006 р.",
	},
	"pt-BR": {
		CurrencySymbol: "R$", CurrencyDecimalSeparator: ",", CurrencyGroupSeparator: ".",
		CurrencyDecimalDigits: 2, CurrencyPositivePattern: 2, CurrencyNegativePattern: 9,
		NumberDecimalSeparator: ",", NumberGroupSeparator: ".", NumberDecimalDigits: 2, NumberNegativePattern: 1,
		PercentSymbol: "%", PercentDecimalSeparator: ",", PercentGroupSeparator: ".",
		PercentDecimalDigits: 2, PercentPositivePattern: 1, PercentNegativePattern: 1,
		ShortDatePattern: "02/01/2006", LongDatePattern: "Monday, 2 de January de 2006",
	},
	"zh-CN": {
		CurrencySymbol: "¥", CurrencyDecimalSeparator: ".", CurrencyGroupSeparator: ",",
		CurrencyDecimalDigits: 2, CurrencyPositivePattern: 0, CurrencyNegativePattern: 2,
		NumberDecimalSeparator: ".", NumberGroupSeparator: ",", NumberDecimalDigits: 2, NumberNegativePattern: 1,
		PercentSymbol: "%", PercentDecimalSeparator: ".", PercentGroupSeparator: ",",
		PercentDecimalDigits: 2, PercentPositivePattern: 1, PercentNegativePattern: 1,
		ShortDatePattern: "2006/1/2", LongDatePattern: "2006年1月2日",
	},
	"ko-KR": {
		CurrencySymbol: "₩", CurrencyDecimalSeparator: ".", CurrencyGroupSeparator: ",",
		CurrencyDecimalDigits: 0, CurrencyPositivePattern: 0, CurrencyNegativePattern: 1,
		NumberDecimalSeparator: ".", NumberGroupSeparator: ",", NumberDecimalDigits: 2, NumberNegativePattern: 1,
		PercentSymbol: "%", PercentDecimalSeparator: ".", PercentGroupSeparator: ",",
		PercentDecimalDigits: 2, PercentPositivePattern: 1, PercentNegativePattern: 1,
		ShortDatePattern: "2006-01-02", LongDatePattern: "2006년 1월 2일 Monday",
	},
	"es-ES": {
		CurrencySymbol: "€", CurrencyDecimalSeparator: ",", CurrencyGroupSeparator: ".",
		CurrencyDecimalDigits: 2, CurrencyPositivePattern: 3, CurrencyNegativePattern: 8,
		NumberDecimalSeparator: ",", NumberGroupSeparator: ".", NumberDecimalDigits: 2, NumberNegativePattern: 1,
		PercentSymbol: "%", PercentDecimalSeparator: ",", PercentGroupSeparator: ".",
		PercentDecimalDigits: 2, PercentPositivePattern: 1, PercentNegativePattern: 1,
		ShortDatePattern: "02/01/2006", LongDatePattern: "Monday, 2 de January de 2006",
	},
	"it-IT": {
		CurrencySymbol: "€", CurrencyDecimalSeparator: ",", CurrencyGroupSeparator: ".",
		CurrencyDecimalDigits: 2, CurrencyPositivePattern: 2, CurrencyNegativePattern: 9,
		NumberDecimalSeparator: ",", NumberGroupSeparator: ".", NumberDecimalDigits: 2, NumberNegativePattern: 1,
		PercentSymbol: "%", PercentDecimalSeparator: ",", PercentGroupSeparator: ".",
		PercentDecimalDigits: 2, PercentPositivePattern: 1, PercentNegativePattern: 1,
		ShortDatePattern: "02/01/2006", LongDatePattern: "Monday 2 January 2006",
	},
	"tr-TR": {
		CurrencySymbol: "₺", CurrencyDecimalSeparator: ",", CurrencyGroupSeparator: ".",
		CurrencyDecimalDigits: 2, CurrencyPositivePattern: 0, CurrencyNegativePattern: 1,
		NumberDecimalSeparator: ",", NumberGroupSeparator: ".", NumberDecimalDigits: 2, NumberNegativePattern: 1,
		PercentSymbol: "%", PercentDecimalSeparator: ",", PercentGroupSeparator: ".",
		PercentDecimalDigits: 2, PercentPositivePattern: 2, PercentNegativePattern: 2,
		ShortDatePattern: "2.01.2006", LongDatePattern: "2 January 2006 Monday",
	},
	"th-TH": {
		CurrencySymbol: "฿", CurrencyDecimalSeparator: ".", CurrencyGroupSeparator: ",",
		CurrencyDecimalDigits: 2, CurrencyPositivePattern: 0, CurrencyNegativePattern: 1,
		NumberDecimalSeparator: ".", NumberGroupSeparator: ",", NumberDecimalDigits: 2, NumberNegativePattern: 1,
		PercentSymbol: "%", PercentDecimalSeparator: ".", PercentGroupSeparator: ",",
		PercentDecimalDigits: 2, PercentPositivePattern: 1, PercentNegativePattern: 1,
		ShortDatePattern: "2/1/2006", LongDatePattern: "2 January 2006",
	},
	"vi-VN": {
		CurrencySymbol: "₫", CurrencyDecimalSeparator: ",", CurrencyGroupSeparator: ".",
		CurrencyDecimalDigits: 2, CurrencyPositivePattern: 3, CurrencyNegativePattern: 8,
		NumberDecimalSeparator: ",", NumberGroupSeparator: ".", NumberDecimalDigits: 2, NumberNegativePattern: 1,
		PercentSymbol: "%", PercentDecimalSeparator: ",", PercentGroupSeparator: ".",
		PercentDecimalDigits: 2, PercentPositivePattern: 1, PercentNegativePattern: 1,
		ShortDatePattern: "02/01/2006", LongDatePattern: "Monday, 2 January, 2006",
	},
	"id-ID": {
		CurrencySymbol: "Rp", CurrencyDecimalSeparator: ",", CurrencyGroupSeparator: ".",
		CurrencyDecimalDigits: 2, CurrencyPositivePattern: 0, CurrencyNegativePattern: 0,
		NumberDecimalSeparator: ",", NumberGroupSeparator: ".", NumberDecimalDigits: 2, NumberNegativePattern: 1,
		PercentSymbol: "%", PercentDecimalSeparator: ",", PercentGroupSeparator: ".",
		PercentDecimalDigits: 2, PercentPositivePattern: 1, PercentNegativePattern: 1,
		ShortDatePattern: "02/01/2006", LongDatePattern: "02 January 2006",
	},
	"hi-IN": {
		CurrencySymbol: "₹", CurrencyDecimalSeparator: ".", CurrencyGroupSeparator: ",",
		CurrencyDecimalDigits: 2, CurrencyPositivePattern: 2, CurrencyNegativePattern: 12,
		NumberDecimalSeparator: ".", NumberGroupSeparator: ",", NumberDecimalDigits: 2, NumberNegativePattern: 1,
		PercentSymbol: "%", PercentDecimalSeparator: ".", PercentGroupSeparator: ",",
		PercentDecimalDigits: 2, PercentPositivePattern: 1, PercentNegativePattern: 1,
		ShortDatePattern: "02-01-2006", LongDatePattern: "02 January 2006",
	},
	"pl-PL": {
		CurrencySymbol: "zł", CurrencyDecimalSeparator: ",", CurrencyGroupSeparator: "\u00a0",
		CurrencyDecimalDigits: 2, CurrencyPositivePattern: 3, CurrencyNegativePattern: 8,
		NumberDecimalSeparator: ",", NumberGroupSeparator: "\u00a0", NumberDecimalDigits: 2, NumberNegativePattern: 1,
		PercentSymbol: "%", PercentDecimalSeparator: ",", PercentGroupSeparator: "\u00a0",
		PercentDecimalDigits: 2, PercentPositivePattern: 1, PercentNegativePattern: 1,
		ShortDatePattern: "02.01.2006", LongDatePattern: "2 January 2006",
	},
}

var (
	cachedLocale     LocaleNumberInfo
	cachedLocaleOnce sync.Once
	localeOverride   string
	localeMu         sync.Mutex
)

// currentLocale returns the cached locale info for the system's current locale.
// Mirrors C#'s CultureInfo.CurrentCulture.NumberFormat.
func currentLocale() LocaleNumberInfo {
	localeMu.Lock()
	override := localeOverride
	localeMu.Unlock()

	if override != "" {
		return GetLocaleInfo(override)
	}

	cachedLocaleOnce.Do(func() {
		code := detectLocale()
		cachedLocale = GetLocaleInfo(code)
	})
	return cachedLocale
}

// SetLocale overrides the auto-detected locale. Use "" to revert to auto-detection.
// Useful for testing or server-side rendering with a specific locale.
func SetLocale(code string) {
	localeMu.Lock()
	localeOverride = code
	localeMu.Unlock()
}

// GetLocaleInfo returns the LocaleNumberInfo for the given locale code.
// Tries exact match (e.g. "fil-PH"), then language-only (e.g. "fil"),
// then falls back to en-US.
func GetLocaleInfo(code string) LocaleNumberInfo {
	code = normalizeLocale(code)
	if info, ok := localeTable[code]; ok {
		return info
	}
	// Try language-only match.
	if idx := strings.IndexByte(code, '-'); idx > 0 {
		lang := code[:idx]
		for k, v := range localeTable {
			if strings.HasPrefix(k, lang+"-") {
				return v
			}
		}
	}
	return localeTable["en-US"]
}

// detectLocale reads the system locale from environment variables.
// Checks LC_ALL, LC_MONETARY, LANG in order. On macOS, falls back to
// the system region via "defaults read NSGlobalDomain AppleLocale".
// Final fallback is "en-US".
func detectLocale() string {
	for _, env := range []string{"LC_ALL", "LC_MONETARY", "LANG"} {
		v := os.Getenv(env)
		if v == "" {
			continue
		}
		// Skip C/POSIX invariant locale (including "C.UTF-8").
		norm := normalizeLocale(v)
		if norm == "C" || norm == "POSIX" {
			continue
		}
		return v
	}
	// macOS: env vars are often "C.UTF-8"; read the actual region setting.
	if runtime.GOOS == "darwin" {
		if out, err := exec.Command("defaults", "read", "NSGlobalDomain", "AppleLocale").Output(); err == nil {
			if s := strings.TrimSpace(string(out)); s != "" {
				return s
			}
		}
	}
	return "en-US"
}

// normalizeLocale converts OS locale strings like "de_DE.UTF-8" to "de-DE".
func normalizeLocale(s string) string {
	// Strip encoding suffix: "de_DE.UTF-8" → "de_DE"
	if idx := strings.IndexByte(s, '.'); idx > 0 {
		s = s[:idx]
	}
	// Strip modifier: "sr_RS@latin" → "sr_RS"
	if idx := strings.IndexByte(s, '@'); idx > 0 {
		s = s[:idx]
	}
	// Underscore → hyphen: "de_DE" → "de-DE"
	s = strings.ReplaceAll(s, "_", "-")
	return s
}
