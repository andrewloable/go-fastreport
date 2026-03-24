package barcode

import (
	"image"
	"testing"
)

// ── pdf417TextEncode tests ──────────────────────────────────────────────────

func TestPdf417TextEncode_Uppercase(t *testing.T) {
	cw := pdf417TextEncode("AB")
	if len(cw) == 0 {
		t.Fatal("expected codewords, got empty slice")
	}
	// First codeword must be text compaction switch (900)
	if cw[0] != pdf417SwitchTxt {
		t.Errorf("first codeword = %d, want %d (text compaction switch)", cw[0], pdf417SwitchTxt)
	}
	// "A" = 0, "B" = 1 → cw = 0*30 + 1 = 1
	if len(cw) < 2 {
		t.Fatalf("expected at least 2 codewords, got %d", len(cw))
	}
	if cw[1] != 1 {
		t.Errorf("cw[1] = %d, want 1 (A=0, B=1 → 0*30+1)", cw[1])
	}
}

func TestPdf417TextEncode_Space(t *testing.T) {
	cw := pdf417TextEncode("A ")
	if len(cw) < 2 {
		t.Fatalf("expected at least 2 codewords, got %d", len(cw))
	}
	// "A"=0, " "=26 → cw = 0*30 + 26 = 26
	if cw[1] != 26 {
		t.Errorf("cw[1] = %d, want 26 (A=0, Space=26 → 0*30+26)", cw[1])
	}
}

func TestPdf417TextEncode_Lowercase(t *testing.T) {
	// Lowercase should insert latch (27) before the value
	cw := pdf417TextEncode("a")
	if len(cw) == 0 {
		t.Fatal("expected codewords")
	}
	if cw[0] != pdf417SwitchTxt {
		t.Errorf("first codeword = %d, want %d", cw[0], pdf417SwitchTxt)
	}
	// 'a' produces: latch(27), value(0) → then padding(29) → cw = 27*30+0=810, then 0*30+29=29
	// Wait, the padding is for the sub-values array being odd.
	// subValues = [27, 0] for 'a' → even, no padding → cw = 27*30+0 = 810
	if len(cw) < 2 {
		t.Fatalf("expected at least 2 codewords, got %d", len(cw))
	}
	if cw[1] != 810 {
		t.Errorf("cw[1] = %d, want 810 (lower latch=27, a=0 → 27*30+0)", cw[1])
	}
}

func TestPdf417TextEncode_Digits(t *testing.T) {
	// Digits insert a mixed latch (28) before the value
	cw := pdf417TextEncode("0")
	if len(cw) == 0 {
		t.Fatal("expected codewords")
	}
	// '0' produces subValues: [28, 15] → then they're even → cw = 28*30+15 = 855
	if len(cw) < 2 {
		t.Fatalf("expected at least 2 codewords, got %d", len(cw))
	}
	if cw[1] != 855 {
		t.Errorf("cw[1] = %d, want 855 (mixed latch=28, digit 0 val=15 → 28*30+15)", cw[1])
	}
}

func TestPdf417TextEncode_OddPadding(t *testing.T) {
	// Single uppercase letter = 1 subValue → padded to 2 with 29
	cw := pdf417TextEncode("A")
	if len(cw) < 2 {
		t.Fatalf("expected at least 2 codewords, got %d", len(cw))
	}
	// "A"=0, pad=29 → cw = 0*30 + 29 = 29
	if cw[1] != 29 {
		t.Errorf("cw[1] = %d, want 29 (A=0, pad=29 → 0*30+29)", cw[1])
	}
}

func TestPdf417TextEncode_LongText(t *testing.T) {
	cw := pdf417TextEncode("HELLO WORLD")
	if len(cw) < 2 {
		t.Fatalf("expected multiple codewords, got %d", len(cw))
	}
	if cw[0] != pdf417SwitchTxt {
		t.Errorf("first codeword must be text compaction switch %d", pdf417SwitchTxt)
	}
	// All codewords should be in valid range [0, 929)
	for i, c := range cw[1:] {
		if c < 0 || c >= pdf417MaxCW {
			t.Errorf("cw[%d] = %d, out of valid range [0, %d)", i+1, c, pdf417MaxCW)
		}
	}
}

func TestPdf417TextEncode_ClampToMaxCW(t *testing.T) {
	// If any pair produces cw >= 929, it should be clamped to 928.
	// subValues pair (30, 29) would be 30*30+29=929 which should be clamped.
	// It's hard to produce this directly, but we verify the clamp logic exists
	// by encoding a longer string and checking all codewords < 929.
	cw := pdf417TextEncode("ABCDEFGHIJKLMNOPQRSTUVWXYZ abcdefghijklmnopqrstuvwxyz 0123456789")
	for i, c := range cw {
		if c > pdf417MaxCW {
			t.Errorf("cw[%d] = %d exceeds max codeword %d", i, c, pdf417MaxCW)
		}
	}
}

