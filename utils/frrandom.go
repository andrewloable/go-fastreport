package utils

// FRRandom is the pseudo-random generator.
// It is the Go port of FastReport.Utils.FRRandom (FRRandom.cs).
//
// This is intended for test-data generation only; it uses math/rand (not
// crypto/rand) to match the C# implementation which uses System.Random.

import (
	"fmt"
	"math/rand"
	"strconv"
	"strings"
	"sync"
	"time"
	"unicode"
)

// lowerLetters and upperLetters mirror the static arrays in FRRandom.cs.
var lowerLetters = []rune("abcdefghijklmnopqrstuvwxyz")
var upperLetters = []rune("ABCDEFGHIJKLMNOPQRSTUVWXYZ")

// FRRandom is a pseudo-random generator used for test-data generation.
// It is safe for concurrent use.
type FRRandom struct {
	mu  sync.Mutex
	rng *rand.Rand
}

// NewFRRandom creates a new FRRandom instance seeded with the current time.
// Mirrors the C# constructor: random = new Random()
func NewFRRandom() *FRRandom {
	//nolint:gosec // math/rand is intentional here; this is test-data generation
	return &FRRandom{
		rng: rand.New(rand.NewSource(time.Now().UnixNano())), //nolint:gosec
	}
}

// NewFRRandomSeed creates a new FRRandom instance with an explicit seed.
// Useful in tests to produce deterministic output.
func NewFRRandomSeed(seed int64) *FRRandom {
	return &FRRandom{
		rng: rand.New(rand.NewSource(seed)), //nolint:gosec
	}
}

// intn returns a random int in [0, n). Caller must hold mu.
func (r *FRRandom) intn(n int) int {
	return r.rng.Intn(n)
}

// NextLetter returns a random letter in the same case as source.
// Non-letter runes are returned unchanged.
// Mirrors FRRandom.NextLetter(char source).
func (r *FRRandom) NextLetter(source rune) rune {
	r.mu.Lock()
	defer r.mu.Unlock()
	if unicode.IsLower(source) {
		return lowerLetters[r.intn(len(lowerLetters))]
	} else if unicode.IsUpper(source) {
		return upperLetters[r.intn(len(upperLetters))]
	}
	return source
}

// NextDigit returns a random int in [0, 9].
// Mirrors FRRandom.NextDigit().
func (r *FRRandom) NextDigit() int {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.intn(10)
}

// NextDigitMax returns a random int in [0, max].
// Mirrors FRRandom.NextDigit(int max).
func (r *FRRandom) NextDigitMax(max int) int {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.intn(max + 1)
}

// NextDigitRange returns a random int in [min, max].
// Mirrors FRRandom.NextDigit(int min, int max).
func (r *FRRandom) NextDigitRange(min, max int) int {
	r.mu.Lock()
	defer r.mu.Unlock()
	if min > max {
		min, max = max, min
	}
	return min + r.intn(max-min+1)
}

// NextDigits returns a string of n random digits 0–9.
// Returns "" when number <= 0.
// Mirrors FRRandom.NextDigits(int number).
func (r *FRRandom) NextDigits(number int) string {
	if number <= 0 {
		return ""
	}
	var sb strings.Builder
	r.mu.Lock()
	defer r.mu.Unlock()
	for i := 0; i < number; i++ {
		sb.WriteByte(byte('0' + r.intn(10)))
	}
	return sb.String()
}

// NextByte returns a random byte in [0, 254].
// Mirrors FRRandom.NextByte() — C# uses random.Next(byte.MaxValue) which is
// exclusive upper bound 255, so the range is [0, 254].
func (r *FRRandom) NextByte() byte {
	r.mu.Lock()
	defer r.mu.Unlock()
	return byte(r.intn(255)) // C#: random.Next(byte.MaxValue) → [0, 254]
}

// NextBytes returns a slice of n random bytes.
// Mirrors FRRandom.NextBytes(int number).
func (r *FRRandom) NextBytes(number int) []byte {
	b := make([]byte, number)
	r.mu.Lock()
	defer r.mu.Unlock()
	for i := range b {
		b[i] = byte(r.intn(256))
	}
	return b
}

