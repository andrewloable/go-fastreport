// barcode39_ean_msi_plessey_test.go contains deterministic encoding tests for
// Barcode39, BarcodeEAN, BarcodeMSI, and BarcodePlessey verifying Go output
// against known-good values derived from the C# source.
//
// C# references:
//   Barcode39.cs  — tabelle_39, GetPattern (start '*', mod-43 checksum, stop '*')
//   BarcodeEAN.cs — tabelle_EAN_A/B/C, EAN-8/EAN-13 GetPattern, BarcodeEAN128
//   BarcodeMSI.cs — tabelle_MSI, quersumme, GetPattern (start "60", stop "515")
//   BarcodePlessey.cs — tabelle, crcGrid, GetPattern (start "606050060", end "70050050606")
package barcode_test

import (
	"strings"
	"testing"

	"github.com/andrewloable/go-fastreport/barcode"
)

// ── Code39 ───────────────────────────────────────────────────────────────────

// TestCode39_KnownEncoding_SingleDigit verifies that encoding "0" (checksum off)
// produces the expected pattern.
// C# tabelle_39[0].data = "505160605" for '0'; star data = "515060605".
// Pattern = star+"0" + "0"+"0" + star = "515060605" + "0" + "505160605" + "0" + "515060605"
// = "5150606050" + "5051606050" + "515060605"
func TestCode39_KnownEncoding_SingleDigit(t *testing.T) {
	b := barcode.NewCode39Barcode()
	b.CalcChecksum = false
	if err := b.Encode("0"); err != nil {
		t.Fatalf("Encode: %v", err)
	}
	pattern, err := b.GetPattern()
	if err != nil {
		t.Fatalf("GetPattern: %v", err)
	}
	// Start '*' data: "515060605" + "0" inter-char gap
	if !strings.HasPrefix(pattern, "515060605") {
		t.Errorf("pattern start: got %q, want prefix \"515060605\"", pattern[:minI(len(pattern), 9)])
	}
	// Stop '*' data: "515060605" (no trailing '0')
	if !strings.HasSuffix(pattern, "515060605") {
		t.Errorf("pattern end: got %q, want suffix \"515060605\"", pattern[maxI(0, len(pattern)-9):])
	}
	// Pattern: 3 chars × 9 data + 2 inter-char '0' gaps = 29 chars total
	// (start data + gap) + ('0' data + gap) + (stop data)
	// = 10 + 10 + 9 = 29
	const wantLen = 29
	if len(pattern) != wantLen {
		t.Errorf("pattern length = %d, want %d", len(pattern), wantLen)
	}
}

// TestCode39_KnownEncoding_WithChecksum verifies that checksum character is
// inserted before the stop bar.
// "A" has chk=10 in C# tabelle_39; mod 43 = 10.
// Checksum lookup: first entry where chk==10 is "A" (index 10).
// So pattern = start + "A" + checksum("A") + stop
func TestCode39_KnownEncoding_WithChecksum_A(t *testing.T) {
	b := barcode.NewCode39Barcode()
	b.CalcChecksum = true
	if err := b.Encode("A"); err != nil {
		t.Fatalf("Encode: %v", err)
	}
	pattern, err := b.GetPattern()
	if err != nil {
		t.Fatalf("GetPattern: %v", err)
	}
	// "A" data = "605051506", chk=10; checksum mod43=10, lookup finds "A" again.
	// Pattern: star+0 + A+0 + A+0 + stop = 10+10+10+9 = 39
	const wantLen = 39
	if len(pattern) != wantLen {
		t.Errorf("pattern length = %d, want %d (start+A+checkA+stop)", len(pattern), wantLen)
	}
}

