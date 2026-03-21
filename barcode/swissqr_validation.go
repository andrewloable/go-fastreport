package barcode

// swissqr_validation.go implements validation helpers for Swiss QR Code payment
// parameters, ported from C# SwissQRCode.cs in the original FastReport.Base.
//
// C# ref: original-dotnet/FastReport.Base/Barcode/SwissQRCode.cs

import (
	"fmt"
	"math/big"
	"regexp"
	"strings"
	"unicode"
)

// swissQRCharsetPattern matches the allowed character set for Swiss QR text
// fields. Pattern sourced from https://qr-validation.iso-payments.ch as
// explained in https://github.com/codebude/QRCoder/issues/97.
//
// C# ref: QRSwissParameters.charsetPattern (SwissQRCode.cs line 27).
var swissQRCharsetPattern = regexp.MustCompile(
	`^([a-zA-Z0-9\.,;:' \+\-/\(\)?\*\[\]\{\}\\` + "`" + `´~ ]|[!"#%&<>÷=@_$£]|[àáâäçèéêëìíîïñòóôöùúûüýßÀÁÂÄÇÈÉÊËÌÍÎÏÒÓÔÖÙÚÛÜÑ])*$`,
)

// swissQRValidCountryCodes is the full ISO 3166-1 two-letter country code list
// used by contact validation.
//
// C# ref: Contact.ValidTwoLetterCodes (SwissQRCode.cs line 382).
var swissQRValidCountryCodes = map[string]struct{}{
	"AF": {}, "AL": {}, "DZ": {}, "AS": {}, "AD": {}, "AO": {}, "AI": {}, "AQ": {}, "AG": {}, "AR": {},
	"AM": {}, "AW": {}, "AU": {}, "AT": {}, "AZ": {}, "BS": {}, "BH": {}, "BD": {}, "BB": {}, "BY": {},
	"BE": {}, "BZ": {}, "BJ": {}, "BM": {}, "BT": {}, "BO": {}, "BQ": {}, "BA": {}, "BW": {}, "BV": {},
	"BR": {}, "IO": {}, "BN": {}, "BG": {}, "BF": {}, "BI": {}, "CV": {}, "KH": {}, "CM": {}, "CA": {},
	"KY": {}, "CF": {}, "TD": {}, "CL": {}, "CN": {}, "CX": {}, "CC": {}, "CO": {}, "KM": {}, "CG": {},
	"CD": {}, "CK": {}, "CR": {}, "CI": {}, "HR": {}, "CU": {}, "CW": {}, "CY": {}, "CZ": {}, "DK": {},
	"DJ": {}, "DM": {}, "DO": {}, "EC": {}, "EG": {}, "SV": {}, "GQ": {}, "ER": {}, "EE": {}, "SZ": {},
	"ET": {}, "FK": {}, "FO": {}, "FJ": {}, "FI": {}, "FR": {}, "GF": {}, "PF": {}, "TF": {}, "GA": {},
	"GM": {}, "GE": {}, "DE": {}, "GH": {}, "GI": {}, "GR": {}, "GL": {}, "GD": {}, "GP": {}, "GU": {},
	"GT": {}, "GG": {}, "GN": {}, "GW": {}, "GY": {}, "HT": {}, "HM": {}, "VA": {}, "HN": {}, "HK": {},
	"HU": {}, "IS": {}, "IN": {}, "ID": {}, "IR": {}, "IQ": {}, "IE": {}, "IM": {}, "IL": {}, "IT": {},
	"JM": {}, "JP": {}, "JE": {}, "JO": {}, "KZ": {}, "KE": {}, "KI": {}, "KP": {}, "KR": {}, "KW": {},
	"KG": {}, "LA": {}, "LV": {}, "LB": {}, "LS": {}, "LR": {}, "LY": {}, "LI": {}, "LT": {}, "LU": {},
	"MO": {}, "MG": {}, "MW": {}, "MY": {}, "MV": {}, "ML": {}, "MT": {}, "MH": {}, "MQ": {}, "MR": {},
	"MU": {}, "YT": {}, "MX": {}, "FM": {}, "MD": {}, "MC": {}, "MN": {}, "ME": {}, "MS": {}, "MA": {},
	"MZ": {}, "MM": {}, "NA": {}, "NR": {}, "NP": {}, "NL": {}, "NC": {}, "NZ": {}, "NI": {}, "NE": {},
	"NG": {}, "NU": {}, "NF": {}, "MP": {}, "MK": {}, "NO": {}, "OM": {}, "PK": {}, "PW": {}, "PS": {},
	"PA": {}, "PG": {}, "PY": {}, "PE": {}, "PH": {}, "PN": {}, "PL": {}, "PT": {}, "PR": {}, "QA": {},
	"RE": {}, "RO": {}, "RU": {}, "RW": {}, "BL": {}, "SH": {}, "KN": {}, "LC": {}, "MF": {}, "PM": {},
	"VC": {}, "WS": {}, "SM": {}, "ST": {}, "SA": {}, "SN": {}, "RS": {}, "SC": {}, "SL": {}, "SG": {},
	"SX": {}, "SK": {}, "SI": {}, "SB": {}, "SO": {}, "ZA": {}, "GS": {}, "SS": {}, "ES": {}, "LK": {},
	"SD": {}, "SR": {}, "SJ": {}, "SE": {}, "CH": {}, "SY": {}, "TW": {}, "TJ": {}, "TZ": {}, "TH": {},
	"TL": {}, "TG": {}, "TK": {}, "TO": {}, "TT": {}, "TN": {}, "TR": {}, "TM": {}, "TC": {}, "TV": {},
	"UG": {}, "UA": {}, "AE": {}, "GB": {}, "US": {}, "UM": {}, "UY": {}, "UZ": {}, "VU": {}, "VE": {},
	"VN": {}, "VG": {}, "VI": {}, "WF": {}, "EH": {}, "YE": {}, "ZM": {}, "ZW": {}, "AX": {},
}

