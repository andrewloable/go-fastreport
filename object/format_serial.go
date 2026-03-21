package object

import (
	"github.com/andrewloable/go-fastreport/format"
	"github.com/andrewloable/go-fastreport/report"
)

// serializeTextFormat writes Format and Format.* attributes for the given
// format value. If f is nil or a GeneralFormat, nothing is written.
func serializeTextFormat(w report.Writer, f format.Format) {
	if f == nil {
		return
	}
	const pfx = "Format."
	switch v := f.(type) {
	case *format.GeneralFormat:
		// GeneralFormat is the default — write nothing.

	case *format.NumberFormat:
		w.WriteStr("Format", "Number")
		dflt := format.NewNumberFormat()
		if v.DecimalDigits != dflt.DecimalDigits {
			w.WriteInt(pfx+"DecimalDigits", v.DecimalDigits)
		}
		if !v.UseLocaleSettings {
			if v.DecimalSeparator != dflt.DecimalSeparator {
				w.WriteStr(pfx+"DecimalSeparator", v.DecimalSeparator)
			}
			if v.GroupSeparator != dflt.GroupSeparator {
				w.WriteStr(pfx+"GroupSeparator", v.GroupSeparator)
			}
			if v.NegativePattern != dflt.NegativePattern {
				w.WriteInt(pfx+"NegativePattern", v.NegativePattern)
			}
		}
		if !v.UseLocaleSettings {
			// C# FRX attribute name is "UseLocale" (not "UseLocaleSettings")
			w.WriteBool(pfx+"UseLocale", false)
		}

	case *format.CurrencyFormat:
		w.WriteStr("Format", "Currency")
		dflt := format.NewCurrencyFormat()
		if v.DecimalDigits != dflt.DecimalDigits {
			w.WriteInt(pfx+"DecimalDigits", v.DecimalDigits)
		}
		if !v.UseLocaleSettings {
			if v.DecimalSeparator != dflt.DecimalSeparator {
				w.WriteStr(pfx+"DecimalSeparator", v.DecimalSeparator)
			}
			if v.GroupSeparator != dflt.GroupSeparator {
				w.WriteStr(pfx+"GroupSeparator", v.GroupSeparator)
			}
			if v.CurrencySymbol != dflt.CurrencySymbol {
				w.WriteStr(pfx+"CurrencySymbol", v.CurrencySymbol)
			}
			if v.PositivePattern != dflt.PositivePattern {
				w.WriteInt(pfx+"PositivePattern", v.PositivePattern)
			}
			if v.NegativePattern != dflt.NegativePattern {
				w.WriteInt(pfx+"NegativePattern", v.NegativePattern)
			}
		}
		if !v.UseLocaleSettings {
			// C# FRX attribute name is "UseLocale" (not "UseLocaleSettings")
			w.WriteBool(pfx+"UseLocale", false)
		}

	case *format.DateFormat:
		w.WriteStr("Format", "Date")
		dflt := format.NewDateFormat()
		if v.Format != dflt.Format {
			w.WriteStr(pfx+"Format", v.Format)
		}
		// C# DateFormat inherits CustomFormat.Serialize which only writes "Format" — no UseLocale.

	case *format.TimeFormat:
		w.WriteStr("Format", "Time")
		dflt := format.NewTimeFormat()
		if v.Format != dflt.Format {
			w.WriteStr(pfx+"Format", v.Format)
		}
		// C# TimeFormat inherits CustomFormat.Serialize which only writes "Format" — no UseLocale.

	case *format.PercentFormat:
		w.WriteStr("Format", "Percent")
		dflt := format.NewPercentFormat()
		if v.DecimalDigits != dflt.DecimalDigits {
			w.WriteInt(pfx+"DecimalDigits", v.DecimalDigits)
		}
		if !v.UseLocaleSettings {
			if v.DecimalSeparator != dflt.DecimalSeparator {
				w.WriteStr(pfx+"DecimalSeparator", v.DecimalSeparator)
			}
			if v.GroupSeparator != dflt.GroupSeparator {
				w.WriteStr(pfx+"GroupSeparator", v.GroupSeparator)
			}
			if v.PercentSymbol != dflt.PercentSymbol {
				w.WriteStr(pfx+"PercentSymbol", v.PercentSymbol)
			}
			if v.PositivePattern != dflt.PositivePattern {
				w.WriteInt(pfx+"PositivePattern", v.PositivePattern)
			}
			if v.NegativePattern != dflt.NegativePattern {
				w.WriteInt(pfx+"NegativePattern", v.NegativePattern)
			}
		}
		if !v.UseLocaleSettings {
			// C# FRX attribute name is "UseLocale" (not "UseLocaleSettings")
			w.WriteBool(pfx+"UseLocale", false)
		}

	case *format.BooleanFormat:
		w.WriteStr("Format", "Boolean")
		dflt := format.NewBooleanFormat()
		if v.TrueText != dflt.TrueText {
			w.WriteStr(pfx+"TrueText", v.TrueText)
		}
		if v.FalseText != dflt.FalseText {
			w.WriteStr(pfx+"FalseText", v.FalseText)
		}

	case *format.CustomFormat:
		w.WriteStr("Format", "Custom")
		dflt := format.NewCustomFormat()
		if v.Format != dflt.Format {
			w.WriteStr(pfx+"Format", v.Format)
		}
	}
}

