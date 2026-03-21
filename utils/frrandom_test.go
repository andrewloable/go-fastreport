package utils_test

import (
	"fmt"
	"strings"
	"testing"
	"time"
	"unicode"

	"github.com/andrewloable/go-fastreport/utils"
)

// seed42 creates a deterministic FRRandom for reproducible tests.
func seed42() *utils.FRRandom { return utils.NewFRRandomSeed(42) }

// ---- NextLetter ----

func TestFRRandom_NextLetter_Lower(t *testing.T) {
	r := seed42()
	for i := 0; i < 100; i++ {
		got := r.NextLetter('a')
		if !unicode.IsLower(got) {
			t.Fatalf("NextLetter('a') returned non-lowercase rune %q", got)
		}
	}
}

func TestFRRandom_NextLetter_Upper(t *testing.T) {
	r := seed42()
	for i := 0; i < 100; i++ {
		got := r.NextLetter('A')
		if !unicode.IsUpper(got) {
			t.Fatalf("NextLetter('A') returned non-uppercase rune %q", got)
		}
	}
}

func TestFRRandom_NextLetter_NonLetter(t *testing.T) {
	r := seed42()
	if got := r.NextLetter('5'); got != '5' {
		t.Errorf("NextLetter('5') = %q, want '5'", got)
	}
	if got := r.NextLetter(' '); got != ' ' {
		t.Errorf("NextLetter(' ') = %q, want ' '", got)
	}
}

// ---- NextDigit ----

func TestFRRandom_NextDigit_Range(t *testing.T) {
	r := seed42()
	for i := 0; i < 500; i++ {
		d := r.NextDigit()
		if d < 0 || d > 9 {
			t.Fatalf("NextDigit() = %d, out of [0,9]", d)
		}
	}
}

func TestFRRandom_NextDigitMax_Range(t *testing.T) {
	r := seed42()
	for i := 0; i < 500; i++ {
		d := r.NextDigitMax(5)
		if d < 0 || d > 5 {
			t.Fatalf("NextDigitMax(5) = %d, out of [0,5]", d)
		}
	}
}

func TestFRRandom_NextDigitRange_Range(t *testing.T) {
	r := seed42()
	for i := 0; i < 500; i++ {
		d := r.NextDigitRange(3, 7)
		if d < 3 || d > 7 {
			t.Fatalf("NextDigitRange(3,7) = %d, out of [3,7]", d)
		}
	}
}

func TestFRRandom_NextDigitRange_Swapped(t *testing.T) {
	// min > max should still work via internal swap.
	r := seed42()
	for i := 0; i < 200; i++ {
		d := r.NextDigitRange(7, 3)
		if d < 3 || d > 7 {
			t.Fatalf("NextDigitRange(7,3) = %d, out of [3,7]", d)
		}
	}
}

// ---- NextDigits ----

func TestFRRandom_NextDigits_Length(t *testing.T) {
	r := seed42()
	for _, n := range []int{0, 1, 5, 10} {
		got := r.NextDigits(n)
		if len(got) != n {
			t.Errorf("NextDigits(%d) has length %d, want %d", n, len(got), n)
		}
		for _, ch := range got {
			if ch < '0' || ch > '9' {
				t.Errorf("NextDigits(%d) contains non-digit %q", n, ch)
			}
		}
	}
}

func TestFRRandom_NextDigits_Negative(t *testing.T) {
	r := seed42()
	if got := r.NextDigits(-1); got != "" {
		t.Errorf("NextDigits(-1) = %q, want empty string", got)
	}
}

// ---- NextByte / NextBytes ----

func TestFRRandom_NextByte_Range(t *testing.T) {
	r := seed42()
	for i := 0; i < 500; i++ {
		b := r.NextByte()
		// C# NextByte uses random.Next(byte.MaxValue) = [0,254]
		if b > 254 {
			t.Fatalf("NextByte() = %d, out of [0,254]", b)
		}
	}
}

func TestFRRandom_NextBytes_Length(t *testing.T) {
	r := seed42()
	b := r.NextBytes(8)
	if len(b) != 8 {
		t.Errorf("NextBytes(8) length = %d, want 8", len(b))
	}
}

func TestFRRandom_NextBytes_Zero(t *testing.T) {
	r := seed42()
	b := r.NextBytes(0)
	if len(b) != 0 {
		t.Errorf("NextBytes(0) length = %d, want 0", len(b))
	}
}

