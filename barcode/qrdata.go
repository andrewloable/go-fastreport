// qrdata.go implements FastReport-specific QR data payload types.
// Ported from C# FastReport.Barcode.QRCode.QRData.cs.
//
// Each concrete type provides Pack() to encode the payload as a string and
// Unpack() to parse a payload string back into typed fields. A top-level
// ParseQRData() dispatcher (equivalent to C# QRData.Parse()) selects the
// appropriate type by inspecting well-known prefixes.
package barcode

import (
	"fmt"
	"regexp"
	"strings"
	"time"
)

// ── QRDataType — payload type enum ──────────────────────────────────────────

// QRDataKind identifies the structured payload type of a QR code.
// C# QRData.cs: abstract class QRData and its concrete subclasses.
type QRDataKind int

const (
	QRDataKindText         QRDataKind = iota // Plain text (default)
	QRDataKindVCard                          // BEGIN:VCARD
	QRDataKindURI                            // Well-formed absolute URI
	QRDataKindEmailAddress                   // Email address (regex-detected)
	QRDataKindEmailMessage                   // MATMSG:
	QRDataKindGeo                            // geo:
	QRDataKindSMS                            // SMSTO:
	QRDataKindCall                           // tel:
	QRDataKindEvent                          // BEGIN:VEVENT
	QRDataKindWifi                           // WIFI:
	QRDataKindSwiss                          // SPC (Swiss QR bill)
	QRDataKindSberBank                       // ST (Sberbank payment)
)

// QRPayload is the interface implemented by all QR data payload types.
// C# QRData.Pack() / QRData.Unpack().
type QRPayload interface {
	// Kind returns the payload type.
	Kind() QRDataKind
	// Pack serialises the structured fields into the QR payload string.
	Pack() string
	// Unpack parses a raw QR payload string into the structured fields.
	Unpack(data string)
}

// ParseQRData selects and returns the appropriate QRPayload for the given raw
// QR payload string. Mirrors C# QRData.Parse() (QRData.cs:33-75).
func ParseQRData(data string) QRPayload {
	if data == "" {
		t := &QRText{}
		t.Unpack(data)
		return t
	}
	// Try each format in the same order as C# QRData.Parse().
	if strings.HasPrefix(data, "BEGIN:VCARD") {
		p := &QRVCard{}
		p.Unpack(data)
		return p
	}
	if strings.HasPrefix(data, "MATMSG:") {
		p := &QREmailMessage{}
		p.Unpack(data)
		return p
	}
	if strings.HasPrefix(data, "geo:") {
		p := &QRGeo{}
		p.Unpack(data)
		return p
	}
	if strings.HasPrefix(data, "SMSTO:") {
		p := &QRSMS{}
		p.Unpack(data)
		return p
	}
	if strings.HasPrefix(data, "tel:") {
		p := &QRCall{}
		p.Unpack(data)
		return p
	}
	if strings.HasPrefix(data, "BEGIN:VEVENT") {
		p := &QREvent{}
		p.Unpack(data)
		return p
	}
	if strings.HasPrefix(data, "WIFI:") {
		p := &QRWifi{}
		p.Unpack(data)
		return p
	}
	// Email address by regex (C# QREmailAddress.TryParse).
	if emailRegex.MatchString(data) {
		p := &QREmailAddress{Data: data}
		return p
	}
	// Swiss QR bill: starts with "SPC"
	if strings.HasPrefix(data, "SPC") {
		p := &QRSwiss{}
		p.Unpack(data)
		return p
	}
	// Sberbank payment: starts with "ST"
	if strings.HasPrefix(data, "ST") {
		p := &QRSberBank{}
		p.Unpack(data)
		return p
	}
	// Plain text fallback.
	t := &QRText{Data: data}
	return t
}

// emailRegex matches a simple email address. Mirrors C# QREmailAddress regex.
var emailRegex = regexp.MustCompile(`(?i)^([\w\-.]+)@(([\w\-]+\.)+)([a-zA-Z]{2,4}|[0-9]{1,3})\]?$`)

// ── QRText ───────────────────────────────────────────────────────────────────