// swissQRCleanAlphanumeric returns only letters and digits from s, uppercased.
func swissQRCleanAlphanumeric(s string) string {
	var b strings.Builder
	for _, r := range s {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			b.WriteRune(unicode.ToUpper(r))
		}
	}
	return b.String()
}

// ValidateIBAN validates an IBAN string using the MOD-97 checksum algorithm and
// verifies it starts with "CH" or "LI" (required by the Swiss QR standard).
//
// The algorithm:
//  1. Strip non-alphanumeric chars and uppercase.
//  2. Check structural pattern: 2 letters, 2 digits, 16-30 alphanumeric chars.
//  3. Rearrange: move first 4 chars to end.
//  4. Convert letters to their numeric equivalents (A=10, B=11 … Z=35).
//  5. Verify the resulting number mod 97 == 1.
//  6. Verify CH or LI prefix (Swiss QR requirement).
//
// C# ref: Iban.IsValidIban and Iban constructor (SwissQRCode.cs lines 498-532).
func ValidateIBAN(iban string) error {
	cleaned := swissQRCleanAlphanumeric(iban)

	// Step 2: structural check.
	structOK := regexp.MustCompile(`^[A-Z]{2}[0-9]{2}[A-Z0-9]{16,30}$`).MatchString(cleaned)
	if !structOK {
		return fmt.Errorf("swissqr: IBAN has invalid structure: %q", iban)
	}

	// Steps 3-5: MOD-97 checksum.
	rearranged := cleaned[4:] + cleaned[:4]
	var numStr strings.Builder
	for _, r := range rearranged {
		if r >= 'A' && r <= 'Z' {
			numStr.WriteString(fmt.Sprintf("%d", int(r-'A')+10))
		} else {
			numStr.WriteRune(r)
		}
	}
	n := new(big.Int)
	if _, ok := n.SetString(numStr.String(), 10); !ok {
		return fmt.Errorf("swissqr: IBAN checksum could not be computed: %q", iban)
	}
	mod := new(big.Int).Mod(n, big.NewInt(97))
	if mod.Int64() != 1 {
		return fmt.Errorf("swissqr: IBAN checksum invalid: %q", iban)
	}

	// Step 6: Swiss QR requires CH or LI prefix.
	if !strings.HasPrefix(cleaned, "CH") && !strings.HasPrefix(cleaned, "LI") {
		return fmt.Errorf("swissqr: IBAN must start with CH or LI, got %q", iban)
	}
	return nil
}