// ---- NextChar ----

func TestFRRandom_NextChar_Range(t *testing.T) {
	r := seed42()
	for i := 0; i < 500; i++ {
		c := r.NextChar()
		// C#: random.Next(char.MaxValue) → [0, 65534]
		if c < 0 || c > 65534 {
			t.Fatalf("NextChar() = %d, out of [0, 65534]", c)
		}
	}
}

// ---- NextDay ----

func TestFRRandom_NextDay_StartAfterToday(t *testing.T) {
	r := seed42()
	future := time.Now().AddDate(1, 0, 0)
	got := r.NextDay(future)
	today := time.Now().Truncate(24 * time.Hour)
	if !got.Equal(today) {
		t.Errorf("NextDay(future) = %v, want today %v", got, today)
	}
}

func TestFRRandom_NextDay_ValidRange(t *testing.T) {
	r := seed42()
	start := time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC)
	today := time.Now().Truncate(24 * time.Hour)
	for i := 0; i < 50; i++ {
		got := r.NextDay(start)
		if got.Before(start) {
			t.Fatalf("NextDay returned %v before start %v", got, start)
		}
		if got.After(today) {
			t.Fatalf("NextDay returned %v after today %v", got, today)
		}
	}
}

// ---- NextTimeSpanBetweenHours ----

func TestFRRandom_NextTimeSpanBetweenHours_Range(t *testing.T) {
	r := seed42()
	for i := 0; i < 200; i++ {
		d := r.NextTimeSpanBetweenHours(8, 17)
		if d < 8*time.Hour || d > 17*time.Hour {
			t.Fatalf("NextTimeSpanBetweenHours(8,17) = %v, out of [8h,17h]", d)
		}
	}
}

func TestFRRandom_NextTimeSpanBetweenHours_Clamped(t *testing.T) {
	r := seed42()
	// Clamping: start < 0, end > 24.
	d := r.NextTimeSpanBetweenHours(-5, 30)
	if d < 0 || d > 24*time.Hour {
		t.Fatalf("NextTimeSpanBetweenHours(-5,30) = %v, out of [0,24h]", d)
	}
}

func TestFRRandom_NextTimeSpanBetweenHours_Swapped(t *testing.T) {
	r := seed42()
	d := r.NextTimeSpanBetweenHours(20, 5)
	if d < 5*time.Hour || d > 20*time.Hour {
		t.Fatalf("NextTimeSpanBetweenHours(20,5) = %v, out of [5h,20h]", d)
	}
}

func TestFRRandom_NextTimeSpanBetweenHours_Equal(t *testing.T) {
	r := seed42()
	d := r.NextTimeSpanBetweenHours(10, 10)
	if d != 10*time.Hour {
		t.Fatalf("NextTimeSpanBetweenHours(10,10) = %v, want 10h", d)
	}
}

// ---- RandomizeDecimal / RandomizeDouble / RandomizeFloat32 ----

func TestFRRandom_RandomizeDecimal_NotSameValue(t *testing.T) {
	r := seed42()
	src := 12345.67
	got := r.RandomizeDecimal(src)
	if got == src {
		t.Error("RandomizeDecimal returned the original value unchanged (highly unlikely with seed 42)")
	}
}

func TestFRRandom_RandomizeDecimal_Negative(t *testing.T) {
	r := seed42()
	got := r.RandomizeDecimal(-42.5)
	if got >= 0 {
		t.Errorf("RandomizeDecimal(-42.5) = %v, expected negative", got)
	}
}

func TestFRRandom_RandomizeDecimal_Integer(t *testing.T) {
	r := seed42()
	// Integer with no fractional part: 1000 has 4 digits, result should be 4-digit.
	got := r.RandomizeDecimal(1000.0)
	if got < 1000 || got >= 10000 {
		t.Errorf("RandomizeDecimal(1000) = %v, expected 4-digit integer", got)
	}
}

func TestFRRandom_RandomizeDouble_NotEqual(t *testing.T) {
	r := seed42()
	src := 9876.543
	got := r.RandomizeDouble(src)
	if got == src {
		t.Error("RandomizeDouble returned original value (unexpected with seed 42)")
	}
}

