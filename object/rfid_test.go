package object_test

import (
	"strings"
	"testing"

	"github.com/andrewloable/go-fastreport/object"
)

func TestRFIDLabel_Defaults(t *testing.T) {
	r := object.NewRFIDLabel()
	if r.TypeName() != "RFIDLabel" {
		t.Errorf("TypeName = %q, want RFIDLabel", r.TypeName())
	}
	if r.BaseName() != "RFIDLabel" {
		t.Errorf("BaseName = %q, want RFIDLabel", r.BaseName())
	}
	if r.ErrorHandle != object.RFIDErrorHandleSkip {
		t.Errorf("default ErrorHandle = %v, want Skip", r.ErrorHandle)
	}
	// C# defaults (RFIDLabel.cs:517-533)
	if r.EpcFormat != "96,8,3,3,20,24,38" {
		t.Errorf("default EpcFormat = %q, want 96,8,3,3,20,24,38", r.EpcFormat)
	}
	if r.AccessPassword != "00000000" {
		t.Errorf("default AccessPassword = %q, want 00000000", r.AccessPassword)
	}
	if r.KillPassword != "00000000" {
		t.Errorf("default KillPassword = %q, want 00000000", r.KillPassword)
	}
	if r.LockKillPassword != object.RFIDLockTypePermanentUnlock {
		t.Errorf("default LockKillPassword = %v, want PermanentUnlock (Open)", r.LockKillPassword)
	}
	if r.LockAccessPassword != object.RFIDLockTypePermanentUnlock {
		t.Errorf("default LockAccessPassword = %v, want PermanentUnlock (Open)", r.LockAccessPassword)
	}
	if r.LockEPCBank != object.RFIDLockTypePermanentUnlock {
		t.Errorf("default LockEPCBank = %v, want PermanentUnlock (Open)", r.LockEPCBank)
	}
	if r.LockUserBank != object.RFIDLockTypePermanentUnlock {
		t.Errorf("default LockUserBank = %v, want PermanentUnlock (Open)", r.LockUserBank)
	}
	if r.ReadPower != 16 {
		t.Errorf("default ReadPower = %d, want 16", r.ReadPower)
	}
	if r.WritePower != 16 {
		t.Errorf("default WritePower = %d, want 16", r.WritePower)
	}
	if r.StartPermaLock != 0 {
		t.Errorf("default StartPermaLock = %d, want 0", r.StartPermaLock)
	}
	if r.CountPermaLock != 0 {
		t.Errorf("default CountPermaLock = %d, want 0", r.CountPermaLock)
	}
	if r.AdaptiveAntenna {
		t.Error("default AdaptiveAntenna should be false")
	}
}

func TestRFIDLabel_PlaceholderText_StaticEPC(t *testing.T) {
	r := object.NewRFIDLabel()
	r.EPCBank.Data = "AABBCCDD"
	text := r.PlaceholderText()
	if !strings.Contains(text, "AABBCCDD") {
		t.Errorf("PlaceholderText = %q, expected EPC data", text)
	}
}

func TestRFIDLabel_PlaceholderText_BoundEPC(t *testing.T) {
	r := object.NewRFIDLabel()
	r.EPCBank.DataColumn = "EpcColumn"
	text := r.PlaceholderText()
	if !strings.Contains(text, "EpcColumn") {
		t.Errorf("PlaceholderText = %q, expected column ref", text)
	}
}

func TestRFIDLabel_PlaceholderText_NoData(t *testing.T) {
	r := object.NewRFIDLabel()
	text := r.PlaceholderText()
	if !strings.Contains(text, "RFID") {
		t.Errorf("PlaceholderText = %q, expected RFID marker", text)
	}
}

func TestRFIDLabel_BankFields(t *testing.T) {
	r := object.NewRFIDLabel()
	r.EPCBank = object.RFIDBank{Data: "EPC123", DataColumn: "col", Offset: 2, DataFormat: object.RFIDBankFormatHex}
	r.UserBank = object.RFIDBank{Data: "USER1", DataFormat: object.RFIDBankFormatASCII}
	r.TIDBank = object.RFIDBank{DataColumn: "tid_col"}

	if r.EPCBank.Data != "EPC123" {
		t.Errorf("EPCBank.Data = %q", r.EPCBank.Data)
	}
	if r.EPCBank.Offset != 2 {
		t.Errorf("EPCBank.Offset = %d", r.EPCBank.Offset)
	}
	if r.UserBank.DataFormat != object.RFIDBankFormatASCII {
		t.Errorf("UserBank.DataFormat = %v", r.UserBank.DataFormat)
	}
	if r.TIDBank.DataColumn != "tid_col" {
		t.Errorf("TIDBank.DataColumn = %q", r.TIDBank.DataColumn)
	}
}

