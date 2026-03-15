package utils

import "testing"

func TestShortPropCode_Known(t *testing.T) {
	code, ok := ShortPropCode("FillColor")
	if !ok {
		t.Fatal("ShortPropCode('FillColor') should return ok=true")
	}
	if code != "fc" {
		t.Errorf("ShortPropCode('FillColor') = %q, want 'fc'", code)
	}
}

func TestShortPropCode_Unknown(t *testing.T) {
	code, ok := ShortPropCode("Unknown")
	if ok {
		t.Error("ShortPropCode('Unknown') should return ok=false")
	}
	if code != "" {
		t.Errorf("ShortPropCode('Unknown') code = %q, want empty", code)
	}
}

func TestShortPropName_Known(t *testing.T) {
	name, ok := ShortPropName("fc")
	if !ok {
		t.Fatal("ShortPropName('fc') should return ok=true")
	}
	if name != "FillColor" {
		t.Errorf("ShortPropName('fc') = %q, want 'FillColor'", name)
	}
}

func TestShortPropName_Unknown(t *testing.T) {
	name, ok := ShortPropName("zz")
	if ok {
		t.Error("ShortPropName('zz') should return ok=false")
	}
	if name != "" {
		t.Errorf("ShortPropName('zz') name = %q, want empty", name)
	}
}

func TestExpandPropName_Known(t *testing.T) {
	if got := ExpandPropName("tx"); got != "Text" {
		t.Errorf("ExpandPropName('tx') = %q, want 'Text'", got)
	}
}

func TestExpandPropName_Unknown(t *testing.T) {
	if got := ExpandPropName("unknown"); got != "unknown" {
		t.Errorf("ExpandPropName('unknown') = %q, want 'unknown'", got)
	}
}

func TestAbbrevPropName_Known(t *testing.T) {
	if got := AbbrevPropName("Text"); got != "tx" {
		t.Errorf("AbbrevPropName('Text') = %q, want 'tx'", got)
	}
}

func TestAbbrevPropName_Unknown(t *testing.T) {
	if got := AbbrevPropName("SomeUnknownProp"); got != "SomeUnknownProp" {
		t.Errorf("AbbrevPropName('SomeUnknownProp') = %q, want 'SomeUnknownProp'", got)
	}
}

func TestShortProps_AllEntriesRoundTrip(t *testing.T) {
	// Every entry in shortToFull must be reversible via fullToShort.
	for code, full := range shortToFull {
		gotCode, ok := ShortPropCode(full)
		if !ok {
			t.Errorf("ShortPropCode(%q) not found (expected code %q)", full, code)
			continue
		}
		if gotCode != code {
			t.Errorf("ShortPropCode(%q) = %q, want %q", full, gotCode, code)
		}
		gotFull, ok2 := ShortPropName(code)
		if !ok2 {
			t.Errorf("ShortPropName(%q) not found (expected full %q)", code, full)
			continue
		}
		if gotFull != full {
			t.Errorf("ShortPropName(%q) = %q, want %q", code, gotFull, full)
		}
	}
}

func TestShortProps_SpecificEntries(t *testing.T) {
	cases := []struct{ short, full string }{
		{"l", "Left"},
		{"t", "Top"},
		{"w", "Width"},
		{"h", "Height"},
		{"tx", "Text"},
		{"ha", "HorzAlign"},
		{"va", "VertAlign"},
		{"ww", "WordWrap"},
		{"fc", "FillColor"},
		{"tc", "TextColor"},
		{"bc", "BackColor"},
		{"fn", "Font.Name"},
		{"fs", "Font.Size"},
		{"fb", "Font.Bold"},
		{"fi", "Font.Italic"},
		{"fu", "Font.Underline"},
		{"bw", "Border.Width"},
		{"bl", "Border.Lines"},
		{"bi", "BlobIdx"},
		{"nm", "Name"},
		{"vi", "Visible"},
		{"ck", "Checked"},
	}
	for _, tc := range cases {
		name, ok := ShortPropName(tc.short)
		if !ok || name != tc.full {
			t.Errorf("ShortPropName(%q) = %q/%v, want %q/true", tc.short, name, ok, tc.full)
		}
		code, ok2 := ShortPropCode(tc.full)
		if !ok2 || code != tc.short {
			t.Errorf("ShortPropCode(%q) = %q/%v, want %q/true", tc.full, code, ok2, tc.short)
		}
	}
}