// QRText is plain text payload. C# class QRText : QRData.
type QRText struct {
	Data string
}

func (q *QRText) Kind() QRDataKind { return QRDataKindText }
func (q *QRText) Pack() string     { return q.Data }
func (q *QRText) Unpack(data string) {
	q.Data = data
}

// ── QRVCard ──────────────────────────────────────────────────────────────────

// QRVCard encodes a vCard 2.1 contact. C# class QRvCard : QRData.
type QRVCard struct {
	FirstName          string
	LastName           string
	FN                 string // FN field (full name as stored)
	Title              string
	Org                string
	URL                string
	TelCell            string
	TelWorkVoice       string
	TelHomeVoice       string
	EmailHomeInternet  string
	EmailWorkInternet  string
	Street             string
	ZipCode            string
	City               string
	Country            string
}

func (q *QRVCard) Kind() QRDataKind { return QRDataKindVCard }

// Pack serialises the vCard. C# QRvCard.Pack() (QRData.cs:123-154).
func (q *QRVCard) Pack() string {
	var sb strings.Builder
	sb.WriteString("BEGIN:VCARD\nVERSION:2.1\n")
	if q.FirstName != "" || q.LastName != "" {
		sb.WriteString("FN:" + q.FirstName + " " + q.LastName + "\n")
		sb.WriteString("N:" + q.LastName + ";" + q.FirstName + "\n")
	}
	appendField := func(prefix, val string) {
		if val != "" {
			sb.WriteString(prefix + val + "\n")
		}
	}
	appendField("TITLE:", q.Title)
	appendField("ORG:", q.Org)
	appendField("URL:", q.URL)
	appendField("TEL;CELL:", q.TelCell)
	appendField("TEL;WORK;VOICE:", q.TelWorkVoice)
	appendField("TEL;HOME;VOICE:", q.TelHomeVoice)
	appendField("EMAIL;HOME;INTERNET:", q.EmailHomeInternet)
	appendField("EMAIL;WORK;INTERNET:", q.EmailWorkInternet)
	if q.Street != "" || q.ZipCode != "" || q.City != "" || q.Country != "" {
		sb.WriteString("ADR:;;" + q.Street + ";" + q.City + ";;" + q.ZipCode + ";" + q.Country + "\n")
	}
	sb.WriteString("END:VCARD")
	return sb.String()
}

// Unpack parses a vCard 2.1 payload. C# QRvCard.Unpack() (QRData.cs:164-215).
func (q *QRVCard) Unpack(data string) {
	lines := strings.Split(data, "\n")
	for _, line := range lines {
		line = strings.TrimRight(line, "\r")
		if line == "" {
			continue
		}
		parts := strings.SplitN(line, ":", 2)
		if len(parts) < 2 {
			continue
		}
		key, val := parts[0], parts[1]
		switch key {
		case "FN":
			q.FN = val
		case "N":
			ns := strings.SplitN(val, ";", 2)
			if len(ns) >= 1 {
				q.LastName = ns[0]
			}
			if len(ns) >= 2 {
				q.FirstName = ns[1]
			}
		case "TITLE":
			q.Title = val
		case "ORG":
			q.Org = val
		case "URL":
			q.URL = val
		case "TEL;CELL":
			q.TelCell = val
		case "TEL;WORK;VOICE":
			q.TelWorkVoice = val
		case "TEL;HOME;VOICE":
			q.TelHomeVoice = val
		case "EMAIL;HOME;INTERNET":
			q.EmailHomeInternet = val
		case "EMAIL;WORK;INTERNET":
			q.EmailWorkInternet = val
		case "ADR":
			// C# QRvCard.Unpack ADR: s[2]=street, s[3]=city, s[5]=zipCode, s[6]=country
			adr := strings.Split(val, ";")
			if len(adr) >= 3 {
				q.Street = adr[2]
			}
			if len(adr) >= 4 {
				q.City = adr[3]
			}
			if len(adr) >= 6 {
				q.ZipCode = adr[5]
			}
			if len(adr) >= 7 {
				q.Country = adr[6]
			}
		}
	}
}

// ── QRURI ────────────────────────────────────────────────────────────────────