// NextChar returns a random rune in [0, 65534].
// Mirrors FRRandom.NextChar() — C# uses random.Next(char.MaxValue) which is
// exclusive upper bound 65535, so the range is [0, 65534].
func (r *FRRandom) NextChar() rune {
	r.mu.Lock()
	defer r.mu.Unlock()
	return rune(r.intn(65535)) // C#: random.Next(char.MaxValue) → [0, 65534]
}

// NextDay returns a random day between start and today (inclusive of start,
// exclusive of today because C# uses random.Next(range) which is [0, range-1]).
// If start is after today, today is returned.
// Mirrors FRRandom.NextDay(DateTime start).
func (r *FRRandom) NextDay(start time.Time) time.Time {
	today := time.Now().Truncate(24 * time.Hour)
	startDay := start.Truncate(24 * time.Hour)
	if startDay.After(today) {
		return today
	}
	rangeDays := int(today.Sub(startDay).Hours() / 24)
	if rangeDays == 0 {
		return startDay
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	return startDay.AddDate(0, 0, r.intn(rangeDays))
}

// NextTimeSpanBetweenHours returns a random duration between startHour and
// endHour (clamped to [0, 24]). The returned duration is relative to midnight.
// Mirrors FRRandom.NextTimeSpanBetweenHours(int start, int end).
func (r *FRRandom) NextTimeSpanBetweenHours(startHour, endHour int) time.Duration {
	if startHour < 0 {
		startHour = 0
	}
	if endHour > 24 {
		endHour = 24
	}
	if startHour > endHour {
		startHour, endHour = endHour, startHour
	}
	startDur := time.Duration(startHour) * time.Hour
	endDur := time.Duration(endHour) * time.Hour
	maxMinutes := int((endDur - startDur).Minutes())
	if maxMinutes <= 0 {
		return startDur
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	return startDur + time.Duration(r.intn(maxMinutes))*time.Minute
}

// RandomizeDecimal returns a random float64 with the same digit layout as
// source (same integer-part length, same fractional-part length, same sign).
// Mirrors FRRandom.RandomizeDecimal(decimal source).
//
// The C# implementation builds a string representation that respects the
// digit count and scientific notation suffix of the source value. We mirror
// that algorithm using float64 and strconv.
func (r *FRRandom) RandomizeDecimal(source float64) float64 {
	repr := strconv.FormatFloat(source, 'f', -1, 64)
	result, ok := r.randomizeNumericString(repr, source < 0)
	if !ok {
		return source
	}
	v, err := strconv.ParseFloat(result, 64)
	if err != nil {
		return source
	}
	return v
}

// randomizeNumericString randomizes the digit characters in a decimal/integer
// string representation, preserving the sign, the decimal point position, and
// any trailing exponent suffix (e.g. "E+10").
// Returns (randomized string, ok). Caller must NOT hold mu.
func (r *FRRandom) randomizeNumericString(repr string, negative bool) (string, bool) {
	// Split off optional exponent suffix (E... or e...)
	eStr := ""
	if idx := strings.IndexAny(repr, "Ee"); idx != -1 {
		eStr = strings.ToUpper(repr[idx:])
		repr = repr[:idx]
	}

	// Remove leading '-' for processing
	if negative {
		repr = strings.TrimPrefix(repr, "-")
	}

	// Split integer and fractional parts
	parts := strings.SplitN(repr, ".", 2)

	var sb strings.Builder
	if negative {
		sb.WriteByte('-')
	}

	// Integer part
	intPart := parts[0]
	if len(intPart) > 0 {
		sb.WriteByte(byte('0' + r.NextDigitRange(1, 9)))
		if len(intPart) > 1 {
			sb.WriteString(r.NextDigits(len(intPart) - 1))
		}
	}

	// Fractional part
	if len(parts) > 1 {
		fracPart := parts[1]
		sb.WriteByte('.')
		if len(fracPart) > 1 {
			sb.WriteString(r.NextDigits(len(fracPart) - 1))
		}
		// Last fractional digit: ensure non-zero (mirrors C#: NextDigit(1, 9))
		sb.WriteByte(byte('0' + r.NextDigitRange(1, 9)))
	}

	sb.WriteString(eStr)
	return sb.String(), true
}

// RandomizeDouble returns a random float64 with the same digit layout as source.
// Mirrors FRRandom.RandomizeDouble(double source).
func (r *FRRandom) RandomizeDouble(source float64) float64 {
	return r.RandomizeDecimal(source)
}

// RandomizeFloat32 returns a random float32 with the same digit layout as source.
// Mirrors FRRandom.RandomizeFloat(float source).
func (r *FRRandom) RandomizeFloat32(source float32) float32 {
	return float32(r.RandomizeDecimal(float64(source)))
}

// RandomizeInt16 returns a random int16 with the same number of digits as source.
// Mirrors FRRandom.RandomizeInt16(Int16 source).
func (r *FRRandom) RandomizeInt16(source int16) int16 {
	const maxLen = 5 // len("32767")
	s := r.randomizeSignedInt(int64(source), maxLen, func(first int) (int, string) {
		// Guarantee < 32000: first digit 1 or 2; if 3 → cap to 3 1 ...
		switch {
		case first < 3:
			return first, r.NextDigits(maxLen - 1)
		default:
			return 3, fmt.Sprintf("%d%s", r.NextDigitMax(1), r.NextDigits(maxLen-2))
		}
	})
	v, err := strconv.ParseInt(s, 10, 16)
	if err != nil {
		return source
	}
	return int16(v)
}

// RandomizeInt32 returns a random int32 with the same number of digits as source.
// Mirrors FRRandom.RandomizeInt32(Int32 source).
func (r *FRRandom) RandomizeInt32(source int32) int32 {
	const maxLen = 10 // len("2147483647")
	s := r.randomizeSignedInt(int64(source), maxLen, func(first int) (int, string) {
		// Guarantee < 2_200_000_000
		switch {
		case first < 2:
			return first, r.NextDigits(maxLen - 1)
		default:
			return 2, fmt.Sprintf("%d%s", r.NextDigitMax(1), r.NextDigits(maxLen-2))
		}
	})
	v, err := strconv.ParseInt(s, 10, 32)
	if err != nil {
		return source
	}
	return int32(v)
}

// RandomizeInt64 returns a random int64 with the same number of digits as source.
// Mirrors FRRandom.RandomizeInt64(Int64 source).
func (r *FRRandom) RandomizeInt64(source int64) int64 {
	const maxLen = 19 // len("9223372036854775807")
	s := r.randomizeSignedInt(source, maxLen, func(first int) (int, string) {
		// Guarantee < 9_200_000_000_000_000_000
		switch {
		case first < 9:
			return first, r.NextDigits(maxLen - 1)
		default:
			return 9, fmt.Sprintf("%d%s", r.NextDigitMax(1), r.NextDigits(maxLen-2))
		}
	})
	v, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return source
	}
	return v
}

// RandomizeSByte returns a random int8 with the same number of digits as source.
// Mirrors FRRandom.RandomizeSByte(SByte source).
func (r *FRRandom) RandomizeSByte(source int8) int8 {
	const maxLen = 3 // len("127")
	s := r.randomizeSignedInt(int64(source), maxLen, func(_ int) (int, string) {
		// Guarantee ≤ 127: first digit 0 or 1;
		// if 1 → second digit ≤ 2, third digit ≤ 7.
		first := r.NextDigitMax(1)
		if first < 1 {
			return first, r.NextDigits(maxLen - 1)
		}
		return first, fmt.Sprintf("%d%d", r.NextDigitMax(2), r.NextDigitMax(7))
	})
	v, err := strconv.ParseInt(s, 10, 8)
	if err != nil {
		return source
	}
	return int8(v)
}

// RandomizeUInt16 returns a random uint16 with the same number of digits as source.
// Mirrors FRRandom.RandomizeUInt16(UInt16 source).
func (r *FRRandom) RandomizeUInt16(source uint16) uint16 {
	const maxLen = 5 // len("65535")
	s := r.randomizeUnsignedInt(uint64(source), maxLen, func() string {
		// Guarantee < 65_000: first digit 1–6; if 6 → second digit ≤ 4.
		first := r.NextDigitRange(1, 6)
		if first < 6 {
			return fmt.Sprintf("%d%s", first, r.NextDigits(maxLen-1))
		}
		return fmt.Sprintf("%d%d%s", first, r.NextDigitMax(4), r.NextDigits(maxLen-2))
	})
	v, err := strconv.ParseUint(s, 10, 16)
	if err != nil {
		return source
	}
	return uint16(v)
}

// RandomizeUInt32 returns a random uint32 with the same number of digits as source.
// Mirrors FRRandom.RandomizeUInt32(UInt32 source).
func (r *FRRandom) RandomizeUInt32(source uint32) uint32 {
	const maxLen = 10 // len("4294967295")
	s := r.randomizeUnsignedInt(uint64(source), maxLen, func() string {
		// Guarantee < 4_200_000_000
		first := r.NextDigitRange(1, 4)
		if first < 4 {
			return fmt.Sprintf("%d%s", first, r.NextDigits(maxLen-1))
		}
		return fmt.Sprintf("%d%d%s", first, r.NextDigitMax(1), r.NextDigits(maxLen-2))
	})
	v, err := strconv.ParseUint(s, 10, 32)
	if err != nil {
		return source
	}
	return uint32(v)
}

// RandomizeUInt64 returns a random uint64 with the same number of digits as source.
// Mirrors FRRandom.RandomizeUInt64(UInt64 source).
func (r *FRRandom) RandomizeUInt64(source uint64) uint64 {
	const maxLen = 20 // len("18446744073709551615")
	s := r.randomizeUnsignedInt(source, maxLen, func() string {
		// Guarantee < 18_400_000_000_000_000_000:
		// always starts with "1", then second digit 0–8;
		// if < 8 → fill rest, else third digit 0–3 then fill.
		second := r.NextDigitMax(8)
		if second < 8 {
			return fmt.Sprintf("1%d%s", second, r.NextDigits(maxLen-2))
		}
		return fmt.Sprintf("1%d%d%s", second, r.NextDigitMax(3), r.NextDigits(maxLen-3))
	})
	v, err := strconv.ParseUint(s, 10, 64)
	if err != nil {
		return source
	}
	return v
}

// RandomizeString returns a string of the same length as source where:
//   - whitespace is preserved,
//   - letters are replaced by random letters of the same case,
//   - digits are replaced by random digits 0–9,
//   - all other characters are preserved.
//
// Mirrors FRRandom.RandomizeString(string source).
func (r *FRRandom) RandomizeString(source string) string {
	var sb strings.Builder
	for _, c := range source {
		switch {
		case unicode.IsSpace(c):
			sb.WriteRune(c)
		case unicode.IsLetter(c):
			sb.WriteRune(r.NextLetter(c))
		case unicode.IsDigit(c):
			sb.WriteRune(rune('0' + r.NextDigit()))
		default:
			sb.WriteRune(c)
		}
	}
	return sb.String()
}

// GetRandomValue returns a randomized value of the same type as source.
// Supported types: string, int, int8, int16, int32, int64,
// uint8, uint16, uint32, uint64, float32, float64, time.Time, time.Duration,
// []byte.
// For unsupported types the original source value is returned unchanged.
// Mirrors FRRandom.GetRandomObject(object source, Type type).
func (r *FRRandom) GetRandomValue(source any) (result any) {
	defer func() {
		if rec := recover(); rec != nil {
			result = source
		}
	}()
	switch v := source.(type) {
	case string:
		return r.RandomizeString(v)
	case int:
		return int(r.RandomizeInt32(int32(v)))
	case int8:
		return r.RandomizeSByte(v)
	case int16:
		return r.RandomizeInt16(v)
	case int32:
		return r.RandomizeInt32(v)
	case int64:
		return r.RandomizeInt64(v)
	case uint8:
		return r.NextByte()
	case uint16:
		return r.RandomizeUInt16(v)
	case uint32:
		return r.RandomizeUInt32(v)
	case uint64:
		return r.RandomizeUInt64(v)
	case float32:
		return r.RandomizeFloat32(v)
	case float64:
		return r.RandomizeDouble(v)
	case time.Time:
		return r.NextDay(time.Date(1990, 1, 1, 0, 0, 0, 0, time.UTC))
	case time.Duration:
		return r.NextTimeSpanBetweenHours(0, 24)
	case []byte:
		return r.NextBytes(len(v))
	}
	return source
}

// ---- helper methods for integer randomization ----

// randomizeSignedInt builds a random decimal string for a signed integer
// whose string representation has the same number of significant digits as
// source. When the digit count would overflow the type (length == maxLen),
// the overflowFn callback is invoked to produce a safe string representation.
// overflowFn receives the randomly chosen first digit (1–9) and must return
// (firstDigit, remainingDigits).
func (r *FRRandom) randomizeSignedInt(source int64, maxLen int, overflowFn func(first int) (int, string)) string {
	repr := strconv.FormatInt(source, 10)
	negative := source < 0
	if negative {
		repr = repr[1:] // strip '-'
	}
	length := len(repr)

	var sb strings.Builder
	if negative {
		sb.WriteByte('-')
	}

	if length < maxLen {
		first := r.NextDigitRange(1, 9)
		sb.WriteByte(byte('0' + first))
		sb.WriteString(r.NextDigits(length - 1))
	} else {
		first := r.NextDigitRange(1, 9)
		f, rest := overflowFn(first)
		sb.WriteByte(byte('0' + f))
		sb.WriteString(rest)
	}
	return sb.String()
}

// randomizeUnsignedInt builds a random decimal string for an unsigned integer
// with the same number of digits as source. When the digit count would overflow
// the type (length == maxLen), overflowFn is invoked to produce a safe string.
func (r *FRRandom) randomizeUnsignedInt(source uint64, maxLen int, overflowFn func() string) string {
	repr := strconv.FormatUint(source, 10)
	length := len(repr)

	if length < maxLen {
		var sb strings.Builder
		first := r.NextDigitRange(1, 9)
		sb.WriteByte(byte('0' + first))
		sb.WriteString(r.NextDigits(length - 1))
		return sb.String()
	}
	return overflowFn()
}

// ---- Support types for datasource randomization ----

// FRColumnInfo holds the type name and row count for a data column.
// Mirrors FRColumnInfo in FRRandom.cs.
type FRColumnInfo struct {
	// TypeName is a Go type descriptor string (e.g. "string", "int32").
	TypeName string
	// Length is the number of rows.
	Length int
}

// FRRandomFieldValue stores the original and randomized value for a field.
// Mirrors FRRandomFieldValue in FRRandom.cs.
type FRRandomFieldValue struct {
	Origin any
	Rand   any
}

// FRRandomFieldValueCollection is an ordered list of FRRandomFieldValue entries.
// Mirrors FRRandomFieldValueCollection in FRRandom.cs.
type FRRandomFieldValueCollection struct {
	list []FRRandomFieldValue
}

// Add appends a value to the collection.
func (c *FRRandomFieldValueCollection) Add(v FRRandomFieldValue) {
	c.list = append(c.list, v)
}

// ContainsOrigin returns true if an entry with the same origin exists.
// Uses == for equality (mirrors C# reference / value equality).
func (c *FRRandomFieldValueCollection) ContainsOrigin(origin any) bool {
	for _, v := range c.list {
		if v.Origin == origin {
			return true
		}
	}
	return false
}

// ContainsRandom returns true if an entry with the same random value exists.
func (c *FRRandomFieldValueCollection) ContainsRandom(randVal any) bool {
	for _, v := range c.list {
		if v.Rand == randVal {
			return true
		}
	}
	return false
}

// GetRandom returns the random value stored for the given origin.
// If not found, origin itself is returned.
func (c *FRRandomFieldValueCollection) GetRandom(origin any) any {
	for _, v := range c.list {
		if v.Origin == origin {
			return v.Rand
		}
	}
	return origin
}
