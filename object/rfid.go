package object

import "github.com/andrewloable/go-fastreport/report"

// ── RFIDBank ──────────────────────────────────────────────────────────────────

// RFIDBankFormat specifies how RFID bank data is encoded.
type RFIDBankFormat int

const (
	// RFIDBankFormatHex encodes data as hexadecimal.
	RFIDBankFormatHex RFIDBankFormat = iota
	// RFIDBankFormatASCII encodes data as ASCII text.
	RFIDBankFormatASCII
	// RFIDBankFormatDecimal encodes data as decimal digits.
	RFIDBankFormatDecimal
)

// RFIDBank represents one RFID memory bank (EPC, TID, or User).
// The Go equivalent of FastReport.RFIDLabel.RFIDBank.
type RFIDBank struct {
	// Data is the static data value written to this bank.
	Data string
	// DataColumn is the bound data-source column name (overrides Data if set).
	DataColumn string
	// Offset is the byte offset within the bank to start writing.
	Offset int
	// DataFormat controls the encoding of Data/DataColumn.
	DataFormat RFIDBankFormat
}

// ── LockType ──────────────────────────────────────────────────────────────────

// RFIDLockType specifies how an RFID bank is locked after writing.
type RFIDLockType int

const (
	// RFIDLockTypeUnlock leaves the bank unlocked (default).
	RFIDLockTypeUnlock RFIDLockType = iota
	// RFIDLockTypeLock locks the bank after writing.
	RFIDLockTypeLock
	// RFIDLockTypePermanentUnlock permanently unlocks the bank.
	RFIDLockTypePermanentUnlock
	// RFIDLockTypePermanentLock permanently locks the bank.
	RFIDLockTypePermanentLock
)

// ── ErrorHandle ───────────────────────────────────────────────────────────────

// RFIDErrorHandle controls behaviour when an RFID write fails during printing.
type RFIDErrorHandle int

const (
	// RFIDErrorHandleSkip skips the failed label and continues printing.
	RFIDErrorHandleSkip RFIDErrorHandle = iota
	// RFIDErrorHandlePause pauses the printer after a failed write.
	RFIDErrorHandlePause
	// RFIDErrorHandleError places the printer in error mode.
	RFIDErrorHandleError
)

// ── RFIDLabel ─────────────────────────────────────────────────────────────────

// RFIDLabel is a report object that encodes data into an RFID tag.
// It acts as a container (like ContainerObject) and is the Go equivalent
// of FastReport.RFIDLabel.
//
// During rendering, it produces a placeholder band with its name and
// encoded values for visualization; actual RFID encoding is performed by
// the printer driver at print time.
type RFIDLabel struct {
	report.ReportComponentBase

	// EPCBank is the Electronic Product Code memory bank.
	EPCBank RFIDBank
	// TIDBank is the Tag Identifier (read-only) bank — not usually written.
	TIDBank RFIDBank
	// UserBank is the User memory bank.
	UserBank RFIDBank

	// AccessPassword is the password required to access protected banks.
	AccessPassword string
	// AccessPasswordDataColumn binds AccessPassword to a data-source column.
	AccessPasswordDataColumn string

	// KillPassword is the password to permanently disable the tag.
	KillPassword string
	// KillPasswordDataColumn binds KillPassword to a data-source column.
	KillPasswordDataColumn string

	// Lock settings per bank.
	LockKillPassword   RFIDLockType
	LockAccessPassword RFIDLockType
	LockEPCBank        RFIDLockType
	LockUserBank       RFIDLockType

	// UseAdjustForEPC enables EPC length auto-adjustment.
	UseAdjustForEPC bool
	// RewriteEPCBank controls whether the EPC bank is overwritten on each print.
	RewriteEPCBank bool
	// ErrorHandle determines printer behaviour on write failure.
	ErrorHandle RFIDErrorHandle
}

// NewRFIDLabel creates an RFIDLabel with safe defaults.
func NewRFIDLabel() *RFIDLabel {
	return &RFIDLabel{
		ReportComponentBase: *report.NewReportComponentBase(),
		ErrorHandle:         RFIDErrorHandleSkip,
	}
}

// BaseName returns the base name prefix for auto-generated names.
func (r *RFIDLabel) BaseName() string { return "RFIDLabel" }

// TypeName returns "RFIDLabel".
func (r *RFIDLabel) TypeName() string { return "RFIDLabel" }

