package utils

import (
	"os"
	"strings"
	"testing"
)

func TestResGet_Builtin(t *testing.T) {
	ResetToBuiltin()

	got := ResGet("Objects,Text")
	if got != "Text" {
		t.Errorf("ResGet(Objects,Text) = %q, want %q", got, "Text")
	}

	got = ResGet("Bands,Data")
	if got != "Data" {
		t.Errorf("ResGet(Bands,Data) = %q, want %q", got, "Data")
	}
}

func TestResGet_Missing(t *testing.T) {
	ResetToBuiltin()

	got := ResGet("NoSuchCategory,NoSuchKey")
	if !strings.Contains(got, "NOT LOCALIZED") {
		t.Errorf("ResGet missing key should contain NOT LOCALIZED, got %q", got)
	}
}

func TestResSet_Override(t *testing.T) {
	ResetToBuiltin()

	ResSet("Objects,Text", "Texto")
	got := ResGet("Objects,Text")
	if got != "Texto" {
		t.Errorf("ResGet after ResSet = %q, want %q", got, "Texto")
	}

	// Reset for other tests.
	ResetToBuiltin()
}

func TestLoadLocaleReader(t *testing.T) {
	xml := `<Locale Name="fr">
  <Objects Name="Objects">
    <Item Name="Text" Text="Texte"/>
  </Objects>
</Locale>`

	err := LoadLocaleReader(strings.NewReader(xml))
	if err != nil {
		t.Fatalf("LoadLocaleReader: %v", err)
	}

	got := ResGet("Objects,Text")
	if got != "Texte" {
		t.Errorf("ResGet after LoadLocaleReader = %q, want %q", got, "Texte")
	}

	if LocaleName() != "fr" {
		t.Errorf("LocaleName = %q, want %q", LocaleName(), "fr")
	}

	// Missing key falls back to built-in English.
	got = ResGet("Bands,Data")
	if got != "Data" {
		t.Errorf("Fallback to builtin: ResGet(Bands,Data) = %q, want %q", got, "Data")
	}

	ResetToBuiltin()
}

func TestLocaleName_Default(t *testing.T) {
	ResetToBuiltin()
	if LocaleName() != "en" {
		t.Errorf("Default LocaleName = %q, want %q", LocaleName(), "en")
	}
}

func TestLoadLocale_ValidFile(t *testing.T) {
	xmlContent := `<Locale Name="de"><Objects Name="Objects"><Item Name="Text" Text="Text_DE"/></Objects></Locale>`
	f, err := os.CreateTemp("", "locale_*.frl")
	if err != nil {
		t.Fatalf("CreateTemp: %v", err)
	}
	defer os.Remove(f.Name())
	if _, err := f.WriteString(xmlContent); err != nil {
		t.Fatalf("write: %v", err)
	}
	f.Close()

	if err := LoadLocale(f.Name()); err != nil {
		t.Fatalf("LoadLocale: %v", err)
	}
	if LocaleName() != "de" {
		t.Errorf("LocaleName = %q, want 'de'", LocaleName())
	}
	ResetToBuiltin()
}

func TestLoadLocale_MissingFile(t *testing.T) {
	err := LoadLocale("/nonexistent/path/to/locale_xyz.frl")
	if err == nil {
		t.Error("expected error for missing file, got nil")
	}
}

func TestLoadLocaleReader_MalformedXML(t *testing.T) {
	err := LoadLocaleReader(strings.NewReader("<invalid xml"))
	if err == nil {
		t.Error("expected error for malformed XML, got nil")
	}
}

func TestLookupLocale_NilRoot(t *testing.T) {
	v := lookupLocale(nil, "Objects,Text")
	if v != "" {
		t.Errorf("lookupLocale(nil, ...) = %q, want empty", v)
	}
}

func TestSetLocale_NilRoot(t *testing.T) {
	// Should not panic when root is nil.
	setLocale(nil, "Objects,Text", "value")
}

func TestLocaleName_NilCurrent(t *testing.T) {
	localeState.mu.Lock()
	orig := localeState.current
	localeState.current = nil
	localeState.mu.Unlock()

	name := LocaleName()
	if name != "en" {
		t.Errorf("LocaleName with nil current = %q, want 'en'", name)
	}

	localeState.mu.Lock()
	localeState.current = orig
	localeState.mu.Unlock()
}

func TestParseLocaleXML_WithChildren(t *testing.T) {
	xmlContent := `<Locale Name="xx">
  <Section Name="Sec">
    <Item Name="K1" Text="V1"/>
    <Item Name="K2" Text="V2"/>
  </Section>
</Locale>`
	root, err := parseLocaleXML(strings.NewReader(xmlContent))
	if err != nil {
		t.Fatalf("parseLocaleXML: %v", err)
	}
	if len(root.Children) == 0 {
		t.Fatal("expected children on root")
	}
	sec := root.Children[0]
	if len(sec.Children) != 2 {
		t.Errorf("expected 2 children in section, got %d", len(sec.Children))
	}
}

func TestXmlToNode_NameFallbackToXMLName(t *testing.T) {
	// Element without Name attr — XMLName.Local should be used as fallback.
	xmlContent := `<Locale Name="xx"><Category><Item Name="Key" Text="Val"/></Category></Locale>`
	root, err := parseLocaleXML(strings.NewReader(xmlContent))
	if err != nil {
		t.Fatalf("parseLocaleXML: %v", err)
	}
	// The Category element has no Name attr, so xmlToNode uses XMLName.Local = "Category".
	if len(root.Children) == 0 {
		t.Fatal("expected children")
	}
	if root.Children[0].Name == "" {
		t.Error("expected XMLName.Local fallback, got empty name")
	}
}

func TestLookupLocale_MultiLevel(t *testing.T) {
	xmlContent := `<Locale Name="xx">
  <Section Name="A">
    <Sub Name="B">
      <Item Name="C" Text="deep"/>
    </Sub>
  </Section>
</Locale>`
	root, err := parseLocaleXML(strings.NewReader(xmlContent))
	if err != nil {
		t.Fatalf("parseLocaleXML: %v", err)
	}
	v := lookupLocale(root, "A,B,C")
	if v != "deep" {
		t.Errorf("lookupLocale multilevel = %q, want 'deep'", v)
	}
}