// readUseLocale reads the UseLocale/UseLocaleSettings attribute from a format
// element. C# FastReport FRX files use "UseLocale"; older Go-generated files may
// use "UseLocaleSettings". Both are checked so that either source round-trips.
func readUseLocale(r report.Reader, pfx string, def bool) bool {
	// Prefer the C# FRX attribute name "UseLocale".
	v := r.ReadBool(pfx+"UseLocale", def)
	if v != def {
		return v
	}
	// Fall back to the old Go-generated attribute name for backward compatibility.
	return r.ReadBool(pfx+"UseLocaleSettings", def)
}

// deserializeTextFormat reads Format.* attributes from r and returns the
// corresponding Format implementation. typeName is the value of the "Format"
// attribute (e.g. "Number", "Date"). Returns nil for unknown/General.
func deserializeTextFormat(typeName string, r report.Reader) format.Format {
	const pfx = "Format."
	switch typeName {
	case "Number":
		f := format.NewNumberFormat()
		f.DecimalDigits = r.ReadInt(pfx+"DecimalDigits", f.DecimalDigits)
		f.UseLocaleSettings = readUseLocale(r, pfx, f.UseLocaleSettings)
		if !f.UseLocaleSettings {
			f.DecimalSeparator = r.ReadStr(pfx+"DecimalSeparator", f.DecimalSeparator)
			f.GroupSeparator = r.ReadStr(pfx+"GroupSeparator", f.GroupSeparator)
			f.NegativePattern = r.ReadInt(pfx+"NegativePattern", f.NegativePattern)
		}
		return f

	case "Currency":
		f := format.NewCurrencyFormat()
		f.DecimalDigits = r.ReadInt(pfx+"DecimalDigits", f.DecimalDigits)
		f.UseLocaleSettings = readUseLocale(r, pfx, f.UseLocaleSettings)
		if !f.UseLocaleSettings {
			f.DecimalSeparator = r.ReadStr(pfx+"DecimalSeparator", f.DecimalSeparator)
			f.GroupSeparator = r.ReadStr(pfx+"GroupSeparator", f.GroupSeparator)
			f.CurrencySymbol = r.ReadStr(pfx+"CurrencySymbol", f.CurrencySymbol)
			f.PositivePattern = r.ReadInt(pfx+"PositivePattern", f.PositivePattern)
			f.NegativePattern = r.ReadInt(pfx+"NegativePattern", f.NegativePattern)
		}
		return f

	case "Date":
		f := format.NewDateFormat()
		f.Format = r.ReadStr(pfx+"Format", f.Format)
		f.UseLocaleSettings = readUseLocale(r, pfx, f.UseLocaleSettings)
		return f

	case "Time":
		f := format.NewTimeFormat()
		f.Format = r.ReadStr(pfx+"Format", f.Format)
		f.UseLocaleSettings = readUseLocale(r, pfx, f.UseLocaleSettings)
		return f

	case "Percent":
		f := format.NewPercentFormat()
		f.DecimalDigits = r.ReadInt(pfx+"DecimalDigits", f.DecimalDigits)
		f.UseLocaleSettings = readUseLocale(r, pfx, f.UseLocaleSettings)
		if !f.UseLocaleSettings {
			f.DecimalSeparator = r.ReadStr(pfx+"DecimalSeparator", f.DecimalSeparator)
			f.GroupSeparator = r.ReadStr(pfx+"GroupSeparator", f.GroupSeparator)
			f.PercentSymbol = r.ReadStr(pfx+"PercentSymbol", f.PercentSymbol)
			f.PositivePattern = r.ReadInt(pfx+"PositivePattern", f.PositivePattern)
			f.NegativePattern = r.ReadInt(pfx+"NegativePattern", f.NegativePattern)
		}
		return f

	case "Boolean":
		f := format.NewBooleanFormat()
		f.TrueText = r.ReadStr(pfx+"TrueText", f.TrueText)
		f.FalseText = r.ReadStr(pfx+"FalseText", f.FalseText)
		return f

	case "Custom":
		f := format.NewCustomFormat()
		f.Format = r.ReadStr(pfx+"Format", f.Format)
		return f

	default:
		return nil
	}
}