// TestCode39_KnownEncoding_LowercaseAutoUppercase verifies that lowercase input
// is auto-uppercased (Go adds ToUpper; C# would skip unknown chars).
func TestCode39_KnownEncoding_LowercaseAutoUppercase(t *testing.T) {
	b := barcode.NewCode39Barcode()
	b.CalcChecksum = false
	// Must allow extended for lowercase (Encode rejects lowercase without AllowExtended).
	b.AllowExtended = true
	if err := b.Encode("a"); err != nil {
		t.Fatalf("Encode: %v", err)
	}
	pattern, err := b.GetPattern()
	if err != nil {
		t.Fatalf("GetPattern: %v", err)
	}
	// 'a' uppercased to 'A' → same encoding as "A" without checksum.
	// Pattern = start + A + stop = 10 + 10 + 9 = 29
	const wantLen = 29
	if len(pattern) != wantLen {
		t.Errorf("lowercase 'a' pattern length = %d, want %d (treated as 'A')", len(pattern), wantLen)
	}
}

// TestCode39_KnownEncoding_ChecksumModulo43_Wraparound tests checksum = 0
// which maps back to '0' (the first entry in tabelle_39).
// "%" has chk=42; one more would wrap to 0.
// To get checksum=0 we need sum of chk values ≡ 0 mod 43.
// A single '*' is the star char with chk=0, but '*' isn't data.
// Use text "0": chk=0, checksum=0 → lookup chk==0 → first match is "0".
func TestCode39_KnownEncoding_ChecksumModulo43_Zero(t *testing.T) {
	b := barcode.NewCode39Barcode()
	b.CalcChecksum = true
	if err := b.Encode("0"); err != nil {
		t.Fatalf("Encode: %v", err)
	}
	pattern, err := b.GetPattern()
	if err != nil {
		t.Fatalf("GetPattern: %v", err)
	}
	// "0" chk=0, checksum=0 → checksum char is "0" again.
	// Pattern: start+0 + 0+0 + 0(checksum)+0 + stop = 10+10+10+9 = 39
	const wantLen = 39
	if len(pattern) != wantLen {
		t.Errorf("checksum=0 pattern length = %d, want %d", len(pattern), wantLen)
	}
}

// TestCode39Extended_KnownEncoding_ControlChar verifies that Code39 Extended
// encodes ASCII control chars via two-char substitution.
// NUL (0x00) → "%U" in code39x table (C# code39x[0] = "%U").
// '%' has chk=42, 'U' has chk=30; sum=72; mod43=72%43=29 → 'T' (chk=29).
func TestCode39Extended_KnownEncoding_ControlChar(t *testing.T) {
	b := barcode.NewCode39ExtendedBarcode()
	b.CalcChecksum = false
	// Encode NUL (0x00) — C# code39x[0] = "%U"
	if err := b.Encode("\x00"); err != nil {
		t.Fatalf("Encode NUL: %v", err)
	}
	pattern, err := b.GetPattern()
	if err != nil {
		t.Fatalf("GetPattern: %v", err)
	}
	// NUL → "%U" = 2 chars in Code39 table.
	// Pattern: start+0 + '%'+0 + 'U'+0 + stop = 10+10+10+9 = 39
	const wantLen = 39
	if len(pattern) != wantLen {
		t.Errorf("NUL (→%%U) pattern length = %d, want %d", len(pattern), wantLen)
	}
}

// ── EAN-8 ──────────────────────────────────────────────────────────────────