// QRURI wraps a URI payload. C# class QRURI : QRData.
type QRURI struct {
	Data string
}

func (q *QRURI) Kind() QRDataKind  { return QRDataKindURI }
func (q *QRURI) Pack() string      { return q.Data }
func (q *QRURI) Unpack(data string) { q.Data = data }

// ── QREmailAddress ───────────────────────────────────────────────────────────

// QREmailAddress wraps a bare email address. C# class QREmailAddress : QRData.
type QREmailAddress struct {
	Data string
}

func (q *QREmailAddress) Kind() QRDataKind  { return QRDataKindEmailAddress }
func (q *QREmailAddress) Pack() string      { return q.Data }
func (q *QREmailAddress) Unpack(data string) { q.Data = data }

// ── QREmailMessage ───────────────────────────────────────────────────────────

// QREmailMessage encodes a MATMSG email message. C# class QREmailMessage : QRData.
type QREmailMessage struct {
	To      string
	Subject string
	Body    string
}

func (q *QREmailMessage) Kind() QRDataKind { return QRDataKindEmailMessage }

// Pack encodes the email message. C# QREmailMessage.Pack() (QRData.cs:286-288).
func (q *QREmailMessage) Pack() string {
	return "MATMSG:TO:" + q.To + ";SUB:" + q.Subject + ";BODY:" + q.Body + ";;"
}

// Unpack parses a MATMSG payload. C# QREmailMessage.Unpack() (QRData.cs:291-298).
func (q *QREmailMessage) Unpack(data string) {
	// MATMSG:TO:<to>;SUB:<sub>;BODY:<body>;;
	data = strings.TrimPrefix(data, "MATMSG:")
	parts := splitKV(data, []string{"TO:", ";SUB:", ";BODY:"})
	if len(parts) >= 1 {
		q.To = parts[0]
	}
	if len(parts) >= 2 {
		q.Subject = parts[1]
	}
	if len(parts) >= 3 {
		body := parts[2]
		// strip trailing ";;" C# does Remove(s[3].Length-2,2)
		body = strings.TrimSuffix(body, ";;")
		q.Body = body
	}
}

// ── QRGeo ────────────────────────────────────────────────────────────────────

// QRGeo encodes a geo: URI. C# class QRGeo : QRData.
type QRGeo struct {
	Latitude  string
	Longitude string
	Meters    string
}

func (q *QRGeo) Kind() QRDataKind { return QRDataKindGeo }

// Pack encodes geo payload. C# QRGeo.Pack() (QRData.cs:661-663).
func (q *QRGeo) Pack() string {
	return "geo:" + q.Latitude + "," + q.Longitude + "," + q.Meters
}

// Unpack parses a geo: payload. C# QRGeo.Unpack() (QRData.cs:666-673).
func (q *QRGeo) Unpack(data string) {
	data = strings.TrimPrefix(data, "geo:")
	parts := strings.SplitN(data, ",", 3)
	if len(parts) >= 1 {
		q.Latitude = parts[0]
	}
	if len(parts) >= 2 {
		q.Longitude = parts[1]
	}
	if len(parts) >= 3 {
		q.Meters = parts[2]
	}
}

// ── QRSMS ────────────────────────────────────────────────────────────────────

// QRSMS encodes an SMS message. C# class QRSMS : QRData.
type QRSMS struct {
	To   string
	Text string
}

func (q *QRSMS) Kind() QRDataKind { return QRDataKindSMS }

// Pack encodes the SMS payload. C# QRSMS.Pack() (QRData.cs:703-705).
func (q *QRSMS) Pack() string {
	return "SMSTO:" + q.To + ":" + q.Text
}

// Unpack parses an SMSTO payload. C# QRSMS.Unpack() (QRData.cs:708-713).
func (q *QRSMS) Unpack(data string) {
	data = strings.TrimPrefix(data, "SMSTO:")
	parts := strings.SplitN(data, ":", 2)
	if len(parts) >= 1 {
		q.To = parts[0]
	}
	if len(parts) >= 2 {
		q.Text = parts[1]
	}
}

// ── QRCall ───────────────────────────────────────────────────────────────────

