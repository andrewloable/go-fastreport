package object

import "github.com/andrewloable/go-fastreport/report"

// ── RFIDBank ──────────────────────────────────────────────────────────────────

// RFIDBankFormat specifies how RFID bank data is encoded.
// C# RFIDLabel.RFIDBank.Format (RFIDLabel.cs:551-554)
type RFIDBankFormat int

const (
	// RFIDBankFormatHex encodes data as hexadecimal ('H').
	RFIDBankFormatHex RFIDBankFormat = iota
	// RFIDBankFormatASCII encodes data as ASCII text ('A').
	RFIDBankFormatASCII
	// RFIDBankFormatDecimal encodes data as decimal ('D').
	// C# RFIDLabel.RFIDBank.Format (RFIDLabel.cs:554)
	RFIDBankFormatDecimal
)

// RFIDBank represents one RFID memory bank (EPC, TID, or User).
// The Go equivalent of FastReport.RFIDLabel.RFIDBank (RFIDLabel.cs:540-639).
type RFIDBank struct {
	// Data is the static data value written to this bank.
	Data string
	// DataColumn is the bound data-source column name (overrides Data if set).
	DataColumn string
	// Offset is the offset within the bank to start writing (measured in 16-bit blocks).
	// C# RFIDBank.Offset
	Offset int
	// DataFormat controls the encoding of Data/DataColumn.
	DataFormat RFIDBankFormat
}

// CountByte returns the number of bytes used by the bank data.
// C# RFIDBank.CountByte (RFIDLabel.cs:595-603)
func (b *RFIDBank) CountByte() int {
	if b.DataFormat == RFIDBankFormatASCII {
		return len(b.Data)
	}
	// Hex: each pair of hex digits = 1 byte, ceiling division
	l := len(b.Data)
	return (l + 1) / 2
}

// ── LockType ──────────────────────────────────────────────────────────────────

// RFIDLockType specifies how an RFID bank is locked after writing.
// C# RFIDLabel.LockType (RFIDLabel.cs:18-39)
type RFIDLockType int

const (
	// RFIDLockTypeUnlock leaves the bank unlocked ('U').
	RFIDLockTypeUnlock RFIDLockType = iota
	// RFIDLockTypeLock locks the bank after writing ('L').
	RFIDLockTypeLock
	// RFIDLockTypePermanentUnlock permanently unlocks the bank ('O' = Open).
	RFIDLockTypePermanentUnlock
	// RFIDLockTypePermanentLock permanently locks the bank ('P' = Protect).
	RFIDLockTypePermanentLock
)

// ── ErrorHandle ───────────────────────────────────────────────────────────────

// RFIDErrorHandle controls behaviour when an RFID write fails during printing.
// C# RFIDLabel.EErrorHandle (RFIDLabel.cs:44-60)
type RFIDErrorHandle int

const (
	// RFIDErrorHandleSkip skips the failed label and continues printing ('N' = None).
	RFIDErrorHandleSkip RFIDErrorHandle = iota
	// RFIDErrorHandlePause pauses the printer after a failed write ('P').
	RFIDErrorHandlePause
	// RFIDErrorHandleError places the printer in error mode ('E').
	RFIDErrorHandleError
)

// ── RFIDLabel ─────────────────────────────────────────────────────────────────