func TestRFIDLabel_LockTypes(t *testing.T) {
	r := object.NewRFIDLabel()
	r.LockEPCBank = object.RFIDLockTypePermanentLock
	r.LockUserBank = object.RFIDLockTypeLock
	r.LockAccessPassword = object.RFIDLockTypeUnlock
	r.LockKillPassword = object.RFIDLockTypePermanentUnlock

	if r.LockEPCBank != object.RFIDLockTypePermanentLock {
		t.Errorf("LockEPCBank = %v", r.LockEPCBank)
	}
	if r.LockUserBank != object.RFIDLockTypeLock {
		t.Errorf("LockUserBank = %v", r.LockUserBank)
	}
	if r.LockAccessPassword != object.RFIDLockTypeUnlock {
		t.Errorf("LockAccessPassword = %v", r.LockAccessPassword)
	}
	if r.LockKillPassword != object.RFIDLockTypePermanentUnlock {
		t.Errorf("LockKillPassword = %v", r.LockKillPassword)
	}
}

func TestRFIDLabel_ErrorHandleModes(t *testing.T) {
	for _, mode := range []object.RFIDErrorHandle{
		object.RFIDErrorHandleSkip,
		object.RFIDErrorHandlePause,
		object.RFIDErrorHandleError,
	} {
		r := object.NewRFIDLabel()
		r.ErrorHandle = mode
		if r.ErrorHandle != mode {
			t.Errorf("ErrorHandle = %v, want %v", r.ErrorHandle, mode)
		}
	}
}

func TestRFIDLabel_Passwords(t *testing.T) {
	r := object.NewRFIDLabel()
	r.AccessPassword = "A1B2"
	r.KillPassword = "DEAD"
	r.AccessPasswordDataColumn = "accCol"
	r.KillPasswordDataColumn = "killCol"

	if r.AccessPassword != "A1B2" {
		t.Errorf("AccessPassword = %q", r.AccessPassword)
	}
	if r.KillPassword != "DEAD" {
		t.Errorf("KillPassword = %q", r.KillPassword)
	}
	if r.AccessPasswordDataColumn != "accCol" {
		t.Errorf("AccessPasswordDataColumn = %q", r.AccessPasswordDataColumn)
	}
	if r.KillPasswordDataColumn != "killCol" {
		t.Errorf("KillPasswordDataColumn = %q", r.KillPasswordDataColumn)
	}
}

// TestRFIDLabel_SetUseAdjustForEPC verifies the mutual-exclusion logic.
// C# RFIDLabel.UseAdjustForEPC setter (RFIDLabel.cs:356-361)
func TestRFIDLabel_SetUseAdjustForEPC(t *testing.T) {
	r := object.NewRFIDLabel()
	r.RewriteEPCBank = true
	r.SetUseAdjustForEPC(true)
	if !r.UseAdjustForEPC {
		t.Error("UseAdjustForEPC should be true after SetUseAdjustForEPC(true)")
	}
	if r.RewriteEPCBank {
		t.Error("RewriteEPCBank should be cleared when UseAdjustForEPC is set true")
	}

	// Setting false should not affect RewriteEPCBank.
	r.RewriteEPCBank = true
	r.SetUseAdjustForEPC(false)
	if r.UseAdjustForEPC {
		t.Error("UseAdjustForEPC should be false after SetUseAdjustForEPC(false)")
	}
	if !r.RewriteEPCBank {
		t.Error("RewriteEPCBank should remain true when UseAdjustForEPC is set false")
	}
}