// IsQRIBAN reports whether the IBAN is a QR-IBAN. A QR-IBAN has the IID
// (5-digit bank identifier, positions 4-8 of the cleaned IBAN) in the range
// 30000–31999. Assumes the IBAN has already passed ValidateIBAN.
//
// C# ref: Iban.IsValidQRIban (SwissQRCode.cs lines 522-533).
func IsQRIBAN(iban string) bool {
	cleaned := swissQRCleanAlphanumeric(iban)
	if len(cleaned) < 9 {
		return false
	}
	iidStr := cleaned[4:9]
	var iid int
	if _, err := fmt.Sscanf(iidStr, "%d", &iid); err != nil {
		return false
	}
	return iid >= 30000 && iid <= 31999
}

// ChecksumMod10 validates a digit string using the Swiss/QR MOD-10 recursive
// (Luhn-variant) algorithm used for QR references.
//
// Returns false if the string is fewer than 2 characters, contains any
// non-digit character, or the check digit is wrong.
//
// C# ref: Reference.ChecksumMod10 (SwissQRCode.cs lines 226-244).
func ChecksumMod10(digits string) bool {
	digits = strings.ReplaceAll(digits, " ", "")
	if len(digits) < 2 {
		return false
	}
	for _, r := range digits {
		if !unicode.IsDigit(r) {
			return false
		}
	}
	mods := [10]int{0, 9, 4, 6, 8, 2, 7, 1, 3, 5}
	remainder := 0
	for i := 0; i < len(digits)-1; i++ {
		num := int(digits[i] - '0')
		remainder = mods[(num+remainder)%10]
	}
	checksum := (10 - remainder) % 10
	return checksum == int(digits[len(digits)-1]-'0')
}

// ValidateReference validates a Swiss QR payment reference string against its
// declared reference type.
//
//   - "NON": reference must be empty (only non-alphanumeric characters allowed
//     since those are stripped before the check).
//   - "QRR" (QR reference): cleaned string must be purely numeric, at most 27
//     digits, and must pass the MOD-10 checksum.
//   - "SCOR" (ISO 11649 Creditor Reference): cleaned string must be at most 25
//     alphanumeric characters.
//
// C# ref: Reference constructor (SwissQRCode.cs lines 158-188).
func ValidateReference(refType, reference string) error {
	cleaned := swissQRCleanAlphanumeric(reference)
	switch refType {
	case "NON":
		if cleaned != "" {
			return fmt.Errorf("swissqr: reference must be empty for reference type NON")
		}
	case "QRR":
		if cleaned == "" {
			return fmt.Errorf("swissqr: reference must not be empty for reference type QRR")
		}
		if len(cleaned) > 27 {
			return fmt.Errorf("swissqr: QRR reference must be at most 27 digits, got %d", len(cleaned))
		}
		for _, r := range cleaned {
			if !unicode.IsDigit(r) {
				return fmt.Errorf("swissqr: QRR reference must contain only digits")
			}
		}
		if !ChecksumMod10(cleaned) {
			return fmt.Errorf("swissqr: QRR reference has invalid MOD-10 checksum")
		}
	case "SCOR":
		if cleaned == "" {
			return fmt.Errorf("swissqr: reference must not be empty for reference type SCOR")
		}
		if len(cleaned) > 25 {
			return fmt.Errorf("swissqr: SCOR reference must be at most 25 alphanumeric characters, got %d", len(cleaned))
		}
	default:
		return fmt.Errorf("swissqr: unknown reference type %q (must be QRR, SCOR, or NON)", refType)
	}
	return nil
}