// RFIDLabel is a report object that encodes data into an RFID tag.
// It acts as a container (like ContainerObject) and is the Go equivalent
// of FastReport.RFIDLabel (RFIDLabel.cs:13-640).
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

	// EpcFormat defines the EPC bit-field layout (default "96,8,3,3,20,24,38").
	// C# RFIDLabel.EpcFormat (RFIDLabel.cs:153-164)
	EpcFormat string

	// AccessPassword is the password required to access protected banks (default "00000000").
	AccessPassword string
	// AccessPasswordDataColumn binds AccessPassword to a data-source column.
	AccessPasswordDataColumn string

	// KillPassword is the password to permanently disable the tag (default "00000000").
	KillPassword string
	// KillPasswordDataColumn binds KillPassword to a data-source column.
	KillPasswordDataColumn string

	// Lock settings per bank.
	// C# defaults: all Open (= PermanentUnlock), RFIDLabel.cs:522-525
	LockKillPassword   RFIDLockType
	LockAccessPassword RFIDLockType
	LockEPCBank        RFIDLockType
	LockUserBank       RFIDLockType

	// StartPermaLock is the start section for permanent lock of user bank.
	// C# RFIDLabel.StartPermaLock (RFIDLabel.cs:289-299)
	StartPermaLock int
	// CountPermaLock is the count of sections for permanent lock of user bank.
	// C# RFIDLabel.CountPermaLock (RFIDLabel.cs:301-313)
	CountPermaLock int

	// AdaptiveAntenna enables the adaptive antenna property.
	// C# RFIDLabel.AdaptiveAntenna (RFIDLabel.cs:398-402)
	AdaptiveAntenna bool

	// ReadPower is the read power level for the label (default 16).
	// C# RFIDLabel.ReadPower (RFIDLabel.cs:319-329)
	ReadPower int16
	// WritePower is the write power level for the label (default 16).
	// C# RFIDLabel.WritePower (RFIDLabel.cs:331-344)
	WritePower int16

	// UseAdjustForEPC enables EPC length auto-adjustment.
	// Setting this true clears RewriteEPCBank (C# RFIDLabel.cs:358-361).
	UseAdjustForEPC bool
	// RewriteEPCBank controls whether the entire EPC bank is overwritten on each print.
	// Setting this true clears UseAdjustForEPC (C# RFIDLabel.cs:373-377).
	RewriteEPCBank bool

	// ErrorHandle determines printer behaviour on write failure.
	ErrorHandle RFIDErrorHandle
}

// NewRFIDLabel creates an RFIDLabel with C#-matching defaults.
// C# RFIDLabel constructor (RFIDLabel.cs:512-535)
func NewRFIDLabel() *RFIDLabel {
	return &RFIDLabel{
		ReportComponentBase:      *report.NewReportComponentBase(),
		EpcFormat:                "96,8,3,3,20,24,38",
		AccessPassword:           "00000000",
		KillPassword:             "00000000",
		AccessPasswordDataColumn: "",
		KillPasswordDataColumn:   "",
		LockKillPassword:         RFIDLockTypePermanentUnlock, // C# LockType.Open
		LockAccessPassword:       RFIDLockTypePermanentUnlock, // C# LockType.Open
		LockEPCBank:              RFIDLockTypePermanentUnlock, // C# LockType.Open
		LockUserBank:             RFIDLockTypePermanentUnlock, // C# LockType.Open
		StartPermaLock:           0,
		CountPermaLock:           0,
		AdaptiveAntenna:          false,
		ReadPower:                16,
		WritePower:               16,
		UseAdjustForEPC:          false,
		RewriteEPCBank:           false,
		ErrorHandle:              RFIDErrorHandleSkip,
	}
}

// SetUseAdjustForEPC sets UseAdjustForEPC; when set to true it clears RewriteEPCBank.
// C# RFIDLabel.UseAdjustForEPC setter (RFIDLabel.cs:356-361)
func (r *RFIDLabel) SetUseAdjustForEPC(v bool) {
	r.UseAdjustForEPC = v
	if v {
		r.RewriteEPCBank = false
	}
}

// SetRewriteEPCBank sets RewriteEPCBank; when set to true it clears UseAdjustForEPC.
// C# RFIDLabel.RewriteEPCbank setter (RFIDLabel.cs:372-377)
func (r *RFIDLabel) SetRewriteEPCBank(v bool) {
	r.RewriteEPCBank = v
	if v {
		r.UseAdjustForEPC = false
	}
}

// BaseName returns the base name prefix for auto-generated names.
func (r *RFIDLabel) BaseName() string { return "RFIDLabel" }

// TypeName returns "RFIDLabel".
func (r *RFIDLabel) TypeName() string { return "RFIDLabel" }

