package utils

import (
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