// QRCall encodes a telephone call URI. C# class QRCall : QRData.
type QRCall struct {
	Tel string
}

func (q *QRCall) Kind() QRDataKind { return QRDataKindCall }

// Pack encodes the tel: payload. C# QRCall.Pack() (QRData.cs:742-744).
func (q *QRCall) Pack() string { return "tel:" + q.Tel }

// Unpack parses a tel: payload. C# QRCall.Unpack() (QRData.cs:747-749).
func (q *QRCall) Unpack(data string) {
	q.Tel = strings.TrimPrefix(data, "tel:")
}

// ── QREvent ──────────────────────────────────────────────────────────────────

// QREvent encodes a calendar event (VEVENT). C# class QREvent : QRData.
type QREvent struct {
	Summary string
	Start   time.Time
	End     time.Time
}

func (q *QREvent) Kind() QRDataKind { return QRDataKindEvent }

// Pack encodes the VEVENT payload. C# QREvent.Pack() (QRData.cs:781-797).
func (q *QREvent) Pack() string {
	fmtDT := func(t time.Time) string {
		return fmt.Sprintf("%04d%02d%02dT%02d%02d%02dZ",
			t.Year(), int(t.Month()), t.Day(),
			t.Hour(), t.Minute(), t.Second())
	}
	return "BEGIN:VEVENT\nSUMMARY:" + q.Summary +
		"\nDTSTART:" + fmtDT(q.Start) +
		"\nDTEND:" + fmtDT(q.End) +
		"\nEND:VEVENT"
}

// Unpack parses a VEVENT payload. C# QREvent.Unpack() (QRData.cs:799-835).
func (q *QREvent) Unpack(data string) {
	lines := strings.Split(data, "\n")
	var dtStart, dtEnd string
	for _, line := range lines {
		line = strings.TrimRight(line, "\r")
		parts := strings.SplitN(line, ":", 2)
		if len(parts) < 2 {
			continue
		}
		switch parts[0] {
		case "SUMMARY":
			q.Summary = parts[1]
		case "DTSTART":
			dtStart = parts[1]
		case "DTEND":
			dtEnd = parts[1]
		}
	}
	q.Start = parseVEventDT(dtStart)
	q.End = parseVEventDT(dtEnd)
}

// parseVEventDT parses a VEVENT date-time string (YYYYMMDDTHHmmssZ).
func parseVEventDT(s string) time.Time {
	s = strings.TrimSuffix(s, "Z")
	if len(s) < 15 {
		return time.Time{}
	}
	t, err := time.Parse("20060102T150405", s)
	if err != nil {
		return time.Time{}
	}
	return t
}

// ── QRWifi ───────────────────────────────────────────────────────────────────

// QRWifi encodes Wi-Fi network credentials. C# class QRWifi : QRData.
type QRWifi struct {
	Encryption  string // "WPA", "WEP", "unencrypted", etc.
	NetworkName string
	Password    string
	Hidden      bool
}

func (q *QRWifi) Kind() QRDataKind { return QRDataKindWifi }

// Pack encodes the WIFI: payload. C# QRWifi.Pack() (QRData.cs:868-872).
func (q *QRWifi) Pack() string {
	enc := q.Encryption
	pass := q.Password
	if enc == "unencrypted" {
		enc = "nopass"
		pass = ""
	}
	hidden := ""
	if q.Hidden {
		hidden = ";H:true;"
	} else {
		hidden = ";;"
	}
	return "WIFI:T:" + enc + ";S:" + q.NetworkName + ";P:" + pass + hidden
}