// Assign copies the contents of another RFIDLabel into this one.
// C# RFIDLabel.Assign (RFIDLabel.cs:430-454)
func (r *RFIDLabel) Assign(src *RFIDLabel) {
	r.TIDBank = src.TIDBank
	r.EPCBank = src.EPCBank
	r.UserBank = src.UserBank
	r.EpcFormat = src.EpcFormat
	r.AccessPassword = src.AccessPassword
	r.KillPassword = src.KillPassword
	r.AccessPasswordDataColumn = src.AccessPasswordDataColumn
	r.KillPasswordDataColumn = src.KillPasswordDataColumn
	r.LockAccessPassword = src.LockAccessPassword
	r.LockKillPassword = src.LockKillPassword
	r.LockEPCBank = src.LockEPCBank
	r.LockUserBank = src.LockUserBank
	r.StartPermaLock = src.StartPermaLock
	r.CountPermaLock = src.CountPermaLock
	r.AdaptiveAntenna = src.AdaptiveAntenna
	r.ReadPower = src.ReadPower
	r.WritePower = src.WritePower
	r.UseAdjustForEPC = src.UseAdjustForEPC
	r.RewriteEPCBank = src.RewriteEPCBank
	r.ErrorHandle = src.ErrorHandle
}

// Serialize writes RFIDLabel properties that differ from defaults.
// C# RFIDLabel.Serialize (RFIDLabel.cs:457-501)
func (r *RFIDLabel) Serialize(w report.Writer) error {
	if err := r.ReportComponentBase.Serialize(w); err != nil {
		return err
	}
	def := NewRFIDLabel()

	serializeBank(w, "EpcBank", r.EPCBank, def.EPCBank)
	serializeBank(w, "TidBank", r.TIDBank, def.TIDBank)
	serializeBank(w, "UserBank", r.UserBank, def.UserBank)

	if r.EpcFormat != def.EpcFormat {
		w.WriteStr("EpcFormat", r.EpcFormat)
	}
	if r.AccessPassword != def.AccessPassword {
		w.WriteStr("AccessPassword", r.AccessPassword)
	}
	if r.KillPassword != def.KillPassword {
		w.WriteStr("KillPassword", r.KillPassword)
	}
	if r.AccessPasswordDataColumn != def.AccessPasswordDataColumn {
		w.WriteStr("AccessPasswordDataColumn", r.AccessPasswordDataColumn)
	}
	if r.KillPasswordDataColumn != def.KillPasswordDataColumn {
		w.WriteStr("KillPasswordDataColumn", r.KillPasswordDataColumn)
	}
	if r.LockAccessPassword != def.LockAccessPassword {
		w.WriteInt("LockAccessPassword", int(r.LockAccessPassword))
	}
	if r.LockKillPassword != def.LockKillPassword {
		w.WriteInt("LockKillPassword", int(r.LockKillPassword))
	}
	if r.LockEPCBank != def.LockEPCBank {
		// C# key: "LockEPCBlock" (RFIDLabel.cs:482)
		w.WriteInt("LockEPCBlock", int(r.LockEPCBank))
	}
	if r.LockUserBank != def.LockUserBank {
		// C# key: "LockUserBlock" (RFIDLabel.cs:484)
		w.WriteInt("LockUserBlock", int(r.LockUserBank))
	}
	if r.StartPermaLock != def.StartPermaLock {
		w.WriteInt("StartPermaLock", r.StartPermaLock)
	}
	if r.CountPermaLock != def.CountPermaLock {
		w.WriteInt("CountPermaLock", r.CountPermaLock)
	}
	if r.AdaptiveAntenna != def.AdaptiveAntenna {
		w.WriteBool("AdaptiveAntenna", r.AdaptiveAntenna)
	}
	if r.ReadPower != def.ReadPower {
		// C# key: "PowerRead" (RFIDLabel.cs:493)
		w.WriteInt("PowerRead", int(r.ReadPower))
	}
	if r.WritePower != def.WritePower {
		// C# key: "PowerWrite" (RFIDLabel.cs:495)
		w.WriteInt("PowerWrite", int(r.WritePower))
	}
	if r.UseAdjustForEPC != def.UseAdjustForEPC {
		w.WriteBool("UseAdjustForEPC", r.UseAdjustForEPC)
	}
	if r.RewriteEPCBank != def.RewriteEPCBank {
		w.WriteBool("RewriteEPCbank", r.RewriteEPCBank)
	}
	if r.ErrorHandle != def.ErrorHandle {
		w.WriteInt("ErrorHandle", int(r.ErrorHandle))
	}
	return nil
}