func TestFRRandom_RandomizeFloat32_NotEqual(t *testing.T) {
	r := seed42()
	src := float32(123.45)
	got := r.RandomizeFloat32(src)
	if got == src {
		t.Errorf("RandomizeFloat32 returned original value %v (unexpected with seed 42)", src)
	}
}

// ---- RandomizeInt16 ----

func TestFRRandom_RandomizeInt16_SignPreserved(t *testing.T) {
	r := seed42()
	for _, src := range []int16{1, 10, 100, 1000, -1, -100} {
		got := r.RandomizeInt16(src)
		if src >= 0 && got < 0 {
			t.Errorf("RandomizeInt16(%d) = %d: sign changed unexpectedly", src, got)
		}
		if src < 0 && got > 0 {
			t.Errorf("RandomizeInt16(%d) = %d: sign changed unexpectedly", src, got)
		}
	}
}

func TestFRRandom_RandomizeInt16_Range(t *testing.T) {
	r := seed42()
	src := int16(1000) // 4 digits
	for i := 0; i < 50; i++ {
		got := r.RandomizeInt16(src)
		if got < 1000 || got > 9999 {
			t.Fatalf("RandomizeInt16(%d) = %d: 4-digit value out of [1000,9999]", src, got)
		}
	}
}

// ---- RandomizeInt32 ----

func TestFRRandom_RandomizeInt32_Range(t *testing.T) {
	r := seed42()
	src := int32(123456) // 6 digits
	for i := 0; i < 50; i++ {
		got := r.RandomizeInt32(src)
		if got < 100000 || got > 999999 {
			t.Fatalf("RandomizeInt32(%d) = %d: 6-digit value out of [100000,999999]", src, got)
		}
	}
}

func TestFRRandom_RandomizeInt32_Negative(t *testing.T) {
	r := seed42()
	got := r.RandomizeInt32(-500)
	if got >= 0 {
		t.Errorf("RandomizeInt32(-500) = %d, expected negative", got)
	}
}

// ---- RandomizeInt64 ----

func TestFRRandom_RandomizeInt64_Positive(t *testing.T) {
	r := seed42()
	src := int64(1234567890) // 10 digits
	got := r.RandomizeInt64(src)
	if got < 1000000000 || got > 9999999999 {
		t.Errorf("RandomizeInt64(%d) = %d: 10-digit value out of range", src, got)
	}
}

func TestFRRandom_RandomizeInt64_Negative(t *testing.T) {
	r := seed42()
	got := r.RandomizeInt64(-12345)
	if got >= 0 {
		t.Errorf("RandomizeInt64(-12345) = %d, expected negative", got)
	}
}

// ---- RandomizeSByte ----

func TestFRRandom_RandomizeSByte_Range(t *testing.T) {
	r := seed42()
	for i := 0; i < 200; i++ {
		got := r.RandomizeSByte(100)
		if got < -128 || got > 127 {
			t.Fatalf("RandomizeSByte(100) = %d, out of int8 range", got)
		}
	}
}

func TestFRRandom_RandomizeSByte_Negative(t *testing.T) {
	r := seed42()
	got := r.RandomizeSByte(-50)
	if got >= 0 {
		t.Errorf("RandomizeSByte(-50) = %d, expected negative", got)
	}
}

// ---- RandomizeUInt16 ----

func TestFRRandom_RandomizeUInt16_SameLength(t *testing.T) {
	r := seed42()
	src := uint16(1234) // 4 digits
	for i := 0; i < 50; i++ {
		got := r.RandomizeUInt16(src)
		if got < 1000 || got > 9999 {
			t.Fatalf("RandomizeUInt16(%d) = %d: 4-digit value out of [1000,9999]", src, got)
		}
	}
}

// ---- RandomizeUInt32 ----

func TestFRRandom_RandomizeUInt32_SameLength(t *testing.T) {
	r := seed42()
	src := uint32(123456) // 6 digits
	got := r.RandomizeUInt32(src)
	if got < 100000 || got > 999999 {
		t.Errorf("RandomizeUInt32(%d) = %d: 6-digit value out of range", src, got)
	}
}

// ---- RandomizeUInt64 ----

func TestFRRandom_RandomizeUInt64_SameLength(t *testing.T) {
	r := seed42()
	src := uint64(123456789) // 9 digits
	got := r.RandomizeUInt64(src)
	if got < 100000000 || got > 999999999 {
		t.Errorf("RandomizeUInt64(%d) = %d: 9-digit value out of range", src, got)
	}
}