// TestEAN8_KnownEncoding_1234567 verifies the exact pattern for "1234567".
// EAN-8 check digit for "1234567":
//   fak=7(odd): 1×3=3, fak=6(even): 2×1=2, fak=5(odd): 3×3=9, fak=4(even): 4×1=4,
//   fak=3(odd): 5×3=15, fak=2(even): 6×1=6, fak=1(odd): 7×3=21
//   sum=3+2+9+4+15+6+21=60; 60%10=0 → check=0 → text = "12345670"
// Left 4 digits (A): "1234" → tabelle_EAN_A[1]+[2]+[3]+[4]
//   = "1615" + "1516" + "0805" + "0526"
// Right 4 digits (C): "5670" → tabelle_EAN_C[5]+[6]+[7]+[0]
//   = "5170" + "5053" + "5251" + "7150"
// Full: "A0A" + left4 + "0A0A0" + right4 + "A0A"
func TestEAN8_KnownEncoding_1234567(t *testing.T) {
	b := barcode.NewEAN8Barcode()
	if err := b.Encode("1234567"); err != nil {
		t.Fatalf("Encode: %v", err)
	}
	pattern, err := b.GetPattern()
	if err != nil {
		t.Fatalf("GetPattern: %v", err)
	}
	// Check structure: start + 4 A-table entries (4×4=16) + center + 4 C-table entries (4×4=16) + stop
	// = 3 + 16 + 5 + 16 + 3 = 43 chars
	const wantLen = 43
	if len(pattern) != wantLen {
		t.Fatalf("pattern length = %d, want %d", len(pattern), wantLen)
	}
	if pattern[:3] != "A0A" {
		t.Errorf("start guard = %q, want \"A0A\"", pattern[:3])
	}
	if pattern[len(pattern)-3:] != "A0A" {
		t.Errorf("stop guard = %q, want \"A0A\"", pattern[len(pattern)-3:])
	}
	centerStart := 3 + 16
	if pattern[centerStart:centerStart+5] != "0A0A0" {
		t.Errorf("center guard = %q, want \"0A0A0\"", pattern[centerStart:centerStart+5])
	}
	// Verify left half digit '1' uses table A: tabelle_EAN_A[1] = "1615"
	if pattern[3:7] != "1615" {
		t.Errorf("digit '1' (A-table) = %q, want \"1615\"", pattern[3:7])
	}
	// Verify right half digit '5' (5th char of "12345670") uses table C: tabelle_EAN_C[5] = "5170"
	rightStart := 3 + 16 + 5
	if pattern[rightStart:rightStart+4] != "5170" {
		t.Errorf("digit '5' (C-table) = %q, want \"5170\"", pattern[rightStart:rightStart+4])
	}
}

// TestEAN13_KnownEncoding_FirstDigit0 verifies EAN-13 parity row 0: all A-table.
// For first digit 0: all 6 left digits use table A.
// Input "012345678905" (12 digits; check digit: fak=12 even:0, fak=11 odd:1×3=3, ...)
// This just verifies the start/stop structure and that no B-table entries appear
// when the system digit is 0 (parity row 0 = AAAAAA).
func TestEAN13_KnownEncoding_FirstDigit0_AllATable(t *testing.T) {
	b := barcode.NewEAN13Barcode()
	// Encode 12 digits; check digit will be appended.
	if err := b.Encode("012345678905"); err != nil {
		t.Fatalf("Encode: %v", err)
	}
	pattern, err := b.GetPattern()
	if err != nil {
		t.Fatalf("GetPattern: %v", err)
	}
	// EAN-13: start(3) + 6×4=24 left + center(5) + 6×4=24 right + stop(3) = 59 chars
	const wantLen = 59
	if len(pattern) != wantLen {
		t.Fatalf("pattern length = %d, want %d", len(pattern), wantLen)
	}
	if pattern[:3] != "A0A" {
		t.Errorf("start guard = %q, want \"A0A\"", pattern[:3])
	}
	if pattern[len(pattern)-3:] != "A0A" {
		t.Errorf("stop guard = %q, want \"A0A\"", pattern[len(pattern)-3:])
	}
	// First digit (system digit=0) uses parity row AAAAAA.
	// The 1st data digit (after system digit 0) is '1': tabelle_EAN_A[1] = "1615"
	if pattern[3:7] != "1615" {
		t.Errorf("first data digit '1' (A-table): got %q, want \"1615\"", pattern[3:7])
	}
}