// deserializeFormatFromChild reads format properties from a child element named
// after the format type (e.g. "NumberFormat", "DateFormat"). Called from
// TextObject.DeserializeChild when handling the <Formats> multi-format case.
func deserializeFormatFromChild(childType string, r report.Reader) format.Format {
	switch childType {
	case "NumberFormat":
		f := format.NewNumberFormat()
		f.DecimalDigits = r.ReadInt("DecimalDigits", f.DecimalDigits)
		f.UseLocaleSettings = readUseLocale(r, "", f.UseLocaleSettings)
		f.DecimalSeparator = r.ReadStr("DecimalSeparator", f.DecimalSeparator)
		f.GroupSeparator = r.ReadStr("GroupSeparator", f.GroupSeparator)
		f.NegativePattern = r.ReadInt("NegativePattern", f.NegativePattern)
		return f

	case "CurrencyFormat":
		f := format.NewCurrencyFormat()
		f.DecimalDigits = r.ReadInt("DecimalDigits", f.DecimalDigits)
		f.UseLocaleSettings = readUseLocale(r, "", f.UseLocaleSettings)
		f.DecimalSeparator = r.ReadStr("DecimalSeparator", f.DecimalSeparator)
		f.GroupSeparator = r.ReadStr("GroupSeparator", f.GroupSeparator)
		f.CurrencySymbol = r.ReadStr("CurrencySymbol", f.CurrencySymbol)
		f.PositivePattern = r.ReadInt("PositivePattern", f.PositivePattern)
		f.NegativePattern = r.ReadInt("NegativePattern", f.NegativePattern)
		return f

	case "DateFormat":
		f := format.NewDateFormat()
		f.Format = r.ReadStr("Format", f.Format)
		f.UseLocaleSettings = readUseLocale(r, "", f.UseLocaleSettings)
		return f

	case "TimeFormat":
		f := format.NewTimeFormat()
		f.Format = r.ReadStr("Format", f.Format)
		f.UseLocaleSettings = readUseLocale(r, "", f.UseLocaleSettings)
		return f

	case "PercentFormat":
		f := format.NewPercentFormat()
		f.DecimalDigits = r.ReadInt("DecimalDigits", f.DecimalDigits)
		f.UseLocaleSettings = readUseLocale(r, "", f.UseLocaleSettings)
		f.DecimalSeparator = r.ReadStr("DecimalSeparator", f.DecimalSeparator)
		f.GroupSeparator = r.ReadStr("GroupSeparator", f.GroupSeparator)
		f.PercentSymbol = r.ReadStr("PercentSymbol", f.PercentSymbol)
		f.PositivePattern = r.ReadInt("PositivePattern", f.PositivePattern)
		f.NegativePattern = r.ReadInt("NegativePattern", f.NegativePattern)
		return f

	case "BooleanFormat":
		f := format.NewBooleanFormat()
		f.TrueText = r.ReadStr("TrueText", f.TrueText)
		f.FalseText = r.ReadStr("FalseText", f.FalseText)
		return f

	case "CustomFormat":
		f := format.NewCustomFormat()
		f.Format = r.ReadStr("Format", f.Format)
		return f

	case "GeneralFormat":
		return format.NewGeneralFormat()

	default:
		return nil
	}
}