// TestRFIDLabel_SetRewriteEPCBank verifies the mutual-exclusion logic.
// C# RFIDLabel.RewriteEPCbank setter (RFIDLabel.cs:372-377)
func TestRFIDLabel_SetRewriteEPCBank(t *testing.T) {
	r := object.NewRFIDLabel()
	r.UseAdjustForEPC = true
	r.SetRewriteEPCBank(true)
	if !r.RewriteEPCBank {
		t.Error("RewriteEPCBank should be true after SetRewriteEPCBank(true)")
	}
	if r.UseAdjustForEPC {
		t.Error("UseAdjustForEPC should be cleared when RewriteEPCBank is set true")
	}

	// Setting false should not affect UseAdjustForEPC.
	r.UseAdjustForEPC = true
	r.SetRewriteEPCBank(false)
	if r.RewriteEPCBank {
		t.Error("RewriteEPCBank should be false after SetRewriteEPCBank(false)")
	}
	if !r.UseAdjustForEPC {
		t.Error("UseAdjustForEPC should remain true when RewriteEPCBank is set false")
	}
}

// TestRFIDLabel_BoolFlags ensures simple bool fields are stored correctly.
func TestRFIDLabel_BoolFlags(t *testing.T) {
	r := object.NewRFIDLabel()
	r.AdaptiveAntenna = true
	if !r.AdaptiveAntenna {
		t.Error("AdaptiveAntenna should be true")
	}
}

// TestRFIDLabel_PowerLevels verifies ReadPower and WritePower are stored.
func TestRFIDLabel_PowerLevels(t *testing.T) {
	r := object.NewRFIDLabel()
	r.ReadPower = 30
	r.WritePower = 20
	if r.ReadPower != 30 {
		t.Errorf("ReadPower = %d, want 30", r.ReadPower)
	}
	if r.WritePower != 20 {
		t.Errorf("WritePower = %d, want 20", r.WritePower)
	}
}

// TestRFIDLabel_PermaLock verifies StartPermaLock and CountPermaLock fields.
func TestRFIDLabel_PermaLock(t *testing.T) {
	r := object.NewRFIDLabel()
	r.StartPermaLock = 4
	r.CountPermaLock = 8
	if r.StartPermaLock != 4 {
		t.Errorf("StartPermaLock = %d, want 4", r.StartPermaLock)
	}
	if r.CountPermaLock != 8 {
		t.Errorf("CountPermaLock = %d, want 8", r.CountPermaLock)
	}
}

// TestRFIDLabel_EpcFormat verifies the EpcFormat field is stored and defaults correctly.
func TestRFIDLabel_EpcFormat(t *testing.T) {
	r := object.NewRFIDLabel()
	if r.EpcFormat != "96,8,3,3,20,24,38" {
		t.Errorf("EpcFormat default = %q", r.EpcFormat)
	}
	r.EpcFormat = "64,8,3,3,20,10,20"
	if r.EpcFormat != "64,8,3,3,20,10,20" {
		t.Errorf("EpcFormat = %q", r.EpcFormat)
	}
}