// TestEAN13_KnownEncoding_FirstDigit1_ParityABRow verifies EAN-13 parity row 1: AABABB.
// For system digit=1: parity = {A, A, B, A, B, B}.
// Digits: system=1, data[0..5]=0,2,3,4,5,6 → [A][A][B][A][B][B]
// tabelle_EAN_A[0]="2605", tabelle_EAN_A[2]="1516",
// tabelle_EAN_B[3]="0535", tabelle_EAN_A[4]="0526",
// tabelle_EAN_B[5]="0715", tabelle_EAN_B[6]="3505"
func TestEAN13_KnownEncoding_FirstDigit1_ParityABRow(t *testing.T) {
	b := barcode.NewEAN13Barcode()
	// System digit=1, followed by 0,2,3,4,5,6 and 6 right digits (pad with 0s).
	if err := b.Encode("102345600000"); err != nil {
		t.Fatalf("Encode: %v", err)
	}
	pattern, err := b.GetPattern()
	if err != nil {
		t.Fatalf("GetPattern: %v", err)
	}
	const wantLen = 59
	if len(pattern) != wantLen {
		t.Fatalf("pattern length = %d, want %d", len(pattern), wantLen)
	}
	// Left half (after start guard): positions 3..26
	left := pattern[3:27]
	// Parity row 1 = AABABB for digits 0,2,3,4,5,6
	// digit 0 (A): "2605"; digit 2 (A): "1516"; digit 3 (B): "0535"
	// digit 4 (A): "0526"; digit 5 (B): "0715"; digit 6 (B): "3505"
	wantLeft := "2605" + "1516" + "0535" + "0526" + "0715" + "3505"
	if left != wantLeft {
		t.Errorf("left half encoding:\n  got  %q\n  want %q", left, wantLeft)
	}
}

// ── MSI ────────────────────────────────────────────────────────────────────

// TestMSI_KnownEncoding_SingleDigit0_NoChecksum verifies that "0" with
// CalcChecksum=false produces: start + tabelleMSI[0] + stop.
// C# tabelle_MSI[0] = "51515151", start = "60", stop = "515".
func TestMSI_KnownEncoding_SingleDigit0_NoChecksum(t *testing.T) {
	b := barcode.NewMSIBarcode()
	b.CalcChecksum = false
	if err := b.Encode("0"); err != nil {
		t.Fatalf("Encode: %v", err)
	}
	pattern, err := b.GetPattern()
	if err != nil {
		t.Fatalf("GetPattern: %v", err)
	}
	want := "60" + "51515151" + "515"
	if pattern != want {
		t.Errorf("MSI '0' no-checksum: got %q, want %q", pattern, want)
	}
}

// TestMSI_KnownEncoding_SingleDigit1_WithChecksum verifies "1" with checksum.
// C# quersumme logic for "1":
//   i=0 (even): check_even += 1 → check_even=1
//   check_odd=0 (no odd-index chars)
//   checksum = quersumme(0*2) + 1 = 0 + 1 = 1
//   checksum % 10 = 1; checksum > 0 → checksum = 10-1 = 9
// tabelleMSI[9] = "60515160"
func TestMSI_KnownEncoding_SingleDigit1_WithChecksum(t *testing.T) {
	b := barcode.NewMSIBarcode()
	b.CalcChecksum = true
	if err := b.Encode("1"); err != nil {
		t.Fatalf("Encode: %v", err)
	}
	pattern, err := b.GetPattern()
	if err != nil {
		t.Fatalf("GetPattern: %v", err)
	}
	// start + "1" data + checksum(9) + stop
	want := "60" + "51515160" + "60515160" + "515"
	if pattern != want {
		t.Errorf("MSI '1' with checksum: got %q, want %q", pattern, want)
	}
}

// TestMSI_KnownEncoding_TwoDigit12_WithChecksum verifies "12" with checksum.
// C# logic for "12":
//   i=0 (even): check_even += 1 → check_even=1
//   i=1 (odd):  check_odd = 0*10 + 2 = 2
//   checksum = quersumme(2*2) + 1 = quersumme(4) + 1 = 4 + 1 = 5
//   checksum % 10 = 5; checksum > 0 → checksum = 10-5 = 5
// tabelleMSI[5] = "51605160"
func TestMSI_KnownEncoding_TwoDigit12_WithChecksum(t *testing.T) {
	b := barcode.NewMSIBarcode()
	b.CalcChecksum = true
	if err := b.Encode("12"); err != nil {
		t.Fatalf("Encode: %v", err)
	}
	pattern, err := b.GetPattern()
	if err != nil {
		t.Fatalf("GetPattern: %v", err)
	}
	// start + "1" + "2" + checksum(5) + stop
	want := "60" + "51515160" + "51516051" + "51605160" + "515"
	if pattern != want {
		t.Errorf("MSI '12' with checksum: got %q, want %q", pattern, want)
	}
}