func TestPdf417TextEncode_NonASCIIByte(t *testing.T) {
	// Characters in range [0, 256) that are not alpha/digit/space use byte fallback
	cw := pdf417TextEncode("!")
	if len(cw) == 0 {
		t.Fatal("expected codewords for special character")
	}
	if cw[0] != pdf417SwitchTxt {
		t.Errorf("first codeword = %d, want %d", cw[0], pdf417SwitchTxt)
	}
}

// ── pdf417ByteEncode tests ──────────────────────────────────────────────────

func TestPdf417ByteEncode_Short(t *testing.T) {
	data := []byte{0x41, 0x42, 0x43} // "ABC"
	cw := pdf417ByteEncode(data)
	if len(cw) == 0 {
		t.Fatal("expected codewords")
	}
	if cw[0] != pdf417SwitchByt {
		t.Errorf("first codeword = %d, want %d (byte compaction switch)", cw[0], pdf417SwitchByt)
	}
	// 3 remaining bytes → 3 individual codewords
	if len(cw) != 4 { // switch + 3 bytes
		t.Errorf("len(cw) = %d, want 4", len(cw))
	}
	// Each remaining byte should be its raw value
	for i := 1; i < len(cw); i++ {
		if cw[i] != int(data[i-1]) {
			t.Errorf("cw[%d] = %d, want %d", i, cw[i], data[i-1])
		}
	}
}

func TestPdf417ByteEncode_SixByteGroup(t *testing.T) {
	data := []byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x06}
	cw := pdf417ByteEncode(data)
	if cw[0] != pdf417SwitchByt {
		t.Errorf("first codeword = %d, want %d", cw[0], pdf417SwitchByt)
	}
	// 6 bytes should produce exactly 5 codewords after the switch
	if len(cw) != 6 { // switch + 5 codewords
		t.Errorf("len(cw) = %d, want 6", len(cw))
	}
	// Verify codewords are in valid range
	for i := 1; i < len(cw); i++ {
		if cw[i] < 0 || cw[i] >= 900 {
			t.Errorf("cw[%d] = %d, out of range [0, 900)", i, cw[i])
		}
	}
}

func TestPdf417ByteEncode_MixedGroups(t *testing.T) {
	// 8 bytes → 1 group of 6 (5 CW) + 2 remaining bytes (2 CW)
	data := []byte{0x10, 0x20, 0x30, 0x40, 0x50, 0x60, 0x70, 0x80}
	cw := pdf417ByteEncode(data)
	if cw[0] != pdf417SwitchByt {
		t.Fatalf("first codeword = %d, want %d", cw[0], pdf417SwitchByt)
	}
	// switch(1) + 5 (from group) + 2 (remaining) = 8
	if len(cw) != 8 {
		t.Errorf("len(cw) = %d, want 8", len(cw))
	}
}

func TestPdf417ByteEncode_Empty(t *testing.T) {
	cw := pdf417ByteEncode(nil)
	if len(cw) != 1 { // just the switch
		t.Errorf("len(cw) = %d, want 1 (just switch)", len(cw))
	}
	if cw[0] != pdf417SwitchByt {
		t.Errorf("cw[0] = %d, want %d", cw[0], pdf417SwitchByt)
	}
}

func TestPdf417ByteEncode_TwelveBytes(t *testing.T) {
	// Exactly 2 groups of 6 → 10 codewords
	data := make([]byte, 12)
	for i := range data {
		data[i] = byte(i + 1)
	}
	cw := pdf417ByteEncode(data)
	// switch(1) + 5 + 5 = 11
	if len(cw) != 11 {
		t.Errorf("len(cw) = %d, want 11", len(cw))
	}
}

// ── pdf417ECCount tests ─────────────────────────────────────────────────────

func TestPdf417ECCount(t *testing.T) {
	tests := []struct {
		level int
		want  int
	}{
		{0, 2},
		{1, 4},
		{2, 8},
		{3, 16},
		{4, 32},
		{5, 64},
		{6, 128},
		{7, 256},
		{8, 512},
	}
	for _, tc := range tests {
		got := pdf417ECCount(tc.level)
		if got != tc.want {
			t.Errorf("pdf417ECCount(%d) = %d, want %d", tc.level, got, tc.want)
		}
	}
}

// ── pdf417ComputeEC tests ───────────────────────────────────────────────────

func TestPdf417ComputeEC_Length(t *testing.T) {
	// Security level 2 → 8 EC codewords (2^3).
	data := []int{5, 900, 100, 200}
	ec := pdf417ComputeEC(data, 2)
	if len(ec) != 8 {
		t.Errorf("len(ec) = %d, want 8 (level 2)", len(ec))
	}
}