// TestRFIDLabel_Assign verifies that Assign copies all properties.
// C# RFIDLabel.Assign (RFIDLabel.cs:430-454)
func TestRFIDLabel_Assign(t *testing.T) {
	src := object.NewRFIDLabel()
	src.EPCBank = object.RFIDBank{Data: "AABB", DataColumn: "epcCol", Offset: 1, DataFormat: object.RFIDBankFormatASCII}
	src.TIDBank = object.RFIDBank{Data: "CCDD"}
	src.UserBank = object.RFIDBank{DataColumn: "userCol"}
	src.EpcFormat = "64,8,3,3,20,10,20"
	src.AccessPassword = "A1A1A1A1"
	src.KillPassword = "B2B2B2B2"
	src.AccessPasswordDataColumn = "accCol"
	src.KillPasswordDataColumn = "killCol"
	src.LockAccessPassword = object.RFIDLockTypeLock
	src.LockKillPassword = object.RFIDLockTypeUnlock
	src.LockEPCBank = object.RFIDLockTypePermanentLock
	src.LockUserBank = object.RFIDLockTypePermanentUnlock
	src.StartPermaLock = 2
	src.CountPermaLock = 4
	src.AdaptiveAntenna = true
	src.ReadPower = 20
	src.WritePower = 18
	src.UseAdjustForEPC = true
	src.RewriteEPCBank = false
	src.ErrorHandle = object.RFIDErrorHandlePause

	dst := object.NewRFIDLabel()
	dst.Assign(src)

	if dst.EPCBank.Data != "AABB" {
		t.Errorf("Assign EPCBank.Data = %q", dst.EPCBank.Data)
	}
	if dst.EPCBank.DataColumn != "epcCol" {
		t.Errorf("Assign EPCBank.DataColumn = %q", dst.EPCBank.DataColumn)
	}
	if dst.TIDBank.Data != "CCDD" {
		t.Errorf("Assign TIDBank.Data = %q", dst.TIDBank.Data)
	}
	if dst.UserBank.DataColumn != "userCol" {
		t.Errorf("Assign UserBank.DataColumn = %q", dst.UserBank.DataColumn)
	}
	if dst.EpcFormat != "64,8,3,3,20,10,20" {
		t.Errorf("Assign EpcFormat = %q", dst.EpcFormat)
	}
	if dst.AccessPassword != "A1A1A1A1" {
		t.Errorf("Assign AccessPassword = %q", dst.AccessPassword)
	}
	if dst.KillPassword != "B2B2B2B2" {
		t.Errorf("Assign KillPassword = %q", dst.KillPassword)
	}
	if dst.AccessPasswordDataColumn != "accCol" {
		t.Errorf("Assign AccessPasswordDataColumn = %q", dst.AccessPasswordDataColumn)
	}
	if dst.KillPasswordDataColumn != "killCol" {
		t.Errorf("Assign KillPasswordDataColumn = %q", dst.KillPasswordDataColumn)
	}
	if dst.LockAccessPassword != object.RFIDLockTypeLock {
		t.Errorf("Assign LockAccessPassword = %v", dst.LockAccessPassword)
	}
	if dst.LockKillPassword != object.RFIDLockTypeUnlock {
		t.Errorf("Assign LockKillPassword = %v", dst.LockKillPassword)
	}
	if dst.LockEPCBank != object.RFIDLockTypePermanentLock {
		t.Errorf("Assign LockEPCBank = %v", dst.LockEPCBank)
	}
	if dst.LockUserBank != object.RFIDLockTypePermanentUnlock {
		t.Errorf("Assign LockUserBank = %v", dst.LockUserBank)
	}
	if dst.StartPermaLock != 2 {
		t.Errorf("Assign StartPermaLock = %d", dst.StartPermaLock)
	}
	if dst.CountPermaLock != 4 {
		t.Errorf("Assign CountPermaLock = %d", dst.CountPermaLock)
	}
	if !dst.AdaptiveAntenna {
		t.Error("Assign AdaptiveAntenna should be true")
	}
	if dst.ReadPower != 20 {
		t.Errorf("Assign ReadPower = %d", dst.ReadPower)
	}
	if dst.WritePower != 18 {
		t.Errorf("Assign WritePower = %d", dst.WritePower)
	}
	if !dst.UseAdjustForEPC {
		t.Error("Assign UseAdjustForEPC should be true")
	}
	if dst.RewriteEPCBank {
		t.Error("Assign RewriteEPCBank should be false")
	}
	if dst.ErrorHandle != object.RFIDErrorHandlePause {
		t.Errorf("Assign ErrorHandle = %v", dst.ErrorHandle)
	}
}

// TestRFIDBank_CountByte verifies byte count calculation.
// C# RFIDBank.CountByte (RFIDLabel.cs:595-603)
func TestRFIDBank_CountByte(t *testing.T) {
	tests := []struct {
		name   string
		bank   object.RFIDBank
		want   int
	}{
		{"ascii 4 chars", object.RFIDBank{Data: "ABCD", DataFormat: object.RFIDBankFormatASCII}, 4},
		{"ascii 0 chars", object.RFIDBank{Data: "", DataFormat: object.RFIDBankFormatASCII}, 0},
		{"hex even", object.RFIDBank{Data: "AABB", DataFormat: object.RFIDBankFormatHex}, 2},
		{"hex odd", object.RFIDBank{Data: "AAB", DataFormat: object.RFIDBankFormatHex}, 2},
		{"hex 1 char", object.RFIDBank{Data: "A", DataFormat: object.RFIDBankFormatHex}, 1},
		{"hex empty", object.RFIDBank{Data: "", DataFormat: object.RFIDBankFormatHex}, 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.bank.CountByte()
			if got != tt.want {
				t.Errorf("CountByte = %d, want %d", got, tt.want)
			}
		})
	}
}

func TestRFIDLabel_ImplementsBase(t *testing.T) {
	// Ensure RFIDLabel satisfies the report.Base interface (via ReportComponentBase).
	r := object.NewRFIDLabel()
	if r.TypeName() == "" {
		t.Error("TypeName should not be empty")
	}
}