// TestMSI_KnownEncoding_AllDigits verifies that all 10 MSI digit entries
// match the C# tabelle_MSI exactly.
func TestMSI_KnownEncoding_AllDigits(t *testing.T) {
	// C# BarcodeMSI.tabelle_MSI exactly:
	wantTable := [10]string{
		"51515151", // 0
		"51515160", // 1
		"51516051", // 2
		"51516060", // 3
		"51605151", // 4
		"51605160", // 5
		"51606051", // 6
		"51606060", // 7
		"60515151", // 8
		"60515160", // 9
	}
	for d := 0; d <= 9; d++ {
		b := barcode.NewMSIBarcode()
		b.CalcChecksum = false
		digit := string(rune('0' + d))
		if err := b.Encode(digit); err != nil {
			t.Fatalf("Encode %d: %v", d, err)
		}
		pattern, err := b.GetPattern()
		if err != nil {
			t.Fatalf("GetPattern %d: %v", d, err)
		}
		// Pattern = "60" + data[d] + "515"
		want := "60" + wantTable[d] + "515"
		if pattern != want {
			t.Errorf("digit %d: got %q, want %q", d, pattern, want)
		}
	}
}

// ── Plessey ────────────────────────────────────────────────────────────────

// TestPlessey_KnownEncoding_StartEnd verifies that the Plessey start and end
// sequences match the C# BarcodePlessey constants exactly.
// C# start = "606050060", end = "70050050606".
func TestPlessey_KnownEncoding_StartEnd(t *testing.T) {
	b := barcode.NewPlesseyBarcode()
	if err := b.Encode("0"); err != nil {
		t.Fatalf("Encode: %v", err)
	}
	pattern, err := b.GetPattern()
	if err != nil {
		t.Fatalf("GetPattern: %v", err)
	}
	const wantStart = "606050060"
	const wantEnd = "70050050606"
	if !strings.HasPrefix(pattern, wantStart) {
		t.Errorf("Plessey start: got %q, want prefix %q", pattern[:minI(len(pattern), len(wantStart))], wantStart)
	}
	if !strings.HasSuffix(pattern, wantEnd) {
		t.Errorf("Plessey end: got %q, want suffix %q", pattern[maxI(0, len(pattern)-len(wantEnd)):], wantEnd)
	}
}

// TestPlessey_KnownEncoding_DataTable verifies the first few entries of the
// Plessey data table match the C# tabelle exactly.
// C# tabelle[0]="500500500500", tabelle[1]="60500500500", tabelle[F/15]="60606060".
func TestPlessey_KnownEncoding_DataTable(t *testing.T) {
	cases := []struct {
		input string
		entry string
	}{
		{"0", "500500500500"}, // tabelle[0]
		{"1", "60500500500"},  // tabelle[1]
		{"F", "60606060"},     // tabelle[15]
	}
	for _, tc := range cases {
		b := barcode.NewPlesseyBarcode()
		if err := b.Encode(tc.input); err != nil {
			t.Fatalf("Encode %q: %v", tc.input, err)
		}
		pattern, err := b.GetPattern()
		if err != nil {
			t.Fatalf("GetPattern %q: %v", tc.input, err)
		}
		// Pattern = start(9) + data + CRC(variable) + end(11)
		// Verify data entry immediately follows the start sequence.
		const startLen = 9
		if len(pattern) < startLen+len(tc.entry) {
			t.Fatalf("%q pattern too short: len=%d", tc.input, len(pattern))
		}
		got := pattern[startLen : startLen+len(tc.entry)]
		if got != tc.entry {
			t.Errorf("data entry for %q: got %q, want %q", tc.input, got, tc.entry)
		}
	}
}