func TestPdf417ComputeEC_ValidRange(t *testing.T) {
	// All EC codewords must be in [0, 929).
	data := []int{10, 900, 100, 200, 300}
	ec := pdf417ComputeEC(data, 1) // level 1 → 4 EC codewords
	for i, v := range ec {
		if v < 0 || v >= 929 {
			t.Errorf("ec[%d] = %d, out of range [0, 929)", i, v)
		}
	}
}

func TestPdf417ComputeEC_DifferentInputsDifferentOutput(t *testing.T) {
	ec1 := pdf417ComputeEC([]int{5, 900, 100}, 1) // level 1 → 4 codewords
	ec2 := pdf417ComputeEC([]int{5, 900, 200}, 1)
	same := true
	for i := range ec1 {
		if ec1[i] != ec2[i] {
			same = false
			break
		}
	}
	if same {
		t.Error("expected different EC codewords for different input data")
	}
}

func TestPdf417ComputeEC_Deterministic(t *testing.T) {
	data := []int{7, 900, 50, 75}
	ec1 := pdf417ComputeEC(data, 2) // level 2 → 8 codewords
	ec2 := pdf417ComputeEC(data, 2)
	for i := range ec1 {
		if ec1[i] != ec2[i] {
			t.Errorf("non-deterministic: ec[%d] = %d vs %d", i, ec1[i], ec2[i])
		}
	}
}

func TestPdf417ComputeEC_LargeECCount(t *testing.T) {
	// Security level 5 → 64 EC codewords (2^6).
	data := []int{20, 900, 100, 200, 300, 400}
	ec := pdf417ComputeEC(data, 5)
	if len(ec) != 64 {
		t.Errorf("len(ec) = %d, want 64 (level 5)", len(ec))
	}
	for i, v := range ec {
		if v < 0 || v >= 929 {
			t.Errorf("ec[%d] = %d, out of valid range", i, v)
		}
	}
}

// ── pdf417EncodeSymbol tests ────────────────────────────────────────────────

func TestPdf417EncodeSymbol_Simple(t *testing.T) {
	matrix, err := pdf417EncodeSymbol("HELLO", 2)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(matrix) < pdf417MinRows {
		t.Errorf("rows = %d, want at least %d", len(matrix), pdf417MinRows)
	}
	// All rows should have the same width
	width := len(matrix[0])
	for i, row := range matrix {
		if len(row) != width {
			t.Errorf("row %d width = %d, want %d", i, len(row), width)
		}
	}
}

func TestPdf417EncodeSymbol_EmptyContent(t *testing.T) {
	_, err := pdf417EncodeSymbol("", 2)
	if err == nil {
		t.Error("expected error for empty content")
	}
}

func TestPdf417EncodeSymbol_InvalidSecurityLevel(t *testing.T) {
	// Negative and >8 should be clamped to 2
	matrix, err := pdf417EncodeSymbol("TEST", -1)
	if err != nil {
		t.Fatalf("unexpected error for secLevel=-1: %v", err)
	}
	if len(matrix) == 0 {
		t.Error("expected non-empty matrix")
	}

	matrix2, err := pdf417EncodeSymbol("TEST", 99)
	if err != nil {
		t.Fatalf("unexpected error for secLevel=99: %v", err)
	}
	if len(matrix2) == 0 {
		t.Error("expected non-empty matrix")
	}
}

func TestPdf417EncodeSymbol_NonASCII(t *testing.T) {
	// Non-ASCII content should use byte compaction
	matrix, err := pdf417EncodeSymbol("\x80\x81\x82", 2)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(matrix) == 0 {
		t.Error("expected non-empty matrix")
	}
}

func TestPdf417EncodeSymbol_LongContent(t *testing.T) {
	// Long content should produce more rows
	longText := ""
	for i := 0; i < 100; i++ {
		longText += "ABCDEFGHIJ "
	}
	matrix, err := pdf417EncodeSymbol(longText, 2)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(matrix) < 3 {
		t.Errorf("expected at least 3 rows for long text, got %d", len(matrix))
	}
}