// ---- RandomizeString ----

func TestFRRandom_RandomizeString_SameLength(t *testing.T) {
	r := seed42()
	src := "Hello, World! 123"
	got := r.RandomizeString(src)
	if len(got) != len(src) {
		t.Errorf("RandomizeString length %d != %d", len(got), len(src))
	}
}

func TestFRRandom_RandomizeString_PreservesSpaceAndPunct(t *testing.T) {
	r := seed42()
	src := "Hello, World!"
	got := r.RandomizeString(src)
	// Comma at index 5, space at 6, exclamation at 12 must be preserved.
	if got[5] != ',' {
		t.Errorf("comma not preserved at position 5: got %q", got[5])
	}
	if got[6] != ' ' {
		t.Errorf("space not preserved at position 6: got %q", got[6])
	}
	if got[12] != '!' {
		t.Errorf("exclamation not preserved at position 12: got %q", got[12])
	}
}

func TestFRRandom_RandomizeString_CasePreserved(t *testing.T) {
	r := seed42()
	src := "AbCdEf"
	got := r.RandomizeString(src)
	for i, ch := range got {
		srcCh := rune(src[i])
		if unicode.IsUpper(srcCh) && !unicode.IsUpper(ch) {
			t.Errorf("pos %d: expected upper, got %q", i, ch)
		}
		if unicode.IsLower(srcCh) && !unicode.IsLower(ch) {
			t.Errorf("pos %d: expected lower, got %q", i, ch)
		}
	}
}

func TestFRRandom_RandomizeString_DigitsReplaced(t *testing.T) {
	r := seed42()
	src := "123-456"
	got := r.RandomizeString(src)
	for i, ch := range got {
		srcCh := rune(src[i])
		if unicode.IsDigit(srcCh) && !unicode.IsDigit(ch) {
			t.Errorf("pos %d: expected digit, got %q", i, ch)
		}
	}
	// Hyphen must be preserved.
	if got[3] != '-' {
		t.Errorf("hyphen not preserved at position 3: got %q", got[3])
	}
}

func TestFRRandom_RandomizeString_Empty(t *testing.T) {
	r := seed42()
	if got := r.RandomizeString(""); got != "" {
		t.Errorf("RandomizeString('') = %q, want empty", got)
	}
}

// ---- GetRandomValue ----

func TestFRRandom_GetRandomValue_String(t *testing.T) {
	r := seed42()
	got := r.GetRandomValue("hello")
	s, ok := got.(string)
	if !ok {
		t.Fatalf("GetRandomValue(string) returned type %T", got)
	}
	if len(s) != 5 {
		t.Errorf("GetRandomValue(string) length %d, want 5", len(s))
	}
}

func TestFRRandom_GetRandomValue_Int32(t *testing.T) {
	r := seed42()
	got := r.GetRandomValue(int32(999))
	if _, ok := got.(int32); !ok {
		t.Fatalf("GetRandomValue(int32) returned type %T", got)
	}
}

func TestFRRandom_GetRandomValue_Float64(t *testing.T) {
	r := seed42()
	got := r.GetRandomValue(float64(3.14))
	if _, ok := got.(float64); !ok {
		t.Fatalf("GetRandomValue(float64) returned type %T", got)
	}
}

func TestFRRandom_GetRandomValue_Time(t *testing.T) {
	r := seed42()
	got := r.GetRandomValue(time.Now())
	if _, ok := got.(time.Time); !ok {
		t.Fatalf("GetRandomValue(time.Time) returned type %T", got)
	}
}

func TestFRRandom_GetRandomValue_Duration(t *testing.T) {
	r := seed42()
	got := r.GetRandomValue(time.Hour)
	if _, ok := got.(time.Duration); !ok {
		t.Fatalf("GetRandomValue(time.Duration) returned type %T", got)
	}
}

func TestFRRandom_GetRandomValue_Bytes(t *testing.T) {
	r := seed42()
	src := []byte{1, 2, 3}
	got := r.GetRandomValue(src)
	b, ok := got.([]byte)
	if !ok {
		t.Fatalf("GetRandomValue([]byte) returned type %T", got)
	}
	if len(b) != 3 {
		t.Errorf("GetRandomValue([]byte) length %d, want 3", len(b))
	}
}