// Unpack parses a WIFI: payload. C# QRWifi.Unpack() (QRData.cs:875-882).
func (q *QRWifi) Unpack(data string) {
	// WIFI:T:<enc>;S:<ssid>;P:<pass>;H:<hidden>;;
	data = strings.TrimPrefix(data, "WIFI:T:")
	// Split on ;S:, ;P:, ;H:, ;;
	// Simple field extraction.
	if i := strings.Index(data, ";S:"); i >= 0 {
		q.Encryption = data[:i]
		data = data[i+3:]
	}
	if i := strings.Index(data, ";P:"); i >= 0 {
		q.NetworkName = data[:i]
		data = data[i+3:]
	}
	if i := strings.Index(data, ";H:"); i >= 0 {
		q.Password = data[:i]
		hidden := data[i+3:]
		hidden = strings.TrimSuffix(hidden, ";;")
		q.Hidden = strings.HasPrefix(hidden, "true")
	} else {
		// no hidden field — strip trailing ";;"
		q.Password = strings.TrimSuffix(data, ";;")
	}
	if q.Encryption == "nopass" {
		q.Encryption = "unencrypted"
	}
}

// ── QRSwiss ──────────────────────────────────────────────────────────────────

// QRSwiss holds a Swiss QR bill payload. C# class QRSwiss : QRData.
// The full Swiss QR payload implementation is in swissqr_validation.go;
// this type provides the QRPayload interface wrapping for ParseQRData.
type QRSwiss struct {
	RawData string
}

func (q *QRSwiss) Kind() QRDataKind  { return QRDataKindSwiss }
func (q *QRSwiss) Pack() string      { return q.RawData }
func (q *QRSwiss) Unpack(data string) { q.RawData = data }

// ── QRSberBank ───────────────────────────────────────────────────────────────

// QRSberBank encodes a Sberbank payment order (ST format).
// C# class QRSberBank : QRData (QRData.cs:318-648).
type QRSberBank struct {
	FormatIdentifier string
	VersionStandart  string
	Encoding         string
	// Required fields
	Name        string
	PersonalAcc string
	BankName    string
	BIC         string
	CorrespAcc  string
	// Optional payment fields
	Sum           string
	Purpose       string
	PayeeINN      string
	PayerINN      string
	DrawerStatus  string
	KPP           string
	CBC           string
	OKTMO         string
	PaytReason    string
	TaxPeriod     string
	DocNo         string
	DocDate       time.Time
	TaxPaytKind   string
	// Additional personal info
	LastName      string
	FirstName     string
	MiddleName    string
	PayerAddress  string
	PersonalAccount string
	DocIdx        string
	PensAcc       string
	Contract      string
	PersAcc       string
	Flat          string
	Phone         string
	PayerIdType   string
	PayerIdNum    string
	ChildFio      string
	BirthDate     time.Time
	PaymTerm      time.Time
	PaymPeriod    string
	Category      string
	ServiceName   string
	CounterId     string
	CounterVal    string
	QuittId       string
	QuittDate     time.Time
	InstNum       string
	ClassNum      string
	SpecFio       string
	AddAmount     string
	RuleId        string
	ExecId        string
	RegType       string
	UIN           string
	TechCode      string

	separator rune
}

func (q *QRSberBank) Kind() QRDataKind { return QRDataKindSberBank }