// Serialize writes RFIDLabel properties that differ from defaults.
func (r *RFIDLabel) Serialize(w report.Writer) error {
	if err := r.ReportComponentBase.Serialize(w); err != nil {
		return err
	}
	serializeBank(w, "EPC", r.EPCBank)
	serializeBank(w, "TID", r.TIDBank)
	serializeBank(w, "User", r.UserBank)

	if r.AccessPassword != "" {
		w.WriteStr("AccessPassword", r.AccessPassword)
	}
	if r.AccessPasswordDataColumn != "" {
		w.WriteStr("AccessPasswordDataColumn", r.AccessPasswordDataColumn)
	}
	if r.KillPassword != "" {
		w.WriteStr("KillPassword", r.KillPassword)
	}
	if r.KillPasswordDataColumn != "" {
		w.WriteStr("KillPasswordDataColumn", r.KillPasswordDataColumn)
	}
	if r.LockKillPassword != 0 {
		w.WriteInt("LockKillPassword", int(r.LockKillPassword))
	}
	if r.LockAccessPassword != 0 {
		w.WriteInt("LockAccessPassword", int(r.LockAccessPassword))
	}
	if r.LockEPCBank != 0 {
		w.WriteInt("LockEPCBank", int(r.LockEPCBank))
	}
	if r.LockUserBank != 0 {
		w.WriteInt("LockUserBank", int(r.LockUserBank))
	}
	if r.UseAdjustForEPC {
		w.WriteBool("UseAdjustForEPC", true)
	}
	if r.RewriteEPCBank {
		w.WriteBool("RewriteEPCBank", true)
	}
	if r.ErrorHandle != 0 {
		w.WriteInt("ErrorHandle", int(r.ErrorHandle))
	}
	return nil
}

func serializeBank(w report.Writer, prefix string, bank RFIDBank) {
	if bank.Data != "" {
		w.WriteStr(prefix+"Bank.Data", bank.Data)
	}
	if bank.DataColumn != "" {
		w.WriteStr(prefix+"Bank.DataColumn", bank.DataColumn)
	}
	if bank.Offset != 0 {
		w.WriteInt(prefix+"Bank.Offset", bank.Offset)
	}
	if bank.DataFormat != 0 {
		w.WriteInt(prefix+"Bank.DataFormat", int(bank.DataFormat))
	}
}

// Deserialize reads RFIDLabel properties.
func (r *RFIDLabel) Deserialize(rd report.Reader) error {
	if err := r.ReportComponentBase.Deserialize(rd); err != nil {
		return err
	}
	r.EPCBank = deserializeBank(rd, "EPC")
	r.TIDBank = deserializeBank(rd, "TID")
	r.UserBank = deserializeBank(rd, "User")

	r.AccessPassword = rd.ReadStr("AccessPassword", "")
	r.AccessPasswordDataColumn = rd.ReadStr("AccessPasswordDataColumn", "")
	r.KillPassword = rd.ReadStr("KillPassword", "")
	r.KillPasswordDataColumn = rd.ReadStr("KillPasswordDataColumn", "")
	r.LockKillPassword = RFIDLockType(rd.ReadInt("LockKillPassword", 0))
	r.LockAccessPassword = RFIDLockType(rd.ReadInt("LockAccessPassword", 0))
	r.LockEPCBank = RFIDLockType(rd.ReadInt("LockEPCBank", 0))
	r.LockUserBank = RFIDLockType(rd.ReadInt("LockUserBank", 0))
	r.UseAdjustForEPC = rd.ReadBool("UseAdjustForEPC", false)
	r.RewriteEPCBank = rd.ReadBool("RewriteEPCBank", false)
	r.ErrorHandle = RFIDErrorHandle(rd.ReadInt("ErrorHandle", 0))
	return nil
}

func deserializeBank(rd report.Reader, prefix string) RFIDBank {
	return RFIDBank{
		Data:       rd.ReadStr(prefix+"Bank.Data", ""),
		DataColumn: rd.ReadStr(prefix+"Bank.DataColumn", ""),
		Offset:     rd.ReadInt(prefix+"Bank.Offset", 0),
		DataFormat: RFIDBankFormat(rd.ReadInt(prefix+"Bank.DataFormat", 0)),
	}
}

// PlaceholderText returns a human-readable text describing the RFID tag
// configuration. Used by exporters to render a visual placeholder.
func (r *RFIDLabel) PlaceholderText() string {
	epc := r.EPCBank.Data
	if r.EPCBank.DataColumn != "" {
		epc = "[" + r.EPCBank.DataColumn + "]"
	}
	if epc == "" {
		epc = "–"
	}
	return "RFID: EPC=" + epc
}