func TestFRRandom_GetRandomValue_Unsupported(t *testing.T) {
	r := seed42()
	type myStruct struct{ X int }
	src := myStruct{X: 42}
	got := r.GetRandomValue(src)
	if got != src {
		t.Errorf("GetRandomValue(unsupported) = %v, want original %v", got, src)
	}
}

func TestFRRandom_GetRandomValue_AllBasicTypes(t *testing.T) {
	r := seed42()
	cases := []any{
		int(10),
		int8(int8(10)),
		int16(int16(100)),
		int64(int64(100000)),
		uint8(uint8(100)),
		uint16(uint16(1000)),
		uint32(uint32(100000)),
		uint64(uint64(1000000)),
		float32(float32(1.5)),
		float64(float64(2.5)),
	}
	for _, src := range cases {
		got := r.GetRandomValue(src)
		if got == nil {
			t.Errorf("GetRandomValue(%T) returned nil", src)
		}
	}
}

// ---- FRRandomFieldValueCollection ----

func TestFRRandomFieldValueCollection_AddAndContains(t *testing.T) {
	c := &utils.FRRandomFieldValueCollection{}
	c.Add(utils.FRRandomFieldValue{Origin: "orig1", Rand: "rand1"})
	c.Add(utils.FRRandomFieldValue{Origin: "orig2", Rand: "rand2"})

	if !c.ContainsOrigin("orig1") {
		t.Error("ContainsOrigin('orig1') should be true")
	}
	if c.ContainsOrigin("missing") {
		t.Error("ContainsOrigin('missing') should be false")
	}
	if !c.ContainsRandom("rand2") {
		t.Error("ContainsRandom('rand2') should be true")
	}
	if c.ContainsRandom("missing") {
		t.Error("ContainsRandom('missing') should be false")
	}
}

func TestFRRandomFieldValueCollection_GetRandom(t *testing.T) {
	c := &utils.FRRandomFieldValueCollection{}
	c.Add(utils.FRRandomFieldValue{Origin: "a", Rand: "x"})
	c.Add(utils.FRRandomFieldValue{Origin: "b", Rand: "y"})

	if got := c.GetRandom("a"); got != "x" {
		t.Errorf("GetRandom('a') = %v, want 'x'", got)
	}
	if got := c.GetRandom("b"); got != "y" {
		t.Errorf("GetRandom('b') = %v, want 'y'", got)
	}
	// Not found → original returned.
	if got := c.GetRandom("z"); got != "z" {
		t.Errorf("GetRandom('z') = %v, want 'z' (original)", got)
	}
}

// ---- NewFRRandom (time-seeded) ----

func TestNewFRRandom_NotNil(t *testing.T) {
	r := utils.NewFRRandom()
	if r == nil {
		t.Fatal("NewFRRandom() returned nil")
	}
	d := r.NextDigit()
	if d < 0 || d > 9 {
		t.Errorf("digit out of range: %d", d)
	}
}

// ---- FRColumnInfo ----

func TestFRColumnInfo_Fields(t *testing.T) {
	ci := utils.FRColumnInfo{TypeName: "int32", Length: 10}
	if ci.TypeName != "int32" {
		t.Errorf("TypeName = %q, want 'int32'", ci.TypeName)
	}
	if ci.Length != 10 {
		t.Errorf("Length = %d, want 10", ci.Length)
	}
}

// ---- Concurrency smoke test ----

func TestFRRandom_ConcurrentSafe(t *testing.T) {
	r := utils.NewFRRandom()
	done := make(chan struct{})
	for i := 0; i < 10; i++ {
		go func() {
			for j := 0; j < 100; j++ {
				r.NextDigit()
				r.NextLetter('a')
				r.RandomizeString("hello world")
				r.RandomizeInt32(12345)
			}
			done <- struct{}{}
		}()
	}
	for i := 0; i < 10; i++ {
		<-done
	}
}

// ---- Additional edge-case coverage ----

func TestFRRandom_RandomizeInt32_MaxLength(t *testing.T) {
	// Source at max int32 digit count (10 digits) triggers overflow path.
	r := seed42()
	src := int32(1234567890) // 10 digits = maxLen for int32
	for i := 0; i < 30; i++ {
		got := r.RandomizeInt32(src)
		if got <= 0 {
			t.Fatalf("RandomizeInt32(max-length source) = %d, expected positive", got)
		}
	}
}