// serializeBank writes bank properties that differ from their defaults.
// C# RFIDBank.Serialize (RFIDLabel.cs:617-627)
func serializeBank(w report.Writer, prefix string, bank, def RFIDBank) {
	if bank.Data != def.Data {
		w.WriteStr(prefix+".Data", bank.Data)
	}
	if bank.DataColumn != def.DataColumn {
		w.WriteStr(prefix+".DataColumn", bank.DataColumn)
	}
	if bank.DataFormat != def.DataFormat {
		w.WriteInt(prefix+".DataFormat", int(bank.DataFormat))
	}
	if bank.Offset != def.Offset {
		w.WriteInt(prefix+".Offset", bank.Offset)
	}
}

// Deserialize reads RFIDLabel properties.
// C# attribute names mirror those written in Serialize above.
func (r *RFIDLabel) Deserialize(rd report.Reader) error {
	if err := r.ReportComponentBase.Deserialize(rd); err != nil {
		return err
	}
	def := NewRFIDLabel()

	r.EPCBank = deserializeBank(rd, "EpcBank", def.EPCBank)
	r.TIDBank = deserializeBank(rd, "TidBank", def.TIDBank)
	r.UserBank = deserializeBank(rd, "UserBank", def.UserBank)

	r.EpcFormat = rd.ReadStr("EpcFormat", def.EpcFormat)
	r.AccessPassword = rd.ReadStr("AccessPassword", def.AccessPassword)
	r.AccessPasswordDataColumn = rd.ReadStr("AccessPasswordDataColumn", def.AccessPasswordDataColumn)
	r.KillPassword = rd.ReadStr("KillPassword", def.KillPassword)
	r.KillPasswordDataColumn = rd.ReadStr("KillPasswordDataColumn", def.KillPasswordDataColumn)
	r.LockKillPassword = RFIDLockType(rd.ReadInt("LockKillPassword", int(def.LockKillPassword)))
	r.LockAccessPassword = RFIDLockType(rd.ReadInt("LockAccessPassword", int(def.LockAccessPassword)))
	// C# serializes as "LockEPCBlock" / "LockUserBlock" (RFIDLabel.cs:482-484)
	r.LockEPCBank = RFIDLockType(rd.ReadInt("LockEPCBlock", int(def.LockEPCBank)))
	r.LockUserBank = RFIDLockType(rd.ReadInt("LockUserBlock", int(def.LockUserBank)))
	r.StartPermaLock = rd.ReadInt("StartPermaLock", def.StartPermaLock)
	r.CountPermaLock = rd.ReadInt("CountPermaLock", def.CountPermaLock)
	r.AdaptiveAntenna = rd.ReadBool("AdaptiveAntenna", def.AdaptiveAntenna)
	r.ReadPower = int16(rd.ReadInt("PowerRead", int(def.ReadPower)))
	r.WritePower = int16(rd.ReadInt("PowerWrite", int(def.WritePower)))
	r.UseAdjustForEPC = rd.ReadBool("UseAdjustForEPC", def.UseAdjustForEPC)
	r.RewriteEPCBank = rd.ReadBool("RewriteEPCbank", def.RewriteEPCBank)
	r.ErrorHandle = RFIDErrorHandle(rd.ReadInt("ErrorHandle", int(def.ErrorHandle)))
	return nil
}

// deserializeBank reads bank properties from the reader.
// C# RFIDBank attribute keys: prefix + ".Data", prefix + ".DataColumn", etc.
func deserializeBank(rd report.Reader, prefix string, def RFIDBank) RFIDBank {
	return RFIDBank{
		Data:       rd.ReadStr(prefix+".Data", def.Data),
		DataColumn: rd.ReadStr(prefix+".DataColumn", def.DataColumn),
		DataFormat: RFIDBankFormat(rd.ReadInt(prefix+".DataFormat", int(def.DataFormat))),
		Offset:     rd.ReadInt(prefix+".Offset", def.Offset),
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