// SwissQRAddressType specifies whether a contact uses structured or combined addressing.
//
// C# ref: Contact.AddressType enum (SwissQRCode.cs lines 424-428).
type SwissQRAddressType int

const (
	// SwissQRStructuredAddress uses separate street, house number, zip, and city fields.
	// Corresponds to C# AddressType.StructuredAddress ("S" in the payload).
	SwissQRStructuredAddress SwissQRAddressType = iota
	// SwissQRCombinedAddress uses two free-form address lines.
	// Corresponds to C# AddressType.CombinedAddress ("K" in the payload).
	SwissQRCombinedAddress
)

// ValidateContact validates a Swiss QR contact (creditor or debtor).
//
// For structured addresses (SwissQRStructuredAddress):
//   - name: required, max 70 chars, charset-checked
//   - streetOrLine1: optional, max 70 chars, charset-checked
//   - houseNumOrLine2: optional, max 16 chars, charset-checked
//   - zip: required, max 16 chars, charset-checked
//   - city: required, max 35 chars, charset-checked
//   - country: valid ISO 3166-1 two-letter code
//
// For combined addresses (SwissQRCombinedAddress):
//   - name: required, max 70 chars, charset-checked
//   - streetOrLine1 (addressLine1): optional, max 70 chars, charset-checked
//   - houseNumOrLine2 (addressLine2): required, max 70 chars, charset-checked
//   - country: valid ISO 3166-1 two-letter code
//
// C# ref: Contact private constructor (SwissQRCode.cs lines 285-373).
func ValidateContact(name, streetOrLine1, houseNumOrLine2, zip, city, country string, addrType SwissQRAddressType) error {
	// Name: required, max 70, charset-checked.
	if name == "" {
		return fmt.Errorf("swissqr: contact name must not be empty")
	}
	if len(name) > 70 {
		return fmt.Errorf("swissqr: contact name must be at most 70 characters, got %d", len(name))
	}
	if !swissQRCharsetPattern.MatchString(name) {
		return fmt.Errorf("swissqr: contact name contains invalid characters: %q", name)
	}

	if addrType == SwissQRStructuredAddress {
		// Street: optional, max 70.
		if streetOrLine1 != "" {
			if len(streetOrLine1) > 70 {
				return fmt.Errorf("swissqr: street must be at most 70 characters, got %d", len(streetOrLine1))
			}
			if !swissQRCharsetPattern.MatchString(streetOrLine1) {
				return fmt.Errorf("swissqr: street contains invalid characters: %q", streetOrLine1)
			}
		}
		// House number: optional, max 16.
		if houseNumOrLine2 != "" {
			if len(houseNumOrLine2) > 16 {
				return fmt.Errorf("swissqr: house number must be at most 16 characters, got %d", len(houseNumOrLine2))
			}
			if !swissQRCharsetPattern.MatchString(houseNumOrLine2) {
				return fmt.Errorf("swissqr: house number contains invalid characters: %q", houseNumOrLine2)
			}
		}
		// Zip: required, max 16.
		if zip == "" {
			return fmt.Errorf("swissqr: zip code must not be empty for structured address")
		}
		if len(zip) > 16 {
			return fmt.Errorf("swissqr: zip code must be at most 16 characters, got %d", len(zip))
		}
		if !swissQRCharsetPattern.MatchString(zip) {
			return fmt.Errorf("swissqr: zip code contains invalid characters: %q", zip)
		}
		// City: required, max 35.
		if city == "" {
			return fmt.Errorf("swissqr: city must not be empty for structured address")
		}
		if len(city) > 35 {
			return fmt.Errorf("swissqr: city must be at most 35 characters, got %d", len(city))
		}
		if !swissQRCharsetPattern.MatchString(city) {
			return fmt.Errorf("swissqr: city contains invalid characters: %q", city)
		}
	} else {
		// Combined address — addressLine1: optional, max 70.
		if streetOrLine1 != "" {
			if len(streetOrLine1) > 70 {
				return fmt.Errorf("swissqr: address line 1 must be at most 70 characters, got %d", len(streetOrLine1))
			}
			if !swissQRCharsetPattern.MatchString(streetOrLine1) {
				return fmt.Errorf("swissqr: address line 1 contains invalid characters: %q", streetOrLine1)
			}
		}
		// AddressLine2: required, max 70.
		if houseNumOrLine2 == "" {
			return fmt.Errorf("swissqr: address line 2 must not be empty for combined address")
		}
		if len(houseNumOrLine2) > 70 {
			return fmt.Errorf("swissqr: address line 2 must be at most 70 characters, got %d", len(houseNumOrLine2))
		}
		if !swissQRCharsetPattern.MatchString(houseNumOrLine2) {
			return fmt.Errorf("swissqr: address line 2 contains invalid characters: %q", houseNumOrLine2)
		}
	}

	// Country: must be a valid ISO 3166-1 two-letter code.
	upper := strings.ToUpper(country)
	if _, ok := swissQRValidCountryCodes[upper]; !ok {
		return fmt.Errorf("swissqr: country code %q is not a valid ISO 3166-1 two-letter code", country)
	}
	return nil
}

