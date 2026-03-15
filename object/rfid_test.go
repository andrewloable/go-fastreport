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
}

func TestRFIDLabel_BoolFlags(t *testing.T) {
	r := object.NewRFIDLabel()
	r.UseAdjustForEPC = true
	r.RewriteEPCBank = true

	if !r.UseAdjustForEPC {
		t.Error("UseAdjustForEPC should be true")
	}
	if !r.RewriteEPCBank {
		t.Error("RewriteEPCBank should be true")
	}
}

func TestRFIDLabel_ImplementsBase(t *testing.T) {
	// Ensure RFIDLabel satisfies the report.Base interface (via ReportComponentBase).
	r := object.NewRFIDLabel()
	if r.TypeName() == "" {
		t.Error("TypeName should not be empty")
	}
}