func TestFRRandom_RandomizeInt64_MaxLength(t *testing.T) {
	// Source at max int64 digit count (19 digits) triggers overflow path.
	r := seed42()
	src := int64(1000000000000000000) // 19 digits = maxLen for int64
	got := r.RandomizeInt64(src)
	if got <= 0 {
		t.Fatalf("RandomizeInt64(max-length source) = %d, expected positive", got)
	}
}

func TestFRRandom_RandomizeUInt16_MaxLength(t *testing.T) {
	// Source with 5 digits (maxLen for uint16) triggers overflow path.
	r := seed42()
	src := uint16(10000) // 5 digits
	for i := 0; i < 30; i++ {
		got := r.RandomizeUInt16(src)
		// Result must fit in uint16 (0..65535)
		if got == 0 {
			t.Fatalf("RandomizeUInt16(max-length source) = 0 (unexpected)", )
		}
	}
}

func TestFRRandom_RandomizeUInt64_MaxLength(t *testing.T) {
	// Source with 20 digits triggers the overflow path.
	r := seed42()
	src := uint64(10000000000000000000) // 20 digits = maxLen for uint64
	got := r.RandomizeUInt64(src)
	// Should start with "1" and be a valid uint64.
	s := fmt.Sprintf("%d", got)
	if s[0] != '1' {
		t.Errorf("RandomizeUInt64(max-length) = %d, expected to start with '1'", got)
	}
}

func TestFRRandom_RandomizeSByte_MaxLength(t *testing.T) {
	// Source with 3 digits (maxLen for int8) triggers the overflow path.
	r := seed42()
	src := int8(100) // 3 digits = maxLen for int8
	for i := 0; i < 50; i++ {
		got := r.RandomizeSByte(src)
		if got < 0 {
			t.Fatalf("RandomizeSByte(100) = %d, expected non-negative for positive source", got)
		}
	}
}

func TestFRRandom_RandomizeUInt32_MaxLength(t *testing.T) {
	// Source with 10 digits (maxLen for uint32) triggers overflow path.
	r := seed42()
	src := uint32(1000000000) // 10 digits
	got := r.RandomizeUInt32(src)
	if got == 0 {
		t.Fatalf("RandomizeUInt32(max-length source) = 0 (unexpected)")
	}
}

func TestFRRandom_RandomizeInt16_MaxLength(t *testing.T) {
	// Source with 5 digits (maxLen for int16) triggers overflow path.
	r := seed42()
	src := int16(10000) // 5 digits
	for i := 0; i < 30; i++ {
		got := r.RandomizeInt16(src)
		if got <= 0 {
			t.Fatalf("RandomizeInt16(max-length source) = %d, expected positive", got)
		}
	}
}

// TestFRRandom_NextDigitMax_Zero checks NextDigitMax(0) always returns 0.
func TestFRRandom_NextDigitMax_Zero(t *testing.T) {
	r := seed42()
	for i := 0; i < 20; i++ {
		if d := r.NextDigitMax(0); d != 0 {
			t.Fatalf("NextDigitMax(0) = %d, want 0", d)
		}
	}
}

// TestFRRandom_NextDigits_One checks a single digit is returned.
func TestFRRandom_NextDigits_One(t *testing.T) {
	r := seed42()
	got := r.NextDigits(1)
	if len(got) != 1 || got[0] < '0' || got[0] > '9' {
		t.Errorf("NextDigits(1) = %q, want single digit", got)
	}
}

// TestFRRandom_RandomizeString_AllWhitespace checks whitespace passthrough.
func TestFRRandom_RandomizeString_AllWhitespace(t *testing.T) {
	r := seed42()
	src := "   \t\n"
	got := r.RandomizeString(src)
	if got != src {
		t.Errorf("RandomizeString(whitespace) = %q, want %q", got, src)
	}
}

// TestFRRandom_NextDay_SameDay checks that when start == today the result equals start.
func TestFRRandom_NextDay_SameDay(t *testing.T) {
	r := seed42()
	today := time.Now().Truncate(24 * time.Hour)
	got := r.NextDay(today)
	if !got.Equal(today) {
		t.Errorf("NextDay(today) = %v, want %v", got, today)
	}
}

// Ensure the fmt import is used by tests that need it.
var _ = strings.Contains
var _ = fmt.Sprintf