// ValidateAdditionalInformation validates the additional information fields used
// in a Swiss QR payment. The combined length of unstructuredMessage and
// billInformation must not exceed 140 characters, and each field must use the
// allowed Swiss QR character set.
//
// C# ref: AdditionalInformation constructor (SwissQRCode.cs lines 86-98).
func ValidateAdditionalInformation(unstructuredMessage, billInformation string) error {
	totalLen := len(unstructuredMessage) + len(billInformation)
	if totalLen > 140 {
		return fmt.Errorf("swissqr: combined length of unstructured message and bill information must not exceed 140 characters, got %d", totalLen)
	}
	if unstructuredMessage != "" && !swissQRCharsetPattern.MatchString(unstructuredMessage) {
		return fmt.Errorf("swissqr: unstructured message contains invalid characters")
	}
	if billInformation != "" && !swissQRCharsetPattern.MatchString(billInformation) {
		return fmt.Errorf("swissqr: bill information contains invalid characters")
	}
	return nil
}

// ValidateSwissQRParameters validates all fields of a SwissQRParameters struct.
// Returns the first error encountered, or nil if all fields are valid.
func ValidateSwissQRParameters(p SwissQRParameters) error {
	// Validate IBAN.
	if err := ValidateIBAN(p.IBAN); err != nil {
		return err
	}

	// Validate currency.
	if p.Currency != "CHF" && p.Currency != "EUR" {
		return fmt.Errorf("swissqr: currency must be CHF or EUR, got %q", p.Currency)
	}

	// Determine address type from Params: if PostalCode and City are both empty,
	// treat it as a combined address (lines 1+2 mode).
	addrType := SwissQRStructuredAddress
	if p.CreditorPostalCode == "" && p.CreditorCity == "" {
		addrType = SwissQRCombinedAddress
	}

	// Validate creditor contact. SwissQRParameters has no separate house-number
	// field, so pass empty string for houseNumOrLine2 in structured mode.
	if err := ValidateContact(
		p.CreditorName,
		p.CreditorStreet,
		"",
		p.CreditorPostalCode,
		p.CreditorCity,
		p.CreditorCountry,
		addrType,
	); err != nil {
		return fmt.Errorf("swissqr: creditor: %w", err)
	}

	// Validate reference.
	refType := p.ReferenceType
	if refType == "" {
		refType = "NON"
	}
	if err := ValidateReference(refType, p.Reference); err != nil {
		return err
	}

	// Validate additional information (unstructured message only, since
	// SwissQRParameters does not expose a separate billInformation field).
	if p.UnstructuredMessage != "" {
		if err := ValidateAdditionalInformation(p.UnstructuredMessage, ""); err != nil {
			return err
		}
	}

	return nil
}