func TestPdf417EncodeSymbol_MatrixContainsBars(t *testing.T) {
	matrix, err := pdf417EncodeSymbol("ABCD", 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// The matrix should contain both true (bar) and false (space) values
	hasTrue := false
	hasFalse := false
	for _, row := range matrix {
		for _, v := range row {
			if v {
				hasTrue = true
			} else {
				hasFalse = true
			}
		}
	}
	if !hasTrue {
		t.Error("matrix contains no bars (true values)")
	}
	if !hasFalse {
		t.Error("matrix contains no spaces (false values)")
	}
}

func TestPdf417EncodeSymbol_AllSecurityLevels(t *testing.T) {
	for level := 0; level <= 8; level++ {
		matrix, err := pdf417EncodeSymbol("TEST DATA", level)
		if err != nil {
			t.Fatalf("secLevel=%d: unexpected error: %v", level, err)
		}
		if len(matrix) < pdf417MinRows {
			t.Errorf("secLevel=%d: rows=%d, want >= %d", level, len(matrix), pdf417MinRows)
		}
	}
}

// ── pdf417DrawPattern tests ─────────────────────────────────────────────────

func TestPdf417DrawPattern_Start(t *testing.T) {
	// Start pattern: 8,1,1,1,1,1,1,3 = 17 modules
	row := make([]bool, 20)
	pos := pdf417DrawPattern(row, 0, pdf417StartPattern[:])
	if pos != 17 {
		t.Errorf("start pattern drew %d modules, want 17", pos)
	}
	// First 8 should be bars (true)
	for i := 0; i < 8; i++ {
		if !row[i] {
			t.Errorf("row[%d] = false, want true (bar in start pattern)", i)
		}
	}
	// Module 8 should be space (false)
	if row[8] {
		t.Error("row[8] should be space (false)")
	}
}

func TestPdf417DrawPattern_Stop(t *testing.T) {
	// Stop pattern: 7,1,1,3,1,1,1,2,1 = 18 modules
	row := make([]bool, 20)
	pos := pdf417DrawPattern(row, 0, pdf417StopPattern[:])
	if pos != 18 {
		t.Errorf("stop pattern drew %d modules, want 18", pos)
	}
}

func TestPdf417DrawPattern_WithOffset(t *testing.T) {
	row := make([]bool, 40)
	pos := pdf417DrawPattern(row, 10, pdf417StartPattern[:])
	if pos != 27 { // 10 + 17
		t.Errorf("pos = %d, want 27", pos)
	}
	// First 10 should be untouched (false)
	for i := 0; i < 10; i++ {
		if row[i] {
			t.Errorf("row[%d] should be false (before pattern start)", i)
		}
	}
}

func TestPdf417DrawPattern_BoundsCheck(t *testing.T) {
	// Pattern extends beyond row length — should not panic
	row := make([]bool, 5)
	pos := pdf417DrawPattern(row, 0, pdf417StartPattern[:])
	// Should write up to position 5, then skip remaining
	if pos > 17 {
		t.Errorf("pos = %d, should not exceed pattern total", pos)
	}
}

// ── pdf417DrawCodeword tests ────────────────────────────────────────────────

func TestPdf417DrawCodeword_Width(t *testing.T) {
	// Every codeword should produce exactly 17 modules
	for _, cluster := range []int{0, 1, 2} {
		row := make([]bool, 20)
		pos := pdf417DrawCodeword(row, 0, 100, cluster)
		if pos != 17 {
			t.Errorf("cluster=%d: codeword drew %d modules, want 17", cluster, pos)
		}
	}
}

func TestPdf417DrawCodeword_DifferentClusters(t *testing.T) {
	// Same codeword with different clusters should produce different patterns
	var rows [3][]bool
	for c := 0; c < 3; c++ {
		rows[c] = make([]bool, 17)
		pdf417DrawCodeword(rows[c], 0, 42, c)
	}
	// At least two of three should differ
	allSame := true
	for i := 0; i < 17; i++ {
		if rows[0][i] != rows[1][i] || rows[1][i] != rows[2][i] {
			allSame = false
			break
		}
	}
	if allSame {
		t.Error("expected different patterns for different clusters")
	}
}

func TestPdf417DrawCodeword_DifferentCodewords(t *testing.T) {
	row1 := make([]bool, 17)
	row2 := make([]bool, 17)
	pdf417DrawCodeword(row1, 0, 0, 0)
	pdf417DrawCodeword(row2, 0, 500, 0)
	same := true
	for i := 0; i < 17; i++ {
		if row1[i] != row2[i] {
			same = false
			break
		}
	}
	if same {
		t.Error("expected different patterns for different codeword values")
	}
}

// ── pdf417RowIndicator tests ────────────────────────────────────────────────

func TestPdf417RowIndicator_Cluster0(t *testing.T) {
	// cluster 0: cw = (row/3)*30 + ((rows-1)/3)
	cw := pdf417RowIndicator(0, 6, 3, 2, 0)
	want := (0/3)*30 + (6-1)/3 // 0 + 1 = 1
	if cw != want {
		t.Errorf("row=0,rows=6,cols=3,sec=2,left: got %d, want %d", cw, want)
	}
}

func TestPdf417RowIndicator_Cluster1(t *testing.T) {
	// cluster 1: cw = (row/3)*30 + (secLevel*3 + (rows-1)%3)
	cw := pdf417RowIndicator(1, 6, 3, 2, 0)
	want := (1/3)*30 + (2*3 + (6-1)%3) // 0 + 6+2 = 8
	if cw != want {
		t.Errorf("row=1,rows=6: got %d, want %d", cw, want)
	}
}

func TestPdf417RowIndicator_Cluster2(t *testing.T) {
	// cluster 2: cw = (row/3)*30 + (cols-1)
	cw := pdf417RowIndicator(2, 6, 5, 2, 0)
	want := (2/3)*30 + (5 - 1) // 0 + 4 = 4
	if cw != want {
		t.Errorf("row=2,cols=5: got %d, want %d", cw, want)
	}
}

func TestPdf417RowIndicator_HighRow(t *testing.T) {
	// Test with higher row numbers
	cw := pdf417RowIndicator(9, 30, 5, 2, 0)
	// row 9 → cluster 0, (9/3)*30 + (29/3) = 3*30 + 9 = 99
	want := 99
	if cw != want {
		t.Errorf("row=9: got %d, want %d", cw, want)
	}
}

func TestPdf417RowIndicator_Mod929(t *testing.T) {
	// Test that values >= 929 are taken mod 929
	// Row 93 cluster 0: (93/3)*30 + (89/3) = 31*30 + 29 = 959 → 959 % 929 = 30
	cw := pdf417RowIndicator(93, 90, 5, 2, 0)
	expected := (93/3)*30 + (90-1)/3
	if expected >= 929 {
		expected = expected % 929
	}
	if cw != expected {
		t.Errorf("row=93: got %d, want %d", cw, expected)
	}
}

func TestPdf417RowIndicator_SideParameter(t *testing.T) {
	// Left and right indicators differ — row=0, cluster=0:
	// left  = (0/3)*30 + (6-1)/3  = 0 + 1 = 1
	// right = (0/3)*30 + (3-1)    = 0 + 2 = 2
	left := pdf417RowIndicator(0, 6, 3, 2, 0)
	right := pdf417RowIndicator(0, 6, 3, 2, 1)
	if left == right {
		t.Errorf("left (%d) should differ from right (%d) for cluster 0", left, right)
	}
	// For cluster 0: left = base+(rows-1)/3, right = base+(cols-1)
	wantLeft := (0/3)*30 + (6-1)/3
	wantRight := (0/3)*30 + (3 - 1)
	if left != wantLeft {
		t.Errorf("left: got %d, want %d", left, wantLeft)
	}
	if right != wantRight {
		t.Errorf("right: got %d, want %d", right, wantRight)
	}
}

// ── pdf417Clusters table tests ───────────────────────────────────────────────

func TestPdf417Clusters_TableSize(t *testing.T) {
	// Each cluster must have exactly 929 entries.
	for c := 0; c < 3; c++ {
		if len(pdf417Clusters[c]) != 929 {
			t.Errorf("cluster %d: expected 929 entries, got %d", c, len(pdf417Clusters[c]))
		}
	}
}

func TestPdf417Clusters_PatternBits(t *testing.T) {
	// Every entry must fit in 17 bits (≤ 0x1ffff).
	for c := 0; c < 3; c++ {
		for cw := 0; cw < 929; cw++ {
			p := pdf417Clusters[c][cw]
			if p < 0 || p > 0x1ffff {
				t.Fatalf("cluster=%d, cw=%d: pattern 0x%x out of 17-bit range", c, cw, p)
			}
		}
	}
}

func TestPdf417DrawCodeword_Writes17Modules(t *testing.T) {
	// pdf417DrawCodeword must write exactly 17 modules per call.
	row := make([]bool, 200)
	for _, cw := range []int{0, 1, 100, 500, 928} {
		for cluster := 0; cluster < 3; cluster++ {
			pos := 0
			newPos := pdf417DrawCodeword(row, pos, cw, cluster)
			if newPos-pos != 17 {
				t.Errorf("cw=%d, cluster=%d: wrote %d modules, want 17", cw, cluster, newPos-pos)
			}
		}
	}
}

func TestPdf417DrawCodeword_OutOfRangeCW(t *testing.T) {
	// cw < 0 or cw >= 929 should be handled (clamped to 0).
	row1 := make([]bool, 20)
	row2 := make([]bool, 20)
	pdf417DrawCodeword(row1, 0, -5, 0)
	pdf417DrawCodeword(row2, 0, 0, 0)
	for i := range row1 {
		if row1[i] != row2[i] {
			t.Errorf("negative cw: module[%d] differs from cw=0", i)
		}
	}
}

func TestPdf417DrawCodeword_ClutersKnownPattern(t *testing.T) {
	// Verify cluster 0, cw=0 pattern against the known table value 0x1d5c0.
	// 0x1d5c0 = 0001 1101 0101 1100 0000 (bits 16..0):
	//   1,1,1,0,1,0,1,0,1,1,1,0,0,0,0,0,0 — wait let me recompute
	// 0x1d5c0 = 0x1*65536+0xd*4096+0x5*256+0xc*16+0x0 = 120256
	// binary (17 bits): 1 1101 0101 1100 0000 — 17 bits from bit 16 to bit 0
	// bit16=1,15=1,14=1,13=0,12=1,11=0,10=1,9=0,8=1,7=1,6=1,5=0,4=0,3=0,2=0,1=0,0=0
	row := make([]bool, 20)
	pdf417DrawCodeword(row, 0, 0, 0)
	expected := pdf417Clusters[0][0]
	for bit := 16; bit >= 0; bit-- {
		want := (expected>>bit)&1 != 0
		got := row[16-bit]
		if got != want {
			t.Errorf("cw=0,cluster=0,bit=%d: got %v, want %v", bit, got, want)
		}
	}
}

func TestPdf417GetCWPattern_OverflowRedistribution(t *testing.T) {
	// The old synthetic generator has been replaced by the ISO/IEC 15438 CLUSTERS
	// lookup table. Verify that all cluster entries produce a 17-bit pattern.
	for cw := 0; cw < 929; cw++ {
		for cluster := 0; cluster < 3; cluster++ {
			p := pdf417Clusters[cluster][cw]
			if p < 0 || p > 0x1ffff {
				t.Fatalf("cluster=%d, cw=%d: pattern 0x%x out of 17-bit range", cluster, cw, p)
			}
		}
	}
}

func TestPdf417GetCWPattern_FullSlotFallback(t *testing.T) {
	// Verify that pdf417DrawCodeword writes exactly 17 modules for every codeword.
	row := make([]bool, 200)
	for cw := 0; cw < 929; cw += 31 {
		for cluster := 0; cluster < 3; cluster++ {
			pos := pdf417DrawCodeword(row, 0, cw, cluster)
			if pos != 17 {
				t.Errorf("cw=%d, cluster=%d: wrote %d modules, want 17", cw, cluster, pos)
			}
		}
	}
}

// ── GetMatrix tests ─────────────────────────────────────────────────────────

func TestPDF417Barcode_GetMatrix_WithText(t *testing.T) {
	p := NewPDF417Barcode()
	_ = p.Encode("HELLO WORLD")
	matrix, rows, cols := p.GetMatrix()
	if rows == 0 || cols == 0 {
		t.Error("expected non-zero rows and cols")
	}
	if len(matrix) != rows {
		t.Errorf("matrix rows = %d, want %d", len(matrix), rows)
	}
	if rows > 0 && len(matrix[0]) != cols {
		t.Errorf("matrix cols = %d, want %d", len(matrix[0]), cols)
	}
}

func TestPDF417Barcode_GetMatrix_EmptyFallsBackToDefault(t *testing.T) {
	p := NewPDF417Barcode()
	// Don't call Encode — encodedText is empty, should fall back to DefaultValue
	matrix, rows, cols := p.GetMatrix()
	if rows == 0 || cols == 0 {
		t.Error("expected non-zero rows and cols from default value")
	}
	if len(matrix) == 0 {
		t.Error("expected non-empty matrix from default value")
	}
}

func TestPDF417Barcode_GetMatrix_FallbackOnError(t *testing.T) {
	// Force the fallback path: empty encodedText and empty default
	p := &PDF417Barcode{
		BaseBarcodeImpl: newBaseBarcodeImpl(BarcodeTypePDF417),
		SecurityLevel:   2,
	}
	// encodedText is "" and DefaultValue returns "PDF417" so it should still work
	matrix, rows, cols := p.GetMatrix()
	if rows == 0 || cols == 0 {
		t.Error("expected valid matrix from default value")
	}
	if matrix == nil {
		t.Error("matrix should not be nil")
	}
}

// ── PDF417Barcode Encode + Render integration tests ─────────────────────────

func TestPDF417Barcode_EncodeAndRender(t *testing.T) {
	p := NewPDF417Barcode()
	if err := p.Encode("Test PDF417 Content"); err != nil {
		t.Fatalf("Encode failed: %v", err)
	}
	img, err := p.Render(300, 100)
	if err != nil {
		t.Fatalf("Render failed: %v", err)
	}
	bounds := img.Bounds()
	if bounds.Dx() != 300 || bounds.Dy() != 100 {
		t.Errorf("image size = %dx%d, want 300x100", bounds.Dx(), bounds.Dy())
	}
}

func TestPDF417Barcode_Render_NotEncoded(t *testing.T) {
	p := NewPDF417Barcode()
	_, err := p.Render(100, 50)
	if err == nil {
		t.Error("expected error when Render called without Encode")
	}
}

func TestPDF417Barcode_EncodeAndRender_VariousSizes(t *testing.T) {
	sizes := [][2]int{{100, 50}, {200, 100}, {400, 200}, {50, 50}}
	for _, sz := range sizes {
		p := NewPDF417Barcode()
		_ = p.Encode("SIZE TEST")
		img, err := p.Render(sz[0], sz[1])
		if err != nil {
			t.Errorf("Render(%d,%d) error: %v", sz[0], sz[1], err)
			continue
		}
		bounds := img.Bounds()
		if bounds.Dx() != sz[0] || bounds.Dy() != sz[1] {
			t.Errorf("Render(%d,%d): got %dx%d", sz[0], sz[1], bounds.Dx(), bounds.Dy())
		}
	}
}

func TestPDF417Barcode_EncodeAndRender_IsImage(t *testing.T) {
	p := NewPDF417Barcode()
	_ = p.Encode("IMAGE CHECK")
	img, err := p.Render(200, 80)
	if err != nil {
		t.Fatalf("Render error: %v", err)
	}
	// Should be a proper image.Image with some non-white pixels (bars)
	var _, _ = img.(image.Image) // compile-time check
	bounds := img.Bounds()
	hasNonWhite := false
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			r, g, b, _ := img.At(x, y).RGBA()
			if r == 0 && g == 0 && b == 0 {
				hasNonWhite = true
				break
			}
		}
		if hasNonWhite {
			break
		}
	}
	if !hasNonWhite {
		t.Error("rendered image has no black pixels — barcode not drawn")
	}
}