// Pack serialises the Sberbank payment order.
// C# QRSberBank.Pack() (QRData.cs:387-455).
func (q *QRSberBank) Pack() string {
	sep := q.separator
	if sep == 0 {
		sep = '|'
	}
	fi := q.FormatIdentifier
	if fi == "" {
		fi = "ST"
	}
	vs := q.VersionStandart
	if vs == "" {
		vs = "0001"
	}
	enc := q.Encoding
	if enc == "" {
		enc = "2"
	}

	s := string(sep)
	result := fi + vs + enc
	result += s + "Name=" + q.Name
	result += s + "PersonalAcc=" + q.PersonalAcc
	result += s + "BankName=" + q.BankName
	result += s + "BIC=" + q.BIC
	result += s + "CorrespAcc=" + q.CorrespAcc
	if q.Sum != "" {
		result += s + "Sum=" + q.Sum
	}
	if q.Purpose != "" {
		result += s + "Purpose=" + q.Purpose
	}
	if q.PayeeINN != "" {
		result += s + "PayeeINN=" + q.PayeeINN
	}
	if q.PayerINN != "" {
		result += s + "PayerINN=" + q.PayerINN
	}
	if q.DrawerStatus != "" {
		result += s + "DrawerStatus=" + q.DrawerStatus
	}
	if q.KPP != "" {
		result += s + "KPP=" + q.KPP
	}
	if q.CBC != "" {
		result += s + "CBC=" + q.CBC
	}
	if q.OKTMO != "" {
		result += s + "OKTMO=" + q.OKTMO
	}
	if q.PaytReason != "" {
		result += s + "PaytReason=" + q.PaytReason
	}
	if q.TaxPeriod != "" {
		result += s + "TaxPeriod=" + q.TaxPeriod
	}
	if q.DocNo != "" {
		result += s + "DocNo=" + q.DocNo
	}
	if !q.DocDate.IsZero() {
		result += s + "DocDate=" + q.DocDate.Format("01.02.2006")
	}
	if q.TaxPaytKind != "" {
		result += s + "TaxPaytKind=" + q.TaxPaytKind
	}
	if q.LastName != "" {
		result += s + "LastName=" + q.LastName
	}
	if q.FirstName != "" {
		result += s + "FirstName=" + q.FirstName
	}
	if q.MiddleName != "" {
		result += s + "MiddleName=" + q.MiddleName
	}
	if q.PayerAddress != "" {
		result += s + "PayerAddress=" + q.PayerAddress
	}
	if q.PersonalAccount != "" {
		result += s + "PersonalAccount=" + q.PersonalAccount
	}
	if q.DocIdx != "" {
		result += s + "DocIdx=" + q.DocIdx
	}
	if q.PensAcc != "" {
		result += s + "PensAcc=" + q.PensAcc
	}
	if q.Contract != "" {
		result += s + "Contract=" + q.Contract
	}
	if q.PersAcc != "" {
		result += s + "PersAcc=" + q.PersAcc
	}
	if q.Flat != "" {
		result += s + "Flat=" + q.Flat
	}
	if q.Phone != "" {
		result += s + "Phone=" + q.Phone
	}
	if q.PayerIdType != "" {
		result += s + "PayerIdType=" + q.PayerIdType
	}
	if q.PayerIdNum != "" {
		result += s + "PayerIdNum=" + q.PayerIdNum
	}
	if q.ChildFio != "" {
		result += s + "ChildFio=" + q.ChildFio
	}
	if !q.BirthDate.IsZero() {
		result += s + "BirthDate=" + q.BirthDate.Format("01.02.2006")
	}
	if !q.PaymTerm.IsZero() {
		result += s + "PaymTerm=" + q.PaymTerm.Format("01.02.2006")
	}
	if q.PaymPeriod != "" {
		result += s + "PaymPeriod=" + q.PaymPeriod
	}
	if q.Category != "" {
		result += s + "Category=" + q.Category
	}
	if q.ServiceName != "" {
		result += s + "ServiceName=" + q.ServiceName
	}
	if q.CounterId != "" {
		result += s + "CounterId=" + q.CounterId
	}
	if q.CounterVal != "" {
		result += s + "CounterVal=" + q.CounterVal
	}
	if q.QuittId != "" {
		result += s + "QuittId=" + q.QuittId
	}
	if !q.QuittDate.IsZero() {
		result += s + "QuittDate=" + q.QuittDate.Format("01.02.2006")
	}
	if q.InstNum != "" {
		result += s + "InstNum=" + q.InstNum
	}
	if q.ClassNum != "" {
		result += s + "ClassNum=" + q.ClassNum
	}
	if q.SpecFio != "" {
		result += s + "SpecFio=" + q.SpecFio
	}
	if q.AddAmount != "" {
		result += s + "AddAmount=" + q.AddAmount
	}
	if q.RuleId != "" {
		result += s + "RuleId=" + q.RuleId
	}
	if q.ExecId != "" {
		result += s + "ExecId=" + q.ExecId
	}
	if q.RegType != "" {
		result += s + "RegType=" + q.RegType
	}
	if q.UIN != "" {
		result += s + "UIN=" + q.UIN
	}
	if q.TechCode != "" {
		result += s + "TechCode=" + q.TechCode
	}
	return result
}