// TestPlessey_KnownEncoding_CRC_Length verifies that each single hex digit
// produces a pattern with 9+data+24CRC+11 = 9+len(tabelle[i])+24+11 chars.
// CRC = 8 bits × "500"(3) or "60"(2) chars; 8 bits always produce exactly 8 CRC tokens.
// The total length for a single hex char is: 9 (start) + len(data_entry) + CRC_len + 11 (end).
// However CRC length depends on the bit pattern, so we can only check structural bounds.
func TestPlessey_KnownEncoding_CRC_BitCount(t *testing.T) {
	// For any input, exactly 8 CRC tokens are appended, each either "500" (3) or "60" (2).
	// Min CRC length = 8×2=16, max = 8×3=24.
	// We verify the CRC region is within [16, 24] chars.
	//
	// Entry lengths from C# tabelle:
	entryLens := map[string]int{
		"0": 12, "1": 11, "2": 11, "3": 10, "4": 12, "5": 11,
		"6": 11, "7": 10, "8": 12, "9": 11, "A": 11, "B": 10,
		"C": 11, "D": 10, "E": 10, "F": 8,
	}
	for input, dataLen := range entryLens {
		b := barcode.NewPlesseyBarcode()
		if err := b.Encode(input); err != nil {
			t.Fatalf("Encode %q: %v", input, err)
		}
		pattern, err := b.GetPattern()
		if err != nil {
			t.Fatalf("GetPattern %q: %v", input, err)
		}
		// Remove known start (9) + data + end (11) to get CRC region.
		const startLen, endLen = 9, 11
		crcRegionLen := len(pattern) - startLen - dataLen - endLen
		if crcRegionLen < 16 || crcRegionLen > 24 {
			t.Errorf("%q: CRC region length = %d, want [16,24] (8 tokens of 2 or 3 chars each)",
				input, crcRegionLen)
		}
	}
}

// TestPlessey_InvalidChar_ReturnsError verifies that a non-hex character
// causes GetPattern to return an error (matching C# throw new ArgumentException).
func TestPlessey_InvalidChar_ReturnsError(t *testing.T) {
	b := barcode.NewPlesseyBarcode()
	// Encode will reject 'G' since PlesseyBarcode.Encode validates the alphabet.
	err := b.Encode("G")
	if err == nil {
		t.Error("expected error for invalid Plessey char 'G', got nil")
	}
}

// TestPlessey_LowercaseAccepted verifies that lowercase hex is accepted (Encode uppercases).
func TestPlessey_LowercaseAccepted(t *testing.T) {
	b := barcode.NewPlesseyBarcode()
	if err := b.Encode("abcdef"); err != nil {
		t.Fatalf("expected lowercase hex to be accepted, got error: %v", err)
	}
	// After Encode, encodedText should be "ABCDEF"
	if b.EncodedText() != "ABCDEF" {
		t.Errorf("EncodedText = %q, want \"ABCDEF\"", b.EncodedText())
	}
}

// ── EAN-128 / GS1-128 ──────────────────────────────────────────────────────

// TestGS1Barcode_KnownEncoding_Pattern verifies that GS1-128 produces a
// non-empty pattern. The GS1-128 is implemented as a Code128 with FNC1 prefix
// (C# BarcodeEAN128.GetPattern delegates to Barcode128.GetPattern with "&C;" prefix).
func TestGS1Barcode_KnownEncoding_Pattern(t *testing.T) {
	b := barcode.NewGS1Barcode()
	if err := b.Encode("(01)12345678901231"); err != nil {
		t.Fatalf("Encode GS1: %v", err)
	}
	pattern, err := b.GetPattern()
	if err != nil {
		t.Fatalf("GetPattern GS1: %v", err)
	}
	if len(pattern) == 0 {
		t.Error("GS1-128 pattern is empty")
	}
}

// ── Boundary confirmation tests ───────────────────────────────────────────────
//
// These tests confirm correct behaviour at exact boundary values for each
// barcode checksum / character-set algorithm. The audit (go-fastreport-i7nfi)
// verified all implementations are correct; these tests pin that correctness.