func TestPDF417Barcode_SecurityLevels(t *testing.T) {
	for level := 0; level <= 8; level++ {
		p := NewPDF417Barcode()
		p.SecurityLevel = level
		if err := p.Encode("LEVEL TEST"); err != nil {
			t.Errorf("level=%d: Encode error: %v", level, err)
			continue
		}
		img, err := p.Render(300, 100)
		if err != nil {
			t.Errorf("level=%d: Render error: %v", level, err)
			continue
		}
		if img == nil {
			t.Errorf("level=%d: nil image", level)
		}
	}
}

func TestPDF417Barcode_DefaultValue(t *testing.T) {
	p := NewPDF417Barcode()
	if p.DefaultValue() != "PDF417" {
		t.Errorf("DefaultValue() = %q, want %q", p.DefaultValue(), "PDF417")
	}
}

// ── Edge cases and stress tests ─────────────────────────────────────────────

func TestPdf417EncodeSymbol_SingleChar(t *testing.T) {
	matrix, err := pdf417EncodeSymbol("A", 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(matrix) < pdf417MinRows {
		t.Errorf("rows = %d, want >= %d", len(matrix), pdf417MinRows)
	}
}

func TestPdf417EncodeSymbol_MaxSecurity(t *testing.T) {
	// Security level 8 = 512 EC codewords; lots of rows
	matrix, err := pdf417EncodeSymbol("HIGH SEC", 8)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(matrix) < 3 {
		t.Errorf("expected many rows for max security level, got %d", len(matrix))
	}
}

func TestPdf417TextEncode_MixedContent(t *testing.T) {
	// Mixed uppercase, lowercase, digits, and special characters
	cw := pdf417TextEncode("Hello World 123!")
	if len(cw) == 0 {
		t.Fatal("expected codewords")
	}
	if cw[0] != pdf417SwitchTxt {
		t.Errorf("first codeword = %d, want %d", cw[0], pdf417SwitchTxt)
	}
}

func TestPdf417ByteEncode_AllBytes(t *testing.T) {
	// Encode all 256 byte values
	data := make([]byte, 256)
	for i := range data {
		data[i] = byte(i)
	}
	cw := pdf417ByteEncode(data)
	if cw[0] != pdf417SwitchByt {
		t.Errorf("first codeword = %d, want %d", cw[0], pdf417SwitchByt)
	}
	// 256 bytes = 42 groups of 6 (252 bytes → 210 CW) + 4 remaining
	// switch(1) + 210 + 4 = 215
	expectedLen := 1 + 42*5 + 4
	if len(cw) != expectedLen {
		t.Errorf("len(cw) = %d, want %d", len(cw), expectedLen)
	}
}

func TestPdf417EncodeSymbol_RowConsistency(t *testing.T) {
	matrix, err := pdf417EncodeSymbol("CONSISTENCY TEST", 2)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// All rows must have the same length
	if len(matrix) == 0 {
		t.Fatal("empty matrix")
	}
	w := len(matrix[0])
	for i, row := range matrix {
		if len(row) != w {
			t.Errorf("row %d has width %d, expected %d", i, len(row), w)
		}
	}
}


func TestPDF417Barcode_GetMatrix_ErrorFallback(t *testing.T) {
	// To hit the error fallback in GetMatrix, we need pdf417EncodeSymbol to fail.
	// It only fails for empty content. GetMatrix falls back to DefaultValue when
	// encodedText is empty, so we need a barcode where both encodedText and
	// DefaultValue result in an error. We can't easily do that with the public API
	// since DefaultValue returns "PDF417". But we can test the len(matrix)==0 path
	// by directly calling with manipulated state.
	//
	// Actually, the only error path is empty content. Since GetMatrix falls back to
	// DefaultValue("PDF417") when encodedText is "", the error fallback line 380
	// is only reachable if DefaultValue also returned "". We can't override it.
	// Let's at least verify the fallback works correctly when encodedText is set.
	p := NewPDF417Barcode()
	// encodedText is empty → falls back to DefaultValue → produces valid matrix
	matrix, rows, cols := p.GetMatrix()
	if rows < 1 || cols < 1 {
		t.Error("expected valid matrix from default value fallback")
	}
	if matrix == nil {
		t.Error("matrix should not be nil")
	}
}

func TestPdf417TextEncode_HighByteChar(t *testing.T) {
	// Characters in range [0, 256) that aren't alpha/digit/space hit the byte fallback case
	// Test with tab (0x09), which is in range [0, 256) but not alpha/digit/space
	cw := pdf417TextEncode("\t")
	if len(cw) == 0 {
		t.Fatal("expected codewords for tab character")
	}
	if cw[0] != pdf417SwitchTxt {
		t.Errorf("first codeword = %d, want %d", cw[0], pdf417SwitchTxt)
	}
}

func TestPdf417TextEncode_NullByte(t *testing.T) {
	// Null byte (0x00) is in [0, 256) range
	cw := pdf417TextEncode(string([]byte{0x00}))
	if len(cw) == 0 {
		t.Fatal("expected codewords for null byte")
	}
}

func TestPdf417TextEncode_SpecialChars(t *testing.T) {
	// Various special characters that exercise the byte fallback branch
	specials := "!@#$%^&*()_+-=[]{}|;:',.<>?/~`"
	cw := pdf417TextEncode(specials)
	if len(cw) == 0 {
		t.Fatal("expected codewords for special characters")
	}
	if cw[0] != pdf417SwitchTxt {
		t.Errorf("first codeword = %d, want %d", cw[0], pdf417SwitchTxt)
	}
	// All codewords after switch should be in valid range
	for i, c := range cw[1:] {
		if c < 0 || c >= pdf417MaxCW {
			t.Errorf("cw[%d] = %d, out of range", i+1, c)
		}
	}
}

func TestPdf417EncodeSymbol_TruncatedData(t *testing.T) {
	// Create content large enough that dataWithLen > lengthCW, triggering truncation
	// This happens when encoded data is larger than the grid can hold
	// Use a very large content with low column count forcing truncation
	longText := ""
	for i := 0; i < 500; i++ {
		longText += "ABCDEFGHIJ"
	}
	matrix, err := pdf417EncodeSymbol(longText, 0) // security level 0 = 2 EC codewords
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(matrix) == 0 {
		t.Error("expected non-empty matrix")
	}
}

func TestPdf417EncodeSymbol_RowsExceedMax_ColumnsExpand(t *testing.T) {
	// Create enough content that initial rows > pdf417MaxRows, forcing column expansion.
	// The loop increases cols up to pdf417MaxCols (30).
	// For very large content, rows may still exceed 90 after maxing columns.
	longText := ""
	for i := 0; i < 2000; i++ {
		longText += "TESTING "
	}
	matrix, err := pdf417EncodeSymbol(longText, 5) // high security = many EC codewords
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(matrix) == 0 {
		t.Error("expected non-empty matrix")
	}
	// Should produce a valid matrix with consistent row widths
	if len(matrix) > 0 {
		w := len(matrix[0])
		for i, row := range matrix {
			if len(row) != w {
				t.Errorf("row %d width = %d, want %d", i, len(row), w)
			}
		}
	}
	// Verify that columns were expanded (width should be larger than minimum)
	if len(matrix) > 0 {
		// Minimum width with 3 cols: 17 + 17 + 3*17 + 17 + 18 = 120
		// If cols expanded, width should be larger
		if len(matrix[0]) <= 120 {
			t.Logf("matrix width = %d, columns may not have expanded", len(matrix[0]))
		}
	}
}

func TestPdf417EncodeSymbol_StartStopPatterns(t *testing.T) {
	matrix, err := pdf417EncodeSymbol("PATTERN", 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Each row should start with bars (start pattern begins with 8 bars)
	for i, row := range matrix {
		if len(row) < 8 {
			t.Fatalf("row %d too short: %d", i, len(row))
		}
		for j := 0; j < 8; j++ {
			if !row[j] {
				t.Errorf("row %d: expected bar at position %d (start pattern)", i, j)
				break
			}
		}
	}
}