// Unpack parses the Sberbank payment order payload.
// C# QRSberBank.Unpack() / RetrieveServiceData() (QRData.cs:456-630).
func (q *QRSberBank) Unpack(data string) {
	q.separator = '|'
	// RetrieveServiceData: read header up to first separator.
	idx := strings.IndexByte(data, '|')
	if idx < 0 {
		return
	}
	header := data[:idx]
	data = data[idx+1:]
	if len(header) >= 2 {
		q.FormatIdentifier = header[:2]
	}
	if len(header) >= 6 {
		q.VersionStandart = header[2:6]
	}
	if len(header) >= 7 {
		q.Encoding = header[6:7]
	}

	pairs := strings.Split(data, "|")
	for _, pair := range pairs {
		kv := strings.SplitN(pair, "=", 2)
		if len(kv) < 2 {
			continue
		}
		key, val := kv[0], kv[1]
		switch key {
		case "Name":
			q.Name = val
		case "PersonalAcc":
			q.PersonalAcc = val
		case "BankName":
			q.BankName = val
		case "BIC":
			q.BIC = val
		case "CorrespAcc":
			q.CorrespAcc = val
		case "Sum":
			q.Sum = val
		case "Purpose":
			q.Purpose = val
		case "PayeeINN":
			q.PayeeINN = val
		case "PayerINN":
			q.PayerINN = val
		case "DrawerStatus":
			q.DrawerStatus = val
		case "KPP":
			q.KPP = val
		case "CBC":
			q.CBC = val
		case "OKTMO":
			q.OKTMO = val
		case "PaytReason":
			q.PaytReason = val
		case "TaxPeriod":
			q.TaxPeriod = val
		case "DocNo":
			q.DocNo = val
		case "DocDate":
			if t, err := time.Parse("01.02.2006", val); err == nil {
				q.DocDate = t
			}
		case "TaxPaytKind":
			q.TaxPaytKind = val
		case "LastName":
			q.LastName = val
		case "FirstName":
			q.FirstName = val
		case "MiddleName":
			q.MiddleName = val
		case "PayerAddress":
			q.PayerAddress = val
		case "PersonalAccount":
			q.PersonalAccount = val
		case "DocIdx":
			q.DocIdx = val
		case "PensAcc":
			q.PensAcc = val
		case "Contract":
			q.Contract = val
		case "PersAcc":
			q.PersAcc = val
		case "Flat":
			q.Flat = val
		case "Phone":
			q.Phone = val
		case "PayerIdType":
			q.PayerIdType = val
		case "PayerIdNum":
			q.PayerIdNum = val
		case "ChildFio":
			q.ChildFio = val
		case "BirthDate":
			if t, err := time.Parse("01.02.2006", val); err == nil {
				q.BirthDate = t
			}
		case "PaymTerm":
			if t, err := time.Parse("01.02.2006", val); err == nil {
				q.PaymTerm = t
			}
		case "PaymPeriod":
			q.PaymPeriod = val
		case "Category":
			q.Category = val
		case "ServiceName":
			q.ServiceName = val
		case "CounterId":
			q.CounterId = val
		case "CounterVal":
			q.CounterVal = val
		case "QuittId":
			q.QuittId = val
		case "QuittDate":
			if t, err := time.Parse("01.02.2006", val); err == nil {
				q.QuittDate = t
			}
		case "InstNum":
			q.InstNum = val
		case "ClassNum":
			q.ClassNum = val
		case "SpecFio":
			q.SpecFio = val
		case "AddAmount":
			q.AddAmount = val
		case "RuleId":
			q.RuleId = val
		case "ExecId":
			q.ExecId = val
		case "RegType":
			q.RegType = val
		case "UIN":
			q.UIN = val
		case "TechCode":
			q.TechCode = val
		}
	}
}

// ── helper ──────────────────────────────────────────────────────────────────

// splitKV splits s by each separator in seps in sequence, returning the pieces
// in between. Used by QREmailMessage.Unpack.
func splitKV(s string, seps []string) []string {
	var result []string
	for _, sep := range seps {
		idx := strings.Index(s, sep)
		if idx < 0 {
			result = append(result, s)
			s = ""
			break
		}
		result = append(result, s[:idx])
		s = s[idx+len(sep):]
	}
	if s != "" {
		result = append(result, s)
	}
	return result
}