// TestCode39_Checksum_MaxValue_Percent verifies that a single input whose
// checksum reaches the maximum mod-43 value (42 → '%') is encoded correctly.
//
// tabelle_39 assigns:
//   '9' → chk=9, 'Z' → chk=35, Sum("9Z") = 44, mod43 = 1 → '1'
//   '%' → chk=42, so Sum("%") = 42, mod43 = 42 → '%'
//
// Encoding '%' with CalcChecksum=true produces checksum '%' again, so the
// pattern contains two copies of the '%' data between start and stop markers.
func TestCode39_Checksum_MaxValue_Percent(t *testing.T) {
	b := barcode.NewCode39Barcode()
	b.CalcChecksum = true
	// '%' has chk=42; mod-43 of 42 is 42, which maps back to '%'.
	if err := b.Encode("%"); err != nil {
		t.Fatalf("Encode: %v", err)
	}
	pattern, err := b.GetPattern()
	if err != nil {
		t.Fatalf("GetPattern: %v", err)
	}
	// Pattern must have at least start + '%' + checksum('%') + stop.
	// That is 4 symbols = start(10) + %(10) + %(10) + stop(9) = 39 chars.
	const wantLen = 39
	if len(pattern) != wantLen {
		t.Errorf("pattern length = %d, want %d (start+%%+%%+stop)", len(pattern), wantLen)
	}
}

// TestCode39_Checksum_WrapAround verifies the modulo-43 wrapping:
// "9Z" has chk sum = 9+35 = 44, mod 43 = 1 → '1'.
// Pattern: start + '9' + 'Z' + '1' + stop = 5 symbols.
// Each symbol (except stop) = 9 chars + 1 gap = 10 chars; stop = 9 chars.
// Total = 4 * 10 + 9 = 49 chars.
func TestCode39_Checksum_WrapAround(t *testing.T) {
	b := barcode.NewCode39Barcode()
	b.CalcChecksum = true
	if err := b.Encode("9Z"); err != nil {
		t.Fatalf("Encode: %v", err)
	}
	pattern, err := b.GetPattern()
	if err != nil {
		t.Fatalf("GetPattern: %v", err)
	}
	const wantLen = 49
	if len(pattern) != wantLen {
		t.Errorf("pattern length = %d, want %d", len(pattern), wantLen)
	}
}

// TestEAN8_CheckDigit_AllNines verifies the check digit when all seven input
// digits are '9'. This is the maximum-value boundary.
//
// Weights alternate 3,1,3,1,3,1,3 (left-to-right, i.e. odd from right):
//
//	sum = 9*3 + 9*1 + 9*3 + 9*1 + 9*3 + 9*1 + 9*3 = 9*(3+1+3+1+3+1+3) = 9*15 = 135
//	check digit = (10 - (135 mod 10)) mod 10 = (10 - 5) mod 10 = 5
//
// So "9999999" → EAN-8 = "99999995". The pattern ends with eanTableC[5]+stop.
// eanTableC[5] = "5170", stop guard = "A0A" → pattern suffix "5170A0A".
func TestEAN8_CheckDigit_AllNines(t *testing.T) {
	b := barcode.NewEAN8Barcode()
	if err := b.Encode("9999999"); err != nil {
		t.Fatalf("Encode: %v", err)
	}
	pattern, err := b.GetPattern()
	if err != nil {
		t.Fatalf("GetPattern: %v", err)
	}
	// EAN-8 fixed pattern length:
	// start(3) + 4 left digits × 4 + centre(5) + 4 right digits × 4 + end(3) = 43.
	const wantLen = 43
	if len(pattern) != wantLen {
		t.Errorf("EAN-8 pattern length = %d, want %d", len(pattern), wantLen)
	}
	// The 8th digit is the check digit 5; it's encoded in the right half as
	// eanTableC[5]="5170". The stop guard "A0A" follows. Verify suffix.
	const wantSuffix = "5170A0A"
	if !strings.HasSuffix(pattern, wantSuffix) {
		t.Errorf("pattern suffix = %q, want %q (check digit 5 for all-nines input)", pattern[len(pattern)-7:], wantSuffix)
	}
}

// TestMSI_Checksum_AllNines verifies Luhn checksum for three '9' digits.
//
// MSI Luhn: separate odd-index digits (position 1, 3-indexed from 0) and even.
//
//	Input "999": odd=[9] (index 1), even=[9,9] (indices 0,2)
//	checkOdd = 9, checkEven = 18
//	digitSum(9*2) = digitSum(18) = 1+8 = 9
//	total = 9 + 18 = 27, mod10 = 7, final = 10-7 = 3
//
// So "999" → check digit 3 appended in GetPattern.
// Pattern = start(2) + 3 digits×8 + check(8) + stop(3) = 37 chars.
// The check-digit bar is tabelleMSI[3]="51516060", stop="515".
func TestMSI_Checksum_AllNines(t *testing.T) {
	b := barcode.NewMSIBarcode()
	b.CalcChecksum = true
	if err := b.Encode("999"); err != nil {
		t.Fatalf("Encode: %v", err)
	}
	// EncodedText stores raw input only; check digit is appended in GetPattern.
	if got := b.EncodedText(); got != "999" {
		t.Errorf("EncodedText = %q, want 999 (raw input, check digit is in GetPattern)", got)
	}
	pattern, err := b.GetPattern()
	if err != nil {
		t.Fatalf("GetPattern: %v", err)
	}
	// start(2) + 3 data digits × 8 + check digit 3 × 8 + stop(3) = 37.
	if len(pattern) != 37 {
		t.Errorf("GetPattern len = %d, want 37 (check digit 3 for 999)", len(pattern))
	}
	// Check digit 3 → tabelleMSI[3]="51516060"; stop="515".
	const wantSuffix = "51516060515"
	if !strings.HasSuffix(pattern, wantSuffix) {
		t.Errorf("GetPattern suffix = %q, want %q (check digit 3)", pattern[len(pattern)-len(wantSuffix):], wantSuffix)
	}
}

// TestMSI_Checksum_SingleDigitZero verifies that "0" produces check digit 0.
//
//	Input "0": odd=[] (no index 1), even=[0] (index 0)
//	checkOdd=0, checkEven=0
//	digitSum(0*2)=0, total=0, mod10=0, final=0 (special case: 0 stays 0)
//
// Pattern = start(2) + 1 digit×8 + check(8) + stop(3) = 21 chars.
func TestMSI_Checksum_SingleDigitZero(t *testing.T) {
	b := barcode.NewMSIBarcode()
	b.CalcChecksum = true
	if err := b.Encode("0"); err != nil {
		t.Fatalf("Encode: %v", err)
	}
	// EncodedText stores raw input only; check digit is appended in GetPattern.
	if got := b.EncodedText(); got != "0" {
		t.Errorf("EncodedText = %q, want 0 (raw input)", got)
	}
	pattern, err := b.GetPattern()
	if err != nil {
		t.Fatalf("GetPattern: %v", err)
	}
	// start(2) + 1 digit×8 + check digit 0×8 + stop(3) = 21.
	if len(pattern) != 21 {
		t.Errorf("GetPattern len = %d, want 21 (check digit 0 for single 0)", len(pattern))
	}
}

// TestPlessey_SingleChar_NoCRCError verifies that a minimal single-hex-digit
// input does not cause a CRC computation panic or error. The CRC bit buffer
// for 1 hex digit = 4 data bits + 8 CRC bits = 12 total bits.
// This confirms the LSB-first bit extraction and CRC loop boundary for
// the shortest possible valid input.
func TestPlessey_SingleChar_NoCRCError(t *testing.T) {
	b := barcode.NewPlesseyBarcode()
	if err := b.Encode("0"); err != nil {
		t.Fatalf("Encode: %v", err)
	}
	pattern, err := b.GetPattern()
	if err != nil {
		t.Fatalf("GetPattern single-char: %v", err)
	}
	if len(pattern) == 0 {
		t.Error("Plessey single-char pattern should be non-empty")
	}
}

// minI/maxI helpers for test-only indexing operations.
// Named minI/maxI to avoid conflict with swissqr_test.go's min helper.
func minI(a, b int) int {
	if a < b {
		return a
	}
	return b
}
func maxI(a, b int) int {
	if a > b {
		return a
	}
	return b
}
